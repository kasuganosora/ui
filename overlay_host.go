package ui

import (
	"fmt"
	"math"

	"github.com/kasuganosora/ui/font/atlas"
	"github.com/kasuganosora/ui/font/textrender"
	"github.com/kasuganosora/ui/icon/material"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// overlayBackendFactory is a seam for tests: a headless harness can supply a
// software backend instead of creating a second GPU pipeline bound to a real
// NSView/HWND. Production always uses the platform backend factory.
var overlayBackendFactory = platformCreateBackend

// overlayWindow owns one auxiliary OS window. Texture handles are local to a
// render backend, so every floating window needs its own glyph atlas, text
// renderer and command buffer; sharing the main window's atlas would make
// drawText silently drop every glyph.
type overlayWindow struct {
	win     platform.Window
	backend render.Backend
	buf     *render.CommandBuffer
	glyphs  *atlas.Atlas
	text    *textrender.Renderer
	cfg     *widget.Config
	draw    func(*widget.Config, *render.CommandBuffer)
	active  bool
	lastW   int
	lastH   int
}

// OpenOverlay implements widget.OverlayWindowHost. It creates (or reuses) a
// borderless, transparent, always-on-top OS window and starts painting it each
// frame. x/y are main-window client coordinates; the host converts them to
// screen coordinates itself.
func (a *App) OpenOverlay(key widget.Widget, draw func(*widget.Config, *render.CommandBuffer), x, y, w, h float32) error {
	if key == nil {
		return fmt.Errorf("overlay: nil key")
	}
	a.overlaysMu.Lock()
	defer a.overlaysMu.Unlock()

	ow := a.overlays[key]
	if ow == nil {
		var err error
		ow, err = a.createOverlay(w, h)
		if err != nil {
			return err
		}
		a.overlays[key] = ow
	}
	ow.draw = draw
	ow.active = true
	a.placeOverlay(ow, x, y, w, h)
	ow.win.SetVisible(true)
	return nil
}

// CloseOverlay implements widget.OverlayWindowHost. The window is hidden but
// kept alive so a repeated hover does not recreate a native window and GPU
// pipeline on every mouse-enter.
func (a *App) CloseOverlay(key widget.Widget) {
	a.overlaysMu.Lock()
	defer a.overlaysMu.Unlock()
	if ow := a.overlays[key]; ow != nil {
		ow.active = false
		ow.win.SetVisible(false)
	}
}

func (a *App) createOverlay(w, h float32) (*overlayWindow, error) {
	iw, ih := int(math.Ceil(float64(w))), int(math.Ceil(float64(h)))
	if iw < 1 {
		iw = 1
	}
	if ih < 1 {
		ih = 1
	}
	win, err := a.plat.CreateWindow(platform.WindowOptions{
		Width:       iw,
		Height:      ih,
		Visible:     false,
		Decorated:   false,
		Resizable:   false,
		Transparent: true,
		TopMost:     true,
		Passthrough: true, // never steal clicks from the main window
	})
	if err != nil {
		return nil, fmt.Errorf("overlay: create window: %w", err)
	}

	backend, err := overlayBackendFactory(a.opts.Backend, win)
	if err != nil {
		win.Destroy()
		return nil, fmt.Errorf("overlay: create backend: %w", err)
	}
	fw, fh := win.FramebufferSize()
	backend.Resize(fw, fh)

	glyphs := atlas.New(atlas.Options{Width: 512, Height: 512, Backend: backend})
	if err := glyphs.EnsureTexture(); err != nil {
		backend.Destroy()
		win.Destroy()
		return nil, fmt.Errorf("overlay: glyph atlas: %w", err)
	}
	text := textrender.New(textrender.Options{
		Manager:   a.fontMgr,
		Atlas:     glyphs,
		DPIScale:  backend.DPIScale(),
		KeepAlive: a.plat.ProcessMessages,
		Backend:   backend,
	})

	// Copy the main config but point every renderer-bound field at this
	// window's resources. OverlayHost is cleared so a floating widget cannot
	// recursively open another window through its draw config.
	cfg := *a.cfg
	cfg.TextRenderer = &textDrawerAdapter{
		renderer:        text,
		fontID:          a.fontID,
		fallbackFontIDs: a.fallbackFontIDs,
		engine:          a.fontEngine,
	}
	cfg.Window = win
	cfg.Backend = backend
	cfg.IconRegistry = material.NewRegistry(backend)
	cfg.OverlayHost = nil

	return &overlayWindow{
		win:     win,
		backend: backend,
		buf:     render.NewCommandBuffer(),
		glyphs:  glyphs,
		text:    text,
		cfg:     &cfg,
		lastW:   fw,
		lastH:   fh,
	}, nil
}

func (a *App) placeOverlay(ow *overlayWindow, x, y, w, h float32) {
	iw, ih := int(math.Ceil(float64(w))), int(math.Ceil(float64(h)))
	if cw, ch := ow.win.Size(); cw != iw || ch != ih {
		ow.win.SetSize(iw, ih)
		fw, fh := ow.win.FramebufferSize()
		ow.backend.Resize(fw, fh)
		ow.lastW, ow.lastH = fw, fh
		ow.cfg.Window = ow.win
	}
	sx, sy := a.clientToTopLeftScreen(int(x), int(y))
	ow.win.SetPosition(sx, sy)
}

// clientToTopLeftScreen returns screen coordinates in SetPosition's
// top-left-origin convention. AppKit's ClientToScreen uses bottom-left origin,
// so darwin exposes an explicit converter; other platforms already return
// top-left coordinates.
func (a *App) clientToTopLeftScreen(x, y int) (int, int) {
	type topLeftConverter interface {
		ClientToTopLeftScreen(x, y int) (int, int)
	}
	if conv, ok := a.win.(topLeftConverter); ok {
		return conv.ClientToTopLeftScreen(x, y)
	}
	return a.win.ClientToScreen(x, y)
}

// renderOverlays paints every active auxiliary window. Called once per frame
// after the main window has been submitted.
func (a *App) renderOverlays() {
	a.overlaysMu.Lock()
	defer a.overlaysMu.Unlock()
	for _, ow := range a.overlays {
		if !ow.active || ow.draw == nil {
			continue
		}
		fw, fh := ow.win.FramebufferSize()
		if fw != ow.lastW || fh != ow.lastH {
			ow.backend.Resize(fw, fh)
			ow.lastW, ow.lastH = fw, fh
		}
		ow.backend.BeginFrame()
		ow.text.BeginFrame()
		ow.buf.Reset()
		ow.draw(ow.cfg, ow.buf)
		_ = ow.text.Upload()
		ow.backend.Submit(ow.buf)
		ow.backend.EndFrame()
	}
}

// destroyOverlays releases auxiliary GPU resources before the platform tears
// down its window list.
func (a *App) destroyOverlays() {
	a.overlaysMu.Lock()
	defer a.overlaysMu.Unlock()
	for key, ow := range a.overlays {
		ow.active = false
		ow.text.Destroy()
		ow.backend.Destroy()
		ow.win.Destroy()
		delete(a.overlays, key)
	}
}

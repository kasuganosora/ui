package ui

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/font"
	"github.com/kasuganosora/ui/godot"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// softwareOverlayBackend builds a software renderer instead of a second GPU
// pipeline, so the auxiliary-window pipeline can be tested without a GUI.
func softwareOverlayBackend(_ BackendType, win platform.Window) (render.Backend, error) {
	b := godot.NewSoftwareBackend()
	if err := b.Init(win); err != nil {
		return nil, err
	}
	return b, nil
}

func newOverlayTestApp(t *testing.T) *App {
	t.Helper()
	var eng font.Engine
	if e := FallbackFontEngine(); e != nil {
		eng = e
	} else {
		eng = NewMockEngine()
	}
	mgr := font.NewManager(eng)
	fid, _ := mgr.RegisterFile("Default", font.WeightRegular, font.StyleNormal, DefaultFont())
	if fid == font.InvalidFontID {
		fid, _ = mgr.Register("Default", font.WeightRegular, font.StyleNormal, nil)
	}

	plat := godot.NewHeadlessPlatform()
	win := godot.NewHeadlessWindow(800, 600, 1.0)
	mainBackend := godot.NewSoftwareBackend()
	if err := mainBackend.Init(win); err != nil {
		t.Fatal(err)
	}

	return &App{
		opts:       AppOptions{Backend: BackendAuto},
		plat:       plat,
		win:        win,
		backend:    mainBackend,
		fontEngine: eng,
		fontMgr:    mgr,
		fontID:     fid,
		cfg:        widget.DefaultConfig(),
		overlays:   make(map[widget.Widget]*overlayWindow),
	}
}

// TestOverlayHostOwnsIndependentRendering verifies the full auxiliary-window
// pipeline: a second window, a second backend, a private glyph atlas, and a
// per-frame Submit that actually produces pixels (including text).
func TestOverlayHostOwnsIndependentRendering(t *testing.T) {
	orig := overlayBackendFactory
	overlayBackendFactory = softwareOverlayBackend
	defer func() { overlayBackendFactory = orig }()

	a := newOverlayTestApp(t)
	defer a.destroyOverlays()
	defer a.backend.Destroy()

	key := widget.NewDiv(core.NewTree(), a.cfg)
	draw := func(cfg *widget.Config, buf *render.CommandBuffer) {
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(0, 0, 160, 40),
			FillColor: uimath.RGBA(0, 0, 0, 0.75),
		}, 1, 1)
		cfg.TextRenderer.DrawText(buf, "Tooltip", 8, 12, 14, 140, uimath.ColorWhite, 1)
	}

	if err := a.OpenOverlay(key, draw, 40, 60, 160, 40); err != nil {
		t.Fatalf("OpenOverlay: %v", err)
	}
	if len(a.overlays) != 1 {
		t.Fatalf("expected 1 auxiliary window, got %d", len(a.overlays))
	}
	ow := a.overlays[key]
	if ow.win == a.win {
		t.Fatal("auxiliary window is the main window")
	}
	if ow.backend == a.backend {
		t.Fatal("auxiliary window reuses the main backend")
	}
	if ow.glyphs == nil || ow.text == nil || ow.buf == nil {
		t.Fatal("auxiliary window is missing its own rendering resources")
	}
	if !ow.active {
		t.Fatal("auxiliary window should be active after OpenOverlay")
	}

	a.renderOverlays()

	// The label must be emitted into the auxiliary window's own command
	// buffer. If it were still bound to the main atlas, the second backend
	// would silently drop the glyphs (texture handles are per backend).
	var texts int
	for _, c := range ow.buf.Commands() {
		if c.Type == render.CmdText && c.Text != nil && len(c.Text.Glyphs) > 0 {
			texts++
		}
	}
	if texts == 0 {
		t.Fatal("no glyph commands submitted to the auxiliary backend")
	}

	img, err := ow.backend.ReadPixels()
	if err != nil {
		t.Fatalf("ReadPixels: %v", err)
	}
	// The bubble is a flat dark fill; any other visible color is glyph
	// coverage. Glyphs are alpha-blended onto that fill, so the resulting
	// pixels are grey rather than pure white — do not assert a bright white.
	nonBlank, glyphPixels := 0, 0
	for i := 0; i < len(img.Pix); i += 4 {
		r, g, b, a := img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3]
		if a == 0 {
			continue
		}
		nonBlank++
		if r != 0 || g != 0 || b != 0 || a != 191 {
			glyphPixels++
		}
	}
	if nonBlank == 0 {
		t.Fatal("auxiliary framebuffer is empty; draw/Submit did not reach the second backend")
	}
	if glyphPixels == 0 {
		t.Fatal("no glyph pixels in the auxiliary framebuffer; the private atlas was not uploaded")
	}

	// A second OpenOverlay must reuse the same native window, not create a new
	// GPU pipeline on every mouse-enter.
	secondWin := ow.win
	if err := a.OpenOverlay(key, draw, 50, 70, 160, 40); err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if len(a.overlays) != 1 || a.overlays[key].win != secondWin {
		t.Fatal("reopening an overlay created a second window")
	}

	a.CloseOverlay(key)
	if a.overlays[key].active {
		t.Fatal("overlay still active after CloseOverlay")
	}
	// Hidden overlays must not be submitted again.
	before := a.overlays[key].buf.Len()
	a.renderOverlays()
	if a.overlays[key].buf.Len() != before {
		t.Fatal("hidden overlay was still rendered")
	}
}

// TestOverlayHostUsesTopLeftScreenCoordinates guards against the AppKit
// bottom-left/top-left mismatch between ClientToScreen and SetPosition.
func TestOverlayHostUsesTopLeftScreenCoordinates(t *testing.T) {
	a := newOverlayTestApp(t)
	defer a.backend.Destroy()

	type converter interface {
		ClientToTopLeftScreen(x, y int) (int, int)
	}
	conv := &topLeftConverterWindow{Window: a.win}
	a.win = conv
	sx, sy := a.clientToTopLeftScreen(30, 40)
	if sx != 130 || sy != 90 {
		t.Fatalf("screen coords = (%d,%d), want (130,90)", sx, sy)
	}
}

type topLeftConverterWindow struct {
	platform.Window
}

func (w *topLeftConverterWindow) ClientToTopLeftScreen(x, y int) (int, int) {
	return x + 100, y + 50
}

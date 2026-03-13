package godot

import (
	"fmt"
	"sync"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/font"
	"github.com/kasuganosora/ui/font/atlas"
	"github.com/kasuganosora/ui/font/textrender"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// UIOptions configures a headless UI instance for Godot embedding.
type UIOptions struct {
	Width    int     // Viewport width in logical pixels (default 960)
	Height   int     // Viewport height in logical pixels (default 540)
	DPIScale float32 // DPI scale factor (default 1.0)

	// Font is the path to the primary font file.
	// If empty, a mock font engine is used (suitable for testing).
	Font string

	// FallbackFonts are additional font paths for symbol/CJK fallback.
	FallbackFonts []string

	// OnLayout is an optional custom layout callback.
	OnLayout func(tree *core.Tree, root widget.Widget, w, h float32)
}

// UI is a headless UI instance designed for embedding in Godot.
// It manages the full widget pipeline (tree, layout, rendering) without
// requiring an OS window or GPU context.
//
// Usage:
//
//	ui := godot.NewUI(godot.UIOptions{Width: 800, Height: 600})
//	defer ui.Destroy()
//	ui.LoadHTML(`<div style="color:white">Hello</div>`)
//	ui.Frame(0.016) // renders one frame
//	pixels := ui.Pixels() // RGBA pixel data
type UI struct {
	mu sync.Mutex

	plat     *HeadlessPlatform
	win      *HeadlessWindow
	backend  *SoftwareBackend
	tree     *core.Tree
	dispatch *core.Dispatcher
	cfg      *widget.Config
	buf      *render.CommandBuffer

	fontEngine      font.Engine
	fontID          font.ID
	fallbackFontIDs []font.ID
	glyphAtlas      *atlas.Atlas
	textRenderer    *textrender.Renderer

	root     widget.Widget
	onLayout func(tree *core.Tree, root widget.Widget, w, h float32)

	// Event state
	lastHoverTarget core.ElementID
}

// NewUI creates a new headless UI instance.
func NewUI(opts UIOptions) (*UI, error) {
	if opts.Width <= 0 {
		opts.Width = 960
	}
	if opts.Height <= 0 {
		opts.Height = 540
	}
	if opts.DPIScale <= 0 {
		opts.DPIScale = 1.0
	}

	u := &UI{
		onLayout: opts.OnLayout,
	}

	// Platform
	u.plat = NewHeadlessPlatform()
	u.win = NewHeadlessWindow(opts.Width, opts.Height, opts.DPIScale)

	// Backend
	u.backend = NewSoftwareBackend()
	if err := u.backend.Init(u.win); err != nil {
		return nil, fmt.Errorf("backend init: %w", err)
	}

	// Font engine
	u.fontEngine = newMockEngine()
	mgr := font.NewManager(u.fontEngine)

	if opts.Font != "" {
		if id, err := mgr.RegisterFile("Default", font.WeightRegular, font.StyleNormal, opts.Font); err == nil {
			u.fontID = id
		}
	}
	if u.fontID == font.InvalidFontID {
		u.fontID, _ = mgr.Register("Default", font.WeightRegular, font.StyleNormal, nil)
	}

	for _, fb := range opts.FallbackFonts {
		if id, err := mgr.RegisterFile("Symbol", font.WeightRegular, font.StyleNormal, fb); err == nil {
			u.fallbackFontIDs = append(u.fallbackFontIDs, id)
		}
	}

	u.fontEngine.SetDPIScale(opts.DPIScale)
	u.glyphAtlas = atlas.New(atlas.Options{Width: 512, Height: 512, Backend: u.backend})
	u.glyphAtlas.EnsureTexture()

	u.textRenderer = textrender.New(textrender.Options{
		Manager:  mgr,
		Atlas:    u.glyphAtlas,
		DPIScale: opts.DPIScale,
	})

	// Widget tree
	u.tree = core.NewTree()
	u.dispatch = core.NewDispatcher(u.tree)
	u.cfg = widget.DefaultConfig()
	u.cfg.TextRenderer = &textDrawer{
		renderer:        u.textRenderer,
		fontID:          u.fontID,
		fallbackFontIDs: u.fallbackFontIDs,
		engine:          u.fontEngine,
	}
	u.cfg.Window = u.win
	u.cfg.Platform = u.plat
	u.cfg.Backend = u.backend
	u.buf = render.NewCommandBuffer()

	return u, nil
}

// LoadHTML parses HTML+CSS and builds the widget tree.
// Returns the root Div for programmatic manipulation.
func (u *UI) LoadHTML(html string) *widget.Div {
	u.mu.Lock()
	defer u.mu.Unlock()

	root := widget.NewDiv(u.tree, u.cfg)
	u.root = root
	u.tree.AppendChild(u.tree.Root(), root.ElementID())
	return root
}

// SetRoot sets the root widget directly (for programmatic UI building).
func (u *UI) SetRoot(w widget.Widget) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.root = w
	u.tree.AppendChild(u.tree.Root(), w.ElementID())
}

// Frame processes one frame: handles pending events, runs layout, and renders
// to the pixel buffer. dt is the delta time in seconds since the last frame.
func (u *UI) Frame(dt float32) {
	u.mu.Lock()
	defer u.mu.Unlock()

	// Always update framebuffer size, even without a root widget.
	fw, fh := u.win.FramebufferSize()
	u.backend.Resize(fw, fh)

	if u.root == nil {
		return
	}

	// Layout first, so hit testing uses correct positions.
	w, h := u.win.Size()

	u.tree.SetLayout(u.tree.Root(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, float32(w), float32(h)),
	})
	if u.onLayout != nil {
		u.onLayout(u.tree, u.root, float32(w), float32(h))
	} else {
		// Default: root fills viewport
		u.tree.SetLayout(u.root.ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(0, 0, float32(w), float32(h)),
		})
	}

	// Process injected events (layout is already set for hit testing)
	events := u.plat.PollEvents()
	for i := range events {
		u.handleEvent(&events[i])
	}

	// Render
	u.backend.BeginFrame()
	u.textRenderer.BeginFrame()
	u.buf.Reset()
	u.root.Draw(u.buf)
	u.textRenderer.Upload()
	u.backend.Submit(u.buf)
	u.backend.EndFrame()
	u.tree.ClearAllDirty()
}

// Pixels returns the rendered RGBA pixel buffer.
// Length = FramebufferWidth * FramebufferHeight * 4.
// Valid until the next Frame() or Resize() call.
func (u *UI) Pixels() []byte {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.backend.Pixels()
}

// FramebufferSize returns the framebuffer dimensions in physical pixels.
func (u *UI) FramebufferSize() (int, int) {
	return u.backend.FramebufferSize()
}

// Resize changes the viewport size (logical pixels).
func (u *UI) Resize(width, height int) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.win.SetSize(width, height)
}

// InjectEvent injects an input event (mouse, keyboard, etc.) into the UI.
func (u *UI) InjectEvent(evt event.Event) {
	u.plat.InjectEvent(evt)
}

// InjectMouseMove injects a mouse move event at the given logical pixel coordinates.
func (u *UI) InjectMouseMove(x, y float32) {
	u.plat.InjectEvent(event.Event{
		Type:    event.MouseMove,
		GlobalX: x,
		GlobalY: y,
	})
}

// InjectMouseClick injects a mouse click at the given logical pixel coordinates.
func (u *UI) InjectMouseClick(x, y float32, button event.MouseButton) {
	u.plat.InjectEvent(event.Event{
		Type:    event.MouseDown,
		GlobalX: x, GlobalY: y,
		Button: button,
	})
	u.plat.InjectEvent(event.Event{
		Type:    event.MouseUp,
		GlobalX: x, GlobalY: y,
		Button: button,
	})
	u.plat.InjectEvent(event.Event{
		Type:    event.MouseClick,
		GlobalX: x, GlobalY: y,
		Button: button,
	})
}

// InjectScroll injects a mouse wheel scroll event.
func (u *UI) InjectScroll(x, y, deltaX, deltaY float32) {
	u.plat.InjectEvent(event.Event{
		Type:    event.MouseWheel,
		GlobalX: x, GlobalY: y,
		WheelDX: deltaX,
		WheelDY: deltaY,
	})
}

// InjectKeyDown injects a key press event.
func (u *UI) InjectKeyDown(key event.Key, modifiers event.Modifiers) {
	u.plat.InjectEvent(event.Event{
		Type:      event.KeyDown,
		Key:       key,
		Modifiers: modifiers,
	})
}

// InjectKeyUp injects a key release event.
func (u *UI) InjectKeyUp(key event.Key, modifiers event.Modifiers) {
	u.plat.InjectEvent(event.Event{
		Type:      event.KeyUp,
		Key:       key,
		Modifiers: modifiers,
	})
}

// InjectChar injects a character input event (for text input).
func (u *UI) InjectChar(ch rune) {
	u.plat.InjectEvent(event.Event{
		Type: event.KeyPress,
		Char: ch,
	})
}

// Tree returns the underlying element tree for advanced manipulation.
func (u *UI) Tree() *core.Tree { return u.tree }

// Config returns the widget configuration.
func (u *UI) Config() *widget.Config { return u.cfg }

// Destroy releases all resources.
func (u *UI) Destroy() {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.textRenderer != nil {
		u.textRenderer.Destroy()
	}
	if u.backend != nil {
		u.backend.Destroy()
	}
}

// handleEvent routes an event to the widget tree.
func (u *UI) handleEvent(evt *event.Event) {
	switch evt.Type {
	case event.MouseMove, event.MouseDown, event.MouseUp, event.MouseClick:
		target := u.tree.HitTest(evt.GlobalX, evt.GlobalY)

		// Hover tracking
		if evt.Type == event.MouseMove {
			u.tree.Walk(u.tree.Root(), func(id core.ElementID, _ int) bool {
				u.tree.SetHovered(id, id == target)
				return true
			})
			if target != u.lastHoverTarget {
				if u.lastHoverTarget != core.InvalidElementID {
					u.dispatch.Dispatch(u.lastHoverTarget, &event.Event{Type: event.MouseLeave})
				}
				if target != core.InvalidElementID {
					u.dispatch.Dispatch(target, &event.Event{Type: event.MouseEnter})
				}
				u.lastHoverTarget = target
			}
		}

		if target != core.InvalidElementID {
			u.dispatch.Dispatch(target, evt)
		}

	case event.MouseWheel:
		// Dispatch to hovered element
		target := u.tree.HitTest(evt.GlobalX, evt.GlobalY)
		if target != core.InvalidElementID {
			u.dispatch.Dispatch(target, evt)
		}

	case event.KeyDown, event.KeyUp, event.KeyPress:
		// Route to focused element
		u.tree.Walk(u.tree.Root(), func(id core.ElementID, _ int) bool {
			if e := u.tree.Get(id); e != nil && e.IsFocused() {
				u.dispatch.Dispatch(id, evt)
				return false
			}
			return true
		})
	}
}

// --- Mock font engine (copied from app.go to avoid circular import) ---

type mockEngine struct {
	glyphs map[rune]font.GlyphID
	nextG  font.GlyphID
}

func newMockEngine() *mockEngine {
	e := &mockEngine{glyphs: make(map[rune]font.GlyphID), nextG: 1}
	for r := rune(32); r < 127; r++ {
		e.glyphs[r] = e.nextG
		e.nextG++
	}
	for r := rune(0x4E00); r <= rune(0x9FFF); r++ {
		e.glyphs[r] = e.nextG
		e.nextG++
	}
	for r := rune(0x3000); r <= rune(0x303F); r++ {
		e.glyphs[r] = e.nextG
		e.nextG++
	}
	for r := rune(0xFF00); r <= rune(0xFFEF); r++ {
		e.glyphs[r] = e.nextG
		e.nextG++
	}
	e.glyphs['…'] = e.nextG
	e.nextG++
	return e
}

func (e *mockEngine) LoadFont([]byte) (font.ID, error)     { return 1, nil }
func (e *mockEngine) LoadFontFile(string) (font.ID, error)  { return 1, nil }
func (e *mockEngine) UnloadFont(font.ID)                    {}
func (e *mockEngine) SetDPIScale(float32)                    {}
func (e *mockEngine) Destroy()                              {}
func (e *mockEngine) Kerning(font.ID, font.GlyphID, font.GlyphID, float32) float32 { return 0 }
func (e *mockEngine) FontMetrics(_ font.ID, size float32) font.Metrics {
	return font.Metrics{Ascent: size * 0.8, Descent: size * 0.2, LineHeight: size * 1.2, UnitsPerEm: 1000}
}
func (e *mockEngine) GlyphIndex(_ font.ID, r rune) font.GlyphID { return e.glyphs[r] }
func (e *mockEngine) GlyphMetrics(_ font.ID, _ font.GlyphID, size float32) font.GlyphMetrics {
	adv := size * 0.6
	return font.GlyphMetrics{Width: adv, Height: size, BearingX: 0, BearingY: size * 0.8, Advance: adv}
}
func (e *mockEngine) RasterizeGlyph(_ font.ID, _ font.GlyphID, size float32, sdf bool) (font.GlyphBitmap, error) {
	w, h := int(size*0.6), int(size)
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	data := make([]byte, w*h)
	for i := range data {
		data[i] = 255
	}
	return font.GlyphBitmap{Width: w, Height: h, Data: data, SDF: sdf}, nil
}
func (e *mockEngine) HasGlyph(_ font.ID, r rune) bool {
	_, ok := e.glyphs[r]
	return ok
}
func (e *mockEngine) HasColorGlyphs(font.ID) bool { return false }

// textDrawer bridges textrender.Renderer to widget.TextDrawer.
type textDrawer struct {
	renderer        *textrender.Renderer
	fontID          font.ID
	fallbackFontIDs []font.ID
	engine          font.Engine
}

func (a *textDrawer) DrawText(buf *render.CommandBuffer, text string, x, y, fontSize, maxWidth float32, color uimath.Color, opacity float32) {
	a.renderer.DrawText(buf, text, textrender.DrawOptions{
		ShapeOpts: font.ShapeOptions{
			FontID:          a.fontID,
			FallbackFontIDs: a.fallbackFontIDs,
			FontSize:        fontSize,
			MaxWidth:        maxWidth,
			Truncate:        font.TruncateChar,
			MaxLines:        1,
		},
		OriginX: x,
		OriginY: y,
		Color:   color,
		Opacity: opacity,
	})
}

func (a *textDrawer) LineHeight(fontSize float32) float32 {
	m := a.engine.FontMetrics(a.fontID, fontSize)
	return m.Ascent + m.Descent
}

func (a *textDrawer) MeasureText(text string, fontSize float32) float32 {
	m := a.renderer.Measure(text, font.ShapeOptions{
		FontID:          a.fontID,
		FallbackFontIDs: a.fallbackFontIDs,
		FontSize:        fontSize,
	})
	return m.Width
}

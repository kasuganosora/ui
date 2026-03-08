//go:build windows

package ui

import (
	"fmt"
	"time"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/font"
	"github.com/kasuganosora/ui/font/atlas"
	"github.com/kasuganosora/ui/font/freetype"
	"github.com/kasuganosora/ui/font/textrender"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/platform/win32"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/render/vulkan"
	"github.com/kasuganosora/ui/widget"
)

// AppOptions configures an App instance.
type AppOptions struct {
	Title  string // Window title
	Width  int    // Window width (logical pixels)
	Height int    // Window height (logical pixels)
	Font   string // Path to font file (e.g. "C:\\Windows\\Fonts\\msyh.ttc")

	// OnLayout is an optional custom layout callback.
	// If nil, the App uses a basic auto-layout.
	OnLayout func(tree *core.Tree, root widget.Widget, w, h float32)
}

// App encapsulates the full UI application lifecycle:
// platform, renderer, font system, widget tree, event routing, and main loop.
type App struct {
	opts AppOptions

	// Subsystems
	plat     *win32.Platform
	win      platform.Window
	backend  render.Backend
	tree     *core.Tree
	dispatch *core.Dispatcher
	cfg      *widget.Config
	buf      *render.CommandBuffer

	// Font
	fontEngine   font.Engine
	fontID       font.ID
	glyphAtlas   *atlas.Atlas
	textRenderer *textrender.Renderer

	// Document
	doc  *Document
	root widget.Widget

	// Event state
	mouseDownTarget core.ElementID
	lastHoverTarget core.ElementID
	lastCursor      platform.CursorShape

	// Scrollable content (auto-detected from <main> tag)
	content *widget.Content
}

// NewApp creates and initializes a new App.
// Call Destroy() when done, or use defer.
func NewApp(opts AppOptions) (*App, error) {
	if opts.Width == 0 {
		opts.Width = 960
	}
	if opts.Height == 0 {
		opts.Height = 640
	}
	if opts.Title == "" {
		opts.Title = "GoUI"
	}

	a := &App{
		opts:       opts,
		lastCursor: platform.CursorArrow,
	}

	// --- Platform ---
	a.plat = win32.New()
	if err := a.plat.Init(); err != nil {
		return nil, fmt.Errorf("platform init: %w", err)
	}

	var err error
	a.win, err = a.plat.CreateWindow(platform.WindowOptions{
		Title:     opts.Title,
		Width:     opts.Width,
		Height:    opts.Height,
		Resizable: true,
		Visible:   true,
		Decorated: true,
	})
	if err != nil {
		a.plat.Terminate()
		return nil, fmt.Errorf("create window: %w", err)
	}

	// --- Renderer ---
	a.backend = vulkan.New()
	if err := a.backend.Init(a.win); err != nil {
		a.win.Destroy()
		a.plat.Terminate()
		return nil, fmt.Errorf("vulkan init: %w", err)
	}

	// --- Font System ---
	if ftEngine, err := freetype.New(); err == nil {
		a.fontEngine = ftEngine
	} else {
		a.fontEngine = newMockEngine()
	}

	mgr := font.NewManager(a.fontEngine)
	fontPath := opts.Font
	if fontPath == "" {
		fontPath = `C:\Windows\Fonts\msyh.ttc`
	}
	a.fontID, _ = mgr.RegisterFile("Default", font.WeightRegular, font.StyleNormal, fontPath)
	if a.fontID == font.InvalidFontID {
		a.fontID, _ = mgr.Register("Default", font.WeightRegular, font.StyleNormal, nil)
	}

	dpi := a.backend.DPIScale()
	a.fontEngine.SetDPIScale(dpi)

	a.glyphAtlas = atlas.New(atlas.Options{Width: 1024, Height: 1024, Backend: a.backend})
	a.glyphAtlas.EnsureTexture()

	a.textRenderer = textrender.New(textrender.Options{
		Manager:   mgr,
		Atlas:     a.glyphAtlas,
		DPIScale:  dpi,
		KeepAlive: a.plat.ProcessMessages,
	})

	// --- Widget Tree ---
	a.tree = core.NewTree()
	a.dispatch = core.NewDispatcher(a.tree)
	a.cfg = widget.DefaultConfig()
	a.cfg.TextRenderer = &textDrawerAdapter{
		renderer: a.textRenderer,
		fontID:   a.fontID,
		engine:   a.fontEngine,
	}
	a.cfg.Window = a.win
	a.cfg.Platform = a.plat
	a.buf = render.NewCommandBuffer()

	return a, nil
}

// LoadHTML parses HTML+CSS and builds the widget tree.
// Returns the Document for event binding via QueryByID, OnClick, etc.
func (a *App) LoadHTML(html string) *Document {
	a.doc = LoadHTMLDocument(a.tree, a.cfg, html, "")

	// Extract root widget
	if len(a.doc.Root.Children()) > 0 {
		a.root = a.doc.Root.Children()[0]
	} else {
		a.root = a.doc.Root
	}

	// Attach to tree
	a.tree.AppendChild(a.tree.Root(), a.root.ElementID())

	// Auto-detect scrollable content (<main> tag)
	if mains := a.doc.QueryByTag("main"); len(mains) > 0 {
		if c, ok := mains[0].(*widget.Content); ok {
			a.content = c
		}
	}

	return a.doc
}

// Tree returns the underlying element tree.
func (a *App) Tree() *core.Tree { return a.tree }

// Config returns the widget configuration.
func (a *App) Config() *widget.Config { return a.cfg }

// Window returns the platform window.
func (a *App) Window() platform.Window { return a.win }

// Run starts the main loop. Blocks until the window is closed.
func (a *App) Run() error {
	var lastW, lastH int
	frameCount := 0
	fpsStart := time.Now()

	var w32 *win32.Window
	if ww, ok := a.win.(*win32.Window); ok {
		w32 = ww
	}

	renderFrame := func() {
		fw, fh := a.win.FramebufferSize()
		if fw != lastW || fh != lastH {
			a.backend.Resize(fw, fh)
			lastW, lastH = fw, fh
			lw, lh := a.win.Size()
			a.tree.SetLayout(a.tree.Root(), core.LayoutResult{
				Bounds: uimath.NewRect(0, 0, float32(lw), float32(lh)),
			})
			if a.opts.OnLayout != nil {
				a.opts.OnLayout(a.tree, a.root, float32(lw), float32(lh))
			} else {
				AutoLayout(a.tree, a.root, float32(lw), float32(lh))
			}
		}

		if w32 != nil && w32.InSizeMove() {
			return
		}

		a.backend.BeginFrame()
		a.textRenderer.BeginFrame()
		a.buf.Reset()
		a.root.Draw(a.buf)
		a.textRenderer.Upload()
		a.backend.Submit(a.buf)
		a.backend.EndFrame()
	}

	if w32 != nil {
		w32.OnResize(renderFrame)
	}

	needsRedraw := true

	for !a.win.ShouldClose() {
		events := a.plat.PollEvents()
		if len(events) > 0 {
			needsRedraw = true
		}
		for i := range events {
			a.handleEvent(&events[i])
		}

		if a.tree.NeedsRender() {
			needsRedraw = true
		}

		if needsRedraw {
			renderFrame()
			a.tree.ClearAllDirty()
			needsRedraw = false
		}

		frameCount++
		if elapsed := time.Since(fpsStart); elapsed >= time.Second {
			fmt.Printf("\rFPS: %d  ", frameCount)
			frameCount = 0
			fpsStart = time.Now()
		}

		time.Sleep(time.Millisecond)
	}

	fmt.Println()
	return nil
}

// Destroy releases all resources.
func (a *App) Destroy() {
	if a.textRenderer != nil {
		a.textRenderer.Destroy()
	}
	if a.backend != nil {
		a.backend.Destroy()
	}
	if a.win != nil {
		a.win.Destroy()
	}
	if a.plat != nil {
		a.plat.Terminate()
	}
}

// handleEvent routes a platform event to the widget tree.
func (a *App) handleEvent(evt *event.Event) {
	switch evt.Type {
	case event.WindowResize:
		// Handled in renderFrame
	case event.WindowClose:
		a.win.SetShouldClose(true)
	case event.MouseWheel:
		if a.content != nil && a.root != nil {
			a.content.HandleWheel(evt.WheelDY)
			lw, lh := a.win.Size()
			if a.opts.OnLayout != nil {
				a.opts.OnLayout(a.tree, a.root, float32(lw), float32(lh))
			} else {
				AutoLayout(a.tree, a.root, float32(lw), float32(lh))
			}
		}
	case event.MouseMove, event.MouseDown, event.MouseUp, event.MouseClick:
		a.handleMouse(evt)
	case event.KeyDown, event.KeyUp, event.KeyPress,
		event.IMECompositionStart, event.IMECompositionUpdate, event.IMECompositionEnd:
		a.tree.Walk(a.tree.Root(), func(id core.ElementID, _ int) bool {
			if e := a.tree.Get(id); e != nil && e.IsFocused() {
				a.dispatch.Dispatch(id, evt)
				return false
			}
			return true
		})
	}
}

func (a *App) handleMouse(evt *event.Event) {
	// Scrollbar drag
	if a.content != nil {
		if a.content.IsScrollBarDragging() {
			if evt.Type == event.MouseMove {
				a.content.HandleScrollBarMove(evt.GlobalY)
				lw, lh := a.win.Size()
				if a.opts.OnLayout != nil {
					a.opts.OnLayout(a.tree, a.root, float32(lw), float32(lh))
				} else {
					AutoLayout(a.tree, a.root, float32(lw), float32(lh))
				}
				return
			}
			if evt.Type == event.MouseUp {
				a.content.HandleScrollBarUp()
				return
			}
		}
		if evt.Type == event.MouseDown && a.content.HandleScrollBarDown(evt.GlobalY) {
			bounds := a.content.Bounds()
			scrollBarX := bounds.X + bounds.Width - 10
			if evt.GlobalX >= scrollBarX {
				return
			}
			a.content.HandleScrollBarUp()
		}
	}

	target := a.tree.HitTest(evt.GlobalX, evt.GlobalY)

	// Drag support
	if a.mouseDownTarget != core.InvalidElementID && target != a.mouseDownTarget {
		if evt.Type == event.MouseMove || evt.Type == event.MouseUp {
			a.dispatch.Dispatch(a.mouseDownTarget, evt)
		}
	}

	// Hover + cursor
	if evt.Type == event.MouseMove {
		a.tree.Walk(a.tree.Root(), func(id core.ElementID, _ int) bool {
			a.tree.SetHovered(id, id == target)
			return true
		})
		if target != a.lastHoverTarget {
			if a.lastHoverTarget != core.InvalidElementID {
				a.dispatch.Dispatch(a.lastHoverTarget, &event.Event{Type: event.MouseLeave})
			}
			if target != core.InvalidElementID {
				a.dispatch.Dispatch(target, &event.Event{Type: event.MouseEnter})
			}
			a.lastHoverTarget = target
		}

		wantCursor := platform.CursorArrow
		for id := target; id != core.InvalidElementID; {
			if e := a.tree.Get(id); e != nil {
				if e.Type() == core.TypeInput {
					wantCursor = platform.CursorIBeam
					break
				}
				if e.Type() == core.TypeButton || e.HasHandler(event.MouseClick) {
					wantCursor = platform.CursorHand
					break
				}
				id = e.ParentID()
			} else {
				break
			}
		}
		if wantCursor != a.lastCursor {
			a.win.SetCursor(wantCursor)
			a.lastCursor = wantCursor
		}
	}

	if target != core.InvalidElementID {
		// Blur on click outside input
		if evt.Type == event.MouseDown {
			isInput := false
			for id := target; id != core.InvalidElementID; {
				if e := a.tree.Get(id); e != nil {
					if e.Type() == core.TypeInput {
						isInput = true
						break
					}
					id = e.ParentID()
				} else {
					break
				}
			}
			if !isInput {
				a.tree.ClearFocus()
			}
		}

		a.dispatch.Dispatch(target, evt)

		// Click synthesis
		if evt.Type == event.MouseDown {
			a.mouseDownTarget = target
		} else if evt.Type == event.MouseUp && target == a.mouseDownTarget {
			clickEvt := *evt
			clickEvt.Type = event.MouseClick
			a.dispatch.Dispatch(target, &clickEvt)
			a.mouseDownTarget = core.InvalidElementID
		}
	}

	if evt.Type == event.MouseUp {
		a.mouseDownTarget = core.InvalidElementID
	}
}

// AutoLayout performs a basic layout for HTML-based UIs.
// Handles the common Layout > Header + Body(Aside + Content) + Footer pattern.
func AutoLayout(tree *core.Tree, root widget.Widget, w, h float32) {
	tree.SetLayout(root.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, w, h),
	})

	children := root.Children()
	if len(children) == 0 {
		return
	}

	// Detect Layout pattern: Header + Body + Footer
	if len(children) >= 3 {
		if isType[*widget.Header](children[0]) && isType[*widget.Footer](children[len(children)-1]) {
			layoutHBF(tree, children, w, h)
			return
		}
	}

	// Fallback: vertical stack
	layoutVerticalStack(tree, root, 0, 0, w, h, 8)
}

func layoutHBF(tree *core.Tree, children []widget.Widget, w, h float32) {
	headerH := float32(56)
	footerH := float32(48)
	bodyH := h - headerH - footerH

	// Header
	tree.SetLayout(children[0].ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, w, headerH),
	})
	layoutHorizontalCenter(tree, children[0], 0, 0, w, headerH, 24)

	// Body (middle children)
	body := children[1]
	tree.SetLayout(body.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, headerH, w, bodyH),
	})

	bodyChildren := body.Children()
	if len(bodyChildren) >= 2 {
		// Aside + Content
		if isType[*widget.Aside](bodyChildren[0]) {
			asideW := float32(220)
			contentW := w - asideW

			tree.SetLayout(bodyChildren[0].ElementID(), core.LayoutResult{
				Bounds: uimath.NewRect(0, headerH, asideW, bodyH),
			})
			layoutVerticalStack(tree, bodyChildren[0], 0, headerH, asideW, bodyH, 8)

			tree.SetLayout(bodyChildren[1].ElementID(), core.LayoutResult{
				Bounds: uimath.NewRect(asideW, headerH, contentW, bodyH),
			})
			layoutContentArea(tree, bodyChildren[1], asideW, headerH, contentW, bodyH)
		}
	} else if len(bodyChildren) == 1 {
		// Just content
		tree.SetLayout(bodyChildren[0].ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(0, headerH, w, bodyH),
		})
		layoutContentArea(tree, bodyChildren[0], 0, headerH, w, bodyH)
	}

	// Footer
	footer := children[len(children)-1]
	tree.SetLayout(footer.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, headerH+bodyH, w, footerH),
	})
	layoutHorizontalCenter(tree, footer, 0, headerH+bodyH, w, footerH, 0)
}

func layoutContentArea(tree *core.Tree, content widget.Widget, x, y, w, h float32) {
	pad := float32(24)
	cx := x + pad
	cw := w - pad*2
	gap := float32(12)

	totalH := pad
	for _, child := range content.Children() {
		totalH += rowHeight(child) + gap
	}
	totalH += pad

	if c, ok := content.(*widget.Content); ok {
		c.SetContentHeight(totalH)
		c.ScrollBy(0)
	}

	scrollY := float32(0)
	if c, ok := content.(*widget.Content); ok {
		scrollY = c.ScrollY()
	}

	cy := y + pad - scrollY
	for _, child := range content.Children() {
		rh := rowHeight(child)
		tree.SetLayout(child.ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(cx, cy, cw, rh),
		})
		if len(child.Children()) > 0 {
			layoutHorizontalCenter(tree, child, cx, cy, cw, rh, 12)
		}
		cy += rh + gap
	}
}

func rowHeight(child widget.Widget) float32 {
	switch child.(type) {
	case *widget.TextArea:
		return 80
	case *widget.Empty:
		return 80
	case *widget.Progress:
		return 12
	}
	return 36
}

func layoutVerticalStack(tree *core.Tree, parent widget.Widget, x, y, w, h, gap float32) {
	children := parent.Children()
	if len(children) == 0 {
		return
	}
	itemH := float32(40)
	cy := y + 8
	for _, child := range children {
		tree.SetLayout(child.ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(x+4, cy, w-8, itemH),
		})
		cy += itemH + gap
	}
}

func layoutHorizontalCenter(tree *core.Tree, parent widget.Widget, x, y, w, h, gap float32) {
	children := parent.Children()
	if len(children) == 0 {
		return
	}
	n := float32(len(children))
	totalGap := gap * (n - 1)
	itemW := (w - totalGap) / n
	cx := x
	for _, child := range children {
		tree.SetLayout(child.ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(cx, y, itemW, h),
		})
		fillDescendants(tree, child, cx+4, y+4, itemW-8, h-8)
		cx += itemW + gap
	}
}

func fillDescendants(tree *core.Tree, parent widget.Widget, x, y, w, h float32) {
	for _, child := range parent.Children() {
		tree.SetLayout(child.ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(x, y, w, h),
		})
		fillDescendants(tree, child, x+2, y+2, w-4, h-4)
	}
}

func isType[T any](w widget.Widget) bool {
	_, ok := w.(T)
	return ok
}

// NewTextDrawer creates a widget.TextDrawer from a textrender.Renderer.
// Useful for tests that set up their own font pipeline.
func NewTextDrawer(renderer *textrender.Renderer, fontID font.ID, engine font.Engine) widget.TextDrawer {
	return &textDrawerAdapter{renderer: renderer, fontID: fontID, engine: engine}
}

// textDrawerAdapter bridges textrender.Renderer to widget.TextDrawer.
type textDrawerAdapter struct {
	renderer *textrender.Renderer
	fontID   font.ID
	engine   font.Engine
}

func (a *textDrawerAdapter) DrawText(buf *render.CommandBuffer, text string, x, y, fontSize, maxWidth float32, color uimath.Color, opacity float32) {
	a.renderer.DrawText(buf, text, textrender.DrawOptions{
		ShapeOpts: font.ShapeOptions{
			FontID:   a.fontID,
			FontSize: fontSize,
			MaxWidth: maxWidth,
			Truncate: font.TruncateChar,
			MaxLines: 1,
		},
		OriginX: x,
		OriginY: y,
		Color:   color,
		Opacity: opacity,
	})
}

func (a *textDrawerAdapter) LineHeight(fontSize float32) float32 {
	m := a.engine.FontMetrics(a.fontID, fontSize)
	return m.Ascent + m.Descent
}

func (a *textDrawerAdapter) MeasureText(text string, fontSize float32) float32 {
	m := a.renderer.Measure(text, font.ShapeOptions{
		FontID:   a.fontID,
		FontSize: fontSize,
	})
	return m.Width
}

// NewMockEngine creates a fallback font engine for testing when FreeType is unavailable.
func NewMockEngine() font.Engine { return newMockEngine() }

// mockEngine is a fallback font engine when FreeType is unavailable.
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

func (e *mockEngine) LoadFont([]byte) (font.ID, error)    { return 1, nil }
func (e *mockEngine) LoadFontFile(string) (font.ID, error) { return 1, nil }
func (e *mockEngine) UnloadFont(font.ID)                   {}
func (e *mockEngine) SetDPIScale(float32)                   {}
func (e *mockEngine) Destroy()                             {}
func (e *mockEngine) Kerning(font.ID, font.GlyphID, font.GlyphID, float32) float32 {
	return 0
}
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

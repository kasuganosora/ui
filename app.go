//go:build windows

package ui

import (
	"fmt"
	"time"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	"github.com/kasuganosora/ui/font"
	"github.com/kasuganosora/ui/font/atlas"
	"github.com/kasuganosora/ui/font/freetype"
	"github.com/kasuganosora/ui/font/textrender"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/theme"
	"github.com/kasuganosora/ui/platform/win32"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/render/dx9"
	"github.com/kasuganosora/ui/render/dx11"
	"github.com/kasuganosora/ui/render/gl"
	"github.com/kasuganosora/ui/render/vulkan"
	"github.com/kasuganosora/ui/icon/material"
	"github.com/kasuganosora/ui/widget"
)

// BackendType selects the rendering backend.
type BackendType int

const (
	BackendAuto    BackendType = iota // Try Vulkan first, fall back to DX11
	BackendVulkan                     // Force Vulkan
	BackendDX11                       // Force DirectX 11
	BackendDX9                        // Force DirectX 9
	BackendOpenGL                     // Force OpenGL 3.3
)

// AppOptions configures an App instance.
type AppOptions struct {
	Title   string      // Window title
	Width   int         // Window width (logical pixels)
	Height  int         // Window height (logical pixels)
	Font    string      // Path to font file (e.g. "C:\\Windows\\Fonts\\msyh.ttc")
	Backend BackendType // Rendering backend (default: auto)

	// OnLayout is an optional custom layout callback.
	// If nil, the App uses a basic auto-layout.
	OnLayout func(tree *core.Tree, root widget.Widget, w, h float32)
}

// createBackend creates a render.Backend based on the selected type.
func createBackend(bt BackendType, win platform.Window) (render.Backend, error) {
	switch bt {
	case BackendDX9:
		b := dx9.New()
		if err := b.Init(win); err != nil {
			return nil, fmt.Errorf("dx9 init: %w", err)
		}
		return b, nil
	case BackendDX11:
		b := dx11.New()
		if err := b.Init(win); err != nil {
			return nil, fmt.Errorf("dx11 init: %w", err)
		}
		return b, nil
	case BackendVulkan:
		b := vulkan.New()
		if err := b.Init(win); err != nil {
			return nil, fmt.Errorf("vulkan init: %w", err)
		}
		return b, nil
	case BackendOpenGL:
		b := gl.New()
		if err := b.Init(win); err != nil {
			return nil, fmt.Errorf("gl init: %w", err)
		}
		return b, nil
	default: // BackendAuto: try Vulkan first, fall back to DX11
		b := vulkan.New()
		if err := b.Init(win); err == nil {
			return b, nil
		}
		d := dx11.New()
		if err := d.Init(win); err != nil {
			return nil, fmt.Errorf("no backend available: dx11: %w", err)
		}
		return d, nil
	}
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
	lastMouseX      float32
	lastMouseY      float32

	// Scrollable content (auto-detected from <main> tag)
	content *widget.Content

	// Scrollable sidebar (auto-detected from aside)
	sidebarDiv  *widget.Div
	sidebarList *widget.List
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
	a.backend, err = createBackend(opts.Backend, a.win)
	if err != nil {
		a.win.Destroy()
		a.plat.Terminate()
		return nil, err
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
	a.cfg.Backend = a.backend
	a.cfg.IconRegistry = material.NewRegistry(a.backend)
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

	// Auto-detect sidebar scrollable element (div or list child of aside)
	if asides := a.doc.QueryByTag("aside"); len(asides) > 0 {
		for _, child := range asides[0].Children() {
			if d, ok := child.(*widget.Div); ok {
				a.sidebarDiv = d
				d.SetScrollable(true) // enable clipping + scroll
				break
			}
			if l, ok := child.(*widget.List); ok {
				a.sidebarList = l
				break
			}
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

// SetTheme applies a theme to the application, updating widget.Config values
// and injecting CSS variables into the current document's stylesheet.
func (a *App) SetTheme(t *theme.Theme) {
	// Update widget.Config from theme
	cv := t.ToConfig()
	a.cfg.PrimaryColor = cv.PrimaryColor
	a.cfg.TextColor = cv.TextColor
	a.cfg.BgColor = cv.BgColor
	a.cfg.BorderColor = cv.BorderColor
	a.cfg.DisabledColor = cv.DisabledColor
	a.cfg.HoverColor = cv.HoverColor
	a.cfg.ActiveColor = cv.ActiveColor
	a.cfg.FocusBorderColor = cv.FocusBorderColor
	a.cfg.ErrorColor = cv.ErrorColor
	a.cfg.FontSize = cv.FontSize
	a.cfg.FontSizeSm = cv.FontSizeSm
	a.cfg.FontSizeLg = cv.FontSizeLg
	a.cfg.SpaceXS = cv.SpaceXS
	a.cfg.SpaceSM = cv.SpaceSM
	a.cfg.SpaceMD = cv.SpaceMD
	a.cfg.SpaceLG = cv.SpaceLG
	a.cfg.SpaceXL = cv.SpaceXL
	a.cfg.BorderRadius = cv.BorderRadius
	a.cfg.BorderWidth = cv.BorderWidth
	a.cfg.ButtonHeight = cv.ButtonHeight
	a.cfg.InputHeight = cv.InputHeight

	// Inject CSS variables into document
	if a.doc != nil {
		a.doc.SetTheme(t)
	}

	// Mark tree dirty so everything redraws
	a.tree.MarkDirty(a.tree.Root())
}

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
		}

		if w32 != nil && w32.InSizeMove() {
			return
		}

		// Always re-run layout (handles Display toggling, scroll changes, etc.)
		lw, lh := a.win.Size()
		a.tree.SetLayout(a.tree.Root(), core.LayoutResult{
			Bounds: uimath.NewRect(0, 0, float32(lw), float32(lh)),
		})
		if a.opts.OnLayout != nil {
			a.opts.OnLayout(a.tree, a.root, float32(lw), float32(lh))
		} else {
			AutoLayout(a.tree, a.root, float32(lw), float32(lh))
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
			// If Draw() marked anything dirty (e.g. animations), schedule next frame
			needsRedraw = a.tree.NeedsRender()
			a.tree.ClearAllDirty()
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
		// WM_MOUSEWHEEL doesn't include cursor position — use last known
		mx, my := a.lastMouseX, a.lastMouseY
		handled := false
		// Check if mouse is over sidebar
		if a.sidebarDiv != nil {
			sb := a.sidebarDiv.Bounds()
			if mx >= sb.X && mx < sb.X+sb.Width &&
				my >= sb.Y && my < sb.Y+sb.Height {
				a.sidebarDiv.ScrollTo(0, a.sidebarDiv.ScrollY()-evt.WheelDY*30)
				handled = true
			}
		}
		if !handled && a.sidebarList != nil {
			sb := a.sidebarList.Bounds()
			if mx >= sb.X && mx < sb.X+sb.Width &&
				my >= sb.Y && my < sb.Y+sb.Height {
				newY := a.sidebarList.ScrollY() - evt.WheelDY*30
				if newY < 0 {
					newY = 0
				}
				maxScroll := a.sidebarList.TotalHeight() - sb.Height
				if maxScroll < 0 {
					maxScroll = 0
				}
				if newY > maxScroll {
					newY = maxScroll
				}
				a.sidebarList.SetScrollY(newY)
				a.tree.MarkDirty(a.sidebarList.ElementID())
				handled = true
			}
		}
		// Otherwise scroll content
		if !handled && a.content != nil {
			a.content.HandleWheel(evt.WheelDY)
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
	// Track mouse position for wheel events
	a.lastMouseX = evt.GlobalX
	a.lastMouseY = evt.GlobalY

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
// Detects common patterns: Header at top, Footer at bottom, Aside+Content body.
func AutoLayout(tree *core.Tree, root widget.Widget, w, h float32) {
	tree.SetLayout(root.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, w, h),
	})

	children := root.Children()
	if len(children) == 0 {
		return
	}

	y := float32(0)

	// Detect header
	headerIdx := -1
	headerH := float32(48)
	if isType[*widget.Header](children[0]) {
		headerIdx = 0
		tree.SetLayout(children[0].ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(0, 0, w, headerH),
		})
		layoutHorizontalCenter(tree, children[0], 0, 0, w, headerH, 24)
		y = headerH
	}

	// Detect footer
	footerIdx := -1
	footerH := float32(48)
	if isType[*widget.Footer](children[len(children)-1]) {
		footerIdx = len(children) - 1
		footerH = 48
	}

	bodyH := h - y
	if footerIdx >= 0 {
		bodyH -= footerH
	}

	// Process body children (everything between header and footer)
	for i, child := range children {
		if i == headerIdx || i == footerIdx {
			continue
		}
		tree.SetLayout(child.ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(0, y, w, bodyH),
		})
		layoutBody(tree, child, 0, y, w, bodyH)
	}

	// Footer
	if footerIdx >= 0 {
		footer := children[footerIdx]
		fy := y + bodyH
		tree.SetLayout(footer.ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(0, fy, w, footerH),
		})
		layoutHorizontalCenter(tree, footer, 0, fy, w, footerH, 0)
	}
}

// layoutBody lays out a body element, detecting Aside+Content patterns.
func layoutBody(tree *core.Tree, body widget.Widget, x, y, w, h float32) {
	bodyChildren := body.Children()
	if len(bodyChildren) == 0 {
		return
	}

	// Find aside and content among children
	asideIdx := -1
	contentIdx := -1
	for i, c := range bodyChildren {
		if isType[*widget.Aside](c) {
			asideIdx = i
		} else if isType[*widget.Content](c) {
			contentIdx = i
		}
	}

	if asideIdx >= 0 && contentIdx >= 0 {
		asideW := float32(180)
		contentX := x + asideW
		contentW := w - asideW

		// Aside
		tree.SetLayout(bodyChildren[asideIdx].ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(x, y, asideW, h),
		})
		layoutSidebar(tree, bodyChildren[asideIdx], x, y, asideW, h)

		// Content
		tree.SetLayout(bodyChildren[contentIdx].ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(contentX, y, contentW, h),
		})
		layoutContentArea(tree, bodyChildren[contentIdx], contentX, y, contentW, h)
		return
	}

	// Single content child
	if len(bodyChildren) == 1 {
		tree.SetLayout(bodyChildren[0].ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(x, y, w, h),
		})
		if isType[*widget.Content](bodyChildren[0]) {
			layoutContentArea(tree, bodyChildren[0], x, y, w, h)
		} else {
			layoutBody(tree, bodyChildren[0], x, y, w, h)
		}
		return
	}

	// Fallback: vertical stack
	layoutVerticalStack(tree, body, x, y, w, h, 8)
}

// layoutSidebar lays out sidebar children vertically, supporting scrollable wrapper divs and lists.
func layoutSidebar(tree *core.Tree, aside widget.Widget, x, y, w, h float32) {
	children := aside.Children()
	if len(children) == 0 {
		return
	}

	// If there's a single List widget, give it the full aside bounds (it handles its own scrolling)
	if len(children) == 1 {
		if _, ok := children[0].(*widget.List); ok {
			tree.SetLayout(children[0].ElementID(), core.LayoutResult{
				Bounds: uimath.NewRect(x, y, w, h),
			})
			return
		}
	}

	// If there's a single wrapper div (e.g. sidebar-scroll), use it as scroll container
	var scrollDiv *widget.Div
	if len(children) == 1 {
		tree.SetLayout(children[0].ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(x, y, w, h),
		})
		if d, ok := children[0].(*widget.Div); ok {
			scrollDiv = d
		}
		children = children[0].Children()
	}

	// Calculate total content height
	itemH := float32(32)
	gap := float32(2)
	totalH := float32(8) // top padding
	for range children {
		totalH += itemH + gap
	}
	totalH += 8 // bottom padding

	// Apply and clamp scroll offset
	scrollY := float32(0)
	if scrollDiv != nil {
		scrollDiv.SetContentHeight(totalH)
		maxScroll := totalH - h
		if maxScroll < 0 {
			maxScroll = 0
		}
		sy := scrollDiv.ScrollY()
		if sy < 0 {
			sy = 0
		}
		if sy > maxScroll {
			sy = maxScroll
		}
		scrollDiv.ScrollTo(0, sy)
		scrollY = sy
	}

	cy := y + 8 - scrollY
	for _, child := range children {
		tree.SetLayout(child.ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(x+8, cy, w-16, itemH),
		})
		cy += itemH + gap
	}
}

func layoutContentArea(tree *core.Tree, content widget.Widget, x, y, w, h float32) {
	pad := float32(24)
	cx := x + pad
	cw := w - pad*2
	gap := float32(16)

	// Collect visible children, zero-out hidden ones
	var visible []widget.Widget
	for _, child := range content.Children() {
		if child.Style().Display == layout.DisplayNone {
			// Clear bounds so hidden sections don't render
			tree.SetLayout(child.ElementID(), core.LayoutResult{})
			continue
		}
		visible = append(visible, child)
	}

	// Calculate total content height
	totalH := pad
	for _, child := range visible {
		totalH += sectionHeight(child) + gap
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
	for _, child := range visible {
		sh := sectionHeight(child)
		tree.SetLayout(child.ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(cx, cy, cw, sh),
		})
		layoutSection(tree, child, cx, cy, cw, sh)
		cy += sh + gap
	}
}

// sectionHeight estimates the height of a section div (h2 + desc + demo-card with children).
func sectionHeight(section widget.Widget) float32 {
	children := section.Children()
	if len(children) == 0 {
		return 36
	}
	h := float32(0)
	for _, child := range children {
		h += itemHeight(child) + 8
	}
	return h
}

// itemHeight estimates height for individual items within a section.
func itemHeight(child widget.Widget) float32 {
	switch v := child.(type) {
	case *widget.TextArea:
		return 100
	case *widget.Empty:
		return 80
	case *widget.Progress:
		return 12
	case *widget.Row:
		return 44 // grid row is a single horizontal line
	case *widget.Space:
		return 36 // space is a single horizontal line
	case *widget.Panel:
		// Panel has a 40px title header + content below
		h := float32(40 + 12) // header + padding
		for _, c := range v.Children() {
			h += itemHeight(c) + 8
		}
		return h
	case *widget.Menu:
		return v.TotalHeight()
	case *widget.Timeline:
		return v.TotalHeight()
	case *widget.Steps:
		return 80 // dot + title + description
	case *widget.Table:
		return v.TotalHeight()
	case *widget.List:
		return v.TotalHeight()
	case *widget.Collapse:
		// Each panel header is 40px + expanded content (60 + 8*2 spacing)
		h := float32(0)
		for _, p := range v.Panels() {
			h += 40
			if v.IsActive(p.Value) {
				h += 60 + 16 // contentH + SpaceSM*2
			}
		}
		if h == 0 {
			h = 40
		}
		return h
	case *widget.Card:
		pad := float32(24)
		headerH := float32(0)
		if v.Title() != "" {
			headerH = 48
		}
		h := headerH + pad*2 // header + top/bottom padding
		for _, c := range v.Children() {
			h += itemHeight(c) + 8
		}
		if h < 100 {
			h = 100
		}
		return h
	case *widget.InputNumber:
		return 32
	case *widget.Text:
		fs := v.FontSize()
		if fs <= 0 {
			fs = 14
		}
		return fs * 1.4 // line-height ~1.4x font size
	}
	// Check if this is a container (demo-card) with children
	children := child.Children()
	if len(children) > 0 {
		h := float32(16) // padding
		for _, c := range children {
			h += itemHeight(c) + 8
		}
		return h + 16
	}
	return 36
}

// layoutSection lays out children of a section (h2, span, demo-card, etc.) vertically.
func layoutSection(tree *core.Tree, section widget.Widget, x, y, w, h float32) {
	children := section.Children()
	if len(children) == 0 {
		return
	}
	cy := y
	for _, child := range children {
		ih := itemHeight(child)
		tree.SetLayout(child.ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(x, cy, w, ih),
		})
		// Recursively lay out children of containers (demo-card divs)
		if len(child.Children()) > 0 {
			layoutSectionContent(tree, child, x+12, cy+8, w-24, ih-16)
		}
		cy += ih + 8
	}
}

// layoutSectionContent lays out items within a demo card.
func layoutSectionContent(tree *core.Tree, parent widget.Widget, x, y, w, h float32) {
	children := parent.Children()
	if len(children) == 0 {
		return
	}
	cy := y
	for _, child := range children {
		ih := itemHeight(child)
		tree.SetLayout(child.ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(x, cy, w, ih),
		})
		// Handle horizontal layout widgets
		if _, ok := child.(*widget.Space); ok {
			layoutSpaceHorizontal(tree, child, x, cy, w, ih)
		} else if _, ok := child.(*widget.Row); ok {
			layoutRowCols(tree, child, x, cy, w, ih)
		} else if _, ok := child.(*widget.Panel); ok {
			// Panel has 40px title header; lay out children below it
			layoutSectionContent(tree, child, x+12, cy+40+4, w-24, ih-40-8)
		} else if card, ok := child.(*widget.Card); ok {
			// Card has header; lay out children below it
			pad := float32(24)
			headerH := float32(0)
			if card.Title() != "" {
				headerH = 48
			}
			layoutSectionContent(tree, child, x+pad, cy+headerH+pad, w-pad*2, ih-headerH-pad*2)
		} else if len(child.Children()) > 0 {
			layoutSectionContent(tree, child, x+4, cy+4, w-8, ih-8)
		}
		cy += ih + 8
	}
}

// layoutSpaceHorizontal lays out Space children horizontally.
func layoutSpaceHorizontal(tree *core.Tree, space widget.Widget, x, y, w, h float32) {
	children := space.Children()
	if len(children) == 0 {
		return
	}
	gap := float32(12)
	n := float32(len(children))
	totalGap := gap * (n - 1)
	itemW := (w - totalGap) / n
	if itemW > 120 {
		itemW = 120
	}
	cx := x
	for _, child := range children {
		tree.SetLayout(child.ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(cx, y, itemW, h),
		})
		fillDescendants(tree, child, cx+2, y+2, itemW-4, h-4)
		cx += itemW + gap
	}
}

// layoutRowCols lays out Row children (Cols) horizontally by span/24 ratio.
func layoutRowCols(tree *core.Tree, row widget.Widget, x, y, w, h float32) {
	children := row.Children()
	if len(children) == 0 {
		return
	}
	r, _ := row.(*widget.Row)
	gutter := float32(0)
	if r != nil {
		gutter = r.Gutter()
	}
	totalGutter := gutter * float32(len(children)-1)
	availW := w - totalGutter

	cx := x
	for _, child := range children {
		span := 24 // default full width
		if col, ok := child.(*widget.Col); ok {
			span = col.Span()
		}
		colW := availW * float32(span) / 24.0
		tree.SetLayout(child.ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(cx, y, colW, h),
		})
		fillDescendants(tree, child, cx+2, y+2, colW-4, h-4)
		cx += colW + gutter
	}
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

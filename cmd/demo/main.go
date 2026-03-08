//go:build windows

// Demo showcases the P0 widget library: Text, Button, Input, Layout, and theme.
// Run: go run ./cmd/demo
package main

import (
	"fmt"
	"os"
	"time"

	ui "github.com/kasuganosora/ui"
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

// mockEngine implements font.Engine with monospace glyphs (no CGO needed).
type mockEngine struct {
	glyphs map[rune]font.GlyphID
	nextG  font.GlyphID
}

func newMockEngine() *mockEngine {
	e := &mockEngine{glyphs: make(map[rune]font.GlyphID), nextG: 1}
	// ASCII
	for r := rune(32); r < 127; r++ {
		e.glyphs[r] = e.nextG
		e.nextG++
	}
	// CJK Unified Ideographs (common range)
	for r := rune(0x4E00); r <= rune(0x9FFF); r++ {
		e.glyphs[r] = e.nextG
		e.nextG++
	}
	// CJK punctuation
	for r := rune(0x3000); r <= rune(0x303F); r++ {
		e.glyphs[r] = e.nextG
		e.nextG++
	}
	// Fullwidth forms
	for r := rune(0xFF00); r <= rune(0xFFEF); r++ {
		e.glyphs[r] = e.nextG
		e.nextG++
	}
	// Ellipsis
	e.glyphs['…'] = e.nextG
	e.nextG++
	return e
}

func (e *mockEngine) LoadFont([]byte) (font.ID, error)        { return 1, nil }
func (e *mockEngine) LoadFontFile(string) (font.ID, error)     { return 1, nil }
func (e *mockEngine) UnloadFont(font.ID)                       {}
func (e *mockEngine) SetDPIScale(float32)                       {}
func (e *mockEngine) Destroy()                                 {}
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
		data[i] = 255 // solid white glyph
	}
	return font.GlyphBitmap{Width: w, Height: h, Data: data, SDF: sdf}, nil
}

func (e *mockEngine) HasGlyph(_ font.ID, r rune) bool {
	_, ok := e.glyphs[r]
	return ok
}

// textDrawerAdapter wraps textrender.Renderer to implement widget.TextDrawer.
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

// mouseDownTarget tracks which element received MouseDown for click synthesis.
var mouseDownTarget core.ElementID

// lastHoverTarget tracks the previously hovered element for MouseEnter/Leave synthesis.
var lastHoverTarget core.ElementID

// lastCursor tracks the current cursor shape to avoid redundant SetCursor calls.
var lastCursor = platform.CursorArrow

// contentWidget is the scrollable content area (set by buildUI).
var contentWidget *widget.Content

// rootWidget is the root layout widget (set in run()).
var rootWidget widget.Widget

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// --- Platform ---
	plat := win32.New()
	if err := plat.Init(); err != nil {
		return fmt.Errorf("platform init: %w", err)
	}
	defer plat.Terminate()

	win, err := plat.CreateWindow(platform.WindowOptions{
		Title:     "GoUI Demo — P0 Widget Showcase",
		Width:     960,
		Height:    640,
		Resizable: true,
		Visible:   true,
		Decorated: true,
	})
	if err != nil {
		return fmt.Errorf("create window: %w", err)
	}
	defer win.Destroy()

	// --- Renderer ---
	backend := vulkan.New()
	if err := backend.Init(win); err != nil {
		return fmt.Errorf("vulkan init: %w", err)
	}
	defer backend.Destroy()

	// --- Font System ---
	t0 := time.Now()
	var fontEngine font.Engine
	var fontID font.ID

	if ftEngine, err := freetype.New(); err == nil {
		fontEngine = ftEngine
		fmt.Printf("[font] FreeType engine loaded (%v)\n", time.Since(t0))
	} else {
		fontEngine = newMockEngine()
		fmt.Printf("[font] FreeType not available (%v), using mock engine\n", err)
	}
	mgr := font.NewManager(fontEngine)
	t1 := time.Now()
	fontID, _ = mgr.RegisterFile("Default", font.WeightRegular, font.StyleNormal, `C:\Windows\Fonts\msyh.ttc`)
	fmt.Printf("[font] RegisterFile took %v\n", time.Since(t1))
	if fontID == font.InvalidFontID {
		fontID, _ = mgr.Register("Default", font.WeightRegular, font.StyleNormal, nil)
	}
	dpi := backend.DPIScale()
	fontEngine.SetDPIScale(dpi)
	glyphAtlas := atlas.New(atlas.Options{Width: 1024, Height: 1024, Backend: backend})
	glyphAtlas.EnsureTexture() // Pre-create GPU texture so first frame text is visible
	textRenderer := textrender.New(textrender.Options{
		Manager:   mgr,
		Atlas:     glyphAtlas,
		DPIScale:  dpi,
		KeepAlive: plat.ProcessMessages,
	})

	defer textRenderer.Destroy()

	// --- Widget Tree ---
	tree := core.NewTree()
	dispatcher := core.NewDispatcher(tree)
	cfg := widget.DefaultConfig()
	cfg.TextRenderer = &textDrawerAdapter{renderer: textRenderer, fontID: fontID, engine: fontEngine}
	cfg.Window = win
	cfg.Platform = plat
	buf := render.NewCommandBuffer()
	_ = dispatcher // used for event routing below

	// -- Build UI --
	root := buildUI(tree, cfg)
	rootWidget = root

	// Attach root widget to tree root
	tree.AppendChild(tree.Root(), root.ElementID())

	// --- Main Loop ---
	var lastW, lastH int
	frameCount := 0
	fpsStart := time.Now()

	// w32 is the Win32-specific window handle (nil on other platforms).
	var w32 *win32.Window
	if ww, ok := win.(*win32.Window); ok {
		w32 = ww
	}

	// renderFrame performs one complete layout-check + render cycle.
	renderFrame := func() {
		fw, fh := win.FramebufferSize()
		if fw != lastW || fh != lastH {
			backend.Resize(fw, fh)
			lastW, lastH = fw, fh
			lw, lh := win.Size()
			tree.SetLayout(tree.Root(), core.LayoutResult{
				Bounds: uimath.NewRect(0, 0, float32(lw), float32(lh)),
			})
			computeLayout(tree, root, float32(lw), float32(lh))
		}

		// During the Win32 modal resize loop (user dragging border),
		// AcquireNextImageKHR blocks for seconds because the compositor
		// holds all swapchain images. Skip rendering and let the DWM
		// stretch the last frame; we render properly when drag ends.
		if w32 != nil && w32.InSizeMove() {
			return
		}

		backend.BeginFrame()

		textRenderer.BeginFrame()
		buf.Reset()
		root.Draw(buf)
		textRenderer.Upload()

		backend.Submit(buf)
		backend.EndFrame()
	}

	// Register resize callback so layout updates during modal resize.
	if w32 != nil {
		w32.OnResize(renderFrame)
	}

	needsRedraw := true // first frame always renders
	firstFrame := true

	for !win.ShouldClose() {
		// Poll events
		events := plat.PollEvents()
		if len(events) > 0 {
			needsRedraw = true
		}
		for i := range events {
			handleEvent(tree, dispatcher, &events[i], win)
		}

		// Check if tree was mutated (layout/paint dirty)
		if tree.NeedsRender() {
			needsRedraw = true
		}

		// Render only when something changed
		if needsRedraw {
			if firstFrame {
				t := time.Now()
				renderFrame()
				fmt.Printf("[perf] first frame took %v\n", time.Since(t))
				firstFrame = false
			} else {
				renderFrame()
			}
			tree.ClearAllDirty()
			needsRedraw = false
		}

		// FPS counter (console)
		frameCount++
		if elapsed := time.Since(fpsStart); elapsed >= time.Second {
			fmt.Printf("\rFPS: %d  ", frameCount)
			frameCount = 0
			fpsStart = time.Now()
		}

		// Small sleep to avoid burning CPU
		time.Sleep(time.Millisecond)
	}

	fmt.Println()
	return nil
}

// demoHTML is the HTML template for the demo UI.
const demoHTML = `
<style>
	/* CSS Variables */
	:root {
		--primary-bg: #001529;
		--aside-bg: #f7f8fa;
		--aside-border: #e8e8e8;
		--content-bg: #ffffff;
		--title-color: #1a1a1a;
		--col-text: #0050b3;
		--footer-text: #ffffff80;
	}

	/* Layout structure */
	layout { background-color: #f0f2f5; }
	header { background-color: var(--primary-bg); }
	footer { background-color: var(--primary-bg); }

	.title { color: white; font-size: 20px; }

	.body {
		display: flex;
		flex-direction: row;
		flex-grow: 1;
		background-color: white;
	}

	main { background-color: var(--content-bg); }

	/* Content sections */
	.section-title { font-size: 24px; color: var(--title-color); }

	/* Grid columns */
	.col-1 { background-color: #e6f7ff; border-radius: 4px; }
	.col-2 { background-color: #bae7ff; border-radius: 4px; }
	.col-3 { background-color: #91d5ff; border-radius: 4px; }
	.col-4 { background-color: #69c0ff; border-radius: 4px; }
	.col-text { color: var(--col-text); }

	/* Footer */
	.footer-text { color: var(--footer-text); font-size: 12px; }
</style>

<layout>
	<!-- Header -->
	<header height="56">
		<span class="title">组件演示平台</span>
	</header>

	<!-- Body: Aside + Content -->
	<div class="body">
		<aside width="220" style="background-color: #f7f8fa">
			<button id="menu-dashboard" variant="text">仪表盘</button>
			<button id="menu-components" variant="text">组件库</button>
			<button id="menu-settings" variant="text">系统设置</button>
			<button id="menu-about" variant="text">关于我们</button>
		</aside>

		<main id="content">
			<!-- Section: Title -->
			<span class="section-title">基础组件展示</span>

			<!-- Section: Buttons -->
			<space gap="12">
				<button id="btn-primary">主要按钮</button>
				<button id="btn-secondary" variant="secondary">次要按钮</button>
				<button id="btn-text" variant="text">文字按钮</button>
				<button id="btn-link" variant="link">链接按钮</button>
				<button disabled>禁用按钮</button>
			</space>

			<!-- Section: Input -->
			<span>输入框：</span>
			<input id="input-name" placeholder="请输入您的姓名..."/>
			<input placeholder="请输入电子邮箱..."/>
			<input value="已禁用的输入框" disabled/>

			<!-- Section: Grid -->
			<span>栅格布局（二十四列）：</span>
			<row gutter="16">
				<col span="6"><div class="col-1"><span class="col-text">第一列</span></div></col>
				<col span="6"><div class="col-2"><span class="col-text">第二列</span></div></col>
				<col span="6"><div class="col-3"><span class="col-text">第三列</span></div></col>
				<col span="6"><div class="col-4"><span class="col-text">第四列</span></div></col>
			</row>

			<!-- Section: Tooltip -->
			<button id="btn-tooltip" variant="secondary">悬停查看提示信息</button>
			<tooltip>这是一个工具提示！</tooltip>

			<!-- Section: Checkbox & Switch -->
			<span>复选框 &amp; 开关：</span>
			<space gap="16">
				<checkbox id="cb-a" checked>选项A</checkbox>
				<checkbox>选项B</checkbox>
				<checkbox disabled>禁用</checkbox>
				<switch id="sw-1" checked></switch>
				<switch disabled></switch>
			</space>

			<!-- Section: Radio -->
			<span>单选按钮：</span>
			<space gap="16">
				<radio group="fruit" checked>苹果</radio>
				<radio group="fruit">香蕉</radio>
				<radio group="fruit">橙子</radio>
			</space>

			<!-- Section: Tags -->
			<span>标签：</span>
			<space gap="8">
				<tag>默认</tag>
				<tag type="success">成功</tag>
				<tag type="warning">警告</tag>
				<tag type="error">错误</tag>
				<tag type="processing">处理中</tag>
			</space>

			<!-- Section: Progress -->
			<span>进度条：</span>
			<progress percent="65"></progress>

			<!-- Section: TextArea -->
			<span>多行输入框：</span>
			<textarea id="textarea" placeholder="请输入多行文本..." rows="3"></textarea>

			<!-- Section: Select -->
			<span>下拉选择：</span>
			<select id="city-select"></select>

			<!-- Section: Messages -->
			<span>消息通知：</span>
			<space gap="12">
				<message>普通消息</message>
				<message type="success">操作成功</message>
				<message type="warning">请注意</message>
				<message type="error">出错了</message>
			</space>

			<!-- Section: Empty -->
			<span>空状态：</span>
			<empty></empty>

			<!-- Section: Loading -->
			<span>加载中：</span>
			<loading tip="正在加载..."></loading>
		</main>
	</div>

	<!-- Footer -->
	<footer>
		<span class="footer-text">GoUI v0.1 — 零CGO跨平台界面库</span>
	</footer>
</layout>
`

// buildUI constructs the widget tree from HTML+CSS template.
func buildUI(tree *core.Tree, cfg *widget.Config) widget.Widget {
	doc := ui.LoadHTMLDocument(tree, cfg, demoHTML, "")

	// Get Content widget reference for scrolling
	if c, ok := doc.QueryByID("content").(*widget.Content); ok {
		contentWidget = c
	}

	// Wire up aside border (not expressible in CSS yet)
	if aside := doc.QueryByTag("aside"); len(aside) > 0 {
		if a, ok := aside[0].(*widget.Aside); ok {
			a.SetBorderRight(1, uimath.ColorHex("#e8e8e8"))
		}
	}

	// Wire up event handlers via QueryByID
	if btn, ok := doc.QueryByID("btn-primary").(*widget.Button); ok {
		btn.OnClick(func() { fmt.Println("[点击] 主要按钮") })
	}
	if btn, ok := doc.QueryByID("btn-secondary").(*widget.Button); ok {
		btn.OnClick(func() { fmt.Println("[点击] 次要按钮") })
	}
	if btn, ok := doc.QueryByID("btn-text").(*widget.Button); ok {
		btn.OnClick(func() { fmt.Println("[点击] 文字按钮") })
	}
	if btn, ok := doc.QueryByID("btn-link").(*widget.Button); ok {
		btn.OnClick(func() { fmt.Println("[点击] 链接按钮") })
	}
	if inp, ok := doc.QueryByID("input-name").(*widget.Input); ok {
		inp.OnChange(func(v string) { fmt.Printf("[输入] 姓名 = %q\n", v) })
	}
	if cb, ok := doc.QueryByID("cb-a").(*widget.Checkbox); ok {
		cb.OnChange(func(v bool) { fmt.Printf("[复选] A = %v\n", v) })
	}
	if sw, ok := doc.QueryByID("sw-1").(*widget.Switch); ok {
		sw.OnChange(func(v bool) { fmt.Printf("[开关] = %v\n", v) })
	}
	if ta, ok := doc.QueryByID("textarea").(*widget.TextArea); ok {
		ta.OnChange(func(v string) { fmt.Printf("[多行] len=%d\n", len(v)) })
	}

	// Select needs options set programmatically
	if sel, ok := doc.QueryByID("city-select").(*widget.Select); ok {
		sel.SetOptions([]widget.SelectOption{
			{Label: "北京", Value: "beijing"},
			{Label: "上海", Value: "shanghai"},
			{Label: "广州", Value: "guangzhou"},
			{Label: "深圳（禁用）", Value: "shenzhen", Disabled: true},
		})
		sel.SetValue("beijing")
		sel.OnChange(func(v string) { fmt.Printf("[选择] = %s\n", v) })
	}

	// The root is a wrapper div; return its first child (the layout)
	if len(doc.Root.Children()) > 0 {
		return doc.Root.Children()[0]
	}
	return doc.Root
}

// computeLayout performs a simplified layout pass for the demo.
// In production this would use the layout.Engine; here we do manual positioning
// to demonstrate the widgets rendering without requiring a full layout integration.
func computeLayout(tree *core.Tree, root widget.Widget, w, h float32) {
	// Root fills the window
	tree.SetLayout(root.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, w, h),
	})

	children := root.Children()
	if len(children) < 3 {
		return
	}

	headerH := float32(56)
	footerH := float32(48)
	bodyH := h - headerH - footerH

	// Header
	tree.SetLayout(children[0].ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, w, headerH),
	})
	layoutChildrenHorizontal(tree, children[0], 0, 0, w, headerH, 24)

	// Body
	body := children[1]
	tree.SetLayout(body.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, headerH, w, bodyH),
	})

	bodyChildren := body.Children()
	if len(bodyChildren) >= 2 {
		asideW := float32(220)
		contentW := w - asideW

		// Aside
		tree.SetLayout(bodyChildren[0].ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(0, headerH, asideW, bodyH),
		})
		layoutChildrenVertical(tree, bodyChildren[0], 0, headerH, asideW, bodyH, 8)

		// Content
		contentX := asideW
		tree.SetLayout(bodyChildren[1].ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(contentX, headerH, contentW, bodyH),
		})
		layoutContentArea(tree, bodyChildren[1], contentX, headerH, contentW, bodyH)
	}

	// Footer
	tree.SetLayout(children[2].ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, headerH+bodyH, w, footerH),
	})
	layoutChildrenHorizontal(tree, children[2], 0, headerH+bodyH, w, footerH, 0)
}

func layoutContentArea(tree *core.Tree, content widget.Widget, x, y, w, h float32) {
	padding := float32(24)
	cx := x + padding
	cw := w - padding*2
	gap := float32(12)

	// First pass: calculate total content height
	totalH := padding
	for _, child := range content.Children() {
		rowH := contentRowHeight(child)
		totalH += rowH + gap
	}
	totalH += padding // bottom padding

	// Set content height on the Content widget for scrollbar
	if c, ok := content.(*widget.Content); ok {
		c.SetContentHeight(totalH)
		c.ScrollBy(0) // re-clamp after layout change
	}

	// Get scroll offset
	scrollY := float32(0)
	if c, ok := content.(*widget.Content); ok {
		scrollY = c.ScrollY()
	}

	cy := y + padding - scrollY

	for _, child := range content.Children() {
		rowH := contentRowHeight(child)

		tree.SetLayout(child.ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(cx, cy, cw, rowH),
		})

		// Layout grandchildren for Space/Row/etc.
		grandChildren := child.Children()
		if len(grandChildren) > 0 {
			layoutChildrenHorizontal(tree, child, cx, cy, cw, rowH, 12)
		}

		cy += rowH + gap
	}
}

func contentRowHeight(child widget.Widget) float32 {
	if _, ok := child.(*widget.TextArea); ok {
		return 80
	}
	if _, ok := child.(*widget.Empty); ok {
		return 80
	}
	if _, ok := child.(*widget.Progress); ok {
		return 12
	}
	return 36
}

func layoutChildrenVertical(tree *core.Tree, parent widget.Widget, x, y, w, h, gap float32) {
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

func layoutChildrenHorizontal(tree *core.Tree, parent widget.Widget, x, y, w, h, gap float32) {
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
		// Recursively fill children with inset bounds
		layoutDescendantsFill(tree, child, cx+4, y+4, itemW-8, h-8)
		cx += itemW + gap
	}
}

// layoutDescendantsFill recursively assigns inset bounds to all descendants.
func layoutDescendantsFill(tree *core.Tree, parent widget.Widget, x, y, w, h float32) {
	for _, child := range parent.Children() {
		tree.SetLayout(child.ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(x, y, w, h),
		})
		layoutDescendantsFill(tree, child, x+2, y+2, w-4, h-4)
	}
}

// handleEvent routes platform events to the element tree.
func handleEvent(tree *core.Tree, dispatcher *core.Dispatcher, evt *event.Event, win platform.Window) {
	switch evt.Type {
	case event.WindowResize:
		// Handled in main loop
	case event.WindowClose:
		win.SetShouldClose(true)
	case event.MouseWheel:
		// Route scroll events to content area
		if contentWidget != nil && rootWidget != nil {
			contentWidget.HandleWheel(evt.WheelDY)
			// Re-layout after scroll (use logical size)
			lw, lh := win.Size()
			computeLayout(tree, rootWidget, float32(lw), float32(lh))
		}
	case event.MouseMove, event.MouseDown, event.MouseUp, event.MouseClick:
		// Handle scrollbar drag
		if contentWidget != nil {
			if contentWidget.IsScrollBarDragging() {
				if evt.Type == event.MouseMove {
					contentWidget.HandleScrollBarMove(evt.GlobalY)
					lw, lh := win.Size()
					computeLayout(tree, rootWidget, float32(lw), float32(lh))
					return
				}
				if evt.Type == event.MouseUp {
					contentWidget.HandleScrollBarUp()
					return
				}
			}
			// Check if clicking on scrollbar thumb
			if evt.Type == event.MouseDown && contentWidget.HandleScrollBarDown(evt.GlobalY) {
				// Check X is in scrollbar region
				bounds := contentWidget.Bounds()
				scrollBarX := bounds.X + bounds.Width - 10
				if evt.GlobalX >= scrollBarX {
					return
				}
				contentWidget.HandleScrollBarUp() // not in scrollbar X range
			}
		}

		target := tree.HitTest(evt.GlobalX, evt.GlobalY)

		// During a drag (mouseDownTarget active), also send MouseMove/MouseUp
		// to the element that received MouseDown, for drag-selection support.
		if mouseDownTarget != core.InvalidElementID && target != mouseDownTarget {
			if evt.Type == event.MouseMove || evt.Type == event.MouseUp {
				dispatcher.Dispatch(mouseDownTarget, evt)
			}
		}

		// Update hover state and cursor shape on mouse move
		if evt.Type == event.MouseMove {
			tree.Walk(tree.Root(), func(id core.ElementID, _ int) bool {
				tree.SetHovered(id, id == target)
				return true
			})
			if target != lastHoverTarget {
				if lastHoverTarget != core.InvalidElementID {
					dispatcher.Dispatch(lastHoverTarget, &event.Event{Type: event.MouseLeave})
				}
				if target != core.InvalidElementID {
					dispatcher.Dispatch(target, &event.Event{Type: event.MouseEnter})
				}
				lastHoverTarget = target
			}
			// Update cursor shape based on element type
			wantCursor := platform.CursorArrow
			for id := target; id != core.InvalidElementID; {
				if e := tree.Get(id); e != nil {
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
			if wantCursor != lastCursor {
				win.SetCursor(wantCursor)
				lastCursor = wantCursor
			}
		}

		if target != core.InvalidElementID {
			// On MouseDown, blur focused element if clicking outside any input
			if evt.Type == event.MouseDown {
				isInput := false
				for id := target; id != core.InvalidElementID; {
					if e := tree.Get(id); e != nil {
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
					tree.ClearFocus()
				}
			}

			dispatcher.Dispatch(target, evt)

			// Synthesize MouseClick from MouseDown+MouseUp on same element
			if evt.Type == event.MouseDown {
				mouseDownTarget = target
			} else if evt.Type == event.MouseUp && target == mouseDownTarget {
				clickEvt := *evt
				clickEvt.Type = event.MouseClick
				dispatcher.Dispatch(target, &clickEvt)
				mouseDownTarget = core.InvalidElementID
			}
		}

		if evt.Type == event.MouseUp {
			mouseDownTarget = core.InvalidElementID
		}
	case event.KeyDown, event.KeyUp, event.KeyPress,
		event.IMECompositionStart, event.IMECompositionUpdate, event.IMECompositionEnd:
		// Send to focused element
		tree.Walk(tree.Root(), func(id core.ElementID, _ int) bool {
			if e := tree.Get(id); e != nil && e.IsFocused() {
				dispatcher.Dispatch(id, evt)
				return false
			}
			return true
		})
	}
}


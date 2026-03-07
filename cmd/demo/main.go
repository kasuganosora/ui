//go:build windows

// Demo showcases the P0 widget library: Text, Button, Input, Layout, and theme.
// Run: go run ./cmd/demo
package main

import (
	"fmt"
	"os"
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
	var fontEngine font.Engine
	var fontID font.ID

	if ftEngine, err := freetype.New(); err == nil {
		fontEngine = ftEngine
		fmt.Println("[font] FreeType engine loaded")
	} else {
		fontEngine = newMockEngine()
		fmt.Printf("[font] FreeType not available (%v), using mock engine\n", err)
	}
	mgr := font.NewManager(fontEngine)
	fontID, _ = mgr.RegisterFile("Default", font.WeightRegular, font.StyleNormal, `C:\Windows\Fonts\msyh.ttc`)
	if fontID == font.InvalidFontID {
		fontID, _ = mgr.Register("Default", font.WeightRegular, font.StyleNormal, nil)
	}
	dpi := backend.DPIScale()
	fontEngine.SetDPIScale(dpi)
	glyphAtlas := atlas.New(atlas.Options{Width: 1024, Height: 1024, Backend: backend})
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

	for !win.ShouldClose() {
		// Poll events
		events := plat.PollEvents()
		for i := range events {
			handleEvent(tree, dispatcher, &events[i], win)
		}

		// Check resize
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

		// Render
		backend.BeginFrame()
		textRenderer.BeginFrame()
		buf.Reset()

		root.Draw(buf)

		textRenderer.Upload()
		backend.Submit(buf)
		backend.EndFrame()

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

// buildUI constructs the widget tree for the demo.
func buildUI(tree *core.Tree, cfg *widget.Config) widget.Widget {
	// Root layout
	root := widget.NewLayout(tree, cfg)
	root.SetBgColor(uimath.ColorHex("#f0f2f5"))

	// -- Header --
	header := widget.NewHeader(tree, cfg)
	header.SetBgColor(uimath.ColorHex("#001529"))
	header.SetHeight(56)

	title := widget.NewText(tree, "组件演示平台", cfg)
	title.SetColor(uimath.ColorWhite)
	title.SetFontSize(20)
	header.AppendChild(title)

	root.AppendChild(header)

	// -- Body: Aside + Content --
	body := widget.NewDiv(tree, cfg)
	body.SetBgColor(uimath.ColorWhite) // covers the gray root bg to avoid SDF anti-alias seams
	bodyStyle := body.Style()
	bodyStyle.FlexGrow = 1
	bodyStyle.FlexDirection = 1 // Row
	body.SetStyle(bodyStyle)

	// Aside
	aside := widget.NewAside(tree, cfg)
	aside.SetBgColor(uimath.ColorHex("#f7f8fa"))
	aside.SetBorderRight(1, uimath.ColorHex("#e8e8e8"))
	aside.SetWidth(220)

	menuItems := []string{"仪表盘", "组件库", "系统设置", "关于我们"}
	for _, label := range menuItems {
		item := widget.NewButton(tree, label, cfg)
		item.SetVariant(widget.ButtonText)
		aside.AppendChild(item)
	}
	body.AppendChild(aside)

	// Content
	content := widget.NewContent(tree, cfg)
	content.SetBgColor(uimath.ColorHex("#ffffff"))
	contentWidget = content

	// Section: Title
	sectionTitle := widget.NewText(tree, "基础组件展示", cfg)
	sectionTitle.SetFontSize(24)
	sectionTitle.SetColor(uimath.ColorHex("#1a1a1a"))
	content.AppendChild(sectionTitle)

	// Section: Buttons
	btnRow := widget.NewSpace(tree, cfg)
	btnRow.SetGap(12)

	primaryBtn := widget.NewButton(tree, "主要按钮", cfg)
	primaryBtn.OnClick(func() { fmt.Println("[点击] 主要按钮") })
	btnRow.AppendChild(primaryBtn)

	secondaryBtn := widget.NewButton(tree, "次要按钮", cfg)
	secondaryBtn.SetVariant(widget.ButtonSecondary)
	secondaryBtn.OnClick(func() { fmt.Println("[点击] 次要按钮") })
	btnRow.AppendChild(secondaryBtn)

	textBtn := widget.NewButton(tree, "文字按钮", cfg)
	textBtn.SetVariant(widget.ButtonText)
	textBtn.OnClick(func() { fmt.Println("[点击] 文字按钮") })
	btnRow.AppendChild(textBtn)

	linkBtn := widget.NewButton(tree, "链接按钮", cfg)
	linkBtn.SetVariant(widget.ButtonLink)
	linkBtn.OnClick(func() { fmt.Println("[点击] 链接按钮") })
	btnRow.AppendChild(linkBtn)

	disabledBtn := widget.NewButton(tree, "禁用按钮", cfg)
	disabledBtn.SetDisabled(true)
	btnRow.AppendChild(disabledBtn)

	content.AppendChild(btnRow)

	// Section: Input
	inputLabel := widget.NewText(tree, "输入框：", cfg)
	content.AppendChild(inputLabel)

	nameInput := widget.NewInput(tree, cfg)
	nameInput.SetPlaceholder("请输入您的姓名...")
	nameInput.OnChange(func(v string) { fmt.Printf("[输入] 姓名 = %q\n", v) })
	content.AppendChild(nameInput)

	emailInput := widget.NewInput(tree, cfg)
	emailInput.SetPlaceholder("请输入电子邮箱...")
	content.AppendChild(emailInput)

	disabledInput := widget.NewInput(tree, cfg)
	disabledInput.SetValue("已禁用的输入框")
	disabledInput.SetDisabled(true)
	content.AppendChild(disabledInput)

	// Section: Grid
	gridLabel := widget.NewText(tree, "栅格布局（二十四列）：", cfg)
	content.AppendChild(gridLabel)

	row := widget.NewRow(tree, cfg)
	row.SetGutter(16)
	colors := []string{"#e6f7ff", "#bae7ff", "#91d5ff", "#69c0ff"}
	spans := []int{6, 6, 6, 6}
	colNames := []string{"第一列", "第二列", "第三列", "第四列"}
	for i, s := range spans {
		col := widget.NewCol(tree, s, cfg)
		colContent := widget.NewDiv(tree, cfg)
		colContent.SetBgColor(uimath.ColorHex(colors[i]))
		colContent.SetBorderRadius(4)
		colTxt := widget.NewText(tree, colNames[i], cfg)
		colTxt.SetColor(uimath.ColorHex("#0050b3"))
		colContent.AppendChild(colTxt)
		col.AppendChild(colContent)
		row.AppendChild(col)
	}
	content.AppendChild(row)

	// Section: Tooltip demo
	tooltipBtn := widget.NewButton(tree, "悬停查看提示信息", cfg)
	tooltipBtn.SetVariant(widget.ButtonSecondary)
	widget.NewTooltip(tree, "这是一个工具提示！", tooltipBtn.ElementID(), cfg)
	content.AppendChild(tooltipBtn)

	// Section: Checkbox & Switch
	checkLabel := widget.NewText(tree, "复选框 & 开关：", cfg)
	content.AppendChild(checkLabel)

	checkRow := widget.NewSpace(tree, cfg)
	checkRow.SetGap(16)

	cb1 := widget.NewCheckbox(tree, "选项A", cfg)
	cb1.SetChecked(true)
	cb1.OnChange(func(v bool) { fmt.Printf("[复选] A = %v\n", v) })
	checkRow.AppendChild(cb1)

	cb2 := widget.NewCheckbox(tree, "选项B", cfg)
	checkRow.AppendChild(cb2)

	cbDisabled := widget.NewCheckbox(tree, "禁用", cfg)
	cbDisabled.SetDisabled(true)
	checkRow.AppendChild(cbDisabled)

	sw1 := widget.NewSwitch(tree, cfg)
	sw1.SetChecked(true)
	sw1.OnChange(func(v bool) { fmt.Printf("[开关] = %v\n", v) })
	checkRow.AppendChild(sw1)

	sw2 := widget.NewSwitch(tree, cfg)
	sw2.SetDisabled(true)
	checkRow.AppendChild(sw2)

	content.AppendChild(checkRow)

	// Section: Radio
	radioLabel := widget.NewText(tree, "单选按钮：", cfg)
	content.AppendChild(radioLabel)

	radioRow := widget.NewSpace(tree, cfg)
	radioRow.SetGap(16)

	rg := widget.NewRadioGroup()
	r1 := widget.NewRadio(tree, "苹果", cfg)
	r1.SetChecked(true)
	rg.Add(r1)
	radioRow.AppendChild(r1)

	r2 := widget.NewRadio(tree, "香蕉", cfg)
	rg.Add(r2)
	radioRow.AppendChild(r2)

	r3 := widget.NewRadio(tree, "橙子", cfg)
	rg.Add(r3)
	radioRow.AppendChild(r3)

	content.AppendChild(radioRow)

	// Section: Tags
	tagLabel := widget.NewText(tree, "标签：", cfg)
	content.AppendChild(tagLabel)

	tagRow := widget.NewSpace(tree, cfg)
	tagRow.SetGap(8)

	tag1 := widget.NewTag(tree, "默认", cfg)
	tagRow.AppendChild(tag1)

	tag2 := widget.NewTag(tree, "成功", cfg)
	tag2.SetTagType(widget.TagSuccess)
	tagRow.AppendChild(tag2)

	tag3 := widget.NewTag(tree, "警告", cfg)
	tag3.SetTagType(widget.TagWarning)
	tagRow.AppendChild(tag3)

	tag4 := widget.NewTag(tree, "错误", cfg)
	tag4.SetTagType(widget.TagError)
	tagRow.AppendChild(tag4)

	tag5 := widget.NewTag(tree, "处理中", cfg)
	tag5.SetTagType(widget.TagProcessing)
	tagRow.AppendChild(tag5)

	content.AppendChild(tagRow)

	// Section: Progress
	progressLabel := widget.NewText(tree, "进度条：", cfg)
	content.AppendChild(progressLabel)

	prog := widget.NewProgress(tree, cfg)
	prog.SetPercent(65)
	content.AppendChild(prog)

	// Section: TextArea
	taLabel := widget.NewText(tree, "多行输入框：", cfg)
	content.AppendChild(taLabel)

	ta := widget.NewTextArea(tree, cfg)
	ta.SetPlaceholder("请输入多行文本...")
	ta.SetRows(3)
	ta.OnChange(func(v string) { fmt.Printf("[多行] len=%d\n", len(v)) })
	content.AppendChild(ta)

	// Section: Select
	selLabel := widget.NewText(tree, "下拉选择：", cfg)
	content.AppendChild(selLabel)

	sel := widget.NewSelect(tree, []widget.SelectOption{
		{Label: "北京", Value: "beijing"},
		{Label: "上海", Value: "shanghai"},
		{Label: "广州", Value: "guangzhou"},
		{Label: "深圳（禁用）", Value: "shenzhen", Disabled: true},
	}, cfg)
	sel.SetValue("beijing")
	sel.OnChange(func(v string) { fmt.Printf("[选择] = %s\n", v) })
	content.AppendChild(sel)

	// Section: Message
	msgLabel := widget.NewText(tree, "消息通知：", cfg)
	content.AppendChild(msgLabel)

	msgRow := widget.NewSpace(tree, cfg)
	msgRow.SetGap(12)

	msgInfo := widget.NewMessage(tree, "普通消息", cfg)
	msgRow.AppendChild(msgInfo)

	msgSuccess := widget.NewMessage(tree, "操作成功", cfg)
	msgSuccess.SetMsgType(widget.MessageSuccess)
	msgRow.AppendChild(msgSuccess)

	msgWarn := widget.NewMessage(tree, "请注意", cfg)
	msgWarn.SetMsgType(widget.MessageWarning)
	msgRow.AppendChild(msgWarn)

	msgErr := widget.NewMessage(tree, "出错了", cfg)
	msgErr.SetMsgType(widget.MessageError)
	msgRow.AppendChild(msgErr)

	content.AppendChild(msgRow)

	// Section: Empty state
	emptyLabel := widget.NewText(tree, "空状态：", cfg)
	content.AppendChild(emptyLabel)

	empty := widget.NewEmpty(tree, cfg)
	content.AppendChild(empty)

	// Section: Loading
	loadLabel := widget.NewText(tree, "加载中：", cfg)
	content.AppendChild(loadLabel)

	loading := widget.NewLoading(tree, cfg)
	loading.SetTip("正在加载...")
	content.AppendChild(loading)

	body.AppendChild(content)
	root.AppendChild(body)

	// -- Footer --
	footer := widget.NewFooter(tree, cfg)
	footer.SetBgColor(uimath.ColorHex("#001529"))
	footerText := widget.NewText(tree, "GoUI v0.1 — 零CGO跨平台界面库", cfg)
	footerText.SetColor(uimath.ColorHex("#ffffff80"))
	footerText.SetFontSize(12)
	footer.AppendChild(footerText)
	root.AppendChild(footer)

	return root
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

		if target != core.InvalidElementID {
			// Update hover state
			if evt.Type == event.MouseMove {
				tree.Walk(tree.Root(), func(id core.ElementID, _ int) bool {
					tree.SetHovered(id, id == target)
					return true
				})
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


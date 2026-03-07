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
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/platform/win32"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/render/vulkan"
	"github.com/kasuganosora/ui/widget"
)

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

	// --- Widget Tree ---
	tree := core.NewTree()
	dispatcher := core.NewDispatcher(tree)
	cfg := widget.DefaultConfig()
	buf := render.NewCommandBuffer()
	_ = dispatcher // used for event routing below

	// -- Build UI --
	root := buildUI(tree, cfg)

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
		w, h := win.FramebufferSize()
		if w != lastW || h != lastH {
			backend.Resize(w, h)
			lastW, lastH = w, h
			// Update root layout to fill window
			tree.SetLayout(tree.Root(), core.LayoutResult{
				Bounds: uimath.NewRect(0, 0, float32(w), float32(h)),
			})
			computeLayout(tree, root, float32(w), float32(h))
		}

		// Render
		backend.BeginFrame()
		buf.Reset()

		root.Draw(buf)

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

	title := widget.NewText(tree, "GoUI Demo", cfg)
	title.SetColor(uimath.ColorWhite)
	title.SetFontSize(20)
	header.AppendChild(title)

	root.AppendChild(header)

	// -- Body: Aside + Content --
	body := widget.NewDiv(tree, cfg)
	bodyStyle := body.Style()
	bodyStyle.FlexGrow = 1
	bodyStyle.FlexDirection = 1 // Row
	body.SetStyle(bodyStyle)

	// Aside
	aside := widget.NewAside(tree, cfg)
	aside.SetBgColor(uimath.ColorWhite)
	aside.SetWidth(220)

	menuItems := []string{"Dashboard", "Components", "Settings", "About"}
	for _, label := range menuItems {
		item := widget.NewButton(tree, label, cfg)
		item.SetVariant(widget.ButtonText)
		aside.AppendChild(item)
	}
	body.AppendChild(aside)

	// Content
	content := widget.NewContent(tree, cfg)
	content.SetBgColor(uimath.ColorHex("#ffffff"))

	// Section: Title
	sectionTitle := widget.NewText(tree, "P0 Widget Showcase", cfg)
	sectionTitle.SetFontSize(24)
	sectionTitle.SetColor(uimath.ColorHex("#1a1a1a"))
	content.AppendChild(sectionTitle)

	// Section: Buttons
	btnRow := widget.NewSpace(tree, cfg)
	btnRow.SetGap(12)

	primaryBtn := widget.NewButton(tree, "Primary", cfg)
	primaryBtn.OnClick(func() { fmt.Println("[click] Primary button") })
	btnRow.AppendChild(primaryBtn)

	secondaryBtn := widget.NewButton(tree, "Secondary", cfg)
	secondaryBtn.SetVariant(widget.ButtonSecondary)
	secondaryBtn.OnClick(func() { fmt.Println("[click] Secondary button") })
	btnRow.AppendChild(secondaryBtn)

	textBtn := widget.NewButton(tree, "Text Button", cfg)
	textBtn.SetVariant(widget.ButtonText)
	textBtn.OnClick(func() { fmt.Println("[click] Text button") })
	btnRow.AppendChild(textBtn)

	linkBtn := widget.NewButton(tree, "Link", cfg)
	linkBtn.SetVariant(widget.ButtonLink)
	linkBtn.OnClick(func() { fmt.Println("[click] Link button") })
	btnRow.AppendChild(linkBtn)

	disabledBtn := widget.NewButton(tree, "Disabled", cfg)
	disabledBtn.SetDisabled(true)
	btnRow.AppendChild(disabledBtn)

	content.AppendChild(btnRow)

	// Section: Input
	inputLabel := widget.NewText(tree, "Input:", cfg)
	content.AppendChild(inputLabel)

	nameInput := widget.NewInput(tree, cfg)
	nameInput.SetPlaceholder("Enter your name...")
	nameInput.OnChange(func(v string) { fmt.Printf("[input] name = %q\n", v) })
	content.AppendChild(nameInput)

	emailInput := widget.NewInput(tree, cfg)
	emailInput.SetPlaceholder("Enter your email...")
	content.AppendChild(emailInput)

	disabledInput := widget.NewInput(tree, cfg)
	disabledInput.SetValue("Disabled input")
	disabledInput.SetDisabled(true)
	content.AppendChild(disabledInput)

	// Section: Grid
	gridLabel := widget.NewText(tree, "Grid (24-column):", cfg)
	content.AppendChild(gridLabel)

	row := widget.NewRow(tree, cfg)
	row.SetGutter(16)
	colors := []string{"#e6f7ff", "#bae7ff", "#91d5ff", "#69c0ff"}
	spans := []int{6, 6, 6, 6}
	for i, s := range spans {
		col := widget.NewCol(tree, s, cfg)
		colContent := widget.NewDiv(tree, cfg)
		colContent.SetBgColor(uimath.ColorHex(colors[i]))
		colContent.SetBorderRadius(4)
		colTxt := widget.NewText(tree, fmt.Sprintf("Col-%d", s), cfg)
		colTxt.SetColor(uimath.ColorHex("#0050b3"))
		colContent.AppendChild(colTxt)
		col.AppendChild(colContent)
		row.AppendChild(col)
	}
	content.AppendChild(row)

	// Section: Tooltip demo
	tooltipBtn := widget.NewButton(tree, "Hover me for tooltip", cfg)
	tooltipBtn.SetVariant(widget.ButtonSecondary)
	widget.NewTooltip(tree, "This is a tooltip!", tooltipBtn.ElementID(), cfg)
	content.AppendChild(tooltipBtn)

	body.AppendChild(content)
	root.AppendChild(body)

	// -- Footer --
	footer := widget.NewFooter(tree, cfg)
	footer.SetBgColor(uimath.ColorHex("#001529"))
	footerText := widget.NewText(tree, "GoUI v0.1 — Zero-CGO Cross-Platform UI Library", cfg)
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
	cy := y + padding
	cw := w - padding*2
	rowH := float32(36)
	gap := float32(16)

	for _, child := range content.Children() {
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
	case event.MouseMove, event.MouseDown, event.MouseUp, event.MouseClick:
		target := tree.HitTest(evt.GlobalX, evt.GlobalY)
		if target != core.InvalidElementID {
			// Update hover state
			if evt.Type == event.MouseMove {
				tree.Walk(tree.Root(), func(id core.ElementID, _ int) bool {
					tree.SetHovered(id, id == target)
					return true
				})
			}
			dispatcher.Dispatch(target, evt)
		}
	case event.KeyDown, event.KeyUp, event.KeyPress:
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

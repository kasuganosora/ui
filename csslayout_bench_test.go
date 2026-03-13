package ui

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	"github.com/kasuganosora/ui/widget"
)

// buildBenchTree creates a widget tree matching the game-demo complexity:
// rootDiv → 20 absolute-positioned windows, each with 3–5 child widgets.
// ~110 total widget nodes.
func buildBenchTree(t testing.TB) (*core.Tree, widget.Widget) {
	t.Helper()
	tree := core.NewTree()
	cfg := widget.DefaultConfig()

	root := widget.NewDiv(tree, cfg)
	root.SetStyle(layout.Style{
		Display: layout.DisplayFlex,
		Width:   layout.Px(1280),
		Height:  layout.Px(800),
	})
	tree.AppendChild(tree.Root(), root.ElementID())

	addWindow := func(x, y, w, h float32, children int) {
		win := widget.NewDiv(tree, cfg)
		win.SetStyle(layout.Style{
			Position: layout.PositionAbsolute,
			Left:     layout.Px(x),
			Top:      layout.Px(y),
			Width:    layout.Px(w),
			Height:   layout.Px(h),
			Display:  layout.DisplayFlex,
			FlexDirection: layout.FlexDirectionColumn,
		})
		for i := 0; i < children; i++ {
			child := widget.NewDiv(tree, cfg)
			child.SetStyle(layout.Style{
				Width:  layout.Px(w - 8),
				Height: layout.Px(float32(18 + i*4)),
			})
			win.AppendChild(child)
		}
		root.AppendChild(win)
	}

	// Simulate game-demo windows
	addWindow(20, 20, 220, 83, 4)
	addWindow(20, 108, 220, 61, 2)
	addWindow(20, 178, 170, 188, 4)
	addWindow(360, 728, 560, 52, 1)
	addWindow(500, 696, 280, 22, 1)
	addWindow(1100, 20, 160, 160, 2)
	addWindow(1030, 200, 230, 200, 4)
	addWindow(490, 8, 300, 24, 3)
	addWindow(540, 36, 200, 30, 2)
	addWindow(390, 68, 220, 52, 2)
	addWindow(480, 350, 100, 30, 2)
	addWindow(560, 420, 90, 26, 2)
	addWindow(400, 440, 90, 26, 2)
	addWindow(10, 520, 340, 228, 3)
	addWindow(980, 250, 287, 247, 3)
	addWindow(20, 346, 320, 228, 3)
	addWindow(270, 216, 360, 248, 5)
	addWindow(640, 256, 200, 208, 3)
	addWindow(400, 352, 480, 208, 4)
	addWindow(640, 100, 200, 80, 3)

	return tree, root
}

// BenchmarkCSSLayoutFull measures a full layout (tree rebuild + compute + apply).
// Represents startup or bringToFront (NeedsLayout=true).
func BenchmarkCSSLayoutFull(b *testing.B) {
	tree, root := buildBenchTree(b)
	cfg := widget.DefaultConfig()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate full layout: create a fresh cache every iteration
		lc := NewCSSLayoutCache()
		lc.Layout(tree, root, 1280, 800, cfg)
	}
}

// BenchmarkCSSLayoutFast measures the fast path (structure unchanged, positions only).
// Represents every frame during drag (NeedsLayout=false, cache valid).
func BenchmarkCSSLayoutFast(b *testing.B) {
	tree, root := buildBenchTree(b)
	cfg := widget.DefaultConfig()

	lc := NewCSSLayoutCache()
	lc.Layout(tree, root, 1280, 800, cfg) // warm up cache
	tree.ClearAllDirty()                   // simulate post-frame cleanup; cache stays valid

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Fast path: NeedsLayout=false, cache valid — only applyLayoutResults runs.
		// After each call, ClearAllDirty so dirty bits don't accumulate.
		lc.Layout(tree, root, 1280, 800, cfg)
		tree.ClearAllDirty()
	}
}

// BenchmarkCSSLayoutFullRepeated simulates bringToFront:
// every iteration marks DirtyLayout then re-layouts (triggers full recompute).
func BenchmarkCSSLayoutFullRepeated(b *testing.B) {
	tree, root := buildBenchTree(b)
	cfg := widget.DefaultConfig()
	lc := NewCSSLayoutCache()
	lc.Layout(tree, root, 1280, 800, cfg)
	tree.ClearAllDirty()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate what bringToFront does: RemoveChild+AppendChild marks DirtyLayout
		lc.Invalidate()
		lc.Layout(tree, root, 1280, 800, cfg)
		tree.ClearAllDirty()
	}
}

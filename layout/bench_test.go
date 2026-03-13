package layout

import "testing"

// buildGameDemoTree builds a layout tree representative of the game HUD demo:
// - 1 root (flex, fills viewport)
// - 20 absolute-positioned windows
// - Each window has 3–5 children (flex column, mixed fixed/auto heights)
// Total ~110 nodes.
func buildGameDemoTree(e *Engine) NodeID {
	root := e.AddNode(Style{
		Display:  DisplayFlex,
		Width:    Px(1280),
		Height:   Px(800),
		Position: PositionRelative,
	})

	addWindow := func(x, y, w, h float32, childCount int) NodeID {
		win := e.AddNode(Style{
			Position: PositionAbsolute,
			Left:     Px(x),
			Top:      Px(y),
			Width:    Px(w),
			Height:   Px(h),
			Display:  DisplayFlex,
			FlexDirection: FlexDirectionColumn,
			Padding:  EdgeValues{Top: Px(4), Bottom: Px(4), Left: Px(4), Right: Px(4)},
		})
		var childIDs []NodeID
		for i := 0; i < childCount; i++ {
			childH := float32(20 + i*4)
			child := e.AddNode(Style{
				Width:  Px(w - 8),
				Height: Px(childH),
				Margin: EdgeValues{Bottom: Px(4)},
			})
			childIDs = append(childIDs, child)
		}
		e.SetChildren(win, childIDs)
		return win
	}

	windows := []NodeID{
		// Status panel + bars
		addWindow(20, 20, 220, 83, 4),
		addWindow(20, 108, 220, 61, 2),
		// Team frames
		addWindow(20, 178, 170, 188, 4),
		// Hotbar
		addWindow(360, 728, 560, 52, 1),
		// Cast bar
		addWindow(500, 696, 280, 22, 1),
		// Minimap
		addWindow(1100, 20, 160, 160, 2),
		// Quest tracker
		addWindow(1030, 200, 230, 200, 4),
		// Currency
		addWindow(490, 8, 300, 24, 3),
		// Countdown
		addWindow(540, 36, 200, 30, 2),
		// Target frame
		addWindow(390, 68, 220, 52, 2),
		// Nameplates (3)
		addWindow(480, 350, 100, 30, 2),
		addWindow(560, 420, 90, 26, 2),
		addWindow(400, 440, 90, 26, 2),
		// Chat window (with children: msgDiv + input)
		addWindow(10, 520, 340, 228, 3),
		// Inventory window
		addWindow(980, 250, 287, 247, 3),
		// Skill tree
		addWindow(20, 346, 320, 228, 3),
		// Scoreboard
		addWindow(270, 216, 360, 248, 5),
		// Loot window
		addWindow(640, 256, 200, 208, 3),
		// Dialogue window
		addWindow(400, 352, 480, 208, 4),
		// Extra HUD
		addWindow(640, 100, 200, 80, 3),
	}

	e.SetChildren(root, windows)
	return root
}

// BenchmarkEngineComputeFull measures a complete layout engine reset + rebuild + compute.
// This represents the "full layout" path in CSSLayoutCache (when NeedsLayout is true).
func BenchmarkEngineComputeFull(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		e := New()
		e.Clear()
		root := buildGameDemoTree(e)
		e.AddRoot(root)
		e.Compute(1280, 800)
	}
}

// BenchmarkEngineComputeReuse measures compute reuse: engine reset + rebuild (no New).
// Represents the case where the engine is reused but cleared for each full layout.
func BenchmarkEngineComputeReuse(b *testing.B) {
	b.ReportAllocs()
	e := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Clear()
		root := buildGameDemoTree(e)
		e.AddRoot(root)
		e.Compute(1280, 800)
	}
}

// BenchmarkEngineComputeOnly measures compute on a pre-built tree.
// Represents the minimal cost if tree construction were free.
func BenchmarkEngineComputeOnly(b *testing.B) {
	e := New()
	root := buildGameDemoTree(e)
	e.AddRoot(root)
	e.Compute(1280, 800) // warm-up

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Re-compute on the same tree (positions change but structure doesn't)
		e.Compute(1280, 800)
	}
}

// BenchmarkChildrenOf measures the allocation cost of childrenOf.
func BenchmarkChildrenOf(b *testing.B) {
	e := New()
	root := buildGameDemoTree(e)
	e.AddRoot(root)
	e.Compute(1280, 800)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// childrenOf is called for every node during layout
		_ = e.childrenOf(int(root))
		for _, child := range e.nodes[root].children {
			_ = e.childrenOf(int(child))
		}
	}
}

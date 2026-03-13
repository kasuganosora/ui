package ui

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/widget"
)

// textMeasurerAdapter bridges widget.TextDrawer to layout.TextMeasurer.
// It approximates multi-line height by dividing total text width by available width.
type textMeasurerAdapter struct {
	drawer widget.TextDrawer
}

func (a *textMeasurerAdapter) MeasureText(text string, _ uint32, fontSize, maxWidth float32) (width, height float32) {
	w := a.drawer.MeasureText(text, fontSize)
	lh := a.drawer.LineHeight(fontSize)
	if lh <= 0 {
		lh = fontSize * 1.2
	}
	if maxWidth > 0 && w > maxWidth {
		// Approximate word-wrap: divide full width by available width.
		lines := float32(int(w/maxWidth) + 1)
		return maxWidth, lines * lh
	}
	return w, lh
}

// CSSLayout performs CSS-based layout on a widget tree using the layout engine.
// It walks the widget tree, builds layout nodes from widget styles, computes
// positions/sizes via flexbox/block flow, and writes results back to the element tree.
//
// Pass cfg to enable accurate text measurement (font size, word-wrap).
// Scrollable containers (Content, Div with overflow:scroll) have their children
// offset by the current scroll position.
func CSSLayout(tree *core.Tree, root widget.Widget, w, h float32, cfg ...*widget.Config) {
	engine := layout.New()
	engine.Clear()

	// Wire up text measurer if a text renderer is available.
	if len(cfg) > 0 && cfg[0] != nil && cfg[0].TextRenderer != nil {
		engine.SetTextMeasurer(&textMeasurerAdapter{cfg[0].TextRenderer})
	}

	// Build layout tree from widget tree
	var widgets []widget.Widget
	var nodeChildren [][]layout.NodeID
	rootNode := buildLayoutNode(engine, root, &widgets, &nodeChildren)
	engine.AddRoot(rootNode)

	// Compute layout
	engine.Compute(w, h)

	// Write results back to the tree, handling scroll offsets
	applyLayoutResults(tree, engine, rootNode, widgets, nodeChildren, 0, 0)
}

// CSSLayoutCache caches layout computation between frames.
// When only scroll offsets change (no DirtyLayout), it skips the expensive
// buildLayoutNode + engine.Compute steps and only re-applies positions.
// This turns a ~90μs full layout into a ~5μs position-only update.
//
// Perf notes:
//   - engine is reused across full layouts (Clear() keeps backing arrays)
//   - widgets/nodeChildren are slices indexed by NodeID (sequential ints),
//     not maps — O(1) slice access vs O(1) amortized hash + cache misses
type CSSLayoutCache struct {
	engine       *layout.Engine
	widgets      []widget.Widget      // indexed by NodeID (0-based sequential)
	nodeChildren [][]layout.NodeID    // indexed by NodeID
	rootNode     layout.NodeID
	lastW, lastH float32
	valid        bool
	measurer     *textMeasurerAdapter
}

// NewCSSLayoutCache creates a new layout cache.
func NewCSSLayoutCache() *CSSLayoutCache {
	return &CSSLayoutCache{engine: layout.New()}
}

// Invalidate forces full recomputation on next Layout call.
func (lc *CSSLayoutCache) Invalidate() {
	lc.valid = false
}

// Layout performs cached CSS layout. If tree structure hasn't changed (no DirtyLayout)
// and viewport size is the same, it only re-applies scroll offsets (~20x faster).
func (lc *CSSLayoutCache) Layout(tree *core.Tree, root widget.Widget, w, h float32, cfg ...*widget.Config) {
	needsFull := !lc.valid || tree.NeedsLayout() || w != lc.lastW || h != lc.lastH

	if needsFull {
		// Reuse engine: Clear() resets length but keeps backing arrays, avoiding
		// repeated allocations from node-slice growth during buildLayoutNode.
		lc.engine.Clear()

		if len(cfg) > 0 && cfg[0] != nil && cfg[0].TextRenderer != nil {
			if lc.measurer == nil {
				lc.measurer = &textMeasurerAdapter{cfg[0].TextRenderer}
			} else {
				lc.measurer.drawer = cfg[0].TextRenderer
			}
			lc.engine.SetTextMeasurer(lc.measurer)
		}

		// Reset slice-based lookup tables (reuse backing arrays when possible).
		lc.widgets = lc.widgets[:0]
		lc.nodeChildren = lc.nodeChildren[:0]
		lc.rootNode = buildLayoutNode(lc.engine, root, &lc.widgets, &lc.nodeChildren)
		lc.engine.AddRoot(lc.rootNode)
		lc.engine.Compute(w, h)
		lc.lastW, lc.lastH = w, h
		lc.valid = true
	}

	// Always re-apply positions (fast path) — updates scroll offsets every frame.
	applyLayoutResults(tree, lc.engine, lc.rootNode, lc.widgets, lc.nodeChildren, 0, 0)
}

// buildLayoutNode recursively creates layout nodes from widgets.
// Text widgets (widget.Text) are registered as text nodes so the layout engine
// can measure their intrinsic size via the TextMeasurer.
// widgets and nodeChildren are slice-indexed by NodeID (sequential 0-based ints).
func buildLayoutNode(engine *layout.Engine, w widget.Widget, widgets *[]widget.Widget, nodeChildren *[][]layout.NodeID) layout.NodeID {
	style := w.Style()

	var nodeID layout.NodeID
	if txt, ok := w.(*widget.Text); ok {
		style.FontSize = txt.FontSize()
		if style.FontSize == 0 {
			style.FontSize = 14
		}
		nodeID = engine.AddTextNode(style, txt.Text())
	} else {
		nodeID = engine.AddNode(style)
	}

	// Extend slice to cover this nodeID (nodeIDs are assigned sequentially).
	id := int(nodeID)
	for len(*widgets) <= id {
		*widgets = append(*widgets, nil)
		*nodeChildren = append(*nodeChildren, nil)
	}
	(*widgets)[id] = w

	children := w.Children()
	if len(children) > 0 {
		childIDs := make([]layout.NodeID, 0, len(children))
		for _, child := range children {
			childID := buildLayoutNode(engine, child, widgets, nodeChildren)
			childIDs = append(childIDs, childID)
		}
		engine.SetChildren(nodeID, childIDs)
		(*nodeChildren)[id] = childIDs
	}

	return nodeID
}

// applyLayoutResults writes computed layout to the tree, accumulating parent offsets
// and applying scroll offsets for scrollable containers.
// widgets and nodeChildren are slice-indexed by NodeID for O(1) access without hashing.
func applyLayoutResults(tree *core.Tree, engine *layout.Engine, nodeID layout.NodeID,
	widgets []widget.Widget, nodeChildren [][]layout.NodeID,
	parentX, parentY float32) {

	id := int(nodeID)
	if id >= len(widgets) {
		return
	}
	w := widgets[id]
	if w == nil {
		return
	}
	result := engine.GetResult(nodeID)

	// Absolute position = parent offset + layout position
	absX := parentX + result.X
	absY := parentY + result.Y

	tree.SetLayout(w.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(absX, absY, result.Width, result.Height),
	})

	// Handle scrollable Content widget
	scrollOffsetY := float32(0)
	if c, ok := w.(*widget.Content); ok {
		if result.ContentHeight > 0 {
			c.SetContentHeight(result.ContentHeight)
			c.ScrollBy(0) // clamp
		}
		scrollOffsetY = c.ScrollY()
	}

	// Handle scrollable Div widget (explicit flag OR CSS overflow:scroll/auto)
	if d, ok := w.(*widget.Div); ok {
		ov := d.Style().Overflow
		if d.IsScrollable() || ov == layout.OverflowScroll || ov == layout.OverflowAuto {
			if result.ContentHeight > 0 {
				d.SetContentHeight(result.ContentHeight)
			}
			scrollOffsetY = d.ScrollY()
		}
	}

	// Recurse into children
	childOffsetX := absX
	childOffsetY := absY - scrollOffsetY
	if id < len(nodeChildren) {
		for _, childID := range nodeChildren[id] {
			applyLayoutResults(tree, engine, childID, widgets, nodeChildren, childOffsetX, childOffsetY)
		}
	}
}

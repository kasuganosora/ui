//go:build windows

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
	widgetMap := make(map[layout.NodeID]widget.Widget)
	nodeToChildren := make(map[layout.NodeID][]layout.NodeID)
	rootNode := buildLayoutNode(engine, root, widgetMap, nodeToChildren)
	engine.AddRoot(rootNode)

	// Compute layout
	engine.Compute(w, h)

	// Write results back to the tree, handling scroll offsets
	applyLayoutResults(tree, engine, rootNode, widgetMap, nodeToChildren, 0, 0)
}

// buildLayoutNode recursively creates layout nodes from widgets.
// Text widgets (widget.Text) are registered as text nodes so the layout engine
// can measure their intrinsic size via the TextMeasurer.
func buildLayoutNode(engine *layout.Engine, w widget.Widget, widgetMap map[layout.NodeID]widget.Widget, nodeChildren map[layout.NodeID][]layout.NodeID) layout.NodeID {
	style := w.Style()

	var nodeID layout.NodeID
	if txt, ok := w.(*widget.Text); ok {
		// Carry font size into the layout style for text measurement.
		style.FontSize = txt.FontSize()
		if style.FontSize == 0 {
			style.FontSize = 14
		}
		nodeID = engine.AddTextNode(style, txt.Text())
	} else {
		nodeID = engine.AddNode(style)
	}
	widgetMap[nodeID] = w

	children := w.Children()
	if len(children) > 0 {
		childIDs := make([]layout.NodeID, 0, len(children))
		for _, child := range children {
			childID := buildLayoutNode(engine, child, widgetMap, nodeChildren)
			childIDs = append(childIDs, childID)
		}
		engine.SetChildren(nodeID, childIDs)
		nodeChildren[nodeID] = childIDs
	}

	return nodeID
}

// applyLayoutResults writes computed layout to the tree, accumulating parent offsets
// and applying scroll offsets for scrollable containers.
func applyLayoutResults(tree *core.Tree, engine *layout.Engine, nodeID layout.NodeID,
	widgetMap map[layout.NodeID]widget.Widget, nodeChildren map[layout.NodeID][]layout.NodeID,
	parentX, parentY float32) {

	w := widgetMap[nodeID]
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

	// Handle scrollable Div widget
	if d, ok := w.(*widget.Div); ok && d.IsScrollable() {
		scrollOffsetY = d.ScrollY()
	}

	// Recurse into children
	childOffsetX := absX
	childOffsetY := absY - scrollOffsetY
	for _, childID := range nodeChildren[nodeID] {
		applyLayoutResults(tree, engine, childID, widgetMap, nodeChildren, childOffsetX, childOffsetY)
	}
}

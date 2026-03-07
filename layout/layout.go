package layout

import uimath "github.com/kasuganosora/ui/math"

// TextMeasurer provides text measurement for layout.
// Implemented by the font system.
type TextMeasurer interface {
	MeasureText(text string, fontID uint32, fontSize float32, maxWidth float32) (width, height float32)
}

// NodeID identifies a layout node.
type NodeID int

const InvalidNode NodeID = -1

// layoutNode is the internal representation of a node during layout.
type layoutNode struct {
	id       NodeID
	parent   NodeID
	children []NodeID
	style    Style
	text     string // For text nodes

	result Result
}

// Result contains the computed layout output for a node.
type Result struct {
	X, Y          float32
	Width, Height float32
}

// Bounds returns the result as a Rect.
func (r Result) Bounds() uimath.Rect {
	return uimath.NewRect(r.X, r.Y, r.Width, r.Height)
}

// ContentBounds returns the content area (bounds minus padding/border).
func (r Result) ContentBounds(style *Style, parentWidth float32) uimath.Rect {
	padTop, padRight, padBottom, padLeft := resolveEdges(style.Padding, parentWidth)
	bdrTop, bdrRight, bdrBottom, bdrLeft := resolveEdges(style.Border, parentWidth)
	return uimath.NewRect(
		r.X+padLeft+bdrLeft,
		r.Y+padTop+bdrTop,
		r.Width-(padLeft+padRight+bdrLeft+bdrRight),
		r.Height-(padTop+padBottom+bdrTop+bdrBottom),
	)
}

// Engine is the layout domain service.
// It takes a tree of nodes with styles and computes positions/sizes.
type Engine struct {
	nodes    []layoutNode
	roots    []NodeID
	measurer TextMeasurer
}

// New creates a new layout engine.
func New() *Engine {
	return &Engine{}
}

// SetTextMeasurer sets the text measurement provider.
func (e *Engine) SetTextMeasurer(m TextMeasurer) {
	e.measurer = m
}

// Clear resets the engine for a new layout pass.
func (e *Engine) Clear() {
	e.nodes = e.nodes[:0]
	e.roots = e.roots[:0]
}

// AddNode adds a node and returns its ID.
func (e *Engine) AddNode(style Style) NodeID {
	id := NodeID(len(e.nodes))
	e.nodes = append(e.nodes, layoutNode{
		id:     id,
		parent: InvalidNode,
		style:  style,
	})
	return id
}

// AddTextNode adds a text node that will be measured during layout.
func (e *Engine) AddTextNode(style Style, text string) NodeID {
	id := NodeID(len(e.nodes))
	e.nodes = append(e.nodes, layoutNode{
		id:     id,
		parent: InvalidNode,
		style:  style,
		text:   text,
	})
	return id
}

// SetChildren sets the children of a node.
func (e *Engine) SetChildren(parent NodeID, children []NodeID) {
	if int(parent) >= len(e.nodes) {
		return
	}
	e.nodes[parent].children = children
	for _, c := range children {
		if int(c) < len(e.nodes) {
			e.nodes[c].parent = parent
		}
	}
}

// AddRoot marks a node as a root for layout.
func (e *Engine) AddRoot(id NodeID) {
	e.roots = append(e.roots, id)
}

// Compute performs layout computation for all root nodes.
func (e *Engine) Compute(viewportWidth, viewportHeight float32) {
	for _, rootID := range e.roots {
		root := &e.nodes[rootID]

		// Resolve root dimensions
		rootW := viewportWidth
		if !root.style.Width.IsAuto() {
			if w, ok := root.style.Width.Resolve(viewportWidth); ok {
				rootW = w
			}
		}
		rootH := float32(0) // auto by default
		if !root.style.Height.IsAuto() {
			if h, ok := root.style.Height.Resolve(viewportHeight); ok {
				rootH = h
			}
		}

		root.result.X = 0
		root.result.Y = 0
		root.result.Width = rootW
		root.result.Height = rootH

		e.layoutNode(int(rootID), rootW, viewportHeight)
	}
}

// GetResult returns the computed layout for a node.
func (e *Engine) GetResult(id NodeID) Result {
	if int(id) >= len(e.nodes) {
		return Result{}
	}
	return e.nodes[id].result
}

// NodeCount returns the number of nodes.
func (e *Engine) NodeCount() int {
	return len(e.nodes)
}

// layoutNode performs layout on a single node and its children.
// The node's result.X/Y/Width/Height should be set by the caller before this.
func (e *Engine) layoutNode(nodeIdx int, availWidth, availHeight float32) {
	node := &e.nodes[nodeIdx]
	style := &node.style

	if style.Display == DisplayNone {
		node.result = Result{}
		return
	}

	// Dispatch to layout algorithm for children
	switch style.Display {
	case DisplayFlex:
		e.layoutFlex(nodeIdx, availWidth, availHeight)
	case DisplayBlock:
		e.layoutBlock(nodeIdx, availWidth, availHeight)
	default:
		e.layoutBlock(nodeIdx, availWidth, availHeight)
	}
}

// layoutAbsolute positions an absolutely positioned element
// relative to its containing block (parentWidth x parentHeight).
func (e *Engine) layoutAbsolute(nodeIdx int, parentWidth, parentHeight float32) {
	node := &e.nodes[nodeIdx]
	style := &node.style

	// Resolve position offsets
	x := float32(0)
	y := float32(0)
	if v, ok := style.Left.Resolve(parentWidth); ok {
		x = v
	}
	if v, ok := style.Top.Resolve(parentHeight); ok {
		y = v
	}

	// Resolve width
	w := float32(0)
	if !style.Width.IsAuto() {
		w, _ = style.Width.Resolve(parentWidth)
	} else if !style.Left.IsAuto() && !style.Right.IsAuto() {
		l, _ := style.Left.Resolve(parentWidth)
		r, _ := style.Right.Resolve(parentWidth)
		w = parentWidth - l - r
	}

	// Resolve height
	h := float32(0)
	if !style.Height.IsAuto() {
		h, _ = style.Height.Resolve(parentHeight)
	} else if !style.Top.IsAuto() && !style.Bottom.IsAuto() {
		t, _ := style.Top.Resolve(parentHeight)
		b, _ := style.Bottom.Resolve(parentHeight)
		h = parentHeight - t - b
	}

	w = constrainSize(w, parentWidth, style.MinWidth, style.MaxWidth)
	h = constrainSize(h, parentHeight, style.MinHeight, style.MaxHeight)

	node.result.X = x
	node.result.Y = y
	node.result.Width = w
	node.result.Height = h

	// Layout children within this absolute element
	e.layoutNode(nodeIdx, w, h)
}

// applyRelativeOffset offsets a relatively positioned element from its normal
// flow position. The offset does not affect sibling layout.
func (e *Engine) applyRelativeOffset(nodeIdx int, parentWidth, parentHeight float32) {
	node := &e.nodes[nodeIdx]
	style := &node.style
	if style.Position != PositionRelative {
		return
	}
	if v, ok := style.Left.Resolve(parentWidth); ok {
		node.result.X += v
	} else if v, ok := style.Right.Resolve(parentWidth); ok {
		node.result.X -= v
	}
	if v, ok := style.Top.Resolve(parentHeight); ok {
		node.result.Y += v
	} else if v, ok := style.Bottom.Resolve(parentHeight); ok {
		node.result.Y -= v
	}
}

// childrenOf returns the child indices for a node.
func (e *Engine) childrenOf(nodeIdx int) []int {
	node := &e.nodes[nodeIdx]
	result := make([]int, 0, len(node.children))
	for _, c := range node.children {
		result = append(result, int(c))
	}
	return result
}

package layout

import (
	"unsafe"

	uimath "github.com/kasuganosora/ui/math"
)

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
	// ContentWidth/ContentHeight track the total extent of children.
	// Non-zero only for overflow:scroll/auto containers.
	ContentWidth  float32
	ContentHeight float32
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
	nodes      []layoutNode
	roots      []NodeID
	measurer   TextMeasurer
	fixedNodes []int // indices of position:fixed elements, laid out after everything
	vpWidth    float32
	vpHeight   float32
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
	e.fixedNodes = e.fixedNodes[:0]
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
	e.vpWidth = viewportWidth
	e.vpHeight = viewportHeight
	e.fixedNodes = e.fixedNodes[:0]

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

	// Layout fixed elements relative to the viewport
	for _, nodeIdx := range e.fixedNodes {
		e.layoutAbsolute(nodeIdx, viewportWidth, viewportHeight)
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
	case DisplayGrid:
		e.layoutGrid(nodeIdx, availWidth, availHeight)
	case DisplayBlock:
		e.layoutBlock(nodeIdx, availWidth, availHeight)
	default:
		e.layoutBlock(nodeIdx, availWidth, availHeight)
	}
}

// layoutAbsolute positions an absolutely positioned element
// relative to its containing block (parentWidth x parentHeight).
//
// Supports the full CSS absolute positioning model:
//   - left/top for edge-anchored placement
//   - right/bottom for opposite-edge anchored placement
//   - left:50% + margin-left:-N for center-offset patterns
//   - top:50% + margin-top:-N for vertical center-offset patterns
//   - auto width from left+right, auto height from top+bottom
func (e *Engine) layoutAbsolute(nodeIdx int, parentWidth, parentHeight float32) {
	node := &e.nodes[nodeIdx]
	style := &node.style

	// Padding+border for border-box adjustment
	absPadH, absPadV := resolveEdgesTotal(style.Padding, parentWidth)
	absBdrH, absBdrV := resolveEdgesTotal(style.Border, parentWidth)

	// Resolve width first (needed for right-edge positioning)
	w := float32(0)
	if !style.Width.IsAuto() {
		w, _ = style.Width.Resolve(parentWidth)
		w = AdjustBoxSizing(w, style.BoxSizing, absPadH, absBdrH)
	} else if !style.Left.IsAuto() && !style.Right.IsAuto() {
		l, _ := style.Left.Resolve(parentWidth)
		r, _ := style.Right.Resolve(parentWidth)
		w = parentWidth - l - r
	}

	// Resolve height first (needed for bottom-edge positioning)
	h := float32(0)
	if !style.Height.IsAuto() {
		h, _ = style.Height.Resolve(parentHeight)
		h = AdjustBoxSizing(h, style.BoxSizing, absPadV, absBdrV)
	} else if !style.Top.IsAuto() && !style.Bottom.IsAuto() {
		t, _ := style.Top.Resolve(parentHeight)
		b, _ := style.Bottom.Resolve(parentHeight)
		h = parentHeight - t - b
	}

	w = constrainSize(w, parentWidth, style.MinWidth, style.MaxWidth)
	h = constrainSize(h, parentHeight, style.MinHeight, style.MaxHeight)

	// Resolve horizontal position: left takes precedence; fall back to right
	x := float32(0)
	if !style.Left.IsAuto() {
		x, _ = style.Left.Resolve(parentWidth)
	} else if !style.Right.IsAuto() {
		r, _ := style.Right.Resolve(parentWidth)
		x = parentWidth - w - r
	}

	// Resolve vertical position: top takes precedence; fall back to bottom
	y := float32(0)
	if !style.Top.IsAuto() {
		y, _ = style.Top.Resolve(parentHeight)
	} else if !style.Bottom.IsAuto() {
		b, _ := style.Bottom.Resolve(parentHeight)
		y = parentHeight - h - b
	}

	// Apply margins (useful for center-offset patterns like left:50% + margin-left:-Npx)
	if ml, ok := style.Margin.Left.Resolve(parentWidth); ok {
		x += ml
	}
	if mr, ok := style.Margin.Right.Resolve(parentWidth); ok && style.Left.IsAuto() && !style.Right.IsAuto() {
		x -= mr
	}
	if mt, ok := style.Margin.Top.Resolve(parentHeight); ok {
		y += mt
	}
	if mb, ok := style.Margin.Bottom.Resolve(parentHeight); ok && style.Top.IsAuto() && !style.Bottom.IsAuto() {
		y -= mb
	}

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
// Returns the internal slice directly (zero allocation) — callers must not modify.
func (e *Engine) childrenOf(nodeIdx int) []int {
	node := &e.nodes[nodeIdx]
	return *(*[]int)(unsafe.Pointer(&node.children))
}

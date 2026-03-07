package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// TreeNode represents a node in the tree.
type TreeNode struct {
	Key      string
	Label    string
	Children []*TreeNode
	Expanded bool
	Selected bool
	Icon     string
	Data     interface{}
}

// Tree is an expandable tree view widget.
type Tree struct {
	Base
	roots      []*TreeNode
	indent     float32
	itemHeight float32
	onSelect   func(node *TreeNode)
	onExpand   func(node *TreeNode)
}

func NewTree(tree *core.Tree, cfg *Config) *Tree {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Tree{
		Base:       NewBase(tree, core.TypeCustom, cfg),
		indent:     20,
		itemHeight: 32,
	}
}

func (t *Tree) Roots() []*TreeNode         { return t.roots }
func (t *Tree) SetIndent(i float32)        { t.indent = i }
func (t *Tree) SetItemHeight(h float32)    { t.itemHeight = h }
func (t *Tree) OnSelect(fn func(*TreeNode)) { t.onSelect = fn }
func (t *Tree) OnExpand(fn func(*TreeNode)) { t.onExpand = fn }

func (t *Tree) SetRoots(roots []*TreeNode) {
	t.roots = roots
}

func (t *Tree) AddRoot(node *TreeNode) {
	t.roots = append(t.roots, node)
}

// FindNode searches for a node by key.
func (t *Tree) FindNode(key string) *TreeNode {
	for _, root := range t.roots {
		if n := findNodeRecursive(root, key); n != nil {
			return n
		}
	}
	return nil
}

func findNodeRecursive(node *TreeNode, key string) *TreeNode {
	if node.Key == key {
		return node
	}
	for _, child := range node.Children {
		if n := findNodeRecursive(child, key); n != nil {
			return n
		}
	}
	return nil
}

// ExpandAll expands all nodes.
func (t *Tree) ExpandAll() {
	for _, root := range t.roots {
		expandAllRecursive(root)
	}
}

func expandAllRecursive(node *TreeNode) {
	node.Expanded = true
	for _, child := range node.Children {
		expandAllRecursive(child)
	}
}

// CollapseAll collapses all nodes.
func (t *Tree) CollapseAll() {
	for _, root := range t.roots {
		collapseAllRecursive(root)
	}
}

func collapseAllRecursive(node *TreeNode) {
	node.Expanded = false
	for _, child := range node.Children {
		collapseAllRecursive(child)
	}
}

func (t *Tree) Draw(buf *render.CommandBuffer) {
	bounds := t.Bounds()
	if bounds.IsEmpty() {
		return
	}
	y := bounds.Y
	for _, root := range t.roots {
		y = t.drawNode(buf, root, bounds.X, y, 0, bounds)
	}
}

func (t *Tree) drawNode(buf *render.CommandBuffer, node *TreeNode, x, y float32, depth int, bounds uimath.Rect) float32 {
	if y+t.itemHeight > bounds.Y+bounds.Height {
		return y
	}
	cfg := t.config
	indent := x + float32(depth)*t.indent

	// Selection highlight
	if node.Selected {
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(bounds.X, y, bounds.Width, t.itemHeight),
			FillColor: uimath.RGBA(0.09, 0.42, 1, 0.08),
		}, 1, 1)
	}

	// Expand/collapse indicator
	if len(node.Children) > 0 {
		arrowX := indent + 4
		arrowY := y + t.itemHeight/2
		arrowSize := float32(6)
		if node.Expanded {
			// Down arrow (V shape as horizontal line)
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(arrowX, arrowY-1, arrowSize, 2),
				FillColor: cfg.TextColor,
			}, 2, 0.6)
		} else {
			// Right arrow (> shape as vertical line)
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(arrowX+2, arrowY-arrowSize/2, 2, arrowSize),
				FillColor: cfg.TextColor,
			}, 2, 0.6)
		}
	}

	// Label
	textX := indent + 20
	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, node.Label, textX, y+(t.itemHeight-lh)/2, cfg.FontSize, bounds.Width-(textX-bounds.X)-cfg.SpaceSM, cfg.TextColor, 1)
	}

	y += t.itemHeight

	// Children
	if node.Expanded {
		for _, child := range node.Children {
			y = t.drawNode(buf, child, x, y, depth+1, bounds)
		}
	}
	return y
}

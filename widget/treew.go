package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
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
	Checked  bool
	Lazy     func(*TreeNode)
}

// Tree is an expandable tree view widget.
type Tree struct {
	Base
	roots      []*TreeNode
	indent     float32
	itemHeight float32
	onSelect   func(node *TreeNode)
	onExpand   func(node *TreeNode)
	cursor     int
	selected   string
	dragKey    string
}

func NewTree(tree *core.Tree, cfg *Config) *Tree {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	tr := &Tree{
		Base:       NewBase(tree, core.TypeCustom, cfg),
		indent:     20,
		itemHeight: 32,
		cursor:     -1,
	}
	bindArrows(tree, tr.id, func(k event.Key) { tr.onKey(k) })
	tree.AddHandler(tr.id, event.MouseDown, func(e *event.Event) {
		if n := tr.nodeAt(e.GlobalY); n != nil {
			tr.dragKey = n.Key
		}
	})
	tree.AddHandler(tr.id, event.MouseUp, func(e *event.Event) {
		if tr.dragKey == "" {
			return
		}
		target := tr.nodeAt(e.GlobalY)
		if target != nil && target.Key != tr.dragKey {
			tr.reorder(tr.dragKey, target.Key)
		}
		tr.dragKey = ""
	})
	return tr
}

func (t *Tree) nodeAt(y float32) *TreeNode {
	nodes := t.flat()
	if len(nodes) == 0 || t.itemHeight <= 0 {
		return nil
	}
	i := int((y - t.Bounds().Y) / t.itemHeight)
	if i < 0 || i >= len(nodes) {
		return nil
	}
	return nodes[i]
}

func (t *Tree) reorder(from, to string) {
	nodes := t.flat()
	var a, b *TreeNode
	for _, n := range nodes {
		if n.Key == from {
			a = n
		}
		if n.Key == to {
			b = n
		}
	}
	if a == nil || b == nil {
		return
	}
	// Reorder among roots when both are roots. Otherwise swap within the same parent slice.
	if t.swapIn(t.roots, a, b) {
		return
	}
}

func (t *Tree) swapIn(nodes []*TreeNode, a, b *TreeNode) bool {
	ia, ib := -1, -1
	for i, n := range nodes {
		if n == a {
			ia = i
		}
		if n == b {
			ib = i
		}
	}
	if ia >= 0 && ib >= 0 {
		nodes[ia], nodes[ib] = nodes[ib], nodes[ia]
		return true
	}
	for _, n := range nodes {
		if t.swapIn(n.Children, a, b) {
			return true
		}
	}
	return false
}

func (t *Tree) SelectedKey() string { return t.selected }

func (t *Tree) flat() []*TreeNode {
	var out []*TreeNode
	var walk func([]*TreeNode)
	walk = func(nodes []*TreeNode) {
		for _, n := range nodes {
			out = append(out, n)
			if n.Expanded {
				walk(n.Children)
			}
		}
	}
	walk(t.roots)
	return out
}

func (t *Tree) onKey(k event.Key) {
	nodes := t.flat()
	if len(nodes) == 0 {
		return
	}
	switch k {
	case event.KeyArrowDown:
		t.cursor++
		if t.cursor >= len(nodes) {
			t.cursor = len(nodes) - 1
		}
	case event.KeyArrowUp:
		if t.cursor < 0 {
			t.cursor = 0
		} else if t.cursor > 0 {
			t.cursor--
		}
	case event.KeyEnter:
		if t.cursor < 0 {
			t.cursor = 0
		}
		if t.cursor >= 0 && t.cursor < len(nodes) {
			t.selected = nodes[t.cursor].Key
			nodes[t.cursor].Selected = true
			if t.onSelect != nil {
				t.onSelect(nodes[t.cursor])
			}
		}
	case event.KeyArrowRight:
		if t.cursor >= 0 && t.cursor < len(nodes) {
			n := nodes[t.cursor]
			if len(n.Children) == 0 && n.Lazy != nil {
				n.Lazy(n)
			}
			n.Expanded = true
			if t.onExpand != nil {
				t.onExpand(n)
			}
		}
	case event.KeyArrowLeft:
		if t.cursor >= 0 && t.cursor < len(nodes) {
			nodes[t.cursor].Expanded = false
		}
	case event.KeySpace:
		if t.cursor >= 0 && t.cursor < len(nodes) {
			nodes[t.cursor].Checked = !nodes[t.cursor].Checked
		}
	}
}

func (t *Tree) Roots() []*TreeNode          { return t.roots }
func (t *Tree) SetIndent(i float32)         { t.indent = i }
func (t *Tree) SetItemHeight(h float32)     { t.itemHeight = h }
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

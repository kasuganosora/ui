package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// TreeSelect is a dropdown with a tree structure for selection.
type TreeSelect struct {
	Base
	roots    []*TreeNode
	selected string
	open     bool
	dropH    float32
	onChange func(string)
}

func NewTreeSelect(tree *core.Tree, cfg *Config) *TreeSelect {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	ts := &TreeSelect{
		Base:  NewBase(tree, core.TypeCustom, cfg),
		dropH: 200,
	}
	tree.AddHandler(ts.id, event.MouseClick, func(e *event.Event) {
		ts.open = !ts.open
	})
	return ts
}

func (ts *TreeSelect) Selected() string              { return ts.selected }
func (ts *TreeSelect) IsOpen() bool                  { return ts.open }
func (ts *TreeSelect) SetOpen(o bool)                { ts.open = o }
func (ts *TreeSelect) OnChange(fn func(string))      { ts.onChange = fn }

func (ts *TreeSelect) SetRoots(roots []*TreeNode) {
	ts.roots = roots
}

func (ts *TreeSelect) SetSelected(key string) {
	ts.selected = key
	if ts.onChange != nil {
		ts.onChange(key)
	}
}

func (ts *TreeSelect) selectedLabel() string {
	for _, root := range ts.roots {
		if n := findNodeRecursive(root, ts.selected); n != nil {
			return n.Label
		}
	}
	return ""
}

func (ts *TreeSelect) Draw(buf *render.CommandBuffer) {
	bounds := ts.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := ts.config

	// Input
	buf.DrawRect(render.RectCmd{
		Bounds:      bounds,
		FillColor:   cfg.BgColor,
		BorderColor: cfg.BorderColor,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 1, 1)

	label := ts.selectedLabel()
	if label == "" {
		label = "Select..."
	}
	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		color := cfg.TextColor
		if ts.selected == "" {
			color = cfg.DisabledColor
		}
		cfg.TextRenderer.DrawText(buf, label, bounds.X+cfg.SpaceSM, bounds.Y+(bounds.Height-lh)/2, cfg.FontSize, bounds.Width-cfg.SpaceSM*2, color, 1)
	}

	// Dropdown tree
	if !ts.open {
		return
	}
	dy := bounds.Y + bounds.Height + 4
	buf.DrawOverlay(render.RectCmd{
		Bounds:      uimath.NewRect(bounds.X, dy, bounds.Width, ts.dropH),
		FillColor:   uimath.ColorWhite,
		BorderColor: cfg.BorderColor,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 40, 1)

	// Render tree nodes in dropdown
	y := dy + 4
	for _, root := range ts.roots {
		y = ts.drawDropNode(buf, root, bounds.X, y, 0, bounds.Width, dy+ts.dropH)
	}
}

func (ts *TreeSelect) drawDropNode(buf *render.CommandBuffer, node *TreeNode, x, y float32, depth int, width, maxY float32) float32 {
	if y > maxY {
		return y
	}
	cfg := ts.config
	itemH := float32(28)
	indent := x + float32(depth)*16 + cfg.SpaceSM

	if node.Key == ts.selected {
		buf.DrawOverlay(render.RectCmd{
			Bounds:    uimath.NewRect(x, y, width, itemH),
			FillColor: uimath.RGBA(0.09, 0.42, 1, 0.08),
		}, 41, 1)
	}
	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
		cfg.TextRenderer.DrawText(buf, node.Label, indent, y+(itemH-lh)/2, cfg.FontSizeSm, width-indent+x, cfg.TextColor, 1)
	}
	y += itemH

	if node.Expanded {
		for _, child := range node.Children {
			y = ts.drawDropNode(buf, child, x, y, depth+1, width, maxY)
		}
	}
	return y
}

package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// CascaderOption is a node in the cascading menu.
type CascaderOption struct {
	Label    string
	Value    string
	Children []*CascaderOption
}

// Cascader is a cascading selection dropdown.
type Cascader struct {
	Base
	options  []*CascaderOption
	selected []string // path of selected values
	open     bool
	onChange func([]string)
	cells    hitCells
}

func NewCascader(tree *core.Tree, cfg *Config) *Cascader {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	c := &Cascader{
		Base: NewBase(tree, core.TypeCustom, cfg),
	}
	c.cells = hitCells{tree: tree, parent: c.id}
	tree.AddHandler(c.id, event.MouseClick, func(e *event.Event) {
		c.onClick(e.GlobalX, e.GlobalY)
	})
	return c
}

func (c *Cascader) onClick(x, y float32) {
	b := c.Bounds()
	if c.open {
		if path, leaf, ok := c.hit(x, y); ok {
			c.selected = path
			if leaf {
				c.open = false
				c.cells.clear()
				if c.onChange != nil {
					c.onChange(append([]string(nil), path...))
				}
			}
			return
		}
	}
	if pointIn(x, y, b) {
		c.open = !c.open
		if !c.open {
			c.cells.clear()
		}
	}
}

func (c *Cascader) hit(x, y float32) (path []string, leaf bool, ok bool) {
	b := c.Bounds()
	panelW := float32(150)
	itemH := float32(32)
	dy := b.Y + b.Height + 4
	opts := c.options
	var prefix []string
	for level := 0; opts != nil; level++ {
		dx := b.X + float32(level)*panelW
		if x >= dx && x < dx+panelW && y >= dy {
			i := int((y - dy) / itemH)
			if i < 0 || i >= len(opts) {
				return nil, false, false
			}
			opt := opts[i]
			path = append(append([]string{}, prefix...), opt.Value)
			return path, len(opt.Children) == 0, true
		}
		if level >= len(c.selected) {
			break
		}
		var next []*CascaderOption
		for _, opt := range opts {
			if opt.Value == c.selected[level] {
				next = opt.Children
				prefix = append(prefix, opt.Value)
				break
			}
		}
		if next == nil {
			break
		}
		opts = next
	}
	return nil, false, false
}

func (c *Cascader) Options() []*CascaderOption { return c.options }
func (c *Cascader) Selected() []string         { return c.selected }
func (c *Cascader) IsOpen() bool               { return c.open }
func (c *Cascader) SetOpen(o bool)             { c.open = o }
func (c *Cascader) OnChange(fn func([]string)) { c.onChange = fn }

func (c *Cascader) SetOptions(opts []*CascaderOption) {
	c.options = opts
}

func (c *Cascader) SetSelected(path []string) {
	c.selected = make([]string, len(path))
	copy(c.selected, path)
	if c.onChange != nil {
		c.onChange(c.selected)
	}
}

func (c *Cascader) selectedLabel() string {
	if len(c.selected) == 0 {
		return ""
	}
	label := ""
	for i, v := range c.selected {
		if i > 0 {
			label += " / "
		}
		label += v
	}
	return label
}

func (c *Cascader) Draw(buf *render.CommandBuffer) {
	bounds := c.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := c.config

	// Input display
	buf.DrawRect(render.RectCmd{
		Bounds:      bounds,
		FillColor:   cfg.BgColor,
		BorderColor: cfg.BorderColor,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 1, 1)

	label := c.selectedLabel()
	if label == "" {
		label = "Select..."
	}
	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		color := cfg.TextColor
		if len(c.selected) == 0 {
			color = cfg.DisabledColor
		}
		cfg.TextRenderer.DrawText(buf, label, bounds.X+cfg.SpaceSM, bounds.Y+(bounds.Height-lh)/2, cfg.FontSize, bounds.Width-cfg.SpaceSM*2, color, 1)
	}

	// Dropdown panels
	if !c.open || len(c.options) == 0 {
		return
	}
	panelW := float32(150)
	itemH := float32(32)
	opts := c.options
	dx := bounds.X
	dy := bounds.Y + bounds.Height + 4

	for level := 0; opts != nil; level++ {
		panelH := float32(len(opts)) * itemH
		buf.DrawOverlay(render.RectCmd{
			Bounds:      uimath.NewRect(dx, dy, panelW, panelH),
			FillColor:   uimath.ColorWhite,
			BorderColor: cfg.BorderColor,
			BorderWidth: 1,
			Corners:     uimath.CornersAll(cfg.BorderRadius),
		}, int32(40+level), 1)

		var nextOpts []*CascaderOption
		for i, opt := range opts {
			iy := dy + float32(i)*itemH
			isSelected := level < len(c.selected) && c.selected[level] == opt.Value
			if isSelected {
				buf.DrawOverlay(render.RectCmd{
					Bounds:    uimath.NewRect(dx, iy, panelW, itemH),
					FillColor: uimath.RGBA(0.09, 0.42, 1, 0.08),
				}, int32(40+level), 1)
				if opt.Children != nil {
					nextOpts = opt.Children
				}
			}
			if cfg.TextRenderer != nil {
				lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
				cfg.TextRenderer.DrawText(buf, opt.Label, dx+cfg.SpaceSM, iy+(itemH-lh)/2, cfg.FontSize, panelW-cfg.SpaceSM*2, cfg.TextColor, 1)
			}
		}
		dx += panelW
		opts = nextOpts
	}
}

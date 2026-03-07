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
}

func NewCascader(tree *core.Tree, cfg *Config) *Cascader {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	c := &Cascader{
		Base: NewBase(tree, core.TypeCustom, cfg),
	}
	tree.AddHandler(c.id, event.MouseClick, func(e *event.Event) {
		c.open = !c.open
	})
	return c
}

func (c *Cascader) Options() []*CascaderOption  { return c.options }
func (c *Cascader) Selected() []string           { return c.selected }
func (c *Cascader) IsOpen() bool                 { return c.open }
func (c *Cascader) SetOpen(o bool)               { c.open = o }
func (c *Cascader) OnChange(fn func([]string))   { c.onChange = fn }

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

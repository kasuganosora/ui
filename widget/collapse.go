package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// CollapsePanel is a single expandable panel.
type CollapsePanel struct {
	Title    string
	Key      string
	Content  Widget
	Disabled bool
}

// Collapse is a container with expandable/collapsible panels.
type Collapse struct {
	Base
	panels    []CollapsePanel
	activeKeys map[string]bool
	accordion  bool
	bordered   bool
	onChange   func(keys []string)
}

func NewCollapse(tree *core.Tree, cfg *Config) *Collapse {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Collapse{
		Base:       NewBase(tree, core.TypeDiv, cfg),
		activeKeys: make(map[string]bool),
		bordered:   true,
	}
}

func (c *Collapse) SetPanels(p []CollapsePanel)   { c.panels = p }
func (c *Collapse) SetAccordion(a bool)            { c.accordion = a }
func (c *Collapse) SetBordered(b bool)             { c.bordered = b }
func (c *Collapse) OnChange(fn func([]string))     { c.onChange = fn }
func (c *Collapse) IsActive(key string) bool       { return c.activeKeys[key] }

func (c *Collapse) Toggle(key string) {
	if c.accordion {
		if c.activeKeys[key] {
			c.activeKeys = make(map[string]bool)
		} else {
			c.activeKeys = map[string]bool{key: true}
		}
	} else {
		if c.activeKeys[key] {
			delete(c.activeKeys, key)
		} else {
			c.activeKeys[key] = true
		}
	}
	if c.onChange != nil {
		var keys []string
		for k := range c.activeKeys {
			keys = append(keys, k)
		}
		c.onChange(keys)
	}
}

func (c *Collapse) Draw(buf *render.CommandBuffer) {
	bounds := c.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := c.config

	if c.bordered {
		buf.DrawRect(render.RectCmd{
			Bounds:      bounds,
			BorderColor: cfg.BorderColor,
			BorderWidth: cfg.BorderWidth,
			Corners:     uimath.CornersAll(cfg.BorderRadius),
		}, 0, 1)
	}

	y := bounds.Y
	headerH := float32(40)
	for _, panel := range c.panels {
		active := c.activeKeys[panel.Key]
		// Header
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(bounds.X, y, bounds.Width, headerH),
			FillColor: uimath.RGBA(0, 0, 0, 0.02),
		}, 1, 1)
		// Divider
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(bounds.X, y+headerH-1, bounds.Width, 1),
			FillColor: uimath.RGBA(0, 0, 0, 0.06),
		}, 1, 1)
		// Arrow indicator
		arrowX := bounds.X + cfg.SpaceMD
		arrowY := y + headerH/2
		arrowSize := float32(4)
		if active {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(arrowX-arrowSize, arrowY, arrowSize*2, arrowSize),
				FillColor: cfg.TextColor,
			}, 2, 0.5)
		} else {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(arrowX, arrowY-arrowSize, arrowSize, arrowSize*2),
				FillColor: cfg.TextColor,
			}, 2, 0.5)
		}
		// Title
		textX := bounds.X + cfg.SpaceMD + 16
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			cfg.TextRenderer.DrawText(buf, panel.Title, textX, y+(headerH-lh)/2, cfg.FontSize, bounds.Width-textX+bounds.X-cfg.SpaceMD, cfg.TextColor, 1)
		} else {
			tw := float32(len(panel.Title)) * cfg.FontSize * 0.55
			th := cfg.FontSize * 1.2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(textX, y+(headerH-th)/2, tw, th),
				FillColor: cfg.TextColor,
				Corners:   uimath.CornersAll(2),
			}, 2, 1)
		}
		y += headerH

		if active && panel.Content != nil {
			panel.Content.Draw(buf)
			y += 60 // placeholder content height
		}
	}
}

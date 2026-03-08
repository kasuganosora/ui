package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// ExpandIconPlacement controls where the expand icon appears.
type ExpandIconPlacement uint8

const (
	ExpandIconLeft  ExpandIconPlacement = iota
	ExpandIconRight
)

// CollapsePanel is a single expandable panel.
type CollapsePanel struct {
	Header             string
	Value              string
	Content            Widget
	Disabled           bool
	DestroyOnCollapse  bool
	HeaderRightContent Widget
}

// Collapse is a container with expandable/collapsible panels.
type Collapse struct {
	Base
	panels              []CollapsePanel
	activeKeys          map[string]bool
	expandMutex         bool
	borderless          bool
	expandOnRowClick    bool
	expandIcon          bool
	expandIconPlacement ExpandIconPlacement
	defaultExpandAll    bool
	disabled            bool
	headerIDs           []core.ElementID // clickable header elements
	onChange            func(keys []string)
}

func NewCollapse(tree *core.Tree, cfg *Config) *Collapse {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Collapse{
		Base:             NewBase(tree, core.TypeDiv, cfg),
		activeKeys:       make(map[string]bool),
		expandOnRowClick: true,
		expandIcon:       true,
	}
}

func (c *Collapse) SetPanels(p []CollapsePanel) {
	// Destroy old header elements
	c.destroyHeaderElements()
	c.panels = p
	c.createHeaderElements()
}

func (c *Collapse) AddPanel(p CollapsePanel) {
	c.panels = append(c.panels, p)
	c.createHeaderElement(len(c.panels) - 1)
}

func (c *Collapse) Panels() []CollapsePanel                      { return c.panels }
func (c *Collapse) SetExpandMutex(a bool)                        { c.expandMutex = a }
func (c *Collapse) SetBorderless(b bool)                         { c.borderless = b }
func (c *Collapse) SetExpandOnRowClick(b bool)                   { c.expandOnRowClick = b }
func (c *Collapse) SetExpandIcon(v bool)                         { c.expandIcon = v }
func (c *Collapse) SetExpandIconPlacement(p ExpandIconPlacement) { c.expandIconPlacement = p }
func (c *Collapse) SetDefaultExpandAll(v bool)                   { c.defaultExpandAll = v }
func (c *Collapse) SetDisabled(v bool)                           { c.disabled = v }
func (c *Collapse) OnChange(fn func([]string))                   { c.onChange = fn }
func (c *Collapse) IsActive(key string) bool                     { return c.activeKeys[key] }

// Deprecated: Use SetExpandMutex instead.
func (c *Collapse) SetAccordion(a bool) { c.expandMutex = a }

// Deprecated: Use SetBorderless instead.
func (c *Collapse) SetBordered(b bool) { c.borderless = !b }

func (c *Collapse) Toggle(key string) {
	// Check if panel is disabled
	for _, p := range c.panels {
		if p.Value == key && p.Disabled {
			return
		}
	}
	if c.expandMutex {
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

func (c *Collapse) createHeaderElements() {
	for i := range c.panels {
		c.createHeaderElement(i)
	}
}

func (c *Collapse) createHeaderElement(idx int) {
	hid := c.tree.CreateElement(core.TypeCustom)
	c.tree.AppendChild(c.id, hid)

	i := idx
	c.tree.AddHandler(hid, event.MouseClick, func(e *event.Event) {
		if !c.expandOnRowClick {
			return
		}
		if i < len(c.panels) {
			c.Toggle(c.panels[i].Value)
			c.tree.MarkDirty(c.id)
		}
	})

	c.headerIDs = append(c.headerIDs, hid)
}

func (c *Collapse) destroyHeaderElements() {
	for _, hid := range c.headerIDs {
		c.tree.DestroyElement(hid)
	}
	c.headerIDs = nil
}

func (c *Collapse) Draw(buf *render.CommandBuffer) {
	bounds := c.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := c.config

	if !c.borderless {
		buf.DrawRect(render.RectCmd{
			Bounds:      bounds,
			BorderColor: cfg.BorderColor,
			BorderWidth: cfg.BorderWidth,
			Corners:     uimath.CornersAll(cfg.BorderRadius),
		}, 0, 1)
	}

	y := bounds.Y
	headerH := float32(40)
	for i, panel := range c.panels {
		active := c.activeKeys[panel.Value]
		disabled := panel.Disabled

		// Set layout on header element for hit testing
		if i < len(c.headerIDs) {
			c.tree.SetLayout(c.headerIDs[i], core.LayoutResult{
				Bounds: uimath.NewRect(bounds.X, y, bounds.Width, headerH),
			})
		}

		// Header background
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(bounds.X, y, bounds.Width, headerH),
			FillColor: uimath.RGBA(0, 0, 0, 0.02),
		}, 1, 1)
		// Divider
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(bounds.X, y+headerH-1, bounds.Width, 1),
			FillColor: uimath.RGBA(0, 0, 0, 0.06),
		}, 1, 1)

		// Chevron arrow: ▸ (right) when collapsed, ▾ (down) when expanded
		arrowX := bounds.X + cfg.SpaceMD
		arrowY := y + headerH/2
		arrowStr := "\u25B8" // ▸
		if active {
			arrowStr = "\u25BE" // ▾
		}

		textOpacity := float32(1)
		if disabled {
			textOpacity = 0.35
		}

		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			cfg.TextRenderer.DrawText(buf, arrowStr, arrowX, arrowY-lh/2, cfg.FontSize, 16, cfg.TextColor, textOpacity)
		} else {
			arrowSize := float32(4)
			arrowClr := cfg.TextColor
			if disabled {
				arrowClr = cfg.DisabledColor
			}
			if active {
				// Down arrow: wider than tall
				buf.DrawRect(render.RectCmd{
					Bounds:    uimath.NewRect(arrowX-arrowSize, arrowY-arrowSize/2, arrowSize*2, arrowSize),
					FillColor: arrowClr,
					Corners:   uimath.CornersAll(1),
				}, 2, 1)
			} else {
				// Right arrow: taller than wide
				buf.DrawRect(render.RectCmd{
					Bounds:    uimath.NewRect(arrowX-arrowSize/2, arrowY-arrowSize, arrowSize, arrowSize*2),
					FillColor: arrowClr,
					Corners:   uimath.CornersAll(1),
				}, 2, 1)
			}
		}

		// Title
		textX := bounds.X + cfg.SpaceMD + 16
		titleClr := cfg.TextColor
		if disabled {
			titleClr = cfg.DisabledColor
		}
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			cfg.TextRenderer.DrawText(buf, panel.Header, textX, y+(headerH-lh)/2, cfg.FontSize, bounds.Width-textX+bounds.X-cfg.SpaceMD, titleClr, textOpacity)
		} else {
			tw := float32(len(panel.Header)) * cfg.FontSize * 0.55
			th := cfg.FontSize * 1.2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(textX, y+(headerH-th)/2, tw, th),
				FillColor: titleClr,
				Corners:   uimath.CornersAll(2),
			}, 2, 1)
		}
		y += headerH

		if active && panel.Content != nil {
			contentH := float32(60) // default content height
			// Try to get content height from its layout bounds
			contentElem := c.tree.Get(panel.Content.ElementID())
			if contentElem != nil {
				cb := contentElem.Layout().Bounds
				if cb.Height > 0 {
					contentH = cb.Height
				}
			}
			// Set content bounds so it draws in the right position
			contentPad := cfg.SpaceMD
			contentBounds := uimath.NewRect(bounds.X+contentPad, y+cfg.SpaceSM, bounds.Width-contentPad*2, contentH)
			if contentElem != nil {
				lo := contentElem.Layout()
				lo.Bounds = contentBounds
				c.tree.SetLayout(panel.Content.ElementID(), lo)
			}
			panel.Content.Draw(buf)
			y += contentH + cfg.SpaceSM*2
		}
	}
}

// Destroy cleans up header elements.
func (c *Collapse) Destroy() {
	c.destroyHeaderElements()
	c.Base.Destroy()
}

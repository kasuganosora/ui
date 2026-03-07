package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Card is a container with optional header and footer.
type Card struct {
	Base
	title       string
	bordered    bool
	bgColor     uimath.Color
	headerExtra Widget
}

func NewCard(tree *core.Tree, cfg *Config) *Card {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	c := &Card{
		Base:     NewBase(tree, core.TypeDiv, cfg),
		bordered: true,
		bgColor:  uimath.ColorWhite,
	}
	c.style.Display = layout.DisplayFlex
	c.style.FlexDirection = layout.FlexDirectionColumn
	return c
}

func (c *Card) SetTitle(t string)       { c.title = t }
func (c *Card) SetBordered(b bool)      { c.bordered = b }
func (c *Card) SetBgColor(cl uimath.Color) { c.bgColor = cl }
func (c *Card) SetHeaderExtra(w Widget)  { c.headerExtra = w }

func (c *Card) Draw(buf *render.CommandBuffer) {
	bounds := c.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := c.config

	borderW := float32(0)
	borderC := uimath.Color{}
	if c.bordered {
		borderW = cfg.BorderWidth
		borderC = cfg.BorderColor
	}

	buf.DrawRect(render.RectCmd{
		Bounds:      bounds,
		FillColor:   c.bgColor,
		BorderColor: borderC,
		BorderWidth: borderW,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 0, 1)

	headerH := float32(0)
	if c.title != "" {
		headerH = 48
		// Header divider
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(bounds.X, bounds.Y+headerH-1, bounds.Width, 1),
			FillColor: uimath.RGBA(0, 0, 0, 0.06),
		}, 1, 1)
		// Title
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			cfg.TextRenderer.DrawText(buf, c.title, bounds.X+cfg.SpaceLG, bounds.Y+(headerH-lh)/2, cfg.FontSize, bounds.Width-cfg.SpaceLG*2, cfg.TextColor, 1)
		} else {
			tw := float32(len(c.title)) * cfg.FontSize * 0.55
			th := cfg.FontSize * 1.2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X+cfg.SpaceLG, bounds.Y+(headerH-th)/2, tw, th),
				FillColor: cfg.TextColor,
				Corners:   uimath.CornersAll(2),
			}, 1, 1)
		}
	}
	_ = headerH
	c.DrawChildren(buf)
}

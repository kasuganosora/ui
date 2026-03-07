package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// AnchorLink represents a navigable section anchor.
type AnchorLink struct {
	Title    string
	Href     string
	Children []AnchorLink
}

// Anchor is a navigation sidebar anchored to page sections.
type Anchor struct {
	Base
	links    []AnchorLink
	active   string
	itemH    float32
	onChange func(string)
}

func NewAnchor(tree *core.Tree, cfg *Config) *Anchor {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Anchor{
		Base:  NewBase(tree, core.TypeCustom, cfg),
		itemH: 28,
	}
}

func (a *Anchor) Links() []AnchorLink        { return a.links }
func (a *Anchor) Active() string              { return a.active }
func (a *Anchor) SetActive(href string)       { a.active = href }
func (a *Anchor) OnChange(fn func(string))    { a.onChange = fn }

func (a *Anchor) SetLinks(links []AnchorLink) {
	a.links = links
}

func (a *Anchor) Draw(buf *render.CommandBuffer) {
	bounds := a.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := a.config

	// Vertical line
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(bounds.X, bounds.Y, 2, bounds.Height),
		FillColor: cfg.BorderColor,
	}, 1, 1)

	y := bounds.Y
	for _, link := range a.links {
		y = a.drawLink(buf, link, bounds.X, y, 0, bounds.Width)
	}
}

func (a *Anchor) drawLink(buf *render.CommandBuffer, link AnchorLink, x, y float32, depth int, width float32) float32 {
	cfg := a.config
	indent := x + 12 + float32(depth)*12
	isActive := link.Href == a.active

	if isActive {
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x, y, 2, a.itemH),
			FillColor: cfg.PrimaryColor,
		}, 2, 1)
	}

	if cfg.TextRenderer != nil {
		color := cfg.TextColor
		if isActive {
			color = cfg.PrimaryColor
		}
		lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
		cfg.TextRenderer.DrawText(buf, link.Title, indent, y+(a.itemH-lh)/2, cfg.FontSizeSm, width-indent+x, color, 1)
	}
	y += a.itemH

	for _, child := range link.Children {
		y = a.drawLink(buf, child, x, y, depth+1, width)
	}
	return y
}

package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// BreadcrumbItem is a single breadcrumb entry.
type BreadcrumbItem struct {
	Label string
	Href  string
}

// Breadcrumb displays a path of navigable links.
type Breadcrumb struct {
	Base
	items     []BreadcrumbItem
	separator string
	onClick   func(index int, href string)
}

func NewBreadcrumb(tree *core.Tree, cfg *Config) *Breadcrumb {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Breadcrumb{
		Base:      NewBase(tree, core.TypeCustom, cfg),
		separator: "/",
	}
}

func (b *Breadcrumb) SetItems(items []BreadcrumbItem) { b.items = items }
func (b *Breadcrumb) SetSeparator(s string)           { b.separator = s }
func (b *Breadcrumb) OnClick(fn func(int, string))    { b.onClick = fn }

func (b *Breadcrumb) Draw(buf *render.CommandBuffer) {
	bounds := b.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := b.config
	x := bounds.X
	for i, item := range b.items {
		isLast := i == len(b.items)-1
		color := cfg.PrimaryColor
		if isLast {
			color = cfg.TextColor
		}
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			tw := cfg.TextRenderer.MeasureText(item.Label, cfg.FontSize)
			cfg.TextRenderer.DrawText(buf, item.Label, x, bounds.Y+(bounds.Height-lh)/2, cfg.FontSize, tw+1, color, 1)
			x += tw
		} else {
			tw := float32(len(item.Label)) * cfg.FontSize * 0.55
			th := cfg.FontSize * 1.2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(x, bounds.Y+(bounds.Height-th)/2, tw, th),
				FillColor: color,
				Corners:   uimath.CornersAll(2),
			}, 0, 1)
			x += tw
		}
		if !isLast {
			sepClr := uimath.RGBA(0, 0, 0, 0.45)
			if cfg.TextRenderer != nil {
				sw := cfg.TextRenderer.MeasureText(b.separator, cfg.FontSize)
				lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
				cfg.TextRenderer.DrawText(buf, b.separator, x+cfg.SpaceXS, bounds.Y+(bounds.Height-lh)/2, cfg.FontSize, sw+1, sepClr, 1)
				x += sw + cfg.SpaceXS*2
			} else {
				x += cfg.SpaceSM
			}
		}
	}
}

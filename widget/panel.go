package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Panel is a titled container with border.
type Panel struct {
	Base
	title    string
	bgColor  uimath.Color
	bordered bool
}

func NewPanel(tree *core.Tree, title string, cfg *Config) *Panel {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	p := &Panel{
		Base:     NewBase(tree, core.TypeDiv, cfg),
		title:    title,
		bgColor:  uimath.ColorWhite,
		bordered: true,
	}
	p.style.Display = layout.DisplayFlex
	p.style.FlexDirection = layout.FlexDirectionColumn
	return p
}

func (p *Panel) SetTitle(t string)          { p.title = t }
func (p *Panel) SetBgColor(c uimath.Color) { p.bgColor = c }
func (p *Panel) SetBordered(b bool)         { p.bordered = b }

func (p *Panel) Draw(buf *render.CommandBuffer) {
	bounds := p.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := p.config

	bw := float32(0)
	bc := uimath.Color{}
	if p.bordered {
		bw = cfg.BorderWidth
		bc = cfg.BorderColor
	}

	buf.DrawRect(render.RectCmd{
		Bounds:      bounds,
		FillColor:   p.bgColor,
		BorderColor: bc,
		BorderWidth: bw,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 0, 1)

	if p.title != "" {
		headerH := float32(40)
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(bounds.X, bounds.Y+headerH-1, bounds.Width, 1),
			FillColor: uimath.RGBA(0, 0, 0, 0.06),
		}, 1, 1)
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			cfg.TextRenderer.DrawText(buf, p.title, bounds.X+cfg.SpaceMD, bounds.Y+(headerH-lh)/2, cfg.FontSize, bounds.Width-cfg.SpaceMD*2, cfg.TextColor, 1)
		} else {
			tw := float32(len(p.title)) * cfg.FontSize * 0.55
			th := cfg.FontSize * 1.2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X+cfg.SpaceMD, bounds.Y+(headerH-th)/2, tw, th),
				FillColor: cfg.TextColor,
				Corners:   uimath.CornersAll(2),
			}, 1, 1)
		}
	}
	p.DrawChildren(buf)
}

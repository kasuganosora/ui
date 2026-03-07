package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// DividerDirection controls orientation.
type DividerDirection uint8

const (
	DividerHorizontal DividerDirection = iota
	DividerVertical
)

// Divider draws a horizontal or vertical separator line.
type Divider struct {
	Base
	direction DividerDirection
	color     uimath.Color
	thickness float32
	text      string
}

func NewDivider(tree *core.Tree, cfg *Config) *Divider {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Divider{
		Base:      NewBase(tree, core.TypeCustom, cfg),
		color:     uimath.RGBA(0, 0, 0, 0.06),
		thickness: 1,
	}
}

func (d *Divider) SetDirection(dir DividerDirection) { d.direction = dir }
func (d *Divider) SetColor(c uimath.Color)           { d.color = c }
func (d *Divider) SetThickness(t float32)             { d.thickness = t }
func (d *Divider) SetText(t string)                   { d.text = t }

func (d *Divider) Draw(buf *render.CommandBuffer) {
	bounds := d.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := d.config
	if d.direction == DividerVertical {
		x := bounds.X + bounds.Width/2 - d.thickness/2
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x, bounds.Y, d.thickness, bounds.Height),
			FillColor: d.color,
		}, 0, 1)
	} else {
		y := bounds.Y + bounds.Height/2 - d.thickness/2
		if d.text == "" {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X, y, bounds.Width, d.thickness),
				FillColor: d.color,
			}, 0, 1)
		} else {
			// Line + centered text
			textW := float32(len(d.text)) * cfg.FontSize * 0.55
			if cfg.TextRenderer != nil {
				textW = cfg.TextRenderer.MeasureText(d.text, cfg.FontSizeSm)
			}
			gap := float32(8)
			lineL := (bounds.Width - textW - gap*2) / 2
			lineR := lineL
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X, y, lineL, d.thickness),
				FillColor: d.color,
			}, 0, 1)
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X+bounds.Width-lineR, y, lineR, d.thickness),
				FillColor: d.color,
			}, 0, 1)
			if cfg.TextRenderer != nil {
				lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
				cfg.TextRenderer.DrawText(buf, d.text, bounds.X+lineL+gap, bounds.Y+(bounds.Height-lh)/2, cfg.FontSizeSm, textW, uimath.RGBA(0, 0, 0, 0.45), 1)
			}
		}
	}
}

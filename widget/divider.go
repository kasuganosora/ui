package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// DividerLayout controls orientation.
type DividerLayout uint8

const (
	DividerHorizontal DividerLayout = iota
	DividerVertical
)

// DividerAlign controls text position on the divider line.
type DividerAlign uint8

const (
	AlignCenter DividerAlign = iota
	AlignLeft
	AlignRight
)

// Divider draws a horizontal or vertical separator line.
type Divider struct {
	Base
	layout    DividerLayout
	color     uimath.Color
	thickness float32
	content   string
	dashed    bool
	align     DividerAlign
	size      float32
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

func (d *Divider) SetLayout(dir DividerLayout)  { d.layout = dir }
func (d *Divider) SetColor(c uimath.Color)      { d.color = c }
func (d *Divider) SetThickness(t float32)        { d.thickness = t }
func (d *Divider) SetContent(t string)           { d.content = t }
func (d *Divider) SetDashed(dashed bool)         { d.dashed = dashed }
func (d *Divider) SetAlign(align DividerAlign)   { d.align = align }
func (d *Divider) SetSize(s float32)             { d.size = s }

// drawDashedLineH draws a horizontal dashed line as repeated 6px segments with 4px gaps.
func (d *Divider) drawDashedLineH(buf *render.CommandBuffer, x, y, width, thickness float32, color uimath.Color, zIndex int32) {
	segLen := float32(6)
	gapLen := float32(4)
	cx := x
	for cx < x+width {
		sw := segLen
		if cx+sw > x+width {
			sw = x + width - cx
		}
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(cx, y, sw, thickness),
			FillColor: color,
		}, zIndex, 1)
		cx += segLen + gapLen
	}
}

// drawDashedLineV draws a vertical dashed line as repeated 6px segments with 4px gaps.
func (d *Divider) drawDashedLineV(buf *render.CommandBuffer, x, y, height, thickness float32, color uimath.Color, zIndex int32) {
	segLen := float32(6)
	gapLen := float32(4)
	cy := y
	for cy < y+height {
		sh := segLen
		if cy+sh > y+height {
			sh = y + height - cy
		}
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x, cy, thickness, sh),
			FillColor: color,
		}, zIndex, 1)
		cy += segLen + gapLen
	}
}

func (d *Divider) Draw(buf *render.CommandBuffer) {
	bounds := d.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := d.config
	if d.layout == DividerVertical {
		x := bounds.X + bounds.Width/2 - d.thickness/2
		if d.dashed {
			d.drawDashedLineV(buf, x, bounds.Y, bounds.Height, d.thickness, d.color, 0)
		} else {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(x, bounds.Y, d.thickness, bounds.Height),
				FillColor: d.color,
			}, 0, 1)
		}
	} else {
		y := bounds.Y + bounds.Height/2 - d.thickness/2
		if d.content == "" {
			// No text — just a line
			if d.dashed {
				d.drawDashedLineH(buf, bounds.X, y, bounds.Width, d.thickness, d.color, 0)
			} else {
				buf.DrawRect(render.RectCmd{
					Bounds:    uimath.NewRect(bounds.X, y, bounds.Width, d.thickness),
					FillColor: d.color,
				}, 0, 1)
			}
		} else {
			// Line + text with alignment
			textW := float32(len(d.content)) * cfg.FontSize * 0.55
			if cfg.TextRenderer != nil {
				textW = cfg.TextRenderer.MeasureText(d.content, cfg.FontSizeSm)
			}
			gap := float32(8)
			indent := float32(16)

			var lineL, lineR, textX float32
			switch d.align {
			case AlignLeft:
				lineL = indent
				textX = bounds.X + lineL + gap
				lineR = bounds.Width - lineL - gap*2 - textW
			case AlignRight:
				lineR = indent
				lineL = bounds.Width - lineR - gap*2 - textW
				textX = bounds.X + lineL + gap
			default: // AlignCenter
				lineL = (bounds.Width - textW - gap*2) / 2
				lineR = lineL
				textX = bounds.X + lineL + gap
			}

			if lineL < 0 {
				lineL = 0
			}
			if lineR < 0 {
				lineR = 0
			}

			// Left line segment
			if lineL > 0 {
				if d.dashed {
					d.drawDashedLineH(buf, bounds.X, y, lineL, d.thickness, d.color, 0)
				} else {
					buf.DrawRect(render.RectCmd{
						Bounds:    uimath.NewRect(bounds.X, y, lineL, d.thickness),
						FillColor: d.color,
					}, 0, 1)
				}
			}

			// Right line segment
			if lineR > 0 {
				rightX := bounds.X + bounds.Width - lineR
				if d.dashed {
					d.drawDashedLineH(buf, rightX, y, lineR, d.thickness, d.color, 0)
				} else {
					buf.DrawRect(render.RectCmd{
						Bounds:    uimath.NewRect(rightX, y, lineR, d.thickness),
						FillColor: d.color,
					}, 0, 1)
				}
			}

			// Text
			if cfg.TextRenderer != nil {
				lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
				cfg.TextRenderer.DrawText(buf, d.content, textX, bounds.Y+(bounds.Height-lh)/2, cfg.FontSizeSm, textW, uimath.RGBA(0, 0, 0, 0.45), 1)
			}
		}
	}
}

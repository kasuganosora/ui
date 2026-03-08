package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Rate is a star rating input.
type Rate struct {
	Base
	value     float32
	count     int
	size      float32
	gap       float32
	disabled  bool
	allowHalf bool
	clearable bool
	texts     []string
	showText  bool
	color     [2]uimath.Color // [active, inactive]
	onChange  func(float32)
}

func NewRate(tree *core.Tree, cfg *Config) *Rate {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	r := &Rate{
		Base:     NewBase(tree, core.TypeCustom, cfg),
		count:    5,
		size:     24,
		gap:      4,
		color:    [2]uimath.Color{uimath.ColorHex("#fadb14"), cfg.BorderColor},
	}
	tree.AddHandler(r.id, event.MouseClick, func(e *event.Event) {
		if r.disabled {
			return
		}
		bounds := r.Bounds()
		relX := e.GlobalX - bounds.X
		starW := r.size + r.gap
		idx := int(relX / starW)
		if idx >= 0 && idx < r.count {
			var newVal float32
			if r.allowHalf {
				// Left half of star = x.5, right half = x+1
				posInStar := relX - float32(idx)*starW
				if posInStar < r.size/2 {
					newVal = float32(idx) + 0.5
				} else {
					newVal = float32(idx) + 1
				}
			} else {
				newVal = float32(idx) + 1
			}
			if r.clearable && newVal == r.value {
				r.SetValue(0)
			} else {
				r.SetValue(newVal)
			}
		}
	})
	return r
}

func (r *Rate) Value() float32                     { return r.value }
func (r *Rate) SetCount(c int)                     { r.count = c }
func (r *Rate) SetSize(s float32)                  { r.size = s }
func (r *Rate) SetDisabled(d bool)                 { r.disabled = d }
func (r *Rate) SetAllowHalf(v bool)                { r.allowHalf = v }
func (r *Rate) SetClearable(v bool)                { r.clearable = v }
func (r *Rate) Clearable() bool                    { return r.clearable }
func (r *Rate) SetTexts(t []string)                { r.texts = t }
func (r *Rate) SetShowText(v bool)                 { r.showText = v }
func (r *Rate) SetColor(c [2]uimath.Color)         { r.color = c }
func (r *Rate) OnChange(fn func(float32))          { r.onChange = fn }

func (r *Rate) SetValue(v float32) {
	if v < 0 {
		v = 0
	}
	max := float32(r.count)
	if v > max {
		v = max
	}
	if r.allowHalf {
		// Snap to nearest 0.5
		v = float32(int(v*2+0.5)) / 2
	} else {
		v = float32(int(v + 0.5))
	}
	if v != r.value {
		r.value = v
		if r.onChange != nil {
			r.onChange(v)
		}
	}
}

func (r *Rate) drawStar(buf *render.CommandBuffer, x, y, s float32, color uimath.Color, z int32) {
	if r.config.TextRenderer != nil {
		// Use text glyph for star
		r.config.TextRenderer.DrawText(buf, "\u2605", x, y, s, s, color, 1)
	} else {
		// Fallback: diamond (rotated square approximation)
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x+s*0.15, y+s*0.15, s*0.7, s*0.7),
			FillColor: color,
			Corners:   uimath.CornersAll(s * 0.15),
		}, z, 1)
	}
}

func (r *Rate) Draw(buf *render.CommandBuffer) {
	bounds := r.Bounds()
	if bounds.IsEmpty() {
		return
	}
	activeColor := r.color[0]
	inactiveColor := r.color[1]

	for i := 0; i < r.count; i++ {
		x := bounds.X + float32(i)*(r.size+r.gap)
		y := bounds.Y + (bounds.Height-r.size)/2
		fi := float32(i)

		if r.allowHalf && r.value > fi && r.value < fi+1 {
			// Half-filled star: left half active, right half inactive
			// Draw inactive full star first
			r.drawStar(buf, x, y, r.size, inactiveColor, 1)
			// Clip left half with a colored rect overlay
			halfW := r.size / 2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(x, y, halfW, r.size),
				FillColor: activeColor,
				Corners:   uimath.CornersAll(0),
			}, 2, 1)
		} else if r.value >= fi+1 {
			// Fully active
			r.drawStar(buf, x, y, r.size, activeColor, 1)
		} else {
			// Inactive
			r.drawStar(buf, x, y, r.size, inactiveColor, 1)
		}
	}

	// Draw text label after stars
	if r.showText && len(r.texts) > 0 {
		idx := int(r.value) - 1
		if idx < 0 {
			idx = 0
		}
		if idx >= len(r.texts) {
			idx = len(r.texts) - 1
		}
		label := r.texts[idx]
		if label != "" {
			textX := bounds.X + float32(r.count)*(r.size+r.gap) + r.gap
			cfg := r.config
			if cfg.TextRenderer != nil {
				lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
				textY := bounds.Y + (bounds.Height-lh)/2
				maxW := bounds.Width - (textX - bounds.X)
				cfg.TextRenderer.DrawText(buf, label, textX, textY, cfg.FontSize, maxW, cfg.TextColor, 1)
			} else {
				tw := float32(len(label)) * cfg.FontSize * 0.55
				th := cfg.FontSize * 1.2
				textY := bounds.Y + (bounds.Height-th)/2
				buf.DrawRect(render.RectCmd{
					Bounds:    uimath.NewRect(textX, textY, tw, th),
					FillColor: cfg.TextColor,
					Corners:   uimath.CornersAll(2),
				}, 1, 0.5)
			}
		}
	}
}

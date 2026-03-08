package widget

import (
	"fmt"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// SliderLayout controls the orientation of the slider.
type SliderLayout uint8

const (
	// SliderHorizontal is the default horizontal slider layout.
	SliderHorizontal SliderLayout = iota
	// SliderVertical renders the slider vertically.
	SliderVertical
)

// Slider is a horizontal slider for selecting a numeric value.
type Slider struct {
	Base
	value       float32
	min         float32
	max         float32
	step        float32
	disabled    bool
	dragging    bool
	showTooltip bool
	layout      SliderLayout
	rangeMode   bool
	showStep    bool
	marks       map[float32]string
	onChange    func(float32)
	onChangeEnd func(float32)
}

func NewSlider(tree *core.Tree, cfg *Config) *Slider {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	s := &Slider{
		Base:        NewBase(tree, core.TypeCustom, cfg),
		min:         0,
		max:         100,
		step:        1,
		showTooltip: true,
	}
	tree.AddHandler(s.id, event.MouseDown, func(e *event.Event) {
		if !s.disabled {
			s.dragging = true
			s.updateFromMouse(e.GlobalX)
		}
	})
	tree.AddHandler(s.id, event.MouseMove, func(e *event.Event) {
		if s.dragging {
			s.updateFromMouse(e.GlobalX)
		}
	})
	tree.AddHandler(s.id, event.MouseUp, func(e *event.Event) {
		if s.dragging {
			s.dragging = false
			if s.onChangeEnd != nil {
				s.onChangeEnd(s.value)
			}
		}
	})
	return s
}

func (s *Slider) Value() float32                  { return s.value }
func (s *Slider) Min() float32                    { return s.min }
func (s *Slider) Max() float32                    { return s.max }
func (s *Slider) SetMin(v float32)                { s.min = v }
func (s *Slider) SetMax(v float32)                { s.max = v }
func (s *Slider) SetStep(v float32)               { s.step = v }
func (s *Slider) SetDisabled(d bool)              { s.disabled = d }
func (s *Slider) SetShowTooltip(v bool)           { s.showTooltip = v }
func (s *Slider) SetLayout(l SliderLayout)        { s.layout = l }
func (s *Slider) Layout() SliderLayout            { return s.layout }
func (s *Slider) SetRange(v bool)                 { s.rangeMode = v }
func (s *Slider) Range() bool                     { return s.rangeMode }
func (s *Slider) SetShowStep(v bool)              { s.showStep = v }
func (s *Slider) ShowStep() bool                  { return s.showStep }
func (s *Slider) SetMarks(m map[float32]string)   { s.marks = m }
func (s *Slider) OnChange(fn func(float32))       { s.onChange = fn }
func (s *Slider) OnChangeEnd(fn func(float32))    { s.onChangeEnd = fn }

func (s *Slider) SetValue(v float32) {
	if v < s.min {
		v = s.min
	}
	if v > s.max {
		v = s.max
	}
	if s.step > 0 {
		v = s.min + float32(int((v-s.min)/s.step+0.5))*s.step
	}
	if v != s.value {
		s.value = v
		if s.onChange != nil {
			s.onChange(v)
		}
	}
}

func (s *Slider) updateFromMouse(mx float32) {
	bounds := s.Bounds()
	if bounds.Width <= 0 {
		return
	}
	t := (mx - bounds.X) / bounds.Width
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	s.SetValue(s.min + t*(s.max-s.min))
}

func (s *Slider) ratio() float32 {
	if s.max <= s.min {
		return 0
	}
	r := (s.value - s.min) / (s.max - s.min)
	if r < 0 {
		return 0
	}
	if r > 1 {
		return 1
	}
	return r
}

func (s *Slider) formatValue(v float32) string {
	if v == float32(int(v)) {
		return fmt.Sprintf("%.0f", v)
	}
	return fmt.Sprintf("%.1f", v)
}

func (s *Slider) Draw(buf *render.CommandBuffer) {
	bounds := s.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := s.config
	trackH := float32(4)
	trackY := bounds.Y + (bounds.Height-trackH)/2

	// Track background
	trackColor := uimath.RGBA(0, 0, 0, 0.06)
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(bounds.X, trackY, bounds.Width, trackH),
		FillColor: trackColor,
		Corners:   uimath.CornersAll(trackH / 2),
	}, 0, 1)

	// Filled portion
	r := s.ratio()
	fillColor := cfg.PrimaryColor
	if s.disabled {
		fillColor = cfg.DisabledColor
	}
	if r > 0 {
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(bounds.X, trackY, bounds.Width*r, trackH),
			FillColor: fillColor,
			Corners:   uimath.CornersAll(trackH / 2),
		}, 1, 1)
	}

	// Thumb
	thumbR := float32(7)
	thumbX := bounds.X + bounds.Width*r - thumbR
	thumbY := bounds.Y + (bounds.Height-thumbR*2)/2
	thumbBorder := cfg.PrimaryColor
	if s.disabled {
		thumbBorder = cfg.DisabledColor
	}
	buf.DrawRect(render.RectCmd{
		Bounds:      uimath.NewRect(thumbX, thumbY, thumbR*2, thumbR*2),
		FillColor:   uimath.ColorWhite,
		BorderColor: thumbBorder,
		BorderWidth: 2,
		Corners:     uimath.CornersAll(thumbR),
	}, 2, 1)

	// Marks: draw ticks and labels below the track
	if len(s.marks) > 0 {
		markTickH := float32(6)
		markTickW := float32(2)
		markTickY := trackY + trackH + 2
		for markVal, label := range s.marks {
			if markVal < s.min || markVal > s.max {
				continue
			}
			mr := (markVal - s.min) / (s.max - s.min)
			mx := bounds.X + bounds.Width*mr - markTickW/2
			// Tick mark
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(mx, markTickY, markTickW, markTickH),
				FillColor: cfg.BorderColor,
			}, 1, 1)
			// Label text below tick
			if label != "" {
				labelY := markTickY + markTickH + 2
				if cfg.TextRenderer != nil {
					tw := cfg.TextRenderer.MeasureText(label, cfg.FontSizeSm)
					lx := bounds.X + bounds.Width*mr - tw/2
					cfg.TextRenderer.DrawText(buf, label, lx, labelY, cfg.FontSizeSm, tw+4, cfg.TextColor, 1)
				} else {
					tw := float32(len(label)) * cfg.FontSizeSm * 0.55
					th := cfg.FontSizeSm * 1.2
					lx := bounds.X + bounds.Width*mr - tw/2
					buf.DrawRect(render.RectCmd{
						Bounds:    uimath.NewRect(lx, labelY, tw, th),
						FillColor: cfg.TextColor,
						Corners:   uimath.CornersAll(1),
					}, 1, 0.5)
				}
			}
		}
	}

	// Tooltip: show value above thumb while dragging
	if s.dragging && s.showTooltip {
		valStr := s.formatValue(s.value)
		tooltipH := float32(24)
		tooltipPadX := float32(8)
		tooltipGap := float32(6)
		var tooltipW float32
		if cfg.TextRenderer != nil {
			tooltipW = cfg.TextRenderer.MeasureText(valStr, cfg.FontSizeSm) + tooltipPadX*2
		} else {
			tooltipW = float32(len(valStr))*cfg.FontSizeSm*0.55 + tooltipPadX*2
		}
		tooltipX := thumbX + thumbR - tooltipW/2
		tooltipY := thumbY - tooltipH - tooltipGap
		bgColor := uimath.RGBA(0, 0, 0, 0.75)
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(tooltipX, tooltipY, tooltipW, tooltipH),
			FillColor: bgColor,
			Corners:   uimath.CornersAll(cfg.BorderRadius),
		}, 20, 1)
		// Tooltip text
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			ty := tooltipY + (tooltipH-lh)/2
			cfg.TextRenderer.DrawText(buf, valStr, tooltipX+tooltipPadX, ty, cfg.FontSizeSm, tooltipW, uimath.ColorWhite, 1)
		} else {
			tw := float32(len(valStr)) * cfg.FontSizeSm * 0.55
			th := cfg.FontSizeSm * 1.2
			ty := tooltipY + (tooltipH-th)/2
			tx := tooltipX + (tooltipW-tw)/2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(tx, ty, tw, th),
				FillColor: uimath.ColorWhite,
				Corners:   uimath.CornersAll(2),
			}, 21, 1)
		}
	}
}

package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Slider is a horizontal slider for selecting a numeric value.
type Slider struct {
	Base
	value    float32
	min      float32
	max      float32
	step     float32
	disabled bool
	dragging bool
	onChange func(float32)
}

func NewSlider(tree *core.Tree, cfg *Config) *Slider {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	s := &Slider{
		Base: NewBase(tree, core.TypeCustom, cfg),
		min:  0,
		max:  100,
		step: 1,
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
		s.dragging = false
	})
	return s
}

func (s *Slider) Value() float32        { return s.value }
func (s *Slider) Min() float32          { return s.min }
func (s *Slider) Max() float32          { return s.max }
func (s *Slider) SetMin(v float32)      { s.min = v }
func (s *Slider) SetMax(v float32)      { s.max = v }
func (s *Slider) SetStep(v float32)     { s.step = v }
func (s *Slider) SetDisabled(d bool)    { s.disabled = d }
func (s *Slider) OnChange(fn func(float32)) { s.onChange = fn }

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
}

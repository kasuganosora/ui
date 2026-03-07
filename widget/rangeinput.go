package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// RangeInput is a dual-thumb range slider.
type RangeInput struct {
	Base
	min       float32
	max       float32
	low       float32
	high      float32
	step      float32
	dragging  int // 0=none, 1=low, 2=high
	onChange  func(low, high float32)
}

func NewRangeInput(tree *core.Tree, min, max float32, cfg *Config) *RangeInput {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	ri := &RangeInput{
		Base: NewBase(tree, core.TypeCustom, cfg),
		min:  min,
		max:  max,
		low:  min,
		high: max,
		step: 1,
	}
	tree.AddHandler(ri.id, event.MouseDown, func(e *event.Event) {
		bounds := ri.Bounds()
		if bounds.IsEmpty() {
			return
		}
		relX := e.GlobalX - bounds.X
		frac := relX / bounds.Width
		val := ri.min + frac*(ri.max-ri.min)
		distLow := val - ri.low
		if distLow < 0 {
			distLow = -distLow
		}
		distHigh := val - ri.high
		if distHigh < 0 {
			distHigh = -distHigh
		}
		if distLow <= distHigh {
			ri.dragging = 1
		} else {
			ri.dragging = 2
		}
	})
	tree.AddHandler(ri.id, event.MouseMove, func(e *event.Event) {
		if ri.dragging == 0 {
			return
		}
		bounds := ri.Bounds()
		frac := (e.GlobalX - bounds.X) / bounds.Width
		if frac < 0 {
			frac = 0
		}
		if frac > 1 {
			frac = 1
		}
		val := ri.min + frac*(ri.max-ri.min)
		if ri.step > 0 {
			val = ri.min + float32(int((val-ri.min)/ri.step+0.5))*ri.step
		}
		if ri.dragging == 1 {
			if val > ri.high {
				val = ri.high
			}
			ri.low = val
		} else {
			if val < ri.low {
				val = ri.low
			}
			ri.high = val
		}
		if ri.onChange != nil {
			ri.onChange(ri.low, ri.high)
		}
	})
	tree.AddHandler(ri.id, event.MouseUp, func(e *event.Event) { ri.dragging = 0 })
	return ri
}

func (ri *RangeInput) Low() float32                       { return ri.low }
func (ri *RangeInput) High() float32                      { return ri.high }
func (ri *RangeInput) SetRange(low, high float32)         { ri.low = low; ri.high = high }
func (ri *RangeInput) SetStep(s float32)                  { ri.step = s }
func (ri *RangeInput) OnChange(fn func(float32, float32)) { ri.onChange = fn }

func (ri *RangeInput) Draw(buf *render.CommandBuffer) {
	bounds := ri.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := ri.config
	trackH := float32(4)
	thumbSize := float32(16)
	trackY := bounds.Y + (bounds.Height-trackH)/2
	rng := ri.max - ri.min
	if rng <= 0 {
		rng = 1
	}

	// Track background
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(bounds.X, trackY, bounds.Width, trackH),
		FillColor: uimath.RGBA(0, 0, 0, 0.06),
		Corners:   uimath.CornersAll(trackH / 2),
	}, 1, 1)

	// Active range
	lowFrac := (ri.low - ri.min) / rng
	highFrac := (ri.high - ri.min) / rng
	activeX := bounds.X + lowFrac*bounds.Width
	activeW := (highFrac - lowFrac) * bounds.Width
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(activeX, trackY, activeW, trackH),
		FillColor: cfg.PrimaryColor,
		Corners:   uimath.CornersAll(trackH / 2),
	}, 2, 1)

	// Thumbs
	for _, frac := range []float32{lowFrac, highFrac} {
		tx := bounds.X + frac*bounds.Width - thumbSize/2
		ty := bounds.Y + (bounds.Height-thumbSize)/2
		buf.DrawRect(render.RectCmd{
			Bounds:      uimath.NewRect(tx, ty, thumbSize, thumbSize),
			FillColor:   uimath.ColorWhite,
			BorderColor: cfg.PrimaryColor,
			BorderWidth: 2,
			Corners:     uimath.CornersAll(thumbSize / 2),
		}, 3, 1)
	}
}

package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// SplitterDirection controls the split orientation.
type SplitterDirection uint8

const (
	SplitterHorizontal SplitterDirection = iota
	SplitterVertical
)

// Splitter divides space into two resizable panels.
type Splitter struct {
	Base
	direction  SplitterDirection
	ratio      float32 // 0-1 split position
	minRatio   float32
	maxRatio   float32
	barSize    float32
	dragging   bool
	first      Widget
	second     Widget
}

func NewSplitter(tree *core.Tree, cfg *Config) *Splitter {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	s := &Splitter{
		Base:     NewBase(tree, core.TypeCustom, cfg),
		ratio:    0.5,
		minRatio: 0.1,
		maxRatio: 0.9,
		barSize:  4,
	}
	tree.AddHandler(s.id, event.MouseDown, func(e *event.Event) { s.dragging = true })
	tree.AddHandler(s.id, event.MouseUp, func(e *event.Event) { s.dragging = false })
	tree.AddHandler(s.id, event.MouseMove, func(e *event.Event) {
		if s.dragging {
			s.updateFromMouse(e.GlobalX, e.GlobalY)
		}
	})
	return s
}

func (s *Splitter) SetDirection(d SplitterDirection) { s.direction = d }
func (s *Splitter) Ratio() float32                   { return s.ratio }
func (s *Splitter) SetRatio(r float32)               { s.ratio = clampF(r, s.minRatio, s.maxRatio) }
func (s *Splitter) SetMinRatio(r float32)            { s.minRatio = r }
func (s *Splitter) SetMaxRatio(r float32)            { s.maxRatio = r }
func (s *Splitter) SetFirst(w Widget)                { s.first = w }
func (s *Splitter) SetSecond(w Widget)               { s.second = w }

func (s *Splitter) updateFromMouse(mx, my float32) {
	bounds := s.Bounds()
	if s.direction == SplitterHorizontal {
		s.ratio = clampF((mx-bounds.X)/bounds.Width, s.minRatio, s.maxRatio)
	} else {
		s.ratio = clampF((my-bounds.Y)/bounds.Height, s.minRatio, s.maxRatio)
	}
}

func (s *Splitter) Draw(buf *render.CommandBuffer) {
	bounds := s.Bounds()
	if bounds.IsEmpty() {
		return
	}

	barClr := uimath.RGBA(0, 0, 0, 0.06)
	if s.dragging {
		barClr = uimath.RGBA(0, 0, 0, 0.15)
	}

	if s.direction == SplitterHorizontal {
		barX := bounds.X + bounds.Width*s.ratio - s.barSize/2
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(barX, bounds.Y, s.barSize, bounds.Height),
			FillColor: barClr,
		}, 1, 1)
	} else {
		barY := bounds.Y + bounds.Height*s.ratio - s.barSize/2
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(bounds.X, barY, bounds.Width, s.barSize),
			FillColor: barClr,
		}, 1, 1)
	}

	if s.first != nil {
		s.first.Draw(buf)
	}
	if s.second != nil {
		s.second.Draw(buf)
	}
}

func clampF(v, min, max float32) float32 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

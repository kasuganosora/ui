package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Swiper is a carousel/slider that cycles through content panels.
type Swiper struct {
	Base
	panels     []Widget
	current    int
	autoplay   bool
	showDots   bool
	onChange   func(int)
}

func NewSwiper(tree *core.Tree, cfg *Config) *Swiper {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Swiper{
		Base:     NewBase(tree, core.TypeCustom, cfg),
		showDots: true,
	}
}

func (s *Swiper) Current() int           { return s.current }
func (s *Swiper) PanelCount() int        { return len(s.panels) }
func (s *Swiper) SetCurrent(c int)       { s.current = c }
func (s *Swiper) SetAutoplay(a bool)     { s.autoplay = a }
func (s *Swiper) SetShowDots(d bool)     { s.showDots = d }
func (s *Swiper) OnChange(fn func(int))  { s.onChange = fn }

func (s *Swiper) AddPanel(w Widget) {
	s.panels = append(s.panels, w)
}

func (s *Swiper) Next() {
	if len(s.panels) == 0 {
		return
	}
	s.current = (s.current + 1) % len(s.panels)
	if s.onChange != nil {
		s.onChange(s.current)
	}
}

func (s *Swiper) Prev() {
	if len(s.panels) == 0 {
		return
	}
	s.current--
	if s.current < 0 {
		s.current = len(s.panels) - 1
	}
	if s.onChange != nil {
		s.onChange(s.current)
	}
}

func (s *Swiper) Draw(buf *render.CommandBuffer) {
	bounds := s.Bounds()
	if bounds.IsEmpty() || len(s.panels) == 0 {
		return
	}
	cfg := s.config

	// Current panel
	if s.current >= 0 && s.current < len(s.panels) {
		s.panels[s.current].Draw(buf)
	}

	// Dots indicator
	if s.showDots && len(s.panels) > 1 {
		dotSize := float32(8)
		dotGap := float32(6)
		totalW := float32(len(s.panels))*(dotSize+dotGap) - dotGap
		dx := bounds.X + (bounds.Width-totalW)/2
		dy := bounds.Y + bounds.Height - dotSize - cfg.SpaceSM

		for i := range s.panels {
			color := uimath.RGBA(0, 0, 0, 0.2)
			if i == s.current {
				color = cfg.PrimaryColor
			}
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(dx+float32(i)*(dotSize+dotGap), dy, dotSize, dotSize),
				FillColor: color,
				Corners:   uimath.CornersAll(dotSize / 2),
			}, 3, 1)
		}
	}
}

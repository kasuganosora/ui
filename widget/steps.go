package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// StepStatus indicates the state of a step.
type StepStatus uint8

const (
	StepWait    StepStatus = iota
	StepProcess
	StepFinish
	StepError
)

// StepItem represents a single step.
type StepItem struct {
	Title       string
	Description string
}

// Steps displays a navigation progress bar.
type Steps struct {
	Base
	items   []StepItem
	current int
}

func NewSteps(tree *core.Tree, cfg *Config) *Steps {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Steps{
		Base: NewBase(tree, core.TypeCustom, cfg),
	}
}

func (s *Steps) Current() int          { return s.current }
func (s *Steps) SetCurrent(c int)      { s.current = c }
func (s *Steps) Items() []StepItem     { return s.items }

func (s *Steps) AddStep(item StepItem) {
	s.items = append(s.items, item)
}

func (s *Steps) ClearSteps() {
	s.items = s.items[:0]
	s.current = 0
}

func (s *Steps) Draw(buf *render.CommandBuffer) {
	bounds := s.Bounds()
	if bounds.IsEmpty() || len(s.items) == 0 {
		return
	}
	cfg := s.config
	n := len(s.items)
	stepW := bounds.Width / float32(n)
	dotSize := float32(24)
	dotR := dotSize / 2

	for i, item := range s.items {
		cx := bounds.X + float32(i)*stepW + stepW/2
		cy := bounds.Y + dotR

		// Connector line
		if i > 0 {
			prevCx := bounds.X + float32(i-1)*stepW + stepW/2
			lineColor := cfg.BorderColor
			if i <= s.current {
				lineColor = cfg.PrimaryColor
			}
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(prevCx+dotR+2, cy-0.5, cx-prevCx-dotSize-4, 1),
				FillColor: lineColor,
			}, 1, 1)
		}

		// Dot
		dotColor := cfg.BorderColor
		if i < s.current {
			dotColor = cfg.PrimaryColor
		} else if i == s.current {
			dotColor = cfg.PrimaryColor
		}
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(cx-dotR, cy-dotR, dotSize, dotSize),
			FillColor: dotColor,
			Corners:   uimath.CornersAll(dotR),
		}, 2, 1)

		// Step number
		if cfg.TextRenderer != nil {
			num := intToStr(i + 1)
			tw := cfg.TextRenderer.MeasureText(num, cfg.FontSizeSm)
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			cfg.TextRenderer.DrawText(buf, num, cx-tw/2, cy-lh/2, cfg.FontSizeSm, dotSize, uimath.ColorWhite, 1)

			// Title below
			titleLh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			titleColor := cfg.TextColor
			if i > s.current {
				titleColor = cfg.DisabledColor
			}
			cfg.TextRenderer.DrawText(buf, item.Title, bounds.X+float32(i)*stepW, cy+dotR+cfg.SpaceXS, cfg.FontSizeSm, stepW, titleColor, 1)
			_ = titleLh
		}
	}
}

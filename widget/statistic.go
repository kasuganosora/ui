package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Statistic displays a numeric value with a label.
type Statistic struct {
	Base
	title  string
	value  string
	prefix string
	suffix string
	color  uimath.Color
}

func NewStatistic(tree *core.Tree, title, value string, cfg *Config) *Statistic {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Statistic{
		Base:  NewBase(tree, core.TypeCustom, cfg),
		title: title,
		value: value,
	}
}

func (s *Statistic) Title() string           { return s.title }
func (s *Statistic) Value() string           { return s.value }
func (s *Statistic) SetTitle(t string)       { s.title = t }
func (s *Statistic) SetValue(v string)       { s.value = v }
func (s *Statistic) SetPrefix(p string)      { s.prefix = p }
func (s *Statistic) SetSuffix(su string)     { s.suffix = su }
func (s *Statistic) SetColor(c uimath.Color) { s.color = c }

func (s *Statistic) Draw(buf *render.CommandBuffer) {
	bounds := s.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := s.config
	if cfg.TextRenderer == nil {
		return
	}

	// Title
	lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
	cfg.TextRenderer.DrawText(buf, s.title, bounds.X, bounds.Y, cfg.FontSizeSm, bounds.Width, cfg.DisabledColor, 1)

	// Value
	valText := s.prefix + s.value + s.suffix
	valSize := cfg.FontSize * 1.8
	color := cfg.TextColor
	if s.color.A > 0 {
		color = s.color
	}
	cfg.TextRenderer.DrawText(buf, valText, bounds.X, bounds.Y+lh+4, valSize, bounds.Width, color, 1)
}

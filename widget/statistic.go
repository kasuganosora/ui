package widget

import (
	"strconv"
	"strings"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Trend indicates the value direction.
type Trend uint8

const (
	TrendNone     Trend = iota
	TrendIncrease       // TDesign: "increase"
	TrendDecrease       // TDesign: "decrease"
)

// TrendPlacement controls where the trend arrow is shown.
type TrendPlacement uint8

const (
	TrendPlacementLeft TrendPlacement = iota
	TrendPlacementRight
)

// StatisticAnimation configures the count-up animation.
type StatisticAnimation struct {
	Duration  int     // milliseconds
	ValueFrom float64 // starting value
}

// Statistic displays a numeric value with a label.
type Statistic struct {
	Base
	title          string
	value          string
	prefix         string
	suffix         string
	color          uimath.Color
	trend          Trend
	trendPlacement TrendPlacement
	separator      string // thousands separator (default ",")
	decimalPlaces  int    // -1 means no formatting
	loading        bool   // show skeleton placeholder
	animation      *StatisticAnimation
	animationStart bool
	extra          string
	format         func(value float64) string
	unit           string
}

func NewStatistic(tree *core.Tree, title, value string, cfg *Config) *Statistic {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Statistic{
		Base:          NewBase(tree, core.TypeCustom, cfg),
		title:         title,
		value:         value,
		separator:     ",",
		decimalPlaces: -1,
	}
}

func (s *Statistic) Title() string              { return s.title }
func (s *Statistic) Value() string              { return s.value }
func (s *Statistic) SetTitle(t string)          { s.title = t }
func (s *Statistic) SetValue(v string)          { s.value = v }
func (s *Statistic) SetPrefix(p string)         { s.prefix = p }
func (s *Statistic) SetSuffix(su string)        { s.suffix = su }
func (s *Statistic) SetColor(c uimath.Color)    { s.color = c }
func (s *Statistic) SetTrend(t Trend)           { s.trend = t }
func (s *Statistic) SetSeparator(sep string)    { s.separator = sep }
func (s *Statistic) SetDecimalPlaces(d int)                  { s.decimalPlaces = d }
func (s *Statistic) SetLoading(l bool)                       { s.loading = l }
func (s *Statistic) SetAnimation(a *StatisticAnimation)      { s.animation = a }
func (s *Statistic) SetAnimationStart(v bool)                { s.animationStart = v }
func (s *Statistic) SetExtra(e string)                       { s.extra = e }
func (s *Statistic) SetFormat(fn func(float64) string)       { s.format = fn }
func (s *Statistic) SetTrendPlacement(tp TrendPlacement)     { s.trendPlacement = tp }
func (s *Statistic) SetUnit(u string)                        { s.unit = u }

// formatValue applies thousands separator and decimal formatting to the value.
func (s *Statistic) formatValue() string {
	val := s.value

	// Try to parse as a number and apply formatting
	if s.separator != "" || s.decimalPlaces >= 0 {
		// Try float first
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			if s.decimalPlaces >= 0 {
				val = strconv.FormatFloat(f, 'f', s.decimalPlaces, 64)
			}
			if s.separator != "" {
				val = addThousandsSeparator(val, s.separator)
			}
		}
	}

	return val
}

// addThousandsSeparator inserts sep every 3 digits in the integer part.
func addThousandsSeparator(numStr, sep string) string {
	parts := strings.SplitN(numStr, ".", 2)
	intPart := parts[0]

	negative := false
	if len(intPart) > 0 && intPart[0] == '-' {
		negative = true
		intPart = intPart[1:]
	}

	if len(intPart) <= 3 {
		result := numStr
		if negative {
			result = "-" + intPart
		} else {
			result = intPart
		}
		if len(parts) == 2 {
			result += "." + parts[1]
		}
		return result
	}

	var b strings.Builder
	remainder := len(intPart) % 3
	if remainder > 0 {
		b.WriteString(intPart[:remainder])
	}
	for i := remainder; i < len(intPart); i += 3 {
		if b.Len() > 0 {
			b.WriteString(sep)
		}
		b.WriteString(intPart[i : i+3])
	}

	result := b.String()
	if negative {
		result = "-" + result
	}
	if len(parts) == 2 {
		result += "." + parts[1]
	}
	return result
}

func (s *Statistic) Draw(buf *render.CommandBuffer) {
	bounds := s.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := s.config
	if cfg.TextRenderer == nil {
		return
	}

	// Loading skeleton
	if s.loading {
		// Title placeholder
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(bounds.X, bounds.Y, bounds.Width*0.4, cfg.FontSizeSm+2),
			FillColor: uimath.ColorHex("#dcdcdc"),
			Corners:   uimath.CornersAll(2),
		}, 0, 1)
		// Value placeholder
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(bounds.X, bounds.Y+cfg.FontSizeSm+8, bounds.Width*0.6, cfg.FontSize*1.8+2),
			FillColor: uimath.ColorHex("#dcdcdc"),
			Corners:   uimath.CornersAll(2),
		}, 0, 1)
		return
	}

	// Title
	lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
	cfg.TextRenderer.DrawText(buf, s.title, bounds.X, bounds.Y, cfg.FontSizeSm, bounds.Width, cfg.DisabledColor, 1)

	// Determine value color
	color := cfg.TextColor
	if s.color.A > 0 {
		color = s.color
	}

	valSize := cfg.FontSize * 1.8
	valY := bounds.Y + lh + 4
	valX := bounds.X

	// Trend arrow
	if s.trend != TrendNone {
		var arrow string
		var arrowColor uimath.Color
		if s.trend == TrendIncrease {
			arrow = "\u25B2" // ▲
			arrowColor = uimath.ColorHex("#2ba471") // green
		} else {
			arrow = "\u25BC" // ▼
			arrowColor = uimath.ColorHex("#d54941") // red
		}
		arrowSize := valSize * 0.7
		aw := cfg.TextRenderer.MeasureText(arrow, arrowSize)
		arrowLH := cfg.TextRenderer.LineHeight(arrowSize)
		// Vertically center arrow with value text
		arrowY := valY + (cfg.TextRenderer.LineHeight(valSize)-arrowLH)/2
		cfg.TextRenderer.DrawText(buf, arrow, valX, arrowY, arrowSize, aw+4, arrowColor, 1)
		valX += aw + 4
	}

	// Build display text: prefix + formatted value + suffix
	formattedVal := s.formatValue()
	valText := s.prefix + formattedVal + s.suffix

	cfg.TextRenderer.DrawText(buf, valText, valX, valY, valSize, bounds.Width-(valX-bounds.X), color, 1)
}

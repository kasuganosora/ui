package widget

import (
	"fmt"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// StepStatus indicates the state of a step (TDesign: StepStatus).
type StepStatus uint8

const (
	StepDefault StepStatus = iota // TDesign: 'default'
	StepProcess                   // TDesign: 'process'
	StepFinish                    // TDesign: 'finish'
	StepError                     // TDesign: 'error'
)

// StepsLayout controls the direction of the steps bar (TDesign: layout).
type StepsLayout int

const (
	StepsLayoutHorizontal StepsLayout = iota // default
	StepsLayoutVertical
)

// StepsSeparator controls the separator style (TDesign: separator).
type StepsSeparator int

const (
	StepsSeparatorLine   StepsSeparator = iota // default
	StepsSeparatorDashed
	StepsSeparatorArrow
)

// StepsSequence controls the display order (TDesign: sequence).
type StepsSequence int

const (
	StepsSequencePositive StepsSequence = iota // default
	StepsSequenceReverse
)

// StepsTheme controls the visual style (TDesign: theme).
type StepsTheme int

const (
	StepsThemeDefault StepsTheme = iota // default
	StepsThemeDot
)

// StepItem represents a single step (TDesign: TdStepItemProps).
type StepItem struct {
	Title   string     // step title (TDesign: title)
	Content string     // step description (TDesign: content)
	Extra   string     // extra content below description (TDesign: extra)
	Status  StepStatus // override auto status (TDesign: status)
	Value   string     // unique step identifier (TDesign: value)
}

// Steps displays a navigation progress bar (TDesign: TdStepsProps).
type Steps struct {
	Base
	options   []StepItem     // step items (TDesign: options)
	current   int            // current step index (TDesign: current)
	layout    StepsLayout    // horizontal or vertical (TDesign: layout)
	readonly  bool           // read-only mode (TDesign: readonly)
	separator StepsSeparator // separator style (TDesign: separator)
	sequence  StepsSequence  // display order (TDesign: sequence)
	theme     StepsTheme     // visual theme (TDesign: theme)
	onChange  func(current int, previous int) // TDesign: onChange
}

func NewSteps(tree *core.Tree, cfg *Config) *Steps {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Steps{
		Base: NewBase(tree, core.TypeCustom, cfg),
	}
}

func (s *Steps) Current() int            { return s.current }
func (s *Steps) SetCurrent(c int)        { s.current = c }
func (s *Steps) Options() []StepItem     { return s.options }
func (s *Steps) SetLayout(l StepsLayout) { s.layout = l }
func (s *Steps) Layout() StepsLayout     { return s.layout }
func (s *Steps) SetReadonly(v bool)       { s.readonly = v }
func (s *Steps) IsReadonly() bool         { return s.readonly }
func (s *Steps) SetSeparator(sep StepsSeparator) { s.separator = sep }
func (s *Steps) Separator() StepsSeparator       { return s.separator }
func (s *Steps) SetSequence(seq StepsSequence)    { s.sequence = seq }
func (s *Steps) Sequence() StepsSequence          { return s.sequence }
func (s *Steps) SetTheme(t StepsTheme)            { s.theme = t }
func (s *Steps) Theme() StepsTheme                { return s.theme }

// OnChange sets the callback for step changes (TDesign: onChange).
func (s *Steps) OnChange(fn func(current int, previous int)) { s.onChange = fn }

// SetOptions sets the step items (TDesign: options).
func (s *Steps) SetOptions(items []StepItem) {
	s.options = items
}

func (s *Steps) AddStep(item StepItem) {
	s.options = append(s.options, item)
}

func (s *Steps) ClearSteps() {
	s.options = s.options[:0]
	s.current = 0
}

// stepStatus returns the effective status for step i.
func (s *Steps) stepStatus(i int) StepStatus {
	item := s.options[i]
	if item.Status != StepDefault {
		return item.Status
	}
	if i < s.current {
		return StepFinish
	}
	if i == s.current {
		return StepProcess
	}
	return StepDefault
}

const (
	stepDotSize = float32(24)
	stepLineH   = float32(1)
)

func (s *Steps) Draw(buf *render.CommandBuffer) {
	bounds := s.Bounds()
	if bounds.IsEmpty() || len(s.options) == 0 {
		return
	}
	cfg := s.config
	n := len(s.options)
	stepW := bounds.Width / float32(n)
	dotR := stepDotSize / 2

	for i := range s.options {
		item := s.options[i]
		status := s.stepStatus(i)
		cx := bounds.X + float32(i)*stepW + stepW/2
		cy := bounds.Y + dotR + 4

		// --- Connector line to previous step ---
		if i > 0 {
			prevCx := bounds.X + float32(i-1)*stepW + stepW/2
			prevStatus := s.stepStatus(i - 1)
			lineColor := cfg.BorderColor
			if prevStatus == StepFinish {
				lineColor = cfg.PrimaryColor
			}
			lineX := prevCx + dotR + 4
			lineW := cx - prevCx - stepDotSize - 8
			if lineW > 0 {
				buf.DrawRect(render.RectCmd{
					Bounds:    uimath.NewRect(lineX, cy-stepLineH/2, lineW, stepLineH),
					FillColor: lineColor,
				}, 1, 1)
			}
		}

		// --- Dot ---
		switch status {
		case StepFinish:
			// Filled primary circle with checkmark
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(cx-dotR, cy-dotR, stepDotSize, stepDotSize),
				FillColor: cfg.PrimaryColor,
				Corners:   uimath.CornersAll(dotR),
			}, 2, 1)
			s.drawCheckmark(buf, cx, cy, 5, uimath.ColorWhite)

		case StepProcess:
			// Filled primary circle with number
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(cx-dotR, cy-dotR, stepDotSize, stepDotSize),
				FillColor: cfg.PrimaryColor,
				Corners:   uimath.CornersAll(dotR),
			}, 2, 1)
			s.drawNumber(buf, cx, cy, i+1, uimath.ColorWhite)

		case StepError:
			// Filled red circle with X
			errColor := uimath.ColorHex("#e34d59")
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(cx-dotR, cy-dotR, stepDotSize, stepDotSize),
				FillColor: errColor,
				Corners:   uimath.CornersAll(dotR),
			}, 2, 1)
			s.drawCross(buf, cx, cy, 5, uimath.ColorWhite)

		default: // StepDefault
			// Outlined circle with number
			buf.DrawRect(render.RectCmd{
				Bounds:      uimath.NewRect(cx-dotR, cy-dotR, stepDotSize, stepDotSize),
				FillColor:   uimath.ColorWhite,
				BorderColor: cfg.BorderColor,
				BorderWidth: 1,
				Corners:     uimath.CornersAll(dotR),
			}, 2, 1)
			s.drawNumber(buf, cx, cy, i+1, cfg.DisabledColor)
		}

		// --- Title ---
		titleColor := cfg.TextColor
		titleFont := cfg.FontSizeSm
		switch status {
		case StepProcess:
			titleColor = cfg.PrimaryColor
		case StepDefault:
			titleColor = cfg.DisabledColor
		case StepError:
			titleColor = uimath.ColorHex("#e34d59")
		}

		titleY := cy + dotR + cfg.SpaceXS
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(titleFont)
			tw := cfg.TextRenderer.MeasureText(item.Title, titleFont)
			tx := cx - tw/2
			cfg.TextRenderer.DrawText(buf, item.Title, tx, titleY, titleFont, stepW, titleColor, 1)

			// --- Content (description) ---
			if item.Content != "" {
				descY := titleY + lh + 2
				descColor := cfg.DisabledColor
				cfg.TextRenderer.DrawText(buf, item.Content, bounds.X+float32(i)*stepW+4, descY, cfg.FontSizeSm, stepW-8, descColor, 1)
			}
		} else {
			tw := float32(len(item.Title)) * titleFont * 0.55
			th := titleFont * 1.2
			tx := cx - tw/2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(tx, titleY, tw, th),
				FillColor: titleColor,
				Corners:   uimath.CornersAll(2),
			}, 3, 1)
		}
	}
}

// drawCheckmark draws a checkmark inside a circle at (cx, cy).
func (s *Steps) drawCheckmark(buf *render.CommandBuffer, cx, cy, size float32, color uimath.Color) {
	cfg := s.config
	if cfg.TextRenderer != nil {
		tw := cfg.TextRenderer.MeasureText("✓", cfg.FontSizeSm)
		lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
		cfg.TextRenderer.DrawText(buf, "✓", cx-tw/2, cy-lh/2, cfg.FontSizeSm, stepDotSize, color, 1)
		return
	}
	// Fallback: draw checkmark with small rects
	t := float32(1.5)
	// Short arm (going down-right at 45 degrees)
	for i := 0; i < 3; i++ {
		fi := float32(i)
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(cx-size*0.4+fi*t, cy+fi*t-t, t, t),
			FillColor: color,
		}, 3, 1)
	}
	// Long arm (going up-right at 45 degrees)
	for i := 0; i < 5; i++ {
		fi := float32(i)
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(cx-size*0.4+3*t+fi*t, cy+2*t-fi*t-t, t, t),
			FillColor: color,
		}, 3, 1)
	}
}

// drawCross draws an X inside a circle at (cx, cy).
func (s *Steps) drawCross(buf *render.CommandBuffer, cx, cy, size float32, color uimath.Color) {
	cfg := s.config
	if cfg.TextRenderer != nil {
		tw := cfg.TextRenderer.MeasureText("×", cfg.FontSizeSm)
		lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
		cfg.TextRenderer.DrawText(buf, "×", cx-tw/2, cy-lh/2, cfg.FontSizeSm, stepDotSize, color, 1)
		return
	}
	t := float32(1.5)
	steps := 5
	for i := 0; i < steps; i++ {
		fi := float32(i)
		off := fi*t - float32(steps/2)*t
		// Diagonal backslash
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(cx+off, cy+off, t, t),
			FillColor: color,
		}, 3, 1)
		// Diagonal forward slash
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(cx+off, cy-off, t, t),
			FillColor: color,
		}, 3, 1)
	}
}

// drawNumber draws a step number centered at (cx, cy).
func (s *Steps) drawNumber(buf *render.CommandBuffer, cx, cy float32, num int, color uimath.Color) {
	cfg := s.config
	str := fmt.Sprintf("%d", num)
	if cfg.TextRenderer != nil {
		tw := cfg.TextRenderer.MeasureText(str, cfg.FontSizeSm)
		lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
		cfg.TextRenderer.DrawText(buf, str, cx-tw/2, cy-lh/2, cfg.FontSizeSm, stepDotSize, color, 1)
	} else {
		tw := float32(len(str)) * cfg.FontSizeSm * 0.55
		th := cfg.FontSizeSm * 1.2
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(cx-tw/2, cy-th/2, tw, th),
			FillColor: color,
			Corners:   uimath.CornersAll(2),
		}, 3, 1)
	}
}

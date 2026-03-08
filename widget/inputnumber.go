package widget

import (
	"fmt"
	"strconv"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// InputNumberAlign controls the text alignment inside the input.
type InputNumberAlign uint8

const (
	InputNumberAlignLeft InputNumberAlign = iota
	InputNumberAlignCenter
	InputNumberAlignRight
)

// InputNumberStatus indicates the validation status of the input.
type InputNumberStatus uint8

const (
	InputNumberStatusDefault InputNumberStatus = iota
	InputNumberStatusSuccess
	InputNumberStatusWarning
	InputNumberStatusError
)

// InputNumberTheme controls the layout of increment/decrement buttons.
type InputNumberTheme uint8

const (
	// InputNumberThemeRow places - and + buttons on left and right sides.
	InputNumberThemeRow InputNumberTheme = iota
	// InputNumberThemeColumn stacks + and - buttons vertically on the right side.
	InputNumberThemeColumn
	// InputNumberThemeNormal hides the increment/decrement buttons.
	InputNumberThemeNormal
)

// Deprecated aliases for backward compatibility.
const (
	ThemeNormal = InputNumberThemeNormal
	ThemeColumn = InputNumberThemeColumn
	ThemeRow    = InputNumberThemeRow
)

// InputNumber is a numeric input with increment/decrement buttons.
type InputNumber struct {
	Base
	value               float64
	min                 float64
	max                 float64
	step                float64
	decimalPlaces       int
	theme               InputNumberTheme
	size                Size
	align               InputNumberAlign
	status              InputNumberStatus
	text                string
	placeholder         string
	label               string
	suffix              string
	tips                string
	disabled            bool
	readonly            bool
	allowInputOverLimit bool
	onChange            func(float64)
	onBlur              func(float64)
	onFocus             func(float64)
	onEnter             func(float64)
	onValidate          func(err string)
}

func NewInputNumber(tree *core.Tree, cfg *Config) *InputNumber {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	n := &InputNumber{
		Base:          NewBase(tree, core.TypeInput, cfg),
		min:           0,
		max:           100,
		step:          1,
		decimalPlaces: -1, // -1 means auto (no trailing zeros)
		theme:         ThemeNormal,
		text:          "0",
	}
	tree.AddHandler(n.id, event.MouseClick, func(e *event.Event) {})
	return n
}

func (n *InputNumber) Value() float64              { return n.value }
func (n *InputNumber) Min() float64                { return n.min }
func (n *InputNumber) Max() float64                { return n.max }
func (n *InputNumber) Step() float64               { return n.step }
func (n *InputNumber) Theme() InputNumberTheme     { return n.theme }
func (n *InputNumber) DecimalPlaces() int          { return n.decimalPlaces }
func (n *InputNumber) Text() string                { return n.text }
func (n *InputNumber) SetMin(v float64)            { n.min = v }
func (n *InputNumber) SetMax(v float64)            { n.max = v }
func (n *InputNumber) SetStep(v float64)           { n.step = v }
func (n *InputNumber) SetDisabled(d bool)              { n.disabled = d }
func (n *InputNumber) SetTheme(t InputNumberTheme)     { n.theme = t }
func (n *InputNumber) SetDecimalPlaces(d int)          { n.decimalPlaces = d }
func (n *InputNumber) SetSize(s Size)                  { n.size = s }
func (n *InputNumber) Size() Size                      { return n.size }
func (n *InputNumber) SetAlign(a InputNumberAlign)     { n.align = a }
func (n *InputNumber) Align() InputNumberAlign         { return n.align }
func (n *InputNumber) SetStatus(s InputNumberStatus)   { n.status = s }
func (n *InputNumber) Status() InputNumberStatus       { return n.status }
func (n *InputNumber) SetPlaceholder(p string)         { n.placeholder = p }
func (n *InputNumber) Placeholder() string             { return n.placeholder }
func (n *InputNumber) SetLabel(l string)               { n.label = l }
func (n *InputNumber) Label() string                   { return n.label }
func (n *InputNumber) SetSuffix(s string)              { n.suffix = s }
func (n *InputNumber) Suffix() string                  { return n.suffix }
func (n *InputNumber) SetTips(t string)                { n.tips = t }
func (n *InputNumber) Tips() string                    { return n.tips }
func (n *InputNumber) SetReadonly(r bool)              { n.readonly = r }
func (n *InputNumber) IsReadonly() bool                { return n.readonly }
func (n *InputNumber) SetAllowInputOverLimit(v bool)   { n.allowInputOverLimit = v }
func (n *InputNumber) AllowInputOverLimit() bool       { return n.allowInputOverLimit }
func (n *InputNumber) OnChange(fn func(float64))       { n.onChange = fn }
func (n *InputNumber) OnBlur(fn func(float64))         { n.onBlur = fn }
func (n *InputNumber) OnFocus(fn func(float64))        { n.onFocus = fn }
func (n *InputNumber) OnEnter(fn func(float64))        { n.onEnter = fn }
func (n *InputNumber) OnValidate(fn func(err string))  { n.onValidate = fn }

func (n *InputNumber) SetValue(v float64) {
	if v < n.min {
		v = n.min
	}
	if v > n.max {
		v = n.max
	}
	if v != n.value {
		n.value = v
		n.text = n.formatValue()
		if n.onChange != nil {
			n.onChange(v)
		}
	}
}

// SetText sets the text input directly. If parseable as float64, the value is updated.
func (n *InputNumber) SetText(s string) {
	n.text = s
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		n.SetValue(v)
	}
}

func (n *InputNumber) Increment() { n.SetValue(n.value + n.step) }
func (n *InputNumber) Decrement() { n.SetValue(n.value - n.step) }

// formatValue formats the current value respecting decimalPlaces.
func (n *InputNumber) formatValue() string {
	if n.decimalPlaces >= 0 {
		return fmt.Sprintf("%.*f", n.decimalPlaces, n.value)
	}
	return strconv.FormatFloat(n.value, 'f', -1, 64)
}

func (n *InputNumber) Draw(buf *render.CommandBuffer) {
	bounds := n.Bounds()
	if bounds.IsEmpty() {
		return
	}

	switch n.theme {
	case InputNumberThemeColumn:
		n.drawColumn(buf, bounds)
	case InputNumberThemeNormal:
		n.drawInputOnly(buf, bounds)
	default: // InputNumberThemeRow
		n.drawNormal(buf, bounds)
	}
}

// drawNormal renders ThemeNormal: [-] [value] [+] side by side.
func (n *InputNumber) drawNormal(buf *render.CommandBuffer, bounds uimath.Rect) {
	cfg := n.config
	btnW := float32(28)
	inputW := bounds.Width - btnW*2

	// Main box
	borderClr := cfg.BorderColor
	if elem := n.Element(); elem != nil && elem.IsFocused() {
		borderClr = cfg.FocusBorderColor
	}
	buf.DrawRect(render.RectCmd{
		Bounds:      bounds,
		FillColor:   uimath.ColorWhite,
		BorderColor: borderClr,
		BorderWidth: cfg.BorderWidth,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 0, 1)

	// Decrement button
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(bounds.X, bounds.Y, btnW, bounds.Height),
		FillColor: uimath.RGBA(0, 0, 0, 0.02),
	}, 1, 1)
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(bounds.X+btnW-1, bounds.Y, 1, bounds.Height),
		FillColor: cfg.BorderColor,
	}, 1, 1)
	// Minus sign
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(bounds.X+btnW/2-5, bounds.Y+bounds.Height/2-0.5, 10, 1),
		FillColor: cfg.TextColor,
	}, 2, 1)

	// Increment button
	incX := bounds.X + btnW + inputW
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(incX, bounds.Y, btnW, bounds.Height),
		FillColor: uimath.RGBA(0, 0, 0, 0.02),
	}, 1, 1)
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(incX, bounds.Y, 1, bounds.Height),
		FillColor: cfg.BorderColor,
	}, 1, 1)
	// Plus sign
	cx := incX + btnW/2
	cy := bounds.Y + bounds.Height/2
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(cx-5, cy-0.5, 10, 1),
		FillColor: cfg.TextColor,
	}, 2, 1)
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(cx-0.5, cy-5, 1, 10),
		FillColor: cfg.TextColor,
	}, 2, 1)

	// Value text
	n.drawValueText(buf, bounds.X+btnW, bounds.Y, inputW, bounds.Height)
}

// drawColumn renders ThemeColumn: [value] [+/-] stacked on right.
func (n *InputNumber) drawColumn(buf *render.CommandBuffer, bounds uimath.Rect) {
	cfg := n.config
	btnW := float32(28)
	btnH := bounds.Height / 2
	inputW := bounds.Width - btnW

	// Main box
	borderClr := cfg.BorderColor
	if elem := n.Element(); elem != nil && elem.IsFocused() {
		borderClr = cfg.FocusBorderColor
	}
	buf.DrawRect(render.RectCmd{
		Bounds:      bounds,
		FillColor:   uimath.ColorWhite,
		BorderColor: borderClr,
		BorderWidth: cfg.BorderWidth,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 0, 1)

	// Separator line between input and buttons
	btnX := bounds.X + inputW
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(btnX, bounds.Y, 1, bounds.Height),
		FillColor: cfg.BorderColor,
	}, 1, 1)

	// Increment button (top)
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(btnX, bounds.Y, btnW, btnH),
		FillColor: uimath.RGBA(0, 0, 0, 0.02),
	}, 1, 1)
	// Plus sign (small)
	cx := btnX + btnW/2
	cy := bounds.Y + btnH/2
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(cx-4, cy-0.5, 8, 1),
		FillColor: cfg.TextColor,
	}, 2, 1)
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(cx-0.5, cy-4, 1, 8),
		FillColor: cfg.TextColor,
	}, 2, 1)

	// Horizontal separator between buttons
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(btnX, bounds.Y+btnH, btnW, 1),
		FillColor: cfg.BorderColor,
	}, 1, 1)

	// Decrement button (bottom)
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(btnX, bounds.Y+btnH, btnW, btnH),
		FillColor: uimath.RGBA(0, 0, 0, 0.02),
	}, 1, 1)
	// Minus sign (small)
	cy2 := bounds.Y + btnH + btnH/2
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(cx-4, cy2-0.5, 8, 1),
		FillColor: cfg.TextColor,
	}, 2, 1)

	// Value text
	n.drawValueText(buf, bounds.X, bounds.Y, inputW, bounds.Height)
}

// drawInputOnly renders ThemeNormal: just the input box, no increment/decrement buttons.
func (n *InputNumber) drawInputOnly(buf *render.CommandBuffer, bounds uimath.Rect) {
	cfg := n.config

	borderClr := cfg.BorderColor
	if elem := n.Element(); elem != nil && elem.IsFocused() {
		borderClr = cfg.FocusBorderColor
	}
	buf.DrawRect(render.RectCmd{
		Bounds:      bounds,
		FillColor:   uimath.ColorWhite,
		BorderColor: borderClr,
		BorderWidth: cfg.BorderWidth,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 0, 1)

	n.drawValueText(buf, bounds.X, bounds.Y, bounds.Width, bounds.Height)
}

// drawValueText draws the text value centered in the given region.
func (n *InputNumber) drawValueText(buf *render.CommandBuffer, x, y, w, h float32) {
	cfg := n.config
	text := n.text
	if text == "" {
		text = n.formatValue()
	}

	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		tw := cfg.TextRenderer.MeasureText(text, cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, text, x+(w-tw)/2, y+(h-lh)/2, cfg.FontSize, w, cfg.TextColor, 1)
	} else {
		tw := float32(len(text)) * cfg.FontSize * 0.55
		th := cfg.FontSize * 1.2
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x+(w-tw)/2, y+(h-th)/2, tw, th),
			FillColor: cfg.TextColor,
			Corners:   uimath.CornersAll(2),
		}, 1, 1)
	}
}

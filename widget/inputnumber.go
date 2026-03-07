package widget

import (
	"strconv"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// InputNumber is a numeric input with increment/decrement buttons.
type InputNumber struct {
	Base
	value    float64
	min      float64
	max      float64
	step     float64
	disabled bool
	onChange func(float64)
}

func NewInputNumber(tree *core.Tree, cfg *Config) *InputNumber {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	n := &InputNumber{
		Base: NewBase(tree, core.TypeInput, cfg),
		min:  0,
		max:  100,
		step: 1,
	}
	tree.AddHandler(n.id, event.MouseClick, func(e *event.Event) {})
	return n
}

func (n *InputNumber) Value() float64        { return n.value }
func (n *InputNumber) Min() float64          { return n.min }
func (n *InputNumber) Max() float64          { return n.max }
func (n *InputNumber) Step() float64         { return n.step }
func (n *InputNumber) SetMin(v float64)      { n.min = v }
func (n *InputNumber) SetMax(v float64)      { n.max = v }
func (n *InputNumber) SetStep(v float64)     { n.step = v }
func (n *InputNumber) SetDisabled(d bool)    { n.disabled = d }
func (n *InputNumber) OnChange(fn func(float64)) { n.onChange = fn }

func (n *InputNumber) SetValue(v float64) {
	if v < n.min {
		v = n.min
	}
	if v > n.max {
		v = n.max
	}
	if v != n.value {
		n.value = v
		if n.onChange != nil {
			n.onChange(v)
		}
	}
}

func (n *InputNumber) Increment() { n.SetValue(n.value + n.step) }
func (n *InputNumber) Decrement() { n.SetValue(n.value - n.step) }

func (n *InputNumber) Draw(buf *render.CommandBuffer) {
	bounds := n.Bounds()
	if bounds.IsEmpty() {
		return
	}
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
	text := strconv.FormatFloat(n.value, 'f', -1, 64)
	textX := bounds.X + btnW + cfg.SpaceSM
	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		tw := cfg.TextRenderer.MeasureText(text, cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, text, textX+(inputW-cfg.SpaceSM*2-tw)/2, bounds.Y+(bounds.Height-lh)/2, cfg.FontSize, inputW-cfg.SpaceSM*2, cfg.TextColor, 1)
	} else {
		tw := float32(len(text)) * cfg.FontSize * 0.55
		th := cfg.FontSize * 1.2
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(textX+(inputW-tw)/2, bounds.Y+(bounds.Height-th)/2, tw, th),
			FillColor: cfg.TextColor,
			Corners:   uimath.CornersAll(2),
		}, 1, 1)
	}
}

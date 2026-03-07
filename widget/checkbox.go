package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Checkbox is a toggle control with a checked/unchecked state.
type Checkbox struct {
	Base
	label    string
	checked  bool
	disabled bool

	onChange func(checked bool)
}

// NewCheckbox creates a checkbox with the given label.
func NewCheckbox(tree *core.Tree, label string, cfg *Config) *Checkbox {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	c := &Checkbox{
		Base:  NewBase(tree, core.TypeCustom, cfg),
		label: label,
	}
	c.style.Display = layout.DisplayFlex
	c.style.AlignItems = layout.AlignCenter
	c.style.Height = layout.Px(cfg.ButtonHeight)
	c.style.Gap = cfg.SpaceSM
	tree.SetProperty(c.id, "text", label)

	c.tree.AddHandler(c.id, event.MouseClick, func(e *event.Event) {
		if !c.disabled {
			c.checked = !c.checked
			if c.onChange != nil {
				c.onChange(c.checked)
			}
		}
	})

	return c
}

func (c *Checkbox) Label() string     { return c.label }
func (c *Checkbox) IsChecked() bool   { return c.checked }
func (c *Checkbox) IsDisabled() bool  { return c.disabled }

func (c *Checkbox) SetLabel(label string) {
	c.label = label
	c.tree.SetProperty(c.id, "text", label)
}

func (c *Checkbox) SetChecked(checked bool) { c.checked = checked }
func (c *Checkbox) SetDisabled(d bool) {
	c.disabled = d
	c.tree.SetEnabled(c.id, !d)
}

func (c *Checkbox) OnChange(fn func(checked bool)) {
	c.onChange = fn
}

const checkboxBoxSize = float32(16)

func (c *Checkbox) Draw(buf *render.CommandBuffer) {
	bounds := c.Bounds()
	if bounds.IsEmpty() {
		return
	}

	cfg := c.config
	elem := c.Element()
	hovered := elem != nil && elem.IsHovered()

	// Box position: vertically centered
	boxY := bounds.Y + (bounds.Height-checkboxBoxSize)/2
	boxRect := uimath.NewRect(bounds.X, boxY, checkboxBoxSize, checkboxBoxSize)

	// Box colors
	var fillColor, borderColor uimath.Color
	if c.disabled {
		fillColor = uimath.ColorHex("#f5f5f5")
		borderColor = cfg.DisabledColor
	} else if c.checked {
		fillColor = cfg.PrimaryColor
		if hovered {
			fillColor = cfg.HoverColor
		}
		borderColor = fillColor
	} else {
		fillColor = uimath.ColorWhite
		borderColor = cfg.BorderColor
		if hovered {
			borderColor = cfg.HoverColor
		}
	}

	// Draw box
	buf.DrawRect(render.RectCmd{
		Bounds:      boxRect,
		FillColor:   fillColor,
		BorderColor: borderColor,
		BorderWidth: cfg.BorderWidth,
		Corners:     uimath.CornersAll(cfg.BorderRadius / 2),
	}, 0, 1)

	// Draw checkmark when checked
	if c.checked {
		checkColor := uimath.ColorWhite
		if c.disabled {
			checkColor = cfg.DisabledColor
		}
		// Draw a ✓ shape using small rectangles along two line segments:
		// Stroke 1: top-left down to bottom-center (short leg)
		// Stroke 2: bottom-center up to top-right (long leg)
		ox := boxRect.X + 3
		oy := boxRect.Y + 3
		const pw = float32(2) // pen width
		// Points relative to box interior (10x10 area):
		// P0=(1,5) → P1=(3.5,8) → P2=(9,2)
		// Short leg: (1,5) → (3.5,8) — 3 steps
		steps := [][2]float32{{1, 5}, {2, 6.5}, {3, 7.5}, {3.5, 8}}
		for _, p := range steps {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(ox+p[0], oy+p[1], pw, pw),
				FillColor: checkColor,
			}, 1, 1)
		}
		// Long leg: (3.5,8) → (9,2) — 6 steps
		steps = [][2]float32{{4.5, 7}, {5.5, 6}, {6.5, 5}, {7.5, 4}, {8.5, 3}, {9, 2}}
		for _, p := range steps {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(ox+p[0], oy+p[1], pw, pw),
				FillColor: checkColor,
			}, 1, 1)
		}
	}

	// Draw label
	if c.label != "" {
		labelX := bounds.X + checkboxBoxSize + cfg.SpaceSM
		labelColor := cfg.TextColor
		if c.disabled {
			labelColor = cfg.DisabledColor
		}
		labelW := bounds.Width - checkboxBoxSize - cfg.SpaceSM
		if labelW > 0 {
			if cfg.TextRenderer != nil {
				lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
				labelY := bounds.Y + (bounds.Height-lh)/2
				cfg.TextRenderer.DrawText(buf, c.label, labelX, labelY, cfg.FontSize, labelW, labelColor, 1)
			} else {
				textW := float32(len(c.label)) * cfg.FontSize * 0.55
				if textW > labelW {
					textW = labelW
				}
				textH := cfg.FontSize * 1.2
				labelY := bounds.Y + (bounds.Height-textH)/2
				buf.DrawRect(render.RectCmd{
					Bounds:    uimath.NewRect(labelX, labelY, textW, textH),
					FillColor: labelColor,
					Corners:   uimath.CornersAll(2),
				}, 1, 1)
			}
		}
	}

	c.DrawChildren(buf)
}

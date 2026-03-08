package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Radio is a single radio button. Use RadioGroup to manage mutual exclusion.
type Radio struct {
	Base
	label        string
	value        string
	checked      bool
	disabled     bool
	readonly     bool
	allowUncheck bool
	group        *RadioGroup

	onChange func(checked bool)
}

// NewRadio creates a radio button with the given label.
func NewRadio(tree *core.Tree, label string, cfg *Config) *Radio {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	r := &Radio{
		Base:  NewBase(tree, core.TypeCustom, cfg),
		label: label,
	}
	r.style.Display = layout.DisplayFlex
	r.style.AlignItems = layout.AlignCenter
	r.style.Height = layout.Px(cfg.ButtonHeight)
	r.style.Gap = cfg.SpaceSM
	tree.SetProperty(r.id, "text", label)

	r.tree.AddHandler(r.id, event.MouseClick, func(e *event.Event) {
		if r.disabled || r.readonly {
			return
		}
		if r.group != nil {
			r.group.Select(r)
		} else {
			if r.allowUncheck && r.checked {
				r.checked = false
				if r.onChange != nil {
					r.onChange(false)
				}
			} else {
				r.checked = true
				if r.onChange != nil {
					r.onChange(true)
				}
			}
		}
	})

	return r
}

func (r *Radio) Label() string        { return r.label }
func (r *Radio) Value() string        { return r.value }
func (r *Radio) IsChecked() bool      { return r.checked }
func (r *Radio) IsDisabled() bool     { return r.disabled }
func (r *Radio) IsReadonly() bool     { return r.readonly }
func (r *Radio) AllowUncheck() bool   { return r.allowUncheck }

func (r *Radio) SetValue(v string)        { r.value = v }
func (r *Radio) SetReadonly(v bool)       { r.readonly = v }
func (r *Radio) SetAllowUncheck(v bool)   { r.allowUncheck = v }

func (r *Radio) SetLabel(label string) {
	r.label = label
	r.tree.SetProperty(r.id, "text", label)
}

func (r *Radio) SetChecked(checked bool) { r.checked = checked }
func (r *Radio) SetDisabled(d bool) {
	r.disabled = d
	r.tree.SetEnabled(r.id, !d)
}

func (r *Radio) OnChange(fn func(checked bool)) {
	r.onChange = fn
}

const radioSize = float32(16)

func (r *Radio) Draw(buf *render.CommandBuffer) {
	bounds := r.Bounds()
	if bounds.IsEmpty() {
		return
	}

	cfg := r.config
	elem := r.Element()
	hovered := elem != nil && elem.IsHovered()

	// Circle position: vertically centered
	circleY := bounds.Y + (bounds.Height-radioSize)/2
	circleRect := uimath.NewRect(bounds.X, circleY, radioSize, radioSize)
	cornerRadius := radioSize / 2

	// Colors
	var fillColor, borderColor uimath.Color
	if r.disabled {
		fillColor = uimath.ColorHex("#f5f5f5")
		borderColor = cfg.DisabledColor
	} else if r.checked {
		fillColor = uimath.ColorWhite
		borderColor = cfg.PrimaryColor
		if hovered {
			borderColor = cfg.HoverColor
		}
	} else {
		fillColor = uimath.ColorWhite
		borderColor = cfg.BorderColor
		if hovered {
			borderColor = cfg.HoverColor
		}
	}

	// Draw outer circle
	buf.DrawRect(render.RectCmd{
		Bounds:      circleRect,
		FillColor:   fillColor,
		BorderColor: borderColor,
		BorderWidth: cfg.BorderWidth,
		Corners:     uimath.CornersAll(cornerRadius),
	}, 0, 1)

	// Draw inner dot when checked
	if r.checked {
		dotSize := radioSize * 0.5
		dotColor := cfg.PrimaryColor
		if r.disabled {
			dotColor = cfg.DisabledColor
		} else if hovered {
			dotColor = cfg.HoverColor
		}
		dotX := circleRect.X + (radioSize-dotSize)/2
		dotY := circleRect.Y + (radioSize-dotSize)/2
		buf.DrawRect(render.RectCmd{
			Bounds:  uimath.NewRect(dotX, dotY, dotSize, dotSize),
			FillColor: dotColor,
			Corners: uimath.CornersAll(dotSize / 2),
		}, 1, 1)
	}

	// Draw label
	if r.label != "" {
		labelX := bounds.X + radioSize + cfg.SpaceSM
		labelColor := cfg.TextColor
		if r.disabled {
			labelColor = cfg.DisabledColor
		}
		labelW := bounds.Width - radioSize - cfg.SpaceSM
		if labelW > 0 {
			if cfg.TextRenderer != nil {
				lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
				labelY := bounds.Y + (bounds.Height-lh)/2
				cfg.TextRenderer.DrawText(buf, r.label, labelX, labelY, cfg.FontSize, labelW, labelColor, 1)
			} else {
				textW := float32(len(r.label)) * cfg.FontSize * 0.55
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

	r.DrawChildren(buf)
}

// RadioGroup manages mutual exclusion among a set of Radio buttons.
type RadioGroup struct {
	radios       []*Radio
	value        string // label of the currently selected radio
	size         Size
	disabled     bool
	readonly     bool
	allowUncheck bool
	onChange     func(value string)
}

// SetSize sets the size for all radios in the group.
func (g *RadioGroup) SetSize(s Size) { g.size = s }

// Size returns the current size setting.
func (g *RadioGroup) Size() Size { return g.size }

func (g *RadioGroup) SetDisabled(d bool)      { g.disabled = d }
func (g *RadioGroup) IsDisabled() bool         { return g.disabled }
func (g *RadioGroup) SetReadonly(r bool)       { g.readonly = r }
func (g *RadioGroup) IsReadonly() bool         { return g.readonly }
func (g *RadioGroup) SetAllowUncheck(v bool)   { g.allowUncheck = v }
func (g *RadioGroup) AllowUncheck() bool       { return g.allowUncheck }

// NewRadioGroup creates a radio group.
func NewRadioGroup() *RadioGroup {
	return &RadioGroup{}
}

// Add adds a radio to the group.
func (g *RadioGroup) Add(r *Radio) {
	r.group = g
	g.radios = append(g.radios, r)
}

// Value returns the label of the currently selected radio.
func (g *RadioGroup) Value() string { return g.value }

// SetValue selects the radio with the given label.
func (g *RadioGroup) SetValue(label string) {
	g.value = label
	for _, r := range g.radios {
		r.checked = r.label == label
	}
}

// OnChange sets a callback for when the selection changes.
func (g *RadioGroup) OnChange(fn func(value string)) {
	g.onChange = fn
}

// Select is called internally when a radio is clicked.
func (g *RadioGroup) Select(r *Radio) {
	if g.disabled || g.readonly {
		return
	}
	// Allow unchecking if allowUncheck is set on group or radio
	if (g.allowUncheck || r.allowUncheck) && r.checked {
		r.checked = false
		g.value = ""
		if r.onChange != nil {
			r.onChange(false)
		}
		if g.onChange != nil {
			g.onChange("")
		}
		return
	}
	for _, other := range g.radios {
		other.checked = other == r
	}
	g.value = r.label
	if r.onChange != nil {
		r.onChange(true)
	}
	if g.onChange != nil {
		g.onChange(r.label)
	}
}

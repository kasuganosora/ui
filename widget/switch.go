package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Switch is a toggle switch control.
type Switch struct {
	Base
	value    bool
	disabled bool
	size     Size
	loading  bool
	label    [2]string // [0]=off text, [1]=on text

	onChange func(value bool)
}

// switchDimensions returns width, height, knob size for the given Size.
func switchDimensions(s Size) (w, h, knob float32) {
	switch s {
	case SizeSmall:
		return 36, 18, 14
	case SizeLarge:
		return 52, 26, 22
	default: // SizeMedium
		return 44, 22, 18
	}
}

// NewSwitch creates a switch widget.
func NewSwitch(tree *core.Tree, cfg *Config) *Switch {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	s := &Switch{
		Base: NewBase(tree, core.TypeCustom, cfg),
	}
	w, h, _ := switchDimensions(s.size)
	s.style.Display = layout.DisplayFlex
	s.style.AlignItems = layout.AlignCenter
	s.style.Width = layout.Px(w)
	s.style.Height = layout.Px(h)
	s.style.FlexShrink = 0

	s.tree.AddHandler(s.id, event.MouseClick, func(e *event.Event) {
		if !s.disabled && !s.loading {
			s.value = !s.value
			if s.onChange != nil {
				s.onChange(s.value)
			}
		}
	})

	return s
}

func (s *Switch) Value() bool      { return s.value }
func (s *Switch) IsDisabled() bool { return s.disabled }
func (s *Switch) IsLoading() bool  { return s.loading }

func (s *Switch) SetValue(v bool) { s.value = v }
func (s *Switch) SetDisabled(d bool) {
	s.disabled = d
	s.tree.SetEnabled(s.id, !d)
}

// SetSize updates the switch size and recalculates layout dimensions.
func (s *Switch) SetSize(sz Size) {
	s.size = sz
	w, h, _ := switchDimensions(sz)
	s.style.Width = layout.Px(w)
	s.style.Height = layout.Px(h)
}

// SetLoading sets the loading state. While loading, clicks are ignored.
func (s *Switch) SetLoading(l bool) { s.loading = l }

// SetLabel sets the on/off label text rendered inside the track.
// label[0] = off text, label[1] = on text.
func (s *Switch) SetLabel(label [2]string) { s.label = label }

func (s *Switch) OnChange(fn func(value bool)) {
	s.onChange = fn
}

func (s *Switch) Draw(buf *render.CommandBuffer) {
	bounds := s.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := s.config
	elem := s.Element()
	hovered := elem != nil && elem.IsHovered()

	sw, sh, sk := switchDimensions(s.size)

	// Track
	trackY := bounds.Y + (bounds.Height-sh)/2
	trackRect := uimath.NewRect(bounds.X, trackY, sw, sh)

	var trackColor uimath.Color
	if s.disabled {
		if s.value {
			trackColor = cfg.DisabledColor
		} else {
			trackColor = uimath.ColorHex("#c6c6c6")
		}
	} else if s.value {
		trackColor = cfg.PrimaryColor
		if hovered {
			trackColor = cfg.HoverColor
		}
	} else {
		trackColor = uimath.ColorHex("#c6c6c6")
		if hovered {
			trackColor = uimath.ColorHex("#8b8b8b")
		}
	}

	buf.DrawRect(render.RectCmd{
		Bounds:    trackRect,
		FillColor: trackColor,
		Corners:   uimath.CornersAll(sh / 2),
	}, 0, 1)

	// Label text inside track
	labelText := s.label[0] // off text
	if s.value {
		labelText = s.label[1] // on text
	}
	if labelText != "" && cfg.TextRenderer != nil {
		labelFontSize := cfg.SizeFontSize(s.size) - 2
		lh := cfg.TextRenderer.LineHeight(labelFontSize)
		labelY := trackRect.Y + (sh-lh)/2
		labelColor := uimath.ColorWhite
		if s.value {
			// Label on the left side (before knob)
			labelX := trackRect.X + 2 + (sh-sk)/2
			maxW := sw - sk - (sh-sk) - 2
			cfg.TextRenderer.DrawText(buf, labelText, labelX, labelY, labelFontSize, maxW, labelColor, 1)
		} else {
			// Label on the right side (after knob)
			labelX := trackRect.X + sk + (sh-sk)/2 + 2
			maxW := sw - sk - (sh-sk) - 2
			cfg.TextRenderer.DrawText(buf, labelText, labelX, labelY, labelFontSize, maxW, labelColor, 1)
		}
	}

	// Knob
	knobPad := (sh - sk) / 2
	var knobX float32
	if s.value {
		knobX = trackRect.X + sw - sk - knobPad
	} else {
		knobX = trackRect.X + knobPad
	}
	knobY := trackRect.Y + knobPad

	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(knobX, knobY, sk, sk),
		FillColor: uimath.ColorWhite,
		Corners:   uimath.CornersAll(sk / 2),
	}, 1, 1)

	// Loading indicator: three small dots rotating on the knob
	if s.loading {
		dotSize := sk * 0.15
		dotColor := cfg.PrimaryColor
		if !s.value {
			dotColor = uimath.ColorHex("#8b8b8b")
		}
		knobCX := knobX + sk/2
		knobCY := knobY + sk/2
		// Draw three dots in a horizontal line
		for i := -1; i <= 1; i++ {
			dx := knobCX + float32(i)*dotSize*1.8 - dotSize/2
			dy := knobCY - dotSize/2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(dx, dy, dotSize, dotSize),
				FillColor: dotColor,
				Corners:   uimath.CornersAll(dotSize / 2),
			}, 2, 1)
		}
	}

	s.DrawChildren(buf)
}

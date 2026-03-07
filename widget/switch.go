package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

const (
	switchWidth  = float32(44)
	switchHeight = float32(22)
	switchKnob   = float32(18)
)

// Switch is a toggle switch control.
type Switch struct {
	Base
	checked  bool
	disabled bool

	onChange func(checked bool)
}

// NewSwitch creates a switch widget.
func NewSwitch(tree *core.Tree, cfg *Config) *Switch {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	s := &Switch{
		Base: NewBase(tree, core.TypeCustom, cfg),
	}
	s.style.Display = layout.DisplayFlex
	s.style.AlignItems = layout.AlignCenter
	s.style.Width = layout.Px(switchWidth)
	s.style.Height = layout.Px(switchHeight)

	s.tree.AddHandler(s.id, event.MouseClick, func(e *event.Event) {
		if !s.disabled {
			s.checked = !s.checked
			if s.onChange != nil {
				s.onChange(s.checked)
			}
		}
	})

	return s
}

func (s *Switch) IsChecked() bool  { return s.checked }
func (s *Switch) IsDisabled() bool { return s.disabled }

func (s *Switch) SetChecked(checked bool) { s.checked = checked }
func (s *Switch) SetDisabled(d bool) {
	s.disabled = d
	s.tree.SetEnabled(s.id, !d)
}

func (s *Switch) OnChange(fn func(checked bool)) {
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

	// Track
	trackY := bounds.Y + (bounds.Height-switchHeight)/2
	trackRect := uimath.NewRect(bounds.X, trackY, switchWidth, switchHeight)

	var trackColor uimath.Color
	if s.disabled {
		if s.checked {
			trackColor = cfg.DisabledColor
		} else {
			trackColor = uimath.ColorHex("#bfbfbf")
		}
	} else if s.checked {
		trackColor = cfg.PrimaryColor
		if hovered {
			trackColor = cfg.HoverColor
		}
	} else {
		trackColor = uimath.ColorHex("#bfbfbf")
		if hovered {
			trackColor = uimath.ColorHex("#8c8c8c")
		}
	}

	buf.DrawRect(render.RectCmd{
		Bounds:  trackRect,
		FillColor: trackColor,
		Corners: uimath.CornersAll(switchHeight / 2),
	}, 0, 1)

	// Knob
	knobPad := (switchHeight - switchKnob) / 2
	var knobX float32
	if s.checked {
		knobX = trackRect.X + switchWidth - switchKnob - knobPad
	} else {
		knobX = trackRect.X + knobPad
	}
	knobY := trackRect.Y + knobPad

	buf.DrawRect(render.RectCmd{
		Bounds:  uimath.NewRect(knobX, knobY, switchKnob, switchKnob),
		FillColor: uimath.ColorWhite,
		Corners: uimath.CornersAll(switchKnob / 2),
	}, 1, 1)

	s.DrawChildren(buf)
}

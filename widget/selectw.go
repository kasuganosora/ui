package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// SelectOption represents a single option in a Select dropdown.
type SelectOption struct {
	Label    string
	Value    string
	Disabled bool
}

// Select is a dropdown selector widget.
type Select struct {
	Base
	options     []SelectOption
	value       string // currently selected value
	placeholder string
	disabled    bool
	open        bool

	optionIDs []core.ElementID
	onChange  func(value string)
}

// NewSelect creates a select dropdown.
func NewSelect(tree *core.Tree, options []SelectOption, cfg *Config) *Select {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	s := &Select{
		Base:        NewBase(tree, core.TypeCustom, cfg),
		options:     options,
		placeholder: "请选择",
	}
	s.style.Display = layout.DisplayFlex
	s.style.AlignItems = layout.AlignCenter
	s.style.Height = layout.Px(cfg.InputHeight)
	s.style.Padding = layout.EdgeValues{
		Left:  layout.Px(cfg.SpaceSM),
		Right: layout.Px(cfg.SpaceSM),
	}

	// Toggle dropdown on click
	s.tree.AddHandler(s.id, event.MouseClick, func(e *event.Event) {
		if s.disabled {
			return
		}
		s.open = !s.open
		if s.open {
			s.createOptionElements()
		} else {
			s.destroyOptionElements()
		}
	})

	return s
}

func (s *Select) Value() string       { return s.value }
func (s *Select) Placeholder() string { return s.placeholder }
func (s *Select) IsDisabled() bool    { return s.disabled }
func (s *Select) IsOpen() bool        { return s.open }
func (s *Select) Options() []SelectOption { return s.options }

func (s *Select) SetValue(v string) {
	s.value = v
	s.open = false
	s.destroyOptionElements()
}

func (s *Select) SetPlaceholder(p string) { s.placeholder = p }

func (s *Select) SetDisabled(d bool) {
	s.disabled = d
	s.tree.SetEnabled(s.id, !d)
}

func (s *Select) SetOptions(opts []SelectOption) {
	s.options = opts
	s.destroyOptionElements()
}

func (s *Select) OnChange(fn func(string)) { s.onChange = fn }

func (s *Select) selectedLabel() string {
	for _, opt := range s.options {
		if opt.Value == s.value {
			return opt.Label
		}
	}
	return ""
}

func (s *Select) createOptionElements() {
	s.destroyOptionElements()
	for i, opt := range s.options {
		eid := s.tree.CreateElement(core.TypeCustom)
		s.tree.SetProperty(eid, "text", opt.Label)
		s.optionIDs = append(s.optionIDs, eid)

		if !opt.Disabled {
			idx := i
			s.tree.AddHandler(eid, event.MouseClick, func(e *event.Event) {
				s.value = s.options[idx].Value
				s.open = false
				s.destroyOptionElements()
				if s.onChange != nil {
					s.onChange(s.value)
				}
			})
		}
	}
}

func (s *Select) destroyOptionElements() {
	for _, eid := range s.optionIDs {
		s.tree.DestroyElement(eid)
	}
	s.optionIDs = nil
}

const selectArrowSize = float32(8)

func (s *Select) Draw(buf *render.CommandBuffer) {
	bounds := s.Bounds()
	if bounds.IsEmpty() {
		return
	}

	cfg := s.config
	elem := s.Element()
	hovered := elem != nil && elem.IsHovered()

	// Border
	borderClr := cfg.BorderColor
	if s.open {
		borderClr = cfg.FocusBorderColor
	} else if hovered {
		borderClr = cfg.HoverColor
	}
	if s.disabled {
		borderClr = cfg.DisabledColor
	}

	bgClr := cfg.BgColor
	if s.disabled {
		bgClr = uimath.ColorHex("#f5f5f5")
	}

	buf.DrawRect(render.RectCmd{
		Bounds:      bounds,
		FillColor:   bgClr,
		BorderColor: borderClr,
		BorderWidth: cfg.BorderWidth,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 0, 1)

	// Display text
	label := s.selectedLabel()
	textColor := cfg.TextColor
	if label == "" {
		label = s.placeholder
		textColor = cfg.DisabledColor
	}
	if s.disabled {
		textColor = cfg.DisabledColor
	}

	padLeft := cfg.SpaceSM
	arrowArea := selectArrowSize + cfg.SpaceSM
	textMaxW := bounds.Width - padLeft - arrowArea

	if label != "" && textMaxW > 0 {
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			tx := bounds.X + padLeft
			ty := bounds.Y + (bounds.Height-lh)/2
			cfg.TextRenderer.DrawText(buf, label, tx, ty, cfg.FontSize, textMaxW, textColor, 1)
		} else {
			textW := float32(len(label)) * cfg.FontSize * 0.55
			if textW > textMaxW {
				textW = textMaxW
			}
			textH := cfg.FontSize * 1.2
			ty := bounds.Y + (bounds.Height-textH)/2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X+padLeft, ty, textW, textH),
				FillColor: textColor,
				Corners:   uimath.CornersAll(2),
			}, 1, 1)
		}
	}

	// Arrow indicator (simple triangle using a small rect)
	arrowX := bounds.X + bounds.Width - cfg.SpaceSM - selectArrowSize
	arrowY := bounds.Y + (bounds.Height-selectArrowSize)/2
	arrowColor := cfg.TextColor
	if s.disabled {
		arrowColor = cfg.DisabledColor
	}
	// Down arrow: horizontal bar
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(arrowX, arrowY+selectArrowSize/2, selectArrowSize, 2),
		FillColor: arrowColor,
	}, 1, 1)

	// Dropdown panel
	if s.open {
		s.drawDropdown(buf, bounds)
	}

	s.DrawChildren(buf)
}

func (s *Select) drawDropdown(buf *render.CommandBuffer, triggerBounds uimath.Rect) {
	cfg := s.config
	optH := cfg.InputHeight
	dropH := optH * float32(len(s.options))
	if dropH > 200 {
		dropH = 200
	}
	dropY := triggerBounds.Y + triggerBounds.Height + 4
	dropRect := uimath.NewRect(triggerBounds.X, dropY, triggerBounds.Width, dropH)

	// Dropdown background
	buf.DrawRect(render.RectCmd{
		Bounds:      dropRect,
		FillColor:   uimath.ColorWhite,
		BorderColor: cfg.BorderColor,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 10, 1)

	// Options
	y := dropY
	for i, opt := range s.options {
		optRect := uimath.NewRect(triggerBounds.X, y, triggerBounds.Width, optH)

		// Hover
		if i < len(s.optionIDs) {
			optElem := s.tree.Get(s.optionIDs[i])
			if optElem != nil && optElem.IsHovered() {
				buf.DrawRect(render.RectCmd{
					Bounds:    optRect,
					FillColor: uimath.ColorHex("#f5f5f5"),
				}, 11, 1)
			}
		}

		// Selected indicator
		if opt.Value == s.value {
			buf.DrawRect(render.RectCmd{
				Bounds:    optRect,
				FillColor: uimath.ColorHex("#e6f4ff"),
			}, 11, 1)
		}

		// Option label
		textColor := cfg.TextColor
		if opt.Disabled {
			textColor = cfg.DisabledColor
		}
		if opt.Value == s.value {
			textColor = cfg.PrimaryColor
		}

		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			tx := optRect.X + cfg.SpaceSM
			ty := optRect.Y + (optH-lh)/2
			maxW := optRect.Width - cfg.SpaceSM*2
			cfg.TextRenderer.DrawText(buf, opt.Label, tx, ty, cfg.FontSize, maxW, textColor, 1)
		} else {
			textW := float32(len(opt.Label)) * cfg.FontSize * 0.55
			maxW := optRect.Width - cfg.SpaceSM*2
			if textW > maxW {
				textW = maxW
			}
			textH := cfg.FontSize * 1.2
			tx := optRect.X + cfg.SpaceSM
			ty := optRect.Y + (optH-textH)/2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(tx, ty, textW, textH),
				FillColor: textColor,
				Corners:   uimath.CornersAll(2),
			}, 12, 1)
		}

		y += optH
	}
}

// Destroy cleans up option elements.
func (s *Select) Destroy() {
	s.destroyOptionElements()
	s.Base.Destroy()
}

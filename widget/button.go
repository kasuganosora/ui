package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// ButtonVariant controls the button appearance.
type ButtonVariant uint8

const (
	ButtonPrimary   ButtonVariant = iota
	ButtonSecondary
	ButtonText
	ButtonLink
)

// Button is an interactive button widget.
type Button struct {
	Base
	label    string
	variant  ButtonVariant
	disabled bool
	pressed  bool

	onClick func()
}

// NewButton creates a button with the given label.
func NewButton(tree *core.Tree, label string, cfg *Config) *Button {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	b := &Button{
		Base:    NewBase(tree, core.TypeButton, cfg),
		label:   label,
		variant: ButtonPrimary,
	}
	b.style.Display = layout.DisplayFlex
	b.style.AlignItems = layout.AlignCenter
	b.style.JustifyContent = layout.JustifyCenter
	b.style.Height = layout.Px(cfg.ButtonHeight)
	b.style.Padding = layout.EdgeValues{
		Left:  layout.Px(cfg.SpaceMD),
		Right: layout.Px(cfg.SpaceMD),
	}
	tree.SetProperty(b.id, "text", label)

	b.tree.AddHandler(b.id, event.MouseDown, func(e *event.Event) {
		if !b.disabled {
			b.pressed = true
		}
	})
	b.tree.AddHandler(b.id, event.MouseUp, func(e *event.Event) {
		b.pressed = false
	})
	b.tree.AddHandler(b.id, event.MouseClick, func(e *event.Event) {
		if !b.disabled && b.onClick != nil {
			b.onClick()
		}
	})

	return b
}

func (b *Button) Label() string          { return b.label }
func (b *Button) Variant() ButtonVariant  { return b.variant }
func (b *Button) IsDisabled() bool        { return b.disabled }
func (b *Button) IsPressed() bool         { return b.pressed }

func (b *Button) SetLabel(label string) {
	b.label = label
	b.tree.SetProperty(b.id, "text", label)
}

func (b *Button) SetVariant(v ButtonVariant) { b.variant = v }

func (b *Button) SetDisabled(d bool) {
	b.disabled = d
	b.tree.SetEnabled(b.id, !d)
}

func (b *Button) OnClick(fn func()) {
	b.onClick = fn
}

func (b *Button) bgColor() uimath.Color {
	cfg := b.config
	if b.disabled {
		return cfg.DisabledColor
	}
	elem := b.Element()
	hovered := elem != nil && elem.IsHovered()

	switch b.variant {
	case ButtonPrimary:
		if b.pressed {
			return cfg.ActiveColor
		}
		if hovered {
			return cfg.HoverColor
		}
		return cfg.PrimaryColor
	case ButtonSecondary:
		if b.pressed {
			return uimath.ColorHex("#e6e6e6")
		}
		if hovered {
			return uimath.ColorHex("#f5f5f5")
		}
		return cfg.BgColor
	case ButtonText, ButtonLink:
		if b.pressed {
			return uimath.ColorHex("#e6f4ff")
		}
		if hovered {
			return uimath.ColorHex("#f0f5ff")
		}
		return uimath.ColorTransparent
	default:
		return cfg.PrimaryColor
	}
}

func (b *Button) textColor() uimath.Color {
	cfg := b.config
	if b.disabled {
		return cfg.BgColor
	}
	elem := b.Element()
	hovered := elem != nil && elem.IsHovered()

	switch b.variant {
	case ButtonPrimary:
		return uimath.ColorWhite
	case ButtonSecondary:
		if b.pressed {
			return cfg.ActiveColor
		}
		if hovered {
			return cfg.HoverColor
		}
		return cfg.TextColor
	case ButtonText:
		if b.pressed {
			return cfg.ActiveColor
		}
		if hovered {
			return cfg.HoverColor
		}
		return cfg.PrimaryColor
	case ButtonLink:
		if b.pressed {
			return cfg.ActiveColor
		}
		if hovered {
			return cfg.HoverColor
		}
		return cfg.PrimaryColor
	default:
		return uimath.ColorWhite
	}
}

func (b *Button) Draw(buf *render.CommandBuffer) {
	bounds := b.Bounds()
	if bounds.IsEmpty() {
		return
	}

	cfg := b.config

	// Background
	bg := b.bgColor()
	elem := b.Element()
	hovered := elem != nil && elem.IsHovered()
	borderClr := cfg.BorderColor
	borderW := cfg.BorderWidth

	switch b.variant {
	case ButtonPrimary:
		borderClr = uimath.ColorTransparent
		borderW = 0
	case ButtonSecondary:
		if b.pressed {
			borderClr = cfg.ActiveColor
		} else if hovered {
			borderClr = cfg.HoverColor
		}
	case ButtonText, ButtonLink:
		borderClr = uimath.ColorTransparent
		borderW = 0
	}

	buf.DrawRect(render.RectCmd{
		Bounds:      bounds,
		FillColor:   bg,
		BorderColor: borderClr,
		BorderWidth: borderW,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 0, 1)

	// Label text
	if b.label != "" {
		if cfg.TextRenderer != nil {
			tx := bounds.X + cfg.SpaceSM
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			ty := bounds.Y + (bounds.Height-lh)/2
			maxW := bounds.Width - cfg.SpaceSM*2
			cfg.TextRenderer.DrawText(buf, b.label, tx, ty, cfg.FontSize, maxW, b.textColor(), 1)
		} else {
			textW := float32(len(b.label)) * cfg.FontSize * 0.55
			textH := cfg.FontSize * 1.2
			if textW > bounds.Width-8 {
				textW = bounds.Width - 8
			}
			tx := bounds.X + (bounds.Width-textW)/2
			ty := bounds.Y + (bounds.Height-textH)/2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(tx, ty, textW, textH),
				FillColor: b.textColor(),
				Corners:   uimath.CornersAll(2),
			}, 1, 1)
		}
	}

	b.DrawChildren(buf)
}

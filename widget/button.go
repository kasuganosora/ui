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
	ButtonBase    ButtonVariant = iota // TDesign: base (filled)
	ButtonOutline                      // TDesign: outline (line border)
	ButtonDashed                       // TDesign: dashed border
	ButtonText                         // TDesign: text only

	// Aliases for backward compatibility
	ButtonPrimary   = ButtonBase
	ButtonSecondary = ButtonOutline
	ButtonLink      = ButtonText
)

// ButtonTheme controls the button color theme (TDesign).
type ButtonTheme uint8

const (
	ThemeDefault ButtonTheme = iota
	ThemePrimary
	ThemeDanger
	ThemeWarning
	ThemeSuccess
)

// ButtonShape controls the button shape (TDesign).
type ButtonShape uint8

const (
	ShapeRectangle ButtonShape = iota
	ShapeSquare
	ShapeRound
	ShapeCircle
)

// Button is an interactive button widget.
type Button struct {
	Base
	content  string
	variant  ButtonVariant
	disabled bool
	pressed  bool
	block    bool

	theme   ButtonTheme
	size    Size
	shape   ButtonShape
	ghost   bool
	loading bool

	onClick func()
}

// NewButton creates a button with the given label.
func NewButton(tree *core.Tree, label string, cfg *Config) *Button {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	b := &Button{
		Base:    NewBase(tree, core.TypeButton, cfg),
		content: label,
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
		if !b.disabled && !b.loading && b.onClick != nil {
			b.onClick()
		}
	})

	return b
}

func (b *Button) Content() string         { return b.content }
func (b *Button) Variant() ButtonVariant  { return b.variant }
func (b *Button) Disabled() bool          { return b.disabled }
func (b *Button) IsDisabled() bool        { return b.disabled } // backward compat
func (b *Button) IsPressed() bool         { return b.pressed }
func (b *Button) Block() bool             { return b.block }

// Label is an alias for Content (backward compatibility).
func (b *Button) Label() string { return b.content }

func (b *Button) SetContent(content string) {
	b.content = content
	b.tree.SetProperty(b.id, "text", content)
}

// SetLabel is an alias for SetContent (backward compatibility).
func (b *Button) SetLabel(label string) { b.SetContent(label) }

func (b *Button) SetBlock(block bool) {
	b.block = block
	if block {
		b.style.Width = layout.Pct(100)
	} else {
		b.style.Width = layout.Auto
	}
}

func (b *Button) SetVariant(v ButtonVariant) { b.variant = v }

func (b *Button) SetDisabled(d bool) {
	b.disabled = d
	b.tree.SetEnabled(b.id, !d)
}

func (b *Button) OnClick(fn func()) {
	b.onClick = fn
}

func (b *Button) Theme() ButtonTheme    { return b.theme }
func (b *Button) Size() Size            { return b.size }
func (b *Button) Shape() ButtonShape    { return b.shape }
func (b *Button) IsGhost() bool         { return b.ghost }
func (b *Button) IsLoading() bool       { return b.loading }

func (b *Button) SetTheme(t ButtonTheme) {
	b.theme = t
}

func (b *Button) SetSize(s Size) {
	b.size = s
	b.style.Height = layout.Px(b.config.SizeHeight(s))
}

func (b *Button) SetShape(s ButtonShape) {
	b.shape = s
}

func (b *Button) SetGhost(g bool) {
	b.ghost = g
}

func (b *Button) SetLoading(l bool) {
	b.loading = l
}

// themeBaseColor returns the base color for the current theme.
func (b *Button) themeBaseColor() uimath.Color {
	cfg := b.config
	switch b.theme {
	case ThemePrimary:
		return cfg.PrimaryColor
	case ThemeDanger:
		return uimath.ColorHex("#e34d59")
	case ThemeWarning:
		return uimath.ColorHex("#ed7b2f")
	case ThemeSuccess:
		return uimath.ColorHex("#2ba471")
	default:
		return uimath.Color{} // no override
	}
}

func (b *Button) bgColor() uimath.Color {
	cfg := b.config
	if b.disabled {
		return cfg.DisabledColor
	}
	elem := b.Element()
	hovered := elem != nil && elem.IsHovered()

	// Ghost mode: transparent background regardless of theme/variant
	if b.ghost {
		if b.pressed {
			base := b.themeBaseColor()
			if base.A == 0 && b.theme == ThemeDefault {
				return uimath.ColorHex("#e6f4ff")
			}
			base.A = 0.1
			return base
		}
		if hovered {
			base := b.themeBaseColor()
			if base.A == 0 && b.theme == ThemeDefault {
				return uimath.ColorHex("#f0f5ff")
			}
			base.A = 0.06
			return base
		}
		return uimath.ColorTransparent
	}

	// Theme override takes precedence for non-default themes
	if b.theme != ThemeDefault {
		base := b.themeBaseColor()
		if b.pressed {
			// Darken slightly
			base.R *= 0.85
			base.G *= 0.85
			base.B *= 0.85
			return base
		}
		if hovered {
			// Lighten slightly
			base.R = base.R + (1-base.R)*0.15
			base.G = base.G + (1-base.G)*0.15
			base.B = base.B + (1-base.B)*0.15
			return base
		}
		return base
	}

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
	case ButtonText:
		if b.pressed {
			return uimath.ColorHex("#e6f4ff")
		}
		if hovered {
			return uimath.ColorHex("#f0f5ff")
		}
		return uimath.ColorTransparent
	case ButtonDashed:
		if b.pressed {
			return uimath.ColorHex("#e6e6e6")
		}
		if hovered {
			return uimath.ColorHex("#f5f5f5")
		}
		return cfg.BgColor
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

	// Ghost mode: text color is the theme color
	if b.ghost {
		if b.theme != ThemeDefault {
			return b.themeBaseColor()
		}
		// Default ghost uses primary color
		if b.pressed {
			return cfg.ActiveColor
		}
		if hovered {
			return cfg.HoverColor
		}
		return cfg.PrimaryColor
	}

	// Non-default theme: white text on colored background
	if b.theme != ThemeDefault {
		return uimath.ColorWhite
	}

	switch b.variant {
	case ButtonBase:
		return uimath.ColorWhite
	case ButtonOutline, ButtonDashed:
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
	fontSize := cfg.SizeFontSize(b.size)

	// Background
	bg := b.bgColor()
	elem := b.Element()
	hovered := elem != nil && elem.IsHovered()
	borderClr := cfg.BorderColor
	borderW := cfg.BorderWidth

	// Ghost mode: outline style with theme-colored border
	if b.ghost {
		if b.theme != ThemeDefault {
			borderClr = b.themeBaseColor()
		} else {
			borderClr = cfg.PrimaryColor
		}
		borderW = cfg.BorderWidth
	} else if b.theme != ThemeDefault {
		borderClr = uimath.ColorTransparent
		borderW = 0
	} else {
		switch b.variant {
		case ButtonBase:
			borderClr = uimath.ColorTransparent
			borderW = 0
		case ButtonOutline:
			if b.pressed {
				borderClr = cfg.ActiveColor
			} else if hovered {
				borderClr = cfg.HoverColor
			}
		case ButtonDashed:
			if b.pressed {
				borderClr = cfg.ActiveColor
			} else if hovered {
				borderClr = cfg.HoverColor
			}
			// TODO: render dashed border style
		case ButtonText:
			borderClr = uimath.ColorTransparent
			borderW = 0
		}
	}

	// Shape determines border radius
	borderRadius := cfg.BorderRadius
	switch b.shape {
	case ShapeSquare:
		borderRadius = 0
	case ShapeRound:
		borderRadius = bounds.Height / 2
	case ShapeCircle:
		borderRadius = bounds.Height / 2
	}

	buf.DrawRect(render.RectCmd{
		Bounds:      bounds,
		FillColor:   bg,
		BorderColor: borderClr,
		BorderWidth: borderW,
		Corners:     uimath.CornersAll(borderRadius),
	}, 0, 1)

	// Label text (show "..." when loading)
	displayLabel := b.content
	if b.loading {
		displayLabel = "..."
	}
	if displayLabel != "" {
		if cfg.TextRenderer != nil {
			tx := bounds.X + cfg.SpaceSM
			lh := cfg.TextRenderer.LineHeight(fontSize)
			ty := bounds.Y + (bounds.Height-lh)/2
			maxW := bounds.Width - cfg.SpaceSM*2
			cfg.TextRenderer.DrawText(buf, displayLabel, tx, ty, fontSize, maxW, b.textColor(), 1)
		} else {
			textW := float32(len(displayLabel)) * fontSize * 0.55
			textH := fontSize * 1.2
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

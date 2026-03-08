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
	ButtonBase    ButtonVariant = iota // base (filled)
	ButtonOutline                      // outline (line border)
	ButtonDashed                       // dashed border
	ButtonText                         // text only

	// Aliases for backward compatibility
	ButtonPrimary   = ButtonBase
	ButtonSecondary = ButtonOutline
	ButtonLink      = ButtonText
)

// ButtonTheme controls the button color theme.
type ButtonTheme uint8

const (
	ThemeDefault ButtonTheme = iota
	ThemePrimary
	ThemeDanger
	ThemeWarning
	ThemeSuccess
)

// ButtonShape controls the button shape.
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
		return uimath.ColorHex("#d54941")
	case ThemeWarning:
		return uimath.ColorHex("#e37318")
	case ThemeSuccess:
		return uimath.ColorHex("#2ba471")
	default:
		return uimath.Color{} // no override
	}
}

func (b *Button) bgColor() uimath.Color {
	cfg := b.config
	if b.disabled {
		return uimath.ColorHex("#eeeeee") // gray-2
	}
	elem := b.Element()
	hovered := elem != nil && elem.IsHovered()

	// Ghost mode: transparent background regardless of theme/variant
	if b.ghost {
		if b.pressed {
			base := b.themeBaseColor()
			if base.A == 0 && b.theme == ThemeDefault {
				return uimath.ColorHex("#d9e1ff")
			}
			base.A = 0.1
			return base
		}
		if hovered {
			base := b.themeBaseColor()
			if base.A == 0 && b.theme == ThemeDefault {
				return uimath.ColorHex("#f2f3ff")
			}
			base.A = 0.06
			return base
		}
		return uimath.ColorTransparent
	}

	// Filled button with non-default theme: colored background
	if b.theme != ThemeDefault && b.variant == ButtonBase {
		base := b.themeBaseColor()
		if b.pressed {
			base.R *= 0.85
			base.G *= 0.85
			base.B *= 0.85
			return base
		}
		if hovered {
			base.R = base.R + (1-base.R)*0.15
			base.G = base.G + (1-base.G)*0.15
			base.B = base.B + (1-base.B)*0.15
			return base
		}
		return base
	}

	// Outline / Dashed / Text variants (all themes) + Default theme Base variant
	switch b.variant {
	case ButtonBase:
		// ThemeDefault + Base = default button (white bg, gray border)
		if b.pressed {
			return uimath.ColorHex("#e8e8e8")
		}
		if hovered {
			return uimath.ColorHex("#f3f3f3")
		}
		return cfg.BgColor
	case ButtonOutline:
		if b.pressed {
			return uimath.ColorHex("#e8e8e8")
		}
		if hovered {
			return uimath.ColorHex("#f3f3f3")
		}
		return cfg.BgColor
	case ButtonText:
		if b.pressed {
			return uimath.ColorHex("#e8e8e8")
		}
		if hovered {
			return uimath.ColorHex("#f3f3f3")
		}
		return uimath.ColorTransparent
	case ButtonDashed:
		if b.pressed {
			return uimath.ColorHex("#e8e8e8")
		}
		if hovered {
			return uimath.ColorHex("#f3f3f3")
		}
		return cfg.BgColor
	default:
		return cfg.BgColor
	}
}

func (b *Button) textColor() uimath.Color {
	cfg := b.config
	if b.disabled {
		return uimath.RGBA(0, 0, 0, 0.26) // text-disabled
	}
	elem := b.Element()
	hovered := elem != nil && elem.IsHovered()

	// Ghost mode: text color is the theme color
	if b.ghost {
		if b.theme != ThemeDefault {
			return b.themeBaseColor()
		}
		if b.pressed {
			return cfg.ActiveColor
		}
		if hovered {
			return cfg.HoverColor
		}
		return cfg.PrimaryColor
	}

	// Filled button with non-default theme: white text on colored bg
	if b.theme != ThemeDefault && b.variant == ButtonBase {
		return uimath.ColorWhite
	}

	// Outline/Dashed/Text with non-default theme: theme-colored text
	if b.theme != ThemeDefault {
		base := b.themeBaseColor()
		if b.pressed {
			base.R *= 0.85
			base.G *= 0.85
			base.B *= 0.85
		} else if hovered {
			base.R = base.R + (1-base.R)*0.15
			base.G = base.G + (1-base.G)*0.15
			base.B = base.B + (1-base.B)*0.15
		}
		return base
	}

	// ThemeDefault
	switch b.variant {
	case ButtonBase:
		// Default button: dark text
		return cfg.TextColor
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
		return cfg.TextColor
	default:
		return cfg.TextColor
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

	// Determine border color based on theme/variant
	if b.ghost {
		if b.theme != ThemeDefault {
			borderClr = b.themeBaseColor()
		} else {
			borderClr = cfg.PrimaryColor
		}
		borderW = cfg.BorderWidth
	} else if b.theme != ThemeDefault && b.variant == ButtonBase {
		// Filled themed button: no visible border
		borderClr = uimath.ColorTransparent
		borderW = 0
	} else if b.theme != ThemeDefault {
		// Outline/Dashed themed button: theme-colored border
		base := b.themeBaseColor()
		if b.pressed {
			base.R *= 0.85
			base.G *= 0.85
			base.B *= 0.85
		} else if hovered {
			base.R = base.R + (1-base.R)*0.15
			base.G = base.G + (1-base.G)*0.15
			base.B = base.B + (1-base.B)*0.15
		}
		borderClr = base
		borderW = cfg.BorderWidth
		if b.variant == ButtonText {
			borderClr = uimath.ColorTransparent
			borderW = 0
		}
	} else {
		// ThemeDefault
		switch b.variant {
		case ButtonBase:
			// Default button: gray border
			if b.pressed {
				borderClr = cfg.ActiveColor
			} else if hovered {
				borderClr = cfg.HoverColor
			}
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

	if b.variant == ButtonDashed && borderW > 0 && borderClr.A > 0 {
		// Background without border
		buf.DrawRect(render.RectCmd{
			Bounds:    bounds,
			FillColor: bg,
			Corners:   uimath.CornersAll(borderRadius),
		}, 0, 1)
		// Dashed border segments
		drawDashedBorder(buf, bounds, borderClr, borderW, borderRadius, 6, 4)
	} else {
		buf.DrawRect(render.RectCmd{
			Bounds:      bounds,
			FillColor:   bg,
			BorderColor: borderClr,
			BorderWidth: borderW,
			Corners:     uimath.CornersAll(borderRadius),
		}, 0, 1)
	}

	// Label text (show "..." when loading)
	displayLabel := b.content
	if b.loading {
		displayLabel = "..."
	}
	if displayLabel != "" {
		if cfg.TextRenderer != nil {
			tw := cfg.TextRenderer.MeasureText(displayLabel, fontSize)
			lh := cfg.TextRenderer.LineHeight(fontSize)
			tx := bounds.X + (bounds.Width-tw)/2
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

// drawDashedBorder draws a dashed border around a rectangle using small rect segments.
// dashLen and gapLen control the dash pattern in pixels.
func drawDashedBorder(buf *render.CommandBuffer, bounds uimath.Rect, color uimath.Color, borderW, radius, dashLen, gapLen float32) {
	x, y, w, h := bounds.X, bounds.Y, bounds.Width, bounds.Height
	step := dashLen + gapLen

	// Top edge (left to right, inset by radius)
	for dx := radius; dx < w-radius; dx += step {
		dw := dashLen
		if dx+dw > w-radius {
			dw = w - radius - dx
		}
		if dw > 0 {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(x+dx, y, dw, borderW),
				FillColor: color,
			}, 1, 1)
		}
	}

	// Bottom edge
	for dx := radius; dx < w-radius; dx += step {
		dw := dashLen
		if dx+dw > w-radius {
			dw = w - radius - dx
		}
		if dw > 0 {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(x+dx, y+h-borderW, dw, borderW),
				FillColor: color,
			}, 1, 1)
		}
	}

	// Left edge (top to bottom, inset by radius)
	for dy := radius; dy < h-radius; dy += step {
		dh := dashLen
		if dy+dh > h-radius {
			dh = h - radius - dy
		}
		if dh > 0 {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(x, y+dy, borderW, dh),
				FillColor: color,
			}, 1, 1)
		}
	}

	// Right edge
	for dy := radius; dy < h-radius; dy += step {
		dh := dashLen
		if dy+dh > h-radius {
			dh = h - radius - dy
		}
		if dh > 0 {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(x+w-borderW, y+dy, borderW, dh),
				FillColor: color,
			}, 1, 1)
		}
	}

	// Corners: draw small arc-approximation dashes at each corner
	// For small radii (typical button 3-6px), a single dash at each corner looks fine
	if radius > 0 {
		r := radius
		// Top-left corner
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x, y+r*0.3, borderW, r*0.5),
			FillColor: color,
		}, 1, 1)
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x+r*0.3, y, r*0.5, borderW),
			FillColor: color,
		}, 1, 1)
		// Top-right corner
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x+w-borderW, y+r*0.3, borderW, r*0.5),
			FillColor: color,
		}, 1, 1)
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x+w-r*0.8, y, r*0.5, borderW),
			FillColor: color,
		}, 1, 1)
		// Bottom-left corner
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x, y+h-r*0.8, borderW, r*0.5),
			FillColor: color,
		}, 1, 1)
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x+r*0.3, y+h-borderW, r*0.5, borderW),
			FillColor: color,
		}, 1, 1)
		// Bottom-right corner
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x+w-borderW, y+h-r*0.8, borderW, r*0.5),
			FillColor: color,
		}, 1, 1)
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x+w-r*0.8, y+h-borderW, r*0.5, borderW),
			FillColor: color,
		}, 1, 1)
	}
}

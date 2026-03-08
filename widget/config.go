package widget

import (
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/render"
)

// TextDrawer renders text into a command buffer.
// When set on Config, widgets use real text rendering instead of placeholder rects.
type TextDrawer interface {
	DrawText(buf *render.CommandBuffer, text string, x, y, fontSize, maxWidth float32, color uimath.Color, opacity float32)
	// LineHeight returns the font line height (ascent + descent) for the given size.
	// Used by widgets for accurate vertical centering.
	LineHeight(fontSize float32) float32
	// MeasureText returns the width of the given text at the specified font size.
	MeasureText(text string, fontSize float32) float32
}

// Size represents component size variants.
type Size uint8

const (
	SizeMedium Size = iota // default
	SizeSmall
	SizeLarge
)

// Status represents input validation state.
type Status uint8

const (
	StatusDefault Status = iota
	StatusSuccess
	StatusWarning
	StatusError
)

// Config holds global UI configuration and theme colors.
// Acts as a ConfigProvider — pass it when creating widgets.
type Config struct {
	// Theme colors
	PrimaryColor     uimath.Color
	TextColor        uimath.Color
	BgColor          uimath.Color
	BorderColor      uimath.Color
	DisabledColor    uimath.Color
	HoverColor       uimath.Color
	ActiveColor      uimath.Color
	FocusBorderColor uimath.Color
	ErrorColor       uimath.Color
	WarningColor     uimath.Color
	SuccessColor     uimath.Color

	// Typography
	FontID     uint32
	FontSize   float32
	FontSizeSm float32
	FontSizeLg float32
	LineHeight float32

	// Text renderer (optional — falls back to placeholder rects if nil)
	TextRenderer TextDrawer

	// Window reference for IME positioning and context menus (optional)
	Window platform.Window

	// Platform reference for clipboard operations (optional)
	Platform platform.Platform

	// Spacing
	SpaceXS float32
	SpaceSM float32
	SpaceMD float32
	SpaceLG float32
	SpaceXL float32

	// Border
	BorderRadius float32
	BorderWidth  float32

	// Component defaults
	ButtonHeight float32
	InputHeight  float32
	IconSize     float32

	// Icon registry for SVG icon lookup by name (e.g., Material Design Icons)
	IconRegistry IconLookup
}

// IconLookup provides icon texture lookup by name and size.
type IconLookup interface {
	Get(name string, size int) (render.TextureHandle, bool)
	Has(name string) bool
}

// DefaultConfig returns a default configuration with light theme defaults.
func DefaultConfig() *Config {
	return &Config{
		// brand-color-7 (#0052d9)
		PrimaryColor:     uimath.ColorHex("#0052d9"),
		TextColor:        uimath.RGBA(0, 0, 0, 0.9),
		BgColor:          uimath.ColorWhite,
		BorderColor:      uimath.ColorHex("#dcdcdc"),
		DisabledColor:    uimath.RGBA(0, 0, 0, 0.26),
		HoverColor:       uimath.ColorHex("#366ef4"),
		ActiveColor:      uimath.ColorHex("#003cab"),
		FocusBorderColor: uimath.ColorHex("#0052d9"),
		ErrorColor:       uimath.ColorHex("#d54941"),
		WarningColor:     uimath.ColorHex("#e37318"),
		SuccessColor:     uimath.ColorHex("#2ba471"),

		FontID:     0,
		FontSize:   14,
		FontSizeSm: 12,
		FontSizeLg: 16,
		LineHeight: 1.5,

		SpaceXS: 4,
		SpaceSM: 8,
		SpaceMD: 16,
		SpaceLG: 24,
		SpaceXL: 32,

		// radius-medium = 6
		BorderRadius: 6,
		BorderWidth:  1,

		ButtonHeight: 32,
		InputHeight:  32,
		IconSize:     16,
	}
}

// SizeHeight returns the component height for a given size.
func (c *Config) SizeHeight(s Size) float32 {
	switch s {
	case SizeSmall:
		return 24
	case SizeLarge:
		return 40
	default:
		return 32
	}
}

// SizeFontSize returns the font size for a given component size.
func (c *Config) SizeFontSize(s Size) float32 {
	switch s {
	case SizeSmall:
		return c.FontSizeSm
	case SizeLarge:
		return c.FontSizeLg
	default:
		return c.FontSize
	}
}

// StatusBorderColor returns the border color for a validation status.
func (c *Config) StatusBorderColor(s Status) uimath.Color {
	switch s {
	case StatusSuccess:
		return c.SuccessColor
	case StatusWarning:
		return c.WarningColor
	case StatusError:
		return c.ErrorColor
	default:
		return c.BorderColor
	}
}

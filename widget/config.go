package widget

import uimath "github.com/kasuganosora/ui/math"

// Config holds global UI configuration and theme colors.
// Acts as a ConfigProvider — pass it when creating widgets.
type Config struct {
	// Theme colors
	PrimaryColor   uimath.Color
	TextColor      uimath.Color
	BgColor        uimath.Color
	BorderColor    uimath.Color
	DisabledColor  uimath.Color
	HoverColor     uimath.Color
	ActiveColor    uimath.Color
	FocusBorderColor uimath.Color
	ErrorColor     uimath.Color

	// Typography
	FontID       uint32
	FontSize     float32
	FontSizeSm   float32
	FontSizeLg   float32
	LineHeight   float32

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
	ButtonHeight    float32
	InputHeight     float32
	IconSize        float32
}

// DefaultConfig returns a default configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		PrimaryColor:     uimath.ColorHex("#1677ff"),
		TextColor:        uimath.ColorHex("#333333"),
		BgColor:          uimath.ColorWhite,
		BorderColor:      uimath.ColorHex("#d9d9d9"),
		DisabledColor:    uimath.ColorHex("#bfbfbf"),
		HoverColor:       uimath.ColorHex("#4096ff"),
		ActiveColor:      uimath.ColorHex("#0958d9"),
		FocusBorderColor: uimath.ColorHex("#4096ff"),
		ErrorColor:       uimath.ColorHex("#ff4d4f"),

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

		BorderRadius: 6,
		BorderWidth:  1,

		ButtonHeight: 32,
		InputHeight:  32,
		IconSize:     16,
	}
}

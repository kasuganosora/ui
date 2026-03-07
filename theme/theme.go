// Package theme provides a token-based theming system.
//
// Themes define color, spacing, typography, and shape tokens.
// Widgets read tokens from the active theme instead of hardcoded values.
package theme

import uimath "github.com/kasuganosora/ui/math"

// Theme holds all design tokens for the UI.
type Theme struct {
	Name string

	// Colors
	Primary      uimath.Color
	PrimaryHover uimath.Color
	PrimaryActive uimath.Color

	Success      uimath.Color
	Warning      uimath.Color
	Error        uimath.Color
	Info         uimath.Color

	TextPrimary   uimath.Color
	TextSecondary uimath.Color
	TextDisabled  uimath.Color
	TextInverse   uimath.Color

	BgPrimary    uimath.Color
	BgSecondary  uimath.Color
	BgElevated   uimath.Color

	Border       uimath.Color
	BorderHover  uimath.Color
	BorderFocus  uimath.Color

	Divider      uimath.Color
	Mask         uimath.Color // Modal overlay color

	// Typography
	FontSizeXS  float32
	FontSizeSM  float32
	FontSizeMD  float32
	FontSizeLG  float32
	FontSizeXL  float32
	FontSizeXXL float32
	LineHeight  float32

	// Spacing
	SpaceXXS float32
	SpaceXS  float32
	SpaceSM  float32
	SpaceMD  float32
	SpaceLG  float32
	SpaceXL  float32
	SpaceXXL float32

	// Shape
	RadiusSM float32
	RadiusMD float32
	RadiusLG float32
	RadiusFull float32 // For pills/circles

	// Borders
	BorderWidth float32

	// Component sizes
	HeightSM float32
	HeightMD float32
	HeightLG float32

	// Shadows (represented as alpha/offset for simple shadow support)
	ShadowColor uimath.Color
	ShadowSmY   float32
	ShadowMdY   float32
	ShadowLgY   float32
}

// Light returns the default light theme (Ant Design inspired).
func Light() *Theme {
	return &Theme{
		Name: "light",

		Primary:       uimath.ColorHex("#1677ff"),
		PrimaryHover:  uimath.ColorHex("#4096ff"),
		PrimaryActive: uimath.ColorHex("#0958d9"),

		Success: uimath.ColorHex("#52c41a"),
		Warning: uimath.ColorHex("#faad14"),
		Error:   uimath.ColorHex("#ff4d4f"),
		Info:    uimath.ColorHex("#1677ff"),

		TextPrimary:   uimath.RGBA(0, 0, 0, 0.88),
		TextSecondary: uimath.RGBA(0, 0, 0, 0.65),
		TextDisabled:  uimath.RGBA(0, 0, 0, 0.25),
		TextInverse:   uimath.ColorWhite,

		BgPrimary:   uimath.ColorWhite,
		BgSecondary: uimath.ColorHex("#f5f5f5"),
		BgElevated:  uimath.ColorWhite,

		Border:      uimath.ColorHex("#d9d9d9"),
		BorderHover: uimath.ColorHex("#4096ff"),
		BorderFocus: uimath.ColorHex("#4096ff"),

		Divider: uimath.RGBA(0, 0, 0, 0.06),
		Mask:    uimath.RGBA(0, 0, 0, 0.45),

		FontSizeXS:  10,
		FontSizeSM:  12,
		FontSizeMD:  14,
		FontSizeLG:  16,
		FontSizeXL:  20,
		FontSizeXXL: 24,
		LineHeight:   1.5,

		SpaceXXS: 2,
		SpaceXS:  4,
		SpaceSM:  8,
		SpaceMD:  16,
		SpaceLG:  24,
		SpaceXL:  32,
		SpaceXXL: 48,

		RadiusSM:   4,
		RadiusMD:   6,
		RadiusLG:   8,
		RadiusFull: 9999,

		BorderWidth: 1,

		HeightSM: 24,
		HeightMD: 32,
		HeightLG: 40,

		ShadowColor: uimath.RGBA(0, 0, 0, 0.12),
		ShadowSmY:   2,
		ShadowMdY:   4,
		ShadowLgY:   8,
	}
}

// Dark returns a dark theme.
func Dark() *Theme {
	return &Theme{
		Name: "dark",

		Primary:       uimath.ColorHex("#1668dc"),
		PrimaryHover:  uimath.ColorHex("#3c89e8"),
		PrimaryActive: uimath.ColorHex("#1554ad"),

		Success: uimath.ColorHex("#49aa19"),
		Warning: uimath.ColorHex("#d89614"),
		Error:   uimath.ColorHex("#dc4446"),
		Info:    uimath.ColorHex("#1668dc"),

		TextPrimary:   uimath.RGBA(1, 1, 1, 0.85),
		TextSecondary: uimath.RGBA(1, 1, 1, 0.65),
		TextDisabled:  uimath.RGBA(1, 1, 1, 0.30),
		TextInverse:   uimath.ColorHex("#141414"),

		BgPrimary:   uimath.ColorHex("#141414"),
		BgSecondary: uimath.ColorHex("#1f1f1f"),
		BgElevated:  uimath.ColorHex("#2a2a2a"),

		Border:      uimath.ColorHex("#424242"),
		BorderHover: uimath.ColorHex("#3c89e8"),
		BorderFocus: uimath.ColorHex("#3c89e8"),

		Divider: uimath.RGBA(1, 1, 1, 0.12),
		Mask:    uimath.RGBA(0, 0, 0, 0.65),

		FontSizeXS:  10,
		FontSizeSM:  12,
		FontSizeMD:  14,
		FontSizeLG:  16,
		FontSizeXL:  20,
		FontSizeXXL: 24,
		LineHeight:   1.5,

		SpaceXXS: 2,
		SpaceXS:  4,
		SpaceSM:  8,
		SpaceMD:  16,
		SpaceLG:  24,
		SpaceXL:  32,
		SpaceXXL: 48,

		RadiusSM:   4,
		RadiusMD:   6,
		RadiusLG:   8,
		RadiusFull: 9999,

		BorderWidth: 1,

		HeightSM: 24,
		HeightMD: 32,
		HeightLG: 40,

		ShadowColor: uimath.RGBA(0, 0, 0, 0.30),
		ShadowSmY:   2,
		ShadowMdY:   4,
		ShadowLgY:   8,
	}
}

// ToConfig converts a Theme to the widget.Config format for backward compatibility.
// This allows gradual migration from Config to Theme.
func (t *Theme) ToConfig() ConfigValues {
	return ConfigValues{
		PrimaryColor:     t.Primary,
		TextColor:        t.TextPrimary,
		BgColor:          t.BgPrimary,
		BorderColor:      t.Border,
		DisabledColor:    t.TextDisabled,
		HoverColor:       t.PrimaryHover,
		ActiveColor:      t.PrimaryActive,
		FocusBorderColor: t.BorderFocus,
		ErrorColor:       t.Error,
		FontSize:         t.FontSizeMD,
		FontSizeSm:       t.FontSizeSM,
		FontSizeLg:       t.FontSizeLG,
		SpaceXS:          t.SpaceXS,
		SpaceSM:          t.SpaceSM,
		SpaceMD:          t.SpaceMD,
		SpaceLG:          t.SpaceLG,
		SpaceXL:          t.SpaceXL,
		BorderRadius:     t.RadiusMD,
		BorderWidth:      t.BorderWidth,
		ButtonHeight:     t.HeightMD,
		InputHeight:      t.HeightMD,
	}
}

// ConfigValues holds the flattened values that map to widget.Config fields.
type ConfigValues struct {
	PrimaryColor     uimath.Color
	TextColor        uimath.Color
	BgColor          uimath.Color
	BorderColor      uimath.Color
	DisabledColor    uimath.Color
	HoverColor       uimath.Color
	ActiveColor      uimath.Color
	FocusBorderColor uimath.Color
	ErrorColor       uimath.Color
	FontSize         float32
	FontSizeSm       float32
	FontSizeLg       float32
	SpaceXS          float32
	SpaceSM          float32
	SpaceMD          float32
	SpaceLG          float32
	SpaceXL          float32
	BorderRadius     float32
	BorderWidth      float32
	ButtonHeight     float32
	InputHeight      float32
}

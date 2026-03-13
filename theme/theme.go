// Package theme provides a token-based theming system.
//
// Themes define color, spacing, typography, and shape tokens.
// Widgets read tokens from the active theme instead of hardcoded values.
package theme

import (
	"fmt"

	uimath "github.com/kasuganosora/ui/math"
)

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

// Light returns the default light theme.
func Light() *Theme {
	return &Theme{
		Name: "light",

		// Brand palette: #0052d9
		Primary:       uimath.ColorHex("#0052d9"),
		PrimaryHover:  uimath.ColorHex("#366ef4"),
		PrimaryActive: uimath.ColorHex("#003cab"),

		// Functional colors
		Success: uimath.ColorHex("#2ba471"),
		Warning: uimath.ColorHex("#e37318"),
		Error:   uimath.ColorHex("#d54941"),
		Info:    uimath.ColorHex("#0052d9"),

		// Text colors
		TextPrimary:   uimath.RGBA(0, 0, 0, 0.9),
		TextSecondary: uimath.RGBA(0, 0, 0, 0.6),
		TextDisabled:  uimath.RGBA(0, 0, 0, 0.26),
		TextInverse:   uimath.ColorWhite,

		// Background colors
		BgPrimary:   uimath.ColorWhite,
		BgSecondary: uimath.ColorHex("#f3f3f3"),
		BgElevated:  uimath.ColorWhite,

		// Border: component-stroke
		Border:      uimath.ColorHex("#dcdcdc"),
		BorderHover: uimath.ColorHex("#366ef4"),
		BorderFocus: uimath.ColorHex("#0052d9"),

		Divider: uimath.RGBA(0, 0, 0, 0.06),
		Mask:    uimath.RGBA(0, 0, 0, 0.6),

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

		// Radius: default=3, medium=6, large=9
		RadiusSM:   3,
		RadiusMD:   6,
		RadiusLG:   9,
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

		Primary:       uimath.ColorHex("#4582e6"),
		PrimaryHover:  uimath.ColorHex("#618dff"),
		PrimaryActive: uimath.ColorHex("#366ef4"),

		Success: uimath.ColorHex("#56c08d"),
		Warning: uimath.ColorHex("#fa9550"),
		Error:   uimath.ColorHex("#f6685d"),
		Info:    uimath.ColorHex("#4582e6"),

		TextPrimary:   uimath.RGBA(1, 1, 1, 0.9),
		TextSecondary: uimath.RGBA(1, 1, 1, 0.6),
		TextDisabled:  uimath.RGBA(1, 1, 1, 0.26),
		TextInverse:   uimath.ColorHex("#181818"),

		BgPrimary:   uimath.ColorHex("#242424"),
		BgSecondary: uimath.ColorHex("#2c2c2c"),
		BgElevated:  uimath.ColorHex("#393939"),

		Border:      uimath.ColorHex("#4b4b4b"),
		BorderHover: uimath.ColorHex("#618dff"),
		BorderFocus: uimath.ColorHex("#4582e6"),

		Divider: uimath.RGBA(1, 1, 1, 0.12),
		Mask:    uimath.RGBA(0, 0, 0, 0.6),

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

		RadiusSM:   3,
		RadiusMD:   6,
		RadiusLG:   9,
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

// Blue returns a soft blue theme suited for dashboards and admin panels.
func Blue() *Theme {
	t := Light()
	t.Name = "blue"
	t.Primary = uimath.ColorHex("#1677ff")
	t.PrimaryHover = uimath.ColorHex("#4096ff")
	t.PrimaryActive = uimath.ColorHex("#0958d9")
	t.Success = uimath.ColorHex("#52c41a")
	t.Warning = uimath.ColorHex("#faad14")
	t.Error = uimath.ColorHex("#ff4d4f")
	t.Info = uimath.ColorHex("#1677ff")
	t.BorderFocus = uimath.ColorHex("#1677ff")
	t.BorderHover = uimath.ColorHex("#4096ff")
	return t
}

// Green returns a nature/eco green theme.
func Green() *Theme {
	t := Light()
	t.Name = "green"
	t.Primary = uimath.ColorHex("#389e0d")
	t.PrimaryHover = uimath.ColorHex("#52c41a")
	t.PrimaryActive = uimath.ColorHex("#237804")
	t.Success = uimath.ColorHex("#389e0d")
	t.Warning = uimath.ColorHex("#d48806")
	t.Error = uimath.ColorHex("#cf1322")
	t.Info = uimath.ColorHex("#389e0d")
	t.BorderFocus = uimath.ColorHex("#389e0d")
	t.BorderHover = uimath.ColorHex("#52c41a")
	return t
}

// Purple returns a creative/design purple theme.
func Purple() *Theme {
	t := Light()
	t.Name = "purple"
	t.Primary = uimath.ColorHex("#722ed1")
	t.PrimaryHover = uimath.ColorHex("#9254de")
	t.PrimaryActive = uimath.ColorHex("#531dab")
	t.Success = uimath.ColorHex("#52c41a")
	t.Warning = uimath.ColorHex("#faad14")
	t.Error = uimath.ColorHex("#f5222d")
	t.Info = uimath.ColorHex("#722ed1")
	t.BorderFocus = uimath.ColorHex("#722ed1")
	t.BorderHover = uimath.ColorHex("#9254de")
	return t
}

// Warm returns a warm-toned theme with orange accents.
func Warm() *Theme {
	t := Light()
	t.Name = "warm"
	t.Primary = uimath.ColorHex("#d4380d")
	t.PrimaryHover = uimath.ColorHex("#ff7a45")
	t.PrimaryActive = uimath.ColorHex("#ad2102")
	t.Success = uimath.ColorHex("#7cb305")
	t.Warning = uimath.ColorHex("#d48806")
	t.Error = uimath.ColorHex("#cf1322")
	t.Info = uimath.ColorHex("#d4380d")
	t.BgPrimary = uimath.ColorHex("#fffbf0")
	t.BgSecondary = uimath.ColorHex("#fff1e6")
	t.BorderFocus = uimath.ColorHex("#d4380d")
	t.BorderHover = uimath.ColorHex("#ff7a45")
	return t
}

// GameDark returns a dark theme optimized for game UIs with high contrast.
func GameDark() *Theme {
	t := Dark()
	t.Name = "game-dark"
	t.Primary = uimath.ColorHex("#f0b400")
	t.PrimaryHover = uimath.ColorHex("#ffd040")
	t.PrimaryActive = uimath.ColorHex("#c89600")
	t.Success = uimath.ColorHex("#73d13d")
	t.Warning = uimath.ColorHex("#ffa940")
	t.Error = uimath.ColorHex("#ff4d4f")
	t.Info = uimath.ColorHex("#40a9ff")
	t.BgPrimary = uimath.ColorHex("#1a1a2e")
	t.BgSecondary = uimath.ColorHex("#16213e")
	t.BgElevated = uimath.ColorHex("#0f3460")
	t.Border = uimath.ColorHex("#2a2a4a")
	t.BorderFocus = uimath.ColorHex("#f0b400")
	t.BorderHover = uimath.ColorHex("#ffd040")
	t.Divider = uimath.RGBA(1, 1, 1, 0.08)
	return t
}

// Compact returns a compact theme with smaller spacing and fonts, suitable for dense UIs.
func Compact() *Theme {
	t := Light()
	t.Name = "compact"
	t.FontSizeXS = 10
	t.FontSizeSM = 11
	t.FontSizeMD = 12
	t.FontSizeLG = 14
	t.FontSizeXL = 16
	t.FontSizeXXL = 20
	t.SpaceXXS = 1
	t.SpaceXS = 2
	t.SpaceSM = 4
	t.SpaceMD = 8
	t.SpaceLG = 12
	t.SpaceXL = 16
	t.SpaceXXL = 24
	t.RadiusSM = 2
	t.RadiusMD = 4
	t.RadiusLG = 6
	t.HeightSM = 20
	t.HeightMD = 28
	t.HeightLG = 36
	return t
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

// formatColorRGBA formats a Color as rgba(R,G,B,A) with 0-255 range.
func formatColorRGBA(c uimath.Color) string {
	r := int(c.R*255 + 0.5)
	g := int(c.G*255 + 0.5)
	b := int(c.B*255 + 0.5)
	a := int(c.A*255 + 0.5)
	return fmt.Sprintf("rgba(%d,%d,%d,%d)", r, g, b, a)
}

// formatPx formats a float32 as a plain number with "px" suffix.
func formatPx(v float32) string {
	return fmt.Sprintf("%gpx", v)
}

// ToCSSVariables converts all Theme tokens to CSS variable names using --ui-* convention.
func (t *Theme) ToCSSVariables() map[string]string {
	m := make(map[string]string)

	// Colors
	m["--ui-brand-color"] = formatColorRGBA(t.Primary)
	m["--ui-brand-color-hover"] = formatColorRGBA(t.PrimaryHover)
	m["--ui-brand-color-active"] = formatColorRGBA(t.PrimaryActive)
	m["--ui-success-color"] = formatColorRGBA(t.Success)
	m["--ui-warning-color"] = formatColorRGBA(t.Warning)
	m["--ui-error-color"] = formatColorRGBA(t.Error)
	m["--ui-info-color"] = formatColorRGBA(t.Info)
	m["--ui-text-color-primary"] = formatColorRGBA(t.TextPrimary)
	m["--ui-text-color-secondary"] = formatColorRGBA(t.TextSecondary)
	m["--ui-text-color-disabled"] = formatColorRGBA(t.TextDisabled)
	m["--ui-text-color-inverse"] = formatColorRGBA(t.TextInverse)
	m["--ui-bg-color-container"] = formatColorRGBA(t.BgPrimary)
	m["--ui-bg-color-secondarycontainer"] = formatColorRGBA(t.BgSecondary)
	m["--ui-bg-color-elevated"] = formatColorRGBA(t.BgElevated)
	m["--ui-component-stroke"] = formatColorRGBA(t.Border)
	m["--ui-component-stroke-hover"] = formatColorRGBA(t.BorderHover)
	m["--ui-component-stroke-focus"] = formatColorRGBA(t.BorderFocus)
	m["--ui-component-stroke-divider"] = formatColorRGBA(t.Divider)
	m["--ui-mask-color"] = formatColorRGBA(t.Mask)

	// Typography
	m["--ui-font-size-body-extra-small"] = formatPx(t.FontSizeXS)
	m["--ui-font-size-body-small"] = formatPx(t.FontSizeSM)
	m["--ui-font-size-body-medium"] = formatPx(t.FontSizeMD)
	m["--ui-font-size-body-large"] = formatPx(t.FontSizeLG)
	m["--ui-font-size-body-extra-large"] = formatPx(t.FontSizeXL)
	m["--ui-font-size-body-extra-extra-large"] = formatPx(t.FontSizeXXL)
	m["--ui-line-height"] = fmt.Sprintf("%g", t.LineHeight)

	// Spacing
	m["--ui-comp-paddingLR-xxs"] = formatPx(t.SpaceXXS)
	m["--ui-comp-paddingLR-xs"] = formatPx(t.SpaceXS)
	m["--ui-comp-paddingLR-s"] = formatPx(t.SpaceSM)
	m["--ui-comp-paddingLR-l"] = formatPx(t.SpaceMD)
	m["--ui-comp-paddingLR-xl"] = formatPx(t.SpaceLG)
	m["--ui-comp-paddingLR-xxl"] = formatPx(t.SpaceXL)
	m["--ui-comp-paddingLR-xxxl"] = formatPx(t.SpaceXXL)

	// Shape
	m["--ui-radius-small"] = formatPx(t.RadiusSM)
	m["--ui-radius-medium"] = formatPx(t.RadiusMD)
	m["--ui-radius-large"] = formatPx(t.RadiusLG)
	m["--ui-radius-round"] = formatPx(t.RadiusFull)

	// Borders
	m["--ui-border-width"] = formatPx(t.BorderWidth)

	// Component sizes
	m["--ui-comp-size-s"] = formatPx(t.HeightSM)
	m["--ui-comp-size-m"] = formatPx(t.HeightMD)
	m["--ui-comp-size-l"] = formatPx(t.HeightLG)

	// Shadows
	m["--ui-shadow-color"] = formatColorRGBA(t.ShadowColor)
	m["--ui-shadow-sm-y"] = formatPx(t.ShadowSmY)
	m["--ui-shadow-md-y"] = formatPx(t.ShadowMdY)
	m["--ui-shadow-lg-y"] = formatPx(t.ShadowLgY)

	return m
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

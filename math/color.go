package math

import "fmt"

// Color is an immutable RGBA color value object.
// Components are in [0, 1] range.
type Color struct {
	R, G, B, A float32
}

func NewColor(r, g, b, a float32) Color {
	return Color{R: r, G: g, B: b, A: a}
}

func RGB(r, g, b float32) Color {
	return Color{R: r, G: g, B: b, A: 1}
}

func RGBA(r, g, b, a float32) Color {
	return Color{R: r, G: g, B: b, A: a}
}

// ColorHex parses a hex color string (#RGB, #RGBA, #RRGGBB, #RRGGBBAA).
func ColorHex(hex string) Color {
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}
	switch len(hex) {
	case 3: // RGB
		r := hexNibble(hex[0])
		g := hexNibble(hex[1])
		b := hexNibble(hex[2])
		return Color{R: r, G: g, B: b, A: 1}
	case 4: // RGBA
		r := hexNibble(hex[0])
		g := hexNibble(hex[1])
		b := hexNibble(hex[2])
		a := hexNibble(hex[3])
		return Color{R: r, G: g, B: b, A: a}
	case 6: // RRGGBB
		r := hexByte(hex[0], hex[1])
		g := hexByte(hex[2], hex[3])
		b := hexByte(hex[4], hex[5])
		return Color{R: r, G: g, B: b, A: 1}
	case 8: // RRGGBBAA
		r := hexByte(hex[0], hex[1])
		g := hexByte(hex[2], hex[3])
		b := hexByte(hex[4], hex[5])
		a := hexByte(hex[6], hex[7])
		return Color{R: r, G: g, B: b, A: a}
	default:
		return Color{}
	}
}

// ColorRGBA8 creates a Color from 8-bit RGBA values (0-255).
func ColorRGBA8(r, g, b, a uint8) Color {
	return Color{
		R: float32(r) / 255,
		G: float32(g) / 255,
		B: float32(b) / 255,
		A: float32(a) / 255,
	}
}

func (c Color) WithAlpha(a float32) Color {
	return Color{R: c.R, G: c.G, B: c.B, A: a}
}

func (c Color) Lerp(other Color, t float32) Color {
	return Color{
		R: lerp32(c.R, other.R, t),
		G: lerp32(c.G, other.G, t),
		B: lerp32(c.B, other.B, t),
		A: lerp32(c.A, other.A, t),
	}
}

func (c Color) Mul(scalar float32) Color {
	return Color{R: c.R * scalar, G: c.G * scalar, B: c.B * scalar, A: c.A}
}

func (c Color) MulColor(other Color) Color {
	return Color{
		R: c.R * other.R,
		G: c.G * other.G,
		B: c.B * other.B,
		A: c.A * other.A,
	}
}

// RGBA8 returns 8-bit RGBA values.
func (c Color) RGBA8() (r, g, b, a uint8) {
	return uint8(clamp32(c.R, 0, 1) * 255),
		uint8(clamp32(c.G, 0, 1) * 255),
		uint8(clamp32(c.B, 0, 1) * 255),
		uint8(clamp32(c.A, 0, 1) * 255)
}

// Hex returns the color as a hex string (#RRGGBB or #RRGGBBAA if alpha != 1).
func (c Color) Hex() string {
	r, g, b, a := c.RGBA8()
	if a == 255 {
		return fmt.Sprintf("#%02x%02x%02x", r, g, b)
	}
	return fmt.Sprintf("#%02x%02x%02x%02x", r, g, b, a)
}

func (c Color) IsTransparent() bool {
	return c.A <= 0
}

func (c Color) Approx(other Color, epsilon float32) bool {
	return abs32(c.R-other.R) < epsilon &&
		abs32(c.G-other.G) < epsilon &&
		abs32(c.B-other.B) < epsilon &&
		abs32(c.A-other.A) < epsilon
}

// Predefined colors
var (
	ColorTransparent = Color{R: 0, G: 0, B: 0, A: 0}
	ColorBlack       = Color{R: 0, G: 0, B: 0, A: 1}
	ColorWhite       = Color{R: 1, G: 1, B: 1, A: 1}
	ColorRed         = Color{R: 1, G: 0, B: 0, A: 1}
	ColorGreen       = Color{R: 0, G: 1, B: 0, A: 1}
	ColorBlue        = Color{R: 0, G: 0, B: 1, A: 1}
	ColorYellow      = Color{R: 1, G: 1, B: 0, A: 1}
	ColorCyan        = Color{R: 0, G: 1, B: 1, A: 1}
	ColorMagenta     = Color{R: 1, G: 0, B: 1, A: 1}
	ColorGray        = Color{R: 0.5, G: 0.5, B: 0.5, A: 1}
)

func hexVal(c byte) float32 {
	switch {
	case c >= '0' && c <= '9':
		return float32(c - '0')
	case c >= 'a' && c <= 'f':
		return float32(c - 'a' + 10)
	case c >= 'A' && c <= 'F':
		return float32(c - 'A' + 10)
	default:
		return 0
	}
}

func hexNibble(c byte) float32 {
	v := hexVal(c)
	return v / 15
}

func hexByte(hi, lo byte) float32 {
	return (hexVal(hi)*16 + hexVal(lo)) / 255
}

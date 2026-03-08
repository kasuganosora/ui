package css

import (
	"strings"

	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
)

// ResolveVar replaces var(--name) and var(--name, fallback) references in a value string.
func ResolveVar(value string, vars map[string]string) string {
	// Limit iterations to prevent infinite loops from circular variable references.
	for range 32 {
		idx := strings.Index(value, "var(")
		if idx < 0 {
			break
		}
		// Find matching )
		depth := 1
		end := idx + 4
		for end < len(value) && depth > 0 {
			if value[end] == '(' {
				depth++
			} else if value[end] == ')' {
				depth--
			}
			if depth > 0 {
				end++
			}
		}
		if depth != 0 {
			break
		}
		inner := value[idx+4 : end] // content between var( and )
		end++                       // skip )

		varName := inner
		fallback := ""
		if ci := strings.Index(inner, ","); ci >= 0 {
			varName = strings.TrimSpace(inner[:ci])
			fallback = strings.TrimSpace(inner[ci+1:])
		} else {
			varName = strings.TrimSpace(varName)
		}

		replacement := fallback
		if v, ok := vars[varName]; ok {
			replacement = v
		}
		value = value[:idx] + replacement + value[end:]
	}
	return value
}

// ParseValue parses a CSS dimension value like "12px", "50%", "auto", "1.5em".
func ParseValue(val string) layout.Value {
	val = strings.TrimSpace(val)
	if val == "" || val == "auto" {
		return layout.Auto
	}
	if val == "0" {
		return layout.Zero
	}

	if strings.HasSuffix(val, "px") {
		n := parseFloat32(strings.TrimSuffix(val, "px"))
		return layout.Px(n)
	}
	if strings.HasSuffix(val, "%") {
		n := parseFloat32(strings.TrimSuffix(val, "%"))
		return layout.Pct(n)
	}
	if strings.HasSuffix(val, "em") || strings.HasSuffix(val, "rem") {
		// Treat em/rem as px * 16 (default font size)
		s := strings.TrimSuffix(strings.TrimSuffix(val, "rem"), "em")
		n := parseFloat32(s)
		return layout.Px(n * 16)
	}

	// Try as plain number (treat as px)
	if len(val) > 0 && (val[0] == '-' || val[0] == '.' || (val[0] >= '0' && val[0] <= '9')) {
		n := parseFloat32(val)
		return layout.Px(n)
	}
	return layout.Auto
}

// ParseFloat parses a float from a CSS value, stripping units.
func ParseFloat(val string) float32 {
	val = strings.TrimSpace(val)
	val = strings.TrimSuffix(val, "px")
	val = strings.TrimSuffix(val, "em")
	val = strings.TrimSuffix(val, "rem")
	val = strings.TrimSuffix(val, "%")
	return parseFloat32(val)
}

// ParseEdgeShorthand parses 1-4 value CSS shorthands (margin, padding).
// Returns top, right, bottom, left.
func ParseEdgeShorthand(val string) (layout.Value, layout.Value, layout.Value, layout.Value) {
	parts := splitValues(val)
	switch len(parts) {
	case 1:
		v := ParseValue(parts[0])
		return v, v, v, v
	case 2:
		tb := ParseValue(parts[0])
		lr := ParseValue(parts[1])
		return tb, lr, tb, lr
	case 3:
		t := ParseValue(parts[0])
		lr := ParseValue(parts[1])
		b := ParseValue(parts[2])
		return t, lr, b, lr
	case 4:
		return ParseValue(parts[0]), ParseValue(parts[1]),
			ParseValue(parts[2]), ParseValue(parts[3])
	default:
		return layout.Auto, layout.Auto, layout.Auto, layout.Auto
	}
}

// ParseEdgeValues parses a shorthand into EdgeValues.
func ParseEdgeValues(val string) layout.EdgeValues {
	t, r, b, l := ParseEdgeShorthand(val)
	return layout.EdgeValues{Top: t, Right: r, Bottom: b, Left: l}
}

// ParseColor parses a CSS color value.
// Supports: hex (#rgb, #rrggbb, #rrggbbaa), named colors, rgb(), rgba().
func ParseColor(val string) (uimath.Color, bool) {
	val = strings.TrimSpace(val)
	if val == "" {
		return uimath.Color{}, false
	}

	// Hex colors
	if strings.HasPrefix(val, "#") {
		return uimath.ColorHex(val), true
	}

	// rgb() / rgba()
	if strings.HasPrefix(val, "rgb") {
		return parseRGBFunc(val)
	}

	// Named colors
	if c, ok := namedColors[strings.ToLower(val)]; ok {
		return c, true
	}

	return uimath.Color{}, false
}

func parseRGBFunc(val string) (uimath.Color, bool) {
	// Extract content between ( and )
	start := strings.Index(val, "(")
	end := strings.LastIndex(val, ")")
	if start < 0 || end < 0 || end <= start {
		return uimath.Color{}, false
	}
	inner := val[start+1 : end]

	// Split by comma or space
	var parts []string
	if strings.Contains(inner, ",") {
		parts = strings.Split(inner, ",")
	} else {
		parts = strings.Fields(inner)
		// Remove "/" separator if present (e.g. "255 128 0 / 0.5")
		filtered := make([]string, 0, len(parts))
		for _, p := range parts {
			if p != "/" {
				filtered = append(filtered, p)
			}
		}
		parts = filtered
	}

	if len(parts) < 3 {
		return uimath.Color{}, false
	}

	r := parseColorComponent(strings.TrimSpace(parts[0]))
	g := parseColorComponent(strings.TrimSpace(parts[1]))
	b := parseColorComponent(strings.TrimSpace(parts[2]))
	a := float32(1.0)
	if len(parts) >= 4 {
		a = parseAlphaComponent(strings.TrimSpace(parts[3]))
	}

	return uimath.Color{R: r, G: g, B: b, A: a}, true
}

func parseColorComponent(s string) float32 {
	if strings.HasSuffix(s, "%") {
		return parseFloat32(strings.TrimSuffix(s, "%")) / 100
	}
	return parseFloat32(s) / 255
}

func parseAlphaComponent(s string) float32 {
	if strings.HasSuffix(s, "%") {
		return parseFloat32(strings.TrimSuffix(s, "%")) / 100
	}
	return parseFloat32(s)
}

// ParseDisplay parses a CSS display value.
func ParseDisplay(val string) (layout.Display, bool) {
	switch strings.TrimSpace(val) {
	case "block":
		return layout.DisplayBlock, true
	case "flex":
		return layout.DisplayFlex, true
	case "inline":
		return layout.DisplayInline, true
	case "none":
		return layout.DisplayNone, true
	case "grid":
		return layout.DisplayGrid, true
	default:
		return 0, false
	}
}

// ParsePosition parses a CSS position value.
func ParsePosition(val string) (layout.Position, bool) {
	switch strings.TrimSpace(val) {
	case "relative":
		return layout.PositionRelative, true
	case "absolute":
		return layout.PositionAbsolute, true
	case "fixed":
		return layout.PositionFixed, true
	default:
		return 0, false
	}
}

// ParseOverflow parses a CSS overflow value.
func ParseOverflow(val string) (layout.Overflow, bool) {
	switch strings.TrimSpace(val) {
	case "visible":
		return layout.OverflowVisible, true
	case "hidden":
		return layout.OverflowHidden, true
	case "scroll":
		return layout.OverflowScroll, true
	case "auto":
		return layout.OverflowAuto, true
	default:
		return 0, false
	}
}

// ParseFlexDirection parses a CSS flex-direction value.
func ParseFlexDirection(val string) (layout.FlexDirection, bool) {
	switch strings.TrimSpace(val) {
	case "row":
		return layout.FlexDirectionRow, true
	case "column":
		return layout.FlexDirectionColumn, true
	case "row-reverse":
		return layout.FlexDirectionRowReverse, true
	case "column-reverse":
		return layout.FlexDirectionColumnReverse, true
	default:
		return 0, false
	}
}

// ParseFlexWrap parses a CSS flex-wrap value.
func ParseFlexWrap(val string) (layout.FlexWrap, bool) {
	switch strings.TrimSpace(val) {
	case "nowrap":
		return layout.FlexWrapNoWrap, true
	case "wrap":
		return layout.FlexWrapWrap, true
	case "wrap-reverse":
		return layout.FlexWrapWrapReverse, true
	default:
		return 0, false
	}
}

// ParseJustifyContent parses a CSS justify-content value.
func ParseJustifyContent(val string) (layout.JustifyContent, bool) {
	switch strings.TrimSpace(val) {
	case "flex-start", "start":
		return layout.JustifyFlexStart, true
	case "flex-end", "end":
		return layout.JustifyFlexEnd, true
	case "center":
		return layout.JustifyCenter, true
	case "space-between":
		return layout.JustifySpaceBetween, true
	case "space-around":
		return layout.JustifySpaceAround, true
	case "space-evenly":
		return layout.JustifySpaceEvenly, true
	default:
		return 0, false
	}
}

// ParseAlignItems parses a CSS align-items value.
func ParseAlignItems(val string) (layout.AlignItems, bool) {
	switch strings.TrimSpace(val) {
	case "stretch":
		return layout.AlignStretch, true
	case "flex-start", "start":
		return layout.AlignFlexStart, true
	case "flex-end", "end":
		return layout.AlignFlexEnd, true
	case "center":
		return layout.AlignCenter, true
	case "baseline":
		return layout.AlignBaseline, true
	default:
		return 0, false
	}
}

// ParseAlignSelf parses a CSS align-self value.
func ParseAlignSelf(val string) (layout.AlignSelf, bool) {
	switch strings.TrimSpace(val) {
	case "auto":
		return layout.AlignSelfAuto, true
	case "stretch":
		return layout.AlignSelfStretch, true
	case "flex-start", "start":
		return layout.AlignSelfFlexStart, true
	case "flex-end", "end":
		return layout.AlignSelfFlexEnd, true
	case "center":
		return layout.AlignSelfCenter, true
	case "baseline":
		return layout.AlignSelfBaseline, true
	default:
		return 0, false
	}
}

// splitValues splits a CSS value by whitespace, respecting parentheses.
func splitValues(val string) []string {
	var parts []string
	var current strings.Builder
	depth := 0
	for i := 0; i < len(val); i++ {
		ch := val[i]
		if ch == '(' {
			depth++
			current.WriteByte(ch)
		} else if ch == ')' {
			depth--
			current.WriteByte(ch)
		} else if ch == ' ' && depth == 0 {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(ch)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

var namedColors = map[string]uimath.Color{
	"transparent": {},
	"white":       {R: 1, G: 1, B: 1, A: 1},
	"black":       {R: 0, G: 0, B: 0, A: 1},
	"red":         {R: 1, G: 0, B: 0, A: 1},
	"green":       {R: 0, G: 0.502, B: 0, A: 1}, // CSS green is #008000
	"blue":        {R: 0, G: 0, B: 1, A: 1},
	"yellow":      {R: 1, G: 1, B: 0, A: 1},
	"cyan":        {R: 0, G: 1, B: 1, A: 1},
	"magenta":     {R: 1, G: 0, B: 1, A: 1},
	"gray":        {R: 0.502, G: 0.502, B: 0.502, A: 1},
	"grey":        {R: 0.502, G: 0.502, B: 0.502, A: 1},
	"orange":      {R: 1, G: 0.647, B: 0, A: 1},
	"purple":      {R: 0.502, G: 0, B: 0.502, A: 1},
	"pink":        {R: 1, G: 0.753, B: 0.796, A: 1},
	"brown":       {R: 0.647, G: 0.165, B: 0.165, A: 1},
	"silver":      {R: 0.753, G: 0.753, B: 0.753, A: 1},
	"gold":        {R: 1, G: 0.843, B: 0, A: 1},
	"navy":        {R: 0, G: 0, B: 0.502, A: 1},
	"teal":        {R: 0, G: 0.502, B: 0.502, A: 1},
	"maroon":      {R: 0.502, G: 0, B: 0, A: 1},
	"olive":       {R: 0.502, G: 0.502, B: 0, A: 1},
	"aqua":        {R: 0, G: 1, B: 1, A: 1},
	"lime":        {R: 0, G: 1, B: 0, A: 1},
	"fuchsia":     {R: 1, G: 0, B: 1, A: 1},
}

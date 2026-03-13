package css

import (
	"math"
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

// ParseNonNegativeValue parses a CSS value and clamps negative values to 0.
// Used for padding and border-width per W3C CSS Box Model Level 3 §6.3/§6.4.
func ParseNonNegativeValue(val string) layout.Value {
	v := ParseValue(val)
	if (v.Unit == layout.UnitPx || v.Unit == layout.UnitPercent) && v.Amount < 0 {
		return layout.Zero
	}
	return v
}

// ParseNonNegativeEdgeValues parses a shorthand into EdgeValues, clamping negatives.
func ParseNonNegativeEdgeValues(val string) layout.EdgeValues {
	ev := ParseEdgeValues(val)
	clamp := func(v *layout.Value) {
		if (v.Unit == layout.UnitPx || v.Unit == layout.UnitPercent) && v.Amount < 0 {
			*v = layout.Zero
		}
	}
	clamp(&ev.Top)
	clamp(&ev.Right)
	clamp(&ev.Bottom)
	clamp(&ev.Left)
	return ev
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

// GradientStop represents a color stop in a gradient.
type GradientStop struct {
	Color    uimath.Color
	Position float32 // 0-1, -1 if not specified
}

// LinearGradient represents a parsed linear-gradient.
type LinearGradient struct {
	Angle float32 // radians, 0 = top to bottom
	Stops []GradientStop
}

// ParseLinearGradient parses a CSS linear-gradient() value.
func ParseLinearGradient(val string) (LinearGradient, bool) {
	val = strings.TrimSpace(val)
	if !strings.HasPrefix(val, "linear-gradient(") {
		return LinearGradient{}, false
	}
	// Extract content between ( and )
	start := strings.Index(val, "(")
	end := strings.LastIndex(val, ")")
	if start < 0 || end < 0 || end <= start {
		return LinearGradient{}, false
	}
	inner := strings.TrimSpace(val[start+1 : end])

	// Split by commas, respecting parentheses (for rgb()/rgba())
	args := splitGradientArgs(inner)
	if len(args) < 2 {
		return LinearGradient{}, false
	}

	grad := LinearGradient{
		Angle: math.Pi, // default: to bottom = 180deg
	}

	firstArg := strings.TrimSpace(args[0])
	stopStart := 0

	// Check if first argument is a direction or angle
	if strings.HasPrefix(firstArg, "to ") {
		grad.Angle = parseDirectionKeyword(firstArg)
		stopStart = 1
	} else if strings.HasSuffix(firstArg, "deg") {
		deg := parseFloat32(strings.TrimSuffix(firstArg, "deg"))
		grad.Angle = deg * math.Pi / 180
		stopStart = 1
	} else if strings.HasSuffix(firstArg, "rad") {
		grad.Angle = parseFloat32(strings.TrimSuffix(firstArg, "rad"))
		stopStart = 1
	}

	// Parse color stops
	for i := stopStart; i < len(args); i++ {
		stop := parseGradientStop(strings.TrimSpace(args[i]))
		if stop.Color.A == 0 && stop.Color.R == 0 && stop.Color.G == 0 && stop.Color.B == 0 && stop.Position < 0 {
			// Failed to parse — skip
			continue
		}
		grad.Stops = append(grad.Stops, stop)
	}

	if len(grad.Stops) < 2 {
		return LinearGradient{}, false
	}

	// Distribute positions for stops that don't have explicit positions
	distributeStopPositions(grad.Stops)

	return grad, true
}

// parseDirectionKeyword converts "to right", "to top left", etc. to radians.
func parseDirectionKeyword(dir string) float32 {
	dir = strings.TrimSpace(strings.TrimPrefix(dir, "to "))
	switch dir {
	case "top":
		return 0
	case "right":
		return math.Pi / 2
	case "bottom":
		return math.Pi
	case "left":
		return 3 * math.Pi / 2
	case "top right":
		return math.Pi / 4
	case "top left":
		return 7 * math.Pi / 4
	case "bottom right":
		return 3 * math.Pi / 4
	case "bottom left":
		return 5 * math.Pi / 4
	default:
		return math.Pi // default to bottom
	}
}

// parseGradientStop parses a single color stop like "red 50%" or "#ff0000".
func parseGradientStop(s string) GradientStop {
	stop := GradientStop{Position: -1}
	parts := splitValues(s)
	if len(parts) == 0 {
		return stop
	}

	// Last part might be a percentage position
	if len(parts) >= 2 {
		last := parts[len(parts)-1]
		if strings.HasSuffix(last, "%") {
			pct := parseFloat32(strings.TrimSuffix(last, "%"))
			stop.Position = pct / 100
			parts = parts[:len(parts)-1]
		}
	}

	// Rejoin remaining parts as the color value
	colorStr := strings.Join(parts, " ")
	if c, ok := ParseColor(colorStr); ok {
		stop.Color = c
	}
	return stop
}

// distributeStopPositions fills in -1 positions with evenly distributed values.
func distributeStopPositions(stops []GradientStop) {
	if len(stops) == 0 {
		return
	}
	// First and last default to 0 and 1
	if stops[0].Position < 0 {
		stops[0].Position = 0
	}
	if stops[len(stops)-1].Position < 0 {
		stops[len(stops)-1].Position = 1
	}
	// Fill gaps
	for i := 1; i < len(stops)-1; i++ {
		if stops[i].Position < 0 {
			// Find next stop with a defined position
			nextIdx := i + 1
			for nextIdx < len(stops) && stops[nextIdx].Position < 0 {
				nextIdx++
			}
			// Interpolate
			prevPos := stops[i-1].Position
			nextPos := stops[nextIdx].Position
			count := float32(nextIdx - i + 1)
			for j := i; j < nextIdx; j++ {
				t := float32(j-i+1) / count
				stops[j].Position = prevPos + t*(nextPos-prevPos)
			}
		}
	}
}

// splitGradientArgs splits gradient arguments by comma, respecting parentheses.
func splitGradientArgs(s string) []string {
	var parts []string
	var current strings.Builder
	depth := 0
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '(' {
			depth++
			current.WriteByte(ch)
		} else if ch == ')' {
			depth--
			current.WriteByte(ch)
		} else if ch == ',' && depth == 0 {
			parts = append(parts, current.String())
			current.Reset()
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

// BoxShadowLayer represents one layer of a CSS box-shadow value.
type BoxShadowLayer struct {
	OffsetX, OffsetY float32
	Blur, Spread     float32
	Color            uimath.Color
	Inset            bool
}

// splitShadowLayers splits a box-shadow value on commas that are not inside
// parentheses (to avoid splitting rgb(r,g,b) color values).
func splitShadowLayers(val string) []string {
	var layers []string
	depth := 0
	start := 0
	for i := 0; i < len(val); i++ {
		switch val[i] {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				layers = append(layers, strings.TrimSpace(val[start:i]))
				start = i + 1
			}
		}
	}
	if start < len(val) {
		layers = append(layers, strings.TrimSpace(val[start:]))
	}
	return layers
}

// ParseBoxShadow parses a CSS box-shadow value, returning one or more shadow layers.
// Each layer: [inset] offset-x offset-y [blur [spread]] [color]
// Multiple layers are separated by commas.
func ParseBoxShadow(val string) []BoxShadowLayer {
	defaultColor := uimath.Color{R: 0, G: 0, B: 0, A: 0.2}
	rawLayers := splitShadowLayers(val)
	var result []BoxShadowLayer
	for _, raw := range rawLayers {
		if raw == "" || raw == "none" {
			continue
		}
		layer := BoxShadowLayer{Color: defaultColor}
		parts := splitValues(raw)

		// Extract inset keyword (can appear anywhere among the tokens)
		var nums []string
		var colorParts []string
		inColorMode := false
		for _, p := range parts {
			if p == "inset" {
				layer.Inset = true
				continue
			}
			if inColorMode {
				colorParts = append(colorParts, p)
				continue
			}
			if isPlainNumber(p) || strings.HasSuffix(p, "px") || strings.HasSuffix(p, "em") || strings.HasSuffix(p, "rem") {
				nums = append(nums, p)
			} else {
				// Treat remainder as color
				inColorMode = true
				colorParts = append(colorParts, p)
			}
		}

		if len(nums) < 2 {
			continue // need at least offset-x and offset-y
		}
		layer.OffsetX = ParseFloat(nums[0])
		layer.OffsetY = ParseFloat(nums[1])
		if len(nums) >= 3 {
			layer.Blur = ParseFloat(nums[2])
		}
		if len(nums) >= 4 {
			layer.Spread = ParseFloat(nums[3])
		}

		if len(colorParts) > 0 {
			if c, ok := ParseColor(strings.Join(colorParts, " ")); ok {
				layer.Color = c
			}
		}

		result = append(result, layer)
	}
	return result
}

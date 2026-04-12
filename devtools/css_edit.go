package devtools

// CSS live-editing support.
//
// When the user modifies a CSS property in Chrome DevTools Styles pane,
// Chrome sends CSS.setPropertyText with the styleSheetId and updated text.
//
// Flow:
//   1. Chrome sends CSS.setPropertyText {styleSheetId: "inline-<nodeId>", range, text}
//   2. We parse the CSS declaration text → layout.Style field modification
//   3. We find the widget via widgetMap[ElementID(nodeId)]
//   4. We call widget.SetStyle(modifiedStyle)
//   5. We trigger markDirty → re-layout → new snapshot → DOM.documentUpdated

import (
	"strconv"
	"strings"

	"github.com/kasuganosora/ui/layout"
)

// --- StyleSheet ID parsing ---

// parseInlineStyleSheetID parses an inline styleSheetId of the form "inline-<nodeId>"
// and returns the numeric node ID. Returns (0, false) if the ID is not an inline ID
// or if the numeric suffix is missing/invalid.
func parseInlineStyleSheetID(id string) (int, bool) {
	if !strings.HasPrefix(id, "inline-") {
		return 0, false
	}
	n, err := strconv.Atoi(id[len("inline-"):])
	if err != nil {
		return 0, false
	}
	return n, true
}

// buildCSSStyle builds a CDP CSSStyle object for a widget's inline style.
// The styleSheetId is set to "inline-<nodeId>".
func buildCSSStyle(nodeID int, st layout.Style) map[string]any {
	declared := declaredStyleProps(st)
	propMaps := make([]map[string]any, len(declared))
	for i, p := range declared {
		text := p.Name + ": " + p.Value + ";"
		propMaps[i] = map[string]any{
			"name":      p.Name,
			"value":     p.Value,
			"important": false,
			"implicit":  false,
			"text":      text,
			"parsedOk":  true,
			"range": map[string]any{
				"startLine": i, "startColumn": 0,
				"endLine": i, "endColumn": len(text),
			},
		}
	}
	return map[string]any{
		"styleSheetId":     "inline-" + strconv.Itoa(nodeID),
		"cssProperties":    propMaps,
		"shorthandEntries": []any{},
	}
}

// --- CDP types for style editing ---

// cssRange is a CDP SourceRange (used in request params).

// --- Style property parsing (CSS text → layout.Style) ---

// parseCSSProperty parses a CSS declaration like "padding: 20px" and returns
// the property name and value, or empty strings if invalid.
func parseCSSProperty(text string) (name, value string) {
	text = strings.TrimSpace(text)
	text = strings.TrimSuffix(text, ";")
	text = strings.TrimSpace(text)

	colon := strings.IndexByte(text, ':')
	if colon <= 0 {
		return "", ""
	}
	name = strings.TrimSpace(text[:colon])
	value = strings.TrimSpace(text[colon+1:])
	return name, value
}

// applyCSSPropertyToStyle applies a single CSS property declaration to a layout.Style.
// Returns the modified style. If the property is unknown, the style is returned unchanged.
func applyCSSPropertyToStyle(st layout.Style, name, value string) layout.Style {
	prop := strings.ToLower(strings.TrimSpace(name))
	val := strings.TrimSpace(value)

	switch prop {
	// --- Display ---
	case "display":
		st.Display = parseDisplay(val)

	// --- Position ---
	case "position":
		st.Position = parsePosition(val)

	// --- Box Sizing ---
	case "box-sizing":
		st.BoxSizing = parseBoxSizing(val)

	// --- Overflow ---
	case "overflow":
		st.Overflow = parseOverflow(val)

	// --- Dimensions ---
	case "width":
		st.Width = parseValue(val)
	case "height":
		st.Height = parseValue(val)
	case "min-width":
		st.MinWidth = parseValue(val)
	case "min-height":
		st.MinHeight = parseValue(val)
	case "max-width":
		st.MaxWidth = parseValue(val)
	case "max-height":
		st.MaxHeight = parseValue(val)

	// --- Margin ---
	case "margin":
		st.Margin = parseEdges(val)
	case "margin-top":
		st.Margin.Top = parseValue(val)
	case "margin-right":
		st.Margin.Right = parseValue(val)
	case "margin-bottom":
		st.Margin.Bottom = parseValue(val)
	case "margin-left":
		st.Margin.Left = parseValue(val)

	// --- Padding ---
	case "padding":
		st.Padding = parseEdges(val)
	case "padding-top":
		st.Padding.Top = parseValue(val)
	case "padding-right":
		st.Padding.Right = parseValue(val)
	case "padding-bottom":
		st.Padding.Bottom = parseValue(val)
	case "padding-left":
		st.Padding.Left = parseValue(val)

	// --- Border ---
	case "border":
		v := parseValue(val)
		st.Border.Top = v
		st.Border.Right = v
		st.Border.Bottom = v
		st.Border.Left = v
	case "border-top":
		st.Border.Top = parseValue(val)
	case "border-right":
		st.Border.Right = parseValue(val)
	case "border-bottom":
		st.Border.Bottom = parseValue(val)
	case "border-left":
		st.Border.Left = parseValue(val)

	// --- Positioning offsets ---
	case "top":
		st.Top = parseValue(val)
	case "right":
		st.Right = parseValue(val)
	case "bottom":
		st.Bottom = parseValue(val)
	case "left":
		st.Left = parseValue(val)

	// --- Flexbox container ---
	case "flex-direction":
		st.FlexDirection = parseFlexDirection(val)
	case "flex-wrap":
		st.FlexWrap = parseFlexWrap(val)
	case "justify-content":
		st.JustifyContent = parseJustifyContent(val)
	case "align-items":
		st.AlignItems = parseAlignItems(val)
	case "align-content":
		st.AlignContent = parseAlignContent(val)
	case "gap":
		st.Gap = parsePx(val)
	case "row-gap":
		st.RowGap = parsePx(val)
	case "column-gap":
		st.ColumnGap = parsePx(val)

	// --- Flexbox item ---
	case "flex-grow":
		st.FlexGrow = parseFloat(val)
	case "flex-shrink":
		st.FlexShrink = parseFloat(val)
	case "flex-basis":
		st.FlexBasis = parseValue(val)
	case "align-self":
		st.AlignSelf = parseAlignSelf(val)
	case "order":
		st.Order = parseInt(val)

	// --- Typography ---
	case "font-size":
		st.FontSize = parsePx(val)

	// --- Text ---
	case "white-space":
		st.WhiteSpace = parseWhiteSpace(val)
	case "text-overflow":
		st.TextOverflow = parseTextOverflow(val)
	}
	return st
}

// resetCSSProperty resets a single CSS property in a layout.Style to its default value.
// Property names match CSS property names (e.g. "padding-top", "flex-direction").
func resetCSSProperty(st layout.Style, prop string) layout.Style {
	switch prop {
	// --- Display & Position ---
	case "display":
		st.Display = layout.DisplayBlock
	case "position":
		st.Position = layout.PositionRelative
	case "box-sizing":
		st.BoxSizing = layout.BoxSizingContentBox
	case "overflow":
		st.Overflow = layout.OverflowVisible

	// --- Dimensions ---
	case "width":
		st.Width = layout.Auto
	case "height":
		st.Height = layout.Auto
	case "min-width":
		st.MinWidth = layout.Auto
	case "min-height":
		st.MinHeight = layout.Auto
	case "max-width":
		st.MaxWidth = layout.Auto
	case "max-height":
		st.MaxHeight = layout.Auto

	// --- Margin (shorthand resets all four sides) ---
	case "margin":
		st.Margin.Top = layout.Value{}
		st.Margin.Right = layout.Value{}
		st.Margin.Bottom = layout.Value{}
		st.Margin.Left = layout.Value{}
	case "margin-top":
		st.Margin.Top = layout.Value{}
	case "margin-right":
		st.Margin.Right = layout.Value{}
	case "margin-bottom":
		st.Margin.Bottom = layout.Value{}
	case "margin-left":
		st.Margin.Left = layout.Value{}

	// --- Padding (shorthand resets all four sides) ---
	case "padding":
		st.Padding.Top = layout.Value{}
		st.Padding.Right = layout.Value{}
		st.Padding.Bottom = layout.Value{}
		st.Padding.Left = layout.Value{}
	case "padding-top":
		st.Padding.Top = layout.Value{}
	case "padding-right":
		st.Padding.Right = layout.Value{}
	case "padding-bottom":
		st.Padding.Bottom = layout.Value{}
	case "padding-left":
		st.Padding.Left = layout.Value{}

	// --- Border (shorthand resets all four sides) ---
	case "border":
		st.Border.Top = layout.Value{}
		st.Border.Right = layout.Value{}
		st.Border.Bottom = layout.Value{}
		st.Border.Left = layout.Value{}
	case "border-top":
		st.Border.Top = layout.Value{}
	case "border-right":
		st.Border.Right = layout.Value{}
	case "border-bottom":
		st.Border.Bottom = layout.Value{}
	case "border-left":
		st.Border.Left = layout.Value{}

	// --- Positioning offsets ---
	case "top":
		st.Top = layout.Auto
	case "right":
		st.Right = layout.Auto
	case "bottom":
		st.Bottom = layout.Auto
	case "left":
		st.Left = layout.Auto

	// --- Flexbox container ---
	case "flex-direction":
		st.FlexDirection = layout.FlexDirectionRow
	case "flex-wrap":
		st.FlexWrap = layout.FlexWrapNoWrap
	case "justify-content":
		st.JustifyContent = layout.JustifyFlexStart
	case "align-items":
		st.AlignItems = layout.AlignStretch
	case "align-content":
		st.AlignContent = layout.AlignContentStretch
	case "gap":
		st.Gap = 0
	case "row-gap":
		st.RowGap = 0
	case "column-gap":
		st.ColumnGap = 0

	// --- Flexbox item ---
	case "flex-grow":
		st.FlexGrow = 0
	case "flex-shrink":
		st.FlexShrink = 1
	case "flex-basis":
		st.FlexBasis = layout.Auto
	case "align-self":
		st.AlignSelf = layout.AlignSelfAuto
	case "order":
		st.Order = 0

	// --- Typography ---
	case "font-size":
		st.FontSize = 0

	// --- Text ---
	case "white-space":
		st.WhiteSpace = layout.WhiteSpaceNormal
	case "text-overflow":
		st.TextOverflow = layout.TextOverflowClip
	}
	return st
}

// propertyNameAtLine returns the CSS property name at the given line index
// in the declared style property list. Used by CSS.setPropertyText to identify
// which property is being deleted (Chrome sends empty text with a range).
func propertyNameAtLine(st layout.Style, line int) string {
	props := declaredStyleProps(st)
	if line < 0 || line >= len(props) {
		return ""
	}
	return props[line].Name
}

// --- CDP method handlers ---

func parseValue(s string) layout.Value {
	s = strings.TrimSpace(s)
	if s == "auto" || s == "" {
		return layout.Auto
	}
	if strings.HasSuffix(s, "%") {
		if n, err := strconv.ParseFloat(strings.TrimSuffix(s, "%"), 32); err == nil {
			return layout.Value{Amount: float32(n), Unit: layout.UnitPercent}
		}
	}
	if n, err := strconv.ParseFloat(strings.TrimSuffix(s, "px"), 32); err == nil {
		return layout.Value{Amount: float32(n), Unit: layout.UnitPx}
	}
	return layout.Auto
}

func parsePx(s string) float32 {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "px")
	n, _ := strconv.ParseFloat(s, 32)
	return float32(n)
}

func parseFloat(s string) float32 {
	s = strings.TrimSpace(s)
	n, _ := strconv.ParseFloat(s, 32)
	return float32(n)
}

func parseInt(s string) int {
	s = strings.TrimSpace(s)
	n, _ := strconv.Atoi(s)
	return n
}

func parseEdges(s string) layout.EdgeValues {
	parts := strings.Fields(s)
	var vals []layout.Value
	for _, p := range parts {
		vals = append(vals, parseValue(p))
	}
	switch len(vals) {
	case 1:
		return layout.EdgeValues{Top: vals[0], Right: vals[0], Bottom: vals[0], Left: vals[0]}
	case 2:
		return layout.EdgeValues{Top: vals[0], Right: vals[1], Bottom: vals[0], Left: vals[1]}
	case 3:
		return layout.EdgeValues{Top: vals[0], Right: vals[1], Bottom: vals[2], Left: vals[1]}
	case 4:
		return layout.EdgeValues{Top: vals[0], Right: vals[1], Bottom: vals[2], Left: vals[3]}
	default:
		return layout.EdgeValues{}
	}
}

// --- Enum parsers (CSS string → layout enum) ---

func parseDisplay(s string) layout.Display {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "none":
		return layout.DisplayNone
	case "block":
		return layout.DisplayBlock
	case "grid":
		return layout.DisplayGrid
	default:
		return layout.DisplayFlex
	}
}

func parsePosition(s string) layout.Position {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "absolute":
		return layout.PositionAbsolute
	case "fixed":
		return layout.PositionFixed
	default:
		return layout.PositionRelative
	}
}

func parseBoxSizing(s string) layout.BoxSizing {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "border-box":
		return layout.BoxSizingBorderBox
	default:
		return layout.BoxSizingContentBox
	}
}

func parseOverflow(s string) layout.Overflow {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "hidden":
		return layout.OverflowHidden
	case "scroll":
		return layout.OverflowScroll
	case "auto":
		return layout.OverflowAuto
	default:
		return layout.OverflowVisible
	}
}

func parseFlexDirection(s string) layout.FlexDirection {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "column":
		return layout.FlexDirectionColumn
	case "column-reverse":
		return layout.FlexDirectionColumnReverse
	case "row-reverse":
		return layout.FlexDirectionRowReverse
	default:
		return layout.FlexDirectionRow
	}
}

func parseFlexWrap(s string) layout.FlexWrap {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "wrap":
		return layout.FlexWrapWrap
	case "wrap-reverse":
		return layout.FlexWrapWrapReverse
	default:
		return layout.FlexWrapNoWrap
	}
}

func parseJustifyContent(s string) layout.JustifyContent {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "flex-end":
		return layout.JustifyFlexEnd
	case "center":
		return layout.JustifyCenter
	case "space-between":
		return layout.JustifySpaceBetween
	case "space-around":
		return layout.JustifySpaceAround
	case "space-evenly":
		return layout.JustifySpaceEvenly
	default:
		return layout.JustifyFlexStart
	}
}

func parseAlignItems(s string) layout.AlignItems {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "flex-start":
		return layout.AlignFlexStart
	case "flex-end":
		return layout.AlignFlexEnd
	case "center":
		return layout.AlignCenter
	case "baseline":
		return layout.AlignBaseline
	default:
		return layout.AlignStretch
	}
}

func parseAlignSelf(s string) layout.AlignSelf {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "flex-start":
		return layout.AlignSelfFlexStart
	case "flex-end":
		return layout.AlignSelfFlexEnd
	case "center":
		return layout.AlignSelfCenter
	case "baseline":
		return layout.AlignSelfBaseline
	case "stretch":
		return layout.AlignSelfStretch
	default:
		return layout.AlignSelfAuto
	}
}

func parseAlignContent(s string) layout.AlignContent {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "flex-start":
		return layout.AlignContentFlexStart
	case "flex-end":
		return layout.AlignContentFlexEnd
	case "center":
		return layout.AlignContentCenter
	case "space-between":
		return layout.AlignContentSpaceBetween
	case "space-around":
		return layout.AlignContentSpaceAround
	default:
		return layout.AlignContentStretch
	}
}

func parseWhiteSpace(s string) layout.WhiteSpace {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "nowrap":
		return layout.WhiteSpaceNowrap
	case "pre":
		return layout.WhiteSpacePre
	default:
		return layout.WhiteSpaceNormal
	}
}

func parseTextOverflow(s string) layout.TextOverflow {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "ellipsis":
		return layout.TextOverflowEllipsis
	default:
		return layout.TextOverflowClip
	}
}

package css

import (
	"sort"
	"strings"

	"github.com/kasuganosora/ui/layout"
)

// ComputedStyle holds all resolved CSS properties for an element.
type ComputedStyle struct {
	Layout layout.Style
	// Visual properties not in layout.Style
	Color           string // text color
	BackgroundColor string // background-color
	FontSize        string // font-size
	BorderRadius    string // border-radius
	BorderColor     string // border-color
	Opacity         string // opacity
	ZIndex          string // z-index
	// Raw map for any property not directly mapped
	Raw map[string]string
}

// matchedRule is a rule that matched an element, with computed specificity.
type matchedRule struct {
	specificity Specificity
	order       int
	decls       []Declaration
}

// ResolveStyle applies stylesheet rules to a single element and returns the computed style.
// inlineStyle is the element's style="" attribute (already parsed as declarations).
func ResolveStyle(sheet *Stylesheet, el *ElementInfo, ancestors []ElementInfo, inlineDecls []Declaration) ComputedStyle {
	var matched []matchedRule

	for _, rule := range sheet.Rules {
		for _, sel := range rule.Selectors {
			if matchSelector(&sel, el, ancestors) {
				spec := SelectorSpecificity(&sel)
				matched = append(matched, matchedRule{
					specificity: spec,
					order:       rule.Order,
					decls:       rule.Declarations,
				})
			}
		}
	}

	// Sort by specificity (lower first), then source order
	sort.SliceStable(matched, func(i, j int) bool {
		si, sj := matched[i].specificity, matched[j].specificity
		if si != sj {
			return si.Less(sj)
		}
		return matched[i].order < matched[j].order
	})

	// Build merged property map (later rules override earlier ones)
	props := make(map[string]string)
	importantProps := make(map[string]string)

	for _, mr := range matched {
		for _, decl := range mr.decls {
			val := ResolveVar(decl.Value, sheet.Variables)
			if decl.Important {
				importantProps[decl.Property] = val
			} else {
				props[decl.Property] = val
			}
		}
	}

	// Inline styles override normal rules (specificity 1,0,0,0)
	for _, decl := range inlineDecls {
		val := ResolveVar(decl.Value, sheet.Variables)
		if decl.Important {
			importantProps[decl.Property] = val
		} else {
			props[decl.Property] = val
		}
	}

	// !important overrides everything
	for k, v := range importantProps {
		props[k] = v
	}

	return buildComputedStyle(props)
}

func buildComputedStyle(props map[string]string) ComputedStyle {
	cs := ComputedStyle{
		Layout: layout.DefaultStyle(),
		Raw:    props,
	}

	for key, val := range props {
		applyProperty(&cs, key, val)
	}
	return cs
}

// applyProperty applies a single CSS property to a ComputedStyle.
func applyProperty(cs *ComputedStyle, prop, val string) {
	switch prop {
	// Display & positioning
	case "display":
		if d, ok := ParseDisplay(val); ok {
			cs.Layout.Display = d
		}
	case "position":
		if p, ok := ParsePosition(val); ok {
			cs.Layout.Position = p
		}
	case "overflow":
		if o, ok := ParseOverflow(val); ok {
			cs.Layout.Overflow = o
		}

	// Dimensions
	case "width":
		cs.Layout.Width = ParseValue(val)
	case "height":
		cs.Layout.Height = ParseValue(val)
	case "min-width":
		cs.Layout.MinWidth = ParseValue(val)
	case "min-height":
		cs.Layout.MinHeight = ParseValue(val)
	case "max-width":
		cs.Layout.MaxWidth = ParseValue(val)
	case "max-height":
		cs.Layout.MaxHeight = ParseValue(val)

	// Margin
	case "margin":
		cs.Layout.Margin = ParseEdgeValues(val)
	case "margin-top":
		cs.Layout.Margin.Top = ParseValue(val)
	case "margin-right":
		cs.Layout.Margin.Right = ParseValue(val)
	case "margin-bottom":
		cs.Layout.Margin.Bottom = ParseValue(val)
	case "margin-left":
		cs.Layout.Margin.Left = ParseValue(val)

	// Padding
	case "padding":
		cs.Layout.Padding = ParseEdgeValues(val)
	case "padding-top":
		cs.Layout.Padding.Top = ParseValue(val)
	case "padding-right":
		cs.Layout.Padding.Right = ParseValue(val)
	case "padding-bottom":
		cs.Layout.Padding.Bottom = ParseValue(val)
	case "padding-left":
		cs.Layout.Padding.Left = ParseValue(val)

	// Border (width)
	case "border-width":
		cs.Layout.Border = ParseEdgeValues(val)
	case "border-top-width":
		cs.Layout.Border.Top = ParseValue(val)
	case "border-right-width":
		cs.Layout.Border.Right = ParseValue(val)
	case "border-bottom-width":
		cs.Layout.Border.Bottom = ParseValue(val)
	case "border-left-width":
		cs.Layout.Border.Left = ParseValue(val)
	case "border":
		// Shorthand: width style color
		parseBorderShorthand(cs, val)

	// Positioning offsets
	case "top":
		cs.Layout.Top = ParseValue(val)
	case "right":
		cs.Layout.Right = ParseValue(val)
	case "bottom":
		cs.Layout.Bottom = ParseValue(val)
	case "left":
		cs.Layout.Left = ParseValue(val)

	// Flexbox container
	case "flex-direction":
		if d, ok := ParseFlexDirection(val); ok {
			cs.Layout.FlexDirection = d
		}
	case "flex-wrap":
		if w, ok := ParseFlexWrap(val); ok {
			cs.Layout.FlexWrap = w
		}
	case "justify-content":
		if j, ok := ParseJustifyContent(val); ok {
			cs.Layout.JustifyContent = j
		}
	case "align-items":
		if a, ok := ParseAlignItems(val); ok {
			cs.Layout.AlignItems = a
		}
	case "align-self":
		if a, ok := ParseAlignSelf(val); ok {
			cs.Layout.AlignSelf = a
		}
	case "gap":
		cs.Layout.Gap = ParseFloat(val)
	case "row-gap":
		cs.Layout.RowGap = ParseFloat(val)
	case "column-gap":
		cs.Layout.ColumnGap = ParseFloat(val)

	// Flexbox item
	case "flex-grow":
		cs.Layout.FlexGrow = ParseFloat(val)
	case "flex-shrink":
		cs.Layout.FlexShrink = ParseFloat(val)
	case "flex-basis":
		cs.Layout.FlexBasis = ParseValue(val)
	case "flex":
		parseFlexShorthand(cs, val)
	case "order":
		cs.Layout.Order = int(ParseFloat(val))

	// Visual properties (stored as strings for widget layer to interpret)
	case "color":
		cs.Color = val
	case "background-color", "background":
		cs.BackgroundColor = val
	case "font-size":
		cs.FontSize = val
	case "border-radius":
		cs.BorderRadius = val
	case "border-color":
		cs.BorderColor = val
	case "opacity":
		cs.Opacity = val
	case "z-index":
		cs.ZIndex = val
	}
}

func parseBorderShorthand(cs *ComputedStyle, val string) {
	parts := splitValues(val)
	for _, p := range parts {
		// Try as dimension (border-width)
		if p == "0" || strings.HasSuffix(p, "px") || strings.HasSuffix(p, "em") || isPlainNumber(p) {
			v := ParseValue(p)
			cs.Layout.Border = layout.EdgeValues{Top: v, Right: v, Bottom: v, Left: v}
		} else if p == "solid" || p == "dashed" || p == "dotted" || p == "none" {
			// border-style — store in raw
			cs.Raw["border-style"] = p
		} else {
			// Assume color
			cs.BorderColor = p
		}
	}
}

func isPlainNumber(s string) bool {
	if len(s) == 0 {
		return false
	}
	i := 0
	if s[i] == '-' {
		i++
	}
	if i >= len(s) {
		return false
	}
	hasDot := false
	hasDigit := false
	for i < len(s) {
		if s[i] >= '0' && s[i] <= '9' {
			hasDigit = true
		} else if s[i] == '.' && !hasDot {
			hasDot = true
		} else {
			return false
		}
		i++
	}
	return hasDigit
}

func parseFlexShorthand(cs *ComputedStyle, val string) {
	val = strings.TrimSpace(val)
	if val == "none" {
		cs.Layout.FlexGrow = 0
		cs.Layout.FlexShrink = 0
		cs.Layout.FlexBasis = layout.Auto
		return
	}
	if val == "auto" {
		cs.Layout.FlexGrow = 1
		cs.Layout.FlexShrink = 1
		cs.Layout.FlexBasis = layout.Auto
		return
	}

	parts := splitValues(val)
	switch len(parts) {
	case 1:
		// Single number = flex-grow
		cs.Layout.FlexGrow = ParseFloat(parts[0])
		cs.Layout.FlexShrink = 1
		cs.Layout.FlexBasis = layout.Zero
	case 2:
		cs.Layout.FlexGrow = ParseFloat(parts[0])
		cs.Layout.FlexShrink = ParseFloat(parts[1])
		cs.Layout.FlexBasis = layout.Zero
	case 3:
		cs.Layout.FlexGrow = ParseFloat(parts[0])
		cs.Layout.FlexShrink = ParseFloat(parts[1])
		cs.Layout.FlexBasis = ParseValue(parts[2])
	}
}

// matchSelector checks if a full selector matches an element given its ancestors.
func matchSelector(sel *Selector, el *ElementInfo, ancestors []ElementInfo) bool {
	if len(sel.Parts) == 0 {
		return false
	}

	// The rightmost compound must match the target element
	rightIdx := len(sel.Parts) - 1
	if sel.Parts[rightIdx].Compound == nil {
		return false
	}
	if !MatchCompound(sel.Parts[rightIdx].Compound, el) {
		return false
	}

	// Walk backwards through selector parts, matching against ancestors
	if rightIdx == 0 {
		return true
	}

	// Collect compound+combinator pairs going right to left
	ancIdx := 0 // index into ancestors (0 = parent, 1 = grandparent, ...)
	partIdx := rightIdx - 1

	for partIdx >= 0 {
		// partIdx should point to a combinator
		if sel.Parts[partIdx].Combinator == CombinatorNone {
			return false
		}
		comb := sel.Parts[partIdx].Combinator
		partIdx--
		if partIdx < 0 {
			return false
		}
		compound := sel.Parts[partIdx].Compound
		if compound == nil {
			return false
		}

		switch comb {
		case CombinatorChild:
			// Must match immediate parent
			if ancIdx >= len(ancestors) {
				return false
			}
			if !MatchCompound(compound, &ancestors[ancIdx]) {
				return false
			}
			ancIdx++

		case CombinatorDescendant:
			// Must match some ancestor
			found := false
			for ancIdx < len(ancestors) {
				if MatchCompound(compound, &ancestors[ancIdx]) {
					ancIdx++
					found = true
					break
				}
				ancIdx++
			}
			if !found {
				return false
			}

		case CombinatorAdjacent:
			// Previous sibling — not directly supported via ancestors
			// For now, skip (would need sibling info)
			return false

		case CombinatorSibling:
			// General sibling — not directly supported via ancestors
			return false
		}

		partIdx--
	}

	return true
}

// ParseInlineDeclarations parses inline style="..." into declarations.
func ParseInlineDeclarations(style string) []Declaration {
	if style == "" {
		return nil
	}
	var decls []Declaration
	for _, part := range strings.Split(style, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		idx := strings.Index(part, ":")
		if idx < 0 {
			continue
		}
		prop := strings.TrimSpace(strings.ToLower(part[:idx]))
		val := strings.TrimSpace(part[idx+1:])
		important := false
		if strings.HasSuffix(val, "!important") {
			val = strings.TrimSpace(strings.TrimSuffix(val, "!important"))
			important = true
		}
		if prop != "" && val != "" {
			decls = append(decls, Declaration{
				Property:  prop,
				Value:     val,
				Important: important,
			})
		}
	}
	return decls
}

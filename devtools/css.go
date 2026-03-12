package devtools

import (
	"encoding/json"
	"fmt"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
)

// cssProperty is a CDP CSS property name/value pair.
type cssProperty struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	Important bool   `json:"important"`
	Implicit  bool   `json:"implicit"`
	Text      string `json:"text"`
	ParsedOk  bool   `json:"parsedOk"`
}

func (s *Session) handleCSS(req Request) {
	switch req.Method {
	case "CSS.enable":
		s.cssEnabled = true
		s.sendResult(req.ID, map[string]any{})

	case "CSS.disable":
		s.cssEnabled = false
		s.sendResult(req.ID, map[string]any{})

	// getComputedStyleForNode returns post-layout computed values (actual pixels).
	case "CSS.getComputedStyleForNode":
		var p struct {
			NodeID int `json:"nodeId"`
		}
		_ = json.Unmarshal(req.Params, &p)

		snap := s.srv.getSnapshot()
		if snap == nil {
			s.sendResult(req.ID, map[string]any{"computedStyle": []cssProperty{}})
			return
		}
		node, ok := snap.Nodes[core.ElementID(p.NodeID)]
		if !ok {
			s.sendResult(req.ID, map[string]any{"computedStyle": []cssProperty{}})
			return
		}
		props := computedStyleProps(node.Style, node.Bounds)
		s.sendResult(req.ID, map[string]any{"computedStyle": props})

	// getMatchedStylesForNode returns the element's pre-layout CSS declarations.
	// We model the widget style as an "inline style" block.
	case "CSS.getMatchedStylesForNode":
		var p struct {
			NodeID int `json:"nodeId"`
		}
		_ = json.Unmarshal(req.Params, &p)

		snap := s.srv.getSnapshot()
		if snap == nil {
			s.sendResult(req.ID, emptyMatchedStyles())
			return
		}
		node, ok := snap.Nodes[core.ElementID(p.NodeID)]
		if !ok {
			s.sendResult(req.ID, emptyMatchedStyles())
			return
		}
		inline := inlineStyleBlock(node.Style)
		s.sendResult(req.ID, map[string]any{
			"inlineStyle":      inline,
			"attributesStyle":  nil,
			"matchedCSSRules":  []any{},
			"pseudoElements":   []any{},
			"inherited":        []any{},
			"cssKeyframesRules": []any{},
		})

	// getInlineStylesForNode — same as above for the inline style pane.
	case "CSS.getInlineStylesForNode":
		var p struct {
			NodeID int `json:"nodeId"`
		}
		_ = json.Unmarshal(req.Params, &p)

		snap := s.srv.getSnapshot()
		if snap == nil {
			s.sendResult(req.ID, map[string]any{"inlineStyle": nil, "attributesStyle": nil})
			return
		}
		node, ok := snap.Nodes[core.ElementID(p.NodeID)]
		if !ok {
			s.sendResult(req.ID, map[string]any{"inlineStyle": nil, "attributesStyle": nil})
			return
		}
		s.sendResult(req.ID, map[string]any{
			"inlineStyle":     inlineStyleBlock(node.Style),
			"attributesStyle": nil,
		})

	case "CSS.getBackgroundColors":
		// Not implemented — return empty.
		s.sendResult(req.ID, map[string]any{
			"backgroundColors":      []string{},
			"computedFontSize":      "",
			"computedFontWeight":    "",
			"computedBodyFontSize":  "",
		})

	case "CSS.getStyleSheetText":
		s.sendResult(req.ID, map[string]any{"text": ""})

	default:
		s.sendResult(req.ID, map[string]any{})
	}
}

// ----- style conversion -----

// computedStyleProps converts a layout.Style + actual bounds to CDP computed style.
// These are the values Chrome shows in the "Computed" tab.
func computedStyleProps(st layout.Style, bounds uimath.Rect) []cssProperty {
	var props []cssProperty
	add := func(name, value string) {
		props = append(props, cssProperty{Name: name, Value: value, ParsedOk: true})
	}

	add("display", displayCSS(st.Display))
	add("position", positionCSS(st.Position))

	// Actual rendered size (from layout result)
	add("width", fmt.Sprintf("%.2fpx", bounds.Width))
	add("height", fmt.Sprintf("%.2fpx", bounds.Height))
	add("left", fmt.Sprintf("%.2fpx", bounds.X))
	add("top", fmt.Sprintf("%.2fpx", bounds.Y))

	// Constraints (from style)
	if !st.MinWidth.IsAuto() {
		add("min-width", valueCSS(st.MinWidth))
	}
	if !st.MinHeight.IsAuto() {
		add("min-height", valueCSS(st.MinHeight))
	}
	if !st.MaxWidth.IsAuto() {
		add("max-width", valueCSS(st.MaxWidth))
	}
	if !st.MaxHeight.IsAuto() {
		add("max-height", valueCSS(st.MaxHeight))
	}

	// Spacing (st.Padding/Margin/Border are layout.EdgeValues with layout.Value fields)
	addEdgeV := func(prefix string, top, right, bottom, left layout.Value) {
		if top == right && right == bottom && bottom == left {
			if !top.IsAuto() && top.Amount != 0 {
				add(prefix, valueCSS(top))
			}
		} else {
			if !top.IsAuto() {
				add(prefix+"-top", valueCSS(top))
			}
			if !right.IsAuto() {
				add(prefix+"-right", valueCSS(right))
			}
			if !bottom.IsAuto() {
				add(prefix+"-bottom", valueCSS(bottom))
			}
			if !left.IsAuto() {
				add(prefix+"-left", valueCSS(left))
			}
		}
	}
	addEdgeV("padding", st.Padding.Top, st.Padding.Right, st.Padding.Bottom, st.Padding.Left)
	addEdgeV("margin", st.Margin.Top, st.Margin.Right, st.Margin.Bottom, st.Margin.Left)
	addEdgeV("border-width", st.Border.Top, st.Border.Right, st.Border.Bottom, st.Border.Left)

	// Flexbox container
	if st.Display == layout.DisplayFlex {
		add("flex-direction", flexDirCSS(st.FlexDirection))
		add("flex-wrap", flexWrapCSS(st.FlexWrap))
		add("justify-content", justifyCSS(st.JustifyContent))
		add("align-items", alignItemsCSS(st.AlignItems))
		if st.Gap != 0 {
			add("gap", fmt.Sprintf("%.2fpx", st.Gap))
		}
	}

	// Flexbox item
	if st.FlexGrow != 0 {
		add("flex-grow", fmt.Sprintf("%g", st.FlexGrow))
	}
	if st.FlexShrink != 0 {
		add("flex-shrink", fmt.Sprintf("%g", st.FlexShrink))
	}
	if !st.FlexBasis.IsAuto() {
		add("flex-basis", valueCSS(st.FlexBasis))
	}

	// Overflow
	if st.Overflow != layout.OverflowVisible {
		add("overflow", overflowCSS(st.Overflow))
	}

	// Typography
	if st.FontSize != 0 {
		add("font-size", fmt.Sprintf("%.2fpx", st.FontSize))
	}
	if st.WhiteSpace != layout.WhiteSpaceNormal {
		add("white-space", whiteSpaceCSS(st.WhiteSpace))
	}
	if st.TextOverflow == layout.TextOverflowEllipsis {
		add("text-overflow", "ellipsis")
	}

	// Positioning overrides
	if st.Position != layout.PositionRelative {
		if !st.Top.IsAuto() {
			add("top", valueCSS(st.Top))
		}
		if !st.Right.IsAuto() {
			add("right", valueCSS(st.Right))
		}
		if !st.Bottom.IsAuto() {
			add("bottom", valueCSS(st.Bottom))
		}
		if !st.Left.IsAuto() {
			add("left", valueCSS(st.Left))
		}
	}

	return props
}

// inlineStyleBlock returns a CDP CSSStyle object for the Styles panel.
// Shows pre-layout declared values (what the developer wrote, not computed pixels).
func inlineStyleBlock(st layout.Style) map[string]any {
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
		"styleSheetId":     "inline-0",
		"cssProperties":    propMaps,
		"shorthandEntries": []any{},
	}
}

// declaredStyleProps returns CSS declarations reflecting what was specified in code,
// using style values (not post-layout computed results).
func declaredStyleProps(st layout.Style) []cssProperty {
	var props []cssProperty
	add := func(name, value string) {
		props = append(props, cssProperty{Name: name, Value: value, ParsedOk: true})
	}

	add("display", displayCSS(st.Display))
	add("position", positionCSS(st.Position))

	if !st.Width.IsAuto() {
		add("width", valueCSS(st.Width))
	}
	if !st.Height.IsAuto() {
		add("height", valueCSS(st.Height))
	}
	if !st.MinWidth.IsAuto() {
		add("min-width", valueCSS(st.MinWidth))
	}
	if !st.MinHeight.IsAuto() {
		add("min-height", valueCSS(st.MinHeight))
	}
	if !st.MaxWidth.IsAuto() {
		add("max-width", valueCSS(st.MaxWidth))
	}
	if !st.MaxHeight.IsAuto() {
		add("max-height", valueCSS(st.MaxHeight))
	}

	addEdgesNonZero := func(prefix string, top, right, bottom, left layout.Value) {
		if top == right && right == bottom && bottom == left && !top.IsAuto() && top.Amount != 0 {
			add(prefix, valueCSS(top))
		} else {
			if !top.IsAuto() && top.Amount != 0 {
				add(prefix+"-top", valueCSS(top))
			}
			if !right.IsAuto() && right.Amount != 0 {
				add(prefix+"-right", valueCSS(right))
			}
			if !bottom.IsAuto() && bottom.Amount != 0 {
				add(prefix+"-bottom", valueCSS(bottom))
			}
			if !left.IsAuto() && left.Amount != 0 {
				add(prefix+"-left", valueCSS(left))
			}
		}
	}
	addEdgesNonZero("padding", st.Padding.Top, st.Padding.Right, st.Padding.Bottom, st.Padding.Left)
	addEdgesNonZero("margin", st.Margin.Top, st.Margin.Right, st.Margin.Bottom, st.Margin.Left)
	addEdgesNonZero("border-width", st.Border.Top, st.Border.Right, st.Border.Bottom, st.Border.Left)

	if st.Display == layout.DisplayFlex {
		add("flex-direction", flexDirCSS(st.FlexDirection))
		add("flex-wrap", flexWrapCSS(st.FlexWrap))
		add("justify-content", justifyCSS(st.JustifyContent))
		add("align-items", alignItemsCSS(st.AlignItems))
		if st.Gap != 0 {
			add("gap", fmt.Sprintf("%.2fpx", st.Gap))
		}
	}

	if st.FlexGrow != 0 {
		add("flex-grow", fmt.Sprintf("%g", st.FlexGrow))
	}
	if st.FlexShrink != 0 {
		add("flex-shrink", fmt.Sprintf("%g", st.FlexShrink))
	}
	if !st.FlexBasis.IsAuto() {
		add("flex-basis", valueCSS(st.FlexBasis))
	}

	if st.Overflow != layout.OverflowVisible {
		add("overflow", overflowCSS(st.Overflow))
	}
	if st.FontSize != 0 {
		add("font-size", fmt.Sprintf("%.2fpx", st.FontSize))
	}
	if st.WhiteSpace != layout.WhiteSpaceNormal {
		add("white-space", whiteSpaceCSS(st.WhiteSpace))
	}
	if st.TextOverflow == layout.TextOverflowEllipsis {
		add("text-overflow", "ellipsis")
	}

	return props
}

func emptyMatchedStyles() map[string]any {
	return map[string]any{
		"inlineStyle":       nil,
		"attributesStyle":   nil,
		"matchedCSSRules":   []any{},
		"pseudoElements":    []any{},
		"inherited":         []any{},
		"cssKeyframesRules": []any{},
	}
}

// ----- CSS value formatters -----

func valueCSS(v layout.Value) string {
	switch v.Unit {
	case layout.UnitPx:
		return fmt.Sprintf("%.2fpx", v.Amount)
	case layout.UnitPercent:
		return fmt.Sprintf("%g%%", v.Amount)
	default:
		return "auto"
	}
}

func displayCSS(d layout.Display) string {
	switch d {
	case layout.DisplayFlex:
		return "flex"
	case layout.DisplayGrid:
		return "grid"
	case layout.DisplayNone:
		return "none"
	case layout.DisplayInline:
		return "inline"
	default:
		return "block"
	}
}

func positionCSS(p layout.Position) string {
	switch p {
	case layout.PositionAbsolute:
		return "absolute"
	case layout.PositionFixed:
		return "fixed"
	default:
		return "relative"
	}
}

func flexDirCSS(d layout.FlexDirection) string {
	switch d {
	case layout.FlexDirectionColumn:
		return "column"
	case layout.FlexDirectionColumnReverse:
		return "column-reverse"
	case layout.FlexDirectionRowReverse:
		return "row-reverse"
	default:
		return "row"
	}
}

func flexWrapCSS(w layout.FlexWrap) string {
	switch w {
	case layout.FlexWrapWrap:
		return "wrap"
	case layout.FlexWrapWrapReverse:
		return "wrap-reverse"
	default:
		return "nowrap"
	}
}

func justifyCSS(j layout.JustifyContent) string {
	switch j {
	case layout.JustifyFlexEnd:
		return "flex-end"
	case layout.JustifyCenter:
		return "center"
	case layout.JustifySpaceBetween:
		return "space-between"
	case layout.JustifySpaceAround:
		return "space-around"
	case layout.JustifySpaceEvenly:
		return "space-evenly"
	default:
		return "flex-start"
	}
}

func alignItemsCSS(a layout.AlignItems) string {
	switch a {
	case layout.AlignFlexStart:
		return "flex-start"
	case layout.AlignFlexEnd:
		return "flex-end"
	case layout.AlignCenter:
		return "center"
	case layout.AlignBaseline:
		return "baseline"
	default:
		return "stretch"
	}
}

func overflowCSS(o layout.Overflow) string {
	switch o {
	case layout.OverflowHidden:
		return "hidden"
	case layout.OverflowScroll:
		return "scroll"
	case layout.OverflowAuto:
		return "auto"
	default:
		return "visible"
	}
}

func whiteSpaceCSS(w layout.WhiteSpace) string {
	switch w {
	case layout.WhiteSpaceNowrap:
		return "nowrap"
	case layout.WhiteSpacePre:
		return "pre"
	default:
		return "normal"
	}
}

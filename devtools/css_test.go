package devtools

import (
	"strings"
	"testing"

	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
)

// ─── CSS value formatters ─────────────────────────────────────────────────────

func TestValueCSS_Px(t *testing.T) {
	if got := valueCSS(pxVal(12.5)); got != "12.50px" {
		t.Errorf("want 12.50px, got %q", got)
	}
}

func TestValueCSS_Percent(t *testing.T) {
	if got := valueCSS(pctVal(50)); got != "50%" {
		t.Errorf("want 50%%, got %q", got)
	}
}

func TestValueCSS_Auto(t *testing.T) {
	auto := layout.Value{} // zero Unit = auto
	if got := valueCSS(auto); got != "auto" {
		t.Errorf("want auto, got %q", got)
	}
}

func TestDisplayCSS(t *testing.T) {
	cases := map[layout.Display]string{
		layout.DisplayFlex:   "flex",
		layout.DisplayGrid:   "grid",
		layout.DisplayNone:   "none",
		layout.DisplayInline: "inline",
		layout.DisplayBlock:  "block",
	}
	for d, want := range cases {
		if got := displayCSS(d); got != want {
			t.Errorf("display %v: want %q, got %q", d, want, got)
		}
	}
}

func TestPositionCSS(t *testing.T) {
	cases := map[layout.Position]string{
		layout.PositionAbsolute: "absolute",
		layout.PositionFixed:    "fixed",
		layout.PositionRelative: "relative",
	}
	for p, want := range cases {
		if got := positionCSS(p); got != want {
			t.Errorf("position %v: want %q, got %q", p, want, got)
		}
	}
}

func TestFlexDirCSS(t *testing.T) {
	cases := map[layout.FlexDirection]string{
		layout.FlexDirectionColumn:        "column",
		layout.FlexDirectionColumnReverse: "column-reverse",
		layout.FlexDirectionRowReverse:    "row-reverse",
		layout.FlexDirectionRow:           "row",
	}
	for d, want := range cases {
		if got := flexDirCSS(d); got != want {
			t.Errorf("dir %v: want %q, got %q", d, want, got)
		}
	}
}

func TestFlexWrapCSS(t *testing.T) {
	cases := map[layout.FlexWrap]string{
		layout.FlexWrapWrap:        "wrap",
		layout.FlexWrapWrapReverse: "wrap-reverse",
		layout.FlexWrapNoWrap:      "nowrap",
	}
	for w, want := range cases {
		if got := flexWrapCSS(w); got != want {
			t.Errorf("wrap %v: want %q, got %q", w, want, got)
		}
	}
}

func TestJustifyCSS(t *testing.T) {
	cases := map[layout.JustifyContent]string{
		layout.JustifyFlexEnd:      "flex-end",
		layout.JustifyCenter:       "center",
		layout.JustifySpaceBetween: "space-between",
		layout.JustifySpaceAround:  "space-around",
		layout.JustifySpaceEvenly:  "space-evenly",
		layout.JustifyFlexStart:    "flex-start",
	}
	for j, want := range cases {
		if got := justifyCSS(j); got != want {
			t.Errorf("justify %v: want %q, got %q", j, want, got)
		}
	}
}

func TestAlignItemsCSS(t *testing.T) {
	cases := map[layout.AlignItems]string{
		layout.AlignFlexStart: "flex-start",
		layout.AlignFlexEnd:   "flex-end",
		layout.AlignCenter:    "center",
		layout.AlignBaseline:  "baseline",
		layout.AlignStretch:   "stretch",
	}
	for a, want := range cases {
		if got := alignItemsCSS(a); got != want {
			t.Errorf("align %v: want %q, got %q", a, want, got)
		}
	}
}

func TestOverflowCSS(t *testing.T) {
	cases := map[layout.Overflow]string{
		layout.OverflowHidden:  "hidden",
		layout.OverflowScroll:  "scroll",
		layout.OverflowAuto:    "auto",
		layout.OverflowVisible: "visible",
	}
	for o, want := range cases {
		if got := overflowCSS(o); got != want {
			t.Errorf("overflow %v: want %q, got %q", o, want, got)
		}
	}
}

// ─── computedStyleProps ───────────────────────────────────────────────────────

func TestComputedStyleProps_Basic(t *testing.T) {
	st := layout.Style{
		Display:  layout.DisplayBlock,
		Position: layout.PositionRelative,
	}
	bounds := uimath.NewRect(10, 20, 100, 50)
	props := computedStyleProps(st, bounds)

	// Must always have display, position, width, height, left, top
	names := make(map[string]string)
	for _, p := range props {
		names[p.Name] = p.Value
	}
	if names["display"] != "block" {
		t.Errorf("display: want block, got %q", names["display"])
	}
	if names["width"] != "100.00px" {
		t.Errorf("width: want 100.00px, got %q", names["width"])
	}
	if names["height"] != "50.00px" {
		t.Errorf("height: want 50.00px, got %q", names["height"])
	}
	if names["left"] != "10.00px" {
		t.Errorf("left: want 10.00px, got %q", names["left"])
	}
	if names["top"] != "20.00px" {
		t.Errorf("top: want 20.00px, got %q", names["top"])
	}
}

func TestComputedStyleProps_FlexContainer(t *testing.T) {
	st := layout.Style{
		Display:        layout.DisplayFlex,
		FlexDirection:  layout.FlexDirectionColumn,
		FlexWrap:       layout.FlexWrapWrap,
		JustifyContent: layout.JustifyCenter,
		AlignItems:     layout.AlignFlexEnd,
		Gap:            8,
	}
	props := computedStyleProps(st, uimath.Rect{})
	names := make(map[string]string)
	for _, p := range props {
		names[p.Name] = p.Value
	}
	if names["flex-direction"] != "column" {
		t.Errorf("flex-direction: got %q", names["flex-direction"])
	}
	if names["flex-wrap"] != "wrap" {
		t.Errorf("flex-wrap: got %q", names["flex-wrap"])
	}
	if names["justify-content"] != "center" {
		t.Errorf("justify-content: got %q", names["justify-content"])
	}
	if names["align-items"] != "flex-end" {
		t.Errorf("align-items: got %q", names["align-items"])
	}
	if names["gap"] != "8.00px" {
		t.Errorf("gap: got %q", names["gap"])
	}
}

func TestComputedStyleProps_FlexItem(t *testing.T) {
	st := layout.Style{
		FlexGrow:   1,
		FlexShrink: 2,
		FlexBasis:  pxVal(100),
	}
	props := computedStyleProps(st, uimath.Rect{})
	names := make(map[string]string)
	for _, p := range props {
		names[p.Name] = p.Value
	}
	if names["flex-grow"] != "1" {
		t.Errorf("flex-grow: got %q", names["flex-grow"])
	}
	if names["flex-shrink"] != "2" {
		t.Errorf("flex-shrink: got %q", names["flex-shrink"])
	}
	if names["flex-basis"] != "100.00px" {
		t.Errorf("flex-basis: got %q", names["flex-basis"])
	}
}

func TestComputedStyleProps_UniformPadding(t *testing.T) {
	padVal := pxVal(10)
	st := layout.Style{
		Padding: layout.EdgeValues{Top: padVal, Right: padVal, Bottom: padVal, Left: padVal},
	}
	props := computedStyleProps(st, uimath.Rect{})
	names := make(map[string]string)
	for _, p := range props {
		names[p.Name] = p.Value
	}
	if names["padding"] != "10.00px" {
		t.Errorf("uniform padding: got %q", names["padding"])
	}
}

func TestComputedStyleProps_NonUniformPadding(t *testing.T) {
	st := layout.Style{
		Padding: layout.EdgeValues{
			Top:    pxVal(1),
			Right:  pxVal(2),
			Bottom: pxVal(3),
			Left:   pxVal(4),
		},
	}
	props := computedStyleProps(st, uimath.Rect{})
	names := make(map[string]string)
	for _, p := range props {
		names[p.Name] = p.Value
	}
	if names["padding-top"] != "1.00px" {
		t.Errorf("padding-top: got %q", names["padding-top"])
	}
	if names["padding-right"] != "2.00px" {
		t.Errorf("padding-right: got %q", names["padding-right"])
	}
}

func TestComputedStyleProps_Overflow(t *testing.T) {
	st := layout.Style{Overflow: layout.OverflowHidden}
	props := computedStyleProps(st, uimath.Rect{})
	names := make(map[string]string)
	for _, p := range props {
		names[p.Name] = p.Value
	}
	if names["overflow"] != "hidden" {
		t.Errorf("overflow: got %q", names["overflow"])
	}
}

func TestComputedStyleProps_Typography(t *testing.T) {
	st := layout.Style{
		FontSize:     14,
		WhiteSpace:   layout.WhiteSpaceNowrap,
		TextOverflow: layout.TextOverflowEllipsis,
	}
	props := computedStyleProps(st, uimath.Rect{})
	names := make(map[string]string)
	for _, p := range props {
		names[p.Name] = p.Value
	}
	if names["font-size"] != "14.00px" {
		t.Errorf("font-size: got %q", names["font-size"])
	}
	if names["text-overflow"] != "ellipsis" {
		t.Errorf("text-overflow: got %q", names["text-overflow"])
	}
}

func TestComputedStyleProps_AbsolutePosition(t *testing.T) {
	st := layout.Style{
		Position: layout.PositionAbsolute,
		Top:      pxVal(10),
		Left:     pxVal(20),
	}
	props := computedStyleProps(st, uimath.Rect{})
	names := make(map[string]string)
	for _, p := range props {
		names[p.Name] = p.Value
	}
	if names["position"] != "absolute" {
		t.Errorf("position: got %q", names["position"])
	}
	if names["top"] != "10.00px" {
		t.Errorf("top (absolute): got %q", names["top"])
	}
}

func TestComputedStyleProps_MinMaxConstraints(t *testing.T) {
	st := layout.Style{
		MinWidth:  pxVal(50),
		MaxWidth:  pxVal(200),
		MinHeight: pxVal(30),
		MaxHeight: pxVal(150),
	}
	props := computedStyleProps(st, uimath.Rect{})
	names := make(map[string]string)
	for _, p := range props {
		names[p.Name] = p.Value
	}
	if names["min-width"] != "50.00px" {
		t.Errorf("min-width: got %q", names["min-width"])
	}
	if names["max-width"] != "200.00px" {
		t.Errorf("max-width: got %q", names["max-width"])
	}
}

// ─── inlineStyleBlock ─────────────────────────────────────────────────────────

func TestInlineStyleBlock_Structure(t *testing.T) {
	st := layout.Style{
		Display: layout.DisplayFlex,
		Width:   pxVal(100),
	}
	block := inlineStyleBlock(st)

	if block["styleSheetId"] != "inline-0" {
		t.Errorf("styleSheetId: got %v", block["styleSheetId"])
	}
	props, ok := block["cssProperties"].([]map[string]any)
	if !ok || len(props) == 0 {
		t.Fatalf("cssProperties missing or empty")
	}
	// Each property should have name, value, text, range
	for _, p := range props {
		if p["name"] == nil || p["value"] == nil {
			t.Errorf("property missing name/value: %v", p)
		}
		if p["range"] == nil {
			t.Errorf("property missing range: %v", p)
		}
		if text, ok := p["text"].(string); !ok || !strings.Contains(text, ":") {
			t.Errorf("property text malformed: %v", p["text"])
		}
	}
}

// ─── declaredStyleProps ───────────────────────────────────────────────────────

func TestDeclaredStyleProps_NoDeclaredValues(t *testing.T) {
	// Zero style — only display+position should appear
	st := layout.Style{}
	props := declaredStyleProps(st)
	// Should at minimum have display and position
	if len(props) < 2 {
		t.Errorf("want at least 2 props (display+position), got %d", len(props))
	}
}

func TestDeclaredStyleProps_WithWidth(t *testing.T) {
	st := layout.Style{Width: pxVal(200)}
	props := declaredStyleProps(st)
	names := make(map[string]string)
	for _, p := range props {
		names[p.Name] = p.Value
	}
	if names["width"] != "200.00px" {
		t.Errorf("width: got %q", names["width"])
	}
}

func TestDeclaredStyleProps_ParsedOk(t *testing.T) {
	st := layout.Style{}
	for _, p := range declaredStyleProps(st) {
		if !p.ParsedOk {
			t.Errorf("parsedOk should be true for %q", p.Name)
		}
	}
}

// ─── emptyMatchedStyles ───────────────────────────────────────────────────────

func TestEmptyMatchedStyles(t *testing.T) {
	m := emptyMatchedStyles()
	if m["inlineStyle"] != nil {
		t.Error("inlineStyle should be nil")
	}
	if rules, ok := m["matchedCSSRules"].([]any); !ok || len(rules) != 0 {
		t.Error("matchedCSSRules should be empty slice")
	}
}

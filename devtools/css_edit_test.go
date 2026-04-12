package devtools

import (
	"testing"

	"github.com/kasuganosora/ui/layout"
)

// ─── CSS property parsing ───────────────────────────────────────────────

func TestParseCSSProperty(t *testing.T) {
	tests := []struct {
		input     string
		wantName  string
		wantValue string
	}{
		{"padding: 20px", "padding", "20px"},
		{"display: flex", "display", "flex"},
		{"  flex-direction : column  ", "flex-direction", "column"},
		{"padding: 20px;", "padding", "20px"},
		{"", "", ""},
		{"no-colon", "", ""},
		{":value", "", ""},
	}
	for _, tt := range tests {
		name, value := parseCSSProperty(tt.input)
		if name != tt.wantName || value != tt.wantValue {
			t.Errorf("parseCSSProperty(%q) = (%q, %q), want (%q, %q)",
				tt.input, name, value, tt.wantName, tt.wantValue)
		}
	}
}

// ─── Value parsing ──────────────────────────────────────────────────────

func TestParseValue(t *testing.T) {
	tests := []struct {
		input string
		want  layout.Value
	}{
		{"auto", layout.Auto},
		{"", layout.Auto},
		{"100px", layout.Value{Amount: 100, Unit: layout.UnitPx}},
		{"50%", layout.Value{Amount: 50, Unit: layout.UnitPercent}},
		{"10.5px", layout.Value{Amount: 10.5, Unit: layout.UnitPx}},
		{"0px", layout.Value{Amount: 0, Unit: layout.UnitPx}},
	}
	for _, tt := range tests {
		got := parseValue(tt.input)
		if got != tt.want {
			t.Errorf("parseValue(%q) = %+v, want %+v", tt.input, got, tt.want)
		}
	}
}

func TestParsePx(t *testing.T) {
	tests := []struct {
		input string
		want  float32
	}{
		{"20px", 20},
		{"20", 20},
		{"10.5px", 10.5},
		{"0px", 0},
		{"", 0},
	}
	for _, tt := range tests {
		got := parsePx(tt.input)
		if got != tt.want {
			t.Errorf("parsePx(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParseEdges(t *testing.T) {
	tests := []struct {
		input string
		want  layout.EdgeValues
	}{
		{"10px", layout.EdgeValues{Top: layout.Px(10), Right: layout.Px(10), Bottom: layout.Px(10), Left: layout.Px(10)}},
		{"10px 20px", layout.EdgeValues{Top: layout.Px(10), Right: layout.Px(20), Bottom: layout.Px(10), Left: layout.Px(20)}},
		{"10px 20px 30px", layout.EdgeValues{Top: layout.Px(10), Right: layout.Px(20), Bottom: layout.Px(30), Left: layout.Px(20)}},
		{"10px 20px 30px 40px", layout.EdgeValues{Top: layout.Px(10), Right: layout.Px(20), Bottom: layout.Px(30), Left: layout.Px(40)}},
		{"", layout.EdgeValues{}},
	}
	for _, tt := range tests {
		got := parseEdges(tt.input)
		if got != tt.want {
			t.Errorf("parseEdges(%q) = %+v, want %+v", tt.input, got, tt.want)
		}
	}
}

// ─── Enum parsing ───────────────────────────────────────────────────────

func TestParseDisplay(t *testing.T) {
	if got := parseDisplay("flex"); got != layout.DisplayFlex {
		t.Errorf("parseDisplay(flex) = %v, want Flex", got)
	}
	if got := parseDisplay("none"); got != layout.DisplayNone {
		t.Errorf("parseDisplay(none) = %v, want None", got)
	}
	if got := parseDisplay("unknown"); got != layout.DisplayFlex {
		t.Errorf("parseDisplay(unknown) = %v, want Flex (default)", got)
	}
}

func TestParseFlexDirection(t *testing.T) {
	tests := []struct {
		input string
		want  layout.FlexDirection
	}{
		{"row", layout.FlexDirectionRow},
		{"column", layout.FlexDirectionColumn},
		{"column-reverse", layout.FlexDirectionColumnReverse},
		{"row-reverse", layout.FlexDirectionRowReverse},
	}
	for _, tt := range tests {
		if got := parseFlexDirection(tt.input); got != tt.want {
			t.Errorf("parseFlexDirection(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParseJustifyContent(t *testing.T) {
	tests := []struct {
		input string
		want  layout.JustifyContent
	}{
		{"flex-start", layout.JustifyFlexStart},
		{"flex-end", layout.JustifyFlexEnd},
		{"center", layout.JustifyCenter},
		{"space-between", layout.JustifySpaceBetween},
		{"space-around", layout.JustifySpaceAround},
		{"space-evenly", layout.JustifySpaceEvenly},
	}
	for _, tt := range tests {
		if got := parseJustifyContent(tt.input); got != tt.want {
			t.Errorf("parseJustifyContent(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParseAlignItems(t *testing.T) {
	tests := []struct {
		input string
		want  layout.AlignItems
	}{
		{"stretch", layout.AlignStretch},
		{"flex-start", layout.AlignFlexStart},
		{"flex-end", layout.AlignFlexEnd},
		{"center", layout.AlignCenter},
		{"baseline", layout.AlignBaseline},
	}
	for _, tt := range tests {
		if got := parseAlignItems(tt.input); got != tt.want {
			t.Errorf("parseAlignItems(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParseOverflow(t *testing.T) {
	tests := []struct {
		input string
		want  layout.Overflow
	}{
		{"visible", layout.OverflowVisible},
		{"hidden", layout.OverflowHidden},
		{"scroll", layout.OverflowScroll},
		{"auto", layout.OverflowAuto},
	}
	for _, tt := range tests {
		if got := parseOverflow(tt.input); got != tt.want {
			t.Errorf("parseOverflow(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

// ─── Apply CSS property to style ────────────────────────────────────────

func TestApplyCSSPropertyToStyle_Dimensions(t *testing.T) {
	st := layout.DefaultStyle()
	st = applyCSSPropertyToStyle(st, "width", "200px")
	if st.Width != layout.Px(200) {
		t.Errorf("width = %+v, want 200px", st.Width)
	}
	st = applyCSSPropertyToStyle(st, "height", "50%")
	if st.Height != layout.Pct(50) {
		t.Errorf("height = %+v, want 50%%", st.Height)
	}
}

func TestApplyCSSPropertyToStyle_Padding(t *testing.T) {
	st := layout.DefaultStyle()
	st = applyCSSPropertyToStyle(st, "padding", "10px 20px")
	want := layout.EdgeValues{Top: layout.Px(10), Right: layout.Px(20), Bottom: layout.Px(10), Left: layout.Px(20)}
	if st.Padding != want {
		t.Errorf("padding = %+v, want %+v", st.Padding, want)
	}
	// Individual sides
	st = applyCSSPropertyToStyle(st, "padding-top", "30px")
	if st.Padding.Top != layout.Px(30) {
		t.Errorf("padding-top = %+v, want 30px", st.Padding.Top)
	}
}

func TestApplyCSSPropertyToStyle_Flexbox(t *testing.T) {
	st := layout.DefaultStyle()
	st = applyCSSPropertyToStyle(st, "flex-direction", "column")
	if st.FlexDirection != layout.FlexDirectionColumn {
		t.Errorf("flex-direction = %v, want Column", st.FlexDirection)
	}
	st = applyCSSPropertyToStyle(st, "justify-content", "center")
	if st.JustifyContent != layout.JustifyCenter {
		t.Errorf("justify-content = %v, want Center", st.JustifyContent)
	}
	st = applyCSSPropertyToStyle(st, "align-items", "center")
	if st.AlignItems != layout.AlignCenter {
		t.Errorf("align-items = %v, want Center", st.AlignItems)
	}
	st = applyCSSPropertyToStyle(st, "flex-grow", "1.5")
	if st.FlexGrow != 1.5 {
		t.Errorf("flex-grow = %v, want 1.5", st.FlexGrow)
	}
}

func TestApplyCSSPropertyToStyle_UnknownProp(t *testing.T) {
	st := layout.DefaultStyle()
	st2 := applyCSSPropertyToStyle(st, "unknown-property", "some-value")
	// Verify a known property is unchanged
	if st2.Display != st.Display {
		t.Error("unknown property should not modify style")
	}
	if st2.Width != st.Width {
		t.Error("unknown property should not modify style")
	}
}

// ─── Declared style props ───────────────────────────────────────────────

func TestDeclaredStylePropsRaw(t *testing.T) {
	st := layout.DefaultStyle()
	props := declaredStyleProps(st)
	if len(props) == 0 {
		t.Error("default style should have at least display and position")
	}

	// Check first property is display
	if props[0].Name != "display" {
		t.Errorf("first property = %q, want display", props[0].Name)
	}
}

func TestDeclaredStylePropsRaw_WithValues(t *testing.T) {
	st := layout.DefaultStyle()
	st.Padding.Top = layout.Px(10)
	st.FlexDirection = layout.FlexDirectionColumn
	props := declaredStyleProps(st)

	var hasPaddingTop, hasFlexDir bool
	for _, p := range props {
		if p.Name == "padding-top" {
			hasPaddingTop = true
		}
		if p.Name == "flex-direction" {
			hasFlexDir = true
		}
	}
	if !hasPaddingTop {
		t.Error("should have padding-top when set")
	}
	if !hasFlexDir {
		t.Error("should have flex-direction when non-default")
	}
}

// ─── StyleSheet ID parsing ──────────────────────────────────────────────

func TestParseInlineStyleSheetID(t *testing.T) {
	tests := []struct {
		input  string
		wantID int
		wantOK bool
	}{
		{"inline-42", 42, true},
		{"inline-0", 0, true},
		{"inline-1", 1, true},
		{"global-42", 0, false},
		{"inline-", 0, false},
		{"inline-abc", 0, false},
		{"", 0, false},
	}
	for _, tt := range tests {
		id, ok := parseInlineStyleSheetID(tt.input)
		if id != tt.wantID || ok != tt.wantOK {
			t.Errorf("parseInlineStyleSheetID(%q) = (%d, %v), want (%d, %v)",
				tt.input, id, ok, tt.wantID, tt.wantOK)
		}
	}
}

// ─── Build CSS Style response ───────────────────────────────────────────

func TestBuildCSSStyle(t *testing.T) {
	st := layout.DefaultStyle()
	st.Width = layout.Px(100)
	result := buildCSSStyle(7, st)

	if result["styleSheetId"] != "inline-7" {
		t.Errorf("styleSheetId = %v, want inline-7", result["styleSheetId"])
	}
	props, ok := result["cssProperties"]
	if !ok {
		t.Fatal("missing cssProperties")
	}
	propSlice := props.([]map[string]any)
	if len(propSlice) == 0 {
		t.Error("should have properties")
	}
	// Check that each property has required fields
	for _, p := range propSlice {
		if _, ok := p["name"]; !ok {
			t.Error("property missing name")
		}
		if _, ok := p["value"]; !ok {
			t.Error("property missing value")
		}
		if _, ok := p["text"]; !ok {
			t.Error("property missing text")
		}
		if _, ok := p["range"]; !ok {
			t.Error("property missing range")
		}
	}
}

// ─── Property reset ─────────────────────────────────────────────────────

func TestResetCSSProperty_Width(t *testing.T) {
	st := layout.DefaultStyle()
	st.Width = layout.Px(200)
	st = resetCSSProperty(st, "width")
	if !st.Width.IsAuto() {
		t.Errorf("width after reset = %+v, want auto", st.Width)
	}
}

func TestResetCSSProperty_Overflow(t *testing.T) {
	st := layout.DefaultStyle()
	st.Overflow = layout.OverflowHidden
	st = resetCSSProperty(st, "overflow")
	if st.Overflow != layout.OverflowVisible {
		t.Errorf("overflow after reset = %v, want Visible", st.Overflow)
	}
}

func TestResetCSSProperty_FlexShrink(t *testing.T) {
	st := layout.DefaultStyle()
	st.FlexShrink = 0
	st = resetCSSProperty(st, "flex-shrink")
	if st.FlexShrink != 1 {
		t.Errorf("flex-shrink after reset = %v, want 1", st.FlexShrink)
	}
}

func TestResetCSSProperty_PaddingTop(t *testing.T) {
	st := layout.DefaultStyle()
	st.Padding.Top = layout.Px(10)
	st = resetCSSProperty(st, "padding-top")
	if st.Padding.Top != (layout.Value{}) {
		t.Errorf("padding-top after reset = %+v, want zero value", st.Padding.Top)
	}
}

func TestResetCSSProperty_FontSize(t *testing.T) {
	st := layout.DefaultStyle()
	st.FontSize = 16
	st = resetCSSProperty(st, "font-size")
	if st.FontSize != 0 {
		t.Errorf("font-size after reset = %v, want 0", st.FontSize)
	}
}

func TestResetCSSProperty_Display(t *testing.T) {
	st := layout.DefaultStyle()
	st.Display = layout.DisplayNone
	st = resetCSSProperty(st, "display")
	if st.Display != layout.DisplayBlock {
		t.Errorf("display after reset = %v, want Block", st.Display)
	}
}

func TestResetCSSProperty_UnknownProp(t *testing.T) {
	st := layout.DefaultStyle()
	st.Width = layout.Px(100)
	st2 := resetCSSProperty(st, "nonexistent-property")
	if st2.Width != layout.Px(100) {
		t.Error("unknown property should not modify style")
	}
}

func TestResetCSSProperty_AlignSelf(t *testing.T) {
	st := layout.DefaultStyle()
	st.AlignSelf = layout.AlignSelfCenter
	st = resetCSSProperty(st, "align-self")
	if st.AlignSelf != layout.AlignSelfAuto {
		t.Errorf("align-self after reset = %v, want Auto", st.AlignSelf)
	}
}

func TestResetCSSProperty_Gap(t *testing.T) {
	st := layout.DefaultStyle()
	st.Gap = 10
	st = resetCSSProperty(st, "gap")
	if st.Gap != 0 {
		t.Errorf("gap after reset = %v, want 0", st.Gap)
	}
}

// ─── propertyNameAtLine ────────────────────────────────────────────────

func TestPropertyNameAtLine(t *testing.T) {
	st := layout.DefaultStyle()

	// Line 0 should be "display" (first declared property)
	name := propertyNameAtLine(st, 0)
	if name != "display" {
		t.Errorf("line 0 = %q, want display", name)
	}

	// Line 1 should be "position"
	name = propertyNameAtLine(st, 1)
	if name != "position" {
		t.Errorf("line 1 = %q, want position", name)
	}

	// Out of bounds should return empty
	name = propertyNameAtLine(st, -1)
	if name != "" {
		t.Errorf("line -1 = %q, want empty", name)
	}
	name = propertyNameAtLine(st, 9999)
	if name != "" {
		t.Errorf("line 9999 = %q, want empty", name)
	}
}

func TestPropertyNameAtLine_WithNonDefault(t *testing.T) {
	st := layout.DefaultStyle()
	st.Width = layout.Px(100)
	st.FlexDirection = layout.FlexDirectionColumn

	props := declaredStyleProps(st)
	// Find the line index of flex-direction
	flexDirLine := -1
	for i, p := range props {
		if p.Name == "flex-direction" {
			flexDirLine = i
			break
		}
	}
	if flexDirLine < 0 {
		t.Fatal("flex-direction not found in declared props")
	}
	name := propertyNameAtLine(st, flexDirLine)
	if name != "flex-direction" {
		t.Errorf("at flex-direction line %d, got %q", flexDirLine, name)
	}
}

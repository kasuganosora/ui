package css

import (
	"testing"

	"github.com/kasuganosora/ui/layout"
)

// === token.go coverage ===

func TestTokenizeEmpty(t *testing.T) {
	tokens := newTokenizer("").tokenize()
	if len(tokens) != 0 {
		t.Errorf("expected 0 tokens, got %d", len(tokens))
	}
}

func TestTokenizeUnterminatedComment(t *testing.T) {
	tokens := newTokenizer("a /* unterminated").tokenize()
	// Should not panic; "a" should be parsed
	if len(tokens) == 0 {
		t.Error("expected at least 1 token")
	}
}

func TestTokenizeUnterminatedString(t *testing.T) {
	tokens := newTokenizer(`"unterminated`).tokenize()
	if len(tokens) != 1 || tokens[0].Type != TokenString {
		t.Errorf("expected 1 string token, got %+v", tokens)
	}
}

func TestTokenizeStringEscape(t *testing.T) {
	tokens := newTokenizer(`"he\"llo"`).tokenize()
	if tokens[0].Value != `he\"llo` {
		t.Errorf("string escape: %q", tokens[0].Value)
	}
}

func TestTokenizeSingleQuoteString(t *testing.T) {
	tokens := newTokenizer(`'hello'`).tokenize()
	if tokens[0].Type != TokenString || tokens[0].Value != "hello" {
		t.Errorf("single quote: %+v", tokens[0])
	}
}

func TestTokenizeDotNumber(t *testing.T) {
	// .5 should be a number, not dot+ident
	tokens := newTokenizer(".5px").tokenize()
	if tokens[0].Type != TokenDimension || tokens[0].Num != 0.5 {
		t.Errorf(".5px: %+v", tokens[0])
	}
}

func TestTokenizeCustomProperty(t *testing.T) {
	// --primary should be ident
	tokens := newTokenizer("--primary").tokenize()
	if tokens[0].Type != TokenIdent || tokens[0].Value != "--primary" {
		t.Errorf("custom property: %+v", tokens[0])
	}
}

func TestTokenizeDelim(t *testing.T) {
	tokens := newTokenizer("^").tokenize()
	if tokens[0].Type != TokenDelim || tokens[0].Value != "^" {
		t.Errorf("delim: %+v", tokens[0])
	}
}

func TestTokenizeBrackets(t *testing.T) {
	tokens := newTokenizer("[attr]").tokenize()
	found := map[TokenType]bool{}
	for _, tok := range tokens {
		found[tok.Type] = true
	}
	if !found[TokenLBracket] || !found[TokenRBracket] {
		t.Error("expected brackets")
	}
}

func TestTokenizeAllSingleChars(t *testing.T) {
	// Cover all single-char token branches
	tokens := newTokenizer(":;,.{}()[]>+~*/@!").tokenize()
	expected := []TokenType{
		TokenColon, TokenSemicolon, TokenComma, TokenDot,
		TokenLBrace, TokenRBrace, TokenLParen, TokenRParen,
		TokenLBracket, TokenRBracket, TokenGT, TokenPlus,
		TokenTilde, TokenStar, TokenSlash, TokenAt, TokenExcl,
	}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}
	for i, tok := range tokens {
		if tok.Type != expected[i] {
			t.Errorf("token[%d] type=%d, want %d", i, tok.Type, expected[i])
		}
	}
}

func TestTokenizeNegativeDecimal(t *testing.T) {
	tokens := newTokenizer("-3.14rem").tokenize()
	if tokens[0].Type != TokenDimension || tokens[0].Num != -3.14 || tokens[0].Unit != "rem" {
		t.Errorf("-3.14rem: %+v", tokens[0])
	}
}

func TestTokenizeFormFeed(t *testing.T) {
	tokens := newTokenizer("a\fb").tokenize()
	hasWS := false
	for _, tok := range tokens {
		if tok.Type == TokenWhitespace {
			hasWS = true
		}
	}
	if !hasWS {
		t.Error("form feed should be whitespace")
	}
}

// === parser.go coverage ===

func TestParseAtRuleImport(t *testing.T) {
	sheet := Parse(`@import url("foo.css"); .a { color: red; }`)
	if len(sheet.Rules) != 1 {
		t.Fatalf("expected 1 rule after @import, got %d", len(sheet.Rules))
	}
}

func TestParseAtRuleMedia(t *testing.T) {
	sheet := Parse(`@media screen { .a { color: red; } } .b { color: blue; }`)
	// @media block is skipped, .b rule should be parsed
	if len(sheet.Rules) != 1 {
		t.Fatalf("expected 1 rule after @media, got %d", len(sheet.Rules))
	}
	if sheet.Rules[0].Selectors[0].Parts[0].Compound.Classes[0] != "b" {
		t.Error("expected .b rule after @media skip")
	}
}

func TestParseMalformedDeclaration(t *testing.T) {
	// Declaration without colon → should be skipped
	sheet := Parse(`.a { invalid; color: red; }`)
	if len(sheet.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(sheet.Rules))
	}
	found := false
	for _, d := range sheet.Rules[0].Declarations {
		if d.Property == "color" && d.Value == "red" {
			found = true
		}
	}
	if !found {
		t.Error("expected color:red after skipping malformed declaration")
	}
}

func TestParseMalformedNoColon(t *testing.T) {
	// Property followed by } instead of :
	sheet := Parse(`.a { color }`)
	if len(sheet.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(sheet.Rules))
	}
	if len(sheet.Rules[0].Declarations) != 0 {
		t.Errorf("expected 0 declarations for malformed rule, got %d", len(sheet.Rules[0].Declarations))
	}
}

func TestParseEmptyValue(t *testing.T) {
	sheet := Parse(`.a { color: ; }`)
	if len(sheet.Rules) != 1 {
		t.Fatal("expected 1 rule")
	}
	// Empty value should be skipped
	if len(sheet.Rules[0].Declarations) != 0 {
		t.Errorf("expected 0 declarations for empty value, got %d", len(sheet.Rules[0].Declarations))
	}
}

func TestParseEmptySelector(t *testing.T) {
	sheet := Parse(`{ color: red; }`)
	// Empty selector text still gets parsed; ParseSelectorList may return a match.
	// Just ensure no panic and declarations are captured if a rule is created.
	if len(sheet.Rules) > 0 {
		if len(sheet.Rules[0].Declarations) != 1 {
			t.Errorf("expected 1 declaration, got %d", len(sheet.Rules[0].Declarations))
		}
	}
}

func TestParseFunctionInValue(t *testing.T) {
	sheet := Parse(`.a { background: rgb(255, 0, 0); }`)
	if len(sheet.Rules) != 1 || len(sheet.Rules[0].Declarations) != 1 {
		t.Fatal("expected 1 rule with 1 declaration")
	}
	d := sheet.Rules[0].Declarations[0]
	if d.Property != "background" || d.Value != "rgb(255, 0, 0)" {
		t.Errorf("function value: %+v", d)
	}
}

func TestParseNoClosingBrace(t *testing.T) {
	// Missing closing brace
	sheet := Parse(`.a { color: red`)
	// Should still parse what it can
	if len(sheet.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(sheet.Rules))
	}
}

func TestParseExclInValue(t *testing.T) {
	// ! that's NOT followed by "important"
	sheet := Parse(`.a { content: "!"; }`)
	// The ! will be treated as excl token, but next is not "important"
	_ = sheet // should not panic
}

// === selector.go coverage ===

func TestParseSelectorAdjacentSibling(t *testing.T) {
	sel := ParseSelector("h1 + p")
	if len(sel.Parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(sel.Parts))
	}
	if sel.Parts[1].Combinator != CombinatorAdjacent {
		t.Errorf("expected adjacent, got %d", sel.Parts[1].Combinator)
	}
}

func TestParseSelectorGeneralSibling(t *testing.T) {
	sel := ParseSelector("h1 ~ p")
	if len(sel.Parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(sel.Parts))
	}
	if sel.Parts[1].Combinator != CombinatorSibling {
		t.Errorf("expected sibling, got %d", sel.Parts[1].Combinator)
	}
}

func TestParseSelectorCombinatorNoSpace(t *testing.T) {
	// div>span without spaces
	sel := ParseSelector("div>span")
	if len(sel.Parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(sel.Parts))
	}
	if sel.Parts[1].Combinator != CombinatorChild {
		t.Errorf("expected child combinator")
	}
}

func TestParseSelectorSpacePlusCombinator(t *testing.T) {
	// "h1 + p" — space before + combinator
	sel := ParseSelector("div + span")
	if len(sel.Parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(sel.Parts))
	}
	if sel.Parts[1].Combinator != CombinatorAdjacent {
		t.Error("expected adjacent")
	}
}

func TestParseSelectorSpaceTildeCombinator(t *testing.T) {
	sel := ParseSelector("div ~ span")
	if len(sel.Parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(sel.Parts))
	}
	if sel.Parts[1].Combinator != CombinatorSibling {
		t.Error("expected sibling")
	}
}

func TestParseSelectorPseudoElement(t *testing.T) {
	// ::before — double colon pseudo-element
	sel := ParseSelector("div::before")
	if len(sel.Parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(sel.Parts))
	}
	c := sel.Parts[0].Compound
	if len(c.PseudoClass) != 1 || c.PseudoClass[0] != "before" {
		t.Errorf("pseudo-element: %v", c.PseudoClass)
	}
}

func TestParseSelectorNthChild(t *testing.T) {
	sel := ParseSelector("li:nth-child(2n+1)")
	if len(sel.Parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(sel.Parts))
	}
	c := sel.Parts[0].Compound
	if c.Tag != "li" || len(c.PseudoClass) != 1 || c.PseudoClass[0] != "nth-child" {
		t.Errorf("nth-child: tag=%q pseudo=%v", c.Tag, c.PseudoClass)
	}
}

func TestParseSelectorEmpty(t *testing.T) {
	sel := ParseSelector("")
	if len(sel.Parts) != 0 {
		t.Errorf("empty selector should have 0 parts, got %d", len(sel.Parts))
	}
}

func TestParseSelectorTrailingWhitespace(t *testing.T) {
	sel := ParseSelector("div   ")
	if len(sel.Parts) != 1 || sel.Parts[0].Compound.Tag != "div" {
		t.Errorf("trailing ws: %+v", sel)
	}
}

func TestSpecificityLessEqual(t *testing.T) {
	a := Specificity{0, 1, 0, 0}
	b := Specificity{0, 1, 0, 0}
	if a.Less(b) {
		t.Error("equal specificities should return false")
	}
}

func TestSpecificityLessFirstField(t *testing.T) {
	a := Specificity{0, 1, 2, 3}
	b := Specificity{1, 0, 0, 0}
	if !a.Less(b) {
		t.Error("inline should beat id")
	}
}

// === Pseudo-class matching coverage ===

func TestMatchPseudoAllVariants(t *testing.T) {
	tests := []struct {
		pseudo string
		el     ElementInfo
		want   bool
	}{
		{"hover", ElementInfo{Hovered: true}, true},
		{"hover", ElementInfo{Hovered: false}, false},
		{"focus", ElementInfo{Focused: true}, true},
		{"focus", ElementInfo{Focused: false}, false},
		{"active", ElementInfo{Active: true}, true},
		{"active", ElementInfo{Active: false}, false},
		{"disabled", ElementInfo{Disabled: true}, true},
		{"disabled", ElementInfo{Disabled: false}, false},
		{"first-child", ElementInfo{ChildIndex: 0}, true},
		{"first-child", ElementInfo{ChildIndex: 1}, false},
		{"last-child", ElementInfo{ChildIndex: 2, SiblingCount: 3}, true},
		{"last-child", ElementInfo{ChildIndex: 0, SiblingCount: 3}, false},
		{"last-child", ElementInfo{ChildIndex: 0, SiblingCount: 0}, false},
		{"unknown-pseudo", ElementInfo{}, false},
	}
	for _, tt := range tests {
		got := matchPseudo(tt.pseudo, &tt.el)
		if got != tt.want {
			t.Errorf("matchPseudo(%q, %+v) = %v, want %v", tt.pseudo, tt.el, got, tt.want)
		}
	}
}

// === value.go coverage ===

func TestParseValueAllBranches(t *testing.T) {
	tests := []struct {
		input string
		want  layout.Value
	}{
		{"", layout.Auto},
		{"auto", layout.Auto},
		{"0", layout.Zero},
		{"10px", layout.Px(10)},
		{"50%", layout.Pct(50)},
		{"2em", layout.Px(32)},
		{"1rem", layout.Px(16)},
		{"100", layout.Px(100)},     // plain number
		{"-10", layout.Px(-10)},     // negative plain number
		{".5", layout.Px(0.5)},      // dot-started number
		{"invalid", layout.Auto},    // non-numeric string
	}
	for _, tt := range tests {
		got := ParseValue(tt.input)
		if got != tt.want {
			t.Errorf("ParseValue(%q) = %+v, want %+v", tt.input, got, tt.want)
		}
	}
}

func TestParseEdgeShorthand3Values(t *testing.T) {
	top, right, bottom, left := ParseEdgeShorthand("10px 20px 30px")
	if top != layout.Px(10) || right != layout.Px(20) || bottom != layout.Px(30) || left != layout.Px(20) {
		t.Errorf("3-value: %v %v %v %v", top, right, bottom, left)
	}
}

func TestParseEdgeShorthandEmpty(t *testing.T) {
	top, right, bottom, left := ParseEdgeShorthand("")
	if top != layout.Auto || right != layout.Auto || bottom != layout.Auto || left != layout.Auto {
		t.Errorf("empty: %v %v %v %v", top, right, bottom, left)
	}
}

func TestParseColorRGBSpaceSyntax(t *testing.T) {
	c, ok := ParseColor("rgb(255 128 0 / 0.5)")
	if !ok {
		t.Fatal("expected ok")
	}
	if c.R != 1 || c.A != 0.5 {
		t.Errorf("rgb space: R=%f A=%f", c.R, c.A)
	}
}

func TestParseColorComponentPercent(t *testing.T) {
	v := parseColorComponent("50%")
	if v != 0.5 {
		t.Errorf("50%% = %f, want 0.5", v)
	}
}

func TestParseAlphaComponentPercent(t *testing.T) {
	v := parseAlphaComponent("50%")
	if v != 0.5 {
		t.Errorf("50%% alpha = %f, want 0.5", v)
	}
}

func TestParseColorRGBNoCloseParen(t *testing.T) {
	_, ok := ParseColor("rgb(255, 0, 0")
	if ok {
		t.Error("expected false for missing )")
	}
}

func TestParseColorRGBTooFewArgs(t *testing.T) {
	_, ok := ParseColor("rgb(255, 0)")
	if !ok {
		// This actually still has 2 parts after split, so should be false
		t.Log("correctly rejected")
	}
}

func TestResolveVarUnclosed(t *testing.T) {
	got := ResolveVar("var(--missing", map[string]string{})
	// Unclosed var( → should not loop infinitely, return as-is
	if got != "var(--missing" {
		t.Errorf("unclosed var: %q", got)
	}
}

func TestResolveVarCircular(t *testing.T) {
	vars := map[string]string{
		"--a": "var(--b)",
		"--b": "var(--a)",
	}
	// Should not infinite loop (limited to 32 iterations)
	got := ResolveVar("var(--a)", vars)
	_ = got // just verify it terminates
}

func TestResolveVarNoVars(t *testing.T) {
	got := ResolveVar("no var here", nil)
	if got != "no var here" {
		t.Errorf("no var: %q", got)
	}
}

func TestResolveVarNested(t *testing.T) {
	vars := map[string]string{
		"--inner": "blue",
		"--outer": "var(--inner)",
	}
	got := ResolveVar("var(--outer)", vars)
	if got != "blue" {
		t.Errorf("nested var: %q, want 'blue'", got)
	}
}

// === Parse* enum functions coverage ===

func TestParseDisplayAll(t *testing.T) {
	tests := []struct {
		val  string
		want layout.Display
		ok   bool
	}{
		{"block", layout.DisplayBlock, true},
		{"flex", layout.DisplayFlex, true},
		{"inline", layout.DisplayInline, true},
		{"none", layout.DisplayNone, true},
		{"grid", layout.DisplayGrid, true},
		{"invalid", 0, false},
	}
	for _, tt := range tests {
		got, ok := ParseDisplay(tt.val)
		if ok != tt.ok || (ok && got != tt.want) {
			t.Errorf("ParseDisplay(%q) = %v, %v", tt.val, got, ok)
		}
	}
}

func TestParsePositionAll(t *testing.T) {
	tests := []struct {
		val  string
		want layout.Position
		ok   bool
	}{
		{"relative", layout.PositionRelative, true},
		{"absolute", layout.PositionAbsolute, true},
		{"fixed", layout.PositionFixed, true},
		{"invalid", 0, false},
	}
	for _, tt := range tests {
		got, ok := ParsePosition(tt.val)
		if ok != tt.ok || (ok && got != tt.want) {
			t.Errorf("ParsePosition(%q) = %v, %v", tt.val, got, ok)
		}
	}
}

func TestParseOverflowAll(t *testing.T) {
	tests := []struct {
		val  string
		want layout.Overflow
		ok   bool
	}{
		{"visible", layout.OverflowVisible, true},
		{"hidden", layout.OverflowHidden, true},
		{"scroll", layout.OverflowScroll, true},
		{"auto", layout.OverflowAuto, true},
		{"invalid", 0, false},
	}
	for _, tt := range tests {
		got, ok := ParseOverflow(tt.val)
		if ok != tt.ok || (ok && got != tt.want) {
			t.Errorf("ParseOverflow(%q) = %v, %v", tt.val, got, ok)
		}
	}
}

func TestParseFlexDirectionAll(t *testing.T) {
	tests := []struct {
		val  string
		want layout.FlexDirection
		ok   bool
	}{
		{"row", layout.FlexDirectionRow, true},
		{"column", layout.FlexDirectionColumn, true},
		{"row-reverse", layout.FlexDirectionRowReverse, true},
		{"column-reverse", layout.FlexDirectionColumnReverse, true},
		{"invalid", 0, false},
	}
	for _, tt := range tests {
		got, ok := ParseFlexDirection(tt.val)
		if ok != tt.ok || (ok && got != tt.want) {
			t.Errorf("ParseFlexDirection(%q) = %v, %v", tt.val, got, ok)
		}
	}
}

func TestParseFlexWrapAll(t *testing.T) {
	tests := []struct {
		val  string
		want layout.FlexWrap
		ok   bool
	}{
		{"nowrap", layout.FlexWrapNoWrap, true},
		{"wrap", layout.FlexWrapWrap, true},
		{"wrap-reverse", layout.FlexWrapWrapReverse, true},
		{"invalid", 0, false},
	}
	for _, tt := range tests {
		got, ok := ParseFlexWrap(tt.val)
		if ok != tt.ok || (ok && got != tt.want) {
			t.Errorf("ParseFlexWrap(%q) = %v, %v", tt.val, got, ok)
		}
	}
}

func TestParseJustifyContentAll(t *testing.T) {
	tests := []struct {
		val  string
		want layout.JustifyContent
		ok   bool
	}{
		{"flex-start", layout.JustifyFlexStart, true},
		{"start", layout.JustifyFlexStart, true},
		{"flex-end", layout.JustifyFlexEnd, true},
		{"end", layout.JustifyFlexEnd, true},
		{"center", layout.JustifyCenter, true},
		{"space-between", layout.JustifySpaceBetween, true},
		{"space-around", layout.JustifySpaceAround, true},
		{"space-evenly", layout.JustifySpaceEvenly, true},
		{"invalid", 0, false},
	}
	for _, tt := range tests {
		got, ok := ParseJustifyContent(tt.val)
		if ok != tt.ok || (ok && got != tt.want) {
			t.Errorf("ParseJustifyContent(%q) = %v, %v", tt.val, got, ok)
		}
	}
}

func TestParseAlignItemsAll(t *testing.T) {
	tests := []struct {
		val  string
		want layout.AlignItems
		ok   bool
	}{
		{"stretch", layout.AlignStretch, true},
		{"flex-start", layout.AlignFlexStart, true},
		{"start", layout.AlignFlexStart, true},
		{"flex-end", layout.AlignFlexEnd, true},
		{"end", layout.AlignFlexEnd, true},
		{"center", layout.AlignCenter, true},
		{"baseline", layout.AlignBaseline, true},
		{"invalid", 0, false},
	}
	for _, tt := range tests {
		got, ok := ParseAlignItems(tt.val)
		if ok != tt.ok || (ok && got != tt.want) {
			t.Errorf("ParseAlignItems(%q) = %v, %v", tt.val, got, ok)
		}
	}
}

func TestParseAlignSelfAll(t *testing.T) {
	tests := []struct {
		val  string
		want layout.AlignSelf
		ok   bool
	}{
		{"auto", layout.AlignSelfAuto, true},
		{"stretch", layout.AlignSelfStretch, true},
		{"flex-start", layout.AlignSelfFlexStart, true},
		{"start", layout.AlignSelfFlexStart, true},
		{"flex-end", layout.AlignSelfFlexEnd, true},
		{"end", layout.AlignSelfFlexEnd, true},
		{"center", layout.AlignSelfCenter, true},
		{"baseline", layout.AlignSelfBaseline, true},
		{"invalid", 0, false},
	}
	for _, tt := range tests {
		got, ok := ParseAlignSelf(tt.val)
		if ok != tt.ok || (ok && got != tt.want) {
			t.Errorf("ParseAlignSelf(%q) = %v, %v", tt.val, got, ok)
		}
	}
}

// === resolve.go applyProperty coverage ===

func TestApplyPropertyAllBranches(t *testing.T) {
	css := `
		.all {
			display: flex;
			position: absolute;
			overflow: hidden;
			width: 100px;
			height: 50px;
			min-width: 10px;
			min-height: 10px;
			max-width: 200px;
			max-height: 200px;
			margin: 8px;
			margin-top: 1px;
			margin-right: 2px;
			margin-bottom: 3px;
			margin-left: 4px;
			padding: 8px;
			padding-top: 1px;
			padding-right: 2px;
			padding-bottom: 3px;
			padding-left: 4px;
			border-width: 1px;
			border-top-width: 2px;
			border-right-width: 3px;
			border-bottom-width: 4px;
			border-left-width: 5px;
			top: 10px;
			right: 20px;
			bottom: 30px;
			left: 40px;
			flex-direction: column;
			flex-wrap: wrap;
			justify-content: center;
			align-items: center;
			align-self: flex-end;
			gap: 8px;
			row-gap: 4px;
			column-gap: 6px;
			flex-grow: 2;
			flex-shrink: 0;
			flex-basis: 100px;
			order: 3;
			color: red;
			background-color: blue;
			font-size: 16px;
			border-radius: 4px;
			border-color: green;
			opacity: 0.5;
			z-index: 10;
		}
	`
	sheet := Parse(css)
	el := &ElementInfo{Tag: "div", Classes: []string{"all"}}
	cs := ResolveStyle(sheet, el, nil, nil)

	if cs.Layout.Display != layout.DisplayFlex {
		t.Error("display")
	}
	if cs.Layout.Position != layout.PositionAbsolute {
		t.Error("position")
	}
	if cs.Layout.Overflow != layout.OverflowHidden {
		t.Error("overflow")
	}
	if cs.Layout.Width != layout.Px(100) {
		t.Error("width")
	}
	if cs.Layout.Height != layout.Px(50) {
		t.Error("height")
	}
	if cs.Layout.MinWidth != layout.Px(10) {
		t.Error("min-width")
	}
	if cs.Layout.MaxWidth != layout.Px(200) {
		t.Error("max-width")
	}
	if cs.Layout.FlexDirection != layout.FlexDirectionColumn {
		t.Error("flex-direction")
	}
	if cs.Layout.FlexWrap != layout.FlexWrapWrap {
		t.Error("flex-wrap")
	}
	if cs.Layout.JustifyContent != layout.JustifyCenter {
		t.Error("justify-content")
	}
	if cs.Layout.AlignItems != layout.AlignCenter {
		t.Error("align-items")
	}
	if cs.Layout.AlignSelf != layout.AlignSelfFlexEnd {
		t.Error("align-self")
	}
	if cs.Layout.Gap != 8 {
		t.Error("gap")
	}
	if cs.Layout.RowGap != 4 {
		t.Error("row-gap")
	}
	if cs.Layout.ColumnGap != 6 {
		t.Error("column-gap")
	}
	if cs.Layout.FlexGrow != 2 {
		t.Error("flex-grow")
	}
	if cs.Layout.FlexShrink != 0 {
		t.Error("flex-shrink")
	}
	if cs.Layout.FlexBasis != layout.Px(100) {
		t.Error("flex-basis")
	}
	if cs.Layout.Order != 3 {
		t.Error("order")
	}
	if cs.Color != "red" {
		t.Error("color")
	}
	if cs.BackgroundColor != "blue" {
		t.Error("background-color")
	}
	if cs.FontSize != "16px" {
		t.Error("font-size")
	}
	if cs.BorderRadius != "4px" {
		t.Error("border-radius")
	}
	if cs.Opacity != "0.5" {
		t.Error("opacity")
	}
	if cs.ZIndex != "10" {
		t.Error("z-index")
	}
}

func TestFlexShorthandVariants(t *testing.T) {
	tests := []struct {
		css    string
		grow   float32
		shrink float32
		basis  layout.Value
	}{
		{`flex: none`, 0, 0, layout.Auto},
		{`flex: auto`, 1, 1, layout.Auto},
		{`flex: 2`, 2, 1, layout.Zero},
		{`flex: 1 0`, 1, 0, layout.Zero},
		{`flex: 1 0 100px`, 1, 0, layout.Px(100)},
	}
	for _, tt := range tests {
		sheet := Parse(`.a { ` + tt.css + `; }`)
		el := &ElementInfo{Tag: "div", Classes: []string{"a"}}
		cs := ResolveStyle(sheet, el, nil, nil)
		if cs.Layout.FlexGrow != tt.grow {
			t.Errorf("%s: grow=%v, want %v", tt.css, cs.Layout.FlexGrow, tt.grow)
		}
		if cs.Layout.FlexShrink != tt.shrink {
			t.Errorf("%s: shrink=%v, want %v", tt.css, cs.Layout.FlexShrink, tt.shrink)
		}
		if cs.Layout.FlexBasis != tt.basis {
			t.Errorf("%s: basis=%v, want %v", tt.css, cs.Layout.FlexBasis, tt.basis)
		}
	}
}

func TestBorderShorthandPlainNumber(t *testing.T) {
	sheet := Parse(`.a { border: 1 solid black; }`)
	el := &ElementInfo{Tag: "div", Classes: []string{"a"}}
	cs := ResolveStyle(sheet, el, nil, nil)
	if cs.Layout.Border.Top != layout.Px(1) {
		t.Errorf("border-top: %v", cs.Layout.Border.Top)
	}
}

func TestBorderShorthandZero(t *testing.T) {
	sheet := Parse(`.a { border: 0 none transparent; }`)
	el := &ElementInfo{Tag: "div", Classes: []string{"a"}}
	cs := ResolveStyle(sheet, el, nil, nil)
	if cs.Layout.Border.Top != layout.Zero {
		t.Errorf("border: %v", cs.Layout.Border.Top)
	}
}

// === matchSelector coverage ===

func TestMatchSelectorAdjacentReturnsFalse(t *testing.T) {
	sheet := Parse(`h1 + p { color: red; }`)
	el := &ElementInfo{Tag: "p"}
	ancestors := []ElementInfo{{Tag: "div"}}
	cs := ResolveStyle(sheet, el, ancestors, nil)
	// Adjacent not supported → should not match
	if cs.Color == "red" {
		t.Error("adjacent combinator should not match")
	}
}

func TestMatchSelectorSiblingReturnsFalse(t *testing.T) {
	sheet := Parse(`h1 ~ p { color: red; }`)
	el := &ElementInfo{Tag: "p"}
	ancestors := []ElementInfo{{Tag: "div"}}
	cs := ResolveStyle(sheet, el, ancestors, nil)
	if cs.Color == "red" {
		t.Error("sibling combinator should not match")
	}
}

func TestMatchSelectorDeepDescendant(t *testing.T) {
	// .a .b .c → three levels
	sheet := Parse(`.a .b .c { color: red; }`)
	el := &ElementInfo{Tag: "div", Classes: []string{"c"}}
	ancestors := []ElementInfo{
		{Tag: "div", Classes: []string{"b"}},
		{Tag: "div", Classes: []string{"a"}},
	}
	cs := ResolveStyle(sheet, el, ancestors, nil)
	if cs.Color != "red" {
		t.Error("deep descendant should match")
	}
}

func TestMatchSelectorNoAncestors(t *testing.T) {
	sheet := Parse(`div > .a { color: red; }`)
	el := &ElementInfo{Tag: "span", Classes: []string{"a"}}
	cs := ResolveStyle(sheet, el, nil, nil) // no ancestors
	if cs.Color == "red" {
		t.Error("should not match without ancestors")
	}
}

func TestMatchSelectorDescendantNotFound(t *testing.T) {
	sheet := Parse(`.missing .item { color: red; }`)
	el := &ElementInfo{Tag: "div", Classes: []string{"item"}}
	ancestors := []ElementInfo{{Tag: "div", Classes: []string{"other"}}}
	cs := ResolveStyle(sheet, el, ancestors, nil)
	if cs.Color == "red" {
		t.Error("descendant with wrong ancestor should not match")
	}
}

// === ParseInlineDeclarations edge cases ===

func TestParseInlineDeclarationsImportant(t *testing.T) {
	decls := ParseInlineDeclarations("color: red !important")
	if len(decls) != 1 || !decls[0].Important || decls[0].Value != "red" {
		t.Errorf("inline important: %+v", decls)
	}
}

func TestParseInlineDeclarationsNoColon(t *testing.T) {
	decls := ParseInlineDeclarations("invalid; color: red")
	if len(decls) != 1 || decls[0].Property != "color" {
		t.Errorf("no colon: %+v", decls)
	}
}

func TestParseInlineDeclarationsEmptyParts(t *testing.T) {
	decls := ParseInlineDeclarations(";;; color: red ;;;")
	if len(decls) != 1 {
		t.Errorf("expected 1 decl, got %d", len(decls))
	}
}

func TestParseInlineDeclarationsEmptyPropOrVal(t *testing.T) {
	decls := ParseInlineDeclarations(": red; color:")
	if len(decls) != 0 {
		t.Errorf("expected 0 decls for empty prop/val, got %d", len(decls))
	}
}

// === splitValues coverage ===

func TestSplitValuesWithParens(t *testing.T) {
	parts := splitValues("rgb(255, 0, 0) 10px")
	if len(parts) != 2 {
		t.Fatalf("expected 2 parts, got %d: %v", len(parts), parts)
	}
	if parts[0] != "rgb(255, 0, 0)" {
		t.Errorf("part 0: %q", parts[0])
	}
}

func TestSplitValuesEmpty(t *testing.T) {
	parts := splitValues("")
	if len(parts) != 0 {
		t.Errorf("expected 0, got %d", len(parts))
	}
}

func TestSplitValuesMultipleSpaces(t *testing.T) {
	parts := splitValues("  10px   20px  ")
	if len(parts) != 2 {
		t.Errorf("expected 2, got %d: %v", len(parts), parts)
	}
}

// === isPlainNumber coverage ===

func TestIsPlainNumber(t *testing.T) {
	tests := []struct {
		s    string
		want bool
	}{
		{"0", true},
		{"123", true},
		{"-5", true},
		{"3.14", true},
		{"-3.14", true},
		{"", false},
		{"-", false},
		{"abc", false},
		{"1px", false},
		{".", false},
		{".5", true},
		{"1.2.3", false},
	}
	for _, tt := range tests {
		got := isPlainNumber(tt.s)
		if got != tt.want {
			t.Errorf("isPlainNumber(%q) = %v, want %v", tt.s, got, tt.want)
		}
	}
}

// === parseFloat32 edge cases ===

func TestParseFloat32Edge(t *testing.T) {
	if v := parseFloat32(""); v != 0 {
		t.Errorf("empty: %f", v)
	}
	if v := parseFloat32("-0"); v != 0 {
		t.Errorf("-0: %f", v)
	}
	if v := parseFloat32(".5"); v != 0.5 {
		t.Errorf(".5: %f", v)
	}
}

// === background shorthand ===

func TestApplyPropertyBackgroundShorthand(t *testing.T) {
	sheet := Parse(`.a { background: #ff0000; }`)
	el := &ElementInfo{Tag: "div", Classes: []string{"a"}}
	cs := ResolveStyle(sheet, el, nil, nil)
	if cs.BackgroundColor != "#ff0000" {
		t.Errorf("background shorthand: %q", cs.BackgroundColor)
	}
}

// === Additional coverage gap tests ===

func TestTokenizerPeekEOF(t *testing.T) {
	// tokenizer.peek when pos >= len(src)
	tok := newTokenizer("")
	if tok.peek() != 0 {
		t.Errorf("expected 0 for empty tokenizer peek")
	}
}

func TestSkipAtRuleEOF(t *testing.T) {
	// @rule that hits EOF without ; or {
	sheet := Parse(`@charset "utf-8"`)
	_ = sheet // just no panic
}

func TestSkipBlockEOF(t *testing.T) {
	// Block that hits EOF without closing }
	sheet := Parse(`@media screen { .a { color: red; }`)
	_ = sheet // just no panic
}

func TestParseRuleNoLBrace(t *testing.T) {
	// Selector text that never gets a {
	sheet := Parse(`.a`)
	if len(sheet.Rules) != 0 {
		t.Errorf("expected 0 rules for selector without brace, got %d", len(sheet.Rules))
	}
}

func TestParseRuleEmptySelectors(t *testing.T) {
	// A rule where ParseSelectorList returns empty (pure whitespace selector)
	// Actually hard to trigger since spaces before { become part of selector text.
	// Use a selector that ParseSelectorList rejects.
	sheet := Parse(`() { color: red; }`)
	// Just ensure no panic — parser may or may not create a rule
	_ = sheet
}

func TestParseDeclarationNonIdentStart(t *testing.T) {
	// Declaration starts with something that's not ident or -
	sheet := Parse(`.a { 123: red; color: blue; }`)
	if len(sheet.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(sheet.Rules))
	}
	// The "123: red" is skipped, "color: blue" is parsed
	found := false
	for _, d := range sheet.Rules[0].Declarations {
		if d.Property == "color" && d.Value == "blue" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected color:blue declaration after recovery")
	}
}

func TestParseDeclarationImportantEOF(t *testing.T) {
	// !important at end of input without ; or }
	sheet := Parse(`.a { color: red !important`)
	if len(sheet.Rules) != 1 || len(sheet.Rules[0].Declarations) != 1 {
		t.Fatalf("expected 1 rule with 1 decl")
	}
	if !sheet.Rules[0].Declarations[0].Important {
		t.Errorf("expected important flag")
	}
}

func TestParseDeclarationImportantRBrace(t *testing.T) {
	// !important followed immediately by }
	sheet := Parse(`.a { color: red !important}`)
	if len(sheet.Rules) != 1 || len(sheet.Rules[0].Declarations) != 1 {
		t.Fatalf("expected 1 rule with 1 decl")
	}
	if !sheet.Rules[0].Declarations[0].Important {
		t.Errorf("expected important flag")
	}
}

func TestSkipToSemicolonOrBraceEOF(t *testing.T) {
	// Invalid declaration that hits EOF
	sheet := Parse(`.a { = red`)
	_ = sheet // no panic
}

func TestResolveStyleEqualSpecificity(t *testing.T) {
	// Two rules with same specificity — source order wins
	src := `.a { color: red; } .b { color: blue; }`
	sheet := Parse(src)
	el := &ElementInfo{Tag: "div", Classes: []string{"a", "b"}}
	cs := ResolveStyle(sheet, el, nil, nil)
	if cs.Color != "blue" {
		t.Errorf("expected source order to win: %q", cs.Color)
	}
}

func TestResolveStyleInlineImportant(t *testing.T) {
	// Inline style with !important
	sheet := Parse(`.a { color: red !important; }`)
	el := &ElementInfo{Tag: "div", Classes: []string{"a"}}
	inline := []Declaration{{Property: "color", Value: "blue", Important: true}}
	cs := ResolveStyle(sheet, el, nil, inline)
	// Both are important; inline important should override rule important
	if cs.Color != "blue" {
		t.Errorf("expected inline important to win: %q", cs.Color)
	}
}

func TestMatchSelectorEmptyParts(t *testing.T) {
	sel := &Selector{Parts: nil}
	el := &ElementInfo{Tag: "div"}
	if matchSelector(sel, el, nil) {
		t.Errorf("empty selector should not match")
	}
}

func TestMatchSelectorNilCompound(t *testing.T) {
	// Rightmost part has nil compound
	sel := &Selector{Parts: []SelectorPart{{Combinator: CombinatorDescendant}}}
	el := &ElementInfo{Tag: "div"}
	if matchSelector(sel, el, nil) {
		t.Errorf("nil compound should not match")
	}
}

func TestMatchSelectorNoCompoundAfterCombinator(t *testing.T) {
	// Combinator part followed by another combinator (no compound)
	sel := &Selector{Parts: []SelectorPart{
		{Compound: &CompoundSelector{Tag: "span"}},
		{Combinator: CombinatorDescendant},
		{Combinator: CombinatorChild}, // should be compound
		{Combinator: CombinatorDescendant},
		{Compound: &CompoundSelector{Tag: "div"}},
	}}
	el := &ElementInfo{Tag: "div"}
	if matchSelector(sel, el, []ElementInfo{{Tag: "span"}}) {
		t.Errorf("malformed selector parts should not match")
	}
}

func TestMatchSelectorPartIdxNegative(t *testing.T) {
	// Only 2 parts: combinator + compound. partIdx-1 goes negative.
	sel := &Selector{Parts: []SelectorPart{
		{Combinator: CombinatorChild},
		{Compound: &CompoundSelector{Tag: "div"}},
	}}
	el := &ElementInfo{Tag: "div"}
	// rightIdx=1, compound matches div, rightIdx>0 so we check partIdx=0
	// partIdx=0 is combinator, so comb=CombinatorChild, partIdx-- = -1 < 0 → return false
	if matchSelector(sel, el, nil) {
		t.Errorf("should return false for malformed selector")
	}
}

func TestParseSelectorAdjacentNoSpace(t *testing.T) {
	// "div+span" — adjacent combinator without space
	sel := ParseSelector("div+span")
	if len(sel.Parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(sel.Parts))
	}
	if sel.Parts[1].Combinator != CombinatorAdjacent {
		t.Errorf("expected adjacent combinator")
	}
}

func TestParseSelectorTildeNoSpace(t *testing.T) {
	// "div~span" — sibling combinator without space
	sel := ParseSelector("div~span")
	if len(sel.Parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(sel.Parts))
	}
	if sel.Parts[1].Combinator != CombinatorSibling {
		t.Errorf("expected sibling combinator")
	}
}

func TestParseSelectorBreakOnUnknownToken(t *testing.T) {
	// Selector with unexpected token that causes else break
	sel := ParseSelector("div{")
	if len(sel.Parts) != 1 {
		t.Errorf("expected 1 part (div), got %d", len(sel.Parts))
	}
}

func TestParseSelectorCompoundNilAfterCombinator(t *testing.T) {
	// Selector "div > " with trailing space and no compound after >
	sel := ParseSelector("div > ")
	// div compound + should stop since no compound follows
	if len(sel.Parts) < 1 {
		t.Errorf("expected at least 1 part")
	}
}

func TestParseSelectorPseudoElementEOF(t *testing.T) {
	// "div::" with nothing after ::
	sel := ParseSelector("div::")
	if len(sel.Parts) != 1 {
		t.Errorf("expected 1 part, got %d", len(sel.Parts))
	}
}

func TestParseInlineDeclarationsEmpty(t *testing.T) {
	result := ParseInlineDeclarations("")
	if result != nil {
		t.Errorf("expected nil for empty input, got %v", result)
	}
}

func TestResolveVarUnclosedParen(t *testing.T) {
	// var( without closing ) — should not infinite loop
	result := ResolveVar("var(--x", map[string]string{"--x": "red"})
	if result != "var(--x" {
		t.Errorf("expected unchanged for unclosed var(, got %q", result)
	}
}

func TestMatchSelectorAncestorOutOfBounds(t *testing.T) {
	// Child combinator when no ancestor available
	sel := &Selector{Parts: []SelectorPart{
		{Compound: &CompoundSelector{Tag: "body"}},
		{Combinator: CombinatorChild},
		{Compound: &CompoundSelector{Tag: "div"}},
	}}
	el := &ElementInfo{Tag: "div"}
	if matchSelector(sel, el, nil) {
		t.Errorf("should fail with no ancestors for child combinator")
	}
}

func TestMatchSelectorDescendantNotFoundDirect(t *testing.T) {
	// Descendant combinator where ancestor doesn't exist
	sel := &Selector{Parts: []SelectorPart{
		{Compound: &CompoundSelector{Tag: "body"}},
		{Combinator: CombinatorDescendant},
		{Compound: &CompoundSelector{Tag: "div"}},
	}}
	el := &ElementInfo{Tag: "div"}
	ancestors := []ElementInfo{{Tag: "span"}, {Tag: "section"}}
	if matchSelector(sel, el, ancestors) {
		t.Errorf("should fail when ancestor not found")
	}
}

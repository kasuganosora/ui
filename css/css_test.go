package css

import (
	"testing"

	"github.com/kasuganosora/ui/layout"
)

// --- Tokenizer tests ---

func TestTokenizeBasic(t *testing.T) {
	tokens := newTokenizer(".foo { color: red; }").tokenize()
	// Should produce: . foo WS { WS color : WS red ; WS }
	types := make([]TokenType, len(tokens))
	for i, tok := range tokens {
		types[i] = tok.Type
	}
	if len(tokens) < 8 {
		t.Fatalf("expected at least 8 tokens, got %d: %v", len(tokens), tokens)
	}
	if tokens[0].Type != TokenDot {
		t.Errorf("expected dot, got %v", tokens[0])
	}
	if tokens[1].Type != TokenIdent || tokens[1].Value != "foo" {
		t.Errorf("expected ident 'foo', got %v", tokens[1])
	}
}

func TestTokenizeNumbers(t *testing.T) {
	tokens := newTokenizer("12px 50% 3.14 -5em").tokenize()
	// Filter non-whitespace
	var filtered []Token
	for _, tok := range tokens {
		if tok.Type != TokenWhitespace {
			filtered = append(filtered, tok)
		}
	}
	if len(filtered) != 4 {
		t.Fatalf("expected 4 non-ws tokens, got %d", len(filtered))
	}
	if filtered[0].Type != TokenDimension || filtered[0].Num != 12 || filtered[0].Unit != "px" {
		t.Errorf("12px: %+v", filtered[0])
	}
	if filtered[1].Type != TokenDimension || filtered[1].Num != 50 || filtered[1].Unit != "%" {
		t.Errorf("50%%: %+v", filtered[1])
	}
	if filtered[2].Type != TokenNumber || filtered[2].Num != 3.14 {
		t.Errorf("3.14: %+v", filtered[2])
	}
	if filtered[3].Type != TokenDimension || filtered[3].Num != -5 || filtered[3].Unit != "em" {
		t.Errorf("-5em: %+v", filtered[3])
	}
}

func TestTokenizeHashAndString(t *testing.T) {
	tokens := newTokenizer(`#ff0000 "hello"`).tokenize()
	var filtered []Token
	for _, tok := range tokens {
		if tok.Type != TokenWhitespace {
			filtered = append(filtered, tok)
		}
	}
	if len(filtered) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(filtered))
	}
	if filtered[0].Type != TokenHash || filtered[0].Value != "#ff0000" {
		t.Errorf("hash: %+v", filtered[0])
	}
	if filtered[1].Type != TokenString || filtered[1].Value != "hello" {
		t.Errorf("string: %+v", filtered[1])
	}
}

func TestTokenizeComment(t *testing.T) {
	tokens := newTokenizer("a /* comment */ b").tokenize()
	var filtered []Token
	for _, tok := range tokens {
		if tok.Type != TokenWhitespace {
			filtered = append(filtered, tok)
		}
	}
	if len(filtered) != 2 {
		t.Fatalf("expected 2 tokens after comment skip, got %d", len(filtered))
	}
}

func TestTokenizeFunction(t *testing.T) {
	tokens := newTokenizer("rgb(255, 0, 0)").tokenize()
	if tokens[0].Type != TokenFunction || tokens[0].Value != "rgb" {
		t.Errorf("expected function 'rgb', got %+v", tokens[0])
	}
}

// --- Selector tests ---

func TestParseSelectorTag(t *testing.T) {
	sel := ParseSelector("div")
	if len(sel.Parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(sel.Parts))
	}
	if sel.Parts[0].Compound.Tag != "div" {
		t.Errorf("expected tag 'div', got %q", sel.Parts[0].Compound.Tag)
	}
}

func TestParseSelectorClass(t *testing.T) {
	sel := ParseSelector(".container")
	if len(sel.Parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(sel.Parts))
	}
	c := sel.Parts[0].Compound
	if len(c.Classes) != 1 || c.Classes[0] != "container" {
		t.Errorf("expected class 'container', got %v", c.Classes)
	}
}

func TestParseSelectorID(t *testing.T) {
	sel := ParseSelector("#main")
	if len(sel.Parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(sel.Parts))
	}
	if sel.Parts[0].Compound.ID != "main" {
		t.Errorf("expected id 'main', got %q", sel.Parts[0].Compound.ID)
	}
}

func TestParseSelectorCompound(t *testing.T) {
	sel := ParseSelector("div.foo#bar")
	if len(sel.Parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(sel.Parts))
	}
	c := sel.Parts[0].Compound
	if c.Tag != "div" || c.ID != "bar" || len(c.Classes) != 1 || c.Classes[0] != "foo" {
		t.Errorf("unexpected compound: tag=%q id=%q classes=%v", c.Tag, c.ID, c.Classes)
	}
}

func TestParseSelectorDescendant(t *testing.T) {
	sel := ParseSelector("div .item")
	if len(sel.Parts) != 3 {
		t.Fatalf("expected 3 parts (compound combinator compound), got %d", len(sel.Parts))
	}
	if sel.Parts[1].Combinator != CombinatorDescendant {
		t.Errorf("expected descendant combinator, got %d", sel.Parts[1].Combinator)
	}
}

func TestParseSelectorChild(t *testing.T) {
	sel := ParseSelector("div > span")
	if len(sel.Parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(sel.Parts))
	}
	if sel.Parts[1].Combinator != CombinatorChild {
		t.Errorf("expected child combinator, got %d", sel.Parts[1].Combinator)
	}
}

func TestParseSelectorPseudoClass(t *testing.T) {
	sel := ParseSelector("button:hover")
	if len(sel.Parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(sel.Parts))
	}
	c := sel.Parts[0].Compound
	if c.Tag != "button" || len(c.PseudoClass) != 1 || c.PseudoClass[0] != "hover" {
		t.Errorf("unexpected: tag=%q pseudo=%v", c.Tag, c.PseudoClass)
	}
}

func TestParseSelectorList(t *testing.T) {
	sels := ParseSelectorList("h1, h2, h3")
	if len(sels) != 3 {
		t.Fatalf("expected 3 selectors, got %d", len(sels))
	}
	for i, tag := range []string{"h1", "h2", "h3"} {
		if sels[i].Parts[0].Compound.Tag != tag {
			t.Errorf("sel[%d] expected %q, got %q", i, tag, sels[i].Parts[0].Compound.Tag)
		}
	}
}

func TestParseSelectorUniversal(t *testing.T) {
	sel := ParseSelector("*")
	if len(sel.Parts) != 1 || !sel.Parts[0].Compound.Universal {
		t.Error("expected universal selector")
	}
}

// --- Specificity tests ---

func TestSpecificity(t *testing.T) {
	tests := []struct {
		sel  string
		spec Specificity
	}{
		{"*", Specificity{0, 0, 0, 0}},
		{"div", Specificity{0, 0, 0, 1}},
		{".foo", Specificity{0, 0, 1, 0}},
		{"#bar", Specificity{0, 1, 0, 0}},
		{"div.foo", Specificity{0, 0, 1, 1}},
		{"div.foo#bar", Specificity{0, 1, 1, 1}},
		{"div:hover", Specificity{0, 0, 1, 1}},
		{".a.b.c", Specificity{0, 0, 3, 0}},
	}
	for _, tt := range tests {
		sel := ParseSelector(tt.sel)
		spec := SelectorSpecificity(&sel)
		if spec != tt.spec {
			t.Errorf("Specificity(%q) = %v, want %v", tt.sel, spec, tt.spec)
		}
	}
}

// --- Selector matching tests ---

func TestMatchCompound(t *testing.T) {
	el := &ElementInfo{Tag: "div", ID: "main", Classes: []string{"container", "wide"}}

	tests := []struct {
		sel   string
		match bool
	}{
		{"div", true},
		{".container", true},
		{"#main", true},
		{"div.container#main", true},
		{"span", false},
		{".narrow", false},
		{"#other", false},
		{"*", true},
	}

	for _, tt := range tests {
		s := ParseSelector(tt.sel)
		got := MatchCompound(s.Parts[0].Compound, el)
		if got != tt.match {
			t.Errorf("MatchCompound(%q) = %v, want %v", tt.sel, got, tt.match)
		}
	}
}

func TestMatchPseudoClass(t *testing.T) {
	el := &ElementInfo{Tag: "button", Hovered: true, ChildIndex: 0, SiblingCount: 3}
	s := ParseSelector("button:hover")
	if !MatchCompound(s.Parts[0].Compound, el) {
		t.Error("expected :hover to match")
	}

	s2 := ParseSelector("button:first-child")
	if !MatchCompound(s2.Parts[0].Compound, el) {
		t.Error("expected :first-child to match")
	}

	s3 := ParseSelector("button:last-child")
	if MatchCompound(s3.Parts[0].Compound, el) {
		t.Error("expected :last-child to NOT match for index 0 of 3")
	}
}

// --- Parser (stylesheet) tests ---

func TestParseStylesheet(t *testing.T) {
	css := `
		.container { display: flex; gap: 12px; }
		#header { color: red; font-size: 24px; }
		div > span { padding: 8px 16px; }
	`
	sheet := Parse(css)
	if len(sheet.Rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(sheet.Rules))
	}

	// Rule 0: .container
	r0 := sheet.Rules[0]
	if len(r0.Declarations) != 2 {
		t.Errorf("rule 0: expected 2 declarations, got %d", len(r0.Declarations))
	}
	if r0.Declarations[0].Property != "display" || r0.Declarations[0].Value != "flex" {
		t.Errorf("rule 0 decl 0: %+v", r0.Declarations[0])
	}

	// Rule 1: #header
	r1 := sheet.Rules[1]
	if len(r1.Selectors) != 1 || r1.Selectors[0].Parts[0].Compound.ID != "header" {
		t.Errorf("rule 1 selector: %+v", r1.Selectors)
	}

	// Rule 2: div > span
	r2 := sheet.Rules[2]
	if len(r2.Selectors[0].Parts) != 3 {
		t.Errorf("rule 2 selector parts: expected 3, got %d", len(r2.Selectors[0].Parts))
	}
}

func TestParseImportant(t *testing.T) {
	sheet := Parse(`.a { color: red !important; }`)
	if len(sheet.Rules) != 1 || len(sheet.Rules[0].Declarations) != 1 {
		t.Fatal("expected 1 rule with 1 declaration")
	}
	d := sheet.Rules[0].Declarations[0]
	if !d.Important {
		t.Error("expected !important")
	}
	if d.Value != "red" {
		t.Errorf("expected value 'red', got %q", d.Value)
	}
}

func TestParseCSSVariables(t *testing.T) {
	sheet := Parse(`:root { --primary: #1890ff; --gap: 12px; }`)
	if sheet.Variables["--primary"] != "#1890ff" {
		t.Errorf("--primary: %q", sheet.Variables["--primary"])
	}
	if sheet.Variables["--gap"] != "12px" {
		t.Errorf("--gap: %q", sheet.Variables["--gap"])
	}
}

func TestParseComment(t *testing.T) {
	sheet := Parse(`/* header style */ .a { color: red; /* inline */ }`)
	if len(sheet.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(sheet.Rules))
	}
}

func TestParseMultipleSelectors(t *testing.T) {
	sheet := Parse(`h1, h2, h3 { font-size: 24px; }`)
	if len(sheet.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(sheet.Rules))
	}
	if len(sheet.Rules[0].Selectors) != 3 {
		t.Fatalf("expected 3 selectors, got %d", len(sheet.Rules[0].Selectors))
	}
}

// --- Value parsing tests ---

func TestParseValue(t *testing.T) {
	tests := []struct {
		input string
		want  layout.Value
	}{
		{"12px", layout.Px(12)},
		{"50%", layout.Pct(50)},
		{"auto", layout.Auto},
		{"0", layout.Zero},
		{"1.5em", layout.Px(24)}, // 1.5 * 16
	}
	for _, tt := range tests {
		got := ParseValue(tt.input)
		if got != tt.want {
			t.Errorf("ParseValue(%q) = %+v, want %+v", tt.input, got, tt.want)
		}
	}
}

func TestParseEdgeShorthand(t *testing.T) {
	// 1 value
	top, right, bottom, left := ParseEdgeShorthand("10px")
	if top != layout.Px(10) || right != layout.Px(10) || bottom != layout.Px(10) || left != layout.Px(10) {
		t.Errorf("1-value shorthand: %v %v %v %v", top, right, bottom, left)
	}

	// 2 values
	top, right, bottom, left = ParseEdgeShorthand("10px 20px")
	if top != layout.Px(10) || right != layout.Px(20) || bottom != layout.Px(10) || left != layout.Px(20) {
		t.Errorf("2-value shorthand: %v %v %v %v", top, right, bottom, left)
	}

	// 4 values
	top, right, bottom, left = ParseEdgeShorthand("10px 20px 30px 40px")
	if top != layout.Px(10) || right != layout.Px(20) || bottom != layout.Px(30) || left != layout.Px(40) {
		t.Errorf("4-value shorthand: %v %v %v %v", top, right, bottom, left)
	}
}

func TestParseColor(t *testing.T) {
	tests := []struct {
		input string
		ok    bool
	}{
		{"#ff0000", true},
		{"red", true},
		{"rgb(255, 0, 0)", true},
		{"rgba(0, 0, 0, 0.5)", true},
		{"transparent", true},
		{"", false},
		{"invalid", false},
	}
	for _, tt := range tests {
		_, ok := ParseColor(tt.input)
		if ok != tt.ok {
			t.Errorf("ParseColor(%q) ok=%v, want %v", tt.input, ok, tt.ok)
		}
	}
}

func TestParseColorRGBA(t *testing.T) {
	c, ok := ParseColor("rgba(255, 128, 0, 0.5)")
	if !ok {
		t.Fatal("expected ok")
	}
	if c.R != 1 || c.A != 0.5 {
		t.Errorf("rgba: R=%f A=%f", c.R, c.A)
	}
}

// --- ResolveVar tests ---

func TestResolveVar(t *testing.T) {
	vars := map[string]string{
		"--primary": "#1890ff",
		"--gap":     "12px",
	}

	if got := ResolveVar("var(--primary)", vars); got != "#1890ff" {
		t.Errorf("got %q", got)
	}
	if got := ResolveVar("var(--missing, blue)", vars); got != "blue" {
		t.Errorf("fallback: got %q", got)
	}
	if got := ResolveVar("1px solid var(--primary)", vars); got != "1px solid #1890ff" {
		t.Errorf("inline var: got %q", got)
	}
}

// --- Style resolution tests ---

func TestResolveStyleBasic(t *testing.T) {
	sheet := Parse(`.container { display: flex; gap: 12px; padding: 8px; }`)
	el := &ElementInfo{Tag: "div", Classes: []string{"container"}}
	cs := ResolveStyle(sheet, el, nil, nil)

	if cs.Layout.Display != layout.DisplayFlex {
		t.Errorf("display: %v", cs.Layout.Display)
	}
	if cs.Layout.Gap != 12 {
		t.Errorf("gap: %v", cs.Layout.Gap)
	}
	if cs.Layout.Padding.Top != layout.Px(8) {
		t.Errorf("padding-top: %v", cs.Layout.Padding.Top)
	}
}

func TestResolveStyleCascade(t *testing.T) {
	sheet := Parse(`
		div { color: red; }
		.special { color: blue; }
	`)
	// .special has higher specificity than div
	el := &ElementInfo{Tag: "div", Classes: []string{"special"}}
	cs := ResolveStyle(sheet, el, nil, nil)
	if cs.Color != "blue" {
		t.Errorf("cascade: color=%q, want 'blue'", cs.Color)
	}
}

func TestResolveStyleImportant(t *testing.T) {
	sheet := Parse(`
		.a { color: red !important; }
		#b { color: blue; }
	`)
	el := &ElementInfo{Tag: "div", ID: "b", Classes: []string{"a"}}
	cs := ResolveStyle(sheet, el, nil, nil)
	if cs.Color != "red" {
		t.Errorf("important: color=%q, want 'red'", cs.Color)
	}
}

func TestResolveStyleInline(t *testing.T) {
	sheet := Parse(`.a { color: red; }`)
	el := &ElementInfo{Tag: "div", Classes: []string{"a"}}
	inline := ParseInlineDeclarations("color: green")
	cs := ResolveStyle(sheet, el, nil, inline)
	if cs.Color != "green" {
		t.Errorf("inline: color=%q, want 'green'", cs.Color)
	}
}

func TestResolveStyleVariables(t *testing.T) {
	sheet := Parse(`
		:root { --bg: #333; }
		.panel { background-color: var(--bg); }
	`)
	el := &ElementInfo{Tag: "div", Classes: []string{"panel"}}
	cs := ResolveStyle(sheet, el, nil, nil)
	if cs.BackgroundColor != "#333" {
		t.Errorf("var: background=%q, want '#333'", cs.BackgroundColor)
	}
}

func TestResolveStyleDescendant(t *testing.T) {
	sheet := Parse(`div .item { color: red; }`)
	el := &ElementInfo{Tag: "span", Classes: []string{"item"}}
	ancestors := []ElementInfo{{Tag: "div"}}
	cs := ResolveStyle(sheet, el, ancestors, nil)
	if cs.Color != "red" {
		t.Errorf("descendant: color=%q, want 'red'", cs.Color)
	}
}

func TestResolveStyleChild(t *testing.T) {
	sheet := Parse(`div > .item { color: blue; }`)
	el := &ElementInfo{Tag: "span", Classes: []string{"item"}}
	ancestors := []ElementInfo{{Tag: "div"}}
	cs := ResolveStyle(sheet, el, ancestors, nil)
	if cs.Color != "blue" {
		t.Errorf("child: color=%q, want 'blue'", cs.Color)
	}

	// Should NOT match if parent is not div
	ancestors2 := []ElementInfo{{Tag: "span"}}
	cs2 := ResolveStyle(sheet, el, ancestors2, nil)
	if cs2.Color == "blue" {
		t.Error("child combinator should not match non-div parent")
	}
}

func TestResolveFlexShorthand(t *testing.T) {
	sheet := Parse(`.item { flex: 1; }`)
	el := &ElementInfo{Tag: "div", Classes: []string{"item"}}
	cs := ResolveStyle(sheet, el, nil, nil)
	if cs.Layout.FlexGrow != 1 {
		t.Errorf("flex-grow: %v", cs.Layout.FlexGrow)
	}
}

func TestResolveBorderShorthand(t *testing.T) {
	sheet := Parse(`.box { border: 2px solid #ccc; }`)
	el := &ElementInfo{Tag: "div", Classes: []string{"box"}}
	cs := ResolveStyle(sheet, el, nil, nil)
	if cs.Layout.Border.Top != layout.Px(2) {
		t.Errorf("border-top: %v", cs.Layout.Border.Top)
	}
	if cs.BorderColor != "#ccc" {
		t.Errorf("border-color: %q", cs.BorderColor)
	}
}

// --- ParseInlineDeclarations tests ---

func TestParseInlineDeclarations(t *testing.T) {
	decls := ParseInlineDeclarations("color: red; font-size: 14px")
	if len(decls) != 2 {
		t.Fatalf("expected 2, got %d", len(decls))
	}
	if decls[0].Property != "color" || decls[0].Value != "red" {
		t.Errorf("decl 0: %+v", decls[0])
	}
	if decls[1].Property != "font-size" || decls[1].Value != "14px" {
		t.Errorf("decl 1: %+v", decls[1])
	}
}

package ui

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	"github.com/kasuganosora/ui/widget"
)

func TestLoadHTMLEmpty(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, "")
	if root == nil {
		t.Fatal("expected root widget")
	}
}

func TestLoadHTMLNilConfig(t *testing.T) {
	tree := core.NewTree()
	root := LoadHTML(tree, nil, "<p>test</p>")
	if root == nil {
		t.Fatal("expected root widget with nil config")
	}
}

func TestLoadHTMLText(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, "<p>Hello World</p>")
	if len(root.Children()) != 1 {
		t.Errorf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestLoadHTMLBareText(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, "Hello plain text")
	if len(root.Children()) != 1 {
		t.Errorf("expected 1 text child, got %d", len(root.Children()))
	}
}

func TestLoadHTMLDiv(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<div><p>A</p><p>B</p></div>`)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child (div), got %d", len(root.Children()))
	}
	div := root.Children()[0]
	if len(div.Children()) != 2 {
		t.Errorf("expected 2 children in div, got %d", len(div.Children()))
	}
}

func TestLoadHTMLButton(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<button>Click</button>`)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestLoadHTMLButtonDisabled(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<button disabled>No</button>`)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestLoadHTMLInput(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<input placeholder="name" value="test"/>`)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestLoadHTMLInputDisabled(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<input disabled/>`)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestLoadHTMLHeadings(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<h1>Title</h1><h3>Subtitle</h3><h6>Small</h6>`)
	if len(root.Children()) != 3 {
		t.Errorf("expected 3 children, got %d", len(root.Children()))
	}
}

func TestLoadHTMLSpan(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<span style="color: red; font-size: 20px">styled</span>`)
	if len(root.Children()) != 1 {
		t.Errorf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestLoadHTMLBr(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<br/>`)
	// br tags don't produce widgets
	_ = root
}

func TestLoadHTMLImg(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<img src="test.png" style="width: 100px; height: 50px"/>`)
	if len(root.Children()) != 1 {
		t.Errorf("expected 1 child (img), got %d", len(root.Children()))
	}
}

func TestLoadHTMLInlineStyle(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<div style="background-color: #ff0000; padding: 16px"><p>styled</p></div>`)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestLoadHTMLDivStyleVariants(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	html := `<div style="border-radius: 8px; border-color: #333; border-width: 2px; display: flex; flex-direction: row; gap: 4px; width: 100px; height: 50px; padding: 8px">test</div>`
	root := LoadHTML(tree, cfg, html)
	if len(root.Children()) != 1 {
		t.Errorf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestLoadHTMLDivDisplayGrid(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<div style="display: grid">grid</div>`)
	if len(root.Children()) != 1 {
		t.Errorf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestLoadHTMLDivDisplayNone(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<div style="display: none">hidden</div>`)
	if len(root.Children()) != 1 {
		t.Errorf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestLoadHTMLDivFlexColumn(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<div style="display: flex; flex-direction: column">col</div>`)
	if len(root.Children()) != 1 {
		t.Errorf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestLoadHTMLDivClass(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<div class="foo bar">classed</div>`)
	if len(root.Children()) != 1 {
		t.Errorf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestLoadHTMLNested(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	html := `
	<div style="display: flex; flex-direction: column; gap: 8px">
		<h2>Form</h2>
		<input placeholder="Name"/>
		<button>Submit</button>
	</div>`
	root := LoadHTML(tree, cfg, html)
	if len(root.Children()) != 1 {
		t.Errorf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestLoadHTMLSelfClosingDiv(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<div/>`)
	// Self-closing div should not panic
	_ = root
}

func TestLoadHTMLUnknownTag(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<custom>content</custom>`)
	// Unknown tags treated as divs
	if len(root.Children()) == 0 {
		t.Error("expected at least 1 child for unknown tag")
	}
}

func TestParseInlineCSS(t *testing.T) {
	props := parseInlineCSS("color: red; font-size: 16px; background: #fff")
	if props["color"] != "red" {
		t.Errorf("expected color=red, got %q", props["color"])
	}
	if props["font-size"] != "16px" {
		t.Errorf("expected font-size=16px, got %q", props["font-size"])
	}
	if props["background"] != "#fff" {
		t.Errorf("expected background=#fff, got %q", props["background"])
	}
}

func TestParseInlineCSSEmpty(t *testing.T) {
	props := parseInlineCSS("")
	if len(props) != 0 {
		t.Errorf("expected 0 props, got %d", len(props))
	}
}

func TestParsePx(t *testing.T) {
	tests := []struct {
		in   string
		want float32
	}{
		{"16px", 16},
		{"24", 24},
		{"0px", 0},
	}
	for _, tt := range tests {
		got, err := parsePx(tt.in)
		if err != nil {
			t.Errorf("parsePx(%q) error: %v", tt.in, err)
		}
		if got != tt.want {
			t.Errorf("parsePx(%q) = %g, want %g", tt.in, got, tt.want)
		}
	}
}

func TestParseColorNamed(t *testing.T) {
	tests := []struct {
		name string
		want bool // just check A > 0 for known colors
	}{
		{"white", true},
		{"black", true},
		{"red", true},
		{"green", true},
		{"blue", true},
		{"transparent", false},
		{"unknown", false},
	}
	for _, tt := range tests {
		c := parseColor(tt.name)
		if tt.want && c.A == 0 {
			t.Errorf("parseColor(%q): expected non-transparent", tt.name)
		}
		if !tt.want && c.A != 0 {
			t.Errorf("parseColor(%q): expected transparent or zero-alpha", tt.name)
		}
	}
}

func TestParseColorHex(t *testing.T) {
	c := parseColor("#ff0000")
	if c.R < 0.9 || c.A < 0.9 {
		t.Errorf("expected red from hex, got %+v", c)
	}
}

func TestApplyTextStyleColor(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<span style="color: blue; font-size: 24px">blue text</span>`)
	if len(root.Children()) != 1 {
		t.Errorf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestReadAttrValueUnquoted(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<input value=hello/>`)
	if len(root.Children()) != 1 {
		t.Errorf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestReadAttrValueSingleQuote(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<input value='world'/>`)
	if len(root.Children()) != 1 {
		t.Errorf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestLoadHTMLDivBackgroundShorthand(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<div style="background: red">bg</div>`)
	if len(root.Children()) != 1 {
		t.Errorf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestLoadHTMLMissingClosingTag(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	// Missing closing tag should not panic
	root := LoadHTML(tree, cfg, `<p>unclosed`)
	_ = root
}

func TestLoadHTMLH2H4H5(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<h2>Two</h2><h4>Four</h4><h5>Five</h5>`)
	if len(root.Children()) != 3 {
		t.Errorf("expected 3 children, got %d", len(root.Children()))
	}
}

// --- HTML+CSS integration tests ---

func TestLoadHTMLStyleBlock(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	html := `
	<style>
		.container { display: flex; gap: 12px; }
	</style>
	<div class="container">
		<p>Hello</p>
	</div>`
	root := LoadHTML(tree, cfg, html)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestLoadHTMLWithCSS(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	html := `<div class="panel"><p>Content</p></div>`
	cssText := `.panel { display: flex; padding: 16px; }`
	root := LoadHTMLWithCSS(tree, cfg, html, cssText)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestLoadHTMLWithCSSVariables(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	html := `<div class="box">Text</div>`
	cssText := `
		:root { --bg: #ff0000; }
		.box { background-color: var(--bg); }
	`
	root := LoadHTMLWithCSS(tree, cfg, html, cssText)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestLoadHTMLAnchorTag(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<a href="https://example.com">Link</a>`)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child (link), got %d", len(root.Children()))
	}
}

func TestLoadHTMLSelectTag(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<select>options</select>`)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child (select), got %d", len(root.Children()))
	}
}

func TestLoadHTMLTextareaTag(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<textarea>initial text</textarea>`)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child (textarea), got %d", len(root.Children()))
	}
}

func TestLoadHTMLIDAttribute(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	html := `<div id="main">content</div>`
	cssText := `#main { padding: 20px; }`
	root := LoadHTMLWithCSS(tree, cfg, html, cssText)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestLoadHTMLMultipleStyleBlocks(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	html := `
	<style>.a { color: red; }</style>
	<style>.b { color: blue; }</style>
	<div class="a">A</div>
	<div class="b">B</div>`
	root := LoadHTML(tree, cfg, html)
	if len(root.Children()) != 2 {
		t.Fatalf("expected 2 children, got %d", len(root.Children()))
	}
}

// === Additional coverage tests ===

func TestLoadHTMLWithCSSNilConfig(t *testing.T) {
	tree := core.NewTree()
	root := LoadHTMLWithCSS(tree, nil, `<p>test</p>`, `p { color: red; }`)
	if root == nil {
		t.Fatal("expected root")
	}
}

func TestApplyVisualPropsColor(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	html := `<span>hello</span>`
	cssText := `span { color: #ff0000; }`
	root := LoadHTMLWithCSS(tree, cfg, html, cssText)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children()))
	}
	txt, ok := root.Children()[0].(*widget.Text)
	if !ok {
		t.Fatal("expected Text widget")
	}
	c := txt.Color()
	if c.R < 0.9 {
		t.Errorf("expected red text color, got R=%f", c.R)
	}
}

func TestApplyVisualPropsBackgroundColor(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	html := `<div class="bg">content</div>`
	cssText := `.bg { background-color: #00ff00; }`
	root := LoadHTMLWithCSS(tree, cfg, html, cssText)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestApplyVisualPropsFontSize(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	html := `<span>big</span>`
	cssText := `span { font-size: 24px; }`
	root := LoadHTMLWithCSS(tree, cfg, html, cssText)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children()))
	}
	txt, ok := root.Children()[0].(*widget.Text)
	if !ok {
		t.Fatal("expected Text widget")
	}
	if txt.FontSize() != 24 {
		t.Errorf("expected font-size 24, got %f", txt.FontSize())
	}
}

func TestApplyVisualPropsBorderRadius(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	html := `<div class="r">content</div>`
	cssText := `.r { border-radius: 8px; }`
	root := LoadHTMLWithCSS(tree, cfg, html, cssText)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestApplyVisualPropsBorderColor(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	html := `<div class="bc">content</div>`
	cssText := `.bc { border-color: #333333; }`
	root := LoadHTMLWithCSS(tree, cfg, html, cssText)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestExtractStyleBlockUnterminatedStyle(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	// <style> without </style>
	root := LoadHTML(tree, cfg, `<style>.a { color: red; }<div>hi</div>`)
	_ = root // no panic
}

func TestExtractStyleBlockNoCloseAngle(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	// <style without closing >
	root := LoadHTML(tree, cfg, `<style .a { color: red; }</style>`)
	_ = root // no panic
}

func TestExtractStyleBlockMergeWithExternalCSS(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	html := `<style>.a { color: red; }</style><div class="a b">text</div>`
	cssText := `.b { font-size: 20px; }`
	root := LoadHTMLWithCSS(tree, cfg, html, cssText)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestParseInlineCSSMalformed(t *testing.T) {
	// Declaration without colon
	props := parseInlineCSS("nocolon; color: red")
	if props["color"] != "red" {
		t.Errorf("expected color=red after malformed entry")
	}
	if _, ok := props["nocolon"]; ok {
		t.Errorf("should not have nocolon key")
	}
}

func TestApplyInlineStyleNonEmpty(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	// Image with width/height via applyInlineStyle
	root := LoadHTML(tree, cfg, `<img src="x" style="width: 200px; height: 100px"/>`)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestReadAttrValueEOF(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	// Attribute value at EOF
	root := LoadHTML(tree, cfg, `<input value=`)
	_ = root // no panic
}

func TestParseColorAllBranches(t *testing.T) {
	// Hex
	c := parseColor("#abcdef")
	if c.A < 0.9 {
		t.Errorf("hex color should have full alpha")
	}
	// Named colors
	for _, name := range []string{"white", "black", "red", "green", "blue", "transparent"} {
		_ = parseColor(name)
	}
	// Unknown
	c = parseColor("notacolor")
	if c.A != 0 {
		t.Errorf("unknown color should be zero")
	}
}

func TestLoadHTMLStyleBlockWithAttrs(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	// <style type="text/css">
	html := `<style type="text/css">.x { color: red; }</style><span class="x">hi</span>`
	root := LoadHTML(tree, cfg, html)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestLoadHTMLNoRules(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	// CSS with no matching rules — applyCSS runs but finds no matches
	html := `<div class="nomatch">text</div>`
	cssText := `.other { color: red; }`
	root := LoadHTMLWithCSS(tree, cfg, html, cssText)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestApplyVisualPropsInvalidColor(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	// CSS color that doesn't parse
	html := `<span>text</span>`
	cssText := `span { color: notacolor; }`
	root := LoadHTMLWithCSS(tree, cfg, html, cssText)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestApplyVisualPropsFontSizeZero(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	// font-size: 0 should not apply (size <= 0)
	html := `<span>text</span>`
	cssText := `span { font-size: 0; }`
	root := LoadHTMLWithCSS(tree, cfg, html, cssText)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestApplyVisualPropsOnButton(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	// Visual props applied to non-div/non-text widget — should be no-op but not crash
	html := `<button>btn</button>`
	cssText := `button { color: red; background-color: blue; border-radius: 4px; border-color: green; font-size: 16px; }`
	root := LoadHTMLWithCSS(tree, cfg, html, cssText)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestApplyInlineStyleUnknownProp(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	// Unknown inline style prop — no crash
	root := LoadHTML(tree, cfg, `<img src="x" style="margin: 10px"/>`)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestParseChildrenEmptyInput(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, "   ")
	if len(root.Children()) != 0 {
		t.Errorf("expected 0 children for whitespace input, got %d", len(root.Children()))
	}
}

// === Document and new tag tests ===

func TestLoadHTMLDocument(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<div id="main" class="box"><p>hi</p></div>`, "")
	if doc.Root == nil {
		t.Fatal("expected root")
	}
	w := doc.QueryByID("main")
	if w == nil {
		t.Fatal("expected widget with id=main")
	}
	ws := doc.QueryByClass("box")
	if len(ws) != 1 {
		t.Errorf("expected 1 widget with class=box, got %d", len(ws))
	}
	divs := doc.QueryByTag("div")
	if len(divs) != 1 {
		t.Errorf("expected 1 div tag, got %d", len(divs))
	}
}

func TestLoadHTMLDocumentNilConfig(t *testing.T) {
	tree := core.NewTree()
	doc := LoadHTMLDocument(tree, nil, `<p>test</p>`, "")
	if doc.Root == nil {
		t.Fatal("expected root")
	}
}

func TestHTMLHeader(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<header height="56"><span>Title</span></header>`, "")
	headers := doc.QueryByTag("header")
	if len(headers) != 1 {
		t.Fatalf("expected 1 header, got %d", len(headers))
	}
	h, ok := headers[0].(*widget.Header)
	if !ok {
		t.Fatal("expected *widget.Header")
	}
	if len(h.Children()) != 1 {
		t.Errorf("expected 1 child in header, got %d", len(h.Children()))
	}
}

func TestHTMLFooter(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<footer><span>Footer</span></footer>`, "")
	if len(doc.QueryByTag("footer")) != 1 {
		t.Fatal("expected 1 footer")
	}
}

func TestHTMLAside(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<aside width="200"><button>Menu</button></aside>`, "")
	asides := doc.QueryByTag("aside")
	if len(asides) != 1 {
		t.Fatal("expected 1 aside")
	}
}

func TestHTMLMain(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<main id="content"><span>Content</span></main>`, "")
	c := doc.QueryByID("content")
	if c == nil {
		t.Fatal("expected content widget")
	}
	if _, ok := c.(*widget.Content); !ok {
		t.Fatal("expected *widget.Content")
	}
}

func TestHTMLLayout(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<layout><header height="40"><span>H</span></header></layout>`, "")
	layouts := doc.QueryByTag("layout")
	if len(layouts) != 1 {
		t.Fatal("expected 1 layout")
	}
}

func TestHTMLSpace(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<space gap="12"><button>A</button><button>B</button></space>`, "")
	spaces := doc.QueryByTag("space")
	if len(spaces) != 1 {
		t.Fatal("expected 1 space")
	}
}

func TestHTMLRowCol(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<row gutter="16"><col span="12"><span>Left</span></col><col span="12"><span>Right</span></col></row>`, "")
	rows := doc.QueryByTag("row")
	if len(rows) != 1 {
		t.Fatal("expected 1 row")
	}
	cols := doc.QueryByTag("col")
	if len(cols) != 2 {
		t.Fatalf("expected 2 cols, got %d", len(cols))
	}
}

func TestHTMLCheckbox(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<checkbox id="cb" checked>Option</checkbox>`, "")
	w := doc.QueryByID("cb")
	if w == nil {
		t.Fatal("expected checkbox")
	}
	cb, ok := w.(*widget.Checkbox)
	if !ok {
		t.Fatal("expected *widget.Checkbox")
	}
	if !cb.IsChecked() {
		t.Error("expected checked=true")
	}
}

func TestHTMLSwitch(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<switch checked></switch>`, "")
	sws := doc.QueryByTag("switch")
	if len(sws) != 1 {
		t.Fatal("expected 1 switch")
	}
}

func TestHTMLRadio(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<radio group="g" checked>A</radio><radio group="g">B</radio>`, "")
	radios := doc.QueryByTag("radio")
	if len(radios) != 2 {
		t.Fatalf("expected 2 radios, got %d", len(radios))
	}
}

func TestHTMLTag(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<tag type="success">OK</tag><tag type="warning">Warn</tag><tag type="error">Err</tag><tag type="processing">...</tag><tag>Default</tag>`, "")
	tags := doc.QueryByTag("tag")
	if len(tags) != 5 {
		t.Fatalf("expected 5 tags, got %d", len(tags))
	}
}

func TestHTMLProgress(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<progress percent="75"></progress>`, "")
	progs := doc.QueryByTag("progress")
	if len(progs) != 1 {
		t.Fatal("expected 1 progress")
	}
}

func TestHTMLMessage(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<message type="success">Done</message><message type="warning">Warn</message><message type="error">Err</message><message>Info</message>`, "")
	msgs := doc.QueryByTag("message")
	if len(msgs) != 4 {
		t.Fatalf("expected 4 messages, got %d", len(msgs))
	}
}

func TestHTMLEmpty(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<empty></empty>`, "")
	if len(doc.QueryByTag("empty")) != 1 {
		t.Fatal("expected 1 empty")
	}
}

func TestHTMLLoading(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<loading tip="Loading..."></loading>`, "")
	if len(doc.QueryByTag("loading")) != 1 {
		t.Fatal("expected 1 loading")
	}
}

func TestHTMLTooltip(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<button>Hover</button><tooltip>Tip text</tooltip>`)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child (button), got %d", len(root.Children()))
	}
}

func TestHTMLComment(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<!-- comment --><p>text</p>`)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestHTMLButtonVariants(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `
		<button id="b1" variant="secondary">S</button>
		<button id="b2" variant="text">T</button>
		<button id="b3" variant="link">L</button>
	`, "")
	if b, ok := doc.QueryByID("b1").(*widget.Button); ok {
		if b.Variant() != widget.ButtonSecondary {
			t.Errorf("expected secondary variant")
		}
	}
	if b, ok := doc.QueryByID("b2").(*widget.Button); ok {
		if b.Variant() != widget.ButtonText {
			t.Errorf("expected text variant")
		}
	}
	if b, ok := doc.QueryByID("b3").(*widget.Button); ok {
		if b.Variant() != widget.ButtonLink {
			t.Errorf("expected link variant")
		}
	}
}

func TestHTMLTextareaAttrs(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<textarea id="ta" placeholder="Enter..." rows="5">initial</textarea>`, "")
	w := doc.QueryByID("ta")
	if w == nil {
		t.Fatal("expected textarea")
	}
}

func TestHTMLCSSOnNewTags(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `
		<layout>
			<header height="50"><span class="title">Test</span></header>
			<main><span>Body</span></main>
			<footer><span>Foot</span></footer>
		</layout>
	`, `
		header { background-color: #001529; }
		.title { color: white; font-size: 20px; }
		footer { background-color: #333; }
	`)
	if doc.Root == nil {
		t.Fatal("expected root")
	}
	titles := doc.QueryByClass("title")
	if len(titles) != 1 {
		t.Fatalf("expected 1 title, got %d", len(titles))
	}
}

func TestDocumentQueryByIDNotFound(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<p>test</p>`, "")
	if doc.QueryByID("nonexistent") != nil {
		t.Error("expected nil for nonexistent ID")
	}
}

func TestDocumentQueryByClassEmpty(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<p>test</p>`, "")
	if len(doc.QueryByClass("nonexistent")) != 0 {
		t.Error("expected empty for nonexistent class")
	}
}

func TestHTMLCheckboxDisabled(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<checkbox disabled>Dis</checkbox>`, "")
	cbs := doc.QueryByTag("checkbox")
	if len(cbs) != 1 {
		t.Fatal("expected 1 checkbox")
	}
}

func TestHTMLSwitchDisabled(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<switch disabled></switch>`, "")
	sws := doc.QueryByTag("switch")
	if len(sws) != 1 {
		t.Fatal("expected 1 switch")
	}
}

// === Feature: Event Delegation (doc.On / doc.QueryAll) ===

func TestDocumentQueryAll(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `
		<button id="b1" class="btn">A</button>
		<button id="b2" class="btn">B</button>
		<p>text</p>
	`, "")

	// By ID
	ws := doc.QueryAll("#b1")
	if len(ws) != 1 {
		t.Fatalf("QueryAll(#b1): expected 1, got %d", len(ws))
	}

	// By class
	ws = doc.QueryAll(".btn")
	if len(ws) != 2 {
		t.Fatalf("QueryAll(.btn): expected 2, got %d", len(ws))
	}

	// By tag
	ws = doc.QueryAll("button")
	if len(ws) != 2 {
		t.Fatalf("QueryAll(button): expected 2, got %d", len(ws))
	}

	// Not found
	ws = doc.QueryAll("#nonexistent")
	if len(ws) != 0 {
		t.Fatalf("QueryAll(#nonexistent): expected 0, got %d", len(ws))
	}

	// Empty selector
	ws = doc.QueryAll("")
	if len(ws) != 0 {
		t.Fatalf("QueryAll(empty): expected 0, got %d", len(ws))
	}

	// By class not found
	ws = doc.QueryAll(".nope")
	if len(ws) != 0 {
		t.Fatalf("QueryAll(.nope): expected 0, got %d", len(ws))
	}
}

func TestDocumentOnClick(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `
		<button id="b1" class="btn">A</button>
		<button id="b2" class="btn">B</button>
	`, "")

	clicked := 0
	doc.On(".btn", "click", func(w widget.Widget) {
		clicked++
	})

	// Verify handlers were registered (buttons have OnClick)
	// We can't easily simulate clicks here, but we can verify no panic
	_ = clicked
}

func TestDocumentOnChange(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `
		<input id="inp" class="field"/>
		<textarea id="ta" class="field">text</textarea>
		<select id="sel" class="field">opts</select>
	`, "")

	doc.On(".field", "change", func(w widget.Widget) {})
	// No panic = success
}

func TestDocumentOnToggle(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `
		<checkbox id="cb" class="toggle">Check</checkbox>
		<switch id="sw" class="toggle"></switch>
	`, "")

	doc.On(".toggle", "toggle", func(w widget.Widget) {})
	// No panic = success
}

// === Feature: Dispose ===

func TestDocumentDispose(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<div id="main"><p>Hello</p></div>`, "")

	if doc.Root == nil {
		t.Fatal("expected root before dispose")
	}

	cleanupCalled := false
	doc.addCleanup(func() { cleanupCalled = true })

	doc.Dispose()

	if !cleanupCalled {
		t.Error("expected cleanup to be called")
	}
	if doc.Root != nil {
		t.Error("expected root to be nil after dispose")
	}
	if len(doc.cleanups) != 0 {
		t.Error("expected cleanups to be cleared")
	}
	if len(doc.bindings) != 0 {
		t.Error("expected bindings to be cleared")
	}
}

func TestDocumentDisposeMultipleCleanups(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<p>test</p>`, "")

	count := 0
	doc.addCleanup(func() { count++ })
	doc.addCleanup(func() { count++ })
	doc.addCleanup(func() { count++ })

	doc.Dispose()
	if count != 3 {
		t.Errorf("expected 3 cleanups called, got %d", count)
	}
}

// === Feature: Data Binding & Templates ===

func TestDocumentSetGetData(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<p>test</p>`, "")

	// Get before set
	if doc.GetData("key") != nil {
		t.Error("expected nil for unset key")
	}

	doc.SetData("name", "Alice")
	if doc.GetData("name") != "Alice" {
		t.Errorf("expected Alice, got %v", doc.GetData("name"))
	}

	doc.SetData("count", 42)
	if doc.GetData("count") != 42 {
		t.Errorf("expected 42, got %v", doc.GetData("count"))
	}
}

func TestDocumentTemplateInterpolation(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<p>Hello {{name}}</p>`, "")

	// Before data set, template key resolves to empty
	p := doc.QueryByTag("p")
	if len(p) != 1 {
		t.Fatalf("expected 1 p, got %d", len(p))
	}
	txt, ok := p[0].(*widget.Text)
	if !ok {
		t.Fatal("expected *widget.Text")
	}
	if txt.Text() != "Hello " {
		t.Errorf("expected 'Hello ', got %q", txt.Text())
	}

	// After SetData, binding should update text
	doc.SetData("name", "World")
	if txt.Text() != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", txt.Text())
	}
}

func TestDocumentTemplateMultipleKeys(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<span>{{first}} {{last}}</span>`, "")

	doc.SetData("first", "Jane")
	doc.SetData("last", "Doe")

	spans := doc.QueryByTag("span")
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	txt := spans[0].(*widget.Text)
	if txt.Text() != "Jane Doe" {
		t.Errorf("expected 'Jane Doe', got %q", txt.Text())
	}
}

func TestDocumentTemplateBareText(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<div>Count: {{count}}</div>`, "")

	doc.SetData("count", 5)

	// The bare text inside div should be interpolated
	div := doc.Root.Children()[0]
	if len(div.Children()) != 1 {
		t.Fatalf("expected 1 child in div, got %d", len(div.Children()))
	}
	txt := div.Children()[0].(*widget.Text)
	if txt.Text() != "Count: 5" {
		t.Errorf("expected 'Count: 5', got %q", txt.Text())
	}
}

func TestDocumentTemplateHeading(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<h1>Welcome {{user}}</h1>`, "")

	doc.SetData("user", "Admin")

	h := doc.QueryByTag("h1")
	if len(h) != 1 {
		t.Fatalf("expected 1 h1, got %d", len(h))
	}
	txt := h[0].(*widget.Text)
	if txt.Text() != "Welcome Admin" {
		t.Errorf("expected 'Welcome Admin', got %q", txt.Text())
	}
}

func TestDocumentDataIf(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<div id="msg" data-if="visible"><p>Hello</p></div>`, "")

	// Initially hidden (no data set)
	w := doc.QueryByID("msg")
	if w == nil {
		t.Fatal("expected widget")
	}
	if w.Style().Display != layout.DisplayNone {
		t.Error("expected display:none when data-if key is not set")
	}

	// Set to true
	doc.SetData("visible", true)
	if w.Style().Display == layout.DisplayNone {
		t.Error("expected visible after SetData(visible, true)")
	}

	// Set to false
	doc.SetData("visible", false)
	if w.Style().Display != layout.DisplayNone {
		t.Error("expected hidden after SetData(visible, false)")
	}
}

func TestDocumentDataModel(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<input id="name" data-model="username"/>`, "")

	// SetData should push value to input
	doc.SetData("username", "alice")
	w := doc.QueryByID("name")
	inp, ok := w.(*widget.Input)
	if !ok {
		t.Fatal("expected *widget.Input")
	}
	if inp.Value() != "alice" {
		t.Errorf("expected 'alice', got %q", inp.Value())
	}
}

func TestDocumentDataModelTextArea(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<textarea id="bio" data-model="bio">init</textarea>`, "")

	doc.SetData("bio", "Hello World")
	w := doc.QueryByID("bio")
	ta, ok := w.(*widget.TextArea)
	if !ok {
		t.Fatal("expected *widget.TextArea")
	}
	if ta.Value() != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", ta.Value())
	}
}

func TestExtractTemplateKeys(t *testing.T) {
	keys := extractTemplateKeys("Hello {{name}}, you have {{count}} items")
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	if keys[0] != "name" || keys[1] != "count" {
		t.Errorf("expected [name, count], got %v", keys)
	}

	// Duplicate keys
	keys = extractTemplateKeys("{{x}} and {{x}}")
	if len(keys) != 1 {
		t.Errorf("expected 1 unique key, got %d", len(keys))
	}

	// No keys
	keys = extractTemplateKeys("no templates here")
	if len(keys) != 0 {
		t.Errorf("expected 0 keys, got %d", len(keys))
	}

	// Unclosed template
	keys = extractTemplateKeys("{{unclosed")
	if len(keys) != 0 {
		t.Errorf("expected 0 keys for unclosed, got %d", len(keys))
	}
}

func TestIsTruthy(t *testing.T) {
	tests := []struct {
		val  interface{}
		want bool
	}{
		{nil, false},
		{true, true},
		{false, false},
		{0, false},
		{1, true},
		{0.0, false},
		{1.5, true},
		{"", false},
		{"hello", true},
		{[]int{1}, true},
	}
	for _, tt := range tests {
		got := isTruthy(tt.val)
		if got != tt.want {
			t.Errorf("isTruthy(%v) = %v, want %v", tt.val, got, tt.want)
		}
	}
}

func TestDocumentInterpolateNoData(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<p>test</p>`, "")

	result := doc.interpolate("Hello {{missing}}")
	if result != "Hello " {
		t.Errorf("expected 'Hello ', got %q", result)
	}
}

func TestDocumentGetDataNilMap(t *testing.T) {
	doc := &Document{}
	if doc.GetData("key") != nil {
		t.Error("expected nil from nil data map")
	}
}

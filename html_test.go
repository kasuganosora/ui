package ui

import (
	"testing"

	"github.com/kasuganosora/ui/core"
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

func TestLoadHTMLText(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, "<p>Hello World</p>")
	if len(root.Children()) != 1 {
		t.Errorf("expected 1 child, got %d", len(root.Children()))
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

func TestLoadHTMLInput(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<input placeholder="name" value="test"/>`)
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestLoadHTMLHeadings(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := LoadHTML(tree, cfg, `<h1>Title</h1><h3>Subtitle</h3>`)
	if len(root.Children()) != 2 {
		t.Errorf("expected 2 children, got %d", len(root.Children()))
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

func TestParsePx(t *testing.T) {
	tests := []struct{ in string; want float32 }{
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

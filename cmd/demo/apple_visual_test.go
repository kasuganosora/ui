package main

import (
	"strings"
	"testing"

	ui "github.com/kasuganosora/ui"
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/css"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// cssRule returns the body of a single CSS rule, so the test can reason about
// one selector instead of matching strings anywhere in the stylesheet.
func cssRule(sheet, selector string) string {
	start := strings.Index(sheet, selector+" {")
	if start < 0 {
		return ""
	}
	rest := sheet[start+len(selector)+2:]
	end := strings.Index(rest, "}")
	if end < 0 {
		return rest
	}
	return rest[:end]
}

func TestDemoCardUsesMaterialDepth(t *testing.T) {
	card := cssRule(demoHTML, ".demo-card")
	if card == "" {
		t.Fatal(".demo-card rule missing from demo stylesheet")
	}
	if strings.Contains(card, "border-width: 1px") {
		t.Errorf("demo card still has a hard 1px outline: %s", card)
	}
	if !strings.Contains(card, "box-shadow:") {
		t.Fatalf("demo card has no material shadow: %s", card)
	}

	layers := css.ParseBoxShadow("0px 2px 8px rgba(0,0,0,0.08)")
	if len(layers) != 1 {
		t.Fatalf("box-shadow value is not understood by the CSS engine: %+v", layers)
	}
	if layers[0].Blur <= 0 || layers[0].Color.A <= 0 {
		t.Errorf("card shadow is not a soft translucent layer: %+v", layers[0])
	}
}

func TestDemoSecondaryTextIsReadable(t *testing.T) {
	if !strings.Contains(demoHTML, "--text-body: rgba(0,0,0,0.7)") {
		t.Error("body copy token fell back to a weaker contrast value")
	}
	// Faint hardcoded captions must not creep back into the stylesheet; the
	// sidebar group headers are drawn by widget.List from the semantic token.
	if strings.Contains(demoHTML, "rgba(0,0,0,0.4)") {
		t.Error("stylesheet still contains a low-opacity black caption color")
	}
}

// TestDemoCardShadowIsRendered verifies the full CSS -> Div -> command buffer
// chain. A stylesheet string alone is not evidence that the shadow reaches
// the renderer; class selectors and the HTML style applier both have to work.
func TestDemoCardShadowIsRendered(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := ui.LoadHTMLDocument(tree, cfg, demoHTML, "")

	var root widget.Widget = doc.Root
	if len(doc.Root.Children()) > 0 {
		root = doc.Root.Children()[0]
	}
	tree.SetLayout(tree.Root(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 1280, 800),
	})
	ui.AutoLayout(tree, root, 1280, 800)

	buf := render.NewCommandBuffer()
	root.Draw(buf)

	shadows := 0
	for _, c := range append(buf.Commands(), buf.Overlays()...) {
		if c.Type != render.CmdShadow || c.Shadow == nil || c.Shadow.BlurRadius <= 0 || c.Shadow.Color.A <= 0 {
			continue
		}
		shadows++
		// Backends stable-sort the whole frame by ZOrder. A negative z puts
		// the shadow in front of every z=0 opaque ancestor background, which
		// then paints over it — the command exists but nothing is visible.
		if c.ZOrder < 0 {
			t.Fatalf("shadow at z=%d would be hidden by z=0 ancestor backgrounds", c.ZOrder)
		}
	}
	if shadows == 0 {
		t.Fatal("no blurred box-shadow reached the command buffer; the card material is not rendered")
	}
}

package main

import (
	"testing"

	ui "github.com/kasuganosora/ui"
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// TestBuildUI validates that the demo UI tree can be built and drawn without crashing.
func TestBuildUI(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()

	doc := ui.LoadHTMLDocument(tree, cfg, demoHTML, "")
	root := doc.Root
	if root == nil {
		t.Fatal("LoadHTMLDocument returned nil root")
	}

	// Run layout so widgets have bounds, then draw
	ui.AutoLayout(tree, root, 1280, 800)
	buf := render.NewCommandBuffer()
	root.Draw(buf)

	if buf.Len() == 0 {
		t.Error("expected draw commands from the demo UI")
	}
	t.Logf("demo UI generated %d render commands", buf.Len())
}

// TestBuildUIWidgetCount validates a reasonable number of widgets exist.
func TestBuildUIWidgetCount(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	ui.LoadHTMLDocument(tree, cfg, demoHTML, "")

	count := tree.ElementCount()
	if count < 20 {
		t.Errorf("expected at least 20 elements, got %d", count)
	}
	t.Logf("total elements: %d", count)
}

// TestCardChildBelowHeader verifies Card body children are positioned below the header.
func TestCardChildBelowHeader(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()

	doc := ui.LoadHTMLDocument(tree, cfg, demoHTML, "")
	root := doc.Root
	if root == nil {
		t.Fatal("LoadHTMLDocument returned nil root")
	}

	// Setup programmatic widgets (creates the Card in demo-card container)
	setupDemoWidgets(doc, tree, cfg)

	// Make the Card section visible (by default only Button section is shown)
	secCard := doc.QueryByID("sec-card")
	if secCard != nil {
		s := secCard.Style()
		s.Display = layout.DisplayBlock
		secCard.SetStyle(s)
	}

	// Hide other sections to avoid conflicts
	secButton := doc.QueryByID("sec-button")
	if secButton != nil {
		s := secButton.Style()
		s.Display = layout.DisplayNone
		secButton.SetStyle(s)
	}

	// Use Layout widget (first child) as root, matching real app behavior
	layoutRoot := root.Children()[0]
	ui.AutoLayout(tree, layoutRoot, 1280, 800)

	// Find the Card widget
	demoCardContainer := doc.QueryByID("demo-card")
	if demoCardContainer == nil {
		t.Fatal("demo-card container not found")
	}

	var card *widget.Card
	var bodyText *widget.Text
	for _, child := range demoCardContainer.Children() {
		if c, ok := child.(*widget.Card); ok && c.Title() == "卡片标题" {
			card = c
			for _, cc := range c.Children() {
				if txt, ok := cc.(*widget.Text); ok {
					bodyText = txt
					break
				}
			}
			break
		}
	}

	if card == nil {
		t.Fatal("Card widget not found")
	}
	if bodyText == nil {
		t.Fatal("Card body text not found")
	}

	cardBounds := card.Bounds()
	textBounds := bodyText.Bounds()

	t.Logf("card bounds: %+v", cardBounds)
	t.Logf("text bounds: %+v", textBounds)

	// Card header is 48px. Body text must start below header.
	headerBottom := cardBounds.Y + 48
	if textBounds.Y < headerBottom {
		t.Errorf("body text Y (%.1f) overlaps with card header bottom (%.1f)",
			textBounds.Y, headerBottom)
	}
}

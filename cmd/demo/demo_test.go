//go:build windows

package main

import (
	"testing"

	ui "github.com/kasuganosora/ui"
	"github.com/kasuganosora/ui/core"
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
	ui.AutoLayout(tree, root, 1280, 720)
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

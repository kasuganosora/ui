//go:build windows

package main

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// TestBuildUI validates that the demo UI tree can be built and drawn without crashing.
func TestBuildUI(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()

	root := buildUI(tree, cfg)
	if root == nil {
		t.Fatal("buildUI returned nil")
	}

	// Simulate layout
	computeLayout(tree, root, 960, 640)

	// Draw into a command buffer
	buf := render.NewCommandBuffer()
	root.Draw(buf)

	if buf.Len() == 0 {
		t.Error("expected draw commands from the demo UI")
	}
	t.Logf("demo UI generated %d render commands", buf.Len())
}

// TestBuildUIResize validates that the layout can handle resize.
func TestBuildUIResize(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := buildUI(tree, cfg)

	sizes := [][2]float32{{800, 600}, {1920, 1080}, {320, 240}}
	for _, s := range sizes {
		computeLayout(tree, root, s[0], s[1])
		buf := render.NewCommandBuffer()
		root.Draw(buf)
		if buf.Len() == 0 {
			t.Errorf("no commands at %gx%g", s[0], s[1])
		}
	}
}

// TestBuildUIWidgetCount validates a reasonable number of widgets exist.
func TestBuildUIWidgetCount(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	buildUI(tree, cfg)

	count := tree.ElementCount()
	// Root + Layout + Header + Title + Body + Aside + 4 menu items + Content + ...
	if count < 20 {
		t.Errorf("expected at least 20 elements, got %d", count)
	}
	t.Logf("total elements: %d", count)
}

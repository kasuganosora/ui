package ui

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

func TestLayerManager(t *testing.T) {
	lm := NewLayerManager()

	base := lm.AddLayer("base", LayerBase, 0)
	hud := lm.AddLayer("hud", LayerHUD, 10)
	dialog := lm.AddLayer("dialog", LayerDialog, 20)

	if base == nil || hud == nil || dialog == nil {
		t.Fatal("expected non-nil layers")
	}

	// GetLayer
	got := lm.GetLayer("hud")
	if got != hud {
		t.Error("GetLayer should return the hud layer")
	}
	if lm.GetLayer("nonexistent") != nil {
		t.Error("GetLayer should return nil for unknown layer")
	}
}

func TestLayerAddWidget(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	lm := NewLayerManager()
	layer := lm.AddLayer("hud", LayerHUD, 10)

	btn := widget.NewButton(tree, "test", cfg)
	layer.AddWidget(btn)

	if len(layer.Widgets) != 1 {
		t.Errorf("expected 1 widget, got %d", len(layer.Widgets))
	}

	layer.RemoveWidget(btn)
	if len(layer.Widgets) != 0 {
		t.Errorf("expected 0 widgets after remove, got %d", len(layer.Widgets))
	}
}

func TestLayerDraw(t *testing.T) {
	lm := NewLayerManager()
	tree := core.NewTree()
	cfg := widget.DefaultConfig()

	layer := lm.AddLayer("hud", LayerHUD, 10)
	btn := widget.NewButton(tree, "test", cfg)
	tree.SetLayout(btn.ElementID(), core.LayoutResult{
		Bounds: rect(0, 0, 100, 32),
	})
	layer.AddWidget(btn)

	buf := render.NewCommandBuffer()
	lm.Draw(buf)

	if buf.Len() == 0 {
		t.Error("expected render commands from layer draw")
	}
}

func TestLayerVisibility(t *testing.T) {
	lm := NewLayerManager()
	tree := core.NewTree()
	cfg := widget.DefaultConfig()

	layer := lm.AddLayer("hud", LayerHUD, 10)
	btn := widget.NewButton(tree, "test", cfg)
	tree.SetLayout(btn.ElementID(), core.LayoutResult{
		Bounds: rect(0, 0, 100, 32),
	})
	layer.AddWidget(btn)

	layer.Visible = false
	buf := render.NewCommandBuffer()
	lm.Draw(buf)
	if buf.Len() != 0 {
		t.Error("expected no commands when layer is hidden")
	}
}

func TestLayerRemove(t *testing.T) {
	lm := NewLayerManager()
	lm.AddLayer("a", LayerBase, 0)
	lm.AddLayer("b", LayerHUD, 10)

	lm.RemoveLayer("a")
	if lm.GetLayer("a") != nil {
		t.Error("layer 'a' should be removed")
	}
	if lm.GetLayer("b") == nil {
		t.Error("layer 'b' should still exist")
	}
}

func TestLayerClear(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	lm := NewLayerManager()
	layer := lm.AddLayer("test", LayerBase, 0)

	layer.AddWidget(widget.NewButton(tree, "a", cfg))
	layer.AddWidget(widget.NewButton(tree, "b", cfg))
	layer.Clear()

	if len(layer.Widgets) != 0 {
		t.Errorf("expected 0 widgets after clear, got %d", len(layer.Widgets))
	}
}

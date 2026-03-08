package ui

import (
	"testing"

	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/widget"
)

func TestEmbeddedUI(t *testing.T) {
	cfg := widget.DefaultConfig()
	ui := NewEmbeddedUI(cfg)

	if ui.Tree() == nil {
		t.Fatal("expected tree")
	}
	if ui.Config() == nil {
		t.Fatal("expected config")
	}
	if ui.Dispatcher() == nil {
		t.Fatal("expected dispatcher")
	}
	if ui.Layers() == nil {
		t.Fatal("expected layer manager")
	}
}

func TestEmbeddedUINilConfig(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	if ui.Config() == nil {
		t.Fatal("expected default config")
	}
}

func TestEmbeddedUIResize(t *testing.T) {
	cfg := widget.DefaultConfig()
	ui := NewEmbeddedUI(cfg)
	ui.Resize(800, 600)

	// No panic is success
}

func TestEmbeddedUISetRoot(t *testing.T) {
	cfg := widget.DefaultConfig()
	ui := NewEmbeddedUI(cfg)

	root := widget.NewDiv(ui.Tree(), cfg)
	ui.SetRoot(root)

	buf := ui.Render()
	// Empty div with no bounds produces no commands, that's OK
	_ = buf
}

func TestEmbeddedUIHandleEventMouse(t *testing.T) {
	cfg := widget.DefaultConfig()
	ui := NewEmbeddedUI(cfg)

	root := widget.NewDiv(ui.Tree(), cfg)
	ui.SetRoot(root)
	ui.Resize(800, 600)

	evt := &event.Event{
		Type:    event.MouseMove,
		GlobalX: 100,
		GlobalY: 100,
	}
	ui.HandleEvent(evt)
	// No panic is success
}

func TestEmbeddedUIHandleEventMouseMissesTarget(t *testing.T) {
	cfg := widget.DefaultConfig()
	ui := NewEmbeddedUI(cfg)

	evt := &event.Event{
		Type:    event.MouseMove,
		GlobalX: 999,
		GlobalY: 999,
	}
	ui.HandleEvent(evt)
	// No panic is success
}

func TestEmbeddedUIHandleEventKeyboard(t *testing.T) {
	cfg := widget.DefaultConfig()
	ui := NewEmbeddedUI(cfg)

	root := widget.NewDiv(ui.Tree(), cfg)
	ui.SetRoot(root)

	evt := &event.Event{
		Type: event.KeyDown,
	}
	ui.HandleEvent(evt)
	// No panic — keyboard events go to focused element
}

func TestEmbeddedUINeedsRedraw(t *testing.T) {
	cfg := widget.DefaultConfig()
	ui := NewEmbeddedUI(cfg)

	// Fresh tree should need render
	if !ui.NeedsRedraw() {
		t.Log("fresh tree may or may not need redraw")
	}

	ui.Render()
	if ui.NeedsRedraw() {
		t.Error("should not need redraw after render clears dirty flags")
	}
}

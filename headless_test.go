package ui

import (
	"encoding/json"
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

func TestHeadlessUI(t *testing.T) {
	cfg := widget.DefaultConfig()
	h := NewHeadlessUI(cfg)

	if h.Tree() == nil {
		t.Fatal("expected tree")
	}
	if h.Config() == nil {
		t.Fatal("expected config")
	}
}

func TestHeadlessUINilConfig(t *testing.T) {
	h := NewHeadlessUI(nil)
	if h.Config() == nil {
		t.Fatal("expected default config")
	}
}

func TestHeadlessUIRender(t *testing.T) {
	cfg := widget.DefaultConfig()
	h := NewHeadlessUI(cfg)

	root := widget.NewDiv(h.Tree(), cfg)
	root.SetBgColor(uimath.ColorHex("#ff0000"))
	h.SetRoot(root)

	// Need bounds for div to render
	h.Tree().SetLayout(root.ElementID(), newLayoutResult(0, 0, 200, 100))

	buf := h.Render()
	if buf.Len() == 0 {
		t.Error("expected render commands")
	}
}

func TestHeadlessUIResize(t *testing.T) {
	cfg := widget.DefaultConfig()
	h := NewHeadlessUI(cfg)
	h.Resize(800, 600)
	// No panic = success
}

func TestHeadlessUIHandleEventMouse(t *testing.T) {
	cfg := widget.DefaultConfig()
	h := NewHeadlessUI(cfg)

	root := widget.NewDiv(h.Tree(), cfg)
	h.SetRoot(root)
	h.Tree().SetLayout(root.ElementID(), newLayoutResult(0, 0, 200, 100))

	evt := &event.Event{
		Type:    event.MouseMove,
		GlobalX: 50,
		GlobalY: 50,
	}
	h.HandleEvent(evt)
	// No panic = success
}

func TestHeadlessUIHandleEventMissesTarget(t *testing.T) {
	cfg := widget.DefaultConfig()
	h := NewHeadlessUI(cfg)

	// No root set, mouse event should hit nothing
	evt := &event.Event{
		Type:    event.MouseMove,
		GlobalX: 999,
		GlobalY: 999,
	}
	h.HandleEvent(evt)
	// No panic = success
}

func TestHeadlessUIRenderEmpty(t *testing.T) {
	cfg := widget.DefaultConfig()
	h := NewHeadlessUI(cfg)
	// No root set
	buf := h.Render()
	if buf.Len() != 0 {
		t.Error("expected 0 commands with no root")
	}
}

func TestExportCommands(t *testing.T) {
	buf := render.NewCommandBuffer()
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(10, 20, 100, 50),
		FillColor: uimath.Color{R: 1, G: 0, B: 0, A: 1},
		Corners:   uimath.CornersAll(4),
	}, 1, 1)

	cmds := ExportCommands(buf)
	if len(cmds) != 1 {
		t.Fatalf("expected 1 exported command, got %d", len(cmds))
	}
	if cmds[0].Type != "rect" {
		t.Errorf("expected type 'rect', got %q", cmds[0].Type)
	}
	if cmds[0].X != 10 || cmds[0].Y != 20 {
		t.Errorf("expected x=10,y=20, got x=%g,y=%g", cmds[0].X, cmds[0].Y)
	}
}

func TestExportCommandsClip(t *testing.T) {
	buf := render.NewCommandBuffer()
	buf.PushClip(uimath.NewRect(0, 0, 100, 100))

	cmds := ExportCommands(buf)
	found := false
	for _, c := range cmds {
		if c.Type == "clip" {
			found = true
		}
	}
	if !found {
		t.Error("expected a clip command in exported commands")
	}
}

func TestExportJSON(t *testing.T) {
	buf := render.NewCommandBuffer()
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(0, 0, 50, 50),
		FillColor: uimath.ColorWhite,
	}, 0, 1)

	data, err := ExportJSON(buf)
	if err != nil {
		t.Fatalf("ExportJSON error: %v", err)
	}

	var cmds []ExportedCommand
	if err := json.Unmarshal(data, &cmds); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(cmds) != 1 {
		t.Errorf("expected 1 command, got %d", len(cmds))
	}
}

// helper
func newLayoutResult(x, y, w, h float32) core.LayoutResult {
	return core.LayoutResult{
		Bounds: rect(x, y, w, h),
	}
}

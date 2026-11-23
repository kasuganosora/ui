package widget

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

func topOverlayAt(buf *render.CommandBuffer, x, y float32) (render.Command, bool) {
	var (
		found bool
		best  int32
		cmd   render.Command
	)
	for _, item := range buf.Overlays() {
		if item.Rect == nil || !item.Rect.Bounds.Contains(uimath.Vec2{X: x, Y: y}) {
			continue
		}
		if !found || item.ZOrder >= best {
			found = true
			best = item.ZOrder
			cmd = item
		}
	}
	return cmd, found
}

func TestLaterOpenedDialogPaintsAboveEarlierDate(t *testing.T) {
	tree := core.NewTree()
	dlg := NewDialog(tree, "early", nil)
	tree.AppendChild(tree.Root(), dlg.ElementID())
	bounds(tree, dlg.ElementID(), 0, 0, 800, 600)
	dp := NewDatePicker(tree, nil)
	dp.SetDate(2026, 1, 1)
	tree.AppendChild(tree.Root(), dp.ElementID())
	bounds(tree, dp.ElementID(), 10, 10, 200, 32)
	dp.SetOpen(true)
	dlg.Open()
	buf := render.NewCommandBuffer()
	dp.Draw(buf)
	dlg.Draw(buf)
	top, ok := topOverlayAt(buf, 10+4*32+16, 10+32+4+36+2*32+16)
	if !ok || top.Rect.Bounds.Width < 800 || top.Rect.FillColor.A < 0.4 {
		t.Fatalf("dialog paint did not cover the date popup: ok %v cmd %+v", ok, top.Rect)
	}
}

func TestInnerDatePaintsAboveOwnDialog(t *testing.T) {
	tree := core.NewTree()
	dlg := NewDialog(tree, "host", nil)
	tree.AppendChild(tree.Root(), dlg.ElementID())
	bounds(tree, dlg.ElementID(), 0, 0, 800, 600)
	dp := NewDatePicker(tree, nil)
	dp.SetDate(2026, 1, 1)
	tree.AppendChild(dlg.ElementID(), dp.ElementID())
	bounds(tree, dp.ElementID(), 30, 40, 200, 32)
	dlg.Open()
	dp.SetOpen(true)
	buf := render.NewCommandBuffer()
	dlg.Draw(buf)
	dp.Draw(buf)
	top, ok := topOverlayAt(buf, 30+4*32+16, 40+32+4+36+2*32+16)
	if !ok || top.Rect.Bounds.Width > 300 {
		t.Fatalf("inner calendar is not the top overlay: ok %v bounds %+v", ok, top.Rect)
	}
}

package ui

import (
	"testing"

	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/godot"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

func TestModalEscapeSkipsCoveredHover(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(400, 300)
	in := widget.NewInput(ui.Tree(), ui.Config())
	card := widget.NewHoverCard(ui.Tree(), ui.Config())
	card.SetDelay(0.05)
	dlg := widget.NewDialog(ui.Tree(), "模态", ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), in.ElementID())
	ui.Tree().AppendChild(ui.Tree().Root(), card.ElementID())
	ui.Tree().AppendChild(ui.Tree().Root(), dlg.ElementID())
	place(ui.Tree(), in.ElementID(), 10, 240, 120, 28)
	place(ui.Tree(), card.ElementID(), 10, 10, 48, 24)
	place(ui.Tree(), dlg.ElementID(), 0, 0, 400, 300)
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: 20, GlobalY: 250})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 20, GlobalY: 250})
	ui.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: 20, GlobalY: 16})
	card.Tick(0.1)
	if !card.IsOpen() || ui.Tree().Focused() != in.ElementID() {
		t.Fatalf("hover setup open %v focus %d", card.IsOpen(), ui.Tree().Focused())
	}
	dlg.Open()
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyEscape})
	if dlg.IsVisible() {
		t.Fatal("first escape did not close the modal")
	}
	if !card.IsOpen() {
		t.Fatal("first escape closed the hover card hidden under the modal")
	}
	if ui.Tree().Focused() != in.ElementID() {
		t.Fatalf("modal close focus %d", ui.Tree().Focused())
	}
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyEscape})
	if card.IsOpen() || ui.Tree().Focused() != in.ElementID() {
		t.Fatalf("second escape open %v focus %d", card.IsOpen(), ui.Tree().Focused())
	}
}

func TestCameraHostPanZoomAndHit(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(240, 200)
	cam := widget.NewCamera(ui.Tree(), ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), cam.ElementID())
	place(ui.Tree(), cam.ElementID(), 0, 0, 200, 160)
	cam.AddMarker("a", 40, 30)
	sx, sy := cam.WorldToScreen(40, 30)
	if cam.MarkerAt(sx, sy) != "a" {
		t.Fatalf("marker screen %v %v id %q", sx, sy, cam.MarkerAt(sx, sy))
	}
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: sx, GlobalY: sy})
	ui.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: sx + 30, GlobalY: sy + 10})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: sx + 30, GlobalY: sy + 10})
	if cam.OriginX() < 20 || cam.MarkerAt(sx, sy) == "a" {
		t.Fatalf("pan origin %v still %q", cam.OriginX(), cam.MarkerAt(sx, sy))
	}
	ui.HandleEvent(&event.Event{Type: event.MouseWheel, GlobalX: 100, GlobalY: 80, WheelDY: -1})
	if cam.Zoom() <= 1 {
		t.Fatalf("zoom %v", cam.Zoom())
	}
	buf := render.NewCommandBuffer()
	cam.Draw(buf)
	be := godot.NewSoftwareBackend()
	be.Resize(200, 160)
	be.BeginFrame()
	be.Submit(buf)
	px := be.Pixels()
	if len(px) < 4 || px[3] == 0 && px[0] == 0 {
		// A marker or grid line must survive the raster somewhere in the buffer.
		lit := false
		for i := 0; i < len(px); i += 4 {
			if px[i] != 0 || px[i+1] != 0 || px[i+2] != 0 {
				lit = true
				break
			}
		}
		if !lit {
			t.Fatal("camera draw produced an empty software frame")
		}
	}
}

func TestPlotHostBarsAxesAndPie(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(360, 240)
	plot := widget.NewPlot(ui.Tree(), ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), plot.ElementID())
	place(ui.Tree(), plot.ElementID(), 0, 0, 320, 200)
	plot.SetBars([]float32{2, 8, 4})
	buf := render.NewCommandBuffer()
	plot.Draw(buf)
	series, index, ok := plot.Hit(plot.BarAnchor(1))
	if !ok || series != 0 || index != 1 {
		t.Fatalf("bar hit series %d index %d ok %v", series, index, ok)
	}
	px, py := plot.BarAnchor(1)
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: px, GlobalY: py})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: px, GlobalY: py})
	if plot.Selected() != 1 {
		t.Fatalf("selected %d", plot.Selected())
	}
	if plot.Canvas().Pixels() == nil {
		t.Fatal("plot has no canvas pixels")
	}
	lit := false
	for i := 0; i < len(plot.Canvas().Pixels()); i += 4 {
		if plot.Canvas().Pixels()[i+3] != 0 {
			lit = true
			break
		}
	}
	if !lit {
		t.Fatal("bar chart did not paint the canvas")
	}
	be := godot.NewSoftwareBackend()
	be.Resize(320, 200)
	be.BeginFrame()
	be.Submit(buf)
	frame := false
	for i := 0; i < len(be.Pixels()); i += 4 {
		if be.Pixels()[i] != 0 || be.Pixels()[i+1] != 0 || be.Pixels()[i+2] != 0 {
			frame = true
			break
		}
	}
	if !frame {
		t.Fatal("bar chart command buffer rasterized empty")
	}
	plot.SetPie([]float32{1, 3})
	plot.Draw(render.NewCommandBuffer())
	_, slice, ok := plot.Hit(plot.PieAnchor(1))
	if !ok || slice != 1 {
		t.Fatalf("pie hit %d ok %v", slice, ok)
	}
	plot.SetLine([]widget.PlotPoint{{X: 0, Y: 1}, {X: 1, Y: 4}, {X: 2, Y: 2}})
	plot.Draw(render.NewCommandBuffer())
	if _, _, ok := plot.Hit(plot.LineAnchor(1)); !ok {
		t.Fatal("line vertex not hittable")
	}
	plot.SetTreemap([]widget.TreemapItem{{Name: "a", Value: 1}, {Name: "b", Value: 3}})
	plot.Draw(render.NewCommandBuffer())
	if plot.TreemapRect(1).Width <= plot.TreemapRect(0).Width {
		t.Fatal("treemap did not give the larger value a wider rect")
	}
	plot.SetSankey(
		[]widget.SankeyNode{{Name: "in", Value: 10}, {Name: "out", Value: 10}},
		[]widget.SankeyLink{{From: 0, To: 1, Value: 10}},
	)
	plot.Draw(render.NewCommandBuffer())
	if plot.SankeyNode(0).Height <= 0 || plot.SankeyNode(1).X <= plot.SankeyNode(0).X {
		t.Fatal("sankey nodes were not laid out in columns")
	}
}

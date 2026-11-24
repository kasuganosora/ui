package ui

import (
	"testing"

	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/godot"
	uimath "github.com/kasuganosora/ui/math"
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

func TestCameraClipsViewportAndAnchorsGrid(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(400, 300)
	cam := widget.NewCamera(ui.Tree(), ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), cam.ElementID())
	place(ui.Tree(), cam.ElementID(), 80, 60, 120, 100)
	cam.AddMarker("out", -30, 10)
	cam.AddMarker("in", 20, 20)
	pr, pg, pb, _ := ui.Config().PrimaryColor.RGBA8()
	buf := render.NewCommandBuffer()
	cam.Draw(buf)
	be := godot.NewSoftwareBackend()
	be.Resize(400, 300)
	be.BeginFrame()
	be.Submit(buf)
	or, og, ob, oa := pixAt(be.Pixels(), 400, 50, 70)
	if oa != 0 && or == pr && og == pg && ob == pb {
		t.Fatalf("marker painted outside the viewport at (50,70) #%02x%02x%02x", or, og, ob)
	}
	inx, iny := cam.WorldToScreen(20, 20)
	ir, ig, ib, ia := pixAt(be.Pixels(), 400, int(inx), int(iny))
	if ia == 0 || ir != pr || ig != pg || ib != pb {
		t.Fatalf("inside marker missing at %v %v color %d %d %d %d", inx, iny, ir, ig, ib, ia)
	}
	phase0 := verticalGridX(buf, cam)
	cam.AddMarker("shift", 0, 0)
	// Slide the world so the origin phase moves by 15px.
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: 100, GlobalY: 80})
	ui.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: 115, GlobalY: 80})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 115, GlobalY: 80})
	buf = render.NewCommandBuffer()
	cam.Draw(buf)
	phase1 := verticalGridX(buf, cam)
	if len(phase0) == 0 || len(phase1) == 0 || phase0[0] == phase1[0] {
		t.Fatalf("grid did not follow the pan: %v -> %v", phase0, phase1)
	}
	sx, _ := cam.WorldToScreen(40, 0)
	found := false
	for _, x := range phase1 {
		if abs32(x-sx) < 1.5 {
			found = true
		}
	}
	if !found {
		t.Fatalf("grid phase %v want a line near world 40 at %v", phase1, sx)
	}
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: 140, GlobalY: 100})
	ui.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: 90, GlobalY: 100})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 90, GlobalY: 100})
	buf = render.NewCommandBuffer()
	cam.Draw(buf)
	be.BeginFrame()
	be.Submit(buf)
	ox, oy := cam.WorldToScreen(-30, 10)
	r, _, _, a := pixAt(be.Pixels(), 400, int(ox), int(oy))
	if ox < 80 || ox >= 200 || a == 0 || r != pr {
		t.Fatalf("marker did not pan back inside the viewport at %v %v color %d %d", ox, oy, r, a)
	}
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: 140, GlobalY: 100})
	ui.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: 190, GlobalY: 100})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 190, GlobalY: 100})
	buf = render.NewCommandBuffer()
	cam.Draw(buf)
	be.BeginFrame()
	be.Submit(buf)
	r, _, _, a = pixAt(be.Pixels(), 400, 50, 70)
	if a != 0 && r == pr {
		t.Fatal("marker stayed painted outside the viewport after panning out")
	}
}

func TestCameraClipDoesNotCoverOtherControls(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(400, 300)
	before := widget.NewMeter(ui.Tree(), ui.Config())
	before.SetValue(0.8)
	before.SetMarks([]float32{0.5})
	cam := widget.NewCamera(ui.Tree(), ui.Config())
	cam.AddMarker("in", 20, 20)
	after := widget.NewMeter(ui.Tree(), ui.Config())
	after.SetValue(0.8)
	after.SetMarks([]float32{0.5})
	cam2 := widget.NewCamera(ui.Tree(), ui.Config())
	cam2.AddMarker("b", 10, 10)
	plot := widget.NewPlot(ui.Tree(), ui.Config())
	plot.SetBars([]float32{4, 8})
	cmp := widget.NewImageCompare(ui.Tree(), ui.Config())
	cmp.SetBefore(widget.NewImage(ui.Tree(), render.TextureHandle(3), ui.Config()))
	cmp.SetAfter(widget.NewImage(ui.Tree(), render.TextureHandle(4), ui.Config()))
	cmp.SetPosition(40)
	ui.Tree().AppendChild(ui.Tree().Root(), before.ElementID())
	ui.Tree().AppendChild(ui.Tree().Root(), cam.ElementID())
	ui.Tree().AppendChild(ui.Tree().Root(), after.ElementID())
	ui.Tree().AppendChild(ui.Tree().Root(), cam2.ElementID())
	ui.Tree().AppendChild(ui.Tree().Root(), plot.ElementID())
	ui.Tree().AppendChild(ui.Tree().Root(), cmp.ElementID())
	place(ui.Tree(), before.ElementID(), 10, 40, 50, 20)
	place(ui.Tree(), cam.ElementID(), 80, 60, 120, 100)
	place(ui.Tree(), after.ElementID(), 240, 70, 100, 20)
	place(ui.Tree(), cam2.ElementID(), 260, 180, 80, 60)
	place(ui.Tree(), plot.ElementID(), 240, 210, 140, 70)
	place(ui.Tree(), cmp.ElementID(), 10, 180, 100, 50)
	buf := render.NewCommandBuffer()
	before.Draw(buf)
	cam.Draw(buf)
	after.Draw(buf)
	cam2.Draw(buf)
	plot.Draw(buf)
	cmp.Draw(buf)
	be := godot.NewSoftwareBackend()
	be.Resize(400, 300)
	be.BeginFrame()
	be.Submit(buf)
	pr, pg, pb, _ := ui.Config().PrimaryColor.RGBA8()
	assertPrimary := func(x, y int, label string) {
		t.Helper()
		r, g, b, a := pixAt(be.Pixels(), 400, x, y)
		if a == 0 || r != pr || g != pg || b != pb {
			t.Fatalf("%s pixel (%d,%d) = %d %d %d %d, want primary %d %d %d", label, x, y, r, g, b, a, pr, pg, pb)
		}
	}
	sx, sy := cam.WorldToScreen(20, 20)
	assertPrimary(int(sx), int(sy), "camera marker")
	assertPrimary(20, 50, "meter before camera")
	assertPrimary(250, 80, "meter after camera")
	r, g, b, _ := pixAt(be.Pixels(), 400, 291, 80)
	if r == 220 && g == 220 && b == 220 {
		t.Fatal("meter mark was clipped away, leaving the gray track")
	}
	assertPrimary(270, 190, "second camera marker")
	bx, by := plot.BarAnchor(1)
	assertPrimary(int(bx), int(by), "plot bar")
	split := 10 + 100*0.4
	r, g, b, a := pixAt(be.Pixels(), 400, int(split), 200)
	if a == 0 || r < 250 || g < 250 || b < 250 {
		t.Fatalf("image compare divider clipped: %d %d %d %d", r, g, b, a)
	}
	r, _, _, a = pixAt(be.Pixels(), 400, 50, 70)
	if a != 0 && r == pr {
		t.Fatal("outside-camera pixel changed to the marker color")
	}
}

func TestPlotScatterHitIsThePoint(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(400, 260)
	plot := widget.NewPlot(ui.Tree(), ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), plot.ElementID())
	place(ui.Tree(), plot.ElementID(), 30, 20, 320, 200)
	plot.SetScatter([]widget.PlotPoint{{X: 0, Y: 0}, {X: 2, Y: 1}, {X: 4, Y: 3}})
	x0, y0 := plot.LineAnchor(0)
	x1, y1 := plot.LineAnchor(1)
	if _, _, ok := plot.Hit((x0+x1)/2, (y0+y1)/2); ok {
		t.Fatal("scatter hit the empty span between points")
	}
	x2, y2 := plot.LineAnchor(2)
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: x2, GlobalY: y2})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: x2, GlobalY: y2})
	if plot.Selected() != 2 {
		t.Fatalf("third point selected %d", plot.Selected())
	}
	buf := render.NewCommandBuffer()
	plot.Draw(buf)
	if !softwareCovers(buf, 400, 260, int(x2), int(y2), ui.Config().PrimaryColor) {
		t.Fatal("scatter point was not rasterized")
	}
}

func TestPlotLinesAreaSankeyPie(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(400, 280)
	plot := widget.NewPlot(ui.Tree(), ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), plot.ElementID())
	place(ui.Tree(), plot.ElementID(), 40, 20, 320, 200)
	plot.SetLine([]widget.PlotPoint{{X: 0, Y: 1}, {X: 2, Y: 3}})
	x0, y0 := plot.LineAnchor(0)
	x1, y1 := plot.LineAnchor(1)
	mx, my := (x0+x1)/2, (y0+y1)/2
	if _, _, ok := plot.Hit(mx, my); !ok {
		t.Fatalf("segment midpoint %v %v missed", mx, my)
	}
	buf := render.NewCommandBuffer()
	plot.Draw(buf)
	if !softwareCovers(buf, 400, 280, int(mx), int(my), ui.Config().PrimaryColor) {
		t.Fatal("line segment was not rasterized away from the vertices")
	}
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: mx, GlobalY: my})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: mx, GlobalY: my})
	if plot.Selected() < 0 {
		t.Fatal("host click on the segment did not select")
	}
	plot.SetArea([]widget.PlotPoint{{X: 0, Y: 0}, {X: 1, Y: 4}, {X: 2, Y: 0}})
	_, peak := plot.LineAnchor(1)
	_, base := plot.LineAnchor(0)
	ix, iy := plot.LineAnchor(1)
	iy = peak + (base-peak)*0.5
	if _, _, ok := plot.Hit(ix, iy); !ok {
		t.Fatalf("area interior %v %v missed", ix, iy)
	}
	if _, _, ok := plot.Hit(4, iy); ok {
		t.Fatal("area hit outside the plot")
	}
	buf = render.NewCommandBuffer()
	plot.Draw(buf)
	if !softwareCovers(buf, 400, 280, int(ix), int(iy), ui.Config().PrimaryColor) {
		t.Fatal("area fill was not rasterized under the line")
	}
	plot.SetSankey(
		[]widget.SankeyNode{{Name: "a", Value: 10}, {Name: "b", Value: 4}, {Name: "c", Value: 10}},
		[]widget.SankeyLink{{From: 0, To: 2, Value: 10}},
	)
	plot.Draw(render.NewCommandBuffer())
	n0, n2 := plot.SankeyNode(0), plot.SankeyNode(2)
	ex := n0.X + n0.Width + 4
	ey := n0.Y + n0.Height/2
	if n2.X <= n0.X || n0.Contains(uimath.Vec2{X: ex, Y: ey}) || n2.Contains(uimath.Vec2{X: ex, Y: ey}) {
		t.Fatalf("edge sample landed on a node %v %v", n0, n2)
	}
	buf = render.NewCommandBuffer()
	plot.Draw(buf)
	if !commandCovers(buf, ex, ey) && !softwareCovers(buf, 400, 280, int(ex), int(ey), ui.Config().WarningColor) {
		t.Fatal("sankey link 0->2 has no edge away from the nodes")
	}
	plot.SetPie([]float32{1, 1, 2})
	if _, _, ok := plot.Hit(42, 52); ok {
		t.Fatal("point outside the pie circle hit a sector")
	}
	for i := 0; i < 3; i++ {
		ax, ay := plot.PieAnchor(i)
		_, got, ok := plot.Hit(ax, ay)
		if !ok || got != i {
			t.Fatalf("wedge %d hit %d ok %v at %v %v", i, got, ok, ax, ay)
		}
	}
	ax, ay := plot.PieAnchor(2)
	buf = render.NewCommandBuffer()
	plot.Draw(buf)
	if !softwareCovers(buf, 400, 280, int(ax), int(ay), ui.Config().SuccessColor) && !softwareCovers(buf, 400, 280, int(ax), int(ay), ui.Config().PrimaryColor) {
		t.Fatal("pie wedge was not rasterized from the command buffer")
	}
	plot.SetBars(nil)
	plot.Draw(render.NewCommandBuffer())
	if _, _, ok := plot.Hit(80, 80); ok {
		t.Fatal("empty bars hit")
	}
	plot.SetBars([]float32{0, 0})
	plot.Draw(render.NewCommandBuffer())
	plot.SetBars([]float32{-4, 8})
	plot.Draw(render.NewCommandBuffer())
	ax, ay = plot.BarAnchor(0)
	if _, idx, ok := plot.Hit(ax, ay); ok && idx == 0 {
		t.Fatal("negative bar was treated as a positive body")
	}
	plot.SetBars([]float32{1e8, 1})
	plot.Draw(render.NewCommandBuffer())
	if _, idx, ok := plot.Hit(plot.BarAnchor(0)); !ok || idx != 0 {
		t.Fatal("extreme bar was not hittable")
	}
}

func pixAt(px []byte, w, x, y int) (byte, byte, byte, byte) {
	if x < 0 || y < 0 || len(px) < (y*w+x+1)*4 {
		return 0, 0, 0, 0
	}
	i := (y*w + x) * 4
	return px[i], px[i+1], px[i+2], px[i+3]
}

func verticalGridX(buf *render.CommandBuffer, cam *widget.Camera) []float32 {
	var xs []float32
	b := cam.Bounds()
	for _, c := range buf.Commands() {
		if c.Rect == nil || c.Rect.Bounds.Width > 1.5 || c.Rect.Bounds.Height < b.Height*0.5 {
			continue
		}
		xs = append(xs, c.Rect.Bounds.X)
	}
	return xs
}

func abs32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

func softwareCovers(buf *render.CommandBuffer, w, h, x, y int, want uimath.Color) bool {
	be := godot.NewSoftwareBackend()
	be.Resize(w, h)
	be.BeginFrame()
	be.Submit(buf)
	r, g, b, a := pixAt(be.Pixels(), w, x, y)
	wr, wg, wb, _ := want.RGBA8()
	return a != 0 && r == wr && g == wg && b == wb
}

func commandCovers(buf *render.CommandBuffer, x, y float32) bool {
	for _, c := range buf.Commands() {
		if c.Rect != nil && c.Rect.Bounds.Contains(uimath.Vec2{X: x, Y: y}) && c.Rect.Bounds.Width > 16 {
			return true
		}
	}
	return false
}

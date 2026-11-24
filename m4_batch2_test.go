package ui

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/godot"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

func TestDataTableHostReorderResizeFrozenWindow(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(480, 320)
	tree := ui.Tree()
	cols := []widget.TableColumn{
		{ColKey: "a", Title: "A", Width: 120},
		{ColKey: "b", Title: "B", Width: 120},
		{ColKey: "c", Title: "C", Width: 120},
		{ColKey: "d", Title: "D", Width: 120},
	}
	dt := widget.NewDataTable(tree, cols, ui.Config())
	meter := widget.NewMeter(tree, ui.Config())
	tree.AppendChild(tree.Root(), dt.ElementID())
	tree.AppendChild(tree.Root(), meter.ElementID())
	place(tree, dt.ElementID(), 40, 30, 360, 140)
	place(tree, meter.ElementID(), 40, 200, 80, 16)
	rows := make([][]string, 3000)
	for i := range rows {
		rows[i] = []string{"a", "b", "c", "d"}
	}
	rows[0] = []string{"a0", "b0", "c0", "d0"}
	dt.SetRows(rows)
	dt.SetRowHeight(20)
	dt.SetHeaderHeight(24)

	if dt.MoveColumn(2, 0) {
		got := dt.ColumnKeys()
		want := []string{"c", "a", "b", "d"}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("keys %v", got)
			}
		}
		if dt.Rows()[0][0] != "c0" || dt.Rows()[0][1] != "a0" {
			t.Fatalf("cells did not follow the column: %v", dt.Rows()[0])
		}
	} else {
		t.Fatal("move of a non-edge column failed")
	}
	fresh := []string{"a", "b", "c", "d"}
	for i, key := range fresh {
		dt.MoveColumn(indexOf(dt.ColumnKeys(), key), i)
	}
	if keys := dt.ColumnKeys(); keys[0] != "a" || keys[3] != "d" {
		t.Fatalf("restore %v", keys)
	}
	if !dt.MoveColumn(1, 2) || dt.ColumnKeys()[2] != "b" {
		t.Fatalf("sibling move %v", dt.ColumnKeys())
	}
	for i, key := range fresh {
		dt.MoveColumn(indexOf(dt.ColumnKeys(), key), i)
	}

	dt.SetFrozen(1, 1)
	dt.SetScrollX(100)
	a := dt.ColumnHeader(0)
	c := dt.ColumnHeader(2)
	d := dt.ColumnHeader(3)
	if a.X != 40 || a.Y != 30 {
		t.Fatalf("frozen left header %v", a)
	}
	if d.X != 40+360-120 {
		t.Fatalf("frozen right header %v", d)
	}
	if c.X != 180 {
		t.Fatalf("scrolled sibling header %v", c)
	}

	buf := render.NewCommandBuffer()
	dt.Draw(buf)
	rects := 0
	for _, cmd := range buf.Commands() {
		if cmd.Type == render.CmdRect {
			rects++
		}
	}
	if rects == 0 || rects >= 200 {
		t.Fatalf("windowed draw emitted %d rects", rects)
	}
	be := godot.NewSoftwareBackend()
	be.Resize(480, 320)
	be.BeginFrame()
	be.Submit(buf)
	sr, sg, sb, _ := ui.Config().SuccessColor.RGBA8()
	pr, pg, pb, _ := ui.Config().PrimaryColor.RGBA8()
	r, g, b, aa := pixAt(be.Pixels(), 480, 50, 36)
	if aa == 0 || r != sr || g != sg || b != sb {
		t.Fatalf("frozen header pixel %d %d %d %d", r, g, b, aa)
	}
	mark := dt.ColumnMark(2)
	r, g, b, aa = pixAt(be.Pixels(), 480, int(mark.X+2), int(mark.Y+2))
	if r != pr || g != pg || b != pb {
		t.Fatalf("scrolled column mark %d %d %d", r, g, b)
	}

	grip := dt.ColumnHeader(0)
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: grip.X + grip.Width - 2, GlobalY: grip.Y + 4})
	ui.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: grip.X + grip.Width + 22, GlobalY: grip.Y + 4})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: grip.X + grip.Width + 22, GlobalY: grip.Y + 4})
	if dt.ColumnWidth(0) < 140 {
		t.Fatalf("width %v", dt.ColumnWidth(0))
	}

	var row, col int
	dt.OnCellClick(func(ctx widget.CellEventContext) {
		row, col = ctx.RowIndex, ctx.ColIndex
	})
	dt.SetScrollY(200)
	bodyY := float32(30 + 24 + 4)
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: 50, GlobalY: bodyY})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 50, GlobalY: bodyY})
	if row != 10 || col != 0 {
		t.Fatalf("scrolled cell %d %d", row, col)
	}
	ui.HandleEvent(&event.Event{Type: event.MouseWheel, GlobalX: 50, GlobalY: bodyY, WheelDY: 1})
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: 50, GlobalY: bodyY})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 50, GlobalY: bodyY})
	if row != 11 {
		t.Fatalf("wheel row %d", row)
	}

	dt.ClearRows()
	row = -1
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: 50, GlobalY: bodyY})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 50, GlobalY: bodyY})
	if row != -1 {
		t.Fatal("cleared table still hit a row")
	}
	id := dt.ElementID()
	dt.Destroy()
	if tree.Get(id) != nil || tree.Get(meter.ElementID()) == nil {
		t.Fatal("destroy removed a sibling or left the table")
	}
}

func TestTextViewMarkdownLinksAndCode(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(480, 360)
	tree := ui.Tree()
	tv := widget.NewTextView(tree, ui.Config())
	tree.AppendChild(tree.Root(), tv.ElementID())
	place(tree, tv.ElementID(), 80, 50, 280, 240)
	src := "# Title\n\n- alpha\n- beta\n\n```\ncode line\n```\n\n[one](https://example.com/one) [two](https://example.com/two)\n"
	tv.SetMarkdown(src)
	blocks := tv.Blocks()
	var heading, listA, listB, code, link1, link2 widget.TextBlock
	var nList int
	for _, b := range blocks {
		switch b.Kind {
		case "heading":
			heading = b
		case "list":
			nList++
			if listA.Text == "" {
				listA = b
			} else {
				listB = b
			}
		case "code":
			if !b.Inline {
				code = b
			}
		case "link":
			if link1.URL == "" {
				link1 = b
			} else {
				link2 = b
			}
		}
	}
	if heading.Text != "Title" || heading.Level != 1 || heading.Bounds.X < 80 {
		t.Fatalf("heading %+v", heading)
	}
	if nList != 2 || listA.Text != "alpha" || listB.Text != "beta" || !(listB.Bounds.Y > listA.Bounds.Y) {
		t.Fatalf("lists %+v %+v", listA, listB)
	}
	if code.Text != "code line" || code.Bounds.X < 80 {
		t.Fatalf("code %+v", code)
	}
	if link1.URL == link2.URL || link2.URL != "https://example.com/two" || link2.Bounds.X <= link1.Bounds.X {
		t.Fatalf("links %+v %+v", link1, link2)
	}
	if url, ok := tv.LinkAt(link1.Bounds.X+2, link1.Bounds.Y+2); !ok || url != link1.URL {
		t.Fatalf("first link hit %q %v", url, ok)
	}
	var got string
	var n int
	tv.OnLink(func(url string) {
		got = url
		n++
	})
	click(ui, link2.Bounds.X+4, link2.Bounds.Y+4)
	if n != 1 || got != "https://example.com/two" {
		t.Fatalf("click %q n %d", got, n)
	}
	click(ui, heading.Bounds.X+4, heading.Bounds.Y+4)
	if n != 1 {
		t.Fatal("heading activated a link")
	}
	buf := render.NewCommandBuffer()
	tv.Draw(buf)
	be := godot.NewSoftwareBackend()
	be.Resize(480, 360)
	be.BeginFrame()
	be.Submit(buf)
	sr, sg, sb, _ := ui.Config().SuccessColor.RGBA8()
	px := int(code.Bounds.X + code.Bounds.Width - 8)
	py := int(code.Bounds.Y + code.Bounds.Height/2)
	r, g, b, a := pixAt(be.Pixels(), 480, px, py)
	if a == 0 || r != sr || g != sg || b != sb {
		t.Fatalf("code pixel (%d,%d) %d %d %d %d", px, py, r, g, b, a)
	}
	tv.SetMarkdown("plain sentence")
	if _, ok := tv.LinkAt(link2.Bounds.X+4, link2.Bounds.Y+4); ok {
		t.Fatal("old link survived a new document")
	}
	click(ui, link2.Bounds.X+4, link2.Bounds.Y+4)
	if n != 1 {
		t.Fatal("cleared document still fired a link")
	}
	id := tv.ElementID()
	tv.Destroy()
	if tree.Get(id) != nil {
		t.Fatal("text view survived destroy")
	}
}

func TestFlowLayersUseCamera(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(520, 360)
	tree := ui.Tree()
	flow := widget.NewFlow(tree, ui.Config())
	tree.AppendChild(tree.Root(), flow.ElementID())
	place(tree, flow.ElementID(), 30, 24, 360, 240)
	flow.AddNode("A", "A")
	flow.AddNode("B", "B")
	flow.AddNode("C", "C")
	flow.AddNode("E", "E")
	flow.AddNode("D", "D")
	flow.AddEdge("A", "B")
	flow.AddEdge("A", "C")
	flow.AddEdge("B", "D")
	flow.AddEdge("C", "D")
	flow.AddEdge("A", "A")
	flow.Layout()
	if flow.Layer("A") != 0 || flow.Layer("E") != 0 || flow.Layer("B") != 1 || flow.Layer("C") != 1 || flow.Layer("D") != 2 {
		t.Fatalf("layers A %d B %d C %d D %d E %d", flow.Layer("A"), flow.Layer("B"), flow.Layer("C"), flow.Layer("D"), flow.Layer("E"))
	}
	sb, sc := flow.NodeScreen("B"), flow.NodeScreen("C")
	if sb.Y >= sc.Y || sb.X < 30 {
		t.Fatalf("siblings B %v C %v", sb, sc)
	}
	if flow.NodeScreen("E").Y <= flow.NodeScreen("A").Y {
		t.Fatal("disconnected sibling did not stay on the first layer below A")
	}
	d := flow.NodeScreen("D")
	if flow.NodeAt(d.X+4, d.Y+4) != "D" {
		t.Fatal("node hit missed D")
	}
	click(ui, d.X+10, d.Y+10)
	if flow.Selected() != "D" {
		t.Fatalf("selected %q", flow.Selected())
	}
	buf := render.NewCommandBuffer()
	flow.Draw(buf)
	be := godot.NewSoftwareBackend()
	be.Resize(520, 360)
	be.BeginFrame()
	be.Submit(buf)
	pr, pg, pb, _ := ui.Config().PrimaryColor.RGBA8()
	a := flow.NodeScreen("A")
	r, g, b, alpha := pixAt(be.Pixels(), 520, int(a.X+8), int(a.Y+8))
	if alpha == 0 || r != pr || g != pg || b != pb {
		t.Fatalf("node pixel %d %d %d %d", r, g, b, alpha)
	}
	r, _, _, alpha = pixAt(be.Pixels(), 520, 40, 200)
	if alpha != 0 && r == pr {
		t.Fatal("empty flow space was painted as a node")
	}
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: 40, GlobalY: 200})
	ui.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: 70, GlobalY: 200})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 70, GlobalY: 200})
	if flow.Camera().OriginX() == 0 {
		t.Fatal("drag did not pan the camera")
	}
	moved := flow.NodeScreen("D")
	if moved.X >= d.X {
		t.Fatalf("pan left the node at %v from %v", moved, d)
	}
	if flow.NodeAt(moved.X+4, moved.Y+4) != "D" {
		t.Fatal("node hit did not follow the camera")
	}
	flow.RemoveNode("B")
	flow.Layout()
	if flow.NodeAt(sb.X+4, sb.Y+4) == "B" || flow.Layer("D") != 2 {
		t.Fatalf("remove B layer D %d", flow.Layer("D"))
	}
	if flow.HasEdge("A", "B") || flow.HasEdge("B", "D") || !flow.HasEdge("C", "D") {
		t.Fatal("edges did not follow the removed node")
	}
	id := flow.ElementID()
	cam := flow.Camera().ElementID()
	flow.Destroy()
	if tree.Get(id) != nil || tree.Get(cam) != nil {
		t.Fatal("flow destroy left an element")
	}
}

func click(ui *EmbeddedUI, x, y float32) {
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: x, GlobalY: y})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: x, GlobalY: y})
}

func TestHTMLDataTableTextViewAndFlow(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<datatable></datatable><textview markdown="# Hi"></textview><flow></flow>`, "")
	if len(doc.QueryByTag("datatable")) != 1 || len(doc.QueryByTag("textview")) != 1 || len(doc.QueryByTag("flow")) != 1 {
		t.Fatalf("tags %d %d %d", len(doc.QueryByTag("datatable")), len(doc.QueryByTag("textview")), len(doc.QueryByTag("flow")))
	}
	tv := doc.QueryByTag("textview")[0].(*widget.TextView)
	place(tree, tv.ElementID(), 0, 0, 200, 80)
	blocks := tv.Blocks()
	if len(blocks) != 1 || blocks[0].Kind != "heading" || blocks[0].Text != "Hi" {
		t.Fatalf("markdown attr %+v", blocks)
	}
}

func indexOf(keys []string, key string) int {
	for i, k := range keys {
		if k == key {
			return i
		}
	}
	return -1
}

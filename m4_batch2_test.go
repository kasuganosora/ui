package ui

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/font"
	"github.com/kasuganosora/ui/font/atlas"
	"github.com/kasuganosora/ui/font/textrender"
	"github.com/kasuganosora/ui/godot"
	uimath "github.com/kasuganosora/ui/math"
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
	r, g, b, aa := pixAt(be.Pixels(), 480, 50, 36)
	if aa == 0 || r != sr || g != sg || b != sb {
		t.Fatalf("frozen header pixel %d %d %d %d", r, g, b, aa)
	}
	r, g, b, aa = pixAt(be.Pixels(), 480, int(c.X+8), int(c.Y+8))
	if r == sr && g == sg && b == sb {
		t.Fatal("scrolled header was painted with the frozen column color")
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

func TestDataTablePaintsLabelsAndClipsRows(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(480, 320)
	fx := newGlyphFixture(t, 480, 320)
	ui.Config().TextRenderer = fx.drawer
	tree := ui.Tree()
	cols := []widget.TableColumn{
		{ColKey: "a", Title: "Aa", Width: 80, Align: widget.TableAlignLeft},
		{ColKey: "b", Title: "Bb", Width: 80, Align: widget.TableAlignRight},
	}
	dt := widget.NewDataTable(tree, cols, ui.Config())
	tree.AppendChild(tree.Root(), dt.ElementID())
	place(tree, dt.ElementID(), 80, 50, 160, 100)
	dt.SetHeaderHeight(28)
	dt.SetRowHeight(28)
	rows := make([][]string, 40)
	for i := range rows {
		rows[i] = []string{"c0", "d0"}
	}
	rows[0] = []string{"c0", "d0"}
	rows[30] = []string{"FAR", "ROW"}
	dt.SetRows(rows)
	dt.SetScrollY(10)

	buf := render.NewCommandBuffer()
	dt.Draw(buf)
	texts := textCmds(buf)
	header := dt.ColumnHeader(0)
	if !hasText(texts, "Aa", header) {
		t.Fatal("header Aa did not reach a text command inside the header cell")
	}
	right := dt.ColumnHeader(1)
	if cmd := textIn(texts, "Bb", right); cmd == nil || cmd.Glyphs[0].X < right.X+right.Width/2 {
		t.Fatal("right-aligned header was not placed in the right half of its cell")
	}
	body := uimath.NewRect(80, 78, 80, 28)
	if !hasText(texts, "c0", body) {
		t.Fatal("visible cell c0 did not reach a text command")
	}
	if hasText(texts, "FAR", uimath.NewRect(0, 0, 480, 320)) {
		t.Fatal("a row outside the window was painted")
	}
	fx.present(buf)
	r, g, b, a := pixAt(fx.be.Pixels(), 480, 90, 152)
	if a != 0 && r > 200 && g > 200 && b > 200 {
		t.Fatalf("row leaked below the table (%d,%d) = %d %d %d %d", 90, 152, r, g, b, a)
	}
	r, g, b, a = pixAt(fx.be.Pixels(), 480, 80+160+4, 90)
	if a != 0 {
		t.Fatalf("table painted a neighbor pixel %d %d %d %d", r, g, b, a)
	}
	glyph := textIn(texts, "Aa", header).Glyphs[0]
	r, g, b, a = pixAt(fx.be.Pixels(), 480, int(glyph.X+glyph.Width/2), int(glyph.Y+glyph.Height/2))
	if a == 0 || r > 80 {
		t.Fatalf("header glyph pixel %d %d %d %d", r, g, b, a)
	}
	var row int = -1
	dt.OnCellClick(func(ctx widget.CellEventContext) { row = ctx.RowIndex })
	click(ui, 90, 149)
	if row != 2 {
		t.Fatalf("partial last row hit %d", row)
	}
	row = -1
	click(ui, 90, 152)
	if row != -1 {
		t.Fatal("hit below the table")
	}
}

func TestDataTableBodyClipLeavesHeader(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(480, 320)
	fx := newGlyphFixture(t, 480, 320)
	ui.Config().TextRenderer = fx.drawer
	tree := ui.Tree()
	cols := []widget.TableColumn{
		{ColKey: "a", Title: "Left", Width: 70},
		{ColKey: "b", Title: "Mid", Width: 50},
		{ColKey: "c", Title: "Right", Width: 40},
	}
	dt := widget.NewDataTable(tree, cols, ui.Config())
	tree.AppendChild(tree.Root(), dt.ElementID())
	place(tree, dt.ElementID(), 80, 50, 160, 100)
	dt.SetHeaderHeight(28)
	dt.SetRowHeight(28)
	dt.SetFrozen(1, 1)
	rows := make([][]string, 12)
	for i := range rows {
		rows[i] = []string{"L", "M", "R"}
	}
	dt.SetRows(rows)
	sr, sg, sb, _ := ui.Config().SuccessColor.RGBA8()
	assertHeader := func(scroll float32) {
		t.Helper()
		dt.SetScrollY(scroll)
		buf := render.NewCommandBuffer()
		dt.Draw(buf)
		fx.present(buf)
		r, g, b, a := pixAt(fx.be.Pixels(), 480, 90, 74)
		if r != sr || g != sg || b != sb || a != 255 {
			t.Fatalf("scroll %v header pixel (90,74) = %d %d %d %d, want frozen green", scroll, r, g, b, a)
		}
		left := textIn(textCmds(buf), "Left", dt.ColumnHeader(0))
		mid := textIn(textCmds(buf), "Mid", dt.ColumnHeader(1))
		if left == nil || mid == nil {
			t.Fatalf("scroll %v missing header glyphs left %v mid %v", scroll, left != nil, mid != nil)
		}
		lg := left.Glyphs[0]
		r, g, b, a = pixAt(fx.be.Pixels(), 480, int(lg.X+lg.Width/2), int(lg.Y+lg.Height/2))
		if a == 0 || r > 80 {
			t.Fatalf("scroll %v frozen header glyph %d %d %d %d", scroll, r, g, b, a)
		}
		mg := mid.Glyphs[0]
		r, g, b, a = pixAt(fx.be.Pixels(), 480, int(mg.X+mg.Width/2), int(mg.Y+mg.Height/2))
		if a == 0 || r > 80 {
			t.Fatalf("scroll %v middle header glyph %d %d %d %d", scroll, r, g, b, a)
		}
	}
	assertHeader(10)
	assertHeader(28)
	assertHeader(38)

	dt.SetScrollY(10)
	buf := render.NewCommandBuffer()
	dt.Draw(buf)
	fx.present(buf)
	r, g, b, a := pixAt(fx.be.Pixels(), 480, 90, 152)
	if a != 0 {
		t.Fatalf("pixel below table %d %d %d %d", r, g, b, a)
	}
	r, g, b, a = pixAt(fx.be.Pixels(), 480, 80+160+4, 90)
	if a != 0 {
		t.Fatalf("neighbor pixel %d %d %d %d", r, g, b, a)
	}
	var row = -1
	dt.OnCellClick(func(ctx widget.CellEventContext) { row = ctx.RowIndex })
	click(ui, 90, 74)
	if row != -1 {
		t.Fatalf("header click hit row %d", row)
	}
	click(ui, 90, 149)
	if row != 2 {
		t.Fatalf("partial body row %d", row)
	}
	row = -1
	click(ui, 90, 152)
	if row != -1 {
		t.Fatal("hit below the table")
	}
	grip := dt.ColumnHeader(0)
	before := dt.ColumnWidth(0)
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: grip.X + grip.Width - 2, GlobalY: grip.Y + 4})
	ui.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: grip.X + grip.Width + 20, GlobalY: grip.Y + 4})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: grip.X + grip.Width + 20, GlobalY: grip.Y + 4})
	if dt.ColumnWidth(0) < before+10 {
		t.Fatalf("header resize %v from %v", dt.ColumnWidth(0), before)
	}
}

func TestTextViewPaintsGlyphs(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(480, 360)
	fx := newGlyphFixture(t, 480, 360)
	ui.Config().TextRenderer = fx.drawer
	tv := widget.NewTextView(ui.Tree(), ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), tv.ElementID())
	place(ui.Tree(), tv.ElementID(), 80, 50, 160, 200)
	tv.SetMarkdown("# Title\n\n- alpha\n\n`inline`\n\n```\ncode line\n```\n\n[one](https://example.com/one) [two](https://example.com/two)\n\nthis paragraph is long enough to wrap onto a second line inside the narrow view\n")
	buf := render.NewCommandBuffer()
	tv.Draw(buf)
	texts := textCmds(buf)
	if !hasText(texts, "Title", uimath.NewRect(80, 50, 160, 40)) {
		t.Fatal("heading was not drawn as glyphs")
	}
	if !hasText(texts, "alpha", uimath.NewRect(80, 50, 160, 200)) {
		t.Fatal("list item was not drawn as glyphs")
	}
	if !hasText(texts, "inline", uimath.NewRect(80, 50, 160, 200)) {
		t.Fatal("inline code was not drawn as glyphs")
	}
	if !hasText(texts, "code", uimath.NewRect(80, 50, 160, 200)) && !hasText(texts, "code line", uimath.NewRect(80, 50, 160, 200)) {
		t.Fatal("fenced code was not drawn as glyphs")
	}
	one := blockURL(tv, "https://example.com/one")
	two := blockURL(tv, "https://example.com/two")
	if !hasText(texts, "one", one) || !hasText(texts, "two", two) {
		t.Fatal("link labels were not drawn as glyphs")
	}
	var y1, y2 float32
	for _, b := range tv.Blocks() {
		if b.Kind == "text" && b.Text == "this" {
			y1 = b.Bounds.Y
		}
		if b.Kind == "text" && (b.Text == "narrow" || b.Text == "view" || b.Text == "line") {
			y2 = b.Bounds.Y
		}
	}
	if y1 == 0 || y2 == 0 || y2 <= y1 {
		t.Fatalf("long line did not wrap (%v, %v)", y1, y2)
	}
	link, ok := tv.LinkAt(linkCenter(tv, "https://example.com/two"))
	if !ok || link != "https://example.com/two" {
		t.Fatalf("wrapped layout lost the second link %q", link)
	}
	fx.present(buf)
	cmd := textIn(texts, "Title", uimath.NewRect(80, 50, 160, 80))
	g := cmd.Glyphs[0]
	r, gg, b, a := pixAt(fx.be.Pixels(), 480, int(g.X+g.Width/2), int(g.Y+g.Height/2))
	if a == 0 || r > 80 {
		t.Fatalf("heading glyph pixel %d %d %d %d", r, gg, b, a)
	}
	tv.SetMarkdown("x")
	buf = render.NewCommandBuffer()
	tv.Draw(buf)
	if hasText(textCmds(buf), "Title", uimath.NewRect(0, 0, 480, 360)) {
		t.Fatal("replaced document still drew the old heading")
	}
}

func TestFlowScalesNodesAndPaintsNames(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(520, 360)
	fx := newGlyphFixture(t, 520, 360)
	ui.Config().TextRenderer = fx.drawer
	flow := widget.NewFlow(ui.Tree(), ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), flow.ElementID())
	place(ui.Tree(), flow.ElementID(), 40, 30, 300, 200)
	flow.AddNode("A", "Alpha")
	flow.AddNode("B", "Beta")
	flow.AddEdge("A", "B")
	flow.Layout()
	flow.Camera().SetOrigin(0, 0)
	flow.Camera().SetZoom(1)
	base := flow.NodeScreen("A")
	worldX := base.X - 40
	worldY := base.Y - 30
	worldW, worldH := base.Width, base.Height
	flow.Camera().SetOrigin(12, -4)
	flow.Camera().SetZoom(2)
	got := flow.NodeScreen("A")
	wantX := float32(40) + (worldX-12)*2
	wantY := float32(30) + (worldY-(-4))*2
	if absf(got.X-wantX) > 0.5 || absf(got.Y-wantY) > 0.5 || absf(got.Width-worldW*2) > 0.5 || absf(got.Height-worldH*2) > 0.5 {
		t.Fatalf("zoom 2 screen %v want %v %v %v %v", got, wantX, wantY, worldW*2, worldH*2)
	}
	buf := render.NewCommandBuffer()
	flow.Draw(buf)
	if !rectNear(buf, got, ui.Config().PrimaryColor) {
		t.Fatal("drawn node rect does not match the scaled world rect")
	}
	if !hasText(textCmds(buf), "Alpha", got) {
		t.Fatal("node name was not drawn inside the scaled node")
	}
	inX, inY := got.X+got.Width*0.5, got.Y+got.Height*0.5
	if flow.NodeAt(inX, inY) != "A" {
		t.Fatal("scaled interior missed A")
	}
	click(ui, inX, inY)
	if flow.Selected() != "A" {
		t.Fatalf("host select %q", flow.Selected())
	}
	if flow.NodeAt(got.X+got.Width+3, got.Y+4) == "A" {
		t.Fatal("hit A outside the scaled rect")
	}
	b := flow.NodeScreen("B")
	if !edgeTouches(buf, got.X+got.Width, b.X) {
		t.Fatal("edge is not attached to the scaled node edges")
	}
	flow.Camera().SetOrigin(-80, 0)
	flow.Camera().SetZoom(2)
	flow.Draw(buf)
	buf = render.NewCommandBuffer()
	flow.Draw(buf)
	fx.present(buf)
	r, g, bcol, a := pixAt(fx.be.Pixels(), 520, int(40+300+2), int(got.Y+4))
	if a != 0 && r < 40 && g < 100 && bcol > 180 {
		t.Fatalf("scaled node leaked outside the view %d %d %d %d", r, g, bcol, a)
	}
	flow.Camera().SetOrigin(0, 0)
	flow.Camera().SetZoom(0.5)
	small := flow.NodeScreen("A")
	if absf(small.Width-worldW*0.5) > 0.5 || absf(small.Height-worldH*0.5) > 0.5 {
		t.Fatalf("zoom 0.5 size %v", small)
	}
	if flow.NodeAt(small.X+small.Width*0.5, small.Y+small.Height*0.5) != "A" {
		t.Fatal("zoom 0.5 interior missed A")
	}
	if flow.NodeAt(small.X+worldW-2, small.Y+4) == "A" {
		t.Fatal("zoom 0.5 hit A outside the visible scaled rect")
	}
}

func newGlyphFixture(t *testing.T, w, h int) glyphFixture {
	t.Helper()
	be := godot.NewSoftwareBackend()
	be.Resize(w, h)
	eng := NewMockEngine()
	mgr := font.NewManager(eng)
	id, err := mgr.Register("Default", font.WeightRegular, font.StyleNormal, nil)
	if err != nil {
		t.Fatal(err)
	}
	atl := atlas.New(atlas.Options{Width: 512, Height: 512, Backend: be})
	if err := atl.EnsureTexture(); err != nil {
		t.Fatal(err)
	}
	r := textrender.New(textrender.Options{Manager: mgr, Atlas: atl, DPIScale: 1})
	return glyphFixture{be: be, atlas: atl, drawer: NewTextDrawer(r, id, eng)}
}

type glyphFixture struct {
	be     *godot.SoftwareBackend
	atlas  *atlas.Atlas
	drawer widget.TextDrawer
}

func (g glyphFixture) present(buf *render.CommandBuffer) {
	_ = g.atlas.Upload()
	g.be.BeginFrame()
	g.be.Submit(buf)
}

func textCmds(buf *render.CommandBuffer) []render.Command {
	var out []render.Command
	for _, c := range buf.Commands() {
		if c.Type == render.CmdText && c.Text != nil && len(c.Text.Glyphs) > 0 {
			out = append(out, c)
		}
	}
	return out
}

func textIn(cmds []render.Command, sample string, area uimath.Rect) *render.TextCmd {
	want := len([]rune(sample))
	for _, c := range cmds {
		if len(c.Text.Glyphs) != want {
			continue
		}
		ok := true
		for _, g := range c.Text.Glyphs {
			if !glyphOverlaps(g, area) {
				ok = false
				break
			}
		}
		if ok {
			return c.Text
		}
	}
	return nil
}

func hasText(cmds []render.Command, sample string, area uimath.Rect) bool {
	return textIn(cmds, sample, area) != nil
}

func glyphOverlaps(g render.GlyphInstance, r uimath.Rect) bool {
	return g.X < r.X+r.Width && g.X+g.Width > r.X && g.Y < r.Y+r.Height && g.Y+g.Height > r.Y
}

func blockURL(tv *widget.TextView, url string) uimath.Rect {
	for _, b := range tv.Blocks() {
		if b.Kind == "link" && b.URL == url {
			return b.Bounds
		}
	}
	return uimath.Rect{}
}

func linkCenter(tv *widget.TextView, url string) (float32, float32) {
	for _, b := range tv.Blocks() {
		if b.Kind == "link" && b.URL == url {
			return b.Bounds.X + 2, b.Bounds.Y + 2
		}
	}
	return 0, 0
}

func rectNear(buf *render.CommandBuffer, want uimath.Rect, col uimath.Color) bool {
	for _, c := range buf.Commands() {
		if c.Type != render.CmdRect || c.Rect == nil {
			continue
		}
		b := c.Rect.Bounds
		if absf(b.X-want.X) < 1 && absf(b.Y-want.Y) < 1 && absf(b.Width-want.Width) < 1 && absf(b.Height-want.Height) < 1 && c.Rect.FillColor == col {
			return true
		}
	}
	return false
}

func edgeTouches(buf *render.CommandBuffer, left, right float32) bool {
	for _, c := range buf.Commands() {
		if c.Type != render.CmdRect || c.Rect == nil || c.Rect.Bounds.Height > 4 {
			continue
		}
		b := c.Rect.Bounds
		if absf(b.X-left) < 2 || absf(b.X+b.Width-right) < 2 || absf(b.X-right) < 2 {
			return true
		}
	}
	return false
}

func absf(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
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

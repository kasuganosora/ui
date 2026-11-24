package ui

import (
	"errors"
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

func TestFilePickerNavigatesInjectedList(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(480, 320)
	fx := newGlyphFixture(t, 480, 320)
	ui.Config().TextRenderer = fx.drawer
	fs := memFiles{
		"/":     {{Name: "docs", Dir: true}, {Name: "readme", Dir: false}},
		"/docs": {{Name: "guide", Dir: false}},
	}
	fp := widget.NewFilePicker(ui.Tree(), ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), fp.ElementID())
	place(ui.Tree(), fp.ElementID(), 40, 30, 220, 120)
	fp.SetSource(fs)
	fp.Open("/")
	if fp.Path() != "/" || len(fp.Entries()) != 2 || fp.Entries()[0].Name != "docs" {
		t.Fatalf("root %+v path %q", fp.Entries(), fp.Path())
	}
	buf := render.NewCommandBuffer()
	fp.Draw(buf)
	if !hasText(textCmds(buf), "docs", uimath.NewRect(40, 30, 220, 120)) || !hasText(textCmds(buf), "/", uimath.NewRect(40, 30, 220, 28)) {
		t.Fatal("directory name or path was not drawn as glyphs")
	}
	fx.present(buf)
	g := textIn(textCmds(buf), "docs", uimath.NewRect(40, 30, 220, 120)).Glyphs[0]
	r, gg, b, a := pixAt(fx.be.Pixels(), 480, int(g.X+g.Width/2), int(g.Y+g.Height/2))
	if a == 0 || r > 80 {
		t.Fatalf("name glyph %d %d %d %d", r, gg, b, a)
	}
	ui.Tree().SetFocused(fp.ElementID(), true)
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyEnter})
	if fp.Path() != "/docs" || len(fp.Entries()) != 1 || fp.Entries()[0].Name != "guide" {
		t.Fatalf("enter dir %+v %q", fp.Entries(), fp.Path())
	}
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyBackspace})
	if fp.Path() != "/" {
		t.Fatalf("up path %q", fp.Path())
	}
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyArrowDown})
	if fp.Selected() != "readme" {
		t.Fatalf("selected %q", fp.Selected())
	}
	var got string
	var dir bool
	fp.OnOpen(func(path string, isDir bool) {
		got = path
		dir = isDir
	})
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyEnter})
	if got != "/readme" || dir {
		t.Fatalf("file open %q dir %v", got, dir)
	}
	if err := fp.Open("/nope"); err == nil || fp.Error() == "" {
		t.Fatal("missing directory did not surface an error")
	}
	buf = render.NewCommandBuffer()
	fp.Draw(buf)
	if !hasText(textCmds(buf), fp.Error(), uimath.NewRect(40, 30, 220, 120)) {
		t.Fatal("error text was not drawn")
	}

	many := memFiles{"/": {}}
	for i := 0; i < 30; i++ {
		many["/"] = append(many["/"], widget.FileEntry{Name: "n" + string(rune('a'+i)), Dir: false})
	}
	many["/"] = append(many["/"], widget.FileEntry{Name: "hidden-row", Dir: false})
	fp.SetSource(many)
	fp.Open("/")
	fp.SetScrollY(0)
	buf = render.NewCommandBuffer()
	fp.Draw(buf)
	if hasText(textCmds(buf), "hidden-row", uimath.NewRect(0, 0, 480, 320)) {
		t.Fatal("offscreen file name was painted")
	}
	fp.SetScrollY(20)
	buf = render.NewCommandBuffer()
	fp.Draw(buf)
	fx.present(buf)
	sr, sg, sb, _ := ui.Config().SuccessColor.RGBA8()
	r, gg, b, a = pixAt(fx.be.Pixels(), 480, 200, 40)
	if r != sr || gg != sg || b != sb {
		t.Fatalf("scrolled row covered the path bar %d %d %d %d", r, gg, b, a)
	}
	id := fp.ElementID()
	fp.Destroy()
	if ui.Tree().Get(id) != nil {
		t.Fatal("file picker survived destroy")
	}
}

func TestCommitGraphLanesAndRows(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(480, 320)
	fx := newGlyphFixture(t, 480, 320)
	ui.Config().TextRenderer = fx.drawer
	g := widget.NewCommitGraph(ui.Tree(), ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), g.ElementID())
	place(ui.Tree(), g.ElementID(), 24, 36, 280, 80)
	g.SetCommits([]widget.Commit{
		{ID: "d4", Parents: []string{"b2", "c3"}, Subject: "merge topic"},
		{ID: "c3", Parents: []string{"a1"}, Subject: "side branch"},
		{ID: "b2", Parents: []string{"a1"}, Subject: "main line"},
		{ID: "a1", Subject: "ROOTROW"},
	})
	if g.Lane("b2") == g.Lane("c3") || g.Lane("d4") != g.Lane("b2") {
		t.Fatalf("lanes d %d b %d c %d a %d", g.Lane("d4"), g.Lane("b2"), g.Lane("c3"), g.Lane("a1"))
	}
	buf := render.NewCommandBuffer()
	g.Draw(buf)
	if !hasText(textCmds(buf), "merge topic", uimath.NewRect(24, 36, 280, 80)) {
		t.Fatal("commit subject was not drawn")
	}
	if hasText(textCmds(buf), "ROOTROW", uimath.NewRect(0, 0, 480, 320)) {
		t.Fatal("commit below the viewport was painted")
	}
	fx.present(buf)
	glyph := textIn(textCmds(buf), "merge topic", uimath.NewRect(24, 36, 280, 80)).Glyphs[0]
	r, gg, b, a := pixAt(fx.be.Pixels(), 480, int(glyph.X+glyph.Width/2), int(glyph.Y+glyph.Height/2))
	if a == 0 || r > 80 {
		t.Fatalf("subject glyph %d %d %d %d", r, gg, b, a)
	}
	row := g.RowScreen(0)
	click(ui, row.X+40, row.Y+8)
	if g.Selected() != "d4" {
		t.Fatalf("click selected %q", g.Selected())
	}
	ui.Tree().SetFocused(g.ElementID(), true)
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyArrowDown})
	if g.Selected() != "c3" {
		t.Fatalf("arrow selected %q", g.Selected())
	}
	g.SetScrollY(80)
	buf = render.NewCommandBuffer()
	g.Draw(buf)
	if !hasText(textCmds(buf), "ROOTROW", uimath.NewRect(24, 36, 280, 80)) {
		t.Fatal("scrolled root commit was not drawn")
	}
}

func TestDiagramBoxArrowLayout(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(480, 320)
	fx := newGlyphFixture(t, 480, 320)
	ui.Config().TextRenderer = fx.drawer
	d := widget.NewDiagram(ui.Tree(), ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), d.ElementID())
	place(ui.Tree(), d.ElementID(), 36, 28, 240, 90)
	d.SetSource("box Start\nbox End\nStart -> End\nsubgraph Ignored\n")
	if d.NodeCount() != 2 || d.EdgeCount() != 1 {
		t.Fatalf("nodes %d edges %d", d.NodeCount(), d.EdgeCount())
	}
	start, end := d.NodeScreen("Start"), d.NodeScreen("End")
	if end.X <= start.X || start.X < 36 {
		t.Fatalf("layout start %v end %v", start, end)
	}
	buf := render.NewCommandBuffer()
	d.Draw(buf)
	if !hasText(textCmds(buf), "Start", start) || !hasText(textCmds(buf), "End", end) {
		t.Fatal("diagram labels were not drawn inside their nodes")
	}
	fx.present(buf)
	glyph := textIn(textCmds(buf), "End", end).Glyphs[0]
	r, gg, b, a := pixAt(fx.be.Pixels(), 480, int(glyph.X+glyph.Width/2), int(glyph.Y+glyph.Height/2))
	if a == 0 || r > 80 {
		t.Fatalf("end glyph %d %d %d %d", r, gg, b, a)
	}
	click(ui, end.X+8, end.Y+8)
	if d.Selected() != "End" {
		t.Fatalf("selected %q", d.Selected())
	}
	ui.Tree().SetFocused(d.ElementID(), true)
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyArrowLeft})
	if d.Selected() != "Start" {
		t.Fatalf("arrow selected %q", d.Selected())
	}

	var chain string
	for i := 0; i < 8; i++ {
		id := "n" + string(rune('a'+i))
		chain += "box " + id + "\n"
		if i > 0 {
			chain += "n" + string(rune('a'+i-1)) + " -> " + id + "\n"
		}
	}
	chain += "box LASTNODE\nnh -> LASTNODE\n"
	d.SetSource(chain)
	place(ui.Tree(), d.ElementID(), 36, 28, 80, 70)
	buf = render.NewCommandBuffer()
	d.Draw(buf)
	if hasText(textCmds(buf), "LASTNODE", uimath.NewRect(0, 0, 480, 320)) {
		t.Fatal("diagram node outside the viewport was painted")
	}
	off := d.NodeScreen("LASTNODE")
	d.SetScrollX(off.X - 36)
	buf = render.NewCommandBuffer()
	d.Draw(buf)
	if !hasText(textCmds(buf), "LASTNODE", uimath.NewRect(36, 28, 80, 70)) {
		t.Fatal("scrolled diagram did not draw the last node inside the view")
	}
}

type memFiles map[string][]widget.FileEntry

func (m memFiles) List(path string) ([]widget.FileEntry, error) {
	ents, ok := m[path]
	if !ok {
		return nil, errors.New("missing")
	}
	out := make([]widget.FileEntry, len(ents))
	copy(out, ents)
	return out, nil
}

func TestHTMLFilePickerCommitGraphDiagram(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<filepicker></filepicker><commitgraph></commitgraph><diagram></diagram>`, "")
	if len(doc.QueryByTag("filepicker")) != 1 || len(doc.QueryByTag("commitgraph")) != 1 || len(doc.QueryByTag("diagram")) != 1 {
		t.Fatalf("tags %d %d %d", len(doc.QueryByTag("filepicker")), len(doc.QueryByTag("commitgraph")), len(doc.QueryByTag("diagram")))
	}
}

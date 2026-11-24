package ui

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

func TestHTMLEditorAndDiff(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<editor></editor><diffview></diffview>`, "")
	if len(doc.QueryByTag("editor")) != 1 || len(doc.QueryByTag("diffview")) != 1 {
		t.Fatalf("editor %d diff %d", len(doc.QueryByTag("editor")), len(doc.QueryByTag("diffview")))
	}
}

func TestEditorRuneCaretAndGlyphs(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(480, 320)
	fx := newGlyphFixture(t, 480, 320)
	ui.Config().TextRenderer = fx.drawer
	ed := widget.NewEditor(ui.Tree(), ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), ed.ElementID())
	place(ui.Tree(), ed.ElementID(), 48, 32, 160, 70)
	ed.SetText("a你b")
	ui.Tree().SetFocused(ed.ElementID(), true)
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyArrowRight})
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyArrowRight})
	if ed.Caret().Line != 0 || ed.Caret().Col != 2 {
		t.Fatalf("caret after two runes %+v", ed.Caret())
	}
	pt := ed.CaretPoint()
	width := fx.drawer.MeasureText("a你", ui.Config().FontSizeSm)
	if pt.X < 48+width-1 || pt.Y < 32 {
		t.Fatalf("caret point %v width %v", pt, width)
	}
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyBackspace})
	if ed.Text() != "ab" {
		t.Fatalf("backspace deleted a byte instead of the rune %q", ed.Text())
	}
	ui.HandleEvent(&event.Event{Type: event.KeyPress, Char: '你'})
	ui.HandleEvent(&event.Event{Type: event.KeyPress, Char: 'c'})
	if ed.Text() != "a你cb" {
		t.Fatalf("insert %q", ed.Text())
	}
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyArrowLeft, Modifiers: event.Modifiers{Shift: true}})
	if ed.SelectedText() != "c" {
		t.Fatalf("selection %q", ed.SelectedText())
	}
	ed.SetText("func hello")
	buf := render.NewCommandBuffer()
	ed.Draw(buf)
	if !hasText(textCmds(buf), "func", uimath.NewRect(48, 32, 160, 70)) || !hasText(textCmds(buf), "hello", uimath.NewRect(48, 32, 160, 70)) {
		t.Fatal("editor text was not drawn as glyphs")
	}
	fx.present(buf)
	g := textIn(textCmds(buf), "func", uimath.NewRect(48, 32, 160, 70)).Glyphs[0]
	r, gg, b, a := pixAt(fx.be.Pixels(), 480, int(g.X+g.Width/2), int(g.Y+g.Height/2))
	if a == 0 || (r > 80 && gg > 80 && b > 80) {
		t.Fatalf("keyword glyph %d %d %d %d", r, gg, b, a)
	}
	r, _, _, a = pixAt(fx.be.Pixels(), 480, 48+160+4, 40)
	if a != 0 {
		t.Fatal("editor painted outside its bounds")
	}

	var lines string
	for i := 0; i < 24; i++ {
		lines += "same\n"
	}
	lines += "BOTTOMROW"
	ed.SetText(lines)
	buf = render.NewCommandBuffer()
	ed.Draw(buf)
	if hasText(textCmds(buf), "BOTTOMROW", uimath.NewRect(0, 0, 480, 320)) {
		t.Fatal("editor drew a line below the viewport")
	}
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyEnd})
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyArrowDown})
	for ed.Caret().Line < 24 {
		ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyArrowDown})
		if ed.Caret().Line > 40 {
			break
		}
	}
	buf = render.NewCommandBuffer()
	ed.Draw(buf)
	if !hasText(textCmds(buf), "BOTTOMROW", uimath.NewRect(48, 32, 160, 70)) {
		t.Fatal("scrolled editor did not draw the last line")
	}
	id := ed.ElementID()
	ed.Destroy()
	if ui.Tree().Get(id) != nil {
		t.Fatal("editor survived destroy")
	}
}

func TestDiffViewSideBySide(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(480, 320)
	fx := newGlyphFixture(t, 480, 320)
	ui.Config().TextRenderer = fx.drawer
	d := widget.NewDiffView(ui.Tree(), ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), d.ElementID())
	place(ui.Tree(), d.ElementID(), 20, 40, 240, 90)
	var same string
	for i := 0; i < 12; i++ {
		same += "same\n"
	}
	d.SetTexts(same+"old\n", same+"new\n")
	buf := render.NewCommandBuffer()
	d.Draw(buf)
	if hasText(textCmds(buf), "new", uimath.NewRect(0, 0, 480, 320)) {
		t.Fatal("diff painted a change below the viewport")
	}
	ui.Tree().SetFocused(d.ElementID(), true)
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyN})
	if d.SelectedOp() != "change" && d.SelectedOp() != "add" && d.SelectedOp() != "del" {
		t.Fatalf("next change op %q", d.SelectedOp())
	}
	buf = render.NewCommandBuffer()
	d.Draw(buf)
	if !hasText(textCmds(buf), "old", uimath.NewRect(20, 40, 120, 90)) {
		t.Fatal("left change was not drawn in the left column")
	}
	if !hasText(textCmds(buf), "new", uimath.NewRect(140, 40, 120, 90)) {
		t.Fatal("right change was not drawn in the right column")
	}
	leftClip, ok := textClip(buf, "old", uimath.NewRect(20, 40, 120, 90))
	if !ok || leftClip.X+leftClip.Width > 140.5 {
		t.Fatalf("left text clip %v", leftClip)
	}
	rightClip, ok := textClip(buf, "new", uimath.NewRect(140, 40, 120, 90))
	if !ok || rightClip.X < 139.5 {
		t.Fatalf("right text clip %v", rightClip)
	}
	fx.present(buf)
	g := textIn(textCmds(buf), "new", uimath.NewRect(140, 40, 120, 90)).Glyphs[0]
	r, gg, b, a := pixAt(fx.be.Pixels(), 480, int(g.X+g.Width/2), int(g.Y+g.Height/2))
	if a == 0 || (r > 80 && gg > 80 && b > 80) {
		t.Fatalf("diff glyph %d %d %d %d", r, gg, b, a)
	}
	r, _, _, a = pixAt(fx.be.Pixels(), 480, 20+240+4, 50)
	if a != 0 {
		t.Fatal("diff painted outside its bounds")
	}
	row := d.RowScreen(d.SelectedIndex())
	click(ui, row.X+row.Width*0.75, row.Y+8)
	if d.Caret().Col < 1 {
		t.Fatalf("click col %d", d.Caret().Col)
	}
	d.SetTexts("alpha\nbeta\ngamma\n", "alpha\nBETA\ngamma\ndelta\n")
	if d.SelectedOp() == "change" || d.SelectedOp() == "add" {
		t.Fatal("reset text should start on unchanged context")
	}
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyN})
	if d.LeftText() != "beta" || d.RightText() != "BETA" {
		t.Fatalf("hunk %q | %q", d.LeftText(), d.RightText())
	}
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyN})
	if d.RightText() != "delta" || d.LeftText() != "" {
		t.Fatalf("second hunk %q | %q op %q", d.LeftText(), d.RightText(), d.SelectedOp())
	}
	id := d.ElementID()
	d.Destroy()
	if ui.Tree().Get(id) != nil {
		t.Fatal("diff survived destroy")
	}
}

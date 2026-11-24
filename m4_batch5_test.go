package ui

import (
	"strings"
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

func TestRichEditorMarksAndGlyphs(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(480, 320)
	fx := newGlyphFixture(t, 480, 320)
	ui.Config().TextRenderer = fx.drawer
	ed := widget.NewRichEditor(ui.Tree(), ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), ed.ElementID())
	place(ui.Tree(), ed.ElementID(), 40, 28, 180, 72)
	ui.Tree().SetFocused(ed.ElementID(), true)
	var changes int
	ed.OnChange(func(string) { changes++ })

	ed.SetText("abcd")
	ed.SetMark(2, 4, "bold")
	changes = 0
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyBackspace})
	if ed.Text() != "abcd" || changes != 0 || !sameMarks(ed.Marks(), []widget.RichMark{{Start: 2, End: 4, Kind: "bold"}}) {
		t.Fatalf("start backspace text %q changes %d marks %+v", ed.Text(), changes, ed.Marks())
	}
	ui.HandleEvent(&event.Event{Type: event.KeyPress, Char: 'X'})
	if ed.Text() != "Xabcd" || changes != 1 || !sameMarks(ed.Marks(), []widget.RichMark{{Start: 3, End: 5, Kind: "bold"}}) {
		t.Fatalf("insert %q changes %d marks %+v", ed.Text(), changes, ed.Marks())
	}

	ed.SetText("ab\ncd\nef")
	ed.SetMark(3, 5, "bold")
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyArrowDown})
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyArrowRight, Modifiers: event.Modifiers{Shift: true}})
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyArrowRight, Modifiers: event.Modifiers{Shift: true}})
	if ed.SelectedText() != "cd" {
		t.Fatalf("selected %q", ed.SelectedText())
	}
	changes = 0
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyBackspace})
	if ed.Text() != "ab\n\nef" || changes != 1 || len(ed.Marks()) != 0 {
		t.Fatalf("delete marked line %q marks %+v changes %d", ed.Text(), ed.Marks(), changes)
	}

	ed.SetText("code")
	ed.SetMark(0, 4, "code")
	buf := render.NewCommandBuffer()
	ed.Draw(buf)
	if !hasText(textCmds(buf), "code", uimath.NewRect(40, 28, 180, 72)) {
		t.Fatal("rich text was not drawn as glyphs")
	}
	fx.present(buf)
	g := textIn(textCmds(buf), "code", uimath.NewRect(40, 28, 180, 72)).Glyphs[0]
	r, gg, b, a := pixAt(fx.be.Pixels(), 480, int(g.X+g.Width/2), int(g.Y+g.Height/2))
	if a == 0 || (r > 80 && gg > 80 && b > 80) {
		t.Fatalf("rich glyph %d %d %d %d", r, gg, b, a)
	}
	r, _, _, a = pixAt(fx.be.Pixels(), 480, 40+180+4, 40)
	if a != 0 {
		t.Fatal("rich text painted outside its bounds")
	}

	ed.SetText("bold")
	ed.SetMark(0, 4, "bold")
	buf = render.NewCommandBuffer()
	ed.Draw(buf)
	boldCmds := 0
	for _, c := range textCmds(buf) {
		if len(c.Text.Glyphs) == len([]rune("bold")) {
			boldCmds++
		}
	}
	if boldCmds < 2 || !hasText(textCmds(buf), "bold", uimath.NewRect(40, 28, 180, 72)) {
		t.Fatalf("bold pass count %d", boldCmds)
	}

	ed.SetText("link")
	ed.SetMark(0, 4, "link")
	buf = render.NewCommandBuffer()
	ed.Draw(buf)
	link := textIn(textCmds(buf), "link", uimath.NewRect(40, 28, 180, 72))
	if link == nil || link.Color != ui.Config().PrimaryColor {
		t.Fatal("link text was not primary-colored glyphs")
	}

	ed.SetText("abcd")
	width := fx.drawer.MeasureText("ab", ui.Config().FontSizeSm)
	click(ui, 40+4+width+1, 28+8)
	if ed.Caret().Col < 2 {
		t.Fatalf("click caret %+v width %v", ed.Caret(), width)
	}

	var doc string
	for i := 0; i < 20; i++ {
		doc += "same\n"
	}
	doc += "RICHEND"
	ed.SetText(doc)
	buf = render.NewCommandBuffer()
	ed.Draw(buf)
	if hasText(textCmds(buf), "RICHEND", uimath.NewRect(0, 0, 480, 320)) {
		t.Fatal("rich editor drew a line below the viewport")
	}
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyEnd})
	for ed.Caret().Line < 20 {
		ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyArrowDown})
		if ed.Caret().Line > 40 {
			break
		}
	}
	buf = render.NewCommandBuffer()
	ed.Draw(buf)
	if !hasText(textCmds(buf), "RICHEND", uimath.NewRect(40, 28, 180, 72)) {
		t.Fatal("scrolled rich editor did not draw the last line")
	}
	id := ed.ElementID()
	ed.Destroy()
	if ui.Tree().Get(id) != nil {
		t.Fatal("rich editor survived destroy")
	}
}

func TestRichEditorMarkTable(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(480, 240)
	clip := &editorClip{}
	ui.Config().Platform = clip
	ed := widget.NewRichEditor(ui.Tree(), ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), ed.ElementID())
	place(ui.Tree(), ed.ElementID(), 16, 20, 180, 80)
	ui.Tree().SetFocused(ed.ElementID(), true)
	var changes int
	ed.OnChange(func(string) { changes++ })

	scripts := []struct {
		name    string
		text    string
		marks   []widget.RichMark
		ops     string
		want    string
		wantM   []widget.RichMark
		changes int
	}{
		{"start backspace", "abcd", []widget.RichMark{{2, 4, "bold"}}, "bs", "abcd", []widget.RichMark{{2, 4, "bold"}}, 0},
		{"insert shifts", "abcd", []widget.RichMark{{2, 4, "bold"}}, "type:X", "Xabcd", []widget.RichMark{{3, 5, "bold"}}, 1},
		{"delete marked middle", "ab\ncd\nef", []widget.RichMark{{3, 5, "bold"}}, "down shiftright shiftright bs", "ab\n\nef", nil, 1},
		{"trim overlapping mark", "abcdef", []widget.RichMark{{1, 5, "bold"}}, "right right shiftright shiftright bs", "abef", []widget.RichMark{{1, 3, "bold"}}, 1},
		{"shift later mark", "abcdef", []widget.RichMark{{4, 6, "code"}}, "right shiftright shiftright bs", "adef", []widget.RichMark{{2, 4, "code"}}, 1},
		{"keep earlier mark", "abcdef", []widget.RichMark{{0, 2, "link"}}, "right right right shiftright shiftright bs", "abcf", []widget.RichMark{{0, 2, "link"}}, 1},
		{"unicode mark", "a你b", []widget.RichMark{{1, 2, "bold"}}, "right right bs", "ab", nil, 1},
		{"unicode start", "你b", []widget.RichMark{{1, 2, "bold"}}, "bs", "你b", []widget.RichMark{{1, 2, "bold"}}, 0},
		{"delete eof", "ab", []widget.RichMark{{0, 2, "bold"}}, "end del", "ab", []widget.RichMark{{0, 2, "bold"}}, 0},
		{"repeat backspace", "ab", []widget.RichMark{{0, 1, "bold"}}, "end bs bs bs", "", nil, 2},
		{"empty paste", "ab", []widget.RichMark{{0, 2, "bold"}}, "paste:", "ab", []widget.RichMark{{0, 2, "bold"}}, 0},
		{"paste over mark", "ab\ncd\nef", []widget.RichMark{{3, 5, "bold"}}, "down shiftright shiftright paste:Z", "ab\nZ\nef", nil, 1},
		{"line merge keeps mark", "ab\ncd", []widget.RichMark{{3, 5, "bold"}}, "down bs", "abcd", []widget.RichMark{{2, 4, "bold"}}, 1},
		{"replace inside mark", "abcdef", []widget.RichMark{{0, 6, "bold"}}, "right right shiftright shiftright type:X", "abXef", []widget.RichMark{{0, 5, "bold"}}, 1},
		{"cleared selection at start", "ab", []widget.RichMark{{0, 2, "bold"}}, "shiftright shiftleft bs", "ab", []widget.RichMark{{0, 2, "bold"}}, 0},
		{"enter", "ab", nil, "right enter", "a\nb", nil, 1},
		{"toggle bold", "abcd", nil, "right right shiftright shiftright toggle:bold", "abcd", []widget.RichMark{{2, 4, "bold"}}, 0},
		{"toggle twice", "abcd", nil, "right right shiftright shiftright toggle:bold toggle:bold", "abcd", nil, 0},
	}
	for _, sc := range scripts {
		t.Run(sc.name, func(t *testing.T) {
			ed.SetText(sc.text)
			for _, m := range sc.marks {
				ed.SetMark(m.Start, m.End, m.Kind)
			}
			changes = 0
			ref := newRichDoc(sc.text, sc.marks)
			for _, op := range strings.Fields(sc.ops) {
				if strings.HasPrefix(op, "toggle:") {
					kind := strings.TrimPrefix(op, "toggle:")
					ref.toggle(kind)
					ed.Toggle(kind)
				} else {
					ref.apply(op, clip)
					ui.HandleEvent(editorOp(op, clip))
				}
				if ed.Text() != string(ref.text) || !sameMarks(ed.Marks(), ref.marks) {
					t.Fatalf("after %s text %q marks %+v want %q %+v", op, ed.Text(), ed.Marks(), string(ref.text), ref.marks)
				}
				if runeOff(ed.Text(), ed.Caret().Line, ed.Caret().Col) != ref.caret || ed.SelectedText() != ref.selected() {
					t.Fatalf("after %s caret %d sel %q want caret %d sel %q", op, runeOff(ed.Text(), ed.Caret().Line, ed.Caret().Col), ed.SelectedText(), ref.caret, ref.selected())
				}
			}
			if ed.Text() != sc.want || !sameMarks(ed.Marks(), sc.wantM) || changes != sc.changes {
				t.Fatalf("got %q %+v changes %d want %q %+v changes %d", ed.Text(), ed.Marks(), changes, sc.want, sc.wantM, sc.changes)
			}
		})
	}
}

type richDoc struct {
	text   []rune
	caret  int
	anchor int
	marks  []widget.RichMark
}

func newRichDoc(s string, marks []widget.RichMark) *richDoc {
	d := &richDoc{text: []rune(s)}
	d.marks = append([]widget.RichMark(nil), marks...)
	return d
}

func (d *richDoc) selected() string {
	a, b := d.anchor, d.caret
	if a > b {
		a, b = b, a
	}
	if a < 0 {
		a = 0
	}
	if b > len(d.text) {
		b = len(d.text)
	}
	if a >= b {
		return ""
	}
	return string(d.text[a:b])
}

func (d *richDoc) apply(op string, clip *editorClip) bool {
	switch {
	case op == "bs":
		return d.backspace()
	case op == "del":
		return d.deleteForward()
	case op == "left":
		d.move(-1, false)
	case op == "right":
		d.move(1, false)
	case op == "shiftleft":
		d.move(-1, true)
	case op == "shiftright":
		d.move(1, true)
	case op == "end":
		d.end(false)
	case op == "down":
		d.vert(1, false)
	case op == "enter":
		return d.insert([]rune("\n"))
	case strings.HasPrefix(op, "type:"):
		return d.insert([]rune(strings.TrimPrefix(op, "type:")))
	case strings.HasPrefix(op, "paste:"):
		clip.text = strings.TrimPrefix(op, "paste:")
		if clip.text == "" {
			return false
		}
		return d.insert([]rune(clip.text))
	}
	return false
}

func (d *richDoc) backspace() bool {
	if d.anchor != d.caret {
		return d.deleteRange()
	}
	if d.caret <= 0 {
		d.caret = 0
		d.anchor = 0
		return false
	}
	d.marks = cutRichRef(d.marks, d.caret-1, d.caret)
	d.text = append(append([]rune{}, d.text[:d.caret-1]...), d.text[d.caret:]...)
	d.caret--
	d.anchor = d.caret
	return true
}

func (d *richDoc) deleteForward() bool {
	if d.anchor != d.caret {
		return d.deleteRange()
	}
	if d.caret >= len(d.text) {
		return false
	}
	d.marks = cutRichRef(d.marks, d.caret, d.caret+1)
	d.text = append(append([]rune{}, d.text[:d.caret]...), d.text[d.caret+1:]...)
	d.anchor = d.caret
	return true
}

func (d *richDoc) deleteRange() bool {
	a, b := d.anchor, d.caret
	if a > b {
		a, b = b, a
	}
	if a < 0 {
		a = 0
	}
	if b > len(d.text) {
		b = len(d.text)
	}
	if a >= b {
		d.caret, d.anchor = a, a
		return false
	}
	d.marks = cutRichRef(d.marks, a, b)
	d.text = append(append([]rune{}, d.text[:a]...), d.text[b:]...)
	d.caret, d.anchor = a, a
	return true
}

func (d *richDoc) insert(s []rune) bool {
	if len(s) == 0 {
		return false
	}
	if d.anchor != d.caret {
		d.deleteRange()
	}
	d.marks = shiftRichRef(d.marks, d.caret, len(s))
	out := append(append([]rune{}, d.text[:d.caret]...), s...)
	out = append(out, d.text[d.caret:]...)
	d.text = out
	d.caret += len(s)
	d.anchor = d.caret
	return true
}

func (d *richDoc) toggle(kind string) {
	a, b := d.anchor, d.caret
	if a > b {
		a, b = b, a
	}
	if a >= b {
		return
	}
	if richCovered(d.marks, kind, a, b) {
		var keep, drop []widget.RichMark
		for _, m := range d.marks {
			if m.Kind == kind {
				drop = append(drop, m)
			} else {
				keep = append(keep, m)
			}
		}
		d.marks = append(keep, cutRichRef(drop, a, b)...)
		return
	}
	d.marks = append(d.marks, widget.RichMark{Start: a, End: b, Kind: kind})
}

func (d *richDoc) move(dir int, keep bool) {
	d.caret += dir
	if d.caret < 0 {
		d.caret = 0
	}
	if d.caret > len(d.text) {
		d.caret = len(d.text)
	}
	if !keep {
		d.anchor = d.caret
	}
}

func (d *richDoc) end(keep bool) {
	for d.caret < len(d.text) && d.text[d.caret] != '\n' {
		d.caret++
	}
	if !keep {
		d.anchor = d.caret
	}
}

func (d *richDoc) vert(dir int, keep bool) {
	line, col := 0, 0
	for i := 0; i < d.caret && i < len(d.text); i++ {
		if d.text[i] == '\n' {
			line++
			col = 0
			continue
		}
		col++
	}
	line += dir
	if line < 0 {
		line = 0
	}
	off := 0
	cur := 0
	for cur < line && off < len(d.text) {
		if d.text[off] == '\n' {
			cur++
		}
		off++
	}
	n := 0
	for off+n < len(d.text) && d.text[off+n] != '\n' && n < col {
		n++
	}
	d.caret = off + n
	if !keep {
		d.anchor = d.caret
	}
}

func shiftRichRef(marks []widget.RichMark, at, n int) []widget.RichMark {
	if n == 0 {
		return marks
	}
	out := make([]widget.RichMark, 0, len(marks))
	for _, m := range marks {
		switch {
		case m.End <= at:
			out = append(out, m)
		case m.Start >= at:
			m.Start += n
			m.End += n
			out = append(out, m)
		default:
			m.End += n
			out = append(out, m)
		}
	}
	return out
}

func cutRichRef(marks []widget.RichMark, a, b int) []widget.RichMark {
	if a >= b {
		return marks
	}
	n := b - a
	out := make([]widget.RichMark, 0, len(marks))
	for _, m := range marks {
		ns, ne := m.Start, m.End
		if ne <= a {
			out = append(out, m)
			continue
		}
		if ns >= b {
			m.Start = ns - n
			m.End = ne - n
			if m.Start < m.End {
				out = append(out, m)
			}
			continue
		}
		if ns < a {
			ns = m.Start
		} else {
			ns = a
		}
		if ne >= b {
			ne -= n
		} else {
			ne = a
		}
		if ns < ne {
			m.Start, m.End = ns, ne
			out = append(out, m)
		}
	}
	return out
}

func richCovered(marks []widget.RichMark, kind string, a, b int) bool {
	for _, m := range marks {
		if m.Kind == kind && m.Start <= a && m.End >= b {
			return true
		}
	}
	return false
}

func sameMarks(a, b []widget.RichMark) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestFormulaLayoutGlyphsAndHit(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(480, 320)
	fx := newGlyphFixture(t, 480, 320)
	ui.Config().TextRenderer = fx.drawer
	f := widget.NewFormula(ui.Tree(), ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), f.ElementID())
	place(ui.Tree(), f.ElementID(), 30, 24, 200, 80)
	if err := f.SetExpr("frac(1,2)+x"); err != nil {
		t.Fatal(err)
	}
	num, den := f.AtomScreen("1"), f.AtomScreen("2")
	if num.Y >= den.Y || num.X < 30 || den.X < 30 {
		t.Fatalf("fraction layout num %v den %v", num, den)
	}
	buf := render.NewCommandBuffer()
	f.Draw(buf)
	cmds := textCmds(buf)
	if hasText(cmds, "frac(1,2)+x", uimath.NewRect(0, 0, 480, 320)) {
		t.Fatal("formula drew the source string")
	}
	if !hasText(cmds, "1", num) || !hasText(cmds, "2", den) || !hasText(cmds, "x", f.AtomScreen("x")) {
		t.Fatal("formula atoms were not drawn as glyphs")
	}
	fx.present(buf)
	g := textIn(cmds, "2", den).Glyphs[0]
	r, gg, b, a := pixAt(fx.be.Pixels(), 480, int(g.X+g.Width/2), int(g.Y+g.Height/2))
	if a == 0 || (r > 80 && gg > 80 && b > 80) {
		t.Fatalf("denominator glyph %d %d %d %d", r, gg, b, a)
	}
	text, role := f.Hit(den.X+2, den.Y+2)
	if text != "2" || role != "den" {
		t.Fatalf("hit %q %q", text, role)
	}
	r, _, _, a = pixAt(fx.be.Pixels(), 480, 30+200+4, 40)
	if a != 0 {
		t.Fatal("formula painted outside its bounds")
	}

	if err := f.SetExpr("sup(x,2)"); err != nil {
		t.Fatal(err)
	}
	base, exp := f.AtomScreen("x"), f.AtomScreen("2")
	if exp.Y >= base.Y {
		t.Fatalf("superscript exp %v base %v", exp, base)
	}
	buf = render.NewCommandBuffer()
	f.Draw(buf)
	if hasText(textCmds(buf), "sup(x,2)", uimath.NewRect(0, 0, 480, 320)) || !hasText(textCmds(buf), "x", base) {
		t.Fatal("superscript was not laid out as glyphs")
	}
	_, role = f.Hit(exp.X+1, exp.Y+1)
	if role != "exp" {
		t.Fatalf("superscript hit role %q", role)
	}

	if err := f.SetExpr("sub(y,3)"); err != nil {
		t.Fatal(err)
	}
	base, sub := f.AtomScreen("y"), f.AtomScreen("3")
	if sub.Y <= base.Y {
		t.Fatalf("subscript sub %v base %v", sub, base)
	}

	if err := f.SetExpr("sqrt(4)"); err != nil {
		t.Fatal(err)
	}
	buf = render.NewCommandBuffer()
	f.Draw(buf)
	if hasText(textCmds(buf), "sqrt(4)", uimath.NewRect(0, 0, 480, 320)) || !hasText(textCmds(buf), "4", f.AtomScreen("4")) {
		t.Fatal("radical content was not a glyph")
	}

	if err := f.SetExpr("frac(1)"); err == nil {
		t.Fatal("incomplete fraction was accepted")
	}

	var wide strings.Builder
	for i := 0; i < 8; i++ {
		if i > 0 {
			wide.WriteByte('+')
		}
		wide.WriteString("sup(n")
		wide.WriteByte(byte('a' + i))
		wide.WriteString(",2)")
	}
	wide.WriteString("+sup(LAST,2)")
	if err := f.SetExpr(wide.String()); err != nil {
		t.Fatal(err)
	}
	place(ui.Tree(), f.ElementID(), 30, 24, 70, 60)
	buf = render.NewCommandBuffer()
	f.Draw(buf)
	if hasText(textCmds(buf), "LAST", uimath.NewRect(0, 0, 480, 320)) {
		t.Fatal("formula atom outside the viewport was painted")
	}
	off := f.AtomScreen("LAST")
	f.SetScrollX(off.X - 30)
	buf = render.NewCommandBuffer()
	f.Draw(buf)
	if !hasText(textCmds(buf), "LAST", uimath.NewRect(30, 24, 70, 60)) {
		t.Fatal("scrolled formula did not draw the last atom")
	}
	fx.present(buf)
	r, _, _, a = pixAt(fx.be.Pixels(), 480, 30+70+4, 40)
	if a != 0 {
		t.Fatal("scrolled formula painted outside its bounds")
	}
	id := f.ElementID()
	f.Destroy()
	if ui.Tree().Get(id) != nil {
		t.Fatal("formula survived destroy")
	}
}

func TestHTMLRichEditorAndFormula(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<richeditor></richeditor><formula></formula>`, "")
	if len(doc.QueryByTag("richeditor")) != 1 || len(doc.QueryByTag("formula")) != 1 {
		t.Fatalf("rich %d formula %d", len(doc.QueryByTag("richeditor")), len(doc.QueryByTag("formula")))
	}
}

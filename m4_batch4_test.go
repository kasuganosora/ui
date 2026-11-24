package ui

import (
	"strings"
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
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

func TestEditorBoundaryEditsMatchRuneModel(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(480, 240)
	clip := &editorClip{}
	ui.Config().Platform = clip
	ed := widget.NewEditor(ui.Tree(), ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), ed.ElementID())
	place(ui.Tree(), ed.ElementID(), 16, 20, 180, 80)
	ui.Tree().SetFocused(ed.ElementID(), true)
	var changes int
	ed.OnChange(func(string) { changes++ })

	scripts := []struct {
		name string
		text string
		ops  string
	}{
		{"start ascii", "ab", "bs"},
		{"start unicode", "你b", "bs"},
		{"empty", "", "bs"},
		{"repeat start", "ab", "end bs bs bs"},
		{"delete eof", "ab", "end del"},
		{"delete empty", "", "del"},
		{"delete mid then eof", "ab", "right del del"},
		{"line merge", "ab\ncd", "down bs"},
		{"real selection", "ab", "shiftright bs"},
		{"cleared selection at start", "ab", "shiftright shiftleft bs"},
		{"empty paste", "ab", "paste:"},
		{"paste", "ab", "paste:Z"},
		{"type", "ab", "type:X"},
		{"enter", "ab", "right enter"},
		{"unicode backspace", "a你b", "right right bs"},
	}
	for _, sc := range scripts {
		t.Run(sc.name, func(t *testing.T) {
			ed.SetText(sc.text)
			changes = 0
			ref := newRuneDoc(sc.text)
			for _, op := range strings.Fields(sc.ops) {
				before := string(ref.text)
				changed := ref.apply(op, clip)
				ui.HandleEvent(editorOp(op, clip))
				if string(ref.text) != ed.Text() {
					t.Fatalf("after %s text %q want %q", op, ed.Text(), string(ref.text))
				}
				if runeOff(ed.Text(), ed.Caret().Line, ed.Caret().Col) != ref.caret || ed.SelectedText() != ref.selected() {
					t.Fatalf("after %s caret %d sel %q want caret %d sel %q", op, runeOff(ed.Text(), ed.Caret().Line, ed.Caret().Col), ed.SelectedText(), ref.caret, ref.selected())
				}
				if changed == (ed.Text() == before) {
					t.Fatalf("after %s change flag %v text %q before %q", op, changed, ed.Text(), before)
				}
			}
			wantN := 0
			tmp := newRuneDoc(sc.text)
			for _, op := range strings.Fields(sc.ops) {
				if tmp.apply(op, clip) {
					wantN++
				}
			}
			if changes != wantN {
				t.Fatalf("OnChange %d want %d text %q", changes, wantN, ed.Text())
			}
		})
	}
}

type runeDoc struct {
	text   []rune
	caret  int
	anchor int
}

func newRuneDoc(s string) *runeDoc {
	d := &runeDoc{text: []rune(s)}
	return d
}

func (d *runeDoc) selected() string {
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

func (d *runeDoc) apply(op string, clip *editorClip) bool {
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
	case op == "home":
		d.home(false)
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
	default:
		return false
	}
	return false
}

func (d *runeDoc) backspace() bool {
	if d.anchor != d.caret {
		return d.deleteRange()
	}
	if d.caret <= 0 {
		d.caret = 0
		d.anchor = 0
		return false
	}
	d.text = append(append([]rune{}, d.text[:d.caret-1]...), d.text[d.caret:]...)
	d.caret--
	d.anchor = d.caret
	return true
}

func (d *runeDoc) deleteForward() bool {
	if d.anchor != d.caret {
		return d.deleteRange()
	}
	if d.caret >= len(d.text) {
		return false
	}
	d.text = append(append([]rune{}, d.text[:d.caret]...), d.text[d.caret+1:]...)
	d.anchor = d.caret
	return true
}

func (d *runeDoc) deleteRange() bool {
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
	d.text = append(append([]rune{}, d.text[:a]...), d.text[b:]...)
	d.caret, d.anchor = a, a
	return true
}

func (d *runeDoc) insert(s []rune) bool {
	if len(s) == 0 {
		return false
	}
	if d.anchor != d.caret {
		d.deleteRange()
	}
	out := append(append([]rune{}, d.text[:d.caret]...), s...)
	out = append(out, d.text[d.caret:]...)
	d.text = out
	d.caret += len(s)
	d.anchor = d.caret
	return true
}

func (d *runeDoc) move(dir int, keep bool) {
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

func (d *runeDoc) home(keep bool) {
	for d.caret > 0 && d.text[d.caret-1] != '\n' {
		d.caret--
	}
	if !keep {
		d.anchor = d.caret
	}
}

func (d *runeDoc) end(keep bool) {
	for d.caret < len(d.text) && d.text[d.caret] != '\n' {
		d.caret++
	}
	if !keep {
		d.anchor = d.caret
	}
}

func (d *runeDoc) vert(dir int, keep bool) {
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

func editorOp(op string, clip *editorClip) *event.Event {
	ev := &event.Event{Type: event.KeyDown}
	switch {
	case op == "bs":
		ev.Key = event.KeyBackspace
	case op == "del":
		ev.Key = event.KeyDelete
	case op == "left":
		ev.Key = event.KeyArrowLeft
	case op == "right":
		ev.Key = event.KeyArrowRight
	case op == "shiftleft":
		ev.Key = event.KeyArrowLeft
		ev.Modifiers.Shift = true
	case op == "shiftright":
		ev.Key = event.KeyArrowRight
		ev.Modifiers.Shift = true
	case op == "down":
		ev.Key = event.KeyArrowDown
	case op == "end":
		ev.Key = event.KeyEnd
	case op == "home":
		ev.Key = event.KeyHome
	case op == "enter":
		ev.Key = event.KeyEnter
	case strings.HasPrefix(op, "type:"):
		ev.Type = event.KeyPress
		ev.Char = []rune(strings.TrimPrefix(op, "type:"))[0]
	case strings.HasPrefix(op, "paste:"):
		clip.text = strings.TrimPrefix(op, "paste:")
		ev.Key = event.KeyV
		ev.Modifiers.Ctrl = true
	}
	return ev
}

func runeOff(text string, line, col int) int {
	lines := strings.Split(text, "\n")
	n := 0
	for i := 0; i < line && i < len(lines); i++ {
		n += len([]rune(lines[i])) + 1
	}
	n += col
	return n
}

func TestEditorDeleteSelectionPreservesNewlines(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(480, 240)
	ed := widget.NewEditor(ui.Tree(), ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), ed.ElementID())
	place(ui.Tree(), ed.ElementID(), 16, 20, 180, 80)
	ui.Tree().SetFocused(ed.ElementID(), true)
	var changes int
	var last string
	ed.OnChange(func(s string) {
		changes++
		last = s
	})

	type step struct {
		name string
		text string
		keys []event.Event
		want string
		line int
		col  int
	}
	right := event.Event{Type: event.KeyDown, Key: event.KeyArrowRight}
	down := event.Event{Type: event.KeyDown, Key: event.KeyArrowDown}
	shiftR := event.Event{Type: event.KeyDown, Key: event.KeyArrowRight, Modifiers: event.Modifiers{Shift: true}}
	shiftL := event.Event{Type: event.KeyDown, Key: event.KeyArrowLeft, Modifiers: event.Modifiers{Shift: true}}
	shiftD := event.Event{Type: event.KeyDown, Key: event.KeyArrowDown, Modifiers: event.Modifiers{Shift: true}}
	back := event.Event{Type: event.KeyDown, Key: event.KeyBackspace}
	del := event.Event{Type: event.KeyDown, Key: event.KeyDelete}
	end := event.Event{Type: event.KeyDown, Key: event.KeyEnd}
	enter := event.Event{Type: event.KeyDown, Key: event.KeyEnter}
	paste := event.Event{Type: event.KeyDown, Key: event.KeyV, Modifiers: event.Modifiers{Ctrl: true}}

	check := func(s step) {
		t.Helper()
		ed.SetText(s.text)
		changes = 0
		for _, k := range s.keys {
			ui.HandleEvent(&k)
		}
		if ed.Text() != s.want || last != s.want {
			t.Fatalf("%s text %q last %q want %q", s.name, ed.Text(), last, s.want)
		}
		if changes != 1 {
			t.Fatalf("%s OnChange %d", s.name, changes)
		}
		if ed.Caret().Line != s.line || ed.Caret().Col != s.col || ed.SelectedText() != "" {
			t.Fatalf("%s caret %+v sel %q", s.name, ed.Caret(), ed.SelectedText())
		}
	}
	check(step{"middle", "ab\ncd\nef", []event.Event{down, shiftR, shiftR, back}, "ab\n\nef", 1, 0})
	check(step{"first", "ab\ncd\nef", []event.Event{shiftR, back}, "b\ncd\nef", 0, 0})
	check(step{"last", "ab\ncd\nef", []event.Event{down, down, shiftR, back}, "ab\ncd\nf", 2, 0})
	check(step{"reverse", "ab\ncd\nef", []event.Event{down, end, shiftL, shiftL, back}, "ab\n\nef", 1, 0})
	check(step{"unicode", "a你b\ncd", []event.Event{right, shiftR, back}, "ab\ncd", 0, 1})
	check(step{"replace", "ab\ncd\nef", []event.Event{down, shiftR, shiftR, {Type: event.KeyPress, Char: 'X'}}, "ab\nX\nef", 1, 1})
	check(step{"span", "ab\ncd\nef", []event.Event{right, shiftD, shiftD, back}, "af", 0, 1})
	check(step{"end col0", "ab\ncd\nef", []event.Event{down, shiftL, back}, "abcd\nef", 0, 2})
	check(step{"empty", "ab\ncd\nef", []event.Event{down, shiftR, shiftL, back}, "abcd\nef", 0, 2})
	check(step{"merge", "ab\ncd\nef", []event.Event{down, back}, "abcd\nef", 0, 2})
	check(step{"delete", "ab\ncd\nef", []event.Event{end, del}, "abcd\nef", 0, 2})
	check(step{"enter", "ab\ncd", []event.Event{right, enter}, "a\nb\ncd", 1, 0})

	clip := &editorClip{text: "Z"}
	ui.Config().Platform = clip
	check(step{"paste", "ab\ncd\nef", []event.Event{down, shiftR, shiftR, paste}, "ab\nZ\nef", 1, 1})
}

type editorClip struct{ text string }

func (c *editorClip) Init() error { return nil }
func (c *editorClip) CreateWindow(platform.WindowOptions) (platform.Window, error) {
	return nil, nil
}
func (c *editorClip) PollEvents() []event.Event     { return nil }
func (c *editorClip) ProcessMessages()              {}
func (c *editorClip) GetClipboardText() string      { return c.text }
func (c *editorClip) SetClipboardText(s string)     { c.text = s }
func (c *editorClip) GetPrimaryMonitorDPI() float32 { return 1 }
func (c *editorClip) GetSystemLocale() string       { return "en" }
func (c *editorClip) Terminate()                    {}

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

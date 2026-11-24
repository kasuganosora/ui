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

type chatStub struct {
	n    int
	last string
	err  error
}

func (s *chatStub) Send(text string) error {
	s.n++
	s.last = text
	return s.err
}

func TestChatMessagesGlyphsScrollAndSend(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(480, 320)
	fx := newGlyphFixture(t, 480, 320)
	ui.Config().TextRenderer = fx.drawer
	ch := widget.NewChat(ui.Tree(), ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), ch.ElementID())
	place(ui.Tree(), ch.ElementID(), 36, 22, 220, 140)
	ui.Tree().SetFocused(ch.ElementID(), true)
	stub := &chatStub{}
	ch.SetSender(stub)

	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyEnter})
	if stub.n != 0 || len(ch.Messages()) != 0 || ch.Draft() != "" {
		t.Fatalf("empty send n %d messages %d draft %q", stub.n, len(ch.Messages()), ch.Draft())
	}
	for _, r := range "HELLOMSG" {
		ui.HandleEvent(&event.Event{Type: event.KeyPress, Char: r})
	}
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyEnter})
	if stub.n != 1 || stub.last != "HELLOMSG" || ch.Draft() != "" || len(ch.Messages()) != 1 || ch.Messages()[0].Text != "HELLOMSG" || !ch.Messages()[0].Mine {
		t.Fatalf("send %+v draft %q", ch.Messages(), ch.Draft())
	}
	stub.err = errChat
	ui.HandleEvent(&event.Event{Type: event.KeyPress, Char: 'Z'})
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyEnter})
	if stub.n != 2 || len(ch.Messages()) != 1 || ch.Draft() != "Z" {
		t.Fatalf("failed send kept state n %d messages %d draft %q", stub.n, len(ch.Messages()), ch.Draft())
	}
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyBackspace})
	if ch.Draft() != "" || stub.n != 2 {
		t.Fatalf("backspace draft %q n %d", ch.Draft(), stub.n)
	}
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyBackspace})
	if ch.Draft() != "" || stub.n != 2 {
		t.Fatal("backspace at an empty draft changed state")
	}

	ch.Append(widget.ChatMessage{ID: "other", Author: "you", Text: "OTHERSIDE", Mine: false})
	if stub.n != 2 || len(ch.Messages()) != 2 {
		t.Fatal("append notified the sender or dropped the local message")
	}
	buf := render.NewCommandBuffer()
	ch.Draw(buf)
	mine := ch.MessageScreen("m1")
	other := ch.MessageScreen("other")
	if mine.X < 36 || other.Y <= mine.Y {
		t.Fatalf("message layout mine %v other %v", mine, other)
	}
	cmds := textCmds(buf)
	if !hasText(cmds, "HELLOMSG", mine) || !hasText(cmds, "OTHERSIDE", other) {
		t.Fatal("chat messages were not drawn as glyphs")
	}
	mg := textIn(cmds, "HELLOMSG", mine).Glyphs[0]
	og := textIn(cmds, "OTHERSIDE", other).Glyphs[0]
	if mg.X <= og.X {
		t.Fatalf("own message x %v other x %v", mg.X, og.X)
	}
	fx.present(buf)
	r, g, b, a := pixAt(fx.be.Pixels(), 480, int(mg.X+mg.Width/2), int(mg.Y+mg.Height/2))
	if a == 0 || (r > 80 && g > 80 && b > 80) {
		t.Fatalf("chat glyph %d %d %d %d", r, g, b, a)
	}
	r, _, _, a = pixAt(fx.be.Pixels(), 480, 36+220+4, 40)
	if a != 0 {
		t.Fatal("chat painted outside its bounds")
	}
	click(ui, other.X+4, other.Y+4)
	if ch.SelectedID() != "other" {
		t.Fatalf("selected %q", ch.SelectedID())
	}

	place(ui.Tree(), ch.ElementID(), 36, 22, 220, 112)
	ch.Append(widget.ChatMessage{ID: "cross", Author: "you", Text: "CROSSING"})
	buf = render.NewCommandBuffer()
	ch.Draw(buf)
	clip, ok := textClip(buf, "CROSSING", uimath.NewRect(36, 22, 220, 80))
	if !ok || clip.Y+clip.Height > 22+80.5 {
		t.Fatalf("message clip %v ok %v", clip, ok)
	}

	var many []widget.ChatMessage
	for i := 0; i < 12; i++ {
		many = append(many, widget.ChatMessage{ID: "row", Author: "you", Text: "same"})
	}
	many = append(many, widget.ChatMessage{ID: "end", Author: "you", Text: "CHATEND"})
	ch.SetMessages(many)
	place(ui.Tree(), ch.ElementID(), 36, 22, 180, 90)
	buf = render.NewCommandBuffer()
	ch.Draw(buf)
	if hasText(textCmds(buf), "CHATEND", uimath.NewRect(0, 0, 480, 320)) {
		t.Fatal("chat drew a message below the list")
	}
	off := ch.MessageScreen("end")
	ch.SetScrollY(off.Y - 22)
	buf = render.NewCommandBuffer()
	ch.Draw(buf)
	if !hasText(textCmds(buf), "CHATEND", uimath.NewRect(36, 22, 180, 58)) {
		t.Fatal("scrolled chat did not draw the last message")
	}
	id := ch.ElementID()
	ch.Destroy()
	if ui.Tree().Get(id) != nil {
		t.Fatal("chat survived destroy")
	}
}

var errChat = stubErr("down")

type stubErr string

func (e stubErr) Error() string { return string(e) }

func TestChatSendMatchesDraft(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(320, 200)
	ch := widget.NewChat(ui.Tree(), ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), ch.ElementID())
	place(ui.Tree(), ch.ElementID(), 8, 8, 200, 120)
	ui.Tree().SetFocused(ch.ElementID(), true)
	stub := &chatStub{}
	ch.SetSender(stub)
	scripts := []struct {
		name  string
		ops   string
		draft string
		texts string
		sends int
		fail  bool
	}{
		{"empty enter", "enter", "", "", 0, false},
		{"send ascii", "type:h type:i enter", "", "hi", 1, false},
		{"send unicode", "type:你 enter", "", "你", 1, false},
		{"backspace start", "bs", "", "", 0, false},
		{"type backspace", "type:Z bs", "", "", 0, false},
		{"failed send", "type:n type:o enter", "no", "", 1, true},
	}
	for _, sc := range scripts {
		t.Run(sc.name, func(t *testing.T) {
			ch.SetMessages(nil)
			stub.n = 0
			stub.err = nil
			if sc.fail {
				stub.err = errChat
			}
			for _, op := range strings.Fields(sc.ops) {
				ui.HandleEvent(editorOp(op, &editorClip{}))
			}
			if ch.Draft() != sc.draft || stub.n != sc.sends {
				t.Fatalf("draft %q sends %d", ch.Draft(), stub.n)
			}
			got := ""
			for i, m := range ch.Messages() {
				if i > 0 {
					got += ","
				}
				got += m.Text
			}
			if got != sc.texts {
				t.Fatalf("messages %q", got)
			}
		})
	}
}

func TestDockRestoreGeometryAndIDs(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(480, 320)
	fx := newGlyphFixture(t, 480, 320)
	ui.Config().TextRenderer = fx.drawer
	dl := widget.NewDockLayout(ui.Tree(), ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), dl.ElementID())
	place(ui.Tree(), dl.ElementID(), 40, 28, 360, 240)
	btn := widget.NewButton(ui.Tree(), "inside", ui.Config())
	dl.AddPanel(&widget.DockPanel{ID: "left", Title: "LEFTDOCK", Content: btn, Position: widget.DockLeft, DockSize: 80})
	dl.AddPanel(&widget.DockPanel{ID: "main", Title: "CENTERTAB", Position: widget.DockCenter})
	dl.AddPanel(&widget.DockPanel{ID: "side", Title: "SIDEPANE", Position: widget.DockCenter})
	dl.AddPanel(&widget.DockPanel{ID: "float", Title: "FLOATWINDOW", Position: widget.DockFloat, FloatX: 80, FloatY: 14, FloatW: 30, FloatH: 16})
	dl.SetActiveTab(1)
	if !dl.SetScale(2) || dl.PanelScreen("left").Width != 160 || dl.PanelScreen("left").X != 40 {
		t.Fatalf("scaled left %v", dl.PanelScreen("left"))
	}
	floatBox := dl.PanelScreen("float")
	if floatBox.X != 160 || floatBox.Y != 28 || floatBox.Width != 60 || floatBox.Height != 32 {
		t.Fatalf("scaled float %v", floatBox)
	}
	raw, err := dl.MarshalLayout()
	if err != nil {
		t.Fatal(err)
	}
	dl.SetActiveTab(0)
	dl.SetScale(1)
	dl.FindPanel("left").DockSize = 10
	if err := dl.UnmarshalLayout(raw); err != nil {
		t.Fatal(err)
	}
	if dl.ActiveID() != "side" || dl.FindPanel("left").Content != btn || dl.FindPanel("left").DockSize != 80 {
		t.Fatalf("restore id %q size %v content %v", dl.ActiveID(), dl.FindPanel("left").DockSize, dl.FindPanel("left").Content != nil)
	}
	if dl.PanelScreen("left").Width != 160 || dl.PanelScreen("float") != floatBox {
		t.Fatalf("restore geom left %v float %v", dl.PanelScreen("left"), dl.PanelScreen("float"))
	}
	buf := render.NewCommandBuffer()
	dl.Draw(buf)
	left := dl.PanelScreen("left")
	if !hasText(textCmds(buf), "LEFTDOCK", left) {
		t.Fatal("dock title was not drawn as glyphs")
	}
	clip, ok := textClip(buf, "LEFTDOCK", left)
	if !ok || clip.X+clip.Width > left.X+left.Width+0.5 {
		t.Fatalf("dock title clip %v region %v", clip, left)
	}
	fx.present(buf)
	g := textIn(textCmds(buf), "LEFTDOCK", left).Glyphs[0]
	r, gg, b, a := pixAt(fx.be.Pixels(), 480, int(g.X+g.Width/2), int(g.Y+g.Height/2))
	if a == 0 || (r > 80 && gg > 80 && b > 80) {
		t.Fatalf("dock glyph %d %d %d %d", r, gg, b, a)
	}
	r, _, _, a = pixAt(fx.be.Pixels(), 480, 40+360+4, 40)
	if a != 0 {
		t.Fatal("dock painted outside its bounds")
	}

	bad := [][]byte{
		[]byte(`{"Panels":[{"ID":"left","Position":0,"DockSize":80},{"ID":"left","Position":4,"DockSize":10}],"Scale":2}`),
		[]byte(`{"Panels":[{"ID":"","Position":0,"DockSize":80}],"Scale":1}`),
		[]byte(`{"Panels":[{"ID":"left","Position":0,"DockSize":-5}],"Scale":1}`),
		[]byte(`{"Panels":[{"ID":"left","Position":0,"DockSize":80}],"ActiveID":"missing","Scale":1}`),
		[]byte(`{"Panels":[{"ID":"left","Position":0,"DockSize":80}],"Scale":0}`),
		[]byte(`{`),
	}
	for _, payload := range bad {
		before := dl.PanelScreen("left").Width
		if err := dl.UnmarshalLayout(payload); err == nil {
			t.Fatalf("accepted %s", payload)
		}
		if dl.PanelScreen("left").Width != before || dl.ActiveID() != "side" || dl.FindPanel("left").Content != btn {
			t.Fatal("rejected layout changed the current dock")
		}
	}
	id := dl.ElementID()
	dl.Destroy()
	if ui.Tree().Get(id) != nil {
		t.Fatal("dock survived destroy")
	}
}

func TestTerminalBytesKeysAndScroll(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(480, 320)
	fx := newGlyphFixture(t, 480, 320)
	ui.Config().TextRenderer = fx.drawer
	term := widget.NewTerminal(ui.Tree(), ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), term.ElementID())
	place(ui.Tree(), term.ElementID(), 24, 18, 200, 80)
	ui.Tree().SetFocused(term.ElementID(), true)
	var got []byte
	term.OnBytes(func(b []byte) { got = append(got, b...) })

	you := []byte("你")
	if n, err := term.Write(you[:1]); n != 1 || err != nil || term.Text() != "" {
		t.Fatalf("partial write n %d err %v text %q", n, err, term.Text())
	}
	if _, err := term.Write(you[1:]); err != nil || term.Text() != "你" {
		t.Fatalf("utf8 text %q", term.Text())
	}
	if _, err := term.Write([]byte("\nkeep\n")); err != nil {
		t.Fatal(err)
	}
	if _, err := term.Write([]byte{0x1b, '[', '2', 'J'}); err != nil {
		t.Fatal(err)
	}
	if term.Text() != "你\nkeep\n\x1b[2J" {
		t.Fatalf("stream text %q", term.Text())
	}
	buf := render.NewCommandBuffer()
	term.Draw(buf)
	if !hasText(textCmds(buf), "keep", uimath.NewRect(24, 18+18, 200, 18)) {
		t.Fatal("terminal stream was not drawn as glyphs")
	}
	fx.present(buf)
	g := textIn(textCmds(buf), "keep", uimath.NewRect(24, 18+18, 200, 18)).Glyphs[0]
	r, gg, b, a := pixAt(fx.be.Pixels(), 480, int(g.X+g.Width/2), int(g.Y+g.Height/2))
	if a == 0 || (r > 80 && gg > 80 && b > 80) {
		t.Fatalf("terminal glyph %d %d %d %d", r, gg, b, a)
	}
	r, _, _, a = pixAt(fx.be.Pixels(), 480, 24+200+4, 30)
	if a != 0 {
		t.Fatal("terminal painted outside its bounds")
	}
	before := term.Text()
	ui.HandleEvent(&event.Event{Type: event.KeyPress, Char: 'Z'})
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyEnter})
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyBackspace})
	if term.Text() != before || string(got) != "Z\n\x7f" {
		t.Fatalf("keys text %q routed %q", term.Text(), got)
	}

	var stream string
	for i := 0; i < 20; i++ {
		stream += "same\n"
	}
	stream += "TERMEND\n"
	term.Clear()
	if _, err := term.Write([]byte(stream)); err != nil {
		t.Fatal(err)
	}
	buf = render.NewCommandBuffer()
	term.Draw(buf)
	if hasText(textCmds(buf), "TERMEND", uimath.NewRect(0, 0, 480, 320)) {
		t.Fatal("terminal drew a line below the viewport")
	}
	term.SetScrollY(20 * 18)
	buf = render.NewCommandBuffer()
	term.Draw(buf)
	if !hasText(textCmds(buf), "TERMEND", uimath.NewRect(24, 18, 200, 80)) {
		t.Fatal("scrolled terminal did not draw the last line")
	}
	id := term.ElementID()
	term.Destroy()
	if ui.Tree().Get(id) != nil {
		t.Fatal("terminal survived destroy")
	}
}

func TestHTMLChatDockAndTerminal(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := LoadHTMLDocument(tree, cfg, `<chat></chat><docklayout></docklayout><terminal></terminal>`, "")
	if len(doc.QueryByTag("chat")) != 1 || len(doc.QueryByTag("docklayout")) != 1 || len(doc.QueryByTag("terminal")) != 1 {
		t.Fatalf("chat %d dock %d terminal %d", len(doc.QueryByTag("chat")), len(doc.QueryByTag("docklayout")), len(doc.QueryByTag("terminal")))
	}
}

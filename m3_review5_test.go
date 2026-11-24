package ui

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/godot"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

func softwarePixel(buf *render.CommandBuffer, w, h, x, y int) (byte, byte, byte, byte) {
	be := godot.NewSoftwareBackend()
	be.Resize(w, h)
	be.BeginFrame()
	be.Submit(buf)
	px := be.Pixels()
	if x < 0 || y < 0 || x >= w || y >= h || len(px) < (y*w+x+1)*4 {
		return 0, 0, 0, 0
	}
	i := (y*w + x) * 4
	return px[i], px[i+1], px[i+2], px[i+3]
}

func TestHoverCardEscapeWithoutStealingFocus(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(400, 300)
	in := widget.NewInput(ui.Tree(), ui.Config())
	card := widget.NewHoverCard(ui.Tree(), ui.Config())
	card.SetDelay(0.1)
	ui.Tree().AppendChild(ui.Tree().Root(), in.ElementID())
	ui.Tree().AppendChild(ui.Tree().Root(), card.ElementID())
	place(ui.Tree(), in.ElementID(), 10, 200, 120, 28)
	place(ui.Tree(), card.ElementID(), 10, 10, 48, 24)
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: 20, GlobalY: 210})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 20, GlobalY: 210})
	if ui.Tree().Focused() != in.ElementID() {
		t.Fatal("input did not take focus")
	}
	ui.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: 20, GlobalY: 16})
	card.Tick(0.2)
	if !card.IsOpen() || ui.Tree().Focused() != in.ElementID() {
		t.Fatalf("hover open %v focus %d", card.IsOpen(), ui.Tree().Focused())
	}
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyEscape})
	if card.IsOpen() || ui.Tree().Focused() != in.ElementID() {
		t.Fatalf("after escape open %v focus %d", card.IsOpen(), ui.Tree().Focused())
	}
	ui.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: 20, GlobalY: 16})
	card.Tick(0.2)
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: 300, GlobalY: 260})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 300, GlobalY: 260})
	if card.IsOpen() {
		t.Fatal("outside click left the hover card open")
	}
	card.Destroy()
	if ui.Tree().HitTest(20, 80) == card.ElementID() {
		t.Fatal("destroyed hover card still hits")
	}
}

func TestSettingsCategoryOpensLiveRows(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(480, 320)
	dlg := widget.NewSettingsDialog(ui.Tree(), ui.Config())
	dlg.Settings().AddCategory("通用", []string{"语言"})
	dlg.Settings().AddCategory("账户", []string{"密码"})
	ui.Tree().AppendChild(ui.Tree().Root(), dlg.ElementID())
	place(ui.Tree(), dlg.ElementID(), 0, 0, 480, 320)
	dlg.Open()
	buf := render.NewCommandBuffer()
	dlg.Draw(buf)
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: 24, GlobalY: 40})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 24, GlobalY: 40})
	if dlg.Settings().Current() != "账户" {
		t.Fatalf("pointer left category %q", dlg.Settings().Current())
	}
	ctrl := dlg.Settings().Control(0)
	if ctrl == nil {
		t.Fatal("category has no row control")
	}
	b := ui.Tree().Get(ctrl.ElementID()).Layout().Bounds
	if b.Width < 20 || b.X < 160 {
		t.Fatalf("row control was not laid out beside the sidebar: %+v", b)
	}
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: b.X + 8, GlobalY: b.Y + 8})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: b.X + 8, GlobalY: b.Y + 8})
	ui.HandleEvent(&event.Event{Type: event.KeyPress, Char: 'p'})
	in, ok := ctrl.(*widget.Input)
	if !ok || in.Value() != "p" {
		t.Fatalf("row control not usable: %T %v", ctrl, ctrl)
	}
	ui.Tree().SetFocused(dlg.ElementID(), true)
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyArrowUp})
	if dlg.Settings().Current() != "通用" {
		t.Fatalf("keyboard left category %q", dlg.Settings().Current())
	}
}

func TestFormRowLaysOutFields(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	section := widget.NewFormSection(ui.Tree(), "账号", ui.Config())
	row := widget.NewFormRow(ui.Tree(), ui.Config())
	left := widget.NewFormField(ui.Tree(), "名", ui.Config())
	right := widget.NewFormField(ui.Tree(), "姓", ui.Config())
	li := widget.NewInput(ui.Tree(), ui.Config())
	ri := widget.NewInput(ui.Tree(), ui.Config())
	left.SetControl(li)
	right.SetControl(ri)
	row.Add(left)
	row.Add(right)
	section.Add(row)
	ui.Tree().AppendChild(ui.Tree().Root(), section.ElementID())
	place(ui.Tree(), section.ElementID(), 0, 0, 320, 80)
	section.Draw(render.NewCommandBuffer())
	lb := ui.Tree().Get(left.ElementID()).Layout().Bounds
	rb := ui.Tree().Get(right.ElementID()).Layout().Bounds
	if lb.Width < 40 || rb.Width < 40 || lb.X+lb.Width > rb.X+1 {
		t.Fatalf("fields overlap or have no width: %+v %+v", lb, rb)
	}
	form := widget.NewForm(ui.Tree(), ui.Config())
	item := widget.NewFormItem(ui.Tree(), "名", li, ui.Config())
	form.AddRule(item, func() string {
		if li.Value() == "" {
			return "必填"
		}
		return ""
	})
	if form.Validate() {
		t.Fatal("form validation was bypassed")
	}
}

func TestCommandRootClickRunsFilteredRow(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(240, 160)
	cmd := widget.NewCommand(ui.Tree(), ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), cmd.ElementID())
	place(ui.Tree(), cmd.ElementID(), 10, 10, 180, 28)
	cmd.Add("打开文件", "open")
	cmd.Add("关闭窗口", "close")
	ran := ""
	cmd.OnRun(func(id string) { ran = id })
	buf := render.NewCommandBuffer()
	cmd.Draw(buf)
	var row *render.RectCmd
	for _, list := range [][]render.Command{buf.Commands(), buf.Overlays()} {
		for i := range list {
			c := list[i]
			if c.Rect == nil || c.Rect.Bounds.Y <= 40 || c.Rect.Bounds.Height < 20 {
				continue
			}
			if row == nil || c.Rect.Bounds.Y < row.Bounds.Y {
				row = c.Rect
			}
		}
	}
	if row == nil {
		t.Fatal("command did not paint a filtered row")
	}
	hit := ui.Tree().HitTest(row.Bounds.X+8, row.Bounds.Y+8)
	if hit == core.InvalidElementID || hit == cmd.ElementID() {
		t.Fatalf("row hit %d", hit)
	}
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: row.Bounds.X + 8, GlobalY: row.Bounds.Y + 8})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: row.Bounds.X + 8, GlobalY: row.Bounds.Y + 8})
	if ran != "open" {
		t.Fatalf("click ran %q", ran)
	}
	r, g, b, a := softwarePixel(buf, 240, 160, int(row.Bounds.X+4), int(row.Bounds.Y+4))
	if a == 0 && r == 0 && g == 0 && b == 0 {
		t.Fatal("software raster left the command row empty")
	}
	cmd.Destroy()
	if ui.Tree().HitTest(row.Bounds.X+8, row.Bounds.Y+8) == hit {
		t.Fatal("destroyed command row still hits")
	}
}

func TestHTMLNavKeepsChildrenAndFollowingMeter(t *testing.T) {
	tree := core.NewTree()
	doc := LoadHTMLDocument(tree, widget.DefaultConfig(), `<nav><p>Hi</p></nav><meter value="0.5"></meter><togglebutton label="粗体"></togglebutton><navstack></navstack>`, "")
	root := doc.Root
	if len(root.Children()) < 3 {
		t.Fatalf("siblings lost, children %d", len(root.Children()))
	}
	if _, ok := root.Children()[0].(*widget.Nav); ok {
		t.Fatal("legacy nav tag became the page stack")
	}
	if len(root.Children()[0].Children()) == 0 {
		t.Fatal("nav dropped the nested paragraph")
	}
	m, ok := root.Children()[1].(*widget.Meter)
	if !ok || m.Value() != 0.5 {
		t.Fatalf("following meter: %T", root.Children()[1])
	}
	if _, ok := root.Children()[2].(*widget.ToggleButton); !ok {
		t.Fatalf("following toggle: %T", root.Children()[2])
	}
	foundStack := false
	for _, ch := range root.Children() {
		if _, ok := ch.(*widget.Nav); ok {
			foundStack = true
		}
	}
	if !foundStack {
		t.Fatal("navstack tag did not create a page nav")
	}
	doc2 := LoadHTMLDocument(tree, widget.DefaultConfig(), `<formsection title="账号"><formrow><formfield label="名"></formfield></formrow></formsection><meter value="0.2"></meter>`, "")
	if len(doc2.Root.Children()) != 2 {
		t.Fatalf("section ate the following meter, children %d", len(doc2.Root.Children()))
	}
	if _, ok := doc2.Root.Children()[1].(*widget.Meter); !ok {
		t.Fatalf("meter after section: %T", doc2.Root.Children()[1])
	}
}

func TestPlainAndSidebarShellLayout(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(400, 240)
	shell := widget.NewPlainShell(ui.Tree(), ui.Config())
	body := widget.NewButton(ui.Tree(), "正文", ui.Config())
	shell.SetTitle("首页")
	shell.SetContent(body)
	ui.Tree().AppendChild(ui.Tree().Root(), shell.ElementID())
	place(ui.Tree(), shell.ElementID(), 0, 0, 300, 180)
	shell.Draw(render.NewCommandBuffer())
	bb := ui.Tree().Get(body.ElementID()).Layout().Bounds
	if bb.Y < 36 || bb.Width < 40 {
		t.Fatalf("plain content not below the header: %+v", bb)
	}
	if ui.Tree().HitTest(bb.X+4, bb.Y+4) != body.ElementID() {
		t.Fatal("plain content is not hittable")
	}
	sideShell := widget.NewSidebarShell(ui.Tree(), ui.Config())
	content := widget.NewButton(ui.Tree(), "内", ui.Config())
	sideShell.Sidebar().AddItem("菜单")
	sideShell.SetContent(content)
	ui.Tree().AppendChild(ui.Tree().Root(), sideShell.ElementID())
	place(ui.Tree(), sideShell.ElementID(), 0, 0, 320, 160)
	sideShell.Draw(render.NewCommandBuffer())
	sb := ui.Tree().Get(sideShell.Sidebar().ElementID()).Layout().Bounds
	cb := ui.Tree().Get(content.ElementID()).Layout().Bounds
	if sb.Width < 40 || cb.Width < 40 || sb.X+sb.Width > cb.X+1 {
		t.Fatalf("sidebar shell rects: side %+v content %+v", sb, cb)
	}
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: cb.X + 8, GlobalY: cb.Y + 8})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: cb.X + 8, GlobalY: cb.Y + 8})
	if ui.Tree().Focused() != content.ElementID() {
		t.Fatal("sidebar content click did not focus the button")
	}
}

func TestImageCompareClipsTwoImages(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	cmp := widget.NewImageCompare(ui.Tree(), ui.Config())
	before := widget.NewImage(ui.Tree(), render.TextureHandle(3), ui.Config())
	after := widget.NewImage(ui.Tree(), render.TextureHandle(4), ui.Config())
	cmp.SetBefore(before)
	cmp.SetAfter(after)
	cmp.SetPosition(25)
	ui.Tree().AppendChild(ui.Tree().Root(), cmp.ElementID())
	place(ui.Tree(), cmp.ElementID(), 0, 0, 200, 80)
	buf := render.NewCommandBuffer()
	cmp.Draw(buf)
	var clips []uimath.Rect
	images := 0
	for _, c := range buf.Commands() {
		if c.Clip != nil {
			clips = append(clips, c.Clip.Bounds)
		}
		if c.Image != nil {
			images++
		}
	}
	if images < 2 || len(clips) < 2 {
		t.Fatalf("images %d clips %d", images, len(clips))
	}
	if clips[0].Width < 40 || clips[0].Width > 60 {
		t.Fatalf("before clip %+v", clips[0])
	}
}

func TestOtpCompleteFiresOnce(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	otp := widget.NewOtpInput(ui.Tree(), 4, ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), otp.ElementID())
	place(ui.Tree(), otp.ElementID(), 10, 10, 160, 32)
	n := 0
	otp.OnComplete(func(string) { n++ })
	ui.Tree().SetFocused(otp.ElementID(), true)
	for _, ch := range []rune{'1', '2', '3', '4', 'z'} {
		ui.HandleEvent(&event.Event{Type: event.KeyPress, Char: ch})
	}
	if n != 1 || otp.Value() != "1234" {
		t.Fatalf("completions %d value %q", n, otp.Value())
	}
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyBackspace})
	ui.HandleEvent(&event.Event{Type: event.KeyPress, Char: '9'})
	if n != 2 || otp.Value() != "1239" {
		t.Fatalf("refill completions %d value %q", n, otp.Value())
	}
}

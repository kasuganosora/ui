package widget

import (
	"testing"
	"time"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/render"
)

func hostTree() (*core.Tree, core.HostInput) {
	tree := core.NewTree()
	bounds(tree, tree.Root(), 0, 0, 800, 600)
	return tree, core.HostInput{Tree: tree}
}

func hostClick(h core.HostInput, x, y float32) {
	h.Handle(&event.Event{Type: event.MouseDown, GlobalX: x, GlobalY: y, Button: event.MouseButtonLeft})
	h.Handle(&event.Event{Type: event.MouseUp, GlobalX: x, GlobalY: y, Button: event.MouseButtonLeft})
}

func hostKey(h core.HostInput, key event.Key) {
	h.Handle(&event.Event{Type: event.KeyDown, Key: key})
}

func TestMeterHostValueAndMarks(t *testing.T) {
	tree, _ := hostTree()
	m := NewMeter(tree, nil)
	tree.AppendChild(tree.Root(), m.ElementID())
	m.SetValue(1.4)
	m.SetMarks([]float32{0.25, 0.75})
	if m.Value() != 1 || len(m.Marks()) != 2 {
		t.Fatalf("value %v marks %v", m.Value(), m.Marks())
	}
	buf := render.NewCommandBuffer()
	bounds(tree, m.ElementID(), 10, 10, 100, 12)
	m.Draw(buf)
	if buf.Len() < 3 {
		t.Fatal("meter did not draw the bar and marks")
	}
	if _, ok := any(m).(*Progress); ok {
		t.Fatal("meter must not be a progress value")
	}
}

func TestHoldToConfirmHostHoldAndRelease(t *testing.T) {
	tree, h := hostTree()
	hold := NewHoldToConfirm(tree, "删除", nil)
	hold.SetDuration(time.Second)
	tree.AppendChild(tree.Root(), hold.ElementID())
	bounds(tree, hold.ElementID(), 10, 10, 80, 32)
	fired := 0
	hold.OnConfirm(func() { fired++ })
	h.Handle(&event.Event{Type: event.MouseDown, GlobalX: 20, GlobalY: 20})
	hold.Tick(0.4)
	h.Handle(&event.Event{Type: event.MouseUp, GlobalX: 20, GlobalY: 20})
	if fired != 0 || hold.Progress() != 0 {
		t.Fatalf("early release fired %d progress %v", fired, hold.Progress())
	}
	h.Handle(&event.Event{Type: event.MouseDown, GlobalX: 20, GlobalY: 20})
	hold.Tick(1.1)
	if fired != 1 {
		t.Fatalf("full hold fired %d", fired)
	}
	tree.SetEnabled(hold.ElementID(), false)
	fired = 0
	h.Handle(&event.Event{Type: event.MouseDown, GlobalX: 20, GlobalY: 20})
	hold.Tick(2)
	if fired != 0 {
		t.Fatal("disabled hold fired")
	}
}

func TestToggleButtonHostSpace(t *testing.T) {
	tree, h := hostTree()
	btn := NewToggleButton(tree, "粗体", nil)
	tree.AppendChild(tree.Root(), btn.ElementID())
	bounds(tree, btn.ElementID(), 10, 10, 60, 28)
	hostClick(h, 20, 20)
	if !btn.Selected() {
		t.Fatal("click did not select")
	}
	tree.SetFocused(btn.ElementID(), true)
	hostKey(h, event.KeySpace)
	if btn.Selected() {
		t.Fatal("space did not clear")
	}
	tree.SetEnabled(btn.ElementID(), false)
	hostKey(h, event.KeySpace)
	if btn.Selected() {
		t.Fatal("disabled toggle changed")
	}
}

func TestImageCompareHostDragAndKeys(t *testing.T) {
	tree, h := hostTree()
	c := NewImageCompare(tree, nil)
	tree.AppendChild(tree.Root(), c.ElementID())
	bounds(tree, c.ElementID(), 0, 0, 200, 40)
	h.Handle(&event.Event{Type: event.MouseDown, GlobalX: 50, GlobalY: 10})
	h.Handle(&event.Event{Type: event.MouseMove, GlobalX: 150, GlobalY: 10})
	h.Handle(&event.Event{Type: event.MouseUp, GlobalX: 150, GlobalY: 10})
	if c.Position() < 70 || c.Position() > 80 {
		t.Fatalf("drag position %v", c.Position())
	}
	tree.SetFocused(c.ElementID(), true)
	hostKey(h, event.KeyArrowRight)
	if c.Position() < 75 || c.Position() > 85 {
		t.Fatalf("arrow position %v", c.Position())
	}
}

func TestOtpInputHostTypeBackspacePaste(t *testing.T) {
	tree, h := hostTree()
	otp := NewOtpInput(tree, 4, nil)
	tree.AppendChild(tree.Root(), otp.ElementID())
	bounds(tree, otp.ElementID(), 10, 10, 160, 32)
	done := ""
	otp.OnComplete(func(v string) { done = v })
	tree.SetFocused(otp.ElementID(), true)
	for _, ch := range []rune{'1', '2'} {
		h.Handle(&event.Event{Type: event.KeyPress, Char: ch})
	}
	hostKey(h, event.KeyBackspace)
	if otp.Value() != "1" {
		t.Fatalf("after backspace %q", otp.Value())
	}
	h.Handle(&event.Event{Type: event.KeyDown, Text: "234"})
	if otp.Value() != "1234" || done != "1234" {
		t.Fatalf("paste %q done %q", otp.Value(), done)
	}
	otp.SetValue("")
	done = ""
	otp.SetDisabled(true)
	tree.SetFocused(otp.ElementID(), true)
	h.Handle(&event.Event{Type: event.KeyPress, Char: '9'})
	if otp.Value() != "" || done != "" {
		t.Fatal("disabled otp accepted input")
	}
}

func TestScrollbarHostDrag(t *testing.T) {
	tree, h := hostTree()
	s := NewScrollbar(tree, nil)
	s.SetMetrics(400, 100)
	tree.AppendChild(tree.Root(), s.ElementID())
	bounds(tree, s.ElementID(), 10, 10, 12, 100)
	h.Handle(&event.Event{Type: event.MouseDown, GlobalX: 12, GlobalY: 10})
	h.Handle(&event.Event{Type: event.MouseMove, GlobalX: 12, GlobalY: 60})
	h.Handle(&event.Event{Type: event.MouseUp, GlobalX: 12, GlobalY: 60})
	if s.Offset() < 140 || s.Offset() > 160 {
		t.Fatalf("offset %v", s.Offset())
	}
}

func TestToggleGroupHostSingleAndMulti(t *testing.T) {
	tree, h := hostTree()
	g := NewToggleGroup(tree, nil)
	g.Add("左", "l")
	g.Add("右", "r")
	tree.AppendChild(tree.Root(), g.ElementID())
	bounds(tree, g.ElementID(), 0, 0, 200, 32)
	tree.SetFocused(g.ElementID(), true)
	hostKey(h, event.KeyArrowRight)
	if g.Value() != "r" {
		t.Fatalf("single %q", g.Value())
	}
	m := NewToggleGroup(tree, nil)
	m.SetMultiple(true)
	m.Add("a", "a")
	m.Add("b", "b")
	tree.AppendChild(tree.Root(), m.ElementID())
	bounds(tree, m.ElementID(), 0, 40, 200, 32)
	tree.SetFocused(m.ElementID(), true)
	hostKey(h, event.KeySpace)
	hostKey(h, event.KeyArrowRight)
	hostKey(h, event.KeySpace)
	got := m.Values()
	if len(got) != 2 {
		t.Fatalf("multi %#v", got)
	}
}

func TestToolbarHostArrowsDoNotActivate(t *testing.T) {
	tree, h := hostTree()
	bar := NewToolbar(tree, nil)
	bar.Add("保存")
	bar.Add("打开")
	tree.AppendChild(tree.Root(), bar.ElementID())
	bounds(tree, bar.ElementID(), 0, 0, 200, 32)
	tree.SetFocused(bar.ElementID(), true)
	hostKey(h, event.KeyArrowRight)
	if bar.Index() != 1 || bar.Clicks(1) != 0 || bar.Clicks(0) != 0 {
		t.Fatalf("arrow index %d clicks %d %d", bar.Index(), bar.Clicks(0), bar.Clicks(1))
	}
	hostKey(h, event.KeyEnter)
	if bar.Clicks(1) != 1 {
		t.Fatal("enter did not activate")
	}
}

func TestHoverCardHostDelayFocusEscape(t *testing.T) {
	tree, h := hostTree()
	card := NewHoverCard(tree, nil)
	card.SetDelay(0.2)
	tree.AppendChild(tree.Root(), card.ElementID())
	bounds(tree, card.ElementID(), 10, 10, 40, 24)
	h.Handle(&event.Event{Type: event.MouseMove, GlobalX: 20, GlobalY: 20})
	card.Tick(0.05)
	if card.IsOpen() {
		t.Fatal("opened before the delay")
	}
	card.Tick(0.2)
	if !card.IsOpen() {
		t.Fatal("hover delay did not open")
	}
	tree.SetFocused(card.ElementID(), true)
	hostKey(h, event.KeyEscape)
	if card.IsOpen() {
		t.Fatal("escape did not close")
	}
	other := NewHoverCard(tree, nil)
	tree.AppendChild(tree.Root(), other.ElementID())
	bounds(tree, other.ElementID(), 200, 10, 40, 24)
	tree.SetFocused(other.ElementID(), true)
	if !other.IsOpen() {
		t.Fatal("focus did not open")
	}
	hostClick(h, 400, 400)
	if other.IsOpen() {
		t.Fatal("outside click did not close")
	}
}

func TestComboboxHostEditAndChoose(t *testing.T) {
	tree, h := hostTree()
	box := NewCombobox(tree, []string{"Apple", "Apricot", "Banana"}, nil)
	tree.AppendChild(tree.Root(), box.ElementID())
	bounds(tree, box.ElementID(), 10, 10, 120, 28)
	tree.SetFocused(box.ElementID(), true)
	h.Handle(&event.Event{Type: event.KeyPress, Char: 'a'})
	if !box.IsOpen() || box.Text() != "a" {
		t.Fatalf("text %q open %v", box.Text(), box.IsOpen())
	}
	hostKey(h, event.KeyEnter)
	if box.Value() != "Apple" || box.IsOpen() {
		t.Fatalf("value %q open %v", box.Value(), box.IsOpen())
	}
}

func TestSuggestListHostDoesNotOwnInput(t *testing.T) {
	tree, h := hostTree()
	in := NewInput(tree, nil)
	in.SetValue("keep")
	list := NewSuggestList(tree, nil)
	list.SetItems([]string{"alpha", "beta"})
	list.SetAnchor(10, 40, 100, 20)
	picked := ""
	list.OnSelect(func(s string) { picked = s })
	tree.AppendChild(tree.Root(), in.ElementID())
	tree.AppendChild(tree.Root(), list.ElementID())
	list.Open()
	tree.SetFocused(list.ElementID(), true)
	hostKey(h, event.KeyArrowDown)
	hostKey(h, event.KeyEnter)
	if picked != "beta" || in.Value() != "keep" {
		t.Fatalf("picked %q input %q", picked, in.Value())
	}
}

func TestCommandHostFilterAndRun(t *testing.T) {
	tree, h := hostTree()
	cmd := NewCommand(tree, nil)
	tree.AppendChild(tree.Root(), cmd.ElementID())
	bounds(tree, cmd.ElementID(), 10, 10, 200, 160)
	cmd.Add("打开文件", "open")
	cmd.Add("关闭窗口", "close")
	ran := ""
	cmd.OnRun(func(id string) { ran = id })
	tree.SetFocused(cmd.ElementID(), true)
	h.Handle(&event.Event{Type: event.KeyPress, Char: '关'})
	hostKey(h, event.KeyEnter)
	if ran != "close" {
		t.Fatalf("ran %q shown %d", ran, len(cmd.Shown()))
	}
}

func TestCommandDialogHostEscapeRestoresFocus(t *testing.T) {
	tree, h := hostTree()
	btn := NewButton(tree, "外", nil)
	tree.AppendChild(tree.Root(), btn.ElementID())
	bounds(tree, btn.ElementID(), 8, 8, 40, 24)
	tree.SetFocused(btn.ElementID(), true)
	dlg := NewCommandDialog(tree, nil)
	tree.AppendChild(tree.Root(), dlg.ElementID())
	bounds(tree, dlg.ElementID(), 0, 0, 800, 600)
	bounds(tree, dlg.Command().ElementID(), 40, 40, 240, 200)
	dlg.Command().Add("保存", "save")
	dlg.Open()
	if tree.FocusScope() != dlg.ElementID() {
		t.Fatal("dialog did not take the focus scope")
	}
	hostKey(h, event.KeyEscape)
	if dlg.IsOpen() || tree.Focused() != btn.ElementID() {
		t.Fatalf("open %v focus %d want %d", dlg.IsOpen(), tree.Focused(), btn.ElementID())
	}
}

func TestEmojiPickerHostSearchAndVirtualCells(t *testing.T) {
	tree, h := hostTree()
	p := NewEmojiPicker(tree, nil)
	tree.AppendChild(tree.Root(), p.ElementID())
	bounds(tree, p.ElementID(), 0, 0, 256, 64)
	buf := render.NewCommandBuffer()
	p.Draw(buf)
	if p.DrawnCount() == 0 || p.DrawnCount() >= len(EmojiCatalog()) {
		t.Fatalf("drawn %d catalog %d", p.DrawnCount(), len(EmojiCatalog()))
	}
	tree.SetFocused(p.ElementID(), true)
	for _, r := range "笑" {
		h.Handle(&event.Event{Type: event.KeyPress, Char: r})
	}
	found := p.Filtered()
	if len(found) == 0 || found[0].Char == "" {
		t.Fatalf("search %#v", found)
	}
	picked := ""
	p.OnSelect(func(e Emoji) { picked = e.Char })
	hostClick(h, 16, 16)
	if picked == "" {
		t.Fatal("click missed the visible cell")
	}
}

func TestNavHostPushPopTitle(t *testing.T) {
	tree, h := hostTree()
	nav := NewNav(tree, nil)
	tree.AppendChild(tree.Root(), nav.ElementID())
	bounds(tree, nav.ElementID(), 0, 0, 300, 200)
	a := NewDiv(tree, nil)
	b := NewDiv(tree, nil)
	nav.Push("第一页", a)
	nav.Push("第二页", b)
	if nav.Title() != "第二页" || nav.Header().Title() != "第二页" {
		t.Fatalf("title %q", nav.Title())
	}
	nav.Draw(render.NewCommandBuffer())
	hostClick(h, 16, 16)
	if nav.Title() != "第一页" || nav.Depth() != 1 {
		t.Fatalf("after back %q depth %d", nav.Title(), nav.Depth())
	}
}

func TestFormSectionHostRowsDoNotReplaceValidate(t *testing.T) {
	tree, _ := hostTree()
	section := NewFormSection(tree, "账号", nil)
	row := NewFormRow(tree, nil)
	field := NewFormField(tree, "名称", nil)
	in := NewInput(tree, nil)
	field.SetControl(in)
	row.Add(field)
	section.Add(row)
	form := NewForm(tree, nil)
	item := NewFormItem(tree, "名称", in, nil)
	form.AddRule(item, func() string {
		if in.Value() == "" {
			return "必填"
		}
		return ""
	})
	if form.Validate() {
		t.Fatal("empty input passed form validation")
	}
	if section.Title() != "账号" || field.Label() != "名称" || field.Control() != in {
		t.Fatal("form layout did not keep its slots")
	}
}

func TestPlainShellHostSlots(t *testing.T) {
	tree, _ := hostTree()
	shell := NewPlainShell(tree, nil)
	body := NewDiv(tree, nil)
	shell.SetTitle("首页")
	shell.SetContent(body)
	if shell.Title() != "首页" || shell.Content() != body {
		t.Fatal("plain shell slots")
	}
}

func TestSidebarHostGroupsAndArrows(t *testing.T) {
	tree, h := hostTree()
	side := NewSidebar(tree, nil)
	g := side.AddGroup("组")
	side.AddGroupItem(g, "子项")
	tree.AppendChild(tree.Root(), side.ElementID())
	bounds(tree, side.ElementID(), 0, 0, 160, 120)
	tree.SetFocused(side.ElementID(), true)
	hostKey(h, event.KeyEnter)
	if side.Expanded(g) {
		t.Fatal("group stayed open")
	}
	if len(side.visible()) != 1 {
		t.Fatalf("collapsed visible %d", len(side.visible()))
	}
	hostKey(h, event.KeyEnter)
	hostKey(h, event.KeyArrowDown)
	if side.Label() != "子项" {
		t.Fatalf("label %q", side.Label())
	}
}

func TestSidebarShellHostSlots(t *testing.T) {
	tree, _ := hostTree()
	shell := NewSidebarShell(tree, nil)
	body := NewDiv(tree, nil)
	shell.SetContent(body)
	shell.Sidebar().AddItem("菜单")
	if shell.Content() != body || shell.Sidebar().Label() != "菜单" && shell.Sidebar().Index() != 0 {
		t.Fatal("sidebar shell")
	}
}

func TestTabShellHostSwitch(t *testing.T) {
	tree, h := hostTree()
	tabs := NewTabShell(tree, nil)
	a := NewDiv(tree, nil)
	b := NewDiv(tree, nil)
	tabs.Add("概览", a)
	tabs.Add("详情", b)
	tree.AppendChild(tree.Root(), tabs.ElementID())
	bounds(tree, tabs.ElementID(), 0, 0, 200, 80)
	hostClick(h, 150, 10)
	if tabs.Title() != "详情" || tree.Get(a.ElementID()).IsVisible() || !tree.Get(b.ElementID()).IsVisible() {
		t.Fatalf("title %q", tabs.Title())
	}
}

func TestSettingsHostCategoryAndDialog(t *testing.T) {
	tree, h := hostTree()
	btn := NewButton(tree, "外", nil)
	tree.AppendChild(tree.Root(), btn.ElementID())
	bounds(tree, btn.ElementID(), 8, 8, 40, 24)
	tree.SetFocused(btn.ElementID(), true)
	dlg := NewSettingsDialog(tree, nil)
	dlg.Settings().AddCategory("通用", []string{"语言"})
	dlg.Settings().AddCategory("账户", []string{"密码", "邮箱"})
	tree.AppendChild(tree.Root(), dlg.ElementID())
	bounds(tree, dlg.ElementID(), 0, 0, 400, 300)
	bounds(tree, dlg.Settings().side.ElementID(), 0, 0, 160, 200)
	dlg.Open()
	hostClick(h, 20, 40)
	if dlg.Settings().Current() != "账户" || len(dlg.Settings().Rows()) != 2 {
		t.Fatalf("current %q rows %v", dlg.Settings().Current(), dlg.Settings().Rows())
	}
	hostKey(h, event.KeyEscape)
	if dlg.IsOpen() || tree.Focused() != btn.ElementID() {
		t.Fatalf("open %v focus %d", dlg.IsOpen(), tree.Focused())
	}
}

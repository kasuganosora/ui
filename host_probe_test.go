package ui

import (
	"testing"
	"time"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

func place(tree *core.Tree, id core.ElementID, x, y, w, h float32) {
	tree.SetLayout(id, core.LayoutResult{Bounds: uimath.NewRect(x, y, w, h)})
}

func TestEmbeddedTabMovesBetweenButtons(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(400, 300)
	a := widget.NewButton(ui.Tree(), "A", ui.Config())
	b := widget.NewButton(ui.Tree(), "B", ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), a.ElementID())
	ui.Tree().AppendChild(ui.Tree().Root(), b.ElementID())
	place(ui.Tree(), a.ElementID(), 10, 10, 80, 32)
	place(ui.Tree(), b.ElementID(), 100, 10, 80, 32)
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: 20, GlobalY: 20})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 20, GlobalY: 20})
	if ui.Tree().Focused() != a.ElementID() {
		t.Fatalf("click focused %d want %d", ui.Tree().Focused(), a.ElementID())
	}
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyTab})
	if ui.Tree().Focused() != b.ElementID() {
		t.Fatalf("tab focused %d want %d", ui.Tree().Focused(), b.ElementID())
	}
	n := 0
	b.OnClick(func() { n++ })
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeySpace})
	if n != 1 {
		t.Fatalf("space clicks %d", n)
	}
}

func TestEmbeddedHiddenDisabledLoadingAndTagInput(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(400, 200)
	a := widget.NewButton(ui.Tree(), "A", ui.Config())
	hidden := widget.NewButton(ui.Tree(), "H", ui.Config())
	load := widget.NewButton(ui.Tree(), "L", ui.Config())
	link := widget.NewLink(ui.Tree(), "docs", "https://example.test", ui.Config())
	tags := widget.NewTagInput(ui.Tree(), ui.Config())
	for _, id := range []core.ElementID{a.ElementID(), hidden.ElementID(), load.ElementID(), link.ElementID(), tags.ElementID()} {
		ui.Tree().AppendChild(ui.Tree().Root(), id)
	}
	place(ui.Tree(), a.ElementID(), 0, 0, 40, 30)
	place(ui.Tree(), hidden.ElementID(), 50, 0, 40, 30)
	place(ui.Tree(), load.ElementID(), 100, 0, 40, 30)
	place(ui.Tree(), link.ElementID(), 150, 0, 40, 30)
	place(ui.Tree(), tags.ElementID(), 0, 40, 120, 30)
	ui.Tree().SetVisible(hidden.ElementID(), false)
	load.SetLoading(true)
	link.SetDisabled(true)
	clicked := 0
	hidden.OnClick(func() { clicked++ })
	ui.Tree().SetFocused(hidden.ElementID(), true)
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeySpace})
	if clicked != 0 {
		t.Fatal("hidden button activated")
	}
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: 10, GlobalY: 10})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 10, GlobalY: 10})
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyTab})
	if ui.Tree().Focused() == hidden.ElementID() || ui.Tree().Focused() == load.ElementID() || ui.Tree().Focused() == link.ElementID() {
		t.Fatalf("tab landed on unavailable %d", ui.Tree().Focused())
	}
	tags.SetInputText("red")
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: 10, GlobalY: 50})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 10, GlobalY: 50})
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyEnter})
	if len(tags.Tags()) != 1 {
		t.Fatalf("tags %v focus %d", tags.Tags(), ui.Tree().Focused())
	}
}

func TestHeadlessPopupHitsOutsideWidget(t *testing.T) {
	h := NewHeadlessUI(nil)
	h.Resize(800, 600)
	tp := widget.NewTimePicker(h.Tree(), h.Config())
	dr := widget.NewDateRangePicker(h.Tree(), h.Config())
	cp := widget.NewColorPicker(h.Tree(), h.Config())
	cp.SetAlphaEnabled(true)
	cs := widget.NewCascader(h.Tree(), h.Config())
	cs.SetOptions([]*widget.CascaderOption{{Label: "A", Value: "a", Children: []*widget.CascaderOption{{Label: "B", Value: "b"}}}})
	ts := widget.NewTreeSelect(h.Tree(), h.Config())
	ts.SetRoots([]*widget.TreeNode{{Key: "root", Label: "Root"}})
	ac := widget.NewAutoComplete(h.Tree(), h.Config())
	ac.SetSuggestions([]string{"apple"})
	for _, id := range []core.ElementID{tp.ElementID(), dr.ElementID(), cp.ElementID(), cs.ElementID(), ts.ElementID(), ac.ElementID()} {
		h.Tree().AppendChild(h.Tree().Root(), id)
	}
	place(h.Tree(), tp.ElementID(), 0, 0, 120, 28)
	place(h.Tree(), dr.ElementID(), 140, 0, 180, 28)
	place(h.Tree(), cp.ElementID(), 340, 0, 36, 36)
	place(h.Tree(), cs.ElementID(), 400, 0, 120, 28)
	place(h.Tree(), ts.ElementID(), 540, 0, 120, 28)
	place(h.Tree(), ac.ElementID(), 0, 200, 140, 28)
	h.HandleEvent(&event.Event{Type: event.MouseClick, GlobalX: 10, GlobalY: 10})
	h.HandleEvent(&event.Event{Type: event.MouseClick, GlobalX: 2*28 + 14, GlobalY: 28 + 4 + 28 + 14})
	if tp.Hour() != 13 {
		t.Fatalf("hour %d", tp.Hour())
	}
	h.HandleEvent(&event.Event{Type: event.MouseClick, GlobalX: 150, GlobalY: 10})
	first := int(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Weekday())
	day := 15
	col := (first + day - 1) % 7
	row := (first + day - 1) / 7
	h.HandleEvent(&event.Event{Type: event.MouseClick, GlobalX: 140 + float32(col)*32 + 16, GlobalY: 28 + 4 + 36 + float32(row)*32 + 16})
	if _, _, sd := dr.StartDate(); sd != 15 {
		t.Fatalf("date range day %d", sd)
	}
	h.HandleEvent(&event.Event{Type: event.MouseClick, GlobalX: 350, GlobalY: 10})
	h.HandleEvent(&event.Event{Type: event.MouseClick, GlobalX: 340 + 90, GlobalY: cp.AlphaBarY() + 4})
	if cp.Alpha() < 0.2 || cp.Alpha() > 0.8 {
		t.Fatalf("alpha %v", cp.Alpha())
	}
	h.HandleEvent(&event.Event{Type: event.MouseClick, GlobalX: 410, GlobalY: 10})
	h.HandleEvent(&event.Event{Type: event.MouseClick, GlobalX: 420, GlobalY: 28 + 4 + 10})
	h.HandleEvent(&event.Event{Type: event.MouseClick, GlobalX: 400 + 150 + 20, GlobalY: 28 + 4 + 10})
	if path := cs.Selected(); len(path) != 2 {
		t.Fatalf("cascader %v", path)
	}
	h.HandleEvent(&event.Event{Type: event.MouseClick, GlobalX: 550, GlobalY: 10})
	h.HandleEvent(&event.Event{Type: event.MouseClick, GlobalX: 550, GlobalY: 28 + 4 + 10})
	if ts.Selected() != "root" {
		t.Fatalf("tree select %q", ts.Selected())
	}
	ac.SetText("a")
	h.HandleEvent(&event.Event{Type: event.MouseClick, GlobalX: 20, GlobalY: 200 + 28 + 4 + 10})
	if ac.Text() != "apple" {
		t.Fatalf("autocomplete %q", ac.Text())
	}
}

func TestEmbeddedDialogLifecycleAndSelectEscape(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(600, 400)
	outside := widget.NewButton(ui.Tree(), "out", ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), outside.ElementID())
	place(ui.Tree(), outside.ElementID(), 10, 10, 60, 28)
	a := widget.NewDialog(ui.Tree(), "A", ui.Config())
	b := widget.NewDialog(ui.Tree(), "B", ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), a.ElementID())
	ui.Tree().AppendChild(ui.Tree().Root(), b.ElementID())
	place(ui.Tree(), a.ElementID(), 0, 0, 600, 400)
	place(ui.Tree(), b.ElementID(), 0, 0, 600, 400)
	confirmed := 0
	a.OnConfirm(func() { confirmed++ })
	a.Open()
	b.Open()
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyEnter})
	if confirmed != 0 {
		t.Fatal("enter confirmed the dialog underneath")
	}
	b.Close()
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyEnter})
	if confirmed != 1 {
		t.Fatalf("enter confirms %d", confirmed)
	}
	a.Close()
	if ui.Tree().IsFocusable(a.ElementID()) {
		t.Fatal("closed dialog focusable")
	}
	ui.Tree().SetFocused(outside.ElementID(), true)
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyTab})
	if ui.Tree().Focused() != outside.ElementID() {
		t.Fatal("tab entered a closed dialog")
	}
	ui.HandleEvent(&event.Event{Type: event.MouseClick, GlobalX: 300, GlobalY: 200})
	if ui.Tree().Focused() != outside.ElementID() && ui.Tree().Get(ui.Tree().Focused()) != nil && ui.Tree().Get(ui.Tree().Focused()).ParentID() == a.ElementID() {
		t.Fatal("click hit a closed dialog")
	}

	d := widget.NewDialog(ui.Tree(), "S", ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), d.ElementID())
	place(ui.Tree(), d.ElementID(), 0, 0, 600, 400)
	sel := widget.NewSelect(ui.Tree(), []widget.SelectOption{{Label: "One", Value: "1"}, {Label: "Two", Value: "2"}}, ui.Config())
	ui.Tree().AppendChild(d.ElementID(), sel.ElementID())
	place(ui.Tree(), sel.ElementID(), 40, 40, 160, 32)
	d.Open()
	ui.HandleEvent(&event.Event{Type: event.MouseClick, GlobalX: 50, GlobalY: 50})
	if !sel.IsOpen() {
		t.Fatal("select did not open")
	}
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyEscape})
	if !d.IsVisible() || sel.IsOpen() {
		t.Fatalf("dialog %v select %v", d.IsVisible(), sel.IsOpen())
	}
	d.Close()
	n := 0
	outside.OnClick(func() { n++ })
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: 20, GlobalY: 20})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 20, GlobalY: 20})
	if n != 1 {
		t.Fatalf("backdrop blocked outside click %d", n)
	}
}

func TestEmbeddedDrawerCaptureAndTooltipFocus(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(400, 300)
	btn := widget.NewButton(ui.Tree(), "tip", ui.Config())
	tip := widget.NewTooltip(ui.Tree(), "hello", btn.ElementID(), ui.Config())
	dr := widget.NewDrawer(ui.Tree(), "d", ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), btn.ElementID())
	ui.Tree().AppendChild(ui.Tree().Root(), dr.ElementID())
	place(ui.Tree(), btn.ElementID(), 10, 10, 40, 24)
	place(ui.Tree(), dr.ElementID(), 0, 40, 200, 100)
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: 20, GlobalY: 16})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 20, GlobalY: 16})
	if !tip.IsVisible() {
		t.Fatal("tooltip did not follow real focus")
	}
	dr.Open()
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: 20, GlobalY: 50})
	ui.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: 20, GlobalY: 250})
	if dr.IsVisible() {
		t.Fatal("drag outside the drawer did not close it")
	}
}

func TestOutOfOrderDialogCloseKeepsUpperScope(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(400, 300)
	out := widget.NewButton(ui.Tree(), "out", ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), out.ElementID())
	place(ui.Tree(), out.ElementID(), 8, 8, 40, 24)
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: 12, GlobalY: 12})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 12, GlobalY: 12})
	a := widget.NewDialog(ui.Tree(), "A", ui.Config())
	b := widget.NewDialog(ui.Tree(), "B", ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), a.ElementID())
	ui.Tree().AppendChild(ui.Tree().Root(), b.ElementID())
	place(ui.Tree(), a.ElementID(), 0, 0, 400, 300)
	place(ui.Tree(), b.ElementID(), 0, 0, 400, 300)
	a.Open()
	b.Open()
	a.Close()
	if ui.Tree().FocusScope() != b.ElementID() {
		t.Fatalf("scope %d want %d", ui.Tree().FocusScope(), b.ElementID())
	}
	if !ui.Tree().InScope(ui.Tree().Focused(), b.ElementID()) {
		t.Fatalf("focus %d left dialog B", ui.Tree().Focused())
	}
	mb := widget.NewMessageBox(ui.Tree(), "m", "c", ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), mb.ElementID())
	place(ui.Tree(), mb.ElementID(), 0, 0, 400, 300)
	mb.Open()
	a.Open()
	a.Close()
	if ui.Tree().FocusScope() != mb.ElementID() || !ui.Tree().InScope(ui.Tree().Focused(), mb.ElementID()) {
		t.Fatalf("message box scope %d focus %d", ui.Tree().FocusScope(), ui.Tree().Focused())
	}
}

func TestSecondEscapeClosesDialogAfterSelect(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(500, 400)
	d := widget.NewDialog(ui.Tree(), "D", ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), d.ElementID())
	place(ui.Tree(), d.ElementID(), 0, 0, 500, 400)
	sel := widget.NewSelect(ui.Tree(), []widget.SelectOption{{Label: "One", Value: "1"}}, ui.Config())
	ui.Tree().AppendChild(d.ElementID(), sel.ElementID())
	place(ui.Tree(), sel.ElementID(), 30, 30, 140, 32)
	d.Open()
	ui.HandleEvent(&event.Event{Type: event.MouseClick, GlobalX: 40, GlobalY: 40})
	if !sel.IsOpen() {
		t.Fatal("select closed")
	}
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyEscape})
	if sel.IsOpen() || !d.IsVisible() {
		t.Fatalf("first escape select %v dialog %v", sel.IsOpen(), d.IsVisible())
	}
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyEscape})
	if d.IsVisible() {
		t.Fatal("second escape did not close the dialog")
	}
}

func TestSelectOptionClickAndEmptyBackdrop(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(400, 300)
	sel := widget.NewSelect(ui.Tree(), []widget.SelectOption{{Label: "One", Value: "1"}, {Label: "Two", Value: "2"}}, ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), sel.ElementID())
	place(ui.Tree(), sel.ElementID(), 10, 10, 160, 32)
	ui.HandleEvent(&event.Event{Type: event.MouseClick, GlobalX: 20, GlobalY: 20})
	sel.Draw(render.NewCommandBuffer())
	optY := float32(10 + 32 + 4 + 8)
	ui.HandleEvent(&event.Event{Type: event.MouseClick, GlobalX: 20, GlobalY: optY})
	if sel.Value() != "1" || sel.IsOpen() {
		t.Fatalf("direct click value %q open %v", sel.Value(), sel.IsOpen())
	}
	ui.HandleEvent(&event.Event{Type: event.MouseClick, GlobalX: 20, GlobalY: 20})
	sel.Draw(render.NewCommandBuffer())
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: 20, GlobalY: optY})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 20, GlobalY: optY})
	if sel.Value() != "1" || sel.IsOpen() {
		t.Fatalf("down/up value %q open %v", sel.Value(), sel.IsOpen())
	}
	empty := widget.NewSelect(ui.Tree(), []widget.SelectOption{{Label: "Apple", Value: "a"}}, ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), empty.ElementID())
	place(ui.Tree(), empty.ElementID(), 10, 180, 160, 32)
	empty.SetFilterable(true)
	empty.SetInputValue("zzz")
	ui.HandleEvent(&event.Event{Type: event.MouseClick, GlobalX: 20, GlobalY: 190})
	ui.HandleEvent(&event.Event{Type: event.MouseClick, GlobalX: 20, GlobalY: 230})
	if empty.IsOpen() || empty.Value() != "" {
		t.Fatalf("empty backdrop open %v value %q", empty.IsOpen(), empty.Value())
	}
}

func TestLaterModalBlocksEarlierOverlay(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(400, 300)
	dp := widget.NewDatePicker(ui.Tree(), ui.Config())
	dp.SetDate(2026, 1, 1)
	ui.Tree().AppendChild(ui.Tree().Root(), dp.ElementID())
	place(ui.Tree(), dp.ElementID(), 10, 10, 200, 32)
	ui.HandleEvent(&event.Event{Type: event.MouseClick, GlobalX: 20, GlobalY: 20})
	first := int(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Weekday())
	day := 15
	col := (first + day - 1) % 7
	row := (first + day - 1) / 7
	x := float32(10) + float32(col)*32 + 16
	y := float32(10+32+4+36) + float32(row)*32 + 16
	dlg := widget.NewDialog(ui.Tree(), "cover", ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), dlg.ElementID())
	place(ui.Tree(), dlg.ElementID(), 0, 0, 400, 300)
	dlg.Open()
	ui.HandleEvent(&event.Event{Type: event.MouseClick, GlobalX: x, GlobalY: y})
	if dp.Day() == 15 {
		t.Fatal("date overlay received a click through the later modal")
	}
}

func TestBlankClickClearsInputButNotItsOwnPress(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(300, 200)
	in := widget.NewInput(ui.Tree(), ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), in.ElementID())
	place(ui.Tree(), in.ElementID(), 10, 10, 120, 28)
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: 20, GlobalY: 20})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 20, GlobalY: 20})
	if ui.Tree().Focused() != in.ElementID() {
		t.Fatal("input click did not focus")
	}
	ui.HandleEvent(&event.Event{Type: event.IMECompositionUpdate, Text: "你"})
	if ui.Tree().Focused() != in.ElementID() {
		t.Fatal("IME cleared input focus")
	}
	ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: 250, GlobalY: 160})
	ui.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 250, GlobalY: 160})
	if ui.Tree().Focused() != core.InvalidElementID {
		t.Fatalf("blank click left focus %d", ui.Tree().Focused())
	}
}

func TestTouchCancelAndWindowBlurClearDrawerDrag(t *testing.T) {
	for _, kind := range []event.Type{event.TouchCancel, event.WindowBlur} {
		ui := NewEmbeddedUI(nil)
		ui.Resize(300, 400)
		dr := widget.NewDrawer(ui.Tree(), "d", ui.Config())
		ui.Tree().AppendChild(ui.Tree().Root(), dr.ElementID())
		place(ui.Tree(), dr.ElementID(), 0, 200, 300, 200)
		dr.Open()
		ui.HandleEvent(&event.Event{Type: event.MouseDown, GlobalX: 40, GlobalY: 240})
		ui.HandleEvent(&event.Event{Type: kind})
		ui.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: 40, GlobalY: 390})
		if !dr.IsVisible() {
			t.Fatalf("%v left the drawer dragging and the move closed it", kind)
		}
	}
}

func TestEmbeddedDisabledAncestorBlocksChild(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(200, 120)
	box := widget.NewDiv(ui.Tree(), ui.Config())
	btn := widget.NewButton(ui.Tree(), "go", ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), box.ElementID())
	ui.Tree().AppendChild(box.ElementID(), btn.ElementID())
	place(ui.Tree(), box.ElementID(), 0, 0, 200, 80)
	place(ui.Tree(), btn.ElementID(), 10, 10, 40, 24)
	ui.Tree().SetEnabled(box.ElementID(), false)
	clicks := 0
	btn.OnClick(func() { clicks++ })
	ui.HandleEvent(&event.Event{Type: event.MouseClick, GlobalX: 20, GlobalY: 20})
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyTab})
	ui.Tree().SetFocused(btn.ElementID(), true)
	if clicks != 0 || ui.Tree().Focused() == btn.ElementID() || ui.Tree().HitTest(20, 20) == btn.ElementID() {
		t.Fatalf("clicks %d focus %d hit %d", clicks, ui.Tree().Focused(), ui.Tree().HitTest(20, 20))
	}
}

func TestEmbeddedGamepadFallbackFocusesCurrent(t *testing.T) {
	ui := NewEmbeddedUI(nil)
	ui.Resize(200, 80)
	gn := widget.NewGamepadNavigator(ui.Tree(), ui.Config())
	a := widget.NewButton(ui.Tree(), "a", ui.Config())
	b := widget.NewButton(ui.Tree(), "b", ui.Config())
	gn.AddWidget(a)
	gn.AddWidget(b)
	ui.Tree().AppendChild(ui.Tree().Root(), gn.ElementID())
	ui.Tree().SetKeyFallback(gn.ElementID())
	ui.HandleEvent(&event.Event{Type: event.KeyDown, Key: event.KeyArrowDown})
	if ui.Tree().Focused() != a.ElementID() {
		t.Fatalf("fallback focused %d want %d", ui.Tree().Focused(), a.ElementID())
	}
}

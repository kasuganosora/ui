package widget

import (
	"testing"
	"time"

	"github.com/kasuganosora/ui/accessibility"
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

func key(tree *core.Tree, id core.ElementID, k event.Key) {
	core.NewDispatcher(tree).Dispatch(id, &event.Event{Type: event.KeyDown, Key: k})
}

func click(tree *core.Tree, id core.ElementID, x, y float32) {
	core.NewDispatcher(tree).Dispatch(id, &event.Event{Type: event.MouseClick, GlobalX: x, GlobalY: y})
}

func bounds(tree *core.Tree, id core.ElementID, x, y, w, h float32) {
	tree.SetLayout(id, core.LayoutResult{Bounds: uimath.NewRect(x, y, w, h)})
}

func TestButtonSpaceActivates(t *testing.T) {
	tree := core.NewTree()
	btn := NewButton(tree, "Go", nil)
	n := 0
	btn.OnClick(func() { n++ })
	tree.SetFocused(btn.ElementID(), true)
	key(tree, btn.ElementID(), event.KeySpace)
	if n != 1 {
		t.Fatalf("clicks %d", n)
	}
	btn.SetDisabled(true)
	key(tree, btn.ElementID(), event.KeyEnter)
	if n != 1 {
		t.Fatalf("disabled clicks %d", n)
	}
	btn.SetDisabled(false)
	btn.SetLoading(true)
	key(tree, btn.ElementID(), event.KeySpace)
	if n != 1 {
		t.Fatalf("loading clicks %d", n)
	}
	btn.SetIcon("check")
	if btn.Icon() != "check" {
		t.Fatal("icon")
	}
}

func TestRadioArrowsAndSingleTabStop(t *testing.T) {
	tree := core.NewTree()
	g := NewRadioGroup()
	a := NewRadio(tree, "a", nil)
	b := NewRadio(tree, "b", nil)
	g.Add(a)
	g.Add(b)
	stops := 0
	if tree.IsFocusable(a.ElementID()) {
		stops++
	}
	if tree.IsFocusable(b.ElementID()) {
		stops++
	}
	if stops != 1 {
		t.Fatalf("tab stops %d", stops)
	}
	tree.SetFocused(a.ElementID(), true)
	key(tree, a.ElementID(), event.KeyArrowDown)
	if g.Value() != "b" || !b.IsChecked() {
		t.Fatalf("value %q checked %v", g.Value(), b.IsChecked())
	}
	if tree.Focused() != b.ElementID() {
		t.Fatal("focus did not follow the arrow")
	}
}

func TestSliderArrowsAndRange(t *testing.T) {
	tree := core.NewTree()
	s := NewSlider(tree, nil)
	s.SetMin(0)
	s.SetMax(10)
	s.SetStep(2)
	s.SetValue(0)
	tree.SetFocused(s.ElementID(), true)
	key(tree, s.ElementID(), event.KeyArrowRight)
	if s.Value() != 2 {
		t.Fatalf("value %v", s.Value())
	}
	s.SetRangeValues(2, 8)
	if s.Value() != 2 {
		t.Fatalf("range value %v", s.Value())
	}
	key(tree, s.ElementID(), event.KeyArrowRight)
	lo, hi := s.RangeValues()
	if lo != 4 || hi != 8 {
		t.Fatalf("range %v %v", lo, hi)
	}
}

func TestDialogEscapeAndFocusRestore(t *testing.T) {
	tree := core.NewTree()
	btn := NewButton(tree, "open", nil)
	tree.SetFocused(btn.ElementID(), true)
	d := NewDialog(tree, "Title", nil)
	d.Open()
	if !d.IsVisible() {
		t.Fatal("not open")
	}
	if tree.Focused() == btn.ElementID() {
		t.Fatal("focus stayed on the opener")
	}
	openedFocus := tree.Focused()
	tree.RouteKey(&event.Event{Type: event.KeyDown, Key: event.KeyEscape})
	if d.IsVisible() {
		t.Fatal("escape did not close")
	}
	if tree.Focused() != btn.ElementID() {
		t.Fatalf("restored %d want %d (dialog had %d)", tree.Focused(), btn.ElementID(), openedFocus)
	}

	d.SetCloseOnEscKeydown(false)
	d.Open()
	tree.RouteKey(&event.Event{Type: event.KeyDown, Key: event.KeyEscape})
	if !d.IsVisible() {
		t.Fatal("escape closed a dialog that opted out")
	}
}

func TestDatePickerSelectsDayAndCleansHitTargets(t *testing.T) {
	tree := core.NewTree()
	dp := NewDatePicker(tree, nil)
	dp.SetDate(2026, 1, 1)
	id := dp.ElementID()
	bounds(tree, id, 10, 20, 260, 36)
	before := tree.ElementCount()
	click(tree, id, 20, 30)
	if !dp.IsOpen() {
		t.Fatal("did not open")
	}
	if tree.ElementCount() <= before {
		t.Fatal("open did not create hit targets")
	}
	first := weekdayIndex(2026, 1, 1)
	day := 15
	col := (first + day - 1) % 7
	row := (first + day - 1) / 7
	x := float32(10) + float32(col)*32 + 16
	y := float32(20+36+4+36) + float32(row)*32 + 16
	click(tree, id, x, y)
	if dp.Day() != 15 {
		t.Fatalf("day %d", dp.Day())
	}
	if dp.IsOpen() {
		t.Fatal("stayed open")
	}
	if tree.ElementCount() != before {
		t.Fatalf("elements %d want %d", tree.ElementCount(), before)
	}
}

func TestColorPickerPreset(t *testing.T) {
	tree := core.NewTree()
	cp := NewColorPicker(tree, nil)
	id := cp.ElementID()
	bounds(tree, id, 0, 0, 40, 40)
	var got uimath.Color
	cp.OnChange(func(c uimath.Color) { got = c })
	click(tree, id, 10, 10)
	click(tree, id, 8+12, 44+8+12)
	want := uimath.ColorHex("#f5222d")
	if got != want {
		t.Fatalf("got %+v want %+v", got, want)
	}
	if cp.IsOpen() {
		t.Fatal("stayed open")
	}
}

func TestCascaderPath(t *testing.T) {
	tree := core.NewTree()
	c := NewCascader(tree, nil)
	c.SetOptions([]*CascaderOption{{
		Label: "A", Value: "a",
		Children: []*CascaderOption{{Label: "B", Value: "b"}},
	}})
	id := c.ElementID()
	bounds(tree, id, 0, 0, 120, 32)
	click(tree, id, 10, 10)
	click(tree, id, 20, 32+4+10)
	click(tree, id, 150+20, 32+4+10)
	got := c.Selected()
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("path %v", got)
	}
}

func TestSelectFilterAndMultiple(t *testing.T) {
	tree := core.NewTree()
	s := NewSelect(tree, []SelectOption{
		{Label: "Apple", Value: "a"},
		{Label: "Banana", Value: "b"},
		{Label: "Berry", Value: "c"},
	}, nil)
	s.SetFilterable(true)
	s.SetInputValue("ber")
	id := s.ElementID()
	bounds(tree, id, 0, 0, 160, 32)
	click(tree, id, 10, 10)
	key(tree, id, event.KeyEnter)
	if s.Value() != "c" {
		t.Fatalf("filtered value %q", s.Value())
	}

	s.SetMultiple(true)
	s.SetFilterable(false)
	s.SetInputValue("")
	s.SetValue("")
	click(tree, id, 10, 10)
	key(tree, id, event.KeyEnter)
	if len(s.SelectedValues()) != 1 || s.SelectedValues()[0] != "a" {
		t.Fatalf("multi %v", s.SelectedValues())
	}
	if !s.IsOpen() {
		t.Fatal("multiple select closed")
	}
}

func TestTableSortPageAndWindow(t *testing.T) {
	tree := core.NewTree()
	tb := NewTable(tree, []TableColumn{{Title: "N", ColKey: "n", Width: 100, Sorter: true}}, nil)
	tb.SetRows([][]string{{"b"}, {"a"}})
	id := tb.ElementID()
	bounds(tree, id, 0, 0, 200, 400)
	var info SortInfo
	tb.OnSortChange(func(s SortInfo) { info = s })
	click(tree, id, 10, 10)
	if tb.Rows()[0][0] != "a" || info.SortBy != "n" || info.Descending {
		t.Fatalf("rows %v info %+v", tb.Rows(), info)
	}

	rows := make([][]string, 10000)
	for i := range rows {
		rows[i] = []string{"x"}
	}
	tb.SetRows(rows)
	bounds(tree, id, 0, 0, 200, 200)
	buf := render.NewCommandBuffer()
	tb.Draw(buf)
	if len(buf.Commands()) >= 200 {
		t.Fatalf("commands %d", len(buf.Commands()))
	}
	tb.SetScrollY(5000 * 48)
	var hit int = -1
	tb.OnRowClick(func(ctx RowEventContext) { hit = ctx.RowIndex })
	click(tree, id, 10, 48+1)
	if hit != 5000 {
		t.Fatalf("hit %d", hit)
	}

	tb.SetPageSize(10)
	var page PageInfo
	tb.OnPageChange(func(p PageInfo) { page = p })
	tb.NextPage()
	if page.Current != 1 || page.PageSize != 10 || page.Previous != 0 {
		t.Fatalf("page %+v", page)
	}
}

func TestFormValidateReset(t *testing.T) {
	tree := core.NewTree()
	in := NewInput(tree, nil)
	item := NewFormItem(tree, "Name", in, nil)
	form := NewForm(tree, nil)
	form.AppendChild(item)
	form.AddRule(item, func() string {
		if in.Value() == "" {
			return "required"
		}
		return ""
	})
	if form.Validate() {
		t.Fatal("empty validated")
	}
	if item.Error() != "required" {
		t.Fatalf("error %q", item.Error())
	}
	in.SetValue("ada")
	if !form.Validate() || item.Error() != "" {
		t.Fatal("valid state")
	}
	form.Reset()
	if in.Value() != "" || item.Error() != "" {
		t.Fatalf("reset value %q error %q", in.Value(), item.Error())
	}
}

func TestClockAndScrollBreakpoints(t *testing.T) {
	tree := core.NewTree()
	sw := NewSwiper(tree, nil)
	sw.AddPanel(NewDiv(tree, nil))
	sw.AddPanel(NewDiv(tree, nil))
	sw.SetAutoplay(true)
	sw.Tick(3)
	if sw.Current() != 1 {
		t.Fatalf("slide %d", sw.Current())
	}

	n := NewNotification(tree, "t", "c", nil)
	n.SetDuration(1000)
	ended := false
	n.OnDurationEnd(func() { ended = true })
	n.Show()
	n.Tick(1)
	if n.IsVisible() || !ended {
		t.Fatalf("visible %v ended %v", n.IsVisible(), ended)
	}

	vl := NewVirtualList(tree, nil)
	vl.SetItemCount(100)
	vl.SetItemHeight(20)
	bounds(tree, vl.ElementID(), 0, 0, 100, 40)
	core.NewDispatcher(tree).Dispatch(vl.ElementID(), &event.Event{Type: event.MouseWheel, WheelDY: 50})
	if vl.ScrollY() != 50 {
		t.Fatalf("scroll %v", vl.ScrollY())
	}

	row := NewRow(tree, nil)
	col := NewCol(tree, 6, nil)
	col.SetSpanAt("lg", 12)
	row.AppendChild(col)
	if col.Span() != 6 {
		t.Fatalf("base span %d", col.Span())
	}
	row.SetViewport(500)
	if col.Span() != 6 {
		t.Fatalf("narrow span %d", col.Span())
	}
	row.SetViewport(1000)
	if col.Span() != 12 {
		t.Fatalf("lg span %d", col.Span())
	}
}

func TestAccessibilityMetadata(t *testing.T) {
	tree := core.NewTree()
	btn := NewButton(tree, "Go", nil)
	sw := NewSwitch(tree, nil)
	sw.SetDisabled(true)
	sl := NewSlider(tree, nil)
	sl.SetMin(0)
	sl.SetMax(10)
	sl.SetValue(4)
	props := accessibility.Collect([]accessibility.Accessible{btn, sw, sl})
	if props[0].Role != accessibility.RoleButton || props[0].Name != "Go" {
		t.Fatalf("button %+v", props[0])
	}
	if !props[1].State.Has(accessibility.StateDisabled) || props[1].Role != accessibility.RoleSwitch {
		t.Fatalf("switch %+v", props[1])
	}
	if props[2].ValueNow != 4 || props[2].ValueMax != 10 {
		t.Fatalf("slider %+v", props[2])
	}
}

func TestCheckboxLinkTabsMenuPaginationRate(t *testing.T) {
	tree := core.NewTree()
	cb := NewCheckbox(tree, "ok", nil)
	tree.SetFocused(cb.ElementID(), true)
	key(tree, cb.ElementID(), event.KeySpace)
	if !cb.IsChecked() {
		t.Fatal("checkbox")
	}
	link := NewLink(tree, "docs", "https://example.test", nil)
	var href string
	link.OnClick(func(h string) { href = h })
	key(tree, link.ElementID(), event.KeyEnter)
	if href != "https://example.test" {
		t.Fatalf("href %q", href)
	}
	tabs := NewTabs(tree, []TabPanel{{Value: "a", Label: "A"}, {Value: "b", Label: "B"}}, nil)
	tree.SetFocused(tabs.ElementID(), true)
	key(tree, tabs.ElementID(), event.KeyArrowRight)
	if tabs.Value() != "b" {
		t.Fatalf("tab %q", tabs.Value())
	}
	menu := NewMenu(tree, nil)
	menu.SetItems([]MenuItem{{Value: "one", Content: "One"}, {Value: "two", Content: "Two"}})
	tree.SetFocused(menu.ElementID(), true)
	key(tree, menu.ElementID(), event.KeyArrowDown)
	key(tree, menu.ElementID(), event.KeyArrowDown)
	key(tree, menu.ElementID(), event.KeyEnter)
	if menu.Value() != "two" {
		t.Fatalf("menu %q", menu.Value())
	}
	pg := NewPagination(tree, nil)
	pg.SetTotal(50)
	pg.SetPageSize(10)
	key(tree, pg.ElementID(), event.KeyArrowRight)
	if pg.Current() != 2 {
		t.Fatalf("page %d", pg.Current())
	}
	rate := NewRate(tree, nil)
	tree.SetFocused(rate.ElementID(), true)
	key(tree, rate.ElementID(), event.KeyArrowRight)
	if rate.Value() != 1 {
		t.Fatalf("rate %v", rate.Value())
	}
	num := NewInputNumber(tree, nil)
	tree.SetFocused(num.ElementID(), true)
	key(tree, num.ElementID(), event.KeyArrowUp)
	if num.Value() != 1 {
		t.Fatalf("number %v", num.Value())
	}
}

func TestTreeTooltipPopoverDrawer(t *testing.T) {
	tree := core.NewTree()
	tr := NewTree(tree, nil)
	tr.SetRoots([]*TreeNode{{Key: "a", Label: "A"}, {Key: "b", Label: "B", Children: []*TreeNode{{Key: "b1", Label: "B1"}}}})
	bounds(tree, tr.ElementID(), 0, 0, 200, 200)
	tree.SetFocused(tr.ElementID(), true)
	key(tree, tr.ElementID(), event.KeyArrowDown)
	key(tree, tr.ElementID(), event.KeyArrowDown)
	key(tree, tr.ElementID(), event.KeyEnter)
	if tr.SelectedKey() != "b" {
		t.Fatalf("selected %q", tr.SelectedKey())
	}
	key(tree, tr.ElementID(), event.KeyArrowRight)
	if !tr.Roots()[1].Expanded {
		t.Fatal("not expanded")
	}
	key(tree, tr.ElementID(), event.KeySpace)
	if !tr.Roots()[1].Checked {
		t.Fatal("not checked")
	}

	anchor := NewButton(tree, "tip", nil)
	tip := NewTooltip(tree, "hello", anchor.ElementID(), nil)
	tree.SetFocused(anchor.ElementID(), true)
	core.NewDispatcher(tree).Dispatch(anchor.ElementID(), &event.Event{Type: event.Focus})
	if !tip.IsVisible() {
		t.Fatal("tooltip hidden on focus")
	}
	key(tree, anchor.ElementID(), event.KeyEscape)
	if tip.IsVisible() {
		t.Fatal("tooltip stayed")
	}

	pop := NewPopover(tree, nil)
	pop.Open()
	tree.SetFocused(pop.ElementID(), true)
	key(tree, pop.ElementID(), event.KeyEscape)
	if pop.IsVisible() {
		t.Fatal("popover stayed")
	}

	dr := NewDrawer(tree, "side", nil)
	dr.Open()
	id := dr.ElementID()
	bounds(tree, id, 0, 0, 200, 100)
	core.NewDispatcher(tree).Dispatch(id, &event.Event{Type: event.MouseDown, GlobalX: 20, GlobalY: 10})
	core.NewDispatcher(tree).Dispatch(id, &event.Event{Type: event.MouseMove, GlobalX: 20, GlobalY: 80})
	if dr.IsVisible() {
		t.Fatal("drawer stayed open")
	}
}

func TestUploadAffixImageViewerDockDrag(t *testing.T) {
	tree := core.NewTree()
	up := NewUpload(tree, nil)
	up.SetMaxCount(1)
	up.SetOpenPicker(func() ([]UploadFile, error) {
		return []UploadFile{{Name: "a.txt", Size: 3, Status: "done"}}, nil
	})
	click(tree, up.ElementID(), 1, 1)
	if len(up.Files()) != 1 || up.Files()[0].Name != "a.txt" {
		t.Fatalf("files %+v", up.Files())
	}

	aff := NewAffix(tree, nil)
	box := NewDiv(tree, nil)
	bounds(tree, box.ElementID(), 0, 80, 10, 10)
	aff.SetContent(box)
	aff.SetOffsetTop(12)
	aff.SetScrollY(40)
	if !aff.IsAffixed() || box.Bounds().Y != 12 {
		t.Fatalf("affixed %v y %v", aff.IsAffixed(), box.Bounds().Y)
	}

	iv := NewImageViewer(tree, nil)
	bounds(tree, iv.ElementID(), 0, 0, 100, 100)
	core.NewDispatcher(tree).Dispatch(iv.ElementID(), &event.Event{Type: event.MouseDown, GlobalX: 10, GlobalY: 10})
	core.NewDispatcher(tree).Dispatch(iv.ElementID(), &event.Event{Type: event.MouseMove, GlobalX: 25, GlobalY: 18})
	px, py := iv.Pan()
	if px != 15 || py != 8 {
		t.Fatalf("pan %v %v", px, py)
	}

	dl := NewDockLayout(tree, nil)
	dl.AddPanel(&DockPanel{ID: "p", Title: "P"})
	var docked DockPosition
	dl.OnDock(func(id string, pos DockPosition) {
		if id == "p" {
			docked = pos
		}
	})
	dl.DockPanel("p", DockLeft)
	if docked != DockLeft || dl.FindPanel("p").Position != DockLeft {
		t.Fatal("dock")
	}

	src := NewButton(tree, "s", nil)
	dst := NewButton(tree, "d", nil)
	tree.AppendChild(tree.Root(), src.ElementID())
	tree.AppendChild(tree.Root(), dst.ElementID())
	bounds(tree, tree.Root(), 0, 0, 400, 200)
	bounds(tree, src.ElementID(), 0, 0, 20, 20)
	bounds(tree, dst.ElementID(), 80, 0, 40, 40)
	dd := NewDragDropManager(tree)
	var dropped string
	dd.RegisterSource(&DragSource{Widget: src, Data: "item"})
	dd.RegisterTarget(&DropTarget{Widget: dst, Accept: func(any) bool { return true }, OnDrop: func(data any) {
		dropped, _ = data.(string)
	}})
	core.NewDispatcher(tree).Dispatch(src.ElementID(), &event.Event{Type: event.MouseDown, GlobalX: 2, GlobalY: 2})
	dd.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: 90, GlobalY: 10})
	dd.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 90, GlobalY: 10})
	if dropped != "item" {
		t.Fatalf("dropped %q", dropped)
	}

	sp := NewSplitter(tree, nil)
	sp.SetMinRatio(0.2)
	sp.SetRatio(0)
	if sp.Ratio() != 0.2 {
		t.Fatalf("ratio %v", sp.Ratio())
	}

	ac := NewAutoComplete(tree, nil)
	ac.SetSuggestions([]string{"apple", "pear"})
	bounds(tree, ac.ElementID(), 0, 0, 120, 28)
	ac.SetText("a")
	key(tree, ac.ElementID(), event.KeyArrowDown)
	key(tree, ac.ElementID(), event.KeyEnter)
	if ac.Text() != "apple" {
		t.Fatalf("complete %q", ac.Text())
	}

	cm := NewContextMenu(tree, nil)
	hit := false
	cm.AddItem(ContextMenuItem{Label: "Copy", OnClick: func() { hit = true }})
	cm.Show(0, 0)
	tree.SetFocused(cm.ElementID(), true)
	key(tree, cm.ElementID(), event.KeyEnter)
	if !hit {
		t.Fatal("context menu")
	}

	mb := NewMenuBar(tree, nil)
	fired := false
	mb.AddItem(MenuBarItem{Label: "File", SubItems: []MenuBarSubItem{{Label: "New", OnClick: func() { fired = true }}}})
	tree.SetFocused(mb.ElementID(), true)
	key(tree, mb.ElementID(), event.KeyArrowRight)
	key(tree, mb.ElementID(), event.KeyEnter)
	if !fired {
		t.Fatal("menubar")
	}

	ti := NewTagInput(tree, nil)
	ti.SetInputText("red")
	tree.SetFocused(ti.ElementID(), true)
	key(tree, ti.ElementID(), event.KeyEnter)
	if len(ti.Tags()) != 1 || !ti.Element().IsFocused() {
		t.Fatal("tag input")
	}
	ti.SetInputText("")
	key(tree, ti.ElementID(), event.KeyBackspace)
	if len(ti.Tags()) != 0 {
		t.Fatal("backspace")
	}

	ts := NewTreeSelect(tree, nil)
	ts.SetRoots([]*TreeNode{{Key: "root", Label: "Root"}})
	bounds(tree, ts.ElementID(), 0, 0, 100, 28)
	click(tree, ts.ElementID(), 10, 10)
	click(tree, ts.ElementID(), 10, 28+4+10)
	if ts.Selected() != "root" {
		t.Fatalf("tree select %q", ts.Selected())
	}

	drp := NewDateRangePicker(tree, nil)
	bounds(tree, drp.ElementID(), 10, 20, 260, 36)
	click(tree, drp.ElementID(), 20, 30)
	first := weekdayIndex(2026, 1, 1)
	pick := func(day int) {
		col := (first + day - 1) % 7
		row := (first + day - 1) / 7
		click(tree, drp.ElementID(), float32(10)+float32(col)*32+16, float32(20+36+4+36)+float32(row)*32+16)
	}
	pick(10)
	pick(3)
	sy, sm, sd := drp.StartDate()
	ey, em, ed := drp.EndDate()
	if sy != 2026 || sm != 1 || sd != 3 || ey != 2026 || em != 1 || ed != 10 {
		t.Fatalf("range %d-%d-%d %d-%d-%d", sy, sm, sd, ey, em, ed)
	}

	msg := NewMessageBox(tree, "Hi", "Body", nil)
	msg.Open()
	tree.RouteKey(&event.Event{Type: event.KeyDown, Key: event.KeyEscape})
	if msg.IsVisible() {
		t.Fatal("message box stayed open")
	}

	pc := NewPopconfirm(tree, "Sure?", nil)
	pc.SetVisible(true)
	tree.SetFocused(pc.ElementID(), true)
	key(tree, pc.ElementID(), event.KeyEscape)
	if pc.IsVisible() {
		t.Fatal("popconfirm stayed open")
	}

	cal := NewCalendar(tree, nil)
	bounds(tree, cal.ElementID(), 0, 0, 300, 400)
	col := (first + 15 - 1) % 7
	row := (first + 15 - 1) / 7
	cs := float32(36)
	click(tree, cal.ElementID(), float32(col)*cs+cs/2, 40+cs+float32(row)*cs+cs/2)
	if cal.Selected() != 15 {
		t.Fatalf("calendar %d", cal.Selected())
	}
	_ = time.Thursday
}

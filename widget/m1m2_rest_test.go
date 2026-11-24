package widget

import (
	"testing"

	"github.com/kasuganosora/ui/accessibility"
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
)

func dispatchHit(tree *core.Tree, x, y float32) core.ElementID {
	id := tree.HitTest(x, y)
	if id == core.InvalidElementID {
		return id
	}
	core.NewDispatcher(tree).Dispatch(id, &event.Event{Type: event.MouseClick, GlobalX: x, GlobalY: y})
	return id
}

func TestInputAndTextAreaAccessibility(t *testing.T) {
	tree := core.NewTree()
	in := NewInput(tree, nil)
	in.SetPlaceholder("name")
	in.SetValue("ada")
	in.SetDisabled(true)
	ip := in.AccessibleProps()
	if ip.Role != accessibility.RoleTextBox || ip.Value != "ada" || ip.Name != "name" || !ip.State.Has(accessibility.StateDisabled) {
		t.Fatalf("input %+v", ip)
	}
	ta := NewTextArea(tree, nil)
	ta.SetPlaceholder("note")
	ta.SetValue("hi")
	tp := ta.AccessibleProps()
	if tp.Role != accessibility.RoleTextArea || tp.Value != "hi" || tp.Name != "note" {
		t.Fatalf("textarea %+v", tp)
	}
}

func TestRemainingAccessibilityTypes(t *testing.T) {
	tree := core.NewTree()
	av := NewAvatar(tree, nil)
	av.SetAlt("lee")
	ag := NewAvatarGroup(tree, nil)
	bd := NewBadge(tree, nil)
	bc := NewBreadcrumb(tree, nil)
	img := NewImage(tree, 0, nil)
	pg := NewProgress(tree, nil)
	pg.SetPercentage(40)
	sk := NewSkeleton(tree, nil)
	sp := NewSpace(tree, nil)
	ld := NewLoading(tree, nil)
	st := NewSteps(tree, nil)
	tx := NewText(tree, "hello", nil)
	cd := NewCard(tree, nil)
	cd.SetTitle("card")
	ic := NewIcon(tree, "home", nil)
	cv := NewCanvas(tree, 4, 4, nil, nil)
	po := NewPortal(tree, nil)
	props := accessibility.Collect([]accessibility.Accessible{av, ag, bd, bc, img, pg, sk, sp, ld, st, tx, cd, ic, cv, po})
	want := []accessibility.Role{
		accessibility.RoleImage, accessibility.RoleGroup, accessibility.RoleLabel, accessibility.RoleGroup,
		accessibility.RoleImage, accessibility.RoleProgressBar, accessibility.RoleGroup, accessibility.RoleGroup,
		accessibility.RoleProgressBar, accessibility.RoleGroup, accessibility.RoleLabel, accessibility.RoleGroup,
		accessibility.RoleImage, accessibility.RoleImage, accessibility.RoleGroup,
	}
	if len(props) != len(want) {
		t.Fatalf("len %d", len(props))
	}
	for i, role := range want {
		if props[i].Role != role {
			t.Fatalf("prop %d role %v want %v (%+v)", i, props[i].Role, role, props[i])
		}
	}
	if props[0].Name != "lee" || props[5].ValueNow != 40 || props[10].Name != "hello" || props[11].Name != "card" || props[12].Name != "home" {
		t.Fatalf("names %+v %+v %+v %+v %+v", props[0], props[5], props[10], props[11], props[12])
	}
}

func TestCollapseAccordionKeyboard(t *testing.T) {
	tree := core.NewTree()
	c := NewCollapse(tree, nil)
	c.SetExpandMutex(true)
	c.SetPanels([]CollapsePanel{{Header: "A", Value: "a"}, {Header: "B", Value: "b", Disabled: true}, {Header: "C", Value: "c"}})
	tree.SetFocused(c.ElementID(), true)
	key(tree, c.ElementID(), event.KeyEnter)
	if !c.IsActive("a") {
		t.Fatal("enter did not open a")
	}
	key(tree, c.ElementID(), event.KeyArrowDown)
	key(tree, c.ElementID(), event.KeyEnter)
	if c.IsActive("b") || !c.IsActive("c") || c.IsActive("a") {
		t.Fatalf("accordion a=%v b=%v c=%v", c.IsActive("a"), c.IsActive("b"), c.IsActive("c"))
	}
	c.SetDisabled(true)
	key(tree, c.ElementID(), event.KeyEnter)
	if !c.IsActive("c") {
		t.Fatal("disabled collapse toggled")
	}

	multi := NewCollapse(tree, nil)
	multi.SetPanels([]CollapsePanel{{Header: "A", Value: "a"}, {Header: "B", Value: "b"}})
	tree.SetFocused(multi.ElementID(), true)
	key(tree, multi.ElementID(), event.KeyEnter)
	key(tree, multi.ElementID(), event.KeyArrowDown)
	key(tree, multi.ElementID(), event.KeyEnter)
	if !multi.IsActive("a") || !multi.IsActive("b") {
		t.Fatal("independent panels did not both stay open")
	}
}

func TestTimePickerHit(t *testing.T) {
	tree := core.NewTree()
	tp := NewTimePicker(tree, nil)
	id := tp.ElementID()
	bounds(tree, id, 0, 0, 200, 32)
	click(tree, id, 10, 10)
	if !tp.IsOpen() {
		t.Fatal("not open")
	}
	// 4 columns x 6 rows, cell 28, origin y = 32+4. Hour 13 is col 2, row 1.
	click(tree, id, float32(2*28+14), float32(36+28+14))
	click(tree, id, float32(2*40+20), float32(36+6*28+8+14))
	if tp.Hour() != 13 || tp.Minute() != 30 {
		t.Fatalf("time %d:%d", tp.Hour(), tp.Minute())
	}
}

func TestScrollViewKeyboard(t *testing.T) {
	tree := core.NewTree()
	d := NewDiv(tree, nil)
	d.SetScrollable(true)
	d.SetContentHeight(500)
	bounds(tree, d.ElementID(), 0, 0, 100, 80)
	tree.SetFocused(d.ElementID(), true)
	key(tree, d.ElementID(), event.KeyPageDown)
	if d.ScrollY() != 80 {
		t.Fatalf("page %v", d.ScrollY())
	}
	key(tree, d.ElementID(), event.KeyPageDown)
	key(tree, d.ElementID(), event.KeyPageDown)
	key(tree, d.ElementID(), event.KeyPageDown)
	key(tree, d.ElementID(), event.KeyPageDown)
	key(tree, d.ElementID(), event.KeyPageDown)
	if d.ScrollY() != 420 {
		t.Fatalf("clamped %v", d.ScrollY())
	}
	tree.SetVisible(d.ElementID(), false)
	before := d.ScrollY()
	key(tree, d.ElementID(), event.KeyPageUp)
	if d.ScrollY() != before {
		t.Fatal("hidden scrolled")
	}
}

func TestColorAlphaAndRecent(t *testing.T) {
	tree := core.NewTree()
	cp := NewColorPicker(tree, nil)
	cp.SetAlphaEnabled(true)
	id := cp.ElementID()
	bounds(tree, id, 0, 0, 40, 40)
	click(tree, id, 10, 10)
	click(tree, id, 8+12, 44+8+12)
	if len(cp.Recent()) == 0 || cp.Recent()[0] != uimath.ColorHex("#f5222d") {
		t.Fatalf("recent %+v", cp.Recent())
	}
	click(tree, id, 10, 10)
	click(tree, id, 90, cp.AlphaBarY()+4)
	if cp.Alpha() < 0.4 || cp.Alpha() > 0.6 {
		t.Fatalf("alpha %v", cp.Alpha())
	}
}

func TestListKeyboard(t *testing.T) {
	tree := core.NewTree()
	l := NewList(tree, nil)
	l.SetSelectable(true)
	l.SetItems([]ListItem{{Value: "a", Title: "A"}, {Value: "b", Title: "B", Disabled: true}, {Value: "c", Title: "C"}})
	var got string
	l.OnChange(func(v string) { got = v })
	tree.SetFocused(l.ElementID(), true)
	key(tree, l.ElementID(), event.KeyArrowDown)
	key(tree, l.ElementID(), event.KeyArrowDown)
	key(tree, l.ElementID(), event.KeyEnter)
	if l.SelectedValue() != "c" || got != "c" {
		t.Fatalf("selected %q got %q", l.SelectedValue(), got)
	}
}

func TestSwiperDrag(t *testing.T) {
	tree := core.NewTree()
	s := NewSwiper(tree, nil)
	s.AddPanel(NewDiv(tree, nil))
	s.AddPanel(NewDiv(tree, nil))
	id := s.ElementID()
	bounds(tree, id, 0, 0, 200, 80)
	core.NewDispatcher(tree).Dispatch(id, &event.Event{Type: event.MouseDown, GlobalX: 100, GlobalY: 20})
	core.NewDispatcher(tree).Dispatch(id, &event.Event{Type: event.MouseMove, GlobalX: 40, GlobalY: 20})
	core.NewDispatcher(tree).Dispatch(id, &event.Event{Type: event.MouseUp, GlobalX: 40, GlobalY: 20})
	if s.Current() != 1 {
		t.Fatalf("current %d", s.Current())
	}
}

func TestMessageTimeout(t *testing.T) {
	tree := core.NewTree()
	m := NewMessage(tree, "saved", nil)
	m.SetDuration(1000)
	ended := false
	m.OnDurationEnd(func() { ended = true })
	m.SetVisible(true)
	m.Tick(1)
	if m.IsVisible() || !ended {
		t.Fatalf("visible %v ended %v", m.IsVisible(), ended)
	}
}

func TestVirtualGridWheel(t *testing.T) {
	tree := core.NewTree()
	g := NewVirtualGrid(tree, 4, 40, nil)
	bounds(tree, g.ElementID(), 0, 0, 200, 100)
	core.NewDispatcher(tree).Dispatch(g.ElementID(), &event.Event{Type: event.MouseWheel, WheelDY: 30})
	if g.ScrollY() != 30 {
		t.Fatalf("scroll %v", g.ScrollY())
	}
}

func TestTreeLazyAndDragSort(t *testing.T) {
	tree := core.NewTree()
	tr := NewTree(tree, nil)
	loaded := false
	root := &TreeNode{Key: "p", Label: "P"}
	root.Lazy = func(n *TreeNode) {
		loaded = true
		n.Children = []*TreeNode{{Key: "c", Label: "C"}}
	}
	tr.SetRoots([]*TreeNode{root, {Key: "q", Label: "Q"}})
	bounds(tree, tr.ElementID(), 0, 0, 160, 120)
	tree.SetFocused(tr.ElementID(), true)
	key(tree, tr.ElementID(), event.KeyArrowDown)
	key(tree, tr.ElementID(), event.KeyArrowRight)
	if !loaded || !tr.Roots()[0].Expanded || len(tr.Roots()[0].Children) != 1 {
		t.Fatal("lazy expand")
	}
	core.NewDispatcher(tree).Dispatch(tr.ElementID(), &event.Event{Type: event.MouseDown, GlobalX: 10, GlobalY: 10})
	core.NewDispatcher(tree).Dispatch(tr.ElementID(), &event.Event{Type: event.MouseUp, GlobalX: 10, GlobalY: 74})
	if tr.Roots()[0].Key != "q" || tr.Roots()[1].Key != "p" {
		t.Fatalf("order %s %s", tr.Roots()[0].Key, tr.Roots()[1].Key)
	}
}

func TestGamepadUsesTreeFocus(t *testing.T) {
	tree := core.NewTree()
	a := NewButton(tree, "a", nil)
	b := NewButton(tree, "b", nil)
	c := NewButton(tree, "c", nil)
	b.SetDisabled(true)
	tree.SetVisible(c.ElementID(), false)
	gn := NewGamepadNavigator(tree, nil)
	gn.AddWidget(a)
	gn.AddWidget(b)
	gn.AddWidget(c)
	tree.AppendChild(tree.Root(), gn.ElementID())
	core.NewDispatcher(tree).Dispatch(gn.ElementID(), &event.Event{Type: event.KeyDown, Key: event.KeyArrowDown})
	if gn.FocusedWidget() == nil || gn.FocusedWidget().ElementID() != a.ElementID() {
		t.Fatal("first focus")
	}
	if tree.Focused() != a.ElementID() {
		t.Fatal("tree focus")
	}
	core.NewDispatcher(tree).Dispatch(gn.ElementID(), &event.Event{Type: event.KeyDown, Key: event.KeyArrowDown})
	if tree.Focused() != a.ElementID() {
		t.Fatal("moved onto disabled or hidden button")
	}
}

func TestRootDispatchFocusScopeAndNestedHit(t *testing.T) {
	tree := core.NewTree()
	bounds(tree, tree.Root(), 0, 0, 800, 800)
	outside := NewButton(tree, "out", nil)
	tree.AppendChild(tree.Root(), outside.ElementID())
	dlg := NewDialog(tree, "d", nil)
	tree.AppendChild(tree.Root(), dlg.ElementID())
	bounds(tree, dlg.ElementID(), 0, 0, 800, 800)
	dp := NewDatePicker(tree, nil)
	dp.SetDate(2026, 1, 1)
	tree.AppendChild(dlg.ElementID(), dp.ElementID())
	bounds(tree, dp.ElementID(), 10, 20, 260, 36)
	tree.SetFocused(outside.ElementID(), true)
	dlg.Open()
	tree.RouteKey(&event.Event{Type: event.KeyDown, Key: event.KeyTab})
	if tree.Focused() == outside.ElementID() {
		t.Fatal("tab left the dialog")
	}
	if dispatchHit(tree, 20, 30) == core.InvalidElementID || !dp.IsOpen() {
		t.Fatal("trigger hit missed")
	}
	first := weekdayIndex(2026, 1, 1)
	day := 15
	col := (first + day - 1) % 7
	row := (first + day - 1) / 7
	x := float32(10) + float32(col)*32 + 16
	y := float32(20+36+4+36) + float32(row)*32 + 16
	hit := dispatchHit(tree, x, y)
	if hit == dp.ElementID() || hit == core.InvalidElementID {
		t.Fatalf("nested hit %d picker %d", hit, dp.ElementID())
	}
	if dp.Day() != 15 || dp.IsOpen() {
		t.Fatalf("day %d open %v", dp.Day(), dp.IsOpen())
	}
}

func TestRangePointerRadioResetDestroyAndSelectEmpty(t *testing.T) {
	tree := core.NewTree()
	s := NewSlider(tree, nil)
	s.SetMin(0)
	s.SetMax(10)
	s.SetRangeValues(2, 8)
	bounds(tree, s.ElementID(), 0, 0, 100, 20)
	core.NewDispatcher(tree).Dispatch(s.ElementID(), &event.Event{Type: event.MouseDown, GlobalX: 90, GlobalY: 10})
	lo, hi := s.RangeValues()
	if lo != 2 || hi < 8.5 {
		t.Fatalf("range %v %v value %v", lo, hi, s.Value())
	}

	g := NewRadioGroup()
	a := NewRadio(tree, "a", nil)
	b := NewRadio(tree, "b", nil)
	g.Add(a)
	g.Add(b)
	g.SetValue("b")
	if tree.IsFocusable(a.ElementID()) || !tree.IsFocusable(b.ElementID()) {
		t.Fatal("setvalue did not move the tab stop")
	}

	beforePicker := tree.ElementCount()
	dp := NewDatePicker(tree, nil)
	bounds(tree, dp.ElementID(), 0, 0, 200, 32)
	click(tree, dp.ElementID(), 10, 10)
	if tree.ElementCount() < beforePicker+10 {
		t.Fatal("expected day cells")
	}
	dp.Destroy()
	if tree.ElementCount() != beforePicker {
		t.Fatalf("cells leaked, count %d want %d", tree.ElementCount(), beforePicker)
	}

	sel := NewSelect(tree, []SelectOption{{Label: "Apple", Value: "a"}}, nil)
	sel.SetFilterable(true)
	sel.SetInputValue("zzz")
	bounds(tree, sel.ElementID(), 0, 0, 120, 32)
	click(tree, sel.ElementID(), 10, 10)
	click(tree, sel.ElementID(), 10, 40)
	if sel.Value() != "" {
		t.Fatalf("empty filter selected %q", sel.Value())
	}

	tb := NewTable(tree, []TableColumn{{Title: "N", ColKey: "n", Width: 80}}, nil)
	rows := make([][]string, 20)
	for i := range rows {
		rows[i] = []string{"r"}
	}
	tb.SetRows(rows)
	bounds(tree, tb.ElementID(), 0, 0, 100, 200)
	tb.SetScrollY(5 * 48)
	core.NewDispatcher(tree).Dispatch(tb.ElementID(), &event.Event{Type: event.MouseMove, GlobalX: 10, GlobalY: 48 + 4})
	if tb.HoveredRow() != 5 {
		t.Fatalf("hover %d", tb.HoveredRow())
	}

	box := NewCheckbox(tree, "c", nil)
	box.SetChecked(true)
	item := NewFormItem(tree, "C", box, nil)
	form := NewForm(tree, nil)
	form.AppendChild(item)
	form.Reset()
	if box.IsChecked() {
		t.Fatal("reset left checkbox checked")
	}

	tr := NewTree(tree, nil)
	n1 := &TreeNode{Key: "a", Label: "A", Selected: true}
	n2 := &TreeNode{Key: "b", Label: "B"}
	tr.SetRoots([]*TreeNode{n1, n2})
	tree.SetFocused(tr.ElementID(), true)
	key(tree, tr.ElementID(), event.KeyArrowDown)
	key(tree, tr.ElementID(), event.KeyArrowDown)
	key(tree, tr.ElementID(), event.KeyEnter)
	if n1.Selected || !n2.Selected {
		t.Fatal("enter left the previous selection")
	}

	pop := NewPopover(tree, nil)
	if tree.IsFocusable(pop.ElementID()) {
		t.Fatal("closed popover is focusable")
	}
	pop.Open()
	if tree.Focused() != pop.ElementID() {
		t.Fatal("open did not take focus")
	}

	l := NewList(tree, nil)
	l.SetSelectable(true)
	l.SetItemHeight(30)
	l.SetItems([]ListItem{{Value: "a", Title: "A"}, {Value: "b", Title: "B"}})
	bounds(tree, l.ElementID(), 0, 100, 100, 80)
	core.NewDispatcher(tree).Dispatch(l.ElementID(), &event.Event{Type: event.MouseClick, Y: 0, GlobalX: 10, GlobalY: 140})
	if l.SelectedValue() != "b" {
		t.Fatalf("list used local Y, selected %q", l.SelectedValue())
	}

	gn := NewGamepadNavigator(tree, nil)
	in := NewButton(tree, "in", nil)
	out := NewButton(tree, "out", nil)
	dlg := NewDialog(tree, "d", nil)
	tree.AppendChild(tree.Root(), out.ElementID())
	tree.AppendChild(tree.Root(), dlg.ElementID())
	tree.AppendChild(dlg.ElementID(), in.ElementID())
	gn.AddWidget(out)
	gn.AddWidget(in)
	dlg.Open()
	gn.Navigate(NavDown)
	if tree.Focused() == out.ElementID() {
		t.Fatal("gamepad left the dialog scope")
	}
}

func TestFocusReentryDeliversSecondFocus(t *testing.T) {
	tree := core.NewTree()
	a := NewButton(tree, "a", nil)
	b := NewButton(tree, "b", nil)
	n := 0
	tree.AddHandler(a.ElementID(), event.Focus, func(*event.Event) {
		tree.SetFocused(b.ElementID(), true)
	})
	tree.AddHandler(b.ElementID(), event.Focus, func(*event.Event) { n++ })
	tree.SetFocused(a.ElementID(), true)
	if tree.Focused() != b.ElementID() || n != 1 {
		t.Fatalf("focus %d notifications %d", tree.Focused(), n)
	}
}

func TestDisabledAncestorBlocksChild(t *testing.T) {
	tree := core.NewTree()
	box := NewDiv(tree, nil)
	btn := NewButton(tree, "go", nil)
	tree.AppendChild(tree.Root(), box.ElementID())
	tree.AppendChild(box.ElementID(), btn.ElementID())
	bounds(tree, tree.Root(), 0, 0, 200, 100)
	bounds(tree, box.ElementID(), 0, 0, 200, 80)
	bounds(tree, btn.ElementID(), 10, 10, 40, 24)
	tree.SetEnabled(box.ElementID(), false)
	clicks := 0
	btn.OnClick(func() { clicks++ })
	core.NewDispatcher(tree).Dispatch(btn.ElementID(), &event.Event{Type: event.MouseClick, GlobalX: 20, GlobalY: 20})
	if clicks != 0 {
		t.Fatal("click reached a child of a disabled ancestor")
	}
	tree.SetFocused(btn.ElementID(), true)
	if tree.Focused() == btn.ElementID() {
		t.Fatal("programmatic focus entered a disabled ancestor")
	}
	if tree.HitTest(20, 20) == btn.ElementID() {
		t.Fatal("hit test returned a child of a disabled ancestor")
	}
}

func TestGamepadFirstDirectionFocusesCurrent(t *testing.T) {
	tree := core.NewTree()
	gn := NewGamepadNavigator(tree, nil)
	a := NewButton(tree, "a", nil)
	b := NewButton(tree, "b", nil)
	gn.AddWidget(a)
	gn.AddWidget(b)
	tree.AppendChild(tree.Root(), gn.ElementID())
	core.NewDispatcher(tree).Dispatch(gn.ElementID(), &event.Event{Type: event.KeyDown, Key: event.KeyArrowDown})
	if tree.Focused() != a.ElementID() {
		t.Fatalf("first direction focused %d want %d", tree.Focused(), a.ElementID())
	}
}

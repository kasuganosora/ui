package widget

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// helpers
func p3tree() *core.Tree            { return core.NewTree() }
func p3cfg() *Config                { return DefaultConfig() }
func p3cfgText() *Config            { return cfgWithTextRenderer() }
func p3buf() *render.CommandBuffer  { return render.NewCommandBuffer() }
func p3lay(tree *core.Tree, w Widget, x, y, ww, hh float32) {
	tree.SetLayout(w.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(x, y, ww, hh)})
}

// === Affix ===
func TestAffixSetOffsetTopAndDraw(t *testing.T) {
	tree := p3tree()
	a := NewAffix(tree, nil)
	a.SetOffsetTop(100)
	a.SetAffixed(true)
	p3lay(tree, a, 0, 0, 200, 50)
	a.Draw(p3buf())
}

// === Alert ===
func TestAlertFullCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	a := NewAlert(tree, "hello", cfg)
	a.SetMessage("world")
	called := false
	a.OnClose(func() { called = true })
	a.SetClosable(true)
	a.SetAlertType(AlertSuccess)
	p3lay(tree, a, 0, 0, 300, 40)
	a.Draw(p3buf())
	a.SetAlertType(AlertWarning)
	a.Draw(p3buf())
	a.SetAlertType(AlertError)
	a.Draw(p3buf())
	a.Close()
	if !called {
		t.Error("onClose not called")
	}
	if a.IsVisible() {
		t.Error("should not be visible after close")
	}
	a.Draw(p3buf())
}

// === Anchor ===
func TestAnchorLinksCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	a := NewAnchor(tree, cfg)
	_ = a.Links()
	a.OnChange(func(string) {})
	a.SetLinks([]AnchorLink{
		{Title: "A", Href: "#a", Children: []AnchorLink{{Title: "B", Href: "#b"}}},
	})
	a.SetActive("#a")
	p3lay(tree, a, 0, 0, 200, 400)
	a.Draw(p3buf())
}

// === AutoComplete ===
func TestAutoCompleteCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	ac := NewAutoComplete(tree, cfg)
	_ = ac.IsOpen()
	ac.OnSelect(func(string) {})
	ac.SetFilterFn(func(input string, items []string) []string { return items })
	ac.SetSuggestions([]string{"alpha", "beta"})
	ac.SetText("al")
	ac.SetOpen(true)
	p3lay(tree, ac, 0, 0, 200, 40)
	ac.Draw(p3buf())
	ac.SelectItem(0)
}

// === Avatar ===
func TestAvatarIconAndBgColor(t *testing.T) {
	tree := p3tree()
	a := NewAvatar(tree, nil)
	a.SetIcon(0)
	a.SetBgColor(uimath.RGBA(1, 0, 0, 1))
	p3lay(tree, a, 0, 0, 40, 40)
	a.Draw(p3buf())
}

// === BackTop ===
func TestBackTopCoverage(t *testing.T) {
	tree := p3tree()
	bt := NewBackTop(tree, nil)
	bt.SetThreshold(200)
	bt.SetPosition(20, 20)
	bt.OnClick(func() {})
	bt.SetVisible(true)
	p3lay(tree, bt, 0, 0, 400, 600)
	bt.Draw(p3buf())
}

// === Badge ===
func TestBadgeDrawCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	b := NewBadge(tree, cfg)
	b.SetColor(uimath.RGBA(1, 0, 0, 1))
	b.SetShowZero(true)
	b.SetCount(0)
	p3lay(tree, b, 0, 0, 100, 30)
	b.Draw(p3buf())
	b.SetCount(5)
	b.Draw(p3buf())
	b.SetCount(999)
	b.SetMaxCount(99)
	b.Draw(p3buf())
	b.SetDot(true)
	b.Draw(p3buf())
}

// === Breadcrumb ===
func TestBreadcrumbCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	b := NewBreadcrumb(tree, cfg)
	b.SetSeparator(">")
	b.OnClick(func(int, string) {})
	b.SetItems([]BreadcrumbItem{{Label: "Home", Href: "/"}, {Label: "Products", Href: "/p"}, {Label: "Detail"}})
	p3lay(tree, b, 0, 0, 400, 30)
	b.Draw(p3buf())
}

// === Calendar ===
func TestCalendarOnSelectAndMonths(t *testing.T) {
	tree := p3tree()
	c := NewCalendar(tree, p3cfg())
	c.OnSelect(func(int, int, int) {})
	c.SetYear(2025)
	c.SetMonth(12)
	c.NextMonth()
	if c.Year() != 2026 || c.Month() != 1 {
		t.Errorf("expected 2026/1, got %d/%d", c.Year(), c.Month())
	}
	c.SetMonth(1)
	c.PrevMonth()
	if c.Month() != 12 {
		t.Errorf("expected 12, got %d", c.Month())
	}
}

// === Card ===
func TestCardDrawCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	c := NewCard(tree, cfg)
	c.SetTitle("Title")
	c.SetBordered(true)
	c.SetBgColor(uimath.RGBA(1, 1, 1, 1))
	c.SetHeaderExtra(nil)
	p3lay(tree, c, 0, 0, 300, 200)
	c.Draw(p3buf())
}

// === Cascader ===
func TestCascaderFullCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	c := NewCascader(tree, cfg)
	_ = c.Options()
	_ = c.IsOpen()
	c.SetOpen(true)
	c.OnChange(func([]string) {})
	c.SetOptions([]*CascaderOption{
		{Label: "A", Value: "a", Children: []*CascaderOption{{Label: "B", Value: "b"}}},
	})
	c.SetSelected([]string{"a", "b"})
	p3lay(tree, c, 0, 0, 300, 40)
	c.Draw(p3buf())
}

// === Collapse ===
func TestCollapseDrawCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	c := NewCollapse(tree, cfg)
	c.SetBordered(true)
	c.OnChange(func([]string) {})
	c.SetPanels([]CollapsePanel{{Key: "1", Title: "One"}, {Key: "2", Title: "Two"}})
	c.Toggle("1")
	c.Toggle("2")
	if !c.IsActive("1") {
		t.Error("expected 1 active")
	}
	c.SetAccordion(true)
	c.Toggle("2")
	p3lay(tree, c, 0, 0, 400, 300)
	c.Draw(p3buf())
}

// === ColorPicker ===
func TestColorPickerFullCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	cp := NewColorPicker(tree, cfg)
	_ = cp.Value()
	cp.SetValue(uimath.RGBA(1, 0, 0, 1))
	cp.SetPresets([]uimath.Color{uimath.RGBA(0, 1, 0, 1)})
	cp.OnChange(func(uimath.Color) {})
	cp.SelectColor(uimath.RGBA(0, 0, 1, 1))
	p3lay(tree, cp, 0, 0, 250, 300)
	cp.Draw(p3buf())
}

// === Comment ===
func TestCommentSetData(t *testing.T) {
	tree := p3tree()
	c := NewComment(tree, CommentData{Author: "A"}, p3cfg())
	c.SetData(CommentData{Author: "B", Content: "hello", Time: "now"})
}

// === ContextMenu ===
func TestContextMenuDrawCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	cm := NewContextMenu(tree, cfg)
	cm.SetWidth(150)
	cm.AddItem(ContextMenuItem{Label: "Cut"})
	cm.AddItem(ContextMenuItem{Label: "Copy"})
	cm.Show(100, 200)
	cm.Draw(p3buf())
	cm.ClearItems()
	if len(cm.Items()) != 0 {
		t.Error("expected empty items")
	}
}

// === DatePicker ===
func TestDatePickerCoverage(t *testing.T) {
	tree := p3tree()
	dp := NewDatePicker(tree, nil)
	_ = dp.IsOpen()
	dp.SetOpen(true)
	dp.OnChange(func(int, int, int) {})
	p3lay(tree, dp, 0, 0, 300, 300)
	dp.SetDate(2024, 2, 1)
	dp.Draw(p3buf())
	dp.SetDate(2023, 2, 1)
	dp.Draw(p3buf())
}

// === TimePicker ===
func TestTimePickerCoverage(t *testing.T) {
	tree := p3tree()
	tp := NewTimePicker(tree, nil)
	_ = tp.Second()
	_ = tp.IsOpen()
	tp.SetShowSeconds(true)
	tp.SetOpen(true)
	tp.OnChange(func(int, int, int) {})
	p3lay(tree, tp, 0, 0, 200, 200)
	tp.Draw(p3buf())
}

// === DateRangePicker ===
func TestDateRangePickerCoverage(t *testing.T) {
	tree := p3tree()
	drp := NewDateRangePicker(tree, nil)
	_, _, _ = drp.EndDate()
	_ = drp.IsOpen()
	drp.SetOpen(true)
	drp.SetEndDate(2025, 12, 31)
	drp.OnChange(func(int, int, int, int, int, int) {})
}

// === Descriptions ===
func TestDescriptionsCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	d := NewDescriptions(tree, cfg)
	_ = d.Title()
	d.SetTitle("Details")
	d.SetColumns(2)
	d.SetBordered(true)
	d.AddItem(DescriptionItem{Label: "Name", Value: "John"})
	d.AddItem(DescriptionItem{Label: "Age", Value: "30"})
	p3lay(tree, d, 0, 0, 400, 200)
	d.Draw(p3buf())
}

// === Divider ===
func TestDividerDrawCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	d := NewDivider(tree, cfg)
	d.SetColor(uimath.RGBA(0, 0, 0, 1))
	d.SetThickness(2)
	d.SetText("OR")
	d.SetDirection(DividerHorizontal)
	p3lay(tree, d, 0, 0, 400, 20)
	d.Draw(p3buf())
	d.SetDirection(DividerVertical)
	d.Draw(p3buf())
}

// === Dock ===
func TestDockLayoutCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	dl := NewDockLayout(tree, cfg)
	dl.OnDock(func(string, DockPosition) {})
	dl.OnUndock(func(string) {})
	dl.AddPanel(&DockPanel{ID: "left", Title: "L", Position: DockLeft})
	dl.AddPanel(&DockPanel{ID: "right", Title: "R", Position: DockRight})
	dl.AddPanel(&DockPanel{ID: "top", Title: "T", Position: DockTop})
	dl.AddPanel(&DockPanel{ID: "bottom", Title: "B", Position: DockBottom})
	dl.AddPanel(&DockPanel{ID: "c1", Title: "C1", Position: DockCenter})
	dl.AddPanel(&DockPanel{ID: "c2", Title: "C2", Position: DockCenter})
	dl.AddPanel(&DockPanel{ID: "float", Title: "F", Position: DockFloat})
	p3lay(tree, dl, 0, 0, 1000, 600)
	dl.Draw(p3buf())
	dl.DockPanel("c1", DockLeft)
	dl.UndockPanel("left")
	dl.RemovePanel("right")
	_ = dl.FindPanel("nonexistent")
	_ = dl.FindPanel("top")
}

// === Drawer ===
func TestDrawerCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	d := NewDrawer(tree, "Test", cfg)
	d.SetTitle("Updated")
	d.SetWidth(300)
	d.SetHeight(200)
	d.SetPlacement(DrawerLeft)
	d.SetClosable(true)
	d.OnClose(func() {})
	d.Open()
	if !d.IsVisible() {
		t.Error("should be visible")
	}
	p3lay(tree, d, 0, 0, 800, 600)
	d.Draw(p3buf())
	d.SetPlacement(DrawerRight)
	d.Draw(p3buf())
	d.SetPlacement(DrawerTop)
	d.Draw(p3buf())
	d.SetPlacement(DrawerBottom)
	d.Draw(p3buf())
	d.Close()
}

// === Div ===
func TestDivDrawCoverage(t *testing.T) {
	tree := p3tree()
	d := NewDiv(tree, p3cfg())
	d.SetBgColor(uimath.RGBA(1, 1, 1, 1))
	d.SetBorderColor(uimath.RGBA(0, 0, 0, 1))
	d.SetBorderWidth(1)
	d.SetBorderRadius(4)
	d.SetScrollable(true)
	d.ScrollTo(0, 50)
	_ = d.ScrollY()
	_ = d.ScrollX()
	_ = d.BgColor()
	_ = d.BorderColor()
	_ = d.BorderWidth()
	_ = d.BorderRadius()
	_ = d.IsScrollable()
	p3lay(tree, d, 0, 0, 300, 200)
	d.Draw(p3buf())
}

// === Dialog ===
func TestDialogCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	d := NewDialog(tree, "Confirm", cfg)
	_ = d.Title()
	_ = d.IsVisible()
	d.SetTitle("New Title")
	d.SetWidth(400)
	d.SetContent(nil)
	d.OnClose(func() {})
	d.Open()
	p3lay(tree, d, 0, 0, 800, 600)
	d.Draw(p3buf())
	d.Close()
}

// === Empty ===
func TestEmptyCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	e := NewEmpty(tree, cfg)
	_ = e.Description()
	e.SetDescription("No data")
	p3lay(tree, e, 0, 0, 300, 200)
	e.Draw(p3buf())
}

// === Form ===
func TestFormCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	f := NewForm(tree, cfg)
	btn := NewButton(tree, "Submit", cfg)
	fi := NewFormItem(tree, "Name", btn, cfg)
	fi.SetRequired(true)
	fi.SetError("required")
	p3lay(tree, fi, 0, 0, 400, 60)
	fi.Draw(p3buf())
	p3lay(tree, f, 0, 0, 400, 200)
	f.Draw(p3buf())
}

// === GamepadNavigator ===
type mockNavWidget struct {
	Base
	focusable bool
}

func (m *mockNavWidget) NavFocusable() bool          { return m.focusable }
func (m *mockNavWidget) Draw(*render.CommandBuffer) {}

func newMockNav(tree *core.Tree, focusable bool) *mockNavWidget {
	m := &mockNavWidget{
		Base:      NewBase(tree, core.TypeCustom, DefaultConfig()),
		focusable: focusable,
	}
	return m
}

func TestGamepadNavigatorFullCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfg()
	gn := NewGamepadNavigator(tree, cfg)
	gn.SetDeadzone(0.2)
	gn.SetRepeatDelay(0.5)
	gn.SetRepeatRate(0.1)
	_ = gn.ShowFocus()
	gn.SetShowFocus(false)
	gn.OnNavigate(func(Navigable) {})
	gn.OnCancel(func() {})
	gn.OnActivate(func(Navigable) {})

	w1 := newMockNav(tree, true)
	w2 := newMockNav(tree, true)
	w3 := newMockNav(tree, false)
	gn.AddWidget(w1)
	gn.AddWidget(w2)
	gn.AddWidget(w3)

	gn.Navigate(NavDown)
	gn.Navigate(NavUp)
	gn.Navigate(NavLeft)
	gn.Navigate(NavRight)
	gn.Activate()
	gn.Cancel()
	gn.SetFocus(0)

	gn.SetEnabled(true)
	gn.Tick(0.1)

	gn.SetShowFocus(true)
	p3lay(tree, w1, 10, 10, 100, 30)
	gn.SetFocus(0)
	gn.Draw(p3buf())

	gn.RemoveWidget(w2)
	gn.ClearWidgets()

	gn.AddWidget(w1)
	gn.AddWidget(newMockNav(tree, true))

	ev := &event.Event{Key: event.KeyArrowDown}
	gn.onKeyDown(ev)
	ev.Key = event.KeyArrowUp
	gn.onKeyDown(ev)
	ev.Key = event.KeyArrowLeft
	gn.onKeyDown(ev)
	ev.Key = event.KeyArrowRight
	gn.onKeyDown(ev)
	ev.Key = event.KeyEnter
	gn.onKeyDown(ev)
	ev.Key = event.KeySpace
	gn.onKeyDown(ev)
	ev.Key = event.KeyEscape
	gn.onKeyDown(ev)

	ge := &event.Event{GamepadButton: GPBtnUp}
	gn.onGamepadButton(ge)
	ge.GamepadButton = GPBtnDown
	gn.onGamepadButton(ge)
	ge.GamepadButton = GPBtnLeft
	gn.onGamepadButton(ge)
	ge.GamepadButton = GPBtnRight
	gn.onGamepadButton(ge)
	ge.GamepadButton = GPBtnA
	gn.onGamepadButton(ge)
	ge.GamepadButton = GPBtnB
	gn.onGamepadButton(ge)

	ae := &event.Event{GamepadAxis: GPAxisLeftX, GamepadValue: 0.5}
	gn.onGamepadAxis(ae)
	ae.GamepadValue = -0.5
	gn.onGamepadAxis(ae)
	ae.GamepadValue = 0
	gn.onGamepadAxis(ae)
	ae.GamepadAxis = GPAxisLeftY
	ae.GamepadValue = 0.5
	gn.onGamepadAxis(ae)
	ae.GamepadValue = -0.5
	gn.onGamepadAxis(ae)
	ae.GamepadValue = 0
	gn.onGamepadAxis(ae)

	gn.moved = true
	gn.lastDir = NavDown
	gn.axisAccum = 0
	gn.Tick(0.5)
}

// === Grid ===
func TestGridRowColDraw(t *testing.T) {
	tree := p3tree()
	cfg := p3cfg()
	r := NewRow(tree, cfg)
	p3lay(tree, r, 0, 0, 600, 40)
	r.Draw(p3buf())
	c := NewCol(tree, 12, cfg)
	p3lay(tree, c, 0, 0, 600, 40)
	c.Draw(p3buf())
}

// === Guide ===
func TestGuideCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	g := NewGuide(tree, cfg)
	_ = g.Steps()
	g.SetMaskColor(uimath.RGBA(0, 0, 0, 0.5))
	g.OnChange(func(int) {})
	g.SetSteps([]GuideStep{
		{Title: "Step 1", Description: "First", TargetX: 10, TargetY: 10, TargetW: 100, TargetH: 30},
		{Title: "Step 2", Description: "Second", TargetX: 200, TargetY: 10, TargetW: 100, TargetH: 30},
	})
	g.Start()
	p3lay(tree, g, 0, 0, 800, 600)
	g.Draw(p3buf())
}

// === Icon ===
func TestIconCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfg()
	ic := NewIcon(tree, "star", cfg)
	_ = ic.Texture()
	ic.SetColor(uimath.RGBA(1, 0, 0, 1))
	ic.SetTexture(0, uimath.NewRect(0, 0, 1, 1))
	p3lay(tree, ic, 0, 0, 24, 24)
	ic.Draw(p3buf())
}

// === Image ===
func TestImageCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfg()
	img := NewImage(tree, 0, cfg)
	_ = img.Texture()
	img.SetTexture(1)
	p3lay(tree, img, 0, 0, 200, 100)
	img.Draw(p3buf())
}

// === ImageViewer ===
func TestImageViewerCoverage(t *testing.T) {
	tree := p3tree()
	iv := NewImageViewer(tree, p3cfg())
	_ = iv.Texture()
	_ = iv.IsVisible()
	iv.SetTexture(1)
	iv.SetVisible(true)
	iv.OnClose(func() {})
	p3lay(tree, iv, 0, 0, 800, 600)
	iv.Draw(p3buf())
}

// === InputNumber ===
func TestInputNumberCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	n := NewInputNumber(tree, cfg)
	_ = n.Min()
	_ = n.Max()
	_ = n.Step()
	n.SetMin(0)
	n.SetMax(100)
	n.SetStep(5)
	n.SetDisabled(false)
	n.OnChange(func(float64) {})
	n.SetValue(50)
	n.Increment()
	n.Decrement()
	n.SetValue(-10)
	n.SetValue(200)
	p3lay(tree, n, 0, 0, 200, 40)
	n.Draw(p3buf())
}

// === Layout ===
func TestLayoutWidgetsCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfg()
	l := NewLayout(tree, cfg)
	l.SetBgColor(uimath.RGBA(1, 1, 1, 1))
	p3lay(tree, l, 0, 0, 800, 600)
	l.Draw(p3buf())

	h := NewHeader(tree, cfg)
	h.SetBgColor(uimath.RGBA(0.9, 0.9, 0.9, 1))
	h.SetHeight(60)
	p3lay(tree, h, 0, 0, 800, 60)
	h.Draw(p3buf())

	c := NewContent(tree, cfg)
	c.SetBgColor(uimath.RGBA(1, 1, 1, 1))
	_ = c.ScrollY()
	_ = c.ContentHeight()
	c.SetContentHeight(1000)
	p3lay(tree, c, 0, 60, 800, 400)
	c.Draw(p3buf())

	f := NewFooter(tree, cfg)
	f.SetBgColor(uimath.RGBA(0.9, 0.9, 0.9, 1))
	f.SetHeight(40)
	p3lay(tree, f, 0, 460, 800, 40)
	f.Draw(p3buf())

	a := NewAside(tree, cfg)
	a.SetBgColor(uimath.RGBA(0.95, 0.95, 0.95, 1))
	a.SetBorderRight(1, uimath.RGBA(0, 0, 0, 0.1))
	a.SetWidth(200)
	p3lay(tree, a, 0, 0, 200, 500)
	a.Draw(p3buf())
}

// === Link ===
func TestLinkDrawCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	l := NewLink(tree, "Click", "/path", cfg)
	l.OnClick(func(string) {})
	p3lay(tree, l, 0, 0, 200, 30)
	l.Draw(p3buf())
	l.SetDisabled(true)
	l.Draw(p3buf())
}

// === List ===
func TestListCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	l := NewList(tree, cfg)
	l.SetItemHeight(40)
	l.SetBordered(true)
	l.OnSelect(func(int) {})
	_ = l.ScrollY()
	l.SetScrollY(0)
	l.SetItems([]ListItem{
		{Title: "Item 1", Description: "Desc 1"},
		{Title: "Item 2", Description: "Desc 2"},
	})
	p3lay(tree, l, 0, 0, 300, 200)
	l.Draw(p3buf())
}

// === VirtualList ===
func TestVirtualListCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfg()
	vl := NewVirtualList(tree, cfg)
	vl.SetItemHeight(30)
	vl.SetScrollY(0)
	_ = vl.ScrollY()
	vl.SetRenderItem(func(idx int, buf *render.CommandBuffer, x, y, w, h float32) {})
	vl.SetItemCount(100)
	_ = vl.ContentHeight()
	p3lay(tree, vl, 0, 0, 300, 400)
	vl.Draw(p3buf())
}

// === Loading ===
func TestLoadingCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	l := NewLoading(tree, cfg)
	_ = l.Tip()
	l.SetTip("Loading...")
	p3lay(tree, l, 0, 0, 200, 100)
	l.Draw(p3buf())
}

// === Menu ===
func TestMenuCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	m := NewMenu(tree, cfg)
	m.SetSelectedKey("home")
	m.OnSelect(func(string) {})
	m.SetItems([]MenuItem{
		{Key: "home", Label: "Home"},
		{Key: "about", Label: "About", Children: []MenuItem{{Key: "team", Label: "Team"}}},
	})
	m.ToggleOpen("about")
	m.SelectItem("home")
	p3lay(tree, m, 0, 0, 200, 400)
	m.Draw(p3buf())
}

// === MenuBar ===
func TestMenuBarCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	mb := NewMenuBar(tree, cfg)
	mb.SetHeight(40)
	_ = mb.IsOpen()
	mb.AddItem(MenuBarItem{Label: "File"})
	mb.AddItem(MenuBarItem{Label: "Edit"})
	p3lay(tree, mb, 0, 0, 600, 40)
	mb.Draw(p3buf())
	mb.ClearItems()
}

// === Message ===
func TestMessageCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	m := NewMessage(tree, "hello", cfg)
	_ = m.IsVisible()
	m.SetContent("world")
	m.SetVisible(true)
	p3lay(tree, m, 0, 0, 400, 40)
	m.Draw(p3buf())
}

// === MessageBox ===
func TestMessageBoxCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	mb := NewMessageBox(tree, "Confirm", "Are you sure?", cfg)
	_ = mb.Title()
	_ = mb.Content()
	_ = mb.BoxType()
	_ = mb.IsVisible()
	mb.SetTitle("New Title")
	mb.SetContent("New Content")
	mb.SetBoxType(MessageBoxInfo)
	mb.SetWidth(400)
	mb.SetShowCancel(true)
	mb.OnOK(func() {})
	mb.OnCancel(func() {})
	mb.Open()
	p3lay(tree, mb, 0, 0, 800, 600)
	mb.Draw(p3buf())
	mb.SetBoxType(MessageBoxSuccess)
	mb.Draw(p3buf())
	mb.SetBoxType(MessageBoxWarning)
	mb.Draw(p3buf())
	mb.SetBoxType(MessageBoxError)
	mb.Draw(p3buf())
	mb.Close()
}

// === Notification ===
func TestNotificationCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	n := NewNotification(tree, "Title", "Message", cfg)
	_ = n.Title()
	_ = n.Message()
	n.SetTitle("T")
	n.SetMessage("M")
	n.SetPosition(10, 10)
	n.OnClose(func() {})
	n.Show()
	p3lay(tree, n, 0, 0, 400, 100)
	n.Draw(p3buf())
}

// === Pagination ===
func TestPaginationCoverage(t *testing.T) {
	tree := p3tree()
	p := NewPagination(tree, p3cfg())
	_ = p.Total()
	_ = p.PageSize()
	p.OnChange(func(int) {})
	p.SetTotal(100)
	p.SetPageSize(10)
	p.SetCurrent(5)
	_ = p.TotalPages()
}

// === Panel ===
func TestPanelDrawCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	p := NewPanel(tree, "Title", cfg)
	p.SetTitle("New")
	p.SetBgColor(uimath.RGBA(1, 1, 1, 1))
	p.SetBordered(true)
	p3lay(tree, p, 0, 0, 300, 200)
	p.Draw(p3buf())
}

// === Popconfirm ===
func TestPopconfirmCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	p := NewPopconfirm(tree, "Sure?", cfg)
	_ = p.Title()
	p.SetTitle("Really?")
	p.SetPlacement(PlacementTop)
	p.SetAnchorRect(10, 10, 100, 30)
	p.OnCancel(func() {})
	p.Show()
	p3lay(tree, p, 0, 0, 800, 600)
	p.Draw(p3buf())
}

// === Popover ===
func TestPopoverCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	p := NewPopover(tree, cfg)
	_ = p.IsVisible()
	p.SetVisible(true)
	p.SetTitle("Info")
	p.SetContent(nil)
	p.SetPlacement(PlacementTop)
	p.SetTrigger(PopoverTriggerHover)
	p.SetWidth(200)
	p.OnClose(func() {})
	p.SetAnchorRect(50, 50, 100, 30)
	p.Open()
	p3lay(tree, p, 0, 0, 800, 600)
	p.Draw(p3buf())
	p.Close()
}

// === Portal ===
func TestPortalCoverage(t *testing.T) {
	tree := p3tree()
	p := NewPortal(tree, p3cfg())
	_ = p.IsVisible()
	p.SetVisible(true)
	p.SetZBase(100)
	p.SetContent(nil)
	p3lay(tree, p, 0, 0, 800, 600)
	p.Draw(p3buf())
}

// === Popup ===
func TestPopupCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfg()
	p := NewPopup(tree, cfg)
	_ = p.IsVisible()
	_ = p.Placement()
	_ = p.AnchorID()
	p.SetVisible(true)
	p.SetPlacement(PlacementBottom)
	p.SetAnchor(core.ElementID(1))
	p.SetBgColor(uimath.RGBA(1, 1, 1, 1))
	p.SetShadow(true)
	p3lay(tree, p, 0, 0, 800, 600)
	p.Draw(p3buf())
}

// === Progress ===
func TestProgressCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	p := NewProgress(tree, cfg)
	p.SetPercent(75)
	p3lay(tree, p, 0, 0, 300, 20)
	p.Draw(p3buf())
}

// === RangeInput ===
func TestRangeInputCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	ri := NewRangeInput(tree, 0, 100, cfg)
	ri.SetStep(5)
	ri.OnChange(func(float32, float32) {})
	p3lay(tree, ri, 0, 0, 300, 40)
	ri.Draw(p3buf())
}

// === Rate ===
func TestRateCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	r := NewRate(tree, cfg)
	_ = r.Value()
	r.SetValue(3)
	r.SetStarSize(24)
	r.SetDisabled(true)
	r.OnChange(func(int) {})
	p3lay(tree, r, 0, 0, 200, 30)
	r.Draw(p3buf())
}

// === RichText ===
func TestRichTextCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	rt := NewRichText(tree, cfg)
	rt.AddSpan(RichSpan{Type: RichSpanText, Text: "Hello"})
	rt.AddText(" World")
	rt.AddStyledText("Bold", uimath.RGBA(1, 0, 0, 1), 16, true)
	rt.AddBreak()
	rt.AddImage(0, 100, 50)
	rt.SetLineSpacing(1.5)
	_ = rt.Spans()
	p3lay(tree, rt, 0, 0, 400, 300)
	rt.Draw(p3buf())
	rt.ClearSpans()
}

// === Slider ===
func TestSliderCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfg()
	s := NewSlider(tree, cfg)
	_ = s.Value()
	_ = s.Min()
	_ = s.Max()
	s.SetMin(0)
	s.SetMax(100)
	s.SetStep(5)
	s.SetDisabled(false)
	s.OnChange(func(float32) {})
	s.SetValue(50)
	s.SetValue(-10)
	s.SetValue(200)
	p3lay(tree, s, 0, 0, 300, 30)
	s.Draw(p3buf())
}

// === Splitter ===
func TestSplitterCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfg()
	s := NewSplitter(tree, cfg)
	s.SetDirection(SplitterHorizontal)
	s.SetMinRatio(0.1)
	s.SetMaxRatio(0.9)
	s.SetFirst(nil)
	s.SetSecond(nil)
	p3lay(tree, s, 0, 0, 600, 400)
	s.Draw(p3buf())
}

// === Statistic ===
func TestStatisticCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	s := NewStatistic(tree, "Users", "1234", cfg)
	_ = s.Title()
	_ = s.Value()
	s.SetTitle("Orders")
	s.SetValue("5678")
	s.SetColor(uimath.RGBA(0, 1, 0, 1))
	p3lay(tree, s, 0, 0, 200, 80)
	s.Draw(p3buf())
}

// === Steps ===
func TestStepsCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	s := NewSteps(tree, cfg)
	_ = s.Items()
	s.AddStep(StepItem{Title: "Step 1"})
	s.AddStep(StepItem{Title: "Step 2"})
	s.SetCurrent(1)
	p3lay(tree, s, 0, 0, 600, 80)
	s.Draw(p3buf())
	s.ClearSteps()
}

// === SubWindow ===
func TestSubWindowCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	w := NewSubWindow(tree, "Win", cfg)
	w.SetTitle("Window")
	w.SetClosable(true)
	w.OnClose(func() {})
	_ = w.IsVisible()
	w.Open()
	w.SetPosition(100, 100)
	p3lay(tree, w, 0, 0, 800, 600)
	w.Draw(p3buf())
	w.Close()
}

// === Swiper ===
func TestSwiperCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfg()
	s := NewSwiper(tree, cfg)
	s.SetCurrent(0)
	s.SetAutoplay(true)
	s.SetShowDots(true)
	s.OnChange(func(int) {})
	s.AddPanel(NewDiv(tree, cfg))
	s.AddPanel(NewDiv(tree, cfg))
	s.AddPanel(NewDiv(tree, cfg))
	s.Next()
	s.Next()
	s.Next()
	s.Prev()
	s.Prev()
	s.Prev()
	p3lay(tree, s, 0, 0, 600, 300)
	s.Draw(p3buf())
}

// === Table ===
func TestTableCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	cols := []TableColumn{{Title: "Name", Width: 100}, {Title: "Age", Width: 50}}
	tb := NewTable(tree, cols, cfg)
	_ = tb.Columns()
	_ = tb.Rows()
	tb.SetStriped(true)
	tb.SetBordered(true)
	tb.SetRowHeight(40)
	tb.SetRows([][]string{{"Alice", "30"}, {"Bob", "25"}})
	p3lay(tree, tb, 0, 0, 400, 200)
	tb.Draw(p3buf())
	tb.ClearRows()
}

// === TagInput ===
func TestTagInputCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	ti := NewTagInput(tree, cfg)
	ti.SetMaxTags(5)
	ti.OnAdd(func(string) {})
	ti.OnRemove(func(string, int) {})
	ti.AddTag("Go")
	ti.AddTag("Rust")
	ti.RemoveTag(0)
	p3lay(tree, ti, 0, 0, 300, 40)
	ti.Draw(p3buf())
}

// === Timeline ===
func TestTimelineCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	tl := NewTimeline(tree, cfg)
	_ = tl.Items()
	tl.SetItemHeight(60)
	tl.AddItem(TimelineItem{Label: "Created", Status: TimelineDefault})
	tl.AddItem(TimelineItem{Label: "In Progress", Status: TimelineSuccess})
	tl.AddItem(TimelineItem{Label: "Error", Status: TimelineError})
	tl.AddItem(TimelineItem{Label: "Warning", Status: TimelineWarning})
	p3lay(tree, tl, 0, 0, 400, 400)
	tl.Draw(p3buf())
	tl.ClearItems()
}

// === Tooltip ===
func TestTooltipCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	anchor := NewButton(tree, "Hover", cfg)
	tt := NewTooltip(tree, "Tooltip text", anchor.ElementID(), cfg)
	_ = tt.IsVisible()
	tt.SetText("Updated")
	tt.SetPlacement(PlacementTop)
	p3lay(tree, anchor, 100, 100, 80, 30)
	tt.Show()
	p3lay(tree, tt, 0, 0, 800, 600)
	tt.Draw(p3buf())
}

// === Transfer ===
func TestTransferCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	tr := NewTransfer(tree, cfg)
	tr.OnChange(func([]string) {})
	tr.SetSource([]TransferItem{
		{Key: "1", Label: "Item 1"},
		{Key: "2", Label: "Item 2"},
	})
	tr.MoveToTarget([]string{"1"})
	tr.MoveToSource([]string{"1"})
	p3lay(tree, tr, 0, 0, 600, 300)
	tr.Draw(p3buf())
}

// === TreeSelect ===
func TestTreeSelectCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	ts := NewTreeSelect(tree, cfg)
	_ = ts.IsOpen()
	ts.SetOpen(true)
	ts.OnChange(func(string) {})
	ts.SetRoots([]*TreeNode{
		{Key: "1", Label: "Root", Children: []*TreeNode{{Key: "1-1", Label: "Child"}}},
	})
	ts.SetSelected("1-1")
	p3lay(tree, ts, 0, 0, 300, 40)
	ts.Draw(p3buf())
}

// === Tree widget ===
func TestTreeWidgetCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	tw := NewTree(tree, cfg)
	_ = tw.Roots()
	tw.SetIndent(20)
	tw.SetItemHeight(28)
	tw.OnSelect(func(*TreeNode) {})
	tw.OnExpand(func(*TreeNode) {})
	tw.SetRoots([]*TreeNode{
		{Key: "1", Label: "Root", Expanded: true, Children: []*TreeNode{
			{Key: "1-1", Label: "Child", Selected: true},
		}},
	})
	_ = tw.FindNode("1-1")
	_ = tw.FindNode("nonexistent")
	p3lay(tree, tw, 0, 0, 300, 400)
	tw.Draw(p3buf())
}

// === Upload ===
func TestUploadCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	u := NewUpload(tree, cfg)
	u.SetMultiple(true)
	u.SetAccept(".png,.jpg")
	u.SetDrag(true)
	u.SetMaxCount(5)
	u.OnUpload(func([]UploadFile) {})
	u.AddFile(UploadFile{Name: "test.png", Size: 1024, Status: "done"})
	u.AddFile(UploadFile{Name: "test2.jpg", Size: 2048, Status: "uploading", Progress: 0.5})
	p3lay(tree, u, 0, 0, 400, 200)
	u.Draw(p3buf())
	u.ClearFiles()
}

// === VirtualGrid ===
func TestVirtualGridCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfg()
	vg := NewVirtualGrid(tree, 3, 10, cfg)
	vg.SetCols(4)
	vg.SetRowCount(20)
	vg.SetGap(8)
	vg.SetScrollY(0)
	_ = vg.ScrollY()
	p3lay(tree, vg, 0, 0, 400, 300)
	vg.Draw(p3buf())
}

// === Watermark ===
func TestWatermarkCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	w := NewWatermark(tree, "Draft", cfg)
	w.SetText("Confidential")
	w.SetGap(100, 80)
	w.SetColor(uimath.RGBA(0, 0, 0, 0.1))
	p3lay(tree, w, 0, 0, 800, 600)
	w.Draw(p3buf())
}

// === Text ===
func TestTextWidgetCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfgText()
	tx := NewText(tree, "Hello", cfg)
	tx.SetText("World")
	tx.SetColor(uimath.RGBA(0, 0, 0, 1))
	p3lay(tree, tx, 0, 0, 200, 30)
	tx.Draw(p3buf())
}

// === Skeleton ===
func TestSkeletonDrawCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfg()
	s := NewSkeleton(tree, cfg)
	s.SetRows(5)
	_ = s.Rows()
	p3lay(tree, s, 0, 0, 400, 200)
	s.Draw(p3buf())
}

// === DragDrop ===
func TestDragDropCoverage(t *testing.T) {
	tree := p3tree()
	cfg := p3cfg()
	dd := NewDragDropManager(tree)
	src := NewButton(tree, "Drag", cfg)
	tgt := NewButton(tree, "Drop", cfg)
	dd.RegisterSource(&DragSource{Widget: src, Data: "hello"})
	dd.RegisterTarget(&DropTarget{
		Widget: tgt,
		Accept: func(data any) bool { return true },
		OnDrop: func(data any) {},
	})
	p3lay(tree, src, 10, 10, 80, 30)
	p3lay(tree, tgt, 200, 10, 80, 30)
	ev := &event.Event{Type: event.MouseDown, GlobalX: 20, GlobalY: 20}
	dd.HandleEvent(ev)
	ev2 := &event.Event{Type: event.MouseMove, GlobalX: 210, GlobalY: 20}
	dd.HandleEvent(ev2)
	ev3 := &event.Event{Type: event.MouseUp, GlobalX: 210, GlobalY: 20}
	dd.HandleEvent(ev3)
}

// === Nil config constructors ===
func TestNilConfigConstructorsP3(t *testing.T) {
	tree := p3tree()
	_ = NewAffix(tree, nil)
	_ = NewAlert(tree, "msg", nil)
	_ = NewAnchor(tree, nil)
	_ = NewAutoComplete(tree, nil)
	_ = NewAvatar(tree, nil)
	_ = NewBackTop(tree, nil)
	_ = NewBreadcrumb(tree, nil)
	_ = NewCard(tree, nil)
	_ = NewCascader(tree, nil)
	_ = NewCollapse(tree, nil)
	_ = NewColorPicker(tree, nil)
	_ = NewDatePicker(tree, nil)
	_ = NewDateRangePicker(tree, nil)
	_ = NewDialog(tree, "t", nil)
	_ = NewDivider(tree, nil)
	_ = NewDockLayout(tree, nil)
	_ = NewDrawer(tree, "t", nil)
	_ = NewEmpty(tree, nil)
	_ = NewForm(tree, nil)
	_ = NewGamepadNavigator(tree, nil)
	_ = NewIcon(tree, "x", nil)
	_ = NewImage(tree, 0, nil)
	_ = NewImageViewer(tree, nil)
	_ = NewInputNumber(tree, nil)
	_ = NewLayout(tree, nil)
	_ = NewHeader(tree, nil)
	_ = NewContent(tree, nil)
	_ = NewFooter(tree, nil)
	_ = NewAside(tree, nil)
	_ = NewLink(tree, "a", "b", nil)
	_ = NewLoading(tree, nil)
	_ = NewMenu(tree, nil)
	_ = NewMessage(tree, "m", nil)
	_ = NewMessageBox(tree, "t", "c", nil)
	_ = NewPagination(tree, nil)
	_ = NewPanel(tree, "t", nil)
	_ = NewPopconfirm(tree, "t", nil)
	_ = NewPopover(tree, nil)
	_ = NewPopup(tree, nil)
	_ = NewPortal(tree, nil)
	_ = NewProgress(tree, nil)
	_ = NewRate(tree, nil)
	_ = NewRangeInput(tree, 0, 100, nil)
	_ = NewRichText(tree, nil)
	_ = NewRow(tree, nil)
	_ = NewCol(tree, 6, nil)
	_ = NewSlider(tree, nil)
	_ = NewSplitter(tree, nil)
	_ = NewSubWindow(tree, "t", nil)
	_ = NewSwiper(tree, nil)
	_ = NewTable(tree, nil, nil)
	_ = NewText(tree, "t", nil)
	_ = NewTooltip(tree, "t", 0, nil)
	_ = NewUpload(tree, nil)
	_ = NewWatermark(tree, "t", nil)
	_ = NewDiv(tree, nil)
}

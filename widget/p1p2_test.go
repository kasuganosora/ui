package widget

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

func p1p2TestCfg() *Config { return DefaultConfig() }

// drawAndVerify sets bounds, calls Draw, and verifies output.
func drawAndVerify(t *testing.T, tree *core.Tree, w Widget, expectOverlay bool) {
	t.Helper()
	setBounds(tree, w, 0, 0, 400, 300)
	buf := render.NewCommandBuffer()
	w.Draw(buf)
	if expectOverlay {
		if len(buf.Overlays()) == 0 && buf.Len() == 0 {
			t.Error("expected overlay or draw output")
		}
	} else {
		if buf.Len() == 0 {
			t.Error("expected draw output")
		}
	}
}

// --- Affix ---

func TestAffixWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	a := NewAffix(tree, cfg)
	btn := NewButton(tree, "test", cfg)
	a.SetContent(btn)
	if a.Content() == nil {
		t.Error("expected content")
	}
	a.SetOffsetTop(10)
	a.SetAffixed(true)
	if !a.IsAffixed() {
		t.Error("expected affixed")
	}
	a.SetAffixed(false)
	if a.IsAffixed() {
		t.Error("expected not affixed")
	}
	setBounds(tree, a, 0, 0, 200, 40)
	setBounds(tree, btn, 0, 0, 200, 40)
	buf := render.NewCommandBuffer()
	a.Draw(buf)
}

func TestAffixDrawAffixed(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	a := NewAffix(tree, cfg)
	btn := NewButton(tree, "test", cfg)
	a.SetContent(btn)
	a.SetAffixed(true)
	setBounds(tree, a, 0, 0, 200, 40)
	setBounds(tree, btn, 0, 0, 200, 40)
	buf := render.NewCommandBuffer()
	a.Draw(buf)
}

func TestAffixDrawEmpty(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	a := NewAffix(tree, cfg)
	buf := render.NewCommandBuffer()
	a.Draw(buf) // no bounds, no content
}

// --- Alert ---

func TestAlertWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	a := NewAlert(tree, "Warning message", cfg)
	a.SetTheme(AlertThemeWarning)
	a.SetCloseBtn(true)
	a.SetMessage("New message")
	if !a.IsVisible() {
		t.Error("alert should be visible")
	}
	drawAndVerify(t, tree, a, false)

	// Test all alert types
	for _, at := range []AlertTheme{AlertThemeSuccess, AlertThemeInfo, AlertThemeWarning, AlertThemeError} {
		a2 := NewAlert(tree, "test", cfg)
		a2.SetTheme(at)
		setBounds(tree, a2, 0, 0, 300, 50)
		buf := render.NewCommandBuffer()
		a2.Draw(buf)
		if buf.Len() == 0 {
			t.Errorf("expected output for alert type %d", at)
		}
	}
}

func TestAlertClose(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	a := NewAlert(tree, "msg", cfg)
	closed := false
	a.OnClose(func() { closed = true })
	a.Close()
	if a.IsVisible() {
		t.Error("should be hidden after close")
	}
	if !closed {
		t.Error("expected close callback")
	}
}

func TestAlertDrawHidden(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	a := NewAlert(tree, "msg", cfg)
	a.Close()
	setBounds(tree, a, 0, 0, 300, 50)
	buf := render.NewCommandBuffer()
	a.Draw(buf)
	if buf.Len() != 0 {
		t.Error("hidden alert should not draw")
	}
}

// --- Anchor ---

func TestAnchorWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	a := NewAnchor(tree, cfg)
	a.SetLinks([]AnchorLink{
		{Title: "Section 1", Href: "#s1"},
		{Title: "Section 2", Href: "#s2", Children: []AnchorLink{
			{Title: "Sub 1", Href: "#s2-1"},
		}},
	})
	if len(a.Links()) != 2 {
		t.Errorf("expected 2 links")
	}
	a.SetActive("#s1")
	if a.Active() != "#s1" {
		t.Errorf("expected active '#s1'")
	}
	changed := false
	a.OnChange(func(href string) { changed = true })
	_ = changed
	drawAndVerify(t, tree, a, false)
}

// --- AutoComplete ---

func TestAutoCompleteWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	ac := NewAutoComplete(tree, cfg)
	ac.SetSuggestions([]string{"Apple", "Application", "Banana"})

	if ac.IsOpen() {
		t.Error("should not be open initially")
	}

	ac.SetText("App")
	if len(ac.Filtered()) != 2 {
		t.Errorf("expected 2 filtered, got %d", len(ac.Filtered()))
	}

	selected := ""
	ac.OnSelect(func(s string) { selected = s })
	ac.SelectItem(0)
	if ac.Text() != "Apple" {
		t.Errorf("expected 'Apple', got %q", ac.Text())
	}
	if selected != "Apple" {
		t.Errorf("expected selected 'Apple'")
	}

	// Custom filter
	ac.SetFilterFn(func(input string, opts []string) []string {
		return []string{"custom"}
	})
	ac.SetText("x")
	if len(ac.Filtered()) != 1 || ac.Filtered()[0] != "custom" {
		t.Error("custom filter not applied")
	}

	ac.SetOpen(true)
	drawAndVerify(t, tree, ac, false)
}

func TestAutoCompleteSelectOutOfRange(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	ac := NewAutoComplete(tree, cfg)
	ac.SetSuggestions([]string{"A"})
	ac.SelectItem(10) // out of range
}

// --- Avatar ---

func TestAvatarWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	a := NewAvatar(tree, cfg)
	a.SetText("AB")
	a.SetShape(AvatarSquare)
	a.SetSize(40)
	a.SetBgColor(uimath.ColorRed)
	a.SetIcon(0)
	drawAndVerify(t, tree, a, false)
}

func TestAvatarCircle(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	a := NewAvatar(tree, cfg)
	a.SetText("XY")
	a.SetShape(AvatarCircle)
	drawAndVerify(t, tree, a, false)
}

// --- BackTop ---

func TestBackTopWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	bt := NewBackTop(tree, cfg)
	bt.SetThreshold(100)
	bt.SetPosition(20, 20)
	clicked := false
	bt.OnClick(func() { clicked = true })

	bt.SetVisible(true)
	if !bt.IsVisible() {
		t.Error("expected visible")
	}
	setBounds(tree, bt, 0, 0, 800, 600)
	buf := render.NewCommandBuffer()
	bt.Draw(buf)
	// BackTop uses overlays
	_ = clicked

	bt.SetVisible(false)
	buf = render.NewCommandBuffer()
	bt.Draw(buf)
}

// --- Badge ---

func TestBadgeWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	b := NewBadge(tree, cfg)
	b.SetCount(5)
	if b.Count() != 5 {
		t.Errorf("expected count 5, got %d", b.Count())
	}
	b.SetDot(true)
	b.SetMaxCount(99)
	b.SetColor(uimath.ColorRed)
	b.SetShowZero(true)
	drawAndVerify(t, tree, b, false)
}

func TestBadgeZeroCount(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	b := NewBadge(tree, cfg)
	b.SetCount(0)
	b.SetShowZero(false)
	setBounds(tree, b, 0, 0, 100, 30)
	buf := render.NewCommandBuffer()
	b.Draw(buf)
}

func TestBadgeOverMax(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	b := NewBadge(tree, cfg)
	b.SetCount(150)
	b.SetMaxCount(99)
	drawAndVerify(t, tree, b, false)
}

// --- Breadcrumb ---

func TestBreadcrumbWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	b := NewBreadcrumb(tree, cfg)
	b.SetOptions([]BreadcrumbItem{
		{Content: "Home"},
		{Content: "Products"},
		{Content: "Detail"},
	})
	b.SetSeparator(">")
	clicked := false
	b.OnClick(func(idx int, label string) { clicked = true })
	_ = clicked
	drawAndVerify(t, tree, b, false)
}

// --- Calendar ---

func TestCalendarWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	c := NewCalendar(tree, cfg)
	c.SetYear(2026)
	c.SetMonth(3)
	c.SetSelected(15)
	if c.Year() != 2026 || c.Month() != 3 || c.Selected() != 15 {
		t.Error("expected date state")
	}

	selected := false
	c.OnSelect(func(y, m, d int) { selected = true })
	_ = selected

	c.NextMonth()
	if c.Month() != 4 {
		t.Errorf("expected month 4 after next")
	}
	c.PrevMonth()
	if c.Month() != 3 {
		t.Errorf("expected month 3 after prev")
	}

	drawAndVerify(t, tree, c, false)
}

func TestCalendarDecemberNextMonth(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	c := NewCalendar(tree, cfg)
	c.SetYear(2026)
	c.SetMonth(12)
	c.NextMonth()
	if c.Month() != 1 || c.Year() != 2027 {
		t.Errorf("expected 2027-01, got %d-%d", c.Year(), c.Month())
	}
}

func TestCalendarJanuaryPrevMonth(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	c := NewCalendar(tree, cfg)
	c.SetYear(2026)
	c.SetMonth(1)
	c.PrevMonth()
	if c.Month() != 12 || c.Year() != 2025 {
		t.Errorf("expected 2025-12, got %d-%d", c.Year(), c.Month())
	}
}

func TestCalendarDrawAllMonths(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	for m := 1; m <= 12; m++ {
		c := NewCalendar(tree, cfg)
		c.SetYear(2024)
		c.SetMonth(m)
		c.SetSelected(1)
		setBounds(tree, c, 0, 0, 400, 300)
		buf := render.NewCommandBuffer()
		c.Draw(buf)
	}
}

// --- Card ---

func TestCardWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	c := NewCard(tree, cfg)
	c.SetTitle("Test Card")
	c.SetBordered(true)
	c.SetBgColor(uimath.ColorWhite)
	c.SetHeaderExtra(NewButton(tree, "extra", cfg))
	drawAndVerify(t, tree, c, false)
}

// --- Cascader ---

func TestCascaderWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	c := NewCascader(tree, cfg)
	c.SetOptions([]*CascaderOption{
		{Label: "Asia", Value: "asia", Children: []*CascaderOption{
			{Label: "China", Value: "china"},
			{Label: "Japan", Value: "japan"},
		}},
	})
	if len(c.Options()) != 1 {
		t.Error("expected 1 option")
	}
	c.SetSelected([]string{"asia", "china"})
	if len(c.Selected()) != 2 {
		t.Errorf("expected 2 selected")
	}
	c.SetOpen(true)
	if !c.IsOpen() {
		t.Error("expected open")
	}
	changed := false
	c.OnChange(func(path []string) { changed = true })
	_ = changed
	drawAndVerify(t, tree, c, false)
}

// --- Collapse ---

func TestCollapseWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	c := NewCollapse(tree, cfg)
	c.SetPanels([]CollapsePanel{
		{Value: "a", Header: "Panel A"},
		{Value: "b", Header: "Panel B"},
		{Value: "c", Header: "Panel C"},
	})
	c.SetBordered(true)
	c.Toggle("a")
	if !c.IsActive("a") {
		t.Error("panel 'a' should be active")
	}
	c.SetAccordion(true)
	c.Toggle("b")
	changed := false
	c.OnChange(func(keys []string) { changed = true })
	_ = changed
	drawAndVerify(t, tree, c, false)
}

func TestCollapseAccordionMode(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	c := NewCollapse(tree, cfg)
	c.SetPanels([]CollapsePanel{{Value: "a", Header: "A"}, {Value: "b", Header: "B"}})
	c.SetAccordion(true)
	c.Toggle("a")
	c.Toggle("b")
	// In accordion mode, only one should be active
}

// --- ColorPicker ---

func TestColorPickerWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	cp := NewColorPicker(tree, cfg)
	cp.SetValue(uimath.ColorRed)
	if cp.Value() != uimath.ColorRed {
		t.Error("expected red")
	}
	cp.SetPresets([]uimath.Color{uimath.ColorRed, uimath.ColorWhite})
	if !cp.IsOpen() {
		// might default to closed
	}
	changed := false
	cp.OnChange(func(c uimath.Color) { changed = true })
	cp.SelectColor(uimath.ColorWhite)
	if !changed {
		t.Error("expected change callback")
	}
	drawAndVerify(t, tree, cp, false)
}

// --- Comment ---

func TestCommentWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	c := NewComment(tree, CommentData{Author: "Alice", Content: "Hello!"}, cfg)
	if c.Data().Author != "Alice" {
		t.Error("expected author Alice")
	}
	c.SetData(CommentData{Author: "Bob", Content: "Hi!", Time: "2h ago"})
	reply := NewComment(tree, CommentData{Author: "Charlie", Content: "Hey!"}, cfg)
	c.AddReply(reply)
	if len(c.Replies()) != 1 {
		t.Error("expected 1 reply")
	}
	drawAndVerify(t, tree, c, false)
}

// --- ContextMenu ---

func TestContextMenuWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	cm := NewContextMenu(tree, cfg)
	cm.SetWidth(200)
	cm.AddItem(ContextMenuItem{Label: "Copy", OnClick: func() {}})
	cm.AddDivider()
	cm.AddItem(ContextMenuItem{Label: "Paste"})
	cm.AddItem(ContextMenuItem{Label: "Disabled", Disabled: true})
	if len(cm.Items()) != 4 {
		t.Errorf("expected 4 items, got %d", len(cm.Items()))
	}
	cm.Show(100, 200)
	if !cm.IsVisible() {
		t.Error("context menu should be visible")
	}
	setBounds(tree, cm, 0, 0, 200, 200)
	buf := render.NewCommandBuffer()
	cm.Draw(buf)

	cm.Hide()
	if cm.IsVisible() {
		t.Error("context menu should be hidden")
	}
	cm.ClearItems()
	if len(cm.Items()) != 0 {
		t.Error("expected 0 items after clear")
	}
}

// --- DatePicker ---

func TestDatePickerWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	dp := NewDatePicker(tree, cfg)
	dp.SetDate(2026, 3, 15)
	if dp.Year() != 2026 || dp.Month() != 3 || dp.Day() != 15 {
		t.Error("expected date 2026-3-15")
	}
	dp.SetOpen(true)
	if !dp.IsOpen() {
		t.Error("expected open")
	}
	changed := false
	dp.OnChange(func(y, m, d int) { changed = true })
	_ = changed
	drawAndVerify(t, tree, dp, false)
}

func TestDatePickerAllMonths(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	dp := NewDatePicker(tree, cfg)
	dp.SetOpen(true)
	setBounds(tree, dp, 0, 0, 300, 40)
	for m := 1; m <= 12; m++ {
		dp.SetDate(2024, m, 1)
		buf := render.NewCommandBuffer()
		dp.Draw(buf)
	}
}

func TestDatePickerClosed(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	dp := NewDatePicker(tree, cfg)
	dp.SetOpen(false)
	drawAndVerify(t, tree, dp, false)
}

// --- TimePicker ---

func TestTimePickerWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	tp := NewTimePicker(tree, cfg)
	tp.SetTime(14, 30, 45)
	if tp.Hour() != 14 || tp.Minute() != 30 || tp.Second() != 45 {
		t.Error("expected time 14:30:45")
	}
	tp.SetShowSeconds(true)
	tp.SetOpen(true)
	if !tp.IsOpen() {
		t.Error("expected open")
	}
	changed := false
	tp.OnChange(func(h, m, s int) { changed = true })
	_ = changed
	drawAndVerify(t, tree, tp, false)
}

// --- DateRangePicker ---

func TestDateRangePickerWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	drp := NewDateRangePicker(tree, cfg)
	drp.SetStartDate(2026, 1, 1)
	drp.SetEndDate(2026, 12, 31)
	sy, sm, sd := drp.StartDate()
	if sy != 2026 || sm != 1 || sd != 1 {
		t.Error("expected start date 2026-1-1")
	}
	ey, em, ed := drp.EndDate()
	if ey != 2026 || em != 12 || ed != 31 {
		t.Error("expected end date 2026-12-31")
	}
	drp.SetOpen(true)
	if !drp.IsOpen() {
		t.Error("expected open")
	}
	changed := false
	drp.OnChange(func(sy, sm, sd, ey, em, ed int) { changed = true })
	_ = changed
	drawAndVerify(t, tree, drp, false)
}

// --- Descriptions ---

func TestDescriptionsWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	d := NewDescriptions(tree, cfg)
	d.SetTitle("User Info")
	if d.Title() != "User Info" {
		t.Error("expected title")
	}
	d.SetColumns(3)
	d.SetBordered(true)
	d.AddItem(DescriptionItem{Label: "Name", Value: "Alice"})
	d.AddItem(DescriptionItem{Label: "Age", Value: "30"})
	if len(d.Items()) != 2 {
		t.Errorf("expected 2 items")
	}
	drawAndVerify(t, tree, d, false)

	d.ClearItems()
	if len(d.Items()) != 0 {
		t.Error("expected 0 items after clear")
	}
}

// --- Divider ---

func TestDividerWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	d := NewDivider(tree, cfg)
	d.SetLayout(DividerVertical)
	d.SetColor(uimath.ColorRed)
	d.SetThickness(2)
	d.SetContent("OR")
	drawAndVerify(t, tree, d, false)
}

func TestDividerHorizontal(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	d := NewDivider(tree, cfg)
	d.SetContent("---")
	drawAndVerify(t, tree, d, false)
}

// --- DockLayout ---

func TestDockLayoutFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	dl := NewDockLayout(tree, cfg)

	dl.AddPanel(&DockPanel{ID: "explorer", Title: "Explorer", Position: DockLeft})
	dl.AddPanel(&DockPanel{ID: "output", Title: "Output", Position: DockBottom})
	dl.AddPanel(&DockPanel{ID: "editor", Title: "Editor", Position: DockCenter})
	dl.AddPanel(&DockPanel{ID: "props", Title: "Properties", Position: DockRight})
	dl.AddPanel(&DockPanel{ID: "top", Title: "Top", Position: DockTop})

	if len(dl.Panels()) != 5 {
		t.Errorf("expected 5 panels, got %d", len(dl.Panels()))
	}

	p := dl.FindPanel("explorer")
	if p == nil || p.Title != "Explorer" {
		t.Error("expected to find explorer panel")
	}

	docked := false
	dl.OnDock(func(id string, pos DockPosition) { docked = true })
	undocked := false
	dl.OnUndock(func(id string) { undocked = true })

	dl.UndockPanel("props")
	p = dl.FindPanel("props")
	if p.Position != DockFloat {
		t.Error("expected float position after undock")
	}

	dl.DockPanel("props", DockRight)
	if p.Position != DockRight {
		t.Error("expected right position after dock")
	}

	dl.RemovePanel("output")
	if len(dl.Panels()) != 4 {
		t.Errorf("expected 4 panels after remove")
	}

	dl.SetActiveTab(0)
	if dl.ActiveTab() != 0 {
		t.Error("expected active tab 0")
	}

	_ = docked
	_ = undocked
	drawAndVerify(t, tree, dl, false)
}

func TestDockLayoutFloat(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	dl := NewDockLayout(tree, cfg)
	dl.AddPanel(&DockPanel{ID: "f1", Title: "Float", Position: DockFloat, FloatX: 100, FloatY: 100, FloatW: 200, FloatH: 150})
	drawAndVerify(t, tree, dl, false)
}

// --- DragDrop ---

func TestDragDropManager(t *testing.T) {
	tree := newTestTree()
	dd := NewDragDropManager(tree)

	if dd.IsDragging() {
		t.Error("should not be dragging initially")
	}
	if dd.DragData() != nil {
		t.Error("should have nil drag data")
	}

	buf := render.NewCommandBuffer()
	dd.Draw(buf) // not dragging, no output
}

func TestDragDropCompleteFlow(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	dd := NewDragDropManager(tree)

	src := NewButton(tree, "src", cfg)
	tgt := NewButton(tree, "tgt", cfg)
	tree.AppendChild(tree.Root(), src.ElementID())
	tree.AppendChild(tree.Root(), tgt.ElementID())
	tree.SetLayout(tree.Root(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 500, 500)})
	setBounds(tree, src, 10, 10, 80, 30)
	setBounds(tree, tgt, 200, 10, 80, 30)

	dd.RegisterSource(&DragSource{Widget: src, Data: "payload"})

	var enterCalled, leaveCalled, dropCalled bool
	var droppedData any
	dd.RegisterTarget(&DropTarget{
		Widget:  tgt,
		Accept:  func(data any) bool { return true },
		OnDrop:  func(data any) { dropCalled = true; droppedData = data },
		OnEnter: func(data any) { enterCalled = true },
		OnLeave: func() { leaveCalled = true },
	})

	// Trigger mousedown via tree handler (last handler is DragDrop's)
	handlers := tree.Handlers(src.ElementID(), event.MouseDown)
	if len(handlers) == 0 {
		t.Fatal("expected mousedown handler on source")
	}
	handlers[len(handlers)-1](&event.Event{Type: event.MouseDown, GlobalX: 50, GlobalY: 25})

	// Move past threshold
	dd.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: 60, GlobalY: 35})
	if !dd.IsDragging() {
		t.Fatal("expected dragging after threshold")
	}
	if dd.DragData() != "payload" {
		t.Errorf("expected payload, got %v", dd.DragData())
	}

	// Draw while dragging (default icon)
	buf := render.NewCommandBuffer()
	dd.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlay from drag indicator")
	}

	// Move over target
	dd.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: 240, GlobalY: 25})
	if !enterCalled {
		t.Error("expected OnEnter callback")
	}

	// Move away from target
	dd.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: 400, GlobalY: 400})
	if !leaveCalled {
		t.Error("expected OnLeave callback")
	}

	// Move back over target and drop
	dd.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: 240, GlobalY: 25})
	dd.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 240, GlobalY: 25})
	if !dropCalled {
		t.Error("expected OnDrop callback")
	}
	if droppedData != "payload" {
		t.Errorf("expected payload in drop, got %v", droppedData)
	}
	if dd.IsDragging() {
		t.Error("should not be dragging after drop")
	}
}

func TestDragDropNoAccept(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	dd := NewDragDropManager(tree)

	src := NewButton(tree, "s", cfg)
	tgt := NewButton(tree, "t", cfg)
	tree.AppendChild(tree.Root(), src.ElementID())
	tree.AppendChild(tree.Root(), tgt.ElementID())
	tree.SetLayout(tree.Root(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 500, 500)})
	setBounds(tree, src, 10, 10, 80, 30)
	setBounds(tree, tgt, 200, 10, 80, 30)

	dd.RegisterSource(&DragSource{Widget: src, Data: "x"})
	dropCalled := false
	dd.RegisterTarget(&DropTarget{
		Widget: tgt,
		Accept: func(data any) bool { return false },
		OnDrop: func(data any) { dropCalled = true },
	})

	handlers := tree.Handlers(src.ElementID(), event.MouseDown)
	handlers[len(handlers)-1](&event.Event{GlobalX: 50, GlobalY: 25})
	dd.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: 60, GlobalY: 35})
	dd.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: 240, GlobalY: 25})
	dd.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 240, GlobalY: 25})

	if dropCalled {
		t.Error("OnDrop should not be called when Accept returns false")
	}
}

func TestDragDropDrawWithIcon(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	dd := NewDragDropManager(tree)

	src := NewButton(tree, "s", cfg)
	icon := NewButton(tree, "icon", cfg)
	tree.AppendChild(tree.Root(), src.ElementID())
	tree.SetLayout(tree.Root(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 500, 500)})
	setBounds(tree, src, 10, 10, 80, 30)
	setBounds(tree, icon, 0, 0, 40, 40)

	dd.RegisterSource(&DragSource{Widget: src, Data: "d", DragIcon: icon})

	handlers := tree.Handlers(src.ElementID(), event.MouseDown)
	handlers[len(handlers)-1](&event.Event{GlobalX: 50, GlobalY: 25})
	dd.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: 60, GlobalY: 35})

	if !dd.IsDragging() {
		t.Fatal("expected dragging")
	}

	buf := render.NewCommandBuffer()
	dd.Draw(buf)
}

func TestDragDropNilSource(t *testing.T) {
	tree := newTestTree()
	dd := NewDragDropManager(tree)
	dd.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: 100, GlobalY: 100})
	dd.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 100, GlobalY: 100})
}

// --- Drawer ---

func TestDrawerWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	d := NewDrawer(tree, "Test Drawer", cfg)
	d.SetHeader("New Title")
	d.SetSize("400")
	d.SetSize("300")
	d.SetCloseBtn(false)
	closeCalled := false
	d.OnClose(func() { closeCalled = true })

	d.Open()
	if !d.IsVisible() {
		t.Error("drawer should be visible")
	}

	// Draw all placements
	for _, p := range []DrawerPlacement{DrawerLeft, DrawerRight, DrawerTop, DrawerBottom} {
		d.SetPlacement(p)
		setBounds(tree, d, 0, 0, 800, 600)
		buf := render.NewCommandBuffer()
		d.Draw(buf)
	}

	d.Close()
	if d.IsVisible() {
		t.Error("drawer should be hidden")
	}
	_ = closeCalled
}

// --- Gamepad ---

type mockNav struct {
	Base
	focusable bool
}

func (m *mockNav) NavFocusable() bool          { return m.focusable }
func (m *mockNav) Draw(buf *render.CommandBuffer) {}

func TestGamepadNavigatorFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	gn := NewGamepadNavigator(tree, cfg)

	w1 := &mockNav{Base: NewBase(tree, "custom", cfg), focusable: true}
	w2 := &mockNav{Base: NewBase(tree, "custom", cfg), focusable: true}
	w3 := &mockNav{Base: NewBase(tree, "custom", cfg), focusable: false}

	gn.AddWidget(w1)
	gn.AddWidget(w2)
	gn.AddWidget(w3)

	if len(gn.Widgets()) != 3 {
		t.Errorf("expected 3 widgets")
	}
	if gn.FocusIndex() != 0 {
		t.Errorf("expected initial focus 0")
	}

	gn.SetDeadzone(0.2)
	gn.SetRepeatDelay(0.5)
	gn.SetRepeatRate(0.1)
	gn.SetShowFocus(true)
	if !gn.ShowFocus() {
		t.Error("expected show focus")
	}

	gn.Navigate(NavDown)
	if gn.FocusIndex() != 1 {
		t.Errorf("expected focus 1 after nav down, got %d", gn.FocusIndex())
	}

	navigated := false
	gn.OnNavigate(func(w Navigable) { navigated = true })
	gn.Navigate(NavDown)
	_ = navigated

	gn.Navigate(NavUp)

	gn.Navigate(NavLeft)
	gn.Navigate(NavRight)

	fw := gn.FocusedWidget()
	_ = fw

	activated := false
	gn.OnActivate(func(w Navigable) { activated = true })
	gn.Activate()
	if !activated {
		t.Error("expected activate callback")
	}

	cancelled := false
	gn.OnCancel(func() { cancelled = true })
	gn.Cancel()
	if !cancelled {
		t.Error("expected cancel callback")
	}

	gn.SetFocus(0)
	gn.Tick(0.1)

	gn.SetEnabled(false)
	if gn.IsEnabled() {
		t.Error("expected disabled")
	}

	gn.SetEnabled(true)
	gn.RemoveWidget(w3)
	if len(gn.Widgets()) != 2 {
		t.Error("expected 2 widgets after remove")
	}
	gn.ClearWidgets()
	if len(gn.Widgets()) != 0 {
		t.Error("expected 0 widgets after clear")
	}

	setBounds(tree, gn, 0, 0, 400, 300)
	buf := render.NewCommandBuffer()
	gn.Draw(buf)
}

// --- Guide ---

func TestGuideWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	g := NewGuide(tree, cfg)
	g.SetSteps([]GuideStep{
		{Title: "Step 1", Description: "Click here"},
		{Title: "Step 2", Description: "Then here"},
		{Title: "Step 3", Description: "Done!"},
	})
	if len(g.Steps()) != 3 {
		t.Error("expected 3 steps")
	}
	g.SetMaskColor(uimath.RGBA(0, 0, 0, 0.5))
	changed := false
	g.OnChange(func(idx int) { changed = true })

	g.Start()
	if !g.IsVisible() {
		t.Error("should be visible after start")
	}
	if g.Current() != 0 {
		t.Errorf("expected current 0")
	}

	g.Next()
	if g.Current() != 1 {
		t.Errorf("expected current 1 after next")
	}

	g.Prev()
	if g.Current() != 0 {
		t.Errorf("expected current 0 after prev")
	}

	finished := false
	g.OnFinish(func() { finished = true })
	g.Next()
	g.Next()
	g.Next()
	if !finished {
		t.Error("expected finish callback")
	}
	if g.IsVisible() {
		t.Error("should be hidden after finish")
	}

	// Draw visible guide
	g.Start()
	setBounds(tree, g, 0, 0, 800, 600)
	buf := render.NewCommandBuffer()
	g.Draw(buf)

	_ = changed
}

func TestGuidePrevAtStart(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	g := NewGuide(tree, cfg)
	g.SetSteps([]GuideStep{{Title: "Only"}})
	g.Start()
	g.Prev()
	if g.Current() != 0 {
		t.Error("prev at start should stay at 0")
	}
}

func TestGuideFinish(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	g := NewGuide(tree, cfg)
	g.SetSteps([]GuideStep{{Title: "Only"}})
	g.Start()
	g.Finish()
	if g.IsVisible() {
		t.Error("should be hidden after finish")
	}
}

// --- ImageViewer ---

func TestImageViewerWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	iv := NewImageViewer(tree, cfg)
	iv.SetTexture(0)
	if iv.Texture() != 0 {
		t.Error("expected texture 0")
	}
	iv.SetZoom(2)
	if iv.Zoom() != 2 {
		t.Errorf("expected zoom 2")
	}
	iv.SetPan(10, 20)
	iv.ZoomIn()
	iv.ZoomOut()
	iv.ResetZoom()
	if iv.Zoom() != 1 {
		t.Error("expected zoom reset to 1")
	}
	iv.SetVisible(true)
	if !iv.IsVisible() {
		t.Error("expected visible")
	}
	closed := false
	iv.OnClose(func() { closed = true })
	_ = closed

	setBounds(tree, iv, 0, 0, 800, 600)
	buf := render.NewCommandBuffer()
	iv.Draw(buf)
}

func TestImageViewerDrawHidden(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	iv := NewImageViewer(tree, cfg)
	iv.SetVisible(false)
	setBounds(tree, iv, 0, 0, 800, 600)
	buf := render.NewCommandBuffer()
	iv.Draw(buf)
}

// --- InputNumber ---

func TestInputNumberWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	in := NewInputNumber(tree, cfg)
	in.SetValue(10)
	in.SetMin(0)
	in.SetMax(100)
	in.SetStep(5)
	if in.Value() != 10 {
		t.Errorf("expected 10")
	}
	if in.Min() != 0 || in.Max() != 100 || in.Step() != 5 {
		t.Error("expected min/max/step")
	}
	in.SetDisabled(true)
	changed := false
	in.OnChange(func(v float64) { changed = true })

	in.SetDisabled(false)
	in.Increment()
	if in.Value() != 15 {
		t.Errorf("expected 15, got %g", in.Value())
	}
	in.Decrement()
	if in.Value() != 10 {
		t.Errorf("expected 10, got %g", in.Value())
	}

	// Clamp at max
	in.SetValue(200)
	if in.Value() != 100 {
		t.Errorf("expected clamped to 100, got %g", in.Value())
	}

	// Clamp at min
	in.SetValue(-5)
	if in.Value() != 0 {
		t.Errorf("expected clamped to 0, got %g", in.Value())
	}

	drawAndVerify(t, tree, in, false)
	_ = changed
}

// --- Link ---

func TestLinkWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	l := NewLink(tree, "Click me", "https://example.com", cfg)
	if l.Text() != "Click me" {
		t.Errorf("expected text 'Click me'")
	}
	if l.Href() != "https://example.com" {
		t.Errorf("expected href")
	}
	l.SetText("New")
	l.SetHref("/new")
	l.SetDisabled(true)
	clicked := false
	l.OnClick(func(url string) { clicked = true })
	_ = clicked
	drawAndVerify(t, tree, l, false)
}

func TestLinkOnClick(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	l := NewLink(tree, "click", "https://example.com", cfg)
	var gotURL string
	l.OnClick(func(url string) { gotURL = url })
	_ = gotURL
}

// --- List ---

func TestListWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	l := NewList(tree, cfg)
	l.SetItems([]ListItem{
		{Title: "Item 1"},
		{Title: "Item 2"},
		{Title: "Item 3"},
	})
	if len(l.Items()) != 3 {
		t.Errorf("expected 3 items")
	}
	l.SetItemHeight(40)
	l.SetBordered(true)
	l.SetScrollY(50)
	if l.ScrollY() != 50 {
		t.Error("expected scrollY 50")
	}
	selected := false
	l.OnSelect(func(i int) { selected = true })
	_ = selected
	drawAndVerify(t, tree, l, false)
}

func TestVirtualListWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	vl := NewVirtualList(tree, cfg)
	vl.SetItemCount(1000)
	vl.SetItemHeight(32)
	vl.SetScrollY(500)
	if vl.ScrollY() != 500 {
		t.Error("expected scrollY 500")
	}
	ch := vl.ContentHeight()
	if ch <= 0 {
		t.Error("expected positive content height")
	}
	vl.SetRenderItem(func(index int, buf *render.CommandBuffer, x, y, w, h float32) {
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x, y, w, h),
			FillColor: uimath.ColorRed,
		}, 0, 1)
	})
	drawAndVerify(t, tree, vl, false)
}

// --- Menu ---

func TestMenuWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	m := NewMenu(tree, cfg)
	m.SetItems([]MenuItem{
		{Value: "a", Content: "Item A", Children: []MenuItem{{Value: "a1", Content: "Sub A"}}},
		{Value: "b", Content: "Item B"},
		{Value: "c", Content: "Item C", Disabled: true},
	})
	m.SetValue("a")
	if m.Value() != "a" {
		t.Error("expected selected 'a'")
	}
	selected := ""
	m.OnChange(func(key string) { selected = key })
	m.SelectItem("b")
	if selected != "b" {
		t.Errorf("expected 'b', got %q", selected)
	}
	m.ToggleExpanded("a")
	m.ToggleExpanded("a") // toggle back
	m.SelectItem("c") // disabled
	drawAndVerify(t, tree, m, false)
}

// --- MenuBar ---

func TestMenuBarWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	mb := NewMenuBar(tree, cfg)
	mb.SetHeight(40)
	mb.AddItem(MenuBarItem{Label: "File"})
	mb.AddItem(MenuBarItem{Label: "Edit"})
	if len(mb.Items()) != 2 {
		t.Errorf("expected 2 items")
	}
	if mb.IsOpen() {
		t.Error("should not be open initially")
	}
	drawAndVerify(t, tree, mb, false)
	mb.ClearItems()
	if len(mb.Items()) != 0 {
		t.Error("expected 0 items after clear")
	}
}

// --- MessageBox ---

func TestMessageBoxWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	mb := NewMessageBox(tree, "Title", "Content", cfg)
	if mb.Title() != "Title" || mb.Content() != "Content" {
		t.Error("expected title/content")
	}
	mb.SetTitle("New Title")
	mb.SetContent("New Content")
	mb.SetBoxType(MessageBoxConfirm)
	if mb.BoxType() != MessageBoxConfirm {
		t.Error("expected confirm type")
	}
	mb.SetWidth(300)
	mb.SetShowCancel(true)
	okCalled := false
	mb.OnOK(func() { okCalled = true })
	cancelCalled := false
	mb.OnCancel(func() { cancelCalled = true })

	mb.Open()
	if !mb.IsVisible() {
		t.Error("expected visible after open")
	}
	setBounds(tree, mb, 0, 0, 800, 600)
	buf := render.NewCommandBuffer()
	mb.Draw(buf)

	mb.Close()
	if mb.IsVisible() {
		t.Error("expected hidden after close")
	}

	_ = okCalled
	_ = cancelCalled

	// Test all box types
	for _, bt := range []MessageBoxType{MessageBoxInfo, MessageBoxSuccess, MessageBoxWarning, MessageBoxError, MessageBoxConfirm} {
		mb2 := NewMessageBox(tree, "T", "C", cfg)
		mb2.SetBoxType(bt)
		mb2.Open()
		setBounds(tree, mb2, 0, 0, 800, 600)
		buf2 := render.NewCommandBuffer()
		mb2.Draw(buf2)
	}
}

// --- Notification ---

func TestNotificationWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	n := NewNotification(tree, "Title", "Message", cfg)
	n.SetTitle("New Title")
	n.SetContent("New Message")
	if n.Title() != "New Title" || n.Content() != "New Message" {
		t.Error("expected new title/message")
	}
	n.SetTheme(NotificationThemeSuccess)
	n.SetPosition(10, 10)
	closeCalled := false
	n.OnClose(func() { closeCalled = true })

	n.Show()
	if !n.IsVisible() {
		t.Error("notification should be visible")
	}
	setBounds(tree, n, 0, 0, 300, 100)
	buf := render.NewCommandBuffer()
	n.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlay commands from Notification")
	}

	n.Close()
	if n.IsVisible() {
		t.Error("notification should be hidden after close")
	}
	_ = closeCalled
}

// --- Pagination ---

func TestPaginationWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	p := NewPagination(tree, cfg)
	p.SetTotal(100)
	p.SetPageSize(10)
	if p.Total() != 100 || p.PageSize() != 10 {
		t.Error("expected total/pageSize")
	}
	if p.TotalPages() != 10 {
		t.Errorf("expected 10 pages, got %d", p.TotalPages())
	}
	p.SetCurrent(5)
	if p.Current() != 5 {
		t.Errorf("expected current 5")
	}
	changed := false
	p.OnChange(func(pageInfo PaginationPageInfo) { changed = true })
	p.GoTo(3)
	if !changed {
		t.Error("expected change callback")
	}
	drawAndVerify(t, tree, p, false)
}

func TestPaginationGoToOutOfRange(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	p := NewPagination(tree, cfg)
	p.SetTotal(50)
	p.SetPageSize(10)
	p.GoTo(0)  // below 1
	p.GoTo(99) // above total pages
}

func TestPaginationZeroTotal(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	p := NewPagination(tree, cfg)
	p.SetTotal(0)
	p.SetPageSize(10)
	if p.TotalPages() < 0 {
		t.Error("should not be negative")
	}
}

// --- Panel ---

func TestPanelWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	p := NewPanel(tree, "My Panel", cfg)
	p.SetTitle("New Title")
	p.SetBgColor(uimath.ColorRed)
	p.SetBordered(true)
	drawAndVerify(t, tree, p, false)
}

// --- Popconfirm ---

func TestPopconfirmWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	p := NewPopconfirm(tree, "Are you sure?", cfg)
	if p.Title() != "Are you sure?" {
		t.Error("expected title")
	}
	p.SetTitle("Really?")
	p.SetPlacement(PlacementTop)
	p.SetAnchorRect(100, 100, 50, 30)
	p.SetVisible(true)

	p.Show()
	if !p.IsVisible() {
		t.Error("expected visible")
	}

	confirmed := false
	p.OnConfirm(func() { confirmed = true })
	p.Confirm()
	if !confirmed {
		t.Error("expected confirm callback")
	}
	if p.IsVisible() {
		t.Error("expected hidden after confirm")
	}

	p.Show()
	cancelled := false
	p.OnCancel(func() { cancelled = true })
	p.Cancel()
	if !cancelled {
		t.Error("expected cancel callback")
	}

	// Draw
	p.Show()
	setBounds(tree, p, 0, 0, 400, 300)
	buf := render.NewCommandBuffer()
	p.Draw(buf)
}

// --- Popover ---

func TestPopoverWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	p := NewPopover(tree, cfg)
	p.SetTitle("Popover Title")
	p.SetContent(NewButton(tree, "test", cfg))
	p.SetPlacement(PlacementBottom)
	p.SetTrigger(PopoverTriggerClick)
	p.SetWidth(200)
	p.SetAnchorRect(100, 100, 50, 30)
	closeCalled := false
	p.OnClose(func() { closeCalled = true })
	p.SetVisible(true)

	p.Open()
	if !p.IsVisible() {
		t.Error("popover should be visible")
	}
	setBounds(tree, p, 0, 0, 400, 300)
	buf := render.NewCommandBuffer()
	p.Draw(buf)

	p.Close()
	if p.IsVisible() {
		t.Error("popover should be hidden")
	}
	_ = closeCalled
}

// --- Portal ---

func TestPortalWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	p := NewPortal(tree, cfg)
	btn := NewButton(tree, "test", cfg)
	p.SetContent(btn)
	if p.Content() == nil {
		t.Error("portal should have content")
	}
	p.SetVisible(true)
	if !p.IsVisible() {
		t.Error("expected visible")
	}
	p.SetZBase(10)
	setBounds(tree, p, 0, 0, 400, 300)
	setBounds(tree, btn, 0, 0, 100, 30)
	buf := render.NewCommandBuffer()
	p.Draw(buf)

	p.SetVisible(false)
	buf = render.NewCommandBuffer()
	p.Draw(buf)
}

// --- RangeInput ---

func TestRangeInputWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	ri := NewRangeInput(tree, 0, 100, cfg)
	ri.SetRange(20, 80)
	if ri.Low() != 20 || ri.High() != 80 {
		t.Errorf("expected range 20-80")
	}
	ri.SetStep(5)
	changed := false
	ri.OnChange(func(lo, hi float32) { changed = true })
	_ = changed
	drawAndVerify(t, tree, ri, false)
}

// --- Rate ---

func TestRateWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	r := NewRate(tree, cfg)
	r.SetValue(3)
	if r.Value() != 3 {
		t.Errorf("expected value 3")
	}
	r.SetCount(10)
	r.SetSize(24)
	r.SetDisabled(true)
	changed := false
	r.OnChange(func(v float32) { changed = true })
	_ = changed
	drawAndVerify(t, tree, r, false)
}

// --- RichText ---

func TestRichTextFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	rt := NewRichText(tree, cfg)
	rt.AddText("Hello ")
	rt.AddStyledText("World", uimath.ColorRed, 20, true)
	rt.AddBreak()
	rt.AddImage(0, 16, 16)
	rt.AddText(" after image")
	if len(rt.Spans()) != 5 {
		t.Errorf("expected 5 spans, got %d", len(rt.Spans()))
	}
	rt.AddSpan(RichSpan{Type: RichSpanText, Text: "Extra", FontSize: 16, Color: uimath.ColorRed})
	rt.SetLineSpacing(1.5)
	drawAndVerify(t, tree, rt, false)

	rt.ClearSpans()
	if len(rt.Spans()) != 0 {
		t.Error("expected 0 spans after clear")
	}
	rt.SetSpans([]RichSpan{{Type: RichSpanText, Text: "test"}})
	if len(rt.Spans()) != 1 {
		t.Error("expected 1 span after SetSpans")
	}
}

// --- Skeleton ---

func TestSkeletonWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	s := NewSkeleton(tree, cfg)
	s.SetRows(5)
	s.SetAvatar(true)
	s.SetLoading(true)
	if s.Rows() != 5 {
		t.Errorf("expected 5 rows")
	}
	drawAndVerify(t, tree, s, false)
}

// --- Slider ---

func TestSliderWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	s := NewSlider(tree, cfg)
	s.SetMin(0)
	s.SetMax(100)
	s.SetStep(5)
	s.SetValue(50)
	if s.Value() != 50 {
		t.Errorf("expected 50, got %g", s.Value())
	}
	if s.Min() != 0 || s.Max() != 100 {
		t.Error("expected min/max")
	}
	s.SetDisabled(true)
	changed := false
	s.OnChange(func(v float32) { changed = true })
	_ = changed

	// Clamp beyond max
	s.SetValue(200)
	if s.Value() > 100 {
		t.Error("expected clamped")
	}
	s.SetValue(-5)
	if s.Value() < 0 {
		t.Error("expected clamped")
	}

	drawAndVerify(t, tree, s, false)
}

// --- Splitter ---

func TestSplitterWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	s := NewSplitter(tree, cfg)
	s.SetRatio(0.3)
	if s.Ratio() != 0.3 {
		t.Errorf("expected ratio 0.3, got %g", s.Ratio())
	}
	s.SetMinRatio(0.1)
	s.SetMaxRatio(0.9)
	s.SetDirection(SplitterVertical)
	s.SetFirst(NewButton(tree, "first", cfg))
	s.SetSecond(NewButton(tree, "second", cfg))
	drawAndVerify(t, tree, s, false)
}

func TestSplitterHorizontal(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	s := NewSplitter(tree, cfg)
	s.SetDirection(SplitterHorizontal)
	s.SetFirst(NewButton(tree, "a", cfg))
	s.SetSecond(NewButton(tree, "b", cfg))
	drawAndVerify(t, tree, s, false)
}

// --- Statistic ---

func TestStatisticWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	s := NewStatistic(tree, "Users", "1,234", cfg)
	if s.Title() != "Users" || s.Value() != "1,234" {
		t.Error("expected title/value")
	}
	s.SetTitle("Views")
	s.SetValue("2,000")
	s.SetPrefix("$")
	s.SetSuffix("/mo")
	s.SetColor(uimath.ColorRed)
	// Draw requires TextRenderer for output, just verify no panic
	setBounds(tree, s, 0, 0, 200, 80)
	buf := render.NewCommandBuffer()
	s.Draw(buf)
}

// --- Steps ---

func TestStepsWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	s := NewSteps(tree, cfg)
	s.AddStep(StepItem{Title: "Step 1", Content: "First"})
	s.AddStep(StepItem{Title: "Step 2", Content: "Second"})
	s.AddStep(StepItem{Title: "Step 3", Content: "Third"})
	if len(s.Options()) != 3 {
		t.Error("expected 3 items")
	}
	s.SetCurrent(1)
	if s.Current() != 1 {
		t.Errorf("expected current 1")
	}
	drawAndVerify(t, tree, s, false)

	s.ClearSteps()
	if len(s.Options()) != 0 {
		t.Error("expected 0 items after clear")
	}
}

// --- SubWindow ---

func TestSubWindowWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	w := NewSubWindow(tree, "Test Window", cfg)
	w.SetPosition(100, 100)
	w.SetSize(400, 300)
	w.SetTitle("New Title")
	w.SetClosable(true)
	closeCalled := false
	w.OnClose(func() { closeCalled = true })

	if !w.IsVisible() {
		t.Error("window should be visible")
	}
	w.Open()

	setBounds(tree, w, 0, 0, 800, 600)
	buf := render.NewCommandBuffer()
	w.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlay commands from SubWindow")
	}

	w.Close()
	if w.IsVisible() {
		t.Error("window should be hidden after close")
	}
	_ = closeCalled
}

// --- Swiper ---

func TestSwiperWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	s := NewSwiper(tree, cfg)
	s.AddPanel(NewButton(tree, "slide1", cfg))
	s.AddPanel(NewButton(tree, "slide2", cfg))
	s.AddPanel(NewButton(tree, "slide3", cfg))
	if s.PanelCount() != 3 {
		t.Errorf("expected 3 panels")
	}
	s.SetAutoplay(true)
	s.SetShowDots(true)
	var lastIdx int
	s.OnChange(func(idx int) { lastIdx = idx })

	s.Next()
	if s.Current() != 1 || lastIdx != 1 {
		t.Errorf("expected current 1 after next")
	}
	s.Prev()
	if s.Current() != 0 || lastIdx != 0 {
		t.Errorf("expected current 0 after prev")
	}
	s.SetCurrent(2)
	drawAndVerify(t, tree, s, false)
}

func TestSwiperNextPrevWrap(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	s := NewSwiper(tree, cfg)
	s.AddPanel(NewDiv(tree, cfg))
	s.AddPanel(NewDiv(tree, cfg))
	s.AddPanel(NewDiv(tree, cfg))

	s.Next()
	s.Next()
	s.Next() // wraps to 0
	if s.Current() != 0 {
		t.Errorf("expected 0 after wrap, got %d", s.Current())
	}
	s.Prev() // wraps to 2
	if s.Current() != 2 {
		t.Errorf("expected 2 after prev wrap, got %d", s.Current())
	}
}

func TestSwiperEmptyNextPrev(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	s := NewSwiper(tree, cfg)
	s.Next() // no panic
	s.Prev() // no panic
}

// --- Table ---

func TestTableWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	tbl := NewTable(tree, []TableColumn{
		{Title: "Name", Width: 100},
		{Title: "Age", Width: 80},
	}, cfg)
	if len(tbl.Columns()) != 2 {
		t.Error("expected 2 columns")
	}
	tbl.AddRow([]string{"Alice", "30"})
	tbl.AddRow([]string{"Bob", "25"})
	if tbl.RowCount() != 2 {
		t.Errorf("expected 2 rows")
	}
	if len(tbl.Rows()) != 2 {
		t.Error("expected rows")
	}
	tbl.SetStripe(true)
	tbl.SetBordered(true)
	tbl.SetRowHeight(40)
	tbl.SetRows([][]string{{"C", "1"}, {"D", "2"}})
	drawAndVerify(t, tree, tbl, false)

	tbl.ClearRows()
	if tbl.RowCount() != 0 {
		t.Errorf("expected 0 rows after clear")
	}
}

// --- TagInput ---

func TestTagInputWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	ti := NewTagInput(tree, cfg)
	ti.SetMaxTags(5)
	addCalled := false
	removeCalled := false
	ti.OnAdd(func(tag string) { addCalled = true })
	ti.OnRemove(func(tag string, idx int) { removeCalled = true })

	ti.AddTag("Go")
	ti.AddTag("Rust")
	ti.AddTag("Python")
	if len(ti.Tags()) != 3 {
		t.Errorf("expected 3 tags, got %d", len(ti.Tags()))
	}
	if !addCalled {
		t.Error("expected add callback")
	}

	ti.RemoveTag(1)
	if len(ti.Tags()) != 2 {
		t.Errorf("expected 2 tags after remove")
	}

	ti.ClearTags()
	if len(ti.Tags()) != 0 {
		t.Error("expected 0 tags after clear")
	}
	_ = removeCalled
	drawAndVerify(t, tree, ti, false)
}

func TestTagInputMaxTags(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	ti := NewTagInput(tree, cfg)
	ti.SetMaxTags(2)
	ti.AddTag("A")
	ti.AddTag("B")
	ti.AddTag("C") // should be rejected
	if len(ti.Tags()) != 2 {
		t.Errorf("expected 2 tags (max), got %d", len(ti.Tags()))
	}
}

// --- Timeline ---

func TestTimelineWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	tl := NewTimeline(tree, cfg)
	tl.SetItemHeight(60)
	tl.AddItem(TimelineItem{Label: "Created", DotColor: "primary"})
	tl.AddItem(TimelineItem{Label: "Processing", DotColor: "default"})
	tl.AddItem(TimelineItem{Label: "Error", DotColor: "error"})
	if len(tl.Items()) != 3 {
		t.Errorf("expected 3 items")
	}
	drawAndVerify(t, tree, tl, false)

	tl.ClearItems()
	if len(tl.Items()) != 0 {
		t.Error("expected 0 items after clear")
	}
}

// --- Transfer ---

func TestTransferWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	tr := NewTransfer(tree, cfg)
	tr.SetSource([]TransferItem{
		{Key: "1", Label: "Item 1"},
		{Key: "2", Label: "Item 2"},
		{Key: "3", Label: "Item 3"},
	})
	var gotKeys []string
	tr.OnChange(func(keys []string) { gotKeys = keys })

	tr.MoveToTarget([]string{"1", "3"})
	if len(tr.Target()) != 2 {
		t.Errorf("expected 2 target items, got %d", len(tr.Target()))
	}
	if len(tr.Source()) != 1 {
		t.Errorf("expected 1 source item, got %d", len(tr.Source()))
	}
	tr.MoveToSource([]string{"1"})
	if len(tr.Source()) != 2 {
		t.Errorf("expected 2 source items after move back")
	}
	_ = gotKeys
	drawAndVerify(t, tree, tr, false)
}

// --- TreeSelect ---

func TestTreeSelectWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	ts := NewTreeSelect(tree, cfg)
	ts.SetRoots([]*TreeNode{
		{Key: "a", Label: "A", Children: []*TreeNode{
			{Key: "a1", Label: "A1"},
		}},
		{Key: "b", Label: "B"},
	})
	ts.SetSelected("a1")
	if ts.Selected() != "a1" {
		t.Errorf("expected selected 'a1'")
	}
	ts.SetOpen(true)
	if !ts.IsOpen() {
		t.Error("expected open")
	}
	changed := false
	ts.OnChange(func(key string) { changed = true })
	_ = changed

	setBounds(tree, ts, 0, 0, 300, 40)
	buf := render.NewCommandBuffer()
	ts.Draw(buf)
}

// --- Tree widget ---

func TestTreeWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	tw := NewTree(tree, cfg)
	tw.SetIndent(24)
	tw.SetItemHeight(32)

	tw.AddRoot(&TreeNode{
		Key:   "root",
		Label: "Root",
		Children: []*TreeNode{
			{Key: "child1", Label: "Child 1"},
			{Key: "child2", Label: "Child 2"},
		},
	})
	if len(tw.Roots()) != 1 {
		t.Error("expected 1 root")
	}
	found := tw.FindNode("child1")
	if found == nil || found.Label != "Child 1" {
		t.Error("expected to find child1")
	}

	selected := false
	tw.OnSelect(func(node *TreeNode) { selected = true })
	expanded := false
	tw.OnExpand(func(node *TreeNode) { expanded = true })

	tw.ExpandAll()
	tw.CollapseAll()

	tw.SetRoots([]*TreeNode{
		{Key: "a", Label: "A"},
		{Key: "b", Label: "B", Children: []*TreeNode{{Key: "b1", Label: "B1"}}},
	})
	tw.ExpandAll()

	_ = selected
	_ = expanded
	drawAndVerify(t, tree, tw, false)
}

func TestTreeFindNodeNotFound(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	tw := NewTree(tree, cfg)
	tw.AddRoot(&TreeNode{Key: "a", Label: "A"})
	if tw.FindNode("nonexistent") != nil {
		t.Error("expected nil for nonexistent node")
	}
}

// --- Upload ---

func TestUploadWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	u := NewUpload(tree, cfg)
	u.SetMultiple(true)
	u.SetAccept(".png,.jpg")
	u.SetDrag(true)
	u.SetMaxCount(5)
	uploaded := false
	u.OnUpload(func(files []UploadFile) { uploaded = true })

	u.AddFile(UploadFile{Name: "test.txt", Size: 1024, Status: "done"})
	u.AddFile(UploadFile{Name: "test2.png", Size: 2048, Status: "uploading"})
	if len(u.Files()) != 2 {
		t.Errorf("expected 2 files")
	}
	u.RemoveFile(0)
	if len(u.Files()) != 1 {
		t.Error("expected 1 file after remove")
	}
	u.ClearFiles()
	if len(u.Files()) != 0 {
		t.Error("expected 0 files after clear")
	}
	_ = uploaded
	drawAndVerify(t, tree, u, false)
}

func TestUploadMaxCount(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	u := NewUpload(tree, cfg)
	u.SetMaxCount(2)
	u.AddFile(UploadFile{Name: "a.txt", Size: 1, Status: "done"})
	u.AddFile(UploadFile{Name: "b.txt", Size: 1, Status: "done"})
	u.AddFile(UploadFile{Name: "c.txt", Size: 1, Status: "done"}) // should be rejected
	if len(u.Files()) != 2 {
		t.Errorf("expected 2 files (max), got %d", len(u.Files()))
	}
}

// --- VirtualGrid ---

func TestVirtualGridWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	vg := NewVirtualGrid(tree, 4, 100, cfg)
	if vg.Cols() != 4 || vg.RowCount() != 100 {
		t.Error("expected 4 cols, 100 rows")
	}
	vg.SetCols(3)
	vg.SetRowCount(50)
	vg.SetCellSize(100, 100)
	vg.SetGap(8)
	vg.SetScrollY(50)
	if vg.ScrollY() != 50 {
		t.Error("expected scrollY 50")
	}
	total := vg.TotalHeight()
	if total <= 0 {
		t.Error("expected positive total height")
	}
	vg.SetRenderCell(func(buf *render.CommandBuffer, row, col int, rect uimath.Rect) {
		buf.DrawRect(render.RectCmd{
			Bounds:    rect,
			FillColor: uimath.ColorRed,
		}, 0, 1)
	})
	drawAndVerify(t, tree, vg, false)
}

// --- Watermark ---

func TestWatermarkWidgetFull(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	w := NewWatermark(tree, "DRAFT", cfg)
	w.SetText("CONFIDENTIAL")
	w.SetColor(uimath.ColorRed)
	w.SetAlpha(0.5)
	w.SetX(100)
	w.SetY(80)
	// Draw requires TextRenderer; just verify no panic
	setBounds(tree, w, 0, 0, 400, 300)
	buf := render.NewCommandBuffer()
	w.Draw(buf)
}

// === Handler coverage tests ===

// --- Slider mouse handlers ---

func TestSliderMouseHandlers(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	s := NewSlider(tree, cfg)
	s.SetMin(0)
	s.SetMax(100)
	setBounds(tree, s, 0, 0, 200, 30)

	changed := false
	s.OnChange(func(v float32) { changed = true })

	// Trigger mouse down
	downH := tree.Handlers(s.ElementID(), event.MouseDown)
	if len(downH) == 0 {
		t.Fatal("expected mousedown handler")
	}
	downH[len(downH)-1](&event.Event{Type: event.MouseDown, GlobalX: 100, GlobalY: 15})

	// Trigger mouse move
	moveH := tree.Handlers(s.ElementID(), event.MouseMove)
	moveH[len(moveH)-1](&event.Event{Type: event.MouseMove, GlobalX: 150, GlobalY: 15})
	if !changed {
		t.Error("expected change callback from slider drag")
	}

	// Trigger mouse up
	upH := tree.Handlers(s.ElementID(), event.MouseUp)
	upH[len(upH)-1](&event.Event{Type: event.MouseUp, GlobalX: 150, GlobalY: 15})
}

func TestSliderDisabledMouseDown(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	s := NewSlider(tree, cfg)
	s.SetDisabled(true)
	setBounds(tree, s, 0, 0, 200, 30)

	downH := tree.Handlers(s.ElementID(), event.MouseDown)
	downH[len(downH)-1](&event.Event{Type: event.MouseDown, GlobalX: 100, GlobalY: 15})
	// Should not start dragging when disabled
}

// --- RangeInput mouse handlers ---

func TestRangeInputMouseHandlers(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	ri := NewRangeInput(tree, 0, 100, cfg)
	ri.SetRange(20, 80)
	setBounds(tree, ri, 0, 0, 200, 30)

	changed := false
	ri.OnChange(func(lo, hi float32) { changed = true })

	// Click near low thumb
	downH := tree.Handlers(ri.ElementID(), event.MouseDown)
	downH[len(downH)-1](&event.Event{Type: event.MouseDown, GlobalX: 40, GlobalY: 15})

	// Drag
	moveH := tree.Handlers(ri.ElementID(), event.MouseMove)
	moveH[len(moveH)-1](&event.Event{Type: event.MouseMove, GlobalX: 60, GlobalY: 15})
	if !changed {
		t.Error("expected change callback")
	}

	// Release
	upH := tree.Handlers(ri.ElementID(), event.MouseUp)
	upH[len(upH)-1](&event.Event{Type: event.MouseUp})

	// Click near high thumb
	downH[len(downH)-1](&event.Event{Type: event.MouseDown, GlobalX: 160, GlobalY: 15})
	moveH[len(moveH)-1](&event.Event{Type: event.MouseMove, GlobalX: 180, GlobalY: 15})
	upH[len(upH)-1](&event.Event{Type: event.MouseUp})
}

func TestRangeInputMouseMoveNotDragging(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	ri := NewRangeInput(tree, 0, 100, cfg)
	setBounds(tree, ri, 0, 0, 200, 30)
	moveH := tree.Handlers(ri.ElementID(), event.MouseMove)
	moveH[len(moveH)-1](&event.Event{Type: event.MouseMove, GlobalX: 100, GlobalY: 15})
	// Should do nothing when not dragging
}

// --- Splitter mouse handlers ---

func TestSplitterMouseHandlers(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	s := NewSplitter(tree, cfg)
	s.SetDirection(SplitterHorizontal)
	setBounds(tree, s, 0, 0, 400, 300)

	downH := tree.Handlers(s.ElementID(), event.MouseDown)
	downH[len(downH)-1](&event.Event{Type: event.MouseDown, GlobalX: 200, GlobalY: 150})

	moveH := tree.Handlers(s.ElementID(), event.MouseMove)
	moveH[len(moveH)-1](&event.Event{Type: event.MouseMove, GlobalX: 250, GlobalY: 150})

	upH := tree.Handlers(s.ElementID(), event.MouseUp)
	upH[len(upH)-1](&event.Event{Type: event.MouseUp})
}

func TestSplitterVerticalMouseHandlers(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	s := NewSplitter(tree, cfg)
	s.SetDirection(SplitterVertical)
	setBounds(tree, s, 0, 0, 400, 300)

	downH := tree.Handlers(s.ElementID(), event.MouseDown)
	downH[len(downH)-1](&event.Event{Type: event.MouseDown, GlobalX: 200, GlobalY: 150})

	moveH := tree.Handlers(s.ElementID(), event.MouseMove)
	moveH[len(moveH)-1](&event.Event{Type: event.MouseMove, GlobalX: 200, GlobalY: 200})

	upH := tree.Handlers(s.ElementID(), event.MouseUp)
	upH[len(upH)-1](&event.Event{Type: event.MouseUp})
}

// --- DockLayout mouse handlers ---

func TestDockLayoutMouseHandlers(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	dl := NewDockLayout(tree, cfg)
	dl.AddPanel(&DockPanel{ID: "left", Title: "Left", Position: DockLeft, DockSize: 200, Visible: true})
	dl.AddPanel(&DockPanel{ID: "center", Title: "Center", Position: DockCenter, Visible: true})
	setBounds(tree, dl, 0, 0, 800, 600)

	// Trigger mouse down on splitter area (right edge of left panel)
	downH := tree.Handlers(dl.ElementID(), event.MouseDown)
	downH[len(downH)-1](&event.Event{Type: event.MouseDown, GlobalX: 201, GlobalY: 300})

	// Mouse move
	moveH := tree.Handlers(dl.ElementID(), event.MouseMove)
	moveH[len(moveH)-1](&event.Event{Type: event.MouseMove, GlobalX: 250, GlobalY: 300})

	// Mouse up
	upH := tree.Handlers(dl.ElementID(), event.MouseUp)
	upH[len(upH)-1](&event.Event{Type: event.MouseUp})
}

// --- Gamepad handlers ---

func TestGamepadKeyDownHandler(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	gn := NewGamepadNavigator(tree, cfg)
	w1 := &mockNav{Base: NewBase(tree, "custom", cfg), focusable: true}
	w2 := &mockNav{Base: NewBase(tree, "custom", cfg), focusable: true}
	gn.AddWidget(w1)
	gn.AddWidget(w2)

	keyH := tree.Handlers(gn.ElementID(), event.KeyDown)
	if len(keyH) == 0 {
		t.Fatal("expected keydown handler")
	}

	// Arrow down
	keyH[len(keyH)-1](&event.Event{Type: event.KeyDown, Key: event.KeyArrowDown})
	if gn.FocusIndex() != 1 {
		t.Errorf("expected focus 1 after arrow down, got %d", gn.FocusIndex())
	}

	// Arrow up
	keyH[len(keyH)-1](&event.Event{Type: event.KeyDown, Key: event.KeyArrowUp})

	// Arrow left/right
	keyH[len(keyH)-1](&event.Event{Type: event.KeyDown, Key: event.KeyArrowLeft})
	keyH[len(keyH)-1](&event.Event{Type: event.KeyDown, Key: event.KeyArrowRight})

	// Enter (activate)
	activated := false
	gn.OnActivate(func(w Navigable) { activated = true })
	keyH[len(keyH)-1](&event.Event{Type: event.KeyDown, Key: event.KeyEnter})
	if !activated {
		t.Error("expected activate from Enter")
	}

	// Space (activate)
	keyH[len(keyH)-1](&event.Event{Type: event.KeyDown, Key: event.KeySpace})

	// Escape (cancel)
	cancelled := false
	gn.OnCancel(func() { cancelled = true })
	keyH[len(keyH)-1](&event.Event{Type: event.KeyDown, Key: event.KeyEscape})
	if !cancelled {
		t.Error("expected cancel from Escape")
	}
}

func TestGamepadButtonHandler(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	gn := NewGamepadNavigator(tree, cfg)
	w1 := &mockNav{Base: NewBase(tree, "custom", cfg), focusable: true}
	w2 := &mockNav{Base: NewBase(tree, "custom", cfg), focusable: true}
	gn.AddWidget(w1)
	gn.AddWidget(w2)

	btnH := tree.Handlers(gn.ElementID(), event.GamepadButtonDown)
	if len(btnH) == 0 {
		t.Fatal("expected gamepad button handler")
	}

	btnH[len(btnH)-1](&event.Event{Type: event.GamepadButtonDown, GamepadButton: GPBtnDown})
	btnH[len(btnH)-1](&event.Event{Type: event.GamepadButtonDown, GamepadButton: GPBtnUp})
	btnH[len(btnH)-1](&event.Event{Type: event.GamepadButtonDown, GamepadButton: GPBtnLeft})
	btnH[len(btnH)-1](&event.Event{Type: event.GamepadButtonDown, GamepadButton: GPBtnRight})

	activated := false
	gn.OnActivate(func(w Navigable) { activated = true })
	btnH[len(btnH)-1](&event.Event{Type: event.GamepadButtonDown, GamepadButton: GPBtnA})
	if !activated {
		t.Error("expected activate from A button")
	}

	cancelled := false
	gn.OnCancel(func() { cancelled = true })
	btnH[len(btnH)-1](&event.Event{Type: event.GamepadButtonDown, GamepadButton: GPBtnB})
	if !cancelled {
		t.Error("expected cancel from B button")
	}
}

func TestGamepadAxisHandler(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	gn := NewGamepadNavigator(tree, cfg)
	gn.SetDeadzone(0.3)
	w1 := &mockNav{Base: NewBase(tree, "custom", cfg), focusable: true}
	w2 := &mockNav{Base: NewBase(tree, "custom", cfg), focusable: true}
	gn.AddWidget(w1)
	gn.AddWidget(w2)

	axisH := tree.Handlers(gn.ElementID(), event.GamepadAxis)
	if len(axisH) == 0 {
		t.Fatal("expected gamepad axis handler")
	}

	// Left stick X positive (right)
	axisH[len(axisH)-1](&event.Event{Type: event.GamepadAxis, GamepadAxis: GPAxisLeftX, GamepadValue: 0.8})
	// Left stick X negative (left)
	axisH[len(axisH)-1](&event.Event{Type: event.GamepadAxis, GamepadAxis: GPAxisLeftX, GamepadValue: -0.8})
	// Left stick X neutral (reset)
	axisH[len(axisH)-1](&event.Event{Type: event.GamepadAxis, GamepadAxis: GPAxisLeftX, GamepadValue: 0.0})

	// Left stick Y positive (down)
	axisH[len(axisH)-1](&event.Event{Type: event.GamepadAxis, GamepadAxis: GPAxisLeftY, GamepadValue: 0.8})
	// Left stick Y negative (up)
	axisH[len(axisH)-1](&event.Event{Type: event.GamepadAxis, GamepadAxis: GPAxisLeftY, GamepadValue: -0.8})
	// Left stick Y neutral
	axisH[len(axisH)-1](&event.Event{Type: event.GamepadAxis, GamepadAxis: GPAxisLeftY, GamepadValue: 0.0})
}

func TestGamepadTickWithMovement(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	gn := NewGamepadNavigator(tree, cfg)
	gn.SetDeadzone(0.3)
	gn.SetRepeatDelay(0.2)
	gn.SetRepeatRate(0.1)
	w1 := &mockNav{Base: NewBase(tree, "custom", cfg), focusable: true}
	w2 := &mockNav{Base: NewBase(tree, "custom", cfg), focusable: true}
	gn.AddWidget(w1)
	gn.AddWidget(w2)

	// Trigger axis movement to set moved=true
	axisH := tree.Handlers(gn.ElementID(), event.GamepadAxis)
	axisH[len(axisH)-1](&event.Event{Type: event.GamepadAxis, GamepadAxis: GPAxisLeftY, GamepadValue: 0.8})

	// Tick past repeat delay
	gn.Tick(0.3)
}

func TestGamepadDisabledHandlers(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	gn := NewGamepadNavigator(tree, cfg)
	gn.SetEnabled(false)
	w1 := &mockNav{Base: NewBase(tree, "custom", cfg), focusable: true}
	gn.AddWidget(w1)

	// All handlers should early-return when disabled
	keyH := tree.Handlers(gn.ElementID(), event.KeyDown)
	keyH[len(keyH)-1](&event.Event{Type: event.KeyDown, Key: event.KeyArrowDown})

	btnH := tree.Handlers(gn.ElementID(), event.GamepadButtonDown)
	btnH[len(btnH)-1](&event.Event{Type: event.GamepadButtonDown, GamepadButton: GPBtnDown})

	axisH := tree.Handlers(gn.ElementID(), event.GamepadAxis)
	axisH[len(axisH)-1](&event.Event{Type: event.GamepadAxis, GamepadAxis: GPAxisLeftX, GamepadValue: 0.8})
}

func TestGamepadDrawWithFocus(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	gn := NewGamepadNavigator(tree, cfg)
	gn.SetShowFocus(true)
	w1 := &mockNav{Base: NewBase(tree, "custom", cfg), focusable: true}
	gn.AddWidget(w1)
	setBounds(tree, w1, 10, 10, 100, 30)
	setBounds(tree, gn, 0, 0, 400, 300)
	buf := render.NewCommandBuffer()
	gn.Draw(buf)
}

// --- Layout coverage ---

func TestLayoutContentHeightAndAside(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()

	c := NewContent(tree, cfg)
	c.SetContentHeight(500)
	if c.ContentHeight() != 500 {
		t.Error("expected content height 500")
	}

	a := NewAside(tree, cfg)
	a.SetBorderRight(2.0, uimath.ColorRed)
	a.SetWidth(200)
	setBounds(tree, a, 0, 0, 200, 600)
	buf := render.NewCommandBuffer()
	a.Draw(buf)
}

// --- clampF ---

func TestClampF(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	s := NewSplitter(tree, cfg)
	s.SetMinRatio(0.2)
	s.SetMaxRatio(0.8)

	s.SetRatio(0.0) // below min
	if s.Ratio() < 0.2 {
		t.Error("expected clamped to min")
	}
	s.SetRatio(1.0) // above max
	if s.Ratio() > 0.8 {
		t.Error("expected clamped to max")
	}
	s.SetRatio(0.5)
	if s.Ratio() != 0.5 {
		t.Error("expected 0.5")
	}
}

// --- ColorPicker Draw with presets and open ---

func TestColorPickerDrawOpen(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	cp := NewColorPicker(tree, cfg)
	cp.SetPresets([]uimath.Color{uimath.ColorRed, uimath.ColorWhite, uimath.RGBA(0, 0, 1, 1)})
	cp.SetValue(uimath.ColorRed)
	setBounds(tree, cp, 0, 0, 300, 40)

	// Toggle open via click handler
	clickH := tree.Handlers(cp.ElementID(), event.MouseClick)
	if len(clickH) > 0 {
		clickH[len(clickH)-1](&event.Event{Type: event.MouseClick})
	}
	if !cp.IsOpen() {
		t.Error("expected open after click")
	}

	buf := render.NewCommandBuffer()
	cp.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlay when color picker is open")
	}
}

func TestRateClickHandler(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	r := NewRate(tree, cfg)
	r.SetCount(5)
	setBounds(tree, r, 0, 0, 200, 30)

	changed := false
	r.OnChange(func(v float32) { changed = true })

	clickH := tree.Handlers(r.ElementID(), event.MouseClick)
	if len(clickH) == 0 {
		t.Fatal("expected click handler on rate")
	}
	// Click on 3rd star (approximate position)
	clickH[len(clickH)-1](&event.Event{Type: event.MouseClick, GlobalX: 70, GlobalY: 15})
	if !changed {
		t.Error("expected change callback")
	}
	if r.Value() == 0 {
		t.Error("expected non-zero value after click")
	}
}

func TestRateClickDisabled(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	r := NewRate(tree, cfg)
	r.SetDisabled(true)
	setBounds(tree, r, 0, 0, 200, 30)
	clickH := tree.Handlers(r.ElementID(), event.MouseClick)
	clickH[len(clickH)-1](&event.Event{Type: event.MouseClick, GlobalX: 70, GlobalY: 15})
	// Should not change value when disabled
}

// --- Collapse toggle with onChange ---

func TestDockLayoutMouseAllSplitters(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	dl := NewDockLayout(tree, cfg)
	dl.AddPanel(&DockPanel{ID: "left", Title: "Left", Position: DockLeft, DockSize: 200, Visible: true})
	dl.AddPanel(&DockPanel{ID: "right", Title: "Right", Position: DockRight, DockSize: 200, Visible: true})
	dl.AddPanel(&DockPanel{ID: "top", Title: "Top", Position: DockTop, DockSize: 100, Visible: true})
	dl.AddPanel(&DockPanel{ID: "bottom", Title: "Bottom", Position: DockBottom, DockSize: 100, Visible: true})
	dl.AddPanel(&DockPanel{ID: "center", Title: "Center", Position: DockCenter, Visible: true})
	setBounds(tree, dl, 0, 0, 800, 600)

	downH := tree.Handlers(dl.ElementID(), event.MouseDown)
	moveH := tree.Handlers(dl.ElementID(), event.MouseMove)
	upH := tree.Handlers(dl.ElementID(), event.MouseUp)

	// Try clicking on each splitter zone
	// Left panel edge (around x=204 since left DockSize=200, splitter after)
	downH[len(downH)-1](&event.Event{Type: event.MouseDown, GlobalX: 202, GlobalY: 350})
	moveH[len(moveH)-1](&event.Event{Type: event.MouseMove, GlobalX: 250, GlobalY: 350})
	upH[len(upH)-1](&event.Event{Type: event.MouseUp})

	// Right panel edge
	downH[len(downH)-1](&event.Event{Type: event.MouseDown, GlobalX: 597, GlobalY: 350})
	moveH[len(moveH)-1](&event.Event{Type: event.MouseMove, GlobalX: 550, GlobalY: 350})
	upH[len(upH)-1](&event.Event{Type: event.MouseUp})

	// Top panel edge
	downH[len(downH)-1](&event.Event{Type: event.MouseDown, GlobalX: 400, GlobalY: 102})
	moveH[len(moveH)-1](&event.Event{Type: event.MouseMove, GlobalX: 400, GlobalY: 150})
	upH[len(upH)-1](&event.Event{Type: event.MouseUp})

	// Bottom panel edge
	downH[len(downH)-1](&event.Event{Type: event.MouseDown, GlobalX: 400, GlobalY: 497})
	moveH[len(moveH)-1](&event.Event{Type: event.MouseMove, GlobalX: 400, GlobalY: 450})
	upH[len(upH)-1](&event.Event{Type: event.MouseUp})
}

func TestCollapseToggleOnChange(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	c := NewCollapse(tree, cfg)
	c.SetPanels([]CollapsePanel{{Value: "a", Header: "A"}, {Value: "b", Header: "B"}})
	var changedKeys []string
	c.OnChange(func(keys []string) { changedKeys = keys })
	c.Toggle("a")
	c.Toggle("b")
	_ = changedKeys
}

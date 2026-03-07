package widget

import (
	"testing"

	"github.com/kasuganosora/ui/render"
)

func p1p2TestCfg() *Config { return DefaultConfig() }

func p1p2TestDraw(t *testing.T, w Widget) {
	t.Helper()
	buf := render.NewCommandBuffer()
	w.Draw(buf)
}

// --- P1 Widgets ---

func TestLinkWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	l := NewLink(tree, "Click me", "https://example.com", cfg)
	if l.Text() != "Click me" {
		t.Errorf("expected text 'Click me', got %q", l.Text())
	}
	if l.Href() != "https://example.com" {
		t.Errorf("expected href")
	}
	l.SetText("New")
	l.SetHref("/new")
	l.SetDisabled(true)
	p1p2TestDraw(t, l)
}

func TestDividerWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	d := NewDivider(tree, cfg)
	d.SetDirection(DividerVertical)
	d.SetText("OR")
	p1p2TestDraw(t, d)
}

func TestBadgeWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	b := NewBadge(tree, cfg)
	b.SetCount(5)
	if b.Count() != 5 {
		t.Errorf("expected count 5, got %d", b.Count())
	}
	b.SetDot(true)
	b.SetMaxCount(99)
	p1p2TestDraw(t, b)
}

func TestAvatarWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	a := NewAvatar(tree, cfg)
	a.SetText("AB")
	a.SetShape(AvatarSquare)
	a.SetSize(40)
	p1p2TestDraw(t, a)
}

func TestCardWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	c := NewCard(tree, cfg)
	c.SetTitle("Test Card")
	p1p2TestDraw(t, c)
}

func TestAlertWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	a := NewAlert(tree, "Warning message", cfg)
	a.SetAlertType(AlertWarning)
	a.SetClosable(true)
	if !a.IsVisible() {
		t.Error("alert should be visible")
	}
	p1p2TestDraw(t, a)
}

func TestPanelWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	p := NewPanel(tree, "My Panel", cfg)
	p.SetBordered(true)
	p1p2TestDraw(t, p)
}

func TestMenuWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	m := NewMenu(tree, cfg)
	m.SetItems([]MenuItem{
		{Key: "a", Label: "Item A"},
		{Key: "b", Label: "Item B"},
	})
	m.SelectItem("a")
	if m.SelectedKey() != "a" {
		t.Errorf("expected selected 'a'")
	}
	p1p2TestDraw(t, m)
}

func TestBreadcrumbWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	b := NewBreadcrumb(tree, cfg)
	b.SetItems([]BreadcrumbItem{
		{Label: "Home"},
		{Label: "Products"},
	})
	p1p2TestDraw(t, b)
}

func TestPaginationWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	p := NewPagination(tree, cfg)
	p.SetTotal(100)
	p.SetPageSize(10)
	if p.TotalPages() != 10 {
		t.Errorf("expected 10 pages, got %d", p.TotalPages())
	}
	p.SetCurrent(5)
	if p.Current() != 5 {
		t.Errorf("expected current 5")
	}
	p1p2TestDraw(t, p)
}

func TestInputNumberWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	in := NewInputNumber(tree, cfg)
	in.SetValue(10)
	in.SetMin(0)
	in.SetMax(100)
	in.SetStep(5)
	in.Increment()
	if in.Value() != 15 {
		t.Errorf("expected 15, got %g", in.Value())
	}
	in.Decrement()
	if in.Value() != 10 {
		t.Errorf("expected 10, got %g", in.Value())
	}
	p1p2TestDraw(t, in)
}

func TestSliderWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	s := NewSlider(tree, cfg)
	s.SetValue(50)
	s.SetMin(0)
	s.SetMax(100)
	if s.Value() != 50 {
		t.Errorf("expected 50, got %g", s.Value())
	}
	p1p2TestDraw(t, s)
}

func TestColorPickerWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	cp := NewColorPicker(tree, cfg)
	if cp.IsOpen() {
		t.Error("should not be open initially")
	}
	p1p2TestDraw(t, cp)
}

func TestListWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	l := NewList(tree, cfg)
	l.SetItems([]ListItem{
		{Title: "Item 1"},
		{Title: "Item 2"},
	})
	if len(l.Items()) != 2 {
		t.Errorf("expected 2 items, got %d", len(l.Items()))
	}
	p1p2TestDraw(t, l)
}

func TestVirtualListWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	vl := NewVirtualList(tree, cfg)
	vl.SetItemCount(1000)
	vl.SetScrollY(500)
	ch := vl.ContentHeight()
	if ch <= 0 {
		t.Error("expected positive content height")
	}
	p1p2TestDraw(t, vl)
}

func TestCollapseWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	c := NewCollapse(tree, cfg)
	c.SetPanels([]CollapsePanel{
		{Key: "a", Title: "Panel A"},
		{Key: "b", Title: "Panel B"},
	})
	c.Toggle("a")
	if !c.IsActive("a") {
		t.Error("panel 'a' should be active")
	}
	c.SetAccordion(true)
	c.Toggle("b")
	p1p2TestDraw(t, c)
}

func TestDrawerWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	d := NewDrawer(tree, "Test Drawer", cfg)
	d.Open()
	if !d.IsVisible() {
		t.Error("drawer should be visible")
	}
	d.Close()
	if d.IsVisible() {
		t.Error("drawer should be hidden")
	}
	p1p2TestDraw(t, d)
}

func TestSplitterWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	s := NewSplitter(tree, cfg)
	s.SetRatio(0.3)
	if s.Ratio() != 0.3 {
		t.Errorf("expected ratio 0.3, got %g", s.Ratio())
	}
	p1p2TestDraw(t, s)
}

func TestSubWindowWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	w := NewSubWindow(tree, "Test Window", cfg)
	w.SetPosition(100, 100)
	w.SetSize(400, 300)
	if !w.IsVisible() {
		t.Error("window should be visible")
	}
	buf := render.NewCommandBuffer()
	w.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlay commands from SubWindow")
	}
	w.Close()
	if w.IsVisible() {
		t.Error("window should be hidden after close")
	}
}

func TestPopoverWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	p := NewPopover(tree, cfg)
	p.SetTitle("Popover Title")
	p.Open()
	if !p.IsVisible() {
		t.Error("popover should be visible")
	}
	p.Close()
	if p.IsVisible() {
		t.Error("popover should be hidden")
	}
}

func TestNotificationWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	n := NewNotification(tree, "Title", "Message", cfg)
	n.SetType(NotificationSuccess)
	n.SetPosition(10, 10)
	if !n.IsVisible() {
		t.Error("notification should be visible")
	}
	buf := render.NewCommandBuffer()
	n.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlay commands from Notification")
	}
	n.Close()
	if n.IsVisible() {
		t.Error("notification should be hidden")
	}
}

func TestContextMenuWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	cm := NewContextMenu(tree, cfg)
	cm.AddItem(ContextMenuItem{Label: "Copy"})
	cm.AddDivider()
	cm.AddItem(ContextMenuItem{Label: "Paste"})
	if len(cm.Items()) != 3 {
		t.Errorf("expected 3 items, got %d", len(cm.Items()))
	}
	cm.Show(100, 200)
	if !cm.IsVisible() {
		t.Error("context menu should be visible")
	}
	cm.Hide()
	if cm.IsVisible() {
		t.Error("context menu should be hidden")
	}
}

func TestPortalWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	p := NewPortal(tree, cfg)
	btn := NewButton(tree, "test", cfg)
	p.SetContent(btn)
	if p.Content() == nil {
		t.Error("portal should have content")
	}
	p1p2TestDraw(t, p)
}

func TestTableWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	tbl := NewTable(tree, []TableColumn{
		{Title: "Name"},
		{Title: "Age", Width: 80},
	}, cfg)
	tbl.AddRow([]string{"Alice", "30"})
	tbl.AddRow([]string{"Bob", "25"})
	if tbl.RowCount() != 2 {
		t.Errorf("expected 2 rows, got %d", tbl.RowCount())
	}
	tbl.ClearRows()
	if tbl.RowCount() != 0 {
		t.Errorf("expected 0 rows after clear")
	}
}

func TestTreeWidgetP1(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	tw := NewTree(tree, cfg)
	tw.AddRoot(&TreeNode{
		Key:   "root",
		Label: "Root",
		Children: []*TreeNode{
			{Key: "child1", Label: "Child 1"},
			{Key: "child2", Label: "Child 2"},
		},
	})
	found := tw.FindNode("child1")
	if found == nil || found.Label != "Child 1" {
		t.Error("expected to find child1")
	}
	tw.ExpandAll()
	tw.CollapseAll()
}

func TestDatePickerWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	dp := NewDatePicker(tree, cfg)
	dp.SetDate(2026, 3, 15)
	if dp.Year() != 2026 || dp.Month() != 3 || dp.Day() != 15 {
		t.Error("expected date 2026-3-15")
	}
	p1p2TestDraw(t, dp)
}

func TestTimePickerWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	tp := NewTimePicker(tree, cfg)
	tp.SetTime(14, 30, 0)
	if tp.Hour() != 14 || tp.Minute() != 30 {
		t.Error("expected time 14:30")
	}
	p1p2TestDraw(t, tp)
}

// --- P2 Widgets ---

func TestStatisticWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	s := NewStatistic(tree, "Users", "1,234", cfg)
	s.SetPrefix("$")
	s.SetSuffix("/mo")
	if s.Value() != "1,234" {
		t.Errorf("expected value '1,234'")
	}
}

func TestSkeletonWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	s := NewSkeleton(tree, cfg)
	s.SetRows(5)
	s.SetAvatar(true)
	if s.Rows() != 5 {
		t.Errorf("expected 5 rows")
	}
}

func TestWatermarkWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	w := NewWatermark(tree, "CONFIDENTIAL", cfg)
	w.SetGap(100, 60)
	p1p2TestDraw(t, w)
}

func TestTimelineWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	tl := NewTimeline(tree, cfg)
	tl.AddItem(TimelineItem{Label: "Created", Status: TimelineSuccess})
	tl.AddItem(TimelineItem{Label: "Processing", Status: TimelineDefault})
	if len(tl.Items()) != 2 {
		t.Errorf("expected 2 items")
	}
	tl.ClearItems()
	if len(tl.Items()) != 0 {
		t.Error("expected 0 items after clear")
	}
}

func TestStepsWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	s := NewSteps(tree, cfg)
	s.AddStep(StepItem{Title: "Step 1"})
	s.AddStep(StepItem{Title: "Step 2"})
	s.AddStep(StepItem{Title: "Step 3"})
	s.SetCurrent(1)
	if s.Current() != 1 {
		t.Errorf("expected current 1")
	}
}

func TestRateWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	r := NewRate(tree, cfg)
	r.SetValue(3)
	if r.Value() != 3 {
		t.Errorf("expected value 3")
	}
	r.SetCount(10)
	p1p2TestDraw(t, r)
}

func TestRangeInputWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	ri := NewRangeInput(tree, 0, 100, cfg)
	ri.SetRange(20, 80)
	if ri.Low() != 20 || ri.High() != 80 {
		t.Errorf("expected range 20-80")
	}
	p1p2TestDraw(t, ri)
}

func TestTagInputWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	ti := NewTagInput(tree, cfg)
	ti.AddTag("Go")
	ti.AddTag("Rust")
	ti.AddTag("Python")
	if len(ti.Tags()) != 3 {
		t.Errorf("expected 3 tags, got %d", len(ti.Tags()))
	}
	ti.RemoveTag(1)
	if len(ti.Tags()) != 2 {
		t.Errorf("expected 2 tags after remove")
	}
	ti.ClearTags()
	if len(ti.Tags()) != 0 {
		t.Error("expected 0 tags after clear")
	}
}

func TestAutoCompleteWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	ac := NewAutoComplete(tree, cfg)
	ac.SetSuggestions([]string{"Apple", "Application", "Banana"})
	ac.SetText("App")
	if len(ac.Filtered()) != 2 {
		t.Errorf("expected 2 filtered, got %d", len(ac.Filtered()))
	}
	ac.SelectItem(0)
	if ac.Text() != "Apple" {
		t.Errorf("expected 'Apple', got %q", ac.Text())
	}
}

func TestCascaderWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	c := NewCascader(tree, cfg)
	c.SetOptions([]*CascaderOption{
		{Label: "Asia", Value: "asia", Children: []*CascaderOption{
			{Label: "China", Value: "china"},
			{Label: "Japan", Value: "japan"},
		}},
	})
	c.SetSelected([]string{"asia", "china"})
	if len(c.Selected()) != 2 {
		t.Errorf("expected 2 selected")
	}
}

func TestTreeSelectWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	ts := NewTreeSelect(tree, cfg)
	ts.SetRoots([]*TreeNode{
		{Key: "a", Label: "A", Children: []*TreeNode{
			{Key: "a1", Label: "A1"},
		}},
	})
	ts.SetSelected("a1")
	if ts.Selected() != "a1" {
		t.Errorf("expected selected 'a1'")
	}
}

func TestTransferWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	tr := NewTransfer(tree, cfg)
	tr.SetSource([]TransferItem{
		{Key: "1", Label: "Item 1"},
		{Key: "2", Label: "Item 2"},
		{Key: "3", Label: "Item 3"},
	})
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
}

func TestUploadWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	u := NewUpload(tree, cfg)
	u.AddFile(UploadFile{Name: "test.txt", Size: 1024, Status: "done"})
	if len(u.Files()) != 1 {
		t.Errorf("expected 1 file")
	}
	u.RemoveFile(0)
	if len(u.Files()) != 0 {
		t.Error("expected 0 files after remove")
	}
}

func TestDateRangePickerWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	drp := NewDateRangePicker(tree, cfg)
	drp.SetStartDate(2026, 1, 1)
	drp.SetEndDate(2026, 12, 31)
	sy, sm, sd := drp.StartDate()
	if sy != 2026 || sm != 1 || sd != 1 {
		t.Error("expected start date 2026-1-1")
	}
	p1p2TestDraw(t, drp)
}

func TestAnchorWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	a := NewAnchor(tree, cfg)
	a.SetLinks([]AnchorLink{
		{Title: "Section 1", Href: "#s1"},
		{Title: "Section 2", Href: "#s2"},
	})
	a.SetActive("#s1")
	if a.Active() != "#s1" {
		t.Errorf("expected active '#s1'")
	}
}

func TestBackTopWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	bt := NewBackTop(tree, cfg)
	bt.SetVisible(true)
	if !bt.IsVisible() {
		t.Error("expected visible")
	}
}

func TestImageViewerWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	iv := NewImageViewer(tree, cfg)
	iv.SetZoom(2)
	if iv.Zoom() != 2 {
		t.Errorf("expected zoom 2")
	}
	iv.ZoomIn()
	iv.ZoomOut()
	iv.ResetZoom()
	if iv.Zoom() != 1 {
		t.Error("expected zoom reset to 1")
	}
}

func TestVirtualGridWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	vg := NewVirtualGrid(tree, 4, 100, cfg)
	if vg.Cols() != 4 || vg.RowCount() != 100 {
		t.Error("expected 4 cols, 100 rows")
	}
	total := vg.TotalHeight()
	if total <= 0 {
		t.Error("expected positive total height")
	}
}

func TestMenuBarWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	mb := NewMenuBar(tree, cfg)
	mb.AddItem(MenuBarItem{Label: "File"})
	mb.AddItem(MenuBarItem{Label: "Edit"})
	if len(mb.Items()) != 2 {
		t.Errorf("expected 2 items")
	}
}

func TestAffixWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	a := NewAffix(tree, cfg)
	btn := NewButton(tree, "test", cfg)
	a.SetContent(btn)
	if a.Content() == nil {
		t.Error("expected content")
	}
	a.SetAffixed(true)
	if !a.IsAffixed() {
		t.Error("expected affixed")
	}
}

func TestSwiperWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	s := NewSwiper(tree, cfg)
	s.AddPanel(NewButton(tree, "slide1", cfg))
	s.AddPanel(NewButton(tree, "slide2", cfg))
	s.AddPanel(NewButton(tree, "slide3", cfg))
	if s.PanelCount() != 3 {
		t.Errorf("expected 3 panels")
	}
	s.Next()
	if s.Current() != 1 {
		t.Errorf("expected current 1 after next")
	}
	s.Prev()
	if s.Current() != 0 {
		t.Errorf("expected current 0 after prev")
	}
}

func TestCalendarWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	c := NewCalendar(tree, cfg)
	c.SetYear(2026)
	c.SetMonth(3)
	c.SetSelected(15)
	if c.Year() != 2026 || c.Month() != 3 || c.Selected() != 15 {
		t.Error("expected date state")
	}
	c.NextMonth()
	if c.Month() != 4 {
		t.Errorf("expected month 4 after next")
	}
	c.PrevMonth()
	if c.Month() != 3 {
		t.Errorf("expected month 3 after prev")
	}
}

func TestCommentWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	c := NewComment(tree, CommentData{Author: "Alice", Content: "Hello!"}, cfg)
	if c.Data().Author != "Alice" {
		t.Error("expected author Alice")
	}
	reply := NewComment(tree, CommentData{Author: "Bob", Content: "Hi!"}, cfg)
	c.AddReply(reply)
	if len(c.Replies()) != 1 {
		t.Error("expected 1 reply")
	}
}

func TestDescriptionsWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	d := NewDescriptions(tree, cfg)
	d.SetTitle("User Info")
	d.AddItem(DescriptionItem{Label: "Name", Value: "Alice"})
	d.AddItem(DescriptionItem{Label: "Age", Value: "30"})
	if len(d.Items()) != 2 {
		t.Errorf("expected 2 items")
	}
	d.ClearItems()
	if len(d.Items()) != 0 {
		t.Error("expected 0 items after clear")
	}
}

func TestPopconfirmWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	p := NewPopconfirm(tree, "Are you sure?", cfg)
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
}

func TestGuideWidget(t *testing.T) {
	tree := newTestTree()
	cfg := p1p2TestCfg()
	g := NewGuide(tree, cfg)
	g.SetSteps([]GuideStep{
		{Title: "Step 1", Description: "Click here"},
		{Title: "Step 2", Description: "Then here"},
		{Title: "Step 3", Description: "Done!"},
	})
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
	g.Next() // past last step
	if !finished {
		t.Error("expected finish callback")
	}
	if g.IsVisible() {
		t.Error("should be hidden after finish")
	}
}

package widget

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// --- helpers ---

func boostTree() *core.Tree { return core.NewTree() }
func boostCfg() *Config     { return DefaultConfig() }
func boostCfgText() *Config { return cfgWithTextRenderer() }

func boostLayout(tree *core.Tree, w Widget, x, y, ww, hh float32) {
	tree.SetLayout(w.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(x, y, ww, hh),
	})
}

// === Nil-config constructor coverage (all the 66.7% New* funcs) ===

func TestNilConfigConstructors(t *testing.T) {
	tree := boostTree()
	_ = NewBadge(tree, nil)
	_ = NewCalendar(tree, nil)
	_ = NewComment(tree, CommentData{}, nil)
	_ = NewContextMenu(tree, nil)
	_ = NewDescriptions(tree, nil)
	_ = NewDivider(tree, nil)
	_ = NewGuide(tree, nil)
	_ = NewList(tree, nil)
	_ = NewVirtualList(tree, nil)
	_ = NewMenuBar(tree, nil)
	_ = NewNotification(tree, "t", "m", nil)
	_ = NewPanel(tree, "p", nil)
	_ = NewPopconfirm(tree, "q", nil)
	_ = NewPortal(tree, nil)
	_ = NewRichText(tree, nil)
	_ = NewSkeleton(tree, nil)
	_ = NewStatistic(tree, "t", "v", nil)
	_ = NewSteps(tree, nil)
	_ = NewSwiper(tree, nil)
	_ = NewTagInput(tree, nil)
	_ = NewTimeline(tree, nil)
	_ = NewTransfer(tree, nil)
	_ = NewTreeSelect(tree, nil)
	_ = NewTree(tree, nil)
	_ = NewUpload(tree, nil)
	_ = NewVirtualGrid(tree, 3, 5, nil)
	_ = NewWatermark(tree, "wm", nil)
	_ = NewTimePicker(tree, nil)
	_ = NewLink(tree, "l", "u", nil)
	_ = NewBackTop(tree, nil)
	_ = NewRangeInput(tree, 0, 100, nil)
	_ = NewImageViewer(tree, nil)
}

// === DragDrop full workflow coverage ===

func TestDragDropFullWorkflow(t *testing.T) {
	tree := boostTree()
	dd := NewDragDropManager(tree)

	cfg := boostCfg()
	src := NewButton(tree, "drag me", cfg)
	boostLayout(tree, src, 10, 10, 80, 30)

	tgt := NewButton(tree, "drop here", cfg)
	boostLayout(tree, tgt, 200, 10, 80, 30)

	var dropped any
	var entered bool
	var left bool
	dd.RegisterSource(&DragSource{Widget: src, Data: "hello"})
	dd.RegisterTarget(&DropTarget{
		Widget:  tgt,
		Accept:  func(data any) bool { return true },
		OnDrop:  func(data any) { dropped = data },
		OnEnter: func(data any) { entered = true },
		OnLeave: func() { left = true },
	})

	// Simulate the tree handler setting sourceID (normally done via tree.AddHandler)
	dd.sourceID = src.ElementID()
	dd.dragStartX = 15
	dd.dragStartY = 15

	// Move past threshold (5px = 25 squared distance)
	dd.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: 25, GlobalY: 25})
	if !dd.IsDragging() {
		t.Error("should be dragging after threshold")
	}
	if dd.DragData() != "hello" {
		t.Error("drag data mismatch")
	}

	// Draw while dragging (default indicator)
	buf := render.NewCommandBuffer()
	dd.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected drag indicator")
	}

	// Move over target area
	dd.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: 210, GlobalY: 15})

	// Move away from target (triggers leave)
	dd.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: 500, GlobalY: 500})

	// Move back over target
	dd.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: 210, GlobalY: 15})

	// Drop
	dd.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 210, GlobalY: 15})
	if dd.IsDragging() {
		t.Error("should not be dragging after drop")
	}
	_ = dropped
	_ = entered
	_ = left
}

func TestDragDropNotDraggingDraw(t *testing.T) {
	tree := boostTree()
	dd := NewDragDropManager(tree)
	buf := render.NewCommandBuffer()
	dd.Draw(buf)
	if buf.Len() != 0 {
		t.Error("should not draw when not dragging")
	}
}

func TestDragDropMouseUpWithoutDrag(t *testing.T) {
	tree := boostTree()
	dd := NewDragDropManager(tree)
	dd.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 0, GlobalY: 0})
}

func TestDragDropStartDragNilSource(t *testing.T) {
	tree := boostTree()
	dd := NewDragDropManager(tree)
	cfg := boostCfg()
	src := NewButton(tree, "s", cfg)
	dd.RegisterSource(&DragSource{Widget: src, Data: "x"})
	// Simulate sourceID set but then remove the source
	dd.sourceID = src.ElementID()
	dd.dragStartX = 0
	dd.dragStartY = 0
	delete(dd.sources, src.ElementID())
	dd.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: 100, GlobalY: 100})
	if dd.IsDragging() {
		t.Error("should not drag with nil source")
	}
}

func TestDragDropFinishNoTarget(t *testing.T) {
	tree := boostTree()
	dd := NewDragDropManager(tree)
	cfg := boostCfg()
	src := NewButton(tree, "s", cfg)
	boostLayout(tree, src, 0, 0, 50, 50)
	dd.RegisterSource(&DragSource{Widget: src, Data: "x"})
	dd.sourceID = src.ElementID()
	dd.dragStartX = 10
	dd.dragStartY = 10
	dd.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: 100, GlobalY: 100})
	dd.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 500, GlobalY: 500})
}

func TestDragDropFinishWithAcceptFalse(t *testing.T) {
	tree := boostTree()
	dd := NewDragDropManager(tree)
	cfg := boostCfg()
	src := NewButton(tree, "s", cfg)
	boostLayout(tree, src, 0, 0, 50, 50)
	tgt := NewButton(tree, "t", cfg)
	boostLayout(tree, tgt, 200, 0, 50, 50)

	dd.RegisterSource(&DragSource{Widget: src, Data: "x"})
	dd.RegisterTarget(&DropTarget{
		Widget: tgt,
		Accept: func(data any) bool { return false },
		OnDrop: func(data any) { t.Error("should not drop") },
	})

	dd.sourceID = src.ElementID()
	dd.dragStartX = 10
	dd.dragStartY = 10
	dd.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: 210, GlobalY: 10})
	dd.HandleEvent(&event.Event{Type: event.MouseUp, GlobalX: 210, GlobalY: 10})
}

func TestDragDropWithDragIcon(t *testing.T) {
	tree := boostTree()
	dd := NewDragDropManager(tree)
	cfg := boostCfg()
	src := NewButton(tree, "s", cfg)
	boostLayout(tree, src, 0, 0, 50, 50)
	icon := NewButton(tree, "icon", cfg)
	boostLayout(tree, icon, 0, 0, 20, 20)

	dd.RegisterSource(&DragSource{Widget: src, Data: "x", DragIcon: icon})
	dd.sourceID = src.ElementID()
	dd.dragStartX = 10
	dd.dragStartY = 10
	dd.HandleEvent(&event.Event{Type: event.MouseMove, GlobalX: 100, GlobalY: 100})

	buf := render.NewCommandBuffer()
	dd.Draw(buf)
}

// === TextRenderer path coverage for widgets ===

func TestCommentDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	c := NewComment(tree, CommentData{Author: "Alice", Content: "Hello", Time: "2m ago"}, cfg)
	boostLayout(tree, c, 0, 0, 400, 100)
	buf := render.NewCommandBuffer()
	c.Draw(buf)
	if buf.Len() < 2 {
		t.Error("expected text render commands")
	}
}

func TestCommentDrawNoTime(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	c := NewComment(tree, CommentData{Author: "Bob", Content: "World"}, cfg)
	boostLayout(tree, c, 0, 0, 400, 100)
	buf := render.NewCommandBuffer()
	c.Draw(buf)
}

func TestStatisticDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	s := NewStatistic(tree, "Users", "1234", cfg)
	s.SetPrefix("$")
	s.SetSuffix(" total")
	s.SetColor(uimath.RGBA(1, 0, 0, 1))
	boostLayout(tree, s, 0, 0, 200, 80)
	buf := render.NewCommandBuffer()
	s.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected text render commands")
	}
}

func TestStatisticDrawNoColor(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	s := NewStatistic(tree, "Count", "42", cfg)
	boostLayout(tree, s, 0, 0, 200, 80)
	buf := render.NewCommandBuffer()
	s.Draw(buf)
}

func TestWatermarkDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	w := NewWatermark(tree, "DRAFT", cfg)
	w.SetX(50)
	w.SetY(40)
	w.SetColor(uimath.RGBA(0, 0, 0, 0.1))
	w.SetAlpha(0.5)
	boostLayout(tree, w, 0, 0, 300, 200)
	buf := render.NewCommandBuffer()
	w.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected watermark text commands")
	}
}

func TestWatermarkDrawEmptyText(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	w := NewWatermark(tree, "", cfg)
	boostLayout(tree, w, 0, 0, 300, 200)
	buf := render.NewCommandBuffer()
	w.Draw(buf)
}

func TestGuideDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	g := NewGuide(tree, cfg)
	g.SetSteps([]GuideStep{
		{Title: "Step 1", Description: "First step", TargetX: 50, TargetY: 50, TargetW: 100, TargetH: 40},
		{Title: "Step 2", Description: "", TargetX: 200, TargetY: 50, TargetW: 0, TargetH: 0},
	})
	g.Start()
	buf := render.NewCommandBuffer()
	g.Draw(buf)
	if buf.Len() < 3 {
		t.Error("expected guide overlay commands")
	}
}

func TestGuideDrawNoTargetSize(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	g := NewGuide(tree, cfg)
	g.SetSteps([]GuideStep{
		{Title: "Only", TargetW: 0, TargetH: 0},
	})
	g.Start()
	buf := render.NewCommandBuffer()
	g.Draw(buf)
}

func TestGuideNextPrevFinish(t *testing.T) {
	tree := boostTree()
	g := NewGuide(tree, boostCfg())
	var finished bool
	var changed int
	g.OnFinish(func() { finished = true })
	g.OnChange(func(i int) { changed = i })
	g.SetSteps([]GuideStep{{Title: "A"}, {Title: "B"}, {Title: "C"}})
	g.Start()
	g.Next()
	if changed != 1 {
		t.Error("expected change to 1")
	}
	g.Prev()
	if changed != 0 {
		t.Error("expected change to 0")
	}
	g.Prev() // at 0, no-op
	g.Next()
	g.Next() // current=2
	g.Next() // at last step, calls Finish
	if !finished {
		t.Error("expected finish")
	}
}

func TestCalendarDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	c := NewCalendar(tree, cfg)
	c.SetMonth(2)
	c.SetYear(2024) // leap year
	c.SetSelected(15)
	boostLayout(tree, c, 0, 0, 300, 300)
	buf := render.NewCommandBuffer()
	c.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected calendar commands")
	}
}

func TestCalendarAllMonths(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	for m := 1; m <= 12; m++ {
		c := NewCalendar(tree, cfg)
		c.SetMonth(m)
		boostLayout(tree, c, 0, 0, 300, 300)
		buf := render.NewCommandBuffer()
		c.Draw(buf)
	}
}

func TestCalendarDaysInMonth(t *testing.T) {
	tree := boostTree()
	cfg := boostCfg()
	// Feb non-leap
	c := NewCalendar(tree, cfg)
	c.SetYear(2025)
	c.SetMonth(2)
	boostLayout(tree, c, 0, 0, 300, 300)
	c.Draw(render.NewCommandBuffer())

	// 30-day month
	c2 := NewCalendar(tree, cfg)
	c2.SetMonth(4)
	boostLayout(tree, c2, 0, 0, 300, 300)
	c2.Draw(render.NewCommandBuffer())

	// Century non-leap (1900)
	c3 := NewCalendar(tree, cfg)
	c3.SetYear(1900)
	c3.SetMonth(2)
	boostLayout(tree, c3, 0, 0, 300, 300)
	c3.Draw(render.NewCommandBuffer())

	// Century leap (2000)
	c4 := NewCalendar(tree, cfg)
	c4.SetYear(2000)
	c4.SetMonth(2)
	boostLayout(tree, c4, 0, 0, 300, 300)
	c4.Draw(render.NewCommandBuffer())
}

func TestNotificationDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	for _, nt := range []NotificationTheme{NotificationThemeInfo, NotificationThemeSuccess, NotificationThemeWarning, NotificationThemeError} {
		n := NewNotification(tree, "Title", "Message", cfg)
		n.SetTheme(nt)
		n.SetPosition(10, 10)
		buf := render.NewCommandBuffer()
		n.Draw(buf)
		if buf.Len() < 3 {
			t.Errorf("expected notification commands for type %d", nt)
		}
	}
}

func TestNotificationDrawNoMessage(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	n := NewNotification(tree, "Only Title", "", cfg)
	n.Draw(render.NewCommandBuffer())
}

func TestNotificationCloseCallback(t *testing.T) {
	tree := boostTree()
	n := NewNotification(tree, "t", "m", boostCfg())
	var closed bool
	n.OnClose(func() { closed = true })
	n.Close()
	if !closed {
		t.Error("expected close callback")
	}
}

func TestImageViewerDrawWithTexture(t *testing.T) {
	tree := boostTree()
	iv := NewImageViewer(tree, boostCfg())
	iv.SetTexture(42)
	iv.SetZoom(1.5)
	iv.SetPan(10, 20)
	boostLayout(tree, iv, 0, 0, 200, 200)
	buf := render.NewCommandBuffer()
	iv.Draw(buf)
	if buf.Len() < 2 {
		t.Error("expected background + image commands")
	}
}

func TestImageViewerNotVisible(t *testing.T) {
	tree := boostTree()
	iv := NewImageViewer(tree, boostCfg())
	iv.SetVisible(false)
	boostLayout(tree, iv, 0, 0, 200, 200)
	buf := render.NewCommandBuffer()
	iv.Draw(buf)
	if buf.Len() != 0 {
		t.Error("hidden viewer should not draw")
	}
}

func TestImageViewerZoomOps(t *testing.T) {
	tree := boostTree()
	iv := NewImageViewer(tree, boostCfg())
	iv.ZoomIn()
	if iv.Zoom() <= 1 {
		t.Error("zoom in should increase")
	}
	iv.ZoomOut()
	iv.ResetZoom()
	if iv.Zoom() != 1 {
		t.Error("reset should return to 1")
	}
}

// === Breadcrumb / Anchor ===

func TestBreadcrumbDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	bc := NewBreadcrumb(tree, cfg)
	bc.SetOptions([]BreadcrumbItem{
		{Content: "Home", Href: "/"},
		{Content: "Products", Href: "/products"},
		{Content: "Detail"},
	})
	boostLayout(tree, bc, 0, 0, 400, 30)
	buf := render.NewCommandBuffer()
	bc.Draw(buf)
}

func TestAnchorDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	a := NewAnchor(tree, cfg)
	a.SetLinks([]AnchorLink{
		{Title: "Section A", Href: "#a"},
		{Title: "Section B", Href: "#b"},
	})
	a.SetActive("#a")
	boostLayout(tree, a, 0, 0, 200, 200)
	buf := render.NewCommandBuffer()
	a.Draw(buf)
}

// === DatePicker / TimePicker with TextRenderer ===

func TestDatePickerDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	dp := NewDatePicker(tree, cfg)
	dp.SetOpen(true)
	boostLayout(tree, dp, 0, 0, 300, 40)
	dp.Draw(render.NewCommandBuffer())
}

func TestTimePickerDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	tp := NewTimePicker(tree, cfg)
	tp.SetOpen(true)
	boostLayout(tree, tp, 0, 0, 300, 40)
	tp.Draw(render.NewCommandBuffer())
}

func TestDateRangePickerDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	drp := NewDateRangePicker(tree, cfg)
	drp.SetOpen(true)
	boostLayout(tree, drp, 0, 0, 300, 40)
	drp.Draw(render.NewCommandBuffer())
}

// === Descriptions with TextRenderer ===

func TestDescriptionsDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	d := NewDescriptions(tree, cfg)
	d.SetTitle("Details")
	d.AddItem(DescriptionItem{Label: "Name", Value: "Test"})
	d.AddItem(DescriptionItem{Label: "Age", Value: "25"})
	boostLayout(tree, d, 0, 0, 400, 200)
	d.Draw(render.NewCommandBuffer())
}

// === List with TextRenderer ===

func TestListDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	l := NewList(tree, cfg)
	l.SetItems([]ListItem{
		{Title: "Item A", Description: "Desc A"},
		{Title: "Item B", Extra: "extra"},
	})
	boostLayout(tree, l, 0, 0, 200, 300)
	l.Draw(render.NewCommandBuffer())
}

// === AutoComplete with TextRenderer ===

func TestAutoCompleteDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	ac := NewAutoComplete(tree, cfg)
	ac.SetSuggestions([]string{"Apple", "Banana"})
	ac.SetOpen(true)
	boostLayout(tree, ac, 0, 0, 200, 40)
	ac.Draw(render.NewCommandBuffer())
}

// === RichText with TextRenderer ===

func TestRichTextDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	rt := NewRichText(tree, cfg)
	rt.SetSpans([]RichSpan{
		{Text: "Bold", Bold: true},
		{Text: " normal"},
		{Text: " colored", Color: uimath.RGBA(1, 0, 0, 1)},
	})
	boostLayout(tree, rt, 0, 0, 300, 100)
	rt.Draw(render.NewCommandBuffer())
}

// === Steps with TextRenderer ===

func TestStepsDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	s := NewSteps(tree, cfg)
	s.AddStep(StepItem{Title: "Start", Content: "Begin"})
	s.AddStep(StepItem{Title: "Middle", Content: "Processing"})
	s.AddStep(StepItem{Title: "End", Content: "Done"})
	s.SetCurrent(1)
	boostLayout(tree, s, 0, 0, 600, 80)
	s.Draw(render.NewCommandBuffer())
}

// === Timeline with TextRenderer ===

func TestTimelineDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	tl := NewTimeline(tree, cfg)
	tl.AddItem(TimelineItem{Label: "Created", Content: "Item created"})
	tl.AddItem(TimelineItem{Label: "Updated", Content: "Item updated", DotColor: "primary"})
	boostLayout(tree, tl, 0, 0, 400, 200)
	tl.Draw(render.NewCommandBuffer())
}

// === TagInput with TextRenderer + RemoveTag ===

func TestTagInputDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	ti := NewTagInput(tree, cfg)
	ti.AddTag("Go")
	ti.AddTag("Rust")
	boostLayout(tree, ti, 0, 0, 300, 40)
	ti.Draw(render.NewCommandBuffer())
}

func TestTagInputRemoveByIndex(t *testing.T) {
	tree := boostTree()
	ti := NewTagInput(tree, boostCfg())
	ti.AddTag("A")
	ti.AddTag("B")
	ti.RemoveTag(0)
	if len(ti.Tags()) != 1 || ti.Tags()[0] != "B" {
		t.Error("expected only B after removing index 0")
	}
}

// === Upload Draw with TextRenderer ===

func TestUploadDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	u := NewUpload(tree, cfg)
	u.AddFile(UploadFile{Name: "test.png", Size: 1024, Status: "done"})
	boostLayout(tree, u, 0, 0, 300, 200)
	u.Draw(render.NewCommandBuffer())
}

// === Tree Draw with TextRenderer ===

func TestTreeDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	tw := NewTree(tree, cfg)
	tw.SetRoots([]*TreeNode{
		{Key: "a", Label: "A", Children: []*TreeNode{
			{Key: "a1", Label: "A1"},
		}},
		{Key: "b", Label: "B"},
	})
	tw.ExpandAll()
	boostLayout(tree, tw, 0, 0, 300, 300)
	tw.Draw(render.NewCommandBuffer())
}

// === Pagination with TextRenderer ===

func TestPaginationDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	p := NewPagination(tree, cfg)
	p.SetTotal(100)
	p.SetPageSize(10)
	p.SetCurrent(3)
	boostLayout(tree, p, 0, 0, 400, 40)
	p.Draw(render.NewCommandBuffer())
}

func TestPaginationTotalPagesZero(t *testing.T) {
	tree := boostTree()
	p := NewPagination(tree, boostCfg())
	p.SetTotal(0)
	tp := p.TotalPages()
	if tp != 0 {
		t.Errorf("expected 0 for empty total, got %d", tp)
	}
}

func TestPaginationGoToEdges(t *testing.T) {
	tree := boostTree()
	p := NewPagination(tree, boostCfg())
	p.SetTotal(30)
	p.SetPageSize(10)
	p.GoTo(0)
	p.GoTo(99)
}

// === MenuBar with TextRenderer ===

func TestMenuBarDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	mb := NewMenuBar(tree, cfg)
	mb.AddItem(MenuBarItem{Label: "File"})
	mb.AddItem(MenuBarItem{Label: "Edit"})
	boostLayout(tree, mb, 0, 0, 600, 30)
	mb.Draw(render.NewCommandBuffer())
}

// === Menu ToggleOpen / SelectItem ===

func TestMenuToggleAndSelect(t *testing.T) {
	tree := boostTree()
	m := NewMenu(tree, boostCfg())
	m.SetItems([]MenuItem{
		{Value: "a", Content: "A", Children: []MenuItem{{Value: "a1", Content: "A1"}}},
		{Value: "b", Content: "B"},
	})
	m.ToggleExpanded("a")
	m.ToggleExpanded("a")
	m.SelectItem("b")
	if m.Value() != "b" {
		t.Error("expected b selected")
	}
}

// === Dock TextRenderer paths ===

func TestDockLayoutDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	dl := NewDockLayout(tree, cfg)
	dl.AddPanel(&DockPanel{ID: "left", Title: "Left", Position: DockLeft})
	dl.AddPanel(&DockPanel{ID: "center1", Title: "C1", Position: DockCenter})
	dl.AddPanel(&DockPanel{ID: "center2", Title: "C2", Position: DockCenter})
	dl.AddPanel(&DockPanel{ID: "float", Title: "Float", Position: DockFloat, FloatX: 50, FloatY: 50, FloatW: 150, FloatH: 100})
	boostLayout(tree, dl, 0, 0, 800, 600)
	dl.Draw(render.NewCommandBuffer())
}

// === Swiper Next/Prev ===

func TestSwiperNavigation(t *testing.T) {
	tree := boostTree()
	cfg := boostCfg()
	s := NewSwiper(tree, cfg)
	// Add panels
	for i := 0; i < 3; i++ {
		p := NewButton(tree, "slide", cfg)
		boostLayout(tree, p, 0, 0, 100, 100)
		s.AddPanel(p)
	}
	s.Next()
	if s.Current() != 1 {
		t.Error("expected 1")
	}
	s.Next()
	s.Next() // wraps to 0
	if s.Current() != 0 {
		t.Error("expected wrap to 0")
	}
	s.Prev() // wraps to 2
	if s.Current() != 2 {
		t.Error("expected wrap to 2")
	}
	s.Prev()
	if s.Current() != 1 {
		t.Error("expected 1")
	}
}

func TestSwiperDrawWithDots(t *testing.T) {
	tree := boostTree()
	cfg := boostCfg()
	s := NewSwiper(tree, cfg)
	for i := 0; i < 3; i++ {
		p := NewButton(tree, "slide", cfg)
		boostLayout(tree, p, 0, 0, 100, 100)
		s.AddPanel(p)
	}
	boostLayout(tree, s, 0, 0, 200, 200)
	buf := render.NewCommandBuffer()
	s.Draw(buf)
}

// === Slider ratio edge case ===

func TestSliderRatioEdge(t *testing.T) {
	tree := boostTree()
	s := NewSlider(tree, boostCfg())
	s.SetMin(0)
	s.SetMax(0)
	boostLayout(tree, s, 0, 0, 200, 30)
	s.Draw(render.NewCommandBuffer())
}

// === Transfer MoveToSource ===

func TestTransferMoveToSource(t *testing.T) {
	tree := boostTree()
	tr := NewTransfer(tree, boostCfg())
	tr.SetSource([]TransferItem{{Key: "a", Label: "A"}, {Key: "b", Label: "B"}})
	tr.MoveToTarget([]string{"a", "b"})
	tr.MoveToSource([]string{"a"})
	if len(tr.Source()) != 1 || tr.Source()[0].Key != "a" {
		t.Error("expected a back in source")
	}
}

func TestTransferDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	tr := NewTransfer(tree, cfg)
	tr.SetSource([]TransferItem{{Key: "a", Label: "A"}, {Key: "b", Label: "B"}})
	tr.MoveToTarget([]string{"a"})
	boostLayout(tree, tr, 0, 0, 500, 300)
	tr.Draw(render.NewCommandBuffer())
}

// === TreeSelect Draw with TextRenderer ===

func TestTreeSelectDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	ts := NewTreeSelect(tree, cfg)
	ts.SetRoots([]*TreeNode{
		{Key: "a", Label: "A", Children: []*TreeNode{{Key: "a1", Label: "A1"}}},
	})
	ts.SetOpen(true)
	boostLayout(tree, ts, 0, 0, 300, 40)
	ts.Draw(render.NewCommandBuffer())
}

// === Cascader SetSelected deeper path ===

func TestCascaderSetSelectedDeep(t *testing.T) {
	tree := boostTree()
	c := NewCascader(tree, boostCfg())
	c.SetOptions([]*CascaderOption{
		{Value: "a", Label: "A", Children: []*CascaderOption{
			{Value: "a1", Label: "A1", Children: []*CascaderOption{
				{Value: "a1x", Label: "A1X"},
			}},
		}},
	})
	c.SetSelected([]string{"a", "a1", "a1x"})
	if len(c.Selected()) != 3 {
		t.Error("expected 3-level selection")
	}
}

// === Select with options + TextRenderer ===

func TestSelectDrawWithOptionsTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	opts := []SelectOption{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}
	s := NewSelect(tree, opts, cfg)
	boostLayout(tree, s, 0, 0, 200, 40)
	s.Draw(render.NewCommandBuffer())
}

// === MessageBox Draw with TextRenderer ===

func TestMessageBoxDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	mb := NewMessageBox(tree, "Confirm", "Are you sure?", cfg)
	boostLayout(tree, mb, 0, 0, 800, 600)
	mb.Draw(render.NewCommandBuffer())
}

// === Popconfirm with TextRenderer ===

func TestPopconfirmDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	pc := NewPopconfirm(tree, "Delete?", cfg)
	pc.SetVisible(true)
	boostLayout(tree, pc, 0, 0, 200, 40)
	pc.Draw(render.NewCommandBuffer())
}

// === Portal Draw ===

func TestPortalDraw(t *testing.T) {
	tree := boostTree()
	p := NewPortal(tree, boostCfg())
	boostLayout(tree, p, 0, 0, 800, 600)
	p.Draw(render.NewCommandBuffer())
}

// === BackTop with TextRenderer ===

func TestBackTopDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	bt := NewBackTop(tree, cfg)
	bt.SetVisible(true)
	boostLayout(tree, bt, 0, 0, 800, 600)
	bt.Draw(render.NewCommandBuffer())
}

// === Skeleton Draw ===

func TestSkeletonDrawActive(t *testing.T) {
	tree := boostTree()
	sk := NewSkeleton(tree, boostCfg())
	sk.SetLoading(true)
	sk.SetRows(3)
	boostLayout(tree, sk, 0, 0, 300, 200)
	sk.Draw(render.NewCommandBuffer())
}

// === VirtualGrid Draw ===

func TestVirtualGridDrawWithRenderer(t *testing.T) {
	tree := boostTree()
	vg := NewVirtualGrid(tree, 3, 5, boostCfg())
	vg.SetCellSize(50, 50)
	vg.SetRenderCell(func(buf *render.CommandBuffer, row, col int, rect uimath.Rect) {
		buf.DrawRect(render.RectCmd{Bounds: rect, FillColor: uimath.ColorWhite}, 1, 1)
	})
	boostLayout(tree, vg, 0, 0, 300, 300)
	buf := render.NewCommandBuffer()
	vg.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected grid cell commands")
	}
}

// === Content layout scrollbar edge ===

func TestContentSmallContent(t *testing.T) {
	tree := boostTree()
	c := NewContent(tree, boostCfg())
	boostLayout(tree, c, 0, 0, 200, 500)
	c.SetContentHeight(100) // smaller than viewport
	c.Draw(render.NewCommandBuffer())
}

// === Avatar Draw with text ===

func TestAvatarDrawWithTextRenderer(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	a := NewAvatar(tree, cfg)
	a.SetText("AB")
	boostLayout(tree, a, 0, 0, 40, 40)
	a.Draw(render.NewCommandBuffer())
}

// === Bounds empty ===

func TestWidgetBoundsEmpty(t *testing.T) {
	tree := boostTree()
	b := NewButton(tree, "test", boostCfg())
	bounds := b.Bounds()
	if !bounds.IsEmpty() {
		t.Error("expected empty bounds without layout")
	}
}

// === ColorPicker open + draw swatches ===

func TestColorPickerDrawOpenBoost(t *testing.T) {
	tree := boostTree()
	cp := NewColorPicker(tree, boostCfg())
	cp.open = true
	boostLayout(tree, cp, 0, 0, 40, 40)
	buf := render.NewCommandBuffer()
	cp.Draw(buf)
	if buf.Len() < 3 {
		t.Errorf("expected swatch + panel + preset commands, got %d", buf.Len())
	}
}

func TestColorPickerSelectColor(t *testing.T) {
	tree := boostTree()
	cp := NewColorPicker(tree, boostCfg())
	var got uimath.Color
	cp.OnChange(func(c uimath.Color) { got = c })
	red := uimath.RGBA(1, 0, 0, 1)
	cp.SelectColor(red)
	if got != red {
		t.Error("expected red")
	}
}

func TestColorPickerSelectedSwatch(t *testing.T) {
	tree := boostTree()
	cp := NewColorPicker(tree, boostCfg())
	cp.SetValue(cp.presets[0])
	cp.open = true
	boostLayout(tree, cp, 0, 0, 40, 40)
	cp.Draw(render.NewCommandBuffer())
}

// === Rate ===

func TestRateDrawWithValue(t *testing.T) {
	tree := boostTree()
	r := NewRate(tree, boostCfg())
	r.SetValue(3)
	r.SetCount(5)
	r.SetSize(24)
	r.SetDisabled(false)
	r.OnChange(func(v float32) {})
	boostLayout(tree, r, 0, 0, 200, 30)
	buf := render.NewCommandBuffer()
	r.Draw(buf)
	if buf.Len() != 5 {
		t.Errorf("expected 5 rects, got %d", buf.Len())
	}
}

// === RangeInput interaction ===

func TestRangeInputDraw(t *testing.T) {
	tree := boostTree()
	ri := NewRangeInput(tree, 0, 100, boostCfg())
	ri.SetRange(20, 80)
	ri.SetStep(5)
	boostLayout(tree, ri, 0, 0, 300, 30)
	buf := render.NewCommandBuffer()
	ri.Draw(buf)
	if buf.Len() < 3 {
		t.Error("expected track + range + thumbs")
	}
}

// === Slider updateFromMouse ===

func TestSliderUpdateFromMouse(t *testing.T) {
	tree := boostTree()
	s := NewSlider(tree, boostCfg())
	s.SetMin(0)
	s.SetMax(100)
	s.SetStep(10)
	boostLayout(tree, s, 0, 0, 200, 30)
	s.dragging = true
	s.updateFromMouse(100) // middle
	if s.Value() == 0 {
		t.Error("expected value change from drag")
	}
}

// === Collapse Toggle ===

func TestCollapseToggle(t *testing.T) {
	tree := boostTree()
	c := NewCollapse(tree, boostCfg())
	c.SetPanels([]CollapsePanel{{Value: "a", Header: "A"}, {Value: "b", Header: "B"}})
	var changedKeys []string
	c.OnChange(func(keys []string) { changedKeys = keys })
	c.Toggle("a")
	if !c.IsActive("a") {
		t.Error("expected a active")
	}
	c.Toggle("a")
	if c.IsActive("a") {
		t.Error("expected a inactive")
	}
	_ = changedKeys
}

func TestCollapseToggleAccordion(t *testing.T) {
	tree := boostTree()
	c := NewCollapse(tree, boostCfg())
	c.SetExpandMutex(true)
	c.SetPanels([]CollapsePanel{{Value: "a", Header: "A"}, {Value: "b", Header: "B"}})
	c.Toggle("a")
	c.Toggle("b") // accordion: only b
	if c.IsActive("a") {
		t.Error("a should be inactive")
	}
	c.Toggle("b") // deactivate
	if c.IsActive("b") {
		t.Error("b should be inactive")
	}
}

func TestCollapseDrawWithContent(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	inner := NewButton(tree, "inner", cfg)
	boostLayout(tree, inner, 0, 0, 300, 50)
	c := NewCollapse(tree, cfg)
	c.SetPanels([]CollapsePanel{
		{Value: "a", Header: "Panel A", Content: inner},
		{Value: "b", Header: "Panel B"},
	})
	c.Toggle("a")
	c.SetBorderless(false)
	boostLayout(tree, c, 0, 0, 300, 400)
	c.Draw(render.NewCommandBuffer())
}

// === Popover Draw placements ===

func TestPopoverDrawPlacements(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	for _, pl := range []PopupPlacement{PlacementTop, PlacementBottom, PlacementLeft, PlacementRight} {
		p := NewPopover(tree, cfg)
		p.SetTitle("Info")
		p.SetPlacement(pl)
		p.SetAnchorRect(100, 100, 80, 30)
		p.SetVisible(true)
		p.Draw(render.NewCommandBuffer())
	}
}

func TestPopoverClose(t *testing.T) {
	tree := boostTree()
	p := NewPopover(tree, boostCfg())
	var closed bool
	p.OnClose(func() { closed = true })
	p.Open()
	p.Close()
	if !closed {
		t.Error("expected close callback")
	}
}

func TestPopoverWithContent(t *testing.T) {
	tree := boostTree()
	cfg := boostCfg()
	content := NewButton(tree, "btn", cfg)
	boostLayout(tree, content, 0, 0, 100, 30)
	p := NewPopover(tree, cfg)
	p.SetContent(content)
	p.SetVisible(true)
	p.Draw(render.NewCommandBuffer())
}

// === Drawer ===

func TestDrawerNilConfig(t *testing.T) {
	tree := boostTree()
	_ = NewDrawer(tree, "Test", nil)
}

func TestDrawerDrawAllPlacements(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	for _, pl := range []DrawerPlacement{DrawerLeft, DrawerRight, DrawerTop, DrawerBottom} {
		d := NewDrawer(tree, "Side", cfg)
		d.SetPlacement(pl)
		d.Open()
		boostLayout(tree, d, 0, 0, 800, 600)
		d.Draw(render.NewCommandBuffer())
	}
}

// === Nil config constructors batch 2 ===

func TestNilConfigConstructors2(t *testing.T) {
	tree := boostTree()
	_ = NewAffix(tree, nil)
	_ = NewAlert(tree, "m", nil)
	_ = NewAnchor(tree, nil)
	_ = NewAvatar(tree, nil)
	_ = NewBreadcrumb(tree, nil)
	_ = NewCascader(tree, nil)
	_ = NewCollapse(tree, nil)
	_ = NewColorPicker(tree, nil)
	_ = NewDatePicker(tree, nil)
	_ = NewDateRangePicker(tree, nil)
	_ = NewDrawer(tree, "d", nil)
	_ = NewMessageBox(tree, "t", "c", nil)
	_ = NewPopover(tree, nil)
	_ = NewRate(tree, nil)
	_ = NewSlider(tree, nil)
	_ = NewSplitter(tree, nil)
}

// === monthName coverage ===

func TestMonthNameAll(t *testing.T) {
	for m := 0; m <= 13; m++ {
		_ = monthName(m)
	}
}

// === Dock RemovePanel + DockPanel ===

func TestDockRemovePanel(t *testing.T) {
	tree := boostTree()
	dl := NewDockLayout(tree, boostCfg())
	dl.AddPanel(&DockPanel{ID: "a", Title: "A", Position: DockLeft})
	dl.AddPanel(&DockPanel{ID: "b", Title: "B", Position: DockRight})
	dl.RemovePanel("a")
	if dl.FindPanel("a") != nil {
		t.Error("expected panel a removed")
	}
}

func TestDockPanelMethod(t *testing.T) {
	tree := boostTree()
	dl := NewDockLayout(tree, boostCfg())
	dl.AddPanel(&DockPanel{ID: "center", Title: "C", Position: DockCenter})
	dl.DockPanel("center", DockLeft)
	p := dl.FindPanel("center")
	if p == nil || p.Position != DockLeft {
		t.Error("expected docked to left")
	}
}

// === Pagination TotalPages pageSize=0 ===

func TestPaginationPageSizeZero(t *testing.T) {
	tree := boostTree()
	p := NewPagination(tree, boostCfg())
	p.SetPageSize(0)
	_ = p.TotalPages()
}

// === Content scrollbar visible ===

func TestContentScrollbar(t *testing.T) {
	tree := boostTree()
	c := NewContent(tree, boostCfg())
	boostLayout(tree, c, 0, 0, 200, 200)
	c.SetContentHeight(500) // trigger scrollbar
	c.Draw(render.NewCommandBuffer())
}

// === Select open dropdown ===

func TestSelectDrawOpenDropdown(t *testing.T) {
	tree := boostTree()
	cfg := boostCfgText()
	opts := []SelectOption{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}
	s := NewSelect(tree, opts, cfg)
	s.SetValue("a")
	s.open = true
	boostLayout(tree, s, 0, 0, 200, 40)
	s.Draw(render.NewCommandBuffer())
}

// === Splitter Draw ===

func TestSplitterDraw(t *testing.T) {
	tree := boostTree()
	sp := NewSplitter(tree, boostCfg())
	boostLayout(tree, sp, 0, 0, 600, 400)
	sp.Draw(render.NewCommandBuffer())
}

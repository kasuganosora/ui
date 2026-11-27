package widget

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// This file covers Apple §9 (soft boundaries). It deliberately tests both the
// shared physics with explicit frame times AND the real widget entry points,
// because a curve that looks correct in isolation says nothing about whether
// the framework's layout pass erases the overscroll every frame.

func TestRubberbandCurve(t *testing.T) {
	const dim = float32(400)
	if got := rubberband(0, dim); got != 0 {
		t.Fatalf("zero overshoot should not move, got %f", got)
	}
	if got := rubberband(80, 0); got != 0 {
		t.Fatalf("a zero-sized viewport must not stretch, got %f", got)
	}

	// Small drags still follow the gesture almost 1:1 (scaled by the constant).
	if got := rubberband(1, dim); got < 0.5 || got > 0.56 {
		t.Errorf("small overshoot should be near-linear, got %f", got)
	}

	// Resistance grows with distance: every extra pixel of drag buys less
	// movement than the previous one, and the surface never leaves the
	// viewport-sized asymptote.
	prevStep := math.MaxFloat64
	for over := float32(10); over <= 1010; over += 10 {
		step := float64(rubberband(over+10, dim) - rubberband(over, dim))
		if step <= 0 {
			t.Fatalf("overshoot %f: displacement stopped increasing", over)
		}
		if step >= prevStep {
			t.Errorf("overshoot %f: marginal movement %.4f did not shrink (prev %.4f)", over, step, prevStep)
		}
		prevStep = step
	}
	for _, over := range []float32{100, 400, 4000, 40000} {
		if got := rubberband(over, dim); got >= dim {
			t.Errorf("overshoot %f: displacement %f escaped the viewport asymptote", over, got)
		}
	}
	if rubberband(40000, dim) <= rubberband(1000, dim) {
		t.Error("large drags should still move slightly further than small ones")
	}

	for _, over := range []float32{7, 120, 3300} {
		if rubberband(-over, dim) != -rubberband(over, dim) {
			t.Errorf("rubberband is not symmetric at %f", over)
		}
	}
}

func TestScrollEdgeDampingGrowsAcrossTicks(t *testing.T) {
	var e ScrollEdge
	pos := float32(0)
	prevStep := math.MaxFloat64
	for i := 0; i < 8; i++ {
		// 40px per wheel tick, 16ms apart = one continuous gesture.
		next := applyScrollGestureAt(&e, pos, -40, 0, 600, 400, int64(1000+i*16), true)
		step := math.Abs(float64(next - pos))
		if step <= 0 {
			t.Fatalf("tick %d stopped hard at the edge", i)
		}
		if step >= prevStep {
			t.Errorf("tick %d: movement %.3f did not shrink versus %.3f", i, step, prevStep)
		}
		if next <= -400 {
			t.Errorf("tick %d: stretched past one viewport: %f", i, next)
		}
		prevStep = step
		pos = next
	}
}

func TestScrollEdgeReleaseSpring(t *testing.T) {
	var e ScrollEdge
	pos := applyScrollGestureAt(&e, 0, -4000, 0, 600, 400, 1000, true)
	if pos >= 0 {
		t.Fatalf("expected overscroll above the top edge, got %f", pos)
	}

	// Ticks are still within the gesture gap: the surface must not be pulled
	// back underneath the user's finger.
	if next, changed := settleScrollGestureAt(&e, pos, 0, 600, 400, 1100); changed || next != pos {
		t.Fatalf("settle fired during an active gesture: pos %f -> %f", pos, next)
	}

	prev := math.Abs(float64(pos))
	for ms := int64(1200); ms <= 3200; ms += 50 {
		if next, changed := settleScrollGestureAt(&e, pos, 0, 600, 400, ms); changed {
			pos = next
		}
		cur := math.Abs(float64(pos))
		if cur > prev+0.01 {
			t.Fatalf("release spring moved away from the boundary at %dms: %f -> %f", ms, prev, cur)
		}
		prev = cur
	}
	if math.Abs(float64(pos)) > 0.01 {
		t.Errorf("surface never returned to the boundary, final pos %f", pos)
	}
	if e.Overscroll(0, 600) != 0 {
		t.Errorf("overscroll state left behind: %f", e.Overscroll(0, 600))
	}
}

func TestScrollEdgeInputInterruptsSettle(t *testing.T) {
	var e ScrollEdge
	pos := applyScrollGestureAt(&e, 0, -2000, 0, 600, 400, 1000, true)
	pos, _ = settleScrollGestureAt(&e, pos, 0, 600, 400, 1300)
	before := pos

	// A new wheel tick arrives mid-settle. It must continue from the current
	// value instead of replaying the old settle curve.
	pos = applyScrollGestureAt(&e, pos, -40, 0, 600, 400, 1310, true)
	if pos >= before {
		t.Errorf("new drag did not continue from the interrupted value: %f -> %f", before, pos)
	}
	if next, changed := settleScrollGestureAt(&e, pos, 0, 600, 400, 1320); changed || next != pos {
		t.Errorf("settle replayed instead of waiting for the new gesture to end")
	}
}

func TestScrollEdgeReducedMotionKeepsHardBoundary(t *testing.T) {
	var e ScrollEdge
	if got := applyScrollGestureAt(&e, 0, -4000, 0, 600, 400, 1000, false); got != 0 {
		t.Errorf("reduced motion should keep a hard boundary, got %f", got)
	}
	if e.Overscroll(0, 600) != 0 {
		t.Error("reduced motion left overscroll state behind")
	}
}

// scrollSurface lets the same assertions run against every widget that owns a
// continuous scroll surface, without hiding their real HandleWheel entry point.
type scrollSurface struct {
	name  string
	pos   *float32
	edge  *ScrollEdge
	max   float32
	dim   float32
	wheel func()
	set   func(float32)
}

func (s scrollSurface) settleAt(nowMS int64) {
	if next, changed := settleScrollGestureAt(s.edge, *s.pos, 0, s.max, s.dim, nowMS); changed {
		*s.pos = next
	}
}

func softBoundarySurfaces(t *testing.T) []scrollSurface {
	t.Helper()
	bounds := uimath.NewRect(0, 0, 800, 400)
	place := func(id core.ElementID, tree *core.Tree) {
		tree.SetLayout(id, core.LayoutResult{Bounds: bounds})
	}

	tree1 := newTestTree()
	c := NewContent(tree1, nil)
	place(c.ElementID(), tree1)
	c.SetContentHeight(1000)

	tree2 := newTestTree()
	rv := NewRecyclerView(tree2, nil)
	place(rv.ElementID(), tree2)
	rv.contentHeight = 1000

	tree3 := newTestTree()
	d := NewDiv(tree3, nil)
	place(d.ElementID(), tree3)
	d.SetScrollable(true)
	d.SetContentHeight(1000)

	tree4 := newTestTree()
	dt := NewDataTable(tree4, []TableColumn{{Title: "a", Width: 120}}, nil)
	place(dt.ElementID(), tree4)
	dt.SetRows(make([][]string, 50))

	tree5 := newTestTree()
	vl := NewVirtualList(tree5, nil)
	place(vl.ElementID(), tree5)
	vl.SetItemCount(50)
	vl.SetItemHeight(40)

	tree6 := newTestTree()
	vg := NewVirtualGrid(tree6, 2, 50, nil)
	place(vg.ElementID(), tree6)

	tree7 := newTestTree()
	l := NewList(tree7, nil)
	place(l.ElementID(), tree7)
	items := make([]ListItem, 20)
	for i := range items {
		items[i] = ListItem{Title: "row"}
	}
	l.SetItems(items)

	tree8 := newTestTree()
	ep := NewEmojiPicker(tree8, nil)
	place(ep.ElementID(), tree8)

	return []scrollSurface{
		{"Content", &c.scrollY, &c.edge, c.maxScroll(), 400,
			func() { c.HandleWheel(100) }, func(v float32) { c.ScrollTo(v) }},
		{"RecyclerView", &rv.scrollY, &rv.edge, rv.maxScroll(), 400,
			func() { rv.HandleWheel(100) }, func(v float32) { rv.ScrollTo(v) }},
		{"Div", &d.scrollY, &d.edge, d.maxScroll(), 400,
			func() { d.HandleWheel(100) }, func(v float32) { d.ScrollTo(0, v) }},
		{"DataTable", &dt.scrollY, &dt.edge, dt.maxScrollY(), dt.scrollDimension(),
			func() {
				for i := 0; i < 100; i++ {
					dt.HandleWheel(-1)
				}
			}, func(v float32) { dt.SetScrollY(v) }},
		{"VirtualList", &vl.scrollY, &vl.edge, vl.maxScroll(), 400,
			func() { vl.HandleWheel(-4000) }, func(v float32) { vl.SetScrollY(v) }},
		{"VirtualGrid", &vg.scrollY, &vg.edge, vg.maxScroll(), 400,
			func() { vg.HandleWheel(-4000) }, func(v float32) { vg.SetScrollY(v) }},
		{"List", &l.scrollY, &l.edge, l.maxScroll(), 400,
			func() { l.HandleWheel(100) }, func(v float32) { l.SetScrollY(v) }},
		{"EmojiPicker", &ep.scroll, &ep.edge, ep.maxScroll(), 400,
			func() { ep.HandleWheel(-4000) }, func(v float32) { ep.SetScrollY(v) }},
	}
}

func TestSoftBoundaryOnEveryScrollSurface(t *testing.T) {
	for _, s := range softBoundarySurfaces(t) {
		s := s
		t.Run(s.name, func(t *testing.T) {
			if s.max < 0 || s.dim <= 0 {
				t.Fatalf("invalid fixture bounds (max=%f dim=%f)", s.max, s.dim)
			}
			s.wheel()
			// HandleWheel uses the wall clock; pin the gesture start so the
			// deterministic frame times below belong to the same timeline.
			s.edge.lastInput = 1000
			pos := *s.pos
			if pos >= 0 && pos <= s.max {
				t.Fatalf("hard stop at the edge: pos=%f, expected a damped overscroll", pos)
			}
			visibleOvershoot := math.Abs(float64(clampScrollValue(pos, 0, s.max) - pos))
			if visibleOvershoot <= 1 || visibleOvershoot >= float64(s.dim) {
				t.Errorf("overscroll %.2f is not progressively damped (viewport %.0f)", visibleOvershoot, s.dim)
			}
			if s.edge.Overscroll(0, s.max) == 0 {
				t.Error("gesture state was not recorded")
			}

			// Explicit frame times: the release spring must bring it home.
			for ms := int64(1200); ms <= 4000; ms += 50 {
				s.settleAt(ms)
			}
			if math.Abs(float64(*s.pos)) > 0.01 {
				t.Errorf("never settled back to the boundary, final pos %f", *s.pos)
			}
			if s.edge.Overscroll(0, s.max) != 0 {
				t.Error("overscroll state left behind after settle")
			}

			// Programmatic scrolls remain exact and end the gesture.
			target := s.max / 2
			s.set(target)
			if math.Abs(float64(*s.pos-target)) > 0.001 {
				t.Errorf("programmatic scroll did not land exactly: got %f want %f", *s.pos, target)
			}
			if s.edge.Overscroll(0, s.max) != 0 {
				t.Error("programmatic scroll did not reset the gesture")
			}
		})
	}
}

func TestLayoutReconcileDoesNotEraseOverscroll(t *testing.T) {
	tree := newTestTree()
	c := NewContent(tree, nil)
	tree.SetLayout(c.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 800, 400)})
	c.SetContentHeight(1000)
	c.HandleWheel(100)
	before := c.ScrollY()

	// Layout callbacks call SetContentHeight every frame with the same value.
	for i := 0; i < 5; i++ {
		c.SetContentHeight(1000)
	}
	if c.ScrollY() != before {
		t.Fatalf("layout pass erased the overscroll: %f -> %f", before, c.ScrollY())
	}

	// A real content-size change legitimately ends the gesture.
	c.SetContentHeight(400)
	if c.ScrollY() != 0 || c.edge.Overscroll(0, 0) != 0 {
		t.Errorf("content change did not reconcile the range: pos=%f", c.ScrollY())
	}
}

func TestScrollbarThumbStaysInTrackWhileStretching(t *testing.T) {
	tree := newTestTree()
	c := NewContent(tree, nil)
	tree.SetLayout(c.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 800, 400)})
	c.SetContentHeight(1000)

	c.HandleWheel(100)
	y, h := c.thumbRect(c.Bounds())
	if y < c.Bounds().Y+2-0.001 || y+h > c.Bounds().Y+c.Bounds().Height+0.001 {
		t.Errorf("thumb escaped its track during overscroll: y=%f h=%f", y, h)
	}

	c.ScrollTo(600)
	c.HandleWheel(-100)
	y, h = c.thumbRect(c.Bounds())
	if y < c.Bounds().Y+2-0.001 || y+h > c.Bounds().Y+c.Bounds().Height+0.001 {
		t.Errorf("thumb escaped its track during bottom overscroll: y=%f h=%f", y, h)
	}
}

func TestRealFramesSpringBack(t *testing.T) {
	tree := newTestTree()
	c := NewContent(tree, nil)
	tree.SetLayout(c.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 800, 400)})
	c.SetContentHeight(1000)
	c.HandleWheel(100)
	if c.ScrollY() >= 0 {
		t.Fatal("fixture did not overscroll")
	}

	// Wall-clock frames through the real Draw() path: this is the evidence
	// that the release spring is actually driven by the render loop.
	time.Sleep(time.Duration(rubberGestureGapMS)*time.Millisecond + 20*time.Millisecond)
	buf := render.NewCommandBuffer()
	for i := 0; i < 24; i++ {
		buf.Reset()
		c.Draw(buf)
		time.Sleep(40 * time.Millisecond)
	}
	if math.Abs(float64(c.ScrollY())) > 2 {
		t.Errorf("real frames never sprang back, final pos %f", c.ScrollY())
	}
}

func TestSoftBoundarySourceGuard(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	gestureFiles := 0
	sharedAdopters := 0 // count of widget types wired to the shared physics
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		text := string(src)
		sharedAdopters += strings.Count(text, "applyScrollGesture(&")
		// Wheel-to-zoom canvases have no scroll boundary and are out of §9's
		// scope; every other wheel handler must use the shared physics.
		zoomOnly := f == "camera.go" || f == "flow.go"
		if strings.Contains(text, "event.MouseWheel") {
			gestureFiles++
			if !zoomOnly && !strings.Contains(text, "HandleWheel(") {
				t.Errorf("%s handles wheel events without the shared soft-boundary entry point", f)
			}
		}
		// Caret-reveal / clear-message clamps are not user gestures and are
		// intentionally excluded from §9.
		discrete := f == "editor.go" || f == "richeditor.go" || f == "chat.go" ||
			f == "terminal.go" || f == "formula.go" || f == "diagram.go" || f == "game/chat.go"
		if !discrete && strings.Contains(text, "scrollY = max") {
			t.Errorf("%s still hard-clamps a scroll boundary", f)
		}
	}
	if gestureFiles < 4 {
		t.Errorf("source guard only saw %d wheel-handling files", gestureFiles)
	}
	if sharedAdopters < 8 {
		t.Errorf("only %d widgets adopt the shared scroll physics", sharedAdopters)
	}
}

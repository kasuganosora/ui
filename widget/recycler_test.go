package widget

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

func TestRecyclerView_BasicReconcile(t *testing.T) {
	tree := core.NewTree()
	cfg := DefaultConfig()
	rv := NewRecyclerView(tree, cfg)

	// Set layout bounds (simulate viewport)
	tree.SetLayout(rv.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 400, 300),
	})

	created := 0
	destroyed := 0

	rv.SetItemCount(1000)
	rv.SetEstimatedItemHeight(50)
	rv.SetCreateItem(func(index int) Widget {
		created++
		return NewDiv(tree, cfg)
	})
	rv.SetDestroyItem(func(index int, w Widget) {
		destroyed++
		w.Destroy()
	})

	rv.Reconcile()

	// With viewport 300px and item height 50, ~6 visible + 3 buffer each side = ~12
	if rv.ActiveCount() == 0 {
		t.Fatal("expected active widgets after reconcile")
	}
	if rv.ActiveCount() > 20 {
		t.Errorf("too many active widgets: %d", rv.ActiveCount())
	}
	t.Logf("initial: created=%d active=%d", created, rv.ActiveCount())

	// Scroll down
	prevCreated := created
	rv.ScrollTo(5000) // scroll far down
	rv.Reconcile()

	// Old widgets should be destroyed, new ones created
	if destroyed == 0 {
		t.Error("expected some widgets to be destroyed after scroll")
	}
	if created == prevCreated {
		t.Error("expected new widgets to be created after scroll")
	}
	t.Logf("after scroll: created=%d destroyed=%d active=%d", created, destroyed, rv.ActiveCount())
}

func TestRecyclerView_MemoryBounded(t *testing.T) {
	tree := core.NewTree()
	cfg := DefaultConfig()
	rv := NewRecyclerView(tree, cfg)

	tree.SetLayout(rv.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 400, 600),
	})

	rv.SetItemCount(100000) // 100K items
	rv.SetEstimatedItemHeight(80)
	rv.SetCreateItem(func(index int) Widget {
		return NewDiv(tree, cfg)
	})
	rv.SetDestroyItem(func(index int, w Widget) {
		w.Destroy()
	})

	// Simulate scrolling through entire list
	maxActive := 0
	for scroll := float32(0); scroll < 50000; scroll += 1000 {
		rv.ScrollTo(scroll)
		rv.Reconcile()
		if rv.ActiveCount() > maxActive {
			maxActive = rv.ActiveCount()
		}
	}

	// Active count should be bounded regardless of scroll position
	if maxActive > 30 {
		t.Errorf("active widgets not bounded: max=%d (expected <30)", maxActive)
	}
	t.Logf("100K items, max active widgets: %d", maxActive)
}

func TestRecyclerView_VariableHeight(t *testing.T) {
	tree := core.NewTree()
	cfg := DefaultConfig()
	rv := NewRecyclerView(tree, cfg)

	tree.SetLayout(rv.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 400, 300),
	})

	rv.SetItemCount(100)
	rv.SetEstimatedItemHeight(50)
	// Every 3rd item is tall
	rv.SetHeightFn(func(i int) float32 {
		if i%3 == 0 {
			return 120
		}
		return 50
	})
	rv.SetCreateItem(func(index int) Widget {
		return NewDiv(tree, cfg)
	})
	rv.SetDestroyItem(func(index int, w Widget) {
		w.Destroy()
	})

	rv.Reconcile()

	if rv.ActiveCount() == 0 {
		t.Fatal("expected active widgets")
	}
	if rv.ContentHeight() <= 0 {
		t.Fatal("expected positive content height")
	}

	// Content height should account for variable heights
	expectedH := float32(0)
	for i := 0; i < 100; i++ {
		if i%3 == 0 {
			expectedH += 120
		} else {
			expectedH += 50
		}
	}
	if rv.ContentHeight() != expectedH {
		t.Errorf("content height: got %f, want %f", rv.ContentHeight(), expectedH)
	}
}

func TestRecyclerView_NearEndCallback(t *testing.T) {
	tree := core.NewTree()
	cfg := DefaultConfig()
	rv := NewRecyclerView(tree, cfg)

	tree.SetLayout(rv.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 400, 300),
	})

	rv.SetItemCount(20)
	rv.SetEstimatedItemHeight(100)
	rv.SetCreateItem(func(index int) Widget {
		return NewDiv(tree, cfg)
	})

	nearEndCalled := false
	rv.SetOnNearEnd(func() {
		nearEndCalled = true
	})

	// Scroll to near bottom
	rv.ScrollTo(rv.ContentHeight() - 300 - 100) // within 200px threshold
	rv.Reconcile()

	if !nearEndCalled {
		t.Error("expected onNearEnd to be called")
	}
}

func TestRecyclerView_DestroyAll(t *testing.T) {
	tree := core.NewTree()
	cfg := DefaultConfig()
	rv := NewRecyclerView(tree, cfg)

	tree.SetLayout(rv.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 400, 300),
	})

	destroyed := 0
	rv.SetItemCount(50)
	rv.SetEstimatedItemHeight(50)
	rv.SetCreateItem(func(index int) Widget {
		return NewDiv(tree, cfg)
	})
	rv.SetDestroyItem(func(index int, w Widget) {
		destroyed++
		w.Destroy()
	})

	rv.Reconcile()
	active := rv.ActiveCount()
	rv.DestroyAll()

	if rv.ActiveCount() != 0 {
		t.Errorf("expected 0 active after DestroyAll, got %d", rv.ActiveCount())
	}
	if destroyed != active {
		t.Errorf("destroyed %d but had %d active", destroyed, active)
	}
}

func TestRecyclerView_Draw(t *testing.T) {
	tree := core.NewTree()
	cfg := DefaultConfig()
	rv := NewRecyclerView(tree, cfg)

	tree.SetLayout(rv.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 400, 300),
	})

	rv.SetItemCount(100)
	rv.SetEstimatedItemHeight(50)
	rv.SetCreateItem(func(index int) Widget {
		d := NewDiv(tree, cfg)
		d.SetBgColor(uimath.ColorHex("#336699"))
		return d
	})
	rv.SetDestroyItem(func(index int, w Widget) {
		w.Destroy()
	})

	rv.Reconcile()

	buf := &render.CommandBuffer{}
	rv.Draw(buf)

	cmds := buf.Commands()
	if len(cmds) == 0 {
		t.Error("expected render commands from Draw")
	}
}

func BenchmarkRecyclerView_Reconcile(b *testing.B) {
	tree := core.NewTree()
	cfg := DefaultConfig()
	rv := NewRecyclerView(tree, cfg)

	tree.SetLayout(rv.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 400, 800),
	})

	rv.SetItemCount(10000)
	rv.SetEstimatedItemHeight(80)
	rv.SetCreateItem(func(index int) Widget {
		return NewDiv(tree, cfg)
	})
	rv.SetDestroyItem(func(index int, w Widget) {
		w.Destroy()
	})

	rv.Reconcile() // warm up

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rv.ScrollTo(float32(i%8000) * 10)
		rv.Reconcile()
	}
}

func BenchmarkRecyclerView_Draw(b *testing.B) {
	tree := core.NewTree()
	cfg := DefaultConfig()
	rv := NewRecyclerView(tree, cfg)

	tree.SetLayout(rv.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 400, 800),
	})

	rv.SetItemCount(10000)
	rv.SetEstimatedItemHeight(80)
	rv.SetCreateItem(func(index int) Widget {
		d := NewDiv(tree, cfg)
		d.SetBgColor(uimath.ColorHex("#336699"))
		return d
	})
	rv.SetDestroyItem(func(index int, w Widget) {
		w.Destroy()
	})

	rv.Reconcile()

	buf := &render.CommandBuffer{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		rv.Draw(buf)
	}
}

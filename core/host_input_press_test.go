package core

import (
	"testing"

	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
)

// §1: the press state is owned by the tree so every renderer can respond on
// pointer-down. These tests pin the input contract itself, independent of any
// widget or backend.

func newPressFixture() (*Tree, ElementID, *int) {
	tree := NewTree()
	// HitTest prunes any subtree whose root bounds do not contain the point,
	// so the fixture must lay out the tree root as well as the control.
	tree.SetLayout(tree.Root(), LayoutResult{Bounds: uimath.NewRect(0, 0, 600, 600)})
	id := tree.CreateElement(TypeButton)
	tree.AppendChild(tree.Root(), id)
	tree.SetLayout(id, LayoutResult{Bounds: uimath.NewRect(0, 0, 100, 40)})
	clicks := 0
	tree.AddHandler(id, event.MouseClick, func(*event.Event) { clicks++ })
	return tree, id, &clicks
}

func downAt(x, y float32) *event.Event {
	return &event.Event{Type: event.MouseDown, GlobalX: x, GlobalY: y}
}

func upAt(x, y float32) *event.Event {
	return &event.Event{Type: event.MouseUp, GlobalX: x, GlobalY: y}
}

func TestHostInputPublishesPressOnDown(t *testing.T) {
	tree, id, clicks := newPressFixture()
	h := HostInput{Tree: tree}

	h.Handle(downAt(10, 10))
	if tree.Pressed() != id {
		t.Fatalf("press not published on pointer-down: got %d want %d", tree.Pressed(), id)
	}
	if *clicks != 0 {
		t.Error("action fired before release; the click contract must stay on MouseUp")
	}

	h.Handle(upAt(10, 10))
	if tree.Pressed() != InvalidElementID {
		t.Errorf("press highlight not released: %d", tree.Pressed())
	}
	if *clicks != 1 {
		t.Errorf("expected exactly one synthesized click, got %d", *clicks)
	}
}

func TestHostInputIgnoresInertPress(t *testing.T) {
	tree := NewTree()
	tree.SetLayout(tree.Root(), LayoutResult{Bounds: uimath.NewRect(0, 0, 600, 600)})
	id := tree.CreateElement(TypeDiv)
	tree.AppendChild(tree.Root(), id)
	tree.SetLayout(id, LayoutResult{Bounds: uimath.NewRect(0, 0, 100, 40)})
	h := HostInput{Tree: tree}

	h.Handle(downAt(10, 10))
	if tree.Pressed() != InvalidElementID {
		t.Errorf("an inert container must not be highlighted, got %d", tree.Pressed())
	}
	h.Handle(upAt(10, 10))
}

func TestHostInputInteractiveTargetWalksToAncestor(t *testing.T) {
	tree, parent, _ := newPressFixture()
	child := tree.CreateElement(TypeText)
	tree.AppendChild(parent, child)
	tree.SetLayout(child, LayoutResult{Bounds: uimath.NewRect(10, 10, 30, 20)})
	if got := tree.InteractiveTarget(child); got != parent {
		t.Fatalf("interactive ancestor not found: got %d want %d", got, parent)
	}

	h := HostInput{Tree: tree}
	h.Handle(downAt(20, 20))
	if tree.Pressed() != parent {
		t.Errorf("press highlight should cover the whole control: got %d want %d", tree.Pressed(), parent)
	}
	h.Handle(upAt(20, 20))
}

func TestHostInputReleasesPressOnCancelAndBlur(t *testing.T) {
	for _, cancel := range []event.Type{event.TouchCancel, event.WindowBlur} {
		tree, _, clicks := newPressFixture()
		h := HostInput{Tree: tree}
		h.Handle(downAt(10, 10))
		if tree.Pressed() == InvalidElementID {
			t.Fatalf("%v: fixture did not press", cancel)
		}
		h.Handle(&event.Event{Type: cancel})
		if tree.Pressed() != InvalidElementID {
			t.Errorf("%v left a stale press highlight", cancel)
		}
		if *clicks != 0 {
			t.Errorf("%v synthesized an unwanted click", cancel)
		}
	}
}

func TestHostInputDragOffCancelsClickAndPress(t *testing.T) {
	tree, _, clicks := newPressFixture()
	h := HostInput{Tree: tree}
	h.Handle(downAt(10, 10))
	h.Handle(upAt(500, 500))
	if *clicks != 0 {
		t.Error("dragging off the control must not activate it")
	}
	if tree.Pressed() != InvalidElementID {
		t.Error("dragging off the control must clear the press highlight")
	}
}

func TestSetPressedMarksBothTargetsDirty(t *testing.T) {
	tree := NewTree()
	a := tree.CreateElement(TypeButton)
	b := tree.CreateElement(TypeButton)
	for _, id := range []ElementID{a, b} {
		tree.AppendChild(tree.Root(), id)
		tree.SetLayout(id, LayoutResult{Bounds: uimath.NewRect(0, 0, 10, 10)})
	}
	tree.ClearAllDirty()

	tree.SetPressed(a)
	if !tree.Get(a).IsDirty(DirtyPaint) {
		t.Error("newly pressed element was not marked dirty")
	}
	tree.ClearAllDirty()
	tree.SetPressed(b)
	if !tree.Get(a).IsDirty(DirtyPaint) || !tree.Get(b).IsDirty(DirtyPaint) {
		t.Error("switching press targets must redraw both the old and new element")
	}
}

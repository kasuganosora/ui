package core

import (
	"testing"

	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
)

// Additional tests to bring core coverage to 80%+.

func TestElementClasses(t *testing.T) {
	tree := NewTree()
	id := tree.CreateElement(TypeButton)
	tree.SetClasses(id, []string{"btn", "primary"})
	elem := tree.Get(id)
	classes := elem.Classes()
	if len(classes) != 2 || classes[0] != "btn" || classes[1] != "primary" {
		t.Errorf("unexpected classes: %v", classes)
	}
}

func TestElementLayout(t *testing.T) {
	tree := NewTree()
	id := tree.CreateElement(TypeDiv)
	layout := LayoutResult{
		Bounds: uimath.NewRect(0, 0, 200, 100),
	}
	tree.SetLayout(id, layout)
	elem := tree.Get(id)
	if elem.Layout().Bounds.Width != 200 {
		t.Errorf("expected width 200, got %v", elem.Layout().Bounds.Width)
	}
}

func TestElementEnabled(t *testing.T) {
	tree := NewTree()
	id := tree.CreateElement(TypeButton)
	elem := tree.Get(id)
	if !elem.IsEnabled() {
		t.Error("element should be enabled by default")
	}
	tree.SetEnabled(id, false)
	if elem.IsEnabled() {
		t.Error("element should be disabled")
	}
}

func TestElementHovered(t *testing.T) {
	tree := NewTree()
	id := tree.CreateElement(TypeButton)
	elem := tree.Get(id)
	if elem.IsHovered() {
		t.Error("element should not be hovered by default")
	}
	tree.SetHovered(id, true)
	if !elem.IsHovered() {
		t.Error("element should be hovered")
	}
}

func TestElementProperty(t *testing.T) {
	tree := NewTree()
	id := tree.CreateElement(TypeText)
	tree.SetProperty(id, "color", "red")
	elem := tree.Get(id)
	v, ok := elem.Property("color")
	if !ok || v != "red" {
		t.Errorf("expected 'red', got %v", v)
	}
	_, ok = elem.Property("nonexistent")
	if ok {
		t.Error("should return false for nonexistent property")
	}
}

func TestTextContentNonString(t *testing.T) {
	tree := NewTree()
	id := tree.CreateElement(TypeText)
	tree.SetProperty(id, "text", 42) // not a string
	elem := tree.Get(id)
	if elem.TextContent() != "" {
		t.Errorf("expected empty string for non-string text, got %q", elem.TextContent())
	}
}

func TestTextContentMissing(t *testing.T) {
	tree := NewTree()
	id := tree.CreateElement(TypeText)
	elem := tree.Get(id)
	if elem.TextContent() != "" {
		t.Errorf("expected empty string for missing text, got %q", elem.TextContent())
	}
}

func TestAppendChildInvalidIDs(t *testing.T) {
	tree := NewTree()
	ok := tree.AppendChild(999, 888) // neither exists
	if ok {
		t.Error("AppendChild with invalid IDs should fail")
	}
}

func TestRemoveChildInvalidIDs(t *testing.T) {
	tree := NewTree()
	ok := tree.RemoveChild(999, 888)
	if ok {
		t.Error("RemoveChild with invalid IDs should fail")
	}
}

func TestInsertBeforeInvalidIDs(t *testing.T) {
	tree := NewTree()
	ok := tree.InsertBefore(999, 888, 777)
	if ok {
		t.Error("InsertBefore with invalid IDs should fail")
	}
}

func TestInsertBeforeNotFound(t *testing.T) {
	tree := NewTree()
	child := tree.CreateElement(TypeDiv)
	// Insert before a nonexistent beforeID => appends
	ok := tree.InsertBefore(tree.Root(), child, 999)
	if !ok {
		t.Error("InsertBefore with missing beforeID should append")
	}
	root := tree.Get(tree.Root())
	if len(root.ChildIDs()) != 1 || root.ChildIDs()[0] != child {
		t.Errorf("child should be appended, got %v", root.ChildIDs())
	}
}

func TestAppendChildReparent(t *testing.T) {
	tree := NewTree()
	a := tree.CreateElement(TypeDiv)
	b := tree.CreateElement(TypeDiv)
	child := tree.CreateElement(TypeText)
	tree.AppendChild(tree.Root(), a)
	tree.AppendChild(tree.Root(), b)
	tree.AppendChild(a, child)

	// Reparent child from a to b
	tree.AppendChild(b, child)
	aElem := tree.Get(a)
	bElem := tree.Get(b)
	if len(aElem.ChildIDs()) != 0 {
		t.Error("a should have no children after reparent")
	}
	if len(bElem.ChildIDs()) != 1 {
		t.Error("b should have 1 child after reparent")
	}
}

func TestDestroyElementNonexistent(t *testing.T) {
	tree := NewTree()
	// Should not panic
	tree.DestroyElement(999)
}

func TestHandlersNilElement(t *testing.T) {
	tree := NewTree()
	h := tree.Handlers(999, event.MouseClick)
	if h != nil {
		t.Error("Handlers for nonexistent element should be nil")
	}
}

func TestSetPropertyNilElement(t *testing.T) {
	tree := NewTree()
	// Should not panic
	tree.SetProperty(999, "key", "value")
}

func TestSetClassesNilElement(t *testing.T) {
	tree := NewTree()
	// Should not panic
	tree.SetClasses(999, []string{"a"})
}

func TestSetVisibleNilElement(t *testing.T) {
	tree := NewTree()
	// Should not panic
	tree.SetVisible(999, false)
}

func TestSetEnabledNilElement(t *testing.T) {
	tree := NewTree()
	tree.SetEnabled(999, false)
}

func TestSetFocusedNilElement(t *testing.T) {
	tree := NewTree()
	tree.SetFocused(999, true)
}

func TestSetHoveredNilElement(t *testing.T) {
	tree := NewTree()
	tree.SetHovered(999, true)
}

func TestSetLayoutNilElement(t *testing.T) {
	tree := NewTree()
	tree.SetLayout(999, LayoutResult{})
}

func TestAddHandlerNilElement(t *testing.T) {
	tree := NewTree()
	tree.AddHandler(999, event.MouseClick, func(e *event.Event) {})
}

func TestClearDirtyNilElement(t *testing.T) {
	tree := NewTree()
	tree.ClearDirty(999, DirtyAll)
}

func TestHitTestInvisibleChild(t *testing.T) {
	tree := NewTree()
	child := tree.CreateElement(TypeButton)
	tree.AppendChild(tree.Root(), child)
	tree.SetLayout(tree.Root(), LayoutResult{
		Bounds: uimath.NewRect(0, 0, 800, 600),
	})
	tree.SetLayout(child, LayoutResult{
		Bounds: uimath.NewRect(100, 100, 200, 50),
	})
	tree.SetVisible(child, false)

	hit := tree.HitTest(150, 120)
	if hit == child {
		t.Error("invisible child should not be hit")
	}
}

func TestWalkSkipSubtree(t *testing.T) {
	tree := NewTree()
	a := tree.CreateElement(TypeDiv)
	b := tree.CreateElement(TypeText) // child of a
	tree.AppendChild(tree.Root(), a)
	tree.AppendChild(a, b)

	var visited []ElementID
	tree.Walk(tree.Root(), func(id ElementID, depth int) bool {
		visited = append(visited, id)
		if id == a {
			return false // skip a's children
		}
		return true
	})
	// Should visit root, a, but not b
	for _, v := range visited {
		if v == b {
			t.Error("b should not be visited when a's subtree is skipped")
		}
	}
}

func TestRemoveChildNotActualChild(t *testing.T) {
	tree := NewTree()
	a := tree.CreateElement(TypeDiv)
	b := tree.CreateElement(TypeDiv)
	tree.AppendChild(tree.Root(), a)
	// b is not a child of root's direct children
	ok := tree.RemoveChild(tree.Root(), b)
	if ok {
		t.Error("RemoveChild should fail for non-child")
	}
}

func TestSetVisibleSameValue(t *testing.T) {
	tree := NewTree()
	id := tree.CreateElement(TypeDiv)
	tree.ClearDirty(id, DirtyAll)
	tree.SetVisible(id, true) // already true
	elem := tree.Get(id)
	if elem.IsDirty(DirtyPaint) {
		t.Error("setting same visibility should not dirty")
	}
}

func TestDispatchToTargetNilTarget(t *testing.T) {
	tree := NewTree()
	d := NewDispatcher(tree)
	evt := &event.Event{Type: event.MouseClick}
	result := d.DispatchToTarget(999, evt)
	if !result {
		t.Error("dispatch to nil target should return true")
	}
}

func TestDispatchNilTarget(t *testing.T) {
	tree := NewTree()
	d := NewDispatcher(tree)
	evt := &event.Event{Type: event.MouseClick}
	result := d.Dispatch(999, evt)
	if !result {
		t.Error("dispatch to nil target should return true")
	}
}

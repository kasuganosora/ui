package core

import (
	"testing"

	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
)

func TestTreeCreateAndGet(t *testing.T) {
	tree := NewTree()
	id := tree.CreateElement(TypeButton)
	elem := tree.Get(id)
	if elem == nil {
		t.Fatal("element should not be nil")
	}
	if elem.Type() != TypeButton {
		t.Errorf("expected TypeButton, got %v", elem.Type())
	}
	if elem.ID() != id {
		t.Errorf("ID mismatch")
	}
}

func TestTreeAppendChild(t *testing.T) {
	tree := NewTree()
	child := tree.CreateElement(TypeText)
	ok := tree.AppendChild(tree.Root(), child)
	if !ok {
		t.Fatal("AppendChild should succeed")
	}
	root := tree.Get(tree.Root())
	if len(root.ChildIDs()) != 1 || root.ChildIDs()[0] != child {
		t.Error("child should be appended to root")
	}
	childElem := tree.Get(child)
	if childElem.ParentID() != tree.Root() {
		t.Error("child's parent should be root")
	}
}

func TestTreeRemoveChild(t *testing.T) {
	tree := NewTree()
	child := tree.CreateElement(TypeText)
	tree.AppendChild(tree.Root(), child)
	tree.RemoveChild(tree.Root(), child)

	root := tree.Get(tree.Root())
	if len(root.ChildIDs()) != 0 {
		t.Error("root should have no children after remove")
	}
	childElem := tree.Get(child)
	if childElem.ParentID() != InvalidElementID {
		t.Error("child should have no parent after remove")
	}
}

func TestTreeDestroyElement(t *testing.T) {
	tree := NewTree()
	parent := tree.CreateElement(TypeDiv)
	child := tree.CreateElement(TypeText)
	tree.AppendChild(tree.Root(), parent)
	tree.AppendChild(parent, child)

	tree.DestroyElement(parent)

	if tree.Get(parent) != nil {
		t.Error("parent should be destroyed")
	}
	if tree.Get(child) != nil {
		t.Error("child should be destroyed recursively")
	}
}

func TestTreeSetProperty(t *testing.T) {
	tree := NewTree()
	id := tree.CreateElement(TypeText)
	tree.SetProperty(id, "text", "hello")

	elem := tree.Get(id)
	if elem.TextContent() != "hello" {
		t.Errorf("expected 'hello', got '%v'", elem.TextContent())
	}
}

func TestTreeVisibility(t *testing.T) {
	tree := NewTree()
	id := tree.CreateElement(TypeDiv)
	elem := tree.Get(id)
	if !elem.IsVisible() {
		t.Error("element should be visible by default")
	}
	tree.SetVisible(id, false)
	if elem.IsVisible() {
		t.Error("element should be hidden")
	}
}

func TestTreeFocus(t *testing.T) {
	tree := NewTree()
	id := tree.CreateElement(TypeInput)
	tree.SetFocused(id, true)
	elem := tree.Get(id)
	if !elem.IsFocused() {
		t.Error("element should be focused")
	}
}

func TestTreeDirtyPropagation(t *testing.T) {
	tree := NewTree()
	parent := tree.CreateElement(TypeDiv)
	child := tree.CreateElement(TypeText)
	tree.AppendChild(tree.Root(), parent)
	tree.AppendChild(parent, child)

	// Clear all dirty flags
	tree.ClearDirty(tree.Root(), DirtyAll)
	tree.ClearDirty(parent, DirtyAll)
	tree.ClearDirty(child, DirtyAll)

	// Modify child -> should propagate dirty up
	tree.SetProperty(child, "text", "changed")

	parentElem := tree.Get(parent)
	if !parentElem.IsDirty(DirtyPaint) {
		t.Error("parent should be dirty after child modification")
	}
	rootElem := tree.Get(tree.Root())
	if !rootElem.IsDirty(DirtyPaint) {
		t.Error("root should be dirty after descendant modification")
	}
}

func TestTreeHitTest(t *testing.T) {
	tree := NewTree()
	child := tree.CreateElement(TypeButton)
	tree.AppendChild(tree.Root(), child)

	tree.SetLayout(tree.Root(), LayoutResult{
		Bounds: uimath.NewRect(0, 0, 800, 600),
	})
	tree.SetLayout(child, LayoutResult{
		Bounds: uimath.NewRect(100, 100, 200, 50),
	})

	hit := tree.HitTest(150, 120)
	if hit != child {
		t.Errorf("expected hit on child, got %v", hit)
	}

	hit = tree.HitTest(50, 50)
	if hit != tree.Root() {
		t.Errorf("expected hit on root, got %v", hit)
	}

	hit = tree.HitTest(900, 900)
	if hit != InvalidElementID {
		t.Errorf("expected no hit, got %v", hit)
	}
}

func TestTreeWalk(t *testing.T) {
	tree := NewTree()
	a := tree.CreateElement(TypeDiv)
	b := tree.CreateElement(TypeText)
	c := tree.CreateElement(TypeButton)
	tree.AppendChild(tree.Root(), a)
	tree.AppendChild(a, b)
	tree.AppendChild(tree.Root(), c)

	var visited []ElementID
	tree.Walk(tree.Root(), func(id ElementID, depth int) bool {
		visited = append(visited, id)
		return true
	})

	if len(visited) != 4 { // root, a, b, c
		t.Errorf("expected 4 elements, visited %d", len(visited))
	}
}

func TestTreeEventHandler(t *testing.T) {
	tree := NewTree()
	id := tree.CreateElement(TypeButton)
	called := false
	tree.AddHandler(id, event.MouseClick, func(e *event.Event) {
		called = true
	})

	handlers := tree.Handlers(id, event.MouseClick)
	if len(handlers) != 1 {
		t.Fatalf("expected 1 handler, got %d", len(handlers))
	}
	handlers[0](&event.Event{})
	if !called {
		t.Error("handler should have been called")
	}
}

func TestTreeElementCount(t *testing.T) {
	tree := NewTree()
	if tree.ElementCount() != 1 { // root
		t.Errorf("expected 1, got %d", tree.ElementCount())
	}
	tree.CreateElement(TypeDiv)
	tree.CreateElement(TypeText)
	if tree.ElementCount() != 3 {
		t.Errorf("expected 3, got %d", tree.ElementCount())
	}
}

func TestTreeInsertBefore(t *testing.T) {
	tree := NewTree()
	a := tree.CreateElement(TypeDiv)
	b := tree.CreateElement(TypeDiv)
	c := tree.CreateElement(TypeDiv)
	tree.AppendChild(tree.Root(), a)
	tree.AppendChild(tree.Root(), c)
	tree.InsertBefore(tree.Root(), b, c)

	root := tree.Get(tree.Root())
	children := root.ChildIDs()
	if len(children) != 3 || children[0] != a || children[1] != b || children[2] != c {
		t.Errorf("expected [a,b,c], got %v", children)
	}
}

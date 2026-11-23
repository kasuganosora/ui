package core

import (
	"testing"

	"github.com/kasuganosora/ui/event"
)

func TestTabCyclesFocusableElements(t *testing.T) {
	tree := NewTree()
	a := tree.CreateElement(TypeButton)
	b := tree.CreateElement(TypeButton)
	tree.AppendChild(tree.Root(), a)
	tree.AppendChild(tree.Root(), b)
	tree.SetFocusable(a, true)
	tree.SetFocusable(b, true)
	tree.SetFocused(a, true)

	tree.RouteKey(&event.Event{Type: event.KeyDown, Key: event.KeyTab})
	if tree.Focused() != b {
		t.Fatalf("tab focused %d, want %d", tree.Focused(), b)
	}
	tree.RouteKey(&event.Event{Type: event.KeyDown, Key: event.KeyTab})
	if tree.Focused() != a {
		t.Fatalf("tab wrap focused %d, want %d", tree.Focused(), a)
	}
}

func TestTabSkipsDisabled(t *testing.T) {
	tree := NewTree()
	a := tree.CreateElement(TypeButton)
	b := tree.CreateElement(TypeButton)
	c := tree.CreateElement(TypeButton)
	tree.AppendChild(tree.Root(), a)
	tree.AppendChild(tree.Root(), b)
	tree.AppendChild(tree.Root(), c)
	tree.SetFocusable(a, true)
	tree.SetFocusable(b, true)
	tree.SetFocusable(c, true)
	tree.SetEnabled(b, false)
	tree.SetFocused(a, true)

	tree.RouteKey(&event.Event{Type: event.KeyDown, Key: event.KeyTab})
	if tree.Focused() != c {
		t.Fatalf("focused %d, want enabled %d", tree.Focused(), c)
	}
}

func TestFocusScopeTrapsTab(t *testing.T) {
	tree := NewTree()
	outside := tree.CreateElement(TypeButton)
	dialog := tree.CreateElement(TypeCustom)
	inside := tree.CreateElement(TypeButton)
	tree.AppendChild(tree.Root(), outside)
	tree.AppendChild(tree.Root(), dialog)
	tree.AppendChild(dialog, inside)
	tree.SetFocusable(outside, true)
	tree.SetFocusable(inside, true)
	tree.SetFocused(outside, true)
	tree.PushFocusScope(dialog)

	tree.RouteKey(&event.Event{Type: event.KeyDown, Key: event.KeyTab})
	if tree.Focused() != inside {
		t.Fatalf("scope focused %d, want %d", tree.Focused(), inside)
	}
	tree.RouteKey(&event.Event{Type: event.KeyDown, Key: event.KeyTab})
	if tree.Focused() != inside {
		t.Fatalf("scope leaked to %d", tree.Focused())
	}
	tree.PopFocusScope()
	tree.SetFocused(inside, true)
	tree.RouteKey(&event.Event{Type: event.KeyDown, Key: event.KeyTab})
	if tree.Focused() != outside {
		t.Fatalf("after pop focused %d, want %d", tree.Focused(), outside)
	}
}

func TestTabSkipsHidden(t *testing.T) {
	tree := NewTree()
	a := tree.CreateElement(TypeButton)
	b := tree.CreateElement(TypeButton)
	tree.AppendChild(tree.Root(), a)
	tree.AppendChild(tree.Root(), b)
	tree.SetFocusable(a, true)
	tree.SetFocusable(b, true)
	tree.SetVisible(a, false)
	tree.SetFocused(b, true)
	tree.RouteKey(&event.Event{Type: event.KeyDown, Key: event.KeyTab})
	if tree.Focused() != b {
		t.Fatalf("hidden stop took focus: %d", tree.Focused())
	}
}

func TestDestroyOpenScopeReleasesTab(t *testing.T) {
	tree := NewTree()
	outside := tree.CreateElement(TypeButton)
	dialog := tree.CreateElement(TypeCustom)
	inside := tree.CreateElement(TypeButton)
	tree.AppendChild(tree.Root(), outside)
	tree.AppendChild(tree.Root(), dialog)
	tree.AppendChild(dialog, inside)
	tree.SetFocusable(outside, true)
	tree.SetFocusable(inside, true)
	tree.SetFocused(inside, true)
	tree.PushFocusScope(dialog)
	tree.DestroyElement(dialog)
	tree.RouteKey(&event.Event{Type: event.KeyDown, Key: event.KeyTab})
	if tree.Focused() != outside {
		t.Fatalf("scope survived destroy, focused %d want %d", tree.Focused(), outside)
	}
}

func TestDestroyClearsFocus(t *testing.T) {
	tree := NewTree()
	a := tree.CreateElement(TypeButton)
	tree.AppendChild(tree.Root(), a)
	tree.SetFocusable(a, true)
	tree.SetFocused(a, true)
	tree.DestroyElement(a)
	if tree.Focused() != InvalidElementID {
		t.Fatalf("focused %d after destroy", tree.Focused())
	}
}

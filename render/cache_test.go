package render

import (
	"testing"
)

func TestCommandCache_NilSafety(t *testing.T) {
	var cc *CommandCache

	// All operations on nil cache should be no-ops without panic
	if got := cc.Get(1); got != nil {
		t.Errorf("Get on nil cache: got %v, want nil", got)
	}
	cc.Store(1, []Command{{Type: CmdRect}})
	cc.Invalidate(1)
	cc.InvalidateAll()
	cc.Remove(1)
	if got := cc.Len(); got != 0 {
		t.Errorf("Len on nil cache: got %d, want 0", got)
	}
	if got := cc.ValidCount(); got != 0 {
		t.Errorf("ValidCount on nil cache: got %d, want 0", got)
	}
}

func TestCommandCache_StoreAndGet(t *testing.T) {
	cc := NewCommandCache()
	var id ElementID = 42

	cmds := []Command{
		{Type: CmdRect, ZOrder: 1, Opacity: 1.0},
		{Type: CmdText, ZOrder: 2, Opacity: 0.5},
	}
	cc.Store(id, cmds)

	got := cc.Get(id)
	if got == nil {
		t.Fatal("Get returned nil after Store")
	}
	if len(got) != len(cmds) {
		t.Fatalf("Get returned %d commands, want %d", len(got), len(cmds))
	}
	for i := range cmds {
		if got[i].Type != cmds[i].Type {
			t.Errorf("command[%d].Type = %v, want %v", i, got[i].Type, cmds[i].Type)
		}
		if got[i].ZOrder != cmds[i].ZOrder {
			t.Errorf("command[%d].ZOrder = %v, want %v", i, got[i].ZOrder, cmds[i].ZOrder)
		}
		if got[i].Opacity != cmds[i].Opacity {
			t.Errorf("command[%d].Opacity = %v, want %v", i, got[i].Opacity, cmds[i].Opacity)
		}
	}
}

func TestCommandCache_StoreIsCopy(t *testing.T) {
	cc := NewCommandCache()
	var id ElementID = 1

	cmds := []Command{{Type: CmdRect, ZOrder: 10}}
	cc.Store(id, cmds)

	// Mutating original should not affect cached copy
	cmds[0].ZOrder = 99

	got := cc.Get(id)
	if got[0].ZOrder != 10 {
		t.Errorf("cached command was mutated: ZOrder = %d, want 10", got[0].ZOrder)
	}
}

func TestCommandCache_GetMiss(t *testing.T) {
	cc := NewCommandCache()
	if got := cc.Get(999); got != nil {
		t.Errorf("Get on missing ID: got %v, want nil", got)
	}
}

func TestCommandCache_Invalidate(t *testing.T) {
	cc := NewCommandCache()
	var id ElementID = 5

	cc.Store(id, []Command{{Type: CmdImage}})
	cc.Invalidate(id)

	if got := cc.Get(id); got != nil {
		t.Errorf("Get after Invalidate: got %v, want nil", got)
	}

	// Invalidate on non-existent entry should not panic
	cc.Invalidate(999)
}

func TestCommandCache_InvalidateAll(t *testing.T) {
	cc := NewCommandCache()
	cc.Store(ElementID(1), []Command{{Type: CmdRect}})
	cc.Store(ElementID(2), []Command{{Type: CmdText}})
	cc.Store(ElementID(3), []Command{{Type: CmdImage}})

	cc.InvalidateAll()

	for _, id := range []ElementID{1, 2, 3} {
		if got := cc.Get(id); got != nil {
			t.Errorf("Get(%d) after InvalidateAll: got %v, want nil", id, got)
		}
	}
}

func TestCommandCache_Remove(t *testing.T) {
	cc := NewCommandCache()
	var id ElementID = 7

	cc.Store(id, []Command{{Type: CmdClip}})
	if cc.Len() != 1 {
		t.Fatalf("Len after Store: got %d, want 1", cc.Len())
	}

	cc.Remove(id)

	if got := cc.Get(id); got != nil {
		t.Errorf("Get after Remove: got %v, want nil", got)
	}
	if cc.Len() != 0 {
		t.Errorf("Len after Remove: got %d, want 0", cc.Len())
	}

	// Remove non-existent should not panic
	cc.Remove(999)
}

func TestCommandCache_StoreOverwrite(t *testing.T) {
	cc := NewCommandCache()
	var id ElementID = 10

	cc.Store(id, []Command{{Type: CmdRect, ZOrder: 1}})
	cc.Store(id, []Command{{Type: CmdText, ZOrder: 2}, {Type: CmdImage, ZOrder: 3}})

	got := cc.Get(id)
	if len(got) != 2 {
		t.Fatalf("Get after overwrite: got %d commands, want 2", len(got))
	}
	if got[0].Type != CmdText || got[1].Type != CmdImage {
		t.Errorf("overwritten commands: got types %v/%v, want CmdText/CmdImage", got[0].Type, got[1].Type)
	}
	if cc.Len() != 1 {
		t.Errorf("Len after overwrite: got %d, want 1", cc.Len())
	}
}

func TestCommandCache_LenAndValidCount(t *testing.T) {
	cc := NewCommandCache()

	if cc.Len() != 0 || cc.ValidCount() != 0 {
		t.Fatalf("empty cache: Len=%d ValidCount=%d, want 0/0", cc.Len(), cc.ValidCount())
	}

	cc.Store(ElementID(1), []Command{{Type: CmdRect}})
	cc.Store(ElementID(2), []Command{{Type: CmdText}})
	cc.Store(ElementID(3), []Command{{Type: CmdImage}})

	if cc.Len() != 3 {
		t.Errorf("Len: got %d, want 3", cc.Len())
	}
	if cc.ValidCount() != 3 {
		t.Errorf("ValidCount: got %d, want 3", cc.ValidCount())
	}

	cc.Invalidate(ElementID(2))

	if cc.Len() != 3 {
		t.Errorf("Len after invalidate: got %d, want 3", cc.Len())
	}
	if cc.ValidCount() != 2 {
		t.Errorf("ValidCount after invalidate: got %d, want 2", cc.ValidCount())
	}

	cc.Remove(ElementID(1))

	if cc.Len() != 2 {
		t.Errorf("Len after remove: got %d, want 2", cc.Len())
	}
	if cc.ValidCount() != 1 {
		t.Errorf("ValidCount after remove: got %d, want 1", cc.ValidCount())
	}
}

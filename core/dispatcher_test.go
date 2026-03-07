package core

import (
	"testing"

	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
)

// buildTestTree creates: root → parent → child
// All elements have layout bounds so hit testing works.
func buildTestTree() (*Tree, ElementID, ElementID, ElementID) {
	tree := NewTree()
	parent := tree.CreateElement(TypeDiv)
	child := tree.CreateElement(TypeButton)
	tree.AppendChild(tree.Root(), parent)
	tree.AppendChild(parent, child)

	tree.SetLayout(tree.Root(), LayoutResult{Bounds: uimath.NewRect(0, 0, 800, 600)})
	tree.SetLayout(parent, LayoutResult{Bounds: uimath.NewRect(0, 0, 400, 300)})
	tree.SetLayout(child, LayoutResult{Bounds: uimath.NewRect(10, 10, 100, 40)})

	return tree, tree.Root(), parent, child
}

func TestDispatchReachesTarget(t *testing.T) {
	tree, _, _, child := buildTestTree()
	d := NewDispatcher(tree)

	called := false
	tree.AddHandler(child, event.MouseClick, func(e *event.Event) {
		called = true
	})

	evt := &event.Event{Type: event.MouseClick}
	d.Dispatch(child, evt)

	if !called {
		t.Error("handler on target should be called")
	}
}

func TestDispatchCapturePhase(t *testing.T) {
	tree, root, parent, child := buildTestTree()
	d := NewDispatcher(tree)

	var order []string
	tree.AddHandler(root, event.MouseClick, func(e *event.Event) {
		if e.Phase == event.PhaseCapture {
			order = append(order, "root-capture")
		}
	})
	tree.AddHandler(parent, event.MouseClick, func(e *event.Event) {
		if e.Phase == event.PhaseCapture {
			order = append(order, "parent-capture")
		}
	})
	tree.AddHandler(child, event.MouseClick, func(e *event.Event) {
		if e.Phase == event.PhaseTarget {
			order = append(order, "child-target")
		}
	})

	d.Dispatch(child, &event.Event{Type: event.MouseClick})

	if len(order) < 3 {
		t.Fatalf("expected 3 phases, got %d: %v", len(order), order)
	}
	if order[0] != "root-capture" || order[1] != "parent-capture" || order[2] != "child-target" {
		t.Errorf("unexpected order: %v", order)
	}
}

func TestDispatchBubblePhase(t *testing.T) {
	tree, root, parent, child := buildTestTree()
	d := NewDispatcher(tree)

	var order []string
	tree.AddHandler(child, event.MouseClick, func(e *event.Event) {
		order = append(order, "child-target")
	})
	tree.AddHandler(parent, event.MouseClick, func(e *event.Event) {
		if e.Phase == event.PhaseBubble {
			order = append(order, "parent-bubble")
		}
	})
	tree.AddHandler(root, event.MouseClick, func(e *event.Event) {
		if e.Phase == event.PhaseBubble {
			order = append(order, "root-bubble")
		}
	})

	d.Dispatch(child, &event.Event{Type: event.MouseClick})

	// Check bubble order appears after target
	found := false
	for i, s := range order {
		if s == "child-target" {
			if i+1 < len(order) && order[i+1] == "parent-bubble" {
				if i+2 < len(order) && order[i+2] == "root-bubble" {
					found = true
				}
			}
		}
	}
	if !found {
		t.Errorf("expected target → parent-bubble → root-bubble, got %v", order)
	}
}

func TestDispatchStopPropagation(t *testing.T) {
	tree, _, parent, child := buildTestTree()
	d := NewDispatcher(tree)

	parentCalled := false
	tree.AddHandler(child, event.MouseClick, func(e *event.Event) {
		e.StopPropagation()
	})
	tree.AddHandler(parent, event.MouseClick, func(e *event.Event) {
		if e.Phase == event.PhaseBubble {
			parentCalled = true
		}
	})

	d.Dispatch(child, &event.Event{Type: event.MouseClick})

	if parentCalled {
		t.Error("parent bubble handler should not be called after StopPropagation")
	}
}

func TestDispatchStopPropagationInCapture(t *testing.T) {
	tree, _, parent, child := buildTestTree()
	d := NewDispatcher(tree)

	childCalled := false
	tree.AddHandler(parent, event.MouseClick, func(e *event.Event) {
		if e.Phase == event.PhaseCapture {
			e.StopPropagation()
		}
	})
	tree.AddHandler(child, event.MouseClick, func(e *event.Event) {
		childCalled = true
	})

	d.Dispatch(child, &event.Event{Type: event.MouseClick})

	if childCalled {
		t.Error("child should not receive event when capture phase stops propagation")
	}
}

func TestDispatchPreventDefault(t *testing.T) {
	tree, _, _, child := buildTestTree()
	d := NewDispatcher(tree)

	tree.AddHandler(child, event.MouseClick, func(e *event.Event) {
		e.PreventDefault()
	})

	evt := &event.Event{Type: event.MouseClick}
	result := d.Dispatch(child, evt)

	if result {
		t.Error("Dispatch should return false when default is prevented")
	}
}

func TestDispatchFocusDoesNotBubble(t *testing.T) {
	tree, _, parent, child := buildTestTree()
	d := NewDispatcher(tree)

	parentCalled := false
	tree.AddHandler(parent, event.Focus, func(e *event.Event) {
		if e.Phase == event.PhaseBubble {
			parentCalled = true
		}
	})

	d.Dispatch(child, &event.Event{Type: event.Focus})

	if parentCalled {
		t.Error("Focus event should not bubble")
	}
}

func TestDispatchFocusInBubbles(t *testing.T) {
	tree, _, parent, child := buildTestTree()
	d := NewDispatcher(tree)

	parentCalled := false
	tree.AddHandler(parent, event.FocusIn, func(e *event.Event) {
		if e.Phase == event.PhaseBubble {
			parentCalled = true
		}
	})

	d.Dispatch(child, &event.Event{Type: event.FocusIn})

	if !parentCalled {
		t.Error("FocusIn event should bubble")
	}
}

func TestDispatchMouseEnterDoesNotBubble(t *testing.T) {
	tree, _, parent, child := buildTestTree()
	d := NewDispatcher(tree)

	parentCalled := false
	tree.AddHandler(parent, event.MouseEnter, func(e *event.Event) {
		if e.Phase == event.PhaseBubble {
			parentCalled = true
		}
	})

	d.Dispatch(child, &event.Event{Type: event.MouseEnter})

	if parentCalled {
		t.Error("MouseEnter should not bubble")
	}
}

func TestDispatchToTargetOnly(t *testing.T) {
	tree, _, parent, child := buildTestTree()
	d := NewDispatcher(tree)

	parentCalled := false
	childCalled := false
	tree.AddHandler(parent, event.MouseClick, func(e *event.Event) {
		parentCalled = true
	})
	tree.AddHandler(child, event.MouseClick, func(e *event.Event) {
		childCalled = true
	})

	d.DispatchToTarget(child, &event.Event{Type: event.MouseClick})

	if parentCalled {
		t.Error("DispatchToTarget should not invoke parent handlers")
	}
	if !childCalled {
		t.Error("DispatchToTarget should invoke target handlers")
	}
}

func TestDispatchNonExistentTarget(t *testing.T) {
	tree := NewTree()
	d := NewDispatcher(tree)

	result := d.Dispatch(999, &event.Event{Type: event.MouseClick})
	if !result {
		t.Error("dispatching to non-existent target should return true")
	}
}

func TestDispatchFullPropagationOrder(t *testing.T) {
	tree, root, parent, child := buildTestTree()
	d := NewDispatcher(tree)

	var phases []string
	handler := func(name string) EventHandler {
		return func(e *event.Event) {
			switch e.Phase {
			case event.PhaseCapture:
				phases = append(phases, name+"-capture")
			case event.PhaseTarget:
				phases = append(phases, name+"-target")
			case event.PhaseBubble:
				phases = append(phases, name+"-bubble")
			}
		}
	}

	tree.AddHandler(root, event.MouseDown, handler("root"))
	tree.AddHandler(parent, event.MouseDown, handler("parent"))
	tree.AddHandler(child, event.MouseDown, handler("child"))

	d.Dispatch(child, &event.Event{Type: event.MouseDown})

	expected := []string{
		"root-capture", "parent-capture",
		"child-target",
		"parent-bubble", "root-bubble",
	}
	if len(phases) != len(expected) {
		t.Fatalf("expected %d phases, got %d: %v", len(expected), len(phases), phases)
	}
	for i, exp := range expected {
		if phases[i] != exp {
			t.Errorf("phase[%d]: expected %s, got %s", i, exp, phases[i])
		}
	}
}

func TestDispatchMultipleHandlersOnSameElement(t *testing.T) {
	tree, _, _, child := buildTestTree()
	d := NewDispatcher(tree)

	count := 0
	tree.AddHandler(child, event.MouseClick, func(e *event.Event) { count++ })
	tree.AddHandler(child, event.MouseClick, func(e *event.Event) { count++ })
	tree.AddHandler(child, event.MouseClick, func(e *event.Event) { count++ })

	d.Dispatch(child, &event.Event{Type: event.MouseClick})

	if count != 3 {
		t.Errorf("expected 3 handler calls, got %d", count)
	}
}

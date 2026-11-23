package core

import "github.com/kasuganosora/ui/event"

// SetFocusable marks an element as a tab stop. Disabled or hidden elements
// are still stored as focusable but are skipped while they cannot be used.
func (t *Tree) SetFocusable(id ElementID, focusable bool) {
	if elem := t.elements[id]; elem != nil {
		elem.focusable = focusable
	}
}

// IsFocusable reports the stored tab-stop flag.
func (t *Tree) IsFocusable(id ElementID) bool {
	elem := t.elements[id]
	return elem != nil && elem.focusable
}

// Focused returns the element that currently holds focus.
func (t *Tree) Focused() ElementID {
	if t.focused != InvalidElementID && t.elements[t.focused] == nil {
		t.focused = InvalidElementID
	}
	return t.focused
}

// PushFocusScope limits tab movement to root and its descendants.
func (t *Tree) PushFocusScope(root ElementID) {
	t.scopes = append(t.scopes, root)
}

// PopFocusScope removes the innermost tab scope.
func (t *Tree) PopFocusScope() {
	if len(t.scopes) == 0 {
		return
	}
	t.scopes = t.scopes[:len(t.scopes)-1]
}

// FocusNext moves focus to the next tab stop after from and returns it.
func (t *Tree) FocusNext(from ElementID) ElementID {
	return t.moveFocus(from, 1)
}

// FocusPrev moves focus to the previous tab stop before from and returns it.
func (t *Tree) FocusPrev(from ElementID) ElementID {
	return t.moveFocus(from, -1)
}

// RouteKey delivers a key. Tab and Shift+Tab move focus. Other keys bubble
// from the focused element.
func (t *Tree) RouteKey(evt *event.Event) {
	if evt == nil {
		return
	}
	if evt.Type == event.KeyDown && evt.Key == event.KeyTab {
		from := t.Focused()
		var next ElementID
		if evt.Modifiers.Shift {
			next = t.FocusPrev(from)
		} else {
			next = t.FocusNext(from)
		}
		if next == InvalidElementID && from != InvalidElementID {
			t.SetFocused(from, true)
		}
		evt.StopPropagation()
		return
	}
	target := t.Focused()
	if target == InvalidElementID {
		target = t.root
	}
	NewDispatcher(t).Dispatch(target, evt)
}

func (t *Tree) moveFocus(from ElementID, dir int) ElementID {
	stops := t.tabStops()
	if len(stops) == 0 {
		return InvalidElementID
	}
	idx := -1
	for i, id := range stops {
		if id == from {
			idx = i
			break
		}
	}
	var next ElementID
	if idx < 0 {
		if dir < 0 {
			next = stops[len(stops)-1]
		} else {
			next = stops[0]
		}
	} else {
		next = stops[(idx+dir+len(stops))%len(stops)]
	}
	t.SetFocused(next, true)
	return next
}

func (t *Tree) tabStops() []ElementID {
	var scope ElementID
	if n := len(t.scopes); n > 0 {
		scope = t.scopes[n-1]
	}
	var out []ElementID
	t.Walk(t.root, func(id ElementID, _ int) bool {
		elem := t.elements[id]
		if elem == nil {
			return true
		}
		if scope != InvalidElementID && !t.within(id, scope) {
			return true
		}
		if elem.focusable && elem.visible && elem.enabled {
			out = append(out, id)
		}
		return true
	})
	return out
}

func (t *Tree) forgetScopes(id ElementID) {
	kept := t.scopes[:0]
	for _, scope := range t.scopes {
		if scope == id || t.within(scope, id) {
			continue
		}
		kept = append(kept, scope)
	}
	t.scopes = kept
}

func (t *Tree) within(id, scope ElementID) bool {
	for id != InvalidElementID {
		if id == scope {
			return true
		}
		elem := t.elements[id]
		if elem == nil {
			return false
		}
		id = elem.parent
	}
	return false
}

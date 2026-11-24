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

// PopFocusScopeOwned removes the scope opened by root, even when another
// scope was pushed later.
func (t *Tree) PopFocusScopeOwned(root ElementID) {
	kept := t.scopes[:0]
	for _, scope := range t.scopes {
		if scope != root {
			kept = append(kept, scope)
		}
	}
	t.scopes = kept
}

// FocusScope returns the innermost tab scope, or InvalidElementID.
func (t *Tree) FocusScope() ElementID {
	if len(t.scopes) == 0 {
		return InvalidElementID
	}
	return t.scopes[len(t.scopes)-1]
}

// SetKeyFallback delivers keys here when nothing is focused.
func (t *Tree) SetKeyFallback(id ElementID) { t.keyFallback = id }

// InScope reports whether id is root or a descendant of root.
// A zero root means every element is in scope.
func (t *Tree) InScope(id, root ElementID) bool {
	if root == InvalidElementID {
		return true
	}
	return t.within(id, root)
}

// PushLayer records id as the top modal layer. Construction order does not
// matter; a later PushLayer paints and hits above earlier layers. Descendants
// share the layer, so a popup owned by the modal stays above that modal and
// below a modal opened afterwards.
func (t *Tree) PushLayer(id ElementID) {
	if id == InvalidElementID {
		return
	}
	t.PopLayer(id)
	t.layerSeq++
	t.layers = append(t.layers, modalLayer{id: id, rank: t.layerSeq})
}

// PopLayer removes the layer owned by id. Earlier layers stay in place.
func (t *Tree) PopLayer(id ElementID) {
	kept := t.layers[:0]
	for _, layer := range t.layers {
		if layer.id != id {
			kept = append(kept, layer)
		}
	}
	t.layers = kept
}

// topModal returns the visible layer with the highest rank.
func (t *Tree) topModal() (ElementID, int32) {
	var id ElementID
	var rank int32
	for _, layer := range t.layers {
		if layer.rank > rank && t.ChainAllows(layer.id) {
			rank = layer.rank
			id = layer.id
		}
	}
	return id, rank
}

// LayerRank returns the rank of the nearest modal layer at id or an ancestor.
// Zero means the element is not inside an open modal layer.
func (t *Tree) LayerRank(id ElementID) int32 {
	best := int32(0)
	for id != InvalidElementID {
		for _, layer := range t.layers {
			if layer.id == id && layer.rank > best {
				best = layer.rank
			}
		}
		elem := t.elements[id]
		if elem == nil {
			break
		}
		id = elem.parent
	}
	return best
}

// PushTransient registers id as an Escape target above the focused element.
// Focus is left where it is. PopTransient removes that registration.
func (t *Tree) PushTransient(id ElementID) {
	if id == InvalidElementID {
		return
	}
	t.PopTransient(id)
	t.transients = append(t.transients, id)
}

// PopTransient removes id from the transient Escape stack.
func (t *Tree) PopTransient(id ElementID) {
	kept := t.transients[:0]
	for _, cur := range t.transients {
		if cur != id {
			kept = append(kept, cur)
		}
	}
	t.transients = kept
}

// ChainAllows reports whether id and every ancestor are visible and enabled.
func (t *Tree) ChainAllows(id ElementID) bool {
	if id == InvalidElementID {
		return false
	}
	for id != InvalidElementID {
		elem := t.elements[id]
		if elem == nil || !elem.visible || !elem.enabled {
			return false
		}
		id = elem.parent
	}
	return true
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
	if evt.Type == event.KeyDown && evt.Key == event.KeyEscape {
		modal, modalRank := t.topModal()
		for i := len(t.transients) - 1; i >= 0; i-- {
			id := t.transients[i]
			if !t.ChainAllows(id) {
				continue
			}
			if modal != InvalidElementID && t.LayerRank(id) < modalRank && !t.InScope(id, modal) {
				continue
			}
			NewDispatcher(t).Dispatch(id, evt)
			if evt.IsStopped() {
				return
			}
		}
	}
	target := t.Focused()
	if target == InvalidElementID {
		target = t.keyFallback
	}
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
		if elem.focusable && t.ChainAllows(id) {
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

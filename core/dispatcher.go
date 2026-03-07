package core

import "github.com/kasuganosora/ui/event"

// Dispatcher dispatches events through the element tree following the
// W3C event model: capture phase (root → target), target phase,
// then bubble phase (target → root).
type Dispatcher struct {
	tree *Tree
}

// NewDispatcher creates a new event dispatcher for the given tree.
func NewDispatcher(tree *Tree) *Dispatcher {
	return &Dispatcher{tree: tree}
}

// Dispatch sends an event to the target element, propagating through
// capture and bubble phases. Returns true if the event was not cancelled
// (i.e., PreventDefault was not called).
func (d *Dispatcher) Dispatch(targetID ElementID, evt *event.Event) bool {
	target := d.tree.Get(targetID)
	if target == nil {
		return true
	}

	// Build propagation path: target → root (excluding target itself)
	path := d.buildPath(targetID)

	// Phase 1: Capture (root → target, excluding target)
	evt.Phase = event.PhaseCapture
	for i := len(path) - 1; i >= 0; i-- {
		if evt.IsStopped() {
			break
		}
		d.invokeHandlers(path[i], evt)
	}

	// Phase 2: Target
	if !evt.IsStopped() {
		evt.Phase = event.PhaseTarget
		d.invokeHandlers(targetID, evt)
	}

	// Phase 3: Bubble (target → root, excluding target)
	// Some events don't bubble (e.g., Focus, Blur, MouseEnter, MouseLeave)
	if !evt.IsStopped() && doesBubble(evt.Type) {
		evt.Phase = event.PhaseBubble
		for _, id := range path {
			if evt.IsStopped() {
				break
			}
			d.invokeHandlers(id, evt)
		}
	}

	evt.Phase = event.PhaseNone
	return !evt.IsDefaultPrevented()
}

// DispatchToTarget sends an event only to the target element (no propagation).
func (d *Dispatcher) DispatchToTarget(targetID ElementID, evt *event.Event) bool {
	target := d.tree.Get(targetID)
	if target == nil {
		return true
	}
	evt.Phase = event.PhaseTarget
	d.invokeHandlers(targetID, evt)
	evt.Phase = event.PhaseNone
	return !evt.IsDefaultPrevented()
}

// buildPath returns the ancestor chain from target's parent up to root.
// The result is ordered [parent, grandparent, ..., root].
func (d *Dispatcher) buildPath(targetID ElementID) []ElementID {
	var path []ElementID
	elem := d.tree.Get(targetID)
	if elem == nil {
		return nil
	}
	current := elem.parent
	for current != InvalidElementID {
		path = append(path, current)
		ancestor := d.tree.Get(current)
		if ancestor == nil {
			break
		}
		current = ancestor.parent
	}
	return path
}

// invokeHandlers calls all registered handlers for the event type on the element.
func (d *Dispatcher) invokeHandlers(id ElementID, evt *event.Event) {
	handlers := d.tree.Handlers(id, evt.Type)
	for _, h := range handlers {
		if evt.IsStopped() {
			break
		}
		h(evt)
	}
}

// doesBubble returns whether an event type participates in the bubble phase.
// Focus/blur and enter/leave events do not bubble per W3C spec.
func doesBubble(t event.Type) bool {
	switch t {
	case event.FocusIn, event.FocusOut: // These DO bubble (focusin/focusout)
		return true
	case event.Focus, event.Blur: // These do NOT bubble
		return false
	case event.MouseEnter, event.MouseLeave: // These do NOT bubble
		return false
	default:
		return true
	}
}

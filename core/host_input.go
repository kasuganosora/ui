package core

import "github.com/kasuganosora/ui/event"

// HostInput is the pointer and key routing shared by the app, embedded,
// headless, and Godot hosts. It owns pointer capture so a drag can leave
// the widget that received the press, and it sends keys through RouteKey.
type HostInput struct {
	Tree    *Tree
	pressed eventTarget
	hover   ElementID
}

type eventTarget = ElementID

// Handle routes one host event. Mouse clicks are synthesized once from a
// matching press and release. A MouseClick delivered by the host is
// dispatched as itself and is not synthesized again.
func (h *HostInput) Handle(evt *event.Event) {
	if h == nil || h.Tree == nil || evt == nil {
		return
	}
	if evt.Type.IsKeyboard() || evt.Type.IsIME() {
		h.Tree.RouteKey(evt)
		return
	}
	if evt.Type.IsMouse() {
		h.pointer(evt)
	}
}

func (h *HostInput) pointer(evt *event.Event) {
	hit := h.Tree.HitTest(evt.GlobalX, evt.GlobalY)
	d := NewDispatcher(h.Tree)
	switch evt.Type {
	case event.MouseMove:
		h.moveHover(hit)
		target := hit
		if h.pressed != InvalidElementID {
			target = h.pressed
		}
		if target != InvalidElementID {
			d.Dispatch(target, evt)
		}
	case event.MouseDown:
		h.acquireFocus(hit)
		if hit != InvalidElementID {
			d.Dispatch(hit, evt)
			h.pressed = hit
		}
	case event.MouseUp:
		target := h.pressed
		if target == InvalidElementID {
			target = hit
		}
		if target != InvalidElementID {
			d.Dispatch(target, evt)
			if h.pressed != InvalidElementID && h.pressed == hit {
				click := *evt
				click.Type = event.MouseClick
				d.Dispatch(target, &click)
			}
		}
		h.pressed = InvalidElementID
	case event.MouseClick:
		h.acquireFocus(hit)
		if hit != InvalidElementID {
			d.Dispatch(hit, evt)
		}
	case event.MouseWheel:
		target := hit
		if h.pressed != InvalidElementID {
			target = h.pressed
		}
		if target != InvalidElementID {
			d.Dispatch(target, evt)
		}
	}
}

func (h *HostInput) moveHover(hit ElementID) {
	if hit == h.hover {
		return
	}
	d := NewDispatcher(h.Tree)
	if h.hover != InvalidElementID {
		h.Tree.SetHovered(h.hover, false)
		d.Dispatch(h.hover, &event.Event{Type: event.MouseLeave})
	}
	if hit != InvalidElementID {
		h.Tree.SetHovered(hit, true)
		d.Dispatch(hit, &event.Event{Type: event.MouseEnter})
	}
	h.hover = hit
}

func (h *HostInput) acquireFocus(hit ElementID) {
	for id := hit; id != InvalidElementID; {
		elem := h.Tree.Get(id)
		if elem == nil {
			return
		}
		if elem.focusable && elem.enabled && elem.visible {
			h.Tree.SetFocused(id, true)
			return
		}
		id = elem.parent
	}
}

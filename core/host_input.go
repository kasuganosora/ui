package core

import "github.com/kasuganosora/ui/event"

// HostInput is the pointer and key routing shared by the app, embedded,
// headless, and Godot hosts. It owns pointer capture so a drag can leave
// the widget that received the press, and it sends keys through RouteKey.
type HostInput struct {
	Tree         *Tree
	pressed      eventTarget
	hover        ElementID
	swallowClick bool
}

type eventTarget = ElementID

// Handle routes one host event. A press and release on the same target
// synthesize one MouseClick. An explicit MouseClick is delivered once,
// and the explicit click that follows that synthesized pair is ignored
// so a host can send Down+Up+Click without activating twice. A release
// on a different target produces no click. TouchCancel and WindowBlur
// release capture with a MouseUp and no click. A press that does not
// land on a focusable element clears focus, except when the press is
// inside the element that already has focus so an input keeps IME.
func (h *HostInput) Handle(evt *event.Event) {
	if h == nil || h.Tree == nil || evt == nil {
		return
	}
	if evt.Type.IsKeyboard() || evt.Type.IsIME() {
		h.Tree.RouteKey(evt)
		return
	}
	if evt.Type == event.TouchCancel || evt.Type == event.WindowBlur {
		h.releaseCapture()
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
				click.Phase = 0
				d.Dispatch(target, &click)
				h.swallowClick = true
			}
		}
		h.pressed = InvalidElementID
	case event.MouseClick:
		if h.swallowClick {
			h.swallowClick = false
			return
		}
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

func (h *HostInput) releaseCapture() {
	if h.pressed != InvalidElementID {
		NewDispatcher(h.Tree).Dispatch(h.pressed, &event.Event{Type: event.MouseUp})
	}
	h.pressed = InvalidElementID
	h.swallowClick = false
}

func (h *HostInput) acquireFocus(hit ElementID) {
	for id := hit; id != InvalidElementID; {
		elem := h.Tree.Get(id)
		if elem == nil {
			return
		}
		if elem.focusable && h.Tree.ChainAllows(id) {
			h.Tree.SetFocused(id, true)
			return
		}
		id = elem.parent
	}
	focused := h.Tree.Focused()
	if focused != InvalidElementID && h.Tree.InScope(hit, focused) {
		return
	}
	h.Tree.ClearFocus()
}

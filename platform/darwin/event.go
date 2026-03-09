//go:build darwin

package darwin

import (
	"github.com/kasuganosora/ui/event"
)

// isKeyPressed returns true if the given key is currently pressed.
// On macOS, we track modifier states from the most recent flags changed event.
func (w *Window) isKeyPressed(key event.Key) bool {
	switch key {
	case event.KeyLeftShift, event.KeyRightShift:
		return w.lastModifiers.Shift
	case event.KeyLeftCtrl, event.KeyRightCtrl:
		return w.lastModifiers.Ctrl
	case event.KeyLeftAlt, event.KeyRightAlt:
		return w.lastModifiers.Alt
	case event.KeyLeftSuper, event.KeyRightSuper:
		return w.lastModifiers.Super
	}
	return false
}

// getCurrentModifiers returns the current modifier state.
func (w *Window) getCurrentModifiers() event.Modifiers {
	return w.lastModifiers
}

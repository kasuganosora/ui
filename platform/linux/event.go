//go:build linux && !android

package linux

import (
	"unicode/utf8"
	"unsafe"

	"github.com/kasuganosora/ui/event"
)

// X11 modifier state bit masks (from X.h)
const (
	x11ShiftMask   = 1 << 0
	x11LockMask    = 1 << 1
	x11ControlMask = 1 << 2
	x11Mod1Mask    = 1 << 3 // Alt
	x11Mod4Mask    = 1 << 6 // Super (Windows key)
)

// translateEvent processes a single X11 event and appends translated events.
func (p *Platform) translateEvent(xe XEvent, win *Window, events *[]event.Event) {
	evType := int(xe[0])

	switch evType {
	case KeyPress:
		kev := (*XKeyEvent)(unsafe.Pointer(&xe))
		keysym := XLookupKeysym(kev, 0)
		key := keysymToKey(keysym)
		mods := modifiersFromState(kev.State)

		*events = append(*events, event.Event{
			Type:      event.KeyDown,
			Key:       key,
			Modifiers: mods,
		})

		// Emit a CharEvent for printable characters (basic Latin and beyond).
		// A full implementation would use XLookupString or XmbLookupString.
		ch := keysymToRune(keysym)
		if ch >= 32 && ch != 127 {
			*events = append(*events, event.Event{
				Type:      event.KeyPress,
				Char:      ch,
				Modifiers: mods,
			})
		}

	case KeyRelease:
		kev := (*XKeyEvent)(unsafe.Pointer(&xe))
		keysym := XLookupKeysym(kev, 0)
		key := keysymToKey(keysym)
		mods := modifiersFromState(kev.State)

		*events = append(*events, event.Event{
			Type:      event.KeyUp,
			Key:       key,
			Modifiers: mods,
		})

	case ButtonPress:
		bev := (*XButtonEvent)(unsafe.Pointer(&xe))
		btn := bev.Button
		mods := modifiersFromState(bev.State)

		switch btn {
		case 4: // Scroll up
			*events = append(*events, event.Event{
				Type:      event.MouseWheel,
				WheelDY:   1.0,
				GlobalX:   float32(bev.X),
				GlobalY:   float32(bev.Y),
				Modifiers: mods,
			})
		case 5: // Scroll down
			*events = append(*events, event.Event{
				Type:      event.MouseWheel,
				WheelDY:   -1.0,
				GlobalX:   float32(bev.X),
				GlobalY:   float32(bev.Y),
				Modifiers: mods,
			})
		case 6: // Scroll left
			*events = append(*events, event.Event{
				Type:      event.MouseWheel,
				WheelDX:   -1.0,
				GlobalX:   float32(bev.X),
				GlobalY:   float32(bev.Y),
				Modifiers: mods,
			})
		case 7: // Scroll right
			*events = append(*events, event.Event{
				Type:      event.MouseWheel,
				WheelDX:   1.0,
				GlobalX:   float32(bev.X),
				GlobalY:   float32(bev.Y),
				Modifiers: mods,
			})
		default:
			*events = append(*events, event.Event{
				Type:      event.MouseDown,
				Button:    x11ButtonToMouseButton(btn),
				GlobalX:   float32(bev.X),
				GlobalY:   float32(bev.Y),
				X:         float32(bev.X),
				Y:         float32(bev.Y),
				Modifiers: mods,
			})
		}

	case ButtonRelease:
		bev := (*XButtonEvent)(unsafe.Pointer(&xe))
		btn := bev.Button
		// Skip scroll wheel synthetic releases (buttons 4-7)
		if btn >= 4 && btn <= 7 {
			return
		}
		mods := modifiersFromState(bev.State)
		*events = append(*events, event.Event{
			Type:      event.MouseUp,
			Button:    x11ButtonToMouseButton(btn),
			GlobalX:   float32(bev.X),
			GlobalY:   float32(bev.Y),
			X:         float32(bev.X),
			Y:         float32(bev.Y),
			Modifiers: mods,
		})

	case MotionNotify:
		mev := (*XMotionEvent)(unsafe.Pointer(&xe))
		*events = append(*events, event.Event{
			Type:    event.MouseMove,
			GlobalX: float32(mev.X),
			GlobalY: float32(mev.Y),
			X:       float32(mev.X),
			Y:       float32(mev.Y),
		})

	case ConfigureNotify:
		cev := (*XConfigureEvent)(unsafe.Pointer(&xe))
		newW := int(cev.Width)
		newH := int(cev.Height)
		if newW != win.width || newH != win.height {
			win.width = newW
			win.height = newH
			*events = append(*events, event.Event{
				Type:         event.WindowResize,
				WindowWidth:  newW,
				WindowHeight: newH,
			})
		}
		// Update stored position
		win.posX = int(cev.X)
		win.posY = int(cev.Y)

	case ClientMessage:
		cmev := (*XClientMessageEvent)(unsafe.Pointer(&xe))
		// WM_DELETE_WINDOW: user clicked the close button
		if Atom(cmev.Data[0]) == win.p.wmDeleteWindow {
			win.shouldClose = true
			*events = append(*events, event.Event{
				Type: event.WindowClose,
			})
		}

	case FocusIn:
		*events = append(*events, event.Event{
			Type:        event.WindowFocus,
		})

	case FocusOut:
		*events = append(*events, event.Event{
			Type: event.WindowBlur,
		})

	case Expose:
		// We redraw every frame; ignore expose events.

	case DestroyNotify:
		win.shouldClose = true
	}
}

// modifiersFromState converts an X11 state bitmask to event.Modifiers.
func modifiersFromState(state uint32) event.Modifiers {
	return event.Modifiers{
		Shift: state&x11ShiftMask != 0,
		Ctrl:  state&x11ControlMask != 0,
		Alt:   state&x11Mod1Mask != 0,
		Super: state&x11Mod4Mask != 0,
	}
}

// x11ButtonToMouseButton maps X11 button numbers to event.MouseButton.
func x11ButtonToMouseButton(btn uint32) event.MouseButton {
	switch btn {
	case 1:
		return event.MouseButtonLeft
	case 2:
		return event.MouseButtonMiddle
	case 3:
		return event.MouseButtonRight
	case 8:
		return event.MouseButton4
	case 9:
		return event.MouseButton5
	default:
		return event.MouseButtonLeft
	}
}

// keysymToRune converts a basic X11 keysym to a Unicode rune for text input.
// This handles Latin-1 and basic Unicode keysyms directly.
func keysymToRune(keysym uint64) rune {
	// X11 keysyms in range 0x01000000–0x0110FFFF encode Unicode directly.
	if keysym&0xFF000000 == 0x01000000 {
		r := rune(keysym & 0x00FFFFFF)
		if utf8.ValidRune(r) {
			return r
		}
		return 0
	}
	// Latin-1 supplement (0x00A0–0x00FF) maps directly to Unicode.
	if keysym >= 0x00A0 && keysym <= 0x00FF {
		return rune(keysym)
	}
	// Basic ASCII printable range (0x0020–0x007E).
	if keysym >= 0x0020 && keysym <= 0x007E {
		return rune(keysym)
	}
	return 0
}

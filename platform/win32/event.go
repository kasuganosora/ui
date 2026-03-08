//go:build windows

package win32

import (
	"unsafe"

	"github.com/kasuganosora/ui/event"
)

// wndProc is the Win32 window procedure callback.
// It translates Win32 messages into domain events.
func wndProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	w := windowFromHWND(hwnd)
	if w == nil {
		ret, _, _ := procDefWindowProcW.Call(hwnd, msg, wParam, lParam)
		return ret
	}

	switch msg {
	case WM_CLOSE:
		w.shouldClose = true
		w.p.pushEvent(event.Event{
			Type:      event.WindowClose,
			Timestamp: queryTimeMicroseconds(),
		})
		return 0

	case WM_ENTERSIZEMOVE:
		w.inSizeMove = true
		return 0

	case WM_EXITSIZEMOVE:
		w.inSizeMove = false
		return 0

	case WM_SIZE:
		w.updateClientSize()
		w.p.pushEvent(event.Event{
			Type:         event.WindowResize,
			Timestamp:    queryTimeMicroseconds(),
			WindowWidth:  w.width,
			WindowHeight: w.height,
		})
		// During the modal resize loop (user dragging border), the main
		// loop is blocked by DefWindowProc. Fire the resize callback so
		// the app can render intermediate frames.
		if w.inSizeMove && w.onResizeFunc != nil {
			w.onResizeFunc()
		}
		return 0

	case WM_MOVE:
		w.updatePosition()
		return 0

	case WM_SETFOCUS:
		w.p.pushEvent(event.Event{
			Type:      event.WindowFocus,
			Timestamp: queryTimeMicroseconds(),
		})
		return 0

	case WM_KILLFOCUS:
		w.p.pushEvent(event.Event{
			Type:      event.WindowBlur,
			Timestamp: queryTimeMicroseconds(),
		})
		return 0

	case WM_DPICHANGED:
		newDPI := float32(hiword(wParam))
		w.dpiScale = newDPI / 96.0
		w.updateClientSize()

		// The lParam contains a suggested RECT
		suggested := (*RECT)(unsafe.Pointer(lParam))
		procSetWindowPos.Call(hwnd, 0,
			uintptr(suggested.Left), uintptr(suggested.Top),
			uintptr(suggested.Right-suggested.Left),
			uintptr(suggested.Bottom-suggested.Top),
			SWP_NOZORDER,
		)

		w.p.pushEvent(event.Event{
			Type:      event.WindowDPIChange,
			Timestamp: queryTimeMicroseconds(),
			DPIScale:  w.dpiScale,
		})
		return 0

	case WM_GETMINMAXINFO:
		mmi := (*MINMAXINFO)(unsafe.Pointer(lParam))
		if w.minWidth > 0 || w.minHeight > 0 {
			mmi.PtMinTrackSize.X = int32(w.minWidth)
			mmi.PtMinTrackSize.Y = int32(w.minHeight)
		}
		if w.maxWidth > 0 || w.maxHeight > 0 {
			mmi.PtMaxTrackSize.X = int32(w.maxWidth)
			mmi.PtMaxTrackSize.Y = int32(w.maxHeight)
		}
		return 0

	// ---- Mouse events ----

	case WM_MOUSEMOVE:
		w.enableMouseTracking()
		w.p.pushEvent(event.Event{
			Type:      event.MouseMove,
			Timestamp: queryTimeMicroseconds(),
			GlobalX:   getXLParam(lParam),
			GlobalY:   getYLParam(lParam),
			X:         getXLParam(lParam),
			Y:         getYLParam(lParam),
			Modifiers: getModifiers(),
		})
		return 0

	case WM_LBUTTONDOWN:
		procSetCapture.Call(hwnd)
		w.p.pushEvent(mouseButtonEvent(event.MouseDown, event.MouseButtonLeft, lParam))
		return 0

	case WM_LBUTTONUP:
		procReleaseCapture.Call()
		w.p.pushEvent(mouseButtonEvent(event.MouseUp, event.MouseButtonLeft, lParam))
		return 0

	case WM_RBUTTONDOWN:
		procSetCapture.Call(hwnd)
		w.p.pushEvent(mouseButtonEvent(event.MouseDown, event.MouseButtonRight, lParam))
		return 0

	case WM_RBUTTONUP:
		procReleaseCapture.Call()
		w.p.pushEvent(mouseButtonEvent(event.MouseUp, event.MouseButtonRight, lParam))
		return 0

	case WM_MBUTTONDOWN:
		procSetCapture.Call(hwnd)
		w.p.pushEvent(mouseButtonEvent(event.MouseDown, event.MouseButtonMiddle, lParam))
		return 0

	case WM_MBUTTONUP:
		procReleaseCapture.Call()
		w.p.pushEvent(mouseButtonEvent(event.MouseUp, event.MouseButtonMiddle, lParam))
		return 0

	case WM_LBUTTONDBLCLK:
		w.p.pushEvent(event.Event{
			Type:       event.MouseDoubleClick,
			Timestamp:  queryTimeMicroseconds(),
			Button:     event.MouseButtonLeft,
			GlobalX:    getXLParam(lParam),
			GlobalY:    getYLParam(lParam),
			X:          getXLParam(lParam),
			Y:          getYLParam(lParam),
			ClickCount: 2,
			Modifiers:  getModifiers(),
		})
		return 0

	case WM_MOUSEWHEEL:
		w.p.pushEvent(event.Event{
			Type:      event.MouseWheel,
			Timestamp: queryTimeMicroseconds(),
			WheelDY:   getWheelDelta(wParam),
			Modifiers: getModifiers(),
		})
		return 0

	case WM_MOUSEHWHEEL:
		w.p.pushEvent(event.Event{
			Type:      event.MouseWheel,
			Timestamp: queryTimeMicroseconds(),
			WheelDX:   getWheelDelta(wParam),
			Modifiers: getModifiers(),
		})
		return 0

	case WM_MOUSELEAVE:
		w.mouseTracked = false
		w.p.pushEvent(event.Event{
			Type:      event.MouseLeave,
			Timestamp: queryTimeMicroseconds(),
		})
		return 0

	case WM_XBUTTONDOWN:
		btn := xButtonFromWParam(wParam)
		procSetCapture.Call(hwnd)
		w.p.pushEvent(mouseButtonEvent(event.MouseDown, btn, lParam))
		return 1 // Must return TRUE for XBUTTON

	case WM_XBUTTONUP:
		btn := xButtonFromWParam(wParam)
		procReleaseCapture.Call()
		w.p.pushEvent(mouseButtonEvent(event.MouseUp, btn, lParam))
		return 1

	// ---- Keyboard events ----

	case WM_KEYDOWN, WM_SYSKEYDOWN:
		key := translateVirtualKey(wParam)
		w.p.pushEvent(event.Event{
			Type:      event.KeyDown,
			Timestamp: queryTimeMicroseconds(),
			Key:       key,
			Modifiers: getModifiers(),
		})
		// Let system handle Alt+F4 etc.
		if msg == WM_SYSKEYDOWN {
			ret, _, _ := procDefWindowProcW.Call(hwnd, msg, wParam, lParam)
			return ret
		}
		return 0

	case WM_KEYUP, WM_SYSKEYUP:
		key := translateVirtualKey(wParam)
		w.p.pushEvent(event.Event{
			Type:      event.KeyUp,
			Timestamp: queryTimeMicroseconds(),
			Key:       key,
			Modifiers: getModifiers(),
		})
		return 0

	case WM_CHAR:
		ch := rune(wParam)
		if ch >= 32 && ch != 127 { // Ignore control characters
			w.p.pushEvent(event.Event{
				Type:      event.KeyPress,
				Timestamp: queryTimeMicroseconds(),
				Char:      ch,
				Modifiers: getModifiers(),
			})
		}
		return 0

	// ---- IME events ----

	case WM_IME_STARTCOMPOSITION:
		// Set IME position before DefWindowProc so the candidate window
		// appears at the cursor location, not at the default position.
		w.applyIMEPosition()
		w.p.pushEvent(event.Event{
			Type:      event.IMECompositionStart,
			Timestamp: queryTimeMicroseconds(),
		})
		ret, _, _ := procDefWindowProcW.Call(hwnd, msg, wParam, lParam)
		return ret

	case WM_IME_COMPOSITION:
		handleIMEComposition(w, lParam)
		ret, _, _ := procDefWindowProcW.Call(hwnd, msg, wParam, lParam)
		return ret

	case WM_IME_ENDCOMPOSITION:
		w.p.pushEvent(event.Event{
			Type:      event.IMECompositionEnd,
			Timestamp: queryTimeMicroseconds(),
		})
		ret, _, _ := procDefWindowProcW.Call(hwnd, msg, wParam, lParam)
		return ret

	// ---- Cursor ----

	case WM_SETCURSOR:
		if loword(lParam) == 1 { // HTCLIENT
			cursor := w.currentCursor
			if cursor == 0 {
				cursor = w.cursorHandles[0] // fallback to arrow
			}
			procSetCursor.Call(cursor)
			return 1
		}

	case WM_ERASEBKGND:
		return 1 // We handle our own painting

	case WM_PAINT:
		// Validate the window to stop WM_PAINT flood
		ret, _, _ := procDefWindowProcW.Call(hwnd, msg, wParam, lParam)
		return ret

	case WM_GETOBJECT:
		if result, handled := w.uiaProvider.HandleGetObject(wParam, lParam); handled {
			return result
		}

	case WM_DESTROY:
		return 0
	}

	ret, _, _ := procDefWindowProcW.Call(hwnd, msg, wParam, lParam)
	return ret
}

// mouseButtonEvent creates a MouseDown/MouseUp event.
func mouseButtonEvent(typ event.Type, btn event.MouseButton, lParam uintptr) event.Event {
	return event.Event{
		Type:      typ,
		Timestamp: queryTimeMicroseconds(),
		Button:    btn,
		GlobalX:   getXLParam(lParam),
		GlobalY:   getYLParam(lParam),
		X:         getXLParam(lParam),
		Y:         getYLParam(lParam),
		Modifiers: getModifiers(),
	}
}

// xButtonFromWParam extracts the X button number from XBUTTON wParam.
func xButtonFromWParam(wp uintptr) event.MouseButton {
	xbutton := hiword(wp)
	if xbutton == 1 {
		return event.MouseButton4
	}
	return event.MouseButton5
}

// getModifiers reads the current keyboard modifier state.
func getModifiers() event.Modifiers {
	return event.Modifiers{
		Ctrl:  isKeyDown(VK_CONTROL),
		Shift: isKeyDown(VK_SHIFT),
		Alt:   isKeyDown(VK_MENU),
		Super: isKeyDown(VK_LWIN) || isKeyDown(VK_RWIN),
	}
}

func isKeyDown(vk int) bool {
	r, _, _ := procGetKeyState.Call(uintptr(vk))
	return r&0x8000 != 0
}

// translateVirtualKey maps Win32 virtual key codes to event.Key.
func translateVirtualKey(vk uintptr) event.Key {
	switch vk {
	case VK_BACK:
		return event.KeyBackspace
	case VK_TAB:
		return event.KeyTab
	case VK_RETURN:
		return event.KeyEnter
	case VK_ESCAPE:
		return event.KeyEscape
	case VK_SPACE:
		return event.KeySpace
	case VK_PRIOR:
		return event.KeyPageUp
	case VK_NEXT:
		return event.KeyPageDown
	case VK_END:
		return event.KeyEnd
	case VK_HOME:
		return event.KeyHome
	case VK_LEFT:
		return event.KeyArrowLeft
	case VK_UP:
		return event.KeyArrowUp
	case VK_RIGHT:
		return event.KeyArrowRight
	case VK_DOWN:
		return event.KeyArrowDown
	case VK_INSERT:
		return event.KeyInsert
	case VK_DELETE:
		return event.KeyDelete
	case VK_SNAPSHOT:
		return event.KeyPrintScreen
	case VK_PAUSE:
		return event.KeyPause
	case VK_CAPITAL:
		return event.KeyCapsLock
	case VK_NUMLOCK:
		return event.KeyNumLock
	case VK_SCROLL:
		return event.KeyScrollLock
	case VK_APPS:
		return event.KeyMenu
	case VK_LSHIFT:
		return event.KeyLeftShift
	case VK_RSHIFT:
		return event.KeyRightShift
	case VK_LCONTROL:
		return event.KeyLeftCtrl
	case VK_RCONTROL:
		return event.KeyRightCtrl
	case VK_LMENU:
		return event.KeyLeftAlt
	case VK_RMENU:
		return event.KeyRightAlt
	case VK_LWIN:
		return event.KeyLeftSuper
	case VK_RWIN:
		return event.KeyRightSuper
	case VK_OEM_MINUS:
		return event.KeyMinus
	case VK_OEM_PLUS:
		return event.KeyEqual
	case VK_OEM_4:
		return event.KeyLeftBracket
	case VK_OEM_6:
		return event.KeyRightBracket
	case VK_OEM_5:
		return event.KeyBackslash
	case VK_OEM_1:
		return event.KeySemicolon
	case VK_OEM_7:
		return event.KeyApostrophe
	case VK_OEM_3:
		return event.KeyGraveAccent
	case VK_OEM_COMMA:
		return event.KeyComma
	case VK_OEM_PERIOD:
		return event.KeyPeriod
	case VK_OEM_2:
		return event.KeySlash
	case VK_MULTIPLY:
		return event.KeyNumpadMultiply
	case VK_ADD:
		return event.KeyNumpadAdd
	case VK_SUBTRACT:
		return event.KeyNumpadSubtract
	case VK_DECIMAL:
		return event.KeyNumpadDecimal
	case VK_DIVIDE:
		return event.KeyNumpadDivide
	}

	// Letters A-Z (VK 0x41-0x5A)
	if vk >= 0x41 && vk <= 0x5A {
		return event.KeyA + event.Key(vk-0x41)
	}
	// Numbers 0-9 (VK 0x30-0x39)
	if vk >= 0x30 && vk <= 0x39 {
		return event.Key0 + event.Key(vk-0x30)
	}
	// Function keys F1-F12 (VK 0x70-0x7B)
	if vk >= VK_F1 && vk <= VK_F12 {
		return event.KeyF1 + event.Key(vk-VK_F1)
	}
	// Numpad 0-9 (VK 0x60-0x69)
	if vk >= VK_NUMPAD0 && vk <= VK_NUMPAD9 {
		return event.KeyNumpad0 + event.Key(vk-VK_NUMPAD0)
	}

	return event.KeyUnknown
}

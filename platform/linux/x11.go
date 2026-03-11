//go:build linux && !android

// Package linux implements the platform.Platform interface for Linux using X11.
// Zero CGO — all X11 calls go through purego (dlopen/dlsym/SyscallN).
package linux

import (
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/kasuganosora/ui/event"
)

// ---- Function pointer variables ----

var (
	fnXOpenDisplay          uintptr
	fnXCloseDisplay         uintptr
	fnXCreateSimpleWindow   uintptr
	fnXDefaultRootWindow    uintptr
	fnXMapWindow            uintptr
	fnXUnmapWindow          uintptr
	fnXDestroyWindow        uintptr
	fnXSelectInput          uintptr
	fnXPending              uintptr
	fnXNextEvent            uintptr
	fnXSendEvent            uintptr
	fnXFlush                uintptr
	fnXSync                 uintptr
	fnXInternAtom           uintptr
	fnXSetWMProtocols       uintptr
	fnXGetWindowAttributes  uintptr
	fnXResizeWindow         uintptr
	fnXMoveWindow           uintptr
	fnXStoreName            uintptr
	fnXDefineCursor         uintptr
	fnXLookupKeysym         uintptr
	fnXSetInputFocus        uintptr
	fnXGetInputFocus        uintptr
	fnXTranslateCoordinates uintptr
	fnXCreateFontCursor     uintptr
	fnXFreeCursor           uintptr
	fnXChangeProperty       uintptr
	fnXGetWindowProperty    uintptr
	fnXFree                 uintptr
	fnXMoveResizeWindow     uintptr
)

func init() {
	h, err := purego.Dlopen("libX11.so.6", purego.RTLD_LAZY|purego.RTLD_GLOBAL)
	if err != nil {
		return // X11 not available; platform will fail at Init time
	}

	sym := func(name string) uintptr {
		s, _ := purego.Dlsym(h, name)
		return s
	}

	fnXOpenDisplay = sym("XOpenDisplay")
	fnXCloseDisplay = sym("XCloseDisplay")
	fnXCreateSimpleWindow = sym("XCreateSimpleWindow")
	fnXDefaultRootWindow = sym("XDefaultRootWindow")
	fnXMapWindow = sym("XMapWindow")
	fnXUnmapWindow = sym("XUnmapWindow")
	fnXDestroyWindow = sym("XDestroyWindow")
	fnXSelectInput = sym("XSelectInput")
	fnXPending = sym("XPending")
	fnXNextEvent = sym("XNextEvent")
	fnXSendEvent = sym("XSendEvent")
	fnXFlush = sym("XFlush")
	fnXSync = sym("XSync")
	fnXInternAtom = sym("XInternAtom")
	fnXSetWMProtocols = sym("XSetWMProtocols")
	fnXGetWindowAttributes = sym("XGetWindowAttributes")
	fnXResizeWindow = sym("XResizeWindow")
	fnXMoveWindow = sym("XMoveWindow")
	fnXStoreName = sym("XStoreName")
	fnXDefineCursor = sym("XDefineCursor")
	fnXLookupKeysym = sym("XLookupKeysym")
	fnXSetInputFocus = sym("XSetInputFocus")
	fnXGetInputFocus = sym("XGetInputFocus")
	fnXTranslateCoordinates = sym("XTranslateCoordinates")
	fnXCreateFontCursor = sym("XCreateFontCursor")
	fnXFreeCursor = sym("XFreeCursor")
	fnXChangeProperty = sym("XChangeProperty")
	fnXGetWindowProperty = sym("XGetWindowProperty")
	fnXFree = sym("XFree")
	fnXMoveResizeWindow = sym("XMoveResizeWindow")
}

// ---- X11 type definitions ----

// Display represents an X11 Display* (opaque pointer).
type Display uintptr

// XWindow is an X11 Window XID.
type XWindow uint64

// Atom is an X11 Atom (interned string identifier).
type Atom uint64

// XID is a generic X resource ID.
type XID uint64

// Bool is X11's Boolean type (int32).
type Bool int32

// Status is X11's Status return type.
type Status int32

// Time is an X11 timestamp.
type Time uint64

// VisualID is an X11 visual identifier.
type VisualID uint64

// XEvent is a fixed-size union (24 longs = 192 bytes on 64-bit Linux).
type XEvent [24]int64

// XConfigureEvent is the configure notification sub-struct within XEvent.
type XConfigureEvent struct {
	Type, Serial int64
	SendEvent    Bool
	Display      Display
	Event, Window XWindow
	X, Y, Width, Height int32
	BorderWidth int32
	Above       XWindow
	OverrideRedirect Bool
}

// XKeyEvent is the key press/release sub-struct within XEvent.
type XKeyEvent struct {
	Type, Serial int64
	SendEvent    Bool
	Display      Display
	Window, Root, Subwindow XWindow
	Time                    Time
	X, Y, XRoot, YRoot      int32
	State, Keycode          uint32
	SameScreen              Bool
}

// XButtonEvent is the button press/release sub-struct within XEvent.
type XButtonEvent struct {
	Type, Serial int64
	SendEvent    Bool
	Display      Display
	Window, Root, Subwindow XWindow
	Time                    Time
	X, Y, XRoot, YRoot      int32
	State, Button           uint32
	SameScreen              Bool
}

// XMotionEvent is the pointer motion sub-struct within XEvent.
type XMotionEvent struct {
	Type, Serial int64
	SendEvent    Bool
	Display      Display
	Window, Root, Subwindow XWindow
	Time                    Time
	X, Y, XRoot, YRoot      int32
	State                   uint32
	IsHint                  int8
	SameScreen              Bool
}

// XClientMessageEvent is the client message sub-struct within XEvent.
type XClientMessageEvent struct {
	Type, Serial int64
	SendEvent    Bool
	Display      Display
	Window      XWindow
	MessageType Atom
	Format      int32
	Data        [5]int64
}

// XFocusChangeEvent is the focus in/out sub-struct within XEvent.
type XFocusChangeEvent struct {
	Type, Serial int64
	SendEvent    Bool
	Display      Display
	Window       XWindow
	Mode, Detail int32
}

// XWindowAttributes holds window geometry and attributes.
type XWindowAttributes struct {
	X, Y, Width, Height, BorderWidth, Depth int32
	_                                        [64]int64
}

// ---- X11 event type constants ----

const (
	KeyPress        = 2
	KeyRelease      = 3
	ButtonPress     = 4
	ButtonRelease   = 5
	MotionNotify    = 6
	EnterNotify     = 7
	LeaveNotify     = 8
	FocusIn         = 9
	FocusOut        = 10
	Expose          = 12
	DestroyNotify   = 17
	ConfigureNotify = 22
	ClientMessage   = 33
)

// ---- X11 event mask constants ----

const (
	NoEventMask         int64 = 0
	KeyPressMask        int64 = 1 << 0
	KeyReleaseMask      int64 = 1 << 1
	ButtonPressMask     int64 = 1 << 2
	ButtonReleaseMask   int64 = 1 << 3
	PointerMotionMask   int64 = 1 << 6
	FocusChangeMask     int64 = 1 << 21
	ExposureMask        int64 = 1 << 15
	StructureNotifyMask int64 = 1 << 17
)

// ---- C string helper ----

// cstring returns a null-terminated byte slice for the given Go string.
func cstring(s string) []byte {
	b := make([]byte, len(s)+1)
	copy(b, s)
	return b
}

// ---- X11 function wrappers ----

// XOpenDisplay opens a connection to the X server.
// Pass empty string for default display ($DISPLAY).
func XOpenDisplay(name string) Display {
	var arg uintptr
	if name != "" {
		b := cstring(name)
		arg = uintptr(unsafe.Pointer(&b[0]))
	}
	r, _, _ := purego.SyscallN(fnXOpenDisplay, arg)
	return Display(r)
}

// XCloseDisplay closes the X server connection.
func XCloseDisplay(dpy Display) {
	purego.SyscallN(fnXCloseDisplay, uintptr(dpy))
}

// XDefaultRootWindow returns the root window for the default screen.
func XDefaultRootWindow(dpy Display) XWindow {
	r, _, _ := purego.SyscallN(fnXDefaultRootWindow, uintptr(dpy))
	return XWindow(r)
}

// XCreateSimpleWindow creates a simple window.
func XCreateSimpleWindow(dpy Display, parent XWindow, x, y int, width, height, borderWidth uint, border, background uint64) XWindow {
	r, _, _ := purego.SyscallN(fnXCreateSimpleWindow,
		uintptr(dpy),
		uintptr(parent),
		uintptr(x), uintptr(y),
		uintptr(width), uintptr(height),
		uintptr(borderWidth),
		uintptr(border),
		uintptr(background),
	)
	return XWindow(r)
}

// XMapWindow maps (shows) the window.
func XMapWindow(dpy Display, w XWindow) {
	purego.SyscallN(fnXMapWindow, uintptr(dpy), uintptr(w))
}

// XUnmapWindow unmaps (hides) the window.
func XUnmapWindow(dpy Display, w XWindow) {
	purego.SyscallN(fnXUnmapWindow, uintptr(dpy), uintptr(w))
}

// XDestroyWindow destroys the window.
func XDestroyWindow(dpy Display, w XWindow) {
	purego.SyscallN(fnXDestroyWindow, uintptr(dpy), uintptr(w))
}

// XSelectInput sets the event mask for the window.
func XSelectInput(dpy Display, w XWindow, eventMask int64) {
	purego.SyscallN(fnXSelectInput, uintptr(dpy), uintptr(w), uintptr(eventMask))
}

// XPending returns the number of events pending in the event queue.
func XPending(dpy Display) int {
	r, _, _ := purego.SyscallN(fnXPending, uintptr(dpy))
	return int(r)
}

// XNextEvent fills the event struct with the next event from the queue.
func XNextEvent(dpy Display, ev *XEvent) {
	purego.SyscallN(fnXNextEvent, uintptr(dpy), uintptr(unsafe.Pointer(ev)))
}

// XSendEvent sends an event to a window.
func XSendEvent(dpy Display, w XWindow, propagate Bool, mask int64, ev *XEvent) Status {
	r, _, _ := purego.SyscallN(fnXSendEvent,
		uintptr(dpy), uintptr(w),
		uintptr(propagate),
		uintptr(mask),
		uintptr(unsafe.Pointer(ev)),
	)
	return Status(r)
}

// XFlush flushes the output buffer.
func XFlush(dpy Display) {
	purego.SyscallN(fnXFlush, uintptr(dpy))
}

// XSync flushes the output buffer and waits for all pending events.
func XSync(dpy Display, discard Bool) {
	purego.SyscallN(fnXSync, uintptr(dpy), uintptr(discard))
}

// XInternAtom returns the atom for a name, optionally failing if it doesn't exist.
func XInternAtom(dpy Display, name string, onlyIfExists Bool) Atom {
	b := cstring(name)
	r, _, _ := purego.SyscallN(fnXInternAtom, uintptr(dpy), uintptr(unsafe.Pointer(&b[0])), uintptr(onlyIfExists))
	return Atom(r)
}

// XSetWMProtocols sets the WM_PROTOCOLS property.
func XSetWMProtocols(dpy Display, w XWindow, protocols *Atom, count int) Status {
	r, _, _ := purego.SyscallN(fnXSetWMProtocols,
		uintptr(dpy), uintptr(w),
		uintptr(unsafe.Pointer(protocols)),
		uintptr(count),
	)
	return Status(r)
}

// XGetWindowAttributes retrieves window geometry and attributes.
func XGetWindowAttributes(dpy Display, w XWindow, attrs *XWindowAttributes) Status {
	r, _, _ := purego.SyscallN(fnXGetWindowAttributes,
		uintptr(dpy), uintptr(w),
		uintptr(unsafe.Pointer(attrs)),
	)
	return Status(r)
}

// XResizeWindow resizes the window.
func XResizeWindow(dpy Display, w XWindow, width, height uint) {
	purego.SyscallN(fnXResizeWindow, uintptr(dpy), uintptr(w), uintptr(width), uintptr(height))
}

// XMoveWindow moves the window.
func XMoveWindow(dpy Display, w XWindow, x, y int) {
	purego.SyscallN(fnXMoveWindow, uintptr(dpy), uintptr(w), uintptr(x), uintptr(y))
}

// XStoreName sets the window title (ASCII/Latin-1 only; use _NET_WM_NAME for UTF-8).
func XStoreName(dpy Display, w XWindow, name string) {
	b := cstring(name)
	purego.SyscallN(fnXStoreName, uintptr(dpy), uintptr(w), uintptr(unsafe.Pointer(&b[0])))
}

// XDefineCursor sets the cursor for the window.
func XDefineCursor(dpy Display, w XWindow, cursor XID) {
	purego.SyscallN(fnXDefineCursor, uintptr(dpy), uintptr(w), uintptr(cursor))
}

// XLookupKeysym returns the keysym for the given key event and index.
func XLookupKeysym(ev *XKeyEvent, index int) uint64 {
	r, _, _ := purego.SyscallN(fnXLookupKeysym, uintptr(unsafe.Pointer(ev)), uintptr(index))
	return uint64(r)
}

// XSetInputFocus sets the keyboard input focus.
func XSetInputFocus(dpy Display, w XWindow, revert int32, time Time) {
	purego.SyscallN(fnXSetInputFocus, uintptr(dpy), uintptr(w), uintptr(revert), uintptr(time))
}

// XGetInputFocus retrieves the keyboard input focus.
func XGetInputFocus(dpy Display, focus *XWindow, revert *int32) {
	purego.SyscallN(fnXGetInputFocus, uintptr(dpy), uintptr(unsafe.Pointer(focus)), uintptr(unsafe.Pointer(revert)))
}

// XTranslateCoordinates translates coordinates between windows.
func XTranslateCoordinates(dpy Display, srcW, destW XWindow, srcX, srcY int, destX, destY *int, child *XWindow) Bool {
	r, _, _ := purego.SyscallN(fnXTranslateCoordinates,
		uintptr(dpy),
		uintptr(srcW), uintptr(destW),
		uintptr(srcX), uintptr(srcY),
		uintptr(unsafe.Pointer(destX)),
		uintptr(unsafe.Pointer(destY)),
		uintptr(unsafe.Pointer(child)),
	)
	return Bool(r)
}

// XCreateFontCursor creates a cursor from the standard cursor font.
func XCreateFontCursor(dpy Display, shape uint) XID {
	r, _, _ := purego.SyscallN(fnXCreateFontCursor, uintptr(dpy), uintptr(shape))
	return XID(r)
}

// XFreeCursor frees a cursor resource.
func XFreeCursor(dpy Display, cursor XID) {
	purego.SyscallN(fnXFreeCursor, uintptr(dpy), uintptr(cursor))
}

// XChangeProperty changes a window property.
func XChangeProperty(dpy Display, w XWindow, property, propType Atom, format int, mode int, data unsafe.Pointer, nelements int) {
	purego.SyscallN(fnXChangeProperty,
		uintptr(dpy), uintptr(w),
		uintptr(property), uintptr(propType),
		uintptr(format), uintptr(mode),
		uintptr(data),
		uintptr(nelements),
	)
}

// XGetWindowProperty retrieves a window property.
func XGetWindowProperty(dpy Display, w XWindow, prop Atom, offset, length int64, del Bool, reqType Atom, actualType *Atom, actualFormat *int, nItems, bytesAfter *uint64, propData **byte) int {
	r, _, _ := purego.SyscallN(fnXGetWindowProperty,
		uintptr(dpy), uintptr(w),
		uintptr(prop),
		uintptr(offset), uintptr(length),
		uintptr(del), uintptr(reqType),
		uintptr(unsafe.Pointer(actualType)),
		uintptr(unsafe.Pointer(actualFormat)),
		uintptr(unsafe.Pointer(nItems)),
		uintptr(unsafe.Pointer(bytesAfter)),
		uintptr(unsafe.Pointer(propData)),
	)
	return int(r)
}

// XFree frees memory allocated by X11.
func XFree(data unsafe.Pointer) {
	purego.SyscallN(fnXFree, uintptr(data))
}

// ---- Keysym to event.Key mapping ----

// X11 keysym constants (selected subset).
const (
	xkBackSpace  = 0xFF08
	xkTab        = 0xFF09
	xkReturn     = 0xFF0D
	xkEscape     = 0xFF1B
	xkDelete     = 0xFFFF
	xkHome       = 0xFF50
	xkLeft       = 0xFF51
	xkUp         = 0xFF52
	xkRight      = 0xFF53
	xkDown       = 0xFF54
	xkPageUp     = 0xFF55
	xkPageDown   = 0xFF56
	xkEnd        = 0xFF57
	xkInsert     = 0xFF63
	xkF1         = 0xFFBE
	xkF2         = 0xFFBF
	xkF3         = 0xFFC0
	xkF4         = 0xFFC1
	xkF5         = 0xFFC2
	xkF6         = 0xFFC3
	xkF7         = 0xFFC4
	xkF8         = 0xFFC5
	xkF9         = 0xFFC6
	xkF10        = 0xFFC7
	xkF11        = 0xFFC8
	xkF12        = 0xFFC9
	xkShiftL     = 0xFFE1
	xkShiftR     = 0xFFE2
	xkControlL   = 0xFFE3
	xkControlR   = 0xFFE4
	xkCapsLock   = 0xFFE5
	xkAltL       = 0xFFE9
	xkAltR       = 0xFFEA
	xkSuperL     = 0xFFEB
	xkSuperR     = 0xFFEC
	xkNumLock    = 0xFF7F
	xkScrollLock = 0xFF14
	xkPrintScreen = 0xFF61
	xkPause      = 0xFF13
	xkMenu       = 0xFF67
	xkSpace      = 0x0020
	xkApostrophe = 0x0027
	xkComma      = 0x002C
	xkMinus      = 0x002D
	xkPeriod     = 0x002E
	xkSlash      = 0x002F
	xkSemicolon  = 0x003B
	xkEqual      = 0x003D
	xkBracketL   = 0x005B
	xkBackslash  = 0x005C
	xkBracketR   = 0x005D
	xkGrave      = 0x0060
	// Numpad
	xkKP0        = 0xFFB0
	xkKP9        = 0xFFB9
	xkKPAdd      = 0xFFAB
	xkKPSubtract = 0xFFAD
	xkKPMultiply = 0xFFAA
	xkKPDivide   = 0xFFAF
	xkKPDecimal  = 0xFFAE
	xkKPEnter    = 0xFF8D
)

// keysymToKey maps an X11 keysym to an event.Key value.
func keysymToKey(keysym uint64) event.Key {
	switch keysym {
	case xkBackSpace:
		return event.KeyBackspace
	case xkTab:
		return event.KeyTab
	case xkReturn, xkKPEnter:
		return event.KeyEnter
	case xkEscape:
		return event.KeyEscape
	case xkDelete:
		return event.KeyDelete
	case xkHome:
		return event.KeyHome
	case xkLeft:
		return event.KeyArrowLeft
	case xkUp:
		return event.KeyArrowUp
	case xkRight:
		return event.KeyArrowRight
	case xkDown:
		return event.KeyArrowDown
	case xkPageUp:
		return event.KeyPageUp
	case xkPageDown:
		return event.KeyPageDown
	case xkEnd:
		return event.KeyEnd
	case xkInsert:
		return event.KeyInsert
	case xkF1:
		return event.KeyF1
	case xkF2:
		return event.KeyF2
	case xkF3:
		return event.KeyF3
	case xkF4:
		return event.KeyF4
	case xkF5:
		return event.KeyF5
	case xkF6:
		return event.KeyF6
	case xkF7:
		return event.KeyF7
	case xkF8:
		return event.KeyF8
	case xkF9:
		return event.KeyF9
	case xkF10:
		return event.KeyF10
	case xkF11:
		return event.KeyF11
	case xkF12:
		return event.KeyF12
	case xkShiftL:
		return event.KeyLeftShift
	case xkShiftR:
		return event.KeyRightShift
	case xkControlL:
		return event.KeyLeftCtrl
	case xkControlR:
		return event.KeyRightCtrl
	case xkAltL:
		return event.KeyLeftAlt
	case xkAltR:
		return event.KeyRightAlt
	case xkSuperL:
		return event.KeyLeftSuper
	case xkSuperR:
		return event.KeyRightSuper
	case xkCapsLock:
		return event.KeyCapsLock
	case xkNumLock:
		return event.KeyNumLock
	case xkScrollLock:
		return event.KeyScrollLock
	case xkPrintScreen:
		return event.KeyPrintScreen
	case xkPause:
		return event.KeyPause
	case xkMenu:
		return event.KeyMenu
	case xkSpace:
		return event.KeySpace
	case xkApostrophe:
		return event.KeyApostrophe
	case xkComma:
		return event.KeyComma
	case xkMinus:
		return event.KeyMinus
	case xkPeriod:
		return event.KeyPeriod
	case xkSlash:
		return event.KeySlash
	case xkSemicolon:
		return event.KeySemicolon
	case xkEqual:
		return event.KeyEqual
	case xkBracketL:
		return event.KeyLeftBracket
	case xkBackslash:
		return event.KeyBackslash
	case xkBracketR:
		return event.KeyRightBracket
	case xkGrave:
		return event.KeyGraveAccent
	case xkKPAdd:
		return event.KeyNumpadAdd
	case xkKPSubtract:
		return event.KeyNumpadSubtract
	case xkKPMultiply:
		return event.KeyNumpadMultiply
	case xkKPDivide:
		return event.KeyNumpadDivide
	case xkKPDecimal:
		return event.KeyNumpadDecimal
	}

	// Letters A-Z (X11 keysyms 0x61-0x7A are lowercase a-z;
	// 0x41-0x5A are uppercase A-Z; both map to the same key)
	if keysym >= 0x61 && keysym <= 0x7A {
		// lowercase a-z → KeyA..KeyZ
		return event.KeyA + event.Key(keysym-0x61)
	}
	if keysym >= 0x41 && keysym <= 0x5A {
		// uppercase A-Z
		return event.KeyA + event.Key(keysym-0x41)
	}
	// Digits 0-9
	if keysym >= 0x30 && keysym <= 0x39 {
		return event.Key0 + event.Key(keysym-0x30)
	}
	// Numpad 0-9 (KP_0..KP_9)
	if keysym >= uint64(xkKP0) && keysym <= uint64(xkKP9) {
		return event.KeyNumpad0 + event.Key(keysym-uint64(xkKP0))
	}

	return event.KeyUnknown
}

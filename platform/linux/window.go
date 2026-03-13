//go:build linux && !android

package linux

import (
	"unsafe"

	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
)

// XC_ cursor font shape constants.
const (
	xcArrow          = 68  // XC_left_ptr
	xcIBeam          = 152 // XC_xterm
	xcCrosshair      = 34  // XC_crosshair
	xcHand           = 60  // XC_hand2
	xcHResize        = 108 // XC_sb_h_double_arrow
	xcVResize        = 116 // XC_sb_v_double_arrow
	xcSizing         = 120 // XC_sizing (used for NW/NE diagonal)
	xcFleur          = 52  // XC_fleur (all-resize)
	xcXCursor        = 0   // XC_X_cursor (not-allowed)
	xcWatch          = 150 // XC_watch
)

// Window implements platform.Window for Linux/X11.
type Window struct {
	p   *Platform
	dpy Display
	xwin XWindow

	width, height       int
	minWidth, minHeight int
	maxWidth, maxHeight int

	dpiScale    float32
	fullscreen  bool
	decorated   bool
	resizable   bool
	visible     bool
	deferredVisible bool
	shouldClose bool

	// Saved pre-fullscreen geometry
	savedX, savedY, savedW, savedH int

	// Per-cursor XID (pre-created on window init).
	cursors       [12]XID
	currentCursor platform.CursorShape

	// Position
	posX, posY int
}

// newWindow creates a new X11 window and sets it up.
func newWindow(p *Platform, opts platform.WindowOptions) (*Window, error) {
	root := XDefaultRootWindow(p.dpy)

	w := &Window{
		p:         p,
		dpy:       p.dpy,
		width:     opts.Width,
		height:    opts.Height,
		minWidth:  opts.MinWidth,
		minHeight: opts.MinHeight,
		maxWidth:  opts.MaxWidth,
		maxHeight: opts.MaxHeight,
		dpiScale:  1.0,
		decorated: opts.Decorated,
		resizable: opts.Resizable,
	}

	// Create the X11 window
	w.xwin = XCreateSimpleWindow(
		p.dpy, root,
		0, 0,
		uint(opts.Width), uint(opts.Height),
		0, // border width
		0, // border color (black)
		0, // background color (black)
	)

	// Register WM_DELETE_WINDOW protocol so close button sends ClientMessage
	wmDelete := p.wmDeleteWindow
	XSetWMProtocols(p.dpy, w.xwin, &wmDelete, 1)

	// Set window title
	if opts.Title != "" {
		w.SetTitle(opts.Title)
	}

	// Select event types we want to receive
	eventMask := KeyPressMask | KeyReleaseMask |
		ButtonPressMask | ButtonReleaseMask |
		PointerMotionMask |
		FocusChangeMask |
		ExposureMask |
		StructureNotifyMask

	XSelectInput(p.dpy, w.xwin, eventMask)

	// Set window manager size hints if constraints are given
	w.applySizeHints()

	// Pre-create cursors for each cursor shape
	w.cursors[platform.CursorArrow] = XCreateFontCursor(p.dpy, xcArrow)
	w.cursors[platform.CursorIBeam] = XCreateFontCursor(p.dpy, xcIBeam)
	w.cursors[platform.CursorCrosshair] = XCreateFontCursor(p.dpy, xcCrosshair)
	w.cursors[platform.CursorHand] = XCreateFontCursor(p.dpy, xcHand)
	w.cursors[platform.CursorHResize] = XCreateFontCursor(p.dpy, xcHResize)
	w.cursors[platform.CursorVResize] = XCreateFontCursor(p.dpy, xcVResize)
	w.cursors[platform.CursorNWSEResize] = XCreateFontCursor(p.dpy, xcSizing)
	w.cursors[platform.CursorNESWResize] = XCreateFontCursor(p.dpy, xcSizing)
	w.cursors[platform.CursorAllResize] = XCreateFontCursor(p.dpy, xcFleur)
	w.cursors[platform.CursorNotAllowed] = XCreateFontCursor(p.dpy, xcXCursor)
	w.cursors[platform.CursorWait] = XCreateFontCursor(p.dpy, xcWatch)
	// CursorNone: create invisible cursor using a 1x1 black pixmap (stub: use arrow)
	w.cursors[platform.CursorNone] = w.cursors[platform.CursorArrow]

	// Apply initial cursor
	XDefineCursor(p.dpy, w.xwin, w.cursors[platform.CursorArrow])

	// Handle visibility
	if opts.Visible {
		// Defer show until first PollEvents (after rendering is ready)
		w.deferredVisible = true
	}

	if opts.Fullscreen {
		w.SetFullscreen(true)
	}

	XFlush(p.dpy)
	return w, nil
}

// applySizeHints sets WM size hints if min/max constraints are set.
// WM_NORMAL_HINTS (XSizeHints) format — simplified via _NET hints.
func (w *Window) applySizeHints() {
	// Simplified: just flush for now. Full WM_NORMAL_HINTS requires XSetWMNormalHints.
	_ = w
}

// ---- platform.Window interface implementation ----

func (w *Window) Size() (int, int) {
	return w.width, w.height
}

func (w *Window) SetSize(width, height int) {
	w.width = width
	w.height = height
	XResizeWindow(w.dpy, w.xwin, uint(width), uint(height))
	XFlush(w.dpy)
}

// FramebufferSize returns the framebuffer size.
// On Linux without HiDPI scaling, framebuffer equals logical size.
func (w *Window) FramebufferSize() (int, int) {
	return w.width, w.height
}

func (w *Window) Position() (int, int) {
	return w.posX, w.posY
}

func (w *Window) SetPosition(x, y int) {
	w.posX = x
	w.posY = y
	XMoveWindow(w.dpy, w.xwin, x, y)
	XFlush(w.dpy)
}

func (w *Window) SetTitle(title string) {
	// Set both the legacy WM_NAME and the UTF-8 _NET_WM_NAME property.
	XStoreName(w.dpy, w.xwin, title)
	b := []byte(title)
	if len(b) > 0 {
		XChangeProperty(
			w.dpy, w.xwin,
			w.p.netWMName, w.p.utf8String,
			8, // format: 8-bit
			0, // PropModeReplace
			unsafe.Pointer(&b[0]), len(b),
		)
	}
}

func (w *Window) SetFullscreen(fullscreen bool) {
	if w.fullscreen == fullscreen {
		return
	}

	if fullscreen {
		// Save current geometry
		var attrs XWindowAttributes
		XGetWindowAttributes(w.dpy, w.xwin, &attrs)
		w.savedX = int(attrs.X)
		w.savedY = int(attrs.Y)
		w.savedW = int(attrs.Width)
		w.savedH = int(attrs.Height)

		// Send _NET_WM_STATE_FULLSCREEN via ClientMessage
		w.sendNetWMState(1, w.p.wmStateFullscreen)
	} else {
		// Remove fullscreen state
		w.sendNetWMState(0, w.p.wmStateFullscreen)
	}
	w.fullscreen = fullscreen
	XFlush(w.dpy)
}

// sendNetWMState sends a _NET_WM_STATE client message.
// action: 0=remove, 1=add, 2=toggle.
func (w *Window) sendNetWMState(action int64, atom Atom) {
	var xe XEvent
	cm := (*XClientMessageEvent)(unsafe.Pointer(&xe))
	cm.Type = ClientMessage
	cm.Window = w.xwin
	cm.MessageType = w.p.wmState
	cm.Format = 32
	cm.Data[0] = action
	cm.Data[1] = int64(atom)
	cm.Data[2] = 0

	root := XDefaultRootWindow(w.dpy)
	XSendEvent(w.dpy, root, 0,
		StructureNotifyMask,
		&xe,
	)
}

func (w *Window) IsFullscreen() bool {
	return w.fullscreen
}

func (w *Window) ShouldClose() bool {
	return w.shouldClose
}

func (w *Window) SetShouldClose(close bool) {
	w.shouldClose = close
}

// NativeHandle returns the X11 Window ID as a uintptr.
func (w *Window) NativeHandle() uintptr {
	return uintptr(w.xwin)
}

func (w *Window) DPIScale() float32 {
	return w.dpiScale
}

func (w *Window) SetVisible(visible bool) {
	w.visible = visible
	if visible {
		XMapWindow(w.dpy, w.xwin)
	} else {
		XUnmapWindow(w.dpy, w.xwin)
	}
	XFlush(w.dpy)
}

func (w *Window) ShowDeferred() {
	if w.deferredVisible {
		w.deferredVisible = false
		w.SetVisible(true)
	}
}

func (w *Window) SetMinSize(width, height int) {
	w.minWidth = width
	w.minHeight = height
}

func (w *Window) SetMaxSize(width, height int) {
	w.maxWidth = width
	w.maxHeight = height
}

func (w *Window) SetCursor(cursor platform.CursorShape) {
	w.currentCursor = cursor
	idx := int(cursor)
	if idx < 0 || idx >= len(w.cursors) {
		idx = int(platform.CursorArrow)
	}
	xid := w.cursors[idx]
	if xid == 0 {
		xid = w.cursors[platform.CursorArrow]
	}
	XDefineCursor(w.dpy, w.xwin, xid)
	XFlush(w.dpy)
}

func (w *Window) SetIMEPosition(caretRect uimath.Rect) {
	// Stub — full XIM integration is in ime.go
	_ = caretRect
}

func (w *Window) ShowContextMenu(clientX, clientY int, items []platform.ContextMenuItem) int {
	// X11 native context menus require GTK or Xt integration.
	// Stub: return -1 (cancelled).
	_ = clientX
	_ = clientY
	_ = items
	return -1
}

func (w *Window) ClientToScreen(x, y int) (int, int) {
	var destX, destY int
	var child XWindow
	root := XDefaultRootWindow(w.dpy)
	XTranslateCoordinates(w.dpy, w.xwin, root, x, y, &destX, &destY, &child)
	return destX, destY
}

func (w *Window) IsTransparent() bool              { return false }
func (w *Window) SetTopMost(topmost bool)           {}
func (w *Window) SetHitTestFunc(func(int, int) bool) {}

func (w *Window) Destroy() {
	// Free pre-created cursors
	for i, c := range w.cursors {
		if c != 0 && i != int(platform.CursorNone) {
			// Don't double-free CursorNone which aliases CursorArrow
			if platform.CursorShape(i) != platform.CursorNone {
				XFreeCursor(w.dpy, c)
			}
		}
		w.cursors[i] = 0
	}
	if w.xwin != 0 {
		XDestroyWindow(w.dpy, w.xwin)
		w.xwin = 0
	}
	XFlush(w.dpy)
}


// XDisplay returns the X11 Display* as a uintptr for Vulkan surface creation.
// This satisfies the vulkan.xlibWindowProvider interface.
func (w *Window) XDisplay() uintptr {
	return uintptr(w.dpy)
}

// Compile-time interface check.
var _ platform.Window = (*Window)(nil)

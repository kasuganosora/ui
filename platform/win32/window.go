//go:build windows

package win32

import (
	"unsafe"

	"github.com/kasuganosora/ui/platform"
	uimath "github.com/kasuganosora/ui/math"
)

// Window implements platform.Window using Win32 HWND.
type Window struct {
	hwnd       uintptr
	hinstance  uintptr
	p          *Platform // back-reference for event queue

	// State
	width, height   int // client area logical pixels
	fbWidth, fbHeight int // framebuffer physical pixels
	posX, posY      int
	dpiScale        float32
	fullscreen      bool
	decorated       bool
	resizable       bool
	visible         bool
	deferredVisible bool // true if ShowWindow is deferred until first PollEvents
	shouldClose     bool
	mouseTracked    bool // for WM_MOUSELEAVE tracking

	// Size constraints
	minWidth, minHeight int
	maxWidth, maxHeight int

	// Saved window state for fullscreen toggle
	savedStyle    uint32
	savedExStyle  uint32
	savedRect     RECT

	// Cursor
	currentCursor uintptr
	cursorInClient bool
}

// newWindow creates a new Win32 window. Called by Platform.CreateWindow.
func newWindow(p *Platform, opts platform.WindowOptions) (*Window, error) {
	w := &Window{
		p:         p,
		hinstance: p.hinstance,
		width:     opts.Width,
		height:    opts.Height,
		decorated: opts.Decorated,
		resizable: opts.Resizable,
		dpiScale:  1.0,
		minWidth:  opts.MinWidth,
		minHeight: opts.MinHeight,
		maxWidth:  opts.MaxWidth,
		maxHeight: opts.MaxHeight,
	}

	// Build window style
	style := uint32(WS_CLIPCHILDREN | WS_CLIPSIBLINGS)
	exStyle := uint32(WS_EX_APPWINDOW)

	if opts.Decorated {
		style |= WS_OVERLAPPEDWINDOW
		if !opts.Resizable {
			style &^= WS_THICKFRAME | WS_MAXIMIZEBOX
		}
	} else {
		style |= WS_POPUP
	}

	// Adjust rect for window decorations
	rect := RECT{
		Left:   0,
		Top:    0,
		Right:  int32(opts.Width),
		Bottom: int32(opts.Height),
	}
	procAdjustWindowRectEx.Call(
		uintptr(unsafe.Pointer(&rect)),
		uintptr(style),
		0, // no menu
		uintptr(exStyle),
	)

	windowWidth := int(rect.Right - rect.Left)
	windowHeight := int(rect.Bottom - rect.Top)

	// Create the window
	hwnd, _, _ := procCreateWindowExW.Call(
		uintptr(exStyle),
		uintptr(unsafe.Pointer(utf16PtrFromString(windowClassName))),
		uintptr(unsafe.Pointer(utf16PtrFromString(opts.Title))),
		uintptr(style),
		0x80000000, // CW_USEDEFAULT
		0x80000000,
		uintptr(windowWidth),
		uintptr(windowHeight),
		0, 0,
		w.hinstance,
		0,
	)
	if hwnd == 0 {
		return nil, lastError("CreateWindowExW")
	}
	w.hwnd = hwnd

	// Store pointer to Window in GWLP_USERDATA for WndProc lookup
	procSetWindowLongPtrW.Call(hwnd, uintptr(uint32ToUintptr(GWLP_USERDATA)), uintptr(unsafe.Pointer(w)))

	// Get DPI
	w.dpiScale = w.queryDPI()
	w.fbWidth = int(float32(w.width) * w.dpiScale)
	w.fbHeight = int(float32(w.height) * w.dpiScale)

	// Load default cursor
	w.currentCursor, _, _ = procLoadCursorW.Call(0, uintptr(IDC_ARROW))

	if opts.Visible {
		// Defer the actual ShowWindow call until the first PollEvents.
		// This keeps the window hidden during heavy initialization
		// (Vulkan, font loading, atlas) so Windows doesn't mark it
		// as "Not Responding" before the app enters its main loop.
		w.deferredVisible = true
	}

	if opts.Fullscreen {
		w.SetFullscreen(true)
	}

	return w, nil
}

func (w *Window) Size() (int, int) {
	return w.width, w.height
}

func (w *Window) SetSize(width, height int) {
	w.width = width
	w.height = height

	style, _, _ := procGetWindowLongPtrW.Call(w.hwnd, uintptr(uint32ToUintptr(GWLP_STYLE)))
	exStyle, _, _ := procGetWindowLongPtrW.Call(w.hwnd, uintptr(uint32ToUintptr(GWLP_EXSTYLE)))

	rect := RECT{Right: int32(width), Bottom: int32(height)}
	procAdjustWindowRectEx.Call(
		uintptr(unsafe.Pointer(&rect)),
		style, 0, exStyle,
	)

	procSetWindowPos.Call(w.hwnd, 0,
		0, 0,
		uintptr(rect.Right-rect.Left),
		uintptr(rect.Bottom-rect.Top),
		SWP_NOMOVE|SWP_NOZORDER,
	)
}

func (w *Window) FramebufferSize() (int, int) {
	return w.fbWidth, w.fbHeight
}

func (w *Window) Position() (int, int) {
	return w.posX, w.posY
}

func (w *Window) SetPosition(x, y int) {
	w.posX = x
	w.posY = y
	procSetWindowPos.Call(w.hwnd, 0,
		uintptr(x), uintptr(y), 0, 0,
		SWP_NOSIZE|SWP_NOZORDER,
	)
}

func (w *Window) SetTitle(title string) {
	procSetWindowTextW.Call(w.hwnd, uintptr(unsafe.Pointer(utf16PtrFromString(title))))
}

func (w *Window) SetFullscreen(fullscreen bool) {
	if w.fullscreen == fullscreen {
		return
	}

	if fullscreen {
		// Save current window state
		w.savedStyle = w.getStyle()
		w.savedExStyle = w.getExStyle()
		procGetWindowRect.Call(w.hwnd, uintptr(unsafe.Pointer(&w.savedRect)))

		// Set fullscreen style
		procSetWindowLongPtrW.Call(w.hwnd,
			uintptr(uint32ToUintptr(GWLP_STYLE)),
			uintptr(WS_POPUP|WS_VISIBLE),
		)

		// Get monitor size
		screenW, _, _ := procGetSystemMetrics.Call(SM_CXSCREEN)
		screenH, _, _ := procGetSystemMetrics.Call(SM_CYSCREEN)

		procSetWindowPos.Call(w.hwnd, 0,
			0, 0, screenW, screenH,
			SWP_FRAMECHANGED,
		)
	} else {
		// Restore saved state
		procSetWindowLongPtrW.Call(w.hwnd,
			uintptr(uint32ToUintptr(GWLP_STYLE)),
			uintptr(w.savedStyle),
		)
		procSetWindowLongPtrW.Call(w.hwnd,
			uintptr(uint32ToUintptr(GWLP_EXSTYLE)),
			uintptr(w.savedExStyle),
		)
		procMoveWindow.Call(w.hwnd,
			uintptr(w.savedRect.Left),
			uintptr(w.savedRect.Top),
			uintptr(w.savedRect.Right-w.savedRect.Left),
			uintptr(w.savedRect.Bottom-w.savedRect.Top),
			1,
		)
	}
	w.fullscreen = fullscreen
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

func (w *Window) NativeHandle() uintptr {
	return w.hwnd
}

func (w *Window) DPIScale() float32 {
	return w.dpiScale
}

func (w *Window) SetVisible(visible bool) {
	w.visible = visible
	if visible {
		procShowWindow.Call(w.hwnd, SW_SHOW)
		procUpdateWindow.Call(w.hwnd)
	} else {
		procShowWindow.Call(w.hwnd, SW_HIDE)
	}
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
	var id uint16
	switch cursor {
	case platform.CursorArrow:
		id = IDC_ARROW
	case platform.CursorIBeam:
		id = IDC_IBEAM
	case platform.CursorCrosshair:
		id = IDC_CROSS
	case platform.CursorHand:
		id = IDC_HAND
	case platform.CursorHResize:
		id = IDC_SIZEWE
	case platform.CursorVResize:
		id = IDC_SIZENS
	case platform.CursorNWSEResize:
		id = IDC_SIZENWSE
	case platform.CursorNESWResize:
		id = IDC_SIZENESW
	case platform.CursorAllResize:
		id = IDC_SIZEALL
	case platform.CursorNotAllowed:
		id = IDC_NO
	case platform.CursorWait:
		id = IDC_WAIT
	case platform.CursorNone:
		w.currentCursor = 0
		procSetCursor.Call(0)
		return
	default:
		id = IDC_ARROW
	}

	h, _, _ := procLoadCursorW.Call(0, uintptr(id))
	w.currentCursor = h
	procSetCursor.Call(h)
}

func (w *Window) SetIMEPosition(pos uimath.Vec2) {
	himc, _, _ := procImmGetContext.Call(w.hwnd)
	if himc == 0 {
		return
	}
	defer procImmReleaseContext.Call(w.hwnd, himc)

	cf := COMPOSITIONFORM{
		DwStyle:      2, // CFS_POINT
		PtCurrentPos: POINT{X: int32(pos.X), Y: int32(pos.Y)},
	}
	procImmSetCompositionWindow.Call(himc, uintptr(unsafe.Pointer(&cf)))

	cand := CANDIDATEFORM{
		DwIndex:      0,
		DwStyle:      0x0040, // CFS_CANDIDATEPOS
		PtCurrentPos: POINT{X: int32(pos.X), Y: int32(pos.Y)},
	}
	procImmSetCandidateWindow.Call(himc, uintptr(unsafe.Pointer(&cand)))
}

func (w *Window) Destroy() {
	if w.hwnd != 0 {
		procDestroyWindow.Call(w.hwnd)
		w.hwnd = 0
	}
}

// internal helpers

func (w *Window) getStyle() uint32 {
	r, _, _ := procGetWindowLongPtrW.Call(w.hwnd, uintptr(uint32ToUintptr(GWLP_STYLE)))
	return uint32(r)
}

func (w *Window) getExStyle() uint32 {
	r, _, _ := procGetWindowLongPtrW.Call(w.hwnd, uintptr(uint32ToUintptr(GWLP_EXSTYLE)))
	return uint32(r)
}

func (w *Window) queryDPI() float32 {
	// Try per-window DPI (Win10 1607+)
	if procGetDpiForWindow.Find() == nil {
		dpi, _, _ := procGetDpiForWindow.Call(w.hwnd)
		if dpi > 0 {
			return float32(dpi) / 96.0
		}
	}
	// Fallback: device caps
	hdc, _, _ := procGetDC.Call(w.hwnd)
	if hdc != 0 {
		dpi, _, _ := procGetDeviceCaps.Call(hdc, LOGPIXELSX)
		procReleaseDC.Call(w.hwnd, hdc)
		if dpi > 0 {
			return float32(dpi) / 96.0
		}
	}
	return 1.0
}

func (w *Window) updateClientSize() {
	var rect RECT
	procGetClientRect.Call(w.hwnd, uintptr(unsafe.Pointer(&rect)))
	w.width = int(rect.Right - rect.Left)
	w.height = int(rect.Bottom - rect.Top)
	w.fbWidth = int(float32(w.width) * w.dpiScale)
	w.fbHeight = int(float32(w.height) * w.dpiScale)
}

func (w *Window) updatePosition() {
	var rect RECT
	procGetWindowRect.Call(w.hwnd, uintptr(unsafe.Pointer(&rect)))
	w.posX = int(rect.Left)
	w.posY = int(rect.Top)
}

func (w *Window) enableMouseTracking() {
	if w.mouseTracked {
		return
	}
	tme := TRACKMOUSEEVENT{
		CbSize:    uint32(unsafe.Sizeof(TRACKMOUSEEVENT{})),
		DwFlags:   TME_LEAVE,
		HwndTrack: w.hwnd,
	}
	// TrackMouseEvent is in user32
	procTrackMouseEvent := user32.NewProc("TrackMouseEvent")
	procTrackMouseEvent.Call(uintptr(unsafe.Pointer(&tme)))
	w.mouseTracked = true
}

// uint32ToUintptr converts an int constant (possibly negative) to uintptr safely.
func uint32ToUintptr(v int) uintptr {
	return uintptr(uint32(int32(v)))
}

// Compile-time interface check.
var _ platform.Window = (*Window)(nil)

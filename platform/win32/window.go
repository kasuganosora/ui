//go:build windows

package win32

import (
	"unsafe"

	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
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
	currentCursor  uintptr
	cursorInClient bool
	cursorHandles  [12]uintptr // pre-loaded system cursor handles

	// IME position (client area coords, stored for use during WM_IME_STARTCOMPOSITION)
	imeX, imeY int32
	imeLineH   int32 // line height for candidate window exclusion rect

	// TSF (Text Services Framework) for modern IME support
	tsfMgr       *TSFManager
	tsfProvider  platform.TSFTextProvider

	// Resize callback — called from WndProc on WM_SIZE during the modal
	// resize loop so the app can re-layout and render while the user drags.
	inSizeMove     bool
	onResizeFunc   func()

	// Accessibility: UI Automation provider
	uiaProvider *UIAProvider
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

	// Pre-load all system cursor handles
	cursorIDs := [12]uint16{
		IDC_ARROW,    // CursorArrow     = 0
		IDC_IBEAM,    // CursorIBeam     = 1
		IDC_CROSS,    // CursorCrosshair = 2
		IDC_HAND,     // CursorHand      = 3
		IDC_SIZEWE,   // CursorHResize   = 4
		IDC_SIZENS,   // CursorVResize   = 5
		IDC_SIZENWSE, // CursorNWSEResize= 6
		IDC_SIZENESW, // CursorNESWResize= 7
		IDC_SIZEALL,  // CursorAllResize = 8
		IDC_NO,       // CursorNotAllowed= 9
		IDC_WAIT,     // CursorWait      = 10
		0,            // CursorNone      = 11
	}
	for i, id := range cursorIDs {
		if id != 0 {
			w.cursorHandles[i], _, _ = procLoadCursorW.Call(0, uintptr(id))
		}
	}
	// Replace system IBeam with a custom cursor that has a black stem
	// and white outline, avoiding the XOR-inversion visibility problem.
	if h := createCustomIBeamCursor(w.hinstance); h != 0 {
		w.cursorHandles[platform.CursorIBeam] = h
	}
	w.currentCursor = w.cursorHandles[platform.CursorArrow]

	if opts.Visible {
		// Defer the actual ShowWindow call until the first PollEvents.
		// This keeps the window hidden during heavy initialization
		// (Vulkan, font loading, atlas) so Windows doesn't mark it
		// as "Not Responding" before the app enters its main loop.
		w.deferredVisible = true
	}

	// Initialize TSF for modern IME support.
	// If TSF is unavailable, tsfMgr.IsActive() will be false and
	// the existing IMM32 code path in ime.go handles everything.
	w.tsfMgr = NewTSFManager(w)

	// Initialize UI Automation provider for accessibility (screen readers).
	w.uiaProvider = NewUIAProvider(w.hwnd)

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

// OnResize sets a callback that fires on every WM_SIZE during modal resize
// (user dragging the window border). This allows the app to re-layout and
// render frames while the Win32 modal resize loop blocks the main loop.
func (w *Window) OnResize(fn func()) {
	w.onResizeFunc = fn
}

// InSizeMove returns true while the user is dragging the window border
// (between WM_ENTERSIZEMOVE and WM_EXITSIZEMOVE).
func (w *Window) InSizeMove() bool {
	return w.inSizeMove
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

// createCustomIBeamCursor loads the system IDC_IBEAM cursor, reads its
// monochrome mask, converts the XOR-inverted pixels to solid black, and
// adds a 1px white outline around them. This fixes the visibility problem
// where the system IBeam uses XOR pixel inversion and becomes invisible
// on mid-tone backgrounds.
func createCustomIBeamCursor(hinstance uintptr) uintptr {
	// Load the system IBeam cursor
	sysCursor, _, _ := procLoadCursorW.Call(0, uintptr(IDC_IBEAM))
	if sysCursor == 0 {
		return 0
	}

	// Get cursor info (hotspot + mask bitmap)
	var info ICONINFO
	ret, _, _ := procGetIconInfo.Call(sysCursor, uintptr(unsafe.Pointer(&info)))
	if ret == 0 {
		return 0
	}
	// GetIconInfo creates copies of bitmaps; we must delete them when done.
	defer func() {
		if info.HbmMask != 0 {
			procDeleteObject.Call(info.HbmMask)
		}
		if info.HbmColor != 0 {
			procDeleteObject.Call(info.HbmColor)
		}
	}()

	// Get mask bitmap dimensions
	var bm BITMAP
	ret, _, _ = procGetObject.Call(info.HbmMask, unsafe.Sizeof(bm), uintptr(unsafe.Pointer(&bm)))
	if ret == 0 {
		return 0
	}

	w := int(bm.BmWidth)
	// For monochrome cursors (no color bitmap), the mask is double-height:
	// top half = AND mask, bottom half = XOR mask
	isMonochrome := info.HbmColor == 0
	h := int(bm.BmHeight)
	cursorH := h
	if isMonochrome {
		cursorH = h / 2
	}
	bytesPerRow := int(bm.BmWidthBytes)
	totalBytes := int(bm.BmWidthBytes) * int(bm.BmHeight)

	// Read mask bits
	maskBits := make([]byte, totalBytes)
	ret, _, _ = procGetBitmapBits.Call(info.HbmMask, uintptr(totalBytes), uintptr(unsafe.Pointer(&maskBits[0])))
	if ret == 0 {
		return 0
	}

	// Helper to read a bit from the mask
	getBit := func(data []byte, x, y, stride int) bool {
		return data[y*stride+x/8]&(0x80>>uint(x%8)) != 0
	}

	// Identify "cursor shape" pixels from the original masks.
	// For monochrome: AND in top half, XOR in bottom half
	// Cursor shape pixels are where XOR=1 (either inverted or white).
	// We also consider AND=0,XOR=0 as shape (black pixels).
	isShape := func(x, y int) bool {
		if x < 0 || x >= w || y < 0 || y >= cursorH {
			return false
		}
		if isMonochrome {
			andBit := getBit(maskBits, x, y, bytesPerRow)
			xorBit := getBit(maskBits, x, y+cursorH, bytesPerRow)
			// Shape = inverted(AND=1,XOR=1) or black(AND=0,XOR=0) or white(AND=0,XOR=1)
			// Transparent = AND=1,XOR=0
			return !(andBit && !xorBit)
		}
		// Color cursor: AND=0 means opaque
		andBit := getBit(maskBits, x, y, bytesPerRow)
		return !andBit
	}

	// Build new AND and XOR masks:
	// - Shape pixels → black (AND=0, XOR=0)
	// - 1px outline around shape → white (AND=0, XOR=1)
	// - Everything else → transparent (AND=1, XOR=0)
	newSize := bytesPerRow * cursorH
	newAND := make([]byte, newSize)
	newXOR := make([]byte, newSize)

	// Start all transparent
	for i := range newAND {
		newAND[i] = 0xFF
	}

	setBit := func(data []byte, x, y, stride int, val bool) {
		idx := y*stride + x/8
		mask := byte(0x80 >> uint(x%8))
		if val {
			data[idx] |= mask
		} else {
			data[idx] &^= mask
		}
	}

	// First pass: add white outline (1px border around every shape pixel)
	for y := 0; y < cursorH; y++ {
		for x := 0; x < w; x++ {
			if isShape(x, y) {
				// Expand outline to neighbors
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						nx, ny := x+dx, y+dy
						if nx >= 0 && nx < w && ny >= 0 && ny < cursorH {
							setBit(newAND, nx, ny, bytesPerRow, false) // AND=0 (opaque)
							setBit(newXOR, nx, ny, bytesPerRow, true)  // XOR=1 (white)
						}
					}
				}
			}
		}
	}

	// Second pass: shape pixels → black (overwrite the white outline center)
	for y := 0; y < cursorH; y++ {
		for x := 0; x < w; x++ {
			if isShape(x, y) {
				setBit(newAND, x, y, bytesPerRow, false) // AND=0 (opaque)
				setBit(newXOR, x, y, bytesPerRow, false) // XOR=0 (black)
			}
		}
	}

	// Create new monochrome mask bitmap (double height: AND on top, XOR on bottom)
	combinedMask := make([]byte, newSize*2)
	copy(combinedMask[:newSize], newAND)
	copy(combinedMask[newSize:], newXOR)

	hbmMask, _, _ := procCreateBitmap.Call(
		uintptr(w), uintptr(cursorH*2), 1, 1,
		uintptr(unsafe.Pointer(&combinedMask[0])),
	)
	if hbmMask == 0 {
		return 0
	}

	newInfo := ICONINFO{
		FIcon:    0, // cursor
		XHotspot: info.XHotspot,
		YHotspot: info.YHotspot,
		HbmMask:  hbmMask,
		HbmColor: 0, // monochrome
	}
	cursor, _, _ := procCreateIconIndirect.Call(uintptr(unsafe.Pointer(&newInfo)))
	procDeleteObject.Call(hbmMask)
	return cursor
}

func (w *Window) SetCursor(cursor platform.CursorShape) {
	if cursor == platform.CursorNone {
		w.currentCursor = 0
		procSetCursor.Call(0)
		return
	}
	idx := int(cursor)
	if idx < 0 || idx >= len(w.cursorHandles) {
		idx = int(platform.CursorArrow)
	}
	h := w.cursorHandles[idx]
	if h == 0 {
		h = w.cursorHandles[platform.CursorArrow]
	}
	w.currentCursor = h
	procSetCursor.Call(h)
}

func (w *Window) SetIMEPosition(caretRect uimath.Rect) {
	// Store position for use during WM_IME_STARTCOMPOSITION
	w.imeX = int32(caretRect.X)
	w.imeY = int32(caretRect.Y)
	w.imeLineH = int32(caretRect.Height)
	if w.imeLineH < 1 {
		w.imeLineH = 20
	}
	w.applyIMEPosition()
}

// applyIMEPosition sets the IME composition and candidate window positions.
func (w *Window) applyIMEPosition() {
	himc, _, _ := procImmGetContext.Call(w.hwnd)
	if himc == 0 {
		return
	}
	defer procImmReleaseContext.Call(w.hwnd, himc)

	pt := POINT{X: w.imeX, Y: w.imeY}

	// Position the inline composition string at the cursor
	cf := COMPOSITIONFORM{
		DwStyle:      2, // CFS_POINT
		PtCurrentPos: pt,
	}
	procImmSetCompositionWindow.Call(himc, uintptr(unsafe.Pointer(&cf)))

	// Use CFS_EXCLUDE to tell the IME to place the candidate window
	// avoiding the text line rectangle. This works with modern IMEs
	// (Microsoft Pinyin, etc.) that ignore CFS_CANDIDATEPOS.
	for i := uint32(0); i < 4; i++ {
		cand := CANDIDATEFORM{
			DwIndex:      i,
			DwStyle:      0x0080, // CFS_EXCLUDE
			PtCurrentPos: pt,
			RcArea: RECT{
				Left:   w.imeX,
				Top:    w.imeY,
				Right:  w.imeX + 1,
				Bottom: w.imeY + w.imeLineH,
			},
		}
		procImmSetCandidateWindow.Call(himc, uintptr(unsafe.Pointer(&cand)))
	}
}

func (w *Window) ClientToScreen(x, y int) (int, int) {
	pt := POINT{X: int32(x), Y: int32(y)}
	procClientToScreen.Call(w.hwnd, uintptr(unsafe.Pointer(&pt)))
	return int(pt.X), int(pt.Y)
}

func (w *Window) ShowContextMenu(clientX, clientY int, items []platform.ContextMenuItem) int {
	hMenu, _, _ := procCreatePopupMenu.Call()
	if hMenu == 0 {
		return -1
	}
	defer procDestroyMenu.Call(hMenu)

	for i, item := range items {
		flags := uintptr(MF_STRING)
		if !item.Enabled {
			flags |= MF_GRAYED
		}
		procAppendMenuW.Call(hMenu, flags, uintptr(i+1),
			uintptr(unsafe.Pointer(utf16PtrFromString(item.Label))))
	}

	// Convert to screen coordinates
	sx, sy := w.ClientToScreen(clientX, clientY)

	ret, _, _ := procTrackPopupMenu.Call(hMenu,
		TPM_RETURNCMD|TPM_RIGHTBUTTON,
		uintptr(sx), uintptr(sy),
		0, w.hwnd, 0)
	if ret == 0 {
		return -1
	}
	return int(ret) - 1
}

// TSF returns the TSF manager for this window, or nil if unavailable.
func (w *Window) TSF() *TSFManager {
	return w.tsfMgr
}

func (w *Window) Destroy() {
	if w.tsfMgr != nil {
		w.tsfMgr.Release()
		w.tsfMgr = nil
	}
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

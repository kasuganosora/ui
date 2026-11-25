//go:build darwin

package darwin

import (
	"fmt"
	"unsafe"

	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
)

// NSWindowLevel values (AppKit). A floating auxiliary panel sits above
// ordinary windows but below screen saver / fullscreen layers.
const (
	nsNormalWindowLevel    = 0
	nsPopUpMenuWindowLevel = 101
)

// Window implements platform.Window using Cocoa NSWindow.
type Window struct {
	nswindow     id
	nsview       id
	trackingArea id
	p            *Platform

	// State
	width, height     int
	fbWidth, fbHeight int
	posX, posY        int
	dpiScale          float32
	fullscreen        bool
	decorated         bool
	resizable         bool
	transparent       bool
	topMost           bool
	passthrough       bool
	visible           bool
	deferredVisible   bool
	shouldClose       bool
	mouseDown         bool
	lastModifiers     event.Modifiers
	hitTest           func(x, y int) bool

	// Size constraints
	minWidth, minHeight int
	maxWidth, maxHeight int

	// Saved window state for fullscreen toggle
	savedFrame     NSRect
	savedStyleMask uint64

	// Cursor
	currentCursor platform.CursorShape
	cursorHidden  bool

	// IME state
	imeX, imeY    int32
	imeLineH      int32
	hasMarkedText bool
	markedText    string
}

// initWindowWithContentRect creates an NSWindow using the designated initializer.
// NOTE: NSWindow's plain -init path is not safe for our purego runtime bridge.
func initWindowWithContentRect(rect NSRect, styleMask uint64) id {
	nsWindowClass := id(objcClass("NSWindow"))
	alloced := msgSend(nsWindowClass, selAlloc)
	if alloced == 0 {
		return 0
	}

	return msgSendInitWindowRect(
		alloced,
		selInitWithContentRect,
		rect,
		styleMask,
		NSBackingStoreBuffered,
		false, // defer = NO
	)
}

// newWindow creates a new Cocoa window.
func newWindow(p *Platform, opts platform.WindowOptions) (*Window, error) {
	w := &Window{
		p:           p,
		width:       opts.Width,
		height:      opts.Height,
		decorated:   opts.Decorated,
		resizable:   opts.Resizable,
		transparent: opts.Transparent,
		topMost:     opts.TopMost,
		passthrough: opts.Passthrough,
		dpiScale:    1.0,
		minWidth:    opts.MinWidth,
		minHeight:   opts.MinHeight,
		maxWidth:    opts.MaxWidth,
		maxHeight:   opts.MaxHeight,
	}

	// Build window style mask
	styleMask := uint64(NSWindowStyleMaskTitled | NSWindowStyleMaskClosable | NSWindowStyleMaskMiniaturizable)
	if opts.Resizable {
		styleMask |= NSWindowStyleMaskResizable
	}
	if !opts.Decorated {
		styleMask = NSWindowStyleMaskBorderless
	}

	// Create window rect (centered on main screen)
	screen := msgSend(id(classNSScreen), selMainScreen)
	var screenFrame NSRect
	if screen != 0 {
		screenFrame = msgSendRectReturn(screen, selFrame)
	}

	// Center the window on screen
	centerX := (screenFrame.Size.Width - float64(opts.Width)) / 2
	centerY := (screenFrame.Size.Height - float64(opts.Height)) / 2

	contentRect := nsRect(centerX, centerY, float64(opts.Width), float64(opts.Height))
	w.posX = int(centerX)
	w.posY = int(centerY)

	// Create the window: initWithContentRect:styleMask:backing:defer:
	w.nswindow = initWindowWithContentRect(contentRect, styleMask)

	if w.nswindow == 0 {
		return nil, fmt.Errorf("darwin: failed to create NSWindow")
	}

	// Get the content view
	w.nsview = msgSend(w.nswindow, selContentView)

	// Calculate DPI scale (must be AFTER window creation so nswindow.screen is valid)
	w.dpiScale = w.queryDPI() / 72.0
	if w.dpiScale <= 0 {
		w.dpiScale = 1.0
	}
	w.fbWidth = int(float32(w.width) * w.dpiScale)
	w.fbHeight = int(float32(w.height) * w.dpiScale)

	// Set window title
	title := nsString(opts.Title)
	msgSend(w.nswindow, selSetTitle, uintptr(title))
	msgSend(title, selRelease)

	// Set min/max size constraints
	if opts.MinWidth > 0 || opts.MinHeight > 0 {
		minSize := nsSize(float64(opts.MinWidth), float64(opts.MinHeight))
		msgSendSizeArg(w.nswindow, selSetMinSize, minSize)
	}
	if opts.MaxWidth > 0 || opts.MaxHeight > 0 {
		maxSize := nsSize(float64(opts.MaxWidth), float64(opts.MaxHeight))
		msgSendSizeArg(w.nswindow, selSetMaxSize, maxSize)
	}

	// Set opacity and background color. Transparent windows need a clear
	// background and an opaque:NO layer so the render backend's alpha clear
	// actually shows through; otherwise the window is a solid white rectangle.
	if opts.Transparent {
		msgSend(w.nswindow, selSetOpaque, 0)
		clearColor := msgSend(id(classNSColor), selClearColor)
		msgSend(w.nswindow, selSetBackgroundColor, uintptr(clearColor))
		// Borderless floating panels have no shadow of their own; the UI draws it.
		msgSend(w.nswindow, selSetHasShadow, 0)
	} else {
		msgSend(w.nswindow, selSetOpaque, 1)
		whiteColor := msgSend(id(classNSColor), objcSelector("whiteColor"))
		msgSend(w.nswindow, selSetBackgroundColor, uintptr(whiteColor))
	}
	if opts.TopMost {
		// NSPopUpMenuWindowLevel keeps auxiliary panels above ordinary windows
		// without making them key windows.
		msgSend(w.nswindow, selSetLevel, nsPopUpMenuWindowLevel)
	}
	if opts.Passthrough {
		// Auxiliary floating panels must not swallow clicks meant for the
		// window below. Transparent windows with per-pixel hit testing keep
		// normal input delivery and filter events in Go instead.
		msgSend(w.nswindow, selSetIgnoresMouseEvents, 1)
	}

	// Create tracking area for mouse enter/exit events
	// NOTE: Disabled temporarily on darwin to avoid ObjC struct-call ABI issues in purego path.
	// w.createTrackingArea()

	// Handle visibility
	if opts.Visible {
		w.deferredVisible = true
	}

	// Setup fullscreen if requested
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
	w.fbWidth = int(float32(width) * w.dpiScale)
	w.fbHeight = int(float32(height) * w.dpiScale)

	// Get current frame
	frame := msgSendRectReturn(w.nswindow, selFrame)

	// Calculate new frame rect that gives us the desired content size
	contentRect := nsRect(0, 0, float64(width), float64(height))
	newFrameRect := msgSendRectArgReturnID(
		w.nswindow,
		selFrameRectForContentRect,
		contentRect,
	)

	// Keep the same top-left position
	newFrameRect.Origin = frame.Origin

	// Update the frame
	msgSendRectArgDisplay(
		w.nswindow,
		selSetFrame,
		newFrameRect,
		true,
	) // display YES
}

func (w *Window) FramebufferSize() (int, int) {
	return w.fbWidth, w.fbHeight
}

func (w *Window) Position() (int, int) {
	frame := msgSendRectReturn(w.nswindow, selFrame)
	w.posX = int(frame.Origin.X)
	w.posY = int(frame.Origin.Y)
	return w.posX, w.posY
}

func (w *Window) SetPosition(x, y int) {
	w.posX = x
	w.posY = y

	// Convert to top-left based coordinates (Cocoa uses bottom-left)
	screen := msgSend(w.nswindow, selScreen)
	if screen != 0 {
		screenFrame := msgSendRectReturn(screen, selFrame)
		y = int(screenFrame.Size.Height) - y - w.height
	}

	msgSendPointArg(w.nswindow, selSetFrameTopLeftPoint, nsPoint(float64(x), float64(y)))
}

func (w *Window) SetTitle(title string) {
	nsTitle := nsString(title)
	msgSend(w.nswindow, selSetTitle, uintptr(nsTitle))
	msgSend(nsTitle, selRelease)
}

func (w *Window) SetFullscreen(fullscreen bool) {
	if w.fullscreen == fullscreen {
		return
	}

	msgSend(w.nswindow, selToggleFullScreen, 0)
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
	if close {
		msgSend(w.nswindow, selClose)
	}
}

func (w *Window) NativeHandle() uintptr {
	// Render backends (Metal, Vulkan/MoltenVK) expect an NSView, not NSWindow.
	return uintptr(w.nsview)
}

// NativeWindowHandle returns the NSWindow pointer for Cocoa-level operations.
func (w *Window) NativeWindowHandle() uintptr {
	return uintptr(w.nswindow)
}

func (w *Window) DPIScale() float32 {
	return w.dpiScale
}

func (w *Window) SetVisible(visible bool) {
	w.visible = visible
	if visible {
		// makeKeyAndOrderFront: makes the window the key window and orders it front.
		msgSend(w.nswindow, selMakeKeyAndOrderFront, 0)
		// orderFrontRegardless forces the window to the front even if the app
		// is not active (essential for non-bundle CLI apps on macOS).
		msgSend(w.nswindow, objcSelector("orderFrontRegardless"))
		// Activate the application so it becomes frontmost.
		ActivateApplication()
	} else {
		// orderOut: hides without releasing the NSWindow, so a floating panel
		// can be shown again on the next hover. close would tear the window
		// down and make Show() a no-op.
		msgSend(w.nswindow, selOrderOut, 0)
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
	minSize := nsSize(float64(width), float64(height))
	msgSendSizeArg(w.nswindow, selSetMinSize, minSize)
}

func (w *Window) SetMaxSize(width, height int) {
	w.maxWidth = width
	w.maxHeight = height
	maxSize := nsSize(float64(width), float64(height))
	msgSendSizeArg(w.nswindow, selSetMaxSize, maxSize)
}

func (w *Window) SetCursor(cursor platform.CursorShape) {
	w.currentCursor = cursor

	switch cursor {
	case platform.CursorNone:
		msgSend(id(classNSCursor), selHide)
		w.cursorHidden = true
	case platform.CursorArrow:
		arrow := msgSend(id(classNSCursor), selArrowCursor)
		msgSend(arrow, selSet)
		w.cursorHidden = false
	case platform.CursorIBeam:
		ibeam := msgSend(id(classNSCursor), selIBeamCursor)
		msgSend(ibeam, selSet)
		w.cursorHidden = false
	case platform.CursorCrosshair:
		crosshair := msgSend(id(classNSCursor), selCrosshairCursor)
		msgSend(crosshair, selSet)
		w.cursorHidden = false
	case platform.CursorHand:
		pointingHand := msgSend(id(classNSCursor), selPointingHandCursor)
		msgSend(pointingHand, selSet)
		w.cursorHidden = false
	case platform.CursorHResize:
		resizeLeftRight := msgSend(id(classNSCursor), selResizeLeftRightCursor)
		msgSend(resizeLeftRight, selSet)
		w.cursorHidden = false
	case platform.CursorVResize:
		resizeUpDown := msgSend(id(classNSCursor), selResizeUpDownCursor)
		msgSend(resizeUpDown, selSet)
		w.cursorHidden = false
	default:
		// Default to arrow for unsupported cursors
		arrow := msgSend(id(classNSCursor), selArrowCursor)
		msgSend(arrow, selSet)
		w.cursorHidden = false
	}

	if w.cursorHidden && cursor != platform.CursorNone {
		msgSend(id(classNSCursor), selUnhide)
	}
}

func (w *Window) SetIMEPosition(caretRect uimath.Rect) {
	// macOS IME position is set through the input context
	// The position is relative to the view coordinates
	// Convert to Cocoa's bottom-left origin
	cocoaY := float64(w.height) - float64(caretRect.Y) - float64(caretRect.Height)

	// Store the IME rect for the input method handler
	w.imeX = int32(caretRect.X)
	w.imeY = int32(cocoaY)
	w.imeLineH = int32(caretRect.Height)

	// Update the input method context if we have marked text
	if w.hasMarkedText {
		w.updateIMECursorPosition()
	}
}

// updateIMECursorPosition updates the IME cursor position.
func (w *Window) updateIMECursorPosition() {
	// Get the input context for this view
	inputContext := msgSend(w.nsview, objcSelector("inputContext"))
	if inputContext == 0 {
		return
	}

	// Invalidate the character coordinates so the IME fetches new ones
	msgSend(inputContext, objcSelector("invalidateCharacterCoordinates"))
}

func (w *Window) ClientToScreen(x, y int) (int, int) {
	// Convert from view coordinates to window coordinates
	point := nsPoint(float64(x), float64(w.height)-float64(y))

	// Convert to screen coordinates
	screenPoint := msgSend(w.nswindow, selConvertRectToScreen,
		*(*uintptr)(unsafe.Pointer(&point.X)),
		*(*uintptr)(unsafe.Pointer(&point.Y)),
		0, 0) // width=0, height=0

	screenPt := pointFromPtr(unsafe.Pointer(&screenPoint))
	return int(screenPt.X), int(screenPt.Y)
}

// ClientToTopLeftScreen converts client coordinates to top-left-origin screen
// coordinates, which is the convention used by SetPosition. ClientToScreen
// returns Cocoa's bottom-left-origin coordinates (used by ShowContextMenu).
func (w *Window) ClientToTopLeftScreen(x, y int) (int, int) {
	sx, sy := w.ClientToScreen(x, y)
	screen := msgSend(w.nswindow, selScreen)
	if screen != 0 {
		frame := msgSendRectReturn(screen, selFrame)
		sy = int(frame.Size.Height) - sy
	}
	return sx, sy
}

func (w *Window) ShowContextMenu(clientX, clientY int, items []platform.ContextMenuItem) int {
	// Create an NSMenu
	menu := msgSend(msgSend(id(classNSMenu), selAlloc), selInitWithTitle, 0)
	defer msgSend(menu, selRelease)

	// Add menu items
	for i, item := range items {
		var menuItem id
		if item.Label == "-" {
			// Separator item
			menuItem = msgSend(id(classNSMenuItem), selSeparatorItem)
		} else {
			itemTitle := nsString(item.Label)
			menuItem = msgSend(
				msgSend(id(classNSMenuItem), selAlloc),
				selInitWithTitle,
				uintptr(itemTitle),
				0, // no action
				0) // no key equivalent
			msgSend(itemTitle, selRelease)

			if !item.Enabled {
				msgSend(menuItem, selSetEnabled, 0)
			} else {
				msgSend(menuItem, selSetEnabled, 1)
			}

			// Set tag to identify the item
			msgSend(menuItem, selSetTag, uintptr(i))
		}

		msgSend(menu, selAddItem, uintptr(menuItem))
		msgSend(menuItem, selRelease)
	}

	// Convert client coordinates to screen coordinates
	screenX, screenY := w.ClientToScreen(clientX, clientY)

	// Position for the menu (flip Y for Cocoa)
	screen := msgSend(w.nswindow, selScreen)
	var screenFrame NSRect
	if screen != 0 {
		screenFrame = msgSendRectReturn(screen, selFrame)
	}
	menuY := int(screenFrame.Size.Height) - screenY

	// Pop up the menu
	// popUpMenuPositioningItem:atLocation:inView:callbackNumber:callback:selector:
	location := nsPoint(float64(screenX), float64(menuY))
	msgSend(menu, selPopUpMenuPositioningItem,
		0, // no positioning item
		*(*uintptr)(unsafe.Pointer(&location.X)),
		*(*uintptr)(unsafe.Pointer(&location.Y)),
		0,       // inView = nil (screen coordinates)
		0, 0, 0) // no callback

	// Get the selected item
	// On macOS, we need to track this differently since popUpMenuPositioningItem is async
	// For simplicity, we return -1 indicating no selection for now
	// A full implementation would use a callback or delegate
	return -1
}

func (w *Window) IsTransparent() bool { return w.transparent }

func (w *Window) SetTopMost(topmost bool) {
	w.topMost = topmost
	level := uintptr(nsNormalWindowLevel)
	if topmost {
		// Must match newWindow(opts.TopMost): a later SetTopMost call on an
		// auxiliary panel must not lower it to a weaker floating level.
		level = nsPopUpMenuWindowLevel
	}
	msgSend(w.nswindow, selSetLevel, level)
}

// SetHitTestFunc stores a per-pixel predicate. Events at rejected points are
// dropped by Platform.convertAndStoreEvent, so transparent shaped windows can
// let clicks through pixel by pixel without a custom NSView subclass.
func (w *Window) SetHitTestFunc(fn func(int, int) bool) {
	w.hitTest = fn
}

func (w *Window) Destroy() {
	if w.trackingArea != 0 {
		msgSend(w.nsview, selRemoveTrackingArea, uintptr(w.trackingArea))
		msgSend(w.trackingArea, selRelease)
		w.trackingArea = 0
	}

	if w.nswindow != 0 {
		msgSend(w.nswindow, selClose)
		w.nswindow = 0
	}
}

// createTrackingArea creates a tracking area for mouse enter/exit events.
func (w *Window) createTrackingArea() {
	// Create tracking area for the entire view
	bounds := nsRect(0, 0, float64(w.width), float64(w.height))

	w.trackingArea = msgSendInitTrackingAreaRect(
		msgSend(id(objcClass("NSTrackingArea")), selAlloc),
		selInitWithRect,
		bounds,
		NSTrackingMouseEnteredAndExited|NSTrackingMouseMoved|NSTrackingActiveAlways,
		w.nsview, // owner
		0,
	) // userInfo

	if w.trackingArea != 0 {
		msgSend(w.nsview, selAddTrackingArea, uintptr(w.trackingArea))
	}
}

// queryDPI returns the DPI for this window's screen.
func (w *Window) queryDPI() float32 {
	screen := msgSend(w.nswindow, selScreen)
	if screen == 0 {
		screen = msgSend(id(classNSScreen), selMainScreen)
	}
	if screen == 0 {
		return 72.0 // Default macOS DPI
	}

	scale := msgSendFloat64Return(screen, selBackingScaleFactor)
	return float32(scale * 72.0)
}

// updateContentSize updates the cached content size from the window.
func (w *Window) updateContentSize() {
	// Get the content view bounds
	bounds := msgSendRectReturn(w.nsview, selBounds)

	w.width = int(bounds.Size.Width)
	w.height = int(bounds.Size.Height)
	w.fbWidth = int(float32(w.width) * w.dpiScale)
	w.fbHeight = int(float32(w.height) * w.dpiScale)
}

// updatePosition updates the cached position from the window.
func (w *Window) updatePosition() {
	frame := msgSendRectReturn(w.nswindow, selFrame)
	w.posX = int(frame.Origin.X)
	w.posY = int(frame.Origin.Y)
}

// Compile-time interface check.
var _ platform.Window = (*Window)(nil)

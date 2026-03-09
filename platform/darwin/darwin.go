//go:build darwin

// Package darwin implements the platform.Platform interface using Cocoa APIs via purego.
// This is a DDD anti-corruption layer: all Cocoa specifics are translated into the
// platform domain model (Window, Event).
// Zero CGO — all system calls go through purego.
package darwin

import (
	"github.com/ebitengine/purego"
	"unsafe"
)

// Foundation Framework symbols
var (
	objc_msgSend        uintptr
	objc_getClass       uintptr
	sel_registerName    uintptr
	class_getName       uintptr
	object_getClass     uintptr
)

// NSObject selectors
var (
	selAlloc           SEL
	selInit            SEL
	selRelease         SEL
	selAutorelease     SEL
	selRetain          SEL
	selClass           SEL
	selDescription     SEL
	selUTF8String      SEL
)

// NSString selectors
var (
	selStringWithUTF8String SEL
	selLength               SEL
	selCharacterAtIndex     SEL
)

// NSApplication selectors
var (
	selSharedApplication    SEL
	selRun                  SEL
	selTerminate            SEL
	selSetActivationPolicy  SEL
	selActivateIgnoringOtherApps SEL
)

// NSWindow selectors
var (
	selInitWithContentRect  SEL
	selMakeKeyAndOrderFront SEL
	selClose                SEL
	selSetTitle             SEL
	selContentView          SEL
	selFrame                SEL
	selSetFrame             SEL
	selSetFrameTopLeftPoint SEL
	selContentRectForFrameRect SEL
	selFrameRectForContentRect SEL
	selMinSize           SEL
	selSetMinSize           SEL
	selMaxSize           SEL
	selSetMaxSize           SEL
	selToggleFullScreen     SEL
	selIsZoomed             SEL
	selSetStyleMask         SEL
	selStyleMask            SEL
	selSetOpaque            SEL
	selSetBackgroundColor   SEL
	selMakeFirstResponder   SEL
	selConvertRectToScreen  SEL
	selScreen               SEL
	selWindow               SEL
)

// NSView selectors
var (
	selSetWantsLayer        SEL
	selLayer                SEL
	selSetLayer             SEL
	selBounds               SEL
	selSetBounds            SEL
	selConvertRectFromView  SEL
	selConvertRectToView    SEL
)

// NSResponder selectors
var (
	selAcceptsFirstResponder SEL
	selBecomeFirstResponder  SEL
	selResignFirstResponder  SEL
)

// NSEvent selectors
var (
	selType           SEL
	selLocationInWindow SEL
	selModifierFlags  SEL
	selButtonNumber   SEL
	selClickCount     SEL
	selKeyCode        SEL
	selCharacters     SEL
	selDeltaX         SEL
	selDeltaY         SEL
	selTimestamp      SEL
)

// NSTextInputClient selectors (for IME)
var (
	selInsertText                     SEL
	selSetMarkedText                  SEL
	selUnmarkText                     SEL
	selSelectedRange                  SEL
	selMarkedRange                    SEL
	selHasMarkedText                  SEL
	selAttributedSubstringFromRange   SEL
	selFirstRectForCharacterRange     SEL
	selCharacterIndexForPoint         SEL
	selValidAttributesForMarkedText   SEL
)

// NSMenu/NSMenuItem selectors
var (
	selAllocMenu        SEL
	selInitWithTitle    SEL
	selAddItem          SEL
	selInsertItem       SEL
	selPopUpMenuPositioningItem SEL
	selMenuItemWithTitle SEL
	selSeparatorItem    SEL
	selSetKeyEquivalent SEL
	selSetKeyEquivalentModifierMask SEL
	selSetTarget        SEL
	selSetAction        SEL
	selSetEnabled       SEL
	selTitle            SEL
	selTag              SEL
	selSetTag           SEL
)

// NSCursor selectors
var (
	selArrowCursor      SEL
	selIBeamCursor      SEL
	selCrosshairCursor  SEL
	selPointingHandCursor SEL
	selResizeLeftRightCursor SEL
	selResizeUpDownCursor SEL
	selPop              SEL
	selPush             SEL
	selSet              SEL
	selHide             SEL
	selUnhide           SEL
	selSetHiddenUntilMouseMoves SEL
)

// NSPasteboard selectors (clipboard)
var (
	selGeneralPasteboard SEL
	selStringForType     SEL
	selSetString         SEL
	selClearContents     SEL
)

// NSScreen selectors
var (
	selMainScreen        SEL
	selScreens           SEL
	selBackingScaleFactor SEL
	selVisibleFrame      SEL
)

// NSLocale selectors
var (
	selCurrentLocale     SEL
	selLocaleIdentifier  SEL
)

// NSTrackingArea selectors
var (
	selInitWithRect      SEL
	selAddTrackingArea   SEL
	selRemoveTrackingArea SEL
)

// Core Foundation functions
var (
	cfRelease           uintptr
	cfStringCreateWithCString uintptr
	cfStringGetCString  uintptr
	cfStringGetLength   uintptr
	cfDataCreate        uintptr
	cfDataGetBytePtr    uintptr
	cfDataGetLength     uintptr
)

// NSApplication activation policies
const (
	NSApplicationActivationPolicyRegular   = 0
	NSApplicationActivationPolicyAccessory = 1
	NSApplicationActivationPolicyProhibited = 2
)

// NSWindow style masks
const (
	NSWindowStyleMaskBorderless             = 0
	NSWindowStyleMaskTitled                 = 1 << 0
	NSWindowStyleMaskClosable               = 1 << 1
	NSWindowStyleMaskMiniaturizable         = 1 << 2
	NSWindowStyleMaskResizable              = 1 << 3
	NSWindowStyleMaskTexturedBackground     = 1 << 8
	NSWindowStyleMaskUnifiedTitleAndToolbar = 1 << 12
	NSWindowStyleMaskFullScreen             = 1 << 14
	NSWindowStyleMaskFullSizeContentView    = 1 << 15
	NSWindowStyleMaskUtilityWindow          = 1 << 4
	NSWindowStyleMaskDocModalWindow         = 1 << 6
	NSWindowStyleMaskNonactivatingPanel     = 1 << 7
	NSWindowStyleMaskHUDWindow              = 1 << 13
)

// NSWindow backing store types
const (
	NSBackingStoreRetained    = 0
	NSBackingStoreNonretained = 1
	NSBackingStoreBuffered    = 2
)

// NSEvent types
const (
	NSEventTypeLeftMouseDown         = 1
	NSEventTypeLeftMouseUp           = 2
	NSEventTypeRightMouseDown        = 3
	NSEventTypeRightMouseUp          = 4
	NSEventTypeMouseMoved            = 5
	NSEventTypeLeftMouseDragged      = 6
	NSEventTypeRightMouseDragged     = 7
	NSEventTypeMouseEntered          = 8
	NSEventTypeMouseExited           = 9
	NSEventTypeKeyDown               = 10
	NSEventTypeKeyUp                 = 11
	NSEventTypeFlagsChanged          = 12
	NSEventTypeAppKitDefined         = 13
	NSEventTypeSystemDefined         = 14
	NSEventTypeApplicationDefined    = 15
	NSEventTypePeriodic              = 16
	NSEventTypeCursorUpdate          = 17
	NSEventTypeScrollWheel           = 22
	NSEventTypeTabletPoint           = 23
	NSEventTypeTabletProximity       = 24
	NSEventTypeOtherMouseDown        = 25
	NSEventTypeOtherMouseUp          = 26
	NSEventTypeOtherMouseDragged     = 27
	NSEventTypeGesture               = 29
	NSEventTypeMagnify               = 30
	NSEventTypeSwipe                 = 31
	NSEventTypeRotate                = 18
	NSEventTypeBeginGesture          = 19
	NSEventTypeEndGesture            = 20
	NSEventTypeSmartMagnify          = 32
	NSEventTypeQuickLook             = 33
	NSEventTypePressure              = 34
	NSEventTypeDirectTouch           = 35
)

// NSEvent modifier flags
const (
	NSEventModifierFlagCapsLock   = 1 << 16
	NSEventModifierFlagShift      = 1 << 17
	NSEventModifierFlagControl    = 1 << 18
	NSEventModifierFlagOption     = 1 << 19
	NSEventModifierFlagCommand    = 1 << 20
	NSEventModifierFlagNumericPad = 1 << 21
	NSEventModifierFlagHelp       = 1 << 22
	NSEventModifierFlagFunction   = 1 << 23
)

// NSTrackingArea options
const (
	NSTrackingMouseEnteredAndExited = 1 << 0
	NSTrackingMouseMoved            = 1 << 1
	NSTrackingCursorUpdate          = 1 << 2
	NSTrackingActiveWhenFirstResponder = 1 << 4
	NSTrackingActiveInKeyWindow     = 1 << 5
	NSTrackingActiveInActiveApp     = 1 << 6
	NSTrackingActiveAlways          = 1 << 7
	NSTrackingAssumeInside          = 1 << 8
	NSTrackingInVisibleRect         = 1 << 9
	NSTrackingEnabledDuringMouseDrag = 1 << 10
)

// NSPasteboard types
const (
	NSPasteboardTypeString = "public.utf8-plain-text"
)

// NSRange constants
const (
	NSNotFound = 0x7FFFFFFFFFFFFFFF // NSInteger max value
)

// NSRect represents a rectangle in Cocoa
type NSRect struct {
	Origin NSPoint
	Size   NSSize
}

// NSPoint represents a point in Cocoa
type NSPoint struct {
	X, Y float64
}

// NSSize represents a size in Cocoa
type NSSize struct {
	Width, Height float64
}

// NSRange represents a range in Cocoa
type NSRange struct {
	Location uint64
	Length   uint64
}

// id represents an Objective-C object pointer
type id uintptr

// SEL represents an Objective-C selector
type SEL uintptr

// Class represents an Objective-C class
type Class uintptr

// IMP represents an Objective-C method implementation
type IMP uintptr

var (
	// Cached class references
	classNSApplication   Class
	classNSWindow        Class
	classNSView          Class
	classNSString        Class
	classNSEvent         Class
	classNSMenu          Class
	classNSMenuItem      Class
	classNSCursor        Class
	classNSPasteboard    Class
	classNSScreen        Class
	classNSLocale        Class
	classNSTrackingArea  Class
	classNSColor         Class
	classCALayer         Class
)

// objcClass gets a class by name, caching the result
func objcClass(name string) Class {
	cstr := cstring(name)
	switch name {
	case "NSApplication":
		if classNSApplication == 0 {
			r1, _, _ := purego.SyscallN(objc_getClass, uintptr(unsafe.Pointer(cstr)))
			classNSApplication = Class(r1)
		}
		return classNSApplication
	case "NSWindow":
		if classNSWindow == 0 {
			r1, _, _ := purego.SyscallN(objc_getClass, uintptr(unsafe.Pointer(cstr)))
			classNSWindow = Class(r1)
		}
		return classNSWindow
	case "NSView":
		if classNSView == 0 {
			r1, _, _ := purego.SyscallN(objc_getClass, uintptr(unsafe.Pointer(cstr)))
			classNSView = Class(r1)
		}
		return classNSView
	case "NSString":
		if classNSString == 0 {
			r1, _, _ := purego.SyscallN(objc_getClass, uintptr(unsafe.Pointer(cstr)))
			classNSString = Class(r1)
		}
		return classNSString
	case "NSEvent":
		if classNSEvent == 0 {
			r1, _, _ := purego.SyscallN(objc_getClass, uintptr(unsafe.Pointer(cstr)))
			classNSEvent = Class(r1)
		}
		return classNSEvent
	case "NSMenu":
		if classNSMenu == 0 {
			r1, _, _ := purego.SyscallN(objc_getClass, uintptr(unsafe.Pointer(cstr)))
			classNSMenu = Class(r1)
		}
		return classNSMenu
	case "NSMenuItem":
		if classNSMenuItem == 0 {
			r1, _, _ := purego.SyscallN(objc_getClass, uintptr(unsafe.Pointer(cstr)))
			classNSMenuItem = Class(r1)
		}
		return classNSMenuItem
	case "NSCursor":
		if classNSCursor == 0 {
			r1, _, _ := purego.SyscallN(objc_getClass, uintptr(unsafe.Pointer(cstr)))
			classNSCursor = Class(r1)
		}
		return classNSCursor
	case "NSPasteboard":
		if classNSPasteboard == 0 {
			r1, _, _ := purego.SyscallN(objc_getClass, uintptr(unsafe.Pointer(cstr)))
			classNSPasteboard = Class(r1)
		}
		return classNSPasteboard
	case "NSScreen":
		if classNSScreen == 0 {
			r1, _, _ := purego.SyscallN(objc_getClass, uintptr(unsafe.Pointer(cstr)))
			classNSScreen = Class(r1)
		}
		return classNSScreen
	case "NSLocale":
		if classNSLocale == 0 {
			r1, _, _ := purego.SyscallN(objc_getClass, uintptr(unsafe.Pointer(cstr)))
			classNSLocale = Class(r1)
		}
		return classNSLocale
	case "NSTrackingArea":
		if classNSTrackingArea == 0 {
			r1, _, _ := purego.SyscallN(objc_getClass, uintptr(unsafe.Pointer(cstr)))
			classNSTrackingArea = Class(r1)
		}
		return classNSTrackingArea
	case "NSColor":
		if classNSColor == 0 {
			r1, _, _ := purego.SyscallN(objc_getClass, uintptr(unsafe.Pointer(cstr)))
			classNSColor = Class(r1)
		}
		return classNSColor
	case "CALayer":
		if classCALayer == 0 {
			r1, _, _ := purego.SyscallN(objc_getClass, uintptr(unsafe.Pointer(cstr)))
			classCALayer = Class(r1)
		}
		return classCALayer
	default:
		r1, _, _ := purego.SyscallN(objc_getClass, uintptr(unsafe.Pointer(cstr)))
		return Class(r1)
	}
}

// selectorCache caches selectors by name to avoid repeated lookups
var selectorCache = make(map[string]SEL)

// objcSelector registers a selector by name, caching the result
func objcSelector(name string) SEL {
	// Check cache first
	if sel, ok := selectorCache[name]; ok {
		return sel
	}
	
	// Register new selector
	r1, _, _ := purego.SyscallN(sel_registerName, uintptr(unsafe.Pointer(cstring(name))))
	sel := SEL(r1)
	
	// Cache it
	selectorCache[name] = sel
	return sel
}

// msgSend sends a message to an object
func msgSend(obj id, sel SEL, args ...uintptr) id {
	var r1 uintptr
	switch len(args) {
	case 0:
		r1, _, _ = purego.SyscallN(objc_msgSend, uintptr(obj), uintptr(sel))
	case 1:
		r1, _, _ = purego.SyscallN(objc_msgSend, uintptr(obj), uintptr(sel), args[0])
	case 2:
		r1, _, _ = purego.SyscallN(objc_msgSend, uintptr(obj), uintptr(sel), args[0], args[1])
	case 3:
		r1, _, _ = purego.SyscallN(objc_msgSend, uintptr(obj), uintptr(sel), args[0], args[1], args[2])
	case 4:
		r1, _, _ = purego.SyscallN(objc_msgSend, uintptr(obj), uintptr(sel), args[0], args[1], args[2], args[3])
	case 5:
		r1, _, _ = purego.SyscallN(objc_msgSend, uintptr(obj), uintptr(sel), args[0], args[1], args[2], args[3], args[4])
	default:
		// For more arguments, build the slice
		allArgs := make([]uintptr, 0, 2+len(args))
		allArgs = append(allArgs, uintptr(obj), uintptr(sel))
		allArgs = append(allArgs, args...)
		r1, _, _ = purego.SyscallN(objc_msgSend, allArgs...)
	}
	return id(r1)
}

// msgSendFloat64 sends a message that returns a float64
func msgSendFloat64(obj id, sel SEL, args ...uintptr) float64 {
	var r1 uintptr
	switch len(args) {
	case 0:
		r1, _, _ = purego.SyscallN(objc_msgSend, uintptr(obj), uintptr(sel))
	case 1:
		r1, _, _ = purego.SyscallN(objc_msgSend, uintptr(obj), uintptr(sel), args[0])
	case 2:
		r1, _, _ = purego.SyscallN(objc_msgSend, uintptr(obj), uintptr(sel), args[0], args[1])
	default:
		allArgs := make([]uintptr, 0, 2+len(args))
		allArgs = append(allArgs, uintptr(obj), uintptr(sel))
		allArgs = append(allArgs, args...)
		r1, _, _ = purego.SyscallN(objc_msgSend, allArgs...)
	}
	return *(*float64)(unsafe.Pointer(&r1))
}

// msgSendPtr sends a message and returns a pointer value (used for getting structs like NSRect)
func msgSendPtr(obj id, sel SEL, retPtr unsafe.Pointer, args ...uintptr) {
	switch len(args) {
	case 0:
		purego.SyscallN(objc_msgSend, uintptr(obj), uintptr(sel), uintptr(retPtr))
	case 1:
		purego.SyscallN(objc_msgSend, uintptr(obj), uintptr(sel), uintptr(retPtr), args[0])
	case 2:
		purego.SyscallN(objc_msgSend, uintptr(obj), uintptr(sel), uintptr(retPtr), args[0], args[1])
	default:
		allArgs := make([]uintptr, 0, 3+len(args))
		allArgs = append(allArgs, uintptr(obj), uintptr(sel), uintptr(retPtr))
		allArgs = append(allArgs, args...)
		purego.SyscallN(objc_msgSend, allArgs...)
	}
}

// msgSend_stret sends a message that returns a structure by value
// On x86_64, structs > 16 bytes use stret, but modern macOS uses regular msgSend for most cases
// We use this as an alias for consistency with other platforms
func msgSend_stret(retPtr unsafe.Pointer, obj id, sel SEL, args ...uintptr) {
	allArgs := make([]uintptr, 0, 2+len(args))
	allArgs = append(allArgs, uintptr(obj), uintptr(sel))
	allArgs = append(allArgs, args...)
	// On arm64 and modern x86_64, objc_msgSend handles struct returns
	purego.SyscallN(objc_msgSend, allArgs...)
	// Copy result - this is simplified; real implementation needs runtime type info
	// For our use case with NSRect, we rely on the calling convention
}

// cstring converts a Go string to a null-terminated C string
func cstring(s string) *byte {
	b := make([]byte, len(s)+1)
	copy(b, s)
	return &b[0]
}

// nsString creates an NSString from a Go string
func nsString(s string) id {
	cls := objcClass("NSString")
	if cls == 0 {
		return 0
	}
	if s == "" {
		return msgSend(msgSend(id(cls), selAlloc), selInit)
	}
	cstr := cstring(s)
	return msgSend(id(cls), selStringWithUTF8String, uintptr(unsafe.Pointer(cstr)))
}

// goString converts an NSString to a Go string
func goString(nsstr id) string {
	if nsstr == 0 {
		return ""
	}
	cstr := msgSend(nsstr, selUTF8String)
	if cstr == 0 {
		return ""
	}
	// Convert C string to Go string
	ptr := (*byte)(unsafe.Pointer(cstr))
	var result []byte
	for *ptr != 0 {
		result = append(result, *ptr)
		ptr = (*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + 1))
	}
	return string(result)
}

// nsRect creates an NSRect from components
func nsRect(x, y, width, height float64) NSRect {
	return NSRect{
		Origin: NSPoint{X: x, Y: y},
		Size:   NSSize{Width: width, Height: height},
	}
}

// nsPoint creates an NSPoint
func nsPoint(x, y float64) NSPoint {
	return NSPoint{X: x, Y: y}
}

// nsSize creates an NSSize
func nsSize(width, height float64) NSSize {
	return NSSize{Width: width, Height: height}
}

// uintptrRect converts an NSRect to uintptr arguments for passing to objc_msgSend
func uintptrRect(r NSRect) (uintptr, uintptr, uintptr, uintptr) {
	return uintptr(r.Origin.X), uintptr(r.Origin.Y), uintptr(r.Size.Width), uintptr(r.Size.Height)
}

// rectFromPtr reads an NSRect from a pointer
func rectFromPtr(ptr unsafe.Pointer) NSRect {
	return *(*NSRect)(ptr)
}

// pointFromPtr reads an NSPoint from a pointer
func pointFromPtr(ptr unsafe.Pointer) NSPoint {
	return *(*NSPoint)(ptr)
}

// sizeFromPtr reads an NSSize from a pointer
func sizeFromPtr(ptr unsafe.Pointer) NSSize {
	return *(*NSSize)(ptr)
}

var foundationLoaded bool

// ensureFoundation loads Foundation framework if not already loaded.
// This is called lazily when creating windows, not in init().
func ensureFoundation() {
	if foundationLoaded {
		return
	}
	_, err := purego.Dlopen("/System/Library/Frameworks/Foundation.framework/Foundation", purego.RTLD_LAZY|purego.RTLD_GLOBAL)
	if err != nil {
		// Foundation framework load failed, but continue
		// Classes will return nil, which callers should handle
	}
	foundationLoaded = true
}

func init() {
	// Load the Objective-C runtime library
	objc, err := purego.Dlopen("/usr/lib/libobjc.A.dylib", purego.RTLD_LAZY)
	if err != nil {
		panic("failed to load libobjc.A.dylib: " + err.Error())
	}

	// Get the required functions from the Objective-C runtime
	objc_msgSend, err = purego.Dlsym(objc, "objc_msgSend")
	if err != nil {
		panic("failed to load objc_msgSend: " + err.Error())
	}
	objc_getClass, err = purego.Dlsym(objc, "objc_getClass")
	if err != nil {
		panic("failed to load objc_getClass: " + err.Error())
	}
	sel_registerName, err = purego.Dlsym(objc, "sel_registerName")
	if err != nil {
		panic("failed to load sel_registerName: " + err.Error())
	}
	class_getName, err = purego.Dlsym(objc, "class_getName")
	if err != nil {
		panic("failed to load class_getName: " + err.Error())
	}
	object_getClass, err = purego.Dlsym(objc, "object_getClass")
	if err != nil {
		panic("failed to load object_getClass: " + err.Error())
	}

	// Note: Class references are not initialized here.
	// They will be loaded lazily when objcClass() is first called.
	// This allows tests to run without loading Foundation framework.

	// Initialize all selectors
	// NSObject
	selAlloc = objcSelector("alloc")
	selInit = objcSelector("init")
	selRelease = objcSelector("release")
	selAutorelease = objcSelector("autorelease")
	selRetain = objcSelector("retain")
	selClass = objcSelector("class")
	selDescription = objcSelector("description")
	selUTF8String = objcSelector("UTF8String")
	
	// NSString
	selStringWithUTF8String = objcSelector("stringWithUTF8String:")
	selLength = objcSelector("length")
	selCharacterAtIndex = objcSelector("characterAtIndex:")
	
	// NSApplication
	selSharedApplication = objcSelector("sharedApplication")
	selRun = objcSelector("run")
	selTerminate = objcSelector("terminate:")
	selSetActivationPolicy = objcSelector("setActivationPolicy:")
	selActivateIgnoringOtherApps = objcSelector("activateIgnoringOtherApps:")
	
	// NSWindow
	selInitWithContentRect = objcSelector("initWithContentRect:styleMask:backing:defer:")
	selMakeKeyAndOrderFront = objcSelector("makeKeyAndOrderFront:")
	selClose = objcSelector("close")
	selSetTitle = objcSelector("setTitle:")
	selContentView = objcSelector("contentView")
	selFrame = objcSelector("frame")
	selSetFrame = objcSelector("setFrame:display:")
	selSetFrameTopLeftPoint = objcSelector("setFrameTopLeftPoint:")
	selContentRectForFrameRect = objcSelector("contentRectForFrameRect:")
	selFrameRectForContentRect = objcSelector("frameRectForContentRect:")
	selSetMinSize = objcSelector("setMinSize:")
	selSetMaxSize = objcSelector("setMaxSize:")
	selToggleFullScreen = objcSelector("toggleFullScreen:")
	selIsZoomed = objcSelector("isZoomed")
	selSetStyleMask = objcSelector("setStyleMask:")
	selStyleMask = objcSelector("styleMask")
	selSetOpaque = objcSelector("setOpaque:")
	selSetBackgroundColor = objcSelector("setBackgroundColor:")
	selMakeFirstResponder = objcSelector("makeFirstResponder:")
	selConvertRectToScreen = objcSelector("convertRectToScreen:")
	selScreen = objcSelector("screen")
	selWindow = objcSelector("window")
	
	// NSView
	selSetWantsLayer = objcSelector("setWantsLayer:")
	selLayer = objcSelector("layer")
	selSetLayer = objcSelector("setLayer:")
	selBounds = objcSelector("bounds")
	selSetBounds = objcSelector("setBounds:")
	selConvertRectFromView = objcSelector("convertRect:fromView:")
	selConvertRectToView = objcSelector("convertRect:toView:")
	
	// NSResponder
	selAcceptsFirstResponder = objcSelector("acceptsFirstResponder")
	selBecomeFirstResponder = objcSelector("becomeFirstResponder")
	selResignFirstResponder = objcSelector("resignFirstResponder")
	
	// NSEvent
	selType = objcSelector("type")
	selLocationInWindow = objcSelector("locationInWindow")
	selModifierFlags = objcSelector("modifierFlags")
	selButtonNumber = objcSelector("buttonNumber")
	selClickCount = objcSelector("clickCount")
	selKeyCode = objcSelector("keyCode")
	selCharacters = objcSelector("characters")
	selDeltaX = objcSelector("deltaX")
	selDeltaY = objcSelector("deltaY")
	selTimestamp = objcSelector("timestamp")
	
	// NSTextInputClient
	selInsertText = objcSelector("insertText:replacementRange:")
	selSetMarkedText = objcSelector("setMarkedText:selectedRange:replacementRange:")
	selUnmarkText = objcSelector("unmarkText")
	selSelectedRange = objcSelector("selectedRange")
	selMarkedRange = objcSelector("markedRange")
	selHasMarkedText = objcSelector("hasMarkedText")
	selAttributedSubstringFromRange = objcSelector("attributedSubstringFromRange:")
	selFirstRectForCharacterRange = objcSelector("firstRectForCharacterRange:actualRange:")
	selCharacterIndexForPoint = objcSelector("characterIndexForPoint:")
	selValidAttributesForMarkedText = objcSelector("validAttributesForMarkedText")
	
	// NSMenu/NSMenuItem
	selAllocMenu = objcSelector("alloc")
	selInitWithTitle = objcSelector("initWithTitle:")
	selAddItem = objcSelector("addItem:")
	selInsertItem = objcSelector("insertItem:atIndex:")
	selPopUpMenuPositioningItem = objcSelector("popUpMenuPositioningItem:atLocation:inView:callbackNumber:callback:selector:")
	selMenuItemWithTitle = objcSelector("itemWithTitle:")
	selSeparatorItem = objcSelector("separatorItem")
	selSetKeyEquivalent = objcSelector("setKeyEquivalent:")
	selSetKeyEquivalentModifierMask = objcSelector("setKeyEquivalentModifierMask:")
	selSetTarget = objcSelector("setTarget:")
	selSetAction = objcSelector("setAction:")
	selSetEnabled = objcSelector("setEnabled:")
	selTitle = objcSelector("title")
	selTag = objcSelector("tag")
	selSetTag = objcSelector("setTag:")
	
	// NSCursor
	selArrowCursor = objcSelector("arrowCursor")
	selIBeamCursor = objcSelector("IBeamCursor")
	selCrosshairCursor = objcSelector("crosshairCursor")
	selPointingHandCursor = objcSelector("pointingHandCursor")
	selResizeLeftRightCursor = objcSelector("resizeLeftRightCursor")
	selResizeUpDownCursor = objcSelector("resizeUpDownCursor")
	selPop = objcSelector("pop")
	selPush = objcSelector("push")
	selSet = objcSelector("set")
	selHide = objcSelector("hide")
	selUnhide = objcSelector("unhide")
	selSetHiddenUntilMouseMoves = objcSelector("setHiddenUntilMouseMoves:")
	
	// NSPasteboard
	selGeneralPasteboard = objcSelector("generalPasteboard")
	selStringForType = objcSelector("stringForType:")
	selSetString = objcSelector("setString:forType:")
	selClearContents = objcSelector("clearContents")
	
	// NSScreen
	selMainScreen = objcSelector("mainScreen")
	selScreens = objcSelector("screens")
	selBackingScaleFactor = objcSelector("backingScaleFactor")
	selVisibleFrame = objcSelector("visibleFrame")
	
	// NSLocale
	selCurrentLocale = objcSelector("currentLocale")
	selLocaleIdentifier = objcSelector("localeIdentifier")
	
	// NSTrackingArea
	selInitWithRect = objcSelector("initWithRect:options:owner:userInfo:")
	selAddTrackingArea = objcSelector("addTrackingArea:")
	selRemoveTrackingArea = objcSelector("removeTrackingArea:")
}

// RunLoop runs the NSApplication main run loop (called from platform.go)
func RunLoop() {
	app := msgSend(id(classNSApplication), selSharedApplication)
	msgSend(app, selRun)
}

// StopApplication stops the NSApplication run loop
func StopApplication() {
	app := msgSend(id(classNSApplication), selSharedApplication)
	msgSend(app, selTerminate, 0)
}

// ActivateApplication activates the application and brings it to front
func ActivateApplication() {
	app := msgSend(id(classNSApplication), selSharedApplication)
	// Set activation policy to regular (dock icon, menu bar)
	msgSend(app, selSetActivationPolicy, NSApplicationActivationPolicyRegular)
	// Activate the app
	msgSend(app, selActivateIgnoringOtherApps, 1)
}

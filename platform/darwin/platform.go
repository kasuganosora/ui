//go:build darwin

package darwin

import (
	"fmt"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/platform"
)

// Platform implements platform.Platform using Cocoa APIs.
type Platform struct {
	app      id
	windows  []*Window
	events   []event.Event
	mu       sync.Mutex
	inited   bool
	running  bool
	
	// Clipboard cache
	pasteboard id

	// Autorelease pool for command-line app lifecycle
	autoreleasePool id
	
	// Cached NSDefaultRunLoopMode constant
	defaultRunLoopMode id

	// Event timestamp base
	timebase uint64
}

// New creates a new macOS platform instance.
func New() *Platform {
	return &Platform{}
}

// Init implements platform.Platform.
func (p *Platform) Init() error {
	if p.inited {
		return nil
	}

	// Lock to OS thread since Cocoa requires main thread for most operations
	runtime.LockOSThread()

	// Ensure Foundation framework is loaded (needed for Cocoa classes)
	ensureFoundation()

	// Create an autorelease pool for this thread (CLI app style Cocoa bootstrap).
	if poolClass := objcClass("NSAutoreleasePool"); poolClass != 0 {
		p.autoreleasePool = msgSend(msgSend(id(poolClass), selAlloc), selInit)
	}

	// Get shared NSApplication instance
	p.app = msgSend(id(objcClass("NSApplication")), selSharedApplication)
	if p.app == 0 {
		return fmt.Errorf("darwin: failed to get NSApplication shared instance")
	}

	// Set activation policy to regular (shows dock icon and menu bar).
	// MUST be done before finishLaunching for non-bundle CLI apps.
	msgSend(p.app, selSetActivationPolicy, NSApplicationActivationPolicyRegular)

	// finishLaunching performs the Cocoa bootstrap (menu bar, dock icon, etc.)
	msgSend(p.app, objcSelector("finishLaunching"))

	// Activate the application so it becomes the frontmost app.
	// For non-bundle CLI processes, this is essential — without it,
	// the app stays in the background and windows are invisible.
	msgSend(p.app, selActivateIgnoringOtherApps, 1)

	// Get the general pasteboard for clipboard operations
	p.pasteboard = msgSend(id(objcClass("NSPasteboard")), selGeneralPasteboard)

	// Load NSDefaultRunLoopMode global constant from Foundation.
	// It is an exported NSString* variable — Dlsym returns a pointer to the variable,
	// and we dereference to get the actual NSString* value.
	if foundationHandle, err := purego.Dlopen("/System/Library/Frameworks/Foundation.framework/Foundation", purego.RTLD_LAZY|purego.RTLD_GLOBAL); err == nil {
		if sym, err := purego.Dlsym(foundationHandle, "NSDefaultRunLoopMode"); err == nil && sym != 0 {
			p.defaultRunLoopMode = *(*id)(unsafe.Pointer(sym))
		}
	}
	// Fallback: if we couldn't load the constant, create an NSString manually.
	// The value of NSDefaultRunLoopMode is the string "kCFRunLoopDefaultMode".
	if p.defaultRunLoopMode == 0 {
		p.defaultRunLoopMode = nsString("kCFRunLoopDefaultMode")
	}

	// Initialize timestamp base
	p.timebase = uint64(time.Now().UnixMicro())

	p.inited = true
	return nil
}

// CreateWindow implements platform.Platform.
func (p *Platform) CreateWindow(opts platform.WindowOptions) (platform.Window, error) {
	if !p.inited {
		return nil, fmt.Errorf("darwin: platform not initialized")
	}

	w, err := newWindow(p, opts)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	p.windows = append(p.windows, w)
	p.mu.Unlock()

	return w, nil
}

// PollEvents implements platform.Platform.
// On macOS, this drains events from our internal queue and processes pending Cocoa events.
func (p *Platform) PollEvents() []event.Event {
	// Process pending Cocoa events without blocking
	p.processPendingEvents()

	// Show any deferred windows
	for _, w := range p.windows {
		w.ShowDeferred()
	}

	// Drain collected events
	p.mu.Lock()
	events := p.events
	p.events = p.events[:0]
	p.mu.Unlock()

	return events
}

// ProcessMessages pumps the Cocoa event loop to keep the window responsive.
// Unlike PollEvents, it does not collect or return events.
func (p *Platform) ProcessMessages() {
	p.processPendingEvents()
}

// processPendingEvents processes all pending Cocoa events without blocking.
func (p *Platform) processPendingEvents() {
	// [NSDate distantPast] — a date far in the past, meaning "don't wait".
	distantPast := msgSend(id(objcClass("NSDate")), objcSelector("distantPast"))
	if distantPast == 0 {
		return
	}

	runLoopMode := p.defaultRunLoopMode
	if runLoopMode == 0 {
		return
	}

	for {
		// nextEventMatchingMask:untilDate:inMode:dequeue:
		event := msgSend(p.app, objcSelector("nextEventMatchingMask:untilDate:inMode:dequeue:"),
			0xFFFFFFFF, // NSEventMaskAny
			uintptr(distantPast),
			uintptr(runLoopMode),
			1, // dequeue
		)
		
		if event == 0 {
			break
		}

		// Send the event to be processed
		msgSend(p.app, objcSelector("sendEvent:"), uintptr(event))
		
		// Convert and store the event
		p.convertAndStoreEvent(event)
	}

	// Update windows
	msgSend(p.app, objcSelector("updateWindows"))
}

// convertAndStoreEvent converts a Cocoa NSEvent to our event.Event type.
func (p *Platform) convertAndStoreEvent(nsevent id) {
	if nsevent == 0 {
		return
	}

	// Get the window for this event
	nswindow := msgSend(nsevent, selWindow)
	
	// Find our Window wrapper
	var w *Window
	if nswindow != 0 {
		for _, win := range p.windows {
			if win.nswindow == nswindow {
				w = win
				break
			}
		}
	}

	// Get event type
	eventType := uint64(msgSend(nsevent, selType))
	
	// Get timestamp
	timestamp := p.currentTimestamp()

	// Get location in window
	// locationInWindow returns NSPoint (2× float64) in floating-point registers
	// (xmm0+xmm1 on amd64, d0+d1 on arm64). Must use typed wrapper, not SyscallN.
	var location NSPoint
	if w != nil {
		location = msgSendPointReturn(nsevent, selLocationInWindow)
		// Flip Y coordinate (Cocoa has Y=0 at bottom, we want Y=0 at top)
		location.Y = float64(w.height) - location.Y
	}

	// Get modifier flags
	modifierFlags := uint64(msgSend(nsevent, selModifierFlags))
	modifiers := convertModifiers(modifierFlags)

	switch eventType {
	case NSEventTypeLeftMouseDown:
		p.pushEvent(event.Event{
			Type:      event.MouseDown,
			Timestamp: timestamp,
			Button:    event.MouseButtonLeft,
			X:         float32(location.X),
			Y:         float32(location.Y),
			GlobalX:   float32(location.X),
			GlobalY:   float32(location.Y),
			Modifiers: modifiers,
		})
		if w != nil {
			w.mouseDown = true
		}

	case NSEventTypeLeftMouseUp:
		p.pushEvent(event.Event{
			Type:      event.MouseUp,
			Timestamp: timestamp,
			Button:    event.MouseButtonLeft,
			X:         float32(location.X),
			Y:         float32(location.Y),
			GlobalX:   float32(location.X),
			GlobalY:   float32(location.Y),
			Modifiers: modifiers,
		})
		if w != nil {
			w.mouseDown = false
		}

	case NSEventTypeRightMouseDown:
		p.pushEvent(event.Event{
			Type:      event.MouseDown,
			Timestamp: timestamp,
			Button:    event.MouseButtonRight,
			X:         float32(location.X),
			Y:         float32(location.Y),
			GlobalX:   float32(location.X),
			GlobalY:   float32(location.Y),
			Modifiers: modifiers,
		})

	case NSEventTypeRightMouseUp:
		p.pushEvent(event.Event{
			Type:      event.MouseUp,
			Timestamp: timestamp,
			Button:    event.MouseButtonRight,
			X:         float32(location.X),
			Y:         float32(location.Y),
			GlobalX:   float32(location.X),
			GlobalY:   float32(location.Y),
			Modifiers: modifiers,
		})

	case NSEventTypeOtherMouseDown:
		buttonNumber := uint64(msgSend(nsevent, selButtonNumber))
		btn := event.MouseButtonMiddle
		if buttonNumber == 2 {
			btn = event.MouseButtonMiddle
		} else if buttonNumber == 3 {
			btn = event.MouseButton4
		} else if buttonNumber == 4 {
			btn = event.MouseButton5
		}
		p.pushEvent(event.Event{
			Type:      event.MouseDown,
			Timestamp: timestamp,
			Button:    btn,
			X:         float32(location.X),
			Y:         float32(location.Y),
			GlobalX:   float32(location.X),
			GlobalY:   float32(location.Y),
			Modifiers: modifiers,
		})

	case NSEventTypeOtherMouseUp:
		buttonNumber := uint64(msgSend(nsevent, selButtonNumber))
		btn := event.MouseButtonMiddle
		if buttonNumber == 2 {
			btn = event.MouseButtonMiddle
		} else if buttonNumber == 3 {
			btn = event.MouseButton4
		} else if buttonNumber == 4 {
			btn = event.MouseButton5
		}
		p.pushEvent(event.Event{
			Type:      event.MouseUp,
			Timestamp: timestamp,
			Button:    btn,
			X:         float32(location.X),
			Y:         float32(location.Y),
			GlobalX:   float32(location.X),
			GlobalY:   float32(location.Y),
			Modifiers: modifiers,
		})

	case NSEventTypeMouseMoved, NSEventTypeLeftMouseDragged, NSEventTypeRightMouseDragged, NSEventTypeOtherMouseDragged:
		p.pushEvent(event.Event{
			Type:      event.MouseMove,
			Timestamp: timestamp,
			X:         float32(location.X),
			Y:         float32(location.Y),
			GlobalX:   float32(location.X),
			GlobalY:   float32(location.Y),
			Modifiers: modifiers,
		})

	case NSEventTypeScrollWheel:
		deltaX := msgSendFloat64Return(nsevent, selDeltaX)
		deltaY := msgSendFloat64Return(nsevent, selDeltaY)
		p.pushEvent(event.Event{
			Type:      event.MouseWheel,
			Timestamp: timestamp,
			WheelDX:   float32(deltaX),
			WheelDY:   float32(deltaY),
			Modifiers: modifiers,
		})

	case NSEventTypeMouseEntered:
		p.pushEvent(event.Event{
			Type:      event.MouseEnter,
			Timestamp: timestamp,
			X:         float32(location.X),
			Y:         float32(location.Y),
			Modifiers: modifiers,
		})

	case NSEventTypeMouseExited:
		p.pushEvent(event.Event{
			Type:      event.MouseLeave,
			Timestamp: timestamp,
			Modifiers: modifiers,
		})

	case NSEventTypeKeyDown:
		keyCode := uint16(msgSend(nsevent, selKeyCode))
		key := translateKeyCode(keyCode)
		p.pushEvent(event.Event{
			Type:      event.KeyDown,
			Timestamp: timestamp,
			Key:       key,
			Modifiers: modifiers,
		})

		// Also generate KeyPress for character input
		chars := msgSend(nsevent, selCharacters)
		if chars != 0 {
			str := goString(chars)
			for _, ch := range str {
				if ch >= 32 && ch != 127 {
					p.pushEvent(event.Event{
						Type:      event.KeyPress,
						Timestamp: timestamp,
						Char:      ch,
						Modifiers: modifiers,
					})
				}
			}
		}

	case NSEventTypeKeyUp:
		keyCode := uint16(msgSend(nsevent, selKeyCode))
		key := translateKeyCode(keyCode)
		p.pushEvent(event.Event{
			Type:      event.KeyUp,
			Timestamp: timestamp,
			Key:       key,
			Modifiers: modifiers,
		})

	case NSEventTypeFlagsChanged:
		// Modifier keys changed - we track this for our internal state
		// but don't generate separate events
		if w != nil {
			w.lastModifiers = modifiers
		}
	}
}

// pushEvent adds an event to the queue.
func (p *Platform) pushEvent(e event.Event) {
	p.mu.Lock()
	p.events = append(p.events, e)
	p.mu.Unlock()
}

// currentTimestamp returns a monotonic timestamp in microseconds.
func (p *Platform) currentTimestamp() uint64 {
	return uint64(time.Now().UnixMicro()) - p.timebase
}

// GetClipboardText implements platform.Platform.
func (p *Platform) GetClipboardText() string {
	if p.pasteboard == 0 {
		return ""
	}
	
	// Create NSString for the pasteboard type
	pasteboardType := nsString(NSPasteboardTypeString)
	defer msgSend(pasteboardType, selRelease)
	
	// Get string for type
	result := msgSend(p.pasteboard, selStringForType, uintptr(pasteboardType))
	if result == 0 {
		return ""
	}
	
	return goString(result)
}

// SetClipboardText implements platform.Platform.
func (p *Platform) SetClipboardText(text string) {
	if p.pasteboard == 0 {
		return
	}
	
	// Clear existing contents
	msgSend(p.pasteboard, selClearContents)
	
	// Create NSString from text
	nsText := nsString(text)
	defer msgSend(nsText, selRelease)
	
	// Create NSString for the pasteboard type
	pasteboardType := nsString(NSPasteboardTypeString)
	defer msgSend(pasteboardType, selRelease)
	
	// Set string for type
	msgSend(p.pasteboard, selSetString, uintptr(nsText), uintptr(pasteboardType))
}

// GetPrimaryMonitorDPI implements platform.Platform.
func (p *Platform) GetPrimaryMonitorDPI() float32 {
	// Get the main screen
	screen := msgSend(id(classNSScreen), selMainScreen)
	if screen == 0 {
		return 96.0 // Default fallback
	}
	
	// Get backing scale factor (1.0 = 72 DPI, 2.0 = 144 DPI for Retina)
	scale := msgSendFloat64Return(screen, selBackingScaleFactor)
	
	// Convert to DPI (macOS uses 72 as base DPI)
	return float32(scale * 72.0)
}

// GetSystemLocale implements platform.Platform.
func (p *Platform) GetSystemLocale() string {
	// Get current locale
	locale := msgSend(id(classNSLocale), selCurrentLocale)
	if locale == 0 {
		return "en-US"
	}
	
	// Get locale identifier
	identifier := msgSend(locale, selLocaleIdentifier)
	if identifier == 0 {
		return "en-US"
	}
	
	return goString(identifier)
}

// Terminate implements platform.Platform.
func (p *Platform) Terminate() {
	for _, w := range p.windows {
		w.Destroy()
	}
	p.windows = nil
	if p.autoreleasePool != 0 {
		msgSend(p.autoreleasePool, selRelease)
		p.autoreleasePool = 0
	}
	p.inited = false
	runtime.UnlockOSThread()
}

// convertModifiers converts Cocoa modifier flags to our Modifiers type.
func convertModifiers(flags uint64) event.Modifiers {
	return event.Modifiers{
		Ctrl:  flags&NSEventModifierFlagControl != 0,
		Shift: flags&NSEventModifierFlagShift != 0,
		Alt:   flags&NSEventModifierFlagOption != 0,
		Super: flags&NSEventModifierFlagCommand != 0,
	}
}

// translateKeyCode converts a Cocoa key code to our Key type.
func translateKeyCode(keyCode uint16) event.Key {
	// macOS virtual key codes (based on USB HID usage)
	switch keyCode {
	case 0x00: return event.KeyA
	case 0x01: return event.KeyS
	case 0x02: return event.KeyD
	case 0x03: return event.KeyF
	case 0x04: return event.KeyH
	case 0x05: return event.KeyG
	case 0x06: return event.KeyZ
	case 0x07: return event.KeyX
	case 0x08: return event.KeyC
	case 0x09: return event.KeyV
	case 0x0B: return event.KeyB
	case 0x0C: return event.KeyQ
	case 0x0D: return event.KeyW
	case 0x0E: return event.KeyE
	case 0x0F: return event.KeyR
	case 0x10: return event.KeyY
	case 0x11: return event.KeyT
	case 0x12: return event.Key1
	case 0x13: return event.Key2
	case 0x14: return event.Key3
	case 0x15: return event.Key4
	case 0x16: return event.Key6
	case 0x17: return event.Key5
	case 0x18: return event.KeyEqual
	case 0x19: return event.Key9
	case 0x1A: return event.Key7
	case 0x1B: return event.KeyMinus
	case 0x1C: return event.Key8
	case 0x1D: return event.Key0
	case 0x1E: return event.KeyRightBracket
	case 0x1F: return event.KeyO
	case 0x20: return event.KeyU
	case 0x21: return event.KeyLeftBracket
	case 0x22: return event.KeyI
	case 0x23: return event.KeyP
	case 0x25: return event.KeyL
	case 0x26: return event.KeyJ
	case 0x27: return event.KeyApostrophe
	case 0x28: return event.KeyK
	case 0x29: return event.KeySemicolon
	case 0x2A: return event.KeyBackslash
	case 0x2B: return event.KeyComma
	case 0x2C: return event.KeySlash
	case 0x2D: return event.KeyN
	case 0x2E: return event.KeyM
	case 0x2F: return event.KeyPeriod
	case 0x32: return event.KeyGraveAccent
	case 0x24: return event.KeyEnter
	case 0x30: return event.KeyTab
	case 0x31: return event.KeySpace
	case 0x33: return event.KeyBackspace
	case 0x34: return event.KeyNumpadEnter
	case 0x35: return event.KeyEscape
	case 0x37: return event.KeyLeftSuper // Command
	case 0x38: return event.KeyLeftShift
	case 0x39: return event.KeyCapsLock
	case 0x3A: return event.KeyLeftAlt // Option
	case 0x3B: return event.KeyLeftCtrl
	case 0x3C: return event.KeyRightShift
	case 0x3D: return event.KeyRightAlt
	case 0x3E: return event.KeyRightCtrl
	case 0x3F: return event.KeyNumpadMultiply
	case 0x45: return event.KeyNumpadAdd
	case 0x4B: return event.KeyNumpadDivide
	case 0x4E: return event.KeyNumpadSubtract
	case 0x47: return event.KeyNumLock // Clear key on macOS numpad
	case 0x52: return event.KeyNumpad0
	case 0x53: return event.KeyNumpad1
	case 0x54: return event.KeyNumpad2
	case 0x55: return event.KeyNumpad3
	case 0x56: return event.KeyNumpad4
	case 0x57: return event.KeyNumpad5
	case 0x58: return event.KeyNumpad6
	case 0x59: return event.KeyNumpad7
	case 0x5B: return event.KeyNumpad8
	case 0x5C: return event.KeyNumpad9
	case 0x60: return event.KeyF5
	case 0x61: return event.KeyF6
	case 0x62: return event.KeyF7
	case 0x63: return event.KeyF3
	case 0x64: return event.KeyF8
	case 0x65: return event.KeyF9
	case 0x67: return event.KeyF11
	case 0x6D: return event.KeyF10
	case 0x6E: return event.KeyF12
	case 0x71: return event.KeyPause // Help key on some keyboards
	case 0x72: return event.KeyInsert
	case 0x73: return event.KeyHome
	case 0x74: return event.KeyPageUp
	case 0x75: return event.KeyDelete
	case 0x76: return event.KeyF4
	case 0x77: return event.KeyEnd
	case 0x78: return event.KeyF2
	case 0x79: return event.KeyPageDown
	case 0x7A: return event.KeyF1
	case 0x7B: return event.KeyArrowLeft
	case 0x7C: return event.KeyArrowRight
	case 0x7D: return event.KeyArrowDown
	case 0x7E: return event.KeyArrowUp
	default: return event.KeyUnknown
	}
}

// Compile-time interface check.
var _ platform.Platform = (*Platform)(nil)

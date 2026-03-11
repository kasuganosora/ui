//go:build linux && !android

package linux

import (
	"fmt"
	"os"
	"sync"
	"unsafe"

	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/platform"
)

// Platform implements platform.Platform for Linux/X11.
type Platform struct {
	dpy     Display
	windows []*Window
	events  []event.Event
	mu      sync.Mutex
	inited  bool

	// WM atoms
	wmDeleteWindow    Atom
	wmState           Atom
	wmStateFullscreen Atom
	netWMName         Atom
	utf8String        Atom
	clipboard         Atom
	targets           Atom
	netWMPid          Atom
}

// New creates a new Linux X11 platform instance.
func New() *Platform {
	return &Platform{}
}

// Init implements platform.Platform.
// Opens the X11 display connection and interns WM atoms.
func (p *Platform) Init() error {
	if p.inited {
		return nil
	}

	p.dpy = XOpenDisplay("")
	if p.dpy == 0 {
		return fmt.Errorf("linux/x11: XOpenDisplay failed — is $DISPLAY set?")
	}

	// Intern WM protocol atoms
	p.wmDeleteWindow = XInternAtom(p.dpy, "WM_DELETE_WINDOW", 0)
	p.wmState = XInternAtom(p.dpy, "_NET_WM_STATE", 0)
	p.wmStateFullscreen = XInternAtom(p.dpy, "_NET_WM_STATE_FULLSCREEN", 0)
	p.netWMName = XInternAtom(p.dpy, "_NET_WM_NAME", 0)
	p.utf8String = XInternAtom(p.dpy, "UTF8_STRING", 0)
	p.clipboard = XInternAtom(p.dpy, "CLIPBOARD", 0)
	p.targets = XInternAtom(p.dpy, "TARGETS", 0)
	p.netWMPid = XInternAtom(p.dpy, "_NET_WM_PID", 0)

	p.inited = true
	return nil
}

// CreateWindow implements platform.Platform.
func (p *Platform) CreateWindow(opts platform.WindowOptions) (platform.Window, error) {
	if !p.inited {
		return nil, fmt.Errorf("linux/x11: platform not initialized")
	}
	w, err := newWindow(p, opts)
	if err != nil {
		return nil, err
	}
	p.windows = append(p.windows, w)
	return w, nil
}

// PollEvents implements platform.Platform.
// Drains all pending X11 events, translates them, and returns the collected events.
func (p *Platform) PollEvents() []event.Event {
	// Show deferred windows on first poll
	for _, w := range p.windows {
		w.ShowDeferred()
	}

	// Drain all pending X11 events
	for XPending(p.dpy) > 0 {
		var xe XEvent
		XNextEvent(p.dpy, &xe)

		// Dispatch to the correct window.
		// xe[0] = type, xe[1] = serial, xe[2] = send_event, xe[3] = display*, xe[4] = window
		// (for most events; ConfigureNotify and ClientMessage have window at index 4 too)
		// We use index 4 which is the 'Window' field in the common event header.
		evType := int(xe[0])
		var xwin XWindow

		switch evType {
		case KeyPress, KeyRelease:
			kev := (*XKeyEvent)(unsafe.Pointer(&xe))
			xwin = kev.Window
		case ButtonPress, ButtonRelease:
			bev := (*XButtonEvent)(unsafe.Pointer(&xe))
			xwin = bev.Window
		case MotionNotify:
			mev := (*XMotionEvent)(unsafe.Pointer(&xe))
			xwin = mev.Window
		case ConfigureNotify:
			cev := (*XConfigureEvent)(unsafe.Pointer(&xe))
			xwin = cev.Window
		case ClientMessage:
			cmev := (*XClientMessageEvent)(unsafe.Pointer(&xe))
			xwin = cmev.Window
		case FocusIn, FocusOut:
			fev := (*XFocusChangeEvent)(unsafe.Pointer(&xe))
			xwin = fev.Window
		default:
			// Generic: window is at offset 4 (int64 index 4)
			xwin = XWindow(xe[4])
		}

		win := p.findWindow(xwin)
		if win != nil {
			p.translateEvent(xe, win, &p.events)
		}
	}

	p.mu.Lock()
	evs := make([]event.Event, len(p.events))
	copy(evs, p.events)
	p.events = p.events[:0]
	p.mu.Unlock()

	return evs
}

// ProcessMessages implements platform.Platform.
// Flushes the X11 output buffer to keep the connection responsive.
func (p *Platform) ProcessMessages() {
	XFlush(p.dpy)
}

// GetClipboardText implements platform.Platform.
// Returns the text content of the CLIPBOARD selection.
// Basic stub — full async ICCCM clipboard requires event loop integration.
func (p *Platform) GetClipboardText() string {
	return ""
}

// SetClipboardText implements platform.Platform.
// Stub — full ICCCM clipboard ownership requires handling SelectionRequest events.
func (p *Platform) SetClipboardText(text string) {}

// GetPrimaryMonitorDPI implements platform.Platform.
// Returns 96.0 (standard Linux DPI). Full Xrandr DPI detection is a future enhancement.
func (p *Platform) GetPrimaryMonitorDPI() float32 {
	return 96.0
}

// GetSystemLocale implements platform.Platform.
// Reads the LANG environment variable.
func (p *Platform) GetSystemLocale() string {
	if lang := os.Getenv("LANG"); lang != "" {
		return lang
	}
	if lc := os.Getenv("LC_ALL"); lc != "" {
		return lc
	}
	return "en_US.UTF-8"
}

// Terminate implements platform.Platform.
func (p *Platform) Terminate() {
	for _, w := range p.windows {
		w.Destroy()
	}
	p.windows = nil
	if p.dpy != 0 {
		XCloseDisplay(p.dpy)
		p.dpy = 0
	}
	p.inited = false
}

// pushEvent appends an event to the queue (called from event translation).
func (p *Platform) pushEvent(e event.Event) {
	p.mu.Lock()
	p.events = append(p.events, e)
	p.mu.Unlock()
}

// findWindow finds the Window struct for the given X11 window ID.
func (p *Platform) findWindow(xwin XWindow) *Window {
	for _, w := range p.windows {
		if w.xwin == xwin {
			return w
		}
	}
	return nil
}

// Display returns the X11 Display connection (used by Vulkan surface creation).
func (p *Platform) Display() Display {
	return p.dpy
}

// Compile-time interface check.
var _ platform.Platform = (*Platform)(nil)

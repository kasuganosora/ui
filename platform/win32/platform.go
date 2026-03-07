//go:build windows

package win32

import (
	"fmt"
	"runtime"
	"sync"
	"syscall"
	"unsafe"

	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/platform"
)

const windowClassName = "GoUI_Window"

// Platform implements platform.Platform using Win32 APIs.
type Platform struct {
	hinstance uintptr
	classAtom uint16
	windows   []*Window
	events    []event.Event
	mu        sync.Mutex
	inited    bool
}

// New creates a new Win32 platform instance.
func New() *Platform {
	return &Platform{}
}

// Init implements platform.Platform.
// Must be called from the main goroutine. Locks the goroutine to the OS thread
// because Win32 requires the message loop to run on the window-creating thread.
func (p *Platform) Init() error {
	if p.inited {
		return nil
	}

	// Win32 message loop must run on the thread that created the window.
	// Go's scheduler can move goroutines between threads, so we lock this one.
	runtime.LockOSThread()

	// Get module handle
	h, _, _ := procGetModuleHandleW.Call(0)
	if h == 0 {
		return fmt.Errorf("win32: GetModuleHandle failed")
	}
	p.hinstance = h

	// Enable DPI awareness (best-effort)
	if procSetProcessDpiAwareness.Find() == nil {
		procSetProcessDpiAwareness.Call(2) // PROCESS_PER_MONITOR_DPI_AWARE
	} else if procSetProcessDPIAware.Find() == nil {
		procSetProcessDPIAware.Call()
	}

	// Register window class
	cursor, _, _ := procLoadCursorW.Call(0, uintptr(IDC_ARROW))

	wc := WNDCLASSEXW{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEXW{})),
		Style:         CS_HREDRAW | CS_VREDRAW | CS_OWNDC | CS_DBLCLKS,
		LpfnWndProc:   syscall.NewCallback(wndProc),
		HInstance:     p.hinstance,
		HCursor:       cursor,
		LpszClassName: utf16PtrFromString(windowClassName),
	}

	atom, _, callErr := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if atom == 0 {
		// ERROR_CLASS_ALREADY_EXISTS (1410) is OK — another instance already registered it.
		if errno, ok := callErr.(syscall.Errno); ok && errno == 1410 {
			// Class already registered, continue.
		} else {
			return fmt.Errorf("win32: RegisterClassExW failed: %v", callErr)
		}
	} else {
		p.classAtom = uint16(atom)
	}

	p.inited = true
	return nil
}

// CreateWindow implements platform.Platform.
func (p *Platform) CreateWindow(opts platform.WindowOptions) (platform.Window, error) {
	if !p.inited {
		return nil, fmt.Errorf("win32: platform not initialized")
	}
	w, err := newWindow(p, opts)
	if err != nil {
		return nil, err
	}
	p.windows = append(p.windows, w)
	return w, nil
}

// PollEvents implements platform.Platform.
func (p *Platform) PollEvents() []event.Event {
	var msg MSG
	for {
		ret, _, _ := procPeekMessageW.Call(
			uintptr(unsafe.Pointer(&msg)),
			0, 0, 0,
			PM_REMOVE,
		)
		if ret == 0 {
			break
		}
		if msg.Message == WM_QUIT {
			// Mark all windows as should close
			for _, w := range p.windows {
				w.shouldClose = true
			}
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}

	// Drain collected events
	p.mu.Lock()
	events := p.events
	p.events = p.events[:0]
	p.mu.Unlock()

	return events
}

// pushEvent adds an event to the queue (called from WndProc).
func (p *Platform) pushEvent(e event.Event) {
	p.mu.Lock()
	p.events = append(p.events, e)
	p.mu.Unlock()
}

// GetClipboardText implements platform.Platform.
func (p *Platform) GetClipboardText() string {
	return getClipboardText()
}

// SetClipboardText implements platform.Platform.
func (p *Platform) SetClipboardText(text string) {
	setClipboardText(text)
}

// GetPrimaryMonitorDPI implements platform.Platform.
func (p *Platform) GetPrimaryMonitorDPI() float32 {
	hdc, _, _ := procGetDC.Call(0)
	if hdc == 0 {
		return 96.0
	}
	defer procReleaseDC.Call(0, hdc)
	dpi, _, _ := procGetDeviceCaps.Call(hdc, LOGPIXELSX)
	if dpi == 0 {
		return 96.0
	}
	return float32(dpi)
}

// GetSystemLocale implements platform.Platform.
func (p *Platform) GetSystemLocale() string {
	// Use GetUserDefaultLocaleName (kernel32)
	procGetUserDefaultLocaleName := kernel32.NewProc("GetUserDefaultLocaleName")
	if procGetUserDefaultLocaleName.Find() != nil {
		return "en-US"
	}
	buf := make([]uint16, 85) // LOCALE_NAME_MAX_LENGTH
	procGetUserDefaultLocaleName.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	return syscall.UTF16ToString(buf)
}

// Terminate implements platform.Platform.
func (p *Platform) Terminate() {
	for _, w := range p.windows {
		w.Destroy()
	}
	p.windows = nil
	p.inited = false
	runtime.UnlockOSThread()
}

// windowFromHWND finds the Window for a given HWND via GWLP_USERDATA.
func windowFromHWND(hwnd uintptr) *Window {
	ptr, _, _ := procGetWindowLongPtrW.Call(hwnd, uintptr(uint32ToUintptr(GWLP_USERDATA)))
	if ptr == 0 {
		return nil
	}
	return (*Window)(unsafe.Pointer(ptr))
}

// lastError returns a formatted error with GetLastError.
func lastError(funcName string) error {
	code, _, _ := procGetLastError.Call()
	return fmt.Errorf("win32: %s failed (error %d)", funcName, code)
}

// Compile-time interface check.
var _ platform.Platform = (*Platform)(nil)

//go:build windows

// Package win32 implements the platform.Platform interface using the Win32 API.
// This is a DDD anti-corruption layer: all Win32 specifics (HWND, MSG, WndProc)
// are translated into the platform domain model (Window, Event).
// Zero CGO — all system calls go through syscall.NewLazyDLL.
package win32

import (
	"syscall"
	"unsafe"
)

// DLL references — loaded lazily on first use.
var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	shcore   = syscall.NewLazyDLL("shcore.dll")
	imm32    = syscall.NewLazyDLL("imm32.dll")
	ole32    = syscall.NewLazyDLL("ole32.dll")
)

// user32 functions
var (
	procRegisterClassExW     = user32.NewProc("RegisterClassExW")
	procCreateWindowExW      = user32.NewProc("CreateWindowExW")
	procDestroyWindow        = user32.NewProc("DestroyWindow")
	procShowWindow           = user32.NewProc("ShowWindow")
	procUpdateWindow         = user32.NewProc("UpdateWindow")
	procDefWindowProcW       = user32.NewProc("DefWindowProcW")
	procPeekMessageW         = user32.NewProc("PeekMessageW")
	procTranslateMessage     = user32.NewProc("TranslateMessage")
	procDispatchMessageW     = user32.NewProc("DispatchMessageW")
	procPostQuitMessage      = user32.NewProc("PostQuitMessage")
	procGetClientRect        = user32.NewProc("GetClientRect")
	procSetWindowPos         = user32.NewProc("SetWindowPos")
	procSetWindowTextW       = user32.NewProc("SetWindowTextW")
	procGetWindowLongPtrW    = user32.NewProc("GetWindowLongPtrW")
	procSetWindowLongPtrW    = user32.NewProc("SetWindowLongPtrW")
	procGetWindowRect        = user32.NewProc("GetWindowRect")
	procMoveWindow           = user32.NewProc("MoveWindow")
	procGetDC                = user32.NewProc("GetDC")
	procReleaseDC            = user32.NewProc("ReleaseDC")
	procGetSystemMetrics     = user32.NewProc("GetSystemMetrics")
	procAdjustWindowRectEx   = user32.NewProc("AdjustWindowRectEx")
	procSetCursor            = user32.NewProc("SetCursor")
	procLoadCursorW          = user32.NewProc("LoadCursorW")
	procSetCapture           = user32.NewProc("SetCapture")
	procReleaseCapture       = user32.NewProc("ReleaseCapture")
	procGetCursorPos         = user32.NewProc("GetCursorPos")
	procScreenToClient       = user32.NewProc("ScreenToClient")
	procClientToScreen       = user32.NewProc("ClientToScreen")
	procSetFocus             = user32.NewProc("SetFocus")
	procGetFocus             = user32.NewProc("GetFocus")
	procGetKeyState          = user32.NewProc("GetKeyState")
	procGetAsyncKeyState     = user32.NewProc("GetAsyncKeyState")
	procMonitorFromWindow    = user32.NewProc("MonitorFromWindow")
	procOpenClipboard        = user32.NewProc("OpenClipboard")
	procCloseClipboard       = user32.NewProc("CloseClipboard")
	procEmptyClipboard       = user32.NewProc("EmptyClipboard")
	procGetClipboardData     = user32.NewProc("GetClipboardData")
	procSetClipboardData     = user32.NewProc("SetClipboardData")
	procChangeWindowMessageFilterEx = user32.NewProc("ChangeWindowMessageFilterEx")
	procGetDpiForWindow      = user32.NewProc("GetDpiForWindow")
	procSetProcessDPIAware   = user32.NewProc("SetProcessDPIAware")
	procSetTimer             = user32.NewProc("SetTimer")
	procKillTimer            = user32.NewProc("KillTimer")
	procPostMessageW         = user32.NewProc("PostMessageW")
)

// kernel32 functions
var (
	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
	procGlobalAlloc      = kernel32.NewProc("GlobalAlloc")
	procGlobalFree       = kernel32.NewProc("GlobalFree")
	procGlobalLock       = kernel32.NewProc("GlobalLock")
	procGlobalUnlock     = kernel32.NewProc("GlobalUnlock")
	procGetLastError     = kernel32.NewProc("GetLastError")
	procQueryPerformanceCounter   = kernel32.NewProc("QueryPerformanceCounter")
	procQueryPerformanceFrequency = kernel32.NewProc("QueryPerformanceFrequency")
)

// gdi32 functions
var (
	procGetDeviceCaps = gdi32.NewProc("GetDeviceCaps")
)

// shcore functions (DPI awareness)
var (
	procSetProcessDpiAwareness = shcore.NewProc("SetProcessDpiAwareness")
	procGetDpiForMonitor       = shcore.NewProc("GetDpiForMonitor")
)

// imm32 functions (IME)
var (
	procImmGetContext        = imm32.NewProc("ImmGetContext")
	procImmReleaseContext    = imm32.NewProc("ImmReleaseContext")
	procImmSetCompositionWindow = imm32.NewProc("ImmSetCompositionWindow")
	procImmGetCompositionStringW = imm32.NewProc("ImmGetCompositionStringW")
	procImmSetCandidateWindow = imm32.NewProc("ImmSetCandidateWindow")
)

// ---- Win32 Constants ----

// Window messages
const (
	WM_CREATE          = 0x0001
	WM_DESTROY         = 0x0002
	WM_MOVE            = 0x0003
	WM_SIZE            = 0x0005
	WM_ACTIVATE        = 0x0006
	WM_SETFOCUS        = 0x0007
	WM_KILLFOCUS       = 0x0008
	WM_CLOSE           = 0x0010
	WM_QUIT            = 0x0012
	WM_PAINT           = 0x000F
	WM_ERASEBKGND      = 0x0014
	WM_SHOWWINDOW      = 0x0018
	WM_GETMINMAXINFO   = 0x0024
	WM_SETCURSOR       = 0x0020
	WM_MOUSEMOVE       = 0x0200
	WM_LBUTTONDOWN     = 0x0201
	WM_LBUTTONUP       = 0x0202
	WM_LBUTTONDBLCLK   = 0x0203
	WM_RBUTTONDOWN     = 0x0204
	WM_RBUTTONUP       = 0x0205
	WM_RBUTTONDBLCLK   = 0x0206
	WM_MBUTTONDOWN     = 0x0207
	WM_MBUTTONUP       = 0x0208
	WM_MBUTTONDBLCLK   = 0x0209
	WM_MOUSEWHEEL      = 0x020A
	WM_MOUSEHWHEEL     = 0x020E
	WM_XBUTTONDOWN     = 0x020B
	WM_XBUTTONUP       = 0x020C
	WM_MOUSELEAVE      = 0x02A3
	WM_KEYDOWN         = 0x0100
	WM_KEYUP           = 0x0101
	WM_CHAR            = 0x0102
	WM_SYSCOMMAND      = 0x0112
	WM_SYSKEYDOWN      = 0x0104
	WM_SYSKEYUP        = 0x0105
	WM_TIMER           = 0x0113
	WM_DPICHANGED      = 0x02E0
	WM_ENTERSIZEMOVE   = 0x0231
	WM_EXITSIZEMOVE    = 0x0232

	// IME messages
	WM_IME_STARTCOMPOSITION  = 0x010D
	WM_IME_ENDCOMPOSITION    = 0x010E
	WM_IME_COMPOSITION       = 0x010F
	WM_IME_SETCONTEXT        = 0x0281
	WM_IME_NOTIFY            = 0x0282
	WM_IME_CHAR              = 0x0286
)

// IME composition string flags
const (
	GCS_COMPSTR       = 0x0008
	GCS_COMPREADSTR   = 0x0001
	GCS_RESULTSTR     = 0x0800
	GCS_RESULTREADSTR = 0x0200
	GCS_CURSORPOS     = 0x0080
)

// Window styles
const (
	WS_OVERLAPPED       = 0x00000000
	WS_CAPTION          = 0x00C00000
	WS_SYSMENU          = 0x00080000
	WS_THICKFRAME       = 0x00040000
	WS_MINIMIZEBOX      = 0x00020000
	WS_MAXIMIZEBOX      = 0x00010000
	WS_OVERLAPPEDWINDOW = WS_OVERLAPPED | WS_CAPTION | WS_SYSMENU | WS_THICKFRAME | WS_MINIMIZEBOX | WS_MAXIMIZEBOX
	WS_POPUP            = 0x80000000
	WS_VISIBLE          = 0x10000000
	WS_CLIPCHILDREN     = 0x02000000
	WS_CLIPSIBLINGS     = 0x04000000

	WS_EX_APPWINDOW  = 0x00040000
	WS_EX_WINDOWEDGE = 0x00000100
)

// Show window commands
const (
	SW_HIDE         = 0
	SW_SHOW         = 5
	SW_SHOWDEFAULT  = 10
	SW_SHOWMAXIMIZED = 3
)

// Window position flags
const (
	SWP_NOMOVE     = 0x0002
	SWP_NOSIZE     = 0x0001
	SWP_NOZORDER   = 0x0004
	SWP_FRAMECHANGED = 0x0020
)

// PeekMessage flags
const (
	PM_NOREMOVE = 0x0000
	PM_REMOVE   = 0x0001
)

// Virtual key codes
const (
	VK_BACK      = 0x08
	VK_TAB       = 0x09
	VK_RETURN    = 0x0D
	VK_SHIFT     = 0x10
	VK_CONTROL   = 0x11
	VK_MENU      = 0x12 // Alt
	VK_PAUSE     = 0x13
	VK_CAPITAL   = 0x14 // Caps Lock
	VK_ESCAPE    = 0x1B
	VK_SPACE     = 0x20
	VK_PRIOR     = 0x21 // Page Up
	VK_NEXT      = 0x22 // Page Down
	VK_END       = 0x23
	VK_HOME      = 0x24
	VK_LEFT      = 0x25
	VK_UP        = 0x26
	VK_RIGHT     = 0x27
	VK_DOWN      = 0x28
	VK_SNAPSHOT  = 0x2C // Print Screen
	VK_INSERT    = 0x2D
	VK_DELETE    = 0x2E
	VK_LWIN      = 0x5B
	VK_RWIN      = 0x5C
	VK_APPS      = 0x5D // Menu key
	VK_NUMPAD0   = 0x60
	VK_NUMPAD9   = 0x69
	VK_MULTIPLY  = 0x6A
	VK_ADD       = 0x6B
	VK_SEPARATOR = 0x6C
	VK_SUBTRACT  = 0x6D
	VK_DECIMAL   = 0x6E
	VK_DIVIDE    = 0x6F
	VK_F1        = 0x70
	VK_F12       = 0x7B
	VK_NUMLOCK   = 0x90
	VK_SCROLL    = 0x91
	VK_LSHIFT    = 0xA0
	VK_RSHIFT    = 0xA1
	VK_LCONTROL  = 0xA2
	VK_RCONTROL  = 0xA3
	VK_LMENU     = 0xA4
	VK_RMENU     = 0xA5

	VK_OEM_MINUS     = 0xBD
	VK_OEM_PLUS      = 0xBB
	VK_OEM_4         = 0xDB // [
	VK_OEM_6         = 0xDD // ]
	VK_OEM_5         = 0xDC // backslash
	VK_OEM_1         = 0xBA // ;
	VK_OEM_7         = 0xDE // '
	VK_OEM_3         = 0xC0 // `
	VK_OEM_COMMA     = 0xBC
	VK_OEM_PERIOD    = 0xBE
	VK_OEM_2         = 0xBF // /
)

// Mouse tracking
const (
	TME_LEAVE = 0x00000002
)

// System metrics
const (
	SM_CXSCREEN = 0
	SM_CYSCREEN = 1
)

// DPI
const (
	LOGPIXELSX               = 88
	LOGPIXELSY               = 90
	MONITOR_DEFAULTTONEAREST = 2
	MDT_EFFECTIVE_DPI        = 0
)

// Clipboard formats
const (
	CF_UNICODETEXT = 13
)

// Global memory flags
const (
	GMEM_MOVEABLE = 0x0002
)

// Cursor IDs
const (
	IDC_ARROW    = 32512
	IDC_IBEAM    = 32513
	IDC_WAIT     = 32514
	IDC_CROSS    = 32515
	IDC_SIZEWE   = 32644
	IDC_SIZENS   = 32645
	IDC_SIZENWSE = 32642
	IDC_SIZENESW = 32643
	IDC_SIZEALL  = 32646
	IDC_NO       = 32648
	IDC_HAND     = 32649
)

// GWLP constants
const (
	GWLP_USERDATA = -21
	GWLP_STYLE    = -16
	GWLP_EXSTYLE  = -20
)

// Class styles
const (
	CS_HREDRAW  = 0x0002
	CS_VREDRAW  = 0x0001
	CS_OWNDC    = 0x0020
	CS_DBLCLKS  = 0x0008
)

// ---- Win32 Structures ----

type WNDCLASSEXW struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm      uintptr
}

type MSG struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type POINT struct {
	X, Y int32
}

type RECT struct {
	Left, Top, Right, Bottom int32
}

type MINMAXINFO struct {
	PtReserved     POINT
	PtMaxSize      POINT
	PtMaxPosition  POINT
	PtMinTrackSize POINT
	PtMaxTrackSize POINT
}

type TRACKMOUSEEVENT struct {
	CbSize      uint32
	DwFlags     uint32
	HwndTrack   uintptr
	DwHoverTime uint32
}

type COMPOSITIONFORM struct {
	DwStyle uint32
	PtCurrentPos POINT
	RcArea  RECT
}

type CANDIDATEFORM struct {
	DwIndex uint32
	DwStyle uint32
	PtCurrentPos POINT
	RcArea  RECT
}

// ---- Helper functions ----

// utf16PtrFromString converts a Go string to a *uint16 for Win32 W functions.
func utf16PtrFromString(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

// utf16FromString converts a Go string to []uint16 with null terminator.
func utf16FromString(s string) []uint16 {
	r, _ := syscall.UTF16FromString(s)
	return r
}

// utf16ToString converts a null-terminated *uint16 to a Go string.
func utf16ToString(p *uint16) string {
	if p == nil {
		return ""
	}
	// Find length
	end := unsafe.Pointer(p)
	n := 0
	for *(*uint16)(end) != 0 {
		end = unsafe.Pointer(uintptr(end) + 2)
		n++
	}
	return syscall.UTF16ToString(unsafe.Slice(p, n))
}

// loword / hiword extract low/high 16-bit words from a uintptr.
func loword(l uintptr) int16 { return int16(l & 0xFFFF) }
func hiword(l uintptr) int16 { return int16((l >> 16) & 0xFFFF) }

// GET_X_LPARAM / GET_Y_LPARAM for mouse messages.
func getXLParam(lp uintptr) float32 { return float32(loword(lp)) }
func getYLParam(lp uintptr) float32 { return float32(hiword(lp)) }

// GET_WHEEL_DELTA_WPARAM
func getWheelDelta(wp uintptr) float32 { return float32(hiword(wp)) / 120.0 }

// MAKEINTRESOURCE
func makeIntResource(id uint16) *uint16 {
	return (*uint16)(unsafe.Pointer(uintptr(id)))
}

// High-resolution timer
var perfFrequency int64

func init() {
	procQueryPerformanceFrequency.Call(uintptr(unsafe.Pointer(&perfFrequency)))
}

func queryTimeMicroseconds() uint64 {
	var counter int64
	procQueryPerformanceCounter.Call(uintptr(unsafe.Pointer(&counter)))
	return uint64(counter) * 1_000_000 / uint64(perfFrequency)
}

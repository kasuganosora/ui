//go:build windows

package win32

import (
	"syscall"
	"unsafe"

	"github.com/kasuganosora/ui/platform"
)

// ---- COM GUIDs for TSF ----

// GUID is the COM GUID structure (128-bit, 16 bytes).
type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

var (
	CLSID_TF_ThreadMgr = GUID{0x529a9e6b, 0x6587, 0x4f23, [8]byte{0xab, 0x9e, 0x9c, 0x7d, 0x68, 0x3e, 0x3c, 0x50}}
	IID_ITfThreadMgr    = GUID{0xaa80e801, 0x2021, 0x11d2, [8]byte{0x93, 0xe0, 0x00, 0x60, 0xb0, 0x67, 0xb8, 0x6e}}
	IID_ITfDocumentMgr  = GUID{0xaa80e7f4, 0x2021, 0x11d2, [8]byte{0x93, 0xe0, 0x00, 0x60, 0xb0, 0x67, 0xb8, 0x6e}}
	IID_ITfSource       = GUID{0x4ea48a35, 0x60ae, 0x446f, [8]byte{0x8f, 0xd6, 0xe6, 0xa8, 0xd8, 0x24, 0x59, 0xf7}}
)

// ---- COM creation functions ----

var (
	procCoInitializeEx  = ole32.NewProc("CoInitializeEx")
	procCoCreateInstance = ole32.NewProc("CoCreateInstance")
	procCoUninitialize  = ole32.NewProc("CoUninitialize")
)

const (
	COINIT_APARTMENTTHREADED = 0x2
	CLSCTX_INPROC_SERVER     = 0x1
)

// ---- TSF Manager ----

// TSFManager wraps the Text Services Framework for modern IME input.
// It manages the COM lifecycle for ITfThreadMgr and provides a fallback
// path: if TSF initialization fails (e.g., on older Windows), the system
// falls back to the existing IMM32 code path transparently.
type TSFManager struct {
	initialized bool
	threadMgr   uintptr // *ITfThreadMgr COM pointer
	clientID    uint32  // TF_CLIENTID
	docMgr      uintptr // *ITfDocumentMgr
	context     uintptr // *ITfContext (editing context)
	window      *Window
}

// NewTSFManager creates and initializes the TSF system.
// If TSF is unavailable the returned manager will report IsActive() == false
// and all methods become safe no-ops, letting the caller fall back to IMM32.
func NewTSFManager(w *Window) *TSFManager {
	tsf := &TSFManager{window: w}

	// Initialize COM (apartment-threaded for UI thread)
	hr, _, _ := procCoInitializeEx.Call(0, COINIT_APARTMENTTHREADED)
	// S_OK (0) or S_FALSE (1, already initialized) are both acceptable
	if hr != 0 && hr != 1 {
		return tsf
	}

	// Create ITfThreadMgr via CoCreateInstance
	var threadMgr uintptr
	hr, _, _ = procCoCreateInstance.Call(
		uintptr(unsafe.Pointer(&CLSID_TF_ThreadMgr)),
		0,
		CLSCTX_INPROC_SERVER,
		uintptr(unsafe.Pointer(&IID_ITfThreadMgr)),
		uintptr(unsafe.Pointer(&threadMgr)),
	)
	if hr != 0 {
		return tsf // TSF not available, fall back to IMM32
	}
	tsf.threadMgr = threadMgr

	// Activate the thread manager.
	// ITfThreadMgr vtable layout (inherits IUnknown):
	//   [0] QueryInterface  [1] AddRef  [2] Release
	//   [3] Activate        [4] Deactivate
	//   [5] CreateDocumentMgr  [6] EnumDocumentMgrs
	//   [7] GetFocus        [8] SetFocus
	//   [9] AssociateFocus  [10] IsThreadFocus
	//   [11] GetFunctionProvider  [12] EnumFunctionProviders
	//   [13] GetGlobalCompartment
	vtbl := *(*[16]uintptr)(unsafe.Pointer(*(*uintptr)(unsafe.Pointer(threadMgr))))
	hr, _, _ = syscall.SyscallN(vtbl[3], threadMgr, uintptr(unsafe.Pointer(&tsf.clientID)))
	if hr != 0 {
		tsf.Release()
		return tsf
	}

	tsf.initialized = true
	return tsf
}

// IsActive returns true if TSF was successfully initialized.
func (tsf *TSFManager) IsActive() bool {
	return tsf != nil && tsf.initialized
}

// Release cleans up all COM objects held by the TSF manager.
// Safe to call multiple times and on a nil receiver.
func (tsf *TSFManager) Release() {
	if tsf == nil {
		return
	}
	if tsf.context != 0 {
		comRelease(tsf.context)
		tsf.context = 0
	}
	if tsf.docMgr != 0 {
		comRelease(tsf.docMgr)
		tsf.docMgr = 0
	}
	if tsf.threadMgr != 0 {
		// Deactivate before releasing
		if tsf.initialized {
			vtbl := *(*[16]uintptr)(unsafe.Pointer(*(*uintptr)(unsafe.Pointer(tsf.threadMgr))))
			syscall.SyscallN(vtbl[4], tsf.threadMgr) // ITfThreadMgr::Deactivate
		}
		comRelease(tsf.threadMgr)
		tsf.threadMgr = 0
	}
	tsf.initialized = false
}

// comRelease calls IUnknown::Release (vtable index 2) on a COM pointer.
func comRelease(ptr uintptr) {
	if ptr == 0 {
		return
	}
	vtbl := *(*[3]uintptr)(unsafe.Pointer(*(*uintptr)(unsafe.Pointer(ptr))))
	syscall.SyscallN(vtbl[2], ptr)
}

// SetFocusedEdit notifies TSF that a text editing control has gained focus.
// The provider will be stored on the window for use during IME callbacks.
func (tsf *TSFManager) SetFocusedEdit(provider platform.TSFTextProvider) {
	if !tsf.IsActive() {
		return
	}
	tsf.window.tsfProvider = provider
}

// ClearFocusedEdit notifies TSF that text editing has ended.
func (tsf *TSFManager) ClearFocusedEdit() {
	if !tsf.IsActive() {
		return
	}
	tsf.window.tsfProvider = nil
}

// SetCandidateWindowPosition sets the position of the IME candidate window.
// This uses the IMM32 candidate window positioning API for broad compatibility,
// as most modern IMEs on Windows still respect IMM32 positioning hints.
func (tsf *TSFManager) SetCandidateWindowPosition(x, y int32) {
	if !tsf.IsActive() {
		return
	}
	setIMECandidatePos(tsf.window, x, y)
}

// setIMECandidatePos positions the IME candidate window using IMM32.
// This is used by both the TSF manager and can be called directly for
// IMM32-only code paths.
func setIMECandidatePos(w *Window, x, y int32) {
	himc, _, _ := procImmGetContext.Call(w.hwnd)
	if himc == 0 {
		return
	}
	defer procImmReleaseContext.Call(w.hwnd, himc)

	cf := CANDIDATEFORM{
		DwStyle:      0x0020, // CFS_CANDIDATEPOS
		PtCurrentPos: POINT{X: x, Y: y},
	}
	procImmSetCandidateWindow.Call(himc, uintptr(unsafe.Pointer(&cf)))
}

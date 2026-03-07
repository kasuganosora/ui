//go:build windows

package win32

import (
	"syscall"
	"unsafe"
)

// getClipboardText reads Unicode text from the Windows clipboard.
func getClipboardText() string {
	ret, _, _ := procOpenClipboard.Call(0)
	if ret == 0 {
		return ""
	}
	defer procCloseClipboard.Call()

	handle, _, _ := procGetClipboardData.Call(CF_UNICODETEXT)
	if handle == 0 {
		return ""
	}

	ptr, _, _ := procGlobalLock.Call(handle)
	if ptr == 0 {
		return ""
	}
	defer procGlobalUnlock.Call(handle)

	return utf16PtrToString((*uint16)(unsafe.Pointer(ptr)))
}

// setClipboardText writes Unicode text to the Windows clipboard.
func setClipboardText(text string) {
	ret, _, _ := procOpenClipboard.Call(0)
	if ret == 0 {
		return
	}
	defer procCloseClipboard.Call()

	procEmptyClipboard.Call()

	u16 := utf16FromString(text)
	size := len(u16) * 2 // uint16 = 2 bytes

	hMem, _, _ := procGlobalAlloc.Call(GMEM_MOVEABLE, uintptr(size))
	if hMem == 0 {
		return
	}

	ptr, _, _ := procGlobalLock.Call(hMem)
	if ptr == 0 {
		procGlobalFree.Call(hMem)
		return
	}

	// Copy UTF-16 data
	dst := unsafe.Slice((*uint16)(unsafe.Pointer(ptr)), len(u16))
	copy(dst, u16)

	procGlobalUnlock.Call(hMem)
	procSetClipboardData.Call(CF_UNICODETEXT, hMem)
}

// utf16PtrToString reads a null-terminated UTF-16 string from a pointer.
func utf16PtrToString(p *uint16) string {
	if p == nil {
		return ""
	}
	end := unsafe.Pointer(p)
	n := 0
	for *(*uint16)(end) != 0 {
		end = unsafe.Pointer(uintptr(end) + 2)
		n++
		if n > 1<<20 { // safety limit: ~1M chars
			break
		}
	}
	return syscall.UTF16ToString(unsafe.Slice(p, n))
}

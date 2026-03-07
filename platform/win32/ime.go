//go:build windows

package win32

import (
	"syscall"
	"unsafe"

	"github.com/kasuganosora/ui/event"
)

// handleIMEComposition processes WM_IME_COMPOSITION and generates events.
func handleIMEComposition(w *Window, lParam uintptr) {
	himc, _, _ := procImmGetContext.Call(w.hwnd)
	if himc == 0 {
		return
	}
	defer procImmReleaseContext.Call(w.hwnd, himc)

	flags := uint32(lParam)

	// Composition string update
	if flags&GCS_COMPSTR != 0 {
		text := getCompositionString(himc, GCS_COMPSTR)
		cursorPos := getCompositionCursorPos(himc)

		w.p.pushEvent(event.Event{
			Type:               event.IMECompositionUpdate,
			Timestamp:          queryTimeMicroseconds(),
			IMECompositionText: text,
			IMECursorPos:       cursorPos,
			Text:               text,
		})
	}

	// Result string (committed text)
	if flags&GCS_RESULTSTR != 0 {
		text := getCompositionString(himc, GCS_RESULTSTR)
		if text != "" {
			w.p.pushEvent(event.Event{
				Type:      event.IMECompositionEnd,
				Timestamp: queryTimeMicroseconds(),
				Text:      text,
			})
		}
	}
}

// getCompositionString retrieves a composition string from IME context.
func getCompositionString(himc uintptr, flag uint32) string {
	// Get required buffer size (in bytes)
	size, _, _ := procImmGetCompositionStringW.Call(himc, uintptr(flag), 0, 0)
	if int32(size) <= 0 {
		return ""
	}

	buf := make([]uint16, size/2+1)
	procImmGetCompositionStringW.Call(himc, uintptr(flag),
		uintptr(unsafe.Pointer(&buf[0])), size,
	)

	return syscall.UTF16ToString(buf)
}

// getCompositionCursorPos retrieves the cursor position within the composition string.
func getCompositionCursorPos(himc uintptr) int {
	pos, _, _ := procImmGetCompositionStringW.Call(himc, GCS_CURSORPOS, 0, 0)
	return int(pos)
}

//go:build android

package freetype

import (
	"fmt"
	"os"
	"syscall"
)

// Types for FreeType handles (opaque pointers).
type ftLibrary uintptr
type ftFace uintptr

// loader holds FreeType function pointers.
type loader struct {
	handle uintptr

	ftInitFreeType  uintptr
	ftDoneFreeType  uintptr
	ftNewFace       uintptr
	ftNewMemoryFace uintptr
	ftDoneFace      uintptr
	ftSetPixelSizes uintptr
	ftLoadGlyph     uintptr
	ftRenderGlyph   uintptr
	ftGetCharIndex  uintptr
	ftGetKerning    uintptr
}

// newLoader returns an error on Android.
// FreeType dynamic loading on Android requires CGO_ENABLED=1 with NDK.
// This causes freetype.New() to return an error so the caller uses the mock engine.
func newLoader() (*loader, error) {
	return nil, fmt.Errorf("freetype: dynamic loading not available on Android without CGO")
}

func (l *loader) close() {}

// ftCall is a stub — never called when newLoader returns an error.
func ftCall(fn uintptr, args ...uintptr) int32 { return -1 }

// syscallN is a stub.
func syscallN(fn uintptr, args ...uintptr) (r1 uintptr, r2 uintptr, err syscall.Errno) {
	return 0, 0, syscall.ENOSYS
}

// readFile reads an entire file into memory.
func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

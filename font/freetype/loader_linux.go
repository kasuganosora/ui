//go:build linux

package freetype

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// Types for FreeType handles (opaque pointers).
type ftLibrary uintptr
type ftFace uintptr

// loader holds FreeType function pointers loaded via dlopen/dlsym.
type loader struct {
	handle uintptr

	ftInitFreeType  uintptr
	ftDoneFreeType  uintptr
	ftNewMemoryFace uintptr
	ftDoneFace      uintptr
	ftSetPixelSizes uintptr
	ftLoadGlyph     uintptr
	ftRenderGlyph   uintptr
	ftGetCharIndex  uintptr
	ftGetKerning    uintptr
}

// soNames lists candidate shared library names in priority order.
var soNames = []string{
	"libfreetype.so.6",
	"libfreetype.so",
}

// dlopen/dlsym/dlclose via manual syscall to libdl.
// On modern Linux, these are in libc (glibc exposes them directly).
// We load libdl.so.2 first, then use dlopen from it.
const (
	rtldLazy   = 0x00001
	rtldGlobal = 0x00100
)

func newLoader() (*loader, error) {
	// First, try to get dlopen from the process (it may already be linked)
	libdlNames := []string{"libdl.so.2", "libdl.so"}
	var dlHandle uintptr
	for _, name := range libdlNames {
		nameBytes := append([]byte(name), 0)
		// Use SYS_OPENAT to open libdl... but this is a file open, not dlopen.
		// Without CGO, we cannot call dlopen directly on Linux.
		// The proper zero-CGO approach requires either:
		//   1. Using Go's plugin package (limited)
		//   2. Inline assembly for syscall to __libc_dlopen_mode
		//   3. A helper binary
		_ = nameBytes
	}
	_ = dlHandle

	return nil, fmt.Errorf("freetype: Linux dynamic loading requires CGO or purego; " +
		"install purego and use the purego-based loader, or set CGO_ENABLED=1")
}

func (l *loader) close() {
	if l.handle != 0 {
		// dlclose would go here
		l.handle = 0
	}
}

// ftCall calls a FreeType function and returns the FT_Error result.
func ftCall(fn uintptr, args ...uintptr) int32 {
	ret, _, _ := syscall.SyscallN(fn, args...)
	return int32(ret)
}

// syscallN wraps syscall.SyscallN.
func syscallN(fn uintptr, args ...uintptr) (r1 uintptr, r2 uintptr, err syscall.Errno) {
	return syscall.SyscallN(fn, args...)
}

// readFile reads an entire file into memory.
func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// Silence unused import warnings for the stub.
var _ = unsafe.Pointer(nil)

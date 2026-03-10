//go:build darwin

package freetype

import "syscall"

// ftLibrary and ftFace are opaque FreeType handle types.
type ftLibrary uintptr
type ftFace uintptr

// ftCall invokes a FreeType function and returns the FT_Error result.
// On darwin the loader is never successfully loaded, so fn is always 0.
func ftCall(fn uintptr, args ...uintptr) int32 {
	// All function pointers are 0 on darwin (newLoader always fails).
	return -1
}

// syscallN is a stub on darwin. Function pointer calling without CGO requires assembly.
func syscallN(fn uintptr, args ...uintptr) (r1 uintptr, r2 uintptr, err syscall.Errno) {
	return 0, 0, 0
}

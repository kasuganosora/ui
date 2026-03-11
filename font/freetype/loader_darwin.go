//go:build darwin

package freetype

import (
	"fmt"
	"os"
	"syscall"

	"github.com/ebitengine/purego"
)

// Types for FreeType handles (opaque pointers).
type ftLibrary uintptr
type ftFace uintptr

// loader holds FreeType function pointers loaded via purego.
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

// newLoader opens libfreetype via purego and resolves all function pointers.
func newLoader() (*loader, error) {
	candidates := []string{
		"/usr/local/lib/libfreetype.dylib",         // Intel Homebrew
		"/opt/homebrew/lib/libfreetype.dylib",       // Apple Silicon Homebrew
		"libfreetype.6.dylib",                       // fallback: search path
		"/usr/local/lib/libfreetype.6.dylib",
		"/opt/homebrew/lib/libfreetype.6.dylib",
	}
	var h uintptr
	for _, name := range candidates {
		handle, err := purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_LOCAL)
		if err == nil {
			h = handle
			break
		}
	}
	if h == 0 {
		return nil, fmt.Errorf("freetype: libfreetype.dylib not found — install via: brew install freetype")
	}

	sym := func(name string) uintptr {
		s, _ := purego.Dlsym(h, name)
		return s
	}

	l := &loader{
		handle:          h,
		ftInitFreeType:  sym("FT_Init_FreeType"),
		ftDoneFreeType:  sym("FT_Done_FreeType"),
		ftNewFace:       sym("FT_New_Face"),
		ftNewMemoryFace: sym("FT_New_Memory_Face"),
		ftDoneFace:      sym("FT_Done_Face"),
		ftSetPixelSizes: sym("FT_Set_Pixel_Sizes"),
		ftLoadGlyph:     sym("FT_Load_Glyph"),
		ftRenderGlyph:   sym("FT_Render_Glyph"),
		ftGetCharIndex:  sym("FT_Get_Char_Index"),
		ftGetKerning:    sym("FT_Get_Kerning"),
	}
	if l.ftInitFreeType == 0 {
		return nil, fmt.Errorf("freetype: FT_Init_FreeType not found in library")
	}
	return l, nil
}

func (l *loader) close() {
	l.handle = 0 // purego does not expose Dlclose
}

// ftCall calls a FreeType function and returns the FT_Error result.
// FreeType functions return FT_Error (int32) in r1.
func ftCall(fn uintptr, args ...uintptr) int32 {
	r1, _, _ := purego.SyscallN(fn, args...)
	return int32(r1)
}

// syscallN wraps purego.SyscallN.
func syscallN(fn uintptr, args ...uintptr) (r1 uintptr, r2 uintptr, err syscall.Errno) {
	var e uintptr
	r1, r2, e = purego.SyscallN(fn, args...)
	err = syscall.Errno(e)
	return
}

// readFile reads an entire file into memory.
func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

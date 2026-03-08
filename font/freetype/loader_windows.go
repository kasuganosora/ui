//go:build windows

package freetype

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// Types for FreeType handles (opaque pointers).
type ftLibrary uintptr
type ftFace uintptr

// loader holds FreeType function pointers loaded from the DLL.
type loader struct {
	dll *syscall.LazyDLL

	ftInitFreeType    uintptr
	ftDoneFreeType    uintptr
	ftNewFace         uintptr
	ftNewMemoryFace   uintptr
	ftDoneFace        uintptr
	ftSetPixelSizes   uintptr
	ftLoadGlyph       uintptr
	ftRenderGlyph     uintptr
	ftGetCharIndex    uintptr
	ftGetKerning      uintptr
}

// dllNames lists candidate DLL names in priority order.
var dllNames = []string{
	"freetype.dll",
	"libfreetype-6.dll",
	"libfreetype.dll",
}

// extraSearchPaths returns additional DLL candidate paths.
func extraSearchPaths() []string {
	var paths []string

	// Next to the executable
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		paths = append(paths,
			filepath.Join(dir, "libfreetype.dll"),
			filepath.Join(dir, "freetype.dll"),
		)
	}

	// Relative to working directory (common in development)
	if wd, err := os.Getwd(); err == nil {
		paths = append(paths,
			filepath.Join(wd, "font", "freetype", "libfreetype.dll"),
			filepath.Join(wd, "libfreetype.dll"),
		)
		// Walk up to find the module root (go.mod) for tests running in subdirs
		dir := wd
		for i := 0; i < 5; i++ {
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
			candidate := filepath.Join(dir, "font", "freetype", "libfreetype.dll")
			if _, err := os.Stat(candidate); err == nil {
				paths = append(paths, candidate)
				break
			}
		}
	}

	return paths
}

func newLoader() (*loader, error) {
	var dll *syscall.LazyDLL
	var loadErr error

	// Try standard DLL names (system PATH search)
	for _, name := range dllNames {
		d := syscall.NewLazyDLL(name)
		if err := d.Load(); err == nil {
			dll = d
			break
		} else {
			loadErr = err
		}
	}

	// Try paths relative to executable
	if dll == nil {
		for _, absPath := range extraSearchPaths() {
			if _, err := os.Stat(absPath); err != nil {
				continue
			}
			d := syscall.NewLazyDLL(absPath)
			if err := d.Load(); err == nil {
				dll = d
				break
			} else {
				loadErr = err
			}
		}
	}

	if dll == nil {
		return nil, fmt.Errorf("cannot load FreeType library: %w", loadErr)
	}

	l := &loader{dll: dll}

	funcs := []struct {
		name string
		ptr  *uintptr
	}{
		{"FT_Init_FreeType", &l.ftInitFreeType},
		{"FT_Done_FreeType", &l.ftDoneFreeType},
		{"FT_New_Face", &l.ftNewFace},
		{"FT_New_Memory_Face", &l.ftNewMemoryFace},
		{"FT_Done_Face", &l.ftDoneFace},
		{"FT_Set_Pixel_Sizes", &l.ftSetPixelSizes},
		{"FT_Load_Glyph", &l.ftLoadGlyph},
		{"FT_Render_Glyph", &l.ftRenderGlyph},
		{"FT_Get_Char_Index", &l.ftGetCharIndex},
		{"FT_Get_Kerning", &l.ftGetKerning},
	}

	for _, f := range funcs {
		proc := dll.NewProc(f.name)
		if err := proc.Find(); err != nil {
			return nil, fmt.Errorf("freetype: %s not found: %w", f.name, err)
		}
		*f.ptr = proc.Addr()
	}

	return l, nil
}

func (l *loader) close() {
	// LazyDLL doesn't need explicit close on Windows
}

// ftCall calls a FreeType function and returns the FT_Error result.
func ftCall(fn uintptr, args ...uintptr) int32 {
	ret, _, _ := syscall.SyscallN(fn, args...)
	return int32(ret)
}

// syscallN wraps syscall.SyscallN for non-error-returning functions.
func syscallN(fn uintptr, args ...uintptr) (r1 uintptr, r2 uintptr, err syscall.Errno) {
	return syscall.SyscallN(fn, args...)
}

// readFile reads an entire file into memory.
func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

//go:build darwin

package freetype

import (
	"fmt"
	"os"
)

// loader holds FreeType function pointers loaded from the shared library.
// On darwin, dynamic loading of FreeType is not yet implemented;
// newLoader returns an error so freetype.New() gracefully falls back to the mock engine.
type loader struct {
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

// newLoader attempts to load libfreetype from known macOS paths.
// Returns an error if the library cannot be found.
func newLoader() (*loader, error) {
	// TODO: implement dlopen-based loading for macOS once a darwin
	// syscall-level dynamic loader is available in this package.
	// Typical locations: /opt/homebrew/lib/libfreetype.dylib,
	//                    /usr/local/lib/libfreetype.dylib
	return nil, fmt.Errorf("freetype: dynamic loading not yet implemented on darwin")
}

func (l *loader) close() {}

// readFile reads a file into memory.
func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

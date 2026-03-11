//go:build darwin

package ui

import (
	"fmt"
	"os"

	"github.com/kasuganosora/ui/font"
	"github.com/kasuganosora/ui/font/freetype"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/platform/darwin"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/render/metal"
	"github.com/kasuganosora/ui/render/vulkan"
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// newPlatform returns the Cocoa/macOS platform implementation.
func newPlatform() platform.Platform {
	return darwin.New()
}

// platformCreateBackend creates a render.Backend for macOS.
// Auto mode tries Metal first, then Vulkan (via MoltenVK) as fallback.
func platformCreateBackend(bt BackendType, win platform.Window) (render.Backend, error) {
	switch bt {
	case BackendMetal:
		b := metal.New()
		if err := b.Init(win); err != nil {
			return nil, fmt.Errorf("metal init: %w", err)
		}
		return b, nil
	case BackendVulkan:
		b := vulkan.New()
		if err := b.Init(win); err != nil {
			return nil, fmt.Errorf("vulkan init: %w", err)
		}
		return b, nil
	default: // BackendAuto: try Metal first, then Vulkan/MoltenVK
		b := metal.New()
		if err := b.Init(win); err == nil {
			return b, nil
		}
		v := vulkan.New()
		if err := v.Init(win); err != nil {
			return nil, fmt.Errorf(
				"no backend available on macOS (tried metal, vulkan/MoltenVK): %w\n"+
					"hint: update macOS (Metal requires 10.11+) or install MoltenVK (brew install molten-vk)",
				err,
			)
		}
		return v, nil
	}
}

// platformNewFontEngine loads FreeType on macOS.
// Returns nil on failure so the caller falls back to the mock engine.
func platformNewFontEngine() font.Engine {
	e, err := freetype.New()
	if err != nil {
		return nil
	}
	return e
}

// platformDefaultFont returns the path to the default CJK-capable font on macOS.
// PingFang SC is the system Chinese font shipped with macOS 10.11+.
func platformDefaultFont() string {
	paths := []string{
		"/System/Library/Fonts/PingFang.ttc",     // macOS 10.11+ (PingFang SC)
		"/Library/Fonts/Arial Unicode.ttf",        // older macOS fallback
		"/System/Library/Fonts/STHeiti Light.ttc", // Snow Leopard / early Lion
	}
	for _, p := range paths {
		if fileExists(p) {
			return p
		}
	}
	return paths[0] // best guess; RegisterFile will fail and app falls back to mock
}

// platformFallbackFonts returns additional fallback font paths for symbol glyphs.
func platformFallbackFonts() []string {
	candidates := []string{
		"/System/Library/Fonts/Apple Symbols.ttf",
		"/System/Library/Fonts/Symbol.ttf",
	}
	var result []string
	for _, p := range candidates {
		if fileExists(p) {
			result = append(result, p)
		}
	}
	return result
}

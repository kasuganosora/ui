//go:build linux && !android

package ui

import (
	"fmt"
	"os"

	"github.com/kasuganosora/ui/font"
	"github.com/kasuganosora/ui/font/freetype"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/platform/linux"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/render/vulkan"
)

// newPlatform returns the Linux X11 platform implementation.
func newPlatform() platform.Platform {
	return linux.New()
}

// platformCreateBackend creates a render.Backend for Linux.
// Currently only Vulkan is supported; OpenGL support is a future enhancement.
func platformCreateBackend(bt BackendType, win platform.Window) (render.Backend, error) {
	switch bt {
	case BackendVulkan:
		b := vulkan.New()
		if err := b.Init(win); err != nil {
			return nil, fmt.Errorf("vulkan init: %w", err)
		}
		return b, nil
	default: // BackendAuto: Vulkan
		b := vulkan.New()
		if err := b.Init(win); err != nil {
			return nil, fmt.Errorf("no backend available (vulkan failed): %w", err)
		}
		return b, nil
	}
}

// platformNewFontEngine loads FreeType from the system library path.
// Returns nil on failure so the caller falls back to the mock engine.
func platformNewFontEngine() font.Engine {
	e, err := freetype.New()
	if err != nil {
		return nil
	}
	return e
}

// platformDefaultFont returns the path to a CJK-capable system font on Linux.
// Tries common distro paths for Noto CJK and WenQuanYi fonts.
func platformDefaultFont() string {
	paths := []string{
		"/usr/share/fonts/noto-cjk/NotoSansCJK-Regular.ttc",
		"/usr/share/fonts/noto/NotoSansCJK-Regular.ttc",
		"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
		"/usr/share/fonts/noto/NotoSansCJKsc-Regular.otf",
		"/usr/share/fonts/google-noto-cjk/NotoSansCJK-Regular.ttc",
		"/usr/share/fonts/truetype/noto/NotoSansCJK-Regular.ttc",
		"/usr/share/fonts/truetype/wqy/wqy-microhei.ttc",
		"/usr/share/fonts/wenquanyi/wqy-microhei/wqy-microhei.ttc",
		"/usr/share/fonts/wqy-microhei/wqy-microhei.ttc",
		"/usr/share/fonts/liberation/LiberationSans-Regular.ttf",
		"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf",
		"/usr/share/fonts/TTF/DejaVuSans.ttf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
	}
	for _, p := range paths {
		if fileExists(p) {
			return p
		}
	}
	// Return first candidate as fallback (error will surface at font load time)
	return paths[0]
}

// platformFallbackFonts returns additional font paths for symbol/emoji glyphs.
func platformFallbackFonts() []string {
	candidates := []string{
		"/usr/share/fonts/noto/NotoSansSymbols-Regular.ttf",
		"/usr/share/fonts/truetype/noto/NotoSansSymbols-Regular.ttf",
		"/usr/share/fonts/noto/NotoSansSymbols2-Regular.ttf",
	}
	var result []string
	for _, p := range candidates {
		if fileExists(p) {
			result = append(result, p)
		}
	}
	return result
}

// fileExists reports whether the file at path exists and is readable.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

//go:build android

package ui

import (
	"fmt"
	"os"

	"github.com/kasuganosora/ui/font"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/platform/android"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/render/vulkan"
)

// newPlatform returns the Android platform implementation.
func newPlatform() platform.Platform {
	return android.New()
}

// platformCreateBackend creates a render.Backend for Android.
// Android uses the Vulkan loader (libvulkan.so) which is always present on
// Android 7.0+ (API level 24+).
func platformCreateBackend(bt BackendType, win platform.Window) (render.Backend, error) {
	switch bt {
	case BackendVulkan:
		b := vulkan.New()
		if err := b.Init(win); err != nil {
			return nil, fmt.Errorf("android vulkan init: %w", err)
		}
		return b, nil
	default: // BackendAuto: Vulkan only on Android
		b := vulkan.New()
		if err := b.Init(win); err != nil {
			return nil, fmt.Errorf("android: no backend available (vulkan failed): %w", err)
		}
		return b, nil
	}
}

// platformNewFontEngine returns nil — FreeType loading on Android requires
// the .so to be bundled in the APK. Use a mock engine or bundled font engine.
func platformNewFontEngine() font.Engine { return nil }

// platformDefaultFont returns the path to a CJK-capable system font on Android.
func platformDefaultFont() string {
	// Android includes Noto CJK fonts starting from Android 7.0.
	// Common paths on AOSP-based systems:
	candidates := []string{
		"/system/fonts/NotoSansCJK-Regular.ttc",
		"/system/fonts/NotoSansCJKsc-Regular.otf",
		"/system/fonts/DroidSansChinese.ttf",
		"/system/fonts/DroidSans.ttf",
		"/system/fonts/Roboto-Regular.ttf",
	}
	for _, p := range candidates {
		if fileExists(p) {
			return p
		}
	}
	return "/system/fonts/Roboto-Regular.ttf"
}

// platformFallbackFonts returns additional fallback font paths for symbol glyphs.
func platformFallbackFonts() []string { return nil }

// fileExists reports whether the file at path exists and is readable.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

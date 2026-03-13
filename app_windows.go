//go:build windows

package ui

import (
	"fmt"

	"github.com/kasuganosora/ui/font"
	"github.com/kasuganosora/ui/font/freetype"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/platform/win32"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/render/dx9"
	"github.com/kasuganosora/ui/render/dx11"
	"github.com/kasuganosora/ui/render/gl"
	"github.com/kasuganosora/ui/render/vulkan"
)

// newPlatform returns the Win32 platform implementation.
func newPlatform() platform.Platform {
	return win32.New()
}

// platformCreateBackend creates a render.Backend for Windows.
// Auto mode: VK → GL → DX11.
func platformCreateBackend(bt BackendType, win platform.Window) (render.Backend, error) {
	switch bt {
	case BackendDX9:
		b := dx9.New()
		if err := b.Init(win); err != nil {
			return nil, fmt.Errorf("dx9 init: %w", err)
		}
		return b, nil
	case BackendDX11:
		b := dx11.New()
		if err := b.Init(win); err != nil {
			return nil, fmt.Errorf("dx11 init: %w", err)
		}
		return b, nil
	case BackendVulkan:
		b := vulkan.New()
		if err := b.Init(win); err != nil {
			return nil, fmt.Errorf("vulkan init: %w", err)
		}
		return b, nil
	case BackendOpenGL:
		b := gl.New()
		if err := b.Init(win); err != nil {
			return nil, fmt.Errorf("gl init: %w", err)
		}
		return b, nil
	default: // BackendAuto: VK → GL → DX11
		vk := vulkan.New()
		if err := vk.Init(win); err == nil {
			return vk, nil
		}
		g := gl.New()
		if err := g.Init(win); err == nil {
			return g, nil
		}
		d := dx11.New()
		if err := d.Init(win); err != nil {
			return nil, fmt.Errorf("no backend available (tried vulkan, gl, dx11): %w", err)
		}
		return d, nil
	}
}

// platformNewFontEngine loads FreeType from the Windows DLL search path.
// Returns nil on failure so the caller falls back to the mock engine.
func platformNewFontEngine() font.Engine {
	e, err := freetype.New()
	if err != nil {
		return nil
	}
	return e
}

// platformDefaultFont returns the path to the default CJK-capable font on Windows.
func platformDefaultFont() string {
	return `C:\Windows\Fonts\msyh.ttc` // Microsoft YaHei
}

// platformFallbackFonts returns additional fallback font paths for symbol glyphs.
func platformFallbackFonts() []string {
	return []string{
		`C:\Windows\Fonts\seguiemj.ttf`, // Segoe UI Emoji: color emoji (COLR/CPAL)
		`C:\Windows\Fonts\seguisym.ttf`, // Segoe UI Symbol: arrows, hearts, misc Unicode
	}
}

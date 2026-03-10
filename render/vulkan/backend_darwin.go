//go:build darwin

package vulkan

import (
	"fmt"
	"unsafe"

	"github.com/kasuganosora/ui/platform"
)

// createPlatformSurface creates a Vulkan surface from a macOS NSView handle.
// Requires VK_MVK_macos_surface or VK_EXT_metal_surface extension (MoltenVK).
func (b *Backend) createPlatformSurface(window platform.Window) (SurfaceKHR, error) {
	handle := window.NativeHandle()
	if handle == 0 {
		return 0, fmt.Errorf("vulkan: invalid native window handle")
	}

	// VkMacOSSurfaceCreateInfoMVK / VkMetalSurfaceCreateInfoEXT
	// Attempt VK_EXT_metal_surface first (supported by MoltenVK 1.1+).
	type metalSurfaceCreateInfo struct {
		sType  uint32
		pNext  uintptr
		flags  uint32
		pLayer uintptr // CAMetalLayer* (NSView.layer after view.wantsLayer = YES)
	}
	const vkStructureTypeMetalSurfaceCreateInfoEXT = 1000217000

	info := metalSurfaceCreateInfo{
		sType:  vkStructureTypeMetalSurfaceCreateInfoEXT,
		pLayer: handle,
	}

	var surface SurfaceKHR
	r, _, _ := syscallN(b.loader.vkCreateMetalSurfaceEXT,
		uintptr(b.instance),
		uintptr(unsafe.Pointer(&info)),
		0,
		uintptr(unsafe.Pointer(&surface)),
	)
	if r == 0 {
		return surface, nil
	}

	// Fallback: try VK_MVK_macos_surface (legacy MoltenVK).
	type macosSurfaceCreateInfo struct {
		sType  uint32
		pNext  uintptr
		flags  uint32
		pView  uintptr // NSView*
	}
	const vkStructureTypeMacosSurfaceCreateInfoMVK = 1000123000

	infoMVK := macosSurfaceCreateInfo{
		sType: vkStructureTypeMacosSurfaceCreateInfoMVK,
		pView: handle,
	}
	r, _, _ = syscallN(b.loader.vkCreateMacOSSurfaceMVK,
		uintptr(b.instance),
		uintptr(unsafe.Pointer(&infoMVK)),
		0,
		uintptr(unsafe.Pointer(&surface)),
	)
	if r != 0 {
		return 0, fmt.Errorf("vulkan: surface creation failed (VkResult %d)", r)
	}
	return surface, nil
}

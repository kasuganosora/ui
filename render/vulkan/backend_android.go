//go:build android

package vulkan

import (
	"fmt"
	"unsafe"

	"github.com/kasuganosora/ui/platform"
)

// createPlatformSurface creates a Vulkan surface for an Android window.
// Uses VK_KHR_android_surface with an ANativeWindow* handle.
func (b *Backend) createPlatformSurface(window platform.Window) (SurfaceKHR, error) {
	handle := window.NativeHandle()
	if handle == 0 {
		return 0, fmt.Errorf("vulkan: invalid ANativeWindow handle — window not yet initialized by Android runtime")
	}

	if b.loader.vkCreateAndroidSurfaceKHR == 0 {
		return 0, fmt.Errorf("vulkan: VK_KHR_android_surface not available")
	}

	// VkAndroidSurfaceCreateInfoKHR
	type androidSurfaceCreateInfoKHR struct {
		SType  uint32  // VK_STRUCTURE_TYPE_ANDROID_SURFACE_CREATE_INFO_KHR = 1000008000
		PNext  uintptr // null
		Flags  uint32  // reserved, must be 0
		Window uintptr // ANativeWindow*
	}

	const vkStructureTypeAndroidSurfaceCreateInfoKHR uint32 = 1000008000

	info := androidSurfaceCreateInfoKHR{
		SType:  vkStructureTypeAndroidSurfaceCreateInfoKHR,
		PNext:  0,
		Flags:  0,
		Window: handle,
	}

	var surface SurfaceKHR
	r, _, _ := syscallN(b.loader.vkCreateAndroidSurfaceKHR,
		uintptr(b.instance),
		uintptr(unsafe.Pointer(&info)),
		0, // pAllocator = NULL
		uintptr(unsafe.Pointer(&surface)),
	)
	if Result(r) != Success {
		return 0, fmt.Errorf("vulkan: vkCreateAndroidSurfaceKHR failed: %v", Result(r))
	}
	return surface, nil
}

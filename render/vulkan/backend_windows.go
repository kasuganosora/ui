//go:build windows

package vulkan

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/kasuganosora/ui/platform"
)

var (
	kernel32        = syscall.NewLazyDLL("kernel32.dll")
	getModuleHandle = kernel32.NewProc("GetModuleHandleW")
)

// createPlatformSurface creates a Vulkan surface from a Win32 window handle.
func (b *Backend) createPlatformSurface(window platform.Window) (SurfaceKHR, error) {
	handle := window.NativeHandle()
	if handle == 0 {
		return 0, fmt.Errorf("vulkan: window native handle is nil")
	}

	// Get the HINSTANCE for the current process
	hinstance, _, _ := getModuleHandle.Call(0)
	if hinstance == 0 {
		return 0, fmt.Errorf("vulkan: GetModuleHandle failed")
	}

	createInfo := Win32SurfaceCreateInfoKHR{
		SType:     StructureTypeWin32SurfaceCreateInfoKHR,
		Hinstance: hinstance,
		Hwnd:      handle,
	}

	var surface SurfaceKHR
	r, _, _ := syscallN(b.loader.vkCreateWin32SurfaceKHR,
		uintptr(b.instance),
		uintptr(unsafe.Pointer(&createInfo)),
		0,
		uintptr(unsafe.Pointer(&surface)),
	)
	if Result(r) != Success {
		return 0, fmt.Errorf("vulkan: vkCreateWin32SurfaceKHR failed: %v", Result(r))
	}
	return surface, nil
}

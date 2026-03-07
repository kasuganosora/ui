//go:build windows

package vulkan

import (
	"fmt"
	"syscall"
	"unsafe"
)

// platformLib wraps a Windows DLL handle.
type platformLib interface {
	lookup(name string) (uintptr, error)
	close()
}

type windowsLib struct {
	dll *syscall.LazyDLL
}

func openVulkanLib() (platformLib, error) {
	dll := syscall.NewLazyDLL("vulkan-1.dll")
	if err := dll.Load(); err != nil {
		return nil, fmt.Errorf("failed to load vulkan-1.dll: %w", err)
	}
	return &windowsLib{dll: dll}, nil
}

func (w *windowsLib) lookup(name string) (uintptr, error) {
	proc := w.dll.NewProc(name)
	if err := proc.Find(); err != nil {
		return 0, err
	}
	return proc.Addr(), nil
}

func (w *windowsLib) close() {
	// LazyDLL doesn't need explicit close on Windows
}

// syscall3 calls a Vulkan function with up to 3 arguments.
func syscall3(fn uintptr, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	return syscall.SyscallN(fn, a1, a2, a3)
}

// syscall6 calls a Vulkan function with up to 6 arguments.
func syscall6(fn uintptr, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	return syscall.SyscallN(fn, a1, a2, a3, a4, a5, a6)
}

// syscallN calls a Vulkan function with N arguments.
func syscallN(fn uintptr, args ...uintptr) (r1 uintptr, r2 uintptr, err syscall.Errno) {
	return syscall.SyscallN(fn, args...)
}

// Win32 surface creation info
type Win32SurfaceCreateInfoKHR struct {
	SType     StructureType
	PNext     unsafe.Pointer
	Flags     uint32
	Hinstance uintptr
	Hwnd      uintptr
}

// CreateWin32SurfaceKHR creates a Vulkan surface for a Win32 window.
func (l *Loader) CreateWin32SurfaceKHR(instance Instance, hinstance, hwnd uintptr) (SurfaceKHR, error) {
	if l.vkCreateWin32SurfaceKHR == 0 {
		return 0, fmt.Errorf("vulkan: VK_KHR_win32_surface not available")
	}
	createInfo := Win32SurfaceCreateInfoKHR{
		SType:     StructureTypeWin32SurfaceCreateInfoKHR,
		Hinstance: hinstance,
		Hwnd:      hwnd,
	}
	var surface SurfaceKHR
	r, _, _ := syscallN(l.vkCreateWin32SurfaceKHR,
		uintptr(instance),
		uintptr(unsafe.Pointer(&createInfo)),
		0, // allocator
		uintptr(unsafe.Pointer(&surface)),
	)
	if Result(r) != Success {
		return 0, fmt.Errorf("vulkan: vkCreateWin32SurfaceKHR failed: %v", Result(r))
	}
	return surface, nil
}

// requiredSurfaceExtensions returns platform-specific surface extensions.
func requiredSurfaceExtensions() []string {
	return []string{"VK_KHR_surface", "VK_KHR_win32_surface"}
}

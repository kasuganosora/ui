//go:build darwin

package vulkan

import (
	"fmt"
	"syscall"

	"github.com/ebitengine/purego"
)

// platformLib abstracts the platform-specific shared library handle.
type platformLib interface {
	lookup(name string) (uintptr, error)
	close()
}

// darwinLib wraps a dylib handle loaded via purego.Dlopen.
type darwinLib struct {
	handle uintptr // library handle from purego.Dlopen
}

func (l *darwinLib) lookup(name string) (uintptr, error) {
	sym, err := purego.Dlsym(l.handle, name)
	if err != nil {
		return 0, fmt.Errorf("vulkan: symbol %q not found: %w", name, err)
	}
	return sym, nil
}

func (l *darwinLib) close() {
	// purego doesn't expose Dlclose; zero the handle to prevent reuse.
	l.handle = 0
}

// openVulkanLib attempts to load libvulkan or MoltenVK from standard locations.
func openVulkanLib() (platformLib, error) {
	candidates := []string{
		"libvulkan.1.dylib",
		"libvulkan.dylib",
		"/opt/homebrew/lib/libvulkan.1.dylib",
		"/usr/local/lib/libvulkan.1.dylib",
		"/usr/local/lib/libMoltenVK.dylib",
		"/opt/homebrew/lib/libMoltenVK.dylib",
	}
	for _, name := range candidates {
		handle, err := purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_LOCAL)
		if err == nil {
			return &darwinLib{handle: uintptr(handle)}, nil
		}
	}
	return nil, fmt.Errorf("vulkan: libvulkan / MoltenVK not found on macOS; install with: brew install molten-vk")
}

// syscall3 calls a function with up to 3 arguments using purego.SyscallN.
// Vulkan functions only use integer/pointer args, so SyscallN is safe.
func syscall3(fn uintptr, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	var e uintptr
	r1, r2, e = purego.SyscallN(fn, a1, a2, a3)
	err = syscall.Errno(e)
	return
}

// syscallN calls a Vulkan function with a variable number of integer/pointer arguments.
func syscallN(fn uintptr, args ...uintptr) (r1 uintptr, r2 uintptr, err syscall.Errno) {
	var e uintptr
	r1, r2, e = purego.SyscallN(fn, args...)
	err = syscall.Errno(e)
	return
}

// requiredSurfaceExtensions returns the Vulkan surface extensions required on macOS.
func requiredSurfaceExtensions() []string {
	return []string{"VK_KHR_surface", "VK_EXT_metal_surface"}
}

//go:build linux && !android

package vulkan

import (
	"fmt"
	"syscall"

	"github.com/ebitengine/purego"
)

// platformLib wraps a Linux shared library handle loaded via purego.
type platformLib interface {
	lookup(name string) (uintptr, error)
	close()
}

type linuxLib struct {
	handle uintptr
}

// openVulkanLib tries to load the Vulkan ICD loader from the standard Linux paths.
func openVulkanLib() (platformLib, error) {
	candidates := []string{
		"libvulkan.so.1",
		"libvulkan.so",
	}
	for _, name := range candidates {
		h, err := purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_LOCAL)
		if err == nil {
			return &linuxLib{handle: h}, nil
		}
	}
	return nil, fmt.Errorf("failed to load libvulkan.so.1 or libvulkan.so — is the Vulkan ICD loader installed?")
}

func (l *linuxLib) lookup(name string) (uintptr, error) {
	sym, err := purego.Dlsym(l.handle, name)
	if err != nil {
		return 0, fmt.Errorf("vulkan: symbol %q not found: %w", name, err)
	}
	return sym, nil
}

func (l *linuxLib) close() {
	// purego does not expose Dlclose; OS reclaims on process exit.
}

// syscall3 calls a Vulkan function with exactly 3 arguments.
func syscall3(fn uintptr, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	var e uintptr
	r1, r2, e = purego.SyscallN(fn, a1, a2, a3)
	err = syscall.Errno(e)
	return
}

// syscall6 calls a Vulkan function with exactly 6 arguments.
func syscall6(fn uintptr, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	var e uintptr
	r1, r2, e = purego.SyscallN(fn, a1, a2, a3, a4, a5, a6)
	err = syscall.Errno(e)
	return
}

// syscallN calls a Vulkan function with N arguments.
func syscallN(fn uintptr, args ...uintptr) (r1 uintptr, r2 uintptr, err syscall.Errno) {
	var e uintptr
	r1, r2, e = purego.SyscallN(fn, args...)
	err = syscall.Errno(e)
	return
}

// requiredSurfaceExtensions returns the Vulkan instance extensions needed for
// creating an X11 (Xlib) surface on Linux.
func requiredSurfaceExtensions() []string {
	return []string{"VK_KHR_surface", "VK_KHR_xlib_surface"}
}

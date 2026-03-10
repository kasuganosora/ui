//go:build darwin

package vulkan

import (
	"fmt"
	"syscall"
)

// platformLib abstracts the platform-specific shared library handle.
// Must match the interface expected by loader.go (same as loader_windows.go).
type platformLib interface {
	lookup(name string) (uintptr, error)
	close()
}

// darwinLib is a stub — MoltenVK/Vulkan dynamic loading on darwin requires
// CGO, assembly, or the purego package for function pointer calls.
// openVulkanLib returns an error so the backend gracefully reports unavailability.
type darwinLib struct{}

func (l *darwinLib) lookup(name string) (uintptr, error) {
	return 0, fmt.Errorf("vulkan: darwin loader stub — %s not available", name)
}

func (l *darwinLib) close() {}

// openVulkanLib is a stub on darwin.
// TODO: implement using assembly (call_darwin_arm64.s) or github.com/ebitengine/purego.
func openVulkanLib() (platformLib, error) {
	return nil, fmt.Errorf("vulkan: Vulkan/MoltenVK loading not yet implemented on darwin (requires assembly or purego for function pointer calls)")
}

// syscall3 calls a function with up to 3 arguments.
// Stub on darwin — function pointer calling without CGO requires assembly.
func syscall3(fn uintptr, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	return syscallN(fn, a1, a2, a3)
}

// syscallN is a stub on darwin.
// Real implementation would need platform-specific assembly to call C function pointers.
func syscallN(fn uintptr, args ...uintptr) (r1 uintptr, r2 uintptr, err syscall.Errno) {
	_ = fn
	_ = args
	return 0, 0, 0
}

// requiredSurfaceExtensions returns the Vulkan surface extensions required on macOS.
func requiredSurfaceExtensions() []string {
	return []string{"VK_KHR_surface", "VK_EXT_metal_surface"}
}

//go:build android

package vulkan

import (
	"fmt"
)

// platformLib wraps an Android shared library handle.
// Android Vulkan loading requires CGO_ENABLED=1 with the NDK.
// This stub causes the Vulkan backend to return an error on Android.
type platformLib interface {
	lookup(name string) (uintptr, error)
	close()
}

type androidLib struct{}

// openVulkanLib returns an error on Android without CGO.
// Full Android Vulkan support requires CGO_ENABLED=1 with the Android NDK.
func openVulkanLib() (platformLib, error) {
	return nil, fmt.Errorf("vulkan: Android dynamic loading requires CGO_ENABLED=1 with NDK")
}

func (a *androidLib) lookup(name string) (uintptr, error) {
	return 0, fmt.Errorf("vulkan: Android not initialized")
}

func (a *androidLib) close() {}

// syscall3, syscall6, syscallN are stubs — never called because openVulkanLib
// returns an error before any function pointers are resolved on Android.

func syscall3(fn uintptr, a1, a2, a3 uintptr) (r1, r2 uintptr, err error) { return }
func syscall6(fn uintptr, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err error) { return }
func syscallN(fn uintptr, args ...uintptr) (r1, r2 uintptr, err error)      { return }

// requiredSurfaceExtensions returns the Vulkan instance extensions needed for
// creating an Android (ANativeWindow) surface.
func requiredSurfaceExtensions() []string {
	return []string{"VK_KHR_surface", "VK_KHR_android_surface"}
}

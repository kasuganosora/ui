//go:build linux && !android

package vulkan

import (
	"fmt"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/kasuganosora/ui/platform"
)

// xlibWindowProvider is an optional interface that Linux platform windows may
// implement to expose the X11 Display pointer alongside the Window ID.
type xlibWindowProvider interface {
	// XDisplay returns the X11 Display* as a uintptr.
	XDisplay() uintptr
}

// createPlatformSurface creates a Vulkan surface for a Linux X11 window.
// Uses VK_KHR_xlib_surface which requires both the Display* and the Window ID.
func (b *Backend) createPlatformSurface(window platform.Window) (SurfaceKHR, error) {
	handle := window.NativeHandle()
	if handle == 0 {
		return 0, fmt.Errorf("vulkan: invalid native X11 window handle")
	}

	if b.loader.vkCreateXlibSurfaceKHR == 0 {
		return 0, fmt.Errorf("vulkan: VK_KHR_xlib_surface not available")
	}

	// Obtain the X11 Display* pointer.
	// If the platform window exposes it via the xlibWindowProvider interface,
	// use that. Otherwise open a new display connection for surface creation.
	var dpy uintptr

	if provider, ok := window.(xlibWindowProvider); ok {
		dpy = provider.XDisplay()
	}

	if dpy == 0 {
		// Fall back: open the default X11 display for surface creation.
		// This is safe because Vulkan manages its own X server connection.
		if h, err := purego.Dlopen("libX11.so.6", purego.RTLD_LAZY|purego.RTLD_GLOBAL); err == nil {
			if fn, err2 := purego.Dlsym(h, "XOpenDisplay"); err2 == nil {
				dpy, _, _ = purego.SyscallN(fn, 0) // NULL = $DISPLAY
			}
		}
	}

	if dpy == 0 {
		return 0, fmt.Errorf("vulkan: XOpenDisplay failed — cannot obtain X11 Display for surface creation")
	}

	// VkXlibSurfaceCreateInfoKHR
	type xlibSurfaceCreateInfoKHR struct {
		SType  uint32  // VK_STRUCTURE_TYPE_XLIB_SURFACE_CREATE_INFO_KHR = 1000004000
		PNext  uintptr // null
		Flags  uint32  // reserved, must be 0
		Dpy    uintptr // Display*
		Window uint64  // Window (XID)
	}

	const vkStructureTypeXlibSurfaceCreateInfoKHR uint32 = 1000004000

	info := xlibSurfaceCreateInfoKHR{
		SType:  vkStructureTypeXlibSurfaceCreateInfoKHR,
		PNext:  0,
		Flags:  0,
		Dpy:    dpy,
		Window: uint64(handle),
	}

	var surface SurfaceKHR
	r, _, _ := syscallN(b.loader.vkCreateXlibSurfaceKHR,
		uintptr(b.instance),
		uintptr(unsafe.Pointer(&info)),
		0, // pAllocator = NULL
		uintptr(unsafe.Pointer(&surface)),
	)
	if Result(r) != Success {
		return 0, fmt.Errorf("vulkan: vkCreateXlibSurfaceKHR failed: %v", Result(r))
	}
	return surface, nil
}

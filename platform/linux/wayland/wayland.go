//go:build linux && !android

// Package wayland implements the platform.Platform interface for Linux using Wayland.
// It loads libwayland-client.so.0 dynamically via purego (zero CGO).
//
// Protocol coverage:
//   - wl_display, wl_registry, wl_compositor, wl_surface
//   - xdg_wm_base, xdg_surface, xdg_toplevel (stable XDG shell)
//   - wl_seat, wl_pointer, wl_keyboard (input)
//
// IME (text-input protocol v3) is stubbed as a future TODO.
package wayland

import (
	"unsafe"

	"github.com/ebitengine/purego"
)

// ---- Function pointer variables ----

var (
	fnWlDisplayConnect                     uintptr
	fnWlDisplayDisconnect                  uintptr
	fnWlDisplayFlush                       uintptr
	fnWlDisplayRoundtrip                   uintptr
	fnWlDisplayDispatch                    uintptr
	fnWlDisplayDispatchPending             uintptr
	fnWlDisplayGetFd                       uintptr
	fnWlProxyDestroy                       uintptr
	fnWlProxyMarshal                       uintptr
	fnWlProxyMarshalConstructorVersioned   uintptr
	fnWlProxyAddListener                   uintptr
	fnWlProxyGetVersion                    uintptr
)

func init() {
	h, err := purego.Dlopen("libwayland-client.so.0", purego.RTLD_LAZY|purego.RTLD_GLOBAL)
	if err != nil {
		return // Wayland not available; platform will fail at Init time
	}

	sym := func(name string) uintptr {
		s, _ := purego.Dlsym(h, name)
		return s
	}

	fnWlDisplayConnect = sym("wl_display_connect")
	fnWlDisplayDisconnect = sym("wl_display_disconnect")
	fnWlDisplayFlush = sym("wl_display_flush")
	fnWlDisplayRoundtrip = sym("wl_display_roundtrip")
	fnWlDisplayDispatch = sym("wl_display_dispatch")
	fnWlDisplayDispatchPending = sym("wl_display_dispatch_pending")
	fnWlDisplayGetFd = sym("wl_display_get_fd")
	fnWlProxyDestroy = sym("wl_proxy_destroy")
	fnWlProxyMarshal = sym("wl_proxy_marshal")
	fnWlProxyMarshalConstructorVersioned = sym("wl_proxy_marshal_constructor_versioned")
	fnWlProxyAddListener = sym("wl_proxy_add_listener")
	fnWlProxyGetVersion = sym("wl_proxy_get_version")
}

// ---- Wayland opaque proxy type ----

// WlProxy represents a wl_proxy* — the base type for all Wayland objects.
type WlProxy uintptr

// WlDisplay represents a wl_display*.
type WlDisplay = WlProxy

// WlRegistry represents a wl_registry*.
type WlRegistry = WlProxy

// WlCompositor represents a wl_compositor*.
type WlCompositor = WlProxy

// WlSurface represents a wl_surface*.
type WlSurface = WlProxy

// WlSeat represents a wl_seat*.
type WlSeat = WlProxy

// WlPointer represents a wl_pointer*.
type WlPointer = WlProxy

// WlKeyboard represents a wl_keyboard*.
type WlKeyboard = WlProxy

// WlCallback represents a wl_callback*.
type WlCallback = WlProxy

// XdgWmBase represents an xdg_wm_base* (stable XDG shell).
type XdgWmBase = WlProxy

// XdgSurface represents an xdg_surface*.
type XdgSurface = WlProxy

// XdgToplevel represents an xdg_toplevel*.
type XdgToplevel = WlProxy

// WlInterface describes a Wayland interface (simplified; only name is used for binding).
type WlInterface struct {
	Name        *byte
	Version     int32
	MethodCount int32
	Methods     uintptr
	EventCount  int32
	Events      uintptr
}

// Wayland request opcodes
const (
	// wl_registry
	wlRegistryBind = 0

	// wl_compositor
	wlCompositorCreateSurface = 0

	// wl_surface
	wlSurfaceDestroy   = 0
	wlSurfaceCommit    = 6

	// xdg_wm_base
	xdgWmBaseDestroy    = 0
	xdgWmBasePong       = 3
	xdgWmBaseGetXdgSurface = 2

	// xdg_surface
	xdgSurfaceDestroy      = 0
	xdgSurfaceGetToplevel  = 1
	xdgSurfaceAckConfigure = 4

	// xdg_toplevel
	xdgToplevelDestroy   = 0
	xdgToplevelSetTitle  = 2
	xdgToplevelSetMinSize = 7
	xdgToplevelSetMaxSize = 8
	xdgToplevelUnsetFullscreen = 12
	xdgToplevelSetFullscreen   = 11

	// wl_seat
	wlSeatGetPointer  = 0
	wlSeatGetKeyboard = 1

	// wl_pointer
	wlPointerRelease = 1

	// wl_keyboard
	wlKeyboardRelease = 0
)

// ---- Wayland interface name strings ----

// Null-terminated C strings for interface names used in wl_registry_bind.
var (
	ifaceNameCompositor = append([]byte("wl_compositor"), 0)
	ifaceNameSeat       = append([]byte("wl_seat"), 0)
	ifaceNameXdgWmBase  = append([]byte("xdg_wm_base"), 0)
)

// ---- Low-level wrappers ----

// wlDisplayConnect connects to the Wayland compositor.
// Pass nil for the default socket ($WAYLAND_DISPLAY or "wayland-0").
func wlDisplayConnect(name *byte) WlDisplay {
	r, _, _ := purego.SyscallN(fnWlDisplayConnect, uintptr(unsafe.Pointer(name)))
	return WlDisplay(r)
}

// wlDisplayDisconnect disconnects from the compositor.
func wlDisplayDisconnect(dpy WlDisplay) {
	purego.SyscallN(fnWlDisplayDisconnect, uintptr(dpy))
}

// wlDisplayFlush flushes the outgoing request buffer.
func wlDisplayFlush(dpy WlDisplay) int {
	r, _, _ := purego.SyscallN(fnWlDisplayFlush, uintptr(dpy))
	return int(int32(r))
}

// wlDisplayRoundtrip blocks until the compositor has processed all pending requests.
func wlDisplayRoundtrip(dpy WlDisplay) int {
	r, _, _ := purego.SyscallN(fnWlDisplayRoundtrip, uintptr(dpy))
	return int(int32(r))
}

// wlDisplayDispatch dispatches pending events (blocking if no events).
func wlDisplayDispatch(dpy WlDisplay) int {
	r, _, _ := purego.SyscallN(fnWlDisplayDispatch, uintptr(dpy))
	return int(int32(r))
}

// wlDisplayDispatchPending dispatches any already-queued events (non-blocking).
func wlDisplayDispatchPending(dpy WlDisplay) int {
	r, _, _ := purego.SyscallN(fnWlDisplayDispatchPending, uintptr(dpy))
	return int(int32(r))
}

// wlDisplayGetFd returns the file descriptor for the Wayland socket.
func wlDisplayGetFd(dpy WlDisplay) int {
	r, _, _ := purego.SyscallN(fnWlDisplayGetFd, uintptr(dpy))
	return int(int32(r))
}

// wlProxyDestroy destroys a Wayland proxy object.
func wlProxyDestroy(proxy WlProxy) {
	if proxy != 0 {
		purego.SyscallN(fnWlProxyDestroy, uintptr(proxy))
	}
}

// wlProxyAddListener attaches an event listener (callback struct) to a proxy.
func wlProxyAddListener(proxy WlProxy, impl unsafe.Pointer, data unsafe.Pointer) int {
	r, _, _ := purego.SyscallN(fnWlProxyAddListener,
		uintptr(proxy),
		uintptr(impl),
		uintptr(data),
	)
	return int(int32(r))
}

// wlProxyMarshalConstructorVersioned creates a new proxy by sending a request
// that returns a new object, using an explicit version.
func wlProxyMarshalConstructorVersioned(proxy WlProxy, opcode uint32, iface *WlInterface, version uint32, args ...uintptr) WlProxy {
	callArgs := []uintptr{
		uintptr(proxy),
		uintptr(opcode),
		uintptr(unsafe.Pointer(iface)),
		uintptr(version),
	}
	callArgs = append(callArgs, args...)
	r, _, _ := purego.SyscallN(fnWlProxyMarshalConstructorVersioned, callArgs...)
	return WlProxy(r)
}

// wlProxyMarshal sends a request on a proxy (fire-and-forget, no return object).
func wlProxyMarshal(proxy WlProxy, opcode uint32, args ...uintptr) {
	callArgs := []uintptr{uintptr(proxy), uintptr(opcode)}
	callArgs = append(callArgs, args...)
	purego.SyscallN(fnWlProxyMarshal, callArgs...)
}

// ---- Wayland listener structs ----
// Each listener is a struct of function pointers (C calling convention).
// The layout must exactly match the Wayland protocol event signatures.

// WlRegistryListener handles wl_registry events.
type WlRegistryListener struct {
	// global(void *data, wl_registry*, uint32_t name, const char *interface, uint32_t version)
	Global uintptr
	// global_remove(void *data, wl_registry*, uint32_t name)
	GlobalRemove uintptr
}

// WlSeatListener handles wl_seat events.
type WlSeatListener struct {
	// capabilities(void *data, wl_seat*, uint32_t caps)
	Capabilities uintptr
	// name(void *data, wl_seat*, const char *name)
	Name uintptr
}

// WlPointerListener handles wl_pointer events.
type WlPointerListener struct {
	Enter  uintptr // (data, wl_pointer*, serial, surface, sx, sy)
	Leave  uintptr // (data, wl_pointer*, serial, surface)
	Motion uintptr // (data, wl_pointer*, time, sx, sy)
	Button uintptr // (data, wl_pointer*, serial, time, button, state)
	Axis   uintptr // (data, wl_pointer*, time, axis, value)
	Frame  uintptr
	AxisSource  uintptr
	AxisStop    uintptr
	AxisDiscrete uintptr
}

// WlKeyboardListener handles wl_keyboard events.
type WlKeyboardListener struct {
	Keymap    uintptr // (data, wl_keyboard*, format, fd, size)
	Enter     uintptr // (data, wl_keyboard*, serial, surface, keys)
	Leave     uintptr // (data, wl_keyboard*, serial, surface)
	Key       uintptr // (data, wl_keyboard*, serial, time, key, state)
	Modifiers uintptr // (data, wl_keyboard*, serial, mods_depressed, mods_latched, mods_locked, group)
	RepeatInfo uintptr // (data, wl_keyboard*, rate, delay)
}

// XdgWmBaseListener handles xdg_wm_base events.
type XdgWmBaseListener struct {
	// ping(void *data, xdg_wm_base*, uint32_t serial)
	Ping uintptr
}

// XdgSurfaceListener handles xdg_surface events.
type XdgSurfaceListener struct {
	// configure(void *data, xdg_surface*, uint32_t serial)
	Configure uintptr
}

// XdgToplevelListener handles xdg_toplevel events.
type XdgToplevelListener struct {
	// configure(void *data, xdg_toplevel*, int32_t width, int32_t height, wl_array *states)
	Configure uintptr
	// close(void *data, xdg_toplevel*)
	Close uintptr
}

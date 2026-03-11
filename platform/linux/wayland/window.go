//go:build linux && !android

package wayland

import (
	"fmt"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
)

// Window implements platform.Window for Linux/Wayland.
type Window struct {
	p       *Platform
	surface WlSurface
	xdgSurface  XdgSurface
	xdgToplevel XdgToplevel

	width, height       int
	minWidth, minHeight int
	maxWidth, maxHeight int

	dpiScale        float32
	fullscreen      bool
	decorated       bool
	resizable       bool
	visible         bool
	deferredVisible bool
	shouldClose     bool

	// Pending configure state
	pendingWidth, pendingHeight int
	configured                  bool

	// XDG surface/toplevel listeners (kept alive by the struct)
	xdgSurfaceListener  XdgSurfaceListener
	xdgToplevelListener XdgToplevelListener

	// Position (not directly available in Wayland; tracked via configure)
	posX, posY int
}

// newWindow creates a new Wayland window.
func newWindow(p *Platform, opts platform.WindowOptions) (*Window, error) {
	w := &Window{
		p:         p,
		width:     opts.Width,
		height:    opts.Height,
		minWidth:  opts.MinWidth,
		minHeight: opts.MinHeight,
		maxWidth:  opts.MaxWidth,
		maxHeight: opts.MaxHeight,
		dpiScale:  1.0,
		decorated: opts.Decorated,
		resizable: opts.Resizable,
	}

	// Create wl_surface via wl_compositor.create_surface (opcode 0)
	w.surface = wlProxyMarshalConstructorVersioned(p.compositor, wlCompositorCreateSurface, &p.surfaceIface, 1)
	if w.surface == 0 {
		return nil, fmt.Errorf("wayland: wl_compositor_create_surface failed")
	}

	// Create xdg_surface via xdg_wm_base.get_xdg_surface (opcode 2)
	w.xdgSurface = wlProxyMarshalConstructorVersioned(p.xdgWmBase, xdgWmBaseGetXdgSurface, &p.xdgSurfaceIface, 1,
		uintptr(w.surface),
	)
	if w.xdgSurface == 0 {
		wlProxyDestroy(w.surface)
		return nil, fmt.Errorf("wayland: xdg_wm_base_get_xdg_surface failed")
	}

	// Attach xdg_surface configure listener
	w.xdgSurfaceListener = XdgSurfaceListener{
		Configure: purego.NewCallback(w.cbXdgSurfaceConfigure),
	}
	wlProxyAddListener(w.xdgSurface, unsafe.Pointer(&w.xdgSurfaceListener), unsafe.Pointer(w))

	// Create xdg_toplevel via xdg_surface.get_toplevel (opcode 1)
	w.xdgToplevel = wlProxyMarshalConstructorVersioned(w.xdgSurface, xdgSurfaceGetToplevel, &p.xdgToplevelIface, 1)
	if w.xdgToplevel == 0 {
		wlProxyDestroy(w.xdgSurface)
		wlProxyDestroy(w.surface)
		return nil, fmt.Errorf("wayland: xdg_surface_get_toplevel failed")
	}

	// Attach xdg_toplevel listener
	w.xdgToplevelListener = XdgToplevelListener{
		Configure: purego.NewCallback(w.cbXdgToplevelConfigure),
		Close:     purego.NewCallback(w.cbXdgToplevelClose),
	}
	wlProxyAddListener(w.xdgToplevel, unsafe.Pointer(&w.xdgToplevelListener), unsafe.Pointer(w))

	// Set window title
	if opts.Title != "" {
		w.SetTitle(opts.Title)
	}

	// Commit to make the window visible to the compositor (triggers initial configure)
	wlProxyMarshal(w.surface, wlSurfaceCommit)
	wlDisplayRoundtrip(p.dpy)

	if opts.Visible {
		w.deferredVisible = true
	}

	if opts.Fullscreen {
		w.SetFullscreen(true)
	}

	return w, nil
}

// ---- XDG surface/toplevel callbacks (all-uintptr for syscall.NewCallback) ----

// cbXdgSurfaceConfigure: void(*)(void *data, xdg_surface*, uint32_t serial)
func (w *Window) cbXdgSurfaceConfigure(data, xdgSurface, serial uintptr) uintptr {
	// Apply pending size if compositor gave us one
	if w.pendingWidth > 0 && w.pendingHeight > 0 {
		if w.width != w.pendingWidth || w.height != w.pendingHeight {
			w.width = w.pendingWidth
			w.height = w.pendingHeight
			w.p.pushEvent(event.Event{
				Type:         event.WindowResize,
				WindowWidth:  w.width,
				WindowHeight: w.height,
			})
		}
	}
	wlProxyMarshal(XdgSurface(xdgSurface), xdgSurfaceAckConfigure, serial)
	wlProxyMarshal(w.surface, wlSurfaceCommit)
	w.configured = true
	return 0
}

// cbXdgToplevelConfigure: void(*)(void *data, xdg_toplevel*, int32_t width, int32_t height, wl_array *states)
func (w *Window) cbXdgToplevelConfigure(data, toplevel, width, height, states uintptr) uintptr {
	if int32(width) > 0 {
		w.pendingWidth = int(int32(width))
	}
	if int32(height) > 0 {
		w.pendingHeight = int(int32(height))
	}
	return 0
}

// cbXdgToplevelClose: void(*)(void *data, xdg_toplevel*)
func (w *Window) cbXdgToplevelClose(data, toplevel uintptr) uintptr {
	w.shouldClose = true
	w.p.pushEvent(event.Event{Type: event.WindowClose})
	return 0
}

// ---- platform.Window interface implementation ----

func (w *Window) Size() (int, int) {
	return w.width, w.height
}

func (w *Window) SetSize(width, height int) {
	// Wayland doesn't allow client-side resize requests; size is compositor-driven.
	// We store the desired size and apply it via the next commit.
	w.width = width
	w.height = height
}

func (w *Window) FramebufferSize() (int, int) {
	return w.width, w.height
}

func (w *Window) Position() (int, int) {
	// Wayland doesn't expose absolute window positions.
	return 0, 0
}

func (w *Window) SetPosition(x, y int) {
	// Wayland doesn't support client-requested window positioning.
}

func (w *Window) SetTitle(title string) {
	b := append([]byte(title), 0)
	wlProxyMarshal(w.xdgToplevel, xdgToplevelSetTitle, uintptr(unsafe.Pointer(&b[0])))
}

func (w *Window) SetFullscreen(fullscreen bool) {
	if w.fullscreen == fullscreen {
		return
	}
	if fullscreen {
		// xdg_toplevel.set_fullscreen(output=NULL)
		wlProxyMarshal(w.xdgToplevel, xdgToplevelSetFullscreen, 0)
	} else {
		wlProxyMarshal(w.xdgToplevel, xdgToplevelUnsetFullscreen)
	}
	w.fullscreen = fullscreen
	wlProxyMarshal(w.surface, wlSurfaceCommit)
	wlDisplayFlush(w.p.dpy)
}

func (w *Window) IsFullscreen() bool {
	return w.fullscreen
}

func (w *Window) ShouldClose() bool {
	return w.shouldClose
}

func (w *Window) SetShouldClose(close bool) {
	w.shouldClose = close
}

// NativeHandle returns the wl_surface pointer as a uintptr.
// This is used by Vulkan VK_KHR_wayland_surface.
func (w *Window) NativeHandle() uintptr {
	return uintptr(w.surface)
}

func (w *Window) DPIScale() float32 {
	return w.dpiScale
}

func (w *Window) SetVisible(visible bool) {
	w.visible = visible
	// In Wayland, visibility is controlled by attaching/detaching the buffer.
	// Commit with no buffer to hide; the compositor will stop rendering.
	wlProxyMarshal(w.surface, wlSurfaceCommit)
	wlDisplayFlush(w.p.dpy)
}

func (w *Window) ShowDeferred() {
	if w.deferredVisible {
		w.deferredVisible = false
		w.SetVisible(true)
	}
}

func (w *Window) SetMinSize(width, height int) {
	w.minWidth = width
	w.minHeight = height
	wlProxyMarshal(w.xdgToplevel, xdgToplevelSetMinSize,
		uintptr(int32(width)), uintptr(int32(height)),
	)
}

func (w *Window) SetMaxSize(width, height int) {
	w.maxWidth = width
	w.maxHeight = height
	wlProxyMarshal(w.xdgToplevel, xdgToplevelSetMaxSize,
		uintptr(int32(width)), uintptr(int32(height)),
	)
}

func (w *Window) SetCursor(cursor platform.CursorShape) {
	// Cursor setting on Wayland requires wl_pointer.set_cursor + wl_cursor_theme.
	// Stub: no-op for now.
	_ = cursor
}

func (w *Window) SetIMEPosition(caretRect uimath.Rect) {
	// TODO: implement via zwp_text_input_v3 protocol
	_ = caretRect
}

func (w *Window) ShowContextMenu(clientX, clientY int, items []platform.ContextMenuItem) int {
	// Wayland doesn't have native context menus; requires custom popup surface.
	_ = clientX
	_ = clientY
	_ = items
	return -1
}

func (w *Window) ClientToScreen(x, y int) (int, int) {
	// Wayland doesn't expose absolute screen coordinates.
	return x, y
}

func (w *Window) Destroy() {
	if w.xdgToplevel != 0 {
		wlProxyDestroy(w.xdgToplevel)
		w.xdgToplevel = 0
	}
	if w.xdgSurface != 0 {
		wlProxyDestroy(w.xdgSurface)
		w.xdgSurface = 0
	}
	if w.surface != 0 {
		wlProxyDestroy(w.surface)
		w.surface = 0
	}
}

// Compile-time interface check.
var _ platform.Window = (*Window)(nil)

//go:build android

package android

import (
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
)

// Window implements platform.Window for Android.
// The nativeWindow field holds an ANativeWindow* provided by the Android runtime
// via ANativeActivity_onCreate → ANativeActivity.window.
type Window struct {
	p      *Platform
	handle uintptr // ANativeWindow* (set by JNI bridge)

	width, height       int
	minWidth, minHeight int
	maxWidth, maxHeight int

	dpiScale    float32
	fullscreen  bool
	visible     bool
	shouldClose bool
}

// SetNativeWindow sets the ANativeWindow* handle from the JNI bridge.
// This must be called before the Vulkan backend initializes its surface.
func (w *Window) SetNativeWindow(handle uintptr) {
	w.handle = handle
}

// ---- platform.Window interface implementation ----

func (w *Window) Size() (int, int) {
	return w.width, w.height
}

func (w *Window) SetSize(width, height int) {
	// Android window size is OS-controlled; store for future reference.
	w.width = width
	w.height = height
}

func (w *Window) FramebufferSize() (int, int) {
	return w.width, w.height
}

func (w *Window) Position() (int, int) {
	// Android doesn't expose window positions.
	return 0, 0
}

func (w *Window) SetPosition(x, y int) {
	// Not applicable on Android.
}

func (w *Window) SetTitle(title string) {
	// Android window titles are set via the Activity intent; stub.
}

func (w *Window) SetFullscreen(fullscreen bool) {
	// Full-screen on Android is managed by the system UI visibility flags.
	// Requires JNI to call View.setSystemUiVisibility.
	w.fullscreen = fullscreen
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

// NativeHandle returns the ANativeWindow* pointer for Vulkan surface creation.
func (w *Window) NativeHandle() uintptr {
	return w.handle
}

func (w *Window) DPIScale() float32 {
	if w.dpiScale == 0 {
		return 1.0
	}
	return w.dpiScale
}

// SetDPIScale allows the JNI bridge to set the DPI scale factor.
func (w *Window) SetDPIScale(scale float32) {
	w.dpiScale = scale
}

func (w *Window) SetVisible(visible bool) {
	w.visible = visible
}

func (w *Window) ShowDeferred() {
	// On Android, window visibility is system-managed.
}

func (w *Window) SetMinSize(width, height int) {
	w.minWidth = width
	w.minHeight = height
}

func (w *Window) SetMaxSize(width, height int) {
	w.maxWidth = width
	w.maxHeight = height
}

func (w *Window) SetCursor(cursor platform.CursorShape) {
	// Cursor shapes require JNI integration with PointerIcon API (Android 7.0+).
	_ = cursor
}

func (w *Window) SetIMEPosition(caretRect uimath.Rect) {
	// Soft keyboard positioning requires JNI; stub.
	_ = caretRect
}

func (w *Window) ShowContextMenu(clientX, clientY int, items []platform.ContextMenuItem) int {
	// Context menus on Android require JNI PopupMenu.
	_ = clientX
	_ = clientY
	_ = items
	return -1
}

func (w *Window) ClientToScreen(x, y int) (int, int) {
	return x, y
}

func (w *Window) Destroy() {
	w.handle = 0
}

// Compile-time interface check.
var _ platform.Window = (*Window)(nil)

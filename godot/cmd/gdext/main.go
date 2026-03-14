// Command gdext builds a GDExtension shared library for Godot integration.
//
// Build:
//   CGO_ENABLED=1 go build -buildmode=c-shared -o goui.dll ./godot/cmd/gdext
//   CGO_ENABLED=1 go build -buildmode=c-shared -o goui.so  ./godot/cmd/gdext   (Linux)
//   CGO_ENABLED=1 go build -buildmode=c-shared -o goui.dylib ./godot/cmd/gdext (macOS)
//
// Install the .dll/.so/.dylib + .gdextension file in your Godot project.
// See godot/doc.go for GDScript usage examples.
package main

import "C"

import (
	"unsafe"

	"github.com/kasuganosora/ui/godot"
)

//export goui_create
func goui_create(width, height C.int, dpiScale C.float) C.longlong {
	id, err := godot.CreateInstance(int(width), int(height), float32(dpiScale))
	if err != nil {
		return 0
	}
	return C.longlong(id)
}

//export goui_destroy
func goui_destroy(id C.longlong) {
	godot.DestroyInstance(int64(id))
}

//export goui_frame
func goui_frame(id C.longlong, dt C.float) {
	godot.InstanceFrame(int64(id), float32(dt))
}

//export goui_resize
func goui_resize(id C.longlong, width, height C.int) {
	godot.InstanceResize(int64(id), int(width), int(height))
}

//export goui_pixels
func goui_pixels(id C.longlong, outLen *C.int) *C.char {
	px := godot.InstancePixels(int64(id))
	if len(px) == 0 {
		*outLen = 0
		return nil
	}
	*outLen = C.int(len(px))
	return (*C.char)(unsafe.Pointer(&px[0]))
}

//export goui_framebuffer_size
func goui_framebuffer_size(id C.longlong, outW, outH *C.int) {
	w, h := godot.InstanceFramebufferSize(int64(id))
	*outW = C.int(w)
	*outH = C.int(h)
}

//export goui_inject_mouse_move
func goui_inject_mouse_move(id C.longlong, x, y C.float) {
	godot.InstanceInjectMouseMove(int64(id), float32(x), float32(y))
}

//export goui_inject_mouse_click
func goui_inject_mouse_click(id C.longlong, x, y C.float, button C.int) {
	godot.InstanceInjectMouseClick(int64(id), float32(x), float32(y), int(button))
}

//export goui_inject_scroll
func goui_inject_scroll(id C.longlong, x, y, dx, dy C.float) {
	godot.InstanceInjectScroll(int64(id), float32(x), float32(y), float32(dx), float32(dy))
}

//export goui_inject_key_down
func goui_inject_key_down(id C.longlong, key, modifiers C.int) {
	godot.InstanceInjectKeyDown(int64(id), int(key), int(modifiers))
}

//export goui_inject_key_up
func goui_inject_key_up(id C.longlong, key, modifiers C.int) {
	godot.InstanceInjectKeyUp(int64(id), int(key), int(modifiers))
}

//export goui_inject_char
func goui_inject_char(id C.longlong, ch C.int) {
	godot.InstanceInjectChar(int64(id), rune(ch))
}

func main() {}

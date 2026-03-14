package godot

// This file defines the C-callable API for GDExtension integration.
// Build as a shared library with:
//   CGO_ENABLED=1 go build -buildmode=c-shared -o goui.dll ./godot/cmd/gdext
//
// The exported functions follow the GDExtension C ABI and can be called
// from Godot's GDScript, C#, or any GDExtension-compatible language.
//
// For the actual CGO exports, see cmd/gdext/main.go which wraps these
// functions with //export directives.

// --- Singleton manager for GDExtension callbacks ---

import (
	"sync"

	"github.com/kasuganosora/ui/event"
)

var (
	instanceMu sync.Mutex
	instances  = make(map[int64]*UI)
	nextInstID int64 = 1
)

// CreateInstance creates a new UI instance and returns its handle.
func CreateInstance(width, height int, dpiScale float32) (int64, error) {
	ui, err := NewUI(UIOptions{
		Width:    width,
		Height:   height,
		DPIScale: dpiScale,
	})
	if err != nil {
		return 0, err
	}

	instanceMu.Lock()
	id := nextInstID
	nextInstID++
	instances[id] = ui
	instanceMu.Unlock()

	return id, nil
}

// GetInstance returns a UI instance by handle. Returns nil if not found.
func GetInstance(id int64) *UI {
	instanceMu.Lock()
	defer instanceMu.Unlock()
	return instances[id]
}

// DestroyInstance destroys a UI instance by handle.
func DestroyInstance(id int64) {
	instanceMu.Lock()
	ui := instances[id]
	delete(instances, id)
	instanceMu.Unlock()

	if ui != nil {
		ui.Destroy()
	}
}

// InstanceFrame processes one frame for the given instance.
func InstanceFrame(id int64, dt float32) {
	if ui := GetInstance(id); ui != nil {
		ui.Frame(dt)
	}
}

// InstancePixels returns the RGBA pixel data for the given instance.
func InstancePixels(id int64) []byte {
	if ui := GetInstance(id); ui != nil {
		return ui.Pixels()
	}
	return nil
}

// InstanceResize resizes the viewport of the given instance.
func InstanceResize(id int64, width, height int) {
	if ui := GetInstance(id); ui != nil {
		ui.Resize(width, height)
	}
}

// InstanceInjectMouseMove injects a mouse move event.
func InstanceInjectMouseMove(id int64, x, y float32) {
	if ui := GetInstance(id); ui != nil {
		ui.InjectMouseMove(x, y)
	}
}

// InstanceInjectMouseClick injects a mouse click event.
func InstanceInjectMouseClick(id int64, x, y float32, button int) {
	if ui := GetInstance(id); ui != nil {
		ui.InjectMouseClick(x, y, event.MouseButton(button))
	}
}

// InstanceInjectScroll injects a scroll event.
func InstanceInjectScroll(id int64, x, y, dx, dy float32) {
	if ui := GetInstance(id); ui != nil {
		ui.InjectScroll(x, y, dx, dy)
	}
}

// Modifier bitmask constants for the C API.
const (
	ModCtrl  = 1 << 0
	ModShift = 1 << 1
	ModAlt   = 1 << 2
	ModSuper = 1 << 3
)

func modifiersFromBitmask(mask int) event.Modifiers {
	return event.Modifiers{
		Ctrl:  mask&ModCtrl != 0,
		Shift: mask&ModShift != 0,
		Alt:   mask&ModAlt != 0,
		Super: mask&ModSuper != 0,
	}
}

// InstanceInjectKeyDown injects a key press event.
// modifiers is a bitmask: ModCtrl=1, ModShift=2, ModAlt=4, ModSuper=8.
func InstanceInjectKeyDown(id int64, key int, modifiers int) {
	if ui := GetInstance(id); ui != nil {
		ui.InjectKeyDown(event.Key(key), modifiersFromBitmask(modifiers))
	}
}

// InstanceInjectKeyUp injects a key release event.
func InstanceInjectKeyUp(id int64, key int, modifiers int) {
	if ui := GetInstance(id); ui != nil {
		ui.InjectKeyUp(event.Key(key), modifiersFromBitmask(modifiers))
	}
}

// InstanceInjectChar injects a character input event.
func InstanceInjectChar(id int64, ch rune) {
	if ui := GetInstance(id); ui != nil {
		ui.InjectChar(ch)
	}
}

// InstanceFramebufferSize returns the framebuffer dimensions.
func InstanceFramebufferSize(id int64) (int, int) {
	if ui := GetInstance(id); ui != nil {
		return ui.FramebufferSize()
	}
	return 0, 0
}

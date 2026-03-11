//go:build linux && !android

package wayland

import (
	"fmt"
	"os"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/platform"
)

// Platform implements platform.Platform for Linux/Wayland.
type Platform struct {
	dpy     WlDisplay
	windows []*Window
	events  []event.Event
	mu      sync.Mutex
	inited  bool

	// Wayland globals obtained from the registry
	compositor WlCompositor
	seat       WlSeat
	xdgWmBase  XdgWmBase

	// Interface descriptors for wl_proxy_marshal_constructor_versioned.
	// These hold C-string name pointers for the binding protocol.
	compositorIface WlInterface
	seatIface       WlInterface
	xdgWmBaseIface  WlInterface
	surfaceIface    WlInterface
	xdgSurfaceIface WlInterface
	xdgToplevelIface WlInterface

	// Compositor-reported seat capabilities
	hasPointer  bool
	hasKeyboard bool

	// Input devices
	pointer  WlPointer
	keyboard WlKeyboard

	// Registered callback structs (kept alive to prevent GC collection)
	registryListener  WlRegistryListener
	seatListener      WlSeatListener
	pointerListener   WlPointerListener
	keyboardListener  WlKeyboardListener
	xdgWmBaseListener XdgWmBaseListener

	// Pointer state
	pointerX, pointerY float32
	pointerWindow       *Window

	// Keyboard modifier state
	keyMods event.Modifiers
}

// New creates a new Wayland platform instance.
func New() *Platform {
	return &Platform{}
}

// Init implements platform.Platform.
func (p *Platform) Init() error {
	if p.inited {
		return nil
	}

	p.dpy = wlDisplayConnect(nil)
	if p.dpy == 0 {
		return fmt.Errorf("wayland: wl_display_connect failed — is $WAYLAND_DISPLAY set?")
	}

	// Initialize WlInterface name pointers.
	// wl_proxy_marshal_constructor_versioned uses the interface name for binding.
	p.compositorIface.Name = &ifaceNameCompositor[0]
	p.seatIface.Name = &ifaceNameSeat[0]
	p.xdgWmBaseIface.Name = &ifaceNameXdgWmBase[0]

	// Construct registry listener with all-uintptr callbacks (required by syscall.NewCallback)
	p.registryListener = WlRegistryListener{
		Global:       purego.NewCallback(p.cbRegistryGlobal),
		GlobalRemove: purego.NewCallback(p.cbRegistryGlobalRemove),
	}

	// Get the registry object (wl_display opcode 1 = get_registry)
	registry := wlProxyMarshalConstructorVersioned(p.dpy, 1, &p.compositorIface, 1)
	if registry == 0 {
		wlDisplayDisconnect(p.dpy)
		return fmt.Errorf("wayland: failed to get wl_registry")
	}
	defer wlProxyDestroy(registry)

	wlProxyAddListener(registry, unsafe.Pointer(&p.registryListener), unsafe.Pointer(p))

	// Block until all registry globals are announced
	wlDisplayRoundtrip(p.dpy)

	if p.compositor == 0 {
		wlDisplayDisconnect(p.dpy)
		return fmt.Errorf("wayland: wl_compositor not available")
	}
	if p.xdgWmBase == 0 {
		wlDisplayDisconnect(p.dpy)
		return fmt.Errorf("wayland: xdg_wm_base not available")
	}

	// Attach xdg_wm_base ping listener
	p.xdgWmBaseListener = XdgWmBaseListener{
		Ping: purego.NewCallback(p.cbXdgWmBasePing),
	}
	wlProxyAddListener(p.xdgWmBase, unsafe.Pointer(&p.xdgWmBaseListener), unsafe.Pointer(p))

	// Attach seat listener if available
	if p.seat != 0 {
		p.seatListener = WlSeatListener{
			Capabilities: purego.NewCallback(p.cbSeatCapabilities),
			Name:         purego.NewCallback(p.cbSeatName),
		}
		wlProxyAddListener(p.seat, unsafe.Pointer(&p.seatListener), unsafe.Pointer(p))
		wlDisplayRoundtrip(p.dpy)
	}

	p.inited = true
	return nil
}

// CreateWindow implements platform.Platform.
func (p *Platform) CreateWindow(opts platform.WindowOptions) (platform.Window, error) {
	if !p.inited {
		return nil, fmt.Errorf("wayland: platform not initialized")
	}
	w, err := newWindow(p, opts)
	if err != nil {
		return nil, err
	}
	p.windows = append(p.windows, w)
	return w, nil
}

// PollEvents implements platform.Platform.
func (p *Platform) PollEvents() []event.Event {
	for _, w := range p.windows {
		w.ShowDeferred()
	}
	wlDisplayDispatchPending(p.dpy)
	wlDisplayFlush(p.dpy)

	p.mu.Lock()
	evs := make([]event.Event, len(p.events))
	copy(evs, p.events)
	p.events = p.events[:0]
	p.mu.Unlock()
	return evs
}

// ProcessMessages implements platform.Platform.
func (p *Platform) ProcessMessages() { wlDisplayFlush(p.dpy) }

func (p *Platform) GetClipboardText() string        { return "" }
func (p *Platform) SetClipboardText(text string)    {}
func (p *Platform) GetPrimaryMonitorDPI() float32   { return 96.0 }

func (p *Platform) GetSystemLocale() string {
	if lang := os.Getenv("LANG"); lang != "" {
		return lang
	}
	return "en_US.UTF-8"
}

// Terminate implements platform.Platform.
func (p *Platform) Terminate() {
	for _, w := range p.windows {
		w.Destroy()
	}
	p.windows = nil
	if p.pointer != 0 {
		wlProxyDestroy(p.pointer)
		p.pointer = 0
	}
	if p.keyboard != 0 {
		wlProxyDestroy(p.keyboard)
		p.keyboard = 0
	}
	if p.seat != 0 {
		wlProxyDestroy(p.seat)
		p.seat = 0
	}
	if p.xdgWmBase != 0 {
		wlProxyDestroy(p.xdgWmBase)
		p.xdgWmBase = 0
	}
	if p.compositor != 0 {
		wlProxyDestroy(p.compositor)
		p.compositor = 0
	}
	if p.dpy != 0 {
		wlDisplayDisconnect(p.dpy)
		p.dpy = 0
	}
	p.inited = false
}

func (p *Platform) pushEvent(e event.Event) {
	p.mu.Lock()
	p.events = append(p.events, e)
	p.mu.Unlock()
}

func (p *Platform) findWindowBySurface(surface WlSurface) *Window {
	for _, w := range p.windows {
		if w.surface == surface {
			return w
		}
	}
	return nil
}

// ---- Registry callbacks (all-uintptr signatures for syscall.NewCallback) ----

// cbRegistryGlobal: void(*)(void *data, wl_registry*, uint32_t name, const char *iface, uint32_t version)
// syscall.NewCallback requires all args to be uintptr.
// uint32 arguments are passed as full uintptr; we cast them back with uint32().
func (p *Platform) cbRegistryGlobal(data, registry, name, iface, version uintptr) uintptr {
	ifaceName := cstringToString((*byte)(unsafe.Pointer(iface)))
	nameU32 := uint32(name)
	verU32 := uint32(version)
	switch ifaceName {
	case "wl_compositor":
		p.compositor = p.bindGlobal(WlRegistry(registry), nameU32, &p.compositorIface, verU32)
	case "wl_seat":
		p.seat = p.bindGlobal(WlRegistry(registry), nameU32, &p.seatIface, verU32)
	case "xdg_wm_base":
		v := verU32
		if v > 4 {
			v = 4
		}
		p.xdgWmBase = p.bindGlobal(WlRegistry(registry), nameU32, &p.xdgWmBaseIface, v)
	}
	return 0
}

// cbRegistryGlobalRemove: void(*)(void *data, wl_registry*, uint32_t name)
func (p *Platform) cbRegistryGlobalRemove(data, registry, name uintptr) uintptr {
	return 0
}

// bindGlobal sends wl_registry.bind to create a new global proxy.
func (p *Platform) bindGlobal(registry WlRegistry, name uint32, iface *WlInterface, version uint32) WlProxy {
	return wlProxyMarshalConstructorVersioned(registry, wlRegistryBind, iface, version,
		uintptr(name),
		uintptr(unsafe.Pointer(iface.Name)),
		uintptr(version),
	)
}

// ---- xdg_wm_base ping callback ----

// cbXdgWmBasePing: void(*)(void *data, xdg_wm_base*, uint32_t serial)
func (p *Platform) cbXdgWmBasePing(data, wmBase, serial uintptr) uintptr {
	wlProxyMarshal(XdgWmBase(wmBase), xdgWmBasePong, serial)
	return 0
}

// ---- wl_seat callbacks ----

const (
	wlSeatCapabilityPointer  = 1
	wlSeatCapabilityKeyboard = 2
	wlSeatCapabilityTouch    = 4
)

// cbSeatCapabilities: void(*)(void *data, wl_seat*, uint32_t caps)
func (p *Platform) cbSeatCapabilities(data, seat, caps uintptr) uintptr {
	capsU32 := uint32(caps)
	wantPointer := capsU32&wlSeatCapabilityPointer != 0
	wantKeyboard := capsU32&wlSeatCapabilityKeyboard != 0

	if wantPointer && p.pointer == 0 {
		p.pointer = wlProxyMarshalConstructorVersioned(WlSeat(seat), wlSeatGetPointer, &p.seatIface, 1)
		if p.pointer != 0 {
			p.pointerListener = WlPointerListener{
				Enter:  purego.NewCallback(p.cbPointerEnter),
				Leave:  purego.NewCallback(p.cbPointerLeave),
				Motion: purego.NewCallback(p.cbPointerMotion),
				Button: purego.NewCallback(p.cbPointerButton),
				Axis:   purego.NewCallback(p.cbPointerAxis),
			}
			wlProxyAddListener(p.pointer, unsafe.Pointer(&p.pointerListener), unsafe.Pointer(p))
		}
	} else if !wantPointer && p.pointer != 0 {
		wlProxyDestroy(p.pointer)
		p.pointer = 0
	}

	if wantKeyboard && p.keyboard == 0 {
		p.keyboard = wlProxyMarshalConstructorVersioned(WlSeat(seat), wlSeatGetKeyboard, &p.seatIface, 1)
		if p.keyboard != 0 {
			p.keyboardListener = WlKeyboardListener{
				Keymap:     purego.NewCallback(p.cbKeyboardKeymap),
				Enter:      purego.NewCallback(p.cbKeyboardEnter),
				Leave:      purego.NewCallback(p.cbKeyboardLeave),
				Key:        purego.NewCallback(p.cbKeyboardKey),
				Modifiers:  purego.NewCallback(p.cbKeyboardModifiers),
				RepeatInfo: purego.NewCallback(p.cbKeyboardRepeatInfo),
			}
			wlProxyAddListener(p.keyboard, unsafe.Pointer(&p.keyboardListener), unsafe.Pointer(p))
		}
	} else if !wantKeyboard && p.keyboard != 0 {
		wlProxyDestroy(p.keyboard)
		p.keyboard = 0
	}

	p.hasPointer = wantPointer
	p.hasKeyboard = wantKeyboard
	return 0
}

// cbSeatName: void(*)(void *data, wl_seat*, const char *name)
func (p *Platform) cbSeatName(data, seat, name uintptr) uintptr { return 0 }

// ---- wl_pointer callbacks ----

// cbPointerEnter: (void *data, wl_pointer*, uint32_t serial, wl_surface*, wl_fixed_t sx, wl_fixed_t sy)
func (p *Platform) cbPointerEnter(data, pointer, serial, surface, sx, sy uintptr) uintptr {
	p.pointerWindow = p.findWindowBySurface(WlSurface(surface))
	p.pointerX = wlFixed2Float(int32(sx))
	p.pointerY = wlFixed2Float(int32(sy))
	return 0
}

// cbPointerLeave: (void *data, wl_pointer*, uint32_t serial, wl_surface*)
func (p *Platform) cbPointerLeave(data, pointer, serial, surface uintptr) uintptr {
	p.pointerWindow = nil
	return 0
}

// cbPointerMotion: (void *data, wl_pointer*, uint32_t time, wl_fixed_t sx, wl_fixed_t sy)
func (p *Platform) cbPointerMotion(data, pointer, time, sx, sy uintptr) uintptr {
	p.pointerX = wlFixed2Float(int32(sx))
	p.pointerY = wlFixed2Float(int32(sy))
	p.pushEvent(event.Event{
		Type:    event.MouseMove,
		X:       p.pointerX,
		Y:       p.pointerY,
		GlobalX: p.pointerX,
		GlobalY: p.pointerY,
	})
	return 0
}

// cbPointerButton: (void *data, wl_pointer*, uint32_t serial, uint32_t time, uint32_t button, uint32_t state)
func (p *Platform) cbPointerButton(data, pointer, serial, time, button, state uintptr) uintptr {
	var btn event.MouseButton
	switch uint32(button) {
	case 0x110:
		btn = event.MouseButtonLeft
	case 0x111:
		btn = event.MouseButtonRight
	case 0x112:
		btn = event.MouseButtonMiddle
	case 0x113:
		btn = event.MouseButton4
	case 0x114:
		btn = event.MouseButton5
	default:
		btn = event.MouseButtonLeft
	}
	evType := event.MouseUp
	if uint32(state) == 1 {
		evType = event.MouseDown
	}
	p.pushEvent(event.Event{
		Type:    evType,
		Button:  btn,
		X:       p.pointerX,
		Y:       p.pointerY,
		GlobalX: p.pointerX,
		GlobalY: p.pointerY,
	})
	return 0
}

// cbPointerAxis: (void *data, wl_pointer*, uint32_t time, uint32_t axis, wl_fixed_t value)
func (p *Platform) cbPointerAxis(data, pointer, time, axis, value uintptr) uintptr {
	dx, dy := float32(0), float32(0)
	delta := wlFixed2Float(int32(value)) / 10.0
	switch uint32(axis) {
	case 0: // WL_POINTER_AXIS_VERTICAL_SCROLL
		dy = -delta
	case 1: // WL_POINTER_AXIS_HORIZONTAL_SCROLL
		dx = delta
	}
	p.pushEvent(event.Event{
		Type:    event.MouseWheel,
		WheelDX: dx,
		WheelDY: dy,
		GlobalX: p.pointerX,
		GlobalY: p.pointerY,
	})
	return 0
}

// ---- wl_keyboard callbacks ----

// cbKeyboardKeymap: (void *data, wl_keyboard*, uint32_t format, int32_t fd, uint32_t size)
func (p *Platform) cbKeyboardKeymap(data, keyboard, format, fd, size uintptr) uintptr {
	// TODO: parse XKB keymap from fd for layout-aware key translation
	return 0
}

// cbKeyboardEnter: (void *data, wl_keyboard*, uint32_t serial, wl_surface*, wl_array *keys)
func (p *Platform) cbKeyboardEnter(data, keyboard, serial, surface, keys uintptr) uintptr {
	p.pushEvent(event.Event{Type: event.WindowFocus})
	return 0
}

// cbKeyboardLeave: (void *data, wl_keyboard*, uint32_t serial, wl_surface*)
func (p *Platform) cbKeyboardLeave(data, keyboard, serial, surface uintptr) uintptr {
	p.pushEvent(event.Event{Type: event.WindowBlur})
	return 0
}

// cbKeyboardKey: (void *data, wl_keyboard*, uint32_t serial, uint32_t time, uint32_t key, uint32_t state)
func (p *Platform) cbKeyboardKey(data, keyboard, serial, time, key, state uintptr) uintptr {
	evKey := evdevToKey(uint32(key))
	evType := event.KeyUp
	if uint32(state) == 1 {
		evType = event.KeyDown
	}
	p.pushEvent(event.Event{
		Type:      evType,
		Key:       evKey,
		Modifiers: p.keyMods,
	})
	if uint32(state) == 1 {
		if ch := evdevToRune(uint32(key), p.keyMods.Shift); ch >= 32 && ch != 127 {
			p.pushEvent(event.Event{
				Type:      event.KeyPress,
				Char:      ch,
				Modifiers: p.keyMods,
			})
		}
	}
	return 0
}

// cbKeyboardModifiers: (void *data, wl_keyboard*, uint32_t serial, uint32_t mods_depressed,
//                       uint32_t mods_latched, uint32_t mods_locked, uint32_t group)
func (p *Platform) cbKeyboardModifiers(data, keyboard, serial, depressed, latched, locked, group uintptr) uintptr {
	active := uint32(depressed) | uint32(latched) | uint32(locked)
	p.keyMods = event.Modifiers{
		Shift: active&(1<<0) != 0,
		Ctrl:  active&(1<<2) != 0,
		Alt:   active&(1<<3) != 0,
		Super: active&(1<<6) != 0,
	}
	return 0
}

// cbKeyboardRepeatInfo: (void *data, wl_keyboard*, int32_t rate, int32_t delay)
func (p *Platform) cbKeyboardRepeatInfo(data, keyboard, rate, delay uintptr) uintptr { return 0 }

// ---- Helpers ----

// wlFixed2Float converts a Wayland wl_fixed_t (24.8 signed fixed-point) to float32.
func wlFixed2Float(f int32) float32 {
	return float32(f) / 256.0
}

// cstringToString converts a null-terminated C string pointer to a Go string.
func cstringToString(p *byte) string {
	if p == nil {
		return ""
	}
	ptr := uintptr(unsafe.Pointer(p))
	n := 0
	for *(*byte)(unsafe.Pointer(ptr + uintptr(n))) != 0 {
		n++
	}
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = *(*byte)(unsafe.Pointer(ptr + uintptr(i)))
	}
	return string(b)
}

// Compile-time interface check.
var _ platform.Platform = (*Platform)(nil)

//go:build linux && !android

package linux

import (
	"os"
	"testing"
	"unsafe"

	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/platform"
)

// TestNew verifies that New() returns a non-nil, usable Platform instance
// without requiring a display connection.
func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
	if p.inited {
		t.Error("New() should return an uninited platform")
	}
	if p.dpy != 0 {
		t.Error("New() should have zero display")
	}
}

// TestInitRequiresDisplay verifies that Init fails gracefully when DISPLAY is not set.
func TestInitRequiresDisplay(t *testing.T) {
	if os.Getenv("DISPLAY") != "" {
		t.Skip("DISPLAY is set; skipping no-display test")
	}
	p := New()
	err := p.Init()
	if err == nil {
		t.Error("Init() should fail when DISPLAY is not set")
		p.Terminate()
	}
}

// TestInitWithDisplay tests a full init/terminate cycle when a display is available.
func TestInitWithDisplay(t *testing.T) {
	if os.Getenv("DISPLAY") == "" {
		t.Skip("DISPLAY not set; skipping X11 display test")
	}
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer p.Terminate()

	if !p.inited {
		t.Error("Init() should mark platform as inited")
	}
	if p.dpy == 0 {
		t.Error("Init() should open a display connection")
	}
	if p.wmDeleteWindow == 0 {
		t.Error("Init() should intern WM_DELETE_WINDOW atom")
	}
}

// TestCreateWindowRequiresInit tests window creation without calling Init first.
func TestCreateWindowRequiresInit(t *testing.T) {
	p := New()
	_, err := p.CreateWindow(platform.WindowOptions{Title: "Test", Width: 800, Height: 600})
	if err == nil {
		t.Error("CreateWindow() should fail when platform is not initialized")
	}
}

// TestGetSystemLocale verifies that GetSystemLocale returns a non-empty string.
func TestGetSystemLocale(t *testing.T) {
	p := New()
	locale := p.GetSystemLocale()
	if locale == "" {
		t.Error("GetSystemLocale() should return a non-empty string")
	}
}

// TestGetPrimaryMonitorDPI verifies that GetPrimaryMonitorDPI returns a sane value.
func TestGetPrimaryMonitorDPI(t *testing.T) {
	p := New()
	dpi := p.GetPrimaryMonitorDPI()
	if dpi <= 0 {
		t.Errorf("GetPrimaryMonitorDPI() = %v; want > 0", dpi)
	}
}

// ---- keysymToKey tests ----

func TestKeysymToKey_BasicLetters(t *testing.T) {
	tests := []struct {
		keysym uint64
		want   event.Key
	}{
		{0x61, event.KeyA}, // 'a'
		{0x62, event.KeyB}, // 'b'
		{0x7A, event.KeyZ}, // 'z'
		{0x41, event.KeyA}, // 'A'
		{0x5A, event.KeyZ}, // 'Z'
	}
	for _, tt := range tests {
		got := keysymToKey(tt.keysym)
		if got != tt.want {
			t.Errorf("keysymToKey(0x%X) = %v; want %v", tt.keysym, got, tt.want)
		}
	}
}

func TestKeysymToKey_Digits(t *testing.T) {
	tests := []struct {
		keysym uint64
		want   event.Key
	}{
		{0x30, event.Key0},
		{0x31, event.Key1},
		{0x39, event.Key9},
	}
	for _, tt := range tests {
		got := keysymToKey(tt.keysym)
		if got != tt.want {
			t.Errorf("keysymToKey(0x%X) = %v; want %v", tt.keysym, got, tt.want)
		}
	}
}

func TestKeysymToKey_SpecialKeys(t *testing.T) {
	tests := []struct {
		keysym uint64
		want   event.Key
	}{
		{xkReturn, event.KeyEnter},
		{xkEscape, event.KeyEscape},
		{xkBackSpace, event.KeyBackspace},
		{xkTab, event.KeyTab},
		{xkDelete, event.KeyDelete},
		{xkHome, event.KeyHome},
		{xkEnd, event.KeyEnd},
		{xkLeft, event.KeyArrowLeft},
		{xkRight, event.KeyArrowRight},
		{xkUp, event.KeyArrowUp},
		{xkDown, event.KeyArrowDown},
		{xkPageUp, event.KeyPageUp},
		{xkPageDown, event.KeyPageDown},
		{xkInsert, event.KeyInsert},
		{xkF1, event.KeyF1},
		{xkF12, event.KeyF12},
		{xkShiftL, event.KeyLeftShift},
		{xkShiftR, event.KeyRightShift},
		{xkControlL, event.KeyLeftCtrl},
		{xkControlR, event.KeyRightCtrl},
		{xkAltL, event.KeyLeftAlt},
		{xkAltR, event.KeyRightAlt},
		{xkSuperL, event.KeyLeftSuper},
		{xkSuperR, event.KeyRightSuper},
		{xkCapsLock, event.KeyCapsLock},
		{xkSpace, event.KeySpace},
	}
	for _, tt := range tests {
		got := keysymToKey(tt.keysym)
		if got != tt.want {
			t.Errorf("keysymToKey(0x%X) = %v; want %v", tt.keysym, got, tt.want)
		}
	}
}

func TestKeysymToKey_Unknown(t *testing.T) {
	// A keysym with no mapping should return KeyUnknown.
	got := keysymToKey(0xFFFF00) // Not a known keysym
	if got != event.KeyUnknown {
		t.Errorf("keysymToKey(0xFFFF00) = %v; want KeyUnknown", got)
	}
}

func TestKeysymToKey_Numpad(t *testing.T) {
	tests := []struct {
		keysym uint64
		want   event.Key
	}{
		{uint64(xkKP0), event.KeyNumpad0},
		{uint64(xkKP9), event.KeyNumpad9},
		{uint64(xkKPAdd), event.KeyNumpadAdd},
		{uint64(xkKPSubtract), event.KeyNumpadSubtract},
		{uint64(xkKPMultiply), event.KeyNumpadMultiply},
		{uint64(xkKPDivide), event.KeyNumpadDivide},
	}
	for _, tt := range tests {
		got := keysymToKey(tt.keysym)
		if got != tt.want {
			t.Errorf("keysymToKey(0x%X) = %v; want %v", tt.keysym, got, tt.want)
		}
	}
}

// ---- modifiersFromState tests ----

func TestModifiersFromState_None(t *testing.T) {
	m := modifiersFromState(0)
	if m.Shift || m.Ctrl || m.Alt || m.Super {
		t.Errorf("modifiersFromState(0) should have no modifiers set, got %+v", m)
	}
}

func TestModifiersFromState_Shift(t *testing.T) {
	m := modifiersFromState(x11ShiftMask)
	if !m.Shift {
		t.Error("expected Shift to be set")
	}
	if m.Ctrl || m.Alt || m.Super {
		t.Error("expected only Shift to be set")
	}
}

func TestModifiersFromState_Ctrl(t *testing.T) {
	m := modifiersFromState(x11ControlMask)
	if !m.Ctrl {
		t.Error("expected Ctrl to be set")
	}
}

func TestModifiersFromState_Alt(t *testing.T) {
	m := modifiersFromState(x11Mod1Mask)
	if !m.Alt {
		t.Error("expected Alt to be set")
	}
}

func TestModifiersFromState_Super(t *testing.T) {
	m := modifiersFromState(x11Mod4Mask)
	if !m.Super {
		t.Error("expected Super to be set")
	}
}

func TestModifiersFromState_All(t *testing.T) {
	all := uint32(x11ShiftMask | x11ControlMask | x11Mod1Mask | x11Mod4Mask)
	m := modifiersFromState(all)
	if !m.Shift || !m.Ctrl || !m.Alt || !m.Super {
		t.Errorf("modifiersFromState(all) = %+v; want all set", m)
	}
}

// ---- XEvent type size/alignment tests ----

func TestXEventSize(t *testing.T) {
	// XEvent must be at least 192 bytes (24 * 8 bytes on 64-bit)
	size := unsafe.Sizeof(XEvent{})
	if size < 192 {
		t.Errorf("XEvent size = %d; want >= 192 bytes", size)
	}
}

func TestXKeyEventLayout(t *testing.T) {
	// XKeyEvent must be castable from XEvent (first struct field is Type at offset 0)
	var xe XEvent
	xe[0] = int64(KeyPress)
	kev := (*XKeyEvent)(unsafe.Pointer(&xe))
	if int64(kev.Type) != xe[0] {
		t.Errorf("XKeyEvent.Type = %v; want %v", kev.Type, xe[0])
	}
}

func TestXButtonEventLayout(t *testing.T) {
	var xe XEvent
	xe[0] = int64(ButtonPress)
	bev := (*XButtonEvent)(unsafe.Pointer(&xe))
	if int64(bev.Type) != xe[0] {
		t.Errorf("XButtonEvent.Type = %v; want %v", bev.Type, xe[0])
	}
}

// ---- Event type constant tests ----

func TestEventTypeConstants(t *testing.T) {
	tests := []struct {
		name string
		val  int
		want int
	}{
		{"KeyPress", KeyPress, 2},
		{"KeyRelease", KeyRelease, 3},
		{"ButtonPress", ButtonPress, 4},
		{"ButtonRelease", ButtonRelease, 5},
		{"MotionNotify", MotionNotify, 6},
		{"FocusIn", FocusIn, 9},
		{"FocusOut", FocusOut, 10},
		{"ConfigureNotify", ConfigureNotify, 22},
		{"ClientMessage", ClientMessage, 33},
	}
	for _, tt := range tests {
		if tt.val != tt.want {
			t.Errorf("%s = %d; want %d", tt.name, tt.val, tt.want)
		}
	}
}


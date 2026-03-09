//go:build darwin

package darwin

import (
	"testing"

	"github.com/kasuganosora/ui/event"
)

// ========== 纯逻辑测试 ==========

func TestConvertModifiers(t *testing.T) {
	skipIfNotDarwin(t)

	tests := []struct {
		name  string
		flags uint64
		want  event.Modifiers
	}{
		{
			name:  "no modifiers",
			flags: 0,
			want:  event.Modifiers{},
		},
		{
			name:  "shift only",
			flags: NSEventModifierFlagShift,
			want:  event.Modifiers{Shift: true},
		},
		{
			name:  "ctrl only",
			flags: NSEventModifierFlagControl,
			want:  event.Modifiers{Ctrl: true},
		},
		{
			name:  "alt (option) only",
			flags: NSEventModifierFlagOption,
			want:  event.Modifiers{Alt: true},
		},
		{
			name:  "super (command) only",
			flags: NSEventModifierFlagCommand,
			want:  event.Modifiers{Super: true},
		},
		{
			name:  "shift+ctrl",
			flags: NSEventModifierFlagShift | NSEventModifierFlagControl,
			want:  event.Modifiers{Shift: true, Ctrl: true},
		},
		{
			name:  "all modifiers",
			flags: NSEventModifierFlagShift | NSEventModifierFlagControl | 
			       NSEventModifierFlagOption | NSEventModifierFlagCommand,
			want:  event.Modifiers{Shift: true, Ctrl: true, Alt: true, Super: true},
		},
		{
			name:  "caps lock (ignored)",
			flags: NSEventModifierFlagCapsLock,
			want:  event.Modifiers{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertModifiers(tt.flags)
			if got != tt.want {
				t.Errorf("convertModifiers() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestTranslateKeyCodeLetters(t *testing.T) {
	skipIfNotDarwin(t)

	// Test A-Z keys (macOS keycodes 0x00-0x0C + others)
	tests := []struct {
		keyCode uint16
		want    event.Key
	}{
		{0x00, event.KeyA},
		{0x01, event.KeyS},
		{0x02, event.KeyD},
		{0x03, event.KeyF},
		{0x04, event.KeyH},
		{0x05, event.KeyG},
		{0x06, event.KeyZ},
		{0x07, event.KeyX},
		{0x08, event.KeyC},
		{0x09, event.KeyV},
		{0x0B, event.KeyB},
		{0x0C, event.KeyQ},
		{0x0D, event.KeyW},
		{0x0E, event.KeyE},
		{0x0F, event.KeyR},
		{0x10, event.KeyY},
		{0x11, event.KeyT},
		{0x1F, event.KeyO},
		{0x20, event.KeyU},
		{0x22, event.KeyI},
		{0x23, event.KeyP},
		{0x25, event.KeyL},
		{0x26, event.KeyJ},
		{0x28, event.KeyK},
		{0x2D, event.KeyN},
		{0x2E, event.KeyM},
	}

	for _, tt := range tests {
		got := translateKeyCode(tt.keyCode)
		if got != tt.want {
			t.Errorf("translateKeyCode(0x%02X) = %v, want %v", tt.keyCode, got, tt.want)
		}
	}
}

func TestTranslateKeyCodeNumbers(t *testing.T) {
	skipIfNotDarwin(t)

	tests := []struct {
		keyCode uint16
		want    event.Key
	}{
		{0x12, event.Key1},
		{0x13, event.Key2},
		{0x14, event.Key3},
		{0x15, event.Key4},
		{0x17, event.Key5},
		{0x16, event.Key6},
		{0x1A, event.Key7},
		{0x1C, event.Key8},
		{0x19, event.Key9},
		{0x1D, event.Key0},
	}

	for _, tt := range tests {
		got := translateKeyCode(tt.keyCode)
		if got != tt.want {
			t.Errorf("translateKeyCode(0x%02X) = %v, want %v", tt.keyCode, got, tt.want)
		}
	}
}

func TestTranslateKeyCodeFunctionKeys(t *testing.T) {
	skipIfNotDarwin(t)

	tests := []struct {
		keyCode uint16
		want    event.Key
	}{
		{0x7A, event.KeyF1},
		{0x78, event.KeyF2},
		{0x63, event.KeyF3},
		{0x76, event.KeyF4},
		{0x60, event.KeyF5},
		{0x61, event.KeyF6},
		{0x62, event.KeyF7},
		{0x64, event.KeyF8},
		{0x65, event.KeyF9},
		{0x6D, event.KeyF10},
		{0x67, event.KeyF11},
		{0x6E, event.KeyF12}, // Fixed: 0x6E not 0x6F
	}

	for _, tt := range tests {
		got := translateKeyCode(tt.keyCode)
		if got != tt.want {
			t.Errorf("translateKeyCode(0x%02X) = %v, want %v", tt.keyCode, got, tt.want)
		}
	}
}

func TestTranslateKeyCodeNumpad(t *testing.T) {
	skipIfNotDarwin(t)

	tests := []struct {
		keyCode uint16
		want    event.Key
	}{
		{0x52, event.KeyNumpad0},
		{0x53, event.KeyNumpad1},
		{0x54, event.KeyNumpad2},
		{0x55, event.KeyNumpad3},
		{0x56, event.KeyNumpad4},
		{0x57, event.KeyNumpad5},
		{0x58, event.KeyNumpad6},
		{0x59, event.KeyNumpad7},
		{0x5B, event.KeyNumpad8},
		{0x5C, event.KeyNumpad9},
		{0x3F, event.KeyNumpadMultiply},
		{0x45, event.KeyNumpadAdd},
		{0x4E, event.KeyNumpadSubtract},
		{0x4B, event.KeyNumpadDivide},
		{0x47, event.KeyNumLock}, // Clear key
	}

	for _, tt := range tests {
		got := translateKeyCode(tt.keyCode)
		if got != tt.want {
			t.Errorf("translateKeyCode(0x%02X) = %v, want %v", tt.keyCode, got, tt.want)
		}
	}
}

func TestTranslateKeyCodeSpecial(t *testing.T) {
	skipIfNotDarwin(t)

	tests := []struct {
		keyCode uint16
		want    event.Key
	}{
		{0x24, event.KeyEnter},
		{0x30, event.KeyTab},
		{0x31, event.KeySpace},
		{0x33, event.KeyBackspace},
		{0x34, event.KeyNumpadEnter},
		{0x35, event.KeyEscape},
		{0x37, event.KeyLeftSuper}, // Command
		{0x38, event.KeyLeftShift},
		{0x39, event.KeyCapsLock},
		{0x3A, event.KeyLeftAlt}, // Option
		{0x3B, event.KeyLeftCtrl},
		{0x3C, event.KeyRightShift},
		{0x3D, event.KeyRightAlt},
		{0x3E, event.KeyRightCtrl},
		{0x73, event.KeyHome},
		{0x77, event.KeyEnd},
		{0x74, event.KeyPageUp},
		{0x79, event.KeyPageDown},
		{0x75, event.KeyDelete},
		{0x72, event.KeyInsert},
		{0x7B, event.KeyArrowLeft},
		{0x7C, event.KeyArrowRight},
		{0x7D, event.KeyArrowDown},
		{0x7E, event.KeyArrowUp},
	}

	for _, tt := range tests {
		got := translateKeyCode(tt.keyCode)
		if got != tt.want {
			t.Errorf("translateKeyCode(0x%02X) = %v, want %v", tt.keyCode, got, tt.want)
		}
	}
}

func TestTranslateKeyCodePunctuation(t *testing.T) {
	skipIfNotDarwin(t)

	tests := []struct {
		keyCode uint16
		want    event.Key
	}{
		{0x18, event.KeyEqual},
		{0x1B, event.KeyMinus},
		{0x1E, event.KeyRightBracket},
		{0x21, event.KeyLeftBracket},
		{0x27, event.KeyApostrophe},
		{0x29, event.KeySemicolon},
		{0x2A, event.KeyBackslash},
		{0x2B, event.KeyComma},
		{0x2C, event.KeySlash},
		{0x2F, event.KeyPeriod},
		{0x32, event.KeyGraveAccent},
	}

	for _, tt := range tests {
		got := translateKeyCode(tt.keyCode)
		if got != tt.want {
			t.Errorf("translateKeyCode(0x%02X) = %v, want %v", tt.keyCode, got, tt.want)
		}
	}
}

func TestTranslateKeyCodeUnknown(t *testing.T) {
	skipIfNotDarwin(t)

	// Test unknown key codes return KeyUnknown
	unknownKeyCodes := []uint16{0xFF, 0xAA, 0xBB, 0xCC}
	for _, keyCode := range unknownKeyCodes {
		got := translateKeyCode(keyCode)
		if got != event.KeyUnknown {
			t.Errorf("translateKeyCode(0x%02X) = %v, want KeyUnknown", keyCode, got)
		}
	}
}

func TestPlatformNew(t *testing.T) {
	skipIfNotDarwin(t)

	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
	if p.inited {
		t.Error("New() platform should not be initialized")
	}
	if p.running {
		t.Error("New() platform should not be running")
	}
	if len(p.windows) != 0 {
		t.Error("New() platform should have no windows")
	}
	if len(p.events) != 0 {
		t.Error("New() platform should have no events")
	}
}

func TestPushEvent(t *testing.T) {
	skipIfNotDarwin(t)

	p := &Platform{}
	
	// Push one event
	evt1 := event.Event{Type: event.MouseDown}
	p.pushEvent(evt1)
	
	if len(p.events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(p.events))
	}
	if p.events[0].Type != event.MouseDown {
		t.Errorf("Expected MouseDown, got %v", p.events[0].Type)
	}
	
	// Push another event
	evt2 := event.Event{Type: event.KeyPress}
	p.pushEvent(evt2)
	
	if len(p.events) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(p.events))
	}
	if p.events[1].Type != event.KeyPress {
		t.Errorf("Expected KeyPress, got %v", p.events[1].Type)
	}
}

func TestCurrentTimestamp(t *testing.T) {
	skipIfNotDarwin(t)

	p := &Platform{
		timebase: 1000000, // 1 second
	}
	
	ts := p.currentTimestamp()
	// Should be a positive number (current time - timebase)
	if ts == 0 {
		t.Error("currentTimestamp() returned 0")
	}
}

// Benchmark tests
func BenchmarkConvertModifiers(b *testing.B) {
	flags := uint64(NSEventModifierFlagShift | NSEventModifierFlagControl | 
	                NSEventModifierFlagOption | NSEventModifierFlagCommand)
	
	for i := 0; i < b.N; i++ {
		_ = convertModifiers(flags)
	}
}

func BenchmarkTranslateKeyCode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = translateKeyCode(0x00) // KeyA
	}
}

func BenchmarkPushEvent(b *testing.B) {
	p := &Platform{}
	evt := event.Event{Type: event.MouseMove, X: 100, Y: 200}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.pushEvent(evt)
	}
}

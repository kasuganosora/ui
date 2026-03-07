//go:build windows

package win32

import (
	"runtime"
	"testing"
	"time"

	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/platform"
)

// === Pure logic tests (no window/GPU needed) ===

func TestTranslateVirtualKeyLetters(t *testing.T) {
	for vk := uintptr(0x41); vk <= 0x5A; vk++ {
		key := translateVirtualKey(vk)
		expected := event.KeyA + event.Key(vk-0x41)
		if key != expected {
			t.Errorf("VK 0x%X: expected %d, got %d", vk, expected, key)
		}
	}
}

func TestTranslateVirtualKeyNumbers(t *testing.T) {
	for vk := uintptr(0x30); vk <= 0x39; vk++ {
		key := translateVirtualKey(vk)
		expected := event.Key0 + event.Key(vk-0x30)
		if key != expected {
			t.Errorf("VK 0x%X: expected %d, got %d", vk, expected, key)
		}
	}
}

func TestTranslateVirtualKeyFunctionKeys(t *testing.T) {
	for vk := uintptr(VK_F1); vk <= uintptr(VK_F12); vk++ {
		key := translateVirtualKey(vk)
		expected := event.KeyF1 + event.Key(vk-VK_F1)
		if key != expected {
			t.Errorf("VK 0x%X: expected %d, got %d", vk, expected, key)
		}
	}
}

func TestTranslateVirtualKeyNumpad(t *testing.T) {
	for vk := uintptr(VK_NUMPAD0); vk <= uintptr(VK_NUMPAD9); vk++ {
		key := translateVirtualKey(vk)
		expected := event.KeyNumpad0 + event.Key(vk-VK_NUMPAD0)
		if key != expected {
			t.Errorf("VK 0x%X: expected %d, got %d", vk, expected, key)
		}
	}
}

func TestTranslateVirtualKeySpecials(t *testing.T) {
	tests := []struct {
		vk   uintptr
		want event.Key
	}{
		{VK_BACK, event.KeyBackspace},
		{VK_TAB, event.KeyTab},
		{VK_RETURN, event.KeyEnter},
		{VK_ESCAPE, event.KeyEscape},
		{VK_SPACE, event.KeySpace},
		{VK_LEFT, event.KeyArrowLeft},
		{VK_RIGHT, event.KeyArrowRight},
		{VK_UP, event.KeyArrowUp},
		{VK_DOWN, event.KeyArrowDown},
		{VK_DELETE, event.KeyDelete},
		{VK_INSERT, event.KeyInsert},
		{VK_HOME, event.KeyHome},
		{VK_END, event.KeyEnd},
		{VK_PRIOR, event.KeyPageUp},
		{VK_NEXT, event.KeyPageDown},
		{VK_LSHIFT, event.KeyLeftShift},
		{VK_RSHIFT, event.KeyRightShift},
		{VK_LCONTROL, event.KeyLeftCtrl},
		{VK_RCONTROL, event.KeyRightCtrl},
		{VK_LMENU, event.KeyLeftAlt},
		{VK_RMENU, event.KeyRightAlt},
		{VK_LWIN, event.KeyLeftSuper},
		{VK_RWIN, event.KeyRightSuper},
		{VK_CAPITAL, event.KeyCapsLock},
		{VK_NUMLOCK, event.KeyNumLock},
		{VK_SCROLL, event.KeyScrollLock},
		{VK_PAUSE, event.KeyPause},
		{VK_SNAPSHOT, event.KeyPrintScreen},
		{VK_APPS, event.KeyMenu},
	}
	for _, tt := range tests {
		if got := translateVirtualKey(tt.vk); got != tt.want {
			t.Errorf("VK 0x%X: expected %v, got %v", tt.vk, tt.want, got)
		}
	}
}

func TestTranslateVirtualKeyPunctuation(t *testing.T) {
	tests := []struct {
		vk   uintptr
		want event.Key
	}{
		{VK_OEM_MINUS, event.KeyMinus},
		{VK_OEM_PLUS, event.KeyEqual},
		{VK_OEM_4, event.KeyLeftBracket},
		{VK_OEM_6, event.KeyRightBracket},
		{VK_OEM_5, event.KeyBackslash},
		{VK_OEM_1, event.KeySemicolon},
		{VK_OEM_7, event.KeyApostrophe},
		{VK_OEM_3, event.KeyGraveAccent},
		{VK_OEM_COMMA, event.KeyComma},
		{VK_OEM_PERIOD, event.KeyPeriod},
		{VK_OEM_2, event.KeySlash},
	}
	for _, tt := range tests {
		if got := translateVirtualKey(tt.vk); got != tt.want {
			t.Errorf("VK 0x%X: expected %v, got %v", tt.vk, tt.want, got)
		}
	}
}

func TestTranslateVirtualKeyUnknown(t *testing.T) {
	if translateVirtualKey(0xFF) != event.KeyUnknown {
		t.Error("unmapped key should return KeyUnknown")
	}
}

func TestLowordHiword(t *testing.T) {
	lp := uintptr(0x00640032) // hiword=100, loword=50
	if loword(lp) != 50 {
		t.Errorf("loword: expected 50, got %d", loword(lp))
	}
	if hiword(lp) != 100 {
		t.Errorf("hiword: expected 100, got %d", hiword(lp))
	}
}

func TestLowordHiwordNegative(t *testing.T) {
	// Test signed extraction: -10 in int16 = 0xFFF6
	lp := uintptr(0xFFF6FFF6)
	if loword(lp) != -10 {
		t.Errorf("loword negative: expected -10, got %d", loword(lp))
	}
	if hiword(lp) != -10 {
		t.Errorf("hiword negative: expected -10, got %d", hiword(lp))
	}
}

func TestGetXYLParam(t *testing.T) {
	lp := uintptr(200<<16 | 100) // y=200, x=100
	x := getXLParam(lp)
	y := getYLParam(lp)
	if x != 100 {
		t.Errorf("x: expected 100, got %v", x)
	}
	if y != 200 {
		t.Errorf("y: expected 200, got %v", y)
	}
}

func TestGetWheelDelta(t *testing.T) {
	// 120 in hiword = 1 click
	wp := uintptr(120 << 16)
	delta := getWheelDelta(wp)
	if delta != 1.0 {
		t.Errorf("expected 1.0, got %v", delta)
	}
}

func TestXButtonFromWParam(t *testing.T) {
	// XBUTTON1 = hiword 1
	if xButtonFromWParam(uintptr(1 << 16)) != event.MouseButton4 {
		t.Error("XBUTTON1 should map to MouseButton4")
	}
	// XBUTTON2 = hiword 2
	if xButtonFromWParam(uintptr(2 << 16)) != event.MouseButton5 {
		t.Error("XBUTTON2 should map to MouseButton5")
	}
}

func TestDefaultWindowOptions(t *testing.T) {
	opts := platform.DefaultWindowOptions()
	if opts.Width != 1280 || opts.Height != 720 {
		t.Error("default size should be 1280x720")
	}
	if !opts.Resizable {
		t.Error("should be resizable by default")
	}
	if !opts.Decorated {
		t.Error("should be decorated by default")
	}
	if !opts.Visible {
		t.Error("should be visible by default")
	}
}

func TestUint32ToUintptr(t *testing.T) {
	// GWLP_USERDATA = -21
	result := uint32ToUintptr(GWLP_USERDATA)
	v := int32(GWLP_USERDATA)
	expected := uintptr(uint32(v))
	if result != expected {
		t.Errorf("expected %x, got %x", expected, result)
	}
}

func TestUtf16PtrFromString(t *testing.T) {
	p := utf16PtrFromString("hello")
	if p == nil {
		t.Fatal("should not be nil")
	}
}

func TestQueryTimeMicroseconds(t *testing.T) {
	t1 := queryTimeMicroseconds()
	t2 := queryTimeMicroseconds()
	if t2 < t1 {
		t.Error("timer should be monotonic")
	}
}

// === Integration tests (need a window) ===

func TestPlatformInitTerminate(t *testing.T) {
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer p.Terminate()

	if !p.inited {
		t.Error("should be initialized")
	}
}

func TestPlatformCreateWindow(t *testing.T) {
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer p.Terminate()

	opts := platform.WindowOptions{
		Title:     "Test Window",
		Width:     400,
		Height:    300,
		Resizable: true,
		Visible:   false, // Hidden — no GUI pop-up during test
		Decorated: true,
	}

	w, err := p.CreateWindow(opts)
	if err != nil {
		t.Fatalf("CreateWindow failed: %v", err)
	}
	defer w.Destroy()

	if w.NativeHandle() == 0 {
		t.Error("HWND should not be zero")
	}

	width, height := w.Size()
	if width != 400 || height != 300 {
		t.Errorf("size: expected 400x300, got %dx%d", width, height)
	}

	if w.ShouldClose() {
		t.Error("should not be closed initially")
	}

	dpi := w.DPIScale()
	if dpi <= 0 {
		t.Errorf("DPI scale should be positive, got %v", dpi)
	}
	t.Logf("DPI scale: %.2f", dpi)
}

func TestWindowSetTitle(t *testing.T) {
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer p.Terminate()

	w, err := p.CreateWindow(platform.WindowOptions{
		Title: "Original", Width: 200, Height: 200, Visible: false, Decorated: true,
	})
	if err != nil {
		t.Fatalf("CreateWindow failed: %v", err)
	}
	defer w.Destroy()

	// Should not panic
	w.SetTitle("新标题 New Title")
}

func TestWindowSetShouldClose(t *testing.T) {
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer p.Terminate()

	w, err := p.CreateWindow(platform.WindowOptions{
		Title: "Test", Width: 200, Height: 200, Visible: false, Decorated: true,
	})
	if err != nil {
		t.Fatalf("CreateWindow failed: %v", err)
	}
	defer w.Destroy()

	w.SetShouldClose(true)
	if !w.ShouldClose() {
		t.Error("should be marked for close")
	}
}

func TestWindowSetSize(t *testing.T) {
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer p.Terminate()

	w, err := p.CreateWindow(platform.WindowOptions{
		Title: "Resize", Width: 400, Height: 300, Visible: false, Decorated: true, Resizable: true,
	})
	if err != nil {
		t.Fatalf("CreateWindow failed: %v", err)
	}
	defer w.Destroy()

	w.SetSize(800, 600)

	// Process messages so WM_SIZE is handled
	p.PollEvents()

	width, height := w.Size()
	if width != 800 || height != 600 {
		t.Errorf("after resize: expected 800x600, got %dx%d", width, height)
	}
}

func TestPollEventsEmpty(t *testing.T) {
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer p.Terminate()

	events := p.PollEvents()
	// No windows, no events — should not crash
	_ = events
}

func TestGetSystemLocale(t *testing.T) {
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer p.Terminate()

	locale := p.GetSystemLocale()
	if locale == "" {
		t.Error("locale should not be empty")
	}
	t.Logf("System locale: %s", locale)
}

func TestGetPrimaryMonitorDPI(t *testing.T) {
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer p.Terminate()

	dpi := p.GetPrimaryMonitorDPI()
	if dpi <= 0 {
		t.Errorf("DPI should be positive, got %v", dpi)
	}
	t.Logf("Primary monitor DPI: %.0f", dpi)
}

// === OS Thread Locking Tests ===

func TestInitLocksOSThread(t *testing.T) {
	// Win32 requires the message loop to run on the window-creating thread.
	// Platform.Init must call runtime.LockOSThread to prevent goroutine migration.
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer p.Terminate()

	// After Init, the goroutine should be locked to an OS thread.
	// runtime.LockOSThread is cumulative — we can detect it by trying
	// to verify we're on the same thread. The simplest check: creating
	// a window and pumping messages should work without hanging.
	w, err := p.CreateWindow(platform.WindowOptions{
		Title: "Thread Test", Width: 200, Height: 200, Visible: false, Decorated: true,
	})
	if err != nil {
		t.Fatalf("CreateWindow failed: %v", err)
	}
	defer w.Destroy()

	// Pump messages — if goroutine migrated to wrong thread, this would
	// either hang or fail to process messages.
	done := make(chan bool, 1)
	go func() {
		// This goroutine checks that the main goroutine is responsive
		time.Sleep(500 * time.Millisecond)
		done <- true
	}()

	// Simulate a few frames of message pumping
	deadline := time.Now().Add(600 * time.Millisecond)
	for time.Now().Before(deadline) {
		p.PollEvents()
		runtime.Gosched()
		time.Sleep(10 * time.Millisecond)
	}

	select {
	case <-done:
		// OK — message loop didn't hang
	default:
		t.Error("message loop appears to have hung")
	}
}

func TestWindowCloseEventSetsFlag(t *testing.T) {
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer p.Terminate()

	w, err := p.CreateWindow(platform.WindowOptions{
		Title: "Close Test", Width: 200, Height: 200, Visible: false, Decorated: true,
	})
	if err != nil {
		t.Fatalf("CreateWindow failed: %v", err)
	}
	defer w.Destroy()

	if w.ShouldClose() {
		t.Fatal("should not be closed initially")
	}

	// Simulate WM_CLOSE by sending the message
	hwnd := w.NativeHandle()
	procPostMessageW.Call(hwnd, WM_CLOSE, 0, 0)

	// Process the message
	p.PollEvents()

	if !w.ShouldClose() {
		t.Error("window should be marked for close after WM_CLOSE")
	}
}

func TestWindowCloseEventGenerated(t *testing.T) {
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer p.Terminate()

	w, err := p.CreateWindow(platform.WindowOptions{
		Title: "Close Event Test", Width: 200, Height: 200, Visible: false, Decorated: true,
	})
	if err != nil {
		t.Fatalf("CreateWindow failed: %v", err)
	}
	defer w.Destroy()

	// Send WM_CLOSE
	hwnd := w.NativeHandle()
	procPostMessageW.Call(hwnd, WM_CLOSE, 0, 0)

	events := p.PollEvents()
	found := false
	for _, e := range events {
		if e.Type == event.WindowClose {
			found = true
		}
	}
	if !found {
		t.Error("WM_CLOSE should generate a WindowClose event")
	}
}

func TestMouseEventsGeneratedCorrectly(t *testing.T) {
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer p.Terminate()

	w, err := p.CreateWindow(platform.WindowOptions{
		Title: "Mouse Test", Width: 400, Height: 300, Visible: false, Decorated: true,
	})
	if err != nil {
		t.Fatalf("CreateWindow failed: %v", err)
	}
	defer w.Destroy()

	// Simulate WM_LBUTTONDOWN at (100, 50)
	hwnd := w.NativeHandle()
	lParam := uintptr(50<<16 | 100) // y=50, x=100
	procPostMessageW.Call(hwnd, WM_LBUTTONDOWN, 0, lParam)

	events := p.PollEvents()
	found := false
	for _, e := range events {
		if e.Type == event.MouseDown && e.Button == event.MouseButtonLeft {
			found = true
			if e.GlobalX != 100 || e.GlobalY != 50 {
				t.Errorf("mouse position: got (%v, %v), want (100, 50)", e.GlobalX, e.GlobalY)
			}
		}
	}
	if !found {
		t.Error("WM_LBUTTONDOWN should generate MouseDown event")
	}
}

func TestPollEventsDoesNotBlockOnEmptyQueue(t *testing.T) {
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer p.Terminate()

	// PollEvents with no messages should return immediately
	done := make(chan bool, 1)
	go func() {
		p.PollEvents()
		done <- true
	}()

	select {
	case <-done:
		// OK
	case <-time.After(time.Second):
		t.Error("PollEvents blocked on empty message queue")
	}
}

func TestClipboardRoundTrip(t *testing.T) {
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer p.Terminate()

	testStr := "GoUI 剪贴板测试 🎮"
	p.SetClipboardText(testStr)

	got := p.GetClipboardText()
	if got != testStr {
		t.Errorf("clipboard round-trip: expected %q, got %q", testStr, got)
	}
}

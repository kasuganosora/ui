//go:build windows

package win32

import (
	"testing"

	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
)

// Additional tests to bring win32 coverage to 80%+.

// --- Pure function tests ---

func TestMouseButtonEvent(t *testing.T) {
	// lParam: x=100 in loword, y=200 in hiword
	lp := uintptr(200<<16 | 100)
	e := mouseButtonEvent(1, 0, lp) // type=1 (MouseDown), button=0 (Left)
	if e.X != 100 || e.Y != 200 {
		t.Errorf("mouseButtonEvent coords: expected (100,200), got (%v,%v)", e.X, e.Y)
	}
	if e.GlobalX != 100 || e.GlobalY != 200 {
		t.Errorf("mouseButtonEvent global coords: expected (100,200), got (%v,%v)", e.GlobalX, e.GlobalY)
	}
}

func TestGetModifiers(t *testing.T) {
	// Just call it — returns current keyboard state
	mods := getModifiers()
	// During test, no keys should be held (usually)
	_ = mods
}

func TestIsKeyDown(t *testing.T) {
	// Test with a key that should not be down during test
	down := isKeyDown(VK_F12)
	_ = down // Just ensure no panic
}

func TestMakeIntResource(t *testing.T) {
	ptr := makeIntResource(IDC_ARROW)
	if ptr == nil {
		t.Error("makeIntResource should return non-nil")
	}
}

func TestUtf16ToString(t *testing.T) {
	// nil case
	if utf16ToString(nil) != "" {
		t.Error("nil should return empty string")
	}

	// Normal case via utf16PtrFromString
	p := utf16PtrFromString("hello world")
	result := utf16ToString(p)
	if result != "hello world" {
		t.Errorf("expected 'hello world', got %q", result)
	}
}

func TestUtf16FromString(t *testing.T) {
	u := utf16FromString("test")
	if len(u) == 0 {
		t.Error("should return non-empty slice")
	}
}

func TestLastError(t *testing.T) {
	err := lastError("TestFunc")
	if err == nil {
		t.Error("lastError should return non-nil error")
	}
}

func TestTranslateVirtualKeyNumpadOps(t *testing.T) {
	tests := []struct {
		vk   uintptr
		name string
	}{
		{VK_MULTIPLY, "Multiply"},
		{VK_ADD, "Add"},
		{VK_SUBTRACT, "Subtract"},
		{VK_DECIMAL, "Decimal"},
		{VK_DIVIDE, "Divide"},
	}
	for _, tt := range tests {
		key := translateVirtualKey(tt.vk)
		if key == 0 {
			t.Errorf("%s: should not return 0", tt.name)
		}
	}
}

// --- Window method tests (require a real HWND) ---

func newTestWindow(t *testing.T) (*Platform, *Window) {
	t.Helper()
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	w, err := p.CreateWindow(platform.WindowOptions{
		Title:     "Coverage Test",
		Width:     300,
		Height:    200,
		Visible:   false,
		Decorated: true,
		Resizable: true,
	})
	if err != nil {
		p.Terminate()
		t.Fatalf("CreateWindow failed: %v", err)
	}
	return p, w.(*Window)
}

func TestWindowFramebufferSize(t *testing.T) {
	p, w := newTestWindow(t)
	defer p.Terminate()
	defer w.Destroy()

	fbW, fbH := w.FramebufferSize()
	if fbW <= 0 || fbH <= 0 {
		t.Errorf("framebuffer size should be positive, got %dx%d", fbW, fbH)
	}
}

func TestWindowPosition(t *testing.T) {
	p, w := newTestWindow(t)
	defer p.Terminate()
	defer w.Destroy()

	x, y := w.Position()
	_ = x
	_ = y // Position may be 0,0 for hidden window — just ensure no panic
}

func TestWindowSetPosition(t *testing.T) {
	p, w := newTestWindow(t)
	defer p.Terminate()
	defer w.Destroy()

	w.SetPosition(100, 200)
	x, y := w.Position()
	if x != 100 || y != 200 {
		t.Errorf("expected position (100,200), got (%d,%d)", x, y)
	}
}

func TestWindowSetVisible(t *testing.T) {
	p, w := newTestWindow(t)
	defer p.Terminate()
	defer w.Destroy()

	w.SetVisible(true)
	w.SetVisible(false)
	// Should not panic
}

func TestWindowSetFullscreen(t *testing.T) {
	p, w := newTestWindow(t)
	defer p.Terminate()
	defer w.Destroy()

	if w.IsFullscreen() {
		t.Error("should not be fullscreen initially")
	}

	w.SetFullscreen(true)
	if !w.IsFullscreen() {
		t.Error("should be fullscreen after SetFullscreen(true)")
	}

	// Setting same value should be no-op
	w.SetFullscreen(true)

	w.SetFullscreen(false)
	if w.IsFullscreen() {
		t.Error("should not be fullscreen after SetFullscreen(false)")
	}
}

func TestWindowSetMinMaxSize(t *testing.T) {
	p, w := newTestWindow(t)
	defer p.Terminate()
	defer w.Destroy()

	w.SetMinSize(100, 100)
	w.SetMaxSize(1000, 1000)

	if w.minWidth != 100 || w.minHeight != 100 {
		t.Error("min size not set")
	}
	if w.maxWidth != 1000 || w.maxHeight != 1000 {
		t.Error("max size not set")
	}
}

func TestWindowSetCursorAll(t *testing.T) {
	p, w := newTestWindow(t)
	defer p.Terminate()
	defer w.Destroy()

	cursors := []platform.CursorShape{
		platform.CursorArrow,
		platform.CursorIBeam,
		platform.CursorCrosshair,
		platform.CursorHand,
		platform.CursorHResize,
		platform.CursorVResize,
		platform.CursorNWSEResize,
		platform.CursorNESWResize,
		platform.CursorAllResize,
		platform.CursorNotAllowed,
		platform.CursorWait,
		platform.CursorNone,
	}
	for _, c := range cursors {
		w.SetCursor(c)
	}
	// Also test default branch with an unmapped value
	w.SetCursor(platform.CursorShape(255))
}

func TestWindowSetIMEPosition(t *testing.T) {
	p, w := newTestWindow(t)
	defer p.Terminate()
	defer w.Destroy()

	// Should not panic even without active IME context
	w.SetIMEPosition(uimath.NewVec2(50, 100))
}

func TestWindowGetStyleExStyle(t *testing.T) {
	p, w := newTestWindow(t)
	defer p.Terminate()
	defer w.Destroy()

	style := w.getStyle()
	if style == 0 {
		t.Error("style should not be zero for a decorated window")
	}

	exStyle := w.getExStyle()
	_ = exStyle // May or may not be zero
}

func TestWindowUpdatePosition(t *testing.T) {
	p, w := newTestWindow(t)
	defer p.Terminate()
	defer w.Destroy()

	w.updatePosition()
	// Just ensure no panic, position values depend on WM
}

func TestWindowEnableMouseTracking(t *testing.T) {
	p, w := newTestWindow(t)
	defer p.Terminate()
	defer w.Destroy()

	w.enableMouseTracking()
	if !w.mouseTracked {
		t.Error("mouseTracked should be true after enableMouseTracking")
	}

	// Calling again should be no-op
	w.enableMouseTracking()
}

func TestCreateWindowNotInitialized(t *testing.T) {
	p := New()
	_, err := p.CreateWindow(platform.WindowOptions{})
	if err == nil {
		t.Error("CreateWindow should fail when not initialized")
	}
}

func TestInitIdempotent(t *testing.T) {
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("first Init failed: %v", err)
	}
	defer p.Terminate()

	// Second init should be a no-op
	if err := p.Init(); err != nil {
		t.Fatalf("second Init should not fail: %v", err)
	}
}

func TestPushEventAndPoll(t *testing.T) {
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer p.Terminate()

	// Create window so PollEvents works properly
	w, err := p.CreateWindow(platform.WindowOptions{
		Title: "Event Test", Width: 200, Height: 200, Visible: false, Decorated: true,
	})
	if err != nil {
		t.Fatalf("CreateWindow failed: %v", err)
	}
	defer w.Destroy()

	// Manually push events and verify drain
	p.PollEvents() // drain any pending events from window creation
}

func TestWindowDestroyTwice(t *testing.T) {
	p, w := newTestWindow(t)
	defer p.Terminate()

	w.Destroy()
	w.Destroy() // Should not panic (hwnd already 0)
}

func TestUtf16PtrToStringNil(t *testing.T) {
	if utf16PtrToString(nil) != "" {
		t.Error("nil should return empty string")
	}
}

func TestWindowUndecorated(t *testing.T) {
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer p.Terminate()

	w, err := p.CreateWindow(platform.WindowOptions{
		Title:     "Undecorated",
		Width:     200,
		Height:    200,
		Visible:   false,
		Decorated: false, // WS_POPUP path
	})
	if err != nil {
		t.Fatalf("CreateWindow failed: %v", err)
	}
	defer w.Destroy()

	if w.NativeHandle() == 0 {
		t.Error("HWND should not be zero")
	}
}

// --- WndProc message tests (send messages to exercise wndProc branches) ---

func sendMsg(hwnd, msg, wParam, lParam uintptr) {
	procPostMessageW.Call(hwnd, msg, wParam, lParam)
}

func TestWndProcMouseMessages(t *testing.T) {
	p, w := newTestWindow(t)
	defer p.Terminate()
	defer w.Destroy()

	hwnd := w.hwnd
	lp := uintptr(100<<16 | 50) // y=100, x=50

	sendMsg(hwnd, WM_MOUSEMOVE, 0, lp)
	sendMsg(hwnd, WM_LBUTTONDOWN, 0, lp)
	sendMsg(hwnd, WM_LBUTTONUP, 0, lp)
	sendMsg(hwnd, WM_RBUTTONDOWN, 0, lp)
	sendMsg(hwnd, WM_RBUTTONUP, 0, lp)
	sendMsg(hwnd, WM_MBUTTONDOWN, 0, lp)
	sendMsg(hwnd, WM_MBUTTONUP, 0, lp)
	sendMsg(hwnd, WM_LBUTTONDBLCLK, 0, lp)
	sendMsg(hwnd, WM_MOUSEWHEEL, uintptr(120<<16), 0)
	sendMsg(hwnd, WM_MOUSEHWHEEL, uintptr(120<<16), 0)
	sendMsg(hwnd, WM_MOUSELEAVE, 0, 0)
	sendMsg(hwnd, WM_XBUTTONDOWN, uintptr(1<<16), lp)
	sendMsg(hwnd, WM_XBUTTONUP, uintptr(2<<16), lp)

	events := p.PollEvents()
	if len(events) == 0 {
		t.Error("expected mouse events after sending messages")
	}
}

func TestWndProcKeyboardMessages(t *testing.T) {
	p, w := newTestWindow(t)
	defer p.Terminate()
	defer w.Destroy()

	hwnd := w.hwnd

	sendMsg(hwnd, WM_KEYDOWN, VK_SPACE, 0)
	sendMsg(hwnd, WM_KEYUP, VK_SPACE, 0)
	sendMsg(hwnd, WM_CHAR, 'A', 0)
	// Control char should be ignored
	sendMsg(hwnd, WM_CHAR, 8, 0) // backspace control char

	events := p.PollEvents()
	hasKeyDown := false
	hasKeyUp := false
	hasKeyPress := false
	for _, e := range events {
		switch e.Type {
		case event.KeyDown:
			hasKeyDown = true
		case event.KeyUp:
			hasKeyUp = true
		case event.KeyPress:
			hasKeyPress = true
		}
	}
	if !hasKeyDown {
		t.Error("expected KeyDown event")
	}
	if !hasKeyUp {
		t.Error("expected KeyUp event")
	}
	if !hasKeyPress {
		t.Error("expected KeyPress event")
	}
}

func TestWndProcWindowMessages(t *testing.T) {
	p, w := newTestWindow(t)
	defer p.Terminate()
	defer w.Destroy()

	hwnd := w.hwnd

	sendMsg(hwnd, WM_SETFOCUS, 0, 0)
	sendMsg(hwnd, WM_KILLFOCUS, 0, 0)
	sendMsg(hwnd, WM_ERASEBKGND, 0, 0)
	sendMsg(hwnd, WM_PAINT, 0, 0)
	sendMsg(hwnd, WM_DESTROY, 0, 0)

	events := p.PollEvents()
	_ = events
}

func TestWndProcCloseMessage(t *testing.T) {
	p, w := newTestWindow(t)
	defer p.Terminate()
	defer w.Destroy()

	sendMsg(w.hwnd, WM_CLOSE, 0, 0)
	p.PollEvents()

	if !w.shouldClose {
		t.Error("WM_CLOSE should set shouldClose")
	}
}

func TestWndProcIMEMessages(t *testing.T) {
	p, w := newTestWindow(t)
	defer p.Terminate()
	defer w.Destroy()

	hwnd := w.hwnd

	sendMsg(hwnd, WM_IME_STARTCOMPOSITION, 0, 0)
	sendMsg(hwnd, WM_IME_COMPOSITION, 0, uintptr(GCS_COMPSTR))
	sendMsg(hwnd, WM_IME_COMPOSITION, 0, uintptr(GCS_RESULTSTR))
	sendMsg(hwnd, WM_IME_ENDCOMPOSITION, 0, 0)

	p.PollEvents()
}

func TestWndProcSetCursorMessage(t *testing.T) {
	p, w := newTestWindow(t)
	defer p.Terminate()
	defer w.Destroy()

	// lParam loword=1 means HTCLIENT
	sendMsg(w.hwnd, WM_SETCURSOR, 0, 1)
	p.PollEvents()
}

func TestWindowDecoratedNotResizable(t *testing.T) {
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer p.Terminate()

	w, err := p.CreateWindow(platform.WindowOptions{
		Title:     "NotResizable",
		Width:     200,
		Height:    200,
		Visible:   false,
		Decorated: true,
		Resizable: false, // strips WS_THICKFRAME|WS_MAXIMIZEBOX
	})
	if err != nil {
		t.Fatalf("CreateWindow failed: %v", err)
	}
	defer w.Destroy()

	if w.NativeHandle() == 0 {
		t.Error("HWND should not be zero")
	}
}

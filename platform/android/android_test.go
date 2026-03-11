//go:build android

package android

import (
	"testing"

	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/platform"
)

// TestNew verifies that New() returns a non-nil Platform.
func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
}

// TestInit verifies that Init() succeeds and marks the platform as initialized.
func TestInit(t *testing.T) {
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}
	if !p.inited {
		t.Error("Init() should mark platform as inited")
	}
	p.Terminate()
}

// TestInitIdempotent verifies that calling Init() twice is safe.
func TestInitIdempotent(t *testing.T) {
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("first Init() failed: %v", err)
	}
	if err := p.Init(); err != nil {
		t.Fatalf("second Init() failed: %v", err)
	}
	p.Terminate()
}

// TestCreateWindow verifies that CreateWindow returns a non-nil Window.
func TestCreateWindow(t *testing.T) {
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer p.Terminate()

	win, err := p.CreateWindow(platform.WindowOptions{
		Title:  "Test",
		Width:  1920,
		Height: 1080,
	})
	if err != nil {
		t.Fatalf("CreateWindow() returned error: %v", err)
	}
	if win == nil {
		t.Fatal("CreateWindow() returned nil window")
	}

	w, ok := win.(*Window)
	if !ok {
		t.Fatal("CreateWindow() returned wrong type")
	}
	if w.width != 1920 || w.height != 1080 {
		t.Errorf("Window size = %dx%d; want 1920x1080", w.width, w.height)
	}
}

// TestCreateWindowDefaultSize verifies that zero dimensions get sensible defaults.
func TestCreateWindowDefaultSize(t *testing.T) {
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer p.Terminate()

	win, err := p.CreateWindow(platform.WindowOptions{})
	if err != nil {
		t.Fatalf("CreateWindow() failed: %v", err)
	}
	w, _ := win.(*Window)
	if w.width <= 0 || w.height <= 0 {
		t.Errorf("Window size = %dx%d; want positive dimensions", w.width, w.height)
	}
}

// TestPollEventsEmpty verifies that PollEvents returns empty slice initially.
func TestPollEventsEmpty(t *testing.T) {
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer p.Terminate()

	evs := p.PollEvents()
	if len(evs) != 0 {
		t.Errorf("PollEvents() = %d events; want 0", len(evs))
	}
}

// TestPollEventsAfterPush verifies that pushed events are returned by PollEvents.
func TestPollEventsAfterPush(t *testing.T) {
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer p.Terminate()

	// Simulate an event pushed from the Android runtime
	p.PushEvent(event.Event{Type: event.WindowClose})

	evs := p.PollEvents()
	if len(evs) != 1 {
		t.Fatalf("PollEvents() = %d events; want 1", len(evs))
	}
	if evs[0].Type != event.WindowClose {
		t.Errorf("event type = %v; want WindowClose", evs[0].Type)
	}

	// Second poll should return empty (events are drained)
	evs2 := p.PollEvents()
	if len(evs2) != 0 {
		t.Errorf("second PollEvents() = %d events; want 0", len(evs2))
	}
}

// TestGetPrimaryMonitorDPI verifies the Android default DPI.
func TestGetPrimaryMonitorDPI(t *testing.T) {
	p := New()
	dpi := p.GetPrimaryMonitorDPI()
	if dpi != 160.0 {
		t.Errorf("GetPrimaryMonitorDPI() = %v; want 160.0", dpi)
	}
}

// TestNativeHandle verifies that the window handle starts at 0.
func TestNativeHandle(t *testing.T) {
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer p.Terminate()

	win, _ := p.CreateWindow(platform.WindowOptions{Width: 100, Height: 100})
	if win.NativeHandle() != 0 {
		t.Errorf("NativeHandle() = %v; want 0 (no ANativeWindow set)", win.NativeHandle())
	}
}

// TestWindowInterface verifies that Window implements platform.Window and all
// interface methods work without panicking.
func TestWindowInterface(t *testing.T) {
	p := New()
	if err := p.Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer p.Terminate()

	win, err := p.CreateWindow(platform.WindowOptions{Width: 800, Height: 600})
	if err != nil {
		t.Fatalf("CreateWindow() failed: %v", err)
	}

	// Exercise all interface methods without crashing
	w, h := win.Size()
	if w != 800 || h != 600 {
		t.Errorf("Size() = %d,%d; want 800,600", w, h)
	}

	win.SetSize(1280, 720)
	w, h = win.Size()
	if w != 1280 || h != 720 {
		t.Errorf("after SetSize, Size() = %d,%d; want 1280,720", w, h)
	}

	fw, fh := win.FramebufferSize()
	if fw != 1280 || fh != 720 {
		t.Errorf("FramebufferSize() = %d,%d; want 1280,720", fw, fh)
	}

	if win.DPIScale() <= 0 {
		t.Error("DPIScale() should be positive")
	}

	win.SetTitle("Hello Android")
	win.SetFullscreen(false)
	win.SetVisible(true)
	win.ShowDeferred() // no-op on Android
	win.SetMinSize(320, 240)
	win.SetMaxSize(3840, 2160)
	win.SetCursor(platform.CursorArrow)

	win.SetShouldClose(false)
	if win.ShouldClose() {
		t.Error("ShouldClose() should be false after SetShouldClose(false)")
	}

	win.SetShouldClose(true)
	if !win.ShouldClose() {
		t.Error("ShouldClose() should be true after SetShouldClose(true)")
	}

	x, y := win.ClientToScreen(10, 20)
	if x != 10 || y != 20 {
		t.Errorf("ClientToScreen(10,20) = %d,%d; want 10,20", x, y)
	}

	result := win.ShowContextMenu(0, 0, nil)
	if result != -1 {
		t.Errorf("ShowContextMenu() = %d; want -1", result)
	}

	win.Destroy()
}

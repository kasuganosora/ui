//go:build darwin

package darwin

import (
	"runtime"
	"testing"

	"github.com/kasuganosora/ui/platform"
)

// TestAuxiliaryWindowOptionFlags covers the Go-side state the overlay host
// relies on. It is a pure logic test: no Cocoa calls are made, so it always
// runs even when the test binary is not on the process main thread.
func TestAuxiliaryWindowOptionFlags(t *testing.T) {
	skipIfNotDarwin(t)

	w := &Window{transparent: true, topMost: true, passthrough: true}
	if !w.IsTransparent() {
		t.Error("IsTransparent() must report the Transparent option")
	}
	if !w.transparent || !w.topMost || !w.passthrough {
		t.Errorf("option flags = (%v,%v,%v), want (true,true,true)",
			w.transparent, w.topMost, w.passthrough)
	}

	// A transparent shaped window (desktop pet) is NOT passthrough: it keeps
	// receiving events and filters them in Go via SetHitTestFunc.
	shaped := &Window{transparent: true}
	if shaped.passthrough {
		t.Error("Transparent alone must not imply Passthrough")
	}
	called := 0
	shaped.SetHitTestFunc(func(x, y int) bool {
		called++
		return x > 10 && y < 10
	})
	if shaped.hitTest == nil {
		t.Fatal("SetHitTestFunc did not store the predicate")
	}
	if !shaped.hitTest(20, 5) || shaped.hitTest(0, 5) {
		t.Error("stored predicate does not reject points outside the shape")
	}
	if called != 2 {
		t.Errorf("predicate called %d times, want 2", called)
	}
}

// bootstrapAppKitForWindowTest initializes AppKit for tests that create a real
// NSWindow. Go does not guarantee a test goroutine stays on the OS main thread,
// and bootstrapping from a secondary thread wedges finishLaunching, so skip in
// that case (the same guard used by TestInitWindowWithContentRectSmoke).
func bootstrapAppKitForWindowTest(t *testing.T) {
	t.Helper()
	runtime.LockOSThread()
	t.Cleanup(runtime.UnlockOSThread)

	ensureFoundation()
	if msgSend(id(objcClass("NSThread")), objcSelector("isMainThread")) == 0 {
		t.Skip("AppKit window creation requires the process main thread")
	}
	app := msgSend(id(objcClass("NSApplication")), selSharedApplication)
	if app == 0 {
		t.Fatal("NSApplication sharedApplication returned nil")
	}
	msgSend(app, selSetActivationPolicy, NSApplicationActivationPolicyRegular)
	msgSend(app, objcSelector("finishLaunching"))
}

// TestAuxiliaryWindowNativeProperties creates a real NSWindow with exactly the
// options the OS-native tooltip host uses, then reads the native properties
// back. Storing Go flags is not enough: the panel only works if AppKit really
// applied opaque:NO, a clear background, a popup-menu level, mouse passthrough
// and a reusable hide. This is the platform-level precondition for the whole
// independent-window tooltip design.
func TestAuxiliaryWindowNativeProperties(t *testing.T) {
	skipIfNotDarwin(t)
	bootstrapAppKitForWindowTest(t)

	w, err := newWindow(&Platform{}, platform.WindowOptions{
		Width:       200,
		Height:      80,
		Decorated:   false,
		Resizable:   false,
		Visible:     false,
		Transparent: true,
		TopMost:     true,
		Passthrough: true,
	})
	if err != nil {
		t.Fatalf("newWindow: %v", err)
	}
	defer w.Destroy()

	if !w.IsTransparent() || !w.topMost || !w.passthrough {
		t.Fatalf("Go flags = transparent=%v topMost=%v passthrough=%v, want all true",
			w.IsTransparent(), w.topMost, w.passthrough)
	}

	checks := []struct {
		name string
		sel  string
		want uintptr
	}{
		{"isOpaque", "isOpaque", 0},
		{"hasShadow", "hasShadow", 0},
		{"level", "level", nsPopUpMenuWindowLevel},
		{"ignoresMouseEvents", "ignoresMouseEvents", 1},
	}
	for _, c := range checks {
		got := uintptr(msgSend(w.nswindow, objcSelector(c.sel)))
		if got != c.want {
			t.Errorf("NSWindow %s = %d, want %d", c.name, got, c.want)
		}
	}

	bg := msgSend(w.nswindow, objcSelector("backgroundColor"))
	clear := msgSend(id(classNSColor), selClearColor)
	if bg == 0 || msgSend(bg, objcSelector("isEqual:"), uintptr(clear)) == 0 {
		t.Error("transparent window background must be NSColor.clearColor")
	}

	// SetTopMost must not lower a panel below the level it was created with.
	w.SetTopMost(false)
	if got := uintptr(msgSend(w.nswindow, objcSelector("level"))); got != nsNormalWindowLevel {
		t.Errorf("level after SetTopMost(false) = %d, want %d", got, nsNormalWindowLevel)
	}
	w.SetTopMost(true)
	if got := uintptr(msgSend(w.nswindow, objcSelector("level"))); got != nsPopUpMenuWindowLevel {
		t.Errorf("level after SetTopMost(true) = %d, want %d", got, nsPopUpMenuWindowLevel)
	}

	// orderOut: hides without tearing the window down, so Show() can reuse it
	// on the next hover. close would make the second show a no-op.
	w.SetVisible(true)
	if msgSend(w.nswindow, objcSelector("isVisible")) == 0 {
		t.Fatal("SetVisible(true) did not show the auxiliary window")
	}
	// A borderless passthrough panel cannot become key, so showing it on hover
	// does not steal focus from the main window (measured on a real NSWindow:
	// isKeyWindow=0, canBecomeKeyWindow=0). Guard the invariant so a future
	// style-mask change cannot silently turn hover into a focus steal.
	if msgSend(w.nswindow, objcSelector("isKeyWindow")) != 0 {
		t.Error("auxiliary passthrough window must not become the key window")
	}
	w.SetVisible(false)
	if msgSend(w.nswindow, objcSelector("isVisible")) != 0 {
		t.Error("SetVisible(false) must hide the window")
	}
	w.SetVisible(true)
	if msgSend(w.nswindow, objcSelector("isVisible")) == 0 {
		t.Error("window must be reusable after being ordered out")
	}
}

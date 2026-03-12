package devtools

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// ─── NewServer defaults ───────────────────────────────────────────────────────

func TestNewServer_Defaults(t *testing.T) {
	s := NewServer(Options{})
	if s.opts.Addr != "127.0.0.1:9222" {
		t.Errorf("default addr: got %q", s.opts.Addr)
	}
	if s.opts.AppName != "UI App" {
		t.Errorf("default appName: got %q", s.opts.AppName)
	}
	if s.router == nil {
		t.Error("router should be initialised")
	}
}

func TestNewServer_Custom(t *testing.T) {
	s := NewServer(Options{Addr: "127.0.0.1:9999", AppName: "My App"})
	if s.opts.Addr != "127.0.0.1:9999" {
		t.Errorf("custom addr: got %q", s.opts.Addr)
	}
}

// ─── localHostPort ────────────────────────────────────────────────────────────

func TestLocalHostPort_Standard(t *testing.T) {
	s := NewServer(Options{Addr: "127.0.0.1:9222"})
	if got := s.localHostPort(); got != "localhost:9222" {
		t.Errorf("want localhost:9222, got %q", got)
	}
}

func TestLocalHostPort_ColonOnly(t *testing.T) {
	s := NewServer(Options{Addr: ":8080"})
	if got := s.localHostPort(); got != "localhost:8080" {
		t.Errorf("want localhost:8080, got %q", got)
	}
}

func TestLocalHostPort_Fallback(t *testing.T) {
	s := NewServer(Options{Addr: "invalid"})
	got := s.localHostPort()
	if got != "localhost:9222" {
		t.Errorf("fallback: got %q", got)
	}
}

// ─── Attach / setHighlight ────────────────────────────────────────────────────

func TestAttach_SetsMarkDirty(t *testing.T) {
	s := NewServer(Options{})
	called := false
	s.Attach(func() { called = true })

	// setHighlight should call markDirty on change
	s.setHighlight(core.ElementID(5))
	if !called {
		t.Error("markDirty should have been called")
	}
}

func TestAttach_NilMarkDirty(t *testing.T) {
	s := NewServer(Options{})
	s.Attach(nil)
	// Should not panic
	s.setHighlight(core.ElementID(5))
}

func TestSetHighlight_SameIDNoCallback(t *testing.T) {
	s := NewServer(Options{})
	callCount := 0
	s.Attach(func() { callCount++ })

	s.setHighlight(core.ElementID(1))
	s.setHighlight(core.ElementID(1)) // same ID, should not call again
	if callCount != 1 {
		t.Errorf("want 1 call, got %d", callCount)
	}
}

func TestSetHighlight_ClearHighlight(t *testing.T) {
	s := NewServer(Options{})
	calls := 0
	s.Attach(func() { calls++ })

	s.setHighlight(core.ElementID(5))
	s.setHighlight(core.InvalidElementID) // clear
	if calls != 2 {
		t.Errorf("want 2 calls, got %d", calls)
	}
}

// ─── getSnapshot ─────────────────────────────────────────────────────────────

func TestGetSnapshot_NilWhenEmpty(t *testing.T) {
	s := NewServer(Options{})
	if s.getSnapshot() != nil {
		t.Error("fresh server should return nil snapshot")
	}
}

func TestGetSnapshot_ReturnsSnapshot(t *testing.T) {
	s := NewServer(Options{})
	snap := &Snapshot{
		Nodes:      make(map[core.ElementID]*NodeSnapshot),
		ViewWidth:  800,
		ViewHeight: 600,
	}
	s.snapMu.Lock()
	s.snapshot = snap
	s.snapMu.Unlock()

	got := s.getSnapshot()
	if got != snap {
		t.Error("getSnapshot should return the stored snapshot")
	}
}

// ─── DrawOverlay ─────────────────────────────────────────────────────────────

func TestDrawOverlay_NoSnapshot(t *testing.T) {
	s := NewServer(Options{})
	buf := render.NewCommandBuffer()
	// Should not panic with nil snapshot
	s.DrawOverlay(buf)
}

func TestDrawOverlay_NoHighlight(t *testing.T) {
	s := NewServer(Options{})
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.NewRect(0, 0, 100, 50))

	s.snapMu.Lock()
	s.snapshot = snap
	s.snapMu.Unlock()

	buf := render.NewCommandBuffer()
	s.DrawOverlay(buf) // highlight = invalid, should draw nothing
	if len(buf.Overlays()) != 0 {
		t.Errorf("no highlight set: expected no overlays, got %d", len(buf.Overlays()))
	}
}

func TestDrawOverlay_WithHighlight(t *testing.T) {
	s := NewServer(Options{})
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.NewRect(10, 20, 100, 50))

	s.snapMu.Lock()
	s.snapshot = snap
	s.snapMu.Unlock()
	s.setHighlight(1)
	// Reset after Attach call
	s.overlayMu.Lock()
	s.highlightID = 1
	s.overlayMu.Unlock()

	buf := render.NewCommandBuffer()
	s.DrawOverlay(buf)
	// Should add at least the content-box overlay
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlay rects to be added")
	}
}

func TestDrawOverlay_InvalidElementInSnapshot(t *testing.T) {
	s := NewServer(Options{})
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.NewRect(10, 20, 100, 50))

	s.snapMu.Lock()
	s.snapshot = snap
	s.snapMu.Unlock()

	// Highlight an ID that doesn't exist
	s.overlayMu.Lock()
	s.highlightID = 999
	s.overlayMu.Unlock()

	buf := render.NewCommandBuffer()
	s.DrawOverlay(buf) // should not panic
}

// ─── DrawOverlayLabel ─────────────────────────────────────────────────────────

func TestDrawOverlayLabel_NilRenderer(t *testing.T) {
	s := NewServer(Options{})
	buf := render.NewCommandBuffer()
	s.DrawOverlayLabel(buf, nil, 0) // should not panic
}

func TestDrawOverlayLabel_NoHighlight(t *testing.T) {
	s := NewServer(Options{})
	buf := render.NewCommandBuffer()
	s.DrawOverlayLabel(buf, nil, 1) // nil renderer, should not panic
}

// ─── Log ─────────────────────────────────────────────────────────────────────

func TestLog_NilWhenNoSessions(t *testing.T) {
	s := NewServer(Options{})
	// Should not panic with no sessions
	s.Log("info", "javascript", "hello")
}

// ─── broadcast ───────────────────────────────────────────────────────────────

func TestBroadcast_NoSessions(t *testing.T) {
	s := NewServer(Options{})
	// Should not panic
	s.broadcast("test.event", map[string]any{})
}

// ─── Stop ────────────────────────────────────────────────────────────────────

func TestStop_WithoutStart(t *testing.T) {
	s := NewServer(Options{})
	// httpSrv is nil; Stop should handle this gracefully
	if err := s.Stop(nil); err == nil {
		// acceptably no-ops or returns error; just should not panic
	}
}

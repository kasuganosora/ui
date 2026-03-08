package perf

import (
	"testing"
	"time"
)

func TestNilProfiler(t *testing.T) {
	var p *Profiler

	// All methods should be no-ops and not panic.
	token := p.Begin("test")
	if token != -1 {
		t.Errorf("nil Begin should return -1, got %d", token)
	}
	p.End(token)
	p.End(-1)
	p.BeginFrame()
	p.EndFrame()

	if idx := p.FrameIndex(); idx != 0 {
		t.Errorf("nil FrameIndex should return 0, got %d", idx)
	}
	if name := p.ScopeName(0); name != "" {
		t.Errorf("nil ScopeName should return empty, got %q", name)
	}

	stats := p.LastFrame()
	if stats.Frame != 0 || stats.Total != 0 || stats.Scopes != nil {
		t.Errorf("nil LastFrame should return zero FrameStats, got %+v", stats)
	}

	history := p.History(10)
	if history != nil {
		t.Errorf("nil History should return nil, got %v", history)
	}
}

func TestBeginEndRecordsTiming(t *testing.T) {
	p := New(10)
	p.BeginFrame()

	tok := p.Begin("work")
	time.Sleep(5 * time.Millisecond)
	p.End(tok)

	p.EndFrame()

	stats := p.LastFrame()
	if stats.Total < 5*time.Millisecond {
		t.Errorf("expected at least 5ms total, got %v", stats.Total)
	}
	if d, ok := stats.Scopes["work"]; !ok {
		t.Error("scope 'work' not found in stats")
	} else if d < 5*time.Millisecond {
		t.Errorf("expected scope 'work' >= 5ms, got %v", d)
	}
}

func TestBeginFrameEndFrameCycles(t *testing.T) {
	p := New(10)

	if idx := p.FrameIndex(); idx != 0 {
		t.Errorf("initial frame index should be 0, got %d", idx)
	}

	p.BeginFrame()
	p.EndFrame()

	if idx := p.FrameIndex(); idx != 1 {
		t.Errorf("after one frame, index should be 1, got %d", idx)
	}

	p.BeginFrame()
	p.EndFrame()

	if idx := p.FrameIndex(); idx != 2 {
		t.Errorf("after two frames, index should be 2, got %d", idx)
	}
}

func TestLastFrameReturnsCorrectData(t *testing.T) {
	p := New(10)

	// Frame 0: scope "A"
	p.BeginFrame()
	tok := p.Begin("A")
	time.Sleep(2 * time.Millisecond)
	p.End(tok)
	p.EndFrame()

	// Frame 1: scope "B"
	p.BeginFrame()
	tok = p.Begin("B")
	time.Sleep(2 * time.Millisecond)
	p.End(tok)
	p.EndFrame()

	stats := p.LastFrame()
	if stats.Frame != 1 {
		t.Errorf("expected frame 1, got %d", stats.Frame)
	}
	if _, ok := stats.Scopes["B"]; !ok {
		t.Error("expected scope 'B' in last frame")
	}
	if _, ok := stats.Scopes["A"]; ok {
		t.Error("scope 'A' should not be in last frame")
	}
}

func TestHistoryReturnsCorrectCount(t *testing.T) {
	p := New(10)

	for i := 0; i < 5; i++ {
		p.BeginFrame()
		tok := p.Begin("work")
		p.End(tok)
		p.EndFrame()
	}

	history := p.History(3)
	if len(history) != 3 {
		t.Fatalf("expected 3 frames in history, got %d", len(history))
	}
	// Should be frames 2, 3, 4
	if history[0].Frame != 2 {
		t.Errorf("expected first history frame to be 2, got %d", history[0].Frame)
	}
	if history[2].Frame != 4 {
		t.Errorf("expected last history frame to be 4, got %d", history[2].Frame)
	}

	// Requesting more than available
	history = p.History(100)
	if len(history) != 5 {
		t.Errorf("expected 5 frames (all available), got %d", len(history))
	}
}

func TestMultipleScopesInOneFrame(t *testing.T) {
	p := New(10)
	p.BeginFrame()

	tok1 := p.Begin("layout")
	time.Sleep(2 * time.Millisecond)
	p.End(tok1)

	tok2 := p.Begin("render")
	time.Sleep(2 * time.Millisecond)
	p.End(tok2)

	tok3 := p.Begin("layout") // same scope again
	time.Sleep(2 * time.Millisecond)
	p.End(tok3)

	p.EndFrame()

	stats := p.LastFrame()
	if len(stats.Scopes) != 2 {
		t.Errorf("expected 2 distinct scopes, got %d", len(stats.Scopes))
	}
	if _, ok := stats.Scopes["layout"]; !ok {
		t.Error("scope 'layout' not found")
	}
	if _, ok := stats.Scopes["render"]; !ok {
		t.Error("scope 'render' not found")
	}
	// layout should have two entries summed
	if stats.Scopes["layout"] < 4*time.Millisecond {
		t.Errorf("expected layout >= 4ms (two calls), got %v", stats.Scopes["layout"])
	}
}

func TestRingBufferWraps(t *testing.T) {
	capacity := 4
	p := New(capacity)

	// Write more frames than capacity
	for i := 0; i < 10; i++ {
		p.BeginFrame()
		tok := p.Begin("step")
		p.End(tok)
		p.EndFrame()
	}

	if idx := p.FrameIndex(); idx != 10 {
		t.Errorf("expected frame index 10, got %d", idx)
	}

	// History should return at most `capacity` frames
	history := p.History(10)
	if len(history) != capacity {
		t.Errorf("expected %d frames from history, got %d", capacity, len(history))
	}

	// Should be the last `capacity` frames: 6, 7, 8, 9
	if history[0].Frame != 6 {
		t.Errorf("expected oldest frame 6, got %d", history[0].Frame)
	}
	if history[capacity-1].Frame != 9 {
		t.Errorf("expected newest frame 9, got %d", history[capacity-1].Frame)
	}
}

func TestScopeName(t *testing.T) {
	p := New(10)
	p.BeginFrame()
	tok := p.Begin("myScope")
	p.End(tok)
	p.EndFrame()

	if name := p.ScopeName(0); name != "myScope" {
		t.Errorf("expected 'myScope', got %q", name)
	}
	// Out of range
	if name := p.ScopeName(999); name != "" {
		t.Errorf("expected empty for invalid ID, got %q", name)
	}
}

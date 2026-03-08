package perf

import (
	"sync"
	"time"
)

// ScopeID identifies a named profiling scope.
type ScopeID uint16

// Profiler provides lightweight per-frame performance instrumentation.
// All methods are nil-safe (no-ops on nil receiver) for zero-cost when disabled.
type Profiler struct {
	mu       sync.Mutex
	scopes   map[string]ScopeID
	names    []string
	records  [][]ScopeRecord // ring buffer of N frames, each frame has slice of records
	current  []scopeEntry    // open scope stack for current frame
	frameIdx uint64
	capacity int // max frames to keep
}

// ScopeRecord holds timing data for one completed scope in a frame.
type ScopeRecord struct {
	Scope    ScopeID
	Start    time.Duration // offset from frame start
	Duration time.Duration
}

type scopeEntry struct {
	scope ScopeID
	start time.Time
}

// FrameStats holds aggregate stats for a single frame.
type FrameStats struct {
	Frame  uint64
	Total  time.Duration
	Scopes map[string]time.Duration // scope name -> total duration
}

// New creates a profiler keeping the last `capacity` frames of data.
func New(capacity int) *Profiler {
	if capacity <= 0 {
		capacity = 120
	}
	return &Profiler{
		scopes:   make(map[string]ScopeID),
		records:  make([][]ScopeRecord, capacity),
		capacity: capacity,
	}
}

// scope returns or creates a ScopeID for the given name.
func (p *Profiler) scope(name string) ScopeID {
	if id, ok := p.scopes[name]; ok {
		return id
	}
	id := ScopeID(len(p.names))
	p.names = append(p.names, name)
	p.scopes[name] = id
	return id
}

// ScopeName returns the name for a ScopeID.
func (p *Profiler) ScopeName(id ScopeID) string {
	if p == nil || int(id) >= len(p.names) {
		return ""
	}
	return p.names[id]
}

// Begin starts a named scope. Returns a token to pass to End.
func (p *Profiler) Begin(name string) int {
	if p == nil {
		return -1
	}
	p.mu.Lock()
	id := p.scope(name)
	idx := len(p.current)
	p.current = append(p.current, scopeEntry{scope: id, start: time.Now()})
	p.mu.Unlock()
	return idx
}

// End closes the scope opened by Begin. Pass the token returned by Begin.
func (p *Profiler) End(token int) {
	if p == nil || token < 0 {
		return
	}
	end := time.Now()
	p.mu.Lock()
	if token < len(p.current) {
		entry := p.current[token]
		rec := ScopeRecord{
			Scope:    entry.scope,
			Start:    entry.start.Sub(entry.start), // will be relative
			Duration: end.Sub(entry.start),
		}
		idx := p.frameIdx % uint64(p.capacity)
		p.records[idx] = append(p.records[idx], rec)
		// Remove from stack (may not be in order for nested scopes)
		if token == len(p.current)-1 {
			p.current = p.current[:token]
		}
	}
	p.mu.Unlock()
}

// BeginFrame starts a new frame, clearing the slot in the ring buffer.
func (p *Profiler) BeginFrame() {
	if p == nil {
		return
	}
	p.mu.Lock()
	idx := p.frameIdx % uint64(p.capacity)
	p.records[idx] = p.records[idx][:0]
	p.current = p.current[:0]
	p.mu.Unlock()
}

// EndFrame advances the frame counter.
func (p *Profiler) EndFrame() {
	if p == nil {
		return
	}
	p.mu.Lock()
	p.frameIdx++
	p.mu.Unlock()
}

// FrameIndex returns the current frame number.
func (p *Profiler) FrameIndex() uint64 {
	if p == nil {
		return 0
	}
	return p.frameIdx
}

// LastFrame returns stats for the most recently completed frame.
func (p *Profiler) LastFrame() FrameStats {
	if p == nil {
		return FrameStats{}
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.frameIdx == 0 {
		return FrameStats{}
	}
	frame := p.frameIdx - 1
	idx := frame % uint64(p.capacity)
	records := p.records[idx]

	stats := FrameStats{
		Frame:  frame,
		Scopes: make(map[string]time.Duration),
	}
	for _, rec := range records {
		name := p.names[rec.Scope]
		stats.Scopes[name] += rec.Duration
		stats.Total += rec.Duration
	}
	return stats
}

// History returns stats for the last N frames.
func (p *Profiler) History(n int) []FrameStats {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	if n > p.capacity {
		n = p.capacity
	}
	if uint64(n) > p.frameIdx {
		n = int(p.frameIdx)
	}

	result := make([]FrameStats, n)
	for i := 0; i < n; i++ {
		frame := p.frameIdx - uint64(n) + uint64(i)
		idx := frame % uint64(p.capacity)
		records := p.records[idx]

		stats := FrameStats{
			Frame:  frame,
			Scopes: make(map[string]time.Duration),
		}
		for _, rec := range records {
			name := p.names[rec.Scope]
			stats.Scopes[name] += rec.Duration
			stats.Total += rec.Duration
		}
		result[i] = stats
	}
	return result
}

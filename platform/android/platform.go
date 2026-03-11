//go:build android

// Package android provides basic Android platform support.
// Uses the Android NDK native activity model via an ANativeWindow* handle.
//
// Current state: compiles and provides the platform.Platform interface.
// Actual window creation is driven by ANativeActivity_onCreate which is
// called by the Android runtime; this package provides the Go-side bridge.
//
// JNI integration and full ANativeActivity lifecycle management are left
// as future work; this package is sufficient for compilation and interface
// compatibility.
package android

import (
	"os"
	"sync"

	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/platform"
)

// Platform implements platform.Platform for Android.
type Platform struct {
	windows []*Window
	events  []event.Event
	mu      sync.Mutex
	inited  bool
}

// New creates a new Android platform instance.
func New() *Platform {
	return &Platform{}
}

// Init implements platform.Platform.
func (p *Platform) Init() error {
	p.inited = true
	return nil
}

// CreateWindow implements platform.Platform.
// On Android, the window is provided by the OS via ANativeActivity_onCreate,
// not created by the application. This method is kept for interface compatibility.
func (p *Platform) CreateWindow(opts platform.WindowOptions) (platform.Window, error) {
	w := &Window{
		p:      p,
		width:  opts.Width,
		height: opts.Height,
	}
	if w.width == 0 {
		w.width = 1920
	}
	if w.height == 0 {
		w.height = 1080
	}
	p.windows = append(p.windows, w)
	return w, nil
}

// PollEvents implements platform.Platform.
func (p *Platform) PollEvents() []event.Event {
	p.mu.Lock()
	evs := make([]event.Event, len(p.events))
	copy(evs, p.events)
	p.events = p.events[:0]
	p.mu.Unlock()
	return evs
}

// ProcessMessages implements platform.Platform.
func (p *Platform) ProcessMessages() {}

// GetClipboardText implements platform.Platform.
// Requires JNI to access Android ClipboardManager; stub returns empty string.
func (p *Platform) GetClipboardText() string { return "" }

// SetClipboardText implements platform.Platform (stub).
func (p *Platform) SetClipboardText(text string) {}

// GetPrimaryMonitorDPI implements platform.Platform.
// Android typical baseline DPI is 160 (mdpi). Actual DPI requires JNI.
func (p *Platform) GetPrimaryMonitorDPI() float32 { return 160.0 }

// GetSystemLocale implements platform.Platform.
// Reads LANG env var as a best-effort; full locale requires JNI.
func (p *Platform) GetSystemLocale() string {
	if lang := os.Getenv("LANG"); lang != "" {
		return lang
	}
	return "en_US"
}

// Terminate implements platform.Platform.
func (p *Platform) Terminate() {
	p.windows = nil
	p.inited = false
}

// PushEvent adds an event to the queue (called from JNI bridge or input callbacks).
func (p *Platform) PushEvent(e event.Event) {
	p.mu.Lock()
	p.events = append(p.events, e)
	p.mu.Unlock()
}

// Compile-time interface check.
var _ platform.Platform = (*Platform)(nil)

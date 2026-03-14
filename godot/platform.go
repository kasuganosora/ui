package godot

import (
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/platform"
)

// HeadlessPlatform implements platform.Platform for headless/embedded operation.
// It creates HeadlessWindow instances and manages an injected event queue.
type HeadlessPlatform struct {
	events    []event.Event
	clipboard string
	locale    string
}

// NewHeadlessPlatform creates a headless platform.
func NewHeadlessPlatform() *HeadlessPlatform {
	return &HeadlessPlatform{locale: "en-US"}
}

func (p *HeadlessPlatform) Init() error { return nil }

func (p *HeadlessPlatform) CreateWindow(opts platform.WindowOptions) (platform.Window, error) {
	return NewHeadlessWindow(opts.Width, opts.Height, 1.0), nil
}

func (p *HeadlessPlatform) PollEvents() []event.Event {
	if len(p.events) == 0 {
		return nil
	}
	out := make([]event.Event, len(p.events))
	copy(out, p.events)
	p.events = p.events[:0]
	return out
}

// InjectEvent adds an event to the queue, returned by the next PollEvents.
func (p *HeadlessPlatform) InjectEvent(evt event.Event) {
	p.events = append(p.events, evt)
}

func (p *HeadlessPlatform) ProcessMessages()          {}
func (p *HeadlessPlatform) GetClipboardText() string   { return p.clipboard }
func (p *HeadlessPlatform) SetClipboardText(t string)  { p.clipboard = t }
func (p *HeadlessPlatform) GetPrimaryMonitorDPI() float32 { return 96 }
func (p *HeadlessPlatform) GetSystemLocale() string    { return p.locale }
func (p *HeadlessPlatform) Terminate()                 {}

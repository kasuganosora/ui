package ui

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// EmbeddedUI runs the UI system embedded within a host application.
// The host provides the render backend and pumps events.
type EmbeddedUI struct {
	tree       *core.Tree
	dispatcher *core.Dispatcher
	config     *widget.Config
	root       widget.Widget
	buf        *render.CommandBuffer
	width      float32
	height     float32
	layers     *LayerManager
}

// NewEmbeddedUI creates a new embedded UI instance.
// The host application provides the render backend.
func NewEmbeddedUI(cfg *widget.Config) *EmbeddedUI {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	tree := core.NewTree()
	return &EmbeddedUI{
		tree:       tree,
		dispatcher: core.NewDispatcher(tree),
		config:     cfg,
		buf:        render.NewCommandBuffer(),
		layers:     NewLayerManager(),
	}
}

// Tree returns the element tree.
func (e *EmbeddedUI) Tree() *core.Tree { return e.tree }

// Config returns the widget config.
func (e *EmbeddedUI) Config() *widget.Config { return e.config }

// Dispatcher returns the event dispatcher.
func (e *EmbeddedUI) Dispatcher() *core.Dispatcher { return e.dispatcher }

// Layers returns the layer manager.
func (e *EmbeddedUI) Layers() *LayerManager { return e.layers }

// SetRoot sets the root widget.
func (e *EmbeddedUI) SetRoot(w widget.Widget) {
	e.root = w
	e.tree.AppendChild(e.tree.Root(), w.ElementID())
}

// Resize updates the viewport size.
func (e *EmbeddedUI) Resize(width, height float32) {
	e.width = width
	e.height = height
	e.tree.SetLayout(e.tree.Root(), core.LayoutResult{
		Bounds: rect(0, 0, width, height),
	})
}

// HandleEvent dispatches an event to the UI.
// The host calls this with translated events from its input system.
func (e *EmbeddedUI) HandleEvent(evt *event.Event) {
	if evt.Type.IsMouse() {
		target := e.tree.HitTest(evt.GlobalX, evt.GlobalY)
		if target != core.InvalidElementID {
			e.dispatcher.Dispatch(target, evt)
		}
	} else if evt.Type.IsKeyboard() || evt.Type.IsIME() {
		// Send to focused element
		e.tree.Walk(e.tree.Root(), func(id core.ElementID, _ int) bool {
			if elem := e.tree.Get(id); elem != nil && elem.IsFocused() {
				e.dispatcher.Dispatch(id, evt)
				return false
			}
			return true
		})
	}
}

// NeedsRedraw returns true if the UI needs to be redrawn.
func (e *EmbeddedUI) NeedsRedraw() bool {
	return e.tree.NeedsRender()
}

// Render generates render commands for the current frame.
// The host submits the returned buffer to its backend.
func (e *EmbeddedUI) Render() *render.CommandBuffer {
	e.buf.Reset()
	if e.root != nil {
		e.root.Draw(e.buf)
	}
	// Draw layers (HUD, dialog, chat, etc.)
	e.layers.Draw(e.buf)
	e.tree.ClearAllDirty()
	return e.buf
}


package ui

import (
	"encoding/json"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// HeadlessUI runs the UI in headless mode, exporting render commands
// as serializable data. Useful for remote rendering, testing, or
// server-side rendering.
type HeadlessUI struct {
	tree       *core.Tree
	dispatcher *core.Dispatcher
	config     *widget.Config
	root       widget.Widget
	buf        *render.CommandBuffer
	width      float32
	height     float32
}

// NewHeadlessUI creates a headless UI instance.
func NewHeadlessUI(cfg *widget.Config) *HeadlessUI {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	tree := core.NewTree()
	return &HeadlessUI{
		tree:       tree,
		dispatcher: core.NewDispatcher(tree),
		config:     cfg,
		buf:        render.NewCommandBuffer(),
	}
}

// Tree returns the element tree.
func (h *HeadlessUI) Tree() *core.Tree { return h.tree }

// Config returns the widget config.
func (h *HeadlessUI) Config() *widget.Config { return h.config }

// SetRoot sets the root widget.
func (h *HeadlessUI) SetRoot(w widget.Widget) {
	h.root = w
	h.tree.AppendChild(h.tree.Root(), w.ElementID())
}

// Resize sets the viewport size.
func (h *HeadlessUI) Resize(width, height float32) {
	h.width = width
	h.height = height
	h.tree.SetLayout(h.tree.Root(), core.LayoutResult{
		Bounds: rect(0, 0, width, height),
	})
}

// HandleEvent dispatches an event.
func (h *HeadlessUI) HandleEvent(evt *event.Event) {
	if evt.Type.IsMouse() {
		target := h.tree.HitTest(evt.GlobalX, evt.GlobalY)
		if target != core.InvalidElementID {
			h.dispatcher.Dispatch(target, evt)
		}
	}
}

// Render generates render commands and returns them.
func (h *HeadlessUI) Render() *render.CommandBuffer {
	h.buf.Reset()
	if h.root != nil {
		h.root.Draw(h.buf)
	}
	h.tree.ClearAllDirty()
	return h.buf
}

// ExportedCommand is a serializable render command.
type ExportedCommand struct {
	Type    string  `json:"type"`
	ZOrder  int32   `json:"z_order"`
	Opacity float32 `json:"opacity"`

	// Rect fields
	X            float32 `json:"x"`
	Y            float32 `json:"y"`
	W            float32 `json:"w"`
	H            float32 `json:"h"`
	FillR        float32 `json:"fill_r,omitempty"`
	FillG        float32 `json:"fill_g,omitempty"`
	FillB        float32 `json:"fill_b,omitempty"`
	FillA        float32 `json:"fill_a,omitempty"`
	BorderR      float32 `json:"border_r,omitempty"`
	BorderG      float32 `json:"border_g,omitempty"`
	BorderB      float32 `json:"border_b,omitempty"`
	BorderA      float32 `json:"border_a,omitempty"`
	BorderWidth  float32 `json:"border_width,omitempty"`
	CornerRadius float32 `json:"corner_radius,omitempty"`
}

// ExportCommands converts render commands to a serializable format.
func ExportCommands(buf *render.CommandBuffer) []ExportedCommand {
	cmds := buf.Commands()
	result := make([]ExportedCommand, 0, len(cmds))
	for _, cmd := range cmds {
		switch cmd.Type {
		case render.CmdRect:
			r := cmd.Rect
			result = append(result, ExportedCommand{
				Type:         "rect",
				ZOrder:       cmd.ZOrder,
				Opacity:      cmd.Opacity,
				X:            r.Bounds.X,
				Y:            r.Bounds.Y,
				W:            r.Bounds.Width,
				H:            r.Bounds.Height,
				FillR:        r.FillColor.R,
				FillG:        r.FillColor.G,
				FillB:        r.FillColor.B,
				FillA:        r.FillColor.A,
				BorderR:      r.BorderColor.R,
				BorderG:      r.BorderColor.G,
				BorderB:      r.BorderColor.B,
				BorderA:      r.BorderColor.A,
				BorderWidth:  r.BorderWidth,
				CornerRadius: r.Corners.TopLeft,
			})
		case render.CmdClip:
			c := cmd.Clip
			result = append(result, ExportedCommand{
				Type: "clip",
				X:    c.Bounds.X,
				Y:    c.Bounds.Y,
				W:    c.Bounds.Width,
				H:    c.Bounds.Height,
			})
		}
	}
	return result
}

// ExportJSON exports render commands as JSON.
func ExportJSON(buf *render.CommandBuffer) ([]byte, error) {
	cmds := ExportCommands(buf)
	return json.Marshal(cmds)
}

package ui

import (
	"sort"

	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// LayerType identifies the purpose of a UI layer.
type LayerType uint8

const (
	LayerBase    LayerType = iota // Normal UI content
	LayerHUD                      // Game HUD overlay
	LayerDialog                   // Modal dialogs
	LayerChat                     // Chat overlay
	LayerTooltip                  // Tooltips (highest)
)

// Layer represents a single UI rendering layer.
type Layer struct {
	Name    string
	Type    LayerType
	ZBase   int32 // Base z-order for this layer
	Visible bool
	Widgets []widget.Widget
}

// LayerManager manages multiple UI layers with z-ordering.
type LayerManager struct {
	layers []*Layer
}

// NewLayerManager creates a layer manager.
func NewLayerManager() *LayerManager {
	return &LayerManager{}
}

// AddLayer adds a new layer. Layers are drawn in order of ZBase.
func (lm *LayerManager) AddLayer(name string, layerType LayerType, zBase int32) *Layer {
	l := &Layer{
		Name:    name,
		Type:    layerType,
		ZBase:   zBase,
		Visible: true,
	}
	lm.layers = append(lm.layers, l)
	sort.Slice(lm.layers, func(i, j int) bool {
		return lm.layers[i].ZBase < lm.layers[j].ZBase
	})
	return l
}

// GetLayer returns a layer by name.
func (lm *LayerManager) GetLayer(name string) *Layer {
	for _, l := range lm.layers {
		if l.Name == name {
			return l
		}
	}
	return nil
}

// RemoveLayer removes a layer by name.
func (lm *LayerManager) RemoveLayer(name string) {
	for i, l := range lm.layers {
		if l.Name == name {
			lm.layers = append(lm.layers[:i], lm.layers[i+1:]...)
			return
		}
	}
}

// Draw renders all visible layers into the command buffer.
func (lm *LayerManager) Draw(buf *render.CommandBuffer) {
	for _, l := range lm.layers {
		if !l.Visible {
			continue
		}
		for _, w := range l.Widgets {
			w.Draw(buf)
		}
	}
}

// AddWidget adds a widget to a layer.
func (l *Layer) AddWidget(w widget.Widget) {
	l.Widgets = append(l.Widgets, w)
}

// RemoveWidget removes a widget from the layer.
func (l *Layer) RemoveWidget(w widget.Widget) {
	for i, lw := range l.Widgets {
		if lw.ElementID() == w.ElementID() {
			l.Widgets = append(l.Widgets[:i], l.Widgets[i+1:]...)
			return
		}
	}
}

// Clear removes all widgets from the layer.
func (l *Layer) Clear() {
	l.Widgets = l.Widgets[:0]
}

// rect helper (avoids importing math in this package file)
func rect(x, y, w, h float32) uimath.Rect {
	return uimath.NewRect(x, y, w, h)
}

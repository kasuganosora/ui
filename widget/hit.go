package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
)

// hitCells are short-lived elements placed over an open popup so hit testing
// and element counts follow the cells that are actually drawn.
type hitCells struct {
	tree   *core.Tree
	parent core.ElementID
	ids    []core.ElementID
}

func (h *hitCells) clear() {
	if h == nil || h.tree == nil {
		return
	}
	for _, id := range h.ids {
		h.tree.DestroyElement(id)
	}
	h.ids = nil
}

func (h *hitCells) set(rects []uimath.Rect, fn func(i int)) {
	h.clear()
	if h.tree == nil || h.parent == core.InvalidElementID {
		return
	}
	for i, r := range rects {
		id := h.tree.CreateElement(core.TypeCustom)
		parent := h.tree.Root()
		if parent == core.InvalidElementID {
			parent = h.parent
		}
		h.tree.AppendChild(parent, id)
		h.tree.SetLayout(id, core.LayoutResult{Bounds: r})
		idx := i
		h.tree.AddHandler(id, event.MouseClick, func(e *event.Event) {
			if fn != nil {
				fn(idx)
			}
			e.StopPropagation()
		})
		h.ids = append(h.ids, id)
	}
}

func pointIn(x, y float32, r uimath.Rect) bool {
	return x >= r.X && y >= r.Y && x < r.X+r.Width && y < r.Y+r.Height
}

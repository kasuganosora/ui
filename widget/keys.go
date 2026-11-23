package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
)

func bindActivate(tree *core.Tree, id core.ElementID, allow func() bool, act func()) {
	tree.SetFocusable(id, true)
	tree.AddHandler(id, event.KeyDown, func(e *event.Event) {
		if e.Phase == event.PhaseCapture {
			return
		}
		if e.Key != event.KeyEnter && e.Key != event.KeySpace {
			return
		}
		if elem := tree.Get(id); elem != nil && (!elem.IsVisible() || !elem.IsEnabled()) {
			return
		}
		if allow != nil && !allow() {
			return
		}
		if act != nil {
			act()
		}
		e.StopPropagation()
	})
}

func bindArrows(tree *core.Tree, id core.ElementID, fn func(event.Key)) {
	tree.SetFocusable(id, true)
	tree.AddHandler(id, event.KeyDown, func(e *event.Event) {
		if e.Phase == event.PhaseCapture {
			return
		}
		switch e.Key {
		case event.KeyArrowLeft, event.KeyArrowRight, event.KeyArrowUp, event.KeyArrowDown, event.KeyEnter, event.KeySpace, event.KeyEscape, event.KeyHome, event.KeyEnd:
			fn(e.Key)
			e.StopPropagation()
		}
	})
}

package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
)

func bindActivate(tree *core.Tree, id core.ElementID, allow func() bool, act func()) {
	tree.SetFocusable(id, true)
	tree.AddHandler(id, event.KeyDown, func(e *event.Event) {
		if e.Key != event.KeyEnter && e.Key != event.KeySpace {
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
		switch e.Key {
		case event.KeyArrowLeft, event.KeyArrowRight, event.KeyArrowUp, event.KeyArrowDown, event.KeyEnter, event.KeySpace, event.KeyEscape, event.KeyHome, event.KeyEnd:
			fn(e.Key)
			e.StopPropagation()
		}
	})
}

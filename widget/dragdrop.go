package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// DragSource marks a widget as draggable.
type DragSource struct {
	Widget   Widget
	Data     any
	DragIcon Widget // Optional custom drag icon
}

// DropTarget marks a widget as a drop target.
type DropTarget struct {
	Widget   Widget
	Accept   func(data any) bool  // Returns true if the data is acceptable
	OnDrop   func(data any)       // Called when a valid drop occurs
	OnEnter  func(data any)       // Called when drag enters
	OnLeave  func()               // Called when drag leaves
}

// DragDropManager coordinates drag and drop operations.
type DragDropManager struct {
	tree       *core.Tree
	sources    map[core.ElementID]*DragSource
	targets    map[core.ElementID]*DropTarget

	// Active drag state
	dragging   bool
	dragData   any
	dragIcon   Widget
	dragStartX float32
	dragStartY float32
	dragX      float32
	dragY      float32
	sourceID   core.ElementID
	overTarget core.ElementID
}

// NewDragDropManager creates a drag-drop manager.
func NewDragDropManager(tree *core.Tree) *DragDropManager {
	return &DragDropManager{
		tree:    tree,
		sources: make(map[core.ElementID]*DragSource),
		targets: make(map[core.ElementID]*DropTarget),
	}
}

// RegisterSource registers a widget as a drag source.
func (dd *DragDropManager) RegisterSource(source *DragSource) {
	id := source.Widget.ElementID()
	dd.sources[id] = source

	dd.tree.AddHandler(id, event.MouseDown, func(e *event.Event) {
		dd.dragStartX = e.GlobalX
		dd.dragStartY = e.GlobalY
		dd.sourceID = id
	})
}

// RegisterTarget registers a widget as a drop target.
func (dd *DragDropManager) RegisterTarget(target *DropTarget) {
	dd.targets[target.Widget.ElementID()] = target
}

// HandleEvent processes mouse events for drag-drop.
// Call this from the main event loop.
func (dd *DragDropManager) HandleEvent(evt *event.Event) {
	switch evt.Type {
	case event.MouseMove:
		if dd.sourceID != core.InvalidElementID && !dd.dragging {
			// Check drag threshold
			dx := evt.GlobalX - dd.dragStartX
			dy := evt.GlobalY - dd.dragStartY
			if dx*dx+dy*dy > 25 { // 5px threshold
				dd.startDrag(evt.GlobalX, evt.GlobalY)
			}
		}
		if dd.dragging {
			dd.dragX = evt.GlobalX
			dd.dragY = evt.GlobalY
			dd.updateDropTarget(evt.GlobalX, evt.GlobalY)
		}

	case event.MouseUp:
		if dd.dragging {
			dd.finishDrag()
		}
		dd.sourceID = core.InvalidElementID
	}
}

func (dd *DragDropManager) startDrag(x, y float32) {
	source := dd.sources[dd.sourceID]
	if source == nil {
		return
	}
	dd.dragging = true
	dd.dragData = source.Data
	dd.dragIcon = source.DragIcon
	dd.dragX = x
	dd.dragY = y
}

func (dd *DragDropManager) updateDropTarget(x, y float32) {
	hit := dd.tree.HitTest(x, y)

	// Find the first ancestor that's a drop target
	targetID := core.InvalidElementID
	for id := hit; id != core.InvalidElementID; {
		if _, ok := dd.targets[id]; ok {
			targetID = id
			break
		}
		elem := dd.tree.Get(id)
		if elem == nil {
			break
		}
		id = elem.ParentID()
	}

	if targetID != dd.overTarget {
		// Leave old target
		if dd.overTarget != core.InvalidElementID {
			if t := dd.targets[dd.overTarget]; t != nil && t.OnLeave != nil {
				t.OnLeave()
			}
		}
		// Enter new target
		if targetID != core.InvalidElementID {
			if t := dd.targets[targetID]; t != nil && t.OnEnter != nil {
				t.OnEnter(dd.dragData)
			}
		}
		dd.overTarget = targetID
	}
}

func (dd *DragDropManager) finishDrag() {
	if dd.overTarget != core.InvalidElementID {
		target := dd.targets[dd.overTarget]
		if target != nil {
			accept := target.Accept == nil || target.Accept(dd.dragData)
			if accept && target.OnDrop != nil {
				target.OnDrop(dd.dragData)
			}
		}
	}
	dd.dragging = false
	dd.dragData = nil
	dd.dragIcon = nil
	dd.overTarget = core.InvalidElementID
}

// IsDragging returns true if a drag operation is in progress.
func (dd *DragDropManager) IsDragging() bool { return dd.dragging }

// DragData returns the data being dragged (nil if not dragging).
func (dd *DragDropManager) DragData() any { return dd.dragData }

// Draw renders the drag icon at the current mouse position.
func (dd *DragDropManager) Draw(buf *render.CommandBuffer) {
	if !dd.dragging {
		return
	}

	if dd.dragIcon != nil {
		dd.dragIcon.Draw(buf)
		return
	}

	// Default drag indicator: a semi-transparent rect
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(dd.dragX-20, dd.dragY-20, 40, 40),
		FillColor: uimath.RGBA(0.1, 0.4, 0.8, 0.5),
		Corners:   uimath.CornersAll(4),
	}, 100, 0.7)
}

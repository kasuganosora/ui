package core

import (
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// ElementID uniquely identifies an element in the tree.
type ElementID uint64

const InvalidElementID ElementID = 0

// ElementType identifies the kind of element.
type ElementType string

// Common element types
const (
	TypeDiv      ElementType = "div"
	TypeText     ElementType = "text"
	TypeButton   ElementType = "button"
	TypeInput    ElementType = "input"
	TypeImage    ElementType = "image"
	TypeScroll   ElementType = "scroll"
	TypeCustom   ElementType = "custom"
)

// HitTestFunc tests whether a point (in local element coordinates) should
// be considered "solid" for hit-testing purposes.  Return true if the point
// is opaque / clickable, false if it should pass through to whatever is
// underneath.  When nil, the default rectangular bounds check is used.
type HitTestFunc func(localX, localY float32) bool

// Element represents a single UI element in the tree.
type Element struct {
	id         ElementID
	elemType   ElementType
	classes    []string
	parent     ElementID
	children   []ElementID
	properties map[string]any

	// Layout results (computed by layout engine)
	layout     LayoutResult

	// State flags
	visible    bool
	enabled    bool
	focused    bool
	hovered    bool
	dirty      DirtyFlags

	// Event handlers
	handlers   map[event.Type][]EventHandler

	// Optional per-pixel hit test (for shaped / transparent regions).
	hitTestFn  HitTestFunc
}

// EventHandler is a function that handles an event.
type EventHandler func(e *event.Event)

// LayoutResult contains computed layout information. Value object.
type LayoutResult struct {
	Bounds   uimath.Rect  // Position and size relative to parent
	Padding  uimath.Edges // Computed padding
	Margin   uimath.Edges // Computed margin
	Border   uimath.Edges // Computed border widths
	ContentBounds uimath.Rect // Content area (bounds minus padding/border)
}

// DirtyFlags tracks what needs to be recalculated.
type DirtyFlags uint8

const (
	DirtyNone    DirtyFlags = 0
	DirtyLayout  DirtyFlags = 1 << iota // Needs layout recalculation
	DirtyPaint                           // Needs repaint
	DirtyStyle                           // Needs style recalculation
	DirtyAll     = DirtyLayout | DirtyPaint | DirtyStyle
)

// NewElement creates a new element with the given type.
func NewElement(id ElementID, elemType ElementType) *Element {
	return &Element{
		id:         id,
		elemType:   elemType,
		visible:    true,
		enabled:    true,
		dirty:      DirtyAll,
		handlers:   make(map[event.Type][]EventHandler),
		properties: make(map[string]any),
	}
}

// --- Getters (elements are accessed through the tree aggregate root) ---

func (e *Element) ID() ElementID        { return e.id }
func (e *Element) Type() ElementType     { return e.elemType }
func (e *Element) Classes() []string     { return e.classes }
func (e *Element) ParentID() ElementID   { return e.parent }
func (e *Element) ChildIDs() []ElementID { return e.children }
func (e *Element) Layout() LayoutResult  { return e.layout }
func (e *Element) IsVisible() bool       { return e.visible }
func (e *Element) IsEnabled() bool       { return e.enabled }
func (e *Element) IsFocused() bool       { return e.focused }
func (e *Element) IsHovered() bool       { return e.hovered }
func (e *Element) IsDirty(flags DirtyFlags) bool { return e.dirty&flags != 0 }
func (e *Element) HasHandler(t event.Type) bool { return len(e.handlers[t]) > 0 }

func (e *Element) Property(key string) (any, bool) {
	v, ok := e.properties[key]
	return v, ok
}

func (e *Element) TextContent() string {
	if v, ok := e.properties["text"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// Tree is the aggregate root for the element tree.
// All element mutations must go through Tree methods.
type Tree struct {
	elements map[ElementID]*Element
	root     ElementID
	nextID   ElementID
	idGen    func() ElementID
}

// NewTree creates a new element tree.
func NewTree() *Tree {
	t := &Tree{
		elements: make(map[ElementID]*Element),
		nextID:   1,
	}
	// Create the root element
	root := NewElement(t.allocID(), TypeDiv)
	t.root = root.id
	t.elements[root.id] = root
	return t
}

func (t *Tree) allocID() ElementID {
	id := t.nextID
	t.nextID++
	return id
}

// Root returns the root element ID.
func (t *Tree) Root() ElementID {
	return t.root
}

// Get returns an element by ID. Returns nil if not found.
func (t *Tree) Get(id ElementID) *Element {
	return t.elements[id]
}

// NeedsRender returns true if the root element is dirty (needs repaint).
func (t *Tree) NeedsRender() bool {
	root := t.elements[t.root]
	return root != nil && root.IsDirty(DirtyPaint|DirtyLayout)
}

// NeedsLayout returns true if any element has the DirtyLayout flag set,
// meaning widget structure or sizing changed and full layout recomputation is needed.
// Scroll-only changes (which only set DirtyPaint) don't require full layout.
func (t *Tree) NeedsLayout() bool {
	root := t.elements[t.root]
	return root != nil && root.IsDirty(DirtyLayout)
}

// ClearAllDirty clears all dirty flags on all elements in the tree.
func (t *Tree) ClearAllDirty() {
	for _, elem := range t.elements {
		if elem != nil {
			elem.dirty = DirtyNone
		}
	}
}

// CreateElement creates a new element and adds it to the tree (unparented).
func (t *Tree) CreateElement(elemType ElementType) ElementID {
	elem := NewElement(t.allocID(), elemType)
	t.elements[elem.id] = elem
	return elem.id
}

// AppendChild adds a child element to a parent.
func (t *Tree) AppendChild(parentID, childID ElementID) bool {
	parent := t.elements[parentID]
	child := t.elements[childID]
	if parent == nil || child == nil {
		return false
	}
	// Remove from old parent if any
	if child.parent != InvalidElementID {
		t.RemoveChild(child.parent, childID)
	}
	child.parent = parentID
	parent.children = append(parent.children, childID)
	t.markDirty(parentID, DirtyLayout)
	return true
}

// InsertBefore inserts a child before a reference child.
func (t *Tree) InsertBefore(parentID, childID, beforeID ElementID) bool {
	parent := t.elements[parentID]
	child := t.elements[childID]
	if parent == nil || child == nil {
		return false
	}
	if child.parent != InvalidElementID {
		t.RemoveChild(child.parent, childID)
	}
	child.parent = parentID
	for i, id := range parent.children {
		if id == beforeID {
			parent.children = append(parent.children[:i+1], parent.children[i:]...)
			parent.children[i] = childID
			t.markDirty(parentID, DirtyLayout)
			return true
		}
	}
	// If beforeID not found, append
	parent.children = append(parent.children, childID)
	t.markDirty(parentID, DirtyLayout)
	return true
}

// BringChildToFront moves a child to the end of the parent's child list so it
// draws on top of siblings. This ONLY marks DirtyPaint (not DirtyLayout) because
// z-order reordering does not affect any element's size or position.
// Use this instead of RemoveChild+AppendChild to avoid triggering a full CSS re-layout.
func (t *Tree) BringChildToFront(parentID, childID ElementID) bool {
	parent := t.elements[parentID]
	if parent == nil {
		return false
	}
	for i, id := range parent.children {
		if id == childID {
			if i == len(parent.children)-1 {
				return true // already at front
			}
			parent.children = append(parent.children[:i], parent.children[i+1:]...)
			parent.children = append(parent.children, childID)
			t.markDirty(parentID, DirtyPaint) // z-order only — layout unchanged
			return true
		}
	}
	return false
}

// RemoveChild removes a child from a parent.
func (t *Tree) RemoveChild(parentID, childID ElementID) bool {
	parent := t.elements[parentID]
	child := t.elements[childID]
	if parent == nil || child == nil {
		return false
	}
	for i, id := range parent.children {
		if id == childID {
			parent.children = append(parent.children[:i], parent.children[i+1:]...)
			child.parent = InvalidElementID
			t.markDirty(parentID, DirtyLayout)
			return true
		}
	}
	return false
}

// DestroyElement removes an element and all its descendants from the tree.
func (t *Tree) DestroyElement(id ElementID) {
	elem := t.elements[id]
	if elem == nil {
		return
	}
	// Remove from parent
	if elem.parent != InvalidElementID {
		t.RemoveChild(elem.parent, id)
	}
	// Recursively destroy children
	t.destroyRecursive(id)
}

func (t *Tree) destroyRecursive(id ElementID) {
	elem := t.elements[id]
	if elem == nil {
		return
	}
	for _, childID := range elem.children {
		t.destroyRecursive(childID)
	}
	delete(t.elements, id)
}

// SetProperty sets a property on an element.
func (t *Tree) SetProperty(id ElementID, key string, value any) {
	if elem := t.elements[id]; elem != nil {
		elem.properties[key] = value
		t.markDirty(id, DirtyPaint)
	}
}

// SetHitTestFunc sets a custom hit-test function for an element.
// When set, after the rectangular bounds check passes the function is called
// with local coordinates (relative to element top-left).  If it returns false
// the element is treated as transparent at that point and the hit test
// continues to elements underneath.
func (t *Tree) SetHitTestFunc(id ElementID, fn HitTestFunc) {
	if elem := t.elements[id]; elem != nil {
		elem.hitTestFn = fn
	}
}

// HitTestFunc returns the custom hit-test function for an element (may be nil).
func (t *Tree) HitTestFunc(id ElementID) HitTestFunc {
	if elem := t.elements[id]; elem != nil {
		return elem.hitTestFn
	}
	return nil
}

// SetClasses sets the CSS classes on an element.
func (t *Tree) SetClasses(id ElementID, classes []string) {
	if elem := t.elements[id]; elem != nil {
		elem.classes = classes
		t.markDirty(id, DirtyStyle|DirtyLayout)
	}
}

// SetVisible sets the visibility of an element.
func (t *Tree) SetVisible(id ElementID, visible bool) {
	if elem := t.elements[id]; elem != nil {
		if elem.visible != visible {
			elem.visible = visible
			t.markDirty(id, DirtyLayout|DirtyPaint)
		}
	}
}

// SetEnabled sets the enabled state of an element.
func (t *Tree) SetEnabled(id ElementID, enabled bool) {
	if elem := t.elements[id]; elem != nil {
		elem.enabled = enabled
		t.markDirty(id, DirtyPaint)
	}
}

// SetFocused sets the focused state. Only one element should be focused at a time.
// When setting focus to true, any previously focused element is automatically unfocused.
func (t *Tree) SetFocused(id ElementID, focused bool) {
	if focused {
		for eid, elem := range t.elements {
			if elem != nil && elem.focused && eid != id {
				elem.focused = false
				t.markDirty(eid, DirtyPaint)
			}
		}
	}
	if elem := t.elements[id]; elem != nil {
		elem.focused = focused
		t.markDirty(id, DirtyPaint)
	}
}

// ClearFocus removes focus from all elements.
func (t *Tree) ClearFocus() {
	for eid, elem := range t.elements {
		if elem != nil && elem.focused {
			elem.focused = false
			t.markDirty(eid, DirtyPaint)
		}
	}
}

// SetHovered sets the hovered state.
// Only marks the element dirty when the state actually changes to avoid
// spurious dirty propagation on every MouseMove event.
func (t *Tree) SetHovered(id ElementID, hovered bool) {
	if elem := t.elements[id]; elem != nil {
		if elem.hovered == hovered {
			return
		}
		elem.hovered = hovered
		t.markDirty(id, DirtyPaint)
	}
}

// SetLayout sets the computed layout for an element (called by layout engine).
func (t *Tree) SetLayout(id ElementID, layout LayoutResult) {
	if elem := t.elements[id]; elem != nil {
		elem.layout = layout
	}
}

// AddHandler registers an event handler on an element.
func (t *Tree) AddHandler(id ElementID, eventType event.Type, handler EventHandler) {
	if elem := t.elements[id]; elem != nil {
		elem.handlers[eventType] = append(elem.handlers[eventType], handler)
	}
}

// Handlers returns the event handlers for an element and event type.
func (t *Tree) Handlers(id ElementID, eventType event.Type) []EventHandler {
	if elem := t.elements[id]; elem != nil {
		return elem.handlers[eventType]
	}
	return nil
}

// ClearDirty clears dirty flags on an element.
func (t *Tree) ClearDirty(id ElementID, flags DirtyFlags) {
	if elem := t.elements[id]; elem != nil {
		elem.dirty &^= flags
	}
}

// MarkDirty marks an element (and its ancestors) as needing repaint.
func (t *Tree) MarkDirty(id ElementID) {
	t.markDirty(id, DirtyPaint)
}

func (t *Tree) markDirty(id ElementID, flags DirtyFlags) {
	elem := t.elements[id]
	for elem != nil {
		elem.dirty |= flags
		elem = t.elements[elem.parent]
	}
}

// Walk traverses the tree depth-first, calling fn for each element.
// If fn returns false, the subtree is skipped.
func (t *Tree) Walk(id ElementID, fn func(id ElementID, depth int) bool) {
	t.walkRecursive(id, 0, fn)
}

func (t *Tree) walkRecursive(id ElementID, depth int, fn func(ElementID, int) bool) {
	elem := t.elements[id]
	if elem == nil {
		return
	}
	if !fn(id, depth) {
		return
	}
	for _, childID := range elem.children {
		t.walkRecursive(childID, depth+1, fn)
	}
}

// ElementCount returns the total number of elements in the tree.
func (t *Tree) ElementCount() int {
	return len(t.elements)
}

// HitTest finds the deepest visible element at the given point.
func (t *Tree) HitTest(x, y float32) ElementID {
	return t.hitTestRecursive(t.root, x, y)
}

func (t *Tree) hitTestRecursive(id ElementID, x, y float32) ElementID {
	elem := t.elements[id]
	if elem == nil || !elem.visible {
		return InvalidElementID
	}
	b := elem.layout.Bounds
	if !b.Contains(uimath.Vec2{X: x, Y: y}) {
		return InvalidElementID
	}
	// Custom per-pixel hit test (shaped / transparent regions).
	if elem.hitTestFn != nil && !elem.hitTestFn(x-b.X, y-b.Y) {
		return InvalidElementID
	}
	// Check children in reverse order (last drawn = on top)
	for i := len(elem.children) - 1; i >= 0; i-- {
		hit := t.hitTestRecursive(elem.children[i], x, y)
		if hit != InvalidElementID {
			return hit
		}
	}
	return id
}

// Drawable is implemented by elements that can generate render commands.
type Drawable interface {
	Draw(elem *Element, buf *render.CommandBuffer)
}

// Package widget provides the P0 UI component library.
//
// Widgets bridge the element tree (core.Tree), layout engine, and render pipeline.
// Each widget owns one or more elements and manages its own style, state, and drawing.
package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Widget is the interface all UI components implement.
type Widget interface {
	// ElementID returns the root element ID in the tree.
	ElementID() core.ElementID

	// Style returns the layout style for this widget.
	Style() layout.Style

	// SetStyle updates the layout style.
	SetStyle(s layout.Style)

	// Draw emits render commands for this widget.
	Draw(buf *render.CommandBuffer)

	// Children returns child widgets (for containers).
	Children() []Widget

	// Destroy removes the widget and its element from the tree.
	Destroy()
}

// Base provides shared widget state. Embed this in concrete widgets.
type Base struct {
	id       core.ElementID
	tree     *core.Tree
	style    layout.Style
	config   *Config
	children []Widget
}

// NewBase creates a base widget with an element in the tree.
func NewBase(tree *core.Tree, elemType core.ElementType, cfg *Config) Base {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return Base{
		id:    tree.CreateElement(elemType),
		tree:  tree,
		style: layout.DefaultStyle(),
		config: cfg,
	}
}

func (b *Base) ElementID() core.ElementID { return b.id }
func (b *Base) Style() layout.Style       { return b.style }
func (b *Base) SetStyle(s layout.Style)   { b.style = s }
func (b *Base) Children() []Widget        { return b.children }
func (b *Base) Tree() *core.Tree          { return b.tree }
func (b *Base) Config() *Config           { return b.config }

func (b *Base) Element() *core.Element {
	return b.tree.Get(b.id)
}

func (b *Base) Bounds() uimath.Rect {
	if e := b.Element(); e != nil {
		return e.Layout().Bounds
	}
	return uimath.Rect{}
}

func (b *Base) Destroy() {
	for _, c := range b.children {
		c.Destroy()
	}
	b.tree.DestroyElement(b.id)
}

// On registers an event handler.
func (b *Base) On(eventType event.Type, handler core.EventHandler) {
	b.tree.AddHandler(b.id, eventType, handler)
}

// AppendChild adds a child widget.
func (b *Base) AppendChild(child Widget) {
	b.children = append(b.children, child)
	b.tree.AppendChild(b.id, child.ElementID())
}

// PrependChild inserts a child widget at the beginning.
func (b *Base) PrependChild(child Widget) {
	b.children = append([]Widget{child}, b.children...)
	if len(b.children) > 1 {
		b.tree.InsertBefore(b.id, child.ElementID(), b.children[1].ElementID())
	} else {
		b.tree.AppendChild(b.id, child.ElementID())
	}
}

// RemoveChild removes a child widget (does not destroy it).
func (b *Base) RemoveChild(child Widget) {
	for i, c := range b.children {
		if c.ElementID() == child.ElementID() {
			b.children = append(b.children[:i], b.children[i+1:]...)
			b.tree.RemoveChild(b.id, child.ElementID())
			return
		}
	}
}

// DrawChildren draws all child widgets.
func (b *Base) DrawChildren(buf *render.CommandBuffer) {
	for _, c := range b.children {
		c.Draw(buf)
	}
}

package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Layout provides a classic page structure: Header, Content, Footer, Aside.
type Layout struct {
	Base
	bgColor uimath.Color
}

// NewLayout creates a full-page layout container.
func NewLayout(tree *core.Tree, cfg *Config) *Layout {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	l := &Layout{
		Base: NewBase(tree, core.TypeDiv, cfg),
	}
	l.style.Display = layout.DisplayFlex
	l.style.FlexDirection = layout.FlexDirectionColumn
	l.style.Width = layout.Pct(100)
	l.style.Height = layout.Pct(100)
	return l
}

func (l *Layout) SetBgColor(c uimath.Color) { l.bgColor = c }

func (l *Layout) Draw(buf *render.CommandBuffer) {
	bounds := l.Bounds()
	if !bounds.IsEmpty() && !l.bgColor.IsTransparent() {
		buf.DrawRect(render.RectCmd{
			Bounds:    bounds,
			FillColor: l.bgColor,
		}, 0, 1)
	}
	l.DrawChildren(buf)
}

// Header is the top section of a Layout.
type Header struct {
	Base
	bgColor uimath.Color
	height  float32
}

// NewHeader creates a header section.
func NewHeader(tree *core.Tree, cfg *Config) *Header {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	h := &Header{
		Base:   NewBase(tree, core.TypeDiv, cfg),
		height: 64,
	}
	h.style.Display = layout.DisplayFlex
	h.style.FlexDirection = layout.FlexDirectionRow
	h.style.AlignItems = layout.AlignCenter
	h.style.Height = layout.Px(64)
	h.style.Padding = layout.EdgeValues{
		Left:  layout.Px(cfg.SpaceLG),
		Right: layout.Px(cfg.SpaceLG),
	}
	return h
}

func (h *Header) SetBgColor(c uimath.Color) { h.bgColor = c }
func (h *Header) SetHeight(v float32) {
	h.height = v
	h.style.Height = layout.Px(v)
}

func (h *Header) Draw(buf *render.CommandBuffer) {
	bounds := h.Bounds()
	if !bounds.IsEmpty() && !h.bgColor.IsTransparent() {
		buf.DrawRect(render.RectCmd{
			Bounds:    bounds,
			FillColor: h.bgColor,
		}, 0, 1)
	}
	h.DrawChildren(buf)
}

// Content is the main content area.
type Content struct {
	Base
	bgColor uimath.Color
}

// NewContent creates a content section that grows to fill available space.
func NewContent(tree *core.Tree, cfg *Config) *Content {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	c := &Content{
		Base: NewBase(tree, core.TypeDiv, cfg),
	}
	c.style.Display = layout.DisplayFlex
	c.style.FlexDirection = layout.FlexDirectionColumn
	c.style.FlexGrow = 1
	c.style.Overflow = layout.OverflowScroll
	return c
}

func (c *Content) SetBgColor(clr uimath.Color) { c.bgColor = clr }

func (c *Content) Draw(buf *render.CommandBuffer) {
	bounds := c.Bounds()
	if !bounds.IsEmpty() && !c.bgColor.IsTransparent() {
		buf.DrawRect(render.RectCmd{
			Bounds:    bounds,
			FillColor: c.bgColor,
		}, 0, 1)
	}
	c.DrawChildren(buf)
}

// Footer is the bottom section of a Layout.
type Footer struct {
	Base
	bgColor uimath.Color
	height  float32
}

// NewFooter creates a footer section.
func NewFooter(tree *core.Tree, cfg *Config) *Footer {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	f := &Footer{
		Base:   NewBase(tree, core.TypeDiv, cfg),
		height: 48,
	}
	f.style.Display = layout.DisplayFlex
	f.style.FlexDirection = layout.FlexDirectionRow
	f.style.AlignItems = layout.AlignCenter
	f.style.JustifyContent = layout.JustifyCenter
	f.style.Height = layout.Px(48)
	return f
}

func (f *Footer) SetBgColor(c uimath.Color) { f.bgColor = c }
func (f *Footer) SetHeight(v float32) {
	f.height = v
	f.style.Height = layout.Px(v)
}

func (f *Footer) Draw(buf *render.CommandBuffer) {
	bounds := f.Bounds()
	if !bounds.IsEmpty() && !f.bgColor.IsTransparent() {
		buf.DrawRect(render.RectCmd{
			Bounds:    bounds,
			FillColor: f.bgColor,
		}, 0, 1)
	}
	f.DrawChildren(buf)
}

// Aside is a sidebar section.
type Aside struct {
	Base
	bgColor uimath.Color
	width   float32
}

// NewAside creates a sidebar section.
func NewAside(tree *core.Tree, cfg *Config) *Aside {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	a := &Aside{
		Base:  NewBase(tree, core.TypeDiv, cfg),
		width: 200,
	}
	a.style.Display = layout.DisplayFlex
	a.style.FlexDirection = layout.FlexDirectionColumn
	a.style.Width = layout.Px(200)
	return a
}

func (a *Aside) SetBgColor(c uimath.Color) { a.bgColor = c }
func (a *Aside) SetWidth(w float32) {
	a.width = w
	a.style.Width = layout.Px(w)
}

func (a *Aside) Draw(buf *render.CommandBuffer) {
	bounds := a.Bounds()
	if !bounds.IsEmpty() && !a.bgColor.IsTransparent() {
		buf.DrawRect(render.RectCmd{
			Bounds:    bounds,
			FillColor: a.bgColor,
		}, 0, 1)
	}
	a.DrawChildren(buf)
}

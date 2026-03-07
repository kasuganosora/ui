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

// Content is the main content area with vertical scrolling support.
type Content struct {
	Base
	bgColor       uimath.Color
	scrollY       float32
	contentHeight float32 // total height of all children
	scrollBarDrag bool    // true when dragging the scrollbar thumb
	dragStartY    float32 // mouse Y when drag started
	dragStartScrl float32 // scrollY when drag started
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
func (c *Content) ScrollY() float32             { return c.scrollY }
func (c *Content) ContentHeight() float32       { return c.contentHeight }
func (c *Content) SetContentHeight(h float32)   { c.contentHeight = h }

// ScrollBy adjusts scroll offset by delta, clamping to valid range.
func (c *Content) ScrollBy(dy float32) {
	c.scrollY += dy
	c.clampScroll()
}

// ScrollTo sets the scroll offset, clamping to valid range.
func (c *Content) ScrollTo(y float32) {
	c.scrollY = y
	c.clampScroll()
}

func (c *Content) clampScroll() {
	bounds := c.Bounds()
	maxScroll := c.contentHeight - bounds.Height
	if maxScroll < 0 {
		maxScroll = 0
	}
	if c.scrollY < 0 {
		c.scrollY = 0
	}
	if c.scrollY > maxScroll {
		c.scrollY = maxScroll
	}
}

// scrollBarWidth is the width of the scrollbar track.
const scrollBarWidth = 8

// needsScroll returns true if content overflows the viewport.
func (c *Content) needsScroll() bool {
	bounds := c.Bounds()
	return c.contentHeight > bounds.Height
}

// HandleWheel processes a mouse wheel event for scrolling.
func (c *Content) HandleWheel(dy float32) {
	c.ScrollBy(-dy * 40) // 40px per scroll tick
}

// HandleScrollBarDown starts a scrollbar thumb drag.
func (c *Content) HandleScrollBarDown(globalY float32) bool {
	if !c.needsScroll() {
		return false
	}
	bounds := c.Bounds()
	thumbY, thumbH := c.thumbRect(bounds)
	if globalY >= thumbY && globalY <= thumbY+thumbH {
		c.scrollBarDrag = true
		c.dragStartY = globalY
		c.dragStartScrl = c.scrollY
		return true
	}
	return false
}

// HandleScrollBarMove updates scroll during a drag.
func (c *Content) HandleScrollBarMove(globalY float32) {
	if !c.scrollBarDrag {
		return
	}
	bounds := c.Bounds()
	trackH := bounds.Height - 4
	thumbH := c.thumbHeight(bounds)
	maxThumbY := trackH - thumbH
	if maxThumbY <= 0 {
		return
	}
	dy := globalY - c.dragStartY
	maxScroll := c.contentHeight - bounds.Height
	scrollDelta := dy * maxScroll / maxThumbY
	c.scrollY = c.dragStartScrl + scrollDelta
	c.clampScroll()
}

// HandleScrollBarUp ends a scrollbar drag.
func (c *Content) HandleScrollBarUp() {
	c.scrollBarDrag = false
}

// IsScrollBarDragging returns true if currently dragging the scrollbar.
func (c *Content) IsScrollBarDragging() bool {
	return c.scrollBarDrag
}

func (c *Content) thumbHeight(bounds uimath.Rect) float32 {
	if c.contentHeight <= 0 {
		return 0
	}
	ratio := bounds.Height / c.contentHeight
	if ratio > 1 {
		ratio = 1
	}
	h := (bounds.Height - 4) * ratio
	if h < 20 {
		h = 20
	}
	return h
}

func (c *Content) thumbRect(bounds uimath.Rect) (y, h float32) {
	h = c.thumbHeight(bounds)
	trackH := bounds.Height - 4
	maxThumbY := trackH - h
	maxScroll := c.contentHeight - bounds.Height
	if maxScroll <= 0 {
		return bounds.Y + 2, h
	}
	ratio := c.scrollY / maxScroll
	return bounds.Y + 2 + maxThumbY*ratio, h
}

func (c *Content) Draw(buf *render.CommandBuffer) {
	bounds := c.Bounds()
	if bounds.IsEmpty() {
		return
	}

	// Background
	if !c.bgColor.IsTransparent() {
		buf.DrawRect(render.RectCmd{
			Bounds:    bounds,
			FillColor: c.bgColor,
		}, 0, 1)
	}

	// Clip children to content area
	buf.PushClip(bounds)
	c.DrawChildren(buf)

	// Draw scrollbar if content overflows
	if c.needsScroll() {
		trackX := bounds.X + bounds.Width - scrollBarWidth - 2
		trackColor := uimath.RGBA(0, 0, 0, 0.05)
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(trackX, bounds.Y+2, scrollBarWidth, bounds.Height-4),
			FillColor: trackColor,
			Corners:   uimath.CornersAll(scrollBarWidth / 2),
		}, 10, 1)

		// Thumb
		thumbY, thumbH := c.thumbRect(bounds)
		thumbColor := uimath.RGBA(0, 0, 0, 0.25)
		if c.scrollBarDrag {
			thumbColor = uimath.RGBA(0, 0, 0, 0.45)
		}
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(trackX, thumbY, scrollBarWidth, thumbH),
			FillColor: thumbColor,
			Corners:   uimath.CornersAll(scrollBarWidth / 2),
		}, 11, 1)
	}

	// Reset clip so subsequent siblings (footer etc.) are not clipped
	buf.PopClip()
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
	bgColor     uimath.Color
	width       float32
	borderColor uimath.Color
	borderWidth float32
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
func (a *Aside) SetBorderRight(w float32, c uimath.Color) {
	a.borderWidth = w
	a.borderColor = c
}
func (a *Aside) SetWidth(w float32) {
	a.width = w
	a.style.Width = layout.Px(w)
}

func (a *Aside) Draw(buf *render.CommandBuffer) {
	bounds := a.Bounds()
	if bounds.IsEmpty() {
		return
	}
	if !a.bgColor.IsTransparent() {
		buf.DrawRect(render.RectCmd{
			Bounds:    bounds,
			FillColor: a.bgColor,
		}, 0, 1)
	}
	a.DrawChildren(buf)
	// Right border divider
	if a.borderWidth > 0 {
		bx := bounds.X + bounds.Width - a.borderWidth
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(bx, bounds.Y, a.borderWidth, bounds.Height),
			FillColor: a.borderColor,
		}, 1, 1)
	}
}

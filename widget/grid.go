package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	"github.com/kasuganosora/ui/render"
)

// Row is a horizontal flex container that divides space into columns.
type Row struct {
	Base
	gutter float32
}

// NewRow creates a row container.
func NewRow(tree *core.Tree, cfg *Config) *Row {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	r := &Row{
		Base: NewBase(tree, core.TypeDiv, cfg),
	}
	r.style.Display = layout.DisplayFlex
	r.style.FlexDirection = layout.FlexDirectionRow
	r.style.FlexWrap = layout.FlexWrapWrap
	return r
}

func (r *Row) Gutter() float32 { return r.gutter }

func (r *Row) SetGutter(g float32) {
	r.gutter = g
	r.style.Gap = g
}

func (r *Row) Draw(buf *render.CommandBuffer) {
	r.DrawChildren(buf)
}

// Col is a column within a Row. Its span determines its flex basis.
type Col struct {
	Base
	span   int // Out of 24 (like Ant Design grid)
	offset int
}

// NewCol creates a column with the given span (1-24).
func NewCol(tree *core.Tree, span int, cfg *Config) *Col {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	if span < 1 {
		span = 1
	}
	if span > 24 {
		span = 24
	}
	c := &Col{
		Base: NewBase(tree, core.TypeDiv, cfg),
		span: span,
	}
	pct := float32(span) / 24.0 * 100.0
	c.style.FlexBasis = layout.Pct(pct)
	c.style.MaxWidth = layout.Pct(pct)
	return c
}

func (c *Col) Span() int   { return c.span }
func (c *Col) Offset() int { return c.offset }

func (c *Col) SetSpan(span int) {
	if span < 1 {
		span = 1
	}
	if span > 24 {
		span = 24
	}
	c.span = span
	pct := float32(span) / 24.0 * 100.0
	c.style.FlexBasis = layout.Pct(pct)
	c.style.MaxWidth = layout.Pct(pct)
}

func (c *Col) SetOffset(offset int) {
	c.offset = offset
	if offset > 0 {
		pct := float32(offset) / 24.0 * 100.0
		c.style.Margin.Left = layout.Pct(pct)
	} else {
		c.style.Margin.Left = layout.Zero
	}
}

func (c *Col) Draw(buf *render.CommandBuffer) {
	c.DrawChildren(buf)
}

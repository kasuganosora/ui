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

func (r *Row) SetViewport(width float32) {
	for _, child := range r.Children() {
		col, ok := child.(*Col)
		if !ok {
			continue
		}
		col.viewport = width
		col.hasViewport = true
		col.applySpan()
	}
}

// Col is a column within a Row. Its span determines its flex basis.
type Col struct {
	Base
	span        int // Out of 24 (like Ant Design grid)
	offset      int
	spans       map[string]int
	viewport    float32
	hasViewport bool
}

var gridBreakpoints = []struct {
	name string
	min  float32
}{
	{"xs", 0},
	{"sm", 576},
	{"md", 768},
	{"lg", 992},
	{"xl", 1200},
	{"xxl", 1600},
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

func (c *Col) Span() int { return c.effectiveSpan() }

func (c *Col) SetSpanAt(breakpoint string, span int) {
	if c.spans == nil {
		c.spans = map[string]int{}
	}
	if span < 1 {
		span = 1
	}
	if span > 24 {
		span = 24
	}
	c.spans[breakpoint] = span
	if c.hasViewport {
		c.applySpan()
	}
}

func (c *Col) effectiveSpan() int {
	if !c.hasViewport {
		return c.span
	}
	best := c.span
	bestMin := float32(-1)
	for _, bp := range gridBreakpoints {
		if c.viewport < bp.min {
			continue
		}
		span, ok := c.spans[bp.name]
		if !ok || bp.min < bestMin {
			continue
		}
		best = span
		bestMin = bp.min
	}
	return best
}

func (c *Col) applySpan() {
	pct := float32(c.effectiveSpan()) / 24.0 * 100.0
	c.style.FlexBasis = layout.Pct(pct)
	c.style.MaxWidth = layout.Pct(pct)
}
func (c *Col) Offset() int { return c.offset }

func (c *Col) SetSpan(span int) {
	if span < 1 {
		span = 1
	}
	if span > 24 {
		span = 24
	}
	c.span = span
	c.applySpan()
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

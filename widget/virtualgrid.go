package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// VirtualGrid renders only visible cells in a large grid.
type VirtualGrid struct {
	Base
	cols      int
	rowCount  int
	cellW     float32
	cellH     float32
	gap       float32
	scrollY   float32
	renderCell func(buf *render.CommandBuffer, row, col int, bounds uimath.Rect)
}

func NewVirtualGrid(tree *core.Tree, cols, rowCount int, cfg *Config) *VirtualGrid {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &VirtualGrid{
		Base:     NewBase(tree, core.TypeCustom, cfg),
		cols:     cols,
		rowCount: rowCount,
		cellW:    100,
		cellH:    100,
		gap:      8,
	}
}

func (vg *VirtualGrid) Cols() int              { return vg.cols }
func (vg *VirtualGrid) RowCount() int          { return vg.rowCount }
func (vg *VirtualGrid) SetCols(c int)          { vg.cols = c }
func (vg *VirtualGrid) SetRowCount(r int)      { vg.rowCount = r }
func (vg *VirtualGrid) SetCellSize(w, h float32) { vg.cellW = w; vg.cellH = h }
func (vg *VirtualGrid) SetGap(g float32)       { vg.gap = g }
func (vg *VirtualGrid) SetScrollY(y float32)   { vg.scrollY = y }
func (vg *VirtualGrid) ScrollY() float32       { return vg.scrollY }

func (vg *VirtualGrid) SetRenderCell(fn func(*render.CommandBuffer, int, int, uimath.Rect)) {
	vg.renderCell = fn
}

func (vg *VirtualGrid) TotalHeight() float32 {
	return float32(vg.rowCount) * (vg.cellH + vg.gap)
}

func (vg *VirtualGrid) Draw(buf *render.CommandBuffer) {
	bounds := vg.Bounds()
	if bounds.IsEmpty() || vg.cols <= 0 {
		return
	}

	rowStride := vg.cellH + vg.gap
	startRow := int(vg.scrollY / rowStride)
	if startRow < 0 {
		startRow = 0
	}
	visibleRows := int(bounds.Height/rowStride) + 2

	for r := startRow; r < startRow+visibleRows && r < vg.rowCount; r++ {
		for c := 0; c < vg.cols; c++ {
			cx := bounds.X + float32(c)*(vg.cellW+vg.gap)
			cy := bounds.Y + float32(r)*rowStride - vg.scrollY
			if cy+vg.cellH < bounds.Y || cy > bounds.Y+bounds.Height {
				continue
			}
			cellBounds := uimath.NewRect(cx, cy, vg.cellW, vg.cellH)
			if vg.renderCell != nil {
				vg.renderCell(buf, r, c, cellBounds)
			} else {
				// Default placeholder
				buf.DrawRect(render.RectCmd{
					Bounds:      cellBounds,
					FillColor:   uimath.RGBA(0, 0, 0, 0.02),
					BorderColor: uimath.RGBA(0, 0, 0, 0.06),
					BorderWidth: 1,
				}, 1, 1)
			}
		}
	}
}

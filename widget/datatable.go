package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// DataTable is a windowed table with reorderable columns, resizable widths,
// and columns pinned to the left and right edges. Rows stay [][]string, in the
// same order as the columns. It does not draw rows outside the visible window.
type DataTable struct {
	Base
	columns []TableColumn
	rows    [][]string
	rowH    float32
	headerH float32
	frozenL int
	frozenR int
	scrollX float32
	scrollY float32
	onCell  func(CellEventContext)
	resize  int
	resizeX float32
	resizeW float32
}

type columnPlace struct {
	index  int
	x, w   float32
	frozen bool
}

// NewDataTable creates an empty table with a copy of columns.
func NewDataTable(tree *core.Tree, columns []TableColumn, cfg *Config) *DataTable {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	cols := make([]TableColumn, len(columns))
	copy(cols, columns)
	d := &DataTable{
		Base:    NewBase(tree, core.TypeCustom, cfg),
		columns: cols,
		rowH:    28,
		headerH: 28,
		resize:  -1,
	}
	tree.AddHandler(d.id, event.MouseDown, func(e *event.Event) {
		if e.Phase == event.PhaseCapture {
			return
		}
		col, edge := d.columnAt(e.GlobalX, e.GlobalY)
		if col >= 0 && edge && e.GlobalY < d.Bounds().Y+d.headerH {
			d.resize = col
			d.resizeX = e.GlobalX
			d.resizeW = d.columns[col].Width
		}
	})
	tree.AddHandler(d.id, event.MouseMove, func(e *event.Event) {
		if d.resize < 0 {
			return
		}
		w := d.resizeW + e.GlobalX - d.resizeX
		if w < 48 {
			w = 48
		}
		d.columns[d.resize].Width = w
	})
	tree.AddHandler(d.id, event.MouseUp, func(*event.Event) {
		d.resize = -1
	})
	tree.AddHandler(d.id, event.MouseClick, func(e *event.Event) {
		if e.Phase == event.PhaseCapture || d.onCell == nil {
			return
		}
		b := d.Bounds()
		if e.GlobalY < b.Y+d.headerH {
			return
		}
		row, col, ok := d.cellAt(e.GlobalX, e.GlobalY)
		if ok {
			d.onCell(CellEventContext{RowIndex: row, ColIndex: col})
		}
	})
	tree.AddHandler(d.id, event.MouseWheel, func(e *event.Event) {
		if e.WheelDY == 0 {
			return
		}
		step := d.rowH
		if e.WheelDY < 0 {
			step = -step
		}
		d.SetScrollY(d.scrollY + step)
	})
	return d
}

func (d *DataTable) SetRows(rows [][]string) {
	d.rows = make([][]string, len(rows))
	for i, row := range rows {
		d.rows[i] = make([]string, len(row))
		copy(d.rows[i], row)
	}
}

func (d *DataTable) Rows() [][]string { return d.rows }

func (d *DataTable) ClearRows() { d.rows = nil }

func (d *DataTable) SetRowHeight(h float32) {
	if h < 1 {
		h = 1
	}
	d.rowH = h
}

func (d *DataTable) SetHeaderHeight(h float32) {
	if h < 1 {
		h = 1
	}
	d.headerH = h
}

func (d *DataTable) OnCellClick(fn func(CellEventContext)) { d.onCell = fn }

func (d *DataTable) ColumnKeys() []string {
	out := make([]string, len(d.columns))
	for i, col := range d.columns {
		out[i] = col.ColKey
		if out[i] == "" {
			out[i] = col.Title
		}
	}
	return out
}

func (d *DataTable) ColumnWidth(i int) float32 {
	if i < 0 || i >= len(d.columns) {
		return 0
	}
	return d.columns[i].Width
}

// MoveColumn removes the column at from and inserts it at to in the list
// that remains. Cells in every row move with the column.
func (d *DataTable) MoveColumn(from, to int) bool {
	n := len(d.columns)
	if from < 0 || to < 0 || from >= n || to >= n || from == to {
		return false
	}
	col := d.columns[from]
	next := make([]TableColumn, 0, len(d.columns)-1)
	next = append(next, d.columns[:from]...)
	next = append(next, d.columns[from+1:]...)
	if to > len(next) {
		to = len(next)
	}
	cols := make([]TableColumn, 0, len(d.columns))
	cols = append(cols, next[:to]...)
	cols = append(cols, col)
	cols = append(cols, next[to:]...)
	d.columns = cols
	for i, row := range d.rows {
		d.rows[i] = moveCell(row, from, to)
	}
	return true
}

func moveCell(row []string, from, to int) []string {
	cell := ""
	if from >= 0 && from < len(row) {
		cell = row[from]
	}
	next := make([]string, 0, len(row))
	for i, v := range row {
		if i != from {
			next = append(next, v)
		}
	}
	if to > len(next) {
		to = len(next)
	}
	out := make([]string, 0, len(next)+1)
	out = append(out, next[:to]...)
	out = append(out, cell)
	out = append(out, next[to:]...)
	return out
}

// SetFrozen pins the first left and last right columns. They stay at the
// widget edges while SetScrollX moves the columns between them.
func (d *DataTable) SetFrozen(left, right int) {
	if left < 0 {
		left = 0
	}
	if right < 0 {
		right = 0
	}
	if left+right > len(d.columns) {
		right = len(d.columns) - left
		if right < 0 {
			left = len(d.columns)
			right = 0
		}
	}
	d.frozenL = left
	d.frozenR = right
	d.clampScroll()
}

func (d *DataTable) SetScrollX(x float32) {
	if x < 0 {
		x = 0
	}
	d.scrollX = x
	d.clampScroll()
}

func (d *DataTable) SetScrollY(y float32) {
	if y < 0 {
		y = 0
	}
	max := d.maxScrollY()
	if y > max {
		y = max
	}
	d.scrollY = y
}

func (d *DataTable) ColumnHeader(i int) uimath.Rect {
	for _, p := range d.placed() {
		if p.index == i {
			return uimath.NewRect(p.x, d.Bounds().Y, p.w, d.headerH)
		}
	}
	return uimath.Rect{}
}

func (d *DataTable) maxScrollY() float32 {
	view := d.Bounds().Height - d.headerH
	total := float32(len(d.rows)) * d.rowH
	if total < view {
		return 0
	}
	return total - view
}

func (d *DataTable) clampScroll() {
	_, _, sum, view := d.middleMetrics()
	max := sum - view
	if max < 0 {
		max = 0
	}
	if d.scrollX > max {
		d.scrollX = max
	}
	if d.scrollX < 0 {
		d.scrollX = 0
	}
}

func (d *DataTable) middleMetrics() (start, n int, sum, view float32) {
	nCols := len(d.columns)
	left, right := d.frozenL, d.frozenR
	if left+right > nCols {
		right = nCols - left
	}
	start = left
	n = nCols - left - right
	b := d.Bounds()
	var leftW, rightW float32
	for i := 0; i < left && i < nCols; i++ {
		leftW += d.columns[i].Width
	}
	for i := nCols - right; i < nCols && i >= 0; i++ {
		rightW += d.columns[i].Width
	}
	view = b.Width - leftW - rightW
	if view < 0 {
		view = 0
	}
	for i := 0; i < n; i++ {
		sum += d.columns[start+i].Width
	}
	return start, n, sum, view
}

func (d *DataTable) placed() []columnPlace {
	b := d.Bounds()
	n := len(d.columns)
	leftN, rightN := d.frozenL, d.frozenR
	if leftN+rightN > n {
		rightN = n - leftN
	}
	out := make([]columnPlace, 0, n)
	x := b.X
	for i := 0; i < leftN; i++ {
		w := d.columns[i].Width
		out = append(out, columnPlace{index: i, x: x, w: w, frozen: true})
		x += w
	}
	leftEdge := x
	var rightW float32
	for i := n - rightN; i < n && i >= 0; i++ {
		rightW += d.columns[i].Width
	}
	rightEdge := b.X + b.Width - rightW
	x = leftEdge - d.scrollX
	for i := leftN; i < n-rightN; i++ {
		w := d.columns[i].Width
		out = append(out, columnPlace{index: i, x: x, w: w})
		x += w
	}
	x = rightEdge
	for i := n - rightN; i < n && i >= 0; i++ {
		w := d.columns[i].Width
		out = append(out, columnPlace{index: i, x: x, w: w, frozen: true})
		x += w
	}
	return out
}

func (d *DataTable) middleClip() uimath.Rect {
	b := d.Bounds()
	var leftW, rightW float32
	n := len(d.columns)
	leftN, rightN := d.frozenL, d.frozenR
	if leftN+rightN > n {
		rightN = n - leftN
	}
	for i := 0; i < leftN && i < n; i++ {
		leftW += d.columns[i].Width
	}
	for i := n - rightN; i < n && i >= 0; i++ {
		rightW += d.columns[i].Width
	}
	return uimath.NewRect(b.X+leftW, b.Y, b.Width-leftW-rightW, b.Height)
}

func (d *DataTable) columnAt(x, y float32) (int, bool) {
	b := d.Bounds()
	if !b.Contains(uimath.Vec2{X: x, Y: y}) {
		return -1, false
	}
	clip := d.middleClip()
	for _, p := range d.placed() {
		if !p.frozen && (x < clip.X || x >= clip.X+clip.Width) {
			continue
		}
		if x >= p.x && x < p.x+p.w {
			return p.index, x >= p.x+p.w-6
		}
	}
	return -1, false
}

func (d *DataTable) cellAt(x, y float32) (row, col int, ok bool) {
	b := d.Bounds()
	if !b.Contains(uimath.Vec2{X: x, Y: y}) || y < b.Y+d.headerH {
		return 0, 0, false
	}
	col, _ = d.columnAt(x, y)
	if col < 0 {
		return 0, 0, false
	}
	content := y - (b.Y + d.headerH) + d.scrollY
	if content < 0 {
		return 0, 0, false
	}
	row = int(content / d.rowH)
	if row < 0 || row >= len(d.rows) {
		return 0, 0, false
	}
	rowY := b.Y + d.headerH + float32(row)*d.rowH - d.scrollY
	visTop := rowY
	if visTop < b.Y+d.headerH {
		visTop = b.Y + d.headerH
	}
	visBot := rowY + d.rowH
	if visBot > b.Y+b.Height {
		visBot = b.Y + b.Height
	}
	if y < visTop || y >= visBot {
		return 0, 0, false
	}
	return row, col, true
}

func (d *DataTable) Draw(buf *render.CommandBuffer) {
	b := d.Bounds()
	if b.IsEmpty() || len(d.columns) == 0 {
		return
	}
	places := d.placed()
	mid := d.middleClip()
	headerBand := uimath.NewRect(b.X, b.Y, b.Width, d.headerH)
	bodyH := b.Height - d.headerH
	if bodyH < 0 {
		bodyH = 0
	}
	bodyBand := uimath.NewRect(b.X, b.Y+d.headerH, b.Width, bodyH)
	regions := []uimath.Rect{
		uimath.NewRect(b.X, b.Y, mid.X-b.X, b.Height),
		uimath.NewRect(mid.X, b.Y, mid.Width, b.Height),
		uimath.NewRect(mid.X+mid.Width, b.Y, b.X+b.Width-(mid.X+mid.Width), b.Height),
	}
	buf.PushClipZ(b, 1)
	for _, region := range regions {
		if region.Width <= 0 {
			continue
		}
		buf.PushClip(headerBand.Intersection(region))
		for _, p := range places {
			if !columnInRegion(p, region) {
				continue
			}
			fill := uimath.ColorWhite
			if p.frozen {
				fill = d.config.SuccessColor
			}
			d.paintHeader(buf, p, fill)
		}
		buf.PopClip()
		if bodyBand.Height <= 0 {
			continue
		}
		buf.PushClip(bodyBand.Intersection(region))
		for _, p := range places {
			if columnInRegion(p, region) {
				d.paintBody(buf, p)
			}
		}
		buf.PopClip()
	}
	buf.PopClipZ(3)
}

func columnInRegion(p columnPlace, region uimath.Rect) bool {
	return p.x < region.X+region.Width && p.x+p.w > region.X
}

func (d *DataTable) paintHeader(buf *render.CommandBuffer, p columnPlace, fill uimath.Color) {
	b := d.Bounds()
	col := d.columns[p.index]
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(p.x, b.Y, p.w, d.headerH),
		FillColor: fill,
	}, 1, 1)
	d.paintString(buf, col.Title, uimath.NewRect(p.x, b.Y, p.w, d.headerH), col.Align)
}

func (d *DataTable) paintBody(buf *render.CommandBuffer, p columnPlace) {
	b := d.Bounds()
	bodyTop := b.Y + d.headerH
	bodyBot := b.Y + b.Height
	if d.rowH <= 0 || bodyBot <= bodyTop {
		return
	}
	col := d.columns[p.index]
	start := int(d.scrollY / d.rowH)
	if start < 0 {
		start = 0
	}
	for ri := start; ri < len(d.rows); ri++ {
		y := bodyTop + float32(ri)*d.rowH - d.scrollY
		if y >= bodyBot {
			break
		}
		if y+d.rowH <= bodyTop {
			continue
		}
		fill := uimath.ColorWhite
		if ri%2 == 1 {
			fill = d.config.BorderColor
		}
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(p.x, y, p.w, d.rowH),
			FillColor: fill,
		}, 1, 1)
		cell := ""
		if p.index < len(d.rows[ri]) {
			cell = d.rows[ri][p.index]
		}
		d.paintString(buf, cell, uimath.NewRect(p.x, y, p.w, d.rowH), col.Align)
	}
}

func (d *DataTable) paintString(buf *render.CommandBuffer, text string, cell uimath.Rect, align TableColumnAlign) {
	cfg := d.config
	if cfg == nil || cfg.TextRenderer == nil || text == "" || cell.Width < 4 || cell.Height < 4 {
		return
	}
	pad := cfg.SpaceSM
	if pad*2+4 >= cell.Width {
		pad = 2
	}
	size := cfg.FontSizeSm
	if size <= 0 {
		size = cfg.FontSize
	}
	maxW := cell.Width - pad*2
	tw := cfg.TextRenderer.MeasureText(text, size)
	if tw > maxW {
		tw = maxW
	}
	tx := cell.X + pad
	switch align {
	case TableAlignCenter:
		tx = cell.X + (cell.Width-tw)/2
	case TableAlignRight:
		tx = cell.X + cell.Width - pad - tw
	}
	lh := cfg.TextRenderer.LineHeight(size)
	ty := cell.Y + (cell.Height-lh)/2
	if ty < cell.Y {
		ty = cell.Y
	}
	buf.PushClip(cell)
	emitText(cfg, buf, text, tx, ty, size, maxW, cfg.TextColor, 1, 4)
	buf.PopClip()
}

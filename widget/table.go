package widget

import (
	"sort"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// TableColumnAlign represents column alignment.
type TableColumnAlign int

const (
	TableAlignLeft TableColumnAlign = iota // default
	TableAlignCenter
	TableAlignRight
)

// TableColumnFixed represents column fixed position.
type TableColumnFixed int

const (
	TableFixedNone TableColumnFixed = iota
	TableFixedLeft
	TableFixedRight
)

// TableColumn defines a column in the table.
type TableColumn struct {
	ColKey   string           // column key for data mapping
	Title    string           // column header text
	Width    float32          // column width, 0 = auto
	Align    TableColumnAlign // horizontal alignment
	Fixed    TableColumnFixed // fixed column position
	Ellipsis bool             // text overflow ellipsis
	Sorter   bool             // whether column is sortable
}

// RowEventContext provides context for row events.
type RowEventContext struct {
	RowIndex int
}

// CellEventContext provides context for cell events.
type CellEventContext struct {
	RowIndex int
	ColIndex int
}

// Table is a data table with column headers and rows.
type Table struct {
	Base
	columns    []TableColumn
	rows       [][]string
	rowHeight  float32
	headerH    float32
	stripe     bool   // striped rows
	bordered   bool   // show borders
	hover      bool   // show hover state
	size       Size   // table size
	rowKey     string // field name for unique row id
	hoveredRow int    // -1 = none

	onRowClick   func(ctx RowEventContext)  // called on row click
	onCellClick  func(ctx CellEventContext) // called on cell click
	onPageChange func(pageInfo PageInfo)    // called on page change
	onSortChange func(sort SortInfo)        // called on sort change

	sortKey  string
	sortDesc bool
	pageSize int
	page     int
	scrollY  float32
}

// SortInfo describes a column sort state.
type SortInfo struct {
	SortBy     string // column key
	Descending bool
}

// PageInfo describes pagination state for table events.
type PageInfo struct {
	Current  int
	Previous int
	PageSize int
}

func NewTable(tree *core.Tree, columns []TableColumn, cfg *Config) *Table {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	cols := make([]TableColumn, len(columns))
	copy(cols, columns)
	t := &Table{
		Base:       NewBase(tree, core.TypeDiv, cfg),
		columns:    cols,
		rowHeight:  48,
		headerH:    48,
		stripe:     true,
		bordered:   true,
		hover:      true,
		size:       SizeMedium,
		rowKey:     "id",
		hoveredRow: -1,
	}

	tree.AddHandler(t.id, event.MouseMove, func(e *event.Event) {
		t.updateHover(e.GlobalY)
	})
	tree.AddHandler(t.id, event.MouseLeave, func(e *event.Event) {
		if t.hoveredRow != -1 {
			t.hoveredRow = -1
			t.tree.MarkDirty(t.id)
		}
	})
	tree.AddHandler(t.id, event.MouseClick, func(e *event.Event) {
		t.handleClick(e.GlobalX, e.GlobalY)
	})

	return t
}

func (t *Table) Columns() []TableColumn { return t.columns }
func (t *Table) Rows() [][]string       { return t.rows }
func (t *Table) RowCount() int          { return len(t.rows) }
func (t *Table) SetStripe(s bool)       { t.stripe = s }
func (t *Table) SetBordered(b bool)     { t.bordered = b }
func (t *Table) SetHover(h bool)        { t.hover = h }
func (t *Table) SetRowHeight(h float32) { t.rowHeight = h }
func (t *Table) SetSize(s Size)         { t.size = s }
func (t *Table) SetRowKey(k string)     { t.rowKey = k }
func (t *Table) RowKey() string         { return t.rowKey }

// OnRowClick sets the callback for row click events.
func (t *Table) OnRowClick(fn func(ctx RowEventContext)) { t.onRowClick = fn }

// OnCellClick sets the callback for cell click events.
func (t *Table) OnCellClick(fn func(ctx CellEventContext)) { t.onCellClick = fn }

// OnPageChange sets the callback for page change events.
func (t *Table) OnPageChange(fn func(pageInfo PageInfo)) { t.onPageChange = fn }

// OnSortChange sets the callback for sort change events.
func (t *Table) OnSortChange(fn func(sort SortInfo)) { t.onSortChange = fn }

func (t *Table) HoveredRow() int { return t.hoveredRow }

func (t *Table) SetScrollY(y float32) {
	if y < 0 {
		y = 0
	}
	t.scrollY = y
}

func (t *Table) SetPageSize(n int) {
	if n < 0 {
		n = 0
	}
	t.pageSize = n
	t.page = 0
}

func (t *Table) NextPage() {
	if t.pageSize <= 0 || len(t.rows) == 0 {
		return
	}
	pages := (len(t.rows) + t.pageSize - 1) / t.pageSize
	if t.page+1 >= pages {
		return
	}
	prev := t.page
	t.page++
	if t.onPageChange != nil {
		t.onPageChange(PageInfo{Current: t.page, Previous: prev, PageSize: t.pageSize})
	}
}

func (t *Table) visibleSpan() (start, end int) {
	n := len(t.rows)
	if n == 0 {
		return 0, 0
	}
	if t.pageSize > 0 {
		start = t.page * t.pageSize
		if start > n {
			start = n
		}
		end = start + t.pageSize
		if end > n {
			end = n
		}
		return start, end
	}
	viewH := t.Bounds().Height - t.headerH
	if viewH < t.rowHeight {
		viewH = t.rowHeight
	}
	count := int(viewH/t.rowHeight) + 1
	if count < 1 {
		count = 1
	}
	start = int(t.scrollY / t.rowHeight)
	if start < 0 {
		start = 0
	}
	if start > n {
		start = n
	}
	end = start + count
	if end > n {
		end = n
	}
	return start, end
}

func (t *Table) sortAt(gx float32) {
	widths := t.colWidths()
	x := t.Bounds().X
	col := -1
	for i, w := range widths {
		if gx >= x && gx < x+w {
			col = i
			break
		}
		x += w
	}
	if col < 0 || col >= len(t.columns) || !t.columns[col].Sorter {
		return
	}
	key := t.columns[col].ColKey
	if key == "" {
		key = t.columns[col].Title
	}
	if t.sortKey == key {
		t.sortDesc = !t.sortDesc
	} else {
		t.sortKey = key
		t.sortDesc = false
	}
	sort.SliceStable(t.rows, func(i, j int) bool {
		ai, aj := "", ""
		if col < len(t.rows[i]) {
			ai = t.rows[i][col]
		}
		if col < len(t.rows[j]) {
			aj = t.rows[j][col]
		}
		if t.sortDesc {
			return ai > aj
		}
		return ai < aj
	})
	if t.onSortChange != nil {
		t.onSortChange(SortInfo{SortBy: key, Descending: t.sortDesc})
	}
}

func (t *Table) TotalHeight() float32 {
	return t.headerH + float32(len(t.rows))*t.rowHeight
}

func (t *Table) SetRows(rows [][]string) {
	t.rows = make([][]string, len(rows))
	for i, row := range rows {
		t.rows[i] = make([]string, len(row))
		copy(t.rows[i], row)
	}
}

func (t *Table) AddRow(row []string) {
	r := make([]string, len(row))
	copy(r, row)
	t.rows = append(t.rows, r)
}

func (t *Table) ClearRows() {
	t.rows = t.rows[:0]
}

func (t *Table) handleClick(gx, gy float32) {
	bounds := t.Bounds()
	if gy < bounds.Y+t.headerH {
		if gy >= bounds.Y {
			t.sortAt(gx)
		}
		return
	}
	start, end := t.visibleSpan()
	local := int((gy - bounds.Y - t.headerH) / t.rowHeight)
	rowIdx := start + local
	if rowIdx < start || rowIdx >= end || rowIdx >= len(t.rows) {
		return
	}
	if t.onRowClick != nil {
		t.onRowClick(RowEventContext{RowIndex: rowIdx})
	}
	if t.onCellClick != nil {
		// Determine column
		colWidths := t.colWidths()
		rx := gx - bounds.X
		cx := float32(0)
		for ci, cw := range colWidths {
			if rx >= cx && rx < cx+cw {
				t.onCellClick(CellEventContext{RowIndex: rowIdx, ColIndex: ci})
				break
			}
			cx += cw
		}
	}
}

func (t *Table) updateHover(globalY float32) {
	if !t.hover {
		return
	}
	bounds := t.Bounds()
	ry := globalY - bounds.Y - t.headerH
	newHovered := -1
	if ry >= 0 {
		start, end := t.visibleSpan()
		idx := start + int(ry/t.rowHeight)
		if idx >= start && idx < end {
			newHovered = idx
		}
	}
	if newHovered != t.hoveredRow {
		t.hoveredRow = newHovered
		t.tree.MarkDirty(t.id)
	}
}

func (t *Table) colWidths() []float32 {
	numCols := len(t.columns)
	widths := make([]float32, numCols)
	bounds := t.Bounds()
	totalFixed := float32(0)
	autoCount := 0
	for i, col := range t.columns {
		if col.Width > 0 {
			widths[i] = col.Width
			totalFixed += col.Width
		} else {
			autoCount++
		}
	}
	if autoCount > 0 {
		autoW := (bounds.Width - totalFixed) / float32(autoCount)
		if autoW < 40 {
			autoW = 40
		}
		for i, col := range t.columns {
			if col.Width <= 0 {
				widths[i] = autoW
			}
		}
	}
	return widths
}

func (t *Table) Draw(buf *render.CommandBuffer) {
	bounds := t.Bounds()
	if bounds.IsEmpty() || len(t.columns) == 0 {
		return
	}
	cfg := t.config
	colWidths := t.colWidths()
	totalH := t.TotalHeight()
	if totalH > bounds.Height {
		totalH = bounds.Height
	}

	// Outer border
	if t.bordered {
		buf.DrawRect(render.RectCmd{
			Bounds:      uimath.NewRect(bounds.X, bounds.Y, bounds.Width, totalH),
			BorderColor: cfg.BorderColor,
			BorderWidth: 1,
			Corners:     uimath.CornersAll(cfg.BorderRadius),
		}, 0, 1)
	}

	// --- Header ---
	headerBg := uimath.ColorHex("#f3f3f3")
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(bounds.X, bounds.Y, bounds.Width, t.headerH),
		FillColor: headerBg,
		Corners:   uimath.CornersAll(cfg.BorderRadius),
	}, 1, 1)
	// Flatten bottom corners of header (overlap with square rect)
	if t.headerH < totalH {
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(bounds.X, bounds.Y+t.headerH-cfg.BorderRadius, bounds.Width, cfg.BorderRadius),
			FillColor: headerBg,
		}, 1, 1)
	}

	// Header text
	headerTextColor := uimath.ColorHex("#8b8b8b")
	cx := bounds.X
	for i, col := range t.columns {
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			tw := cfg.TextRenderer.MeasureText(col.Title, cfg.FontSizeSm)
			tx := cx + cfg.SpaceMD
			switch col.Align {
			case TableAlignCenter:
				tx = cx + (colWidths[i]-tw)/2
			case TableAlignRight:
				tx = cx + colWidths[i] - tw - cfg.SpaceMD
			}
			cfg.TextRenderer.DrawText(buf, col.Title, tx, bounds.Y+(t.headerH-lh)/2, cfg.FontSizeSm, colWidths[i]-cfg.SpaceMD*2, headerTextColor, 1)
		}

		// Vertical column divider in header (bordered mode)
		if t.bordered && i < len(t.columns)-1 {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(cx+colWidths[i]-0.5, bounds.Y+8, 1, t.headerH-16),
				FillColor: uimath.ColorHex("#e8e8e8"),
			}, 2, 1)
		}
		cx += colWidths[i]
	}

	// Header bottom border
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(bounds.X, bounds.Y+t.headerH-1, bounds.Width, 1),
		FillColor: cfg.BorderColor,
	}, 2, 1)

	// --- Data Rows (only the visible window) ---
	start, end := t.visibleSpan()
	for ri := start; ri < end; ri++ {
		row := t.rows[ri]
		local := ri - start
		ry := bounds.Y + t.headerH + float32(local)*t.rowHeight
		if t.pageSize == 0 && ry >= bounds.Y+bounds.Height {
			break
		}

		// Hover background
		if t.hover && ri == t.hoveredRow {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X+1, ry, bounds.Width-2, t.rowHeight),
				FillColor: uimath.ColorHex("#f3f3f3"),
			}, 1, 1)
		} else if t.stripe && ri%2 == 1 {
			// Striped background
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X+1, ry, bounds.Width-2, t.rowHeight),
				FillColor: uimath.ColorHex("#f3f3f3"),
			}, 1, 1)
		}

		// Row bottom divider
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(bounds.X, ry+t.rowHeight-1, bounds.Width, 1),
			FillColor: uimath.ColorHex("#e8e8e8"),
		}, 2, 1)

		// Cell text
		cx = bounds.X
		for ci, cell := range row {
			if ci >= len(t.columns) {
				break
			}
			col := t.columns[ci]
			if cfg.TextRenderer != nil {
				lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
				tw := cfg.TextRenderer.MeasureText(cell, cfg.FontSize)
				tx := cx + cfg.SpaceMD
				switch col.Align {
				case TableAlignCenter:
					tx = cx + (colWidths[ci]-tw)/2
				case TableAlignRight:
					tx = cx + colWidths[ci] - tw - cfg.SpaceMD
				}
				cfg.TextRenderer.DrawText(buf, cell, tx, ry+(t.rowHeight-lh)/2, cfg.FontSize, colWidths[ci]-cfg.SpaceMD*2, cfg.TextColor, 1)
			}

			// Vertical column divider (bordered mode)
			if t.bordered && ci < len(t.columns)-1 {
				buf.DrawRect(render.RectCmd{
					Bounds:    uimath.NewRect(cx+colWidths[ci]-0.5, ry, 1, t.rowHeight),
					FillColor: uimath.ColorHex("#e8e8e8"),
				}, 2, 1)
			}
			cx += colWidths[ci]
		}
	}
}

package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// TableColumn defines a column in the table.
type TableColumn struct {
	Title string
	Width float32 // 0 = auto (equal share)
	Align uint8   // 0=left, 1=center, 2=right
}

// Table is a data table with column headers and rows.
type Table struct {
	Base
	columns   []TableColumn
	rows      [][]string
	rowHeight float32
	headerH   float32
	striped   bool
	bordered  bool
}

func NewTable(tree *core.Tree, columns []TableColumn, cfg *Config) *Table {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	cols := make([]TableColumn, len(columns))
	copy(cols, columns)
	return &Table{
		Base:      NewBase(tree, core.TypeCustom, cfg),
		columns:   cols,
		rowHeight: 40,
		headerH:   44,
		striped:   true,
		bordered:  true,
	}
}

func (t *Table) Columns() []TableColumn { return t.columns }
func (t *Table) Rows() [][]string       { return t.rows }
func (t *Table) RowCount() int          { return len(t.rows) }
func (t *Table) SetStriped(s bool)      { t.striped = s }
func (t *Table) SetBordered(b bool)     { t.bordered = b }
func (t *Table) SetRowHeight(h float32) { t.rowHeight = h }

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

func (t *Table) Draw(buf *render.CommandBuffer) {
	bounds := t.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := t.config
	numCols := len(t.columns)
	if numCols == 0 {
		return
	}

	// Calculate column widths
	colWidths := make([]float32, numCols)
	totalFixed := float32(0)
	autoCount := 0
	for i, col := range t.columns {
		if col.Width > 0 {
			colWidths[i] = col.Width
			totalFixed += col.Width
		} else {
			autoCount++
		}
	}
	autoW := float32(0)
	if autoCount > 0 {
		autoW = (bounds.Width - totalFixed) / float32(autoCount)
	}
	for i, col := range t.columns {
		if col.Width <= 0 {
			colWidths[i] = autoW
		}
	}

	// Header background
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(bounds.X, bounds.Y, bounds.Width, t.headerH),
		FillColor: uimath.RGBA(0, 0, 0, 0.02),
	}, 1, 1)

	// Header text
	cx := bounds.X
	for i, col := range t.columns {
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			cfg.TextRenderer.DrawText(buf, col.Title, cx+cfg.SpaceSM, bounds.Y+(t.headerH-lh)/2, cfg.FontSize, colWidths[i]-cfg.SpaceSM*2, cfg.TextColor, 1)
		}
		cx += colWidths[i]
	}

	// Header divider
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(bounds.X, bounds.Y+t.headerH-1, bounds.Width, 1),
		FillColor: cfg.BorderColor,
	}, 1, 1)

	// Rows
	for ri, row := range t.rows {
		ry := bounds.Y + t.headerH + float32(ri)*t.rowHeight
		if ry+t.rowHeight > bounds.Y+bounds.Height {
			break
		}

		// Striped background
		if t.striped && ri%2 == 1 {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X, ry, bounds.Width, t.rowHeight),
				FillColor: uimath.RGBA(0, 0, 0, 0.015),
			}, 1, 1)
		}

		// Row divider
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(bounds.X, ry+t.rowHeight-1, bounds.Width, 1),
			FillColor: uimath.RGBA(0, 0, 0, 0.04),
		}, 1, 1)

		// Cell text
		cx = bounds.X
		for ci, cell := range row {
			if ci >= numCols {
				break
			}
			if cfg.TextRenderer != nil {
				lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
				cfg.TextRenderer.DrawText(buf, cell, cx+cfg.SpaceSM, ry+(t.rowHeight-lh)/2, cfg.FontSize, colWidths[ci]-cfg.SpaceSM*2, cfg.TextColor, 1)
			}
			cx += colWidths[ci]
		}
	}

	// Outer border
	if t.bordered {
		totalH := t.headerH + float32(len(t.rows))*t.rowHeight
		if totalH > bounds.Height {
			totalH = bounds.Height
		}
		buf.DrawRect(render.RectCmd{
			Bounds:      uimath.NewRect(bounds.X, bounds.Y, bounds.Width, totalH),
			BorderColor: cfg.BorderColor,
			BorderWidth: 1,
		}, 1, 1)
	}
}

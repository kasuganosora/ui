package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// DiffRow is one side-by-side line. Op is "same", "del", "add", or "change".
// Columns are rune offsets in Left or Right, shared with Editor's TextPos.
type DiffRow struct {
	Op          string
	Left, Right string
}

// DiffView shows a line diff in two columns. It is not a word-level or
// three-way merge.
type DiffView struct {
	Base
	rows   []DiffRow
	index  int
	caret  TextPos
	scroll float32
	rowH   float32
}

// NewDiffView creates an empty side-by-side diff.
func NewDiffView(tree *core.Tree, cfg *Config) *DiffView {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	d := &DiffView{Base: NewBase(tree, core.TypeCustom, cfg), rowH: 22}
	tree.SetFocusable(d.id, true)
	tree.AddHandler(d.id, event.KeyDown, func(ev *event.Event) {
		if !targetPhase(ev) {
			return
		}
		switch ev.Key {
		case event.KeyN, event.KeyArrowDown:
			d.NextChange(1)
		case event.KeyP, event.KeyArrowUp:
			d.NextChange(-1)
		default:
			return
		}
		ev.StopPropagation()
	})
	tree.AddHandler(d.id, event.MouseClick, func(ev *event.Event) {
		if !targetPhase(ev) {
			return
		}
		d.click(ev.GlobalX, ev.GlobalY)
	})
	return d
}

// SetTexts replaces both sides and resets the selection to the first row.
func (d *DiffView) SetTexts(left, right string) {
	d.rows = diffLines(textLines(left), textLines(right))
	d.index = 0
	d.caret = TextPos{}
	d.scroll = 0
}

func (d *DiffView) SelectedIndex() int {
	if len(d.rows) == 0 {
		return 0
	}
	if d.index < 0 {
		return 0
	}
	if d.index >= len(d.rows) {
		return len(d.rows) - 1
	}
	return d.index
}

func (d *DiffView) SelectedOp() string {
	if len(d.rows) == 0 {
		return ""
	}
	return d.rows[d.SelectedIndex()].Op
}

func (d *DiffView) LeftText() string {
	if len(d.rows) == 0 {
		return ""
	}
	return d.rows[d.SelectedIndex()].Left
}

func (d *DiffView) RightText() string {
	if len(d.rows) == 0 {
		return ""
	}
	return d.rows[d.SelectedIndex()].Right
}

func (d *DiffView) Caret() TextPos { return d.caret }

// NextChange moves to the next row whose op is not "same".
func (d *DiffView) NextChange(dir int) {
	if len(d.rows) == 0 || dir == 0 {
		return
	}
	i := d.index
	for {
		i += dir
		if i < 0 || i >= len(d.rows) {
			return
		}
		if d.rows[i].Op != "same" {
			d.index = i
			d.caret = TextPos{Line: i}
			d.reveal()
			return
		}
	}
}

// RowScreen is the full-width row in screen space.
func (d *DiffView) RowScreen(i int) uimath.Rect {
	b := d.Bounds()
	return uimath.NewRect(b.X, b.Y+float32(i)*d.rowH-d.scroll, b.Width, d.rowH)
}

func (d *DiffView) reveal() {
	b := d.Bounds()
	y := float32(d.index) * d.rowH
	if y < d.scroll {
		d.scroll = y
	}
	if b.Height > 0 && y+d.rowH > d.scroll+b.Height {
		d.scroll = y + d.rowH - b.Height
	}
	if d.scroll < 0 {
		d.scroll = 0
	}
}

func (d *DiffView) sides() (left, right uimath.Rect) {
	b := d.Bounds()
	w := b.Width / 2
	return uimath.NewRect(b.X, b.Y, w, b.Height), uimath.NewRect(b.X+w, b.Y, b.Width-w, b.Height)
}

func (d *DiffView) click(x, y float32) {
	b := d.Bounds()
	if !b.Contains(uimath.Vec2{X: x, Y: y}) || d.rowH <= 0 {
		return
	}
	i := int((y - b.Y + d.scroll) / d.rowH)
	if i < 0 || i >= len(d.rows) {
		return
	}
	d.index = i
	left, right := d.sides()
	rowY := b.Y + float32(i)*d.rowH - d.scroll
	side := left
	text := d.rows[i].Left
	if x >= right.X {
		side = right
		text = d.rows[i].Right
	}
	local := x - (side.X + 6)
	d.caret = TextPos{Line: i, Col: colAtX(text, local, d.measure)}
	_ = rowY
}

func (d *DiffView) measure(s string) float32 {
	size := d.config.FontSizeSm
	if d.config.TextRenderer != nil {
		return d.config.TextRenderer.MeasureText(s, size)
	}
	return float32(len([]rune(s))) * size * 0.55
}

func (d *DiffView) Draw(buf *render.CommandBuffer) {
	b := d.Bounds()
	if b.IsEmpty() {
		return
	}
	cfg := d.config
	left, right := d.sides()
	buf.PushClip(b)
	for i, row := range d.rows {
		y := b.Y + float32(i)*d.rowH - d.scroll
		if y+d.rowH <= b.Y {
			continue
		}
		if y >= b.Y+b.Height {
			break
		}
		d.paintSide(buf, left, y, row.Left, sideFill(cfg, row.Op, true), i == d.index)
		d.paintSide(buf, right, y, row.Right, sideFill(cfg, row.Op, false), i == d.index)
	}
	buf.PopClip()
}

func sideFill(cfg *Config, op string, isLeft bool) uimath.Color {
	switch op {
	case "del":
		if isLeft {
			return cfg.ErrorColor
		}
	case "add":
		if !isLeft {
			return cfg.SuccessColor
		}
	case "change":
		return cfg.WarningColor
	}
	return uimath.ColorWhite
}

func (d *DiffView) paintSide(buf *render.CommandBuffer, side uimath.Rect, y float32, text string, fill uimath.Color, selected bool) {
	cfg := d.config
	cell := uimath.NewRect(side.X, y, side.Width, d.rowH)
	buf.PushClip(cell.Intersection(side))
	buf.DrawRect(render.RectCmd{Bounds: cell, FillColor: fill}, 1, 1)
	if selected {
		buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(side.X, y, 3, d.rowH), FillColor: cfg.PrimaryColor}, 2, 1)
	}
	if text != "" {
		emitText(cfg, buf, text, side.X+6, y+3, cfg.FontSizeSm, side.Width-10, cfg.TextColor, 1, 4)
	}
	buf.PopClip()
}

func diffLines(a, b []string) []DiffRow {
	n, m := len(a), len(b)
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}
	var ops []DiffRow
	i, j := 0, 0
	for i < n && j < m {
		if a[i] == b[j] {
			ops = append(ops, DiffRow{Op: "same", Left: a[i], Right: b[j]})
			i++
			j++
			continue
		}
		if dp[i+1][j] >= dp[i][j+1] {
			ops = append(ops, DiffRow{Op: "del", Left: a[i]})
			i++
			continue
		}
		ops = append(ops, DiffRow{Op: "add", Right: b[j]})
		j++
	}
	for i < n {
		ops = append(ops, DiffRow{Op: "del", Left: a[i]})
		i++
	}
	for j < m {
		ops = append(ops, DiffRow{Op: "add", Right: b[j]})
		j++
	}
	return pairChanges(ops)
}

func pairChanges(ops []DiffRow) []DiffRow {
	var out []DiffRow
	for i := 0; i < len(ops); i++ {
		if ops[i].Op == "del" {
			j := i
			var dels, adds []DiffRow
			for j < len(ops) && ops[j].Op == "del" {
				dels = append(dels, ops[j])
				j++
			}
			k := j
			for k < len(ops) && ops[k].Op == "add" {
				adds = append(adds, ops[k])
				k++
			}
			n := len(dels)
			if len(adds) < n {
				n = len(adds)
			}
			for t := 0; t < n; t++ {
				out = append(out, DiffRow{Op: "change", Left: dels[t].Left, Right: adds[t].Right})
			}
			for t := n; t < len(dels); t++ {
				out = append(out, dels[t])
			}
			for t := n; t < len(adds); t++ {
				out = append(out, adds[t])
			}
			i = k - 1
			continue
		}
		out = append(out, ops[i])
	}
	return out
}

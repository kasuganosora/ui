package widget

import (
	"strings"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// EmojiPicker searches the handwritten catalog and draws only the visible cells.
type EmojiPicker struct {
	Base
	items    []Emoji
	query    string
	category string
	onPick   func(Emoji)
	scroll   float32
	edge     ScrollEdge
	drawn    int
	cols     int
	cell     float32
}

func NewEmojiPicker(tree *core.Tree, cfg *Config) *EmojiPicker {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	e := &EmojiPicker{Base: NewBase(tree, core.TypeCustom, cfg), items: EmojiCatalog(), cols: 8, cell: 32}
	tree.SetFocusable(e.id, true)
	tree.AddHandler(e.id, event.KeyPress, func(ev *event.Event) {
		if !targetPhase(ev) || ev.Char == 0 {
			return
		}
		e.query += string(ev.Char)
		e.edge.Reset()
		e.scroll = 0
		ev.StopPropagation()
	})
	tree.AddHandler(e.id, event.KeyDown, func(ev *event.Event) {
		if !targetPhase(ev) {
			return
		}
		if ev.Key == event.KeyBackspace && e.query != "" {
			e.query = e.query[:len(e.query)-1]
			ev.StopPropagation()
		}
	})
	tree.AddHandler(e.id, event.MouseWheel, func(ev *event.Event) {
		e.HandleWheel(ev.WheelDY)
	})
	tree.AddHandler(e.id, event.MouseClick, func(ev *event.Event) {
		if !targetPhase(ev) {
			return
		}
		if item, ok := e.hit(ev.GlobalX, ev.GlobalY); ok && e.onPick != nil {
			e.onPick(item)
		}
	})
	return e
}

func (e *EmojiPicker) SetCategory(c string) {
	e.edge.Reset()
	e.category = c
	e.scroll = 0
}
func (e *EmojiPicker) SetQuery(q string) {
	e.edge.Reset()
	e.query = q
	e.scroll = 0
}

// HandleWheel is the user-gesture entry point for the emoji grid.
func (e *EmojiPicker) HandleWheel(dy float32) { e.applyScrollDelta(dy) }
func (e *EmojiPicker) ScrollY() float32       { return e.scroll }
func (e *EmojiPicker) SetScrollY(y float32)   { e.edge.Reset(); e.scroll = y }

func (e *EmojiPicker) maxScroll() float32 {
	if e.cols < 1 || e.cell <= 0 {
		return 0
	}
	n := len(e.Filtered())
	rows := n / e.cols
	if n%e.cols != 0 {
		rows++
	}
	max := float32(rows)*e.cell - e.Bounds().Height
	if max < 0 {
		return 0
	}
	return max
}

func (e *EmojiPicker) applyScrollDelta(delta float32) {
	e.scroll = applyScrollGesture(&e.edge, e.scroll, delta, 0, e.maxScroll(),
		e.Bounds().Height, e.config.Animate())
	e.tree.MarkDirty(e.id)
}

func (e *EmojiPicker) settleScroll() {
	if next, changed := settleScrollGesture(&e.edge, e.scroll, 0, e.maxScroll(), e.Bounds().Height); changed {
		e.scroll = next
		e.tree.MarkDirty(e.id)
	}
}
func (e *EmojiPicker) Query() string           { return e.query }
func (e *EmojiPicker) OnSelect(fn func(Emoji)) { e.onPick = fn }
func (e *EmojiPicker) DrawnCount() int         { return e.drawn }
func (e *EmojiPicker) Filtered() []Emoji {
	q := strings.ToLower(e.query)
	var out []Emoji
	for _, item := range e.items {
		if e.category != "" && item.Category != e.category {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(item.Label), q) && !strings.Contains(item.Char, e.query) {
			continue
		}
		out = append(out, item)
	}
	return out
}
func (e *EmojiPicker) visibleRange() (start, end int, items []Emoji) {
	items = e.Filtered()
	b := e.Bounds()
	if e.cols < 1 || e.cell <= 0 {
		return 0, 0, items
	}
	rows := len(items) / e.cols
	if len(items)%e.cols != 0 {
		rows++
	}
	firstRow := int(e.scroll / e.cell)
	if firstRow < 0 {
		firstRow = 0
	}
	visRows := int(b.Height/e.cell) + 1
	if visRows < 1 {
		visRows = 1
	}
	start = firstRow * e.cols
	end = (firstRow + visRows) * e.cols
	if start > len(items) {
		start = len(items)
	}
	if end > len(items) {
		end = len(items)
	}
	return start, end, items
}
func (e *EmojiPicker) hit(x, y float32) (Emoji, bool) {
	start, end, items := e.visibleRange()
	b := e.Bounds()
	for i := start; i < end; i++ {
		col := i % e.cols
		row := i / e.cols
		r := uimath.NewRect(b.X+float32(col)*e.cell, b.Y+float32(row)*e.cell-e.scroll, e.cell, e.cell)
		if r.Contains(uimath.Vec2{X: x, Y: y}) {
			return items[i], true
		}
	}
	return Emoji{}, false
}
func (e *EmojiPicker) Draw(buf *render.CommandBuffer) {
	b := e.Bounds()
	e.drawn = 0
	if b.IsEmpty() {
		return
	}
	e.settleScroll()
	buf.DrawRect(render.RectCmd{Bounds: b, FillColor: e.config.BgColor, BorderColor: e.config.BorderColor, BorderWidth: 1}, 1, 1)
	start, end, _ := e.visibleRange()
	for i := start; i < end; i++ {
		col := i % e.cols
		row := i / e.cols
		r := uimath.NewRect(b.X+float32(col)*e.cell+2, b.Y+float32(row)*e.cell-e.scroll+2, e.cell-4, e.cell-4)
		if r.Y+r.Height < b.Y || r.Y > b.Y+b.Height {
			continue
		}
		e.drawn++
		buf.DrawRect(render.RectCmd{Bounds: r, FillColor: e.config.PrimaryColor}, 2, 1)
	}
}

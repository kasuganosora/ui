package widget

import (
	"strings"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// RichMark covers [Start, End) in rune offsets of the whole document,
// including newline runes. Kind is bold, code, or link.
type RichMark struct {
	Start int
	End   int
	Kind  string
}

type richStyle struct {
	bold bool
	code bool
	link bool
}

// RichEditor is a plain-text document plus rune-range marks. Editing uses
// the same rune offsets as Editor. Bold is a second pass of the same face.
// It is not a ProseMirror or Tiptap document.
type RichEditor struct {
	Base
	text     string
	marks    []RichMark
	caret    TextPos
	anchor   TextPos
	scrollX  float32
	scrollY  float32
	lineH    float32
	pad      float32
	onChange func(string)
}

// NewRichEditor creates an empty rich editor.
func NewRichEditor(tree *core.Tree, cfg *Config) *RichEditor {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	e := &RichEditor{Base: NewBase(tree, core.TypeCustom, cfg), lineH: 18, pad: 4}
	tree.SetFocusable(e.id, true)
	tree.AddHandler(e.id, event.KeyDown, func(ev *event.Event) {
		if !targetPhase(ev) {
			return
		}
		shift := ev.Modifiers.Shift
		switch ev.Key {
		case event.KeyArrowLeft:
			e.moveCaret(-1, 0, shift)
		case event.KeyArrowRight:
			e.moveCaret(1, 0, shift)
		case event.KeyArrowUp:
			e.moveCaret(0, -1, shift)
		case event.KeyArrowDown:
			e.moveCaret(0, 1, shift)
		case event.KeyHome:
			e.caret.Col = 0
			if !shift {
				e.anchor = e.caret
			}
			e.reveal()
		case event.KeyEnd:
			lines := textLines(e.text)
			e.caret = clampPos(lines, e.caret)
			e.caret.Col = len([]rune(lines[e.caret.Line]))
			if !shift {
				e.anchor = e.caret
			}
			e.reveal()
		case event.KeyBackspace:
			e.backspace()
		case event.KeyDelete:
			e.deleteForward()
		case event.KeyEnter:
			e.insert("\n")
		default:
			if ev.Modifiers.Ctrl && ev.Key == event.KeyV {
				e.paste()
				break
			}
			return
		}
		ev.StopPropagation()
	})
	tree.AddHandler(e.id, event.KeyPress, func(ev *event.Event) {
		if !targetPhase(ev) || ev.Char == 0 || ev.Modifiers.Ctrl {
			return
		}
		e.insert(string(ev.Char))
		ev.StopPropagation()
	})
	tree.AddHandler(e.id, event.MouseDown, func(ev *event.Event) {
		if !targetPhase(ev) {
			return
		}
		e.caret = e.posAt(ev.GlobalX, ev.GlobalY)
		e.anchor = e.caret
		e.reveal()
	})
	return e
}

// SetText replaces the document, drops marks, and does not notify.
func (e *RichEditor) SetText(s string) {
	e.text = strings.ReplaceAll(s, "\r\n", "\n")
	e.marks = nil
	e.caret = TextPos{}
	e.anchor = e.caret
	e.scrollX, e.scrollY = 0, 0
}

func (e *RichEditor) Text() string   { return e.text }
func (e *RichEditor) Caret() TextPos { return e.caret }

// Marks returns a copy of the rune-range marks.
func (e *RichEditor) Marks() []RichMark {
	out := make([]RichMark, len(e.marks))
	copy(out, e.marks)
	return out
}

// SetMark adds a half-open rune range. It does not notify.
func (e *RichEditor) SetMark(start, end int, kind string) {
	n := len([]rune(e.text))
	if start < 0 {
		start = 0
	}
	if end > n {
		end = n
	}
	if start >= end || kind == "" {
		return
	}
	e.marks = append(e.marks, RichMark{Start: start, End: end, Kind: kind})
}

// Toggle applies or removes kind on the current selection. Text is unchanged,
// so OnChange does not run.
func (e *RichEditor) Toggle(kind string) {
	lines := textLines(e.text)
	lo, hi := orderPos(clampPos(lines, e.anchor), clampPos(lines, e.caret))
	a := posToRune(e.text, lo)
	b := posToRune(e.text, hi)
	if a >= b || kind == "" {
		return
	}
	if richCovered(e.marks, kind, a, b) {
		var keep, drop []RichMark
		for _, m := range e.marks {
			if m.Kind == kind {
				drop = append(drop, m)
			} else {
				keep = append(keep, m)
			}
		}
		e.marks = append(keep, cutRich(drop, a, b)...)
		return
	}
	e.marks = append(e.marks, RichMark{Start: a, End: b, Kind: kind})
}

// OnChange reports the full text after one user edit that changed it.
func (e *RichEditor) OnChange(fn func(string)) { e.onChange = fn }

func (e *RichEditor) changed() {
	if e.onChange != nil {
		e.onChange(e.text)
	}
}

// SelectedText is the rune range between the anchor and the caret.
func (e *RichEditor) SelectedText() string { return selectedText(e.text, e.anchor, e.caret) }

func (e *RichEditor) measure(s string) float32 {
	size := e.config.FontSizeSm
	if e.config.TextRenderer != nil {
		return e.config.TextRenderer.MeasureText(s, size)
	}
	return float32(len([]rune(s))) * size * 0.55
}

func (e *RichEditor) measureCol(line string, col int) float32 {
	return prefixWidth(line, col, e.measure)
}

func (e *RichEditor) posAt(x, y float32) TextPos {
	b := e.Bounds()
	lines := textLines(e.text)
	localY := y - b.Y + e.scrollY
	line := int(localY / e.lineH)
	if line < 0 {
		line = 0
	}
	if line >= len(lines) {
		line = len(lines) - 1
	}
	localX := x - (b.X + e.pad) + e.scrollX
	return clampPos(lines, TextPos{Line: line, Col: colAtX(lines[line], localX, e.measure)})
}

func (e *RichEditor) moveCaret(dCol, dLine int, keepAnchor bool) {
	lines := textLines(e.text)
	p := clampPos(lines, e.caret)
	if dLine != 0 {
		p.Line += dLine
		p = clampPos(lines, p)
	}
	if dCol != 0 {
		r := []rune(lines[p.Line])
		p.Col += dCol
		if p.Col < 0 {
			if p.Line > 0 {
				p.Line--
				p.Col = len([]rune(lines[p.Line]))
			} else {
				p.Col = 0
			}
		} else if p.Col > len(r) {
			if p.Line+1 < len(lines) {
				p.Line++
				p.Col = 0
			} else {
				p.Col = len(r)
			}
		}
	}
	e.caret = clampPos(lines, p)
	if !keepAnchor {
		e.anchor = e.caret
	}
	e.reveal()
}

func (e *RichEditor) backspace() {
	if e.SelectedText() != "" {
		if e.deleteSelection() {
			e.changed()
		}
		return
	}
	r := []rune(e.text)
	off := posToRune(e.text, e.caret)
	if off <= 0 {
		return
	}
	e.marks = cutRich(e.marks, off-1, off)
	out := make([]rune, 0, len(r)-1)
	out = append(out, r[:off-1]...)
	out = append(out, r[off:]...)
	e.text = string(out)
	e.caret = runeToPos(e.text, off-1)
	e.anchor = e.caret
	e.reveal()
	e.changed()
}

func (e *RichEditor) deleteForward() {
	if e.SelectedText() != "" {
		if e.deleteSelection() {
			e.changed()
		}
		return
	}
	r := []rune(e.text)
	off := posToRune(e.text, e.caret)
	if off >= len(r) {
		return
	}
	e.marks = cutRich(e.marks, off, off+1)
	out := make([]rune, 0, len(r)-1)
	out = append(out, r[:off]...)
	out = append(out, r[off+1:]...)
	e.text = string(out)
	e.caret = runeToPos(e.text, off)
	e.anchor = e.caret
	e.reveal()
	e.changed()
}

func (e *RichEditor) deleteSelection() bool {
	lines := textLines(e.text)
	lo, hi := orderPos(clampPos(lines, e.anchor), clampPos(lines, e.caret))
	if lo == hi {
		return false
	}
	r := []rune(e.text)
	a := posToRune(e.text, lo)
	b := posToRune(e.text, hi)
	if a < 0 {
		a = 0
	}
	if b > len(r) {
		b = len(r)
	}
	if a >= b {
		return false
	}
	e.marks = cutRich(e.marks, a, b)
	out := make([]rune, 0, len(r)-(b-a))
	out = append(out, r[:a]...)
	out = append(out, r[b:]...)
	e.text = string(out)
	e.caret = runeToPos(e.text, a)
	e.anchor = e.caret
	e.reveal()
	return true
}

func (e *RichEditor) insert(s string) {
	if s == "" {
		return
	}
	if e.SelectedText() != "" {
		e.deleteSelection()
	}
	r := []rune(e.text)
	off := posToRune(e.text, e.caret)
	add := []rune(s)
	e.marks = shiftRich(e.marks, off, len(add))
	out := make([]rune, 0, len(r)+len(add))
	out = append(out, r[:off]...)
	out = append(out, add...)
	out = append(out, r[off:]...)
	e.text = string(out)
	e.caret = runeToPos(e.text, off+len(add))
	e.anchor = e.caret
	e.reveal()
	e.changed()
}

func (e *RichEditor) paste() {
	if e.config == nil || e.config.Platform == nil {
		return
	}
	text := e.config.Platform.GetClipboardText()
	if text == "" {
		return
	}
	e.insert(text)
}

func (e *RichEditor) reveal() {
	b := e.Bounds()
	lines := textLines(e.text)
	p := clampPos(lines, e.caret)
	y := float32(p.Line) * e.lineH
	if y < e.scrollY {
		e.scrollY = y
	}
	if b.Height > 0 && y+e.lineH > e.scrollY+b.Height {
		e.scrollY = y + e.lineH - b.Height
	}
	if e.scrollY < 0 {
		e.scrollY = 0
	}
	x := e.measureCol(lines[p.Line], p.Col)
	if x < e.scrollX {
		e.scrollX = x
	}
	if b.Width > e.pad*2 && x > e.scrollX+b.Width-e.pad*2 {
		e.scrollX = x - b.Width + e.pad*2
	}
	if e.scrollX < 0 {
		e.scrollX = 0
	}
}

// Draw paints real glyph runs. Bold repeats the same face one pixel to the
// right. Code may also fill a background. Lines below the viewport are skipped.
// Lines do not soft-wrap.
func (e *RichEditor) Draw(buf *render.CommandBuffer) {
	b := e.Bounds()
	if b.IsEmpty() {
		return
	}
	cfg := e.config
	lines := textLines(e.text)
	buf.PushClip(b)
	buf.DrawRect(render.RectCmd{Bounds: b, FillColor: cfg.BgColor}, 0, 1)
	lo, hi := orderPos(e.anchor, e.caret)
	for i, line := range lines {
		y := b.Y + float32(i)*e.lineH - e.scrollY
		if y+e.lineH <= b.Y {
			continue
		}
		if y >= b.Y+b.Height {
			break
		}
		if i >= lo.Line && i <= hi.Line && (lo.Line != hi.Line || lo.Col != hi.Col) {
			r := []rune(line)
			start, end := 0, len(r)
			if i == lo.Line {
				start = lo.Col
			}
			if i == hi.Line {
				end = hi.Col
			}
			x0 := b.X + e.pad + e.measureCol(line, start) - e.scrollX
			x1 := b.X + e.pad + e.measureCol(line, end) - e.scrollX
			if x1 < x0 {
				x0, x1 = x1, x0
			}
			buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(x0, y, x1-x0, e.lineH), FillColor: cfg.HoverColor}, 1, 1)
		}
		e.drawLine(buf, line, i, b.X+e.pad-e.scrollX, y+1)
	}
	pt := e.caretPoint()
	buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(pt.X, pt.Y, 1, e.lineH-2), FillColor: cfg.TextColor}, 6, 1)
	buf.PopClip()
}

func (e *RichEditor) caretPoint() uimath.Vec2 {
	b := e.Bounds()
	lines := textLines(e.text)
	p := clampPos(lines, e.caret)
	return uimath.Vec2{
		X: b.X + e.pad + e.measureCol(lines[p.Line], p.Col) - e.scrollX,
		Y: b.Y + float32(p.Line)*e.lineH - e.scrollY,
	}
}

func (e *RichEditor) drawLine(buf *render.CommandBuffer, line string, lineNo int, x, y float32) {
	cfg := e.config
	size := cfg.FontSizeSm
	origin := posToRune(e.text, TextPos{Line: lineNo, Col: 0})
	r := []rune(line)
	i := 0
	for i < len(r) {
		st := e.styleAt(origin + i)
		j := i + 1
		for j < len(r) && e.styleAt(origin+j) == st {
			j++
		}
		seg := string(r[i:j])
		w := e.measure(seg)
		color := cfg.TextColor
		if st.link {
			color = cfg.PrimaryColor
		}
		if st.code {
			buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(x, y-1, w, e.lineH), FillColor: cfg.BorderColor}, 1, 1)
		}
		emitText(cfg, buf, seg, x, y, size, w+1, color, 1, 4)
		if st.bold {
			emitText(cfg, buf, seg, x+1, y, size, w+1, color, 1, 4)
		}
		x += w
		i = j
	}
}

func (e *RichEditor) styleAt(off int) richStyle {
	var s richStyle
	for _, m := range e.marks {
		if off < m.Start || off >= m.End {
			continue
		}
		switch m.Kind {
		case "bold":
			s.bold = true
		case "code":
			s.code = true
		case "link":
			s.link = true
		}
	}
	return s
}

func richCovered(marks []RichMark, kind string, a, b int) bool {
	for _, m := range marks {
		if m.Kind == kind && m.Start <= a && m.End >= b {
			return true
		}
	}
	return false
}

func shiftRich(marks []RichMark, at, n int) []RichMark {
	if n == 0 {
		return marks
	}
	out := make([]RichMark, 0, len(marks))
	for _, m := range marks {
		switch {
		case m.End <= at:
			out = append(out, m)
		case m.Start >= at:
			m.Start += n
			m.End += n
			out = append(out, m)
		default:
			m.End += n
			out = append(out, m)
		}
	}
	return out
}

func cutRich(marks []RichMark, a, b int) []RichMark {
	if a >= b {
		return marks
	}
	n := b - a
	out := make([]RichMark, 0, len(marks))
	for _, m := range marks {
		ns, ne := m.Start, m.End
		if ne <= a {
			out = append(out, m)
			continue
		}
		if ns >= b {
			m.Start = ns - n
			m.End = ne - n
			if m.Start < m.End {
				out = append(out, m)
			}
			continue
		}
		if ns < a {
			ns = m.Start
		} else {
			ns = a
		}
		if ne >= b {
			ne -= n
		} else {
			ne = a
		}
		if ns < ne {
			m.Start, m.End = ns, ne
			out = append(out, m)
		}
	}
	return out
}

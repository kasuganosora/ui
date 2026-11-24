package widget

import (
	"strings"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Editor is a plain-text code editor. Offsets are rune indexes. Keywords are
// a fixed set colored with the text pipeline. It does not use chroma.
type Editor struct {
	Base
	text    string
	caret   TextPos
	anchor  TextPos
	scrollX float32
	scrollY float32
	lineH   float32
	pad     float32
}

// NewEditor creates an empty editor.
func NewEditor(tree *core.Tree, cfg *Config) *Editor {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	e := &Editor{Base: NewBase(tree, core.TypeCustom, cfg), lineH: 18, pad: 4}
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
		default:
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

// SetText replaces the buffer and puts the caret at the start.
func (e *Editor) SetText(s string) {
	e.text = strings.ReplaceAll(s, "\r\n", "\n")
	e.caret = TextPos{}
	e.anchor = e.caret
	e.scrollX, e.scrollY = 0, 0
}

func (e *Editor) Text() string   { return e.text }
func (e *Editor) Caret() TextPos { return e.caret }

// SelectedText is the rune range between the anchor and the caret.
func (e *Editor) SelectedText() string { return selectedText(e.text, e.anchor, e.caret) }

// CaretPoint is the caret's top-left in screen space.
func (e *Editor) CaretPoint() uimath.Vec2 {
	b := e.Bounds()
	lines := textLines(e.text)
	p := clampPos(lines, e.caret)
	return uimath.Vec2{
		X: b.X + e.pad + e.measureCol(lines[p.Line], p.Col) - e.scrollX,
		Y: b.Y + float32(p.Line)*e.lineH - e.scrollY,
	}
}

func (e *Editor) measureCol(line string, col int) float32 {
	return prefixWidth(line, col, e.measure)
}

func (e *Editor) measure(s string) float32 {
	size := e.config.FontSizeSm
	if e.config.TextRenderer != nil {
		return e.config.TextRenderer.MeasureText(s, size)
	}
	return float32(len([]rune(s))) * size * 0.55
}

func (e *Editor) posAt(x, y float32) TextPos {
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

func (e *Editor) moveCaret(dCol, dLine int, keepAnchor bool) {
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

func (e *Editor) backspace() {
	if e.SelectedText() != "" {
		e.deleteSelection()
		return
	}
	e.moveCaret(-1, 0, false)
	e.deleteForward()
}

func (e *Editor) deleteForward() {
	if e.SelectedText() != "" {
		e.deleteSelection()
		return
	}
	lines := textLines(e.text)
	p := clampPos(lines, e.caret)
	r := []rune(lines[p.Line])
	if p.Col < len(r) {
		r = append(r[:p.Col], r[p.Col+1:]...)
		lines[p.Line] = string(r)
	} else if p.Line+1 < len(lines) {
		lines[p.Line] = lines[p.Line] + lines[p.Line+1]
		lines = append(lines[:p.Line+1], lines[p.Line+2:]...)
	}
	e.text = strings.Join(lines, "\n")
	e.caret = clampPos(textLines(e.text), p)
	e.anchor = e.caret
}

func (e *Editor) deleteSelection() {
	lines := textLines(e.text)
	lo, hi := orderPos(clampPos(lines, e.anchor), clampPos(lines, e.caret))
	if lo == hi {
		return
	}
	var b strings.Builder
	for i, line := range lines {
		r := []rune(line)
		start, end := 0, len(r)
		if i == lo.Line {
			start = lo.Col
		}
		if i == hi.Line {
			end = hi.Col
		}
		if i < lo.Line || i > hi.Line {
			if b.Len() > 0 {
				b.WriteByte('\n')
			}
			b.WriteString(line)
			continue
		}
		if i == lo.Line {
			if b.Len() > 0 && lo.Line != 0 {
				// joined below
			}
			b.WriteString(string(r[:start]))
			if i == hi.Line {
				b.WriteString(string(r[end:]))
			}
			continue
		}
		if i == hi.Line {
			b.WriteString(string(r[end:]))
		}
	}
	e.text = b.String()
	e.caret = lo
	e.anchor = lo
	e.reveal()
}

func (e *Editor) insert(s string) {
	if e.SelectedText() != "" {
		e.deleteSelection()
	}
	lines := textLines(e.text)
	p := clampPos(lines, e.caret)
	r := []rune(lines[p.Line])
	add := []rune(s)
	r = append(r[:p.Col], append(add, r[p.Col:]...)...)
	lines[p.Line] = string(r)
	p.Col += len(add)
	e.text = strings.Join(lines, "\n")
	e.caret = p
	e.anchor = p
	e.reveal()
}

func (e *Editor) reveal() {
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

func (e *Editor) Draw(buf *render.CommandBuffer) {
	b := e.Bounds()
	if b.IsEmpty() {
		return
	}
	cfg := e.config
	lines := textLines(e.text)
	buf.PushClip(b)
	buf.DrawRect(render.RectCmd{Bounds: b, FillColor: uimath.ColorWhite}, 0, 1)
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
		e.drawLine(buf, line, b.X+e.pad-e.scrollX, y+1)
	}
	pt := e.CaretPoint()
	buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(pt.X, pt.Y, 1, e.lineH-2), FillColor: cfg.TextColor}, 6, 1)
	buf.PopClip()
}

func (e *Editor) drawLine(buf *render.CommandBuffer, line string, x, y float32) {
	cfg := e.config
	size := cfg.FontSizeSm
	for _, tok := range codeTokens(line) {
		color := cfg.TextColor
		if tok.kw {
			color = cfg.PrimaryColor
		}
		w := e.measure(tok.text)
		emitText(cfg, buf, tok.text, x, y, size, w+1, color, 1, 4)
		x += w
	}
}

type codeTok struct {
	text string
	kw   bool
}

func codeTokens(line string) []codeTok {
	var out []codeTok
	r := []rune(line)
	i := 0
	for i < len(r) {
		if isWord(r[i]) {
			j := i + 1
			for j < len(r) && isWord(r[j]) {
				j++
			}
			word := string(r[i:j])
			out = append(out, codeTok{text: word, kw: isKeyword(word)})
			i = j
			continue
		}
		j := i + 1
		for j < len(r) && !isWord(r[j]) {
			j++
		}
		out = append(out, codeTok{text: string(r[i:j])})
		i = j
	}
	return out
}

func isWord(r rune) bool {
	return r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

func isKeyword(w string) bool {
	switch w {
	case "func", "return", "if", "var", "for", "type", "package":
		return true
	default:
		return false
	}
}

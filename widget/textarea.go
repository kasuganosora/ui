package widget

import (
	"strings"
	"time"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/render"
)

// TextArea is a multi-line text input widget.
type TextArea struct {
	Base
	value       string
	placeholder string
	disabled    bool
	rows        int // visible rows (default 4)
	cursorPos   int // cursor position in runes
	selAnchor   int // selection anchor (-1 = no selection)

	dragging     bool
	dragSelected bool

	onChange func(value string)
}

// NewTextArea creates a multi-line text area.
func NewTextArea(tree *core.Tree, cfg *Config) *TextArea {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	ta := &TextArea{
		Base:      NewBase(tree, core.TypeInput, cfg),
		rows:      4,
		selAnchor: -1,
	}

	lineH := cfg.FontSize * cfg.LineHeight
	height := lineH*float32(ta.rows) + cfg.SpaceSM*2
	ta.style.Display = layout.DisplayBlock
	ta.style.Height = layout.Px(height)
	ta.style.Padding = layout.EdgeValues{
		Top: layout.Px(cfg.SpaceSM), Bottom: layout.Px(cfg.SpaceSM),
		Left: layout.Px(cfg.SpaceSM), Right: layout.Px(cfg.SpaceSM),
	}

	ta.tree.AddHandler(ta.id, event.MouseClick, func(e *event.Event) {
		if ta.disabled {
			return
		}
		ta.tree.SetFocused(ta.id, true)
		if ta.dragSelected {
			ta.dragSelected = false
			return
		}
		pos := ta.hitTestChar(e.X, e.Y)
		if e.Modifiers.Shift && ta.selAnchor >= 0 {
			ta.cursorPos = pos
		} else {
			ta.cursorPos = pos
			ta.selAnchor = -1
		}
	})

	ta.tree.AddHandler(ta.id, event.MouseDown, func(e *event.Event) {
		if ta.disabled {
			return
		}
		if e.Button == event.MouseButtonRight {
			ta.showContextMenu(int(e.GlobalX), int(e.GlobalY))
			return
		}
		ta.dragging = true
		ta.dragSelected = false
		pos := ta.hitTestChar(e.X, e.Y)
		ta.selAnchor = pos
		ta.cursorPos = pos
	})

	ta.tree.AddHandler(ta.id, event.MouseMove, func(e *event.Event) {
		if !ta.dragging || ta.disabled {
			return
		}
		pos := ta.hitTestChar(e.X, e.Y)
		ta.cursorPos = pos
	})

	ta.tree.AddHandler(ta.id, event.MouseUp, func(e *event.Event) {
		if ta.dragging {
			ta.dragging = false
			if ta.selAnchor == ta.cursorPos {
				ta.selAnchor = -1
				ta.dragSelected = false
			} else {
				ta.dragSelected = true
			}
		}
	})

	ta.tree.AddHandler(ta.id, event.KeyPress, func(e *event.Event) {
		if ta.disabled {
			return
		}
		if e.Char != 0 {
			ta.deleteSelection()
			ta.insertChar(e.Char)
		}
	})

	ta.tree.AddHandler(ta.id, event.KeyDown, func(e *event.Event) {
		if ta.disabled {
			return
		}

		if e.Modifiers.Ctrl {
			switch e.Key {
			case event.KeyA:
				ta.selectAll()
			case event.KeyC:
				ta.copySelection()
			case event.KeyX:
				ta.cutSelection()
			case event.KeyV:
				ta.paste()
			}
			return
		}

		shift := e.Modifiers.Shift
		switch e.Key {
		case event.KeyEnter:
			ta.deleteSelection()
			ta.insertChar('\n')
		case event.KeyBackspace:
			if ta.hasSelection() {
				ta.deleteSelection()
			} else {
				ta.deleteBack()
			}
		case event.KeyDelete:
			if ta.hasSelection() {
				ta.deleteSelection()
			} else {
				ta.deleteForward()
			}
		case event.KeyArrowLeft:
			if shift {
				ta.startSelection()
				if ta.cursorPos > 0 {
					ta.cursorPos--
				}
			} else {
				if ta.hasSelection() {
					ta.cursorPos = ta.selMin()
					ta.selAnchor = -1
				} else if ta.cursorPos > 0 {
					ta.cursorPos--
				}
			}
		case event.KeyArrowRight:
			if shift {
				ta.startSelection()
				if ta.cursorPos < ta.runeLen() {
					ta.cursorPos++
				}
			} else {
				if ta.hasSelection() {
					ta.cursorPos = ta.selMax()
					ta.selAnchor = -1
				} else if ta.cursorPos < ta.runeLen() {
					ta.cursorPos++
				}
			}
		case event.KeyArrowUp:
			if shift {
				ta.startSelection()
			} else {
				ta.selAnchor = -1
			}
			ta.moveCursorVertical(-1)
		case event.KeyArrowDown:
			if shift {
				ta.startSelection()
			} else {
				ta.selAnchor = -1
			}
			ta.moveCursorVertical(1)
		case event.KeyHome:
			if shift {
				ta.startSelection()
			} else {
				ta.selAnchor = -1
			}
			ta.cursorPos = ta.lineStart(ta.cursorPos)
		case event.KeyEnd:
			if shift {
				ta.startSelection()
			} else {
				ta.selAnchor = -1
			}
			ta.cursorPos = ta.lineEnd(ta.cursorPos)
		}
	})

	// IME support
	ta.tree.AddHandler(ta.id, event.IMECompositionEnd, func(e *event.Event) {
		if ta.disabled || e.Text == "" {
			return
		}
		ta.deleteSelection()
		for _, ch := range e.Text {
			ta.insertChar(ch)
		}
	})

	return ta
}

func (ta *TextArea) Value() string       { return ta.value }
func (ta *TextArea) Placeholder() string { return ta.placeholder }
func (ta *TextArea) IsDisabled() bool    { return ta.disabled }
func (ta *TextArea) Rows() int           { return ta.rows }

func (ta *TextArea) SetValue(v string) {
	ta.value = v
	if ta.cursorPos > len([]rune(v)) {
		ta.cursorPos = len([]rune(v))
	}
	ta.selAnchor = -1
	ta.tree.SetProperty(ta.id, "text", v)
}

func (ta *TextArea) SetPlaceholder(p string) { ta.placeholder = p }

func (ta *TextArea) SetDisabled(d bool) {
	ta.disabled = d
	ta.tree.SetEnabled(ta.id, !d)
}

func (ta *TextArea) SetRows(rows int) {
	if rows < 1 {
		rows = 1
	}
	ta.rows = rows
	lineH := ta.config.FontSize * ta.config.LineHeight
	ta.style.Height = layout.Px(lineH*float32(rows) + ta.config.SpaceSM*2)
}

func (ta *TextArea) OnChange(fn func(string)) { ta.onChange = fn }

// --- Internal helpers ---

func (ta *TextArea) runeLen() int { return len([]rune(ta.value)) }

func (ta *TextArea) hasSelection() bool {
	return ta.selAnchor >= 0 && ta.selAnchor != ta.cursorPos
}

func (ta *TextArea) selMin() int {
	if ta.selAnchor < ta.cursorPos {
		return ta.selAnchor
	}
	return ta.cursorPos
}

func (ta *TextArea) selMax() int {
	if ta.selAnchor > ta.cursorPos {
		return ta.selAnchor
	}
	return ta.cursorPos
}

func (ta *TextArea) startSelection() {
	if ta.selAnchor < 0 {
		ta.selAnchor = ta.cursorPos
	}
}

func (ta *TextArea) selectAll() {
	ta.selAnchor = 0
	ta.cursorPos = ta.runeLen()
}

func (ta *TextArea) selectedText() string {
	if !ta.hasSelection() {
		return ""
	}
	runes := []rune(ta.value)
	lo, hi := ta.selMin(), ta.selMax()
	if hi > len(runes) {
		hi = len(runes)
	}
	return string(runes[lo:hi])
}

func (ta *TextArea) deleteSelection() {
	if !ta.hasSelection() {
		return
	}
	runes := []rune(ta.value)
	lo, hi := ta.selMin(), ta.selMax()
	if hi > len(runes) {
		hi = len(runes)
	}
	runes = append(runes[:lo], runes[hi:]...)
	ta.value = string(runes)
	ta.cursorPos = lo
	ta.selAnchor = -1
	ta.tree.SetProperty(ta.id, "text", ta.value)
	if ta.onChange != nil {
		ta.onChange(ta.value)
	}
}

func (ta *TextArea) insertChar(ch rune) {
	runes := []rune(ta.value)
	pos := ta.cursorPos
	if pos > len(runes) {
		pos = len(runes)
	}
	runes = append(runes[:pos], append([]rune{ch}, runes[pos:]...)...)
	ta.value = string(runes)
	ta.cursorPos = pos + 1
	ta.selAnchor = -1
	ta.tree.SetProperty(ta.id, "text", ta.value)
	if ta.onChange != nil {
		ta.onChange(ta.value)
	}
}

func (ta *TextArea) deleteBack() {
	if ta.cursorPos <= 0 {
		return
	}
	runes := []rune(ta.value)
	pos := ta.cursorPos
	if pos > len(runes) {
		pos = len(runes)
	}
	runes = append(runes[:pos-1], runes[pos:]...)
	ta.value = string(runes)
	ta.cursorPos = pos - 1
	ta.tree.SetProperty(ta.id, "text", ta.value)
	if ta.onChange != nil {
		ta.onChange(ta.value)
	}
}

func (ta *TextArea) deleteForward() {
	runes := []rune(ta.value)
	pos := ta.cursorPos
	if pos >= len(runes) {
		return
	}
	runes = append(runes[:pos], runes[pos+1:]...)
	ta.value = string(runes)
	ta.tree.SetProperty(ta.id, "text", ta.value)
	if ta.onChange != nil {
		ta.onChange(ta.value)
	}
}

// --- Clipboard ---

func (ta *TextArea) copySelection() {
	text := ta.selectedText()
	if text == "" {
		return
	}
	if ta.config.Platform != nil {
		ta.config.Platform.SetClipboardText(text)
	}
}

func (ta *TextArea) cutSelection() {
	ta.copySelection()
	ta.deleteSelection()
}

func (ta *TextArea) paste() {
	if ta.config.Platform == nil {
		return
	}
	text := ta.config.Platform.GetClipboardText()
	if text == "" {
		return
	}
	ta.deleteSelection()
	runes := []rune(ta.value)
	pos := ta.cursorPos
	if pos > len(runes) {
		pos = len(runes)
	}
	inserted := []rune(text)
	runes = append(runes[:pos], append(inserted, runes[pos:]...)...)
	ta.value = string(runes)
	ta.cursorPos = pos + len(inserted)
	ta.selAnchor = -1
	ta.tree.SetProperty(ta.id, "text", ta.value)
	if ta.onChange != nil {
		ta.onChange(ta.value)
	}
}

func (ta *TextArea) showContextMenu(clientX, clientY int) {
	win := ta.config.Window
	if win == nil {
		return
	}
	hasSel := ta.hasSelection()
	hasClip := false
	if ta.config.Platform != nil {
		hasClip = ta.config.Platform.GetClipboardText() != ""
	}
	items := []platform.ContextMenuItem{
		{Label: "剪切(X)", Enabled: hasSel},
		{Label: "复制(C)", Enabled: hasSel},
		{Label: "粘贴(V)", Enabled: hasClip},
		{Label: "全选(A)", Enabled: ta.runeLen() > 0},
	}
	idx := win.ShowContextMenu(clientX, clientY, items)
	switch idx {
	case 0:
		ta.cutSelection()
	case 1:
		ta.copySelection()
	case 2:
		ta.paste()
	case 3:
		ta.selectAll()
	}
}

// --- Line helpers ---

// lines splits the value into visual lines.
func (ta *TextArea) lines() []string {
	if ta.value == "" {
		return []string{""}
	}
	return strings.Split(ta.value, "\n")
}

// runeOffsetToLineCol converts a rune offset to (line, col).
func (ta *TextArea) runeOffsetToLineCol(pos int) (line, col int) {
	runes := []rune(ta.value)
	if pos > len(runes) {
		pos = len(runes)
	}
	for i := 0; i < pos; i++ {
		if runes[i] == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}
	return
}

// lineColToRuneOffset converts (line, col) to a rune offset.
func (ta *TextArea) lineColToRuneOffset(line, col int) int {
	ll := ta.lines()
	if line >= len(ll) {
		line = len(ll) - 1
	}
	if line < 0 {
		line = 0
	}
	offset := 0
	for i := 0; i < line; i++ {
		offset += len([]rune(ll[i])) + 1 // +1 for \n
	}
	lineRunes := len([]rune(ll[line]))
	if col > lineRunes {
		col = lineRunes
	}
	return offset + col
}

// lineStart returns the rune offset of the start of the line containing pos.
func (ta *TextArea) lineStart(pos int) int {
	line, _ := ta.runeOffsetToLineCol(pos)
	return ta.lineColToRuneOffset(line, 0)
}

// lineEnd returns the rune offset of the end of the line containing pos.
func (ta *TextArea) lineEnd(pos int) int {
	line, _ := ta.runeOffsetToLineCol(pos)
	ll := ta.lines()
	if line < len(ll) {
		return ta.lineColToRuneOffset(line, len([]rune(ll[line])))
	}
	return ta.runeLen()
}

// moveCursorVertical moves the cursor up (dir=-1) or down (dir=1).
func (ta *TextArea) moveCursorVertical(dir int) {
	line, col := ta.runeOffsetToLineCol(ta.cursorPos)
	newLine := line + dir
	ta.cursorPos = ta.lineColToRuneOffset(newLine, col)
}

// --- Hit testing ---

func (ta *TextArea) measureRunes(runes []rune) float32 {
	cfg := ta.config
	if cfg.TextRenderer != nil {
		return cfg.TextRenderer.MeasureText(string(runes), cfg.FontSize)
	}
	return float32(len(runes)) * cfg.FontSize * 0.55
}

func (ta *TextArea) hitTestChar(localX, localY float32) int {
	bounds := ta.Bounds()
	cfg := ta.config
	padLeft := cfg.SpaceSM
	padTop := cfg.SpaceSM
	lineH := cfg.FontSize * cfg.LineHeight

	relX := localX - bounds.X - padLeft
	relY := localY - bounds.Y - padTop
	if relY < 0 {
		relY = 0
	}

	lineIdx := int(relY / lineH)
	ll := ta.lines()
	if lineIdx >= len(ll) {
		lineIdx = len(ll) - 1
	}
	if lineIdx < 0 {
		lineIdx = 0
	}

	// Find character within line
	lineRunes := []rune(ll[lineIdx])
	col := 0
	if relX > 0 {
		for i := 1; i <= len(lineRunes); i++ {
			w := ta.measureRunes(lineRunes[:i])
			if cfg.TextRenderer != nil {
				prevW := float32(0)
				if i > 1 {
					prevW = ta.measureRunes(lineRunes[:i-1])
				}
				mid := prevW + (w-prevW)/2
				if relX < mid {
					col = i - 1
					goto done
				}
			} else {
				if relX < w {
					col = i - 1
					goto done
				}
			}
		}
		col = len(lineRunes)
	}

done:
	return ta.lineColToRuneOffset(lineIdx, col)
}

// --- Drawing ---

func (ta *TextArea) Draw(buf *render.CommandBuffer) {
	bounds := ta.Bounds()
	if bounds.IsEmpty() {
		return
	}

	cfg := ta.config
	elem := ta.Element()
	focused := elem != nil && elem.IsFocused()

	// Border
	borderClr := cfg.BorderColor
	if focused {
		borderClr = cfg.FocusBorderColor
	}
	if ta.disabled {
		borderClr = cfg.DisabledColor
	}

	bgClr := cfg.BgColor
	if ta.disabled {
		bgClr = uimath.ColorHex("#f5f5f5")
	}

	buf.DrawRect(render.RectCmd{
		Bounds:      bounds,
		FillColor:   bgClr,
		BorderColor: borderClr,
		BorderWidth: cfg.BorderWidth,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 0, 1)

	padLeft := cfg.SpaceSM
	padTop := cfg.SpaceSM
	lineH := cfg.FontSize * cfg.LineHeight
	contentW := bounds.Width - padLeft*2

	// Draw selection highlight
	if focused && ta.hasSelection() {
		ta.drawSelection(buf, bounds, padLeft, padTop, lineH, contentW)
	}

	// Draw text or placeholder
	displayText := ta.value
	textColor := cfg.TextColor
	if displayText == "" && ta.placeholder != "" {
		displayText = ta.placeholder
		textColor = cfg.DisabledColor
	}

	if displayText != "" {
		ll := strings.Split(displayText, "\n")
		for i, line := range ll {
			y := bounds.Y + padTop + float32(i)*lineH
			if y > bounds.Y+bounds.Height {
				break
			}
			if line == "" {
				continue
			}
			if cfg.TextRenderer != nil {
				lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
				ty := y + (lineH-lh)/2
				cfg.TextRenderer.DrawText(buf, line, bounds.X+padLeft, ty, cfg.FontSize, contentW, textColor, 1)
			} else {
				textW := float32(len(line)) * cfg.FontSize * 0.55
				if textW > contentW {
					textW = contentW
				}
				textH := cfg.FontSize * 1.2
				ty := y + (lineH-textH)/2
				buf.DrawRect(render.RectCmd{
					Bounds:    uimath.NewRect(bounds.X+padLeft, ty, textW, textH),
					FillColor: textColor,
					Corners:   uimath.CornersAll(2),
				}, 1, 1)
			}
		}
	}

	// Cursor
	if focused && !ta.disabled {
		ms := time.Now().UnixMilli()
		if (ms/500)%2 == 0 {
			line, col := ta.runeOffsetToLineCol(ta.cursorPos)
			ll := ta.lines()
			cx := float32(0)
			if line < len(ll) {
				lineRunes := []rune(ll[line])
				if col > len(lineRunes) {
					col = len(lineRunes)
				}
				cx = ta.measureRunes(lineRunes[:col])
			}
			cursorX := bounds.X + padLeft + cx
			cursorY := bounds.Y + padTop + float32(line)*lineH
			var cursorH float32
			if cfg.TextRenderer != nil {
				cursorH = cfg.TextRenderer.LineHeight(cfg.FontSize)
			} else {
				cursorH = cfg.FontSize * 1.2
			}
			cursorY += (lineH - cursorH) / 2

			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(cursorX, cursorY, 1, cursorH),
				FillColor: cfg.TextColor,
			}, 2, 1)
		}
	}
}

func (ta *TextArea) drawSelection(buf *render.CommandBuffer, bounds uimath.Rect, padLeft, padTop, lineH, contentW float32) {
	cfg := ta.config
	lo, hi := ta.selMin(), ta.selMax()
	loLine, loCol := ta.runeOffsetToLineCol(lo)
	hiLine, hiCol := ta.runeOffsetToLineCol(hi)

	selColor := uimath.ColorHex("#1677ff")
	selColor.A = 0.3

	ll := ta.lines()

	var lh float32
	if cfg.TextRenderer != nil {
		lh = cfg.TextRenderer.LineHeight(cfg.FontSize)
	} else {
		lh = cfg.FontSize * 1.2
	}

	for lineIdx := loLine; lineIdx <= hiLine && lineIdx < len(ll); lineIdx++ {
		lineRunes := []rune(ll[lineIdx])
		startCol := 0
		endCol := len(lineRunes)
		if lineIdx == loLine {
			startCol = loCol
		}
		if lineIdx == hiLine {
			endCol = hiCol
		}
		if startCol > len(lineRunes) {
			startCol = len(lineRunes)
		}
		if endCol > len(lineRunes) {
			endCol = len(lineRunes)
		}

		sx := ta.measureRunes(lineRunes[:startCol])
		ex := ta.measureRunes(lineRunes[:endCol])

		y := bounds.Y + padTop + float32(lineIdx)*lineH + (lineH-lh)/2

		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(bounds.X+padLeft+sx, y, ex-sx, lh),
			FillColor: selColor,
		}, 1, 1)
	}
}

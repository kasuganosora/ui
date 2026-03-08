package widget

import (
	"fmt"
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

	status            Status
	maxLength         int // 0 = unlimited
	tips              string
	showLimitNumber   bool
	readonly          bool
	allowInputOverMax bool
	autosize          bool
	autosizeMinRows   int
	autosizeMaxRows   int

	onChange  func(value string)
	onBlur   func(value string)
	onFocus  func(value string)
	onKeydown func(value string, key event.Key)
	onKeyup   func(value string, key event.Key)
	onValidate func(error string)
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
		ta.updateIMEPosition()
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
		if ta.disabled || ta.readonly {
			return
		}
		if e.Char != 0 {
			ta.deleteSelection()
			ta.insertChar(e.Char)
			ta.updateIMEPosition()
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
				ta.updateIMEPosition()
			}
			return
		}

		shift := e.Modifiers.Shift
		switch e.Key {
		case event.KeyEnter:
			ta.deleteSelection()
			ta.insertChar('\n')
			ta.updateIMEPosition()
		case event.KeyBackspace:
			if ta.hasSelection() {
				ta.deleteSelection()
			} else {
				ta.deleteBack()
			}
			ta.updateIMEPosition()
		case event.KeyDelete:
			if ta.hasSelection() {
				ta.deleteSelection()
			} else {
				ta.deleteForward()
			}
			ta.updateIMEPosition()
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
			ta.updateIMEPosition()
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
			ta.updateIMEPosition()
		case event.KeyArrowUp:
			if shift {
				ta.startSelection()
			} else {
				ta.selAnchor = -1
			}
			ta.moveCursorVertical(-1)
			ta.updateIMEPosition()
		case event.KeyArrowDown:
			if shift {
				ta.startSelection()
			} else {
				ta.selAnchor = -1
			}
			ta.moveCursorVertical(1)
			ta.updateIMEPosition()
		case event.KeyHome:
			if shift {
				ta.startSelection()
			} else {
				ta.selAnchor = -1
			}
			ta.cursorPos = ta.lineStart(ta.cursorPos)
			ta.updateIMEPosition()
		case event.KeyEnd:
			if shift {
				ta.startSelection()
			} else {
				ta.selAnchor = -1
			}
			ta.cursorPos = ta.lineEnd(ta.cursorPos)
			ta.updateIMEPosition()
		}

		// Fire onKeydown callback
		if ta.onKeydown != nil {
			ta.onKeydown(ta.value, e.Key)
		}
	})

	ta.tree.AddHandler(ta.id, event.KeyUp, func(e *event.Event) {
		if ta.disabled {
			return
		}
		if ta.onKeyup != nil {
			ta.onKeyup(ta.value, e.Key)
		}
	})

	ta.tree.AddHandler(ta.id, event.Focus, func(e *event.Event) {
		if ta.onFocus != nil {
			ta.onFocus(ta.value)
		}
	})

	ta.tree.AddHandler(ta.id, event.Blur, func(e *event.Event) {
		if ta.onBlur != nil {
			ta.onBlur(ta.value)
		}
	})

	// IME support
	ta.tree.AddHandler(ta.id, event.IMECompositionEnd, func(e *event.Event) {
		if ta.disabled || ta.readonly || e.Text == "" {
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

func (ta *TextArea) SetStatus(s Status)       { ta.status = s }
func (ta *TextArea) SetMaxLength(n int)        { ta.maxLength = n }
func (ta *TextArea) SetTips(t string)          { ta.tips = t }
func (ta *TextArea) SetShowLimitNumber(s bool) { ta.showLimitNumber = s }
func (ta *TextArea) Readonly() bool            { return ta.readonly }
func (ta *TextArea) SetReadonly(r bool)        { ta.readonly = r }
func (ta *TextArea) AllowInputOverMax() bool   { return ta.allowInputOverMax }
func (ta *TextArea) SetAllowInputOverMax(v bool) { ta.allowInputOverMax = v }
func (ta *TextArea) Autosize() bool            { return ta.autosize }
func (ta *TextArea) SetAutosize(v bool)        { ta.autosize = v }
func (ta *TextArea) AutosizeMinRows() int      { return ta.autosizeMinRows }
func (ta *TextArea) AutosizeMaxRows() int      { return ta.autosizeMaxRows }
func (ta *TextArea) SetAutosizeRows(minRows, maxRows int) {
	ta.autosizeMinRows = minRows
	ta.autosizeMaxRows = maxRows
	ta.autosize = true
}

// TDesign event setters
func (ta *TextArea) OnBlur(fn func(string))                { ta.onBlur = fn }
func (ta *TextArea) OnFocus(fn func(string))               { ta.onFocus = fn }
func (ta *TextArea) OnKeydown(fn func(string, event.Key))  { ta.onKeydown = fn }
func (ta *TextArea) OnKeyup(fn func(string, event.Key))    { ta.onKeyup = fn }
func (ta *TextArea) OnValidate(fn func(string))            { ta.onValidate = fn }

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
	if ta.readonly {
		return
	}
	runes := []rune(ta.value)
	if ta.maxLength > 0 && len(runes) >= ta.maxLength && !ta.allowInputOverMax {
		if ta.onValidate != nil {
			ta.onValidate("exceed-maximum")
		}
		return
	}
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
	// Enforce maxLength
	if ta.maxLength > 0 {
		remaining := ta.maxLength - len(runes)
		if remaining <= 0 {
			return
		}
		if len(inserted) > remaining {
			inserted = inserted[:remaining]
		}
	}
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

// --- Visual line helpers (soft word wrap) ---

// visualLine represents one visual (possibly wrapped) line of text.
type visualLine struct {
	text      string // display text of this visual line
	runeStart int    // rune offset in ta.value where this line starts
	runeCount int    // number of runes in this visual line
}

// contentWidth returns the available text width inside the textarea.
func (ta *TextArea) contentWidth() float32 {
	bounds := ta.Bounds()
	return bounds.Width - ta.config.SpaceSM*2
}

// visualLines breaks the value into visual lines, wrapping at contentW.
func (ta *TextArea) visualLines(contentW float32) []visualLine {
	if ta.value == "" {
		return []visualLine{{text: "", runeStart: 0, runeCount: 0}}
	}

	logicalLines := strings.Split(ta.value, "\n")
	var result []visualLine
	runeOff := 0

	for i, ll := range logicalLines {
		if i > 0 {
			runeOff++ // account for \n
		}
		runes := []rune(ll)
		if len(runes) == 0 {
			result = append(result, visualLine{text: "", runeStart: runeOff, runeCount: 0})
			continue
		}

		if contentW <= 0 {
			// No layout yet, don't wrap
			result = append(result, visualLine{text: ll, runeStart: runeOff, runeCount: len(runes)})
			runeOff += len(runes)
			continue
		}

		// Break logical line into visual lines by character width
		start := 0
		for start < len(runes) {
			end := start
			for end < len(runes) {
				w := ta.measureRunes(runes[start : end+1])
				if w > contentW && end > start {
					break
				}
				end++
			}
			result = append(result, visualLine{
				text:      string(runes[start:end]),
				runeStart: runeOff + start,
				runeCount: end - start,
			})
			start = end
		}
		runeOff += len(runes)
	}

	return result
}

// runeOffsetToVisualLineCol maps a rune offset to (visual line index, column).
func (ta *TextArea) runeOffsetToVisualLineCol(pos int, contentW float32) (int, int) {
	vlines := ta.visualLines(contentW)
	rlen := ta.runeLen()
	if pos > rlen {
		pos = rlen
	}
	if pos < 0 {
		pos = 0
	}

	line := 0
	for i, vl := range vlines {
		if vl.runeStart <= pos {
			line = i
		} else {
			break
		}
	}
	return line, pos - vlines[line].runeStart
}

// visualLineColToRuneOffset maps (visual line index, column) to a rune offset.
func (ta *TextArea) visualLineColToRuneOffset(line, col int, contentW float32) int {
	vlines := ta.visualLines(contentW)
	if line < 0 {
		line = 0
	}
	if line >= len(vlines) {
		line = len(vlines) - 1
	}
	vl := vlines[line]
	if col < 0 {
		col = 0
	}
	if col > vl.runeCount {
		col = vl.runeCount
	}
	return vl.runeStart + col
}

// lineStart returns the rune offset of the start of the visual line containing pos.
func (ta *TextArea) lineStart(pos int) int {
	contentW := ta.contentWidth()
	line, _ := ta.runeOffsetToVisualLineCol(pos, contentW)
	return ta.visualLineColToRuneOffset(line, 0, contentW)
}

// lineEnd returns the rune offset of the end of the visual line containing pos.
func (ta *TextArea) lineEnd(pos int) int {
	contentW := ta.contentWidth()
	vlines := ta.visualLines(contentW)
	line, _ := ta.runeOffsetToVisualLineCol(pos, contentW)
	if line < len(vlines) {
		return ta.visualLineColToRuneOffset(line, vlines[line].runeCount, contentW)
	}
	return ta.runeLen()
}

// moveCursorVertical moves the cursor up (dir=-1) or down (dir=1).
func (ta *TextArea) moveCursorVertical(dir int) {
	contentW := ta.contentWidth()
	line, col := ta.runeOffsetToVisualLineCol(ta.cursorPos, contentW)
	newLine := line + dir
	ta.cursorPos = ta.visualLineColToRuneOffset(newLine, col, contentW)
}

// updateIMEPosition tells the OS where to place the IME candidate window.
func (ta *TextArea) updateIMEPosition() {
	cfg := ta.config
	if cfg.Window == nil {
		return
	}
	bounds := ta.Bounds()
	padLeft := cfg.SpaceSM
	padTop := cfg.SpaceSM
	lineH := cfg.FontSize * cfg.LineHeight
	contentW := bounds.Width - padLeft*2

	vlines := ta.visualLines(contentW)
	vLine, vCol := ta.runeOffsetToVisualLineCol(ta.cursorPos, contentW)
	cx := float32(0)
	if vLine < len(vlines) {
		lineRunes := []rune(vlines[vLine].text)
		if vCol > len(lineRunes) {
			vCol = len(lineRunes)
		}
		cx = ta.measureRunes(lineRunes[:vCol])
	}

	var lh float32
	if cfg.TextRenderer != nil {
		lh = cfg.TextRenderer.LineHeight(cfg.FontSize)
	} else {
		lh = cfg.FontSize * 1.2
	}

	cursorX := bounds.X + padLeft + cx
	cursorY := bounds.Y + padTop + float32(vLine)*lineH + (lineH-lh)/2

	cfg.Window.SetIMEPosition(uimath.NewRect(cursorX, cursorY, 1, lh))
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
	contentW := bounds.Width - padLeft*2

	relX := localX - bounds.X - padLeft
	relY := localY - bounds.Y - padTop
	if relY < 0 {
		relY = 0
	}

	vlines := ta.visualLines(contentW)
	lineIdx := int(relY / lineH)
	if lineIdx >= len(vlines) {
		lineIdx = len(vlines) - 1
	}
	if lineIdx < 0 {
		lineIdx = 0
	}

	// Find character within visual line
	lineRunes := []rune(vlines[lineIdx].text)
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
	return ta.visualLineColToRuneOffset(lineIdx, col, contentW)
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

	// Border: status overrides default, then focus, then disabled
	borderClr := cfg.StatusBorderColor(ta.status)
	if ta.status == StatusDefault && focused {
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
		var vlines []visualLine
		if displayText == ta.value {
			vlines = ta.visualLines(contentW)
		} else {
			// Placeholder: no wrapping needed, just split on newlines
			for _, l := range strings.Split(displayText, "\n") {
				vlines = append(vlines, visualLine{text: l})
			}
		}
		for i, vl := range vlines {
			y := bounds.Y + padTop + float32(i)*lineH
			if y > bounds.Y+bounds.Height {
				break
			}
			if vl.text == "" {
				continue
			}
			if cfg.TextRenderer != nil {
				lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
				ty := y + (lineH-lh)/2
				cfg.TextRenderer.DrawText(buf, vl.text, bounds.X+padLeft, ty, cfg.FontSize, contentW, textColor, 1)
			} else {
				textW := float32(len([]rune(vl.text))) * cfg.FontSize * 0.55
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
			vlines := ta.visualLines(contentW)
			vLine, vCol := ta.runeOffsetToVisualLineCol(ta.cursorPos, contentW)
			cx := float32(0)
			if vLine < len(vlines) {
				lineRunes := []rune(vlines[vLine].text)
				if vCol > len(lineRunes) {
					vCol = len(lineRunes)
				}
				cx = ta.measureRunes(lineRunes[:vCol])
			}
			cursorX := bounds.X + padLeft + cx
			cursorY := bounds.Y + padTop + float32(vLine)*lineH
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

	// Draw character counter at bottom-right
	if ta.showLimitNumber && ta.maxLength > 0 {
		counterText := fmt.Sprintf("%d/%d", len([]rune(ta.value)), ta.maxLength)
		counterColor := cfg.DisabledColor
		if len([]rune(ta.value)) > ta.maxLength {
			counterColor = cfg.ErrorColor
		}
		counterFontSize := cfg.FontSizeSm
		counterW := float32(len(counterText)) * counterFontSize * 0.55
		counterX := bounds.X + bounds.Width - cfg.SpaceSM - counterW
		counterY := bounds.Y + bounds.Height - cfg.SpaceSM - counterFontSize*1.2
		if cfg.TextRenderer != nil {
			cfg.TextRenderer.DrawText(buf, counterText, counterX, counterY, counterFontSize, counterW, counterColor, 1)
		} else {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(counterX, counterY, counterW, counterFontSize*1.2),
				FillColor: counterColor,
				Corners:   uimath.CornersAll(2),
			}, 3, 1)
		}
	}

	// Draw tips text below the textarea
	if ta.tips != "" {
		tipsColor := cfg.StatusBorderColor(ta.status)
		if ta.status == StatusDefault {
			tipsColor = cfg.DisabledColor
		}
		tipsFontSize := cfg.FontSizeSm
		tipsY := bounds.Y + bounds.Height + 4
		if cfg.TextRenderer != nil {
			cfg.TextRenderer.DrawText(buf, ta.tips, bounds.X, tipsY, tipsFontSize, bounds.Width, tipsColor, 1)
		} else {
			tipsW := float32(len([]rune(ta.tips))) * tipsFontSize * 0.55
			tipsH := tipsFontSize * 1.2
			if tipsW > bounds.Width {
				tipsW = bounds.Width
			}
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X, tipsY, tipsW, tipsH),
				FillColor: tipsColor,
				Corners:   uimath.CornersAll(2),
			}, 3, 1)
		}
	}
}

func (ta *TextArea) drawSelection(buf *render.CommandBuffer, bounds uimath.Rect, padLeft, padTop, lineH, contentW float32) {
	cfg := ta.config
	lo, hi := ta.selMin(), ta.selMax()
	loLine, loCol := ta.runeOffsetToVisualLineCol(lo, contentW)
	hiLine, hiCol := ta.runeOffsetToVisualLineCol(hi, contentW)

	selColor := uimath.ColorHex("#1677ff")
	selColor.A = 0.3

	vlines := ta.visualLines(contentW)

	var lh float32
	if cfg.TextRenderer != nil {
		lh = cfg.TextRenderer.LineHeight(cfg.FontSize)
	} else {
		lh = cfg.FontSize * 1.2
	}

	for lineIdx := loLine; lineIdx <= hiLine && lineIdx < len(vlines); lineIdx++ {
		lineRunes := []rune(vlines[lineIdx].text)
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

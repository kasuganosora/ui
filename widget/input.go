package widget

import (
	"fmt"
	"time"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/render"
)

// Input is a single-line text input widget with selection and clipboard support.
type Input struct {
	Base
	value       string
	placeholder string
	disabled    bool
	cursorPos   int // cursor position in runes
	selAnchor   int // selection anchor in runes (-1 = no selection)

	dragging     bool // mouse drag selection in progress
	dragSelected bool // true if the last drag produced a selection

	size            Size
	status          Status
	clearable       bool
	maxLength       int // 0 = unlimited
	readonly        bool
	tips            string
	showLimitNumber bool
	align           InputAlign
	allowInputOverMax bool
	borderless      bool
	label           string
	inputType       InputType

	onChange         func(value string)
	onBlur           func(value string)
	onFocus          func(value string)
	onClear          func()
	onEnter          func(value string)
	onKeydown        func(value string, key event.Key)
	onKeyup          func(value string, key event.Key)
	onValidate       func(error string)
}

// InputAlign controls text alignment within the input.
type InputAlign uint8

const (
	InputAlignLeft   InputAlign = iota // default
	InputAlignCenter
	InputAlignRight
)

// InputType controls the input type.
type InputType uint8

const (
	InputTypeText     InputType = iota // default
	InputTypePassword
	InputTypeNumber
)

// NewInput creates a text input widget.
func NewInput(tree *core.Tree, cfg *Config) *Input {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	inp := &Input{
		Base:      NewBase(tree, core.TypeInput, cfg),
		selAnchor: -1,
	}
	inp.style.Display = layout.DisplayFlex
	inp.style.AlignItems = layout.AlignCenter
	inp.style.Height = layout.Px(cfg.InputHeight)
	inp.style.Padding = layout.EdgeValues{
		Left:  layout.Px(cfg.SpaceSM),
		Right: layout.Px(cfg.SpaceSM),
	}

	inp.tree.AddHandler(inp.id, event.MouseClick, func(e *event.Event) {
		if inp.disabled {
			return
		}

		// Check if click is on clear button
		if inp.clearable && inp.value != "" && !inp.readonly {
			b := inp.Bounds()
			clearBtnSize := cfg.SizeFontSize(inp.size)
			clearX := b.X + b.Width - cfg.SpaceSM - clearBtnSize
			if e.X >= clearX && e.X <= clearX+clearBtnSize {
				inp.value = ""
				inp.cursorPos = 0
				inp.selAnchor = -1
				inp.tree.SetProperty(inp.id, "text", "")
				if inp.onChange != nil {
					inp.onChange("")
				}
				if inp.onClear != nil {
					inp.onClear()
				}
				return
			}
		}

		inp.tree.SetFocused(inp.id, true)

		// Don't clear selection if we just finished a drag-select
		if inp.dragSelected {
			inp.dragSelected = false
			return
		}

		pos := inp.hitTestChar(e.X)
		if e.Modifiers.Shift && inp.selAnchor >= 0 {
			inp.cursorPos = pos
		} else {
			inp.cursorPos = pos
			inp.selAnchor = -1
		}
		inp.updateIMEPosition()
	})

	inp.tree.AddHandler(inp.id, event.MouseDown, func(e *event.Event) {
		if inp.disabled || e.Button != event.MouseButtonLeft {
			return
		}
		pos := inp.hitTestChar(e.X)
		if e.Modifiers.Shift {
			if inp.selAnchor < 0 {
				inp.selAnchor = inp.cursorPos
			}
			inp.cursorPos = pos
		} else {
			inp.cursorPos = pos
			inp.selAnchor = pos
			inp.dragging = true
		}
	})

	inp.tree.AddHandler(inp.id, event.MouseMove, func(e *event.Event) {
		if !inp.dragging {
			return
		}
		pos := inp.hitTestChar(e.X)
		inp.cursorPos = pos
	})

	inp.tree.AddHandler(inp.id, event.MouseUp, func(e *event.Event) {
		if inp.dragging {
			inp.dragging = false
			if inp.selAnchor == inp.cursorPos {
				inp.selAnchor = -1
				inp.dragSelected = false
			} else {
				inp.dragSelected = true
			}
		}
	})

	// Right-click context menu
	inp.tree.AddHandler(inp.id, event.MouseDown, func(e *event.Event) {
		if inp.disabled || e.Button != event.MouseButtonRight {
			return
		}
		inp.tree.SetFocused(inp.id, true)
		inp.showContextMenu(int(e.GlobalX), int(e.GlobalY))
	})

	inp.tree.AddHandler(inp.id, event.KeyPress, func(e *event.Event) {
		if inp.disabled {
			return
		}
		// Ctrl+key combos are handled in KeyDown, not KeyPress
		if e.Modifiers.Ctrl {
			return
		}
		if e.Char != 0 {
			inp.deleteSelection()
			inp.insertChar(e.Char)
			inp.updateIMEPosition()
		}
	})

	inp.tree.AddHandler(inp.id, event.KeyDown, func(e *event.Event) {
		if inp.disabled {
			return
		}

		// Ctrl shortcuts
		if e.Modifiers.Ctrl {
			switch e.Key {
			case event.KeyA:
				inp.selectAll()
			case event.KeyC:
				inp.copySelection()
			case event.KeyX:
				inp.cutSelection()
			case event.KeyV:
				inp.paste()
				inp.updateIMEPosition()
			}
			return
		}

		shift := e.Modifiers.Shift
		switch e.Key {
		case event.KeyBackspace:
			if inp.hasSelection() {
				inp.deleteSelection()
			} else {
				inp.deleteBack()
			}
			inp.updateIMEPosition()
		case event.KeyDelete:
			if inp.hasSelection() {
				inp.deleteSelection()
			} else {
				inp.deleteForward()
			}
			inp.updateIMEPosition()
		case event.KeyArrowLeft:
			if shift {
				inp.startSelection()
				if inp.cursorPos > 0 {
					inp.cursorPos--
				}
			} else {
				if inp.hasSelection() {
					inp.cursorPos = inp.selMin()
					inp.selAnchor = -1
				} else if inp.cursorPos > 0 {
					inp.cursorPos--
				}
			}
			inp.updateIMEPosition()
		case event.KeyArrowRight:
			if shift {
				inp.startSelection()
				if inp.cursorPos < inp.runeLen() {
					inp.cursorPos++
				}
			} else {
				if inp.hasSelection() {
					inp.cursorPos = inp.selMax()
					inp.selAnchor = -1
				} else if inp.cursorPos < inp.runeLen() {
					inp.cursorPos++
				}
			}
			inp.updateIMEPosition()
		case event.KeyHome:
			if shift {
				inp.startSelection()
			} else {
				inp.selAnchor = -1
			}
			inp.cursorPos = 0
			inp.updateIMEPosition()
		case event.KeyEnter:
			if inp.onEnter != nil {
				inp.onEnter(inp.value)
			}
		case event.KeyEnd:
			if shift {
				inp.startSelection()
			} else {
				inp.selAnchor = -1
			}
			inp.cursorPos = inp.runeLen()
			inp.updateIMEPosition()
		}

		// Fire onKeydown callback
		if inp.onKeydown != nil {
			inp.onKeydown(inp.value, e.Key)
		}
	})

	inp.tree.AddHandler(inp.id, event.KeyUp, func(e *event.Event) {
		if inp.disabled {
			return
		}
		if inp.onKeyup != nil {
			inp.onKeyup(inp.value, e.Key)
		}
	})

	inp.tree.AddHandler(inp.id, event.Focus, func(e *event.Event) {
		if inp.onFocus != nil {
			inp.onFocus(inp.value)
		}
	})

	inp.tree.AddHandler(inp.id, event.Blur, func(e *event.Event) {
		if inp.onBlur != nil {
			inp.onBlur(inp.value)
		}
	})

	return inp
}

func (inp *Input) Value() string            { return inp.value }
func (inp *Input) Placeholder() string      { return inp.placeholder }
func (inp *Input) IsDisabled() bool         { return inp.disabled }
func (inp *Input) CursorPos() int           { return inp.cursorPos }
func (inp *Input) Selection() (int, int) {
	if inp.selAnchor < 0 {
		return inp.cursorPos, inp.cursorPos
	}
	return inp.selMin(), inp.selMax()
}

func (inp *Input) SetValue(v string) {
	inp.value = v
	if inp.cursorPos > len([]rune(v)) {
		inp.cursorPos = len([]rune(v))
	}
	inp.selAnchor = -1
	inp.tree.SetProperty(inp.id, "text", v)
}

func (inp *Input) SetPlaceholder(p string) {
	inp.placeholder = p
}

func (inp *Input) SetDisabled(d bool) {
	inp.disabled = d
	inp.tree.SetEnabled(inp.id, !d)
}

func (inp *Input) OnChange(fn func(string)) {
	inp.onChange = fn
}

func (inp *Input) SetSize(s Size) {
	inp.size = s
	inp.style.Height = layout.Px(inp.config.SizeHeight(s))
}

func (inp *Input) SetStatus(s Status)          { inp.status = s }
func (inp *Input) SetClearable(c bool)          { inp.clearable = c }
func (inp *Input) SetMaxLength(n int)           { inp.maxLength = n }
func (inp *Input) SetReadonly(r bool)           { inp.readonly = r }
func (inp *Input) SetTips(t string)             { inp.tips = t }
func (inp *Input) SetShowLimitNumber(s bool)    { inp.showLimitNumber = s }

// TDesign additional prop getters/setters
func (inp *Input) Align() InputAlign             { return inp.align }
func (inp *Input) SetAlign(a InputAlign)         { inp.align = a }
func (inp *Input) AllowInputOverMax() bool       { return inp.allowInputOverMax }
func (inp *Input) SetAllowInputOverMax(v bool)   { inp.allowInputOverMax = v }
func (inp *Input) Borderless() bool              { return inp.borderless }
func (inp *Input) SetBorderless(v bool)          { inp.borderless = v }
func (inp *Input) InputLabel() string            { return inp.label }
func (inp *Input) SetLabel(l string)             { inp.label = l }
func (inp *Input) Type() InputType               { return inp.inputType }
func (inp *Input) SetType(t InputType)           { inp.inputType = t }

// TDesign event setters
func (inp *Input) OnBlur(fn func(string))                  { inp.onBlur = fn }
func (inp *Input) OnFocus(fn func(string))                 { inp.onFocus = fn }
func (inp *Input) OnClear(fn func())                       { inp.onClear = fn }
func (inp *Input) OnEnter(fn func(string))                 { inp.onEnter = fn }
func (inp *Input) OnKeydown(fn func(string, event.Key))    { inp.onKeydown = fn }
func (inp *Input) OnKeyup(fn func(string, event.Key))      { inp.onKeyup = fn }
func (inp *Input) OnValidate(fn func(string))              { inp.onValidate = fn }

// --- Selection helpers ---

func (inp *Input) runeLen() int { return len([]rune(inp.value)) }

func (inp *Input) hasSelection() bool {
	return inp.selAnchor >= 0 && inp.selAnchor != inp.cursorPos
}

func (inp *Input) selMin() int {
	if inp.selAnchor < inp.cursorPos {
		return inp.selAnchor
	}
	return inp.cursorPos
}

func (inp *Input) selMax() int {
	if inp.selAnchor > inp.cursorPos {
		return inp.selAnchor
	}
	return inp.cursorPos
}

func (inp *Input) startSelection() {
	if inp.selAnchor < 0 {
		inp.selAnchor = inp.cursorPos
	}
}

func (inp *Input) selectAll() {
	inp.selAnchor = 0
	inp.cursorPos = inp.runeLen()
}

func (inp *Input) selectedText() string {
	if !inp.hasSelection() {
		return ""
	}
	runes := []rune(inp.value)
	lo, hi := inp.selMin(), inp.selMax()
	if hi > len(runes) {
		hi = len(runes)
	}
	return string(runes[lo:hi])
}

func (inp *Input) deleteSelection() {
	if !inp.hasSelection() {
		return
	}
	runes := []rune(inp.value)
	lo, hi := inp.selMin(), inp.selMax()
	if hi > len(runes) {
		hi = len(runes)
	}
	runes = append(runes[:lo], runes[hi:]...)
	inp.value = string(runes)
	inp.cursorPos = lo
	inp.selAnchor = -1
	inp.tree.SetProperty(inp.id, "text", inp.value)
	if inp.onChange != nil {
		inp.onChange(inp.value)
	}
}

// --- Clipboard ---

func (inp *Input) copySelection() {
	text := inp.selectedText()
	if text == "" {
		return
	}
	if inp.config.Platform != nil {
		inp.config.Platform.SetClipboardText(text)
	}
}

func (inp *Input) cutSelection() {
	inp.copySelection()
	inp.deleteSelection()
}

func (inp *Input) paste() {
	if inp.readonly {
		return
	}
	if inp.config.Platform == nil {
		return
	}
	text := inp.config.Platform.GetClipboardText()
	if text == "" {
		return
	}
	inp.deleteSelection()
	// Insert pasted text character by character (to filter newlines)
	runes := []rune(inp.value)
	pos := inp.cursorPos
	if pos > len(runes) {
		pos = len(runes)
	}
	var filtered []rune
	for _, r := range text {
		if r != '\n' && r != '\r' {
			filtered = append(filtered, r)
		}
	}
	// Enforce maxLength
	if inp.maxLength > 0 {
		remaining := inp.maxLength - len(runes)
		if remaining <= 0 {
			return
		}
		if len(filtered) > remaining {
			filtered = filtered[:remaining]
		}
	}
	runes = append(runes[:pos], append(filtered, runes[pos:]...)...)
	inp.value = string(runes)
	inp.cursorPos = pos + len(filtered)
	inp.selAnchor = -1
	inp.tree.SetProperty(inp.id, "text", inp.value)
	if inp.onChange != nil {
		inp.onChange(inp.value)
	}
}

// --- Context menu ---

func (inp *Input) showContextMenu(clientX, clientY int) {
	win := inp.config.Window
	if win == nil {
		return
	}

	hasSel := inp.hasSelection()
	hasClip := false
	if inp.config.Platform != nil {
		hasClip = inp.config.Platform.GetClipboardText() != ""
	}

	items := []platform.ContextMenuItem{
		{Label: "剪切(X)", Enabled: hasSel},
		{Label: "复制(C)", Enabled: hasSel},
		{Label: "粘贴(V)", Enabled: hasClip},
		{Label: "全选(A)", Enabled: inp.runeLen() > 0},
	}

	idx := win.ShowContextMenu(clientX, clientY, items)
	switch idx {
	case 0: // Cut
		inp.cutSelection()
	case 1: // Copy
		inp.copySelection()
	case 2: // Paste
		inp.paste()
	case 3: // Select All
		inp.selectAll()
	}
}

// --- Text editing ---

func (inp *Input) insertChar(ch rune) {
	if inp.readonly {
		return
	}
	runes := []rune(inp.value)
	if inp.maxLength > 0 && len(runes) >= inp.maxLength && !inp.allowInputOverMax {
		if inp.onValidate != nil {
			inp.onValidate("exceed-maximum")
		}
		return
	}
	pos := inp.cursorPos
	if pos > len(runes) {
		pos = len(runes)
	}
	runes = append(runes[:pos], append([]rune{ch}, runes[pos:]...)...)
	inp.value = string(runes)
	inp.cursorPos = pos + 1
	inp.selAnchor = -1
	inp.tree.SetProperty(inp.id, "text", inp.value)
	if inp.onChange != nil {
		inp.onChange(inp.value)
	}
}

func (inp *Input) deleteBack() {
	if inp.readonly || inp.cursorPos <= 0 {
		return
	}
	runes := []rune(inp.value)
	pos := inp.cursorPos
	if pos > len(runes) {
		pos = len(runes)
	}
	runes = append(runes[:pos-1], runes[pos:]...)
	inp.value = string(runes)
	inp.cursorPos = pos - 1
	inp.tree.SetProperty(inp.id, "text", inp.value)
	if inp.onChange != nil {
		inp.onChange(inp.value)
	}
}

func (inp *Input) deleteForward() {
	if inp.readonly {
		return
	}
	runes := []rune(inp.value)
	pos := inp.cursorPos
	if pos >= len(runes) {
		return
	}
	runes = append(runes[:pos], runes[pos+1:]...)
	inp.value = string(runes)
	inp.tree.SetProperty(inp.id, "text", inp.value)
	if inp.onChange != nil {
		inp.onChange(inp.value)
	}
}

// --- Hit testing ---

// hitTestChar returns the rune index closest to the given x coordinate (element-local).
func (inp *Input) hitTestChar(localX float32) int {
	bounds := inp.Bounds()
	padLeft := inp.config.SpaceSM
	relX := localX - bounds.X - padLeft
	if relX <= 0 {
		return 0
	}

	runes := []rune(inp.value)
	cfg := inp.config

	for i := 1; i <= len(runes); i++ {
		w := inp.measureRunes(runes[:i])
		if cfg.TextRenderer != nil {
			// Check midpoint of the character
			prevW := float32(0)
			if i > 0 {
				prevW = inp.measureRunes(runes[:i-1])
			}
			mid := prevW + (w-prevW)/2
			if relX < mid {
				return i - 1
			}
		} else {
			if relX < w {
				return i - 1
			}
		}
	}
	return len(runes)
}

func (inp *Input) measureRunes(runes []rune) float32 {
	cfg := inp.config
	if cfg.TextRenderer != nil {
		return cfg.TextRenderer.MeasureText(string(runes), cfg.FontSize)
	}
	return float32(len(runes)) * cfg.FontSize * 0.55
}

// cursorX returns the X offset of the cursor within the text area.
func (inp *Input) cursorX() float32 {
	runes := []rune(inp.value)
	pos := inp.cursorPos
	if pos > len(runes) {
		pos = len(runes)
	}
	return inp.measureRunes(runes[:pos])
}

// textX returns the X offset for the given rune position.
func (inp *Input) textX(pos int) float32 {
	runes := []rune(inp.value)
	if pos > len(runes) {
		pos = len(runes)
	}
	return inp.measureRunes(runes[:pos])
}

// updateIMEPosition tells the OS where to place the IME candidate window.
func (inp *Input) updateIMEPosition() {
	cfg := inp.config
	if cfg.Window == nil {
		return
	}
	bounds := inp.Bounds()
	padLeft := cfg.SpaceSM
	cx := bounds.X + padLeft + inp.cursorX()

	var lh float32
	if cfg.TextRenderer != nil {
		lh = cfg.TextRenderer.LineHeight(cfg.FontSize)
	} else {
		lh = cfg.FontSize * 1.2
	}
	cy := bounds.Y + (bounds.Height-lh)/2

	cfg.Window.SetIMEPosition(uimath.NewRect(cx, cy, 1, lh))
}

func (inp *Input) Draw(buf *render.CommandBuffer) {
	bounds := inp.Bounds()
	if bounds.IsEmpty() {
		return
	}

	cfg := inp.config
	fontSize := cfg.SizeFontSize(inp.size)
	elem := inp.Element()
	focused := elem != nil && elem.IsFocused()
	hovered := elem != nil && elem.IsHovered()

	// Border color: status overrides default, then focus, then disabled
	borderClr := cfg.StatusBorderColor(inp.status)
	if inp.status == StatusDefault && focused {
		borderClr = cfg.FocusBorderColor
	}
	if inp.disabled {
		borderClr = cfg.DisabledColor
	}

	bgClr := cfg.BgColor
	if inp.disabled {
		bgClr = uimath.ColorHex("#f5f5f5")
	} else if inp.readonly {
		bgClr = uimath.ColorHex("#fafafa")
	}

	// Background + border
	buf.DrawRect(render.RectCmd{
		Bounds:      bounds,
		FillColor:   bgClr,
		BorderColor: borderClr,
		BorderWidth: cfg.BorderWidth,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 0, 1)

	padLeft := cfg.SpaceSM
	// Reserve space on right for clear button and/or limit counter
	padRight := cfg.SpaceSM
	rightExtra := float32(0)

	// Clear button area
	clearBtnSize := fontSize
	showClear := inp.clearable && inp.value != "" && (hovered || focused) && !inp.disabled && !inp.readonly
	if showClear {
		rightExtra += clearBtnSize + cfg.SpaceXS
	}

	// Limit counter area
	if inp.showLimitNumber && inp.maxLength > 0 {
		counterText := fmt.Sprintf("%d/%d", len([]rune(inp.value)), inp.maxLength)
		counterW := float32(len(counterText)) * fontSize * 0.55
		rightExtra += counterW + cfg.SpaceXS
	}

	// Selection highlight
	if focused && inp.hasSelection() {
		lo, hi := inp.selMin(), inp.selMax()
		selStartX := bounds.X + padLeft + inp.textX(lo)
		selEndX := bounds.X + padLeft + inp.textX(hi)

		var lh float32
		if cfg.TextRenderer != nil {
			lh = cfg.TextRenderer.LineHeight(fontSize)
		} else {
			lh = fontSize * 1.2
		}
		selY := bounds.Y + (bounds.Height-lh)/2

		selColor := uimath.ColorHex("#1677ff")
		selColor.A = 0.3

		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(selStartX, selY, selEndX-selStartX, lh),
			FillColor: selColor,
		}, 1, 1)
	}

	// Text or placeholder
	displayText := inp.value
	textColor := cfg.TextColor
	if displayText == "" && inp.placeholder != "" {
		displayText = inp.placeholder
		textColor = cfg.DisabledColor
	}

	if displayText != "" {
		if cfg.TextRenderer != nil {
			tx := bounds.X + padLeft
			lh := cfg.TextRenderer.LineHeight(fontSize)
			ty := bounds.Y + (bounds.Height-lh)/2
			maxW := bounds.Width - padLeft - padRight - rightExtra
			cfg.TextRenderer.DrawText(buf, displayText, tx, ty, fontSize, maxW, textColor, 1)
		} else {
			textW := float32(len(displayText)) * fontSize * 0.55
			textH := fontSize * 1.2
			maxW := bounds.Width - padLeft - padRight - rightExtra
			if textW > maxW {
				textW = maxW
			}
			ty := bounds.Y + (bounds.Height-textH)/2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X+padLeft, ty, textW, textH),
				FillColor: textColor,
				Corners:   uimath.CornersAll(2),
			}, 1, 1)
		}
	}

	// Draw limit counter on right side
	if inp.showLimitNumber && inp.maxLength > 0 {
		counterText := fmt.Sprintf("%d/%d", len([]rune(inp.value)), inp.maxLength)
		counterColor := cfg.DisabledColor
		if len([]rune(inp.value)) > inp.maxLength {
			counterColor = cfg.ErrorColor
		}
		counterW := float32(len(counterText)) * fontSize * 0.55
		counterX := bounds.X + bounds.Width - padRight - counterW
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(fontSize)
			counterY := bounds.Y + (bounds.Height-lh)/2
			cfg.TextRenderer.DrawText(buf, counterText, counterX, counterY, fontSize, counterW, counterColor, 1)
		} else {
			textH := fontSize * 1.2
			counterY := bounds.Y + (bounds.Height-textH)/2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(counterX, counterY, counterW, textH),
				FillColor: counterColor,
				Corners:   uimath.CornersAll(2),
			}, 1, 1)
		}
	}

	// Draw clear "x" button
	if showClear {
		clearX := bounds.X + bounds.Width - padRight - clearBtnSize
		if inp.showLimitNumber && inp.maxLength > 0 {
			counterText := fmt.Sprintf("%d/%d", len([]rune(inp.value)), inp.maxLength)
			counterW := float32(len(counterText)) * fontSize * 0.55
			clearX = bounds.X + bounds.Width - padRight - counterW - cfg.SpaceXS - clearBtnSize
		}
		clearY := bounds.Y + (bounds.Height-clearBtnSize)/2
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(fontSize)
			clearTY := bounds.Y + (bounds.Height-lh)/2
			cfg.TextRenderer.DrawText(buf, "\u00d7", clearX, clearTY, fontSize, clearBtnSize, cfg.DisabledColor, 1)
		} else {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(clearX, clearY, clearBtnSize, clearBtnSize),
				FillColor: cfg.DisabledColor,
				Corners:   uimath.CornersAll(clearBtnSize / 2),
			}, 1, 1)
		}
	}

	// Blinking cursor when focused
	if focused && !inp.disabled && !inp.readonly {
		ms := time.Now().UnixMilli()
		if (ms/500)%2 == 0 {
			cx := bounds.X + padLeft + inp.cursorX()
			var lh float32
			if cfg.TextRenderer != nil {
				lh = cfg.TextRenderer.LineHeight(fontSize)
			} else {
				lh = fontSize * 1.2
			}
			cursorH := lh
			cursorY := bounds.Y + (bounds.Height-cursorH)/2

			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(cx, cursorY, 1, cursorH),
				FillColor: cfg.TextColor,
			}, 2, 1)
		}
	}

	// Draw tips text below the input
	if inp.tips != "" {
		tipsColor := cfg.StatusBorderColor(inp.status)
		if inp.status == StatusDefault {
			tipsColor = cfg.DisabledColor
		}
		tipsFontSize := cfg.FontSizeSm
		tipsY := bounds.Y + bounds.Height + 4
		if cfg.TextRenderer != nil {
			cfg.TextRenderer.DrawText(buf, inp.tips, bounds.X, tipsY, tipsFontSize, bounds.Width, tipsColor, 1)
		} else {
			tipsW := float32(len([]rune(inp.tips))) * tipsFontSize * 0.55
			tipsH := tipsFontSize * 1.2
			if tipsW > bounds.Width {
				tipsW = bounds.Width
			}
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X, tipsY, tipsW, tipsH),
				FillColor: tipsColor,
				Corners:   uimath.CornersAll(2),
			}, 1, 1)
		}
	}
}


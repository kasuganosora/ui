package widget

import (
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

	dragging    bool // mouse drag selection in progress
	dragSelected bool // true if the last drag produced a selection

	onChange func(value string)
}

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
		case event.KeyEnd:
			if shift {
				inp.startSelection()
			} else {
				inp.selAnchor = -1
			}
			inp.cursorPos = inp.runeLen()
			inp.updateIMEPosition()
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
	runes := []rune(inp.value)
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
	if inp.cursorPos <= 0 {
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
	elem := inp.Element()
	focused := elem != nil && elem.IsFocused()

	// Border color based on state
	borderClr := cfg.BorderColor
	if focused {
		borderClr = cfg.FocusBorderColor
	}
	if inp.disabled {
		borderClr = cfg.DisabledColor
	}

	bgClr := cfg.BgColor
	if inp.disabled {
		bgClr = uimath.ColorHex("#f5f5f5")
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

	// Selection highlight
	if focused && inp.hasSelection() {
		lo, hi := inp.selMin(), inp.selMax()
		selStartX := bounds.X + padLeft + inp.textX(lo)
		selEndX := bounds.X + padLeft + inp.textX(hi)

		var lh float32
		if cfg.TextRenderer != nil {
			lh = cfg.TextRenderer.LineHeight(cfg.FontSize)
		} else {
			lh = cfg.FontSize * 1.2
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
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			ty := bounds.Y + (bounds.Height-lh)/2
			maxW := bounds.Width - padLeft*2
			cfg.TextRenderer.DrawText(buf, displayText, tx, ty, cfg.FontSize, maxW, textColor, 1)
		} else {
			textW := float32(len(displayText)) * cfg.FontSize * 0.55
			textH := cfg.FontSize * 1.2
			maxW := bounds.Width - padLeft*2
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

	// Blinking cursor when focused
	if focused && !inp.disabled {
		// Blink at ~500ms intervals
		ms := time.Now().UnixMilli()
		if (ms/500)%2 == 0 {
			cx := bounds.X + padLeft + inp.cursorX()
			var lh float32
			if cfg.TextRenderer != nil {
				lh = cfg.TextRenderer.LineHeight(cfg.FontSize)
			} else {
				lh = cfg.FontSize * 1.2
			}
			cursorH := lh
			cursorY := bounds.Y + (bounds.Height-cursorH)/2

			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(cx, cursorY, 1, cursorH),
				FillColor: cfg.TextColor,
			}, 2, 1)
		}
	}
}


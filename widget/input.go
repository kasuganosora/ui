package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Input is a single-line text input widget.
type Input struct {
	Base
	value       string
	placeholder string
	disabled    bool
	cursorPos   int
	selStart    int
	selEnd      int

	onChange func(value string)
}

// NewInput creates a text input widget.
func NewInput(tree *core.Tree, cfg *Config) *Input {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	inp := &Input{
		Base: NewBase(tree, core.TypeInput, cfg),
	}
	inp.style.Display = layout.DisplayFlex
	inp.style.AlignItems = layout.AlignCenter
	inp.style.Height = layout.Px(cfg.InputHeight)
	inp.style.Padding = layout.EdgeValues{
		Left:  layout.Px(cfg.SpaceSM),
		Right: layout.Px(cfg.SpaceSM),
	}

	inp.tree.AddHandler(inp.id, event.MouseClick, func(e *event.Event) {
		if !inp.disabled {
			inp.tree.SetFocused(inp.id, true)
		}
	})

	inp.tree.AddHandler(inp.id, event.KeyPress, func(e *event.Event) {
		if inp.disabled {
			return
		}
		if e.Char != 0 {
			inp.insertChar(e.Char)
		}
	})

	inp.tree.AddHandler(inp.id, event.KeyDown, func(e *event.Event) {
		if inp.disabled {
			return
		}
		switch e.Key {
		case event.KeyBackspace:
			inp.deleteBack()
		case event.KeyDelete:
			inp.deleteForward()
		case event.KeyArrowLeft:
			if inp.cursorPos > 0 {
				inp.cursorPos--
			}
		case event.KeyArrowRight:
			if inp.cursorPos < len(inp.value) {
				inp.cursorPos++
			}
		case event.KeyHome:
			inp.cursorPos = 0
		case event.KeyEnd:
			inp.cursorPos = len(inp.value)
		}
	})

	return inp
}

func (inp *Input) Value() string       { return inp.value }
func (inp *Input) Placeholder() string { return inp.placeholder }
func (inp *Input) IsDisabled() bool    { return inp.disabled }
func (inp *Input) CursorPos() int      { return inp.cursorPos }
func (inp *Input) Selection() (int, int) { return inp.selStart, inp.selEnd }

func (inp *Input) SetValue(v string) {
	inp.value = v
	if inp.cursorPos > len(v) {
		inp.cursorPos = len(v)
	}
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

func (inp *Input) insertChar(ch rune) {
	runes := []rune(inp.value)
	pos := inp.cursorPos
	if pos > len(runes) {
		pos = len(runes)
	}
	runes = append(runes[:pos], append([]rune{ch}, runes[pos:]...)...)
	inp.value = string(runes)
	inp.cursorPos = pos + 1
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

func (inp *Input) Draw(buf *render.CommandBuffer) {
	bounds := inp.Bounds()
	if bounds.IsEmpty() {
		return
	}

	cfg := inp.config
	elem := inp.Element()

	// Border color based on state
	borderClr := cfg.BorderColor
	if elem != nil && elem.IsFocused() {
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

	// Text or placeholder
	displayText := inp.value
	textColor := cfg.TextColor
	if displayText == "" && inp.placeholder != "" {
		displayText = inp.placeholder
		textColor = cfg.DisabledColor
	}

	if displayText != "" {
		padLeft := cfg.SpaceSM
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

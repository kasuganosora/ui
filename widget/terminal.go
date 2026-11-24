package widget

import (
	"unicode/utf8"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Terminal shows a UTF-8 byte stream and routes keyboard bytes to the host.
// It does not start a shell or PTY, and it does not interpret VT sequences.
type Terminal struct {
	Base
	text    string
	pending []byte
	scrollY float32
	lineH   float32
	pad     float32
	onBytes func([]byte)
}

// NewTerminal creates an empty stream view.
func NewTerminal(tree *core.Tree, cfg *Config) *Terminal {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	t := &Terminal{Base: NewBase(tree, core.TypeCustom, cfg), lineH: 18, pad: 4}
	tree.SetFocusable(t.id, true)
	tree.AddHandler(t.id, event.KeyDown, func(ev *event.Event) {
		if !targetPhase(ev) {
			return
		}
		switch ev.Key {
		case event.KeyEnter:
			t.emit([]byte{'\n'})
		case event.KeyBackspace:
			t.emit([]byte{0x7f})
		default:
			return
		}
		ev.StopPropagation()
	})
	tree.AddHandler(t.id, event.KeyPress, func(ev *event.Event) {
		if !targetPhase(ev) || ev.Char == 0 || ev.Modifiers.Ctrl {
			return
		}
		t.emit([]byte(string(ev.Char)))
		ev.StopPropagation()
	})
	return t
}

// OnBytes receives key bytes. The display changes only through Write.
func (t *Terminal) OnBytes(fn func([]byte)) { t.onBytes = fn }

func (t *Terminal) emit(b []byte) {
	if t.onBytes == nil || len(b) == 0 {
		return
	}
	t.onBytes(append([]byte(nil), b...))
}

// Write appends UTF-8 bytes. A split rune is kept until it is complete.
// Escape sequences stay in the text.
func (t *Terminal) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	t.pending = append(t.pending, p...)
	for len(t.pending) > 0 && utf8.FullRune(t.pending) {
		r, size := utf8.DecodeRune(t.pending)
		t.pending = t.pending[size:]
		if r == utf8.RuneError && size == 1 {
			r = '\uFFFD'
		}
		t.text += string(r)
	}
	return len(p), nil
}

// Text is the decoded stream. Incomplete trailing bytes are omitted.
func (t *Terminal) Text() string { return t.text }

// Clear drops the stream and the scroll offset.
func (t *Terminal) Clear() {
	t.text = ""
	t.pending = nil
	t.scrollY = 0
}

// SetScrollY moves the stream vertically.
func (t *Terminal) SetScrollY(y float32) {
	if y < 0 {
		y = 0
	}
	t.scrollY = y
}

// Draw paints each line as glyphs. Lines outside the widget are skipped.
func (t *Terminal) Draw(buf *render.CommandBuffer) {
	b := t.Bounds()
	if b.IsEmpty() {
		return
	}
	buf.PushClip(b)
	buf.DrawRect(render.RectCmd{Bounds: b, FillColor: uimath.ColorWhite}, 0, 1)
	lines := textLines(t.text)
	size := t.config.FontSizeSm
	for i, line := range lines {
		y := b.Y + float32(i)*t.lineH - t.scrollY
		if y+t.lineH <= b.Y {
			continue
		}
		if y >= b.Y+b.Height {
			break
		}
		if line == "" {
			continue
		}
		emitText(t.config, buf, line, b.X+t.pad, y+1, size, b.Width, t.config.TextColor, 1, 4)
	}
	buf.PopClip()
}

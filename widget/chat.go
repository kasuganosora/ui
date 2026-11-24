package widget

import (
	"fmt"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// ChatMessage is one transcript row. The widget draws it; it does not call a model.
type ChatMessage struct {
	ID     string
	Author string
	Text   string
	Mine   bool
}

// ChatSender receives a composer submission. The host injects it.
// The widget does not open a network connection.
type ChatSender interface {
	Send(text string) error
}

// Chat is a transcript plus a composer. Enter submits the draft.
// An empty draft and a sender error leave the transcript unchanged.
type Chat struct {
	Base
	messages  []ChatMessage
	draft     string
	selected  string
	scrollY   float32
	rowH      float32
	composerH float32
	pad       float32
	seq       int
	sender    ChatSender
}

// NewChat creates an empty conversation.
func NewChat(tree *core.Tree, cfg *Config) *Chat {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	c := &Chat{Base: NewBase(tree, core.TypeCustom, cfg), rowH: 36, composerH: 32, pad: 8}
	tree.SetFocusable(c.id, true)
	tree.AddHandler(c.id, event.KeyDown, func(ev *event.Event) {
		if !targetPhase(ev) {
			return
		}
		switch ev.Key {
		case event.KeyBackspace:
			c.backspace()
		case event.KeyEnter:
			c.commit()
		default:
			return
		}
		ev.StopPropagation()
	})
	tree.AddHandler(c.id, event.KeyPress, func(ev *event.Event) {
		if !targetPhase(ev) || ev.Char == 0 || ev.Modifiers.Ctrl {
			return
		}
		c.draft += string(ev.Char)
		ev.StopPropagation()
	})
	tree.AddHandler(c.id, event.MouseDown, func(ev *event.Event) {
		if !targetPhase(ev) {
			return
		}
		if id := c.messageAt(ev.GlobalX, ev.GlobalY); id != "" {
			c.selected = id
		}
	})
	return c
}

// SetSender installs the host delivery hook.
func (c *Chat) SetSender(s ChatSender) { c.sender = s }

// Messages returns a copy of the transcript.
func (c *Chat) Messages() []ChatMessage {
	out := make([]ChatMessage, len(c.messages))
	copy(out, c.messages)
	return out
}

// SetMessages replaces the transcript and clears the draft, selection, and scroll.
func (c *Chat) SetMessages(m []ChatMessage) {
	c.messages = append([]ChatMessage(nil), m...)
	c.draft = ""
	c.selected = ""
	c.scrollY = 0
}

// Append adds a row without calling the sender.
func (c *Chat) Append(m ChatMessage) { c.messages = append(c.messages, m) }

func (c *Chat) Draft() string      { return c.draft }
func (c *Chat) SelectedID() string { return c.selected }
func (c *Chat) ScrollY() float32   { return c.scrollY }

// SetScrollY moves the transcript. The composer stays put.
func (c *Chat) SetScrollY(y float32) {
	if y < 0 {
		y = 0
	}
	c.scrollY = y
}

// MessageScreen is the row rectangle for id, including rows outside the clip.
func (c *Chat) MessageScreen(id string) uimath.Rect {
	list := c.listBounds()
	for i, m := range c.messages {
		if m.ID == id {
			return uimath.NewRect(list.X, list.Y+float32(i)*c.rowH-c.scrollY, list.Width, c.rowH)
		}
	}
	return uimath.Rect{}
}

func (c *Chat) listBounds() uimath.Rect {
	b := c.Bounds()
	h := b.Height - c.composerH
	if h < 0 {
		h = 0
	}
	return uimath.NewRect(b.X, b.Y, b.Width, h)
}

func (c *Chat) backspace() {
	r := []rune(c.draft)
	if len(r) == 0 {
		return
	}
	c.draft = string(r[:len(r)-1])
}

func (c *Chat) commit() {
	if c.draft == "" {
		return
	}
	if c.sender != nil {
		if err := c.sender.Send(c.draft); err != nil {
			return
		}
	}
	c.seq++
	c.messages = append(c.messages, ChatMessage{
		ID: fmt.Sprintf("m%d", c.seq), Author: "me", Text: c.draft, Mine: true,
	})
	c.draft = ""
}

func (c *Chat) messageAt(x, y float32) string {
	list := c.listBounds()
	if x < list.X || y < list.Y || x >= list.X+list.Width || y >= list.Y+list.Height {
		return ""
	}
	i := int((y - list.Y + c.scrollY) / c.rowH)
	if i < 0 || i >= len(c.messages) {
		return ""
	}
	return c.messages[i].ID
}

func (c *Chat) measure(s string) float32 {
	size := c.config.FontSizeSm
	if c.config.TextRenderer != nil {
		return c.config.TextRenderer.MeasureText(s, size)
	}
	return float32(len([]rune(s))) * size * 0.6
}

// Draw paints author and body glyphs. The list and the composer have separate clips.
func (c *Chat) Draw(buf *render.CommandBuffer) {
	b := c.Bounds()
	if b.IsEmpty() {
		return
	}
	list := c.listBounds()
	buf.PushClip(b)
	buf.DrawRect(render.RectCmd{Bounds: b, FillColor: c.config.BgColor}, 0, 1)
	buf.PushClip(list)
	for i, m := range c.messages {
		y := list.Y + float32(i)*c.rowH - c.scrollY
		if y+c.rowH <= list.Y {
			continue
		}
		if y >= list.Y+list.Height {
			break
		}
		c.drawMessage(buf, m, uimath.NewRect(list.X, y, list.Width, c.rowH))
	}
	buf.PopClip()
	comp := uimath.NewRect(b.X, list.Y+list.Height, b.Width, c.composerH)
	buf.PushClip(comp)
	buf.DrawRect(render.RectCmd{Bounds: comp, FillColor: uimath.ColorWhite}, 1, 1)
	if c.draft != "" {
		emitText(c.config, buf, c.draft, comp.X+c.pad, comp.Y+6, c.config.FontSizeSm, comp.Width, c.config.TextColor, 1, 4)
	}
	buf.PopClip()
	buf.PopClip()
}

func (c *Chat) drawMessage(buf *render.CommandBuffer, m ChatMessage, row uimath.Rect) {
	cfg := c.config
	author := m.Author
	if author == "" {
		author = "me"
	}
	aw := c.measure(author)
	bw := c.measure(m.Text)
	ax := row.X + c.pad
	bx := ax
	if m.Mine {
		ax = row.X + row.Width - c.pad - aw
		bx = row.X + row.Width - c.pad - bw
		if ax < row.X+c.pad {
			ax = row.X + c.pad
		}
		if bx < row.X+c.pad {
			bx = row.X + c.pad
		}
	}
	emitText(cfg, buf, author, ax, row.Y+1, cfg.FontSizeSm, row.Width, cfg.TextColor, 1, 4)
	emitText(cfg, buf, m.Text, bx, row.Y+1, cfg.FontSizeSm, row.Width, cfg.TextColor, 1, 4)
}

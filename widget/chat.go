package widget

import (
	"fmt"
	"strings"

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
	composerH float32
	pad       float32
	seq       int
	sending   bool
	sender    ChatSender
}

type chatLine struct {
	text string
	x, y float32
	w    float32
}

type chatBlock struct {
	id    string
	row   uimath.Rect
	lines []chatLine
}

// NewChat creates an empty conversation.
func NewChat(tree *core.Tree, cfg *Config) *Chat {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	c := &Chat{Base: NewBase(tree, core.TypeCustom, cfg), composerH: 32, pad: 8}
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
	for _, b := range c.blocks() {
		if b.id == id {
			return b.row
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
	if c.sending || c.draft == "" {
		return
	}
	snap := c.draft
	c.sending = true
	var err error
	if c.sender != nil {
		err = c.sender.Send(snap)
	}
	if err != nil {
		c.sending = false
		return
	}
	c.seq++
	c.messages = append(c.messages, ChatMessage{
		ID: fmt.Sprintf("m%d", c.seq), Author: "me", Text: snap, Mine: true,
	})
	if strings.HasPrefix(c.draft, snap) {
		c.draft = c.draft[len(snap):]
	}
	c.sending = false
}

func (c *Chat) messageAt(x, y float32) string {
	list := c.listBounds()
	if x < list.X || y < list.Y || x >= list.X+list.Width || y >= list.Y+list.Height {
		return ""
	}
	for _, b := range c.blocks() {
		if y >= b.row.Y && y < b.row.Y+b.row.Height {
			return b.id
		}
	}
	return ""
}

func (c *Chat) measure(s string) float32 {
	if s == "" {
		return 0
	}
	size := c.config.FontSizeSm
	if c.config.TextRenderer != nil {
		return c.config.TextRenderer.MeasureText(s, size)
	}
	return float32(len([]rune(s))) * size * 0.6
}

func (c *Chat) lineStep() float32 {
	size := c.config.FontSizeSm
	if size <= 0 {
		size = 12
	}
	if c.config.TextRenderer != nil {
		if h := c.config.TextRenderer.LineHeight(size); h > 0 {
			return h
		}
	}
	return size
}

func (c *Chat) wrap(s string, maxW float32) []string {
	if s == "" {
		return nil
	}
	var out []string
	for _, para := range strings.Split(s, "\n") {
		r := []rune(para)
		if len(r) == 0 {
			out = append(out, "")
			continue
		}
		i := 0
		for i < len(r) {
			j := i + 1
			for j < len(r) && c.measure(string(r[i:j+1])) <= maxW {
				j++
			}
			out = append(out, string(r[i:j]))
			i = j
		}
	}
	return out
}

func (c *Chat) blocks() []chatBlock {
	list := c.listBounds()
	y := list.Y - c.scrollY
	out := make([]chatBlock, 0, len(c.messages))
	for _, m := range c.messages {
		lines, h := c.placeLines(m, uimath.NewRect(list.X, y, list.Width, 0))
		out = append(out, chatBlock{id: m.ID, row: uimath.NewRect(list.X, y, list.Width, h), lines: lines})
		y += h
	}
	return out
}

func (c *Chat) placeLines(m ChatMessage, origin uimath.Rect) ([]chatLine, float32) {
	maxW := origin.Width - 2*c.pad
	if maxW < 1 {
		maxW = origin.Width
	}
	step := c.lineStep()
	author := m.Author
	if author == "" {
		author = "me"
	}
	var lines []chatLine
	y := origin.Y + 1
	add := func(text string) {
		w := c.measure(text)
		x := origin.X + c.pad
		if m.Mine {
			x = origin.X + origin.Width - c.pad - w
			if x < origin.X+c.pad {
				x = origin.X + c.pad
			}
		}
		if text != "" {
			lines = append(lines, chatLine{text: text, x: x, y: y, w: w})
		}
		y += step
	}
	for _, line := range c.wrap(author, maxW) {
		add(line)
	}
	for _, line := range c.wrap(m.Text, maxW) {
		add(line)
	}
	h := y - origin.Y + 1
	if h < step+2 {
		h = step + 2
	}
	return lines, h
}

// Draw paints the author above the body. The list and the composer have separate clips.
func (c *Chat) Draw(buf *render.CommandBuffer) {
	b := c.Bounds()
	if b.IsEmpty() {
		return
	}
	list := c.listBounds()
	buf.PushClip(b)
	buf.DrawRect(render.RectCmd{Bounds: b, FillColor: c.config.BgColor}, 0, 1)
	buf.PushClip(list)
	for _, block := range c.blocks() {
		if block.row.Y+block.row.Height <= list.Y {
			continue
		}
		if block.row.Y >= list.Y+list.Height {
			break
		}
		for _, line := range block.lines {
			emitText(c.config, buf, line.text, line.x, line.y, c.config.FontSizeSm, line.w+1, c.config.TextColor, 1, 4)
		}
	}
	buf.PopClip()
	comp := uimath.NewRect(b.X, list.Y+list.Height, b.Width, c.composerH)
	buf.PushClip(comp)
	buf.DrawRect(render.RectCmd{Bounds: comp, FillColor: c.config.BgColor}, 1, 1)
	if c.draft != "" {
		emitText(c.config, buf, c.draft, comp.X+c.pad, comp.Y+6, c.config.FontSizeSm, comp.Width, c.config.TextColor, 1, 4)
	}
	buf.PopClip()
	buf.PopClip()
}

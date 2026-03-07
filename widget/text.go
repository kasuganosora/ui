package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Text displays a string of text.
type Text struct {
	Base
	text     string
	color    uimath.Color
	fontSize float32
}

// NewText creates a text widget.
func NewText(tree *core.Tree, text string, cfg *Config) *Text {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	t := &Text{
		Base:     NewBase(tree, core.TypeText, cfg),
		text:     text,
		color:    cfg.TextColor,
		fontSize: cfg.FontSize,
	}
	tree.SetProperty(t.id, "text", text)
	return t
}

func (t *Text) Text() string         { return t.text }
func (t *Text) Color() uimath.Color  { return t.color }
func (t *Text) FontSize() float32    { return t.fontSize }

func (t *Text) SetText(text string) {
	t.text = text
	t.tree.SetProperty(t.id, "text", text)
}

func (t *Text) SetColor(c uimath.Color) {
	t.color = c
}

func (t *Text) SetFontSize(size float32) {
	t.fontSize = size
}

// Style returns a block style with auto dimensions (sized by text content).
func (t *Text) Style() layout.Style {
	s := t.Base.Style()
	s.Display = layout.DisplayBlock
	return s
}

func (t *Text) Draw(buf *render.CommandBuffer) {
	bounds := t.Bounds()
	if bounds.IsEmpty() || t.text == "" {
		return
	}
	buf.DrawText(render.TextCmd{
		X:        bounds.X,
		Y:        bounds.Y,
		Color:    t.color,
		FontSize: t.fontSize,
	}, 0, 1)
}

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

	// Use real text rendering if available
	if t.config.TextRenderer != nil {
		tx := bounds.X
		lh := t.config.TextRenderer.LineHeight(t.fontSize)
		ty := bounds.Y + (bounds.Height-lh)/2
		t.config.TextRenderer.DrawText(buf, t.text, tx, ty, t.fontSize, bounds.Width, t.color, 1)
		return
	}

	// Fallback: colored rectangle placeholder
	textW := float32(len(t.text)) * t.fontSize * 0.55
	if textW > bounds.Width {
		textW = bounds.Width
	}
	textH := t.fontSize * 1.2
	if textH > bounds.Height {
		textH = bounds.Height
	}
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(bounds.X, bounds.Y+(bounds.Height-textH)/2, textW, textH),
		FillColor: t.color,
		Corners:   uimath.CornersAll(2),
	}, 0, 1)
}

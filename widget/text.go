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

		style := t.Base.Style()
		nowrap := style.WhiteSpace == layout.WhiteSpaceNowrap
		ellipsis := style.TextOverflow == layout.TextOverflowEllipsis

		if ellipsis && bounds.Width > 0 {
			// Check if text fits; truncate with "…" if not
			tw := t.config.TextRenderer.MeasureText(t.text, t.fontSize)
			if tw > bounds.Width {
				const dots = "…"
				dw := t.config.TextRenderer.MeasureText(dots, t.fontSize)
				avail := bounds.Width - dw
				runes := []rune(t.text)
				// Binary search for the longest prefix that fits
				lo, hi := 0, len(runes)
				for lo < hi {
					mid := (lo + hi + 1) / 2
					if t.config.TextRenderer.MeasureText(string(runes[:mid]), t.fontSize) <= avail {
						lo = mid
					} else {
						hi = mid - 1
					}
				}
				text := string(runes[:lo]) + dots
				t.config.TextRenderer.DrawText(buf, text, tx, ty, t.fontSize, 0, t.color, 1)
				return
			}
		}

		maxW := bounds.Width
		if nowrap {
			maxW = 0 // disable word-wrap
		}
		t.config.TextRenderer.DrawText(buf, t.text, tx, ty, t.fontSize, maxW, t.color, 1)
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

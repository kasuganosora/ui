package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Watermark renders repeating text across its bounds.
type Watermark struct {
	Base
	text    string
	gapX    float32
	gapY    float32
	color   uimath.Color
	opacity float32
}

func NewWatermark(tree *core.Tree, text string, cfg *Config) *Watermark {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Watermark{
		Base:    NewBase(tree, core.TypeCustom, cfg),
		text:    text,
		gapX:    120,
		gapY:    80,
		color:   uimath.RGBA(0, 0, 0, 0.06),
		opacity: 1,
	}
}

func (w *Watermark) SetText(t string)          { w.text = t }
func (w *Watermark) SetGap(x, y float32)       { w.gapX = x; w.gapY = y }
func (w *Watermark) SetColor(c uimath.Color)   { w.color = c }
func (w *Watermark) SetOpacity(o float32)       { w.opacity = o }

func (w *Watermark) Draw(buf *render.CommandBuffer) {
	bounds := w.Bounds()
	if bounds.IsEmpty() || w.text == "" {
		return
	}
	cfg := w.config
	if cfg.TextRenderer == nil {
		return
	}
	tw := cfg.TextRenderer.MeasureText(w.text, cfg.FontSizeSm)
	stepX := tw + w.gapX
	stepY := cfg.TextRenderer.LineHeight(cfg.FontSizeSm) + w.gapY

	for y := bounds.Y; y < bounds.Y+bounds.Height; y += stepY {
		for x := bounds.X; x < bounds.X+bounds.Width; x += stepX {
			cfg.TextRenderer.DrawText(buf, w.text, x, y, cfg.FontSizeSm, tw+10, w.color, w.opacity)
		}
	}

	w.DrawChildren(buf)
}

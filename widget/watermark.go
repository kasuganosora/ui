package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// WatermarkLayout controls the watermark tiling pattern.
type WatermarkLayout uint8

const (
	WatermarkRectangular WatermarkLayout = iota
	WatermarkHexagonal
)

// WatermarkText describes text content for a watermark.
type WatermarkText struct {
	FontColor  string
	FontFamily string
	FontSize   float32
	FontWeight string
	Text       string
}

// WatermarkImage describes image content for a watermark.
type WatermarkImage struct {
	IsGrayscale bool
	URL         string
}

// Watermark renders repeating text across its bounds.
type Watermark struct {
	Base
	text             string
	x                float32 // horizontal gap between watermarks
	y                float32 // vertical gap between watermarks
	color            uimath.Color
	alpha            float32
	height           float32
	width            float32
	isRepeat         bool
	layout           WatermarkLayout
	lineSpace        float32
	movable          bool
	moveInterval     int // milliseconds
	offset           [2]float32
	removable        bool
	rotate           float32
	watermarkContent interface{} // WatermarkText, WatermarkImage, or slice
	zIndex           int
}

func NewWatermark(tree *core.Tree, text string, cfg *Config) *Watermark {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Watermark{
		Base:         NewBase(tree, core.TypeCustom, cfg),
		text:         text,
		x:            120,
		y:            80,
		color:        uimath.RGBA(0, 0, 0, 0.06),
		alpha:        1,
		isRepeat:     true,
		lineSpace:    16,
		moveInterval: 3000,
		removable:    true,
		rotate:       -22,
	}
}

func (w *Watermark) SetText(t string)                    { w.text = t }
func (w *Watermark) SetX(v float32)                      { w.x = v }
func (w *Watermark) SetY(v float32)                      { w.y = v }
func (w *Watermark) SetColor(c uimath.Color)             { w.color = c }
func (w *Watermark) SetAlpha(a float32)                  { w.alpha = a }
func (w *Watermark) SetHeight(h float32)                 { w.height = h }
func (w *Watermark) SetWidth(v float32)                  { w.width = v }
func (w *Watermark) SetIsRepeat(v bool)                  { w.isRepeat = v }
func (w *Watermark) SetLayout(l WatermarkLayout)         { w.layout = l }
func (w *Watermark) SetLineSpace(v float32)              { w.lineSpace = v }
func (w *Watermark) SetMovable(v bool)                   { w.movable = v }
func (w *Watermark) SetMoveInterval(ms int)              { w.moveInterval = ms }
func (w *Watermark) SetOffset(o [2]float32)              { w.offset = o }
func (w *Watermark) SetRemovable(v bool)                 { w.removable = v }
func (w *Watermark) SetRotate(deg float32)               { w.rotate = deg }
func (w *Watermark) SetWatermarkContent(c interface{})   { w.watermarkContent = c }
func (w *Watermark) SetZIndex(z int)                     { w.zIndex = z }

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
	stepX := tw + w.x
	stepY := cfg.TextRenderer.LineHeight(cfg.FontSizeSm) + w.y

	for y := bounds.Y; y < bounds.Y+bounds.Height; y += stepY {
		for x := bounds.X; x < bounds.X+bounds.Width; x += stepX {
			cfg.TextRenderer.DrawText(buf, w.text, x, y, cfg.FontSizeSm, tw+10, w.color, w.alpha)
		}
	}

	w.DrawChildren(buf)
}

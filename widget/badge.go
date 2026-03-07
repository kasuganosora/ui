package widget

import (
	"strconv"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Badge displays a small count or dot indicator.
type Badge struct {
	Base
	count    int
	dot      bool
	maxCount int
	color    uimath.Color
	showZero bool
}

func NewBadge(tree *core.Tree, cfg *Config) *Badge {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Badge{
		Base:     NewBase(tree, core.TypeCustom, cfg),
		maxCount: 99,
		color:    uimath.ColorHex("#ff4d4f"),
	}
}

func (b *Badge) Count() int         { return b.count }
func (b *Badge) SetCount(c int)     { b.count = c }
func (b *Badge) SetDot(d bool)      { b.dot = d }
func (b *Badge) SetMaxCount(m int)  { b.maxCount = m }
func (b *Badge) SetColor(c uimath.Color) { b.color = c }
func (b *Badge) SetShowZero(v bool) { b.showZero = v }

func (b *Badge) Draw(buf *render.CommandBuffer) {
	b.DrawChildren(buf)

	bounds := b.Bounds()
	if bounds.IsEmpty() {
		return
	}

	if b.dot {
		dotSize := float32(8)
		dx := bounds.X + bounds.Width - dotSize/2
		dy := bounds.Y - dotSize/2
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(dx, dy, dotSize, dotSize),
			FillColor: b.color,
			Corners:   uimath.CornersAll(dotSize / 2),
		}, 5, 1)
		return
	}

	if b.count == 0 && !b.showZero {
		return
	}

	cfg := b.config
	text := strconv.Itoa(b.count)
	if b.count > b.maxCount {
		text = strconv.Itoa(b.maxCount) + "+"
	}

	badgeH := float32(18)
	badgeW := float32(len(text))*8 + 8
	if badgeW < badgeH {
		badgeW = badgeH
	}

	bx := bounds.X + bounds.Width - badgeW/2
	by := bounds.Y - badgeH/2

	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(bx, by, badgeW, badgeH),
		FillColor: b.color,
		Corners:   uimath.CornersAll(badgeH / 2),
	}, 5, 1)

	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(10)
		tw := cfg.TextRenderer.MeasureText(text, 10)
		cfg.TextRenderer.DrawText(buf, text, bx+(badgeW-tw)/2, by+(badgeH-lh)/2, 10, badgeW, uimath.ColorWhite, 1)
	}
}

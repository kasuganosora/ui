package widget

import (
	"strconv"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// BadgeShape controls the badge background shape.
type BadgeShape uint8

const (
	BadgeShapeCircle BadgeShape = iota // default circular badge
	BadgeShapeRound                    // pill / rounded-rect
)

// Deprecated: BadgeShapeRibbon is not part of TDesign spec.
const BadgeShapeRibbon BadgeShape = 2

// Badge displays a small count, text, or dot indicator.
type Badge struct {
	Base
	count    int
	content  string     // text content (overrides count display, e.g. "new")
	dot      bool       // dot-only mode (small colored dot, no text)
	maxCount int        // max count before showing "99+"
	color    uimath.Color
	showZero bool       // show badge when count is 0
	shape    BadgeShape // circle, round (pill), or ribbon
	size     Size       // SizeSmall uses smaller font/dimensions
	offset   [2]float32 // custom [dx,dy] offset from top-right corner
}

func NewBadge(tree *core.Tree, cfg *Config) *Badge {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Badge{
		Base:     NewBase(tree, core.TypeCustom, cfg),
		maxCount: 99,
		color:    uimath.ColorHex("#d54941"),
		shape:    BadgeShapeCircle,
		size:     SizeMedium,
	}
}

func (b *Badge) Count() int                  { return b.count }
func (b *Badge) SetCount(c int)              { b.count = c }
func (b *Badge) SetContent(c string)         { b.content = c }
func (b *Badge) SetDot(d bool)               { b.dot = d }
func (b *Badge) SetMaxCount(m int)           { b.maxCount = m }
func (b *Badge) SetColor(c uimath.Color)     { b.color = c }
func (b *Badge) SetShowZero(v bool)          { b.showZero = v }
func (b *Badge) SetShape(s BadgeShape)       { b.shape = s }
func (b *Badge) SetSize(s Size)              { b.size = s }
func (b *Badge) SetOffset(dx, dy float32)    { b.offset = [2]float32{dx, dy} }

func (b *Badge) Draw(buf *render.CommandBuffer) {
	b.DrawChildren(buf)

	bounds := b.Bounds()
	if bounds.IsEmpty() {
		return
	}

	// Dot mode: small colored dot, no text
	if b.dot {
		dotSize := float32(8)
		if b.size == SizeSmall {
			dotSize = 6
		}
		dx := bounds.X + bounds.Width - dotSize/2 + b.offset[0]
		dy := bounds.Y - dotSize/2 + b.offset[1]
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(dx, dy, dotSize, dotSize),
			FillColor: b.color,
			Corners:   uimath.CornersAll(dotSize / 2),
		}, 5, 1)
		return
	}

	// Determine display text
	text := b.content
	if text == "" {
		if b.count == 0 && !b.showZero {
			return
		}
		text = strconv.Itoa(b.count)
		if b.count > b.maxCount {
			text = strconv.Itoa(b.maxCount) + "+"
		}
	}

	cfg := b.config

	// Size-dependent dimensions
	fontSize := float32(10)
	badgeH := float32(18)
	charW := float32(8)
	if b.size == SizeSmall {
		fontSize = 8
		badgeH = 14
		charW = 6
	}

	badgeW := float32(len(text))*charW + charW
	if badgeW < badgeH {
		badgeW = badgeH
	}

	bx := bounds.X + bounds.Width - badgeW/2 + b.offset[0]
	by := bounds.Y - badgeH/2 + b.offset[1]

	// Determine corner radius based on shape
	var corners uimath.Corners
	switch b.shape {
	case BadgeShapeRound:
		corners = uimath.CornersAll(badgeH / 4) // pill / rounded-rect with less rounding
	case BadgeShapeRibbon:
		// Ribbon: only round left corners, right side is flush
		half := badgeH / 2
		corners = uimath.Corners{TopLeft: half, BottomLeft: half}
		// Shift ribbon to be fully to the right of the host
		bx = bounds.X + bounds.Width + b.offset[0]
	default: // BadgeShapeCircle
		corners = uimath.CornersAll(badgeH / 2)
	}

	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(bx, by, badgeW, badgeH),
		FillColor: b.color,
		Corners:   corners,
	}, 5, 1)

	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(fontSize)
		tw := cfg.TextRenderer.MeasureText(text, fontSize)
		cfg.TextRenderer.DrawText(buf, text, bx+(badgeW-tw)/2, by+(badgeH-lh)/2, fontSize, badgeW, uimath.ColorWhite, 1)
	}
}

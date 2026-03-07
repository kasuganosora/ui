package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

const (
	loadingSize    = float32(32)
	loadingDotSize = float32(6)
)

// Loading displays a loading indicator (three dots).
type Loading struct {
	Base
	tip string
}

// NewLoading creates a loading indicator.
func NewLoading(tree *core.Tree, cfg *Config) *Loading {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	l := &Loading{
		Base: NewBase(tree, core.TypeCustom, cfg),
	}
	l.style.Display = layout.DisplayFlex
	l.style.AlignItems = layout.AlignCenter
	l.style.JustifyContent = layout.JustifyCenter
	l.style.FlexDirection = layout.FlexDirectionColumn
	l.style.Gap = cfg.SpaceSM
	return l
}

func (l *Loading) Tip() string     { return l.tip }
func (l *Loading) SetTip(tip string) { l.tip = tip }

func (l *Loading) Draw(buf *render.CommandBuffer) {
	bounds := l.Bounds()
	if bounds.IsEmpty() {
		return
	}

	cfg := l.config
	color := cfg.PrimaryColor

	// Draw three dots in a row
	dotY := bounds.Y + (bounds.Height-loadingDotSize)/2
	if l.tip != "" {
		dotY -= cfg.FontSize // shift up to make room for tip
	}
	totalDotsW := loadingDotSize*3 + cfg.SpaceSM*2
	dotX := bounds.X + (bounds.Width-totalDotsW)/2

	for i := 0; i < 3; i++ {
		buf.DrawRect(render.RectCmd{
			Bounds:  uimath.NewRect(dotX, dotY, loadingDotSize, loadingDotSize),
			FillColor: color,
			Corners: uimath.CornersAll(loadingDotSize / 2),
		}, 0, 1)
		dotX += loadingDotSize + cfg.SpaceSM
	}

	// Tip text
	if l.tip != "" && cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		tipY := dotY + loadingDotSize + cfg.SpaceSM
		tipW := cfg.TextRenderer.MeasureText(l.tip, cfg.FontSize)
		tipX := bounds.X + (bounds.Width-tipW)/2
		cfg.TextRenderer.DrawText(buf, l.tip, tipX, tipY, cfg.FontSize, bounds.Width, cfg.TextColor, 1)
		_ = lh
	}

	l.DrawChildren(buf)
}

package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// BackTop is a floating button that scrolls to the top.
type BackTop struct {
	Base
	visible   bool
	threshold float32
	size      float32
	right     float32
	bottom    float32
	onClick   func()
}

func NewBackTop(tree *core.Tree, cfg *Config) *BackTop {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	bt := &BackTop{
		Base:      NewBase(tree, core.TypeCustom, cfg),
		threshold: 400,
		size:      40,
		right:     24,
		bottom:    24,
	}
	tree.AddHandler(bt.id, event.MouseClick, func(e *event.Event) {
		if bt.onClick != nil {
			bt.onClick()
		}
	})
	return bt
}

func (bt *BackTop) IsVisible() bool           { return bt.visible }
func (bt *BackTop) SetVisible(v bool)         { bt.visible = v }
func (bt *BackTop) SetThreshold(t float32)    { bt.threshold = t }
func (bt *BackTop) SetPosition(r, b float32)  { bt.right = r; bt.bottom = b }
func (bt *BackTop) OnClick(fn func())         { bt.onClick = fn }

func (bt *BackTop) Draw(buf *render.CommandBuffer) {
	if !bt.visible {
		return
	}
	cfg := bt.config
	s := bt.size
	r := s / 2

	// Use overlay rendering
	buf.DrawOverlay(render.RectCmd{
		Bounds:      uimath.NewRect(0, 0, s, s), // position set externally
		FillColor:   uimath.ColorWhite,
		BorderColor: cfg.BorderColor,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(r),
	}, 90, 1)

	// Up arrow
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(r-0.5, r-6, 1, 12),
		FillColor: cfg.TextColor,
	}, 91, 1)
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(r-4, r-4, 8, 1),
		FillColor: cfg.TextColor,
	}, 91, 0.6)
}

package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Popconfirm is a confirmation popup anchored to a trigger element.
type Popconfirm struct {
	Base
	title     string
	visible   bool
	placement PopupPlacement
	anchorX   float32
	anchorY   float32
	anchorW   float32
	anchorH   float32
	width     float32
	onConfirm func()
	onCancel  func()
}

func NewPopconfirm(tree *core.Tree, title string, cfg *Config) *Popconfirm {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Popconfirm{
		Base:      NewBase(tree, core.TypeCustom, cfg),
		title:     title,
		placement: PlacementTop,
		width:     220,
	}
}

func (p *Popconfirm) Title() string                  { return p.title }
func (p *Popconfirm) IsVisible() bool                { return p.visible }
func (p *Popconfirm) SetTitle(t string)              { p.title = t }
func (p *Popconfirm) SetVisible(v bool)              { p.visible = v }
func (p *Popconfirm) SetPlacement(pl PopupPlacement) { p.placement = pl }
func (p *Popconfirm) SetAnchorRect(x, y, w, h float32) {
	p.anchorX = x; p.anchorY = y; p.anchorW = w; p.anchorH = h
}
func (p *Popconfirm) OnConfirm(fn func()) { p.onConfirm = fn }
func (p *Popconfirm) OnCancel(fn func())  { p.onCancel = fn }

func (p *Popconfirm) Show()    { p.visible = true }
func (p *Popconfirm) Confirm() {
	p.visible = false
	if p.onConfirm != nil {
		p.onConfirm()
	}
}
func (p *Popconfirm) Cancel() {
	p.visible = false
	if p.onCancel != nil {
		p.onCancel()
	}
}

func (p *Popconfirm) Draw(buf *render.CommandBuffer) {
	if !p.visible {
		return
	}
	cfg := p.config
	h := float32(80)
	btnH := float32(24)

	var x, y float32
	switch p.placement {
	case PlacementBottom:
		x = p.anchorX + p.anchorW/2 - p.width/2
		y = p.anchorY + p.anchorH + 8
	default: // Top
		x = p.anchorX + p.anchorW/2 - p.width/2
		y = p.anchorY - h - 8
	}

	// Shadow
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(x+2, y+2, p.width, h),
		FillColor: uimath.RGBA(0, 0, 0, 0.12),
		Corners:   uimath.CornersAll(cfg.BorderRadius),
	}, 50, 1)

	// Background
	buf.DrawOverlay(render.RectCmd{
		Bounds:      uimath.NewRect(x, y, p.width, h),
		FillColor:   uimath.ColorWhite,
		BorderColor: cfg.BorderColor,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 51, 1)

	// Warning icon (dot)
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(x+cfg.SpaceSM, y+cfg.SpaceSM+2, 6, 6),
		FillColor: uimath.ColorHex("#faad14"),
		Corners:   uimath.CornersAll(3),
	}, 52, 1)

	// Title text
	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, p.title, x+cfg.SpaceSM+12, y+cfg.SpaceSM, cfg.FontSize, p.width-cfg.SpaceSM*2-12, cfg.TextColor, 1)
		_ = lh
	}

	// Buttons area
	btnW := float32(56)
	btnY := y + h - btnH - cfg.SpaceSM

	// Cancel button
	buf.DrawOverlay(render.RectCmd{
		Bounds:      uimath.NewRect(x+p.width-cfg.SpaceSM-btnW*2-cfg.SpaceXS, btnY, btnW, btnH),
		FillColor:   uimath.ColorWhite,
		BorderColor: cfg.BorderColor,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 52, 1)

	// OK button
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(x+p.width-cfg.SpaceSM-btnW, btnY, btnW, btnH),
		FillColor: cfg.PrimaryColor,
		Corners:   uimath.CornersAll(cfg.BorderRadius),
	}, 52, 1)
}

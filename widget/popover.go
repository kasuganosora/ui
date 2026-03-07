package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// PopoverTrigger determines how the popover is activated.
type PopoverTrigger uint8

const (
	PopoverTriggerClick PopoverTrigger = iota
	PopoverTriggerHover
)

// Popover is a floating content panel anchored to a trigger element.
type Popover struct {
	Base
	title     string
	content   Widget
	visible   bool
	placement PopupPlacement
	trigger   PopoverTrigger
	anchorX   float32
	anchorY   float32
	anchorW   float32
	anchorH   float32
	width     float32
	onClose   func()
}

func NewPopover(tree *core.Tree, cfg *Config) *Popover {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	p := &Popover{
		Base:      NewBase(tree, core.TypeCustom, cfg),
		placement: PlacementBottom,
		trigger:   PopoverTriggerClick,
		width:     200,
	}
	tree.AddHandler(p.id, event.MouseClick, func(e *event.Event) {
		if p.trigger == PopoverTriggerClick {
			p.visible = !p.visible
		}
	})
	return p
}

func (p *Popover) IsVisible() bool                  { return p.visible }
func (p *Popover) SetVisible(v bool)                { p.visible = v }
func (p *Popover) SetTitle(t string)                { p.title = t }
func (p *Popover) SetContent(w Widget)              { p.content = w }
func (p *Popover) SetPlacement(pl PopupPlacement)   { p.placement = pl }
func (p *Popover) SetTrigger(t PopoverTrigger)      { p.trigger = t }
func (p *Popover) SetWidth(w float32)               { p.width = w }
func (p *Popover) OnClose(fn func())                { p.onClose = fn }
func (p *Popover) SetAnchorRect(x, y, w, h float32) { p.anchorX = x; p.anchorY = y; p.anchorW = w; p.anchorH = h }

func (p *Popover) Open()  { p.visible = true }
func (p *Popover) Close() {
	p.visible = false
	if p.onClose != nil {
		p.onClose()
	}
}

func (p *Popover) Draw(buf *render.CommandBuffer) {
	if !p.visible {
		return
	}
	cfg := p.config
	h := float32(120)

	// Position based on placement
	var x, y float32
	switch p.placement {
	case PlacementTop:
		x = p.anchorX + p.anchorW/2 - p.width/2
		y = p.anchorY - h - 8
	case PlacementLeft:
		x = p.anchorX - p.width - 8
		y = p.anchorY + p.anchorH/2 - h/2
	case PlacementRight:
		x = p.anchorX + p.anchorW + 8
		y = p.anchorY + p.anchorH/2 - h/2
	default: // Bottom
		x = p.anchorX + p.anchorW/2 - p.width/2
		y = p.anchorY + p.anchorH + 8
	}

	// Shadow
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(x+2, y+2, p.width, h),
		FillColor: uimath.RGBA(0, 0, 0, 0.12),
		Corners:   uimath.CornersAll(cfg.BorderRadius),
	}, 40, 1)

	// Background
	buf.DrawOverlay(render.RectCmd{
		Bounds:      uimath.NewRect(x, y, p.width, h),
		FillColor:   uimath.ColorWhite,
		BorderColor: cfg.BorderColor,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 41, 1)

	// Title
	if p.title != "" && cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, p.title, x+cfg.SpaceMD, y+cfg.SpaceSM, cfg.FontSize, p.width-cfg.SpaceMD*2, cfg.TextColor, 1)
		_ = lh
	}

	// Content widget
	if p.content != nil {
		p.content.Draw(buf)
	}
}

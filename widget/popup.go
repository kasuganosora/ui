package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// PopupPlacement determines where the popup appears relative to its anchor.
type PopupPlacement uint8

const (
	PlacementTop PopupPlacement = iota
	PlacementBottom
	PlacementLeft
	PlacementRight
	PlacementTopLeft
	PlacementTopRight
	PlacementBottomLeft
	PlacementBottomRight
	PlacementLeftTop
	PlacementLeftBottom
	PlacementRightTop
	PlacementRightBottom
)

// Deprecated aliases for backward compatibility.
const (
	PlacementTopStart    = PlacementTopLeft
	PlacementTopEnd      = PlacementTopRight
	PlacementBottomStart = PlacementBottomLeft
	PlacementBottomEnd   = PlacementBottomRight
)

// PopupTrigger determines how the popup is triggered.
type PopupTrigger uint8

const (
	TriggerHover      PopupTrigger = iota
	TriggerClick
	TriggerFocus
	TriggerMousedown
	TriggerContextMenu
	TriggerManual
)

// Popup is a floating layer positioned relative to an anchor element.
type Popup struct {
	Base
	visible        bool
	placement      PopupPlacement
	anchorID       core.ElementID
	bgColor        uimath.Color
	shadow         bool
	showArrow      bool
	trigger        PopupTrigger
	disabled       bool
	hideEmptyPopup bool
	destroyOnClose bool
	zIndex         int

	onVisibleChange  func(visible bool)
	onOverlayClick   func()
	onScroll         func()
	onScrollToBottom func()
}

// NewPopup creates a popup widget.
func NewPopup(tree *core.Tree, cfg *Config) *Popup {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	p := &Popup{
		Base:      NewBase(tree, core.TypeDiv, cfg),
		placement: PlacementBottom,
		bgColor:   cfg.BgColor,
		shadow:    true,
		trigger:   TriggerHover,
	}
	p.style.Display = layout.DisplayFlex
	p.style.FlexDirection = layout.FlexDirectionColumn
	p.style.Position = layout.PositionAbsolute
	return p
}

func (p *Popup) IsVisible() bool           { return p.visible }
func (p *Popup) Placement() PopupPlacement { return p.placement }
func (p *Popup) AnchorID() core.ElementID  { return p.anchorID }
func (p *Popup) Trigger() PopupTrigger     { return p.trigger }

func (p *Popup) SetVisible(v bool) {
	p.visible = v
	p.tree.SetVisible(p.id, v)
}

func (p *Popup) SetPlacement(pl PopupPlacement) {
	p.placement = pl
}

func (p *Popup) SetAnchor(id core.ElementID) {
	p.anchorID = id
}

func (p *Popup) SetBgColor(c uimath.Color)        { p.bgColor = c }
func (p *Popup) SetShadow(s bool)                 { p.shadow = s }
func (p *Popup) SetShowArrow(v bool)               { p.showArrow = v }
func (p *Popup) SetTrigger(t PopupTrigger)         { p.trigger = t }
func (p *Popup) Disabled() bool                    { return p.disabled }
func (p *Popup) SetDisabled(v bool)                { p.disabled = v }
func (p *Popup) HideEmptyPopup() bool              { return p.hideEmptyPopup }
func (p *Popup) SetHideEmptyPopup(v bool)          { p.hideEmptyPopup = v }
func (p *Popup) DestroyOnClose() bool              { return p.destroyOnClose }
func (p *Popup) SetDestroyOnClose(v bool)          { p.destroyOnClose = v }
func (p *Popup) ZIndex() int                       { return p.zIndex }
func (p *Popup) SetZIndex(z int)                   { p.zIndex = z }
func (p *Popup) OnVisibleChange(fn func(bool))     { p.onVisibleChange = fn }
func (p *Popup) OnOverlayClick(fn func())          { p.onOverlayClick = fn }
func (p *Popup) OnScroll(fn func())                { p.onScroll = fn }
func (p *Popup) OnScrollToBottom(fn func())        { p.onScrollToBottom = fn }

// UpdatePosition calculates popup position based on anchor and placement.
func (p *Popup) UpdatePosition() {
	anchor := p.tree.Get(p.anchorID)
	if anchor == nil {
		return
	}
	anchorBounds := anchor.Layout().Bounds
	popupBounds := p.Bounds()

	var x, y float32
	switch p.placement {
	case PlacementTop:
		x = anchorBounds.X + (anchorBounds.Width-popupBounds.Width)/2
		y = anchorBounds.Y - popupBounds.Height
	case PlacementBottom:
		x = anchorBounds.X + (anchorBounds.Width-popupBounds.Width)/2
		y = anchorBounds.Y + anchorBounds.Height
	case PlacementLeft:
		x = anchorBounds.X - popupBounds.Width
		y = anchorBounds.Y + (anchorBounds.Height-popupBounds.Height)/2
	case PlacementRight:
		x = anchorBounds.X + anchorBounds.Width
		y = anchorBounds.Y + (anchorBounds.Height-popupBounds.Height)/2
	case PlacementTopStart:
		x = anchorBounds.X
		y = anchorBounds.Y - popupBounds.Height
	case PlacementTopEnd:
		x = anchorBounds.X + anchorBounds.Width - popupBounds.Width
		y = anchorBounds.Y - popupBounds.Height
	case PlacementBottomStart:
		x = anchorBounds.X
		y = anchorBounds.Y + anchorBounds.Height
	case PlacementBottomEnd:
		x = anchorBounds.X + anchorBounds.Width - popupBounds.Width
		y = anchorBounds.Y + anchorBounds.Height
	}

	p.style.Left = layout.Px(x)
	p.style.Top = layout.Px(y)
}

// drawArrow draws a small triangle pointing toward the anchor using stacked rects.
func (p *Popup) drawArrow(buf *render.CommandBuffer, bounds uimath.Rect, bgColor uimath.Color, z int32) {
	arrowSize := float32(6)
	switch p.placement {
	case PlacementTop, PlacementTopStart, PlacementTopEnd:
		// Arrow points down at bottom-center
		ax := bounds.X + bounds.Width/2 - arrowSize
		ay := bounds.Y + bounds.Height
		for i := 0; i < 3; i++ {
			fi := float32(i)
			w := arrowSize*2 - fi*4
			if w < 2 {
				w = 2
			}
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(ax+fi*2, ay+fi*2, w, 2),
				FillColor: bgColor,
			}, z, 1)
		}
	case PlacementBottom, PlacementBottomStart, PlacementBottomEnd:
		// Arrow points up at top-center
		ax := bounds.X + bounds.Width/2 - arrowSize
		ay := bounds.Y - arrowSize
		for i := 0; i < 3; i++ {
			fi := float32(i)
			w := arrowSize*2 - fi*4
			if w < 2 {
				w = 2
			}
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(ax+fi*2, ay+(2-fi)*2, w, 2),
				FillColor: bgColor,
			}, z, 1)
		}
	case PlacementLeft:
		// Arrow points right
		ax := bounds.X + bounds.Width
		ay := bounds.Y + bounds.Height/2 - arrowSize
		for i := 0; i < 3; i++ {
			fi := float32(i)
			h := arrowSize*2 - fi*4
			if h < 2 {
				h = 2
			}
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(ax+fi*2, ay+fi*2, 2, h),
				FillColor: bgColor,
			}, z, 1)
		}
	case PlacementRight:
		// Arrow points left
		ax := bounds.X - arrowSize
		ay := bounds.Y + bounds.Height/2 - arrowSize
		for i := 0; i < 3; i++ {
			fi := float32(i)
			h := arrowSize*2 - fi*4
			if h < 2 {
				h = 2
			}
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(ax+(2-fi)*2, ay+fi*2, 2, h),
				FillColor: bgColor,
			}, z, 1)
		}
	}
}

func (p *Popup) Draw(buf *render.CommandBuffer) {
	if !p.visible {
		return
	}
	bounds := p.Bounds()
	if bounds.IsEmpty() {
		return
	}

	cfg := p.config

	buf.DrawRect(render.RectCmd{
		Bounds:      bounds,
		FillColor:   p.bgColor,
		BorderColor: cfg.BorderColor,
		BorderWidth: cfg.BorderWidth,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 10, 1) // Higher z-order for popups

	// Arrow
	if p.showArrow {
		p.drawArrow(buf, bounds, p.bgColor, 10)
	}

	p.DrawChildren(buf)
}

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
	PlacementTopStart
	PlacementTopEnd
	PlacementBottomStart
	PlacementBottomEnd
)

// Popup is a floating layer positioned relative to an anchor element.
type Popup struct {
	Base
	visible   bool
	placement PopupPlacement
	anchorID  core.ElementID
	bgColor   uimath.Color
	shadow    bool
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
	}
	p.style.Display = layout.DisplayFlex
	p.style.FlexDirection = layout.FlexDirectionColumn
	p.style.Position = layout.PositionAbsolute
	return p
}

func (p *Popup) IsVisible() bool             { return p.visible }
func (p *Popup) Placement() PopupPlacement   { return p.placement }
func (p *Popup) AnchorID() core.ElementID    { return p.anchorID }

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

func (p *Popup) SetBgColor(c uimath.Color) { p.bgColor = c }
func (p *Popup) SetShadow(s bool)           { p.shadow = s }

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

	p.DrawChildren(buf)
}

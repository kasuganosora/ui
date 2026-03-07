package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Tooltip shows a text hint when hovering over an anchor element.
type Tooltip struct {
	Base
	text      string
	visible   bool
	placement PopupPlacement
	anchorID  core.ElementID
}

// NewTooltip creates a tooltip for an anchor element.
func NewTooltip(tree *core.Tree, text string, anchorID core.ElementID, cfg *Config) *Tooltip {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	t := &Tooltip{
		Base:      NewBase(tree, core.TypeDiv, cfg),
		text:      text,
		placement: PlacementTop,
		anchorID:  anchorID,
	}
	t.style.Display = layout.DisplayFlex
	t.style.Position = layout.PositionAbsolute
	t.style.Padding = layout.EdgeValues{
		Top:    layout.Px(cfg.SpaceXS),
		Bottom: layout.Px(cfg.SpaceXS),
		Left:   layout.Px(cfg.SpaceSM),
		Right:  layout.Px(cfg.SpaceSM),
	}
	t.tree.SetVisible(t.id, false)

	// Show on hover
	tree.AddHandler(anchorID, event.MouseEnter, func(e *event.Event) {
		t.Show()
	})
	tree.AddHandler(anchorID, event.MouseLeave, func(e *event.Event) {
		t.Hide()
	})

	return t
}

func (t *Tooltip) Text() string              { return t.text }
func (t *Tooltip) IsVisible() bool            { return t.visible }
func (t *Tooltip) Placement() PopupPlacement  { return t.placement }
func (t *Tooltip) AnchorID() core.ElementID   { return t.anchorID }

func (t *Tooltip) SetText(text string)          { t.text = text }
func (t *Tooltip) SetPlacement(p PopupPlacement) { t.placement = p }

func (t *Tooltip) Show() {
	t.visible = true
	t.tree.SetVisible(t.id, true)
}

func (t *Tooltip) Hide() {
	t.visible = false
	t.tree.SetVisible(t.id, false)
}

func (t *Tooltip) Draw(buf *render.CommandBuffer) {
	if !t.visible {
		return
	}
	bounds := t.Bounds()
	if bounds.IsEmpty() {
		return
	}

	// Dark background for tooltip
	bgColor := uimath.RGBA(0, 0, 0, 0.75)

	buf.DrawRect(render.RectCmd{
		Bounds:    bounds,
		FillColor: bgColor,
		Corners:   uimath.CornersAll(t.config.BorderRadius),
	}, 20, 1) // Highest z-order

	// White text
	if t.text != "" {
		if t.config.TextRenderer != nil {
			tx := bounds.X + t.config.SpaceSM
			lh := t.config.TextRenderer.LineHeight(t.config.FontSizeSm)
			ty := bounds.Y + (bounds.Height-lh)/2
			maxW := bounds.Width - t.config.SpaceSM*2
			t.config.TextRenderer.DrawText(buf, t.text, tx, ty, t.config.FontSizeSm, maxW, uimath.ColorWhite, 1)
		} else {
			textW := float32(len(t.text)) * t.config.FontSizeSm * 0.55
			textH := t.config.FontSizeSm * 1.2
			maxW := bounds.Width - t.config.SpaceSM*2
			if textW > maxW {
				textW = maxW
			}
			ty := bounds.Y + (bounds.Height-textH)/2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X+t.config.SpaceSM, ty, textW, textH),
				FillColor: uimath.ColorWhite,
				Corners:   uimath.CornersAll(2),
			}, 21, 1)
		}
	}
}

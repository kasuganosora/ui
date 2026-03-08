package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// TooltipTheme determines the color theme of the tooltip.
type TooltipTheme uint8

const (
	TooltipThemeDefault TooltipTheme = iota // dark bg + white text
	TooltipThemeLight                       // white bg + dark text + border
	TooltipThemePrimary                     // primary color bg + white text
	TooltipThemeSuccess                     // success color bg + white text
	TooltipThemeDanger                      // danger/error color bg + white text
	TooltipThemeWarning                     // warning color bg + white text
)

// Tooltip shows a text hint when hovering over an anchor element.
type Tooltip struct {
	Base
	text           string
	visible        bool
	placement      PopupPlacement
	anchorID       core.ElementID
	theme          TooltipTheme
	showArrow      bool
	delay          int // milliseconds before showing
	destroyOnClose bool
	duration       int // milliseconds before auto-hide (0 = no auto-hide)
}

// NewTooltip creates a tooltip for an anchor element.
func NewTooltip(tree *core.Tree, text string, anchorID core.ElementID, cfg *Config) *Tooltip {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	t := &Tooltip{
		Base:           NewBase(tree, core.TypeDiv, cfg),
		text:           text,
		placement:      PlacementTop,
		anchorID:       anchorID,
		theme:          TooltipThemeDefault,
		showArrow:      true,
		destroyOnClose: true,
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
func (t *Tooltip) IsVisible() bool           { return t.visible }
func (t *Tooltip) Placement() PopupPlacement { return t.placement }
func (t *Tooltip) AnchorID() core.ElementID  { return t.anchorID }
func (t *Tooltip) Theme() TooltipTheme       { return t.theme }
func (t *Tooltip) ShowArrow() bool           { return t.showArrow }

func (t *Tooltip) SetText(text string)             { t.text = text }
func (t *Tooltip) SetPlacement(p PopupPlacement)   { t.placement = p }
func (t *Tooltip) SetTheme(th TooltipTheme)        { t.theme = th }
func (t *Tooltip) SetShowArrow(v bool)             { t.showArrow = v }
func (t *Tooltip) Delay() int                      { return t.delay }
func (t *Tooltip) SetDelay(ms int)                 { t.delay = ms }
func (t *Tooltip) DestroyOnClose() bool            { return t.destroyOnClose }
func (t *Tooltip) SetDestroyOnClose(v bool)        { t.destroyOnClose = v }
func (t *Tooltip) Duration() int                   { return t.duration }
func (t *Tooltip) SetDuration(ms int)              { t.duration = ms }

func (t *Tooltip) Show() {
	t.visible = true
	t.tree.SetVisible(t.id, true)
}

func (t *Tooltip) Hide() {
	t.visible = false
	t.tree.SetVisible(t.id, false)
}

// themeColors returns (bgColor, textColor, borderColor, hasBorder) for the current theme.
func (t *Tooltip) themeColors() (uimath.Color, uimath.Color, uimath.Color, bool) {
	cfg := t.config
	switch t.theme {
	case TooltipThemeLight:
		return uimath.ColorWhite, cfg.TextColor, cfg.BorderColor, true
	case TooltipThemePrimary:
		return cfg.PrimaryColor, uimath.ColorWhite, cfg.PrimaryColor, false
	case TooltipThemeSuccess:
		return cfg.SuccessColor, uimath.ColorWhite, cfg.SuccessColor, false
	case TooltipThemeDanger:
		return cfg.ErrorColor, uimath.ColorWhite, cfg.ErrorColor, false
	case TooltipThemeWarning:
		return cfg.WarningColor, uimath.ColorWhite, cfg.WarningColor, false
	default:
		return uimath.RGBA(0, 0, 0, 0.75), uimath.ColorWhite, uimath.RGBA(0, 0, 0, 0.75), false
	}
}

// drawArrow draws a small triangle pointing toward the anchor using stacked rects.
func (t *Tooltip) drawArrow(buf *render.CommandBuffer, bounds uimath.Rect, bgColor uimath.Color, z int32) {
	arrowSize := float32(6)
	switch t.placement {
	case PlacementTop, PlacementTopStart, PlacementTopEnd:
		// Arrow points down at bottom-center of tooltip
		ax := bounds.X + bounds.Width/2 - arrowSize
		ay := bounds.Y + bounds.Height
		// 3 stacked rects getting narrower (widest on top)
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
		// Arrow points up at top-center of tooltip
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
		// Arrow points right at right-center of tooltip
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
		// Arrow points left at left-center of tooltip
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

func (t *Tooltip) Draw(buf *render.CommandBuffer) {
	if !t.visible {
		return
	}
	bounds := t.Bounds()
	if bounds.IsEmpty() {
		return
	}

	bgColor, textColor, borderColor, hasBorder := t.themeColors()

	cmd := render.RectCmd{
		Bounds:    bounds,
		FillColor: bgColor,
		Corners:   uimath.CornersAll(t.config.BorderRadius),
	}
	if hasBorder {
		cmd.BorderColor = borderColor
		cmd.BorderWidth = t.config.BorderWidth
	}
	buf.DrawRect(cmd, 20, 1) // Highest z-order

	// Arrow
	if t.showArrow {
		t.drawArrow(buf, bounds, bgColor, 20)
	}

	// Text
	if t.text != "" {
		if t.config.TextRenderer != nil {
			tx := bounds.X + t.config.SpaceSM
			lh := t.config.TextRenderer.LineHeight(t.config.FontSizeSm)
			ty := bounds.Y + (bounds.Height-lh)/2
			maxW := bounds.Width - t.config.SpaceSM*2
			t.config.TextRenderer.DrawText(buf, t.text, tx, ty, t.config.FontSizeSm, maxW, textColor, 1)
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
				FillColor: textColor,
				Corners:   uimath.CornersAll(2),
			}, 21, 1)
		}
	}
}

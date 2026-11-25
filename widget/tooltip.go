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

	// Size of the auxiliary window and whether it is currently owned by an
	// OverlayHost. When overlayOpen is false the tooltip falls back to drawing
	// inside the main widget tree (used by hosts without OS window support).
	width, height float32
	overlayOpen   bool
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
	if anchorID != core.InvalidElementID {
		tree.AddHandler(anchorID, event.Focus, func(e *event.Event) { t.Show() })
		tree.AddHandler(anchorID, event.KeyDown, func(e *event.Event) {
			if e.Key == event.KeyEscape {
				t.Hide()
			}
		})
	}

	return t
}

func (t *Tooltip) Text() string              { return t.text }
func (t *Tooltip) IsVisible() bool           { return t.visible }
func (t *Tooltip) Placement() PopupPlacement { return t.placement }
func (t *Tooltip) AnchorID() core.ElementID  { return t.anchorID }
func (t *Tooltip) Theme() TooltipTheme       { return t.theme }
func (t *Tooltip) ShowArrow() bool           { return t.showArrow }

func (t *Tooltip) SetText(text string)           { t.text = text }
func (t *Tooltip) SetPlacement(p PopupPlacement) { t.placement = p }
func (t *Tooltip) SetTheme(th TooltipTheme)      { t.theme = th }
func (t *Tooltip) SetShowArrow(v bool)           { t.showArrow = v }
func (t *Tooltip) Delay() int                    { return t.delay }
func (t *Tooltip) SetDelay(ms int)               { t.delay = ms }
func (t *Tooltip) DestroyOnClose() bool          { return t.destroyOnClose }
func (t *Tooltip) SetDestroyOnClose(v bool)      { t.destroyOnClose = v }
func (t *Tooltip) Duration() int                 { return t.duration }
func (t *Tooltip) SetDuration(ms int)            { t.duration = ms }

func (t *Tooltip) Show() {
	t.visible = true
	t.tree.SetVisible(t.id, true)

	cfg := t.config
	if cfg == nil || cfg.OverlayHost == nil {
		return // no OS window host: legacy in-tree rendering
	}
	w, h := t.measure(cfg)
	x, y := t.position(cfg, w, h)
	t.width, t.height = w, h
	if err := cfg.OverlayHost.OpenOverlay(t, t.drawOverlay, x, y, w, h); err == nil {
		t.overlayOpen = true
	} else {
		// Keep the tooltip usable if the host cannot create a window (e.g. a
		// minimal runtime). The legacy in-tree path then takes over.
		t.overlayOpen = false
	}
}

func (t *Tooltip) Hide() {
	t.visible = false
	t.tree.SetVisible(t.id, false)
	if t.overlayOpen && t.config != nil && t.config.OverlayHost != nil {
		t.config.OverlayHost.CloseOverlay(t)
		t.overlayOpen = false
	}
}

// measure returns the auxiliary window size from the label's real metrics.
func (t *Tooltip) measure(cfg *Config) (float32, float32) {
	var textW, lh float32
	if cfg.TextRenderer != nil {
		textW = cfg.TextRenderer.MeasureText(t.text, cfg.FontSizeSm)
		lh = cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
	} else {
		textW = float32(len([]rune(t.text))) * cfg.FontSizeSm * 0.55
		lh = cfg.FontSizeSm * 1.2
	}
	w := textW + cfg.SpaceSM*2
	h := lh + cfg.SpaceXS*2
	if w < 48 {
		w = 48
	}
	if h < 24 {
		h = 24
	}
	return w, h
}

// position places the tooltip relative to its anchor, in main-window client
// coordinates. The host converts these to screen coordinates.
func (t *Tooltip) position(cfg *Config, w, h float32) (float32, float32) {
	anchor := t.tree.Get(t.anchorID)
	if anchor == nil {
		return 0, 0
	}
	a := anchor.Layout().Bounds
	gap := float32(8)
	var x, y float32
	switch t.placement {
	case PlacementBottom, PlacementBottomLeft, PlacementBottomRight:
		x, y = a.X+(a.Width-w)/2, a.Y+a.Height+gap
	case PlacementLeft, PlacementLeftTop, PlacementLeftBottom:
		x, y = a.X-w-gap, a.Y+(a.Height-h)/2
	case PlacementRight, PlacementRightTop, PlacementRightBottom:
		x, y = a.X+a.Width+gap, a.Y+(a.Height-h)/2
	default: // PlacementTop and its start/end variants
		x, y = a.X+(a.Width-w)/2, a.Y-h-gap
	}
	if win := cfg.Window; win != nil {
		vw, vh := win.Size()
		x = clampFloat(x, 0, float32(vw)-w)
		y = clampFloat(y, 0, float32(vh)-h)
	}
	return x, y
}

func clampFloat(v, lo, hi float32) float32 {
	if hi < lo {
		return lo
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// drawOverlay paints into the auxiliary window's own command buffer. Bounds
// are local to that window, so the origin is always (0,0).
func (t *Tooltip) drawOverlay(cfg *Config, buf *render.CommandBuffer) {
	t.paint(cfg, buf, uimath.NewRect(0, 0, t.width, t.height))
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
	// When an auxiliary OS window is open the host paints it with its own
	// command buffer; drawing here too would duplicate the bubble in the main
	// window (and use the wrong glyph atlas).
	if !t.visible || t.overlayOpen {
		return
	}
	bounds := t.Bounds()
	if bounds.IsEmpty() {
		return
	}
	t.paint(t.config, buf, bounds)
}

// paint emits the bubble into buf using the supplied Config. cfg must belong
// to the same render backend as buf, because glyph atlas handles are local to
// a backend.
func (t *Tooltip) paint(cfg *Config, buf *render.CommandBuffer, bounds uimath.Rect) {
	bgColor, textColor, borderColor, hasBorder := t.themeColors()

	cmd := render.RectCmd{
		Bounds:    bounds,
		FillColor: bgColor,
		Corners:   uimath.CornersAll(cfg.BorderRadius),
	}
	if hasBorder {
		cmd.BorderColor = borderColor
		cmd.BorderWidth = cfg.BorderWidth
	}
	buf.DrawRect(cmd, 20, 1) // Highest z-order

	// Arrow
	if t.showArrow {
		t.drawArrow(buf, bounds, bgColor, 20)
	}

	// Text
	if t.text != "" {
		if cfg.TextRenderer != nil {
			tx := bounds.X + cfg.SpaceSM
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			ty := bounds.Y + (bounds.Height-lh)/2
			maxW := bounds.Width - cfg.SpaceSM*2
			emitText(cfg, buf, t.text, tx, ty, cfg.FontSizeSm, maxW, textColor, 1, tooltipTextZ)
		} else {
			textW := float32(len([]rune(t.text))) * cfg.FontSizeSm * 0.55
			textH := cfg.FontSizeSm * 1.2
			maxW := bounds.Width - cfg.SpaceSM*2
			if textW > maxW {
				textW = maxW
			}
			ty := bounds.Y + (bounds.Height-textH)/2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X+cfg.SpaceSM, ty, textW, textH),
				FillColor: textColor,
				Corners:   uimath.CornersAll(2),
			}, 21, 1)
		}
	}
}

package game

import (
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// Panel draws a standard game window frame: background, optional title bar,
// and returns the content area origin so widgets don't have to calculate offsets manually.
//
// Usage in a widget's Draw method:
//
//	p := game.Panel{Title: "Inventory", Width: 300, Height: 400}
//	cx, cy, cw := p.Draw(buf, bounds, cfg)
//	// draw content starting at (cx, cy) with max width cw
type Panel struct {
	Title   string
	Width   float32 // panel width (used if bounds is empty)
	Height  float32 // panel height (0 = use auto from content)
	TitleH  float32 // title bar height (0 = default 30; negative = no title bar)
	Padding float32 // content padding (0 = default 8)

	// Appearance
	BgColor     uimath.Color // zero = default dark
	BorderColor uimath.Color // zero = default gray
	BorderWidth float32      // 0 = default 1
	TitleColor  uimath.Color // zero = default gold
	Shadow      bool         // draw drop shadow
}

// PanelResult contains the layout info returned by Panel.Draw.
type PanelResult struct {
	ContentX float32 // left edge of content area
	ContentY float32 // top edge of content area (below title bar)
	ContentW float32 // available content width
	PanelX   float32 // panel left edge
	PanelY   float32 // panel top edge
	PanelW   float32 // panel width
	PanelH   float32 // panel height (as drawn)
}

// Draw draws the panel frame and returns the content area layout.
// If contentH > 0, it overrides Height for auto-sizing panels.
func (p *Panel) Draw(buf *render.CommandBuffer, bounds uimath.Rect, cfg *widget.Config, contentH float32) PanelResult {
	// Defaults
	titleH := p.TitleH
	if titleH == 0 && p.Title != "" {
		titleH = 30
	}
	if titleH < 0 {
		titleH = 0
	}
	pad := p.Padding
	if pad == 0 {
		pad = 8
	}

	// Position
	x, y := bounds.X, bounds.Y
	w := p.Width
	if bounds.Width > 0 {
		w = bounds.Width
	}
	if w == 0 {
		w = 300
	}

	// Height
	h := p.Height
	if contentH > 0 {
		h = titleH + contentH + pad
	}
	if h == 0 {
		h = bounds.Height
	}

	// Colors
	bgColor := p.BgColor
	if bgColor.A == 0 {
		bgColor = uimath.RGBA(0.06, 0.06, 0.1, 0.92)
	}
	borderColor := p.BorderColor
	if borderColor.A == 0 {
		borderColor = uimath.RGBA(0.35, 0.35, 0.45, 0.8)
	}
	borderW := p.BorderWidth
	if borderW == 0 {
		borderW = 1
	}
	titleColor := p.TitleColor
	if titleColor.A == 0 {
		titleColor = uimath.ColorHex("#ffd700")
	}

	radius := cfg.BorderRadius

	// Shadow
	if p.Shadow {
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x+3, y+3, w, h),
			FillColor: uimath.RGBA(0, 0, 0, 0.2),
			Corners:   uimath.CornersAll(radius),
		}, 1, 1)
	}

	// Background
	buf.DrawRect(render.RectCmd{
		Bounds:      uimath.NewRect(x, y, w, h),
		FillColor:   bgColor,
		BorderColor: borderColor,
		BorderWidth: borderW,
		Corners:     uimath.CornersAll(radius),
	}, 2, 1)

	// Title bar separator + text
	if p.Title != "" && titleH > 0 {
		// Separator line
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x+pad, y+titleH-1, w-pad*2, 1),
			FillColor: uimath.RGBA(0.3, 0.3, 0.4, 0.5),
		}, 3, 1)

		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			cfg.TextRenderer.DrawText(buf, p.Title, x+pad, y+(titleH-lh)/2, cfg.FontSize, w-pad*2, titleColor, 1)
		}
	}

	return PanelResult{
		ContentX: x + pad,
		ContentY: y + titleH,
		ContentW: w - pad*2,
		PanelX:   x,
		PanelY:   y,
		PanelW:   w,
		PanelH:   h,
	}
}

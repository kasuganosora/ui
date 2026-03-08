package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// PopconfirmTheme controls the popconfirm visual style.
type PopconfirmTheme uint8

const (
	PopconfirmThemeDefault PopconfirmTheme = iota
	PopconfirmThemeWarning
	PopconfirmThemeDanger
)

// Popconfirm is a confirmation popup anchored to a trigger element.
type Popconfirm struct {
	Base
	content        string
	visible        bool
	placement      PopupPlacement
	anchorX        float32
	anchorY        float32
	anchorW        float32
	anchorH        float32
	width          float32
	showArrow      bool
	confirmBtn     string
	cancelBtn      string
	destroyOnClose bool
	theme          PopconfirmTheme
	onConfirm      func()
	onCancel       func()
	onVisibleChange func(visible bool)

	// Internal button elements for hit testing
	confirmEl core.ElementID
	cancelEl  core.ElementID
}

func NewPopconfirm(tree *core.Tree, title string, cfg *Config) *Popconfirm {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	p := &Popconfirm{
		Base:        NewBase(tree, core.TypeCustom, cfg),
		content:        title,
		placement:      PlacementTop,
		width:          220,
		showArrow:      true,
		confirmBtn:     "确定",
		cancelBtn:      "取消",
		destroyOnClose: true,
	}

	// Create clickable button elements
	p.confirmEl = tree.CreateElement(core.TypeDiv)
	tree.AppendChild(p.id, p.confirmEl)

	p.cancelEl = tree.CreateElement(core.TypeDiv)
	tree.AppendChild(p.id, p.cancelEl)

	// Confirm button handler
	tree.AddHandler(p.confirmEl, event.MouseClick, func(e *event.Event) {
		p.visible = false
		if p.onConfirm != nil {
			p.onConfirm()
		}
	})

	// Cancel button handler
	tree.AddHandler(p.cancelEl, event.MouseClick, func(e *event.Event) {
		p.visible = false
		if p.onCancel != nil {
			p.onCancel()
		}
	})

	return p
}

func (p *Popconfirm) Content() string                { return p.content }
func (p *Popconfirm) IsVisible() bool                { return p.visible }
func (p *Popconfirm) SetContent(t string)            { p.content = t }
func (p *Popconfirm) SetVisible(v bool)              { p.visible = v }
func (p *Popconfirm) SetPlacement(pl PopupPlacement) { p.placement = pl }
func (p *Popconfirm) SetAnchorRect(x, y, w, h float32) {
	p.anchorX = x; p.anchorY = y; p.anchorW = w; p.anchorH = h
}
func (p *Popconfirm) OnConfirm(fn func())            { p.onConfirm = fn }
func (p *Popconfirm) OnCancel(fn func())             { p.onCancel = fn }
func (p *Popconfirm) OnVisibleChange(fn func(bool))  { p.onVisibleChange = fn }
func (p *Popconfirm) SetShowArrow(v bool)            { p.showArrow = v }
func (p *Popconfirm) SetConfirmBtn(t string)         { p.confirmBtn = t }
func (p *Popconfirm) SetCancelBtn(t string)          { p.cancelBtn = t }
func (p *Popconfirm) DestroyOnClose() bool           { return p.destroyOnClose }
func (p *Popconfirm) SetDestroyOnClose(v bool)       { p.destroyOnClose = v }
func (p *Popconfirm) Theme() PopconfirmTheme         { return p.theme }
func (p *Popconfirm) SetTheme(t PopconfirmTheme)     { p.theme = t }

// Deprecated: Use Content instead.
func (p *Popconfirm) Title() string     { return p.content }

// Deprecated: Use SetContent instead.
func (p *Popconfirm) SetTitle(t string) { p.content = t }

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

// drawArrow draws a small triangle pointing toward the anchor using stacked rects.
func (p *Popconfirm) drawArrow(buf *render.CommandBuffer, popBounds uimath.Rect, z int32) {
	arrowSize := float32(6)
	bgColor := uimath.ColorWhite
	switch p.placement {
	case PlacementBottom:
		// Arrow points up at top-center
		ax := popBounds.X + popBounds.Width/2 - arrowSize
		ay := popBounds.Y - arrowSize
		for i := 0; i < 3; i++ {
			fi := float32(i)
			w := arrowSize*2 - fi*4
			if w < 2 {
				w = 2
			}
			buf.DrawOverlay(render.RectCmd{
				Bounds:    uimath.NewRect(ax+fi*2, ay+(2-fi)*2, w, 2),
				FillColor: bgColor,
			}, z, 1)
		}
	default: // PlacementTop
		// Arrow points down at bottom-center
		ax := popBounds.X + popBounds.Width/2 - arrowSize
		ay := popBounds.Y + popBounds.Height
		for i := 0; i < 3; i++ {
			fi := float32(i)
			w := arrowSize*2 - fi*4
			if w < 2 {
				w = 2
			}
			buf.DrawOverlay(render.RectCmd{
				Bounds:    uimath.NewRect(ax+fi*2, ay+fi*2, w, 2),
				FillColor: bgColor,
			}, z, 1)
		}
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

	popBounds := uimath.NewRect(x, y, p.width, h)

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

	// Arrow
	if p.showArrow {
		p.drawArrow(buf, popBounds, 52)
	}

	// Warning icon: "!" text in orange circle
	iconSize := float32(16)
	iconX := x + cfg.SpaceSM
	iconY := y + cfg.SpaceSM
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(iconX, iconY, iconSize, iconSize),
		FillColor: cfg.WarningColor,
		Corners:   uimath.CornersAll(iconSize / 2),
	}, 52, 1)
	if cfg.TextRenderer != nil {
		// Draw "!" centered in the circle
		bangW := cfg.TextRenderer.MeasureText("!", cfg.FontSizeSm)
		lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
		cfg.TextRenderer.DrawText(buf, "!", iconX+(iconSize-bangW)/2, iconY+(iconSize-lh)/2, cfg.FontSizeSm, iconSize, uimath.ColorWhite, 1)
	} else {
		// Fallback: small white rect for "!"
		buf.DrawOverlay(render.RectCmd{
			Bounds:    uimath.NewRect(iconX+6, iconY+3, 4, 10),
			FillColor: uimath.ColorWhite,
			Corners:   uimath.CornersAll(1),
		}, 53, 1)
	}

	// Title text
	if cfg.TextRenderer != nil {
		cfg.TextRenderer.DrawText(buf, p.content, x+cfg.SpaceSM+iconSize+4, y+cfg.SpaceSM, cfg.FontSize, p.width-cfg.SpaceSM*2-iconSize-4, cfg.TextColor, 1)
	}

	// Buttons area
	btnW := float32(56)
	btnY := y + h - btnH - cfg.SpaceSM

	// Cancel button (outline style)
	cancelX := x + p.width - cfg.SpaceSM - btnW*2 - cfg.SpaceXS
	cancelBounds := uimath.NewRect(cancelX, btnY, btnW, btnH)
	buf.DrawOverlay(render.RectCmd{
		Bounds:      cancelBounds,
		FillColor:   uimath.ColorWhite,
		BorderColor: cfg.BorderColor,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 52, 1)

	// Cancel button text
	if cfg.TextRenderer != nil {
		tw := cfg.TextRenderer.MeasureText(p.cancelBtn, cfg.FontSizeSm)
		lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
		cfg.TextRenderer.DrawText(buf, p.cancelBtn, cancelX+(btnW-tw)/2, btnY+(btnH-lh)/2, cfg.FontSizeSm, btnW, cfg.TextColor, 1)
	}

	// Set layout on cancel button element for hit testing
	p.tree.SetLayout(p.cancelEl, core.LayoutResult{
		Bounds: cancelBounds,
	})

	// Confirm button (primary filled)
	confirmX := x + p.width - cfg.SpaceSM - btnW
	confirmBounds := uimath.NewRect(confirmX, btnY, btnW, btnH)
	buf.DrawOverlay(render.RectCmd{
		Bounds:    confirmBounds,
		FillColor: cfg.PrimaryColor,
		Corners:   uimath.CornersAll(cfg.BorderRadius),
	}, 52, 1)

	// Confirm button text
	if cfg.TextRenderer != nil {
		tw := cfg.TextRenderer.MeasureText(p.confirmBtn, cfg.FontSizeSm)
		lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
		cfg.TextRenderer.DrawText(buf, p.confirmBtn, confirmX+(btnW-tw)/2, btnY+(btnH-lh)/2, cfg.FontSizeSm, btnW, uimath.ColorWhite, 1)
	}

	// Set layout on confirm button element for hit testing
	p.tree.SetLayout(p.confirmEl, core.LayoutResult{
		Bounds: confirmBounds,
	})
}

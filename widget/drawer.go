package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// DrawerPlacement controls which edge the drawer slides from.
type DrawerPlacement uint8

const (
	DrawerRight  DrawerPlacement = iota
	DrawerLeft
	DrawerTop
	DrawerBottom
)

// DrawerMode controls how the drawer appears.
type DrawerMode uint8

const (
	DrawerModeOverlay DrawerMode = iota // overlay on top of content (default)
	DrawerModePush                      // push content aside
)

// Drawer is a sliding panel from the edge of the screen.
type Drawer struct {
	Base
	header              string
	visible             bool
	size                string
	placement           DrawerPlacement
	closeBtn            bool
	onClose             func()
	onCloseBtnClick     func()
	onBeforeClose       func()
	onBeforeOpen        func()
	onEscKeydown        func()
	onOverlayClick      func()
	confirmBtn          string
	cancelBtn           string
	onConfirm           func()
	onCancel            func()
	footer              bool
	closeOnOverlayClick bool
	closeOnEscKeydown   bool
	showOverlay         bool
	mode                DrawerMode
	body                Widget

	closeID   core.ElementID
	confirmID core.ElementID
	cancelID  core.ElementID
}

func NewDrawer(tree *core.Tree, title string, cfg *Config) *Drawer {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	d := &Drawer{
		Base:                NewBase(tree, core.TypeCustom, cfg),
		header:              title,
		size:                "378",
		placement:           DrawerRight,
		closeBtn:            true,
		confirmBtn:          "\u786e\u5b9a",
		cancelBtn:           "\u53d6\u6d88",
		footer:              true,
		closeOnOverlayClick: true,
		closeOnEscKeydown:   true,
		showOverlay:         true,
	}

	// Backdrop click to close
	tree.AddHandler(d.id, event.MouseClick, func(e *event.Event) {
		if d.closeOnOverlayClick {
			if d.onOverlayClick != nil {
				d.onOverlayClick()
			}
			if d.onClose != nil {
				d.onClose()
			}
		}
	})

	// Close button element
	d.closeID = tree.CreateElement(core.TypeCustom)
	tree.AppendChild(d.id, d.closeID)
	tree.AddHandler(d.closeID, event.MouseClick, func(e *event.Event) {
		if d.onCloseBtnClick != nil {
			d.onCloseBtnClick()
		}
		if d.onClose != nil {
			d.onClose()
		}
	})

	// Confirm button element
	d.confirmID = tree.CreateElement(core.TypeCustom)
	tree.AppendChild(d.id, d.confirmID)
	tree.AddHandler(d.confirmID, event.MouseClick, func(e *event.Event) {
		if d.onConfirm != nil {
			d.onConfirm()
		}
	})

	// Cancel button element
	d.cancelID = tree.CreateElement(core.TypeCustom)
	tree.AppendChild(d.id, d.cancelID)
	tree.AddHandler(d.cancelID, event.MouseClick, func(e *event.Event) {
		if d.onCancel != nil {
			d.onCancel()
		}
	})

	return d
}

func (d *Drawer) SetHeader(t string)                { d.header = t }
func (d *Drawer) SetSize(s string)                  { d.size = s }
func (d *Drawer) SetPlacement(p DrawerPlacement)    { d.placement = p }
func (d *Drawer) SetCloseBtn(c bool)                { d.closeBtn = c }
func (d *Drawer) OnClose(fn func())                 { d.onClose = fn }
func (d *Drawer) OnCloseBtnClick(fn func())         { d.onCloseBtnClick = fn }
func (d *Drawer) OnBeforeClose(fn func())           { d.onBeforeClose = fn }
func (d *Drawer) OnBeforeOpen(fn func())            { d.onBeforeOpen = fn }
func (d *Drawer) OnEscKeydown(fn func())            { d.onEscKeydown = fn }
func (d *Drawer) OnOverlayClick(fn func())          { d.onOverlayClick = fn }
func (d *Drawer) IsVisible() bool                   { return d.visible }
func (d *Drawer) SetConfirmBtn(t string)            { d.confirmBtn = t }
func (d *Drawer) SetCancelBtn(t string)             { d.cancelBtn = t }
func (d *Drawer) OnConfirm(fn func())               { d.onConfirm = fn }
func (d *Drawer) OnCancel(fn func())                { d.onCancel = fn }
func (d *Drawer) SetFooter(show bool)               { d.footer = show }
func (d *Drawer) SetCloseOnOverlayClick(close bool) { d.closeOnOverlayClick = close }
func (d *Drawer) SetCloseOnEscKeydown(close bool)   { d.closeOnEscKeydown = close }
func (d *Drawer) SetShowOverlay(show bool)          { d.showOverlay = show }
func (d *Drawer) SetMode(m DrawerMode)              { d.mode = m }
func (d *Drawer) SetBody(w Widget)                  { d.body = w }

func (d *Drawer) Open() {
	if d.onBeforeOpen != nil {
		d.onBeforeOpen()
	}
	d.visible = true
	d.tree.MarkDirty(d.id)
}

func (d *Drawer) Close() {
	if d.onBeforeClose != nil {
		d.onBeforeClose()
	}
	d.visible = false
	d.tree.MarkDirty(d.id)
}

// drawerSize parses the size string to a float32 pixel value.
func (d *Drawer) drawerSize() float32 {
	size := float32(378)
	if d.size == "small" {
		size = 300
	} else if d.size == "medium" {
		size = 378
	} else if d.size == "large" {
		size = 520
	} else {
		var v float32
		n := 0
		for _, c := range d.size {
			if c >= '0' && c <= '9' {
				v = v*10 + float32(c-'0')
				n++
			} else {
				break
			}
		}
		if n > 0 {
			size = v
		}
	}
	return size
}

func (d *Drawer) Draw(buf *render.CommandBuffer) {
	if !d.visible {
		return
	}
	bounds := d.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := d.config

	// Backdrop
	if d.showOverlay {
		buf.DrawOverlay(render.RectCmd{
			Bounds:    bounds,
			FillColor: uimath.RGBA(0, 0, 0, 0.45),
		}, 15, 1)
	}

	// Panel
	sz := d.drawerSize()
	var px, py, pw, ph float32
	switch d.placement {
	case DrawerRight:
		pw, ph = sz, bounds.Height
		px, py = bounds.X+bounds.Width-pw, bounds.Y
	case DrawerLeft:
		pw, ph = sz, bounds.Height
		px, py = bounds.X, bounds.Y
	case DrawerTop:
		pw, ph = bounds.Width, sz
		px, py = bounds.X, bounds.Y
	case DrawerBottom:
		pw, ph = bounds.Width, sz
		px, py = bounds.X, bounds.Y+bounds.Height-ph
	}

	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(px, py, pw, ph),
		FillColor: uimath.ColorWhite,
	}, 16, 1)

	// Header
	headerH := float32(48)
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(px, py+headerH-1, pw, 1),
		FillColor: uimath.RGBA(0, 0, 0, 0.06),
	}, 17, 1)

	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSizeLg)
		cfg.TextRenderer.DrawText(buf, d.header, px+cfg.SpaceLG, py+(headerH-lh)/2, cfg.FontSizeLg, pw-cfg.SpaceLG*2, cfg.TextColor, 1)
	}

	// Close button (X) in title bar top-right
	if d.closeBtn {
		closeSize := float32(16)
		closeX := px + pw - cfg.SpaceLG - closeSize
		closeY := py + (headerH-closeSize)/2

		buf.DrawOverlay(render.RectCmd{
			Bounds:    uimath.NewRect(closeX+closeSize/2-1, closeY, 2, closeSize),
			FillColor: cfg.TextColor,
		}, 18, 1)
		buf.DrawOverlay(render.RectCmd{
			Bounds:    uimath.NewRect(closeX, closeY+closeSize/2-1, closeSize, 2),
			FillColor: cfg.TextColor,
		}, 18, 1)

		closeHit := closeSize + 8
		d.tree.SetLayout(d.closeID, core.LayoutResult{
			Bounds: uimath.NewRect(closeX-4, closeY-4, closeHit, closeHit),
		})
	}

	// Footer area with buttons
	if d.footer {
		footerH := float32(56)
		footerY := py + ph - footerH

		// Footer top border
		buf.DrawOverlay(render.RectCmd{
			Bounds:    uimath.NewRect(px, footerY, pw, 1),
			FillColor: cfg.BorderColor,
		}, 17, 1)

		// Buttons area (right-aligned)
		btnH := float32(32)
		btnY := footerY + (footerH-btnH)/2
		btnPadX := float32(8)
		btnGap := cfg.SpaceSM

		// Measure button text widths
		confirmTextW := float32(len(d.confirmBtn)) * cfg.FontSize * 0.55
		cancelTextW := float32(len(d.cancelBtn)) * cfg.FontSize * 0.55
		if cfg.TextRenderer != nil {
			confirmTextW = cfg.TextRenderer.MeasureText(d.confirmBtn, cfg.FontSize)
			cancelTextW = cfg.TextRenderer.MeasureText(d.cancelBtn, cfg.FontSize)
		}
		confirmBtnW := confirmTextW + btnPadX*2
		cancelBtnW := cancelTextW + btnPadX*2
		if confirmBtnW < 64 {
			confirmBtnW = 64
		}
		if cancelBtnW < 64 {
			cancelBtnW = 64
		}

		// Position from right
		confirmBtnX := px + pw - cfg.SpaceLG - confirmBtnW
		cancelBtnX := confirmBtnX - btnGap - cancelBtnW

		// Cancel button (outline style)
		buf.DrawOverlay(render.RectCmd{
			Bounds:      uimath.NewRect(cancelBtnX, btnY, cancelBtnW, btnH),
			FillColor:   uimath.ColorWhite,
			BorderColor: cfg.BorderColor,
			BorderWidth: 1,
			Corners:     uimath.CornersAll(cfg.BorderRadius),
		}, 18, 1)
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			tw := cfg.TextRenderer.MeasureText(d.cancelBtn, cfg.FontSize)
			cfg.TextRenderer.DrawText(buf, d.cancelBtn,
				cancelBtnX+(cancelBtnW-tw)/2, btnY+(btnH-lh)/2,
				cfg.FontSize, cancelBtnW, cfg.TextColor, 1)
		} else {
			th := cfg.FontSize * 1.2
			buf.DrawOverlay(render.RectCmd{
				Bounds:    uimath.NewRect(cancelBtnX+(cancelBtnW-cancelTextW)/2, btnY+(btnH-th)/2, cancelTextW, th),
				FillColor: cfg.TextColor,
				Corners:   uimath.CornersAll(2),
			}, 19, 1)
		}

		// Confirm button (primary filled)
		buf.DrawOverlay(render.RectCmd{
			Bounds:    uimath.NewRect(confirmBtnX, btnY, confirmBtnW, btnH),
			FillColor: cfg.PrimaryColor,
			Corners:   uimath.CornersAll(cfg.BorderRadius),
		}, 18, 1)
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			tw := cfg.TextRenderer.MeasureText(d.confirmBtn, cfg.FontSize)
			cfg.TextRenderer.DrawText(buf, d.confirmBtn,
				confirmBtnX+(confirmBtnW-tw)/2, btnY+(btnH-lh)/2,
				cfg.FontSize, confirmBtnW, uimath.ColorWhite, 1)
		} else {
			th := cfg.FontSize * 1.2
			buf.DrawOverlay(render.RectCmd{
				Bounds:    uimath.NewRect(confirmBtnX+(confirmBtnW-confirmTextW)/2, btnY+(btnH-th)/2, confirmTextW, th),
				FillColor: uimath.ColorWhite,
				Corners:   uimath.CornersAll(2),
			}, 19, 1)
		}

		// Set layout on button elements for hit testing
		d.tree.SetLayout(d.cancelID, core.LayoutResult{
			Bounds: uimath.NewRect(cancelBtnX, btnY, cancelBtnW, btnH),
		})
		d.tree.SetLayout(d.confirmID, core.LayoutResult{
			Bounds: uimath.NewRect(confirmBtnX, btnY, confirmBtnW, btnH),
		})
	}

	d.DrawChildren(buf)
}

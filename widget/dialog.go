package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// DialogTheme controls the dialog style.
type DialogTheme uint8

const (
	DialogThemeDefault DialogTheme = iota
	DialogThemeInfo
	DialogThemeWarning
	DialogThemeDanger
	DialogThemeSuccess
)

// DialogPlacement controls the dialog vertical position.
type DialogPlacement uint8

const (
	DialogPlacementTop    DialogPlacement = iota // top:20% (default)
	DialogPlacementCenter                        // vertically centered
)

// DialogMode controls the dialog modality.
type DialogMode uint8

const (
	DialogModeModal    DialogMode = iota // modal (default)
	DialogModeModeless                   // non-modal
	DialogModeNormal                     // in-flow
)

// Dialog is a modal overlay dialog.
type Dialog struct {
	Base
	header  string
	visible bool
	width   float32
	body    Widget

	onClose            func()
	onCloseBtnClick    func()
	onClosed           func()
	onOpened           func()
	onBeforeClose      func()
	onBeforeOpen       func()
	onEscKeydown       func()
	onOverlayClick     func()
	confirmBtn         string
	cancelBtn          string
	onConfirm          func()
	onCancel           func()
	footer             bool
	closeOnOverlayClick bool
	closeOnEscKeydown  bool
	showOverlay        bool
	theme              DialogTheme
	placement          DialogPlacement
	mode               DialogMode

	closeID   core.ElementID // close X button element
	confirmID core.ElementID // confirm button element
	cancelID  core.ElementID // cancel button element
}

// NewDialog creates a dialog.
func NewDialog(tree *core.Tree, title string, cfg *Config) *Dialog {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	d := &Dialog{
		Base:                NewBase(tree, core.TypeCustom, cfg),
		header:              title,
		width:               520,
		confirmBtn:          "\u786e\u5b9a",
		cancelBtn:           "\u53d6\u6d88",
		footer:              true,
		closeOnOverlayClick: true,
		closeOnEscKeydown:   true,
		showOverlay:         true,
	}
	d.style.Display = layout.DisplayNone

	// Click on backdrop closes dialog
	d.tree.AddHandler(d.id, event.MouseClick, func(e *event.Event) {
		if d.closeOnOverlayClick {
			if d.onClose != nil {
				d.onClose()
			}
		}
	})

	// Close button element
	d.closeID = tree.CreateElement(core.TypeCustom)
	tree.AppendChild(d.id, d.closeID)
	tree.AddHandler(d.closeID, event.MouseClick, func(e *event.Event) {
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

func (d *Dialog) Header() string   { return d.header }
func (d *Dialog) IsVisible() bool  { return d.visible }

func (d *Dialog) SetHeader(header string)              { d.header = header }
func (d *Dialog) SetWidth(w float32)                   { d.width = w }
func (d *Dialog) SetBody(w Widget)                     { d.body = w }
func (d *Dialog) OnClose(fn func())                    { d.onClose = fn }
func (d *Dialog) OnCloseBtnClick(fn func())            { d.onCloseBtnClick = fn }
func (d *Dialog) OnClosed(fn func())                   { d.onClosed = fn }
func (d *Dialog) OnOpened(fn func())                   { d.onOpened = fn }
func (d *Dialog) OnBeforeClose(fn func())              { d.onBeforeClose = fn }
func (d *Dialog) OnBeforeOpen(fn func())               { d.onBeforeOpen = fn }
func (d *Dialog) OnEscKeydown(fn func())               { d.onEscKeydown = fn }
func (d *Dialog) OnOverlayClick(fn func())             { d.onOverlayClick = fn }
func (d *Dialog) SetConfirmBtn(t string)               { d.confirmBtn = t }
func (d *Dialog) SetCancelBtn(t string)                { d.cancelBtn = t }
func (d *Dialog) OnConfirm(fn func())                  { d.onConfirm = fn }
func (d *Dialog) OnCancel(fn func())                   { d.onCancel = fn }
func (d *Dialog) SetFooter(show bool)                  { d.footer = show }
func (d *Dialog) SetCloseOnOverlayClick(close bool)    { d.closeOnOverlayClick = close }
func (d *Dialog) SetCloseOnEscKeydown(close bool)      { d.closeOnEscKeydown = close }
func (d *Dialog) SetShowOverlay(show bool)             { d.showOverlay = show }
func (d *Dialog) SetTheme(t DialogTheme)               { d.theme = t }
func (d *Dialog) SetPlacement(p DialogPlacement)       { d.placement = p }
func (d *Dialog) SetMode(m DialogMode)                 { d.mode = m }

func (d *Dialog) Open() {
	d.visible = true
	d.style.Display = layout.DisplayFlex
}

func (d *Dialog) Close() {
	d.visible = false
	d.style.Display = layout.DisplayNone
}

func (d *Dialog) Draw(buf *render.CommandBuffer) {
	if !d.visible {
		return
	}
	bounds := d.Bounds()
	if bounds.IsEmpty() {
		return
	}

	cfg := d.config

	// Backdrop (semi-transparent overlay)
	buf.DrawRect(render.RectCmd{
		Bounds:    bounds,
		FillColor: uimath.Color{R: 0, G: 0, B: 0, A: 0.45},
	}, 10, 1)

	// Dialog panel centered
	panelW := d.width
	panelH := float32(200)
	if d.body != nil {
		panelH = 300
	}
	footerH := float32(0)
	if d.footer {
		footerH = 56
		panelH += footerH
	}
	panelX := bounds.X + (bounds.Width-panelW)/2
	panelY := bounds.Y + (bounds.Height-panelH)/2

	// Panel background
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(panelX, panelY, panelW, panelH),
		FillColor: uimath.ColorWhite,
		Corners:   uimath.CornersAll(cfg.BorderRadius),
	}, 11, 1)

	// Title bar
	titleH := float32(56)
	if d.header != "" {
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeLg)
			tx := panelX + cfg.SpaceLG
			ty := panelY + (titleH-lh)/2
			cfg.TextRenderer.DrawText(buf, d.header, tx, ty, cfg.FontSizeLg, panelW-cfg.SpaceLG*2, cfg.TextColor, 1)
		} else {
			textW := float32(len(d.header)) * cfg.FontSizeLg * 0.55
			textH := cfg.FontSizeLg * 1.2
			tx := panelX + cfg.SpaceLG
			ty := panelY + (titleH-textH)/2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(tx, ty, textW, textH),
				FillColor: cfg.TextColor,
				Corners:   uimath.CornersAll(2),
			}, 12, 1)
		}
	}

	// Close button (X) in title bar top-right
	closeSize := float32(16)
	closeX := panelX + panelW - cfg.SpaceLG - closeSize
	closeY := panelY + (titleH-closeSize)/2
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(closeX+closeSize/2-1, closeY, 2, closeSize),
		FillColor: cfg.TextColor,
	}, 12, 1)
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(closeX, closeY+closeSize/2-1, closeSize, 2),
		FillColor: cfg.TextColor,
	}, 12, 1)

	// Set layout on close element for hit testing
	closeHit := closeSize + 8
	d.tree.SetLayout(d.closeID, core.LayoutResult{
		Bounds: uimath.NewRect(closeX-4, closeY-4, closeHit, closeHit),
	})

	// Title divider
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(panelX, panelY+titleH, panelW, 1),
		FillColor: cfg.BorderColor,
	}, 12, 1)

	// Content area
	if d.body != nil {
		d.body.Draw(buf)
	}

	// Footer area with buttons
	if d.footer {
		footerY := panelY + panelH - footerH

		// Footer top border
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(panelX, footerY, panelW, 1),
			FillColor: cfg.BorderColor,
		}, 12, 1)

		// Buttons area (right-aligned)
		btnH := float32(32)
		btnY := footerY + (footerH-btnH)/2
		btnPadX := float32(8) // horizontal padding inside buttons
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
		confirmBtnX := panelX + panelW - cfg.SpaceLG - confirmBtnW
		cancelBtnX := confirmBtnX - btnGap - cancelBtnW

		// Cancel button (outline style)
		buf.DrawRect(render.RectCmd{
			Bounds:      uimath.NewRect(cancelBtnX, btnY, cancelBtnW, btnH),
			FillColor:   uimath.ColorWhite,
			BorderColor: cfg.BorderColor,
			BorderWidth: 1,
			Corners:     uimath.CornersAll(cfg.BorderRadius),
		}, 12, 1)
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			tw := cfg.TextRenderer.MeasureText(d.cancelBtn, cfg.FontSize)
			cfg.TextRenderer.DrawText(buf, d.cancelBtn,
				cancelBtnX+(cancelBtnW-tw)/2, btnY+(btnH-lh)/2,
				cfg.FontSize, cancelBtnW, cfg.TextColor, 1)
		} else {
			th := cfg.FontSize * 1.2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(cancelBtnX+(cancelBtnW-cancelTextW)/2, btnY+(btnH-th)/2, cancelTextW, th),
				FillColor: cfg.TextColor,
				Corners:   uimath.CornersAll(2),
			}, 13, 1)
		}

		// Confirm button (primary filled)
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(confirmBtnX, btnY, confirmBtnW, btnH),
			FillColor: cfg.PrimaryColor,
			Corners:   uimath.CornersAll(cfg.BorderRadius),
		}, 12, 1)
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			tw := cfg.TextRenderer.MeasureText(d.confirmBtn, cfg.FontSize)
			cfg.TextRenderer.DrawText(buf, d.confirmBtn,
				confirmBtnX+(confirmBtnW-tw)/2, btnY+(btnH-lh)/2,
				cfg.FontSize, confirmBtnW, uimath.ColorWhite, 1)
		} else {
			th := cfg.FontSize * 1.2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(confirmBtnX+(confirmBtnW-confirmTextW)/2, btnY+(btnH-th)/2, confirmTextW, th),
				FillColor: uimath.ColorWhite,
				Corners:   uimath.CornersAll(2),
			}, 13, 1)
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

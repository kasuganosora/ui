package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// AlertTheme controls the alert style (maps to TDesign theme).
type AlertTheme uint8

const (
	AlertThemeInfo    AlertTheme = iota
	AlertThemeSuccess
	AlertThemeWarning
	AlertThemeError
)

// Alert displays a colored banner with an icon and message.
type Alert struct {
	Base
	message   string
	title     string
	theme     AlertTheme
	closeBtn  bool
	visible   bool
	onClose   func()
	onClosed  func()
	operation string
	maxLine   int

	closeID core.ElementID // clickable close button element
}

func NewAlert(tree *core.Tree, message string, cfg *Config) *Alert {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	a := &Alert{
		Base:    NewBase(tree, core.TypeCustom, cfg),
		message: message,
		visible: true,
	}

	// Create close button element
	a.closeID = tree.CreateElement(core.TypeCustom)
	tree.AppendChild(a.id, a.closeID)
	tree.AddHandler(a.closeID, event.MouseClick, func(e *event.Event) {
		if a.closeBtn {
			a.Close()
		}
	})

	return a
}

func (a *Alert) SetMessage(m string)       { a.message = m }
func (a *Alert) SetTheme(t AlertTheme)     { a.theme = t }
func (a *Alert) SetCloseBtn(c bool)        { a.closeBtn = c }
func (a *Alert) OnClose(fn func())         { a.onClose = fn }
func (a *Alert) OnClosed(fn func())        { a.onClosed = fn }
func (a *Alert) IsVisible() bool           { return a.visible }
func (a *Alert) SetTitle(t string)         { a.title = t }
func (a *Alert) SetOperation(op string)    { a.operation = op }
func (a *Alert) SetMaxLine(n int)          { a.maxLine = n }

func (a *Alert) Close() {
	a.visible = false
	if a.onClose != nil {
		a.onClose()
	}
	if a.onClosed != nil {
		a.onClosed()
	}
}

func (a *Alert) alertColors() (bg, border, text uimath.Color) {
	switch a.theme {
	case AlertThemeSuccess:
		return uimath.ColorHex("#f6ffed"), uimath.ColorHex("#b7eb8f"), uimath.ColorHex("#52c41a")
	case AlertThemeWarning:
		return uimath.ColorHex("#fffbe6"), uimath.ColorHex("#ffe58f"), uimath.ColorHex("#faad14")
	case AlertThemeError:
		return uimath.ColorHex("#fff2f0"), uimath.ColorHex("#ffccc7"), uimath.ColorHex("#ff4d4f")
	default:
		return uimath.ColorHex("#e6f4ff"), uimath.ColorHex("#91caff"), uimath.ColorHex("#1677ff")
	}
}

// alertIconText returns the icon character and color for the alert type.
func (a *Alert) alertIconText() (string, uimath.Color) {
	switch a.theme {
	case AlertThemeSuccess:
		return "\u2713", uimath.ColorHex("#52c41a") // ✓
	case AlertThemeWarning:
		return "!", uimath.ColorHex("#faad14")
	case AlertThemeError:
		return "\u00d7", uimath.ColorHex("#ff4d4f") // ×
	default: // Info
		return "i", uimath.ColorHex("#1677ff")
	}
}

func (a *Alert) Draw(buf *render.CommandBuffer) {
	if !a.visible {
		return
	}
	bounds := a.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := a.config
	bg, border, textClr := a.alertColors()

	buf.DrawRect(render.RectCmd{
		Bounds:      bounds,
		FillColor:   bg,
		BorderColor: border,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 0, 1)

	// Status icon (circle with symbol)
	dotSize := float32(16)
	dotX := bounds.X + cfg.SpaceMD
	hasTitle := a.title != ""
	dotY := bounds.Y + cfg.SpaceMD
	if !hasTitle {
		dotY = bounds.Y + (bounds.Height-dotSize)/2
	}

	iconText, iconColor := a.alertIconText()

	// Draw circle background
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(dotX, dotY, dotSize, dotSize),
		FillColor: iconColor,
		Corners:   uimath.CornersAll(dotSize / 2),
	}, 1, 1)

	// Draw icon character inside circle
	if cfg.TextRenderer != nil {
		iconFontSize := float32(10)
		lh := cfg.TextRenderer.LineHeight(iconFontSize)
		iw := cfg.TextRenderer.MeasureText(iconText, iconFontSize)
		cfg.TextRenderer.DrawText(buf, iconText,
			dotX+(dotSize-iw)/2, dotY+(dotSize-lh)/2,
			iconFontSize, dotSize, uimath.ColorWhite, 1)
	}

	// Text area
	textX := dotX + dotSize + cfg.SpaceSM
	closeSpace := float32(0)
	if a.closeBtn {
		closeSpace = 24 // space for close button
	}
	maxW := bounds.Width - cfg.SpaceMD*2 - dotSize - cfg.SpaceSM - closeSpace

	if hasTitle {
		// Title (bold/larger text above message)
		titleY := bounds.Y + cfg.SpaceMD
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			cfg.TextRenderer.DrawText(buf, a.title, textX, titleY, cfg.FontSize, maxW, cfg.TextColor, 1)
			// Message below title
			msgY := titleY + lh + 4
			cfg.TextRenderer.DrawText(buf, a.message, textX, msgY, cfg.FontSizeSm, maxW, cfg.TextColor, 1)
			// Operation text after message
			if a.operation != "" {
				msgW := cfg.TextRenderer.MeasureText(a.message, cfg.FontSizeSm)
				opX := textX + msgW + cfg.SpaceSM
				cfg.TextRenderer.DrawText(buf, a.operation, opX, msgY, cfg.FontSizeSm, maxW-msgW-cfg.SpaceSM, textClr, 1)
			}
		} else {
			// Title placeholder
			titleW := float32(len(a.title)) * cfg.FontSize * 0.55
			titleH := cfg.FontSize * 1.2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(textX, titleY, titleW, titleH),
				FillColor: cfg.TextColor,
				Corners:   uimath.CornersAll(2),
			}, 1, 1)
			// Message placeholder
			msgY := titleY + titleH + 4
			tw := float32(len(a.message)) * cfg.FontSizeSm * 0.55
			th := cfg.FontSizeSm * 1.2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(textX, msgY, tw, th),
				FillColor: cfg.TextColor,
				Corners:   uimath.CornersAll(2),
			}, 1, 1)
		}
	} else {
		// Single line: message only
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			msgY := bounds.Y + (bounds.Height-lh)/2
			cfg.TextRenderer.DrawText(buf, a.message, textX, msgY, cfg.FontSize, maxW, cfg.TextColor, 1)
			// Operation text after message
			if a.operation != "" {
				msgW := cfg.TextRenderer.MeasureText(a.message, cfg.FontSize)
				opX := textX + msgW + cfg.SpaceSM
				cfg.TextRenderer.DrawText(buf, a.operation, opX, msgY, cfg.FontSize, maxW-msgW-cfg.SpaceSM, textClr, 1)
			}
		} else {
			tw := float32(len(a.message)) * cfg.FontSize * 0.55
			th := cfg.FontSize * 1.2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(textX, bounds.Y+(bounds.Height-th)/2, tw, th),
				FillColor: cfg.TextColor,
				Corners:   uimath.CornersAll(2),
			}, 1, 1)
		}
	}

	// Close button (X) in top-right
	if a.closeBtn {
		closeSize := float32(14)
		closeX := bounds.X + bounds.Width - cfg.SpaceMD - closeSize
		closeY := bounds.Y + cfg.SpaceMD
		if !hasTitle {
			closeY = bounds.Y + (bounds.Height-closeSize)/2
		}

		// Draw X as cross
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(closeX+closeSize/2-1, closeY, 2, closeSize),
			FillColor: cfg.TextColor,
		}, 1, 1)
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(closeX, closeY+closeSize/2-1, closeSize, 2),
			FillColor: cfg.TextColor,
		}, 1, 1)

		// Set layout for hit testing
		hitSize := closeSize + 4
		a.tree.SetLayout(a.closeID, core.LayoutResult{
			Bounds: uimath.NewRect(closeX-2, closeY-2, hitSize, hitSize),
		})
	}
}

package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// AlertType controls the alert style.
type AlertType uint8

const (
	AlertInfo    AlertType = iota
	AlertSuccess
	AlertWarning
	AlertError
)

// Alert displays a colored banner with an icon and message.
type Alert struct {
	Base
	message   string
	alertType AlertType
	closable  bool
	visible   bool
	onClose   func()
}

func NewAlert(tree *core.Tree, message string, cfg *Config) *Alert {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Alert{
		Base:    NewBase(tree, core.TypeCustom, cfg),
		message: message,
		visible: true,
	}
}

func (a *Alert) SetMessage(m string)     { a.message = m }
func (a *Alert) SetAlertType(t AlertType) { a.alertType = t }
func (a *Alert) SetClosable(c bool)      { a.closable = c }
func (a *Alert) OnClose(fn func())       { a.onClose = fn }
func (a *Alert) IsVisible() bool         { return a.visible }

func (a *Alert) Close() {
	a.visible = false
	if a.onClose != nil {
		a.onClose()
	}
}

func (a *Alert) alertColors() (bg, border, text uimath.Color) {
	switch a.alertType {
	case AlertSuccess:
		return uimath.ColorHex("#f6ffed"), uimath.ColorHex("#b7eb8f"), uimath.ColorHex("#52c41a")
	case AlertWarning:
		return uimath.ColorHex("#fffbe6"), uimath.ColorHex("#ffe58f"), uimath.ColorHex("#faad14")
	case AlertError:
		return uimath.ColorHex("#fff2f0"), uimath.ColorHex("#ffccc7"), uimath.ColorHex("#ff4d4f")
	default:
		return uimath.ColorHex("#e6f4ff"), uimath.ColorHex("#91caff"), uimath.ColorHex("#1677ff")
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

	// Icon dot
	dotSize := float32(14)
	dotX := bounds.X + cfg.SpaceMD
	dotY := bounds.Y + (bounds.Height-dotSize)/2
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(dotX, dotY, dotSize, dotSize),
		FillColor: textClr,
		Corners:   uimath.CornersAll(dotSize / 2),
	}, 1, 1)

	// Message text
	textX := dotX + dotSize + cfg.SpaceSM
	maxW := bounds.Width - cfg.SpaceMD*2 - dotSize - cfg.SpaceSM
	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, a.message, textX, bounds.Y+(bounds.Height-lh)/2, cfg.FontSize, maxW, cfg.TextColor, 1)
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

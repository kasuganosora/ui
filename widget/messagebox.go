package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// MessageBoxType controls the icon and style.
type MessageBoxType uint8

const (
	MessageBoxInfo    MessageBoxType = iota
	MessageBoxSuccess
	MessageBoxWarning
	MessageBoxError
	MessageBoxConfirm
)

// MessageBox is a modal message dialog with OK/Cancel buttons.
type MessageBox struct {
	Base
	title     string
	content   string
	boxType   MessageBoxType
	visible   bool
	width     float32
	onOK      func()
	onCancel  func()
	showCancel bool
}

// NewMessageBox creates a message box dialog.
func NewMessageBox(tree *core.Tree, title, content string, cfg *Config) *MessageBox {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	mb := &MessageBox{
		Base:       NewBase(tree, core.TypeCustom, cfg),
		title:      title,
		content:    content,
		boxType:    MessageBoxInfo,
		width:      420,
		showCancel: false,
	}
	mb.style.Display = layout.DisplayNone

	// Click backdrop to close
	tree.AddHandler(mb.id, event.MouseClick, func(e *event.Event) {
		if mb.onCancel != nil {
			mb.onCancel()
		}
	})

	return mb
}

func (mb *MessageBox) Title() string          { return mb.title }
func (mb *MessageBox) Content() string        { return mb.content }
func (mb *MessageBox) BoxType() MessageBoxType { return mb.boxType }
func (mb *MessageBox) IsVisible() bool        { return mb.visible }

func (mb *MessageBox) SetTitle(title string)        { mb.title = title }
func (mb *MessageBox) SetContent(content string)    { mb.content = content }
func (mb *MessageBox) SetBoxType(t MessageBoxType)   { mb.boxType = t }
func (mb *MessageBox) SetWidth(w float32)            { mb.width = w }
func (mb *MessageBox) SetShowCancel(v bool)          { mb.showCancel = v }
func (mb *MessageBox) OnOK(fn func())                { mb.onOK = fn }
func (mb *MessageBox) OnCancel(fn func())            { mb.onCancel = fn }

func (mb *MessageBox) Open() {
	mb.visible = true
	mb.style.Display = layout.DisplayFlex
	mb.tree.MarkDirty(mb.id)
}

func (mb *MessageBox) Close() {
	mb.visible = false
	mb.style.Display = layout.DisplayNone
	mb.tree.MarkDirty(mb.id)
}

func (mb *MessageBox) typeColor() uimath.Color {
	switch mb.boxType {
	case MessageBoxSuccess:
		return uimath.ColorHex("#52c41a")
	case MessageBoxWarning:
		return uimath.ColorHex("#faad14")
	case MessageBoxError:
		return uimath.ColorHex("#ff4d4f")
	case MessageBoxConfirm:
		return uimath.ColorHex("#faad14")
	default:
		return mb.config.PrimaryColor
	}
}

func (mb *MessageBox) Draw(buf *render.CommandBuffer) {
	if !mb.visible {
		return
	}
	bounds := mb.Bounds()
	if bounds.IsEmpty() {
		return
	}

	cfg := mb.config

	// Backdrop
	buf.DrawRect(render.RectCmd{
		Bounds:    bounds,
		FillColor: uimath.RGBA(0, 0, 0, 0.45),
	}, 10, 1)

	// Panel
	panelW := mb.width
	panelH := float32(180)
	if mb.content != "" {
		panelH = 220
	}
	panelX := bounds.X + (bounds.Width-panelW)/2
	panelY := bounds.Y + (bounds.Height-panelH)/2

	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(panelX, panelY, panelW, panelH),
		FillColor: uimath.ColorWhite,
		Corners:   uimath.CornersAll(cfg.BorderRadius),
	}, 11, 1)

	// Icon dot
	iconSize := float32(22)
	iconX := panelX + cfg.SpaceLG
	iconY := panelY + cfg.SpaceLG
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(iconX, iconY, iconSize, iconSize),
		FillColor: mb.typeColor(),
		Corners:   uimath.CornersAll(iconSize / 2),
	}, 12, 1)

	// Title
	titleX := iconX + iconSize + cfg.SpaceSM
	if mb.title != "" {
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeLg)
			ty := panelY + cfg.SpaceLG + (iconSize-lh)/2
			cfg.TextRenderer.DrawText(buf, mb.title, titleX, ty, cfg.FontSizeLg, panelW-cfg.SpaceLG*2-iconSize-cfg.SpaceSM, cfg.TextColor, 1)
		} else {
			tw := float32(len(mb.title)) * cfg.FontSizeLg * 0.55
			th := cfg.FontSizeLg * 1.2
			ty := panelY + cfg.SpaceLG + (iconSize-th)/2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(titleX, ty, tw, th),
				FillColor: cfg.TextColor,
				Corners:   uimath.CornersAll(2),
			}, 12, 1)
		}
	}

	// Content
	if mb.content != "" {
		contentY := panelY + cfg.SpaceLG + iconSize + cfg.SpaceMD
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			cfg.TextRenderer.DrawText(buf, mb.content, titleX, contentY, cfg.FontSize, panelW-cfg.SpaceLG*2-iconSize-cfg.SpaceSM, uimath.RGBA(0, 0, 0, 0.65), 1)
			_ = lh
		} else {
			tw := float32(len(mb.content)) * cfg.FontSize * 0.55
			th := cfg.FontSize * 1.2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(titleX, contentY, tw, th),
				FillColor: uimath.RGBA(0, 0, 0, 0.65),
				Corners:   uimath.CornersAll(2),
			}, 12, 1)
		}
	}

	// Button area
	btnH := float32(32)
	btnY := panelY + panelH - cfg.SpaceMD - btnH
	btnGap := cfg.SpaceSM

	// OK button
	okW := float32(60)
	okX := panelX + panelW - cfg.SpaceLG - okW
	if mb.showCancel {
		cancelW := float32(60)
		okX -= cancelW + btnGap
		// Cancel button
		cancelX := panelX + panelW - cfg.SpaceLG - cancelW
		buf.DrawRect(render.RectCmd{
			Bounds:      uimath.NewRect(cancelX, btnY, cancelW, btnH),
			FillColor:   uimath.ColorWhite,
			BorderColor: cfg.BorderColor,
			BorderWidth: 1,
			Corners:     uimath.CornersAll(cfg.BorderRadius),
		}, 12, 1)
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			tw := cfg.TextRenderer.MeasureText("取消", cfg.FontSize)
			cfg.TextRenderer.DrawText(buf, "取消", cancelX+(cancelW-tw)/2, btnY+(btnH-lh)/2, cfg.FontSize, cancelW, cfg.TextColor, 1)
		}
	}

	// OK button
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(okX, btnY, okW, btnH),
		FillColor: cfg.PrimaryColor,
		Corners:   uimath.CornersAll(cfg.BorderRadius),
	}, 12, 1)
	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		tw := cfg.TextRenderer.MeasureText("确定", cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, "确定", okX+(okW-tw)/2, btnY+(btnH-lh)/2, cfg.FontSize, okW, uimath.ColorWhite, 1)
	}

	mb.DrawChildren(buf)
}

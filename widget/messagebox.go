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
	MessageBoxInfo MessageBoxType = iota
	MessageBoxSuccess
	MessageBoxWarning
	MessageBoxError
	MessageBoxConfirm
)

// MessageBox is a modal message dialog with OK/Cancel buttons.
type MessageBox struct {
	Base
	title      string
	content    string
	boxType    MessageBoxType
	visible    bool
	width      float32
	onOK       func()
	onCancel   func()
	showCancel bool
	prior      core.ElementID
	scoped     bool
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
	tree.SetVisible(mb.id, false)

	// Click backdrop to close
	tree.AddHandler(mb.id, event.MouseClick, func(e *event.Event) {
		if mb.onCancel != nil {
			mb.onCancel()
		}
	})
	tree.AddHandler(mb.id, event.KeyDown, func(e *event.Event) {
		if e.Phase == event.PhaseCapture || e.Key != event.KeyEscape || !mb.visible {
			return
		}
		if mb.onCancel != nil {
			mb.onCancel()
		}
		mb.Close()
		e.StopPropagation()
	})

	return mb
}

func (mb *MessageBox) Title() string           { return mb.title }
func (mb *MessageBox) Content() string         { return mb.content }
func (mb *MessageBox) BoxType() MessageBoxType { return mb.boxType }
func (mb *MessageBox) IsVisible() bool         { return mb.visible }

func (mb *MessageBox) SetTitle(title string)       { mb.title = title }
func (mb *MessageBox) SetContent(content string)   { mb.content = content }
func (mb *MessageBox) SetBoxType(t MessageBoxType) { mb.boxType = t }
func (mb *MessageBox) SetWidth(w float32)          { mb.width = w }
func (mb *MessageBox) SetShowCancel(v bool)        { mb.showCancel = v }
func (mb *MessageBox) OnOK(fn func())              { mb.onOK = fn }
func (mb *MessageBox) OnCancel(fn func())          { mb.onCancel = fn }

func (mb *MessageBox) Open() {
	if !mb.visible {
		mb.prior = mb.tree.Focused()
		mb.tree.PushFocusScope(mb.id)
		mb.tree.SetFocusable(mb.id, true)
		mb.scoped = true
	}
	mb.visible = true
	mb.style.Display = layout.DisplayFlex
	mb.tree.SetVisible(mb.id, true)
	mb.tree.SetFocused(mb.id, true)
	mb.tree.PushLayer(mb.id)
	if elem := mb.tree.Get(mb.id); elem != nil && elem.ParentID() != core.InvalidElementID {
		mb.tree.BringChildToFront(elem.ParentID(), mb.id)
	}
	mb.tree.MarkDirty(mb.id)
}

func (mb *MessageBox) Close() {
	was := mb.visible
	focusWasInside := mb.tree.InScope(mb.tree.Focused(), mb.id)
	mb.visible = false
	mb.style.Display = layout.DisplayNone
	mb.tree.SetVisible(mb.id, false)
	mb.tree.SetFocusable(mb.id, false)
	mb.tree.PopLayer(mb.id)
	if mb.scoped {
		mb.tree.PopFocusScopeOwned(mb.id)
		mb.scoped = false
	}
	if was && focusWasInside {
		scope := mb.tree.FocusScope()
		if mb.prior != core.InvalidElementID && mb.tree.Get(mb.prior) != nil && mb.tree.InScope(mb.prior, scope) && mb.tree.ChainAllows(mb.prior) {
			mb.tree.SetFocused(mb.prior, true)
		} else if scope != core.InvalidElementID {
			mb.tree.FocusNext(core.InvalidElementID)
		}
	}
	mb.tree.MarkDirty(mb.id)
}

func (mb *MessageBox) typeColor() uimath.Color {
	switch mb.boxType {
	case MessageBoxSuccess:
		return uimath.ColorHex("#2ba471")
	case MessageBoxWarning:
		return uimath.ColorHex("#e37318")
	case MessageBoxError:
		return uimath.ColorHex("#d54941")
	case MessageBoxConfirm:
		return uimath.ColorHex("#e37318")
	default:
		return mb.config.PrimaryColor
	}
}

func (mb *MessageBox) Draw(buf *render.CommandBuffer) {
	if !mb.visible {
		return
	}
	defer withLayer(buf, mb.tree, mb.id)()
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
			emitText(cfg, buf, mb.title, titleX, ty, cfg.FontSizeLg, panelW-cfg.SpaceLG*2-iconSize-cfg.SpaceSM, cfg.TextColor, 1, dialogTextZ)
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
			emitText(cfg, buf, mb.content, titleX, contentY, cfg.FontSize, panelW-cfg.SpaceLG*2-iconSize-cfg.SpaceSM, uimath.RGBA(0, 0, 0, 0.65), 1, dialogTextZ)
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
			emitText(cfg, buf, "取消", cancelX+(cancelW-tw)/2, btnY+(btnH-lh)/2, cfg.FontSize, cancelW, cfg.TextColor, 1, dialogTextZ)
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
		emitText(cfg, buf, "确定", okX+(okW-tw)/2, btnY+(btnH-lh)/2, cfg.FontSize, okW, uimath.ColorWhite, 1, dialogTextZ)
	}

	mb.DrawChildren(buf)
}

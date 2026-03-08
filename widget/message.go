package widget

import (
	"time"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// MessageTheme controls the message appearance (maps to TDesign theme).
type MessageTheme uint8

const (
	MessageThemeInfo    MessageTheme = iota
	MessageThemeSuccess
	MessageThemeWarning
	MessageThemeError
	MessageThemeQuestion
	MessageThemeLoading
)

// Message is a global notification message (toast).
type Message struct {
	Base
	content         string
	theme           MessageTheme
	visible         bool
	duration        int // milliseconds, default 3000; 0 = no auto-dismiss
	closeBtn        bool
	onClose         func()
	onCloseBtnClick func()
	onDurationEnd   func()
	startTime       time.Time
	closeID         core.ElementID
}

// NewMessage creates a message notification.
func NewMessage(tree *core.Tree, content string, cfg *Config) *Message {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	m := &Message{
		Base:     NewBase(tree, core.TypeCustom, cfg),
		content:  content,
		visible:  true,
		duration: 3000,
	}
	m.style.Display = layout.DisplayFlex
	m.style.AlignItems = layout.AlignCenter
	m.style.JustifyContent = layout.JustifyCenter
	m.style.Height = layout.Px(36)
	m.style.Padding = layout.EdgeValues{
		Left:  layout.Px(cfg.SpaceMD),
		Right: layout.Px(cfg.SpaceMD),
	}
	m.startTime = time.Now()

	// Close button element
	m.closeID = tree.CreateElement(core.TypeCustom)
	tree.AppendChild(m.id, m.closeID)
	tree.AddHandler(m.closeID, event.MouseClick, func(e *event.Event) {
		m.SetVisible(false)
		if m.onCloseBtnClick != nil {
			m.onCloseBtnClick()
		}
		if m.onClose != nil {
			m.onClose()
		}
	})

	return m
}

func (m *Message) Content() string       { return m.content }
func (m *Message) Theme() MessageTheme   { return m.theme }
func (m *Message) IsVisible() bool       { return m.visible }

func (m *Message) SetContent(content string) { m.content = content }
func (m *Message) SetTheme(t MessageTheme)   { m.theme = t }

func (m *Message) SetVisible(v bool) {
	m.visible = v
	if v {
		m.startTime = time.Now()
	}
}

func (m *Message) SetDuration(ms int)        { m.duration = ms }
func (m *Message) SetCloseBtn(c bool)        { m.closeBtn = c }
func (m *Message) OnClose(fn func())         { m.onClose = fn }
func (m *Message) OnCloseBtnClick(fn func()) { m.onCloseBtnClick = fn }
func (m *Message) OnDurationEnd(fn func())   { m.onDurationEnd = fn }

func (m *Message) iconColor() uimath.Color {
	switch m.theme {
	case MessageThemeSuccess:
		return m.config.SuccessColor
	case MessageThemeWarning:
		return m.config.WarningColor
	case MessageThemeError:
		return m.config.ErrorColor
	default:
		return m.config.PrimaryColor
	}
}

func (m *Message) iconText() string {
	switch m.theme {
	case MessageThemeSuccess:
		return "\u2713" // ✓
	case MessageThemeWarning:
		return "!"
	case MessageThemeError:
		return "\u00d7" // ×
	default:
		return "i"
	}
}

func (m *Message) Draw(buf *render.CommandBuffer) {
	if !m.visible {
		return
	}

	// Auto-dismiss after duration
	if m.duration > 0 && time.Since(m.startTime) > time.Duration(m.duration)*time.Millisecond {
		m.SetVisible(false)
		if m.onDurationEnd != nil {
			m.onDurationEnd()
		}
		if m.onClose != nil {
			m.onClose()
		}
		return
	}
	// Keep redrawing for timer
	if m.duration > 0 {
		m.tree.MarkDirty(m.id)
	}

	bounds := m.Bounds()
	if bounds.IsEmpty() {
		return
	}

	cfg := m.config

	// Message pill background with shadow effect
	buf.DrawRect(render.RectCmd{
		Bounds:      bounds,
		FillColor:   uimath.ColorWhite,
		BorderColor: uimath.ColorHex("#e8e8e8"),
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 20, 1)

	// Status icon
	dotSize := float32(16)
	dotX := bounds.X + cfg.SpaceMD
	dotY := bounds.Y + (bounds.Height-dotSize)/2
	iconClr := m.iconColor()

	// Draw icon circle background
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(dotX, dotY, dotSize, dotSize),
		FillColor: iconClr,
		Corners:   uimath.CornersAll(dotSize / 2),
	}, 21, 1)

	// Draw icon character
	if cfg.TextRenderer != nil {
		iconStr := m.iconText()
		iconFontSize := float32(10)
		tw := cfg.TextRenderer.MeasureText(iconStr, iconFontSize)
		lh := cfg.TextRenderer.LineHeight(iconFontSize)
		cfg.TextRenderer.DrawText(buf, iconStr,
			dotX+(dotSize-tw)/2, dotY+(dotSize-lh)/2,
			iconFontSize, dotSize, uimath.ColorWhite, 1)
	}

	// Text
	rightPad := cfg.SpaceMD
	if m.closeBtn {
		rightPad += 20 // space for close button
	}
	if m.content != "" {
		textX := dotX + dotSize + cfg.SpaceSM
		textW := bounds.Width - cfg.SpaceMD - dotSize - cfg.SpaceSM - rightPad
		if textW > 0 {
			if cfg.TextRenderer != nil {
				lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
				textY := bounds.Y + (bounds.Height-lh)/2
				cfg.TextRenderer.DrawText(buf, m.content, textX, textY, cfg.FontSize, textW, cfg.TextColor, 1)
			} else {
				tw := float32(len(m.content)) * cfg.FontSize * 0.55
				if tw > textW {
					tw = textW
				}
				th := cfg.FontSize * 1.2
				textY := bounds.Y + (bounds.Height-th)/2
				buf.DrawRect(render.RectCmd{
					Bounds:    uimath.NewRect(textX, textY, tw, th),
					FillColor: cfg.TextColor,
					Corners:   uimath.CornersAll(2),
				}, 21, 1)
			}
		}
	}

	// Close button (X)
	if m.closeBtn {
		closeSize := float32(12)
		closeX := bounds.X + bounds.Width - cfg.SpaceMD - closeSize
		closeY := bounds.Y + (bounds.Height-closeSize)/2

		// Draw X as two crossing lines
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(closeX+closeSize/2-1, closeY, 2, closeSize),
			FillColor: cfg.DisabledColor,
		}, 22, 1)
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(closeX, closeY+closeSize/2-1, closeSize, 2),
			FillColor: cfg.DisabledColor,
		}, 22, 1)

		// Set layout on close element for hit testing
		closeHit := closeSize + 8
		m.tree.SetLayout(m.closeID, core.LayoutResult{
			Bounds: uimath.NewRect(closeX-4, closeY-4, closeHit, closeHit),
		})
	}

	m.DrawChildren(buf)
}

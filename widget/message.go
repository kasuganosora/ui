package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// MessageType controls the message appearance.
type MessageType uint8

const (
	MessageInfo    MessageType = iota
	MessageSuccess
	MessageWarning
	MessageError
)

// Message is a global notification message (toast).
type Message struct {
	Base
	content string
	msgType MessageType
	visible bool
}

// NewMessage creates a message notification.
func NewMessage(tree *core.Tree, content string, cfg *Config) *Message {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	m := &Message{
		Base:    NewBase(tree, core.TypeCustom, cfg),
		content: content,
		visible: true,
	}
	m.style.Display = layout.DisplayFlex
	m.style.AlignItems = layout.AlignCenter
	m.style.JustifyContent = layout.JustifyCenter
	m.style.Height = layout.Px(36)
	m.style.Padding = layout.EdgeValues{
		Left:  layout.Px(cfg.SpaceMD),
		Right: layout.Px(cfg.SpaceMD),
	}
	return m
}

func (m *Message) Content() string     { return m.content }
func (m *Message) MsgType() MessageType { return m.msgType }
func (m *Message) IsVisible() bool     { return m.visible }

func (m *Message) SetContent(content string) { m.content = content }
func (m *Message) SetMsgType(t MessageType)  { m.msgType = t }
func (m *Message) SetVisible(v bool)         { m.visible = v }

func (m *Message) iconColor() uimath.Color {
	switch m.msgType {
	case MessageSuccess:
		return uimath.ColorHex("#52c41a")
	case MessageWarning:
		return uimath.ColorHex("#faad14")
	case MessageError:
		return uimath.ColorHex("#ff4d4f")
	default:
		return m.config.PrimaryColor
	}
}

func (m *Message) Draw(buf *render.CommandBuffer) {
	if !m.visible {
		return
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

	// Status dot
	dotSize := float32(8)
	dotX := bounds.X + cfg.SpaceMD
	dotY := bounds.Y + (bounds.Height-dotSize)/2
	buf.DrawRect(render.RectCmd{
		Bounds:  uimath.NewRect(dotX, dotY, dotSize, dotSize),
		FillColor: m.iconColor(),
		Corners: uimath.CornersAll(dotSize / 2),
	}, 21, 1)

	// Text
	if m.content != "" {
		textX := dotX + dotSize + cfg.SpaceSM
		textW := bounds.Width - cfg.SpaceMD*2 - dotSize - cfg.SpaceSM
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

	m.DrawChildren(buf)
}

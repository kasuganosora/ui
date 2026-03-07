package game

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// ChatMessage represents a single chat message.
type ChatMessage struct {
	Sender  string
	Text    string
	Color   uimath.Color // Sender name color
	Channel string       // e.g., "world", "party", "whisper"
}

// ChatBox displays a scrollable list of chat messages with an input area.
type ChatBox struct {
	widget.Base
	messages   []ChatMessage
	maxVisible int
	scrollY    int
	width      float32
	height     float32
	inputText  string
	onSend     func(text string)
}

// NewChatBox creates a chat box.
func NewChatBox(tree *core.Tree, cfg *widget.Config) *ChatBox {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	return &ChatBox{
		Base:       widget.NewBase(tree, core.TypeCustom, cfg),
		maxVisible: 10,
		width:      350,
		height:     250,
	}
}

func (cb *ChatBox) Messages() []ChatMessage   { return cb.messages }
func (cb *ChatBox) InputText() string          { return cb.inputText }
func (cb *ChatBox) SetSize(w, h float32)       { cb.width = w; cb.height = h }
func (cb *ChatBox) SetMaxVisible(n int)        { cb.maxVisible = n }
func (cb *ChatBox) OnSend(fn func(string))     { cb.onSend = fn }

func (cb *ChatBox) AddMessage(msg ChatMessage) {
	cb.messages = append(cb.messages, msg)
	// Auto-scroll to bottom
	if len(cb.messages) > cb.maxVisible {
		cb.scrollY = len(cb.messages) - cb.maxVisible
	}
}

func (cb *ChatBox) ClearMessages() {
	cb.messages = cb.messages[:0]
	cb.scrollY = 0
}

func (cb *ChatBox) ScrollUp() {
	if cb.scrollY > 0 {
		cb.scrollY--
	}
}

func (cb *ChatBox) ScrollDown() {
	max := len(cb.messages) - cb.maxVisible
	if max < 0 {
		max = 0
	}
	if cb.scrollY < max {
		cb.scrollY++
	}
}

func (cb *ChatBox) Draw(buf *render.CommandBuffer) {
	bounds := cb.Bounds()
	if bounds.IsEmpty() {
		bounds = uimath.NewRect(0, 0, cb.width, cb.height)
	}

	cfg := cb.Config()
	inputH := float32(28)
	msgAreaH := bounds.Height - inputH - 4

	// Background
	buf.DrawRect(render.RectCmd{
		Bounds:    bounds,
		FillColor: uimath.RGBA(0, 0, 0, 0.6),
		Corners:   uimath.CornersAll(cfg.BorderRadius),
	}, 1, 1)

	// Messages area
	if cfg.TextRenderer != nil {
		lineH := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
		y := bounds.Y + 4
		start := cb.scrollY
		end := start + cb.maxVisible
		if end > len(cb.messages) {
			end = len(cb.messages)
		}

		for i := start; i < end; i++ {
			msg := cb.messages[i]
			if y+lineH > bounds.Y+msgAreaH {
				break
			}
			// Sender name
			senderText := "[" + msg.Sender + "] "
			senderW := cfg.TextRenderer.MeasureText(senderText, cfg.FontSizeSm)
			senderColor := msg.Color
			if senderColor == (uimath.Color{}) {
				senderColor = uimath.ColorHex("#aaaaaa")
			}
			cfg.TextRenderer.DrawText(buf, senderText, bounds.X+4, y, cfg.FontSizeSm, senderW, senderColor, 1)
			// Message text
			textW := bounds.Width - 8 - senderW
			cfg.TextRenderer.DrawText(buf, msg.Text, bounds.X+4+senderW, y, cfg.FontSizeSm, textW, uimath.ColorWhite, 0.9)
			y += lineH + 2
		}
	}

	// Input area
	inputY := bounds.Y + bounds.Height - inputH
	buf.DrawRect(render.RectCmd{
		Bounds:      uimath.NewRect(bounds.X, inputY, bounds.Width, inputH),
		FillColor:   uimath.RGBA(0.15, 0.15, 0.15, 0.9),
		BorderColor: uimath.RGBA(0.3, 0.3, 0.3, 1),
		BorderWidth: 1,
		Corners: uimath.Corners{
			BottomLeft:  cfg.BorderRadius,
			BottomRight: cfg.BorderRadius,
		},
	}, 2, 1)

	if cb.inputText != "" && cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
		cfg.TextRenderer.DrawText(buf, cb.inputText, bounds.X+4, inputY+(inputH-lh)/2, cfg.FontSizeSm, bounds.Width-8, uimath.ColorWhite, 1)
	}
}

// FloatingText displays temporary floating text (damage numbers, XP gains, etc.).
type FloatingText struct {
	widget.Base
	text  string
	color uimath.Color
	x, y  float32
}

// NewFloatingText creates a floating text element.
func NewFloatingText(tree *core.Tree, text string, x, y float32, color uimath.Color, cfg *widget.Config) *FloatingText {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	return &FloatingText{
		Base:  widget.NewBase(tree, core.TypeCustom, cfg),
		text:  text,
		color: color,
		x:     x,
		y:     y,
	}
}

func (ft *FloatingText) SetPosition(x, y float32) { ft.x = x; ft.y = y }
func (ft *FloatingText) SetText(t string)          { ft.text = t }
func (ft *FloatingText) SetColor(c uimath.Color)   { ft.color = c }

func (ft *FloatingText) Draw(buf *render.CommandBuffer) {
	cfg := ft.Config()
	if cfg.TextRenderer != nil {
		cfg.TextRenderer.DrawText(buf, ft.text, ft.x, ft.y, cfg.FontSizeLg, 200, ft.color, 1)
	} else {
		tw := float32(len(ft.text)) * cfg.FontSizeLg * 0.55
		th := cfg.FontSizeLg * 1.2
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(ft.x, ft.y, tw, th),
			FillColor: ft.color,
			Corners:   uimath.CornersAll(2),
		}, 10, 1)
	}
}

// ItemTooltip displays detailed item information on hover.
type ItemTooltip struct {
	widget.Base
	item    *ItemData
	visible bool
	x, y    float32
	width   float32
}

// NewItemTooltip creates an item tooltip.
func NewItemTooltip(tree *core.Tree, cfg *widget.Config) *ItemTooltip {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	return &ItemTooltip{
		Base:  widget.NewBase(tree, core.TypeCustom, cfg),
		width: 220,
	}
}

func (it *ItemTooltip) SetItem(item *ItemData)    { it.item = item }
func (it *ItemTooltip) SetPosition(x, y float32)  { it.x = x; it.y = y }
func (it *ItemTooltip) SetVisible(v bool)          { it.visible = v }
func (it *ItemTooltip) IsVisible() bool            { return it.visible }

func (it *ItemTooltip) Draw(buf *render.CommandBuffer) {
	if !it.visible || it.item == nil {
		return
	}

	cfg := it.Config()
	lineH := cfg.FontSize * 1.5
	h := lineH*3 + 8 // name + rarity + placeholder description

	// Background
	buf.DrawOverlay(render.RectCmd{
		Bounds:      uimath.NewRect(it.x, it.y, it.width, h),
		FillColor:   uimath.RGBA(0.05, 0.05, 0.05, 0.95),
		BorderColor: rarityColor(it.item.Rarity),
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 100, 1)

	// Item name
	if cfg.TextRenderer != nil {
		cfg.TextRenderer.DrawText(buf, it.item.Name, it.x+8, it.y+4, cfg.FontSize, it.width-16, rarityColor(it.item.Rarity), 1)
	}
}

// NotificationToast displays a temporary notification message.
type NotificationToast struct {
	widget.Base
	text     string
	icon     string
	toastType MessageType
	visible  bool
	x, y     float32
	width    float32
}

// MessageType for toasts.
type MessageType uint8

const (
	ToastInfo    MessageType = iota
	ToastSuccess
	ToastWarning
	ToastError
)

// NewNotificationToast creates a toast notification.
func NewNotificationToast(tree *core.Tree, text string, cfg *widget.Config) *NotificationToast {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	return &NotificationToast{
		Base:      widget.NewBase(tree, core.TypeCustom, cfg),
		text:      text,
		toastType: ToastInfo,
		visible:   true,
		width:     300,
	}
}

func (nt *NotificationToast) SetText(t string)        { nt.text = t }
func (nt *NotificationToast) SetToastType(t MessageType) { nt.toastType = t }
func (nt *NotificationToast) SetVisible(v bool)        { nt.visible = v }
func (nt *NotificationToast) SetPosition(x, y float32) { nt.x = x; nt.y = y }
func (nt *NotificationToast) IsVisible() bool          { return nt.visible }

func (nt *NotificationToast) toastColor() uimath.Color {
	switch nt.toastType {
	case ToastSuccess:
		return uimath.ColorHex("#52c41a")
	case ToastWarning:
		return uimath.ColorHex("#faad14")
	case ToastError:
		return uimath.ColorHex("#ff4d4f")
	default:
		return uimath.ColorHex("#1677ff")
	}
}

func (nt *NotificationToast) Draw(buf *render.CommandBuffer) {
	if !nt.visible {
		return
	}

	cfg := nt.Config()
	h := float32(40)

	// Background
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(nt.x, nt.y, nt.width, h),
		FillColor: uimath.RGBA(0.1, 0.1, 0.1, 0.9),
		Corners:   uimath.CornersAll(cfg.BorderRadius),
	}, 90, 1)

	// Color bar
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(nt.x, nt.y, 4, h),
		FillColor: nt.toastColor(),
		Corners: uimath.Corners{
			TopLeft:    cfg.BorderRadius,
			BottomLeft: cfg.BorderRadius,
		},
	}, 91, 1)

	// Text
	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, nt.text, nt.x+12, nt.y+(h-lh)/2, cfg.FontSize, nt.width-20, uimath.ColorWhite, 1)
	}
}

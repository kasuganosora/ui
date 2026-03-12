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
	inputH     float32 // height of the input area (default 28)
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
		inputH:     28,
	}
}

func (cb *ChatBox) Messages() []ChatMessage   { return cb.messages }
func (cb *ChatBox) InputText() string          { return cb.inputText }
func (cb *ChatBox) SetSize(w, h float32)       { cb.width = w; cb.height = h }
func (cb *ChatBox) SetMaxVisible(n int)        { cb.maxVisible = n }
func (cb *ChatBox) OnSend(fn func(string))     { cb.onSend = fn }
func (cb *ChatBox) InputH() float32            { return cb.inputH }

// InputBounds returns the bounding rectangle for the input area (for external Input widget placement).
func (cb *ChatBox) InputBounds() uimath.Rect {
	b := cb.Bounds()
	if b.IsEmpty() {
		b = uimath.NewRect(0, 0, cb.width, cb.height)
	}
	return uimath.NewRect(b.X, b.Y+b.Height-cb.inputH, b.Width, cb.inputH)
}

// MessageBounds returns the bounding rectangle for the message area (for external scrollable Div placement).
func (cb *ChatBox) MessageBounds() uimath.Rect {
	b := cb.Bounds()
	if b.IsEmpty() {
		b = uimath.NewRect(0, 0, cb.width, cb.height)
	}
	msgH := b.Height - cb.inputH
	return uimath.NewRect(b.X, b.Y, b.Width, msgH)
}

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

// HandleWheel scrolls the message area. deltaY < 0 = scroll up, > 0 = scroll down.
// Returns true if the event was consumed (mouse was within chat bounds).
func (cb *ChatBox) HandleWheel(x, y, deltaY float32) bool {
	b := cb.Bounds()
	if b.IsEmpty() {
		return false
	}
	if x < b.X || x >= b.X+b.Width || y < b.Y || y >= b.Y+b.Height {
		return false
	}
	if deltaY < 0 {
		cb.ScrollUp()
	} else if deltaY > 0 {
		cb.ScrollDown()
	}
	return true
}

func (cb *ChatBox) Draw(buf *render.CommandBuffer) {
	bounds := cb.Bounds()
	if bounds.IsEmpty() {
		bounds = uimath.NewRect(0, 0, cb.width, cb.height)
	}

	cfg := cb.Config()

	// Background (message area only, input area handled by external Input widget)
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(bounds.X, bounds.Y, bounds.Width, bounds.Height-cb.inputH),
		FillColor: uimath.RGBA(0, 0, 0, 0.6),
		Corners: uimath.Corners{
			TopLeft:  cfg.BorderRadius,
			TopRight: cfg.BorderRadius,
		},
	}, 1, 1)

	// Messages are rendered by an external scrollable Div — not drawn here.
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
	buf.DrawRect(render.RectCmd{
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
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(nt.x, nt.y, nt.width, h),
		FillColor: uimath.RGBA(0.1, 0.1, 0.1, 0.9),
		Corners:   uimath.CornersAll(cfg.BorderRadius),
	}, 90, 1)

	// Color bar
	buf.DrawRect(render.RectCmd{
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

package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// NotificationType determines the visual style.
type NotificationType uint8

const (
	NotificationInfo NotificationType = iota
	NotificationSuccess
	NotificationWarning
	NotificationError
)

// Notification is a stacking toast notification.
type Notification struct {
	Base
	title    string
	message  string
	ntype    NotificationType
	visible  bool
	x, y     float32
	width    float32
	onClose  func()
}

func NewNotification(tree *core.Tree, title, message string, cfg *Config) *Notification {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Notification{
		Base:    NewBase(tree, core.TypeCustom, cfg),
		title:   title,
		message: message,
		ntype:   NotificationInfo,
		visible: true,
		width:   320,
	}
}

func (n *Notification) Title() string              { return n.title }
func (n *Notification) Message() string            { return n.message }
func (n *Notification) IsVisible() bool            { return n.visible }
func (n *Notification) SetTitle(t string)          { n.title = t }
func (n *Notification) SetMessage(m string)        { n.message = m }
func (n *Notification) SetType(t NotificationType) { n.ntype = t }
func (n *Notification) SetPosition(x, y float32)   { n.x = x; n.y = y }
func (n *Notification) OnClose(fn func())          { n.onClose = fn }

func (n *Notification) Show() { n.visible = true }
func (n *Notification) Close() {
	n.visible = false
	if n.onClose != nil {
		n.onClose()
	}
}

func notificationColor(t NotificationType) uimath.Color {
	switch t {
	case NotificationSuccess:
		return uimath.ColorHex("#52c41a")
	case NotificationWarning:
		return uimath.ColorHex("#faad14")
	case NotificationError:
		return uimath.ColorHex("#ff4d4f")
	default:
		return uimath.ColorHex("#1890ff")
	}
}

func (n *Notification) Draw(buf *render.CommandBuffer) {
	if !n.visible {
		return
	}
	cfg := n.config
	h := float32(80)

	// Shadow
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(n.x+2, n.y+2, n.width, h),
		FillColor: uimath.RGBA(0, 0, 0, 0.1),
		Corners:   uimath.CornersAll(cfg.BorderRadius),
	}, 50, 1)

	// Background
	buf.DrawOverlay(render.RectCmd{
		Bounds:      uimath.NewRect(n.x, n.y, n.width, h),
		FillColor:   uimath.ColorWhite,
		BorderColor: cfg.BorderColor,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 51, 1)

	// Color accent bar
	accentColor := notificationColor(n.ntype)
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(n.x, n.y, 4, h),
		FillColor: accentColor,
		Corners: uimath.Corners{
			TopLeft:    cfg.BorderRadius,
			BottomLeft: cfg.BorderRadius,
		},
	}, 52, 1)

	// Title
	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, n.title, n.x+cfg.SpaceMD, n.y+cfg.SpaceSM, cfg.FontSize, n.width-cfg.SpaceMD*2, cfg.TextColor, 1)
		// Message
		if n.message != "" {
			cfg.TextRenderer.DrawText(buf, n.message, n.x+cfg.SpaceMD, n.y+cfg.SpaceSM+lh+4, cfg.FontSizeSm, n.width-cfg.SpaceMD*2, cfg.DisabledColor, 1)
		}
	}
}

package widget

import (
	"time"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// NotificationTheme determines the visual style.
type NotificationTheme uint8

const (
	NotificationThemeInfo NotificationTheme = iota
	NotificationThemeSuccess
	NotificationThemeWarning
	NotificationThemeError
)

// Notification is a stacking toast notification.
type Notification struct {
	Base
	title           string
	content         string
	theme           NotificationTheme
	visible         bool
	x, y            float32
	width           float32
	onClose         func()
	onCloseBtnClick func()
	onDurationEnd   func()
	duration        int // milliseconds, default 3000; 0 = no auto-dismiss
	closeBtn        bool
	footer          string
	startTime       time.Time
	closeID         core.ElementID
}

func NewNotification(tree *core.Tree, title, content string, cfg *Config) *Notification {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	n := &Notification{
		Base:     NewBase(tree, core.TypeCustom, cfg),
		title:    title,
		content:  content,
		theme:    NotificationThemeInfo,
		visible:  true,
		width:    320,
		duration: 3000,
		closeBtn: true,
	}
	n.startTime = time.Now()

	// Close button element
	n.closeID = tree.CreateElement(core.TypeCustom)
	tree.AppendChild(n.id, n.closeID)
	tree.AddHandler(n.closeID, event.MouseClick, func(e *event.Event) {
		if n.onCloseBtnClick != nil {
			n.onCloseBtnClick()
		}
		n.Close()
	})

	return n
}

func (n *Notification) Title() string                { return n.title }
func (n *Notification) Content() string              { return n.content }
func (n *Notification) IsVisible() bool              { return n.visible }
func (n *Notification) SetTitle(t string)            { n.title = t }
func (n *Notification) SetContent(c string)          { n.content = c }
func (n *Notification) SetTheme(t NotificationTheme) { n.theme = t }
func (n *Notification) SetPosition(x, y float32)     { n.x = x; n.y = y }
func (n *Notification) OnClose(fn func())            { n.onClose = fn }
func (n *Notification) OnCloseBtnClick(fn func())    { n.onCloseBtnClick = fn }
func (n *Notification) OnDurationEnd(fn func())      { n.onDurationEnd = fn }
func (n *Notification) SetDuration(ms int)           { n.duration = ms }
func (n *Notification) SetCloseBtn(c bool)           { n.closeBtn = c }
func (n *Notification) SetFooter(f string)           { n.footer = f }

func (n *Notification) Show() {
	n.visible = true
	n.startTime = time.Now()
}

func (n *Notification) Close() {
	n.visible = false
	if n.onClose != nil {
		n.onClose()
	}
}

func notificationColor(t NotificationTheme) uimath.Color {
	switch t {
	case NotificationThemeSuccess:
		return uimath.ColorHex("#2ba471")
	case NotificationThemeWarning:
		return uimath.ColorHex("#e37318")
	case NotificationThemeError:
		return uimath.ColorHex("#d54941")
	default:
		return uimath.ColorHex("#0052d9")
	}
}

func notificationIconInfo(t NotificationTheme) (string, string) {
	switch t {
	case NotificationThemeSuccess:
		return "check_circle", "\u2713"
	case NotificationThemeWarning:
		return "warning", "!"
	case NotificationThemeError:
		return "cancel", "\u00d7"
	default:
		return "info", "i"
	}
}

func (n *Notification) Draw(buf *render.CommandBuffer) {
	if !n.visible {
		return
	}

	// Auto-dismiss after duration
	if n.duration > 0 && time.Since(n.startTime) > time.Duration(n.duration)*time.Millisecond {
		if n.onDurationEnd != nil {
			n.onDurationEnd()
		}
		n.Close()
		return
	}
	// Keep redrawing for timer
	if n.duration > 0 {
		n.tree.MarkDirty(n.id)
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
	accentColor := notificationColor(n.theme)
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(n.x, n.y, 4, h),
		FillColor: accentColor,
		Corners: uimath.Corners{
			TopLeft:    cfg.BorderRadius,
			BottomLeft: cfg.BorderRadius,
		},
	}, 52, 1)

	// Status icon in accent bar area
	{
		iconName, iconFallback := notificationIconInfo(n.theme)
		iconSize := float32(16)
		iconX := n.x + 12
		iconY := n.y + (h-iconSize)/2
		if !cfg.DrawMDIconOverlay(buf, iconName, iconX, iconY, iconSize, accentColor, 53, 1) {
			buf.DrawOverlay(render.RectCmd{
				Bounds:    uimath.NewRect(iconX, iconY, iconSize, iconSize),
				FillColor: accentColor,
				Corners:   uimath.CornersAll(iconSize / 2),
			}, 53, 1)
			if cfg.TextRenderer != nil {
				iconFontSize := float32(10)
				tw := cfg.TextRenderer.MeasureText(iconFallback, iconFontSize)
				lh := cfg.TextRenderer.LineHeight(iconFontSize)
				cfg.TextRenderer.DrawText(buf, iconFallback,
					iconX+(iconSize-tw)/2, iconY+(iconSize-lh)/2,
					iconFontSize, iconSize, uimath.ColorWhite, 1)
			}
		}
	}

	// Title and content text
	textLeftOffset := float32(36) // after icon area
	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, n.title, n.x+textLeftOffset, n.y+cfg.SpaceSM, cfg.FontSize, n.width-textLeftOffset-cfg.SpaceMD, cfg.TextColor, 1)
		// Content
		if n.content != "" {
			cfg.TextRenderer.DrawText(buf, n.content, n.x+textLeftOffset, n.y+cfg.SpaceSM+lh+4, cfg.FontSizeSm, n.width-textLeftOffset-cfg.SpaceMD, cfg.DisabledColor, 1)
		}
	}

	// Close button (X) in top-right
	if n.closeBtn {
		closeSize := float32(12)
		closeX := n.x + n.width - cfg.SpaceMD - closeSize
		closeY := n.y + cfg.SpaceSM

		// Draw close icon
		if !cfg.DrawMDIconOverlay(buf, "close", closeX, closeY, closeSize, cfg.DisabledColor, 54, 1) {
			buf.DrawOverlay(render.RectCmd{
				Bounds:    uimath.NewRect(closeX+closeSize/2-1, closeY, 2, closeSize),
				FillColor: cfg.DisabledColor,
			}, 54, 1)
			buf.DrawOverlay(render.RectCmd{
				Bounds:    uimath.NewRect(closeX, closeY+closeSize/2-1, closeSize, 2),
				FillColor: cfg.DisabledColor,
			}, 54, 1)
		}

		// Set layout on close element for hit testing
		closeHit := closeSize + 8
		n.tree.SetLayout(n.closeID, core.LayoutResult{
			Bounds: uimath.NewRect(closeX-4, closeY-4, closeHit, closeHit),
		})
	}
}

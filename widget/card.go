package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// CardTheme controls the card visual layout style.
type CardTheme uint8

const (
	CardThemeNormal  CardTheme = iota // default style
	CardThemePoster1                  // actions in header
	CardThemePoster2                  // actions in footer
)

// Card is a container with optional header and footer.
type Card struct {
	Base
	title          string
	subtitle       string
	bordered       bool
	bgColor        uimath.Color
	actions        Widget
	footer         Widget
	hoverShadow    bool
	headerBordered bool
	size           Size
	hovered        bool
	description    string
	cover          string
	avatar         string
	shadow         bool
	status         string
	theme          CardTheme
	loading        bool
}

func NewCard(tree *core.Tree, cfg *Config) *Card {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	c := &Card{
		Base:           NewBase(tree, core.TypeDiv, cfg),
		bordered:       true,
		bgColor:        uimath.ColorWhite,
		headerBordered: true,
		size:           SizeMedium,
	}
	c.style.Display = layout.DisplayFlex
	c.style.FlexDirection = layout.FlexDirectionColumn

	// Track hover for shadow
	tree.AddHandler(c.id, event.MouseEnter, func(e *event.Event) {
		c.hovered = true
	})
	tree.AddHandler(c.id, event.MouseLeave, func(e *event.Event) {
		c.hovered = false
	})

	return c
}

func (c *Card) SetTitle(t string)          { c.title = t }
func (c *Card) SetSubtitle(s string)       { c.subtitle = s }
func (c *Card) SetBordered(b bool)         { c.bordered = b }
func (c *Card) SetBgColor(cl uimath.Color) { c.bgColor = cl }
func (c *Card) SetActions(w Widget)        { c.actions = w }
func (c *Card) SetFooter(w Widget)         { c.footer = w }
func (c *Card) SetHoverShadow(v bool)      { c.hoverShadow = v }
func (c *Card) SetHeaderBordered(v bool)   { c.headerBordered = v }
func (c *Card) SetSize(s Size)             { c.size = s }
func (c *Card) SetDescription(d string)    { c.description = d }
func (c *Card) SetCover(url string)        { c.cover = url }
func (c *Card) SetAvatar(url string)       { c.avatar = url }
func (c *Card) SetShadow(v bool)           { c.shadow = v }
func (c *Card) SetStatus(s string)         { c.status = s }
func (c *Card) SetTheme(t CardTheme)       { c.theme = t }
func (c *Card) SetLoading(v bool)          { c.loading = v }

// Deprecated: Use SetActions instead.
func (c *Card) SetHeaderExtra(w Widget) { c.actions = w }

func (c *Card) cardPadding() float32 {
	if c.size == SizeSmall {
		return c.config.SpaceSM
	}
	return c.config.SpaceLG
}

func (c *Card) Draw(buf *render.CommandBuffer) {
	bounds := c.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := c.config
	pad := c.cardPadding()

	// Hover shadow
	if c.hoverShadow && c.hovered {
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(bounds.X+2, bounds.Y+2, bounds.Width+4, bounds.Height+4),
			FillColor: uimath.RGBA(0, 0, 0, 0.08),
			Corners:   uimath.CornersAll(cfg.BorderRadius + 2),
		}, -1, 1)
	}

	borderW := float32(0)
	borderC := uimath.Color{}
	if c.bordered {
		borderW = cfg.BorderWidth
		borderC = cfg.BorderColor
	}

	buf.DrawRect(render.RectCmd{
		Bounds:      bounds,
		FillColor:   c.bgColor,
		BorderColor: borderC,
		BorderWidth: borderW,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 0, 1)

	headerH := float32(0)
	if c.title != "" {
		headerH = 48
		if c.subtitle != "" {
			headerH = 60
		}
		if c.size == SizeSmall {
			headerH = 40
			if c.subtitle != "" {
				headerH = 52
			}
		}

		// Header divider
		if c.headerBordered {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X, bounds.Y+headerH-1, bounds.Width, 1),
				FillColor: uimath.RGBA(0, 0, 0, 0.06),
			}, 1, 1)
		}

		// Title
		titleMaxW := bounds.Width - pad*2
		headerExtraW := float32(0)
		if c.actions != nil {
			// Reserve space for headerExtra
			headerExtraW = 80 // default reservation
		}
		titleMaxW -= headerExtraW

		titleY := bounds.Y + (headerH-cfg.FontSize*1.2)/2
		if c.subtitle != "" {
			titleY = bounds.Y + pad/2
		}

		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			if c.subtitle == "" {
				titleY = bounds.Y + (headerH-lh)/2
			}
			cfg.TextRenderer.DrawText(buf, c.title, bounds.X+pad, titleY, cfg.FontSize, titleMaxW, cfg.TextColor, 1)
		} else {
			tw := float32(len(c.title)) * cfg.FontSize * 0.55
			th := cfg.FontSize * 1.2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X+pad, titleY, tw, th),
				FillColor: cfg.TextColor,
				Corners:   uimath.CornersAll(2),
			}, 1, 1)
		}

		// Subtitle
		if c.subtitle != "" {
			subtitleY := titleY + cfg.FontSize*1.4
			subtitleColor := uimath.RGBA(0, 0, 0, 0.45)
			if cfg.TextRenderer != nil {
				cfg.TextRenderer.DrawText(buf, c.subtitle, bounds.X+pad, subtitleY, cfg.FontSizeSm, titleMaxW, subtitleColor, 1)
			} else {
				sw := float32(len(c.subtitle)) * cfg.FontSizeSm * 0.55
				sh := cfg.FontSizeSm * 1.2
				buf.DrawRect(render.RectCmd{
					Bounds:    uimath.NewRect(bounds.X+pad, subtitleY, sw, sh),
					FillColor: subtitleColor,
					Corners:   uimath.CornersAll(2),
				}, 1, 1)
			}
		}

		// Header extra widget
		if c.actions != nil {
			c.actions.Draw(buf)
		}
	}

	_ = headerH
	c.DrawChildren(buf)

	// Footer
	if c.footer != nil {
		footerH := float32(48)
		if c.size == SizeSmall {
			footerH = 40
		}
		footerY := bounds.Y + bounds.Height - footerH

		// Footer top divider
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(bounds.X, footerY, bounds.Width, 1),
			FillColor: uimath.RGBA(0, 0, 0, 0.06),
		}, 1, 1)

		c.footer.Draw(buf)
	}
}

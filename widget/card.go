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

// Card is a container with optional header, body content, and footer.
// Matches TDesign Card component: bordered/borderless, shadow, header divider,
// title + subtitle + actions, footer with action icons.
type Card struct {
	Base
	title          string
	subtitle       string
	description    string
	bordered       bool
	shadow         bool // show shadow (useful for borderless cards)
	bgColor        uimath.Color
	actions        Widget // header right-side actions
	footer         Widget // footer content
	footerActions  []Widget // footer action items (like, comment, share icons)
	hoverShadow    bool
	headerBordered bool // divider between header and body
	size           Size
	hovered        bool
	cover          string // cover image URL
	avatar         string // avatar URL
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

	tree.AddHandler(c.id, event.MouseEnter, func(e *event.Event) {
		c.hovered = true
	})
	tree.AddHandler(c.id, event.MouseLeave, func(e *event.Event) {
		c.hovered = false
	})

	c.updatePadding()
	return c
}

func (c *Card) SetTitle(t string) {
	c.title = t
	c.updatePadding()
}
func (c *Card) SetSubtitle(s string) {
	c.subtitle = s
	c.updatePadding()
}
func (c *Card) SetDescription(d string)    { c.description = d }
func (c *Card) SetBordered(b bool)         { c.bordered = b }
func (c *Card) SetShadow(v bool)          { c.shadow = v }
func (c *Card) SetBgColor(cl uimath.Color) { c.bgColor = cl }
func (c *Card) SetActions(w Widget)        { c.actions = w }
func (c *Card) SetFooter(w Widget)         { c.footer = w; c.updatePadding() }
func (c *Card) SetHoverShadow(v bool)      { c.hoverShadow = v }
func (c *Card) SetHeaderBordered(v bool)   { c.headerBordered = v }
func (c *Card) SetSize(s Size)             { c.size = s; c.updatePadding() }
func (c *Card) SetCover(url string)        { c.cover = url }
func (c *Card) SetAvatar(url string)       { c.avatar = url }
func (c *Card) SetStatus(s string)         { c.status = s }
func (c *Card) SetTheme(t CardTheme)       { c.theme = t }
func (c *Card) SetLoading(v bool)          { c.loading = v }
func (c *Card) Title() string              { return c.title }
func (c *Card) Bordered() bool             { return c.bordered }
func (c *Card) Shadow() bool               { return c.shadow }
func (c *Card) HeaderBordered() bool       { return c.headerBordered }

// AddFooterAction adds a widget to the footer action bar.
func (c *Card) AddFooterAction(w Widget) {
	c.footerActions = append(c.footerActions, w)
	c.updatePadding()
}

// FooterActions returns the footer action widgets.
func (c *Card) FooterActions() []Widget { return c.footerActions }

// Deprecated: Use SetActions instead.
func (c *Card) SetHeaderExtra(w Widget) { c.actions = w }

func (c *Card) cardPadding() float32 {
	if c.size == SizeSmall {
		return c.config.SpaceSM
	}
	return c.config.SpaceLG
}

// headerHeight returns the height of the header area (0 if no title).
func (c *Card) headerHeight() float32 {
	if c.title == "" {
		return 0
	}
	if c.size == SizeSmall {
		if c.subtitle != "" {
			return 52
		}
		return 40
	}
	if c.subtitle != "" {
		return 64
	}
	return 48
}

// footerHeight returns the height of the footer area.
func (c *Card) footerHeight() float32 {
	if c.footer == nil && len(c.footerActions) == 0 {
		return 0
	}
	if c.size == SizeSmall {
		return 40
	}
	return 48
}

// updatePadding adjusts the style padding so children don't overlap header/footer.
func (c *Card) updatePadding() {
	pad := c.cardPadding()
	c.style.Padding = layout.EdgeValues{
		Top:    layout.Px(c.headerHeight() + pad),
		Bottom: layout.Px(c.footerHeight() + pad),
		Left:   layout.Px(pad),
		Right:  layout.Px(pad),
	}
}

func (c *Card) Draw(buf *render.CommandBuffer) {
	bounds := c.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := c.config
	pad := c.cardPadding()
	radius := cfg.BorderRadius

	// Shadow (drawn beneath everything)
	if c.shadow || (c.hoverShadow && c.hovered) {
		spread := float32(4)
		offsetY := float32(2)
		shadowAlpha := float32(0.08)
		if c.hoverShadow && c.hovered {
			spread = 8
			offsetY = 4
			shadowAlpha = 0.15
		}
		buf.DrawRect(render.RectCmd{
			Bounds: uimath.NewRect(
				bounds.X-spread, bounds.Y-spread+offsetY,
				bounds.Width+spread*2, bounds.Height+spread*2,
			),
			FillColor: uimath.RGBA(0, 0, 0, shadowAlpha),
			Corners:   uimath.CornersAll(radius + spread),
		}, -1, 1)
	}

	// Card background
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
		Corners:     uimath.CornersAll(radius),
	}, 0, 1)

	y := bounds.Y

	// Header section (title + subtitle + actions)
	headerH := c.headerHeight()
	if c.title != "" {

		// Title
		titleFs := cfg.FontSizeLg
		if c.size == SizeSmall {
			titleFs = cfg.FontSize
		}
		titleMaxW := bounds.Width - pad*2
		if c.actions != nil {
			titleMaxW -= 80
		}

		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(titleFs)
			titleY := y + (headerH-lh)/2
			if c.subtitle != "" {
				titleY = y + pad*0.6
			}
			// Title text (bold would be set by font, we draw it larger)
			cfg.TextRenderer.DrawText(buf, c.title,
				bounds.X+pad, titleY,
				titleFs, titleMaxW, cfg.TextColor, 1)

			// Subtitle
			if c.subtitle != "" {
				subtitleFs := cfg.FontSizeSm
				subtitleColor := uimath.RGBA(0, 0, 0, 0.4)
				subY := titleY + lh + 2
				cfg.TextRenderer.DrawText(buf, c.subtitle,
					bounds.X+pad, subY,
					subtitleFs, titleMaxW, subtitleColor, 1)
			}
		} else {
			// Placeholder rects
			titleY := y + (headerH-titleFs*1.2)/2
			if c.subtitle != "" {
				titleY = y + pad*0.6
			}
			tw := float32(len(c.title)) * titleFs * 0.55
			th := titleFs * 1.2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X+pad, titleY, tw, th),
				FillColor: cfg.TextColor,
				Corners:   uimath.CornersAll(2),
			}, 1, 1)

			if c.subtitle != "" {
				sw := float32(len(c.subtitle)) * cfg.FontSizeSm * 0.55
				sh := cfg.FontSizeSm * 1.2
				buf.DrawRect(render.RectCmd{
					Bounds:    uimath.NewRect(bounds.X+pad, titleY+th+4, sw, sh),
					FillColor: uimath.RGBA(0, 0, 0, 0.4),
					Corners:   uimath.CornersAll(2),
				}, 1, 1)
			}
		}

		// Actions (right side of header)
		if c.actions != nil {
			c.actions.Draw(buf)
		}

		// Header divider line
		if c.headerBordered {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X, y+headerH-1, bounds.Width, 1),
				FillColor: uimath.RGBA(0, 0, 0, 0.06),
			}, 1, 1)
		}

		y += headerH
	}

	// Body (children)
	c.DrawChildren(buf)

	// Footer section
	footerH := c.footerHeight()
	if footerH > 0 {
		footerY := bounds.Y + bounds.Height - footerH

		// Footer top divider
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(bounds.X, footerY, bounds.Width, 1),
			FillColor: uimath.RGBA(0, 0, 0, 0.06),
		}, 1, 1)

		if c.footer != nil {
			c.footer.Draw(buf)
		}

		// Footer actions (evenly distributed)
		if len(c.footerActions) > 0 {
			actionW := bounds.Width / float32(len(c.footerActions))
			ax := bounds.X
			for i, act := range c.footerActions {
				// Draw vertical divider between actions (not before first)
				if i > 0 {
					buf.DrawRect(render.RectCmd{
						Bounds:    uimath.NewRect(ax, footerY+8, 1, footerH-16),
						FillColor: uimath.RGBA(0, 0, 0, 0.06),
					}, 1, 1)
				}
				act.Draw(buf)
				ax += actionW
			}
		}
	}
	_ = footerH
}

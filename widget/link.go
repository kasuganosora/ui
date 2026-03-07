package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Link is a clickable text link.
type Link struct {
	Base
	text     string
	href     string
	disabled bool
	onClick  func(href string)
}

func NewLink(tree *core.Tree, text, href string, cfg *Config) *Link {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	l := &Link{
		Base: NewBase(tree, core.TypeButton, cfg),
		text: text,
		href: href,
	}
	tree.SetProperty(l.id, "text", text)
	tree.AddHandler(l.id, event.MouseClick, func(e *event.Event) {
		if !l.disabled && l.onClick != nil {
			l.onClick(l.href)
		}
	})
	return l
}

func (l *Link) Text() string         { return l.text }
func (l *Link) Href() string         { return l.href }
func (l *Link) SetText(t string)     { l.text = t; l.tree.SetProperty(l.id, "text", t) }
func (l *Link) SetHref(h string)     { l.href = h }
func (l *Link) SetDisabled(d bool)   { l.disabled = d }
func (l *Link) OnClick(fn func(string)) { l.onClick = fn }

func (l *Link) Draw(buf *render.CommandBuffer) {
	bounds := l.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := l.config
	color := cfg.PrimaryColor
	if l.disabled {
		color = cfg.DisabledColor
	} else if elem := l.Element(); elem != nil && elem.IsHovered() {
		color = cfg.HoverColor
	}
	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, l.text, bounds.X, bounds.Y+(bounds.Height-lh)/2, cfg.FontSize, bounds.Width, color, 1)
	} else {
		tw := float32(len(l.text)) * cfg.FontSize * 0.55
		th := cfg.FontSize * 1.2
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(bounds.X, bounds.Y+(bounds.Height-th)/2, tw, th),
			FillColor: color,
			Corners:   uimath.CornersAll(2),
		}, 1, 1)
	}
	// Underline
	underY := bounds.Y + bounds.Height - 2
	tw := bounds.Width
	if cfg.TextRenderer != nil {
		tw = cfg.TextRenderer.MeasureText(l.text, cfg.FontSize)
	}
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(bounds.X, underY, tw, 1),
		FillColor: color,
	}, 1, 0.5)
}

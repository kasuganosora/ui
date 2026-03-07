package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// TagType controls the tag appearance.
type TagType uint8

const (
	TagDefault TagType = iota
	TagSuccess
	TagWarning
	TagError
	TagProcessing
)

// Tag displays a small labeled tag/badge.
type Tag struct {
	Base
	label   string
	tagType TagType
	color   uimath.Color // custom color (zero = use tagType default)
}

// NewTag creates a tag with the given label.
func NewTag(tree *core.Tree, label string, cfg *Config) *Tag {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	t := &Tag{
		Base:  NewBase(tree, core.TypeCustom, cfg),
		label: label,
	}
	t.style.Display = layout.DisplayFlex
	t.style.AlignItems = layout.AlignCenter
	t.style.JustifyContent = layout.JustifyCenter
	t.style.Height = layout.Px(22)
	t.style.Padding = layout.EdgeValues{
		Left:  layout.Px(cfg.SpaceSM),
		Right: layout.Px(cfg.SpaceSM),
	}
	tree.SetProperty(t.id, "text", label)
	return t
}

func (t *Tag) Label() string   { return t.label }
func (t *Tag) TagType() TagType { return t.tagType }

func (t *Tag) SetLabel(label string) {
	t.label = label
	t.tree.SetProperty(t.id, "text", label)
}

func (t *Tag) SetTagType(tt TagType) { t.tagType = tt }
func (t *Tag) SetColor(c uimath.Color) { t.color = c }

func (t *Tag) tagColors() (bg, border, text uimath.Color) {
	if t.color != (uimath.Color{}) {
		return uimath.Color{R: t.color.R, G: t.color.G, B: t.color.B, A: 0.1},
			t.color, t.color
	}
	switch t.tagType {
	case TagSuccess:
		return uimath.ColorHex("#f6ffed"), uimath.ColorHex("#b7eb8f"), uimath.ColorHex("#52c41a")
	case TagWarning:
		return uimath.ColorHex("#fffbe6"), uimath.ColorHex("#ffe58f"), uimath.ColorHex("#faad14")
	case TagError:
		return uimath.ColorHex("#fff2f0"), uimath.ColorHex("#ffccc7"), uimath.ColorHex("#ff4d4f")
	case TagProcessing:
		return uimath.ColorHex("#e6f4ff"), uimath.ColorHex("#91caff"), uimath.ColorHex("#1677ff")
	default:
		return uimath.ColorHex("#fafafa"), uimath.ColorHex("#d9d9d9"), t.config.TextColor
	}
}

func (t *Tag) Draw(buf *render.CommandBuffer) {
	bounds := t.Bounds()
	if bounds.IsEmpty() {
		return
	}

	cfg := t.config
	bgColor, borderColor, textColor := t.tagColors()

	// Background
	buf.DrawRect(render.RectCmd{
		Bounds:      bounds,
		FillColor:   bgColor,
		BorderColor: borderColor,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius / 2),
	}, 0, 1)

	// Label
	if t.label != "" {
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			tx := bounds.X + cfg.SpaceSM
			ty := bounds.Y + (bounds.Height-lh)/2
			maxW := bounds.Width - cfg.SpaceSM*2
			cfg.TextRenderer.DrawText(buf, t.label, tx, ty, cfg.FontSizeSm, maxW, textColor, 1)
		} else {
			textW := float32(len(t.label)) * cfg.FontSizeSm * 0.55
			maxW := bounds.Width - cfg.SpaceSM*2
			if textW > maxW {
				textW = maxW
			}
			textH := cfg.FontSizeSm * 1.2
			tx := bounds.X + (bounds.Width-textW)/2
			ty := bounds.Y + (bounds.Height-textH)/2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(tx, ty, textW, textH),
				FillColor: textColor,
				Corners:   uimath.CornersAll(2),
			}, 1, 1)
		}
	}

	t.DrawChildren(buf)
}

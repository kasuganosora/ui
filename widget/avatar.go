package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// AvatarShape controls the avatar shape.
type AvatarShape uint8

const (
	AvatarCircle AvatarShape = iota
	AvatarSquare
)

// Avatar displays a user avatar (image or initials).
type Avatar struct {
	Base
	text    string
	icon    render.TextureHandle
	bgColor uimath.Color
	size    float32
	shape   AvatarShape
}

func NewAvatar(tree *core.Tree, cfg *Config) *Avatar {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Avatar{
		Base:    NewBase(tree, core.TypeCustom, cfg),
		bgColor: uimath.ColorHex("#1677ff"),
		size:    32,
		shape:   AvatarCircle,
	}
}

func (a *Avatar) SetText(t string)              { a.text = t }
func (a *Avatar) SetIcon(h render.TextureHandle) { a.icon = h }
func (a *Avatar) SetBgColor(c uimath.Color)     { a.bgColor = c }
func (a *Avatar) SetSize(s float32)              { a.size = s }
func (a *Avatar) SetShape(s AvatarShape)         { a.shape = s }

func (a *Avatar) Draw(buf *render.CommandBuffer) {
	bounds := a.Bounds()
	if bounds.IsEmpty() {
		bounds = uimath.NewRect(0, 0, a.size, a.size)
	}

	cfg := a.config
	radius := float32(0)
	if a.shape == AvatarCircle {
		radius = bounds.Width / 2
	} else {
		radius = cfg.BorderRadius
	}

	buf.DrawRect(render.RectCmd{
		Bounds:    bounds,
		FillColor: a.bgColor,
		Corners:   uimath.CornersAll(radius),
	}, 0, 1)

	if a.icon != 0 {
		buf.DrawImage(render.ImageCmd{
			Texture: a.icon,
			DstRect: bounds,
			Tint:    uimath.ColorWhite,
			Corners: uimath.CornersAll(radius),
		}, 1, 1)
	} else if a.text != "" && cfg.TextRenderer != nil {
		fs := bounds.Height * 0.45
		lh := cfg.TextRenderer.LineHeight(fs)
		tw := cfg.TextRenderer.MeasureText(a.text, fs)
		cfg.TextRenderer.DrawText(buf, a.text, bounds.X+(bounds.Width-tw)/2, bounds.Y+(bounds.Height-lh)/2, fs, bounds.Width, uimath.ColorWhite, 1)
	}
}

package widget

import (
	"strconv"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// AvatarShape controls the avatar shape.
type AvatarShape uint8

const (
	AvatarCircle AvatarShape = iota
	AvatarRound
	AvatarSquare
)

// Avatar displays a user avatar (image or initials).
type Avatar struct {
	Base
	content          string
	alt              string
	icon             render.TextureHandle
	image            string // image URL/path
	bgColor          uimath.Color
	size             Size // SizeSmall=24, SizeMedium=40, SizeLarge=64
	shape            AvatarShape
	hideOnLoadFailed bool
	onError          func()
}

// avatarSizePx returns the pixel dimension for a given avatar size.
func avatarSizePx(s Size) float32 {
	switch s {
	case SizeSmall:
		return 24
	case SizeLarge:
		return 64
	default: // SizeMedium
		return 40
	}
}

func NewAvatar(tree *core.Tree, cfg *Config) *Avatar {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Avatar{
		Base:    NewBase(tree, core.TypeCustom, cfg),
		bgColor: uimath.ColorHex("#1677ff"),
		size:    SizeMedium,
		shape:   AvatarCircle,
	}
}

func (a *Avatar) SetContent(t string)            { a.content = t }
func (a *Avatar) SetAlt(alt string)              { a.alt = alt }
func (a *Avatar) SetIcon(h render.TextureHandle) { a.icon = h }
func (a *Avatar) SetImage(path string)           { a.image = path }
func (a *Avatar) SetBgColor(c uimath.Color)      { a.bgColor = c }
func (a *Avatar) SetSize(s Size)                 { a.size = s }
func (a *Avatar) SetShape(s AvatarShape)         { a.shape = s }
func (a *Avatar) SetHideOnLoadFailed(v bool)     { a.hideOnLoadFailed = v }
func (a *Avatar) OnError(fn func())              { a.onError = fn }
func (a *Avatar) Image() string                  { return a.image }
func (a *Avatar) AvatarSize() float32            { return avatarSizePx(a.size) }

// Deprecated: Use SetContent instead.
func (a *Avatar) SetText(t string) { a.content = t }

func (a *Avatar) Draw(buf *render.CommandBuffer) {
	px := avatarSizePx(a.size)
	bounds := a.Bounds()
	if bounds.IsEmpty() {
		bounds = uimath.NewRect(0, 0, px, px)
	}

	cfg := a.config
	radius := float32(0)
	switch a.shape {
	case AvatarCircle:
		radius = bounds.Width / 2
	case AvatarRound:
		radius = cfg.BorderRadius
	default: // AvatarSquare
		radius = 0
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
	} else if a.content != "" && cfg.TextRenderer != nil {
		fs := bounds.Height * 0.45
		lh := cfg.TextRenderer.LineHeight(fs)
		tw := cfg.TextRenderer.MeasureText(a.content, fs)
		cfg.TextRenderer.DrawText(buf, a.content, bounds.X+(bounds.Width-tw)/2, bounds.Y+(bounds.Height-lh)/2, fs, bounds.Width, uimath.ColorWhite, 1)
	}
}

// CascadingValue controls the stacking direction of avatars.
type CascadingValue uint8

const (
	CascadingRightUp CascadingValue = iota // right avatar on top (default)
	CascadingLeftUp                        // left avatar on top
)

// AvatarGroup holds multiple Avatars and draws them overlapping.
type AvatarGroup struct {
	Base
	avatars        []*Avatar
	max            int            // max visible avatars (0 = show all)
	spacing        float32        // overlap offset (negative = overlap), default -8
	size           Size           // size applied to all child avatars
	cascading      CascadingValue // stacking direction
	collapseAvatar string         // custom text for the "+N" element
}

func NewAvatarGroup(tree *core.Tree, cfg *Config) *AvatarGroup {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &AvatarGroup{
		Base:    NewBase(tree, core.TypeCustom, cfg),
		spacing: -8,
		size:    SizeMedium,
	}
}

func (g *AvatarGroup) SetMax(m int)                    { g.max = m }
func (g *AvatarGroup) SetSpacing(s float32)            { g.spacing = s }
func (g *AvatarGroup) SetSize(s Size)                  { g.size = s }
func (g *AvatarGroup) SetCascading(c CascadingValue)   { g.cascading = c }
func (g *AvatarGroup) SetCollapseAvatar(text string)   { g.collapseAvatar = text }

// Deprecated: Use SetMax instead.
func (g *AvatarGroup) SetMaxCount(m int) { g.max = m }

// AddAvatar appends an avatar to the group.
func (g *AvatarGroup) AddAvatar(a *Avatar) {
	a.SetSize(g.size)
	g.avatars = append(g.avatars, a)
}

// Avatars returns the group's avatar slice.
func (g *AvatarGroup) Avatars() []*Avatar { return g.avatars }

func (g *AvatarGroup) Draw(buf *render.CommandBuffer) {
	bounds := g.Bounds()
	if bounds.IsEmpty() {
		return
	}

	px := avatarSizePx(g.size)
	visible := g.avatars
	if g.max > 0 && len(visible) > g.max {
		visible = visible[:g.max]
	}

	x := bounds.X
	for _, av := range visible {
		avBounds := uimath.NewRect(x, bounds.Y, px, px)
		// Set layout bounds so the avatar draws at the right position
		lo := av.tree.Get(av.id).Layout()
		lo.Bounds = avBounds
		av.tree.SetLayout(av.id, lo)
		av.Draw(buf)

		x += px + g.spacing
	}

	// If truncated, draw a "+N" indicator
	if g.max > 0 && len(g.avatars) > g.max {
		remaining := len(g.avatars) - g.max
		cfg := g.config
		extraBounds := uimath.NewRect(x, bounds.Y, px, px)
		buf.DrawRect(render.RectCmd{
			Bounds:    extraBounds,
			FillColor: uimath.ColorHex("#c0c0c0"),
			Corners:   uimath.CornersAll(px / 2),
		}, int32(len(visible)), 1)
		if cfg.TextRenderer != nil {
			text := "+" + strconv.Itoa(remaining)
			fs := px * 0.4
			lh := cfg.TextRenderer.LineHeight(fs)
			tw := cfg.TextRenderer.MeasureText(text, fs)
			cfg.TextRenderer.DrawText(buf, text, extraBounds.X+(px-tw)/2, extraBounds.Y+(px-lh)/2, fs, px, uimath.ColorWhite, 1)
		}
	}
}

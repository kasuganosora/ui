package widget

import (
	"strconv"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// AvatarShape controls the avatar shape.
type AvatarShape uint8

const (
	AvatarCircle AvatarShape = iota
	AvatarRound              // rounded square
	AvatarSquare
)

// Avatar displays a user avatar (image, icon, or initials).
type Avatar struct {
	Base
	content          string
	alt              string
	icon             render.TextureHandle
	image            render.TextureHandle // image texture handle
	imagePath        string               // image URL/path (for reference)
	bgColor          uimath.Color
	size             Size    // SizeSmall=32, SizeMedium=48, SizeLarge=64
	customSize       float32 // if > 0, overrides size enum
	shape            AvatarShape
	hideOnLoadFailed bool
	onError          func()
}

// avatarSizePx returns the pixel dimension for a given avatar size.
func avatarSizePx(s Size) float32 {
	switch s {
	case SizeSmall:
		return 32
	case SizeLarge:
		return 64
	default: // SizeMedium
		return 48
	}
}

func NewAvatar(tree *core.Tree, cfg *Config) *Avatar {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	a := &Avatar{
		Base:    NewBase(tree, core.TypeCustom, cfg),
		bgColor: uimath.ColorHex("#bcc4d0"),
		size:    SizeMedium,
		shape:   AvatarCircle,
	}
	px := avatarSizePx(a.size)
	a.style.Width = layout.Px(px)
	a.style.Height = layout.Px(px)
	a.style.FlexShrink = 0
	a.style.FlexGrow = 0
	a.style.Display = layout.DisplayBlock
	return a
}

func (a *Avatar) SetContent(t string)            { a.content = t }
func (a *Avatar) SetAlt(alt string)              { a.alt = alt }
func (a *Avatar) SetIcon(h render.TextureHandle) { a.icon = h }
func (a *Avatar) SetImageTexture(h render.TextureHandle) { a.image = h }
func (a *Avatar) SetImagePath(path string)       { a.imagePath = path }
func (a *Avatar) SetBgColor(c uimath.Color)      { a.bgColor = c }
func (a *Avatar) SetHideOnLoadFailed(v bool)     { a.hideOnLoadFailed = v }
func (a *Avatar) OnError(fn func())              { a.onError = fn }
func (a *Avatar) ImagePath() string              { return a.imagePath }
func (a *Avatar) Content() string                { return a.content }
func (a *Avatar) Shape() AvatarShape             { return a.shape }

// SetSize sets one of the preset sizes (Small=32, Medium=48, Large=64).
func (a *Avatar) SetSize(s Size) {
	a.size = s
	a.customSize = 0
	px := avatarSizePx(s)
	a.style.Width = layout.Px(px)
	a.style.Height = layout.Px(px)
	a.style.FlexShrink = 0
	a.style.FlexGrow = 0
}

// SetCustomSize sets a custom pixel size, overriding the preset.
func (a *Avatar) SetCustomSize(px float32) {
	a.customSize = px
	a.style.Width = layout.Px(px)
	a.style.Height = layout.Px(px)
	a.style.FlexShrink = 0
	a.style.FlexGrow = 0
}

// SetShape sets the avatar shape (Circle, Round, Square).
func (a *Avatar) SetShape(s AvatarShape) { a.shape = s }

// AvatarSize returns the effective pixel size.
func (a *Avatar) AvatarSize() float32 {
	if a.customSize > 0 {
		return a.customSize
	}
	return avatarSizePx(a.size)
}

// Deprecated: Use SetContent instead.
func (a *Avatar) SetText(t string) { a.content = t }

// Deprecated: Use SetImagePath instead.
func (a *Avatar) SetImage(path string) { a.imagePath = path }

// Deprecated: Use ImagePath instead.
func (a *Avatar) Image() string { return a.imagePath }

func (a *Avatar) Draw(buf *render.CommandBuffer) {
	px := a.AvatarSize()
	bounds := a.Bounds()
	if bounds.IsEmpty() {
		bounds = uimath.NewRect(0, 0, px, px)
	}
	// Force square bounds — use the configured size, centered within layout bounds.
	// This prevents flex layout distortion (e.g., circles becoming ellipses).
	if bounds.Width != px || bounds.Height != px {
		cx := bounds.X + bounds.Width/2
		cy := bounds.Y + bounds.Height/2
		bounds = uimath.NewRect(cx-px/2, cy-px/2, px, px)
	}

	// Scale font size based on avatar size
	fontSize := px * 0.45
	if a.customSize > 0 && a.customSize > 64 {
		fontSize = a.customSize * 0.4
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

	// Background
	buf.DrawRect(render.RectCmd{
		Bounds:    bounds,
		FillColor: a.bgColor,
		Corners:   uimath.CornersAll(radius),
	}, 0, 1)

	// Image (texture) takes priority over icon and text
	if a.image != 0 {
		buf.DrawImage(render.ImageCmd{
			Texture: a.image,
			DstRect: bounds,
			Tint:    uimath.ColorWhite,
			Corners: uimath.CornersAll(radius),
		}, 1, 1)
	} else if a.icon != 0 {
		// Icon centered within bounds, sized proportionally
		iconSize := px * 0.5
		iconBounds := uimath.NewRect(
			bounds.X+(bounds.Width-iconSize)/2,
			bounds.Y+(bounds.Height-iconSize)/2,
			iconSize, iconSize,
		)
		buf.DrawImage(render.ImageCmd{
			Texture: a.icon,
			DstRect: iconBounds,
			Tint:    uimath.ColorWhite,
		}, 1, 1)
	} else if a.content != "" {
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(fontSize)
			tw := cfg.TextRenderer.MeasureText(a.content, fontSize)
			cfg.TextRenderer.DrawText(buf, a.content,
				bounds.X+(bounds.Width-tw)/2,
				bounds.Y+(bounds.Height-lh)/2,
				fontSize, bounds.Width, uimath.ColorWhite, 1)
		} else {
			// Placeholder rect for text
			tw := float32(len(a.content)) * fontSize * 0.6
			th := fontSize
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X+(bounds.Width-tw)/2, bounds.Y+(bounds.Height-th)/2, tw, th),
				FillColor: uimath.ColorWhite,
				Corners:   uimath.CornersAll(2),
			}, 1, 1)
		}
	} else {
		// Default: draw a user icon placeholder (simple silhouette)
		a.drawDefaultIcon(buf, bounds, radius)
	}
}

// drawDefaultIcon draws a simple user silhouette placeholder.
func (a *Avatar) drawDefaultIcon(buf *render.CommandBuffer, bounds uimath.Rect, radius float32) {
	cx := bounds.X + bounds.Width/2
	cy := bounds.Y + bounds.Height/2
	px := bounds.Width

	// Head (circle)
	headR := px * 0.18
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(cx-headR, cy-px*0.18-headR, headR*2, headR*2),
		FillColor: uimath.ColorWhite,
		Corners:   uimath.CornersAll(headR),
	}, 1, 1)

	// Body (wider ellipse at bottom)
	bodyW := px * 0.38
	bodyH := px * 0.22
	bodyY := cy + px*0.05
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(cx-bodyW/2, bodyY, bodyW, bodyH),
		FillColor: uimath.ColorWhite,
		Corners:   uimath.CornersAll(bodyW / 2),
	}, 1, 1)
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
	size           Size           // size applied to all child avatars
	cascading      CascadingValue // stacking direction
	collapseAvatar string         // custom text for the "+N" element
}

func NewAvatarGroup(tree *core.Tree, cfg *Config) *AvatarGroup {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &AvatarGroup{
		Base: NewBase(tree, core.TypeCustom, cfg),
		size: SizeMedium,
	}
}

func (g *AvatarGroup) SetMax(m int)                  { g.max = m }
func (g *AvatarGroup) SetSize(s Size)                { g.size = s }
func (g *AvatarGroup) SetCascading(c CascadingValue) { g.cascading = c }
func (g *AvatarGroup) SetCollapseAvatar(text string) { g.collapseAvatar = text }

// Deprecated: Use SetMax instead.
func (g *AvatarGroup) SetMaxCount(m int) { g.max = m }

// Deprecated: spacing is now auto-calculated from size.
func (g *AvatarGroup) SetSpacing(s float32) {}

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
	// Overlap is ~25% of avatar size
	overlap := px * 0.25
	step := px - overlap

	visible := g.avatars
	hasMore := false
	if g.max > 0 && len(visible) > g.max {
		visible = visible[:g.max]
		hasMore = true
	}

	x := bounds.X
	for i, av := range visible {
		avBounds := uimath.NewRect(x, bounds.Y, px, px)
		lo := av.tree.Get(av.id).Layout()
		lo.Bounds = avBounds
		av.tree.SetLayout(av.id, lo)

		// Z-order based on cascading direction
		var z int32
		if g.cascading == CascadingLeftUp {
			z = int32(len(visible) - i)
		} else {
			z = int32(i)
		}

		// Draw border ring (white outline for overlap visibility)
		borderBounds := uimath.NewRect(avBounds.X-1, avBounds.Y-1, px+2, px+2)
		buf.DrawRect(render.RectCmd{
			Bounds:    borderBounds,
			FillColor: uimath.ColorWhite,
			Corners:   uimath.CornersAll((px + 2) / 2),
		}, z, 1)

		av.Draw(buf)
		x += step
	}

	// "+N" collapse indicator
	if hasMore {
		remaining := len(g.avatars) - g.max
		cfg := g.config
		extraBounds := uimath.NewRect(x, bounds.Y, px, px)

		// White border
		borderBounds := uimath.NewRect(extraBounds.X-1, extraBounds.Y-1, px+2, px+2)
		buf.DrawRect(render.RectCmd{
			Bounds:    borderBounds,
			FillColor: uimath.ColorWhite,
			Corners:   uimath.CornersAll((px + 2) / 2),
		}, int32(len(visible)), 1)

		buf.DrawRect(render.RectCmd{
			Bounds:    extraBounds,
			FillColor: uimath.ColorHex("#e7e7e7"),
			Corners:   uimath.CornersAll(px / 2),
		}, int32(len(visible)), 1)

		text := g.collapseAvatar
		if text == "" {
			text = "+" + strconv.Itoa(remaining)
		}
		if cfg.TextRenderer != nil {
			fs := px * 0.38
			lh := cfg.TextRenderer.LineHeight(fs)
			tw := cfg.TextRenderer.MeasureText(text, fs)
			cfg.TextRenderer.DrawText(buf, text,
				extraBounds.X+(px-tw)/2,
				extraBounds.Y+(px-lh)/2,
				fs, px, uimath.ColorHex("#999"), 1)
		}
	}
}

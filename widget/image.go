package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// ImageFit controls how the image fits within its bounds.
type ImageFit uint8

const (
	ImageFitFill    ImageFit = iota // Stretch to fill (default)
	ImageFitContain                 // Scale to fit, preserving aspect ratio
	ImageFitCover                   // Scale to cover, preserving aspect ratio
)

// Image displays a texture.
type Image struct {
	Base
	texture render.TextureHandle
	srcRect uimath.Rect
	fit     ImageFit
	tint    uimath.Color
	src     string // source URL/path for lazy loading
}

// NewImage creates an image widget.
func NewImage(tree *core.Tree, texture render.TextureHandle, cfg *Config) *Image {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	img := &Image{
		Base:    NewBase(tree, core.TypeImage, cfg),
		texture: texture,
		tint:    uimath.ColorWhite,
	}
	img.style.Display = layout.DisplayBlock
	return img
}

func (img *Image) Texture() render.TextureHandle { return img.texture }
func (img *Image) Fit() ImageFit                 { return img.fit }
func (img *Image) Src() string                   { return img.src }

func (img *Image) SetTexture(t render.TextureHandle) { img.texture = t }
func (img *Image) SetSrcRect(r uimath.Rect)          { img.srcRect = r }
func (img *Image) SetFit(f ImageFit)                  { img.fit = f }
func (img *Image) SetTint(c uimath.Color)             { img.tint = c }
func (img *Image) SetSrc(s string)                    { img.src = s }

func (img *Image) Draw(buf *render.CommandBuffer) {
	bounds := img.Bounds()
	if bounds.IsEmpty() || img.texture == render.InvalidTexture {
		return
	}

	dst := bounds
	if img.fit == ImageFitContain || img.fit == ImageFitCover {
		// These require knowing source dimensions; for now just fill
		dst = bounds
	}

	buf.DrawImage(render.ImageCmd{
		Texture: img.texture,
		SrcRect: img.srcRect,
		DstRect: dst,
		Tint:    img.tint,
	}, 0, 1)

	img.DrawChildren(buf)
}

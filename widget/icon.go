package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Icon displays an icon from a texture atlas or glyph.
type Icon struct {
	Base
	name    string
	size    float32
	color   uimath.Color
	texture render.TextureHandle
	uvRect  uimath.Rect
}

// NewIcon creates an icon widget.
func NewIcon(tree *core.Tree, name string, cfg *Config) *Icon {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	ic := &Icon{
		Base:  NewBase(tree, core.TypeImage, cfg),
		name:  name,
		size:  cfg.IconSize,
		color: cfg.TextColor,
	}
	ic.style.Width = layout.Px(cfg.IconSize)
	ic.style.Height = layout.Px(cfg.IconSize)
	return ic
}

func (ic *Icon) Name() string              { return ic.name }
func (ic *Icon) Size() float32             { return ic.size }
func (ic *Icon) Color() uimath.Color       { return ic.color }
func (ic *Icon) Texture() render.TextureHandle { return ic.texture }

func (ic *Icon) SetName(name string) { ic.name = name }

func (ic *Icon) SetSize(size float32) {
	ic.size = size
	ic.style.Width = layout.Px(size)
	ic.style.Height = layout.Px(size)
}

func (ic *Icon) SetColor(c uimath.Color) { ic.color = c }

func (ic *Icon) SetTexture(tex render.TextureHandle, uv uimath.Rect) {
	ic.texture = tex
	ic.uvRect = uv
}

func (ic *Icon) Draw(buf *render.CommandBuffer) {
	bounds := ic.Bounds()
	if bounds.IsEmpty() {
		return
	}

	tex := ic.texture
	uv := ic.uvRect

	// If no texture set, try to resolve from icon registry by name
	if tex == render.InvalidTexture && ic.name != "" && ic.config.IconRegistry != nil {
		size := int(ic.size)
		if size < 1 {
			size = int(ic.config.IconSize)
		}
		if t, ok := ic.config.IconRegistry.Get(ic.name, size); ok {
			tex = t
			uv = uimath.NewRect(0, 0, 1, 1)
		}
	}

	if tex != render.InvalidTexture {
		buf.DrawImage(render.ImageCmd{
			Texture: tex,
			SrcRect: uv,
			DstRect: bounds,
			Tint:    ic.color,
		}, 0, 1)
	}
}

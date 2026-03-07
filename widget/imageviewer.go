package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// ImageViewer displays an image with zoom and pan.
type ImageViewer struct {
	Base
	texture render.TextureHandle
	zoom    float32
	panX    float32
	panY    float32
	visible bool
	onClose func()
}

func NewImageViewer(tree *core.Tree, cfg *Config) *ImageViewer {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	iv := &ImageViewer{
		Base:    NewBase(tree, core.TypeCustom, cfg),
		zoom:    1,
		visible: true,
	}
	tree.AddHandler(iv.id, event.MouseClick, func(e *event.Event) {
		// Double click could reset zoom
	})
	return iv
}

func (iv *ImageViewer) Texture() render.TextureHandle { return iv.texture }
func (iv *ImageViewer) Zoom() float32                 { return iv.zoom }
func (iv *ImageViewer) IsVisible() bool               { return iv.visible }
func (iv *ImageViewer) SetTexture(t render.TextureHandle) { iv.texture = t }
func (iv *ImageViewer) SetZoom(z float32)             { iv.zoom = z }
func (iv *ImageViewer) SetPan(x, y float32)           { iv.panX = x; iv.panY = y }
func (iv *ImageViewer) SetVisible(v bool)             { iv.visible = v }
func (iv *ImageViewer) OnClose(fn func())             { iv.onClose = fn }

func (iv *ImageViewer) ZoomIn()  { iv.zoom *= 1.25 }
func (iv *ImageViewer) ZoomOut() { iv.zoom *= 0.8 }
func (iv *ImageViewer) ResetZoom() { iv.zoom = 1; iv.panX = 0; iv.panY = 0 }

func (iv *ImageViewer) Draw(buf *render.CommandBuffer) {
	if !iv.visible {
		return
	}
	bounds := iv.Bounds()
	if bounds.IsEmpty() {
		return
	}

	// Background
	buf.DrawRect(render.RectCmd{
		Bounds:    bounds,
		FillColor: uimath.RGBA(0, 0, 0, 0.02),
	}, 1, 1)

	// Image (if texture available)
	if iv.texture != 0 {
		imgW := bounds.Width * iv.zoom
		imgH := bounds.Height * iv.zoom
		imgX := bounds.X + (bounds.Width-imgW)/2 + iv.panX
		imgY := bounds.Y + (bounds.Height-imgH)/2 + iv.panY
		buf.DrawImage(render.ImageCmd{
			Texture: iv.texture,
			DstRect: uimath.NewRect(imgX, imgY, imgW, imgH),
		}, 2, 1)
	}
}

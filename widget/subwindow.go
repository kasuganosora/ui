package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// SubWindow is a draggable, floating window within the UI.
type SubWindow struct {
	Base
	title    string
	x, y     float32
	width    float32
	height   float32
	visible  bool
	closable bool
	dragOffX float32
	dragOffY float32
	dragging bool
	onClose  func()
}

func NewSubWindow(tree *core.Tree, title string, cfg *Config) *SubWindow {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	w := &SubWindow{
		Base:     NewBase(tree, core.TypeCustom, cfg),
		title:    title,
		width:    320,
		height:   240,
		visible:  true,
		closable: true,
	}
	tree.AddHandler(w.id, event.MouseDown, func(e *event.Event) {
		headerH := float32(36)
		if e.GlobalY < w.y+headerH {
			w.dragging = true
			w.dragOffX = e.GlobalX - w.x
			w.dragOffY = e.GlobalY - w.y
		}
	})
	tree.AddHandler(w.id, event.MouseMove, func(e *event.Event) {
		if w.dragging {
			w.x = e.GlobalX - w.dragOffX
			w.y = e.GlobalY - w.dragOffY
		}
	})
	tree.AddHandler(w.id, event.MouseUp, func(e *event.Event) { w.dragging = false })
	return w
}

func (w *SubWindow) SetPosition(x, y float32) { w.x = x; w.y = y }
func (w *SubWindow) SetSize(wi, h float32)    { w.width = wi; w.height = h }
func (w *SubWindow) SetTitle(t string)         { w.title = t }
func (w *SubWindow) SetClosable(c bool)        { w.closable = c }
func (w *SubWindow) OnClose(fn func())         { w.onClose = fn }
func (w *SubWindow) IsVisible() bool           { return w.visible }
func (w *SubWindow) Open()                     { w.visible = true; w.tree.MarkDirty(w.id) }
func (w *SubWindow) Close() {
	w.visible = false
	w.tree.MarkDirty(w.id)
	if w.onClose != nil {
		w.onClose()
	}
}

func (w *SubWindow) Draw(buf *render.CommandBuffer) {
	if !w.visible {
		return
	}
	cfg := w.config
	headerH := float32(36)

	// Shadow
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(w.x+2, w.y+2, w.width, w.height),
		FillColor: uimath.RGBA(0, 0, 0, 0.15),
		Corners:   uimath.CornersAll(cfg.BorderRadius),
	}, 30, 1)

	// Window body
	buf.DrawOverlay(render.RectCmd{
		Bounds:      uimath.NewRect(w.x, w.y, w.width, w.height),
		FillColor:   uimath.ColorWhite,
		BorderColor: cfg.BorderColor,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 31, 1)

	// Title bar
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(w.x, w.y, w.width, headerH),
		FillColor: uimath.RGBA(0, 0, 0, 0.02),
		Corners: uimath.Corners{
			TopLeft:  cfg.BorderRadius,
			TopRight: cfg.BorderRadius,
		},
	}, 32, 1)

	// Title bar divider
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(w.x, w.y+headerH-1, w.width, 1),
		FillColor: uimath.RGBA(0, 0, 0, 0.06),
	}, 32, 1)

	// Title text
	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, w.title, w.x+cfg.SpaceMD, w.y+(headerH-lh)/2, cfg.FontSize, w.width-cfg.SpaceMD*2, cfg.TextColor, 1)
	}

	// Close button (X)
	if w.closable {
		cx := w.x + w.width - cfg.SpaceMD - 8
		cy := w.y + headerH/2
		buf.DrawOverlay(render.RectCmd{
			Bounds:    uimath.NewRect(cx-4, cy-0.5, 8, 1),
			FillColor: cfg.TextColor,
		}, 33, 0.5)
		buf.DrawOverlay(render.RectCmd{
			Bounds:    uimath.NewRect(cx-0.5, cy-4, 1, 8),
			FillColor: cfg.TextColor,
		}, 33, 0.5)
	}

	w.DrawChildren(buf)
}

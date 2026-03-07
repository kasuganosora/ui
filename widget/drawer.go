package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// DrawerPlacement controls which edge the drawer slides from.
type DrawerPlacement uint8

const (
	DrawerRight  DrawerPlacement = iota
	DrawerLeft
	DrawerTop
	DrawerBottom
)

// Drawer is a sliding panel from the edge of the screen.
type Drawer struct {
	Base
	title     string
	visible   bool
	width     float32
	height    float32
	placement DrawerPlacement
	closable  bool
	onClose   func()
}

func NewDrawer(tree *core.Tree, title string, cfg *Config) *Drawer {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	d := &Drawer{
		Base:      NewBase(tree, core.TypeCustom, cfg),
		title:     title,
		width:     378,
		height:    378,
		placement: DrawerRight,
		closable:  true,
	}
	tree.AddHandler(d.id, event.MouseClick, func(e *event.Event) {
		// Backdrop click to close
		if d.onClose != nil {
			d.onClose()
		}
	})
	return d
}

func (d *Drawer) SetTitle(t string)              { d.title = t }
func (d *Drawer) SetWidth(w float32)              { d.width = w }
func (d *Drawer) SetHeight(h float32)             { d.height = h }
func (d *Drawer) SetPlacement(p DrawerPlacement)   { d.placement = p }
func (d *Drawer) SetClosable(c bool)              { d.closable = c }
func (d *Drawer) OnClose(fn func())               { d.onClose = fn }
func (d *Drawer) IsVisible() bool                 { return d.visible }

func (d *Drawer) Open()  { d.visible = true; d.tree.MarkDirty(d.id) }
func (d *Drawer) Close() { d.visible = false; d.tree.MarkDirty(d.id) }

func (d *Drawer) Draw(buf *render.CommandBuffer) {
	if !d.visible {
		return
	}
	bounds := d.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := d.config

	// Backdrop
	buf.DrawOverlay(render.RectCmd{
		Bounds:    bounds,
		FillColor: uimath.RGBA(0, 0, 0, 0.45),
	}, 15, 1)

	// Panel
	var px, py, pw, ph float32
	switch d.placement {
	case DrawerRight:
		pw, ph = d.width, bounds.Height
		px, py = bounds.X+bounds.Width-pw, bounds.Y
	case DrawerLeft:
		pw, ph = d.width, bounds.Height
		px, py = bounds.X, bounds.Y
	case DrawerTop:
		pw, ph = bounds.Width, d.height
		px, py = bounds.X, bounds.Y
	case DrawerBottom:
		pw, ph = bounds.Width, d.height
		px, py = bounds.X, bounds.Y+bounds.Height-ph
	}

	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(px, py, pw, ph),
		FillColor: uimath.ColorWhite,
	}, 16, 1)

	// Header
	headerH := float32(48)
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(px, py+headerH-1, pw, 1),
		FillColor: uimath.RGBA(0, 0, 0, 0.06),
	}, 17, 1)

	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSizeLg)
		cfg.TextRenderer.DrawText(buf, d.title, px+cfg.SpaceLG, py+(headerH-lh)/2, cfg.FontSizeLg, pw-cfg.SpaceLG*2, cfg.TextColor, 1)
	}

	d.DrawChildren(buf)
}

package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Dialog is a modal overlay dialog.
type Dialog struct {
	Base
	title   string
	visible bool
	width   float32
	content Widget

	onClose func()
}

// NewDialog creates a dialog.
func NewDialog(tree *core.Tree, title string, cfg *Config) *Dialog {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	d := &Dialog{
		Base:  NewBase(tree, core.TypeCustom, cfg),
		title: title,
		width: 520,
	}
	d.style.Display = layout.DisplayNone

	// Click on backdrop closes dialog
	d.tree.AddHandler(d.id, event.MouseClick, func(e *event.Event) {
		if d.onClose != nil {
			d.onClose()
		}
	})

	return d
}

func (d *Dialog) Title() string   { return d.title }
func (d *Dialog) IsVisible() bool { return d.visible }

func (d *Dialog) SetTitle(title string) { d.title = title }
func (d *Dialog) SetWidth(w float32)    { d.width = w }
func (d *Dialog) SetContent(w Widget)   { d.content = w }
func (d *Dialog) OnClose(fn func())     { d.onClose = fn }

func (d *Dialog) Open() {
	d.visible = true
	d.style.Display = layout.DisplayFlex
}

func (d *Dialog) Close() {
	d.visible = false
	d.style.Display = layout.DisplayNone
}

func (d *Dialog) Draw(buf *render.CommandBuffer) {
	if !d.visible {
		return
	}
	bounds := d.Bounds()
	if bounds.IsEmpty() {
		return
	}

	cfg := d.config

	// Backdrop (semi-transparent overlay)
	buf.DrawRect(render.RectCmd{
		Bounds:    bounds,
		FillColor: uimath.Color{R: 0, G: 0, B: 0, A: 0.45},
	}, 10, 1)

	// Dialog panel centered
	panelW := d.width
	panelH := float32(200)
	if d.content != nil {
		panelH = 300 // allow more space for content
	}
	panelX := bounds.X + (bounds.Width-panelW)/2
	panelY := bounds.Y + (bounds.Height-panelH)/2

	// Panel background
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(panelX, panelY, panelW, panelH),
		FillColor: uimath.ColorWhite,
		Corners:   uimath.CornersAll(cfg.BorderRadius),
	}, 11, 1)

	// Title bar
	titleH := float32(56)
	if d.title != "" {
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeLg)
			tx := panelX + cfg.SpaceLG
			ty := panelY + (titleH-lh)/2
			cfg.TextRenderer.DrawText(buf, d.title, tx, ty, cfg.FontSizeLg, panelW-cfg.SpaceLG*2, cfg.TextColor, 1)
		} else {
			textW := float32(len(d.title)) * cfg.FontSizeLg * 0.55
			textH := cfg.FontSizeLg * 1.2
			tx := panelX + cfg.SpaceLG
			ty := panelY + (titleH-textH)/2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(tx, ty, textW, textH),
				FillColor: cfg.TextColor,
				Corners:   uimath.CornersAll(2),
			}, 12, 1)
		}
	}

	// Close button (X) in top-right
	closeSize := float32(16)
	closeX := panelX + panelW - cfg.SpaceLG - closeSize
	closeY := panelY + (titleH-closeSize)/2
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(closeX, closeY+closeSize/2-1, closeSize, 2),
		FillColor: cfg.TextColor,
	}, 12, 1)
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(closeX+closeSize/2-1, closeY, 2, closeSize),
		FillColor: cfg.TextColor,
	}, 12, 1)

	// Divider
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(panelX, panelY+titleH, panelW, 1),
		FillColor: cfg.BorderColor,
	}, 12, 1)

	// Content area
	if d.content != nil {
		d.content.Draw(buf)
	}

	d.DrawChildren(buf)
}

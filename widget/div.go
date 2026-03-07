package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Div is a container widget that arranges children via flexbox.
type Div struct {
	Base
	bgColor      uimath.Color
	borderColor  uimath.Color
	borderWidth  float32
	borderRadius float32
	scrollX      float32
	scrollY      float32
	scrollable   bool
}

// NewDiv creates a div container.
func NewDiv(tree *core.Tree, cfg *Config) *Div {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	d := &Div{
		Base: NewBase(tree, core.TypeDiv, cfg),
	}
	d.style.Display = layout.DisplayFlex
	d.style.FlexDirection = layout.FlexDirectionColumn
	return d
}

func (d *Div) BgColor() uimath.Color   { return d.bgColor }
func (d *Div) BorderColor() uimath.Color { return d.borderColor }
func (d *Div) BorderWidth() float32     { return d.borderWidth }
func (d *Div) BorderRadius() float32    { return d.borderRadius }
func (d *Div) ScrollX() float32         { return d.scrollX }
func (d *Div) ScrollY() float32         { return d.scrollY }
func (d *Div) IsScrollable() bool       { return d.scrollable }

func (d *Div) SetBgColor(c uimath.Color)       { d.bgColor = c }
func (d *Div) SetBorderColor(c uimath.Color)    { d.borderColor = c }
func (d *Div) SetBorderWidth(w float32)          { d.borderWidth = w }
func (d *Div) SetBorderRadius(r float32)         { d.borderRadius = r }
func (d *Div) SetScrollable(v bool)              { d.scrollable = v }

func (d *Div) ScrollTo(x, y float32) {
	d.scrollX = x
	d.scrollY = y
}

func (d *Div) Draw(buf *render.CommandBuffer) {
	bounds := d.Bounds()
	if bounds.IsEmpty() {
		return
	}

	// Draw background
	if !d.bgColor.IsTransparent() || d.borderWidth > 0 {
		buf.DrawRect(render.RectCmd{
			Bounds:      bounds,
			FillColor:   d.bgColor,
			BorderColor: d.borderColor,
			BorderWidth: d.borderWidth,
			Corners:     uimath.CornersAll(d.borderRadius),
		}, 0, 1)
	}

	// Clip if scrollable
	if d.scrollable {
		buf.PushClip(bounds)
	}

	d.DrawChildren(buf)

	if d.scrollable {
		buf.PopClip()
	}
}

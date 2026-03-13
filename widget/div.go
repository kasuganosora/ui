package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/css"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Div is a container widget that arranges children via flexbox.
type Div struct {
	Base
	bgColor       uimath.Color
	borderColor   uimath.Color
	borderWidth   float32
	borderRadius  float32
	scrollX       float32
	scrollY       float32
	contentHeight float32
	scrollable    bool
	gradientStart uimath.Color
	gradientEnd   uimath.Color
	gradientAngle float32 // radians; 0 = no gradient
	shadows       []css.BoxShadowLayer
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

func (d *Div) GradientStart() uimath.Color      { return d.gradientStart }
func (d *Div) GradientEnd() uimath.Color        { return d.gradientEnd }
func (d *Div) GradientAngle() float32           { return d.gradientAngle }

func (d *Div) SetGradient(start, end uimath.Color, angle float32) {
	d.gradientStart = start
	d.gradientEnd = end
	d.gradientAngle = angle
}

func (d *Div) SetBoxShadow(layers []css.BoxShadowLayer) {
	d.shadows = layers
}

func (d *Div) ContentHeight() float32          { return d.contentHeight }
func (d *Div) SetContentHeight(h float32)       { d.contentHeight = h }

func (d *Div) ScrollTo(x, y float32) {
	d.scrollX = x
	d.scrollY = y
}

func (d *Div) Draw(buf *render.CommandBuffer) {
	bounds := d.Bounds()
	if bounds.IsEmpty() {
		return
	}

	// Draw box shadows (behind the background), back-to-front (last layer first)
	for i := len(d.shadows) - 1; i >= 0; i-- {
		sh := d.shadows[i]
		if sh.Inset {
			continue // inset shadows drawn after background; skip here
		}
		buf.DrawShadow(render.ShadowCmd{
			Bounds:       bounds,
			Corners:      uimath.CornersAll(d.borderRadius),
			OffsetX:      sh.OffsetX,
			OffsetY:      sh.OffsetY,
			BlurRadius:   sh.Blur,
			SpreadRadius: sh.Spread,
			Color:        sh.Color,
		}, -1, 1)
	}

	// Draw background
	if !d.bgColor.IsTransparent() || d.borderWidth > 0 || d.gradientAngle != 0 {
		buf.DrawRect(render.RectCmd{
			Bounds:        bounds,
			FillColor:     d.bgColor,
			BorderColor:   d.borderColor,
			BorderWidth:   d.borderWidth,
			Corners:       uimath.CornersAll(d.borderRadius),
			GradientStart: d.gradientStart,
			GradientEnd:   d.gradientEnd,
			GradientAngle: d.gradientAngle,
		}, 0, 1)
	}

	// Clip if scrollable or overflow is hidden/scroll/auto
	needsClip := d.scrollable ||
		d.style.Overflow == layout.OverflowHidden ||
		d.style.Overflow == layout.OverflowScroll ||
		d.style.Overflow == layout.OverflowAuto
	if needsClip {
		buf.PushClip(bounds)
	}

	// Apply scroll offset: temporarily move all descendant elements
	if d.scrollX != 0 || d.scrollY != 0 {
		d.offsetDescendants(-d.scrollX, -d.scrollY)
	}

	d.DrawChildren(buf)

	// Restore scroll offset
	if d.scrollX != 0 || d.scrollY != 0 {
		d.offsetDescendants(d.scrollX, d.scrollY)
	}

	if needsClip {
		buf.PopClip()
	}

	// Draw scrollbar if content overflows
	hasScroll := d.scrollable ||
		d.style.Overflow == layout.OverflowScroll ||
		d.style.Overflow == layout.OverflowAuto
	if hasScroll && d.contentHeight > bounds.Height {
		const barW = 4
		trackX := bounds.X + bounds.Width - barW - 1
		ratio := bounds.Height / d.contentHeight
		thumbH := bounds.Height * ratio
		if thumbH < 20 {
			thumbH = 20
		}
		maxScroll := d.contentHeight - bounds.Height
		thumbY := bounds.Y
		if maxScroll > 0 {
			thumbY += (bounds.Height - thumbH) * (d.scrollY / maxScroll)
		}
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(trackX, thumbY, barW, thumbH),
			FillColor: uimath.ColorHex("#ffffff40"),
			Corners:   uimath.CornersAll(barW / 2),
		}, 0, 1)
	}
}

// offsetDescendants moves all descendant element layout bounds by (dx, dy).
// Used to apply/restore scroll offset during drawing.
func (d *Div) offsetDescendants(dx, dy float32) {
	for _, child := range d.children {
		d.offsetElement(child.ElementID(), dx, dy)
	}
}

func (d *Div) offsetElement(id core.ElementID, dx, dy float32) {
	elem := d.tree.Get(id)
	if elem == nil {
		return
	}
	b := elem.Layout().Bounds
	d.tree.SetLayout(id, core.LayoutResult{
		Bounds: uimath.NewRect(b.X+dx, b.Y+dy, b.Width, b.Height),
	})
	for _, child := range elem.ChildIDs() {
		d.offsetElement(child, dx, dy)
	}
}

package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Camera is a pan and zoom plane. Marker positions are in world units.
// Dragging slides the viewport across the world. The wheel zooms around the cursor.
type Camera struct {
	Base
	originX float32
	originY float32
	zoom    float32
	drag    bool
	lastX   float32
	lastY   float32
	ids     []string
	wx      []float32
	wy      []float32
}

func NewCamera(tree *core.Tree, cfg *Config) *Camera {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	c := &Camera{Base: NewBase(tree, core.TypeCustom, cfg), zoom: 1}
	tree.AddHandler(c.id, event.MouseDown, func(e *event.Event) {
		c.drag = true
		c.lastX, c.lastY = e.GlobalX, e.GlobalY
	})
	tree.AddHandler(c.id, event.MouseMove, func(e *event.Event) {
		if !c.drag {
			return
		}
		dx, dy := e.GlobalX-c.lastX, e.GlobalY-c.lastY
		c.lastX, c.lastY = e.GlobalX, e.GlobalY
		if c.zoom == 0 {
			return
		}
		c.originX += dx / c.zoom
		c.originY += dy / c.zoom
	})
	tree.AddHandler(c.id, event.MouseUp, func(*event.Event) { c.drag = false })
	tree.AddHandler(c.id, event.MouseWheel, func(e *event.Event) {
		factor := float32(1.1)
		if e.WheelDY > 0 {
			factor = 1 / 1.1
		}
		wx, wy := c.ScreenToWorld(e.GlobalX, e.GlobalY)
		c.zoom *= factor
		if c.zoom < 0.2 {
			c.zoom = 0.2
		}
		if c.zoom > 8 {
			c.zoom = 8
		}
		b := c.Bounds()
		c.originX = wx - (e.GlobalX-b.X)/c.zoom
		c.originY = wy - (e.GlobalY-b.Y)/c.zoom
	})
	return c
}

func (c *Camera) OriginX() float32 { return c.originX }
func (c *Camera) OriginY() float32 { return c.originY }
func (c *Camera) Zoom() float32    { return c.zoom }
func (c *Camera) SetZoom(z float32) {
	if z < 0.2 {
		z = 0.2
	}
	if z > 8 {
		z = 8
	}
	c.zoom = z
}

func (c *Camera) AddMarker(id string, x, y float32) {
	c.ids = append(c.ids, id)
	c.wx = append(c.wx, x)
	c.wy = append(c.wy, y)
}

func (c *Camera) WorldToScreen(x, y float32) (float32, float32) {
	b := c.Bounds()
	return b.X + (x-c.originX)*c.zoom, b.Y + (y-c.originY)*c.zoom
}

func (c *Camera) ScreenToWorld(x, y float32) (float32, float32) {
	b := c.Bounds()
	if c.zoom == 0 {
		return 0, 0
	}
	return c.originX + (x-b.X)/c.zoom, c.originY + (y-b.Y)/c.zoom
}

func (c *Camera) MarkerAt(x, y float32) string {
	const r = float32(8)
	for i := len(c.ids) - 1; i >= 0; i-- {
		sx, sy := c.WorldToScreen(c.wx[i], c.wy[i])
		if x >= sx-r && x <= sx+r && y >= sy-r && y <= sy+r {
			return c.ids[i]
		}
	}
	return ""
}

func (c *Camera) Draw(buf *render.CommandBuffer) {
	b := c.Bounds()
	if b.IsEmpty() {
		return
	}
	buf.DrawRect(render.RectCmd{Bounds: b, FillColor: c.config.BgColor, BorderColor: c.config.BorderColor, BorderWidth: 1}, 1, 1)
	step := float32(40) * c.zoom
	if step < 8 {
		step = 8
	}
	for x := b.X; x < b.X+b.Width; x += step {
		buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(x, b.Y, 1, b.Height), FillColor: c.config.BorderColor}, 2, 1)
	}
	for i := range c.ids {
		sx, sy := c.WorldToScreen(c.wx[i], c.wy[i])
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(sx-5, sy-5, 10, 10),
			FillColor: c.config.PrimaryColor,
		}, 3, 1)
	}
}

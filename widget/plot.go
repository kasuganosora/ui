package widget

import (
	"math"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// PlotPoint is one cartesian sample.
type PlotPoint struct {
	X, Y float32
}

// TreemapItem is one value in a slice-and-dice treemap.
type TreemapItem struct {
	Name  string
	Value float32
}

// SankeyNode is a column node. Value is its throughput.
type SankeyNode struct {
	Name  string
	Value float32
}

// SankeyLink connects two nodes by index.
type SankeyLink struct {
	From, To int
	Value    float32
}

type plotMode uint8

const (
	plotBars plotMode = iota
	plotLine
	plotArea
	plotScatter
	plotPie
	plotTreemap
	plotSankey
)

// Plot draws charts into a Canvas and into retained rectangles.
// Sankey and treemap share this type; they do not depend on d3.
type Plot struct {
	Base
	canvas   *Canvas
	mode     plotMode
	bars     []float32
	points   []PlotPoint
	pie      []float32
	tiles    []TreemapItem
	nodes    []SankeyNode
	links    []SankeyLink
	selected int
	rects    []uimath.Rect
}

func NewPlot(tree *core.Tree, cfg *Config) *Plot {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	p := &Plot{
		Base:     NewBase(tree, core.TypeCustom, cfg),
		canvas:   NewCanvas(tree, 320, 200, nil, cfg),
		selected: -1,
	}
	p.AppendChild(p.canvas)
	tree.SetHitTestFunc(p.canvas.ElementID(), func(float32, float32) bool { return false })
	tree.AddHandler(p.id, event.MouseClick, func(e *event.Event) {
		if !targetPhase(e) {
			return
		}
		_, idx, ok := p.Hit(e.GlobalX, e.GlobalY)
		if ok {
			p.selected = idx
		}
	})
	return p
}

func (p *Plot) Canvas() *Canvas { return p.canvas }
func (p *Plot) Selected() int   { return p.selected }
func (p *Plot) SetBars(v []float32) {
	p.mode = plotBars
	p.bars = append(p.bars[:0], v...)
	p.rebuild()
}
func (p *Plot) SetLine(pts []PlotPoint)    { p.setPoints(plotLine, pts) }
func (p *Plot) SetArea(pts []PlotPoint)    { p.setPoints(plotArea, pts) }
func (p *Plot) SetScatter(pts []PlotPoint) { p.setPoints(plotScatter, pts) }
func (p *Plot) setPoints(mode plotMode, pts []PlotPoint) {
	p.mode = mode
	p.points = append(p.points[:0], pts...)
	p.rebuild()
}
func (p *Plot) SetPie(v []float32) {
	p.mode = plotPie
	p.pie = append(p.pie[:0], v...)
	p.rebuild()
}
func (p *Plot) SetTreemap(items []TreemapItem) {
	p.mode = plotTreemap
	p.tiles = append(p.tiles[:0], items...)
	p.rebuild()
}
func (p *Plot) SetSankey(nodes []SankeyNode, links []SankeyLink) {
	p.mode = plotSankey
	p.nodes = append(p.nodes[:0], nodes...)
	p.links = append(p.links[:0], links...)
	p.rebuild()
}

func (p *Plot) plotBounds() uimath.Rect {
	b := p.Bounds()
	if b.IsEmpty() {
		b = uimath.NewRect(0, 0, float32(p.canvas.Width()), float32(p.canvas.Height()))
	}
	return uimath.NewRect(b.X+36, b.Y+12, b.Width-48, b.Height-36)
}

func (p *Plot) domain() (minX, maxX, minY, maxY float32) {
	minX, maxX, minY, maxY = 0, 1, 0, 1
	switch p.mode {
	case plotBars:
		maxX = float32(len(p.bars))
		for _, v := range p.bars {
			if v > maxY {
				maxY = v
			}
		}
	default:
		if len(p.points) == 0 {
			return
		}
		minX, maxX = p.points[0].X, p.points[0].X
		minY, maxY = p.points[0].Y, p.points[0].Y
		for _, pt := range p.points {
			if pt.X < minX {
				minX = pt.X
			}
			if pt.X > maxX {
				maxX = pt.X
			}
			if pt.Y < minY {
				minY = pt.Y
			}
			if pt.Y > maxY {
				maxY = pt.Y
			}
		}
		if minX == maxX {
			maxX = minX + 1
		}
		if minY == maxY {
			maxY = minY + 1
		}
	}
	return
}

func (p *Plot) mapPoint(x, y float32) (float32, float32) {
	minX, maxX, minY, maxY := p.domain()
	a := p.plotBounds()
	sx := a.X + (x-minX)/(maxX-minX)*a.Width
	sy := a.Y + a.Height - (y-minY)/(maxY-minY)*a.Height
	return sx, sy
}

func (p *Plot) rebuild() {
	p.rects = p.rects[:0]
	switch p.mode {
	case plotBars:
		a := p.plotBounds()
		n := len(p.bars)
		if n == 0 || a.Width <= 0 {
			return
		}
		slot := a.Width / float32(n)
		_, _, _, maxY := p.domain()
		for i, v := range p.bars {
			h := float32(0)
			if maxY > 0 {
				h = v / maxY * a.Height
			}
			bw := slot * 0.7
			x := a.X + float32(i)*slot + (slot-bw)/2
			p.rects = append(p.rects, uimath.NewRect(x, a.Y+a.Height-h, bw, h))
		}
	case plotLine, plotArea, plotScatter:
		for _, pt := range p.points {
			x, y := p.mapPoint(pt.X, pt.Y)
			p.rects = append(p.rects, uimath.NewRect(x-4, y-4, 8, 8))
		}
	case plotTreemap:
		a := p.plotBounds()
		var sum float32
		for _, item := range p.tiles {
			if item.Value > 0 {
				sum += item.Value
			}
		}
		x := a.X
		if sum <= 0 {
			return
		}
		for _, item := range p.tiles {
			w := a.Width * item.Value / sum
			p.rects = append(p.rects, uimath.NewRect(x, a.Y, w, a.Height))
			x += w
		}
	case plotSankey:
		a := p.plotBounds()
		if len(p.nodes) == 0 {
			return
		}
		colW := a.Width / float32(len(p.nodes)+1)
		var maxV float32
		for _, n := range p.nodes {
			if n.Value > maxV {
				maxV = n.Value
			}
		}
		if maxV <= 0 {
			maxV = 1
		}
		for i, n := range p.nodes {
			h := n.Value / maxV * a.Height
			x := a.X + float32(i+1)*colW - 16
			y := a.Y + (a.Height-h)/2
			p.rects = append(p.rects, uimath.NewRect(x, y, 32, h))
		}
	}
}

func (p *Plot) BarAnchor(i int) (float32, float32) {
	p.rebuild()
	if i < 0 || i >= len(p.rects) {
		return 0, 0
	}
	r := p.rects[i]
	return r.X + r.Width/2, r.Y + r.Height/2
}

func (p *Plot) LineAnchor(i int) (float32, float32) { return p.BarAnchor(i) }
func (p *Plot) TreemapRect(i int) uimath.Rect {
	p.rebuild()
	if i < 0 || i >= len(p.rects) {
		return uimath.Rect{}
	}
	return p.rects[i]
}
func (p *Plot) SankeyNode(i int) uimath.Rect { return p.TreemapRect(i) }

func (p *Plot) PieAnchor(i int) (float32, float32) {
	a := p.plotBounds()
	cx := a.X + a.Width/2
	cy := a.Y + a.Height/2
	r := a.Height * 0.3
	if a.Width < a.Height {
		r = a.Width * 0.3
	}
	var sum float32
	for _, v := range p.pie {
		sum += v
	}
	if sum <= 0 || i < 0 || i >= len(p.pie) {
		return cx, cy
	}
	angle := float32(-math.Pi / 2)
	for n, v := range p.pie {
		sweep := v / sum * float32(2*math.Pi)
		if n == i {
			mid := angle + sweep/2
			return cx + float32(math.Cos(float64(mid)))*r, cy + float32(math.Sin(float64(mid)))*r
		}
		angle += sweep
	}
	return cx, cy
}

func (p *Plot) Hit(x, y float32) (series, index int, ok bool) {
	if p.mode == plotPie {
		a := p.plotBounds()
		cx := a.X + a.Width/2
		cy := a.Y + a.Height/2
		ang := math.Atan2(float64(y-cy), float64(x-cx))
		// Convert atan2 (-pi, pi] so it starts at -pi/2 and increases clockwise in screen space.
		// Screen y grows downward, which matches atan2's y.
		start := -math.Pi / 2
		rel := ang - start
		for rel < 0 {
			rel += 2 * math.Pi
		}
		var sum float32
		for _, v := range p.pie {
			sum += v
		}
		if sum <= 0 {
			return 0, 0, false
		}
		acc := 0.0
		for i, v := range p.pie {
			acc += float64(v / sum * float32(2*math.Pi))
			if rel <= acc {
				return 0, i, true
			}
		}
		return 0, len(p.pie) - 1, true
	}
	p.rebuild()
	for i := len(p.rects) - 1; i >= 0; i-- {
		if p.rects[i].Contains(uimath.Vec2{X: x, Y: y}) {
			return 0, i, true
		}
	}
	return 0, 0, false
}

func (p *Plot) Draw(buf *render.CommandBuffer) {
	b := p.Bounds()
	if b.Width >= 2 && b.Height >= 2 {
		p.canvas.SetSize(int(b.Width), int(b.Height))
		p.tree.SetLayout(p.canvas.ElementID(), core.LayoutResult{Bounds: b})
	}
	p.rebuild()
	ctx := p.canvas.GetContext2D()
	ctx.SetFillColor(uimath.Color{R: 1, G: 1, B: 1, A: 1})
	ctx.FillRect(0, 0, float64(p.canvas.Width()), float64(p.canvas.Height()))
	ctx.SetFillColor(p.config.PrimaryColor)
	for _, r := range p.rects {
		ctx.FillRect(float64(r.X-b.X), float64(r.Y-b.Y), float64(r.Width), float64(r.Height))
		buf.DrawRect(render.RectCmd{Bounds: r, FillColor: p.config.PrimaryColor}, 2, 1)
	}
	if p.mode == plotPie {
		p.paintPie(ctx, buf, b)
	}
	axis := p.plotBounds()
	buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(axis.X, axis.Y, 1, axis.Height), FillColor: p.config.TextColor}, 3, 1)
	buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(axis.X, axis.Y+axis.Height, axis.Width, 1), FillColor: p.config.TextColor}, 3, 1)
	p.canvas.Draw(buf)
}

func (p *Plot) paintPie(ctx *CanvasContext2D, buf *render.CommandBuffer, b uimath.Rect) {
	a := p.plotBounds()
	cx := a.X + a.Width/2
	cy := a.Y + a.Height/2
	radius := a.Height * 0.35
	if a.Width < a.Height {
		radius = a.Width * 0.35
	}
	var sum float32
	for _, v := range p.pie {
		sum += v
	}
	if sum <= 0 || radius <= 0 {
		return
	}
	colors := []uimath.Color{p.config.PrimaryColor, p.config.ErrorColor, p.config.SuccessColor, p.config.WarningColor}
	for py := int(cy - radius); py <= int(cy+radius); py++ {
		for px := int(cx - radius); px <= int(cx+radius); px++ {
			dx, dy := float32(px)-cx, float32(py)-cy
			if dx*dx+dy*dy > radius*radius {
				continue
			}
			ang := math.Atan2(float64(dy), float64(dx))
			rel := ang + math.Pi/2
			for rel < 0 {
				rel += 2 * math.Pi
			}
			acc := 0.0
			idx := 0
			for i, v := range p.pie {
				acc += float64(v / sum * float32(2*math.Pi))
				idx = i
				if rel <= acc {
					break
				}
			}
			col := colors[idx%len(colors)]
			ctx.SetFillColor(col)
			ctx.FillRect(float64(px)-float64(b.X), float64(py)-float64(b.Y), 1, 1)
		}
	}
	// One retained rectangle per slice so the software raster shows both colors.
	angle := -math.Pi / 2
	for i, v := range p.pie {
		sweep := float64(v / sum * float32(2*math.Pi))
		mid := angle + sweep/2
		x := cx + float32(math.Cos(mid))*radius*0.6
		y := cy + float32(math.Sin(mid))*radius*0.6
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x-6, y-6, 12, 12),
			FillColor: colors[i%len(colors)],
		}, 4, 1)
		angle += sweep
	}
}

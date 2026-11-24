package widget

import (
	"strings"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Diagram lays out a small box-and-arrow description. Lines are "box ID",
// "box ID label", or "A -> B". Subgraphs, styles, and other Mermaid syntax
// are ignored. This is not a Mermaid implementation and does not use limu.
type Diagram struct {
	Base
	ids     []string
	labels  map[string]string
	edges   [][2]string
	x, y    map[string]float32
	w, h    map[string]float32
	index   int
	scrollX float32
	scrollY float32
}

// NewDiagram creates an empty diagram.
func NewDiagram(tree *core.Tree, cfg *Config) *Diagram {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	d := &Diagram{
		Base:   NewBase(tree, core.TypeCustom, cfg),
		labels: map[string]string{},
		x:      map[string]float32{},
		y:      map[string]float32{},
		w:      map[string]float32{},
		h:      map[string]float32{},
	}
	tree.SetFocusable(d.id, true)
	tree.AddHandler(d.id, event.KeyDown, func(e *event.Event) {
		if !targetPhase(e) || len(d.ids) == 0 {
			return
		}
		switch e.Key {
		case event.KeyArrowLeft, event.KeyArrowUp:
			d.move(-1)
		case event.KeyArrowRight, event.KeyArrowDown:
			d.move(1)
		default:
			return
		}
		e.StopPropagation()
	})
	tree.AddHandler(d.id, event.MouseClick, func(e *event.Event) {
		if !targetPhase(e) {
			return
		}
		if id := d.hit(e.GlobalX, e.GlobalY); id != "" {
			for i, cur := range d.ids {
				if cur == id {
					d.index = i
				}
			}
		}
	})
	return d
}

// SetSource replaces the diagram from the small box-and-arrow subset.
func (d *Diagram) SetSource(src string) {
	d.ids = nil
	d.labels = map[string]string{}
	d.edges = nil
	d.index = 0
	d.scrollX, d.scrollY = 0, 0
	for _, line := range strings.Split(src, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "subgraph") || strings.HasPrefix(line, "end") {
			continue
		}
		if strings.HasPrefix(line, "box ") {
			rest := strings.TrimSpace(line[len("box "):])
			id, label := rest, rest
			if i := strings.IndexByte(rest, ' '); i >= 0 {
				id = rest[:i]
				label = strings.TrimSpace(rest[i+1:])
				if label == "" {
					label = id
				}
			}
			if _, ok := d.labels[id]; !ok {
				d.ids = append(d.ids, id)
			}
			d.labels[id] = label
			continue
		}
		parts := strings.Split(line, "->")
		if len(parts) != 2 {
			continue
		}
		a := strings.TrimSpace(parts[0])
		b := strings.TrimSpace(parts[1])
		if a == "" || b == "" {
			continue
		}
		d.edges = append(d.edges, [2]string{a, b})
		for _, id := range []string{a, b} {
			if _, ok := d.labels[id]; !ok {
				d.ids = append(d.ids, id)
				d.labels[id] = id
			}
		}
	}
	d.layout()
}

func (d *Diagram) NodeCount() int { return len(d.ids) }
func (d *Diagram) EdgeCount() int { return len(d.edges) }

func (d *Diagram) Selected() string {
	if d.index < 0 || d.index >= len(d.ids) {
		return ""
	}
	return d.ids[d.index]
}

func (d *Diagram) SetScrollX(x float32) {
	if x < 0 {
		x = 0
	}
	d.scrollX = x
}
func (d *Diagram) SetScrollY(y float32) {
	if y < 0 {
		y = 0
	}
	d.scrollY = y
}

// NodeScreen is the node rectangle in screen space, including scroll.
func (d *Diagram) NodeScreen(id string) uimath.Rect {
	b := d.Bounds()
	if _, ok := d.labels[id]; !ok {
		return uimath.Rect{}
	}
	return uimath.NewRect(b.X+d.x[id]-d.scrollX, b.Y+d.y[id]-d.scrollY, d.w[id], d.h[id])
}

func (d *Diagram) move(dir int) {
	d.index += dir
	if d.index < 0 {
		d.index = 0
	}
	if d.index >= len(d.ids) {
		d.index = len(d.ids) - 1
	}
	s := d.NodeScreen(d.ids[d.index])
	b := d.Bounds()
	if s.X < b.X {
		d.scrollX -= b.X - s.X
	}
	if s.X+s.Width > b.X+b.Width {
		d.scrollX += s.X + s.Width - (b.X + b.Width)
	}
	if d.scrollX < 0 {
		d.scrollX = 0
	}
}

func (d *Diagram) layout() {
	d.x = map[string]float32{}
	d.y = map[string]float32{}
	d.w = map[string]float32{}
	d.h = map[string]float32{}
	indeg := map[string]int{}
	for _, id := range d.ids {
		indeg[id] = 0
	}
	for _, e := range d.edges {
		if e[0] != e[1] {
			indeg[e[1]]++
		}
	}
	layer := map[string]int{}
	var q []string
	for _, id := range d.ids {
		if indeg[id] == 0 {
			q = append(q, id)
			layer[id] = 0
		}
	}
	seen := map[string]bool{}
	for len(q) > 0 {
		id := q[0]
		q = q[1:]
		if seen[id] {
			continue
		}
		seen[id] = true
		for _, e := range d.edges {
			if e[0] != id || e[0] == e[1] {
				continue
			}
			next := layer[id] + 1
			if have, ok := layer[e[1]]; !ok || next > have {
				layer[e[1]] = next
			}
			indeg[e[1]]--
			if indeg[e[1]] <= 0 {
				q = append(q, e[1])
			}
		}
	}
	for _, id := range d.ids {
		if _, ok := layer[id]; !ok {
			layer[id] = 0
		}
	}
	slot := map[int]int{}
	const gapX, gapY = float32(24), float32(16)
	const minW, nodeH = float32(64), float32(32)
	colW := map[int]float32{}
	for _, id := range d.ids {
		w := d.measure(d.labels[id]) + 20
		if w < minW {
			w = minW
		}
		d.w[id] = w
		d.h[id] = nodeH
		if w > colW[layer[id]] {
			colW[layer[id]] = w
		}
	}
	colX := map[int]float32{}
	maxLayer := 0
	for _, id := range d.ids {
		if layer[id] > maxLayer {
			maxLayer = layer[id]
		}
	}
	var x float32
	for i := 0; i <= maxLayer; i++ {
		colX[i] = x
		w := colW[i]
		if w == 0 {
			w = minW
		}
		x += w + gapX
	}
	for _, id := range d.ids {
		l := layer[id]
		d.x[id] = colX[l]
		d.y[id] = float32(slot[l]) * (nodeH + gapY)
		slot[l]++
	}
}

func (d *Diagram) measure(text string) float32 {
	size := d.config.FontSizeSm
	if size <= 0 {
		size = d.config.FontSize
	}
	if d.config.TextRenderer != nil {
		return d.config.TextRenderer.MeasureText(text, size)
	}
	return float32(len([]rune(text))) * size * 0.55
}

func (d *Diagram) hit(x, y float32) string {
	b := d.Bounds()
	if !b.Contains(uimath.Vec2{X: x, Y: y}) {
		return ""
	}
	var found string
	for _, id := range d.ids {
		if d.NodeScreen(id).Contains(uimath.Vec2{X: x, Y: y}) {
			found = id
		}
	}
	return found
}

func (d *Diagram) Draw(buf *render.CommandBuffer) {
	b := d.Bounds()
	if b.IsEmpty() || len(d.ids) == 0 {
		return
	}
	cfg := d.config
	buf.PushClip(b)
	for _, e := range d.edges {
		a := d.NodeScreen(e[0])
		c := d.NodeScreen(e[1])
		x0 := a.X + a.Width
		y0 := a.Y + a.Height/2
		x1 := c.X
		y1 := c.Y + c.Height/2
		buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(min32(x0, x1), y0, abs32(x1-x0)+2, 2), FillColor: cfg.BorderColor}, 2, 1)
		buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(x1, min32(y0, y1), 2, abs32(y1-y0)+2), FillColor: cfg.BorderColor}, 2, 1)
	}
	size := cfg.FontSizeSm
	for _, id := range d.ids {
		s := d.NodeScreen(id)
		if s.X+s.Width < b.X || s.X > b.X+b.Width || s.Y+s.Height < b.Y || s.Y > b.Y+b.Height {
			continue
		}
		fill := uimath.ColorWhite
		if id == d.Selected() {
			fill = cfg.HoverColor
		}
		buf.PushClip(s.Intersection(b))
		buf.DrawRect(render.RectCmd{Bounds: s, FillColor: fill, BorderColor: cfg.BorderColor, BorderWidth: 1}, 1, 1)
		emitText(cfg, buf, d.labels[id], s.X+8, s.Y+6, size, s.Width-16, cfg.TextColor, 1, 4)
		buf.PopClip()
	}
	buf.PopClip()
}

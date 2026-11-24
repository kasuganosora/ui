package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

const (
	flowNodeW = 100
	flowNodeH = 36
	flowGapX  = 36
	flowGapY  = 20
)

// Flow is a layered flowchart drawn on a Camera. Layers follow the longest
// path from sources. A self-loop does not change a node's layer. Nodes that
// only participate in a cycle share one extra layer; this is not a full
// cycle layout.
type Flow struct {
	Base
	cam      *Camera
	ids      []string
	labels   map[string]string
	edges    [][2]string
	layer    map[string]int
	slot     map[string]int
	selected string
	panning  bool
	lastX    float32
	lastY    float32
}

// NewFlow creates an empty diagram and the camera it pans.
func NewFlow(tree *core.Tree, cfg *Config) *Flow {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	f := &Flow{
		Base:   NewBase(tree, core.TypeCustom, cfg),
		cam:    NewCamera(tree, cfg),
		labels: map[string]string{},
		layer:  map[string]int{},
		slot:   map[string]int{},
	}
	tree.AddHandler(f.id, event.MouseDown, func(e *event.Event) {
		if e.Phase == event.PhaseCapture {
			return
		}
		f.sync()
		if id := f.NodeAt(e.GlobalX, e.GlobalY); id != "" {
			f.selected = id
			f.panning = false
			return
		}
		f.panning = true
		f.lastX, f.lastY = e.GlobalX, e.GlobalY
	})
	tree.AddHandler(f.id, event.MouseMove, func(e *event.Event) {
		if !f.panning || f.cam.Zoom() == 0 {
			return
		}
		dx, dy := e.GlobalX-f.lastX, e.GlobalY-f.lastY
		f.lastX, f.lastY = e.GlobalX, e.GlobalY
		f.cam.SetOrigin(f.cam.OriginX()+dx/f.cam.Zoom(), f.cam.OriginY()+dy/f.cam.Zoom())
	})
	tree.AddHandler(f.id, event.MouseUp, func(*event.Event) { f.panning = false })
	tree.AddHandler(f.id, event.MouseWheel, func(e *event.Event) {
		f.sync()
		factor := float32(1.1)
		if e.WheelDY > 0 {
			factor = 1 / 1.1
		}
		wx, wy := f.cam.ScreenToWorld(e.GlobalX, e.GlobalY)
		f.cam.SetZoom(f.cam.Zoom() * factor)
		b := f.cam.Bounds()
		if f.cam.Zoom() == 0 {
			return
		}
		f.cam.SetOrigin(wx-(e.GlobalX-b.X)/f.cam.Zoom(), wy-(e.GlobalY-b.Y)/f.cam.Zoom())
	})
	return f
}

// Camera is the plane this diagram pans and zooms.
func (f *Flow) Camera() *Camera { return f.cam }

func (f *Flow) AddNode(id, label string) {
	if id == "" {
		return
	}
	if _, ok := f.labels[id]; ok {
		f.labels[id] = label
		return
	}
	f.ids = append(f.ids, id)
	f.labels[id] = label
}

func (f *Flow) AddEdge(from, to string) {
	if from == "" || to == "" {
		return
	}
	f.edges = append(f.edges, [2]string{from, to})
}

func (f *Flow) HasEdge(from, to string) bool {
	for _, e := range f.edges {
		if e[0] == from && e[1] == to {
			return true
		}
	}
	return false
}

// RemoveNode deletes the node and every edge that touches it.
func (f *Flow) RemoveNode(id string) {
	kept := f.ids[:0]
	for _, cur := range f.ids {
		if cur != id {
			kept = append(kept, cur)
		}
	}
	f.ids = kept
	delete(f.labels, id)
	next := f.edges[:0]
	for _, e := range f.edges {
		if e[0] != id && e[1] != id {
			next = append(next, e)
		}
	}
	f.edges = next
	if f.selected == id {
		f.selected = ""
	}
}

func (f *Flow) Selected() string { return f.selected }

// Layout assigns layers. Sources and nodes with no incoming edge are layer 0.
// A later sibling on the same layer is placed below an earlier one.
func (f *Flow) Layout() {
	f.layer = map[string]int{}
	f.slot = map[string]int{}
	indeg := map[string]int{}
	for _, id := range f.ids {
		indeg[id] = 0
	}
	for _, e := range f.edges {
		if e[0] == e[1] {
			continue
		}
		if _, ok := indeg[e[1]]; ok {
			indeg[e[1]]++
		}
	}
	type item struct {
		id    string
		layer int
	}
	var queue []item
	for _, id := range f.ids {
		if indeg[id] == 0 {
			queue = append(queue, item{id, 0})
			f.layer[id] = 0
		}
	}
	seen := map[string]bool{}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if seen[cur.id] {
			continue
		}
		seen[cur.id] = true
		for _, e := range f.edges {
			if e[0] != cur.id || e[0] == e[1] {
				continue
			}
			if _, ok := f.labels[e[1]]; !ok {
				continue
			}
			next := cur.layer + 1
			if have, ok := f.layer[e[1]]; !ok || next > have {
				f.layer[e[1]] = next
			}
			indeg[e[1]]--
			if indeg[e[1]] <= 0 && !seen[e[1]] {
				queue = append(queue, item{id: e[1], layer: f.layer[e[1]]})
			}
		}
	}
	extra := 0
	for _, layer := range f.layer {
		if layer+1 > extra {
			extra = layer + 1
		}
	}
	for _, id := range f.ids {
		if _, ok := f.layer[id]; !ok {
			f.layer[id] = extra
		}
	}
	count := map[int]int{}
	for _, id := range f.ids {
		layer := f.layer[id]
		f.slot[id] = count[layer]
		count[layer]++
	}
}

func (f *Flow) Layer(id string) int {
	if f.layer == nil {
		return -1
	}
	layer, ok := f.layer[id]
	if !ok {
		return -1
	}
	return layer
}

func (f *Flow) nodeWorld(id string) uimath.Rect {
	return uimath.NewRect(
		float32(f.layer[id])*(flowNodeW+flowGapX),
		float32(f.slot[id])*(flowNodeH+flowGapY),
		flowNodeW, flowNodeH,
	)
}

// NodeScreen returns the node rectangle in screen space. Position and size
// both follow the camera zoom.
func (f *Flow) NodeScreen(id string) uimath.Rect {
	f.sync()
	w := f.nodeWorld(id)
	x, y := f.cam.WorldToScreen(w.X, w.Y)
	z := f.cam.Zoom()
	if z < 0 {
		z = 0
	}
	return uimath.NewRect(x, y, w.Width*z, w.Height*z)
}

// NodeAt returns the top-most node under a screen point, or "".
func (f *Flow) NodeAt(x, y float32) string {
	b := f.Bounds()
	if !b.Contains(uimath.Vec2{X: x, Y: y}) {
		return ""
	}
	f.sync()
	wx, wy := f.cam.ScreenToWorld(x, y)
	p := uimath.Vec2{X: wx, Y: wy}
	for i := len(f.ids) - 1; i >= 0; i-- {
		id := f.ids[i]
		if _, ok := f.layer[id]; !ok {
			continue
		}
		if f.nodeWorld(id).Contains(p) {
			return id
		}
	}
	return ""
}

func (f *Flow) sync() {
	if f.cam == nil {
		return
	}
	f.tree.SetLayout(f.cam.ElementID(), core.LayoutResult{Bounds: f.Bounds()})
}

func (f *Flow) Draw(buf *render.CommandBuffer) {
	b := f.Bounds()
	if b.IsEmpty() {
		return
	}
	f.sync()
	f.cam.Draw(buf)
	buf.PushClipZ(b, 5)
	cfg := f.config
	for _, e := range f.edges {
		if e[0] == e[1] {
			continue
		}
		if _, ok := f.layer[e[0]]; !ok {
			continue
		}
		if _, ok := f.layer[e[1]]; !ok {
			continue
		}
		a := f.NodeScreen(e[0])
		c := f.NodeScreen(e[1])
		x0 := a.X + a.Width
		y0 := a.Y + a.Height/2
		x1 := c.X
		y1 := c.Y + c.Height/2
		mid := (x0 + x1) / 2
		buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(min32(x0, mid), y0, abs32(mid-x0), 2), FillColor: cfg.BorderColor}, 5, 1)
		buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(mid, min32(y0, y1), 2, abs32(y1-y0)), FillColor: cfg.BorderColor}, 5, 1)
		buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(min32(mid, x1), y1, abs32(x1-mid), 2), FillColor: cfg.BorderColor}, 5, 1)
	}
	for _, id := range f.ids {
		if _, ok := f.layer[id]; !ok {
			continue
		}
		s := f.NodeScreen(id)
		fill := cfg.PrimaryColor
		if id == f.selected {
			fill = cfg.ActiveColor
		}
		buf.DrawRect(render.RectCmd{Bounds: s, FillColor: fill}, 6, 1)
		label := f.labels[id]
		if label != "" && cfg.TextRenderer != nil {
			buf.PushClip(s)
			size := cfg.FontSize * f.cam.Zoom()
			if size < 1 {
				size = 1
			}
			tw := cfg.TextRenderer.MeasureText(label, size)
			if tw > s.Width {
				tw = s.Width
			}
			lh := cfg.TextRenderer.LineHeight(size)
			tx := s.X + (s.Width-tw)/2
			ty := s.Y + (s.Height-lh)/2
			if tx < s.X {
				tx = s.X
			}
			if ty < s.Y {
				ty = s.Y
			}
			emitText(cfg, buf, label, tx, ty, size, s.Width, cfg.TextColor, 1, 8)
			buf.PopClip()
		}
	}
	buf.PopClipZ(7)
}

func (f *Flow) Destroy() {
	if f.cam != nil {
		f.cam.Destroy()
		f.cam = nil
	}
	f.Base.Destroy()
}

func min32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

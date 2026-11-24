package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Commit is one row in a history graph. Parents refer to other commit IDs.
// The list is newest first.
type Commit struct {
	ID      string
	Parents []string
	Subject string
	Author  string
}

// CommitGraph draws a lane graph of commits as list rows. It is a view of
// data the caller supplies, not a git client.
type CommitGraph struct {
	Base
	commits []Commit
	laneOf  map[string]int
	index   int
	scroll  float32
	rowH    float32
}

// NewCommitGraph creates an empty history view.
func NewCommitGraph(tree *core.Tree, cfg *Config) *CommitGraph {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	g := &CommitGraph{Base: NewBase(tree, core.TypeCustom, cfg), laneOf: map[string]int{}, rowH: 28}
	tree.SetFocusable(g.id, true)
	tree.AddHandler(g.id, event.KeyDown, func(e *event.Event) {
		if !targetPhase(e) || len(g.commits) == 0 {
			return
		}
		switch e.Key {
		case event.KeyArrowDown:
			g.move(1)
		case event.KeyArrowUp:
			g.move(-1)
		default:
			return
		}
		e.StopPropagation()
	})
	tree.AddHandler(g.id, event.MouseClick, func(e *event.Event) {
		if !targetPhase(e) {
			return
		}
		if i := g.rowAt(e.GlobalX, e.GlobalY); i >= 0 {
			g.index = i
			g.reveal()
		}
	})
	return g
}

// SetCommits replaces the history and assigns lanes. The first parent keeps
// the child's lane. Extra parents open a new lane. This is not a full
// force-directed branch layout.
func (g *CommitGraph) SetCommits(rows []Commit) {
	g.commits = append([]Commit(nil), rows...)
	g.laneOf = map[string]int{}
	g.index = 0
	g.scroll = 0
	active := []string{}
	for _, c := range g.commits {
		lane := -1
		for i, id := range active {
			if id == c.ID {
				lane = i
				break
			}
		}
		if lane < 0 {
			lane = len(active)
			active = append(active, "")
		}
		g.laneOf[c.ID] = lane
		if len(c.Parents) == 0 {
			active[lane] = ""
			continue
		}
		active[lane] = c.Parents[0]
		for _, p := range c.Parents[1:] {
			lane++
			active = append(active, "")
			copy(active[lane+1:], active[lane:])
			active[lane] = p
		}
	}
}

func (g *CommitGraph) Lane(id string) int {
	lane, ok := g.laneOf[id]
	if !ok {
		return -1
	}
	return lane
}

func (g *CommitGraph) Selected() string {
	if g.index < 0 || g.index >= len(g.commits) {
		return ""
	}
	return g.commits[g.index].ID
}

func (g *CommitGraph) SetScrollY(y float32) {
	if y < 0 {
		y = 0
	}
	g.scroll = y
}

// RowScreen is the row rectangle in screen space, before viewport clipping.
func (g *CommitGraph) RowScreen(i int) uimath.Rect {
	b := g.Bounds()
	return uimath.NewRect(b.X, b.Y+float32(i)*g.rowH-g.scroll, b.Width, g.rowH)
}

func (g *CommitGraph) move(dir int) {
	g.index += dir
	if g.index < 0 {
		g.index = 0
	}
	if g.index >= len(g.commits) {
		g.index = len(g.commits) - 1
	}
	g.reveal()
}

func (g *CommitGraph) reveal() {
	b := g.Bounds()
	top := float32(g.index) * g.rowH
	if top < g.scroll {
		g.scroll = top
	}
	if top+g.rowH > g.scroll+b.Height {
		g.scroll = top + g.rowH - b.Height
	}
}

func (g *CommitGraph) rowAt(x, y float32) int {
	b := g.Bounds()
	if !b.Contains(uimath.Vec2{X: x, Y: y}) {
		return -1
	}
	i := int((y - b.Y + g.scroll) / g.rowH)
	if i < 0 || i >= len(g.commits) {
		return -1
	}
	return i
}

func (g *CommitGraph) Draw(buf *render.CommandBuffer) {
	b := g.Bounds()
	if b.IsEmpty() {
		return
	}
	cfg := g.config
	buf.PushClip(b)
	rowOf := map[string]int{}
	for i, c := range g.commits {
		rowOf[c.ID] = i
	}
	maxLane := 0
	for _, lane := range g.laneOf {
		if lane > maxLane {
			maxLane = lane
		}
	}
	laneX := func(lane int, y float32) (float32, float32) {
		return b.X + 10 + float32(lane)*14, y
	}
	for i, c := range g.commits {
		y0 := b.Y + float32(i)*g.rowH - g.scroll + g.rowH/2
		for _, p := range c.Parents {
			j, ok := rowOf[p]
			if !ok {
				continue
			}
			y1 := b.Y + float32(j)*g.rowH - g.scroll + g.rowH/2
			x0, _ := laneX(g.laneOf[c.ID], y0)
			x1, _ := laneX(g.laneOf[p], y1)
			g.stroke(buf, x0, y0, x1, y1, cfg.BorderColor)
		}
	}
	textX := b.X + 18 + float32(maxLane+1)*14
	for i, c := range g.commits {
		y := b.Y + float32(i)*g.rowH - g.scroll
		if y+g.rowH <= b.Y || y >= b.Y+b.Height {
			continue
		}
		fill := uimath.ColorWhite
		if i == g.index {
			fill = cfg.HoverColor
		}
		buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(b.X, y, b.Width, g.rowH), FillColor: fill}, 1, 1)
		lx, ly := laneX(g.laneOf[c.ID], y+g.rowH/2)
		buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(lx-3, ly-3, 6, 6), FillColor: cfg.PrimaryColor}, 3, 1)
		emitText(cfg, buf, c.Subject, textX, y+4, cfg.FontSizeSm, b.X+b.Width-textX-8, cfg.TextColor, 1, 4)
	}
	buf.PopClip()
}

func (g *CommitGraph) stroke(buf *render.CommandBuffer, x0, y0, x1, y1 float32, color uimath.Color) {
	mid := (y0 + y1) / 2
	buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(min32(x0, x1), y0, abs32(x1-x0)+2, 2), FillColor: color}, 2, 1)
	buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(x1, min32(y0, y1), 2, abs32(y1-y0)+2), FillColor: color}, 2, 1)
	_ = mid
}

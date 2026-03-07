package game

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// SkillNodeState indicates whether a skill is available.
type SkillNodeState uint8

const (
	SkillLocked    SkillNodeState = iota
	SkillAvailable
	SkillUnlocked
	SkillMaxed
)

// SkillNode represents a single node in the skill tree.
type SkillNode struct {
	ID          string
	Name        string
	Description string
	Icon        render.TextureHandle
	X, Y        float32 // position in tree space
	State       SkillNodeState
	Level       int
	MaxLevel    int
	Cost        int
	Requires    []string // IDs of prerequisite skills
}

// SkillTree displays an interconnected skill/talent tree.
type SkillTree struct {
	widget.Base
	nodes      []*SkillNode
	scrollX    float32
	scrollY    float32
	nodeSize   float32
	points     int // available skill points
	onUnlock   func(nodeID string)
	onSelect   func(nodeID string)
	selected   string
}

func NewSkillTree(tree *core.Tree, cfg *widget.Config) *SkillTree {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	return &SkillTree{
		Base:     widget.NewBase(tree, core.TypeCustom, cfg),
		nodeSize: 48,
	}
}

func (st *SkillTree) Nodes() []*SkillNode        { return st.nodes }
func (st *SkillTree) Points() int                  { return st.points }
func (st *SkillTree) Selected() string             { return st.selected }
func (st *SkillTree) SetPoints(p int)              { st.points = p }
func (st *SkillTree) SetSelected(id string)        { st.selected = id }
func (st *SkillTree) SetNodeSize(s float32)        { st.nodeSize = s }
func (st *SkillTree) SetScroll(x, y float32)       { st.scrollX = x; st.scrollY = y }
func (st *SkillTree) OnUnlock(fn func(string))     { st.onUnlock = fn }
func (st *SkillTree) OnSelect(fn func(string))     { st.onSelect = fn }

func (st *SkillTree) AddNode(node *SkillNode) {
	st.nodes = append(st.nodes, node)
}

func (st *SkillTree) FindNode(id string) *SkillNode {
	for _, n := range st.nodes {
		if n.ID == id {
			return n
		}
	}
	return nil
}

func (st *SkillTree) UnlockNode(id string) bool {
	node := st.FindNode(id)
	if node == nil || node.State == SkillMaxed || st.points < node.Cost {
		return false
	}
	// Check prerequisites
	for _, req := range node.Requires {
		rn := st.FindNode(req)
		if rn == nil || (rn.State != SkillUnlocked && rn.State != SkillMaxed) {
			return false
		}
	}
	st.points -= node.Cost
	node.Level++
	if node.Level >= node.MaxLevel {
		node.State = SkillMaxed
	} else {
		node.State = SkillUnlocked
	}
	if st.onUnlock != nil {
		st.onUnlock(id)
	}
	return true
}

func (st *SkillTree) ClearNodes() {
	st.nodes = st.nodes[:0]
}

func skillNodeColor(state SkillNodeState) uimath.Color {
	switch state {
	case SkillAvailable:
		return uimath.ColorHex("#1890ff")
	case SkillUnlocked:
		return uimath.ColorHex("#52c41a")
	case SkillMaxed:
		return uimath.ColorHex("#ffd700")
	default: // locked
		return uimath.RGBA(0.3, 0.3, 0.3, 1)
	}
}

func (st *SkillTree) Draw(buf *render.CommandBuffer) {
	bounds := st.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := st.Config()
	s := st.nodeSize

	// Build ID->node map for connections
	nodeMap := make(map[string]*SkillNode, len(st.nodes))
	for _, n := range st.nodes {
		nodeMap[n.ID] = n
	}

	// Draw connections first
	for _, node := range st.nodes {
		nx := bounds.X + node.X - st.scrollX + s/2
		ny := bounds.Y + node.Y - st.scrollY + s/2
		for _, reqID := range node.Requires {
			req := nodeMap[reqID]
			if req == nil {
				continue
			}
			rx := bounds.X + req.X - st.scrollX + s/2
			ry := bounds.Y + req.Y - st.scrollY + s/2

			lineColor := uimath.RGBA(0.3, 0.3, 0.3, 0.6)
			if req.State == SkillUnlocked || req.State == SkillMaxed {
				lineColor = uimath.RGBA(0.4, 0.8, 0.4, 0.6)
			}
			// Horizontal then vertical line
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(minF(rx, nx), ry-0.5, absF(nx-rx), 1),
				FillColor: lineColor,
			}, 1, 1)
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(nx-0.5, minF(ry, ny), 1, absF(ny-ry)),
				FillColor: lineColor,
			}, 1, 1)
		}
	}

	// Draw nodes
	for _, node := range st.nodes {
		nx := bounds.X + node.X - st.scrollX
		ny := bounds.Y + node.Y - st.scrollY

		// Skip if off-screen
		if nx+s < bounds.X || nx > bounds.X+bounds.Width || ny+s < bounds.Y || ny > bounds.Y+bounds.Height {
			continue
		}

		color := skillNodeColor(node.State)
		borderW := float32(2)
		if node.ID == st.selected {
			borderW = 3
		}

		buf.DrawRect(render.RectCmd{
			Bounds:      uimath.NewRect(nx, ny, s, s),
			FillColor:   uimath.RGBA(0.12, 0.12, 0.15, 0.9),
			BorderColor: color,
			BorderWidth: borderW,
			Corners:     uimath.CornersAll(8),
		}, 2, 1)

		// Level indicator
		if cfg.TextRenderer != nil && node.MaxLevel > 0 {
			lvlText := itoa(node.Level) + "/" + itoa(node.MaxLevel)
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			tw := cfg.TextRenderer.MeasureText(lvlText, cfg.FontSizeSm)
			cfg.TextRenderer.DrawText(buf, lvlText, nx+(s-tw)/2, ny+s-lh-2, cfg.FontSizeSm, s, color, 1)
		}

		// Name below node
		if cfg.TextRenderer != nil {
			tw := cfg.TextRenderer.MeasureText(node.Name, cfg.FontSizeSm)
			cfg.TextRenderer.DrawText(buf, node.Name, nx+(s-tw)/2, ny+s+2, cfg.FontSizeSm, s*2, uimath.RGBA(0.8, 0.8, 0.8, 1), 1)
		}
	}
}

func minF(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func absF(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

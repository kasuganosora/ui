package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Skeleton displays a placeholder loading animation.
type Skeleton struct {
	Base
	rows    int
	avatar  bool
	active  bool
	rowGap  float32
	rowH    float32
}

func NewSkeleton(tree *core.Tree, cfg *Config) *Skeleton {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Skeleton{
		Base:   NewBase(tree, core.TypeCustom, cfg),
		rows:   3,
		active: true,
		rowGap: 12,
		rowH:   16,
	}
}

func (s *Skeleton) SetRows(r int)     { s.rows = r }
func (s *Skeleton) SetAvatar(a bool)  { s.avatar = a }
func (s *Skeleton) SetActive(a bool)  { s.active = a }
func (s *Skeleton) Rows() int         { return s.rows }

func (s *Skeleton) Draw(buf *render.CommandBuffer) {
	bounds := s.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := s.config
	skColor := uimath.RGBA(0, 0, 0, 0.06)
	x := bounds.X
	y := bounds.Y

	// Avatar circle
	if s.avatar {
		size := float32(40)
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x, y, size, size),
			FillColor: skColor,
			Corners:   uimath.CornersAll(size / 2),
		}, 1, 1)
		x += size + cfg.SpaceMD
	}

	// Rows
	for i := 0; i < s.rows; i++ {
		w := bounds.Width - (x - bounds.X)
		// Last row is shorter
		if i == s.rows-1 {
			w *= 0.6
		}
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x, y+float32(i)*(s.rowH+s.rowGap), w, s.rowH),
			FillColor: skColor,
			Corners:   uimath.CornersAll(cfg.BorderRadius / 2),
		}, 1, 1)
	}
}

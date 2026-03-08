package widget

import (
	gomath "math"
	"time"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// SkeletonAnimation controls the skeleton animation style.
type SkeletonAnimation uint8

const (
	AnimationNone     SkeletonAnimation = iota
	AnimationGradient                   // shimmer effect using sin(time) opacity modulation
	AnimationFlashed                    // flashing effect
)

// SkeletonTheme controls the skeleton layout preset.
type SkeletonTheme uint8

const (
	ThemeText       SkeletonTheme = iota // 3 rows of different widths
	ThemeParagraph                       // 4 rows
	ThemeAvatar                          // single circle
	ThemeAvatarText                      // circle on left + 3 text rows on right
)

// SkeletonRowColObj describes a single cell in a skeleton row.
type SkeletonRowColObj struct {
	Width  string
	Height string
	Type   string // "rect", "circle", "text"
}

// SkeletonRowCol defines custom row/column layout for the skeleton.
type SkeletonRowCol = []interface{}

// Skeleton displays a placeholder loading animation.
type Skeleton struct {
	Base
	rows      int
	avatar    bool
	rowGap    float32
	rowH      float32
	loading   bool
	animation SkeletonAnimation
	theme     SkeletonTheme
	delay     int // milliseconds before showing skeleton
	rowCol    SkeletonRowCol
	startTime time.Time
}

func NewSkeleton(tree *core.Tree, cfg *Config) *Skeleton {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Skeleton{
		Base:      NewBase(tree, core.TypeCustom, cfg),
		rows:      3,
		rowGap:    12,
		rowH:      16,
		loading:   true,
		animation: AnimationNone,
		theme:     ThemeText,
		startTime: time.Now(),
	}
}

func (s *Skeleton) SetRows(r int)                    { s.rows = r }
func (s *Skeleton) SetAvatar(a bool)                 { s.avatar = a }
func (s *Skeleton) Rows() int                        { return s.rows }
func (s *Skeleton) SetLoading(l bool)                { s.loading = l }
func (s *Skeleton) SetAnimation(a SkeletonAnimation) { s.animation = a }
func (s *Skeleton) SetTheme(t SkeletonTheme)         { s.theme = t }
func (s *Skeleton) SetDelay(ms int)                  { s.delay = ms }
func (s *Skeleton) SetRowCol(rc SkeletonRowCol)      { s.rowCol = rc }

func (s *Skeleton) Draw(buf *render.CommandBuffer) {
	if !s.loading {
		s.DrawChildren(buf)
		return
	}

	bounds := s.Bounds()
	if bounds.IsEmpty() {
		return
	}

	// Animation opacity modulation
	opacity := float32(1.0)
	if s.animation == AnimationGradient {
		elapsed := time.Since(s.startTime).Seconds()
		opacity = float32(0.4 + 0.6*gomath.Abs(gomath.Sin(elapsed*2.0)))
		s.tree.MarkDirty(s.id) // continuous redraw
	} else if s.animation == AnimationFlashed {
		elapsed := time.Since(s.startTime).Seconds()
		// Flash: alternate between 0.3 and 1.0
		phase := gomath.Mod(elapsed*2.0, 2.0)
		if phase < 1.0 {
			opacity = 0.3
		} else {
			opacity = 1.0
		}
		s.tree.MarkDirty(s.id)
	}

	cfg := s.config
	skColor := uimath.RGBA(0, 0, 0, 0.06*opacity)

	switch s.theme {
	case ThemeAvatar:
		s.drawAvatar(buf, bounds, skColor)
	case ThemeAvatarText:
		s.drawAvatarText(buf, bounds, cfg, skColor)
	case ThemeParagraph:
		s.drawRows(buf, bounds, cfg, skColor, 4, bounds.X)
	default: // ThemeText
		s.drawRows(buf, bounds, cfg, skColor, 3, bounds.X)
	}
}

func (s *Skeleton) drawAvatar(buf *render.CommandBuffer, bounds uimath.Rect, color uimath.Color) {
	size := float32(40)
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(bounds.X, bounds.Y, size, size),
		FillColor: color,
		Corners:   uimath.CornersAll(size / 2),
	}, 1, 1)
}

func (s *Skeleton) drawRows(buf *render.CommandBuffer, bounds uimath.Rect, cfg *Config, color uimath.Color, rowCount int, startX float32) {
	widths := []float32{1.0, 0.8, 0.6, 0.9} // row width ratios
	for i := 0; i < rowCount; i++ {
		w := bounds.Width - (startX - bounds.X)
		ratio := float32(1.0)
		if i < len(widths) {
			ratio = widths[i]
		}
		w *= ratio
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(startX, bounds.Y+float32(i)*(s.rowH+s.rowGap), w, s.rowH),
			FillColor: color,
			Corners:   uimath.CornersAll(cfg.BorderRadius / 2),
		}, 1, 1)
	}
}

func (s *Skeleton) drawAvatarText(buf *render.CommandBuffer, bounds uimath.Rect, cfg *Config, color uimath.Color) {
	// Avatar circle on left
	avatarSize := float32(40)
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(bounds.X, bounds.Y, avatarSize, avatarSize),
		FillColor: color,
		Corners:   uimath.CornersAll(avatarSize / 2),
	}, 1, 1)

	// Text rows on right
	textX := bounds.X + avatarSize + cfg.SpaceMD
	s.drawRows(buf, bounds, cfg, color, 3, textX)
}

package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	"github.com/kasuganosora/ui/render"
)

// SpaceDirection controls whether items are laid out horizontally or vertically.
type SpaceDirection uint8

const (
	SpaceHorizontal SpaceDirection = iota
	SpaceVertical
)

// Space arranges children with consistent spacing between them.
type Space struct {
	Base
	direction SpaceDirection
	gap       float32
}

// NewSpace creates a space layout widget.
func NewSpace(tree *core.Tree, cfg *Config) *Space {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	s := &Space{
		Base:      NewBase(tree, core.TypeDiv, cfg),
		direction: SpaceHorizontal,
		gap:       cfg.SpaceSM,
	}
	s.style.Display = layout.DisplayFlex
	s.style.FlexDirection = layout.FlexDirectionRow
	s.style.Gap = cfg.SpaceSM
	s.style.AlignItems = layout.AlignCenter
	return s
}

func (s *Space) Direction() SpaceDirection { return s.direction }
func (s *Space) Gap() float32              { return s.gap }

func (s *Space) SetDirection(d SpaceDirection) {
	s.direction = d
	if d == SpaceVertical {
		s.style.FlexDirection = layout.FlexDirectionColumn
	} else {
		s.style.FlexDirection = layout.FlexDirectionRow
	}
}

func (s *Space) SetGap(gap float32) {
	s.gap = gap
	s.style.Gap = gap
}

func (s *Space) Draw(buf *render.CommandBuffer) {
	s.DrawChildren(buf)
}

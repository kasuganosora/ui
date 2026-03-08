package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// SpaceDirection controls whether items are laid out horizontally or vertically.
type SpaceDirection uint8

const (
	SpaceHorizontal SpaceDirection = iota
	SpaceVertical
)

// SpaceAlign controls cross-axis alignment of children.
type SpaceAlign uint8

const (
	SpaceAlignStart    SpaceAlign = iota
	SpaceAlignCenter
	SpaceAlignEnd
	SpaceAlignBaseline
)

// Space arranges children with consistent spacing between them.
type Space struct {
	Base
	direction SpaceDirection
	gap       float32
	align     SpaceAlign
	breakLine bool
	separator string
	size      Size
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
		align:     SpaceAlignCenter,
		size:      SizeMedium,
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

func (s *Space) SetAlign(a SpaceAlign) {
	s.align = a
	switch a {
	case SpaceAlignStart:
		s.style.AlignItems = layout.AlignFlexStart
	case SpaceAlignEnd:
		s.style.AlignItems = layout.AlignFlexEnd
	case SpaceAlignBaseline:
		s.style.AlignItems = layout.AlignBaseline
	default:
		s.style.AlignItems = layout.AlignCenter
	}
}

func (s *Space) SetBreakLine(v bool) {
	s.breakLine = v
	if v {
		s.style.FlexWrap = layout.FlexWrapWrap
	} else {
		s.style.FlexWrap = layout.FlexWrapNoWrap
	}
}

func (s *Space) SetSeparator(sep string) {
	s.separator = sep
}

func (s *Space) SetSize(sz Size) {
	s.size = sz
	switch sz {
	case SizeSmall:
		s.SetGap(8)
	case SizeLarge:
		s.SetGap(24)
	default:
		s.SetGap(16)
	}
}

func (s *Space) childBounds(child Widget) uimath.Rect {
	if e := s.tree.Get(child.ElementID()); e != nil {
		return e.Layout().Bounds
	}
	return uimath.Rect{}
}

func (s *Space) Draw(buf *render.CommandBuffer) {
	if s.separator != "" && len(s.children) > 1 {
		cfg := s.config
		for i, child := range s.children {
			child.Draw(buf)
			if i < len(s.children)-1 {
				// Draw separator between children
				childBounds := s.childBounds(child)
				if !childBounds.IsEmpty() && cfg.TextRenderer != nil {
					if s.direction == SpaceVertical {
						// Horizontal separator between vertically stacked children
						sepX := childBounds.X
						sepY := childBounds.Y + childBounds.Height + s.gap/2
						tw := cfg.TextRenderer.MeasureText(s.separator, cfg.FontSizeSm)
						lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
						cfg.TextRenderer.DrawText(buf, s.separator, sepX+(childBounds.Width-tw)/2, sepY-lh/2, cfg.FontSizeSm, childBounds.Width, cfg.DisabledColor, 1)
					} else {
						// Vertical separator between horizontally placed children
						sepX := childBounds.X + childBounds.Width + s.gap/2
						sepY := childBounds.Y
						tw := cfg.TextRenderer.MeasureText(s.separator, cfg.FontSizeSm)
						lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
						cfg.TextRenderer.DrawText(buf, s.separator, sepX-tw/2, sepY+(childBounds.Height-lh)/2, cfg.FontSizeSm, tw+4, cfg.DisabledColor, 1)
					}
				}
			}
		}
		return
	}
	s.DrawChildren(buf)
}

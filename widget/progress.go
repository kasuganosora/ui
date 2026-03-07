package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// ProgressType controls the progress bar style.
type ProgressType uint8

const (
	ProgressLine   ProgressType = iota // Horizontal bar (default)
	ProgressCircle                     // Circular (future)
)

// ProgressStatus controls the progress bar color.
type ProgressStatus uint8

const (
	ProgressNormal  ProgressStatus = iota
	ProgressSuccess
	ProgressError
	ProgressActive
)

const progressHeight = float32(8)

// Progress displays a progress bar.
type Progress struct {
	Base
	percent float32 // 0-100
	status  ProgressStatus
}

// NewProgress creates a progress bar.
func NewProgress(tree *core.Tree, cfg *Config) *Progress {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	p := &Progress{
		Base: NewBase(tree, core.TypeCustom, cfg),
	}
	p.style.Display = layout.DisplayBlock
	p.style.Height = layout.Px(progressHeight)
	return p
}

func (p *Progress) Percent() float32       { return p.percent }
func (p *Progress) Status() ProgressStatus { return p.status }

func (p *Progress) SetPercent(pct float32) {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	p.percent = pct
}

func (p *Progress) SetStatus(s ProgressStatus) { p.status = s }

func (p *Progress) barColor() uimath.Color {
	switch p.status {
	case ProgressSuccess:
		return uimath.ColorHex("#52c41a")
	case ProgressError:
		return uimath.ColorHex("#ff4d4f")
	default:
		return p.config.PrimaryColor
	}
}

func (p *Progress) Draw(buf *render.CommandBuffer) {
	bounds := p.Bounds()
	if bounds.IsEmpty() {
		return
	}

	radius := bounds.Height / 2

	// Track
	buf.DrawRect(render.RectCmd{
		Bounds:  bounds,
		FillColor: uimath.ColorHex("#f5f5f5"),
		Corners: uimath.CornersAll(radius),
	}, 0, 1)

	// Fill
	if p.percent > 0 {
		fillW := bounds.Width * p.percent / 100
		if fillW < bounds.Height {
			fillW = bounds.Height // minimum width = height for rounded ends
		}
		buf.DrawRect(render.RectCmd{
			Bounds:  uimath.NewRect(bounds.X, bounds.Y, fillW, bounds.Height),
			FillColor: p.barColor(),
			Corners: uimath.CornersAll(radius),
		}, 1, 1)
	}

	p.DrawChildren(buf)
}

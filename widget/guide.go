package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// GuideStep represents a single step in the guided tour.
type GuideStep struct {
	Title       string
	Description string
	TargetX     float32
	TargetY     float32
	TargetW     float32
	TargetH     float32
}

// Guide provides a step-by-step onboarding overlay.
type Guide struct {
	Base
	steps     []GuideStep
	current   int
	visible   bool
	maskColor uimath.Color
	onFinish  func()
	onChange  func(int)
}

func NewGuide(tree *core.Tree, cfg *Config) *Guide {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Guide{
		Base:      NewBase(tree, core.TypeCustom, cfg),
		maskColor: uimath.RGBA(0, 0, 0, 0.5),
	}
}

func (g *Guide) Steps() []GuideStep    { return g.steps }
func (g *Guide) Current() int          { return g.current }
func (g *Guide) IsVisible() bool       { return g.visible }
func (g *Guide) SetMaskColor(c uimath.Color) { g.maskColor = c }
func (g *Guide) OnFinish(fn func())    { g.onFinish = fn }
func (g *Guide) OnChange(fn func(int)) { g.onChange = fn }

func (g *Guide) SetSteps(steps []GuideStep) {
	g.steps = make([]GuideStep, len(steps))
	copy(g.steps, steps)
}

func (g *Guide) Start() {
	if len(g.steps) > 0 {
		g.current = 0
		g.visible = true
	}
}

func (g *Guide) Next() {
	if g.current < len(g.steps)-1 {
		g.current++
		if g.onChange != nil {
			g.onChange(g.current)
		}
	} else {
		g.Finish()
	}
}

func (g *Guide) Prev() {
	if g.current > 0 {
		g.current--
		if g.onChange != nil {
			g.onChange(g.current)
		}
	}
}

func (g *Guide) Finish() {
	g.visible = false
	if g.onFinish != nil {
		g.onFinish()
	}
}

func (g *Guide) Draw(buf *render.CommandBuffer) {
	if !g.visible || len(g.steps) == 0 || g.current >= len(g.steps) {
		return
	}
	cfg := g.config
	step := g.steps[g.current]

	// Semi-transparent mask overlay (full screen — using large bounds)
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(0, 0, 9999, 9999),
		FillColor: g.maskColor,
	}, 80, 1)

	// Highlight cutout (clear area around target)
	if step.TargetW > 0 && step.TargetH > 0 {
		buf.DrawOverlay(render.RectCmd{
			Bounds:      uimath.NewRect(step.TargetX-4, step.TargetY-4, step.TargetW+8, step.TargetH+8),
			BorderColor: uimath.ColorHex("#0052d9"),
			BorderWidth: 2,
			Corners:     uimath.CornersAll(cfg.BorderRadius),
		}, 82, 1)
	}

	// Tooltip card
	cardW := float32(280)
	cardH := float32(100)
	cardX := step.TargetX + step.TargetW + 12
	cardY := step.TargetY

	buf.DrawOverlay(render.RectCmd{
		Bounds:      uimath.NewRect(cardX, cardY, cardW, cardH),
		FillColor:   uimath.ColorWhite,
		BorderColor: cfg.BorderColor,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 83, 1)

	if cfg.TextRenderer != nil {
		// Title
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, step.Title, cardX+cfg.SpaceSM, cardY+cfg.SpaceSM, cfg.FontSize, cardW-cfg.SpaceSM*2, cfg.TextColor, 1)

		// Description
		if step.Description != "" {
			cfg.TextRenderer.DrawText(buf, step.Description, cardX+cfg.SpaceSM, cardY+cfg.SpaceSM+lh+4, cfg.FontSizeSm, cardW-cfg.SpaceSM*2, cfg.DisabledColor, 1)
		}

		// Step indicator
		indicator := intToStr(g.current+1) + " / " + intToStr(len(g.steps))
		tw := cfg.TextRenderer.MeasureText(indicator, cfg.FontSizeSm)
		cfg.TextRenderer.DrawText(buf, indicator, cardX+cardW-tw-cfg.SpaceSM, cardY+cardH-cfg.TextRenderer.LineHeight(cfg.FontSizeSm)-cfg.SpaceSM, cfg.FontSizeSm, tw+4, cfg.DisabledColor, 1)
	}
}

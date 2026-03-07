package game

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// CastBar shows a spell/ability casting progress bar.
type CastBar struct {
	widget.Base
	spellName  string
	progress   float32 // 0-1
	castTime   float32 // total cast time in seconds
	elapsed    float32 // elapsed time
	color      uimath.Color
	visible    bool
	width      float32
	height     float32
	onComplete func()
	onInterrupt func()
}

func NewCastBar(tree *core.Tree, cfg *widget.Config) *CastBar {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	return &CastBar{
		Base:   widget.NewBase(tree, core.TypeCustom, cfg),
		color:  uimath.ColorHex("#ffd700"),
		width:  250,
		height: 20,
	}
}

func (cb *CastBar) SpellName() string          { return cb.spellName }
func (cb *CastBar) Progress() float32          { return cb.progress }
func (cb *CastBar) IsVisible() bool            { return cb.visible }
func (cb *CastBar) SetColor(c uimath.Color)    { cb.color = c }
func (cb *CastBar) SetSize(w, h float32)       { cb.width = w; cb.height = h }
func (cb *CastBar) OnComplete(fn func())       { cb.onComplete = fn }
func (cb *CastBar) OnInterrupt(fn func())      { cb.onInterrupt = fn }

func (cb *CastBar) StartCast(spellName string, castTime float32) {
	cb.spellName = spellName
	cb.castTime = castTime
	cb.elapsed = 0
	cb.progress = 0
	cb.visible = true
}

func (cb *CastBar) Interrupt() {
	cb.visible = false
	if cb.onInterrupt != nil {
		cb.onInterrupt()
	}
}

func (cb *CastBar) Tick(dt float32) {
	if !cb.visible || cb.castTime <= 0 {
		return
	}
	cb.elapsed += dt
	cb.progress = cb.elapsed / cb.castTime
	if cb.progress >= 1 {
		cb.progress = 1
		cb.visible = false
		if cb.onComplete != nil {
			cb.onComplete()
		}
	}
}

func (cb *CastBar) Draw(buf *render.CommandBuffer) {
	if !cb.visible {
		return
	}
	cfg := cb.Config()
	bounds := cb.Bounds()
	x, y := bounds.X, bounds.Y
	if bounds.IsEmpty() {
		x, y = 0, 0
	}
	r := cb.height / 2

	// Background
	buf.DrawOverlay(render.RectCmd{
		Bounds:      uimath.NewRect(x, y, cb.width, cb.height),
		FillColor:   uimath.RGBA(0, 0, 0, 0.7),
		BorderColor: uimath.RGBA(0.4, 0.4, 0.4, 0.8),
		BorderWidth: 1,
		Corners:     uimath.CornersAll(r),
	}, 30, 1)

	// Fill
	if cb.progress > 0 {
		fillW := cb.width * cb.progress
		buf.DrawOverlay(render.RectCmd{
			Bounds:    uimath.NewRect(x, y, fillW, cb.height),
			FillColor: cb.color,
			Corners:   uimath.CornersAll(r),
		}, 31, 1)
	}

	// Spell name
	if cfg.TextRenderer != nil && cb.spellName != "" {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
		tw := cfg.TextRenderer.MeasureText(cb.spellName, cfg.FontSizeSm)
		cfg.TextRenderer.DrawText(buf, cb.spellName, x+(cb.width-tw)/2, y+(cb.height-lh)/2, cfg.FontSizeSm, cb.width, uimath.ColorWhite, 1)
	}
}

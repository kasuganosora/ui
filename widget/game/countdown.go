package game

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// CountdownTimer displays a countdown timer.
type CountdownTimer struct {
	widget.Base
	seconds   float32
	label     string
	color     uimath.Color
	fontSize  float32
	onExpire  func()
}

func NewCountdownTimer(tree *core.Tree, cfg *widget.Config) *CountdownTimer {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	return &CountdownTimer{
		Base:     widget.NewBase(tree, core.TypeCustom, cfg),
		color:    uimath.ColorWhite,
		fontSize: 0, // 0 = use cfg.FontSize
	}
}

func (ct *CountdownTimer) Seconds() float32         { return ct.seconds }
func (ct *CountdownTimer) SetSeconds(s float32)     { ct.seconds = s }
func (ct *CountdownTimer) SetLabel(l string)        { ct.label = l }
func (ct *CountdownTimer) SetColor(c uimath.Color)  { ct.color = c }
func (ct *CountdownTimer) SetFontSize(s float32)    { ct.fontSize = s }
func (ct *CountdownTimer) OnExpire(fn func())       { ct.onExpire = fn }
func (ct *CountdownTimer) IsExpired() bool          { return ct.seconds <= 0 }

func (ct *CountdownTimer) Tick(dt float32) {
	if ct.seconds <= 0 {
		return
	}
	ct.seconds -= dt
	if ct.seconds <= 0 {
		ct.seconds = 0
		if ct.onExpire != nil {
			ct.onExpire()
		}
	}
}

func (ct *CountdownTimer) formatTime() string {
	total := int(ct.seconds)
	if total < 0 {
		total = 0
	}
	m := total / 60
	s := total % 60
	if m > 0 {
		return itoa(m) + ":" + pad2(s)
	}
	return itoa(s)
}

func pad2(n int) string {
	if n < 10 {
		return "0" + itoa(n)
	}
	return itoa(n)
}

func (ct *CountdownTimer) Draw(buf *render.CommandBuffer) {
	bounds := ct.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := ct.Config()
	fs := ct.fontSize
	if fs <= 0 {
		fs = cfg.FontSize
	}

	text := ct.formatTime()
	color := ct.color
	// Flash red when low
	if ct.seconds > 0 && ct.seconds < 10 {
		color = uimath.ColorHex("#ff4d4f")
	}

	if cfg.TextRenderer != nil {
		tw := cfg.TextRenderer.MeasureText(text, fs)
		lh := cfg.TextRenderer.LineHeight(fs)
		tx := bounds.X + (bounds.Width-tw)/2
		ty := bounds.Y + (bounds.Height-lh)/2
		cfg.TextRenderer.DrawText(buf, text, tx, ty, fs, bounds.Width, color, 1)

		// Label above
		if ct.label != "" {
			llh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			ltw := cfg.TextRenderer.MeasureText(ct.label, cfg.FontSizeSm)
			cfg.TextRenderer.DrawText(buf, ct.label, bounds.X+(bounds.Width-ltw)/2, ty-llh-2, cfg.FontSizeSm, bounds.Width, uimath.RGBA(0.7, 0.7, 0.7, 1), 1)
		}
	}
}

package widget

import (
	"math"
	"time"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

const (
	loadingSize    = float32(32)
	loadingDotSize = float32(6)
)

// Loading displays a loading indicator (three dots with animation).
type Loading struct {
	Base
	text         string
	start        time.Time
	delay        int    // milliseconds before showing
	fullscreen   bool
	indicator    bool
	inheritColor bool
	loading      bool
	size         string // "small", "medium", "large", or CSS value
}

// NewLoading creates a loading indicator.
func NewLoading(tree *core.Tree, cfg *Config) *Loading {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	l := &Loading{
		Base:      NewBase(tree, core.TypeCustom, cfg),
		start:     time.Now(),
		indicator: true,
		loading:   true,
		size:      "medium",
	}
	l.style.Display = layout.DisplayFlex
	l.style.AlignItems = layout.AlignCenter
	l.style.JustifyContent = layout.JustifyCenter
	l.style.FlexDirection = layout.FlexDirectionColumn
	l.style.Gap = cfg.SpaceSM
	return l
}

func (l *Loading) Text() string              { return l.text }
func (l *Loading) SetText(t string)          { l.text = t }
func (l *Loading) SetDelay(ms int)           { l.delay = ms }
func (l *Loading) SetFullscreen(v bool)      { l.fullscreen = v }
func (l *Loading) SetIndicator(v bool)       { l.indicator = v }
func (l *Loading) SetInheritColor(v bool)    { l.inheritColor = v }
func (l *Loading) SetLoading(v bool)         { l.loading = v }
func (l *Loading) SetSize(s string)          { l.size = s }

func (l *Loading) Draw(buf *render.CommandBuffer) {
	bounds := l.Bounds()
	if bounds.IsEmpty() {
		return
	}

	// Request continuous redraws for animation
	l.tree.MarkDirty(l.id)

	cfg := l.config
	color := cfg.PrimaryColor

	// Animated three dots: each dot bounces up with a phase offset
	elapsed := float32(time.Since(l.start).Seconds())
	const period = float32(1.2) // full cycle duration in seconds

	dotY := bounds.Y + (bounds.Height-loadingDotSize)/2
	if l.text != "" {
		dotY -= cfg.FontSize
	}
	totalDotsW := loadingDotSize*3 + cfg.SpaceSM*2
	dotX := bounds.X + (bounds.Width-totalDotsW)/2

	for i := 0; i < 3; i++ {
		// Each dot has a phase offset: 0, 0.15, 0.3 seconds
		phase := elapsed - float32(i)*0.15
		// Normalize to [0, period), then compute bounce
		t := float32(math.Mod(float64(phase), float64(period))) / period
		// Bounce: active during first half, using sin curve
		bounce := float32(0)
		if t < 0.5 {
			bounce = float32(math.Sin(float64(t) * 2 * math.Pi)) * 6
		}
		// Opacity: brighter when bouncing
		opacity := float32(0.35)
		if t < 0.5 {
			opacity = 0.35 + 0.65*float32(math.Sin(float64(t)*2*math.Pi))
		}

		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(dotX, dotY-bounce, loadingDotSize, loadingDotSize),
			FillColor: color,
			Corners:   uimath.CornersAll(loadingDotSize / 2),
		}, 0, opacity)
		dotX += loadingDotSize + cfg.SpaceSM
	}

	// Tip text
	if l.text != "" && cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		tipY := dotY + loadingDotSize + cfg.SpaceSM
		tipW := cfg.TextRenderer.MeasureText(l.text, cfg.FontSize)
		tipX := bounds.X + (bounds.Width-tipW)/2
		cfg.TextRenderer.DrawText(buf, l.text, tipX, tipY, cfg.FontSize, bounds.Width, cfg.TextColor, 1)
		_ = lh
	}

	l.DrawChildren(buf)
}

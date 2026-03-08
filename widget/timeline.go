package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// TimelineLabelAlign controls label position relative to the axis.
type TimelineLabelAlign uint8

const (
	TimelineLabelLeft TimelineLabelAlign = iota
	TimelineLabelRight
	TimelineLabelAlternate
	TimelineLabelTop
	TimelineLabelBottom
)

// TimelineLayout controls timeline orientation.
type TimelineLayout uint8

const (
	TimelineVertical TimelineLayout = iota
	TimelineHorizontal
)

// TimelineMode controls label and content placement.
type TimelineMode uint8

const (
	TimelineModeAlternate TimelineMode = iota
	TimelineModeSame
)

// TimelineTheme controls dot styling.
type TimelineTheme uint8

const (
	TimelineThemeDefault TimelineTheme = iota
	TimelineThemeDot
)

// TimelineItem represents a single event in the timeline.
type TimelineItem struct {
	Content    string
	Label      string
	DotColor   string // hex color or "primary"/"warning"/"error"/"default"
	LabelAlign TimelineLabelAlign
	Loading    bool
	OnClick    func()
}

// Timeline displays a vertical list of events.
type Timeline struct {
	Base
	items      []TimelineItem
	itemH      float32
	dotSize    float32
	labelAlign TimelineLabelAlign
	layout     TimelineLayout
	mode       TimelineMode
	reverse    bool
	theme      TimelineTheme
}

func NewTimeline(tree *core.Tree, cfg *Config) *Timeline {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Timeline{
		Base:    NewBase(tree, core.TypeCustom, cfg),
		itemH:   60,
		dotSize: 10,
	}
}

func (t *Timeline) Items() []TimelineItem             { return t.items }
func (t *Timeline) SetItemHeight(h float32)            { t.itemH = h }
func (t *Timeline) SetLabelAlign(a TimelineLabelAlign) { t.labelAlign = a }
func (t *Timeline) SetLayout(l TimelineLayout)         { t.layout = l }
func (t *Timeline) SetMode(m TimelineMode)             { t.mode = m }
func (t *Timeline) SetReverse(v bool)                  { t.reverse = v }
func (t *Timeline) SetTheme(th TimelineTheme)          { t.theme = th }

func (t *Timeline) AddItem(item TimelineItem) {
	t.items = append(t.items, item)
}

func (t *Timeline) ClearItems() {
	t.items = t.items[:0]
}

func (t *Timeline) TotalHeight() float32 {
	return float32(len(t.items)) * t.itemH
}

func timelineDotColor(dotColor string) uimath.Color {
	switch dotColor {
	case "warning":
		return uimath.ColorHex("#faad14")
	case "error":
		return uimath.ColorHex("#ff4d4f")
	case "default":
		return uimath.ColorHex("#c0c4cc")
	case "primary", "":
		return uimath.ColorHex("#1890ff")
	default:
		// Treat as hex color string
		return uimath.ColorHex(dotColor)
	}
}

func (t *Timeline) Draw(buf *render.CommandBuffer) {
	bounds := t.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := t.config
	dotR := t.dotSize / 2
	lineX := bounds.X + dotR

	for i, item := range t.items {
		y := bounds.Y + float32(i)*t.itemH
		if y+t.itemH > bounds.Y+bounds.Height {
			break
		}
		dotColor := timelineDotColor(item.DotColor)

		// Connecting line (not for last item)
		if i < len(t.items)-1 {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(lineX-0.5, y+t.dotSize, 1, t.itemH-t.dotSize),
				FillColor: cfg.BorderColor,
			}, 1, 1)
		}

		// Dot (outlined ring like TDesign)
		buf.DrawRect(render.RectCmd{
			Bounds:      uimath.NewRect(bounds.X, y, t.dotSize, t.dotSize),
			FillColor:   uimath.ColorWhite,
			BorderColor: dotColor,
			BorderWidth: 2,
			Corners:     uimath.CornersAll(dotR),
		}, 2, 1)

		// Label
		textX := bounds.X + t.dotSize + cfg.SpaceMD
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			cfg.TextRenderer.DrawText(buf, item.Label, textX, y, cfg.FontSize, bounds.Width-textX+bounds.X, cfg.TextColor, 1)
			if item.Content != "" {
				cfg.TextRenderer.DrawText(buf, item.Content, textX, y+lh+2, cfg.FontSizeSm, bounds.Width-textX+bounds.X, cfg.DisabledColor, 1)
			}
		}
	}
}

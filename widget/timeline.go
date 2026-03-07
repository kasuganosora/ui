package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// TimelineItemStatus determines the dot color.
type TimelineItemStatus uint8

const (
	TimelineDefault TimelineItemStatus = iota
	TimelineSuccess
	TimelineWarning
	TimelineError
)

// TimelineItem represents a single event in the timeline.
type TimelineItem struct {
	Label  string
	Detail string
	Status TimelineItemStatus
}

// Timeline displays a vertical list of events.
type Timeline struct {
	Base
	items    []TimelineItem
	itemH    float32
	dotSize  float32
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

func (t *Timeline) Items() []TimelineItem  { return t.items }
func (t *Timeline) SetItemHeight(h float32) { t.itemH = h }

func (t *Timeline) AddItem(item TimelineItem) {
	t.items = append(t.items, item)
}

func (t *Timeline) ClearItems() {
	t.items = t.items[:0]
}

func timelineStatusColor(s TimelineItemStatus) uimath.Color {
	switch s {
	case TimelineSuccess:
		return uimath.ColorHex("#52c41a")
	case TimelineWarning:
		return uimath.ColorHex("#faad14")
	case TimelineError:
		return uimath.ColorHex("#ff4d4f")
	default:
		return uimath.ColorHex("#1890ff")
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
		dotColor := timelineStatusColor(item.Status)

		// Connecting line (not for last item)
		if i < len(t.items)-1 {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(lineX-0.5, y+t.dotSize, 1, t.itemH-t.dotSize),
				FillColor: cfg.BorderColor,
			}, 1, 1)
		}

		// Dot
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(bounds.X, y, t.dotSize, t.dotSize),
			FillColor: dotColor,
			Corners:   uimath.CornersAll(dotR),
		}, 2, 1)

		// Label
		textX := bounds.X + t.dotSize + cfg.SpaceMD
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			cfg.TextRenderer.DrawText(buf, item.Label, textX, y, cfg.FontSize, bounds.Width-textX+bounds.X, cfg.TextColor, 1)
			if item.Detail != "" {
				cfg.TextRenderer.DrawText(buf, item.Detail, textX, y+lh+2, cfg.FontSizeSm, bounds.Width-textX+bounds.X, cfg.DisabledColor, 1)
			}
		}
	}
}

package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// DateRangePicker selects a start and end date.
type DateRangePicker struct {
	Base
	startYear, startMonth, startDay int
	endYear, endMonth, endDay       int
	open                            bool
	onChange                         func(sy, sm, sd, ey, em, ed int)
}

func NewDateRangePicker(tree *core.Tree, cfg *Config) *DateRangePicker {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	drp := &DateRangePicker{
		Base:       NewBase(tree, core.TypeCustom, cfg),
		startYear:  2026, startMonth: 1, startDay: 1,
		endYear: 2026, endMonth: 1, endDay: 31,
	}
	tree.AddHandler(drp.id, event.MouseClick, func(e *event.Event) {
		drp.open = !drp.open
	})
	return drp
}

func (drp *DateRangePicker) StartDate() (int, int, int) {
	return drp.startYear, drp.startMonth, drp.startDay
}
func (drp *DateRangePicker) EndDate() (int, int, int) {
	return drp.endYear, drp.endMonth, drp.endDay
}
func (drp *DateRangePicker) IsOpen() bool { return drp.open }
func (drp *DateRangePicker) SetOpen(o bool) { drp.open = o }

func (drp *DateRangePicker) SetStartDate(y, m, d int) {
	drp.startYear = y; drp.startMonth = m; drp.startDay = d
}
func (drp *DateRangePicker) SetEndDate(y, m, d int) {
	drp.endYear = y; drp.endMonth = m; drp.endDay = d
}
func (drp *DateRangePicker) OnChange(fn func(int, int, int, int, int, int)) {
	drp.onChange = fn
}

func (drp *DateRangePicker) Draw(buf *render.CommandBuffer) {
	bounds := drp.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := drp.config

	label := dateStr(drp.startYear, drp.startMonth, drp.startDay) + " ~ " + dateStr(drp.endYear, drp.endMonth, drp.endDay)
	buf.DrawRect(render.RectCmd{
		Bounds:      bounds,
		FillColor:   cfg.BgColor,
		BorderColor: cfg.BorderColor,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 1, 1)
	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, label, bounds.X+cfg.SpaceSM, bounds.Y+(bounds.Height-lh)/2, cfg.FontSize, bounds.Width-cfg.SpaceSM*2, cfg.TextColor, 1)
	}
}

package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// DatePicker provides a date selection interface.
type DatePicker struct {
	Base
	year     int
	month    int // 1-12
	day      int // 1-31
	open     bool
	width    float32
	onChange func(year, month, day int)
}

func NewDatePicker(tree *core.Tree, cfg *Config) *DatePicker {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	dp := &DatePicker{
		Base:  NewBase(tree, core.TypeCustom, cfg),
		year:  2026,
		month: 1,
		day:   1,
		width: 260,
	}
	tree.AddHandler(dp.id, event.MouseClick, func(e *event.Event) {
		dp.open = !dp.open
	})
	return dp
}

func (dp *DatePicker) Year() int                        { return dp.year }
func (dp *DatePicker) Month() int                       { return dp.month }
func (dp *DatePicker) Day() int                         { return dp.day }
func (dp *DatePicker) IsOpen() bool                     { return dp.open }
func (dp *DatePicker) SetDate(y, m, d int)              { dp.year = y; dp.month = m; dp.day = d }
func (dp *DatePicker) SetOpen(o bool)                   { dp.open = o }
func (dp *DatePicker) OnChange(fn func(int, int, int))  { dp.onChange = fn }

func (dp *DatePicker) daysInMonth() int {
	switch dp.month {
	case 2:
		if dp.year%4 == 0 && (dp.year%100 != 0 || dp.year%400 == 0) {
			return 29
		}
		return 28
	case 4, 6, 9, 11:
		return 30
	default:
		return 31
	}
}

// dayOfWeek returns 0=Sun..6=Sat for first day of current month (Zeller's).
func (dp *DatePicker) firstDayOfWeek() int {
	y := dp.year
	m := dp.month
	if m < 3 {
		m += 12
		y--
	}
	return (1 + (13*(m+1))/5 + y + y/4 - y/100 + y/400) % 7
}

func (dp *DatePicker) Draw(buf *render.CommandBuffer) {
	bounds := dp.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := dp.config

	// Input field
	label := dateStr(dp.year, dp.month, dp.day)
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

	// Calendar dropdown
	if !dp.open {
		return
	}
	cellSize := float32(32)
	calW := cellSize * 7
	calH := cellSize*7 + 36 // header + 6 rows
	cx := bounds.X
	cy := bounds.Y + bounds.Height + 4

	// Shadow + background
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(cx+2, cy+2, calW, calH),
		FillColor: uimath.RGBA(0, 0, 0, 0.1),
		Corners:   uimath.CornersAll(cfg.BorderRadius),
	}, 40, 1)
	buf.DrawOverlay(render.RectCmd{
		Bounds:      uimath.NewRect(cx, cy, calW, calH),
		FillColor:   uimath.ColorWhite,
		BorderColor: cfg.BorderColor,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 41, 1)

	// Month/year header
	if cfg.TextRenderer != nil {
		header := monthName(dp.month) + " " + intToStr(dp.year)
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, header, cx+cfg.SpaceSM, cy+(36-lh)/2, cfg.FontSize, calW-cfg.SpaceSM*2, cfg.TextColor, 1)
	}

	// Day grid
	startDay := dp.firstDayOfWeek()
	daysInM := dp.daysInMonth()
	row := 0
	col := startDay
	for d := 1; d <= daysInM; d++ {
		dx := cx + float32(col)*cellSize
		dy := cy + 36 + float32(row)*cellSize
		if d == dp.day {
			buf.DrawOverlay(render.RectCmd{
				Bounds:    uimath.NewRect(dx+2, dy+2, cellSize-4, cellSize-4),
				FillColor: cfg.PrimaryColor,
				Corners:   uimath.CornersAll(cellSize / 2),
			}, 42, 1)
		}
		if cfg.TextRenderer != nil {
			txt := intToStr(d)
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			tw := cfg.TextRenderer.MeasureText(txt, cfg.FontSizeSm)
			color := cfg.TextColor
			if d == dp.day {
				color = uimath.ColorWhite
			}
			cfg.TextRenderer.DrawText(buf, txt, dx+(cellSize-tw)/2, dy+(cellSize-lh)/2, cfg.FontSizeSm, cellSize, color, 1)
		}
		col++
		if col > 6 {
			col = 0
			row++
		}
	}
}

// TimePicker provides a time selection interface.
type TimePicker struct {
	Base
	hour     int
	minute   int
	second   int
	showSec  bool
	open     bool
	onChange func(h, m, s int)
}

func NewTimePicker(tree *core.Tree, cfg *Config) *TimePicker {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	tp := &TimePicker{
		Base: NewBase(tree, core.TypeCustom, cfg),
	}
	tree.AddHandler(tp.id, event.MouseClick, func(e *event.Event) {
		tp.open = !tp.open
	})
	return tp
}

func (tp *TimePicker) Hour() int   { return tp.hour }
func (tp *TimePicker) Minute() int { return tp.minute }
func (tp *TimePicker) Second() int { return tp.second }
func (tp *TimePicker) IsOpen() bool { return tp.open }
func (tp *TimePicker) SetTime(h, m, s int) {
	tp.hour = h
	tp.minute = m
	tp.second = s
}
func (tp *TimePicker) SetShowSeconds(s bool)           { tp.showSec = s }
func (tp *TimePicker) SetOpen(o bool)                  { tp.open = o }
func (tp *TimePicker) OnChange(fn func(int, int, int)) { tp.onChange = fn }

func (tp *TimePicker) Draw(buf *render.CommandBuffer) {
	bounds := tp.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := tp.config

	label := pad2(tp.hour) + ":" + pad2(tp.minute)
	if tp.showSec {
		label += ":" + pad2(tp.second)
	}

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

// helpers

func dateStr(y, m, d int) string {
	return intToStr(y) + "-" + pad2(m) + "-" + pad2(d)
}

func pad2(n int) string {
	if n < 10 {
		return "0" + intToStr(n)
	}
	return intToStr(n)
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append(buf, byte('0'+n%10))
		n /= 10
	}
	if neg {
		buf = append(buf, '-')
	}
	for l, r := 0, len(buf)-1; l < r; l, r = l+1, r-1 {
		buf[l], buf[r] = buf[r], buf[l]
	}
	return string(buf)
}

var monthNames = [12]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

func monthName(m int) string {
	if m >= 1 && m <= 12 {
		return monthNames[m-1]
	}
	return "???"
}

package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Calendar displays a full month calendar grid.
type Calendar struct {
	Base
	year      int
	month     int
	selected  int // day
	cellSize  float32
	onSelect  func(year, month, day int)
}

func NewCalendar(tree *core.Tree, cfg *Config) *Calendar {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Calendar{
		Base:     NewBase(tree, core.TypeCustom, cfg),
		year:     2026,
		month:    1,
		selected: 1,
		cellSize: 36,
	}
}

func (c *Calendar) Year() int                          { return c.year }
func (c *Calendar) Month() int                         { return c.month }
func (c *Calendar) Selected() int                      { return c.selected }
func (c *Calendar) SetYear(y int)                      { c.year = y }
func (c *Calendar) SetMonth(m int)                     { c.month = m }
func (c *Calendar) SetSelected(d int)                  { c.selected = d }
func (c *Calendar) OnSelect(fn func(int, int, int))    { c.onSelect = fn }

func (c *Calendar) NextMonth() {
	c.month++
	if c.month > 12 {
		c.month = 1
		c.year++
	}
}

func (c *Calendar) PrevMonth() {
	c.month--
	if c.month < 1 {
		c.month = 12
		c.year--
	}
}

func (c *Calendar) daysInMonth() int {
	switch c.month {
	case 2:
		if c.year%4 == 0 && (c.year%100 != 0 || c.year%400 == 0) {
			return 29
		}
		return 28
	case 4, 6, 9, 11:
		return 30
	default:
		return 31
	}
}

func (c *Calendar) firstDayOfWeek() int {
	y := c.year
	m := c.month
	if m < 3 {
		m += 12
		y--
	}
	return (1 + (13*(m+1))/5 + y + y/4 - y/100 + y/400) % 7
}

func (c *Calendar) Draw(buf *render.CommandBuffer) {
	bounds := c.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := c.config
	cs := c.cellSize
	headerH := float32(40)

	// Header: month/year
	if cfg.TextRenderer != nil {
		header := monthName(c.month) + " " + intToStr(c.year)
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, header, bounds.X+cfg.SpaceSM, bounds.Y+(headerH-lh)/2, cfg.FontSize, bounds.Width-cfg.SpaceSM*2, cfg.TextColor, 1)
	}

	// Weekday headers
	days := [7]string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	for i, d := range days {
		dx := bounds.X + float32(i)*cs
		if cfg.TextRenderer != nil {
			tw := cfg.TextRenderer.MeasureText(d, cfg.FontSizeSm)
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			cfg.TextRenderer.DrawText(buf, d, dx+(cs-tw)/2, bounds.Y+headerH+(cs-lh)/2, cfg.FontSizeSm, cs, cfg.DisabledColor, 1)
		}
	}

	// Days grid
	startDay := c.firstDayOfWeek()
	daysInM := c.daysInMonth()
	row := 0
	col := startDay
	gridY := bounds.Y + headerH + cs
	for d := 1; d <= daysInM; d++ {
		dx := bounds.X + float32(col)*cs
		dy := gridY + float32(row)*cs

		if d == c.selected {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(dx+2, dy+2, cs-4, cs-4),
				FillColor: cfg.PrimaryColor,
				Corners:   uimath.CornersAll(cs / 2),
			}, 2, 1)
		}

		if cfg.TextRenderer != nil {
			txt := intToStr(d)
			tw := cfg.TextRenderer.MeasureText(txt, cfg.FontSizeSm)
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			color := cfg.TextColor
			if d == c.selected {
				color = uimath.ColorWhite
			}
			cfg.TextRenderer.DrawText(buf, txt, dx+(cs-tw)/2, dy+(cs-lh)/2, cfg.FontSizeSm, cs, color, 1)
		}

		col++
		if col > 6 {
			col = 0
			row++
		}
	}
}

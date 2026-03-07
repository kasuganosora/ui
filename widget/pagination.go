package widget

import (
	"strconv"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Pagination displays page navigation buttons.
type Pagination struct {
	Base
	current  int
	total    int
	pageSize int
	onChange func(page int)
}

func NewPagination(tree *core.Tree, cfg *Config) *Pagination {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	p := &Pagination{
		Base:     NewBase(tree, core.TypeCustom, cfg),
		current:  1,
		total:    0,
		pageSize: 10,
	}
	tree.AddHandler(p.id, event.MouseClick, func(e *event.Event) {})
	return p
}

func (p *Pagination) Current() int        { return p.current }
func (p *Pagination) Total() int          { return p.total }
func (p *Pagination) PageSize() int       { return p.pageSize }
func (p *Pagination) SetCurrent(c int)    { p.current = c }
func (p *Pagination) SetTotal(t int)      { p.total = t }
func (p *Pagination) SetPageSize(s int)   { p.pageSize = s }
func (p *Pagination) OnChange(fn func(int)) { p.onChange = fn }

func (p *Pagination) TotalPages() int {
	if p.pageSize <= 0 {
		return 0
	}
	return (p.total + p.pageSize - 1) / p.pageSize
}

func (p *Pagination) GoTo(page int) {
	total := p.TotalPages()
	if page < 1 {
		page = 1
	}
	if page > total {
		page = total
	}
	if page != p.current {
		p.current = page
		if p.onChange != nil {
			p.onChange(page)
		}
	}
}

func (p *Pagination) Draw(buf *render.CommandBuffer) {
	bounds := p.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := p.config
	totalPages := p.TotalPages()
	if totalPages <= 0 {
		return
	}

	btnSize := float32(32)
	gap := float32(4)
	x := bounds.X

	// Prev button
	p.drawPageBtn(buf, x, bounds.Y, btnSize, "<", p.current > 1, false, cfg)
	x += btnSize + gap

	// Page numbers (show up to 7)
	start, end := 1, totalPages
	if totalPages > 7 {
		start = p.current - 3
		end = p.current + 3
		if start < 1 {
			start = 1
			end = 7
		}
		if end > totalPages {
			end = totalPages
			start = totalPages - 6
		}
	}
	for i := start; i <= end; i++ {
		p.drawPageBtn(buf, x, bounds.Y, btnSize, strconv.Itoa(i), true, i == p.current, cfg)
		x += btnSize + gap
	}

	// Next button
	p.drawPageBtn(buf, x, bounds.Y, btnSize, ">", p.current < totalPages, false, cfg)
}

func (p *Pagination) drawPageBtn(buf *render.CommandBuffer, x, y, size float32, label string, enabled, active bool, cfg *Config) {
	bg := uimath.ColorWhite
	borderClr := cfg.BorderColor
	textClr := cfg.TextColor
	if active {
		bg = cfg.PrimaryColor
		borderClr = cfg.PrimaryColor
		textClr = uimath.ColorWhite
	}
	if !enabled {
		textClr = cfg.DisabledColor
	}

	buf.DrawRect(render.RectCmd{
		Bounds:      uimath.NewRect(x, y, size, size),
		FillColor:   bg,
		BorderColor: borderClr,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 0, 1)

	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		tw := cfg.TextRenderer.MeasureText(label, cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, label, x+(size-tw)/2, y+(size-lh)/2, cfg.FontSize, size, textClr, 1)
	} else {
		tw := float32(len(label)) * cfg.FontSize * 0.55
		th := cfg.FontSize * 1.2
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x+(size-tw)/2, y+(size-th)/2, tw, th),
			FillColor: textClr,
			Corners:   uimath.CornersAll(2),
		}, 1, 1)
	}
}

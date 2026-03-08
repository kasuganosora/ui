package widget

import (
	"strconv"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// PaginationTheme controls the pagination visual style (TDesign: theme).
type PaginationTheme int

const (
	PaginationThemeDefault PaginationTheme = iota
	PaginationThemeSimple
)

// PageEllipsisMode controls how ellipsis is shown (TDesign: pageEllipsisMode).
type PageEllipsisMode int

const (
	PageEllipsisModeMid      PageEllipsisMode = iota // default
	PageEllipsisModeBothEnds
)

// PaginationPageInfo is the argument to pagination callbacks (TDesign: PageInfo).
type PaginationPageInfo struct {
	Current  int
	Previous int
	PageSize int
}

// Pagination displays page navigation buttons (TDesign: TdPaginationProps).
type Pagination struct {
	Base
	current                int
	total                  int
	pageSize               int
	size                   Size
	theme                  PaginationTheme  // TDesign: theme
	disabled               bool             // TDesign: disabled
	foldedMaxPageBtn       int              // TDesign: foldedMaxPageBtn
	maxPageBtn             int              // TDesign: maxPageBtn
	pageEllipsisMode       PageEllipsisMode // TDesign: pageEllipsisMode
	showFirstAndLastPageBtn bool            // TDesign: showFirstAndLastPageBtn
	showJumper             bool             // TDesign: showJumper
	showPageNumber         bool             // TDesign: showPageNumber
	showPageSize           bool             // TDesign: showPageSize
	showPreviousAndNextBtn bool             // TDesign: showPreviousAndNextBtn
	totalContent           bool             // TDesign: totalContent (show total count)
	onChange               func(pageInfo PaginationPageInfo) // TDesign: onChange
	onCurrentChange        func(current int, pageInfo PaginationPageInfo) // TDesign: onCurrentChange
	onPageSizeChange       func(pageSize int, pageInfo PaginationPageInfo) // TDesign: onPageSizeChange

	// Clickable child elements for hit testing
	prevID    core.ElementID
	nextID    core.ElementID
	pageIDs   []core.ElementID // one per displayed page button
	lastPages int              // track when page count changes to rebuild
}

func NewPagination(tree *core.Tree, cfg *Config) *Pagination {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	p := &Pagination{
		Base:                   NewBase(tree, core.TypeCustom, cfg),
		current:                1,
		total:                  0,
		pageSize:               10,
		size:                   SizeMedium,
		foldedMaxPageBtn:       5,
		maxPageBtn:             10,
		showPageNumber:         true,
		showPageSize:           true,
		showPreviousAndNextBtn: true,
		totalContent:           true,
	}

	// Create prev/next button elements
	p.prevID = tree.CreateElement(core.TypeCustom)
	tree.AppendChild(p.id, p.prevID)
	tree.AddHandler(p.prevID, event.MouseClick, func(e *event.Event) {
		if p.disabled {
			return
		}
		if p.current > 1 {
			p.GoTo(p.current - 1)
			p.tree.MarkDirty(p.id)
		}
	})

	p.nextID = tree.CreateElement(core.TypeCustom)
	tree.AppendChild(p.id, p.nextID)
	tree.AddHandler(p.nextID, event.MouseClick, func(e *event.Event) {
		if p.disabled {
			return
		}
		if p.current < p.TotalPages() {
			p.GoTo(p.current + 1)
			p.tree.MarkDirty(p.id)
		}
	})

	return p
}

func (p *Pagination) Current() int                 { return p.current }
func (p *Pagination) Total() int                   { return p.total }
func (p *Pagination) PageSize() int                { return p.pageSize }
func (p *Pagination) SetCurrent(c int)             { p.current = c }
func (p *Pagination) SetTotal(t int)               { p.total = t }
func (p *Pagination) SetPageSize(s int)            { p.pageSize = s }
func (p *Pagination) SetSize(s Size)               { p.size = s }
func (p *Pagination) SetTheme(t PaginationTheme)   { p.theme = t }
func (p *Pagination) SetDisabled(d bool)           { p.disabled = d }
func (p *Pagination) SetFoldedMaxPageBtn(n int)    { p.foldedMaxPageBtn = n }
func (p *Pagination) SetMaxPageBtn(n int)          { p.maxPageBtn = n }
func (p *Pagination) SetPageEllipsisMode(m PageEllipsisMode) { p.pageEllipsisMode = m }
func (p *Pagination) SetShowFirstAndLastPageBtn(v bool)      { p.showFirstAndLastPageBtn = v }
func (p *Pagination) SetShowJumper(v bool)                   { p.showJumper = v }
func (p *Pagination) SetShowPageNumber(v bool)               { p.showPageNumber = v }
func (p *Pagination) SetShowPageSize(v bool)                 { p.showPageSize = v }
func (p *Pagination) SetShowPreviousAndNextBtn(v bool)       { p.showPreviousAndNextBtn = v }
func (p *Pagination) SetTotalContent(v bool)                 { p.totalContent = v }

// OnChange sets the callback for any page/size change (TDesign: onChange).
func (p *Pagination) OnChange(fn func(pageInfo PaginationPageInfo)) { p.onChange = fn }

// OnCurrentChange sets the callback for current page change (TDesign: onCurrentChange).
func (p *Pagination) OnCurrentChange(fn func(current int, pageInfo PaginationPageInfo)) {
	p.onCurrentChange = fn
}

// OnPageSizeChange sets the callback for page size change (TDesign: onPageSizeChange).
func (p *Pagination) OnPageSizeChange(fn func(pageSize int, pageInfo PaginationPageInfo)) {
	p.onPageSizeChange = fn
}

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
		prev := p.current
		p.current = page
		pi := PaginationPageInfo{
			Current:  p.current,
			Previous: prev,
			PageSize: p.pageSize,
		}
		if p.onChange != nil {
			p.onChange(pi)
		}
		if p.onCurrentChange != nil {
			p.onCurrentChange(p.current, pi)
		}
	}
}

// rebuildPageElements destroys old page elements and creates new ones.
func (p *Pagination) rebuildPageElements(count int) {
	// Destroy old page elements
	for _, pid := range p.pageIDs {
		p.tree.DestroyElement(pid)
	}
	p.pageIDs = nil

	for i := 0; i < count; i++ {
		pid := p.tree.CreateElement(core.TypeCustom)
		p.tree.AppendChild(p.id, pid)
		p.tree.AddHandler(pid, event.MouseClick, func(e *event.Event) {
			if p.disabled {
				return
			}
			// Page number is stored as property on the element
			if v, ok := p.tree.Get(pid).Property("page"); ok {
				if pg, ok := v.(int); ok {
					p.GoTo(pg)
					p.tree.MarkDirty(p.id)
				}
			}
		})
		p.pageIDs = append(p.pageIDs, pid)
	}
	p.lastPages = count
}

// pageRange computes the displayed page buttons with ellipsis gaps.
// Returns a slice of page numbers; 0 means ellipsis.
func (p *Pagination) pageRange(totalPages int) []int {
	if totalPages <= 7 {
		pages := make([]int, totalPages)
		for i := range pages {
			pages[i] = i + 1
		}
		return pages
	}

	var pages []int
	cur := p.current

	// Always show first page
	pages = append(pages, 1)

	// Left ellipsis?
	if cur > 4 {
		pages = append(pages, 0) // ellipsis
	}

	// Middle range
	start := cur - 2
	end := cur + 2
	if start < 2 {
		start = 2
	}
	if end > totalPages-1 {
		end = totalPages - 1
	}
	// Expand range if near edges
	if cur <= 4 {
		end = 5
		if end > totalPages-1 {
			end = totalPages - 1
		}
	}
	if cur >= totalPages-3 {
		start = totalPages - 4
		if start < 2 {
			start = 2
		}
	}
	for i := start; i <= end; i++ {
		pages = append(pages, i)
	}

	// Right ellipsis?
	if cur < totalPages-3 {
		pages = append(pages, 0) // ellipsis
	}

	// Always show last page
	pages = append(pages, totalPages)

	return pages
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

	btnSize := cfg.SizeHeight(p.size)
	fontSize := cfg.SizeFontSize(p.size)
	gap := float32(4)
	x := bounds.X

	// Show total count
	if p.totalContent {
		totalText := "共 " + strconv.Itoa(p.total) + " 条"
		totalW := float32(len(totalText)) * fontSize * 0.55
		if cfg.TextRenderer != nil {
			totalW = cfg.TextRenderer.MeasureText(totalText, fontSize)
		}
		p.drawText(buf, totalText, x, bounds.Y, totalW, btnSize, fontSize, cfg.TextColor, cfg)
		x += totalW + gap*2
	}

	// Prev button
	p.tree.SetLayout(p.prevID, core.LayoutResult{
		Bounds: uimath.NewRect(x, bounds.Y, btnSize, btnSize),
	})
	p.drawPageBtn(buf, x, bounds.Y, btnSize, "<", p.current > 1 && !p.disabled, false, fontSize, cfg)
	x += btnSize + gap

	// Compute page range with ellipsis
	pages := p.pageRange(totalPages)

	// Rebuild page elements if count changed
	if len(pages) != p.lastPages {
		p.rebuildPageElements(len(pages))
	}

	for i, pg := range pages {
		if pg == 0 {
			// Ellipsis
			p.drawText(buf, "...", x, bounds.Y, btnSize, btnSize, fontSize, cfg.TextColor, cfg)
			// Set layout for the element even for ellipsis (non-clickable but consistent)
			if i < len(p.pageIDs) {
				p.tree.SetLayout(p.pageIDs[i], core.LayoutResult{
					Bounds: uimath.NewRect(x, bounds.Y, btnSize, btnSize),
				})
				p.tree.SetProperty(p.pageIDs[i], "page", 0)
			}
		} else {
			if i < len(p.pageIDs) {
				p.tree.SetLayout(p.pageIDs[i], core.LayoutResult{
					Bounds: uimath.NewRect(x, bounds.Y, btnSize, btnSize),
				})
				p.tree.SetProperty(p.pageIDs[i], "page", pg)
			}
			p.drawPageBtn(buf, x, bounds.Y, btnSize, strconv.Itoa(pg), !p.disabled, pg == p.current, fontSize, cfg)
		}
		x += btnSize + gap
	}

	// Next button
	p.tree.SetLayout(p.nextID, core.LayoutResult{
		Bounds: uimath.NewRect(x, bounds.Y, btnSize, btnSize),
	})
	p.drawPageBtn(buf, x, bounds.Y, btnSize, ">", p.current < totalPages && !p.disabled, false, fontSize, cfg)
}

func (p *Pagination) drawText(buf *render.CommandBuffer, text string, x, y, w, h, fontSize float32, clr uimath.Color, cfg *Config) {
	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(fontSize)
		tw := cfg.TextRenderer.MeasureText(text, fontSize)
		cfg.TextRenderer.DrawText(buf, text, x+(w-tw)/2, y+(h-lh)/2, fontSize, w, clr, 1)
	} else {
		tw := float32(len(text)) * fontSize * 0.55
		th := fontSize * 1.2
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x+(w-tw)/2, y+(h-th)/2, tw, th),
			FillColor: clr,
			Corners:   uimath.CornersAll(2),
		}, 1, 1)
	}
}

func (p *Pagination) drawPageBtn(buf *render.CommandBuffer, x, y, size float32, label string, enabled, active bool, fontSize float32, cfg *Config) {
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

	p.drawText(buf, label, x, y, size, size, fontSize, textClr, cfg)
}

// Destroy cleans up child elements.
func (p *Pagination) Destroy() {
	p.tree.DestroyElement(p.prevID)
	p.tree.DestroyElement(p.nextID)
	for _, pid := range p.pageIDs {
		p.tree.DestroyElement(pid)
	}
	p.Base.Destroy()
}

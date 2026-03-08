package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// ListItem represents a single list entry.
type ListItem struct {
	Title       string
	Description string
	Extra       string
}

// ListLayout controls whether items are arranged horizontally or vertically.
type ListLayout uint8

const (
	ListHorizontal ListLayout = iota
	ListVertical
)

// List displays a scrollable list of items.
type List struct {
	Base
	items        []ListItem
	itemHeight   float32
	scrollY      float32
	bordered     bool
	split        bool
	stripe       bool
	size         Size
	hoverable    bool
	hoveredIndex int
	layout       ListLayout
	asyncLoading string
	header       string
	footer       string
	onSelect     func(index int)
	onLoadMore   func()
	onScroll     func(scrollTop, scrollBottom float32)
}

func NewList(tree *core.Tree, cfg *Config) *List {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	l := &List{
		Base:         NewBase(tree, core.TypeDiv, cfg),
		itemHeight:   48,
		bordered:     true,
		split:        true,
		hoverable:    true,
		hoveredIndex: -1,
		size:         SizeMedium,
	}
	l.style.Display = layout.DisplayFlex
	l.style.FlexDirection = layout.FlexDirectionColumn

	// Track hover
	tree.AddHandler(l.id, event.MouseMove, func(e *event.Event) {
		if !l.hoverable {
			return
		}
		bounds := l.Bounds()
		relY := e.Y - bounds.Y + l.scrollY
		idx := int(relY / l.itemHeight)
		if idx >= 0 && idx < len(l.items) {
			l.hoveredIndex = idx
		} else {
			l.hoveredIndex = -1
		}
	})
	tree.AddHandler(l.id, event.MouseLeave, func(e *event.Event) {
		l.hoveredIndex = -1
	})

	return l
}

func (l *List) SetItems(items []ListItem) { l.items = items }
func (l *List) Items() []ListItem         { return l.items }
func (l *List) SetItemHeight(h float32)   { l.itemHeight = h }
func (l *List) SetBordered(b bool)        { l.bordered = b }
func (l *List) OnSelect(fn func(int))     { l.onSelect = fn }
func (l *List) ScrollY() float32          { return l.scrollY }
func (l *List) SetScrollY(y float32)      { l.scrollY = y }
func (l *List) SetSplit(v bool)                                    { l.split = v }
func (l *List) SetStripe(v bool)                                   { l.stripe = v }
func (l *List) SetHoverable(v bool)                                { l.hoverable = v }
func (l *List) SetLayout(v ListLayout)                             { l.layout = v }
func (l *List) SetAsyncLoading(v string)                           { l.asyncLoading = v }
func (l *List) SetHeader(v string)                                 { l.header = v }
func (l *List) SetFooter(v string)                                 { l.footer = v }
func (l *List) OnLoadMore(fn func())                               { l.onLoadMore = fn }
func (l *List) OnScroll(fn func(scrollTop, scrollBottom float32))  { l.onScroll = fn }

func (l *List) SetSize(s Size) {
	l.size = s
	switch s {
	case SizeSmall:
		l.itemHeight = 36
	case SizeLarge:
		l.itemHeight = 64
	default:
		l.itemHeight = 48
	}
}

func (l *List) sizeItemHeight() float32 {
	return l.itemHeight
}

func (l *List) Draw(buf *render.CommandBuffer) {
	bounds := l.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := l.config

	if l.bordered {
		buf.DrawRect(render.RectCmd{
			Bounds:      bounds,
			BorderColor: cfg.BorderColor,
			BorderWidth: cfg.BorderWidth,
			Corners:     uimath.CornersAll(cfg.BorderRadius),
		}, 0, 1)
	}

	buf.PushClip(bounds)
	y := bounds.Y - l.scrollY
	ih := l.sizeItemHeight()
	for i, item := range l.items {
		if y+ih < bounds.Y {
			y += ih
			continue
		}
		if y > bounds.Y+bounds.Height {
			break
		}

		// Stripe background
		if l.stripe && i%2 == 1 {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X, y, bounds.Width, ih),
				FillColor: uimath.RGBA(0, 0, 0, 0.02),
			}, 1, 1)
		}

		// Hover highlight
		if l.hoverable && i == l.hoveredIndex {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X, y, bounds.Width, ih),
				FillColor: uimath.RGBA(0, 0, 0, 0.04),
			}, 1, 1)
		}

		// Divider between items
		if l.split && i < len(l.items)-1 {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X, y+ih-1, bounds.Width, 1),
				FillColor: uimath.RGBA(0, 0, 0, 0.06),
			}, 2, 1)
		}

		// Calculate text areas
		leftPad := bounds.X + cfg.SpaceMD
		rightPad := cfg.SpaceMD
		extraW := float32(0)

		// Measure extra text width
		if item.Extra != "" {
			if cfg.TextRenderer != nil {
				extraW = cfg.TextRenderer.MeasureText(item.Extra, cfg.FontSizeSm) + cfg.SpaceMD
			} else {
				extraW = float32(len(item.Extra))*cfg.FontSizeSm*0.55 + cfg.SpaceMD
			}
		}

		titleMaxW := bounds.Width - cfg.SpaceMD*2 - extraW

		hasDesc := item.Description != ""

		// Title
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			var titleY float32
			if hasDesc {
				// Title in upper portion
				titleY = y + cfg.SpaceXS
			} else {
				titleY = y + (ih-lh)/2
			}
			cfg.TextRenderer.DrawText(buf, item.Title, leftPad, titleY, cfg.FontSize, titleMaxW, cfg.TextColor, 1)
		} else {
			tw := float32(len(item.Title)) * cfg.FontSize * 0.55
			if tw > titleMaxW {
				tw = titleMaxW
			}
			th := cfg.FontSize * 1.2
			var titleY float32
			if hasDesc {
				titleY = y + cfg.SpaceXS
			} else {
				titleY = y + (ih-th)/2
			}
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(leftPad, titleY, tw, th),
				FillColor: cfg.TextColor,
				Corners:   uimath.CornersAll(2),
			}, 3, 1)
		}

		// Description below title
		if hasDesc {
			descY := y + ih/2
			descMaxW := titleMaxW
			if cfg.TextRenderer != nil {
				descColor := uimath.RGBA(0, 0, 0, 0.45)
				cfg.TextRenderer.DrawText(buf, item.Description, leftPad, descY, cfg.FontSizeSm, descMaxW, descColor, 1)
			} else {
				dw := float32(len(item.Description)) * cfg.FontSizeSm * 0.55
				if dw > descMaxW {
					dw = descMaxW
				}
				dh := cfg.FontSizeSm * 1.2
				buf.DrawRect(render.RectCmd{
					Bounds:    uimath.NewRect(leftPad, descY, dw, dh),
					FillColor: uimath.RGBA(0, 0, 0, 0.25),
					Corners:   uimath.CornersAll(2),
				}, 3, 1)
			}
		}

		// Extra text on the right
		if item.Extra != "" {
			extraX := bounds.X + bounds.Width - rightPad
			if cfg.TextRenderer != nil {
				ew := cfg.TextRenderer.MeasureText(item.Extra, cfg.FontSizeSm)
				lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
				extraColor := uimath.RGBA(0, 0, 0, 0.45)
				cfg.TextRenderer.DrawText(buf, item.Extra, extraX-ew, y+(ih-lh)/2, cfg.FontSizeSm, ew, extraColor, 1)
			} else {
				ew := float32(len(item.Extra)) * cfg.FontSizeSm * 0.55
				eh := cfg.FontSizeSm * 1.2
				buf.DrawRect(render.RectCmd{
					Bounds:    uimath.NewRect(extraX-ew, y+(ih-eh)/2, ew, eh),
					FillColor: uimath.RGBA(0, 0, 0, 0.25),
					Corners:   uimath.CornersAll(2),
				}, 3, 1)
			}
		}

		y += ih
	}
	buf.PopClip()
}

// VirtualList renders only visible items for large datasets.
type VirtualList struct {
	Base
	itemCount  int
	itemHeight float32
	scrollY    float32
	renderItem func(index int, buf *render.CommandBuffer, x, y, w, h float32)
}

func NewVirtualList(tree *core.Tree, cfg *Config) *VirtualList {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &VirtualList{
		Base:       NewBase(tree, core.TypeDiv, cfg),
		itemHeight: 40,
	}
}

func (vl *VirtualList) SetItemCount(n int)     { vl.itemCount = n }
func (vl *VirtualList) SetItemHeight(h float32) { vl.itemHeight = h }
func (vl *VirtualList) SetScrollY(y float32)    { vl.scrollY = y }
func (vl *VirtualList) ScrollY() float32        { return vl.scrollY }
func (vl *VirtualList) SetRenderItem(fn func(int, *render.CommandBuffer, float32, float32, float32, float32)) {
	vl.renderItem = fn
}

func (vl *VirtualList) ContentHeight() float32 {
	return float32(vl.itemCount) * vl.itemHeight
}

func (vl *VirtualList) Draw(buf *render.CommandBuffer) {
	bounds := vl.Bounds()
	if bounds.IsEmpty() || vl.renderItem == nil {
		return
	}
	buf.PushClip(bounds)

	startIdx := int(vl.scrollY / vl.itemHeight)
	if startIdx < 0 {
		startIdx = 0
	}
	endIdx := int((vl.scrollY + bounds.Height) / vl.itemHeight) + 1
	if endIdx > vl.itemCount {
		endIdx = vl.itemCount
	}

	for i := startIdx; i < endIdx; i++ {
		y := bounds.Y + float32(i)*vl.itemHeight - vl.scrollY
		vl.renderItem(i, buf, bounds.X, y, bounds.Width, vl.itemHeight)
	}
	buf.PopClip()
}

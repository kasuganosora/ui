package widget

import (
	"github.com/kasuganosora/ui/core"
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

// List displays a scrollable list of items.
type List struct {
	Base
	items      []ListItem
	itemHeight float32
	scrollY    float32
	bordered   bool
	onSelect   func(index int)
}

func NewList(tree *core.Tree, cfg *Config) *List {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	l := &List{
		Base:       NewBase(tree, core.TypeDiv, cfg),
		itemHeight: 48,
		bordered:   true,
	}
	l.style.Display = layout.DisplayFlex
	l.style.FlexDirection = layout.FlexDirectionColumn
	return l
}

func (l *List) SetItems(items []ListItem) { l.items = items }
func (l *List) Items() []ListItem         { return l.items }
func (l *List) SetItemHeight(h float32)   { l.itemHeight = h }
func (l *List) SetBordered(b bool)        { l.bordered = b }
func (l *List) OnSelect(fn func(int))     { l.onSelect = fn }
func (l *List) ScrollY() float32          { return l.scrollY }
func (l *List) SetScrollY(y float32)      { l.scrollY = y }

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
	for _, item := range l.items {
		if y+l.itemHeight < bounds.Y {
			y += l.itemHeight
			continue
		}
		if y > bounds.Y+bounds.Height {
			break
		}
		// Divider
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(bounds.X, y+l.itemHeight-1, bounds.Width, 1),
			FillColor: uimath.RGBA(0, 0, 0, 0.06),
		}, 1, 1)
		// Title
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			cfg.TextRenderer.DrawText(buf, item.Title, bounds.X+cfg.SpaceMD, y+(l.itemHeight-lh)/2, cfg.FontSize, bounds.Width-cfg.SpaceMD*2, cfg.TextColor, 1)
		} else {
			tw := float32(len(item.Title)) * cfg.FontSize * 0.55
			th := cfg.FontSize * 1.2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X+cfg.SpaceMD, y+(l.itemHeight-th)/2, tw, th),
				FillColor: cfg.TextColor,
				Corners:   uimath.CornersAll(2),
			}, 1, 1)
		}
		y += l.itemHeight
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

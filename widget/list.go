package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// ListItem represents a single list entry.
// Supports text-only, multi-line (title+description), avatar, and actions.
type ListItem struct {
	Value       string   // unique identifier (used with selectable mode)
	Group       string   // group header text (items with same Group are grouped together)
	Title       string
	Description string
	Extra       string   // single extra text on the right (deprecated, use Actions)
	Actions     []string // action text links on the right (e.g. "操作1", "操作2")
	Avatar      string   // avatar text/initials (rendered as circle)
	AvatarColor uimath.Color // avatar background color (zero = default)
	Image       render.TextureHandle // avatar image texture
}

// ListLayout controls whether items are arranged horizontally or vertically.
type ListLayout uint8

const (
	ListHorizontal ListLayout = iota
	ListVertical
)

// List displays a scrollable list of items.
// Supports text-only, multi-line, avatar+text, and action items.
// When selectable is true, items can be clicked to select, with active highlight.
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
	selectable    bool   // enable selection mode (nav list)
	selectedValue string // active item value (selectable mode)
	groupHeight   float32 // height of group header rows
	onSelect     func(index int)
	onChange     func(value string) // called when selected value changes (selectable mode)
	onAction     func(itemIndex, actionIndex int)
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
	l.groupHeight = 28

	// Track hover
	tree.AddHandler(l.id, event.MouseMove, func(e *event.Event) {
		if !l.hoverable {
			return
		}
		bounds := l.Bounds()
		idx := l.hitTestItem(e.Y - bounds.Y + l.scrollY)
		if idx != l.hoveredIndex {
			l.hoveredIndex = idx
			l.tree.MarkDirty(l.id)
		}
	})
	tree.AddHandler(l.id, event.MouseLeave, func(e *event.Event) {
		if l.hoveredIndex != -1 {
			l.hoveredIndex = -1
			l.tree.MarkDirty(l.id)
		}
	})

	// Click handler for item selection
	tree.AddHandler(l.id, event.MouseClick, func(e *event.Event) {
		bounds := l.Bounds()
		idx := l.hitTestItem(e.Y - bounds.Y + l.scrollY)
		if idx >= 0 && idx < len(l.items) {
			if l.selectable && l.items[idx].Value != "" {
				l.selectedValue = l.items[idx].Value
				if l.onChange != nil {
					l.onChange(l.selectedValue)
				}
				l.tree.MarkDirty(l.id)
			}
			if l.onSelect != nil {
				l.onSelect(idx)
			}
		}
	})

	return l
}

func (l *List) SetItems(items []ListItem) {
	l.items = items
	l.updatePreferredHeight()
}
func (l *List) Items() []ListItem         { return l.items }
func (l *List) SetItemHeight(h float32)   { l.itemHeight = h }
func (l *List) SetBordered(b bool)        { l.bordered = b }
func (l *List) OnSelect(fn func(int))     { l.onSelect = fn }
func (l *List) ScrollY() float32          { return l.scrollY }
func (l *List) SetScrollY(y float32)      { l.scrollY = y }
func (l *List) SetSplit(v bool)           { l.split = v }
func (l *List) SetStripe(v bool)          { l.stripe = v }
func (l *List) SetHoverable(v bool)       { l.hoverable = v }
func (l *List) SetLayout(v ListLayout)    { l.layout = v }
func (l *List) SetAsyncLoading(v string)  { l.asyncLoading = v }
func (l *List) SetHeader(v string)        { l.header = v }
func (l *List) SetFooter(v string)        { l.footer = v }
func (l *List) OnLoadMore(fn func())      { l.onLoadMore = fn }
func (l *List) OnAction(fn func(itemIndex, actionIndex int)) { l.onAction = fn }
func (l *List) OnScroll(fn func(scrollTop, scrollBottom float32)) { l.onScroll = fn }
func (l *List) SetSelectable(v bool)       { l.selectable = v }
func (l *List) IsSelectable() bool         { return l.selectable }
func (l *List) SetSelectedValue(v string)  { l.selectedValue = v }
func (l *List) SelectedValue() string      { return l.selectedValue }
func (l *List) OnChange(fn func(string))   { l.onChange = fn }

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

// effectiveItemHeight returns the actual item height, auto-detecting
// whether items have avatars/descriptions for taller rows.
func (l *List) effectiveItemHeight() float32 {
	if l.itemHeight != 48 && l.itemHeight != 36 && l.itemHeight != 64 {
		return l.itemHeight // custom height set by user
	}
	// Auto-detect: if any item has avatar or image, use taller rows
	for _, item := range l.items {
		if item.Avatar != "" || item.Image != 0 {
			if l.itemHeight < 72 {
				return 72
			}
		}
	}
	// If items have descriptions, use medium-tall rows
	for _, item := range l.items {
		if item.Description != "" && l.itemHeight < 64 {
			return 64
		}
	}
	return l.itemHeight
}

// groupCount returns the number of distinct group headers.
func (l *List) groupCount() int {
	seen := map[string]bool{}
	count := 0
	for _, item := range l.items {
		if item.Group != "" && !seen[item.Group] {
			seen[item.Group] = true
			count++
		}
	}
	return count
}

// TotalHeight returns the total height needed to render all items.
func (l *List) TotalHeight() float32 {
	ih := l.effectiveItemHeight()
	return ih*float32(len(l.items)) + l.groupHeight*float32(l.groupCount())
}

// updatePreferredHeight sets the widget's preferred height based on item count.
func (l *List) updatePreferredHeight() {
	l.style.Height = layout.Px(l.TotalHeight())
}

// hitTestItem returns the item index at the given relative Y position,
// accounting for group headers. Returns -1 if no item hit.
func (l *List) hitTestItem(relY float32) int {
	ih := l.effectiveItemHeight()
	y := float32(0)
	lastGroup := ""
	for i, item := range l.items {
		if item.Group != "" && item.Group != lastGroup {
			y += l.groupHeight
			lastGroup = item.Group
		}
		if relY >= y && relY < y+ih {
			return i
		}
		y += ih
	}
	return -1
}

func (l *List) Draw(buf *render.CommandBuffer) {
	bounds := l.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := l.config

	// Background (skip for selectable nav lists — use parent's bg)
	if !l.selectable {
		buf.DrawRect(render.RectCmd{
			Bounds:    bounds,
			FillColor: uimath.ColorWhite,
			Corners:   uimath.CornersAll(cfg.BorderRadius),
		}, 0, 1)
	}

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
	ih := l.effectiveItemHeight()
	lastGroup := ""
	for i, item := range l.items {
		// Group header
		if item.Group != "" && item.Group != lastGroup {
			lastGroup = item.Group
			if y+l.groupHeight >= bounds.Y && y <= bounds.Y+bounds.Height {
				l.drawGroupHeader(buf, item.Group, bounds.X, y, bounds.Width, cfg)
			}
			y += l.groupHeight
		}

		if y+ih < bounds.Y {
			y += ih
			continue
		}
		if y > bounds.Y+bounds.Height {
			break
		}

		// Selected highlight (selectable mode)
		if l.selectable && item.Value != "" && item.Value == l.selectedValue {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X, y, bounds.Width, ih),
				FillColor: uimath.RGBA(0, 0, 0, 0.06),
			}, 1, 1)
			// Left accent bar
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X, y, 3, ih),
				FillColor: cfg.PrimaryColor,
			}, 2, 1)
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
				Bounds:    uimath.NewRect(bounds.X+cfg.SpaceMD, y+ih-1, bounds.Width-cfg.SpaceMD*2, 1),
				FillColor: uimath.RGBA(0, 0, 0, 0.06),
			}, 2, 1)
		}

		// Text color override for selected item
		l.drawItem(buf, item, bounds, y, ih, cfg)
		y += ih
	}

	// Draw scrollbar if content overflows
	totalH := l.TotalHeight()
	if totalH > bounds.Height {
		sbWidth := float32(8)
		trackX := bounds.X + bounds.Width - sbWidth - 2
		// Track
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(trackX, bounds.Y+2, sbWidth, bounds.Height-4),
			FillColor: uimath.RGBA(0, 0, 0, 0.05),
			Corners:   uimath.CornersAll(sbWidth / 2),
		}, 10, 1)
		// Thumb
		trackH := bounds.Height - 4
		ratio := bounds.Height / totalH
		if ratio > 1 {
			ratio = 1
		}
		thumbH := trackH * ratio
		if thumbH < 20 {
			thumbH = 20
		}
		maxScroll := totalH - bounds.Height
		scrollRatio := float32(0)
		if maxScroll > 0 {
			scrollRatio = l.scrollY / maxScroll
		}
		thumbY := bounds.Y + 2 + (trackH-thumbH)*scrollRatio
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(trackX, thumbY, sbWidth, thumbH),
			FillColor: uimath.RGBA(0, 0, 0, 0.25),
			Corners:   uimath.CornersAll(sbWidth / 2),
		}, 11, 1)
	}

	buf.PopClip()
}

// drawGroupHeader renders a group title row.
func (l *List) drawGroupHeader(buf *render.CommandBuffer, title string, x, y, w float32, cfg *Config) {
	gh := l.groupHeight
	textClr := uimath.RGBA(0, 0, 0, 0.4)
	fs := cfg.FontSizeSm
	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(fs)
		cfg.TextRenderer.DrawText(buf, title, x+cfg.SpaceMD, y+(gh-lh)/2, fs, w-cfg.SpaceMD*2, textClr, 1)
	} else {
		tw := float32(len(title)) * fs * 0.55
		th := fs * 1.2
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x+cfg.SpaceMD, y+(gh-th)/2, tw, th),
			FillColor: textClr,
			Corners:   uimath.CornersAll(2),
		}, 1, 1)
	}
}

// drawItem renders a single list item.
func (l *List) drawItem(buf *render.CommandBuffer, item ListItem, bounds uimath.Rect, y, ih float32, cfg *Config) {
	leftPad := bounds.X + cfg.SpaceMD
	leftOffset := float32(0)

	// Avatar / Image
	hasAvatar := item.Avatar != "" || item.Image != 0
	if hasAvatar {
		avatarSize := float32(40)
		if ih < 64 {
			avatarSize = 32
		}
		avatarX := leftPad
		avatarY := y + (ih-avatarSize)/2

		if item.Image != 0 {
			// Draw image in circle
			buf.DrawImage(render.ImageCmd{
				Texture: item.Image,
				DstRect: uimath.NewRect(avatarX, avatarY, avatarSize, avatarSize),
				Tint:    uimath.ColorWhite,
				Corners: uimath.CornersAll(avatarSize / 2),
			}, 3, 1)
		} else {
			// Draw text avatar
			avatarBg := item.AvatarColor
			if avatarBg == (uimath.Color{}) {
				avatarBg = uimath.ColorHex("#c6c6c6")
			}
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(avatarX, avatarY, avatarSize, avatarSize),
				FillColor: avatarBg,
				Corners:   uimath.CornersAll(avatarSize / 2),
			}, 3, 1)
			if cfg.TextRenderer != nil && item.Avatar != "" {
				fs := avatarSize * 0.45
				lh := cfg.TextRenderer.LineHeight(fs)
				tw := cfg.TextRenderer.MeasureText(item.Avatar, fs)
				cfg.TextRenderer.DrawText(buf, item.Avatar,
					avatarX+(avatarSize-tw)/2, avatarY+(avatarSize-lh)/2,
					fs, avatarSize, uimath.ColorWhite, 1)
			}
		}
		leftOffset = avatarSize + cfg.SpaceMD
	}

	// Calculate right-side width for actions/extra
	rightPad := cfg.SpaceMD
	actionsW := l.measureActionsWidth(item, cfg)

	// Text area
	textX := leftPad + leftOffset
	textMaxW := bounds.Width - cfg.SpaceMD*2 - leftOffset - actionsW
	if textMaxW < 0 {
		textMaxW = 0
	}

	hasDesc := item.Description != ""

	// Title
	titleFs := cfg.FontSize
	if hasDesc || hasAvatar {
		titleFs = cfg.FontSize // bold in real rendering, same size for now
	}

	titleColor := cfg.TextColor
	if l.selectable && item.Value != "" && item.Value == l.selectedValue {
		titleColor = cfg.PrimaryColor
	}

	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(titleFs)
		var titleY float32
		if hasDesc {
			titleY = y + (ih/2-lh) - 2
		} else {
			titleY = y + (ih-lh)/2
		}
		cfg.TextRenderer.DrawText(buf, item.Title, textX, titleY, titleFs, textMaxW, titleColor, 1)
	} else {
		tw := float32(len(item.Title)) * titleFs * 0.55
		if tw > textMaxW {
			tw = textMaxW
		}
		th := titleFs * 1.2
		var titleY float32
		if hasDesc {
			titleY = y + (ih/2-th) - 2
		} else {
			titleY = y + (ih-th)/2
		}
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(textX, titleY, tw, th),
			FillColor: titleColor,
			Corners:   uimath.CornersAll(2),
		}, 3, 1)
	}

	// Description below title
	if hasDesc {
		descFs := cfg.FontSizeSm
		descColor := uimath.RGBA(0, 0, 0, 0.45)
		if cfg.TextRenderer != nil {
			descY := y + ih/2 + 2
			cfg.TextRenderer.DrawText(buf, item.Description, textX, descY, descFs, textMaxW, descColor, 1)
		} else {
			dw := float32(len(item.Description)) * descFs * 0.55
			if dw > textMaxW {
				dw = textMaxW
			}
			dh := descFs * 1.2
			descY := y + ih/2 + 2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(textX, descY, dw, dh),
				FillColor: uimath.RGBA(0, 0, 0, 0.25),
				Corners:   uimath.CornersAll(2),
			}, 3, 1)
		}
	}

	// Actions (right side)
	if len(item.Actions) > 0 {
		l.drawActions(buf, item.Actions, bounds, y, ih, cfg)
	} else if item.Extra != "" {
		l.drawExtra(buf, item.Extra, bounds, y, ih, cfg, rightPad)
	}
}

// measureActionsWidth calculates the total width needed for action links.
func (l *List) measureActionsWidth(item ListItem, cfg *Config) float32 {
	if len(item.Actions) > 0 {
		w := float32(0)
		for _, act := range item.Actions {
			if cfg.TextRenderer != nil {
				w += cfg.TextRenderer.MeasureText(act, cfg.FontSizeSm) + cfg.SpaceMD
			} else {
				w += float32(len(act))*cfg.FontSizeSm*0.55 + cfg.SpaceMD
			}
		}
		return w + cfg.SpaceMD
	}
	if item.Extra != "" {
		if cfg.TextRenderer != nil {
			return cfg.TextRenderer.MeasureText(item.Extra, cfg.FontSizeSm) + cfg.SpaceMD*2
		}
		return float32(len(item.Extra))*cfg.FontSizeSm*0.55 + cfg.SpaceMD*2
	}
	return 0
}

// drawActions renders action text links on the right side of an item.
func (l *List) drawActions(buf *render.CommandBuffer, actions []string, bounds uimath.Rect, y, ih float32, cfg *Config) {
	rightX := bounds.X + bounds.Width - cfg.SpaceMD
	actionColor := cfg.PrimaryColor

	// Draw actions right-to-left
	for i := len(actions) - 1; i >= 0; i-- {
		act := actions[i]
		var aw float32
		if cfg.TextRenderer != nil {
			aw = cfg.TextRenderer.MeasureText(act, cfg.FontSizeSm)
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			cfg.TextRenderer.DrawText(buf, act,
				rightX-aw, y+(ih-lh)/2,
				cfg.FontSizeSm, aw, actionColor, 1)
		} else {
			aw = float32(len(act)) * cfg.FontSizeSm * 0.55
			ah := cfg.FontSizeSm * 1.2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(rightX-aw, y+(ih-ah)/2, aw, ah),
				FillColor: actionColor,
				Corners:   uimath.CornersAll(2),
			}, 3, 1)
		}
		rightX -= aw + cfg.SpaceMD
	}
}

// drawExtra renders a single extra text on the right.
func (l *List) drawExtra(buf *render.CommandBuffer, extra string, bounds uimath.Rect, y, ih float32, cfg *Config, rightPad float32) {
	extraX := bounds.X + bounds.Width - rightPad
	extraColor := uimath.RGBA(0, 0, 0, 0.45)
	if cfg.TextRenderer != nil {
		ew := cfg.TextRenderer.MeasureText(extra, cfg.FontSizeSm)
		lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
		cfg.TextRenderer.DrawText(buf, extra, extraX-ew, y+(ih-lh)/2, cfg.FontSizeSm, ew, extraColor, 1)
	} else {
		ew := float32(len(extra)) * cfg.FontSizeSm * 0.55
		eh := cfg.FontSizeSm * 1.2
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(extraX-ew, y+(ih-eh)/2, ew, eh),
			FillColor: uimath.RGBA(0, 0, 0, 0.25),
			Corners:   uimath.CornersAll(2),
		}, 3, 1)
	}
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

func (vl *VirtualList) SetItemCount(n int)      { vl.itemCount = n }
func (vl *VirtualList) SetItemHeight(h float32)  { vl.itemHeight = h }
func (vl *VirtualList) SetScrollY(y float32)     { vl.scrollY = y }
func (vl *VirtualList) ScrollY() float32         { return vl.scrollY }
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

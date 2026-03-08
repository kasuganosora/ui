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
// Matches TDesign List: text-only, multi-line, avatar+text, actions.
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

	// Track hover
	tree.AddHandler(l.id, event.MouseMove, func(e *event.Event) {
		if !l.hoverable {
			return
		}
		bounds := l.Bounds()
		relY := e.Y - bounds.Y + l.scrollY
		ih := l.effectiveItemHeight()
		idx := int(relY / ih)
		if idx >= 0 && idx < len(l.items) {
			l.hoveredIndex = idx
		} else {
			l.hoveredIndex = -1
		}
	})
	tree.AddHandler(l.id, event.MouseLeave, func(e *event.Event) {
		l.hoveredIndex = -1
	})

	// Click handler for item selection
	tree.AddHandler(l.id, event.MouseClick, func(e *event.Event) {
		bounds := l.Bounds()
		relY := e.Y - bounds.Y + l.scrollY
		ih := l.effectiveItemHeight()
		idx := int(relY / ih)
		if idx >= 0 && idx < len(l.items) {
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

// TotalHeight returns the total height needed to render all items.
func (l *List) TotalHeight() float32 {
	return l.effectiveItemHeight() * float32(len(l.items))
}

// updatePreferredHeight sets the widget's preferred height based on item count.
func (l *List) updatePreferredHeight() {
	l.style.Height = layout.Px(l.TotalHeight())
}

func (l *List) Draw(buf *render.CommandBuffer) {
	bounds := l.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := l.config

	// Background
	buf.DrawRect(render.RectCmd{
		Bounds:    bounds,
		FillColor: uimath.ColorWhite,
		Corners:   uimath.CornersAll(cfg.BorderRadius),
	}, 0, 1)

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
				Bounds:    uimath.NewRect(bounds.X+cfg.SpaceMD, y+ih-1, bounds.Width-cfg.SpaceMD*2, 1),
				FillColor: uimath.RGBA(0, 0, 0, 0.06),
			}, 2, 1)
		}

		l.drawItem(buf, item, bounds, y, ih, cfg)
		y += ih
	}
	buf.PopClip()
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

	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(titleFs)
		var titleY float32
		if hasDesc {
			titleY = y + (ih/2-lh) - 2
		} else {
			titleY = y + (ih-lh)/2
		}
		cfg.TextRenderer.DrawText(buf, item.Title, textX, titleY, titleFs, textMaxW, cfg.TextColor, 1)
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
			FillColor: cfg.TextColor,
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

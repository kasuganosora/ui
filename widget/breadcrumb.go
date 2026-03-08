package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// BreadcrumbItem is a single breadcrumb entry (TDesign: TdBreadcrumbItemProps).
type BreadcrumbItem struct {
	Content  string // display text (TDesign: content)
	Href     string // navigation link (TDesign: href)
	Disabled bool   // whether click is disabled (TDesign: disabled)
	MaxWidth string // max width with ellipsis (TDesign: maxWidth)
}

// Breadcrumb displays a path of navigable links (TDesign: TdBreadcrumbProps).
type Breadcrumb struct {
	Base
	options      []BreadcrumbItem // breadcrumb items (TDesign: options)
	separator    string           // separator character (TDesign: separator)
	maxItemWidth string           // max width per item (TDesign: maxItemWidth)
	maxItems     int              // max visible items, 0 = unlimited (TDesign: maxItems)
	theme        string           // visual theme (TDesign: theme)
	onClick      func(index int, href string)
	itemIDs      []core.ElementID
}

func NewBreadcrumb(tree *core.Tree, cfg *Config) *Breadcrumb {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Breadcrumb{
		Base:      NewBase(tree, core.TypeCustom, cfg),
		separator: "/",
		theme:     "light",
	}
}

// SetOptions sets the breadcrumb items (TDesign: options).
func (b *Breadcrumb) SetOptions(items []BreadcrumbItem) {
	// Clean up old item elements
	for _, id := range b.itemIDs {
		b.tree.DestroyElement(id)
	}
	b.itemIDs = nil

	b.options = items

	// Create clickable elements for non-last, non-disabled items
	for i, item := range items {
		isLast := i == len(items)-1
		if isLast {
			break // Last item is not clickable
		}
		if item.Disabled {
			continue
		}
		itemID := b.tree.CreateElement(core.TypeCustom)
		b.tree.AppendChild(b.id, itemID)
		b.itemIDs = append(b.itemIDs, itemID)

		idx := i
		href := item.Href
		b.tree.AddHandler(itemID, event.MouseClick, func(e *event.Event) {
			if b.onClick != nil {
				b.onClick(idx, href)
			}
		})
	}
}

func (b *Breadcrumb) SetSeparator(s string)    { b.separator = s }
func (b *Breadcrumb) SetMaxItemWidth(w string)  { b.maxItemWidth = w }
func (b *Breadcrumb) SetMaxItems(n int)         { b.maxItems = n }
func (b *Breadcrumb) SetTheme(t string)         { b.theme = t }

// OnClick sets the callback for breadcrumb item clicks (TDesign: onClick on BreadcrumbItem).
func (b *Breadcrumb) OnClick(fn func(int, string)) { b.onClick = fn }

func (b *Breadcrumb) Draw(buf *render.CommandBuffer) {
	bounds := b.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := b.config
	x := bounds.X
	itemIdx := 0
	for i, item := range b.options {
		isLast := i == len(b.options)-1
		color := cfg.PrimaryColor
		if isLast || item.Disabled {
			color = cfg.TextColor
		}

		var tw float32
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			tw = cfg.TextRenderer.MeasureText(item.Content, cfg.FontSize)

			// Check hover for non-last, non-disabled items
			if !isLast && !item.Disabled && itemIdx < len(b.itemIDs) {
				elem := b.tree.Get(b.itemIDs[itemIdx])
				if elem != nil && elem.IsHovered() {
					color = cfg.HoverColor
				}
			}

			cfg.TextRenderer.DrawText(buf, item.Content, x, bounds.Y+(bounds.Height-lh)/2, cfg.FontSize, tw+1, color, 1)
		} else {
			tw = float32(len(item.Content)) * cfg.FontSize * 0.55
			th := cfg.FontSize * 1.2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(x, bounds.Y+(bounds.Height-th)/2, tw, th),
				FillColor: color,
				Corners:   uimath.CornersAll(2),
			}, 0, 1)
		}

		// Set bounds on clickable element for hit testing
		if !isLast && !item.Disabled && itemIdx < len(b.itemIDs) {
			b.tree.SetLayout(b.itemIDs[itemIdx], core.LayoutResult{
				Bounds: uimath.NewRect(x, bounds.Y, tw, bounds.Height),
			})
			itemIdx++
		}

		x += tw
		if !isLast {
			sepClr := uimath.RGBA(0, 0, 0, 0.45)
			if cfg.TextRenderer != nil {
				sw := cfg.TextRenderer.MeasureText(b.separator, cfg.FontSize)
				lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
				cfg.TextRenderer.DrawText(buf, b.separator, x+cfg.SpaceXS, bounds.Y+(bounds.Height-lh)/2, cfg.FontSize, sw+1, sepClr, 1)
				x += sw + cfg.SpaceXS*2
			} else {
				x += cfg.SpaceSM
			}
		}
	}
}

func (b *Breadcrumb) Destroy() {
	for _, id := range b.itemIDs {
		b.tree.DestroyElement(id)
	}
	b.Base.Destroy()
}

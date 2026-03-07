package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// TabItem defines a single tab.
type TabItem struct {
	Key     string
	Label   string
	Content Widget
}

// Tabs is a tabbed container widget.
type Tabs struct {
	Base
	items     []TabItem
	activeKey string
	tabIDs    []core.ElementID // element IDs for clickable tab headers

	onChange func(key string)
}

// NewTabs creates a tabs widget.
func NewTabs(tree *core.Tree, items []TabItem, cfg *Config) *Tabs {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	t := &Tabs{
		Base:  NewBase(tree, core.TypeCustom, cfg),
		items: items,
	}
	t.style.Display = layout.DisplayFlex
	t.style.FlexDirection = layout.FlexDirectionColumn

	if len(items) > 0 {
		t.activeKey = items[0].Key
	}

	// Create clickable element for each tab header
	for i, item := range items {
		tabID := tree.CreateElement(core.TypeCustom)
		tree.SetProperty(tabID, "text", item.Label)
		t.tabIDs = append(t.tabIDs, tabID)

		idx := i
		tree.AddHandler(tabID, event.MouseClick, func(e *event.Event) {
			t.activeKey = t.items[idx].Key
			if t.onChange != nil {
				t.onChange(t.activeKey)
			}
		})
	}

	return t
}

func (t *Tabs) ActiveKey() string { return t.activeKey }

func (t *Tabs) SetActiveKey(key string) { t.activeKey = key }

func (t *Tabs) OnChange(fn func(key string)) { t.onChange = fn }

const tabHeaderHeight = float32(40)

func (t *Tabs) Draw(buf *render.CommandBuffer) {
	bounds := t.Bounds()
	if bounds.IsEmpty() {
		return
	}

	cfg := t.config

	// Draw tab header bar
	headerRect := uimath.NewRect(bounds.X, bounds.Y, bounds.Width, tabHeaderHeight)

	// Bottom border of header
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(bounds.X, bounds.Y+tabHeaderHeight-1, bounds.Width, 1),
		FillColor: cfg.BorderColor,
	}, 0, 1)

	// Draw each tab header
	tabX := bounds.X
	for i, item := range t.items {
		active := item.Key == t.activeKey
		elem := t.tree.Get(t.tabIDs[i])
		hovered := elem != nil && elem.IsHovered()

		// Measure tab width
		tabW := float32(len(item.Label))*cfg.FontSize*0.55 + cfg.SpaceMD*2
		if cfg.TextRenderer != nil {
			tabW = cfg.TextRenderer.MeasureText(item.Label, cfg.FontSize) + cfg.SpaceMD*2
		}

		// Tab background on hover
		if hovered && !active {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(tabX, bounds.Y, tabW, tabHeaderHeight-1),
				FillColor: uimath.ColorHex("#fafafa"),
			}, 0, 1)
		}

		// Tab label
		textColor := cfg.TextColor
		if active {
			textColor = cfg.PrimaryColor
		} else if hovered {
			textColor = cfg.HoverColor
		}

		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			tx := tabX + cfg.SpaceMD
			ty := headerRect.Y + (tabHeaderHeight-lh)/2
			cfg.TextRenderer.DrawText(buf, item.Label, tx, ty, cfg.FontSize, tabW-cfg.SpaceMD*2, textColor, 1)
		} else {
			textW := float32(len(item.Label)) * cfg.FontSize * 0.55
			textH := cfg.FontSize * 1.2
			tx := tabX + (tabW-textW)/2
			ty := headerRect.Y + (tabHeaderHeight-textH)/2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(tx, ty, textW, textH),
				FillColor: textColor,
				Corners:   uimath.CornersAll(2),
			}, 1, 1)
		}

		// Active indicator
		if active {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(tabX, bounds.Y+tabHeaderHeight-2, tabW, 2),
				FillColor: cfg.PrimaryColor,
			}, 1, 1)
		}

		tabX += tabW
	}

	// Draw active tab content
	for _, item := range t.items {
		if item.Key == t.activeKey && item.Content != nil {
			item.Content.Draw(buf)
		}
	}
}

// Destroy cleans up tab header elements.
func (t *Tabs) Destroy() {
	for _, tabID := range t.tabIDs {
		t.tree.DestroyElement(tabID)
	}
	t.Base.Destroy()
}

package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// MenuItem represents a single menu entry.
type MenuItem struct {
	Key      string
	Label    string
	Icon     render.TextureHandle
	Disabled bool
	Children []MenuItem
}

// Menu is a vertical navigation menu.
type Menu struct {
	Base
	items       []MenuItem
	selectedKey string
	openKeys    map[string]bool
	onSelect    func(key string)
	itemHeight  float32
	indent      float32
}

func NewMenu(tree *core.Tree, cfg *Config) *Menu {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	m := &Menu{
		Base:       NewBase(tree, core.TypeDiv, cfg),
		openKeys:   make(map[string]bool),
		itemHeight: 40,
		indent:     24,
	}
	m.style.Display = layout.DisplayFlex
	m.style.FlexDirection = layout.FlexDirectionColumn

	tree.AddHandler(m.id, event.MouseClick, func(e *event.Event) {
		// Hit test handled externally
	})
	return m
}

func (m *Menu) SetItems(items []MenuItem) { m.items = items }
func (m *Menu) SelectedKey() string       { return m.selectedKey }
func (m *Menu) SetSelectedKey(k string)   { m.selectedKey = k }
func (m *Menu) OnSelect(fn func(string))  { m.onSelect = fn }

func (m *Menu) ToggleOpen(key string) {
	if m.openKeys[key] {
		delete(m.openKeys, key)
	} else {
		m.openKeys[key] = true
	}
}

func (m *Menu) SelectItem(key string) {
	m.selectedKey = key
	if m.onSelect != nil {
		m.onSelect(key)
	}
}

func (m *Menu) Draw(buf *render.CommandBuffer) {
	bounds := m.Bounds()
	if bounds.IsEmpty() {
		return
	}
	y := bounds.Y
	for _, item := range m.items {
		y = m.drawItem(buf, item, bounds.X, y, bounds.Width, 0)
	}
}

func (m *Menu) drawItem(buf *render.CommandBuffer, item MenuItem, x, y, w float32, depth int) float32 {
	cfg := m.config
	h := m.itemHeight
	indentX := x + float32(depth)*m.indent

	selected := item.Key == m.selectedKey
	elem := m.Element()
	hovered := elem != nil && elem.IsHovered()

	bg := uimath.Color{}
	if selected {
		bg = uimath.ColorHex("#e6f4ff")
	} else if hovered {
		bg = uimath.RGBA(0, 0, 0, 0.04)
	}
	if bg.A > 0 {
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x, y, w, h),
			FillColor: bg,
		}, 0, 1)
	}

	if selected {
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x+w-3, y, 3, h),
			FillColor: cfg.PrimaryColor,
		}, 1, 1)
	}

	textClr := cfg.TextColor
	if item.Disabled {
		textClr = cfg.DisabledColor
	} else if selected {
		textClr = cfg.PrimaryColor
	}

	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, item.Label, indentX+cfg.SpaceMD, y+(h-lh)/2, cfg.FontSize, w-float32(depth)*m.indent-cfg.SpaceMD*2, textClr, 1)
	} else {
		tw := float32(len(item.Label)) * cfg.FontSize * 0.55
		th := cfg.FontSize * 1.2
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(indentX+cfg.SpaceMD, y+(h-th)/2, tw, th),
			FillColor: textClr,
			Corners:   uimath.CornersAll(2),
		}, 1, 1)
	}

	y += h
	if m.openKeys[item.Key] {
		for _, child := range item.Children {
			y = m.drawItem(buf, child, x, y, w, depth+1)
		}
	}
	return y
}

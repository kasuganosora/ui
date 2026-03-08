package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// MenuTheme controls the color scheme of the menu (TDesign: theme).
type MenuTheme int

const (
	// MenuThemeLight is the default light theme.
	MenuThemeLight MenuTheme = iota
	// MenuThemeDark uses a dark background with light text.
	MenuThemeDark
)

// MenuItem represents a single menu entry (TDesign: TdMenuItemProps).
type MenuItem struct {
	Value    string            // unique identifier (TDesign: value)
	Content  string            // menu item text content (TDesign: content)
	Icon     render.TextureHandle // icon (TDesign: icon)
	Disabled bool              // whether disabled (TDesign: disabled)
	Children []MenuItem        // submenu items
}

// menuItemRect stores layout info for hit testing.
type menuItemRect struct {
	value  string
	y, h   float32
	hasSub bool
}

// Menu is a vertical navigation menu (TDesign: TdMenuProps).
type Menu struct {
	Base
	items       []MenuItem
	value       string          // active menu item value (TDesign: value)
	expanded    map[string]bool // expanded submenu values (TDesign: expanded)
	onChange    func(value string)        // TDesign: onChange
	onExpand    func(expanded []string)   // TDesign: onExpand
	itemHeight  float32
	indent      float32
	hoveredValue string
	itemRects   []menuItemRect // rebuilt each Draw for hit testing
	theme       MenuTheme
	collapsed   bool    // collapsed sidebar mode (TDesign: collapsed)
	expandMutex bool    // same-level mutual exclusion (TDesign: expandMutex)
	width       float32 // menu width, 0 = auto (TDesign: width)
}

func NewMenu(tree *core.Tree, cfg *Config) *Menu {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	m := &Menu{
		Base:       NewBase(tree, core.TypeDiv, cfg),
		expanded:   make(map[string]bool),
		itemHeight: 40,
		indent:     24,
	}
	m.style.Display = layout.DisplayFlex
	m.style.FlexDirection = layout.FlexDirectionColumn

	tree.AddHandler(m.id, event.MouseClick, func(e *event.Event) {
		m.handleClick(e.GlobalX, e.GlobalY)
	})
	tree.AddHandler(m.id, event.MouseMove, func(e *event.Event) {
		m.handleHover(e.GlobalX, e.GlobalY)
	})
	return m
}

func (m *Menu) SetItems(items []MenuItem) { m.items = items }

// Value returns the active menu item value.
func (m *Menu) Value() string { return m.value }

// SetValue sets the active menu item value.
func (m *Menu) SetValue(v string) {
	m.value = v
	// Auto-open parent submenus for the selected value
	m.autoOpenParents(v)
}

// Expanded returns the list of expanded submenu values.
func (m *Menu) Expanded() []string {
	keys := make([]string, 0, len(m.expanded))
	for k := range m.expanded {
		keys = append(keys, k)
	}
	return keys
}

// SetExpanded sets which submenus are expanded.
func (m *Menu) SetExpanded(values []string) {
	m.expanded = make(map[string]bool, len(values))
	for _, v := range values {
		m.expanded[v] = true
	}
}

// OnChange sets the callback invoked when the active menu item changes (TDesign: onChange).
func (m *Menu) OnChange(fn func(value string)) { m.onChange = fn }

// OnExpand sets the callback invoked when expanded submenus change (TDesign: onExpand).
func (m *Menu) OnExpand(fn func(expanded []string)) { m.onExpand = fn }

// SetTheme sets the menu color theme.
func (m *Menu) SetTheme(t MenuTheme) { m.theme = t }

// Theme returns the current menu theme.
func (m *Menu) Theme() MenuTheme { return m.theme }

// SetCollapsed enables or disables collapsed sidebar mode (icons only).
func (m *Menu) SetCollapsed(v bool) { m.collapsed = v }

// IsCollapsed returns whether the menu is in collapsed mode.
func (m *Menu) IsCollapsed() bool { return m.collapsed }

// SetExpandMutex sets same-level mutual exclusion for submenu expansion.
func (m *Menu) SetExpandMutex(v bool) { m.expandMutex = v }

// IsExpandMutex returns whether expand mutex is enabled.
func (m *Menu) IsExpandMutex() bool { return m.expandMutex }

// SetWidth sets the menu width.
func (m *Menu) SetWidth(w float32) { m.width = w }

// Width returns the menu width.
func (m *Menu) Width() float32 { return m.width }

func (m *Menu) ToggleExpanded(value string) {
	if m.expanded[value] {
		delete(m.expanded, value)
	} else {
		if m.expandMutex {
			// Close siblings at same level
			m.closeSiblings(value)
		}
		m.expanded[value] = true
	}
	if m.onExpand != nil {
		m.onExpand(m.Expanded())
	}
}

// closeSiblings closes sibling submenus at the same level as the given value.
func (m *Menu) closeSiblings(value string) {
	// Find siblings of the given value and close them
	for i := range m.items {
		if m.items[i].Value == value {
			// Top-level: close all other top-level expanded
			for _, item := range m.items {
				if item.Value != value {
					delete(m.expanded, item.Value)
				}
			}
			return
		}
		if m.closeSiblingsIn(&m.items[i], value) {
			return
		}
	}
}

func (m *Menu) closeSiblingsIn(parent *MenuItem, value string) bool {
	for _, child := range parent.Children {
		if child.Value == value {
			// Found: close siblings
			for _, sib := range parent.Children {
				if sib.Value != value {
					delete(m.expanded, sib.Value)
				}
			}
			return true
		}
		if m.closeSiblingsIn(&child, value) {
			return true
		}
	}
	return false
}

func (m *Menu) SelectItem(value string) {
	m.value = value
	if m.onChange != nil {
		m.onChange(value)
	}
}

// autoOpenParents opens all ancestor submenus that contain the given value.
func (m *Menu) autoOpenParents(value string) {
	for i := range m.items {
		if m.openParentPath(&m.items[i], value) {
			return
		}
	}
}

func (m *Menu) openParentPath(item *MenuItem, value string) bool {
	if item.Value == value {
		return true
	}
	for i := range item.Children {
		if m.openParentPath(&item.Children[i], value) {
			m.expanded[item.Value] = true
			return true
		}
	}
	return false
}

func (m *Menu) handleClick(gx, gy float32) {
	for _, r := range m.itemRects {
		if gy >= r.y && gy < r.y+r.h {
			if r.hasSub {
				m.ToggleExpanded(r.value)
			} else {
				m.SelectItem(r.value)
			}
			m.tree.MarkDirty(m.id)
			return
		}
	}
}

func (m *Menu) handleHover(gx, gy float32) {
	newHovered := ""
	for _, r := range m.itemRects {
		if gy >= r.y && gy < r.y+r.h {
			newHovered = r.value
			break
		}
	}
	if newHovered != m.hoveredValue {
		m.hoveredValue = newHovered
		m.tree.MarkDirty(m.id)
	}
}

// TotalHeight returns the total height needed to render all visible items.
func (m *Menu) TotalHeight() float32 {
	h := float32(0)
	for _, item := range m.items {
		h += m.calcItemHeight(item)
	}
	return h
}

func (m *Menu) calcItemHeight(item MenuItem) float32 {
	h := m.itemHeight
	if !m.collapsed && m.expanded[item.Value] {
		for _, child := range item.Children {
			h += m.calcItemHeight(child)
		}
	}
	return h
}

// themeColors returns (background, text, selectedBg, selectedText, hoverBg, arrowColor)
// based on the current theme.
func (m *Menu) themeColors() (bg, text, selBg, selText, hoverBg, arrowClr uimath.Color) {
	cfg := m.config
	if m.theme == MenuThemeDark {
		bg = uimath.ColorHex("#001529")
		text = uimath.RGBA(1, 1, 1, 0.65)
		selBg = cfg.PrimaryColor
		selText = uimath.ColorWhite
		hoverBg = uimath.RGBA(1, 1, 1, 0.08)
		arrowClr = uimath.RGBA(1, 1, 1, 0.45)
	} else {
		bg = uimath.ColorWhite
		text = cfg.TextColor
		selBg = uimath.ColorHex("#f2f3ff")
		selText = cfg.PrimaryColor
		hoverBg = uimath.RGBA(0, 0, 0, 0.04)
		arrowClr = uimath.ColorHex("#8b8b8b")
	}
	return
}

func (m *Menu) Draw(buf *render.CommandBuffer) {
	bounds := m.Bounds()
	if bounds.IsEmpty() {
		return
	}

	bgColor, _, _, _, _, _ := m.themeColors()

	// Background
	buf.DrawRect(render.RectCmd{
		Bounds:    bounds,
		FillColor: bgColor,
	}, 0, 1)

	// Rebuild item rects for hit testing
	m.itemRects = m.itemRects[:0]
	y := bounds.Y
	for _, item := range m.items {
		y = m.drawItem(buf, item, bounds.X, y, bounds.Width, 0)
	}
}

func (m *Menu) drawItem(buf *render.CommandBuffer, item MenuItem, x, y, w float32, depth int) float32 {
	cfg := m.config
	h := m.itemHeight
	_, textClrBase, selBg, selText, hoverBg, arrowClr := m.themeColors()

	effectiveIndent := float32(0)
	if !m.collapsed {
		effectiveIndent = float32(depth) * m.indent
	}
	indentX := x + effectiveIndent

	// Record rect for hit testing
	m.itemRects = append(m.itemRects, menuItemRect{
		value:  item.Value,
		y:      y,
		h:      h,
		hasSub: len(item.Children) > 0,
	})

	selected := item.Value == m.value
	hovered := item.Value == m.hoveredValue

	// Background
	bg := uimath.Color{}
	if selected {
		bg = selBg
	} else if hovered && !item.Disabled {
		bg = hoverBg
	}
	if bg.A > 0 {
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x, y, w, h),
			FillColor: bg,
		}, 0, 1)
	}

	// Selected indicator (right bar)
	if selected {
		barColor := cfg.PrimaryColor
		if m.theme == MenuThemeDark {
			barColor = uimath.ColorWhite
		}
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x+w-3, y, 3, h),
			FillColor: barColor,
		}, 1, 1)
	}

	// Text color
	textClr := textClrBase
	if item.Disabled {
		textClr = cfg.DisabledColor
	} else if selected {
		textClr = selText
	}

	// Icon placeholder (draw always, even in collapsed mode)
	iconSize := float32(16)
	iconX := indentX + cfg.SpaceMD
	if item.Icon != 0 {
		// Draw icon as a small colored square placeholder
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(iconX, y+(h-iconSize)/2, iconSize, iconSize),
			FillColor: textClr,
			Corners:   uimath.CornersAll(2),
		}, 1, 1)
	}

	// Label (skip in collapsed mode)
	if !m.collapsed {
		labelX := indentX + cfg.SpaceMD
		if item.Icon != 0 {
			labelX += iconSize + cfg.SpaceXS
		}
		availW := w - effectiveIndent - cfg.SpaceMD*2
		if item.Icon != 0 {
			availW -= iconSize + cfg.SpaceXS
		}

		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			cfg.TextRenderer.DrawText(buf, item.Content, labelX, y+(h-lh)/2, cfg.FontSize, availW, textClr, 1)
		} else {
			tw := float32(len(item.Content)) * cfg.FontSize * 0.55
			th := cfg.FontSize * 1.2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(labelX, y+(h-th)/2, tw, th),
				FillColor: textClr,
				Corners:   uimath.CornersAll(2),
			}, 1, 1)
		}
	}

	// Expand/collapse arrow for items with children (hidden when collapsed)
	if len(item.Children) > 0 && !m.collapsed {
		open := m.expanded[item.Value]
		arrowX := x + w - cfg.SpaceMD - 8
		arrowY := y + h/2
		m.drawArrow(buf, arrowX, arrowY, 5, open, arrowClr)
	}

	y += h

	// Draw children if expanded (never show in collapsed mode)
	if !m.collapsed && m.expanded[item.Value] {
		for _, child := range item.Children {
			y = m.drawItem(buf, child, x, y, w, depth+1)
		}
	}
	return y
}

// drawArrow draws a small triangle arrow (right when closed, down when open).
func (m *Menu) drawArrow(buf *render.CommandBuffer, cx, cy, size float32, open bool, color uimath.Color) {
	// Approximate arrow with small rects (since we don't have triangle primitives)
	// Right arrow: > , Down arrow: v
	s := size
	if open {
		// Down arrow: 3 rects getting narrower
		for i := 0; i < 3; i++ {
			fi := float32(i)
			rw := s - fi*2
			if rw < 1 {
				rw = 1
			}
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(cx-rw/2, cy-s/2+fi*2, rw, 1.5),
				FillColor: color,
			}, 2, 1)
		}
	} else {
		// Right arrow: 3 rects getting narrower
		for i := 0; i < 3; i++ {
			fi := float32(i)
			rh := s - fi*2
			if rh < 1 {
				rh = 1
			}
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(cx-s/2+fi*2, cy-rh/2, 1.5, rh),
				FillColor: color,
			}, 2, 1)
		}
	}
}

// findItemAt returns the value of the item at the given global position, or "".
func (m *Menu) findItemAt(gy float32) string {
	for _, r := range m.itemRects {
		if gy >= r.y && gy < r.y+r.h {
			return r.value
		}
	}
	return ""
}

package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// TabTheme controls the visual style of the tabs.
type TabTheme int

const (
	// TabThemeNormal renders tabs with an underline indicator (default).
	TabThemeNormal TabTheme = iota
	// TabThemeCard renders tabs as bordered cards; the active tab has a white
	// background and no bottom border.
	TabThemeCard
)

// TabPlacement controls where the tab header bar is rendered.
type TabPlacement int

const (
	TabPlacementTop    TabPlacement = iota // default
	TabPlacementBottom
	TabPlacementLeft
	TabPlacementRight
)

// TabPanel defines a single tab panel (TDesign: TdTabPanelProps).
type TabPanel struct {
	Value     string // unique tab identifier (TDesign: value)
	Label     string // tab header text (TDesign: label)
	Content   Widget // tab panel content (TDesign: panel)
	Removable bool   // per-tab remove button (TDesign: removable)
	Disabled  bool   // whether this tab is disabled (TDesign: disabled)
	Draggable bool   // whether drag-sorting is allowed for this tab (TDesign: draggable)
}

// Tabs is a tabbed container widget (TDesign: TdTabsProps).
type Tabs struct {
	Base
	list     []TabPanel
	value    string // active tab value (TDesign: value)
	tabIDs   []core.ElementID // element IDs for clickable tab headers

	theme     TabTheme     // visual style: normal or card (TDesign: theme)
	placement TabPlacement // tab header position (TDesign: placement)
	size      Size         // component size (TDesign: size)
	addable   bool         // whether tabs can be added (TDesign: addable)
	disabled  bool         // whether all tabs are disabled (TDesign: disabled)
	onAdd     func()
	onRemove  func(options TabRemoveOptions) // TDesign: onRemove
	onChange  func(value string)             // TDesign: onChange

	addBtnID core.ElementID // element for the "+" button (0 if not addable)
}

// TabRemoveOptions is the argument to the OnRemove callback.
type TabRemoveOptions struct {
	Value string
	Index int
}

// NewTabs creates a tabs widget.
func NewTabs(tree *core.Tree, list []TabPanel, cfg *Config) *Tabs {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	t := &Tabs{
		Base: NewBase(tree, core.TypeCustom, cfg),
		list: list,
		size: SizeMedium,
	}
	t.style.Display = layout.DisplayFlex
	t.style.FlexDirection = layout.FlexDirectionColumn

	if len(list) > 0 {
		t.value = list[0].Value
	}

	// Create clickable element for each tab header (parented to Tabs for hit testing)
	for i, item := range list {
		tabID := tree.CreateElement(core.TypeCustom)
		tree.AppendChild(t.id, tabID)
		tree.SetProperty(tabID, "text", item.Label)
		t.tabIDs = append(t.tabIDs, tabID)

		idx := i
		tree.AddHandler(tabID, event.MouseClick, func(e *event.Event) {
			if t.disabled || t.list[idx].Disabled {
				return
			}
			t.value = t.list[idx].Value
			if t.onChange != nil {
				t.onChange(t.value)
			}
			t.tree.MarkDirty(t.id)
		})
	}

	return t
}

// Value returns the active tab value.
func (t *Tabs) Value() string { return t.value }

// SetValue sets the active tab value.
func (t *Tabs) SetValue(v string) { t.value = v }

// List returns the tab panels.
func (t *Tabs) List() []TabPanel { return t.list }

// OnChange sets the callback invoked when the active tab changes.
func (t *Tabs) OnChange(fn func(value string)) { t.onChange = fn }

// SetTheme sets the visual style (TabThemeNormal or TabThemeCard).
func (t *Tabs) SetTheme(theme TabTheme) { t.theme = theme }

// Theme returns the current tab theme.
func (t *Tabs) Theme() TabTheme { return t.theme }

// SetPlacement sets where the tab headers are rendered.
func (t *Tabs) SetPlacement(p TabPlacement) { t.placement = p }

// Placement returns the current placement.
func (t *Tabs) Placement() TabPlacement { return t.placement }

// SetSize sets the component size.
func (t *Tabs) SetSize(s Size) { t.size = s }

// Size returns the component size.
func (t *Tabs) Size() Size { return t.size }

// SetDisabled disables or enables all tabs.
func (t *Tabs) SetDisabled(v bool) { t.disabled = v }

// IsDisabled returns whether all tabs are disabled.
func (t *Tabs) IsDisabled() bool { return t.disabled }

// SetAddable enables or disables the "+" add-tab button.
func (t *Tabs) SetAddable(v bool) { t.addable = v }

// OnAdd sets the callback invoked when the "+" button is clicked.
func (t *Tabs) OnAdd(fn func()) {
	t.onAdd = fn
	t.addable = true
	// Lazily create add-button element
	if t.addBtnID == 0 {
		t.addBtnID = t.tree.CreateElement(core.TypeCustom)
		t.tree.AppendChild(t.id, t.addBtnID)
		t.tree.SetProperty(t.addBtnID, "text", "+")
		t.tree.AddHandler(t.addBtnID, event.MouseClick, func(e *event.Event) {
			if t.onAdd != nil {
				t.onAdd()
			}
		})
	}
}

// OnRemove sets the callback invoked when a tab's remove button is clicked.
func (t *Tabs) OnRemove(fn func(options TabRemoveOptions)) { t.onRemove = fn }

const tabHeaderHeight = float32(40)
const tabCloseSize = float32(14) // width/height of close button area

// headerIsVertical returns true for Left/Right placements.
func (t *Tabs) headerIsVertical() bool {
	return t.placement == TabPlacementLeft || t.placement == TabPlacementRight
}

func (t *Tabs) Draw(buf *render.CommandBuffer) {
	bounds := t.Bounds()
	if bounds.IsEmpty() {
		return
	}

	vertical := t.headerIsVertical()

	// Determine header and content rectangles.
	var headerBounds, contentBounds uimath.Rect
	headerSize := tabHeaderHeight
	if vertical {
		headerW := float32(120) // fixed sidebar width for vertical tabs
		switch t.placement {
		case TabPlacementLeft:
			headerBounds = uimath.NewRect(bounds.X, bounds.Y, headerW, bounds.Height)
			contentBounds = uimath.NewRect(bounds.X+headerW, bounds.Y, bounds.Width-headerW, bounds.Height)
		case TabPlacementRight:
			headerBounds = uimath.NewRect(bounds.X+bounds.Width-headerW, bounds.Y, headerW, bounds.Height)
			contentBounds = uimath.NewRect(bounds.X, bounds.Y, bounds.Width-headerW, bounds.Height)
		}
	} else {
		switch t.placement {
		case TabPlacementBottom:
			headerBounds = uimath.NewRect(bounds.X, bounds.Y+bounds.Height-headerSize, bounds.Width, headerSize)
			contentBounds = uimath.NewRect(bounds.X, bounds.Y, bounds.Width, bounds.Height-headerSize)
		default: // Top
			headerBounds = uimath.NewRect(bounds.X, bounds.Y, bounds.Width, headerSize)
			contentBounds = uimath.NewRect(bounds.X, bounds.Y+headerSize, bounds.Width, bounds.Height-headerSize)
		}
	}

	// Draw header background for card style
	if t.theme == TabThemeCard {
		buf.DrawRect(render.RectCmd{
			Bounds:    headerBounds,
			FillColor: uimath.ColorHex("#f3f3f3"),
		}, 0, 1)
	}

	// Draw border line along the content edge of the header
	t.drawHeaderBorder(buf, headerBounds)

	// Draw each tab header
	if vertical {
		t.drawVerticalHeaders(buf, headerBounds)
	} else {
		t.drawHorizontalHeaders(buf, headerBounds)
	}

	// Draw active tab content (set layout bounds to contentBounds)
	for _, item := range t.list {
		if item.Value == t.value && item.Content != nil {
			_ = contentBounds
			item.Content.Draw(buf)
		}
	}
}

func (t *Tabs) drawHeaderBorder(buf *render.CommandBuffer, hb uimath.Rect) {
	cfg := t.config
	switch t.placement {
	case TabPlacementBottom:
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(hb.X, hb.Y, hb.Width, 1),
			FillColor: cfg.BorderColor,
		}, 0, 1)
	case TabPlacementLeft:
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(hb.X+hb.Width-1, hb.Y, 1, hb.Height),
			FillColor: cfg.BorderColor,
		}, 0, 1)
	case TabPlacementRight:
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(hb.X, hb.Y, 1, hb.Height),
			FillColor: cfg.BorderColor,
		}, 0, 1)
	default: // Top
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(hb.X, hb.Y+hb.Height-1, hb.Width, 1),
			FillColor: cfg.BorderColor,
		}, 0, 1)
	}
}

func (t *Tabs) drawHorizontalHeaders(buf *render.CommandBuffer, hb uimath.Rect) {
	cfg := t.config
	tabX := hb.X
	for i, item := range t.list {
		active := item.Value == t.value

		// Measure tab width
		tabW := float32(len(item.Label))*cfg.FontSize*0.55 + cfg.SpaceMD*2
		if cfg.TextRenderer != nil {
			tabW = cfg.TextRenderer.MeasureText(item.Label, cfg.FontSize) + cfg.SpaceMD*2
		}
		if item.Removable {
			tabW += tabCloseSize + cfg.SpaceXS
		}

		// Set bounds on tab header element so hit testing works
		if i < len(t.tabIDs) {
			t.tree.SetLayout(t.tabIDs[i], core.LayoutResult{
				Bounds: uimath.NewRect(tabX, hb.Y, tabW, hb.Height),
			})
			elem := t.tree.Get(t.tabIDs[i])
			hovered := elem != nil && elem.IsHovered()

			// Tab background
			if t.theme == TabThemeCard {
				t.drawCardTab(buf, tabX, hb.Y, tabW, hb.Height, active, hovered)
			} else {
				// Normal: hover highlight
				if hovered && !active {
					buf.DrawRect(render.RectCmd{
						Bounds:    uimath.NewRect(tabX, hb.Y, tabW, hb.Height-1),
						FillColor: uimath.ColorHex("#f3f3f3"),
					}, 0, 1)
				}
			}

			// Tab label
			textColor := cfg.TextColor
			if item.Disabled || t.disabled {
				textColor = cfg.DisabledColor
			} else if active {
				textColor = cfg.PrimaryColor
			} else if hovered {
				textColor = cfg.HoverColor
			}

			labelW := tabW - cfg.SpaceMD*2
			if item.Removable {
				labelW -= tabCloseSize + cfg.SpaceXS
			}

			if cfg.TextRenderer != nil {
				lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
				tx := tabX + cfg.SpaceMD
				ty := hb.Y + (hb.Height-lh)/2
				cfg.TextRenderer.DrawText(buf, item.Label, tx, ty, cfg.FontSize, labelW, textColor, 1)
			} else {
				textW := float32(len(item.Label)) * cfg.FontSize * 0.55
				textH := cfg.FontSize * 1.2
				tx := tabX + (tabW-textW)/2
				ty := hb.Y + (hb.Height-textH)/2
				buf.DrawRect(render.RectCmd{
					Bounds:    uimath.NewRect(tx, ty, textW, textH),
					FillColor: textColor,
					Corners:   uimath.CornersAll(2),
				}, 1, 1)
			}

			// Active indicator (underline) for normal style
			if active && t.theme == TabThemeNormal {
				buf.DrawRect(render.RectCmd{
					Bounds:    uimath.NewRect(tabX, hb.Y+hb.Height-2, tabW, 2),
					FillColor: cfg.PrimaryColor,
				}, 1, 1)
			}

			// Remove button
			if item.Removable {
				t.drawCloseButton(buf, tabX+tabW-cfg.SpaceMD-tabCloseSize, hb.Y+(hb.Height-tabCloseSize)/2, item.Value, i)
			}
		}

		tabX += tabW
	}

	// "+" add button
	if t.addable {
		addW := tabHeaderHeight
		if t.addBtnID != 0 {
			t.tree.SetLayout(t.addBtnID, core.LayoutResult{
				Bounds: uimath.NewRect(tabX, hb.Y, addW, hb.Height),
			})
		}
		t.drawAddButton(buf, tabX, hb.Y, addW, hb.Height)
	}
}

func (t *Tabs) drawVerticalHeaders(buf *render.CommandBuffer, hb uimath.Rect) {
	cfg := t.config
	tabY := hb.Y
	for i, item := range t.list {
		active := item.Value == t.value

		tabH := t.verticalTabHeight()
		tabW := hb.Width

		if i < len(t.tabIDs) {
			t.tree.SetLayout(t.tabIDs[i], core.LayoutResult{
				Bounds: uimath.NewRect(hb.X, tabY, tabW, tabH),
			})
			elem := t.tree.Get(t.tabIDs[i])
			hovered := elem != nil && elem.IsHovered()

			// Backgrounds
			if active {
				buf.DrawRect(render.RectCmd{
					Bounds:    uimath.NewRect(hb.X, tabY, tabW, tabH),
					FillColor: uimath.ColorHex("#f2f3ff"),
				}, 0, 1)
			} else if hovered {
				buf.DrawRect(render.RectCmd{
					Bounds:    uimath.NewRect(hb.X, tabY, tabW, tabH),
					FillColor: uimath.ColorHex("#f3f3f3"),
				}, 0, 1)
			}

			textColor := cfg.TextColor
			if item.Disabled || t.disabled {
				textColor = cfg.DisabledColor
			} else if active {
				textColor = cfg.PrimaryColor
			} else if hovered {
				textColor = cfg.HoverColor
			}

			if cfg.TextRenderer != nil {
				lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
				cfg.TextRenderer.DrawText(buf, item.Label, hb.X+cfg.SpaceMD, tabY+(tabH-lh)/2, cfg.FontSize, tabW-cfg.SpaceMD*2, textColor, 1)
			} else {
				tw := float32(len(item.Label)) * cfg.FontSize * 0.55
				th := cfg.FontSize * 1.2
				buf.DrawRect(render.RectCmd{
					Bounds:    uimath.NewRect(hb.X+cfg.SpaceMD, tabY+(tabH-th)/2, tw, th),
					FillColor: textColor,
					Corners:   uimath.CornersAll(2),
				}, 1, 1)
			}

			// Active indicator (vertical bar)
			if active && t.theme == TabThemeNormal {
				if t.placement == TabPlacementLeft {
					buf.DrawRect(render.RectCmd{
						Bounds:    uimath.NewRect(hb.X+hb.Width-2, tabY, 2, tabH),
						FillColor: cfg.PrimaryColor,
					}, 1, 1)
				} else {
					buf.DrawRect(render.RectCmd{
						Bounds:    uimath.NewRect(hb.X, tabY, 2, tabH),
						FillColor: cfg.PrimaryColor,
					}, 1, 1)
				}
			}

			// Remove button
			if item.Removable {
				t.drawCloseButton(buf, hb.X+tabW-cfg.SpaceMD-tabCloseSize, tabY+(tabH-tabCloseSize)/2, item.Value, i)
			}
		}

		tabY += tabH
	}

	// "+" add button
	if t.addable {
		addH := t.verticalTabHeight()
		if t.addBtnID != 0 {
			t.tree.SetLayout(t.addBtnID, core.LayoutResult{
				Bounds: uimath.NewRect(hb.X, tabY, hb.Width, addH),
			})
		}
		t.drawAddButton(buf, hb.X, tabY, hb.Width, addH)
	}
}

func (t *Tabs) verticalTabHeight() float32 { return tabHeaderHeight }

// drawCardTab draws a single card-style tab header.
func (t *Tabs) drawCardTab(buf *render.CommandBuffer, x, y, w, h float32, active, hovered bool) {
	borderClr := t.config.BorderColor
	radius := float32(4)

	if active {
		// Active card: white bg, border on top/left/right, no bottom border
		// Draw white background covering the bottom border line
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x, y, w, h),
			FillColor: uimath.ColorWhite,
			Corners:   uimath.Corners{TopLeft: radius, TopRight: radius},
		}, 0, 1)
		// Top border
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x, y, w, 1),
			FillColor: borderClr,
		}, 1, 1)
		// Left border
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x, y, 1, h),
			FillColor: borderClr,
		}, 1, 1)
		// Right border
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x+w-1, y, 1, h),
			FillColor: borderClr,
		}, 1, 1)
	} else if hovered {
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x, y, w, h),
			FillColor: uimath.ColorHex("#f3f3f3"),
		}, 0, 1)
	}
}

// drawCloseButton draws a small "x" close/remove button.
func (t *Tabs) drawCloseButton(buf *render.CommandBuffer, x, y float32, value string, index int) {
	cfg := t.config
	s := tabCloseSize
	clr := uimath.ColorHex("#8b8b8b")
	// Draw two diagonal lines as small rects to approximate "x"
	for i := 0; i < int(s); i++ {
		fi := float32(i)
		// "\" diagonal
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x+fi, y+fi, 1, 1),
			FillColor: clr,
		}, 2, 1)
		// "/" diagonal
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x+s-1-fi, y+fi, 1, 1),
			FillColor: clr,
		}, 2, 1)
	}
	_ = cfg
	_ = value
	_ = index
}

// drawAddButton draws the "+" add-tab button.
func (t *Tabs) drawAddButton(buf *render.CommandBuffer, x, y, w, h float32) {
	cfg := t.config
	clr := uimath.ColorHex("#8b8b8b")

	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, "+", x+(w-cfg.FontSize*0.55)/2, y+(h-lh)/2, cfg.FontSize, w, clr, 1)
	} else {
		// Draw "+" as two crossing rects
		cx := x + w/2
		cy := y + h/2
		barLen := float32(10)
		barThick := float32(2)
		// Horizontal bar
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(cx-barLen/2, cy-barThick/2, barLen, barThick),
			FillColor: clr,
		}, 2, 1)
		// Vertical bar
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(cx-barThick/2, cy-barLen/2, barThick, barLen),
			FillColor: clr,
		}, 2, 1)
	}
}

// Destroy cleans up tab header elements.
func (t *Tabs) Destroy() {
	for _, tabID := range t.tabIDs {
		t.tree.DestroyElement(tabID)
	}
	if t.addBtnID != 0 {
		t.tree.DestroyElement(t.addBtnID)
	}
	t.Base.Destroy()
}

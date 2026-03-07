package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// ContextMenuItem represents a single menu item.
type ContextMenuItem struct {
	Label    string
	Icon     string
	Disabled bool
	Divider  bool
	OnClick  func()
}

// ContextMenu is a right-click context menu.
type ContextMenu struct {
	Base
	items   []ContextMenuItem
	visible bool
	x, y    float32
	width   float32
}

func NewContextMenu(tree *core.Tree, cfg *Config) *ContextMenu {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &ContextMenu{
		Base:  NewBase(tree, core.TypeCustom, cfg),
		width: 180,
	}
}

func (cm *ContextMenu) IsVisible() bool { return cm.visible }
func (cm *ContextMenu) SetWidth(w float32) { cm.width = w }

func (cm *ContextMenu) AddItem(item ContextMenuItem) {
	cm.items = append(cm.items, item)
}

func (cm *ContextMenu) AddDivider() {
	cm.items = append(cm.items, ContextMenuItem{Divider: true})
}

func (cm *ContextMenu) ClearItems() {
	cm.items = cm.items[:0]
}

func (cm *ContextMenu) Items() []ContextMenuItem { return cm.items }

func (cm *ContextMenu) Show(x, y float32) {
	cm.x = x
	cm.y = y
	cm.visible = true
}

func (cm *ContextMenu) Hide() {
	cm.visible = false
}

func (cm *ContextMenu) Draw(buf *render.CommandBuffer) {
	if !cm.visible || len(cm.items) == 0 {
		return
	}
	cfg := cm.config
	itemH := float32(32)
	divH := float32(9)
	totalH := float32(0)
	for _, item := range cm.items {
		if item.Divider {
			totalH += divH
		} else {
			totalH += itemH
		}
	}
	totalH += cfg.SpaceXS * 2 // padding top/bottom

	// Shadow
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(cm.x+2, cm.y+2, cm.width, totalH),
		FillColor: uimath.RGBA(0, 0, 0, 0.12),
		Corners:   uimath.CornersAll(cfg.BorderRadius),
	}, 60, 1)

	// Background
	buf.DrawOverlay(render.RectCmd{
		Bounds:      uimath.NewRect(cm.x, cm.y, cm.width, totalH),
		FillColor:   uimath.ColorWhite,
		BorderColor: cfg.BorderColor,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 61, 1)

	// Items
	cy := cm.y + cfg.SpaceXS
	for _, item := range cm.items {
		if item.Divider {
			buf.DrawOverlay(render.RectCmd{
				Bounds:    uimath.NewRect(cm.x+cfg.SpaceSM, cy+4, cm.width-cfg.SpaceSM*2, 1),
				FillColor: cfg.BorderColor,
			}, 62, 1)
			cy += divH
			continue
		}
		color := cfg.TextColor
		if item.Disabled {
			color = cfg.DisabledColor
		}
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			cfg.TextRenderer.DrawText(buf, item.Label, cm.x+cfg.SpaceMD, cy+(itemH-lh)/2, cfg.FontSize, cm.width-cfg.SpaceMD*2, color, 1)
		}
		cy += itemH
	}
}

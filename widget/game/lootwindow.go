package game

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// LootItem represents a single loot entry.
type LootItem struct {
	Item     *ItemData
	Quantity int
	Claimed  bool
}

// LootWindow displays items available for pickup from a defeated enemy or chest.
type LootWindow struct {
	widget.Base
	title    string
	items    []LootItem
	visible  bool
	width    float32
	slotSize float32
	gap      float32
	onLoot   func(index int)
	onClose  func()
}

func NewLootWindow(tree *core.Tree, cfg *widget.Config) *LootWindow {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	return &LootWindow{
		Base:     widget.NewBase(tree, core.TypeCustom, cfg),
		title:    "Loot",
		width:    220,
		slotSize: 40,
		gap:      4,
	}
}

func (lw *LootWindow) Title() string           { return lw.title }
func (lw *LootWindow) Items() []LootItem       { return lw.items }
func (lw *LootWindow) IsVisible() bool         { return lw.visible }
func (lw *LootWindow) SetTitle(t string)       { lw.title = t }
func (lw *LootWindow) SetWidth(w float32)      { lw.width = w }
func (lw *LootWindow) OnLoot(fn func(int))     { lw.onLoot = fn }
func (lw *LootWindow) OnClose(fn func())       { lw.onClose = fn }

func (lw *LootWindow) AddItem(item LootItem) {
	lw.items = append(lw.items, item)
}

func (lw *LootWindow) ClearItems() {
	lw.items = lw.items[:0]
}

func (lw *LootWindow) Open() {
	lw.visible = true
}

func (lw *LootWindow) Close() {
	lw.visible = false
	if lw.onClose != nil {
		lw.onClose()
	}
}

func (lw *LootWindow) LootItem(index int) {
	if index >= 0 && index < len(lw.items) && !lw.items[index].Claimed {
		lw.items[index].Claimed = true
		if lw.onLoot != nil {
			lw.onLoot(index)
		}
	}
}

func (lw *LootWindow) LootAll() {
	for i := range lw.items {
		if !lw.items[i].Claimed {
			lw.items[i].Claimed = true
			if lw.onLoot != nil {
				lw.onLoot(i)
			}
		}
	}
}

func (lw *LootWindow) Draw(buf *render.CommandBuffer) {
	if !lw.visible || len(lw.items) == 0 {
		return
	}
	cfg := lw.Config()
	headerH := float32(32)
	itemH := lw.slotSize + lw.gap
	totalH := headerH + float32(len(lw.items))*itemH + cfg.SpaceSM
	bounds := lw.Bounds()
	x, y := bounds.X, bounds.Y
	if bounds.IsEmpty() {
		x, y = 0, 0
	}

	// Shadow + background
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(x+2, y+2, lw.width, totalH),
		FillColor: uimath.RGBA(0, 0, 0, 0.15),
		Corners:   uimath.CornersAll(cfg.BorderRadius),
	}, 40, 1)
	buf.DrawOverlay(render.RectCmd{
		Bounds:      uimath.NewRect(x, y, lw.width, totalH),
		FillColor:   uimath.RGBA(0.08, 0.08, 0.12, 0.95),
		BorderColor: uimath.RGBA(0.4, 0.4, 0.5, 0.8),
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 41, 1)

	// Title
	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, lw.title, x+cfg.SpaceSM, y+(headerH-lh)/2, cfg.FontSize, lw.width-cfg.SpaceSM*2, uimath.ColorHex("#ffd700"), 1)
	}

	// Items
	for i, li := range lw.items {
		iy := y + headerH + float32(i)*itemH
		s := lw.slotSize

		// Slot background
		bgColor := uimath.RGBA(0.15, 0.15, 0.15, 0.85)
		if li.Claimed {
			bgColor = uimath.RGBA(0.1, 0.1, 0.1, 0.5)
		}
		borderColor := uimath.RGBA(0.3, 0.3, 0.3, 1)
		if li.Item != nil {
			borderColor = rarityColor(li.Item.Rarity)
		}
		buf.DrawOverlay(render.RectCmd{
			Bounds:      uimath.NewRect(x+cfg.SpaceSM, iy, s, s),
			FillColor:   bgColor,
			BorderColor: borderColor,
			BorderWidth: 1,
			Corners:     uimath.CornersAll(4),
		}, 42, 1)

		// Item name
		if li.Item != nil && cfg.TextRenderer != nil {
			nameColor := rarityColor(li.Item.Rarity)
			if li.Claimed {
				nameColor = uimath.RGBA(0.4, 0.4, 0.4, 1)
			}
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			cfg.TextRenderer.DrawText(buf, li.Item.Name, x+cfg.SpaceSM+s+cfg.SpaceXS, iy+(s-lh)/2, cfg.FontSizeSm, lw.width-s-cfg.SpaceSM*2-cfg.SpaceXS, nameColor, 1)
		}
	}
}

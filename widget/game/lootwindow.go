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
	itemH := lw.slotSize + lw.gap
	contentH := float32(len(lw.items))*itemH + cfg.SpaceSM

	panel := Panel{
		Title:   lw.title,
		Width:   lw.width,
		TitleH:  32,
		BgColor: uimath.RGBA(0.08, 0.08, 0.12, 0.95),
		Shadow:  true,
	}
	r := panel.Draw(buf, lw.Bounds(), cfg, contentH)

	// Items
	for i, li := range lw.items {
		iy := r.ContentY + float32(i)*itemH
		s := lw.slotSize

		bgColor := uimath.RGBA(0.15, 0.15, 0.15, 0.85)
		if li.Claimed {
			bgColor = uimath.RGBA(0.1, 0.1, 0.1, 0.5)
		}
		borderColor := uimath.RGBA(0.3, 0.3, 0.3, 1)
		if li.Item != nil {
			borderColor = rarityColor(li.Item.Rarity)
		}
		buf.DrawRect(render.RectCmd{
			Bounds:      uimath.NewRect(r.ContentX, iy, s, s),
			FillColor:   bgColor,
			BorderColor: borderColor,
			BorderWidth: 1,
			Corners:     uimath.CornersAll(4),
		}, 4, 1)

		if li.Item != nil && cfg.TextRenderer != nil {
			nameColor := rarityColor(li.Item.Rarity)
			if li.Claimed {
				nameColor = uimath.RGBA(0.4, 0.4, 0.4, 1)
			}
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			cfg.TextRenderer.DrawText(buf, li.Item.Name, r.ContentX+s+cfg.SpaceXS, iy+(s-lh)/2, cfg.FontSizeSm, r.ContentW-s-cfg.SpaceXS, nameColor, 1)
		}
	}
}

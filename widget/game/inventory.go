package game

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// ItemData represents an item in the inventory.
type ItemData struct {
	ID       string
	Name     string
	Icon     render.TextureHandle
	Quantity int
	Rarity   ItemRarity
}

// ItemRarity determines the border color of an item slot.
type ItemRarity uint8

const (
	RarityCommon    ItemRarity = iota
	RarityUncommon
	RarityRare
	RarityEpic
	RarityLegendary
)

// Inventory is a grid-based item container (bag, chest, etc.).
type Inventory struct {
	widget.Base
	rows     int
	cols     int
	slotSize float32
	gap      float32
	items    map[int]*ItemData // slot index -> item
	title    string
	onDrop   func(slotIndex int, data any)
	onSelect func(slotIndex int, item *ItemData)
}

// NewInventory creates an inventory grid.
func NewInventory(tree *core.Tree, rows, cols int, cfg *widget.Config) *Inventory {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	return &Inventory{
		Base:     widget.NewBase(tree, core.TypeCustom, cfg),
		rows:     rows,
		cols:     cols,
		slotSize: 48,
		gap:      4,
		items:    make(map[int]*ItemData),
	}
}

func (inv *Inventory) Rows() int                  { return inv.rows }
func (inv *Inventory) Cols() int                   { return inv.cols }
func (inv *Inventory) Title() string               { return inv.title }
func (inv *Inventory) SetTitle(t string)            { inv.title = t }
func (inv *Inventory) SetSlotSize(s float32)        { inv.slotSize = s }
func (inv *Inventory) SetGap(g float32)             { inv.gap = g }
func (inv *Inventory) OnDrop(fn func(int, any))     { inv.onDrop = fn }
func (inv *Inventory) OnSelect(fn func(int, *ItemData)) { inv.onSelect = fn }

func (inv *Inventory) SetItem(slot int, item *ItemData) {
	if slot >= 0 && slot < inv.rows*inv.cols {
		inv.items[slot] = item
	}
}

func (inv *Inventory) GetItem(slot int) *ItemData {
	return inv.items[slot]
}

func (inv *Inventory) RemoveItem(slot int) *ItemData {
	item := inv.items[slot]
	delete(inv.items, slot)
	return item
}

func (inv *Inventory) ClearAll() {
	inv.items = make(map[int]*ItemData)
}

func rarityColor(r ItemRarity) uimath.Color {
	switch r {
	case RarityUncommon:
		return uimath.ColorHex("#1eff00")
	case RarityRare:
		return uimath.ColorHex("#0070dd")
	case RarityEpic:
		return uimath.ColorHex("#a335ee")
	case RarityLegendary:
		return uimath.ColorHex("#ff8000")
	default:
		return uimath.RGBA(0.5, 0.5, 0.5, 1)
	}
}

func (inv *Inventory) Draw(buf *render.CommandBuffer) {
	bounds := inv.Bounds()
	if bounds.IsEmpty() {
		return
	}

	cfg := inv.Config()
	s := inv.slotSize
	pad := float32(8)
	titleH := float32(0)

	// Panel background
	totalW := float32(inv.cols)*s + float32(inv.cols-1)*inv.gap + pad*2
	totalH := float32(inv.rows)*s + float32(inv.rows-1)*inv.gap + pad*2
	if inv.title != "" {
		titleH = 28
		totalH += titleH
	}

	buf.DrawRect(render.RectCmd{
		Bounds:      uimath.NewRect(bounds.X, bounds.Y, totalW, totalH),
		FillColor:   uimath.RGBA(0.1, 0.1, 0.1, 0.9),
		BorderColor: uimath.RGBA(0.3, 0.3, 0.3, 1),
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 1, 1)

	// Title
	if inv.title != "" && cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, inv.title, bounds.X+pad, bounds.Y+(titleH-lh)/2, cfg.FontSize, totalW-pad*2, uimath.ColorWhite, 1)
	}

	// Slots
	for r := 0; r < inv.rows; r++ {
		for c := 0; c < inv.cols; c++ {
			idx := r*inv.cols + c
			sx := bounds.X + pad + float32(c)*(s+inv.gap)
			sy := bounds.Y + pad + titleH + float32(r)*(s+inv.gap)

			// Slot background
			bgColor := uimath.RGBA(0.2, 0.2, 0.2, 0.8)
			borderColor := uimath.RGBA(0.35, 0.35, 0.35, 1)

			item := inv.items[idx]
			if item != nil {
				borderColor = rarityColor(item.Rarity)
			}

			buf.DrawRect(render.RectCmd{
				Bounds:      uimath.NewRect(sx, sy, s, s),
				FillColor:   bgColor,
				BorderColor: borderColor,
				BorderWidth: 1,
				Corners:     uimath.CornersAll(3),
			}, 2, 1)

			// Item icon placeholder
			if item != nil && item.Icon == 0 {
				// Draw a colored square as placeholder
				iconPad := float32(4)
				buf.DrawRect(render.RectCmd{
					Bounds:    uimath.NewRect(sx+iconPad, sy+iconPad, s-iconPad*2, s-iconPad*2),
					FillColor: rarityColor(item.Rarity),
					Corners:   uimath.CornersAll(2),
				}, 3, 0.3)
			}

			// Quantity badge
			if item != nil && item.Quantity > 1 && cfg.TextRenderer != nil {
				qText := itoa(item.Quantity)
				lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
				tw := cfg.TextRenderer.MeasureText(qText, cfg.FontSizeSm)
				cfg.TextRenderer.DrawText(buf, qText, sx+s-tw-2, sy+s-lh-2, cfg.FontSizeSm, s, uimath.ColorWhite, 1)
			}
		}
	}
}

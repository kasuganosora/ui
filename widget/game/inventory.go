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
	embedded bool // if true, skip Panel chrome (used inside Window)
	onDrop   func(slotIndex int, data any)
	onSelect func(slotIndex int, item *ItemData)

	// Drag state for item rearrangement
	dragSlot   int     // slot being dragged, or -1
	dragMouseX float32 // current mouse position
	dragMouseY float32
	hoverSlot  int // slot under cursor during drag, or -1
}

// NewInventory creates an inventory grid.
func NewInventory(tree *core.Tree, rows, cols int, cfg *widget.Config) *Inventory {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	return &Inventory{
		Base:      widget.NewBase(tree, core.TypeCustom, cfg),
		rows:      rows,
		cols:      cols,
		slotSize:  48,
		gap:       4,
		items:     make(map[int]*ItemData),
		dragSlot:  -1,
		hoverSlot: -1,
	}
}

func (inv *Inventory) Rows() int                  { return inv.rows }
func (inv *Inventory) Cols() int                   { return inv.cols }
func (inv *Inventory) Title() string               { return inv.title }
func (inv *Inventory) SetTitle(t string)            { inv.title = t }
func (inv *Inventory) SetEmbedded(v bool)           { inv.embedded = v }
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

// IsDragging reports whether an item drag is in progress.
func (inv *Inventory) IsDragging() bool { return inv.dragSlot >= 0 }

// contentOrigin returns the top-left of the grid area (below title bar + padding).
func (inv *Inventory) contentOrigin() (float32, float32) {
	b := inv.Bounds()
	pad := float32(8)
	titleH := float32(0)
	if inv.title != "" {
		titleH = 28
	}
	return b.X + pad, b.Y + titleH
}

// slotAt returns the slot index at pixel (x, y), or -1 if none.
func (inv *Inventory) slotAt(x, y float32) int {
	cx, cy := inv.contentOrigin()
	s := inv.slotSize
	g := inv.gap
	col := int((x - cx) / (s + g))
	row := int((y - cy) / (s + g))
	if col < 0 || col >= inv.cols || row < 0 || row >= inv.rows {
		return -1
	}
	// Verify we're inside the slot, not in the gap
	sx := cx + float32(col)*(s+g)
	sy := cy + float32(row)*(s+g)
	if x >= sx && x < sx+s && y >= sy && y < sy+s {
		return row*inv.cols + col
	}
	return -1
}

// HandleMouseDown starts an item drag if clicking a slot that has an item.
// Returns true if an item drag was started.
func (inv *Inventory) HandleMouseDown(x, y float32) bool {
	slot := inv.slotAt(x, y)
	if slot < 0 {
		return false
	}
	item := inv.items[slot]
	if item == nil {
		return false
	}
	inv.dragSlot = slot
	inv.dragMouseX = x
	inv.dragMouseY = y
	inv.hoverSlot = slot
	if inv.onSelect != nil {
		inv.onSelect(slot, item)
	}
	return true
}

// HandleMouseMove updates drag position. Returns true if dragging.
func (inv *Inventory) HandleMouseMove(x, y float32) bool {
	if inv.dragSlot < 0 {
		return false
	}
	inv.dragMouseX = x
	inv.dragMouseY = y
	inv.hoverSlot = inv.slotAt(x, y)
	return true
}

// HandleMouseUp ends the drag. If dropped on a different slot, swaps the items.
// Returns true if a drag was active.
func (inv *Inventory) HandleMouseUp(x, y float32) bool {
	if inv.dragSlot < 0 {
		return false
	}
	target := inv.slotAt(x, y)
	if target >= 0 && target != inv.dragSlot {
		// Swap items between slots
		inv.items[inv.dragSlot], inv.items[target] = inv.items[target], inv.items[inv.dragSlot]
		if inv.onDrop != nil {
			inv.onDrop(target, inv.items[target])
		}
	}
	inv.dragSlot = -1
	inv.hoverSlot = -1
	return true
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
	gridW := float32(inv.cols)*s + float32(inv.cols-1)*inv.gap
	gridH := float32(inv.rows)*s + float32(inv.rows-1)*inv.gap

	var contentX, contentY, contentW float32

	if inv.embedded {
		// Embedded mode: no Panel chrome
		pad := float32(8)
		contentX = bounds.X + pad
		contentY = bounds.Y + pad
		contentW = bounds.Width - pad*2
		if contentW <= 0 {
			contentW = gridW
		}
	} else {
		titleH := float32(-1) // no title bar
		if inv.title != "" {
			titleH = 28
		}
		panel := Panel{
			Title:       inv.title,
			Width:       gridW + 16,
			TitleH:      titleH,
			BgColor:     uimath.RGBA(0.1, 0.1, 0.1, 0.9),
			BorderColor: uimath.RGBA(0.3, 0.3, 0.3, 1),
			TitleColor:  uimath.ColorWhite,
		}
		r := panel.Draw(buf, bounds, cfg, gridH+8)
		contentX, contentY, contentW = r.ContentX, r.ContentY, r.ContentW
	}
	_ = contentW // used for future layout

	// Slots
	for row := 0; row < inv.rows; row++ {
		for c := 0; c < inv.cols; c++ {
			idx := row*inv.cols + c
			sx := contentX + float32(c)*(s+inv.gap)
			sy := contentY + float32(row)*(s+inv.gap)

			// Slot background
			bgColor := uimath.RGBA(0.2, 0.2, 0.2, 0.8)
			borderColor := uimath.RGBA(0.35, 0.35, 0.35, 1)

			item := inv.items[idx]
			if item != nil {
				borderColor = rarityColor(item.Rarity)
			}

			// Highlight drop target during drag
			if inv.dragSlot >= 0 && idx == inv.hoverSlot && idx != inv.dragSlot {
				bgColor = uimath.RGBA(0.3, 0.3, 0.4, 0.9)
				borderColor = uimath.ColorHex("#ffd700")
			}

			buf.DrawRect(render.RectCmd{
				Bounds:      uimath.NewRect(sx, sy, s, s),
				FillColor:   bgColor,
				BorderColor: borderColor,
				BorderWidth: 1,
				Corners:     uimath.CornersAll(3),
			}, 2, 1)

			// Skip drawing the item being dragged in its original slot
			if inv.dragSlot >= 0 && idx == inv.dragSlot {
				continue
			}

			// Item icon placeholder
			if item != nil && item.Icon == 0 {
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

	// Draw dragged item at cursor
	if inv.dragSlot >= 0 {
		if dragItem := inv.items[inv.dragSlot]; dragItem != nil {
			half := s / 2
			dx := inv.dragMouseX - half
			dy := inv.dragMouseY - half
			if dragItem.Icon == 0 {
				iconPad := float32(4)
				buf.DrawRect(render.RectCmd{
					Bounds:      uimath.NewRect(dx, dy, s, s),
					FillColor:   uimath.RGBA(0.2, 0.2, 0.2, 0.6),
					BorderColor: rarityColor(dragItem.Rarity),
					BorderWidth: 1,
					Corners:     uimath.CornersAll(3),
				}, 10, 0.8)
				buf.DrawRect(render.RectCmd{
					Bounds:    uimath.NewRect(dx+iconPad, dy+iconPad, s-iconPad*2, s-iconPad*2),
					FillColor: rarityColor(dragItem.Rarity),
					Corners:   uimath.CornersAll(2),
				}, 11, 0.5)
			}
			if dragItem.Quantity > 1 && cfg.TextRenderer != nil {
				qText := itoa(dragItem.Quantity)
				lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
				tw := cfg.TextRenderer.MeasureText(qText, cfg.FontSizeSm)
				cfg.TextRenderer.DrawText(buf, qText, dx+s-tw-2, dy+s-lh-2, cfg.FontSizeSm, s, uimath.ColorWhite, 1)
			}
		}
	}
}

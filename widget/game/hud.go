// Package game provides game-specific UI widgets.
//
// These widgets are designed for game HUDs, inventories, chat boxes,
// and other game-specific UI elements.
package game

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// HUD is a head-up display overlay layer.
// It manages multiple HUD elements positioned around the screen.
type HUD struct {
	widget.Base
	elements []HUDElement
}

// HUDElement is a positioned element within the HUD.
type HUDElement struct {
	Widget  widget.Widget
	Anchor  HUDAnchor
	OffsetX float32
	OffsetY float32
}

// HUDAnchor specifies where an element is anchored on screen.
type HUDAnchor uint8

const (
	AnchorTopLeft     HUDAnchor = iota
	AnchorTopCenter
	AnchorTopRight
	AnchorMiddleLeft
	AnchorMiddleCenter
	AnchorMiddleRight
	AnchorBottomLeft
	AnchorBottomCenter
	AnchorBottomRight
)

// NewHUD creates a HUD overlay.
func NewHUD(tree *core.Tree, cfg *widget.Config) *HUD {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	h := &HUD{
		Base: widget.NewBase(tree, core.TypeCustom, cfg),
	}
	h.SetStyle(layout.Style{Display: layout.DisplayNone})
	return h
}

// AddElement adds a widget to the HUD at the specified anchor.
func (h *HUD) AddElement(w widget.Widget, anchor HUDAnchor, offsetX, offsetY float32) {
	h.elements = append(h.elements, HUDElement{
		Widget:  w,
		Anchor:  anchor,
		OffsetX: offsetX,
		OffsetY: offsetY,
	})
}

// LayoutElements positions HUD elements based on viewport size.
func (h *HUD) LayoutElements(vpW, vpH float32) {
	for _, elem := range h.elements {
		var x, y float32
		switch elem.Anchor {
		case AnchorTopLeft:
			x, y = 0, 0
		case AnchorTopCenter:
			x, y = vpW/2, 0
		case AnchorTopRight:
			x, y = vpW, 0
		case AnchorMiddleLeft:
			x, y = 0, vpH/2
		case AnchorMiddleCenter:
			x, y = vpW/2, vpH/2
		case AnchorMiddleRight:
			x, y = vpW, vpH/2
		case AnchorBottomLeft:
			x, y = 0, vpH
		case AnchorBottomCenter:
			x, y = vpW/2, vpH
		case AnchorBottomRight:
			x, y = vpW, vpH
		}
		x += elem.OffsetX
		y += elem.OffsetY
		_ = x
		_ = y
	}
}

func (h *HUD) Draw(buf *render.CommandBuffer) {
	for _, elem := range h.elements {
		elem.Widget.Draw(buf)
	}
}

// HealthBar displays a horizontal resource bar (HP, MP, XP, etc.).
type HealthBar struct {
	widget.Base
	current    float32
	max        float32
	barColor   uimath.Color
	bgColor    uimath.Color
	showText   bool
	width      float32
	height     float32
}

// NewHealthBar creates a resource bar.
func NewHealthBar(tree *core.Tree, cfg *widget.Config) *HealthBar {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	return &HealthBar{
		Base:     widget.NewBase(tree, core.TypeCustom, cfg),
		current:  100,
		max:      100,
		barColor: uimath.ColorHex("#52c41a"),
		bgColor:  uimath.RGBA(0, 0, 0, 0.6),
		width:    200,
		height:   20,
	}
}

func (hb *HealthBar) Current() float32        { return hb.current }
func (hb *HealthBar) Max() float32             { return hb.max }

func (hb *HealthBar) SetCurrent(v float32)     { hb.current = v }
func (hb *HealthBar) SetMax(v float32)          { hb.max = v }
func (hb *HealthBar) SetBarColor(c uimath.Color) { hb.barColor = c }
func (hb *HealthBar) SetBgColor(c uimath.Color)  { hb.bgColor = c }
func (hb *HealthBar) SetShowText(v bool)         { hb.showText = v }
func (hb *HealthBar) SetSize(w, h float32)       { hb.width = w; hb.height = h }

func (hb *HealthBar) Ratio() float32 {
	if hb.max <= 0 {
		return 0
	}
	r := hb.current / hb.max
	if r < 0 {
		r = 0
	}
	if r > 1 {
		r = 1
	}
	return r
}

func (hb *HealthBar) Draw(buf *render.CommandBuffer) {
	bounds := hb.Bounds()
	if bounds.IsEmpty() {
		bounds = uimath.NewRect(0, 0, hb.width, hb.height)
	}

	cfg := hb.Config()
	radius := hb.height / 2

	// Background
	buf.DrawRect(render.RectCmd{
		Bounds:    bounds,
		FillColor: hb.bgColor,
		Corners:   uimath.CornersAll(radius),
	}, 1, 1)

	// Fill
	ratio := hb.Ratio()
	if ratio > 0 {
		fillW := bounds.Width * ratio
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(bounds.X, bounds.Y, fillW, bounds.Height),
			FillColor: hb.barColor,
			Corners:   uimath.CornersAll(radius),
		}, 2, 1)
	}

	// Text
	if hb.showText && cfg.TextRenderer != nil {
		text := formatFloat(hb.current) + "/" + formatFloat(hb.max)
		lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
		tw := cfg.TextRenderer.MeasureText(text, cfg.FontSizeSm)
		tx := bounds.X + (bounds.Width-tw)/2
		ty := bounds.Y + (bounds.Height-lh)/2
		cfg.TextRenderer.DrawText(buf, text, tx, ty, cfg.FontSizeSm, bounds.Width, uimath.ColorWhite, 1)
	}
}

// Hotbar is a row of action slots (abilities, items).
type Hotbar struct {
	widget.Base
	slots     []HotbarSlot
	slotSize  float32
	gap       float32
	selected  int
}

// HotbarSlot represents a single slot in the hotbar.
type HotbarSlot struct {
	Icon      render.TextureHandle
	Label     string
	Cooldown  float32 // 0-1, fraction remaining
	Keybind   string  // e.g., "1", "Q"
	Available bool
}

// NewHotbar creates a hotbar with the given number of slots.
func NewHotbar(tree *core.Tree, numSlots int, cfg *widget.Config) *Hotbar {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	slots := make([]HotbarSlot, numSlots)
	for i := range slots {
		slots[i].Available = true
	}
	return &Hotbar{
		Base:     widget.NewBase(tree, core.TypeCustom, cfg),
		slots:    slots,
		slotSize: 48,
		gap:      4,
		selected: -1,
	}
}

func (h *Hotbar) SlotCount() int           { return len(h.slots) }
func (h *Hotbar) Selected() int             { return h.selected }
func (h *Hotbar) SetSelected(i int)         { h.selected = i }
func (h *Hotbar) SetSlotSize(s float32)     { h.slotSize = s }
func (h *Hotbar) SetGap(g float32)          { h.gap = g }

func (h *Hotbar) SetSlot(index int, slot HotbarSlot) {
	if index >= 0 && index < len(h.slots) {
		h.slots[index] = slot
	}
}

func (h *Hotbar) GetSlot(index int) HotbarSlot {
	if index >= 0 && index < len(h.slots) {
		return h.slots[index]
	}
	return HotbarSlot{}
}

func (h *Hotbar) Draw(buf *render.CommandBuffer) {
	bounds := h.Bounds()
	if bounds.IsEmpty() {
		return
	}

	cfg := h.Config()
	x := bounds.X
	y := bounds.Y
	s := h.slotSize

	for i, slot := range h.slots {
		sx := x + float32(i)*(s+h.gap)

		// Slot background
		bgColor := uimath.RGBA(0.15, 0.15, 0.15, 0.85)
		borderColor := uimath.RGBA(0.4, 0.4, 0.4, 1)
		if i == h.selected {
			borderColor = uimath.ColorHex("#ffd700")
		}
		buf.DrawRect(render.RectCmd{
			Bounds:      uimath.NewRect(sx, y, s, s),
			FillColor:   bgColor,
			BorderColor: borderColor,
			BorderWidth: 1,
			Corners:     uimath.CornersAll(4),
		}, 1, 1)

		// Cooldown overlay
		if slot.Cooldown > 0 {
			coolH := s * slot.Cooldown
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(sx, y+s-coolH, s, coolH),
				FillColor: uimath.RGBA(0, 0, 0, 0.6),
			}, 3, 1)
		}

		// Keybind label
		if slot.Keybind != "" && cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			cfg.TextRenderer.DrawText(buf, slot.Keybind, sx+2, y+s-lh-2, cfg.FontSizeSm, s, uimath.RGBA(1, 1, 1, 0.7), 1)
		}
	}
}

// CooldownMask displays a cooldown overlay on a slot.
type CooldownMask struct {
	widget.Base
	ratio float32 // 0-1
}

// NewCooldownMask creates a cooldown mask overlay.
func NewCooldownMask(tree *core.Tree, cfg *widget.Config) *CooldownMask {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	return &CooldownMask{
		Base: widget.NewBase(tree, core.TypeCustom, cfg),
	}
}

func (cm *CooldownMask) SetRatio(r float32)  { cm.ratio = r }
func (cm *CooldownMask) Ratio() float32       { return cm.ratio }

func (cm *CooldownMask) Draw(buf *render.CommandBuffer) {
	if cm.ratio <= 0 {
		return
	}
	bounds := cm.Bounds()
	if bounds.IsEmpty() {
		return
	}
	h := bounds.Height * cm.ratio
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(bounds.X, bounds.Y+bounds.Height-h, bounds.Width, h),
		FillColor: uimath.RGBA(0, 0, 0, 0.6),
	}, 5, 1)
}

// helper
func formatFloat(v float32) string {
	i := int(v)
	if float32(i) == v {
		return itoa(i)
	}
	return itoa(i)
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := false
	if i < 0 {
		neg = true
		i = -i
	}
	buf := make([]byte, 0, 10)
	for i > 0 {
		buf = append(buf, byte('0'+i%10))
		i /= 10
	}
	if neg {
		buf = append(buf, '-')
	}
	// reverse
	for l, r := 0, len(buf)-1; l < r; l, r = l+1, r-1 {
		buf[l], buf[r] = buf[r], buf[l]
	}
	return string(buf)
}

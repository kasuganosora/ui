package game

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// RadialMenuItem represents a slice of the radial menu.
type RadialMenuItem struct {
	Label    string
	Icon     render.TextureHandle
	Disabled bool
	OnClick  func()
}

// RadialMenu is a circular context menu with items around a center.
type RadialMenu struct {
	widget.Base
	items     []RadialMenuItem
	visible   bool
	cx, cy    float32
	radius    float32
	innerR    float32
	hovered   int
	onSelect  func(int)
}

func NewRadialMenu(tree *core.Tree, cfg *widget.Config) *RadialMenu {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	return &RadialMenu{
		Base:    widget.NewBase(tree, core.TypeCustom, cfg),
		radius:  120,
		innerR:  40,
		hovered: -1,
	}
}

func (rm *RadialMenu) IsVisible() bool           { return rm.visible }
func (rm *RadialMenu) SetRadius(r float32)        { rm.radius = r }
func (rm *RadialMenu) SetInnerRadius(r float32)   { rm.innerR = r }
func (rm *RadialMenu) Hovered() int               { return rm.hovered }
func (rm *RadialMenu) SetHovered(h int)           { rm.hovered = h }
func (rm *RadialMenu) OnSelect(fn func(int))      { rm.onSelect = fn }
func (rm *RadialMenu) Items() []RadialMenuItem    { return rm.items }

func (rm *RadialMenu) AddItem(item RadialMenuItem) {
	rm.items = append(rm.items, item)
}

func (rm *RadialMenu) ClearItems() {
	rm.items = rm.items[:0]
}

func (rm *RadialMenu) Show(cx, cy float32) {
	rm.cx = cx
	rm.cy = cy
	rm.visible = true
	rm.hovered = -1
}

func (rm *RadialMenu) Hide() {
	rm.visible = false
}

func (rm *RadialMenu) Select(index int) {
	if index >= 0 && index < len(rm.items) && !rm.items[index].Disabled {
		if rm.items[index].OnClick != nil {
			rm.items[index].OnClick()
		}
		if rm.onSelect != nil {
			rm.onSelect(index)
		}
	}
	rm.Hide()
}

func (rm *RadialMenu) Draw(buf *render.CommandBuffer) {
	if !rm.visible || len(rm.items) == 0 {
		return
	}
	cfg := rm.Config()
	n := len(rm.items)

	// Outer circle background
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(rm.cx-rm.radius, rm.cy-rm.radius, rm.radius*2, rm.radius*2),
		FillColor: uimath.RGBA(0.1, 0.1, 0.1, 0.85),
		Corners:   uimath.CornersAll(rm.radius),
	}, 70, 1)

	// Inner circle
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(rm.cx-rm.innerR, rm.cy-rm.innerR, rm.innerR*2, rm.innerR*2),
		FillColor: uimath.RGBA(0.2, 0.2, 0.2, 0.9),
		Corners:   uimath.CornersAll(rm.innerR),
	}, 71, 1)

	// Item labels positioned around the circle
	midR := (rm.radius + rm.innerR) / 2
	for i, item := range rm.items {
		// Angle for this item (evenly distributed)
		angle := float32(i) * 6.2831853 / float32(n) // 2*PI / n
		ix := rm.cx + midR*sinApprox(angle)
		iy := rm.cy - midR*cosApprox(angle)

		// Highlight hovered
		dotSize := float32(8)
		if i == rm.hovered {
			dotSize = 12
		}
		color := uimath.RGBA(0.8, 0.8, 0.8, 1)
		if item.Disabled {
			color = uimath.RGBA(0.4, 0.4, 0.4, 1)
		} else if i == rm.hovered {
			color = uimath.ColorHex("#ffd700")
		}

		buf.DrawOverlay(render.RectCmd{
			Bounds:    uimath.NewRect(ix-dotSize/2, iy-dotSize/2, dotSize, dotSize),
			FillColor: color,
			Corners:   uimath.CornersAll(dotSize / 2),
		}, 72, 1)

		// Label
		if cfg.TextRenderer != nil && item.Label != "" {
			tw := cfg.TextRenderer.MeasureText(item.Label, cfg.FontSizeSm)
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			cfg.TextRenderer.DrawText(buf, item.Label, ix-tw/2, iy+dotSize/2+2, cfg.FontSizeSm, tw+4, color, 1)
			_ = lh
		}
	}
}

// Simple sin/cos approximations for layout (avoid math import)
func sinApprox(x float32) float32 {
	// Normalize to [-PI, PI]
	for x > 3.14159 {
		x -= 6.28318
	}
	for x < -3.14159 {
		x += 6.28318
	}
	// Taylor series: sin(x) ≈ x - x³/6 + x⁵/120
	x3 := x * x * x
	x5 := x3 * x * x
	return x - x3/6 + x5/120
}

func cosApprox(x float32) float32 {
	return sinApprox(x + 1.5708)
}

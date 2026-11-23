package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// ColorPicker allows selecting a color from a preset swatch grid.
type ColorPicker struct {
	Base
	value        uimath.Color
	presets      []uimath.Color
	open         bool
	onChange     func(uimath.Color)
	cells        hitCells
	alphaEnabled bool
	recent       []uimath.Color
}

func NewColorPicker(tree *core.Tree, cfg *Config) *ColorPicker {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	cp := &ColorPicker{
		Base:  NewBase(tree, core.TypeCustom, cfg),
		value: cfg.PrimaryColor,
		presets: []uimath.Color{
			uimath.ColorHex("#f5222d"), uimath.ColorHex("#fa541c"), uimath.ColorHex("#fa8c16"),
			uimath.ColorHex("#faad14"), uimath.ColorHex("#fadb14"), uimath.ColorHex("#a0d911"),
			uimath.ColorHex("#52c41a"), uimath.ColorHex("#13c2c2"), uimath.ColorHex("#1677ff"),
			uimath.ColorHex("#2f54eb"), uimath.ColorHex("#722ed1"), uimath.ColorHex("#eb2f96"),
		},
	}
	cp.cells = hitCells{tree: tree, parent: cp.id}
	tree.AddHandler(cp.id, event.MouseClick, func(e *event.Event) {
		cp.onClick(e.GlobalX, e.GlobalY)
	})
	return cp
}

func (cp *ColorPicker) onClick(x, y float32) {
	b := cp.Bounds()
	sw := b.Height
	if sw > b.Width {
		sw = b.Width
	}
	if cp.open {
		const cols = 6
		swSize := float32(24)
		gap := float32(4)
		pad := float32(8)
		panelX := b.X
		panelY := b.Y + sw + 4
		for i := range cp.presets {
			r := i / cols
			col := i % cols
			sx := panelX + pad + float32(col)*(swSize+gap)
			sy := panelY + pad + float32(r)*(swSize+gap)
			if x >= sx && y >= sy && x < sx+swSize && y < sy+swSize {
				cp.SelectColor(cp.presets[i])
				cp.cells.clear()
				return
			}
		}
		if cp.alphaEnabled {
			barY := cp.AlphaBarY()
			panelW := float32(6)*24 + float32(5)*4 + 16
			if y >= barY && y < barY+16 && x >= b.X && x < b.X+panelW {
				cp.value.A = (x - b.X) / panelW
				cp.open = false
				cp.cells.clear()
				if cp.onChange != nil {
					cp.onChange(cp.value)
				}
				return
			}
		}
	}
	if pointIn(x, y, uimath.NewRect(b.X, b.Y, sw, sw)) {
		cp.open = !cp.open
		if cp.open {
			cp.placeCells(sw)
		} else {
			cp.cells.clear()
		}
	}
}

func (cp *ColorPicker) placeCells(sw float32) {
	b := cp.Bounds()
	const cols = 6
	swSize := float32(24)
	gap := float32(4)
	pad := float32(8)
	panelX := b.X
	panelY := b.Y + sw + 4
	rects := make([]uimath.Rect, len(cp.presets))
	for i := range cp.presets {
		r := i / cols
		col := i % cols
		rects[i] = uimath.NewRect(panelX+pad+float32(col)*(swSize+gap), panelY+pad+float32(r)*(swSize+gap), swSize, swSize)
	}
	cp.cells.set(rects, func(i int) {
		if i >= 0 && i < len(cp.presets) {
			cp.SelectColor(cp.presets[i])
			cp.cells.clear()
		}
	})
}

func (cp *ColorPicker) Value() uimath.Color            { return cp.value }
func (cp *ColorPicker) SetValue(c uimath.Color)        { cp.value = c }
func (cp *ColorPicker) SetPresets(p []uimath.Color)    { cp.presets = p }
func (cp *ColorPicker) OnChange(fn func(uimath.Color)) { cp.onChange = fn }
func (cp *ColorPicker) IsOpen() bool                   { return cp.open }

func (cp *ColorPicker) SelectColor(c uimath.Color) {
	if cp.alphaEnabled {
		c.A = cp.value.A
		if c.A == 0 {
			c.A = 1
		}
	}
	cp.value = c
	cp.remember(c)
	cp.open = false
	if cp.onChange != nil {
		cp.onChange(c)
	}
}

func (cp *ColorPicker) SetAlphaEnabled(v bool) { cp.alphaEnabled = v }

func (cp *ColorPicker) Alpha() float32 { return cp.value.A }

func (cp *ColorPicker) Recent() []uimath.Color {
	out := make([]uimath.Color, len(cp.recent))
	copy(out, cp.recent)
	return out
}

func (cp *ColorPicker) AlphaBarY() float32 {
	b := cp.Bounds()
	sw := b.Height
	if sw > b.Width {
		sw = b.Width
	}
	return b.Y + sw + 4 + cp.presetPanelH() + 4
}

func (cp *ColorPicker) remember(c uimath.Color) {
	next := []uimath.Color{c}
	for _, old := range cp.recent {
		if old == c || len(next) >= 8 {
			continue
		}
		next = append(next, old)
	}
	cp.recent = next
}

func (cp *ColorPicker) presetPanelH() float32 {
	const cols = 6
	swSize := float32(24)
	gap := float32(4)
	pad := float32(8)
	rows := (len(cp.presets) + cols - 1) / cols
	if rows < 1 {
		rows = 1
	}
	return float32(rows)*swSize + float32(rows-1)*gap + pad*2
}

func (cp *ColorPicker) Draw(buf *render.CommandBuffer) {
	bounds := cp.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := cp.config

	// Color swatch button
	swatchSize := bounds.Height
	if swatchSize > bounds.Width {
		swatchSize = bounds.Width
	}
	buf.DrawRect(render.RectCmd{
		Bounds:      uimath.NewRect(bounds.X, bounds.Y, swatchSize, swatchSize),
		FillColor:   cp.value,
		BorderColor: cfg.BorderColor,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 0, 1)

	if !cp.open {
		return
	}

	// Dropdown panel
	cols := 6
	swSize := float32(24)
	gap := float32(4)
	pad := float32(8)
	rows := (len(cp.presets) + cols - 1) / cols
	panelW := float32(cols)*swSize + float32(cols-1)*gap + pad*2
	panelH := float32(rows)*swSize + float32(rows-1)*gap + pad*2
	panelX := bounds.X
	panelY := bounds.Y + swatchSize + 4

	buf.DrawOverlay(render.RectCmd{
		Bounds:      uimath.NewRect(panelX, panelY, panelW, panelH),
		FillColor:   uimath.ColorWhite,
		BorderColor: cfg.BorderColor,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 20, 1)

	for i, c := range cp.presets {
		r := i / cols
		col := i % cols
		sx := panelX + pad + float32(col)*(swSize+gap)
		sy := panelY + pad + float32(r)*(swSize+gap)
		borderC := uimath.Color{}
		if c == cp.value {
			borderC = uimath.Color{R: 0, G: 0, B: 0, A: 1}
		}
		buf.DrawOverlay(render.RectCmd{
			Bounds:      uimath.NewRect(sx, sy, swSize, swSize),
			FillColor:   c,
			BorderColor: borderC,
			BorderWidth: 1,
			Corners:     uimath.CornersAll(3),
		}, 21, 1)
	}
}

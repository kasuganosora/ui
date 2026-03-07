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
	value    uimath.Color
	presets  []uimath.Color
	open     bool
	onChange func(uimath.Color)
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
	tree.AddHandler(cp.id, event.MouseClick, func(e *event.Event) {
		cp.open = !cp.open
	})
	return cp
}

func (cp *ColorPicker) Value() uimath.Color        { return cp.value }
func (cp *ColorPicker) SetValue(c uimath.Color)    { cp.value = c }
func (cp *ColorPicker) SetPresets(p []uimath.Color) { cp.presets = p }
func (cp *ColorPicker) OnChange(fn func(uimath.Color)) { cp.onChange = fn }
func (cp *ColorPicker) IsOpen() bool                { return cp.open }

func (cp *ColorPicker) SelectColor(c uimath.Color) {
	cp.value = c
	cp.open = false
	if cp.onChange != nil {
		cp.onChange(c)
	}
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

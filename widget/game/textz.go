package game

import (
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// Glyph z-orders for game HUD widgets. These components paint opaque fills at
// z>0 (panels, bars, radial menus, tooltips) while a plain TextDrawer emits
// glyphs at z=0, so the label disappears behind its own background on backends
// that submit commands by ascending ZOrder. Each value is one above the highest
// fill in that component family.
const (
	barTextZ       int32 = 3   // buff bar / skill tree / health bar: up to z=2
	panelTextZ     int32 = 4   // shared Panel frame: up to z=3
	lootTextZ      int32 = 5   // loot window rows: up to z=4
	castBarTextZ   int32 = 32  // cast bar: up to z=31
	unitFrameTextZ int32 = 23  // unit frames: up to z=22
	radialTextZ    int32 = 73  // radial menu: up to z=72
	toastTextZ     int32 = 92  // notification toast: up to z=91
	itemTipTextZ   int32 = 101 // item tooltip: up to z=100
)

// emitTextAt draws text at an explicit z-order when the host supports it and
// degrades to a plain DrawText for simpler TextDrawer implementations. This is
// the game subpackage counterpart of widget.emitText (unexported there).
func emitTextAt(cfg *widget.Config, buf *render.CommandBuffer, text string, x, y, size, maxW float32, color uimath.Color, opacity float32, z int32) {
	if cfg == nil || cfg.TextRenderer == nil || text == "" {
		return
	}
	if d, ok := cfg.TextRenderer.(widget.ZTextDrawer); ok {
		d.DrawTextZ(buf, text, x, y, size, maxW, color, opacity, z)
		return
	}
	cfg.TextRenderer.DrawText(buf, text, x, y, size, maxW, color, opacity)
}

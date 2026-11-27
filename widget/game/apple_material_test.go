package game

import (
	"testing"

	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// appleShadowChecker inspects the whole frame (Commands + Overlays) for the
// material contract of a floating surface: a blurred SDF shadow, and no
// offset dark rectangle pretending to be one.
type appleShadowChecker struct {
	shadowBlur float32
	shadows    int
	fakeRects  int
}

func (c *appleShadowChecker) inspect(buf *render.CommandBuffer, x, y float32) {
	all := append([]render.Command{}, buf.Commands()...)
	all = append(all, buf.Overlays()...)
	for _, cmd := range all {
		switch {
		case cmd.Type == render.CmdShadow && cmd.Shadow != nil:
			c.shadows++
			if cmd.Shadow.BlurRadius > c.shadowBlur {
				c.shadowBlur = cmd.Shadow.BlurRadius
			}
		case cmd.Type == render.CmdRect && cmd.Rect != nil:
			r := cmd.Rect.Bounds
			// The old fake shadow was a dark, semi-transparent rectangle offset
			// by exactly +3,+3 from the panel itself.
			if r.X == x+3 && r.Y == y+3 && cmd.Rect.FillColor.A < 1 && cmd.Rect.FillColor.A > 0 {
				c.fakeRects++
			}
		}
	}
}

func (c *appleShadowChecker) requireAppleMaterial(t *testing.T, name string) {
	t.Helper()
	if c.shadows == 0 || c.shadowBlur <= 0 {
		t.Errorf("%s has no blurred SDF shadow (shadows=%d blur=%g); a flat panel reads as a cut-out, not material", name, c.shadows, c.shadowBlur)
	}
	if c.fakeRects > 0 {
		t.Errorf("%s still paints %d offset dark rectangle(s) as a fake shadow", name, c.fakeRects)
	}
}

// TestWindowUsesMaterialShadow keeps game windows on the same depth system as
// the rest of the toolkit: a real SDF shadow from Config.FloatingShadow rather
// than the old x+3,y+3 black rectangle (apple-design §12).
func TestWindowUsesMaterialShadow(t *testing.T) {
	tree := newTree()
	cfg := newCfgWithText()
	win := NewWindow(tree, "统计", cfg)
	setBounds(tree, win, 270, 216, 360, 248)

	buf := render.NewCommandBuffer()
	win.Draw(buf)

	got := &appleShadowChecker{}
	got.inspect(buf, 270, 216)
	got.requireAppleMaterial(t, "game.Window")
}

// TestPanelUsesMaterialShadow covers the second floating-surface path, which
// is drawn through Panel.Draw instead of Widget.Draw.
func TestPanelUsesMaterialShadow(t *testing.T) {
	cfg := newCfgWithText()
	p := &Panel{Title: "背包", Width: 208, Height: 181, Shadow: true}
	buf := render.NewCommandBuffer()
	p.Draw(buf, uimath.NewRect(980, 250, 208, 181), cfg, 0)

	got := &appleShadowChecker{}
	got.inspect(buf, 980, 250)
	got.requireAppleMaterial(t, "game.Panel")
}

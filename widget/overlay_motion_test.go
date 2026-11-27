package widget

import (
	"math"
	"os"
	"strings"
	"testing"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Apple §7 regression: floating surfaces must enter and leave along the same
// anchored path, never through two independent easings.

func TestPanelPlacementAndAnchorDirection(t *testing.T) {
	anchor := uimath.NewRect(300, 200, 80, 30)
	const w, h = 200.0, 120.0
	placements := []PopupPlacement{
		PlacementTop, PlacementBottom, PlacementLeft, PlacementRight,
		PlacementTopLeft, PlacementTopRight, PlacementBottomLeft, PlacementBottomRight,
		PlacementLeftTop, PlacementLeftBottom, PlacementRightTop, PlacementRightBottom,
	}
	for _, pl := range placements {
		x, y := PlacePanel(anchor, w, h, pl, uimath.Rect{})
		panel := uimath.NewRect(x, y, w, h)
		if !panel.IsEmpty() {
			// The panel must not sit on top of its trigger.
			if panel.Intersection(anchor).Width > 0 && panel.Intersection(anchor).Height > 0 {
				t.Errorf("placement %d overlaps the trigger: %+v", pl, panel)
			}
		}
		// The anchor vector must point from the panel back toward the trigger.
		dir := AnchorVector(pl)
		toTrigger := uimath.Vec2{
			X: anchor.X + anchor.Width/2 - (x + w/2),
			Y: anchor.Y + anchor.Height/2 - (y + h/2),
		}
		if dir.X*toTrigger.X+dir.Y*toTrigger.Y <= 0 {
			t.Errorf("placement %d anchor vector %+v points away from trigger (to %+v)", pl, dir, toTrigger)
		}
	}
	// Gap is shared, not duplicated with a different number per widget.
	_, bottom := PlacePanel(anchor, w, h, PlacementBottom, uimath.Rect{})
	if math.Abs(float64(bottom-(anchor.Y+anchor.Height+overlayPanelGap))) > 0.001 {
		t.Errorf("bottom placement gap is not the shared %f: y=%f", overlayPanelGap, bottom)
	}
}

func TestPlacePanelClampsToViewport(t *testing.T) {
	vp := uimath.NewRect(0, 0, 640, 480)
	for _, pl := range []PopupPlacement{PlacementTop, PlacementBottom, PlacementLeft, PlacementRight} {
		anchor := uimath.NewRect(5, 5, 40, 20)
		x, y := PlacePanel(anchor, 200, 120, pl, vp)
		if x < vp.X || y < vp.Y || x+200 > vp.X+vp.Width || y+120 > vp.Y+vp.Height {
			t.Errorf("placement %d escaped the viewport: (%f,%f)", pl, x, y)
		}
	}
}

func frameAtOpenness(m *OverlayMotion, o float32, animate bool) (float32, float32, float32) {
	m.spring.X = o
	m.spring.V = 0
	m.spring.settled = false
	return m.Frame(m.at, animate)
}

func TestOverlayMotionEnterAndExitShareOnePath(t *testing.T) {
	var enter OverlayMotion
	enter.Show(1000, AnchorVector(PlacementBottom))
	var exit OverlayMotion
	exit.Show(1000, AnchorVector(PlacementBottom))
	exit.Close(1000)

	for _, o := range []float32{0, 0.15, 0.35, 0.6, 0.85, 1} {
		edx, edy, ea := frameAtOpenness(&enter, o, true)
		xdx, xdy, xa := frameAtOpenness(&exit, o, true)
		if edx != xdx || edy != xdy || ea != xa {
			t.Errorf("openness %f: enter (%f,%f,%f) != exit (%f,%f,%f)", o, edx, edy, ea, xdx, xdy, xa)
		}
		if math.Abs(float64(ea-o)) > 0.001 {
			t.Errorf("openness %f produced alpha %f", o, ea)
		}
		// Bottom placement emerges from above its trigger (negative Y).
		if o < 1 && edy >= 0 {
			t.Errorf("openness %f: panel did not emerge from the anchored direction: dy=%f", o, edy)
		}
		if o == 1 && (edx != 0 || edy != 0) {
			t.Errorf("fully open panel is still displaced: (%f,%f)", edx, edy)
		}
	}
}

func TestOverlayMotionInterruptKeepsPresentationValue(t *testing.T) {
	var m OverlayMotion
	m.Show(1000, AnchorVector(PlacementTop))
	m.Frame(1100, true)
	live, liveV := m.spring.X, m.spring.V
	if live <= 0 || live >= 1 {
		t.Fatalf("fixture did not put the transition in flight: %f", live)
	}

	m.Close(1100)
	if m.spring.X != live || m.spring.V != liveV {
		t.Errorf("close jumped from the presentation value: %f/%f -> %f/%f",
			live, liveV, m.spring.X, m.spring.V)
	}
	m.Frame(1150, true)
	mid, midV := m.spring.X, m.spring.V
	m.Show(1150, AnchorVector(PlacementTop))
	if m.spring.X != mid || m.spring.V != midV {
		t.Errorf("reopen jumped from the presentation value: %f/%f -> %f/%f",
			mid, midV, m.spring.X, m.spring.V)
	}
	if m.Closing() {
		t.Error("reopened panel is still on the exit path")
	}
}

func TestOverlayMotionReducedMotionDropsSpatialTravel(t *testing.T) {
	var m OverlayMotion
	m.Show(1000, AnchorVector(PlacementBottom))
	dx, dy, alpha := m.Frame(1000, false)
	if dx != 0 || dy != 0 {
		t.Errorf("reduced motion still moves the panel: (%f,%f)", dx, dy)
	}
	if alpha != 1 {
		t.Errorf("reduced motion did not settle to a static visible state: %f", alpha)
	}
	if m.Active() {
		t.Error("reduced motion left a spring running")
	}
}

func TestTransformScopeMovesWholeCommandSubtree(t *testing.T) {
	buf := render.NewCommandBuffer()
	buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(10, 10, 20, 20)}, 0, 1)
	beforeScope := buf.MarkScope()

	glyphs := []render.GlyphInstance{{X: 30, Y: 40, Width: 8, Height: 8}}
	buf.DrawText(render.TextCmd{X: 30, Y: 40, Glyphs: glyphs}, 1, 1)
	buf.DrawOverlay(render.RectCmd{Bounds: uimath.NewRect(50, 50, 10, 10)}, 2, 1)
	buf.DrawOverlayShadow(render.ShadowCmd{Bounds: uimath.NewRect(50, 50, 10, 10)}, 2, 1)
	buf.PushClip(uimath.NewRect(0, 0, 200, 200))
	buf.TransformScope(beforeScope, 5, -3, 0.5)

	if got := buf.Commands()[0].Rect.Bounds; got.X != 10 || got.Y != 10 {
		t.Errorf("command before the scope was transformed: %+v", got)
	}
	if got := buf.Commands()[0].Opacity; got != 1 {
		t.Errorf("command before the scope was faded: %f", got)
	}
	for i := 1; i < len(buf.Commands()); i++ {
		c := &buf.Commands()[i]
		// Clip commands carry no opacity (backends ignore it); only paint
		// commands are expected to fade.
		if c.Type != render.CmdClip && c.Opacity != 0.5 {
			t.Errorf("command %d opacity = %f, want 0.5", i, c.Opacity)
		}
		switch c.Type {
		case render.CmdText:
			if c.Text.X != 35 || c.Text.Y != 37 || c.Text.Glyphs[0].X != 35 || c.Text.Glyphs[0].Y != 37 {
				t.Errorf("text and glyphs did not move together: %+v", c.Text)
			}
			if glyphs[0].X != 30 {
				t.Error("transform mutated the shared glyph slice")
			}
		case render.CmdClip:
			if c.Clip.Bounds.X != 5 || c.Clip.Bounds.Y != -3 {
				t.Errorf("clip did not follow the panel: %+v", c.Clip.Bounds)
			}
		}
	}
	for i := range buf.Overlays() {
		c := &buf.Overlays()[i]
		if c.Opacity != 0.5 {
			t.Errorf("overlay %d opacity = %f, want 0.5", i, c.Opacity)
		}
		switch c.Type {
		case render.CmdRect:
			if c.Rect.Bounds.X != 55 || c.Rect.Bounds.Y != 47 {
				t.Errorf("overlay rect did not move: %+v", c.Rect.Bounds)
			}
		case render.CmdShadow:
			if c.Shadow.Bounds.X != 55 || c.Shadow.Bounds.Y != 47 {
				t.Errorf("overlay shadow did not move: %+v", c.Shadow.Bounds)
			}
		}
	}
}

func TestFloatingWidgetsPlayExitPath(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	cfg.AnimateOverlays = true
	p := NewPopover(tree, cfg)
	setBounds(tree, p, 0, 0, 640, 480)
	p.SetAnchorRect(100, 100, 80, 30)
	p.SetPlacement(PlacementBottom)
	p.Open()

	buf := render.NewCommandBuffer()
	p.Draw(buf)
	if len(buf.Commands())+len(buf.Overlays()) == 0 {
		t.Fatal("open popover painted nothing")
	}
	faded := false
	for i := range buf.Overlays() {
		if buf.Overlays()[i].Opacity < 1 {
			faded = true
		}
	}
	if !faded {
		t.Error("enter transition is not fading the panel in")
	}

	p.Close()
	if p.IsVisible() {
		t.Error("Close must still hide the widget immediately")
	}
	// Sync runs in Draw, so the exit transition is observed on the next frame.
	buf = render.NewCommandBuffer()
	p.Draw(buf)
	if !p.motion.Closing() {
		t.Fatal("Close did not start the exit transition")
	}
	if len(buf.Commands())+len(buf.Overlays()) == 0 {
		t.Error("exit path was never painted; the panel vanished instead of retracing its path")
	}

	// Reduced motion settles the exit immediately; the next frame is empty.
	cfg.ReducedMotion = true
	p.Draw(render.NewCommandBuffer())
	buf = render.NewCommandBuffer()
	p.Draw(buf)
	if len(buf.Commands())+len(buf.Overlays()) != 0 {
		t.Error("exit transition kept painting after it settled")
	}
}

func TestOverlayMotionSourceGuard(t *testing.T) {
	adopters := []string{
		"popover.go", "contextmenu.go", "drawer.go", "dialog.go",
		"notification.go", "message.go", "popconfirm.go", "subwindow.go",
	}
	for _, f := range adopters {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		text := string(src)
		if !strings.Contains(text, "motion OverlayMotion") || !strings.Contains(text, ".DrawScope(buf,") {
			t.Errorf("%s does not adopt the shared anchored motion", f)
		}
		if !strings.Contains(text, ".motion.Closing()") {
			t.Errorf("%s hides immediately and would skip the exit path", f)
		}
	}
	cmd, err := os.ReadFile("../render/command.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(cmd), "TransformScope(") {
		t.Error("command buffer no longer supports scoped subtree transforms")
	}
	cfg, err := os.ReadFile("../app.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(cfg), "a.cfg.AnimateOverlays = true") {
		t.Error("GUI host does not enable anchored transitions")
	}
	_ = core.InvalidElementID
}

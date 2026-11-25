package widget

import (
	"testing"

	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// zProbeDrawer implements ZTextDrawer and emits one glyph covering the label
// box, so a test can compare that text command's z-order with the fills around
// it. Backends paint by ascending ZOrder: a label below an opaque fill is
// invisible even though a text command was recorded.
type zProbeDrawer struct {
	zs []int32
}

func (d *zProbeDrawer) DrawText(buf *render.CommandBuffer, text string, x, y, fontSize, maxWidth float32, color uimath.Color, opacity float32) {
	d.DrawTextZ(buf, text, x, y, fontSize, maxWidth, color, opacity, 0)
}

func (d *zProbeDrawer) DrawTextZ(buf *render.CommandBuffer, text string, x, y, fontSize, maxWidth float32, color uimath.Color, opacity float32, z int32) {
	d.zs = append(d.zs, z)
	buf.DrawText(render.TextCmd{
		X: x, Y: y, FontSize: fontSize,
		Glyphs: []render.GlyphInstance{{X: x, Y: y, Width: 8, Height: fontSize}},
	}, z, opacity)
}

func (d *zProbeDrawer) LineHeight(fontSize float32) float32 { return fontSize * 1.2 }
func (d *zProbeDrawer) MeasureText(text string, fontSize float32) float32 {
	return float32(len([]rune(text))) * fontSize * 0.7
}

func minF(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func maxF(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func overlapRatio(a, b uimath.Rect) float32 {
	ix := minF(a.X+a.Width, b.X+b.Width) - maxF(a.X, b.X)
	iy := minF(a.Y+a.Height, b.Y+b.Height) - maxF(a.Y, b.Y)
	if ix <= 0 || iy <= 0 || a.Width <= 0 || a.Height <= 0 {
		return 0
	}
	return ix * iy / (a.Width * a.Height)
}

// assertLabelsAboveFills checks that no text command is covered by an opaque
// rectangle painted later. Backends submit two independent, z-sorted lists:
// Commands() first, then Overlays(). A fill in the overlay list therefore
// hides a glyph in the command list regardless of their z values, so both the
// list membership and the z-order must be considered.
type labeledCommand struct {
	cmd     render.Command
	overlay bool
}

func paintCommands(buf *render.CommandBuffer) []labeledCommand {
	var out []labeledCommand
	for _, c := range buf.Commands() {
		out = append(out, labeledCommand{cmd: c})
	}
	for _, c := range buf.Overlays() {
		out = append(out, labeledCommand{cmd: c, overlay: true})
	}
	return out
}

func assertLabelsAboveFills(t *testing.T, name string, buf *render.CommandBuffer) {
	t.Helper()
	cmds := paintCommands(buf)

	type fill struct {
		rect    uimath.Rect
		z       int32
		overlay bool
	}
	var fills []fill
	for _, c := range cmds {
		if c.cmd.Type != render.CmdRect || c.cmd.Rect == nil {
			continue
		}
		// A fill at A=0.75 (tooltip / slider bubble) already hides a glyph,
		// so the cutoff must stay below that. Only near-transparent tints are
		// allowed through.
		if c.cmd.Rect.FillColor.A*c.cmd.Opacity < 0.5 {
			continue
		}
		fills = append(fills, fill{rect: c.cmd.Rect.Bounds, z: c.cmd.ZOrder, overlay: c.overlay})
	}

	texts := 0
	for _, c := range cmds {
		if c.cmd.Type != render.CmdText || c.cmd.Text == nil || len(c.cmd.Text.Glyphs) == 0 {
			continue
		}
		texts++
		minX, minY := float32(1e9), float32(1e9)
		maxX, maxY := float32(-1e9), float32(-1e9)
		for _, g := range c.cmd.Text.Glyphs {
			minX, minY = minF(minX, g.X), minF(minY, g.Y)
			maxX, maxY = maxF(maxX, g.X+g.Width), maxF(maxY, g.Y+g.Height)
		}
		box := uimath.NewRect(minX, minY, maxX-minX, maxY-minY)
		for _, f := range fills {
			if overlapRatio(box, f.rect) < 0.5 {
				continue
			}
			if f.overlay == c.overlay {
				if f.z <= c.cmd.ZOrder {
					continue // same list, glyph painted at/after the fill
				}
			} else if !f.overlay {
				continue // fill in Commands, glyph in Overlays: glyph wins
			}
			t.Errorf("%s: text (z=%d overlay=%v) is covered by an opaque fill (z=%d overlay=%v)",
				name, c.cmd.ZOrder, c.overlay, f.z, f.overlay)
		}
	}
	if texts == 0 {
		t.Fatalf("%s: no text commands emitted", name)
	}
}

// TestWidgetLabelsPaintAboveTheirFills guards the five widgets that used to
// emit labels at the default z=0 while their own backgrounds used z>0.
func TestWidgetLabelsPaintAboveTheirFills(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TextRenderer = &zProbeDrawer{}

	cases := []struct {
		name string
		draw func() *render.CommandBuffer
	}{
		{"badge", func() *render.CommandBuffer {
			tree := newTestTree()
			b := NewBadge(tree, cfg)
			b.SetCount(99)
			setBounds(tree, b, 0, 0, 120, 40)
			buf := render.NewCommandBuffer()
			b.Draw(buf)
			return buf
		}},
		{"table", func() *render.CommandBuffer {
			tree := newTestTree()
			tab := NewTable(tree, []TableColumn{
				{Title: "序号", Width: 80},
				{Title: "名称", Width: 220},
			}, cfg)
			tab.SetRows([][]string{{"1", "稳定"}, {"2", "开发中"}})
			tab.SetStripe(true)
			tab.SetBordered(true)
			setBounds(tree, tab, 0, 0, 400, 160)
			buf := render.NewCommandBuffer()
			tab.Draw(buf)
			return buf
		}},
		{"steps", func() *render.CommandBuffer {
			tree := newTestTree()
			s := NewSteps(tree, cfg)
			s.SetOptions([]StepItem{
				{Title: "第一步", Content: "描述一"},
				{Title: "第二步", Content: "描述二"},
				{Title: "第三步", Content: "描述三"},
			})
			s.SetCurrent(1)
			setBounds(tree, s, 0, 0, 600, 120)
			buf := render.NewCommandBuffer()
			s.Draw(buf)
			return buf
		}},
		{"alert", func() *render.CommandBuffer {
			tree := newTestTree()
			a := NewAlert(tree, "这是一条提示", cfg)
			a.SetTitle("标题")
			a.SetCloseBtn(true)
			setBounds(tree, a, 0, 0, 400, 90)
			buf := render.NewCommandBuffer()
			a.Draw(buf)
			return buf
		}},
		{"message", func() *render.CommandBuffer {
			tree := newTestTree()
			m := NewMessage(tree, "普通消息", cfg)
			setBounds(tree, m, 0, 0, 220, 36)
			buf := render.NewCommandBuffer()
			m.Draw(buf)
			return buf
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertLabelsAboveFills(t, tc.name, tc.draw())
		})
	}
}

func TestEmitTextUsesZDrawerWhenAvailable(t *testing.T) {
	probe := &zProbeDrawer{}
	cfg := DefaultConfig()
	cfg.TextRenderer = probe
	buf := render.NewCommandBuffer()

	emitText(cfg, buf, "x", 1, 2, 14, 50, uimath.ColorWhite, 1, 7)

	if len(probe.zs) != 1 || probe.zs[0] != 7 {
		t.Fatalf("z-aware drawer got %v, want [7]", probe.zs)
	}
	if len(buf.Commands()) != 1 || buf.Commands()[0].ZOrder != 7 {
		t.Fatalf("text command z = %v", buf.Commands())
	}
}

func TestEmitTextFallsBackForPlainDrawer(t *testing.T) {
	cfg := cfgWithTextRenderer() // plain TextDrawer, no DrawTextZ
	buf := render.NewCommandBuffer()

	emitText(cfg, buf, "x", 1, 2, 14, 50, uimath.ColorWhite, 1, 7)

	if buf.Len() == 0 {
		t.Fatal("emitText must still draw when the host has no z-aware drawer")
	}
}

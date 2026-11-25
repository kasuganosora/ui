package game

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// zGameProbe emits one real glyph covering the label box so a test can compare
// that text command's z-order with the opaque fills around it.
type zGameProbe struct{}

func (p *zGameProbe) DrawText(buf *render.CommandBuffer, text string, x, y, fontSize, maxWidth float32, color uimath.Color, opacity float32) {
	p.DrawTextZ(buf, text, x, y, fontSize, maxWidth, color, opacity, 0)
}
func (p *zGameProbe) DrawTextZ(buf *render.CommandBuffer, text string, x, y, fontSize, maxWidth float32, color uimath.Color, opacity float32, z int32) {
	buf.DrawText(render.TextCmd{
		X: x, Y: y, FontSize: fontSize, Color: color,
		Glyphs: []render.GlyphInstance{{X: x, Y: y, Width: 8, Height: fontSize}},
	}, z, opacity)
}
func (p *zGameProbe) LineHeight(fontSize float32) float32 { return fontSize * 1.2 }
func (p *zGameProbe) MeasureText(text string, fontSize float32) float32 {
	return float32(len([]rune(text))) * fontSize * 0.6
}

func zMin(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
func zMax(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
func zOverlap(a, b uimath.Rect) float32 {
	ix := zMin(a.X+a.Width, b.X+b.Width) - zMax(a.X, b.X)
	iy := zMin(a.Y+a.Height, b.Y+b.Height) - zMax(a.Y, b.Y)
	if ix <= 0 || iy <= 0 || a.Width <= 0 || a.Height <= 0 {
		return 0
	}
	return ix * iy / (a.Width * a.Height)
}

func assertGameLabelsAboveFills(t *testing.T, name string, buf *render.CommandBuffer) {
	t.Helper()
	// Backends submit Commands() and Overlays() as two separately z-sorted
	// batches, overlays last. Account for both list membership and z-order.
	type draw struct {
		cmd     render.Command
		overlay bool
	}
	var cmds []draw
	for _, c := range buf.Commands() {
		cmds = append(cmds, draw{cmd: c})
	}
	for _, c := range buf.Overlays() {
		cmds = append(cmds, draw{cmd: c, overlay: true})
	}
	type fill struct {
		rect    uimath.Rect
		z       int32
		overlay bool
	}
	var fills []fill
	for _, c := range cmds {
		if c.cmd.Type != render.CmdRect || c.cmd.Rect == nil || c.cmd.Rect.FillColor.A*c.cmd.Opacity < 0.5 {
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
			minX, minY = zMin(minX, g.X), zMin(minY, g.Y)
			maxX, maxY = zMax(maxX, g.X+g.Width), zMax(maxY, g.Y+g.Height)
		}
		box := uimath.NewRect(minX, minY, maxX-minX, maxY-minY)
		for _, f := range fills {
			if zOverlap(box, f.rect) < 0.5 {
				continue
			}
			if f.overlay == c.overlay {
				if f.z <= c.cmd.ZOrder {
					continue
				}
			} else if !f.overlay {
				continue
			}
			t.Errorf("%s: text (z=%d overlay=%v) is covered by an opaque fill (z=%d overlay=%v)",
				name, c.cmd.ZOrder, c.overlay, f.z, f.overlay)
		}
	}
	if texts == 0 {
		t.Fatalf("%s: no text commands emitted (test would pass vacuously)", name)
	}
}

// TestGameWidgetLabelsPaintAboveTheirFills guards every game HUD component
// that paints an opaque panel/bar/menu behind its own labels. FloatingText,
// CountdownTimer, CurrencyDisplay, Nameplate and QuestTracker are included as
// controls: they were verified to be safe and must stay that way.
func TestGameWidgetLabelsPaintAboveTheirFills(t *testing.T) {
	cases := []struct {
		name  string
		build func(*core.Tree, *widget.Config) widget.Widget
	}{
		{"buffbar", func(tree *core.Tree, cfg *widget.Config) widget.Widget {
			w := NewBuffBar(tree, cfg)
			w.AddBuff(Buff{ID: "b1", Label: "增益", Stacks: 2})
			return w
		}},
		{"castbar", func(tree *core.Tree, cfg *widget.Config) widget.Widget {
			w := NewCastBar(tree, cfg)
			w.StartCast("火球术", 2)
			w.Tick(0.5)
			return w
		}},
		{"floatingtext", func(tree *core.Tree, cfg *widget.Config) widget.Widget {
			return NewFloatingText(tree, "+100", 40, 40, uimath.ColorWhite, cfg)
		}},
		{"itemtooltip", func(tree *core.Tree, cfg *widget.Config) widget.Widget {
			w := NewItemTooltip(tree, cfg)
			w.SetItem(&ItemData{ID: "i1", Name: "长剑", Quantity: 1})
			w.SetPosition(40, 40)
			w.SetVisible(true)
			return w
		}},
		{"notificationtoast", func(tree *core.Tree, cfg *widget.Config) widget.Widget {
			w := NewNotificationToast(tree, "获得新成就", cfg)
			w.SetPosition(40, 40)
			w.SetVisible(true)
			return w
		}},
		{"countdown", func(tree *core.Tree, cfg *widget.Config) widget.Widget {
			w := NewCountdownTimer(tree, cfg)
			w.SetSeconds(9)
			w.SetLabel("倒计时")
			return w
		}},
		{"currency", func(tree *core.Tree, cfg *widget.Config) widget.Widget {
			w := NewCurrencyDisplay(tree, cfg)
			w.AddCurrency(CurrencyEntry{Symbol: "G", Amount: 100})
			return w
		}},
		{"dialoguebox", func(tree *core.Tree, cfg *widget.Config) widget.Widget {
			w := NewDialogueBox(tree, cfg)
			w.SetChoices([]DialogueChoice{{Text: "继续"}})
			w.Show("旁白", "这是一段对话文本")
			return w
		}},
		{"healthbar", func(tree *core.Tree, cfg *widget.Config) widget.Widget {
			w := NewHealthBar(tree, cfg)
			w.SetMax(100)
			w.SetCurrent(60)
			w.SetShowText(true)
			return w
		}},
		{"hotbar", func(tree *core.Tree, cfg *widget.Config) widget.Widget {
			w := NewHotbar(tree, 3, cfg)
			w.SetSlot(0, HotbarSlot{Keybind: "1", Label: "技能"})
			return w
		}},
		{"inventory", func(tree *core.Tree, cfg *widget.Config) widget.Widget {
			w := NewInventory(tree, 2, 3, cfg)
			w.SetItem(0, &ItemData{ID: "p", Name: "药水", Quantity: 3})
			return w
		}},
		{"lootwindow", func(tree *core.Tree, cfg *widget.Config) widget.Widget {
			w := NewLootWindow(tree, cfg)
			w.AddItem(LootItem{Item: &ItemData{ID: "l", Name: "战利品"}})
			w.Open()
			return w
		}},
		{"nameplate", func(tree *core.Tree, cfg *widget.Config) widget.Widget {
			w := NewNameplate(tree, "玩家甲", cfg)
			w.SetHP(80, 100)
			w.SetVisible(true)
			return w
		}},
		{"questtracker", func(tree *core.Tree, cfg *widget.Config) widget.Widget {
			w := NewQuestTracker(tree, cfg)
			w.AddQuest(Quest{Title: "主线任务", Active: true, Objectives: []QuestObjective{{Text: "击败 3 只怪物", Current: 1, Required: 3}}})
			return w
		}},
		{"radialmenu", func(tree *core.Tree, cfg *widget.Config) widget.Widget {
			w := NewRadialMenu(tree, cfg)
			w.AddItem(RadialMenuItem{Label: "攻击"})
			w.AddItem(RadialMenuItem{Label: "防御"})
			w.Show(200, 200)
			return w
		}},
		{"scoreboard", func(tree *core.Tree, cfg *widget.Config) widget.Widget {
			w := NewScoreboard(tree, cfg)
			w.AddEntry(ScoreEntry{Name: "玩家甲", Score: 10, Kills: 2, Deaths: 1})
			w.SetVisible(true)
			return w
		}},
		{"skilltree", func(tree *core.Tree, cfg *widget.Config) widget.Widget {
			w := NewSkillTree(tree, cfg)
			w.AddNode(&SkillNode{ID: "s1", Name: "火球", X: 40, Y: 40, Level: 1, MaxLevel: 3})
			return w
		}},
		{"targetframe", func(tree *core.Tree, cfg *widget.Config) widget.Widget {
			w := NewTargetFrame(tree, cfg)
			w.SetTarget(&UnitFrameData{Name: "目标", HP: 80, HPMax: 100})
			return w
		}},
		{"teamframe", func(tree *core.Tree, cfg *widget.Config) widget.Widget {
			w := NewTeamFrame(tree, cfg)
			w.SetMembers([]UnitFrameData{{Name: "队友", HP: 80, HPMax: 100}})
			return w
		}},
		{"gamewindow", func(tree *core.Tree, cfg *widget.Config) widget.Widget {
			w := NewWindow(tree, "背包", cfg)
			w.SetVisible(true)
			return w
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tree := newTree()
			cfg := newCfg()
			cfg.TextRenderer = &zGameProbe{}
			w := tc.build(tree, cfg)
			setBounds(tree, w, 0, 0, 640, 480)
			buf := render.NewCommandBuffer()
			w.Draw(buf)
			assertGameLabelsAboveFills(t, tc.name, buf)
		})
	}
}

func TestEmitTextAtUsesZDrawerWhenAvailable(t *testing.T) {
	cfg := newCfg()
	cfg.TextRenderer = &zGameProbe{}
	buf := render.NewCommandBuffer()

	emitTextAt(cfg, buf, "x", 1, 2, 14, 50, uimath.ColorWhite, 1, 7)

	if len(buf.Commands()) != 1 || buf.Commands()[0].ZOrder != 7 {
		t.Fatalf("text command z = %v, want 7", buf.Commands())
	}
}

func TestEmitTextAtFallsBackForPlainDrawer(t *testing.T) {
	cfg := newCfg()
	cfg.TextRenderer = &mockTextDrawer{} // plain TextDrawer, no DrawTextZ
	buf := render.NewCommandBuffer()

	emitTextAt(cfg, buf, "x", 1, 2, 14, 50, uimath.ColorWhite, 1, 7)

	if buf.Len() == 0 {
		t.Fatal("emitTextAt must still draw when the host has no z-aware drawer")
	}
}

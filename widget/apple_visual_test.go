package widget

import (
	"math"
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/css"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// compositeOver flattens a translucent foreground onto an opaque background,
// the way every backend alpha-blends a text color onto its surface.
func compositeOver(fg, bg uimath.Color) uimath.Color {
	a := fg.A
	return uimath.Color{
		R: fg.R*a + bg.R*(1-a),
		G: fg.G*a + bg.G*(1-a),
		B: fg.B*a + bg.B*(1-a),
		A: 1,
	}
}

func relativeLuminance(c uimath.Color) float64 {
	linear := func(v float32) float64 {
		f := float64(v)
		if f <= 0.03928 {
			return f / 12.92
		}
		return math.Pow((f+0.055)/1.055, 2.4)
	}
	return 0.2126*linear(c.R) + 0.7152*linear(c.G) + 0.0722*linear(c.B)
}

func contrastRatio(fg, bg uimath.Color) float64 {
	l1, l2 := relativeLuminance(compositeOver(fg, bg)), relativeLuminance(bg)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

// TestSecondaryTextTokenIsLegible guards the semantic split between body copy
// and the disabled state. DisabledColor (alpha 0.26) sits near 2:1 on white;
// reusing it for metadata and descriptions made that copy unreadable.
func TestSecondaryTextTokenIsLegible(t *testing.T) {
	cfg := DefaultConfig()
	secondary := contrastRatio(cfg.TextSecondaryColor, cfg.BgColor)
	disabled := contrastRatio(cfg.DisabledColor, cfg.BgColor)
	if secondary < 4.5 {
		t.Errorf("TextSecondaryColor contrast is only %.2f:1, below the 4.5:1 body-text threshold", secondary)
	}
	if secondary <= disabled {
		t.Errorf("secondary text (%.2f:1) is no more legible than DisabledColor (%.2f:1)", secondary, disabled)
	}
}

// TestHairlineRecedesFromBorder keeps section dividers subtler than the border
// they are derived from, instead of a hard 1px rule.
func TestHairlineRecedesFromBorder(t *testing.T) {
	cfg := DefaultConfig()
	h := cfg.Hairline()
	if h.A >= cfg.BorderColor.A {
		t.Fatalf("hairline alpha %g is not softer than border alpha %g", h.A, cfg.BorderColor.A)
	}
	if h.R != cfg.BorderColor.R || h.G != cfg.BorderColor.G || h.B != cfg.BorderColor.B {
		t.Fatalf("hairline must derive from the theme border, got %+v want channels of %+v", h, cfg.BorderColor)
	}
}

// TestDrawOverlayShadowSharesOverlayBatch pins the two-phase submit contract:
// backends paint Commands() and Overlays() as independently z-sorted batches,
// so a shadow behind a DrawOverlay panel must live in Overlays() too.
func TestDrawOverlayShadowSharesOverlayBatch(t *testing.T) {
	buf := render.NewCommandBuffer()
	buf.SetLayerBase(10000)
	buf.DrawOverlayShadow(render.ShadowCmd{
		Bounds:     uimath.NewRect(0, 0, 100, 40),
		BlurRadius: 16,
		Color:      uimath.RGBA(0, 0, 0, 0.12),
	}, 40, 1)
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(0, 0, 100, 40),
		FillColor: uimath.ColorWhite,
	}, 41, 1)

	if len(buf.Commands()) != 0 {
		t.Fatalf("shadow leaked into Commands(): %v", buf.Commands())
	}
	ov := buf.Overlays()
	if len(ov) != 2 || ov[0].Type != render.CmdShadow || ov[1].Type != render.CmdRect {
		t.Fatalf("overlay batch = %v, want [shadow, rect]", ov)
	}
	if ov[0].ZOrder != 10040 || ov[1].ZOrder != 10041 {
		t.Fatalf("layer base not applied uniformly: shadow z=%d panel z=%d", ov[0].ZOrder, ov[1].ZOrder)
	}

	ordered := render.OrderedForPaint(ov)
	shadowAt, panelAt := -1, -1
	for i, c := range ordered {
		switch c.Type {
		case render.CmdShadow:
			shadowAt = i
		case render.CmdRect:
			panelAt = i
		}
	}
	if shadowAt < 0 || panelAt < 0 || shadowAt > panelAt {
		t.Fatalf("painted order puts the panel before its shadow (shadow=%d panel=%d)", shadowAt, panelAt)
	}
}

// TestDivShadowPaintsAboveAncestorBackground reproduces the frame-wide z-sort
// bug that made CSS box-shadow invisible: a negative shadow z sorts in front
// of every z=0 ancestor background, so an opaque layout/main paints over it.
func TestDivShadowPaintsAboveAncestorBackground(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()

	parent := NewDiv(tree, cfg)
	parent.SetBgColor(uimath.ColorWhite)
	setBounds(tree, parent, 0, 0, 400, 300)

	child := NewDiv(tree, cfg)
	child.SetBgColor(uimath.ColorWhite)
	child.SetBoxShadow([]css.BoxShadowLayer{{
		OffsetY: 2,
		Blur:    8,
		Color:   uimath.RGBA(0, 0, 0, 0.08),
	}})
	parent.AppendChild(child)
	setBounds(tree, child, 40, 40, 200, 100)

	buf := render.NewCommandBuffer()
	parent.Draw(buf)
	ordered := render.OrderedForPaint(buf.Commands())

	parentBg, shadow, childBg := -1, -1, -1
	for i, c := range ordered {
		switch c.Type {
		case render.CmdRect:
			if c.Rect == nil {
				continue
			}
			if c.Rect.Bounds.Width == 400 {
				parentBg = i
			} else if c.Rect.Bounds.Width == 200 {
				childBg = i
			}
		case render.CmdShadow:
			shadow = i
		}
	}
	if parentBg < 0 || shadow < 0 || childBg < 0 {
		t.Fatalf("missing commands: parentBg=%d shadow=%d childBg=%d", parentBg, shadow, childBg)
	}
	if !(parentBg < shadow && shadow < childBg) {
		t.Fatalf("paint order parentBg=%d shadow=%d childBg=%d: shadow must sit between the ancestor and its own background", parentBg, shadow, childBg)
	}
}

// TestFloatingSurfacesUseMaterialDepth verifies that floating panels read as
// real material: a blurred SDF shadow in the same submit batch as the panel,
// a theme-driven surface color, and no hard 1px outline. It also catches the
// old hardcoded white panel, which broke dark themes.
func TestFloatingSurfacesUseMaterialDepth(t *testing.T) {
	cases := []struct {
		name  string
		build func(*core.Tree, *Config) Widget
	}{
		{"popover", func(tree *core.Tree, cfg *Config) Widget {
			w := NewPopover(tree, cfg)
			w.title = "气泡标题"
			w.Open()
			return w
		}},
		{"notification", func(tree *core.Tree, cfg *Config) Widget {
			w := NewNotification(tree, "通知标题", "通知内容文本", cfg)
			w.SetCloseBtn(true)
			w.Show()
			return w
		}},
		{"message", func(tree *core.Tree, cfg *Config) Widget {
			return NewMessage(tree, "操作成功", cfg)
		}},
		{"dialog", func(tree *core.Tree, cfg *Config) Widget {
			w := NewDialog(tree, "对话框", cfg)
			w.SetHeader("对话框标题")
			w.SetFooter(true)
			w.Open()
			return w
		}},
		{"drawer", func(tree *core.Tree, cfg *Config) Widget {
			w := NewDrawer(tree, "抽屉", cfg)
			w.SetHeader("抽屉标题")
			w.SetFooter(true)
			w.Open()
			return w
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tree := newTestTree()
			cfg := DefaultConfig()
			cfg.TextRenderer = &zProbeDrawer{}
			surface := uimath.ColorHex("#f7f3ec")
			cfg.BgColor = surface

			w := tc.build(tree, cfg)
			setBounds(tree, w, 0, 0, 640, 480)
			buf := render.NewCommandBuffer()
			w.Draw(buf)

			cmds := paintCommands(buf)
			shadowAt, shadowZ, shadowOverlay, shadowBlur := -1, int32(0), false, float32(0)
			panelAt, panelZ, panelOverlay, panelArea := -1, int32(0), false, float32(0)
			for i, c := range cmds {
				switch {
				case c.cmd.Type == render.CmdShadow && c.cmd.Shadow != nil:
					if c.cmd.Shadow.BlurRadius > shadowBlur {
						shadowAt, shadowZ, shadowOverlay = i, c.cmd.ZOrder, c.overlay
						shadowBlur = c.cmd.Shadow.BlurRadius
					}
				case c.cmd.Type == render.CmdRect && c.cmd.Rect != nil && c.cmd.Rect.FillColor == surface:
					r := c.cmd.Rect.Bounds
					area := r.Width * r.Height
					if area > panelArea {
						panelAt, panelZ, panelOverlay, panelArea = i, c.cmd.ZOrder, c.overlay, area
					}
				}
			}
			if shadowAt < 0 || shadowBlur <= 0 {
				t.Fatalf("no blurred shadow; the panel still uses a flat offset rectangle")
			}
			if panelAt < 0 {
				t.Fatalf("panel is not painted with cfg.BgColor (hardcoded white surface?)")
			}
			if shadowOverlay != panelOverlay {
				t.Fatalf("shadow (overlay=%v) and panel (overlay=%v) land in different submit batches", shadowOverlay, panelOverlay)
			}
			if shadowZ > panelZ {
				t.Fatalf("shadow z=%d paints above panel z=%d", shadowZ, panelZ)
			}
			if cmds[panelAt].cmd.Rect.BorderWidth > 0 {
				t.Fatalf("panel still has a hard %.0fpx outline", cmds[panelAt].cmd.Rect.BorderWidth)
			}
			for _, c := range cmds {
				if c.cmd.Type == render.CmdRect && c.cmd.Rect != nil && c.cmd.Rect.FillColor == uimath.ColorWhite {
					t.Fatalf("hardcoded white rectangle remains in %s (breaks dark themes)", tc.name)
				}
			}
		})
	}
}

// TestModalScrimUsesThemeToken checks that modal tasks dim with the theme's
// mask rather than a fixed light-mode black overlay.
func TestModalScrimUsesThemeToken(t *testing.T) {
	cases := []struct {
		name  string
		build func(*core.Tree, *Config) Widget
	}{
		{"dialog", func(tree *core.Tree, cfg *Config) Widget {
			w := NewDialog(tree, "对话框", cfg)
			w.Open()
			return w
		}},
		{"drawer", func(tree *core.Tree, cfg *Config) Widget {
			w := NewDrawer(tree, "抽屉", cfg)
			w.Open()
			return w
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tree := newTestTree()
			cfg := DefaultConfig()
			cfg.TextRenderer = &zProbeDrawer{}
			mask := uimath.RGBA(0.08, 0.12, 0.18, 0.55)
			cfg.MaskColor = mask

			w := tc.build(tree, cfg)
			setBounds(tree, w, 0, 0, 640, 480)
			buf := render.NewCommandBuffer()
			w.Draw(buf)

			for _, c := range paintCommands(buf) {
				if c.cmd.Type == render.CmdRect && c.cmd.Rect != nil && c.cmd.Rect.FillColor == mask {
					return
				}
			}
			t.Fatalf("scrim does not use cfg.MaskColor")
		})
	}
}

// colorProbeDrawer records the color each label was drawn with. It implements
// ZTextDrawer so both the plain DrawText calls and emitText/emitOverlayText
// paths are covered.
type colorProbeDrawer struct {
	texts  []string
	colors []uimath.Color
}

func (d *colorProbeDrawer) record(buf *render.CommandBuffer, text string, x, y, fontSize float32, color uimath.Color, opacity float32, z int32) {
	d.texts = append(d.texts, text)
	d.colors = append(d.colors, color)
	buf.DrawText(render.TextCmd{
		X: x, Y: y, FontSize: fontSize,
		Glyphs: []render.GlyphInstance{{X: x, Y: y, Width: 8, Height: fontSize}},
	}, z, opacity)
}

func (d *colorProbeDrawer) DrawText(buf *render.CommandBuffer, text string, x, y, fontSize, maxWidth float32, color uimath.Color, opacity float32) {
	d.record(buf, text, x, y, fontSize, color, opacity, 0)
}

func (d *colorProbeDrawer) DrawTextZ(buf *render.CommandBuffer, text string, x, y, fontSize, maxWidth float32, color uimath.Color, opacity float32, z int32) {
	d.record(buf, text, x, y, fontSize, color, opacity, z)
}

func (d *colorProbeDrawer) LineHeight(fontSize float32) float32 { return fontSize * 1.2 }
func (d *colorProbeDrawer) MeasureText(text string, fontSize float32) float32 {
	return float32(len([]rune(text))) * fontSize * 0.7
}

// TestSecondaryBodyCopyUsesSemanticColor keeps metadata, labels and help text
// on the legible secondary token. DisabledColor is still valid for genuinely
// disabled controls, but not for copy the user is expected to read.
func TestSecondaryBodyCopyUsesSemanticColor(t *testing.T) {
	cases := []struct {
		name      string
		secondary string
		build     func(*core.Tree, *Config) Widget
	}{
		{"statistic title", "总额", func(tree *core.Tree, cfg *Config) Widget {
			return NewStatistic(tree, "总额", "1234", cfg)
		}},
		{"comment timestamp", "2 小时前", func(tree *core.Tree, cfg *Config) Widget {
			return NewComment(tree, CommentData{Author: "甲", Content: "内容", Time: "2 小时前"}, cfg)
		}},
		{"timeline content", "已发布", func(tree *core.Tree, cfg *Config) Widget {
			w := NewTimeline(tree, cfg)
			w.AddItem(TimelineItem{Label: "里程碑", Content: "已发布"})
			return w
		}},
		{"descriptions label", "负责人", func(tree *core.Tree, cfg *Config) Widget {
			w := NewDescriptions(tree, cfg)
			w.AddItem(DescriptionItem{Label: "负责人", Value: "甲"})
			return w
		}},
		{"empty description", "暂无数据", func(tree *core.Tree, cfg *Config) Widget {
			w := NewEmpty(tree, cfg)
			w.SetDescription("暂无数据")
			return w
		}},
		{"guide description", "这里是说明文本", func(tree *core.Tree, cfg *Config) Widget {
			w := NewGuide(tree, cfg)
			w.SetSteps([]GuideStep{{Title: "第一步", Description: "这里是说明文本", TargetX: 100, TargetY: 100, TargetW: 80, TargetH: 40}})
			w.visible = true
			return w
		}},
		{"list group caption", "基础", func(tree *core.Tree, cfg *Config) Widget {
			w := NewList(tree, cfg)
			w.SetItems([]ListItem{
				{Value: "button", Group: "基础", Title: "Button 按钮"},
				{Value: "tag", Group: "数据展示", Title: "Tag 标签"},
			})
			return w
		}},
		{"card subtitle", "卡片副标题", func(tree *core.Tree, cfg *Config) Widget {
			w := NewCard(tree, cfg)
			w.SetTitle("卡片标题")
			w.SetSubtitle("卡片副标题")
			return w
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tree := newTestTree()
			cfg := DefaultConfig()
			probe := &colorProbeDrawer{}
			cfg.TextRenderer = probe

			w := tc.build(tree, cfg)
			setBounds(tree, w, 0, 0, 640, 480)
			buf := render.NewCommandBuffer()
			w.Draw(buf)

			idx := -1
			for i, text := range probe.texts {
				if text == tc.secondary {
					idx = i
					break
				}
			}
			if idx < 0 {
				t.Fatalf("secondary label %q was not drawn at all", tc.secondary)
			}
			if probe.colors[idx] != cfg.TextSecondaryColor {
				t.Errorf("%q uses %+v, want cfg.TextSecondaryColor %+v", tc.secondary, probe.colors[idx], cfg.TextSecondaryColor)
			}
			for i, c := range probe.colors {
				if c == cfg.DisabledColor {
					t.Errorf("%q still uses the unreadable DisabledColor", probe.texts[i])
				}
			}
		})
	}
}

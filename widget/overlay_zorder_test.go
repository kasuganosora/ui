package widget

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// TestOverlayWidgetLabelsPaintAboveTheirFills guards the overlay, dropdown
// and modal states that are normally inactive in a frame. They all used to
// emit labels at z=0 while painting their own opaque background at z>0, so
// the labels vanished as soon as the state was activated.
//
// Every case is built into the state a user actually sees (shown tooltip,
// dragging slider, opened dialog, expanded calendar, ...), not just the
// default hidden state. select-open is kept as a control: it lifts its labels
// with MoveToOverlay instead of a constant, and must stay visible too.
func TestOverlayWidgetLabelsPaintAboveTheirFills(t *testing.T) {
	cases := []struct {
		name  string
		build func(*core.Tree, *Config) Widget
	}{
		{"tooltip", func(tree *core.Tree, cfg *Config) Widget {
			w := NewTooltip(tree, "这是一条提示", 0, cfg)
			w.Show()
			return w
		}},
		{"slider-tooltip+marks", func(tree *core.Tree, cfg *Config) Widget {
			w := NewSlider(tree, cfg)
			w.SetShowTooltip(true)
			w.SetShowStep(true)
			w.SetMarks(map[float32]string{0: "低", 50: "中", 100: "高"})
			w.SetValue(50)
			w.dragging = true // no public setter; mirror the real drag state
			return w
		}},
		{"select-open", func(tree *core.Tree, cfg *Config) Widget {
			w := NewSelect(tree, []SelectOption{{Label: "北京", Value: "bj"}, {Label: "上海", Value: "sh"}}, cfg)
			// Draw() gates the dropdown on open+shown rather than popupVisible.
			w.open = true
			w.shown = w.visibleOptions()
			return w
		}},
		{"dialog", func(tree *core.Tree, cfg *Config) Widget {
			w := NewDialog(tree, "对话框", cfg)
			w.SetHeader("对话框标题")
			w.SetFooter(true)
			w.Open()
			return w
		}},
		{"notification", func(tree *core.Tree, cfg *Config) Widget {
			w := NewNotification(tree, "通知标题", "通知内容文本", cfg)
			w.SetCloseBtn(true)
			w.SetFooter("查看详情")
			w.Show()
			return w
		}},
		{"popconfirm", func(tree *core.Tree, cfg *Config) Widget {
			w := NewPopconfirm(tree, "确定删除吗", cfg)
			w.Show()
			return w
		}},
		{"guide", func(tree *core.Tree, cfg *Config) Widget {
			w := NewGuide(tree, cfg)
			w.SetSteps([]GuideStep{{Title: "第一步", Description: "这里是说明文本", TargetX: 100, TargetY: 100, TargetW: 80, TargetH: 40}})
			w.visible = true // SetSteps alone does not start the tour
			return w
		}},
		{"messagebox", func(tree *core.Tree, cfg *Config) Widget {
			w := NewMessageBox(tree, "确认操作", "确定要继续吗？", cfg)
			w.SetShowCancel(true)
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
		{"popover", func(tree *core.Tree, cfg *Config) Widget {
			w := NewPopover(tree, cfg)
			w.title = "气泡标题"
			w.Open()
			return w
		}},
		{"contextmenu", func(tree *core.Tree, cfg *Config) Widget {
			w := NewContextMenu(tree, cfg)
			w.AddItem(ContextMenuItem{Label: "复制"})
			w.AddItem(ContextMenuItem{Label: "粘贴"})
			w.Show(20, 20)
			return w
		}},
		{"subwindow", func(tree *core.Tree, cfg *Config) Widget {
			w := NewSubWindow(tree, "子窗口", cfg)
			w.SetPosition(20, 20)
			w.SetSize(320, 200)
			w.Open()
			return w
		}},
		{"datepicker", func(tree *core.Tree, cfg *Config) Widget {
			w := NewDatePicker(tree, cfg)
			w.SetDate(2026, 11, 24)
			w.SetOpen(true)
			return w
		}},
		{"timepicker", func(tree *core.Tree, cfg *Config) Widget {
			w := NewTimePicker(tree, cfg)
			w.SetTime(10, 30, 0)
			w.SetOpen(true)
			return w
		}},
		{"autocomplete", func(tree *core.Tree, cfg *Config) Widget {
			w := NewAutoComplete(tree, cfg)
			w.SetSuggestions([]string{"apple", "apply", "banana"})
			w.SetText("app")
			w.SetOpen(true)
			return w
		}},
		{"treeselect", func(tree *core.Tree, cfg *Config) Widget {
			w := NewTreeSelect(tree, cfg)
			w.SetRoots([]*TreeNode{{Key: "a", Label: "节点甲", Expanded: true, Children: []*TreeNode{{Key: "a1", Label: "子节点"}}}})
			w.SetOpen(true)
			return w
		}},
		{"cascader", func(tree *core.Tree, cfg *Config) Widget {
			w := NewCascader(tree, cfg)
			w.SetOptions([]*CascaderOption{{Label: "选项甲", Value: "a", Children: []*CascaderOption{{Label: "子项", Value: "a1"}}}})
			w.SetOpen(true)
			return w
		}},
		{"dock-floating", func(tree *core.Tree, cfg *Config) Widget {
			w := NewDockLayout(tree, cfg)
			w.AddPanel(&DockPanel{ID: "p1", Title: "浮动面板", Position: DockFloat, FloatX: 40, FloatY: 40, FloatW: 260, FloatH: 180, Visible: true})
			return w
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tree := newTestTree()
			cfg := DefaultConfig()
			cfg.TextRenderer = &zProbeDrawer{}
			w := tc.build(tree, cfg)
			setBounds(tree, w, 0, 0, 640, 480)
			buf := render.NewCommandBuffer()
			w.Draw(buf)
			assertLabelsAboveFills(t, tc.name, buf)
		})
	}
}

// TestOverlayTextZConstantsBeatTheirFills pins each family constant above the
// highest fill it is meant to clear. A new fill added to a modal must not
// silently outgrow the glyph z-order used by its labels.
func TestOverlayTextZConstantsBeatTheirFills(t *testing.T) {
	cases := []struct {
		name        string
		textZ       int32
		highestFill int32
	}{
		{"slider mark", sliderMarkTextZ, 2},
		{"dropdown", dropdownTextZ, 42},
		{"tooltip", tooltipTextZ, 21},
		{"dialog", dialogTextZ, 13},
		{"drawer", drawerTextZ, 19},
		{"popover", popoverTextZ, 41},
		{"popconfirm", popconfirmTextZ, 53},
		{"notification", notificationTextZ, 54},
		{"dock float", dockFloatTextZ, 52},
		{"context menu", menuTextZ, 62},
		{"guide", guideTextZ, 83},
		{"sub window", subwindowTextZ, 32},
	}
	for _, tc := range cases {
		if tc.textZ <= tc.highestFill {
			t.Errorf("%s: text z=%d is not above fill z=%d", tc.name, tc.textZ, tc.highestFill)
		}
	}
}

// The overlay-list routing must stay explicit: a glyph left in Commands() is
// hidden by a DrawOverlay panel regardless of its z-order.
func TestEmitOverlayTextMovesGlyphsToOverlays(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TextRenderer = &zProbeDrawer{}
	buf := render.NewCommandBuffer()

	emitOverlayText(cfg, buf, "x", 1, 2, 14, 50, uimath.ColorWhite, 1, 42)

	if len(buf.Commands()) != 0 {
		t.Fatalf("glyph left in Commands(): %v", buf.Commands())
	}
	ov := buf.Overlays()
	if len(ov) != 1 || ov[0].Type != render.CmdText || ov[0].ZOrder != 42 {
		t.Fatalf("overlay glyph = %v, want one CmdText at z=42", ov)
	}
}

func TestEmitOverlayTextRespectsLayerBase(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TextRenderer = &zProbeDrawer{}
	buf := render.NewCommandBuffer()
	buf.SetLayerBase(10000)

	emitOverlayText(cfg, buf, "x", 1, 2, 14, 50, uimath.ColorWhite, 1, 42)

	ov := buf.Overlays()
	if len(ov) != 1 || ov[0].ZOrder != 10042 {
		t.Fatalf("overlay glyph z = %v, want 10042", ov)
	}
}

func TestEmitOverlayTextFallsBackForPlainDrawer(t *testing.T) {
	cfg := cfgWithTextRenderer() // plain TextDrawer, no DrawTextZ
	buf := render.NewCommandBuffer()

	emitOverlayText(cfg, buf, "x", 1, 2, 14, 50, uimath.ColorWhite, 1, 42)

	if len(buf.Overlays()) == 0 {
		t.Fatal("emitOverlayText must still draw into overlays for a plain drawer")
	}
}

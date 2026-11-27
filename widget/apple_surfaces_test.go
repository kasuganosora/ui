package widget

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// appleSurfaceCases enumerates the non-floating widgets touched by the
// repo-wide material/contrast pass. Keeping the list explicit makes it
// obvious when a new widget is added without being reviewed.
var appleSurfaceCases = []struct {
	name  string
	build func(*core.Tree, *Config) Widget
}{
	{"list", func(tree *core.Tree, cfg *Config) Widget {
		w := NewList(tree, cfg)
		w.SetItems([]ListItem{{Value: "a", Title: "按钮"}, {Value: "b", Title: "标签"}})
		return w
	}},
	{"card", func(tree *core.Tree, cfg *Config) Widget {
		w := NewCard(tree, cfg)
		w.SetTitle("标题")
		w.SetSubtitle("副标题")
		return w
	}},
	{"breadcrumb", func(tree *core.Tree, cfg *Config) Widget {
		return NewBreadcrumb(tree, cfg)
	}},
	{"divider", func(tree *core.Tree, cfg *Config) Widget {
		return NewDivider(tree, cfg)
	}},
	{"taginput", func(tree *core.Tree, cfg *Config) Widget {
		return NewTagInput(tree, cfg)
	}},
	{"chat", func(tree *core.Tree, cfg *Config) Widget {
		return NewChat(tree, cfg)
	}},
	{"editor", func(tree *core.Tree, cfg *Config) Widget {
		return NewEditor(tree, cfg)
	}},
	{"formula", func(tree *core.Tree, cfg *Config) Widget {
		return NewFormula(tree, cfg)
	}},
	{"richeditor", func(tree *core.Tree, cfg *Config) Widget {
		return NewRichEditor(tree, cfg)
	}},
	{"terminal", func(tree *core.Tree, cfg *Config) Widget {
		return NewTerminal(tree, cfg)
	}},
	{"inputnumber", func(tree *core.Tree, cfg *Config) Widget {
		return NewInputNumber(tree, cfg)
	}},
	{"tabs", func(tree *core.Tree, cfg *Config) Widget {
		return NewTabs(tree, []TabPanel{{Value: "a", Label: "标签一"}, {Value: "b", Label: "标签二"}}, cfg)
	}},
	{"steps", func(tree *core.Tree, cfg *Config) Widget {
		w := NewSteps(tree, cfg)
		w.SetOptions([]StepItem{{Title: "第一步", Content: "说明"}})
		return w
	}},
	{"timeline", func(tree *core.Tree, cfg *Config) Widget {
		w := NewTimeline(tree, cfg)
		w.AddItem(TimelineItem{Label: "里程碑", Content: "已发布"})
		return w
	}},
	{"panel", func(tree *core.Tree, cfg *Config) Widget {
		return NewPanel(tree, "面板", cfg)
	}},
	{"input", func(tree *core.Tree, cfg *Config) Widget {
		return NewInput(tree, cfg)
	}},
	{"textarea", func(tree *core.Tree, cfg *Config) Widget {
		return NewTextArea(tree, cfg)
	}},
	{"upload", func(tree *core.Tree, cfg *Config) Widget {
		return NewUpload(tree, cfg)
	}},
	{"img", func(tree *core.Tree, cfg *Config) Widget {
		return NewImg(tree, cfg)
	}},
	{"space", func(tree *core.Tree, cfg *Config) Widget {
		return NewSpace(tree, cfg)
	}},
	{"empty", func(tree *core.Tree, cfg *Config) Widget {
		w := NewEmpty(tree, cfg)
		w.SetDescription("暂无数据")
		return w
	}},
	{"calendar", func(tree *core.Tree, cfg *Config) Widget {
		return NewCalendar(tree, cfg)
	}},
	{"comment", func(tree *core.Tree, cfg *Config) Widget {
		return NewComment(tree, CommentData{Author: "甲", Content: "内容", Time: "2 小时前"}, cfg)
	}},
	{"descriptions", func(tree *core.Tree, cfg *Config) Widget {
		w := NewDescriptions(tree, cfg)
		w.AddItem(DescriptionItem{Label: "负责人", Value: "甲"})
		return w
	}},
	{"statistic", func(tree *core.Tree, cfg *Config) Widget {
		return NewStatistic(tree, "总额", "1234", cfg)
	}},
	{"notification", func(tree *core.Tree, cfg *Config) Widget {
		return NewNotification(tree, "通知标题", "通知内容", cfg)
	}},
	{"message", func(tree *core.Tree, cfg *Config) Widget {
		return NewMessage(tree, "操作成功", cfg)
	}},
}

// TestAllSurfaceWidgetsRespectTheme verifies that every reviewed widget
// paints its surface from the theme and does not leave hardcoded white
// rectangles (which break dark themes) or unreadable DisabledColor body copy.
func TestAllSurfaceWidgetsRespectTheme(t *testing.T) {
	for _, tc := range appleSurfaceCases {
		t.Run(tc.name, func(t *testing.T) {
			tree := newTestTree()
			cfg := DefaultConfig()
			cfg.TextRenderer = &colorProbeDrawer{}
			cfg.BgColor = uimath.ColorHex("#f7f3ec")

			w := tc.build(tree, cfg)
			setBounds(tree, w, 0, 0, 640, 480)
			buf := render.NewCommandBuffer()
			w.Draw(buf)

			themeRects, whiteRects := 0, 0
			for _, lc := range paintCommands(buf) {
				c := lc.cmd
				if c.Type != render.CmdRect || c.Rect == nil {
					continue
				}
				switch c.Rect.FillColor {
				case cfg.BgColor:
					themeRects++
				case uimath.ColorWhite:
					whiteRects++
				}
			}
			probe := cfg.TextRenderer.(*colorProbeDrawer)
			disabledTexts := 0
			for i, color := range probe.colors {
				if color == cfg.DisabledColor {
					t.Logf("DisabledColor used for %q", probe.texts[i])
					disabledTexts++
				}
			}
			t.Logf("themeRects=%d whiteRects=%d texts=%d disabledTexts=%d",
				themeRects, whiteRects, len(probe.texts), disabledTexts)

			if whiteRects > 0 {
				t.Errorf("%d hardcoded white rectangle(s) remain (breaks dark themes)", whiteRects)
			}
			if disabledTexts > 0 {
				t.Errorf("%d body label(s) still use the unreadable DisabledColor", disabledTexts)
			}
		})
	}
}

package widget

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// appleFloatingCases enumerates every floating surface covered by the
// repo-wide material pass. Each case must actually activate the surface so
// Draw() emits the panel, not just the closed-state placeholder.
var appleFloatingCases = []struct {
	name  string
	build func(*core.Tree, *Config) Widget
}{
	{"contextmenu", func(tree *core.Tree, cfg *Config) Widget {
		w := NewContextMenu(tree, cfg)
		w.AddItem(ContextMenuItem{Label: "复制"})
		w.Show(40, 40)
		return w
	}},
	{"datepicker", func(tree *core.Tree, cfg *Config) Widget {
		w := NewDatePicker(tree, cfg)
		w.SetOpen(true)
		return w
	}},
	{"timepicker", func(tree *core.Tree, cfg *Config) Widget {
		w := NewTimePicker(tree, cfg)
		w.open = true
		return w
	}},
	{"popconfirm", func(tree *core.Tree, cfg *Config) Widget {
		w := NewPopconfirm(tree, "确认删除?", cfg)
		w.SetAnchorRect(40, 40, 80, 30)
		w.Show()
		return w
	}},
	{"subwindow", func(tree *core.Tree, cfg *Config) Widget {
		w := NewSubWindow(tree, "子窗口", cfg)
		w.Open()
		return w
	}},
	{"select dropdown", func(tree *core.Tree, cfg *Config) Widget {
		w := NewSelect(tree, []SelectOption{{Label: "甲", Value: "a"}, {Label: "乙", Value: "b"}}, cfg)
		w.open = true
		w.shown = w.visibleOptions()
		return w
	}},
	{"autocomplete", func(tree *core.Tree, cfg *Config) Widget {
		w := NewAutoComplete(tree, cfg)
		w.SetSuggestions([]string{"甲", "乙"})
		w.SetText("甲")
		return w
	}},
	{"cascader", func(tree *core.Tree, cfg *Config) Widget {
		w := NewCascader(tree, cfg)
		w.SetOptions([]*CascaderOption{{Label: "甲", Value: "a"}})
		w.SetOpen(true)
		return w
	}},
	{"colorpicker", func(tree *core.Tree, cfg *Config) Widget {
		w := NewColorPicker(tree, cfg)
		w.open = true
		return w
	}},
	{"treeselect", func(tree *core.Tree, cfg *Config) Widget {
		w := NewTreeSelect(tree, cfg)
		w.SetOpen(true)
		return w
	}},
	{"backtop", func(tree *core.Tree, cfg *Config) Widget {
		w := NewBackTop(tree, cfg)
		w.SetVisible(true)
		return w
	}},
	{"guide", func(tree *core.Tree, cfg *Config) Widget {
		w := NewGuide(tree, cfg)
		w.SetSteps([]GuideStep{{Title: "第一步", Description: "说明文本", TargetX: 100, TargetY: 100, TargetW: 80, TargetH: 40}})
		w.visible = true
		return w
	}},
	{"messagebox", func(tree *core.Tree, cfg *Config) Widget {
		w := NewMessageBox(tree, "标题", "内容文本", cfg)
		w.Open()
		return w
	}},
	{"command dialog", func(tree *core.Tree, cfg *Config) Widget {
		w := NewCommandDialog(tree, cfg)
		w.Open()
		// The child command panel gets its own layout box in the real tree.
		tree.SetLayout(w.cmd.ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(120, 80, 400, 250),
		})
		return w
	}},
}

// appleSurfaceStats is the measured material state of a widget's frame.
type appleSurfaceStats struct {
	commands      int
	shadows       int
	maxBlur       float32
	shadowOverlay bool
	shadowZ       int32
	surfaceRects  int
	maxArea       float32
	panelOverlay  bool
	panelZ        int32
	panelBorder   float32
	whiteRects    int
	fakeShadows   int
}

func measureAppleSurface(buf *render.CommandBuffer, surface uimath.Color) appleSurfaceStats {
	s := appleSurfaceStats{shadowZ: -1 << 30, panelZ: -1 << 30}
	type rectInfo struct {
		area    float32
		overlay bool
		z       int32
		border  float32
	}
	var surfaces []rectInfo
	for _, lc := range paintCommands(buf) {
		c := lc.cmd
		s.commands++
		switch {
		case c.Type == render.CmdShadow && c.Shadow != nil:
			s.shadows++
			if c.Shadow.BlurRadius > s.maxBlur {
				s.maxBlur = c.Shadow.BlurRadius
				s.shadowOverlay = lc.overlay
				s.shadowZ = c.ZOrder
			}
		case c.Type == render.CmdRect && c.Rect != nil:
			r := c.Rect.Bounds
			if c.Rect.FillColor == uimath.ColorWhite {
				s.whiteRects++
			}
			if c.Rect.FillColor.A > 0 && c.Rect.FillColor.A < 0.5 &&
				r.X > 0 && r.Y > 0 && r.Width > 100 && r.Height > 40 {
				// A large translucent rectangle offset from the panel is the
				// old flat fake-shadow technique.
				s.fakeShadows++
			}
			if c.Rect.FillColor == surface {
				s.surfaceRects++
				surfaces = append(surfaces, rectInfo{
					area:    r.Width * r.Height,
					overlay: lc.overlay,
					z:       c.ZOrder,
					border:  c.Rect.BorderWidth,
				})
			}
		}
	}

	// A dropdown also paints its (opaque, 1px-bordered) input field into
	// Commands(), which is larger than the floating panel. Select the panel
	// by matching the shadow's submit batch first, then fall back to the
	// largest surface. Input borders are legitimate; panel outlines are not.
	var best *rectInfo
	for i := range surfaces {
		cand := &surfaces[i]
		if s.shadows > 0 && cand.overlay != s.shadowOverlay {
			continue
		}
		if best == nil || cand.area > best.area {
			best = cand
		}
	}
	if best == nil {
		for i := range surfaces {
			if best == nil || surfaces[i].area > best.area {
				best = &surfaces[i]
			}
		}
	}
	if best != nil {
		s.maxArea = best.area
		s.panelOverlay = best.overlay
		s.panelZ = best.z
		s.panelBorder = best.border
	}
	return s
}

// TestAllFloatingSurfacesUseThemeMaterial is the repo-wide extension of
// TestFloatingSurfacesUseMaterialDepth: every dropdown, popover, menu, modal
// and subwindow must read as real material (apple-design §12) instead of a
// white cut-out with a hard outline and an offset black rectangle.
func TestAllFloatingSurfacesUseThemeMaterial(t *testing.T) {
	const surfaceHex = "#f7f3ec"
	for _, tc := range appleFloatingCases {
		t.Run(tc.name, func(t *testing.T) {
			tree := newTestTree()
			cfg := DefaultConfig()
			cfg.TextRenderer = &zProbeDrawer{}
			surface := uimath.ColorHex(surfaceHex)
			cfg.BgColor = surface

			w := tc.build(tree, cfg)
			setBounds(tree, w, 0, 0, 640, 480)
			buf := render.NewCommandBuffer()
			w.Draw(buf)

			s := measureAppleSurface(buf, surface)
			t.Logf("commands=%d shadows=%d blur=%g shadowBatch=(overlay=%v z=%d) surfaceRects=%d panel=(overlay=%v z=%d border=%g area=%g) whiteRects=%d fakeShadows=%d",
				s.commands, s.shadows, s.maxBlur, s.shadowOverlay, s.shadowZ,
				s.surfaceRects, s.panelOverlay, s.panelZ, s.panelBorder, s.maxArea,
				s.whiteRects, s.fakeShadows)

			if s.shadows == 0 || s.maxBlur <= 0 {
				t.Errorf("no blurred SDF shadow (shadows=%d blur=%g): panel is a flat cut-out", s.shadows, s.maxBlur)
			}
			if s.surfaceRects == 0 {
				t.Errorf("panel is not painted with cfg.BgColor (hardcoded surface?)")
			}
			if s.fakeShadows > 0 {
				t.Errorf("%d offset dark rectangle(s) still impersonate a shadow", s.fakeShadows)
			}
			if s.whiteRects > 0 {
				t.Errorf("%d hardcoded white rectangle(s) break dark themes", s.whiteRects)
			}
			if s.surfaceRects > 0 && s.shadows > 0 && s.shadowOverlay != s.panelOverlay {
				t.Errorf("shadow (overlay=%v) and panel (overlay=%v) land in different submit batches", s.shadowOverlay, s.panelOverlay)
			}
			if s.panelBorder > 0 {
				t.Errorf("panel still has a hard %.0fpx outline", s.panelBorder)
			}
		})
	}
}

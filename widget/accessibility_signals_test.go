package widget

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// largestOverlaySurface finds the floating panel in a frame. Input fields and
// scrims are filtered out by matching the surface color and requiring the
// overlay submit batch, which is where every floating panel is emitted.
func largestOverlaySurface(buf *render.CommandBuffer, surface uimath.Color) *render.RectCmd {
	var best *render.RectCmd
	bestArea := float32(0)
	for _, lc := range paintCommands(buf) {
		if !lc.overlay || lc.cmd.Type != render.CmdRect || lc.cmd.Rect == nil {
			continue
		}
		if lc.cmd.Rect.FillColor != surface {
			continue
		}
		r := lc.cmd.Rect.Bounds
		if area := r.Width * r.Height; area > bestArea {
			best, bestArea = lc.cmd.Rect, area
		}
	}
	return best
}

// TestAccessibilitySignalsAreIndependent guards the §14 requirement that the
// three OS signals are separate. Enabling one must not silently change the
// materials or borders governed by the others.
func TestAccessibilitySignalsAreIndependent(t *testing.T) {
	base := DefaultConfig()
	base.ApplyAccessibility(AccessibilityPreferences{ReducedMotion: true})
	if base.Animate() {
		t.Error("ReducedMotion did not freeze motion")
	}
	if base.EffectiveMaskColor() != base.MaskColor {
		t.Error("ReducedMotion changed the modal scrim")
	}
	if base.PanelBorderWidth() != 0 || base.PanelBorderColor() != (uimath.Color{}) {
		t.Error("ReducedMotion added a panel outline")
	}
	if base.Hairline() == base.BorderColor {
		t.Error("ReducedMotion unexpectedly solidified the hairline")
	}

	onlyTransparency := DefaultConfig()
	onlyTransparency.ApplyAccessibility(AccessibilityPreferences{ReducedTransparency: true})
	if !onlyTransparency.Animate() {
		t.Error("ReducedTransparency must not freeze motion")
	}
	if onlyTransparency.PanelBorderWidth() != 0 {
		t.Error("ReducedTransparency must not add a panel outline")
	}
	if onlyTransparency.EffectiveMaskColor().A <= onlyTransparency.MaskColor.A {
		t.Error("ReducedTransparency did not solidify the scrim")
	}

	onlyContrast := DefaultConfig()
	onlyContrast.ApplyAccessibility(AccessibilityPreferences{MoreContrast: true})
	if !onlyContrast.Animate() {
		t.Error("MoreContrast must not freeze motion")
	}
	if onlyContrast.PanelBorderWidth() == 0 {
		t.Error("MoreContrast did not add a panel outline")
	}
}

// TestReducedTransparencySolidifiesMaterials covers the three translucent
// layers that §14 asks to make frostier: the modal scrim, the SDF shadow and
// the hairline divider.
func TestReducedTransparencySolidifiesMaterials(t *testing.T) {
	cfg := DefaultConfig()
	normalHairline := cfg.Hairline()
	normalShadow := cfg.ShadowColor

	cfg.ReducedTransparency = true
	if cfg.EffectiveMaskColor().A <= cfg.MaskColor.A || cfg.EffectiveMaskColor().A < 0.75 {
		t.Errorf("scrim alpha only %g, want a near-solid mask", cfg.EffectiveMaskColor().A)
	}
	shadow := cfg.FloatingShadow(uimath.NewRect(0, 0, 100, 40), false)
	if shadow.Color.A <= normalShadow.A {
		t.Errorf("shadow alpha %g did not increase from %g", shadow.Color.A, normalShadow.A)
	}
	if shadow.BlurRadius <= 0 {
		t.Error("reduced transparency must not remove the depth cue entirely")
	}
	if cfg.Hairline() != cfg.BorderColor || normalHairline == cfg.BorderColor {
		t.Error("hairline did not become a solid theme border")
	}
	// The signal must be reversible.
	cfg.ReducedTransparency = false
	if cfg.EffectiveMaskColor() != cfg.MaskColor {
		t.Error("scrim did not return to the theme mask")
	}
}

// TestMoreContrastOutlinesEveryFloatingPanel is the repo-wide behaviour check:
// all 14 floating surfaces must gain a defined contrasting edge, and none may
// keep an outline when the preference is off (the §12 material contract).
func TestMoreContrastOutlinesEveryFloatingPanel(t *testing.T) {
	const surfaceHex = "#f7f3ec"
	for _, tc := range appleFloatingCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, more := range []bool{true, false} {
				cfg := DefaultConfig()
				cfg.TextRenderer = &zProbeDrawer{}
				surface := uimath.ColorHex(surfaceHex)
				cfg.BgColor = surface
				cfg.MoreContrast = more

				tree := newTestTree()
				w := tc.build(tree, cfg)
				setBounds(tree, w, 0, 0, 640, 480)
				buf := render.NewCommandBuffer()
				w.Draw(buf)

				panel := largestOverlaySurface(buf, surface)
				if panel == nil {
					t.Fatalf("more=%v: no floating panel found", more)
				}
				if more {
					if panel.BorderWidth <= 0 {
						t.Errorf("prefers-contrast: panel has no defined outline (width=%g)", panel.BorderWidth)
					}
					if panel.BorderColor != cfg.TextColor {
						t.Errorf("outline %+v is not the contrasting text color %+v", panel.BorderColor, cfg.TextColor)
					}
				} else if panel.BorderWidth > 0 || panel.BorderColor != (uimath.Color{}) {
					t.Errorf("default rendering regained a hard outline (width=%g color=%+v)", panel.BorderWidth, panel.BorderColor)
				}
			}
		})
	}
}

// TestModalScrimsSolidifyUnderReducedTransparency verifies the helper is
// actually wired into every modal, including Guide's previously cached mask.
func TestModalScrimsSolidifyUnderReducedTransparency(t *testing.T) {
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
		{"messagebox", func(tree *core.Tree, cfg *Config) Widget {
			w := NewMessageBox(tree, "标题", "内容", cfg)
			w.Open()
			return w
		}},
		{"guide", func(tree *core.Tree, cfg *Config) Widget {
			w := NewGuide(tree, cfg)
			w.SetSteps([]GuideStep{{Title: "第一步", Description: "说明", TargetX: 100, TargetY: 100, TargetW: 80, TargetH: 40}})
			w.visible = true
			return w
		}},
		{"command dialog", func(tree *core.Tree, cfg *Config) Widget {
			w := NewCommandDialog(tree, cfg)
			w.Open()
			tree.SetLayout(w.cmd.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(120, 80, 400, 250)})
			return w
		}},
	}
	mask := uimath.RGBA(0.1, 0.2, 0.3, 0.3)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.TextRenderer = &zProbeDrawer{}
			cfg.MaskColor = mask
			cfg.ReducedTransparency = true

			tree := newTestTree()
			w := tc.build(tree, cfg)
			setBounds(tree, w, 0, 0, 640, 480)
			buf := render.NewCommandBuffer()
			w.Draw(buf)

			found := false
			for _, lc := range paintCommands(buf) {
				if lc.cmd.Type != render.CmdRect || lc.cmd.Rect == nil {
					continue
				}
				c := lc.cmd.Rect.FillColor
				if c.R == mask.R && c.G == mask.G && c.B == mask.B && c.A >= 0.75 {
					found = true
				}
			}
			if !found {
				t.Error("modal scrim is still translucent; EffectiveMaskColor is not wired in")
			}
		})
	}
}

// TestGuideMaskTracksPreference ensures Guide no longer caches the scrim at
// construction time, which would ignore a preference applied afterwards.
func TestGuideMaskTracksPreference(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TextRenderer = &zProbeDrawer{}
	cfg.MaskColor = uimath.RGBA(0.1, 0.2, 0.3, 0.3)
	tree := newTestTree()
	g := NewGuide(tree, cfg) // preference applied AFTER construction
	g.SetSteps([]GuideStep{{Title: "第一步", Description: "说明", TargetX: 100, TargetY: 100, TargetW: 80, TargetH: 40}})
	g.visible = true
	setBounds(tree, g, 0, 0, 640, 480)
	cfg.ReducedTransparency = true

	buf := render.NewCommandBuffer()
	g.Draw(buf)
	found := false
	for _, lc := range paintCommands(buf) {
		if lc.cmd.Type == render.CmdRect && lc.cmd.Rect != nil {
			c := lc.cmd.Rect.FillColor
			if c.R == cfg.MaskColor.R && c.A >= 0.75 {
				found = true
			}
		}
	}
	if !found {
		t.Error("Guide kept the stale construction-time mask")
	}

	// An explicit override must still win over the theme token.
	g.SetMaskColor(uimath.RGBA(0.9, 0.1, 0.1, 0.5))
	buf = render.NewCommandBuffer()
	g.Draw(buf)
	overrideFound := false
	for _, lc := range paintCommands(buf) {
		if lc.cmd.Type == render.CmdRect && lc.cmd.Rect != nil &&
			lc.cmd.Rect.FillColor == (uimath.RGBA(0.9, 0.1, 0.1, 0.5)) {
			overrideFound = true
		}
	}
	if !overrideFound {
		t.Error("SetMaskColor override was ignored")
	}
}

func TestAccessibilityFromEnv(t *testing.T) {
	t.Setenv(ReducedMotionEnv, "1")
	t.Setenv(ReducedTransparencyEnv, "yes")
	t.Setenv(MoreContrastEnv, "on")
	p := AccessibilityFromEnv()
	if !p.ReducedMotion || !p.ReducedTransparency || !p.MoreContrast {
		t.Fatalf("env preferences not read: %+v", p)
	}
	cfg := DefaultConfig()
	if got := cfg.Accessibility(); got != p {
		t.Fatalf("DefaultConfig applied %+v, want %+v", got, p)
	}

	t.Setenv(ReducedTransparencyEnv, "0")
	t.Setenv(MoreContrastEnv, "0")
	if DefaultConfig().MoreContrast || DefaultConfig().ReducedTransparency {
		t.Error("\"0\" must not enable the preferences")
	}
}

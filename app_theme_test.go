package ui

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/theme"
	"github.com/kasuganosora/ui/widget"
)

// TestSetThemeWiresMaterialTokens guards the token plumbing that the visual
// review exposed: theme tokens existed but never reached widget.Config, so
// floating panels kept hardcoded light-mode colors and fixed scrims.
func TestSetThemeWiresMaterialTokens(t *testing.T) {
	app := &App{cfg: widget.DefaultConfig(), tree: core.NewTree()}

	for _, th := range []*theme.Theme{theme.Light(), theme.Dark()} {
		t.Run(th.Name, func(t *testing.T) {
			app.SetTheme(th)
			cv := th.ToConfig()

			if cv.TextSecondaryColor != th.TextSecondary {
				t.Errorf("ToConfig dropped TextSecondary: got %+v want %+v", cv.TextSecondaryColor, th.TextSecondary)
			}
			if cv.MaskColor != th.Mask {
				t.Errorf("ToConfig dropped Mask: got %+v want %+v", cv.MaskColor, th.Mask)
			}
			if cv.ShadowColor != th.ShadowColor || cv.ShadowY != th.ShadowMdY {
				t.Errorf("ToConfig dropped shadow tokens: got %+v/%g want %+v/%g",
					cv.ShadowColor, cv.ShadowY, th.ShadowColor, th.ShadowMdY)
			}

			if app.cfg.TextSecondaryColor != cv.TextSecondaryColor {
				t.Errorf("SetTheme did not wire TextSecondaryColor: got %+v want %+v", app.cfg.TextSecondaryColor, cv.TextSecondaryColor)
			}
			if app.cfg.MaskColor != cv.MaskColor {
				t.Errorf("SetTheme did not wire MaskColor: got %+v want %+v", app.cfg.MaskColor, cv.MaskColor)
			}
			if app.cfg.ShadowColor != cv.ShadowColor || app.cfg.ShadowY != cv.ShadowY {
				t.Errorf("SetTheme did not wire shadow tokens: got %+v/%g want %+v/%g",
					app.cfg.ShadowColor, app.cfg.ShadowY, cv.ShadowColor, cv.ShadowY)
			}
		})
	}

	if app.cfg.TextSecondaryColor == theme.Light().TextSecondary {
		t.Fatalf("dark theme did not replace the light-theme secondary token")
	}
	if app.cfg.ShadowColor == theme.Light().ShadowColor {
		t.Fatalf("dark theme did not replace the light-theme shadow token")
	}
}

// TestSetReducedMotionReachesWidgets verifies the App-level accessibility
// switch actually changes widget behaviour through the shared Config, rather
// than only setting an unused field (apple-design §14).
func TestSetReducedMotionReachesWidgets(t *testing.T) {
	cfg := widget.DefaultConfig()
	app := &App{cfg: cfg, tree: core.NewTree()}
	loading := widget.NewLoading(app.tree, cfg)
	app.tree.SetLayout(loading.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 200, 80),
	})

	requestsFrame := func() bool {
		app.tree.ClearAllDirty()
		loading.Draw(render.NewCommandBuffer())
		elem := app.tree.Get(loading.ElementID())
		return elem != nil && elem.IsDirty(core.DirtyPaint)
	}

	if !requestsFrame() {
		t.Fatal("control arm: loading indicator should request frames when motion is enabled")
	}
	app.SetReducedMotion(true)
	if !app.ReducedMotion() || cfg.Animate() {
		t.Fatal("SetReducedMotion(true) did not reach the shared Config")
	}
	if requestsFrame() {
		t.Error("widget still requests animation frames after SetReducedMotion(true)")
	}
	app.SetReducedMotion(false)
	if !requestsFrame() {
		t.Error("motion did not resume after SetReducedMotion(false)")
	}
}

// TestSetAccessibilityKeepsSignalsIndependent guards the three-signal design:
// changing motion later must not clear transparency/contrast, and the shared
// Config must reflect every preference on the next frame.
func TestSetAccessibilityKeepsSignalsIndependent(t *testing.T) {
	cfg := widget.DefaultConfig()
	app := &App{cfg: cfg, tree: core.NewTree()}
	baseMask := cfg.MaskColor.A

	app.SetAccessibility(widget.AccessibilityPreferences{
		ReducedTransparency: true,
		MoreContrast:        true,
	})
	app.SetReducedMotion(true)

	got := app.Accessibility()
	if !got.ReducedMotion || !got.ReducedTransparency || !got.MoreContrast {
		t.Fatalf("SetReducedMotion reset the other signals: %+v", got)
	}
	if cfg.EffectiveMaskColor().A <= baseMask {
		t.Error("reduced transparency did not reach the shared Config")
	}
	if cfg.PanelBorderWidth() == 0 || cfg.PanelBorderColor() != cfg.TextColor {
		t.Error("more contrast did not reach the shared Config")
	}
	if cfg.Animate() {
		t.Error("reduced motion did not reach the shared Config")
	}

	app.SetAccessibility(widget.AccessibilityPreferences{})
	if app.Accessibility() != (widget.AccessibilityPreferences{}) {
		t.Fatal("preferences were not reversible")
	}
	if cfg.EffectiveMaskColor() != cfg.MaskColor || cfg.PanelBorderWidth() != 0 {
		t.Error("materials did not return to their default state")
	}
}

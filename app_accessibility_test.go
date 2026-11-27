package ui

import (
	"os"
	"strings"
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/widget"
)

// accessibilityFakePlatform embeds the interface so the test does not have to
// implement every platform method; resolveAccessibility only needs the
// optional capability.
type accessibilityFakePlatform struct {
	platform.Platform
	prefs platform.AccessibilityPreferences
}

func (f accessibilityFakePlatform) AccessibilityPreferences() platform.AccessibilityPreferences {
	return f.prefs
}

var _ platform.AccessibilityProvider = accessibilityFakePlatform{}

// plainFakePlatform implements Platform but deliberately not the optional
// AccessibilityProvider capability.
type plainFakePlatform struct{ platform.Platform }

func clearAccessibilityEnvRoot(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		widget.ReducedMotionEnv, widget.ReducedTransparencyEnv, widget.MoreContrastEnv,
	} {
		old, existed := os.LookupEnv(name)
		if err := os.Unsetenv(name); err != nil {
			t.Fatal(err)
		}
		n, v, ok := name, old, existed
		t.Cleanup(func() {
			if ok {
				_ = os.Setenv(n, v)
			} else {
				_ = os.Unsetenv(n)
			}
		})
	}
}

func TestResolveAccessibilityAdoptsSystemPreferences(t *testing.T) {
	clearAccessibilityEnvRoot(t)
	plat := accessibilityFakePlatform{prefs: platform.AccessibilityPreferences{
		ReduceMotion: true, ReduceTransparency: true, IncreaseContrast: true,
	}}
	got := resolveAccessibility(plat)
	if !got.ReducedMotion || !got.ReducedTransparency || !got.MoreContrast {
		t.Fatalf("OS signals were dropped: %+v", got)
	}
	if got := resolveAccessibility(plainFakePlatform{}); got != (widget.AccessibilityPreferences{}) {
		t.Errorf("platform without the capability should default to off: %+v", got)
	}
}

// An explicit "0" must be able to switch off a system-level true, while the
// two untouched signals keep their OS values.
func TestResolveAccessibilityEnvOverridesSystem(t *testing.T) {
	clearAccessibilityEnvRoot(t)
	plat := accessibilityFakePlatform{prefs: platform.AccessibilityPreferences{
		ReduceMotion: true, ReduceTransparency: true, IncreaseContrast: true,
	}}
	t.Setenv(widget.ReducedMotionEnv, "0")
	t.Setenv(widget.ReducedTransparencyEnv, "false")

	got := resolveAccessibility(plat)
	if got.ReducedMotion || got.ReducedTransparency {
		t.Errorf("explicit env false did not override system true: %+v", got)
	}
	if !got.MoreContrast {
		t.Errorf("overriding motion/transparency reset contrast: %+v", got)
	}

	t.Setenv(widget.MoreContrastEnv, "1")
	if got := resolveAccessibility(accessibilityFakePlatform{}); !got.MoreContrast {
		t.Error("explicit env true did not enable contrast without OS support")
	}
}

func TestResolveAccessibilityKeepsSignalsIndependent(t *testing.T) {
	clearAccessibilityEnvRoot(t)
	plat := accessibilityFakePlatform{prefs: platform.AccessibilityPreferences{
		ReduceTransparency: true,
	}}
	t.Setenv(widget.ReducedMotionEnv, "1")
	got := resolveAccessibility(plat)
	if !got.ReducedMotion || !got.ReducedTransparency || got.MoreContrast {
		t.Errorf("signals bled into each other: %+v", got)
	}
}

// NewApp must actually call the resolver; unit tests alone cannot prove the
// startup wiring. This guards the call site without creating a real window.
func TestNewAppAppliesSystemAccessibility(t *testing.T) {
	src, err := os.ReadFile("app.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(src)
	for _, want := range []string{
		"resolveAccessibility(a.plat)",
		"a.cfg.ApplyAccessibility(",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("NewApp startup wiring lost %q", want)
		}
	}

	// The runtime API remains the highest priority: it is applied after
	// startup and therefore cannot be undone by the resolver.
	app := &App{cfg: widget.DefaultConfig(), tree: core.NewTree()}
	app.cfg.ApplyAccessibility(resolveAccessibility(accessibilityFakePlatform{
		prefs: platform.AccessibilityPreferences{ReduceMotion: true},
	}))
	if !app.ReducedMotion() {
		t.Fatal("system preference did not reach Config")
	}
	app.SetReducedMotion(false)
	if app.ReducedMotion() || app.Accessibility().MoreContrast {
		t.Error("runtime override failed or reset unrelated signals")
	}
}

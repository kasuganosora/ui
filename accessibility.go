package ui

import (
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/widget"
)

// resolveAccessibility merges the OS accessibility signals with the process
// environment. Priority, highest first:
//
//  1. SetAccessibility() at runtime (applied later, wins by construction)
//  2. an explicitly set GOUI_* variable (including "0" to force off)
//  3. the platform's OS preferences (NSWorkspace on macOS)
//  4. defaults (everything off)
//
// The three signals stay independent: overriding motion must never reset the
// contrast or transparency signals (apple-design §14).
func resolveAccessibility(plat platform.Platform) widget.AccessibilityPreferences {
	prefs := widget.AccessibilityPreferences{}
	if provider, ok := plat.(platform.AccessibilityProvider); ok {
		sys := provider.AccessibilityPreferences()
		prefs.ReducedMotion = sys.ReduceMotion
		prefs.ReducedTransparency = sys.ReduceTransparency
		prefs.MoreContrast = sys.IncreaseContrast
	}

	env, mask := widget.AccessibilityEnvOverride()
	if mask.ReducedMotion {
		prefs.ReducedMotion = env.ReducedMotion
	}
	if mask.ReducedTransparency {
		prefs.ReducedTransparency = env.ReducedTransparency
	}
	if mask.MoreContrast {
		prefs.MoreContrast = env.MoreContrast
	}
	return prefs
}

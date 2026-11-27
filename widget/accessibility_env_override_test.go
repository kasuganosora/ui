package widget

import (
	"os"
	"testing"
)

func clearAccessibilityEnv(t *testing.T) {
	t.Helper()
	for _, name := range []string{ReducedMotionEnv, ReducedTransparencyEnv, MoreContrastEnv} {
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

// An unset variable must not override an OS-level true; an explicit false
// must. The original envFlag could not express that difference.
func TestAccessibilityEnvOverrideDistinguishesUnsetFromFalse(t *testing.T) {
	clearAccessibilityEnv(t)

	_, mask := AccessibilityEnvOverride()
	if mask.ReducedMotion || mask.ReducedTransparency || mask.MoreContrast {
		t.Fatalf("unset environment produced an override mask: %+v", mask)
	}
	if AccessibilityFromEnv() != (AccessibilityPreferences{}) {
		t.Error("unset environment should leave all signals off by default")
	}

	t.Setenv(ReducedMotionEnv, "0")
	t.Setenv(MoreContrastEnv, "yes")
	prefs, mask := AccessibilityEnvOverride()
	if !mask.ReducedMotion || !mask.MoreContrast || mask.ReducedTransparency {
		t.Errorf("wrong override mask: %+v", mask)
	}
	if prefs.ReducedMotion || !prefs.MoreContrast || prefs.ReducedTransparency {
		t.Errorf("wrong override values: %+v", prefs)
	}
	fromEnv := AccessibilityFromEnv()
	if fromEnv.ReducedMotion || !fromEnv.MoreContrast {
		t.Errorf("AccessibilityFromEnv changed semantics: %+v", fromEnv)
	}
}

func TestAccessibilityEnvOverrideEmptyValueIsUnset(t *testing.T) {
	clearAccessibilityEnv(t)
	t.Setenv(ReducedTransparencyEnv, "   ")
	prefs, mask := AccessibilityEnvOverride()
	if mask.ReducedTransparency || prefs.ReducedTransparency {
		t.Errorf("blank value should count as unset: %+v / %+v", prefs, mask)
	}
}

func TestAccessibilityEnvOverrideRejectsUnknownValues(t *testing.T) {
	clearAccessibilityEnv(t)
	t.Setenv(ReducedMotionEnv, "maybe")
	prefs, mask := AccessibilityEnvOverride()
	if !mask.ReducedMotion {
		t.Error("an explicit unknown value should still count as an override")
	}
	if prefs.ReducedMotion {
		t.Error("an unknown value must not enable reduced motion")
	}
}

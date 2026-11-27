//go:build darwin

package darwin

import (
	"testing"

	"github.com/kasuganosora/ui/platform"
)

// compile-time assertion: the darwin platform satisfies the optional
// capability the application shell looks for.
var _ platform.AccessibilityProvider = (*Platform)(nil)

// TestSystemAccessibilityPreferencesReadsRealWorkspace is the real-GUI
// counterpart of the environment-variable tests: after a genuine AppKit
// bootstrap, NSWorkspace must answer all three accessibility properties and
// the values must be stable across reads.
func TestSystemAccessibilityPreferencesReadsRealWorkspace(t *testing.T) {
	skipIfNotDarwin(t)
	bootstrapAppKitForWindowTest(t)
	ensureAccessibilitySelectors()

	cls := objcClass("NSWorkspace")
	if cls == 0 {
		t.Fatal("NSWorkspace class is unavailable after AppKit bootstrap")
	}
	ws := msgSend(id(cls), selSharedWorkspaceAcc)
	if ws == 0 {
		t.Fatal("NSWorkspace.sharedWorkspace returned nil")
	}

	for name, sel := range map[string]SEL{
		"accessibilityDisplayShouldReduceMotion":       selShouldReduceMotion,
		"accessibilityDisplayShouldReduceTransparency": selShouldReduceTransp,
		"accessibilityDisplayShouldIncreaseContrast":   selShouldIncreaseContr,
	} {
		if sel == 0 {
			t.Errorf("selector %s was not registered", name)
			continue
		}
		if msgSend(ws, selRespondsToSelector, uintptr(sel)) == 0 {
			// Older macOS releases legitimately lack some of these. The
			// reader must degrade to false instead of raising an exception.
			t.Logf("OS does not implement %s; reader will report false", name)
		}
	}

	p := New()
	got := p.AccessibilityPreferences()
	again := SystemAccessibilityPreferences()
	if got != again {
		t.Errorf("preferences changed between two reads: %+v vs %+v", got, again)
	}
	t.Logf("real system accessibility: reduceMotion=%v reduceTransparency=%v increaseContrast=%v",
		got.ReduceMotion, got.ReduceTransparency, got.IncreaseContrast)
}

// A wrong or unavailable selector must be a safe "false", not an Objective-C
// exception. This is the guard that makes the reader usable on older systems.
func TestWorkspaceBoolDegradesForUnknownSelector(t *testing.T) {
	skipIfNotDarwin(t)
	bootstrapAppKitForWindowTest(t)
	ensureAccessibilitySelectors()

	ws := msgSend(id(objcClass("NSWorkspace")), selSharedWorkspaceAcc)
	if ws == 0 {
		t.Fatal("NSWorkspace.sharedWorkspace returned nil")
	}
	if workspaceBool(ws, objcSelector("gouiDefinitelyNotARealSelector")) {
		t.Error("unknown selector reported true")
	}
	if workspaceBool(0, selShouldReduceMotion) {
		t.Error("nil workspace reported true")
	}
	if workspaceBool(ws, 0) {
		t.Error("nil selector reported true")
	}
	if msgSend(ws, selRespondsToSelector, uintptr(selShouldReduceMotion)) == 0 {
		t.Skip("OS predates NSWorkspace reduce-motion property")
	}
}

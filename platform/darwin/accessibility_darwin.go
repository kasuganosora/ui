//go:build darwin

package darwin

import (
	"sync"

	"github.com/kasuganosora/ui/platform"
)

// NSWorkspace exposes the system accessibility toggles as read-only BOOL
// properties (macOS 10.15+). They are queried through the Objective-C runtime
// so this package stays zero-CGO.
//
// The selectors are resolved once and lazily: sel_registerName is available
// from the package init, but AppKit itself is only loaded by ensureFoundation.
var (
	accessibilityOnce      sync.Once
	selSharedWorkspaceAcc  SEL
	selRespondsToSelector  SEL
	selShouldReduceMotion  SEL
	selShouldReduceTransp  SEL
	selShouldIncreaseContr SEL
)

func ensureAccessibilitySelectors() {
	accessibilityOnce.Do(func() {
		selSharedWorkspaceAcc = objcSelector("sharedWorkspace")
		selRespondsToSelector = objcSelector("respondsToSelector:")
		selShouldReduceMotion = objcSelector("accessibilityDisplayShouldReduceMotion")
		selShouldReduceTransp = objcSelector("accessibilityDisplayShouldReduceTransparency")
		selShouldIncreaseContr = objcSelector("accessibilityDisplayShouldIncreaseContrast")
	})
}

// workspaceBool reads a BOOL property safely. Older macOS releases do not
// implement every selector; sending an unknown selector to NSWorkspace would
// raise an ObjC exception, so respondsToSelector: gates the call. A missing
// class, singleton or selector degrades to false (no accessibility override),
// never to a crash or a garbage value.
func workspaceBool(ws id, sel SEL) bool {
	if ws == 0 || sel == 0 {
		return false
	}
	if msgSend(ws, selRespondsToSelector, uintptr(sel)) == 0 {
		return false
	}
	return msgSend(ws, sel) != 0
}

// SystemAccessibilityPreferences reads the three independent signals from
// NSWorkspace. This is a real OS query, not an environment variable: the
// values reflect System Settings > Accessibility > Display on the host.
func SystemAccessibilityPreferences() platform.AccessibilityPreferences {
	ensureFoundation()
	ensureAccessibilitySelectors()

	cls := objcClass("NSWorkspace")
	if cls == 0 {
		return platform.AccessibilityPreferences{}
	}
	ws := msgSend(id(cls), selSharedWorkspaceAcc)
	if ws == 0 {
		return platform.AccessibilityPreferences{}
	}
	return platform.AccessibilityPreferences{
		ReduceMotion:       workspaceBool(ws, selShouldReduceMotion),
		ReduceTransparency: workspaceBool(ws, selShouldReduceTransp),
		IncreaseContrast:   workspaceBool(ws, selShouldIncreaseContr),
	}
}

// AccessibilityPreferences implements platform.AccessibilityProvider. The
// application shell detects this optional capability and applies the result
// to the shared widget.Config at startup.
func (p *Platform) AccessibilityPreferences() platform.AccessibilityPreferences {
	return SystemAccessibilityPreferences()
}

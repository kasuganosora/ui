package widget

import (
	"os"
	"strings"
)

// Environment signals for hosts that cannot query OS accessibility settings
// themselves. Accepted values are 1/true/yes/on (case-insensitive); anything
// else leaves the preference off.
const (
	ReducedMotionEnv       = "GOUI_REDUCED_MOTION"
	ReducedTransparencyEnv = "GOUI_REDUCED_TRANSPARENCY"
	MoreContrastEnv        = "GOUI_MORE_CONTRAST"
)

// AccessibilityPreferences holds the three independent OS signals described
// in apple-design §14. They are deliberately separate: a user may want less
// motion but still want translucent materials, or need more contrast without
// any change to animation.
type AccessibilityPreferences struct {
	ReducedMotion       bool
	ReducedTransparency bool
	MoreContrast        bool
}

// envFlagExplicit distinguishes "not set" from "explicitly turned off".
// That matters once the OS itself reports a preference: an explicit 0/false
// must be able to override a system-level true, while an unset variable must
// leave the system value alone.
func envFlagExplicit(name string) (value, explicit bool) {
	raw, ok := os.LookupEnv(name)
	if !ok {
		return false, false
	}
	v := strings.ToLower(strings.TrimSpace(raw))
	if v == "" {
		return false, false
	}
	switch v {
	case "1", "true", "yes", "on":
		return true, true
	default: // 0/false/no/off and any unknown value
		return false, true
	}
}

func envFlag(name string) bool {
	v, _ := envFlagExplicit(name)
	return v
}

// AccessibilityMask marks which of the three signals were explicitly set in
// the process environment. Unmasked signals keep the value supplied by the OS.
type AccessibilityMask struct {
	ReducedMotion       bool
	ReducedTransparency bool
	MoreContrast        bool
}

// AccessibilityEnvOverride returns the environment values together with a
// mask telling the host which signals it is allowed to override. It is used
// when merging OS preferences, not for the default Config construction.
func AccessibilityEnvOverride() (AccessibilityPreferences, AccessibilityMask) {
	motion, motionSet := envFlagExplicit(ReducedMotionEnv)
	transparency, transparencySet := envFlagExplicit(ReducedTransparencyEnv)
	contrast, contrastSet := envFlagExplicit(MoreContrastEnv)
	return AccessibilityPreferences{
			ReducedMotion:       motion,
			ReducedTransparency: transparency,
			MoreContrast:        contrast,
		}, AccessibilityMask{
			ReducedMotion:       motionSet,
			ReducedTransparency: transparencySet,
			MoreContrast:        contrastSet,
		}
}

// AccessibilityFromEnv reads all three preferences from the process
// environment. DefaultConfig applies it so an application honours the
// settings without extra wiring; a host can still override Config directly.
func AccessibilityFromEnv() AccessibilityPreferences {
	return AccessibilityPreferences{
		ReducedMotion:       envFlag(ReducedMotionEnv),
		ReducedTransparency: envFlag(ReducedTransparencyEnv),
		MoreContrast:        envFlag(MoreContrastEnv),
	}
}

// ReducedMotionFromEnv is kept as the single-signal convenience used by hosts
// that only care about motion.
func ReducedMotionFromEnv() bool { return envFlag(ReducedMotionEnv) }

// ApplyAccessibility updates the shared Config in place. Widgets hold the
// same pointer, so changes take effect on the next frame without rebuilding
// the widget tree.
func (c *Config) ApplyAccessibility(p AccessibilityPreferences) {
	if c == nil {
		return
	}
	c.ReducedMotion = p.ReducedMotion
	c.ReducedTransparency = p.ReducedTransparency
	c.MoreContrast = p.MoreContrast
}

// Accessibility returns the current preferences.
func (c *Config) Accessibility() AccessibilityPreferences {
	if c == nil {
		return AccessibilityPreferences{}
	}
	return AccessibilityPreferences{
		ReducedMotion:       c.ReducedMotion,
		ReducedTransparency: c.ReducedTransparency,
		MoreContrast:        c.MoreContrast,
	}
}

// Animate is the single gate for decorative looping motion. When reduced
// motion is requested, effects freeze at a legible static state and stop
// asking the tree for continuous frames (apple-design §14: a gentler,
// non-vestibular equivalent, not the absence of feedback).
//
// Functional state changes are deliberately NOT gated by this: auto-dismiss
// timers, progress indicators, focus and hover still behave normally.
func (c *Config) Animate() bool {
	return c != nil && !c.ReducedMotion
}

package ui

import (
	"github.com/kasuganosora/ui/font"
	"github.com/kasuganosora/ui/widget"
)

// ConfigureCaptionFont registers a medium-weight face of the host font and
// wires it into the text drawer used by widgets. Apple §15 builds small-text
// hierarchy from weight + size + leading as a set; at 12px a regular CJK
// outline has almost no fully covered pixels, so a caption reads as grey even
// with a high-contrast theme color.
//
// The capability is optional: native engines that cannot select a face by
// weight, single-font files, and test drawers all fall back to the regular
// face. The caller receives font.InvalidFontID in that case, so a host can
// report the downgrade without treating it as an error.
func ConfigureCaptionFont(mgr font.Manager, path string, drawer widget.TextDrawer) font.ID {
	adapter, ok := drawer.(*textDrawerAdapter)
	if !ok || mgr == nil || path == "" {
		return font.InvalidFontID
	}
	wm, ok := mgr.(font.WeightedManager)
	if !ok {
		return font.InvalidFontID
	}
	id, err := wm.RegisterFileWithWeight("Caption", font.WeightMedium, font.StyleNormal, path)
	if err != nil || id == font.InvalidFontID {
		return font.InvalidFontID
	}
	adapter.SetCaptionFont(id)
	return id
}

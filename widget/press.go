package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Apple §1 — Response: the surface must respond on pointer-down, not on the
// synthesized click at release.
//
// Instead of adding a MouseDown handler and a pressed field to every one of
// the dozens of clickable widgets, HostInput publishes the held-down target
// on core.Tree and each renderer calls DrawPressFeedback once after drawing
// the widget tree. That gives every control an immediate, consistent press
// highlight without changing the click contract: the action itself still
// fires on MouseClick, so dragging off a control still cancels it.
const (
	pressFeedbackZ     int32   = 95
	pressFeedbackAlpha float32 = 0.12
)

// DrawPressFeedback emits the instantaneous highlight for the element
// currently held down. It is a no-op when nothing is pressed, when the target
// became disabled/hidden, or when it has no layout bounds yet.
//
// The tint uses the theme text color so it darkens light surfaces and
// lightens dark ones. It goes into the overlay batch at the target's own
// layer rank, so a button inside a modal is highlighted above that modal's
// panel rather than being swallowed by the two-pass command submission.
func DrawPressFeedback(buf *render.CommandBuffer, tree *core.Tree, cfg *Config) {
	if buf == nil || tree == nil {
		return
	}
	id := tree.Pressed()
	if id == core.InvalidElementID {
		return
	}
	elem := tree.Get(id)
	if elem == nil || !elem.IsVisible() || !elem.IsEnabled() {
		return
	}
	bounds := elem.Layout().Bounds
	if bounds.IsEmpty() {
		return
	}
	if cfg == nil {
		cfg = DefaultConfig()
	}
	tint := cfg.TextColor
	tint.A = pressFeedbackAlpha

	prev := buf.LayerBase()
	buf.SetLayerBase(tree.LayerRank(id) * 10000)
	buf.DrawOverlay(render.RectCmd{
		Bounds:    bounds,
		FillColor: tint,
		Corners:   uimath.CornersAll(cfg.BorderRadius),
	}, pressFeedbackZ, 1)
	buf.SetLayerBase(prev)
}

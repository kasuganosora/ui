package widget

import (
	"time"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Apple §7 — Spatial consistency. "If something disappears one way, we
// expect it to emerge from where it came." A floating surface must enter and
// exit along the same path, and that path must be anchored to the element
// that triggered it rather than to the screen centre.
//
// The mechanism is deliberately shared: OverlayMotion owns one spring whose
// value is "openness" (0 = fully gone, 1 = fully present). Enter springs
// 0→1, exit springs 1→0 on the very same spring with the live velocity
// carried over (§3). Both directions read the same frame function, so the
// outbound path is literally the inbound path reversed — there is no second
// easing curve to drift out of sync.
const (
	overlayPanelGap float32 = 8  // shared anchor-to-panel distance
	overlayTravelPx float32 = 10 // how far the panel starts from its trigger
)

// AnchorVector returns the unit direction from a panel back toward its
// trigger. A top-placed panel sits above its trigger, so it emerges from
// below (0,+1) and returns there; a bottom-placed panel uses (0,-1), and so
// on. Corner placements inherit the edge they primarily grow from.
func AnchorVector(placement PopupPlacement) uimath.Vec2 {
	switch placement {
	case PlacementTop, PlacementTopLeft, PlacementTopRight:
		return uimath.Vec2{X: 0, Y: 1}
	case PlacementLeft, PlacementLeftTop, PlacementLeftBottom:
		return uimath.Vec2{X: 1, Y: 0}
	case PlacementRight, PlacementRightTop, PlacementRightBottom:
		return uimath.Vec2{X: -1, Y: 0}
	default: // bottom variants
		return uimath.Vec2{X: 0, Y: -1}
	}
}

// DrawerAnchorVector is the edge-anchored equivalent for sheets. A right
// drawer emerges from the right edge and dismisses back to the right edge.
func DrawerAnchorVector(placement DrawerPlacement) uimath.Vec2 {
	switch placement {
	case DrawerLeft:
		return uimath.Vec2{X: -1, Y: 0}
	case DrawerTop:
		return uimath.Vec2{X: 0, Y: -1}
	case DrawerBottom:
		return uimath.Vec2{X: 0, Y: 1}
	default: // DrawerRight
		return uimath.Vec2{X: 1, Y: 0}
	}
}

// PlacePanel positions a panel against its trigger. It is the shared
// implementation of the placement switch that used to be duplicated across
// every dropdown, popover and tooltip, so they all agree on the gap and on
// viewport clamping. Passing an empty viewport skips clamping (widgets that
// manage their own bounds keep their existing coordinates).
func PlacePanel(anchor uimath.Rect, panelW, panelH float32, placement PopupPlacement, viewport uimath.Rect) (float32, float32) {
	gap := overlayPanelGap
	var x, y float32
	switch placement {
	case PlacementTop, PlacementTopLeft, PlacementTopRight:
		y = anchor.Y - panelH - gap
		switch placement {
		case PlacementTopLeft:
			x = anchor.X
		case PlacementTopRight:
			x = anchor.X + anchor.Width - panelW
		default:
			x = anchor.X + anchor.Width/2 - panelW/2
		}
	case PlacementLeft, PlacementLeftTop, PlacementLeftBottom:
		x = anchor.X - panelW - gap
		switch placement {
		case PlacementLeftTop:
			y = anchor.Y
		case PlacementLeftBottom:
			y = anchor.Y + anchor.Height - panelH
		default:
			y = anchor.Y + anchor.Height/2 - panelH/2
		}
	case PlacementRight, PlacementRightTop, PlacementRightBottom:
		x = anchor.X + anchor.Width + gap
		switch placement {
		case PlacementRightTop:
			y = anchor.Y
		case PlacementRightBottom:
			y = anchor.Y + anchor.Height - panelH
		default:
			y = anchor.Y + anchor.Height/2 - panelH/2
		}
	default: // bottom variants
		y = anchor.Y + anchor.Height + gap
		switch placement {
		case PlacementBottomLeft:
			x = anchor.X
		case PlacementBottomRight:
			x = anchor.X + anchor.Width - panelW
		default:
			x = anchor.X + anchor.Width/2 - panelW/2
		}
	}
	if !viewport.IsEmpty() {
		if panelW <= viewport.Width {
			x = clampScrollValue(x, viewport.X, viewport.X+viewport.Width-panelW)
		}
		if panelH <= viewport.Height {
			y = clampScrollValue(y, viewport.Y, viewport.Y+viewport.Height-panelH)
		}
	}
	return x, y
}

// OverlayMotion is the open/close spring shared by floating widgets.
type OverlayMotion struct {
	spring     Spring
	active     bool
	closing    bool
	at         int64
	anchor     uimath.Vec2
	wasVisible bool
}

// Show starts (or re-targets) the enter animation. Reopening a panel that is
// half-way through closing continues from its live value and velocity (§3),
// never snapping back to the closed state first.
func (m *OverlayMotion) Show(nowMS int64, anchor uimath.Vec2) {
	from, velocity := float32(0), float32(0)
	if m.active {
		from, velocity = m.spring.X, m.spring.V
	}
	m.spring = NewSpringWithSnap(from, velocity, 1, springCriticalDamping, 0.001, 0.001)
	m.active = true
	m.closing = false
	m.at = nowMS
	m.anchor = anchor
}

// Close starts the exit animation along the same anchored path. It never
// resets the value, so an interrupted enter reverses smoothly.
func (m *OverlayMotion) Close(nowMS int64) {
	from, velocity := float32(1), float32(0)
	if m.active {
		from, velocity = m.spring.X, m.spring.V
	}
	m.spring = NewSpringWithSnap(from, velocity, 0, springCriticalDamping, 0.001, 0.001)
	m.active = true
	m.closing = true
	m.at = nowMS
}

// Sync derives the transition from the widget's visible flag. Keeping this
// in Draw means Open/Show/Close/Hide keep their existing semantics and every
// entry point (including auto-dismiss and Escape) gets the exit path.
func (m *OverlayMotion) Sync(visible bool, anchor uimath.Vec2, nowMS int64) {
	if visible && !m.wasVisible {
		m.Show(nowMS, anchor)
	} else if !visible && m.wasVisible {
		m.Close(nowMS)
	}
	m.wasVisible = visible
}

// Active reports whether a transition still needs frames.
func (m *OverlayMotion) Active() bool { return m.active }

// Closing reports whether the widget is playing its exit path. A widget that
// already flipped its visible flag to false should keep drawing while this is
// true, otherwise the exit would never be seen.
func (m *OverlayMotion) Closing() bool { return m.active && m.closing }

// Openness returns the live 0..1 presentation value.
func (m *OverlayMotion) Openness() float32 {
	if !m.active {
		return 1
	}
	return clampScrollValue(m.spring.X, 0, 1)
}

// Frame advances the spring and returns the transform for this frame:
// (dx, dy, alpha). Reduced motion keeps a static opacity change with no
// spatial travel (§14).
func (m *OverlayMotion) Frame(nowMS int64, animate bool) (float32, float32, float32) {
	if !m.active {
		return 0, 0, 1
	}
	dt := float32(nowMS - m.at)
	m.at = nowMS
	if !animate {
		m.spring.X = m.spring.Target
		m.spring.V = 0
		m.spring.settled = true
	} else if dt > 0 {
		m.spring.Step(dt)
	}
	openness := clampScrollValue(m.spring.X, 0, 1)
	if m.spring.Settled() || !animate {
		m.spring.X = openness
		// Once the transition reaches its target it is over in both
		// directions. Keeping active=true here would make an opened panel
		// repaint forever and a closed one never stop drawing.
		m.active = false
	}
	if !animate {
		return 0, 0, openness
	}
	travel := (1 - openness) * overlayTravelPx
	return m.anchor.X * travel, m.anchor.Y * travel, openness
}

// DrawScope is the one-line integration point for a floating widget:
//
//	defer motion.DrawScope(buf, cfg, tree, id, visible, anchor)()
//
// It snapshots the command stream at the start of Draw and applies the
// anchored transform to every command the widget emitted, including child
// widgets, glyphs, shadows and clips. It also keeps the tree dirty while the
// transition still needs frames.
//
// AnimateOverlays is off for Config values created directly (headless
// renderers and visual regression must stay deterministic); the GUI host
// enables it. ReducedMotion still gates the spatial travel through Animate().
func (m *OverlayMotion) DrawScope(buf *render.CommandBuffer, cfg *Config, tree *core.Tree, id core.ElementID, visible bool, anchor uimath.Vec2) func() {
	if buf == nil || cfg == nil || !cfg.AnimateOverlays {
		return func() {}
	}
	nowMS := time.Now().UnixMilli()
	m.Sync(visible, anchor, nowMS)
	scope := buf.MarkScope()
	return func() {
		dx, dy, alpha := m.Frame(nowMS, cfg.Animate())
		buf.TransformScope(scope, dx, dy, alpha)
		if m.Active() && tree != nil {
			tree.MarkDirty(id)
		}
	}
}

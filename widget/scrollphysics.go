package widget

import (
	"math"
	"time"
)

// Sections 9, 3, 4, 5 and 6 of the Apple design guidance meet in this file.
//
// §9 asks for soft boundaries: at an edge the surface resists progressively
// instead of stopping hard. §3 asks for every animation to be interruptible
// from its live value, §4 for springs instead of scripted durations, §5 for
// the release velocity to be handed to the spring, and §6 for the resting
// position to be projected from that velocity.
//
// This is the single source of truth for that behaviour. Every continuously
// scrollable surface keeps a ScrollEdge and routes user-driven gestures (wheel
// ticks, scrollbar thumb drags) through Apply; Draw calls Settle so the
// surface springs back once the gesture stops, or coasts when it was a flick.
// Programmatic scrolling (ScrollTo/ScrollBy) is deliberately NOT animated:
// callers such as "scroll to bottom" must land on an exact offset. Call Reset
// there.
const rubberGestureGapMS int64 = 140 // consecutive wheel ticks belong to one gesture

// rubberband maps an undamped overshoot to the visible displacement. Small
// overshoots still follow the gesture; larger ones asymptotically approach
// the viewport-sized limit, so the further past the edge the user drags, the
// less the content follows.
func rubberband(overshoot, dimension float32) float32 {
	const c = float32(0.55) // from the design guidance
	if overshoot == 0 || dimension <= 0 {
		return 0
	}
	abs := float32(math.Abs(float64(overshoot)))
	return (overshoot * dimension * c) / (dimension + c*abs)
}

func clampScrollValue(v, min, max float32) float32 {
	if max < min {
		max = min
	}
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// gestureSample is one position/time observation. Apple §2 asks for a short
// history rather than a single pointer position, so release velocity can be
// estimated without being dominated by one noisy tick.
type gestureSample struct {
	at  int64
	raw float32
}

// ScrollEdge tracks the undamped scroll target across one gesture. raw is the
// absolute position the gesture is asking for; it may sit outside [min,max].
// The rendered position is always the damped projection of raw. The spring,
// velocity history and carried velocity live here so re-targeting never has
// to reconstruct state from a logical target value.
type ScrollEdge struct {
	raw       float32
	hasRaw    bool
	lastInput int64

	spring   Spring
	springAt int64
	released bool

	samples    [4]gestureSample // short velocity history (§2)
	sampleN    int
	carriedVel float32 // velocity blended through a re-target (§3)
}

// Target returns the undamped target of the active gesture, or pos when no
// gesture is active. Used by absolute gestures such as scrollbar thumb drags.
func (e *ScrollEdge) Target(pos float32) float32 {
	if !e.hasRaw {
		return pos
	}
	return e.raw
}

// Overscroll reports the undamped distance past the legal bounds. A value of
// zero means the surface is fully settled.
func (e *ScrollEdge) Overscroll(min, max float32) float32 {
	if !e.hasRaw {
		return 0
	}
	return e.raw - clampScrollValue(e.raw, min, max)
}

// Reset abandons any overscroll, spring and velocity state. Programmatic
// scroll entry points must call it so an in-flight gesture cannot fight an
// exact requested offset.
func (e *ScrollEdge) Reset() {
	*e = ScrollEdge{}
}

func (e *ScrollEdge) pushSample(at int64, raw float32) {
	if e.sampleN < len(e.samples) {
		e.samples[e.sampleN] = gestureSample{at: at, raw: raw}
		e.sampleN++
		return
	}
	copy(e.samples[:], e.samples[1:])
	e.samples[len(e.samples)-1] = gestureSample{at: at, raw: raw}
}

// releaseVelocity estimates px/ms from the oldest to newest sample. A run of
// ticks inside the same millisecond has no measurable time, so it contributes
// no velocity rather than an infinite one.
func (e *ScrollEdge) releaseVelocity() float32 {
	if e.sampleN < 2 {
		return 0
	}
	first := e.samples[0]
	last := e.samples[e.sampleN-1]
	if last.at <= first.at {
		return 0
	}
	return (last.raw - first.raw) / float32(last.at-first.at)
}

// Apply consumes a user-driven scroll delta and returns the new visible
// position. min/max are the legal bounds and dimension is the viewport size
// along the scroll axis. When animate is false (reduced motion) the surface
// keeps the hard boundary instead of playing a decorative spring (§14).
//
// §3: an in-flight spring is not discarded. Its velocity is blended into the
// new gesture so reversing direction does not hit a brick wall.
func (e *ScrollEdge) Apply(pos, delta, min, max, dimension float32, nowMS int64, animate bool) float32 {
	if !animate {
		e.Reset()
		return clampScrollValue(pos+delta, min, max)
	}
	if e.spring.Active() {
		e.carriedVel = e.spring.V * momentumReversalBlend
		e.spring = Spring{}
	}
	if !e.hasRaw {
		e.raw = pos
		e.hasRaw = true
		e.samples = [4]gestureSample{}
		e.sampleN = 0
		e.carriedVel = 0
	}
	e.raw += delta
	e.lastInput = nowMS
	e.released = false
	e.pushSample(nowMS, e.raw)
	return e.position(min, max, dimension)
}

// Settle advances the release spring and returns (position, changed). It is
// time-driven, never a fixed-duration scripted animation: a new wheel tick
// interrupts the settle and continues from the current value (§3).
//
// Two release behaviours are decided once, from the live value:
//   - past a boundary: the rubber band already absorbed the gesture, so the
//     spring critically damps back to the boundary (§9);
//   - inside the bounds with a flick: project the resting position from the
//     release velocity (§6), clamp it to the legal range as the nearest snap
//     point, and hand the velocity to a slightly bouncy spring (§4/§5).
func (e *ScrollEdge) Settle(pos, min, max, dimension float32, nowMS int64) (float32, bool) {
	if !e.hasRaw || e.released {
		return pos, false
	}
	// While ticks keep arriving the gesture is still active; do not pull the
	// content back underneath the user's finger.
	if nowMS-e.lastInput < rubberGestureGapMS {
		return pos, false
	}
	base := clampScrollValue(e.raw, min, max)

	if !e.spring.Active() {
		velocity := e.releaseVelocity() + e.carriedVel
		e.carriedVel = 0
		target := base
		initialV := float32(0)
		damping := springCriticalDamping

		if e.raw == base && float32(math.Abs(float64(velocity))) >= momentumMinVelocity {
			projected := clampScrollValue(e.raw+projectMomentum(velocity), min, max)
			if projected != e.raw {
				target = projected
				initialV = velocity
				damping = springMomentumDamping
			}
		} else if e.raw != base && velocity*(e.raw-base) < 0 {
			// The gesture reversed while stretched out of bounds: hand the
			// inward velocity to the spring instead of dropping it (§5).
			initialV = velocity
		}
		// An outward velocity past the edge is intentionally NOT handed off:
		// the rubber band absorbed it, and handing it over would first push
		// the surface further away from the boundary.

		if target == e.raw && initialV == 0 {
			e.raw = base
			e.released = true
			return pos, false
		}
		e.spring = NewSpring(e.raw, initialV, target, damping)
		e.springAt = nowMS
	}

	dt := float32(nowMS - e.springAt)
	e.springAt = nowMS
	if dt > 0 {
		e.spring.Step(dt)
	}
	e.raw = e.spring.X
	if e.spring.Settled() {
		e.raw = e.spring.Target
		e.spring = Spring{}
		e.released = true
	}
	next := e.position(min, max, dimension)
	return next, math.Abs(float64(next-pos)) > 0.01
}

func (e *ScrollEdge) position(min, max, dimension float32) float32 {
	base := clampScrollValue(e.raw, min, max)
	return base + rubberband(e.raw-base, dimension)
}

// applyScrollGesture is the widget-facing wrapper: widgets supply their
// current position, bounds and animation preference, and get back the visible
// position for this input event.
func applyScrollGesture(e *ScrollEdge, pos, delta, min, max, dimension float32, animate bool) float32 {
	return applyScrollGestureAt(e, pos, delta, min, max, dimension, time.Now().UnixMilli(), animate)
}

func applyScrollGestureAt(e *ScrollEdge, pos, delta, min, max, dimension float32, nowMS int64, animate bool) float32 {
	return e.Apply(pos, delta, min, max, dimension, nowMS, animate)
}

// settleScrollGesture is called once per Draw so the release spring keeps
// repainting until the surface reaches its resting position.
func settleScrollGesture(e *ScrollEdge, pos, min, max, dimension float32) (float32, bool) {
	return settleScrollGestureAt(e, pos, min, max, dimension, time.Now().UnixMilli())
}

// settleScrollGestureAt is the deterministic form used by tests: frame times
// are supplied explicitly instead of sleeping on a wall clock.
func settleScrollGestureAt(e *ScrollEdge, pos, min, max, dimension float32, nowMS int64) (float32, bool) {
	return e.Settle(pos, min, max, dimension, nowMS)
}

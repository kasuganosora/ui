package widget

import "math"

// Apple §4 — Behavior over animation. A gesture-driven surface must not run
// a pre-scripted, fixed-duration animation: that cannot be grabbed or
// redirected. Springs have no duration; their settle time emerges from the
// damping ratio and the response, and new input just changes the target.
//
// §5 asks for the gesture's release velocity to be handed to the spring, and
// §6 asks for the resting position to be projected from that velocity. Both
// are implemented on top of this one stateful spring so the value always
// continues from its live on-screen presentation value (§3).
const (
	springResponseMS      float32 = 350   // Apple's 0.3–0.4s house range, in ms
	springCriticalDamping float32 = 1.0   // default: no overshoot
	springMomentumDamping float32 = 0.8   // only when a flick carried momentum
	springMaxStepMS       float32 = 16    // integration substep, ~one frame
	springSnapDistance    float32 = 0.5   // finish the settle below this distance
	springSnapVelocity    float32 = 0.01  // px/ms
	momentumDeceleration  float64 = 0.998 // Apple's sample-code scroll decay
	momentumMinVelocity   float32 = 0.15  // px/ms, below this a release is not a flick
	momentumReversalBlend float32 = 0.5   // §3: carry, don't hard-cut, spring velocity
)

// Spring is a damped harmonic oscillator integrated with semi-implicit Euler.
// X is the live value and V its velocity in units of X per millisecond. A new
// gesture simply re-targets the spring while keeping X and V, which is why an
// interrupted animation never jumps back to a logical target.
type Spring struct {
	X       float32
	V       float32
	Target  float32
	Damping float32
	active  bool
	settled bool
	snapD   float32
	snapV   float32
}

// NewSpring starts a spring at the presentation value. velocity is handed
// straight over from the gesture (§5); damping below 1.0 is reserved for
// momentum-driven interactions (§4). The default snap thresholds are in the
// pixel units used by scroll surfaces.
func NewSpring(x, velocity, target, damping float32) Spring {
	return NewSpringWithSnap(x, velocity, target, damping, springSnapDistance, springSnapVelocity)
}

// NewSpringWithSnap is for springs whose value is not measured in pixels —
// for example the 0..1 openness of a floating panel. A 0.5px snap distance
// would settle that spring on the very first frame.
func NewSpringWithSnap(x, velocity, target, damping, snapDistance, snapVelocity float32) Spring {
	if damping <= 0 {
		damping = springCriticalDamping
	}
	return Spring{
		X: x, V: velocity, Target: target, Damping: damping,
		active: true, snapD: snapDistance, snapV: snapVelocity,
	}
}

// Active reports whether the spring still has work to do.
func (s *Spring) Active() bool { return s.active && !s.settled }

// Settled reports whether the spring reached its target.
func (s *Spring) Settled() bool { return s.settled }

// Step advances the spring by dtMS. Large frame gaps are split into short
// substeps so a stalled frame cannot make the integration explode.
func (s *Spring) Step(dtMS float32) {
	if !s.active || s.settled || dtMS <= 0 {
		return
	}
	omega := float32(2*math.Pi) / springResponseMS // rad/ms
	for remaining := dtMS; remaining > 0; {
		step := remaining
		if step > springMaxStepMS {
			step = springMaxStepMS
		}
		accel := -omega*omega*(s.X-s.Target) - 2*s.Damping*omega*s.V
		s.V += accel * step
		s.X += s.V * step
		remaining -= step
	}
	snapD, snapV := s.snapD, s.snapV
	if snapD <= 0 {
		snapD = springSnapDistance
	}
	if snapV <= 0 {
		snapV = springSnapVelocity
	}
	if float32(math.Abs(float64(s.X-s.Target))) < snapD &&
		float32(math.Abs(float64(s.V))) < snapV {
		s.X = s.Target
		s.V = 0
		s.settled = true
	}
}

// projectMomentum is Apple's exponential-decay projection from Designing
// Fluid Interfaces, not the physics-textbook v²/(2·decel):
//
//	project(v_px_per_s, d) = (v / 1000) * d / (1 - d)
//
// The argument is in px/ms (our internal unit), so the /1000 conversion is
// applied here before the formula.
func projectMomentum(velocityPxPerMS float32) float32 {
	d := momentumDeceleration
	return float32(float64(velocityPxPerMS) * d / (1 - d))
}

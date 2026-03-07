// Package anim provides an animation system with transitions and keyframes.
//
// Animations interpolate values over time using easing functions. A Scheduler
// manages active animations and advances them each frame via Tick().
package anim

import "math"

// Easing is a timing function that maps t in [0,1] to an output value.
type Easing func(t float32) float32

// Standard easing functions.
var (
	Linear      Easing = func(t float32) float32 { return t }
	EaseIn      Easing = func(t float32) float32 { return t * t }
	EaseOut     Easing = func(t float32) float32 { return t * (2 - t) }
	EaseInOut   Easing = easeInOut
	EaseInCubic Easing = func(t float32) float32 { return t * t * t }
	EaseOutCubic Easing = func(t float32) float32 {
		t--
		return t*t*t + 1
	}
	EaseInOutCubic Easing = func(t float32) float32 {
		if t < 0.5 {
			return 4 * t * t * t
		}
		t = 2*t - 2
		return 0.5*t*t*t + 1
	}
	EaseOutBack Easing = func(t float32) float32 {
		const c1 = 1.70158
		const c3 = c1 + 1
		t--
		return 1 + c3*t*t*t + c1*t*t
	}
	EaseOutElastic Easing = func(t float32) float32 {
		if t <= 0 {
			return 0
		}
		if t >= 1 {
			return 1
		}
		return float32(math.Pow(2, -10*float64(t))*math.Sin((float64(t)*10-0.75)*(2*math.Pi/3))) + 1
	}
	EaseOutBounce Easing = easeOutBounce
)

func easeInOut(t float32) float32 {
	if t < 0.5 {
		return 2 * t * t
	}
	return -1 + (4-2*t)*t
}

func easeOutBounce(t float32) float32 {
	const n1 = 7.5625
	const d1 = 2.75
	switch {
	case t < 1/d1:
		return n1 * t * t
	case t < 2/d1:
		t -= 1.5 / d1
		return n1*t*t + 0.75
	case t < 2.5/d1:
		t -= 2.25 / d1
		return n1*t*t + 0.9375
	default:
		t -= 2.625 / d1
		return n1*t*t + 0.984375
	}
}

// State tracks the lifecycle of an animation.
type State uint8

const (
	StatePending  State = iota // Not yet started (has delay)
	StateRunning               // Actively interpolating
	StatePaused                // Paused mid-animation
	StateFinished              // Completed (may loop)
)

// FillMode controls what happens when animation is not running.
type FillMode uint8

const (
	FillNone     FillMode = iota // Revert to original value
	FillForwards                 // Keep final value
	FillBackwards                // Apply initial value during delay
	FillBoth                     // Forwards + Backwards
)

// Direction controls playback direction.
type Direction uint8

const (
	DirNormal    Direction = iota
	DirReverse
	DirAlternate           // Normal then reverse
	DirAlternateReverse    // Reverse then normal
)

// Animation is a single property animation from start to end values.
type Animation struct {
	ID        uint64
	From      float32
	To        float32
	Duration  float32 // seconds
	Delay     float32 // seconds before start
	Easing    Easing
	Fill      FillMode
	Direction Direction
	Repeat    int     // 0 = play once, -1 = infinite, N = repeat N times

	elapsed   float32
	state     State
	iteration int
	paused    bool
	onUpdate  func(value float32)
	onFinish  func()
}

// Value returns the current interpolated value.
func (a *Animation) Value() float32 {
	if a.Duration <= 0 {
		return a.To
	}
	t := a.progress()
	eased := t
	if a.Easing != nil {
		eased = a.Easing(t)
	}
	from, to := a.From, a.To
	if a.isReversed() {
		from, to = to, from
	}
	return from + (to-from)*eased
}

// State returns the animation state.
func (a *Animation) State() State { return a.state }

// IsFinished returns true if the animation has completed.
func (a *Animation) IsFinished() bool { return a.state == StateFinished }

// OnUpdate sets a callback invoked each tick with the current value.
func (a *Animation) OnUpdate(fn func(float32)) { a.onUpdate = fn }

// OnFinish sets a callback invoked when the animation completes.
func (a *Animation) OnFinish(fn func()) { a.onFinish = fn }

// Pause pauses the animation.
func (a *Animation) Pause() {
	if a.state == StateRunning {
		a.state = StatePaused
		a.paused = true
	}
}

// Resume resumes a paused animation.
func (a *Animation) Resume() {
	if a.paused {
		a.state = StateRunning
		a.paused = false
	}
}

func (a *Animation) progress() float32 {
	active := a.elapsed - a.Delay
	if active < 0 {
		if a.Fill == FillBackwards || a.Fill == FillBoth {
			return 0
		}
		return 0
	}
	if a.Duration <= 0 {
		return 1
	}
	t := active / a.Duration
	if t > 1 {
		t = 1
	}
	return t
}

func (a *Animation) isReversed() bool {
	switch a.Direction {
	case DirReverse:
		return true
	case DirAlternate:
		return a.iteration%2 == 1
	case DirAlternateReverse:
		return a.iteration%2 == 0
	}
	return false
}

func (a *Animation) tick(dt float32) {
	if a.state == StateFinished || a.state == StatePaused {
		return
	}
	a.elapsed += dt

	if a.elapsed < a.Delay {
		a.state = StatePending
		return
	}
	a.state = StateRunning

	active := a.elapsed - a.Delay
	if active >= a.Duration && a.Duration > 0 {
		// Iteration complete
		if a.Repeat == -1 || a.iteration < a.Repeat {
			a.iteration++
			a.elapsed = a.Delay // reset for next iteration
			if a.onUpdate != nil {
				a.onUpdate(a.Value())
			}
			return
		}
		a.state = StateFinished
		if a.onUpdate != nil {
			a.onUpdate(a.Value())
		}
		if a.onFinish != nil {
			a.onFinish()
		}
		return
	}

	if a.onUpdate != nil {
		a.onUpdate(a.Value())
	}
}

// Keyframe defines a point in a keyframe animation.
type Keyframe struct {
	Offset float32 // 0.0 to 1.0 (percentage of duration)
	Value  float32
	Easing Easing  // Easing to next keyframe (nil = linear)
}

// KeyframeAnimation interpolates through multiple keyframes.
type KeyframeAnimation struct {
	ID        uint64
	Keyframes []Keyframe
	Duration  float32
	Delay     float32
	Fill      FillMode
	Direction Direction
	Repeat    int

	elapsed   float32
	state     State
	iteration int
	paused    bool
	onUpdate  func(value float32)
	onFinish  func()
}

// Value returns the current interpolated value.
func (ka *KeyframeAnimation) Value() float32 {
	if len(ka.Keyframes) == 0 {
		return 0
	}
	if len(ka.Keyframes) == 1 {
		return ka.Keyframes[0].Value
	}
	t := ka.progress()
	if ka.isReversed() {
		t = 1 - t
	}

	// Find surrounding keyframes
	var prev, next Keyframe
	prev = ka.Keyframes[0]
	next = ka.Keyframes[len(ka.Keyframes)-1]
	for i := 0; i < len(ka.Keyframes)-1; i++ {
		if t >= ka.Keyframes[i].Offset && t <= ka.Keyframes[i+1].Offset {
			prev = ka.Keyframes[i]
			next = ka.Keyframes[i+1]
			break
		}
	}

	span := next.Offset - prev.Offset
	if span <= 0 {
		return next.Value
	}
	local := (t - prev.Offset) / span
	if prev.Easing != nil {
		local = prev.Easing(local)
	}
	return prev.Value + (next.Value-prev.Value)*local
}

func (ka *KeyframeAnimation) State() State   { return ka.state }
func (ka *KeyframeAnimation) IsFinished() bool { return ka.state == StateFinished }
func (ka *KeyframeAnimation) OnUpdate(fn func(float32)) { ka.onUpdate = fn }
func (ka *KeyframeAnimation) OnFinish(fn func())        { ka.onFinish = fn }

func (ka *KeyframeAnimation) Pause() {
	if ka.state == StateRunning {
		ka.state = StatePaused
		ka.paused = true
	}
}

func (ka *KeyframeAnimation) Resume() {
	if ka.paused {
		ka.state = StateRunning
		ka.paused = false
	}
}

func (ka *KeyframeAnimation) progress() float32 {
	active := ka.elapsed - ka.Delay
	if active < 0 {
		return 0
	}
	if ka.Duration <= 0 {
		return 1
	}
	t := active / ka.Duration
	if t > 1 {
		t = 1
	}
	return t
}

func (ka *KeyframeAnimation) isReversed() bool {
	switch ka.Direction {
	case DirReverse:
		return true
	case DirAlternate:
		return ka.iteration%2 == 1
	case DirAlternateReverse:
		return ka.iteration%2 == 0
	}
	return false
}

func (ka *KeyframeAnimation) tick(dt float32) {
	if ka.state == StateFinished || ka.state == StatePaused {
		return
	}
	ka.elapsed += dt

	if ka.elapsed < ka.Delay {
		ka.state = StatePending
		return
	}
	ka.state = StateRunning

	active := ka.elapsed - ka.Delay
	if active >= ka.Duration && ka.Duration > 0 {
		if ka.Repeat == -1 || ka.iteration < ka.Repeat {
			ka.iteration++
			ka.elapsed = ka.Delay
			if ka.onUpdate != nil {
				ka.onUpdate(ka.Value())
			}
			return
		}
		ka.state = StateFinished
		if ka.onUpdate != nil {
			ka.onUpdate(ka.Value())
		}
		if ka.onFinish != nil {
			ka.onFinish()
		}
		return
	}

	if ka.onUpdate != nil {
		ka.onUpdate(ka.Value())
	}
}

// Scheduler manages active animations and advances them each frame.
type Scheduler struct {
	anims     []*Animation
	keyframes []*KeyframeAnimation
	nextID    uint64
}

// NewScheduler creates a new animation scheduler.
func NewScheduler() *Scheduler {
	return &Scheduler{nextID: 1}
}

// Add creates and registers a simple from→to animation.
func (s *Scheduler) Add(from, to, duration float32, easing Easing) *Animation {
	a := &Animation{
		ID:       s.nextID,
		From:     from,
		To:       to,
		Duration: duration,
		Easing:   easing,
	}
	s.nextID++
	s.anims = append(s.anims, a)
	return a
}

// AddKeyframes creates and registers a keyframe animation.
func (s *Scheduler) AddKeyframes(keyframes []Keyframe, duration float32) *KeyframeAnimation {
	ka := &KeyframeAnimation{
		ID:        s.nextID,
		Keyframes: keyframes,
		Duration:  duration,
	}
	s.nextID++
	s.keyframes = append(s.keyframes, ka)
	return ka
}

// Remove cancels an animation by ID.
func (s *Scheduler) Remove(id uint64) {
	for i, a := range s.anims {
		if a.ID == id {
			s.anims = append(s.anims[:i], s.anims[i+1:]...)
			return
		}
	}
	for i, ka := range s.keyframes {
		if ka.ID == id {
			s.keyframes = append(s.keyframes[:i], s.keyframes[i+1:]...)
			return
		}
	}
}

// Tick advances all animations by dt seconds and removes finished ones.
func (s *Scheduler) Tick(dt float32) {
	// Tick simple animations
	n := 0
	for _, a := range s.anims {
		a.tick(dt)
		if !a.IsFinished() || a.Fill == FillForwards || a.Fill == FillBoth {
			s.anims[n] = a
			n++
		}
	}
	s.anims = s.anims[:n]

	// Tick keyframe animations
	n = 0
	for _, ka := range s.keyframes {
		ka.tick(dt)
		if !ka.IsFinished() || ka.Fill == FillForwards || ka.Fill == FillBoth {
			s.keyframes[n] = ka
			n++
		}
	}
	s.keyframes = s.keyframes[:n]
}

// Active returns the number of active animations.
func (s *Scheduler) Active() int {
	return len(s.anims) + len(s.keyframes)
}

// HasActive returns true if any animations are running.
func (s *Scheduler) HasActive() bool {
	return s.Active() > 0
}

// Clear removes all animations.
func (s *Scheduler) Clear() {
	s.anims = s.anims[:0]
	s.keyframes = s.keyframes[:0]
}

// Transition is a convenience for creating property transitions.
// It manages a single animated value that can be retargeted mid-flight.
type Transition struct {
	current  float32
	target   float32
	duration float32
	easing   Easing
	anim     *Animation
	sched    *Scheduler
}

// NewTransition creates a transition for a value.
func NewTransition(sched *Scheduler, initial, duration float32, easing Easing) *Transition {
	if easing == nil {
		easing = EaseInOut
	}
	return &Transition{
		current:  initial,
		target:   initial,
		duration: duration,
		easing:   easing,
		sched:    sched,
	}
}

// Set starts transitioning to a new target value.
func (tr *Transition) Set(target float32) {
	if target == tr.target {
		return
	}
	// Cancel any in-flight animation
	if tr.anim != nil {
		tr.current = tr.anim.Value()
		tr.sched.Remove(tr.anim.ID)
	}
	tr.target = target
	tr.anim = tr.sched.Add(tr.current, target, tr.duration, tr.easing)
	tr.anim.OnUpdate(func(v float32) { tr.current = v })
	tr.anim.OnFinish(func() { tr.current = target; tr.anim = nil })
}

// Value returns the current interpolated value.
func (tr *Transition) Value() float32 {
	if tr.anim != nil {
		return tr.anim.Value()
	}
	return tr.current
}

// IsAnimating returns true if a transition is in progress.
func (tr *Transition) IsAnimating() bool {
	return tr.anim != nil && !tr.anim.IsFinished()
}

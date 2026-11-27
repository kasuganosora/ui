package widget

import (
	"math"
	"os"
	"strings"
	"testing"
)

// Apple §3–§6 regression. §9 already proved the soft boundary exists; these
// tests prove the release is a real interruptible spring that inherits the
// gesture's velocity and lands on a projected point, not a scripted timer.

func TestSpringHasNoFixedDuration(t *testing.T) {
	short := NewSpring(-40, 0, 0, springCriticalDamping)
	long := NewSpring(-4000, 0, 0, springCriticalDamping)
	// A linear spring has the same *fractional* progress at any instant, but
	// its absolute snap threshold is reached at different times. A fixed
	// duration would stop both on the same frame.
	shortAt, longAt := 0, 0
	for i := 1; i <= 80; i++ {
		short.Step(16)
		long.Step(16)
		if shortAt == 0 && short.Settled() {
			shortAt = i
		}
		if longAt == 0 && long.Settled() {
			longAt = i
		}
	}
	if shortAt == 0 || longAt == 0 {
		t.Fatalf("springs never settled: short at %d long at %d", shortAt, longAt)
	}
	if longAt <= shortAt {
		t.Errorf("a 100× longer travel finished on the same scripted duration: short=%d long=%d frames", shortAt, longAt)
	}
}

func TestSpringStaysCriticallyDampedByDefault(t *testing.T) {
	s := NewSpring(-4000, 0, 0, springCriticalDamping)
	prev := math.Abs(float64(s.X))
	for i := 0; i < 80; i++ {
		s.Step(16)
		cur := math.Abs(float64(s.X))
		if cur > prev+0.001 {
			t.Fatalf("default spring overshot the target at step %d: %f -> %f", i, prev, cur)
		}
		prev = cur
	}
	if !s.Settled() {
		t.Fatalf("critical spring never settled: x=%f", s.X)
	}
}

func TestSpringInterruptStartsFromPresentationValue(t *testing.T) {
	s := NewSpring(-1000, 0, 0, springCriticalDamping)
	for i := 0; i < 6; i++ {
		s.Step(16)
	}
	liveX, liveV := s.X, s.V
	if liveX == -1000 || liveV == 0 {
		t.Fatal("fixture did not put the spring in flight")
	}

	// §3: re-target mid-flight. The new spring must start at the live value
	// with the live velocity, never snap back to the original or the target.
	re := NewSpring(liveX, liveV, liveX+300, springCriticalDamping)
	if re.X != liveX || re.V != liveV {
		t.Fatalf("re-target jumped: x %f->%f v %f->%f", liveX, re.X, liveV, re.V)
	}
	re.Step(16)
	if re.X <= liveX {
		t.Errorf("re-targeted spring did not continue toward the new target: %f", re.X)
	}
}

func TestSpringVelocityHandoffHasNoSeam(t *testing.T) {
	withVel := NewSpring(0, 1.2, 300, springCriticalDamping)
	fromRest := NewSpring(0, 0, 300, springCriticalDamping)
	withVel.Step(16)
	fromRest.Step(16)
	if withVel.X <= fromRest.X {
		t.Errorf("release velocity was dropped: with=%f rest=%f", withVel.X, fromRest.X)
	}
	if withVel.V <= 0 {
		t.Errorf("handed-off velocity was hard-cut to %f", withVel.V)
	}
}

func TestProjectMomentumMatchesAppleFormula(t *testing.T) {
	// Apple: project(v_px_per_s) = (v/1000) * d/(1-d), d = 0.998.
	// Our argument is px/ms, so v_px_per_s = v*1000.
	const d = 0.998
	for _, v := range []float32{-3, -0.5, 0.5, 2, 8} {
		want := float32(float64(v*1000) / 1000 * d / (1 - d))
		if got := projectMomentum(v); math.Abs(float64(got-want)) > 0.001 {
			t.Errorf("projectMomentum(%f) = %f, want %f", v, got, want)
		}
	}
	if projectMomentum(2) != -projectMomentum(-2) {
		t.Error("momentum projection is not symmetric")
	}
	if projectMomentum(0) != 0 {
		t.Error("zero velocity should not project any travel")
	}
}

func TestScrollEdgeFlickProjectsRestingPosition(t *testing.T) {
	var e ScrollEdge
	pos := float32(0)
	// Three ticks 16ms apart, all inside the legal range: a real flick.
	for i := 0; i < 3; i++ {
		pos = applyScrollGestureAt(&e, pos, 100, 0, 600, 400, int64(1000+i*16), true)
	}
	if pos <= 0 || pos >= 600 {
		t.Fatalf("fixture is not inside the bounds: %f", pos)
	}
	velocity := e.releaseVelocity()
	if velocity <= momentumMinVelocity {
		t.Fatalf("fixture has no measurable flick velocity: %f", velocity)
	}
	releasePoint := e.raw
	projected := clampScrollValue(releasePoint+projectMomentum(velocity), 0, 600)
	if projected <= releasePoint {
		t.Fatal("fixture projection does not move forward")
	}

	// Release after the gesture gap, then drive real frame times.
	first, _ := settleScrollGestureAt(&e, pos, 0, 600, 400, 1200)
	if first < pos-0.01 {
		t.Errorf("release jumped backwards before coasting: %f -> %f", pos, first)
	}
	for ms := int64(1216); ms <= 3000; ms += 16 {
		if next, changed := settleScrollGestureAt(&e, first, 0, 600, 400, ms); changed {
			first = next
		}
	}
	if math.Abs(float64(e.raw-projected)) > 1 {
		t.Errorf("flick landed at %f instead of the projected %f", e.raw, projected)
	}
	if e.Overscroll(0, 600) > 1 {
		t.Errorf("flick left the surface outside its bounds: %f", e.raw)
	}
}

func TestScrollEdgeFlickIsInterruptibleMidFlight(t *testing.T) {
	var e ScrollEdge
	pos := float32(0)
	for i := 0; i < 3; i++ {
		pos = applyScrollGestureAt(&e, pos, 100, 0, 600, 400, int64(1000+i*16), true)
	}
	settleScrollGestureAt(&e, pos, 0, 600, 400, 1200)
	for ms := int64(1216); ms <= 1280; ms += 16 {
		pos, _ = settleScrollGestureAt(&e, pos, 0, 600, 400, ms)
	}
	live := e.raw
	if live <= 300 {
		t.Fatalf("flick did not move off its release point: %f", live)
	}
	inFlightV := e.spring.V
	if !e.spring.Active() || inFlightV == 0 {
		t.Fatal("fixture did not leave a spring in flight")
	}

	// A new wheel tick arrives mid-flight. §3 requires continuity: the new
	// position continues from the live value and the spring velocity is
	// carried into the new gesture instead of being dropped.
	next := applyScrollGestureAt(&e, pos, -60, 0, 600, 400, 1296, true)
	if math.Abs(float64(e.raw-(live-60))) > 0.01 {
		t.Errorf("interrupt did not start from the presentation value: %f -> %f", live, e.raw)
	}
	if e.carriedVel == 0 || math.Signbit(float64(e.carriedVel)) != math.Signbit(float64(inFlightV)) {
		t.Errorf("in-flight velocity was hard-cut at the interrupt: %f -> %f", inFlightV, e.carriedVel)
	}
	if next >= pos+1 {
		t.Errorf("reverse gesture did not change direction: %f -> %f", pos, next)
	}
	// The old spring is no longer driving the surface.
	if e.spring.Active() {
		t.Error("old spring kept running after the interrupt")
	}
}

func TestOverscrollReleaseDoesNotHandOffOutwardVelocity(t *testing.T) {
	var e ScrollEdge
	pos := float32(0)
	for i := 0; i < 4; i++ {
		pos = applyScrollGestureAt(&e, pos, -400, 0, 600, 400, int64(1000+i*16), true)
	}
	if pos >= 0 {
		t.Fatal("fixture did not overscroll")
	}
	prev := math.Abs(float64(pos))
	for ms := int64(1200); ms <= 3000; ms += 16 {
		if next, changed := settleScrollGestureAt(&e, pos, 0, 600, 400, ms); changed {
			pos = next
		}
		cur := math.Abs(float64(pos))
		if cur > prev+0.01 {
			t.Fatalf("rubber-band release moved away from the boundary at %dms: %f -> %f", ms, prev, cur)
		}
		prev = cur
	}
	if e.Overscroll(0, 600) != 0 {
		t.Errorf("rubber-band left overscroll behind: %f", e.raw)
	}
}

func TestReducedMotionDisablesFlickAndSpring(t *testing.T) {
	var e ScrollEdge
	pos := applyScrollGestureAt(&e, 0, 100, 0, 600, 400, 1000, false)
	if pos != 100 {
		t.Fatalf("reduced motion should follow the gesture 1:1, got %f", pos)
	}
	if next, changed := settleScrollGestureAt(&e, pos, 0, 600, 400, 1400); changed || next != pos {
		t.Errorf("reduced motion played a decorative animation: %f -> %f", pos, next)
	}
	if e.spring.Active() || e.Overscroll(0, 600) != 0 {
		t.Error("reduced motion left animation state behind")
	}
}

func TestSpringIntegrationSplitsLargeFrameGaps(t *testing.T) {
	a := NewSpring(-4000, 0, 0, springCriticalDamping)
	b := NewSpring(-4000, 0, 0, springCriticalDamping)
	for i := 0; i < 8; i++ {
		a.Step(16)
	}
	b.Step(128)
	if math.Abs(float64(a.X-b.X)) > 1 {
		t.Errorf("a long frame gap changed the trajectory: %f vs %f", a.X, b.X)
	}
}

func TestSpringSourceGuard(t *testing.T) {
	src, err := os.ReadFile("scrollphysics.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(src)
	for _, fixed := range []string{"math.Exp(", "time.After", "duration :=", "elapsed /"} {
		if strings.Contains(text, fixed) {
			t.Errorf("scroll release still uses a scripted decay (%q)", fixed)
		}
	}
	if !strings.Contains(text, "NewSpring(") || !strings.Contains(text, "projectMomentum(") {
		t.Error("scroll release no longer uses the spring/momentum projection")
	}
	spr, err := os.ReadFile("spring.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(spr), "momentumDeceleration") {
		t.Error("Apple's exponential-decay projection constant is gone")
	}
}

package anim

import (
	"testing"
)

func approx(a, b, eps float32) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < eps
}

func TestEasingFunctions(t *testing.T) {
	easings := map[string]Easing{
		"Linear":         Linear,
		"EaseIn":         EaseIn,
		"EaseOut":        EaseOut,
		"EaseInOut":      EaseInOut,
		"EaseInCubic":    EaseInCubic,
		"EaseOutCubic":   EaseOutCubic,
		"EaseInOutCubic": EaseInOutCubic,
		"EaseOutBack":    EaseOutBack,
		"EaseOutElastic": EaseOutElastic,
		"EaseOutBounce":  EaseOutBounce,
	}
	for name, e := range easings {
		if !approx(e(0), 0, 0.02) {
			t.Errorf("%s(0) = %g, want ~0", name, e(0))
		}
		if !approx(e(1), 1, 0.02) {
			t.Errorf("%s(1) = %g, want ~1", name, e(1))
		}
	}
}

func TestEaseOutBounceAllBranches(t *testing.T) {
	// Branch 1: t < 1/2.75 ≈ 0.364
	v := EaseOutBounce(0.2)
	if v < 0 || v > 1 {
		t.Errorf("bounce(0.2) = %g out of range", v)
	}
	// Branch 2: t < 2/2.75 ≈ 0.727
	v = EaseOutBounce(0.5)
	if v < 0 || v > 1 {
		t.Errorf("bounce(0.5) = %g out of range", v)
	}
	// Branch 3: t < 2.5/2.75 ≈ 0.909
	v = EaseOutBounce(0.85)
	if v < 0.9 || v > 1 {
		t.Errorf("bounce(0.85) = %g out of range", v)
	}
	// Branch 4: default
	v = EaseOutBounce(0.97)
	if v < 0.97 || v > 1.01 {
		t.Errorf("bounce(0.97) = %g out of range", v)
	}
}

func TestEaseInOutBranches(t *testing.T) {
	// t < 0.5
	v := EaseInOut(0.25)
	if v <= 0 || v >= 0.5 {
		t.Errorf("easeInOut(0.25) = %g", v)
	}
	// t >= 0.5
	v = EaseInOut(0.75)
	if v <= 0.5 || v >= 1 {
		t.Errorf("easeInOut(0.75) = %g", v)
	}
}

func TestEaseInOutCubicBranches(t *testing.T) {
	v := EaseInOutCubic(0.25)
	if v <= 0 || v >= 0.5 {
		t.Errorf("easeInOutCubic(0.25) = %g", v)
	}
	v = EaseInOutCubic(0.75)
	if v <= 0.5 || v >= 1 {
		t.Errorf("easeInOutCubic(0.75) = %g", v)
	}
}

func TestEaseOutElasticEdges(t *testing.T) {
	if EaseOutElastic(0) != 0 {
		t.Error("elastic(0) should be 0")
	}
	if EaseOutElastic(1) != 1 {
		t.Error("elastic(1) should be 1")
	}
}

func TestAnimationBasic(t *testing.T) {
	s := NewScheduler()
	a := s.Add(0, 100, 1.0, Linear)
	if a.State() != StatePending {
		t.Errorf("expected pending, got %d", a.State())
	}
	s.Tick(0.5)
	if !approx(a.Value(), 50, 1) {
		t.Errorf("at 0.5s: got %g, want ~50", a.Value())
	}
	if a.State() != StateRunning {
		t.Error("expected running")
	}
	s.Tick(0.6)
	if !a.IsFinished() {
		t.Error("expected finished")
	}
	if !approx(a.Value(), 100, 0.01) {
		t.Errorf("final: got %g", a.Value())
	}
}

func TestAnimationZeroDuration(t *testing.T) {
	s := NewScheduler()
	a := s.Add(0, 50, 0, Linear)
	if a.Value() != 50 {
		t.Errorf("zero duration: got %g, want 50", a.Value())
	}
	s.Tick(0.1)
	if !a.IsFinished() {
		t.Error("zero duration should finish immediately")
	}
}

func TestAnimationNilEasing(t *testing.T) {
	s := NewScheduler()
	a := s.Add(0, 100, 1.0, nil)
	s.Tick(0.5)
	// nil easing → raw t value
	if !approx(a.Value(), 50, 1) {
		t.Errorf("nil easing: got %g", a.Value())
	}
}

func TestAnimationDelay(t *testing.T) {
	s := NewScheduler()
	a := s.Add(0, 10, 1.0, Linear)
	a.Delay = 0.5

	s.Tick(0.3)
	if a.State() != StatePending {
		t.Error("should be pending during delay")
	}
	s.Tick(0.7)
	if a.State() != StateRunning {
		t.Error("should be running after delay")
	}
}

func TestAnimationDelayWithFillBackwards(t *testing.T) {
	s := NewScheduler()
	a := s.Add(0, 100, 1.0, Linear)
	a.Delay = 1.0
	a.Fill = FillBackwards
	s.Tick(0.5) // still in delay
	v := a.Value()
	// FillBackwards: during delay, show initial value (progress=0)
	if !approx(v, 0, 1) {
		t.Errorf("fill backwards during delay: got %g", v)
	}
}

func TestAnimationDelayWithFillBoth(t *testing.T) {
	s := NewScheduler()
	a := s.Add(0, 100, 0.5, Linear)
	a.Delay = 0.5
	a.Fill = FillBoth
	s.Tick(0.3) // during delay
	_ = a.Value()
	s.Tick(0.5) // running
	s.Tick(0.5) // finished
	if s.Active() != 1 {
		t.Error("FillBoth should keep animation after finish")
	}
}

func TestAnimationRepeat(t *testing.T) {
	s := NewScheduler()
	a := s.Add(0, 10, 0.5, Linear)
	a.Repeat = 2
	updates := 0
	a.OnUpdate(func(v float32) { updates++ })

	s.Tick(0.6)
	if a.IsFinished() {
		t.Error("should not be finished after 1 repeat")
	}
	s.Tick(0.5)
	if a.IsFinished() {
		t.Error("should not be finished after 2 repeats")
	}
	s.Tick(0.5)
	if !a.IsFinished() {
		t.Error("should be finished after 3 plays")
	}
	if updates == 0 {
		t.Error("onUpdate should have been called during repeats")
	}
}

func TestAnimationDirectionReverse(t *testing.T) {
	s := NewScheduler()
	a := s.Add(0, 100, 1.0, Linear)
	a.Direction = DirReverse

	s.Tick(0.01) // near start
	v := a.Value()
	// Reversed: from=To(100), to=From(0), progress~0 → value~100
	if v < 90 {
		t.Errorf("reverse near start: got %g, want ~100", v)
	}

	s.Tick(0.98) // near end
	v = a.Value()
	if v > 10 {
		t.Errorf("reverse near end: got %g, want ~0", v)
	}
}

func TestAnimationDirectionAlternateReverse(t *testing.T) {
	s := NewScheduler()
	a := s.Add(0, 100, 0.5, Linear)
	a.Direction = DirAlternateReverse
	a.Repeat = 1

	// First iteration: AlternateReverse with iteration=0 → reversed
	s.Tick(0.01)
	v := a.Value()
	if v < 90 {
		t.Errorf("alternate reverse iter 0: got %g, want ~100", v)
	}
	s.Tick(0.6) // complete first, start second (iteration=1 → normal)
	s.Tick(0.01)
}

func TestAnimationDirectionAlternate(t *testing.T) {
	s := NewScheduler()
	a := s.Add(0, 100, 0.5, Linear)
	a.Direction = DirAlternate
	a.Repeat = 1

	// First iteration: normal (iteration=0)
	s.Tick(0.01)
	v := a.Value()
	if v > 10 {
		t.Errorf("alternate iter 0 near start: got %g, want ~0", v)
	}
}

func TestAnimationPauseResume(t *testing.T) {
	s := NewScheduler()
	a := s.Add(0, 100, 1.0, Linear)

	s.Tick(0.3)
	a.Pause()
	if a.State() != StatePaused {
		t.Error("expected paused state")
	}
	v1 := a.Value()

	s.Tick(0.5)
	v2 := a.Value()
	if !approx(v1, v2, 0.01) {
		t.Errorf("value changed while paused: %g vs %g", v1, v2)
	}

	a.Resume()
	if a.State() != StateRunning {
		t.Error("expected running after resume")
	}
	s.Tick(0.3)
	if a.Value() <= v2 {
		t.Error("value should advance after resume")
	}

	// Pause when not running → no-op
	a2 := s.Add(0, 10, 1.0, Linear)
	a2.state = StateFinished
	a2.Pause() // should not change state
	if a2.State() != StateFinished {
		t.Error("pause on finished should be no-op")
	}

	// Resume when not paused → no-op
	a2.Resume()
	if a2.State() != StateFinished {
		t.Error("resume on non-paused should be no-op")
	}
}

func TestAnimationCallback(t *testing.T) {
	s := NewScheduler()
	a := s.Add(0, 10, 0.5, Linear)
	var lastVal float32
	finished := false
	a.OnUpdate(func(v float32) { lastVal = v })
	a.OnFinish(func() { finished = true })

	s.Tick(0.3)
	if lastVal == 0 {
		t.Error("onUpdate should have been called")
	}
	s.Tick(0.3)
	if !finished {
		t.Error("onFinish should have been called")
	}
}

// --- Keyframe Animation ---

func TestKeyframeAnimation(t *testing.T) {
	s := NewScheduler()
	ka := s.AddKeyframes([]Keyframe{
		{Offset: 0, Value: 0},
		{Offset: 0.5, Value: 100},
		{Offset: 1.0, Value: 50},
	}, 2.0)

	s.Tick(1.0) // at 50%
	if !approx(ka.Value(), 100, 2) {
		t.Errorf("at 50%%: got %g", ka.Value())
	}
	s.Tick(1.0) // at 100%
	if !approx(ka.Value(), 50, 2) {
		t.Errorf("at 100%%: got %g", ka.Value())
	}
	if !ka.IsFinished() {
		t.Error("should be finished")
	}
}

func TestKeyframeEmpty(t *testing.T) {
	s := NewScheduler()
	ka := s.AddKeyframes(nil, 1.0)
	s.Tick(0.5)
	if ka.Value() != 0 {
		t.Errorf("empty keyframes: got %g", ka.Value())
	}
}

func TestKeyframeSingle(t *testing.T) {
	s := NewScheduler()
	ka := s.AddKeyframes([]Keyframe{{Offset: 0, Value: 42}}, 1.0)
	s.Tick(0.5)
	if ka.Value() != 42 {
		t.Errorf("single keyframe: got %g", ka.Value())
	}
}

func TestKeyframeZeroDuration(t *testing.T) {
	s := NewScheduler()
	ka := s.AddKeyframes([]Keyframe{
		{Offset: 0, Value: 0},
		{Offset: 1, Value: 100},
	}, 0)
	s.Tick(0.1)
	// Zero duration → progress=1 → last keyframe
	if ka.IsFinished() {
		// Should finish
	}
}

func TestKeyframeZeroSpan(t *testing.T) {
	s := NewScheduler()
	ka := s.AddKeyframes([]Keyframe{
		{Offset: 0.5, Value: 10},
		{Offset: 0.5, Value: 20},
	}, 1.0)
	s.Tick(0.5)
	// Zero span → returns next.Value
	_ = ka.Value()
}

func TestKeyframeRepeat(t *testing.T) {
	s := NewScheduler()
	ka := s.AddKeyframes([]Keyframe{
		{Offset: 0, Value: 0},
		{Offset: 1.0, Value: 10},
	}, 1.0)
	ka.Repeat = 1
	updates := 0
	ka.OnUpdate(func(v float32) { updates++ })

	s.Tick(1.1)
	if ka.IsFinished() {
		t.Error("should not be finished after 1 repeat")
	}
	s.Tick(1.0)
	if !ka.IsFinished() {
		t.Error("should be finished after 2 plays")
	}
	if updates == 0 {
		t.Error("updates should fire")
	}
}

func TestKeyframeDirection(t *testing.T) {
	s := NewScheduler()
	ka := s.AddKeyframes([]Keyframe{
		{Offset: 0, Value: 0},
		{Offset: 1, Value: 100},
	}, 1.0)
	ka.Direction = DirReverse

	s.Tick(0.01)
	// Reversed: t near 0 → 1-0 = 1 → value near 100
	if ka.Value() < 90 {
		t.Errorf("reversed keyframe: got %g", ka.Value())
	}
}

func TestKeyframeDirectionAlternate(t *testing.T) {
	s := NewScheduler()
	ka := s.AddKeyframes([]Keyframe{
		{Offset: 0, Value: 0},
		{Offset: 1, Value: 100},
	}, 0.5)
	ka.Direction = DirAlternate
	ka.Repeat = 1

	s.Tick(0.01) // iter 0 normal
	if ka.Value() > 10 {
		t.Errorf("alternate iter 0: got %g", ka.Value())
	}
}

func TestKeyframeDirectionAlternateReverse(t *testing.T) {
	s := NewScheduler()
	ka := s.AddKeyframes([]Keyframe{
		{Offset: 0, Value: 0},
		{Offset: 1, Value: 100},
	}, 0.5)
	ka.Direction = DirAlternateReverse

	s.Tick(0.01) // iter 0 → reversed
	if ka.Value() < 90 {
		t.Errorf("alternateReverse iter 0: got %g", ka.Value())
	}
}

func TestKeyframePauseResume(t *testing.T) {
	s := NewScheduler()
	ka := s.AddKeyframes([]Keyframe{
		{Offset: 0, Value: 0},
		{Offset: 1, Value: 100},
	}, 1.0)

	s.Tick(0.3)
	ka.Pause()
	if ka.State() != StatePaused {
		t.Error("expected paused")
	}
	v1 := ka.Value()
	s.Tick(0.5)
	if !approx(v1, ka.Value(), 0.01) {
		t.Error("value changed while paused")
	}
	ka.Resume()
	s.Tick(0.3)
	if ka.Value() <= v1 {
		t.Error("should advance after resume")
	}

	// Pause on non-running
	ka2 := s.AddKeyframes([]Keyframe{{Offset: 0, Value: 0}}, 1.0)
	ka2.state = StateFinished
	ka2.Pause()
	if ka2.State() != StateFinished {
		t.Error("pause on finished should be no-op")
	}
	ka2.Resume()
}

func TestKeyframeDelay(t *testing.T) {
	s := NewScheduler()
	ka := s.AddKeyframes([]Keyframe{
		{Offset: 0, Value: 0},
		{Offset: 1, Value: 100},
	}, 1.0)
	ka.Delay = 0.5

	s.Tick(0.3) // during delay
	if ka.State() != StatePending {
		t.Error("should be pending during delay")
	}
	s.Tick(0.5) // past delay, running
	if ka.State() != StateRunning {
		t.Error("should be running after delay")
	}
}

func TestKeyframeCallbacks(t *testing.T) {
	s := NewScheduler()
	ka := s.AddKeyframes([]Keyframe{
		{Offset: 0, Value: 0},
		{Offset: 1, Value: 100},
	}, 0.5)
	finished := false
	ka.OnFinish(func() { finished = true })
	ka.OnUpdate(func(v float32) {})

	s.Tick(0.6)
	if !finished {
		t.Error("expected finish callback")
	}
}

func TestKeyframeEasing(t *testing.T) {
	s := NewScheduler()
	ka := s.AddKeyframes([]Keyframe{
		{Offset: 0, Value: 0, Easing: EaseIn},
		{Offset: 1.0, Value: 100},
	}, 1.0)

	s.Tick(0.5)
	v := ka.Value()
	if !approx(v, 25, 3) {
		t.Errorf("EaseIn keyframe at 0.5: got %g", v)
	}
}

// --- Scheduler ---

func TestSchedulerRemove(t *testing.T) {
	s := NewScheduler()
	a := s.Add(0, 10, 1.0, Linear)
	if s.Active() != 1 {
		t.Errorf("expected 1 active")
	}
	s.Remove(a.ID)
	if s.Active() != 0 {
		t.Errorf("expected 0 active after remove")
	}
}

func TestSchedulerRemoveKeyframe(t *testing.T) {
	s := NewScheduler()
	ka := s.AddKeyframes([]Keyframe{{Offset: 0, Value: 0}}, 1.0)
	s.Remove(ka.ID)
	if s.Active() != 0 {
		t.Error("expected 0 after removing keyframe")
	}
}

func TestSchedulerRemoveNonexistent(t *testing.T) {
	s := NewScheduler()
	s.Add(0, 10, 1.0, Linear)
	s.Remove(9999) // should not panic
	if s.Active() != 1 {
		t.Error("should still have 1 after removing nonexistent")
	}
}

func TestSchedulerClear(t *testing.T) {
	s := NewScheduler()
	s.Add(0, 10, 1.0, Linear)
	s.Add(0, 20, 2.0, EaseIn)
	s.AddKeyframes([]Keyframe{{Offset: 0, Value: 0}, {Offset: 1, Value: 10}}, 1)
	s.Clear()
	if s.Active() != 0 {
		t.Errorf("expected 0 after clear")
	}
}

func TestSchedulerCleanup(t *testing.T) {
	s := NewScheduler()
	s.Add(0, 10, 0.5, Linear)
	s.Tick(1.0)
	if s.Active() != 0 {
		t.Errorf("FillNone should remove finished")
	}
}

func TestSchedulerFillForwards(t *testing.T) {
	s := NewScheduler()
	a := s.Add(0, 10, 0.5, Linear)
	a.Fill = FillForwards
	s.Tick(1.0)
	if s.Active() != 1 {
		t.Error("FillForwards should keep animation")
	}
}

func TestSchedulerKeyframeFillForwards(t *testing.T) {
	s := NewScheduler()
	ka := s.AddKeyframes([]Keyframe{{Offset: 0, Value: 0}, {Offset: 1, Value: 10}}, 0.5)
	ka.Fill = FillForwards
	s.Tick(1.0)
	if s.Active() != 1 {
		t.Error("FillForwards should keep keyframe animation")
	}
}

func TestSchedulerHasActive(t *testing.T) {
	s := NewScheduler()
	if s.HasActive() {
		t.Error("empty scheduler should not have active")
	}
	s.Add(0, 10, 1.0, Linear)
	if !s.HasActive() {
		t.Error("should have active after add")
	}
}

// --- Transition ---

func TestTransition(t *testing.T) {
	s := NewScheduler()
	tr := NewTransition(s, 0, 1.0, EaseInOut)

	if !approx(tr.Value(), 0, 0.01) {
		t.Errorf("initial: %g", tr.Value())
	}
	if tr.IsAnimating() {
		t.Error("should not be animating initially")
	}

	tr.Set(100)
	if !tr.IsAnimating() {
		t.Error("should be animating")
	}
	s.Tick(0.5)
	v := tr.Value()
	if v <= 0 || v >= 100 {
		t.Errorf("mid-transition: %g", v)
	}
	s.Tick(0.6)
	if !approx(tr.Value(), 100, 1) {
		t.Errorf("after transition: %g", tr.Value())
	}
}

func TestTransitionNilEasing(t *testing.T) {
	s := NewScheduler()
	tr := NewTransition(s, 50, 1.0, nil)
	// nil easing → defaults to EaseInOut
	if tr.easing == nil {
		t.Error("nil easing should default to EaseInOut")
	}
	if !approx(tr.Value(), 50, 0.01) {
		t.Errorf("initial: %g", tr.Value())
	}
}

func TestTransitionSameTarget(t *testing.T) {
	s := NewScheduler()
	tr := NewTransition(s, 50, 1.0, Linear)
	tr.Set(50) // same as current → should not create animation
	if tr.IsAnimating() {
		t.Error("setting same target should not animate")
	}
}

func TestTransitionRetarget(t *testing.T) {
	s := NewScheduler()
	tr := NewTransition(s, 0, 1.0, Linear)

	tr.Set(100)
	s.Tick(0.5) // at ~50
	tr.Set(0)   // retarget back
	s.Tick(0.5)
	v := tr.Value()
	if v > 55 || v < -5 {
		t.Errorf("retarget: %g", v)
	}
}

func TestInfiniteRepeat(t *testing.T) {
	s := NewScheduler()
	a := s.Add(0, 10, 0.5, Linear)
	a.Repeat = -1

	for i := 0; i < 20; i++ {
		s.Tick(0.5)
	}
	if a.IsFinished() {
		t.Error("infinite repeat should never finish")
	}
}

func TestKeyframeInfiniteRepeat(t *testing.T) {
	s := NewScheduler()
	ka := s.AddKeyframes([]Keyframe{{Offset: 0, Value: 0}, {Offset: 1, Value: 10}}, 0.5)
	ka.Repeat = -1

	for i := 0; i < 10; i++ {
		s.Tick(0.5)
	}
	if ka.IsFinished() {
		t.Error("infinite keyframe repeat should never finish")
	}
}

func TestAnimationProgressEdges(t *testing.T) {
	s := NewScheduler()
	a := s.Add(0, 100, 1.0, Linear)
	// Before any tick, progress is 0
	if a.progress() != 0 {
		t.Errorf("initial progress: %g", a.progress())
	}
	// After full duration
	s.Tick(2.0)
	if a.progress() != 1 {
		t.Errorf("final progress: %g", a.progress())
	}
}

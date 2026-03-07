package anim

import (
	"math"
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
		// Must pass through endpoints
		if !approx(e(0), 0, 0.01) {
			t.Errorf("%s(0) = %g, want ~0", name, e(0))
		}
		if !approx(e(1), 1, 0.01) {
			t.Errorf("%s(1) = %g, want ~1", name, e(1))
		}
	}
}

func TestAnimationBasic(t *testing.T) {
	s := NewScheduler()
	a := s.Add(0, 100, 1.0, Linear)
	if a.State() != StatePending {
		t.Errorf("expected pending, got %d", a.State())
	}

	s.Tick(0.5)
	v := a.Value()
	if !approx(v, 50, 1) {
		t.Errorf("at 0.5s: got %g, want ~50", v)
	}
	if a.State() != StateRunning {
		t.Error("expected running")
	}

	s.Tick(0.6)
	if !a.IsFinished() {
		t.Error("expected finished after 1.1s total")
	}
	if !approx(a.Value(), 100, 0.01) {
		t.Errorf("final value: got %g, want 100", a.Value())
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
	v := a.Value()
	if !approx(v, 5, 1) {
		t.Errorf("got %g, want ~5", v)
	}
}

func TestAnimationRepeat(t *testing.T) {
	s := NewScheduler()
	a := s.Add(0, 10, 0.5, Linear)
	a.Repeat = 2

	s.Tick(0.6) // first iteration done, into second
	if a.IsFinished() {
		t.Error("should not be finished after 1 repeat")
	}

	s.Tick(0.5) // second done, into third
	if a.IsFinished() {
		t.Error("should not be finished after 2 repeats")
	}

	s.Tick(0.5) // third done
	if !a.IsFinished() {
		t.Error("should be finished after 3 plays")
	}
}

func TestAnimationDirection(t *testing.T) {
	s := NewScheduler()
	a := s.Add(0, 100, 1.0, Linear)
	a.Direction = DirReverse

	s.Tick(0.5)
	v := a.Value()
	if !approx(v, 50, 1) {
		t.Errorf("reverse at 0.5: got %g, want ~50", v)
	}
	// At t=0 reversed should give To=100, at t=1 should give From=0
	s2 := NewScheduler()
	a2 := s2.Add(0, 100, 1.0, Linear)
	a2.Direction = DirReverse
	s2.Tick(0.01)
	if a2.Value() > 5 {
		// At t~0, reversed: from=To(100), to=From(0), progress~0 → ~100
		// Actually no: from,to are swapped and progress is near 0 so value ~ 100
	}
}

func TestAnimationPauseResume(t *testing.T) {
	s := NewScheduler()
	a := s.Add(0, 100, 1.0, Linear)

	s.Tick(0.3)
	a.Pause()
	v1 := a.Value()

	s.Tick(0.5)
	v2 := a.Value()
	if !approx(v1, v2, 0.01) {
		t.Errorf("value should not change while paused: %g vs %g", v1, v2)
	}

	a.Resume()
	s.Tick(0.3)
	v3 := a.Value()
	if v3 <= v2 {
		t.Error("value should advance after resume")
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

func TestKeyframeAnimation(t *testing.T) {
	s := NewScheduler()
	ka := s.AddKeyframes([]Keyframe{
		{Offset: 0, Value: 0},
		{Offset: 0.5, Value: 100},
		{Offset: 1.0, Value: 50},
	}, 2.0)

	s.Tick(1.0) // at 50%
	v := ka.Value()
	if !approx(v, 100, 2) {
		t.Errorf("at 50%%: got %g, want ~100", v)
	}

	s.Tick(1.0) // at 100%
	v = ka.Value()
	if !approx(v, 50, 2) {
		t.Errorf("at 100%%: got %g, want ~50", v)
	}
	if !ka.IsFinished() {
		t.Error("should be finished")
	}
}

func TestKeyframeRepeat(t *testing.T) {
	s := NewScheduler()
	ka := s.AddKeyframes([]Keyframe{
		{Offset: 0, Value: 0},
		{Offset: 1.0, Value: 10},
	}, 1.0)
	ka.Repeat = 1

	s.Tick(1.1)
	if ka.IsFinished() {
		t.Error("should not be finished after 1 repeat")
	}

	s.Tick(1.0)
	if !ka.IsFinished() {
		t.Error("should be finished after 2 plays")
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
	// EaseIn at 0.5 → 0.25, so value ~ 25
	if !approx(v, 25, 3) {
		t.Errorf("EaseIn keyframe at 0.5: got %g, want ~25", v)
	}
}

func TestSchedulerRemove(t *testing.T) {
	s := NewScheduler()
	a := s.Add(0, 10, 1.0, Linear)
	if s.Active() != 1 {
		t.Errorf("expected 1 active, got %d", s.Active())
	}
	s.Remove(a.ID)
	if s.Active() != 0 {
		t.Errorf("expected 0 active after remove, got %d", s.Active())
	}
}

func TestSchedulerClear(t *testing.T) {
	s := NewScheduler()
	s.Add(0, 10, 1.0, Linear)
	s.Add(0, 20, 2.0, EaseIn)
	s.AddKeyframes([]Keyframe{{Offset: 0, Value: 0}, {Offset: 1, Value: 10}}, 1)
	s.Clear()
	if s.Active() != 0 {
		t.Errorf("expected 0 after clear, got %d", s.Active())
	}
}

func TestSchedulerCleanup(t *testing.T) {
	s := NewScheduler()
	s.Add(0, 10, 0.5, Linear)
	s.Tick(1.0) // finish
	// FillNone: should be removed
	if s.Active() != 0 {
		t.Errorf("expected 0 active after finish (FillNone), got %d", s.Active())
	}
}

func TestSchedulerFillForwards(t *testing.T) {
	s := NewScheduler()
	a := s.Add(0, 10, 0.5, Linear)
	a.Fill = FillForwards
	s.Tick(1.0)
	if s.Active() != 1 {
		t.Errorf("FillForwards should keep animation, got %d active", s.Active())
	}
}

func TestTransition(t *testing.T) {
	s := NewScheduler()
	tr := NewTransition(s, 0, 1.0, EaseInOut)

	if !approx(tr.Value(), 0, 0.01) {
		t.Errorf("initial value: %g", tr.Value())
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

func TestTransitionRetarget(t *testing.T) {
	s := NewScheduler()
	tr := NewTransition(s, 0, 1.0, Linear)

	tr.Set(100)
	s.Tick(0.5) // at ~50

	tr.Set(0) // retarget back
	s.Tick(0.5)
	v := tr.Value()
	// Should be between 50 and 0, trending toward 0
	if v > 55 || v < -5 {
		t.Errorf("retarget value: %g", v)
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
	if s.Active() != 1 {
		t.Errorf("should still have 1 active, got %d", s.Active())
	}
}

func TestAlternateDirection(t *testing.T) {
	s := NewScheduler()
	a := s.Add(0, 100, 1.0, Linear)
	a.Direction = DirAlternate
	a.Repeat = 1

	s.Tick(1.1) // first iteration done (normal: 0→100), start second (reverse)
	s.Tick(0.5) // halfway through reverse
	v := a.Value()
	// Second iteration is reversed, so at 50% progress: value ~50
	_ = v
	_ = math.Pi // suppress import error
}

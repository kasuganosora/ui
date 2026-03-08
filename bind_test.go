package ui

import (
	"testing"
)

func TestStateGetSet(t *testing.T) {
	s := NewState(42)
	if s.Get() != 42 {
		t.Errorf("expected 42, got %d", s.Get())
	}
	s.Set(100)
	if s.Get() != 100 {
		t.Errorf("expected 100, got %d", s.Get())
	}
}

func TestStateWatch(t *testing.T) {
	s := NewState("")
	var received string
	s.Watch(func(v string) { received = v })

	s.Set("hello")
	if received != "hello" {
		t.Errorf("expected 'hello', got %q", received)
	}
}

func TestStateNoNotifyOnSameValue(t *testing.T) {
	s := NewState(5)
	calls := 0
	s.Watch(func(v int) { calls++ })

	s.Set(5)
	if calls != 0 {
		t.Errorf("expected 0 calls on same value, got %d", calls)
	}
	s.Set(10)
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestStateUnsubscribe(t *testing.T) {
	s := NewState(0)
	calls := 0
	unsub := s.Watch(func(v int) { calls++ })

	s.Set(1)
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}

	unsub()
	s.Set(2)
	if calls != 1 {
		t.Errorf("expected still 1 call after unsub, got %d", calls)
	}
}

func TestBind(t *testing.T) {
	s := NewState("init")
	var applied string
	Bind(s, func(v string) { applied = v })

	if applied != "init" {
		t.Errorf("Bind should apply current value, got %q", applied)
	}

	s.Set("updated")
	if applied != "updated" {
		t.Errorf("expected 'updated', got %q", applied)
	}
}

func TestComputed(t *testing.T) {
	s := NewState(5)
	doubled := Computed(s, func(v int) int { return v * 2 })

	if doubled.Get() != 10 {
		t.Errorf("expected 10, got %d", doubled.Get())
	}

	s.Set(7)
	if doubled.Get() != 14 {
		t.Errorf("expected 14, got %d", doubled.Get())
	}
}

func TestListState(t *testing.T) {
	ls := NewListState([]string{"a", "b"})

	if ls.Len() != 2 {
		t.Errorf("expected 2 items, got %d", ls.Len())
	}

	var received []string
	ls.Watch(func(items []string) { received = items })

	ls.Append("c")
	if ls.Len() != 3 {
		t.Errorf("expected 3 items, got %d", ls.Len())
	}
	if len(received) != 3 {
		t.Errorf("observer should have received 3 items, got %d", len(received))
	}

	ls.RemoveAt(0)
	if ls.Len() != 2 {
		t.Errorf("expected 2 items after remove, got %d", ls.Len())
	}
	items := ls.Get()
	if items[0] != "b" {
		t.Errorf("expected first item 'b', got %q", items[0])
	}
}

func TestListStateSet(t *testing.T) {
	ls := NewListState([]int{1, 2, 3})
	ls.Set([]int{10, 20})

	if ls.Len() != 2 {
		t.Errorf("expected 2 items, got %d", ls.Len())
	}
}

func TestListStateRemoveOutOfBounds(t *testing.T) {
	ls := NewListState([]int{1})
	ls.RemoveAt(-1)
	ls.RemoveAt(5)
	if ls.Len() != 1 {
		t.Errorf("expected 1 item, got %d", ls.Len())
	}
}

func TestListStateWatchUnsubscribe(t *testing.T) {
	ls := NewListState([]string{"a"})
	calls := 0
	unsub := ls.Watch(func(items []string) { calls++ })

	ls.Set([]string{"b"})
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}

	unsub()
	ls.Set([]string{"c"})
	if calls != 1 {
		t.Errorf("expected still 1 call after unsub, got %d", calls)
	}
}

func TestListStateSetNotifiesObservers(t *testing.T) {
	ls := NewListState([]int{})
	var received []int
	ls.Watch(func(items []int) { received = items })

	ls.Set([]int{10, 20})
	if len(received) != 2 {
		t.Errorf("expected 2 items in notification, got %d", len(received))
	}
}

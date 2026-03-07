package ui

import "sync"

// State is a reactive value container. When the value changes,
// all registered observers are notified.
type State[T comparable] struct {
	mu        sync.RWMutex
	value     T
	observers []func(T)
}

// NewState creates a reactive state with an initial value.
func NewState[T comparable](initial T) *State[T] {
	return &State[T]{value: initial}
}

// Get returns the current value.
func (s *State[T]) Get() T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value
}

// Set updates the value and notifies observers if changed.
func (s *State[T]) Set(v T) {
	s.mu.Lock()
	if s.value == v {
		s.mu.Unlock()
		return
	}
	s.value = v
	// Copy observers under lock to avoid holding lock during callbacks
	obs := make([]func(T), len(s.observers))
	copy(obs, s.observers)
	s.mu.Unlock()

	for _, fn := range obs {
		fn(v)
	}
}

// Watch registers an observer that is called when the value changes.
// Returns an unsubscribe function.
func (s *State[T]) Watch(fn func(T)) func() {
	s.mu.Lock()
	s.observers = append(s.observers, fn)
	idx := len(s.observers) - 1
	s.mu.Unlock()

	return func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if idx < len(s.observers) {
			s.observers = append(s.observers[:idx], s.observers[idx+1:]...)
		}
	}
}

// Bind connects a State to a setter function. When the state changes,
// the setter is called with the new value.
func Bind[T comparable](state *State[T], setter func(T)) func() {
	// Apply current value immediately
	setter(state.Get())
	return state.Watch(setter)
}

// Computed creates a derived state that updates when the source changes.
func Computed[S comparable, T comparable](source *State[S], transform func(S) T) *State[T] {
	derived := NewState(transform(source.Get()))
	source.Watch(func(v S) {
		derived.Set(transform(v))
	})
	return derived
}

// ListState is a reactive list container.
type ListState[T any] struct {
	mu        sync.RWMutex
	items     []T
	observers []func([]T)
}

// NewListState creates a reactive list.
func NewListState[T any](initial []T) *ListState[T] {
	items := make([]T, len(initial))
	copy(items, initial)
	return &ListState[T]{items: items}
}

// Get returns a copy of the current items.
func (ls *ListState[T]) Get() []T {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	out := make([]T, len(ls.items))
	copy(out, ls.items)
	return out
}

// Len returns the number of items.
func (ls *ListState[T]) Len() int {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	return len(ls.items)
}

// Set replaces all items and notifies observers.
func (ls *ListState[T]) Set(items []T) {
	ls.mu.Lock()
	ls.items = make([]T, len(items))
	copy(ls.items, items)
	obs := make([]func([]T), len(ls.observers))
	copy(obs, ls.observers)
	snapshot := make([]T, len(ls.items))
	copy(snapshot, ls.items)
	ls.mu.Unlock()

	for _, fn := range obs {
		fn(snapshot)
	}
}

// Append adds items and notifies observers.
func (ls *ListState[T]) Append(items ...T) {
	ls.mu.Lock()
	ls.items = append(ls.items, items...)
	obs := make([]func([]T), len(ls.observers))
	copy(obs, ls.observers)
	snapshot := make([]T, len(ls.items))
	copy(snapshot, ls.items)
	ls.mu.Unlock()

	for _, fn := range obs {
		fn(snapshot)
	}
}

// RemoveAt removes an item at the given index.
func (ls *ListState[T]) RemoveAt(index int) {
	ls.mu.Lock()
	if index < 0 || index >= len(ls.items) {
		ls.mu.Unlock()
		return
	}
	ls.items = append(ls.items[:index], ls.items[index+1:]...)
	obs := make([]func([]T), len(ls.observers))
	copy(obs, ls.observers)
	snapshot := make([]T, len(ls.items))
	copy(snapshot, ls.items)
	ls.mu.Unlock()

	for _, fn := range obs {
		fn(snapshot)
	}
}

// Watch registers an observer on the list.
func (ls *ListState[T]) Watch(fn func([]T)) func() {
	ls.mu.Lock()
	ls.observers = append(ls.observers, fn)
	idx := len(ls.observers) - 1
	ls.mu.Unlock()

	return func() {
		ls.mu.Lock()
		defer ls.mu.Unlock()
		if idx < len(ls.observers) {
			ls.observers = append(ls.observers[:idx], ls.observers[idx+1:]...)
		}
	}
}

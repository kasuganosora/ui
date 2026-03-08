package ui

import "sync"

// observer pairs a unique ID with a callback.
type observer[T any] struct {
	id uint64
	fn func(T)
}

// State is a reactive value container. When the value changes,
// all registered observers are notified.
type State[T comparable] struct {
	mu        sync.RWMutex
	value     T
	observers []observer[T]
	nextID    uint64
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
	obs := make([]observer[T], len(s.observers))
	copy(obs, s.observers)
	s.mu.Unlock()

	for _, o := range obs {
		o.fn(v)
	}
}

// Watch registers an observer that is called when the value changes.
// Returns an unsubscribe function.
func (s *State[T]) Watch(fn func(T)) func() {
	s.mu.Lock()
	id := s.nextID
	s.nextID++
	s.observers = append(s.observers, observer[T]{id: id, fn: fn})
	s.mu.Unlock()

	return func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		for i, o := range s.observers {
			if o.id == id {
				s.observers = append(s.observers[:i], s.observers[i+1:]...)
				return
			}
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
	observers []observer[[]T]
	nextID    uint64
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
	obs := make([]observer[[]T], len(ls.observers))
	copy(obs, ls.observers)
	snapshot := make([]T, len(ls.items))
	copy(snapshot, ls.items)
	ls.mu.Unlock()

	for _, o := range obs {
		o.fn(snapshot)
	}
}

// Append adds items and notifies observers.
func (ls *ListState[T]) Append(items ...T) {
	ls.mu.Lock()
	ls.items = append(ls.items, items...)
	obs := make([]observer[[]T], len(ls.observers))
	copy(obs, ls.observers)
	snapshot := make([]T, len(ls.items))
	copy(snapshot, ls.items)
	ls.mu.Unlock()

	for _, o := range obs {
		o.fn(snapshot)
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
	obs := make([]observer[[]T], len(ls.observers))
	copy(obs, ls.observers)
	snapshot := make([]T, len(ls.items))
	copy(snapshot, ls.items)
	ls.mu.Unlock()

	for _, o := range obs {
		o.fn(snapshot)
	}
}

// Watch registers an observer on the list.
func (ls *ListState[T]) Watch(fn func([]T)) func() {
	ls.mu.Lock()
	id := ls.nextID
	ls.nextID++
	ls.observers = append(ls.observers, observer[[]T]{id: id, fn: fn})
	ls.mu.Unlock()

	return func() {
		ls.mu.Lock()
		defer ls.mu.Unlock()
		for i, o := range ls.observers {
			if o.id == id {
				ls.observers = append(ls.observers[:i], ls.observers[i+1:]...)
				return
			}
		}
	}
}

package render

// ElementID is a local type matching core.ElementID to avoid import cycles.
// Callers should pass core.ElementID values directly (both are uint64).
type ElementID = uint64

// CommandCache stores snapshots of render commands per element.
// When an element hasn't changed (not dirty), cached commands can be replayed
// instead of calling the widget's Draw method.
type CommandCache struct {
	entries map[ElementID]*cacheEntry
}

type cacheEntry struct {
	commands []Command
	valid    bool
}

// NewCommandCache creates an empty command cache.
func NewCommandCache() *CommandCache {
	return &CommandCache{
		entries: make(map[ElementID]*cacheEntry),
	}
}

// Get returns cached commands for an element, or nil if not cached/invalid.
func (cc *CommandCache) Get(id ElementID) []Command {
	if cc == nil {
		return nil
	}
	entry := cc.entries[id]
	if entry == nil || !entry.valid {
		return nil
	}
	return entry.commands
}

// Store saves a snapshot of commands for an element.
func (cc *CommandCache) Store(id ElementID, commands []Command) {
	if cc == nil {
		return
	}
	entry := cc.entries[id]
	if entry == nil {
		entry = &cacheEntry{}
		cc.entries[id] = entry
	}
	// Copy commands to avoid referencing the original buffer
	if cap(entry.commands) >= len(commands) {
		entry.commands = entry.commands[:len(commands)]
	} else {
		entry.commands = make([]Command, len(commands))
	}
	copy(entry.commands, commands)
	entry.valid = true
}

// Invalidate marks an element's cache as stale.
func (cc *CommandCache) Invalidate(id ElementID) {
	if cc == nil {
		return
	}
	if entry := cc.entries[id]; entry != nil {
		entry.valid = false
	}
}

// InvalidateAll marks all entries as stale.
func (cc *CommandCache) InvalidateAll() {
	if cc == nil {
		return
	}
	for _, entry := range cc.entries {
		entry.valid = false
	}
}

// Remove deletes an element's cache entry entirely.
func (cc *CommandCache) Remove(id ElementID) {
	if cc == nil {
		return
	}
	delete(cc.entries, id)
}

// Len returns the number of cached entries.
func (cc *CommandCache) Len() int {
	if cc == nil {
		return 0
	}
	return len(cc.entries)
}

// ValidCount returns how many entries are currently valid.
func (cc *CommandCache) ValidCount() int {
	if cc == nil {
		return 0
	}
	n := 0
	for _, entry := range cc.entries {
		if entry.valid {
			n++
		}
	}
	return n
}

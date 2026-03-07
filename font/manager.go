package font

import (
	"fmt"
	"sync"
)

// fontEntry stores a registered font with its metadata.
type fontEntry struct {
	id     ID
	family string
	weight Weight
	style  Style
}

// fontManager is the default Manager implementation.
// It manages font registration, resolution by properties, and fallback chains.
type fontManager struct {
	mu      sync.RWMutex
	engine  Engine
	entries map[ID]*fontEntry
	nextID  ID

	// families maps family name -> list of font entries (sorted by weight).
	families map[string][]*fontEntry

	// fallbackChains maps locale -> ordered list of family names.
	fallbackChains map[string][]string

	// defaultFallback is the fallback chain when no locale match.
	defaultFallback []string
}

// NewManager creates a new font manager backed by the given engine.
func NewManager(engine Engine) Manager {
	return &fontManager{
		engine:          engine,
		entries:         make(map[ID]*fontEntry),
		families:        make(map[string][]*fontEntry),
		fallbackChains:  make(map[string][]string),
		defaultFallback: nil,
		nextID:          1,
	}
}

func (m *fontManager) Register(family string, weight Weight, style Style, data []byte) (ID, error) {
	id, err := m.engine.LoadFont(data)
	if err != nil {
		return InvalidFontID, fmt.Errorf("font: load failed: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	entry := &fontEntry{
		id:     id,
		family: family,
		weight: weight,
		style:  style,
	}
	m.entries[id] = entry
	m.families[family] = append(m.families[family], entry)
	return id, nil
}

func (m *fontManager) RegisterFile(family string, weight Weight, style Style, path string) (ID, error) {
	id, err := m.engine.LoadFontFile(path)
	if err != nil {
		return InvalidFontID, fmt.Errorf("font: load file failed: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	entry := &fontEntry{
		id:     id,
		family: family,
		weight: weight,
		style:  style,
	}
	m.entries[id] = entry
	m.families[family] = append(m.families[family], entry)
	return id, nil
}

func (m *fontManager) Resolve(props Properties) (ID, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.resolveInFamily(props.Family, props.Weight, props.Style)
}

func (m *fontManager) ResolveRune(props Properties, r rune) (ID, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Try primary family first
	if id, ok := m.resolveRuneInFamily(props.Family, props.Weight, props.Style, r); ok {
		return id, true
	}

	// Try fallback chains
	chains := m.fallbackChainsFor(props.Family)
	for _, family := range chains {
		if family == props.Family {
			continue
		}
		if id, ok := m.resolveRuneInFamily(family, props.Weight, props.Style, r); ok {
			return id, true
		}
	}

	return InvalidFontID, false
}

func (m *fontManager) SetFallbackChain(locale string, families []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if locale == "" {
		m.defaultFallback = families
	} else {
		m.fallbackChains[locale] = families
	}
}

func (m *fontManager) Engine() Engine {
	return m.engine
}

// resolveInFamily finds the best matching font within a family.
// Uses CSS font matching: exact weight+style > closest weight > any style.
func (m *fontManager) resolveInFamily(family string, weight Weight, style Style) (ID, bool) {
	entries := m.families[family]
	if len(entries) == 0 {
		return InvalidFontID, false
	}

	// Pass 1: exact match
	for _, e := range entries {
		if e.weight == weight && e.style == style {
			return e.id, true
		}
	}

	// Pass 2: same style, closest weight
	bestID := InvalidFontID
	bestDist := int(^uint(0) >> 1) // max int
	for _, e := range entries {
		if e.style == style {
			dist := weightDistance(e.weight, weight)
			if dist < bestDist {
				bestDist = dist
				bestID = e.id
			}
		}
	}
	if bestID != InvalidFontID {
		return bestID, true
	}

	// Pass 3: any style, closest weight
	for _, e := range entries {
		dist := weightDistance(e.weight, weight)
		if dist < bestDist {
			bestDist = dist
			bestID = e.id
		}
	}
	if bestID != InvalidFontID {
		return bestID, true
	}

	// Pass 4: just return first
	return entries[0].id, true
}

// resolveRuneInFamily finds a font in the family that contains the given rune.
func (m *fontManager) resolveRuneInFamily(family string, weight Weight, style Style, r rune) (ID, bool) {
	id, ok := m.resolveInFamily(family, weight, style)
	if !ok {
		return InvalidFontID, false
	}
	if m.engine.HasGlyph(id, r) {
		return id, true
	}

	// Try other faces in the same family
	entries := m.families[family]
	for _, e := range entries {
		if e.id != id && m.engine.HasGlyph(e.id, r) {
			return e.id, true
		}
	}

	return InvalidFontID, false
}

// fallbackChainsFor returns the fallback chain for a given primary family.
// It checks locale-specific chains first, then default.
func (m *fontManager) fallbackChainsFor(primary string) []string {
	// For now, return default fallback. Locale support can be added
	// when the caller passes locale through ShapeOptions.
	if m.defaultFallback != nil {
		return m.defaultFallback
	}
	return nil
}

func weightDistance(a, b Weight) int {
	d := int(a) - int(b)
	if d < 0 {
		return -d
	}
	return d
}

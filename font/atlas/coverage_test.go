package atlas

import (
	"testing"

	"github.com/kasuganosora/ui/font"
)

// Additional tests to bring atlas coverage to 80%+.

func TestAtlasTexture(t *testing.T) {
	a := New(Options{Width: 64, Height: 64})
	defer a.Destroy()
	// No backend, texture should be invalid (zero)
	if a.Texture() != 0 {
		t.Errorf("expected zero texture handle, got %v", a.Texture())
	}
}

func TestAtlasOccupancy(t *testing.T) {
	a := New(Options{Width: 256, Height: 256})
	if a.Occupancy() != 0 {
		t.Errorf("expected 0, got %v", a.Occupancy())
	}
	bitmap := font.GlyphBitmap{Width: 10, Height: 10, Data: make([]byte, 100)}
	a.Insert(MakeKey(1, 1, 16), bitmap, font.GlyphMetrics{})
	occ := a.Occupancy()
	if occ <= 0 {
		t.Errorf("expected positive occupancy, got %v", occ)
	}
}

func TestAtlasUploadNoDirty(t *testing.T) {
	a := New(Options{Width: 64, Height: 64})
	// No inserts, so nothing dirty
	err := a.Upload()
	if err != nil {
		t.Errorf("Upload with no dirty data should not error: %v", err)
	}
}

func TestAtlasEvictionTriggered(t *testing.T) {
	// Very small atlas: can only fit one 10x10 glyph per shelf, maybe 2 shelves
	a := New(Options{Width: 16, Height: 16})
	bitmap := font.GlyphBitmap{Width: 10, Height: 10, Data: make([]byte, 100)}
	metrics := font.GlyphMetrics{Advance: 10}

	// Insert first glyph — should succeed
	e1 := a.Insert(MakeKey(1, 1, 16), bitmap, metrics)
	if e1 == nil {
		t.Fatal("first insert should succeed")
	}
	if a.GlyphCount() != 1 {
		t.Errorf("expected 1 glyph, got %d", a.GlyphCount())
	}

	// Insert second glyph — should trigger eviction then succeed
	e2 := a.Insert(MakeKey(1, 2, 16), bitmap, metrics)
	if e2 == nil {
		t.Fatal("second insert should succeed after eviction")
	}

	// First glyph should be gone (eviction clears all)
	if a.Lookup(MakeKey(1, 1, 16)) != nil {
		t.Error("first glyph should be evicted")
	}
}

func TestAtlasInsertTooLarge(t *testing.T) {
	a := New(Options{Width: 8, Height: 8})
	// Glyph larger than atlas
	bitmap := font.GlyphBitmap{Width: 100, Height: 100, Data: make([]byte, 10000)}
	entry := a.Insert(MakeKey(1, 1, 16), bitmap, font.GlyphMetrics{})
	if entry != nil {
		t.Error("insert of oversized glyph should return nil")
	}
}

func TestAtlasDestroyWithoutBackend(t *testing.T) {
	a := New(Options{Width: 64, Height: 64})
	// Should not panic
	a.Destroy()
}

func TestAtlasBlitBitmapBoundary(t *testing.T) {
	a := New(Options{Width: 32, Height: 32})
	// Insert a glyph with data shorter than width*height
	shortData := make([]byte, 5) // less than 10*10
	bitmap := font.GlyphBitmap{Width: 10, Height: 10, Data: shortData}
	entry := a.Insert(MakeKey(1, 1, 16), bitmap, font.GlyphMetrics{Advance: 10})
	// Should handle gracefully without panic
	if entry == nil {
		t.Fatal("insert should succeed even with short bitmap data")
	}
}

func TestAtlasBeginFrameLRU(t *testing.T) {
	a := New(Options{Width: 256, Height: 256})
	bitmap := font.GlyphBitmap{Width: 5, Height: 5, Data: make([]byte, 25)}

	a.Insert(MakeKey(1, 1, 16), bitmap, font.GlyphMetrics{})
	a.BeginFrame()
	a.BeginFrame()

	// Lookup should update LRU
	entry := a.Lookup(MakeKey(1, 1, 16))
	if entry == nil {
		t.Fatal("lookup should return entry")
	}
}

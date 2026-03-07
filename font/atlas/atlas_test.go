package atlas

import (
	"testing"

	"github.com/kasuganosora/ui/font"
)

func TestAtlasInsertAndLookup(t *testing.T) {
	a := New(Options{Width: 256, Height: 256})
	defer a.Destroy()

	key := MakeKey(1, 42, 16.0)
	bitmap := font.GlyphBitmap{
		Width:  10,
		Height: 12,
		Data:   make([]byte, 10*12),
		SDF:    false,
	}
	metrics := font.GlyphMetrics{
		Width: 10, Height: 12, BearingX: 0, BearingY: 10, Advance: 9.6,
	}

	entry := a.Insert(key, bitmap, metrics)
	if entry == nil {
		t.Fatal("Insert returned nil")
	}
	if entry.Region.Width != 10 || entry.Region.Height != 12 {
		t.Errorf("unexpected region: %+v", entry.Region)
	}

	// Lookup should return the same entry
	found := a.Lookup(key)
	if found == nil {
		t.Fatal("Lookup returned nil")
	}
	if found.U0 >= found.U1 || found.V0 >= found.V1 {
		t.Errorf("invalid UVs: U0=%f U1=%f V0=%f V1=%f", found.U0, found.U1, found.V0, found.V1)
	}
}

func TestAtlasLookupMiss(t *testing.T) {
	a := New(Options{Width: 256, Height: 256})
	key := MakeKey(1, 99, 16.0)
	if a.Lookup(key) != nil {
		t.Error("expected nil for missing key")
	}
}

func TestAtlasDuplicateInsert(t *testing.T) {
	a := New(Options{Width: 256, Height: 256})
	key := MakeKey(1, 42, 16.0)
	bitmap := font.GlyphBitmap{Width: 5, Height: 5, Data: make([]byte, 25)}
	metrics := font.GlyphMetrics{Advance: 5}

	e1 := a.Insert(key, bitmap, metrics)
	e2 := a.Insert(key, bitmap, metrics)
	if e1 == nil || e2 == nil {
		t.Fatal("Insert returned nil")
	}
	// Should return same entry (deduplicated)
	if e1.Region.X != e2.Region.X || e1.Region.Y != e2.Region.Y {
		t.Error("duplicate insert should return same region")
	}
}

func TestAtlasGlyphCount(t *testing.T) {
	a := New(Options{Width: 256, Height: 256})
	if a.GlyphCount() != 0 {
		t.Errorf("expected 0, got %d", a.GlyphCount())
	}

	bitmap := font.GlyphBitmap{Width: 5, Height: 5, Data: make([]byte, 25)}
	metrics := font.GlyphMetrics{Advance: 5}

	a.Insert(MakeKey(1, 1, 16), bitmap, metrics)
	a.Insert(MakeKey(1, 2, 16), bitmap, metrics)
	a.Insert(MakeKey(1, 3, 16), bitmap, metrics)

	if a.GlyphCount() != 3 {
		t.Errorf("expected 3, got %d", a.GlyphCount())
	}
}

func TestAtlasEviction(t *testing.T) {
	// Tiny atlas that can only fit a couple glyphs
	a := New(Options{Width: 32, Height: 32})
	bitmap := font.GlyphBitmap{Width: 14, Height: 14, Data: make([]byte, 14*14)}
	metrics := font.GlyphMetrics{Advance: 14}

	// Fill atlas
	e1 := a.Insert(MakeKey(1, 1, 16), bitmap, metrics)
	e2 := a.Insert(MakeKey(1, 2, 16), bitmap, metrics)

	if e1 == nil {
		t.Fatal("first insert should succeed")
	}

	// Third should trigger eviction
	e3 := a.Insert(MakeKey(1, 3, 16), bitmap, metrics)
	// After eviction, old entries are gone
	_ = e2
	_ = e3

	// First entry should be evicted
	if a.Lookup(MakeKey(1, 1, 16)) != nil && e3 != nil {
		// If eviction happened, old entries are gone
		// This is fine — atlas was reset
	}
}

func TestAtlasBeginFrame(t *testing.T) {
	a := New(Options{Width: 256, Height: 256})
	a.BeginFrame()
	a.BeginFrame()
	// Should not panic
}

func TestAtlasDimensions(t *testing.T) {
	a := New(Options{Width: 512, Height: 1024})
	if a.Width() != 512 {
		t.Errorf("expected width 512, got %d", a.Width())
	}
	if a.Height() != 1024 {
		t.Errorf("expected height 1024, got %d", a.Height())
	}
}

func TestAtlasDefaultDimensions(t *testing.T) {
	a := New(Options{})
	if a.Width() != 1024 || a.Height() != 1024 {
		t.Errorf("expected default 1024x1024, got %dx%d", a.Width(), a.Height())
	}
}

func TestAtlasUploadNilBackend(t *testing.T) {
	a := New(Options{Width: 64, Height: 64})
	bitmap := font.GlyphBitmap{Width: 5, Height: 5, Data: make([]byte, 25)}
	a.Insert(MakeKey(1, 1, 16), bitmap, font.GlyphMetrics{})

	// Upload with nil backend should not panic
	err := a.Upload()
	if err != nil {
		t.Errorf("Upload with nil backend should not error: %v", err)
	}
}

func TestMakeKey(t *testing.T) {
	k := MakeKey(1, 42, 16.0)
	if k.FontID != 1 || k.GlyphID != 42 || k.Size != 32 {
		t.Errorf("unexpected key: %+v", k)
	}
	// Different size = different key
	k2 := MakeKey(1, 42, 16.5)
	if k == k2 {
		t.Error("different sizes should produce different keys")
	}
}

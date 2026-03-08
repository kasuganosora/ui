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
	// Tiny atlas: 16x16 with MaxSize=16 to prevent growth.
	// 10x10 glyph + 1px pad = 11x11 per slot, only one fits.
	a := New(Options{Width: 16, Height: 16, MaxSize: 16})
	bitmap := font.GlyphBitmap{Width: 10, Height: 10, Data: make([]byte, 100)}
	metrics := font.GlyphMetrics{Advance: 10}

	// Advance past the stale threshold so eviction can work
	for i := 0; i < 15; i++ {
		a.BeginFrame()
	}

	// Fill atlas — only one 10x10 glyph fits in 16x16
	e1 := a.Insert(MakeKey(1, 1, 16), bitmap, metrics)
	if e1 == nil {
		t.Fatal("first insert should succeed")
	}

	// Advance frames so glyph becomes stale
	for i := 0; i < 15; i++ {
		a.BeginFrame()
	}

	// Second should trigger stale eviction and succeed
	e2 := a.Insert(MakeKey(1, 2, 16), bitmap, metrics)
	if e2 == nil {
		t.Fatal("insert after eviction should succeed")
	}

	// First glyph should be evicted (stale)
	if a.Lookup(MakeKey(1, 1, 16)) != nil {
		t.Error("stale glyph 1 should have been evicted")
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

func TestAtlasGrow(t *testing.T) {
	// 16x16 atlas can only fit one 10x10 glyph (10+1 pad = 11 per slot).
	// MaxSize allows growth up to 128.
	a := New(Options{Width: 16, Height: 16, MaxSize: 128})
	bitmap := font.GlyphBitmap{Width: 10, Height: 10, Data: make([]byte, 100)}
	metrics := font.GlyphMetrics{Advance: 10}

	e1 := a.Insert(MakeKey(1, 1, 16), bitmap, metrics)
	if e1 == nil {
		t.Fatal("first insert should succeed")
	}

	origW, origH := a.Width(), a.Height()

	// Second glyph won't fit — should trigger growth (no stale glyphs to evict)
	e2 := a.Insert(MakeKey(1, 2, 16), bitmap, metrics)
	if e2 == nil {
		t.Fatal("insert after growth should succeed")
	}

	// Atlas should have grown
	if a.Width() == origW && a.Height() == origH {
		t.Errorf("atlas should have grown from %dx%d", origW, origH)
	}
	if a.Width()*a.Height() <= origW*origH {
		t.Errorf("atlas area should have increased: was %d, now %d",
			origW*origH, a.Width()*a.Height())
	}

	// First glyph should still be accessible (preserved during growth)
	if a.Lookup(MakeKey(1, 1, 16)) == nil {
		t.Error("glyph 1 should be preserved after growth")
	}
}

func TestAtlasEvictStale(t *testing.T) {
	// 16x16 atlas, MaxSize=16 prevents growth so eviction must work
	a := New(Options{Width: 16, Height: 16, MaxSize: 16})
	bitmap := font.GlyphBitmap{Width: 10, Height: 10, Data: make([]byte, 100)}
	metrics := font.GlyphMetrics{Advance: 10}

	// Advance past stale threshold
	for i := 0; i < 15; i++ {
		a.BeginFrame()
	}

	// Insert one glyph (only one fits in 16x16)
	a.Insert(MakeKey(1, 1, 16), bitmap, metrics)
	if a.GlyphCount() != 1 {
		t.Fatalf("expected 1 glyph, got %d", a.GlyphCount())
	}

	// Advance frames to make it stale
	for i := 0; i < 15; i++ {
		a.BeginFrame()
	}

	// Insert a new glyph — should trigger stale eviction
	e := a.Insert(MakeKey(1, 10, 16), bitmap, metrics)
	if e == nil {
		t.Fatal("insert after stale eviction should succeed")
	}

	// Stale glyph should be gone (rebuild clears all)
	if a.Lookup(MakeKey(1, 1, 16)) != nil {
		t.Error("stale glyph 1 should have been evicted")
	}
}

func TestAtlasGrowMaxSize(t *testing.T) {
	// Atlas already at max size — growth should fail
	a := New(Options{Width: 64, Height: 64, MaxSize: 64})
	bitmap := font.GlyphBitmap{Width: 30, Height: 30, Data: make([]byte, 900)}
	metrics := font.GlyphMetrics{Advance: 30}

	// Fill atlas
	a.Insert(MakeKey(1, 1, 16), bitmap, metrics)
	a.Insert(MakeKey(1, 2, 16), bitmap, metrics)

	// Advance frames so glyphs become stale, then insert another to trigger eviction
	for i := 0; i < 15; i++ {
		a.BeginFrame()
	}

	// This should succeed via eviction (not growth)
	e := a.Insert(MakeKey(1, 3, 16), bitmap, metrics)
	if e == nil {
		t.Fatal("insert should succeed via eviction")
	}

	// Atlas should NOT have grown past max
	if a.Width() > 64 || a.Height() > 64 {
		t.Errorf("atlas should not exceed max size 64, got %dx%d", a.Width(), a.Height())
	}
}

func TestAtlasUVsAfterGrow(t *testing.T) {
	// 16x16 atlas, one 10x10 glyph fits. Growth to 32x16 on second insert.
	a := New(Options{Width: 16, Height: 16, MaxSize: 128})
	bitmap := font.GlyphBitmap{Width: 10, Height: 10, Data: make([]byte, 100)}
	metrics := font.GlyphMetrics{Advance: 10}

	e1 := a.Insert(MakeKey(1, 1, 16), bitmap, metrics)
	if e1 == nil {
		t.Fatal("first insert should succeed")
	}

	// Record pre-growth UVs
	preU1 := e1.U1
	preV1 := e1.V1
	preRegion := e1.Region

	// Second insert triggers growth
	e2 := a.Insert(MakeKey(1, 2, 16), bitmap, metrics)
	if e2 == nil {
		t.Fatal("second insert should succeed after growth")
	}

	if a.Width() == 16 && a.Height() == 16 {
		t.Fatal("atlas should have grown")
	}

	// e1 should still exist with updated UVs
	found := a.Lookup(MakeKey(1, 1, 16))
	if found == nil {
		t.Fatal("glyph 1 should be preserved after growth")
	}

	// Region pixel position should not change
	if found.Region.X != preRegion.X || found.Region.Y != preRegion.Y {
		t.Error("region position should not change after growth")
	}

	// UV should have changed since atlas is bigger (width doubled)
	if found.U1 == preU1 && a.Width() > 16 {
		t.Errorf("U1 should have been updated after growth: still %f", found.U1)
	}
	_ = preV1

	// UVs should be valid
	if found.U0 >= found.U1 || found.V0 >= found.V1 {
		t.Errorf("invalid UVs after growth: U0=%f U1=%f V0=%f V1=%f",
			found.U0, found.U1, found.V0, found.V1)
	}
}

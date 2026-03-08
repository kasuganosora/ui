package icon

import (
	"testing"

	"github.com/kasuganosora/ui/render"
)

func TestRasterizeSimplePath(t *testing.T) {
	// Simple square: M2 2 L22 2 L22 22 L2 22 Z (24x24 viewBox)
	pixels := rasterizeIcon("M2 2L22 2L22 22L2 22Z", 24)
	if pixels == nil {
		t.Fatal("rasterizeIcon returned nil")
	}
	if len(pixels) != 24*24*4 {
		t.Fatalf("expected %d bytes, got %d", 24*24*4, len(pixels))
	}

	// Check center pixel is filled (white with alpha)
	idx := (12*24 + 12) * 4
	if pixels[idx+3] == 0 {
		t.Error("center pixel should be filled")
	}

	// Check corner pixel (0,0) is empty
	if pixels[3] != 0 {
		t.Error("top-left corner should be empty")
	}
}

func TestRasterizeCirclePath(t *testing.T) {
	// Circle approximation using MDI-style path
	pixels := rasterizeIcon("M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2z", 48)
	if pixels == nil {
		t.Fatal("rasterizeIcon returned nil")
	}
	if len(pixels) != 48*48*4 {
		t.Fatalf("expected %d bytes, got %d", 48*48*4, len(pixels))
	}

	// Center should be filled
	idx := (24*48 + 24) * 4
	if pixels[idx+3] == 0 {
		t.Error("center pixel should be filled")
	}
}

func TestRasterizeEmptyPath(t *testing.T) {
	pixels := rasterizeIcon("", 24)
	if pixels != nil {
		t.Error("empty path should return nil")
	}
}

func TestBuildEdgesMoveTo(t *testing.T) {
	path := render.ParseSVGPath("M5 5L15 5L15 15L5 15Z")
	edges := buildEdges(path.Commands, 1.0)
	if len(edges) == 0 {
		t.Error("expected edges from closed path")
	}

	// Should have 2 vertical edges (horizontal edges are skipped)
	vertCount := 0
	for _, e := range edges {
		if e.y0 != e.y1 {
			vertCount++
		}
	}
	if vertCount < 2 {
		t.Errorf("expected at least 2 non-horizontal edges, got %d", vertCount)
	}
}

func TestRegistryBasic(t *testing.T) {
	reg := &Registry{
		icons: make(map[string]*Icon),
		cache: make(map[cacheKey]*CachedIcon),
	}

	reg.Register("test", "M2 2L22 2L22 22L2 22Z")

	if !reg.Has("test") {
		t.Error("should have 'test' icon")
	}
	if reg.Has("missing") {
		t.Error("should not have 'missing' icon")
	}
	if reg.Count() != 1 {
		t.Errorf("expected count 1, got %d", reg.Count())
	}
}

func TestRegistryRegisterAll(t *testing.T) {
	reg := &Registry{
		icons: make(map[string]*Icon),
		cache: make(map[cacheKey]*CachedIcon),
	}

	icons := map[string]string{
		"home":   "M10 20v-6h4v6h5v-8h3L12 3 2 12h3v8z",
		"search": "M15.5 14h-.79l-.28-.27z",
	}
	reg.RegisterAll(icons)

	if reg.Count() != 2 {
		t.Errorf("expected 2, got %d", reg.Count())
	}
	names := reg.Names()
	if len(names) != 2 {
		t.Errorf("expected 2 names, got %d", len(names))
	}
}

func TestSortFloat64s(t *testing.T) {
	a := []float64{5, 2, 8, 1, 3}
	sortFloat64s(a)
	for i := 1; i < len(a); i++ {
		if a[i] < a[i-1] {
			t.Errorf("not sorted: %v", a)
			break
		}
	}
}

func TestScanlineFillAA(t *testing.T) {
	// Test that AA produces intermediate alpha values
	pixels := rasterizeIcon("M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2z", 48)
	if pixels == nil {
		t.Fatal("nil pixels")
	}

	// Check for intermediate alpha values (AA should produce non-0, non-255)
	hasIntermediate := false
	for i := 3; i < len(pixels); i += 4 {
		a := pixels[i]
		if a > 0 && a < 255 {
			hasIntermediate = true
			break
		}
	}
	if !hasIntermediate {
		t.Error("expected antialiased edges with intermediate alpha values")
	}
}

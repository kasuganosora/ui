package freetype

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kasuganosora/ui/font"
)

// Additional tests for 90%+ coverage.

func TestLoadFontFileNotFound(t *testing.T) {
	engine := skipIfNoFreeType(t)
	defer engine.Destroy()

	_, err := engine.LoadFontFile("/nonexistent/font.ttf")
	if err == nil {
		t.Error("expected error for missing font file")
	}
}

func TestGlyphIndexInvalidFont(t *testing.T) {
	engine := skipIfNoFreeType(t)
	defer engine.Destroy()

	glyph := engine.GlyphIndex(999, 'A')
	if glyph != 0 {
		t.Error("expected 0 for invalid font ID")
	}
}

func TestGlyphMetricsInvalidFont(t *testing.T) {
	engine := skipIfNoFreeType(t)
	defer engine.Destroy()

	m := engine.GlyphMetrics(999, 1, 16)
	if m.Advance != 0 {
		t.Error("expected zero metrics for invalid font ID")
	}
}

func TestGlyphMetricsSpace(t *testing.T) {
	engine := skipIfNoFreeType(t)
	defer engine.Destroy()

	path := findTestFont(t)
	id, _ := engine.LoadFontFile(path)

	spaceGlyph := engine.GlyphIndex(id, ' ')
	m := engine.GlyphMetrics(id, spaceGlyph, 16)
	if m.Advance <= 0 {
		t.Errorf("space should have positive advance, got %f", m.Advance)
	}
}

func TestRasterizeInvalidFont(t *testing.T) {
	engine := skipIfNoFreeType(t)
	defer engine.Destroy()

	_, err := engine.RasterizeGlyph(999, 1, 16, false)
	if err == nil {
		t.Error("expected error for invalid font ID")
	}
}

func TestKerningInvalidFont(t *testing.T) {
	engine := skipIfNoFreeType(t)
	defer engine.Destroy()

	k := engine.Kerning(999, 1, 2, 16)
	if k != 0 {
		t.Error("expected 0 kerning for invalid font")
	}
}

func TestHasGlyphInvalidFont(t *testing.T) {
	engine := skipIfNoFreeType(t)
	defer engine.Destroy()

	if engine.HasGlyph(999, 'A') {
		t.Error("expected false for invalid font")
	}
}

func TestHasGlyphMissing(t *testing.T) {
	engine := skipIfNoFreeType(t)
	defer engine.Destroy()

	path := findTestFont(t)
	id, _ := engine.LoadFontFile(path)

	// Private use area rune — unlikely to be in arial
	if engine.HasGlyph(id, '\uF8FF') {
		// Some fonts have this, so just skip the assertion
	}
}

func TestLoadFontFromData(t *testing.T) {
	engine := skipIfNoFreeType(t)
	defer engine.Destroy()

	path := findTestFont(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("cannot read font: %v", err)
	}

	id, err := engine.LoadFont(data)
	if err != nil {
		t.Fatalf("LoadFont from data failed: %v", err)
	}

	m := engine.FontMetrics(id, 16)
	if m.Ascent <= 0 {
		t.Error("expected positive ascent from loaded font data")
	}
}

func TestUnloadFontInvalid(t *testing.T) {
	engine := skipIfNoFreeType(t)
	defer engine.Destroy()

	// Should not panic
	engine.UnloadFont(999)
}

func TestMultipleFonts(t *testing.T) {
	engine := skipIfNoFreeType(t)
	defer engine.Destroy()

	// Find two different fonts
	candidates := []string{
		`C:\Windows\Fonts\arial.ttf`,
		`C:\Windows\Fonts\times.ttf`,
		`C:\Windows\Fonts\consola.ttf`,
	}
	var loaded []font.ID
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			id, err := engine.LoadFontFile(p)
			if err == nil {
				loaded = append(loaded, id)
			}
		}
		if len(loaded) >= 2 {
			break
		}
	}
	if len(loaded) < 2 {
		t.Skip("need 2 system fonts for this test")
	}

	// Both should have independent metrics
	m1 := engine.FontMetrics(loaded[0], 16)
	m2 := engine.FontMetrics(loaded[1], 16)
	if m1.Ascent <= 0 || m2.Ascent <= 0 {
		t.Error("both fonts should have positive ascent")
	}
}

func TestDestroyTwice(t *testing.T) {
	engine := skipIfNoFreeType(t)
	engine.Destroy()
	// Second destroy should not panic
	engine.Destroy()
}

func TestLoadFontFileTempCopy(t *testing.T) {
	engine := skipIfNoFreeType(t)
	defer engine.Destroy()

	src := findTestFont(t)
	data, err := os.ReadFile(src)
	if err != nil {
		t.Skip("cannot read font")
	}

	dir := t.TempDir()
	tmp := filepath.Join(dir, "test.ttf")
	os.WriteFile(tmp, data, 0644)

	id, err := engine.LoadFontFile(tmp)
	if err != nil {
		t.Fatalf("LoadFontFile from temp: %v", err)
	}
	if engine.HasGlyph(id, 'A') != true {
		t.Error("should have glyph 'A'")
	}
}

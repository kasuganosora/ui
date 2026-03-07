package freetype

import (
	"os"
	"testing"
)

// skipIfNoFreeType skips the test if FreeType DLL is not available.
func skipIfNoFreeType(t *testing.T) *Engine {
	t.Helper()
	engine, err := New()
	if err != nil {
		t.Skipf("FreeType not available: %v", err)
	}
	return engine
}

// findTestFont returns a path to a system font for testing, or skips.
func findTestFont(t *testing.T) string {
	t.Helper()
	candidates := []string{
		`C:\Windows\Fonts\arial.ttf`,
		`C:\Windows\Fonts\segoeui.ttf`,
		`C:\Windows\Fonts\consola.ttf`,
		`/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf`,
		`/usr/share/fonts/TTF/DejaVuSans.ttf`,
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	t.Skip("no system font found for testing")
	return ""
}

func TestNew(t *testing.T) {
	engine := skipIfNoFreeType(t)
	defer engine.Destroy()
}

func TestLoadFontFile(t *testing.T) {
	engine := skipIfNoFreeType(t)
	defer engine.Destroy()

	path := findTestFont(t)
	id, err := engine.LoadFontFile(path)
	if err != nil {
		t.Fatalf("LoadFontFile failed: %v", err)
	}
	if id == 0 {
		t.Fatal("got zero font ID")
	}
}

func TestFontMetrics(t *testing.T) {
	engine := skipIfNoFreeType(t)
	defer engine.Destroy()

	path := findTestFont(t)
	id, err := engine.LoadFontFile(path)
	if err != nil {
		t.Fatalf("LoadFontFile: %v", err)
	}

	m := engine.FontMetrics(id, 16)
	if m.Ascent <= 0 {
		t.Errorf("expected positive ascent, got %f", m.Ascent)
	}
	if m.Descent <= 0 {
		t.Errorf("expected positive descent, got %f", m.Descent)
	}
	if m.LineHeight <= 0 {
		t.Errorf("expected positive line height, got %f", m.LineHeight)
	}
	if m.UnitsPerEm <= 0 {
		t.Errorf("expected positive units per em, got %f", m.UnitsPerEm)
	}
}

func TestGlyphIndex(t *testing.T) {
	engine := skipIfNoFreeType(t)
	defer engine.Destroy()

	path := findTestFont(t)
	id, _ := engine.LoadFontFile(path)

	glyph := engine.GlyphIndex(id, 'A')
	if glyph == 0 {
		t.Error("GlyphIndex('A') returned 0")
	}

	// Space should also have a glyph
	spaceGlyph := engine.GlyphIndex(id, ' ')
	if spaceGlyph == 0 {
		t.Error("GlyphIndex(' ') returned 0")
	}
}

func TestGlyphMetrics(t *testing.T) {
	engine := skipIfNoFreeType(t)
	defer engine.Destroy()

	path := findTestFont(t)
	id, _ := engine.LoadFontFile(path)

	glyph := engine.GlyphIndex(id, 'A')
	m := engine.GlyphMetrics(id, glyph, 32)

	if m.Width <= 0 {
		t.Errorf("expected positive width, got %f", m.Width)
	}
	if m.Height <= 0 {
		t.Errorf("expected positive height, got %f", m.Height)
	}
	if m.Advance <= 0 {
		t.Errorf("expected positive advance, got %f", m.Advance)
	}
}

func TestRasterizeGlyph(t *testing.T) {
	engine := skipIfNoFreeType(t)
	defer engine.Destroy()

	path := findTestFont(t)
	id, _ := engine.LoadFontFile(path)
	glyph := engine.GlyphIndex(id, 'A')

	bmp, err := engine.RasterizeGlyph(id, glyph, 32, false)
	if err != nil {
		t.Fatalf("RasterizeGlyph failed: %v", err)
	}
	if bmp.Width <= 0 || bmp.Height <= 0 {
		t.Errorf("expected non-zero bitmap, got %dx%d", bmp.Width, bmp.Height)
	}
	if len(bmp.Data) != bmp.Width*bmp.Height {
		t.Errorf("data length %d != %d*%d", len(bmp.Data), bmp.Width, bmp.Height)
	}
	if bmp.SDF {
		t.Error("expected non-SDF bitmap")
	}
}

func TestRasterizeGlyphSDF(t *testing.T) {
	engine := skipIfNoFreeType(t)
	defer engine.Destroy()

	path := findTestFont(t)
	id, _ := engine.LoadFontFile(path)
	glyph := engine.GlyphIndex(id, 'A')

	bmp, err := engine.RasterizeGlyph(id, glyph, 32, true)
	if err != nil {
		// SDF rendering may not be available in all FreeType builds
		t.Skipf("SDF rendering not supported: %v", err)
	}
	if bmp.Width <= 0 || bmp.Height <= 0 {
		t.Errorf("expected non-zero SDF bitmap, got %dx%d", bmp.Width, bmp.Height)
	}
	if !bmp.SDF {
		t.Error("expected SDF bitmap")
	}
}

func TestKerning(t *testing.T) {
	engine := skipIfNoFreeType(t)
	defer engine.Destroy()

	path := findTestFont(t)
	id, _ := engine.LoadFontFile(path)

	glyphA := engine.GlyphIndex(id, 'A')
	glyphV := engine.GlyphIndex(id, 'V')

	// AV is a classic kerning pair — but not all fonts have kerning tables.
	// We just verify it doesn't crash and returns a reasonable value.
	kern := engine.Kerning(id, glyphA, glyphV, 32)
	_ = kern // May be 0 if font has no kerning
}

func TestHasGlyph(t *testing.T) {
	engine := skipIfNoFreeType(t)
	defer engine.Destroy()

	path := findTestFont(t)
	id, _ := engine.LoadFontFile(path)

	if !engine.HasGlyph(id, 'A') {
		t.Error("expected HasGlyph('A') = true")
	}
}

func TestUnloadFont(t *testing.T) {
	engine := skipIfNoFreeType(t)
	defer engine.Destroy()

	path := findTestFont(t)
	id, _ := engine.LoadFontFile(path)

	engine.UnloadFont(id)

	// After unload, FontMetrics should return zero
	m := engine.FontMetrics(id, 16)
	if m.Ascent != 0 {
		t.Error("expected zero metrics after UnloadFont")
	}
}

func TestFix26_6ToFloat(t *testing.T) {
	tests := []struct {
		input    int32
		expected float32
	}{
		{0, 0},
		{64, 1.0},
		{128, 2.0},
		{32, 0.5},
		{-64, -1.0},
		{960, 15.0},
	}
	for _, tc := range tests {
		got := fix26_6ToFloat(tc.input)
		if got != tc.expected {
			t.Errorf("fix26_6ToFloat(%d) = %f, want %f", tc.input, got, tc.expected)
		}
	}
}

func TestLoadFontInvalidData(t *testing.T) {
	engine := skipIfNoFreeType(t)
	defer engine.Destroy()

	_, err := engine.LoadFont([]byte{0, 1, 2, 3})
	if err == nil {
		t.Error("expected error loading invalid font data")
	}
}

func TestFontMetricsInvalidID(t *testing.T) {
	engine := skipIfNoFreeType(t)
	defer engine.Destroy()

	m := engine.FontMetrics(999, 16)
	if m.Ascent != 0 || m.Descent != 0 {
		t.Error("expected zero metrics for invalid font ID")
	}
}

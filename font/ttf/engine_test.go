package ttf

import (
	"testing"

	"github.com/kasuganosora/ui/font"
	"golang.org/x/image/font/gofont/goregular"
)

func newTestEngine(t *testing.T) (*Engine, font.ID) {
	t.Helper()
	e := New()
	id, err := e.LoadFont(goregular.TTF)
	if err != nil {
		t.Fatalf("LoadFont: %v", err)
	}
	t.Cleanup(e.Destroy)
	return e, id
}

func coverage(b font.GlyphBitmap) float64 {
	if b.Width <= 0 || b.Height <= 0 {
		return 0
	}
	n := 0
	for _, v := range b.Data {
		if v > 0 {
			n++
		}
	}
	return float64(n) / float64(b.Width*b.Height)
}

// The regression this engine fixes: the old fallback filled every glyph with
// opaque pixels, so text appeared as solid blocks. A real outline has partial
// coverage and antialiased edges.
func TestRasterizeGlyphIsOutline(t *testing.T) {
	e, id := newTestEngine(t)
	gid := e.GlyphIndex(id, 'A')
	if gid == 0 {
		t.Fatal("GlyphIndex(A) returned 0")
	}
	bmp, err := e.RasterizeGlyph(id, gid, 24, false)
	if err != nil {
		t.Fatalf("RasterizeGlyph: %v", err)
	}
	if bmp.Width == 0 || bmp.Height == 0 {
		t.Fatalf("empty bitmap for A: %dx%d", bmp.Width, bmp.Height)
	}
	if len(bmp.Data) != bmp.Width*bmp.Height {
		t.Errorf("bitmap data length = %d, want %d", len(bmp.Data), bmp.Width*bmp.Height)
	}
	if c := coverage(bmp); c < 0.05 || c > 0.85 {
		t.Errorf("coverage %.3f looks like a solid block or a blank glyph", c)
	}
	edges := 0
	for _, v := range bmp.Data {
		if v > 0 && v < 255 {
			edges++
		}
	}
	if edges == 0 {
		t.Error("no antialiased edge pixels: glyph was not rasterized from an outline")
	}
}

func TestSpaceProducesEmptyBitmap(t *testing.T) {
	e, id := newTestEngine(t)
	gid := e.GlyphIndex(id, ' ')
	bmp, err := e.RasterizeGlyph(id, gid, 24, false)
	if err != nil {
		t.Fatalf("RasterizeGlyph: %v", err)
	}
	if bmp.Width != 0 || bmp.Height != 0 || len(bmp.Data) != 0 {
		t.Errorf("space bitmap = %dx%d (%d bytes), want empty", bmp.Width, bmp.Height, len(bmp.Data))
	}
}

func TestMetricsAreLogicalPixels(t *testing.T) {
	e, id := newTestEngine(t)
	gid := e.GlyphIndex(id, 'A')

	fm := e.FontMetrics(id, 24)
	if fm.Ascent <= 0 || fm.Descent <= 0 || fm.LineHeight <= 0 || fm.UnitsPerEm <= 0 {
		t.Errorf("invalid font metrics: %+v", fm)
	}
	gm1 := e.GlyphMetrics(id, gid, 24)
	if gm1.Advance <= 0 || gm1.Width <= 0 || gm1.Height <= 0 {
		t.Errorf("invalid glyph metrics: %+v", gm1)
	}

	e.SetDPIScale(2)
	gm2 := e.GlyphMetrics(id, gid, 24)
	// Physical pixels double, but metrics are divided by dpiScale, so the
	// logical advance must stay roughly the same.
	if gm2.Advance < gm1.Advance*0.8 || gm2.Advance > gm1.Advance*1.2 {
		t.Errorf("advance changed with DPI: %.2f -> %.2f", gm1.Advance, gm2.Advance)
	}
	bmp2, err := e.RasterizeGlyph(id, gid, 24, false)
	if err != nil {
		t.Fatalf("RasterizeGlyph at 2x DPI: %v", err)
	}
	if bmp2.Width <= 0 {
		t.Error("empty bitmap at 2x DPI")
	}
}

func TestGlyphLookupAndKerning(t *testing.T) {
	e, id := newTestEngine(t)
	if !e.HasGlyph(id, 'A') {
		t.Error("HasGlyph(A) = false")
	}
	if e.HasGlyph(id, '中') {
		t.Error("HasGlyph(中) = true for an ASCII-only font")
	}
	if e.HasColorGlyphs(id) {
		t.Error("outline engine must not claim color glyphs")
	}
	if got := e.Kerning(id, e.GlyphIndex(id, 'A'), e.GlyphIndex(id, 'V'), 24); got > 100 {
		t.Errorf("implausible kerning: %v", got)
	}
}

func TestLoadInvalidFont(t *testing.T) {
	if _, err := New().LoadFont([]byte("not a font")); err == nil {
		t.Error("LoadFont accepted garbage data")
	}
}

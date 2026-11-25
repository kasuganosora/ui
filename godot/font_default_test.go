package godot

import (
	"testing"

	"github.com/kasuganosora/ui/font"
)

// TestHeadlessUsesRealGlyphOutlines guards against regressing to the
// placeholder engine, whose RasterizeGlyph returns a fully opaque block for
// every rune (all labels render as solid rectangles).
func TestHeadlessUsesRealGlyphOutlines(t *testing.T) {
	if systemFontPath() == "" {
		t.Skip("host has no system font; placeholder engine is expected")
	}
	ui, err := NewUI(UIOptions{Width: 320, Height: 240})
	if err != nil {
		t.Fatal(err)
	}
	defer ui.Destroy()
	if _, ok := ui.fontEngine.(*mockEngine); ok {
		t.Fatal("headless UI fell back to the placeholder engine despite a usable system font")
	}

	// A real outline has both covered and transparent pixels. The placeholder
	// engine produces 100% coverage, which is exactly the visual defect.
	id, err := ui.fontEngine.LoadFontFile(systemFontPath())
	if err != nil {
		t.Fatalf("load system font: %v", err)
	}
	glyph := ui.fontEngine.GlyphIndex(id, 'A')
	if glyph == 0 {
		t.Fatal("system font has no glyph for 'A'")
	}
	bm, err := ui.fontEngine.RasterizeGlyph(id, glyph, 24, false)
	if err != nil {
		t.Fatalf("rasterize: %v", err)
	}
	if len(bm.Data) == 0 {
		t.Fatal("empty glyph bitmap")
	}
	transparent, covered := 0, 0
	for _, a := range bm.Data {
		if a == 0 {
			transparent++
		} else {
			covered++
		}
	}
	if transparent == 0 || covered == 0 {
		t.Fatalf("glyph looks like a solid block: transparent=%d covered=%d", transparent, covered)
	}
}

// TestHeadlessFallsBackForMissingFont keeps the headless UI usable on minimal
// runtime images where no font can be found, instead of failing to render.
func TestHeadlessFallsBackForMissingFont(t *testing.T) {
	ui, err := NewUI(UIOptions{
		Width: 320, Height: 240,
		Font: "/nonexistent-path/ui-test-font.ttf",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer ui.Destroy()
	if _, ok := ui.fontEngine.(*mockEngine); !ok {
		t.Fatalf("expected placeholder engine for a missing font, got %T", ui.fontEngine)
	}
	if ui.fontID == font.InvalidFontID {
		t.Fatal("no font registered after falling back to the placeholder engine")
	}
	ui.Frame(0.016) // must not panic and must produce a frame
	if _, h := ui.FramebufferSize(); h == 0 {
		t.Fatal("framebuffer has zero height")
	}
}

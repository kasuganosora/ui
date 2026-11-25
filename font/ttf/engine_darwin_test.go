//go:build darwin

package ttf

import (
	"os"
	"testing"
)

// The real-world case that produced solid blocks: macOS ships its CJK font as
// a large TrueType collection, which the pure Go engine must parse.
func TestSystemTTCChineseGlyph(t *testing.T) {
	const path = "/System/Library/Fonts/PingFang.ttc"
	if _, err := os.Stat(path); err != nil {
		t.Skipf("system CJK font not available: %v", err)
	}

	e := New()
	defer e.Destroy()
	id, err := e.LoadFontFile(path)
	if err != nil {
		t.Fatalf("LoadFontFile: %v", err)
	}

	gid := e.GlyphIndex(id, '中')
	if gid == 0 {
		t.Fatal("GlyphIndex(中) returned 0")
	}
	bmp, err := e.RasterizeGlyph(id, gid, 24, false)
	if err != nil {
		t.Fatalf("RasterizeGlyph: %v", err)
	}
	if bmp.Width == 0 || bmp.Height == 0 {
		t.Fatalf("empty bitmap for 中: %dx%d", bmp.Width, bmp.Height)
	}
	if c := coverage(bmp); c < 0.05 || c > 0.85 {
		t.Errorf("coverage %.3f for 中 looks like a solid block", c)
	}
	if !e.HasGlyph(id, '按') {
		t.Error("HasGlyph(按) = false")
	}
}

// TestSystemTTCUsesWidestFace guards the collection face choice. Face 0 of
// the system CJK collection is a regional variant whose cmap omits some
// simplified characters (e.g. 稳), so blindly taking face 0 drops glyphs from
// otherwise ordinary UI text.
func TestSystemTTCUsesWidestFace(t *testing.T) {
	const path = "/System/Library/Fonts/PingFang.ttc"
	if _, err := os.Stat(path); err != nil {
		t.Skipf("system CJK font not available: %v", err)
	}

	e := New()
	defer e.Destroy()
	id, err := e.LoadFontFile(path)
	if err != nil {
		t.Fatalf("LoadFontFile: %v", err)
	}

	for _, r := range "稳发这为产图书说话时实现样应该简体中文测试" {
		gid := e.GlyphIndex(id, r)
		if gid == 0 {
			t.Errorf("GlyphIndex(%c) = 0: selected face has an incomplete cmap", r)
			continue
		}
		m := e.GlyphMetrics(id, gid, 14)
		if m.Advance <= 0 {
			t.Errorf("GlyphMetrics(%c).Advance = %.2f", r, m.Advance)
		}
		bmp, err := e.RasterizeGlyph(id, gid, 14, false)
		if err != nil || bmp.Width == 0 || bmp.Height == 0 {
			t.Errorf("RasterizeGlyph(%c): %dx%d err=%v", r, bmp.Width, bmp.Height, err)
		}
	}
}

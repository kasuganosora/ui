//go:build darwin

package ttf

import (
	"os"
	"strings"
	"testing"

	"github.com/kasuganosora/ui/font"
	"golang.org/x/image/font/sfnt"
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

// rasterInk sums the alpha coverage of a glyph bitmap. At 12px a regular CJK
// outline barely reaches full coverage, which is why small captions read as
// grey mush; a medium outline fixes that without changing the theme color.
func rasterInk(t *testing.T, e *Engine, id font.ID, size float32, r rune) float64 {
	t.Helper()
	gid := e.GlyphIndex(id, r)
	if gid == 0 {
		t.Fatalf("glyph %q missing from face %d", r, id)
	}
	bmp, err := e.RasterizeGlyph(id, gid, size, false)
	if err != nil {
		t.Fatalf("RasterizeGlyph(%q): %v", r, err)
	}
	sum := 0.0
	for _, v := range bmp.Data {
		sum += float64(v)
	}
	return sum / 255.0
}

// TestWeightedFaceKeepsCoverageAndAddsInk verifies the §15 caption weight on
// the real system collection rather than on a synthetic sample: the medium
// face must be genuinely heavier at caption sizes AND must not reintroduce
// the missing-simplified-glyph bug that the coverage-based selection fixed.
func TestWeightedFaceKeepsCoverageAndAddsInk(t *testing.T) {
	const path = "/System/Library/Fonts/PingFang.ttc"
	if _, err := os.Stat(path); err != nil {
		t.Skipf("system CJK font not available: %v", err)
	}

	e := New()
	defer e.Destroy()
	regular, err := e.LoadFontFile(path)
	if err != nil {
		t.Fatalf("regular LoadFontFile: %v", err)
	}
	medium, err := e.LoadFontFileWithWeight(path, font.WeightMedium)
	if err != nil {
		t.Fatalf("medium LoadFontFileWithWeight: %v", err)
	}
	if regular == medium {
		t.Fatal("weighted load returned the regular face ID")
	}

	// The selected face really is Medium, not just a different regional
	// variant of Regular.
	f := e.faces[medium]
	if f == nil {
		t.Fatal("medium face not stored in the engine")
	}
	sub, err := f.f.Name(f.buf, sfnt.NameIDSubfamily)
	if err != nil || !strings.EqualFold(sub, "Medium") {
		t.Fatalf("selected subfamily = %q (err=%v), want Medium", sub, err)
	}

	for _, r := range []rune{'稳', '定', '分', '类', 'A'} {
		base := rasterInk(t, e, regular, 12, r)
		heavy := rasterInk(t, e, medium, 12, r)
		gain := heavy/base - 1
		t.Logf("12px %q ink regular=%.2f medium=%.2f gain=%+.1f%%", r, base, heavy, gain*100)
		if gain < 0.15 {
			t.Errorf("medium face did not add ink for %q: gain %.1f%%", r, gain*100)
		}
	}

	missing := 0
	for _, r := range "稳定开发中分类标题按钮链接输入选择器列表卡片设置搜索消息通知时间日期" {
		if e.GlyphIndex(medium, r) == 0 {
			missing++
		}
	}
	if missing != 0 {
		t.Errorf("medium face lost %d simplified UI characters", missing)
	}

	// A weight the collection does not ship must fall back to the existing
	// coverage-based default instead of failing registration.
	fallback, err := e.LoadFontFileWithWeight(path, font.WeightBlack)
	if err != nil {
		t.Fatalf("unavailable weight should degrade, not fail: %v", err)
	}
	if ink := rasterInk(t, e, fallback, 12, '稳'); ink != rasterInk(t, e, regular, 12, '稳') {
		t.Errorf("unknown weight did not fall back to the default face: ink %.2f", ink)
	}
}

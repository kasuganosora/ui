package capture

import (
	"image"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"testing"

	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/render"
)

// mockBackend implements render.Backend with a fixed ReadPixels result.
type mockBackend struct {
	img *image.RGBA
}

func (m *mockBackend) Init(platform.Window) error                                     { return nil }
func (m *mockBackend) BeginFrame()                                                    {}
func (m *mockBackend) EndFrame()                                                      {}
func (m *mockBackend) Submit(*render.CommandBuffer)                                   {}
func (m *mockBackend) Resize(int, int)                                                {}
func (m *mockBackend) CreateTexture(render.TextureDesc) (render.TextureHandle, error) { return 0, nil }
func (m *mockBackend) UpdateTexture(render.TextureHandle, uimath.Rect, []byte) error  { return nil }
func (m *mockBackend) DestroyTexture(render.TextureHandle)                            {}
func (m *mockBackend) MaxTextureSize() int                                            { return 4096 }
func (m *mockBackend) Destroy()                                                       {}
func (m *mockBackend) ReadPixels() (*image.RGBA, error)                               { return m.img, nil }

func solidImage(w, h int, c color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, c)
		}
	}
	return img
}

func TestCompareIdentical(t *testing.T) {
	a := solidImage(10, 10, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	b := solidImage(10, 10, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	result, err := Compare(a, b, 0.01)
	if err != nil {
		t.Fatal(err)
	}
	if result.MeanError != 0 {
		t.Errorf("expected 0 mean error, got %f", result.MeanError)
	}
	if result.DiffPixels != 0 {
		t.Errorf("expected 0 diff pixels, got %d", result.DiffPixels)
	}
	if result.DiffImage != nil {
		t.Error("expected nil diff image for identical images")
	}
}

func TestCompareDifferent(t *testing.T) {
	a := solidImage(10, 10, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	b := solidImage(10, 10, color.RGBA{R: 0, G: 255, B: 0, A: 255})

	result, err := Compare(a, b, 0.01)
	if err != nil {
		t.Fatal(err)
	}
	if result.MeanError == 0 {
		t.Error("expected non-zero mean error")
	}
	if result.DiffPixels != 100 {
		t.Errorf("expected 100 diff pixels, got %d", result.DiffPixels)
	}
	if result.DiffImage == nil {
		t.Error("expected non-nil diff image")
	}
}

func TestCompareSizeMismatch(t *testing.T) {
	a := solidImage(10, 10, color.RGBA{R: 255, A: 255})
	b := solidImage(20, 20, color.RGBA{R: 255, A: 255})

	_, err := Compare(a, b, 0.01)
	if err == nil {
		t.Error("expected error for size mismatch")
	}
}

func TestCompareWithThreshold(t *testing.T) {
	a := solidImage(10, 10, color.RGBA{R: 100, G: 100, B: 100, A: 255})
	b := solidImage(10, 10, color.RGBA{R: 102, G: 100, B: 100, A: 255})

	// With high threshold, should report 0 diff pixels
	result, err := Compare(a, b, 0.05)
	if err != nil {
		t.Fatal(err)
	}
	if result.DiffPixels != 0 {
		t.Errorf("expected 0 diff pixels with threshold 0.05, got %d", result.DiffPixels)
	}

	// With zero threshold, all pixels differ
	result2, err := Compare(a, b, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result2.DiffPixels != 100 {
		t.Errorf("expected 100 diff pixels with threshold 0, got %d", result2.DiffPixels)
	}
}

func TestDiff(t *testing.T) {
	a := solidImage(10, 10, color.RGBA{R: 255, A: 255})
	b := solidImage(10, 10, color.RGBA{R: 255, A: 255})

	d, err := Diff(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if d != 0 {
		t.Errorf("expected 0 diff, got %f", d)
	}
}

func TestPSNRIdentical(t *testing.T) {
	a := solidImage(10, 10, color.RGBA{R: 128, G: 64, B: 32, A: 255})

	psnr, err := PSNR(a, a)
	if err != nil {
		t.Fatal(err)
	}
	if !math.IsInf(psnr, 1) {
		t.Errorf("expected +Inf PSNR for identical images, got %f", psnr)
	}
}

func TestPSNRDifferent(t *testing.T) {
	a := solidImage(10, 10, color.RGBA{R: 0, A: 255})
	b := solidImage(10, 10, color.RGBA{R: 255, A: 255})

	psnr, err := PSNR(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if psnr <= 0 || math.IsInf(psnr, 1) {
		t.Errorf("expected finite positive PSNR, got %f", psnr)
	}
}

func TestPSNRSizeMismatch(t *testing.T) {
	a := solidImage(10, 10, color.RGBA{A: 255})
	b := solidImage(5, 5, color.RGBA{A: 255})

	_, err := PSNR(a, b)
	if err == nil {
		t.Error("expected error for size mismatch")
	}
}

func TestSavePNGAndLoadPNG(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.png")

	original := solidImage(8, 8, color.RGBA{R: 100, G: 200, B: 50, A: 255})
	if err := SavePNG(original, path); err != nil {
		t.Fatalf("SavePNG failed: %v", err)
	}

	loaded, err := LoadPNG(path)
	if err != nil {
		t.Fatalf("LoadPNG failed: %v", err)
	}

	result, err := Compare(original, loaded, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.DiffPixels != 0 {
		t.Errorf("round-trip should be lossless, got %d diff pixels", result.DiffPixels)
	}
}

func TestLoadPNGNotFound(t *testing.T) {
	_, err := LoadPNG("/nonexistent/path.png")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestSavePNGBadPath(t *testing.T) {
	img := solidImage(1, 1, color.RGBA{A: 255})
	err := SavePNG(img, "/nonexistent/dir/file.png")
	if err == nil {
		t.Error("expected error for bad path")
	}
}

func TestMatchesGoldenBootstrap(t *testing.T) {
	dir := t.TempDir()
	goldenPath := filepath.Join(dir, "golden.png")

	img := solidImage(4, 4, color.RGBA{R: 255, G: 128, B: 0, A: 255})

	// First call: golden doesn't exist, should create it
	result, err := MatchesGolden(img, goldenPath, 0.01)
	if err != nil {
		t.Fatalf("MatchesGolden bootstrap: %v", err)
	}
	if result.DiffPixels != 0 {
		t.Error("bootstrap should report 0 diff pixels")
	}

	// Verify file was created
	if _, err := os.Stat(goldenPath); os.IsNotExist(err) {
		t.Error("golden file should have been created")
	}

	// Second call: golden exists, should match
	result, err = MatchesGolden(img, goldenPath, 0.01)
	if err != nil {
		t.Fatalf("MatchesGolden match: %v", err)
	}
	if result.DiffPixels != 0 {
		t.Error("should match after bootstrap")
	}
}

func TestMatchesGoldenMismatch(t *testing.T) {
	dir := t.TempDir()
	goldenPath := filepath.Join(dir, "golden.png")

	// Bootstrap with red
	red := solidImage(4, 4, color.RGBA{R: 255, A: 255})
	MatchesGolden(red, goldenPath, 0.01)

	// Compare with green — should detect diff
	green := solidImage(4, 4, color.RGBA{G: 255, A: 255})
	result, err := MatchesGolden(green, goldenPath, 0.01)
	if err != nil {
		t.Fatalf("MatchesGolden mismatch: %v", err)
	}
	if result.DiffPixels == 0 {
		t.Error("should detect mismatch between red and green")
	}
}

func TestAbsDiff16(t *testing.T) {
	if absDiff16(100, 50) != 50 {
		t.Error("absDiff16(100, 50) should be 50")
	}
	if absDiff16(50, 100) != 50 {
		t.Error("absDiff16(50, 100) should be 50")
	}
	if absDiff16(0, 0) != 0 {
		t.Error("absDiff16(0, 0) should be 0")
	}
}

func TestMaxU8(t *testing.T) {
	if maxU8(1, 5, 3, 2) != 5 {
		t.Error("maxU8 should return 5")
	}
	if maxU8(10, 5, 3, 2) != 10 {
		t.Error("maxU8 should return 10")
	}
	if maxU8(1, 5, 3, 20) != 20 {
		t.Error("maxU8 should return 20")
	}
}

func TestScreenshotWithMockBackend(t *testing.T) {
	expected := solidImage(4, 4, color.RGBA{R: 128, G: 64, B: 32, A: 255})
	backend := &mockBackend{img: expected}

	img, err := Screenshot(backend)
	if err != nil {
		t.Fatalf("Screenshot: %v", err)
	}
	if img != expected {
		t.Error("Screenshot should return backend's image")
	}
}

func TestMustMatchGoldenWithMock(t *testing.T) {
	dir := t.TempDir()
	goldenPath := filepath.Join(dir, "golden.png")

	img := solidImage(4, 4, color.RGBA{R: 200, G: 100, B: 50, A: 255})
	backend := &mockBackend{img: img}

	// First run bootstraps golden
	MustMatchGolden(t, backend, goldenPath, 0.01)

	// Second run should match
	MustMatchGolden(t, backend, goldenPath, 0.01)
}

func TestUpdateGolden(t *testing.T) {
	dir := t.TempDir()
	goldenPath := filepath.Join(dir, "golden.png")

	img := solidImage(4, 4, color.RGBA{R: 50, G: 150, B: 250, A: 255})
	backend := &mockBackend{img: img}

	err := UpdateGolden(backend, goldenPath)
	if err != nil {
		t.Fatalf("UpdateGolden: %v", err)
	}

	if _, err := os.Stat(goldenPath); os.IsNotExist(err) {
		t.Error("golden file should exist after UpdateGolden")
	}
}

func TestAssertImageEqual(t *testing.T) {
	a := solidImage(4, 4, color.RGBA{R: 100, A: 255})
	b := solidImage(4, 4, color.RGBA{R: 100, A: 255})
	AssertImageEqual(t, a, b, 0.01)
}

func TestDiffWithDifferentImages(t *testing.T) {
	a := solidImage(4, 4, color.RGBA{R: 0, A: 255})
	b := solidImage(4, 4, color.RGBA{R: 255, A: 255})

	d, err := Diff(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if d == 0 {
		t.Error("expected non-zero diff")
	}
}

func TestDiffSizeMismatch(t *testing.T) {
	a := solidImage(4, 4, color.RGBA{A: 255})
	b := solidImage(8, 8, color.RGBA{A: 255})

	d, err := Diff(a, b)
	if err == nil {
		t.Error("expected error")
	}
	if d != 1.0 {
		t.Error("expected 1.0 on error")
	}
}

func TestLoadPNGInvalidData(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.png")
	os.WriteFile(path, []byte("not a png"), 0644)

	_, err := LoadPNG(path)
	if err == nil {
		t.Error("expected error for invalid PNG")
	}
}

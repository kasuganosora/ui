package capture

import (
	"fmt"
	"image"
	"path/filepath"
	"testing"

	"github.com/kasuganosora/ui/render"
)

// MustMatchGolden captures a screenshot and compares against a golden image.
// On first run (golden doesn't exist), it saves the screenshot as the golden.
// On subsequent runs, it fails the test if the diff exceeds the threshold.
//
// threshold is the per-channel tolerance (0..1). Typically 0.01 for exact match,
// 0.05 for minor rendering differences across GPUs.
func MustMatchGolden(t *testing.T, backend render.Backend, goldenPath string, threshold float64) {
	t.Helper()

	img, err := Screenshot(backend)
	if err != nil {
		t.Fatalf("capture: screenshot failed: %v", err)
	}

	result, err := MatchesGolden(img, goldenPath, threshold)
	if err != nil {
		t.Fatalf("capture: golden comparison failed: %v", err)
	}

	if result.DiffPixels > 0 {
		pct := float64(result.DiffPixels) / float64(result.TotalPixels) * 100
		t.Errorf("capture: %d/%d pixels differ (%.2f%%), mean error=%.4f, max error=%.4f",
			result.DiffPixels, result.TotalPixels, pct, result.MeanError, result.MaxError)

		// Save actual and diff images for debugging
		if result.DiffImage != nil {
			diffPath := goldenPath[:len(goldenPath)-len(filepath.Ext(goldenPath))] + "_diff.png"
			SavePNG(result.DiffImage, diffPath)
			t.Logf("capture: diff image saved to %s", diffPath)
		}
		actualPath := goldenPath[:len(goldenPath)-len(filepath.Ext(goldenPath))] + "_actual.png"
		SavePNG(img, actualPath)
		t.Logf("capture: actual image saved to %s", actualPath)
	}
}

// AssertImageEqual fails the test if two images differ by more than threshold.
func AssertImageEqual(t *testing.T, expected, actual *image.RGBA, threshold float64) {
	t.Helper()

	result, err := Compare(expected, actual, threshold)
	if err != nil {
		t.Fatalf("capture: compare failed: %v", err)
	}

	if result.DiffPixels > 0 {
		pct := float64(result.DiffPixels) / float64(result.TotalPixels) * 100
		t.Errorf("capture: images differ — %d/%d pixels (%.2f%%), mean=%.4f",
			result.DiffPixels, result.TotalPixels, pct, result.MeanError)
	}
}

// UpdateGolden saves the current screenshot as the new golden image.
// Useful for intentionally updating golden images after visual changes.
func UpdateGolden(backend render.Backend, goldenPath string) error {
	img, err := Screenshot(backend)
	if err != nil {
		return fmt.Errorf("capture: screenshot failed: %w", err)
	}
	return SavePNG(img, goldenPath)
}

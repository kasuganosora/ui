// Package capture provides screenshot and visual regression testing utilities.
//
// Usage for acceptance testing:
//
//	img, err := capture.Screenshot(backend)
//	capture.SavePNG(img, "testdata/actual.png")
//	capture.MustMatchGolden(t, img, "testdata/golden.png", 0.01)
package capture

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Screenshot captures the current framebuffer from the backend as an RGBA image.
func Screenshot(backend render.Backend) (*image.RGBA, error) {
	return backend.ReadPixels()
}

// ScreenshotNode captures a specific element and all its descendants.
// It takes a full framebuffer screenshot and crops to the element's layout bounds.
// The dpiScale parameter converts logical layout coordinates to physical pixels
// (pass 1.0 if layout coordinates already match pixel coordinates).
func ScreenshotNode(backend render.Backend, tree *core.Tree, id core.ElementID, dpiScale float32) (*image.RGBA, error) {
	elem := tree.Get(id)
	if elem == nil {
		return nil, fmt.Errorf("capture: element %d not found", id)
	}
	bounds := elem.Layout().Bounds
	if bounds.IsEmpty() {
		return nil, fmt.Errorf("capture: element %d has empty bounds", id)
	}
	full, err := backend.ReadPixels()
	if err != nil {
		return nil, err
	}
	return CropRect(full, bounds, dpiScale), nil
}

// ScreenshotBounds captures a region of the framebuffer defined by logical bounds.
// The dpiScale parameter converts logical coordinates to physical pixels.
func ScreenshotBounds(backend render.Backend, bounds uimath.Rect, dpiScale float32) (*image.RGBA, error) {
	full, err := backend.ReadPixels()
	if err != nil {
		return nil, err
	}
	return CropRect(full, bounds, dpiScale), nil
}

// CropRect crops an image to the given logical rect, applying dpiScale.
func CropRect(img *image.RGBA, bounds uimath.Rect, dpiScale float32) *image.RGBA {
	x0 := int(bounds.X * dpiScale)
	y0 := int(bounds.Y * dpiScale)
	x1 := int((bounds.X + bounds.Width) * dpiScale)
	y1 := int((bounds.Y + bounds.Height) * dpiScale)

	// Clamp to image bounds
	imgBounds := img.Bounds()
	if x0 < imgBounds.Min.X {
		x0 = imgBounds.Min.X
	}
	if y0 < imgBounds.Min.Y {
		y0 = imgBounds.Min.Y
	}
	if x1 > imgBounds.Max.X {
		x1 = imgBounds.Max.X
	}
	if y1 > imgBounds.Max.Y {
		y1 = imgBounds.Max.Y
	}

	w := x1 - x0
	h := y1 - y0
	if w <= 0 || h <= 0 {
		return image.NewRGBA(image.Rect(0, 0, 0, 0))
	}

	cropped := image.NewRGBA(image.Rect(0, 0, w, h))
	for dy := 0; dy < h; dy++ {
		srcOff := img.PixOffset(x0, y0+dy)
		dstOff := cropped.PixOffset(0, dy)
		copy(cropped.Pix[dstOff:dstOff+w*4], img.Pix[srcOff:srcOff+w*4])
	}
	return cropped
}

// Crop extracts a sub-region from an image using pixel coordinates.
func Crop(img *image.RGBA, x, y, w, h int) *image.RGBA {
	return CropRect(img, uimath.NewRect(float32(x), float32(y), float32(w), float32(h)), 1.0)
}

// SavePNG saves an RGBA image to a PNG file at the given path.
func SavePNG(img *image.RGBA, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("capture: create %s: %w", path, err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("capture: encode png: %w", err)
	}
	return nil
}

// LoadPNG loads a PNG file as an RGBA image.
func LoadPNG(path string) (*image.RGBA, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("capture: open %s: %w", path, err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("capture: decode png: %w", err)
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}
	return rgba, nil
}

// DiffResult contains the result of comparing two images.
type DiffResult struct {
	// MeanError is the average per-channel absolute difference (0..1 range).
	MeanError float64
	// MaxError is the maximum per-channel absolute difference (0..1 range).
	MaxError float64
	// DiffPixels is the number of pixels that differ by more than threshold.
	DiffPixels int
	// TotalPixels is the total number of pixels compared.
	TotalPixels int
	// DiffImage is the visual diff (amplified differences), nil if images match.
	DiffImage *image.RGBA
}

// Compare compares two RGBA images pixel by pixel.
// The threshold (0..1) determines the per-channel tolerance for considering pixels "different".
func Compare(a, b *image.RGBA, threshold float64) (*DiffResult, error) {
	if a.Bounds() != b.Bounds() {
		return nil, fmt.Errorf("capture: image size mismatch: %v vs %v", a.Bounds().Size(), b.Bounds().Size())
	}

	bounds := a.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	total := w * h
	threshByte := uint8(threshold * 255)

	var sumErr float64
	var maxErr float64
	diffCount := 0
	diff := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			ar, ag, ab, aa := a.At(x, y).RGBA()
			br, bg, bb, ba := b.At(x, y).RGBA()

			dr := absDiff16(ar, br)
			dg := absDiff16(ag, bg)
			db := absDiff16(ab, bb)
			da := absDiff16(aa, ba)

			chErr := float64(dr+dg+db+da) / (4.0 * 65535.0)
			sumErr += chErr
			if chErr > maxErr {
				maxErr = chErr
			}

			maxCh := maxU8(uint8(dr>>8), uint8(dg>>8), uint8(db>>8), uint8(da>>8))
			if maxCh > threshByte {
				diffCount++
				// Amplify diff for visibility
				diff.Set(x, y, color.RGBA{R: 255, A: 255})
			}
		}
	}

	result := &DiffResult{
		MeanError:   sumErr / float64(total),
		MaxError:    maxErr,
		DiffPixels:  diffCount,
		TotalPixels: total,
	}
	if diffCount > 0 {
		result.DiffImage = diff
	}
	return result, nil
}

// MatchesGolden compares the image against a golden file.
// If the golden file doesn't exist, it creates it (first-run bootstrapping).
// Returns the diff result and any error.
func MatchesGolden(img *image.RGBA, goldenPath string, threshold float64) (*DiffResult, error) {
	// If golden doesn't exist, create it (bootstrap mode)
	if _, err := os.Stat(goldenPath); os.IsNotExist(err) {
		if err := SavePNG(img, goldenPath); err != nil {
			return nil, fmt.Errorf("capture: bootstrap golden: %w", err)
		}
		return &DiffResult{TotalPixels: img.Bounds().Dx() * img.Bounds().Dy()}, nil
	}

	golden, err := LoadPNG(goldenPath)
	if err != nil {
		return nil, err
	}

	return Compare(img, golden, threshold)
}

// Diff returns the normalized mean error between two images.
// Convenience wrapper around Compare for simple pass/fail checks.
func Diff(a, b *image.RGBA) (float64, error) {
	result, err := Compare(a, b, 0)
	if err != nil {
		return 1.0, err
	}
	return result.MeanError, nil
}

func absDiff16(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}

func maxU8(a, b, c, d uint8) uint8 {
	m := a
	if b > m {
		m = b
	}
	if c > m {
		m = c
	}
	if d > m {
		m = d
	}
	return m
}

// PSNR calculates the Peak Signal-to-Noise Ratio between two images.
// Higher values mean more similar images. Returns +Inf for identical images.
func PSNR(a, b *image.RGBA) (float64, error) {
	if a.Bounds() != b.Bounds() {
		return 0, fmt.Errorf("capture: image size mismatch")
	}

	bounds := a.Bounds()
	var mse float64
	n := 0

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			ar, ag, ab, _ := a.At(x, y).RGBA()
			br, bg, bb, _ := b.At(x, y).RGBA()

			dr := float64(ar) - float64(br)
			dg := float64(ag) - float64(bg)
			db := float64(ab) - float64(bb)

			mse += (dr*dr + dg*dg + db*db) / 3.0
			n++
		}
	}

	if n == 0 {
		return math.Inf(1), nil
	}
	mse /= float64(n)
	if mse == 0 {
		return math.Inf(1), nil
	}

	maxVal := 65535.0
	return 10 * math.Log10(maxVal*maxVal/mse), nil
}

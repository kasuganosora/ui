package core

import "image"

// HitTestFromAlpha creates a HitTestFunc that samples the alpha channel of
// img at the point corresponding to (localX, localY) within a widget of size
// (w × h).  Pixels with alpha < threshold are treated as transparent
// (pass-through).
//
// The image is sampled using nearest-neighbour mapping:
//
//	imgX = localX / w * imgWidth
//	imgY = localY / h * imgHeight
//
// This is useful for "shaped" or irregularly-outlined windows: supply the
// same PNG used as the window background, and clicks on fully-transparent
// regions will fall through to whatever is underneath.
//
// Example:
//
//	img, _ := png.Decode(f)
//	win.SetHitTestFunc(core.HitTestFromAlpha(img, 300, 200, 0))
func HitTestFromAlpha(img image.Image, w, h float32, threshold uint8) HitTestFunc {
	bounds := img.Bounds()
	imgW := float32(bounds.Dx())
	imgH := float32(bounds.Dy())
	if imgW == 0 || imgH == 0 || w == 0 || h == 0 {
		return nil
	}
	return func(localX, localY float32) bool {
		if localX < 0 || localY < 0 || localX >= w || localY >= h {
			return false
		}
		ix := int(localX / w * imgW)
		iy := int(localY / h * imgH)
		if ix >= bounds.Dx() {
			ix = bounds.Dx() - 1
		}
		if iy >= bounds.Dy() {
			iy = bounds.Dy() - 1
		}
		_, _, _, a := img.At(bounds.Min.X+ix, bounds.Min.Y+iy).RGBA()
		// RGBA returns 16-bit; threshold is 8-bit
		return uint8(a>>8) > threshold
	}
}

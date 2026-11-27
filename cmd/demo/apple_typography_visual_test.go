package main

import (
	"image"
	"testing"

	"github.com/kasuganosora/ui/font"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/render/capture"
	"github.com/kasuganosora/ui/widget"
)

// renderCommands submits a single frame built from raw commands. The shared
// renderFrames() helper needs a Widget root; typography A/B specimens are
// drawn directly so both rows go through exactly the same layout path.
// Called from a single test body, never a t.Run subtest: PollEvents() must
// stay on the AppKit main thread.
func (e *testEnv) renderCommands(t *testing.T, draw func(*render.CommandBuffer)) {
	t.Helper()
	e.plat.PollEvents()
	// TextCmd captures the atlas texture handle at draw time. Atlas.Upload()
	// only creates the texture when glyphs are dirty, so a cold first frame
	// would emit handle 0 and Metal would silently skip every glyph. Ensure
	// the texture exists before the first command is built.
	if err := e.glyphAtlas.EnsureTexture(); err != nil {
		t.Fatalf("glyph atlas texture: %v", err)
	}
	e.backend.BeginFrame()
	e.textRenderer.BeginFrame()
	e.buf.Reset()
	draw(e.buf)
	if err := e.textRenderer.Upload(); err != nil {
		t.Fatalf("glyph atlas upload: %v", err)
	}
	e.backend.Submit(e.buf)
	e.backend.EndFrame()
}

// regionInk measures how much dark ink a text row actually put on the
// framebuffer. A command-buffer assertion cannot see this: the earlier
// #A3A3A3 caption bug was invisible until pixels were read back.
func regionInk(img *image.RGBA, r image.Rectangle) (float64, uint8) {
	ink := 0.0
	darkest := uint8(255)
	for y := r.Min.Y; y < r.Max.Y && y < img.Bounds().Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X && x < img.Bounds().Max.X; x++ {
			lr, _, _, _ := img.At(x, y).RGBA()
			lum := uint8(lr >> 8)
			ink += float64(255 - lum)
			if lum < darkest {
				darkest = lum
			}
		}
	}
	return ink, darkest
}

func logicalRegion(e *testEnv, x, y, w, h float32) image.Rectangle {
	s := e.backend.DPIScale()
	return image.Rect(int(x*s), int(y*s), int((x+w)*s), int((y+h)*s))
}

// typographyChangedPixels is local to this batch so the typography specimen
// does not depend on helpers from another review batch.
func typographyChangedPixels(a, b *image.RGBA) int {
	n := 0
	for y := 0; y < a.Bounds().Dy(); y++ {
		for x := 0; x < a.Bounds().Dx(); x++ {
			r1, g1, b1, a1 := a.At(x, y).RGBA()
			r2, g2, b2, a2 := b.At(x, y).RGBA()
			if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
				n++
			}
		}
	}
	return n
}

// TestTypographyReachesRealFrame closes the loop for §15 on real Metal
// pixels: caption weight, optical tracking and Dynamic Type must all change
// what is painted, not just what sits in the command buffer.
func TestTypographyReachesRealFrame(t *testing.T) {
	e := newTestEnv(t, 720, 420)
	defer e.close()

	if id := e.enableCaptionFont(); id == font.InvalidFontID {
		t.Skip("font engine cannot select a medium-weight face")
	}
	caption, ok := e.cfg.TextRenderer.(widget.CaptionTextDrawer)
	if !ok {
		t.Fatal("text drawer does not implement the caption path")
	}
	tracked, ok := e.cfg.TextRenderer.(widget.TrackingTextDrawer)
	if !ok {
		t.Fatal("text drawer does not implement tracking")
	}

	// Reproduce the original symptom: 0.6-alpha secondary text on a light
	// surface. With pure black, both rows saturate at 0 and the coverage
	// difference (the #A3A3A3 "grey mush" bug) becomes invisible. Use the
	// literal here so this typography batch stays independent of theme tokens.
	ink := uimath.RGBA(0, 0, 0, 0.6)
	bg := func(buf *render.CommandBuffer) {
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(0, 0, 720, 420),
			FillColor: uimath.ColorWhite,
		}, 0, 1)
	}

	// --- 1. Caption weight: same text, size, tracking and color, only the
	// registered face differs. Regular goes on row A, medium on row B.
	const label = "分类标题稳定开发中"
	const size float32 = 12
	e.renderCommands(t, func(buf *render.CommandBuffer) {
		bg(buf)
		tracked.DrawTextTracked(buf, label, 20, 40, size, 660, ink, 1, 0, e.cfg.Tracking(size))
		caption.DrawCaptionText(buf, label, 20, 140, size, 660, ink, 1, 0, e.cfg.Tracking(size))
	})
	frame, err := capture.Screenshot(e.backend)
	if err != nil {
		t.Fatalf("caption frame readback: %v", err)
	}
	regularInk, regularDarkest := regionInk(frame, logicalRegion(e, 20, 32, 660, 30))
	mediumInk, mediumDarkest := regionInk(frame, logicalRegion(e, 20, 132, 660, 30))
	t.Logf("12px caption: regular ink=%.0f darkest=%d, medium ink=%.0f darkest=%d (0.6-black floor ~102)",
		regularInk, regularDarkest, mediumInk, mediumDarkest)
	if regularInk <= 0 || mediumInk <= 0 {
		t.Fatal("one of the caption rows produced no ink")
	}
	if mediumInk < regularInk*1.15 {
		t.Errorf("medium caption did not reach the framebuffer: ink %.0f vs %.0f", mediumInk, regularInk)
	}
	if mediumDarkest > regularDarkest {
		t.Errorf("medium caption has worse peak coverage: %d vs %d", mediumDarkest, regularDarkest)
	}
	if regularDarkest <= 102 {
		t.Error("fixture no longer reproduces the thin-outline case; the caption weight test is vacuous")
	}

	// --- 2. Optical tracking: a 72px heading with negative tracking must
	// occupy different pixels than the same heading with zero tracking.
	const heading = "Typography Hamburgefonts"
	const big float32 = 72
	drawHeading := func(tracking float32) *image.RGBA {
		e.renderCommands(t, func(buf *render.CommandBuffer) {
			bg(buf)
			tracked.DrawTextTracked(buf, heading, 20, 240, big, 690, ink, 1, 0, tracking)
		})
		img, err := capture.Screenshot(e.backend)
		if err != nil {
			t.Fatalf("tracking frame readback: %v", err)
		}
		return img
	}
	noTracking := drawHeading(0)
	withTracking := drawHeading(e.cfg.Tracking(big))
	delta := typographyChangedPixels(noTracking, withTracking)
	t.Logf("72px tracking=%g changed %d pixels", e.cfg.Tracking(big), delta)
	if e.cfg.Tracking(big) >= 0 {
		t.Fatal("large heading tracking should be negative")
	}
	if delta < 500 {
		t.Errorf("tracking never reached the framebuffer (%d changed pixels)", delta)
	}

	// --- 3. Dynamic Type: scaling from 1x to 2x must visibly grow the caption.
	e.cfg.SetTextScale(1)
	small := e.cfg.CaptionSize()
	e.renderCommands(t, func(buf *render.CommandBuffer) {
		bg(buf)
		caption.DrawCaptionText(buf, label, 20, 340, small, 660, ink, 1, 0, e.cfg.Tracking(small))
	})
	smallFrame, err := capture.Screenshot(e.backend)
	if err != nil {
		t.Fatalf("1x caption frame readback: %v", err)
	}
	e.cfg.SetTextScale(2)
	large := e.cfg.CaptionSize()
	e.renderCommands(t, func(buf *render.CommandBuffer) {
		bg(buf)
		caption.DrawCaptionText(buf, label, 20, 340, large, 660, ink, 1, 0, e.cfg.Tracking(large))
	})
	largeFrame, err := capture.Screenshot(e.backend)
	if err != nil {
		t.Fatalf("2x caption frame readback: %v", err)
	}
	smallInk, _ := regionInk(smallFrame, logicalRegion(e, 20, 330, 660, 70))
	largeInk, _ := regionInk(largeFrame, logicalRegion(e, 20, 330, 660, 70))
	t.Logf("Dynamic Type caption: 1x size=%g ink=%.0f, 2x size=%g ink=%.0f", small, smallInk, large, largeInk)
	if large != small*2 {
		t.Errorf("caption size did not double: %g -> %g", small, large)
	}
	if largeInk < smallInk*1.8 {
		t.Errorf("Dynamic Type did not change painted text: ink %.0f -> %.0f", smallInk, largeInk)
	}
}

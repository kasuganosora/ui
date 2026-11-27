//go:build darwin

package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/render/capture"
	"github.com/kasuganosora/ui/widget"
)

// TestTooltipOverlayPixelsOnRealGUI closes the last gap in the "tooltip is an
// independent OS window" decision: earlier tests proved the host creates a
// second window and submits commands to a second backend, but only with a
// software backend. Here a real AppKit NSWindow + Metal pipeline must produce
// actual pixels — a dark bubble with readable glyphs — while the main window
// shows no bubble at all.
//
// Single test body, never t.Run: NewApp bootstraps Cocoa on the main thread.
func TestTooltipOverlayPixelsOnRealGUI(t *testing.T) {
	app, err := NewApp(AppOptions{
		Title:  "GoUI Tooltip Overlay",
		Width:  640,
		Height: 480,
	})
	if err != nil {
		t.Fatalf("NewApp on real GUI: %v", err)
	}
	defer app.Destroy()

	tree := app.Tree()
	anchor := tree.CreateElement(core.TypeButton)
	tree.SetLayout(anchor, core.LayoutResult{Bounds: uimath.NewRect(200, 200, 120, 32)})

	// Simplified Chinese + Latin: the label must survive the TTC face
	// selection and the auxiliary window's private glyph atlas.
	tt := widget.NewTooltip(tree, "提示文本 Tooltip", anchor, app.Config())
	tt.Show()

	if len(app.overlays) != 1 {
		t.Fatalf("expected exactly one auxiliary OS window, got %d", len(app.overlays))
	}
	ow := app.overlays[tt]
	if ow == nil {
		t.Fatal("tooltip did not register an auxiliary window")
	}
	if ow.win == app.win || ow.backend == app.backend {
		t.Fatal("tooltip reused the main window or its backend")
	}
	if ow.win.NativeHandle() == 0 {
		t.Fatal("auxiliary window has no native OS handle")
	}
	if !ow.win.IsTransparent() {
		t.Error("auxiliary window is not transparent; the bubble would be a grey rectangle")
	}
	if ow.glyphs == nil || ow.text == nil {
		t.Fatal("auxiliary window is missing its private glyph resources")
	}

	app.renderOverlays()
	img, err := ow.backend.ReadPixels()
	if err != nil {
		t.Fatalf("auxiliary Metal framebuffer readback: %v", err)
	}

	// Default tooltip theme: 75% black bubble, white glyphs, fully
	// transparent outside. Count both so "a dark rectangle with no text"
	// cannot pass.
	bubblePixels, glyphPixels := 0, 0
	for i := 0; i < len(img.Pix); i += 4 {
		r, g, b, a := img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3]
		if a == 0 {
			continue
		}
		bubblePixels++
		if r > 90 && g > 90 && b > 90 {
			glyphPixels++
		}
	}
	t.Logf("auxiliary window %dx%d: bubble pixels=%d glyph pixels=%d",
		img.Bounds().Dx(), img.Bounds().Dy(), bubblePixels, glyphPixels)
	if bubblePixels < 2000 {
		t.Errorf("bubble never reached the auxiliary framebuffer (%d pixels)", bubblePixels)
	}
	if glyphPixels < 200 {
		t.Errorf("label glyphs never reached the auxiliary framebuffer (%d pixels)", glyphPixels)
	}

	// Save a specimen so the bubble can also be inspected by eye; this is a
	// temporary diagnostic outside the repository, not a golden screenshot.
	shot := filepath.Join(os.TempDir(), "goui_tooltip_overlay_metal.png")
	if err := capture.SavePNG(img, shot); err != nil {
		t.Logf("could not save overlay specimen: %v", err)
	} else {
		t.Logf("overlay specimen: %s", shot)
	}

	// The same widget must not also paint a bubble into the main window.
	app.backend.BeginFrame()
	app.buf.Reset()
	app.buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(0, 0, 640, 480),
		FillColor: uimath.ColorWhite,
	}, 0, 1)
	tt.Draw(app.buf)
	app.backend.Submit(app.buf)
	app.backend.EndFrame()
	mainImg, err := app.backend.ReadPixels()
	if err != nil {
		t.Fatalf("main Metal framebuffer readback: %v", err)
	}
	scale := app.backend.DPIScale()
	dark := 0
	for y := int(140 * scale); y < int(210*scale) && y < mainImg.Bounds().Dy(); y++ {
		for x := int(180 * scale); x < int(340*scale) && x < mainImg.Bounds().Dx(); x++ {
			r, _, _, _ := mainImg.At(x, y).RGBA()
			if uint8(r>>8) < 120 {
				dark++
			}
		}
	}
	t.Logf("main window bubble region: %d dark pixels", dark)
	if dark != 0 {
		t.Errorf("tooltip also painted %d dark pixels into the main window", dark)
	}

	tt.Hide()
	if ow.active {
		t.Error("auxiliary window stayed active after Hide()")
	}
}

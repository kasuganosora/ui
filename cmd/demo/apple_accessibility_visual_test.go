package main

import (
	"image"
	"testing"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render/capture"
	"github.com/kasuganosora/ui/widget"
)

func countChangedPixels(a, b *image.RGBA) int {
	changed := 0
	for y := 0; y < a.Bounds().Dy(); y++ {
		for x := 0; x < a.Bounds().Dx(); x++ {
			r1, g1, b1, a1 := a.At(x, y).RGBA()
			r2, g2, b2, a2 := b.At(x, y).RGBA()
			if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
				changed++
			}
		}
	}
	return changed
}

// TestAccessibilityPreferencesChangeRealFrame closes the loop the command
// buffer tests cannot: a border or scrim in the command list is only useful if
// the backend actually paints different pixels. Each signal is applied on its
// own so neither can hide behind the other. No screenshot is written, so this
// test does not touch visual golden files.
//
// All frames are rendered in this single test body, NOT in t.Run subtests:
// renderFrames calls PollEvents(), and AppKit requires nextEventMatchingMask
// to run on the OS main thread. A subtest goroutine would abort the process.
func TestAccessibilityPreferencesChangeRealFrame(t *testing.T) {
	e := newTestEnv(t, 640, 480)
	defer e.close()

	dlg := widget.NewDialog(e.tree, "对话框", e.cfg)
	dlg.SetHeader("标题")
	dlg.SetFooter(true)
	dlg.Open()
	e.tree.SetLayout(dlg.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 640, 480),
	})

	e.renderFrames(dlg, 2)
	before, err := capture.Screenshot(e.backend)
	if err != nil {
		t.Fatalf("default frame readback: %v", err)
	}

	cases := []struct {
		name  string
		prefs widget.AccessibilityPreferences
	}{
		{"more contrast", widget.AccessibilityPreferences{MoreContrast: true}},
		{"reduced transparency", widget.AccessibilityPreferences{ReducedTransparency: true}},
	}
	var solidFrame *image.RGBA
	for _, tc := range cases {
		e.cfg.ApplyAccessibility(tc.prefs)
		e.renderFrames(dlg, 2)
		after, err := capture.Screenshot(e.backend)
		if err != nil {
			t.Fatalf("%s frame readback: %v", tc.name, err)
		}
		if before.Bounds() != after.Bounds() {
			t.Fatalf("%s: framebuffer changed size: %v -> %v", tc.name, before.Bounds(), after.Bounds())
		}
		changed := countChangedPixels(before, after)
		t.Logf("%s: pixels changed = %d", tc.name, changed)
		if changed == 0 {
			t.Errorf("%s produced no visible change; it never reached the backend", tc.name)
		}
		if tc.name == "reduced transparency" {
			solidFrame = after
		}
	}

	// Both signals solidify the scrim identically, so comparing these two
	// frames isolates the contrasting panel outline. A command-buffer border
	// that the backend never painted would leave this difference at zero.
	e.cfg.ApplyAccessibility(widget.AccessibilityPreferences{
		ReducedTransparency: true,
		MoreContrast:        true,
	})
	e.renderFrames(dlg, 2)
	withOutline, err := capture.Screenshot(e.backend)
	if err != nil {
		t.Fatalf("contrast frame readback: %v", err)
	}
	outlineChanged := countChangedPixels(solidFrame, withOutline)
	t.Logf("outline-only pixels changed = %d", outlineChanged)
	if outlineChanged == 0 {
		t.Error("prefers-contrast outline did not paint any pixels")
	}
}

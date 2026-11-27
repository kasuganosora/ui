package main

import (
	"testing"
	"time"

	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render/capture"
	"github.com/kasuganosora/ui/widget"
)

// TestMomentumProjectionCoastsInRealFrames closes the loop for Apple §5/§6:
// a flick must keep moving after release and land near the projected point,
// not stop at the last wheel tick. Frames are read back from the real Metal
// backend; no golden PNG is written.
//
// Single test body, NOT t.Run subtests: renderFrames calls PollEvents(), and
// AppKit requires nextEventMatchingMask to run on the OS main thread.
func TestMomentumProjectionCoastsInRealFrames(t *testing.T) {
	e := newTestEnv(t, 640, 480)
	defer e.close()
	e.cfg.ApplyAccessibility(widget.AccessibilityPreferences{})

	root := widget.NewLayout(e.tree, e.cfg)
	body := widget.NewDiv(e.tree, e.cfg)
	content := widget.NewContent(e.tree, e.cfg)
	content.SetBgColor(uimath.ColorHex("#f2f2f2"))
	for i := 0; i < 24; i++ {
		row := widget.NewDiv(e.tree, e.cfg)
		if i%2 == 0 {
			row.SetBgColor(uimath.ColorHex("#3b82f6"))
		} else {
			row.SetBgColor(uimath.ColorHex("#ef4444"))
		}
		content.AppendChild(row)
	}
	body.AppendChild(content)
	root.AppendChild(body)
	e.tree.AppendChild(e.tree.Root(), root.ElementID())

	e.renderFrames(root, 3)
	if content.ContentHeight() <= 480 {
		t.Fatalf("fixture does not overflow: contentHeight=%f", content.ContentHeight())
	}
	baseline, err := capture.Screenshot(e.backend)
	if err != nil {
		t.Fatalf("baseline readback: %v", err)
	}

	// Three wheel ticks ~16ms apart form one measurable flick inside the
	// legal range. Reduced motion is off, so the release should project.
	for i := 0; i < 3; i++ {
		content.HandleWheel(-1)
		e.renderFrames(root, 1)
		time.Sleep(16 * time.Millisecond)
	}
	atRelease := content.ScrollY()
	if atRelease <= 0 {
		t.Fatalf("flick did not scroll inside the bounds: %f", atRelease)
	}
	atReleaseFrame, err := capture.Screenshot(e.backend)
	if err != nil {
		t.Fatalf("release readback: %v", err)
	}
	if countChangedPixels(baseline, atReleaseFrame) == 0 {
		t.Fatal("the three ticks never reached the framebuffer")
	}

	// Let the gesture gap pass, then drive real frames through Draw().
	time.Sleep(160 * time.Millisecond)
	for i := 0; i < 40; i++ {
		e.renderFrames(root, 1)
		time.Sleep(16 * time.Millisecond)
	}
	coasted, err := capture.Screenshot(e.backend)
	if err != nil {
		t.Fatalf("coasted readback: %v", err)
	}
	delta := countChangedPixels(atReleaseFrame, coasted)
	t.Logf("release pos=%f coasted pos=%f, pixels changed after release=%d", atRelease, content.ScrollY(), delta)
	if content.ScrollY() <= atRelease+20 {
		t.Errorf("surface stopped at the release point instead of coasting: %f -> %f", atRelease, content.ScrollY())
	}
	if delta < 10000 {
		t.Errorf("momentum never reached the framebuffer (%d changed pixels)", delta)
	}

	// §14 gate: with reduced motion the same flick must not coast.
	e.cfg.ApplyAccessibility(widget.AccessibilityPreferences{ReducedMotion: true})
	content.ScrollTo(0)
	e.renderFrames(root, 2)
	for i := 0; i < 3; i++ {
		content.HandleWheel(-1)
		e.renderFrames(root, 1)
		time.Sleep(16 * time.Millisecond)
	}
	hardStop := content.ScrollY()
	time.Sleep(160 * time.Millisecond)
	for i := 0; i < 20; i++ {
		e.renderFrames(root, 1)
		time.Sleep(16 * time.Millisecond)
	}
	if content.ScrollY() != hardStop {
		t.Errorf("reduced motion played a decorative coast: %f -> %f", hardStop, content.ScrollY())
	}
}

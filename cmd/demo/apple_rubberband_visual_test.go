package main

import (
	"testing"
	"time"

	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render/capture"
	"github.com/kasuganosora/ui/widget"
)

// TestRubberbandOverscrollSurvivesRealLayoutFrames closes the loop the widget
// unit tests cannot: scroll offsets in this framework are applied by the
// layout pass, so a per-frame clamp in AutoLayout would silently erase the
// soft boundary. This renders real Metal frames (no golden PNG is written).
//
// Single test body, NOT t.Run subtests: renderFrames calls PollEvents(), and
// AppKit requires nextEventMatchingMask to run on the OS main thread.
func TestRubberbandOverscrollSurvivesRealLayoutFrames(t *testing.T) {
	e := newTestEnv(t, 640, 480)
	defer e.close()

	// Keep the motion gate deterministic even if the host sets GOUI_REDUCED_MOTION.
	e.cfg.ApplyAccessibility(widget.AccessibilityPreferences{})

	root := widget.NewLayout(e.tree, e.cfg)
	body := widget.NewDiv(e.tree, e.cfg)
	content := widget.NewContent(e.tree, e.cfg)
	content.SetBgColor(uimath.ColorHex("#f2f2f2"))
	for i := 0; i < 20; i++ {
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

	// Let the layout callback settle contentHeight before the gesture starts.
	e.renderFrames(root, 3)
	if content.ContentHeight() <= 480 {
		t.Fatalf("fixture does not overflow: contentHeight=%f", content.ContentHeight())
	}
	baseline, err := capture.Screenshot(e.backend)
	if err != nil {
		t.Fatalf("baseline readback: %v", err)
	}

	content.HandleWheel(100)
	if content.ScrollY() >= 0 {
		t.Fatalf("fixture did not overscroll: %f", content.ScrollY())
	}
	e.renderFrames(root, 2)
	if content.ScrollY() >= 0 {
		t.Fatalf("AutoLayout erased the overscroll: %f", content.ScrollY())
	}
	stretched, err := capture.Screenshot(e.backend)
	if err != nil {
		t.Fatalf("stretched readback: %v", err)
	}
	stretchDelta := countChangedPixels(baseline, stretched)
	t.Logf("overscroll frame: pixels changed = %d", stretchDelta)
	if stretchDelta < 10000 {
		t.Errorf("overscroll never reached the framebuffer (%d changed pixels)", stretchDelta)
	}

	// Release and let real frames drive the spring back through Draw().
	time.Sleep(160 * time.Millisecond)
	for i := 0; i < 24; i++ {
		e.renderFrames(root, 1)
		time.Sleep(40 * time.Millisecond)
	}
	if content.ScrollY() > 2 || content.ScrollY() < -2 {
		t.Errorf("surface did not spring back to the boundary: %f", content.ScrollY())
	}
	settled, err := capture.Screenshot(e.backend)
	if err != nil {
		t.Fatalf("settled readback: %v", err)
	}
	backDelta := countChangedPixels(baseline, settled)
	t.Logf("settled frame: pixels changed vs baseline = %d", backDelta)
	if backDelta > stretchDelta/10 {
		t.Errorf("surface did not return to its original frame (%d residual pixels)", backDelta)
	}
}

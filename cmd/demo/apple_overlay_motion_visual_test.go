package main

import (
	"testing"
	"time"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render/capture"
	"github.com/kasuganosora/ui/widget"
)

// TestAnchoredOverlayMotionInRealFrames closes the §7 loop: the enter and
// exit transitions must be visible in the Metal framebuffer, and the surface
// must end exactly where it started. No golden PNG is written.
//
// Single test body, NOT t.Run subtests: renderFrames calls PollEvents(), and
// AppKit requires nextEventMatchingMask to run on the OS main thread.
func TestAnchoredOverlayMotionInRealFrames(t *testing.T) {
	e := newTestEnv(t, 640, 480)
	defer e.close()
	e.cfg.ApplyAccessibility(widget.AccessibilityPreferences{})
	e.cfg.AnimateOverlays = true

	p := widget.NewPopover(e.tree, e.cfg)
	p.SetTitle("锚定浮层")
	p.SetWidth(220)
	p.SetPlacement(widget.PlacementBottom)
	p.SetAnchorRect(200, 120, 80, 30)
	e.tree.AppendChild(e.tree.Root(), p.ElementID())
	e.tree.SetLayout(p.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 640, 480),
	})

	e.renderFrames(p, 2)
	baseline, err := capture.Screenshot(e.backend)
	if err != nil {
		t.Fatalf("baseline readback: %v", err)
	}

	p.Open()
	e.renderFrames(p, 1)
	early, err := capture.Screenshot(e.backend)
	if err != nil {
		t.Fatalf("early enter readback: %v", err)
	}
	for i := 0; i < 30; i++ {
		e.renderFrames(p, 1)
		time.Sleep(16 * time.Millisecond)
	}
	settled, err := capture.Screenshot(e.backend)
	if err != nil {
		t.Fatalf("settled readback: %v", err)
	}
	enterProgress := countChangedPixels(early, settled)
	enterVsBaseline := countChangedPixels(baseline, settled)
	t.Logf("enter: early->settled %d pixels, settled vs baseline %d pixels", enterProgress, enterVsBaseline)
	if enterProgress < 5000 || enterVsBaseline < 5000 {
		t.Errorf("enter transition never reached the framebuffer (progress=%d visible=%d)", enterProgress, enterVsBaseline)
	}

	p.Close()
	e.renderFrames(p, 1)
	exitEarly, err := capture.Screenshot(e.backend)
	if err != nil {
		t.Fatalf("early exit readback: %v", err)
	}
	for i := 0; i < 30; i++ {
		e.renderFrames(p, 1)
		time.Sleep(16 * time.Millisecond)
	}
	gone, err := capture.Screenshot(e.backend)
	if err != nil {
		t.Fatalf("gone readback: %v", err)
	}
	exitProgress := countChangedPixels(exitEarly, gone)
	back := countChangedPixels(baseline, gone)
	t.Logf("exit: early->gone %d pixels, gone vs baseline %d pixels", exitProgress, back)
	if exitProgress < 5000 {
		t.Errorf("exit path was never painted (%d changed pixels)", exitProgress)
	}
	if back != 0 {
		t.Errorf("panel did not return to the baseline frame (%d residual pixels)", back)
	}

	// §14: reduced motion keeps a static appearance with no spatial travel.
	// After one frame the panel must be fully drawn rather than displaced.
	e.cfg.ReducedMotion = true
	p.Open()
	e.renderFrames(p, 1)
	reduced, err := capture.Screenshot(e.backend)
	if err != nil {
		t.Fatalf("reduced-motion readback: %v", err)
	}
	if countChangedPixels(settled, reduced) != 0 {
		t.Error("reduced motion still displaced the panel instead of showing it statically")
	}
}

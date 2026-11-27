package main

import (
	"image"
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render/capture"
	"github.com/kasuganosora/ui/widget"
)

// pressFeedbackAABand is the logical-pixel halo allowed around the pressed
// control. The Metal SDF rect shader pads its antialiasing quad, so a tinted
// rounded rect also changes a 1px fringe just outside its bounds. That fringe
// belongs to the same command; anything beyond it would be a real leak onto
// neighbouring controls.
const pressFeedbackAABand float32 = 4

// changedPixelsByRegion splits a frame difference into pixels inside the
// pressed control's bounds (plus its AA band) and pixels outside it. A stray
// full-frame repaint would pass a plain "any pixels changed" assertion, but
// not this one.
func changedPixelsByRegion(a, b *image.RGBA, r uimath.Rect, scale float32) (inside, outside int) {
	band := pressFeedbackAABand * scale
	region := image.Rect(
		int(r.X*scale-band), int(r.Y*scale-band),
		int((r.X+r.Width)*scale+band), int((r.Y+r.Height)*scale+band),
	)
	for y := 0; y < a.Bounds().Dy(); y++ {
		for x := 0; x < a.Bounds().Dx(); x++ {
			r1, g1, b1, a1 := a.At(x, y).RGBA()
			r2, g2, b2, a2 := b.At(x, y).RGBA()
			if r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2 {
				continue
			}
			if x >= region.Min.X && x < region.Max.X && y >= region.Min.Y && y < region.Max.Y {
				inside++
			} else {
				outside++
			}
		}
	}
	return
}

// TestPressFeedbackReachesRealFrame closes the loop the command-buffer tests
// cannot: a highlight emitted into Overlays() is only useful if the backend
// paints it on the very frame of pointer-down. The fixture uses a Link, which
// registers MouseClick only and has no press state of its own, so any visible
// change comes from the shared §1 feedback layer.
//
// Single test body, NOT t.Run subtests: renderFrames calls PollEvents(), and
// AppKit requires nextEventMatchingMask to run on the OS main thread.
func TestPressFeedbackReachesRealFrame(t *testing.T) {
	e := newTestEnv(t, 640, 480)
	defer e.close()

	root := widget.NewLayout(e.tree, e.cfg)
	body := widget.NewDiv(e.tree, e.cfg)
	var links []*widget.Link
	clicks := 0
	for _, label := range []string{"First", "Second", "Third"} {
		link := widget.NewLink(e.tree, label, "#", e.cfg)
		link.OnClick(func(string) { clicks++ })
		body.AppendChild(link)
		links = append(links, link)
	}
	root.AppendChild(body)
	// HitTest walks from the tree root; rendering alone does not parent the
	// layout, so attach it explicitly before routing any pointer input.
	e.tree.AppendChild(e.tree.Root(), root.ElementID())

	e.renderFrames(root, 2)
	baseline, err := capture.Screenshot(e.backend)
	if err != nil {
		t.Fatalf("baseline readback: %v", err)
	}
	target := links[0]
	bounds := e.tree.Get(target.ElementID()).Layout().Bounds
	if bounds.IsEmpty() {
		t.Fatal("AutoLayout did not lay out the link")
	}
	px, py := bounds.X+bounds.Width/2, bounds.Y+bounds.Height/2
	scale := e.backend.DPIScale()

	host := core.HostInput{Tree: e.tree}
	host.Handle(&event.Event{Type: event.MouseDown, GlobalX: px, GlobalY: py})
	if e.tree.Pressed() != target.ElementID() {
		t.Fatalf("press published on %d, want link %d", e.tree.Pressed(), target.ElementID())
	}
	if clicks != 0 {
		t.Fatal("action fired on pointer-down; the click contract must stay on release")
	}

	e.renderFrames(root, 1)
	pressed, err := capture.Screenshot(e.backend)
	if err != nil {
		t.Fatalf("pressed readback: %v", err)
	}
	inside, outside := changedPixelsByRegion(baseline, pressed, bounds, scale)
	t.Logf("pointer-down frame: %d pixels changed on the link, %d elsewhere", inside, outside)
	if inside < 1000 {
		t.Errorf("press highlight never reached the framebuffer (%d changed pixels)", inside)
	}
	if outside != 0 {
		t.Errorf("press feedback repainted pixels outside the pressed control (%d pixels)", outside)
	}

	host.Handle(&event.Event{Type: event.MouseUp, GlobalX: px, GlobalY: py})
	if e.tree.Pressed() != core.InvalidElementID {
		t.Error("press state survived pointer release")
	}
	if clicks != 1 {
		t.Errorf("release synthesized %d clicks, want exactly 1", clicks)
	}

	e.renderFrames(root, 1)
	released, err := capture.Screenshot(e.backend)
	if err != nil {
		t.Fatalf("released readback: %v", err)
	}
	backInside, backOutside := changedPixelsByRegion(baseline, released, bounds, scale)
	t.Logf("released frame: %d residual pixels on the link, %d elsewhere", backInside, backOutside)
	if backInside != 0 || backOutside != 0 {
		t.Errorf("frame did not return to baseline after release (%d/%d pixels)", backInside, backOutside)
	}
}

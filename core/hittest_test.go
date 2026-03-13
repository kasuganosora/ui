package core

import (
	"image"
	"image/color"
	"testing"

	uimath "github.com/kasuganosora/ui/math"
)

func TestHitTestFuncOnElement(t *testing.T) {
	tree := NewTree()
	parent := tree.Root()

	child := tree.CreateElement(TypeButton)
	tree.AppendChild(parent, child)
	tree.SetLayout(parent, LayoutResult{Bounds: uimath.NewRect(0, 0, 200, 200)})
	tree.SetLayout(child, LayoutResult{Bounds: uimath.NewRect(0, 0, 100, 100)})

	// Without HitTestFunc: hit inside child bounds
	if hit := tree.HitTest(50, 50); hit != child {
		t.Fatalf("expected child, got %v", hit)
	}

	// Set HitTestFunc that makes center transparent
	tree.SetHitTestFunc(child, func(lx, ly float32) bool {
		return lx < 25 || lx > 75 || ly < 25 || ly > 75
	})

	// Hit in transparent center → should fall through to parent
	hit := tree.HitTest(50, 50)
	if hit != parent {
		t.Fatalf("expected parent (transparent center), got %v", hit)
	}

	// Hit in opaque corner → should still hit child
	hit = tree.HitTest(10, 10)
	if hit != child {
		t.Fatalf("expected child (opaque corner), got %v", hit)
	}

	// Clear HitTestFunc → back to default
	tree.SetHitTestFunc(child, nil)
	if tree.HitTestFunc(child) != nil {
		t.Fatal("expected nil after clearing")
	}
	hit = tree.HitTest(50, 50)
	if hit != child {
		t.Fatal("expected child after clearing HitTestFunc")
	}
}

func TestHitTestFuncGetterNonExistent(t *testing.T) {
	tree := NewTree()
	if fn := tree.HitTestFunc(999); fn != nil {
		t.Fatal("expected nil for non-existent element")
	}
}

func TestSetHitTestFuncNonExistent(t *testing.T) {
	tree := NewTree()
	// Should not panic
	tree.SetHitTestFunc(999, func(lx, ly float32) bool { return true })
}

func TestHitTestFromAlpha(t *testing.T) {
	// Create 4×4 NRGBA image: top-left quadrant opaque, rest transparent
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			if x < 2 && y < 2 {
				img.Set(x, y, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
			} else {
				img.Set(x, y, color.NRGBA{R: 0, G: 0, B: 0, A: 0})
			}
		}
	}

	fn := HitTestFromAlpha(img, 100, 100, 0)
	if fn == nil {
		t.Fatal("expected non-nil func")
	}

	// Top-left quadrant (0–49) → opaque
	if !fn(10, 10) {
		t.Error("top-left should be opaque")
	}
	// Bottom-right quadrant (50–99) → transparent
	if fn(75, 75) {
		t.Error("bottom-right should be transparent")
	}
	// Out of bounds
	if fn(-1, 0) {
		t.Error("negative coords should be false")
	}
	if fn(100, 50) {
		t.Error("beyond width should be false")
	}
}

func TestHitTestFromAlphaThreshold(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.NRGBA{R: 255, A: 100})
	img.Set(1, 0, color.NRGBA{R: 255, A: 200})

	fn := HitTestFromAlpha(img, 100, 100, 128)
	if fn(10, 10) {
		t.Error("alpha 100 should be below threshold 128")
	}
	if !fn(75, 10) {
		t.Error("alpha 200 should be above threshold 128")
	}
}

func TestHitTestFromAlphaEdgeCases(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 0, 0))
	if fn := HitTestFromAlpha(img, 100, 100, 0); fn != nil {
		t.Error("zero-size image should return nil")
	}
	img2 := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	if fn := HitTestFromAlpha(img2, 0, 100, 0); fn != nil {
		t.Error("zero widget width should return nil")
	}
}

func TestHitTestTransparentParentBlocksChildren(t *testing.T) {
	tree := NewTree()
	root := tree.Root()

	parent := tree.CreateElement(TypeDiv)
	tree.AppendChild(root, parent)
	tree.SetLayout(root, LayoutResult{Bounds: uimath.NewRect(0, 0, 200, 200)})
	tree.SetLayout(parent, LayoutResult{Bounds: uimath.NewRect(0, 0, 200, 200)})

	// Make parent transparent everywhere
	tree.SetHitTestFunc(parent, func(lx, ly float32) bool { return false })

	child := tree.CreateElement(TypeButton)
	tree.AppendChild(parent, child)
	tree.SetLayout(child, LayoutResult{Bounds: uimath.NewRect(50, 50, 50, 50)})

	// Parent is transparent → child is never reached → falls to root
	hit := tree.HitTest(60, 60)
	if hit != root {
		t.Fatalf("expected root (parent transparent), got %v", hit)
	}
}

func TestHitTestFromAlphaBoundaryPixels(t *testing.T) {
	// 2x2 image, all opaque
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			img.Set(x, y, color.NRGBA{R: 255, A: 255})
		}
	}
	fn := HitTestFromAlpha(img, 200, 200, 0)
	// Test edge pixels
	if !fn(0, 0) {
		t.Error("origin should be opaque")
	}
	if !fn(199, 199) {
		t.Error("max corner should be opaque")
	}
}

package math

import "testing"

func TestEdgesShrinkRect(t *testing.T) {
	r := NewRect(0, 0, 100, 100)
	e := NewEdges(10, 20, 10, 20)
	shrunk := e.ShrinkRect(r)
	expected := NewRect(20, 10, 60, 80)
	if !shrunk.Approx(expected, 1e-6) {
		t.Errorf("expected %v, got %v", expected, shrunk)
	}
}

func TestEdgesExpandRect(t *testing.T) {
	r := NewRect(20, 10, 60, 80)
	e := NewEdges(10, 20, 10, 20)
	expanded := e.ExpandRect(r)
	expected := NewRect(0, 0, 100, 100)
	if !expanded.Approx(expected, 1e-6) {
		t.Errorf("expected %v, got %v", expected, expanded)
	}
}

func TestEdgesSymmetric(t *testing.T) {
	e := EdgesSymmetric(10, 20)
	if e.Top != 10 || e.Bottom != 10 || e.Left != 20 || e.Right != 20 {
		t.Errorf("unexpected symmetric edges: %v", e)
	}
}

func TestEdgesHorizontalVertical(t *testing.T) {
	e := NewEdges(5, 10, 15, 20)
	if e.Horizontal() != 30 {
		t.Errorf("horizontal: expected 30, got %v", e.Horizontal())
	}
	if e.Vertical() != 20 {
		t.Errorf("vertical: expected 20, got %v", e.Vertical())
	}
}

func TestCornersAll(t *testing.T) {
	c := CornersAll(8)
	if !c.IsUniform() {
		t.Error("CornersAll should be uniform")
	}
	if c.TopLeft != 8 {
		t.Errorf("expected 8, got %v", c.TopLeft)
	}
}

func TestCornersZero(t *testing.T) {
	c := CornersZero()
	if !c.IsZero() {
		t.Error("CornersZero should be zero")
	}
}

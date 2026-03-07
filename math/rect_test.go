package math

import "testing"

func TestRectContains(t *testing.T) {
	r := NewRect(10, 20, 100, 50)
	tests := []struct {
		p    Vec2
		want bool
	}{
		{NewVec2(50, 40), true},
		{NewVec2(10, 20), true},  // top-left corner
		{NewVec2(5, 40), false},  // left of rect
		{NewVec2(50, 80), false}, // below rect
		{NewVec2(110, 40), false}, // right edge (exclusive)
	}
	for _, tt := range tests {
		if got := r.Contains(tt.p); got != tt.want {
			t.Errorf("Contains(%v): got %v, want %v", tt.p, got, tt.want)
		}
	}
}

func TestRectIntersects(t *testing.T) {
	a := NewRect(0, 0, 100, 100)
	b := NewRect(50, 50, 100, 100)
	c := NewRect(200, 200, 10, 10)

	if !a.Intersects(b) {
		t.Error("a and b should intersect")
	}
	if a.Intersects(c) {
		t.Error("a and c should not intersect")
	}
}

func TestRectIntersection(t *testing.T) {
	a := NewRect(0, 0, 100, 100)
	b := NewRect(50, 50, 100, 100)
	inter := a.Intersection(b)
	expected := NewRect(50, 50, 50, 50)
	if !inter.Approx(expected, 1e-6) {
		t.Errorf("expected %v, got %v", expected, inter)
	}
}

func TestRectIntersectionNoOverlap(t *testing.T) {
	a := NewRect(0, 0, 10, 10)
	b := NewRect(20, 20, 10, 10)
	inter := a.Intersection(b)
	if !inter.IsEmpty() {
		t.Errorf("expected empty rect, got %v", inter)
	}
}

func TestRectUnion(t *testing.T) {
	a := NewRect(10, 10, 20, 20)
	b := NewRect(50, 50, 30, 30)
	u := a.Union(b)
	expected := NewRect(10, 10, 70, 70)
	if !u.Approx(expected, 1e-6) {
		t.Errorf("expected %v, got %v", expected, u)
	}
}

func TestRectExpand(t *testing.T) {
	r := NewRect(10, 10, 50, 50)
	expanded := r.Expand(5)
	if expanded.X != 5 || expanded.Y != 5 || expanded.Width != 60 || expanded.Height != 60 {
		t.Errorf("unexpected expand result: %v", expanded)
	}
}

func TestRectOffset(t *testing.T) {
	r := NewRect(10, 20, 30, 40)
	moved := r.Offset(5, -5)
	if moved.X != 15 || moved.Y != 15 || moved.Width != 30 || moved.Height != 40 {
		t.Errorf("unexpected offset result: %v", moved)
	}
}

func TestRectCenter(t *testing.T) {
	r := NewRect(0, 0, 100, 200)
	c := r.Center()
	if c.X != 50 || c.Y != 100 {
		t.Errorf("expected (50,100), got (%v,%v)", c.X, c.Y)
	}
}

func TestRectFromMinMax(t *testing.T) {
	r := RectFromMinMax(10, 20, 110, 70)
	if r.X != 10 || r.Y != 20 || r.Width != 100 || r.Height != 50 {
		t.Errorf("unexpected rect: %v", r)
	}
}

func TestRectContainsRect(t *testing.T) {
	outer := NewRect(0, 0, 100, 100)
	inner := NewRect(10, 10, 50, 50)
	outside := NewRect(50, 50, 100, 100)

	if !outer.ContainsRect(inner) {
		t.Error("outer should contain inner")
	}
	if outer.ContainsRect(outside) {
		t.Error("outer should not contain outside")
	}
}

func TestRectArea(t *testing.T) {
	r := NewRect(0, 0, 10, 20)
	if r.Area() != 200 {
		t.Errorf("expected 200, got %v", r.Area())
	}
}

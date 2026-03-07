package math

import "testing"

func TestVec2Add(t *testing.T) {
	a := NewVec2(1, 2)
	b := NewVec2(3, 4)
	r := a.Add(b)
	if r.X != 4 || r.Y != 6 {
		t.Errorf("expected (4,6), got (%v,%v)", r.X, r.Y)
	}
}

func TestVec2Sub(t *testing.T) {
	a := NewVec2(5, 7)
	b := NewVec2(2, 3)
	r := a.Sub(b)
	if r.X != 3 || r.Y != 4 {
		t.Errorf("expected (3,4), got (%v,%v)", r.X, r.Y)
	}
}

func TestVec2Mul(t *testing.T) {
	v := NewVec2(3, 4)
	r := v.Mul(2)
	if r.X != 6 || r.Y != 8 {
		t.Errorf("expected (6,8), got (%v,%v)", r.X, r.Y)
	}
}

func TestVec2Dot(t *testing.T) {
	a := NewVec2(1, 0)
	b := NewVec2(0, 1)
	if a.Dot(b) != 0 {
		t.Errorf("perpendicular vectors should have dot product 0")
	}
	c := NewVec2(2, 3)
	d := NewVec2(4, 5)
	if c.Dot(d) != 23 {
		t.Errorf("expected 23, got %v", c.Dot(d))
	}
}

func TestVec2Length(t *testing.T) {
	v := NewVec2(3, 4)
	if v.Length() != 5 {
		t.Errorf("expected 5, got %v", v.Length())
	}
}

func TestVec2Normalized(t *testing.T) {
	v := NewVec2(3, 4)
	n := v.Normalized()
	if !n.Approx(NewVec2(0.6, 0.8), 1e-6) {
		t.Errorf("expected (0.6,0.8), got (%v,%v)", n.X, n.Y)
	}
}

func TestVec2NormalizedZero(t *testing.T) {
	v := Vec2Zero()
	n := v.Normalized()
	if n.X != 0 || n.Y != 0 {
		t.Errorf("normalized zero should be zero")
	}
}

func TestVec2Distance(t *testing.T) {
	a := NewVec2(0, 0)
	b := NewVec2(3, 4)
	if a.Distance(b) != 5 {
		t.Errorf("expected 5, got %v", a.Distance(b))
	}
}

func TestVec2Lerp(t *testing.T) {
	a := NewVec2(0, 0)
	b := NewVec2(10, 20)
	r := a.Lerp(b, 0.5)
	if !r.Approx(NewVec2(5, 10), 1e-6) {
		t.Errorf("expected (5,10), got (%v,%v)", r.X, r.Y)
	}
}

func TestVec2Cross(t *testing.T) {
	a := NewVec2(1, 0)
	b := NewVec2(0, 1)
	if a.Cross(b) != 1 {
		t.Errorf("expected 1, got %v", a.Cross(b))
	}
}

func TestVec2MinMax(t *testing.T) {
	a := NewVec2(1, 5)
	b := NewVec2(3, 2)
	mn := a.Min(b)
	mx := a.Max(b)
	if mn.X != 1 || mn.Y != 2 {
		t.Errorf("min: expected (1,2), got (%v,%v)", mn.X, mn.Y)
	}
	if mx.X != 3 || mx.Y != 5 {
		t.Errorf("max: expected (3,5), got (%v,%v)", mx.X, mx.Y)
	}
}

func TestVec2Neg(t *testing.T) {
	v := NewVec2(3, -4)
	n := v.Neg()
	if n.X != -3 || n.Y != 4 {
		t.Errorf("expected (-3,4), got (%v,%v)", n.X, n.Y)
	}
}

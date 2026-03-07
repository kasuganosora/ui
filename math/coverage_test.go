package math

import "testing"

// Additional tests to bring math package to 80%+ coverage.

// --- vec2 uncovered ---

func TestVec2One(t *testing.T) {
	v := Vec2One()
	if v.X != 1 || v.Y != 1 {
		t.Errorf("expected (1,1), got (%v,%v)", v.X, v.Y)
	}
}

func TestVec2UnitX(t *testing.T) {
	v := Vec2UnitX()
	if v.X != 1 || v.Y != 0 {
		t.Errorf("expected (1,0), got (%v,%v)", v.X, v.Y)
	}
}

func TestVec2UnitY(t *testing.T) {
	v := Vec2UnitY()
	if v.X != 0 || v.Y != 1 {
		t.Errorf("expected (0,1), got (%v,%v)", v.X, v.Y)
	}
}

func TestVec2Div(t *testing.T) {
	v := NewVec2(6, 8)
	r := v.Div(2)
	if r.X != 3 || r.Y != 4 {
		t.Errorf("expected (3,4), got (%v,%v)", r.X, r.Y)
	}
}

func TestVec2MulVec(t *testing.T) {
	a := NewVec2(2, 3)
	b := NewVec2(4, 5)
	r := a.MulVec(b)
	if r.X != 8 || r.Y != 15 {
		t.Errorf("expected (8,15), got (%v,%v)", r.X, r.Y)
	}
}

func TestVec2LengthSq(t *testing.T) {
	v := NewVec2(3, 4)
	if v.LengthSq() != 25 {
		t.Errorf("expected 25, got %v", v.LengthSq())
	}
}

func TestVec2DistanceSq(t *testing.T) {
	a := NewVec2(0, 0)
	b := NewVec2(3, 4)
	if a.DistanceSq(b) != 25 {
		t.Errorf("expected 25, got %v", a.DistanceSq(b))
	}
}

func TestVec2Clamp(t *testing.T) {
	v := NewVec2(15, -5)
	lo := NewVec2(0, 0)
	hi := NewVec2(10, 10)
	r := v.Clamp(lo, hi)
	if r.X != 10 || r.Y != 0 {
		t.Errorf("expected (10,0), got (%v,%v)", r.X, r.Y)
	}
}

// --- rect uncovered ---

func TestRectFromPosSize(t *testing.T) {
	r := RectFromPosSize(NewVec2(10, 20), NewVec2(30, 40))
	if r.X != 10 || r.Y != 20 || r.Width != 30 || r.Height != 40 {
		t.Errorf("unexpected rect: %v", r)
	}
}

func TestRectMinMaxSize(t *testing.T) {
	r := NewRect(10, 20, 30, 40)
	min := r.Min()
	max := r.Max()
	sz := r.Size()
	if min.X != 10 || min.Y != 20 {
		t.Errorf("min: expected (10,20), got %v", min)
	}
	if max.X != 40 || max.Y != 60 {
		t.Errorf("max: expected (40,60), got %v", max)
	}
	if sz.X != 30 || sz.Y != 40 {
		t.Errorf("size: expected (30,40), got %v", sz)
	}
}

func TestRectCorners(t *testing.T) {
	r := NewRect(10, 20, 100, 50)
	if tl := r.TopLeft(); tl.X != 10 || tl.Y != 20 {
		t.Errorf("TopLeft: %v", tl)
	}
	if tr := r.TopRight(); tr.X != 110 || tr.Y != 20 {
		t.Errorf("TopRight: %v", tr)
	}
	if bl := r.BottomLeft(); bl.X != 10 || bl.Y != 70 {
		t.Errorf("BottomLeft: %v", bl)
	}
	if br := r.BottomRight(); br.X != 110 || br.Y != 70 {
		t.Errorf("BottomRight: %v", br)
	}
}

func TestRectRightBottom(t *testing.T) {
	r := NewRect(10, 20, 100, 50)
	if r.Right() != 110 {
		t.Errorf("Right: expected 110, got %v", r.Right())
	}
	if r.Bottom() != 70 {
		t.Errorf("Bottom: expected 70, got %v", r.Bottom())
	}
}

func TestRectShrink(t *testing.T) {
	r := NewRect(5, 5, 60, 60)
	s := r.Shrink(5)
	if s.X != 10 || s.Y != 10 || s.Width != 50 || s.Height != 50 {
		t.Errorf("unexpected shrink: %v", s)
	}
}

func TestRectOffsetVec(t *testing.T) {
	r := NewRect(10, 20, 30, 40)
	moved := r.OffsetVec(NewVec2(5, -5))
	if moved.X != 15 || moved.Y != 15 {
		t.Errorf("unexpected offset: %v", moved)
	}
}

// --- color uncovered ---

func TestNewColor(t *testing.T) {
	c := NewColor(0.1, 0.2, 0.3, 0.4)
	if c.R != 0.1 || c.G != 0.2 || c.B != 0.3 || c.A != 0.4 {
		t.Errorf("unexpected color: %v", c)
	}
}

func TestRGB(t *testing.T) {
	c := RGB(0.5, 0.6, 0.7)
	if c.A != 1 {
		t.Errorf("RGB should set alpha to 1, got %v", c.A)
	}
}

func TestRGBA(t *testing.T) {
	c := RGBA(0.1, 0.2, 0.3, 0.4)
	if c.R != 0.1 || c.A != 0.4 {
		t.Errorf("unexpected RGBA: %v", c)
	}
}

func TestColorMul(t *testing.T) {
	c := RGB(0.5, 0.4, 0.3)
	r := c.Mul(2)
	if !r.Approx(Color{R: 1, G: 0.8, B: 0.6, A: 1}, 0.01) {
		t.Errorf("unexpected Mul: %v", r)
	}
}

func TestColorMulColor(t *testing.T) {
	a := RGB(0.5, 0.5, 0.5)
	b := RGB(0.5, 1, 0)
	r := a.MulColor(b)
	if !r.Approx(Color{R: 0.25, G: 0.5, B: 0, A: 1}, 0.01) {
		t.Errorf("unexpected MulColor: %v", r)
	}
}

func TestColorHexAlpha(t *testing.T) {
	c := RGB(1, 0, 0).WithAlpha(0.5)
	hex := c.Hex()
	if len(hex) != 9 { // #rrggbbaa
		t.Errorf("expected 9-char hex with alpha, got %q", hex)
	}
}

func TestColorHex4(t *testing.T) {
	c := ColorHex("#f008")
	if c.A > 0.6 || c.A < 0.4 {
		t.Errorf("expected ~0.53 alpha from '8', got %v", c.A)
	}
}

func TestColorHexInvalid(t *testing.T) {
	c := ColorHex("#12345") // 5 chars = invalid
	if c.R != 0 && c.G != 0 && c.B != 0 && c.A != 0 {
		t.Errorf("invalid hex should return zero color, got %v", c)
	}
}

func TestHexValUppercase(t *testing.T) {
	c := ColorHex("#FF0000")
	if !c.Approx(ColorRed, 0.01) {
		t.Errorf("uppercase hex should parse, got %v", c)
	}
}

// --- edges uncovered ---

func TestEdgesAll(t *testing.T) {
	e := EdgesAll(5)
	if e.Top != 5 || e.Right != 5 || e.Bottom != 5 || e.Left != 5 {
		t.Errorf("unexpected EdgesAll: %v", e)
	}
}

func TestEdgesZero(t *testing.T) {
	e := EdgesZero()
	if e.Top != 0 || e.Right != 0 || e.Bottom != 0 || e.Left != 0 {
		t.Errorf("unexpected EdgesZero: %v", e)
	}
}

func TestNewCorners(t *testing.T) {
	c := NewCorners(1, 2, 3, 4)
	if c.TopLeft != 1 || c.TopRight != 2 || c.BottomRight != 3 || c.BottomLeft != 4 {
		t.Errorf("unexpected NewCorners: %v", c)
	}
}

// --- mat3 uncovered ---

func TestMat3ScaleComponent(t *testing.T) {
	m := Mat3Scale(2, 3)
	sc := m.ScaleComponent()
	if !sc.Approx(NewVec2(2, 3), 0.01) {
		t.Errorf("expected (2,3), got %v", sc)
	}
}

func TestMat3InverseSingular(t *testing.T) {
	// Singular matrix (all zeros)
	m := Mat3{}
	inv := m.Inverse()
	// Should return identity
	id := Mat3Identity()
	if inv != id {
		t.Errorf("singular inverse should return identity, got %v", inv)
	}
}

// --- util uncovered ---

func TestPublicClamp(t *testing.T) {
	if Clamp(5, 0, 10) != 5 {
		t.Error("5 clamped to [0,10] should be 5")
	}
	if Clamp(-1, 0, 10) != 0 {
		t.Error("-1 clamped to [0,10] should be 0")
	}
	if Clamp(15, 0, 10) != 10 {
		t.Error("15 clamped to [0,10] should be 10")
	}
}

func TestPublicLerp(t *testing.T) {
	if Lerp(0, 10, 0.5) != 5 {
		t.Errorf("expected 5, got %v", Lerp(0, 10, 0.5))
	}
}

func TestPublicAbs(t *testing.T) {
	if Abs(-3) != 3 {
		t.Error("Abs(-3) should be 3")
	}
	if Abs(3) != 3 {
		t.Error("Abs(3) should be 3")
	}
}

func TestPublicMinMax(t *testing.T) {
	if Min(3, 5) != 3 {
		t.Error("Min(3,5) should be 3")
	}
	if Max(3, 5) != 5 {
		t.Error("Max(3,5) should be 5")
	}
}

func TestPublicFloorCeilRound(t *testing.T) {
	if Floor(3.7) != 3 {
		t.Errorf("Floor(3.7) = %v", Floor(3.7))
	}
	if Ceil(3.2) != 4 {
		t.Errorf("Ceil(3.2) = %v", Ceil(3.2))
	}
	if Round(3.5) != 4 {
		t.Errorf("Round(3.5) = %v", Round(3.5))
	}
	if Round(3.4) != 3 {
		t.Errorf("Round(3.4) = %v", Round(3.4))
	}
}

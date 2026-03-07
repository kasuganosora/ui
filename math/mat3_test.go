package math

import (
	"math"
	"testing"
)

func TestMat3Identity(t *testing.T) {
	m := Mat3Identity()
	p := NewVec2(3, 7)
	r := m.TransformPoint(p)
	if !r.Approx(p, 1e-6) {
		t.Errorf("identity should not change point: got %v", r)
	}
}

func TestMat3Translate(t *testing.T) {
	m := Mat3Translate(10, 20)
	p := NewVec2(5, 5)
	r := m.TransformPoint(p)
	if !r.Approx(NewVec2(15, 25), 1e-6) {
		t.Errorf("expected (15,25), got (%v,%v)", r.X, r.Y)
	}
}

func TestMat3Scale(t *testing.T) {
	m := Mat3Scale(2, 3)
	p := NewVec2(4, 5)
	r := m.TransformPoint(p)
	if !r.Approx(NewVec2(8, 15), 1e-6) {
		t.Errorf("expected (8,15), got (%v,%v)", r.X, r.Y)
	}
}

func TestMat3Rotate90(t *testing.T) {
	m := Mat3Rotate(math.Pi / 2)
	p := NewVec2(1, 0)
	r := m.TransformPoint(p)
	if !r.Approx(NewVec2(0, 1), 1e-5) {
		t.Errorf("expected (0,1), got (%v,%v)", r.X, r.Y)
	}
}

func TestMat3Mul(t *testing.T) {
	s := Mat3Scale(2, 2)
	tr := Mat3Translate(10, 0)
	m := tr.Mul(s) // first scale, then translate
	p := NewVec2(5, 0)
	r := m.TransformPoint(p)
	if !r.Approx(NewVec2(20, 0), 1e-5) {
		t.Errorf("expected (20,0), got (%v,%v)", r.X, r.Y)
	}
}

func TestMat3Inverse(t *testing.T) {
	m := Mat3Translate(10, 20).Mul(Mat3Scale(2, 3))
	inv := m.Inverse()
	product := m.Mul(inv)
	identity := Mat3Identity()
	for i := 0; i < 9; i++ {
		if abs32(product[i]-identity[i]) > 1e-4 {
			t.Errorf("M*M^-1 should be identity, got %v", product)
			break
		}
	}
}

func TestMat3Determinant(t *testing.T) {
	m := Mat3Identity()
	if abs32(m.Determinant()-1) > 1e-6 {
		t.Errorf("identity determinant should be 1, got %v", m.Determinant())
	}
	s := Mat3Scale(2, 3)
	if abs32(s.Determinant()-6) > 1e-6 {
		t.Errorf("scale(2,3) determinant should be 6, got %v", s.Determinant())
	}
}

func TestMat3TransformVec(t *testing.T) {
	m := Mat3Translate(100, 200)
	v := NewVec2(1, 0)
	r := m.TransformVec(v)
	// TransformVec should ignore translation
	if !r.Approx(NewVec2(1, 0), 1e-6) {
		t.Errorf("TransformVec should ignore translation, got %v", r)
	}
}

func TestMat3Translation(t *testing.T) {
	m := Mat3Translate(42, 99)
	tr := m.Translation()
	if !tr.Approx(NewVec2(42, 99), 1e-6) {
		t.Errorf("expected (42,99), got %v", tr)
	}
}

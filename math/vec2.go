package math

import "math"

// Vec2 is an immutable 2D vector value object.
type Vec2 struct {
	X, Y float32
}

func NewVec2(x, y float32) Vec2 {
	return Vec2{X: x, Y: y}
}

func Vec2Zero() Vec2  { return Vec2{} }
func Vec2One() Vec2   { return Vec2{X: 1, Y: 1} }
func Vec2UnitX() Vec2 { return Vec2{X: 1} }
func Vec2UnitY() Vec2 { return Vec2{Y: 1} }

func (v Vec2) Add(other Vec2) Vec2 {
	return Vec2{X: v.X + other.X, Y: v.Y + other.Y}
}

func (v Vec2) Sub(other Vec2) Vec2 {
	return Vec2{X: v.X - other.X, Y: v.Y - other.Y}
}

func (v Vec2) Mul(scalar float32) Vec2 {
	return Vec2{X: v.X * scalar, Y: v.Y * scalar}
}

func (v Vec2) Div(scalar float32) Vec2 {
	return Vec2{X: v.X / scalar, Y: v.Y / scalar}
}

func (v Vec2) MulVec(other Vec2) Vec2 {
	return Vec2{X: v.X * other.X, Y: v.Y * other.Y}
}

func (v Vec2) Dot(other Vec2) float32 {
	return v.X*other.X + v.Y*other.Y
}

func (v Vec2) Cross(other Vec2) float32 {
	return v.X*other.Y - v.Y*other.X
}

func (v Vec2) Length() float32 {
	return float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y)))
}

func (v Vec2) LengthSq() float32 {
	return v.X*v.X + v.Y*v.Y
}

func (v Vec2) Normalized() Vec2 {
	l := v.Length()
	if l == 0 {
		return Vec2Zero()
	}
	return v.Div(l)
}

func (v Vec2) Distance(other Vec2) float32 {
	return v.Sub(other).Length()
}

func (v Vec2) DistanceSq(other Vec2) float32 {
	return v.Sub(other).LengthSq()
}

func (v Vec2) Lerp(other Vec2, t float32) Vec2 {
	return Vec2{
		X: v.X + (other.X-v.X)*t,
		Y: v.Y + (other.Y-v.Y)*t,
	}
}

func (v Vec2) Min(other Vec2) Vec2 {
	return Vec2{
		X: min32(v.X, other.X),
		Y: min32(v.Y, other.Y),
	}
}

func (v Vec2) Max(other Vec2) Vec2 {
	return Vec2{
		X: max32(v.X, other.X),
		Y: max32(v.Y, other.Y),
	}
}

func (v Vec2) Clamp(lo, hi Vec2) Vec2 {
	return v.Max(lo).Min(hi)
}

func (v Vec2) Neg() Vec2 {
	return Vec2{X: -v.X, Y: -v.Y}
}

func (v Vec2) Approx(other Vec2, epsilon float32) bool {
	return abs32(v.X-other.X) < epsilon && abs32(v.Y-other.Y) < epsilon
}

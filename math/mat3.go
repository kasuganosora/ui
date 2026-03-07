package math

import "math"

// Mat3 is an immutable 3x3 matrix value object, stored in column-major order.
// Used for 2D affine transformations.
//
//	[M00 M01 M02]   [0 3 6]
//	[M10 M11 M12] = [1 4 7]
//	[M20 M21 M22]   [2 5 8]
type Mat3 [9]float32

func Mat3Identity() Mat3 {
	return Mat3{
		1, 0, 0,
		0, 1, 0,
		0, 0, 1,
	}
}

func Mat3Translate(tx, ty float32) Mat3 {
	return Mat3{
		1, 0, 0,
		0, 1, 0,
		tx, ty, 1,
	}
}

func Mat3Scale(sx, sy float32) Mat3 {
	return Mat3{
		sx, 0, 0,
		0, sy, 0,
		0, 0, 1,
	}
}

func Mat3Rotate(radians float32) Mat3 {
	c := float32(math.Cos(float64(radians)))
	s := float32(math.Sin(float64(radians)))
	return Mat3{
		c, s, 0,
		-s, c, 0,
		0, 0, 1,
	}
}

func (m Mat3) Mul(other Mat3) Mat3 {
	return Mat3{
		m[0]*other[0] + m[3]*other[1] + m[6]*other[2],
		m[1]*other[0] + m[4]*other[1] + m[7]*other[2],
		m[2]*other[0] + m[5]*other[1] + m[8]*other[2],

		m[0]*other[3] + m[3]*other[4] + m[6]*other[5],
		m[1]*other[3] + m[4]*other[4] + m[7]*other[5],
		m[2]*other[3] + m[5]*other[4] + m[8]*other[5],

		m[0]*other[6] + m[3]*other[7] + m[6]*other[8],
		m[1]*other[6] + m[4]*other[7] + m[7]*other[8],
		m[2]*other[6] + m[5]*other[7] + m[8]*other[8],
	}
}

// TransformPoint applies the affine transformation to a 2D point.
func (m Mat3) TransformPoint(p Vec2) Vec2 {
	return Vec2{
		X: m[0]*p.X + m[3]*p.Y + m[6],
		Y: m[1]*p.X + m[4]*p.Y + m[7],
	}
}

// TransformVec applies the linear part (no translation) to a 2D vector.
func (m Mat3) TransformVec(v Vec2) Vec2 {
	return Vec2{
		X: m[0]*v.X + m[3]*v.Y,
		Y: m[1]*v.X + m[4]*v.Y,
	}
}

// Determinant returns the determinant of the matrix.
func (m Mat3) Determinant() float32 {
	return m[0]*(m[4]*m[8]-m[7]*m[5]) -
		m[3]*(m[1]*m[8]-m[7]*m[2]) +
		m[6]*(m[1]*m[5]-m[4]*m[2])
}

// Inverse returns the inverse of the matrix. Returns identity if not invertible.
func (m Mat3) Inverse() Mat3 {
	det := m.Determinant()
	if abs32(det) < 1e-10 {
		return Mat3Identity()
	}
	invDet := 1.0 / det
	return Mat3{
		(m[4]*m[8] - m[5]*m[7]) * invDet,
		(m[2]*m[7] - m[1]*m[8]) * invDet,
		(m[1]*m[5] - m[2]*m[4]) * invDet,

		(m[5]*m[6] - m[3]*m[8]) * invDet,
		(m[0]*m[8] - m[2]*m[6]) * invDet,
		(m[2]*m[3] - m[0]*m[5]) * invDet,

		(m[3]*m[7] - m[4]*m[6]) * invDet,
		(m[1]*m[6] - m[0]*m[7]) * invDet,
		(m[0]*m[4] - m[1]*m[3]) * invDet,
	}
}

// Translation returns the translation component.
func (m Mat3) Translation() Vec2 {
	return Vec2{X: m[6], Y: m[7]}
}

// ScaleComponent returns the approximate scale component.
func (m Mat3) ScaleComponent() Vec2 {
	sx := Vec2{X: m[0], Y: m[1]}.Length()
	sy := Vec2{X: m[3], Y: m[4]}.Length()
	return Vec2{X: sx, Y: sy}
}

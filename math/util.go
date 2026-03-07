package math

import "math"

func min32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func max32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func abs32(a float32) float32 {
	if a < 0 {
		return -a
	}
	return a
}

func clamp32(v, lo, hi float32) float32 {
	return min32(max32(v, lo), hi)
}

func lerp32(a, b, t float32) float32 {
	return a + (b-a)*t
}

func floor32(a float32) float32 {
	return float32(math.Floor(float64(a)))
}

func ceil32(a float32) float32 {
	return float32(math.Ceil(float64(a)))
}

func round32(a float32) float32 {
	return float32(math.Round(float64(a)))
}

// Clamp clamps v to [lo, hi].
func Clamp(v, lo, hi float32) float32 {
	return clamp32(v, lo, hi)
}

// Lerp linearly interpolates between a and b by t.
func Lerp(a, b, t float32) float32 {
	return lerp32(a, b, t)
}

// Abs returns the absolute value of a.
func Abs(a float32) float32 {
	return abs32(a)
}

// Min returns the smaller of a or b.
func Min(a, b float32) float32 {
	return min32(a, b)
}

// Max returns the larger of a or b.
func Max(a, b float32) float32 {
	return max32(a, b)
}

// Floor returns the greatest integer value less than or equal to a.
func Floor(a float32) float32 {
	return floor32(a)
}

// Ceil returns the least integer value greater than or equal to a.
func Ceil(a float32) float32 {
	return ceil32(a)
}

// Round returns the nearest integer, rounding half away from zero.
func Round(a float32) float32 {
	return round32(a)
}

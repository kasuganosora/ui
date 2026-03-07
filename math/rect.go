package math

// Rect is an immutable axis-aligned rectangle value object.
// Origin is top-left corner.
type Rect struct {
	X, Y, Width, Height float32
}

func NewRect(x, y, w, h float32) Rect {
	return Rect{X: x, Y: y, Width: w, Height: h}
}

func RectFromMinMax(minX, minY, maxX, maxY float32) Rect {
	return Rect{
		X:      minX,
		Y:      minY,
		Width:  maxX - minX,
		Height: maxY - minY,
	}
}

func RectFromPosSize(pos Vec2, size Vec2) Rect {
	return Rect{X: pos.X, Y: pos.Y, Width: size.X, Height: size.Y}
}

func (r Rect) Min() Vec2 { return Vec2{X: r.X, Y: r.Y} }
func (r Rect) Max() Vec2 { return Vec2{X: r.X + r.Width, Y: r.Y + r.Height} }
func (r Rect) Size() Vec2 { return Vec2{X: r.Width, Y: r.Height} }

func (r Rect) Center() Vec2 {
	return Vec2{X: r.X + r.Width*0.5, Y: r.Y + r.Height*0.5}
}

func (r Rect) TopLeft() Vec2     { return Vec2{X: r.X, Y: r.Y} }
func (r Rect) TopRight() Vec2    { return Vec2{X: r.X + r.Width, Y: r.Y} }
func (r Rect) BottomLeft() Vec2  { return Vec2{X: r.X, Y: r.Y + r.Height} }
func (r Rect) BottomRight() Vec2 { return Vec2{X: r.X + r.Width, Y: r.Y + r.Height} }

func (r Rect) Right() float32  { return r.X + r.Width }
func (r Rect) Bottom() float32 { return r.Y + r.Height }

func (r Rect) Contains(p Vec2) bool {
	return p.X >= r.X && p.X < r.X+r.Width &&
		p.Y >= r.Y && p.Y < r.Y+r.Height
}

func (r Rect) ContainsRect(other Rect) bool {
	return other.X >= r.X && other.X+other.Width <= r.X+r.Width &&
		other.Y >= r.Y && other.Y+other.Height <= r.Y+r.Height
}

func (r Rect) Intersects(other Rect) bool {
	return r.X < other.X+other.Width && r.X+r.Width > other.X &&
		r.Y < other.Y+other.Height && r.Y+r.Height > other.Y
}

// Intersection returns the overlap area of two rectangles.
// Returns a zero Rect if they don't intersect.
func (r Rect) Intersection(other Rect) Rect {
	x1 := max32(r.X, other.X)
	y1 := max32(r.Y, other.Y)
	x2 := min32(r.X+r.Width, other.X+other.Width)
	y2 := min32(r.Y+r.Height, other.Y+other.Height)
	if x2 <= x1 || y2 <= y1 {
		return Rect{}
	}
	return RectFromMinMax(x1, y1, x2, y2)
}

// Union returns the smallest rectangle containing both rectangles.
func (r Rect) Union(other Rect) Rect {
	x1 := min32(r.X, other.X)
	y1 := min32(r.Y, other.Y)
	x2 := max32(r.X+r.Width, other.X+other.Width)
	y2 := max32(r.Y+r.Height, other.Y+other.Height)
	return RectFromMinMax(x1, y1, x2, y2)
}

// Expand returns a rect grown by the given amount on each side.
func (r Rect) Expand(amount float32) Rect {
	return Rect{
		X:      r.X - amount,
		Y:      r.Y - amount,
		Width:  r.Width + amount*2,
		Height: r.Height + amount*2,
	}
}

// Shrink returns a rect shrunk by the given amount on each side.
func (r Rect) Shrink(amount float32) Rect {
	return r.Expand(-amount)
}

// Offset returns a rect moved by the given delta.
func (r Rect) Offset(dx, dy float32) Rect {
	return Rect{X: r.X + dx, Y: r.Y + dy, Width: r.Width, Height: r.Height}
}

// OffsetVec returns a rect moved by the given vector.
func (r Rect) OffsetVec(v Vec2) Rect {
	return Rect{X: r.X + v.X, Y: r.Y + v.Y, Width: r.Width, Height: r.Height}
}

func (r Rect) IsEmpty() bool {
	return r.Width <= 0 || r.Height <= 0
}

func (r Rect) Area() float32 {
	return r.Width * r.Height
}

func (r Rect) Approx(other Rect, epsilon float32) bool {
	return abs32(r.X-other.X) < epsilon &&
		abs32(r.Y-other.Y) < epsilon &&
		abs32(r.Width-other.Width) < epsilon &&
		abs32(r.Height-other.Height) < epsilon
}

package math

// Edges represents insets/margins/padding on four sides. Immutable value object.
type Edges struct {
	Top, Right, Bottom, Left float32
}

func NewEdges(top, right, bottom, left float32) Edges {
	return Edges{Top: top, Right: right, Bottom: bottom, Left: left}
}

func EdgesAll(v float32) Edges {
	return Edges{Top: v, Right: v, Bottom: v, Left: v}
}

func EdgesSymmetric(vertical, horizontal float32) Edges {
	return Edges{Top: vertical, Right: horizontal, Bottom: vertical, Left: horizontal}
}

func EdgesZero() Edges {
	return Edges{}
}

func (e Edges) Horizontal() float32 { return e.Left + e.Right }
func (e Edges) Vertical() float32   { return e.Top + e.Bottom }

// ShrinkRect returns a rect shrunk by these edges.
func (e Edges) ShrinkRect(r Rect) Rect {
	return Rect{
		X:      r.X + e.Left,
		Y:      r.Y + e.Top,
		Width:  r.Width - e.Horizontal(),
		Height: r.Height - e.Vertical(),
	}
}

// ExpandRect returns a rect expanded by these edges.
func (e Edges) ExpandRect(r Rect) Rect {
	return Rect{
		X:      r.X - e.Left,
		Y:      r.Y - e.Top,
		Width:  r.Width + e.Horizontal(),
		Height: r.Height + e.Vertical(),
	}
}

// Corners represents corner radii. Immutable value object.
type Corners struct {
	TopLeft, TopRight, BottomRight, BottomLeft float32
}

func NewCorners(tl, tr, br, bl float32) Corners {
	return Corners{TopLeft: tl, TopRight: tr, BottomRight: br, BottomLeft: bl}
}

func CornersAll(v float32) Corners {
	return Corners{TopLeft: v, TopRight: v, BottomRight: v, BottomLeft: v}
}

func CornersZero() Corners {
	return Corners{}
}

func (c Corners) IsZero() bool {
	return c.TopLeft == 0 && c.TopRight == 0 && c.BottomRight == 0 && c.BottomLeft == 0
}

func (c Corners) IsUniform() bool {
	return c.TopLeft == c.TopRight && c.TopRight == c.BottomRight && c.BottomRight == c.BottomLeft
}

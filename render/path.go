package render

import (
	"math"
	"strings"
	"unicode"

	uimath "github.com/kasuganosora/ui/math"
)

// PathCmd identifies the type of path command.
type PathCmdType uint8

const (
	PathMoveTo  PathCmdType = iota // Move pen (no draw)
	PathLineTo                     // Straight line
	PathQuadTo                     // Quadratic bezier
	PathCubicTo                    // Cubic bezier
	PathArcTo                      // Elliptical arc
	PathClose                      // Close subpath
)

// PathCommand is a single path drawing instruction.
type PathCommand struct {
	Type PathCmdType
	// Points depend on type:
	//   MoveTo/LineTo: X1,Y1
	//   QuadTo: X1,Y1 (control), X2,Y2 (end)
	//   CubicTo: X1,Y1 (cp1), X2,Y2 (cp2), X3,Y3 (end)
	//   ArcTo: X1,Y1 (radii), X2(rotation), Y2(largeArc), X3(sweep), Y3 unused, X3,Y3 (end) — see ArcParams
	X1, Y1 float32
	X2, Y2 float32
	X3, Y3 float32
}

// ArcParams provides arc-specific fields stored in PathCommand.
type ArcParams struct {
	RX, RY   float32 // Ellipse radii
	Rotation float32 // X-axis rotation in degrees
	LargeArc bool
	Sweep    bool
	EndX     float32
	EndY     float32
}

// Path represents a 2D vector path made of subpaths.
type Path struct {
	Commands []PathCommand
	curX     float32
	curY     float32
	startX   float32
	startY   float32
}

// NewPath creates an empty path.
func NewPath() *Path {
	return &Path{}
}

// MoveTo starts a new subpath at (x, y).
func (p *Path) MoveTo(x, y float32) {
	p.Commands = append(p.Commands, PathCommand{Type: PathMoveTo, X1: x, Y1: y})
	p.curX, p.curY = x, y
	p.startX, p.startY = x, y
}

// LineTo draws a line to (x, y).
func (p *Path) LineTo(x, y float32) {
	p.Commands = append(p.Commands, PathCommand{Type: PathLineTo, X1: x, Y1: y})
	p.curX, p.curY = x, y
}

// QuadTo draws a quadratic bezier curve.
func (p *Path) QuadTo(cx, cy, x, y float32) {
	p.Commands = append(p.Commands, PathCommand{Type: PathQuadTo, X1: cx, Y1: cy, X2: x, Y2: y})
	p.curX, p.curY = x, y
}

// CubicTo draws a cubic bezier curve.
func (p *Path) CubicTo(cx1, cy1, cx2, cy2, x, y float32) {
	p.Commands = append(p.Commands, PathCommand{Type: PathCubicTo, X1: cx1, Y1: cy1, X2: cx2, Y2: cy2, X3: x, Y3: y})
	p.curX, p.curY = x, y
}

// ArcTo draws an elliptical arc.
func (p *Path) ArcTo(rx, ry, rotation float32, largeArc, sweep bool, x, y float32) {
	la := float32(0)
	if largeArc {
		la = 1
	}
	sw := float32(0)
	if sweep {
		sw = 1
	}
	p.Commands = append(p.Commands, PathCommand{
		Type: PathArcTo,
		X1: rx, Y1: ry,
		X2: rotation, Y2: la,
		X3: x,
	})
	// Store sweep and endY in a second command to avoid overloading fields
	p.Commands[len(p.Commands)-1].Y3 = y
	// We repurpose X3,Y3 for end point, and use Y2 for largeArc flag,
	// and we need sweep somewhere — let's use a convention:
	// Y2 encodes: largeArc*2 + sweep
	p.Commands[len(p.Commands)-1].Y2 = la*2 + sw
	p.curX, p.curY = x, y
}

// Close closes the current subpath.
func (p *Path) Close() {
	p.Commands = append(p.Commands, PathCommand{Type: PathClose})
	p.curX, p.curY = p.startX, p.startY
}

// Bounds computes the axis-aligned bounding box of the path.
func (p *Path) Bounds() uimath.Rect {
	if len(p.Commands) == 0 {
		return uimath.Rect{}
	}
	minX, minY := float32(1e9), float32(1e9)
	maxX, maxY := float32(-1e9), float32(-1e9)
	update := func(x, y float32) {
		if x < minX { minX = x }
		if x > maxX { maxX = x }
		if y < minY { minY = y }
		if y > maxY { maxY = y }
	}
	for _, c := range p.Commands {
		switch c.Type {
		case PathMoveTo, PathLineTo:
			update(c.X1, c.Y1)
		case PathQuadTo:
			update(c.X1, c.Y1)
			update(c.X2, c.Y2)
		case PathCubicTo:
			update(c.X1, c.Y1)
			update(c.X2, c.Y2)
			update(c.X3, c.Y3)
		case PathArcTo:
			update(c.X3, c.Y3)
		}
	}
	if minX > maxX {
		return uimath.Rect{}
	}
	return uimath.NewRect(minX, minY, maxX-minX, maxY-minY)
}

// Flatten converts curves to line segments for rasterization.
// tolerance is the max error in pixels.
func (p *Path) Flatten(tolerance float32) []uimath.Vec2 {
	var points []uimath.Vec2
	var cx, cy float32
	for _, cmd := range p.Commands {
		switch cmd.Type {
		case PathMoveTo:
			cx, cy = cmd.X1, cmd.Y1
			points = append(points, uimath.NewVec2(cx, cy))
		case PathLineTo:
			cx, cy = cmd.X1, cmd.Y1
			points = append(points, uimath.NewVec2(cx, cy))
		case PathQuadTo:
			pts := flattenQuad(cx, cy, cmd.X1, cmd.Y1, cmd.X2, cmd.Y2, tolerance)
			points = append(points, pts...)
			cx, cy = cmd.X2, cmd.Y2
		case PathCubicTo:
			pts := flattenCubic(cx, cy, cmd.X1, cmd.Y1, cmd.X2, cmd.Y2, cmd.X3, cmd.Y3, tolerance)
			points = append(points, pts...)
			cx, cy = cmd.X3, cmd.Y3
		case PathArcTo:
			pts := flattenArc(cx, cy, cmd, tolerance)
			points = append(points, pts...)
			cx, cy = cmd.X3, cmd.Y3
		case PathClose:
			// Close line back to subpath start handled by caller
		}
	}
	return points
}

func flattenQuad(x0, y0, cx, cy, x1, y1, tol float32) []uimath.Vec2 {
	// Recursive subdivision
	dx := x1 - x0
	dy := y1 - y0
	d := absF(((cx-x1)*dy - (cy-y1)*dx))
	if d < tol*tol*0.25 {
		return []uimath.Vec2{uimath.NewVec2(x1, y1)}
	}
	mx0 := (x0 + cx) * 0.5
	my0 := (y0 + cy) * 0.5
	mx1 := (cx + x1) * 0.5
	my1 := (cy + y1) * 0.5
	mx := (mx0 + mx1) * 0.5
	my := (my0 + my1) * 0.5
	pts := flattenQuad(x0, y0, mx0, my0, mx, my, tol)
	pts = append(pts, flattenQuad(mx, my, mx1, my1, x1, y1, tol)...)
	return pts
}

func flattenCubic(x0, y0, cx1, cy1, cx2, cy2, x1, y1, tol float32) []uimath.Vec2 {
	dx := x1 - x0
	dy := y1 - y0
	d1 := absF((cx1-x1)*dy - (cy1-y1)*dx)
	d2 := absF((cx2-x1)*dy - (cy2-y1)*dx)
	if (d1+d2)*(d1+d2) < tol*tol*(dx*dx+dy*dy)*0.25 {
		return []uimath.Vec2{uimath.NewVec2(x1, y1)}
	}
	// de Casteljau subdivision at t=0.5
	m01x := (x0 + cx1) * 0.5
	m01y := (y0 + cy1) * 0.5
	m12x := (cx1 + cx2) * 0.5
	m12y := (cy1 + cy2) * 0.5
	m23x := (cx2 + x1) * 0.5
	m23y := (cy2 + y1) * 0.5
	m012x := (m01x + m12x) * 0.5
	m012y := (m01y + m12y) * 0.5
	m123x := (m12x + m23x) * 0.5
	m123y := (m12y + m23y) * 0.5
	mx := (m012x + m123x) * 0.5
	my := (m012y + m123y) * 0.5
	pts := flattenCubic(x0, y0, m01x, m01y, m012x, m012y, mx, my, tol)
	pts = append(pts, flattenCubic(mx, my, m123x, m123y, m23x, m23y, x1, y1, tol)...)
	return pts
}

func flattenArc(cx, cy float32, cmd PathCommand, tol float32) []uimath.Vec2 {
	rx, ry := cmd.X1, cmd.Y1
	endX, endY := cmd.X3, cmd.Y3
	if rx <= 0 || ry <= 0 {
		return []uimath.Vec2{uimath.NewVec2(endX, endY)}
	}
	// Approximate arc with line segments
	dx := endX - cx
	dy := endY - cy
	dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	steps := int(dist/tol) + 8
	if steps > 100 {
		steps = 100
	}
	pts := make([]uimath.Vec2, 0, steps)
	for i := 1; i <= steps; i++ {
		t := float32(i) / float32(steps)
		x := cx + dx*t
		y := cy + dy*t
		pts = append(pts, uimath.NewVec2(x, y))
	}
	return pts
}

func absF(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

// ParseSVGPath parses an SVG path data string (d attribute) into a Path.
func ParseSVGPath(d string) *Path {
	p := NewPath()
	tokens := tokenizeSVGPath(d)
	i := 0
	var curCmd byte
	var cx, cy float32 // current point
	var sx, sy float32 // subpath start

	nextFloat := func() float32 {
		if i >= len(tokens) {
			return 0
		}
		v := parseFloat32(tokens[i])
		i++
		return v
	}

	for i < len(tokens) {
		tok := tokens[i]
		if len(tok) == 1 && isAlpha(tok[0]) {
			curCmd = tok[0]
			i++
		}

		switch curCmd {
		case 'M':
			x, y := nextFloat(), nextFloat()
			p.MoveTo(x, y)
			cx, cy = x, y
			sx, sy = x, y
			curCmd = 'L' // subsequent coords are LineTo
		case 'm':
			dx, dy := nextFloat(), nextFloat()
			x, y := cx+dx, cy+dy
			p.MoveTo(x, y)
			cx, cy = x, y
			sx, sy = x, y
			curCmd = 'l'
		case 'L':
			x, y := nextFloat(), nextFloat()
			p.LineTo(x, y)
			cx, cy = x, y
		case 'l':
			dx, dy := nextFloat(), nextFloat()
			x, y := cx+dx, cy+dy
			p.LineTo(x, y)
			cx, cy = x, y
		case 'H':
			x := nextFloat()
			p.LineTo(x, cy)
			cx = x
		case 'h':
			dx := nextFloat()
			p.LineTo(cx+dx, cy)
			cx += dx
		case 'V':
			y := nextFloat()
			p.LineTo(cx, y)
			cy = y
		case 'v':
			dy := nextFloat()
			p.LineTo(cx, cy+dy)
			cy += dy
		case 'Q':
			cpx, cpy := nextFloat(), nextFloat()
			x, y := nextFloat(), nextFloat()
			p.QuadTo(cpx, cpy, x, y)
			cx, cy = x, y
		case 'q':
			dcpx, dcpy := nextFloat(), nextFloat()
			dx, dy := nextFloat(), nextFloat()
			p.QuadTo(cx+dcpx, cy+dcpy, cx+dx, cy+dy)
			cx, cy = cx+dx, cy+dy
		case 'C':
			cp1x, cp1y := nextFloat(), nextFloat()
			cp2x, cp2y := nextFloat(), nextFloat()
			x, y := nextFloat(), nextFloat()
			p.CubicTo(cp1x, cp1y, cp2x, cp2y, x, y)
			cx, cy = x, y
		case 'c':
			d1x, d1y := nextFloat(), nextFloat()
			d2x, d2y := nextFloat(), nextFloat()
			dx, dy := nextFloat(), nextFloat()
			p.CubicTo(cx+d1x, cy+d1y, cx+d2x, cy+d2y, cx+dx, cy+dy)
			cx, cy = cx+dx, cy+dy
		case 'S':
			cp2x, cp2y := nextFloat(), nextFloat()
			x, y := nextFloat(), nextFloat()
			// Reflect previous control point
			p.CubicTo(cx, cy, cp2x, cp2y, x, y)
			cx, cy = x, y
		case 's':
			d2x, d2y := nextFloat(), nextFloat()
			dx, dy := nextFloat(), nextFloat()
			p.CubicTo(cx, cy, cx+d2x, cy+d2y, cx+dx, cy+dy)
			cx, cy = cx+dx, cy+dy
		case 'T':
			x, y := nextFloat(), nextFloat()
			p.QuadTo(cx, cy, x, y)
			cx, cy = x, y
		case 't':
			dx, dy := nextFloat(), nextFloat()
			p.QuadTo(cx, cy, cx+dx, cy+dy)
			cx, cy = cx+dx, cy+dy
		case 'A':
			rx, ry := nextFloat(), nextFloat()
			rot := nextFloat()
			la := nextFloat()
			sw := nextFloat()
			x, y := nextFloat(), nextFloat()
			p.ArcTo(rx, ry, rot, la != 0, sw != 0, x, y)
			cx, cy = x, y
		case 'a':
			rx, ry := nextFloat(), nextFloat()
			rot := nextFloat()
			la := nextFloat()
			sw := nextFloat()
			dx, dy := nextFloat(), nextFloat()
			p.ArcTo(rx, ry, rot, la != 0, sw != 0, cx+dx, cy+dy)
			cx, cy = cx+dx, cy+dy
		case 'Z', 'z':
			p.Close()
			cx, cy = sx, sy
		default:
			i++ // skip unknown
		}
	}
	return p
}

func isAlpha(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

// tokenizeSVGPath splits an SVG path data string into command letters and numbers.
func tokenizeSVGPath(d string) []string {
	var tokens []string
	d = strings.TrimSpace(d)
	i := 0
	for i < len(d) {
		c := d[i]
		if c == ',' || c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			i++
			continue
		}
		if isAlpha(c) {
			tokens = append(tokens, string(c))
			i++
			continue
		}
		// Number: optional sign, digits, optional decimal, optional exponent
		start := i
		if c == '-' || c == '+' {
			i++
		}
		for i < len(d) && (d[i] >= '0' && d[i] <= '9') {
			i++
		}
		if i < len(d) && d[i] == '.' {
			i++
			for i < len(d) && (d[i] >= '0' && d[i] <= '9') {
				i++
			}
		}
		if i < len(d) && (d[i] == 'e' || d[i] == 'E') {
			i++
			if i < len(d) && (d[i] == '+' || d[i] == '-') {
				i++
			}
			for i < len(d) && (d[i] >= '0' && d[i] <= '9') {
				i++
			}
		}
		if i > start {
			tokens = append(tokens, d[start:i])
		} else {
			i++ // skip unrecognized character
		}
	}
	return tokens
}

func parseFloat32(s string) float32 {
	// Simple float parser
	neg := false
	i := 0
	if i < len(s) && s[i] == '-' {
		neg = true
		i++
	} else if i < len(s) && s[i] == '+' {
		i++
	}
	var intPart float64
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		intPart = intPart*10 + float64(s[i]-'0')
		i++
	}
	var fracPart float64
	if i < len(s) && s[i] == '.' {
		i++
		scale := 0.1
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			fracPart += float64(s[i]-'0') * scale
			scale *= 0.1
			i++
		}
	}
	v := intPart + fracPart
	// Exponent
	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		i++
		expNeg := false
		if i < len(s) && s[i] == '-' {
			expNeg = true
			i++
		} else if i < len(s) && s[i] == '+' {
			i++
		}
		var exp float64
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			exp = exp*10 + float64(s[i]-'0')
			i++
		}
		if expNeg {
			v /= math.Pow(10, exp)
		} else {
			v *= math.Pow(10, exp)
		}
	}
	if neg {
		v = -v
	}
	return float32(v)
}

// DrawPath renders a path as filled/stroked rectangles (line segments).
// This is a software rasterizer fallback that emits thin rect commands.
func (cb *CommandBuffer) DrawPath(path *Path, strokeColor uimath.Color, strokeWidth float32, fillColor uimath.Color, zOrder int32, opacity float32) {
	points := path.Flatten(1.0)
	if len(points) < 2 && fillColor.IsTransparent() {
		return
	}

	// Fill: approximate with bounding rect (real fill requires scanline raster)
	if !fillColor.IsTransparent() && len(points) > 2 {
		bounds := path.Bounds()
		cb.DrawRect(RectCmd{
			Bounds:    bounds,
			FillColor: fillColor,
		}, zOrder, opacity)
	}

	// Stroke: emit thin rectangles for each segment
	if !strokeColor.IsTransparent() && strokeWidth > 0 {
		for i := 0; i < len(points)-1; i++ {
			a, b := points[i], points[i+1]
			drawLineSegment(cb, a.X, a.Y, b.X, b.Y, strokeWidth, strokeColor, zOrder+1, opacity)
		}
	}
}

func drawLineSegment(cb *CommandBuffer, x1, y1, x2, y2, width float32, color uimath.Color, z int32, opacity float32) {
	dx := x2 - x1
	dy := y2 - y1
	length := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	if length < 0.001 {
		return
	}
	// For axis-aligned or near-axis lines, use a simple rect
	// For angled lines, approximate with a thin rect at the midpoint
	if absF(dy) < 0.5 {
		// Horizontal-ish
		minX := x1
		if x2 < minX { minX = x2 }
		cb.DrawRect(RectCmd{
			Bounds:    uimath.NewRect(minX, y1-width/2, absF(dx), width),
			FillColor: color,
		}, z, opacity)
	} else if absF(dx) < 0.5 {
		// Vertical-ish
		minY := y1
		if y2 < minY { minY = y2 }
		cb.DrawRect(RectCmd{
			Bounds:    uimath.NewRect(x1-width/2, minY, width, absF(dy)),
			FillColor: color,
		}, z, opacity)
	} else {
		// Diagonal: emit a small rect at the midpoint
		// This is an approximation; real line drawing needs a proper rasterizer
		midX := (x1 + x2) / 2
		midY := (y1 + y2) / 2
		cb.DrawRect(RectCmd{
			Bounds:    uimath.NewRect(midX-length/2, midY-width/2, length, width),
			FillColor: color,
		}, z, opacity)
	}
}

// Suppress unused import warning
var _ = unicode.IsLetter

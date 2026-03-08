package render

import (
	"math"
	"strings"

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
// Fields: X1=rx, Y1=ry, X2=rotation, Y2=flags (largeArc*2+sweep), X3=endX, Y3=endY.
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
		X1:   rx, Y1: ry,
		X2:   rotation, Y2: la*2 + sw,
		X3:   x, Y3: y,
	})
	p.curX, p.curY = x, y
}

// Close closes the current subpath.
func (p *Path) Close() {
	p.Commands = append(p.Commands, PathCommand{Type: PathClose})
	p.curX, p.curY = p.startX, p.startY
}

// Bounds computes the axis-aligned bounding box of the path,
// accounting for bezier curve extrema.
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
	var cx, cy float32
	for _, c := range p.Commands {
		switch c.Type {
		case PathMoveTo:
			update(c.X1, c.Y1)
			cx, cy = c.X1, c.Y1
		case PathLineTo:
			update(c.X1, c.Y1)
			cx, cy = c.X1, c.Y1
		case PathQuadTo:
			// Check extrema: derivative is linear, solve for t where d/dt = 0
			// B'(t) = 2(1-t)(cp-p0) + 2t(p1-cp) = 0 → t = (p0-cp)/(p0-2cp+p1)
			update(c.X2, c.Y2) // endpoint
			quadExtrema := func(p0, cp, p1 float32) float32 {
				denom := p0 - 2*cp + p1
				if absF(denom) < 1e-6 {
					return -1
				}
				return (p0 - cp) / denom
			}
			if t := quadExtrema(cx, c.X1, c.X2); t > 0 && t < 1 {
				x := (1-t)*(1-t)*cx + 2*(1-t)*t*c.X1 + t*t*c.X2
				update(x, 0)
				update(x, maxY) // only x matters
			}
			if t := quadExtrema(cy, c.Y1, c.Y2); t > 0 && t < 1 {
				y := (1-t)*(1-t)*cy + 2*(1-t)*t*c.Y1 + t*t*c.Y2
				update(0, y)
				update(maxX, y) // only y matters
			}
			cx, cy = c.X2, c.Y2
		case PathCubicTo:
			update(c.X3, c.Y3) // endpoint
			// Cubic extrema: solve quadratic 3(1-t)²(c1-p0) + 6(1-t)t(c2-c1) + 3t²(p1-c2) = 0
			cubicExtremaT := func(p0, c1, c2, p1 float32) (float32, float32) {
				a := -p0 + 3*c1 - 3*c2 + p1
				b := 2*p0 - 4*c1 + 2*c2
				c := -p0 + c1
				if absF(a) < 1e-6 {
					if absF(b) < 1e-6 {
						return -1, -1
					}
					return -c / b, -1
				}
				disc := b*b - 4*a*c
				if disc < 0 {
					return -1, -1
				}
				sq := float32(math.Sqrt(float64(disc)))
				return (-b + sq) / (2 * a), (-b - sq) / (2 * a)
			}
			evalCubic := func(p0, c1, c2, p1, t float32) float32 {
				u := 1 - t
				return u*u*u*p0 + 3*u*u*t*c1 + 3*u*t*t*c2 + t*t*t*p1
			}
			t1, t2 := cubicExtremaT(cx, c.X1, c.X2, c.X3)
			if t1 > 0 && t1 < 1 {
				update(evalCubic(cx, c.X1, c.X2, c.X3, t1), minY)
			}
			if t2 > 0 && t2 < 1 {
				update(evalCubic(cx, c.X1, c.X2, c.X3, t2), minY)
			}
			t1, t2 = cubicExtremaT(cy, c.Y1, c.Y2, c.Y3)
			if t1 > 0 && t1 < 1 {
				update(minX, evalCubic(cy, c.Y1, c.Y2, c.Y3, t1))
			}
			if t2 > 0 && t2 < 1 {
				update(minX, evalCubic(cy, c.Y1, c.Y2, c.Y3, t2))
			}
			cx, cy = c.X3, c.Y3
		case PathArcTo:
			update(c.X3, c.Y3)
			cx, cy = c.X3, c.Y3
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
	var sx, sy float32 // subpath start
	for _, cmd := range p.Commands {
		switch cmd.Type {
		case PathMoveTo:
			cx, cy = cmd.X1, cmd.Y1
			sx, sy = cx, cy
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
			// Close line back to subpath start
			if cx != sx || cy != sy {
				points = append(points, uimath.NewVec2(sx, sy))
			}
			cx, cy = sx, sy
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
	rx, ry := float64(cmd.X1), float64(cmd.Y1)
	endX, endY := float64(cmd.X3), float64(cmd.Y3)
	if rx <= 0 || ry <= 0 {
		return []uimath.Vec2{uimath.NewVec2(cmd.X3, cmd.Y3)}
	}

	// Decode flags from Y2: largeArc*2 + sweep
	flags := cmd.Y2
	largeArc := flags >= 2
	sweep := int(flags)%2 == 1
	phi := float64(cmd.X2) * math.Pi / 180 // rotation in radians

	// SVG endpoint-to-center arc conversion (https://www.w3.org/TR/SVG/implnote.html#ArcConversionEndpointToCenter)
	x1, y1 := float64(cx), float64(cy)
	x2, y2 := endX, endY

	cosPhi := math.Cos(phi)
	sinPhi := math.Sin(phi)

	dx2 := (x1 - x2) / 2
	dy2 := (y1 - y2) / 2
	x1p := cosPhi*dx2 + sinPhi*dy2
	y1p := -sinPhi*dx2 + cosPhi*dy2

	// Ensure radii are large enough
	x1pSq := x1p * x1p
	y1pSq := y1p * y1p
	rxSq := rx * rx
	rySq := ry * ry
	lambda := x1pSq/rxSq + y1pSq/rySq
	if lambda > 1 {
		s := math.Sqrt(lambda)
		rx *= s
		ry *= s
		rxSq = rx * rx
		rySq = ry * ry
	}

	// Compute center point
	num := rxSq*rySq - rxSq*y1pSq - rySq*x1pSq
	den := rxSq*y1pSq + rySq*x1pSq
	sq := 0.0
	if den > 0 {
		sq = math.Sqrt(math.Abs(num / den))
	}
	if largeArc == sweep {
		sq = -sq
	}
	cxp := sq * rx * y1p / ry
	cyp := -sq * ry * x1p / rx

	centerX := cosPhi*cxp - sinPhi*cyp + (x1+x2)/2
	centerY := sinPhi*cxp + cosPhi*cyp + (y1+y2)/2

	// Compute start/end angles
	angleVec := func(ux, uy, vx, vy float64) float64 {
		n := math.Sqrt((ux*ux+uy*uy) * (vx*vx + vy*vy))
		if n == 0 {
			return 0
		}
		c := (ux*vx + uy*vy) / n
		if c > 1 {
			c = 1
		} else if c < -1 {
			c = -1
		}
		a := math.Acos(c)
		if ux*vy-uy*vx < 0 {
			a = -a
		}
		return a
	}

	theta1 := angleVec(1, 0, (x1p-cxp)/rx, (y1p-cyp)/ry)
	dtheta := angleVec((x1p-cxp)/rx, (y1p-cyp)/ry, (-x1p-cxp)/rx, (-y1p-cyp)/ry)
	if !sweep && dtheta > 0 {
		dtheta -= 2 * math.Pi
	} else if sweep && dtheta < 0 {
		dtheta += 2 * math.Pi
	}

	// Number of segments based on arc length approximation
	steps := int(math.Abs(dtheta)/(math.Pi/8)) + 4
	if steps > 100 {
		steps = 100
	}

	pts := make([]uimath.Vec2, 0, steps)
	for i := 1; i <= steps; i++ {
		t := float64(i) / float64(steps)
		angle := theta1 + dtheta*t
		xp := rx*math.Cos(angle)
		yp := ry*math.Sin(angle)
		px := cosPhi*xp - sinPhi*yp + centerX
		py := sinPhi*xp + cosPhi*yp + centerY
		pts = append(pts, uimath.NewVec2(float32(px), float32(py)))
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
	var cx, cy float32     // current point
	var sx, sy float32     // subpath start
	var lastCp2x, lastCp2y float32 // last cubic control point (for S/s)
	var lastCpx, lastCpy float32   // last quadratic control point (for T/t)
	var lastWasCubic, lastWasQuad bool

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
			lastWasCubic, lastWasQuad = false, false
			curCmd = 'L' // subsequent coords are LineTo
		case 'm':
			dx, dy := nextFloat(), nextFloat()
			x, y := cx+dx, cy+dy
			p.MoveTo(x, y)
			cx, cy = x, y
			sx, sy = x, y
			lastWasCubic, lastWasQuad = false, false
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
			lastCpx, lastCpy = cpx, cpy
			lastWasQuad = true
			lastWasCubic = false
			cx, cy = x, y
		case 'q':
			dcpx, dcpy := nextFloat(), nextFloat()
			dx, dy := nextFloat(), nextFloat()
			qcpx, qcpy := cx+dcpx, cy+dcpy
			p.QuadTo(qcpx, qcpy, cx+dx, cy+dy)
			lastCpx, lastCpy = qcpx, qcpy
			lastWasQuad = true
			lastWasCubic = false
			cx, cy = cx+dx, cy+dy
		case 'C':
			cp1x, cp1y := nextFloat(), nextFloat()
			cp2x, cp2y := nextFloat(), nextFloat()
			x, y := nextFloat(), nextFloat()
			p.CubicTo(cp1x, cp1y, cp2x, cp2y, x, y)
			lastCp2x, lastCp2y = cp2x, cp2y
			lastWasCubic = true
			lastWasQuad = false
			cx, cy = x, y
		case 'c':
			d1x, d1y := nextFloat(), nextFloat()
			d2x, d2y := nextFloat(), nextFloat()
			dx, dy := nextFloat(), nextFloat()
			acp2x, acp2y := cx+d2x, cy+d2y
			p.CubicTo(cx+d1x, cy+d1y, acp2x, acp2y, cx+dx, cy+dy)
			lastCp2x, lastCp2y = acp2x, acp2y
			lastWasCubic = true
			lastWasQuad = false
			cx, cy = cx+dx, cy+dy
		case 'S':
			cp2x, cp2y := nextFloat(), nextFloat()
			x, y := nextFloat(), nextFloat()
			// Reflect previous cp2, or use current point if last cmd wasn't cubic
			cp1x, cp1y := cx, cy
			if lastWasCubic {
				cp1x = 2*cx - lastCp2x
				cp1y = 2*cy - lastCp2y
			}
			p.CubicTo(cp1x, cp1y, cp2x, cp2y, x, y)
			lastCp2x, lastCp2y = cp2x, cp2y
			lastWasCubic = true
			lastWasQuad = false
			cx, cy = x, y
		case 's':
			d2x, d2y := nextFloat(), nextFloat()
			dx, dy := nextFloat(), nextFloat()
			cp1x, cp1y := cx, cy
			if lastWasCubic {
				cp1x = 2*cx - lastCp2x
				cp1y = 2*cy - lastCp2y
			}
			acp2x, acp2y := cx+d2x, cy+d2y
			p.CubicTo(cp1x, cp1y, acp2x, acp2y, cx+dx, cy+dy)
			lastCp2x, lastCp2y = acp2x, acp2y
			lastWasCubic = true
			lastWasQuad = false
			cx, cy = cx+dx, cy+dy
		case 'T':
			x, y := nextFloat(), nextFloat()
			// Reflect previous quadratic cp, or use current point
			cpx, cpy := cx, cy
			if lastWasQuad {
				cpx = 2*cx - lastCpx
				cpy = 2*cy - lastCpy
			}
			p.QuadTo(cpx, cpy, x, y)
			lastCpx, lastCpy = cpx, cpy
			lastWasQuad = true
			lastWasCubic = false
			cx, cy = x, y
		case 't':
			dx, dy := nextFloat(), nextFloat()
			cpx, cpy := cx, cy
			if lastWasQuad {
				cpx = 2*cx - lastCpx
				cpy = 2*cy - lastCpy
			}
			p.QuadTo(cpx, cpy, cx+dx, cy+dy)
			lastCpx, lastCpy = cpx, cpy
			lastWasQuad = true
			lastWasCubic = false
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
		// In SVG, '-' and '.' can act as separators between numbers (e.g., "100-50" = "100" "-50", "1.5.3" = "1.5" ".3")
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

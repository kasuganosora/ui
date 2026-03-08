package render

import (
	"testing"

	uimath "github.com/kasuganosora/ui/math"
)

func TestPathMoveTo(t *testing.T) {
	p := NewPath()
	p.MoveTo(10, 20)
	if len(p.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(p.Commands))
	}
	if p.Commands[0].Type != PathMoveTo {
		t.Error("expected MoveTo")
	}
}

func TestPathLineTo(t *testing.T) {
	p := NewPath()
	p.MoveTo(0, 0)
	p.LineTo(100, 0)
	p.LineTo(100, 100)
	p.Close()
	if len(p.Commands) != 4 {
		t.Fatalf("expected 4 commands, got %d", len(p.Commands))
	}
}

func TestPathBounds(t *testing.T) {
	p := NewPath()
	p.MoveTo(10, 20)
	p.LineTo(50, 80)
	p.LineTo(30, 10)
	b := p.Bounds()
	if b.X != 10 || b.Y != 10 || b.Width != 40 || b.Height != 70 {
		t.Errorf("unexpected bounds: %+v", b)
	}
}

func TestPathBoundsQuad(t *testing.T) {
	p := NewPath()
	p.MoveTo(0, 0)
	p.QuadTo(50, 100, 100, 0)
	b := p.Bounds()
	if b.IsEmpty() {
		t.Error("expected non-empty bounds for quad path")
	}
	// Quadratic bezier (0,0)→cp(50,100)→(100,0): curve peak at y=50, not at control point
	if b.Width < 100 || b.Height < 49 {
		t.Errorf("bounds too small for quad: %+v", b)
	}
}

func TestPathBoundsCubic(t *testing.T) {
	p := NewPath()
	p.MoveTo(0, 0)
	p.CubicTo(10, 50, 90, 50, 100, 0)
	b := p.Bounds()
	if b.IsEmpty() {
		t.Error("expected non-empty bounds for cubic path")
	}
}

func TestPathBoundsArc(t *testing.T) {
	p := NewPath()
	p.MoveTo(0, 0)
	p.ArcTo(25, 25, 0, false, true, 50, 50)
	b := p.Bounds()
	if b.IsEmpty() {
		t.Error("expected non-empty bounds for arc path")
	}
}

func TestPathFlatten(t *testing.T) {
	p := NewPath()
	p.MoveTo(0, 0)
	p.LineTo(100, 0)
	p.LineTo(100, 100)
	pts := p.Flatten(1.0)
	if len(pts) < 3 {
		t.Errorf("expected at least 3 points, got %d", len(pts))
	}
}

func TestPathFlattenQuad(t *testing.T) {
	p := NewPath()
	p.MoveTo(0, 0)
	p.QuadTo(50, 100, 100, 0)
	pts := p.Flatten(1.0)
	if len(pts) < 4 {
		t.Errorf("expected several points from quad flatten, got %d", len(pts))
	}
}

func TestPathFlattenCubic(t *testing.T) {
	p := NewPath()
	p.MoveTo(0, 0)
	p.CubicTo(33, 100, 66, -50, 100, 0)
	pts := p.Flatten(1.0)
	if len(pts) < 4 {
		t.Errorf("expected several points from cubic flatten, got %d", len(pts))
	}
}

func TestPathFlattenArc(t *testing.T) {
	p := NewPath()
	p.MoveTo(10, 80)
	p.ArcTo(25, 25, 0, false, true, 50, 80)
	pts := p.Flatten(1.0)
	if len(pts) < 2 {
		t.Errorf("expected at least 2 points from arc flatten, got %d", len(pts))
	}
	// Last point should be at arc endpoint
	last := pts[len(pts)-1]
	if last.X != 50 || last.Y != 80 {
		t.Errorf("arc endpoint: got (%g,%g), want (50,80)", last.X, last.Y)
	}
}

func TestPathFlattenArcZeroRadius(t *testing.T) {
	p := NewPath()
	p.MoveTo(0, 0)
	// Zero radius arc should just produce endpoint
	p.ArcTo(0, 0, 0, false, false, 50, 50)
	pts := p.Flatten(1.0)
	// Should have start + 1 point for degenerate arc
	found := false
	for _, pt := range pts {
		if pt.X == 50 && pt.Y == 50 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected arc endpoint in flattened points")
	}
}

func TestPathFlattenClose(t *testing.T) {
	p := NewPath()
	p.MoveTo(0, 0)
	p.LineTo(100, 0)
	p.Close()
	pts := p.Flatten(1.0)
	if len(pts) < 2 {
		t.Errorf("expected at least 2 points, got %d", len(pts))
	}
}

func TestParseSVGPathSimple(t *testing.T) {
	path := ParseSVGPath("M 10 20 L 50 80 L 30 10 Z")
	if len(path.Commands) != 4 {
		t.Fatalf("expected 4 commands, got %d", len(path.Commands))
	}
	if path.Commands[0].Type != PathMoveTo {
		t.Error("expected MoveTo first")
	}
	if path.Commands[3].Type != PathClose {
		t.Error("expected Close last")
	}
}

func TestParseSVGPathRelative(t *testing.T) {
	path := ParseSVGPath("m 0 0 l 100 0 l 0 100 z")
	if len(path.Commands) != 4 {
		t.Fatalf("expected 4 commands, got %d", len(path.Commands))
	}
	if path.Commands[1].X1 != 100 || path.Commands[1].Y1 != 0 {
		t.Errorf("expected (100,0), got (%g,%g)", path.Commands[1].X1, path.Commands[1].Y1)
	}
}

func TestParseSVGPathHV(t *testing.T) {
	path := ParseSVGPath("M 0 0 H 100 V 50")
	if len(path.Commands) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(path.Commands))
	}
	if path.Commands[1].X1 != 100 || path.Commands[1].Y1 != 0 {
		t.Errorf("H: expected (100,0), got (%g,%g)", path.Commands[1].X1, path.Commands[1].Y1)
	}
	if path.Commands[2].X1 != 100 || path.Commands[2].Y1 != 50 {
		t.Errorf("V: expected (100,50), got (%g,%g)", path.Commands[2].X1, path.Commands[2].Y1)
	}
}

func TestParseSVGPathRelativeHV(t *testing.T) {
	path := ParseSVGPath("M 10 20 h 50 v 30")
	if len(path.Commands) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(path.Commands))
	}
	// h 50 from (10,20) → LineTo(60, 20)
	if path.Commands[1].X1 != 60 || path.Commands[1].Y1 != 20 {
		t.Errorf("h: expected (60,20), got (%g,%g)", path.Commands[1].X1, path.Commands[1].Y1)
	}
	// v 30 from (60,20) → LineTo(60, 50)
	if path.Commands[2].X1 != 60 || path.Commands[2].Y1 != 50 {
		t.Errorf("v: expected (60,50), got (%g,%g)", path.Commands[2].X1, path.Commands[2].Y1)
	}
}

func TestParseSVGPathCurve(t *testing.T) {
	path := ParseSVGPath("M 0 0 C 10 20 30 40 50 60")
	if len(path.Commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(path.Commands))
	}
	c := path.Commands[1]
	if c.Type != PathCubicTo {
		t.Error("expected CubicTo")
	}
	if c.X3 != 50 || c.Y3 != 60 {
		t.Errorf("expected end (50,60), got (%g,%g)", c.X3, c.Y3)
	}
}

func TestParseSVGPathRelativeCubic(t *testing.T) {
	path := ParseSVGPath("M 10 10 c 10 20 30 40 50 60")
	if len(path.Commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(path.Commands))
	}
	c := path.Commands[1]
	if c.Type != PathCubicTo {
		t.Error("expected CubicTo")
	}
	// Relative: cp1=(20,30), cp2=(40,50), end=(60,70)
	if c.X3 != 60 || c.Y3 != 70 {
		t.Errorf("expected end (60,70), got (%g,%g)", c.X3, c.Y3)
	}
}

func TestParseSVGPathQuad(t *testing.T) {
	path := ParseSVGPath("M 0 0 Q 50 100 100 0")
	if len(path.Commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(path.Commands))
	}
	if path.Commands[1].Type != PathQuadTo {
		t.Error("expected QuadTo")
	}
}

func TestParseSVGPathRelativeQuad(t *testing.T) {
	path := ParseSVGPath("M 10 10 q 40 90 90 -10")
	if len(path.Commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(path.Commands))
	}
	q := path.Commands[1]
	if q.Type != PathQuadTo {
		t.Error("expected QuadTo")
	}
	// Endpoint: (10+90, 10-10) = (100, 0)
	if q.X2 != 100 || q.Y2 != 0 {
		t.Errorf("expected end (100,0), got (%g,%g)", q.X2, q.Y2)
	}
}

func TestParseSVGPathSmoothCubic(t *testing.T) {
	path := ParseSVGPath("M 0 0 S 30 40 50 60")
	if len(path.Commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(path.Commands))
	}
	c := path.Commands[1]
	if c.Type != PathCubicTo {
		t.Error("expected CubicTo from S command")
	}
	if c.X3 != 50 || c.Y3 != 60 {
		t.Errorf("S end: expected (50,60), got (%g,%g)", c.X3, c.Y3)
	}
}

func TestParseSVGPathRelativeSmoothCubic(t *testing.T) {
	path := ParseSVGPath("M 10 10 s 20 30 40 50")
	if len(path.Commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(path.Commands))
	}
	c := path.Commands[1]
	if c.Type != PathCubicTo {
		t.Error("expected CubicTo from s command")
	}
	// end = (10+40, 10+50) = (50, 60)
	if c.X3 != 50 || c.Y3 != 60 {
		t.Errorf("s end: expected (50,60), got (%g,%g)", c.X3, c.Y3)
	}
}

func TestParseSVGPathSmoothQuad(t *testing.T) {
	path := ParseSVGPath("M 0 0 T 50 60")
	if len(path.Commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(path.Commands))
	}
	q := path.Commands[1]
	if q.Type != PathQuadTo {
		t.Error("expected QuadTo from T command")
	}
}

func TestParseSVGPathRelativeSmoothQuad(t *testing.T) {
	path := ParseSVGPath("M 10 10 t 40 50")
	if len(path.Commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(path.Commands))
	}
	q := path.Commands[1]
	if q.Type != PathQuadTo {
		t.Error("expected QuadTo from t command")
	}
	// end = (10+40, 10+50) = (50, 60)
	if q.X2 != 50 || q.Y2 != 60 {
		t.Errorf("t end: expected (50,60), got (%g,%g)", q.X2, q.Y2)
	}
}

func TestParseSVGPathArc(t *testing.T) {
	path := ParseSVGPath("M 10 80 A 25 25 0 0 1 50 80")
	if len(path.Commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(path.Commands))
	}
	if path.Commands[1].Type != PathArcTo {
		t.Error("expected ArcTo")
	}
}

func TestParseSVGPathRelativeArc(t *testing.T) {
	path := ParseSVGPath("M 10 80 a 25 25 0 1 0 40 0")
	if len(path.Commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(path.Commands))
	}
	a := path.Commands[1]
	if a.Type != PathArcTo {
		t.Error("expected ArcTo from 'a' command")
	}
	// end = (10+40, 80+0) = (50, 80)
	if a.X3 != 50 || a.Y3 != 80 {
		t.Errorf("a end: expected (50,80), got (%g,%g)", a.X3, a.Y3)
	}
}

func TestParseSVGPathMultipleSubpaths(t *testing.T) {
	path := ParseSVGPath("M 0 0 L 10 10 Z M 20 20 L 30 30 Z")
	if len(path.Commands) != 6 {
		t.Fatalf("expected 6 commands, got %d", len(path.Commands))
	}
}

func TestParseSVGPathImplicitLineTo(t *testing.T) {
	// After M, subsequent coordinate pairs become L
	path := ParseSVGPath("M 0 0 10 10 20 20")
	if len(path.Commands) != 3 {
		t.Fatalf("expected 3 commands (M + 2 implicit L), got %d", len(path.Commands))
	}
	if path.Commands[1].Type != PathLineTo {
		t.Error("expected implicit LineTo")
	}
}

func TestParseSVGPathRelativeImplicitLineTo(t *testing.T) {
	// After m, subsequent coordinate pairs become l
	path := ParseSVGPath("m 0 0 10 10 20 20")
	if len(path.Commands) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(path.Commands))
	}
	// Second command: relative l 10 10 → LineTo(10, 10)
	if path.Commands[1].X1 != 10 || path.Commands[1].Y1 != 10 {
		t.Errorf("implicit l: expected (10,10), got (%g,%g)", path.Commands[1].X1, path.Commands[1].Y1)
	}
}

func TestParseSVGPathNegativeNumbers(t *testing.T) {
	path := ParseSVGPath("M-10-20L-30-40")
	if len(path.Commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(path.Commands))
	}
	if path.Commands[0].X1 != -10 || path.Commands[0].Y1 != -20 {
		t.Errorf("M: expected (-10,-20), got (%g,%g)", path.Commands[0].X1, path.Commands[0].Y1)
	}
}

func TestParseSVGPathEmpty(t *testing.T) {
	path := ParseSVGPath("")
	if len(path.Commands) != 0 {
		t.Errorf("expected 0 commands for empty path, got %d", len(path.Commands))
	}
}

func TestParseSVGPathUnknownCommand(t *testing.T) {
	// Unknown commands should be skipped without panic
	path := ParseSVGPath("M 0 0 X 10 20 L 30 40")
	if path == nil {
		t.Fatal("expected non-nil path")
	}
}

func TestDrawPath(t *testing.T) {
	buf := NewCommandBuffer()
	p := NewPath()
	p.MoveTo(0, 0)
	p.LineTo(100, 0)
	p.LineTo(100, 100)
	p.Close()

	buf.DrawPath(p, uimath.ColorBlack, 2, uimath.ColorRed, 0, 1)
	if buf.Len() == 0 {
		t.Error("expected commands from DrawPath")
	}
}

func TestDrawPathStrokeOnly(t *testing.T) {
	buf := NewCommandBuffer()
	p := NewPath()
	p.MoveTo(0, 0)
	p.LineTo(100, 0)

	buf.DrawPath(p, uimath.ColorBlack, 1, uimath.ColorTransparent, 0, 1)
	if buf.Len() == 0 {
		t.Error("expected stroke commands")
	}
}

func TestDrawPathFillOnly(t *testing.T) {
	buf := NewCommandBuffer()
	p := NewPath()
	p.MoveTo(0, 0)
	p.LineTo(100, 0)
	p.LineTo(100, 100)
	p.Close()

	buf.DrawPath(p, uimath.ColorTransparent, 0, uimath.ColorRed, 0, 1)
	if buf.Len() == 0 {
		t.Error("expected fill commands")
	}
}

func TestDrawPathEmptyPath(t *testing.T) {
	buf := NewCommandBuffer()
	p := NewPath()
	// No points, transparent fill → should produce nothing
	buf.DrawPath(p, uimath.ColorBlack, 1, uimath.ColorTransparent, 0, 1)
	if buf.Len() != 0 {
		t.Error("expected no commands for empty path")
	}
}

func TestDrawPathDiagonalLine(t *testing.T) {
	buf := NewCommandBuffer()
	p := NewPath()
	p.MoveTo(0, 0)
	p.LineTo(100, 100) // diagonal line → exercises the else branch
	buf.DrawPath(p, uimath.ColorBlack, 2, uimath.ColorTransparent, 0, 1)
	if buf.Len() == 0 {
		t.Error("expected commands for diagonal line")
	}
}

func TestDrawPathVerticalLine(t *testing.T) {
	buf := NewCommandBuffer()
	p := NewPath()
	p.MoveTo(50, 0)
	p.LineTo(50, 100) // vertical line → exercises the vertical-ish branch
	buf.DrawPath(p, uimath.ColorBlack, 2, uimath.ColorTransparent, 0, 1)
	if buf.Len() == 0 {
		t.Error("expected commands for vertical line")
	}
}

func TestDrawPathHorizontalLine(t *testing.T) {
	buf := NewCommandBuffer()
	p := NewPath()
	p.MoveTo(0, 50)
	p.LineTo(100, 50) // horizontal line
	buf.DrawPath(p, uimath.ColorBlack, 2, uimath.ColorTransparent, 0, 1)
	if buf.Len() == 0 {
		t.Error("expected commands for horizontal line")
	}
}

func TestDrawPathZeroLengthSegment(t *testing.T) {
	buf := NewCommandBuffer()
	p := NewPath()
	p.MoveTo(50, 50)
	p.LineTo(50, 50) // zero-length segment → should be skipped
	buf.DrawPath(p, uimath.ColorBlack, 2, uimath.ColorTransparent, 0, 1)
	// Zero-length skipped, no commands expected
}

func TestPathEmpty(t *testing.T) {
	p := NewPath()
	b := p.Bounds()
	if !b.IsEmpty() {
		t.Error("expected empty bounds for empty path")
	}
	pts := p.Flatten(1.0)
	if len(pts) != 0 {
		t.Error("expected no points for empty path")
	}
}

func TestTokenizeSVGPathWithCommas(t *testing.T) {
	tokens := tokenizeSVGPath("M 0,0 L 100,50")
	// Should produce: M, 0, 0, L, 100, 50
	expected := []string{"M", "0", "0", "L", "100", "50"}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d: %v", len(expected), len(tokens), tokens)
	}
	for i, want := range expected {
		if tokens[i] != want {
			t.Errorf("token[%d]: want %q, got %q", i, want, tokens[i])
		}
	}
}

func TestTokenizeSVGPathWithTabs(t *testing.T) {
	tokens := tokenizeSVGPath("M\t10\t20\nL\t30\t40")
	if len(tokens) != 6 {
		t.Fatalf("expected 6 tokens, got %d: %v", len(tokens), tokens)
	}
}

func TestTokenizeSVGPathWithExponent(t *testing.T) {
	tokens := tokenizeSVGPath("M 1e2 2.5E-1")
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d: %v", len(tokens), tokens)
	}
	if tokens[1] != "1e2" {
		t.Errorf("expected '1e2', got %q", tokens[1])
	}
	if tokens[2] != "2.5E-1" {
		t.Errorf("expected '2.5E-1', got %q", tokens[2])
	}
}

func TestTokenizeSVGPathWithPlusSign(t *testing.T) {
	tokens := tokenizeSVGPath("M +10 +20")
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d: %v", len(tokens), tokens)
	}
	if tokens[1] != "+10" {
		t.Errorf("expected '+10', got %q", tokens[1])
	}
}

func TestTokenizeSVGPathSkipUnrecognized(t *testing.T) {
	// Characters like @ should be skipped
	tokens := tokenizeSVGPath("M 0 0 @ L 10 10")
	found := false
	for _, tok := range tokens {
		if tok == "L" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'L' token after skipping unrecognized char")
	}
}

func TestTokenizeSVGPathExponentPlusSign(t *testing.T) {
	tokens := tokenizeSVGPath("1.5e+2")
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d: %v", len(tokens), tokens)
	}
	if tokens[0] != "1.5e+2" {
		t.Errorf("expected '1.5e+2', got %q", tokens[0])
	}
}

func TestParseFloat32(t *testing.T) {
	tests := []struct {
		s    string
		want float32
	}{
		{"0", 0},
		{"100", 100},
		{"-50", -50},
		{"3.14", 3.14},
		{"-0.5", -0.5},
		{"1e2", 100},
		{"1.5e-1", 0.15},
		{"+42", 42},
		{"1.5E+2", 150},
		{"0.0", 0},
	}
	for _, tt := range tests {
		got := parseFloat32(tt.s)
		diff := got - tt.want
		if diff < 0 {
			diff = -diff
		}
		if diff > 0.01 {
			t.Errorf("parseFloat32(%q) = %g, want %g", tt.s, got, tt.want)
		}
	}
}

func TestAbsF(t *testing.T) {
	if absF(-5) != 5 {
		t.Error("absF(-5) should be 5")
	}
	if absF(5) != 5 {
		t.Error("absF(5) should be 5")
	}
	if absF(0) != 0 {
		t.Error("absF(0) should be 0")
	}
}

func TestIsAlpha(t *testing.T) {
	if !isAlpha('M') {
		t.Error("M should be alpha")
	}
	if !isAlpha('z') {
		t.Error("z should be alpha")
	}
	if isAlpha('5') {
		t.Error("5 should not be alpha")
	}
	if isAlpha(' ') {
		t.Error("space should not be alpha")
	}
}

func TestArcToLargeArcAndSweep(t *testing.T) {
	p := NewPath()
	p.MoveTo(0, 0)
	p.ArcTo(25, 25, 0, true, true, 50, 0) // largeArc=true, sweep=true
	if len(p.Commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(p.Commands))
	}
	a := p.Commands[1]
	// Y2 encodes: largeArc*2 + sweep = 1*2 + 1 = 3
	if a.Y2 != 3 {
		t.Errorf("Y2 encoding: expected 3, got %g", a.Y2)
	}
}

func TestArcToNoLargeArcNoSweep(t *testing.T) {
	p := NewPath()
	p.MoveTo(0, 0)
	p.ArcTo(25, 25, 0, false, false, 50, 0) // la=false, sw=false
	a := p.Commands[1]
	// Y2 encodes: 0*2 + 0 = 0
	if a.Y2 != 0 {
		t.Errorf("Y2 encoding: expected 0, got %g", a.Y2)
	}
}

func TestFlattenArcLargeDistance(t *testing.T) {
	// Test arc with large distance to hit steps > 100 cap
	p := NewPath()
	p.MoveTo(0, 0)
	p.ArcTo(500, 500, 0, true, true, 1000, 0)
	pts := p.Flatten(0.1) // small tolerance → more steps, but capped at 100
	if len(pts) < 8 {
		t.Errorf("expected many points from large arc, got %d", len(pts))
	}
}

func TestParseSVGPathLowercaseZ(t *testing.T) {
	path := ParseSVGPath("M 0 0 L 100 0 L 100 100 z")
	found := false
	for _, c := range path.Commands {
		if c.Type == PathClose {
			found = true
		}
	}
	if !found {
		t.Error("expected Close from lowercase 'z'")
	}
}

func TestParseSVGPathMixed(t *testing.T) {
	// Complex path with many command types
	path := ParseSVGPath("M 0 0 L 50 0 Q 75 25 50 50 C 25 75 0 50 0 25 A 10 10 0 0 1 20 20 S 30 40 50 50 T 60 70 H 80 V 90 Z")
	if len(path.Commands) == 0 {
		t.Error("expected commands from complex path")
	}
	// Should have Close at end
	last := path.Commands[len(path.Commands)-1]
	if last.Type != PathClose {
		t.Error("expected Close at end")
	}
}

func TestDrawLineSegmentMinX(t *testing.T) {
	// Test horizontal line where x2 < x1 (reverse direction)
	buf := NewCommandBuffer()
	p := NewPath()
	p.MoveTo(100, 50)
	p.LineTo(0, 50) // horizontal, x2 < x1
	buf.DrawPath(p, uimath.ColorBlack, 2, uimath.ColorTransparent, 0, 1)
	if buf.Len() == 0 {
		t.Error("expected commands for reverse horizontal line")
	}
}

func TestDrawLineSegmentMinY(t *testing.T) {
	// Test vertical line where y2 < y1 (reverse direction)
	buf := NewCommandBuffer()
	p := NewPath()
	p.MoveTo(50, 100)
	p.LineTo(50, 0) // vertical, y2 < y1
	buf.DrawPath(p, uimath.ColorBlack, 2, uimath.ColorTransparent, 0, 1)
	if buf.Len() == 0 {
		t.Error("expected commands for reverse vertical line")
	}
}

func TestParseSVGPathSmoothCubicReflection(t *testing.T) {
	// C 10 20 30 40 50 0 S 70 -40 90 0
	// After C: last cp2 = (30,40), endpoint = (50,0)
	// S reflects cp2: cp1 = 2*(50,0) - (30,40) = (70,-40)
	path := ParseSVGPath("M 0 0 C 10 20 30 40 50 0 S 70 -40 90 0")
	if len(path.Commands) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(path.Commands))
	}
	s := path.Commands[2] // the S command → CubicTo
	if s.Type != PathCubicTo {
		t.Error("expected CubicTo from S")
	}
	// cp1 should be reflected: (70, -40)
	if s.X1 != 70 || s.Y1 != -40 {
		t.Errorf("S reflected cp1: expected (70,-40), got (%g,%g)", s.X1, s.Y1)
	}
}

func TestParseSVGPathSmoothQuadReflection(t *testing.T) {
	// Q 50 100 100 0 T 200 0
	// After Q: last cp = (50,100), endpoint = (100,0)
	// T reflects: cp = 2*(100,0) - (50,100) = (150,-100)
	path := ParseSVGPath("M 0 0 Q 50 100 100 0 T 200 0")
	if len(path.Commands) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(path.Commands))
	}
	tCmd := path.Commands[2]
	if tCmd.Type != PathQuadTo {
		t.Error("expected QuadTo from T")
	}
	// Reflected cp should be (150, -100)
	if tCmd.X1 != 150 || tCmd.Y1 != -100 {
		t.Errorf("T reflected cp: expected (150,-100), got (%g,%g)", tCmd.X1, tCmd.Y1)
	}
}

func TestParseSVGPathNextFloatOutOfBounds(t *testing.T) {
	// Path with insufficient numbers for a command
	path := ParseSVGPath("M 10")
	// Should not panic; M gets (10, 0)
	if len(path.Commands) == 0 {
		t.Error("expected at least 1 command")
	}
}

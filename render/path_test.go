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
	// Second command should be absolute (100, 0)
	if path.Commands[1].X1 != 100 || path.Commands[1].Y1 != 0 {
		t.Errorf("expected (100,0), got (%g,%g)", path.Commands[1].X1, path.Commands[1].Y1)
	}
}

func TestParseSVGPathHV(t *testing.T) {
	path := ParseSVGPath("M 0 0 H 100 V 50")
	if len(path.Commands) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(path.Commands))
	}
	// H 100 → LineTo(100, 0)
	if path.Commands[1].X1 != 100 || path.Commands[1].Y1 != 0 {
		t.Errorf("H: expected (100,0), got (%g,%g)", path.Commands[1].X1, path.Commands[1].Y1)
	}
	// V 50 → LineTo(100, 50)
	if path.Commands[2].X1 != 100 || path.Commands[2].Y1 != 50 {
		t.Errorf("V: expected (100,50), got (%g,%g)", path.Commands[2].X1, path.Commands[2].Y1)
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

func TestParseSVGPathQuad(t *testing.T) {
	path := ParseSVGPath("M 0 0 Q 50 100 100 0")
	if len(path.Commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(path.Commands))
	}
	if path.Commands[1].Type != PathQuadTo {
		t.Error("expected QuadTo")
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

func TestParseSVGPathMultipleSubpaths(t *testing.T) {
	path := ParseSVGPath("M 0 0 L 10 10 Z M 20 20 L 30 30 Z")
	if len(path.Commands) != 6 {
		t.Fatalf("expected 6 commands, got %d", len(path.Commands))
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

func TestParseSVGPathNegativeNumbers(t *testing.T) {
	path := ParseSVGPath("M-10-20L-30-40")
	if len(path.Commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(path.Commands))
	}
	if path.Commands[0].X1 != -10 || path.Commands[0].Y1 != -20 {
		t.Errorf("M: expected (-10,-20), got (%g,%g)", path.Commands[0].X1, path.Commands[0].Y1)
	}
}

func TestParseFloat32(t *testing.T) {
	tests := []struct{ s string; want float32 }{
		{"0", 0},
		{"100", 100},
		{"-50", -50},
		{"3.14", 3.14},
		{"-0.5", -0.5},
		{"1e2", 100},
		{"1.5e-1", 0.15},
	}
	for _, tt := range tests {
		got := parseFloat32(tt.s)
		diff := got - tt.want
		if diff < 0 { diff = -diff }
		if diff > 0.01 {
			t.Errorf("parseFloat32(%q) = %g, want %g", tt.s, got, tt.want)
		}
	}
}

package widget

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/render"
)

// ---------------------------------------------------------------------------
// Mock backend for texture tests
// ---------------------------------------------------------------------------

type mockBackend struct {
	created   int
	updated   int
	destroyed int
	failCreate bool
}

func (m *mockBackend) Init(_ platform.Window) error                    { return nil }
func (m *mockBackend) BeginFrame()                                    {}
func (m *mockBackend) EndFrame()                                      {}
func (m *mockBackend) Resize(_, _ int)                                {}
func (m *mockBackend) Submit(_ *render.CommandBuffer)                 {}
func (m *mockBackend) MaxTextureSize() int                            { return 4096 }
func (m *mockBackend) DPIScale() float32                              { return 1 }
func (m *mockBackend) ReadPixels() (*image.RGBA, error)               { return nil, nil }
func (m *mockBackend) Destroy()                                       {}
func (m *mockBackend) DestroyTexture(_ render.TextureHandle)          { m.destroyed++ }
func (m *mockBackend) UpdateTexture(_ render.TextureHandle, _ uimath.Rect, _ []byte) error {
	m.updated++
	return nil
}
func (m *mockBackend) CreateTexture(_ render.TextureDesc) (render.TextureHandle, error) {
	m.created++
	if m.failCreate {
		return render.InvalidTexture, image.ErrFormat
	}
	return render.TextureHandle(m.created), nil
}

func newTestCanvas(w, h int) *Canvas {
	tree := core.NewTree()
	return NewCanvas(tree, w, h, nil, nil)
}

func newTestCanvasWithBackend(w, h int) (*Canvas, *mockBackend) {
	tree := core.NewTree()
	mb := &mockBackend{}
	c := NewCanvas(tree, w, h, mb, nil)
	return c, mb
}

func newTestCanvasWithText(w, h int) *Canvas {
	tree := core.NewTree()
	cfg := DefaultConfig()
	cfg.TextRenderer = &mockTextDrawer{}
	return NewCanvas(tree, w, h, nil, cfg)
}

// ---------------------------------------------------------------------------
// Construction & basics
// ---------------------------------------------------------------------------

func TestCanvasNewAndSize(t *testing.T) {
	c := newTestCanvas(100, 80)
	if c.Width() != 100 || c.Height() != 80 {
		t.Fatalf("expected 100x80, got %dx%d", c.Width(), c.Height())
	}
	if len(c.Pixels()) != 100*80*4 {
		t.Fatalf("pixel buffer size: got %d, want %d", len(c.Pixels()), 100*80*4)
	}

	c.SetSize(50, 40)
	if c.Width() != 50 || c.Height() != 40 {
		t.Fatalf("after resize: expected 50x40, got %dx%d", c.Width(), c.Height())
	}
	if len(c.Pixels()) != 50*40*4 {
		t.Fatal("pixel buffer not resized")
	}

	// SetSize same dimensions — no-op
	c.SetSize(50, 40)
	if c.Width() != 50 {
		t.Fatal("SetSize same should be no-op")
	}
}

func TestCanvasGetContext2D(t *testing.T) {
	c := newTestCanvas(64, 64)
	ctx := c.GetContext2D()
	if ctx == nil {
		t.Fatal("GetContext2D returned nil")
	}
}

// ---------------------------------------------------------------------------
// Backend integration: Sync, Draw, Destroy with texture
// ---------------------------------------------------------------------------

func TestCanvasSyncCreateTexture(t *testing.T) {
	c, mb := newTestCanvasWithBackend(10, 10)
	c.Sync()
	if mb.created != 1 {
		t.Fatalf("expected 1 texture created, got %d", mb.created)
	}
	// Second sync without changes — update (dirty=false after first sync, so no-op)
	c.Sync()
	if mb.updated != 0 {
		t.Fatalf("expected 0 updates (not dirty), got %d", mb.updated)
	}
	// Mark dirty and sync again — should update
	c.GetContext2D().FillRect(0, 0, 1, 1) // marks dirty
	c.Sync()
	if mb.updated != 1 {
		t.Fatalf("expected 1 update after dirty, got %d", mb.updated)
	}
}

func TestCanvasSyncCreateFail(t *testing.T) {
	c, mb := newTestCanvasWithBackend(10, 10)
	mb.failCreate = true
	c.Sync()
	// Texture should remain invalid
	if c.texture != render.InvalidTexture {
		t.Fatal("texture should be invalid after failed create")
	}
}

func TestCanvasDrawWithBackend(t *testing.T) {
	c, _ := newTestCanvasWithBackend(10, 10)
	// Set layout bounds so Draw produces a command
	tree := c.tree
	tree.SetLayout(c.id, core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 10, 10),
	})
	buf := render.NewCommandBuffer()
	c.Draw(buf)
	if buf.Len() != 1 {
		t.Fatalf("expected 1 draw command, got %d", buf.Len())
	}
}

func TestCanvasDrawEmptyBounds(t *testing.T) {
	c, _ := newTestCanvasWithBackend(10, 10)
	// Don't set layout — bounds will be empty
	buf := render.NewCommandBuffer()
	c.Draw(buf)
	if buf.Len() != 0 {
		t.Fatalf("expected 0 commands with empty bounds, got %d", buf.Len())
	}
}

func TestCanvasDestroyWithTexture(t *testing.T) {
	c, mb := newTestCanvasWithBackend(10, 10)
	c.Sync() // creates texture
	c.Destroy()
	if mb.destroyed != 1 {
		t.Fatalf("expected 1 texture destroyed, got %d", mb.destroyed)
	}
}

func TestCanvasSetSizeWithBackend(t *testing.T) {
	c, mb := newTestCanvasWithBackend(10, 10)
	c.Sync() // creates texture
	c.SetSize(20, 20)
	if mb.destroyed != 1 {
		t.Fatalf("expected old texture destroyed on resize, got %d", mb.destroyed)
	}
}

// ---------------------------------------------------------------------------
// Rectangle operations
// ---------------------------------------------------------------------------

func TestCanvasFillRect(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()

	ctx.SetFillColor(uimath.Color{R: 1, G: 0, B: 0, A: 1})
	ctx.FillRect(2, 2, 4, 4)

	i := (3*10 + 3) * 4
	if c.pixels[i] != 255 || c.pixels[i+1] != 0 || c.pixels[i+2] != 0 || c.pixels[i+3] != 255 {
		t.Fatalf("pixel at (3,3): got RGBA(%d,%d,%d,%d), want (255,0,0,255)",
			c.pixels[i], c.pixels[i+1], c.pixels[i+2], c.pixels[i+3])
	}

	if c.pixels[0] != 0 || c.pixels[3] != 0 {
		t.Fatalf("pixel at (0,0) should be empty")
	}
}

func TestCanvasFillRectOutOfBounds(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	ctx.SetFillColor(uimath.Color{R: 1, G: 0, B: 0, A: 1})
	// Partially out of bounds — should not panic
	ctx.FillRect(-5, -5, 20, 20)
	i := (0*10 + 0) * 4
	if c.pixels[i+3] == 0 {
		t.Fatal("pixel at (0,0) should be filled")
	}
}

func TestCanvasClearRect(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()

	ctx.SetFillColor(uimath.Color{R: 0, G: 0, B: 1, A: 1})
	ctx.FillRect(0, 0, 10, 10)
	ctx.ClearRect(2, 2, 3, 3)

	i := (3*10 + 3) * 4
	if c.pixels[i+3] != 0 {
		t.Fatalf("cleared pixel alpha: got %d, want 0", c.pixels[i+3])
	}
	i = 0
	if c.pixels[i+2] != 255 {
		t.Fatalf("uncleaned pixel blue: got %d, want 255", c.pixels[i+2])
	}
}

func TestCanvasClearRectOutOfBounds(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	ctx.SetFillColor(uimath.Color{R: 1, A: 1})
	ctx.FillRect(0, 0, 10, 10)
	ctx.ClearRect(-5, -5, 20, 20) // should not panic
	if c.pixels[3] != 0 {
		t.Fatal("pixel should be cleared")
	}
}

func TestCanvasStrokeRect(t *testing.T) {
	c := newTestCanvas(20, 20)
	ctx := c.GetContext2D()
	ctx.SetStrokeColor(uimath.Color{R: 0, G: 1, B: 0, A: 1})
	ctx.SetLineWidth(1)
	ctx.StrokeRect(5, 5, 10, 10)

	i := (5*20 + 10) * 4
	if c.pixels[i+1] != 255 {
		t.Fatalf("stroke pixel green: got %d, want 255", c.pixels[i+1])
	}
	i = (10*20 + 10) * 4
	if c.pixels[i+3] != 0 {
		t.Fatalf("interior should be empty, got alpha %d", c.pixels[i+3])
	}
}

// ---------------------------------------------------------------------------
// Path fill & stroke
// ---------------------------------------------------------------------------

func TestCanvasPathFill(t *testing.T) {
	c := newTestCanvas(20, 20)
	ctx := c.GetContext2D()
	ctx.SetFillColor(uimath.Color{R: 1, G: 1, B: 0, A: 1})

	ctx.BeginPath()
	ctx.MoveTo(10, 0)
	ctx.LineTo(20, 20)
	ctx.LineTo(0, 20)
	ctx.ClosePath()
	ctx.Fill()

	i := (15*20 + 10) * 4
	if c.pixels[i+3] == 0 {
		t.Fatal("triangle center should be filled")
	}
}

func TestCanvasPathFillEmpty(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	ctx.BeginPath()
	ctx.Fill() // empty path — should not panic
}

func TestCanvasPathStroke(t *testing.T) {
	c := newTestCanvas(20, 20)
	ctx := c.GetContext2D()
	ctx.SetStrokeColor(uimath.Color{R: 1, G: 0, B: 1, A: 1})
	ctx.SetLineWidth(1)

	ctx.BeginPath()
	ctx.MoveTo(0, 10)
	ctx.LineTo(20, 10)
	ctx.Stroke()

	i := (10*20 + 10) * 4
	if c.pixels[i+3] == 0 {
		t.Fatal("line pixel should be filled")
	}
}

func TestCanvasPathStrokeEmpty(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	ctx.BeginPath()
	ctx.Stroke() // empty path — should not panic
}

func TestCanvasPathStrokeClose(t *testing.T) {
	c := newTestCanvas(20, 20)
	ctx := c.GetContext2D()
	ctx.SetStrokeColor(uimath.Color{R: 1, A: 1})
	ctx.SetLineWidth(1)

	ctx.BeginPath()
	ctx.MoveTo(5, 5)
	ctx.LineTo(15, 5)
	ctx.LineTo(15, 15)
	ctx.ClosePath()
	ctx.Stroke()

	// Closing edge from (15,15) to (5,5) should have pixels
	hasPixels := false
	for y := 6; y < 14; y++ {
		for x := 6; x < 14; x++ {
			if c.pixels[(y*20+x)*4+3] > 0 {
				hasPixels = true
				break
			}
		}
		if hasPixels {
			break
		}
	}
	if !hasPixels {
		t.Fatal("closing stroke should produce pixels")
	}
}

func TestCanvasPathStrokeThick(t *testing.T) {
	c := newTestCanvas(30, 30)
	ctx := c.GetContext2D()
	ctx.SetStrokeColor(uimath.Color{R: 1, A: 1})
	ctx.SetLineWidth(5)

	ctx.BeginPath()
	ctx.MoveTo(5, 15)
	ctx.LineTo(25, 15)
	ctx.Stroke()

	// Should have pixels on or near the center line
	hasPixels := false
	for y := 12; y < 18; y++ {
		if c.pixels[(y*30+15)*4+3] > 0 {
			hasPixels = true
			break
		}
	}
	if !hasPixels {
		t.Fatal("thick line should have width")
	}
}

func TestCanvasPathFillNoClose(t *testing.T) {
	// Fill with path that has no ClosePath — still has edges from line segments
	c := newTestCanvas(20, 20)
	ctx := c.GetContext2D()
	ctx.SetFillColor(uimath.Color{R: 1, A: 1})
	ctx.BeginPath()
	ctx.MoveTo(0, 0)
	ctx.LineTo(20, 0)
	ctx.LineTo(20, 20)
	ctx.LineTo(0, 20)
	ctx.ClosePath()
	ctx.Fill()

	i := (10*20 + 10) * 4
	if c.pixels[i+3] == 0 {
		t.Fatal("fill should work")
	}
}

// ---------------------------------------------------------------------------
// Save / Restore
// ---------------------------------------------------------------------------

func TestCanvasSaveRestore(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()

	ctx.SetGlobalAlpha(0.5)
	ctx.Save()
	ctx.SetGlobalAlpha(0.2)
	if ctx.GlobalAlpha() != 0.2 {
		t.Fatalf("expected 0.2, got %f", ctx.GlobalAlpha())
	}
	ctx.Restore()
	if ctx.GlobalAlpha() != 0.5 {
		t.Fatalf("expected 0.5 after restore, got %f", ctx.GlobalAlpha())
	}
}

func TestCanvasRestoreEmpty(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	ctx.Restore() // empty stack — should not panic
}

// ---------------------------------------------------------------------------
// Transforms
// ---------------------------------------------------------------------------

func TestCanvasTranslate(t *testing.T) {
	c := newTestCanvas(20, 20)
	ctx := c.GetContext2D()

	ctx.Translate(5, 5)
	ctx.SetFillColor(uimath.Color{R: 1, G: 0, B: 0, A: 1})
	ctx.FillRect(0, 0, 5, 5)

	i := (7*20 + 7) * 4
	if c.pixels[i] != 255 {
		t.Fatalf("translated pixel at (7,7): R=%d, want 255", c.pixels[i])
	}
	i = (2*20 + 2) * 4
	if c.pixels[i+3] != 0 {
		t.Fatal("pixel at (2,2) should be empty")
	}
}

func TestCanvasResetTransform(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	ctx.Translate(100, 100)
	ctx.ResetTransform()

	a, b, cc, d, e, f := ctx.GetTransform()
	if a != 1 || b != 0 || cc != 0 || d != 1 || e != 0 || f != 0 {
		t.Fatal("transform not reset to identity")
	}
}

func TestCanvasTransformMethod(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	// Apply manual transform (scale 2x)
	ctx.Transform(2, 0, 0, 2, 0, 0)
	a, _, _, d, _, _ := ctx.GetTransform()
	if a != 2 || d != 2 {
		t.Fatalf("expected scale 2x, got a=%f d=%f", a, d)
	}
}

func TestCanvasSetTransform(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	ctx.Translate(50, 50) // will be overwritten
	ctx.SetTransform(3, 0, 0, 3, 10, 20)
	a, b, cc, d, e, f := ctx.GetTransform()
	if a != 3 || b != 0 || cc != 0 || d != 3 || e != 10 || f != 20 {
		t.Fatalf("SetTransform mismatch: %f %f %f %f %f %f", a, b, cc, d, e, f)
	}
}

func TestCanvasScale(t *testing.T) {
	c := newTestCanvas(20, 20)
	ctx := c.GetContext2D()
	ctx.Scale(2, 2)
	ctx.SetFillColor(uimath.Color{R: 1, G: 0, B: 0, A: 1})
	ctx.FillRect(0, 0, 5, 5)

	i := (8*20 + 8) * 4
	if c.pixels[i+3] == 0 {
		t.Fatal("scaled rect should extend to (8,8)")
	}
	i = (12*20 + 12) * 4
	if c.pixels[i+3] != 0 {
		t.Fatal("pixel at (12,12) should be empty")
	}
}

func TestCanvasRotate(t *testing.T) {
	c := newTestCanvas(30, 30)
	ctx := c.GetContext2D()
	ctx.Translate(15, 15)
	ctx.Rotate(math.Pi / 4)
	ctx.SetFillColor(uimath.Color{R: 1, G: 0, B: 0, A: 1})
	ctx.FillRect(-5, -5, 10, 10)

	i := (15*30 + 15) * 4
	if c.pixels[i+3] == 0 {
		t.Fatal("rotated rect center should be filled")
	}
}

// ---------------------------------------------------------------------------
// Style properties — getters/setters
// ---------------------------------------------------------------------------

func TestCanvasStyleGettersSetters(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()

	// FillColor / StrokeColor
	red := uimath.Color{R: 1, G: 0, B: 0, A: 1}
	ctx.SetFillColor(red)
	if ctx.FillColor() != red {
		t.Fatal("FillColor mismatch")
	}
	blue := uimath.Color{R: 0, G: 0, B: 1, A: 1}
	ctx.SetStrokeColor(blue)
	if ctx.StrokeColor() != blue {
		t.Fatal("StrokeColor mismatch")
	}

	// RGBA convenience
	ctx.SetFillStyleRGBA(0.5, 0.6, 0.7, 0.8)
	fc := ctx.FillColor()
	if fc.R != 0.5 || fc.G != 0.6 || fc.B != 0.7 || fc.A != 0.8 {
		t.Fatalf("SetFillStyleRGBA: got %v", fc)
	}
	ctx.SetStrokeStyleRGBA(0.1, 0.2, 0.3, 0.4)
	sc := ctx.StrokeColor()
	if sc.R != 0.1 || sc.G != 0.2 || sc.B != 0.3 || sc.A != 0.4 {
		t.Fatalf("SetStrokeStyleRGBA: got %v", sc)
	}

	// LineWidth
	ctx.SetLineWidth(3)
	if ctx.LineWidth() != 3 {
		t.Fatal("LineWidth mismatch")
	}

	// GlobalCompositeOperation
	ctx.SetGlobalCompositeOperation(CompositeLighter)
	if ctx.GlobalCompositeOperation() != CompositeLighter {
		t.Fatal("GlobalCompositeOperation mismatch")
	}

	// Font
	ctx.SetFont(16)
	if ctx.Font() != 16 {
		t.Fatal("Font mismatch")
	}

	// TextAlign
	ctx.SetTextAlign(CanvasTextAlignCenter)
	if ctx.TextAlign() != CanvasTextAlignCenter {
		t.Fatal("TextAlign mismatch")
	}

	// TextBaseline
	ctx.SetTextBaseline(CanvasTextBaselineMiddle)
	if ctx.TextBaseline() != CanvasTextBaselineMiddle {
		t.Fatal("TextBaseline mismatch")
	}
}

func TestCanvasStrokeStyleGradient(t *testing.T) {
	c := newTestCanvas(20, 20)
	ctx := c.GetContext2D()
	g := ctx.CreateLinearGradient(0, 0, 20, 0)
	g.AddColorStop(0, uimath.Color{R: 1, A: 1})
	g.AddColorStop(1, uimath.Color{B: 1, A: 1})
	ctx.SetStrokeStyleGradient(g)
	ctx.SetLineWidth(3)
	ctx.StrokeRect(2, 2, 16, 16)

	// Some edge pixels should be drawn
	hasColor := false
	for i := 3; i < len(c.pixels); i += 4 {
		if c.pixels[i] > 0 {
			hasColor = true
			break
		}
	}
	if !hasColor {
		t.Fatal("stroke gradient should produce pixels")
	}
}

func TestCanvasLineStyles(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()

	ctx.SetLineCap(LineCapRound)
	ctx.SetLineJoin(LineJoinBevel)
	ctx.SetMiterLimit(5)

	if ctx.LineCap() != LineCapRound {
		t.Fatal("lineCap mismatch")
	}
	if ctx.LineJoin() != LineJoinBevel {
		t.Fatal("lineJoin mismatch")
	}
	if ctx.MiterLimit() != 5 {
		t.Fatal("miterLimit mismatch")
	}
}

// ---------------------------------------------------------------------------
// Arcs & curves
// ---------------------------------------------------------------------------

func TestCanvasArc(t *testing.T) {
	c := newTestCanvas(30, 30)
	ctx := c.GetContext2D()
	ctx.SetFillColor(uimath.Color{R: 0, G: 0, B: 1, A: 1})

	ctx.BeginPath()
	ctx.Arc(15, 15, 10, 0, 2*math.Pi, false)
	ctx.ClosePath()
	ctx.Fill()

	i := (15*30 + 15) * 4
	if c.pixels[i+3] == 0 {
		t.Fatal("circle center should be filled")
	}
	if c.pixels[3] != 0 {
		t.Fatal("corner (0,0) should be empty")
	}
}

func TestCanvasArcCounterClockwise(t *testing.T) {
	c := newTestCanvas(30, 30)
	ctx := c.GetContext2D()
	ctx.SetStrokeColor(uimath.Color{R: 1, A: 1})
	ctx.SetLineWidth(2)

	ctx.BeginPath()
	ctx.Arc(15, 15, 10, 0, math.Pi, true) // CCW
	ctx.Stroke()

	hasPixels := false
	for i := 3; i < len(c.pixels); i += 4 {
		if c.pixels[i] > 0 {
			hasPixels = true
			break
		}
	}
	if !hasPixels {
		t.Fatal("CCW arc should produce pixels")
	}
}

func TestCanvasArcToStartOnEmptyPath(t *testing.T) {
	c := newTestCanvas(30, 30)
	ctx := c.GetContext2D()
	ctx.SetStrokeColor(uimath.Color{R: 1, A: 1})
	ctx.SetLineWidth(1)

	// Arc on empty path should create an initial MoveTo
	ctx.BeginPath()
	ctx.Arc(15, 15, 5, 0, math.Pi/2, false)
	ctx.Stroke()

	hasPixels := false
	for i := 3; i < len(c.pixels); i += 4 {
		if c.pixels[i] > 0 {
			hasPixels = true
			break
		}
	}
	if !hasPixels {
		t.Fatal("arc on empty path should produce pixels")
	}
}

func TestCanvasArcTo(t *testing.T) {
	c := newTestCanvas(30, 30)
	ctx := c.GetContext2D()
	ctx.SetStrokeColor(uimath.Color{R: 1, A: 1})
	ctx.SetLineWidth(2)

	ctx.BeginPath()
	ctx.MoveTo(5, 5)
	ctx.ArcTo(25, 5, 25, 25, 10)
	ctx.Stroke()

	hasPixels := false
	for i := 3; i < len(c.pixels); i += 4 {
		if c.pixels[i] > 0 {
			hasPixels = true
			break
		}
	}
	if !hasPixels {
		t.Fatal("ArcTo should produce pixels")
	}
}

func TestCanvasQuadraticCurve(t *testing.T) {
	c := newTestCanvas(30, 30)
	ctx := c.GetContext2D()
	ctx.SetStrokeColor(uimath.Color{R: 1, G: 1, B: 1, A: 1})
	ctx.SetLineWidth(2)

	ctx.BeginPath()
	ctx.MoveTo(0, 15)
	ctx.QuadraticCurveTo(15, 0, 30, 15)
	ctx.Stroke()

	filled := false
	for x := 0; x < 30; x++ {
		for y := 0; y < 15; y++ {
			if c.pixels[(y*30+x)*4+3] > 0 {
				filled = true
				break
			}
		}
		if filled {
			break
		}
	}
	if !filled {
		t.Fatal("quadratic curve should have filled some pixels")
	}
}

func TestCanvasBezierCurve(t *testing.T) {
	c := newTestCanvas(30, 30)
	ctx := c.GetContext2D()
	ctx.SetStrokeColor(uimath.Color{R: 1, G: 1, B: 1, A: 1})
	ctx.SetLineWidth(2)

	ctx.BeginPath()
	ctx.MoveTo(0, 15)
	ctx.BezierCurveTo(10, 0, 20, 30, 30, 15)
	ctx.Stroke()

	filled := false
	for i := 3; i < len(c.pixels); i += 4 {
		if c.pixels[i] > 0 {
			filled = true
			break
		}
	}
	if !filled {
		t.Fatal("bezier curve should have filled some pixels")
	}
}

// ---------------------------------------------------------------------------
// Compositing
// ---------------------------------------------------------------------------

func TestCanvasGlobalAlpha(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	ctx.SetGlobalAlpha(0.5)
	ctx.SetFillColor(uimath.Color{R: 1, G: 0, B: 0, A: 1})
	ctx.FillRect(0, 0, 10, 10)

	a := c.pixels[3]
	if a < 120 || a > 135 {
		t.Fatalf("expected alpha ~128, got %d", a)
	}
}

func TestCanvasGlobalAlphaZero(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	ctx.SetGlobalAlpha(0)
	ctx.SetFillColor(uimath.Color{R: 1, A: 1})
	ctx.FillRect(0, 0, 10, 10)
	if c.pixels[3] != 0 {
		t.Fatal("alpha 0 should produce no visible pixels")
	}
}

func TestCanvasCompositeOps(t *testing.T) {
	// CompositeCopy
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	ctx.SetFillColor(uimath.Color{R: 1, A: 0.5})
	ctx.FillRect(0, 0, 10, 10)
	ctx.SetGlobalCompositeOperation(CompositeCopy)
	ctx.SetFillColor(uimath.Color{G: 1, A: 1})
	ctx.FillRect(0, 0, 5, 5)
	i := (2*10 + 2) * 4
	if c.pixels[i] != 0 || c.pixels[i+1] != 255 || c.pixels[i+2] != 0 {
		t.Fatalf("CompositeCopy: got RGB(%d,%d,%d)", c.pixels[i], c.pixels[i+1], c.pixels[i+2])
	}
}

func TestCanvasCompositeLighter(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	ctx.SetFillColor(uimath.Color{R: 0.5, A: 1})
	ctx.FillRect(0, 0, 10, 10)
	ctx.SetGlobalCompositeOperation(CompositeLighter)
	ctx.SetFillColor(uimath.Color{R: 0.5, A: 1})
	ctx.FillRect(0, 0, 10, 10)
	// R should be >= 250 (additive, may have rounding)
	if c.pixels[0] < 250 {
		t.Fatalf("CompositeLighter: R=%d, want >=250", c.pixels[0])
	}
}

func TestCanvasCompositeXOR(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	ctx.SetFillColor(uimath.Color{R: 1, A: 0.5})
	ctx.FillRect(0, 0, 10, 10)
	ctx.SetGlobalCompositeOperation(CompositeXOR)
	ctx.SetFillColor(uimath.Color{G: 1, A: 0.5})
	ctx.FillRect(0, 0, 10, 10)
	// XOR with equal alpha produces some result — just verify it executed
	// The key point is that the XOR code path is exercised
	i := (5*10 + 5) * 4
	_ = c.pixels[i] // no panic
}

func TestCanvasCompositeDefault(t *testing.T) {
	// Hit the default branch with an unimplemented op
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	ctx.SetFillColor(uimath.Color{R: 1, A: 0.5})
	ctx.FillRect(0, 0, 10, 10)
	ctx.SetGlobalCompositeOperation(CompositeSourceIn) // falls to default
	ctx.SetFillColor(uimath.Color{G: 1, A: 0.5})
	ctx.FillRect(0, 0, 10, 10)
	// Should blend (default = source-over fallback)
	if c.pixels[3] == 0 {
		t.Fatal("default composite should produce pixels")
	}
}

// ---------------------------------------------------------------------------
// Pixel manipulation
// ---------------------------------------------------------------------------

func TestCanvasGetPutImageData(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()

	ctx.SetFillColor(uimath.Color{R: 1, A: 1})
	ctx.FillRect(0, 0, 10, 10)

	data := ctx.GetImageData(2, 2, 3, 3)
	if data.Width != 3 || data.Height != 3 {
		t.Fatal("wrong image data size")
	}
	if data.Data[0] != 255 {
		t.Fatalf("expected R=255, got %d", data.Data[0])
	}

	for i := range data.Data {
		data.Data[i] = 0
	}
	ctx.PutImageData(data, 2, 2)

	i := (2*10 + 2) * 4
	if c.pixels[i+3] != 0 {
		t.Fatal("put image data should have cleared the region")
	}
}

func TestCanvasGetImageDataOutOfBounds(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	ctx.SetFillColor(uimath.Color{R: 1, A: 1})
	ctx.FillRect(0, 0, 10, 10)

	// Partially out of bounds
	data := ctx.GetImageData(-2, -2, 5, 5)
	if data.Width != 5 || data.Height != 5 {
		t.Fatal("wrong size")
	}
	// (-2,-2) to (2,2): pixels at data coords (2,2)..(4,4) should have R=255
	i := (2*5 + 2) * 4
	if data.Data[i] != 255 {
		t.Fatalf("in-bounds pixel R=%d, want 255", data.Data[i])
	}
	// (0,0) in data is out of canvas bounds — should be 0
	if data.Data[0] != 0 {
		t.Fatal("out-of-bounds pixel should be 0")
	}
}

func TestCanvasPutImageDataOutOfBounds(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	data := NewImageData(5, 5)
	for i := range data.Data {
		data.Data[i] = 255
	}
	// Partially out of bounds — should not panic
	ctx.PutImageData(data, -2, -2)
	// Pixel at (0,0) should be 255 (from data coords 2,2)
	if c.pixels[3] != 255 {
		t.Fatal("in-bounds portion should be written")
	}
}

func TestCanvasCreateImageData(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	data := ctx.CreateImageData(5, 5)
	if data.Width != 5 || data.Height != 5 || len(data.Data) != 100 {
		t.Fatal("wrong image data dimensions")
	}
}

// ---------------------------------------------------------------------------
// IsPointInPath
// ---------------------------------------------------------------------------

func TestCanvasIsPointInPath(t *testing.T) {
	c := newTestCanvas(20, 20)
	ctx := c.GetContext2D()

	ctx.BeginPath()
	ctx.Rect(5, 5, 10, 10)

	if !ctx.IsPointInPath(10, 10) {
		t.Fatal("(10,10) should be in rect path")
	}
	if ctx.IsPointInPath(0, 0) {
		t.Fatal("(0,0) should NOT be in rect path")
	}
}

// ---------------------------------------------------------------------------
// Gradients
// ---------------------------------------------------------------------------

func TestCanvasLinearGradient(t *testing.T) {
	c := newTestCanvas(20, 10)
	ctx := c.GetContext2D()

	g := ctx.CreateLinearGradient(0, 0, 20, 0)
	g.AddColorStop(0, uimath.Color{R: 1, A: 1})
	g.AddColorStop(1, uimath.Color{B: 1, A: 1})
	ctx.SetFillStyleGradient(g)
	ctx.FillRect(0, 0, 20, 10)

	if c.pixels[0] < 200 {
		t.Fatalf("left pixel R=%d, want mostly red", c.pixels[0])
	}
	i := (5*20 + 19) * 4
	if c.pixels[i+2] < 200 {
		t.Fatalf("right pixel B=%d, want mostly blue", c.pixels[i+2])
	}
}

func TestCanvasLinearGradientZeroLength(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	g := ctx.CreateLinearGradient(5, 5, 5, 5) // zero length
	g.AddColorStop(0, uimath.Color{R: 1, A: 1})
	g.AddColorStop(1, uimath.Color{B: 1, A: 1})
	ctx.SetFillStyleGradient(g)
	ctx.FillRect(0, 0, 10, 10)
	// Should use first stop color
	if c.pixels[0] < 200 {
		t.Fatalf("zero-length gradient should use first stop, R=%d", c.pixels[0])
	}
}

func TestCanvasRadialGradient(t *testing.T) {
	c := newTestCanvas(30, 30)
	ctx := c.GetContext2D()
	g := ctx.CreateRadialGradient(15, 15, 0, 15, 15, 15)
	g.AddColorStop(0, uimath.Color{R: 1, A: 1})
	g.AddColorStop(1, uimath.Color{B: 1, A: 1})
	ctx.SetFillStyleGradient(g)
	ctx.FillRect(0, 0, 30, 30)

	// Center should be reddish
	ci := (15*30 + 15) * 4
	if c.pixels[ci] < 200 {
		t.Fatalf("center R=%d, want mostly red", c.pixels[ci])
	}
}

func TestCanvasRadialGradientZeroRange(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	g := ctx.CreateRadialGradient(5, 5, 5, 5, 5, 5) // r0 == r1
	g.AddColorStop(0, uimath.Color{R: 1, A: 1})
	ctx.SetFillStyleGradient(g)
	ctx.FillRect(0, 0, 10, 10) // should not panic
}

func TestCanvasGradientColorAt(t *testing.T) {
	g := &CanvasGradient{linear: true}

	// No stops
	c := g.colorAt(0.5)
	if c != (uimath.Color{}) {
		t.Fatal("no stops should return zero color")
	}

	// Single stop
	g.AddColorStop(0.5, uimath.Color{R: 1, A: 1})
	c = g.colorAt(0)
	if c.R != 1 {
		t.Fatal("single stop: t<offset should return that stop")
	}
	c = g.colorAt(1)
	if c.R != 1 {
		t.Fatal("single stop: t>offset should return that stop")
	}

	// Multiple stops — test interpolation
	g2 := &CanvasGradient{linear: true}
	g2.AddColorStop(0, uimath.Color{R: 1, A: 1})
	g2.AddColorStop(0.5, uimath.Color{G: 1, A: 1})
	g2.AddColorStop(1, uimath.Color{B: 1, A: 1})

	c = g2.colorAt(0.25) // between stop 0 and 0.5
	if c.R < 0.3 || c.G < 0.3 {
		t.Fatalf("mid-interpolation unexpected: %v", c)
	}
	c = g2.colorAt(1.5) // beyond last
	if c.B != 1 {
		t.Fatal("beyond last stop should return last")
	}
	c = g2.colorAt(-0.5) // before first
	if c.R != 1 {
		t.Fatal("before first stop should return first")
	}
}

// ---------------------------------------------------------------------------
// Text
// ---------------------------------------------------------------------------

func TestCanvasFillTextNoRenderer(t *testing.T) {
	c := newTestCanvas(50, 50)
	ctx := c.GetContext2D()
	ctx.SetFont(12)
	ctx.FillText("hello", 5, 20) // No TextRenderer — should not panic, no pixels
}

func TestCanvasFillTextWithRenderer(t *testing.T) {
	c := newTestCanvasWithText(100, 50)
	ctx := c.GetContext2D()
	ctx.SetFont(12)
	ctx.SetFillColor(uimath.Color{R: 1, A: 1})
	ctx.FillText("Hi", 5, 30)

	hasPixels := false
	for i := 3; i < len(c.pixels); i += 4 {
		if c.pixels[i] > 0 {
			hasPixels = true
			break
		}
	}
	if !hasPixels {
		t.Fatal("FillText with renderer should produce pixels")
	}
}

func TestCanvasFillTextAlignCenter(t *testing.T) {
	c := newTestCanvasWithText(100, 50)
	ctx := c.GetContext2D()
	ctx.SetFont(12)
	ctx.SetTextAlign(CanvasTextAlignCenter)
	ctx.SetFillColor(uimath.Color{R: 1, A: 1})
	ctx.FillText("AB", 50, 30)
	if !c.dirty {
		t.Fatal("should be dirty after FillText")
	}
}

func TestCanvasFillTextAlignRight(t *testing.T) {
	c := newTestCanvasWithText(100, 50)
	ctx := c.GetContext2D()
	ctx.SetFont(12)
	ctx.SetTextAlign(CanvasTextAlignRight)
	ctx.SetFillColor(uimath.Color{R: 1, A: 1})
	ctx.FillText("AB", 90, 30)
	if !c.dirty {
		t.Fatal("should be dirty")
	}
}

func TestCanvasFillTextAlignEnd(t *testing.T) {
	c := newTestCanvasWithText(100, 50)
	ctx := c.GetContext2D()
	ctx.SetFont(12)
	ctx.SetTextAlign(CanvasTextAlignEnd)
	ctx.SetFillColor(uimath.Color{R: 1, A: 1})
	ctx.FillText("AB", 90, 30)
	if !c.dirty {
		t.Fatal("should be dirty")
	}
}

func TestCanvasFillTextBaselineTop(t *testing.T) {
	c := newTestCanvasWithText(100, 50)
	ctx := c.GetContext2D()
	ctx.SetFont(12)
	ctx.SetTextBaseline(CanvasTextBaselineTop)
	ctx.SetFillColor(uimath.Color{R: 1, A: 1})
	ctx.FillText("AB", 5, 5)
}

func TestCanvasFillTextBaselineHanging(t *testing.T) {
	c := newTestCanvasWithText(100, 50)
	ctx := c.GetContext2D()
	ctx.SetFont(12)
	ctx.SetTextBaseline(CanvasTextBaselineHanging)
	ctx.SetFillColor(uimath.Color{R: 1, A: 1})
	ctx.FillText("AB", 5, 5)
}

func TestCanvasFillTextBaselineMiddle(t *testing.T) {
	c := newTestCanvasWithText(100, 50)
	ctx := c.GetContext2D()
	ctx.SetFont(12)
	ctx.SetTextBaseline(CanvasTextBaselineMiddle)
	ctx.SetFillColor(uimath.Color{R: 1, A: 1})
	ctx.FillText("AB", 5, 25)
}

func TestCanvasFillTextBaselineBottom(t *testing.T) {
	c := newTestCanvasWithText(100, 50)
	ctx := c.GetContext2D()
	ctx.SetFont(12)
	ctx.SetTextBaseline(CanvasTextBaselineBottom)
	ctx.SetFillColor(uimath.Color{R: 1, A: 1})
	ctx.FillText("AB", 5, 45)
}

func TestCanvasFillTextBaselineIdeographic(t *testing.T) {
	c := newTestCanvasWithText(100, 50)
	ctx := c.GetContext2D()
	ctx.SetFont(12)
	ctx.SetTextBaseline(CanvasTextBaselineIdeographic)
	ctx.SetFillColor(uimath.Color{R: 1, A: 1})
	ctx.FillText("AB", 5, 45)
}

func TestCanvasMeasureText(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	ctx.SetFont(12)
	m := ctx.MeasureText("hello")
	if m.Width <= 0 {
		t.Fatal("MeasureText should return positive width")
	}
}

func TestCanvasMeasureTextWithRenderer(t *testing.T) {
	c := newTestCanvasWithText(100, 50)
	ctx := c.GetContext2D()
	ctx.SetFont(14)
	m := ctx.MeasureText("hello")
	if m.Width <= 0 {
		t.Fatal("MeasureText with renderer should return positive width")
	}
}

// ---------------------------------------------------------------------------
// DrawImage / DrawImageScaled
// ---------------------------------------------------------------------------

func TestCanvasDrawImage(t *testing.T) {
	c := newTestCanvas(20, 20)
	ctx := c.GetContext2D()

	src := NewImageData(5, 5)
	for i := 0; i < 5*5*4; i += 4 {
		src.Data[i] = 255
		src.Data[i+3] = 255
	}

	ctx.DrawImage(src, 3, 3)

	i := (5*20 + 5) * 4
	if c.pixels[i] != 255 || c.pixels[i+3] != 255 {
		t.Fatalf("drawImage pixel at (5,5): R=%d A=%d", c.pixels[i], c.pixels[i+3])
	}
}

func TestCanvasDrawImageScaled(t *testing.T) {
	c := newTestCanvas(20, 20)
	ctx := c.GetContext2D()

	src := NewImageData(2, 2)
	for i := 0; i < 2*2*4; i += 4 {
		src.Data[i] = 255   // R
		src.Data[i+3] = 255 // A
	}

	// Scale 2x2 → 10x10
	ctx.DrawImageScaled(src, 0, 0, 10, 10)

	i := (5*20 + 5) * 4
	if c.pixels[i] != 255 || c.pixels[i+3] != 255 {
		t.Fatalf("scaled image pixel at (5,5): R=%d A=%d", c.pixels[i], c.pixels[i+3])
	}
}

// ---------------------------------------------------------------------------
// Clip
// ---------------------------------------------------------------------------

func TestCanvasClip(t *testing.T) {
	c := newTestCanvas(20, 20)
	ctx := c.GetContext2D()

	ctx.BeginPath()
	ctx.Rect(5, 5, 10, 10)
	ctx.Clip()

	ctx.SetFillColor(uimath.Color{R: 1, A: 1})
	ctx.FillRect(0, 0, 20, 20)

	if c.pixels[3] != 0 {
		t.Fatal("pixel at (0,0) should be clipped")
	}
	i := (10*20 + 10) * 4
	if c.pixels[i+3] == 0 {
		t.Fatal("pixel at (10,10) should be filled within clip")
	}
}

func TestCanvasClipEmpty(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	ctx.BeginPath()
	ctx.Clip() // empty path — should be no-op, no clip
	ctx.SetFillColor(uimath.Color{R: 1, A: 1})
	ctx.FillRect(0, 0, 10, 10)
	if c.pixels[3] == 0 {
		t.Fatal("empty clip should not restrict drawing")
	}
}

func TestCanvasClipRestored(t *testing.T) {
	c := newTestCanvas(20, 20)
	ctx := c.GetContext2D()

	ctx.Save()
	ctx.BeginPath()
	ctx.Rect(5, 5, 10, 10)
	ctx.Clip()
	ctx.Restore()

	// After restore, clip should be gone
	ctx.SetFillColor(uimath.Color{R: 1, A: 1})
	ctx.FillRect(0, 0, 20, 20)
	if c.pixels[3] == 0 {
		t.Fatal("clip should be restored away")
	}
}

// ---------------------------------------------------------------------------
// Go image interop
// ---------------------------------------------------------------------------

func TestCanvasToImage(t *testing.T) {
	c := newTestCanvas(4, 4)
	ctx := c.GetContext2D()
	ctx.SetFillColor(uimath.Color{R: 1, G: 1, B: 1, A: 1})
	ctx.FillRect(0, 0, 4, 4)

	img := c.ToImage()
	if img.Bounds().Dx() != 4 || img.Bounds().Dy() != 4 {
		t.Fatal("image size mismatch")
	}
	r, g, b, a := img.At(2, 2).RGBA()
	if r>>8 != 255 || g>>8 != 255 || b>>8 != 255 || a>>8 != 255 {
		t.Fatalf("pixel at (2,2): got RGBA(%d,%d,%d,%d)", r>>8, g>>8, b>>8, a>>8)
	}
}

func TestCanvasPutGoImage(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()

	img := image.NewRGBA(image.Rect(0, 0, 3, 3))
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			img.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}

	ctx.PutGoImage(img, 2, 2)
	i := (3*10 + 3) * 4
	if c.pixels[i] != 255 || c.pixels[i+3] != 255 {
		t.Fatalf("PutGoImage pixel at (3,3): R=%d A=%d", c.pixels[i], c.pixels[i+3])
	}
}

func TestCanvasPutGoImageOutOfBounds(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	img := image.NewRGBA(image.Rect(0, 0, 5, 5))
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			img.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	ctx.PutGoImage(img, -2, -2) // should not panic
	if c.pixels[3] != 255 {
		t.Fatal("in-bounds portion should be written")
	}
}

func TestCanvasToGoColor(t *testing.T) {
	c := ToGoColor(uimath.Color{R: 1, G: 0.5, B: 0, A: 1})
	if c.R != 255 || c.G != 127 || c.B != 0 || c.A != 255 {
		t.Fatalf("ToGoColor: got %v", c)
	}
}

// ---------------------------------------------------------------------------
// Draw without backend
// ---------------------------------------------------------------------------

func TestCanvasDrawNoBackend(t *testing.T) {
	c := newTestCanvas(10, 10)
	buf := render.NewCommandBuffer()
	c.Draw(buf)
	if buf.Len() != 0 {
		t.Fatalf("expected 0 commands with no backend, got %d", buf.Len())
	}
}

// ---------------------------------------------------------------------------
// Destroy
// ---------------------------------------------------------------------------

func TestCanvasDestroy(t *testing.T) {
	c := newTestCanvas(10, 10)
	c.Destroy()
}

func TestCanvasDestroyWithBackend(t *testing.T) {
	c, mb := newTestCanvasWithBackend(10, 10)
	c.Sync()
	c.Destroy()
	if mb.destroyed != 1 {
		t.Fatalf("texture not destroyed, got %d", mb.destroyed)
	}
}

// ---------------------------------------------------------------------------
// Bresenham directions (vertical, horizontal, diagonal, reverse)
// ---------------------------------------------------------------------------

func TestCanvasBresenhamDirections(t *testing.T) {
	c := newTestCanvas(20, 20)
	ctx := c.GetContext2D()
	ctx.SetStrokeColor(uimath.Color{R: 1, A: 1})
	ctx.SetLineWidth(1)

	// Vertical line (x constant)
	ctx.BeginPath()
	ctx.MoveTo(5, 0)
	ctx.LineTo(5, 19)
	ctx.Stroke()
	if c.pixels[(10*20+5)*4+3] == 0 {
		t.Fatal("vertical line should have pixels")
	}

	// Horizontal line (y constant) — right to left
	ctx.BeginPath()
	ctx.MoveTo(19, 15)
	ctx.LineTo(0, 15)
	ctx.Stroke()
	if c.pixels[(15*20+10)*4+3] == 0 {
		t.Fatal("horizontal R→L line should have pixels")
	}

	// Diagonal: bottom-left to top-right
	ctx.BeginPath()
	ctx.MoveTo(0, 19)
	ctx.LineTo(19, 0)
	ctx.Stroke()
	if c.pixels[(10*20+9)*4+3] == 0 {
		t.Fatal("diagonal line should have pixels")
	}
}

// ---------------------------------------------------------------------------
// Thick line (zero length)
// ---------------------------------------------------------------------------

func TestCanvasThickLineZeroLength(t *testing.T) {
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	ctx.SetStrokeColor(uimath.Color{R: 1, A: 1})
	ctx.SetLineWidth(3)
	ctx.BeginPath()
	ctx.MoveTo(5, 5)
	ctx.LineTo(5, 5) // zero length
	ctx.Stroke()
	// Should not panic
}

// ---------------------------------------------------------------------------
// setPixel at boundaries
// ---------------------------------------------------------------------------

func TestCanvasSetPixelOutOfBounds(t *testing.T) {
	c := newTestCanvas(5, 5)
	ctx := c.GetContext2D()
	// These should not panic
	ctx.setPixel(-1, 0, uimath.Color{R: 1, A: 1})
	ctx.setPixel(0, -1, uimath.Color{R: 1, A: 1})
	ctx.setPixel(5, 0, uimath.Color{R: 1, A: 1})
	ctx.setPixel(0, 5, uimath.Color{R: 1, A: 1})
}

// ---------------------------------------------------------------------------
// min4 / max4 / abs full coverage
// ---------------------------------------------------------------------------

func TestCanvasHelpers(t *testing.T) {
	// min4
	if min4(1, 2, 3, 4) != 1 {
		t.Fatal("min4 first")
	}
	if min4(4, 1, 3, 2) != 1 {
		t.Fatal("min4 second")
	}
	if min4(4, 3, 1, 2) != 1 {
		t.Fatal("min4 third")
	}
	if min4(4, 3, 2, 1) != 1 {
		t.Fatal("min4 fourth")
	}

	// max4
	if max4(4, 3, 2, 1) != 4 {
		t.Fatal("max4 first")
	}
	if max4(1, 4, 2, 3) != 4 {
		t.Fatal("max4 second")
	}
	if max4(1, 2, 4, 3) != 4 {
		t.Fatal("max4 third")
	}
	if max4(1, 2, 3, 4) != 4 {
		t.Fatal("max4 fourth")
	}

	// abs
	if abs(5) != 5 {
		t.Fatal("abs positive")
	}
	if abs(-5) != 5 {
		t.Fatal("abs negative")
	}
	if abs(0) != 0 {
		t.Fatal("abs zero")
	}
}

// ---------------------------------------------------------------------------
// clampByte edge cases
// ---------------------------------------------------------------------------

func TestCanvasClampByte(t *testing.T) {
	if clampByte(-0.5) != 0 {
		t.Fatal("clampByte negative")
	}
	if clampByte(0) != 0 {
		t.Fatal("clampByte zero")
	}
	if clampByte(1.5) != 255 {
		t.Fatal("clampByte > 1")
	}
	if clampByte(0.5) != 127 {
		t.Fatalf("clampByte 0.5: got %d", clampByte(0.5))
	}
}

// ---------------------------------------------------------------------------
// Rect fill with gradient stroke
// ---------------------------------------------------------------------------

func TestCanvasStrokeRectGradient(t *testing.T) {
	c := newTestCanvas(20, 20)
	ctx := c.GetContext2D()
	g := ctx.CreateLinearGradient(0, 0, 20, 0)
	g.AddColorStop(0, uimath.Color{R: 1, A: 1})
	g.AddColorStop(1, uimath.Color{B: 1, A: 1})
	ctx.SetStrokeStyleGradient(g)
	ctx.SetLineWidth(2)
	ctx.StrokeRect(3, 3, 14, 14)
	hasColor := false
	for i := 3; i < len(c.pixels); i += 4 {
		if c.pixels[i] > 0 {
			hasColor = true
			break
		}
	}
	if !hasColor {
		t.Fatal("gradient stroke rect should produce pixels")
	}
}

// ---------------------------------------------------------------------------
// Fill path with gradient
// ---------------------------------------------------------------------------

func TestCanvasFillPathGradient(t *testing.T) {
	c := newTestCanvas(20, 20)
	ctx := c.GetContext2D()
	g := ctx.CreateLinearGradient(0, 0, 20, 0)
	g.AddColorStop(0, uimath.Color{R: 1, A: 1})
	g.AddColorStop(1, uimath.Color{B: 1, A: 1})
	ctx.SetFillStyleGradient(g)
	ctx.BeginPath()
	ctx.Rect(2, 2, 16, 16)
	ctx.Fill()
	if c.pixels[(10*20+2)*4+3] == 0 {
		t.Fatal("filled path with gradient should produce pixels")
	}
}

// ---------------------------------------------------------------------------
// Gradient edge cases in rasterization
// ---------------------------------------------------------------------------

func TestCanvasRadialGradientClampT(t *testing.T) {
	// Radial gradient where dist < r0 (t < 0) and dist > r1 (t > 1)
	c := newTestCanvas(30, 30)
	ctx := c.GetContext2D()
	g := ctx.CreateRadialGradient(15, 15, 5, 15, 15, 10)
	g.AddColorStop(0, uimath.Color{R: 1, A: 1})
	g.AddColorStop(1, uimath.Color{B: 1, A: 1})
	ctx.SetFillStyleGradient(g)
	ctx.FillRect(0, 0, 30, 30)
	// Center (dist=0 < r0=5) should clamp to first stop (red)
	ci := (15*30 + 15) * 4
	if c.pixels[ci] < 200 {
		t.Fatalf("center R=%d, want red (t clamped to 0)", c.pixels[ci])
	}
}

func TestCanvasLinearGradientClampT(t *testing.T) {
	// Linear gradient where some pixels project before start or after end
	c := newTestCanvas(30, 10)
	ctx := c.GetContext2D()
	g := ctx.CreateLinearGradient(10, 0, 20, 0) // gradient only covers x=10..20
	g.AddColorStop(0, uimath.Color{R: 1, A: 1})
	g.AddColorStop(1, uimath.Color{B: 1, A: 1})
	ctx.SetFillStyleGradient(g)
	ctx.FillRect(0, 0, 30, 10)
	// x=0 projects to t<0, should clamp → first stop (red)
	if c.pixels[0] < 200 {
		t.Fatalf("left R=%d, want red (t<0 clamped)", c.pixels[0])
	}
	// x=29 projects to t>1, should clamp → last stop (blue)
	i := (5*30 + 29) * 4
	if c.pixels[i+2] < 200 {
		t.Fatalf("right B=%d, want blue (t>1 clamped)", c.pixels[i+2])
	}
}

func TestCanvasGradientStrokeClamp(t *testing.T) {
	// Stroke rect with gradient to exercise rasterFillRectRaw with gradient and boundary clamp
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	g := ctx.CreateLinearGradient(0, 0, 10, 0)
	g.AddColorStop(0, uimath.Color{R: 1, A: 1})
	g.AddColorStop(1, uimath.Color{B: 1, A: 1})
	ctx.SetStrokeStyleGradient(g)
	ctx.SetLineWidth(2)
	// Partially out of bounds to hit clamp paths
	ctx.StrokeRect(-1, -1, 12, 12)
	if c.pixels[3] == 0 {
		t.Fatal("stroke should produce pixels")
	}
}

func TestCanvasRasterFillRectRawBoundary(t *testing.T) {
	// Fill rect that extends beyond canvas bounds to hit all clamp paths
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	ctx.SetFillColor(uimath.Color{R: 1, A: 1})
	ctx.FillRect(-5, -5, 20, 20) // extends beyond all edges
	if c.pixels[3] == 0 {
		t.Fatal("should fill visible region")
	}
}

func TestCanvasThickLineBoundary(t *testing.T) {
	// Thick line that extends beyond canvas
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	ctx.SetStrokeColor(uimath.Color{R: 1, A: 1})
	ctx.SetLineWidth(4)
	ctx.BeginPath()
	ctx.MoveTo(-5, 5)
	ctx.LineTo(15, 5)
	ctx.Stroke()
	// Some pixels should be drawn
	if c.pixels[(5*10+5)*4+3] == 0 {
		t.Fatal("thick line should have pixels in bounds")
	}
}

func TestCanvasRasterFillBoundary(t *testing.T) {
	// Path fill that extends beyond canvas
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	ctx.SetFillColor(uimath.Color{R: 1, A: 1})
	ctx.BeginPath()
	ctx.MoveTo(-5, -5)
	ctx.LineTo(15, -5)
	ctx.LineTo(15, 15)
	ctx.LineTo(-5, 15)
	ctx.ClosePath()
	ctx.Fill()
	if c.pixels[(5*10+5)*4+3] == 0 {
		t.Fatal("should fill visible portion")
	}
}

func TestCanvasArcSmallSweep(t *testing.T) {
	// Very small arc (< pi/16) to test n=1 segment case
	c := newTestCanvas(20, 20)
	ctx := c.GetContext2D()
	ctx.SetStrokeColor(uimath.Color{R: 1, A: 1})
	ctx.SetLineWidth(1)
	ctx.BeginPath()
	ctx.Arc(10, 10, 5, 0, 0.01, false)
	ctx.Stroke()
}

func TestCanvasArcCCWSweepAdjust(t *testing.T) {
	// CCW arc where sweep starts positive
	c := newTestCanvas(30, 30)
	ctx := c.GetContext2D()
	ctx.SetStrokeColor(uimath.Color{R: 1, A: 1})
	ctx.SetLineWidth(1)
	ctx.BeginPath()
	ctx.Arc(15, 15, 8, 0, math.Pi*1.5, true)
	ctx.Stroke()
	hasPixels := false
	for i := 3; i < len(c.pixels); i += 4 {
		if c.pixels[i] > 0 {
			hasPixels = true
			break
		}
	}
	if !hasPixels {
		t.Fatal("CCW arc with sweep>0 should produce pixels")
	}
}

func TestCanvasArcCWSweepAdjust(t *testing.T) {
	// CW arc where endAngle < startAngle (negative sweep adjusted positive)
	c := newTestCanvas(30, 30)
	ctx := c.GetContext2D()
	ctx.SetStrokeColor(uimath.Color{R: 1, A: 1})
	ctx.SetLineWidth(1)
	ctx.BeginPath()
	ctx.Arc(15, 15, 8, math.Pi, 0, false) // CW, end < start
	ctx.Stroke()
	hasPixels := false
	for i := 3; i < len(c.pixels); i += 4 {
		if c.pixels[i] > 0 {
			hasPixels = true
			break
		}
	}
	if !hasPixels {
		t.Fatal("CW arc with negative sweep should produce pixels")
	}
}

func TestCanvasDrawImageScaledBeyondSrc(t *testing.T) {
	// Scale where some dest pixels map beyond src dimensions
	c := newTestCanvas(20, 20)
	ctx := c.GetContext2D()
	src := NewImageData(2, 2)
	src.Data[0] = 255
	src.Data[3] = 255
	// dw/dh < src size causes scaleX/Y > 1, potentially mapping beyond
	ctx.DrawImageScaled(src, 0, 0, 1, 1) // scale down
	// Should not panic
}

func TestCanvasGradientColorAtFallthrough(t *testing.T) {
	// 3 stops, t exactly at last boundary — exercises the final return
	g := &CanvasGradient{linear: true}
	g.AddColorStop(0, uimath.Color{R: 1, A: 1})
	g.AddColorStop(0.5, uimath.Color{G: 1, A: 1})
	g.AddColorStop(1, uimath.Color{B: 1, A: 1})

	// t = 0.5 exactly at second stop
	c := g.colorAt(0.5)
	if c.G < 0.9 {
		t.Fatalf("at stop 0.5: G=%f, want ~1", c.G)
	}
}

func TestCanvasArcOnExistingPath(t *testing.T) {
	// Arc when path already has segments (i==0 && len(path)>0 → LineTo branch)
	c := newTestCanvas(30, 30)
	ctx := c.GetContext2D()
	ctx.SetStrokeColor(uimath.Color{R: 1, A: 1})
	ctx.SetLineWidth(1)
	ctx.BeginPath()
	ctx.MoveTo(0, 15)
	ctx.LineTo(5, 15) // path is non-empty
	ctx.Arc(15, 15, 8, math.Pi, 0, false)
	ctx.Stroke()
	hasPixels := false
	for i := 3; i < len(c.pixels); i += 4 {
		if c.pixels[i] > 0 {
			hasPixels = true
			break
		}
	}
	if !hasPixels {
		t.Fatal("arc on existing path should produce pixels")
	}
}

func TestCanvasFillOnlyMoveTo(t *testing.T) {
	// Path with only MoveTo → no edges → early return
	c := newTestCanvas(10, 10)
	ctx := c.GetContext2D()
	ctx.SetFillColor(uimath.Color{R: 1, A: 1})
	ctx.BeginPath()
	ctx.MoveTo(5, 5) // only moveTo, no lineTo
	ctx.Fill()
	// Should not fill anything
	if c.pixels[(5*10+5)*4+3] != 0 {
		t.Fatal("fill with only moveTo should not produce pixels")
	}
}

func TestCanvasArcZeroSweep(t *testing.T) {
	// Arc with startAngle == endAngle (sweep=0, n=0→1)
	c := newTestCanvas(20, 20)
	ctx := c.GetContext2D()
	ctx.SetStrokeColor(uimath.Color{R: 1, A: 1})
	ctx.BeginPath()
	ctx.Arc(10, 10, 5, math.Pi/4, math.Pi/4, false) // zero sweep
	ctx.Stroke()
	// Should not panic
}

func TestCanvasArcCCWAlreadyNegative(t *testing.T) {
	// CCW arc where sweep is already negative (endAngle < startAngle)
	c := newTestCanvas(30, 30)
	ctx := c.GetContext2D()
	ctx.SetStrokeColor(uimath.Color{R: 1, A: 1})
	ctx.SetLineWidth(1)
	ctx.BeginPath()
	ctx.Arc(15, 15, 8, math.Pi, 0, true) // ccw, sweep = 0 - Pi < 0
	ctx.Stroke()
	hasPixels := false
	for i := 3; i < len(c.pixels); i += 4 {
		if c.pixels[i] > 0 {
			hasPixels = true
			break
		}
	}
	if !hasPixels {
		t.Fatal("ccw arc with already-negative sweep should produce pixels")
	}
}

func TestCanvasThickLineOutOfBoundsX(t *testing.T) {
	// Thick line where ix0 < 0 and ix1 > width
	c := newTestCanvas(5, 5)
	ctx := c.GetContext2D()
	ctx.SetStrokeColor(uimath.Color{R: 1, A: 1})
	ctx.SetLineWidth(4)
	ctx.BeginPath()
	ctx.MoveTo(-10, 2)
	ctx.LineTo(15, 2)
	ctx.Stroke()
	// Should not panic and have some pixels
	if c.pixels[(2*5+2)*4+3] == 0 {
		t.Fatal("thick line crossing bounds should produce pixels")
	}
}

func TestCanvasThickLineOutOfBoundsY(t *testing.T) {
	// Thick vertical line outside Y range
	c := newTestCanvas(5, 5)
	ctx := c.GetContext2D()
	ctx.SetStrokeColor(uimath.Color{R: 1, A: 1})
	ctx.SetLineWidth(4)
	ctx.BeginPath()
	ctx.MoveTo(2, -10)
	ctx.LineTo(2, 15)
	ctx.Stroke()
	if c.pixels[(2*5+2)*4+3] == 0 {
		t.Fatal("thick line crossing Y bounds should have pixels")
	}
}

func TestCanvasRasterFillXClamp(t *testing.T) {
	// Path fill where scanline X intersections extend beyond canvas
	c := newTestCanvas(5, 5)
	ctx := c.GetContext2D()
	ctx.SetFillColor(uimath.Color{R: 1, A: 1})
	ctx.BeginPath()
	ctx.MoveTo(-10, 0)
	ctx.LineTo(15, 0)
	ctx.LineTo(15, 5)
	ctx.LineTo(-10, 5)
	ctx.ClosePath()
	ctx.Fill()
	if c.pixels[(2*5+2)*4+3] == 0 {
		t.Fatal("fill with extended X should work within bounds")
	}
}

func TestCanvasDrawImageScaledSkipOutOfSrc(t *testing.T) {
	// Very small src scaled very large — some dest pixels map beyond src
	c := newTestCanvas(20, 20)
	ctx := c.GetContext2D()
	src := NewImageData(1, 1)
	src.Data[0] = 255
	src.Data[3] = 255
	ctx.DrawImageScaled(src, 0, 0, 20, 20) // 1x1 → 20x20
	if c.pixels[(10*20+10)*4+3] == 0 {
		t.Fatal("scaled 1x1→20x20 should fill center")
	}
}

func TestCanvasDrawImageScaledExceedSrc(t *testing.T) {
	// Scale where some dest pixels map to sx >= src.Width
	c := newTestCanvas(20, 20)
	ctx := c.GetContext2D()
	src := NewImageData(3, 3)
	for i := range src.Data {
		src.Data[i] = 128
	}
	// dw=10, src.Width=3: scaleX = 3/10 = 0.3
	// px=10 → sx = int(3.0) = 3 >= 3 → continue (skip)
	ctx.DrawImageScaled(src, 0, 0, 10, 10)
}

func TestCanvasDrawImageScaledDownExceed(t *testing.T) {
	// Scale DOWN: dw < src.Width → scaleX > 1 → sx can exceed src.Width
	// src.Width=10, dw=3: scaleX = 10/3 = 3.33
	// px=2 → sx = int(2*3.33) = int(6.66) = 6 < 10 ✓
	// px=3 → beyond loop (px < idw=3)
	// Actually px never reaches 3 since loop is px < 3
	// Try src.Width=5, dw=2: scaleX = 5/2 = 2.5
	// px=1 → sx = int(2.5) = 2 < 5 ✓
	// The continue is effectively unreachable when dw = int(dw_float)
	// because px goes 0..dw-1 and sx = int(px * src.Width/dw) < src.Width
	// when px < dw. It can only trigger with floating point edge cases.
	c := newTestCanvas(20, 20)
	ctx := c.GetContext2D()
	src := NewImageData(5, 5)
	for i := range src.Data {
		src.Data[i] = 128
	}
	ctx.DrawImageScaled(src, 0, 0, 3, 3) // scale down 5→3
}

func TestCanvasGradientUnsortedStops(t *testing.T) {
	// Unsorted stops cause colorAt loop to fall through to final return
	g := &CanvasGradient{linear: true}
	g.stops = []GradientStop{
		{Offset: 0.8, Color: uimath.Color{R: 1, A: 1}},
		{Offset: 0.2, Color: uimath.Color{B: 1, A: 1}},
	}
	// t=0.5: not <= stops[0].Offset(0.8)? Actually 0.5 <= 0.8 → returns from loop
	// Need t between first and last but not matching any ordered stop
	// With stops [0.8, 0.2], t=0.5:
	//   t <= stops[0].Offset(0.8) → true, returns first
	// Try t=0.9:
	//   t >= last.Offset(0.2) → true, returns last (line 222)
	// Actually this function is hard to reach line 231 because:
	//   - if t <= first stop → return first
	//   - if t >= last stop → return last
	//   - loop always finds a stop where t <= stops[i].Offset
	// Line 231 is truly unreachable for sorted stops
	// For unsorted: stops = [0.2, 0.8] with t=0.5
	//   first check: t(0.5) <= stops[0].Offset(0.2) → false
	//   second check: t(0.5) >= last.Offset(0.8) → false
	//   loop i=1: t(0.5) <= stops[1].Offset(0.8) → true → returns
	// The line is indeed unreachable in normal cases.
	c := g.colorAt(0.5)
	_ = c
}

// ---------------------------------------------------------------------------
// Stroke path with gradient
// ---------------------------------------------------------------------------

func TestCanvasStrokePathGradient(t *testing.T) {
	c := newTestCanvas(20, 20)
	ctx := c.GetContext2D()
	g := ctx.CreateLinearGradient(0, 0, 20, 20)
	g.AddColorStop(0, uimath.Color{R: 1, A: 1})
	g.AddColorStop(1, uimath.Color{B: 1, A: 1})
	ctx.SetStrokeStyleGradient(g)
	ctx.SetLineWidth(2)
	ctx.BeginPath()
	ctx.MoveTo(0, 10)
	ctx.LineTo(20, 10)
	ctx.Stroke()
	hasColor := false
	for i := 3; i < len(c.pixels); i += 4 {
		if c.pixels[i] > 0 {
			hasColor = true
			break
		}
	}
	if !hasColor {
		t.Fatal("stroked path with gradient should produce pixels")
	}
}

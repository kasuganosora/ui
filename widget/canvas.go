package widget

import (
	"image"
	"image/color"
	"math"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Canvas is a pixel-level drawing surface with W3C Canvas 2D Context API.
// It maintains an in-memory RGBA pixel buffer and uploads to GPU via Backend.
// The API is platform-independent — Backend abstracts the GPU layer.
type Canvas struct {
	Base
	width   int
	height  int
	pixels  []byte // RGBA, length = width * height * 4
	texture render.TextureHandle
	backend render.Backend
	dirty   bool
	ctx     *CanvasContext2D
}

// NewCanvas creates a canvas widget with the given pixel dimensions.
// backend is used for GPU texture upload; pass nil for offscreen/test use.
func NewCanvas(tree *core.Tree, width, height int, backend render.Backend, cfg *Config) *Canvas {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	c := &Canvas{
		Base:    NewBase(tree, core.TypeCustom, cfg),
		width:   width,
		height:  height,
		pixels:  make([]byte, width*height*4),
		backend: backend,
		dirty:   true,
	}
	c.style.Display = layout.DisplayBlock
	c.style.Width = layout.Px(float32(width))
	c.style.Height = layout.Px(float32(height))
	c.ctx = newCanvasContext2D(c)
	return c
}

func (c *Canvas) Width() int                    { return c.width }
func (c *Canvas) Height() int                   { return c.height }
func (c *Canvas) GetContext2D() *CanvasContext2D { return c.ctx }

// Pixels returns the raw RGBA pixel buffer (read-only view).
func (c *Canvas) Pixels() []byte { return c.pixels }

// SetSize resizes the canvas pixel buffer. Existing content is lost.
func (c *Canvas) SetSize(w, h int) {
	if w == c.width && h == c.height {
		return
	}
	c.width = w
	c.height = h
	c.pixels = make([]byte, w*h*4)
	c.style.Width = layout.Px(float32(w))
	c.style.Height = layout.Px(float32(h))
	// Destroy old texture; a new one will be created on next Sync
	if c.texture != render.InvalidTexture && c.backend != nil {
		c.backend.DestroyTexture(c.texture)
		c.texture = render.InvalidTexture
	}
	c.dirty = true
}

// Sync uploads the pixel buffer to the GPU texture.
// Called automatically by Draw, but can be called explicitly.
func (c *Canvas) Sync() {
	if !c.dirty || c.backend == nil {
		return
	}
	if c.texture == render.InvalidTexture {
		tex, err := c.backend.CreateTexture(render.TextureDesc{
			Width:  c.width,
			Height: c.height,
			Format: render.TextureFormatRGBA8,
			Filter: render.TextureFilterNearest,
			Data:   c.pixels,
		})
		if err == nil {
			c.texture = tex
		}
	} else {
		_ = c.backend.UpdateTexture(
			c.texture,
			uimath.NewRect(0, 0, float32(c.width), float32(c.height)),
			c.pixels,
		)
	}
	c.dirty = false
}

func (c *Canvas) Draw(buf *render.CommandBuffer) {
	c.Sync()
	bounds := c.Bounds()
	if bounds.IsEmpty() || c.texture == render.InvalidTexture {
		return
	}
	buf.DrawImage(render.ImageCmd{
		Texture: c.texture,
		SrcRect: uimath.NewRect(0, 0, float32(c.width), float32(c.height)),
		DstRect: bounds,
		Tint:    uimath.ColorWhite,
	}, 0, 1)
}

func (c *Canvas) Destroy() {
	if c.texture != render.InvalidTexture && c.backend != nil {
		c.backend.DestroyTexture(c.texture)
		c.texture = render.InvalidTexture
	}
	c.Base.Destroy()
}

// ToImage converts the pixel buffer to a Go image.RGBA.
func (c *Canvas) ToImage() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, c.width, c.height))
	copy(img.Pix, c.pixels)
	return img
}

// ---------------------------------------------------------------------------
// LineCap / LineJoin enums (W3C names)
// ---------------------------------------------------------------------------

type LineCap uint8

const (
	LineCapButt   LineCap = iota // "butt"
	LineCapRound                 // "round"
	LineCapSquare                // "square"
)

type LineJoin uint8

const (
	LineJoinMiter LineJoin = iota // "miter"
	LineJoinRound                // "round"
	LineJoinBevel                // "bevel"
)

// TextAlign for canvas text drawing.
type CanvasTextAlign uint8

const (
	CanvasTextAlignStart  CanvasTextAlign = iota // "start" (default, same as left for LTR)
	CanvasTextAlignEnd                           // "end"
	CanvasTextAlignLeft                          // "left"
	CanvasTextAlignRight                         // "right"
	CanvasTextAlignCenter                        // "center"
)

// TextBaseline for canvas text drawing.
type CanvasTextBaseline uint8

const (
	CanvasTextBaselineAlphabetic  CanvasTextBaseline = iota // default
	CanvasTextBaselineTop
	CanvasTextBaselineHanging
	CanvasTextBaselineMiddle
	CanvasTextBaselineIdeographic
	CanvasTextBaselineBottom
)

// CompositeOp is globalCompositeOperation.
type CompositeOp uint8

const (
	CompositeSourceOver CompositeOp = iota // default
	CompositeSourceIn
	CompositeSourceOut
	CompositeSourceAtop
	CompositeDestinationOver
	CompositeDestinationIn
	CompositeDestinationOut
	CompositeDestinationAtop
	CompositeLighter
	CompositeCopy
	CompositeXOR
)

// ---------------------------------------------------------------------------
// CanvasGradient
// ---------------------------------------------------------------------------

type GradientStop struct {
	Offset float64
	Color  uimath.Color
}

type CanvasGradient struct {
	linear bool
	// Linear: x0,y0 → x1,y1
	x0, y0, x1, y1 float64
	// Radial: (x0,y0,r0) → (x1,y1,r1)
	r0, r1 float64
	stops  []GradientStop
}

func (g *CanvasGradient) AddColorStop(offset float64, c uimath.Color) {
	g.stops = append(g.stops, GradientStop{Offset: offset, Color: c})
}

// colorAt evaluates the gradient at parameter t ∈ [0,1].
func (g *CanvasGradient) colorAt(t float64) uimath.Color {
	if len(g.stops) == 0 {
		return uimath.Color{}
	}
	if t <= g.stops[0].Offset || len(g.stops) == 1 {
		return g.stops[0].Color
	}
	last := g.stops[len(g.stops)-1]
	if t >= last.Offset {
		return last.Color
	}
	for i := 1; i < len(g.stops); i++ {
		if t <= g.stops[i].Offset {
			s0, s1 := g.stops[i-1], g.stops[i]
			f := (t - s0.Offset) / (s1.Offset - s0.Offset)
			return lerpColor(s0.Color, s1.Color, f)
		}
	}
	return last.Color
}

func lerpColor(a, b uimath.Color, t float64) uimath.Color {
	return uimath.Color{
		R: float32(float64(a.R) + t*(float64(b.R)-float64(a.R))),
		G: float32(float64(a.G) + t*(float64(b.G)-float64(a.G))),
		B: float32(float64(a.B) + t*(float64(b.B)-float64(a.B))),
		A: float32(float64(a.A) + t*(float64(b.A)-float64(a.A))),
	}
}

// ---------------------------------------------------------------------------
// ImageData (W3C)
// ---------------------------------------------------------------------------

type ImageData struct {
	Width  int
	Height int
	Data   []byte // RGBA
}

// NewImageData creates a blank ImageData.
func NewImageData(width, height int) *ImageData {
	return &ImageData{
		Width:  width,
		Height: height,
		Data:   make([]byte, width*height*4),
	}
}

// TextMetrics returned by MeasureText.
type TextMetrics struct {
	Width float64
}

// ---------------------------------------------------------------------------
// CanvasStyle — can be solid color or gradient
// ---------------------------------------------------------------------------

type canvasStyle struct {
	color    uimath.Color
	gradient *CanvasGradient
}

func solidStyle(c uimath.Color) canvasStyle {
	return canvasStyle{color: c}
}

func gradientStyle(g *CanvasGradient) canvasStyle {
	return canvasStyle{gradient: g}
}

// ---------------------------------------------------------------------------
// drawState — saved/restored via Save/Restore
// ---------------------------------------------------------------------------

type drawState struct {
	fillStyle    canvasStyle
	strokeStyle  canvasStyle
	lineWidth    float64
	lineCap      LineCap
	lineJoin     LineJoin
	miterLimit   float64
	globalAlpha  float64
	compositeOp  CompositeOp
	fontSize     float64
	textAlign    CanvasTextAlign
	textBaseline CanvasTextBaseline
	transform    affineMatrix
	clipX, clipY, clipW, clipH int // clip rect (-1 = no clip)
}

func defaultDrawState() drawState {
	return drawState{
		fillStyle:   solidStyle(uimath.Color{R: 0, G: 0, B: 0, A: 1}),
		strokeStyle: solidStyle(uimath.Color{R: 0, G: 0, B: 0, A: 1}),
		lineWidth:   1,
		lineCap:     LineCapButt,
		lineJoin:    LineJoinMiter,
		miterLimit:  10,
		globalAlpha: 1,
		compositeOp: CompositeSourceOver,
		fontSize:    10,
		textAlign:   CanvasTextAlignStart,
		textBaseline: CanvasTextBaselineAlphabetic,
		transform:   identityMatrix(),
		clipX:       -1,
	}
}

// ---------------------------------------------------------------------------
// affineMatrix — 2D affine transform [a c e; b d f; 0 0 1]
// ---------------------------------------------------------------------------

type affineMatrix struct {
	A, B, C, D, E, F float64
}

func identityMatrix() affineMatrix {
	return affineMatrix{A: 1, D: 1}
}

func (m affineMatrix) transformPoint(x, y float64) (float64, float64) {
	return m.A*x + m.C*y + m.E, m.B*x + m.D*y + m.F
}

func (m affineMatrix) multiply(n affineMatrix) affineMatrix {
	return affineMatrix{
		A: m.A*n.A + m.C*n.B,
		B: m.B*n.A + m.D*n.B,
		C: m.A*n.C + m.C*n.D,
		D: m.B*n.C + m.D*n.D,
		E: m.A*n.E + m.C*n.F + m.E,
		F: m.B*n.E + m.D*n.F + m.F,
	}
}

// ---------------------------------------------------------------------------
// Path segment types
// ---------------------------------------------------------------------------

type pathSegType uint8

const (
	pathMoveTo pathSegType = iota
	pathLineTo
	pathClose
)

type pathSeg struct {
	typ  pathSegType
	x, y float64
}

// ---------------------------------------------------------------------------
// CanvasContext2D — the main 2D drawing API
// ---------------------------------------------------------------------------

type CanvasContext2D struct {
	canvas    *Canvas
	state     drawState
	stateStack []drawState
	path      []pathSeg
	pathX, pathY float64 // current point
}

func newCanvasContext2D(c *Canvas) *CanvasContext2D {
	return &CanvasContext2D{
		canvas: c,
		state:  defaultDrawState(),
	}
}

// --- State ---

func (ctx *CanvasContext2D) Save() {
	ctx.stateStack = append(ctx.stateStack, ctx.state)
}

func (ctx *CanvasContext2D) Restore() {
	n := len(ctx.stateStack)
	if n == 0 {
		return
	}
	ctx.state = ctx.stateStack[n-1]
	ctx.stateStack = ctx.stateStack[:n-1]
}

// --- Transform ---

func (ctx *CanvasContext2D) Scale(x, y float64) {
	ctx.state.transform = ctx.state.transform.multiply(affineMatrix{A: x, D: y})
}

func (ctx *CanvasContext2D) Rotate(angle float64) {
	cos, sin := math.Cos(angle), math.Sin(angle)
	ctx.state.transform = ctx.state.transform.multiply(affineMatrix{A: cos, B: sin, C: -sin, D: cos})
}

func (ctx *CanvasContext2D) Translate(x, y float64) {
	ctx.state.transform = ctx.state.transform.multiply(affineMatrix{A: 1, D: 1, E: x, F: y})
}

func (ctx *CanvasContext2D) Transform(a, b, c, d, e, f float64) {
	ctx.state.transform = ctx.state.transform.multiply(affineMatrix{A: a, B: b, C: c, D: d, E: e, F: f})
}

func (ctx *CanvasContext2D) SetTransform(a, b, c, d, e, f float64) {
	ctx.state.transform = affineMatrix{A: a, B: b, C: c, D: d, E: e, F: f}
}

func (ctx *CanvasContext2D) ResetTransform() {
	ctx.state.transform = identityMatrix()
}

func (ctx *CanvasContext2D) GetTransform() (a, b, c, d, e, f float64) {
	m := ctx.state.transform
	return m.A, m.B, m.C, m.D, m.E, m.F
}

// --- Style properties ---

func (ctx *CanvasContext2D) SetFillColor(c uimath.Color)   { ctx.state.fillStyle = solidStyle(c) }
func (ctx *CanvasContext2D) SetStrokeColor(c uimath.Color) { ctx.state.strokeStyle = solidStyle(c) }

// SetFillStyleGradient sets fill to a gradient.
func (ctx *CanvasContext2D) SetFillStyleGradient(g *CanvasGradient) {
	ctx.state.fillStyle = gradientStyle(g)
}
func (ctx *CanvasContext2D) SetStrokeStyleGradient(g *CanvasGradient) {
	ctx.state.strokeStyle = gradientStyle(g)
}

// SetFillStyleRGBA is a convenience for setting fill color from RGBA values.
func (ctx *CanvasContext2D) SetFillStyleRGBA(r, g, b, a float32) {
	ctx.state.fillStyle = solidStyle(uimath.Color{R: r, G: g, B: b, A: a})
}
func (ctx *CanvasContext2D) SetStrokeStyleRGBA(r, g, b, a float32) {
	ctx.state.strokeStyle = solidStyle(uimath.Color{R: r, G: g, B: b, A: a})
}

func (ctx *CanvasContext2D) FillColor() uimath.Color   { return ctx.state.fillStyle.color }
func (ctx *CanvasContext2D) StrokeColor() uimath.Color { return ctx.state.strokeStyle.color }

func (ctx *CanvasContext2D) SetLineWidth(w float64)  { ctx.state.lineWidth = w }
func (ctx *CanvasContext2D) LineWidth() float64       { return ctx.state.lineWidth }
func (ctx *CanvasContext2D) SetLineCap(c LineCap)     { ctx.state.lineCap = c }
func (ctx *CanvasContext2D) LineCap() LineCap         { return ctx.state.lineCap }
func (ctx *CanvasContext2D) SetLineJoin(j LineJoin)   { ctx.state.lineJoin = j }
func (ctx *CanvasContext2D) LineJoin() LineJoin        { return ctx.state.lineJoin }
func (ctx *CanvasContext2D) SetMiterLimit(m float64)  { ctx.state.miterLimit = m }
func (ctx *CanvasContext2D) MiterLimit() float64      { return ctx.state.miterLimit }

func (ctx *CanvasContext2D) SetGlobalAlpha(a float64) { ctx.state.globalAlpha = a }
func (ctx *CanvasContext2D) GlobalAlpha() float64     { return ctx.state.globalAlpha }

func (ctx *CanvasContext2D) SetGlobalCompositeOperation(op CompositeOp) { ctx.state.compositeOp = op }
func (ctx *CanvasContext2D) GlobalCompositeOperation() CompositeOp      { return ctx.state.compositeOp }

func (ctx *CanvasContext2D) SetFont(size float64)              { ctx.state.fontSize = size }
func (ctx *CanvasContext2D) Font() float64                     { return ctx.state.fontSize }
func (ctx *CanvasContext2D) SetTextAlign(a CanvasTextAlign)    { ctx.state.textAlign = a }
func (ctx *CanvasContext2D) TextAlign() CanvasTextAlign        { return ctx.state.textAlign }
func (ctx *CanvasContext2D) SetTextBaseline(b CanvasTextBaseline) { ctx.state.textBaseline = b }
func (ctx *CanvasContext2D) TextBaseline() CanvasTextBaseline  { return ctx.state.textBaseline }

// --- Rectangles ---

func (ctx *CanvasContext2D) ClearRect(x, y, w, h float64) {
	ctx.rasterClearRect(x, y, w, h)
}

func (ctx *CanvasContext2D) FillRect(x, y, w, h float64) {
	ctx.rasterFillRect(x, y, w, h)
}

func (ctx *CanvasContext2D) StrokeRect(x, y, w, h float64) {
	ctx.rasterStrokeRect(x, y, w, h)
}

// --- Path ---

func (ctx *CanvasContext2D) BeginPath() {
	ctx.path = ctx.path[:0]
}

func (ctx *CanvasContext2D) MoveTo(x, y float64) {
	tx, ty := ctx.state.transform.transformPoint(x, y)
	ctx.path = append(ctx.path, pathSeg{typ: pathMoveTo, x: tx, y: ty})
	ctx.pathX, ctx.pathY = tx, ty
}

func (ctx *CanvasContext2D) LineTo(x, y float64) {
	tx, ty := ctx.state.transform.transformPoint(x, y)
	ctx.path = append(ctx.path, pathSeg{typ: pathLineTo, x: tx, y: ty})
	ctx.pathX, ctx.pathY = tx, ty
}

func (ctx *CanvasContext2D) ClosePath() {
	ctx.path = append(ctx.path, pathSeg{typ: pathClose})
	// Find last MoveTo to update current point
	for i := len(ctx.path) - 2; i >= 0; i-- {
		if ctx.path[i].typ == pathMoveTo {
			ctx.pathX, ctx.pathY = ctx.path[i].x, ctx.path[i].y
			break
		}
	}
}

func (ctx *CanvasContext2D) Rect(x, y, w, h float64) {
	ctx.MoveTo(x, y)
	ctx.LineTo(x+w, y)
	ctx.LineTo(x+w, y+h)
	ctx.LineTo(x, y+h)
	ctx.ClosePath()
}

// Arc adds an arc to the path (angles in radians).
func (ctx *CanvasContext2D) Arc(cx, cy, radius, startAngle, endAngle float64, counterclockwise bool) {
	ctx.arcToSegments(cx, cy, radius, startAngle, endAngle, counterclockwise)
}

// ArcTo adds an arc defined by tangent lines (W3C arcTo).
func (ctx *CanvasContext2D) ArcTo(x1, y1, x2, y2, radius float64) {
	ctx.arcToTangent(x1, y1, x2, y2, radius)
}

// QuadraticCurveTo adds a quadratic Bézier curve.
func (ctx *CanvasContext2D) QuadraticCurveTo(cpx, cpy, x, y float64) {
	ctx.quadraticToSegments(cpx, cpy, x, y)
}

// BezierCurveTo adds a cubic Bézier curve.
func (ctx *CanvasContext2D) BezierCurveTo(cp1x, cp1y, cp2x, cp2y, x, y float64) {
	ctx.cubicToSegments(cp1x, cp1y, cp2x, cp2y, x, y)
}

func (ctx *CanvasContext2D) Fill()   { ctx.rasterFill() }
func (ctx *CanvasContext2D) Stroke() { ctx.rasterStroke() }

// IsPointInPath tests if (x,y) is inside the current path (even-odd rule).
func (ctx *CanvasContext2D) IsPointInPath(x, y float64) bool {
	tx, ty := ctx.state.transform.transformPoint(x, y)
	return ctx.pointInPath(tx, ty)
}

// --- Text ---

// FillText draws filled text at (x, y).
func (ctx *CanvasContext2D) FillText(text string, x, y float64) {
	ctx.rasterFillText(text, x, y)
}

// MeasureText returns text metrics for the given string.
func (ctx *CanvasContext2D) MeasureText(text string) TextMetrics {
	cfg := ctx.canvas.config
	if cfg.TextRenderer != nil {
		w := cfg.TextRenderer.MeasureText(text, float32(ctx.state.fontSize))
		return TextMetrics{Width: float64(w)}
	}
	// Rough estimate: fontSize * 0.6 per character
	return TextMetrics{Width: float64(len([]rune(text))) * ctx.state.fontSize * 0.6}
}

// --- Image ---

// DrawImage draws a source ImageData onto the canvas at (dx, dy).
func (ctx *CanvasContext2D) DrawImage(src *ImageData, dx, dy float64) {
	ctx.rasterDrawImage(src, dx, dy)
}

// DrawImageScaled draws a source ImageData scaled to (dw, dh).
func (ctx *CanvasContext2D) DrawImageScaled(src *ImageData, dx, dy, dw, dh float64) {
	ctx.rasterDrawImageScaled(src, dx, dy, dw, dh)
}

// --- Pixel manipulation ---

func (ctx *CanvasContext2D) CreateImageData(width, height int) *ImageData {
	return NewImageData(width, height)
}

func (ctx *CanvasContext2D) GetImageData(sx, sy, sw, sh int) *ImageData {
	c := ctx.canvas
	data := NewImageData(sw, sh)
	for dy := 0; dy < sh; dy++ {
		for dx := 0; dx < sw; dx++ {
			srcX, srcY := sx+dx, sy+dy
			if srcX >= 0 && srcX < c.width && srcY >= 0 && srcY < c.height {
				si := (srcY*c.width + srcX) * 4
				di := (dy*sw + dx) * 4
				copy(data.Data[di:di+4], c.pixels[si:si+4])
			}
		}
	}
	return data
}

func (ctx *CanvasContext2D) PutImageData(data *ImageData, dx, dy int) {
	c := ctx.canvas
	for sy := 0; sy < data.Height; sy++ {
		for sx := 0; sx < data.Width; sx++ {
			px, py := dx+sx, dy+sy
			if px >= 0 && px < c.width && py >= 0 && py < c.height {
				si := (sy*data.Width + sx) * 4
				di := (py*c.width + px) * 4
				copy(c.pixels[di:di+4], data.Data[si:si+4])
			}
		}
	}
	c.dirty = true
}

// --- Gradients ---

func (ctx *CanvasContext2D) CreateLinearGradient(x0, y0, x1, y1 float64) *CanvasGradient {
	return &CanvasGradient{linear: true, x0: x0, y0: y0, x1: x1, y1: y1}
}

func (ctx *CanvasContext2D) CreateRadialGradient(x0, y0, r0, x1, y1, r1 float64) *CanvasGradient {
	return &CanvasGradient{linear: false, x0: x0, y0: y0, r0: r0, x1: x1, y1: y1, r1: r1}
}

// --- Clipping ---

// Clip sets the clipping region to the current path's bounding box.
// (Simplified: uses bounding rect, not exact path shape.)
func (ctx *CanvasContext2D) Clip() {
	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64
	for _, seg := range ctx.path {
		if seg.typ == pathClose {
			continue
		}
		if seg.x < minX {
			minX = seg.x
		}
		if seg.y < minY {
			minY = seg.y
		}
		if seg.x > maxX {
			maxX = seg.x
		}
		if seg.y > maxY {
			maxY = seg.y
		}
	}
	if minX > maxX {
		return
	}
	ctx.state.clipX = int(minX)
	ctx.state.clipY = int(minY)
	ctx.state.clipW = int(maxX-minX) + 1
	ctx.state.clipH = int(maxY-minY) + 1
}

// --- Convenience: Go image.Image interop ---

// PutGoImage draws a Go image.Image onto the canvas at (dx, dy).
func (ctx *CanvasContext2D) PutGoImage(img image.Image, dx, dy int) {
	bounds := img.Bounds()
	c := ctx.canvas
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			px, py := dx+x-bounds.Min.X, dy+y-bounds.Min.Y
			if px >= 0 && px < c.width && py >= 0 && py < c.height {
				r, g, b, a := img.At(x, y).RGBA()
				di := (py*c.width + px) * 4
				c.pixels[di] = uint8(r >> 8)
				c.pixels[di+1] = uint8(g >> 8)
				c.pixels[di+2] = uint8(b >> 8)
				c.pixels[di+3] = uint8(a >> 8)
			}
		}
	}
	c.dirty = true
}

// ToGoColor converts a uimath.Color to color.RGBA.
func ToGoColor(c uimath.Color) color.RGBA {
	return color.RGBA{
		R: uint8(c.R * 255),
		G: uint8(c.G * 255),
		B: uint8(c.B * 255),
		A: uint8(c.A * 255),
	}
}

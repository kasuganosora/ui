package widget

import (
	"math"

	uimath "github.com/kasuganosora/ui/math"
)

// canvas_raster.go — software rasterization for CanvasContext2D.
// All coordinates are in canvas pixel space (post-transform where applicable).

// ---------------------------------------------------------------------------
// Pixel operations
// ---------------------------------------------------------------------------

// setPixel blends a color into the pixel at (x, y) respecting globalAlpha and compositeOp.
func (ctx *CanvasContext2D) setPixel(x, y int, col uimath.Color) {
	c := ctx.canvas
	if x < 0 || y < 0 || x >= c.width || y >= c.height {
		return
	}
	// Clip check
	s := &ctx.state
	if s.clipX >= 0 {
		if x < s.clipX || x >= s.clipX+s.clipW || y < s.clipY || y >= s.clipY+s.clipH {
			return
		}
	}

	alpha := col.A * float32(s.globalAlpha)
	if alpha <= 0 {
		return
	}

	i := (y*c.width + x) * 4
	sr, sg, sb, sa := col.R, col.G, col.B, alpha

	switch s.compositeOp {
	case CompositeCopy:
		c.pixels[i] = uint8(sr * 255)
		c.pixels[i+1] = uint8(sg * 255)
		c.pixels[i+2] = uint8(sb * 255)
		c.pixels[i+3] = uint8(sa * 255)
	case CompositeSourceOver:
		// Standard alpha blending: out = src * srcA + dst * (1 - srcA)
		dr := float32(c.pixels[i]) / 255
		dg := float32(c.pixels[i+1]) / 255
		db := float32(c.pixels[i+2]) / 255
		da := float32(c.pixels[i+3]) / 255
		inv := 1 - sa
		outA := sa + da*inv
		if outA > 0 {
			outR := (sr*sa + dr*da*inv) / outA
			outG := (sg*sa + dg*da*inv) / outA
			outB := (sb*sa + db*da*inv) / outA
			c.pixels[i] = clampByte(outR)
			c.pixels[i+1] = clampByte(outG)
			c.pixels[i+2] = clampByte(outB)
			c.pixels[i+3] = clampByte(outA)
		}
	case CompositeLighter:
		dr := float32(c.pixels[i]) / 255
		dg := float32(c.pixels[i+1]) / 255
		db := float32(c.pixels[i+2]) / 255
		da := float32(c.pixels[i+3]) / 255
		c.pixels[i] = clampByte(dr + sr*sa)
		c.pixels[i+1] = clampByte(dg + sg*sa)
		c.pixels[i+2] = clampByte(db + sb*sa)
		c.pixels[i+3] = clampByte(da + sa)
	case CompositeXOR:
		dr := float32(c.pixels[i]) / 255
		dg := float32(c.pixels[i+1]) / 255
		db := float32(c.pixels[i+2]) / 255
		da := float32(c.pixels[i+3]) / 255
		outA := sa*(1-da) + da*(1-sa)
		if outA > 0 {
			c.pixels[i] = clampByte((sr*sa*(1-da) + dr*da*(1-sa)) / outA)
			c.pixels[i+1] = clampByte((sg*sa*(1-da) + dg*da*(1-sa)) / outA)
			c.pixels[i+2] = clampByte((sb*sa*(1-da) + db*da*(1-sa)) / outA)
			c.pixels[i+3] = clampByte(outA)
		}
	default:
		// Fallback to source-over for unimplemented ops
		dr := float32(c.pixels[i]) / 255
		dg := float32(c.pixels[i+1]) / 255
		db := float32(c.pixels[i+2]) / 255
		da := float32(c.pixels[i+3]) / 255
		inv := 1 - sa
		outA := sa + da*inv
		if outA > 0 {
			c.pixels[i] = clampByte((sr*sa + dr*da*inv) / outA)
			c.pixels[i+1] = clampByte((sg*sa + dg*da*inv) / outA)
			c.pixels[i+2] = clampByte((sb*sa + db*da*inv) / outA)
			c.pixels[i+3] = clampByte(outA)
		}
	}
}

func clampByte(v float32) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	return uint8(v * 255)
}

// resolveStyleAt returns the color for a style at canvas coordinates (x, y).
func (ctx *CanvasContext2D) resolveStyleAt(style canvasStyle, x, y int) uimath.Color {
	if style.gradient == nil {
		return style.color
	}
	g := style.gradient
	fx, fy := float64(x), float64(y)
	if g.linear {
		dx, dy := g.x1-g.x0, g.y1-g.y0
		lenSq := dx*dx + dy*dy
		if lenSq == 0 {
			return g.colorAt(0)
		}
		t := ((fx-g.x0)*dx + (fy-g.y0)*dy) / lenSq
		if t < 0 {
			t = 0
		} else if t > 1 {
			t = 1
		}
		return g.colorAt(t)
	}
	// Radial gradient
	dist := math.Sqrt((fx-g.x0)*(fx-g.x0) + (fy-g.y0)*(fy-g.y0))
	rRange := g.r1 - g.r0
	if rRange == 0 {
		return g.colorAt(0)
	}
	t := (dist - g.r0) / rRange
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	return g.colorAt(t)
}

// ---------------------------------------------------------------------------
// Rectangle rasterization
// ---------------------------------------------------------------------------

func (ctx *CanvasContext2D) rasterClearRect(x, y, w, h float64) {
	c := ctx.canvas
	m := ctx.state.transform
	// Transform all four corners
	x0, y0 := m.transformPoint(x, y)
	x1, y1 := m.transformPoint(x+w, y)
	x2, y2 := m.transformPoint(x+w, y+h)
	x3, y3 := m.transformPoint(x, y+h)

	minX := int(math.Floor(min4(x0, x1, x2, x3)))
	minY := int(math.Floor(min4(y0, y1, y2, y3)))
	maxX := int(math.Ceil(max4(x0, x1, x2, x3)))
	maxY := int(math.Ceil(max4(y0, y1, y2, y3)))

	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX > c.width {
		maxX = c.width
	}
	if maxY > c.height {
		maxY = c.height
	}

	for py := minY; py < maxY; py++ {
		for px := minX; px < maxX; px++ {
			i := (py*c.width + px) * 4
			c.pixels[i] = 0
			c.pixels[i+1] = 0
			c.pixels[i+2] = 0
			c.pixels[i+3] = 0
		}
	}
	c.dirty = true
}

func (ctx *CanvasContext2D) rasterFillRect(x, y, w, h float64) {
	c := ctx.canvas
	m := ctx.state.transform
	x0, y0 := m.transformPoint(x, y)
	x1, y1 := m.transformPoint(x+w, y)
	x2, y2 := m.transformPoint(x+w, y+h)
	x3, y3 := m.transformPoint(x, y+h)

	minX := int(math.Floor(min4(x0, x1, x2, x3)))
	minY := int(math.Floor(min4(y0, y1, y2, y3)))
	maxX := int(math.Ceil(max4(x0, x1, x2, x3)))
	maxY := int(math.Ceil(max4(y0, y1, y2, y3)))

	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX > c.width {
		maxX = c.width
	}
	if maxY > c.height {
		maxY = c.height
	}

	style := ctx.state.fillStyle
	for py := minY; py < maxY; py++ {
		for px := minX; px < maxX; px++ {
			col := ctx.resolveStyleAt(style, px, py)
			ctx.setPixel(px, py, col)
		}
	}
	c.dirty = true
}

func (ctx *CanvasContext2D) rasterStrokeRect(x, y, w, h float64) {
	lw := ctx.state.lineWidth
	// Top
	ctx.rasterFillRectRaw(x, y, w, lw, ctx.state.strokeStyle)
	// Bottom
	ctx.rasterFillRectRaw(x, y+h-lw, w, lw, ctx.state.strokeStyle)
	// Left
	ctx.rasterFillRectRaw(x, y, lw, h, ctx.state.strokeStyle)
	// Right
	ctx.rasterFillRectRaw(x+w-lw, y, lw, h, ctx.state.strokeStyle)
}

func (ctx *CanvasContext2D) rasterFillRectRaw(x, y, w, h float64, style canvasStyle) {
	c := ctx.canvas
	m := ctx.state.transform
	x0, y0 := m.transformPoint(x, y)
	x1, y1 := m.transformPoint(x+w, y)
	x2, y2 := m.transformPoint(x+w, y+h)
	x3, y3 := m.transformPoint(x, y+h)

	minX := int(math.Floor(min4(x0, x1, x2, x3)))
	minY := int(math.Floor(min4(y0, y1, y2, y3)))
	maxX := int(math.Ceil(max4(x0, x1, x2, x3)))
	maxY := int(math.Ceil(max4(y0, y1, y2, y3)))

	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX > c.width {
		maxX = c.width
	}
	if maxY > c.height {
		maxY = c.height
	}

	for py := minY; py < maxY; py++ {
		for px := minX; px < maxX; px++ {
			col := ctx.resolveStyleAt(style, px, py)
			ctx.setPixel(px, py, col)
		}
	}
	c.dirty = true
}

// ---------------------------------------------------------------------------
// Line drawing (Bresenham with width)
// ---------------------------------------------------------------------------

func (ctx *CanvasContext2D) drawLine(x0, y0, x1, y1 float64, style canvasStyle) {
	lw := ctx.state.lineWidth
	if lw <= 1 {
		ctx.bresenham(int(math.Round(x0)), int(math.Round(y0)),
			int(math.Round(x1)), int(math.Round(y1)), style)
	} else {
		ctx.thickLine(x0, y0, x1, y1, lw, style)
	}
	ctx.canvas.dirty = true
}

func (ctx *CanvasContext2D) bresenham(x0, y0, x1, y1 int, style canvasStyle) {
	dx := abs(x1 - x0)
	dy := -abs(y1 - y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy
	for {
		col := ctx.resolveStyleAt(style, x0, y0)
		ctx.setPixel(x0, y0, col)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := err * 2
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func (ctx *CanvasContext2D) thickLine(x0, y0, x1, y1, width float64, style canvasStyle) {
	// Draw a thick line as a filled rectangle rotated along the line direction
	dx, dy := x1-x0, y1-y0
	length := math.Sqrt(dx*dx + dy*dy)
	if length == 0 {
		return
	}
	hw := width / 2
	// Perpendicular unit vector
	px, py := -dy/length*hw, dx/length*hw

	// Four corners of the thick line
	corners := [4][2]float64{
		{x0 + px, y0 + py},
		{x0 - px, y0 - py},
		{x1 - px, y1 - py},
		{x1 + px, y1 + py},
	}

	// Scanline fill the quad
	minY, maxY := corners[0][1], corners[0][1]
	for _, c := range corners[1:] {
		if c[1] < minY {
			minY = c[1]
		}
		if c[1] > maxY {
			maxY = c[1]
		}
	}

	iy0 := int(math.Floor(minY))
	iy1 := int(math.Ceil(maxY))
	cvs := ctx.canvas
	if iy0 < 0 {
		iy0 = 0
	}
	if iy1 > cvs.height {
		iy1 = cvs.height
	}

	for iy := iy0; iy < iy1; iy++ {
		fy := float64(iy) + 0.5
		// Find x-intersections of scanline with polygon edges
		var xs []float64
		for i := 0; i < 4; i++ {
			j := (i + 1) % 4
			y1e, y2e := corners[i][1], corners[j][1]
			if (y1e <= fy && y2e > fy) || (y2e <= fy && y1e > fy) {
				t := (fy - y1e) / (y2e - y1e)
				xs = append(xs, corners[i][0]+t*(corners[j][0]-corners[i][0]))
			}
		}
		if len(xs) >= 2 {
			if xs[0] > xs[1] {
				xs[0], xs[1] = xs[1], xs[0]
			}
			ix0 := int(math.Floor(xs[0]))
			ix1 := int(math.Ceil(xs[1]))
			if ix0 < 0 {
				ix0 = 0
			}
			if ix1 > cvs.width {
				ix1 = cvs.width
			}
			for ix := ix0; ix < ix1; ix++ {
				col := ctx.resolveStyleAt(style, ix, iy)
				ctx.setPixel(ix, iy, col)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Path rasterization
// ---------------------------------------------------------------------------

// rasterFill fills the current path using the even-odd rule.
func (ctx *CanvasContext2D) rasterFill() {
	if len(ctx.path) == 0 {
		return
	}

	// Collect edges from path segments
	type edge struct {
		x0, y0, x1, y1 float64
	}
	var edges []edge
	var startX, startY, curX, curY float64

	for _, seg := range ctx.path {
		switch seg.typ {
		case pathMoveTo:
			startX, startY = seg.x, seg.y
			curX, curY = seg.x, seg.y
		case pathLineTo:
			edges = append(edges, edge{curX, curY, seg.x, seg.y})
			curX, curY = seg.x, seg.y
		case pathClose:
			if curX != startX || curY != startY {
				edges = append(edges, edge{curX, curY, startX, startY})
			}
			curX, curY = startX, startY
		}
	}

	if len(edges) == 0 {
		return
	}

	// Find bounding box
	minY, maxY := edges[0].y0, edges[0].y0
	minX, maxX := edges[0].x0, edges[0].x0
	for _, e := range edges {
		for _, v := range []float64{e.y0, e.y1} {
			if v < minY {
				minY = v
			}
			if v > maxY {
				maxY = v
			}
		}
		for _, v := range []float64{e.x0, e.x1} {
			if v < minX {
				minX = v
			}
			if v > maxX {
				maxX = v
			}
		}
	}

	cvs := ctx.canvas
	iy0 := int(math.Floor(minY))
	iy1 := int(math.Ceil(maxY))
	if iy0 < 0 {
		iy0 = 0
	}
	if iy1 > cvs.height {
		iy1 = cvs.height
	}

	style := ctx.state.fillStyle

	// Scanline fill
	for iy := iy0; iy < iy1; iy++ {
		fy := float64(iy) + 0.5
		var xs []float64

		for _, e := range edges {
			y0, y1 := e.y0, e.y1
			if (y0 <= fy && y1 > fy) || (y1 <= fy && y0 > fy) {
				t := (fy - y0) / (y1 - y0)
				xs = append(xs, e.x0+t*(e.x1-e.x0))
			}
		}

		// Sort intersections
		for i := 1; i < len(xs); i++ {
			for j := i; j > 0 && xs[j-1] > xs[j]; j-- {
				xs[j-1], xs[j] = xs[j], xs[j-1]
			}
		}

		// Fill between pairs (even-odd rule)
		for i := 0; i+1 < len(xs); i += 2 {
			ix0 := int(math.Floor(xs[i]))
			ix1 := int(math.Ceil(xs[i+1]))
			if ix0 < 0 {
				ix0 = 0
			}
			if ix1 > cvs.width {
				ix1 = cvs.width
			}
			for ix := ix0; ix < ix1; ix++ {
				col := ctx.resolveStyleAt(style, ix, iy)
				ctx.setPixel(ix, iy, col)
			}
		}
	}
	cvs.dirty = true
}

// rasterStroke strokes the current path.
func (ctx *CanvasContext2D) rasterStroke() {
	if len(ctx.path) == 0 {
		return
	}

	style := ctx.state.strokeStyle
	var startX, startY, curX, curY float64

	for _, seg := range ctx.path {
		switch seg.typ {
		case pathMoveTo:
			startX, startY = seg.x, seg.y
			curX, curY = seg.x, seg.y
		case pathLineTo:
			ctx.drawLine(curX, curY, seg.x, seg.y, style)
			curX, curY = seg.x, seg.y
		case pathClose:
			if curX != startX || curY != startY {
				ctx.drawLine(curX, curY, startX, startY, style)
			}
			curX, curY = startX, startY
		}
	}
	ctx.canvas.dirty = true
}

// pointInPath tests if (x,y) is inside the path using even-odd rule.
func (ctx *CanvasContext2D) pointInPath(x, y float64) bool {
	crossings := 0
	var startX, startY, curX, curY float64

	testEdge := func(x0, y0, x1, y1 float64) {
		if (y0 <= y && y1 > y) || (y1 <= y && y0 > y) {
			t := (y - y0) / (y1 - y0)
			ix := x0 + t*(x1-x0)
			if x < ix {
				crossings++
			}
		}
	}

	for _, seg := range ctx.path {
		switch seg.typ {
		case pathMoveTo:
			startX, startY = seg.x, seg.y
			curX, curY = seg.x, seg.y
		case pathLineTo:
			testEdge(curX, curY, seg.x, seg.y)
			curX, curY = seg.x, seg.y
		case pathClose:
			testEdge(curX, curY, startX, startY)
			curX, curY = startX, startY
		}
	}
	return crossings%2 == 1
}

// ---------------------------------------------------------------------------
// Arc / curve → line segment conversion
// ---------------------------------------------------------------------------

func (ctx *CanvasContext2D) arcToSegments(cx, cy, r, startAngle, endAngle float64, ccw bool) {
	m := ctx.state.transform

	// Determine sweep
	sweep := endAngle - startAngle
	if ccw {
		if sweep > 0 {
			sweep -= 2 * math.Pi
		}
	} else {
		if sweep < 0 {
			sweep += 2 * math.Pi
		}
	}

	// Number of segments proportional to arc length
	n := int(math.Ceil(math.Abs(sweep) / (math.Pi / 16)))
	if n < 1 {
		n = 1
	}

	for i := 0; i <= n; i++ {
		t := float64(i) / float64(n)
		angle := startAngle + t*sweep
		px := cx + r*math.Cos(angle)
		py := cy + r*math.Sin(angle)
		tx, ty := m.transformPoint(px, py)

		if i == 0 && len(ctx.path) == 0 {
			ctx.path = append(ctx.path, pathSeg{typ: pathMoveTo, x: tx, y: ty})
		} else if i == 0 {
			ctx.path = append(ctx.path, pathSeg{typ: pathLineTo, x: tx, y: ty})
		} else {
			ctx.path = append(ctx.path, pathSeg{typ: pathLineTo, x: tx, y: ty})
		}
		ctx.pathX, ctx.pathY = tx, ty
	}
}

func (ctx *CanvasContext2D) arcToTangent(x1, y1, x2, y2, radius float64) {
	// Current point in user space (need to inverse-transform pathX, pathY)
	// Simplified: assume transform is identity for arcTo control points
	m := ctx.state.transform
	tx1, ty1 := m.transformPoint(x1, y1)
	tx2, ty2 := m.transformPoint(x2, y2)

	// Line from current point to (tx1, ty1) to (tx2, ty2)
	// Just draw line segments as approximation
	ctx.path = append(ctx.path, pathSeg{typ: pathLineTo, x: tx1, y: ty1})
	ctx.path = append(ctx.path, pathSeg{typ: pathLineTo, x: tx2, y: ty2})
	ctx.pathX, ctx.pathY = tx2, ty2
}

func (ctx *CanvasContext2D) quadraticToSegments(cpx, cpy, x, y float64) {
	m := ctx.state.transform
	// Start point
	sx, sy := ctx.pathX, ctx.pathY
	// Control and end in transformed space
	tcpx, tcpy := m.transformPoint(cpx, cpy)
	tex, tey := m.transformPoint(x, y)

	n := 16
	for i := 1; i <= n; i++ {
		t := float64(i) / float64(n)
		t1 := 1 - t
		px := t1*t1*sx + 2*t1*t*tcpx + t*t*tex
		py := t1*t1*sy + 2*t1*t*tcpy + t*t*tey
		ctx.path = append(ctx.path, pathSeg{typ: pathLineTo, x: px, y: py})
	}
	ctx.pathX, ctx.pathY = tex, tey
}

func (ctx *CanvasContext2D) cubicToSegments(cp1x, cp1y, cp2x, cp2y, x, y float64) {
	m := ctx.state.transform
	sx, sy := ctx.pathX, ctx.pathY
	tc1x, tc1y := m.transformPoint(cp1x, cp1y)
	tc2x, tc2y := m.transformPoint(cp2x, cp2y)
	tex, tey := m.transformPoint(x, y)

	n := 24
	for i := 1; i <= n; i++ {
		t := float64(i) / float64(n)
		t1 := 1 - t
		px := t1*t1*t1*sx + 3*t1*t1*t*tc1x + 3*t1*t*t*tc2x + t*t*t*tex
		py := t1*t1*t1*sy + 3*t1*t1*t*tc1y + 3*t1*t*t*tc2y + t*t*t*tey
		ctx.path = append(ctx.path, pathSeg{typ: pathLineTo, x: px, y: py})
	}
	ctx.pathX, ctx.pathY = tex, tey
}

// ---------------------------------------------------------------------------
// Text rasterization
// ---------------------------------------------------------------------------

func (ctx *CanvasContext2D) rasterFillText(text string, x, y float64) {
	// Delegate to config TextRenderer if available, otherwise simple bitmap fallback
	cfg := ctx.canvas.config
	if cfg.TextRenderer == nil {
		return
	}
	// For now, render text via the TextRenderer into the pixel buffer
	// This is a simplified approach — proper text rasterization would use glyph bitmaps
	// For the W3C Canvas API, we stamp glyphs into the pixel buffer
	m := ctx.state.transform
	tx, ty := m.transformPoint(x, y)

	// TextRenderer measures
	fontSize := float32(ctx.state.fontSize)
	tw := cfg.TextRenderer.MeasureText(text, fontSize)

	// Align
	switch ctx.state.textAlign {
	case CanvasTextAlignCenter:
		tx -= float64(tw) / 2
	case CanvasTextAlignRight, CanvasTextAlignEnd:
		tx -= float64(tw)
	}

	// Baseline adjustment
	lh := cfg.TextRenderer.LineHeight(fontSize)
	switch ctx.state.textBaseline {
	case CanvasTextBaselineTop, CanvasTextBaselineHanging:
		// ty is already at top
	case CanvasTextBaselineMiddle:
		ty -= float64(lh) / 2
	case CanvasTextBaselineBottom, CanvasTextBaselineIdeographic:
		ty -= float64(lh)
	default: // alphabetic
		ty -= float64(lh) * 0.8
	}

	// Simple fallback: fill rectangles for each character as colored blocks
	// Real implementation would rasterize glyph bitmaps
	col := ctx.state.fillStyle.color
	charW := tw / float32(len([]rune(text)))
	cx := float64(0)
	for range []rune(text) {
		// Draw a small filled rect per character
		rx := int(tx + cx)
		ry := int(ty)
		rw := int(charW)
		rh := int(lh)
		for py := ry; py < ry+rh; py++ {
			for px := rx; px < rx+rw; px++ {
				ctx.setPixel(px, py, col)
			}
		}
		cx += float64(charW)
	}
	ctx.canvas.dirty = true
}

// ---------------------------------------------------------------------------
// Image rasterization
// ---------------------------------------------------------------------------

func (ctx *CanvasContext2D) rasterDrawImage(src *ImageData, dx, dy float64) {
	ctx.rasterDrawImageScaled(src, dx, dy, float64(src.Width), float64(src.Height))
}

func (ctx *CanvasContext2D) rasterDrawImageScaled(src *ImageData, dx, dy, dw, dh float64) {
	c := ctx.canvas
	m := ctx.state.transform

	scaleX := float64(src.Width) / dw
	scaleY := float64(src.Height) / dh

	// Transform destination rect corners
	tx0, ty0 := m.transformPoint(dx, dy)

	idw := int(dw)
	idh := int(dh)

	for py := 0; py < idh; py++ {
		for px := 0; px < idw; px++ {
			// Source pixel
			sx := int(float64(px) * scaleX)
			sy := int(float64(py) * scaleY)
			if sx >= src.Width || sy >= src.Height {
				continue
			}
			si := (sy*src.Width + sx) * 4
			col := uimath.Color{
				R: float32(src.Data[si]) / 255,
				G: float32(src.Data[si+1]) / 255,
				B: float32(src.Data[si+2]) / 255,
				A: float32(src.Data[si+3]) / 255,
			}
			ctx.setPixel(int(tx0)+px, int(ty0)+py, col)
		}
	}
	c.dirty = true
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func min4(a, b, c, d float64) float64 {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	if d < m {
		m = d
	}
	return m
}

func max4(a, b, c, d float64) float64 {
	m := a
	if b > m {
		m = b
	}
	if c > m {
		m = c
	}
	if d > m {
		m = d
	}
	return m
}


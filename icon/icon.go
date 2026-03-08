// Package icon provides an icon registry and rendering system.
// Icons are stored as SVG path data and rasterized on demand to GPU textures.
package icon

import (
	"image"
	"image/color"
	"math"
	"sync"

	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Icon holds the SVG path data for a single icon.
type Icon struct {
	Name string
	Path string // SVG path d-attribute (24x24 viewBox)
}

// Registry manages a set of icons and caches their rasterized textures.
type Registry struct {
	mu      sync.RWMutex
	icons   map[string]*Icon
	cache   map[cacheKey]*CachedIcon
	backend render.Backend
}

type cacheKey struct {
	name string
	size int
}

// CachedIcon holds a rasterized icon texture ready for rendering.
type CachedIcon struct {
	Texture render.TextureHandle
	Size    int
}

// NewRegistry creates an icon registry bound to a render backend.
func NewRegistry(backend render.Backend) *Registry {
	return &Registry{
		icons:   make(map[string]*Icon),
		cache:   make(map[cacheKey]*CachedIcon),
		backend: backend,
	}
}

// Register adds an icon to the registry.
func (r *Registry) Register(name, svgPath string) {
	r.mu.Lock()
	r.icons[name] = &Icon{Name: name, Path: svgPath}
	r.mu.Unlock()
}

// RegisterAll adds multiple icons from a name→path map.
func (r *Registry) RegisterAll(icons map[string]string) {
	r.mu.Lock()
	for name, path := range icons {
		r.icons[name] = &Icon{Name: name, Path: path}
	}
	r.mu.Unlock()
}

// Has returns true if the named icon exists.
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	_, ok := r.icons[name]
	r.mu.RUnlock()
	return ok
}

// Names returns all registered icon names sorted.
func (r *Registry) Names() []string {
	r.mu.RLock()
	names := make([]string, 0, len(r.icons))
	for name := range r.icons {
		names = append(names, name)
	}
	r.mu.RUnlock()
	return names
}

// Count returns the number of registered icons.
func (r *Registry) Count() int {
	r.mu.RLock()
	n := len(r.icons)
	r.mu.RUnlock()
	return n
}

// Get returns the rasterized icon texture at the given pixel size.
// Returns (InvalidTexture, false) if the icon is not found or rasterization fails.
func (r *Registry) Get(name string, size int) (render.TextureHandle, bool) {
	key := cacheKey{name, size}

	r.mu.RLock()
	if cached, ok := r.cache[key]; ok {
		r.mu.RUnlock()
		return cached.Texture, true
	}
	icon, ok := r.icons[name]
	r.mu.RUnlock()
	if !ok {
		return render.InvalidTexture, false
	}

	// Rasterize the icon
	pixels := rasterizeIcon(icon.Path, size)
	if pixels == nil {
		return render.InvalidTexture, false
	}

	// Upload to GPU
	tex, err := r.backend.CreateTexture(render.TextureDesc{
		Width:  size,
		Height: size,
		Format: render.TextureFormatRGBA8,
		Data:   pixels,
	})
	if err != nil {
		return render.InvalidTexture, false
	}

	r.mu.Lock()
	r.cache[key] = &CachedIcon{Texture: tex, Size: size}
	r.mu.Unlock()

	return tex, true
}

// Destroy releases all cached textures.
func (r *Registry) Destroy() {
	r.mu.Lock()
	for _, cached := range r.cache {
		r.backend.DestroyTexture(cached.Texture)
	}
	r.cache = make(map[cacheKey]*CachedIcon)
	r.mu.Unlock()
}

// rasterizeIcon renders an SVG path (24x24 viewBox) to an RGBA pixel buffer.
// The icon is rendered as white (RGB=255) with alpha from coverage,
// so it can be tinted to any color via the Tint field when drawing.
func rasterizeIcon(svgPath string, size int) []byte {
	path := render.ParseSVGPath(svgPath)
	if path == nil || len(path.Commands) == 0 {
		return nil
	}

	// Create RGBA image
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Scale from 24x24 viewBox to target size
	scale := float64(size) / 24.0

	// Build edge list from path commands (flatten curves inline)
	edges := buildEdges(path.Commands, scale)
	if len(edges) == 0 {
		return nil
	}

	// Scanline fill with even-odd rule + 4x AA
	scanlineFillAA(img, edges)

	return img.Pix
}

type edge struct {
	x0, y0, x1, y1 float64
}

// buildEdges converts path commands to scaled edge segments for scanline fill.
func buildEdges(commands []render.PathCommand, scale float64) []edge {
	var edges []edge
	var cx, cy float64
	var mx, my float64

	addEdge := func(ax, ay, bx, by float64) {
		if ay != by {
			edges = append(edges, edge{ax, ay, bx, by})
		}
	}

	for _, cmd := range commands {
		switch cmd.Type {
		case render.PathMoveTo:
			cx, cy = float64(cmd.X1)*scale, float64(cmd.Y1)*scale
			mx, my = cx, cy

		case render.PathLineTo:
			nx, ny := float64(cmd.X1)*scale, float64(cmd.Y1)*scale
			addEdge(cx, cy, nx, ny)
			cx, cy = nx, ny

		case render.PathQuadTo:
			// Flatten quadratic bezier
			cpx, cpy := float64(cmd.X1)*scale, float64(cmd.Y1)*scale
			ex, ey := float64(cmd.X2)*scale, float64(cmd.Y2)*scale
			flattenQuadEdges(&edges, cx, cy, cpx, cpy, ex, ey, 0.25)
			cx, cy = ex, ey

		case render.PathCubicTo:
			// Flatten cubic bezier
			cp1x, cp1y := float64(cmd.X1)*scale, float64(cmd.Y1)*scale
			cp2x, cp2y := float64(cmd.X2)*scale, float64(cmd.Y2)*scale
			ex, ey := float64(cmd.X3)*scale, float64(cmd.Y3)*scale
			flattenCubicEdges(&edges, cx, cy, cp1x, cp1y, cp2x, cp2y, ex, ey, 0.25)
			cx, cy = ex, ey

		case render.PathArcTo:
			// Use render.Path's Flatten to handle arcs correctly.
			// Build a temporary path with just MoveTo + ArcTo to flatten it.
			ex, ey := float64(cmd.X3)*scale, float64(cmd.Y3)*scale
			tmpPath := render.NewPath()
			tmpPath.MoveTo(float32(cx/scale), float32(cy/scale))
			// Re-encode the arc command
			tmpPath.Commands = append(tmpPath.Commands, cmd)
			pts := tmpPath.Flatten(float32(0.25 / scale))
			// pts[0] is the moveTo, rest are line segments from flattened arc
			prev := [2]float64{cx, cy}
			for i := 1; i < len(pts); i++ {
				nx, ny := float64(pts[i].X)*scale, float64(pts[i].Y)*scale
				addEdge(prev[0], prev[1], nx, ny)
				prev = [2]float64{nx, ny}
			}
			cx, cy = ex, ey

		case render.PathClose:
			addEdge(cx, cy, mx, my)
			cx, cy = mx, my
		}
	}

	return edges
}

func flattenQuadEdges(edges *[]edge, x0, y0, cpx, cpy, x1, y1, tol float64) {
	// Check if flat enough
	dx := x1 - x0
	dy := y1 - y0
	d := math.Abs((cpx-x1)*dy - (cpy-y1)*dx)
	if d < tol*tol*0.25 {
		if y0 != y1 {
			*edges = append(*edges, edge{x0, y0, x1, y1})
		}
		return
	}
	// Subdivide
	mx0 := (x0 + cpx) * 0.5
	my0 := (y0 + cpy) * 0.5
	mx1 := (cpx + x1) * 0.5
	my1 := (cpy + y1) * 0.5
	mx := (mx0 + mx1) * 0.5
	my := (my0 + my1) * 0.5
	flattenQuadEdges(edges, x0, y0, mx0, my0, mx, my, tol)
	flattenQuadEdges(edges, mx, my, mx1, my1, x1, y1, tol)
}

func flattenCubicEdges(edges *[]edge, x0, y0, cp1x, cp1y, cp2x, cp2y, x1, y1, tol float64) {
	// Check if flat enough using distance of control points from line
	dx := x1 - x0
	dy := y1 - y0
	d1 := math.Abs((cp1x-x1)*dy - (cp1y-y1)*dx)
	d2 := math.Abs((cp2x-x1)*dy - (cp2y-y1)*dx)
	if (d1+d2)*(d1+d2) < tol*tol*(dx*dx+dy*dy)*4 {
		if y0 != y1 {
			*edges = append(*edges, edge{x0, y0, x1, y1})
		}
		return
	}
	// De Casteljau subdivision
	m01x := (x0 + cp1x) * 0.5
	m01y := (y0 + cp1y) * 0.5
	m12x := (cp1x + cp2x) * 0.5
	m12y := (cp1y + cp2y) * 0.5
	m23x := (cp2x + x1) * 0.5
	m23y := (cp2y + y1) * 0.5
	m012x := (m01x + m12x) * 0.5
	m012y := (m01y + m12y) * 0.5
	m123x := (m12x + m23x) * 0.5
	m123y := (m12y + m23y) * 0.5
	mx := (m012x + m123x) * 0.5
	my := (m012y + m123y) * 0.5
	flattenCubicEdges(edges, x0, y0, m01x, m01y, m012x, m012y, mx, my, tol)
	flattenCubicEdges(edges, mx, my, m123x, m123y, m23x, m23y, x1, y1, tol)
}

// scanlineFillAA fills the shape using 4x vertical supersampling for antialiasing.
func scanlineFillAA(img *image.RGBA, edges []edge) {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	const samples = 4

	for y := 0; y < h; y++ {
		// 4 sub-pixel scanlines per pixel row
		var coverage [4096]uint8 // per-pixel coverage count (0..samples)
		if w > 4096 {
			continue
		}

		for s := 0; s < samples; s++ {
			scanY := float64(y) + (float64(s)+0.5)/float64(samples)

			var intersections []float64
			for _, e := range edges {
				ey0, ey1 := e.y0, e.y1
				if ey0 > ey1 {
					ey0, ey1 = ey1, ey0
				}
				if scanY < ey0 || scanY >= ey1 {
					continue
				}
				t := (scanY - e.y0) / (e.y1 - e.y0)
				ix := e.x0 + t*(e.x1-e.x0)
				intersections = append(intersections, ix)
			}

			sortFloat64s(intersections)

			for i := 0; i+1 < len(intersections); i += 2 {
				xStart := int(math.Floor(intersections[i]))
				xEnd := int(math.Ceil(intersections[i+1]))
				if xStart < 0 {
					xStart = 0
				}
				if xEnd > w {
					xEnd = w
				}
				for x := xStart; x < xEnd; x++ {
					coverage[x]++
				}
			}
		}

		// Write pixels with alpha from coverage
		for x := 0; x < w; x++ {
			if coverage[x] > 0 {
				alpha := uint8(int(coverage[x]) * 255 / samples)
				img.SetRGBA(x, y, color.RGBA{255, 255, 255, alpha})
			}
		}
	}
}

func sortFloat64s(a []float64) {
	for i := 1; i < len(a); i++ {
		key := a[i]
		j := i - 1
		for j >= 0 && a[j] > key {
			a[j+1] = a[j]
			j--
		}
		a[j+1] = key
	}
}

// DrawIcon draws a named icon at the given position and size with color tinting.
func DrawIcon(buf *render.CommandBuffer, reg *Registry, name string, x, y, size float32, clr uimath.Color, zOrder int32, opacity float32) bool {
	pixelSize := int(size)
	if pixelSize < 1 {
		pixelSize = 1
	}

	tex, ok := reg.Get(name, pixelSize)
	if !ok {
		return false
	}

	buf.DrawImage(render.ImageCmd{
		Texture: tex,
		SrcRect: uimath.NewRect(0, 0, 1, 1),
		DstRect: uimath.NewRect(x, y, size, size),
		Tint:    clr,
	}, zOrder, opacity)
	return true
}

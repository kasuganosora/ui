// Package ttf implements a pure Go font.Engine backed by SFNT outlines.
//
// It is the fallback when the platform's native font engine (FreeType) is not
// available: instead of drawing a solid placeholder block for every character,
// it rasterizes real glyphs from a system TrueType/OpenType font.
package ttf

import (
	"fmt"
	"image"
	"image/draw"
	"os"
	"sync"

	"github.com/kasuganosora/ui/font"
	xfont "golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
	"golang.org/x/image/vector"
)

// fontFileCache keeps one copy of a font file's bytes. A font collection such
// as the system CJK font can be tens of megabytes; sfnt keeps a reference to
// its source slice, so sharing it avoids duplicating that allocation per face.
var fontFileCache sync.Map // path -> []byte

type face struct {
	f   *sfnt.Font
	buf *sfnt.Buffer
}

// Engine rasterizes glyph outlines with a pure Go vector rasterizer.
type Engine struct {
	mu          sync.Mutex
	faces       map[font.ID]*face
	nextID      font.ID
	dpiScale    float32
	pixelsPerPt float32
}

// New returns an empty engine. Register faces with LoadFont/LoadFontFile.
func New() *Engine {
	return &Engine{
		faces:       make(map[font.ID]*face),
		nextID:      1,
		dpiScale:    1,
		pixelsPerPt: 96.0 / 72.0,
	}
}

// NewFromFile returns an engine with the font at path already registered.
func NewFromFile(path string) (*Engine, error) {
	e := New()
	if _, err := e.LoadFontFile(path); err != nil {
		return nil, err
	}
	return e, nil
}

// faceCoverage samples the ranges the UI actually renders (ASCII, CJK, and a
// few UI symbols) and counts how many of them the face's cmap can resolve.
// Collections frequently ship regional variants whose cmap subtables differ:
// face 0 of the macOS CJK collection is the HK variant and misses simplified
// characters such as 稳, while the SC faces cover them all.
func faceCoverage(f *sfnt.Font, buf *sfnt.Buffer) int {
	hits := 0
	check := func(r rune) {
		if g, err := f.GlyphIndex(buf, r); err == nil && g != 0 {
			hits++
		}
	}
	for r := rune(' '); r <= '~'; r++ {
		check(r)
	}
	for r := rune(0x4E00); r <= 0x9FA5; r += 17 {
		check(r)
	}
	for _, r := range []rune{'✓', '×', '…', '●', '★', '—', '、', '。', '，', '！', '？', '￥'} {
		check(r)
	}
	return hits
}

// chooseCollectionFace returns the face with the widest character coverage.
// Ties keep the earliest face, which for typical collections is the Regular
// weight; this avoids silently switching every label to a light/bold face.
func chooseCollectionFace(col *sfnt.Collection) (*sfnt.Font, error) {
	var best *sfnt.Font
	bestScore := -1
	for i := 0; i < col.NumFonts(); i++ {
		f, err := col.Font(i)
		if err != nil {
			continue
		}
		score := faceCoverage(f, &sfnt.Buffer{})
		if score > bestScore {
			best, bestScore = f, score
		}
	}
	if best == nil {
		return nil, fmt.Errorf("ttf: no usable face in collection")
	}
	return best, nil
}

// LoadFont implements font.Engine. Both single fonts and font collections are
// accepted; for a collection the face with the widest cmap coverage is used.
func (e *Engine) LoadFont(data []byte) (font.ID, error) {
	var f *sfnt.Font
	var err error
	if len(data) >= 4 && string(data[:4]) == "ttcf" {
		var col *sfnt.Collection
		col, err = sfnt.ParseCollection(data)
		if err == nil {
			f, err = chooseCollectionFace(col)
		}
	} else {
		f, err = sfnt.Parse(data)
	}
	if err != nil {
		return font.InvalidFontID, fmt.Errorf("ttf: parse font: %w", err)
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	id := e.nextID
	e.nextID++
	e.faces[id] = &face{f: f, buf: &sfnt.Buffer{}}
	return id, nil
}

// LoadFontFile implements font.Engine.
func (e *Engine) LoadFontFile(path string) (font.ID, error) {
	if cached, ok := fontFileCache.Load(path); ok {
		return e.LoadFont(cached.([]byte))
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return font.InvalidFontID, fmt.Errorf("ttf: read font %q: %w", path, err)
	}
	fontFileCache.Store(path, data)
	return e.LoadFont(data)
}

// UnloadFont implements font.Engine.
func (e *Engine) UnloadFont(id font.ID) {
	e.mu.Lock()
	delete(e.faces, id)
	e.mu.Unlock()
}

// Destroy implements font.Engine.
func (e *Engine) Destroy() {
	e.mu.Lock()
	e.faces = make(map[font.ID]*face)
	e.mu.Unlock()
}

// SetDPIScale implements font.Engine.
func (e *Engine) SetDPIScale(scale float32) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if scale <= 0 {
		scale = 1
	}
	e.dpiScale = scale
	e.pixelsPerPt = scale * 96.0 / 72.0
}

// ppem converts a point size to pixels-per-em in 26.6 fixed point.
func (e *Engine) ppem(size float32) fixed.Int26_6 {
	pixels := int(size*e.pixelsPerPt + 0.5)
	if pixels < 1 {
		pixels = 1
	}
	return fixed.I(pixels)
}

// FontMetrics implements font.Engine. Results are logical pixels, matching
// the FreeType engine (physical pixels divided by dpiScale).
func (e *Engine) FontMetrics(id font.ID, size float32) font.Metrics {
	e.mu.Lock()
	defer e.mu.Unlock()
	fc := e.faces[id]
	if fc == nil {
		return font.Metrics{}
	}
	m, err := fc.f.Metrics(fc.buf, e.ppem(size), xfont.HintingNone)
	if err != nil {
		return font.Metrics{}
	}
	s := e.dpiScale
	positive := func(v fixed.Int26_6) float32 {
		f := float32(v) / 64 / s
		if f < 0 {
			return -f
		}
		return f
	}
	return font.Metrics{
		Ascent:     positive(m.Ascent),
		Descent:    positive(m.Descent),
		LineHeight: positive(m.Height),
		UnitsPerEm: float32(fc.f.UnitsPerEm()),
	}
}

func (e *Engine) glyphIndexLocked(fc *face, r rune) font.GlyphID {
	gid, err := fc.f.GlyphIndex(fc.buf, r)
	if err != nil {
		return 0
	}
	return font.GlyphID(gid)
}

// GlyphIndex implements font.Engine.
func (e *Engine) GlyphIndex(id font.ID, r rune) font.GlyphID {
	e.mu.Lock()
	defer e.mu.Unlock()
	fc := e.faces[id]
	if fc == nil {
		return 0
	}
	return e.glyphIndexLocked(fc, r)
}

// HasGlyph implements font.Engine.
func (e *Engine) HasGlyph(id font.ID, r rune) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	fc := e.faces[id]
	return fc != nil && e.glyphIndexLocked(fc, r) != 0
}

// HasColorGlyphs implements font.Engine. Outline rasterization does not read
// color bitmap tables (sbix/CBDT/COLR), so color emoji still needs FreeType.
func (e *Engine) HasColorGlyphs(font.ID) bool { return false }

// glyphBox returns the integer pixel box and advance of a glyph at size.
func (e *Engine) glyphBox(fc *face, glyph font.GlyphID, size float32) (x0, y0, x1, y1 int, advance fixed.Int26_6, err error) {
	b, adv, err := fc.f.GlyphBounds(fc.buf, sfnt.GlyphIndex(glyph), e.ppem(size), xfont.HintingNone)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	x0, y0 = b.Min.X.Floor(), b.Min.Y.Floor()
	x1, y1 = b.Max.X.Ceil(), b.Max.Y.Ceil()
	return x0, y0, x1, y1, adv, nil
}

// GlyphMetrics implements font.Engine.
func (e *Engine) GlyphMetrics(id font.ID, glyph font.GlyphID, size float32) font.GlyphMetrics {
	e.mu.Lock()
	defer e.mu.Unlock()
	fc := e.faces[id]
	if fc == nil || glyph == 0 {
		return font.GlyphMetrics{}
	}
	x0, y0, x1, y1, adv, err := e.glyphBox(fc, glyph, size)
	if err != nil {
		return font.GlyphMetrics{}
	}
	s := e.dpiScale
	return font.GlyphMetrics{
		Width:    float32(x1-x0) / s,
		Height:   float32(y1-y0) / s,
		BearingX: float32(x0) / s,
		BearingY: -float32(y0) / s,
		Advance:  float32(adv) / 64 / s,
	}
}

// RasterizeGlyph implements font.Engine. It produces an 8-bit coverage mask.
// SDF is not supported: the vector rasterizer yields antialiased coverage,
// not a signed distance field.
func (e *Engine) RasterizeGlyph(id font.ID, glyph font.GlyphID, size float32, sdf bool) (font.GlyphBitmap, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	fc := e.faces[id]
	if fc == nil {
		return font.GlyphBitmap{}, fmt.Errorf("ttf: font %d not found", id)
	}
	if glyph == 0 {
		return font.GlyphBitmap{}, nil
	}
	segs, err := fc.f.LoadGlyph(fc.buf, sfnt.GlyphIndex(glyph), e.ppem(size), nil)
	if err != nil {
		return font.GlyphBitmap{}, fmt.Errorf("ttf: load glyph %d: %w", glyph, err)
	}
	b := segs.Bounds()
	x0, y0 := b.Min.X.Floor(), b.Min.Y.Floor()
	x1, y1 := b.Max.X.Ceil(), b.Max.Y.Ceil()
	w, h := x1-x0, y1-y0
	if w <= 0 || h <= 0 {
		return font.GlyphBitmap{}, nil
	}

	r := vector.NewRasterizer(w, h)
	r.DrawOp = draw.Src
	for _, seg := range segs {
		pt := func(i int) (float32, float32) {
			return float32(seg.Args[i].X)/64 - float32(x0),
				float32(seg.Args[i].Y)/64 - float32(y0)
		}
		switch seg.Op {
		case sfnt.SegmentOpMoveTo:
			x, y := pt(0)
			r.MoveTo(x, y)
		case sfnt.SegmentOpLineTo:
			x, y := pt(0)
			r.LineTo(x, y)
		case sfnt.SegmentOpQuadTo:
			bx, by := pt(0)
			cx, cy := pt(1)
			r.QuadTo(bx, by, cx, cy)
		case sfnt.SegmentOpCubeTo:
			bx, by := pt(0)
			cx, cy := pt(1)
			dx, dy := pt(2)
			r.CubeTo(bx, by, cx, cy, dx, dy)
		}
	}
	mask := image.NewAlpha(image.Rect(0, 0, w, h))
	r.Draw(mask, mask.Bounds(), image.Opaque, image.Point{})

	data := make([]byte, w*h)
	copy(data, mask.Pix)
	return font.GlyphBitmap{Width: w, Height: h, Data: data}, nil
}

// Kerning implements font.Engine.
func (e *Engine) Kerning(id font.ID, left, right font.GlyphID, size float32) float32 {
	e.mu.Lock()
	defer e.mu.Unlock()
	fc := e.faces[id]
	if fc == nil {
		return 0
	}
	k, err := fc.f.Kern(fc.buf, sfnt.GlyphIndex(left), sfnt.GlyphIndex(right), e.ppem(size), xfont.HintingNone)
	if err != nil {
		return 0
	}
	return float32(k) / 64 / e.dpiScale
}

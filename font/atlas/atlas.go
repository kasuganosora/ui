package atlas

import (
	"sync"

	"github.com/kasuganosora/ui/font"
	"github.com/kasuganosora/ui/render"
)

// GlyphKey uniquely identifies a glyph in the atlas.
type GlyphKey struct {
	FontID   font.ID
	GlyphID  font.GlyphID
	Size     uint16 // Font size in half-pixels (size*2) for sub-pixel precision
}

// GlyphEntry contains the atlas location and metrics of a cached glyph.
type GlyphEntry struct {
	Region  Region  // Location in atlas texture
	Metrics font.GlyphMetrics

	// UV coordinates (computed from region and atlas size)
	U0, V0, U1, V1 float32

	// LRU tracking
	lastUsedFrame uint64
}

// Atlas manages a texture atlas of rasterized glyphs.
// It handles bin-packing, LRU eviction, and GPU texture updates.
type Atlas struct {
	mu sync.Mutex

	width, height int
	packer        *shelfPacker

	// Glyph cache: key -> entry
	glyphs map[GlyphKey]*GlyphEntry

	// GPU texture handle (created lazily)
	texture render.TextureHandle
	backend render.Backend

	// Pixel data buffer (CPU side, for batching uploads)
	pixels []byte
	dirty  bool // True if pixels have been modified since last upload
	bpp    int  // Bytes per pixel: 1 for R8 (SDF/gray), 4 for RGBA8 (color emoji)

	// Dirty region tracking for partial uploads
	dirtyMinX, dirtyMinY int
	dirtyMaxX, dirtyMaxY int

	// LRU
	currentFrame uint64

	// Maximum texture dimension for growth
	maxSize int

	// SDF mode
	sdf   bool
	color bool // True if this is an RGBA color atlas
}

// Options for creating an atlas.
type Options struct {
	Width   int
	Height  int
	MaxSize int  // Maximum atlas dimension (0 = 4096)
	SDF     bool // Use SDF rendering
	Color   bool // True for RGBA color glyph atlas (e.g., color emoji)
	Backend render.Backend
}

// New creates a new glyph atlas.
func New(opts Options) *Atlas {
	if opts.Width == 0 {
		opts.Width = 1024
	}
	if opts.Height == 0 {
		opts.Height = 1024
	}
	maxSize := opts.MaxSize
	if maxSize <= 0 {
		maxSize = 4096
	}

	bpp := 1
	if opts.Color {
		bpp = 4
	}

	return &Atlas{
		width:     opts.Width,
		height:    opts.Height,
		packer:    newShelfPacker(opts.Width, opts.Height),
		glyphs:    make(map[GlyphKey]*GlyphEntry),
		pixels:    make([]byte, opts.Width*opts.Height*bpp),
		bpp:       bpp,
		backend:   opts.Backend,
		maxSize:   maxSize,
		sdf:       opts.SDF,
		color:     opts.Color,
		dirtyMinX: opts.Width,
		dirtyMinY: opts.Height,
	}
}

// MakeKey creates a GlyphKey from font parameters.
// Size is stored as half-pixels for sub-pixel precision.
func MakeKey(fontID font.ID, glyphID font.GlyphID, size float32) GlyphKey {
	return GlyphKey{
		FontID:  fontID,
		GlyphID: glyphID,
		Size:    uint16(size * 2),
	}
}

// Lookup returns a cached glyph entry, or nil if not cached.
// Updates LRU timestamp on hit.
func (a *Atlas) Lookup(key GlyphKey) *GlyphEntry {
	a.mu.Lock()
	defer a.mu.Unlock()

	entry, ok := a.glyphs[key]
	if !ok {
		return nil
	}
	entry.lastUsedFrame = a.currentFrame
	return entry
}

// Insert adds a rasterized glyph to the atlas.
// Returns the entry on success, or nil if the atlas is full (after attempting eviction).
func (a *Atlas) Insert(key GlyphKey, bitmap font.GlyphBitmap, metrics font.GlyphMetrics) *GlyphEntry {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Already cached?
	if entry, ok := a.glyphs[key]; ok {
		entry.lastUsedFrame = a.currentFrame
		return entry
	}

	region, ok := a.packer.Pack(bitmap.Width, bitmap.Height)
	if !ok {
		// Try evicting stale glyphs first
		a.evictStale()
		region, ok = a.packer.Pack(bitmap.Width, bitmap.Height)
	}
	if !ok {
		// Try growing the atlas
		if a.grow() {
			region, ok = a.packer.Pack(bitmap.Width, bitmap.Height)
		}
	}
	if !ok {
		return nil // Can't fit even after eviction and growth
	}

	// Copy bitmap data into atlas pixel buffer
	a.blitBitmap(region, bitmap)

	entry := &GlyphEntry{
		Region:        region,
		Metrics:       metrics,
		U0:            float32(region.X) / float32(a.width),
		V0:            float32(region.Y) / float32(a.height),
		U1:            float32(region.X+region.Width) / float32(a.width),
		V1:            float32(region.Y+region.Height) / float32(a.height),
		lastUsedFrame: a.currentFrame,
	}

	a.glyphs[key] = entry
	return entry
}

// blitBitmap copies glyph bitmap data into the atlas pixel buffer.
func (a *Atlas) blitBitmap(region Region, bitmap font.GlyphBitmap) {
	bpp := a.bpp
	srcBPP := 1
	if bitmap.Color {
		srcBPP = 4
	}
	for y := 0; y < bitmap.Height && y < region.Height; y++ {
		srcOffset := y * bitmap.Width * srcBPP
		if srcOffset >= len(bitmap.Data) {
			break
		}
		srcRowBytes := bitmap.Width * srcBPP
		srcEnd := srcOffset + srcRowBytes
		if srcEnd > len(bitmap.Data) {
			srcEnd = len(bitmap.Data)
		}
		dstOffset := ((region.Y+y)*a.width + region.X) * bpp

		if srcBPP == bpp {
			// Same format: direct copy
			copyLen := srcEnd - srcOffset
			dstEnd := dstOffset + copyLen
			if dstEnd > len(a.pixels) {
				dstEnd = len(a.pixels)
			}
			if dstOffset < dstEnd {
				copy(a.pixels[dstOffset:dstEnd], bitmap.Data[srcOffset:srcEnd])
			}
		} else if srcBPP == 1 && bpp == 4 {
			// R8 → RGBA: expand each byte to (255, 255, 255, alpha)
			for x := 0; x < bitmap.Width && srcOffset+x < srcEnd; x++ {
				di := dstOffset + x*4
				if di+3 >= len(a.pixels) {
					break
				}
				a.pixels[di+0] = 255
				a.pixels[di+1] = 255
				a.pixels[di+2] = 255
				a.pixels[di+3] = bitmap.Data[srcOffset+x]
			}
		} else if srcBPP == 4 && bpp == 1 {
			// RGBA → R8: extract alpha channel
			for x := 0; x < bitmap.Width && srcOffset+x*4+3 < srcEnd; x++ {
				di := dstOffset + x
				if di >= len(a.pixels) {
					break
				}
				a.pixels[di] = bitmap.Data[srcOffset+x*4+3]
			}
		}
	}

	// Track dirty region
	if region.X < a.dirtyMinX {
		a.dirtyMinX = region.X
	}
	if region.Y < a.dirtyMinY {
		a.dirtyMinY = region.Y
	}
	rx := region.X + region.Width
	if rx > a.dirtyMaxX {
		a.dirtyMaxX = rx
	}
	ry := region.Y + region.Height
	if ry > a.dirtyMaxY {
		a.dirtyMaxY = ry
	}
	a.dirty = true
}

// Upload uploads dirty regions to the GPU texture.
// Must be called once per frame before rendering text.
func (a *Atlas) Upload() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.dirty {
		return nil
	}

	format := render.TextureFormatR8
	if a.color {
		format = render.TextureFormatRGBA8
	}

	// Create texture on first upload
	if a.texture == render.InvalidTexture && a.backend != nil {
		var err error
		a.texture, err = a.backend.CreateTexture(render.TextureDesc{
			Width:  a.width,
			Height: a.height,
			Format: format,
			Filter: render.TextureFilterLinear,
			Data:   a.pixels,
		})
		if err != nil {
			return err
		}
		a.resetDirty()
		return nil
	}

	if a.backend == nil || a.texture == render.InvalidTexture {
		a.resetDirty()
		return nil
	}

	// Partial upload of dirty region
	dx := a.dirtyMinX
	dy := a.dirtyMinY
	dw := a.dirtyMaxX - a.dirtyMinX
	dh := a.dirtyMaxY - a.dirtyMinY
	if dw <= 0 || dh <= 0 {
		a.resetDirty()
		return nil
	}

	// Extract the dirty sub-rectangle
	bpp := a.bpp
	subData := make([]byte, dw*dh*bpp)
	for y := 0; y < dh; y++ {
		srcStart := ((dy+y)*a.width + dx) * bpp
		copy(subData[y*dw*bpp:(y+1)*dw*bpp], a.pixels[srcStart:srcStart+dw*bpp])
	}

	uiRect := uimathRect(float32(dx), float32(dy), float32(dw), float32(dh))
	err := a.backend.UpdateTexture(a.texture, uiRect, subData)

	a.resetDirty()
	return err
}

func (a *Atlas) resetDirty() {
	a.dirty = false
	a.dirtyMinX = a.width
	a.dirtyMinY = a.height
	a.dirtyMaxX = 0
	a.dirtyMaxY = 0
}

// EnsureTexture pre-creates the GPU texture if it doesn't exist yet.
// Call this after atlas creation to ensure Texture() returns a valid handle
// before the first Draw/Upload cycle. Without this, the first frame's text
// commands reference InvalidTexture since Upload() hasn't been called yet.
func (a *Atlas) EnsureTexture() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.texture != render.InvalidTexture || a.backend == nil {
		return nil
	}

	format := render.TextureFormatR8
	if a.color {
		format = render.TextureFormatRGBA8
	}
	var err error
	a.texture, err = a.backend.CreateTexture(render.TextureDesc{
		Width:  a.width,
		Height: a.height,
		Format: format,
		Filter: render.TextureFilterLinear,
		Data:   a.pixels,
	})
	return err
}

// Texture returns the GPU texture handle for this atlas.
func (a *Atlas) Texture() render.TextureHandle {
	return a.texture
}

// BeginFrame advances the LRU frame counter.
func (a *Atlas) BeginFrame() {
	a.mu.Lock()
	a.currentFrame++
	a.mu.Unlock()
}

// evictStale removes glyphs not used in the last 10 frames.
// Returns the number of evicted glyphs. If any were evicted, the packer
// and pixel buffer are rebuilt since shelf space can't be reclaimed individually.
func (a *Atlas) evictStale() int {
	threshold := a.currentFrame - 10
	if a.currentFrame < 10 {
		threshold = 0
	}
	evicted := 0
	for k, entry := range a.glyphs {
		if entry.lastUsedFrame < threshold {
			delete(a.glyphs, k)
			evicted++
		}
	}
	if evicted > 0 {
		a.rebuildPacker()
	}
	return evicted
}

// rebuildPacker resets the packer and pixel buffer.
// Since we don't retain raw bitmap data for remaining glyphs, they are also
// cleared and will be re-rasterized on the next lookup miss.
func (a *Atlas) rebuildPacker() {
	a.packer.Reset()
	for i := range a.pixels {
		a.pixels[i] = 0
	}
	// Without stored bitmaps we can't re-pack remaining glyphs,
	// so clear the glyph map as well.
	for k := range a.glyphs {
		delete(a.glyphs, k)
	}
	a.dirtyMinX = 0
	a.dirtyMinY = 0
	a.dirtyMaxX = a.width
	a.dirtyMaxY = a.height
	a.dirty = true
}

// grow doubles the atlas in one dimension (the smaller one first).
// Returns true if growth succeeded, false if already at max size.
func (a *Atlas) grow() bool {
	maxSize := a.maxSize
	if a.backend != nil {
		if ms := a.backend.MaxTextureSize(); ms > 0 && ms < maxSize {
			maxSize = ms
		}
	}

	newW, newH := a.width, a.height
	// Double the smaller dimension, or width if equal
	if newW <= newH && newW*2 <= maxSize {
		newW *= 2
	} else if newH*2 <= maxSize {
		newH *= 2
	} else {
		return false // Can't grow further
	}

	// Allocate new pixel buffer and copy old data row by row
	bpp := a.bpp
	newPixels := make([]byte, newW*newH*bpp)
	for y := 0; y < a.height; y++ {
		srcOff := y * a.width * bpp
		dstOff := y * newW * bpp
		copy(newPixels[dstOff:dstOff+a.width*bpp], a.pixels[srcOff:srcOff+a.width*bpp])
	}

	// Update packer dimensions
	a.packer.Grow(newW, newH)

	// Update UV coordinates for existing glyphs
	for _, entry := range a.glyphs {
		entry.U0 = float32(entry.Region.X) / float32(newW)
		entry.V0 = float32(entry.Region.Y) / float32(newH)
		entry.U1 = float32(entry.Region.X+entry.Region.Width) / float32(newW)
		entry.V1 = float32(entry.Region.Y+entry.Region.Height) / float32(newH)
	}

	// Destroy old GPU texture — a new one will be created on next Upload
	if a.backend != nil && a.texture != render.InvalidTexture {
		a.backend.DestroyTexture(a.texture)
		a.texture = render.InvalidTexture
	}

	a.pixels = newPixels
	a.width = newW
	a.height = newH

	// Mark entire atlas dirty for full re-upload
	a.dirtyMinX = 0
	a.dirtyMinY = 0
	a.dirtyMaxX = newW
	a.dirtyMaxY = newH
	a.dirty = true

	return true
}

// Destroy releases GPU resources.
func (a *Atlas) Destroy() {
	if a.backend != nil && a.texture != render.InvalidTexture {
		a.backend.DestroyTexture(a.texture)
		a.texture = render.InvalidTexture
	}
}

// IsColor returns true if this is an RGBA color atlas.
func (a *Atlas) IsColor() bool { return a.color }

// Width returns the atlas width in pixels.
func (a *Atlas) Width() int { return a.width }

// Height returns the atlas height in pixels.
func (a *Atlas) Height() int { return a.height }

// GlyphCount returns the number of cached glyphs.
func (a *Atlas) GlyphCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.glyphs)
}

// Occupancy returns the fraction of atlas area used.
func (a *Atlas) Occupancy() float32 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.packer.Occupancy()
}

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

	// Dirty region tracking for partial uploads
	dirtyMinX, dirtyMinY int
	dirtyMaxX, dirtyMaxY int

	// LRU
	currentFrame uint64

	// SDF mode
	sdf bool
}

// Options for creating an atlas.
type Options struct {
	Width   int
	Height  int
	SDF     bool // Use SDF rendering
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

	return &Atlas{
		width:     opts.Width,
		height:    opts.Height,
		packer:    newShelfPacker(opts.Width, opts.Height),
		glyphs:    make(map[GlyphKey]*GlyphEntry),
		pixels:    make([]byte, opts.Width*opts.Height),
		backend:   opts.Backend,
		sdf:       opts.SDF,
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
		// Atlas full — try eviction
		a.evictOldest()
		region, ok = a.packer.Pack(bitmap.Width, bitmap.Height)
		if !ok {
			return nil // Still can't fit
		}
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
	for y := 0; y < bitmap.Height && y < region.Height; y++ {
		srcOffset := y * bitmap.Width
		if srcOffset >= len(bitmap.Data) {
			break
		}
		srcEnd := srcOffset + bitmap.Width
		if srcEnd > len(bitmap.Data) {
			srcEnd = len(bitmap.Data)
		}
		copyLen := srcEnd - srcOffset
		dstOffset := (region.Y+y)*a.width + region.X
		dstEnd := dstOffset + copyLen
		if dstEnd > len(a.pixels) {
			dstEnd = len(a.pixels)
		}
		if dstOffset < dstEnd {
			copy(a.pixels[dstOffset:dstEnd], bitmap.Data[srcOffset:srcEnd])
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

	// Create texture on first upload
	if a.texture == render.InvalidTexture && a.backend != nil {
		var err error
		a.texture, err = a.backend.CreateTexture(render.TextureDesc{
			Width:  a.width,
			Height: a.height,
			Format: render.TextureFormatR8,
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
	subData := make([]byte, dw*dh)
	for y := 0; y < dh; y++ {
		srcStart := (dy+y)*a.width + dx
		copy(subData[y*dw:(y+1)*dw], a.pixels[srcStart:srcStart+dw])
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

// evictOldest clears the atlas entirely when full.
// A more sophisticated approach would evict only old entries,
// but full-reset is simpler and sufficient for most UI workloads.
func (a *Atlas) evictOldest() {
	// Clear all cached glyphs
	for k := range a.glyphs {
		delete(a.glyphs, k)
	}
	a.packer.Reset()

	// Clear pixel data
	for i := range a.pixels {
		a.pixels[i] = 0
	}

	// Mark entire atlas dirty for re-upload
	a.dirtyMinX = 0
	a.dirtyMinY = 0
	a.dirtyMaxX = a.width
	a.dirtyMaxY = a.height
	a.dirty = true
}

// Destroy releases GPU resources.
func (a *Atlas) Destroy() {
	if a.backend != nil && a.texture != render.InvalidTexture {
		a.backend.DestroyTexture(a.texture)
		a.texture = render.InvalidTexture
	}
}

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

// Package atlas provides a glyph texture atlas with bin-packing and LRU eviction.
package atlas

// Region represents a rectangular region in the atlas.
type Region struct {
	X, Y          int
	Width, Height int
}

// shelfPacker implements a shelf-based bin packing algorithm.
// Shelves are horizontal rows. Glyphs are placed left-to-right on the current shelf.
// When a glyph doesn't fit, a new shelf is started.
// This is simple, fast, and well-suited for similarly-sized glyphs.
type shelfPacker struct {
	width, height int

	// Current shelves
	shelves []shelf
}

type shelf struct {
	y      int // Top Y of this shelf
	height int // Height of this shelf
	cursorX int // Next available X position
}

func newShelfPacker(width, height int) *shelfPacker {
	return &shelfPacker{
		width:  width,
		height: height,
	}
}

// Pack tries to place a rectangle of the given size.
// Returns the region and true if successful, or zero Region and false if full.
func (p *shelfPacker) Pack(w, h int) (Region, bool) {
	if w > p.width || h > p.height {
		return Region{}, false
	}

	// Padding between glyphs to avoid bleeding
	const pad = 1
	pw := w + pad
	ph := h + pad

	// Try to fit on an existing shelf
	for i := range p.shelves {
		s := &p.shelves[i]
		if ph <= s.height && s.cursorX+pw <= p.width {
			r := Region{X: s.cursorX, Y: s.y, Width: w, Height: h}
			s.cursorX += pw
			return r, true
		}
	}

	// Start a new shelf
	nextY := 0
	if len(p.shelves) > 0 {
		last := p.shelves[len(p.shelves)-1]
		nextY = last.y + last.height + pad
	}

	if nextY+ph > p.height {
		return Region{}, false // Atlas is full
	}

	p.shelves = append(p.shelves, shelf{
		y:       nextY,
		height:  ph,
		cursorX: pw,
	})

	return Region{X: 0, Y: nextY, Width: w, Height: h}, true
}

// Reset clears all shelves, making the entire atlas available again.
func (p *shelfPacker) Reset() {
	p.shelves = p.shelves[:0]
}

// Grow expands the packer dimensions. Existing shelves remain valid
// since growth only adds space to the right and/or bottom.
func (p *shelfPacker) Grow(newWidth, newHeight int) {
	p.width = newWidth
	p.height = newHeight
}

// Occupancy returns the fraction of atlas area used (0.0 to 1.0).
func (p *shelfPacker) Occupancy() float32 {
	used := 0
	for _, s := range p.shelves {
		used += s.cursorX * s.height
	}
	total := p.width * p.height
	if total == 0 {
		return 0
	}
	return float32(used) / float32(total)
}

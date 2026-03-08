package render

import "sync"

// Object pools for render command structs.
// These reduce GC pressure by reusing allocations across frames.

var (
	rectCmdPool  = sync.Pool{New: func() any { return &RectCmd{} }}
	textCmdPool  = sync.Pool{New: func() any { return &TextCmd{} }}
	imageCmdPool = sync.Pool{New: func() any { return &ImageCmd{} }}
	clipCmdPool  = sync.Pool{New: func() any { return &ClipCmd{} }}
)

// AcquireRectCmd gets a RectCmd from the pool.
func AcquireRectCmd() *RectCmd { return rectCmdPool.Get().(*RectCmd) }

// ReleaseRectCmd zeroes and returns a RectCmd to the pool.
func ReleaseRectCmd(c *RectCmd) { *c = RectCmd{}; rectCmdPool.Put(c) }

// AcquireTextCmd gets a TextCmd from the pool.
func AcquireTextCmd() *TextCmd { return textCmdPool.Get().(*TextCmd) }

// ReleaseTextCmd zeroes and returns a TextCmd to the pool.
func ReleaseTextCmd(c *TextCmd) { *c = TextCmd{}; textCmdPool.Put(c) }

// AcquireImageCmd gets an ImageCmd from the pool.
func AcquireImageCmd() *ImageCmd { return imageCmdPool.Get().(*ImageCmd) }

// ReleaseImageCmd zeroes and returns an ImageCmd to the pool.
func ReleaseImageCmd(c *ImageCmd) { *c = ImageCmd{}; imageCmdPool.Put(c) }

// AcquireClipCmd gets a ClipCmd from the pool.
func AcquireClipCmd() *ClipCmd { return clipCmdPool.Get().(*ClipCmd) }

// ReleaseClipCmd zeroes and returns a ClipCmd to the pool.
func ReleaseClipCmd(c *ClipCmd) { *c = ClipCmd{}; clipCmdPool.Put(c) }

// GlyphSlice pool for reusing []GlyphInstance slices.
var glyphSlicePool = sync.Pool{New: func() any { s := make([]GlyphInstance, 0, 64); return &s }}

// AcquireGlyphSlice gets a []GlyphInstance slice from the pool, reset to length 0.
func AcquireGlyphSlice() []GlyphInstance {
	sp := glyphSlicePool.Get().(*[]GlyphInstance)
	return (*sp)[:0]
}

// ReleaseGlyphSlice returns a []GlyphInstance slice to the pool.
func ReleaseGlyphSlice(s []GlyphInstance) {
	s = s[:0]
	glyphSlicePool.Put(&s)
}

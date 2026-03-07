// Package textrender bridges the font system (shaper + atlas + engine) with the
// render system. It takes text, shapes it, rasterizes missing glyphs into an
// atlas, uploads dirty regions to the GPU, and emits render.TextCmd commands.
package textrender

import (
	"github.com/kasuganosora/ui/font"
	"github.com/kasuganosora/ui/font/atlas"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Renderer bridges font shaping/rasterization with the render command buffer.
type Renderer struct {
	engine Engine
	shaper font.Shaper
	atlas  *atlas.Atlas
	sdf    bool
}

// Engine is the subset of font.Engine needed by the renderer.
type Engine interface {
	GlyphMetrics(id font.ID, glyph font.GlyphID, size float32) font.GlyphMetrics
	RasterizeGlyph(id font.ID, glyph font.GlyphID, size float32, sdf bool) (font.GlyphBitmap, error)
}

// Options configures a Renderer.
type Options struct {
	Manager font.Manager
	Atlas   *atlas.Atlas
	SDF     bool
}

// New creates a new text renderer.
func New(opts Options) *Renderer {
	return &Renderer{
		engine: opts.Manager.Engine(),
		shaper: font.NewShaper(opts.Manager),
		atlas:  opts.Atlas,
		sdf:    opts.SDF,
	}
}

// DrawText shapes text and emits render commands into the command buffer.
func (r *Renderer) DrawText(buf *render.CommandBuffer, text string, opts DrawOptions) {
	if text == "" {
		return
	}
	runs := r.shaper.Shape(text, opts.ShapeOpts)
	r.DrawRuns(buf, runs, opts)
}

// DrawRuns takes pre-shaped glyph runs and emits render commands.
func (r *Renderer) DrawRuns(buf *render.CommandBuffer, runs []font.GlyphRun, opts DrawOptions) {
	for _, run := range runs {
		r.drawRun(buf, run, opts)
	}
}

// drawRun processes a single glyph run.
func (r *Renderer) drawRun(buf *render.CommandBuffer, run font.GlyphRun, opts DrawOptions) {
	if len(run.Glyphs) == 0 {
		return
	}

	instances := make([]render.GlyphInstance, 0, len(run.Glyphs))

	for _, pg := range run.Glyphs {
		entry := r.ensureGlyph(run.FontID, pg.GlyphID, run.FontSize)
		if entry == nil || entry.Region.Width == 0 || entry.Region.Height == 0 {
			continue
		}

		gm := entry.Metrics
		instances = append(instances, render.GlyphInstance{
			X:      opts.OriginX + pg.X + gm.BearingX,
			Y:      opts.OriginY + pg.Y - gm.BearingY,
			Width:  float32(entry.Region.Width),
			Height: float32(entry.Region.Height),
			U0:     entry.U0,
			V0:     entry.V0,
			U1:     entry.U1,
			V1:     entry.V1,
		})
	}

	if len(instances) == 0 {
		return
	}

	buf.DrawText(render.TextCmd{
		X:        opts.OriginX,
		Y:        opts.OriginY,
		Glyphs:   instances,
		Color:    opts.Color,
		FontSize: run.FontSize,
		Atlas:    r.atlas.Texture(),
	}, opts.ZOrder, opts.Opacity)
}

// ensureGlyph checks the atlas and rasterizes on cache miss.
func (r *Renderer) ensureGlyph(fontID font.ID, glyphID font.GlyphID, size float32) *atlas.GlyphEntry {
	key := atlas.MakeKey(fontID, glyphID, size)

	if entry := r.atlas.Lookup(key); entry != nil {
		return entry
	}

	bitmap, err := r.engine.RasterizeGlyph(fontID, glyphID, size, r.sdf)
	if err != nil {
		return nil
	}

	metrics := r.engine.GlyphMetrics(fontID, glyphID, size)
	return r.atlas.Insert(key, bitmap, metrics)
}

// Upload uploads dirty atlas regions to the GPU.
// Call once per frame after all DrawText calls and before rendering.
func (r *Renderer) Upload() error {
	return r.atlas.Upload()
}

// BeginFrame advances the atlas LRU counter.
func (r *Renderer) BeginFrame() {
	r.atlas.BeginFrame()
}

// Measure measures text without rendering.
func (r *Renderer) Measure(text string, opts font.ShapeOptions) font.TextMetrics {
	return r.shaper.Measure(text, opts)
}

// Shape shapes text into glyph runs without rendering.
func (r *Renderer) Shape(text string, opts font.ShapeOptions) []font.GlyphRun {
	return r.shaper.Shape(text, opts)
}

// Atlas returns the underlying glyph atlas.
func (r *Renderer) Atlas() *atlas.Atlas {
	return r.atlas
}

// Shaper returns the underlying text shaper.
func (r *Renderer) Shaper() font.Shaper {
	return r.shaper
}

// Destroy releases resources.
func (r *Renderer) Destroy() {
	r.atlas.Destroy()
}

// DrawOptions controls how text is rendered.
type DrawOptions struct {
	ShapeOpts font.ShapeOptions
	OriginX   float32     // Screen X origin
	OriginY   float32     // Screen Y origin
	Color     uimath.Color
	ZOrder    int32
	Opacity   float32     // 0..1
}

// Package textrender bridges the font system (shaper + atlas + engine) with the
// render system. It takes text, shapes it, rasterizes missing glyphs into an
// atlas, uploads dirty regions to the GPU, and emits render.TextCmd commands.
// Color emoji glyphs are stored in a separate RGBA atlas and rendered through
// the image/textured pipeline.
package textrender

import (
	"math"

	"github.com/kasuganosora/ui/font"
	"github.com/kasuganosora/ui/font/atlas"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Renderer bridges font shaping/rasterization with the render command buffer.
type Renderer struct {
	engine    Engine
	shaper    font.Shaper
	atlas     *atlas.Atlas // R8 atlas for SDF/grayscale text
	sdf       bool
	dpiScale  float32 // DPI scale for converting atlas bitmap (physical) to logical pixels
	keepAlive func() // Called during heavy rasterization to keep the window responsive

	// Color emoji support
	colorAtlas *atlas.Atlas     // RGBA atlas for color emoji glyphs (created lazily)
	colorFonts map[font.ID]bool // Cache: fontID -> HasColorGlyphs result
	colorBack  render.Backend   // Backend for creating color atlas texture

	// Batched keepAlive: only call every N glyph misses to reduce overhead
	missCount int
}

// Engine is the subset of font.Engine needed by the renderer.
type Engine interface {
	GlyphMetrics(id font.ID, glyph font.GlyphID, size float32) font.GlyphMetrics
	RasterizeGlyph(id font.ID, glyph font.GlyphID, size float32, sdf bool) (font.GlyphBitmap, error)
}

// ColorEngine extends Engine with color glyph detection.
type ColorEngine interface {
	Engine
	HasColorGlyphs(id font.ID) bool
}

// Options configures a Renderer.
type Options struct {
	Manager   font.Manager
	Atlas     *atlas.Atlas
	SDF       bool
	DPIScale  float32 // DPI scale factor (1.0 = 96 DPI). Defaults to 1.0 if zero.
	KeepAlive func() // Optional: called during glyph rasterization to pump the OS message queue
	Backend   render.Backend // For creating color atlas texture (optional)
}

// New creates a new text renderer.
func New(opts Options) *Renderer {
	dpi := opts.DPIScale
	if dpi <= 0 {
		dpi = 1.0
	}
	return &Renderer{
		engine:     opts.Manager.Engine(),
		shaper:     font.NewShaper(opts.Manager),
		atlas:      opts.Atlas,
		sdf:        opts.SDF,
		dpiScale:   dpi,
		keepAlive:  opts.KeepAlive,
		colorFonts: make(map[font.ID]bool),
		colorBack:  opts.Backend,
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

// isColorFont checks (with caching) whether a font has color glyphs.
func (r *Renderer) isColorFont(fontID font.ID) bool {
	if v, ok := r.colorFonts[fontID]; ok {
		return v
	}
	ce, ok := r.engine.(ColorEngine)
	if !ok {
		r.colorFonts[fontID] = false
		return false
	}
	v := ce.HasColorGlyphs(fontID)
	r.colorFonts[fontID] = v
	return v
}

// ensureColorAtlas creates the color atlas lazily on first use.
func (r *Renderer) ensureColorAtlas() *atlas.Atlas {
	if r.colorAtlas != nil {
		return r.colorAtlas
	}
	r.colorAtlas = atlas.New(atlas.Options{
		Width:   512,
		Height:  512,
		MaxSize: 4096,
		Color:   true,
		Backend: r.colorBack,
	})
	return r.colorAtlas
}

// drawRun processes a single glyph run.
func (r *Renderer) drawRun(buf *render.CommandBuffer, run font.GlyphRun, opts DrawOptions) {
	if len(run.Glyphs) == 0 {
		return
	}

	isColor := r.isColorFont(run.FontID)

	if isColor {
		r.drawColorRun(buf, run, opts)
		return
	}

	instances := make([]render.GlyphInstance, 0, len(run.Glyphs))

	for _, pg := range run.Glyphs {
		entry := r.ensureGlyph(run.FontID, pg.GlyphID, run.FontSize)
		if entry == nil || entry.Region.Width == 0 || entry.Region.Height == 0 {
			continue
		}

		gm := entry.Metrics
		// Snap glyph positions to pixel grid for crisp rendering.
		// All values are in logical pixels (metrics already divided by dpiScale).
		gx := float32(math.Floor(float64(opts.OriginX + pg.X + gm.BearingX)))
		gy := float32(math.Floor(float64(opts.OriginY + pg.Y - gm.BearingY)))
		// Atlas bitmap is in physical pixels; convert to logical.
		instances = append(instances, render.GlyphInstance{
			X:      gx,
			Y:      gy,
			Width:  float32(entry.Region.Width) / r.dpiScale,
			Height: float32(entry.Region.Height) / r.dpiScale,
			U0:     entry.U0,
			V0:     entry.V0,
			U1:     entry.U1,
			V1:     entry.V1,
		})
	}

	if len(instances) == 0 {
		return
	}

	cmd := render.TextCmd{
		X:        opts.OriginX,
		Y:        opts.OriginY,
		Glyphs:   instances,
		Color:    opts.Color,
		FontSize: run.FontSize,
		Atlas:    r.atlas.Texture(),
	}
	if opts.Overlay {
		buf.DrawOverlayTextCmd(cmd, opts.ZOrder, opts.Opacity)
	} else {
		buf.DrawText(cmd, opts.ZOrder, opts.Opacity)
	}
}

// drawColorRun renders color emoji glyphs as individual ImageCmd commands.
// Each glyph is a separate textured quad using the RGBA color atlas.
func (r *Renderer) drawColorRun(buf *render.CommandBuffer, run font.GlyphRun, opts DrawOptions) {
	ca := r.ensureColorAtlas()
	tex := ca.Texture()

	for _, pg := range run.Glyphs {
		entry := r.ensureColorGlyph(run.FontID, pg.GlyphID, run.FontSize)
		if entry == nil || entry.Region.Width == 0 || entry.Region.Height == 0 {
			continue
		}

		gm := entry.Metrics
		gx := float32(math.Floor(float64(opts.OriginX + pg.X + gm.BearingX)))
		gy := float32(math.Floor(float64(opts.OriginY + pg.Y - gm.BearingY)))
		w := float32(entry.Region.Width) / r.dpiScale
		h := float32(entry.Region.Height) / r.dpiScale

		// Re-fetch texture handle after potential atlas growth/upload
		tex = ca.Texture()

		cmd := render.ImageCmd{
			Texture: tex,
			SrcRect: uimath.NewRect(entry.U0, entry.V0, entry.U1-entry.U0, entry.V1-entry.V0),
			DstRect: uimath.NewRect(gx, gy, w, h),
			Tint:    uimath.Color{R: 1, G: 1, B: 1, A: opts.Color.A},
		}
		buf.DrawImage(cmd, opts.ZOrder, opts.Opacity)
	}
}

// ensureGlyph checks the SDF atlas and rasterizes on cache miss.
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
	entry := r.atlas.Insert(key, bitmap, metrics)

	// Pump OS messages periodically to prevent "Not Responding".
	// Batching reduces syscall overhead from per-glyph to every 16 glyphs.
	r.missCount++
	if r.keepAlive != nil && r.missCount&15 == 0 {
		r.keepAlive()
	}

	return entry
}

// ensureColorGlyph checks the color atlas and rasterizes on cache miss.
func (r *Renderer) ensureColorGlyph(fontID font.ID, glyphID font.GlyphID, size float32) *atlas.GlyphEntry {
	ca := r.ensureColorAtlas()
	key := atlas.MakeKey(fontID, glyphID, size)

	if entry := ca.Lookup(key); entry != nil {
		return entry
	}

	// Rasterize without SDF — the engine will use FT_LOAD_COLOR for color fonts
	bitmap, err := r.engine.RasterizeGlyph(fontID, glyphID, size, false)
	if err != nil {
		return nil
	}

	metrics := r.engine.GlyphMetrics(fontID, glyphID, size)
	entry := ca.Insert(key, bitmap, metrics)

	r.missCount++
	if r.keepAlive != nil && r.missCount&15 == 0 {
		r.keepAlive()
	}

	return entry
}

// Upload uploads dirty atlas regions to the GPU.
// Call once per frame after all DrawText calls and before rendering.
func (r *Renderer) Upload() error {
	if err := r.atlas.Upload(); err != nil {
		return err
	}
	if r.colorAtlas != nil {
		return r.colorAtlas.Upload()
	}
	return nil
}

// BeginFrame advances the atlas LRU counter.
func (r *Renderer) BeginFrame() {
	r.atlas.BeginFrame()
	if r.colorAtlas != nil {
		r.colorAtlas.BeginFrame()
	}
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

// ColorAtlas returns the color emoji atlas (may be nil if unused).
func (r *Renderer) ColorAtlas() *atlas.Atlas {
	return r.colorAtlas
}

// Shaper returns the underlying text shaper.
func (r *Renderer) Shaper() font.Shaper {
	return r.shaper
}

// Destroy releases resources.
func (r *Renderer) Destroy() {
	r.atlas.Destroy()
	if r.colorAtlas != nil {
		r.colorAtlas.Destroy()
	}
}

// DrawOptions controls how text is rendered.
type DrawOptions struct {
	ShapeOpts font.ShapeOptions
	OriginX   float32      // Screen X origin
	OriginY   float32      // Screen Y origin
	Color     uimath.Color
	ZOrder    int32
	Opacity   float32 // 0..1
	// Overlay routes the text command into the overlay layer (rendered above all
	// normal content, no clip applied). Used for DevTools labels and debug UI.
	Overlay bool
}

package font

import uimath "github.com/kasuganosora/ui/math"

// ID uniquely identifies a registered font.
type ID uint32

const InvalidFontID ID = 0

// Style represents font style (normal, italic, etc.).
type Style uint8

const (
	StyleNormal Style = iota
	StyleItalic
)

// Weight represents font weight.
type Weight uint16

const (
	WeightThin       Weight = 100
	WeightExtraLight Weight = 200
	WeightLight      Weight = 300
	WeightRegular    Weight = 400
	WeightMedium     Weight = 500
	WeightSemiBold   Weight = 600
	WeightBold       Weight = 700
	WeightExtraBold  Weight = 800
	WeightBlack      Weight = 900
)

// Properties describes a font for matching/selection. Value object.
type Properties struct {
	Family string
	Weight Weight
	Style  Style
}

// GlyphID is an index into a font's glyph table.
type GlyphID uint32

// Metrics contains font-level metrics for a given size. Value object.
type Metrics struct {
	Ascent      float32 // Distance from baseline to top (positive up)
	Descent     float32 // Distance from baseline to bottom (positive down)
	LineHeight  float32 // Recommended line spacing
	UnitsPerEm  float32 // Font design units per em
}

// GlyphMetrics contains metrics for a single glyph. Value object.
type GlyphMetrics struct {
	Width    float32 // Glyph bitmap width
	Height   float32 // Glyph bitmap height
	BearingX float32 // Offset from origin to left edge of bitmap
	BearingY float32 // Offset from baseline to top edge of bitmap
	Advance  float32 // Horizontal advance to next glyph
}

// GlyphBitmap contains a rasterized glyph. Value object.
type GlyphBitmap struct {
	Width  int
	Height int
	Data   []byte // Pixel data (R8 for SDF/gray, RGBA8 for color emoji)
	SDF    bool   // True if this is an SDF bitmap
	Color  bool   // True if this is an RGBA color bitmap (e.g., color emoji)
}

// TextMetrics describes measured text dimensions. Value object.
type TextMetrics struct {
	Width      float32       // Total advance width
	Height     float32       // Total height (ascent + descent)
	LineCount  int           // Number of lines
	Lines      []LineMetrics // Per-line metrics
}

// LineMetrics describes a single line of text. Value object.
type LineMetrics struct {
	Width   float32 // Width of this line
	Ascent  float32
	Descent float32
	Start   int     // Start byte offset in source text
	End     int     // End byte offset in source text
}

// GlyphRun is a sequence of positioned glyphs ready for rendering. Value object.
type GlyphRun struct {
	FontID   ID
	FontSize float32
	Glyphs   []PositionedGlyph
	Bounds   uimath.Rect
}

// PositionedGlyph is a glyph with its screen position. Value object.
type PositionedGlyph struct {
	GlyphID  GlyphID
	X        float32 // Screen X position
	Y        float32 // Screen Y position (baseline)
	Advance  float32
}

// Engine is the font system interface.
// Implementations: FreeType (CGO), pure Go fallback.
type Engine interface {
	// LoadFont loads a font from raw data and returns a face ID.
	LoadFont(data []byte) (ID, error)

	// LoadFontFile loads a font from a file path.
	LoadFontFile(path string) (ID, error)

	// UnloadFont releases a loaded font.
	UnloadFont(id ID)

	// FontMetrics returns font-level metrics at the given size.
	FontMetrics(id ID, size float32) Metrics

	// GlyphIndex maps a rune to a glyph ID. Returns 0 if not found.
	GlyphIndex(id ID, r rune) GlyphID

	// GlyphMetrics returns metrics for a glyph at the given size.
	GlyphMetrics(id ID, glyph GlyphID, size float32) GlyphMetrics

	// RasterizeGlyph rasterizes a glyph to a bitmap.
	RasterizeGlyph(id ID, glyph GlyphID, size float32, sdf bool) (GlyphBitmap, error)

	// Kerning returns the kerning adjustment between two glyphs.
	Kerning(id ID, left, right GlyphID, size float32) float32

	// HasGlyph returns true if the font contains a glyph for the given rune.
	HasGlyph(id ID, r rune) bool

	// HasColorGlyphs returns true if the font contains color glyph data
	// (e.g., COLR/CPAL, CBDT/CBLC, or sbix tables for color emoji).
	HasColorGlyphs(id ID) bool

	// SetDPIScale sets the display DPI scale factor (1.0 = 96 DPI).
	// Font sizes passed to other methods are treated as points and
	// converted to pixels using: pixels = points * dpiScale * 96/72.
	// Default is 1.0 (standard 96 DPI).
	SetDPIScale(scale float32)

	// Destroy releases all resources.
	Destroy()
}

// Manager manages font registration, lookup, and fallback chains.
// This is an aggregate root for the font bounded context.
type Manager interface {
	// Register registers a font with a family name.
	Register(family string, weight Weight, style Style, data []byte) (ID, error)

	// RegisterFile registers a font from a file path.
	RegisterFile(family string, weight Weight, style Style, path string) (ID, error)

	// Resolve finds the best matching font for the given properties.
	Resolve(props Properties) (ID, bool)

	// ResolveRune finds a font that can render the given rune,
	// following the fallback chain if needed.
	ResolveRune(props Properties, r rune) (ID, bool)

	// SetFallbackChain sets the font fallback order for a locale.
	SetFallbackChain(locale string, families []string)

	// Engine returns the underlying font engine.
	Engine() Engine
}

// Shaper performs text layout (shaping, line breaking, alignment).
type Shaper interface {
	// Shape converts text into positioned glyph runs.
	Shape(text string, opts ShapeOptions) []GlyphRun

	// Measure measures text without full shaping.
	Measure(text string, opts ShapeOptions) TextMetrics
}

// ShapeOptions controls text shaping. Value object.
type ShapeOptions struct {
	FontID          ID
	FallbackFontIDs []ID     // Fallback fonts tried in order when a glyph is missing from FontID
	FontSize        float32
	MaxWidth        float32      // 0 = no wrapping
	LineHeight      float32      // 0 = use font default
	Align           TextAlign
	Locale          string       // For locale-specific line breaking rules
	Segmenter       Segmenter    // Optional: CJK word segmenter for smart line breaking
	Truncate        TruncateMode // Truncation mode when text exceeds MaxWidth
	MaxLines        int          // 0 = unlimited; used with Truncate
	Ellipsis        string       // Ellipsis string, defaults to "…"
}

// TextAlign specifies horizontal text alignment.
type TextAlign uint8

const (
	TextAlignLeft TextAlign = iota
	TextAlignCenter
	TextAlignRight
	TextAlignJustify
)

// Segmenter performs word segmentation for CJK text.
// Used by the shaper for word-boundary line breaking and truncation.
type Segmenter interface {
	// Segment splits text into word tokens.
	Segment(text string) []string
}

// TruncateMode specifies how text is truncated when it exceeds MaxWidth.
type TruncateMode uint8

const (
	TruncateNone     TruncateMode = iota // No truncation (default)
	TruncateChar                          // Truncate at character boundary + ellipsis
	TruncateWord                          // Truncate at word boundary + ellipsis
)

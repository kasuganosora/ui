//go:build windows

package gl

// OpenGL 3.3 Core constants.

// Data types
const (
	GL_FLOAT         = 0x1406
	GL_UNSIGNED_BYTE = 0x1401
	GL_UNSIGNED_INT  = 0x1405
	GL_INT           = 0x1404
	GL_TRUE          = 1
	GL_FALSE         = 0
)

// Primitive types
const (
	GL_TRIANGLES = 0x0004
)

// Buffer targets
const (
	GL_ARRAY_BUFFER         = 0x8892
	GL_ELEMENT_ARRAY_BUFFER = 0x8893
)

// Buffer usage
const (
	GL_STATIC_DRAW  = 0x88E4
	GL_DYNAMIC_DRAW = 0x88E8
	GL_STREAM_DRAW  = 0x88E0
)

// Shader types
const (
	GL_VERTEX_SHADER   = 0x8B31
	GL_FRAGMENT_SHADER = 0x8B30
)

// Shader queries
const (
	GL_COMPILE_STATUS  = 0x8B81
	GL_LINK_STATUS     = 0x8B82
	GL_INFO_LOG_LENGTH = 0x8B84
)

// Texture targets and params
const (
	GL_TEXTURE_2D = 0x0DE1
	GL_TEXTURE0   = 0x84C0

	GL_TEXTURE_MIN_FILTER = 0x2801
	GL_TEXTURE_MAG_FILTER = 0x2800
	GL_TEXTURE_WRAP_S     = 0x2802
	GL_TEXTURE_WRAP_T     = 0x2803

	GL_LINEAR            = 0x2601
	GL_NEAREST           = 0x2600
	GL_CLAMP_TO_EDGE     = 0x812F
	GL_UNPACK_ALIGNMENT  = 0x0CF5
	GL_PACK_ALIGNMENT    = 0x0D05
	GL_UNPACK_ROW_LENGTH = 0x0CF2
)

// Pixel formats
const (
	GL_RED  = 0x1903
	GL_RGBA = 0x1908
	GL_BGRA = 0x80E1
	GL_R8   = 0x8229
	GL_RGBA8 = 0x8058
	GL_SRGB8_ALPHA8 = 0x8C43
)

// Enable caps
const (
	GL_BLEND        = 0x0BE2
	GL_SCISSOR_TEST = 0x0C11
	GL_DEPTH_TEST   = 0x0B71
)

// Blend factors / equations
const (
	GL_SRC_ALPHA           = 0x0302
	GL_ONE_MINUS_SRC_ALPHA = 0x0303
	GL_ONE                 = 1
	GL_ZERO                = 0
	GL_FUNC_ADD            = 0x8006
)

// Framebuffer
const (
	GL_FRAMEBUFFER               = 0x8D40
	GL_READ_FRAMEBUFFER          = 0x8C36
	GL_COLOR_ATTACHMENT0         = 0x8CE0
	GL_FRAMEBUFFER_COMPLETE      = 0x8CD5
	GL_FRAMEBUFFER_SRGB          = 0x8DB9
)

// Get queries
const (
	GL_MAX_TEXTURE_SIZE = 0x0D33
	GL_VIEWPORT         = 0x0BA2
	GL_SCISSOR_BOX      = 0x0C10
)

// Clear bits
const (
	GL_COLOR_BUFFER_BIT = 0x00004000
)

// Map access
const (
	GL_MAP_WRITE_BIT              = 0x0002
	GL_MAP_INVALIDATE_BUFFER_BIT  = 0x0008
)

// RectVertex matches the vertex layout used by both Vulkan and DX11 backends.
// 19 float32 fields = 76 bytes per vertex.
type RectVertex struct {
	PosX, PosY     float32 // NDC position
	U, V           float32 // UV for SDF (0..1)
	ColorR, ColorG, ColorB, ColorA float32 // Fill color
	RectW, RectH   float32 // Rect size in pixels (for SDF)
	RadiusTL, RadiusTR, RadiusBR, RadiusBL float32 // Corner radii
	BorderWidth    float32
	BorderR, BorderG, BorderB, BorderA float32 // Border color
}

// TexturedVertex for text glyphs and images.
// 8 float32 fields = 32 bytes per vertex.
type TexturedVertex struct {
	PosX, PosY     float32 // NDC position
	U, V           float32 // Texture UV
	ColorR, ColorG, ColorB, ColorA float32 // Tint color
}

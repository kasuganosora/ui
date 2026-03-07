package render

import (
	"image"

	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
)

// Backend is the rendering abstraction interface.
// Implementations: Vulkan, OpenGL, DirectX, Software.
// This is a DDD anti-corruption layer boundary.
type Backend interface {
	// Init initializes the rendering backend for the given window.
	Init(window platform.Window) error

	// BeginFrame starts a new rendering frame.
	BeginFrame()

	// EndFrame finishes the current frame and presents it.
	EndFrame()

	// Resize handles window resize events.
	Resize(width, height int)

	// Submit submits a command buffer for rendering.
	Submit(buf *CommandBuffer)

	// CreateTexture creates a GPU texture.
	CreateTexture(desc TextureDesc) (TextureHandle, error)

	// UpdateTexture uploads data to a region of an existing texture.
	UpdateTexture(handle TextureHandle, region uimath.Rect, data []byte) error

	// DestroyTexture frees a GPU texture.
	DestroyTexture(handle TextureHandle)

	// MaxTextureSize returns the maximum supported texture dimension.
	MaxTextureSize() int

	// ReadPixels reads the current framebuffer contents as an RGBA image.
	// Must be called after Submit and before the next BeginFrame.
	// Returns nil if readback is not supported by the backend.
	ReadPixels() (*image.RGBA, error)

	// Destroy releases all GPU resources.
	Destroy()
}

// TextureHandle is an opaque handle to a GPU texture.
type TextureHandle uint64

const InvalidTexture TextureHandle = 0

// TextureDesc describes a texture to create.
type TextureDesc struct {
	Width  int
	Height int
	Format TextureFormat
	Data   []byte // Initial data (optional, can be nil)
}

// TextureFormat specifies the pixel format of a texture.
type TextureFormat uint8

const (
	TextureFormatR8     TextureFormat = iota // Single channel, 8 bits (SDF glyphs)
	TextureFormatRGBA8                       // 4 channels, 8 bits each (color images)
	TextureFormatBGRA8                       // 4 channels, BGRA order
)

package render

import uimath "github.com/kasuganosora/ui/math"

// CommandType identifies the type of render command.
type CommandType uint8

const (
	CmdRect  CommandType = iota + 1 // Filled/stroked rectangle
	CmdText                         // Text run
	CmdImage                        // Textured rectangle
	CmdClip                         // Set scissor rect
	CmdPath                         // Custom path (future)
)

// Command is a single render command. Value object.
type Command struct {
	Type    CommandType
	ZOrder  int32
	Opacity float32 // 0..1

	// Used by CmdRect
	Rect       *RectCmd
	// Used by CmdText
	Text       *TextCmd
	// Used by CmdImage
	Image      *ImageCmd
	// Used by CmdClip
	Clip       *ClipCmd
}

// RectCmd describes a rectangle draw.
type RectCmd struct {
	Bounds      uimath.Rect
	Corners     uimath.Corners // Corner radii
	FillColor   uimath.Color
	BorderColor uimath.Color
	BorderWidth float32
	// Gradient (optional)
	GradientStart uimath.Color
	GradientEnd   uimath.Color
	GradientAngle float32 // In radians; 0 = no gradient
}

// TextCmd describes a text draw.
type TextCmd struct {
	X, Y     float32
	Glyphs   []GlyphInstance
	Color    uimath.Color
	FontSize float32
	Atlas    TextureHandle // Glyph atlas texture
}

// GlyphInstance describes a single glyph to render from an atlas.
type GlyphInstance struct {
	// Position in screen space (top-left of glyph)
	X, Y float32
	// Size in screen space
	Width, Height float32
	// UV coordinates in the atlas texture
	U0, V0, U1, V1 float32
}

// ImageCmd describes an image draw.
type ImageCmd struct {
	Texture  TextureHandle
	SrcRect  uimath.Rect // Region in texture (UV space or pixels)
	DstRect  uimath.Rect // Destination on screen
	Tint     uimath.Color
	Corners  uimath.Corners // Corner radii for rounded image
}

// ClipCmd sets the scissor rectangle.
type ClipCmd struct {
	Bounds uimath.Rect
}

// CommandBuffer collects render commands for a frame.
// This is an aggregate root - external code must use its methods.
type CommandBuffer struct {
	commands []Command
}

// NewCommandBuffer creates an empty command buffer.
func NewCommandBuffer() *CommandBuffer {
	return &CommandBuffer{
		commands: make([]Command, 0, 256),
	}
}

// Reset clears all commands for reuse.
func (cb *CommandBuffer) Reset() {
	cb.commands = cb.commands[:0]
}

// DrawRect adds a rectangle draw command.
func (cb *CommandBuffer) DrawRect(cmd RectCmd, zOrder int32, opacity float32) {
	cb.commands = append(cb.commands, Command{
		Type:    CmdRect,
		ZOrder:  zOrder,
		Opacity: opacity,
		Rect:    &cmd,
	})
}

// DrawText adds a text draw command.
func (cb *CommandBuffer) DrawText(cmd TextCmd, zOrder int32, opacity float32) {
	cb.commands = append(cb.commands, Command{
		Type:    CmdText,
		ZOrder:  zOrder,
		Opacity: opacity,
		Text:    &cmd,
	})
}

// DrawImage adds an image draw command.
func (cb *CommandBuffer) DrawImage(cmd ImageCmd, zOrder int32, opacity float32) {
	cb.commands = append(cb.commands, Command{
		Type:    CmdImage,
		ZOrder:  zOrder,
		Opacity: opacity,
		Image:   &cmd,
	})
}

// PushClip pushes a scissor rectangle.
func (cb *CommandBuffer) PushClip(bounds uimath.Rect) {
	cb.commands = append(cb.commands, Command{
		Type: CmdClip,
		Clip: &ClipCmd{Bounds: bounds},
	})
}

// Commands returns the collected commands (read-only view).
func (cb *CommandBuffer) Commands() []Command {
	return cb.commands
}

// Len returns the number of commands.
func (cb *CommandBuffer) Len() int {
	return len(cb.commands)
}

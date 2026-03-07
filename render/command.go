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
	overlays []Command // rendered after all commands, without clip
}

// NewCommandBuffer creates an empty command buffer.
func NewCommandBuffer() *CommandBuffer {
	return &CommandBuffer{
		commands: make([]Command, 0, 256),
		overlays: make([]Command, 0, 16),
	}
}

// Reset clears all commands for reuse.
func (cb *CommandBuffer) Reset() {
	cb.commands = cb.commands[:0]
	cb.overlays = cb.overlays[:0]
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

// PopClip resets the scissor to the full viewport (max bounds).
func (cb *CommandBuffer) PopClip() {
	cb.commands = append(cb.commands, Command{
		Type: CmdClip,
		Clip: &ClipCmd{Bounds: uimath.NewRect(0, 0, 1e6, 1e6)},
	})
}

// DrawOverlay adds a rect command to the overlay layer.
// Overlays are rendered after all normal commands with no clip applied.
func (cb *CommandBuffer) DrawOverlay(cmd RectCmd, zOrder int32, opacity float32) {
	cb.overlays = append(cb.overlays, Command{
		Type:    CmdRect,
		ZOrder:  zOrder,
		Opacity: opacity,
		Rect:    &cmd,
	})
}

// DrawOverlayTextCmd adds a text command to the overlay layer.
func (cb *CommandBuffer) DrawOverlayTextCmd(cmd TextCmd, zOrder int32, opacity float32) {
	cb.overlays = append(cb.overlays, Command{
		Type:    CmdText,
		ZOrder:  zOrder,
		Opacity: opacity,
		Text:    &cmd,
	})
}

// MoveToOverlay moves commands added after position 'from' (by count of
// commands + overlays) from the main command list into the overlay list,
// overriding their z-order.
func (cb *CommandBuffer) MoveToOverlay(fromLen int, zOrder int32) {
	// fromLen is the value of Len() before the commands were added.
	// Since overlays count is unchanged, the new commands are in cb.commands.
	mainBefore := fromLen - len(cb.overlays)
	if mainBefore < 0 {
		mainBefore = 0
	}
	if mainBefore >= len(cb.commands) {
		return
	}
	for _, c := range cb.commands[mainBefore:] {
		c.ZOrder = zOrder
		cb.overlays = append(cb.overlays, c)
	}
	cb.commands = cb.commands[:mainBefore]
}

// Commands returns the collected commands (read-only view).
func (cb *CommandBuffer) Commands() []Command {
	return cb.commands
}

// Overlays returns the overlay commands (read-only view).
func (cb *CommandBuffer) Overlays() []Command {
	return cb.overlays
}

// Len returns the number of commands.
func (cb *CommandBuffer) Len() int {
	return len(cb.commands) + len(cb.overlays)
}

package render

import uimath "github.com/kasuganosora/ui/math"

// CommandType identifies the type of render command.
type CommandType uint8

const (
	CmdRect   CommandType = iota + 1 // Filled/stroked rectangle
	CmdText                          // Text run
	CmdImage                         // Textured rectangle
	CmdClip                          // Set scissor rect
	CmdShadow                        // Box shadow (SDF-blurred rounded rect)
	CmdPath                          // Custom path (future)
)

// Command is a single render command. Value object.
type Command struct {
	Type    CommandType
	ZOrder  int32
	Opacity float32 // 0..1

	// Used by CmdRect
	Rect *RectCmd
	// Used by CmdText
	Text *TextCmd
	// Used by CmdImage
	Image *ImageCmd
	// Used by CmdClip
	Clip *ClipCmd
	// Used by CmdShadow
	Shadow *ShadowCmd
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

// ShadowCmd describes a CSS box-shadow layer.
// The shadow is rendered as a SDF-blurred rounded rect behind the element.
type ShadowCmd struct {
	Bounds       uimath.Rect    // element's own bounds (logical px)
	Corners      uimath.Corners // element's corner radii (logical px)
	OffsetX      float32        // shadow X offset (logical px)
	OffsetY      float32        // shadow Y offset (logical px)
	BlurRadius   float32        // blur radius (logical px); 0 = hard edge
	SpreadRadius float32        // positive = expand, negative = shrink shadow shape
	Color        uimath.Color   // shadow color (including alpha)
	Inset        bool           // true = inset shadow (rendered inside the element)
}

// ClipCmd sets the scissor rectangle.
type ClipCmd struct {
	Bounds uimath.Rect
}

// CommandBuffer collects render commands for a frame.
// Uses arena-style allocation for command structs — zero GC pressure per frame.
type CommandBuffer struct {
	commands  []Command
	overlays  []Command // rendered after all commands, without clip
	clipStack []uimath.Rect

	// Arena pools: grow as needed, reset index to 0 each frame (no allocs after warm-up).
	rectArena   []RectCmd
	rectIdx     int
	textArena   []TextCmd
	textIdx     int
	imageArena  []ImageCmd
	imageIdx    int
	clipArena   []ClipCmd
	clipIdx     int
	shadowArena []ShadowCmd
	shadowIdx   int
}

// NewCommandBuffer creates an empty command buffer.
func NewCommandBuffer() *CommandBuffer {
	return &CommandBuffer{
		commands:    make([]Command, 0, 256),
		overlays:    make([]Command, 0, 16),
		rectArena:   make([]RectCmd, 0, 128),
		textArena:   make([]TextCmd, 0, 64),
		imageArena:  make([]ImageCmd, 0, 32),
		clipArena:   make([]ClipCmd, 0, 32),
		shadowArena: make([]ShadowCmd, 0, 16),
	}
}

// Reset clears the buffer for reuse. Arena memory is retained for next frame.
func (cb *CommandBuffer) Reset() {
	cb.commands = cb.commands[:0]
	cb.overlays = cb.overlays[:0]
	cb.clipStack = cb.clipStack[:0]
	cb.rectIdx = 0
	cb.textIdx = 0
	cb.imageIdx = 0
	cb.clipIdx = 0
	cb.shadowIdx = 0
}

// acquireRect returns a pointer to a RectCmd in the arena.
func (cb *CommandBuffer) acquireRect() *RectCmd {
	if cb.rectIdx >= len(cb.rectArena) {
		cb.rectArena = append(cb.rectArena, RectCmd{})
	}
	p := &cb.rectArena[cb.rectIdx]
	cb.rectIdx++
	return p
}

func (cb *CommandBuffer) acquireText() *TextCmd {
	if cb.textIdx >= len(cb.textArena) {
		cb.textArena = append(cb.textArena, TextCmd{})
	}
	p := &cb.textArena[cb.textIdx]
	cb.textIdx++
	return p
}

func (cb *CommandBuffer) acquireImage() *ImageCmd {
	if cb.imageIdx >= len(cb.imageArena) {
		cb.imageArena = append(cb.imageArena, ImageCmd{})
	}
	p := &cb.imageArena[cb.imageIdx]
	cb.imageIdx++
	return p
}

func (cb *CommandBuffer) acquireClip() *ClipCmd {
	if cb.clipIdx >= len(cb.clipArena) {
		cb.clipArena = append(cb.clipArena, ClipCmd{})
	}
	p := &cb.clipArena[cb.clipIdx]
	cb.clipIdx++
	return p
}

func (cb *CommandBuffer) acquireShadow() *ShadowCmd {
	if cb.shadowIdx >= len(cb.shadowArena) {
		cb.shadowArena = append(cb.shadowArena, ShadowCmd{})
	}
	p := &cb.shadowArena[cb.shadowIdx]
	cb.shadowIdx++
	return p
}

// DrawRect adds a rectangle draw command.
func (cb *CommandBuffer) DrawRect(cmd RectCmd, zOrder int32, opacity float32) {
	rc := cb.acquireRect()
	*rc = cmd
	cb.commands = append(cb.commands, Command{
		Type:    CmdRect,
		ZOrder:  zOrder,
		Opacity: opacity,
		Rect:    rc,
	})
}

// DrawText adds a text draw command.
func (cb *CommandBuffer) DrawText(cmd TextCmd, zOrder int32, opacity float32) {
	tc := cb.acquireText()
	*tc = cmd
	cb.commands = append(cb.commands, Command{
		Type:    CmdText,
		ZOrder:  zOrder,
		Opacity: opacity,
		Text:    tc,
	})
}

// DrawImage adds an image draw command.
func (cb *CommandBuffer) DrawImage(cmd ImageCmd, zOrder int32, opacity float32) {
	ic := cb.acquireImage()
	*ic = cmd
	cb.commands = append(cb.commands, Command{
		Type:    CmdImage,
		ZOrder:  zOrder,
		Opacity: opacity,
		Image:   ic,
	})
}

// DrawShadow adds a box-shadow draw command. The shadow is rendered behind
// the element using SDF distance-field blur. Call before DrawRect for the
// same element so the shadow appears beneath it.
func (cb *CommandBuffer) DrawShadow(cmd ShadowCmd, zOrder int32, opacity float32) {
	sc := cb.acquireShadow()
	*sc = cmd
	cb.commands = append(cb.commands, Command{
		Type:    CmdShadow,
		ZOrder:  zOrder,
		Opacity: opacity,
		Shadow:  sc,
	})
}

// CurrentClip returns the current active clip rectangle.
// Returns a very large rect if no clip is active (everything visible).
func (cb *CommandBuffer) CurrentClip() uimath.Rect {
	if len(cb.clipStack) > 0 {
		return cb.clipStack[len(cb.clipStack)-1]
	}
	return uimath.NewRect(0, 0, 1e6, 1e6)
}

// PushClip pushes a scissor rectangle, intersecting with any active clip.
func (cb *CommandBuffer) PushClip(bounds uimath.Rect) {
	if len(cb.clipStack) > 0 {
		bounds = cb.clipStack[len(cb.clipStack)-1].Intersection(bounds)
	}
	cb.clipStack = append(cb.clipStack, bounds)
	cc := cb.acquireClip()
	cc.Bounds = bounds
	cb.commands = append(cb.commands, Command{
		Type: CmdClip,
		Clip: cc,
	})
}

// PopClip restores the scissor to the previous clip rect (or full viewport if none).
func (cb *CommandBuffer) PopClip() {
	if len(cb.clipStack) > 0 {
		cb.clipStack = cb.clipStack[:len(cb.clipStack)-1]
	}
	cc := cb.acquireClip()
	if len(cb.clipStack) > 0 {
		cc.Bounds = cb.clipStack[len(cb.clipStack)-1]
	} else {
		cc.Bounds = uimath.NewRect(0, 0, 1e6, 1e6)
	}
	cb.commands = append(cb.commands, Command{
		Type: CmdClip,
		Clip: cc,
	})
}

// DrawOverlay adds a rect command to the overlay layer.
// Overlays are rendered after all normal commands with no clip applied.
func (cb *CommandBuffer) DrawOverlay(cmd RectCmd, zOrder int32, opacity float32) {
	rc := cb.acquireRect()
	*rc = cmd
	cb.overlays = append(cb.overlays, Command{
		Type:    CmdRect,
		ZOrder:  zOrder,
		Opacity: opacity,
		Rect:    rc,
	})
}

// DrawOverlayTextCmd adds a text command to the overlay layer.
func (cb *CommandBuffer) DrawOverlayTextCmd(cmd TextCmd, zOrder int32, opacity float32) {
	tc := cb.acquireText()
	*tc = cmd
	cb.overlays = append(cb.overlays, Command{
		Type:    CmdText,
		ZOrder:  zOrder,
		Opacity: opacity,
		Text:    tc,
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

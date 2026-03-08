//go:build windows

package gl

import (
	"fmt"
	"image"
	"math"
	"sort"
	"unsafe"

	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/render"
)

// textureEntry stores GPU resources for a texture.
type textureEntry struct {
	glTexture uint32
	width     int
	height    int
	format    render.TextureFormat
}

// Backend implements render.Backend using OpenGL 3.3 Core.
// Zero-CGO: all GL calls go through syscall.SyscallN.
type Backend struct {
	loader *Loader

	// Platform (Windows)
	hwnd  uintptr
	hdc   uintptr
	hglrc uintptr

	// Shader programs
	rectProgram     uint32
	texturedProgram uint32
	textProgram     uint32

	// Uniform locations
	texturedTexLoc int32
	textTexLoc     int32

	// VAO + VBO
	rectVAO, rectVBO         uint32
	texturedVAO, texturedVBO uint32

	// State
	width, height   int // from FramebufferSize (for NDC calculations, matches VK/DX11)
	vpWidth, vpHeight int // actual physical viewport pixels (from GetClientRect)
	dpiScale        float32
	shared          bool // true when using external GL context (embedded mode)

	// Textures
	nextTextureID render.TextureHandle
	textures      map[render.TextureHandle]*textureEntry
}

// New creates a new OpenGL backend.
func New() *Backend {
	return &Backend{
		nextTextureID: 1,
		textures:      make(map[render.TextureHandle]*textureEntry),
	}
}

// Init initializes the GL backend for the given window.
func (b *Backend) Init(window platform.Window) error {
	var err error
	b.loader, err = NewLoader()
	if err != nil {
		return err
	}

	b.hwnd = window.NativeHandle()
	b.dpiScale = window.DPIScale()
	w, h := window.FramebufferSize()
	b.width, b.height = w, h
	// Get actual physical viewport from GetClientRect (FramebufferSize may overestimate on DPI-aware windows)
	b.vpWidth, b.vpHeight = b.loader.GetClientRect(b.hwnd)

	// Create WGL context
	b.hdc, b.hglrc, err = b.createContext(b.hwnd)
	if err != nil {
		return err
	}

	// Load extension functions (need current context)
	if err := b.loader.LoadExtensionFunctions(); err != nil {
		b.destroyContext()
		return err
	}

	return b.initPipeline()
}

// InitShared initializes the GL backend for use in an existing GL context.
// The caller must ensure the GL context is current before calling this.
func (b *Backend) InitShared(width, height int, dpiScale float32) error {
	var err error
	b.loader, err = NewLoader()
	if err != nil {
		return err
	}

	b.shared = true
	b.width, b.height = width, height
	b.vpWidth, b.vpHeight = width, height // shared mode: caller provides real viewport size
	b.dpiScale = dpiScale

	// Load functions — context must already be current
	b.loader.LoadCoreFunctions()
	if err := b.loader.LoadExtensionFunctions(); err != nil {
		return err
	}

	return b.initPipeline()
}

// initPipeline creates shaders, VAOs, VBOs.
func (b *Backend) initPipeline() error {
	l := b.loader
	var err error

	// Compile shader programs
	b.rectProgram, err = l.createProgram(rectVertSrc, rectFragSrc)
	if err != nil {
		return fmt.Errorf("gl: rect program: %w", err)
	}
	b.texturedProgram, err = l.createProgram(texturedVertSrc, texturedFragSrc)
	if err != nil {
		return fmt.Errorf("gl: textured program: %w", err)
	}
	b.textProgram, err = l.createProgram(texturedVertSrc, textFragSrc)
	if err != nil {
		return fmt.Errorf("gl: text program: %w", err)
	}

	// Get uniform locations
	b.texturedTexLoc = l.getUniformLocation(b.texturedProgram, "tex")
	b.textTexLoc = l.getUniformLocation(b.textProgram, "glyphAtlas")

	// Create rect VAO/VBO
	b.rectVAO, b.rectVBO = b.createRectVAO()
	// Create textured VAO/VBO (shared by text + image pipelines)
	b.texturedVAO, b.texturedVBO = b.createTexturedVAO()

	// Enable sRGB framebuffer for linear-space alpha blending
	glCall(l.glEnable, GL_FRAMEBUFFER_SRGB)

	return nil
}

func (b *Backend) createRectVAO() (uint32, uint32) {
	l := b.loader
	var vao, vbo uint32
	glCall(l.glGenVertexArrays, 1, uintptr(unsafe.Pointer(&vao)))
	glCall(l.glGenBuffers, 1, uintptr(unsafe.Pointer(&vbo)))

	glCall(l.glBindVertexArray, uintptr(vao))
	glCall(l.glBindBuffer, GL_ARRAY_BUFFER, uintptr(vbo))

	stride := uintptr(unsafe.Sizeof(RectVertex{}))

	// location 0: inPos (vec2)
	glCall(l.glEnableVertexAttribArray, 0)
	glCall(l.glVertexAttribPointer, 0, 2, GL_FLOAT, GL_FALSE, stride, 0)
	// location 1: inUV (vec2)
	glCall(l.glEnableVertexAttribArray, 1)
	glCall(l.glVertexAttribPointer, 1, 2, GL_FLOAT, GL_FALSE, stride, 2*4)
	// location 2: inColor (vec4)
	glCall(l.glEnableVertexAttribArray, 2)
	glCall(l.glVertexAttribPointer, 2, 4, GL_FLOAT, GL_FALSE, stride, 4*4)
	// location 3: inRectSize (vec2)
	glCall(l.glEnableVertexAttribArray, 3)
	glCall(l.glVertexAttribPointer, 3, 2, GL_FLOAT, GL_FALSE, stride, 8*4)
	// location 4: inRadii (vec4)
	glCall(l.glEnableVertexAttribArray, 4)
	glCall(l.glVertexAttribPointer, 4, 4, GL_FLOAT, GL_FALSE, stride, 10*4)
	// location 5: inBorderWidth (float)
	glCall(l.glEnableVertexAttribArray, 5)
	glCall(l.glVertexAttribPointer, 5, 1, GL_FLOAT, GL_FALSE, stride, 14*4)
	// location 6: inBorderColor (vec4)
	glCall(l.glEnableVertexAttribArray, 6)
	glCall(l.glVertexAttribPointer, 6, 4, GL_FLOAT, GL_FALSE, stride, 15*4)

	glCall(l.glBindVertexArray, 0)
	return vao, vbo
}

func (b *Backend) createTexturedVAO() (uint32, uint32) {
	l := b.loader
	var vao, vbo uint32
	glCall(l.glGenVertexArrays, 1, uintptr(unsafe.Pointer(&vao)))
	glCall(l.glGenBuffers, 1, uintptr(unsafe.Pointer(&vbo)))

	glCall(l.glBindVertexArray, uintptr(vao))
	glCall(l.glBindBuffer, GL_ARRAY_BUFFER, uintptr(vbo))

	stride := uintptr(unsafe.Sizeof(TexturedVertex{}))

	// location 0: inPos (vec2)
	glCall(l.glEnableVertexAttribArray, 0)
	glCall(l.glVertexAttribPointer, 0, 2, GL_FLOAT, GL_FALSE, stride, 0)
	// location 1: inUV (vec2)
	glCall(l.glEnableVertexAttribArray, 1)
	glCall(l.glVertexAttribPointer, 1, 2, GL_FLOAT, GL_FALSE, stride, 2*4)
	// location 2: inColor (vec4)
	glCall(l.glEnableVertexAttribArray, 2)
	glCall(l.glVertexAttribPointer, 2, 4, GL_FLOAT, GL_FALSE, stride, 4*4)

	glCall(l.glBindVertexArray, 0)
	return vao, vbo
}

// BeginFrame starts a new frame.
func (b *Backend) BeginFrame() {}

// EndFrame finishes the current frame.
func (b *Backend) EndFrame() {}

// Resize handles window resize.
func (b *Backend) Resize(width, height int) {
	b.width, b.height = width, height
	if b.hwnd != 0 {
		b.vpWidth, b.vpHeight = b.loader.GetClientRect(b.hwnd)
	} else {
		b.vpWidth, b.vpHeight = width, height
	}
}

// Submit renders a command buffer.
func (b *Backend) Submit(buf *render.CommandBuffer) {
	l := b.loader

	// Set viewport to actual physical window size
	glCall(l.glViewport, 0, 0, uintptr(b.vpWidth), uintptr(b.vpHeight))

	// Clear
	glCall(l.glClearColor,
		uintptr(math.Float32bits(0)),
		uintptr(math.Float32bits(0)),
		uintptr(math.Float32bits(0)),
		uintptr(math.Float32bits(1)))
	glCall(l.glClear, GL_COLOR_BUFFER_BIT)

	// Enable blending
	glCall(l.glEnable, GL_BLEND)
	glCall(l.glBlendFuncSeparate,
		GL_SRC_ALPHA, GL_ONE_MINUS_SRC_ALPHA,
		GL_ONE, GL_ONE_MINUS_SRC_ALPHA)
	glCall(l.glBlendEquation, GL_FUNC_ADD)

	// Disable depth test
	glCall(l.glDisable, GL_DEPTH_TEST)

	// Enable scissor
	glCall(l.glEnable, GL_SCISSOR_TEST)
	glCall(l.glScissor, 0, 0, uintptr(b.vpWidth), uintptr(b.vpHeight))

	// Process commands in order (with overlays appended)
	cmds := buf.Commands()
	overlays := buf.Overlays()

	// Sort by z-order (stable)
	all := make([]render.Command, 0, len(cmds)+len(overlays))
	all = append(all, cmds...)
	all = append(all, overlays...)
	sort.SliceStable(all, func(i, j int) bool {
		return all[i].ZOrder < all[j].ZOrder
	})

	for _, c := range all {
		switch c.Type {
		case render.CmdClip:
			b.applyScissor(c.Clip)
		case render.CmdRect:
			b.renderRect(c)
		case render.CmdText:
			b.renderText(c)
		case render.CmdImage:
			b.renderImage(c)
		}
	}

	glCall(l.glDisable, GL_SCISSOR_TEST)

	// Present (only if we own the context)
	if !b.shared {
		b.swapBuffers()
	}
}

func (b *Backend) applyScissor(clip *render.ClipCmd) {
	l := b.loader
	logW := float32(b.width) / b.dpiScale
	logH := float32(b.height) / b.dpiScale
	vpW := float32(b.vpWidth)
	vpH := float32(b.vpHeight)

	x := int32(clip.Bounds.X / logW * vpW)
	y := int32(clip.Bounds.Y / logH * vpH)
	w := int32(clip.Bounds.Width / logW * vpW)
	h := int32(clip.Bounds.Height / logH * vpH)

	if x < 0 {
		w += x
		x = 0
	}
	if y < 0 {
		h += y
		y = 0
	}
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	if w > int32(b.vpWidth) {
		w = int32(b.vpWidth)
	}
	if h > int32(b.vpHeight) {
		h = int32(b.vpHeight)
	}

	// OpenGL scissor Y is bottom-up
	glCall(l.glScissor, uintptr(x), uintptr(int32(b.vpHeight)-y-h), uintptr(w), uintptr(h))
}

func (b *Backend) renderRect(c render.Command) {
	l := b.loader
	rect := c.Rect
	opacity := c.Opacity

	logW := float32(b.width) / b.dpiScale
	logH := float32(b.height) / b.dpiScale
	x, y, w, h := rect.Bounds.X, rect.Bounds.Y, rect.Bounds.Width, rect.Bounds.Height

	pad := float32(1.0)
	qx, qy := x-pad, y-pad
	qw, qh := w+pad*2, h+pad*2

	// NDC: OpenGL has Y-up, but we use same convention as Vulkan (Y-down in input, flip here)
	ndcX := (qx/logW)*2 - 1
	ndcY := 1 - (qy/logH)*2
	ndcW := (qw / logW) * 2
	ndcH := (qh / logH) * 2

	uvL := -pad / w
	uvT := -pad / h
	uvR := 1.0 + pad/w
	uvB := 1.0 + pad/h

	s := b.dpiScale
	r, g, bl, a := rect.FillColor.R, rect.FillColor.G, rect.FillColor.B, rect.FillColor.A*opacity

	makeVertex := func(px, py, u, v float32) RectVertex {
		return RectVertex{
			PosX: px, PosY: py, U: u, V: v,
			ColorR: r, ColorG: g, ColorB: bl, ColorA: a,
			RectW: w * s, RectH: h * s,
			RadiusTL: rect.Corners.TopLeft * s, RadiusTR: rect.Corners.TopRight * s,
			RadiusBR: rect.Corners.BottomRight * s, RadiusBL: rect.Corners.BottomLeft * s,
			BorderWidth: rect.BorderWidth * s,
			BorderR: rect.BorderColor.R, BorderG: rect.BorderColor.G,
			BorderB: rect.BorderColor.B, BorderA: rect.BorderColor.A,
		}
	}

	vertices := [6]RectVertex{
		makeVertex(ndcX, ndcY, uvL, uvT),
		makeVertex(ndcX+ndcW, ndcY, uvR, uvT),
		makeVertex(ndcX+ndcW, ndcY-ndcH, uvR, uvB),
		makeVertex(ndcX, ndcY, uvL, uvT),
		makeVertex(ndcX+ndcW, ndcY-ndcH, uvR, uvB),
		makeVertex(ndcX, ndcY-ndcH, uvL, uvB),
	}

	glCall(l.glUseProgram, uintptr(b.rectProgram))
	glCall(l.glBindVertexArray, uintptr(b.rectVAO))
	glCall(l.glBindBuffer, GL_ARRAY_BUFFER, uintptr(b.rectVBO))

	size := unsafe.Sizeof(RectVertex{}) * 6
	glCall(l.glBufferData, GL_ARRAY_BUFFER, size,
		uintptr(unsafe.Pointer(&vertices[0])), GL_STREAM_DRAW)
	glCall(l.glDrawArrays, GL_TRIANGLES, 0, 6)
}

func (b *Backend) renderText(c render.Command) {
	l := b.loader
	tc := c.Text
	entry, ok := b.textures[tc.Atlas]
	if !ok || len(tc.Glyphs) == 0 {
		return
	}

	logW := float32(b.width) / b.dpiScale
	logH := float32(b.height) / b.dpiScale

	vertices := make([]TexturedVertex, 0, len(tc.Glyphs)*6)
	for _, g := range tc.Glyphs {
		x0 := (g.X/logW)*2 - 1
		y0 := 1 - (g.Y/logH)*2
		x1 := ((g.X + g.Width) / logW) * 2 - 1
		y1 := 1 - ((g.Y+g.Height)/logH)*2

		clr := TexturedVertex{
			ColorR: tc.Color.R, ColorG: tc.Color.G,
			ColorB: tc.Color.B, ColorA: tc.Color.A * c.Opacity,
		}

		v0 := clr
		v0.PosX, v0.PosY, v0.U, v0.V = x0, y0, g.U0, g.V0
		v1 := clr
		v1.PosX, v1.PosY, v1.U, v1.V = x1, y0, g.U1, g.V0
		v2 := clr
		v2.PosX, v2.PosY, v2.U, v2.V = x1, y1, g.U1, g.V1
		v3 := clr
		v3.PosX, v3.PosY, v3.U, v3.V = x0, y1, g.U0, g.V1

		vertices = append(vertices, v0, v1, v2, v0, v2, v3)
	}

	glCall(l.glUseProgram, uintptr(b.textProgram))
	glCall(l.glActiveTexture, GL_TEXTURE0)
	glCall(l.glBindTexture, GL_TEXTURE_2D, uintptr(entry.glTexture))
	glCall(l.glUniform1i, uintptr(b.textTexLoc), 0)

	glCall(l.glBindVertexArray, uintptr(b.texturedVAO))
	glCall(l.glBindBuffer, GL_ARRAY_BUFFER, uintptr(b.texturedVBO))

	size := unsafe.Sizeof(TexturedVertex{}) * uintptr(len(vertices))
	glCall(l.glBufferData, GL_ARRAY_BUFFER, size,
		uintptr(unsafe.Pointer(&vertices[0])), GL_STREAM_DRAW)
	glCall(l.glDrawArrays, GL_TRIANGLES, 0, uintptr(len(vertices)))
}

func (b *Backend) renderImage(c render.Command) {
	l := b.loader
	ic := c.Image
	entry, ok := b.textures[ic.Texture]
	if !ok {
		return
	}

	logW := float32(b.width) / b.dpiScale
	logH := float32(b.height) / b.dpiScale

	dx, dy := ic.DstRect.X, ic.DstRect.Y
	dw, dh := ic.DstRect.Width, ic.DstRect.Height

	x0 := (dx/logW)*2 - 1
	y0 := 1 - (dy/logH)*2
	x1 := ((dx + dw) / logW) * 2 - 1
	y1 := 1 - ((dy+dh)/logH)*2

	// UV: use SrcRect if set, otherwise full texture
	u0, v0, u1, v1 := float32(0), float32(0), float32(1), float32(1)
	if ic.SrcRect.Width > 0 && ic.SrcRect.Height > 0 {
		tw, th := float32(entry.width), float32(entry.height)
		u0 = ic.SrcRect.X / tw
		v0 = ic.SrcRect.Y / th
		u1 = (ic.SrcRect.X + ic.SrcRect.Width) / tw
		v1 = (ic.SrcRect.Y + ic.SrcRect.Height) / th
	}

	tint := ic.Tint
	if tint == (uimath.Color{}) {
		tint = uimath.ColorWhite
	}
	tint.A *= c.Opacity

	vertices := [6]TexturedVertex{
		{x0, y0, u0, v0, tint.R, tint.G, tint.B, tint.A},
		{x1, y0, u1, v0, tint.R, tint.G, tint.B, tint.A},
		{x1, y1, u1, v1, tint.R, tint.G, tint.B, tint.A},
		{x0, y0, u0, v0, tint.R, tint.G, tint.B, tint.A},
		{x1, y1, u1, v1, tint.R, tint.G, tint.B, tint.A},
		{x0, y1, u0, v1, tint.R, tint.G, tint.B, tint.A},
	}

	glCall(l.glUseProgram, uintptr(b.texturedProgram))
	glCall(l.glActiveTexture, GL_TEXTURE0)
	glCall(l.glBindTexture, GL_TEXTURE_2D, uintptr(entry.glTexture))
	glCall(l.glUniform1i, uintptr(b.texturedTexLoc), 0)

	glCall(l.glBindVertexArray, uintptr(b.texturedVAO))
	glCall(l.glBindBuffer, GL_ARRAY_BUFFER, uintptr(b.texturedVBO))

	size := unsafe.Sizeof(TexturedVertex{}) * 6
	glCall(l.glBufferData, GL_ARRAY_BUFFER, size,
		uintptr(unsafe.Pointer(&vertices[0])), GL_STREAM_DRAW)
	glCall(l.glDrawArrays, GL_TRIANGLES, 0, 6)
}

// --- Texture management ---

func (b *Backend) CreateTexture(desc render.TextureDesc) (render.TextureHandle, error) {
	l := b.loader
	var tex uint32
	glCall(l.glGenTextures, 1, uintptr(unsafe.Pointer(&tex)))
	glCall(l.glBindTexture, GL_TEXTURE_2D, uintptr(tex))

	// Filter
	filter := uint32(GL_LINEAR)
	if desc.Filter == render.TextureFilterNearest {
		filter = GL_NEAREST
	}
	glCall(l.glTexParameteri, GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, uintptr(filter))
	glCall(l.glTexParameteri, GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, uintptr(filter))
	glCall(l.glTexParameteri, GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, GL_CLAMP_TO_EDGE)
	glCall(l.glTexParameteri, GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, GL_CLAMP_TO_EDGE)

	// Determine internal format and pixel format
	var internalFmt, pixelFmt, pixelType uint32
	switch desc.Format {
	case render.TextureFormatR8:
		internalFmt = GL_R8
		pixelFmt = GL_RED
		pixelType = GL_UNSIGNED_BYTE
		glCall(l.glPixelStorei, GL_UNPACK_ALIGNMENT, 1)
	case render.TextureFormatRGBA8:
		internalFmt = GL_SRGB8_ALPHA8
		pixelFmt = GL_RGBA
		pixelType = GL_UNSIGNED_BYTE
	case render.TextureFormatBGRA8:
		internalFmt = GL_SRGB8_ALPHA8
		pixelFmt = GL_BGRA
		pixelType = GL_UNSIGNED_BYTE
	default:
		internalFmt = GL_RGBA8
		pixelFmt = GL_RGBA
		pixelType = GL_UNSIGNED_BYTE
	}

	var dataPtr uintptr
	if len(desc.Data) > 0 {
		dataPtr = uintptr(unsafe.Pointer(&desc.Data[0]))
	}

	glCall(l.glTexImage2D, GL_TEXTURE_2D, 0, uintptr(internalFmt),
		uintptr(desc.Width), uintptr(desc.Height), 0,
		uintptr(pixelFmt), uintptr(pixelType), dataPtr)

	// Reset alignment
	glCall(l.glPixelStorei, GL_UNPACK_ALIGNMENT, 4)

	handle := b.nextTextureID
	b.nextTextureID++
	b.textures[handle] = &textureEntry{
		glTexture: tex,
		width:     desc.Width,
		height:    desc.Height,
		format:    desc.Format,
	}
	return handle, nil
}

func (b *Backend) UpdateTexture(handle render.TextureHandle, region uimath.Rect, data []byte) error {
	l := b.loader
	entry, ok := b.textures[handle]
	if !ok {
		return fmt.Errorf("gl: texture %d not found", handle)
	}

	glCall(l.glBindTexture, GL_TEXTURE_2D, uintptr(entry.glTexture))

	var pixelFmt, pixelType uint32
	switch entry.format {
	case render.TextureFormatR8:
		pixelFmt = GL_RED
		pixelType = GL_UNSIGNED_BYTE
		glCall(l.glPixelStorei, GL_UNPACK_ALIGNMENT, 1)
	case render.TextureFormatRGBA8:
		pixelFmt = GL_RGBA
		pixelType = GL_UNSIGNED_BYTE
	case render.TextureFormatBGRA8:
		pixelFmt = GL_BGRA
		pixelType = GL_UNSIGNED_BYTE
	default:
		pixelFmt = GL_RGBA
		pixelType = GL_UNSIGNED_BYTE
	}

	x, y := int32(region.X), int32(region.Y)
	w, h := int32(region.Width), int32(region.Height)

	glCall(l.glTexSubImage2D, GL_TEXTURE_2D, 0,
		uintptr(x), uintptr(y), uintptr(w), uintptr(h),
		uintptr(pixelFmt), uintptr(pixelType),
		uintptr(unsafe.Pointer(&data[0])))

	// Reset
	glCall(l.glPixelStorei, GL_UNPACK_ALIGNMENT, 4)
	glCall(l.glPixelStorei, GL_UNPACK_ROW_LENGTH, 0)

	return nil
}

func (b *Backend) DestroyTexture(handle render.TextureHandle) {
	entry, ok := b.textures[handle]
	if !ok {
		return
	}
	glCall(b.loader.glDeleteTextures, 1, uintptr(unsafe.Pointer(&entry.glTexture)))
	delete(b.textures, handle)
}

func (b *Backend) MaxTextureSize() int {
	var size int32
	glCall(b.loader.glGetIntegerv, GL_MAX_TEXTURE_SIZE, uintptr(unsafe.Pointer(&size)))
	if size <= 0 {
		return 4096
	}
	return int(size)
}

func (b *Backend) DPIScale() float32 {
	return b.dpiScale
}

func (b *Backend) ReadPixels() (*image.RGBA, error) {
	l := b.loader
	w, h := b.vpWidth, b.vpHeight
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	glCall(l.glReadPixels, 0, 0, uintptr(w), uintptr(h),
		GL_RGBA, GL_UNSIGNED_BYTE, uintptr(unsafe.Pointer(&img.Pix[0])))

	// OpenGL reads bottom-up, flip vertically
	stride := w * 4
	row := make([]byte, stride)
	for y := 0; y < h/2; y++ {
		top := y * stride
		bot := (h - 1 - y) * stride
		copy(row, img.Pix[top:top+stride])
		copy(img.Pix[top:top+stride], img.Pix[bot:bot+stride])
		copy(img.Pix[bot:bot+stride], row)
	}
	return img, nil
}

func (b *Backend) Destroy() {
	l := b.loader
	if l == nil {
		return
	}

	// Delete textures
	for _, entry := range b.textures {
		glCall(l.glDeleteTextures, 1, uintptr(unsafe.Pointer(&entry.glTexture)))
	}
	b.textures = nil

	// Delete VAO/VBO
	if b.rectVAO != 0 {
		glCall(l.glDeleteVertexArrays, 1, uintptr(unsafe.Pointer(&b.rectVAO)))
		glCall(l.glDeleteBuffers, 1, uintptr(unsafe.Pointer(&b.rectVBO)))
	}
	if b.texturedVAO != 0 {
		glCall(l.glDeleteVertexArrays, 1, uintptr(unsafe.Pointer(&b.texturedVAO)))
		glCall(l.glDeleteBuffers, 1, uintptr(unsafe.Pointer(&b.texturedVBO)))
	}

	// Delete programs
	if b.rectProgram != 0 {
		glCall(l.glDeleteProgram, uintptr(b.rectProgram))
	}
	if b.texturedProgram != 0 {
		glCall(l.glDeleteProgram, uintptr(b.texturedProgram))
	}
	if b.textProgram != 0 {
		glCall(l.glDeleteProgram, uintptr(b.textProgram))
	}

	// Release WGL context (only if we own it)
	if !b.shared {
		b.destroyContext()
	}
}

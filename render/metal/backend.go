//go:build darwin

package metal

import (
	"fmt"
	"image"
	"sync"
	"unsafe"

	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/render"
	"github.com/ebitengine/purego"
)

// ---- Vertex types (matching shaders.go attributes and render/dx11 layout) ----

// RectVertex: 19 float32 = 76 bytes
type RectVertex struct {
	PosX, PosY                              float32
	U, V                                    float32
	ColorR, ColorG, ColorB, ColorA          float32
	RectW, RectH                            float32
	RadiusTL, RadiusTR, RadiusBR, RadiusBL float32
	BorderWidth                             float32
	BorderR, BorderG, BorderB, BorderA      float32
}

// ShadowVertex: 15 float32 = 60 bytes
type ShadowVertex struct {
	PosX, PosY                              float32
	U, V                                    float32
	ColorR, ColorG, ColorB, ColorA          float32
	ElemW, ElemH                            float32
	RadiusTL, RadiusTR, RadiusBR, RadiusBL float32
	Blur                                    float32
}

// TexturedVertex: 8 float32 = 32 bytes
type TexturedVertex struct {
	PosX, PosY                     float32
	U, V                           float32
	ColorR, ColorG, ColorB, ColorA float32
}

// metalTexture holds GPU-side texture resources.
type metalTexture struct {
	tex    uintptr // id<MTLTexture>
	width  int
	height int
	format render.TextureFormat
	filter render.TextureFilter
}

// vertexBufSize is 8 MiB — sufficient for a dense frame of UI commands.
const vertexBufSize = 8 << 20

// Backend implements render.Backend using Metal via purego ObjC calls.
type Backend struct {
	device     uintptr // id<MTLDevice>
	cmdQueue   uintptr // id<MTLCommandQueue>
	metalLayer uintptr // CAMetalLayer*

	// Pipelines (id<MTLRenderPipelineState>)
	rectPipeline   uintptr
	shadowPipeline uintptr
	textPipeline   uintptr
	imagePipeline  uintptr

	// Samplers
	linearSampler  uintptr // id<MTLSamplerState>
	nearestSampler uintptr

	// Per-frame: current drawable and command buffer
	drawable  uintptr // id<CAMetalDrawable>
	cmdBuffer uintptr // id<MTLCommandBuffer>
	encoder   uintptr // id<MTLRenderCommandEncoder>

	// Vertex ring buffer
	vertexBuf    uintptr        // id<MTLBuffer>
	vertexPtr    unsafe.Pointer // mapped CPU pointer
	vertexOffset int            // current byte offset

	// State
	width, height int
	dpiScale      float32
	transparent   bool

	// Textures
	texMu     sync.RWMutex
	nextTexID render.TextureHandle
	textures  map[render.TextureHandle]*metalTexture
}

// New creates a new Metal backend.
func New() *Backend {
	return &Backend{
		nextTexID: 1,
		textures:  make(map[render.TextureHandle]*metalTexture),
		dpiScale:  1.0,
	}
}

// Init implements render.Backend.
func (b *Backend) Init(win platform.Window) error {
	if objc_msgSend == 0 {
		return fmt.Errorf("metal: ObjC runtime not loaded")
	}
	if MTLCreateSystemDefaultDevice == 0 {
		return fmt.Errorf("metal: Metal framework not loaded")
	}

	// Create Metal device
	b.device, _, _ = purego.SyscallN(MTLCreateSystemDefaultDevice)
	if b.device == 0 {
		return fmt.Errorf("metal: MTLCreateSystemDefaultDevice returned nil — Metal not supported on this device")
	}

	b.transparent = win.IsTransparent()

	// Framebuffer size (physical pixels)
	fbW, fbH := win.FramebufferSize()
	b.width, b.height = fbW, fbH
	b.dpiScale = win.DPIScale()
	if b.dpiScale <= 0 {
		b.dpiScale = 1.0
	}

	// Get NSView from window
	view := win.NativeHandle()
	if view == 0 {
		return fmt.Errorf("metal: window NativeHandle() returned nil")
	}

	// view.setWantsLayer(YES)
	msgSend(view, selSetWantsLayer, 1)

	// Create CAMetalLayer: [CAMetalLayer layer]
	caMetalLayerClass := objcClass("CAMetalLayer")
	if caMetalLayerClass == 0 {
		return fmt.Errorf("metal: CAMetalLayer class not found — QuartzCore not loaded")
	}
	b.metalLayer = msgSend(caMetalLayerClass, selLayer)
	if b.metalLayer == 0 {
		return fmt.Errorf("metal: CAMetalLayer layer returned nil")
	}

	// [view setLayer:metalLayer]
	msgSend(view, selSetLayerSel, b.metalLayer)

	// [metalLayer setDevice:device]
	msgSend(b.metalLayer, selSetDevice, b.device)

	// [metalLayer setPixelFormat:MTLPixelFormatBGRA8Unorm_sRGB]
	msgSend(b.metalLayer, selSetPixelFormat, uintptr(MTLPixelFormatBGRA8Unorm_sRGB))

	// [metalLayer setFramebufferOnly:NO] — we may need to read pixels
	msgSend(b.metalLayer, selSetFramebufferOnly, 0)

	// [metalLayer setDrawableSize:CGSize{fbW, fbH}]
	msgSendCGSize(b.metalLayer, selSetDrawableSize, float64(fbW), float64(fbH))

	// [metalLayer setContentsScale:dpiScale]
	msgSendCGFloat(b.metalLayer, selSetContentsScale, float64(b.dpiScale))

	// For transparent windows, make the layer non-opaque
	if b.transparent {
		msgSend(b.metalLayer, selSetOpaque, 0) // [metalLayer setOpaque:NO]
	}

	// Create command queue
	b.cmdQueue = msgSend(b.device, selNewCommandQueueSel)
	if b.cmdQueue == 0 {
		return fmt.Errorf("metal: newCommandQueue returned nil")
	}

	// Compile MSL shaders and create pipeline states
	if err := b.createPipelines(); err != nil {
		return err
	}

	// Create samplers
	if err := b.createSamplers(); err != nil {
		return err
	}

	// Allocate vertex ring buffer: [device newBufferWithLength:8MB options:0]
	b.vertexBuf = msgSend(b.device, selNewBufferWithLength,
		uintptr(vertexBufSize), uintptr(MTLResourceStorageModeShared))
	if b.vertexBuf == 0 {
		return fmt.Errorf("metal: failed to allocate vertex buffer")
	}

	// Map vertex buffer pointer: [buffer contents]
	b.vertexPtr = unsafe.Pointer(msgSend(b.vertexBuf, selContents))
	if b.vertexPtr == nil {
		return fmt.Errorf("metal: vertex buffer contents returned nil")
	}

	return nil
}

// nsStringMetal creates an NSString from a Go string using the metal package's symbols.
func nsStringMetal(s string) uintptr {
	cls := objcClass("NSString")
	if cls == 0 {
		return 0
	}
	b := append([]byte(s), 0)
	return msgSend(cls, selStringWithUTF8StringMetal, uintptr(unsafe.Pointer(&b[0])))
}

// errFromNSError extracts a string from an NSError pointer.
func errFromNSError(errPtr uintptr) string {
	if errPtr == 0 {
		return "<nil error>"
	}
	desc := msgSend(errPtr, selLocalizedDescription)
	if desc == 0 {
		return "<no description>"
	}
	cstr := msgSend(desc, selUTF8String)
	if cstr == 0 {
		return "<no string>"
	}
	return goStringMetal(cstr)
}

// goStringMetal converts a C string pointer to a Go string.
func goStringMetal(cstrPtr uintptr) string {
	if cstrPtr == 0 {
		return ""
	}
	ptr := (*byte)(unsafe.Pointer(cstrPtr))
	var result []byte
	for *ptr != 0 {
		result = append(result, *ptr)
		ptr = (*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + 1))
	}
	return string(result)
}

// createPipelines compiles MSL and creates all 4 render pipeline states.
func (b *Backend) createPipelines() error {
	// Compile library from MSL source
	srcNS := nsStringMetal(metalShaderSource)
	if srcNS == 0 {
		return fmt.Errorf("metal: failed to create NSString for shader source")
	}
	defer msgSend(srcNS, selReleaseMetal)

	var nsErr uintptr
	lib := msgSend(b.device, selNewDefaultLibraryWithSource, srcNS, 0, uintptr(unsafe.Pointer(&nsErr)))
	if lib == 0 {
		errStr := errFromNSError(nsErr)
		return fmt.Errorf("metal: shader compilation failed: %s", errStr)
	}
	defer msgSend(lib, selReleaseMetal)

	// Helper: get function from library
	getFunc := func(name string) (uintptr, error) {
		nameNS := nsStringMetal(name)
		if nameNS == 0 {
			return 0, fmt.Errorf("metal: failed to create NSString for function name %q", name)
		}
		defer msgSend(nameNS, selReleaseMetal)
		fn := msgSend(lib, selNewFunctionWithName, nameNS)
		if fn == 0 {
			return 0, fmt.Errorf("metal: function %q not found in shader library", name)
		}
		return fn, nil
	}

	rectVS, err := getFunc("rectVertex")
	if err != nil {
		return err
	}
	defer msgSend(rectVS, selReleaseMetal)

	rectFS, err := getFunc("rectFragment")
	if err != nil {
		return err
	}
	defer msgSend(rectFS, selReleaseMetal)

	shadowVS, err := getFunc("shadowVertex")
	if err != nil {
		return err
	}
	defer msgSend(shadowVS, selReleaseMetal)

	shadowFS, err := getFunc("shadowFragment")
	if err != nil {
		return err
	}
	defer msgSend(shadowFS, selReleaseMetal)

	texturedVS, err := getFunc("texturedVertex")
	if err != nil {
		return err
	}
	defer msgSend(texturedVS, selReleaseMetal)

	textFS, err := getFunc("textFragment")
	if err != nil {
		return err
	}
	defer msgSend(textFS, selReleaseMetal)

	imageFS, err := getFunc("imageFragment")
	if err != nil {
		return err
	}
	defer msgSend(imageFS, selReleaseMetal)

	// Build rect pipeline
	b.rectPipeline, err = b.buildPipeline(rectVS, rectFS, b.rectVertexDescriptor())
	if err != nil {
		return fmt.Errorf("metal: rect pipeline: %w", err)
	}

	// Build shadow pipeline
	b.shadowPipeline, err = b.buildPipeline(shadowVS, shadowFS, b.shadowVertexDescriptor())
	if err != nil {
		return fmt.Errorf("metal: shadow pipeline: %w", err)
	}

	// Build text pipeline (texturedVertex + textFragment)
	b.textPipeline, err = b.buildPipeline(texturedVS, textFS, b.texturedVertexDescriptor())
	if err != nil {
		return fmt.Errorf("metal: text pipeline: %w", err)
	}

	// Build image pipeline (texturedVertex + imageFragment)
	b.imagePipeline, err = b.buildPipeline(texturedVS, imageFS, b.texturedVertexDescriptor())
	if err != nil {
		return fmt.Errorf("metal: image pipeline: %w", err)
	}

	return nil
}

// buildPipeline creates a MTLRenderPipelineState from vertex/fragment functions and a vertex descriptor.
func (b *Backend) buildPipeline(vertFn, fragFn, vertDesc uintptr) (uintptr, error) {
	// MTLRenderPipelineDescriptor: alloc + init
	pipelineDescClass := objcClass("MTLRenderPipelineDescriptor")
	if pipelineDescClass == 0 {
		return 0, fmt.Errorf("MTLRenderPipelineDescriptor class not found")
	}
	pDesc := msgSend(msgSend(pipelineDescClass, selAllocMetal), selInitMetal)
	if pDesc == 0 {
		return 0, fmt.Errorf("failed to create MTLRenderPipelineDescriptor")
	}
	defer msgSend(pDesc, selReleaseMetal)

	msgSend(pDesc, selSetVertexFunction, vertFn)
	msgSend(pDesc, selSetFragmentFunction, fragFn)

	if vertDesc != 0 {
		msgSend(pDesc, selSetVertexDescriptor, vertDesc)
		msgSend(vertDesc, selReleaseMetal)
	}

	// Configure color attachment 0
	colorAttachments := msgSend(pDesc, selColorAttachmentsDescriptor)
	att0 := msgSend(colorAttachments, selObjectAtIndexedSubscript, 0)

	msgSend(att0, selSetPixelFormatPipeline, uintptr(MTLPixelFormatBGRA8Unorm_sRGB))
	msgSend(att0, selSetBlendingEnabled, 1)

	// src = srcAlpha, dst = 1-srcAlpha  (standard alpha blend)
	msgSend(att0, selSetSourceRGBBlendFactor, uintptr(MTLBlendFactorSourceAlpha))
	msgSend(att0, selSetDestinationRGBBlendFactor, uintptr(MTLBlendFactorOneMinusSourceAlpha))
	msgSend(att0, selSetRGBBlendOperation, uintptr(MTLBlendOperationAdd))
	msgSend(att0, selSetSourceAlphaBlendFactor, uintptr(MTLBlendFactorOne))
	msgSend(att0, selSetDestinationAlphaBlendFactor, uintptr(MTLBlendFactorOneMinusSourceAlpha))
	msgSend(att0, selSetAlphaBlendOperation, uintptr(MTLBlendOperationAdd))
	msgSend(att0, selSetWriteMask, uintptr(MTLColorWriteMaskAll))

	// Create pipeline state
	var nsErr uintptr
	state := msgSend(b.device, selNewRenderPipelineStateWithDescriptor,
		pDesc, uintptr(unsafe.Pointer(&nsErr)))
	if state == 0 {
		errStr := errFromNSError(nsErr)
		return 0, fmt.Errorf("newRenderPipelineStateWithDescriptor failed: %s", errStr)
	}
	return state, nil
}

// ---- Vertex descriptor builders ----

// MTLVertexFormat constants
const (
	mtlVertexFormatFloat  = 28
	mtlVertexFormatFloat2 = 29
	mtlVertexFormatFloat3 = 30
	mtlVertexFormatFloat4 = 31
)

// MTLVertexStepFunction constants
const (
	mtlVertexStepFunctionPerVertex = 1
)

// rectVertexDescriptor builds a MTLVertexDescriptor for RectVertex (76 bytes).
// Layout:
//   0: pos     float2  @ offset 0
//   1: uv      float2  @ offset 8
//   2: color   float4  @ offset 16
//   3: rectSize float2 @ offset 32
//   4: radii   float4  @ offset 40
//   5: borderWidth float @ offset 56
//   6: borderColor float4 @ offset 60
func (b *Backend) rectVertexDescriptor() uintptr {
	vdClass := objcClass("MTLVertexDescriptor")
	var vd uintptr
	if vdClass != 0 {
		vd = msgSend(msgSend(vdClass, selAllocMetal), selInitMetal)
	}
	if vd == 0 {
		return 0
	}

	attrs := msgSend(vd, selAttributes)
	setAttr := func(idx, format, offset int) {
		a := msgSend(attrs, selObjectAtIndexedSubscript, uintptr(idx))
		msgSend(a, selSetFormat, uintptr(format))
		msgSend(a, selSetOffset, uintptr(offset))
		msgSend(a, selSetBufferIndex, 0)
	}

	setAttr(0, mtlVertexFormatFloat2, 0)  // pos
	setAttr(1, mtlVertexFormatFloat2, 8)  // uv
	setAttr(2, mtlVertexFormatFloat4, 16) // color
	setAttr(3, mtlVertexFormatFloat2, 32) // rectSize
	setAttr(4, mtlVertexFormatFloat4, 40) // radii
	setAttr(5, mtlVertexFormatFloat, 56)  // borderWidth
	setAttr(6, mtlVertexFormatFloat4, 60) // borderColor

	layouts := msgSend(vd, selLayouts)
	l0 := msgSend(layouts, selObjectAtIndexedSubscript, 0)
	msgSend(l0, selSetStride, uintptr(76)) // sizeof(RectVertex)
	msgSend(l0, selSetStepFunction, uintptr(mtlVertexStepFunctionPerVertex))
	msgSend(l0, selSetStepRate, uintptr(1))

	return vd
}

// shadowVertexDescriptor builds a MTLVertexDescriptor for ShadowVertex (60 bytes).
// Layout:
//   0: pos      float2  @ offset 0
//   1: uv       float2  @ offset 8
//   2: color    float4  @ offset 16
//   3: elemSize float2  @ offset 32
//   4: radii    float4  @ offset 40
//   5: blur     float   @ offset 56
func (b *Backend) shadowVertexDescriptor() uintptr {
	vdClass := objcClass("MTLVertexDescriptor")
	var vd uintptr
	if vdClass != 0 {
		vd = msgSend(msgSend(vdClass, selAllocMetal), selInitMetal)
	}
	if vd == 0 {
		return 0
	}

	attrs := msgSend(vd, selAttributes)
	setAttr := func(idx, format, offset int) {
		a := msgSend(attrs, selObjectAtIndexedSubscript, uintptr(idx))
		msgSend(a, selSetFormat, uintptr(format))
		msgSend(a, selSetOffset, uintptr(offset))
		msgSend(a, selSetBufferIndex, 0)
	}

	setAttr(0, mtlVertexFormatFloat2, 0)  // pos
	setAttr(1, mtlVertexFormatFloat2, 8)  // uv
	setAttr(2, mtlVertexFormatFloat4, 16) // color
	setAttr(3, mtlVertexFormatFloat2, 32) // elemSize
	setAttr(4, mtlVertexFormatFloat4, 40) // radii
	setAttr(5, mtlVertexFormatFloat, 56)  // blur

	layouts := msgSend(vd, selLayouts)
	l0 := msgSend(layouts, selObjectAtIndexedSubscript, 0)
	msgSend(l0, selSetStride, uintptr(60)) // sizeof(ShadowVertex)
	msgSend(l0, selSetStepFunction, uintptr(mtlVertexStepFunctionPerVertex))
	msgSend(l0, selSetStepRate, uintptr(1))

	return vd
}

// texturedVertexDescriptor builds a MTLVertexDescriptor for TexturedVertex (32 bytes).
// Layout:
//   0: pos   float2  @ offset 0
//   1: uv    float2  @ offset 8
//   2: color float4  @ offset 16
func (b *Backend) texturedVertexDescriptor() uintptr {
	vdClass := objcClass("MTLVertexDescriptor")
	var vd uintptr
	if vdClass != 0 {
		vd = msgSend(msgSend(vdClass, selAllocMetal), selInitMetal)
	}
	if vd == 0 {
		return 0
	}

	attrs := msgSend(vd, selAttributes)
	setAttr := func(idx, format, offset int) {
		a := msgSend(attrs, selObjectAtIndexedSubscript, uintptr(idx))
		msgSend(a, selSetFormat, uintptr(format))
		msgSend(a, selSetOffset, uintptr(offset))
		msgSend(a, selSetBufferIndex, 0)
	}

	setAttr(0, mtlVertexFormatFloat2, 0)  // pos
	setAttr(1, mtlVertexFormatFloat2, 8)  // uv
	setAttr(2, mtlVertexFormatFloat4, 16) // color

	layouts := msgSend(vd, selLayouts)
	l0 := msgSend(layouts, selObjectAtIndexedSubscript, 0)
	msgSend(l0, selSetStride, uintptr(32)) // sizeof(TexturedVertex)
	msgSend(l0, selSetStepFunction, uintptr(mtlVertexStepFunctionPerVertex))
	msgSend(l0, selSetStepRate, uintptr(1))

	return vd
}

// createSamplers creates linear and nearest MTLSamplerState objects.
func (b *Backend) createSamplers() error {
	makeSampler := func(filter uintptr) (uintptr, error) {
		descClass := objcClass("MTLSamplerDescriptor")
		var desc uintptr
		if descClass != 0 {
			desc = msgSend(msgSend(descClass, selAllocMetal), selInitMetal)
		}
		if desc == 0 {
			return 0, fmt.Errorf("metal: failed to create MTLSamplerDescriptor")
		}
		defer msgSend(desc, selReleaseMetal)

		msgSend(desc, selSetMinFilter, filter)
		msgSend(desc, selSetMagFilter, filter)
		msgSend(desc, selSetSAddressMode, uintptr(MTLSamplerAddressModeClampToEdge))
		msgSend(desc, selSetTAddressMode, uintptr(MTLSamplerAddressModeClampToEdge))

		s := msgSend(b.device, selNewSamplerStateWithDescriptor, desc)
		if s == 0 {
			return 0, fmt.Errorf("metal: newSamplerStateWithDescriptor returned nil")
		}
		return s, nil
	}

	var err error
	b.linearSampler, err = makeSampler(uintptr(MTLSamplerMinMagFilterLinear))
	if err != nil {
		return err
	}
	b.nearestSampler, err = makeSampler(uintptr(MTLSamplerMinMagFilterNearest))
	if err != nil {
		return err
	}
	return nil
}

// BeginFrame implements render.Backend.
// Acquires the next drawable, creates a command buffer and render encoder.
func (b *Backend) BeginFrame() {
	if b.metalLayer == 0 {
		fmt.Println("[METAL] BeginFrame: metalLayer is 0!")
		return
	}

	// Reset vertex ring buffer offset each frame
	b.vertexOffset = 0

	// Get next drawable
	b.drawable = msgSend(b.metalLayer, selNextDrawable)
	if b.drawable == 0 {
		fmt.Println("[METAL] BeginFrame: drawable is nil!")
		return
	}

	// Drawable texture
	drawableTex := msgSend(b.drawable, selTexture)

	// Create command buffer
	b.cmdBuffer = msgSend(b.cmdQueue, selCommandBuffer)
	if b.cmdBuffer == 0 {
		return
	}

	// Build render pass descriptor
	rpdClass := objcClass("MTLRenderPassDescriptor")
	var rpd uintptr
	if rpdClass != 0 {
		rpd = msgSend(msgSend(rpdClass, selAllocMetal), selInitMetal)
	}
	if rpd == 0 {
		// fallback to class method
		rpd = msgSend(b.metalLayer, selCurrentRenderPassDescriptor)
	}
	if rpd == 0 {
		return
	}
	defer msgSend(rpd, selReleaseMetal)

	// Set up color attachment 0
	colorAttachments := msgSend(rpd, selColorAttachments)
	att0 := msgSend(colorAttachments, selObjectAtIndexedSubscript, 0)

	msgSend(att0, selSetTexture, drawableTex)
	msgSend(att0, selSetLoadAction, uintptr(MTLLoadActionClear))
	msgSend(att0, selSetStoreAction, uintptr(MTLStoreActionStore))
	// Clear color
	if b.transparent {
		msgSendClearColor(att0, selSetClearColor, 0.0, 0.0, 0.0, 0.0)
	} else {
		msgSendClearColor(att0, selSetClearColor, 1.0, 1.0, 1.0, 1.0)
	}

	// Create render encoder
	b.encoder = msgSend(b.cmdBuffer, selRenderCommandEncoderWithDescriptor, rpd)
}

// EndFrame implements render.Backend (no-op; present happens in Submit).
func (b *Backend) EndFrame() {}

// Submit implements render.Backend.
// Processes all commands, then presents and commits.
func (b *Backend) Submit(buf *render.CommandBuffer) {
	if b.encoder == 0 || buf == nil {
		b.presentAndCommit()
		return
	}

	// Process main commands then overlays
	b.processCommands(buf.Commands())
	b.processCommands(buf.Overlays())

	b.presentAndCommit()
}

// processCommands renders a slice of render.Command objects.
func (b *Backend) processCommands(commands []render.Command) {
	for _, c := range commands {
		switch c.Type {
		case render.CmdClip:
			if c.Clip != nil {
				b.setScissor(c.Clip)
			}
		case render.CmdRect:
			if c.Rect != nil {
				b.drawRect(c.Rect, c.Opacity)
			}
		case render.CmdShadow:
			if c.Shadow != nil {
				b.drawShadow(c.Shadow, c.Opacity)
			}
		case render.CmdText:
			if c.Text != nil {
				b.drawText(c.Text, c.Opacity)
			}
		case render.CmdImage:
			if c.Image != nil {
				b.drawImage(c.Image, c.Opacity)
			}
		}
	}
}

// presentAndCommit ends encoding, presents the drawable, and commits the command buffer.
func (b *Backend) presentAndCommit() {
	if b.encoder != 0 {
		msgSend(b.encoder, selEndEncoding)
		b.encoder = 0
	}
	if b.cmdBuffer != 0 {
		if b.drawable != 0 {
			msgSend(b.cmdBuffer, selPresentDrawable, b.drawable)
			b.drawable = 0
		}
		msgSend(b.cmdBuffer, selCommit)
		b.cmdBuffer = 0
	}
}

// Resize implements render.Backend.
func (b *Backend) Resize(w, h int) {
	b.width = w
	b.height = h
	if b.metalLayer != 0 {
		msgSendCGSize(b.metalLayer, selSetDrawableSize, float64(w), float64(h))
	}
}

// ndcCoords converts logical pixel coordinates to Metal NDC (Y-up).
// logW = b.width / b.dpiScale, logH = b.height / b.dpiScale
func (b *Backend) ndcCoords(x, y float32) (ndcX, ndcY float32) {
	logW := float32(b.width) / b.dpiScale
	logH := float32(b.height) / b.dpiScale
	ndcX = (x/logW)*2 - 1
	ndcY = 1 - (y/logH)*2 // Metal Y-up
	return
}

// writeVertices copies vertex data into the ring buffer and returns the byte offset where it starts.
func (b *Backend) writeVertices(data []byte) int {
	start := b.vertexOffset
	needed := start + len(data)
	if needed > vertexBufSize {
		// Wrap — lose the frame's earlier data rather than panic
		b.vertexOffset = 0
		start = 0
	}
	dst := unsafe.Slice((*byte)(unsafe.Add(b.vertexPtr, start)), len(data))
	copy(dst, data)
	b.vertexOffset = start + len(data)
	return start
}

// drawRect writes 6 RectVertex quads and issues a draw call.
func (b *Backend) drawRect(r *render.RectCmd, opacity float32) {
	x := r.Bounds.X
	y := r.Bounds.Y
	w := r.Bounds.Width
	h := r.Bounds.Height

	logW := float32(b.width) / b.dpiScale
	logH := float32(b.height) / b.dpiScale

	pad := float32(1.0)
	qx := x - pad
	qy := y - pad
	qw := w + pad*2
	qh := h + pad*2

	// NDC (Metal Y-up)
	ndcX := (qx/logW)*2 - 1
	ndcY := 1 - (qy/logH)*2
	ndcW := (qw / logW) * 2
	ndcH := (qh / logH) * 2

	uvL := -pad / w
	uvT := -pad / h
	uvR := 1.0 + pad/w
	uvB := 1.0 + pad/h

	col := r.FillColor
	a := col.A * opacity
	s := b.dpiScale

	rv := RectVertex{
		RectW: w * s, RectH: h * s,
		RadiusTL:    r.Corners.TopLeft * s,
		RadiusTR:    r.Corners.TopRight * s,
		RadiusBR:    r.Corners.BottomRight * s,
		RadiusBL:    r.Corners.BottomLeft * s,
		BorderWidth: r.BorderWidth * s,
		BorderR:     r.BorderColor.R,
		BorderG:     r.BorderColor.G,
		BorderB:     r.BorderColor.B,
		BorderA:     r.BorderColor.A,
	}

	// 4 corners of the padded quad (Y-down logical, Y-up NDC)
	// Top-left in logical = bottom-left in NDC
	// v0: top-left (logical) → NDC (ndcX, ndcY)
	// v1: top-right (logical) → NDC (ndcX+ndcW, ndcY)
	// v2: bottom-right (logical) → NDC (ndcX+ndcW, ndcY-ndcH)
	// v3: bottom-left (logical) → NDC (ndcX, ndcY-ndcH)
	// UV follows logical top-down order.

	v := func(px, py, u, v float32) RectVertex {
		rv2 := rv
		rv2.PosX = px
		rv2.PosY = py
		rv2.U = u
		rv2.V = v
		rv2.ColorR = col.R
		rv2.ColorG = col.G
		rv2.ColorB = col.B
		rv2.ColorA = a
		return rv2
	}

	// Two triangles: (v0,v1,v2), (v0,v2,v3)
	vertices := [6]RectVertex{
		v(ndcX, ndcY, uvL, uvT),                // top-left
		v(ndcX+ndcW, ndcY, uvR, uvT),           // top-right
		v(ndcX+ndcW, ndcY-ndcH, uvR, uvB),      // bottom-right
		v(ndcX, ndcY, uvL, uvT),                // top-left
		v(ndcX+ndcW, ndcY-ndcH, uvR, uvB),      // bottom-right
		v(ndcX, ndcY-ndcH, uvL, uvB),           // bottom-left
	}

	data := unsafe.Slice((*byte)(unsafe.Pointer(&vertices[0])), 6*int(unsafe.Sizeof(RectVertex{})))
	offset := b.writeVertices(data)

	msgSend(b.encoder, selSetRenderPipelineState, b.rectPipeline)
	msgSend(b.encoder, selSetVertexBuffer, b.vertexBuf, uintptr(offset), 0)
	msgSend(b.encoder, selDrawPrimitives,
		uintptr(MTLPrimitiveTypeTriangle), 0, 6)
}

// drawShadow writes 6 ShadowVertex quads and issues a draw call.
func (b *Backend) drawShadow(sh *render.ShadowCmd, opacity float32) {
	spread := sh.SpreadRadius
	blur := sh.BlurRadius
	const pad = float32(1.0)

	elemW := sh.Bounds.Width + 2*spread
	elemH := sh.Bounds.Height + 2*spread
	if elemW < 1 {
		elemW = 1
	}
	if elemH < 1 {
		elemH = 1
	}

	expand := blur*2 + pad

	qx := sh.Bounds.X + sh.OffsetX - spread - expand
	qy := sh.Bounds.Y + sh.OffsetY - spread - expand
	qw := elemW + 2*expand
	qh := elemH + 2*expand

	logW := float32(b.width) / b.dpiScale
	logH := float32(b.height) / b.dpiScale

	// NDC (Metal Y-up)
	ndcX := (qx/logW)*2 - 1
	ndcY := 1 - (qy/logH)*2
	ndcW := (qw / logW) * 2
	ndcH := (qh / logH) * 2

	uvL := -expand / elemW
	uvT := -expand / elemH
	uvR := 1.0 + expand/elemW
	uvB := 1.0 + expand/elemH

	s := b.dpiScale
	maxR := float32Min(elemW, elemH) * 0.5
	radTL := float32Clamp(sh.Corners.TopLeft+spread, 0, maxR) * s
	radTR := float32Clamp(sh.Corners.TopRight+spread, 0, maxR) * s
	radBR := float32Clamp(sh.Corners.BottomRight+spread, 0, maxR) * s
	radBL := float32Clamp(sh.Corners.BottomLeft+spread, 0, maxR) * s

	col := sh.Color
	a := col.A * opacity

	sv := ShadowVertex{
		ElemW:    elemW * s,
		ElemH:    elemH * s,
		RadiusTL: radTL,
		RadiusTR: radTR,
		RadiusBR: radBR,
		RadiusBL: radBL,
		Blur:     blur * s,
		ColorR:   col.R,
		ColorG:   col.G,
		ColorB:   col.B,
		ColorA:   a,
	}

	vf := func(px, py, u, v float32) ShadowVertex {
		sv2 := sv
		sv2.PosX, sv2.PosY, sv2.U, sv2.V = px, py, u, v
		return sv2
	}

	vertices := [6]ShadowVertex{
		vf(ndcX, ndcY, uvL, uvT),
		vf(ndcX+ndcW, ndcY, uvR, uvT),
		vf(ndcX+ndcW, ndcY-ndcH, uvR, uvB),
		vf(ndcX, ndcY, uvL, uvT),
		vf(ndcX+ndcW, ndcY-ndcH, uvR, uvB),
		vf(ndcX, ndcY-ndcH, uvL, uvB),
	}

	data := unsafe.Slice((*byte)(unsafe.Pointer(&vertices[0])), 6*int(unsafe.Sizeof(ShadowVertex{})))
	offset := b.writeVertices(data)

	msgSend(b.encoder, selSetRenderPipelineState, b.shadowPipeline)
	msgSend(b.encoder, selSetVertexBuffer, b.vertexBuf, uintptr(offset), 0)
	msgSend(b.encoder, selDrawPrimitives,
		uintptr(MTLPrimitiveTypeTriangle), 0, 6)
}

// drawText renders glyphs from a text command using the SDF text pipeline.
func (b *Backend) drawText(t *render.TextCmd, opacity float32) {
	if len(t.Glyphs) == 0 {
		return
	}

	b.texMu.RLock()
	mt, ok := b.textures[t.Atlas]
	b.texMu.RUnlock()
	if !ok || mt == nil {
		return
	}

	logW := float32(b.width) / b.dpiScale
	logH := float32(b.height) / b.dpiScale

	r, g, bl, a := t.Color.R, t.Color.G, t.Color.B, t.Color.A*opacity

	vertices := make([]TexturedVertex, 0, len(t.Glyphs)*6)
	for _, glyph := range t.Glyphs {
		// Metal Y-up NDC
		x0 := (glyph.X/logW)*2 - 1
		y0 := 1 - (glyph.Y/logH)*2
		x1 := ((glyph.X + glyph.Width) / logW) * 2 - 1
		y1 := 1 - ((glyph.Y+glyph.Height)/logH)*2

		vertices = append(vertices,
			TexturedVertex{x0, y0, glyph.U0, glyph.V0, r, g, bl, a},
			TexturedVertex{x1, y0, glyph.U1, glyph.V0, r, g, bl, a},
			TexturedVertex{x1, y1, glyph.U1, glyph.V1, r, g, bl, a},
			TexturedVertex{x0, y0, glyph.U0, glyph.V0, r, g, bl, a},
			TexturedVertex{x1, y1, glyph.U1, glyph.V1, r, g, bl, a},
			TexturedVertex{x0, y1, glyph.U0, glyph.V1, r, g, bl, a},
		)
	}

	data := unsafe.Slice((*byte)(unsafe.Pointer(&vertices[0])), len(vertices)*int(unsafe.Sizeof(TexturedVertex{})))
	offset := b.writeVertices(data)

	// Select sampler based on texture filter
	smp := b.nearestSampler
	if mt.filter == render.TextureFilterLinear {
		smp = b.linearSampler
	}

	msgSend(b.encoder, selSetRenderPipelineState, b.textPipeline)
	msgSend(b.encoder, selSetVertexBuffer, b.vertexBuf, uintptr(offset), 0)
	msgSend(b.encoder, selSetFragmentTexture, mt.tex, 0)
	msgSend(b.encoder, selSetFragmentSamplerState, smp, 0)
	msgSend(b.encoder, selDrawPrimitives,
		uintptr(MTLPrimitiveTypeTriangle), 0, uintptr(len(vertices)))
}

// drawImage renders a textured image command.
func (b *Backend) drawImage(img *render.ImageCmd, opacity float32) {
	b.texMu.RLock()
	mt, ok := b.textures[img.Texture]
	b.texMu.RUnlock()
	if !ok || mt == nil {
		return
	}

	logW := float32(b.width) / b.dpiScale
	logH := float32(b.height) / b.dpiScale

	// Metal Y-up NDC
	x0 := (img.DstRect.X/logW)*2 - 1
	y0 := 1 - (img.DstRect.Y/logH)*2
	x1 := ((img.DstRect.X + img.DstRect.Width) / logW) * 2 - 1
	y1 := 1 - ((img.DstRect.Y+img.DstRect.Height)/logH)*2

	// Default SrcRect to full texture when empty
	srcRect := img.SrcRect
	if srcRect.Width == 0 || srcRect.Height == 0 {
		srcRect.X, srcRect.Y, srcRect.Width, srcRect.Height = 0, 0, 1, 1
	}
	u0 := srcRect.X
	v0 := srcRect.Y
	u1 := srcRect.X + srcRect.Width
	v1 := srcRect.Y + srcRect.Height

	r, g, bl, a := img.Tint.R, img.Tint.G, img.Tint.B, img.Tint.A*opacity

	vertices := [6]TexturedVertex{
		{x0, y0, u0, v0, r, g, bl, a},
		{x1, y0, u1, v0, r, g, bl, a},
		{x1, y1, u1, v1, r, g, bl, a},
		{x0, y0, u0, v0, r, g, bl, a},
		{x1, y1, u1, v1, r, g, bl, a},
		{x0, y1, u0, v1, r, g, bl, a},
	}

	data := unsafe.Slice((*byte)(unsafe.Pointer(&vertices[0])), 6*int(unsafe.Sizeof(TexturedVertex{})))
	offset := b.writeVertices(data)

	smp := b.linearSampler
	if mt.filter == render.TextureFilterNearest {
		smp = b.nearestSampler
	}

	msgSend(b.encoder, selSetRenderPipelineState, b.imagePipeline)
	msgSend(b.encoder, selSetVertexBuffer, b.vertexBuf, uintptr(offset), 0)
	msgSend(b.encoder, selSetFragmentTexture, mt.tex, 0)
	msgSend(b.encoder, selSetFragmentSamplerState, smp, 0)
	msgSend(b.encoder, selDrawPrimitives,
		uintptr(MTLPrimitiveTypeTriangle), 0, 6)
}

// setScissor sets a scissor rect on the encoder.
// MTLScissorRect is {x, y, width, height uint64} — pass as 4 uintptr args.
// Clip bounds are logical pixels; convert to physical pixels.
func (b *Backend) setScissor(clip *render.ClipCmd) {
	if b.encoder == 0 {
		return
	}

	logW := float32(b.width) / b.dpiScale
	logH := float32(b.height) / b.dpiScale

	sx := int(clip.Bounds.X / logW * float32(b.width))
	sy := int(clip.Bounds.Y / logH * float32(b.height))
	sw := int(clip.Bounds.Width / logW * float32(b.width))
	sh := int(clip.Bounds.Height / logH * float32(b.height))

	// Clamp to framebuffer bounds
	if sx < 0 {
		sw += sx
		sx = 0
	}
	if sy < 0 {
		sh += sy
		sy = 0
	}
	if sw < 0 {
		sw = 0
	}
	if sh < 0 {
		sh = 0
	}
	if sx+sw > b.width {
		sw = b.width - sx
	}
	if sy+sh > b.height {
		sh = b.height - sy
	}

	// MTLScissorRect: {x, y, width, height} — 32 bytes struct
	// Must use typed wrapper because struct > 16 bytes on amd64.
	scissor := MTLScissorRect{
		X:      uintptr(sx),
		Y:      uintptr(sy),
		Width:  uintptr(sw),
		Height: uintptr(sh),
	}
	msgSendSetScissorRect(b.encoder, selSetScissorRect, scissor)
}

// CreateTexture implements render.Backend.
func (b *Backend) CreateTexture(desc render.TextureDesc) (render.TextureHandle, error) {
	if b.device == 0 {
		return render.InvalidTexture, fmt.Errorf("metal: device not initialized")
	}

	// Determine MTLPixelFormat
	var pixelFormat uintptr
	switch desc.Format {
	case render.TextureFormatR8:
		pixelFormat = uintptr(MTLPixelFormatR8Unorm)
	case render.TextureFormatRGBA8:
		pixelFormat = uintptr(MTLPixelFormatRGBA8Unorm)
	case render.TextureFormatBGRA8:
		pixelFormat = uintptr(MTLPixelFormatBGRA8Unorm)
	default:
		pixelFormat = uintptr(MTLPixelFormatRGBA8Unorm)
	}

	// [MTLTextureDescriptor texture2DDescriptorWithPixelFormat:width:height:mipmapped:]
	mtlTexDescClass := objcClass("MTLTextureDescriptor")
	if mtlTexDescClass == 0 {
		return render.InvalidTexture, fmt.Errorf("metal: MTLTextureDescriptor class not found")
	}
	texDesc := msgSend(mtlTexDescClass, selTexture2DDescriptorWithPixelFormat,
		pixelFormat, uintptr(desc.Width), uintptr(desc.Height), 0 /*mipmapped=NO*/)
	if texDesc == 0 {
		return render.InvalidTexture, fmt.Errorf("metal: texture2DDescriptorWithPixelFormat returned nil")
	}
	defer msgSend(texDesc, selReleaseMetal)

	msgSend(texDesc, selSetUsage, uintptr(MTLTextureUsageShaderRead))
	// MTLStorageModeShared = 0
	msgSend(texDesc, selSetStorageMode, 0)

	// Create texture
	tex := msgSend(b.device, selNewTextureWithDescriptor, texDesc)
	if tex == 0 {
		return render.InvalidTexture, fmt.Errorf("metal: newTextureWithDescriptor returned nil")
	}

	// Upload initial data if provided
	if len(desc.Data) > 0 {
		bytesPerRow := bytesPerRowForFormat(desc.Format, desc.Width)
		region := MTLRegion{
			Origin: MTLOrigin{X: 0, Y: 0, Z: 0},
			Size:   MTLSize{Width: uintptr(desc.Width), Height: uintptr(desc.Height), Depth: 1},
		}
		// replaceRegion:mipmapLevel:withBytes:bytesPerRow:
		// MTLRegion (48 bytes) MUST be passed via typed wrapper — SyscallN would
		// break the ABI by splitting struct fields across registers instead of
		// laying them contiguously on the stack (System V AMD64 ABI requirement
		// for structs > 16 bytes).
		msgSendReplaceRegion(tex, selReplaceRegion, region,
			0, // mipmapLevel
			uintptr(unsafe.Pointer(&desc.Data[0])),
			uintptr(bytesPerRow),
		)
	}

	b.texMu.Lock()
	handle := b.nextTexID
	b.nextTexID++
	b.textures[handle] = &metalTexture{
		tex:    tex,
		width:  desc.Width,
		height: desc.Height,
		format: desc.Format,
		filter: desc.Filter,
	}
	b.texMu.Unlock()

	return handle, nil
}

// UpdateTexture implements render.Backend.
func (b *Backend) UpdateTexture(handle render.TextureHandle, region uimath.Rect, data []byte) error {
	b.texMu.RLock()
	mt, ok := b.textures[handle]
	b.texMu.RUnlock()
	if !ok || mt == nil {
		return fmt.Errorf("metal: invalid texture handle %d", handle)
	}

	rx := int(region.X)
	ry := int(region.Y)
	rw := int(region.Width)
	rh := int(region.Height)

	if len(data) == 0 || rw <= 0 || rh <= 0 {
		return nil
	}

	bytesPerRow := bytesPerRowForFormat(mt.format, rw)

	mtlRegion := MTLRegion{
		Origin: MTLOrigin{X: uintptr(rx), Y: uintptr(ry), Z: 0},
		Size:   MTLSize{Width: uintptr(rw), Height: uintptr(rh), Depth: 1},
	}
	// replaceRegion:mipmapLevel:withBytes:bytesPerRow:
	msgSendReplaceRegion(mt.tex, selReplaceRegion, mtlRegion,
		0, // mipmapLevel
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(bytesPerRow),
	)
	return nil
}

// DestroyTexture implements render.Backend.
func (b *Backend) DestroyTexture(handle render.TextureHandle) {
	b.texMu.Lock()
	mt, ok := b.textures[handle]
	if ok {
		delete(b.textures, handle)
	}
	b.texMu.Unlock()
	if !ok || mt == nil {
		return
	}
	if mt.tex != 0 {
		msgSend(mt.tex, selReleaseMetal)
	}
}

// MaxTextureSize implements render.Backend.
func (b *Backend) MaxTextureSize() int {
	return 16384
}

// DPIScale implements render.Backend.
func (b *Backend) DPIScale() float32 {
	return b.dpiScale
}

// ReadPixels implements render.Backend. Returns nil (not implemented).
func (b *Backend) ReadPixels() (*image.RGBA, error) {
	return nil, nil
}

// Destroy implements render.Backend.
func (b *Backend) Destroy() {
	// Release all textures
	b.texMu.Lock()
	for handle, mt := range b.textures {
		if mt != nil && mt.tex != 0 {
			msgSend(mt.tex, selReleaseMetal)
		}
		delete(b.textures, handle)
	}
	b.texMu.Unlock()

	// Release samplers
	if b.linearSampler != 0 {
		msgSend(b.linearSampler, selReleaseMetal)
		b.linearSampler = 0
	}
	if b.nearestSampler != 0 {
		msgSend(b.nearestSampler, selReleaseMetal)
		b.nearestSampler = 0
	}

	// Release pipelines
	for _, pipe := range []uintptr{b.rectPipeline, b.shadowPipeline, b.textPipeline, b.imagePipeline} {
		if pipe != 0 {
			msgSend(pipe, selReleaseMetal)
		}
	}
	b.rectPipeline = 0
	b.shadowPipeline = 0
	b.textPipeline = 0
	b.imagePipeline = 0

	// Release vertex buffer
	if b.vertexBuf != 0 {
		msgSend(b.vertexBuf, selReleaseMetal)
		b.vertexBuf = 0
	}
	b.vertexPtr = nil

	// Release command queue
	if b.cmdQueue != 0 {
		msgSend(b.cmdQueue, selReleaseMetal)
		b.cmdQueue = 0
	}

	// Release device (device is retained by CAMetalLayer, release our reference)
	if b.device != 0 {
		msgSend(b.device, selReleaseMetal)
		b.device = 0
	}

	b.metalLayer = 0
}

// ---- Helpers ----

func bytesPerRowForFormat(format render.TextureFormat, width int) int {
	switch format {
	case render.TextureFormatR8:
		return width
	case render.TextureFormatRGBA8, render.TextureFormatBGRA8:
		return width * 4
	default:
		return width * 4
	}
}

func float32Min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func float32Clamp(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// Compile-time interface check.
var _ render.Backend = (*Backend)(nil)

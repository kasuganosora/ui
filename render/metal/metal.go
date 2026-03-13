//go:build darwin

// Package metal implements the render.Backend interface using Apple Metal via purego.
// Zero CGO — all Metal and ObjC calls go through purego.SyscallN or purego.RegisterFunc.
package metal

import (
	"unsafe"

	"github.com/ebitengine/purego"
)

// id is an Objective-C object pointer (alias for uintptr).
type id = uintptr

// ---- Metal struct types (matching C layout) ----

// MTLOrigin matches the C struct: {uint64, uint64, uint64}.
type MTLOrigin struct {
	X, Y, Z uintptr
}

// MTLSize matches the C struct: {uint64, uint64, uint64}.
type MTLSize struct {
	Width, Height, Depth uintptr
}

// MTLRegion matches the C struct: {MTLOrigin, MTLSize} — 48 bytes on 64-bit.
// This is a large struct (> 16 bytes) so it MUST be passed via purego.RegisterFunc
// typed wrappers, not through raw SyscallN (which would break the System V AMD64 ABI
// by placing struct fields in registers instead of on the stack).
type MTLRegion struct {
	Origin MTLOrigin
	Size   MTLSize
}

// MTLScissorRect matches the C struct: {uint64, uint64, uint64, uint64} — 32 bytes.
// Also > 16 bytes, requires typed wrapper on amd64.
type MTLScissorRect struct {
	X, Y, Width, Height uintptr
}

// ---- Library handles ----
var (
	objc_msgSend              uintptr
	objc_getClass             uintptr
	sel_registerName          uintptr
	MTLCreateSystemDefaultDevice uintptr
)

// ---- Typed function registrations for float/struct arguments ----
// These are needed on ARM64 where float args go in FP registers,
// not integer registers, so plain SyscallN won't work.
// They are ALSO needed on AMD64 for methods with struct parameters > 16 bytes,
// because SyscallN treats each field as a separate integer arg, but the C ABI
// requires large structs to be laid out contiguously on the stack.
var (
	// msgSendCGSize: setDrawableSize: (w, h float64)
	msgSendCGSize func(obj, sel uintptr, w, h float64)
	// msgSendCGFloat: setContentsScale: (f float64)
	msgSendCGFloat func(obj, sel uintptr, f float64)
	// msgSendClearColor: setClearColor: (r, g, b, a float64)
	msgSendClearColor func(obj, sel uintptr, r, g, b, a float64)
	// msgSendID: -> id (no extra args, returns id)
	msgSendID func(obj, sel uintptr) uintptr
	// msgSendBool: bool arg, returns id
	msgSendBool func(obj, sel uintptr, v uintptr) uintptr

	// msgSendReplaceRegion: replaceRegion:mipmapLevel:withBytes:bytesPerRow:
	// MTLRegion (48 bytes) is passed as a struct, then mipmapLevel, bytes ptr, bytesPerRow.
	msgSendReplaceRegion func(obj, sel uintptr, region MTLRegion, level uintptr, bytes uintptr, bytesPerRow uintptr)

	// msgSendSetScissorRect: setScissorRect: (MTLScissorRect — 32 bytes struct)
	msgSendSetScissorRect func(obj, sel uintptr, rect MTLScissorRect)
)

// ---- Metal pixel format constants ----
const (
	MTLPixelFormatBGRA8Unorm      = 80
	MTLPixelFormatBGRA8Unorm_sRGB = 81
	MTLPixelFormatR8Unorm         = 10
	MTLPixelFormatRGBA8Unorm      = 70
	MTLPixelFormatRGBA8Unorm_sRGB = 71
)

// ---- Metal load/store action constants ----
const (
	MTLLoadActionDontCare = 0
	MTLLoadActionLoad     = 1
	MTLLoadActionClear    = 2

	MTLStoreActionDontCare            = 0
	MTLStoreActionStore               = 1
	MTLStoreActionMultisampleResolve  = 2
)

// ---- Metal primitive type constants ----
const (
	MTLPrimitiveTypePoint         = 0
	MTLPrimitiveTypeLine          = 1
	MTLPrimitiveTypeLineStrip     = 2
	MTLPrimitiveTypeTriangle      = 3
	MTLPrimitiveTypeTriangleStrip = 4
)

// ---- Metal buffer options ----
const (
	MTLResourceStorageModeShared  = 0 << 4
	MTLResourceStorageModeManaged = 1 << 4
	MTLResourceStorageModePrivate = 2 << 4
	MTLResourceCPUCacheModeDefaultCache = 0
)

// ---- Metal texture type ----
const (
	MTLTextureType2D = 2
)

// ---- Metal texture usage ----
const (
	MTLTextureUsageShaderRead  = 1
	MTLTextureUsageShaderWrite = 2
	MTLTextureUsageRenderTarget = 4
)

// ---- Metal sampler address mode ----
const (
	MTLSamplerAddressModeClampToEdge  = 0
	MTLSamplerAddressModeRepeat       = 2
	MTLSamplerAddressModeClampToZero  = 3
)

// ---- Metal sampler min/mag filter ----
const (
	MTLSamplerMinMagFilterNearest = 0
	MTLSamplerMinMagFilterLinear  = 1
)

// ---- Metal blend factors ----
const (
	MTLBlendFactorZero                  = 0
	MTLBlendFactorOne                   = 1
	MTLBlendFactorSourceColor           = 2
	MTLBlendFactorOneMinusSourceColor   = 3
	MTLBlendFactorSourceAlpha           = 4
	MTLBlendFactorOneMinusSourceAlpha   = 5
	MTLBlendFactorDestinationColor      = 6
	MTLBlendFactorOneMinusDestinationColor = 7
	MTLBlendFactorDestinationAlpha      = 8
	MTLBlendFactorOneMinusDestinationAlpha = 9
)

// ---- Metal blend operations ----
const (
	MTLBlendOperationAdd             = 0
	MTLBlendOperationSubtract        = 1
	MTLBlendOperationReverseSubtract = 2
	MTLBlendOperationMin             = 3
	MTLBlendOperationMax             = 4
)

// ---- Metal color write mask ----
const (
	MTLColorWriteMaskNone  = 0
	MTLColorWriteMaskRed   = 0x1
	MTLColorWriteMaskGreen = 0x2
	MTLColorWriteMaskBlue  = 0x4
	MTLColorWriteMaskAlpha = 0x8
	MTLColorWriteMaskAll   = 0xf
)

// ---- Selectors (filled in init) ----
var (
	selNextDrawable               uintptr
	selCurrentRenderPassDescriptor uintptr
	selCommandBuffer              uintptr
	selRenderCommandEncoderWithDescriptor uintptr
	selSetRenderPipelineState     uintptr
	selSetVertexBuffer            uintptr
	selDrawPrimitives             uintptr
	selEndEncoding                uintptr
	selPresentDrawable            uintptr
	selCommit                     uintptr
	selTexture                    uintptr
	selSetFragmentTexture         uintptr
	selSetFragmentSamplerState    uintptr
	selSetScissorRect             uintptr
	selSetViewport                uintptr

	// device
	selNewCommandQueueSel              uintptr
	selNewBufferWithLength             uintptr
	selContents                        uintptr
	selNewTextureWithDescriptor        uintptr
	selReplaceRegion                   uintptr
	selNewDefaultLibraryWithSource     uintptr
	selNewFunctionWithName             uintptr
	selNewRenderPipelineStateWithDescriptor uintptr
	selNewSamplerStateWithDescriptor   uintptr

	// CAMetalLayer
	selLayer                     uintptr
	selSetDevice                 uintptr
	selSetPixelFormat            uintptr
	selSetDrawableSize           uintptr
	selSetContentsScale          uintptr
	selSetFramebufferOnly        uintptr
	selSetOpaque                 uintptr

	// NSView
	selSetWantsLayer             uintptr
	selSetLayerSel               uintptr

	// MTLRenderPassDescriptor
	selRenderPassDescriptor      uintptr
	selColorAttachments          uintptr
	selObjectAtIndexedSubscript  uintptr
	selSetTexture                uintptr
	selSetLoadAction             uintptr
	selSetStoreAction            uintptr
	selSetClearColor             uintptr

	// MTLRenderPipelineDescriptor
	selNewRenderPipelineDescriptor uintptr
	selSetVertexFunction           uintptr
	selSetFragmentFunction         uintptr
	selColorAttachmentsDescriptor  uintptr // returns array
	selSetPixelFormatPipeline      uintptr
	selSetBlendingEnabled          uintptr
	selSetSourceRGBBlendFactor     uintptr
	selSetDestinationRGBBlendFactor uintptr
	selSetRGBBlendOperation        uintptr
	selSetSourceAlphaBlendFactor   uintptr
	selSetDestinationAlphaBlendFactor uintptr
	selSetAlphaBlendOperation      uintptr
	selSetWriteMask                uintptr

	// MTLVertexDescriptor
	selVertexDescriptor          uintptr
	selAttributes                uintptr
	selLayouts                   uintptr
	selSetFormat                 uintptr
	selSetOffset                 uintptr
	selSetBufferIndex            uintptr
	selSetStride                 uintptr
	selSetStepFunction           uintptr
	selSetStepRate               uintptr
	selSetVertexDescriptor       uintptr

	// MTLTextureDescriptor
	selTexture2DDescriptorWithPixelFormat uintptr
	selSetUsage                  uintptr
	selSetStorageMode            uintptr

	// MTLSamplerDescriptor
	selNewSamplerDescriptor      uintptr
	selSetMinFilter              uintptr
	selSetMagFilter              uintptr
	selSetSAddressMode           uintptr
	selSetTAddressMode           uintptr

	// NSError
	selLocalizedDescription      uintptr
	selUTF8String                uintptr

	// NSString
	selStringWithUTF8StringMetal uintptr

	// release
	selReleaseMetal              uintptr

	// alloc / init
	selAllocMetal                uintptr
	selInitMetal                 uintptr
)

// helper: call sel_registerName to register a selector string
func sel(name string) uintptr {
	b := append([]byte(name), 0)
	r, _, _ := purego.SyscallN(sel_registerName, uintptr(unsafe.Pointer(&b[0])))
	return r
}

// helper: call objc_getClass by name
func objcClass(name string) uintptr {
	b := append([]byte(name), 0)
	r, _, _ := purego.SyscallN(objc_getClass, uintptr(unsafe.Pointer(&b[0])))
	return r
}

// msgSend calls objc_msgSend with integer/pointer args only.
func msgSend(obj, selector uintptr, args ...uintptr) uintptr {
	allArgs := make([]uintptr, 0, 2+len(args))
	allArgs = append(allArgs, obj, selector)
	allArgs = append(allArgs, args...)
	r, _, _ := purego.SyscallN(objc_msgSend, allArgs...)
	return r
}

func init() {
	// ---- Load libobjc ----
	objcHandle, err := purego.Dlopen("/usr/lib/libobjc.A.dylib", purego.RTLD_LAZY|purego.RTLD_GLOBAL)
	if err != nil {
		// Not on macOS — silently skip (build tag should prevent this, but be safe)
		return
	}

	objc_msgSend, _ = purego.Dlsym(objcHandle, "objc_msgSend")
	objc_getClass, _ = purego.Dlsym(objcHandle, "objc_getClass")
	sel_registerName, _ = purego.Dlsym(objcHandle, "sel_registerName")

	if objc_msgSend == 0 {
		return
	}

	// ---- Register typed function wrappers (needed for float/struct ObjC args on ARM64) ----
	// Also needed on AMD64 for methods with struct params > 16 bytes (e.g. MTLRegion, MTLScissorRect).
	purego.RegisterFunc(&msgSendCGSize, objc_msgSend)
	purego.RegisterFunc(&msgSendCGFloat, objc_msgSend)
	purego.RegisterFunc(&msgSendClearColor, objc_msgSend)
	purego.RegisterFunc(&msgSendID, objc_msgSend)
	purego.RegisterFunc(&msgSendBool, objc_msgSend)
	purego.RegisterFunc(&msgSendReplaceRegion, objc_msgSend)
	purego.RegisterFunc(&msgSendSetScissorRect, objc_msgSend)

	// ---- Load Metal framework (optional: Metal may not be available on very old macOS) ----
	_, _ = purego.Dlopen("/System/Library/Frameworks/Metal.framework/Metal", purego.RTLD_LAZY|purego.RTLD_GLOBAL)

	// ---- Load QuartzCore framework (for CAMetalLayer) ----
	_, _ = purego.Dlopen("/System/Library/Frameworks/QuartzCore.framework/QuartzCore", purego.RTLD_LAZY|purego.RTLD_GLOBAL)

	// ---- Load MTLCreateSystemDefaultDevice (Metal entry point) ----
	metalHandle, err2 := purego.Dlopen("/System/Library/Frameworks/Metal.framework/Metal", purego.RTLD_LAZY|purego.RTLD_GLOBAL)
	if err2 == nil {
		MTLCreateSystemDefaultDevice, _ = purego.Dlsym(metalHandle, "MTLCreateSystemDefaultDevice")
	}

	// ---- Register all selectors ----
	selReleaseMetal = sel("release")
	selAllocMetal = sel("alloc")
	selInitMetal = sel("init")

	// CAMetalLayer
	selLayer = sel("layer")
	selSetDevice = sel("setDevice:")
	selSetPixelFormat = sel("setPixelFormat:")
	selSetDrawableSize = sel("setDrawableSize:")
	selSetContentsScale = sel("setContentsScale:")
	selSetFramebufferOnly = sel("setFramebufferOnly:")
	selSetOpaque = sel("setOpaque:")

	// NSView
	selSetWantsLayer = sel("setWantsLayer:")
	selSetLayerSel = sel("setLayer:")

	// Command queue / buffer / encoder
	selNewCommandQueueSel = sel("newCommandQueue")
	selCommandBuffer = sel("commandBuffer")
	selNextDrawable = sel("nextDrawable")
	selCurrentRenderPassDescriptor = sel("currentRenderPassDescriptor")
	selRenderCommandEncoderWithDescriptor = sel("renderCommandEncoderWithDescriptor:")
	selEndEncoding = sel("endEncoding")
	selPresentDrawable = sel("presentDrawable:")
	selCommit = sel("commit")

	// Encoder state
	selSetRenderPipelineState = sel("setRenderPipelineState:")
	selSetVertexBuffer = sel("setVertexBuffer:offset:atIndex:")
	selDrawPrimitives = sel("drawPrimitives:vertexStart:vertexCount:")
	selSetScissorRect = sel("setScissorRect:")
	selSetViewport = sel("setViewport:")
	selTexture = sel("texture")
	selSetFragmentTexture = sel("setFragmentTexture:atIndex:")
	selSetFragmentSamplerState = sel("setFragmentSamplerState:atIndex:")

	// MTLBuffer
	selNewBufferWithLength = sel("newBufferWithLength:options:")
	selContents = sel("contents")

	// Texture / descriptor
	selNewTextureWithDescriptor = sel("newTextureWithDescriptor:")
	selReplaceRegion = sel("replaceRegion:mipmapLevel:withBytes:bytesPerRow:")
	selTexture2DDescriptorWithPixelFormat = sel("texture2DDescriptorWithPixelFormat:width:height:mipmapped:")
	selSetUsage = sel("setUsage:")
	selSetStorageMode = sel("setStorageMode:")

	// MTLRenderPassDescriptor
	selRenderPassDescriptor = sel("renderPassDescriptor")
	selColorAttachments = sel("colorAttachments")
	selObjectAtIndexedSubscript = sel("objectAtIndexedSubscript:")
	selSetTexture = sel("setTexture:")
	selSetLoadAction = sel("setLoadAction:")
	selSetStoreAction = sel("setStoreAction:")
	selSetClearColor = sel("setClearColor:")

	// Shader library / pipeline
	selNewDefaultLibraryWithSource = sel("newLibraryWithSource:options:error:")
	selNewFunctionWithName = sel("newFunctionWithName:")
	selNewRenderPipelineStateWithDescriptor = sel("newRenderPipelineStateWithDescriptor:error:")
	selNewRenderPipelineDescriptor = sel("renderPipelineDescriptor") // [MTLRenderPipelineDescriptor new] or alloc/init
	selSetVertexFunction = sel("setVertexFunction:")
	selSetFragmentFunction = sel("setFragmentFunction:")
	selColorAttachmentsDescriptor = sel("colorAttachments")
	selSetPixelFormatPipeline = sel("setPixelFormat:")
	selSetBlendingEnabled = sel("setBlendingEnabled:")
	selSetSourceRGBBlendFactor = sel("setSourceRGBBlendFactor:")
	selSetDestinationRGBBlendFactor = sel("setDestinationRGBBlendFactor:")
	selSetRGBBlendOperation = sel("setRgbBlendOperation:")
	selSetSourceAlphaBlendFactor = sel("setSourceAlphaBlendFactor:")
	selSetDestinationAlphaBlendFactor = sel("setDestinationAlphaBlendFactor:")
	selSetAlphaBlendOperation = sel("setAlphaBlendOperation:")
	selSetWriteMask = sel("setWriteMask:")

	// Vertex descriptor
	selVertexDescriptor = sel("vertexDescriptor")
	selAttributes = sel("attributes")
	selLayouts = sel("layouts")
	selSetFormat = sel("setFormat:")
	selSetOffset = sel("setOffset:")
	selSetBufferIndex = sel("setBufferIndex:")
	selSetStride = sel("setStride:")
	selSetStepFunction = sel("setStepFunction:")
	selSetStepRate = sel("setStepRate:")
	selSetVertexDescriptor = sel("setVertexDescriptor:")

	// Sampler
	selNewSamplerDescriptor = sel("samplerDescriptor")
	selNewSamplerStateWithDescriptor = sel("newSamplerStateWithDescriptor:")
	selSetMinFilter = sel("setMinFilter:")
	selSetMagFilter = sel("setMagFilter:")
	selSetSAddressMode = sel("setSAddressMode:")
	selSetTAddressMode = sel("setTAddressMode:")

	// NSError / NSString
	selLocalizedDescription = sel("localizedDescription")
	selUTF8String = sel("UTF8String")
	selStringWithUTF8StringMetal = sel("stringWithUTF8String:")

}

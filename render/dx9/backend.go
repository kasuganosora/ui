//go:build windows

package dx9

import (
	"fmt"
	"image"
	"math"
	"syscall"
	"unsafe"

	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/render"
)

// RectVertex matches the rect shader input layout (vs_3_0/ps_3_0).
// Position (float2) + UV (float2) + Color (float4) + RectSize (float2) +
// Radius (float4) + BorderWidth (float1) + BorderColor (float4) = 19 floats = 76 bytes
type RectVertex struct {
	PosX, PosY                                 float32
	U, V                                       float32
	ColorR, ColorG, ColorB, ColorA             float32
	RectW, RectH                               float32
	RadiusTL, RadiusTR, RadiusBR, RadiusBL     float32
	BorderWidth                                float32
	BorderR, BorderG, BorderB, BorderA         float32
}

// ShadowVertex matches the shadow shader input layout — 15 float32, 60 bytes.
type ShadowVertex struct {
	PosX, PosY                              float32
	U, V                                    float32
	ColorR, ColorG, ColorB, ColorA          float32
	ElemW, ElemH                            float32
	RadiusTL, RadiusTR, RadiusBR, RadiusBL float32
	Blur                                    float32
}

// TexturedVertex for text and image rendering.
type TexturedVertex struct {
	PosX, PosY                     float32
	U, V                           float32
	ColorR, ColorG, ColorB, ColorA float32
}

// dx9Pipeline identifies the active rendering pipeline to avoid redundant state switches.
type dx9Pipeline uint8

const (
	pipelineNone     dx9Pipeline = iota
	pipelineRect                         // SDF rect shader
	pipelineTextured                     // image shader (sRGB texture)
	pipelineText                         // text shader (coverage atlas, no sRGB)
	pipelineShadow                       // shadow shader
)

type textureEntry struct {
	texture uintptr // IDirect3DTexture9*
	width   int
	height  int
	format  render.TextureFormat
}

// ---- IDirect3D9 vtable indices (IUnknown: 0-2) ----
const (
	d3d9RegisterSoftwareDevice = 3
	d3d9GetAdapterCount        = 4
	d3d9GetAdapterIdentifier   = 5
	d3d9GetAdapterModeCount    = 6
	d3d9EnumAdapterModes       = 7
	d3d9GetAdapterDisplayMode  = 8
	d3d9CheckDeviceType        = 9
	d3d9CheckDeviceFormat      = 10
	d3d9CheckDeviceMultiSampleType = 11
	d3d9CheckDepthStencilMatch = 12
	d3d9CheckDeviceFormatConversion = 13
	d3d9GetDeviceCaps          = 14
	d3d9GetAdapterMonitor      = 15
	d3d9CreateDevice           = 16
)

// ---- IDirect3DDevice9 vtable indices (IUnknown: 0-2) ----
const (
	devTestCooperativeLevel = 3
	devGetAvailableTextureMem = 4
	devEvictManagedResources = 5
	devGetDirect3D          = 6
	devGetDeviceCaps        = 7
	devGetDisplayMode       = 8
	devGetCreationParameters = 9
	devSetCursorProperties  = 10
	devSetCursorPosition    = 11
	devShowCursor           = 12
	devCreateAdditionalSwapChain = 13
	devGetSwapChain         = 14
	devGetNumberOfSwapChains = 15
	devReset                = 16
	devPresent              = 17
	devGetBackBuffer        = 18
	devGetRasterStatus      = 19
	devSetDialogBoxMode     = 20
	devSetGammaRamp         = 21
	devGetGammaRamp         = 22
	devCreateTexture        = 23
	devCreateVolumeTexture  = 24
	devCreateCubeTexture    = 25
	devCreateVertexBuffer   = 26
	devCreateIndexBuffer    = 27
	devCreateRenderTarget   = 28
	devCreateDepthStencilSurface = 29
	devUpdateSurface        = 30
	devUpdateTexture        = 31
	devGetRenderTargetData  = 32
	devGetFrontBufferData   = 33
	devStretchRect          = 34
	devColorFill            = 35
	devCreateOffscreenPlainSurface = 36
	devSetRenderTarget      = 37
	devGetRenderTarget      = 38
	devSetDepthStencilSurface = 39
	devGetDepthStencilSurface = 40
	devBeginScene           = 41
	devEndScene             = 42
	devClear                = 43
	devSetTransform         = 44
	devGetTransform         = 45
	devMultiplyTransform    = 46
	devSetViewport          = 47
	devGetViewport          = 48
	devSetMaterial          = 49
	devGetMaterial          = 50
	devSetLight             = 51
	devGetLight             = 52
	devLightEnable          = 53
	devGetLightEnable       = 54
	devSetClipPlane         = 55
	devGetClipPlane         = 56
	devSetRenderState       = 57
	devGetRenderState       = 58
	devCreateStateBlock     = 59
	devBeginStateBlock      = 60
	devEndStateBlock        = 61
	devSetClipStatus        = 62
	devGetClipStatus        = 63
	devGetTexture           = 64
	devSetTexture           = 65
	devGetTextureStageState = 66
	devSetTextureStageState = 67
	devGetSamplerState      = 68
	devSetSamplerState      = 69
	devValidateDevice       = 70
	devSetPaletteEntries    = 71
	devGetPaletteEntries    = 72
	devSetCurrentTexturePalette = 73
	devGetCurrentTexturePalette = 74
	devSetScissorRect       = 75
	devGetScissorRect       = 76
	devSetSoftwareVertexProcessing = 77
	devGetSoftwareVertexProcessing = 78
	devSetNPatchMode        = 79
	devGetNPatchMode        = 80
	devDrawPrimitive        = 81
	devDrawIndexedPrimitive = 82
	devDrawPrimitiveUP      = 83
	devDrawIndexedPrimitiveUP = 84
	devProcessVertices      = 85
	devCreateVertexDeclaration = 86
	devSetVertexDeclaration = 87
	devGetVertexDeclaration = 88
	devSetFVF               = 89
	devGetFVF               = 90
	devCreateVertexShaderFn = 91
	devSetVertexShader      = 92
	devGetVertexShader      = 93
	devSetVertexShaderConstantF = 94
	devGetVertexShaderConstantF = 95
	devSetVertexShaderConstantI = 96
	devGetVertexShaderConstantI = 97
	devSetVertexShaderConstantB = 98
	devGetVertexShaderConstantB = 99
	devSetStreamSource      = 100
	devGetStreamSource      = 101
	devSetStreamSourceFreq  = 102
	devGetStreamSourceFreq  = 103
	devSetIndices           = 104
	devGetIndices           = 105
	devCreatePixelShaderFn  = 106
	devSetPixelShader       = 107
	devGetPixelShader       = 108
	devSetPixelShaderConstantF = 109
	devGetPixelShaderConstantF = 110
	devSetPixelShaderConstantI = 111
	devGetPixelShaderConstantI = 112
	devSetPixelShaderConstantB = 113
	devGetPixelShaderConstantB = 114
	devDrawRectPatch        = 115
	devDrawTriPatch         = 116
	devDeletePatch          = 117
	devCreateQuery          = 118
)

// ---- IDirect3DTexture9 vtable indices ----
// IUnknown(0-2) + IDirect3DResource9(3-9) + IDirect3DBaseTexture9(10-17)
const (
	texGetLevelDesc  = 17
	texGetSurfaceLevel = 18
	texLockRect      = 19
	texUnlockRect    = 20
	texAddDirtyRect  = 21
)

// ---- IDirect3DSurface9 vtable indices ----
const (
	surfLockRect   = 13
	surfUnlockRect = 14
)

// Backend implements render.Backend using DirectX 9.
type Backend struct {
	loader *Loader

	d3d9   uintptr // IDirect3D9*
	device uintptr // IDirect3DDevice9*

	// Shaders (SM 3.0)
	rectVS     uintptr // IDirect3DVertexShader9*
	rectPS     uintptr // IDirect3DPixelShader9*
	texturedVS uintptr // IDirect3DVertexShader9*
	texturedPS uintptr // IDirect3DPixelShader9*
	textPS     uintptr // IDirect3DPixelShader9* (SDF text)
	shadowVS   uintptr // IDirect3DVertexShader9*
	shadowPS   uintptr // IDirect3DPixelShader9*

	// Vertex declarations
	rectDecl     uintptr // IDirect3DVertexDeclaration9*
	texturedDecl uintptr // IDirect3DVertexDeclaration9*
	shadowDecl   uintptr // IDirect3DVertexDeclaration9*

	// Dynamic vertex buffers
	rectVB         uintptr // IDirect3DVertexBuffer9*
	rectVBSize     uint32
	texturedVB     uintptr
	texturedVBSize uint32
	shadowVB       uintptr
	shadowVBSize   uint32

	// State
	width, height int
	dpiScale      float32
	hwnd          uintptr

	// Texture management
	nextTextureID render.TextureHandle
	textures      map[render.TextureHandle]*textureEntry

	// Present params (needed for Reset)
	pp D3DPRESENT_PARAMETERS

	// Pipeline state tracking (avoid redundant COM calls)
	activePipeline dx9Pipeline
	activeTexture  uintptr // currently bound IDirect3DTexture9*

	// Staging buffers (reused across frames to avoid GC allocations)
	rectStaging     []RectVertex
	texturedStaging []TexturedVertex
	shadowStaging   []ShadowVertex
}

// New creates a new DX9 backend.
func New() *Backend {
	return &Backend{
		nextTextureID: 1,
		textures:      make(map[render.TextureHandle]*textureEntry),
		dpiScale:      1.0,
	}
}

// comVtbl reads a single vtable entry from a COM object at the given index.
func comVtbl(ptr uintptr, index int) uintptr {
	vtblPtr := *(*uintptr)(unsafe.Pointer(ptr))
	return *(*uintptr)(unsafe.Pointer(vtblPtr + uintptr(index)*unsafe.Sizeof(uintptr(0))))
}

// comCall invokes a COM method via vtable entry.
func comCall(method uintptr, args ...uintptr) uintptr {
	ret, _, _ := syscall.SyscallN(method, args...)
	return ret
}

// comRelease calls IUnknown::Release (vtable index 2).
func comRelease(ptr uintptr) {
	if ptr == 0 {
		return
	}
	comCall(comVtbl(ptr, 2), ptr)
}

// Init implements render.Backend.
func (b *Backend) Init(window platform.Window) error {
	var err error
	b.loader, err = NewLoader()
	if err != nil {
		return err
	}

	b.hwnd = window.NativeHandle()
	b.width, b.height = window.FramebufferSize()
	b.dpiScale = window.DPIScale()
	if b.dpiScale <= 0 {
		b.dpiScale = 1.0
	}

	// Create IDirect3D9
	b.d3d9, _, _ = b.loader.direct3DCreate9.Call(D3D_SDK_VERSION)
	if b.d3d9 == 0 {
		return fmt.Errorf("dx9: Direct3DCreate9 failed")
	}

	// Create device
	b.pp = D3DPRESENT_PARAMETERS{
		BackBufferWidth:  uint32(b.width),
		BackBufferHeight: uint32(b.height),
		BackBufferFormat: D3DFMT_A8R8G8B8,
		BackBufferCount:  1,
		MultiSampleType:  D3DMULTISAMPLE_NONE,
		SwapEffect:       D3DSWAPEFFECT_DISCARD,
		HDeviceWindow:    b.hwnd,
		Windowed:         1,
		PresentationInterval: 0, // D3DPRESENT_INTERVAL_IMMEDIATE
	}

	hr := comCall(comVtbl(b.d3d9, d3d9CreateDevice),
		b.d3d9,
		0, // Adapter (D3DADAPTER_DEFAULT)
		D3DDEVTYPE_HAL,
		b.hwnd,
		D3DCREATE_HARDWARE_VERTEXPROCESSING|D3DCREATE_FPU_PRESERVE,
		uintptr(unsafe.Pointer(&b.pp)),
		uintptr(unsafe.Pointer(&b.device)))
	if hr != S_OK {
		// Fall back to software vertex processing
		hr = comCall(comVtbl(b.d3d9, d3d9CreateDevice),
			b.d3d9,
			0,
			D3DDEVTYPE_HAL,
			b.hwnd,
			D3DCREATE_SOFTWARE_VERTEXPROCESSING|D3DCREATE_FPU_PRESERVE,
			uintptr(unsafe.Pointer(&b.pp)),
			uintptr(unsafe.Pointer(&b.device)))
		if hr != S_OK {
			return fmt.Errorf("dx9: CreateDevice failed: 0x%x", hr)
		}
	}

	// Set up render states
	b.setRenderStates()

	// Create shaders
	if err := b.createShaders(); err != nil {
		return err
	}

	// Create initial vertex buffers
	b.rectVBSize = 65536
	b.texturedVBSize = 65536
	if err := b.createVertexBuffers(); err != nil {
		return err
	}

	return nil
}

func (b *Backend) setRenderStates() {
	dev := b.device

	// Disable Z-buffer and lighting
	comCall(comVtbl(dev, devSetRenderState), dev, D3DRS_ZENABLE, 0)
	comCall(comVtbl(dev, devSetRenderState), dev, D3DRS_LIGHTING, 0)
	comCall(comVtbl(dev, devSetRenderState), dev, D3DRS_CULLMODE, D3DCULL_NONE)

	// Alpha blending: SrcAlpha, InvSrcAlpha
	comCall(comVtbl(dev, devSetRenderState), dev, D3DRS_ALPHABLENDENABLE, 1)
	comCall(comVtbl(dev, devSetRenderState), dev, D3DRS_SRCBLEND, D3DBLEND_SRCALPHA)
	comCall(comVtbl(dev, devSetRenderState), dev, D3DRS_DESTBLEND, D3DBLEND_INVSRCALPHA)
	comCall(comVtbl(dev, devSetRenderState), dev, D3DRS_BLENDOP, D3DBLENDOP_ADD)

	// Separate alpha blending
	comCall(comVtbl(dev, devSetRenderState), dev, D3DRS_SEPARATEALPHABLENDENABLE, 1)
	comCall(comVtbl(dev, devSetRenderState), dev, D3DRS_SRCBLENDALPHA, D3DBLEND_ONE)
	comCall(comVtbl(dev, devSetRenderState), dev, D3DRS_DESTBLENDALPHA, D3DBLEND_INVSRCALPHA)
	comCall(comVtbl(dev, devSetRenderState), dev, D3DRS_BLENDOPALPHA, D3DBLENDOP_ADD)

	// Enable scissor test
	comCall(comVtbl(dev, devSetRenderState), dev, D3DRS_SCISSORTESTENABLE, 1)

	// sRGB write for linear-space blending (if supported by driver)
	comCall(comVtbl(dev, devSetRenderState), dev, D3DRS_SRGBWRITEENABLE, 1)

	// Sampler state: linear filtering, clamp
	comCall(comVtbl(dev, devSetSamplerState), dev, 0, D3DSAMP_MINFILTER, D3DTEXF_LINEAR)
	comCall(comVtbl(dev, devSetSamplerState), dev, 0, D3DSAMP_MAGFILTER, D3DTEXF_LINEAR)
	comCall(comVtbl(dev, devSetSamplerState), dev, 0, D3DSAMP_MIPFILTER, D3DTEXF_NONE)
	comCall(comVtbl(dev, devSetSamplerState), dev, 0, D3DSAMP_ADDRESSU, D3DTADDRESS_CLAMP)
	comCall(comVtbl(dev, devSetSamplerState), dev, 0, D3DSAMP_ADDRESSV, D3DTADDRESS_CLAMP)
}

// D3DCompile from d3dcompiler_47.dll
var (
	d3dcompiler    = syscall.NewLazyDLL("d3dcompiler_47.dll")
	procD3DCompile = d3dcompiler.NewProc("D3DCompile")
)

func (b *Backend) createShaders() error {
	// ---- Rect vertex shader (SM 3.0) ----
	// Passes all attributes to pixel shader for SDF evaluation
	rectVSCode := `
struct VS_INPUT {
    float2 pos : POSITION;
    float2 uv : TEXCOORD0;
    float4 color : COLOR0;
    float2 rectSize : TEXCOORD1;
    float4 radius : TEXCOORD2;
    float borderWidth : TEXCOORD3;
    float4 borderColor : COLOR1;
};
struct PS_INPUT {
    float4 pos : POSITION;
    float2 uv : TEXCOORD0;
    float4 color : COLOR1;
    float2 rectSize : TEXCOORD1;
    float4 radius : TEXCOORD2;
    float borderWidth : TEXCOORD3;
    float4 borderColor : COLOR2;
};
float3 srgbToLinear(float3 c) {
    return pow(max(c, 0.0), 2.2);
}
PS_INPUT main(VS_INPUT input) {
    PS_INPUT output;
    output.pos = float4(input.pos, 0.0, 1.0);
    output.uv = input.uv;
    output.color = float4(srgbToLinear(input.color.rgb), input.color.a);
    output.rectSize = input.rectSize;
    output.radius = input.radius;
    output.borderWidth = input.borderWidth;
    output.borderColor = float4(srgbToLinear(input.borderColor.rgb), input.borderColor.a);
    return output;
}`

	rectPSCode := `
struct PS_INPUT {
    float4 pos : POSITION;
    float2 uv : TEXCOORD0;
    float4 color : COLOR1;
    float2 rectSize : TEXCOORD1;
    float4 radius : TEXCOORD2;
    float borderWidth : TEXCOORD3;
    float4 borderColor : COLOR2;
};
float roundedBoxSDF(float2 p, float2 b, float4 r) {
    float radius = (p.x > 0.0) ? ((p.y > 0.0) ? r.z : r.y) : ((p.y > 0.0) ? r.w : r.x);
    float2 q = abs(p) - b + float2(radius, radius);
    return min(max(q.x, q.y), 0.0) + length(max(q, 0.0)) - radius;
}
float4 main(PS_INPUT input) : COLOR0 {
    float2 p = (input.uv - 0.5) * input.rectSize;
    float2 b = input.rectSize * 0.5;
    float dist = roundedBoxSDF(p, b, input.radius);
    float aa = max(fwidth(dist), 0.5);
    float fillAlpha = 1.0 - smoothstep(0.0, aa, dist);
    if (input.borderWidth > 0.0) {
        float innerDist = dist + input.borderWidth;
        float fillMask = 1.0 - smoothstep(0.0, aa, innerDist);
        float4 color = lerp(input.borderColor, input.color, fillMask);
        return float4(color.rgb, color.a * fillAlpha);
    }
    return float4(input.color.rgb, input.color.a * fillAlpha);
}`

	// ---- Textured vertex shader (images) ----
	texturedVSCode := `
struct VS_INPUT {
    float2 pos : POSITION;
    float2 uv : TEXCOORD0;
    float4 color : COLOR0;
};
struct PS_INPUT {
    float4 pos : POSITION;
    float2 uv : TEXCOORD0;
    float4 color : COLOR1;
};
float3 srgbToLinear(float3 c) {
    return pow(max(c, 0.0), 2.2);
}
PS_INPUT main(VS_INPUT input) {
    PS_INPUT output;
    output.pos = float4(input.pos, 0.0, 1.0);
    output.uv = input.uv;
    output.color = float4(srgbToLinear(input.color.rgb), input.color.a);
    return output;
}`

	texturedPSCode := `
sampler2D tex : register(s0);
struct PS_INPUT {
    float4 pos : POSITION;
    float2 uv : TEXCOORD0;
    float4 color : COLOR1;
};
float4 main(PS_INPUT input) : COLOR0 {
    float4 texColor = tex2D(tex, input.uv);
    return texColor * input.color;
}`

	// ---- Text pixel shader (coverage from R channel) ----
	textPSCode := `
sampler2D tex : register(s0);
struct PS_INPUT {
    float4 pos : POSITION;
    float2 uv : TEXCOORD0;
    float4 color : COLOR1;
};
float4 main(PS_INPUT input) : COLOR0 {
    // Font atlas uses D3DFMT_L8: coverage in RGB (r=g=b=luminance), a=1.0
    float coverage = tex2D(tex, input.uv).r;
    return float4(input.color.rgb, input.color.a * coverage);
}`

	var err error

	// Compile rect shaders
	b.rectVS, err = b.compileVertexShader(rectVSCode)
	if err != nil {
		return fmt.Errorf("dx9: rect VS: %w", err)
	}
	b.rectPS, err = b.compilePixelShader(rectPSCode)
	if err != nil {
		return fmt.Errorf("dx9: rect PS: %w", err)
	}

	// Compile textured shaders
	b.texturedVS, err = b.compileVertexShader(texturedVSCode)
	if err != nil {
		return fmt.Errorf("dx9: textured VS: %w", err)
	}
	b.texturedPS, err = b.compilePixelShader(texturedPSCode)
	if err != nil {
		return fmt.Errorf("dx9: textured PS: %w", err)
	}
	b.textPS, err = b.compilePixelShader(textPSCode)
	if err != nil {
		return fmt.Errorf("dx9: text PS: %w", err)
	}

	// ---- Shadow vertex shader (SM 3.0) ----
	shadowVSCode := `
struct VS_INPUT {
    float2 pos      : POSITION;
    float2 uv       : TEXCOORD0;
    float4 color    : COLOR0;
    float2 elemSize : TEXCOORD1;
    float4 radii    : TEXCOORD2;
    float  blur     : TEXCOORD3;
};
struct PS_INPUT {
    float4 pos      : POSITION;
    float2 uv       : TEXCOORD0;
    float4 color    : COLOR1;
    float2 elemSize : TEXCOORD1;
    float4 radii    : TEXCOORD2;
    float  blur     : TEXCOORD3;
};
float3 srgbToLinear(float3 c) {
    return pow(max(c, 0.0), 2.2);
}
PS_INPUT main(VS_INPUT input) {
    PS_INPUT output;
    output.pos      = float4(input.pos, 0.0, 1.0);
    output.uv       = input.uv;
    output.color    = float4(srgbToLinear(input.color.rgb), input.color.a);
    output.elemSize = input.elemSize;
    output.radii    = input.radii;
    output.blur     = input.blur;
    return output;
}`

	shadowPSCode := `
struct PS_INPUT {
    float4 pos      : POSITION;
    float2 uv       : TEXCOORD0;
    float4 color    : COLOR1;
    float2 elemSize : TEXCOORD1;
    float4 radii    : TEXCOORD2;
    float  blur     : TEXCOORD3;
};
float roundedRectSDF(float2 p, float2 halfSize, float4 r) {
    float radius = (p.x > 0.0) ? ((p.y > 0.0) ? r.z : r.y)
                               : ((p.y > 0.0) ? r.w : r.x);
    float2 q = abs(p) - halfSize + float2(radius, radius);
    return min(max(q.x, q.y), 0.0) + length(max(q, 0.0)) - radius;
}
float4 main(PS_INPUT input) : COLOR0 {
    float2 elemHalf = input.elemSize * 0.5;
    float2 p        = (input.uv - 0.5) * input.elemSize;
    float  dist     = roundedRectSDF(p, elemHalf, input.radii);
    float  aa       = 1.0;
    float  alpha;
    float  blur     = input.blur;
    if (blur < 1.0) {
        alpha = 1.0 - smoothstep(-aa, aa, dist);
    } else {
        float sigma = blur * 0.5;
        float t     = max(0.0, dist) / sigma;
        alpha       = exp(-t * t * 0.5);
        // Fade out at 3-sigma from the element boundary.
        alpha      *= 1.0 - smoothstep(-aa, aa, dist - sigma * 3.0);
    }
    return float4(input.color.rgb, input.color.a * alpha);
}`

	// Compile shadow shaders
	b.shadowVS, err = b.compileVertexShader(shadowVSCode)
	if err != nil {
		return fmt.Errorf("dx9: shadow VS: %w", err)
	}
	b.shadowPS, err = b.compilePixelShader(shadowPSCode)
	if err != nil {
		return fmt.Errorf("dx9: shadow PS: %w", err)
	}

	// Create vertex declarations
	if err := b.createVertexDeclarations(); err != nil {
		return err
	}

	return nil
}

func (b *Backend) createVertexDeclarations() error {
	// Rect vertex declaration
	// Position(float2) + UV(float2) + Color(float4) + RectSize(float2) +
	// Radius(float4) + BorderWidth(float1→pad to float2) + BorderColor(float4)
	rectElements := []D3DVERTEXELEMENT9{
		{0, 0, D3DDECLTYPE_FLOAT2, 0, D3DDECLUSAGE_POSITION, 0},
		{0, 8, D3DDECLTYPE_FLOAT2, 0, D3DDECLUSAGE_TEXCOORD, 0},
		{0, 16, D3DDECLTYPE_FLOAT4, 0, D3DDECLUSAGE_COLOR, 0},
		{0, 32, D3DDECLTYPE_FLOAT2, 0, D3DDECLUSAGE_TEXCOORD, 1},
		{0, 40, D3DDECLTYPE_FLOAT4, 0, D3DDECLUSAGE_TEXCOORD, 2},
		{0, 56, 0, 0, D3DDECLUSAGE_TEXCOORD, 3}, // float1 = D3DDECLTYPE_FLOAT1 = 0
		{0, 60, D3DDECLTYPE_FLOAT4, 0, D3DDECLUSAGE_COLOR, 1},
		D3DVERTEXELEMENT9_END,
	}

	hr := comCall(comVtbl(b.device, devCreateVertexDeclaration),
		b.device,
		uintptr(unsafe.Pointer(&rectElements[0])),
		uintptr(unsafe.Pointer(&b.rectDecl)))
	if hr != S_OK {
		return fmt.Errorf("dx9: CreateVertexDeclaration (rect) failed: 0x%x", hr)
	}

	// Textured vertex declaration
	texturedElements := []D3DVERTEXELEMENT9{
		{0, 0, D3DDECLTYPE_FLOAT2, 0, D3DDECLUSAGE_POSITION, 0},
		{0, 8, D3DDECLTYPE_FLOAT2, 0, D3DDECLUSAGE_TEXCOORD, 0},
		{0, 16, D3DDECLTYPE_FLOAT4, 0, D3DDECLUSAGE_COLOR, 0},
		D3DVERTEXELEMENT9_END,
	}

	hr = comCall(comVtbl(b.device, devCreateVertexDeclaration),
		b.device,
		uintptr(unsafe.Pointer(&texturedElements[0])),
		uintptr(unsafe.Pointer(&b.texturedDecl)))
	if hr != S_OK {
		return fmt.Errorf("dx9: CreateVertexDeclaration (textured) failed: 0x%x", hr)
	}

	// Shadow vertex declaration:
	// POSITION(float2,0) + TEXCOORD0(float2,8) + COLOR0(float4,16) +
	// TEXCOORD1(float2,32) + TEXCOORD2(float4,40) + TEXCOORD3(float1,56)
	shadowElements := []D3DVERTEXELEMENT9{
		{0, 0, D3DDECLTYPE_FLOAT2, 0, D3DDECLUSAGE_POSITION, 0},
		{0, 8, D3DDECLTYPE_FLOAT2, 0, D3DDECLUSAGE_TEXCOORD, 0},
		{0, 16, D3DDECLTYPE_FLOAT4, 0, D3DDECLUSAGE_COLOR, 0},
		{0, 32, D3DDECLTYPE_FLOAT2, 0, D3DDECLUSAGE_TEXCOORD, 1},
		{0, 40, D3DDECLTYPE_FLOAT4, 0, D3DDECLUSAGE_TEXCOORD, 2},
		{0, 56, 0, 0, D3DDECLUSAGE_TEXCOORD, 3}, // D3DDECLTYPE_FLOAT1 = 0
		D3DVERTEXELEMENT9_END,
	}

	hr = comCall(comVtbl(b.device, devCreateVertexDeclaration),
		b.device,
		uintptr(unsafe.Pointer(&shadowElements[0])),
		uintptr(unsafe.Pointer(&b.shadowDecl)))
	if hr != S_OK {
		return fmt.Errorf("dx9: CreateVertexDeclaration (shadow) failed: 0x%x", hr)
	}

	return nil
}

func (b *Backend) createVertexBuffers() error {
	// Rect VB
	hr := comCall(comVtbl(b.device, devCreateVertexBuffer),
		b.device,
		uintptr(b.rectVBSize),
		uintptr(D3DUSAGE_DYNAMIC|D3DUSAGE_WRITEONLY),
		0, // FVF (not used with vertex declaration)
		uintptr(D3DPOOL_DEFAULT),
		uintptr(unsafe.Pointer(&b.rectVB)),
		0)
	if hr != S_OK {
		return fmt.Errorf("dx9: CreateVertexBuffer (rect) failed: 0x%x", hr)
	}

	// Textured VB
	hr = comCall(comVtbl(b.device, devCreateVertexBuffer),
		b.device,
		uintptr(b.texturedVBSize),
		uintptr(D3DUSAGE_DYNAMIC|D3DUSAGE_WRITEONLY),
		0,
		uintptr(D3DPOOL_DEFAULT),
		uintptr(unsafe.Pointer(&b.texturedVB)),
		0)
	if hr != S_OK {
		return fmt.Errorf("dx9: CreateVertexBuffer (textured) failed: 0x%x", hr)
	}

	// Shadow VB
	if b.shadowVBSize == 0 {
		b.shadowVBSize = 65536
	}
	hr = comCall(comVtbl(b.device, devCreateVertexBuffer),
		b.device,
		uintptr(b.shadowVBSize),
		uintptr(D3DUSAGE_DYNAMIC|D3DUSAGE_WRITEONLY),
		0,
		uintptr(D3DPOOL_DEFAULT),
		uintptr(unsafe.Pointer(&b.shadowVB)),
		0)
	if hr != S_OK {
		return fmt.Errorf("dx9: CreateVertexBuffer (shadow) failed: 0x%x", hr)
	}

	return nil
}

func (b *Backend) compileShader(code, entryPoint, target string) (uintptr, error) {
	codeBytes := []byte(code)
	ep := append([]byte(entryPoint), 0)
	tgt := append([]byte(target), 0)

	var blob, errBlob uintptr
	hr, _, _ := procD3DCompile.Call(
		uintptr(unsafe.Pointer(&codeBytes[0])),
		uintptr(len(codeBytes)),
		0, 0, 0,
		uintptr(unsafe.Pointer(&ep[0])),
		uintptr(unsafe.Pointer(&tgt[0])),
		0, 0,
		uintptr(unsafe.Pointer(&blob)),
		uintptr(unsafe.Pointer(&errBlob)),
	)
	if hr != S_OK {
		msg := "unknown error"
		if errBlob != 0 {
			ptr := comCall(comVtbl(errBlob, 3), errBlob)
			size := comCall(comVtbl(errBlob, 4), errBlob)
			if ptr != 0 && size > 0 {
				msg = string(unsafe.Slice((*byte)(unsafe.Pointer(ptr)), int(size)))
			}
			comRelease(errBlob)
		}
		return 0, fmt.Errorf("D3DCompile failed: %s", msg)
	}
	if errBlob != 0 {
		comRelease(errBlob)
	}
	return blob, nil
}

func (b *Backend) compileVertexShader(code string) (uintptr, error) {
	blob, err := b.compileShader(code, "main", "vs_3_0")
	if err != nil {
		return 0, err
	}
	defer comRelease(blob)

	blobPtr := comCall(comVtbl(blob, 3), blob)

	var vs uintptr
	hr := comCall(comVtbl(b.device, devCreateVertexShaderFn),
		b.device, blobPtr,
		uintptr(unsafe.Pointer(&vs)))
	if hr != S_OK {
		return 0, fmt.Errorf("CreateVertexShader failed: 0x%x", hr)
	}
	return vs, nil
}

func (b *Backend) compilePixelShader(code string) (uintptr, error) {
	blob, err := b.compileShader(code, "main", "ps_3_0")
	if err != nil {
		return 0, err
	}
	defer comRelease(blob)

	blobPtr := comCall(comVtbl(blob, 3), blob)

	var ps uintptr
	hr := comCall(comVtbl(b.device, devCreatePixelShaderFn),
		b.device, blobPtr,
		uintptr(unsafe.Pointer(&ps)))
	if hr != S_OK {
		return 0, fmt.Errorf("CreatePixelShader failed: 0x%x", hr)
	}
	return ps, nil
}

// BeginFrame implements render.Backend.
func (b *Backend) BeginFrame() {
	// Reset pipeline state for new frame
	b.activePipeline = pipelineNone
	b.activeTexture = 0

	dev := b.device

	// Check for lost device
	hr := comCall(comVtbl(dev, devTestCooperativeLevel), dev)
	if hr == D3DERR_DEVICENOTRESET {
		b.resetDevice()
	}

	// Set viewport
	vp := D3DVIEWPORT9{
		Width:  uint32(b.width),
		Height: uint32(b.height),
		MaxZ:   1.0,
	}
	comCall(comVtbl(dev, devSetViewport), dev, uintptr(unsafe.Pointer(&vp)))

	// Set default scissor to full viewport
	scissor := RECT{Right: int32(b.width), Bottom: int32(b.height)}
	comCall(comVtbl(dev, devSetScissorRect), dev, uintptr(unsafe.Pointer(&scissor)))

	// Clear + BeginScene
	comCall(comVtbl(dev, devClear), dev,
		0, 0, // no rects
		uintptr(1|2), // D3DCLEAR_TARGET | D3DCLEAR_ZBUFFER
		uintptr(0xFF000000), // black
		uintptr(math.Float64bits(1.0)), // Z
		0)

	comCall(comVtbl(dev, devBeginScene), dev)
}

// EndFrame implements render.Backend.
func (b *Backend) EndFrame() {
	dev := b.device
	comCall(comVtbl(dev, devEndScene), dev)
	comCall(comVtbl(dev, devPresent), dev, 0, 0, 0, 0)
}

// Resize implements render.Backend.
func (b *Backend) Resize(width, height int) {
	if width <= 0 || height <= 0 {
		return
	}
	b.width = width
	b.height = height
	b.pp.BackBufferWidth = uint32(width)
	b.pp.BackBufferHeight = uint32(height)
	b.resetDevice()
}

func (b *Backend) resetDevice() {
	// Release D3DPOOL_DEFAULT resources before Reset
	if b.rectVB != 0 {
		comRelease(b.rectVB)
		b.rectVB = 0
	}
	if b.texturedVB != 0 {
		comRelease(b.texturedVB)
		b.texturedVB = 0
	}
	if b.shadowVB != 0 {
		comRelease(b.shadowVB)
		b.shadowVB = 0
	}

	hr := comCall(comVtbl(b.device, devReset), b.device, uintptr(unsafe.Pointer(&b.pp)))
	if hr != S_OK {
		return // Will retry next frame
	}

	// Restore render states
	b.setRenderStates()

	// Recreate D3DPOOL_DEFAULT resources
	b.createVertexBuffers()
}

// Submit implements render.Backend.
func (b *Backend) Submit(buf *render.CommandBuffer) {
	if buf == nil {
		return
	}
	b.renderCommands(buf.Commands())
	b.renderCommands(buf.Overlays())
}

func (b *Backend) renderCommands(commands []render.Command) {
	i := 0
	n := len(commands)
	for i < n {
		c := commands[i]
		switch c.Type {
		case render.CmdClip:
			if c.Clip != nil {
				b.applyScissor(c.Clip)
			}
			i++

		case render.CmdRect:
			if c.Rect == nil {
				i++
				continue
			}
			// Scan ahead: batch consecutive CmdRect commands into one draw call.
			j := i + 1
			for j < n && commands[j].Type == render.CmdRect && commands[j].Rect != nil {
				j++
			}
			b.renderRectBatch(commands[i:j])
			i = j

		case render.CmdText:
			if c.Text == nil {
				i++
				continue
			}
			// Scan ahead: batch consecutive CmdText with the same atlas texture.
			atlas := c.Text.Atlas
			j := i + 1
			for j < n && commands[j].Type == render.CmdText &&
				commands[j].Text != nil && commands[j].Text.Atlas == atlas {
				j++
			}
			b.renderTextBatch(commands[i:j])
			i = j

		case render.CmdImage:
			if c.Image == nil {
				i++
				continue
			}
			// Scan ahead: batch consecutive CmdImage with the same texture.
			tex := c.Image.Texture
			j := i + 1
			for j < n && commands[j].Type == render.CmdImage &&
				commands[j].Image != nil && commands[j].Image.Texture == tex {
				j++
			}
			b.renderImageBatch(commands[i:j])
			i = j

		case render.CmdShadow:
			if c.Shadow == nil {
				i++
				continue
			}
			// Scan ahead: batch consecutive CmdShadow commands.
			j := i + 1
			for j < n && commands[j].Type == render.CmdShadow && commands[j].Shadow != nil {
				j++
			}
			b.renderShadowBatch(commands[i:j])
			i = j

		default:
			i++
		}
	}
}

func (b *Backend) applyScissor(clip *render.ClipCmd) {
	logW := float32(b.width) / b.dpiScale
	logH := float32(b.height) / b.dpiScale
	vpW := float32(b.width)
	vpH := float32(b.height)

	x := int32(clip.Bounds.X / logW * vpW)
	y := int32(clip.Bounds.Y / logH * vpH)
	right := int32((clip.Bounds.X + clip.Bounds.Width) / logW * vpW)
	bottom := int32((clip.Bounds.Y + clip.Bounds.Height) / logH * vpH)

	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if right > int32(b.width) {
		right = int32(b.width)
	}
	if bottom > int32(b.height) {
		bottom = int32(b.height)
	}

	scissor := RECT{Left: x, Top: y, Right: right, Bottom: bottom}
	comCall(comVtbl(b.device, devSetScissorRect), b.device, uintptr(unsafe.Pointer(&scissor)))
}

// texelHalfPixelOffset returns the DX9 half-pixel correction in NDC space
// for texture sampling alignment. DX9 pixel centers are at integers (0,0),
// while DX10+ are at (0.5,0.5). Applied only to textured rendering (text/images),
// not to SDF rects which don't sample textures.
func (b *Backend) texelHalfPixelOffset() (float32, float32) {
	return -1.0 / float32(b.width), 1.0 / float32(b.height)
}

// renderRectBatch draws multiple consecutive CmdRect commands in a single draw call.
func (b *Backend) renderRectBatch(commands []render.Command) {
	logW := float32(b.width) / b.dpiScale
	logH := float32(b.height) / b.dpiScale
	s := b.dpiScale

	// Reuse staging buffer
	needed := len(commands) * 6
	if cap(b.rectStaging) < needed {
		b.rectStaging = make([]RectVertex, 0, needed*2)
	}
	b.rectStaging = b.rectStaging[:0]

	for _, c := range commands {
		rect := c.Rect
		opacity := c.Opacity
		x, y, w, h := rect.Bounds.X, rect.Bounds.Y, rect.Bounds.Width, rect.Bounds.Height

		pad := float32(1.0)
		qx, qy := x-pad, y-pad
		qw, qh := w+pad*2, h+pad*2

		ndcX := (qx/logW)*2 - 1
		ndcY := 1 - (qy/logH)*2
		ndcW := (qw / logW) * 2
		ndcH := (qh / logH) * 2

		uvL := -pad / w
		uvT := -pad / h
		uvR := 1.0 + pad/w
		uvB := 1.0 + pad/h

		r, g, bl, a := rect.FillColor.R, rect.FillColor.G, rect.FillColor.B, rect.FillColor.A*opacity
		v := RectVertex{
			ColorR: r, ColorG: g, ColorB: bl, ColorA: a,
			RectW: w * s, RectH: h * s,
			RadiusTL: rect.Corners.TopLeft * s, RadiusTR: rect.Corners.TopRight * s,
			RadiusBR: rect.Corners.BottomRight * s, RadiusBL: rect.Corners.BottomLeft * s,
			BorderWidth: rect.BorderWidth * s,
			BorderR: rect.BorderColor.R, BorderG: rect.BorderColor.G,
			BorderB: rect.BorderColor.B, BorderA: rect.BorderColor.A,
		}

		v0 := v
		v0.PosX, v0.PosY, v0.U, v0.V = ndcX, ndcY, uvL, uvT
		v1 := v
		v1.PosX, v1.PosY, v1.U, v1.V = ndcX+ndcW, ndcY, uvR, uvT
		v2 := v
		v2.PosX, v2.PosY, v2.U, v2.V = ndcX+ndcW, ndcY-ndcH, uvR, uvB
		v3 := v
		v3.PosX, v3.PosY, v3.U, v3.V = ndcX, ndcY-ndcH, uvL, uvB

		b.rectStaging = append(b.rectStaging, v0, v1, v2, v0, v2, v3)
	}

	if len(b.rectStaging) == 0 {
		return
	}

	stride := uint32(unsafe.Sizeof(RectVertex{}))
	dataSize := uint32(len(b.rectStaging)) * stride
	b.uploadToVB(&b.rectVB, &b.rectVBSize, unsafe.Pointer(&b.rectStaging[0]), dataSize)

	b.setRectPipeline()
	comCall(comVtbl(b.device, devSetStreamSource), b.device, 0, b.rectVB, 0, uintptr(stride))
	comCall(comVtbl(b.device, devDrawPrimitive), b.device, D3DPT_TRIANGLELIST, 0, uintptr(len(b.rectStaging)/3))
}

// renderTextBatch draws multiple consecutive same-atlas CmdText commands in a single draw call.
// This is the most impactful optimization — text is the most frequent command in typical UIs.
func (b *Backend) renderTextBatch(commands []render.Command) {
	atlas := commands[0].Text.Atlas
	entry, ok := b.textures[atlas]
	if !ok {
		return
	}

	// Count total glyphs for capacity hint
	totalGlyphs := 0
	for _, c := range commands {
		totalGlyphs += len(c.Text.Glyphs)
	}
	if totalGlyphs == 0 {
		return
	}

	logW := float32(b.width) / b.dpiScale
	logH := float32(b.height) / b.dpiScale
	hpX, hpY := b.texelHalfPixelOffset()

	// Reuse staging buffer (grows but never shrinks)
	needed := totalGlyphs * 6
	if cap(b.texturedStaging) < needed {
		b.texturedStaging = make([]TexturedVertex, 0, needed*2)
	}
	b.texturedStaging = b.texturedStaging[:0]

	for _, c := range commands {
		tc := c.Text
		for _, g := range tc.Glyphs {
			x0 := (g.X/logW)*2 - 1 + hpX
			y0 := 1 - (g.Y/logH)*2 + hpY
			x1 := ((g.X+g.Width)/logW)*2 - 1 + hpX
			y1 := 1 - ((g.Y+g.Height)/logH)*2 + hpY

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

			b.texturedStaging = append(b.texturedStaging, v0, v1, v2, v0, v2, v3)
		}
	}

	stride := uint32(unsafe.Sizeof(TexturedVertex{}))
	dataSize := uint32(len(b.texturedStaging)) * stride
	b.uploadToVB(&b.texturedVB, &b.texturedVBSize, unsafe.Pointer(&b.texturedStaging[0]), dataSize)

	b.setTextPipeline(entry.texture)
	comCall(comVtbl(b.device, devSetStreamSource), b.device, 0, b.texturedVB, 0, uintptr(stride))
	comCall(comVtbl(b.device, devDrawPrimitive), b.device, D3DPT_TRIANGLELIST, 0, uintptr(len(b.texturedStaging)/3))
}

// renderImageBatch draws multiple consecutive same-texture CmdImage commands in a single draw call.
func (b *Backend) renderImageBatch(commands []render.Command) {
	tex := commands[0].Image.Texture
	entry, ok := b.textures[tex]
	if !ok {
		return
	}

	logW := float32(b.width) / b.dpiScale
	logH := float32(b.height) / b.dpiScale
	hpX, hpY := b.texelHalfPixelOffset()

	// Reuse staging buffer (shared with text — safe because batches are sequential)
	needed := len(commands) * 6
	if cap(b.texturedStaging) < needed {
		b.texturedStaging = make([]TexturedVertex, 0, needed*2)
	}
	b.texturedStaging = b.texturedStaging[:0]

	for _, c := range commands {
		ic := c.Image
		x0 := (ic.DstRect.X/logW)*2 - 1 + hpX
		y0 := 1 - (ic.DstRect.Y/logH)*2 + hpY
		x1 := ((ic.DstRect.X+ic.DstRect.Width)/logW)*2 - 1 + hpX
		y1 := 1 - ((ic.DstRect.Y+ic.DstRect.Height)/logH)*2 + hpY

		u0, v0 := ic.SrcRect.X, ic.SrcRect.Y
		u1 := ic.SrcRect.X + ic.SrcRect.Width
		v1 := ic.SrcRect.Y + ic.SrcRect.Height

		tint := ic.Tint
		if tint.A == 0 && tint.R == 0 && tint.G == 0 && tint.B == 0 {
			tint = uimath.Color{R: 1, G: 1, B: 1, A: c.Opacity}
		} else {
			tint.A *= c.Opacity
		}

		clr := TexturedVertex{ColorR: tint.R, ColorG: tint.G, ColorB: tint.B, ColorA: tint.A}
		tv0 := clr
		tv0.PosX, tv0.PosY, tv0.U, tv0.V = x0, y0, u0, v0
		tv1 := clr
		tv1.PosX, tv1.PosY, tv1.U, tv1.V = x1, y0, u1, v0
		tv2 := clr
		tv2.PosX, tv2.PosY, tv2.U, tv2.V = x1, y1, u1, v1
		tv3 := clr
		tv3.PosX, tv3.PosY, tv3.U, tv3.V = x0, y1, u0, v1

		b.texturedStaging = append(b.texturedStaging, tv0, tv1, tv2, tv0, tv2, tv3)
	}

	if len(b.texturedStaging) == 0 {
		return
	}

	stride := uint32(unsafe.Sizeof(TexturedVertex{}))
	dataSize := uint32(len(b.texturedStaging)) * stride
	b.uploadToVB(&b.texturedVB, &b.texturedVBSize, unsafe.Pointer(&b.texturedStaging[0]), dataSize)

	b.setImagePipeline(entry.texture)
	comCall(comVtbl(b.device, devSetStreamSource), b.device, 0, b.texturedVB, 0, uintptr(stride))
	comCall(comVtbl(b.device, devDrawPrimitive), b.device, D3DPT_TRIANGLELIST, 0, uintptr(len(b.texturedStaging)/3))
}

func min32dx9(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func clamp32dx9(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// renderShadowBatch draws multiple consecutive CmdShadow commands in a single draw call.
func (b *Backend) renderShadowBatch(commands []render.Command) {
	logW := float32(b.width) / b.dpiScale
	logH := float32(b.height) / b.dpiScale
	s := b.dpiScale

	// Reuse staging buffer
	needed := len(commands) * 6
	if cap(b.shadowStaging) < needed {
		b.shadowStaging = make([]ShadowVertex, 0, needed*2)
	}
	b.shadowStaging = b.shadowStaging[:0]

	for _, c := range commands {
		sc := c.Shadow
		opacity := c.Opacity

		x, y, w, h := sc.Bounds.X, sc.Bounds.Y, sc.Bounds.Width, sc.Bounds.Height
		spread := sc.SpreadRadius
		blur := sc.BlurRadius
		const pad = float32(1.0)

		elemWlog := w + 2*spread
		elemHlog := h + 2*spread
		if elemWlog < 1 {
			elemWlog = 1
		}
		if elemHlog < 1 {
			elemHlog = 1
		}

		expand := blur*2 + pad
		qx := x + sc.OffsetX - spread - expand
		qy := y + sc.OffsetY - spread - expand
		qw := elemWlog + 2*expand
		qh := elemHlog + 2*expand

		ndcX := (qx/logW)*2 - 1
		ndcY := 1 - (qy/logH)*2
		ndcW := (qw / logW) * 2
		ndcH := (qh / logH) * 2

		uvL := -expand / elemWlog
		uvT := -expand / elemHlog
		uvR := 1.0 + expand/elemWlog
		uvB := 1.0 + expand/elemHlog

		elemW := elemWlog * s
		elemH := elemHlog * s
		maxR := min32dx9(elemW, elemH) * 0.5
		rTL := clamp32dx9(sc.Corners.TopLeft*s+spread*s, 0, maxR)
		rTR := clamp32dx9(sc.Corners.TopRight*s+spread*s, 0, maxR)
		rBR := clamp32dx9(sc.Corners.BottomRight*s+spread*s, 0, maxR)
		rBL := clamp32dx9(sc.Corners.BottomLeft*s+spread*s, 0, maxR)
		blurPx := blur * s

		r, g, bl, a := sc.Color.R, sc.Color.G, sc.Color.B, sc.Color.A*opacity
		sv := ShadowVertex{
			ColorR: r, ColorG: g, ColorB: bl, ColorA: a,
			ElemW: elemW, ElemH: elemH,
			RadiusTL: rTL, RadiusTR: rTR, RadiusBR: rBR, RadiusBL: rBL,
			Blur: blurPx,
		}

		sv0 := sv
		sv0.PosX, sv0.PosY, sv0.U, sv0.V = ndcX, ndcY, uvL, uvT
		sv1 := sv
		sv1.PosX, sv1.PosY, sv1.U, sv1.V = ndcX+ndcW, ndcY, uvR, uvT
		sv2 := sv
		sv2.PosX, sv2.PosY, sv2.U, sv2.V = ndcX+ndcW, ndcY-ndcH, uvR, uvB
		sv3 := sv
		sv3.PosX, sv3.PosY, sv3.U, sv3.V = ndcX, ndcY-ndcH, uvL, uvB

		b.shadowStaging = append(b.shadowStaging, sv0, sv1, sv2, sv0, sv2, sv3)
	}

	if len(b.shadowStaging) == 0 {
		return
	}

	stride := uint32(unsafe.Sizeof(ShadowVertex{}))
	dataSize := uint32(len(b.shadowStaging)) * stride
	b.uploadToVB(&b.shadowVB, &b.shadowVBSize, unsafe.Pointer(&b.shadowStaging[0]), dataSize)

	b.setShadowPipeline()
	comCall(comVtbl(b.device, devSetStreamSource), b.device, 0, b.shadowVB, 0, uintptr(stride))
	comCall(comVtbl(b.device, devDrawPrimitive), b.device, D3DPT_TRIANGLELIST, 0, uintptr(len(b.shadowStaging)/3))
}

// ---- Pipeline state management ----
// These helpers track the active pipeline and skip redundant COM state changes.

func (b *Backend) setRectPipeline() {
	if b.activePipeline == pipelineRect {
		return
	}
	dev := b.device
	comCall(comVtbl(dev, devSetVertexDeclaration), dev, b.rectDecl)
	comCall(comVtbl(dev, devSetVertexShader), dev, b.rectVS)
	comCall(comVtbl(dev, devSetPixelShader), dev, b.rectPS)
	comCall(comVtbl(dev, devSetTexture), dev, 0, 0)
	b.activePipeline = pipelineRect
	b.activeTexture = 0
}

func (b *Backend) setTextPipeline(texture uintptr) {
	dev := b.device
	if b.activePipeline != pipelineText {
		comCall(comVtbl(dev, devSetVertexDeclaration), dev, b.texturedDecl)
		comCall(comVtbl(dev, devSetVertexShader), dev, b.texturedVS)
		comCall(comVtbl(dev, devSetPixelShader), dev, b.textPS)
		// Font atlas is coverage data, not color — disable sRGB texture read
		comCall(comVtbl(dev, devSetSamplerState), dev, 0, D3DSAMP_SRGBTEXTURE, 0)
		b.activePipeline = pipelineText
		b.activeTexture = 0 // force texture rebind below
	}
	if b.activeTexture != texture {
		comCall(comVtbl(dev, devSetTexture), dev, 0, texture)
		b.activeTexture = texture
	}
}

func (b *Backend) setImagePipeline(texture uintptr) {
	dev := b.device
	if b.activePipeline != pipelineTextured {
		comCall(comVtbl(dev, devSetVertexDeclaration), dev, b.texturedDecl)
		comCall(comVtbl(dev, devSetVertexShader), dev, b.texturedVS)
		comCall(comVtbl(dev, devSetPixelShader), dev, b.texturedPS)
		b.activePipeline = pipelineTextured
		b.activeTexture = 0 // force texture rebind below
	}
	if b.activeTexture != texture {
		comCall(comVtbl(dev, devSetTexture), dev, 0, texture)
		b.activeTexture = texture
	}
}

func (b *Backend) setShadowPipeline() {
	if b.activePipeline == pipelineShadow {
		return
	}
	dev := b.device
	comCall(comVtbl(dev, devSetVertexDeclaration), dev, b.shadowDecl)
	comCall(comVtbl(dev, devSetVertexShader), dev, b.shadowVS)
	comCall(comVtbl(dev, devSetPixelShader), dev, b.shadowPS)
	comCall(comVtbl(dev, devSetTexture), dev, 0, 0)
	b.activePipeline = pipelineShadow
	b.activeTexture = 0
}

// uploadToVB locks and writes data to a dynamic vertex buffer, growing if needed.
func (b *Backend) uploadToVB(vb *uintptr, vbSize *uint32, data unsafe.Pointer, dataSize uint32) {
	if dataSize > *vbSize {
		if *vb != 0 {
			comRelease(*vb)
			*vb = 0
		}
		*vbSize = dataSize * 2
		hr := comCall(comVtbl(b.device, devCreateVertexBuffer),
			b.device,
			uintptr(*vbSize),
			uintptr(D3DUSAGE_DYNAMIC|D3DUSAGE_WRITEONLY),
			0,
			uintptr(D3DPOOL_DEFAULT),
			uintptr(unsafe.Pointer(vb)),
			0)
		if hr != S_OK {
			return
		}
	}

	// IDirect3DVertexBuffer9 vtable: IUnknown(0-2), IDirect3DResource9(3-9), Lock(10), Unlock(11)
	var pData unsafe.Pointer
	hr := comCall(comVtbl(*vb, 11), // Lock is at vtable index 11
		*vb, 0, uintptr(dataSize),
		uintptr(unsafe.Pointer(&pData)),
		D3DLOCK_DISCARD)
	if hr != S_OK {
		return
	}
	copy(unsafe.Slice((*byte)(pData), dataSize), unsafe.Slice((*byte)(data), dataSize))
	comCall(comVtbl(*vb, 12), *vb) // Unlock is at vtable index 12
}

// CreateTexture implements render.Backend.
func (b *Backend) CreateTexture(desc render.TextureDesc) (render.TextureHandle, error) {
	d3dFormat := uint32(D3DFMT_A8R8G8B8)
	bytesPerPixel := 4
	switch desc.Format {
	case render.TextureFormatR8:
		d3dFormat = D3DFMT_L8 // Luminance for single-channel on DX9
		bytesPerPixel = 1
	case render.TextureFormatRGBA8:
		d3dFormat = D3DFMT_A8R8G8B8 // DX9 uses BGRA order
	case render.TextureFormatBGRA8:
		d3dFormat = D3DFMT_A8R8G8B8
	}

	var texture uintptr
	hr := comCall(comVtbl(b.device, devCreateTexture),
		b.device,
		uintptr(desc.Width), uintptr(desc.Height),
		1,    // levels
		0,    // usage
		uintptr(d3dFormat),
		uintptr(D3DPOOL_MANAGED),
		uintptr(unsafe.Pointer(&texture)),
		0)
	if hr != S_OK {
		return render.InvalidTexture, fmt.Errorf("dx9: CreateTexture failed: 0x%x", hr)
	}

	// Upload initial data if provided
	if desc.Data != nil {
		b.uploadTextureData(texture, 0, 0, desc.Width, desc.Height, desc.Data, bytesPerPixel, desc.Format)
	}

	handle := b.nextTextureID
	b.nextTextureID++
	b.textures[handle] = &textureEntry{
		texture: texture,
		width:   desc.Width,
		height:  desc.Height,
		format:  desc.Format,
	}

	return handle, nil
}

func (b *Backend) uploadTextureData(texture uintptr, x, y, w, h int, data []byte, bpp int, format render.TextureFormat) {
	var locked D3DLOCKED_RECT
	rect := RECT{
		Left:   int32(x),
		Top:    int32(y),
		Right:  int32(x + w),
		Bottom: int32(y + h),
	}
	hr := comCall(comVtbl(texture, texLockRect),
		texture, 0,
		uintptr(unsafe.Pointer(&locked)),
		uintptr(unsafe.Pointer(&rect)),
		0)
	if hr != S_OK {
		return
	}

	srcPitch := w * bpp
	for row := 0; row < h; row++ {
		srcOff := row * srcPitch
		dstPtr := unsafe.Add(locked.PBits, row*int(locked.Pitch))
		src := data[srcOff : srcOff+srcPitch]

		if format == render.TextureFormatRGBA8 && bpp == 4 {
			// Convert RGBA → BGRA for DX9
			dst := unsafe.Slice((*byte)(dstPtr), srcPitch)
			for i := 0; i < len(src); i += 4 {
				dst[i+0] = src[i+2] // B
				dst[i+1] = src[i+1] // G
				dst[i+2] = src[i+0] // R
				dst[i+3] = src[i+3] // A
			}
		} else {
			copy(unsafe.Slice((*byte)(dstPtr), srcPitch), src)
		}
	}

	comCall(comVtbl(texture, texUnlockRect), texture, 0)
}

// UpdateTexture implements render.Backend.
func (b *Backend) UpdateTexture(handle render.TextureHandle, region uimath.Rect, data []byte) error {
	entry, ok := b.textures[handle]
	if !ok {
		return fmt.Errorf("dx9: texture %d not found", handle)
	}

	bpp := 4
	if entry.format == render.TextureFormatR8 {
		bpp = 1
	}

	b.uploadTextureData(entry.texture, int(region.X), int(region.Y),
		int(region.Width), int(region.Height), data, bpp, entry.format)
	return nil
}

// DestroyTexture implements render.Backend.
func (b *Backend) DestroyTexture(handle render.TextureHandle) {
	entry, ok := b.textures[handle]
	if !ok {
		return
	}
	if entry.texture != 0 {
		comRelease(entry.texture)
	}
	delete(b.textures, handle)
}

// MaxTextureSize implements render.Backend.
func (b *Backend) MaxTextureSize() int {
	return 4096 // DX9 minimum guaranteed is 4096
}

// DPIScale implements render.Backend.
func (b *Backend) DPIScale() float32 {
	return b.dpiScale
}

// ReadPixels implements render.Backend.
func (b *Backend) ReadPixels() (*image.RGBA, error) {
	if b.width <= 0 || b.height <= 0 {
		return nil, fmt.Errorf("dx9: invalid dimensions")
	}

	// Get the back buffer surface
	var backBuffer uintptr
	hr := comCall(comVtbl(b.device, devGetBackBuffer),
		b.device, 0, 0, 0, // SwapChain=0, BackBuffer=0, Type=D3DBACKBUFFER_TYPE_MONO
		uintptr(unsafe.Pointer(&backBuffer)))
	if hr != S_OK {
		return nil, fmt.Errorf("dx9: GetBackBuffer failed: 0x%x", hr)
	}
	defer comRelease(backBuffer)

	// Create an offscreen plain surface in SYSTEMMEM for readback
	var readbackSurf uintptr
	hr = comCall(comVtbl(b.device, devCreateOffscreenPlainSurface),
		b.device,
		uintptr(b.width), uintptr(b.height),
		D3DFMT_A8R8G8B8,
		uintptr(D3DPOOL_SYSTEMMEM),
		uintptr(unsafe.Pointer(&readbackSurf)),
		0)
	if hr != S_OK {
		return nil, fmt.Errorf("dx9: CreateOffscreenPlainSurface failed: 0x%x", hr)
	}
	defer comRelease(readbackSurf)

	// GetRenderTargetData copies from video memory surface to system memory surface
	hr = comCall(comVtbl(b.device, devGetRenderTargetData),
		b.device, backBuffer, readbackSurf)
	if hr != S_OK {
		return nil, fmt.Errorf("dx9: GetRenderTargetData failed: 0x%x", hr)
	}

	// Lock the readback surface
	var locked D3DLOCKED_RECT
	hr = comCall(comVtbl(readbackSurf, surfLockRect),
		readbackSurf, uintptr(unsafe.Pointer(&locked)), 0, 0) // D3DLOCK_READONLY=0x10
	if hr != S_OK {
		return nil, fmt.Errorf("dx9: LockRect failed: 0x%x", hr)
	}

	img := image.NewRGBA(image.Rect(0, 0, b.width, b.height))
	for y := 0; y < b.height; y++ {
		src := unsafe.Add(locked.PBits, y*int(locked.Pitch))
		srcSlice := unsafe.Slice((*byte)(src), b.width*4)
		dstOff := y * img.Stride
		// Convert BGRA → RGBA
		for x := 0; x < b.width; x++ {
			si := x * 4
			di := dstOff + x*4
			img.Pix[di+0] = srcSlice[si+2] // R
			img.Pix[di+1] = srcSlice[si+1] // G
			img.Pix[di+2] = srcSlice[si+0] // B
			img.Pix[di+3] = srcSlice[si+3] // A
		}
	}

	comCall(comVtbl(readbackSurf, surfUnlockRect), readbackSurf)
	return img, nil
}

// Destroy implements render.Backend.
func (b *Backend) Destroy() {
	for handle := range b.textures {
		b.DestroyTexture(handle)
	}

	if b.rectVB != 0 {
		comRelease(b.rectVB)
	}
	if b.texturedVB != 0 {
		comRelease(b.texturedVB)
	}
	if b.shadowVB != 0 {
		comRelease(b.shadowVB)
	}
	if b.rectDecl != 0 {
		comRelease(b.rectDecl)
	}
	if b.texturedDecl != 0 {
		comRelease(b.texturedDecl)
	}
	if b.shadowDecl != 0 {
		comRelease(b.shadowDecl)
	}
	if b.rectVS != 0 {
		comRelease(b.rectVS)
	}
	if b.rectPS != 0 {
		comRelease(b.rectPS)
	}
	if b.texturedVS != 0 {
		comRelease(b.texturedVS)
	}
	if b.texturedPS != 0 {
		comRelease(b.texturedPS)
	}
	if b.textPS != 0 {
		comRelease(b.textPS)
	}
	if b.shadowVS != 0 {
		comRelease(b.shadowVS)
	}
	if b.shadowPS != 0 {
		comRelease(b.shadowPS)
	}
	if b.device != 0 {
		comRelease(b.device)
	}
	if b.d3d9 != 0 {
		comRelease(b.d3d9)
	}
}

// Verify Backend implements render.Backend at compile time.
var _ render.Backend = (*Backend)(nil)

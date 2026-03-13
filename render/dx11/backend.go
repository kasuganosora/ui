//go:build windows

package dx11

import (
	"fmt"
	"image"
	"syscall"
	"unsafe"

	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/render"
)

// RectVertex matches the rect shader input layout.
type RectVertex struct {
	PosX, PosY                                     float32 // NDC position
	U, V                                           float32 // UV for SDF
	ColorR, ColorG, ColorB, ColorA                 float32
	RectW, RectH                                   float32 // Rect size in physical pixels
	RadiusTL, RadiusTR, RadiusBR, RadiusBL         float32
	BorderWidth                                    float32
	BorderR, BorderG, BorderB, BorderA             float32
}

// ShadowVertex matches the shadow shader input layout — 15 float32, 60 bytes.
type ShadowVertex struct {
	PosX, PosY                              float32
	U, V                                    float32
	ColorR, ColorG, ColorB, ColorA          float32
	ElemW, ElemH                            float32 // Spread-adjusted element size in physical px
	RadiusTL, RadiusTR, RadiusBR, RadiusBL float32 // Corner radii in physical px
	Blur                                    float32
}

// TexturedVertex for text and image rendering.
// Text leaves RectW..RadiusBL as zero; images populate them for SDF rounded corners.
type TexturedVertex struct {
	PosX, PosY                             float32
	U, V                                   float32
	ColorR, ColorG, ColorB, ColorA         float32
	RectW, RectH                           float32 // Rect size in physical pixels (0 = no SDF clipping)
	RadiusTL, RadiusTR, RadiusBR, RadiusBL float32
}

type textureEntry struct {
	texture uintptr // ID3D11Texture2D*
	srv     uintptr // ID3D11ShaderResourceView*
	width   int
	height  int
	format  render.TextureFormat
}

// ---- ID3D11Device vtable indices (IUnknown: 0-2) ----
// Order follows d3d11.h: CreateBuffer(3), CreateTexture1D(4), CreateTexture2D(5),
// CreateTexture3D(6), CreateShaderResourceView(7), CreateUnorderedAccessView(8),
// CreateRenderTargetView(9), CreateDepthStencilView(10), CreateInputLayout(11),
// CreateVertexShader(12), CreateGeometryShader(13), CreateGeometryShaderWithStreamOutput(14),
// CreatePixelShader(15), CreateHullShader(16), CreateDomainShader(17),
// CreateComputeShader(18), CreateClassLinkage(19), CreateBlendState(20),
// CreateDepthStencilState(21), CreateRasterizerState(22), CreateSamplerState(23).
const (
	devCreateBuffer              = 3
	devCreateTexture2D           = 5
	devCreateShaderResourceView  = 7
	devCreateRenderTargetView    = 9
	devCreateInputLayout         = 11
	devCreateVertexShader        = 12
	devCreatePixelShader         = 15
	devCreateBlendState          = 20
	devCreateRasterizerState     = 22
	devCreateSamplerState        = 23
)

// ---- ID3D11DeviceContext vtable indices (IUnknown: 0-2, ID3D11DeviceChild: 3-6) ----
const (
	ctxPSSetShaderResources = 8
	ctxPSSetShader          = 9
	ctxPSSetSamplers        = 10
	ctxVSSetShader          = 11
	ctxDraw                 = 13
	ctxMap                  = 14
	ctxUnmap                = 15
	ctxIASetInputLayout     = 17
	ctxIASetVertexBuffers   = 18
	ctxIASetPrimitiveTopology = 24
	ctxOMSetRenderTargets   = 33
	ctxOMSetBlendState      = 35
	ctxRSSetState           = 43
	ctxRSSetViewports       = 44
	ctxRSSetScissorRects    = 45
	ctxCopyResource         = 47
	ctxUpdateSubresource    = 48
	ctxClearRenderTargetView = 50
)

// ---- IDXGISwapChain vtable indices ----
const (
	scPresent       = 8
	scGetBuffer     = 9
	scResizeBuffers = 13
)

// Backend implements render.Backend using DirectX 11.
type Backend struct {
	loader *Loader

	device       uintptr // ID3D11Device*
	context      uintptr // ID3D11DeviceContext*
	swapChain    uintptr // IDXGISwapChain*
	renderTarget uintptr // ID3D11RenderTargetView*

	// Pipeline state objects
	blendState      uintptr // ID3D11BlendState*
	rasterizerState uintptr // ID3D11RasterizerState*

	// Rect pipeline
	rectVS          uintptr // ID3D11VertexShader*
	rectPS          uintptr // ID3D11PixelShader*
	rectInputLayout uintptr // ID3D11InputLayout*
	shadowVS          uintptr // ID3D11VertexShader* (shadow)
	shadowPS          uintptr // ID3D11PixelShader*  (shadow)
	shadowInputLayout uintptr // ID3D11InputLayout*  (shadow)

	// Textured pipeline (images)
	texturedVS          uintptr // ID3D11VertexShader*
	texturedPS          uintptr // ID3D11PixelShader*
	texturedInputLayout uintptr

	// Text pipeline
	textPS uintptr // ID3D11PixelShader* (SDF)

	// Samplers
	linearSampler uintptr // ID3D11SamplerState*

	// Dynamic vertex buffer
	vertexBuffer  uintptr // ID3D11Buffer*
	vertexBufSize uint32

	// State
	width, height int
	dpiScale      float32
	hwnd          uintptr

	// Texture management
	nextTextureID render.TextureHandle
	textures      map[render.TextureHandle]*textureEntry
}

// New creates a new DX11 backend.
func New() *Backend {
	return &Backend{
		nextTextureID: 1,
		textures:      make(map[render.TextureHandle]*textureEntry),
		dpiScale:      1.0,
	}
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

	// Create device and swap chain
	sd := DXGI_SWAP_CHAIN_DESC{
		BufferDesc: DXGI_MODE_DESC{
			Width:  uint32(b.width),
			Height: uint32(b.height),
			Format: DXGI_FORMAT_R8G8B8A8_UNORM,
			RefreshRate: DXGI_RATIONAL{
				Numerator:   60,
				Denominator: 1,
			},
		},
		SampleDesc:   DXGI_SAMPLE_DESC{Count: 1},
		BufferUsage:  DXGI_USAGE_RENDER_TARGET_OUTPUT,
		BufferCount:  2,
		OutputWindow: b.hwnd,
		Windowed:     1,
		SwapEffect:   DXGI_SWAP_EFFECT_FLIP_DISCARD,
	}

	featureLevel := uint32(0)
	featureLevels := [3]uint32{
		D3D_FEATURE_LEVEL_11_0,
		D3D_FEATURE_LEVEL_10_1,
		D3D_FEATURE_LEVEL_10_0,
	}

	hr, _, _ := b.loader.d3d11CreateDeviceAndSwapChain.Call(
		0, // pAdapter
		D3D_DRIVER_TYPE_HARDWARE,
		0, // Software
		0, // Flags (no debug)
		uintptr(unsafe.Pointer(&featureLevels[0])),
		uintptr(len(featureLevels)),
		D3D11_SDK_VERSION,
		uintptr(unsafe.Pointer(&sd)),
		uintptr(unsafe.Pointer(&b.swapChain)),
		uintptr(unsafe.Pointer(&b.device)),
		uintptr(unsafe.Pointer(&featureLevel)),
		uintptr(unsafe.Pointer(&b.context)),
	)
	if hr != S_OK {
		return fmt.Errorf("dx11: D3D11CreateDeviceAndSwapChain failed: 0x%x", hr)
	}

	// Create render target from back buffer
	if err := b.createRenderTarget(); err != nil {
		return err
	}

	// Create pipeline state
	if err := b.createPipelineState(); err != nil {
		return err
	}

	// Create initial vertex buffer (dynamic, 64KB)
	b.vertexBufSize = 65536
	if err := b.createVertexBuffer(b.vertexBufSize); err != nil {
		return err
	}

	return nil
}

func (b *Backend) createRenderTarget() error {
	// IDXGISwapChain::GetBuffer(0, IID_ID3D11Texture2D, &backBuffer)
	var backBuffer uintptr
	iidTexture2D := GUID{
		0x6f15aaf2, 0xd208, 0x4e89,
		[8]byte{0x9a, 0xb4, 0x48, 0x95, 0x35, 0xd3, 0x4f, 0x9c},
	}

	hr := comCall(comVtbl(b.swapChain, scGetBuffer),
		b.swapChain, 0,
		uintptr(unsafe.Pointer(&iidTexture2D)),
		uintptr(unsafe.Pointer(&backBuffer)))
	if hr != S_OK {
		return fmt.Errorf("dx11: GetBuffer failed: 0x%x", hr)
	}
	defer comRelease(backBuffer)

	// Create RTV with sRGB format for linear-space alpha blending
	// (matching Vulkan VK_FORMAT_B8G8R8A8_SRGB).
	// FLIP_DISCARD supports cross-format RTV: UNORM swap chain + SRGB RTV.
	rtvDesc := D3D11_RENDER_TARGET_VIEW_DESC{
		Format:        DXGI_FORMAT_R8G8B8A8_UNORM_SRGB,
		ViewDimension: 4, // D3D11_RTV_DIMENSION_TEXTURE2D
	}
	hr = comCall(comVtbl(b.device, devCreateRenderTargetView),
		b.device, backBuffer,
		uintptr(unsafe.Pointer(&rtvDesc)),
		uintptr(unsafe.Pointer(&b.renderTarget)))
	if hr != S_OK {
		return fmt.Errorf("dx11: CreateRenderTargetView failed: 0x%x", hr)
	}

	return nil
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

func (b *Backend) createPipelineState() error {
	// Create blend state (alpha blending)
	blendDesc := D3D11_BLEND_DESC{}
	blendDesc.RenderTarget[0] = D3D11_RENDER_TARGET_BLEND_DESC{
		BlendEnable:           1,
		SrcBlend:              D3D11_BLEND_SRC_ALPHA,
		DestBlend:             D3D11_BLEND_INV_SRC_ALPHA,
		BlendOp:               D3D11_BLEND_OP_ADD,
		SrcBlendAlpha:         D3D11_BLEND_ONE,
		DestBlendAlpha:        D3D11_BLEND_INV_SRC_ALPHA,
		BlendOpAlpha:          D3D11_BLEND_OP_ADD,
		RenderTargetWriteMask: D3D11_COLOR_WRITE_ENABLE_ALL,
	}

	hr := comCall(comVtbl(b.device, devCreateBlendState),
		b.device,
		uintptr(unsafe.Pointer(&blendDesc)),
		uintptr(unsafe.Pointer(&b.blendState)))
	if hr != S_OK {
		return fmt.Errorf("dx11: CreateBlendState failed: 0x%x", hr)
	}

	// Create rasterizer state (no culling, scissor enabled)
	rastDesc := D3D11_RASTERIZER_DESC{
		FillMode:        D3D11_FILL_SOLID,
		CullMode:        D3D11_CULL_NONE,
		ScissorEnable:   1,
		DepthClipEnable: 1,
	}
	hr = comCall(comVtbl(b.device, devCreateRasterizerState),
		b.device,
		uintptr(unsafe.Pointer(&rastDesc)),
		uintptr(unsafe.Pointer(&b.rasterizerState)))
	if hr != S_OK {
		return fmt.Errorf("dx11: CreateRasterizerState failed: 0x%x", hr)
	}

	// Create linear sampler
	samplerDesc := D3D11_SAMPLER_DESC{
		Filter:         D3D11_FILTER_MIN_MAG_LINEAR_MIP_POINT,
		AddressU:       D3D11_TEXTURE_ADDRESS_CLAMP,
		AddressV:       D3D11_TEXTURE_ADDRESS_CLAMP,
		AddressW:       D3D11_TEXTURE_ADDRESS_CLAMP,
		ComparisonFunc: D3D11_COMPARISON_NEVER,
		MaxLOD:         0,
	}
	hr = comCall(comVtbl(b.device, devCreateSamplerState),
		b.device,
		uintptr(unsafe.Pointer(&samplerDesc)),
		uintptr(unsafe.Pointer(&b.linearSampler)))
	if hr != S_OK {
		return fmt.Errorf("dx11: CreateSamplerState failed: 0x%x", hr)
	}

	// Compile and create shaders
	if err := b.createShaders(); err != nil {
		return err
	}

	return nil
}

// D3DCompile from d3dcompiler_47.dll
var (
	d3dcompiler    = syscall.NewLazyDLL("d3dcompiler_47.dll")
	procD3DCompile = d3dcompiler.NewProc("D3DCompile")
)

func (b *Backend) createShaders() error {
	// ---- Rect shaders (SDF rounded rectangles) ----

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
    float4 pos : SV_POSITION;
    float2 uv : TEXCOORD0;
    float4 color : COLOR0;
    float2 rectSize : TEXCOORD1;
    float4 radius : TEXCOORD2;
    float borderWidth : TEXCOORD3;
    float4 borderColor : COLOR1;
};
float3 srgbToLinear(float3 c) {
    return c <= 0.04045 ? c / 12.92 : pow((c + 0.055) / 1.055, 2.4);
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
    float4 pos : SV_POSITION;
    float2 uv : TEXCOORD0;
    float4 color : COLOR0;
    float2 rectSize : TEXCOORD1;
    float4 radius : TEXCOORD2;
    float borderWidth : TEXCOORD3;
    float4 borderColor : COLOR1;
};
float roundedBoxSDF(float2 p, float2 b, float4 r) {
    float radius = (p.x > 0.0) ? ((p.y > 0.0) ? r.z : r.y) : ((p.y > 0.0) ? r.w : r.x);
    float2 q = abs(p) - b + float2(radius, radius);
    return min(max(q.x, q.y), 0.0) + length(max(q, 0.0)) - radius;
}
float4 main(PS_INPUT input) : SV_TARGET {
    float2 p = (input.uv - 0.5) * input.rectSize;
    float2 b = input.rectSize * 0.5;
    float dist = roundedBoxSDF(p, b, input.radius);
    float aa = fwidth(dist);
    float fillAlpha = 1.0 - smoothstep(0.0, aa, dist);
    if (input.borderWidth > 0.0) {
        float innerDist = dist + input.borderWidth;
        float fillMask = 1.0 - smoothstep(0.0, aa, innerDist);
        float4 color = lerp(input.borderColor, input.color, fillMask);
        return float4(color.rgb, color.a * fillAlpha);
    }
    return float4(input.color.rgb, input.color.a * fillAlpha);
}`

	var err error
	b.rectVS, b.rectInputLayout, err = b.compileVertexShader(rectVSCode, rectInputElements())
	if err != nil {
		return fmt.Errorf("dx11: rect VS: %w", err)
	}
	b.rectPS, err = b.compilePixelShader(rectPSCode)
	if err != nil {
		return fmt.Errorf("dx11: rect PS: %w", err)
	}

	// ---- Shadow shaders ----

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
    float4 pos      : SV_POSITION;
    float2 uv       : TEXCOORD0;
    float4 color    : COLOR0;
    float2 elemSize : TEXCOORD1;
    float4 radii    : TEXCOORD2;
    float  blur     : TEXCOORD3;
};
float3 srgbToLinear(float3 c) {
    return c <= 0.04045 ? c / 12.92 : pow((c + 0.055) / 1.055, 2.4);
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
    float4 pos      : SV_POSITION;
    float2 uv       : TEXCOORD0;
    float4 color    : COLOR0;
    float2 elemSize : TEXCOORD1;
    float4 radii    : TEXCOORD2;
    float  blur     : TEXCOORD3;
};
float roundedBoxSDF(float2 p, float2 b, float4 r) {
    float radius = (p.x > 0.0) ? ((p.y > 0.0) ? r.z : r.y) : ((p.y > 0.0) ? r.w : r.x);
    float2 q = abs(p) - b + float2(radius, radius);
    return min(max(q.x, q.y), 0.0) + length(max(q, 0.0)) - radius;
}
float4 main(PS_INPUT input) : SV_TARGET {
    float2 elemHalf = input.elemSize * 0.5;
    float2 p = (input.uv - 0.5) * input.elemSize;
    float dist = roundedBoxSDF(p, elemHalf, input.radii);
    float blur = max(input.blur, 0.0);
    float alpha;
    float aa = max(max(abs(ddx(dist)), abs(ddy(dist))), 0.5);
    if (blur < 1.0) {
        alpha = 1.0 - smoothstep(-aa, aa, dist);
    } else {
        float sigma = blur * 0.5;
        float outside = max(0.0, dist);
        float t = outside / sigma;
        alpha = exp(-t * t * 0.5);
        // Fade out at 3-sigma from the element boundary (dist - sigma*3 transitions
        // from negative to positive at the 3-sigma cutoff point).
        alpha *= 1.0 - smoothstep(-aa, aa, dist - sigma * 3.0);
        alpha = clamp(alpha, 0.0, 1.0);
    }
    return float4(input.color.rgb, input.color.a * alpha);
}`

	b.shadowVS, b.shadowInputLayout, err = b.compileVertexShader(shadowVSCode, shadowInputElements())
	if err != nil {
		return fmt.Errorf("dx11: shadow VS: %w", err)
	}
	b.shadowPS, err = b.compilePixelShader(shadowPSCode)
	if err != nil {
		return fmt.Errorf("dx11: shadow PS: %w", err)
	}

	// ---- Textured shaders (images) ----

	texturedVSCode := `
struct VS_INPUT {
    float2 pos : POSITION;
    float2 uv : TEXCOORD0;
    float4 color : COLOR0;
    float2 rectSize : TEXCOORD1;
    float4 radius : TEXCOORD2;
};
struct PS_INPUT {
    float4 pos : SV_POSITION;
    float2 uv : TEXCOORD0;
    float4 color : COLOR0;
    float2 rectSize : TEXCOORD1;
    float4 radius : TEXCOORD2;
};
float3 srgbToLinear(float3 c) {
    return c <= 0.04045 ? c / 12.92 : pow((c + 0.055) / 1.055, 2.4);
}
PS_INPUT main(VS_INPUT input) {
    PS_INPUT output;
    output.pos = float4(input.pos, 0.0, 1.0);
    output.uv = input.uv;
    output.color = float4(srgbToLinear(input.color.rgb), input.color.a);
    output.rectSize = input.rectSize;
    output.radius = input.radius;
    return output;
}`

	texturedPSCode := `
Texture2D tex : register(t0);
SamplerState samp : register(s0);
struct PS_INPUT {
    float4 pos : SV_POSITION;
    float2 uv : TEXCOORD0;
    float4 color : COLOR0;
    float2 rectSize : TEXCOORD1;
    float4 radius : TEXCOORD2;
};
float roundedBoxSDF(float2 p, float2 b, float4 r) {
    float rad = (p.x > 0.0) ? ((p.y > 0.0) ? r.z : r.y) : ((p.y > 0.0) ? r.w : r.x);
    float2 q = abs(p) - b + float2(rad, rad);
    return min(max(q.x, q.y), 0.0) + length(max(q, 0.0)) - rad;
}
float4 main(PS_INPUT input) : SV_TARGET {
    float4 texColor = tex.Sample(samp, input.uv);
    float4 result = texColor * input.color;
    if (input.rectSize.x > 0.0) {
        float2 p = (input.uv - 0.5) * input.rectSize;
        float2 b = input.rectSize * 0.5;
        float dist = roundedBoxSDF(p, b, input.radius);
        float aa = fwidth(dist);
        result.a *= 1.0 - smoothstep(0.0, aa, dist);
    }
    return result;
}`

	// ---- Text shader (SDF alpha) ----

	textPSCode := `
Texture2D tex : register(t0);
SamplerState samp : register(s0);
struct PS_INPUT {
    float4 pos : SV_POSITION;
    float2 uv : TEXCOORD0;
    float4 color : COLOR0;
};
float4 main(PS_INPUT input) : SV_TARGET {
    float coverage = tex.Sample(samp, input.uv).r;
    return float4(input.color.rgb, input.color.a * coverage);
}`

	b.texturedVS, b.texturedInputLayout, err = b.compileVertexShader(texturedVSCode, texturedInputElements())
	if err != nil {
		return fmt.Errorf("dx11: textured VS: %w", err)
	}
	b.texturedPS, err = b.compilePixelShader(texturedPSCode)
	if err != nil {
		return fmt.Errorf("dx11: textured PS: %w", err)
	}
	b.textPS, err = b.compilePixelShader(textPSCode)
	if err != nil {
		return fmt.Errorf("dx11: text PS: %w", err)
	}

	return nil
}

// semanticNames caches null-terminated C strings for D3D11 input element semantics.
var semanticNames = map[string]*byte{}

func semanticName(name string) *byte {
	if p, ok := semanticNames[name]; ok {
		return p
	}
	b := append([]byte(name), 0)
	p := &b[0]
	semanticNames[name] = p
	return p
}

// rectInputElements returns the input layout for rect vertices.
func rectInputElements() []D3D11_INPUT_ELEMENT_DESC {
	return []D3D11_INPUT_ELEMENT_DESC{
		{semanticName("POSITION"), 0, DXGI_FORMAT_R32G32_FLOAT, 0, 0, 0, 0},
		{semanticName("TEXCOORD"), 0, DXGI_FORMAT_R32G32_FLOAT, 0, 8, 0, 0},
		{semanticName("COLOR"), 0, DXGI_FORMAT_R32G32B32A32_FLOAT, 0, 16, 0, 0},
		{semanticName("TEXCOORD"), 1, DXGI_FORMAT_R32G32_FLOAT, 0, 32, 0, 0},
		{semanticName("TEXCOORD"), 2, DXGI_FORMAT_R32G32B32A32_FLOAT, 0, 40, 0, 0},
		{semanticName("TEXCOORD"), 3, DXGI_FORMAT_R32_FLOAT, 0, 56, 0, 0},
		{semanticName("COLOR"), 1, DXGI_FORMAT_R32G32B32A32_FLOAT, 0, 60, 0, 0},
	}
}

// shadowInputElements returns the input layout for shadow vertices (ShadowVertex, 60 bytes).
func shadowInputElements() []D3D11_INPUT_ELEMENT_DESC {
	return []D3D11_INPUT_ELEMENT_DESC{
		{semanticName("POSITION"), 0, DXGI_FORMAT_R32G32_FLOAT, 0, 0, 0, 0},         // pos (0)
		{semanticName("TEXCOORD"), 0, DXGI_FORMAT_R32G32_FLOAT, 0, 8, 0, 0},         // uv (8)
		{semanticName("COLOR"), 0, DXGI_FORMAT_R32G32B32A32_FLOAT, 0, 16, 0, 0},     // color (16)
		{semanticName("TEXCOORD"), 1, DXGI_FORMAT_R32G32_FLOAT, 0, 32, 0, 0},        // elemSize (32)
		{semanticName("TEXCOORD"), 2, DXGI_FORMAT_R32G32B32A32_FLOAT, 0, 40, 0, 0},  // radii (40)
		{semanticName("TEXCOORD"), 3, DXGI_FORMAT_R32_FLOAT, 0, 56, 0, 0},           // blur (56)
	}
}

// texturedInputElements returns the input layout for textured vertices.
func texturedInputElements() []D3D11_INPUT_ELEMENT_DESC {
	return []D3D11_INPUT_ELEMENT_DESC{
		{semanticName("POSITION"), 0, DXGI_FORMAT_R32G32_FLOAT, 0, 0, 0, 0},
		{semanticName("TEXCOORD"), 0, DXGI_FORMAT_R32G32_FLOAT, 0, 8, 0, 0},
		{semanticName("COLOR"), 0, DXGI_FORMAT_R32G32B32A32_FLOAT, 0, 16, 0, 0},
		{semanticName("TEXCOORD"), 1, DXGI_FORMAT_R32G32_FLOAT, 0, 32, 0, 0},           // rectSize
		{semanticName("TEXCOORD"), 2, DXGI_FORMAT_R32G32B32A32_FLOAT, 0, 40, 0, 0},     // radius (TL,TR,BR,BL)
	}
}

func (b *Backend) createVertexBuffer(size uint32) error {
	desc := D3D11_BUFFER_DESC{
		ByteWidth:      size,
		Usage:          D3D11_USAGE_DYNAMIC,
		BindFlags:      D3D11_BIND_VERTEX_BUFFER,
		CPUAccessFlags: D3D11_CPU_ACCESS_WRITE,
	}
	hr := comCall(comVtbl(b.device, devCreateBuffer),
		b.device,
		uintptr(unsafe.Pointer(&desc)),
		0,
		uintptr(unsafe.Pointer(&b.vertexBuffer)))
	if hr != S_OK {
		return fmt.Errorf("dx11: CreateBuffer failed: 0x%x", hr)
	}
	return nil
}

func (b *Backend) compileVertexShader(code string, inputElems []D3D11_INPUT_ELEMENT_DESC) (vs uintptr, il uintptr, err error) {
	blob, err := b.compileShader(code, "main", "vs_5_0")
	if err != nil {
		return 0, 0, err
	}
	defer comRelease(blob)

	// ID3DBlob::GetBufferPointer (vtable[3]), GetBufferSize (vtable[4])
	blobPtr := comCall(comVtbl(blob, 3), blob)
	blobSize := comCall(comVtbl(blob, 4), blob)

	// Create vertex shader
	hr := comCall(comVtbl(b.device, devCreateVertexShader),
		b.device, blobPtr, blobSize, 0,
		uintptr(unsafe.Pointer(&vs)))
	if hr != S_OK {
		return 0, 0, fmt.Errorf("CreateVertexShader failed: 0x%x", hr)
	}

	// Create input layout
	hr = comCall(comVtbl(b.device, devCreateInputLayout),
		b.device,
		uintptr(unsafe.Pointer(&inputElems[0])),
		uintptr(len(inputElems)),
		blobPtr, blobSize,
		uintptr(unsafe.Pointer(&il)))
	if hr != S_OK {
		comRelease(vs)
		return 0, 0, fmt.Errorf("CreateInputLayout failed: 0x%x", hr)
	}

	return vs, il, nil
}

func (b *Backend) compilePixelShader(code string) (uintptr, error) {
	blob, err := b.compileShader(code, "main", "ps_5_0")
	if err != nil {
		return 0, err
	}
	defer comRelease(blob)

	blobPtr := comCall(comVtbl(blob, 3), blob)
	blobSize := comCall(comVtbl(blob, 4), blob)

	var ps uintptr
	hr := comCall(comVtbl(b.device, devCreatePixelShader),
		b.device, blobPtr, blobSize, 0,
		uintptr(unsafe.Pointer(&ps)))
	if hr != S_OK {
		return 0, fmt.Errorf("CreatePixelShader failed: 0x%x", hr)
	}
	return ps, nil
}

func (b *Backend) compileShader(code, entryPoint, target string) (uintptr, error) {
	codeBytes := []byte(code)
	ep := append([]byte(entryPoint), 0)
	tgt := append([]byte(target), 0)

	var blob, errBlob uintptr
	hr, _, _ := procD3DCompile.Call(
		uintptr(unsafe.Pointer(&codeBytes[0])),
		uintptr(len(codeBytes)),
		0, // pSourceName
		0, // pDefines
		0, // pInclude
		uintptr(unsafe.Pointer(&ep[0])),
		uintptr(unsafe.Pointer(&tgt[0])),
		0, // Flags1
		0, // Flags2
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

// BeginFrame implements render.Backend.
func (b *Backend) BeginFrame() {
	// Set render target
	syscall.SyscallN(comVtbl(b.context, ctxOMSetRenderTargets),
		b.context, 1,
		uintptr(unsafe.Pointer(&b.renderTarget)),
		0)

	// Set viewport
	vp := D3D11_VIEWPORT{
		Width:    float32(b.width),
		Height:   float32(b.height),
		MaxDepth: 1.0,
	}
	syscall.SyscallN(comVtbl(b.context, ctxRSSetViewports),
		b.context, 1,
		uintptr(unsafe.Pointer(&vp)))

	// Set blend state
	blendFactor := [4]float32{0, 0, 0, 0}
	syscall.SyscallN(comVtbl(b.context, ctxOMSetBlendState),
		b.context, b.blendState,
		uintptr(unsafe.Pointer(&blendFactor[0])),
		0xFFFFFFFF)

	// Set rasterizer state
	syscall.SyscallN(comVtbl(b.context, ctxRSSetState),
		b.context, b.rasterizerState)

	// Set default scissor to full viewport
	scissor := D3D11_RECT{Right: int32(b.width), Bottom: int32(b.height)}
	syscall.SyscallN(comVtbl(b.context, ctxRSSetScissorRects),
		b.context, 1,
		uintptr(unsafe.Pointer(&scissor)))

	// Set topology
	syscall.SyscallN(comVtbl(b.context, ctxIASetPrimitiveTopology),
		b.context, D3D11_PRIMITIVE_TOPOLOGY_TRIANGLELIST)

	// Clear render target
	clearColor := [4]float32{0, 0, 0, 1}
	syscall.SyscallN(comVtbl(b.context, ctxClearRenderTargetView),
		b.context, b.renderTarget,
		uintptr(unsafe.Pointer(&clearColor[0])))
}

// EndFrame implements render.Backend.
func (b *Backend) EndFrame() {
	// IDXGISwapChain::Present(0, 0) — 0 = no vsync
	comCall(comVtbl(b.swapChain, scPresent), b.swapChain, 0, 0)
}

// Resize implements render.Backend.
func (b *Backend) Resize(width, height int) {
	if width <= 0 || height <= 0 {
		return
	}
	b.width = width
	b.height = height

	// Release old render target
	if b.renderTarget != 0 {
		comRelease(b.renderTarget)
		b.renderTarget = 0
	}

	// IDXGISwapChain::ResizeBuffers(BufferCount, Width, Height, NewFormat, Flags)
	comCall(comVtbl(b.swapChain, scResizeBuffers),
		b.swapChain, 0,
		uintptr(width), uintptr(height),
		DXGI_FORMAT_R8G8B8A8_UNORM, 0)

	// Recreate render target
	b.createRenderTarget()
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
	for _, c := range commands {
		switch c.Type {
		case render.CmdClip:
			if c.Clip != nil {
				b.applyScissor(c.Clip)
			}
		case render.CmdRect:
			if c.Rect != nil {
				b.renderRect(c)
			}
		case render.CmdText:
			if c.Text != nil {
				b.renderText(c)
			}
		case render.CmdImage:
			if c.Image != nil {
				b.renderImage(c)
			}
		case render.CmdShadow:
			if c.Shadow != nil {
				b.renderShadow(c)
			}
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

	scissor := D3D11_RECT{Left: x, Top: y, Right: right, Bottom: bottom}
	syscall.SyscallN(comVtbl(b.context, ctxRSSetScissorRects),
		b.context, 1,
		uintptr(unsafe.Pointer(&scissor)))
}

func (b *Backend) renderRect(c render.Command) {
	rect := c.Rect
	opacity := c.Opacity

	logW := float32(b.width) / b.dpiScale
	logH := float32(b.height) / b.dpiScale
	x, y, w, h := rect.Bounds.X, rect.Bounds.Y, rect.Bounds.Width, rect.Bounds.Height

	// Padding for anti-aliasing
	pad := float32(1.0)
	qx, qy := x-pad, y-pad
	qw, qh := w+pad*2, h+pad*2

	// Convert to NDC: DX11 NDC is X [-1,1] left-to-right, Y [1,-1] top-to-bottom in clip space
	// But SV_POSITION after vertex shader is in clip space where Y+ is up.
	// Screen Y increases downward, so we flip Y.
	ndcX := (qx/logW)*2 - 1
	ndcY := 1 - (qy/logH)*2 // flip Y for DX11
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
		makeVertex(ndcX, ndcY, uvL, uvT),                 // top-left
		makeVertex(ndcX+ndcW, ndcY, uvR, uvT),            // top-right
		makeVertex(ndcX+ndcW, ndcY-ndcH, uvR, uvB),       // bottom-right
		makeVertex(ndcX, ndcY, uvL, uvT),                  // top-left
		makeVertex(ndcX+ndcW, ndcY-ndcH, uvR, uvB),       // bottom-right
		makeVertex(ndcX, ndcY-ndcH, uvL, uvB),             // bottom-left
	}

	stride := uint32(unsafe.Sizeof(RectVertex{}))
	data := unsafe.Slice((*byte)(unsafe.Pointer(&vertices[0])), 6*int(stride))
	b.uploadAndDraw(data, stride, 6, b.rectVS, b.rectPS, b.rectInputLayout, 0)
}

func (b *Backend) renderShadow(c render.Command) {
	sh := c.Shadow
	opacity := c.Opacity

	logW := float32(b.width) / b.dpiScale
	logH := float32(b.height) / b.dpiScale
	s := b.dpiScale

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

	ndcX := (qx/logW)*2 - 1
	ndcY := 1 - (qy/logH)*2
	ndcW := (qw / logW) * 2
	ndcH := (qh / logH) * 2

	uvL := -expand / elemW
	uvT := -expand / elemH
	uvR := 1.0 + expand/elemW
	uvB := 1.0 + expand/elemH

	maxR := min32dx(elemW, elemH) * 0.5
	radTL := clamp32dx(sh.Corners.TopLeft+spread, 0, maxR) * s
	radTR := clamp32dx(sh.Corners.TopRight+spread, 0, maxR) * s
	radBR := clamp32dx(sh.Corners.BottomRight+spread, 0, maxR) * s
	radBL := clamp32dx(sh.Corners.BottomLeft+spread, 0, maxR) * s

	col := sh.Color
	a := col.A * opacity

	makeVertex := func(px, py, u, v float32) ShadowVertex {
		return ShadowVertex{
			PosX: px, PosY: py, U: u, V: v,
			ColorR: col.R, ColorG: col.G, ColorB: col.B, ColorA: a,
			ElemW: elemW * s, ElemH: elemH * s,
			RadiusTL: radTL, RadiusTR: radTR, RadiusBR: radBR, RadiusBL: radBL,
			Blur: blur * s,
		}
	}
	vertices := [6]ShadowVertex{
		makeVertex(ndcX, ndcY, uvL, uvT),
		makeVertex(ndcX+ndcW, ndcY, uvR, uvT),
		makeVertex(ndcX+ndcW, ndcY-ndcH, uvR, uvB),
		makeVertex(ndcX, ndcY, uvL, uvT),
		makeVertex(ndcX+ndcW, ndcY-ndcH, uvR, uvB),
		makeVertex(ndcX, ndcY-ndcH, uvL, uvB),
	}
	stride := uint32(unsafe.Sizeof(ShadowVertex{}))
	data := unsafe.Slice((*byte)(unsafe.Pointer(&vertices[0])), 6*int(stride))
	b.uploadAndDraw(data, stride, 6, b.shadowVS, b.shadowPS, b.shadowInputLayout, 0)
}

func min32dx(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func clamp32dx(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func (b *Backend) renderText(c render.Command) {
	tc := c.Text
	entry, ok := b.textures[tc.Atlas]
	if !ok {
		return
	}

	logW := float32(b.width) / b.dpiScale
	logH := float32(b.height) / b.dpiScale

	glyphs := tc.Glyphs
	if len(glyphs) == 0 {
		return
	}

	vertices := make([]TexturedVertex, 0, len(glyphs)*6)
	for _, g := range glyphs {
		x0 := (g.X/logW)*2 - 1
		y0 := 1 - (g.Y/logH)*2 // flip Y
		x1 := ((g.X + g.Width) / logW) * 2 - 1
		y1 := 1 - ((g.Y+g.Height)/logH)*2 // flip Y

		clr := TexturedVertex{
			ColorR: tc.Color.R, ColorG: tc.Color.G,
			ColorB: tc.Color.B, ColorA: tc.Color.A * c.Opacity,
		}

		v0 := clr
		v0.PosX = x0
		v0.PosY = y0
		v0.U = g.U0
		v0.V = g.V0
		v1 := clr
		v1.PosX = x1
		v1.PosY = y0
		v1.U = g.U1
		v1.V = g.V0
		v2 := clr
		v2.PosX = x1
		v2.PosY = y1
		v2.U = g.U1
		v2.V = g.V1
		v3 := clr
		v3.PosX = x0
		v3.PosY = y1
		v3.U = g.U0
		v3.V = g.V1

		vertices = append(vertices, v0, v1, v2, v0, v2, v3)
	}

	stride := uint32(unsafe.Sizeof(TexturedVertex{}))
	data := unsafe.Slice((*byte)(unsafe.Pointer(&vertices[0])), len(vertices)*int(stride))
	b.uploadAndDraw(data, stride, uint32(len(vertices)), b.texturedVS, b.textPS, b.texturedInputLayout, entry.srv)
}

func (b *Backend) renderImage(c render.Command) {
	ic := c.Image
	entry, ok := b.textures[ic.Texture]
	if !ok {
		return
	}

	logW := float32(b.width) / b.dpiScale
	logH := float32(b.height) / b.dpiScale

	x0 := (ic.DstRect.X/logW)*2 - 1
	y0 := 1 - (ic.DstRect.Y/logH)*2 // flip Y
	x1 := ((ic.DstRect.X + ic.DstRect.Width) / logW) * 2 - 1
	y1 := 1 - ((ic.DstRect.Y+ic.DstRect.Height)/logH)*2 // flip Y

	// UV from source rect; default to full texture when empty
	srcRect := ic.SrcRect
	if srcRect.Width == 0 || srcRect.Height == 0 {
		srcRect = uimath.NewRect(0, 0, 1, 1)
	}
	u0, v0 := srcRect.X, srcRect.Y
	u1 := srcRect.X + srcRect.Width
	v1 := srcRect.Y + srcRect.Height

	tint := ic.Tint
	if tint.A == 0 && tint.R == 0 && tint.G == 0 && tint.B == 0 {
		tint = uimath.Color{R: 1, G: 1, B: 1, A: c.Opacity}
	} else {
		tint.A *= c.Opacity
	}

	// SDF rounded corner data (physical pixels)
	s := b.dpiScale
	rw := ic.DstRect.Width * s
	rh := ic.DstRect.Height * s
	rtl := ic.Corners.TopLeft * s
	rtr := ic.Corners.TopRight * s
	rbr := ic.Corners.BottomRight * s
	rbl := ic.Corners.BottomLeft * s

	makeVertex := func(px, py, u, v float32) TexturedVertex {
		return TexturedVertex{
			PosX: px, PosY: py, U: u, V: v,
			ColorR: tint.R, ColorG: tint.G, ColorB: tint.B, ColorA: tint.A,
			RectW: rw, RectH: rh,
			RadiusTL: rtl, RadiusTR: rtr, RadiusBR: rbr, RadiusBL: rbl,
		}
	}

	vertices := [6]TexturedVertex{
		makeVertex(x0, y0, u0, v0),
		makeVertex(x1, y0, u1, v0),
		makeVertex(x1, y1, u1, v1),
		makeVertex(x0, y0, u0, v0),
		makeVertex(x1, y1, u1, v1),
		makeVertex(x0, y1, u0, v1),
	}

	stride := uint32(unsafe.Sizeof(TexturedVertex{}))
	data := unsafe.Slice((*byte)(unsafe.Pointer(&vertices[0])), 6*int(stride))
	b.uploadAndDraw(data, stride, 6, b.texturedVS, b.texturedPS, b.texturedInputLayout, entry.srv)
}

func (b *Backend) uploadAndDraw(data []byte, stride, vertexCount uint32, vs, ps, il, srv uintptr) {
	dataSize := uint32(len(data))

	// Grow buffer if needed
	if dataSize > b.vertexBufSize {
		if b.vertexBuffer != 0 {
			comRelease(b.vertexBuffer)
			b.vertexBuffer = 0
		}
		b.vertexBufSize = dataSize * 2
		if err := b.createVertexBuffer(b.vertexBufSize); err != nil {
			return
		}
	}

	// Map buffer
	var mapped D3D11_MAPPED_SUBRESOURCE
	hr := comCall(comVtbl(b.context, ctxMap),
		b.context, b.vertexBuffer, 0,
		D3D11_MAP_WRITE_DISCARD, 0,
		uintptr(unsafe.Pointer(&mapped)))
	if hr != S_OK {
		return
	}
	copy(unsafe.Slice((*byte)(mapped.PData), dataSize), data)
	syscall.SyscallN(comVtbl(b.context, ctxUnmap),
		b.context, b.vertexBuffer, 0)

	// Bind shaders
	syscall.SyscallN(comVtbl(b.context, ctxVSSetShader),
		b.context, vs, 0, 0)
	syscall.SyscallN(comVtbl(b.context, ctxPSSetShader),
		b.context, ps, 0, 0)

	// Bind input layout
	syscall.SyscallN(comVtbl(b.context, ctxIASetInputLayout),
		b.context, il)

	// Bind vertex buffer
	offset := uint32(0)
	syscall.SyscallN(comVtbl(b.context, ctxIASetVertexBuffers),
		b.context, 0, 1,
		uintptr(unsafe.Pointer(&b.vertexBuffer)),
		uintptr(unsafe.Pointer(&stride)),
		uintptr(unsafe.Pointer(&offset)))

	// Bind texture if present
	if srv != 0 {
		syscall.SyscallN(comVtbl(b.context, ctxPSSetShaderResources),
			b.context, 0, 1,
			uintptr(unsafe.Pointer(&srv)))
		syscall.SyscallN(comVtbl(b.context, ctxPSSetSamplers),
			b.context, 0, 1,
			uintptr(unsafe.Pointer(&b.linearSampler)))
	}

	// Draw
	syscall.SyscallN(comVtbl(b.context, ctxDraw),
		b.context, uintptr(vertexCount), 0)
}

// CreateTexture implements render.Backend.
func (b *Backend) CreateTexture(desc render.TextureDesc) (render.TextureHandle, error) {
	format := uint32(DXGI_FORMAT_R8G8B8A8_UNORM)
	bytesPerPixel := 4
	switch desc.Format {
	case render.TextureFormatR8:
		format = DXGI_FORMAT_R8_UNORM
		bytesPerPixel = 1
	case render.TextureFormatRGBA8:
		format = DXGI_FORMAT_R8G8B8A8_UNORM
	case render.TextureFormatBGRA8:
		format = DXGI_FORMAT_B8G8R8A8_UNORM
	}

	texDesc := D3D11_TEXTURE2D_DESC{
		Width:      uint32(desc.Width),
		Height:     uint32(desc.Height),
		MipLevels:  1,
		ArraySize:  1,
		Format:     format,
		SampleDesc: DXGI_SAMPLE_DESC{Count: 1},
		Usage:      D3D11_USAGE_DEFAULT,
		BindFlags:  D3D11_BIND_SHADER_RESOURCE,
	}

	var initDataPtr uintptr
	var sd D3D11_SUBRESOURCE_DATA
	if desc.Data != nil {
		sd.PSysMem = unsafe.Pointer(&desc.Data[0])
		sd.SysMemPitch = uint32(desc.Width * bytesPerPixel)
		initDataPtr = uintptr(unsafe.Pointer(&sd))
	}

	var texture uintptr
	hr := comCall(comVtbl(b.device, devCreateTexture2D),
		b.device,
		uintptr(unsafe.Pointer(&texDesc)),
		initDataPtr,
		uintptr(unsafe.Pointer(&texture)))
	if hr != S_OK {
		return render.InvalidTexture, fmt.Errorf("dx11: CreateTexture2D failed: 0x%x", hr)
	}

	// Create shader resource view
	srvDesc := D3D11_SHADER_RESOURCE_VIEW_DESC{
		Format:        format,
		ViewDimension: 4, // D3D11_SRV_DIMENSION_TEXTURE2D
	}
	srvDesc.Texture2D.MipLevels = 1

	var srv uintptr
	hr = comCall(comVtbl(b.device, devCreateShaderResourceView),
		b.device, texture,
		uintptr(unsafe.Pointer(&srvDesc)),
		uintptr(unsafe.Pointer(&srv)))
	if hr != S_OK {
		comRelease(texture)
		return render.InvalidTexture, fmt.Errorf("dx11: CreateShaderResourceView failed: 0x%x", hr)
	}

	handle := b.nextTextureID
	b.nextTextureID++
	b.textures[handle] = &textureEntry{
		texture: texture,
		srv:     srv,
		width:   desc.Width,
		height:  desc.Height,
		format:  desc.Format,
	}

	return handle, nil
}

// UpdateTexture implements render.Backend.
func (b *Backend) UpdateTexture(handle render.TextureHandle, region uimath.Rect, data []byte) error {
	entry, ok := b.textures[handle]
	if !ok {
		return fmt.Errorf("dx11: texture %d not found", handle)
	}

	bytesPerPixel := 4
	if entry.format == render.TextureFormatR8 {
		bytesPerPixel = 1
	}

	box := D3D11_BOX{
		Left:   uint32(region.X),
		Top:    uint32(region.Y),
		Right:  uint32(region.X + region.Width),
		Bottom: uint32(region.Y + region.Height),
		Front:  0,
		Back:   1,
	}

	// ID3D11DeviceContext::UpdateSubresource
	syscall.SyscallN(comVtbl(b.context, ctxUpdateSubresource),
		b.context,
		entry.texture, 0,
		uintptr(unsafe.Pointer(&box)),
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(uint32(int(region.Width)*bytesPerPixel)),
		0)

	return nil
}

// DestroyTexture implements render.Backend.
func (b *Backend) DestroyTexture(handle render.TextureHandle) {
	entry, ok := b.textures[handle]
	if !ok {
		return
	}
	if entry.srv != 0 {
		comRelease(entry.srv)
	}
	if entry.texture != 0 {
		comRelease(entry.texture)
	}
	delete(b.textures, handle)
}

// MaxTextureSize implements render.Backend.
func (b *Backend) MaxTextureSize() int {
	return 16384 // D3D11 maximum texture dimension
}

// DPIScale implements render.Backend.
func (b *Backend) DPIScale() float32 {
	return b.dpiScale
}

// ReadPixels implements render.Backend.
func (b *Backend) ReadPixels() (*image.RGBA, error) {
	if b.width <= 0 || b.height <= 0 {
		return nil, fmt.Errorf("dx11: invalid dimensions")
	}

	// Get back buffer
	var backBuffer uintptr
	iidTexture2D := GUID{
		0x6f15aaf2, 0xd208, 0x4e89,
		[8]byte{0x9a, 0xb4, 0x48, 0x95, 0x35, 0xd3, 0x4f, 0x9c},
	}
	hr := comCall(comVtbl(b.swapChain, scGetBuffer),
		b.swapChain, 0,
		uintptr(unsafe.Pointer(&iidTexture2D)),
		uintptr(unsafe.Pointer(&backBuffer)))
	if hr != S_OK {
		return nil, fmt.Errorf("dx11: GetBuffer failed for ReadPixels: 0x%x", hr)
	}
	defer comRelease(backBuffer)

	// Create staging texture for CPU readback
	stagingDesc := D3D11_TEXTURE2D_DESC{
		Width:          uint32(b.width),
		Height:         uint32(b.height),
		MipLevels:      1,
		ArraySize:      1,
		Format:         DXGI_FORMAT_R8G8B8A8_UNORM,
		SampleDesc:     DXGI_SAMPLE_DESC{Count: 1},
		Usage:          D3D11_USAGE_STAGING,
		CPUAccessFlags: D3D11_CPU_ACCESS_READ,
	}

	var staging uintptr
	hr = comCall(comVtbl(b.device, devCreateTexture2D),
		b.device,
		uintptr(unsafe.Pointer(&stagingDesc)),
		0,
		uintptr(unsafe.Pointer(&staging)))
	if hr != S_OK {
		return nil, fmt.Errorf("dx11: staging texture failed: 0x%x", hr)
	}
	defer comRelease(staging)

	// CopyResource(dst, src)
	syscall.SyscallN(comVtbl(b.context, ctxCopyResource),
		b.context, staging, backBuffer)

	// Map the staging texture
	var mapped D3D11_MAPPED_SUBRESOURCE
	hr = comCall(comVtbl(b.context, ctxMap),
		b.context, staging, 0,
		D3D11_MAP_READ, 0,
		uintptr(unsafe.Pointer(&mapped)))
	if hr != S_OK {
		return nil, fmt.Errorf("dx11: map staging failed: 0x%x", hr)
	}

	img := image.NewRGBA(image.Rect(0, 0, b.width, b.height))
	for y := 0; y < b.height; y++ {
		src := unsafe.Add(mapped.PData, y*int(mapped.RowPitch))
		srcSlice := unsafe.Slice((*byte)(src), b.width*4)
		dstOff := y * img.Stride
		copy(img.Pix[dstOff:dstOff+b.width*4], srcSlice)
	}

	syscall.SyscallN(comVtbl(b.context, ctxUnmap),
		b.context, staging, 0)

	return img, nil
}

// Destroy implements render.Backend.
func (b *Backend) Destroy() {
	for handle := range b.textures {
		b.DestroyTexture(handle)
	}

	if b.vertexBuffer != 0 {
		comRelease(b.vertexBuffer)
	}
	if b.linearSampler != 0 {
		comRelease(b.linearSampler)
	}
	if b.textPS != 0 {
		comRelease(b.textPS)
	}
	if b.texturedPS != 0 {
		comRelease(b.texturedPS)
	}
	if b.texturedInputLayout != 0 {
		comRelease(b.texturedInputLayout)
	}
	if b.texturedVS != 0 {
		comRelease(b.texturedVS)
	}
	if b.rectPS != 0 {
		comRelease(b.rectPS)
	}
	if b.rectInputLayout != 0 {
		comRelease(b.rectInputLayout)
	}
	if b.rectVS != 0 {
		comRelease(b.rectVS)
	}
	if b.shadowPS != 0 {
		comRelease(b.shadowPS)
	}
	if b.shadowInputLayout != 0 {
		comRelease(b.shadowInputLayout)
	}
	if b.shadowVS != 0 {
		comRelease(b.shadowVS)
	}
	if b.rasterizerState != 0 {
		comRelease(b.rasterizerState)
	}
	if b.blendState != 0 {
		comRelease(b.blendState)
	}
	if b.renderTarget != 0 {
		comRelease(b.renderTarget)
	}
	if b.swapChain != 0 {
		comRelease(b.swapChain)
	}
	if b.context != 0 {
		comRelease(b.context)
	}
	if b.device != 0 {
		comRelease(b.device)
	}
}

// Verify Backend implements render.Backend at compile time.
var _ render.Backend = (*Backend)(nil)

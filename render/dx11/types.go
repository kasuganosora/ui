//go:build windows

package dx11

import "unsafe"

// COM result codes
const (
	S_OK    = 0
	S_FALSE = 1
	E_FAIL  = 0x80004005
)

// DXGI formats
const (
	DXGI_FORMAT_R8G8B8A8_UNORM      = 28
	DXGI_FORMAT_R8G8B8A8_UNORM_SRGB = 29
	DXGI_FORMAT_B8G8R8A8_UNORM      = 87
	DXGI_FORMAT_R8_UNORM            = 61
	DXGI_FORMAT_R32G32_FLOAT        = 16
	DXGI_FORMAT_R32G32B32_FLOAT     = 6
	DXGI_FORMAT_R32G32B32A32_FLOAT  = 2
	DXGI_FORMAT_R32_FLOAT           = 41
)

// D3D11 bind flags
const (
	D3D11_BIND_VERTEX_BUFFER   = 0x1
	D3D11_BIND_INDEX_BUFFER    = 0x2
	D3D11_BIND_CONSTANT_BUFFER = 0x4
	D3D11_BIND_SHADER_RESOURCE = 0x8
	D3D11_BIND_RENDER_TARGET   = 0x20
)

// D3D11 usage
const (
	D3D11_USAGE_DEFAULT = 0
	D3D11_USAGE_DYNAMIC = 2
	D3D11_USAGE_STAGING = 3
)

// D3D11 CPU access
const (
	D3D11_CPU_ACCESS_WRITE = 0x10000
	D3D11_CPU_ACCESS_READ  = 0x20000
)

// D3D11 map types
const (
	D3D11_MAP_READ          = 1
	D3D11_MAP_WRITE_DISCARD = 4
)

// D3D11 primitive topology
const (
	D3D11_PRIMITIVE_TOPOLOGY_TRIANGLELIST = 4
)

// D3D11 filter
const (
	D3D11_FILTER_MIN_MAG_MIP_POINT       = 0
	D3D11_FILTER_MIN_MAG_LINEAR_MIP_POINT = 0x14
	D3D11_FILTER_MIN_MAG_MIP_LINEAR      = 0x15
)

// D3D11 texture address mode
const (
	D3D11_TEXTURE_ADDRESS_CLAMP = 3
)

// D3D11 comparison func
const (
	D3D11_COMPARISON_NEVER = 1
)

// D3D11 fill mode
const (
	D3D11_FILL_SOLID = 3
)

// D3D11 cull mode
const (
	D3D11_CULL_NONE = 1
)

// DXGI swap effect
const (
	DXGI_SWAP_EFFECT_DISCARD      = 0
	DXGI_SWAP_EFFECT_FLIP_DISCARD = 4
)

// D3D11 blend
const (
	D3D11_BLEND_ZERO          = 1
	D3D11_BLEND_ONE           = 2
	D3D11_BLEND_SRC_ALPHA     = 5
	D3D11_BLEND_INV_SRC_ALPHA = 6
	D3D11_BLEND_OP_ADD        = 1
)

// D3D11 color write enable
const (
	D3D11_COLOR_WRITE_ENABLE_ALL = 0xF
)

// D3D feature level
const (
	D3D_FEATURE_LEVEL_10_0 = 0xa000
	D3D_FEATURE_LEVEL_10_1 = 0xa100
	D3D_FEATURE_LEVEL_11_0 = 0xb000
)

// D3D driver type
const (
	D3D_DRIVER_TYPE_HARDWARE = 1
)

// D3D11 create device flags
const (
	D3D11_CREATE_DEVICE_DEBUG = 0x2
)

// DXGI usage
const (
	DXGI_USAGE_RENDER_TARGET_OUTPUT = 0x20
)

// D3D11 SDK version
const (
	D3D11_SDK_VERSION = 7
)

// GUID is a COM interface identifier.
type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

// Structs matching D3D11 layout

// DXGI_SWAP_CHAIN_DESC describes a swap chain.
type DXGI_SWAP_CHAIN_DESC struct {
	BufferDesc   DXGI_MODE_DESC
	SampleDesc   DXGI_SAMPLE_DESC
	BufferUsage  uint32
	BufferCount  uint32
	OutputWindow uintptr // HWND
	Windowed     int32
	SwapEffect   uint32
	Flags        uint32
}

// DXGI_MODE_DESC describes a display mode.
type DXGI_MODE_DESC struct {
	Width            uint32
	Height           uint32
	RefreshRate      DXGI_RATIONAL
	Format           uint32
	ScanlineOrdering uint32
	Scaling          uint32
}

// DXGI_RATIONAL represents a rational number (fraction).
type DXGI_RATIONAL struct {
	Numerator   uint32
	Denominator uint32
}

// DXGI_SAMPLE_DESC describes multi-sampling parameters.
type DXGI_SAMPLE_DESC struct {
	Count   uint32
	Quality uint32
}

// D3D11_BUFFER_DESC describes a buffer resource.
type D3D11_BUFFER_DESC struct {
	ByteWidth           uint32
	Usage               uint32
	BindFlags           uint32
	CPUAccessFlags      uint32
	MiscFlags           uint32
	StructureByteStride uint32
}

// D3D11_SUBRESOURCE_DATA provides data for initializing a subresource.
type D3D11_SUBRESOURCE_DATA struct {
	PSysMem          unsafe.Pointer
	SysMemPitch      uint32
	SysMemSlicePitch uint32
}

// D3D11_MAPPED_SUBRESOURCE provides access to a mapped resource.
type D3D11_MAPPED_SUBRESOURCE struct {
	PData      unsafe.Pointer
	RowPitch   uint32
	DepthPitch uint32
}

// D3D11_TEXTURE2D_DESC describes a 2D texture resource.
type D3D11_TEXTURE2D_DESC struct {
	Width          uint32
	Height         uint32
	MipLevels      uint32
	ArraySize      uint32
	Format         uint32
	SampleDesc     DXGI_SAMPLE_DESC
	Usage          uint32
	BindFlags      uint32
	CPUAccessFlags uint32
	MiscFlags      uint32
}

// D3D11_SHADER_RESOURCE_VIEW_DESC describes a shader resource view.
type D3D11_SHADER_RESOURCE_VIEW_DESC struct {
	Format        uint32
	ViewDimension uint32 // D3D11_SRV_DIMENSION_TEXTURE2D = 4
	Texture2D     struct {
		MostDetailedMip uint32
		MipLevels       uint32
	}
	_pad [8]byte // union padding
}

// D3D11_SAMPLER_DESC describes a sampler state.
type D3D11_SAMPLER_DESC struct {
	Filter         uint32
	AddressU       uint32
	AddressV       uint32
	AddressW       uint32
	MipLODBias     float32
	MaxAnisotropy  uint32
	ComparisonFunc uint32
	BorderColor    [4]float32
	MinLOD         float32
	MaxLOD         float32
}

// D3D11_BLEND_DESC describes the blend state.
type D3D11_BLEND_DESC struct {
	AlphaToCoverageEnable  int32
	IndependentBlendEnable int32
	RenderTarget           [8]D3D11_RENDER_TARGET_BLEND_DESC
}

// D3D11_RENDER_TARGET_BLEND_DESC describes the blend state for a render target.
type D3D11_RENDER_TARGET_BLEND_DESC struct {
	BlendEnable           int32
	SrcBlend              uint32
	DestBlend             uint32
	BlendOp               uint32
	SrcBlendAlpha         uint32
	DestBlendAlpha        uint32
	BlendOpAlpha          uint32
	RenderTargetWriteMask uint8
	_pad                  [3]byte
}

// D3D11_RASTERIZER_DESC describes rasterizer state.
type D3D11_RASTERIZER_DESC struct {
	FillMode              uint32
	CullMode              uint32
	FrontCounterClockwise int32
	DepthBias             int32
	DepthBiasClamp        float32
	SlopeScaledDepthBias  float32
	DepthClipEnable       int32
	ScissorEnable         int32
	MultisampleEnable     int32
	AntialiasedLineEnable int32
}

// D3D11_VIEWPORT describes a viewport.
type D3D11_VIEWPORT struct {
	TopLeftX float32
	TopLeftY float32
	Width    float32
	Height   float32
	MinDepth float32
	MaxDepth float32
}

// D3D11_RECT defines a rectangle (LONG coordinates).
type D3D11_RECT struct {
	Left, Top, Right, Bottom int32
}

// D3D11_INPUT_ELEMENT_DESC describes a single element in a vertex input layout.
type D3D11_INPUT_ELEMENT_DESC struct {
	SemanticName         *byte
	SemanticIndex        uint32
	Format               uint32
	InputSlot            uint32
	AlignedByteOffset    uint32
	InputSlotClass       uint32
	InstanceDataStepRate uint32
}

// D3D11_BOX defines a 3D box region for UpdateSubresource.
type D3D11_BOX struct {
	Left, Top, Front    uint32
	Right, Bottom, Back uint32
}

// D3D11_RENDER_TARGET_VIEW_DESC describes a render target view.
type D3D11_RENDER_TARGET_VIEW_DESC struct {
	Format        uint32
	ViewDimension uint32 // D3D11_RTV_DIMENSION_TEXTURE2D = 4
	Texture2D     struct {
		MipSlice uint32
	}
	_pad [8]byte // union padding for largest member
}

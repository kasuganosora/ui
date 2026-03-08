//go:build windows

package dx9

import "unsafe"

// COM result codes
const (
	S_OK                              = 0
	D3D_OK                            = 0
	D3DERR_DEVICELOST      uintptr   = 0x88760868
	D3DERR_DEVICENOTRESET  uintptr   = 0x88760869
	D3DERR_NOTAVAILABLE    uintptr   = 0x8876086A
)

// D3DFORMAT constants
const (
	D3DFMT_UNKNOWN       = 0
	D3DFMT_R8G8B8        = 20
	D3DFMT_A8R8G8B8      = 21 // BGRA in memory
	D3DFMT_X8R8G8B8      = 22
	D3DFMT_A8            = 28 // Single-channel alpha
	D3DFMT_L8            = 50 // Single-channel luminance
	D3DFMT_A8L8          = 51
	D3DFMT_D16           = 80
	D3DFMT_D24S8         = 75
	D3DFMT_INDEX16       = 101
)

// D3DPOOL constants
const (
	D3DPOOL_DEFAULT   = 0
	D3DPOOL_MANAGED   = 1
	D3DPOOL_SYSTEMMEM = 2
)

// D3DUSAGE constants
const (
	D3DUSAGE_DYNAMIC     = 0x200
	D3DUSAGE_WRITEONLY   = 0x8
	D3DUSAGE_SOFTWAREPROCESSING = 0x10
)

// D3DRS (render state) constants
const (
	D3DRS_ZENABLE          = 7
	D3DRS_LIGHTING         = 137
	D3DRS_CULLMODE         = 22
	D3DRS_ALPHABLENDENABLE = 27
	D3DRS_SRCBLEND         = 19
	D3DRS_DESTBLEND        = 20
	D3DRS_BLENDOP          = 171
	D3DRS_SEPARATEALPHABLENDENABLE = 206
	D3DRS_SRCBLENDALPHA    = 207
	D3DRS_DESTBLENDALPHA   = 208
	D3DRS_BLENDOPALPHA     = 209
	D3DRS_SCISSORTESTENABLE = 174
	D3DRS_SRGBWRITEENABLE  = 194
	D3DRS_COLORWRITEENABLE = 168
)

// D3DBLEND constants
const (
	D3DBLEND_ZERO         = 1
	D3DBLEND_ONE          = 2
	D3DBLEND_SRCALPHA     = 5
	D3DBLEND_INVSRCALPHA  = 6
	D3DBLEND_DESTALPHA    = 7
	D3DBLEND_INVDESTALPHA = 8
)

// D3DBLENDOP constants
const (
	D3DBLENDOP_ADD = 1
)

// D3DCULL constants
const (
	D3DCULL_NONE = 1
	D3DCULL_CW   = 2
	D3DCULL_CCW  = 3
)

// D3DPRIMITIVETYPE constants
const (
	D3DPT_TRIANGLELIST = 4
)

// D3DLOCK constants
const (
	D3DLOCK_DISCARD = 0x2000
)

// D3DFVF (Flexible Vertex Format) — not used with shaders, but needed for CreateVertexBuffer
const (
	D3DFVF_XYZ    = 0x002
	D3DFVF_XYZRHW = 0x004
	D3DFVF_TEX1   = 0x100
)

// D3DDEVTYPE constants
const (
	D3DDEVTYPE_HAL = 1
	D3DDEVTYPE_REF = 2
)

// D3DCREATE constants
const (
	D3DCREATE_SOFTWARE_VERTEXPROCESSING = 0x20
	D3DCREATE_HARDWARE_VERTEXPROCESSING = 0x40
	D3DCREATE_MIXED_VERTEXPROCESSING    = 0x80
	D3DCREATE_FPU_PRESERVE             = 0x2
)

// D3DSWAPEFFECT constants
const (
	D3DSWAPEFFECT_DISCARD = 1
)

// D3DMULTISAMPLE_TYPE
const (
	D3DMULTISAMPLE_NONE = 0
)

// D3DTSS (texture stage state)
const (
	D3DTSS_COLOROP   = 1
	D3DTSS_COLORARG1 = 2
	D3DTSS_ALPHAOP   = 4
	D3DTSS_ALPHAARG1 = 5
)

// D3DTOP (texture operation)
const (
	D3DTOP_SELECTARG1 = 2
)

// D3DTA (texture argument)
const (
	D3DTA_TEXTURE = 2
)

// D3DSAMP (sampler state)
const (
	D3DSAMP_MINFILTER = 5
	D3DSAMP_MAGFILTER = 6
	D3DSAMP_MIPFILTER = 7
	D3DSAMP_ADDRESSU  = 1
	D3DSAMP_ADDRESSV  = 2
	D3DSAMP_SRGBTEXTURE = 13
)

// D3DTEXF (texture filter)
const (
	D3DTEXF_NONE   = 0
	D3DTEXF_POINT  = 1
	D3DTEXF_LINEAR = 2
)

// D3DTADDRESS (texture address mode)
const (
	D3DTADDRESS_CLAMP = 3
)

// D3DDECLUSAGE for vertex declarations
const (
	D3DDECLUSAGE_POSITION    = 0
	D3DDECLUSAGE_TEXCOORD    = 5
	D3DDECLUSAGE_COLOR       = 10
)

// D3DDECLTYPE for vertex declarations
const (
	D3DDECLTYPE_FLOAT2  = 1 // 2 floats
	D3DDECLTYPE_FLOAT4  = 3 // 4 floats
	D3DDECLTYPE_D3DCOLOR = 4 // DWORD color (not used, we use float4)
)

// D3DDECLMETHOD
const (
	D3DDECLMETHOD_DEFAULT = 0
)

// D3DPRESENT_PARAMETERS describes the presentation parameters.
type D3DPRESENT_PARAMETERS struct {
	BackBufferWidth            uint32
	BackBufferHeight           uint32
	BackBufferFormat           uint32
	BackBufferCount            uint32
	MultiSampleType            uint32
	MultiSampleQuality         uint32
	SwapEffect                 uint32
	HDeviceWindow              uintptr
	Windowed                   int32
	EnableAutoDepthStencil     int32
	AutoDepthStencilFormat     uint32
	Flags                      uint32
	FullScreen_RefreshRateInHz uint32
	PresentationInterval       uint32
}

// D3DVERTEXELEMENT9 describes a vertex declaration element.
type D3DVERTEXELEMENT9 struct {
	Stream     uint16
	Offset     uint16
	Type       byte
	Method     byte
	Usage      byte
	UsageIndex byte
}

// D3DVERTEXELEMENT9_END marks the end of a vertex element array.
var D3DVERTEXELEMENT9_END = D3DVERTEXELEMENT9{
	Stream: 0xFF,
	Offset: 0,
	Type:   17, // D3DDECLTYPE_UNUSED
	Method: 0,
	Usage:  0,
	UsageIndex: 0,
}

// D3DLOCKED_RECT provides access to locked texture data.
type D3DLOCKED_RECT struct {
	Pitch int32
	PBits unsafe.Pointer
}

// RECT for scissor test.
type RECT struct {
	Left, Top, Right, Bottom int32
}

// D3DVIEWPORT9 describes a viewport.
type D3DVIEWPORT9 struct {
	X, Y          uint32
	Width, Height  uint32
	MinZ, MaxZ     float32
}

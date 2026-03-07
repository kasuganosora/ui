package vulkan

import "unsafe"

// Vulkan handles - opaque pointers or uint64 depending on platform.
type (
	Instance       uintptr
	PhysicalDevice uintptr
	Device         uintptr
	Queue          uintptr
	CommandPool    uintptr
	CommandBuffer  uintptr
	RenderPass     uintptr
	Framebuffer    uintptr
	Pipeline       uintptr
	PipelineLayout uintptr
	PipelineCache  uintptr
	ShaderModule   uintptr
	Buffer         uintptr
	DeviceMemory   uintptr
	Image          uintptr
	ImageView      uintptr
	Sampler        uintptr
	Fence          uintptr
	Semaphore      uintptr
	DescriptorPool       uintptr
	DescriptorSet        uintptr
	DescriptorSetLayout  uintptr
	SurfaceKHR           uintptr
	SwapchainKHR         uintptr
)

// Result codes
type Result int32

const (
	Success                  Result = 0
	NotReady                 Result = 1
	Timeout                  Result = 2
	Incomplete               Result = 5
	ErrorOutOfHostMemory     Result = -1
	ErrorOutOfDeviceMemory   Result = -2
	ErrorInitializationFailed Result = -3
	ErrorDeviceLost          Result = -4
	ErrorMemoryMapFailed     Result = -5
	ErrorLayerNotPresent     Result = -6
	ErrorExtensionNotPresent Result = -7
	ErrorFeatureNotPresent   Result = -8
	ErrorTooManyObjects      Result = -10
	ErrorFormatNotSupported  Result = -11
	ErrorSurfaceLostKHR      Result = -1000000000
	ErrorOutOfDateKHR        Result = -1000001004
	SuboptimalKHR            Result = 1000001003
)

func (r Result) Error() string {
	switch r {
	case Success:
		return "VK_SUCCESS"
	case ErrorOutOfHostMemory:
		return "VK_ERROR_OUT_OF_HOST_MEMORY"
	case ErrorOutOfDeviceMemory:
		return "VK_ERROR_OUT_OF_DEVICE_MEMORY"
	case ErrorInitializationFailed:
		return "VK_ERROR_INITIALIZATION_FAILED"
	case ErrorDeviceLost:
		return "VK_ERROR_DEVICE_LOST"
	case ErrorSurfaceLostKHR:
		return "VK_ERROR_SURFACE_LOST_KHR"
	case ErrorOutOfDateKHR:
		return "VK_ERROR_OUT_OF_DATE_KHR"
	default:
		return "VK_ERROR_UNKNOWN"
	}
}

// Format
type Format int32

const (
	FormatUndefined          Format = 0
	FormatR8Unorm            Format = 9
	FormatR8G8B8A8Unorm      Format = 37
	FormatR8G8B8A8Srgb       Format = 43
	FormatB8G8R8A8Unorm      Format = 44
	FormatB8G8R8A8Srgb       Format = 50
	FormatD32Sfloat          Format = 126
	FormatD24UnormS8Uint     Format = 129
)

// Color space
type ColorSpaceKHR int32

const (
	ColorSpaceSrgbNonlinearKHR ColorSpaceKHR = 0
)

// Present mode
type PresentModeKHR int32

const (
	PresentModeImmediateKHR    PresentModeKHR = 0
	PresentModeMailboxKHR      PresentModeKHR = 1
	PresentModeFifoKHR         PresentModeKHR = 2
	PresentModeFifoRelaxedKHR  PresentModeKHR = 3
)

// Image usage flags
type ImageUsageFlags uint32

const (
	ImageUsageTransferSrcBit         ImageUsageFlags = 0x00000001
	ImageUsageTransferDstBit         ImageUsageFlags = 0x00000002
	ImageUsageSampledBit             ImageUsageFlags = 0x00000004
	ImageUsageColorAttachmentBit     ImageUsageFlags = 0x00000010
	ImageUsageDepthStencilAttachmentBit ImageUsageFlags = 0x00000020
)

// Sharing mode
type SharingMode int32

const (
	SharingModeExclusive  SharingMode = 0
	SharingModeConcurrent SharingMode = 1
)

// Composite alpha
type CompositeAlphaFlagBitsKHR uint32

const (
	CompositeAlphaOpaqueBitKHR CompositeAlphaFlagBitsKHR = 0x00000001
)

// Surface transform
type SurfaceTransformFlagBitsKHR uint32

const (
	SurfaceTransformIdentityBitKHR SurfaceTransformFlagBitsKHR = 0x00000001
)

// Image layout
type ImageLayout int32

const (
	ImageLayoutUndefined                ImageLayout = 0
	ImageLayoutGeneral                  ImageLayout = 1
	ImageLayoutColorAttachmentOptimal   ImageLayout = 2
	ImageLayoutDepthStencilAttachmentOptimal ImageLayout = 3
	ImageLayoutShaderReadOnlyOptimal    ImageLayout = 5
	ImageLayoutTransferSrcOptimal       ImageLayout = 6
	ImageLayoutTransferDstOptimal       ImageLayout = 7
	ImageLayoutPresentSrcKHR            ImageLayout = 1000001002
)

// Image aspect
type ImageAspectFlags uint32

const (
	ImageAspectColorBit   ImageAspectFlags = 0x00000001
	ImageAspectDepthBit   ImageAspectFlags = 0x00000002
	ImageAspectStencilBit ImageAspectFlags = 0x00000004
)

// Component swizzle
type ComponentSwizzle int32

const ComponentSwizzleIdentity ComponentSwizzle = 0

// Image view type
type ImageViewType int32

const ImageViewType2D ImageViewType = 1

// Attachment load/store ops
type AttachmentLoadOp int32

const (
	AttachmentLoadOpLoad     AttachmentLoadOp = 0
	AttachmentLoadOpClear    AttachmentLoadOp = 1
	AttachmentLoadOpDontCare AttachmentLoadOp = 2
)

type AttachmentStoreOp int32

const (
	AttachmentStoreOpStore    AttachmentStoreOp = 0
	AttachmentStoreOpDontCare AttachmentStoreOp = 1
)

// Sample count
type SampleCountFlagBits uint32

const SampleCount1Bit SampleCountFlagBits = 0x00000001

// Pipeline bind point
type PipelineBindPoint int32

const PipelineBindPointGraphics PipelineBindPoint = 0

// Pipeline stage
type PipelineStageFlags uint32

const (
	PipelineStageColorAttachmentOutputBit PipelineStageFlags = 0x00000400
)

// Access flags
type AccessFlags uint32

const (
	AccessColorAttachmentWriteBit AccessFlags = 0x00000100
)

// Subpass
const SubpassExternal = ^uint32(0)

// Command buffer level
type CommandBufferLevel int32

const CommandBufferLevelPrimary CommandBufferLevel = 0

// Subpass contents
type SubpassContents int32

const SubpassContentsInline SubpassContents = 0

// Index type
type IndexType int32

const (
	IndexTypeUint16 IndexType = 0
	IndexTypeUint32 IndexType = 1
)

// Buffer usage
type BufferUsageFlags uint32

const (
	BufferUsageTransferSrcBit  BufferUsageFlags = 0x00000001
	BufferUsageTransferDstBit  BufferUsageFlags = 0x00000002
	BufferUsageVertexBufferBit BufferUsageFlags = 0x00000080
	BufferUsageIndexBufferBit  BufferUsageFlags = 0x00000040
	BufferUsageUniformBufferBit BufferUsageFlags = 0x00000010
)

// Memory property
type MemoryPropertyFlags uint32

const (
	MemoryPropertyDeviceLocalBit  MemoryPropertyFlags = 0x00000001
	MemoryPropertyHostVisibleBit  MemoryPropertyFlags = 0x00000002
	MemoryPropertyHostCoherentBit MemoryPropertyFlags = 0x00000004
)

// Shader stage
type ShaderStageFlags uint32

const (
	ShaderStageVertexBit   ShaderStageFlags = 0x00000001
	ShaderStageFragmentBit ShaderStageFlags = 0x00000010
)

// Vertex input rate
type VertexInputRate int32

const (
	VertexInputRateVertex   VertexInputRate = 0
	VertexInputRateInstance VertexInputRate = 1
)

// Primitive topology
type PrimitiveTopology int32

const (
	PrimitiveTopologyTriangleList PrimitiveTopology = 3
)

// Polygon mode
type PolygonMode int32

const PolygonModeFill PolygonMode = 0

// Cull mode
type CullModeFlags uint32

const (
	CullModeNone     CullModeFlags = 0
	CullModeBackBit  CullModeFlags = 0x00000002
)

// Front face
type FrontFace int32

const (
	FrontFaceCounterClockwise FrontFace = 0
	FrontFaceClockwise        FrontFace = 1
)

// Blend factor
type BlendFactor int32

const (
	BlendFactorZero             BlendFactor = 0
	BlendFactorOne              BlendFactor = 1
	BlendFactorSrcAlpha         BlendFactor = 6
	BlendFactorOneMinusSrcAlpha BlendFactor = 7
)

// Blend op
type BlendOp int32

const BlendOpAdd BlendOp = 0

// Color component
type ColorComponentFlags uint32

const (
	ColorComponentRBit ColorComponentFlags = 0x00000001
	ColorComponentGBit ColorComponentFlags = 0x00000002
	ColorComponentBBit ColorComponentFlags = 0x00000004
	ColorComponentABit ColorComponentFlags = 0x00000008
	ColorComponentAll  ColorComponentFlags = 0x0000000F
)

// Dynamic state
type DynamicState int32

const (
	DynamicStateViewport DynamicState = 0
	DynamicStateScissor  DynamicState = 1
)

// Filter
type Filter int32

const (
	FilterNearest Filter = 0
	FilterLinear  Filter = 1
)

// Sampler address mode
type SamplerAddressMode int32

const (
	SamplerAddressModeRepeat         SamplerAddressMode = 0
	SamplerAddressModeMirroredRepeat SamplerAddressMode = 1
	SamplerAddressModeClampToEdge    SamplerAddressMode = 2
)

// Sampler mipmap mode
type SamplerMipmapMode int32

const SamplerMipmapModeLinear SamplerMipmapMode = 1

// Descriptor type
type DescriptorType int32

const (
	DescriptorTypeUniformBuffer        DescriptorType = 6
	DescriptorTypeCombinedImageSampler DescriptorType = 1
)

// Queue family properties
type QueueFamilyProperties struct {
	QueueFlags                uint32
	QueueCount                uint32
	TimestampValidBits        uint32
	MinImageTransferGranularity Extent3D
}

const (
	QueueGraphicsBit uint32 = 0x00000001
	QueueComputeBit  uint32 = 0x00000002
	QueueTransferBit uint32 = 0x00000004
)

// Structure types (sType field)
type StructureType int32

const (
	StructureTypeApplicationInfo                StructureType = 0
	StructureTypeInstanceCreateInfo             StructureType = 1
	StructureTypeDeviceQueueCreateInfo          StructureType = 2
	StructureTypeDeviceCreateInfo               StructureType = 3
	StructureTypeSubmitInfo                     StructureType = 4
	StructureTypeMemoryAllocateInfo             StructureType = 5
	StructureTypeFenceCreateInfo                StructureType = 8
	StructureTypeSemaphoreCreateInfo            StructureType = 9
	StructureTypeBufferCreateInfo               StructureType = 12
	StructureTypeImageCreateInfo                StructureType = 14
	StructureTypeImageViewCreateInfo            StructureType = 15
	StructureTypeShaderModuleCreateInfo         StructureType = 16
	StructureTypePipelineShaderStageCreateInfo  StructureType = 18
	StructureTypePipelineVertexInputStateCreateInfo StructureType = 19
	StructureTypePipelineInputAssemblyStateCreateInfo StructureType = 20
	StructureTypePipelineViewportStateCreateInfo StructureType = 22
	StructureTypePipelineRasterizationStateCreateInfo StructureType = 23
	StructureTypePipelineMultisampleStateCreateInfo StructureType = 24
	StructureTypePipelineColorBlendStateCreateInfo StructureType = 26
	StructureTypePipelineDynamicStateCreateInfo StructureType = 27
	StructureTypePipelineLayoutCreateInfo       StructureType = 30
	StructureTypeRenderPassCreateInfo           StructureType = 38
	StructureTypeGraphicsPipelineCreateInfo     StructureType = 28
	StructureTypeFramebufferCreateInfo          StructureType = 37
	StructureTypeCommandPoolCreateInfo          StructureType = 39
	StructureTypeCommandBufferAllocateInfo      StructureType = 40
	StructureTypeCommandBufferBeginInfo         StructureType = 42
	StructureTypeRenderPassBeginInfo            StructureType = 43
	StructureTypePresentInfoKHR                 StructureType = 1000001001
	StructureTypeSwapchainCreateInfoKHR         StructureType = 1000001000
	StructureTypeDescriptorSetLayoutCreateInfo  StructureType = 32
	StructureTypeDescriptorPoolCreateInfo       StructureType = 33
	StructureTypeDescriptorSetAllocateInfo      StructureType = 34
	StructureTypeWriteDescriptorSet             StructureType = 35
	StructureTypeSamplerCreateInfo              StructureType = 31
	StructureTypeWin32SurfaceCreateInfoKHR      StructureType = 1000009000
	StructureTypeXlibSurfaceCreateInfoKHR       StructureType = 1000004000
	StructureTypeWaylandSurfaceCreateInfoKHR    StructureType = 1000006000
	StructureTypeMappedMemoryRange              StructureType = 6
	StructureTypeBufferMemoryBarrier            StructureType = 44
	StructureTypeImageMemoryBarrier             StructureType = 45
)

// Fence create flags
const FenceCreateSignaledBit uint32 = 0x00000001

// Command pool create flags
const CommandPoolCreateResetCommandBufferBit uint32 = 0x00000002

// Geometry structs
type Extent2D struct {
	Width  uint32
	Height uint32
}

type Extent3D struct {
	Width  uint32
	Height uint32
	Depth  uint32
}

type Offset2D struct {
	X int32
	Y int32
}

type Rect2D struct {
	Offset Offset2D
	Extent Extent2D
}

type Viewport struct {
	X        float32
	Y        float32
	Width    float32
	Height   float32
	MinDepth float32
	MaxDepth float32
}

type ClearValue struct {
	Color [4]float32
}

// Surface capabilities
type SurfaceCapabilitiesKHR struct {
	MinImageCount           uint32
	MaxImageCount           uint32
	CurrentExtent           Extent2D
	MinImageExtent          Extent2D
	MaxImageExtent          Extent2D
	MaxImageArrayLayers     uint32
	SupportedTransforms     uint32
	CurrentTransform        SurfaceTransformFlagBitsKHR
	SupportedCompositeAlpha uint32
	SupportedUsageFlags     ImageUsageFlags
}

// Surface format
type SurfaceFormatKHR struct {
	Format     Format
	ColorSpace ColorSpaceKHR
}

// Memory requirements
type MemoryRequirements struct {
	Size           uint64
	Alignment      uint64
	MemoryTypeBits uint32
}

// Physical device memory properties
type PhysicalDeviceMemoryProperties struct {
	MemoryTypeCount uint32
	MemoryTypes     [32]MemoryType
	MemoryHeapCount uint32
	MemoryHeaps     [16]MemoryHeap
}

type MemoryType struct {
	PropertyFlags MemoryPropertyFlags
	HeapIndex     uint32
}

type MemoryHeap struct {
	Size  uint64
	Flags uint32
}

// Null handle
const NullHandle = 0

// API version
func MakeAPIVersion(major, minor, patch uint32) uint32 {
	return (major << 22) | (minor << 12) | patch
}

// WholeSize for buffer mapping
const WholeSize = ^uint64(0)

// Max timeout for fences
const MaxTimeout = ^uint64(0)

// Pointer helpers
func ptrToBytes(p unsafe.Pointer, size int) []byte {
	return unsafe.Slice((*byte)(p), size)
}

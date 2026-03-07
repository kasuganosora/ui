package vulkan

import (
	"fmt"
	"unsafe"

	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// imageCreateInfo for vkCreateImage.
type imageCreateInfo struct {
	SType                 StructureType
	PNext                 unsafe.Pointer
	Flags                 uint32
	ImageType             int32 // VkImageType
	Format                Format
	Extent                Extent3D
	MipLevels             uint32
	ArrayLayers           uint32
	Samples               SampleCountFlagBits
	Tiling                int32 // VkImageTiling
	Usage                 ImageUsageFlags
	SharingMode           SharingMode
	QueueFamilyIndexCount uint32
	PQueueFamilyIndices   *uint32
	InitialLayout         ImageLayout
}

// samplerCreateInfo for vkCreateSampler.
type samplerCreateInfo struct {
	SType                   StructureType
	PNext                   unsafe.Pointer
	Flags                   uint32
	MagFilter               Filter
	MinFilter               Filter
	MipmapMode              SamplerMipmapMode
	AddressModeU            SamplerAddressMode
	AddressModeV            SamplerAddressMode
	AddressModeW            SamplerAddressMode
	MipLodBias              float32
	AnisotropyEnable        uint32
	MaxAnisotropy           float32
	CompareEnable           uint32
	CompareOp               int32
	MinLod                  float32
	MaxLod                  float32
	BorderColor             int32
	UnnormalizedCoordinates uint32
}

// Image tiling
const (
	imageTilingOptimal int32 = 0
	imageTilingLinear  int32 = 1
)

// Image type
const imageType2D int32 = 1

// Border color
const borderColorFloatTransparentBlack int32 = 0

// textureFormatToVk converts render.TextureFormat to Vulkan Format.
func textureFormatToVk(f render.TextureFormat) Format {
	switch f {
	case render.TextureFormatR8:
		return FormatR8Unorm
	case render.TextureFormatRGBA8:
		return FormatR8G8B8A8Unorm
	case render.TextureFormatBGRA8:
		return FormatB8G8R8A8Unorm
	default:
		return FormatR8G8B8A8Unorm
	}
}

// textureBytesPerPixel returns bytes per pixel for a texture format.
func textureBytesPerPixel(f render.TextureFormat) int {
	switch f {
	case render.TextureFormatR8:
		return 1
	case render.TextureFormatRGBA8, render.TextureFormatBGRA8:
		return 4
	default:
		return 4
	}
}

// createTextureImage creates a VkImage + VkDeviceMemory for a texture.
func (b *Backend) createTextureImage(width, height int, vkFormat Format) (Image, DeviceMemory, error) {
	ci := imageCreateInfo{
		SType:     StructureTypeImageCreateInfo,
		ImageType: imageType2D,
		Format:    vkFormat,
		Extent: Extent3D{
			Width:  uint32(width),
			Height: uint32(height),
			Depth:  1,
		},
		MipLevels:     1,
		ArrayLayers:   1,
		Samples:       SampleCount1Bit,
		Tiling:        imageTilingOptimal,
		Usage:         ImageUsageTransferDstBit | ImageUsageSampledBit,
		SharingMode:   SharingModeExclusive,
		InitialLayout: ImageLayoutUndefined,
	}

	var image Image
	r, _, _ := syscallN(b.loader.vkCreateImage,
		uintptr(b.device), uintptr(unsafe.Pointer(&ci)), 0, uintptr(unsafe.Pointer(&image)),
	)
	if Result(r) != Success {
		return 0, 0, fmt.Errorf("vulkan: vkCreateImage failed: %v", Result(r))
	}

	var memReqs MemoryRequirements
	syscallN(b.loader.vkGetImageMemoryRequirements,
		uintptr(b.device), uintptr(image), uintptr(unsafe.Pointer(&memReqs)),
	)

	memTypeIdx, ok := FindMemoryType(b.memProps, memReqs.MemoryTypeBits, MemoryPropertyDeviceLocalBit)
	if !ok {
		syscallN(b.loader.vkDestroyImage, uintptr(b.device), uintptr(image), 0)
		return 0, 0, fmt.Errorf("vulkan: no suitable memory type for texture")
	}

	ai := memoryAllocateInfo{
		SType:           StructureTypeMemoryAllocateInfo,
		AllocationSize:  memReqs.Size,
		MemoryTypeIndex: memTypeIdx,
	}

	var memory DeviceMemory
	r, _, _ = syscallN(b.loader.vkAllocateMemory,
		uintptr(b.device), uintptr(unsafe.Pointer(&ai)), 0, uintptr(unsafe.Pointer(&memory)),
	)
	if Result(r) != Success {
		syscallN(b.loader.vkDestroyImage, uintptr(b.device), uintptr(image), 0)
		return 0, 0, fmt.Errorf("vulkan: vkAllocateMemory for texture failed: %v", Result(r))
	}

	r, _, _ = syscallN(b.loader.vkBindImageMemory,
		uintptr(b.device), uintptr(image), uintptr(memory), 0,
	)
	if Result(r) != Success {
		syscallN(b.loader.vkFreeMemory, uintptr(b.device), uintptr(memory), 0)
		syscallN(b.loader.vkDestroyImage, uintptr(b.device), uintptr(image), 0)
		return 0, 0, fmt.Errorf("vulkan: vkBindImageMemory failed: %v", Result(r))
	}

	return image, memory, nil
}

// createTextureImageView creates a VkImageView for a texture image.
func (b *Backend) createTextureImageView(image Image, vkFormat Format) (ImageView, error) {
	ci := imageViewCreateInfo{
		SType:    StructureTypeImageViewCreateInfo,
		Image:    image,
		ViewType: ImageViewType2D,
		Format:   vkFormat,
		Components: [4]ComponentSwizzle{
			ComponentSwizzleIdentity,
			ComponentSwizzleIdentity,
			ComponentSwizzleIdentity,
			ComponentSwizzleIdentity,
		},
		SubresourceRange: imageSubresourceRange{
			AspectMask:     ImageAspectColorBit,
			BaseMipLevel:   0,
			LevelCount:     1,
			BaseArrayLayer: 0,
			LayerCount:     1,
		},
	}

	var view ImageView
	r, _, _ := syscallN(b.loader.vkCreateImageView,
		uintptr(b.device), uintptr(unsafe.Pointer(&ci)), 0, uintptr(unsafe.Pointer(&view)),
	)
	if Result(r) != Success {
		return 0, fmt.Errorf("vulkan: vkCreateImageView for texture failed: %v", Result(r))
	}
	return view, nil
}

// createTextureSampler creates a VkSampler with linear filtering and clamp-to-edge.
func (b *Backend) createTextureSampler() (Sampler, error) {
	return b.createTextureSamplerWithFilter(render.TextureFilterLinear)
}

// createTextureSamplerWithFilter creates a VkSampler with the specified filtering.
func (b *Backend) createTextureSamplerWithFilter(filter render.TextureFilter) (Sampler, error) {
	vkFilter := FilterLinear
	if filter == render.TextureFilterNearest {
		vkFilter = FilterNearest
	}
	ci := samplerCreateInfo{
		SType:        StructureTypeSamplerCreateInfo,
		MagFilter:    vkFilter,
		MinFilter:    vkFilter,
		MipmapMode:   SamplerMipmapModeLinear,
		AddressModeU: SamplerAddressModeClampToEdge,
		AddressModeV: SamplerAddressModeClampToEdge,
		AddressModeW: SamplerAddressModeClampToEdge,
		MaxLod:       1.0,
		BorderColor:  borderColorFloatTransparentBlack,
	}

	var sampler Sampler
	r, _, _ := syscallN(b.loader.vkCreateSampler,
		uintptr(b.device), uintptr(unsafe.Pointer(&ci)), 0, uintptr(unsafe.Pointer(&sampler)),
	)
	if Result(r) != Success {
		return 0, fmt.Errorf("vulkan: vkCreateSampler failed: %v", Result(r))
	}
	return sampler, nil
}

// imageMemoryBarrier for layout transitions.
type imageMemoryBarrier struct {
	SType               StructureType
	PNext               unsafe.Pointer
	SrcAccessMask       AccessFlags
	DstAccessMask       AccessFlags
	OldLayout           ImageLayout
	NewLayout           ImageLayout
	SrcQueueFamilyIndex uint32
	DstQueueFamilyIndex uint32
	Image               Image
	SubresourceRange    imageSubresourceRange
}

// Pipeline stage flags needed for barriers.
const (
	PipelineStageTopOfPipeBit    PipelineStageFlags = 0x00000001
	PipelineStageTransferBit     PipelineStageFlags = 0x00001000
	PipelineStageFragmentShaderBit PipelineStageFlags = 0x00000080
)

// Access flags for transfer operations.
const (
	AccessTransferReadBit   AccessFlags = 0x00000800
	AccessTransferWriteBit  AccessFlags = 0x00002000
	AccessShaderReadBit     AccessFlags = 0x00000020
)

const queueFamilyIgnored = ^uint32(0)

// bufferImageCopy for vkCmdCopyBufferToImage.
type bufferImageCopy struct {
	BufferOffset      uint64
	BufferRowLength   uint32
	BufferImageHeight uint32
	ImageSubresource  imageSubresourceLayers
	ImageOffset       Offset3D
	ImageExtent       Extent3D
}

type imageSubresourceLayers struct {
	AspectMask     ImageAspectFlags
	MipLevel       uint32
	BaseArrayLayer uint32
	LayerCount     uint32
}

type Offset3D struct {
	X, Y, Z int32
}

// transitionImageLayout records a pipeline barrier to transition an image layout.
func (b *Backend) transitionImageLayout(cmd CommandBuffer, image Image, oldLayout, newLayout ImageLayout, srcStage, dstStage PipelineStageFlags, srcAccess, dstAccess AccessFlags) {
	barrier := imageMemoryBarrier{
		SType:               StructureTypeImageMemoryBarrier,
		SrcAccessMask:       srcAccess,
		DstAccessMask:       dstAccess,
		OldLayout:           oldLayout,
		NewLayout:           newLayout,
		SrcQueueFamilyIndex: queueFamilyIgnored,
		DstQueueFamilyIndex: queueFamilyIgnored,
		Image:               image,
		SubresourceRange: imageSubresourceRange{
			AspectMask:   ImageAspectColorBit,
			LevelCount:   1,
			LayerCount:   1,
		},
	}

	syscallN(b.loader.vkCmdPipelineBarrier,
		uintptr(cmd),
		uintptr(srcStage), uintptr(dstStage),
		0, // dependency flags
		0, 0, // memory barriers
		0, 0, // buffer memory barriers
		1, uintptr(unsafe.Pointer(&barrier)),
	)
}

// beginOneTimeCommands allocates and begins a single-use command buffer.
func (b *Backend) beginOneTimeCommands() (CommandBuffer, error) {
	allocInfo := commandBufferAllocateInfo{
		SType:              StructureTypeCommandBufferAllocateInfo,
		CommandPool:        b.commandPool,
		Level:              CommandBufferLevelPrimary,
		CommandBufferCount: 1,
	}

	var cmd CommandBuffer
	r, _, _ := syscallN(b.loader.vkAllocateCommandBuffers,
		uintptr(b.device), uintptr(unsafe.Pointer(&allocInfo)), uintptr(unsafe.Pointer(&cmd)),
	)
	if Result(r) != Success {
		return 0, fmt.Errorf("vulkan: failed to allocate one-time command buffer: %v", Result(r))
	}

	beginInfo := commandBufferBeginInfo{
		SType: StructureTypeCommandBufferBeginInfo,
		Flags: 0x00000001, // VK_COMMAND_BUFFER_USAGE_ONE_TIME_SUBMIT_BIT
	}
	syscallN(b.loader.vkBeginCommandBuffer, uintptr(cmd), uintptr(unsafe.Pointer(&beginInfo)))

	return cmd, nil
}

// endOneTimeCommands ends, submits, and waits for a one-time command buffer.
func (b *Backend) endOneTimeCommands(cmd CommandBuffer) {
	syscallN(b.loader.vkEndCommandBuffer, uintptr(cmd))

	si := submitInfo{
		SType:              StructureTypeSubmitInfo,
		CommandBufferCount: 1,
		PCommandBuffers:    &cmd,
	}

	r, _, _ := syscallN(b.loader.vkQueueSubmit,
		uintptr(b.graphicsQueue), 1, uintptr(unsafe.Pointer(&si)), 0,
	)
	if Result(r) != Success {
		fmt.Printf("[vk] endOneTimeCommands: vkQueueSubmit failed: %v\n", Result(r))
	}
	r2, _, _ := syscallN(b.loader.vkQueueWaitIdle, uintptr(b.graphicsQueue))
	if Result(r2) != Success {
		fmt.Printf("[vk] endOneTimeCommands: vkQueueWaitIdle failed: %v\n", Result(r2))
	}

	syscallN(b.loader.vkFreeCommandBuffers,
		uintptr(b.device), uintptr(b.commandPool), 1, uintptr(unsafe.Pointer(&cmd)),
	)
}

// uploadTextureData uploads pixel data to a texture using a staging buffer.
func (b *Backend) uploadTextureData(image Image, width, height int, data []byte) error {
	dataSize := uint64(len(data))

	// Create staging buffer
	stagingBuf, stagingMem, err := b.createBuffer(
		dataSize,
		BufferUsageTransferSrcBit,
		MemoryPropertyHostVisibleBit|MemoryPropertyHostCoherentBit,
	)
	if err != nil {
		return fmt.Errorf("vulkan: staging buffer: %w", err)
	}
	defer b.destroyBuffer(stagingBuf, stagingMem)

	// Map and copy data to staging buffer
	var mapped unsafe.Pointer
	syscallN(b.loader.vkMapMemory,
		uintptr(b.device), uintptr(stagingMem), 0, uintptr(dataSize), 0,
		uintptr(unsafe.Pointer(&mapped)),
	)
	copy(unsafe.Slice((*byte)(mapped), dataSize), data)
	syscallN(b.loader.vkUnmapMemory, uintptr(b.device), uintptr(stagingMem))

	// Record copy commands
	cmd, err := b.beginOneTimeCommands()
	if err != nil {
		return err
	}

	// Transition image to transfer dst
	b.transitionImageLayout(cmd, image,
		ImageLayoutUndefined, ImageLayoutTransferDstOptimal,
		PipelineStageTopOfPipeBit, PipelineStageTransferBit,
		0, AccessTransferWriteBit,
	)

	// Copy buffer to image
	region := bufferImageCopy{
		ImageSubresource: imageSubresourceLayers{
			AspectMask: ImageAspectColorBit,
			LayerCount: 1,
		},
		ImageExtent: Extent3D{
			Width:  uint32(width),
			Height: uint32(height),
			Depth:  1,
		},
	}

	syscallN(b.loader.vkCmdCopyBufferToImage,
		uintptr(cmd), uintptr(stagingBuf), uintptr(image),
		uintptr(ImageLayoutTransferDstOptimal),
		1, uintptr(unsafe.Pointer(&region)),
	)

	// Transition image to shader read
	b.transitionImageLayout(cmd, image,
		ImageLayoutTransferDstOptimal, ImageLayoutShaderReadOnlyOptimal,
		PipelineStageTransferBit, PipelineStageFragmentShaderBit,
		AccessTransferWriteBit, AccessShaderReadBit,
	)

	b.endOneTimeCommands(cmd)
	return nil
}

// uploadTextureRegion uploads pixel data to a sub-region of a texture.
func (b *Backend) uploadTextureRegion(image Image, region uimath.Rect, rowBytes int, data []byte) error {
	dataSize := uint64(len(data))

	stagingBuf, stagingMem, err := b.createBuffer(
		dataSize,
		BufferUsageTransferSrcBit,
		MemoryPropertyHostVisibleBit|MemoryPropertyHostCoherentBit,
	)
	if err != nil {
		return fmt.Errorf("vulkan: staging buffer: %w", err)
	}
	defer b.destroyBuffer(stagingBuf, stagingMem)

	var mapped unsafe.Pointer
	syscallN(b.loader.vkMapMemory,
		uintptr(b.device), uintptr(stagingMem), 0, uintptr(dataSize), 0,
		uintptr(unsafe.Pointer(&mapped)),
	)
	copy(unsafe.Slice((*byte)(mapped), dataSize), data)
	syscallN(b.loader.vkUnmapMemory, uintptr(b.device), uintptr(stagingMem))

	cmd, err := b.beginOneTimeCommands()
	if err != nil {
		return err
	}

	// Transition to transfer dst
	b.transitionImageLayout(cmd, image,
		ImageLayoutShaderReadOnlyOptimal, ImageLayoutTransferDstOptimal,
		PipelineStageFragmentShaderBit, PipelineStageTransferBit,
		AccessShaderReadBit, AccessTransferWriteBit,
	)

	copyRegion := bufferImageCopy{
		ImageSubresource: imageSubresourceLayers{
			AspectMask: ImageAspectColorBit,
			LayerCount: 1,
		},
		ImageOffset: Offset3D{
			X: int32(region.X),
			Y: int32(region.Y),
		},
		ImageExtent: Extent3D{
			Width:  uint32(region.Width),
			Height: uint32(region.Height),
			Depth:  1,
		},
	}

	syscallN(b.loader.vkCmdCopyBufferToImage,
		uintptr(cmd), uintptr(stagingBuf), uintptr(image),
		uintptr(ImageLayoutTransferDstOptimal),
		1, uintptr(unsafe.Pointer(&copyRegion)),
	)

	// Transition back to shader read
	b.transitionImageLayout(cmd, image,
		ImageLayoutTransferDstOptimal, ImageLayoutShaderReadOnlyOptimal,
		PipelineStageTransferBit, PipelineStageFragmentShaderBit,
		AccessTransferWriteBit, AccessShaderReadBit,
	)

	b.endOneTimeCommands(cmd)
	return nil
}

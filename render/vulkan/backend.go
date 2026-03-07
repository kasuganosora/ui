package vulkan

import (
	"fmt"
	"sort"
	"unsafe"

	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/render"
)

const maxFramesInFlight = 2

// Backend implements render.Backend using Vulkan.
type Backend struct {
	loader *Loader

	instance       Instance
	physicalDevice PhysicalDevice
	device         Device
	surface        SurfaceKHR
	graphicsQueue  Queue
	presentQueue   Queue
	queueIndices   QueueFamilyIndices
	memProps       PhysicalDeviceMemoryProperties

	swapchain  *Swapchain
	renderPass RenderPass

	commandPool    CommandPool
	commandBuffers [maxFramesInFlight]CommandBuffer

	// Synchronization
	imageAvailableSems [maxFramesInFlight]Semaphore
	renderFinishedSems [maxFramesInFlight]Semaphore
	inFlightFences     [maxFramesInFlight]Fence
	currentFrame       int

	// Pipelines
	rectPipeline       Pipeline
	rectPipelineLayout PipelineLayout

	// Textured pipeline (for text and images)
	texturedPipeline       Pipeline
	texturedPipelineLayout PipelineLayout
	textPipeline           Pipeline
	textPipelineLayout     PipelineLayout

	// Descriptor infrastructure
	texDescSetLayout DescriptorSetLayout
	texDescPool      DescriptorPool

	// Dynamic vertex buffer for per-frame data
	vertexBuffers [maxFramesInFlight]Buffer
	vertexMemory  [maxFramesInFlight]DeviceMemory
	vertexSize    uint64

	// State
	width, height int
	vsync         bool

	// Texture management
	nextTextureID render.TextureHandle
	textures      map[render.TextureHandle]*textureEntry
}

type textureEntry struct {
	image         Image
	memory        DeviceMemory
	view          ImageView
	sampler       Sampler
	descriptorSet DescriptorSet
	width         int
	height        int
	format        render.TextureFormat
}

// New creates a new Vulkan backend.
func New() *Backend {
	return &Backend{
		nextTextureID: 1,
		textures:      make(map[render.TextureHandle]*textureEntry),
		vertexSize:    64 * 1024, // 64KB initial vertex buffer
	}
}

// Init implements render.Backend.
func (b *Backend) Init(window platform.Window) error {
	var err error

	// Load Vulkan
	b.loader, err = NewLoader()
	if err != nil {
		return err
	}

	// Create instance
	b.instance, err = b.loader.CreateInstance("GoUI", "GoUI Engine", false)
	if err != nil {
		return err
	}

	// Load instance functions
	b.loader.LoadInstanceFunctions(b.instance)

	// Create surface from native window handle
	b.surface, err = b.createSurface(window)
	if err != nil {
		return err
	}

	// Pick physical device
	b.physicalDevice, b.queueIndices, err = b.pickPhysicalDevice()
	if err != nil {
		return err
	}

	b.memProps = b.loader.GetPhysicalDeviceMemoryProperties(b.physicalDevice)

	// Create logical device
	b.device, err = b.loader.CreateDevice(b.physicalDevice, b.queueIndices)
	if err != nil {
		return err
	}

	// Load device functions
	b.loader.LoadDeviceFunctions(b.instance)

	// Get queues
	b.graphicsQueue = b.loader.GetDeviceQueue(b.device, uint32(b.queueIndices.Graphics), 0)
	b.presentQueue = b.loader.GetDeviceQueue(b.device, uint32(b.queueIndices.Present), 0)

	// Get window size
	w, h := window.FramebufferSize()
	b.width = w
	b.height = h

	// Create swapchain
	err = b.createSwapchain()
	if err != nil {
		return err
	}

	// Create render pass
	err = b.createRenderPass()
	if err != nil {
		return err
	}

	// Create framebuffers
	err = b.loader.CreateFramebuffers(b.device, b.renderPass, b.swapchain)
	if err != nil {
		return err
	}

	// Create command pool and buffers
	err = b.createCommandResources()
	if err != nil {
		return err
	}

	// Create synchronization objects
	err = b.createSyncObjects()
	if err != nil {
		return err
	}

	// Create descriptor infrastructure
	err = b.createDescriptorInfrastructure()
	if err != nil {
		return err
	}

	// Create rect pipeline
	err = b.createRectPipeline()
	if err != nil {
		return err
	}

	// Create textured pipelines (for images and text)
	err = b.createTexturedPipeline()
	if err != nil {
		return err
	}
	err = b.createTextPipeline()
	if err != nil {
		return err
	}

	// Create vertex buffers
	for i := 0; i < maxFramesInFlight; i++ {
		b.vertexBuffers[i], b.vertexMemory[i], err = b.createBuffer(
			b.vertexSize,
			BufferUsageVertexBufferBit,
			MemoryPropertyHostVisibleBit|MemoryPropertyHostCoherentBit,
		)
		if err != nil {
			return fmt.Errorf("vulkan: failed to create vertex buffer %d: %w", i, err)
		}
	}

	return nil
}

func (b *Backend) createSurface(window platform.Window) (SurfaceKHR, error) {
	return b.createPlatformSurface(window)
}

func (b *Backend) pickPhysicalDevice() (PhysicalDevice, QueueFamilyIndices, error) {
	devices, err := b.loader.EnumeratePhysicalDevices(b.instance)
	if err != nil {
		return 0, QueueFamilyIndices{}, err
	}

	for _, dev := range devices {
		indices, ok := b.loader.FindQueueFamilies(dev, b.surface)
		if ok {
			return dev, indices, nil
		}
	}
	return 0, QueueFamilyIndices{}, fmt.Errorf("vulkan: no suitable physical device found")
}

func (b *Backend) createSwapchain() error {
	caps, err := b.loader.GetPhysicalDeviceSurfaceCapabilitiesKHR(b.physicalDevice, b.surface)
	if err != nil {
		return err
	}

	formats, err := b.loader.GetPhysicalDeviceSurfaceFormatsKHR(b.physicalDevice, b.surface)
	if err != nil {
		return err
	}

	presentModes, err := b.loader.GetPhysicalDeviceSurfacePresentModesKHR(b.physicalDevice, b.surface)
	if err != nil {
		return err
	}

	surfaceFormat := ChooseSurfaceFormat(formats)
	presentMode := ChoosePresentMode(presentModes, b.vsync)

	var oldSwapchain SwapchainKHR
	if b.swapchain != nil {
		oldSwapchain = b.swapchain.handle
	}

	b.swapchain, err = b.loader.CreateSwapchain(
		b.device, b.surface, caps, surfaceFormat, presentMode,
		uint32(b.width), uint32(b.height),
		b.queueIndices, oldSwapchain,
	)
	if err != nil {
		return err
	}

	return b.loader.CreateImageViews(b.device, b.swapchain)
}

// attachmentDescription for render pass.
type attachmentDescription struct {
	Flags          uint32
	Format         Format
	Samples        SampleCountFlagBits
	LoadOp         AttachmentLoadOp
	StoreOp        AttachmentStoreOp
	StencilLoadOp  AttachmentLoadOp
	StencilStoreOp AttachmentStoreOp
	InitialLayout  ImageLayout
	FinalLayout    ImageLayout
}

type attachmentReference struct {
	Attachment uint32
	Layout     ImageLayout
}

type subpassDescription struct {
	Flags                   uint32
	PipelineBindPoint       PipelineBindPoint
	InputAttachmentCount    uint32
	PInputAttachments       *attachmentReference
	ColorAttachmentCount    uint32
	PColorAttachments       *attachmentReference
	PResolveAttachments     *attachmentReference
	PDepthStencilAttachment *attachmentReference
	PreserveAttachmentCount uint32
	PPreserveAttachments    *uint32
}

type subpassDependency struct {
	SrcSubpass    uint32
	DstSubpass    uint32
	SrcStageMask  PipelineStageFlags
	DstStageMask  PipelineStageFlags
	SrcAccessMask AccessFlags
	DstAccessMask AccessFlags
	DependencyFlags uint32
}

type renderPassCreateInfo struct {
	SType           StructureType
	PNext           unsafe.Pointer
	Flags           uint32
	AttachmentCount uint32
	PAttachments    *attachmentDescription
	SubpassCount    uint32
	PSubpasses      *subpassDescription
	DependencyCount uint32
	PDependencies   *subpassDependency
}

func (b *Backend) createRenderPass() error {
	colorAttachment := attachmentDescription{
		Format:         b.swapchain.format,
		Samples:        SampleCount1Bit,
		LoadOp:         AttachmentLoadOpClear,
		StoreOp:        AttachmentStoreOpStore,
		StencilLoadOp:  AttachmentLoadOpDontCare,
		StencilStoreOp: AttachmentStoreOpDontCare,
		InitialLayout:  ImageLayoutUndefined,
		FinalLayout:    ImageLayoutPresentSrcKHR,
	}

	colorRef := attachmentReference{
		Attachment: 0,
		Layout:     ImageLayoutColorAttachmentOptimal,
	}

	subpass := subpassDescription{
		PipelineBindPoint:    PipelineBindPointGraphics,
		ColorAttachmentCount: 1,
		PColorAttachments:    &colorRef,
	}

	dependency := subpassDependency{
		SrcSubpass:    SubpassExternal,
		DstSubpass:    0,
		SrcStageMask:  PipelineStageColorAttachmentOutputBit,
		DstStageMask:  PipelineStageColorAttachmentOutputBit,
		DstAccessMask: AccessColorAttachmentWriteBit,
	}

	createInfo := renderPassCreateInfo{
		SType:           StructureTypeRenderPassCreateInfo,
		AttachmentCount: 1,
		PAttachments:    &colorAttachment,
		SubpassCount:    1,
		PSubpasses:      &subpass,
		DependencyCount: 1,
		PDependencies:   &dependency,
	}

	r, _, _ := syscallN(b.loader.vkCreateRenderPass,
		uintptr(b.device),
		uintptr(unsafe.Pointer(&createInfo)),
		0,
		uintptr(unsafe.Pointer(&b.renderPass)),
	)
	if Result(r) != Success {
		return fmt.Errorf("vulkan: vkCreateRenderPass failed: %v", Result(r))
	}
	return nil
}

// commandPoolCreateInfo
type commandPoolCreateInfo struct {
	SType            StructureType
	PNext            unsafe.Pointer
	Flags            uint32
	QueueFamilyIndex uint32
}

// commandBufferAllocateInfo
type commandBufferAllocateInfo struct {
	SType              StructureType
	PNext              unsafe.Pointer
	CommandPool        CommandPool
	Level              CommandBufferLevel
	CommandBufferCount uint32
}

func (b *Backend) createCommandResources() error {
	poolInfo := commandPoolCreateInfo{
		SType:            StructureTypeCommandPoolCreateInfo,
		Flags:            CommandPoolCreateResetCommandBufferBit,
		QueueFamilyIndex: uint32(b.queueIndices.Graphics),
	}

	r, _, _ := syscallN(b.loader.vkCreateCommandPool,
		uintptr(b.device),
		uintptr(unsafe.Pointer(&poolInfo)),
		0,
		uintptr(unsafe.Pointer(&b.commandPool)),
	)
	if Result(r) != Success {
		return fmt.Errorf("vulkan: vkCreateCommandPool failed: %v", Result(r))
	}

	allocInfo := commandBufferAllocateInfo{
		SType:              StructureTypeCommandBufferAllocateInfo,
		CommandPool:        b.commandPool,
		Level:              CommandBufferLevelPrimary,
		CommandBufferCount: maxFramesInFlight,
	}

	r, _, _ = syscallN(b.loader.vkAllocateCommandBuffers,
		uintptr(b.device),
		uintptr(unsafe.Pointer(&allocInfo)),
		uintptr(unsafe.Pointer(&b.commandBuffers[0])),
	)
	if Result(r) != Success {
		return fmt.Errorf("vulkan: vkAllocateCommandBuffers failed: %v", Result(r))
	}
	return nil
}

// semaphoreCreateInfo
type semaphoreCreateInfo struct {
	SType StructureType
	PNext unsafe.Pointer
	Flags uint32
}

// fenceCreateInfo
type fenceCreateInfo struct {
	SType StructureType
	PNext unsafe.Pointer
	Flags uint32
}

func (b *Backend) createSyncObjects() error {
	semInfo := semaphoreCreateInfo{SType: StructureTypeSemaphoreCreateInfo}
	fenceInfo := fenceCreateInfo{
		SType: StructureTypeFenceCreateInfo,
		Flags: FenceCreateSignaledBit,
	}

	for i := 0; i < maxFramesInFlight; i++ {
		r, _, _ := syscallN(b.loader.vkCreateSemaphore,
			uintptr(b.device), uintptr(unsafe.Pointer(&semInfo)), 0,
			uintptr(unsafe.Pointer(&b.imageAvailableSems[i])),
		)
		if Result(r) != Success {
			return fmt.Errorf("vulkan: failed to create imageAvailable semaphore: %v", Result(r))
		}
		r, _, _ = syscallN(b.loader.vkCreateSemaphore,
			uintptr(b.device), uintptr(unsafe.Pointer(&semInfo)), 0,
			uintptr(unsafe.Pointer(&b.renderFinishedSems[i])),
		)
		if Result(r) != Success {
			return fmt.Errorf("vulkan: failed to create renderFinished semaphore: %v", Result(r))
		}
		r, _, _ = syscallN(b.loader.vkCreateFence,
			uintptr(b.device), uintptr(unsafe.Pointer(&fenceInfo)), 0,
			uintptr(unsafe.Pointer(&b.inFlightFences[i])),
		)
		if Result(r) != Success {
			return fmt.Errorf("vulkan: failed to create fence: %v", Result(r))
		}
	}
	return nil
}

// RectVertex for rect rendering — 19 float32 fields, 76 bytes.
type RectVertex struct {
	PosX, PosY                             float32 // NDC position
	U, V                                   float32 // UV for SDF computation
	ColorR, ColorG, ColorB, ColorA         float32 // Fill color
	RectW, RectH                           float32 // Rect size in pixels for SDF
	RadiusTL, RadiusTR, RadiusBR, RadiusBL float32 // Corner radii
	BorderWidth                            float32
	BorderR, BorderG, BorderB, BorderA     float32 // Border color
}

// BeginFrame implements render.Backend.
func (b *Backend) BeginFrame() {
	// Wait for this frame's fence
	fence := b.inFlightFences[b.currentFrame]
	syscallN(b.loader.vkWaitForFences,
		uintptr(b.device), 1, uintptr(unsafe.Pointer(&fence)), 1, uintptr(MaxTimeout),
	)
	syscallN(b.loader.vkResetFences,
		uintptr(b.device), 1, uintptr(unsafe.Pointer(&fence)),
	)
}

// EndFrame implements render.Backend.
func (b *Backend) EndFrame() {
	// Present is handled in Submit
}

// Submit implements render.Backend.
func (b *Backend) Submit(buf *render.CommandBuffer) {
	if buf == nil || buf.Len() == 0 {
		// Still need to acquire and present for the frame to progress
		b.submitEmpty()
		return
	}

	// Acquire next image
	imageIndex, result := b.loader.AcquireNextImageKHR(
		b.device, b.swapchain.handle, MaxTimeout,
		b.imageAvailableSems[b.currentFrame], 0,
	)
	if result == ErrorOutOfDateKHR {
		b.recreateSwapchain()
		return
	}

	cmd := b.commandBuffers[b.currentFrame]

	// Begin command buffer
	beginInfo := commandBufferBeginInfo{SType: StructureTypeCommandBufferBeginInfo}
	syscallN(b.loader.vkResetCommandBuffer, uintptr(cmd), 0)
	syscallN(b.loader.vkBeginCommandBuffer, uintptr(cmd), uintptr(unsafe.Pointer(&beginInfo)))

	// Begin render pass
	clearColor := ClearValue{Color: [4]float32{0.0, 0.0, 0.0, 1.0}}
	rpBeginInfo := renderPassBeginInfo{
		SType:       StructureTypeRenderPassBeginInfo,
		RenderPass:  b.renderPass,
		Framebuffer: b.swapchain.framebuffers[imageIndex],
		RenderArea: Rect2D{
			Extent: b.swapchain.extent,
		},
		ClearValueCount: 1,
		PClearValues:    &clearColor,
	}
	syscallN(b.loader.vkCmdBeginRenderPass,
		uintptr(cmd), uintptr(unsafe.Pointer(&rpBeginInfo)), uintptr(SubpassContentsInline),
	)

	// Set viewport and scissor
	viewport := Viewport{
		Width:    float32(b.swapchain.extent.Width),
		Height:   float32(b.swapchain.extent.Height),
		MaxDepth: 1.0,
	}
	scissor := Rect2D{Extent: b.swapchain.extent}
	syscallN(b.loader.vkCmdSetViewport, uintptr(cmd), 0, 1, uintptr(unsafe.Pointer(&viewport)))
	syscallN(b.loader.vkCmdSetScissor, uintptr(cmd), 0, 1, uintptr(unsafe.Pointer(&scissor)))

	// Process commands - sort by z-order
	commands := buf.Commands()
	sorted := make([]render.Command, len(commands))
	copy(sorted, commands)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ZOrder < sorted[j].ZOrder
	})

	// Render commands in z-order, handling clips inline
	b.renderAllCommands(cmd, sorted)

	// End render pass and command buffer
	syscallN(b.loader.vkCmdEndRenderPass, uintptr(cmd))
	syscallN(b.loader.vkEndCommandBuffer, uintptr(cmd))

	// Submit
	b.submitCommandBuffer(cmd, imageIndex)

	b.currentFrame = (b.currentFrame + 1) % maxFramesInFlight
}


// commandBufferBeginInfo
type commandBufferBeginInfo struct {
	SType            StructureType
	PNext            unsafe.Pointer
	Flags            uint32
	PInheritanceInfo unsafe.Pointer
}

// renderPassBeginInfo
type renderPassBeginInfo struct {
	SType           StructureType
	PNext           unsafe.Pointer
	RenderPass      RenderPass
	Framebuffer     Framebuffer
	RenderArea      Rect2D
	ClearValueCount uint32
	PClearValues    *ClearValue
}

// submitInfo
type submitInfo struct {
	SType                StructureType
	PNext                unsafe.Pointer
	WaitSemaphoreCount   uint32
	PWaitSemaphores      *Semaphore
	PWaitDstStageMask    *PipelineStageFlags
	CommandBufferCount   uint32
	PCommandBuffers      *CommandBuffer
	SignalSemaphoreCount uint32
	PSignalSemaphores    *Semaphore
}

// presentInfoKHR
type presentInfoKHR struct {
	SType              StructureType
	PNext              unsafe.Pointer
	WaitSemaphoreCount uint32
	PWaitSemaphores    *Semaphore
	SwapchainCount     uint32
	PSwapchains        *SwapchainKHR
	PImageIndices      *uint32
	PResults           *Result
}

func (b *Backend) submitCommandBuffer(cmd CommandBuffer, imageIndex uint32) {
	waitSem := b.imageAvailableSems[b.currentFrame]
	signalSem := b.renderFinishedSems[b.currentFrame]
	waitStage := PipelineStageColorAttachmentOutputBit

	si := submitInfo{
		SType:                StructureTypeSubmitInfo,
		WaitSemaphoreCount:   1,
		PWaitSemaphores:      &waitSem,
		PWaitDstStageMask:    &waitStage,
		CommandBufferCount:   1,
		PCommandBuffers:      &cmd,
		SignalSemaphoreCount: 1,
		PSignalSemaphores:    &signalSem,
	}

	fence := b.inFlightFences[b.currentFrame]
	syscallN(b.loader.vkQueueSubmit,
		uintptr(b.graphicsQueue), 1, uintptr(unsafe.Pointer(&si)), uintptr(fence),
	)

	// Present
	swapchain := b.swapchain.handle
	pi := presentInfoKHR{
		SType:              StructureTypePresentInfoKHR,
		WaitSemaphoreCount: 1,
		PWaitSemaphores:    &signalSem,
		SwapchainCount:     1,
		PSwapchains:        &swapchain,
		PImageIndices:      &imageIndex,
	}
	r, _, _ := syscallN(b.loader.vkQueuePresentKHR,
		uintptr(b.presentQueue), uintptr(unsafe.Pointer(&pi)),
	)
	if Result(r) == ErrorOutOfDateKHR || Result(r) == SuboptimalKHR {
		b.recreateSwapchain()
	}
}

func (b *Backend) submitEmpty() {
	imageIndex, result := b.loader.AcquireNextImageKHR(
		b.device, b.swapchain.handle, MaxTimeout,
		b.imageAvailableSems[b.currentFrame], 0,
	)
	if result == ErrorOutOfDateKHR {
		b.recreateSwapchain()
		return
	}

	cmd := b.commandBuffers[b.currentFrame]
	beginInfo := commandBufferBeginInfo{SType: StructureTypeCommandBufferBeginInfo}
	syscallN(b.loader.vkResetCommandBuffer, uintptr(cmd), 0)
	syscallN(b.loader.vkBeginCommandBuffer, uintptr(cmd), uintptr(unsafe.Pointer(&beginInfo)))

	clearColor := ClearValue{Color: [4]float32{0, 0, 0, 1}}
	rpBeginInfo := renderPassBeginInfo{
		SType:           StructureTypeRenderPassBeginInfo,
		RenderPass:      b.renderPass,
		Framebuffer:     b.swapchain.framebuffers[imageIndex],
		RenderArea:      Rect2D{Extent: b.swapchain.extent},
		ClearValueCount: 1,
		PClearValues:    &clearColor,
	}
	syscallN(b.loader.vkCmdBeginRenderPass, uintptr(cmd), uintptr(unsafe.Pointer(&rpBeginInfo)), uintptr(SubpassContentsInline))
	syscallN(b.loader.vkCmdEndRenderPass, uintptr(cmd))
	syscallN(b.loader.vkEndCommandBuffer, uintptr(cmd))

	b.submitCommandBuffer(cmd, imageIndex)
	b.currentFrame = (b.currentFrame + 1) % maxFramesInFlight
}

// Resize implements render.Backend.
func (b *Backend) Resize(width, height int) {
	b.width = width
	b.height = height
	b.recreateSwapchain()
}

func (b *Backend) recreateSwapchain() {
	b.loader.DeviceWaitIdle(b.device)

	oldSwapchain := b.swapchain
	b.createSwapchain()
	b.loader.CreateFramebuffers(b.device, b.renderPass, b.swapchain)

	if oldSwapchain != nil {
		// Clean up old framebuffers and views (handle was already reused via oldSwapchain)
		for _, fb := range oldSwapchain.framebuffers {
			if fb != 0 {
				syscallN(b.loader.vkDestroyFramebuffer, uintptr(b.device), uintptr(fb), 0)
			}
		}
		for _, view := range oldSwapchain.imageViews {
			if view != 0 {
				syscallN(b.loader.vkDestroyImageView, uintptr(b.device), uintptr(view), 0)
			}
		}
	}
}

// CreateTexture implements render.Backend.
func (b *Backend) CreateTexture(desc render.TextureDesc) (render.TextureHandle, error) {
	vkFormat := textureFormatToVk(desc.Format)

	image, memory, err := b.createTextureImage(desc.Width, desc.Height, vkFormat)
	if err != nil {
		return render.InvalidTexture, err
	}

	view, err := b.createTextureImageView(image, vkFormat)
	if err != nil {
		syscallN(b.loader.vkFreeMemory, uintptr(b.device), uintptr(memory), 0)
		syscallN(b.loader.vkDestroyImage, uintptr(b.device), uintptr(image), 0)
		return render.InvalidTexture, err
	}

	sampler, err := b.createTextureSampler()
	if err != nil {
		syscallN(b.loader.vkDestroyImageView, uintptr(b.device), uintptr(view), 0)
		syscallN(b.loader.vkFreeMemory, uintptr(b.device), uintptr(memory), 0)
		syscallN(b.loader.vkDestroyImage, uintptr(b.device), uintptr(image), 0)
		return render.InvalidTexture, err
	}

	id := b.nextTextureID
	b.nextTextureID++
	entry := &textureEntry{
		image:   image,
		memory:  memory,
		view:    view,
		sampler: sampler,
		width:   desc.Width,
		height:  desc.Height,
		format:  desc.Format,
	}
	b.textures[id] = entry

	// Upload initial data if provided
	if len(desc.Data) > 0 {
		if err := b.uploadTextureData(image, desc.Width, desc.Height, desc.Data); err != nil {
			b.DestroyTexture(id)
			return render.InvalidTexture, fmt.Errorf("vulkan: initial texture upload: %w", err)
		}
	} else {
		// Transition to shader read layout even without data
		cmd, err := b.beginOneTimeCommands()
		if err != nil {
			b.DestroyTexture(id)
			return render.InvalidTexture, err
		}
		b.transitionImageLayout(cmd, image,
			ImageLayoutUndefined, ImageLayoutShaderReadOnlyOptimal,
			PipelineStageTopOfPipeBit, PipelineStageFragmentShaderBit,
			0, AccessShaderReadBit,
		)
		b.endOneTimeCommands(cmd)
	}

	// Allocate descriptor set for this texture
	if err := b.allocateTextureDescriptorSet(entry); err != nil {
		b.DestroyTexture(id)
		return render.InvalidTexture, err
	}

	return id, nil
}

// UpdateTexture implements render.Backend.
func (b *Backend) UpdateTexture(handle render.TextureHandle, region uimath.Rect, data []byte) error {
	entry, ok := b.textures[handle]
	if !ok {
		return fmt.Errorf("vulkan: texture %d not found", handle)
	}

	bpp := textureBytesPerPixel(entry.format)
	return b.uploadTextureRegion(entry.image, region, int(region.Width)*bpp, data)
}

// DestroyTexture implements render.Backend.
func (b *Backend) DestroyTexture(handle render.TextureHandle) {
	if entry, ok := b.textures[handle]; ok {
		if entry.sampler != 0 {
			syscallN(b.loader.vkDestroySampler, uintptr(b.device), uintptr(entry.sampler), 0)
		}
		if entry.view != 0 {
			syscallN(b.loader.vkDestroyImageView, uintptr(b.device), uintptr(entry.view), 0)
		}
		if entry.image != 0 {
			syscallN(b.loader.vkDestroyImage, uintptr(b.device), uintptr(entry.image), 0)
		}
		if entry.memory != 0 {
			syscallN(b.loader.vkFreeMemory, uintptr(b.device), uintptr(entry.memory), 0)
		}
		delete(b.textures, handle)
	}
}

// MaxTextureSize implements render.Backend.
func (b *Backend) MaxTextureSize() int {
	return 4096 // Conservative default
}

// Destroy implements render.Backend.
func (b *Backend) Destroy() {
	if b.device == 0 {
		return
	}
	b.loader.DeviceWaitIdle(b.device)

	// Destroy textures
	for id := range b.textures {
		b.DestroyTexture(id)
	}

	// Destroy vertex buffers
	for i := 0; i < maxFramesInFlight; i++ {
		b.destroyBuffer(b.vertexBuffers[i], b.vertexMemory[i])
	}

	// Destroy descriptor infrastructure
	if b.texDescPool != 0 {
		syscallN(b.loader.vkDestroyDescriptorPool, uintptr(b.device), uintptr(b.texDescPool), 0)
	}
	if b.texDescSetLayout != 0 {
		syscallN(b.loader.vkDestroyDescriptorSetLayout, uintptr(b.device), uintptr(b.texDescSetLayout), 0)
	}

	// Destroy pipelines
	if b.textPipeline != 0 {
		syscallN(b.loader.vkDestroyPipeline, uintptr(b.device), uintptr(b.textPipeline), 0)
	}
	if b.textPipelineLayout != 0 {
		syscallN(b.loader.vkDestroyPipelineLayout, uintptr(b.device), uintptr(b.textPipelineLayout), 0)
	}
	if b.texturedPipeline != 0 {
		syscallN(b.loader.vkDestroyPipeline, uintptr(b.device), uintptr(b.texturedPipeline), 0)
	}
	if b.texturedPipelineLayout != 0 {
		syscallN(b.loader.vkDestroyPipelineLayout, uintptr(b.device), uintptr(b.texturedPipelineLayout), 0)
	}
	if b.rectPipeline != 0 {
		syscallN(b.loader.vkDestroyPipeline, uintptr(b.device), uintptr(b.rectPipeline), 0)
	}
	if b.rectPipelineLayout != 0 {
		syscallN(b.loader.vkDestroyPipelineLayout, uintptr(b.device), uintptr(b.rectPipelineLayout), 0)
	}

	// Destroy sync objects
	for i := 0; i < maxFramesInFlight; i++ {
		syscallN(b.loader.vkDestroySemaphore, uintptr(b.device), uintptr(b.imageAvailableSems[i]), 0)
		syscallN(b.loader.vkDestroySemaphore, uintptr(b.device), uintptr(b.renderFinishedSems[i]), 0)
		syscallN(b.loader.vkDestroyFence, uintptr(b.device), uintptr(b.inFlightFences[i]), 0)
	}

	// Destroy command pool
	if b.commandPool != 0 {
		syscallN(b.loader.vkDestroyCommandPool, uintptr(b.device), uintptr(b.commandPool), 0)
	}

	// Destroy swapchain
	if b.swapchain != nil {
		b.loader.DestroySwapchain(b.device, b.swapchain)
	}

	// Destroy render pass
	if b.renderPass != 0 {
		syscallN(b.loader.vkDestroyRenderPass, uintptr(b.device), uintptr(b.renderPass), 0)
	}

	// Destroy device
	b.loader.DestroyDevice(b.device)

	// Destroy surface and instance
	if b.surface != 0 {
		b.loader.DestroySurfaceKHR(b.instance, b.surface)
	}
	b.loader.DestroyInstance(b.instance)
	b.loader.Close()
}

// Buffer helpers

// bufferCreateInfo
type bufferCreateInfo struct {
	SType                 StructureType
	PNext                 unsafe.Pointer
	Flags                 uint32
	Size                  uint64
	Usage                 BufferUsageFlags
	SharingMode           SharingMode
	QueueFamilyIndexCount uint32
	PQueueFamilyIndices   *uint32
}

// memoryAllocateInfo
type memoryAllocateInfo struct {
	SType           StructureType
	PNext           unsafe.Pointer
	AllocationSize  uint64
	MemoryTypeIndex uint32
}

func (b *Backend) createBuffer(size uint64, usage BufferUsageFlags, properties MemoryPropertyFlags) (Buffer, DeviceMemory, error) {
	ci := bufferCreateInfo{
		SType:       StructureTypeBufferCreateInfo,
		Size:        size,
		Usage:       usage,
		SharingMode: SharingModeExclusive,
	}

	var buffer Buffer
	r, _, _ := syscallN(b.loader.vkCreateBuffer,
		uintptr(b.device), uintptr(unsafe.Pointer(&ci)), 0, uintptr(unsafe.Pointer(&buffer)),
	)
	if Result(r) != Success {
		return 0, 0, fmt.Errorf("vulkan: vkCreateBuffer failed: %v", Result(r))
	}

	var memReqs MemoryRequirements
	syscallN(b.loader.vkGetBufferMemoryRequirements,
		uintptr(b.device), uintptr(buffer), uintptr(unsafe.Pointer(&memReqs)),
	)

	memTypeIdx, ok := FindMemoryType(b.memProps, memReqs.MemoryTypeBits, properties)
	if !ok {
		syscallN(b.loader.vkDestroyBuffer, uintptr(b.device), uintptr(buffer), 0)
		return 0, 0, fmt.Errorf("vulkan: no suitable memory type")
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
		syscallN(b.loader.vkDestroyBuffer, uintptr(b.device), uintptr(buffer), 0)
		return 0, 0, fmt.Errorf("vulkan: vkAllocateMemory failed: %v", Result(r))
	}

	syscallN(b.loader.vkBindBufferMemory, uintptr(b.device), uintptr(buffer), uintptr(memory), 0)

	return buffer, memory, nil
}

func (b *Backend) destroyBuffer(buffer Buffer, memory DeviceMemory) {
	if buffer != 0 {
		syscallN(b.loader.vkDestroyBuffer, uintptr(b.device), uintptr(buffer), 0)
	}
	if memory != 0 {
		syscallN(b.loader.vkFreeMemory, uintptr(b.device), uintptr(memory), 0)
	}
}

// Ensure Backend implements render.Backend at compile time.
var _ render.Backend = (*Backend)(nil)

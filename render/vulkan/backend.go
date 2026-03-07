package vulkan

import (
	"fmt"
	"image"
	"image/color"
	"unsafe"

	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/render"
)

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
	commandBuffers []CommandBuffer

	// Synchronization (sized per swapchain image)
	imageAvailableSems []Semaphore
	renderFinishedSems []Semaphore
	inFlightFences     []Fence
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
	vertexBuffers      []Buffer
	vertexMemory       []DeviceMemory
	vertexSizes        []uint64       // Per-frame-slot: allocated buffer size
	frameVertexOffset  uint64         // Current write offset in vertex buffer for this frame
	mappedVertexPtr    unsafe.Pointer // Mapped pointer for current frame's vertex buffer
	staleVertexBuffers [][]Buffer       // Per-frame-slot: old vertex buffers pending destruction
	staleVertexMemory  [][]DeviceMemory // Per-frame-slot: corresponding memory

	// State
	width, height    int
	dpiScale         float32 // DPI scale factor (1.0 = 96 DPI); coordinates are logical pixels
	vsync            bool
	resizePending    bool   // Set by Resize; cleared after swapchain recreation in Submit
	lastImageIndex   uint32 // Last rendered swapchain image (for ReadPixels)
	lastFrameValid   bool   // Whether lastImageIndex is valid
	deviceLost       bool   // Set when VK_ERROR_DEVICE_LOST is detected; rendering becomes no-op

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

	// Log GPU info for diagnostics
	gpuName, apiVer, driverVer := b.loader.GetPhysicalDeviceNameAndDriver(b.physicalDevice)
	fmt.Printf("[vk] GPU: %s (API %d.%d.%d, driver %d.%d.%d)\n",
		gpuName,
		apiVer>>22, (apiVer>>12)&0x3FF, apiVer&0xFFF,
		driverVer>>22, (driverVer>>12)&0x3FF, driverVer&0xFFF,
	)

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

	// Get window size and DPI
	w, h := window.FramebufferSize()
	b.width = w
	b.height = h
	b.dpiScale = window.DPIScale()
	if b.dpiScale <= 0 {
		b.dpiScale = 1.0
	}

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

	// Create vertex buffers (one per swapchain image)
	vbCount := int(b.swapchain.imageCount)
	b.vertexBuffers = make([]Buffer, vbCount)
	b.vertexMemory = make([]DeviceMemory, vbCount)
	b.vertexSizes = make([]uint64, vbCount)
	const initialVertexSize uint64 = 64 * 1024 // 64KB
	for i := 0; i < vbCount; i++ {
		b.vertexSizes[i] = initialVertexSize
		b.vertexBuffers[i], b.vertexMemory[i], err = b.createBuffer(
			initialVertexSize,
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

	n := b.swapchain.imageCount
	b.commandBuffers = make([]CommandBuffer, n)
	allocInfo := commandBufferAllocateInfo{
		SType:              StructureTypeCommandBufferAllocateInfo,
		CommandPool:        b.commandPool,
		Level:              CommandBufferLevelPrimary,
		CommandBufferCount: n,
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
	n := int(b.swapchain.imageCount)
	b.imageAvailableSems = make([]Semaphore, n)
	b.renderFinishedSems = make([]Semaphore, n)
	b.inFlightFences = make([]Fence, n)
	b.staleVertexBuffers = make([][]Buffer, n)
	b.staleVertexMemory = make([][]DeviceMemory, n)

	semInfo := semaphoreCreateInfo{SType: StructureTypeSemaphoreCreateInfo}
	fenceInfo := fenceCreateInfo{
		SType: StructureTypeFenceCreateInfo,
		Flags: FenceCreateSignaledBit,
	}

	for i := 0; i < n; i++ {
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

// DPIScale returns the DPI scale factor. Apps use logical coordinates;
// the backend converts to physical pixels for the GPU.
func (b *Backend) DPIScale() float32 { return b.dpiScale }

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
	if b.deviceLost {
		return
	}
	// Wait for this frame's fence to ensure the GPU finished the previous
	// use of this frame slot. The fence is NOT reset here — it's reset in
	// Submit right before vkQueueSubmit, so that if AcquireNextImageKHR
	// fails (e.g. ErrorOutOfDateKHR after window show/resize), the fence
	// remains signaled and the next BeginFrame won't hang.
	fence := b.inFlightFences[b.currentFrame]
	// 100ms timeout — keeps resize responsive; if we timeout, recover gracefully.
	const fenceTimeout = 100_000_000 // 100ms in nanoseconds
	r, _, _ := syscallN(b.loader.vkWaitForFences,
		uintptr(b.device), 1, uintptr(unsafe.Pointer(&fence)), 1, uintptr(fenceTimeout),
	)
	if Result(r) == ErrorDeviceLost {
		fmt.Printf("[vk] FATAL: VK_ERROR_DEVICE_LOST at WaitForFences (slot %d) — GPU crashed, rendering disabled\n", b.currentFrame)
		b.deviceLost = true
		return
	} else if Result(r) == Timeout {
		// Fence was not signaled within timeout — likely a lost submit. Recover.
		fmt.Printf("[vk] ERROR: WaitForFences TIMEOUT on slot %d — recovering\n", b.currentFrame)
		b.loader.DeviceWaitIdle(b.device)
		// Recreate the fence in signaled state so we can proceed.
		syscallN(b.loader.vkDestroyFence, uintptr(b.device), uintptr(fence), 0)
		fi := fenceCreateInfo{
			SType: StructureTypeFenceCreateInfo,
			Flags: FenceCreateSignaledBit,
		}
		syscallN(b.loader.vkCreateFence,
			uintptr(b.device), uintptr(unsafe.Pointer(&fi)), 0,
			uintptr(unsafe.Pointer(&b.inFlightFences[b.currentFrame])),
		)
	} else if Result(r) != Success {
		fmt.Printf("[vk] WaitForFences error: %v\n", Result(r))
	}

	// Destroy stale vertex buffers from the previous use of this frame slot.
	// The fence guarantees the GPU has finished with them.
	for i, buf := range b.staleVertexBuffers[b.currentFrame] {
		b.destroyBuffer(buf, b.staleVertexMemory[b.currentFrame][i])
	}
	b.staleVertexBuffers[b.currentFrame] = b.staleVertexBuffers[b.currentFrame][:0]
	b.staleVertexMemory[b.currentFrame] = b.staleVertexMemory[b.currentFrame][:0]
}

// EndFrame implements render.Backend.
func (b *Backend) EndFrame() {
	// Present is handled in Submit
}

// Submit implements render.Backend.
func (b *Backend) Submit(buf *render.CommandBuffer) {
	if b.deviceLost {
		return
	}
	if buf == nil || buf.Len() == 0 {
		b.submitEmpty()
		return
	}

	// Recreate swapchain if a resize was requested.
	if b.resizePending {
		b.recreateSwapchain()
		b.resizePending = false
	}

	// Acquire next image.
	imageIndex, result := b.loader.AcquireNextImageKHR(
		b.device, b.swapchain.handle, MaxTimeout,
		b.imageAvailableSems[b.currentFrame], 0,
	)
	if result == ErrorOutOfDateKHR {
		b.resizePending = true
		return
	}
	if result == NotReady || result == Timeout {
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

	// Reset per-frame vertex offset and map vertex buffer for the frame
	b.frameVertexOffset = 0
	b.mapVertexBuffer()

	// Process commands in painters algorithm order (as emitted by the widget tree).
	// The widget Draw methods already emit commands in the correct back-to-front order.
	commands := buf.Commands()
	b.renderAllCommands(cmd, commands)

	// Unmap vertex buffer
	b.unmapVertexBuffer()

	// End render pass and command buffer
	syscallN(b.loader.vkCmdEndRenderPass, uintptr(cmd))
	syscallN(b.loader.vkEndCommandBuffer, uintptr(cmd))

	// Submit
	b.submitCommandBuffer(cmd, imageIndex)
	b.lastImageIndex = imageIndex
	b.lastFrameValid = true

	b.currentFrame = (b.currentFrame + 1) % len(b.inFlightFences)
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
	// Reset the fence right before submit so it will be signaled when the GPU finishes.
	// This is done here (not in BeginFrame) so that if AcquireNextImageKHR fails,
	// the fence remains signaled and the next BeginFrame won't deadlock.
	syscallN(b.loader.vkResetFences,
		uintptr(b.device), 1, uintptr(unsafe.Pointer(&fence)),
	)
	r, _, _ := syscallN(b.loader.vkQueueSubmit,
		uintptr(b.graphicsQueue), 1, uintptr(unsafe.Pointer(&si)), uintptr(fence),
	)
	if Result(r) != Success {
		if Result(r) == ErrorDeviceLost {
			fmt.Printf("[vk] FATAL: VK_ERROR_DEVICE_LOST at vkQueueSubmit (slot %d) — GPU crashed, rendering disabled\n", b.currentFrame)
			b.deviceLost = true
			return
		}
		// Submit failed — the fence was reset but will never be signaled.
		// Recreate the fence in signaled state to prevent deadlock.
		fmt.Printf("[vk] ERROR: vkQueueSubmit failed: %v, recovering fence\n", Result(r))
		b.loader.DeviceWaitIdle(b.device)
		syscallN(b.loader.vkDestroyFence, uintptr(b.device), uintptr(fence), 0)
		fi := fenceCreateInfo{
			SType: StructureTypeFenceCreateInfo,
			Flags: FenceCreateSignaledBit,
		}
		syscallN(b.loader.vkCreateFence,
			uintptr(b.device), uintptr(unsafe.Pointer(&fi)), 0,
			uintptr(unsafe.Pointer(&b.inFlightFences[b.currentFrame])),
		)
		return
	}

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
	r, _, _ = syscallN(b.loader.vkQueuePresentKHR,
		uintptr(b.presentQueue), uintptr(unsafe.Pointer(&pi)),
	)
	if Result(r) == ErrorOutOfDateKHR || Result(r) == SuboptimalKHR {
		b.resizePending = true
	}
}

func (b *Backend) submitEmpty() {
	if b.resizePending {
		b.recreateSwapchain()
		b.resizePending = false
	}

	imageIndex, result := b.loader.AcquireNextImageKHR(
		b.device, b.swapchain.handle, MaxTimeout,
		b.imageAvailableSems[b.currentFrame], 0,
	)
	if result == ErrorOutOfDateKHR {
		b.resizePending = true
		return
	}
	if result == NotReady || result == Timeout {
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
	b.lastImageIndex = imageIndex
	b.lastFrameValid = true

	b.currentFrame = (b.currentFrame + 1) % len(b.inFlightFences)
}

// Resize implements render.Backend.
// This is lazy — it only records the new size and sets a flag.
// The actual swapchain recreation happens in Submit when AcquireNextImageKHR
// returns ErrorOutOfDateKHR, or at the start of Submit if the flag is set.
// This avoids expensive DeviceWaitIdle + swapchain recreation on every WM_SIZE
// during rapid window resizing.
func (b *Backend) Resize(width, height int) {
	if width <= 0 || height <= 0 {
		return
	}
	b.width = width
	b.height = height
	b.resizePending = true
}

func (b *Backend) recreateSwapchain() {
	b.loader.DeviceWaitIdle(b.device)

	// Fully destroy old swapchain resources BEFORE creating a new one.
	// Passing oldSwapchain to vkCreateSwapchainKHR lets the driver reuse
	// internal state, but on AMD Windows this causes vkAcquireNextImageKHR
	// to block for seconds after resize (images stuck in presentation engine).
	// Destroying first forces a clean break.
	if b.swapchain != nil {
		for _, fb := range b.swapchain.framebuffers {
			if fb != 0 {
				syscallN(b.loader.vkDestroyFramebuffer, uintptr(b.device), uintptr(fb), 0)
			}
		}
		for _, view := range b.swapchain.imageViews {
			if view != 0 {
				syscallN(b.loader.vkDestroyImageView, uintptr(b.device), uintptr(view), 0)
			}
		}
		syscallN(b.loader.vkDestroySwapchainKHR, uintptr(b.device), uintptr(b.swapchain.handle), 0)
		b.swapchain = nil
	}

	b.createSwapchain()
	b.loader.CreateFramebuffers(b.device, b.renderPass, b.swapchain)
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

	sampler, err := b.createTextureSamplerWithFilter(desc.Filter)
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

// ReadPixels implements render.Backend.
// Reads back the last rendered swapchain image as an RGBA image.
func (b *Backend) ReadPixels() (*image.RGBA, error) {
	if !b.lastFrameValid || b.swapchain == nil {
		return nil, fmt.Errorf("vulkan: no frame has been rendered yet")
	}

	// Wait for GPU to finish all work
	b.loader.DeviceWaitIdle(b.device)

	w := int(b.swapchain.extent.Width)
	h := int(b.swapchain.extent.Height)
	if w == 0 || h == 0 {
		return nil, fmt.Errorf("vulkan: swapchain extent is zero")
	}

	// 4 bytes per pixel (BGRA or RGBA)
	dataSize := uint64(w * h * 4)

	// Create host-visible staging buffer for readback
	stagingBuf, stagingMem, err := b.createBuffer(
		dataSize,
		BufferUsageTransferDstBit,
		MemoryPropertyHostVisibleBit|MemoryPropertyHostCoherentBit,
	)
	if err != nil {
		return nil, fmt.Errorf("vulkan: readback staging buffer: %w", err)
	}
	defer b.destroyBuffer(stagingBuf, stagingMem)

	srcImage := b.swapchain.images[b.lastImageIndex]

	cmd, err := b.beginOneTimeCommands()
	if err != nil {
		return nil, err
	}

	// Transition swapchain image: PresentSrc → TransferSrc
	b.transitionImageLayout(cmd, srcImage,
		ImageLayoutPresentSrcKHR, ImageLayoutTransferSrcOptimal,
		PipelineStageColorAttachmentOutputBit, PipelineStageTransferBit,
		0, AccessTransferReadBit,
	)

	// Copy image to buffer
	region := bufferImageCopy{
		ImageSubresource: imageSubresourceLayers{
			AspectMask: ImageAspectColorBit,
			LayerCount: 1,
		},
		ImageExtent: Extent3D{
			Width:  uint32(w),
			Height: uint32(h),
			Depth:  1,
		},
	}
	syscallN(b.loader.vkCmdCopyImageToBuffer,
		uintptr(cmd), uintptr(srcImage),
		uintptr(ImageLayoutTransferSrcOptimal),
		uintptr(stagingBuf),
		1, uintptr(unsafe.Pointer(&region)),
	)

	// Transition back: TransferSrc → PresentSrc
	b.transitionImageLayout(cmd, srcImage,
		ImageLayoutTransferSrcOptimal, ImageLayoutPresentSrcKHR,
		PipelineStageTransferBit, PipelineStageColorAttachmentOutputBit,
		AccessTransferReadBit, 0,
	)

	b.endOneTimeCommands(cmd)

	// Map staging buffer and read data
	var mapped unsafe.Pointer
	syscallN(b.loader.vkMapMemory,
		uintptr(b.device), uintptr(stagingMem), 0, uintptr(dataSize), 0,
		uintptr(unsafe.Pointer(&mapped)),
	)
	raw := make([]byte, dataSize)
	copy(raw, unsafe.Slice((*byte)(mapped), dataSize))
	syscallN(b.loader.vkUnmapMemory, uintptr(b.device), uintptr(stagingMem))

	// Convert to image.RGBA (swapchain is typically BGRA, need to swap R and B)
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	isBGRA := b.swapchain.format == FormatB8G8R8A8Srgb || b.swapchain.format == FormatB8G8R8A8Unorm

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			offset := (y*w + x) * 4
			r, g, bl, a := raw[offset], raw[offset+1], raw[offset+2], raw[offset+3]
			if isBGRA {
				r, bl = bl, r // swap B and R
			}
			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: bl, A: a})
		}
	}

	return img, nil
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

	// Destroy vertex buffers (including stale ones from resizes)
	for i := range b.vertexBuffers {
		b.destroyBuffer(b.vertexBuffers[i], b.vertexMemory[i])
	}
	for f := range b.staleVertexBuffers {
		for i, buf := range b.staleVertexBuffers[f] {
			b.destroyBuffer(buf, b.staleVertexMemory[f][i])
		}
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
	for i := range b.inFlightFences {
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

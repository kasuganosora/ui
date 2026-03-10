package vulkan

import (
	"fmt"
	"unsafe"
)

// Swapchain manages the Vulkan swapchain and associated resources.
type Swapchain struct {
	handle       SwapchainKHR
	images       []Image
	imageViews   []ImageView
	framebuffers []Framebuffer
	format       Format
	extent       Extent2D
	imageCount   uint32
}

// swapchainCreateInfoKHR for vkCreateSwapchainKHR.
type swapchainCreateInfoKHR struct {
	SType                 StructureType
	PNext                 unsafe.Pointer
	Flags                 uint32
	Surface               SurfaceKHR
	MinImageCount         uint32
	ImageFormat           Format
	ImageColorSpace       ColorSpaceKHR
	ImageExtent           Extent2D
	ImageArrayLayers      uint32
	ImageUsage            ImageUsageFlags
	ImageSharingMode      SharingMode
	QueueFamilyIndexCount uint32
	PQueueFamilyIndices   *uint32
	PreTransform          SurfaceTransformFlagBitsKHR
	CompositeAlpha        CompositeAlphaFlagBitsKHR
	PresentMode           PresentModeKHR
	Clipped               uint32
	OldSwapchain          SwapchainKHR
}

// CreateSwapchain creates a swapchain.
func (l *Loader) CreateSwapchain(
	device Device,
	surface SurfaceKHR,
	caps SurfaceCapabilitiesKHR,
	format SurfaceFormatKHR,
	presentMode PresentModeKHR,
	width, height uint32,
	indices QueueFamilyIndices,
	oldSwapchain SwapchainKHR,
) (*Swapchain, error) {
	// Determine image count
	imageCount := caps.MinImageCount + 1
	if caps.MaxImageCount > 0 && imageCount > caps.MaxImageCount {
		imageCount = caps.MaxImageCount
	}

	// Determine extent
	extent := Extent2D{Width: width, Height: height}
	if caps.CurrentExtent.Width != 0xFFFFFFFF {
		extent = caps.CurrentExtent
	} else {
		if extent.Width < caps.MinImageExtent.Width {
			extent.Width = caps.MinImageExtent.Width
		}
		if extent.Width > caps.MaxImageExtent.Width {
			extent.Width = caps.MaxImageExtent.Width
		}
		if extent.Height < caps.MinImageExtent.Height {
			extent.Height = caps.MinImageExtent.Height
		}
		if extent.Height > caps.MaxImageExtent.Height {
			extent.Height = caps.MaxImageExtent.Height
		}
	}

	createInfo := swapchainCreateInfoKHR{
		SType:            StructureTypeSwapchainCreateInfoKHR,
		Surface:          surface,
		MinImageCount:    imageCount,
		ImageFormat:      format.Format,
		ImageColorSpace:  format.ColorSpace,
		ImageExtent:      extent,
		ImageArrayLayers: 1,
		ImageUsage:       ImageUsageColorAttachmentBit | ImageUsageTransferSrcBit,
		PreTransform:     caps.CurrentTransform,
		CompositeAlpha:   CompositeAlphaOpaqueBitKHR,
		PresentMode:      presentMode,
		Clipped:          1, // VK_TRUE
		OldSwapchain:     oldSwapchain,
	}

	queueFamilyIndices := [2]uint32{uint32(indices.Graphics), uint32(indices.Present)}
	if indices.Graphics != indices.Present {
		createInfo.ImageSharingMode = SharingModeConcurrent
		createInfo.QueueFamilyIndexCount = 2
		createInfo.PQueueFamilyIndices = &queueFamilyIndices[0]
	} else {
		createInfo.ImageSharingMode = SharingModeExclusive
	}

	var swapchain SwapchainKHR
	r, _, _ := syscallN(l.vkCreateSwapchainKHR,
		uintptr(device),
		uintptr(unsafe.Pointer(&createInfo)),
		0,
		uintptr(unsafe.Pointer(&swapchain)),
	)
	if Result(r) != Success {
		return nil, fmt.Errorf("vulkan: vkCreateSwapchainKHR failed: %v", Result(r))
	}

	// Get swapchain images
	var count uint32
	syscallN(l.vkGetSwapchainImagesKHR,
		uintptr(device), uintptr(swapchain), uintptr(unsafe.Pointer(&count)), 0,
	)
	images := make([]Image, count)
	syscallN(l.vkGetSwapchainImagesKHR,
		uintptr(device), uintptr(swapchain), uintptr(unsafe.Pointer(&count)), uintptr(unsafe.Pointer(&images[0])),
	)

	return &Swapchain{
		handle:     swapchain,
		images:     images,
		format:     format.Format,
		extent:     extent,
		imageCount: count,
	}, nil
}

// imageViewCreateInfo for vkCreateImageView.
type imageViewCreateInfo struct {
	SType            StructureType
	PNext            unsafe.Pointer
	Flags            uint32
	Image            Image
	ViewType         ImageViewType
	Format           Format
	Components       [4]ComponentSwizzle // r,g,b,a
	SubresourceRange imageSubresourceRange
}

type imageSubresourceRange struct {
	AspectMask     ImageAspectFlags
	BaseMipLevel   uint32
	LevelCount     uint32
	BaseArrayLayer uint32
	LayerCount     uint32
}

// CreateImageViews creates image views for all swapchain images.
func (l *Loader) CreateImageViews(device Device, sc *Swapchain) error {
	sc.imageViews = make([]ImageView, len(sc.images))
	for i, img := range sc.images {
		createInfo := imageViewCreateInfo{
			SType:    StructureTypeImageViewCreateInfo,
			Image:    img,
			ViewType: ImageViewType2D,
			Format:   sc.format,
			SubresourceRange: imageSubresourceRange{
				AspectMask:   ImageAspectColorBit,
				BaseMipLevel: 0,
				LevelCount:   1,
				LayerCount:   1,
			},
		}
		var view ImageView
		r, _, _ := syscallN(l.vkCreateImageView,
			uintptr(device),
			uintptr(unsafe.Pointer(&createInfo)),
			0,
			uintptr(unsafe.Pointer(&view)),
		)
		if Result(r) != Success {
			return fmt.Errorf("vulkan: vkCreateImageView failed for image %d: %v", i, Result(r))
		}
		sc.imageViews[i] = view
	}
	return nil
}

// framebufferCreateInfo for vkCreateFramebuffer.
type framebufferCreateInfo struct {
	SType           StructureType
	PNext           unsafe.Pointer
	Flags           uint32
	RenderPass      RenderPass
	AttachmentCount uint32
	PAttachments    *ImageView
	Width           uint32
	Height          uint32
	Layers          uint32
}

// CreateFramebuffers creates framebuffers for all swapchain image views.
func (l *Loader) CreateFramebuffers(device Device, renderPass RenderPass, sc *Swapchain) error {
	sc.framebuffers = make([]Framebuffer, len(sc.imageViews))
	for i, view := range sc.imageViews {
		createInfo := framebufferCreateInfo{
			SType:           StructureTypeFramebufferCreateInfo,
			RenderPass:      renderPass,
			AttachmentCount: 1,
			PAttachments:    &view,
			Width:           sc.extent.Width,
			Height:          sc.extent.Height,
			Layers:          1,
		}
		var fb Framebuffer
		r, _, _ := syscallN(l.vkCreateFramebuffer,
			uintptr(device),
			uintptr(unsafe.Pointer(&createInfo)),
			0,
			uintptr(unsafe.Pointer(&fb)),
		)
		if Result(r) != Success {
			return fmt.Errorf("vulkan: vkCreateFramebuffer failed for view %d: %v", i, Result(r))
		}
		sc.framebuffers[i] = fb
	}
	return nil
}

// DestroySwapchain destroys swapchain and associated resources.
func (l *Loader) DestroySwapchain(device Device, sc *Swapchain) {
	for _, fb := range sc.framebuffers {
		if fb != 0 {
			syscallN(l.vkDestroyFramebuffer, uintptr(device), uintptr(fb), 0)
		}
	}
	for _, view := range sc.imageViews {
		if view != 0 {
			syscallN(l.vkDestroyImageView, uintptr(device), uintptr(view), 0)
		}
	}
	if sc.handle != 0 {
		syscallN(l.vkDestroySwapchainKHR, uintptr(device), uintptr(sc.handle), 0)
	}
}

// AcquireNextImageKHR acquires the next swapchain image.
func (l *Loader) AcquireNextImageKHR(device Device, swapchain SwapchainKHR, timeout uint64, semaphore Semaphore, fence Fence) (uint32, Result) {
	var imageIndex uint32
	r, _, _ := syscallN(l.vkAcquireNextImageKHR,
		uintptr(device),
		uintptr(swapchain),
		uintptr(timeout),
		uintptr(semaphore),
		uintptr(fence),
		uintptr(unsafe.Pointer(&imageIndex)),
	)
	return imageIndex, Result(r)
}

// ChooseSurfaceFormat picks the best surface format.
// Prefer sRGB for correct linear-space alpha blending.
func ChooseSurfaceFormat(formats []SurfaceFormatKHR) SurfaceFormatKHR {
	for _, f := range formats {
		if f.Format == FormatB8G8R8A8Srgb && f.ColorSpace == ColorSpaceSrgbNonlinearKHR {
			return f
		}
	}
	// Prefer B8G8R8A8_UNORM as fallback (native format on most Windows GPUs)
	for _, f := range formats {
		if f.Format == FormatB8G8R8A8Unorm {
			return f
		}
	}
	return formats[0]
}

// ChoosePresentMode picks the best present mode.
// When vsync is off, prefer IMMEDIATE > MAILBOX > FIFO.
// IMMEDIATE avoids the compositor holding swapchain images during/after
// Win32 modal resize, which causes AcquireNextImageKHR to block for seconds
// on some drivers (e.g. AMD on Windows).
func ChoosePresentMode(modes []PresentModeKHR, vsync bool) PresentModeKHR {
	if vsync {
		return PresentModeFifoKHR
	}
	// Prefer IMMEDIATE — images are returned immediately after scanout,
	// avoiding compositor-induced AcquireNextImageKHR stalls.
	for _, m := range modes {
		if m == PresentModeImmediateKHR {
			return m
		}
	}
	for _, m := range modes {
		if m == PresentModeMailboxKHR {
			return m
		}
	}
	return PresentModeFifoKHR
}

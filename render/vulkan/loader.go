package vulkan

import (
	"fmt"
	"unsafe"
)

// Loader dynamically loads Vulkan functions.
// This is the anti-corruption layer between the C Vulkan API and Go.
type Loader struct {
	lib platformLib // Platform-specific library handle

	// Global functions (loaded from library)
	vkGetInstanceProcAddr           uintptr
	vkCreateInstance                uintptr
	vkEnumerateInstanceLayerProperties uintptr
	vkEnumerateInstanceExtensionProperties uintptr

	// Instance functions (loaded after instance creation)
	vkDestroyInstance                      uintptr
	vkEnumeratePhysicalDevices             uintptr
	vkGetPhysicalDeviceProperties          uintptr
	vkGetPhysicalDeviceFeatures            uintptr
	vkGetPhysicalDeviceQueueFamilyProperties uintptr
	vkGetPhysicalDeviceMemoryProperties    uintptr
	vkGetPhysicalDeviceSurfaceSupportKHR   uintptr
	vkGetPhysicalDeviceSurfaceCapabilitiesKHR uintptr
	vkGetPhysicalDeviceSurfaceFormatsKHR   uintptr
	vkGetPhysicalDeviceSurfacePresentModesKHR uintptr
	vkCreateDevice                         uintptr
	vkDestroySurfaceKHR                    uintptr

	// Platform surface creation (loaded per-platform)
	vkCreateWin32SurfaceKHR   uintptr
	vkCreateXlibSurfaceKHR    uintptr

	// Device functions (loaded after device creation)
	vkDestroyDevice             uintptr
	vkGetDeviceQueue            uintptr
	vkDeviceWaitIdle            uintptr
	vkCreateSwapchainKHR        uintptr
	vkDestroySwapchainKHR       uintptr
	vkGetSwapchainImagesKHR     uintptr
	vkAcquireNextImageKHR       uintptr
	vkQueuePresentKHR           uintptr
	vkQueueSubmit               uintptr
	vkQueueWaitIdle             uintptr
	vkCreateImageView           uintptr
	vkDestroyImageView          uintptr
	vkCreateRenderPass          uintptr
	vkDestroyRenderPass         uintptr
	vkCreateFramebuffer         uintptr
	vkDestroyFramebuffer        uintptr
	vkCreateShaderModule        uintptr
	vkDestroyShaderModule       uintptr
	vkCreatePipelineLayout      uintptr
	vkDestroyPipelineLayout     uintptr
	vkCreateGraphicsPipelines   uintptr
	vkDestroyPipeline           uintptr
	vkCreateCommandPool         uintptr
	vkDestroyCommandPool        uintptr
	vkAllocateCommandBuffers    uintptr
	vkFreeCommandBuffers        uintptr
	vkBeginCommandBuffer        uintptr
	vkEndCommandBuffer          uintptr
	vkResetCommandBuffer        uintptr
	vkCmdBeginRenderPass        uintptr
	vkCmdEndRenderPass          uintptr
	vkCmdBindPipeline           uintptr
	vkCmdSetViewport            uintptr
	vkCmdSetScissor             uintptr
	vkCmdDraw                   uintptr
	vkCmdDrawIndexed            uintptr
	vkCmdBindVertexBuffers      uintptr
	vkCmdBindIndexBuffer        uintptr
	vkCmdBindDescriptorSets     uintptr
	vkCmdPushConstants          uintptr
	vkCmdPipelineBarrier        uintptr
	vkCmdCopyBufferToImage      uintptr
	vkCmdCopyImageToBuffer      uintptr
	vkCreateFence               uintptr
	vkDestroyFence              uintptr
	vkWaitForFences             uintptr
	vkResetFences               uintptr
	vkCreateSemaphore           uintptr
	vkDestroySemaphore          uintptr
	vkCreateBuffer              uintptr
	vkDestroyBuffer             uintptr
	vkGetBufferMemoryRequirements uintptr
	vkAllocateMemory            uintptr
	vkFreeMemory                uintptr
	vkBindBufferMemory          uintptr
	vkMapMemory                 uintptr
	vkUnmapMemory               uintptr
	vkFlushMappedMemoryRanges   uintptr
	vkCreateImage               uintptr
	vkDestroyImage              uintptr
	vkGetImageMemoryRequirements uintptr
	vkBindImageMemory           uintptr
	vkCreateSampler             uintptr
	vkDestroySampler            uintptr
	vkCreateDescriptorSetLayout uintptr
	vkDestroyDescriptorSetLayout uintptr
	vkCreateDescriptorPool      uintptr
	vkDestroyDescriptorPool     uintptr
	vkAllocateDescriptorSets    uintptr
	vkUpdateDescriptorSets      uintptr
}

// NewLoader creates a new Vulkan loader by loading the Vulkan library.
func NewLoader() (*Loader, error) {
	l := &Loader{}
	var err error
	l.lib, err = openVulkanLib()
	if err != nil {
		return nil, fmt.Errorf("vulkan: failed to load library: %w", err)
	}

	// Load the bootstrap function
	l.vkGetInstanceProcAddr, err = l.lib.lookup("vkGetInstanceProcAddr")
	if err != nil {
		l.lib.close()
		return nil, fmt.Errorf("vulkan: vkGetInstanceProcAddr not found: %w", err)
	}

	// Load global functions via vkGetInstanceProcAddr(NULL, ...)
	l.vkCreateInstance = l.getInstanceProcAddr(0, "vkCreateInstance")
	l.vkEnumerateInstanceLayerProperties = l.getInstanceProcAddr(0, "vkEnumerateInstanceLayerProperties")
	l.vkEnumerateInstanceExtensionProperties = l.getInstanceProcAddr(0, "vkEnumerateInstanceExtensionProperties")

	return l, nil
}

// LoadInstanceFunctions loads all instance-level Vulkan functions.
func (l *Loader) LoadInstanceFunctions(instance Instance) {
	inst := uintptr(instance)
	l.vkDestroyInstance = l.getInstanceProcAddr(inst, "vkDestroyInstance")
	l.vkEnumeratePhysicalDevices = l.getInstanceProcAddr(inst, "vkEnumeratePhysicalDevices")
	l.vkGetPhysicalDeviceProperties = l.getInstanceProcAddr(inst, "vkGetPhysicalDeviceProperties")
	l.vkGetPhysicalDeviceFeatures = l.getInstanceProcAddr(inst, "vkGetPhysicalDeviceFeatures")
	l.vkGetPhysicalDeviceQueueFamilyProperties = l.getInstanceProcAddr(inst, "vkGetPhysicalDeviceQueueFamilyProperties")
	l.vkGetPhysicalDeviceMemoryProperties = l.getInstanceProcAddr(inst, "vkGetPhysicalDeviceMemoryProperties")
	l.vkGetPhysicalDeviceSurfaceSupportKHR = l.getInstanceProcAddr(inst, "vkGetPhysicalDeviceSurfaceSupportKHR")
	l.vkGetPhysicalDeviceSurfaceCapabilitiesKHR = l.getInstanceProcAddr(inst, "vkGetPhysicalDeviceSurfaceCapabilitiesKHR")
	l.vkGetPhysicalDeviceSurfaceFormatsKHR = l.getInstanceProcAddr(inst, "vkGetPhysicalDeviceSurfaceFormatsKHR")
	l.vkGetPhysicalDeviceSurfacePresentModesKHR = l.getInstanceProcAddr(inst, "vkGetPhysicalDeviceSurfacePresentModesKHR")
	l.vkCreateDevice = l.getInstanceProcAddr(inst, "vkCreateDevice")
	l.vkDestroySurfaceKHR = l.getInstanceProcAddr(inst, "vkDestroySurfaceKHR")

	// Platform surface extensions
	l.vkCreateWin32SurfaceKHR = l.getInstanceProcAddr(inst, "vkCreateWin32SurfaceKHR")
	l.vkCreateXlibSurfaceKHR = l.getInstanceProcAddr(inst, "vkCreateXlibSurfaceKHR")
}

// LoadDeviceFunctions loads all device-level Vulkan functions.
func (l *Loader) LoadDeviceFunctions(instance Instance) {
	inst := uintptr(instance)
	l.vkDestroyDevice = l.getInstanceProcAddr(inst, "vkDestroyDevice")
	l.vkGetDeviceQueue = l.getInstanceProcAddr(inst, "vkGetDeviceQueue")
	l.vkDeviceWaitIdle = l.getInstanceProcAddr(inst, "vkDeviceWaitIdle")
	l.vkCreateSwapchainKHR = l.getInstanceProcAddr(inst, "vkCreateSwapchainKHR")
	l.vkDestroySwapchainKHR = l.getInstanceProcAddr(inst, "vkDestroySwapchainKHR")
	l.vkGetSwapchainImagesKHR = l.getInstanceProcAddr(inst, "vkGetSwapchainImagesKHR")
	l.vkAcquireNextImageKHR = l.getInstanceProcAddr(inst, "vkAcquireNextImageKHR")
	l.vkQueuePresentKHR = l.getInstanceProcAddr(inst, "vkQueuePresentKHR")
	l.vkQueueSubmit = l.getInstanceProcAddr(inst, "vkQueueSubmit")
	l.vkQueueWaitIdle = l.getInstanceProcAddr(inst, "vkQueueWaitIdle")
	l.vkCreateImageView = l.getInstanceProcAddr(inst, "vkCreateImageView")
	l.vkDestroyImageView = l.getInstanceProcAddr(inst, "vkDestroyImageView")
	l.vkCreateRenderPass = l.getInstanceProcAddr(inst, "vkCreateRenderPass")
	l.vkDestroyRenderPass = l.getInstanceProcAddr(inst, "vkDestroyRenderPass")
	l.vkCreateFramebuffer = l.getInstanceProcAddr(inst, "vkCreateFramebuffer")
	l.vkDestroyFramebuffer = l.getInstanceProcAddr(inst, "vkDestroyFramebuffer")
	l.vkCreateShaderModule = l.getInstanceProcAddr(inst, "vkCreateShaderModule")
	l.vkDestroyShaderModule = l.getInstanceProcAddr(inst, "vkDestroyShaderModule")
	l.vkCreatePipelineLayout = l.getInstanceProcAddr(inst, "vkCreatePipelineLayout")
	l.vkDestroyPipelineLayout = l.getInstanceProcAddr(inst, "vkDestroyPipelineLayout")
	l.vkCreateGraphicsPipelines = l.getInstanceProcAddr(inst, "vkCreateGraphicsPipelines")
	l.vkDestroyPipeline = l.getInstanceProcAddr(inst, "vkDestroyPipeline")
	l.vkCreateCommandPool = l.getInstanceProcAddr(inst, "vkCreateCommandPool")
	l.vkDestroyCommandPool = l.getInstanceProcAddr(inst, "vkDestroyCommandPool")
	l.vkAllocateCommandBuffers = l.getInstanceProcAddr(inst, "vkAllocateCommandBuffers")
	l.vkFreeCommandBuffers = l.getInstanceProcAddr(inst, "vkFreeCommandBuffers")
	l.vkBeginCommandBuffer = l.getInstanceProcAddr(inst, "vkBeginCommandBuffer")
	l.vkEndCommandBuffer = l.getInstanceProcAddr(inst, "vkEndCommandBuffer")
	l.vkResetCommandBuffer = l.getInstanceProcAddr(inst, "vkResetCommandBuffer")
	l.vkCmdBeginRenderPass = l.getInstanceProcAddr(inst, "vkCmdBeginRenderPass")
	l.vkCmdEndRenderPass = l.getInstanceProcAddr(inst, "vkCmdEndRenderPass")
	l.vkCmdBindPipeline = l.getInstanceProcAddr(inst, "vkCmdBindPipeline")
	l.vkCmdSetViewport = l.getInstanceProcAddr(inst, "vkCmdSetViewport")
	l.vkCmdSetScissor = l.getInstanceProcAddr(inst, "vkCmdSetScissor")
	l.vkCmdDraw = l.getInstanceProcAddr(inst, "vkCmdDraw")
	l.vkCmdDrawIndexed = l.getInstanceProcAddr(inst, "vkCmdDrawIndexed")
	l.vkCmdBindVertexBuffers = l.getInstanceProcAddr(inst, "vkCmdBindVertexBuffers")
	l.vkCmdBindIndexBuffer = l.getInstanceProcAddr(inst, "vkCmdBindIndexBuffer")
	l.vkCmdBindDescriptorSets = l.getInstanceProcAddr(inst, "vkCmdBindDescriptorSets")
	l.vkCmdPushConstants = l.getInstanceProcAddr(inst, "vkCmdPushConstants")
	l.vkCmdPipelineBarrier = l.getInstanceProcAddr(inst, "vkCmdPipelineBarrier")
	l.vkCmdCopyBufferToImage = l.getInstanceProcAddr(inst, "vkCmdCopyBufferToImage")
	l.vkCmdCopyImageToBuffer = l.getInstanceProcAddr(inst, "vkCmdCopyImageToBuffer")
	l.vkCreateFence = l.getInstanceProcAddr(inst, "vkCreateFence")
	l.vkDestroyFence = l.getInstanceProcAddr(inst, "vkDestroyFence")
	l.vkWaitForFences = l.getInstanceProcAddr(inst, "vkWaitForFences")
	l.vkResetFences = l.getInstanceProcAddr(inst, "vkResetFences")
	l.vkCreateSemaphore = l.getInstanceProcAddr(inst, "vkCreateSemaphore")
	l.vkDestroySemaphore = l.getInstanceProcAddr(inst, "vkDestroySemaphore")
	l.vkCreateBuffer = l.getInstanceProcAddr(inst, "vkCreateBuffer")
	l.vkDestroyBuffer = l.getInstanceProcAddr(inst, "vkDestroyBuffer")
	l.vkGetBufferMemoryRequirements = l.getInstanceProcAddr(inst, "vkGetBufferMemoryRequirements")
	l.vkAllocateMemory = l.getInstanceProcAddr(inst, "vkAllocateMemory")
	l.vkFreeMemory = l.getInstanceProcAddr(inst, "vkFreeMemory")
	l.vkBindBufferMemory = l.getInstanceProcAddr(inst, "vkBindBufferMemory")
	l.vkMapMemory = l.getInstanceProcAddr(inst, "vkMapMemory")
	l.vkUnmapMemory = l.getInstanceProcAddr(inst, "vkUnmapMemory")
	l.vkFlushMappedMemoryRanges = l.getInstanceProcAddr(inst, "vkFlushMappedMemoryRanges")
	l.vkCreateImage = l.getInstanceProcAddr(inst, "vkCreateImage")
	l.vkDestroyImage = l.getInstanceProcAddr(inst, "vkDestroyImage")
	l.vkGetImageMemoryRequirements = l.getInstanceProcAddr(inst, "vkGetImageMemoryRequirements")
	l.vkBindImageMemory = l.getInstanceProcAddr(inst, "vkBindImageMemory")
	l.vkCreateSampler = l.getInstanceProcAddr(inst, "vkCreateSampler")
	l.vkDestroySampler = l.getInstanceProcAddr(inst, "vkDestroySampler")
	l.vkCreateDescriptorSetLayout = l.getInstanceProcAddr(inst, "vkCreateDescriptorSetLayout")
	l.vkDestroyDescriptorSetLayout = l.getInstanceProcAddr(inst, "vkDestroyDescriptorSetLayout")
	l.vkCreateDescriptorPool = l.getInstanceProcAddr(inst, "vkCreateDescriptorPool")
	l.vkDestroyDescriptorPool = l.getInstanceProcAddr(inst, "vkDestroyDescriptorPool")
	l.vkAllocateDescriptorSets = l.getInstanceProcAddr(inst, "vkAllocateDescriptorSets")
	l.vkUpdateDescriptorSets = l.getInstanceProcAddr(inst, "vkUpdateDescriptorSets")
}

// Close releases the loaded library.
func (l *Loader) Close() {
	if l.lib != nil {
		l.lib.close()
	}
}

// getInstanceProcAddr calls vkGetInstanceProcAddr to look up a function.
func (l *Loader) getInstanceProcAddr(instance uintptr, name string) uintptr {
	nameBytes := append([]byte(name), 0) // null-terminated
	ret, _, _ := syscall3(l.vkGetInstanceProcAddr, instance, uintptr(unsafe.Pointer(&nameBytes[0])), 0)
	return ret
}

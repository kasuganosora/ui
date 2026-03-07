package vulkan

import (
	"fmt"
	"unsafe"
)

// QueueFamilyIndices holds the queue family indices we need.
type QueueFamilyIndices struct {
	Graphics int
	Present  int
}

// FindQueueFamilies finds queue families that support graphics and presentation.
func (l *Loader) FindQueueFamilies(device PhysicalDevice, surface SurfaceKHR) (QueueFamilyIndices, bool) {
	indices := QueueFamilyIndices{Graphics: -1, Present: -1}
	families := l.GetPhysicalDeviceQueueFamilyProperties(device)

	for i, fam := range families {
		if fam.QueueFlags&QueueGraphicsBit != 0 {
			indices.Graphics = i
		}
		if l.GetPhysicalDeviceSurfaceSupportKHR(device, uint32(i), surface) {
			indices.Present = i
		}
		if indices.Graphics >= 0 && indices.Present >= 0 {
			return indices, true
		}
	}
	return indices, false
}

// deviceQueueCreateInfo for vkCreateDevice.
type deviceQueueCreateInfo struct {
	SType            StructureType
	PNext            unsafe.Pointer
	Flags            uint32
	QueueFamilyIndex uint32
	QueueCount       uint32
	PQueuePriorities *float32
}

// deviceCreateInfo for vkCreateDevice.
type deviceCreateInfo struct {
	SType                   StructureType
	PNext                   unsafe.Pointer
	Flags                   uint32
	QueueCreateInfoCount    uint32
	PQueueCreateInfos       *deviceQueueCreateInfo
	EnabledLayerCount       uint32
	PPEnabledLayerNames     **byte
	EnabledExtensionCount   uint32
	PPEnabledExtensionNames **byte
	PEnabledFeatures        unsafe.Pointer
}

// CreateDevice creates a logical device.
func (l *Loader) CreateDevice(physDevice PhysicalDevice, indices QueueFamilyIndices) (Device, error) {
	queuePriority := float32(1.0)

	// Unique queue families
	uniqueFamilies := map[int]bool{
		indices.Graphics: true,
		indices.Present:  true,
	}

	queueCreateInfos := make([]deviceQueueCreateInfo, 0, len(uniqueFamilies))
	for family := range uniqueFamilies {
		queueCreateInfos = append(queueCreateInfos, deviceQueueCreateInfo{
			SType:            StructureTypeDeviceQueueCreateInfo,
			QueueFamilyIndex: uint32(family),
			QueueCount:       1,
			PQueuePriorities: &queuePriority,
		})
	}

	// Required device extensions
	deviceExtensions := []string{"VK_KHR_swapchain"}
	extPtrs := cstrArray(deviceExtensions)

	createInfo := deviceCreateInfo{
		SType:                 StructureTypeDeviceCreateInfo,
		QueueCreateInfoCount:  uint32(len(queueCreateInfos)),
		PQueueCreateInfos:     &queueCreateInfos[0],
		EnabledExtensionCount: uint32(len(deviceExtensions)),
	}
	if len(extPtrs) > 0 {
		createInfo.PPEnabledExtensionNames = &extPtrs[0]
	}

	var device Device
	r, _, _ := syscallN(l.vkCreateDevice,
		uintptr(physDevice),
		uintptr(unsafe.Pointer(&createInfo)),
		0, // allocator
		uintptr(unsafe.Pointer(&device)),
	)
	if Result(r) != Success {
		return 0, fmt.Errorf("vulkan: vkCreateDevice failed: %v", Result(r))
	}
	return device, nil
}

// DestroyDevice destroys a logical device.
func (l *Loader) DestroyDevice(device Device) {
	if l.vkDestroyDevice != 0 {
		syscallN(l.vkDestroyDevice, uintptr(device), 0)
	}
}

// GetDeviceQueue retrieves a queue handle.
func (l *Loader) GetDeviceQueue(device Device, familyIndex, queueIndex uint32) Queue {
	var queue Queue
	syscallN(l.vkGetDeviceQueue,
		uintptr(device),
		uintptr(familyIndex),
		uintptr(queueIndex),
		uintptr(unsafe.Pointer(&queue)),
	)
	return queue
}

// DeviceWaitIdle waits for the device to finish all work.
func (l *Loader) DeviceWaitIdle(device Device) {
	syscallN(l.vkDeviceWaitIdle, uintptr(device))
}

// FindMemoryType finds a suitable memory type index.
func FindMemoryType(memProps PhysicalDeviceMemoryProperties, typeFilter uint32, properties MemoryPropertyFlags) (uint32, bool) {
	for i := uint32(0); i < memProps.MemoryTypeCount; i++ {
		if typeFilter&(1<<i) != 0 && memProps.MemoryTypes[i].PropertyFlags&properties == properties {
			return i, true
		}
	}
	return 0, false
}

package vulkan

import (
	"fmt"
	"unsafe"
)

// ApplicationInfo for vkCreateInstance.
type applicationInfo struct {
	SType              StructureType
	PNext              unsafe.Pointer
	PApplicationName   *byte
	ApplicationVersion uint32
	PEngineName        *byte
	EngineVersion      uint32
	APIVersion         uint32
}

// InstanceCreateInfo for vkCreateInstance.
type instanceCreateInfo struct {
	SType                   StructureType
	PNext                   unsafe.Pointer
	Flags                   uint32
	PApplicationInfo        *applicationInfo
	EnabledLayerCount       uint32
	PPEnabledLayerNames     **byte
	EnabledExtensionCount   uint32
	PPEnabledExtensionNames **byte
}

// CreateInstance creates a Vulkan instance.
func (l *Loader) CreateInstance(appName, engineName string, enableValidation bool) (Instance, error) {
	appNameC := cstr(appName)
	engineNameC := cstr(engineName)

	appInfo := applicationInfo{
		SType:              StructureTypeApplicationInfo,
		PApplicationName:   &appNameC[0],
		ApplicationVersion: MakeAPIVersion(1, 0, 0),
		PEngineName:        &engineNameC[0],
		EngineVersion:      MakeAPIVersion(1, 0, 0),
		APIVersion:         MakeAPIVersion(1, 0, 0),
	}

	// Extensions
	exts := requiredSurfaceExtensions()
	extPtrs := cstrArray(exts)

	// Layers
	var layerPtrs []*byte
	var layers [][]byte
	if enableValidation {
		validationLayer := cstr("VK_LAYER_KHRONOS_validation")
		layers = append(layers, validationLayer)
		layerPtrs = append(layerPtrs, &validationLayer[0])
	}

	createInfo := instanceCreateInfo{
		SType:                 StructureTypeInstanceCreateInfo,
		PApplicationInfo:      &appInfo,
		EnabledExtensionCount: uint32(len(exts)),
	}
	if len(extPtrs) > 0 {
		createInfo.PPEnabledExtensionNames = &extPtrs[0]
	}
	if len(layerPtrs) > 0 {
		createInfo.EnabledLayerCount = uint32(len(layerPtrs))
		createInfo.PPEnabledLayerNames = &layerPtrs[0]
	}

	var instance Instance
	r, _, _ := syscallN(l.vkCreateInstance,
		uintptr(unsafe.Pointer(&createInfo)),
		0, // allocator
		uintptr(unsafe.Pointer(&instance)),
	)
	if Result(r) != Success {
		return 0, fmt.Errorf("vulkan: vkCreateInstance failed: %v", Result(r))
	}

	// Keep references alive
	_ = appNameC
	_ = engineNameC
	_ = layers

	return instance, nil
}

// DestroyInstance destroys a Vulkan instance.
func (l *Loader) DestroyInstance(instance Instance) {
	if l.vkDestroyInstance != 0 {
		syscallN(l.vkDestroyInstance, uintptr(instance), 0)
	}
}

// EnumeratePhysicalDevices returns all physical devices.
func (l *Loader) EnumeratePhysicalDevices(instance Instance) ([]PhysicalDevice, error) {
	var count uint32
	r, _, _ := syscallN(l.vkEnumeratePhysicalDevices,
		uintptr(instance),
		uintptr(unsafe.Pointer(&count)),
		0,
	)
	if Result(r) != Success {
		return nil, fmt.Errorf("vulkan: vkEnumeratePhysicalDevices count failed: %v", Result(r))
	}
	if count == 0 {
		return nil, fmt.Errorf("vulkan: no physical devices found")
	}
	devices := make([]PhysicalDevice, count)
	r, _, _ = syscallN(l.vkEnumeratePhysicalDevices,
		uintptr(instance),
		uintptr(unsafe.Pointer(&count)),
		uintptr(unsafe.Pointer(&devices[0])),
	)
	if Result(r) != Success {
		return nil, fmt.Errorf("vulkan: vkEnumeratePhysicalDevices failed: %v", Result(r))
	}
	return devices[:count], nil
}

// GetPhysicalDeviceQueueFamilyProperties returns queue family properties.
func (l *Loader) GetPhysicalDeviceQueueFamilyProperties(device PhysicalDevice) []QueueFamilyProperties {
	var count uint32
	syscallN(l.vkGetPhysicalDeviceQueueFamilyProperties,
		uintptr(device),
		uintptr(unsafe.Pointer(&count)),
		0,
	)
	if count == 0 {
		return nil
	}
	props := make([]QueueFamilyProperties, count)
	syscallN(l.vkGetPhysicalDeviceQueueFamilyProperties,
		uintptr(device),
		uintptr(unsafe.Pointer(&count)),
		uintptr(unsafe.Pointer(&props[0])),
	)
	return props[:count]
}

// GetPhysicalDeviceNameAndDriver returns the device name and driver version for logging.
func (l *Loader) GetPhysicalDeviceNameAndDriver(device PhysicalDevice) (name string, apiVer, driverVer uint32) {
	// VkPhysicalDeviceProperties is ~824 bytes. We only need the first fields:
	//   uint32 apiVersion       (offset 0)
	//   uint32 driverVersion    (offset 4)
	//   uint32 vendorID         (offset 8)
	//   uint32 deviceID         (offset 12)
	//   uint32 deviceType       (offset 16)
	//   char   deviceName[256]  (offset 20)
	var buf [1024]byte
	syscallN(l.vkGetPhysicalDeviceProperties,
		uintptr(device),
		uintptr(unsafe.Pointer(&buf[0])),
	)
	apiVer = *(*uint32)(unsafe.Pointer(&buf[0]))
	driverVer = *(*uint32)(unsafe.Pointer(&buf[4]))
	// deviceName is a null-terminated C string at offset 20
	nameBytes := buf[20:276]
	for i, b := range nameBytes {
		if b == 0 {
			name = string(nameBytes[:i])
			return
		}
	}
	name = string(nameBytes)
	return
}

// GetPhysicalDeviceMemoryProperties returns memory properties.
func (l *Loader) GetPhysicalDeviceMemoryProperties(device PhysicalDevice) PhysicalDeviceMemoryProperties {
	var props PhysicalDeviceMemoryProperties
	syscallN(l.vkGetPhysicalDeviceMemoryProperties,
		uintptr(device),
		uintptr(unsafe.Pointer(&props)),
	)
	return props
}

// GetPhysicalDeviceSurfaceSupportKHR checks if a queue family supports presentation.
func (l *Loader) GetPhysicalDeviceSurfaceSupportKHR(device PhysicalDevice, queueFamily uint32, surface SurfaceKHR) bool {
	var supported uint32
	syscallN(l.vkGetPhysicalDeviceSurfaceSupportKHR,
		uintptr(device),
		uintptr(queueFamily),
		uintptr(surface),
		uintptr(unsafe.Pointer(&supported)),
	)
	return supported != 0
}

// GetPhysicalDeviceSurfaceCapabilitiesKHR returns surface capabilities.
func (l *Loader) GetPhysicalDeviceSurfaceCapabilitiesKHR(device PhysicalDevice, surface SurfaceKHR) (SurfaceCapabilitiesKHR, error) {
	var caps SurfaceCapabilitiesKHR
	r, _, _ := syscallN(l.vkGetPhysicalDeviceSurfaceCapabilitiesKHR,
		uintptr(device),
		uintptr(surface),
		uintptr(unsafe.Pointer(&caps)),
	)
	if Result(r) != Success {
		return caps, fmt.Errorf("vulkan: vkGetPhysicalDeviceSurfaceCapabilitiesKHR failed: %v", Result(r))
	}
	return caps, nil
}

// GetPhysicalDeviceSurfaceFormatsKHR returns supported surface formats.
func (l *Loader) GetPhysicalDeviceSurfaceFormatsKHR(device PhysicalDevice, surface SurfaceKHR) ([]SurfaceFormatKHR, error) {
	var count uint32
	r, _, _ := syscallN(l.vkGetPhysicalDeviceSurfaceFormatsKHR,
		uintptr(device), uintptr(surface), uintptr(unsafe.Pointer(&count)), 0,
	)
	if Result(r) != Success {
		return nil, fmt.Errorf("vulkan: vkGetPhysicalDeviceSurfaceFormatsKHR count failed: %v", Result(r))
	}
	formats := make([]SurfaceFormatKHR, count)
	r, _, _ = syscallN(l.vkGetPhysicalDeviceSurfaceFormatsKHR,
		uintptr(device), uintptr(surface), uintptr(unsafe.Pointer(&count)), uintptr(unsafe.Pointer(&formats[0])),
	)
	if Result(r) != Success {
		return nil, fmt.Errorf("vulkan: vkGetPhysicalDeviceSurfaceFormatsKHR failed: %v", Result(r))
	}
	return formats[:count], nil
}

// GetPhysicalDeviceSurfacePresentModesKHR returns supported present modes.
func (l *Loader) GetPhysicalDeviceSurfacePresentModesKHR(device PhysicalDevice, surface SurfaceKHR) ([]PresentModeKHR, error) {
	var count uint32
	r, _, _ := syscallN(l.vkGetPhysicalDeviceSurfacePresentModesKHR,
		uintptr(device), uintptr(surface), uintptr(unsafe.Pointer(&count)), 0,
	)
	if Result(r) != Success {
		return nil, fmt.Errorf("vulkan: vkGetPhysicalDeviceSurfacePresentModesKHR count failed: %v", Result(r))
	}
	modes := make([]PresentModeKHR, count)
	r, _, _ = syscallN(l.vkGetPhysicalDeviceSurfacePresentModesKHR,
		uintptr(device), uintptr(surface), uintptr(unsafe.Pointer(&count)), uintptr(unsafe.Pointer(&modes[0])),
	)
	if Result(r) != Success {
		return nil, fmt.Errorf("vulkan: vkGetPhysicalDeviceSurfacePresentModesKHR failed: %v", Result(r))
	}
	return modes[:count], nil
}

// DestroySurfaceKHR destroys a surface.
func (l *Loader) DestroySurfaceKHR(instance Instance, surface SurfaceKHR) {
	if l.vkDestroySurfaceKHR != 0 {
		syscallN(l.vkDestroySurfaceKHR, uintptr(instance), uintptr(surface), 0)
	}
}

// cstr creates a null-terminated C string from a Go string.
func cstr(s string) []byte {
	b := make([]byte, len(s)+1)
	copy(b, s)
	return b
}

// cstrArray creates an array of C string pointers.
func cstrArray(strs []string) []*byte {
	ptrs := make([]*byte, len(strs))
	bufs := make([][]byte, len(strs)) // keep alive
	for i, s := range strs {
		bufs[i] = cstr(s)
		ptrs[i] = &bufs[i][0]
	}
	return ptrs
}

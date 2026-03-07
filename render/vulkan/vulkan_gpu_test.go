//go:build windows

package vulkan

import (
	"testing"
)

// requireVulkan tries to load the Vulkan library and create an instance.
// Returns loader+instance on success, or calls t.Skip if hardware is unavailable.
func requireVulkan(t *testing.T) (*Loader, Instance) {
	t.Helper()

	loader, err := NewLoader()
	if err != nil {
		t.Skipf("Vulkan not available: %v", err)
	}

	instance, err := loader.CreateInstance("GoUI-Test", "GoUI-Test", false)
	if err != nil {
		loader.Close()
		t.Skipf("Cannot create Vulkan instance: %v", err)
	}

	loader.LoadInstanceFunctions(instance)
	return loader, instance
}

// requirePhysicalDevice enumerates and picks a physical device.
func requirePhysicalDevice(t *testing.T, loader *Loader, instance Instance) PhysicalDevice {
	t.Helper()

	devices, err := loader.EnumeratePhysicalDevices(instance)
	if err != nil || len(devices) == 0 {
		t.Skipf("No physical devices: %v", err)
	}
	return devices[0]
}

func TestGPULoaderInit(t *testing.T) {
	loader, instance := requireVulkan(t)
	defer loader.DestroyInstance(instance)
	defer loader.Close()

	t.Log("Vulkan instance created successfully")
}

func TestGPUEnumeratePhysicalDevices(t *testing.T) {
	loader, instance := requireVulkan(t)
	defer loader.DestroyInstance(instance)
	defer loader.Close()

	devices, err := loader.EnumeratePhysicalDevices(instance)
	if err != nil {
		t.Fatalf("EnumeratePhysicalDevices failed: %v", err)
	}
	if len(devices) == 0 {
		t.Fatal("expected at least 1 physical device")
	}
	t.Logf("Found %d physical device(s)", len(devices))
}

func TestGPUQueueFamilyProperties(t *testing.T) {
	loader, instance := requireVulkan(t)
	defer loader.DestroyInstance(instance)
	defer loader.Close()

	device := requirePhysicalDevice(t, loader, instance)

	families := loader.GetPhysicalDeviceQueueFamilyProperties(device)
	if len(families) == 0 {
		t.Fatal("expected at least 1 queue family")
	}

	hasGraphics := false
	for i, fam := range families {
		if fam.QueueFlags&QueueGraphicsBit != 0 {
			hasGraphics = true
			t.Logf("Queue family %d: Graphics, %d queues", i, fam.QueueCount)
		}
	}
	if !hasGraphics {
		t.Error("expected at least one graphics queue family")
	}
}

func TestGPUMemoryProperties(t *testing.T) {
	loader, instance := requireVulkan(t)
	defer loader.DestroyInstance(instance)
	defer loader.Close()

	device := requirePhysicalDevice(t, loader, instance)

	memProps := loader.GetPhysicalDeviceMemoryProperties(device)
	if memProps.MemoryTypeCount == 0 {
		t.Error("expected at least 1 memory type")
	}
	if memProps.MemoryHeapCount == 0 {
		t.Error("expected at least 1 memory heap")
	}

	t.Logf("Memory types: %d, heaps: %d", memProps.MemoryTypeCount, memProps.MemoryHeapCount)

	// Verify we can find host-visible memory (needed for vertex buffers)
	_, ok := FindMemoryType(memProps, ^uint32(0), MemoryPropertyHostVisibleBit|MemoryPropertyHostCoherentBit)
	if !ok {
		t.Error("expected to find host-visible coherent memory type")
	}

	// Verify we can find device-local memory (needed for textures)
	_, ok = FindMemoryType(memProps, ^uint32(0), MemoryPropertyDeviceLocalBit)
	if !ok {
		t.Error("expected to find device-local memory type")
	}
}

func TestGPUCreateDeviceWithoutSurface(t *testing.T) {
	loader, instance := requireVulkan(t)
	defer loader.DestroyInstance(instance)
	defer loader.Close()

	device := requirePhysicalDevice(t, loader, instance)

	families := loader.GetPhysicalDeviceQueueFamilyProperties(device)
	graphicsIdx := -1
	for i, fam := range families {
		if fam.QueueFlags&QueueGraphicsBit != 0 {
			graphicsIdx = i
			break
		}
	}
	if graphicsIdx < 0 {
		t.Skip("No graphics queue family")
	}

	// Create device with graphics queue only (no present, no surface)
	indices := QueueFamilyIndices{
		Graphics: graphicsIdx,
		Present:  graphicsIdx, // Same family, no surface check needed
	}

	logicalDevice, err := loader.CreateDevice(device, indices)
	if err != nil {
		t.Fatalf("CreateDevice failed: %v", err)
	}
	defer loader.DestroyDevice(logicalDevice)

	loader.LoadDeviceFunctions(instance)

	t.Log("Logical device created successfully")

	// Verify we can get the queue
	queue := loader.GetDeviceQueue(logicalDevice, uint32(graphicsIdx), 0)
	if queue == 0 {
		t.Error("expected non-zero queue handle")
	}

	// Verify DeviceWaitIdle works
	loader.DeviceWaitIdle(logicalDevice)
}

func TestGPUCreateAndDestroyBuffer(t *testing.T) {
	loader, instance := requireVulkan(t)
	defer loader.DestroyInstance(instance)
	defer loader.Close()

	physDevice := requirePhysicalDevice(t, loader, instance)

	families := loader.GetPhysicalDeviceQueueFamilyProperties(physDevice)
	graphicsIdx := -1
	for i, fam := range families {
		if fam.QueueFlags&QueueGraphicsBit != 0 {
			graphicsIdx = i
			break
		}
	}
	if graphicsIdx < 0 {
		t.Skip("No graphics queue family")
	}

	indices := QueueFamilyIndices{Graphics: graphicsIdx, Present: graphicsIdx}
	device, err := loader.CreateDevice(physDevice, indices)
	if err != nil {
		t.Fatalf("CreateDevice failed: %v", err)
	}
	defer loader.DestroyDevice(device)

	loader.LoadDeviceFunctions(instance)

	// Use the Backend helper to create a buffer
	b := New()
	b.loader = loader
	b.device = device
	b.physicalDevice = physDevice
	b.memProps = loader.GetPhysicalDeviceMemoryProperties(physDevice)

	buffer, memory, err := b.createBuffer(
		1024,
		BufferUsageVertexBufferBit,
		MemoryPropertyHostVisibleBit|MemoryPropertyHostCoherentBit,
	)
	if err != nil {
		t.Fatalf("createBuffer failed: %v", err)
	}
	if buffer == 0 {
		t.Error("expected non-zero buffer handle")
	}
	if memory == 0 {
		t.Error("expected non-zero memory handle")
	}

	b.destroyBuffer(buffer, memory)
	t.Log("Buffer created and destroyed successfully")
}

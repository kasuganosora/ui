package vulkan

import (
	"testing"
	"unsafe"
)

// === Types & Constants Tests ===

func TestMakeAPIVersion(t *testing.T) {
	v := MakeAPIVersion(1, 2, 3)
	major := v >> 22
	minor := (v >> 12) & 0x3FF
	patch := v & 0xFFF
	if major != 1 || minor != 2 || patch != 3 {
		t.Errorf("expected 1.2.3, got %d.%d.%d", major, minor, patch)
	}
}

func TestMakeAPIVersionVulkan10(t *testing.T) {
	v := MakeAPIVersion(1, 0, 0)
	if v != (1 << 22) {
		t.Errorf("Vulkan 1.0.0 should be %d, got %d", 1<<22, v)
	}
}

func TestResultError(t *testing.T) {
	tests := []struct {
		r    Result
		want string
	}{
		{Success, "VK_SUCCESS"},
		{ErrorOutOfHostMemory, "VK_ERROR_OUT_OF_HOST_MEMORY"},
		{ErrorOutOfDeviceMemory, "VK_ERROR_OUT_OF_DEVICE_MEMORY"},
		{ErrorInitializationFailed, "VK_ERROR_INITIALIZATION_FAILED"},
		{ErrorDeviceLost, "VK_ERROR_DEVICE_LOST"},
		{ErrorSurfaceLostKHR, "VK_ERROR_SURFACE_LOST_KHR"},
		{ErrorOutOfDateKHR, "VK_ERROR_OUT_OF_DATE_KHR"},
		{Result(9999), "VK_ERROR_UNKNOWN"},
	}
	for _, tt := range tests {
		if got := tt.r.Error(); got != tt.want {
			t.Errorf("Result(%d).Error() = %q, want %q", tt.r, got, tt.want)
		}
	}
}

// === FindMemoryType Tests ===

func TestFindMemoryTypeSuccess(t *testing.T) {
	memProps := PhysicalDeviceMemoryProperties{
		MemoryTypeCount: 3,
	}
	memProps.MemoryTypes[0] = MemoryType{PropertyFlags: MemoryPropertyDeviceLocalBit, HeapIndex: 0}
	memProps.MemoryTypes[1] = MemoryType{PropertyFlags: MemoryPropertyHostVisibleBit | MemoryPropertyHostCoherentBit, HeapIndex: 1}
	memProps.MemoryTypes[2] = MemoryType{PropertyFlags: MemoryPropertyDeviceLocalBit | MemoryPropertyHostVisibleBit, HeapIndex: 0}

	// Find host-visible + coherent (type bit filter = all bits set)
	idx, ok := FindMemoryType(memProps, 0x7, MemoryPropertyHostVisibleBit|MemoryPropertyHostCoherentBit)
	if !ok {
		t.Fatal("expected to find memory type")
	}
	if idx != 1 {
		t.Errorf("expected index 1, got %d", idx)
	}
}

func TestFindMemoryTypeNoMatch(t *testing.T) {
	memProps := PhysicalDeviceMemoryProperties{
		MemoryTypeCount: 1,
	}
	memProps.MemoryTypes[0] = MemoryType{PropertyFlags: MemoryPropertyDeviceLocalBit}

	_, ok := FindMemoryType(memProps, 0x1, MemoryPropertyHostVisibleBit)
	if ok {
		t.Error("should not find host-visible in device-local only")
	}
}

func TestFindMemoryTypeFilterBits(t *testing.T) {
	memProps := PhysicalDeviceMemoryProperties{
		MemoryTypeCount: 4,
	}
	memProps.MemoryTypes[0] = MemoryType{PropertyFlags: MemoryPropertyHostVisibleBit}
	memProps.MemoryTypes[1] = MemoryType{PropertyFlags: MemoryPropertyHostVisibleBit}
	memProps.MemoryTypes[2] = MemoryType{PropertyFlags: MemoryPropertyHostVisibleBit}
	memProps.MemoryTypes[3] = MemoryType{PropertyFlags: MemoryPropertyHostVisibleBit}

	// typeFilter only allows bit 2 (index 2)
	idx, ok := FindMemoryType(memProps, 0x4, MemoryPropertyHostVisibleBit)
	if !ok {
		t.Fatal("expected match")
	}
	if idx != 2 {
		t.Errorf("expected index 2, got %d", idx)
	}
}

// === ChooseSurfaceFormat Tests ===

func TestChooseSurfaceFormatPrefersSrgb(t *testing.T) {
	formats := []SurfaceFormatKHR{
		{Format: FormatR8G8B8A8Unorm, ColorSpace: ColorSpaceSrgbNonlinearKHR},
		{Format: FormatB8G8R8A8Srgb, ColorSpace: ColorSpaceSrgbNonlinearKHR},
		{Format: FormatB8G8R8A8Unorm, ColorSpace: ColorSpaceSrgbNonlinearKHR},
	}
	chosen := ChooseSurfaceFormat(formats)
	if chosen.Format != FormatB8G8R8A8Srgb {
		t.Errorf("expected B8G8R8A8_SRGB, got %v", chosen.Format)
	}
}

func TestChooseSurfaceFormatFallbackUnorm(t *testing.T) {
	formats := []SurfaceFormatKHR{
		{Format: FormatR8G8B8A8Unorm, ColorSpace: ColorSpaceSrgbNonlinearKHR},
		{Format: FormatB8G8R8A8Unorm, ColorSpace: ColorSpaceSrgbNonlinearKHR},
	}
	chosen := ChooseSurfaceFormat(formats)
	if chosen.Format != FormatB8G8R8A8Unorm {
		t.Errorf("expected B8G8R8A8_UNORM fallback, got %v", chosen.Format)
	}
}

func TestChooseSurfaceFormatFallbackFirst(t *testing.T) {
	formats := []SurfaceFormatKHR{
		{Format: FormatR8G8B8A8Unorm, ColorSpace: ColorSpaceSrgbNonlinearKHR},
	}
	chosen := ChooseSurfaceFormat(formats)
	if chosen.Format != FormatR8G8B8A8Unorm {
		t.Errorf("expected first format as fallback, got %v", chosen.Format)
	}
}

// === ChoosePresentMode Tests ===

func TestChoosePresentModeVsync(t *testing.T) {
	modes := []PresentModeKHR{PresentModeImmediateKHR, PresentModeMailboxKHR, PresentModeFifoKHR}
	chosen := ChoosePresentMode(modes, true)
	if chosen != PresentModeFifoKHR {
		t.Errorf("vsync should choose FIFO, got %v", chosen)
	}
}

func TestChoosePresentModeNoVsyncPrefersMailbox(t *testing.T) {
	modes := []PresentModeKHR{PresentModeImmediateKHR, PresentModeMailboxKHR, PresentModeFifoKHR}
	chosen := ChoosePresentMode(modes, false)
	if chosen != PresentModeMailboxKHR {
		t.Errorf("no-vsync should prefer mailbox, got %v", chosen)
	}
}

func TestChoosePresentModeNoVsyncFallbackFIFO(t *testing.T) {
	modes := []PresentModeKHR{PresentModeImmediateKHR, PresentModeFifoKHR}
	chosen := ChoosePresentMode(modes, false)
	if chosen != PresentModeFifoKHR {
		t.Errorf("no mailbox available should fallback to FIFO, got %v", chosen)
	}
}

// === Backend Factory Tests ===

func TestNewBackend(t *testing.T) {
	b := New()
	if b == nil {
		t.Fatal("New() should not return nil")
	}
	if b.nextTextureID != 1 {
		t.Errorf("expected nextTextureID=1, got %d", b.nextTextureID)
	}
	if b.textures == nil {
		t.Error("textures map should be initialized")
	}
	if b.vertexSize != 64*1024 {
		t.Errorf("expected 64KB vertex buffer, got %d", b.vertexSize)
	}
}

// === Vertex Layout Tests ===

func TestRectVertexSize(t *testing.T) {
	// RectVertex has 19 float32 fields = 76 bytes
	expected := uintptr(76)
	actual := unsafe.Sizeof(RectVertex{})
	if actual != expected {
		t.Errorf("RectVertex size: expected %d bytes, got %d", expected, actual)
	}
}

func TestRectVertexBindingDescription(t *testing.T) {
	binding := rectVertexBindingDescription()
	if binding.Binding != 0 {
		t.Errorf("expected binding 0, got %d", binding.Binding)
	}
	if binding.Stride != 76 {
		t.Errorf("expected stride 76, got %d", binding.Stride)
	}
	if binding.InputRate != VertexInputRateVertex {
		t.Errorf("expected per-vertex input rate")
	}
}

func TestRectVertexAttributeDescriptions(t *testing.T) {
	attrs := rectVertexAttributeDescriptions()
	if len(attrs) != 7 {
		t.Fatalf("expected 7 attributes, got %d", len(attrs))
	}

	// Verify locations are sequential 0..6
	for i, attr := range attrs {
		if attr.Location != uint32(i) {
			t.Errorf("attr[%d].Location = %d, want %d", i, attr.Location, i)
		}
		if attr.Binding != 0 {
			t.Errorf("attr[%d].Binding = %d, want 0", i, attr.Binding)
		}
	}

	// Verify formats
	if attrs[0].Format != FormatR32G32Sfloat { // vec2 pos
		t.Error("attr 0 (pos) should be R32G32")
	}
	if attrs[1].Format != FormatR32G32Sfloat { // vec2 uv
		t.Error("attr 1 (uv) should be R32G32")
	}
	if attrs[2].Format != FormatR32G32B32A32Sfloat { // vec4 color
		t.Error("attr 2 (color) should be R32G32B32A32")
	}
	if attrs[3].Format != FormatR32G32Sfloat { // vec2 rectSize
		t.Error("attr 3 (rectSize) should be R32G32")
	}
	if attrs[4].Format != FormatR32G32B32A32Sfloat { // vec4 radii
		t.Error("attr 4 (radii) should be R32G32B32A32")
	}
	if attrs[5].Format != FormatR32Sfloat { // float borderWidth
		t.Error("attr 5 (borderWidth) should be R32")
	}
	if attrs[6].Format != FormatR32G32B32A32Sfloat { // vec4 borderColor
		t.Error("attr 6 (borderColor) should be R32G32B32A32")
	}

	// Verify offsets match struct layout
	expectedOffsets := []uint32{0, 8, 16, 32, 40, 56, 60}
	for i, attr := range attrs {
		if attr.Offset != expectedOffsets[i] {
			t.Errorf("attr[%d].Offset = %d, want %d", i, attr.Offset, expectedOffsets[i])
		}
	}
}

func TestRectVertexAttributesCoverFullStruct(t *testing.T) {
	attrs := rectVertexAttributeDescriptions()
	// Last attribute (borderColor vec4) starts at offset 60, size 16 = 76
	// But struct is 80 bytes (20 * 4). Wait - actually:
	// offset 60 + vec4(16 bytes) = 76, but struct has 20 floats = 80.
	// The last attr (borderColor) is at offset 60 with vec4 = 76.
	// But BorderA is at offset 76 (float, 4 bytes) -> 80 total.
	// Actually borderColor is vec4 covering BorderR(60), BorderG(64), BorderB(68), BorderA(72) = ends at 76.
	// Hmm, that's 19 floats covered (76 bytes). Let's check:
	// PosX(0), PosY(4) = 8
	// U(8), V(12) = 16
	// ColorR(16), ColorG(20), ColorB(24), ColorA(28) = 32
	// RectW(32), RectH(36) = 40
	// RadiusTL(40), RadiusTR(44), RadiusBR(48), RadiusBL(52) = 56
	// BorderWidth(56) = 60
	// BorderR(60), BorderG(64), BorderB(68), BorderA(72) = 76
	// Wait - that's only 19 floats * 4 = 76. But RectVertex has 20 fields = 80.
	// Let me recount fields in the struct...
	// Actually looking at the struct: it has exactly 20 float32 fields.
	// The offset math: borderColor starts at 60, is a vec4 (16 bytes), ends at 76.
	// 76 != 80, so 4 bytes (1 float) gap at end? No, struct packing...
	// Actually the struct has no padding because all fields are float32.
	// Let me count: PosX, PosY, U, V, ColorR, ColorG, ColorB, ColorA,
	//   RectW, RectH, RadiusTL, RadiusTR, RadiusBR, RadiusBL,
	//   BorderWidth, BorderR, BorderG, BorderB, BorderA = 19 floats!
	// Wait - that's 19. The struct says 20 float32 in the comment but let me verify.
	// Re-checking: the binding says stride=80 which is 20*4. Let me check sizeof.
	size := unsafe.Sizeof(RectVertex{})
	lastAttr := attrs[len(attrs)-1]
	lastEnd := lastAttr.Offset + 16 // vec4 = 16 bytes
	if uintptr(lastEnd) > size {
		t.Errorf("last attribute extends beyond struct: attr ends at %d, struct size %d", lastEnd, size)
	}
}

// === C String Helper Tests ===

func TestCstr(t *testing.T) {
	b := cstr("hello")
	if len(b) != 6 {
		t.Errorf("expected 6 bytes, got %d", len(b))
	}
	if b[5] != 0 {
		t.Error("cstr should be null-terminated")
	}
	if string(b[:5]) != "hello" {
		t.Errorf("expected 'hello', got '%s'", string(b[:5]))
	}
}

func TestCstrEmpty(t *testing.T) {
	b := cstr("")
	if len(b) != 1 || b[0] != 0 {
		t.Error("empty cstr should be a single null byte")
	}
}

func TestCstrArray(t *testing.T) {
	strs := []string{"VK_KHR_surface", "VK_KHR_win32_surface"}
	ptrs := cstrArray(strs)
	if len(ptrs) != 2 {
		t.Fatalf("expected 2 pointers, got %d", len(ptrs))
	}
	for i, p := range ptrs {
		if p == nil {
			t.Errorf("pointer %d is nil", i)
		}
	}
}

func TestCstrArrayEmpty(t *testing.T) {
	ptrs := cstrArray(nil)
	if len(ptrs) != 0 {
		t.Errorf("expected 0 pointers, got %d", len(ptrs))
	}
}

// === StructureType Constants Tests ===

func TestStructureTypeValues(t *testing.T) {
	// Verify critical sType values match Vulkan spec
	if StructureTypeApplicationInfo != 0 {
		t.Error("VK_STRUCTURE_TYPE_APPLICATION_INFO should be 0")
	}
	if StructureTypeInstanceCreateInfo != 1 {
		t.Error("VK_STRUCTURE_TYPE_INSTANCE_CREATE_INFO should be 1")
	}
	if StructureTypeSwapchainCreateInfoKHR != 1000001000 {
		t.Error("VK_STRUCTURE_TYPE_SWAPCHAIN_CREATE_INFO_KHR should be 1000001000")
	}
	if StructureTypeWin32SurfaceCreateInfoKHR != 1000009000 {
		t.Error("VK_STRUCTURE_TYPE_WIN32_SURFACE_CREATE_INFO_KHR should be 1000009000")
	}
}

// === Shader Module Tests (no GPU) ===

func TestMainEntryPoint(t *testing.T) {
	if len(mainEntryPoint) != 5 {
		t.Errorf("expected 5 bytes ('main' + null), got %d", len(mainEntryPoint))
	}
	if string(mainEntryPoint[:4]) != "main" {
		t.Error("entry point should be 'main'")
	}
	if mainEntryPoint[4] != 0 {
		t.Error("entry point should be null-terminated")
	}
}

func TestShaderBytecodeEmbedded(t *testing.T) {
	if len(rectVertSPV) == 0 {
		t.Fatal("rect vertex shader SPIR-V not embedded")
	}
	if len(rectFragSPV) == 0 {
		t.Fatal("rect fragment shader SPIR-V not embedded")
	}
	// SPIR-V magic number: 0x07230203
	if rectVertSPV[0] != 0x03 || rectVertSPV[1] != 0x02 || rectVertSPV[2] != 0x23 || rectVertSPV[3] != 0x07 {
		t.Error("rect vertex shader does not have SPIR-V magic number")
	}
	if rectFragSPV[0] != 0x03 || rectFragSPV[1] != 0x02 || rectFragSPV[2] != 0x23 || rectFragSPV[3] != 0x07 {
		t.Error("rect fragment shader does not have SPIR-V magic number")
	}
	t.Logf("rect.vert.spv: %d bytes, rect.frag.spv: %d bytes", len(rectVertSPV), len(rectFragSPV))
}

// === Backend Compile-time Interface Check ===

func TestBackendImplementsRenderBackend(t *testing.T) {
	// This is already checked at compile time by the var _ line in backend.go,
	// but we include it here for test coverage documentation.
	var _ interface {
		BeginFrame()
		EndFrame()
		Resize(width, height int)
		Destroy()
		MaxTextureSize() int
	} = (*Backend)(nil)
}

func TestMaxTextureSize(t *testing.T) {
	b := New()
	if b.MaxTextureSize() != 4096 {
		t.Errorf("expected 4096, got %d", b.MaxTextureSize())
	}
}

// === Handle Constants Tests ===

func TestNullHandle(t *testing.T) {
	if NullHandle != 0 {
		t.Error("NullHandle should be 0")
	}
}

func TestWholeSize(t *testing.T) {
	if WholeSize != ^uint64(0) {
		t.Error("WholeSize should be max uint64")
	}
}

func TestMaxTimeout(t *testing.T) {
	if MaxTimeout != ^uint64(0) {
		t.Error("MaxTimeout should be max uint64")
	}
}

// === Textured Shader Embeds ===

func TestTexturedShaderBytecodeEmbedded(t *testing.T) {
	shaders := map[string][]byte{
		"textured.vert.spv": texturedVertSPV,
		"textured.frag.spv": texturedFragSPV,
		"text.frag.spv":     textFragSPV,
	}
	for name, spv := range shaders {
		if len(spv) == 0 {
			t.Fatalf("%s not embedded", name)
		}
		if spv[0] != 0x03 || spv[1] != 0x02 || spv[2] != 0x23 || spv[3] != 0x07 {
			t.Errorf("%s does not have SPIR-V magic number", name)
		}
		t.Logf("%s: %d bytes", name, len(spv))
	}
}

// === Textured Vertex Layout ===

func TestTexturedVertexSize(t *testing.T) {
	size := unsafe.Sizeof(TexturedVertex{})
	if size != 32 {
		t.Errorf("TexturedVertex should be 32 bytes, got %d", size)
	}
}

func TestTexturedVertexAttributeDescriptions(t *testing.T) {
	attrs := texturedVertexAttributeDescriptions()
	if len(attrs) != 3 {
		t.Fatalf("expected 3 attributes, got %d", len(attrs))
	}
	// location 0: vec2 pos at offset 0
	if attrs[0].Location != 0 || attrs[0].Offset != 0 || attrs[0].Format != FormatR32G32Sfloat {
		t.Error("attribute 0 (pos) incorrect")
	}
	// location 1: vec2 uv at offset 8
	if attrs[1].Location != 1 || attrs[1].Offset != 8 || attrs[1].Format != FormatR32G32Sfloat {
		t.Error("attribute 1 (uv) incorrect")
	}
	// location 2: vec4 color at offset 16
	if attrs[2].Location != 2 || attrs[2].Offset != 16 || attrs[2].Format != FormatR32G32B32A32Sfloat {
		t.Error("attribute 2 (color) incorrect")
	}
}

func TestTexturedVertexBindingDescription(t *testing.T) {
	binding := texturedVertexBindingDescription()
	if binding.Stride != 32 {
		t.Errorf("expected stride 32, got %d", binding.Stride)
	}
	if binding.InputRate != VertexInputRateVertex {
		t.Error("expected per-vertex input rate")
	}
}

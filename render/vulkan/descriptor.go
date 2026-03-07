package vulkan

import (
	"fmt"
	"unsafe"
)

// descriptorSetLayoutBinding for VkDescriptorSetLayoutBinding.
type descriptorSetLayoutBinding struct {
	Binding            uint32
	DescriptorType     DescriptorType
	DescriptorCount    uint32
	StageFlags         ShaderStageFlags
	PImmutableSamplers *Sampler
}

// descriptorSetLayoutCreateInfo for VkDescriptorSetLayoutCreateInfo.
type descriptorSetLayoutCreateInfo struct {
	SType        StructureType
	PNext        unsafe.Pointer
	Flags        uint32
	BindingCount uint32
	PBindings    *descriptorSetLayoutBinding
}

// descriptorPoolSize for VkDescriptorPoolSize.
type descriptorPoolSize struct {
	Type            DescriptorType
	DescriptorCount uint32
}

// descriptorPoolCreateInfo for VkDescriptorPoolCreateInfo.
type descriptorPoolCreateInfo struct {
	SType         StructureType
	PNext         unsafe.Pointer
	Flags         uint32
	MaxSets       uint32
	PoolSizeCount uint32
	PPoolSizes    *descriptorPoolSize
}

// descriptorSetAllocateInfo for VkDescriptorSetAllocateInfo.
type descriptorSetAllocateInfo struct {
	SType              StructureType
	PNext              unsafe.Pointer
	DescriptorPool     DescriptorPool
	DescriptorSetCount uint32
	PSetLayouts        *DescriptorSetLayout
}

// writeDescriptorSet for VkWriteDescriptorSet.
type writeDescriptorSet struct {
	SType            StructureType
	PNext            unsafe.Pointer
	DstSet           DescriptorSet
	DstBinding       uint32
	DstArrayElement  uint32
	DescriptorCount  uint32
	DescriptorType   DescriptorType
	PImageInfo       *descriptorImageInfo
	PBufferInfo      unsafe.Pointer
	PTexelBufferView unsafe.Pointer
}

// descriptorImageInfo for VkDescriptorImageInfo.
type descriptorImageInfo struct {
	Sampler     Sampler
	ImageView   ImageView
	ImageLayout ImageLayout
}

const maxTextureDescriptorSets = 256

// Descriptor pool create flags
const descriptorPoolCreateFreeDescriptorSetBit uint32 = 0x00000001

// createDescriptorInfrastructure creates the shared descriptor set layout and pool
// for texture binding (binding 0 = combined image sampler, fragment stage).
func (b *Backend) createDescriptorInfrastructure() error {
	// Descriptor set layout: one combined image sampler at binding 0
	binding := descriptorSetLayoutBinding{
		Binding:         0,
		DescriptorType:  DescriptorTypeCombinedImageSampler,
		DescriptorCount: 1,
		StageFlags:      ShaderStageFragmentBit,
	}

	layoutCI := descriptorSetLayoutCreateInfo{
		SType:        StructureTypeDescriptorSetLayoutCreateInfo,
		BindingCount: 1,
		PBindings:    &binding,
	}

	r, _, _ := syscallN(b.loader.vkCreateDescriptorSetLayout,
		uintptr(b.device), uintptr(unsafe.Pointer(&layoutCI)), 0,
		uintptr(unsafe.Pointer(&b.texDescSetLayout)),
	)
	if Result(r) != Success {
		return fmt.Errorf("vulkan: vkCreateDescriptorSetLayout failed: %v", Result(r))
	}

	// Descriptor pool
	poolSize := descriptorPoolSize{
		Type:            DescriptorTypeCombinedImageSampler,
		DescriptorCount: maxTextureDescriptorSets,
	}

	poolCI := descriptorPoolCreateInfo{
		SType:         StructureTypeDescriptorPoolCreateInfo,
		Flags:         descriptorPoolCreateFreeDescriptorSetBit,
		MaxSets:       maxTextureDescriptorSets,
		PoolSizeCount: 1,
		PPoolSizes:    &poolSize,
	}

	r, _, _ = syscallN(b.loader.vkCreateDescriptorPool,
		uintptr(b.device), uintptr(unsafe.Pointer(&poolCI)), 0,
		uintptr(unsafe.Pointer(&b.texDescPool)),
	)
	if Result(r) != Success {
		return fmt.Errorf("vulkan: vkCreateDescriptorPool failed: %v", Result(r))
	}

	return nil
}

// allocateTextureDescriptorSet allocates and writes a descriptor set for a texture entry.
func (b *Backend) allocateTextureDescriptorSet(entry *textureEntry) error {
	layout := b.texDescSetLayout
	allocInfo := descriptorSetAllocateInfo{
		SType:              StructureTypeDescriptorSetAllocateInfo,
		DescriptorPool:     b.texDescPool,
		DescriptorSetCount: 1,
		PSetLayouts:        &layout,
	}

	r, _, _ := syscallN(b.loader.vkAllocateDescriptorSets,
		uintptr(b.device), uintptr(unsafe.Pointer(&allocInfo)),
		uintptr(unsafe.Pointer(&entry.descriptorSet)),
	)
	if Result(r) != Success {
		return fmt.Errorf("vulkan: vkAllocateDescriptorSets failed: %v", Result(r))
	}

	// Write the combined image sampler descriptor
	imageInfo := descriptorImageInfo{
		Sampler:     entry.sampler,
		ImageView:   entry.view,
		ImageLayout: ImageLayoutShaderReadOnlyOptimal,
	}

	write := writeDescriptorSet{
		SType:           StructureTypeWriteDescriptorSet,
		DstSet:          entry.descriptorSet,
		DstBinding:      0,
		DescriptorCount: 1,
		DescriptorType:  DescriptorTypeCombinedImageSampler,
		PImageInfo:      &imageInfo,
	}

	syscallN(b.loader.vkUpdateDescriptorSets,
		uintptr(b.device), 1, uintptr(unsafe.Pointer(&write)), 0, 0,
	)

	return nil
}

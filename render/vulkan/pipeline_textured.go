package vulkan

import (
	"fmt"
	"unsafe"
)

// TexturedVertex for textured quad rendering — 8 float32 fields, 32 bytes.
type TexturedVertex struct {
	PosX, PosY     float32 // NDC position
	U, V           float32 // Texture UV
	ColorR, ColorG float32 // Tint color
	ColorB, ColorA float32
}

// texturedVertexBindingDescription returns the vertex binding for TexturedVertex.
func texturedVertexBindingDescription() vertexInputBindingDescription {
	return vertexInputBindingDescription{
		Binding:   0,
		Stride:    uint32(unsafe.Sizeof(TexturedVertex{})),
		InputRate: VertexInputRateVertex,
	}
}

// texturedVertexAttributeDescriptions returns the per-attribute layout for TexturedVertex.
func texturedVertexAttributeDescriptions() [3]vertexInputAttributeDescription {
	return [3]vertexInputAttributeDescription{
		// location 0: vec2 inPos
		{Location: 0, Binding: 0, Format: FormatR32G32Sfloat, Offset: 0},
		// location 1: vec2 inUV
		{Location: 1, Binding: 0, Format: FormatR32G32Sfloat, Offset: 8},
		// location 2: vec4 inColor
		{Location: 2, Binding: 0, Format: FormatR32G32B32A32Sfloat, Offset: 16},
	}
}

// createTexturedPipelineWithShaders creates a pipeline using the textured vertex shader
// and the specified fragment shader, with a descriptor set layout for texture binding.
func (b *Backend) createTexturedPipelineWithShaders(fragSPV []byte) (Pipeline, PipelineLayout, error) {
	vertModule, err := b.createShaderModule(texturedVertSPV)
	if err != nil {
		return 0, 0, fmt.Errorf("vulkan: textured vertex shader: %w", err)
	}
	defer syscallN(b.loader.vkDestroyShaderModule, uintptr(b.device), uintptr(vertModule), 0)

	fragModule, err := b.createShaderModule(fragSPV)
	if err != nil {
		return 0, 0, fmt.Errorf("vulkan: textured fragment shader: %w", err)
	}
	defer syscallN(b.loader.vkDestroyShaderModule, uintptr(b.device), uintptr(fragModule), 0)

	shaderStages := [2]pipelineShaderStageCreateInfo{
		{
			SType:  StructureTypePipelineShaderStageCreateInfo,
			Stage:  ShaderStageVertexBit,
			Module: vertModule,
			PName:  &mainEntryPoint[0],
		},
		{
			SType:  StructureTypePipelineShaderStageCreateInfo,
			Stage:  ShaderStageFragmentBit,
			Module: fragModule,
			PName:  &mainEntryPoint[0],
		},
	}

	binding := texturedVertexBindingDescription()
	attributes := texturedVertexAttributeDescriptions()
	vertexInput := pipelineVertexInputStateCreateInfo{
		SType:                           StructureTypePipelineVertexInputStateCreateInfo,
		VertexBindingDescriptionCount:   1,
		PVertexBindingDescriptions:      &binding,
		VertexAttributeDescriptionCount: uint32(len(attributes)),
		PVertexAttributeDescriptions:    &attributes[0],
	}

	inputAssembly := pipelineInputAssemblyStateCreateInfo{
		SType:    StructureTypePipelineInputAssemblyStateCreateInfo,
		Topology: PrimitiveTopologyTriangleList,
	}

	viewportState := pipelineViewportStateCreateInfo{
		SType:         StructureTypePipelineViewportStateCreateInfo,
		ViewportCount: 1,
		ScissorCount:  1,
	}

	rasterizer := pipelineRasterizationStateCreateInfo{
		SType:       StructureTypePipelineRasterizationStateCreateInfo,
		PolygonMode: PolygonModeFill,
		CullMode:    CullModeNone,
		FrontFace:   FrontFaceClockwise,
		LineWidth:   1.0,
	}

	multisample := pipelineMultisampleStateCreateInfo{
		SType:                StructureTypePipelineMultisampleStateCreateInfo,
		RasterizationSamples: SampleCount1Bit,
	}

	colorAttachment := pipelineColorBlendAttachmentState{
		BlendEnable:         1,
		SrcColorBlendFactor: BlendFactorSrcAlpha,
		DstColorBlendFactor: BlendFactorOneMinusSrcAlpha,
		ColorBlendOp:        BlendOpAdd,
		SrcAlphaBlendFactor: BlendFactorOne,
		DstAlphaBlendFactor: BlendFactorOneMinusSrcAlpha,
		AlphaBlendOp:        BlendOpAdd,
		ColorWriteMask:      ColorComponentAll,
	}

	colorBlend := pipelineColorBlendStateCreateInfo{
		SType:           StructureTypePipelineColorBlendStateCreateInfo,
		AttachmentCount: 1,
		PAttachments:    &colorAttachment,
	}

	dynamicStates := [2]DynamicState{DynamicStateViewport, DynamicStateScissor}
	dynamicState := pipelineDynamicStateCreateInfo{
		SType:             StructureTypePipelineDynamicStateCreateInfo,
		DynamicStateCount: 2,
		PDynamicStates:    &dynamicStates[0],
	}

	// Pipeline layout with one descriptor set (texture sampler)
	setLayout := b.texDescSetLayout
	layoutInfo := pipelineLayoutCreateInfo{
		SType:          StructureTypePipelineLayoutCreateInfo,
		SetLayoutCount: 1,
		PSetLayouts:    &setLayout,
	}

	var pipelineLayout PipelineLayout
	r, _, _ := syscallN(b.loader.vkCreatePipelineLayout,
		uintptr(b.device), uintptr(unsafe.Pointer(&layoutInfo)), 0,
		uintptr(unsafe.Pointer(&pipelineLayout)),
	)
	if Result(r) != Success {
		return 0, 0, fmt.Errorf("vulkan: vkCreatePipelineLayout for textured pipeline failed: %v", Result(r))
	}

	pipelineInfo := graphicsPipelineCreateInfo{
		SType:               StructureTypeGraphicsPipelineCreateInfo,
		StageCount:          2,
		PStages:             &shaderStages[0],
		PVertexInputState:   &vertexInput,
		PInputAssemblyState: &inputAssembly,
		PViewportState:      &viewportState,
		PRasterizationState: &rasterizer,
		PMultisampleState:   &multisample,
		PColorBlendState:    &colorBlend,
		PDynamicState:       &dynamicState,
		Layout:              pipelineLayout,
		RenderPass:          b.renderPass,
		Subpass:             0,
		BasePipelineIndex:   -1,
	}

	var pipeline Pipeline
	r, _, _ = syscallN(b.loader.vkCreateGraphicsPipelines,
		uintptr(b.device), 0, 1, uintptr(unsafe.Pointer(&pipelineInfo)), 0,
		uintptr(unsafe.Pointer(&pipeline)),
	)
	if Result(r) != Success {
		syscallN(b.loader.vkDestroyPipelineLayout, uintptr(b.device), uintptr(pipelineLayout), 0)
		return 0, 0, fmt.Errorf("vulkan: vkCreateGraphicsPipelines for textured pipeline failed: %v", Result(r))
	}

	return pipeline, pipelineLayout, nil
}

// createTexturedPipeline creates the pipeline for image rendering (textured.frag).
func (b *Backend) createTexturedPipeline() error {
	var err error
	b.texturedPipeline, b.texturedPipelineLayout, err = b.createTexturedPipelineWithShaders(texturedFragSPV)
	if err != nil {
		return fmt.Errorf("vulkan: textured pipeline: %w", err)
	}
	return nil
}

// createTextPipeline creates the pipeline for SDF text rendering (text.frag).
func (b *Backend) createTextPipeline() error {
	var err error
	b.textPipeline, b.textPipelineLayout, err = b.createTexturedPipelineWithShaders(textFragSPV)
	if err != nil {
		return fmt.Errorf("vulkan: text pipeline: %w", err)
	}
	return nil
}

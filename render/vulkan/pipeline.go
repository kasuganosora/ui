package vulkan

import (
	"fmt"
	"unsafe"
)

// Vertex input binding description
type vertexInputBindingDescription struct {
	Binding   uint32
	Stride    uint32
	InputRate VertexInputRate
}

// Vertex input attribute description
type vertexInputAttributeDescription struct {
	Location uint32
	Binding  uint32
	Format   Format
	Offset   uint32
}

// Pipeline shader stage create info
type pipelineShaderStageCreateInfo struct {
	SType               StructureType
	PNext               unsafe.Pointer
	Flags               uint32
	Stage               ShaderStageFlags
	Module              ShaderModule
	PName               *byte
	PSpecializationInfo unsafe.Pointer
}

// Pipeline vertex input state
type pipelineVertexInputStateCreateInfo struct {
	SType                           StructureType
	PNext                           unsafe.Pointer
	Flags                           uint32
	VertexBindingDescriptionCount   uint32
	PVertexBindingDescriptions      *vertexInputBindingDescription
	VertexAttributeDescriptionCount uint32
	PVertexAttributeDescriptions    *vertexInputAttributeDescription
}

// Pipeline input assembly state
type pipelineInputAssemblyStateCreateInfo struct {
	SType                  StructureType
	PNext                  unsafe.Pointer
	Flags                  uint32
	Topology               PrimitiveTopology
	PrimitiveRestartEnable uint32
}

// Pipeline viewport state
type pipelineViewportStateCreateInfo struct {
	SType         StructureType
	PNext         unsafe.Pointer
	Flags         uint32
	ViewportCount uint32
	PViewports    *Viewport
	ScissorCount  uint32
	PScissors     *Rect2D
}

// Pipeline rasterization state
type pipelineRasterizationStateCreateInfo struct {
	SType                   StructureType
	PNext                   unsafe.Pointer
	Flags                   uint32
	DepthClampEnable        uint32
	RasterizerDiscardEnable uint32
	PolygonMode             PolygonMode
	CullMode                CullModeFlags
	FrontFace               FrontFace
	DepthBiasEnable         uint32
	DepthBiasConstantFactor float32
	DepthBiasClamp          float32
	DepthBiasSlopeFactor    float32
	LineWidth               float32
}

// Pipeline multisample state
type pipelineMultisampleStateCreateInfo struct {
	SType                 StructureType
	PNext                 unsafe.Pointer
	Flags                 uint32
	RasterizationSamples  SampleCountFlagBits
	SampleShadingEnable   uint32
	MinSampleShading      float32
	PSampleMask           *uint32
	AlphaToCoverageEnable uint32
	AlphaToOneEnable      uint32
}

// Pipeline color blend attachment state
type pipelineColorBlendAttachmentState struct {
	BlendEnable         uint32
	SrcColorBlendFactor BlendFactor
	DstColorBlendFactor BlendFactor
	ColorBlendOp        BlendOp
	SrcAlphaBlendFactor BlendFactor
	DstAlphaBlendFactor BlendFactor
	AlphaBlendOp        BlendOp
	ColorWriteMask      ColorComponentFlags
}

// Pipeline color blend state
type pipelineColorBlendStateCreateInfo struct {
	SType           StructureType
	PNext           unsafe.Pointer
	Flags           uint32
	LogicOpEnable   uint32
	LogicOp         int32
	AttachmentCount uint32
	PAttachments    *pipelineColorBlendAttachmentState
	BlendConstants  [4]float32
}

// Pipeline dynamic state
type pipelineDynamicStateCreateInfo struct {
	SType             StructureType
	PNext             unsafe.Pointer
	Flags             uint32
	DynamicStateCount uint32
	PDynamicStates    *DynamicState
}

// Pipeline layout create info
type pipelineLayoutCreateInfo struct {
	SType                  StructureType
	PNext                  unsafe.Pointer
	Flags                  uint32
	SetLayoutCount         uint32
	PSetLayouts            *DescriptorSetLayout
	PushConstantRangeCount uint32
	PPushConstantRanges    unsafe.Pointer
}

// Graphics pipeline create info
type graphicsPipelineCreateInfo struct {
	SType               StructureType
	PNext               unsafe.Pointer
	Flags               uint32
	StageCount          uint32
	PStages             *pipelineShaderStageCreateInfo
	PVertexInputState   *pipelineVertexInputStateCreateInfo
	PInputAssemblyState *pipelineInputAssemblyStateCreateInfo
	PTessellationState  unsafe.Pointer
	PViewportState      *pipelineViewportStateCreateInfo
	PRasterizationState *pipelineRasterizationStateCreateInfo
	PMultisampleState   *pipelineMultisampleStateCreateInfo
	PDepthStencilState  unsafe.Pointer
	PColorBlendState    *pipelineColorBlendStateCreateInfo
	PDynamicState       *pipelineDynamicStateCreateInfo
	Layout              PipelineLayout
	RenderPass          RenderPass
	Subpass             uint32
	BasePipelineHandle  Pipeline
	BasePipelineIndex   int32
}

// Shader module create info
type shaderModuleCreateInfo struct {
	SType    StructureType
	PNext    unsafe.Pointer
	Flags    uint32
	CodeSize uintptr
	PCode    *uint32
}

// createShaderModule creates a Vulkan shader module from SPIR-V bytecode.
func (b *Backend) createShaderModule(code []byte) (ShaderModule, error) {
	if len(code) == 0 {
		return 0, fmt.Errorf("vulkan: empty shader bytecode")
	}

	createInfo := shaderModuleCreateInfo{
		SType:    StructureTypeShaderModuleCreateInfo,
		CodeSize: uintptr(len(code)),
		PCode:    (*uint32)(unsafe.Pointer(&code[0])),
	}

	var module ShaderModule
	r, _, _ := syscallN(b.loader.vkCreateShaderModule,
		uintptr(b.device),
		uintptr(unsafe.Pointer(&createInfo)),
		0,
		uintptr(unsafe.Pointer(&module)),
	)
	if Result(r) != Success {
		return 0, fmt.Errorf("vulkan: vkCreateShaderModule failed: %v", Result(r))
	}
	return module, nil
}

// rectVertexBindingDescription returns the vertex binding for RectVertex.
func rectVertexBindingDescription() vertexInputBindingDescription {
	return vertexInputBindingDescription{
		Binding:   0,
		Stride:    uint32(unsafe.Sizeof(RectVertex{})), // 76 bytes (19 float32)
		InputRate: VertexInputRateVertex,
	}
}

// FormatR32G32Sfloat = vec2
const FormatR32G32Sfloat Format = 103

// FormatR32G32B32A32Sfloat = vec4
const FormatR32G32B32A32Sfloat Format = 109

// FormatR32Sfloat = float
const FormatR32Sfloat Format = 100

// rectVertexAttributeDescriptions returns the per-attribute layout for RectVertex.
func rectVertexAttributeDescriptions() [7]vertexInputAttributeDescription {
	return [7]vertexInputAttributeDescription{
		// location 0: vec2 pos (PosX, PosY)
		{Location: 0, Binding: 0, Format: FormatR32G32Sfloat, Offset: 0},
		// location 1: vec2 uv (U, V)
		{Location: 1, Binding: 0, Format: FormatR32G32Sfloat, Offset: 8},
		// location 2: vec4 color (ColorR, ColorG, ColorB, ColorA)
		{Location: 2, Binding: 0, Format: FormatR32G32B32A32Sfloat, Offset: 16},
		// location 3: vec2 rectSize (RectW, RectH)
		{Location: 3, Binding: 0, Format: FormatR32G32Sfloat, Offset: 32},
		// location 4: vec4 radii (RadiusTL, RadiusTR, RadiusBR, RadiusBL)
		{Location: 4, Binding: 0, Format: FormatR32G32B32A32Sfloat, Offset: 40},
		// location 5: float borderWidth
		{Location: 5, Binding: 0, Format: FormatR32Sfloat, Offset: 56},
		// location 6: vec4 borderColor (BorderR, BorderG, BorderB, BorderA)
		{Location: 6, Binding: 0, Format: FormatR32G32B32A32Sfloat, Offset: 60},
	}
}

var mainEntryPoint = append([]byte("main"), 0)

// createRectPipeline creates the graphics pipeline for rendering rounded rectangles.
func (b *Backend) createRectPipeline() error {
	// Create shader modules
	vertModule, err := b.createShaderModule(rectVertSPV)
	if err != nil {
		return fmt.Errorf("vulkan: rect vertex shader: %w", err)
	}
	defer syscallN(b.loader.vkDestroyShaderModule, uintptr(b.device), uintptr(vertModule), 0)

	fragModule, err := b.createShaderModule(rectFragSPV)
	if err != nil {
		return fmt.Errorf("vulkan: rect fragment shader: %w", err)
	}
	defer syscallN(b.loader.vkDestroyShaderModule, uintptr(b.device), uintptr(fragModule), 0)

	// Shader stages
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

	// Vertex input
	binding := rectVertexBindingDescription()
	attributes := rectVertexAttributeDescriptions()
	vertexInput := pipelineVertexInputStateCreateInfo{
		SType:                           StructureTypePipelineVertexInputStateCreateInfo,
		VertexBindingDescriptionCount:   1,
		PVertexBindingDescriptions:      &binding,
		VertexAttributeDescriptionCount: uint32(len(attributes)),
		PVertexAttributeDescriptions:    &attributes[0],
	}

	// Input assembly
	inputAssembly := pipelineInputAssemblyStateCreateInfo{
		SType:    StructureTypePipelineInputAssemblyStateCreateInfo,
		Topology: PrimitiveTopologyTriangleList,
	}

	// Viewport state (dynamic, so counts only)
	viewportState := pipelineViewportStateCreateInfo{
		SType:         StructureTypePipelineViewportStateCreateInfo,
		ViewportCount: 1,
		ScissorCount:  1,
	}

	// Rasterization - no culling for UI quads
	rasterizer := pipelineRasterizationStateCreateInfo{
		SType:       StructureTypePipelineRasterizationStateCreateInfo,
		PolygonMode: PolygonModeFill,
		CullMode:    CullModeNone,
		FrontFace:   FrontFaceClockwise,
		LineWidth:   1.0,
	}

	// Multisample - no MSAA
	multisample := pipelineMultisampleStateCreateInfo{
		SType:                StructureTypePipelineMultisampleStateCreateInfo,
		RasterizationSamples: SampleCount1Bit,
	}

	// Color blend - standard alpha blending for UI
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

	// Dynamic state
	dynamicStates := [2]DynamicState{DynamicStateViewport, DynamicStateScissor}
	dynamicState := pipelineDynamicStateCreateInfo{
		SType:             StructureTypePipelineDynamicStateCreateInfo,
		DynamicStateCount: 2,
		PDynamicStates:    &dynamicStates[0],
	}

	// Pipeline layout (no descriptors or push constants for rect pipeline)
	layoutInfo := pipelineLayoutCreateInfo{
		SType: StructureTypePipelineLayoutCreateInfo,
	}

	r, _, _ := syscallN(b.loader.vkCreatePipelineLayout,
		uintptr(b.device),
		uintptr(unsafe.Pointer(&layoutInfo)),
		0,
		uintptr(unsafe.Pointer(&b.rectPipelineLayout)),
	)
	if Result(r) != Success {
		return fmt.Errorf("vulkan: vkCreatePipelineLayout failed: %v", Result(r))
	}

	// Graphics pipeline
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
		Layout:              b.rectPipelineLayout,
		RenderPass:          b.renderPass,
		Subpass:             0,
		BasePipelineIndex:   -1,
	}

	r, _, _ = syscallN(b.loader.vkCreateGraphicsPipelines,
		uintptr(b.device),
		0, // pipeline cache
		1,
		uintptr(unsafe.Pointer(&pipelineInfo)),
		0,
		uintptr(unsafe.Pointer(&b.rectPipeline)),
	)
	if Result(r) != Success {
		return fmt.Errorf("vulkan: vkCreateGraphicsPipelines failed: %v", Result(r))
	}

	return nil
}

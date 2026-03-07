package vulkan

import _ "embed"

// To regenerate SPIR-V bytecode:
//   1. Install the Vulkan SDK (provides glslc)
//   2. Run:
//        cd render/vulkan/shaders
//        glslc -O --target-env=vulkan1.0 -fshader-stage=vertex rect.vert.glsl -o rect.vert.spv
//        glslc -O --target-env=vulkan1.0 -fshader-stage=fragment rect.frag.glsl -o rect.frag.spv
//        glslc -O --target-env=vulkan1.0 -fshader-stage=vertex textured.vert.glsl -o textured.vert.spv
//        glslc -O --target-env=vulkan1.0 -fshader-stage=fragment textured.frag.glsl -o textured.frag.spv
//        glslc -O --target-env=vulkan1.0 -fshader-stage=fragment text.frag.glsl -o text.frag.spv

// rectVertSPV contains the compiled SPIR-V for the rect vertex shader.
//
//go:embed shaders/rect.vert.spv
var rectVertSPV []byte

// rectFragSPV contains the compiled SPIR-V for the rect fragment shader.
//
//go:embed shaders/rect.frag.spv
var rectFragSPV []byte

// texturedVertSPV contains the compiled SPIR-V for the textured vertex shader.
//
//go:embed shaders/textured.vert.spv
var texturedVertSPV []byte

// texturedFragSPV contains the compiled SPIR-V for the textured fragment shader.
//
//go:embed shaders/textured.frag.spv
var texturedFragSPV []byte

// textFragSPV contains the compiled SPIR-V for the SDF text fragment shader.
//
//go:embed shaders/text.frag.spv
var textFragSPV []byte

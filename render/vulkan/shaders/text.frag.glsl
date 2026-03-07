#version 450

layout(location = 0) in vec2 fragUV;
layout(location = 1) in vec4 fragColor;

layout(binding = 0) uniform sampler2D glyphAtlas;

layout(location = 0) out vec4 outColor;

void main() {
    // Glyph atlas: single-channel coverage (grayscale or SDF)
    float coverage = texture(glyphAtlas, fragUV).r;

    // Use coverage directly as alpha (works for both grayscale and SDF bitmaps).
    // Grayscale: 0 = transparent, 1 = fully covered.
    // SDF: values > 0.5 are inside the glyph. For SDF, a smoothstep would be
    // better, but since FreeType SDF is not always available we use direct alpha
    // which still looks good at typical UI font sizes (12-24px).
    outColor = vec4(fragColor.rgb, fragColor.a * coverage);
}

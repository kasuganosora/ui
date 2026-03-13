#version 450

layout(location = 0) in vec2 inPos;      // NDC position
layout(location = 1) in vec2 inUV;       // texture UV
layout(location = 2) in vec4 inColor;    // tint color
layout(location = 3) in vec2 inRectSize; // rect size in physical px (0 = no SDF)
layout(location = 4) in vec4 inRadius;   // corner radii (TL, TR, BR, BL)

layout(location = 0) out vec2 fragUV;
layout(location = 1) out vec4 fragColor;
layout(location = 2) out vec2 fragRectSize;
layout(location = 3) out vec4 fragRadius;

// sRGB to linear conversion (GPU will encode back to sRGB on framebuffer write)
vec3 srgbToLinear(vec3 c) {
    return mix(c / 12.92, pow((c + 0.055) / 1.055, vec3(2.4)), step(0.04045, c));
}

void main() {
    gl_Position = vec4(inPos, 0.0, 1.0);
    fragUV = inUV;
    fragColor = vec4(srgbToLinear(inColor.rgb), inColor.a);
    fragRectSize = inRectSize;
    fragRadius = inRadius;
}

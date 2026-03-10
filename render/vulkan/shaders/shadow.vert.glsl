#version 450

// Vertex inputs for shadow rendering.
// The shadow quad covers the element bounds expanded by blur+spread on all sides.
// UV (0,0)=(element top-left), UV (1,1)=(element bottom-right); outside [0,1] = blur region.
layout(location = 0) in vec2 inPos;      // NDC position
layout(location = 1) in vec2 inUV;       // UV relative to element (0,0)..(1,1), blur region goes outside
layout(location = 2) in vec4 inColor;    // Shadow color (sRGB + alpha)
layout(location = 3) in vec2 inElemSize; // Spread-adjusted element size in physical px (W, H)
layout(location = 4) in vec4 inRadii;    // Corner radii in physical px (TL, TR, BR, BL), spread-adjusted
layout(location = 5) in float inBlur;    // Blur radius in physical px

layout(location = 0) out vec2 fragUV;
layout(location = 1) out vec4 fragColor;
layout(location = 2) out vec2 fragElemSize;
layout(location = 3) out vec4 fragRadii;
layout(location = 4) out float fragBlur;

// sRGB to linear conversion for correct blending in linear-space framebuffer.
vec3 srgbToLinear(vec3 c) {
    return mix(c / 12.92, pow((c + 0.055) / 1.055, vec3(2.4)), step(0.04045, c));
}

void main() {
    gl_Position = vec4(inPos, 0.0, 1.0);
    fragUV = inUV;
    fragColor = vec4(srgbToLinear(inColor.rgb), inColor.a);
    fragElemSize = inElemSize;
    fragRadii = inRadii;
    fragBlur = inBlur;
}

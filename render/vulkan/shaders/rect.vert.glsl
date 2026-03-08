#version 450

// Vertex inputs matching RectVertex struct layout
layout(location = 0) in vec2 inPos;        // NDC position
layout(location = 1) in vec2 inUV;         // UV for SDF (0..1)
layout(location = 2) in vec4 inColor;      // Fill color (premultiplied alpha)
layout(location = 3) in vec2 inRectSize;   // Rect size in pixels (for SDF)
layout(location = 4) in vec4 inRadii;      // Corner radii (TL, TR, BR, BL)
layout(location = 5) in float inBorderWidth;
layout(location = 6) in vec4 inBorderColor;

// Outputs to fragment shader
layout(location = 0) out vec2 fragUV;
layout(location = 1) out vec4 fragColor;
layout(location = 2) out vec2 fragRectSize;
layout(location = 3) out vec4 fragRadii;
layout(location = 4) out float fragBorderWidth;
layout(location = 5) out vec4 fragBorderColor;

// sRGB to linear conversion (GPU will encode back to sRGB on framebuffer write)
vec3 srgbToLinear(vec3 c) {
    return mix(c / 12.92, pow((c + 0.055) / 1.055, vec3(2.4)), step(0.04045, c));
}

void main() {
    gl_Position = vec4(inPos, 0.0, 1.0);
    fragUV = inUV;
    fragColor = vec4(srgbToLinear(inColor.rgb), inColor.a);
    fragRectSize = inRectSize;
    fragRadii = inRadii;
    fragBorderWidth = inBorderWidth;
    fragBorderColor = vec4(srgbToLinear(inBorderColor.rgb), inBorderColor.a);
}

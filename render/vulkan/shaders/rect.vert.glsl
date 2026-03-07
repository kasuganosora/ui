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

void main() {
    gl_Position = vec4(inPos, 0.0, 1.0);
    fragUV = inUV;
    fragColor = inColor;
    fragRectSize = inRectSize;
    fragRadii = inRadii;
    fragBorderWidth = inBorderWidth;
    fragBorderColor = inBorderColor;
}

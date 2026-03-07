#version 450

layout(location = 0) in vec2 inPos;    // NDC position
layout(location = 1) in vec2 inUV;     // texture UV
layout(location = 2) in vec4 inColor;  // tint color

layout(location = 0) out vec2 fragUV;
layout(location = 1) out vec4 fragColor;

void main() {
    gl_Position = vec4(inPos, 0.0, 1.0);
    fragUV = inUV;
    fragColor = inColor;
}

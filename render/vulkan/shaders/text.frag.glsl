#version 450

layout(location = 0) in vec2 fragUV;
layout(location = 1) in vec4 fragColor;

layout(binding = 0) uniform sampler2D glyphAtlas;

layout(location = 0) out vec4 outColor;

void main() {
    // SDF glyph atlas: single-channel distance field
    float dist = texture(glyphAtlas, fragUV).r;

    // SDF threshold with anti-aliasing
    float aa = fwidth(dist);
    float alpha = smoothstep(0.5 - aa, 0.5 + aa, dist);

    outColor = vec4(fragColor.rgb, fragColor.a * alpha);
}

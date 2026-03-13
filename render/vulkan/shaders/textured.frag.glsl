#version 450

layout(location = 0) in vec2 fragUV;
layout(location = 1) in vec4 fragColor;
layout(location = 2) in vec2 fragRectSize;
layout(location = 3) in vec4 fragRadius;

layout(binding = 0) uniform sampler2D tex;

layout(location = 0) out vec4 outColor;

float roundedBoxSDF(vec2 p, vec2 b, vec4 r) {
    float rad = (p.x > 0.0) ? ((p.y > 0.0) ? r.z : r.y) : ((p.y > 0.0) ? r.w : r.x);
    vec2 q = abs(p) - b + vec2(rad);
    return min(max(q.x, q.y), 0.0) + length(max(q, 0.0)) - rad;
}

void main() {
    vec4 texel = texture(tex, fragUV);
    vec4 result = texel * fragColor;

    if (fragRectSize.x > 0.0) {
        vec2 p = (fragUV - 0.5) * fragRectSize;
        vec2 b = fragRectSize * 0.5;
        float dist = roundedBoxSDF(p, b, fragRadius);
        float aa = fwidth(dist);
        result.a *= 1.0 - smoothstep(0.0, aa, dist);
    }

    outColor = result;
}

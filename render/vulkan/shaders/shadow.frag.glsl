#version 450

layout(location = 0) in vec2 fragUV;
layout(location = 1) in vec4 fragColor;
layout(location = 2) in vec2 fragElemSize;  // Spread-adjusted element size in physical px
layout(location = 3) in vec4 fragRadii;     // Corner radii in physical px (TL, TR, BR, BL)
layout(location = 4) in float fragBlur;     // Blur radius in physical px

layout(location = 0) out vec4 outColor;

// SDF for a rounded rectangle.
// p: point relative to rect center (physical px)
// b: rect half-size (physical px)
// r: corner radii (TL, TR, BR, BL)
float roundedRectSDF(vec2 p, vec2 b, vec4 r) {
    float radius = (p.x > 0.0)
        ? ((p.y > 0.0) ? r.z : r.y)
        : ((p.y > 0.0) ? r.w : r.x);
    vec2 q = abs(p) - b + vec2(radius);
    return min(max(q.x, q.y), 0.0) + length(max(q, 0.0)) - radius;
}

void main() {
    // Map UV to position relative to element center in physical px.
    // UV (0,0)=element top-left, UV (1,1)=element bottom-right.
    // Outside [0,1] range = blur/spread area.
    vec2 elemHalf = fragElemSize * 0.5;
    vec2 p = (fragUV - 0.5) * fragElemSize;  // position relative to element center

    float dist = roundedRectSDF(p, elemHalf, fragRadii);

    // Alpha falloff: gaussian approximation of CSS blur.
    float blur = max(fragBlur, 0.0);
    float alpha;
    // Guard against fwidth=0 (uniform regions) to avoid smoothstep UB.
    float aa = max(fwidth(dist), 0.5);

    if (blur < 1.0) {
        // Hard edge shadow (blur ≈ 0), with sub-pixel AA.
        alpha = 1.0 - smoothstep(-aa, aa, dist);
    } else {
        // Soft shadow: gaussian-like falloff using squared exponential.
        // Inside the shape (dist < 0): full alpha.
        // Outside: falloff over the blur distance.
        float sigma = blur * 0.5;  // effective gaussian sigma
        float outside = max(0.0, dist);
        float t = outside / sigma;
        alpha = exp(-t * t * 0.5);  // Gaussian: e^(-t²/2)
        // Fade out sharply beyond 3-sigma from the element boundary.
        // (dist - sigma*3) transitions from negative (keep) to positive (cut) at 3-sigma distance.
        alpha *= 1.0 - smoothstep(-aa, aa, dist - sigma * 3.0);
        alpha = clamp(alpha, 0.0, 1.0);
    }

    outColor = vec4(fragColor.rgb, fragColor.a * alpha);
}

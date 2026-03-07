#version 450

layout(location = 0) in vec2 fragUV;
layout(location = 1) in vec4 fragColor;
layout(location = 2) in vec2 fragRectSize;
layout(location = 3) in vec4 fragRadii;
layout(location = 4) in float fragBorderWidth;
layout(location = 5) in vec4 fragBorderColor;

layout(location = 0) out vec4 outColor;

// SDF for a rounded rectangle.
// p: point relative to rect center
// b: rect half-size
// r: corner radii (TL, TR, BR, BL) - we pick the correct one based on quadrant
float roundedRectSDF(vec2 p, vec2 b, vec4 r) {
    // Select radius based on which quadrant p is in
    float radius = (p.x > 0.0)
        ? ((p.y > 0.0) ? r.z : r.y)   // right: BR or TR
        : ((p.y > 0.0) ? r.w : r.x);  // left:  BL or TL

    vec2 q = abs(p) - b + vec2(radius);
    return min(max(q.x, q.y), 0.0) + length(max(q, 0.0)) - radius;
}

void main() {
    vec2 halfSize = fragRectSize * 0.5;

    // Map UV (0..1) to rect-local coordinates centered at origin
    vec2 p = (fragUV - 0.5) * fragRectSize;

    float dist = roundedRectSDF(p, halfSize, fragRadii);

    // Anti-aliased edge (1 pixel smoothstep)
    float aa = fwidth(dist);
    float fillAlpha = 1.0 - smoothstep(-aa, aa, dist);

    if (fragBorderWidth > 0.0) {
        // Border: the region between dist and dist+borderWidth
        float innerDist = dist + fragBorderWidth;
        float borderAlpha = 1.0 - smoothstep(-aa, aa, innerDist);

        // Inside the border ring: use border color
        // Inside the fill area (past border): use fill color
        float fillMask = 1.0 - smoothstep(-aa, aa, innerDist);
        vec4 color = mix(fragBorderColor, fragColor, fillMask);
        outColor = vec4(color.rgb, color.a * fillAlpha);
    } else {
        outColor = vec4(fragColor.rgb, fragColor.a * fillAlpha);
    }
}

//go:build darwin

package metal

// metalShaderSource contains all MSL shader source code for the Metal backend.
// It includes:
//   - Shared helpers: srgbToLinear(), roundedRectSDF()
//   - Rect pipeline: rectVertex + rectFragment (SDF rounded rects)
//   - Shadow pipeline: shadowVertex + shadowFragment (gaussian box shadow)
//   - Textured pipeline: texturedVertex + textFragment (SDF text atlas, R8)
//   - Image pipeline: texturedVertex + imageFragment (RGBA/BGRA texture)
//
// NDC coordinate system: Metal uses Y-up clip space, so ndcY = 1 - (y/logH)*2.
const metalShaderSource = `
#include <metal_stdlib>
using namespace metal;

// ---- Shared helpers ----

// Convert sRGB color component to linear light.
// Alpha channel is NOT converted (doesn't participate in sRGB).
inline float3 srgbToLinear(float3 c) {
    return select(
        c / 12.92,
        pow((c + 0.055) / 1.055, float3(2.4)),
        c > 0.04045
    );
}

// SDF for a rounded rectangle.
// p:  position relative to rect center (in physical/logical pixels)
// b:  half-size of the rect
// r:  per-corner radii (TL, TR, BR, BL)
inline float roundedRectSDF(float2 p, float2 b, float4 r) {
    float radius = (p.x > 0.0)
        ? ((p.y > 0.0) ? r.z : r.y)
        : ((p.y > 0.0) ? r.w : r.x);
    float2 q = abs(p) - b + float2(radius);
    return min(max(q.x, q.y), 0.0) + length(max(q, float2(0.0))) - radius;
}

// ============================================================
// RECT PIPELINE
// ============================================================

// Input from vertex buffer (RectVertex: 19 float32 = 76 bytes)
struct RectVertexIn {
    float2 pos         [[attribute(0)]];
    float2 uv          [[attribute(1)]];
    float4 color       [[attribute(2)]];
    float2 rectSize    [[attribute(3)]];
    float4 radii       [[attribute(4)]];
    float  borderWidth [[attribute(5)]];
    float4 borderColor [[attribute(6)]];
};

// Interpolated output from rectVertex to rectFragment
struct RectVarying {
    float4 clipPos     [[position]];
    float2 uv;
    float4 color;
    float2 rectSize;
    float4 radii;
    float  borderWidth;
    float4 borderColor;
};

vertex RectVarying rectVertex(RectVertexIn in [[stage_in]]) {
    RectVarying out;
    // pos is already NDC (computed by CPU in Go)
    out.clipPos    = float4(in.pos.x, in.pos.y, 0.0, 1.0);
    out.uv         = in.uv;
    // Convert sRGB colors to linear for correct blending in sRGB framebuffer
    out.color      = float4(srgbToLinear(in.color.rgb), in.color.a);
    out.rectSize   = in.rectSize;
    out.radii      = in.radii;
    out.borderWidth = in.borderWidth;
    out.borderColor = float4(srgbToLinear(in.borderColor.rgb), in.borderColor.a);
    return out;
}

fragment float4 rectFragment(RectVarying in [[stage_in]]) {
    float2 halfSize = in.rectSize * 0.5;
    // Map UV (0..1) to rect-local coordinates centered at origin
    float2 p = (in.uv - 0.5) * in.rectSize;

    float dist = roundedRectSDF(p, halfSize, in.radii);

    // Anti-aliased edge: AA band extends OUTSIDE the shape boundary only.
    // At dist=0 (the geometric edge), fillAlpha = 1.0 (fully opaque).
    float aa = fwidth(dist);
    float fillAlpha = 1.0 - smoothstep(0.0, aa, dist);

    if (in.borderWidth > 0.0) {
        float innerDist = dist + in.borderWidth;
        float fillMask  = 1.0 - smoothstep(0.0, aa, innerDist);
        float4 color    = mix(in.borderColor, in.color, fillMask);
        return float4(color.rgb, color.a * fillAlpha);
    } else {
        return float4(in.color.rgb, in.color.a * fillAlpha);
    }
}

// ============================================================
// SHADOW PIPELINE
// ============================================================

// Input from vertex buffer (ShadowVertex: 15 float32 = 60 bytes)
struct ShadowVertexIn {
    float2 pos     [[attribute(0)]];
    float2 uv      [[attribute(1)]];
    float4 color   [[attribute(2)]];
    float2 elemSize [[attribute(3)]];
    float4 radii   [[attribute(4)]];
    float  blur    [[attribute(5)]];
};

struct ShadowVarying {
    float4 clipPos  [[position]];
    float2 uv;
    float4 color;
    float2 elemSize;
    float4 radii;
    float  blur;
};

vertex ShadowVarying shadowVertex(ShadowVertexIn in [[stage_in]]) {
    ShadowVarying out;
    out.clipPos  = float4(in.pos.x, in.pos.y, 0.0, 1.0);
    out.uv       = in.uv;
    // Shadow color stays in sRGB (fragment handles alpha only)
    out.color    = float4(srgbToLinear(in.color.rgb), in.color.a);
    out.elemSize = in.elemSize;
    out.radii    = in.radii;
    out.blur     = in.blur;
    return out;
}

fragment float4 shadowFragment(ShadowVarying in [[stage_in]]) {
    // Map UV to position relative to element center in physical px.
    // UV (0,0) = element top-left, UV (1,1) = element bottom-right.
    // Outside [0,1] range = blur/spread area.
    float2 elemHalf = in.elemSize * 0.5;
    float2 p = (in.uv - 0.5) * in.elemSize; // position relative to element center

    float dist = roundedRectSDF(p, elemHalf, in.radii);

    float blur_r = max(in.blur, 0.0);
    float alpha;
    float aa = max(fwidth(dist), 0.5);

    if (blur_r < 1.0) {
        // Hard edge shadow (blur ~ 0), with sub-pixel AA.
        alpha = 1.0 - smoothstep(-aa, aa, dist);
    } else {
        // Soft shadow: gaussian-like falloff using squared exponential.
        // Inside the shape (dist < 0): full alpha.
        // Outside: falloff over the blur distance.
        float sigma   = blur_r * 0.5; // effective gaussian sigma
        float outside = max(0.0, dist);
        float t       = outside / sigma;
        alpha = exp(-t * t * 0.5); // Gaussian: e^(-t^2/2)
        // Fade out sharply beyond 3-sigma from the element boundary.
        alpha *= 1.0 - smoothstep(-aa, aa, dist - sigma * 3.0);
        alpha = clamp(alpha, 0.0, 1.0);
    }

    return float4(in.color.rgb, in.color.a * alpha);
}

// ============================================================
// TEXTURED PIPELINE (shared vertex shader for text and image)
// ============================================================

// Input from vertex buffer (TexturedVertex: 8 float32 = 32 bytes)
struct TexturedVertexIn {
    float2 pos   [[attribute(0)]];
    float2 uv    [[attribute(1)]];
    float4 color [[attribute(2)]];
};

struct TexturedVarying {
    float4 clipPos [[position]];
    float2 uv;
    float4 color;
};

vertex TexturedVarying texturedVertex(TexturedVertexIn in [[stage_in]]) {
    TexturedVarying out;
    out.clipPos = float4(in.pos.x, in.pos.y, 0.0, 1.0);
    out.uv      = in.uv;
    // Color is pre-multiplied sRGB; convert to linear for sRGB framebuffer
    out.color   = float4(srgbToLinear(in.color.rgb), in.color.a);
    return out;
}

// textFragment: SDF text rendering using an R8 glyph atlas.
// The R8 texture encodes a signed distance field.
// Alpha is derived from the SDF value with sub-pixel AA.
fragment float4 textFragment(
    TexturedVarying   in  [[stage_in]],
    texture2d<float>  tex [[texture(0)]],
    sampler           smp [[sampler(0)]])
{
    float sdf = tex.sample(smp, in.uv).r;
    // smoothstep AA over 0.5 (the SDF boundary)
    float aa    = fwidth(sdf) * 0.5;
    float alpha = smoothstep(0.5 - aa, 0.5 + aa, sdf);
    return float4(in.color.rgb, in.color.a * alpha);
}

// imageFragment: RGBA/BGRA texture rendering (images, icons, etc.).
// Tint color is multiplied with the sampled texel (premultiplied alpha approach).
fragment float4 imageFragment(
    TexturedVarying   in  [[stage_in]],
    texture2d<float>  tex [[texture(0)]],
    sampler           smp [[sampler(0)]])
{
    float4 texel = tex.sample(smp, in.uv);
    // Convert sampled texel from sRGB to linear (Metal doesn't auto-convert R8Unorm / RGBA8Unorm)
    float4 linearTexel = float4(srgbToLinear(texel.rgb), texel.a);
    // Apply tint: multiply colors, use tint alpha as overall opacity
    return float4(linearTexel.rgb * in.color.rgb, linearTexel.a * in.color.a);
}
`

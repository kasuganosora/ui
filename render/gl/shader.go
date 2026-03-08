//go:build windows

package gl

import (
	"fmt"
	"unsafe"
)

// Shader sources for OpenGL 3.3 Core.
// These are functionally identical to the Vulkan GLSL shaders.

const rectVertSrc = `#version 330 core
layout(location = 0) in vec2 inPos;
layout(location = 1) in vec2 inUV;
layout(location = 2) in vec4 inColor;
layout(location = 3) in vec2 inRectSize;
layout(location = 4) in vec4 inRadii;
layout(location = 5) in float inBorderWidth;
layout(location = 6) in vec4 inBorderColor;

out vec2 fragUV;
out vec4 fragColor;
out vec2 fragRectSize;
out vec4 fragRadii;
out float fragBorderWidth;
out vec4 fragBorderColor;

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
`

const rectFragSrc = `#version 330 core
in vec2 fragUV;
in vec4 fragColor;
in vec2 fragRectSize;
in vec4 fragRadii;
in float fragBorderWidth;
in vec4 fragBorderColor;

out vec4 outColor;

float roundedRectSDF(vec2 p, vec2 b, vec4 r) {
    float radius = (p.x > 0.0)
        ? ((p.y > 0.0) ? r.z : r.y)
        : ((p.y > 0.0) ? r.w : r.x);
    vec2 q = abs(p) - b + vec2(radius);
    return min(max(q.x, q.y), 0.0) + length(max(q, 0.0)) - radius;
}

void main() {
    vec2 halfSize = fragRectSize * 0.5;
    vec2 p = (fragUV - 0.5) * fragRectSize;
    float dist = roundedRectSDF(p, halfSize, fragRadii);
    float aa = fwidth(dist);
    float fillAlpha = 1.0 - smoothstep(0.0, aa, dist);

    if (fragBorderWidth > 0.0) {
        float innerDist = dist + fragBorderWidth;
        float fillMask = 1.0 - smoothstep(0.0, aa, innerDist);
        vec4 color = mix(fragBorderColor, fragColor, fillMask);
        outColor = vec4(color.rgb, color.a * fillAlpha);
    } else {
        outColor = vec4(fragColor.rgb, fragColor.a * fillAlpha);
    }
}
`

const texturedVertSrc = `#version 330 core
layout(location = 0) in vec2 inPos;
layout(location = 1) in vec2 inUV;
layout(location = 2) in vec4 inColor;

out vec2 fragUV;
out vec4 fragColor;

vec3 srgbToLinear(vec3 c) {
    return mix(c / 12.92, pow((c + 0.055) / 1.055, vec3(2.4)), step(0.04045, c));
}

void main() {
    gl_Position = vec4(inPos, 0.0, 1.0);
    fragUV = inUV;
    fragColor = vec4(srgbToLinear(inColor.rgb), inColor.a);
}
`

const texturedFragSrc = `#version 330 core
in vec2 fragUV;
in vec4 fragColor;

uniform sampler2D tex;

out vec4 outColor;

void main() {
    outColor = texture(tex, fragUV) * fragColor;
}
`

const textFragSrc = `#version 330 core
in vec2 fragUV;
in vec4 fragColor;

uniform sampler2D glyphAtlas;

out vec4 outColor;

void main() {
    float coverage = texture(glyphAtlas, fragUV).r;
    outColor = vec4(fragColor.rgb, fragColor.a * coverage);
}
`

// compileShader compiles a GLSL shader and returns its ID.
func (l *Loader) compileShader(shaderType uint32, source string) (uint32, error) {
	id := uint32(glCall(l.glCreateShader, uintptr(shaderType)))
	if id == 0 {
		return 0, fmt.Errorf("gl: glCreateShader failed")
	}

	// glShaderSource(id, 1, &srcPtr, &srcLen)
	srcBytes := append([]byte(source), 0)
	srcPtr := unsafe.Pointer(&srcBytes[0])
	srcLen := int32(len(source))
	glCall(l.glShaderSource, uintptr(id), 1,
		uintptr(unsafe.Pointer(&srcPtr)),
		uintptr(unsafe.Pointer(&srcLen)))

	glCall(l.glCompileShader, uintptr(id))

	var status int32
	glCall(l.glGetShaderiv, uintptr(id), GL_COMPILE_STATUS, uintptr(unsafe.Pointer(&status)))
	if status == GL_FALSE {
		var logLen int32
		glCall(l.glGetShaderiv, uintptr(id), GL_INFO_LOG_LENGTH, uintptr(unsafe.Pointer(&logLen)))
		if logLen > 0 {
			logBuf := make([]byte, logLen)
			glCall(l.glGetShaderInfoLog, uintptr(id), uintptr(logLen), 0, uintptr(unsafe.Pointer(&logBuf[0])))
			glCall(l.glDeleteShader, uintptr(id))
			return 0, fmt.Errorf("gl: shader compile error: %s", string(logBuf))
		}
		glCall(l.glDeleteShader, uintptr(id))
		return 0, fmt.Errorf("gl: shader compile failed (no log)")
	}
	return id, nil
}

// linkProgram links a vertex and fragment shader into a program.
func (l *Loader) linkProgram(vertID, fragID uint32) (uint32, error) {
	id := uint32(glCall(l.glCreateProgram))
	if id == 0 {
		return 0, fmt.Errorf("gl: glCreateProgram failed")
	}

	glCall(l.glAttachShader, uintptr(id), uintptr(vertID))
	glCall(l.glAttachShader, uintptr(id), uintptr(fragID))
	glCall(l.glLinkProgram, uintptr(id))

	var status int32
	glCall(l.glGetProgramiv, uintptr(id), GL_LINK_STATUS, uintptr(unsafe.Pointer(&status)))
	if status == GL_FALSE {
		var logLen int32
		glCall(l.glGetProgramiv, uintptr(id), GL_INFO_LOG_LENGTH, uintptr(unsafe.Pointer(&logLen)))
		if logLen > 0 {
			logBuf := make([]byte, logLen)
			glCall(l.glGetProgramInfoLog, uintptr(id), uintptr(logLen), 0, uintptr(unsafe.Pointer(&logBuf[0])))
			glCall(l.glDeleteProgram, uintptr(id))
			return 0, fmt.Errorf("gl: program link error: %s", string(logBuf))
		}
		glCall(l.glDeleteProgram, uintptr(id))
		return 0, fmt.Errorf("gl: program link failed (no log)")
	}
	return id, nil
}

// createProgram compiles and links a vertex+fragment shader pair.
func (l *Loader) createProgram(vertSrc, fragSrc string) (uint32, error) {
	vs, err := l.compileShader(GL_VERTEX_SHADER, vertSrc)
	if err != nil {
		return 0, err
	}
	defer glCall(l.glDeleteShader, uintptr(vs))

	fs, err := l.compileShader(GL_FRAGMENT_SHADER, fragSrc)
	if err != nil {
		return 0, err
	}
	defer glCall(l.glDeleteShader, uintptr(fs))

	return l.linkProgram(vs, fs)
}

// getUniformLocation returns the location of a uniform variable.
func (l *Loader) getUniformLocation(program uint32, name string) int32 {
	cstr := append([]byte(name), 0)
	return int32(glCall(l.glGetUniformLocation, uintptr(program), uintptr(unsafe.Pointer(&cstr[0]))))
}

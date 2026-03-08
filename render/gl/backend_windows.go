//go:build windows

package gl

import (
	"fmt"
	"unsafe"
)

// createContext creates a WGL OpenGL 3.3 core profile context for the given HWND.
// Returns (hdc, hglrc) on success.
func (b *Backend) createContext(hwnd uintptr) (uintptr, uintptr, error) {
	l := b.loader

	// Get device context
	hdc, _, _ := l.getDC.Call(hwnd)
	if hdc == 0 {
		return 0, 0, fmt.Errorf("gl: GetDC failed")
	}

	// Set pixel format
	pfd := PIXELFORMATDESCRIPTOR{
		NSize:      uint16(unsafe.Sizeof(PIXELFORMATDESCRIPTOR{})),
		NVersion:   1,
		DwFlags:    PFD_DRAW_TO_WINDOW | PFD_SUPPORT_OPENGL | PFD_DOUBLEBUFFER,
		IPixelType: PFD_TYPE_RGBA,
		CColorBits: 32,
		CAlphaBits: 8,
		CDepthBits: 0,
		ILayerType: PFD_MAIN_PLANE,
	}

	pf, _, _ := l.choosePF.Call(hdc, uintptr(unsafe.Pointer(&pfd)))
	if pf == 0 {
		return 0, 0, fmt.Errorf("gl: ChoosePixelFormat failed")
	}
	r, _, _ := l.setPF.Call(hdc, pf, uintptr(unsafe.Pointer(&pfd)))
	if r == 0 {
		return 0, 0, fmt.Errorf("gl: SetPixelFormat failed")
	}

	// Create legacy context first (needed to get wglCreateContextAttribsARB)
	legacyRC, _, _ := l.wglCreateContext.Call(hdc)
	if legacyRC == 0 {
		return 0, 0, fmt.Errorf("gl: wglCreateContext failed")
	}
	r, _, _ = l.wglMakeCurrent.Call(hdc, legacyRC)
	if r == 0 {
		l.wglDeleteContext.Call(legacyRC)
		return 0, 0, fmt.Errorf("gl: wglMakeCurrent failed for legacy context")
	}

	// Load extension functions with legacy context
	l.LoadCoreFunctions()

	// Try to create 3.3 core profile context
	wglCreateContextAttribs := l.getProc("wglCreateContextAttribsARB")
	if wglCreateContextAttribs != 0 {
		attribs := [...]int32{
			WGL_CONTEXT_MAJOR_VERSION_ARB, 3,
			WGL_CONTEXT_MINOR_VERSION_ARB, 3,
			WGL_CONTEXT_PROFILE_MASK_ARB, WGL_CONTEXT_CORE_PROFILE_BIT_ARB,
			0, // terminator
		}
		coreRC := glCall(wglCreateContextAttribs, hdc, 0,
			uintptr(unsafe.Pointer(&attribs[0])))
		if coreRC != 0 {
			// Switch to core context
			l.wglMakeCurrent.Call(0, 0)
			l.wglDeleteContext.Call(legacyRC)
			r, _, _ = l.wglMakeCurrent.Call(hdc, coreRC)
			if r == 0 {
				l.wglDeleteContext.Call(coreRC)
				return 0, 0, fmt.Errorf("gl: wglMakeCurrent failed for core context")
			}
			return hdc, coreRC, nil
		}
	}

	// Fall back to legacy context (may still support GL 3.3 on some drivers)
	return hdc, legacyRC, nil
}

// destroyContext releases the WGL context.
func (b *Backend) destroyContext() {
	if b.hglrc != 0 {
		b.loader.wglMakeCurrent.Call(0, 0)
		b.loader.wglDeleteContext.Call(b.hglrc)
		b.hglrc = 0
	}
	if b.hdc != 0 && b.hwnd != 0 {
		b.loader.releaseDC.Call(b.hwnd, b.hdc)
		b.hdc = 0
	}
}

// swapBuffers presents the frame.
func (b *Backend) swapBuffers() {
	b.loader.wglSwapBuffers.Call(b.hdc)
}

//go:build windows

package gl

import (
	"fmt"
	"syscall"
	"unsafe"
)

// Loader dynamically loads OpenGL functions via opengl32.dll + wglGetProcAddress.
// Zero-CGO: all calls go through syscall.SyscallN.
type Loader struct {
	opengl32 *syscall.LazyDLL
	gdi32    *syscall.LazyDLL

	// WGL functions (from opengl32.dll)
	wglCreateContext  *syscall.LazyProc
	wglDeleteContext  *syscall.LazyProc
	wglMakeCurrent   *syscall.LazyProc
	wglGetProcAddress *syscall.LazyProc
	wglSwapBuffers    *syscall.LazyProc // actually SwapBuffers from gdi32

	// GDI functions
	choosePF  *syscall.LazyProc // ChoosePixelFormat
	setPF     *syscall.LazyProc // SetPixelFormat
	describePF *syscall.LazyProc // DescribePixelFormat
	getDC          *syscall.LazyProc // from user32
	releaseDC      *syscall.LazyProc // from user32
	getClientRect  *syscall.LazyProc // from user32

	user32 *syscall.LazyDLL

	// OpenGL 1.x functions (from opengl32.dll)
	glGetIntegerv uintptr
	glEnable      uintptr
	glDisable     uintptr
	glViewport    uintptr
	glScissor     uintptr
	glClearColor  uintptr
	glClear       uintptr
	glGetError    uintptr
	glBlendFunc   uintptr
	glPixelStorei uintptr
	glTexImage2D  uintptr
	glTexSubImage2D uintptr
	glTexParameteri uintptr
	glGenTextures   uintptr
	glDeleteTextures uintptr
	glBindTexture   uintptr
	glActiveTexture uintptr
	glDrawArrays    uintptr
	glReadPixels    uintptr
	glFlush         uintptr
	glFinish        uintptr

	// OpenGL 2.0+ extension functions (via wglGetProcAddress)
	glCreateShader       uintptr
	glShaderSource       uintptr
	glCompileShader      uintptr
	glGetShaderiv        uintptr
	glGetShaderInfoLog   uintptr
	glDeleteShader       uintptr
	glCreateProgram      uintptr
	glAttachShader       uintptr
	glLinkProgram        uintptr
	glGetProgramiv       uintptr
	glGetProgramInfoLog  uintptr
	glUseProgram         uintptr
	glDeleteProgram      uintptr
	glGetUniformLocation uintptr
	glUniform1i          uintptr
	glGenBuffers         uintptr
	glDeleteBuffers      uintptr
	glBindBuffer         uintptr
	glBufferData         uintptr
	glBufferSubData      uintptr
	glMapBufferRange     uintptr
	glUnmapBuffer        uintptr
	glGenVertexArrays    uintptr
	glDeleteVertexArrays uintptr
	glBindVertexArray    uintptr
	glEnableVertexAttribArray  uintptr
	glDisableVertexAttribArray uintptr
	glVertexAttribPointer      uintptr
	glBlendFuncSeparate  uintptr
	glBlendEquation      uintptr

	// Framebuffer (for sRGB and readback)
	glGenFramebuffers    uintptr
	glDeleteFramebuffers uintptr
	glBindFramebuffer    uintptr
	glFramebufferTexture2D uintptr
	glCheckFramebufferStatus uintptr
}

// NewLoader creates a new OpenGL loader.
func NewLoader() (*Loader, error) {
	l := &Loader{
		opengl32: syscall.NewLazyDLL("opengl32.dll"),
		gdi32:    syscall.NewLazyDLL("gdi32.dll"),
		user32:   syscall.NewLazyDLL("user32.dll"),
	}

	if err := l.opengl32.Load(); err != nil {
		return nil, fmt.Errorf("gl: failed to load opengl32.dll: %w", err)
	}

	// WGL functions
	l.wglCreateContext = l.opengl32.NewProc("wglCreateContext")
	l.wglDeleteContext = l.opengl32.NewProc("wglDeleteContext")
	l.wglMakeCurrent = l.opengl32.NewProc("wglMakeCurrent")
	l.wglGetProcAddress = l.opengl32.NewProc("wglGetProcAddress")

	// GDI
	l.wglSwapBuffers = l.gdi32.NewProc("SwapBuffers")
	l.choosePF = l.gdi32.NewProc("ChoosePixelFormat")
	l.setPF = l.gdi32.NewProc("SetPixelFormat")
	l.describePF = l.gdi32.NewProc("DescribePixelFormat")
	l.getDC = l.user32.NewProc("GetDC")
	l.releaseDC = l.user32.NewProc("ReleaseDC")
	l.getClientRect = l.user32.NewProc("GetClientRect")

	return l, nil
}

// getProc loads an OpenGL extension function via wglGetProcAddress.
// Falls back to GetProcAddress on opengl32.dll for GL 1.x functions.
func (l *Loader) getProc(name string) uintptr {
	cstr := append([]byte(name), 0)
	r, _, _ := l.wglGetProcAddress.Call(uintptr(unsafe.Pointer(&cstr[0])))
	if r != 0 {
		return r
	}
	// Fallback: some GL 1.x functions are in opengl32.dll
	proc := l.opengl32.NewProc(name)
	if err := proc.Find(); err == nil {
		return proc.Addr()
	}
	return 0
}

// LoadCoreFunctions loads all GL 1.x functions from opengl32.dll.
func (l *Loader) LoadCoreFunctions() {
	l.glGetIntegerv = l.getProc("glGetIntegerv")
	l.glEnable = l.getProc("glEnable")
	l.glDisable = l.getProc("glDisable")
	l.glViewport = l.getProc("glViewport")
	l.glScissor = l.getProc("glScissor")
	l.glClearColor = l.getProc("glClearColor")
	l.glClear = l.getProc("glClear")
	l.glGetError = l.getProc("glGetError")
	l.glBlendFunc = l.getProc("glBlendFunc")
	l.glPixelStorei = l.getProc("glPixelStorei")
	l.glTexImage2D = l.getProc("glTexImage2D")
	l.glTexSubImage2D = l.getProc("glTexSubImage2D")
	l.glTexParameteri = l.getProc("glTexParameteri")
	l.glGenTextures = l.getProc("glGenTextures")
	l.glDeleteTextures = l.getProc("glDeleteTextures")
	l.glBindTexture = l.getProc("glBindTexture")
	l.glActiveTexture = l.getProc("glActiveTexture")
	l.glDrawArrays = l.getProc("glDrawArrays")
	l.glReadPixels = l.getProc("glReadPixels")
	l.glFlush = l.getProc("glFlush")
	l.glFinish = l.getProc("glFinish")
}

// LoadExtensionFunctions loads GL 2.0+ / 3.3 functions via wglGetProcAddress.
// Must be called after a GL context is current.
func (l *Loader) LoadExtensionFunctions() error {
	l.glCreateShader = l.getProc("glCreateShader")
	l.glShaderSource = l.getProc("glShaderSource")
	l.glCompileShader = l.getProc("glCompileShader")
	l.glGetShaderiv = l.getProc("glGetShaderiv")
	l.glGetShaderInfoLog = l.getProc("glGetShaderInfoLog")
	l.glDeleteShader = l.getProc("glDeleteShader")
	l.glCreateProgram = l.getProc("glCreateProgram")
	l.glAttachShader = l.getProc("glAttachShader")
	l.glLinkProgram = l.getProc("glLinkProgram")
	l.glGetProgramiv = l.getProc("glGetProgramiv")
	l.glGetProgramInfoLog = l.getProc("glGetProgramInfoLog")
	l.glUseProgram = l.getProc("glUseProgram")
	l.glDeleteProgram = l.getProc("glDeleteProgram")
	l.glGetUniformLocation = l.getProc("glGetUniformLocation")
	l.glUniform1i = l.getProc("glUniform1i")
	l.glGenBuffers = l.getProc("glGenBuffers")
	l.glDeleteBuffers = l.getProc("glDeleteBuffers")
	l.glBindBuffer = l.getProc("glBindBuffer")
	l.glBufferData = l.getProc("glBufferData")
	l.glBufferSubData = l.getProc("glBufferSubData")
	l.glMapBufferRange = l.getProc("glMapBufferRange")
	l.glUnmapBuffer = l.getProc("glUnmapBuffer")
	l.glGenVertexArrays = l.getProc("glGenVertexArrays")
	l.glDeleteVertexArrays = l.getProc("glDeleteVertexArrays")
	l.glBindVertexArray = l.getProc("glBindVertexArray")
	l.glEnableVertexAttribArray = l.getProc("glEnableVertexAttribArray")
	l.glDisableVertexAttribArray = l.getProc("glDisableVertexAttribArray")
	l.glVertexAttribPointer = l.getProc("glVertexAttribPointer")
	l.glBlendFuncSeparate = l.getProc("glBlendFuncSeparate")
	l.glBlendEquation = l.getProc("glBlendEquation")
	l.glGenFramebuffers = l.getProc("glGenFramebuffers")
	l.glDeleteFramebuffers = l.getProc("glDeleteFramebuffers")
	l.glBindFramebuffer = l.getProc("glBindFramebuffer")
	l.glFramebufferTexture2D = l.getProc("glFramebufferTexture2D")
	l.glCheckFramebufferStatus = l.getProc("glCheckFramebufferStatus")

	// Verify critical functions loaded
	if l.glCreateShader == 0 || l.glCreateProgram == 0 || l.glGenVertexArrays == 0 {
		return fmt.Errorf("gl: failed to load required OpenGL 3.3 functions (no GL 3.3 context?)")
	}
	return nil
}

// glCall wraps syscall.SyscallN for GL function pointers.
func glCall(fn uintptr, args ...uintptr) uintptr {
	r, _, _ := syscall.SyscallN(fn, args...)
	return r
}

// PIXELFORMATDESCRIPTOR for SetPixelFormat.
type PIXELFORMATDESCRIPTOR struct {
	NSize           uint16
	NVersion        uint16
	DwFlags         uint32
	IPixelType      byte
	CColorBits      byte
	CRedBits        byte
	CRedShift       byte
	CGreenBits      byte
	CGreenShift     byte
	CBlueBits       byte
	CBlueShift      byte
	CAlphaBits      byte
	CAlphaShift     byte
	CAccumBits      byte
	CAccumRedBits   byte
	CAccumGreenBits byte
	CAccumBlueBits  byte
	CAccumAlphaBits byte
	CDepthBits      byte
	CStencilBits    byte
	CAuxBuffers     byte
	ILayerType      byte
	BReserved       byte
	DwLayerMask     uint32
	DwVisibleMask   uint32
	DwDamageMask    uint32
}

// PFD flags
const (
	PFD_DRAW_TO_WINDOW = 0x00000004
	PFD_SUPPORT_OPENGL = 0x00000020
	PFD_DOUBLEBUFFER   = 0x00000001
	PFD_TYPE_RGBA      = 0
	PFD_MAIN_PLANE     = 0
)

// RECT for GetClientRect
type RECT struct {
	Left, Top, Right, Bottom int32
}

// GetClientRect returns the actual physical client area size of a window.
func (l *Loader) GetClientRect(hwnd uintptr) (int, int) {
	var r RECT
	l.getClientRect.Call(hwnd, uintptr(unsafe.Pointer(&r)))
	return int(r.Right - r.Left), int(r.Bottom - r.Top)
}

// WGL_ARB constants for wglCreateContextAttribsARB
const (
	WGL_CONTEXT_MAJOR_VERSION_ARB = 0x2091
	WGL_CONTEXT_MINOR_VERSION_ARB = 0x2092
	WGL_CONTEXT_PROFILE_MASK_ARB  = 0x9126
	WGL_CONTEXT_CORE_PROFILE_BIT_ARB = 0x00000001
)

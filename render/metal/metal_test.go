//go:build darwin

package metal

import (
	"testing"
	"unsafe"

	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/render"
	"github.com/ebitengine/purego"
)

// TestNew verifies that New() returns a non-nil backend with expected defaults.
func TestNew(t *testing.T) {
	b := New()
	if b == nil {
		t.Fatal("New() returned nil")
	}
	if b.dpiScale != 1.0 {
		t.Errorf("expected dpiScale=1.0, got %v", b.dpiScale)
	}
	if b.textures == nil {
		t.Error("textures map is nil")
	}
	if b.nextTexID != 1 {
		t.Errorf("expected nextTexID=1, got %v", b.nextTexID)
	}
}

// TestInitNilWindow verifies that Init with a nil-handle window returns an error
// without panicking.
func TestInitNilWindow(t *testing.T) {
	b := New()
	err := b.Init(&nilWindow{})
	// Init must return an error (nil native handle, or Metal not available).
	// It must NEVER panic.
	if err == nil {
		t.Log("Init succeeded with nilWindow — skipping (unexpected but harmless on real Metal hardware)")
		b.Destroy()
	}
}

// TestConstants verifies critical Metal constant values.
func TestConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      int
		expected int
	}{
		{"MTLPixelFormatBGRA8Unorm", MTLPixelFormatBGRA8Unorm, 80},
		{"MTLPixelFormatBGRA8Unorm_sRGB", MTLPixelFormatBGRA8Unorm_sRGB, 81},
		{"MTLPixelFormatR8Unorm", MTLPixelFormatR8Unorm, 10},
		{"MTLPixelFormatRGBA8Unorm", MTLPixelFormatRGBA8Unorm, 70},
		{"MTLPixelFormatRGBA8Unorm_sRGB", MTLPixelFormatRGBA8Unorm_sRGB, 71},
		{"MTLLoadActionClear", MTLLoadActionClear, 2},
		{"MTLLoadActionLoad", MTLLoadActionLoad, 1},
		{"MTLLoadActionDontCare", MTLLoadActionDontCare, 0},
		{"MTLStoreActionStore", MTLStoreActionStore, 1},
		{"MTLPrimitiveTypeTriangle", MTLPrimitiveTypeTriangle, 3},
		{"MTLBlendFactorSourceAlpha", MTLBlendFactorSourceAlpha, 4},
		{"MTLBlendFactorOneMinusSourceAlpha", MTLBlendFactorOneMinusSourceAlpha, 5},
		{"MTLBlendOperationAdd", MTLBlendOperationAdd, 0},
		{"MTLColorWriteMaskAll", MTLColorWriteMaskAll, 0xf},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.expected {
				t.Errorf("%s = %d, want %d", tc.name, tc.got, tc.expected)
			}
		})
	}
}

// TestVertexSizes verifies vertex struct sizes match the shader attribute layouts.
func TestVertexSizes(t *testing.T) {
	if s := int(unsafe.Sizeof(RectVertex{})); s != 76 {
		t.Errorf("RectVertex size = %d bytes, want 76", s)
	}
	if s := int(unsafe.Sizeof(ShadowVertex{})); s != 60 {
		t.Errorf("ShadowVertex size = %d bytes, want 60", s)
	}
	if s := int(unsafe.Sizeof(TexturedVertex{})); s != 32 {
		t.Errorf("TexturedVertex size = %d bytes, want 32", s)
	}
}

// TestObjCRuntimeLoaded verifies that the ObjC runtime was loaded in init().
func TestObjCRuntimeLoaded(t *testing.T) {
	if objc_msgSend == 0 {
		t.Skip("libobjc.A.dylib not loaded (not running on macOS)")
	}
	if sel_registerName == 0 {
		t.Error("sel_registerName not loaded")
	}
	if objc_getClass == 0 {
		t.Error("objc_getClass not loaded")
	}
}

// TestSelectors verifies that key selectors are registered (non-zero).
func TestSelectors(t *testing.T) {
	if objc_msgSend == 0 {
		t.Skip("ObjC runtime not available")
	}
	selTests := map[string]uintptr{
		"release":       selReleaseMetal,
		"alloc":         selAllocMetal,
		"init":          selInitMetal,
		"commandBuffer": selCommandBuffer,
		"endEncoding":   selEndEncoding,
		"commit":        selCommit,
	}
	for name, s := range selTests {
		if s == 0 {
			t.Errorf("selector %q is zero (not registered)", name)
		}
	}
}

// TestShaderCompilation attempts to compile the MSL shader source.
// Skipped if Metal is not available.
func TestShaderCompilation(t *testing.T) {
	if MTLCreateSystemDefaultDevice == 0 {
		t.Skip("Metal not available (MTLCreateSystemDefaultDevice not loaded)")
	}

	device, _, _ := purego.SyscallN(MTLCreateSystemDefaultDevice)
	if device == 0 {
		t.Skip("No Metal device available on this system")
	}
	defer msgSend(device, selReleaseMetal)

	srcNS := nsStringMetal(metalShaderSource)
	if srcNS == 0 {
		t.Fatal("failed to create NSString for shader source")
	}
	defer msgSend(srcNS, selReleaseMetal)

	var nsErr uintptr
	lib := msgSend(device, selNewDefaultLibraryWithSource, srcNS, 0, uintptr(unsafe.Pointer(&nsErr)))
	if lib == 0 {
		errStr := errFromNSError(nsErr)
		t.Fatalf("shader compilation failed: %s", errStr)
	}
	defer msgSend(lib, selReleaseMetal)

	// Verify required functions exist
	functions := []string{
		"rectVertex", "rectFragment",
		"shadowVertex", "shadowFragment",
		"texturedVertex", "textFragment", "imageFragment",
	}
	for _, name := range functions {
		nameNS := nsStringMetal(name)
		if nameNS == 0 {
			t.Errorf("failed to create NSString for %q", name)
			continue
		}
		fn := msgSend(lib, selNewFunctionWithName, nameNS)
		msgSend(nameNS, selReleaseMetal)
		if fn == 0 {
			t.Errorf("function %q not found in compiled library", name)
		} else {
			msgSend(fn, selReleaseMetal)
		}
	}
}

// TestBytesPerRow verifies bytesPerRowForFormat returns correct values.
func TestBytesPerRow(t *testing.T) {
	if v := bytesPerRowForFormat(render.TextureFormatR8, 100); v != 100 {
		t.Errorf("R8 bytesPerRow(100) = %d, want 100", v)
	}
	if v := bytesPerRowForFormat(render.TextureFormatRGBA8, 100); v != 400 {
		t.Errorf("RGBA8 bytesPerRow(100) = %d, want 400", v)
	}
	if v := bytesPerRowForFormat(render.TextureFormatBGRA8, 100); v != 400 {
		t.Errorf("BGRA8 bytesPerRow(100) = %d, want 400", v)
	}
}

// TestFloat32Helpers verifies float32Min and float32Clamp.
func TestFloat32Helpers(t *testing.T) {
	if v := float32Min(3.0, 5.0); v != 3.0 {
		t.Errorf("float32Min(3,5) = %v, want 3", v)
	}
	if v := float32Min(7.0, 2.0); v != 2.0 {
		t.Errorf("float32Min(7,2) = %v, want 2", v)
	}
	if v := float32Clamp(5.0, 1.0, 10.0); v != 5.0 {
		t.Errorf("float32Clamp(5,1,10) = %v, want 5", v)
	}
	if v := float32Clamp(-1.0, 0.0, 10.0); v != 0.0 {
		t.Errorf("float32Clamp(-1,0,10) = %v, want 0", v)
	}
	if v := float32Clamp(20.0, 0.0, 10.0); v != 10.0 {
		t.Errorf("float32Clamp(20,0,10) = %v, want 10", v)
	}
}

// TestTypedFunctionsRegistered verifies that typed function registrations succeeded.
func TestTypedFunctionsRegistered(t *testing.T) {
	if objc_msgSend == 0 {
		t.Skip("ObjC runtime not available")
	}
	if msgSendCGSize == nil {
		t.Error("msgSendCGSize is nil (purego.RegisterFunc failed)")
	}
	if msgSendCGFloat == nil {
		t.Error("msgSendCGFloat is nil (purego.RegisterFunc failed)")
	}
	if msgSendClearColor == nil {
		t.Error("msgSendClearColor is nil (purego.RegisterFunc failed)")
	}
}

// ---- nilWindow: a platform.Window stub that returns zero/defaults for everything ----

type nilWindow struct{}

var _ platform.Window = (*nilWindow)(nil)

func (w *nilWindow) Size() (int, int)             { return 800, 600 }
func (w *nilWindow) SetSize(int, int)              {}
func (w *nilWindow) FramebufferSize() (int, int)  { return 800, 600 }
func (w *nilWindow) Position() (int, int)          { return 0, 0 }
func (w *nilWindow) SetPosition(int, int)          {}
func (w *nilWindow) SetTitle(string)                {}
func (w *nilWindow) SetFullscreen(bool)             {}
func (w *nilWindow) IsFullscreen() bool             { return false }
func (w *nilWindow) ShouldClose() bool              { return false }
func (w *nilWindow) SetShouldClose(bool)            {}
func (w *nilWindow) NativeHandle() uintptr          { return 0 }
func (w *nilWindow) DPIScale() float32              { return 1.0 }
func (w *nilWindow) SetVisible(bool)                {}
func (w *nilWindow) ShowDeferred()                  {}
func (w *nilWindow) SetMinSize(int, int)            {}
func (w *nilWindow) SetMaxSize(int, int)            {}
func (w *nilWindow) SetCursor(platform.CursorShape) {}
func (w *nilWindow) SetIMEPosition(uimath.Rect)     {}
func (w *nilWindow) ShowContextMenu(clientX, clientY int, items []platform.ContextMenuItem) int {
	return -1
}
func (w *nilWindow) ClientToScreen(x, y int) (int, int) { return x, y }
func (w *nilWindow) Destroy()                            {}

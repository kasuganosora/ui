//go:build windows

package dx9

import (
	"testing"
	"unsafe"

	"github.com/kasuganosora/ui/render"
)

func TestNew(t *testing.T) {
	b := New()
	if b == nil {
		t.Fatal("New() returned nil")
	}
	if b.nextTextureID != 1 {
		t.Errorf("nextTextureID = %d, want 1", b.nextTextureID)
	}
	if b.textures == nil {
		t.Error("textures map is nil")
	}
	if b.dpiScale != 1.0 {
		t.Errorf("dpiScale = %f, want 1.0", b.dpiScale)
	}
}

func TestBackendInterface(t *testing.T) {
	var _ render.Backend = (*Backend)(nil)
}

func TestRectVertexSize(t *testing.T) {
	// 19 float32s = 76 bytes
	if got := unsafe.Sizeof(RectVertex{}); got != 76 {
		t.Errorf("RectVertex size = %d, want 76", got)
	}
}

func TestTexturedVertexSize(t *testing.T) {
	// 14 float32s = 56 bytes (pos + uv + color + rectSize + radius)
	if got := unsafe.Sizeof(TexturedVertex{}); got != 56 {
		t.Errorf("TexturedVertex size = %d, want 56", got)
	}
}

func TestMaxTextureSize(t *testing.T) {
	b := New()
	if got := b.MaxTextureSize(); got != 4096 {
		t.Errorf("MaxTextureSize = %d, want 4096", got)
	}
}

func TestDPIScale(t *testing.T) {
	b := New()
	b.dpiScale = 1.5
	if got := b.DPIScale(); got != 1.5 {
		t.Errorf("DPIScale = %f, want 1.5", got)
	}
}

func TestDestroyTextureNotFound(t *testing.T) {
	b := New()
	// Should not panic
	b.DestroyTexture(999)
}

func TestVertexElement9End(t *testing.T) {
	end := D3DVERTEXELEMENT9_END
	if end.Stream != 0xFF {
		t.Errorf("END stream = 0x%x, want 0xFF", end.Stream)
	}
	if end.Type != 17 {
		t.Errorf("END type = %d, want 17", end.Type)
	}
}

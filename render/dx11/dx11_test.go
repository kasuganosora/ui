//go:build windows

package dx11

import (
	"testing"
	"unsafe"
)

func TestBackendNew(t *testing.T) {
	b := New()
	if b == nil {
		t.Fatal("New returned nil")
	}
	if b.nextTextureID != 1 {
		t.Errorf("nextTextureID = %d, want 1", b.nextTextureID)
	}
	if b.dpiScale != 1.0 {
		t.Errorf("dpiScale = %f, want 1.0", b.dpiScale)
	}
}

func TestMaxTextureSize(t *testing.T) {
	b := New()
	if s := b.MaxTextureSize(); s != 16384 {
		t.Errorf("MaxTextureSize = %d, want 16384", s)
	}
}

func TestRectVertexSize(t *testing.T) {
	// RectVertex: 19 float32 fields = 76 bytes
	if s := unsafe.Sizeof(RectVertex{}); s != 76 {
		t.Errorf("RectVertex size = %d, want 76", s)
	}
}

func TestTexturedVertexSize(t *testing.T) {
	// TexturedVertex: 8 float32 fields = 32 bytes
	if s := unsafe.Sizeof(TexturedVertex{}); s != 32 {
		t.Errorf("TexturedVertex size = %d, want 32", s)
	}
}

func TestGUIDSize(t *testing.T) {
	if s := unsafe.Sizeof(GUID{}); s != 16 {
		t.Errorf("GUID size = %d, want 16", s)
	}
}

func TestDXGISwapChainDescSize(t *testing.T) {
	// Basic structural check — just verify it doesn't crash.
	var sd DXGI_SWAP_CHAIN_DESC
	_ = sd
}

func TestDestroyNilSafe(t *testing.T) {
	b := New()
	// Destroy without Init should not panic.
	b.Destroy()
}

func TestDPIScale(t *testing.T) {
	b := New()
	b.dpiScale = 1.5
	if s := b.DPIScale(); s != 1.5 {
		t.Errorf("DPIScale = %f, want 1.5", s)
	}
}

func TestSemanticName(t *testing.T) {
	// Reset cache for test isolation.
	semanticNames = map[string]*byte{}

	p := semanticName("POSITION")
	if p == nil {
		t.Fatal("semanticName returned nil")
	}
	// Same pointer on second call (cached).
	p2 := semanticName("POSITION")
	if p != p2 {
		t.Error("semanticName should cache pointers")
	}
}

func TestD3D11BoxSize(t *testing.T) {
	if s := unsafe.Sizeof(D3D11_BOX{}); s != 24 {
		t.Errorf("D3D11_BOX size = %d, want 24", s)
	}
}

func TestInputElementDescSize(t *testing.T) {
	// On 64-bit: *byte(8) + 5*uint32(20) + padding = 32 bytes
	s := unsafe.Sizeof(D3D11_INPUT_ELEMENT_DESC{})
	if s == 0 {
		t.Error("D3D11_INPUT_ELEMENT_DESC has zero size")
	}
}

func TestRectInputElements(t *testing.T) {
	// Reset cache.
	semanticNames = map[string]*byte{}

	elems := rectInputElements()
	if len(elems) != 7 {
		t.Fatalf("rectInputElements count = %d, want 7", len(elems))
	}
	// First element should be POSITION at offset 0.
	if elems[0].AlignedByteOffset != 0 {
		t.Errorf("first element offset = %d, want 0", elems[0].AlignedByteOffset)
	}
	// Last element (borderColor) at offset 60.
	if elems[6].AlignedByteOffset != 60 {
		t.Errorf("last element offset = %d, want 60", elems[6].AlignedByteOffset)
	}
}

func TestTexturedInputElements(t *testing.T) {
	// Reset cache.
	semanticNames = map[string]*byte{}

	elems := texturedInputElements()
	if len(elems) != 3 {
		t.Fatalf("texturedInputElements count = %d, want 3", len(elems))
	}
	// COLOR at offset 16.
	if elems[2].AlignedByteOffset != 16 {
		t.Errorf("color element offset = %d, want 16", elems[2].AlignedByteOffset)
	}
}

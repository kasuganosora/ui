package atlas

import (
	"image"
	"testing"

	"github.com/kasuganosora/ui/font"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/render"
)

// mockBackend satisfies render.Backend for testing Upload/Destroy with a backend.
type mockBackend struct {
	textures    map[render.TextureHandle]bool
	nextTexture render.TextureHandle
	updated     bool
}

func newMockBackend() *mockBackend {
	return &mockBackend{textures: make(map[render.TextureHandle]bool), nextTexture: 1}
}

func (b *mockBackend) Init(window platform.Window) error  { return nil }
func (b *mockBackend) BeginFrame()                        {}
func (b *mockBackend) EndFrame()                          {}
func (b *mockBackend) Submit(buf *render.CommandBuffer)    {}
func (b *mockBackend) Resize(w, h int)                    {}
func (b *mockBackend) MaxTextureSize() int                { return 4096 }
func (b *mockBackend) ReadPixels() (*image.RGBA, error)   { return nil, nil }
func (b *mockBackend) Destroy()                           {}
func (b *mockBackend) CreateTexture(desc render.TextureDesc) (render.TextureHandle, error) {
	h := b.nextTexture
	b.nextTexture++
	b.textures[h] = true
	return h, nil
}
func (b *mockBackend) UpdateTexture(h render.TextureHandle, region uimath.Rect, data []byte) error {
	b.updated = true
	return nil
}
func (b *mockBackend) DestroyTexture(h render.TextureHandle) {
	delete(b.textures, h)
}

func TestUimathRect(t *testing.T) {
	r := uimathRect(10, 20, 30, 40)
	if r.X != 10 || r.Y != 20 || r.Width != 30 || r.Height != 40 {
		t.Errorf("unexpected rect: %v", r)
	}
}

func TestAtlasUploadWithBackend(t *testing.T) {
	backend := newMockBackend()
	a := New(Options{Width: 128, Height: 128, Backend: backend})
	defer a.Destroy()

	bitmap := font.GlyphBitmap{Width: 10, Height: 10, Data: make([]byte, 100)}
	a.Insert(MakeKey(1, 1, 16), bitmap, font.GlyphMetrics{Advance: 10})

	// First upload creates texture
	err := a.Upload()
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}
	if a.Texture() == render.InvalidTexture {
		t.Error("texture should be created after upload")
	}

	// Insert another glyph and upload partial update
	bitmap2 := font.GlyphBitmap{Width: 8, Height: 8, Data: make([]byte, 64)}
	a.Insert(MakeKey(1, 2, 16), bitmap2, font.GlyphMetrics{Advance: 8})

	err = a.Upload()
	if err != nil {
		t.Fatalf("partial Upload failed: %v", err)
	}
	if !backend.updated {
		t.Error("backend should have been updated")
	}
}

func TestAtlasDestroyWithBackend(t *testing.T) {
	backend := newMockBackend()
	a := New(Options{Width: 64, Height: 64, Backend: backend})

	// Insert and upload to create texture
	bitmap := font.GlyphBitmap{Width: 5, Height: 5, Data: make([]byte, 25)}
	a.Insert(MakeKey(1, 1, 16), bitmap, font.GlyphMetrics{})
	a.Upload()

	h := a.Texture()
	if h == render.InvalidTexture {
		t.Fatal("expected valid texture")
	}

	a.Destroy()
	if a.Texture() != render.InvalidTexture {
		t.Error("texture should be invalid after destroy")
	}
	if backend.textures[h] {
		t.Error("backend texture should be destroyed")
	}
}

func TestAtlasUploadNoChangeAfterReset(t *testing.T) {
	backend := newMockBackend()
	a := New(Options{Width: 64, Height: 64, Backend: backend})
	defer a.Destroy()

	// Upload with no inserts
	err := a.Upload()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAtlasInsertDuplicate(t *testing.T) {
	a := New(Options{Width: 128, Height: 128})
	bitmap := font.GlyphBitmap{Width: 5, Height: 5, Data: make([]byte, 25)}
	key := MakeKey(1, 1, 16)

	e1 := a.Insert(key, bitmap, font.GlyphMetrics{Advance: 5})
	e2 := a.Insert(key, bitmap, font.GlyphMetrics{Advance: 5})

	if e1 == nil || e2 == nil {
		t.Fatal("both inserts should return entries")
	}
	if a.GlyphCount() != 1 {
		t.Errorf("duplicate insert should not increase count, got %d", a.GlyphCount())
	}
}

func TestAtlasDefaultSize(t *testing.T) {
	a := New(Options{}) // zero width/height -> defaults
	if a.Width() != 1024 || a.Height() != 1024 {
		t.Errorf("expected 1024x1024 default, got %dx%d", a.Width(), a.Height())
	}
}

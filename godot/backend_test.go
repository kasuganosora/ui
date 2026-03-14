package godot

import (
	"image/color"
	"testing"

	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

func TestSoftwareBackend_Init(t *testing.T) {
	win := NewHeadlessWindow(800, 600, 1.0)
	b := NewSoftwareBackend()
	if err := b.Init(win); err != nil {
		t.Fatal(err)
	}
	defer b.Destroy()

	w, h := b.FramebufferSize()
	if w != 800 || h != 600 {
		t.Errorf("framebuffer size: got %dx%d, want 800x600", w, h)
	}
	if len(b.Pixels()) != 800*600*4 {
		t.Errorf("pixel buffer length: got %d, want %d", len(b.Pixels()), 800*600*4)
	}
}

func TestSoftwareBackend_InitDPI(t *testing.T) {
	win := NewHeadlessWindow(800, 600, 2.0)
	b := NewSoftwareBackend()
	if err := b.Init(win); err != nil {
		t.Fatal(err)
	}
	defer b.Destroy()

	w, h := b.FramebufferSize()
	if w != 1600 || h != 1200 {
		t.Errorf("framebuffer size at 2x DPI: got %dx%d, want 1600x1200", w, h)
	}
	if b.DPIScale() != 2.0 {
		t.Errorf("DPI scale: got %f, want 2.0", b.DPIScale())
	}
}

func TestSoftwareBackend_ClearOnBeginFrame(t *testing.T) {
	win := NewHeadlessWindow(4, 4, 1.0)
	b := NewSoftwareBackend()
	b.Init(win)
	defer b.Destroy()

	// Dirty the buffer
	for i := range b.pixels {
		b.pixels[i] = 0xFF
	}

	b.BeginFrame()

	// Should be cleared to transparent black
	for i, v := range b.pixels {
		if v != 0 {
			t.Fatalf("pixel[%d] = %d after BeginFrame, want 0", i, v)
		}
	}
}

func TestSoftwareBackend_DrawRect(t *testing.T) {
	win := NewHeadlessWindow(100, 100, 1.0)
	b := NewSoftwareBackend()
	b.Init(win)
	defer b.Destroy()

	buf := render.NewCommandBuffer()
	b.BeginFrame()
	buf.Reset()
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(10, 10, 20, 20),
		FillColor: uimath.NewColor(1, 0, 0, 1), // red
	}, 0, 1.0)
	b.Submit(buf)

	// Check pixel inside the rect (15, 15)
	c := b.pixelAt(15, 15)
	if c.R != 255 || c.G != 0 || c.B != 0 || c.A != 255 {
		t.Errorf("pixel at (15,15): got %v, want red", c)
	}

	// Check pixel outside (5, 5)
	c = b.pixelAt(5, 5)
	if c.A != 0 {
		t.Errorf("pixel at (5,5): got alpha %d, want 0", c.A)
	}
}

func TestSoftwareBackend_DrawRectWithBorder(t *testing.T) {
	win := NewHeadlessWindow(100, 100, 1.0)
	b := NewSoftwareBackend()
	b.Init(win)
	defer b.Destroy()

	buf := render.NewCommandBuffer()
	b.BeginFrame()
	buf.Reset()
	buf.DrawRect(render.RectCmd{
		Bounds:      uimath.NewRect(10, 10, 30, 30),
		FillColor:   uimath.NewColor(0, 0, 1, 1), // blue fill
		BorderColor: uimath.NewColor(1, 1, 0, 1), // yellow border
		BorderWidth: 2,
	}, 0, 1.0)
	b.Submit(buf)

	// Border pixel (top edge, at y=10, x=20)
	c := b.pixelAt(20, 10)
	if c.R != 255 || c.G != 255 || c.B != 0 {
		t.Errorf("border pixel at (20,10): got %v, want yellow", c)
	}

	// Interior pixel
	c = b.pixelAt(20, 20)
	// Border is drawn on top of fill, interior should be blue
	if c.B < 200 {
		t.Errorf("interior pixel at (20,20): expected blue, got %v", c)
	}
}

func TestSoftwareBackend_AlphaBlending(t *testing.T) {
	win := NewHeadlessWindow(100, 100, 1.0)
	b := NewSoftwareBackend()
	b.Init(win)
	defer b.Destroy()

	buf := render.NewCommandBuffer()
	b.BeginFrame()
	buf.Reset()

	// Draw blue background
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(0, 0, 100, 100),
		FillColor: uimath.NewColor(0, 0, 1, 1),
	}, 0, 1.0)
	// Draw 50% transparent red on top
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(0, 0, 100, 100),
		FillColor: uimath.NewColor(1, 0, 0, 0.5),
	}, 1, 1.0)
	b.Submit(buf)

	c := b.pixelAt(50, 50)
	// Blended: R should be ~128, B should be ~127
	if c.R < 100 || c.R > 160 {
		t.Errorf("blended R: got %d, want ~128", c.R)
	}
	if c.B < 80 || c.B > 160 {
		t.Errorf("blended B: got %d, want ~127", c.B)
	}
}

func TestSoftwareBackend_Scissor(t *testing.T) {
	win := NewHeadlessWindow(100, 100, 1.0)
	b := NewSoftwareBackend()
	b.Init(win)
	defer b.Destroy()

	buf := render.NewCommandBuffer()
	b.BeginFrame()
	buf.Reset()

	// Set scissor to top-left quadrant
	buf.PushClip(uimath.NewRect(0, 0, 50, 50))
	// Draw full-screen rect
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(0, 0, 100, 100),
		FillColor: uimath.NewColor(0, 1, 0, 1), // green
	}, 0, 1.0)
	buf.PopClip()
	b.Submit(buf)

	// Inside scissor
	c := b.pixelAt(25, 25)
	if c.G != 255 {
		t.Errorf("inside scissor (25,25): got G=%d, want 255", c.G)
	}

	// Outside scissor
	c = b.pixelAt(75, 75)
	if c.A != 0 {
		t.Errorf("outside scissor (75,75): got alpha %d, want 0", c.A)
	}
}

func TestSoftwareBackend_TextureCreateUpdate(t *testing.T) {
	win := NewHeadlessWindow(100, 100, 1.0)
	b := NewSoftwareBackend()
	b.Init(win)
	defer b.Destroy()

	// Create 2x2 RGBA texture
	data := []byte{
		255, 0, 0, 255, 0, 255, 0, 255,
		0, 0, 255, 255, 255, 255, 0, 255,
	}
	h, err := b.CreateTexture(render.TextureDesc{
		Width:  2,
		Height: 2,
		Format: render.TextureFormatRGBA8,
		Data:   data,
	})
	if err != nil {
		t.Fatal(err)
	}
	if h == render.InvalidTexture {
		t.Fatal("got invalid texture handle")
	}

	// Update top-left pixel to white
	err = b.UpdateTexture(h, uimath.NewRect(0, 0, 1, 1), []byte{255, 255, 255, 255})
	if err != nil {
		t.Fatal(err)
	}

	b.DestroyTexture(h)
}

func TestSoftwareBackend_R8Texture(t *testing.T) {
	win := NewHeadlessWindow(100, 100, 1.0)
	b := NewSoftwareBackend()
	b.Init(win)
	defer b.Destroy()

	// Create 2x2 R8 texture (glyph atlas)
	data := []byte{128, 255, 0, 64}
	h, err := b.CreateTexture(render.TextureDesc{
		Width:  2,
		Height: 2,
		Format: render.TextureFormatR8,
		Data:   data,
	})
	if err != nil {
		t.Fatal(err)
	}

	tex := b.textures[h]
	if tex == nil {
		t.Fatal("texture not found")
	}

	// R8 → RGBA: value becomes alpha, RGB is white
	// Pixel (0,0): alpha = 128
	if tex.data[3] != 128 {
		t.Errorf("R8 pixel (0,0) alpha: got %d, want 128", tex.data[3])
	}
	// Pixel (1,0): alpha = 255
	if tex.data[7] != 255 {
		t.Errorf("R8 pixel (1,0) alpha: got %d, want 255", tex.data[7])
	}
}

func TestSoftwareBackend_ReadPixels(t *testing.T) {
	win := NewHeadlessWindow(10, 10, 1.0)
	b := NewSoftwareBackend()
	b.Init(win)
	defer b.Destroy()

	buf := render.NewCommandBuffer()
	b.BeginFrame()
	buf.Reset()
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(0, 0, 10, 10),
		FillColor: uimath.NewColor(1, 0, 1, 1), // magenta
	}, 0, 1.0)
	b.Submit(buf)

	img, err := b.ReadPixels()
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() != 10 || img.Bounds().Dy() != 10 {
		t.Errorf("image size: %v, want 10x10", img.Bounds())
	}
	c := img.At(5, 5).(color.RGBA)
	if c.R != 255 || c.B != 255 {
		t.Errorf("ReadPixels pixel: got %v, want magenta", c)
	}
}

func BenchmarkSoftwareBackend_FillRect(b *testing.B) {
	win := NewHeadlessWindow(1920, 1080, 1.0)
	be := NewSoftwareBackend()
	be.Init(win)
	defer be.Destroy()

	buf := render.NewCommandBuffer()
	for i := 0; i < b.N; i++ {
		be.BeginFrame()
		buf.Reset()
		// Simulate ~20 UI rectangles
		for j := 0; j < 20; j++ {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(float32(j*50), float32(j*30), 200, 100),
				FillColor: uimath.NewColor(0.2, 0.4, 0.8, 0.9),
			}, int32(j), 1.0)
		}
		be.Submit(buf)
	}
}

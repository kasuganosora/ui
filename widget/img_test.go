package widget

import (
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

func TestImgNew(t *testing.T) {
	tree := core.NewTree()
	img := NewImg(tree, nil)
	if img == nil {
		t.Fatal("NewImg returned nil")
	}
	if img.State() != ImgStateIdle {
		t.Errorf("expected idle state, got %d", img.State())
	}
	if img.Src() != "" {
		t.Error("expected empty src")
	}
	if img.texture != render.InvalidTexture {
		t.Error("expected invalid texture")
	}
}

func TestImgSetAlt(t *testing.T) {
	tree := core.NewTree()
	img := NewImg(tree, nil)
	img.SetAlt("test image")
	if img.Alt() != "test image" {
		t.Errorf("expected 'test image', got %q", img.Alt())
	}
}

func TestImgObjectFit(t *testing.T) {
	tree := core.NewTree()
	img := NewImg(tree, nil)
	img.SetObjectFit(ObjectFitContain)
	if img.ObjectFit() != ObjectFitContain {
		t.Errorf("expected contain, got %d", img.ObjectFit())
	}
}

func TestImgSetSrcNotFound(t *testing.T) {
	tree := core.NewTree()
	img := NewImg(tree, nil)
	img.SetSrc("/nonexistent/path.png")
	if img.State() != ImgStateError {
		t.Errorf("expected error state, got %d", img.State())
	}
	if img.Error() == "" {
		t.Error("expected error message")
	}
}

func TestImgSetSrcEmpty(t *testing.T) {
	tree := core.NewTree()
	img := NewImg(tree, nil)
	img.SetSrc("")
	if img.State() != ImgStateIdle {
		t.Errorf("expected idle state, got %d", img.State())
	}
}

func TestImgLoadPNG(t *testing.T) {
	// Create a temp PNG file
	dir := t.TempDir()
	path := filepath.Join(dir, "test.png")

	rgba := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			rgba.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	png.Encode(f, rgba)
	f.Close()

	tree := core.NewTree()
	img := NewImg(tree, nil)
	img.SetSrc(path)

	// Without backend, state should be loaded but no texture
	if img.State() != ImgStateLoaded {
		t.Errorf("expected loaded state, got %d (err: %s)", img.State(), img.Error())
	}
	if img.NaturalWidth() != 32 || img.NaturalHeight() != 32 {
		t.Errorf("expected 32x32, got %dx%d", img.NaturalWidth(), img.NaturalHeight())
	}
	if img.IsAnimated() {
		t.Error("PNG should not be animated")
	}
}

func TestImgLoadGIF(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.gif")

	// Create a 2-frame animated GIF
	g := &gif.GIF{
		Image: []*image.Paletted{
			image.NewPaletted(image.Rect(0, 0, 16, 16), color.Palette{color.Black, color.White}),
			image.NewPaletted(image.Rect(0, 0, 16, 16), color.Palette{color.Black, color.White}),
		},
		Delay:     []int{10, 10}, // 100ms each
		LoopCount: 0,             // infinite
		Config:    image.Config{Width: 16, Height: 16},
	}
	// Fill frames with different colors
	for x := 0; x < 16; x++ {
		for y := 0; y < 16; y++ {
			g.Image[0].SetColorIndex(x, y, 0) // black
			g.Image[1].SetColorIndex(x, y, 1) // white
		}
	}

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	gif.EncodeAll(f, g)
	f.Close()

	tree := core.NewTree()
	img := NewImg(tree, nil)
	img.SetSrc(path)

	if img.State() != ImgStateLoaded {
		t.Errorf("expected loaded state, got %d (err: %s)", img.State(), img.Error())
	}
	if !img.IsAnimated() {
		t.Error("expected animated GIF")
	}
	if img.NaturalWidth() != 16 || img.NaturalHeight() != 16 {
		t.Errorf("expected 16x16, got %dx%d", img.NaturalWidth(), img.NaturalHeight())
	}
	if len(img.frames) != 2 {
		t.Errorf("expected 2 frames, got %d", len(img.frames))
	}
	// Without backend, playing should be true but texture is invalid
	if !img.Playing() {
		t.Error("expected auto-play for animated GIF")
	}
}

func TestImgFitContain(t *testing.T) {
	tree := core.NewTree()
	img := NewImg(tree, nil)
	img.naturalW = 200
	img.naturalH = 100
	img.SetObjectFit(ObjectFitContain)

	bounds := uimath.NewRect(0, 0, 100, 100)
	dst := img.fitRect(bounds)

	// 200x100 into 100x100 → 100x50, centered vertically
	if dst.Width != 100 || dst.Height != 50 {
		t.Errorf("expected 100x50, got %vx%v", dst.Width, dst.Height)
	}
	if dst.Y != 25 {
		t.Errorf("expected Y=25 (centered), got %v", dst.Y)
	}
}

func TestImgFitCover(t *testing.T) {
	tree := core.NewTree()
	img := NewImg(tree, nil)
	img.naturalW = 200
	img.naturalH = 100
	img.SetObjectFit(ObjectFitCover)

	bounds := uimath.NewRect(0, 0, 100, 100)
	dst := img.fitRect(bounds)

	// 200x100 into 100x100 → 200x100 scaled to cover → 100 height → 200x100 → w=200, h=100
	// Cover: use dimension that's smaller ratio. bounds aspect = 1, image aspect = 2
	// aspect < bAspect is false (2 < 1 is false), so h = 100, w = 200
	if dst.Width != 200 || dst.Height != 100 {
		t.Errorf("expected 200x100, got %vx%v", dst.Width, dst.Height)
	}
}

func TestImgFitNone(t *testing.T) {
	tree := core.NewTree()
	img := NewImg(tree, nil)
	img.naturalW = 50
	img.naturalH = 30
	img.SetObjectFit(ObjectFitNone)

	bounds := uimath.NewRect(10, 10, 100, 100)
	dst := img.fitRect(bounds)

	if dst.Width != 50 || dst.Height != 30 {
		t.Errorf("expected 50x30, got %vx%v", dst.Width, dst.Height)
	}
	// Should be centered
	if dst.X != 35 || dst.Y != 45 {
		t.Errorf("expected centered at 35,45, got %v,%v", dst.X, dst.Y)
	}
}

func TestImgFitScaleDown(t *testing.T) {
	tree := core.NewTree()
	img := NewImg(tree, nil)

	// Small image that fits without scaling
	img.naturalW = 50
	img.naturalH = 30
	img.SetObjectFit(ObjectFitScaleDown)

	bounds := uimath.NewRect(0, 0, 100, 100)
	dst := img.fitRect(bounds)

	// Should show at natural size since it fits
	if dst.Width != 50 || dst.Height != 30 {
		t.Errorf("expected 50x30, got %vx%v", dst.Width, dst.Height)
	}

	// Large image that needs scaling
	img.naturalW = 200
	img.naturalH = 100
	dst = img.fitRect(bounds)

	// Should scale down like contain
	if dst.Width != 100 || dst.Height != 50 {
		t.Errorf("expected 100x50, got %vx%v", dst.Width, dst.Height)
	}
}

func TestImgPlayPause(t *testing.T) {
	tree := core.NewTree()
	img := NewImg(tree, nil)
	img.frames = make([]gifFrame, 3)
	img.playing = true

	img.Pause()
	if img.Playing() {
		t.Error("expected paused")
	}

	img.Play()
	if !img.Playing() {
		t.Error("expected playing")
	}
}

func TestImgSetFrame(t *testing.T) {
	tree := core.NewTree()
	img := NewImg(tree, nil)
	img.frames = make([]gifFrame, 3)

	img.SetFrame(2)
	if img.frameIdx != 2 {
		t.Errorf("expected frame 2, got %d", img.frameIdx)
	}

	// Out of bounds should be ignored
	img.SetFrame(5)
	if img.frameIdx != 2 {
		t.Errorf("expected frame 2 (unchanged), got %d", img.frameIdx)
	}
}

func TestImgDraw(t *testing.T) {
	tree := core.NewTree()
	img := NewImg(tree, nil)
	buf := render.NewCommandBuffer()

	// Draw with no image should not panic
	tree.SetLayout(img.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 100, 100),
	})
	img.Draw(buf)
}

func TestImgOnLoadCallback(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.png")

	rgba := image.NewRGBA(image.Rect(0, 0, 4, 4))
	f, _ := os.Create(path)
	png.Encode(f, rgba)
	f.Close()

	tree := core.NewTree()
	img := NewImg(tree, nil)

	called := false
	img.OnLoad(func() { called = true })
	img.SetSrc(path)

	if !called {
		t.Error("onLoad callback not called")
	}
}

func TestImgOnErrorCallback(t *testing.T) {
	tree := core.NewTree()
	img := NewImg(tree, nil)

	var gotErr error
	img.OnError(func(err error) { gotErr = err })
	img.SetSrc("/nonexistent.png")

	if gotErr == nil {
		t.Error("onError callback not called")
	}
}

func TestIsGIF(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"test.gif", true},
		{"test.GIF", true},
		{"test.Gif", true},
		{"test.png", false},
		{"test.jpg", false},
		{"giffile.png", false},
	}
	for _, tt := range tests {
		if got := isGIF(tt.path); got != tt.want {
			t.Errorf("isGIF(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

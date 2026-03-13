package widget

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"strings"
	"time"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// ImgState represents the loading state of an Img widget.
type ImgState uint8

const (
	ImgStateIdle    ImgState = iota // No source set
	ImgStateLoading                 // Loading in progress
	ImgStateLoaded                  // Successfully loaded
	ImgStateError                   // Load failed
)

// ObjectFit controls how the image fits within its bounds (CSS object-fit).
type ObjectFit uint8

const (
	ObjectFitFill      ObjectFit = iota // Stretch to fill (default)
	ObjectFitContain                    // Scale to fit, preserving aspect ratio
	ObjectFitCover                      // Scale to cover, preserving aspect ratio
	ObjectFitNone                       // No scaling, original size
	ObjectFitScaleDown                  // Like contain, but never upscale
)

// gifFrame holds a pre-composited frame for animated GIF playback.
type gifFrame struct {
	pixels []byte       // RGBA pixel data
	delay  time.Duration // display duration
}

// Img loads and displays images from file paths.
// Supports PNG, JPEG, and animated GIF with frame playback.
// Follows the W3C <img> element specification.
type Img struct {
	Base
	src       string
	alt       string
	objectFit ObjectFit
	state     ImgState
	errMsg    string

	// Image data
	naturalW int // original image width
	naturalH int // original image height
	texture  render.TextureHandle

	// Pending texture data from background goroutine (decoded on bg thread,
	// GPU upload deferred to main thread via Draw).
	pendingPixels []byte
	pendingW      int
	pendingH      int

	// Animated GIF
	frames     []gifFrame
	frameIdx   int
	loopCount  int // 0 = infinite
	loopsDone  int
	lastUpdate time.Time
	playing    bool

	// Callbacks
	onLoad  func()
	onError func(error)
}

// NewImg creates an image widget.
func NewImg(tree *core.Tree, cfg *Config) *Img {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	img := &Img{
		Base:    NewBase(tree, core.TypeImage, cfg),
		texture: render.InvalidTexture,
	}
	img.style.Display = layout.DisplayBlock
	return img
}

// Getters
func (img *Img) Src() string       { return img.src }
func (img *Img) Alt() string       { return img.alt }
func (img *Img) State() ImgState   { return img.state }
func (img *Img) Error() string     { return img.errMsg }
func (img *Img) ObjectFit() ObjectFit { return img.objectFit }
func (img *Img) NaturalWidth() int  { return img.naturalW }
func (img *Img) NaturalHeight() int { return img.naturalH }
func (img *Img) IsAnimated() bool  { return len(img.frames) > 1 }
func (img *Img) Playing() bool     { return img.playing }

// Setters
func (img *Img) SetAlt(alt string)         { img.alt = alt }
func (img *Img) SetObjectFit(f ObjectFit)  { img.objectFit = f }
func (img *Img) OnLoad(fn func())          { img.onLoad = fn }
func (img *Img) OnError(fn func(error))    { img.onError = fn }

// SetSrc sets the image source path and triggers loading.
func (img *Img) SetSrc(src string) {
	if src == img.src && img.state == ImgStateLoaded {
		return
	}
	img.src = src
	img.destroyTexture()
	img.frames = nil
	img.frameIdx = 0
	img.loopsDone = 0
	img.playing = false

	if src == "" {
		img.state = ImgStateIdle
		return
	}

	img.state = ImgStateLoading
	if isRemoteURL(src) {
		img.loadFromURL(src)
	} else {
		img.load()
	}
}

// Play starts animated GIF playback.
func (img *Img) Play() {
	if len(img.frames) > 1 {
		img.playing = true
		img.lastUpdate = time.Now()
		img.tree.MarkDirty(img.id)
	}
}

// Pause stops animated GIF playback.
func (img *Img) Pause() {
	img.playing = false
}

// SetFrame sets the current frame index for animated GIFs.
func (img *Img) SetFrame(idx int) {
	if idx >= 0 && idx < len(img.frames) {
		img.frameIdx = idx
		img.uploadFrame(idx)
	}
}

func (img *Img) load() {
	f, err := os.Open(img.src)
	if err != nil {
		img.setError(err)
		return
	}
	defer f.Close()

	// Check if GIF (try animated decode first)
	if isGIF(img.src) {
		img.loadGIF(f)
		return
	}

	// Decode PNG/JPEG/other
	decoded, _, err := image.Decode(f)
	if err != nil {
		img.setError(err)
		return
	}

	// Convert to RGBA
	bounds := decoded.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, decoded, bounds.Min, draw.Src)

	img.naturalW = bounds.Dx()
	img.naturalH = bounds.Dy()
	// Defer GPU texture creation to the main thread (Draw).
	// Vulkan command submission is not thread-safe.
	img.pendingPixels = rgba.Pix
	img.pendingW = img.naturalW
	img.pendingH = img.naturalH
	img.state = ImgStateLoading
}

func (img *Img) loadGIF(f *os.File) {
	img.loadGIFFromReader(f)
}

// loadGIFFromReader is like loadGIF but reads from an io.Reader.
func (img *Img) loadGIFFromReader(r io.Reader) {
	g, err := gif.DecodeAll(r)
	if err != nil {
		img.setError(err)
		return
	}

	if len(g.Image) == 0 {
		img.setError(errNoFrames)
		return
	}

	// Canvas size from config or first frame
	canvasW := g.Config.Width
	canvasH := g.Config.Height
	if canvasW == 0 || canvasH == 0 {
		canvasW = g.Image[0].Bounds().Dx()
		canvasH = g.Image[0].Bounds().Dy()
	}

	img.naturalW = canvasW
	img.naturalH = canvasH
	img.loopCount = g.LoopCount

	// Composite all frames
	canvas := image.NewRGBA(image.Rect(0, 0, canvasW, canvasH))
	var prevCanvas *image.RGBA

	img.frames = make([]gifFrame, len(g.Image))
	for i, frame := range g.Image {
		disposal := byte(gif.DisposalNone)
		if i < len(g.Disposal) {
			disposal = g.Disposal[i]
		}

		// Save canvas before drawing for DisposalPrevious
		if disposal == gif.DisposalPrevious {
			prevCanvas = cloneRGBA(canvas)
		}

		// Draw frame onto canvas
		draw.Draw(canvas, frame.Bounds(), frame, frame.Bounds().Min, draw.Over)

		// Store composited frame
		pixels := make([]byte, len(canvas.Pix))
		copy(pixels, canvas.Pix)

		delay := time.Duration(100) * time.Millisecond // default 100ms
		if i < len(g.Delay) && g.Delay[i] > 0 {
			delay = time.Duration(g.Delay[i]) * 10 * time.Millisecond
		}

		img.frames[i] = gifFrame{
			pixels: pixels,
			delay:  delay,
		}

		// Apply disposal
		switch disposal {
		case gif.DisposalBackground:
			// Clear the frame area to transparent
			bg := image.NewUniform(color.Transparent)
			draw.Draw(canvas, frame.Bounds(), bg, image.Point{}, draw.Src)
		case gif.DisposalPrevious:
			if prevCanvas != nil {
				copy(canvas.Pix, prevCanvas.Pix)
			}
		}
	}

	// Defer GPU texture creation to the main thread (Draw).
	// Vulkan command submission is not thread-safe.
	img.pendingPixels = img.frames[0].pixels
	img.pendingW = canvasW
	img.pendingH = canvasH
	img.state = ImgStateLoading

	// Auto-play animated GIFs
	if len(img.frames) > 1 {
		img.playing = true
		img.lastUpdate = time.Now()
	}
}

// loadFromURL fetches a remote URL or decodes a data: URL asynchronously.
func (img *Img) loadFromURL(src string) {
	nc := img.config.NetClient
	if nc == nil {
		img.setError(fmt.Errorf("net: no NetClient configured on Config"))
		return
	}
	nc.FetchAsync(src, func(data []byte, err error) {
		if err != nil {
			img.errMsg = err.Error()
			img.state = ImgStateError
			if img.onError != nil {
				img.onError(err)
			}
			img.tree.MarkDirty(img.id)
			return
		}
		img.loadFromBytes(data, src)
		img.tree.MarkDirty(img.id)
	})
}

// loadFromBytes decodes image data from an in-memory byte slice.
func (img *Img) loadFromBytes(data []byte, hint string) {
	if isGIF(hint) || looksLikeGIF(data) {
		img.loadGIFFromReader(bytes.NewReader(data))
		return
	}
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		img.setError(err)
		return
	}
	bounds := decoded.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, decoded, bounds.Min, draw.Src)
	img.naturalW = bounds.Dx()
	img.naturalH = bounds.Dy()
	// Defer GPU texture creation to the main thread (Draw).
	// Vulkan command submission is not thread-safe.
	img.pendingPixels = rgba.Pix
	img.pendingW = img.naturalW
	img.pendingH = img.naturalH
	img.state = ImgStateLoading
}

func isRemoteURL(src string) bool {
	return len(src) > 7 && (src[:7] == "http://" || len(src) > 8 && src[:8] == "https://" || len(src) > 5 && src[:5] == "data:")
}

func looksLikeGIF(data []byte) bool {
	return len(data) >= 6 && (string(data[:6]) == "GIF89a" || string(data[:6]) == "GIF87a")
}

func (img *Img) createTexture(pixels []byte, w, h int) {
	backend := img.config.Backend
	if backend == nil {
		img.state = ImgStateLoaded
		if img.onLoad != nil {
			img.onLoad()
		}
		return
	}

	tex, err := backend.CreateTexture(render.TextureDesc{
		Width:  w,
		Height: h,
		Format: render.TextureFormatRGBA8,
		Filter: render.TextureFilterLinear,
		Data:   pixels,
	})
	if err != nil {
		img.setError(err)
		return
	}

	img.texture = tex
	img.state = ImgStateLoaded
	if img.onLoad != nil {
		img.onLoad()
	}
}

func (img *Img) uploadFrame(idx int) {
	if idx < 0 || idx >= len(img.frames) || img.texture == render.InvalidTexture {
		return
	}
	backend := img.config.Backend
	if backend == nil {
		return
	}
	region := uimath.NewRect(0, 0, float32(img.naturalW), float32(img.naturalH))
	backend.UpdateTexture(img.texture, region, img.frames[idx].pixels)
}

func (img *Img) setError(err error) {
	img.state = ImgStateError
	if err != nil {
		img.errMsg = err.Error()
	}
	if img.onError != nil {
		img.onError(err)
	}
}

func (img *Img) destroyTexture() {
	if img.texture != render.InvalidTexture && img.config.Backend != nil {
		img.config.Backend.DestroyTexture(img.texture)
		img.texture = render.InvalidTexture
	}
}

// Destroy cleans up GPU resources.
func (img *Img) Destroy() {
	img.destroyTexture()
	img.Base.Destroy()
}

// Draw renders the image.
func (img *Img) Draw(buf *render.CommandBuffer) {
	// Flush pending texture from background goroutine (GPU ops must run on main thread)
	if img.pendingPixels != nil {
		img.createTexture(img.pendingPixels, img.pendingW, img.pendingH)
		img.pendingPixels = nil
	}

	bounds := img.Bounds()
	if bounds.IsEmpty() {
		return
	}

	// Advance GIF animation
	if img.playing && len(img.frames) > 1 {
		now := time.Now()
		elapsed := now.Sub(img.lastUpdate)
		frame := img.frames[img.frameIdx]
		if elapsed >= frame.delay {
			img.lastUpdate = now
			nextIdx := img.frameIdx + 1
			if nextIdx >= len(img.frames) {
				// Loop handling
				if img.loopCount == 0 {
					// Infinite loop
					nextIdx = 0
				} else {
					img.loopsDone++
					if img.loopsDone >= img.loopCount {
						img.playing = false
						nextIdx = img.frameIdx // stay on last frame
					} else {
						nextIdx = 0
					}
				}
			}
			if nextIdx != img.frameIdx {
				img.frameIdx = nextIdx
				img.uploadFrame(nextIdx)
			}
		}
		if img.playing {
			img.tree.MarkDirty(img.id)
		}
	}

	if img.texture == render.InvalidTexture {
		// Show alt text or error placeholder
		img.drawFallback(buf, bounds)
		return
	}

	// Calculate destination rect based on object-fit
	dst := img.fitRect(bounds)

	buf.DrawImage(render.ImageCmd{
		Texture: img.texture,
		SrcRect: uimath.NewRect(0, 0, 1, 1),
		DstRect: dst,
		Tint:    uimath.ColorWhite,
	}, 0, 1)

	img.DrawChildren(buf)
}

func (img *Img) fitRect(bounds uimath.Rect) uimath.Rect {
	if img.naturalW == 0 || img.naturalH == 0 {
		return bounds
	}

	natW := float32(img.naturalW)
	natH := float32(img.naturalH)
	aspect := natW / natH

	switch img.objectFit {
	case ObjectFitContain:
		return fitContain(bounds, aspect)
	case ObjectFitCover:
		return fitCover(bounds, aspect)
	case ObjectFitNone:
		// Center at natural size
		x := bounds.X + (bounds.Width-natW)/2
		y := bounds.Y + (bounds.Height-natH)/2
		return uimath.NewRect(x, y, natW, natH)
	case ObjectFitScaleDown:
		// Like contain, but never upscale
		if natW <= bounds.Width && natH <= bounds.Height {
			x := bounds.X + (bounds.Width-natW)/2
			y := bounds.Y + (bounds.Height-natH)/2
			return uimath.NewRect(x, y, natW, natH)
		}
		return fitContain(bounds, aspect)
	default: // ObjectFitFill
		return bounds
	}
}

func fitContain(bounds uimath.Rect, aspect float32) uimath.Rect {
	bAspect := bounds.Width / bounds.Height
	var w, h float32
	if aspect > bAspect {
		w = bounds.Width
		h = w / aspect
	} else {
		h = bounds.Height
		w = h * aspect
	}
	x := bounds.X + (bounds.Width-w)/2
	y := bounds.Y + (bounds.Height-h)/2
	return uimath.NewRect(x, y, w, h)
}

func fitCover(bounds uimath.Rect, aspect float32) uimath.Rect {
	bAspect := bounds.Width / bounds.Height
	var w, h float32
	if aspect < bAspect {
		w = bounds.Width
		h = w / aspect
	} else {
		h = bounds.Height
		w = h * aspect
	}
	x := bounds.X + (bounds.Width-w)/2
	y := bounds.Y + (bounds.Height-h)/2
	return uimath.NewRect(x, y, w, h)
}

func (img *Img) drawFallback(buf *render.CommandBuffer, bounds uimath.Rect) {
	cfg := img.config

	if img.state == ImgStateError {
		// Draw broken image placeholder
		buf.DrawRect(render.RectCmd{
			Bounds:      bounds,
			FillColor:   uimath.ColorHex("#f5f5f5"),
			BorderColor: cfg.BorderColor,
			BorderWidth: 1,
			Corners:     uimath.CornersAll(cfg.BorderRadius),
		}, 0, 1)

		// Draw broken image icon
		iconSize := float32(24)
		if iconSize > bounds.Width-8 {
			iconSize = bounds.Width - 8
		}
		if iconSize > bounds.Height-8 {
			iconSize = bounds.Height - 8
		}
		if iconSize > 0 {
			ix := bounds.X + (bounds.Width-iconSize)/2
			iy := bounds.Y + (bounds.Height-iconSize)/2
			if img.alt != "" && cfg.TextRenderer != nil {
				// Icon above, alt text below
				iy = bounds.Y + (bounds.Height-iconSize-cfg.FontSizeSm*1.4)/2
			}
			cfg.DrawMDIcon(buf, "broken_image", ix, iy, iconSize, cfg.DisabledColor, 0, 1)

			// Draw alt text below icon
			if img.alt != "" && cfg.TextRenderer != nil {
				textY := iy + iconSize + 4
				maxW := bounds.Width - 8
				cfg.TextRenderer.DrawText(buf, img.alt, bounds.X+4, textY, cfg.FontSizeSm, maxW, cfg.DisabledColor, 1)
			}
		}
		return
	}

	if img.state == ImgStateLoading {
		// Loading placeholder
		buf.DrawRect(render.RectCmd{
			Bounds:    bounds,
			FillColor: uimath.ColorHex("#f0f0f0"),
			Corners:   uimath.CornersAll(cfg.BorderRadius),
		}, 0, 1)
		return
	}

	// Idle state with alt text
	if img.alt != "" && cfg.TextRenderer != nil {
		cfg.TextRenderer.DrawText(buf, img.alt, bounds.X+4, bounds.Y+4, cfg.FontSize, bounds.Width-8, cfg.TextColor, 1)
	}
}

func isGIF(path string) bool {
	lower := strings.ToLower(path)
	return strings.HasSuffix(lower, ".gif")
}

func cloneRGBA(src *image.RGBA) *image.RGBA {
	dst := image.NewRGBA(src.Bounds())
	copy(dst.Pix, src.Pix)
	return dst
}

type errNoFramesType struct{}

func (errNoFramesType) Error() string { return "GIF has no frames" }

var errNoFrames error = errNoFramesType{}

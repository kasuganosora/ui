// Desktop Pet — a transparent, always-on-top animated sprite on the desktop.
// Uses VPet animation frames from https://github.com/LorisYounger/VPet.
//
// Run:  go run ./cmd/pet
package main

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	ui "github.com/kasuganosora/ui"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/render"
)

// animFrame is one frame of a sprite animation.
type animFrame struct {
	path     string
	index    int
	duration time.Duration
}

const (
	vpetBaseURL  = "https://raw.githubusercontent.com/LorisYounger/VPet/main/VPet-Simulator.Windows/mod/0000_core/pet/vup"
	vpetAnimPath = "Default/Nomal/1"
	cacheDir     = "petcache"
)

var vpetFrames = []string{
	"_000_250.png",
	"_001_125.png",
	"_002_125.png",
	"_003_375.png",
	"_004_125.png",
	"_005_250.png",
	"_006_125.png",
	"_007_125.png",
}

func main() {
	os.MkdirAll(cacheDir, 0755)

	// Download animation frames
	fmt.Println("[pet] downloading animation frames...")
	frames, err := downloadFrames()
	if err != nil {
		fmt.Fprintf(os.Stderr, "download failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[pet] loaded %d frames\n", len(frames))

	// Measure sprite size from first frame
	spriteW, spriteH := measureFrame(frames[0].path)
	if spriteW == 0 {
		spriteW, spriteH = 500, 500
	}

	// Init platform
	plat := ui.NewPlatform()
	if err := plat.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "platform init: %v\n", err)
		os.Exit(1)
	}
	defer plat.Terminate()

	win, err := plat.CreateWindow(platform.WindowOptions{
		Title:       "Desktop Pet",
		Width:       spriteW,
		Height:      spriteH,
		Resizable:   false,
		Visible:     true,
		Decorated:   false,
		Transparent: true,
		TopMost:     true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "create window: %v\n", err)
		os.Exit(1)
	}
	defer win.Destroy()

	// Create rendering backend (DX11 on Windows for DirectComposition transparency)
	backend, err := ui.CreateBackend(ui.BackendDX11, win)
	if err != nil {
		// Fallback to auto
		backend, err = ui.CreateBackend(ui.BackendAuto, win)
		if err != nil {
			fmt.Fprintf(os.Stderr, "backend init: %v\n", err)
			os.Exit(1)
		}
	}
	defer backend.Destroy()

	fw, fh := win.FramebufferSize()
	backend.Resize(fw, fh)

	// Load textures for all frames
	textures := make([]render.TextureHandle, len(frames))
	texW := make([]int, len(frames))
	texH := make([]int, len(frames))
	for i, frame := range frames {
		img, w, h := loadPNG(frame.path)
		if img == nil {
			fmt.Printf("  skip frame %d\n", i)
			continue
		}
		tex, err := backend.CreateTexture(render.TextureDesc{
			Width:  w,
			Height: h,
			Format: render.TextureFormatRGBA8,
		})
		if err != nil {
			fmt.Printf("  texture error: %v\n", err)
			continue
		}
		backend.UpdateTexture(tex, uimath.NewRect(0, 0, float32(w), float32(h)), img.Pix)
		textures[i] = tex
		texW[i] = w
		texH[i] = h
		fmt.Printf("  frame %d: %dx%d, %v\n", frame.index, w, h, frame.duration)
	}

	// Animation state
	currentFrame := 0
	lastSwitch := time.Now()

	// Drag state
	dragging := false
	dragScreenX, dragScreenY := 0, 0
	winStartX, winStartY := 0, 0

	buf := render.NewCommandBuffer()

	fmt.Println("[pet] running! drag to move, right-click to quit.")

	for !win.ShouldClose() {
		events := plat.PollEvents()

		for _, ev := range events {
			switch ev.Type {
			case event.MouseDown:
				if ev.Button == event.MouseButtonRight {
					win.SetShouldClose(true)
				}
				if ev.Button == event.MouseButtonLeft {
					dragging = true
					dragScreenX, dragScreenY = win.ClientToScreen(int(ev.GlobalX), int(ev.GlobalY))
					winStartX, winStartY = win.Position()
				}
			case event.MouseUp:
				if ev.Button == event.MouseButtonLeft {
					dragging = false
				}
			case event.MouseMove:
				if dragging {
					screenX, screenY := win.ClientToScreen(int(ev.GlobalX), int(ev.GlobalY))
					dx := screenX - dragScreenX
					dy := screenY - dragScreenY
					win.SetPosition(winStartX+dx, winStartY+dy)
				}
			}
		}

		// Advance animation
		now := time.Now()
		dur := frames[currentFrame].duration
		if dur <= 0 {
			dur = 125 * time.Millisecond
		}
		if now.Sub(lastSwitch) >= dur {
			currentFrame = (currentFrame + 1) % len(frames)
			lastSwitch = now
		}

		// Render
		backend.BeginFrame()
		buf.Reset()

		if textures[currentFrame] != 0 {
			w := float32(texW[currentFrame])
			h := float32(texH[currentFrame])
			buf.DrawImage(render.ImageCmd{
				Texture: textures[currentFrame],
				SrcRect: uimath.NewRect(0, 0, w, h),
				DstRect: uimath.NewRect(0, 0, float32(spriteW), float32(spriteH)),
				Tint:    uimath.Color{R: 1, G: 1, B: 1, A: 1},
			}, 0, 1.0)
		}

		backend.Submit(buf)
		backend.EndFrame()

		time.Sleep(16 * time.Millisecond)
	}
}

func downloadFrames() ([]animFrame, error) {
	var frames []animFrame
	for _, name := range vpetFrames {
		localPath := filepath.Join(cacheDir, name)
		if _, err := os.Stat(localPath); os.IsNotExist(err) {
			url := fmt.Sprintf("%s/%s/%s", vpetBaseURL, vpetAnimPath, name)
			fmt.Printf("  downloading %s\n", name)
			if err := downloadFile(url, localPath); err != nil {
				return nil, fmt.Errorf("download %s: %w", name, err)
			}
		}
		frame, err := parseFrameName(name, localPath)
		if err != nil {
			continue
		}
		frames = append(frames, frame)
	}
	sort.Slice(frames, func(i, j int) bool { return frames[i].index < frames[j].index })
	return frames, nil
}

func parseFrameName(name, path string) (animFrame, error) {
	base := strings.TrimSuffix(name, ".png")
	parts := strings.Split(base, "_")
	if len(parts) < 3 {
		return animFrame{}, fmt.Errorf("invalid: %s", name)
	}
	idx, err := strconv.Atoi(parts[1])
	if err != nil {
		return animFrame{}, err
	}
	dur, err := strconv.Atoi(parts[2])
	if err != nil {
		return animFrame{}, err
	}
	return animFrame{path: path, index: idx, duration: time.Duration(dur) * time.Millisecond}, nil
}

func downloadFile(url, dst string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func measureFrame(path string) (int, int) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0
	}
	defer f.Close()
	cfg, err := png.DecodeConfig(f)
	if err != nil {
		return 0, 0
	}
	return cfg.Width, cfg.Height
}

func loadPNG(path string) (*image.RGBA, int, int) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, 0
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil, 0, 0
	}
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}
	return rgba, bounds.Dx(), bounds.Dy()
}

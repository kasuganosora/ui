//go:build windows

// Font Viewer - Windows-style font specimen viewer for diagnosing text rendering.
// Run: go run ./cmd/fontviewer
// Optional: go run ./cmd/fontviewer C:\Windows\Fonts\simsun.ttc
package main

import (
	"fmt"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/kasuganosora/ui/font"
	"github.com/kasuganosora/ui/font/atlas"
	"github.com/kasuganosora/ui/font/freetype"
	"github.com/kasuganosora/ui/font/textrender"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/platform/win32"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/render/vulkan"
)

const (
	winWidth  = 1200
	winHeight = 800
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Determine font path from args or default
	fontPath := `C:\Windows\Fonts\msyh.ttc`
	if len(os.Args) > 1 && os.Args[1] != "--screenshot" {
		fontPath = os.Args[1]
	}
	fontName := filepath.Base(fontPath)

	// --- Platform ---
	plat := win32.New()
	if err := plat.Init(); err != nil {
		return fmt.Errorf("platform init: %w", err)
	}
	defer plat.Terminate()

	win, err := plat.CreateWindow(platform.WindowOptions{
		Title:     fmt.Sprintf("Font Viewer — %s", fontName),
		Width:     winWidth,
		Height:    winHeight,
		Resizable: true,
		Visible:   true,
		Decorated: true,
	})
	if err != nil {
		return fmt.Errorf("create window: %w", err)
	}
	defer win.Destroy()

	// --- Renderer ---
	backend := vulkan.New()
	if err := backend.Init(win); err != nil {
		return fmt.Errorf("vulkan init: %w", err)
	}
	defer backend.Destroy()

	// --- Font System ---
	ftEngine, err := freetype.New()
	if err != nil {
		return fmt.Errorf("freetype init: %w", err)
	}
	dpi := backend.DPIScale()
	ftEngine.SetDPIScale(dpi)
	mgr := font.NewManager(ftEngine)
	fontID, err := mgr.RegisterFile("Specimen", font.WeightRegular, font.StyleNormal, fontPath)
	if err != nil || fontID == font.InvalidFontID {
		return fmt.Errorf("load font %s: %w", fontPath, err)
	}
	fmt.Printf("[font] Loaded: %s (DPI scale: %.2f)\n", fontPath, dpi)

	glyphAtlas := atlas.New(atlas.Options{Width: 2048, Height: 2048, Backend: backend})
	tr := textrender.New(textrender.Options{
		Manager:   mgr,
		Atlas:     glyphAtlas,
		DPIScale:  dpi,
		KeepAlive: plat.ProcessMessages,
	})
	defer tr.Destroy()

	buf := render.NewCommandBuffer()

	// --- Specimen data ---
	sampleText := "Innovation in China 中国智造，慧及全球 0123456789"
	sizes := []float32{12, 18, 24, 36, 48, 60, 72}

	headerLines := []string{
		fmt.Sprintf("字体名称: %s", fontName),
		fmt.Sprintf("路径: %s", fontPath),
		"OpenType Layout, TrueType Outlines",
	}
	charSamples := []string{
		"abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		`1234567890.:;' " (!?) +-*/=`,
	}

	// --- Main Loop ---
	var lastW, lastH int
	screenshotMode := len(os.Args) > 1 && os.Args[len(os.Args)-1] == "--screenshot"
	frameNum := 0

	for !win.ShouldClose() {
		plat.PollEvents()

		w, h := win.FramebufferSize()
		if w == 0 || h == 0 {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		if w != lastW || h != lastH {
			backend.Resize(w, h)
			lastW, lastH = w, h
		}

		backend.BeginFrame()
		tr.BeginFrame()
		buf.Reset()

		// White background (logical size = physical / dpiScale)
		logW := float32(w) / dpi
		logH := float32(h) / dpi
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(0, 0, logW, logH),
			FillColor: uimath.ColorWhite,
		}, 0, 1)

		drawSpecimen(buf, tr, ftEngine, fontID, logW, headerLines, charSamples, sampleText, sizes)

		tr.Upload()
		backend.Submit(buf)
		backend.EndFrame()

		frameNum++

		// In screenshot mode, capture after 3 frames (ensures everything is stable)
		if screenshotMode && frameNum == 3 {
			img, err := backend.ReadPixels()
			if err != nil {
				return fmt.Errorf("ReadPixels: %w", err)
			}
			f, err := os.Create("fontviewer_screenshot.png")
			if err != nil {
				return err
			}
			if err := png.Encode(f, img); err != nil {
				f.Close()
				return err
			}
			f.Close()
			fmt.Println("[screenshot] Saved fontviewer_screenshot.png")
			return nil
		}

		time.Sleep(time.Millisecond)
	}

	return nil
}

func drawSpecimen(buf *render.CommandBuffer, tr *textrender.Renderer, engine font.Engine, fontID font.ID, viewW float32, headerLines, charSamples []string, sampleText string, sizes []float32) {
	black := uimath.ColorHex("#000000")
	gray := uimath.ColorHex("#999999")
	leftX := float32(20)

	curY := float32(16)

	// Header
	for _, text := range headerLines {
		tr.DrawText(buf, text, textrender.DrawOptions{
			ShapeOpts: font.ShapeOptions{FontID: fontID, FontSize: 14},
			OriginX:   leftX,
			OriginY:   curY,
			Color:     black,
			Opacity:   1,
		})
		m := engine.FontMetrics(fontID, 14)
		curY += m.Ascent + m.Descent + 2
	}
	curY += 6

	// Character samples
	for _, text := range charSamples {
		tr.DrawText(buf, text, textrender.DrawOptions{
			ShapeOpts: font.ShapeOptions{FontID: fontID, FontSize: 16},
			OriginX:   leftX,
			OriginY:   curY,
			Color:     black,
			Opacity:   1,
		})
		m := engine.FontMetrics(fontID, 16)
		curY += m.Ascent + m.Descent + 4
	}
	curY += 8

	// Separator
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(leftX, curY, viewW-leftX*2, 1),
		FillColor: uimath.ColorHex("#CCCCCC"),
	}, 0, 1)
	curY += 10

	// Specimen at each size
	for _, sz := range sizes {
		// Size label
		label := fmt.Sprintf("%.0f", sz)
		tr.DrawText(buf, label, textrender.DrawOptions{
			ShapeOpts: font.ShapeOptions{FontID: fontID, FontSize: 11},
			OriginX:   leftX,
			OriginY:   curY,
			Color:     gray,
			Opacity:   1,
		})

		// Sample text
		tr.DrawText(buf, sampleText, textrender.DrawOptions{
			ShapeOpts: font.ShapeOptions{FontID: fontID, FontSize: sz},
			OriginX:   leftX + 40,
			OriginY:   curY,
			Color:     black,
			Opacity:   1,
		})

		m := engine.FontMetrics(fontID, sz)
		curY += float32(math.Ceil(float64(m.Ascent+m.Descent))) + 8
	}
}

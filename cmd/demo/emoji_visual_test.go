package main

import (
	"image"
	"os"
	"path/filepath"
	"testing"
	"time"

	ui "github.com/kasuganosora/ui"
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/font"
	"github.com/kasuganosora/ui/font/atlas"
	"github.com/kasuganosora/ui/font/textrender"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/render/capture"
	"github.com/kasuganosora/ui/widget"
)

// emojiTestEnv holds a test environment with color emoji font support.
type emojiTestEnv struct {
	plat         platform.Platform
	win          platform.Window
	backend      render.Backend
	tree         *core.Tree
	cfg          *widget.Config
	buf          *render.CommandBuffer
	textRenderer *textrender.Renderer
	fontEngine   font.Engine
	fontID       font.ID
	fallbackIDs  []font.ID
	width        int
	height       int
}

func newEmojiTestEnv(t *testing.T, width, height int) *emojiTestEnv {
	t.Helper()

	plat := ui.NewPlatform()
	if err := plat.Init(); err != nil {
		t.Fatalf("platform init: %v", err)
	}

	win, err := plat.CreateWindow(platform.WindowOptions{
		Title:     "Emoji Visual Test",
		Width:     width,
		Height:    height,
		Resizable: false,
		Visible:   false,
		Decorated: true,
	})
	if err != nil {
		plat.Terminate()
		t.Fatalf("create window: %v", err)
	}

	backend, err := ui.CreateBackend(ui.BackendAuto, win)
	if err != nil {
		win.Destroy()
		plat.Terminate()
		t.Fatalf("backend init: %v", err)
	}
	fw, fh := win.FramebufferSize()
	backend.Resize(fw, fh)

	var fontEngine font.Engine
	if fe := ui.NewFontEngine(); fe != nil {
		fontEngine = fe
	} else {
		fontEngine = ui.NewMockEngine()
	}
	dpi := backend.DPIScale()
	fontEngine.SetDPIScale(dpi)

	fontMgr := font.NewManager(fontEngine)
	fontID, _ := fontMgr.RegisterFile("Default", font.WeightRegular, font.StyleNormal, ui.DefaultFont())
	if fontID == font.InvalidFontID {
		fontID, _ = fontMgr.Register("Default", font.WeightRegular, font.StyleNormal, nil)
	}

	// Register fallback fonts including emoji
	var fallbackIDs []font.ID
	for _, fbPath := range ui.FallbackFonts() {
		if fbID, err := fontMgr.RegisterFile("Fallback", font.WeightRegular, font.StyleNormal, fbPath); err == nil && fbID != font.InvalidFontID {
			fallbackIDs = append(fallbackIDs, fbID)
			if fontEngine.HasColorGlyphs(fbID) {
				t.Logf("color emoji font loaded: %s (ID=%d)", fbPath, fbID)
			}
		}
	}

	glyphAtlas := atlas.New(atlas.Options{Width: 1024, Height: 1024, Backend: backend})
	tr := textrender.New(textrender.Options{
		Manager:  fontMgr,
		Atlas:    glyphAtlas,
		DPIScale: dpi,
		Backend:  backend,
	})

	cfg := widget.DefaultConfig()
	cfg.TextRenderer = ui.NewTextDrawer(tr, fontID, fontEngine, fallbackIDs...)
	cfg.Backend = backend

	return &emojiTestEnv{
		plat:         plat,
		win:          win,
		backend:      backend,
		tree:         core.NewTree(),
		cfg:          cfg,
		buf:          render.NewCommandBuffer(),
		textRenderer: tr,
		fontEngine:   fontEngine,
		fontID:       fontID,
		fallbackIDs:  fallbackIDs,
		width:        width,
		height:       height,
	}
}

func (e *emojiTestEnv) close() {
	e.textRenderer.Destroy()
	e.backend.Destroy()
	e.win.Destroy()
	e.plat.Terminate()
}

func (e *emojiTestEnv) renderFrames(root widget.Widget, n int) {
	w, h := float32(e.width), float32(e.height)
	for i := 0; i < n; i++ {
		e.plat.PollEvents()
		e.tree.SetLayout(e.tree.Root(), core.LayoutResult{
			Bounds: uimath.NewRect(0, 0, w, h),
		})
		ui.AutoLayout(e.tree, root, w, h)

		e.backend.BeginFrame()
		e.textRenderer.BeginFrame()
		e.buf.Reset()
		root.Draw(e.buf)
		e.textRenderer.Upload()
		e.backend.Submit(e.buf)
		e.backend.EndFrame()
	}
}

func (e *emojiTestEnv) emojiScreenshot(t *testing.T, name string) (string, *image.RGBA) {
	t.Helper()

	img, err := capture.Screenshot(e.backend)
	if err != nil {
		t.Fatalf("screenshot failed: %v", err)
	}

	dir := filepath.Join("testdata", "screenshots")
	os.MkdirAll(dir, 0755)

	path := filepath.Join(dir, name+".png")
	if err := capture.SavePNG(img, path); err != nil {
		t.Fatalf("save screenshot: %v", err)
	}
	t.Logf("screenshot saved: %s (%dx%d)", path, img.Bounds().Dx(), img.Bounds().Dy())
	return path, img
}

const emojiDemoHTML = `<div style="padding: 20px; background: #1a1a2e; width: 100%; height: 100%;">
  <div style="color: #ffffff; font-size: 16px; margin-bottom: 10px;">Emoji 渲染测试</div>

  <div style="color: #e0e0e0; font-size: 14px; margin-bottom: 8px;">表情: 😀 😃 😄 😁 😆 🥹 😅 😂 🤣</div>
  <div style="color: #e0e0e0; font-size: 14px; margin-bottom: 8px;">手势: 👍 👎 👋 🤝 👏 🙏 ✌️ 🤞 🫶</div>
  <div style="color: #e0e0e0; font-size: 14px; margin-bottom: 8px;">动物: 🐶 🐱 🐭 🐹 🐰 🦊 🐻 🐼 🐨</div>
  <div style="color: #e0e0e0; font-size: 14px; margin-bottom: 8px;">食物: 🍎 🍊 🍋 🍇 🍉 🍓 🫐 🍑 🍒</div>
  <div style="color: #e0e0e0; font-size: 14px; margin-bottom: 8px;">旗帜: 🏁 🚩 🎌 🏴 🏳️</div>

  <div style="color: #e0e0e0; font-size: 14px; margin-bottom: 8px;">混合文本: Hello 你好 🌍 World 世界 🎉</div>
  <div style="color: #e0e0e0; font-size: 14px; margin-bottom: 8px;">连续emoji: 🔥🔥🔥💯💯💯✨✨✨</div>

  <div style="color: #ffffff; font-size: 24px; margin-top: 10px;">大号: 🎮 🎲 🎯 🏆 🥇</div>
</div>`

func TestVisualEmoji(t *testing.T) {
	env := newEmojiTestEnv(t, 800, 500)
	defer env.close()

	doc := ui.LoadHTMLDocument(env.tree, env.cfg, emojiDemoHTML, "")
	root := doc.Root
	if len(root.Children()) > 0 {
		root = root.Children()[0]
	}
	env.tree.AppendChild(env.tree.Root(), root.ElementID())

	start := time.Now()
	env.renderFrames(root, 3)
	t.Logf("render time: %v", time.Since(start))

	_, img := env.emojiScreenshot(t, "emoji_demo")

	// Basic sanity checks
	verifyNotUniform(t, img, "emoji_demo")

	// Check that the center region has more than just background + text colors
	// (emoji should contribute many distinct colors if rendered correctly)
	midX := img.Bounds().Dx() / 2
	midY := img.Bounds().Dy() / 2
	colors := countDistinctColors(img, midX-200, midY-100, 400, 200)
	t.Logf("distinct colors in center region: %d", colors)

	// Color emoji should produce many colors; monochrome text only produces 2-3
	if colors < 5 {
		t.Logf("WARNING: very few colors in center region (%d) — emoji may not be rendering in color", colors)
	}

	// Check the color atlas was actually used
	if ca := env.textRenderer.ColorAtlas(); ca != nil {
		t.Logf("color atlas: %dx%d, %d glyphs, occupancy %.1f%%",
			ca.Width(), ca.Height(), ca.GlyphCount(), ca.Occupancy()*100)
		if ca.GlyphCount() == 0 {
			t.Error("color atlas has 0 glyphs — emoji font may not have loaded")
		}
	} else {
		t.Log("WARNING: color atlas was not created — no color emoji font available")
	}
}

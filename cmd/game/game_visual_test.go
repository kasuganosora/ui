package main

import (
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
	"testing"

	ui "github.com/kasuganosora/ui"
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/font"
	"github.com/kasuganosora/ui/font/atlas"
	"github.com/kasuganosora/ui/font/textrender"
	"github.com/kasuganosora/ui/icon/material"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/render/capture"
	"github.com/kasuganosora/ui/widget"
	"github.com/kasuganosora/ui/widget/game"
)

type gameTestEnv struct {
	plat         platform.Platform
	win          platform.Window
	backend      render.Backend
	tree         *core.Tree
	cfg          *widget.Config
	buf          *render.CommandBuffer
	textRenderer *textrender.Renderer
	fontEngine   font.Engine
	fontID       font.ID
	width, height int
}

func newGameTestEnv(t *testing.T, w, h int) *gameTestEnv {
	t.Helper()
	plat := ui.NewPlatform()
	if err := plat.Init(); err != nil {
		t.Fatalf("platform init: %v", err)
	}
	win, err := plat.CreateWindow(platform.WindowOptions{
		Title: "Game Visual Test", Width: w, Height: h, Visible: false,
	})
	if err != nil {
		plat.Terminate()
		t.Fatalf("create window: %v", err)
	}
	backend, err := ui.CreateBackend(ui.BackendAuto, win)
	if err != nil {
		win.Destroy(); plat.Terminate()
		t.Fatalf("backend: %v", err)
	}
	fw, fh := win.FramebufferSize()
	backend.Resize(fw, fh)

	var fe font.Engine
	if e := ui.NewFontEngine(); e != nil {
		fe = e
	} else {
		fe = ui.NewMockEngine()
	}
	mgr := font.NewManager(fe)
	fid, _ := mgr.RegisterFile("Default", font.WeightRegular, font.StyleNormal, ui.DefaultFont())
	if fid == font.InvalidFontID {
		fid, _ = mgr.Register("Default", font.WeightRegular, font.StyleNormal, nil)
	}
	ga := atlas.New(atlas.Options{Width: 1024, Height: 1024, Backend: backend})
	tr := textrender.New(textrender.Options{Manager: mgr, Atlas: ga})
	cfg := widget.DefaultConfig()
	cfg.TextRenderer = ui.NewTextDrawer(tr, fid, fe)
	cfg.IconRegistry = material.NewRegistry(backend)

	return &gameTestEnv{
		plat: plat, win: win, backend: backend,
		tree: core.NewTree(), cfg: cfg,
		buf: render.NewCommandBuffer(), textRenderer: tr,
		fontEngine: fe, fontID: fid,
		width: w, height: h,
	}
}

func (e *gameTestEnv) close() {
	e.textRenderer.Destroy()
	e.backend.Destroy()
	e.win.Destroy()
	e.plat.Terminate()
}

func (e *gameTestEnv) renderFrame(root widget.Widget, onLayout func(float32, float32)) {
	w, h := float32(e.width), float32(e.height)
	e.tree.SetLayout(e.tree.Root(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, w, h),
	})
	if onLayout != nil {
		onLayout(w, h)
	}

	e.backend.BeginFrame()
	e.textRenderer.BeginFrame()
	e.buf.Reset()
	root.Draw(e.buf)
	e.textRenderer.Upload()
	e.backend.Submit(e.buf)
	e.backend.EndFrame()
}

func (e *gameTestEnv) screenshot(t *testing.T, name string) *image.RGBA {
	t.Helper()
	img, err := capture.Screenshot(e.backend)
	if err != nil {
		t.Fatalf("screenshot: %v", err)
	}
	dir := filepath.Join("testdata", "screenshots")
	os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, name+".png")
	capture.SavePNG(img, path)
	t.Logf("screenshot: %s (%dx%d)", path, img.Bounds().Dx(), img.Bounds().Dy())
	return img
}

// countDistinctColors counts unique RGBA values in a region.
func countDistinctColors(img *image.RGBA, x, y, w, h int) int {
	colors := make(map[uint32]struct{})
	for py := y; py < y+h && py < img.Bounds().Dy(); py++ {
		for px := x; px < x+w && px < img.Bounds().Dx(); px++ {
			r, g, b, a := img.At(px, py).RGBA()
			key := (r>>8)<<24 | (g>>8)<<16 | (b>>8)<<8 | (a >> 8)
			colors[key] = struct{}{}
		}
	}
	return len(colors)
}

// verifyNotUniform checks that the image is not all one color.
func verifyNotUniform(t *testing.T, img *image.RGBA) {
	t.Helper()
	n := countDistinctColors(img, 0, 0, img.Bounds().Dx(), img.Bounds().Dy())
	if n <= 1 {
		t.Errorf("image is uniform (%d colors) — nothing rendered", n)
	}
}

// TestVisualGameHUD builds a minimal game HUD and verifies it renders.
func TestVisualGameHUD(t *testing.T) {
	env := newGameTestEnv(t, 1280, 800)
	defer env.close()

	tree := env.tree
	cfg := env.cfg
	cfg.BgColor = uimath.ColorHex("#0a0e17")
	cfg.TextColor = uimath.ColorHex("#c8ccd0")

	// Build a root div manually (simulating what LoadHTML does)
	rootDiv := widget.NewDiv(tree, cfg)
	rootDiv.SetBgColor(uimath.ColorHex("#0a0e17"))
	tree.AppendChild(tree.Root(), rootDiv.ElementID())

	// HUD
	hud := game.NewHUD(tree, cfg)

	// Health bar
	hpBar := game.NewHealthBar(tree, cfg)
	hpBar.SetCurrent(780)
	hpBar.SetMax(1200)
	hpBar.SetBarColor(uimath.ColorHex("#52c41a"))
	hpBar.SetShowText(true)
	hpBar.SetSize(220, 22)
	tree.SetLayout(hpBar.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 220, 22)})
	hud.AddElement(hpBar, game.AnchorTopLeft, 20, 20)

	// Mana bar
	mpBar := game.NewHealthBar(tree, cfg)
	mpBar.SetCurrent(350)
	mpBar.SetMax(600)
	mpBar.SetBarColor(uimath.ColorHex("#1890ff"))
	mpBar.SetShowText(true)
	mpBar.SetSize(220, 18)
	tree.SetLayout(mpBar.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 220, 18)})
	hud.AddElement(mpBar, game.AnchorTopLeft, 20, 48)

	// Hotbar
	hotbar := game.NewHotbar(tree, 10, cfg)
	hotbar.SetSlotSize(52)
	hotbar.SetGap(4)
	hotbar.SetSelected(0)
	for i := 0; i < 10; i++ {
		hotbar.SetSlot(i, game.HotbarSlot{Keybind: fmt.Sprintf("%d", (i+1)%10), Available: true})
	}
	tree.SetLayout(hotbar.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 10*(52+4), 52)})
	hud.AddElement(hotbar, game.AnchorBottomCenter, 0, -20)

	// Minimap
	minimap := game.NewMinimap(tree, cfg)
	minimap.SetSize(160)
	minimap.SetCircular(true)
	minimap.SetPlayerPos(80, 80)
	tree.SetLayout(minimap.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 160, 160)})
	hud.AddElement(minimap, game.AnchorTopRight, -20, 20)

	// Nameplate
	np := game.NewNameplate(tree, "暗影守卫", cfg)
	np.SetLevel(55)
	np.SetHP(3200, 5000)
	np.SetType(game.NameplateHostile)
	np.SetBarSize(100, 6)
	np.SetPosition(480, 350)
	np.SetVisible(true)

	// Append to root div
	rootDiv.AppendChild(hud)
	rootDiv.AppendChild(np)

	// Layout callback
	onLayout := func(w, h float32) {
		tree.SetLayout(rootDiv.ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(0, 0, w, h),
		})
		hud.LayoutElements(w, h)
		tree.SetLayout(np.ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(480, 350, 100, 30),
		})
	}

	// Render 3 frames
	for i := 0; i < 3; i++ {
		env.renderFrame(rootDiv, onLayout)
	}

	img := env.screenshot(t, "game_hud")

	// Verify: image should not be uniform (HUD elements visible)
	verifyNotUniform(t, img)

	// Check distinct colors > 5 (bg + green hp + blue mp + hotbar slots + minimap)
	totalColors := countDistinctColors(img, 0, 0, img.Bounds().Dx(), img.Bounds().Dy())
	t.Logf("total distinct colors: %d", totalColors)
	if totalColors < 5 {
		t.Errorf("expected >= 5 distinct colors, got %d", totalColors)
	}

	// Check HP bar region has green pixels (top-left area)
	dpi := env.win.DPIScale()
	hpRegionColors := countDistinctColors(img,
		int(20*dpi), int(20*dpi), int(220*dpi), int(22*dpi))
	t.Logf("HP bar region colors: %d", hpRegionColors)
	if hpRegionColors < 2 {
		t.Errorf("HP bar region has %d colors, expected >= 2 (bg + green fill)", hpRegionColors)
	}

	// Check hotbar region (bottom-center) — screenshot is at logical resolution
	imgW, imgH := img.Bounds().Dx(), img.Bounds().Dy()
	hotbarY := imgH - 72
	hotbarX := imgW/2 - 280
	if hotbarX < 0 {
		hotbarX = 0
	}
	hotbarColors := countDistinctColors(img, hotbarX, hotbarY, 560, 52)
	t.Logf("Hotbar region colors: %d (at %d,%d)", hotbarColors, hotbarX, hotbarY)
	if hotbarColors < 2 {
		t.Errorf("Hotbar region has %d colors, expected >= 2", hotbarColors)
	}

	// Dump command buffer stats
	t.Logf("command buffer: %d commands, %d overlays", env.buf.Len(), len(env.buf.Overlays()))

	// Dump tree
	tree.Walk(tree.Root(), func(id core.ElementID, depth int) bool {
		elem := tree.Get(id)
		if elem == nil {
			return false
		}
		indent := ""
		for i := 0; i < depth; i++ {
			indent += "  "
		}
		b := elem.Layout().Bounds
		t.Logf("%s[%d] %s (%.0f,%.0f %.0fx%.0f) children=%d",
			indent, id, elem.Type(), b.X, b.Y, b.Width, b.Height, len(elem.ChildIDs()))
		return true
	})
}

// TestVisualGameHUD_WithCSSLayout tests the exact same flow as the real app.
func TestVisualGameHUD_WithCSSLayout(t *testing.T) {
	env := newGameTestEnv(t, 1280, 800)
	defer env.close()

	tree := env.tree
	cfg := env.cfg
	cfg.BgColor = uimath.ColorHex("#0a0e17")
	cfg.TextColor = uimath.ColorHex("#c8ccd0")

	// Use LoadHTMLDocument (same as app.LoadHTML)
	doc := ui.LoadHTMLDocument(tree, cfg, `<div style="width:100%; height:100%; background:#0a0e17;"></div>`, "")

	// Get the inner div (same as app does: doc.Root.Children()[0])
	var root widget.Widget
	if len(doc.Root.Children()) > 0 {
		root = doc.Root.Children()[0]
	} else {
		root = doc.Root
	}

	// Attach to tree root (same as app)
	tree.AppendChild(tree.Root(), root.ElementID())

	// Cast to Div to use AppendChild
	rootDiv, ok := root.(*widget.Div)
	if !ok {
		t.Fatal("root is not a Div")
	}

	// Build HUD
	hud := game.NewHUD(tree, cfg)
	hpBar := game.NewHealthBar(tree, cfg)
	hpBar.SetCurrent(780)
	hpBar.SetMax(1200)
	hpBar.SetBarColor(uimath.ColorHex("#52c41a"))
	hpBar.SetShowText(true)
	hpBar.SetSize(220, 22)
	tree.SetLayout(hpBar.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 220, 22)})
	hud.AddElement(hpBar, game.AnchorTopLeft, 20, 20)

	hotbar := game.NewHotbar(tree, 10, cfg)
	hotbar.SetSlotSize(52)
	hotbar.SetGap(4)
	hotbar.SetSelected(0)
	for i := 0; i < 10; i++ {
		hotbar.SetSlot(i, game.HotbarSlot{Keybind: fmt.Sprintf("%d", (i+1)%10), Available: true})
	}
	tree.SetLayout(hotbar.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 10*(52+4), 52)})
	hud.AddElement(hotbar, game.AnchorBottomCenter, 0, -20)

	np := game.NewNameplate(tree, "暗影守卫", cfg)
	np.SetLevel(55)
	np.SetHP(3200, 5000)
	np.SetType(game.NameplateHostile)
	np.SetBarSize(100, 6)
	np.SetPosition(480, 350)
	np.SetVisible(true)

	rootDiv.AppendChild(hud)
	rootDiv.AppendChild(np)

	// Layout callback (exactly like real app)
	onLayout := func(w, h float32) {
		ui.CSSLayout(tree, root, w, h, cfg)
		// Force root to fill viewport (same fix as in cmd/game/main.go)
		tree.SetLayout(root.ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(0, 0, w, h),
		})
		hud.LayoutElements(w, h)
		tree.SetLayout(np.ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(480, 350, 100, 30),
		})
	}

	// Render
	for i := 0; i < 3; i++ {
		env.renderFrame(root, onLayout)
	}

	img := env.screenshot(t, "game_hud_csslayout")
	verifyNotUniform(t, img)

	totalColors := countDistinctColors(img, 0, 0, img.Bounds().Dx(), img.Bounds().Dy())
	t.Logf("total distinct colors (CSSLayout path): %d", totalColors)
	if totalColors < 5 {
		t.Errorf("expected >= 5 distinct colors, got %d", totalColors)
	}

	// Dump tree
	tree.Walk(tree.Root(), func(id core.ElementID, depth int) bool {
		elem := tree.Get(id)
		if elem == nil {
			return false
		}
		indent := ""
		for i := 0; i < depth; i++ {
			indent += "  "
		}
		b := elem.Layout().Bounds
		t.Logf("%s[%d] %s (%.0f,%.0f %.0fx%.0f) children=%d",
			indent, id, elem.Type(), b.X, b.Y, b.Width, b.Height, len(elem.ChildIDs()))
		return true
	})

	t.Logf("command buffer: %d commands, %d overlays", env.buf.Len(), len(env.buf.Overlays()))
}

// TestVisualGameWindows tests the WindowManager layout: windows should have
// title bars, content should be contained within window bounds, no overflow.
func TestVisualGameWindows(t *testing.T) {
	env := newGameTestEnv(t, 1280, 800)
	defer env.close()

	tree := env.tree
	cfg := env.cfg
	cfg.BgColor = uimath.ColorHex("#0a0e17")
	cfg.TextColor = uimath.ColorHex("#c8ccd0")

	doc := ui.LoadHTMLDocument(tree, cfg, `<div style="position:relative; width:100%; height:100%; background:#0a0e17;"></div>`, "")
	rootDiv := doc.Root.Children()[0].(*widget.Div)
	tree.AppendChild(tree.Root(), rootDiv.ElementID())

	wm := game.NewWindowManager(tree, rootDiv)

	// ── Chat Window ──
	chatWin := game.NewWindow(tree, "聊天", cfg)
	chatWin.SetTitleH(24)
	chatContent := widget.NewDiv(tree, cfg)
	chatContent.SetBgColor(uimath.RGBA(0.1, 0.1, 0.15, 1))
	chatContent.SetStyle(layout.Style{Width: layout.Px(340), Height: layout.Px(172)})
	chatWin.AppendChild(chatContent)
	wm.Add(chatWin, 10, 520, 340, 228)

	// ── Inventory Window ──
	inv := game.NewInventory(tree, 3, 4, cfg)
	inv.SetSlotSize(44)
	inv.SetGap(3)
	inv.SetEmbedded(true)
	invW := float32(4*(44+3)) + 20
	invH := float32(3*(44+3)) + 12
	invWin := game.NewWindow(tree, "背包", cfg)
	inv.SetStyle(layout.Style{Width: layout.Px(invW), Height: layout.Px(invH)})
	invWin.AppendChild(inv)
	wm.Add(invWin, 980, 250, invW, invH+28)

	// ── Dialogue Window ──
	dialogue := game.NewDialogueBox(tree, cfg)
	dialogue.SetSize(480, 130)
	dialogue.SetEmbedded(true)
	dialogue.Show("旅行商人", "你好，旅人。")
	dialogueWin := game.NewWindow(tree, "对话", cfg)
	dialogue.SetStyle(layout.Style{Width: layout.Px(480), Height: layout.Px(130)})
	dialogueWin.AppendChild(dialogue)
	wm.Add(dialogueWin, 400, 352, 480, 158)

	// ── Score Window ──
	scoreboard := game.NewScoreboard(tree, cfg)
	scoreboard.SetTitle("")
	scoreboard.SetEmbedded(true)
	scoreboard.SetWidth(360)
	scoreboard.AddEntry(game.ScoreEntry{Name: "Player1", Score: 100, Team: 1})
	scoreboard.SetVisible(true)
	scoreWin := game.NewWindow(tree, "统计", cfg)
	scoreboard.SetStyle(layout.Style{Width: layout.Px(360), Height: layout.Px(220)})
	scoreWin.AppendChild(scoreboard)
	wm.Add(scoreWin, 270, 216, 360, 248)

	layoutCache := ui.NewCSSLayoutCache()
	onLayout := func(w, h float32) {
		layoutCache.Layout(tree, rootDiv, w, h, cfg)
		wm.PostLayout()
	}

	for i := 0; i < 3; i++ {
		env.renderFrame(rootDiv, onLayout)
	}

	img := env.screenshot(t, "game_windows")
	verifyNotUniform(t, img)

	dpi := env.win.DPIScale()

	// Verify each window's content is within its bounds.
	// Check that pixels OUTSIDE window bounds (but near them) are dark background.
	type windowCheck struct {
		name       string
		x, y, w, h float32
	}
	checks := []windowCheck{
		{"chat", 10, 520, 340, 228},
		{"inv", 980, 250, invW, invH + 28},
		{"dialogue", 400, 352, 480, 158},
		{"score", 270, 216, 360, 248},
	}

	for _, wc := range checks {
		// Check inside window has some content (not uniform background)
		px := int((wc.x + 10) * dpi)
		py := int((wc.y + wc.h/2) * dpi)
		pw := int((wc.w - 20) * dpi)
		ph := int(40 * dpi)
		if px < 0 || py < 0 || px+pw > img.Bounds().Dx() || py+ph > img.Bounds().Dy() {
			continue
		}
		inner := countDistinctColors(img, px, py, pw, ph)
		t.Logf("window %q inner colors: %d (region %d,%d %dx%d)", wc.name, inner, px, py, pw, ph)

		// Check just below the window: should be dark background
		belowY := int((wc.y + wc.h + 5) * dpi)
		if belowY+10 < img.Bounds().Dy() {
			belowColors := countDistinctColors(img, int(wc.x*dpi), belowY, int(wc.w*dpi), 10)
			t.Logf("window %q below colors: %d", wc.name, belowColors)
			if belowColors > 5 {
				t.Errorf("window %q content leaks below bounds (%d colors below)", wc.name, belowColors)
			}
		}
	}

	// Dump tree for debugging
	tree.Walk(tree.Root(), func(id core.ElementID, depth int) bool {
		elem := tree.Get(id)
		if elem == nil {
			return false
		}
		indent := ""
		for i := 0; i < depth; i++ {
			indent += "  "
		}
		b := elem.Layout().Bounds
		t.Logf("%s[%d] %s (%.0f,%.0f %.0fx%.0f) children=%d",
			indent, id, elem.Type(), b.X, b.Y, b.Width, b.Height, len(elem.ChildIDs()))
		return true
	})

	t.Logf("commands: %d", env.buf.Len())
}

// Suppress unused import warnings
var _ = math.Sin

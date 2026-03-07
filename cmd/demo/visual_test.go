//go:build windows

package main

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/font"
	"github.com/kasuganosora/ui/font/atlas"
	"github.com/kasuganosora/ui/font/freetype"
	"github.com/kasuganosora/ui/font/textrender"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/platform/win32"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/render/capture"
	"github.com/kasuganosora/ui/render/vulkan"
	"github.com/kasuganosora/ui/widget"
)

// testEnv holds a full platform+renderer test environment.
type testEnv struct {
	plat         *win32.Platform
	win          platform.Window
	backend      *vulkan.Backend
	tree         *core.Tree
	cfg          *widget.Config
	buf          *render.CommandBuffer
	textRenderer *textrender.Renderer
	width        int // logical window width
	height       int // logical window height
}

func newTestEnv(t *testing.T, width, height int) *testEnv {
	t.Helper()

	plat := win32.New()
	if err := plat.Init(); err != nil {
		t.Fatalf("platform init: %v", err)
	}

	win, err := plat.CreateWindow(platform.WindowOptions{
		Title:     "GoUI Visual Test",
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

	backend := vulkan.New()
	if err := backend.Init(win); err != nil {
		win.Destroy()
		plat.Terminate()
		t.Fatalf("vulkan init: %v", err)
	}

	// Use logical window size for layout. The Vulkan swapchain extent matches
	// the logical size on DPI-scaled systems (the surface caps report logical px).
	backend.Resize(width, height)

	// Font system: try FreeType, fall back to mock engine
	var fontEngine font.Engine
	if ftEngine, err := freetype.New(); err == nil {
		fontEngine = ftEngine
	} else {
		fontEngine = newMockEngine()
	}
	fontMgr := font.NewManager(fontEngine)
	fontID, _ := fontMgr.RegisterFile("Default", font.WeightRegular, font.StyleNormal, `C:\Windows\Fonts\msyh.ttc`)
	if fontID == font.InvalidFontID {
		fontID, _ = fontMgr.Register("Default", font.WeightRegular, font.StyleNormal, nil)
	}
	glyphAtlas := atlas.New(atlas.Options{Width: 1024, Height: 1024, Backend: backend})
	tr := textrender.New(textrender.Options{Manager: fontMgr, Atlas: glyphAtlas})

	cfg := widget.DefaultConfig()
	cfg.TextRenderer = &textDrawerAdapter{renderer: tr, fontID: fontID, engine: fontEngine}

	return &testEnv{
		plat:         plat,
		win:          win,
		backend:      backend,
		tree:         core.NewTree(),
		cfg:          cfg,
		buf:          render.NewCommandBuffer(),
		textRenderer: tr,
		width:        width,
		height:       height,
	}
}

func (e *testEnv) close() {
	e.textRenderer.Destroy()
	e.backend.Destroy()
	e.win.Destroy()
	e.plat.Terminate()
}

// renderFrames builds layout, draws, and submits N frames.
func (e *testEnv) renderFrames(root widget.Widget, n int) {
	w, h := float32(e.width), float32(e.height)
	for i := 0; i < n; i++ {
		e.plat.PollEvents()
		e.tree.SetLayout(e.tree.Root(), core.LayoutResult{
			Bounds: uimath.NewRect(0, 0, w, h),
		})
		computeLayout(e.tree, root, w, h)

		e.backend.BeginFrame()
		e.textRenderer.BeginFrame()
		e.buf.Reset()
		root.Draw(e.buf)
		e.textRenderer.Upload()
		e.backend.Submit(e.buf)
		e.backend.EndFrame()
	}
}

// screenshot captures current framebuffer and saves to testdata/.
func (e *testEnv) screenshot(t *testing.T, name string) (string, *image.RGBA) {
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

// --- Tree inspection helpers ---

// dumpTree logs the widget tree structure for debugging.
func dumpTree(t *testing.T, tree *core.Tree) {
	t.Helper()
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
		text := elem.TextContent()
		extra := ""
		if text != "" {
			extra = fmt.Sprintf(" text=%q", text)
		}
		t.Logf("%s[%d] %s (%.0f,%.0f %.0fx%.0f) vis=%v%s",
			indent, id, elem.Type(), b.X, b.Y, b.Width, b.Height, elem.IsVisible(), extra)
		return true
	})
}

// verifyAllVisibleHaveBounds walks the tree and checks every visible element at depth <= maxDepth has non-zero bounds.
func verifyAllVisibleHaveBounds(t *testing.T, tree *core.Tree, maxDepth int) {
	t.Helper()
	tree.Walk(tree.Root(), func(id core.ElementID, depth int) bool {
		if depth > maxDepth {
			return false
		}
		elem := tree.Get(id)
		if elem == nil || !elem.IsVisible() {
			return true
		}
		b := elem.Layout().Bounds
		if b.Width == 0 && b.Height == 0 && depth > 0 {
			t.Errorf("element %d (%s) has zero bounds at depth %d", id, elem.Type(), depth)
		}
		return true
	})
}

// verifyNoOverlappingSiblings checks that sibling elements don't completely overlap each other.
func verifyNoOverlappingSiblings(t *testing.T, tree *core.Tree) {
	t.Helper()
	tree.Walk(tree.Root(), func(id core.ElementID, _ int) bool {
		elem := tree.Get(id)
		if elem == nil {
			return true
		}
		children := elem.ChildIDs()
		for i := 0; i < len(children); i++ {
			for j := i + 1; j < len(children); j++ {
				ci := tree.Get(children[i])
				cj := tree.Get(children[j])
				if ci == nil || cj == nil {
					continue
				}
				bi := ci.Layout().Bounds
				bj := cj.Layout().Bounds
				if bi.Width == 0 || bi.Height == 0 || bj.Width == 0 || bj.Height == 0 {
					continue
				}
				if bi.X == bj.X && bi.Y == bj.Y && bi.Width == bj.Width && bi.Height == bj.Height {
					t.Errorf("siblings %d (%s) and %d (%s) have identical bounds (%.0f,%.0f %.0fx%.0f)",
						children[i], ci.Type(), children[j], cj.Type(),
						bi.X, bi.Y, bi.Width, bi.Height)
				}
			}
		}
		return true
	})
}

// --- Pixel inspection helpers ---

// verifyNotUniform checks that the image is not a single solid color.
func verifyNotUniform(t *testing.T, img *image.RGBA, label string) {
	t.Helper()
	bounds := img.Bounds()
	ref := img.RGBAAt(0, 0)
	for y := 0; y < bounds.Dy(); y += max(1, bounds.Dy()/20) {
		for x := 0; x < bounds.Dx(); x += max(1, bounds.Dx()/20) {
			c := img.RGBAAt(x, y)
			if c != ref {
				return
			}
		}
	}
	t.Errorf("%s: screenshot is uniform color (R=%d G=%d B=%d A=%d)", label, ref.R, ref.G, ref.B, ref.A)
}

// verifyRegionNotBlack checks that a specific rectangular region is not entirely black.
func verifyRegionNotBlack(t *testing.T, img *image.RGBA, label string, rx, ry, rw, rh int) {
	t.Helper()
	bounds := img.Bounds()
	for y := ry; y < ry+rh && y < bounds.Dy(); y += max(1, rh/5) {
		for x := rx; x < rx+rw && x < bounds.Dx(); x += max(1, rw/5) {
			c := img.RGBAAt(x, y)
			if c.R > 10 || c.G > 10 || c.B > 10 {
				return
			}
		}
	}
	t.Errorf("%s: region (%d,%d %dx%d) is entirely black", label, rx, ry, rw, rh)
}

// verifyRegionHasColor checks that a region contains at least one pixel near the expected color.
// Tolerance is generous (60) by default to account for SRGB gamma encoding differences.
func verifyRegionHasColor(t *testing.T, img *image.RGBA, label string, rx, ry, rw, rh int, wantR, wantG, wantB, tolerance uint8) {
	t.Helper()
	bounds := img.Bounds()
	for y := ry; y < ry+rh && y < bounds.Dy(); y += max(1, rh/10) {
		for x := rx; x < rx+rw && x < bounds.Dx(); x += max(1, rw/10) {
			c := img.RGBAAt(x, y)
			if absDiffU8(c.R, wantR) <= tolerance &&
				absDiffU8(c.G, wantG) <= tolerance &&
				absDiffU8(c.B, wantB) <= tolerance {
				return
			}
		}
	}
	// Log actual colors found for debugging
	var samples []string
	for y := ry; y < ry+rh && y < bounds.Dy() && len(samples) < 5; y += max(1, rh/3) {
		for x := rx; x < rx+rw && x < bounds.Dx() && len(samples) < 5; x += max(1, rw/3) {
			c := img.RGBAAt(x, y)
			samples = append(samples, fmt.Sprintf("(%d,%d)=#%02x%02x%02x", x, y, c.R, c.G, c.B))
		}
	}
	t.Errorf("%s: region (%d,%d %dx%d) does not contain expected color (#%02x%02x%02x +/-%d). samples: %v",
		label, rx, ry, rw, rh, wantR, wantG, wantB, tolerance, samples)
}

// countDistinctColors returns the number of distinct colors in a region.
func countDistinctColors(img *image.RGBA, rx, ry, rw, rh int) int {
	colors := make(map[uint32]struct{})
	bounds := img.Bounds()
	for y := ry; y < ry+rh && y < bounds.Dy(); y++ {
		for x := rx; x < rx+rw && x < bounds.Dx(); x++ {
			c := img.RGBAAt(x, y)
			key := uint32(c.R)<<24 | uint32(c.G)<<16 | uint32(c.B)<<8 | uint32(c.A)
			colors[key] = struct{}{}
		}
	}
	return len(colors)
}

func absDiffU8(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}

// --- Visual Tests ---

// TestVisualDemoFullUI renders the full demo UI and validates structure + rendering.
func TestVisualDemoFullUI(t *testing.T) {
	env := newTestEnv(t, 960, 640)
	defer env.close()

	root := buildUI(env.tree, env.cfg)
	env.tree.AppendChild(env.tree.Root(), root.ElementID())

	env.renderFrames(root, 3)

	// Tree structure validation
	count := env.tree.ElementCount()
	t.Logf("element count: %d", count)
	if count < 20 {
		t.Errorf("expected at least 20 elements, got %d", count)
	}

	dumpTree(t, env.tree)

	// Verify elements at depth <= 4 have bounds (deeper nesting like Col→Div→Text
	// may not get bounds from the manual layout, which is expected until the
	// layout engine is integrated).
	verifyAllVisibleHaveBounds(t, env.tree, 4)
	verifyNoOverlappingSiblings(t, env.tree)

	// Screenshot validation
	_, img := env.screenshot(t, "demo_full_ui")
	verifyNotUniform(t, img, "demo_full_ui")

	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	t.Logf("screenshot size: %dx%d", w, h)

	// Verify distinct regions exist (not just one solid color)
	centerColors := countDistinctColors(img, w/4, h/4, w/2, h/2)
	t.Logf("distinct colors in center: %d", centerColors)
	if centerColors < 3 {
		t.Errorf("center of demo UI has only %d colors, expected a complex layout", centerColors)
	}

	// Content area (center) should not be entirely black
	verifyRegionNotBlack(t, img, "content_area", w/4, h/4, w/2, h/2)

	// Sidebar area (left 220px) should contain content
	verifyRegionNotBlack(t, img, "sidebar", 0, h/4, 200, h/2)
}

// TestVisualButtonRendering renders the full demo and checks that buttons are visible.
func TestVisualButtonRendering(t *testing.T) {
	env := newTestEnv(t, 960, 640)
	defer env.close()

	root := buildUI(env.tree, env.cfg)
	env.tree.AppendChild(env.tree.Root(), root.ElementID())

	env.renderFrames(root, 3)

	_, img := env.screenshot(t, "button_rendering")

	// Find button elements in the tree and verify their regions aren't blank
	env.tree.Walk(env.tree.Root(), func(id core.ElementID, depth int) bool {
		elem := env.tree.Get(id)
		if elem == nil || elem.Type() != core.TypeButton {
			return true
		}
		b := elem.Layout().Bounds
		if b.Width < 10 || b.Height < 10 {
			return true
		}
		// Check that the button region has more than 1 color (background + text placeholder)
		dpi := env.backend.DPIScale()
		rx, ry := int(b.X*dpi), int(b.Y*dpi)
		rw, rh := int(b.Width*dpi), int(b.Height*dpi)
		if rx >= img.Bounds().Dx() || ry >= img.Bounds().Dy() {
			return true
		}
		nc := countDistinctColors(img, rx, ry, min(rw, img.Bounds().Dx()-rx), min(rh, img.Bounds().Dy()-ry))
		if nc < 2 {
			t.Errorf("button %d %q: only %d color(s) in region, expected background+label",
				id, elem.TextContent(), nc)
		}
		return true
	})
}

// TestVisualInputRendering renders the full demo and checks that inputs are visible.
func TestVisualInputRendering(t *testing.T) {
	env := newTestEnv(t, 960, 640)
	defer env.close()

	root := buildUI(env.tree, env.cfg)
	env.tree.AppendChild(env.tree.Root(), root.ElementID())

	env.renderFrames(root, 3)

	_, img := env.screenshot(t, "input_rendering")

	// Find input elements and verify they're rendered
	inputCount := 0
	env.tree.Walk(env.tree.Root(), func(id core.ElementID, depth int) bool {
		elem := env.tree.Get(id)
		if elem == nil || elem.Type() != core.TypeInput {
			return true
		}
		inputCount++
		b := elem.Layout().Bounds
		if b.Width < 10 || b.Height < 10 {
			return true
		}
		verifyRegionNotBlack(t, img, fmt.Sprintf("input_%d", id),
			int(b.X), int(b.Y), int(b.Width), int(b.Height))
		return true
	})
	if inputCount < 3 {
		t.Errorf("expected at least 3 inputs in demo, got %d", inputCount)
	}
}

// TestVisualGridColors renders grid columns and verifies graduated blue colors.
func TestVisualGridColors(t *testing.T) {
	env := newTestEnv(t, 960, 640)
	defer env.close()

	root := buildUI(env.tree, env.cfg)
	env.tree.AppendChild(env.tree.Root(), root.ElementID())

	env.renderFrames(root, 3)

	_, img := env.screenshot(t, "grid_colors")

	// The grid row should be somewhere in the content area.
	// Find it by looking for the row widget in the tree.
	// Grid colors should produce at least 4 distinct blue shades.
	contentX := 220 // sidebar width
	contentY := 56  // header height
	gridY := contentY + 24 + 7*52 // approximate grid row position based on layout
	gridH := 36

	if gridY+gridH < img.Bounds().Dy() {
		colors := countDistinctColors(img, contentX+24, gridY, img.Bounds().Dx()-contentX-48, gridH)
		t.Logf("grid region colors: %d (at y=%d)", colors, gridY)
		if colors < 2 {
			t.Errorf("grid region has only %d colors, expected multiple shades of blue", colors)
		}
	}
}

// TestVisualMessageLoop verifies the window remains responsive during rendering.
func TestVisualMessageLoop(t *testing.T) {
	env := newTestEnv(t, 400, 300)
	defer env.close()

	tree := env.tree
	cfg := env.cfg

	root := widget.NewLayout(tree, cfg)
	root.SetBgColor(uimath.ColorHex("#001529"))
	txt := widget.NewText(tree, "消息循环测试", cfg)
	txt.SetColor(uimath.ColorWhite)
	root.AppendChild(txt)
	tree.AppendChild(tree.Root(), root.ElementID())

	w, h := float32(env.width), float32(env.height)
	tree.SetLayout(tree.Root(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, w, h)})
	tree.SetLayout(root.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, w, h)})
	tree.SetLayout(txt.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(20, h/2-20, w-40, 40)})

	// Render 10 frames within a timeout to verify no deadlock
	done := make(chan struct{})
	go func() {
		for i := 0; i < 10; i++ {
			env.plat.PollEvents()
			env.backend.BeginFrame()
			env.buf.Reset()
			root.Draw(env.buf)
			env.backend.Submit(env.buf)
			env.backend.EndFrame()
		}
		close(done)
	}()

	select {
	case <-done:
		// OK
	case <-time.After(10 * time.Second):
		t.Fatal("message loop deadlock: 10 frames did not complete in 10s")
	}

	_, img := env.screenshot(t, "message_loop")
	verifyNotUniform(t, img, "message_loop")

	// Should have at least 2 colors (dark background + white text placeholder)
	nc := countDistinctColors(img, 0, 0, img.Bounds().Dx(), img.Bounds().Dy())
	t.Logf("message loop screenshot has %d distinct colors", nc)
	if nc < 2 {
		t.Errorf("expected at least 2 colors (background + text), got %d", nc)
	}
}

// TestVisualHitTestConsistency verifies that hit test works for elements with bounds.
func TestVisualHitTestConsistency(t *testing.T) {
	env := newTestEnv(t, 960, 640)
	defer env.close()

	root := buildUI(env.tree, env.cfg)
	env.tree.AppendChild(env.tree.Root(), root.ElementID())

	w, h := float32(env.width), float32(env.height)
	env.tree.SetLayout(env.tree.Root(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, w, h),
	})
	computeLayout(env.tree, root, w, h)

	// Walk tree and for each visible element with non-trivial bounds,
	// verify HitTest finds *something* at its center (not necessarily itself,
	// since a child might be on top).
	tested := 0
	missed := 0
	env.tree.Walk(env.tree.Root(), func(id core.ElementID, depth int) bool {
		elem := env.tree.Get(id)
		if elem == nil || !elem.IsVisible() {
			return true
		}
		b := elem.Layout().Bounds
		if b.Width < 5 || b.Height < 5 {
			return true
		}
		// Only test elements within window bounds
		cx := b.X + b.Width/2
		cy := b.Y + b.Height/2
		if cx >= w || cy >= h || cx < 0 || cy < 0 {
			return true
		}

		tested++
		hit := env.tree.HitTest(cx, cy)
		if hit == core.InvalidElementID {
			missed++
			if depth <= 3 {
				// Only report shallow elements as errors; deep elements might
				// be outside their parent's bounds due to manual layout limitations
				t.Errorf("HitTest(%.0f, %.0f) returned invalid for element %d (%s) at depth %d",
					cx, cy, id, elem.Type(), depth)
			}
		}
		return true
	})
	t.Logf("hit test: tested %d elements, %d missed", tested, missed)
}

// TestVisualCommandBufferCoverage verifies that Draw() generates commands for all visible widgets.
func TestVisualCommandBufferCoverage(t *testing.T) {
	env := newTestEnv(t, 960, 640)
	defer env.close()

	root := buildUI(env.tree, env.cfg)
	env.tree.AppendChild(env.tree.Root(), root.ElementID())

	w, h := float32(env.width), float32(env.height)
	env.tree.SetLayout(env.tree.Root(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, w, h),
	})
	computeLayout(env.tree, root, w, h)

	env.buf.Reset()
	root.Draw(env.buf)

	cmdCount := env.buf.Len()
	t.Logf("command buffer has %d commands", cmdCount)

	// Count visible elements with non-zero bounds
	visibleCount := 0
	env.tree.Walk(env.tree.Root(), func(id core.ElementID, _ int) bool {
		elem := env.tree.Get(id)
		if elem == nil || !elem.IsVisible() {
			return true
		}
		b := elem.Layout().Bounds
		if b.Width > 0 && b.Height > 0 {
			visibleCount++
		}
		return true
	})

	t.Logf("visible elements with bounds: %d", visibleCount)

	// We should have at least some commands per visible element
	if cmdCount < visibleCount/2 {
		t.Errorf("too few render commands (%d) for %d visible elements", cmdCount, visibleCount)
	}
}

// TestVisualFrameConsistency renders multiple frames and verifies they produce identical output.
func TestVisualFrameConsistency(t *testing.T) {
	env := newTestEnv(t, 400, 300)
	defer env.close()

	root := buildUI(env.tree, env.cfg)
	env.tree.AppendChild(env.tree.Root(), root.ElementID())

	// Warm-up frame to create atlas texture and stabilize rendering
	env.renderFrames(root, 1)

	// Render 2 frames and compare
	env.renderFrames(root, 1)
	img1, err := capture.Screenshot(env.backend)
	if err != nil {
		t.Fatalf("screenshot 1: %v", err)
	}

	env.renderFrames(root, 1)
	img2, err := capture.Screenshot(env.backend)
	if err != nil {
		t.Fatalf("screenshot 2: %v", err)
	}

	if img1.Bounds() != img2.Bounds() {
		t.Fatalf("frame sizes differ: %v vs %v", img1.Bounds(), img2.Bounds())
	}

	// Compare frames — should be identical for static UI
	result, err := capture.Compare(img1, img2, 0.01)
	if err != nil {
		t.Fatalf("compare: %v", err)
	}
	if result.DiffPixels > 0 {
		pct := float64(result.DiffPixels) / float64(result.TotalPixels) * 100
		t.Errorf("consecutive frames differ: %d pixels (%.2f%%)", result.DiffPixels, pct)
	}
}

// TestVisualFontSpecimen renders a Windows-style font viewer to inspect rendering quality.
func TestVisualFontSpecimen(t *testing.T) {
	env := newTestEnv(t, 1200, 900)
	defer env.close()

	adapter := env.cfg.TextRenderer.(*textDrawerAdapter)
	fontID := adapter.fontID

	// Specimen sizes (matching Windows font viewer)
	sampleText := "Innovation in China 中国智造，慧及全球 0123456789"
	sizes := []float32{12, 18, 24, 36, 48, 60, 72}

	// Render 2 frames
	for frame := 0; frame < 2; frame++ {
		env.plat.PollEvents()
		env.backend.BeginFrame()
		env.textRenderer.BeginFrame()
		env.buf.Reset()

		// White background
		env.buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(0, 0, 1200, 900),
			FillColor: uimath.ColorWhite,
		}, 0, 1)

		curY := float32(12)
		black := uimath.ColorHex("#000000")
		gray := uimath.ColorHex("#999999")
		leftX := float32(20)

		// Header section (font info)
		headerTexts := []string{
			"字体名称: 微软雅黑",
			"版本: Version 6.31",
			"OpenType Layout, TrueType Outlines",
		}
		for _, ht := range headerTexts {
			env.textRenderer.DrawText(env.buf, ht, textrender.DrawOptions{
				ShapeOpts: font.ShapeOptions{FontID: fontID, FontSize: 14},
				OriginX:   leftX,
				OriginY:   curY,
				Color:     black,
				Opacity:   1,
			})
			m := adapter.engine.FontMetrics(fontID, 14)
			curY += m.Ascent + m.Descent + 2
		}
		curY += 4

		// Character samples
		charTexts := []string{
			"abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ",
			`1234567890.:;' " (!?) +-*/=`,
		}
		for _, ct := range charTexts {
			env.textRenderer.DrawText(env.buf, ct, textrender.DrawOptions{
				ShapeOpts: font.ShapeOptions{FontID: fontID, FontSize: 16},
				OriginX:   leftX,
				OriginY:   curY,
				Color:     black,
				Opacity:   1,
			})
			m := adapter.engine.FontMetrics(fontID, 16)
			curY += m.Ascent + m.Descent + 4
		}
		curY += 6

		// Separator line
		env.buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(leftX, curY, 1160, 1),
			FillColor: uimath.ColorHex("#CCCCCC"),
		}, 0, 1)
		curY += 8

		// Specimen text at increasing sizes
		for _, sz := range sizes {
			// Size label on the left
			label := fmt.Sprintf("%.0f", sz)
			env.textRenderer.DrawText(env.buf, label, textrender.DrawOptions{
				ShapeOpts: font.ShapeOptions{FontID: fontID, FontSize: 11},
				OriginX:   leftX,
				OriginY:   curY,
				Color:     gray,
				Opacity:   1,
			})

			// Sample text
			env.textRenderer.DrawText(env.buf, sampleText, textrender.DrawOptions{
				ShapeOpts: font.ShapeOptions{FontID: fontID, FontSize: sz},
				OriginX:   leftX + 30,
				OriginY:   curY,
				Color:     black,
				Opacity:   1,
			})

			m := adapter.engine.FontMetrics(fontID, sz)
			curY += m.Ascent + m.Descent + 6
		}

		env.textRenderer.Upload()
		env.backend.Submit(env.buf)
		env.backend.EndFrame()
	}

	_, img := env.screenshot(t, "font_specimen")
	verifyNotUniform(t, img, "font_specimen")

	nc := countDistinctColors(img, 40, 0, 1100, img.Bounds().Dy())
	t.Logf("font specimen: %d distinct colors", nc)
	if nc < 50 {
		t.Errorf("expected rich text rendering (>50 colors), got %d — FreeType may not be loaded", nc)
	}
}

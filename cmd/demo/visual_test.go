package main

import (
	"fmt"
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
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/render/capture"
	"github.com/kasuganosora/ui/widget"
)

// testEnv holds a full platform+renderer test environment.
type testEnv struct {
	plat         platform.Platform
	win          platform.Window
	backend      render.Backend
	tree         *core.Tree
	cfg          *widget.Config
	buf          *render.CommandBuffer
	textRenderer *textrender.Renderer
	glyphAtlas   *atlas.Atlas
	fontEngine   font.Engine
	fontMgr      font.Manager
	fontID       font.ID
	width        int
	height       int
}

func newTestEnv(t *testing.T, width, height int) *testEnv {
	t.Helper()

	plat := ui.NewPlatform()
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
	} else if fe := ui.FallbackFontEngine(); fe != nil {
		fontEngine = fe
	} else {
		fontEngine = ui.NewMockEngine()
	}
	mgr := font.NewManager(fontEngine)
	fontID, _ := mgr.RegisterFile("Default", font.WeightRegular, font.StyleNormal, ui.DefaultFont())
	if fontID == font.InvalidFontID {
		fontID, _ = mgr.Register("Default", font.WeightRegular, font.StyleNormal, nil)
	}
	glyphAtlas := atlas.New(atlas.Options{Width: 1024, Height: 1024, Backend: backend})
	tr := textrender.New(textrender.Options{Manager: mgr, Atlas: glyphAtlas})

	cfg := widget.DefaultConfig()
	cfg.TextRenderer = ui.NewTextDrawer(tr, fontID, fontEngine)

	return &testEnv{
		plat:         plat,
		win:          win,
		backend:      backend,
		tree:         core.NewTree(),
		cfg:          cfg,
		buf:          render.NewCommandBuffer(),
		textRenderer: tr,
		glyphAtlas:   glyphAtlas,
		fontEngine:   fontEngine,
		fontMgr:      mgr,
		fontID:       fontID,
		width:        width,
		height:       height,
	}
}

// enableCaptionFont opts a single visual test into the §15 medium-weight
// caption face. Keeping it opt-in leaves the existing golden screenshots on
// the regular face (they are static regression baselines, not typography
// specimens).
func (e *testEnv) enableCaptionFont() font.ID {
	e.cfg.TextRenderer = ui.NewTextDrawer(e.textRenderer, e.fontID, e.fontEngine)
	return ui.ConfigureCaptionFont(e.fontMgr, ui.DefaultFont(), e.cfg.TextRenderer)
}

func (e *testEnv) close() {
	e.textRenderer.Destroy()
	e.backend.Destroy()
	e.win.Destroy()
	e.plat.Terminate()
}

func (e *testEnv) buildUI() widget.Widget {
	doc := ui.LoadHTMLDocument(e.tree, e.cfg, demoHTML, "")
	setupDemoWidgets(doc, e.tree, e.cfg)
	if len(doc.Root.Children()) > 0 {
		return doc.Root.Children()[0]
	}
	return doc.Root
}

// minControlColors is the distinct-color floor for a rendered control region.
// Measured on real Metal frames: the flattest button (disabled) has 39 colors
// and the flattest input (empty placeholder) has 35. One third of that floor
// keeps ~3x headroom while still failing a control that collapsed into a flat
// fill with a single 1px edge. The old threshold of 2 could never fail.
const minControlColors = 12

// assertSingleSection enforces the precondition every section-focused visual
// test relies on: navigation must have hidden all other sections. Without it
// the content column is ~10000pt tall and sections below the 800pt viewport
// silently fall outside the framebuffer. Checking the precondition catches a
// broken navigation path even for the first section, whose own pixels look
// perfectly fine either way.
func assertSingleSection(t *testing.T, doc *ui.Document, sectionID string) {
	t.Helper()
	target := doc.QueryByID(sectionID)
	if target == nil {
		t.Fatalf("section %q missing from demoHTML", sectionID)
	}
	if target.Style().Display == layout.DisplayNone {
		t.Errorf("section %q is hidden; navigation did not select it", sectionID)
	}
	for _, n := range nav {
		if n.section == sectionID {
			continue
		}
		w := doc.QueryByID(n.section)
		if w == nil {
			continue
		}
		if d := w.Style().Display; d != layout.DisplayNone {
			t.Errorf("section %q is still displayed (display=%d); only %q should be visible",
				n.section, d, sectionID)
		}
	}
}

// buildSectionUI loads the real demo document and switches it to a single
// section through setupNavigation — the exact path the running demo takes.
//
// buildUI() deliberately keeps every section visible, which makes the content
// column ~10000pt tall. Four visual tests used to call it and therefore saved
// the identical first screen under four different names; sections below the
// fold were never actually captured. Showing one section brings its controls
// into the viewport so their bounds are real and assertable.
func (e *testEnv) buildSectionUI(t *testing.T, sectionID string) (widget.Widget, *ui.Document) {
	t.Helper()
	doc := ui.LoadHTMLDocument(e.tree, e.cfg, demoHTML, "")
	setupNavigation(doc, sectionID)
	setupDemoWidgets(doc, e.tree, e.cfg)
	assertSingleSection(t, doc, sectionID)
	root := widget.Widget(doc.Root)
	if len(doc.Root.Children()) > 0 {
		root = doc.Root.Children()[0]
	}
	return root, doc
}

// renderFrames builds layout, draws, and submits N frames.
func (e *testEnv) renderFrames(root widget.Widget, n int) {
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
		widget.DrawPressFeedback(e.buf, e.tree, e.cfg)
		_ = e.textRenderer.Upload()
		e.backend.Submit(e.buf)
		e.backend.EndFrame()
	}
}

func saveScreenshot(t *testing.T, name string, img *image.RGBA) string {
	t.Helper()

	dir := filepath.Join("testdata", "screenshots")
	_ = os.MkdirAll(dir, 0755)

	path := filepath.Join(dir, name+".png")
	if err := capture.SavePNG(img, path); err != nil {
		t.Fatalf("save screenshot: %v", err)
	}
	t.Logf("screenshot saved: %s (%dx%d)", path, img.Bounds().Dx(), img.Bounds().Dy())
	return path
}

// screenshot captures the full framebuffer and saves it to testdata/.
func (e *testEnv) screenshot(t *testing.T, name string) (string, *image.RGBA) {
	t.Helper()

	img, err := capture.Screenshot(e.backend)
	if err != nil {
		t.Fatalf("screenshot failed: %v", err)
	}
	return saveScreenshot(t, name, img), img
}

// logicalRect converts logical UI coordinates into framebuffer pixels.
// Layout bounds are in points while Metal runs at 2x on Retina; every visual
// assertion has to go through this conversion or it silently samples the
// wrong quarter of the frame.
func (e *testEnv) logicalRect(r uimath.Rect) image.Rectangle {
	dpi := e.backend.DPIScale()
	return image.Rect(
		int(r.X*dpi),
		int(r.Y*dpi),
		int((r.X+r.Width)*dpi),
		int((r.Y+r.Height)*dpi),
	)
}

// cropLogical clips a framebuffer to a logical rectangle. The result is a
// fresh image with origin (0,0), so the pixel helpers below keep working on
// the cropped image without any coordinate translation.
func (e *testEnv) cropLogical(img *image.RGBA, r uimath.Rect) *image.RGBA {
	px := e.logicalRect(r).Intersect(img.Bounds())
	if px.Empty() {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}
	out := image.NewRGBA(image.Rect(0, 0, px.Dx(), px.Dy()))
	for y := 0; y < px.Dy(); y++ {
		for x := 0; x < px.Dx(); x++ {
			out.SetRGBA(x, y, img.RGBAAt(px.Min.X+x, px.Min.Y+y))
		}
	}
	return out
}

// screenshotSection captures only the section's own rectangle instead of the
// whole window, and returns its logical bounds as the coordinate origin for
// converting element bounds into cropped-image coordinates.
func (e *testEnv) screenshotSection(t *testing.T, name string, section widget.Widget) (string, *image.RGBA, uimath.Rect) {
	t.Helper()
	img, err := capture.Screenshot(e.backend)
	if err != nil {
		t.Fatalf("screenshot failed: %v", err)
	}
	elem := e.tree.Get(section.ElementID())
	if elem == nil {
		t.Fatalf("section %d has no tree element", section.ElementID())
	}
	origin := elem.Layout().Bounds
	if origin.Width < 10 || origin.Height < 10 {
		t.Fatalf("section %d has no layout bounds: %+v", section.ElementID(), origin)
	}

	// Cross-check the DPI used by logicalRect against the real framebuffer.
	// These come from independent sources (backend.DPIScale vs. window size),
	// so a wrong scale cannot quietly satisfy the crop and the pixel
	// assertions at the same time.
	dpiFromFrame := float64(img.Bounds().Dx()) / float64(e.width)
	if diff := float64(e.backend.DPIScale()) - dpiFromFrame; diff > 0.01 || diff < -0.01 {
		t.Fatalf("DPI mismatch: backend=%.2f framebuffer/window=%.2f",
			e.backend.DPIScale(), dpiFromFrame)
	}
	wantW := int(float64(origin.Width)*dpiFromFrame + 0.5)
	wantH := int(float64(origin.Height)*dpiFromFrame + 0.5)

	cropped := e.cropLogical(img, origin)
	if d := cropped.Bounds().Dx() - wantW; d > 2 || d < -2 {
		t.Errorf("cropped width %d, want %d (section %.0fpt @ %.1fx)",
			cropped.Bounds().Dx(), wantW, origin.Width, dpiFromFrame)
	}
	if d := cropped.Bounds().Dy() - wantH; d > 2 || d < -2 {
		t.Errorf("cropped height %d, want %d (section %.0fpt @ %.1fx)",
			cropped.Bounds().Dy(), wantH, origin.Height, dpiFromFrame)
	}
	return saveScreenshot(t, name, cropped), cropped, origin
}

// relativeRect re-bases an element's absolute logical bounds into the cropped
// section image's coordinate space.
func relativeRect(origin, b uimath.Rect) uimath.Rect {
	return uimath.NewRect(b.X-origin.X, b.Y-origin.Y, b.Width, b.Height)
}

// sectionPixels converts an absolute element bounds into a pixel rectangle
// inside a cropped section image. It returns false for elements that are not
// fully inside the section (hidden sections, off-screen controls, etc.).
func (e *testEnv) sectionPixels(origin, b uimath.Rect, img *image.RGBA) (image.Rectangle, bool) {
	rel := relativeRect(origin, b)
	if rel.X < 0 || rel.Y < 0 || rel.X+rel.Width > origin.Width || rel.Y+rel.Height > origin.Height {
		return image.Rectangle{}, false
	}
	px := e.logicalRect(rel)
	if px.Empty() || px.Min.X < 0 || px.Min.Y < 0 ||
		px.Max.X > img.Bounds().Dx() || px.Max.Y > img.Bounds().Dy() {
		return image.Rectangle{}, false
	}
	return px, true
}

// dominantColorKey returns the most frequent RGBA value in a pixel rectangle.
// It isolates a swatch's fill from its rounded border and centered label.
func dominantColorKey(img *image.RGBA, r image.Rectangle) uint32 {
	counts := make(map[uint32]int)
	best, bestN := uint32(0), -1
	bounds := img.Bounds().Intersect(r)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.RGBAAt(x, y)
			key := uint32(c.R)<<24 | uint32(c.G)<<16 | uint32(c.B)<<8 | uint32(c.A)
			counts[key]++
			if counts[key] > bestN {
				bestN, best = counts[key], key
			}
		}
	}
	return best
}

// --- Tree inspection helpers ---

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
func verifyNotUniform(t *testing.T, img *image.RGBA, label string) {
	t.Helper()
	bounds := img.Bounds()
	ref := img.RGBAAt(bounds.Min.X, bounds.Min.Y)
	// Sample every pixel: a coarse 20x20 grid misses thin text and separators
	// and reports a perfectly rendered frame as uniform.
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if img.RGBAAt(x, y) != ref {
				return
			}
		}
	}
	t.Errorf("%s: screenshot is uniform color (R=%d G=%d B=%d A=%d)", label, ref.R, ref.G, ref.B, ref.A)
}

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
func TestVisualDemoFullUI(t *testing.T) {
	env := newTestEnv(t, 1280, 800)
	defer env.close()

	root := env.buildUI()
	env.tree.AppendChild(env.tree.Root(), root.ElementID())

	env.renderFrames(root, 3)

	count := env.tree.ElementCount()
	t.Logf("element count: %d", count)
	if count < 20 {
		t.Errorf("expected at least 20 elements, got %d", count)
	}

	dumpTree(t, env.tree)
	verifyAllVisibleHaveBounds(t, env.tree, 2)

	_, img := env.screenshot(t, "demo_full_ui")
	verifyNotUniform(t, img, "demo_full_ui")

	w := img.Bounds().Dx()
	// 60 is a logical height (the demo header is 48pt tall); the framebuffer
	// is 2x on Retina, so the sampled band must be scaled too.
	top := int(60 * env.backend.DPIScale())
	topColors := countDistinctColors(img, 0, 0, w, top)
	t.Logf("distinct colors in top %dpx (60pt): %d", top, topColors)
	if topColors < 2 {
		t.Errorf("top region has only %d colors, expected header rendering", topColors)
	}
}

func TestVisualButtonRendering(t *testing.T) {
	env := newTestEnv(t, 1280, 800)
	defer env.close()

	root, doc := env.buildSectionUI(t, "sec-button")
	env.tree.AppendChild(env.tree.Root(), root.ElementID())
	env.renderFrames(root, 3)

	section := doc.QueryByID("sec-button")
	if section == nil {
		t.Fatal("sec-button missing from demoHTML")
	}
	_, img, origin := env.screenshotSection(t, "button_rendering", section)
	verifyNotUniform(t, img, "button_rendering")

	checked, minNC := 0, 1<<30
	env.tree.Walk(env.tree.Root(), func(id core.ElementID, depth int) bool {
		elem := env.tree.Get(id)
		if elem == nil || elem.Type() != core.TypeButton {
			return true
		}
		b := elem.Layout().Bounds
		if b.Width < 10 || b.Height < 10 {
			return true
		}
		px, ok := env.sectionPixels(origin, b, img)
		if !ok {
			return true // belongs to a hidden section or falls outside the crop
		}
		checked++
		nc := countDistinctColors(img, px.Min.X, px.Min.Y, px.Dx(), px.Dy())
		if nc < minNC {
			minNC = nc
		}
		if nc < minControlColors {
			t.Errorf("button %d %q rendered only %d color(s), want >= %d",
				id, elem.TextContent(), nc, minControlColors)
		}
		return true
	})
	// A log line is not an assertion: if navigation broke and no button was
	// visible at all, the old test still passed.
	if checked < 10 {
		t.Errorf("only %d buttons were checked in sec-button, expected at least 10", checked)
	}
	if minNC == 1<<30 {
		minNC = 0
	}
	t.Logf("checked %d buttons in sec-button; minimum distinct colors=%d (floor=%d)",
		checked, minNC, minControlColors)
}

func TestVisualInputRendering(t *testing.T) {
	env := newTestEnv(t, 1280, 800)
	defer env.close()

	root, doc := env.buildSectionUI(t, "sec-input")
	env.tree.AppendChild(env.tree.Root(), root.ElementID())
	env.renderFrames(root, 3)

	section := doc.QueryByID("sec-input")
	if section == nil {
		t.Fatal("sec-input missing from demoHTML")
	}
	_, img, origin := env.screenshotSection(t, "input_rendering", section)
	verifyNotUniform(t, img, "input_rendering")

	checked, minNC := 0, 1<<30
	env.tree.Walk(env.tree.Root(), func(id core.ElementID, depth int) bool {
		elem := env.tree.Get(id)
		if elem == nil || elem.Type() != core.TypeInput {
			return true
		}
		b := elem.Layout().Bounds
		if b.Width < 10 || b.Height < 10 {
			return true
		}
		px, ok := env.sectionPixels(origin, b, img)
		if !ok {
			return true
		}
		checked++
		// Coordinates are now DPI-scaled and rebased to the cropped section.
		verifyRegionNotBlack(t, img, fmt.Sprintf("input_%d", id),
			px.Min.X, px.Min.Y, px.Dx(), px.Dy())
		// A single bright pixel used to satisfy verifyRegionNotBlack; require
		// the full border/background/text antialiasing footprint instead.
		nc := countDistinctColors(img, px.Min.X, px.Min.Y, px.Dx(), px.Dy())
		if nc < minNC {
			minNC = nc
		}
		if nc < minControlColors {
			t.Errorf("input %d rendered only %d color(s), want >= %d", id, nc, minControlColors)
		}
		return true
	})
	if checked < 3 {
		t.Errorf("only %d inputs were checked in sec-input, expected at least 3", checked)
	}
	if minNC == 1<<30 {
		minNC = 0
	}
	t.Logf("checked %d inputs in sec-input; minimum distinct colors=%d (floor=%d)",
		checked, minNC, minControlColors)
}

func TestVisualGridColors(t *testing.T) {
	env := newTestEnv(t, 1280, 800)
	defer env.close()

	root, doc := env.buildSectionUI(t, "sec-grid")
	env.tree.AppendChild(env.tree.Root(), root.ElementID())
	env.renderFrames(root, 3)

	section := doc.QueryByID("sec-grid")
	if section == nil {
		t.Fatal("sec-grid missing from demoHTML")
	}
	_, img, origin := env.screenshotSection(t, "grid_colors", section)
	verifyNotUniform(t, img, "grid_colors")

	// "GridColors" should check the four column swatches, not the window's
	// top 60px (which only ever contained the header).
	seen := make(map[uint32]string)
	for _, cls := range []string{"grid-col-1", "grid-col-2", "grid-col-3", "grid-col-4"} {
		ws := doc.QueryByClass(cls)
		if len(ws) == 0 {
			t.Errorf("no widget with class %q in sec-grid", cls)
			continue
		}
		b := env.tree.Get(ws[0].ElementID()).Layout().Bounds
		px, ok := env.sectionPixels(origin, b, img)
		if !ok {
			t.Errorf("swatch %q at (%.0f,%.0f %.0fx%.0f) is outside sec-grid", cls, b.X, b.Y, b.Width, b.Height)
			continue
		}
		// Inset past the rounded border and the centered "col-N" label.
		key := dominantColorKey(img, px.Inset(4))
		if prev, dup := seen[key]; dup {
			t.Errorf("swatch %q has the same dominant color as %q (#%06X)", cls, prev, key>>8)
		}
		seen[key] = cls
		t.Logf("swatch %q dominant color #%06X", cls, key>>8)
	}
	if len(seen) < 4 {
		t.Errorf("only %d distinct grid swatch colors, expected 4", len(seen))
	}
	nc := countDistinctColors(img, 0, 0, img.Bounds().Dx(), img.Bounds().Dy())
	if nc < 4 {
		t.Errorf("grid section has only %d colors, expected a rendered grid", nc)
	}
}

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

	// AppKit requires nextEventMatchingMask (PollEvents) to run on the main
	// thread, so the loop stays on the test goroutine instead of a spawned one.
	// A deadline still guards against a stalled frame; go test's own timeout is
	// the final backstop if a single call blocks entirely.
	start := time.Now()
	for i := 0; i < 10; i++ {
		env.plat.PollEvents()
		env.backend.BeginFrame()
		env.textRenderer.BeginFrame()
		env.buf.Reset()
		root.Draw(env.buf)
		_ = env.textRenderer.Upload()
		env.backend.Submit(env.buf)
		env.backend.EndFrame()
		if time.Since(start) > 10*time.Second {
			t.Fatal("message loop deadlock")
		}
	}

	_, img := env.screenshot(t, "message_loop")
	verifyNotUniform(t, img, "message_loop")
}

func TestVisualHitTestConsistency(t *testing.T) {
	env := newTestEnv(t, 1280, 800)
	defer env.close()

	root := env.buildUI()
	env.tree.AppendChild(env.tree.Root(), root.ElementID())

	w, h := float32(env.width), float32(env.height)
	env.tree.SetLayout(env.tree.Root(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, w, h)})
	ui.AutoLayout(env.tree, root, w, h)

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
				t.Errorf("HitTest(%.0f, %.0f) invalid for element %d (%s) at depth %d",
					cx, cy, id, elem.Type(), depth)
			}
		}
		return true
	})
	t.Logf("hit test: tested %d elements, %d missed", tested, missed)
}

func TestVisualCommandBufferCoverage(t *testing.T) {
	env := newTestEnv(t, 1280, 800)
	defer env.close()

	root := env.buildUI()
	env.tree.AppendChild(env.tree.Root(), root.ElementID())

	w, h := float32(env.width), float32(env.height)
	env.tree.SetLayout(env.tree.Root(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, w, h)})
	ui.AutoLayout(env.tree, root, w, h)

	env.buf.Reset()
	root.Draw(env.buf)

	cmdCount := env.buf.Len()
	t.Logf("command buffer has %d commands", cmdCount)

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
	if cmdCount < visibleCount/2 {
		t.Errorf("too few render commands (%d) for %d visible elements", cmdCount, visibleCount)
	}
}

func TestVisualFrameConsistency(t *testing.T) {
	env := newTestEnv(t, 400, 300)
	defer env.close()

	root := env.buildUI()
	env.tree.AppendChild(env.tree.Root(), root.ElementID())

	env.renderFrames(root, 1) // warm-up

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

	result, err := capture.Compare(img1, img2, 0.01)
	if err != nil {
		t.Fatalf("compare: %v", err)
	}
	if result.DiffPixels > 0 {
		pct := float64(result.DiffPixels) / float64(result.TotalPixels) * 100
		t.Errorf("consecutive frames differ: %d pixels (%.2f%%)", result.DiffPixels, pct)
	}
}

func TestVisualFontSpecimen(t *testing.T) {
	// This measures real glyph antialiasing. Both the native engine and the
	// pure Go outline fallback produce it, so the old FreeType-only skip is
	// no longer needed.
	env := newTestEnv(t, 1200, 900)
	defer env.close()

	sampleText := "Innovation in China 中国智造，慧及全球 0123456789"
	sizes := []float32{12, 18, 24, 36, 48, 60, 72}

	for frame := 0; frame < 2; frame++ {
		env.plat.PollEvents()
		env.backend.BeginFrame()
		env.textRenderer.BeginFrame()
		env.buf.Reset()

		env.buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(0, 0, 1200, 900),
			FillColor: uimath.ColorWhite,
		}, 0, 1)

		curY := float32(12)
		black := uimath.ColorHex("#000000")
		gray := uimath.ColorHex("#999999")
		leftX := float32(20)

		headerTexts := []string{
			"字体名称: 微软雅黑",
			"版本: Version 6.31",
			"OpenType Layout, TrueType Outlines",
		}
		for _, ht := range headerTexts {
			env.textRenderer.DrawText(env.buf, ht, textrender.DrawOptions{
				ShapeOpts: font.ShapeOptions{FontID: env.fontID, FontSize: 14},
				OriginX:   leftX,
				OriginY:   curY,
				Color:     black,
				Opacity:   1,
			})
			m := env.fontEngine.FontMetrics(env.fontID, 14)
			curY += m.Ascent + m.Descent + 2
		}
		curY += 4

		charTexts := []string{
			"abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ",
			`1234567890.:;' " (!?) +-*/=`,
		}
		for _, ct := range charTexts {
			env.textRenderer.DrawText(env.buf, ct, textrender.DrawOptions{
				ShapeOpts: font.ShapeOptions{FontID: env.fontID, FontSize: 16},
				OriginX:   leftX,
				OriginY:   curY,
				Color:     black,
				Opacity:   1,
			})
			m := env.fontEngine.FontMetrics(env.fontID, 16)
			curY += m.Ascent + m.Descent + 4
		}
		curY += 6

		env.buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(leftX, curY, 1160, 1),
			FillColor: uimath.ColorHex("#CCCCCC"),
		}, 0, 1)
		curY += 8

		for _, sz := range sizes {
			label := fmt.Sprintf("%.0f", sz)
			env.textRenderer.DrawText(env.buf, label, textrender.DrawOptions{
				ShapeOpts: font.ShapeOptions{FontID: env.fontID, FontSize: 11},
				OriginX:   leftX,
				OriginY:   curY,
				Color:     gray,
				Opacity:   1,
			})

			env.textRenderer.DrawText(env.buf, sampleText, textrender.DrawOptions{
				ShapeOpts: font.ShapeOptions{FontID: env.fontID, FontSize: sz},
				OriginX:   leftX + 30,
				OriginY:   curY,
				Color:     black,
				Opacity:   1,
			})

			m := env.fontEngine.FontMetrics(env.fontID, sz)
			curY += m.Ascent + m.Descent + 6
		}

		_ = env.textRenderer.Upload()
		env.backend.Submit(env.buf)
		env.backend.EndFrame()
	}

	_, img := env.screenshot(t, "font_specimen")
	verifyNotUniform(t, img, "font_specimen")

	nc := countDistinctColors(img, 40, 0, 1100, img.Bounds().Dy())
	t.Logf("font specimen: %d distinct colors", nc)
	if nc < 50 {
		t.Errorf("expected rich text rendering (>50 colors), got %d", nc)
	}
}

func TestVisualListWidget(t *testing.T) {
	env := newTestEnv(t, 600, 300)
	defer env.close()

	tree := env.tree
	cfg := env.cfg

	l := widget.NewList(tree, cfg)
	l.SetItems([]widget.ListItem{
		{Title: "列表主内容", Description: "列表内容列表内容列表内容", Actions: []string{"操作1", "操作2", "操作3"}},
		{Title: "列表主内容", Description: "列表内容列表内容列表内容", Actions: []string{"操作1", "操作2", "操作3"}},
		{Title: "列表主内容", Description: "列表内容列表内容列表内容", Actions: []string{"操作1", "操作2", "操作3"}},
	})

	tree.AppendChild(tree.Root(), l.ElementID())

	w, h := float32(env.width), float32(env.height)
	tree.SetLayout(tree.Root(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, w, h)})

	totalH := l.TotalHeight()
	t.Logf("List TotalHeight: %.0f", totalH)
	if totalH < 180 {
		t.Errorf("expected List TotalHeight >= 180, got %.0f", totalH)
	}
	tree.SetLayout(l.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(10, 10, w-20, totalH)})

	env.plat.PollEvents()
	env.backend.BeginFrame()
	env.textRenderer.BeginFrame()
	env.buf.Reset()
	l.Draw(env.buf)
	_ = env.textRenderer.Upload()
	env.backend.Submit(env.buf)
	env.backend.EndFrame()
	_, img := env.screenshot(t, "list_widget")
	verifyNotUniform(t, img, "list_widget")

	thirdH := img.Bounds().Dy() / 3
	topColors := countDistinctColors(img, 10, 10, img.Bounds().Dx()-20, thirdH)
	midColors := countDistinctColors(img, 10, thirdH, img.Bounds().Dx()-20, thirdH)
	botColors := countDistinctColors(img, 10, thirdH*2, img.Bounds().Dx()-20, thirdH)
	t.Logf("List row colors: top=%d mid=%d bot=%d", topColors, midColors, botColors)
	if topColors < 3 {
		t.Errorf("top third has only %d colors, expected list item rendering", topColors)
	}
	if midColors < 3 {
		t.Errorf("middle third has only %d colors, expected list item rendering", midColors)
	}
	if botColors < 2 {
		t.Errorf("bottom third has only %d colors, expected list item rendering", botColors)
	}
}

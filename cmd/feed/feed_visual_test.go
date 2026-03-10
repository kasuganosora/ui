//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"testing"

	ui "github.com/kasuganosora/ui"
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

// feedTestEnv holds a full platform+renderer test environment for feed tests.
type feedTestEnv struct {
	plat         *win32.Platform
	win          platform.Window
	backend      *vulkan.Backend
	tree         *core.Tree
	cfg          *widget.Config
	buf          *render.CommandBuffer
	textRenderer *textrender.Renderer
	width        int
	height       int
}

func newFeedTestEnv(t *testing.T, width, height int) *feedTestEnv {
	t.Helper()

	plat := win32.New()
	if err := plat.Init(); err != nil {
		t.Fatalf("platform init: %v", err)
	}

	win, err := plat.CreateWindow(platform.WindowOptions{
		Title:     "Feed Visual Test",
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
	fw, fh := win.FramebufferSize()
	backend.Resize(fw, fh)

	var fontEngine font.Engine
	if ftEngine, err := freetype.New(); err == nil {
		fontEngine = ftEngine
	} else {
		fontEngine = ui.NewMockEngine()
	}
	fontMgr := font.NewManager(fontEngine)
	fontID, _ := fontMgr.RegisterFile("Default", font.WeightRegular, font.StyleNormal, `C:\Windows\Fonts\msyh.ttc`)
	if fontID == font.InvalidFontID {
		fontID, _ = fontMgr.Register("Default", font.WeightRegular, font.StyleNormal, nil)
	}
	// Register symbol fallback fonts for glyphs missing from the primary font.
	var fallbackIDs []font.ID
	for _, fbPath := range []string{`C:\Windows\Fonts\seguisym.ttf`} {
		if fbID, err := fontMgr.RegisterFile("Symbol", font.WeightRegular, font.StyleNormal, fbPath); err == nil && fbID != font.InvalidFontID {
			fallbackIDs = append(fallbackIDs, fbID)
		}
	}
	glyphAtlas := atlas.New(atlas.Options{Width: 1024, Height: 1024, Backend: backend})
	tr := textrender.New(textrender.Options{Manager: fontMgr, Atlas: glyphAtlas})

	cfg := widget.DefaultConfig()
	cfg.TextRenderer = ui.NewTextDrawer(tr, fontID, fontEngine, fallbackIDs...)
	cfg.Backend = backend

	return &feedTestEnv{
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

func (e *feedTestEnv) close() {
	e.textRenderer.Destroy()
	e.backend.Destroy()
	e.win.Destroy()
	e.plat.Terminate()
}

func (e *feedTestEnv) renderFrames(root widget.Widget, n int) {
	w, h := float32(e.width), float32(e.height)
	for i := 0; i < n; i++ {
		e.plat.PollEvents()
		e.tree.SetLayout(e.tree.Root(), core.LayoutResult{
			Bounds: uimath.NewRect(0, 0, w, h),
		})
		ui.CSSLayout(e.tree, root, w, h, e.cfg)

		e.backend.BeginFrame()
		e.textRenderer.BeginFrame()
		e.buf.Reset()
		root.Draw(e.buf)
		e.textRenderer.Upload()
		e.backend.Submit(e.buf)
		e.backend.EndFrame()
	}
}

func (e *feedTestEnv) screenshot(t *testing.T, name string) (string, *image.RGBA) {
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

func dumpFeedTree(t *testing.T, tree *core.Tree) {
	t.Helper()
	count := 0
	tree.Walk(tree.Root(), func(id core.ElementID, depth int) bool {
		if depth > 8 {
			return false
		}
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
		count++
		return true
	})
	t.Logf("total elements: %d", count)
}

// dumpWidgetTree recursively dumps widget bounds from the widget.Widget hierarchy.
func dumpWidgetTree(t *testing.T, tree *core.Tree, w widget.Widget, depth int, maxDepth int) {
	t.Helper()
	if depth > maxDepth {
		return
	}
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	elem := tree.Get(w.ElementID())
	typeStr := "?"
	textStr := ""
	if elem != nil {
		typeStr = string(elem.Type())
		if tc := elem.TextContent(); tc != "" && len(tc) < 40 {
			textStr = fmt.Sprintf(" %q", tc)
		}
	}
	b := uimath.Rect{}
	if elem != nil {
		b = elem.Layout().Bounds
	}
	t.Logf("%s[%s] (%.0f,%.0f %.0fx%.0f)%s", indent, typeStr, b.X, b.Y, b.Width, b.Height, textStr)
	for _, c := range w.Children() {
		dumpWidgetTree(t, tree, c, depth+1, maxDepth)
	}
}

// countNonBlackPixels counts pixels that are not black (> threshold in any channel).
func countNonBlackPixels(img *image.RGBA, threshold uint8) int {
	count := 0
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			c := img.RGBAAt(x, y)
			if c.R > threshold || c.G > threshold || c.B > threshold {
				count++
			}
		}
	}
	return count
}

// countDistinctColorsInRegion counts unique colors in a region.
func countDistinctColorsInRegion(img *image.RGBA, x, y, w, h int) int {
	seen := make(map[[3]uint8]bool)
	for py := y; py < y+h && py < img.Bounds().Max.Y; py++ {
		for px := x; px < x+w && px < img.Bounds().Max.X; px++ {
			c := img.RGBAAt(px, py)
			seen[[3]uint8{c.R, c.G, c.B}] = true
		}
	}
	return len(seen)
}

// TestVisualFeedTimeline renders the feed UI and captures a screenshot for visual inspection.
func TestVisualFeedTimeline(t *testing.T) {
	// Load tweet data
	if err := json.Unmarshal(tweetsJSON, &allTweets); err != nil {
		t.Fatalf("parse tweets: %v", err)
	}

	env := newFeedTestEnv(t, 600, 860)
	defer env.close()

	// Build feed UI using the same HTML/CSS as main.go
	doc := ui.LoadHTMLDocument(env.tree, env.cfg, feedHTML+"<style>"+feedCSS+"</style>", "")
	var root widget.Widget = doc.Root
	if len(doc.Root.Children()) > 0 {
		root = doc.Root.Children()[0]
	}

	// Find the timeline container
	timeline := doc.QueryByID("timeline")
	if timeline == nil {
		t.Fatal("timeline element not found")
	}

	type feedContainer interface {
		widget.Widget
		AppendChild(widget.Widget)
	}
	container := timeline.(feedContainer)

	// Seed 10 tweets
	count := min(10, len(allTweets))
	for i := range count {
		var rtBy string
		if i > 0 && i%5 == 0 {
			rtBy = allTweets[(i+1)%len(allTweets)].Name
		}
		td := &allTweets[i]
		timeStr := parseCreatedAt(td.Created)
		html := tweetHTML(td, timeStr, rtBy)
		tw := ui.LoadHTMLWithCSS(env.tree, env.cfg, html, feedCSS)
		if len(tw.Children()) > 0 {
			container.AppendChild(tw.Children()[0])
		} else {
			container.AppendChild(tw)
		}
	}

	// Debug: check what style the root has
	// Debug: check what style the root has
	rs := root.Style()
	t.Logf("=== Root style: display=%v W=%v H=%v ===", rs.Display, rs.Width, rs.Height)
	if len(root.Children()) > 0 {
		hs := root.Children()[0].Style()
		t.Logf("=== Header style: display=%v W=%v H=%v ===", hs.Display, hs.Width, hs.Height)
		if len(root.Children()[0].Children()) > 0 {
			ts := root.Children()[0].Children()[0].Style()
			t.Logf("=== Tabs style: display=%v W=%v H=%v ===", ts.Display, ts.Width, ts.Height)
		}
	}

	// Render 3 frames to stabilize
	env.renderFrames(root, 3)

	// Capture screenshot
	_, img := env.screenshot(t, "feed_timeline")

	// Dump widget tree (first 4 levels) for debugging bounds
	t.Logf("=== Widget Tree (first 4 levels) ===")
	dumpWidgetTree(t, env.tree, root, 0, 4)
	// Dump core tree for element IDs
	dumpFeedTree(t, env.tree)

	// Visual assertions
	totalPixels := img.Bounds().Dx() * img.Bounds().Dy()
	nonBlack := countNonBlackPixels(img, 10)
	nonBlackPct := float64(nonBlack) / float64(totalPixels) * 100
	t.Logf("non-black pixels: %d / %d (%.1f%%)", nonBlack, totalPixels, nonBlackPct)

	// Expect significant non-black content (header, tweets, text)
	if nonBlackPct < 5 {
		t.Errorf("too few non-black pixels (%.1f%%): feed may not be rendering", nonBlackPct)
	}

	// Check header region has content (tab bar at top ~53px)
	headerColors := countDistinctColorsInRegion(img, 0, 0, 600, 53)
	t.Logf("header distinct colors: %d", headerColors)
	if headerColors < 3 {
		t.Errorf("header region has too few colors (%d): tab bar not rendering", headerColors)
	}

	// Check timeline area has tweet cards (below header ~150px from top)
	timelineColors := countDistinctColorsInRegion(img, 0, 150, 600, 400)
	t.Logf("timeline distinct colors: %d", timelineColors)
	if timelineColors < 5 {
		t.Errorf("timeline region has too few colors (%d): tweets not rendering", timelineColors)
	}
}

// TestVisualFeedTweetCard tests a single tweet card renders correctly.
func TestVisualFeedTweetCard(t *testing.T) {
	if err := json.Unmarshal(tweetsJSON, &allTweets); err != nil {
		t.Fatalf("parse tweets: %v", err)
	}

	env := newFeedTestEnv(t, 600, 200)
	defer env.close()

	// Render a single tweet card
	td := &allTweets[0]
	html := tweetHTML(td, "2分", "")
	root := ui.LoadHTMLWithCSS(env.tree, env.cfg, html, feedCSS)

	env.renderFrames(root, 3)
	_, img := env.screenshot(t, "feed_tweet_card")

	dumpFeedTree(t, env.tree)

	// Check that the tweet card has non-trivial content
	nonBlack := countNonBlackPixels(img, 10)
	totalPixels := img.Bounds().Dx() * img.Bounds().Dy()
	pct := float64(nonBlack) / float64(totalPixels) * 100
	t.Logf("tweet card non-black: %.1f%%", pct)

	// Avatar color check: left ~40px should have the avatar color
	avatarColors := countDistinctColorsInRegion(img, 16, 12, 40, 40)
	t.Logf("avatar region colors: %d", avatarColors)

	// Text should be present
	if pct < 1 {
		t.Errorf("tweet card renders as blank (%.1f%% non-black)", pct)
	}

	// Check distinct colors in the name area (top right, after avatar)
	nameColors := countDistinctColorsInRegion(img, 68, 12, 400, 20)
	t.Logf("name row colors: %d", nameColors)
	if nameColors < 2 {
		t.Errorf("tweet name/meta not rendering (only %d colors in name area)", nameColors)
	}
}

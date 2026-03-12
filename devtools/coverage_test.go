// coverage_test.go contains targeted tests to push coverage above 95%.
package devtools

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/font"
	"github.com/kasuganosora/ui/font/atlas"
	"github.com/kasuganosora/ui/font/textrender"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// ─── Start / Stop ─────────────────────────────────────────────────────────────

func TestStart_Stop(t *testing.T) {
	srv := NewServer(Options{Addr: "127.0.0.1:0"})
	// Swap the router's server so we can use a random port.
	// We use an httptest server that binds on :0 to get a free port.
	ts := httptest.NewServer(srv.router)
	defer ts.Close()

	// Test Stop path when httpSrv is set (set it manually)
	srv.httpSrv = &http.Server{}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	// Shutdown of an unstarted Server.Shutdown should succeed
	_ = srv.httpSrv.Shutdown(ctx) // pre-warm
	err := srv.Stop(ctx)
	// Shutdown of already closed server may return error; just should not panic.
	_ = err
}

func TestStop_NilHttpSrv(t *testing.T) {
	srv := NewServer(Options{})
	// Stop with nil httpSrv is a no-op (returns nil)
	if err := srv.Stop(nil); err != nil {
		t.Errorf("Stop with nil httpSrv should return nil, got %v", err)
	}
}

// ─── jsonList / jsonVersion with empty Host header ────────────────────────────

func TestHTTP_JsonList_EmptyHost(t *testing.T) {
	srv := NewServer(Options{Addr: "127.0.0.1:9222", AppName: "TestHost"})
	ts := httptest.NewServer(srv.router)
	defer ts.Close()

	// Make request with explicit empty Host to exercise fallback
	req, _ := http.NewRequest("GET", ts.URL+"/json/list", nil)
	req.Host = "" // force empty host
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
}

func TestHTTP_JsonVersion_EmptyHost(t *testing.T) {
	srv := NewServer(Options{Addr: "127.0.0.1:9222"})
	ts := httptest.NewServer(srv.router)
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL+"/json/version", nil)
	req.Host = ""
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
}

// ─── declaredStyleProps – missing branches ───────────────────────────────────

func TestDeclaredStyleProps_WithAllConstraints(t *testing.T) {
	st := layout.Style{
		Width:     pxVal(100),
		Height:    pxVal(200),
		MinWidth:  pxVal(50),
		MinHeight: pxVal(30),
		MaxWidth:  pxVal(300),
		MaxHeight: pxVal(400),
		FlexBasis: pxVal(80),
		FlexGrow:  2,
		FlexShrink: 1,
		Overflow:  layout.OverflowScroll,
		FontSize:  16,
		WhiteSpace: layout.WhiteSpacePre,
		TextOverflow: layout.TextOverflowEllipsis,
	}
	props := declaredStyleProps(st)
	names := make(map[string]string)
	for _, p := range props {
		names[p.Name] = p.Value
	}
	checks := map[string]string{
		"width":        "100.00px",
		"height":       "200.00px",
		"min-width":    "50.00px",
		"max-height":   "400.00px",
		"flex-basis":   "80.00px",
		"overflow":     "scroll",
		"font-size":    "16.00px",
		"white-space":  "pre",
		"text-overflow": "ellipsis",
	}
	for k, want := range checks {
		if got := names[k]; got != want {
			t.Errorf("%s: want %q, got %q", k, want, got)
		}
	}
}

func TestDeclaredStyleProps_UniformBorder(t *testing.T) {
	v := pxVal(2)
	st := layout.Style{
		Border: layout.EdgeValues{Top: v, Right: v, Bottom: v, Left: v},
	}
	props := declaredStyleProps(st)
	names := make(map[string]string)
	for _, p := range props {
		names[p.Name] = p.Value
	}
	if names["border-width"] != "2.00px" {
		t.Errorf("uniform border: got %q (all: %v)", names["border-width"], names)
	}
}

func TestDeclaredStyleProps_PercentWidth(t *testing.T) {
	st := layout.Style{Width: pctVal(50)}
	props := declaredStyleProps(st)
	names := make(map[string]string)
	for _, p := range props {
		names[p.Name] = p.Value
	}
	if names["width"] != "50%" {
		t.Errorf("percent width: got %q", names["width"])
	}
}

func TestDeclaredStyleProps_FlexWithGap(t *testing.T) {
	st := layout.Style{
		Display:        layout.DisplayFlex,
		FlexDirection:  layout.FlexDirectionRowReverse,
		FlexWrap:       layout.FlexWrapWrapReverse,
		JustifyContent: layout.JustifySpaceEvenly,
		AlignItems:     layout.AlignBaseline,
		Gap:            12,
	}
	props := declaredStyleProps(st)
	names := make(map[string]string)
	for _, p := range props {
		names[p.Name] = p.Value
	}
	if names["gap"] != "12.00px" {
		t.Errorf("gap: got %q", names["gap"])
	}
	if names["flex-direction"] != "row-reverse" {
		t.Errorf("flex-direction: got %q", names["flex-direction"])
	}
}

// ─── walkSnap with early return ───────────────────────────────────────────────

func TestWalkSnap_EarlyReturn(t *testing.T) {
	snap := makeSnap(800, 600)
	snap.Root = 1
	p := addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.Rect{})
	addNode(snap, 2, 1, "span", nil, "", uimath.Rect{})
	p.ChildIDs = []core.ElementID{2}

	visited := 0
	walkSnap(snap, 1, func(id core.ElementID) bool {
		visited++
		return false // stop after first node
	})
	if visited != 1 {
		t.Errorf("early return should stop after 1 node, visited %d", visited)
	}
}

// ─── collectTextContent with mixed text + children ───────────────────────────

func TestCollectTextContent_NodeWithOwnTextAndChildren(t *testing.T) {
	snap := makeSnap(800, 600)
	p := addNode(snap, 1, core.InvalidElementID, "div", nil, "prefix:", uimath.Rect{})
	addNode(snap, 2, 1, "span", nil, "child", uimath.Rect{})
	p.ChildIDs = []core.ElementID{2}

	got := collectTextContent(snap, 1)
	if !strings.Contains(got, "prefix:") || !strings.Contains(got, "child") {
		t.Errorf("want 'prefix:child', got %q", got)
	}
}

// ─── handleDOM missing paths ──────────────────────────────────────────────────

func TestDOM_GetBoxModel_NodeNotFound(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.snapMu.Lock()
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.Rect{})
	tp.srv.snapshot = snap
	tp.srv.snapMu.Unlock()

	m := tp.send(t, 1, "DOM.getBoxModel", map[string]any{"nodeId": 999})
	checkError(t, m)
}

func TestDOM_DescribeNode_NodeNotFound(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.snapMu.Lock()
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.Rect{})
	tp.srv.snapshot = snap
	tp.srv.snapMu.Unlock()

	m := tp.send(t, 1, "DOM.describeNode", map[string]any{"nodeId": 999})
	checkError(t, m)
}

func TestDOM_GetNodeForLocation_NilSnapshot(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "DOM.getNodeForLocation", map[string]any{"x": 10, "y": 10})
	checkError(t, m)
}

func TestDOM_QuerySelector_NilSnapshot(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "DOM.querySelector", map[string]any{"nodeId": 1, "selector": "*"})
	checkResult(t, m)
	result := m["result"].(map[string]any)
	if result["nodeId"] != float64(0) {
		t.Errorf("nil snapshot: want nodeId=0, got %v", result["nodeId"])
	}
}

func TestDOM_QuerySelectorAll_NilSnapshot(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "DOM.querySelectorAll", map[string]any{"nodeId": 1, "selector": "*"})
	checkResult(t, m)
	result := m["result"].(map[string]any)
	ids := result["nodeIds"].([]any)
	if len(ids) != 0 {
		t.Errorf("nil snapshot: want empty nodeIds, got %v", ids)
	}
}

func TestDOM_RequestChildNodes_DefaultDepth(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.snapMu.Lock()
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.Rect{})
	tp.srv.snapshot = snap
	tp.srv.snapMu.Unlock()

	// depth=0 should default to 1
	msgs := tp.sendN(t, 1, "DOM.requestChildNodes", map[string]any{"nodeId": 1, "depth": 0}, 2)
	found := false
	for _, m := range msgs {
		if m["method"] == "DOM.setChildNodes" {
			found = true
		}
	}
	if !found {
		t.Error("expected DOM.setChildNodes")
	}
}

// ─── handleCSS missing paths ──────────────────────────────────────────────────

func TestCSS_GetComputedStyle_NodeNotFound(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.snapMu.Lock()
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.Rect{})
	tp.srv.snapshot = snap
	tp.srv.snapMu.Unlock()

	m := tp.send(t, 1, "CSS.getComputedStyleForNode", map[string]any{"nodeId": 999})
	checkResult(t, m)
}

func TestCSS_GetMatchedStyles_NodeNotFound(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.snapMu.Lock()
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.Rect{})
	tp.srv.snapshot = snap
	tp.srv.snapMu.Unlock()

	m := tp.send(t, 1, "CSS.getMatchedStylesForNode", map[string]any{"nodeId": 999})
	checkResult(t, m)
}

func TestCSS_GetInlineStyles_NodeNotFound(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.snapMu.Lock()
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.Rect{})
	tp.srv.snapshot = snap
	tp.srv.snapMu.Unlock()

	m := tp.send(t, 1, "CSS.getInlineStylesForNode", map[string]any{"nodeId": 999})
	checkResult(t, m)
}

// ─── buildSnapshotNode – nil property values ─────────────────────────────────

func TestBuildSnapshotNode_NoHtmlTagProperty(t *testing.T) {
	tree := core.NewTree()
	root := widget.NewDiv(tree, nil)
	// No "html-tag" property set
	snap := buildSnapshot(tree, root, 800, 600)
	node := snap.Nodes[root.ElementID()]
	if node.HTMLTag != "" {
		t.Errorf("want empty HTMLTag, got %q", node.HTMLTag)
	}
}

func TestBuildSnapshotNode_NonStringProperties(t *testing.T) {
	tree := core.NewTree()
	root := widget.NewDiv(tree, nil)
	// Set non-string values for html-tag and id (should not panic, should be ignored)
	tree.SetProperty(root.ElementID(), "html-tag", 42)
	tree.SetProperty(root.ElementID(), "id", true)
	snap := buildSnapshot(tree, root, 800, 600)
	node := snap.Nodes[root.ElementID()]
	if node.HTMLTag != "" {
		t.Errorf("non-string html-tag should be empty, got %q", node.HTMLTag)
	}
	if node.IDAttr != "" {
		t.Errorf("non-string id should be empty, got %q", node.IDAttr)
	}
}

// ─── querySelectorAll with nil node ──────────────────────────────────────────

func TestQuerySelectorAll_NilNode(t *testing.T) {
	snap := makeSnap(800, 600)
	// Root ID has no node in map
	snap.Root = 1
	// walkSnap will try to get node 1 → ok, but ChildIDs has id 99 → nil
	p := &NodeSnapshot{ID: 1, ChildIDs: []core.ElementID{99}}
	snap.Nodes[1] = p

	// Should not panic
	ids := querySelectorAll(snap, 1, "*")
	if len(ids) < 1 {
		t.Errorf("want at least root, got %d", len(ids))
	}
}

// ─── buildOuterHTML – indentation ────────────────────────────────────────────

func TestBuildOuterHTML_Indented(t *testing.T) {
	snap := makeSnap(800, 600)
	snap.Root = 1
	p := addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.Rect{})
	addNode(snap, 2, 1, "span", nil, "text", uimath.Rect{})
	p.ChildIDs = []core.ElementID{2}

	html := buildOuterHTML(snap, 1, 1) // indent=1 → 2-space pad
	if !strings.HasPrefix(html, "  <div>") {
		t.Errorf("indent=1 should add 2-space prefix: %q", html[:min(30, len(html))])
	}
}

// ─── Page.getLayoutMetrics with nil snapshot ─────────────────────────────────

func TestPage_GetLayoutMetrics_NilSnapshot(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "Page.getLayoutMetrics", nil)
	checkResult(t, m)
	result := m["result"].(map[string]any)
	lv := result["layoutViewport"].(map[string]any)
	// Default size = 1280×720
	if lv["clientWidth"] != float64(1280) {
		t.Errorf("want 1280, got %v", lv["clientWidth"])
	}
}

// ─── DrawOverlay – zero-size highlight ───────────────────────────────────────

func TestDrawOverlay_ZeroSizeBounds(t *testing.T) {
	srv := NewServer(Options{})
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.NewRect(10, 20, 0, 0))

	srv.snapMu.Lock()
	srv.snapshot = snap
	srv.snapMu.Unlock()
	srv.overlayMu.Lock()
	srv.highlightID = 1
	srv.overlayMu.Unlock()

	buf := render.NewCommandBuffer()
	srv.DrawOverlay(buf) // zero bounds → should not panic, no rects
}

// ─── Runtime.callFunctionOn – invalid node ID format ─────────────────────────

func TestRuntime_CallFunctionOn_InvalidNodeFormat(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "Runtime.callFunctionOn", map[string]any{
		"objectId":            "node-abc", // not a valid integer
		"functionDeclaration": "function() { return this.outerHTML; }",
	})
	result := m["result"].(map[string]any)
	r := result["result"].(map[string]any)
	if r["type"] != "undefined" {
		t.Errorf("invalid node format: want undefined, got %v", r)
	}
}

// ─── Start / Stop lifecycle ───────────────────────────────────────────────────

func TestStart_Stop_Lifecycle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	srv := NewServer(Options{Addr: "127.0.0.1:0"})
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Start() }()

	// Wait until httpSrv is set.
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if srv.httpSrv != nil {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	if err := srv.Stop(ctx); err != nil {
		t.Errorf("Stop: %v", err)
	}
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Start returned: %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("Start goroutine did not exit")
	}
}

// ─── jsonList / jsonVersion direct handler call with empty Host ──────────────

func TestJsonList_EmptyHostDirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	srv := NewServer(Options{Addr: "127.0.0.1:9222", AppName: "direct"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/json/list", nil)
	req.Host = ""
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	srv.jsonList(c)
	if w.Code != http.StatusOK {
		t.Errorf("want 200, got %d", w.Code)
	}
}

func TestJsonVersion_EmptyHostDirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	srv := NewServer(Options{Addr: "127.0.0.1:9222"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/json/version", nil)
	req.Host = ""
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	srv.jsonVersion(c)
	if w.Code != http.StatusOK {
		t.Errorf("want 200, got %d", w.Code)
	}
}

// ─── session.send marshal error path ─────────────────────────────────────────

func TestSession_Send_MarshalError(t *testing.T) {
	tp := newTestPair(t)
	// Channels are not JSON-serialisable → json.Marshal returns an error.
	// send() must silently drop the message without panicking.
	tp.sess.send(make(chan int))
}

// ─── DrawOverlayLabel with a real renderer ───────────────────────────────────

// coverageFontEngine is a minimal font.Engine that satisfies all interface
// methods for testing DrawOverlayLabel without a real font file.
type coverageFontEngine struct {
	glyphs map[rune]font.GlyphID
}

func newCoverageFontEngine() *coverageFontEngine {
	e := &coverageFontEngine{glyphs: make(map[rune]font.GlyphID)}
	for i, r := range " ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789.×" {
		e.glyphs[r] = font.GlyphID(i + 1)
	}
	return e
}
func (e *coverageFontEngine) LoadFont([]byte) (font.ID, error)              { return 1, nil }
func (e *coverageFontEngine) LoadFontFile(string) (font.ID, error)          { return 1, nil }
func (e *coverageFontEngine) UnloadFont(font.ID)                            {}
func (e *coverageFontEngine) SetDPIScale(float32)                           {}
func (e *coverageFontEngine) Destroy()                                      {}
func (e *coverageFontEngine) Kerning(font.ID, font.GlyphID, font.GlyphID, float32) float32 { return 0 }
func (e *coverageFontEngine) FontMetrics(_ font.ID, size float32) font.Metrics {
	return font.Metrics{Ascent: size * 0.8, Descent: size * 0.2, LineHeight: size * 1.2, UnitsPerEm: 1000}
}
func (e *coverageFontEngine) GlyphIndex(_ font.ID, r rune) font.GlyphID { return e.glyphs[r] }
func (e *coverageFontEngine) GlyphMetrics(_ font.ID, _ font.GlyphID, size float32) font.GlyphMetrics {
	adv := size * 0.6
	return font.GlyphMetrics{Width: adv, Height: size, BearingX: 0, BearingY: size * 0.8, Advance: adv}
}
func (e *coverageFontEngine) RasterizeGlyph(_ font.ID, _ font.GlyphID, size float32, sdf bool) (font.GlyphBitmap, error) {
	w, h := int(size*0.6)+1, int(size)+1
	return font.GlyphBitmap{Width: w, Height: h, Data: make([]byte, w*h), SDF: sdf}, nil
}
func (e *coverageFontEngine) HasGlyph(_ font.ID, r rune) bool {
	_, ok := e.glyphs[r]
	return ok
}

func newTestRenderer(t *testing.T) (*textrender.Renderer, font.ID) {
	t.Helper()
	eng := newCoverageFontEngine()
	mgr := font.NewManager(eng)
	fid, err := mgr.Register("Default", font.WeightRegular, font.StyleNormal, nil)
	if err != nil {
		t.Fatalf("Register font: %v", err)
	}
	gAtlas := atlas.New(atlas.Options{Width: 256, Height: 256}) // nil backend
	tr := textrender.New(textrender.Options{Manager: mgr, Atlas: gAtlas})
	return tr, fid
}

func TestDrawOverlayLabel_WithRenderer(t *testing.T) {
	tr, fid := newTestRenderer(t)
	srv := NewServer(Options{})
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "div", []string{"foo"}, "", uimath.NewRect(100, 100, 200, 50))
	srv.snapMu.Lock()
	srv.snapshot = snap
	srv.snapMu.Unlock()
	srv.overlayMu.Lock()
	srv.highlightID = 1
	srv.overlayMu.Unlock()

	buf := render.NewCommandBuffer()
	srv.DrawOverlayLabel(buf, tr, fid) // should not panic
}

func TestDrawOverlayLabel_LabelFlipsAbove(t *testing.T) {
	// Element near bottom: label should flip to render above the element.
	tr, fid := newTestRenderer(t)
	srv := NewServer(Options{})
	snap := makeSnap(800, 600)
	snap.Root = 1
	// Place element near the bottom so ly+boxH > ViewHeight
	addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.NewRect(100, 560, 200, 30))
	srv.snapMu.Lock()
	srv.snapshot = snap
	srv.snapMu.Unlock()
	srv.overlayMu.Lock()
	srv.highlightID = 1
	srv.overlayMu.Unlock()

	buf := render.NewCommandBuffer()
	srv.DrawOverlayLabel(buf, tr, fid)
}

func TestDrawOverlayLabel_NodeNotFound(t *testing.T) {
	tr, fid := newTestRenderer(t)
	srv := NewServer(Options{})
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.NewRect(0, 0, 100, 50))
	srv.snapMu.Lock()
	srv.snapshot = snap
	srv.snapMu.Unlock()
	srv.overlayMu.Lock()
	srv.highlightID = 999 // not in snapshot
	srv.overlayMu.Unlock()

	buf := render.NewCommandBuffer()
	srv.DrawOverlayLabel(buf, tr, fid)
}

func TestDrawOverlayLabel_NilSnapshot(t *testing.T) {
	tr, fid := newTestRenderer(t)
	srv := NewServer(Options{})
	srv.overlayMu.Lock()
	srv.highlightID = 1
	srv.overlayMu.Unlock()
	// snapshot is nil

	buf := render.NewCommandBuffer()
	srv.DrawOverlayLabel(buf, tr, fid)
}

func TestDrawOverlayLabel_NoClasses(t *testing.T) {
	tr, fid := newTestRenderer(t)
	srv := NewServer(Options{})
	snap := makeSnap(800, 600)
	snap.Root = 1
	// Node with no HTMLTag, no classes → uses ElemType fallback
	n := addNode(snap, 1, core.InvalidElementID, "", nil, "", uimath.NewRect(10, 10, 100, 40))
	n.ElemType = "button"
	srv.snapMu.Lock()
	srv.snapshot = snap
	srv.snapMu.Unlock()
	srv.overlayMu.Lock()
	srv.highlightID = 1
	srv.overlayMu.Unlock()

	buf := render.NewCommandBuffer()
	srv.DrawOverlayLabel(buf, tr, fid)
}

// ─── misc helpers ─────────────────────────────────────────────────────────────

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Ensure unused imports compile
var _ = fmt.Sprintf
var _ layout.Style
var _ uimath.Rect

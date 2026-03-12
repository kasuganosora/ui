package devtools

import (
	"encoding/json"
	"testing"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
)

// ─── handleMessage ────────────────────────────────────────────────────────────

func TestHandleMessage_InvalidJSON(t *testing.T) {
	tp := newTestPair(t)
	// Invalid JSON should be silently ignored (no response).
	// Use a timeout-based read to verify nothing is sent.
	tp.sess.handleMessage([]byte("not json"))
	msgs := tp.readAll(t)
	if len(msgs) != 0 {
		t.Errorf("invalid JSON should produce no response, got %d msgs", len(msgs))
	}
}

func TestHandleMessage_NoDot(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "ping", nil)
	if _, ok := m["result"]; !ok {
		t.Errorf("no-dot method should return empty result: %v", m)
	}
}

func TestHandleMessage_UnknownDomain(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "Foo.bar", nil)
	if _, ok := m["result"]; !ok {
		t.Errorf("unknown domain should return empty result: %v", m)
	}
}

func TestHandleMessage_KnownAcknowledgedDomains(t *testing.T) {
	tp := newTestPair(t)
	for _, method := range []string{
		"Target.activateTarget",
		"Browser.getVersion",
		"Network.enable",
		"Emulation.setDeviceMetricsOverride",
		"Debugger.enable",
	} {
		m := tp.send(t, 1, method, nil)
		if _, ok := m["result"]; !ok {
			t.Errorf("%s: expected result, got %v", method, m)
		}
	}
}

// ─── DOM domain ───────────────────────────────────────────────────────────────

func TestDOM_Enable(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "DOM.enable", nil)
	checkResult(t, m)
	if !tp.sess.domEnabled {
		t.Error("domEnabled should be true after enable")
	}
}

func TestDOM_Disable(t *testing.T) {
	tp := newTestPair(t)
	tp.send(t, 1, "DOM.enable", nil)
	m := tp.send(t, 2, "DOM.disable", nil)
	checkResult(t, m)
	if tp.sess.domEnabled {
		t.Error("domEnabled should be false after disable")
	}
}

func TestDOM_GetDocument_NilSnapshot(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "DOM.getDocument", map[string]any{"depth": -1})
	checkResult(t, m)
	result := m["result"].(map[string]any)
	root := result["root"].(map[string]any)
	if root["nodeName"] != "#document" {
		t.Errorf("want #document, got %v", root["nodeName"])
	}
}

func TestDOM_GetDocument_WithSnapshot(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.snapMu.Lock()
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "div", []string{"app"}, "", uimath.NewRect(0, 0, 800, 600))
	tp.srv.snapshot = snap
	tp.srv.snapMu.Unlock()

	m := tp.send(t, 1, "DOM.getDocument", map[string]any{"depth": 1})
	checkResult(t, m)
	result := m["result"].(map[string]any)
	root := result["root"].(map[string]any)
	if root["nodeName"] != "#document" {
		t.Errorf("want #document root")
	}
}

func TestDOM_RequestChildNodes(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.snapMu.Lock()
	snap := makeSnap(800, 600)
	snap.Root = 1
	p := addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.Rect{})
	addNode(snap, 2, 1, "span", nil, "text", uimath.Rect{})
	p.ChildIDs = []core.ElementID{2}
	tp.srv.snapshot = snap
	tp.srv.snapMu.Unlock()

	// requestChildNodes sends a result + DOM.setChildNodes event
	msgs := tp.sendN(t, 1, "DOM.requestChildNodes", map[string]any{"nodeId": 1, "depth": 1}, 2)
	if len(msgs) < 2 {
		t.Fatalf("want 2 messages, got %d: %v", len(msgs), msgs)
	}
	// One should be DOM.setChildNodes event
	found := false
	for _, msg := range msgs {
		if msg["method"] == "DOM.setChildNodes" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected DOM.setChildNodes event among: %v", msgs)
	}
}

func TestDOM_RequestChildNodes_TextChild(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.snapMu.Lock()
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "span", nil, "hello", uimath.Rect{})
	tp.srv.snapshot = snap
	tp.srv.snapMu.Unlock()

	msgs := tp.sendN(t, 1, "DOM.requestChildNodes", map[string]any{"nodeId": 1}, 2)
	// Should send DOM.setChildNodes with a #text child
	for _, msg := range msgs {
		if msg["method"] == "DOM.setChildNodes" {
			params := msg["params"].(map[string]any)
			nodes := params["nodes"].([]any)
			if len(nodes) != 1 {
				t.Fatalf("want 1 child, got %d", len(nodes))
			}
			child := nodes[0].(map[string]any)
			if child["nodeName"] != "#text" {
				t.Errorf("want #text child, got %v", child["nodeName"])
			}
		}
	}
}

func TestDOM_RequestChildNodes_NoSnapshot(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "DOM.requestChildNodes", map[string]any{"nodeId": 1})
	checkResult(t, m)
}

func TestDOM_GetBoxModel_NilSnapshot(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "DOM.getBoxModel", map[string]any{"nodeId": 1})
	checkError(t, m)
}

func TestDOM_GetBoxModel_WithNode(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.snapMu.Lock()
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.NewRect(10, 20, 100, 50))
	tp.srv.snapshot = snap
	tp.srv.snapMu.Unlock()

	m := tp.send(t, 1, "DOM.getBoxModel", map[string]any{"nodeId": 1})
	checkResult(t, m)
	result := m["result"].(map[string]any)
	model := result["model"].(map[string]any)
	if model["content"] == nil {
		t.Errorf("model.content missing")
	}
}

func TestDOM_GetOuterHTML(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.snapMu.Lock()
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "div", []string{"root"}, "", uimath.Rect{})
	tp.srv.snapshot = snap
	tp.srv.snapMu.Unlock()

	m := tp.send(t, 1, "DOM.getOuterHTML", map[string]any{"nodeId": 1})
	checkResult(t, m)
	result := m["result"].(map[string]any)
	html, ok := result["outerHTML"].(string)
	if !ok || html == "" {
		t.Errorf("outerHTML missing or empty: %v", result)
	}
	if !containsStr(html, "div") {
		t.Errorf("outerHTML should contain div: %q", html)
	}
}

func TestDOM_GetOuterHTML_NilSnapshot(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "DOM.getOuterHTML", map[string]any{"nodeId": 1})
	checkResult(t, m)
	result := m["result"].(map[string]any)
	if result["outerHTML"] != "" {
		t.Errorf("nil snapshot should return empty outerHTML")
	}
}

func TestDOM_DescribeNode(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.snapMu.Lock()
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.Rect{})
	tp.srv.snapshot = snap
	tp.srv.snapMu.Unlock()

	m := tp.send(t, 1, "DOM.describeNode", map[string]any{"nodeId": 1})
	checkResult(t, m)
}

func TestDOM_DescribeNode_NilSnapshot(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "DOM.describeNode", map[string]any{"nodeId": 1})
	checkError(t, m)
}

func TestDOM_GetNodeForLocation_Hit(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.snapMu.Lock()
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.NewRect(0, 0, 200, 200))
	tp.srv.snapshot = snap
	tp.srv.snapMu.Unlock()

	m := tp.send(t, 1, "DOM.getNodeForLocation", map[string]any{"x": 50, "y": 50})
	checkResult(t, m)
	result := m["result"].(map[string]any)
	if result["nodeId"] == nil {
		t.Errorf("expected nodeId in result: %v", result)
	}
}

func TestDOM_GetNodeForLocation_Miss(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.snapMu.Lock()
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.NewRect(0, 0, 100, 100))
	tp.srv.snapshot = snap
	tp.srv.snapMu.Unlock()

	m := tp.send(t, 1, "DOM.getNodeForLocation", map[string]any{"x": 500, "y": 500})
	checkError(t, m)
}

func TestDOM_QuerySelector(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.snapMu.Lock()
	snap := makeSnap(800, 600)
	snap.Root = 1
	p := addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.Rect{})
	addNode(snap, 2, 1, "span", []string{"target"}, "", uimath.Rect{})
	p.ChildIDs = []core.ElementID{2}
	tp.srv.snapshot = snap
	tp.srv.snapMu.Unlock()

	m := tp.send(t, 1, "DOM.querySelector", map[string]any{"nodeId": 1, "selector": ".target"})
	checkResult(t, m)
	result := m["result"].(map[string]any)
	if id, ok := result["nodeId"].(float64); !ok || id != 2 {
		t.Errorf("want nodeId=2, got %v", result["nodeId"])
	}
}

func TestDOM_QuerySelectorAll(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.snapMu.Lock()
	snap := makeSnap(800, 600)
	snap.Root = 1
	p := addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.Rect{})
	addNode(snap, 2, 1, "span", []string{"item"}, "", uimath.Rect{})
	addNode(snap, 3, 1, "span", []string{"item"}, "", uimath.Rect{})
	p.ChildIDs = []core.ElementID{2, 3}
	tp.srv.snapshot = snap
	tp.srv.snapMu.Unlock()

	m := tp.send(t, 1, "DOM.querySelectorAll", map[string]any{"nodeId": 1, "selector": ".item"})
	checkResult(t, m)
	result := m["result"].(map[string]any)
	ids, ok := result["nodeIds"].([]any)
	if !ok || len(ids) != 2 {
		t.Errorf("want 2 nodeIds, got %v", result["nodeIds"])
	}
}

func TestDOM_PushNodesByBackendIds(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "DOM.pushNodesByBackendIdsToFrontend", map[string]any{"backendNodeIds": []int{1, 2}})
	checkResult(t, m)
}

func TestDOM_GetAttributes(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.snapMu.Lock()
	snap := makeSnap(800, 600)
	snap.Root = 1
	n := addNode(snap, 1, core.InvalidElementID, "div", []string{"foo"}, "", uimath.NewRect(10, 20, 100, 50))
	n.IDAttr = "root"
	tp.srv.snapshot = snap
	tp.srv.snapMu.Unlock()

	m := tp.send(t, 1, "DOM.getAttributes", map[string]any{"nodeId": 1})
	checkResult(t, m)
	result := m["result"].(map[string]any)
	attrs, ok := result["attributes"].([]any)
	if !ok || len(attrs) == 0 {
		t.Errorf("expected attributes: %v", result)
	}
}

func TestDOM_SetOuterHTML(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "DOM.setOuterHTML", nil)
	checkResult(t, m)
}

func TestDOM_SetInspectedNode(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "DOM.setInspectedNode", nil)
	checkResult(t, m)
}

func TestDOM_ResolveNode(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "DOM.resolveNode", map[string]any{"nodeId": 42})
	checkResult(t, m)
	result := m["result"].(map[string]any)
	obj, ok := result["object"].(map[string]any)
	if !ok || obj["type"] != "object" {
		t.Errorf("unexpected resolveNode result: %v", result)
	}
}

func TestDOM_Unknown(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "DOM.unknownMethod", nil)
	checkResult(t, m)
}

// ─── CSS domain ───────────────────────────────────────────────────────────────

func TestCSS_Enable(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "CSS.enable", nil)
	checkResult(t, m)
	if !tp.sess.cssEnabled {
		t.Error("cssEnabled should be true")
	}
}

func TestCSS_Disable(t *testing.T) {
	tp := newTestPair(t)
	tp.send(t, 1, "CSS.enable", nil)
	m := tp.send(t, 2, "CSS.disable", nil)
	checkResult(t, m)
	if tp.sess.cssEnabled {
		t.Error("cssEnabled should be false")
	}
}

func TestCSS_GetComputedStyle_NilSnapshot(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "CSS.getComputedStyleForNode", map[string]any{"nodeId": 1})
	checkResult(t, m)
}

func TestCSS_GetComputedStyle_WithNode(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.snapMu.Lock()
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.NewRect(0, 0, 100, 50))
	tp.srv.snapshot = snap
	tp.srv.snapMu.Unlock()

	m := tp.send(t, 1, "CSS.getComputedStyleForNode", map[string]any{"nodeId": 1})
	checkResult(t, m)
	result := m["result"].(map[string]any)
	if result["computedStyle"] == nil {
		t.Error("computedStyle missing")
	}
}

func TestCSS_GetMatchedStyles_NilSnapshot(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "CSS.getMatchedStylesForNode", map[string]any{"nodeId": 1})
	checkResult(t, m)
}

func TestCSS_GetMatchedStyles_WithNode(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.snapMu.Lock()
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.Rect{})
	tp.srv.snapshot = snap
	tp.srv.snapMu.Unlock()

	m := tp.send(t, 1, "CSS.getMatchedStylesForNode", map[string]any{"nodeId": 1})
	checkResult(t, m)
	result := m["result"].(map[string]any)
	if result["inlineStyle"] == nil {
		t.Error("inlineStyle missing")
	}
}

func TestCSS_GetInlineStyles_NilSnapshot(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "CSS.getInlineStylesForNode", map[string]any{"nodeId": 1})
	checkResult(t, m)
}

func TestCSS_GetInlineStyles_WithNode(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.snapMu.Lock()
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.Rect{})
	tp.srv.snapshot = snap
	tp.srv.snapMu.Unlock()

	m := tp.send(t, 1, "CSS.getInlineStylesForNode", map[string]any{"nodeId": 1})
	checkResult(t, m)
}

func TestCSS_BackgroundColors(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "CSS.getBackgroundColors", nil)
	checkResult(t, m)
}

func TestCSS_StyleSheetText(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "CSS.getStyleSheetText", nil)
	checkResult(t, m)
}

func TestCSS_Unknown(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "CSS.unknownMethod", nil)
	checkResult(t, m)
}

// ─── Overlay domain ───────────────────────────────────────────────────────────

func TestOverlay_Enable(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "Overlay.enable", nil)
	checkResult(t, m)
	if !tp.sess.overlayEnabled {
		t.Error("overlayEnabled should be true")
	}
}

func TestOverlay_Disable(t *testing.T) {
	tp := newTestPair(t)
	tp.send(t, 1, "Overlay.enable", nil)
	m := tp.send(t, 2, "Overlay.disable", nil)
	checkResult(t, m)
	if tp.sess.overlayEnabled {
		t.Error("overlayEnabled should be false")
	}
	if tp.srv.highlightID != core.InvalidElementID {
		t.Error("highlight should be cleared on disable")
	}
}

func TestOverlay_HighlightNode(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "Overlay.highlightNode", map[string]any{"nodeId": 42})
	checkResult(t, m)
	tp.srv.overlayMu.Lock()
	id := tp.srv.highlightID
	tp.srv.overlayMu.Unlock()
	if id != 42 {
		t.Errorf("highlight should be 42, got %d", id)
	}
}

func TestOverlay_HighlightNode_BackendNodeID(t *testing.T) {
	tp := newTestPair(t)
	// nodeId=0, backendNodeId=7 → should use backendNodeId
	m := tp.send(t, 1, "Overlay.highlightNode", map[string]any{"nodeId": 0, "backendNodeId": 7})
	checkResult(t, m)
	tp.srv.overlayMu.Lock()
	id := tp.srv.highlightID
	tp.srv.overlayMu.Unlock()
	if id != 7 {
		t.Errorf("want highlightID=7, got %d", id)
	}
}

func TestOverlay_HideHighlight(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.overlayMu.Lock()
	tp.srv.highlightID = 5
	tp.srv.overlayMu.Unlock()

	m := tp.send(t, 1, "Overlay.hideHighlight", nil)
	checkResult(t, m)
	tp.srv.overlayMu.Lock()
	id := tp.srv.highlightID
	tp.srv.overlayMu.Unlock()
	if id != core.InvalidElementID {
		t.Errorf("highlight should be cleared")
	}
}

func TestOverlay_HighlightRect(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "Overlay.highlightRect", nil)
	checkResult(t, m)
}

func TestOverlay_SetInspectMode_None(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.overlayMu.Lock()
	tp.srv.highlightID = 5
	tp.srv.overlayMu.Unlock()

	m := tp.send(t, 1, "Overlay.setInspectMode", map[string]any{"mode": "none"})
	checkResult(t, m)
	tp.srv.overlayMu.Lock()
	id := tp.srv.highlightID
	tp.srv.overlayMu.Unlock()
	if id != core.InvalidElementID {
		t.Errorf("mode=none should clear highlight")
	}
}

func TestOverlay_SetInspectMode_Search(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "Overlay.setInspectMode", map[string]any{"mode": "searchForNode"})
	checkResult(t, m)
}

func TestOverlay_NoOp(t *testing.T) {
	tp := newTestPair(t)
	for _, method := range []string{
		"Overlay.setPausedInDebuggerMessage",
		"Overlay.setShowFPSCounter",
		"Overlay.setShowPaintRects",
		"Overlay.setShowLayoutShiftRegions",
		"Overlay.setShowScrollBottleneckRects",
		"Overlay.setShowHitTestBorders",
		"Overlay.setShowWebVitals",
		"Overlay.setShowViewportSizeOnResize",
		"Overlay.setShowAdHighlights",
		"Overlay.getHighlightObjectForTest",
	} {
		m := tp.send(t, 1, method, nil)
		checkResult(t, m)
	}
}

func TestOverlay_Unknown(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "Overlay.unknownMethod", nil)
	checkResult(t, m)
}

// ─── Page domain ─────────────────────────────────────────────────────────────

func TestPage_Enable(t *testing.T) {
	tp := newTestPair(t)
	// Page.enable sends result + 2 events
	msgs := tp.sendN(t, 1, "Page.enable", nil, 3)
	if len(msgs) < 1 {
		t.Fatalf("want ≥1 messages, got 0")
	}
	if !tp.sess.pageEnabled {
		t.Error("pageEnabled should be true")
	}
	// Should include load events
	methods := collectMethods(msgs)
	if !containsStr(methods, "Page.loadEventFired") {
		t.Errorf("expected Page.loadEventFired among %v", methods)
	}
}

func TestPage_Disable(t *testing.T) {
	tp := newTestPair(t)
	tp.sendN(t, 1, "Page.enable", nil, 3)
	m := tp.send(t, 2, "Page.disable", nil)
	checkResult(t, m)
	if tp.sess.pageEnabled {
		t.Error("pageEnabled should be false")
	}
}

func TestPage_GetFrameTree(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "Page.getFrameTree", nil)
	checkResult(t, m)
	result := m["result"].(map[string]any)
	if result["frameTree"] == nil {
		t.Error("frameTree missing")
	}
}

func TestPage_GetResourceTree(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "Page.getResourceTree", nil)
	checkResult(t, m)
}

func TestPage_GetLayoutMetrics(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.snapMu.Lock()
	snap := makeSnap(1280, 720)
	tp.srv.snapshot = snap
	tp.srv.snapMu.Unlock()

	m := tp.send(t, 1, "Page.getLayoutMetrics", nil)
	checkResult(t, m)
	result := m["result"].(map[string]any)
	lv := result["layoutViewport"].(map[string]any)
	if lv["clientWidth"] != float64(1280) {
		t.Errorf("want 1280, got %v", lv["clientWidth"])
	}
}

func TestPage_Reload(t *testing.T) {
	tp := newTestPair(t)
	called := false
	tp.srv.Attach(func() { called = true })
	m := tp.send(t, 1, "Page.reload", nil)
	checkResult(t, m)
	if !called {
		t.Error("reload should call markDirty")
	}
}

func TestPage_Navigate(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "Page.navigate", nil)
	checkResult(t, m)
}

func TestPage_NoOp(t *testing.T) {
	tp := newTestPair(t)
	for _, method := range []string{
		"Page.stopLoading",
		"Page.setLifecycleEventsEnabled",
		"Page.createIsolatedWorld",
		"Page.addScriptToEvaluateOnNewDocument",
		"Page.setBypassCSP",
		"Page.captureScreenshot",
	} {
		m := tp.send(t, 1, method, nil)
		checkResult(t, m)
	}
}

// ─── Runtime domain ───────────────────────────────────────────────────────────

func TestRuntime_Enable(t *testing.T) {
	tp := newTestPair(t)
	msgs := tp.sendN(t, 1, "Runtime.enable", nil, 2)
	if len(msgs) < 1 {
		t.Fatalf("want ≥1 messages")
	}
	if !tp.sess.runtimeEnabled {
		t.Error("runtimeEnabled should be true")
	}
	// Should fire executionContextCreated
	methods := collectMethods(msgs)
	if !containsStr(methods, "Runtime.executionContextCreated") {
		t.Errorf("expected Runtime.executionContextCreated among %v", methods)
	}
}

func TestRuntime_Disable(t *testing.T) {
	tp := newTestPair(t)
	tp.sendN(t, 1, "Runtime.enable", nil, 2)
	m := tp.send(t, 2, "Runtime.disable", nil)
	checkResult(t, m)
	if tp.sess.runtimeEnabled {
		t.Error("runtimeEnabled should be false")
	}
}

func TestRuntime_Evaluate(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "Runtime.evaluate", map[string]any{"expression": "document.title"})
	checkResult(t, m)
	result := m["result"].(map[string]any)
	r := result["result"].(map[string]any)
	if r["type"] != "undefined" {
		t.Errorf("evaluate should return undefined, got %v", r["type"])
	}
}

func TestRuntime_CallFunctionOn_OuterHTML(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.snapMu.Lock()
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "div", []string{"app"}, "", uimath.Rect{})
	tp.srv.snapshot = snap
	tp.srv.snapMu.Unlock()

	m := tp.send(t, 1, "Runtime.callFunctionOn", map[string]any{
		"objectId":            "node-1",
		"functionDeclaration": "function() { return this.outerHTML; }",
	})
	checkResult(t, m)
	result := m["result"].(map[string]any)
	r := result["result"].(map[string]any)
	if r["type"] != "string" {
		t.Errorf("expected string type, got %v", r["type"])
	}
	if html, ok := r["value"].(string); !ok || !containsStr(html, "div") {
		t.Errorf("outerHTML should contain div: %v", r["value"])
	}
}

func TestRuntime_CallFunctionOn_InnerHTML(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.snapMu.Lock()
	snap := makeSnap(800, 600)
	snap.Root = 1
	p := addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.Rect{})
	addNode(snap, 2, 1, "span", nil, "hello", uimath.Rect{})
	p.ChildIDs = []core.ElementID{2}
	tp.srv.snapshot = snap
	tp.srv.snapMu.Unlock()

	m := tp.send(t, 1, "Runtime.callFunctionOn", map[string]any{
		"objectId":            "node-1",
		"functionDeclaration": "function() { return this.innerHTML; }",
	})
	checkResult(t, m)
	result := m["result"].(map[string]any)
	r := result["result"].(map[string]any)
	if r["type"] != "string" {
		t.Errorf("expected string, got %v", r["type"])
	}
	if !containsStr(r["value"].(string), "span") {
		t.Errorf("innerHTML should contain span: %v", r["value"])
	}
}

func TestRuntime_CallFunctionOn_TextContent(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.snapMu.Lock()
	snap := makeSnap(800, 600)
	snap.Root = 1
	p := addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.Rect{})
	addNode(snap, 2, 1, "span", nil, "Hello", uimath.Rect{})
	addNode(snap, 3, 1, "span", nil, "World", uimath.Rect{})
	p.ChildIDs = []core.ElementID{2, 3}
	tp.srv.snapshot = snap
	tp.srv.snapMu.Unlock()

	m := tp.send(t, 1, "Runtime.callFunctionOn", map[string]any{
		"objectId":            "node-1",
		"functionDeclaration": "function() { return this.textContent; }",
	})
	checkResult(t, m)
	result := m["result"].(map[string]any)
	r := result["result"].(map[string]any)
	if r["value"] != "HelloWorld" {
		t.Errorf("textContent: want 'HelloWorld', got %v", r["value"])
	}
}

func TestRuntime_CallFunctionOn_LeafInnerHTML(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.snapMu.Lock()
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "span", nil, "leaf text", uimath.Rect{})
	tp.srv.snapshot = snap
	tp.srv.snapMu.Unlock()

	m := tp.send(t, 1, "Runtime.callFunctionOn", map[string]any{
		"objectId":            "node-1",
		"functionDeclaration": "function() { return this.innerHTML; }",
	})
	result := m["result"].(map[string]any)
	r := result["result"].(map[string]any)
	if r["value"] != "leaf text" {
		t.Errorf("leaf innerHTML: want 'leaf text', got %v", r["value"])
	}
}

func TestRuntime_CallFunctionOn_NilSnapshot(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "Runtime.callFunctionOn", map[string]any{
		"objectId":            "node-1",
		"functionDeclaration": "function() { return this.outerHTML; }",
	})
	// nil snapshot → falls through to undefined
	result := m["result"].(map[string]any)
	r := result["result"].(map[string]any)
	if r["type"] != "undefined" {
		t.Errorf("nil snapshot should return undefined, got %v", r["type"])
	}
}

func TestRuntime_CallFunctionOn_UnknownObjectId(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "Runtime.callFunctionOn", map[string]any{
		"objectId":            "other-123",
		"functionDeclaration": "function() { return 1; }",
	})
	result := m["result"].(map[string]any)
	r := result["result"].(map[string]any)
	if r["type"] != "undefined" {
		t.Errorf("unknown objectId should return undefined, got %v", r["type"])
	}
}

func TestRuntime_GetProperties(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "Runtime.getProperties", nil)
	checkResult(t, m)
}

func TestRuntime_NoOp(t *testing.T) {
	tp := newTestPair(t)
	for _, method := range []string{
		"Runtime.releaseObject",
		"Runtime.releaseObjectGroup",
		"Runtime.runIfWaitingForDebugger",
		"Runtime.discardConsoleEntries",
	} {
		m := tp.send(t, 1, method, nil)
		checkResult(t, m)
	}
}

// ─── Log domain ───────────────────────────────────────────────────────────────

func TestLog_Enable(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "Log.enable", nil)
	checkResult(t, m)
	if !tp.sess.logEnabled {
		t.Error("logEnabled should be true")
	}
}

func TestLog_Disable(t *testing.T) {
	tp := newTestPair(t)
	tp.send(t, 1, "Log.enable", nil)
	m := tp.send(t, 2, "Log.disable", nil)
	checkResult(t, m)
	if tp.sess.logEnabled {
		t.Error("logEnabled should be false")
	}
}

func TestLog_Clear(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "Log.clear", nil)
	checkResult(t, m)
}

func TestLog_ViolationsNoOp(t *testing.T) {
	tp := newTestPair(t)
	for _, method := range []string{"Log.startViolationsReport", "Log.stopViolationsReport"} {
		m := tp.send(t, 1, method, nil)
		checkResult(t, m)
	}
}

func TestLog_Unknown(t *testing.T) {
	tp := newTestPair(t)
	m := tp.send(t, 1, "Log.unknownMethod", nil)
	checkResult(t, m)
}

// ─── Server.Log → Log.entryAdded event ───────────────────────────────────────

func TestServerLog_SendsEventToEnabledSession(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.addSession(tp.sess)

	tp.send(t, 1, "Log.enable", nil)
	tp.srv.Log("info", "javascript", "test message")

	msgs := tp.readAll(t)
	found := false
	for _, m := range msgs {
		if m["method"] == "Log.entryAdded" {
			found = true
			params := m["params"].(map[string]any)
			entry := params["entry"].(map[string]any)
			if entry["text"] != "test message" {
				t.Errorf("unexpected text: %v", entry["text"])
			}
		}
	}
	if !found {
		t.Error("expected Log.entryAdded event")
	}
}

func TestServerLog_NoEventToDisabledSession(t *testing.T) {
	tp := newTestPair(t)
	tp.srv.addSession(tp.sess)
	// log NOT enabled
	tp.srv.Log("info", "javascript", "test message")
	msgs := tp.readAll(t)
	for _, m := range msgs {
		if m["method"] == "Log.entryAdded" {
			t.Error("should not receive Log.entryAdded when log disabled")
		}
	}
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func checkResult(t *testing.T, m map[string]any) {
	t.Helper()
	if m["error"] != nil {
		t.Errorf("unexpected error: %v", m["error"])
	}
	if _, ok := m["result"]; !ok {
		t.Errorf("expected 'result' key in: %v", m)
	}
}

func checkError(t *testing.T, m map[string]any) {
	t.Helper()
	if m["error"] == nil {
		t.Errorf("expected error, got: %v", m)
	}
}

func collectMethods(msgs []map[string]any) string {
	b, _ := json.Marshal(msgs)
	return string(b)
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}

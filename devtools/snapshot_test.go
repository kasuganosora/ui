package devtools

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// ─── buildSnapshot / buildSnapshotNode ───────────────────────────────────────

func TestBuildSnapshot_BasicTree(t *testing.T) {
	tree := core.NewTree()
	root := widget.NewDiv(tree, nil)

	snap := buildSnapshot(tree, root, 800, 600)
	if snap == nil {
		t.Fatal("expected non-nil snapshot")
	}
	if snap.ViewWidth != 800 || snap.ViewHeight != 600 {
		t.Errorf("view size: want 800×600, got %.0f×%.0f", snap.ViewWidth, snap.ViewHeight)
	}
	if snap.Root != root.ElementID() {
		t.Errorf("root ID mismatch: want %d, got %d", root.ElementID(), snap.Root)
	}
	if _, ok := snap.Nodes[root.ElementID()]; !ok {
		t.Error("root node missing from snapshot")
	}
}

func TestBuildSnapshot_WithChildren(t *testing.T) {
	tree := core.NewTree()
	root := widget.NewDiv(tree, nil)
	span := widget.NewText(tree, "hello", nil)
	root.AppendChild(span)

	snap := buildSnapshot(tree, root, 800, 600)
	if len(snap.Nodes) != 2 {
		t.Errorf("want 2 nodes (root+child), got %d", len(snap.Nodes))
	}
	rootNode := snap.Nodes[root.ElementID()]
	if len(rootNode.ChildIDs) != 1 {
		t.Errorf("root should have 1 child, got %d", len(rootNode.ChildIDs))
	}
	childNode := snap.Nodes[span.ElementID()]
	if childNode == nil {
		t.Fatal("child node missing from snapshot")
	}
	if childNode.Text != "hello" {
		t.Errorf("want text 'hello', got %q", childNode.Text)
	}
	if childNode.ParentID != root.ElementID() {
		t.Errorf("child parentID wrong: want %d, got %d", root.ElementID(), childNode.ParentID)
	}
}

func TestBuildSnapshot_HTMLTagProperty(t *testing.T) {
	tree := core.NewTree()
	root := widget.NewDiv(tree, nil)
	tree.SetProperty(root.ElementID(), "html-tag", "section")

	snap := buildSnapshot(tree, root, 800, 600)
	node := snap.Nodes[root.ElementID()]
	if node.HTMLTag != "section" {
		t.Errorf("want HTMLTag='section', got %q", node.HTMLTag)
	}
}

func TestBuildSnapshot_IDAttrProperty(t *testing.T) {
	tree := core.NewTree()
	root := widget.NewDiv(tree, nil)
	tree.SetProperty(root.ElementID(), "id", "my-root")

	snap := buildSnapshot(tree, root, 800, 600)
	node := snap.Nodes[root.ElementID()]
	if node.IDAttr != "my-root" {
		t.Errorf("want IDAttr='my-root', got %q", node.IDAttr)
	}
}

func TestBuildSnapshot_ElemType(t *testing.T) {
	tree := core.NewTree()
	root := widget.NewText(tree, "text node", nil)

	snap := buildSnapshot(tree, root, 800, 600)
	node := snap.Nodes[root.ElementID()]
	if node.ElemType != core.TypeText {
		t.Errorf("want TypeText, got %q", node.ElemType)
	}
}

func TestBuildSnapshot_Classes(t *testing.T) {
	tree := core.NewTree()
	root := widget.NewDiv(tree, nil)
	tree.SetClasses(root.ElementID(), []string{"foo", "bar"})

	snap := buildSnapshot(tree, root, 800, 600)
	node := snap.Nodes[root.ElementID()]
	if len(node.Classes) != 2 {
		t.Errorf("want 2 classes, got %d: %v", len(node.Classes), node.Classes)
	}
}

// ─── AfterLayout ─────────────────────────────────────────────────────────────

func TestAfterLayout_SetsSnapshot(t *testing.T) {
	srv := NewServer(Options{})
	tree := core.NewTree()
	root := widget.NewDiv(tree, nil)

	srv.AfterLayout(tree, root, 800, 600)

	snap := srv.getSnapshot()
	if snap == nil {
		t.Fatal("snapshot should be set after AfterLayout")
	}
	if snap.ViewWidth != 800 || snap.ViewHeight != 600 {
		t.Errorf("view size wrong after AfterLayout: %.0f×%.0f", snap.ViewWidth, snap.ViewHeight)
	}
}

func TestAfterLayout_BroadcastsDocumentUpdated(t *testing.T) {
	srv := NewServer(Options{})

	tp := newTestPair(t)
	srv.addSession(tp.sess)
	tp.sess.domEnabled = true

	tree := core.NewTree()
	root := widget.NewDiv(tree, nil)
	srv.AfterLayout(tree, root, 800, 600)

	msgs := tp.readAll(t)
	found := false
	for _, m := range msgs {
		if m["method"] == "DOM.documentUpdated" {
			found = true
		}
	}
	if !found {
		t.Error("expected DOM.documentUpdated broadcast after AfterLayout")
	}
}

// ─── removeSession / broadcast ────────────────────────────────────────────────

func TestRemoveSession(t *testing.T) {
	srv := NewServer(Options{})
	tp := newTestPair(t)
	srv.addSession(tp.sess)

	if len(srv.sessions) != 1 {
		t.Fatalf("want 1 session, got %d", len(srv.sessions))
	}
	srv.removeSession(tp.sess)
	if len(srv.sessions) != 0 {
		t.Errorf("want 0 sessions after remove, got %d", len(srv.sessions))
	}
}

func TestBroadcast_ToSessions(t *testing.T) {
	srv := NewServer(Options{})
	tp := newTestPair(t)
	srv.addSession(tp.sess)

	srv.broadcast("Test.event", map[string]any{"data": "hello"})
	msgs := tp.readAll(t)

	if len(msgs) != 1 {
		t.Fatalf("want 1 message, got %d", len(msgs))
	}
	if msgs[0]["method"] != "Test.event" {
		t.Errorf("unexpected method: %v", msgs[0]["method"])
	}
}

// ─── HTTP endpoints (jsonList, jsonVersion, jsonProtocol) ────────────────────

func TestHTTP_JsonList(t *testing.T) {
	srv := NewServer(Options{Addr: "127.0.0.1:9222", AppName: "TestApp"})
	ts := httptest.NewServer(srv.router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/json/list")
	if err != nil {
		t.Fatalf("GET /json/list: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var result []map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("want 1 target, got %d", len(result))
	}
	if result[0]["title"] != "TestApp" {
		t.Errorf("want title=TestApp, got %v", result[0]["title"])
	}
	if result[0]["type"] != "page" {
		t.Errorf("want type=page, got %v", result[0]["type"])
	}
	if result[0]["webSocketDebuggerUrl"] == nil {
		t.Errorf("webSocketDebuggerUrl missing")
	}
}

func TestHTTP_JsonVersion(t *testing.T) {
	srv := NewServer(Options{Addr: "127.0.0.1:9222"})
	ts := httptest.NewServer(srv.router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/json/version")
	if err != nil {
		t.Fatalf("GET /json/version: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if result["Browser"] == "" {
		t.Error("Browser field missing")
	}
}

func TestHTTP_JsonProtocol(t *testing.T) {
	srv := NewServer(Options{})
	ts := httptest.NewServer(srv.router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/json/protocol")
	if err != nil {
		t.Fatalf("GET /json/protocol: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
}

func TestHTTP_JsonAlias(t *testing.T) {
	srv := NewServer(Options{})
	ts := httptest.NewServer(srv.router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/json")
	if err != nil {
		t.Fatalf("GET /json: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
}

func TestHTTP_WebSocket_Session(t *testing.T) {
	srv := NewServer(Options{})
	ts := httptest.NewServer(srv.router)
	defer ts.Close()

	wsURL := "ws" + ts.URL[4:] + "/devtools/page/" + targetID
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial WS: %v", err)
	}
	defer conn.Close()

	req := map[string]any{"id": 1, "method": "DOM.enable"}
	data, _ := json.Marshal(req)
	conn.WriteMessage(websocket.TextMessage, data)

	conn.SetReadDeadline(time.Now().Add(time.Second))
	_, resp, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var m map[string]any
	json.Unmarshal(resp, &m)
	if _, ok := m["result"]; !ok {
		t.Errorf("expected result, got %v", m)
	}
}

func TestRun_ClosesOnDisconnect(t *testing.T) {
	srv := NewServer(Options{})
	ts := httptest.NewServer(srv.router)
	defer ts.Close()

	wsURL := "ws" + ts.URL[4:] + "/devtools/page/" + targetID
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	// Close immediately — session.run() should exit cleanly
	conn.Close()
	time.Sleep(50 * time.Millisecond) // allow goroutine to finish
}

// ─── extra CSS coverage ───────────────────────────────────────────────────────

func TestWhiteSpaceCSS_AllVariants(t *testing.T) {
	cases := []struct {
		ws   layout.WhiteSpace
		want string
	}{
		{layout.WhiteSpaceNowrap, "nowrap"},
		{layout.WhiteSpacePre, "pre"},
		{layout.WhiteSpaceNormal, "normal"},
		{layout.WhiteSpace(99), "normal"}, // unknown → normal
	}
	for _, c := range cases {
		if got := whiteSpaceCSS(c.ws); got != c.want {
			t.Errorf("whiteSpaceCSS(%v): want %q, got %q", c.ws, c.want, got)
		}
	}
}

func TestDeclaredStyleProps_ZeroPaddingNotEmitted(t *testing.T) {
	tree := core.NewTree()
	root := widget.NewText(tree, "", nil)
	props := declaredStyleProps(root.Style())
	for _, p := range props {
		if strings.HasPrefix(p.Name, "padding") && p.Value == "0.00px" {
			t.Errorf("zero padding should not be emitted: %q=%q", p.Name, p.Value)
		}
	}
}

// ─── DrawOverlay with highlight ───────────────────────────────────────────────

func TestDrawOverlay_WithNonZeroBounds(t *testing.T) {
	srv := NewServer(Options{})
	snap := makeSnap(800, 600)
	snap.Root = 1
	n := addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.NewRect(10, 20, 100, 50))
	n.Padding = uimath.EdgesAll(4)
	n.Margin = uimath.EdgesAll(8)
	n.Border = uimath.EdgesAll(1)

	srv.snapMu.Lock()
	srv.snapshot = snap
	srv.snapMu.Unlock()
	srv.overlayMu.Lock()
	srv.highlightID = 1
	srv.overlayMu.Unlock()

	buf := render.NewCommandBuffer()
	srv.DrawOverlay(buf)
	// Should emit at least content + padding + margin = 3 overlay rects
	if len(buf.Overlays()) < 3 {
		t.Errorf("want ≥3 overlay rects, got %d", len(buf.Overlays()))
	}
}

func TestDeclaredStyleProps_PerSideMargin(t *testing.T) {
	st := layout.Style{
		Margin: layout.EdgeValues{
			Top:   pxVal(1),
			Right: pxVal(2),
			Bottom: pxVal(3),
			Left:  pxVal(4),
		},
	}
	props := declaredStyleProps(st)
	names := make(map[string]string)
	for _, p := range props {
		names[p.Name] = p.Value
	}
	if names["margin-top"] != "1.00px" {
		t.Errorf("margin-top: got %q", names["margin-top"])
	}
}

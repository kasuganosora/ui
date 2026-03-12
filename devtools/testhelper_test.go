package devtools

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
)

// ─── WebSocket test pair ──────────────────────────────────────────────────────

// testPair holds a Server + in-memory WebSocket session pair for handler tests.
type testPair struct {
	srv    *Server
	sess   *Session
	client *websocket.Conn
}

func newTestPair(t *testing.T) *testPair {
	t.Helper()
	srv := NewServer(Options{Addr: "127.0.0.1:0", AppName: "test"})

	sessCh := make(chan *Session, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		s := &Session{srv: srv, conn: conn}
		sessCh <- s
		// Keep server-side connection alive so the client can read responses.
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	t.Cleanup(ts.Close)

	client, _, err := websocket.DefaultDialer.Dial("ws"+ts.URL[4:]+"/", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { client.Close() })

	return &testPair{srv: srv, sess: <-sessCh, client: client}
}

// send dispatches a CDP request directly to the session (bypassing the WebSocket
// read loop) and returns the first response the client sees.
func (tp *testPair) send(t *testing.T, id int, method string, params any) map[string]any {
	t.Helper()
	raw := map[string]any{"id": id, "method": method}
	if params != nil {
		raw["params"] = params
	}
	data, _ := json.Marshal(raw)
	tp.sess.handleMessage(data)

	tp.client.SetReadDeadline(time.Now().Add(time.Second))
	_, resp, err := tp.client.ReadMessage()
	tp.client.SetReadDeadline(time.Time{})
	if err != nil {
		t.Fatalf("read response for %s: %v", method, err)
	}
	var m map[string]any
	json.Unmarshal(resp, &m)
	return m
}

// readAll drains any pending messages from the client within a short timeout.
func (tp *testPair) readAll(t *testing.T) []map[string]any {
	t.Helper()
	var msgs []map[string]any
	for {
		tp.client.SetReadDeadline(time.Now().Add(30 * time.Millisecond))
		_, data, err := tp.client.ReadMessage()
		if err != nil {
			break
		}
		var m map[string]any
		json.Unmarshal(data, &m)
		msgs = append(msgs, m)
	}
	tp.client.SetReadDeadline(time.Time{})
	return msgs
}

// sendRaw dispatches a CDP request and collects n messages (result + any events).
func (tp *testPair) sendN(t *testing.T, id int, method string, params any, n int) []map[string]any {
	t.Helper()
	raw := map[string]any{"id": id, "method": method}
	if params != nil {
		raw["params"] = params
	}
	data, _ := json.Marshal(raw)
	tp.sess.handleMessage(data)

	var msgs []map[string]any
	for i := 0; i < n; i++ {
		tp.client.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		_, resp, err := tp.client.ReadMessage()
		tp.client.SetReadDeadline(time.Time{})
		if err != nil {
			break
		}
		var m map[string]any
		json.Unmarshal(resp, &m)
		msgs = append(msgs, m)
	}
	return msgs
}

// ─── Snapshot builders ────────────────────────────────────────────────────────

// makeSnap builds a minimal Snapshot for use in pure-function tests.
// Pass triples of (id, parentID, text) for leaf nodes.
func makeSnap(vw, vh float32) *Snapshot {
	return &Snapshot{
		Nodes:      make(map[core.ElementID]*NodeSnapshot),
		ViewWidth:  vw,
		ViewHeight: vh,
	}
}

func addNode(snap *Snapshot, id, parent core.ElementID, htmlTag string, classes []string,
	text string, bounds uimath.Rect) *NodeSnapshot {
	n := &NodeSnapshot{
		ID:       id,
		ParentID: parent,
		HTMLTag:  htmlTag,
		ElemType: core.TypeDiv,
		Classes:  classes,
		Text:     text,
		Bounds:   bounds,
	}
	if parent != core.InvalidElementID {
		if p, ok := snap.Nodes[parent]; ok {
			p.ChildIDs = append(p.ChildIDs, id)
		}
	}
	snap.Nodes[id] = n
	return n
}

// ─── Layout helper ────────────────────────────────────────────────────────────

func pxVal(v float32) layout.Value { return layout.Value{Unit: layout.UnitPx, Amount: v} }
func pctVal(v float32) layout.Value {
	return layout.Value{Unit: layout.UnitPercent, Amount: v}
}

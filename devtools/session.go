package devtools

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

// Request is an incoming CDP JSON-RPC message from the DevTools client.
type Request struct {
	ID     int             `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

// Response is a CDP JSON-RPC result message sent back to the client.
type Response struct {
	ID     int       `json:"id"`
	Result any       `json:"result"`
	Error  *RPCError `json:"error,omitempty"`
}

// cdpEvent is an unsolicited CDP event pushed to the client.
type cdpEvent struct {
	Method string `json:"method"`
	Params any    `json:"params"`
}

// RPCError is a CDP protocol error.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Session is one active Chrome DevTools WebSocket connection.
type Session struct {
	srv  *Server
	conn *websocket.Conn

	sendMu sync.Mutex // serialises WebSocket writes

	// Per-session domain enable flags
	domEnabled     bool
	cssEnabled     bool
	logEnabled     bool
	runtimeEnabled bool
	overlayEnabled bool
	pageEnabled    bool
}

func newSession(srv *Server, conn *websocket.Conn) *Session {
	return &Session{srv: srv, conn: conn}
}

// run is the session's read loop. Blocks until the WebSocket closes.
func (s *Session) run() {
	defer s.conn.Close()
	for {
		_, msg, err := s.conn.ReadMessage()
		if err != nil {
			break
		}
		s.handleMessage(msg)
	}
}

func (s *Session) handleMessage(msg []byte) {
	var req Request
	if err := json.Unmarshal(msg, &req); err != nil {
		return
	}

	dot := strings.IndexByte(req.Method, '.')
	if dot < 0 {
		s.sendResult(req.ID, map[string]any{})
		return
	}
	domain := req.Method[:dot]

	switch domain {
	case "DOM":
		s.handleDOM(req)
	case "CSS":
		s.handleCSS(req)
	case "Overlay":
		s.handleOverlay(req)
	case "Page":
		s.handlePage(req)
	case "Runtime":
		s.handleRuntime(req)
	case "Log":
		s.handleLog(req)
	// Domains we acknowledge but don't fully implement
	case "Target", "Browser", "Inspector",
		"Network", "Emulation", "Input",
		"Security", "ServiceWorker", "Storage",
		"Performance", "Profiler", "Debugger", "HeapProfiler":
		s.sendResult(req.ID, map[string]any{})
	default:
		s.sendResult(req.ID, map[string]any{})
	}
}

// ----- send helpers -----

func (s *Session) sendResult(id int, result any) {
	s.send(Response{ID: id, Result: result})
}

func (s *Session) sendError(id, code int, message string) {
	s.send(Response{ID: id, Error: &RPCError{Code: code, Message: message}})
}

func (s *Session) sendEvent(method string, params any) {
	s.send(cdpEvent{Method: method, Params: params})
}

func (s *Session) send(v any) {
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	s.sendMu.Lock()
	defer s.sendMu.Unlock()
	_ = s.conn.WriteMessage(websocket.TextMessage, data)
}

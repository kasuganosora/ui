package devtools

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/font"
	"github.com/kasuganosora/ui/font/textrender"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// Options configures the DevTools server.
type Options struct {
	// Addr is the TCP address to listen on (default ":9222").
	Addr string
	// AppName is shown in chrome://inspect (default "UI App").
	AppName string
}

// LogEntry is a console log entry visible in the DevTools Console panel.
type LogEntry struct {
	// Source: "javascript" | "network" | "storage" | "other"
	Source string `json:"source"`
	// Level: "verbose" | "info" | "warning" | "error"
	Level     string  `json:"level"`
	Text      string  `json:"text"`
	Timestamp float64 `json:"timestamp"` // milliseconds since epoch
	URL       string  `json:"url,omitempty"`
}

// Server implements the Chrome DevTools Protocol (CDP) over WebSocket.
// It exposes the UI widget tree for inspection, layout debugging, and console logging.
type Server struct {
	opts Options

	// Snapshot of the widget tree (rebuilt after each layout pass).
	// Protected by snapMu; written on render goroutine, read on session goroutines.
	snapMu   sync.RWMutex
	snapshot *Snapshot

	// Overlay: which element to highlight in the next frame.
	// Protected by overlayMu; written by sessions, read by render goroutine.
	overlayMu   sync.Mutex
	highlightID core.ElementID

	// markDirty triggers a redraw when the overlay state changes.
	// Set by App via attach().
	markDirty func()

	// Active WebSocket sessions.
	sessionsMu sync.Mutex
	sessions   []*Session

	// Gin HTTP router and underlying http.Server.
	router  *gin.Engine
	httpSrv *http.Server
}

// NewServer creates a DevTools server. Call Start() to begin accepting connections.
func NewServer(opts Options) *Server {
	if opts.Addr == "" {
		opts.Addr = "127.0.0.1:9222"
	}
	if opts.AppName == "" {
		opts.AppName = "UI App"
	}

	gin.SetMode(gin.ReleaseMode)
	s := &Server{opts: opts}
	s.router = gin.New()
	s.router.Use(gin.Recovery())
	s.setupRoutes()
	return s
}

// Start begins accepting DevTools connections. Blocks until Stop is called.
// Typical usage: go srv.Start()
func (s *Server) Start() error {
	s.httpSrv = &http.Server{
		Addr:    s.opts.Addr,
		Handler: s.router,
	}
	fmt.Printf("[devtools] listening on http://%s\n", s.opts.Addr)
	fmt.Printf("[devtools] open chrome://inspect and click '%s' to connect\n", s.opts.AppName)
	if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Stop gracefully shuts down the server.
func (s *Server) Stop(ctx context.Context) error {
	if s.httpSrv == nil {
		return nil
	}
	return s.httpSrv.Shutdown(ctx)
}

// AfterLayout is called by App after each layout pass to rebuild the tree snapshot.
// Must be called from the render goroutine.
func (s *Server) AfterLayout(tree *core.Tree, root widget.Widget, w, h float32) {
	snap := buildSnapshot(tree, root, w, h)
	s.snapMu.Lock()
	s.snapshot = snap
	s.snapMu.Unlock()

	// Notify connected sessions that the DOM tree may have been updated.
	s.broadcast("DOM.documentUpdated", map[string]any{})
}

// DrawOverlay draws the element highlight box if an element is currently inspected.
// Must be called from the render goroutine with an active command buffer
// (after root.Draw, before backend.Submit).
func (s *Server) DrawOverlay(buf *render.CommandBuffer) {
	s.overlayMu.Lock()
	id := s.highlightID
	s.overlayMu.Unlock()

	if id == core.InvalidElementID {
		return
	}

	snap := s.getSnapshot()
	if snap == nil {
		return
	}

	node, ok := snap.Nodes[id]
	if !ok {
		return
	}

	b := node.Bounds
	if b.Width <= 0 || b.Height <= 0 {
		return
	}

	// Content box — blue (matches Chrome DevTools default)
	buf.DrawOverlay(render.RectCmd{
		Bounds:      uimath.NewRect(b.X, b.Y, b.Width, b.Height),
		FillColor:   uimath.ColorHex("#0080ff26"),
		BorderColor: uimath.ColorHex("#0080ffb3"),
		BorderWidth: 1,
	}, 1000, 1)

	// Padding box — green, expanded outward by padding amounts
	p := node.Padding
	if p.Top+p.Right+p.Bottom+p.Left > 0 {
		buf.DrawOverlay(render.RectCmd{
			Bounds: uimath.NewRect(
				b.X-p.Left, b.Y-p.Top,
				b.Width+p.Left+p.Right, b.Height+p.Top+p.Bottom,
			),
			FillColor: uimath.ColorHex("#00cc6619"),
		}, 999, 1)
	}

	// Margin box — orange, expanded further by margins
	m := node.Margin
	if m.Top+m.Right+m.Bottom+m.Left > 0 {
		totalPadL := p.Left + node.Border.Left
		totalPadT := p.Top + node.Border.Top
		totalPadR := p.Right + node.Border.Right
		totalPadB := p.Bottom + node.Border.Bottom
		buf.DrawOverlay(render.RectCmd{
			Bounds: uimath.NewRect(
				b.X-totalPadL-m.Left, b.Y-totalPadT-m.Top,
				b.Width+totalPadL+totalPadR+m.Left+m.Right,
				b.Height+totalPadT+totalPadB+m.Top+m.Bottom,
			),
			FillColor: uimath.ColorHex("#ff880010"),
		}, 998, 1)
	}
}

// DrawOverlayLabel draws a Chrome-style info tooltip near the highlighted element.
// It shows the tag name with classes (e.g. "div.compose-avatar") and the element
// dimensions (e.g. "40 × 40"). Call this after DrawOverlay in the render loop.
// tr must be the app's text renderer; fontID must be the primary font used by the app.
// If tr is nil or fontID is invalid the label is silently skipped.
func (s *Server) DrawOverlayLabel(buf *render.CommandBuffer, tr *textrender.Renderer, fontID font.ID) {
	if tr == nil || fontID == font.InvalidFontID {
		return
	}
	s.overlayMu.Lock()
	id := s.highlightID
	s.overlayMu.Unlock()
	if id == core.InvalidElementID {
		return
	}
	snap := s.getSnapshot()
	if snap == nil {
		return
	}
	node, ok := snap.Nodes[id]
	if !ok {
		return
	}

	// Build label lines.
	tag := node.HTMLTag
	if tag == "" {
		tag = string(node.ElemType)
	}
	tagLine := tag
	if len(node.Classes) > 0 {
		tagLine += "." + strings.Join(node.Classes, ".")
	}
	dimsLine := fmt.Sprintf("%.0f × %.0f", node.Bounds.Width, node.Bounds.Height)

	const (
		fontSize  = 11.0
		padH      = 8.0  // horizontal padding inside label box
		padV      = 5.0  // vertical padding
		lineGap   = 3.0  // gap between the two text lines
		cornerR   = 4.0
	)
	shapeOpts := font.ShapeOptions{FontID: fontID, FontSize: fontSize}
	tm1 := tr.Measure(tagLine, shapeOpts)
	tm2 := tr.Measure(dimsLine, shapeOpts)

	lineH := tm1.Height
	if lineH <= 0 {
		lineH = fontSize + 2
	}
	boxW := tm1.Width
	if tm2.Width > boxW {
		boxW = tm2.Width
	}
	boxW += padH * 2
	boxH := lineH*2 + lineGap + padV*2

	// Position the label below the element; flip above if it would go off-screen.
	b := node.Bounds
	lx := b.X
	ly := b.Y + b.Height + 6
	if ly+boxH > snap.ViewHeight && b.Y-boxH-6 >= 0 {
		ly = b.Y - boxH - 6
	}
	// Clamp horizontally.
	if lx+boxW > snap.ViewWidth {
		lx = snap.ViewWidth - boxW - 4
	}
	if lx < 0 {
		lx = 0
	}

	// Dark background (matches Chrome DevTools tooltip style).
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(lx, ly, boxW, boxH),
		FillColor: uimath.ColorHex("#1a1a1aee"),
		Corners:   uimath.CornersAll(cornerR),
	}, 1002, 1)

	textX := lx + padH
	// Line 1: tag.class in blue (Chrome uses #9cdcfe for tag, #d7ba7d for class —
	// we keep it simple with a single highlight colour).
	tr.DrawText(buf, tagLine, textrender.DrawOptions{
		ShapeOpts: shapeOpts,
		OriginX:   textX,
		OriginY:   ly + padV + lineH*0.8,
		Color:     uimath.ColorHex("#5db0d7"), // Chrome tag-name blue
		ZOrder:    1003,
		Opacity:   1,
		Overlay:   true,
	})
	// Line 2: dimensions in grey.
	tr.DrawText(buf, dimsLine, textrender.DrawOptions{
		ShapeOpts: shapeOpts,
		OriginX:   textX,
		OriginY:   ly + padV + lineH*0.8 + lineH + lineGap,
		Color:     uimath.ColorHex("#cccccc"),
		ZOrder:    1003,
		Opacity:   1,
		Overlay:   true,
	})
}

// Log adds an entry to the DevTools Console panel.
//
//	srv.Log("info",  "other", "layout complete")
//	srv.Log("error", "other", "texture failed to load")
func (s *Server) Log(level, source, text string) {
	entry := map[string]any{
		"entry": LogEntry{
			Source:    source,
			Level:     level,
			Text:      text,
			Timestamp: float64(time.Now().UnixMilli()),
		},
	}
	s.sessionsMu.Lock()
	sessions := make([]*Session, len(s.sessions))
	copy(sessions, s.sessions)
	s.sessionsMu.Unlock()
	for _, sess := range sessions {
		if sess.logEnabled {
			sess.sendEvent("Log.entryAdded", entry)
		}
	}
}

// ----- internal helpers -----

func (s *Server) getSnapshot() *Snapshot {
	s.snapMu.RLock()
	defer s.snapMu.RUnlock()
	return s.snapshot
}

func (s *Server) setHighlight(id core.ElementID) {
	s.overlayMu.Lock()
	changed := s.highlightID != id
	s.highlightID = id
	dirty := s.markDirty
	s.overlayMu.Unlock()
	if changed && dirty != nil {
		dirty()
	}
}

// Attach is called by App to give the server a way to trigger redraws.
// It is called automatically when DevTools is set in AppOptions.
func (s *Server) Attach(markDirty func()) {
	s.overlayMu.Lock()
	s.markDirty = markDirty
	s.overlayMu.Unlock()
}

func (s *Server) broadcast(method string, params any) {
	s.sessionsMu.Lock()
	sessions := make([]*Session, len(s.sessions))
	copy(sessions, s.sessions)
	s.sessionsMu.Unlock()

	for _, sess := range sessions {
		sess.sendEvent(method, params)
	}
}

func (s *Server) addSession(sess *Session) {
	s.sessionsMu.Lock()
	s.sessions = append(s.sessions, sess)
	s.sessionsMu.Unlock()
}

func (s *Server) removeSession(sess *Session) {
	s.sessionsMu.Lock()
	for i, s2 := range s.sessions {
		if s2 == sess {
			s.sessions = append(s.sessions[:i], s.sessions[i+1:]...)
			break
		}
	}
	s.sessionsMu.Unlock()
}

// ----- HTTP routes -----

var wsUpgrader = &websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 64 * 1024,
}

const targetID = "main-ui"

func (s *Server) setupRoutes() {
	// Chrome DevTools discovery endpoints
	s.router.GET("/json", s.jsonList)
	s.router.GET("/json/list", s.jsonList)
	s.router.GET("/json/version", s.jsonVersion)
	s.router.GET("/json/protocol", s.jsonProtocol)

	// CDP WebSocket endpoint
	s.router.GET("/devtools/page/:id", func(c *gin.Context) {
		conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		sess := newSession(s, conn)
		s.addSession(sess)
		defer s.removeSession(sess)
		sess.run()
	})
}

func (s *Server) jsonList(c *gin.Context) {
	host := c.Request.Host
	if host == "" {
		host = s.localHostPort()
	}
	c.JSON(http.StatusOK, []map[string]any{
		{
			"description":          "",
			"devtoolsFrontendUrl":  fmt.Sprintf("/devtools/inspector.html?ws=%s/devtools/page/%s", host, targetID),
			"faviconUrl":           "",
			"id":                   targetID,
			"title":                s.opts.AppName,
			"type":                 "page",
			"url":                  "ui://app",
			"webSocketDebuggerUrl": fmt.Sprintf("ws://%s/devtools/page/%s", host, targetID),
		},
	})
}

func (s *Server) jsonVersion(c *gin.Context) {
	host := c.Request.Host
	if host == "" {
		host = s.localHostPort()
	}
	c.JSON(http.StatusOK, map[string]string{
		"Browser":              "GoUI/1.0",
		"Protocol-Version":     "1.3",
		"User-Agent":           "GoUI DevTools Server",
		"V8-Version":           "",
		"WebKit-Version":       "",
		"webSocketDebuggerUrl": fmt.Sprintf("ws://%s/devtools/page/%s", host, targetID),
	})
}

func (s *Server) jsonProtocol(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]any{"domains": []any{}})
}

// localHostPort returns "localhost:<port>" derived from opts.Addr.
// Used as a fallback when the HTTP Host header is absent (e.g. in tests).
func (s *Server) localHostPort() string {
	if _, port, err := net.SplitHostPort(s.opts.Addr); err == nil {
		return "localhost:" + port
	}
	return "localhost:9222"
}

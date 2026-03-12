package devtools

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/kasuganosora/ui/core"
)

// ----- Page domain -----

func (s *Session) handlePage(req Request) {
	switch req.Method {
	case "Page.enable":
		s.pageEnabled = true
		s.sendResult(req.ID, map[string]any{})
		// Emit a synthetic load event so DevTools considers the page ready.
		s.sendEvent("Page.loadEventFired", map[string]any{
			"timestamp": nowSeconds(),
		})
		s.sendEvent("Page.domContentEventFired", map[string]any{
			"timestamp": nowSeconds(),
		})

	case "Page.disable":
		s.pageEnabled = false
		s.sendResult(req.ID, map[string]any{})

	case "Page.getFrameTree":
		s.sendResult(req.ID, map[string]any{
			"frameTree": frameTree(),
		})

	case "Page.getResourceTree":
		s.sendResult(req.ID, map[string]any{
			"frameTree": map[string]any{
				"frame":     frameInfo(),
				"resources": []any{},
			},
		})

	case "Page.getLayoutMetrics":
		snap := s.srv.getSnapshot()
		vw, vh := float32(1280), float32(720)
		if snap != nil {
			vw, vh = snap.ViewWidth, snap.ViewHeight
		}
		s.sendResult(req.ID, map[string]any{
			"layoutViewport": map[string]any{
				"pageX":        0,
				"pageY":        0,
				"clientWidth":  vw,
				"clientHeight": vh,
			},
			"visualViewport": map[string]any{
				"offsetX":      0,
				"offsetY":      0,
				"pageX":        0,
				"pageY":        0,
				"clientWidth":  vw,
				"clientHeight": vh,
				"scale":        1,
				"zoom":         1,
			},
			"contentSize": map[string]any{
				"x":      0,
				"y":      0,
				"width":  vw,
				"height": vh,
			},
		})

	case "Page.reload":
		if s.srv.markDirty != nil {
			s.srv.markDirty()
		}
		s.sendResult(req.ID, map[string]any{})

	case "Page.navigate":
		s.sendResult(req.ID, map[string]any{
			"frameId":  "main",
			"loaderId": "loader-1",
		})

	case "Page.stopLoading",
		"Page.setLifecycleEventsEnabled",
		"Page.createIsolatedWorld",
		"Page.addScriptToEvaluateOnNewDocument",
		"Page.setBypassCSP",
		"Page.captureScreenshot":
		s.sendResult(req.ID, map[string]any{})

	default:
		s.sendResult(req.ID, map[string]any{})
	}
}

// ----- Runtime domain -----
// We expose a minimal read-only JavaScript runtime that supports console output.

func (s *Session) handleRuntime(req Request) {
	switch req.Method {
	case "Runtime.enable":
		s.runtimeEnabled = true
		s.sendResult(req.ID, map[string]any{})
		// Signal that the execution context is ready.
		s.sendEvent("Runtime.executionContextCreated", map[string]any{
			"context": map[string]any{
				"id":      1,
				"origin":  "ui://app",
				"name":    "main",
				"auxData": map[string]any{"isDefault": true, "frameId": "main"},
			},
		})

	case "Runtime.disable":
		s.runtimeEnabled = false
		s.sendResult(req.ID, map[string]any{})

	case "Runtime.evaluate":
		// We don't execute JavaScript; return undefined for all expressions.
		var p struct {
			Expression string `json:"expression"`
		}
		_ = json.Unmarshal(req.Params, &p)
		s.sendResult(req.ID, map[string]any{
			"result": map[string]any{
				"type":        "undefined",
				"description": "undefined",
			},
		})

	case "Runtime.callFunctionOn":
		var p struct {
			ObjectID            string `json:"objectId"`
			FunctionDeclaration string `json:"functionDeclaration"`
		}
		_ = json.Unmarshal(req.Params, &p)

		// Handle node remote objects (objectId = "node-<elementID>").
		// Detect common DevTools property reads (e.g. outerHTML, innerHTML, textContent)
		// and return the appropriate value so that "Copy > outerHTML" works.
		if strings.HasPrefix(p.ObjectID, "node-") {
			var nodeIDVal int
			if _, err := fmt.Sscanf(p.ObjectID, "node-%d", &nodeIDVal); err == nil {
				nodeID := core.ElementID(nodeIDVal)
				snap := s.srv.getSnapshot()
				fn := p.FunctionDeclaration
				switch {
				case snap != nil && strings.Contains(fn, "outerHTML"):
					html := buildOuterHTML(snap, nodeID, 0)
					s.sendResult(req.ID, map[string]any{
						"result": map[string]any{"type": "string", "value": html},
					})
					return
				case snap != nil && strings.Contains(fn, "innerHTML"):
					node := snap.Nodes[nodeID]
					var inner strings.Builder
					if node != nil {
						for _, cid := range node.ChildIDs {
							inner.WriteString(buildOuterHTML(snap, cid, 0))
						}
						if node.Text != "" && len(node.ChildIDs) == 0 {
							inner.WriteString(htmlEscaper.Replace(node.Text))
						}
					}
					s.sendResult(req.ID, map[string]any{
						"result": map[string]any{"type": "string", "value": inner.String()},
					})
					return
				case snap != nil && strings.Contains(fn, "textContent"):
					text := collectTextContent(snap, nodeID)
					s.sendResult(req.ID, map[string]any{
						"result": map[string]any{"type": "string", "value": text},
					})
					return
				}
			}
		}
		s.sendResult(req.ID, map[string]any{
			"result": map[string]any{"type": "undefined"},
		})

	case "Runtime.getProperties":
		s.sendResult(req.ID, map[string]any{
			"result":             []any{},
			"internalProperties": []any{},
		})

	case "Runtime.releaseObject",
		"Runtime.releaseObjectGroup",
		"Runtime.runIfWaitingForDebugger",
		"Runtime.discardConsoleEntries":
		s.sendResult(req.ID, map[string]any{})

	default:
		s.sendResult(req.ID, map[string]any{})
	}
}

// ----- Log domain -----

func (s *Session) handleLog(req Request) {
	switch req.Method {
	case "Log.enable":
		s.logEnabled = true
		s.sendResult(req.ID, map[string]any{})

	case "Log.disable":
		s.logEnabled = false
		s.sendResult(req.ID, map[string]any{})

	case "Log.clear":
		s.sendResult(req.ID, map[string]any{})

	case "Log.startViolationsReport",
		"Log.stopViolationsReport":
		s.sendResult(req.ID, map[string]any{})

	default:
		s.sendResult(req.ID, map[string]any{})
	}
}

// ----- helpers -----

func frameInfo() map[string]any {
	return map[string]any{
		"id":             "main",
		"loaderId":       "loader-1",
		"url":            "ui://app",
		"securityOrigin": "ui://app",
		"mimeType":       "text/html",
	}
}

func frameTree() map[string]any {
	return map[string]any{
		"frame":       frameInfo(),
		"childFrames": []any{},
	}
}

func nowSeconds() float64 {
	return float64(time.Now().UnixMilli()) / 1000
}

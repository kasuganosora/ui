package devtools

import (
	"encoding/json"

	"github.com/kasuganosora/ui/core"
)

// rgbaColor is a CDP RGBA color (used in highlight configs).
type rgbaColor struct {
	R int     `json:"r"`
	G int     `json:"g"`
	B int     `json:"b"`
	A float64 `json:"a"`
}

func (s *Session) handleOverlay(req Request) {
	switch req.Method {
	case "Overlay.enable":
		s.overlayEnabled = true
		s.sendResult(req.ID, map[string]any{})

	case "Overlay.disable":
		s.overlayEnabled = false
		s.srv.setHighlight(core.InvalidElementID)
		s.sendResult(req.ID, map[string]any{})

	// highlightNode — highlight the element box in the live window.
	case "Overlay.highlightNode":
		var p struct {
			NodeID       int `json:"nodeId"`
			BackendNodeID int `json:"backendNodeId"`
		}
		_ = json.Unmarshal(req.Params, &p)

		id := core.ElementID(p.NodeID)
		if id == 0 && p.BackendNodeID != 0 {
			id = core.ElementID(p.BackendNodeID)
		}
		s.srv.setHighlight(id)
		s.sendResult(req.ID, map[string]any{})

	// highlightRect — highlight an arbitrary rectangle (not tied to a node).
	case "Overlay.highlightRect":
		// We don't support arbitrary rect highlights (no node ID),
		// so clear the current highlight and acknowledge.
		s.srv.setHighlight(core.InvalidElementID)
		s.sendResult(req.ID, map[string]any{})

	case "Overlay.hideHighlight":
		s.srv.setHighlight(core.InvalidElementID)
		s.sendResult(req.ID, map[string]any{})

	// setInspectMode — enables mouse-hover inspect in the application window.
	// We return success; actual pointer-based picking would require UI-side hooks.
	case "Overlay.setInspectMode":
		var p struct {
			Mode string `json:"mode"` // "searchForNode" | "none"
		}
		_ = json.Unmarshal(req.Params, &p)
		if p.Mode == "none" {
			s.srv.setHighlight(core.InvalidElementID)
		}
		s.sendResult(req.ID, map[string]any{})

	// Miscellaneous Overlay methods we acknowledge but don't implement.
	case "Overlay.setPausedInDebuggerMessage",
		"Overlay.setShowFPSCounter",
		"Overlay.setShowPaintRects",
		"Overlay.setShowLayoutShiftRegions",
		"Overlay.setShowScrollBottleneckRects",
		"Overlay.setShowHitTestBorders",
		"Overlay.setShowWebVitals",
		"Overlay.setShowViewportSizeOnResize",
		"Overlay.setShowAdHighlights",
		"Overlay.getHighlightObjectForTest":
		s.sendResult(req.ID, map[string]any{})

	default:
		s.sendResult(req.ID, map[string]any{})
	}
}

package devtools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kasuganosora/ui/core"
)

var htmlEscaper = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	`"`, "&#34;",
)

// CDP DOM node type constants (W3C spec).
const (
	nodeTypeElement  = 1
	nodeTypeText     = 3
	nodeTypeDocument = 9
)

// domNode is the CDP DOM.Node wire structure.
type domNode struct {
	NodeID         int        `json:"nodeId"`
	BackendNodeID  int        `json:"backendNodeId"`
	ParentID       int        `json:"parentId,omitempty"`
	NodeType       int        `json:"nodeType"`
	NodeName       string     `json:"nodeName"`
	LocalName      string     `json:"localName"`
	NodeValue      string     `json:"nodeValue"`
	ChildNodeCount int        `json:"childNodeCount"`
	Children       []*domNode `json:"children,omitempty"`
	Attributes     []string   `json:"attributes"` // interleaved key/value pairs
	FrameID        string     `json:"frameId,omitempty"`
}

// syntheticDocumentNodeID is the node ID we use for the virtual document root.
// We use math.MaxInt32 so it can't clash with real ElementIDs (which start at 1).
const syntheticDocumentNodeID = 2147483647

func (s *Session) handleDOM(req Request) {
	switch req.Method {
	case "DOM.enable":
		s.domEnabled = true
		s.sendResult(req.ID, map[string]any{})

	case "DOM.disable":
		s.domEnabled = false
		s.sendResult(req.ID, map[string]any{})

	case "DOM.getDocument":
		var p struct {
			Depth  int  `json:"depth"`  // -1 = full tree
			Pierce bool `json:"pierce"` // ignored
		}
		_ = json.Unmarshal(req.Params, &p)
		depth := p.Depth
		if depth == 0 {
			depth = 2 // DevTools default for initial load
		}

		snap := s.srv.getSnapshot()
		if snap == nil {
			s.sendResult(req.ID, map[string]any{"root": s.emptyDocument()})
			return
		}

		rootElem := s.buildDOMNode(snap, snap.Root, depth)
		doc := &domNode{
			NodeID:         syntheticDocumentNodeID,
			BackendNodeID:  syntheticDocumentNodeID,
			NodeType:       nodeTypeDocument,
			NodeName:       "#document",
			LocalName:      "",
			NodeValue:      "",
			ChildNodeCount: 1,
			Children:       []*domNode{rootElem},
			FrameID:        "main",
		}
		s.sendResult(req.ID, map[string]any{"root": doc})

	case "DOM.requestChildNodes":
		var p struct {
			NodeID int `json:"nodeId"`
			Depth  int `json:"depth"`
		}
		_ = json.Unmarshal(req.Params, &p)
		depth := p.Depth
		if depth == 0 {
			depth = 1
		}

		snap := s.srv.getSnapshot()
		if snap == nil {
			s.sendResult(req.ID, map[string]any{})
			return
		}

		nodeID := core.ElementID(p.NodeID)
		node, ok := snap.Nodes[nodeID]
		if !ok {
			s.sendResult(req.ID, map[string]any{})
			return
		}

		children := make([]*domNode, 0, len(node.ChildIDs)+1)
		for _, cid := range node.ChildIDs {
			if dn := s.buildDOMNode(snap, cid, depth-1); dn != nil {
				children = append(children, dn)
			}
		}
		// Leaf text node: add synthetic #text child (mirrors browser DOM).
		if node.Text != "" && len(node.ChildIDs) == 0 {
			children = append(children, &domNode{
				NodeID:        -int(nodeID),
				BackendNodeID: -int(nodeID),
				ParentID:      int(nodeID),
				NodeType:      nodeTypeText,
				NodeName:      "#text",
				LocalName:     "",
				NodeValue:     node.Text,
				Attributes:    []string{},
			})
		}
		s.sendEvent("DOM.setChildNodes", map[string]any{
			"parentId": p.NodeID,
			"nodes":    children,
		})
		s.sendResult(req.ID, map[string]any{})

	case "DOM.getBoxModel":
		var p struct {
			NodeID int `json:"nodeId"`
		}
		_ = json.Unmarshal(req.Params, &p)

		snap := s.srv.getSnapshot()
		if snap == nil {
			s.sendError(req.ID, -32000, "node not found")
			return
		}
		node, ok := snap.Nodes[core.ElementID(p.NodeID)]
		if !ok {
			s.sendError(req.ID, -32000, "node not found")
			return
		}

		b := node.Bounds
		pad := node.Padding
		bdr := node.Border
		mar := node.Margin

		// CDP quads: 8 floats [x1,y1, x2,y1, x2,y2, x1,y2] (clockwise from top-left)
		contentQ := quad(b.X, b.Y, b.Width, b.Height)
		paddingQ := quad(b.X-pad.Left, b.Y-pad.Top,
			b.Width+pad.Left+pad.Right, b.Height+pad.Top+pad.Bottom)
		borderQ := quad(b.X-pad.Left-bdr.Left, b.Y-pad.Top-bdr.Top,
			b.Width+pad.Left+pad.Right+bdr.Left+bdr.Right,
			b.Height+pad.Top+pad.Bottom+bdr.Top+bdr.Bottom)
		marginQ := quad(b.X-pad.Left-bdr.Left-mar.Left, b.Y-pad.Top-bdr.Top-mar.Top,
			b.Width+pad.Left+pad.Right+bdr.Left+bdr.Right+mar.Left+mar.Right,
			b.Height+pad.Top+pad.Bottom+bdr.Top+bdr.Bottom+mar.Top+mar.Bottom)

		s.sendResult(req.ID, map[string]any{
			"model": map[string]any{
				"content": contentQ,
				"padding": paddingQ,
				"border":  borderQ,
				"margin":  marginQ,
				"width":   int(b.Width + 0.5),
				"height":  int(b.Height + 0.5),
			},
		})

	case "DOM.describeNode":
		var p struct {
			NodeID int `json:"nodeId"`
		}
		_ = json.Unmarshal(req.Params, &p)

		snap := s.srv.getSnapshot()
		if snap == nil {
			s.sendError(req.ID, -32000, "node not found")
			return
		}
		node, ok := snap.Nodes[core.ElementID(p.NodeID)]
		if !ok {
			s.sendError(req.ID, -32000, "node not found")
			return
		}
		s.sendResult(req.ID, map[string]any{
			"node": s.buildDOMNode(snap, node.ID, 0),
		})

	case "DOM.getNodeForLocation":
		var p struct {
			X int `json:"x"`
			Y int `json:"y"`
		}
		_ = json.Unmarshal(req.Params, &p)

		snap := s.srv.getSnapshot()
		if snap == nil {
			s.sendError(req.ID, -32000, "no snapshot")
			return
		}
		id := hitTestSnap(snap, snap.Root, float32(p.X), float32(p.Y))
		if id == core.InvalidElementID {
			s.sendError(req.ID, -32000, "no node at location")
			return
		}
		s.sendResult(req.ID, map[string]any{
			"nodeId":        int(id),
			"backendNodeId": int(id),
			"frameId":       "main",
		})

	case "DOM.getOuterHTML":
		var p struct {
			NodeID int `json:"nodeId"`
		}
		_ = json.Unmarshal(req.Params, &p)
		snap := s.srv.getSnapshot()
		if snap == nil {
			s.sendResult(req.ID, map[string]any{"outerHTML": ""})
			return
		}
		html := buildOuterHTML(snap, core.ElementID(p.NodeID), 0)
		s.sendResult(req.ID, map[string]any{"outerHTML": html})

	case "DOM.setOuterHTML":
		// We don't support live editing; just acknowledge.
		s.sendResult(req.ID, map[string]any{})

	case "DOM.setInspectedNode":
		// Acknowledged; used by DevTools when pinning an element.
		s.sendResult(req.ID, map[string]any{})

	case "DOM.resolveNode":
		var p struct {
			NodeID int `json:"nodeId"`
		}
		_ = json.Unmarshal(req.Params, &p)
		s.sendResult(req.ID, map[string]any{
			"object": map[string]any{
				"type":     "object",
				"subtype":  "node",
				"objectId": fmt.Sprintf("node-%d", p.NodeID),
				"description": fmt.Sprintf("node#%d", p.NodeID),
			},
		})

	case "DOM.querySelector":
		// Minimal implementation: class/tag selector on snapshot
		var p struct {
			NodeID   int    `json:"nodeId"`
			Selector string `json:"selector"`
		}
		_ = json.Unmarshal(req.Params, &p)
		snap := s.srv.getSnapshot()
		if snap != nil {
			if id := querySelector(snap, core.ElementID(p.NodeID), p.Selector); id != core.InvalidElementID {
				s.sendResult(req.ID, map[string]any{"nodeId": int(id)})
				return
			}
		}
		s.sendResult(req.ID, map[string]any{"nodeId": 0})

	case "DOM.querySelectorAll":
		var p struct {
			NodeID   int    `json:"nodeId"`
			Selector string `json:"selector"`
		}
		_ = json.Unmarshal(req.Params, &p)
		snap := s.srv.getSnapshot()
		var ids []int
		if snap != nil {
			for _, id := range querySelectorAll(snap, core.ElementID(p.NodeID), p.Selector) {
				ids = append(ids, int(id))
			}
		}
		if ids == nil {
			ids = []int{}
		}
		s.sendResult(req.ID, map[string]any{"nodeIds": ids})

	case "DOM.pushNodesByBackendIdsToFrontend":
		var p struct {
			BackendNodeIDs []int `json:"backendNodeIds"`
		}
		_ = json.Unmarshal(req.Params, &p)
		nodeIDs := make([]int, len(p.BackendNodeIDs))
		copy(nodeIDs, p.BackendNodeIDs)
		s.sendResult(req.ID, map[string]any{"nodeIds": nodeIDs})

	case "DOM.getAttributes":
		var p struct {
			NodeID int `json:"nodeId"`
		}
		_ = json.Unmarshal(req.Params, &p)
		snap := s.srv.getSnapshot()
		attrs := []string{}
		if snap != nil {
			if node, ok := snap.Nodes[core.ElementID(p.NodeID)]; ok {
				attrs = buildAttributes(node)
			}
		}
		s.sendResult(req.ID, map[string]any{"attributes": attrs})

	default:
		s.sendResult(req.ID, map[string]any{})
	}
}

// buildDOMNode converts a snapshot node to a CDP domNode.
// depth controls recursive child expansion: 0 = leaf (no children array),
// -1 = full tree, positive = levels remaining.
func (s *Session) buildDOMNode(snap *Snapshot, id core.ElementID, depth int) *domNode {
	node, ok := snap.Nodes[id]
	if !ok {
		return nil
	}

	// Prefer original HTML tag name (e.g. "span", "p", "h1") over internal
	// element type (e.g. "text"). Fall back to element type for programmatic widgets.
	localName := node.HTMLTag
	if localName == "" {
		localName = string(node.ElemType)
	}
	nodeName := strings.ToUpper(localName)

	// A #text child is added for leaf elements that carry text content
	// (e.g. <span>foo</span> → SPAN > #text "foo"), matching browser DOM structure.
	hasTextChild := node.Text != "" && len(node.ChildIDs) == 0
	childCount := len(node.ChildIDs)
	if hasTextChild {
		childCount = 1
	}

	dn := &domNode{
		NodeID:         int(id),
		BackendNodeID:  int(id),
		ParentID:       int(node.ParentID),
		NodeType:       nodeTypeElement,
		NodeName:       nodeName,
		LocalName:      localName,
		NodeValue:      "",
		ChildNodeCount: childCount,
		Attributes:     buildAttributes(node),
	}

	// Expand children if depth allows
	if depth != 0 {
		nextDepth := depth - 1
		if depth < 0 {
			nextDepth = -1 // full tree
		}

		// Real widget children
		for _, cid := range node.ChildIDs {
			if child := s.buildDOMNode(snap, cid, nextDepth); child != nil {
				dn.Children = append(dn.Children, child)
			}
		}

		// Synthetic #text child for leaf text elements
		if hasTextChild {
			dn.Children = append(dn.Children, &domNode{
				// Use a synthetic node ID: negate the parent ID to avoid clashing
				// with real element IDs (which are always positive).
				NodeID:        -int(id),
				BackendNodeID: -int(id),
				ParentID:      int(id),
				NodeType:      nodeTypeText,
				NodeName:      "#text",
				LocalName:     "",
				NodeValue:     node.Text,
				Attributes:    []string{},
			})
		}
	}

	return dn
}

func (s *Session) emptyDocument() *domNode {
	return &domNode{
		NodeID:         syntheticDocumentNodeID,
		BackendNodeID:  syntheticDocumentNodeID,
		NodeType:       nodeTypeDocument,
		NodeName:       "#document",
		LocalName:      "",
		NodeValue:      "",
		ChildNodeCount: 0,
	}
}

// buildAttributes returns CDP attribute pairs ["key","value",...].
func buildAttributes(node *NodeSnapshot) []string {
	var attrs []string
	if node.IDAttr != "" {
		attrs = append(attrs, "id", node.IDAttr)
	}
	if len(node.Classes) > 0 {
		attrs = append(attrs, "class", strings.Join(node.Classes, " "))
	}
	// Bounding-box data attributes (useful for quick inspection in the Attributes pane)
	b := node.Bounds
	attrs = append(attrs,
		"data-x", fmt.Sprintf("%.0f", b.X),
		"data-y", fmt.Sprintf("%.0f", b.Y),
		"data-w", fmt.Sprintf("%.0f", b.Width),
		"data-h", fmt.Sprintf("%.0f", b.Height),
	)
	return attrs
}

// quad returns a CDP quad [x1,y1,x2,y1,x2,y2,x1,y2].
func quad(x, y, w, h float32) []float64 {
	return []float64{
		float64(x), float64(y),
		float64(x + w), float64(y),
		float64(x + w), float64(y + h),
		float64(x), float64(y + h),
	}
}

// hitTestSnap finds the deepest node in the snapshot at the given screen position.
func hitTestSnap(snap *Snapshot, id core.ElementID, x, y float32) core.ElementID {
	node, ok := snap.Nodes[id]
	if !ok {
		return core.InvalidElementID
	}
	b := node.Bounds
	if x < b.X || x > b.X+b.Width || y < b.Y || y > b.Y+b.Height {
		return core.InvalidElementID
	}
	// Check children in reverse (last painted = visually on top)
	for i := len(node.ChildIDs) - 1; i >= 0; i-- {
		if hit := hitTestSnap(snap, node.ChildIDs[i], x, y); hit != core.InvalidElementID {
			return hit
		}
	}
	return id
}

// querySelector returns the first node under root that matches selector.
// Supports: tag, .class, #id (treated as data-id attr), * (wildcard).
func querySelector(snap *Snapshot, root core.ElementID, sel string) core.ElementID {
	results := querySelectorAll(snap, root, sel)
	if len(results) > 0 {
		return results[0]
	}
	return core.InvalidElementID
}

func querySelectorAll(snap *Snapshot, root core.ElementID, sel string) []core.ElementID {
	sel = strings.TrimSpace(sel)
	var results []core.ElementID
	walkSnap(snap, root, func(id core.ElementID) bool {
		node := snap.Nodes[id]
		if node == nil {
			return true
		}
		if matchesSelector(node, sel) {
			results = append(results, id)
		}
		return true
	})
	return results
}

func matchesSelector(node *NodeSnapshot, sel string) bool {
	if sel == "*" {
		return true
	}
	if strings.HasPrefix(sel, "#") {
		return node.IDAttr == sel[1:]
	}
	if strings.HasPrefix(sel, ".") {
		cls := sel[1:]
		for _, c := range node.Classes {
			if c == cls {
				return true
			}
		}
		return false
	}
	// Tag selector: match against HTMLTag first (e.g. "span"), then ElemType.
	if node.HTMLTag != "" && strings.EqualFold(node.HTMLTag, sel) {
		return true
	}
	return strings.EqualFold(string(node.ElemType), sel)
}

func walkSnap(snap *Snapshot, id core.ElementID, fn func(core.ElementID) bool) {
	if !fn(id) {
		return
	}
	node, ok := snap.Nodes[id]
	if !ok {
		return
	}
	for _, cid := range node.ChildIDs {
		walkSnap(snap, cid, fn)
	}
}

// buildOuterHTML serialises a snapshot node (and its subtree) to an HTML string.
// indent is the current nesting level for pretty-printing.
func buildOuterHTML(snap *Snapshot, id core.ElementID, indent int) string {
	node, ok := snap.Nodes[id]
	if !ok {
		return ""
	}

	// Determine tag name (use original HTML tag if available)
	tag := node.HTMLTag
	if tag == "" {
		tag = string(node.ElemType)
	}

	pad := strings.Repeat("  ", indent)

	// Build opening tag with id and class attributes
	var sb strings.Builder
	sb.WriteString(pad)
	sb.WriteByte('<')
	sb.WriteString(tag)
	if node.IDAttr != "" {
		sb.WriteString(` id="`)
		sb.WriteString(htmlEscaper.Replace(node.IDAttr))
		sb.WriteByte('"')
	}
	if len(node.Classes) > 0 {
		sb.WriteString(` class="`)
		sb.WriteString(htmlEscaper.Replace(strings.Join(node.Classes, " ")))
		sb.WriteByte('"')
	}
	sb.WriteByte('>')

	hasChildren := len(node.ChildIDs) > 0
	hasText := node.Text != "" && !hasChildren

	if hasText {
		// Inline text: <span>content</span>
		sb.WriteString(htmlEscaper.Replace(node.Text))
		sb.WriteString("</")
		sb.WriteString(tag)
		sb.WriteByte('>')
		return sb.String()
	}

	if !hasChildren {
		// Self-closing-style: void or empty element
		switch tag {
		case "input", "img", "br", "hr":
			// Replace > with />
			s := sb.String()
			return s[:len(s)-1] + " />"
		}
		sb.WriteString("</")
		sb.WriteString(tag)
		sb.WriteByte('>')
		return sb.String()
	}

	// Block children
	sb.WriteByte('\n')
	for _, cid := range node.ChildIDs {
		sb.WriteString(buildOuterHTML(snap, cid, indent+1))
		sb.WriteByte('\n')
	}
	sb.WriteString(pad)
	sb.WriteString("</")
	sb.WriteString(tag)
	sb.WriteByte('>')
	return sb.String()
}

// collectTextContent recursively gathers all text content from a node and its
// descendants, matching the browser's Node.textContent semantics.
func collectTextContent(snap *Snapshot, id core.ElementID) string {
	node, ok := snap.Nodes[id]
	if !ok {
		return ""
	}
	if len(node.ChildIDs) == 0 {
		return node.Text
	}
	var sb strings.Builder
	if node.Text != "" {
		sb.WriteString(node.Text)
	}
	for _, cid := range node.ChildIDs {
		sb.WriteString(collectTextContent(snap, cid))
	}
	return sb.String()
}

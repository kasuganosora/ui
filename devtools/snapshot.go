// Package devtools provides a Chrome DevTools Protocol (CDP) server
// that lets Chrome DevTools inspect, debug, and profile the UI framework.
//
// # Quick start
//
//	srv := devtools.NewServer(devtools.Options{Addr: ":9222", AppName: "My App"})
//	app, _ := ui.NewApp(ui.AppOptions{DevTools: srv, ...})
//	go srv.Start()
//	app.Run()
//
// Then navigate to chrome://inspect in Chrome and click "inspect" next to "My App".
package devtools

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/widget"
)

// NodeSnapshot is a point-in-time immutable copy of one element's state.
type NodeSnapshot struct {
	ID       core.ElementID
	ParentID core.ElementID
	ChildIDs []core.ElementID

	// From core.Element
	ElemType core.ElementType
	HTMLTag  string   // original HTML tag (e.g. "span", "p", "h1"); empty for programmatic widgets
	Classes  []string
	Text     string // text content (set by widget via element property "text")

	// Layout results (absolute screen coordinates after layout)
	Bounds  uimath.Rect
	Padding uimath.Edges
	Margin  uimath.Edges
	Border  uimath.Edges

	// Pre-layout CSS style from widget
	Style layout.Style
}

// Snapshot holds the full widget tree state at a single point in time.
// Built on the render goroutine; read from DevTools session goroutines under a lock.
type Snapshot struct {
	Nodes      map[core.ElementID]*NodeSnapshot
	Root       core.ElementID
	ViewWidth  float32
	ViewHeight float32
}

// buildSnapshot walks the widget tree depth-first and captures a Snapshot.
// Must be called from the render goroutine (no locking needed here).
func buildSnapshot(tree *core.Tree, root widget.Widget, vw, vh float32) *Snapshot {
	snap := &Snapshot{
		Nodes:      make(map[core.ElementID]*NodeSnapshot),
		Root:       root.ElementID(),
		ViewWidth:  vw,
		ViewHeight: vh,
	}
	buildSnapshotNode(snap, tree, root, core.InvalidElementID)
	return snap
}

func buildSnapshotNode(snap *Snapshot, tree *core.Tree, w widget.Widget, parentID core.ElementID) {
	id := w.ElementID()
	elem := tree.Get(id)
	if elem == nil {
		return
	}

	lr := elem.Layout()
	htmlTag := ""
	if v, ok := elem.Property("html-tag"); ok {
		if s, ok := v.(string); ok {
			htmlTag = s
		}
	}
	node := &NodeSnapshot{
		ID:       id,
		ParentID: parentID,
		ElemType: elem.Type(),
		HTMLTag:  htmlTag,
		Classes:  append([]string(nil), elem.Classes()...),
		Text:     elem.TextContent(),
		Bounds:   lr.Bounds,
		Padding:  lr.Padding,
		Margin:   lr.Margin,
		Border:   lr.Border,
		Style:    w.Style(),
	}

	children := w.Children()
	node.ChildIDs = make([]core.ElementID, 0, len(children))
	for _, child := range children {
		node.ChildIDs = append(node.ChildIDs, child.ElementID())
	}

	snap.Nodes[id] = node

	for _, child := range children {
		buildSnapshotNode(snap, tree, child, id)
	}
}

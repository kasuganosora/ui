package devtools

import (
	"strings"
	"testing"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
)

// ─── buildAttributes ─────────────────────────────────────────────────────────

func TestBuildAttributes_Empty(t *testing.T) {
	n := &NodeSnapshot{Bounds: uimath.NewRect(0, 0, 100, 50)}
	attrs := buildAttributes(n)
	// No id, no class — only data-x/y/w/h
	if len(attrs) != 8 {
		t.Fatalf("want 8 attrs (4 pairs), got %d", len(attrs))
	}
	if attrs[0] != "data-x" || attrs[1] != "0" {
		t.Errorf("unexpected data-x pair: %v %v", attrs[0], attrs[1])
	}
}

func TestBuildAttributes_WithIDAndClass(t *testing.T) {
	n := &NodeSnapshot{
		IDAttr:  "root",
		Classes: []string{"foo", "bar"},
		Bounds:  uimath.NewRect(10, 20, 30, 40),
	}
	attrs := buildAttributes(n)
	if attrs[0] != "id" || attrs[1] != "root" {
		t.Errorf("want id=root, got %v=%v", attrs[0], attrs[1])
	}
	if attrs[2] != "class" || attrs[3] != "foo bar" {
		t.Errorf("want class=foo bar, got %v=%v", attrs[2], attrs[3])
	}
	// data-x should follow
	if attrs[4] != "data-x" || attrs[5] != "10" {
		t.Errorf("unexpected data-x: %v=%v", attrs[4], attrs[5])
	}
}

// ─── buildDOMNode ─────────────────────────────────────────────────────────────

func TestBuildDOMNode_LeafWithText(t *testing.T) {
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "span", []string{"lbl"}, "hello", uimath.NewRect(0, 0, 50, 20))
	snap.Nodes[1].ElemType = core.TypeText

	sess := &Session{srv: &Server{}}
	dn := sess.buildDOMNode(snap, 1, 1)
	if dn == nil {
		t.Fatal("expected non-nil domNode")
	}
	if dn.NodeName != "SPAN" {
		t.Errorf("want SPAN, got %s", dn.NodeName)
	}
	if dn.LocalName != "span" {
		t.Errorf("want span, got %s", dn.LocalName)
	}
	// Should have one synthetic #text child
	if len(dn.Children) != 1 {
		t.Fatalf("want 1 child (#text), got %d", len(dn.Children))
	}
	if dn.Children[0].NodeName != "#text" {
		t.Errorf("want #text, got %s", dn.Children[0].NodeName)
	}
	if dn.Children[0].NodeValue != "hello" {
		t.Errorf("want 'hello', got %q", dn.Children[0].NodeValue)
	}
	// Synthetic text node has negative ID
	if dn.Children[0].NodeID >= 0 {
		t.Errorf("expected negative ID for synthetic text node, got %d", dn.Children[0].NodeID)
	}
}

func TestBuildDOMNode_FallbackToElemType(t *testing.T) {
	snap := makeSnap(800, 600)
	snap.Root = 1
	// No HTMLTag set — falls back to ElemType
	addNode(snap, 1, core.InvalidElementID, "", nil, "", uimath.NewRect(0, 0, 100, 100))
	snap.Nodes[1].ElemType = core.TypeDiv

	sess := &Session{srv: &Server{}}
	dn := sess.buildDOMNode(snap, 1, 0)
	if dn.NodeName != "DIV" {
		t.Errorf("want DIV, got %s", dn.NodeName)
	}
}

func TestBuildDOMNode_Depth0(t *testing.T) {
	snap := makeSnap(800, 600)
	snap.Root = 1
	p := addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.NewRect(0, 0, 200, 100))
	addNode(snap, 2, 1, "span", nil, "child", uimath.NewRect(0, 0, 50, 20))
	p.ChildIDs = []core.ElementID{2}

	sess := &Session{srv: &Server{}}
	dn := sess.buildDOMNode(snap, 1, 0)
	// depth=0: children not expanded but ChildNodeCount reflects actual count
	if dn.Children != nil {
		t.Errorf("depth=0: expected no Children slice, got %v", dn.Children)
	}
	if dn.ChildNodeCount != 1 {
		t.Errorf("want ChildNodeCount=1, got %d", dn.ChildNodeCount)
	}
}

func TestBuildDOMNode_FullTree(t *testing.T) {
	snap := makeSnap(800, 600)
	snap.Root = 1
	p := addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.NewRect(0, 0, 200, 100))
	addNode(snap, 2, 1, "span", nil, "hello", uimath.NewRect(0, 0, 50, 20))
	p.ChildIDs = []core.ElementID{2}

	sess := &Session{srv: &Server{}}
	dn := sess.buildDOMNode(snap, 1, -1)
	if len(dn.Children) != 1 {
		t.Fatalf("want 1 child, got %d", len(dn.Children))
	}
	// Child expanded with its #text
	if len(dn.Children[0].Children) != 1 {
		t.Fatalf("want #text grandchild, got %d children", len(dn.Children[0].Children))
	}
}

func TestBuildDOMNode_NonExistentID(t *testing.T) {
	snap := makeSnap(800, 600)
	sess := &Session{srv: &Server{}}
	if dn := sess.buildDOMNode(snap, 99, 1); dn != nil {
		t.Errorf("expected nil for non-existent ID")
	}
}

func TestBuildDOMNode_WithIDAttr(t *testing.T) {
	snap := makeSnap(800, 600)
	snap.Root = 1
	n := addNode(snap, 1, core.InvalidElementID, "div", []string{"container"}, "", uimath.NewRect(0, 0, 100, 100))
	n.IDAttr = "main"

	sess := &Session{srv: &Server{}}
	dn := sess.buildDOMNode(snap, 1, 0)
	// id should appear in attributes
	found := false
	for i := 0; i+1 < len(dn.Attributes); i += 2 {
		if dn.Attributes[i] == "id" && dn.Attributes[i+1] == "main" {
			found = true
		}
	}
	if !found {
		t.Errorf("id attribute not found in %v", dn.Attributes)
	}
}

// ─── buildOuterHTML ───────────────────────────────────────────────────────────

func TestBuildOuterHTML_LeafText(t *testing.T) {
	snap := makeSnap(800, 600)
	snap.Root = 1
	n := addNode(snap, 1, core.InvalidElementID, "span", []string{"lbl"}, "hello & world", uimath.Rect{})
	n.IDAttr = "s1"

	html := buildOuterHTML(snap, 1, 0)
	if !strings.Contains(html, `id="s1"`) {
		t.Errorf("missing id: %s", html)
	}
	if !strings.Contains(html, `class="lbl"`) {
		t.Errorf("missing class: %s", html)
	}
	if !strings.Contains(html, "hello &amp; world") {
		t.Errorf("text not escaped: %s", html)
	}
	if !strings.HasSuffix(html, "</span>") {
		t.Errorf("missing close tag: %s", html)
	}
}

func TestBuildOuterHTML_VoidElement(t *testing.T) {
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "input", nil, "", uimath.Rect{})

	html := buildOuterHTML(snap, 1, 0)
	if !strings.HasSuffix(html, "/>") {
		t.Errorf("expected self-closing input, got: %s", html)
	}
}

func TestBuildOuterHTML_WithChildren(t *testing.T) {
	snap := makeSnap(800, 600)
	snap.Root = 1
	p := addNode(snap, 1, core.InvalidElementID, "div", []string{"wrap"}, "", uimath.Rect{})
	addNode(snap, 2, 1, "span", nil, "text", uimath.Rect{})
	p.ChildIDs = []core.ElementID{2}

	html := buildOuterHTML(snap, 1, 0)
	if !strings.Contains(html, "<div") {
		t.Errorf("missing div: %s", html)
	}
	if !strings.Contains(html, "<span>text</span>") {
		t.Errorf("missing span child: %s", html)
	}
	if !strings.Contains(html, "</div>") {
		t.Errorf("missing closing div: %s", html)
	}
}

func TestBuildOuterHTML_EmptyNoChildren(t *testing.T) {
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.Rect{})

	html := buildOuterHTML(snap, 1, 0)
	if html != "<div></div>" {
		t.Errorf("want <div></div>, got %q", html)
	}
}

func TestBuildOuterHTML_NonExistent(t *testing.T) {
	snap := makeSnap(800, 600)
	if got := buildOuterHTML(snap, 999, 0); got != "" {
		t.Errorf("want empty string, got %q", got)
	}
}

func TestBuildOuterHTML_ImgBrHr(t *testing.T) {
	for _, tag := range []string{"img", "br", "hr"} {
		snap := makeSnap(800, 600)
		snap.Root = 1
		addNode(snap, 1, core.InvalidElementID, tag, nil, "", uimath.Rect{})
		html := buildOuterHTML(snap, 1, 0)
		if !strings.HasSuffix(html, "/>") {
			t.Errorf("tag %s: expected self-closing, got %q", tag, html)
		}
	}
}

// ─── collectTextContent ───────────────────────────────────────────────────────

func TestCollectTextContent_Leaf(t *testing.T) {
	snap := makeSnap(800, 600)
	addNode(snap, 1, core.InvalidElementID, "span", nil, "hello", uimath.Rect{})
	if got := collectTextContent(snap, 1); got != "hello" {
		t.Errorf("want 'hello', got %q", got)
	}
}

func TestCollectTextContent_Recursive(t *testing.T) {
	snap := makeSnap(800, 600)
	p := addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.Rect{})
	addNode(snap, 2, 1, "span", nil, "Hello", uimath.Rect{})
	addNode(snap, 3, 1, "span", nil, "World", uimath.Rect{})
	p.ChildIDs = []core.ElementID{2, 3}

	if got := collectTextContent(snap, 1); got != "HelloWorld" {
		t.Errorf("want 'HelloWorld', got %q", got)
	}
}

func TestCollectTextContent_NonExistent(t *testing.T) {
	snap := makeSnap(800, 600)
	if got := collectTextContent(snap, 99); got != "" {
		t.Errorf("want empty, got %q", got)
	}
}

func TestCollectTextContent_Empty(t *testing.T) {
	snap := makeSnap(800, 600)
	addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.Rect{})
	if got := collectTextContent(snap, 1); got != "" {
		t.Errorf("want empty, got %q", got)
	}
}

// ─── hitTestSnap ─────────────────────────────────────────────────────────────

func TestHitTestSnap_Miss(t *testing.T) {
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.NewRect(0, 0, 100, 100))
	id := hitTestSnap(snap, 1, 200, 200) // outside
	if id != core.InvalidElementID {
		t.Errorf("expected miss, got %d", id)
	}
}

func TestHitTestSnap_HitParent(t *testing.T) {
	snap := makeSnap(800, 600)
	snap.Root = 1
	addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.NewRect(0, 0, 200, 200))

	id := hitTestSnap(snap, 1, 50, 50)
	if id != 1 {
		t.Errorf("want 1, got %d", id)
	}
}

func TestHitTestSnap_HitChild(t *testing.T) {
	snap := makeSnap(800, 600)
	snap.Root = 1
	p := addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.NewRect(0, 0, 200, 200))
	addNode(snap, 2, 1, "span", nil, "", uimath.NewRect(10, 10, 50, 50))
	p.ChildIDs = []core.ElementID{2}

	id := hitTestSnap(snap, 1, 20, 20) // inside child bounds
	if id != 2 {
		t.Errorf("want 2 (child), got %d", id)
	}
}

func TestHitTestSnap_NonExistent(t *testing.T) {
	snap := makeSnap(800, 600)
	if id := hitTestSnap(snap, 99, 0, 0); id != core.InvalidElementID {
		t.Errorf("expected invalid, got %d", id)
	}
}

// ─── matchesSelector ─────────────────────────────────────────────────────────

func TestMatchesSelector_Wildcard(t *testing.T) {
	n := &NodeSnapshot{ElemType: core.TypeDiv}
	if !matchesSelector(n, "*") {
		t.Error("wildcard should match everything")
	}
}

func TestMatchesSelector_Class(t *testing.T) {
	n := &NodeSnapshot{Classes: []string{"foo", "bar"}}
	if !matchesSelector(n, ".foo") {
		t.Error("should match .foo")
	}
	if matchesSelector(n, ".baz") {
		t.Error("should not match .baz")
	}
}

func TestMatchesSelector_HTMLTag(t *testing.T) {
	n := &NodeSnapshot{HTMLTag: "span", ElemType: core.TypeText}
	if !matchesSelector(n, "span") {
		t.Error("should match HTMLTag 'span'")
	}
	if !matchesSelector(n, "SPAN") {
		t.Error("should match case-insensitive SPAN")
	}
}

func TestMatchesSelector_ElemTypeFallback(t *testing.T) {
	n := &NodeSnapshot{ElemType: core.TypeDiv}
	if !matchesSelector(n, "div") {
		t.Error("should fall back to ElemType 'div'")
	}
}

func TestMatchesSelector_IDSelector(t *testing.T) {
	n := &NodeSnapshot{IDAttr: "header"}
	if !matchesSelector(n, "#header") {
		t.Error("#header should match node with IDAttr=header")
	}
	if matchesSelector(n, "#footer") {
		t.Error("#footer should not match node with IDAttr=header")
	}
	empty := &NodeSnapshot{}
	if matchesSelector(empty, "#header") {
		t.Error("#header should not match node with empty IDAttr")
	}
}

// ─── querySelector / querySelectorAll ────────────────────────────────────────

func TestQuerySelectorAll_Class(t *testing.T) {
	snap := makeSnap(800, 600)
	snap.Root = 1
	p := addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.Rect{})
	addNode(snap, 2, 1, "span", []string{"item"}, "a", uimath.Rect{})
	addNode(snap, 3, 1, "span", []string{"item"}, "b", uimath.Rect{})
	addNode(snap, 4, 1, "span", []string{"other"}, "c", uimath.Rect{})
	p.ChildIDs = []core.ElementID{2, 3, 4}

	ids := querySelectorAll(snap, 1, ".item")
	if len(ids) != 2 {
		t.Errorf("want 2 results, got %d: %v", len(ids), ids)
	}
}

func TestQuerySelectorAll_Tag(t *testing.T) {
	snap := makeSnap(800, 600)
	snap.Root = 1
	p := addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.Rect{})
	addNode(snap, 2, 1, "span", nil, "a", uimath.Rect{})
	addNode(snap, 3, 1, "p", nil, "b", uimath.Rect{})
	p.ChildIDs = []core.ElementID{2, 3}

	ids := querySelectorAll(snap, 1, "span")
	if len(ids) != 1 || ids[0] != 2 {
		t.Errorf("want [2], got %v", ids)
	}
}

func TestQuerySelectorAll_Wildcard(t *testing.T) {
	snap := makeSnap(800, 600)
	snap.Root = 1
	p := addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.Rect{})
	addNode(snap, 2, 1, "span", nil, "", uimath.Rect{})
	p.ChildIDs = []core.ElementID{2}

	ids := querySelectorAll(snap, 1, "*")
	if len(ids) != 2 { // root + child
		t.Errorf("wildcard should match all: got %d", len(ids))
	}
}

func TestQuerySelectorAll_NoMatch(t *testing.T) {
	snap := makeSnap(800, 600)
	addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.Rect{})
	if ids := querySelectorAll(snap, 1, ".nope"); len(ids) != 0 {
		t.Errorf("expected no match")
	}
}

func TestQuerySelector_ReturnsFirst(t *testing.T) {
	snap := makeSnap(800, 600)
	snap.Root = 1
	p := addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.Rect{})
	addNode(snap, 2, 1, "span", []string{"x"}, "", uimath.Rect{})
	addNode(snap, 3, 1, "span", []string{"x"}, "", uimath.Rect{})
	p.ChildIDs = []core.ElementID{2, 3}

	id := querySelector(snap, 1, ".x")
	if id != 2 {
		t.Errorf("want 2, got %d", id)
	}
}

func TestQuerySelector_NoMatch(t *testing.T) {
	snap := makeSnap(800, 600)
	addNode(snap, 1, core.InvalidElementID, "div", nil, "", uimath.Rect{})
	if id := querySelector(snap, 1, ".nope"); id != core.InvalidElementID {
		t.Errorf("expected InvalidElementID")
	}
}

// ─── quad ─────────────────────────────────────────────────────────────────────

func TestQuad(t *testing.T) {
	q := quad(10, 20, 100, 50)
	want := []float64{10, 20, 110, 20, 110, 70, 10, 70}
	if len(q) != 8 {
		t.Fatalf("want 8 floats, got %d", len(q))
	}
	for i, v := range want {
		if q[i] != v {
			t.Errorf("[%d] want %v, got %v", i, v, q[i])
		}
	}
}

// ─── emptyDocument ───────────────────────────────────────────────────────────

func TestEmptyDocument(t *testing.T) {
	sess := &Session{srv: &Server{}}
	doc := sess.emptyDocument()
	if doc.NodeType != nodeTypeDocument {
		t.Errorf("want nodeTypeDocument (%d), got %d", nodeTypeDocument, doc.NodeType)
	}
	if doc.NodeName != "#document" {
		t.Errorf("want #document, got %s", doc.NodeName)
	}
}

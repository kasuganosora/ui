package main

import (
	"regexp"
	"strings"
	"testing"

	ui "github.com/kasuganosora/ui"
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// TestBuildUI validates that the demo UI tree can be built and drawn without crashing.
func TestBuildUI(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()

	doc := ui.LoadHTMLDocument(tree, cfg, demoHTML, "")
	root := doc.Root
	if root == nil {
		t.Fatal("LoadHTMLDocument returned nil root")
	}

	// Run layout so widgets have bounds, then draw
	ui.AutoLayout(tree, root, 1280, 800)
	buf := render.NewCommandBuffer()
	root.Draw(buf)

	if buf.Len() == 0 {
		t.Error("expected draw commands from the demo UI")
	}
	t.Logf("demo UI generated %d render commands", buf.Len())
}

// TestBuildUIWidgetCount validates a reasonable number of widgets exist.
func TestBuildUIWidgetCount(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	ui.LoadHTMLDocument(tree, cfg, demoHTML, "")

	count := tree.ElementCount()
	if count < 20 {
		t.Errorf("expected at least 20 elements, got %d", count)
	}
	t.Logf("total elements: %d", count)
}

// TestCardChildBelowHeader verifies Card body children are positioned below the header.
func TestCardChildBelowHeader(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()

	doc := ui.LoadHTMLDocument(tree, cfg, demoHTML, "")
	root := doc.Root
	if root == nil {
		t.Fatal("LoadHTMLDocument returned nil root")
	}

	// Setup programmatic widgets (creates the Card in demo-card container)
	setupDemoWidgets(doc, tree, cfg)

	// Make the Card section visible (by default only Button section is shown)
	secCard := doc.QueryByID("sec-card")
	if secCard != nil {
		s := secCard.Style()
		s.Display = layout.DisplayBlock
		secCard.SetStyle(s)
	}

	// Hide other sections to avoid conflicts
	secButton := doc.QueryByID("sec-button")
	if secButton != nil {
		s := secButton.Style()
		s.Display = layout.DisplayNone
		secButton.SetStyle(s)
	}

	// Use Layout widget (first child) as root, matching real app behavior
	layoutRoot := root.Children()[0]
	ui.AutoLayout(tree, layoutRoot, 1280, 800)

	// Find the Card widget
	demoCardContainer := doc.QueryByID("demo-card")
	if demoCardContainer == nil {
		t.Fatal("demo-card container not found")
	}

	var card *widget.Card
	var bodyText *widget.Text
	for _, child := range demoCardContainer.Children() {
		if c, ok := child.(*widget.Card); ok && c.Title() == "卡片标题" {
			card = c
			for _, cc := range c.Children() {
				if txt, ok := cc.(*widget.Text); ok {
					bodyText = txt
					break
				}
			}
			break
		}
	}

	if card == nil {
		t.Fatal("Card widget not found")
	}
	if bodyText == nil {
		t.Fatal("Card body text not found")
	}

	cardBounds := card.Bounds()
	textBounds := bodyText.Bounds()

	t.Logf("card bounds: %+v", cardBounds)
	t.Logf("text bounds: %+v", textBounds)

	// Card header is 48px. Body text must start below header.
	headerBottom := cardBounds.Y + 48
	if textBounds.Y < headerBottom {
		t.Errorf("body text Y (%.1f) overlaps with card header bottom (%.1f)",
			textBounds.Y, headerBottom)
	}
}

// TestNavigationConsistency guards the three parallel lists that make the
// sidebar work: nav entries, <list-item value> and <div id="sec-*">.
func TestNavigationConsistency(t *testing.T) {
	itemRe := regexp.MustCompile(`<list-item value="([a-z0-9]+)"`)
	sectionRe := regexp.MustCompile(`<div id="sec-([a-z0-9]+)"`)
	sidebarMatches := itemRe.FindAllStringSubmatch(demoHTML, -1)
	sectionMatches := sectionRe.FindAllStringSubmatch(demoHTML, -1)

	sidebarSet := make(map[string]bool)
	for _, m := range sidebarMatches {
		sidebarSet[m[1]] = true
	}
	sectionSet := make(map[string]bool)
	for _, m := range sectionMatches {
		sectionSet[m[1]] = true
	}
	if len(sidebarMatches) != len(sidebarSet) {
		t.Errorf("duplicate sidebar values: %d items, %d unique", len(sidebarMatches), len(sidebarSet))
	}
	if len(sectionMatches) != len(sectionSet) {
		t.Errorf("duplicate section ids: %d sections, %d unique", len(sectionMatches), len(sectionSet))
	}

	for _, v := range sectionValues() {
		if !sidebarSet[v] {
			t.Errorf("nav %q has no sidebar item", v)
		}
		if !sectionSet[v] {
			t.Errorf("nav %q has no section", v)
		}
		sec, ok := sectionByValue(v)
		if !ok || sec != "sec-"+v {
			t.Errorf("sectionByValue(%q) = %q, %v", v, sec, ok)
		}
	}
	for v := range sidebarSet {
		if _, ok := sectionByValue(v); !ok {
			t.Errorf("sidebar value %q has no nav entry", v)
		}
	}
	for v := range sectionSet {
		if _, ok := sectionByValue(v); !ok {
			t.Errorf("section sec-%q has no nav entry", v)
		}
	}
	if _, ok := sectionByValue("does-not-exist"); ok {
		t.Error("sectionByValue accepted an unknown value")
	}
}

// TestSetupNavigationInitialSection verifies that -section shows exactly one
// panel, keeps the sidebar highlight in sync, and falls back for bad input.
func TestSetupNavigationInitialSection(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := ui.LoadHTMLDocument(tree, cfg, demoHTML, "")

	sec, ok := sectionByValue("emoji")
	if !ok {
		t.Fatal("emoji section missing")
	}
	setupNavigation(doc, sec)

	visible := 0
	for _, n := range nav {
		w := doc.QueryByID(n.section)
		if w == nil {
			t.Fatalf("section %q not found in document", n.section)
		}
		if w.Style().Display == layout.DisplayBlock {
			visible++
			if n.section != sec {
				t.Errorf("unexpected visible section %q", n.section)
			}
		}
	}
	if visible != 1 {
		t.Errorf("expected exactly one visible section, got %d", visible)
	}
	ls, ok := doc.QueryByID("sidebar-nav").(*widget.List)
	if !ok {
		t.Fatal("sidebar-nav is not a List")
	}
	if ls.SelectedValue() != "emoji" {
		t.Errorf("sidebar highlight = %q, want emoji", ls.SelectedValue())
	}

	// An unknown section must fall back to Button rather than blank the window.
	setupNavigation(doc, "sec-unknown")
	if w := doc.QueryByID("sec-button"); w == nil || w.Style().Display != layout.DisplayBlock {
		t.Error("unknown section did not fall back to sec-button")
	}
	if ls.SelectedValue() != "button" {
		t.Errorf("sidebar highlight after fallback = %q, want button", ls.SelectedValue())
	}
}

func TestParseBackend(t *testing.T) {
	cases := []struct {
		in   string
		want ui.BackendType
		ok   bool
	}{
		{"", ui.BackendAuto, true},
		{"auto", ui.BackendAuto, true},
		{"AUTO", ui.BackendAuto, true},
		{"metal", ui.BackendMetal, true},
		{"mtl", ui.BackendMetal, true},
		{"vulkan", ui.BackendVulkan, true},
		{"vk", ui.BackendVulkan, true},
		{"dx11", ui.BackendDX11, true},
		{"d3d11", ui.BackendDX11, true},
		{"dx9", ui.BackendDX9, true},
		{"gl", ui.BackendOpenGL, true},
		{"opengl", ui.BackendOpenGL, true},
		{"nope", ui.BackendAuto, false},
	}
	for _, c := range cases {
		got, err := parseBackend(c.in)
		if c.ok && err != nil {
			t.Errorf("parseBackend(%q) returned error: %v", c.in, err)
		}
		if !c.ok && err == nil {
			t.Errorf("parseBackend(%q) should reject an unknown backend", c.in)
		}
		if got != c.want {
			t.Errorf("parseBackend(%q) = %v, want %v", c.in, got, c.want)
		}
	}
	for _, v := range backendValues {
		if _, err := parseBackend(v); err != nil {
			t.Errorf("documented backend %q does not parse: %v", v, err)
		}
	}
}

// TestButtonBindings makes sure every clickable button in the Button section
// has a handler and that no handler points at a non-existent id.
func TestButtonBindings(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	doc := ui.LoadHTMLDocument(tree, cfg, demoHTML, "")

	seen := make(map[string]bool)
	for _, b := range buttonBindings {
		if seen[b.id] {
			t.Errorf("duplicate binding id %q", b.id)
		}
		seen[b.id] = true
		w := doc.QueryByID(b.id)
		if w == nil {
			t.Errorf("binding id %q is missing from demoHTML", b.id)
			continue
		}
		if _, ok := w.(*widget.Button); !ok {
			t.Errorf("binding id %q resolves to %T, not *widget.Button", b.id, w)
		}
	}

	start := strings.Index(demoHTML, `<div id="sec-button">`)
	end := strings.Index(demoHTML, `<!-- ===== Link ===== -->`)
	if start < 0 || end < 0 || end <= start {
		t.Fatal("could not locate the Button section in demoHTML")
	}
	buttonSection := demoHTML[start:end]
	btnRe := regexp.MustCompile(`<button([^>]*)>`)
	idRe := regexp.MustCompile(`id="([a-zA-Z0-9_-]+)"`)
	for _, m := range btnRe.FindAllStringSubmatch(buttonSection, -1) {
		attrs := m[1]
		if strings.Contains(attrs, "disabled") {
			continue // disabled buttons never fire a click
		}
		match := idRe.FindStringSubmatch(attrs)
		if match == nil {
			t.Errorf("clickable button without id: %s", attrs)
			continue
		}
		if !seen[match[1]] {
			t.Errorf("button id %q has no binding", match[1])
		}
	}
}

package widget

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Apple §1 regression. The important assertion is that a widget which only
// ever registered MouseClick still produces visible feedback on MouseDown —
// through the shared tree state and one DrawPressFeedback call per renderer,
// not through a per-widget patch.

func pressTintCommand(buf *render.CommandBuffer, cfg *Config) *render.Command {
	for i := range buf.Overlays() {
		cmd := &buf.Overlays()[i]
		if cmd.Type != render.CmdRect || cmd.Rect == nil || cmd.ZOrder != pressFeedbackZ {
			continue
		}
		c := cmd.Rect.FillColor
		want := cfg.TextColor
		if c.A == pressFeedbackAlpha && c.R == want.R && c.G == want.G && c.B == want.B {
			return cmd
		}
	}
	return nil
}

type pressFixture struct {
	name    string
	make    func(*core.Tree, *Config) Widget
	prepare func(Widget)
}

// A representative slice of the 41 widgets that register MouseClick without
// any MouseDown handler, plus the two that already had their own press state.
// Composite controls need real content: pagination pages, breadcrumb crumbs
// and collapse headers are child elements whose click handlers and layout
// bounds only exist after content is set and a frame has been drawn.
func pressFixtures() []pressFixture {
	return []pressFixture{
		{"Checkbox", func(t *core.Tree, c *Config) Widget { return NewCheckbox(t, "check", c) }, nil},
		{"Switch", func(t *core.Tree, c *Config) Widget { return NewSwitch(t, c) }, nil},
		{"Radio", func(t *core.Tree, c *Config) Widget { return NewRadio(t, "radio", c) }, nil},
		{"Tag", func(t *core.Tree, c *Config) Widget { return NewTag(t, "tag", c) }, nil},
		{"Link", func(t *core.Tree, c *Config) Widget { return NewLink(t, "link", "#", c) }, nil},
		{"Pagination", func(t *core.Tree, c *Config) Widget { return NewPagination(t, c) }, func(w Widget) {
			w.(*Pagination).SetTotal(100)
		}},
		{"Rate", func(t *core.Tree, c *Config) Widget { return NewRate(t, c) }, nil},
		{"List", func(t *core.Tree, c *Config) Widget { return NewList(t, c) }, nil},
		{"Menu", func(t *core.Tree, c *Config) Widget { return NewMenu(t, c) }, nil},
		{"Breadcrumb", func(t *core.Tree, c *Config) Widget { return NewBreadcrumb(t, c) }, func(w Widget) {
			w.(*Breadcrumb).SetOptions([]BreadcrumbItem{
				{Content: "Home", Href: "#"},
				{Content: "Components", Href: "#"},
				{Content: "Breadcrumb"},
			})
		}},
		{"EmojiPicker", func(t *core.Tree, c *Config) Widget { return NewEmojiPicker(t, c) }, nil},
		{"Collapse", func(t *core.Tree, c *Config) Widget { return NewCollapse(t, c) }, func(w Widget) {
			w.(*Collapse).SetPanels([]CollapsePanel{{Header: "Panel one", Value: "one"}})
		}},
		{"Button", func(t *core.Tree, c *Config) Widget { return NewButton(t, "press", c) }, nil},
		{"Slider", func(t *core.Tree, c *Config) Widget { return NewSlider(t, c) }, nil},
	}
}

// pressPoint finds a point that is actually hittable and resolves to an
// interactive element. For composite controls that point sits on a child
// (a page button, a crumb, a header), so the highlight must cover that child
// rather than the whole container.
func pressPoint(tree *core.Tree, root core.ElementID) (x, y float32, target core.ElementID) {
	tree.Walk(root, func(id core.ElementID, _ int) bool {
		elem := tree.Get(id)
		if elem == nil {
			return true
		}
		b := elem.Layout().Bounds
		if b.IsEmpty() {
			return true
		}
		if !(elem.HasHandler(event.MouseClick) || elem.HasHandler(event.MouseDown) || elem.Type() == core.TypeButton) {
			return true
		}
		px, py := b.X+b.Width/2, b.Y+b.Height/2
		resolved := tree.InteractiveTarget(tree.HitTest(px, py))
		if resolved != core.InvalidElementID {
			x, y, target = px, py, resolved
			return false
		}
		return true
	})
	return
}

func TestEveryClickableWidgetRespondsOnPointerDown(t *testing.T) {
	for _, tc := range pressFixtures() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tree := newTestTree()
			// HitTest prunes subtrees whose root bounds miss the point.
			tree.SetLayout(tree.Root(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 600, 600)})
			cfg := DefaultConfig()
			w := tc.make(tree, cfg)
			bounds := uimath.NewRect(0, 0, 200, 80)
			tree.AppendChild(tree.Root(), w.ElementID())
			tree.SetLayout(w.ElementID(), core.LayoutResult{Bounds: bounds})
			if tc.prepare != nil {
				tc.prepare(w)
			}
			// In a real frame the tree draws before input is routed, and
			// composite controls lay out their clickable children in Draw.
			w.Draw(render.NewCommandBuffer())

			px, py, want := pressPoint(tree, w.ElementID())
			if want == core.InvalidElementID {
				t.Fatal("fixture has no hittable interactive region")
			}

			host := core.HostInput{Tree: tree}
			host.Handle(&event.Event{Type: event.MouseDown, GlobalX: px, GlobalY: py})
			if tree.Pressed() == core.InvalidElementID {
				t.Fatal("pointer-down did not reach an interactive target")
			}
			if tree.Pressed() != want {
				t.Errorf("press published on %d, want %d", tree.Pressed(), want)
			}

			buf := render.NewCommandBuffer()
			w.Draw(buf)
			DrawPressFeedback(buf, tree, cfg)
			tint := pressTintCommand(buf, cfg)
			if tint == nil {
				t.Fatal("no press highlight emitted on pointer-down")
			}
			pressedElem := tree.Get(tree.Pressed())
			if pressedElem == nil {
				t.Fatal("press target disappeared during the frame")
			}
			if tint.Rect.Bounds != pressedElem.Layout().Bounds {
				t.Errorf("highlight bounds %+v do not cover the pressed control %+v", tint.Rect.Bounds, pressedElem.Layout().Bounds)
			}
			if !bounds.Contains(uimath.Vec2{X: tint.Rect.Bounds.X, Y: tint.Rect.Bounds.Y}) {
				t.Errorf("highlight %+v escapes the widget bounds %+v", tint.Rect.Bounds, bounds)
			}

			host.Handle(&event.Event{Type: event.MouseUp, GlobalX: px, GlobalY: py})
			released := render.NewCommandBuffer()
			w.Draw(released)
			DrawPressFeedback(released, tree, cfg)
			if pressTintCommand(released, cfg) != nil {
				t.Error("press highlight survived pointer release")
			}
		})
	}
}

func TestPressFeedbackSkipsInactiveTargets(t *testing.T) {
	tree := newTestTree()
	tree.SetLayout(tree.Root(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 600, 600)})
	cfg := DefaultConfig()
	b := NewButton(tree, "press", cfg)
	tree.AppendChild(tree.Root(), b.ElementID())
	tree.SetLayout(b.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 100, 40)})

	host := core.HostInput{Tree: tree}
	host.Handle(&event.Event{Type: event.MouseDown, GlobalX: 10, GlobalY: 10})

	b.SetDisabled(true)
	buf := render.NewCommandBuffer()
	DrawPressFeedback(buf, tree, cfg)
	if pressTintCommand(buf, cfg) != nil {
		t.Error("disabled control must not show press feedback")
	}
	b.SetDisabled(false)

	tree.SetVisible(b.ElementID(), false)
	buf = render.NewCommandBuffer()
	DrawPressFeedback(buf, tree, cfg)
	if pressTintCommand(buf, cfg) != nil {
		t.Error("hidden control must not show press feedback")
	}
	tree.SetVisible(b.ElementID(), true)

	tree.SetLayout(b.ElementID(), core.LayoutResult{})
	buf = render.NewCommandBuffer()
	DrawPressFeedback(buf, tree, cfg)
	if pressTintCommand(buf, cfg) != nil {
		t.Error("unlaid-out control must not show press feedback")
	}
}

func TestPressFeedbackUsesThemeTextColor(t *testing.T) {
	tree := newTestTree()
	tree.SetLayout(tree.Root(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 600, 600)})
	cfg := DefaultConfig()
	cfg.TextColor = uimath.ColorHex("#123456")
	b := NewButton(tree, "press", cfg)
	tree.AppendChild(tree.Root(), b.ElementID())
	tree.SetLayout(b.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 100, 40)})
	host := core.HostInput{Tree: tree}
	host.Handle(&event.Event{Type: event.MouseDown, GlobalX: 10, GlobalY: 10})

	buf := render.NewCommandBuffer()
	DrawPressFeedback(buf, tree, cfg)
	cmd := pressTintCommand(buf, cfg)
	if cmd == nil {
		t.Fatal("no press highlight")
	}
	if got := cmd.Rect.FillColor; got.R != cfg.TextColor.R || got.G != cfg.TextColor.G || got.B != cfg.TextColor.B {
		t.Errorf("highlight does not follow the theme text color: %+v", got)
	}
}

// §1 also requires feedback to stay continuous during a drag, not only at
// the end of the gesture. The slider already tracks MouseMove; this pins that
// behaviour so a future refactor cannot move it back to MouseUp.
func TestSliderTracksPointerThroughoutDrag(t *testing.T) {
	tree := newTestTree()
	tree.SetLayout(tree.Root(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 600, 600)})
	cfg := DefaultConfig()
	s := NewSlider(tree, cfg)
	tree.AppendChild(tree.Root(), s.ElementID())
	tree.SetLayout(s.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 200, 40)})

	host := core.HostInput{Tree: tree}
	host.Handle(&event.Event{Type: event.MouseDown, GlobalX: 20, GlobalY: 20})
	first := s.Value()

	var values []float32
	for _, x := range []float32{60, 100, 140, 180} {
		host.Handle(&event.Event{Type: event.MouseMove, GlobalX: x, GlobalY: 20})
		values = append(values, s.Value())
	}
	for i := 1; i < len(values); i++ {
		if values[i] <= values[i-1] {
			t.Errorf("value did not follow the pointer: %v (start %f)", values, first)
			break
		}
	}
	if values[0] <= first {
		t.Errorf("first move produced no update: start=%f values=%v", first, values)
	}
	// 1:1 with the pointer on a 0..100 slider across a 200px track.
	if math.Abs(float64(values[len(values)-1]-90)) > 15 {
		t.Errorf("drag is not tracking the pointer position: final=%f (want ~90)", values[len(values)-1])
	}
	host.Handle(&event.Event{Type: event.MouseUp, GlobalX: 180, GlobalY: 20})
}

func TestPressFeedbackSourceGuard(t *testing.T) {
	widgetFiles, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	clickFiles := 0
	for _, f := range widgetFiles {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(src), "event.MouseClick") {
			clickFiles++
		}
	}
	if clickFiles < 40 {
		t.Errorf("source guard only saw %d click-handling widget files", clickFiles)
	}

	// Every renderer that draws a widget tree must emit the shared highlight.
	hosts := []string{"../app.go", "../headless.go", "../embedded.go", "../godot/ui.go", "../cmd/demo/visual_test.go"}
	for _, f := range hosts {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(src), "DrawPressFeedback(") {
			t.Errorf("%s draws a widget tree without §1 press feedback", f)
		}
	}
	hostInput, err := os.ReadFile("../core/host_input.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(hostInput), "SetPressed(") {
		t.Error("HostInput no longer publishes press state on pointer-down")
	}
	elem, err := os.ReadFile("../core/element.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(elem), "InteractiveTarget(") {
		t.Error("tree no longer resolves the nearest interactive ancestor")
	}
}

package widget

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

func newTestTree() *core.Tree {
	return core.NewTree()
}

// --- Config tests ---

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.FontSize != 14 {
		t.Errorf("expected FontSize 14, got %f", cfg.FontSize)
	}
	if cfg.BorderRadius != 6 {
		t.Errorf("expected BorderRadius 6, got %f", cfg.BorderRadius)
	}
	if cfg.ButtonHeight != 32 {
		t.Errorf("expected ButtonHeight 32, got %f", cfg.ButtonHeight)
	}
	if cfg.PrimaryColor.IsTransparent() {
		t.Error("primary color should not be transparent")
	}
}

// --- Base tests ---

func TestBaseWidget(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	base := NewBase(tree, core.TypeDiv, cfg)

	if base.ElementID() == core.InvalidElementID {
		t.Error("expected valid element ID")
	}
	if base.Element() == nil {
		t.Error("expected non-nil element")
	}
	if base.Config() != cfg {
		t.Error("config mismatch")
	}
	if base.Tree() != tree {
		t.Error("tree mismatch")
	}
}

func TestBaseNilConfig(t *testing.T) {
	tree := newTestTree()
	base := NewBase(tree, core.TypeDiv, nil)
	if base.Config() == nil {
		t.Error("nil config should default")
	}
}

func TestBaseStyle(t *testing.T) {
	tree := newTestTree()
	base := NewBase(tree, core.TypeDiv, nil)
	s := base.Style()
	if s.Display != layout.DisplayBlock {
		t.Error("default should be block")
	}
	s.Display = layout.DisplayFlex
	base.SetStyle(s)
	if base.Style().Display != layout.DisplayFlex {
		t.Error("style should be updated")
	}
}

func TestBaseChildren(t *testing.T) {
	tree := newTestTree()
	parent := NewBase(tree, core.TypeDiv, nil)
	child := NewBase(tree, core.TypeDiv, nil)

	if len(parent.Children()) != 0 {
		t.Error("should start with no children")
	}

	// Create a wrapper to satisfy Widget interface
	cw := &testWidget{Base: child}
	parent.AppendChild(cw)
	if len(parent.Children()) != 1 {
		t.Error("should have 1 child")
	}

	parent.RemoveChild(cw)
	if len(parent.Children()) != 0 {
		t.Error("should have 0 children after remove")
	}
}

func TestBaseRemoveNonexistentChild(t *testing.T) {
	tree := newTestTree()
	parent := NewBase(tree, core.TypeDiv, nil)
	fake := NewBase(tree, core.TypeDiv, nil)
	fw := &testWidget{Base: fake}
	parent.RemoveChild(fw) // should not panic
}

func TestBaseDestroy(t *testing.T) {
	tree := newTestTree()
	base := NewBase(tree, core.TypeDiv, nil)
	id := base.ElementID()
	base.Destroy()
	if tree.Get(id) != nil {
		t.Error("element should be removed after destroy")
	}
}

func TestBaseDestroyWithChildren(t *testing.T) {
	tree := newTestTree()
	parent := NewBase(tree, core.TypeDiv, nil)
	child := NewBase(tree, core.TypeDiv, nil)
	cw := &testWidget{Base: child}
	parent.AppendChild(cw)
	childID := child.ElementID()
	parent.Destroy()
	if tree.Get(childID) != nil {
		t.Error("child should be destroyed with parent")
	}
}

func TestBaseBounds(t *testing.T) {
	tree := newTestTree()
	base := NewBase(tree, core.TypeDiv, nil)
	// No layout set, should return empty rect
	b := base.Bounds()
	if b.Width != 0 || b.Height != 0 {
		t.Error("should be empty without layout")
	}
}

func TestBaseOn(t *testing.T) {
	tree := newTestTree()
	base := NewBase(tree, core.TypeDiv, nil)
	called := false
	base.On(event.MouseClick, func(e *event.Event) {
		called = true
	})
	handlers := tree.Handlers(base.ElementID(), event.MouseClick)
	if len(handlers) != 1 {
		t.Fatal("expected 1 handler")
	}
	handlers[0](&event.Event{})
	if !called {
		t.Error("handler should have been called")
	}
}

func TestBaseDrawChildren(t *testing.T) {
	tree := newTestTree()
	parent := NewBase(tree, core.TypeDiv, nil)
	child := &drawCounter{}
	parent.children = append(parent.children, child)
	buf := render.NewCommandBuffer()
	parent.DrawChildren(buf)
	if child.count != 1 {
		t.Error("child should have been drawn once")
	}
}

// --- Text tests ---

func TestText(t *testing.T) {
	tree := newTestTree()
	txt := NewText(tree, "hello", nil)
	if txt.Text() != "hello" {
		t.Error("text mismatch")
	}
	if txt.FontSize() != 14 {
		t.Error("default font size should be 14")
	}
	if txt.Color().IsTransparent() {
		t.Error("text color should not be transparent")
	}

	txt.SetText("world")
	if txt.Text() != "world" {
		t.Error("text should be updated")
	}

	txt.SetColor(uimath.ColorRed)
	if txt.Color() != uimath.ColorRed {
		t.Error("color should be updated")
	}

	txt.SetFontSize(20)
	if txt.FontSize() != 20 {
		t.Error("font size should be updated")
	}
}

func TestTextNilConfig(t *testing.T) {
	tree := newTestTree()
	txt := NewText(tree, "test", nil)
	if txt.config == nil {
		t.Error("should get default config")
	}
}

func TestTextStyle(t *testing.T) {
	tree := newTestTree()
	txt := NewText(tree, "test", nil)
	s := txt.Style()
	if s.Display != layout.DisplayBlock {
		t.Error("text should be block display")
	}
}

func TestTextDraw(t *testing.T) {
	tree := newTestTree()
	txt := NewText(tree, "hello", nil)
	buf := render.NewCommandBuffer()

	// Empty bounds — no draw
	txt.Draw(buf)
	if buf.Len() != 0 {
		t.Error("should not draw with empty bounds")
	}

	// Set layout bounds
	tree.SetLayout(txt.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(10, 20, 100, 30),
	})
	txt.Draw(buf)
	if buf.Len() != 1 {
		t.Errorf("expected 1 command, got %d", buf.Len())
	}
}

func TestTextDrawEmpty(t *testing.T) {
	tree := newTestTree()
	txt := NewText(tree, "", nil)
	tree.SetLayout(txt.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 100, 30),
	})
	buf := render.NewCommandBuffer()
	txt.Draw(buf)
	if buf.Len() != 0 {
		t.Error("empty text should not draw")
	}
}

// --- Div tests ---

func TestDiv(t *testing.T) {
	tree := newTestTree()
	d := NewDiv(tree, nil)

	if d.Style().Display != layout.DisplayFlex {
		t.Error("div should be flex")
	}
	if d.IsScrollable() {
		t.Error("not scrollable by default")
	}

	d.SetBgColor(uimath.ColorWhite)
	if d.BgColor() != uimath.ColorWhite {
		t.Error("bg color mismatch")
	}

	d.SetBorderColor(uimath.ColorBlack)
	if d.BorderColor() != uimath.ColorBlack {
		t.Error("border color mismatch")
	}

	d.SetBorderWidth(2)
	if d.BorderWidth() != 2 {
		t.Error("border width mismatch")
	}

	d.SetBorderRadius(8)
	if d.BorderRadius() != 8 {
		t.Error("border radius mismatch")
	}

	d.SetScrollable(true)
	if !d.IsScrollable() {
		t.Error("should be scrollable")
	}

	d.ScrollTo(10, 20)
	if d.ScrollX() != 10 || d.ScrollY() != 20 {
		t.Error("scroll mismatch")
	}
}

func TestDivDraw(t *testing.T) {
	tree := newTestTree()
	d := NewDiv(tree, nil)
	d.SetBgColor(uimath.ColorWhite)
	tree.SetLayout(d.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 200, 100),
	})

	buf := render.NewCommandBuffer()
	d.Draw(buf)
	if buf.Len() != 1 {
		t.Errorf("expected 1 rect command, got %d", buf.Len())
	}
}

func TestDivDrawScrollable(t *testing.T) {
	tree := newTestTree()
	d := NewDiv(tree, nil)
	d.SetBgColor(uimath.ColorWhite)
	d.SetScrollable(true)
	tree.SetLayout(d.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 200, 100),
	})

	buf := render.NewCommandBuffer()
	d.Draw(buf)
	// rect + clip
	if buf.Len() != 2 {
		t.Errorf("expected 2 commands (rect+clip), got %d", buf.Len())
	}
}

func TestDivDrawEmpty(t *testing.T) {
	tree := newTestTree()
	d := NewDiv(tree, nil)
	buf := render.NewCommandBuffer()
	d.Draw(buf)
	if buf.Len() != 0 {
		t.Error("empty bounds should not draw")
	}
}

func TestDivDrawTransparent(t *testing.T) {
	tree := newTestTree()
	d := NewDiv(tree, nil)
	tree.SetLayout(d.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 200, 100),
	})
	buf := render.NewCommandBuffer()
	d.Draw(buf)
	// transparent bg + no border = no rect
	if buf.Len() != 0 {
		t.Error("transparent bg with no border should not draw rect")
	}
}

// --- Button tests ---

func TestButton(t *testing.T) {
	tree := newTestTree()
	btn := NewButton(tree, "Click", nil)

	if btn.Label() != "Click" {
		t.Error("label mismatch")
	}
	if btn.Variant() != ButtonPrimary {
		t.Error("default should be primary")
	}
	if btn.IsDisabled() {
		t.Error("should not be disabled by default")
	}
	if btn.IsPressed() {
		t.Error("should not be pressed")
	}

	btn.SetLabel("OK")
	if btn.Label() != "OK" {
		t.Error("label should be updated")
	}

	btn.SetVariant(ButtonSecondary)
	if btn.Variant() != ButtonSecondary {
		t.Error("variant should be updated")
	}

	btn.SetDisabled(true)
	if !btn.IsDisabled() {
		t.Error("should be disabled")
	}
}

func TestButtonClick(t *testing.T) {
	tree := newTestTree()
	btn := NewButton(tree, "Click", nil)
	clicked := false
	btn.OnClick(func() { clicked = true })

	// Simulate click via handlers
	handlers := tree.Handlers(btn.ElementID(), event.MouseClick)
	for _, h := range handlers {
		h(&event.Event{Type: event.MouseClick})
	}
	if !clicked {
		t.Error("click handler should fire")
	}
}

func TestButtonClickDisabled(t *testing.T) {
	tree := newTestTree()
	btn := NewButton(tree, "Click", nil)
	btn.SetDisabled(true)
	clicked := false
	btn.OnClick(func() { clicked = true })

	handlers := tree.Handlers(btn.ElementID(), event.MouseClick)
	for _, h := range handlers {
		h(&event.Event{Type: event.MouseClick})
	}
	if clicked {
		t.Error("disabled button should not fire click")
	}
}

func TestButtonPress(t *testing.T) {
	tree := newTestTree()
	btn := NewButton(tree, "Click", nil)

	// MouseDown
	handlers := tree.Handlers(btn.ElementID(), event.MouseDown)
	for _, h := range handlers {
		h(&event.Event{Type: event.MouseDown})
	}
	if !btn.IsPressed() {
		t.Error("should be pressed after mouse down")
	}

	// MouseUp
	handlers = tree.Handlers(btn.ElementID(), event.MouseUp)
	for _, h := range handlers {
		h(&event.Event{Type: event.MouseUp})
	}
	if btn.IsPressed() {
		t.Error("should not be pressed after mouse up")
	}
}

func TestButtonPressDisabled(t *testing.T) {
	tree := newTestTree()
	btn := NewButton(tree, "Click", nil)
	btn.SetDisabled(true)

	handlers := tree.Handlers(btn.ElementID(), event.MouseDown)
	for _, h := range handlers {
		h(&event.Event{Type: event.MouseDown})
	}
	if btn.IsPressed() {
		t.Error("disabled button should not become pressed")
	}
}

func TestButtonDraw(t *testing.T) {
	tree := newTestTree()
	btn := NewButton(tree, "OK", nil)
	tree.SetLayout(btn.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 100, 32),
	})

	buf := render.NewCommandBuffer()
	btn.Draw(buf)
	// rect + text
	if buf.Len() != 2 {
		t.Errorf("expected 2 commands, got %d", buf.Len())
	}
}

func TestButtonDrawEmpty(t *testing.T) {
	tree := newTestTree()
	btn := NewButton(tree, "OK", nil)
	buf := render.NewCommandBuffer()
	btn.Draw(buf)
	if buf.Len() != 0 {
		t.Error("empty bounds should not draw")
	}
}

func TestButtonDrawEmptyLabel(t *testing.T) {
	tree := newTestTree()
	btn := NewButton(tree, "", nil)
	tree.SetLayout(btn.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 100, 32),
	})
	buf := render.NewCommandBuffer()
	btn.Draw(buf)
	// Just rect, no text
	if buf.Len() != 1 {
		t.Errorf("expected 1 command for empty label, got %d", buf.Len())
	}
}

func TestButtonVariantColors(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()

	for _, v := range []ButtonVariant{ButtonPrimary, ButtonSecondary, ButtonText, ButtonLink} {
		btn := NewButton(tree, "X", cfg)
		btn.SetVariant(v)
		bg := btn.bgColor()
		tc := btn.textColor()
		_ = bg
		_ = tc
	}

	// Disabled
	btn := NewButton(tree, "X", cfg)
	btn.SetDisabled(true)
	bg := btn.bgColor()
	if bg != cfg.DisabledColor {
		t.Error("disabled should use disabled color")
	}
	tc := btn.textColor()
	if tc != cfg.BgColor {
		t.Error("disabled text should use bg color")
	}
}

func TestButtonHoverAndPressed(t *testing.T) {
	tree := newTestTree()
	btn := NewButton(tree, "X", nil)

	// Hovered
	tree.SetHovered(btn.ElementID(), true)
	bg := btn.bgColor()
	if bg != btn.config.HoverColor {
		t.Error("hovered primary should use hover color")
	}

	// Pressed
	btn.pressed = true
	bg = btn.bgColor()
	if bg != btn.config.ActiveColor {
		t.Error("pressed primary should use active color")
	}
}

func TestButtonDrawVariants(t *testing.T) {
	tree := newTestTree()

	for _, v := range []ButtonVariant{ButtonPrimary, ButtonSecondary, ButtonText, ButtonLink} {
		btn := NewButton(tree, "X", nil)
		btn.SetVariant(v)
		tree.SetLayout(btn.ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(0, 0, 80, 32),
		})
		buf := render.NewCommandBuffer()
		btn.Draw(buf)
		if buf.Len() < 1 {
			t.Errorf("variant %d should produce at least 1 command", v)
		}
	}
}

// --- Input tests ---

func TestInput(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)

	if inp.Value() != "" {
		t.Error("should start empty")
	}
	if inp.Placeholder() != "" {
		t.Error("should have no placeholder")
	}
	if inp.IsDisabled() {
		t.Error("should not be disabled")
	}
	if inp.CursorPos() != 0 {
		t.Error("cursor should be at 0")
	}

	inp.SetValue("hello")
	if inp.Value() != "hello" {
		t.Error("value mismatch")
	}

	inp.SetPlaceholder("Type here...")
	if inp.Placeholder() != "Type here..." {
		t.Error("placeholder mismatch")
	}

	inp.SetDisabled(true)
	if !inp.IsDisabled() {
		t.Error("should be disabled")
	}
}

func TestInputSetValueTruncatesCursor(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.SetValue("abcde")
	inp.cursorPos = 5
	inp.SetValue("ab")
	if inp.CursorPos() != 2 {
		t.Errorf("cursor should be truncated to 2, got %d", inp.CursorPos())
	}
}

func TestInputSelection(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	s, e := inp.Selection()
	if s != 0 || e != 0 {
		t.Error("selection should be 0,0")
	}
}

func TestInputTyping(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	var lastValue string
	inp.OnChange(func(v string) { lastValue = v })

	// Simulate KeyPress
	handlers := tree.Handlers(inp.ElementID(), event.KeyPress)
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyPress, Char: 'a'})
	}
	if inp.Value() != "a" {
		t.Errorf("expected 'a', got '%s'", inp.Value())
	}
	if lastValue != "a" {
		t.Error("onChange should fire")
	}
	if inp.CursorPos() != 1 {
		t.Errorf("cursor should be 1, got %d", inp.CursorPos())
	}

	// Type another char
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyPress, Char: 'b'})
	}
	if inp.Value() != "ab" {
		t.Errorf("expected 'ab', got '%s'", inp.Value())
	}
}

func TestInputTypingDisabled(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.SetDisabled(true)

	handlers := tree.Handlers(inp.ElementID(), event.KeyPress)
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyPress, Char: 'a'})
	}
	if inp.Value() != "" {
		t.Error("disabled input should not accept input")
	}
}

func TestInputTypingZeroChar(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)

	handlers := tree.Handlers(inp.ElementID(), event.KeyPress)
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyPress, Char: 0})
	}
	if inp.Value() != "" {
		t.Error("zero char should not insert")
	}
}

func TestInputBackspace(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.SetValue("abc")
	inp.cursorPos = 3

	handlers := tree.Handlers(inp.ElementID(), event.KeyDown)
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyBackspace})
	}
	if inp.Value() != "ab" {
		t.Errorf("expected 'ab', got '%s'", inp.Value())
	}
	if inp.CursorPos() != 2 {
		t.Error("cursor should be 2")
	}
}

func TestInputBackspaceAtStart(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.SetValue("abc")
	inp.cursorPos = 0

	handlers := tree.Handlers(inp.ElementID(), event.KeyDown)
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyBackspace})
	}
	if inp.Value() != "abc" {
		t.Error("should not delete at start")
	}
}

func TestInputDelete(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.SetValue("abc")
	inp.cursorPos = 0

	handlers := tree.Handlers(inp.ElementID(), event.KeyDown)
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyDelete})
	}
	if inp.Value() != "bc" {
		t.Errorf("expected 'bc', got '%s'", inp.Value())
	}
}

func TestInputDeleteAtEnd(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.SetValue("abc")
	inp.cursorPos = 3

	handlers := tree.Handlers(inp.ElementID(), event.KeyDown)
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyDelete})
	}
	if inp.Value() != "abc" {
		t.Error("should not delete at end")
	}
}

func TestInputArrowKeys(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.SetValue("abc")
	inp.cursorPos = 1

	handlers := tree.Handlers(inp.ElementID(), event.KeyDown)

	// Left
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyArrowLeft})
	}
	if inp.CursorPos() != 0 {
		t.Error("left arrow should move cursor left")
	}

	// Left at 0 stays 0
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyArrowLeft})
	}
	if inp.CursorPos() != 0 {
		t.Error("cursor should stay at 0")
	}

	// Right
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyArrowRight})
	}
	if inp.CursorPos() != 1 {
		t.Error("right arrow should move cursor right")
	}

	// Right past end
	inp.cursorPos = 3
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyArrowRight})
	}
	if inp.CursorPos() != 3 {
		t.Error("cursor should not go past end")
	}
}

func TestInputHomeEnd(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.SetValue("abc")
	inp.cursorPos = 1

	handlers := tree.Handlers(inp.ElementID(), event.KeyDown)

	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyHome})
	}
	if inp.CursorPos() != 0 {
		t.Error("Home should move to start")
	}

	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyEnd})
	}
	if inp.CursorPos() != 3 {
		t.Error("End should move to end")
	}
}

func TestInputKeyDownDisabled(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.SetValue("abc")
	inp.cursorPos = 3
	inp.SetDisabled(true)

	handlers := tree.Handlers(inp.ElementID(), event.KeyDown)
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyBackspace})
	}
	if inp.Value() != "abc" {
		t.Error("disabled input should not handle key down")
	}
}

func TestInputClickFocus(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)

	handlers := tree.Handlers(inp.ElementID(), event.MouseClick)
	for _, h := range handlers {
		h(&event.Event{Type: event.MouseClick})
	}
	if !inp.Element().IsFocused() {
		t.Error("click should focus input")
	}
}

func TestInputClickDisabledNoFocus(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.SetDisabled(true)

	handlers := tree.Handlers(inp.ElementID(), event.MouseClick)
	for _, h := range handlers {
		h(&event.Event{Type: event.MouseClick})
	}
	if inp.Element().IsFocused() {
		t.Error("disabled input should not gain focus on click")
	}
}

func TestInputDraw(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.SetValue("hello")
	tree.SetLayout(inp.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 200, 32),
	})

	buf := render.NewCommandBuffer()
	inp.Draw(buf)
	// rect + text
	if buf.Len() != 2 {
		t.Errorf("expected 2 commands, got %d", buf.Len())
	}
}

func TestInputDrawPlaceholder(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.SetPlaceholder("Enter...")
	tree.SetLayout(inp.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 200, 32),
	})

	buf := render.NewCommandBuffer()
	inp.Draw(buf)
	// rect + placeholder text
	if buf.Len() != 2 {
		t.Errorf("expected 2 commands, got %d", buf.Len())
	}
}

func TestInputDrawEmpty(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	buf := render.NewCommandBuffer()
	inp.Draw(buf)
	if buf.Len() != 0 {
		t.Error("empty bounds should not draw")
	}
}

func TestInputDrawDisabled(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.SetValue("hi")
	inp.SetDisabled(true)
	tree.SetLayout(inp.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 200, 32),
	})
	buf := render.NewCommandBuffer()
	inp.Draw(buf)
	if buf.Len() < 1 {
		t.Error("disabled input should still draw")
	}
}

func TestInputDrawFocused(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.SetValue("hi")
	tree.SetFocused(inp.ElementID(), true)
	tree.SetLayout(inp.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 200, 32),
	})
	buf := render.NewCommandBuffer()
	inp.Draw(buf)
	if buf.Len() < 1 {
		t.Error("focused input should draw")
	}
}

func TestInputDrawNoPlaceholderNoValue(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	tree.SetLayout(inp.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 200, 32),
	})
	buf := render.NewCommandBuffer()
	inp.Draw(buf)
	// Just rect, no text
	if buf.Len() != 1 {
		t.Errorf("expected 1 command for empty input, got %d", buf.Len())
	}
}

// --- Icon tests ---

func TestIcon(t *testing.T) {
	tree := newTestTree()
	ic := NewIcon(tree, "check", nil)
	if ic.Name() != "check" {
		t.Error("name mismatch")
	}
	if ic.Size() != 16 {
		t.Error("default size should be 16")
	}
	if ic.Texture() != render.InvalidTexture {
		t.Error("should have no texture initially")
	}

	ic.SetName("close")
	if ic.Name() != "close" {
		t.Error("name should update")
	}

	ic.SetSize(24)
	if ic.Size() != 24 {
		t.Error("size should update")
	}

	ic.SetColor(uimath.ColorRed)
	if ic.Color() != uimath.ColorRed {
		t.Error("color should update")
	}

	ic.SetTexture(42, uimath.NewRect(0, 0, 1, 1))
	if ic.Texture() != 42 {
		t.Error("texture should update")
	}
}

func TestIconDraw(t *testing.T) {
	tree := newTestTree()
	ic := NewIcon(tree, "check", nil)
	ic.SetTexture(1, uimath.NewRect(0, 0, 16, 16))
	tree.SetLayout(ic.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 16, 16),
	})

	buf := render.NewCommandBuffer()
	ic.Draw(buf)
	if buf.Len() != 1 {
		t.Errorf("expected 1 image command, got %d", buf.Len())
	}
}

func TestIconDrawNoTexture(t *testing.T) {
	tree := newTestTree()
	ic := NewIcon(tree, "check", nil)
	tree.SetLayout(ic.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 16, 16),
	})
	buf := render.NewCommandBuffer()
	ic.Draw(buf)
	if buf.Len() != 0 {
		t.Error("no texture means no draw")
	}
}

func TestIconDrawEmpty(t *testing.T) {
	tree := newTestTree()
	ic := NewIcon(tree, "check", nil)
	buf := render.NewCommandBuffer()
	ic.Draw(buf)
	if buf.Len() != 0 {
		t.Error("empty bounds should not draw")
	}
}

// --- Space tests ---

func TestSpace(t *testing.T) {
	tree := newTestTree()
	s := NewSpace(tree, nil)
	if s.Direction() != SpaceHorizontal {
		t.Error("default should be horizontal")
	}
	if s.Gap() != 8 {
		t.Errorf("default gap should be 8, got %f", s.Gap())
	}

	s.SetDirection(SpaceVertical)
	if s.Direction() != SpaceVertical {
		t.Error("direction should update")
	}
	if s.style.FlexDirection != layout.FlexDirectionColumn {
		t.Error("style should be column")
	}

	s.SetDirection(SpaceHorizontal)
	if s.style.FlexDirection != layout.FlexDirectionRow {
		t.Error("style should be row")
	}

	s.SetGap(16)
	if s.Gap() != 16 {
		t.Error("gap should update")
	}
	if s.style.Gap != 16 {
		t.Error("style gap should update")
	}
}

func TestSpaceDraw(t *testing.T) {
	tree := newTestTree()
	s := NewSpace(tree, nil)
	buf := render.NewCommandBuffer()
	s.Draw(buf) // should not panic
}

// --- Grid tests ---

func TestRow(t *testing.T) {
	tree := newTestTree()
	r := NewRow(tree, nil)
	if r.Style().Display != layout.DisplayFlex {
		t.Error("row should be flex")
	}
	if r.Gutter() != 0 {
		t.Error("default gutter should be 0")
	}
	r.SetGutter(16)
	if r.Gutter() != 16 {
		t.Error("gutter should update")
	}
}

func TestCol(t *testing.T) {
	tree := newTestTree()
	c := NewCol(tree, 12, nil)
	if c.Span() != 12 {
		t.Error("span should be 12")
	}
	if c.Offset() != 0 {
		t.Error("offset should be 0")
	}

	c.SetSpan(6)
	if c.Span() != 6 {
		t.Error("span should update")
	}

	c.SetOffset(3)
	if c.Offset() != 3 {
		t.Error("offset should update")
	}

	// Reset offset
	c.SetOffset(0)
	if c.Offset() != 0 {
		t.Error("offset should be 0")
	}
}

func TestColClamp(t *testing.T) {
	tree := newTestTree()
	c := NewCol(tree, 0, nil)
	if c.Span() != 1 {
		t.Error("span should be clamped to 1")
	}
	c2 := NewCol(tree, 30, nil)
	if c2.Span() != 24 {
		t.Error("span should be clamped to 24")
	}
	c.SetSpan(0)
	if c.Span() != 1 {
		t.Error("SetSpan should clamp to 1")
	}
	c.SetSpan(30)
	if c.Span() != 24 {
		t.Error("SetSpan should clamp to 24")
	}
}

func TestRowDraw(t *testing.T) {
	tree := newTestTree()
	r := NewRow(tree, nil)
	buf := render.NewCommandBuffer()
	r.Draw(buf)
}

func TestColDraw(t *testing.T) {
	tree := newTestTree()
	c := NewCol(tree, 12, nil)
	buf := render.NewCommandBuffer()
	c.Draw(buf)
}

// --- Layout tests ---

func TestLayout(t *testing.T) {
	tree := newTestTree()
	l := NewLayout(tree, nil)
	if l.Style().Display != layout.DisplayFlex {
		t.Error("layout should be flex")
	}
	l.SetBgColor(uimath.ColorWhite)
	tree.SetLayout(l.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 800, 600),
	})
	buf := render.NewCommandBuffer()
	l.Draw(buf)
	if buf.Len() != 1 {
		t.Errorf("expected 1 rect, got %d", buf.Len())
	}
}

func TestLayoutDrawTransparent(t *testing.T) {
	tree := newTestTree()
	l := NewLayout(tree, nil)
	tree.SetLayout(l.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 800, 600),
	})
	buf := render.NewCommandBuffer()
	l.Draw(buf)
	if buf.Len() != 0 {
		t.Error("transparent layout should not draw rect")
	}
}

func TestLayoutDrawEmpty(t *testing.T) {
	tree := newTestTree()
	l := NewLayout(tree, nil)
	l.SetBgColor(uimath.ColorWhite)
	buf := render.NewCommandBuffer()
	l.Draw(buf)
	if buf.Len() != 0 {
		t.Error("empty bounds should not draw")
	}
}

func TestHeader(t *testing.T) {
	tree := newTestTree()
	h := NewHeader(tree, nil)
	if h.Style().Display != layout.DisplayFlex {
		t.Error("header should be flex")
	}
	h.SetBgColor(uimath.ColorBlue)
	h.SetHeight(80)
	tree.SetLayout(h.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 800, 80),
	})
	buf := render.NewCommandBuffer()
	h.Draw(buf)
	if buf.Len() != 1 {
		t.Errorf("expected 1 rect, got %d", buf.Len())
	}
}

func TestHeaderDrawTransparent(t *testing.T) {
	tree := newTestTree()
	h := NewHeader(tree, nil)
	tree.SetLayout(h.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 800, 64),
	})
	buf := render.NewCommandBuffer()
	h.Draw(buf)
	if buf.Len() != 0 {
		t.Error("transparent header should not draw")
	}
}

func TestContent(t *testing.T) {
	tree := newTestTree()
	c := NewContent(tree, nil)
	if c.Style().FlexGrow != 1 {
		t.Error("content should have FlexGrow=1")
	}
	c.SetBgColor(uimath.ColorGray)
	tree.SetLayout(c.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 800, 400),
	})
	buf := render.NewCommandBuffer()
	c.Draw(buf)
	if buf.Len() != 1 {
		t.Errorf("expected 1 rect, got %d", buf.Len())
	}
}

func TestContentDrawTransparent(t *testing.T) {
	tree := newTestTree()
	c := NewContent(tree, nil)
	tree.SetLayout(c.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 800, 400),
	})
	buf := render.NewCommandBuffer()
	c.Draw(buf)
	if buf.Len() != 0 {
		t.Error("transparent content should not draw")
	}
}

func TestFooter(t *testing.T) {
	tree := newTestTree()
	f := NewFooter(tree, nil)
	f.SetBgColor(uimath.ColorBlack)
	f.SetHeight(60)
	tree.SetLayout(f.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 800, 60),
	})
	buf := render.NewCommandBuffer()
	f.Draw(buf)
	if buf.Len() != 1 {
		t.Errorf("expected 1 rect, got %d", buf.Len())
	}
}

func TestFooterDrawTransparent(t *testing.T) {
	tree := newTestTree()
	f := NewFooter(tree, nil)
	tree.SetLayout(f.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 800, 48),
	})
	buf := render.NewCommandBuffer()
	f.Draw(buf)
	if buf.Len() != 0 {
		t.Error("transparent footer should not draw")
	}
}

func TestAside(t *testing.T) {
	tree := newTestTree()
	a := NewAside(tree, nil)
	a.SetBgColor(uimath.ColorGray)
	a.SetWidth(250)
	tree.SetLayout(a.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 250, 600),
	})
	buf := render.NewCommandBuffer()
	a.Draw(buf)
	if buf.Len() != 1 {
		t.Errorf("expected 1 rect, got %d", buf.Len())
	}
}

func TestAsideDrawTransparent(t *testing.T) {
	tree := newTestTree()
	a := NewAside(tree, nil)
	tree.SetLayout(a.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 200, 600),
	})
	buf := render.NewCommandBuffer()
	a.Draw(buf)
	if buf.Len() != 0 {
		t.Error("transparent aside should not draw")
	}
}

// --- Popup tests ---

func TestPopup(t *testing.T) {
	tree := newTestTree()
	p := NewPopup(tree, nil)

	if p.IsVisible() {
		t.Error("popup should start hidden")
	}
	if p.Placement() != PlacementBottom {
		t.Error("default placement should be bottom")
	}

	p.SetVisible(true)
	if !p.IsVisible() {
		t.Error("should be visible")
	}

	p.SetPlacement(PlacementTop)
	if p.Placement() != PlacementTop {
		t.Error("placement should update")
	}

	p.SetBgColor(uimath.ColorWhite)
	p.SetShadow(false)
}

func TestPopupAnchor(t *testing.T) {
	tree := newTestTree()
	anchor := tree.CreateElement(core.TypeButton)
	p := NewPopup(tree, nil)
	p.SetAnchor(anchor)
	if p.AnchorID() != anchor {
		t.Error("anchor should be set")
	}
}

func TestPopupUpdatePosition(t *testing.T) {
	tree := newTestTree()
	anchor := tree.CreateElement(core.TypeButton)
	tree.SetLayout(anchor, core.LayoutResult{
		Bounds: uimath.NewRect(100, 50, 80, 32),
	})

	p := NewPopup(tree, nil)
	p.SetAnchor(anchor)
	tree.SetLayout(p.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 120, 40),
	})

	for _, pl := range []PopupPlacement{
		PlacementTop, PlacementBottom, PlacementLeft, PlacementRight,
		PlacementTopStart, PlacementTopEnd, PlacementBottomStart, PlacementBottomEnd,
	} {
		p.SetPlacement(pl)
		p.UpdatePosition()
	}
}

func TestPopupUpdatePositionNoAnchor(t *testing.T) {
	tree := newTestTree()
	p := NewPopup(tree, nil)
	p.UpdatePosition() // should not panic
}

func TestPopupDraw(t *testing.T) {
	tree := newTestTree()
	p := NewPopup(tree, nil)
	p.SetVisible(true)
	tree.SetLayout(p.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(100, 82, 120, 40),
	})
	buf := render.NewCommandBuffer()
	p.Draw(buf)
	if buf.Len() != 1 {
		t.Errorf("expected 1 rect, got %d", buf.Len())
	}
}

func TestPopupDrawHidden(t *testing.T) {
	tree := newTestTree()
	p := NewPopup(tree, nil)
	tree.SetLayout(p.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 120, 40),
	})
	buf := render.NewCommandBuffer()
	p.Draw(buf)
	if buf.Len() != 0 {
		t.Error("hidden popup should not draw")
	}
}

func TestPopupDrawEmpty(t *testing.T) {
	tree := newTestTree()
	p := NewPopup(tree, nil)
	p.SetVisible(true)
	buf := render.NewCommandBuffer()
	p.Draw(buf)
	if buf.Len() != 0 {
		t.Error("empty bounds should not draw")
	}
}

// --- Tooltip tests ---

func TestTooltip(t *testing.T) {
	tree := newTestTree()
	anchor := tree.CreateElement(core.TypeButton)
	tt := NewTooltip(tree, "Hint", anchor, nil)

	if tt.Text() != "Hint" {
		t.Error("text mismatch")
	}
	if tt.IsVisible() {
		t.Error("should start hidden")
	}
	if tt.Placement() != PlacementTop {
		t.Error("default placement should be top")
	}
	if tt.AnchorID() != anchor {
		t.Error("anchor mismatch")
	}

	tt.SetText("Updated")
	if tt.Text() != "Updated" {
		t.Error("text should update")
	}

	tt.SetPlacement(PlacementBottom)
	if tt.Placement() != PlacementBottom {
		t.Error("placement should update")
	}
}

func TestTooltipShowHide(t *testing.T) {
	tree := newTestTree()
	anchor := tree.CreateElement(core.TypeButton)
	tt := NewTooltip(tree, "Hint", anchor, nil)

	tt.Show()
	if !tt.IsVisible() {
		t.Error("should be visible after Show")
	}

	tt.Hide()
	if tt.IsVisible() {
		t.Error("should be hidden after Hide")
	}
}

func TestTooltipHoverShowHide(t *testing.T) {
	tree := newTestTree()
	anchor := tree.CreateElement(core.TypeButton)
	tt := NewTooltip(tree, "Hint", anchor, nil)

	// MouseEnter on anchor
	handlers := tree.Handlers(anchor, event.MouseEnter)
	for _, h := range handlers {
		h(&event.Event{Type: event.MouseEnter})
	}
	if !tt.IsVisible() {
		t.Error("should show on mouse enter")
	}

	// MouseLeave on anchor
	handlers = tree.Handlers(anchor, event.MouseLeave)
	for _, h := range handlers {
		h(&event.Event{Type: event.MouseLeave})
	}
	if tt.IsVisible() {
		t.Error("should hide on mouse leave")
	}
}

func TestTooltipDraw(t *testing.T) {
	tree := newTestTree()
	anchor := tree.CreateElement(core.TypeButton)
	tt := NewTooltip(tree, "Hint", anchor, nil)
	tt.Show()
	tree.SetLayout(tt.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(100, 30, 60, 24),
	})

	buf := render.NewCommandBuffer()
	tt.Draw(buf)
	// rect + text
	if buf.Len() != 2 {
		t.Errorf("expected 2 commands, got %d", buf.Len())
	}
}

func TestTooltipDrawHidden(t *testing.T) {
	tree := newTestTree()
	anchor := tree.CreateElement(core.TypeButton)
	tt := NewTooltip(tree, "Hint", anchor, nil)
	tree.SetLayout(tt.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 60, 24),
	})
	buf := render.NewCommandBuffer()
	tt.Draw(buf)
	if buf.Len() != 0 {
		t.Error("hidden tooltip should not draw")
	}
}

func TestTooltipDrawEmpty(t *testing.T) {
	tree := newTestTree()
	anchor := tree.CreateElement(core.TypeButton)
	tt := NewTooltip(tree, "Hint", anchor, nil)
	tt.Show()
	buf := render.NewCommandBuffer()
	tt.Draw(buf)
	if buf.Len() != 0 {
		t.Error("empty bounds should not draw")
	}
}

func TestTooltipDrawEmptyText(t *testing.T) {
	tree := newTestTree()
	anchor := tree.CreateElement(core.TypeButton)
	tt := NewTooltip(tree, "", anchor, nil)
	tt.Show()
	tree.SetLayout(tt.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 60, 24),
	})
	buf := render.NewCommandBuffer()
	tt.Draw(buf)
	// Just rect, no text
	if buf.Len() != 1 {
		t.Errorf("expected 1 rect for empty tooltip text, got %d", buf.Len())
	}
}

// --- Test helpers ---

type testWidget struct {
	Base
}

func (w *testWidget) Draw(buf *render.CommandBuffer) {}

type drawCounter struct {
	count int
}

func (d *drawCounter) ElementID() core.ElementID  { return 0 }
func (d *drawCounter) Style() layout.Style         { return layout.Style{} }
func (d *drawCounter) SetStyle(layout.Style)       {}
func (d *drawCounter) Children() []Widget           { return nil }
func (d *drawCounter) Destroy()                     {}
func (d *drawCounter) Draw(buf *render.CommandBuffer) {
	d.count++
}

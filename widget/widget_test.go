package widget

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
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
	// rect + pushclip + popclip
	if buf.Len() != 3 {
		t.Errorf("expected 3 commands (rect+pushclip+popclip), got %d", buf.Len())
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
	// 1 bg rect + 1 push clip + 1 pop clip = 3 commands
	if buf.Len() != 3 {
		t.Errorf("expected 3 commands (bg+pushclip+popclip), got %d", buf.Len())
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
	// push clip + pop clip = 2 commands (no bg)
	if buf.Len() != 2 {
		t.Errorf("expected 2 commands (pushclip+popclip), got %d", buf.Len())
	}
}

func TestContentScroll(t *testing.T) {
	tree := newTestTree()
	c := NewContent(tree, nil)
	c.SetBgColor(uimath.ColorWhite)
	tree.SetLayout(c.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 800, 400),
	})
	c.SetContentHeight(1000)

	// Initial scroll is 0
	if c.ScrollY() != 0 {
		t.Errorf("initial scrollY should be 0, got %f", c.ScrollY())
	}

	// Scroll down via wheel (negative dy = scroll down)
	c.HandleWheel(-3) // 3 ticks * 40px = 120px
	if c.ScrollY() != 120 {
		t.Errorf("expected scrollY=120, got %f", c.ScrollY())
	}

	// Scroll up
	c.HandleWheel(1) // 1 tick * 40px = 40px up
	if c.ScrollY() != 80 {
		t.Errorf("expected scrollY=80, got %f", c.ScrollY())
	}

	// Clamp to 0
	c.HandleWheel(100)
	if c.ScrollY() != 0 {
		t.Errorf("scrollY should clamp to 0, got %f", c.ScrollY())
	}

	// Clamp to max (1000 - 400 = 600)
	c.HandleWheel(-1000)
	if c.ScrollY() != 600 {
		t.Errorf("scrollY should clamp to 600, got %f", c.ScrollY())
	}

	// ScrollTo
	c.ScrollTo(200)
	if c.ScrollY() != 200 {
		t.Errorf("expected scrollY=200, got %f", c.ScrollY())
	}

	// ScrollBy
	c.ScrollBy(50)
	if c.ScrollY() != 250 {
		t.Errorf("expected scrollY=250, got %f", c.ScrollY())
	}

	// needsScroll
	if !c.needsScroll() {
		t.Error("should need scroll when contentHeight > bounds.Height")
	}

	c.SetContentHeight(200) // less than viewport
	c.ScrollBy(0)           // re-clamp
	if c.ScrollY() != 0 {
		t.Errorf("scrollY should be 0 when content fits, got %f", c.ScrollY())
	}
}

func TestContentScrollbarDraw(t *testing.T) {
	tree := newTestTree()
	c := NewContent(tree, nil)
	c.SetBgColor(uimath.ColorWhite)
	tree.SetLayout(c.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 800, 400),
	})
	c.SetContentHeight(1000)

	buf := render.NewCommandBuffer()
	c.Draw(buf)
	// bg + pushclip + scrollbar track + scrollbar thumb + popclip = 5
	if buf.Len() != 5 {
		t.Errorf("expected 5 commands with scrollbar, got %d", buf.Len())
	}
}

func TestContentScrollbarDrag(t *testing.T) {
	tree := newTestTree()
	c := NewContent(tree, nil)
	tree.SetLayout(c.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 800, 400),
	})
	c.SetContentHeight(1000)

	// Not dragging initially
	if c.IsScrollBarDragging() {
		t.Error("should not be dragging initially")
	}

	// Simulate drag on thumb (thumb is at top when scrollY=0)
	thumbY, _ := c.thumbRect(c.Bounds())
	c.HandleScrollBarDown(thumbY + 5)
	if !c.IsScrollBarDragging() {
		t.Error("should be dragging after HandleScrollBarDown on thumb")
	}

	// Move
	c.HandleScrollBarMove(thumbY + 50)
	if c.ScrollY() == 0 {
		t.Error("scrollY should have changed during drag")
	}

	// Release
	c.HandleScrollBarUp()
	if c.IsScrollBarDragging() {
		t.Error("should not be dragging after HandleScrollBarUp")
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

// mockTextDrawer implements TextDrawer for testing.
type mockTextDrawer struct{}

func (m *mockTextDrawer) DrawText(buf *render.CommandBuffer, text string, x, y, fontSize, maxWidth float32, color uimath.Color, opacity float32) {
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(x, y, maxWidth, fontSize*1.2),
		FillColor: color,
	}, 0, opacity)
}
func (m *mockTextDrawer) LineHeight(fontSize float32) float32 { return fontSize * 1.2 }
func (m *mockTextDrawer) MeasureText(text string, fontSize float32) float32 {
	return float32(len([]rune(text))) * fontSize * 0.6
}

// cfgWithTextRenderer returns a config with a mock text renderer.
func cfgWithTextRenderer() *Config {
	cfg := DefaultConfig()
	cfg.TextRenderer = &mockTextDrawer{}
	return cfg
}

// === New widget tests ===

// setBounds is a test helper that sets layout bounds on a widget.
func setBounds(tree *core.Tree, w Widget, x, y, width, height float32) {
	tree.SetLayout(w.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(x, y, width, height),
	})
}

// --- Checkbox ---

func TestCheckbox(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	cb := NewCheckbox(tree, "Accept", cfg)

	if cb.Label() != "Accept" {t.Errorf("got '%s'", cb.Label())}
	if cb.IsChecked() {t.Error("should not be checked")}
	if cb.IsDisabled() {t.Error("should not be disabled")}

	cb.SetChecked(true)
	if !cb.IsChecked() {t.Error("should be checked")}
	cb.SetDisabled(true)
	if !cb.IsDisabled() {t.Error("should be disabled")}
	cb.SetLabel("New Label")
	if cb.Label() != "New Label" {t.Errorf("got '%s'", cb.Label())}
}

func TestCheckboxNilConfig(t *testing.T) {
	tree := newTestTree()
	cb := NewCheckbox(tree, "Test", nil)
	if cb.Config() == nil {t.Error("nil config should default")}
}

func TestCheckboxClick(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	cb := NewCheckbox(tree, "Accept", cfg)
	var changed bool
	cb.OnChange(func(checked bool) { changed = checked })

	dispatcher := core.NewDispatcher(tree)
	dispatcher.Dispatch(cb.ElementID(), &event.Event{Type: event.MouseClick})
	if !cb.IsChecked() || !changed {t.Error("click should toggle")}

	// Click when disabled: no toggle
	cb.SetDisabled(true)
	dispatcher.Dispatch(cb.ElementID(), &event.Event{Type: event.MouseClick})
	if !cb.IsChecked() {t.Error("should stay checked when disabled")}
}

func TestCheckboxDraw(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	buf := render.NewCommandBuffer()

	// Draw unchecked
	cb := NewCheckbox(tree, "Label", cfg)
	setBounds(tree, cb, 10, 10, 200, 32)
	cb.Draw(buf)

	// Draw checked
	cb.SetChecked(true)
	cb.Draw(buf)

	// Draw disabled + checked
	cb.SetDisabled(true)
	cb.Draw(buf)

	// Draw disabled + unchecked
	cb.SetChecked(false)
	cb.Draw(buf)

	// Draw empty label
	cb2 := NewCheckbox(tree, "", cfg)
	setBounds(tree, cb2, 10, 50, 200, 32)
	cb2.Draw(buf)

	// Draw empty bounds (should be no-op)
	cb3 := NewCheckbox(tree, "X", cfg)
	cb3.Draw(buf)
}

// --- Radio ---

func TestRadio(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	r := NewRadio(tree, "Option A", cfg)

	if r.Label() != "Option A" {t.Errorf("got '%s'", r.Label())}
	if r.IsChecked() {t.Error("should not be checked")}
	if r.IsDisabled() {t.Error("should not be disabled")}

	r.SetChecked(true)
	if !r.IsChecked() {t.Error("should be checked")}
	r.SetLabel("Option B")
	if r.Label() != "Option B" {t.Errorf("got '%s'", r.Label())}
	r.SetDisabled(true)
	if !r.IsDisabled() {t.Error("should be disabled")}
	var called bool
	r.OnChange(func(bool) { called = true })
	_ = called
}

func TestRadioGroup(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	r1 := NewRadio(tree, "A", cfg)
	r2 := NewRadio(tree, "B", cfg)

	group := NewRadioGroup()
	group.Add(r1)
	group.Add(r2)

	var selected string
	group.OnChange(func(v string) { selected = v })

	group.SetValue("A")
	if !r1.IsChecked() || r2.IsChecked() {t.Error("only r1 should be checked")}
	if group.Value() != "A" {t.Errorf("got '%s'", group.Value())}

	dispatcher := core.NewDispatcher(tree)
	dispatcher.Dispatch(r2.ElementID(), &event.Event{Type: event.MouseClick})
	if r1.IsChecked() || !r2.IsChecked() {t.Error("only r2 checked after click")}
	if selected != "B" {t.Errorf("got '%s'", selected)}

	// Click disabled radio
	r1.SetDisabled(true)
	dispatcher.Dispatch(r1.ElementID(), &event.Event{Type: event.MouseClick})
	if r1.IsChecked() {t.Error("disabled radio should not check")}
}

func TestRadioDraw(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	buf := render.NewCommandBuffer()

	r := NewRadio(tree, "Opt", cfg)
	setBounds(tree, r, 10, 10, 200, 32)

	// Draw unchecked
	r.Draw(buf)

	// Draw checked
	r.SetChecked(true)
	r.Draw(buf)

	// Draw disabled + checked
	r.SetDisabled(true)
	r.Draw(buf)

	// Draw disabled + unchecked
	r.SetChecked(false)
	r.Draw(buf)

	// Draw empty label
	r2 := NewRadio(tree, "", cfg)
	setBounds(tree, r2, 10, 50, 200, 32)
	r2.Draw(buf)

	// No bounds
	r3 := NewRadio(tree, "X", cfg)
	r3.Draw(buf)
}

// --- Switch ---

func TestSwitch(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	sw := NewSwitch(tree, cfg)

	if sw.IsChecked() {t.Error("should not be checked")}
	if sw.IsDisabled() {t.Error("should not be disabled")}
	sw.SetChecked(true)
	if !sw.IsChecked() {t.Error("should be checked")}

	var toggled bool
	sw.OnChange(func(checked bool) { toggled = checked })
	sw.SetChecked(false)

	dispatcher := core.NewDispatcher(tree)
	dispatcher.Dispatch(sw.ElementID(), &event.Event{Type: event.MouseClick})
	if !sw.IsChecked() || !toggled {t.Error("click should toggle")}

	sw.SetDisabled(true)
	dispatcher.Dispatch(sw.ElementID(), &event.Event{Type: event.MouseClick})
	if !sw.IsChecked() {t.Error("should stay checked when disabled")}
}

func TestSwitchDraw(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	buf := render.NewCommandBuffer()

	sw := NewSwitch(tree, cfg)
	setBounds(tree, sw, 10, 10, 44, 22)

	// Unchecked
	sw.Draw(buf)
	// Checked
	sw.SetChecked(true)
	sw.Draw(buf)
	// Disabled + checked
	sw.SetDisabled(true)
	sw.Draw(buf)
	// Disabled + unchecked
	sw.SetChecked(false)
	sw.Draw(buf)
	// No bounds
	sw2 := NewSwitch(tree, cfg)
	sw2.Draw(buf)
}

// --- Image ---

func TestImage(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	img := NewImage(tree, 42, cfg)

	if img.Texture() != 42 {t.Errorf("got %d", img.Texture())}
	if img.Fit() != ImageFitFill {t.Error("default fit should be fill")}

	img.SetTexture(100)
	img.SetFit(ImageFitContain)
	img.SetTint(uimath.ColorWhite)
	img.SetSrcRect(uimath.NewRect(0, 0, 64, 64))
}

func TestImageDraw(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	buf := render.NewCommandBuffer()

	img := NewImage(tree, 42, cfg)
	setBounds(tree, img, 10, 10, 100, 80)
	img.Draw(buf)

	// With contain fit
	img.SetFit(ImageFitContain)
	img.Draw(buf)

	// Invalid texture
	img2 := NewImage(tree, 0, cfg) // InvalidTexture = 0
	setBounds(tree, img2, 10, 10, 100, 80)
	img2.Draw(buf)

	// No bounds
	img3 := NewImage(tree, 42, cfg)
	img3.Draw(buf)

	// Nil config
	img4 := NewImage(tree, 42, nil)
	if img4.Config() == nil {t.Error("nil config should default")}
}

// --- Tag ---

func TestTag(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	tag := NewTag(tree, "New", cfg)

	if tag.Label() != "New" {t.Errorf("got '%s'", tag.Label())}
	tag.SetTagType(TagSuccess)
	if tag.TagType() != TagSuccess {t.Error("expected TagSuccess")}
	tag.SetLabel("Updated")
	tag.SetColor(uimath.ColorHex("#ff0000"))
}

func TestTagDraw(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	buf := render.NewCommandBuffer()

	// All tag types
	types := []TagType{TagDefault, TagSuccess, TagWarning, TagError, TagProcessing}
	for _, tt := range types {
		tag := NewTag(tree, "Tag", cfg)
		tag.SetTagType(tt)
		setBounds(tree, tag, 10, 10, 80, 22)
		tag.Draw(buf)
	}

	// Custom color
	tag2 := NewTag(tree, "Custom", cfg)
	tag2.SetColor(uimath.ColorHex("#ff0000"))
	setBounds(tree, tag2, 10, 10, 80, 22)
	tag2.Draw(buf)

	// Empty label
	tag3 := NewTag(tree, "", cfg)
	setBounds(tree, tag3, 10, 10, 80, 22)
	tag3.Draw(buf)

	// No bounds
	tag4 := NewTag(tree, "X", cfg)
	tag4.Draw(buf)
}

// --- Progress ---

func TestProgress(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	p := NewProgress(tree, cfg)

	if p.Percent() != 0 {t.Errorf("got %f", p.Percent())}
	p.SetPercent(75)
	if p.Percent() != 75 {t.Errorf("got %f", p.Percent())}
	p.SetPercent(150) // clamp
	if p.Percent() != 100 {t.Errorf("got %f", p.Percent())}
	p.SetPercent(-10) // clamp
	if p.Percent() != 0 {t.Errorf("got %f", p.Percent())}

	p.SetStatus(ProgressSuccess)
	if p.Status() != ProgressSuccess {t.Error("expected ProgressSuccess")}
}

func TestProgressDraw(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	buf := render.NewCommandBuffer()

	p := NewProgress(tree, cfg)
	setBounds(tree, p, 10, 10, 200, 8)

	// 0%
	p.Draw(buf)

	// 50%
	p.SetPercent(50)
	p.Draw(buf)

	// All statuses
	statuses := []ProgressStatus{ProgressNormal, ProgressSuccess, ProgressError, ProgressActive}
	for _, s := range statuses {
		p.SetStatus(s)
		p.SetPercent(75)
		p.Draw(buf)
	}

	// No bounds
	p2 := NewProgress(tree, cfg)
	p2.SetPercent(50)
	p2.Draw(buf)
}

// --- Loading ---

func TestLoading(t *testing.T) {
	tree := newTestTree()
	l := NewLoading(tree, nil)
	l.SetTip("Loading...")
	if l.Tip() != "Loading..." {t.Errorf("got '%s'", l.Tip())}
}

func TestLoadingDraw(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	buf := render.NewCommandBuffer()

	l := NewLoading(tree, cfg)
	setBounds(tree, l, 10, 10, 200, 100)
	l.Draw(buf)

	// With tip
	l.SetTip("Please wait...")
	l.Draw(buf)

	// No bounds
	l2 := NewLoading(tree, cfg)
	l2.Draw(buf)
}

// --- Empty ---

func TestEmpty(t *testing.T) {
	tree := newTestTree()
	e := NewEmpty(tree, nil)
	if e.Description() != "暂无数据" {t.Errorf("got '%s'", e.Description())}
	e.SetDescription("No data")
	if e.Description() != "No data" {t.Errorf("got '%s'", e.Description())}
}

func TestEmptyDraw(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	buf := render.NewCommandBuffer()

	e := NewEmpty(tree, cfg)
	setBounds(tree, e, 10, 10, 300, 120)
	e.Draw(buf)

	// Custom description
	e.SetDescription("Nothing here")
	e.Draw(buf)

	// Empty description
	e.SetDescription("")
	e.Draw(buf)

	// No bounds
	e2 := NewEmpty(tree, cfg)
	e2.Draw(buf)
}

// --- Tabs ---

func TestTabs(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	items := []TabItem{
		{Key: "tab1", Label: "Tab 1"},
		{Key: "tab2", Label: "Tab 2"},
	}
	tabs := NewTabs(tree, items, cfg)

	if tabs.ActiveKey() != "tab1" {t.Errorf("got '%s'", tabs.ActiveKey())}
	tabs.SetActiveKey("tab2")
	if tabs.ActiveKey() != "tab2" {t.Errorf("got '%s'", tabs.ActiveKey())}

	var changed string
	tabs.OnChange(func(key string) { changed = key })
	_ = changed

	tabs.Destroy()
}

func TestTabsDraw(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	buf := render.NewCommandBuffer()

	items := []TabItem{
		{Key: "t1", Label: "First"},
		{Key: "t2", Label: "Second"},
		{Key: "t3", Label: "Third"},
	}
	tabs := NewTabs(tree, items, cfg)
	setBounds(tree, tabs, 10, 10, 400, 300)
	tabs.Draw(buf)

	// Switch tab
	tabs.SetActiveKey("t2")
	tabs.Draw(buf)

	// No bounds
	tabs2 := NewTabs(tree, items, cfg)
	tabs2.Draw(buf)

	// Empty tabs
	tabs3 := NewTabs(tree, nil, cfg)
	setBounds(tree, tabs3, 10, 10, 400, 300)
	tabs3.Draw(buf)

	tabs.Destroy()
	tabs2.Destroy()
	tabs3.Destroy()
}

// --- Dialog ---

func TestDialog(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	d := NewDialog(tree, "Confirm", cfg)

	if d.Title() != "Confirm" {t.Errorf("got '%s'", d.Title())}
	if d.IsVisible() {t.Error("should not be visible")}

	d.Open()
	if !d.IsVisible() {t.Error("should be visible")}
	d.Close()
	if d.IsVisible() {t.Error("should not be visible")}

	d.SetTitle("Delete?")
	d.SetWidth(600)
	content := NewText(tree, "Are you sure?", cfg)
	d.SetContent(content)

	var closed bool
	d.OnClose(func() { closed = true })
	_ = closed
}

func TestDialogDraw(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	buf := render.NewCommandBuffer()

	d := NewDialog(tree, "Test Dialog", cfg)
	setBounds(tree, d, 0, 0, 800, 600)

	// Closed - should not draw
	d.Draw(buf)

	// Open
	d.Open()
	d.Draw(buf)

	// With content
	content := NewText(tree, "Content", cfg)
	d.SetContent(content)
	d.Draw(buf)

	// No bounds
	d2 := NewDialog(tree, "X", cfg)
	d2.Open()
	d2.Draw(buf)
}

// --- Message ---

func TestMessage(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	m := NewMessage(tree, "OK", cfg)

	if m.Content() != "OK" {t.Errorf("got '%s'", m.Content())}
	if m.MsgType() != MessageInfo {t.Error("expected MessageInfo")}
	if !m.IsVisible() {t.Error("should be visible")}

	m.SetMsgType(MessageSuccess)
	m.SetContent("Done!")
	m.SetVisible(false)
	if m.IsVisible() {t.Error("should not be visible")}
}

func TestMessageDraw(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	buf := render.NewCommandBuffer()

	// All types
	types := []MessageType{MessageInfo, MessageSuccess, MessageWarning, MessageError}
	for _, mt := range types {
		m := NewMessage(tree, "Test message", cfg)
		m.SetMsgType(mt)
		setBounds(tree, m, 100, 10, 300, 36)
		m.Draw(buf)
	}

	// Hidden
	m2 := NewMessage(tree, "Hidden", cfg)
	m2.SetVisible(false)
	setBounds(tree, m2, 100, 10, 300, 36)
	m2.Draw(buf) // no-op

	// Empty content
	m3 := NewMessage(tree, "", cfg)
	setBounds(tree, m3, 100, 10, 300, 36)
	m3.Draw(buf)

	// No bounds
	m4 := NewMessage(tree, "X", cfg)
	m4.Draw(buf)
}

// --- Select ---

func TestSelect(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	opts := []SelectOption{
		{Label: "Apple", Value: "apple"},
		{Label: "Banana", Value: "banana"},
		{Label: "Cherry", Value: "cherry", Disabled: true},
	}
	s := NewSelect(tree, opts, cfg)

	if s.Value() != "" {t.Error("should have no value")}
	if s.Placeholder() != "请选择" {t.Errorf("got '%s'", s.Placeholder())}
	if s.IsDisabled() {t.Error("should not be disabled")}
	if s.IsOpen() {t.Error("should not be open")}
	if len(s.Options()) != 3 {t.Errorf("got %d", len(s.Options()))}

	s.SetValue("banana")
	if s.Value() != "banana" {t.Errorf("got '%s'", s.Value())}
	s.SetPlaceholder("Choose...")
	s.SetDisabled(true)

	s.SetOptions([]SelectOption{{Label: "X", Value: "x"}})
	if len(s.Options()) != 1 {t.Errorf("got %d", len(s.Options()))}

	var changed string
	s.OnChange(func(v string) { changed = v })
	_ = changed
	s.Destroy()
}

func TestSelectDraw(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	buf := render.NewCommandBuffer()
	opts := []SelectOption{
		{Label: "Apple", Value: "apple"},
		{Label: "Banana", Value: "banana"},
	}

	// Closed, no value (placeholder)
	s := NewSelect(tree, opts, cfg)
	setBounds(tree, s, 10, 10, 200, 32)
	s.Draw(buf)

	// With value
	s.SetValue("apple")
	s.Draw(buf)

	// Disabled
	s.SetDisabled(true)
	s.Draw(buf)

	// Open dropdown
	s2 := NewSelect(tree, opts, cfg)
	setBounds(tree, s2, 10, 10, 200, 32)
	s2.SetValue("banana")
	// Manually open
	dispatcher := core.NewDispatcher(tree)
	s2.SetDisabled(false)
	dispatcher.Dispatch(s2.ElementID(), &event.Event{Type: event.MouseClick})
	if !s2.IsOpen() {t.Error("should be open after click")}
	s2.Draw(buf)

	// Close
	dispatcher.Dispatch(s2.ElementID(), &event.Event{Type: event.MouseClick})
	if s2.IsOpen() {t.Error("should be closed after second click")}

	// No bounds
	s3 := NewSelect(tree, opts, cfg)
	s3.Draw(buf)

	s.Destroy()
	s2.Destroy()
	s3.Destroy()
}

// --- TextArea ---

func TestTextArea(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	ta := NewTextArea(tree, cfg)

	if ta.Value() != "" {t.Error("should be empty")}
	if ta.Rows() != 4 {t.Errorf("got %d", ta.Rows())}
	if ta.Placeholder() != "" {t.Error("should have no placeholder")}

	ta.SetValue("Hello\nWorld")
	if ta.Value() != "Hello\nWorld" {t.Errorf("got '%s'", ta.Value())}
	ta.SetPlaceholder("Enter text...")
	ta.SetRows(6)
	if ta.Rows() != 6 {t.Errorf("got %d", ta.Rows())}
	ta.SetRows(0) // should clamp to 1
	if ta.Rows() != 1 {t.Errorf("got %d", ta.Rows())}

	ta.SetDisabled(true)
	if !ta.IsDisabled() {t.Error("should be disabled")}

	var changed string
	ta.OnChange(func(v string) { changed = v })
	_ = changed
}

func TestTextAreaEditing(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	ta := NewTextArea(tree, cfg)
	setBounds(tree, ta, 10, 10, 300, 200)

	dispatcher := core.NewDispatcher(tree)

	// Type characters
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.KeyPress, Char: 'H'})
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.KeyPress, Char: 'i'})
	if ta.Value() != "Hi" {t.Errorf("got '%s'", ta.Value())}

	// Enter for newline
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.KeyDown, Key: event.KeyEnter})
	if ta.Value() != "Hi\n" {t.Errorf("got '%s'", ta.Value())}

	// Type on second line
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.KeyPress, Char: 'X'})
	if ta.Value() != "Hi\nX" {t.Errorf("got '%s'", ta.Value())}

	// Backspace
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.KeyDown, Key: event.KeyBackspace})
	if ta.Value() != "Hi\n" {t.Errorf("got '%s'", ta.Value())}

	// Delete (forward)
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.KeyDown, Key: event.KeyHome})
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.KeyDown, Key: event.KeyDelete})

	// Arrow keys
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.KeyDown, Key: event.KeyArrowLeft})
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.KeyDown, Key: event.KeyArrowRight})
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.KeyDown, Key: event.KeyArrowUp})
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.KeyDown, Key: event.KeyArrowDown})
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.KeyDown, Key: event.KeyEnd})

	// Select all
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.KeyDown, Key: event.KeyA, Modifiers: event.Modifiers{Ctrl: true}})

	// Shift+Arrow selection
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.KeyDown, Key: event.KeyHome})
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.KeyDown, Key: event.KeyArrowRight, Modifiers: event.Modifiers{Shift: true}})
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.KeyDown, Key: event.KeyArrowLeft, Modifiers: event.Modifiers{Shift: true}})
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.KeyDown, Key: event.KeyArrowDown, Modifiers: event.Modifiers{Shift: true}})
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.KeyDown, Key: event.KeyArrowUp, Modifiers: event.Modifiers{Shift: true}})
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.KeyDown, Key: event.KeyHome, Modifiers: event.Modifiers{Shift: true}})
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.KeyDown, Key: event.KeyEnd, Modifiers: event.Modifiers{Shift: true}})

	// Delete selection
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.KeyDown, Key: event.KeyA, Modifiers: event.Modifiers{Ctrl: true}})
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.KeyDown, Key: event.KeyBackspace})

	// Disabled
	ta.SetDisabled(true)
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.KeyPress, Char: 'Z'})
	if ta.Value() != "" {t.Errorf("should not type when disabled, got '%s'", ta.Value())}
}

func TestTextAreaIME(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	ta := NewTextArea(tree, cfg)
	setBounds(tree, ta, 10, 10, 300, 200)

	dispatcher := core.NewDispatcher(tree)
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.IMECompositionEnd, Text: "你好"})
	if ta.Value() != "你好" {t.Errorf("got '%s'", ta.Value())}

	// Disabled
	ta.SetDisabled(true)
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.IMECompositionEnd, Text: "世界"})
	if ta.Value() != "你好" {t.Errorf("should not insert when disabled, got '%s'", ta.Value())}
}

func TestTextAreaMouse(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	ta := NewTextArea(tree, cfg)
	ta.SetValue("Hello\nWorld")
	setBounds(tree, ta, 10, 10, 300, 200)

	dispatcher := core.NewDispatcher(tree)

	// Click to focus
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.MouseClick, X: 20, Y: 20})

	// Mouse drag
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.MouseDown, X: 20, Y: 20, Button: event.MouseButtonLeft})
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.MouseMove, X: 50, Y: 20})
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.MouseUp, X: 50, Y: 20})

	// Disabled: mouse ignored
	ta.SetDisabled(true)
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.MouseClick, X: 20, Y: 20})
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.MouseDown, X: 20, Y: 20})
	dispatcher.Dispatch(ta.ElementID(), &event.Event{Type: event.MouseMove, X: 50, Y: 20})
}

func TestTextAreaDraw(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	buf := render.NewCommandBuffer()

	ta := NewTextArea(tree, cfg)
	setBounds(tree, ta, 10, 10, 300, 200)

	// Empty with placeholder
	ta.SetPlaceholder("Enter text...")
	ta.Draw(buf)

	// With content
	ta.SetValue("Line 1\nLine 2\nLine 3")
	ta.Draw(buf)

	// Focused with cursor
	tree.SetFocused(ta.ElementID(), true)
	ta.Draw(buf)

	// Disabled
	ta.SetDisabled(true)
	ta.Draw(buf)

	// No bounds
	ta2 := NewTextArea(tree, cfg)
	ta2.Draw(buf)
}

func TestTextAreaLineHelpers(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	ta := NewTextArea(tree, cfg)
	ta.SetValue("AB\nCD\nEF")
	// Set bounds wide enough so no wrapping occurs
	setBounds(tree, ta, 0, 0, 400, 200)
	contentW := ta.contentWidth()

	// visualLines
	vl := ta.visualLines(contentW)
	if len(vl) != 3 {t.Errorf("expected 3 visual lines, got %d", len(vl))}

	// runeOffsetToVisualLineCol
	line, col := ta.runeOffsetToVisualLineCol(0, contentW)
	if line != 0 || col != 0 {t.Errorf("got %d,%d", line, col)}
	line, col = ta.runeOffsetToVisualLineCol(3, contentW) // after "AB\n"
	if line != 1 || col != 0 {t.Errorf("got %d,%d", line, col)}
	line, col = ta.runeOffsetToVisualLineCol(5, contentW) // "CD" on line 1
	if line != 1 || col != 2 {t.Errorf("got %d,%d", line, col)}

	// visualLineColToRuneOffset
	if ta.visualLineColToRuneOffset(0, 0, contentW) != 0 {t.Error("bad offset")}
	if ta.visualLineColToRuneOffset(1, 0, contentW) != 3 {t.Error("bad offset")}
	if ta.visualLineColToRuneOffset(2, 1, contentW) != 7 {t.Error("bad offset")}
	// Clamping
	if ta.visualLineColToRuneOffset(99, 0, contentW) != 6 {t.Errorf("got %d", ta.visualLineColToRuneOffset(99, 0, contentW))}
	if ta.visualLineColToRuneOffset(0, 99, contentW) != 2 {t.Errorf("got %d", ta.visualLineColToRuneOffset(0, 99, contentW))}

	// lineStart / lineEnd
	if ta.lineStart(4) != 3 {t.Errorf("got %d", ta.lineStart(4))} // pos 4 = line 1
	if ta.lineEnd(4) != 5 {t.Errorf("got %d", ta.lineEnd(4))}
}

func TestTextAreaDrawSelection(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	buf := render.NewCommandBuffer()

	ta := NewTextArea(tree, cfg)
	ta.SetValue("Hello\nWorld")
	setBounds(tree, ta, 10, 10, 300, 200)
	tree.SetFocused(ta.ElementID(), true)

	// Set selection manually
	ta.selAnchor = 0
	ta.cursorPos = 8
	ta.Draw(buf)
}

// --- Form ---

func TestForm(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	form := NewForm(tree, cfg)

	if form.Layout() != FormLayoutHorizontal {t.Error("expected horizontal")}
	form.SetLayout(FormLayoutVertical)
	form.SetLabelWidth(120)
	if form.LabelWidth() != 120 {t.Errorf("got %f", form.LabelWidth())}

	buf := render.NewCommandBuffer()
	form.Draw(buf)
}

func TestFormItem(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()

	inp := NewInput(tree, cfg)
	fi := NewFormItem(tree, "Name", inp, cfg)
	if fi.Label() != "Name" {t.Errorf("got '%s'", fi.Label())}
	if fi.Control() != inp {t.Error("control mismatch")}
	fi.SetLabel("Email")
	fi.SetRequired(true)
	if !fi.IsRequired() {t.Error("should be required")}
	fi.SetError("Required field")
	if fi.Error() != "Required field" {t.Errorf("got '%s'", fi.Error())}
}

func TestFormItemDraw(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	buf := render.NewCommandBuffer()

	inp := NewInput(tree, cfg)
	fi := NewFormItem(tree, "Name", inp, cfg)
	setBounds(tree, fi, 10, 10, 400, 32)
	fi.Draw(buf)

	// Required
	fi.SetRequired(true)
	fi.Draw(buf)

	// With error
	fi.SetError("This field is required")
	fi.Draw(buf)

	// No label
	fi2 := NewFormItem(tree, "", nil, cfg)
	setBounds(tree, fi2, 10, 50, 400, 32)
	fi2.Draw(buf)

	// No bounds
	fi3 := NewFormItem(tree, "X", nil, cfg)
	fi3.Draw(buf)
}

// === TextRenderer path coverage ===

func TestCheckboxDrawWithTextRenderer(t *testing.T) {
	tree := newTestTree()
	cfg := cfgWithTextRenderer()
	buf := render.NewCommandBuffer()
	cb := NewCheckbox(tree, "Checked", cfg)
	setBounds(tree, cb, 10, 10, 200, 32)
	cb.SetChecked(true)
	cb.Draw(buf)
	cb.SetChecked(false)
	cb.Draw(buf)
}

func TestRadioDrawWithTextRenderer(t *testing.T) {
	tree := newTestTree()
	cfg := cfgWithTextRenderer()
	buf := render.NewCommandBuffer()
	r := NewRadio(tree, "Opt", cfg)
	setBounds(tree, r, 10, 10, 200, 32)
	r.SetChecked(true)
	r.Draw(buf)
	r.SetChecked(false)
	r.Draw(buf)
}

func TestTagDrawWithTextRenderer(t *testing.T) {
	tree := newTestTree()
	cfg := cfgWithTextRenderer()
	buf := render.NewCommandBuffer()
	tag := NewTag(tree, "Info", cfg)
	setBounds(tree, tag, 10, 10, 80, 22)
	tag.Draw(buf)
}

func TestProgressDrawWithTextRenderer(t *testing.T) {
	tree := newTestTree()
	cfg := cfgWithTextRenderer()
	buf := render.NewCommandBuffer()
	p := NewProgress(tree, cfg)
	p.SetPercent(50)
	setBounds(tree, p, 10, 10, 200, 8)
	p.Draw(buf)
}

func TestLoadingDrawWithTextRenderer(t *testing.T) {
	tree := newTestTree()
	cfg := cfgWithTextRenderer()
	buf := render.NewCommandBuffer()
	l := NewLoading(tree, cfg)
	l.SetTip("Loading...")
	setBounds(tree, l, 10, 10, 200, 100)
	l.Draw(buf)
}

func TestEmptyDrawWithTextRenderer(t *testing.T) {
	tree := newTestTree()
	cfg := cfgWithTextRenderer()
	buf := render.NewCommandBuffer()
	e := NewEmpty(tree, cfg)
	setBounds(tree, e, 10, 10, 300, 120)
	e.Draw(buf)
}

func TestTabsDrawWithTextRenderer(t *testing.T) {
	tree := newTestTree()
	cfg := cfgWithTextRenderer()
	buf := render.NewCommandBuffer()
	items := []TabItem{
		{Key: "t1", Label: "First"},
		{Key: "t2", Label: "Second"},
	}
	tabs := NewTabs(tree, items, cfg)
	setBounds(tree, tabs, 10, 10, 400, 300)
	tabs.Draw(buf)
	tabs.SetActiveKey("t2")
	tabs.Draw(buf)
	tabs.Destroy()
}

func TestDialogDrawWithTextRenderer(t *testing.T) {
	tree := newTestTree()
	cfg := cfgWithTextRenderer()
	buf := render.NewCommandBuffer()
	d := NewDialog(tree, "Title", cfg)
	setBounds(tree, d, 0, 0, 800, 600)
	d.Open()
	d.Draw(buf)
}

func TestMessageDrawWithTextRenderer(t *testing.T) {
	tree := newTestTree()
	cfg := cfgWithTextRenderer()
	buf := render.NewCommandBuffer()
	types := []MessageType{MessageInfo, MessageSuccess, MessageWarning, MessageError}
	for _, mt := range types {
		m := NewMessage(tree, "Msg", cfg)
		m.SetMsgType(mt)
		setBounds(tree, m, 100, 10, 300, 36)
		m.Draw(buf)
	}
}

func TestSelectDrawWithTextRenderer(t *testing.T) {
	tree := newTestTree()
	cfg := cfgWithTextRenderer()
	buf := render.NewCommandBuffer()
	opts := []SelectOption{
		{Label: "Apple", Value: "apple"},
		{Label: "Banana", Value: "banana"},
	}
	s := NewSelect(tree, opts, cfg)
	s.SetValue("apple")
	setBounds(tree, s, 10, 10, 200, 32)
	s.Draw(buf)

	// Open dropdown
	s.open = true
	s.createOptionElements()
	s.Draw(buf)
	s.Destroy()
}

func TestFormItemDrawWithTextRenderer(t *testing.T) {
	tree := newTestTree()
	cfg := cfgWithTextRenderer()
	buf := render.NewCommandBuffer()
	inp := NewInput(tree, cfg)
	fi := NewFormItem(tree, "Name", inp, cfg)
	fi.SetRequired(true)
	fi.SetError("Required")
	setBounds(tree, fi, 10, 10, 400, 32)
	fi.Draw(buf)
}

func TestTextAreaDrawWithTextRenderer(t *testing.T) {
	tree := newTestTree()
	cfg := cfgWithTextRenderer()
	buf := render.NewCommandBuffer()

	ta := NewTextArea(tree, cfg)
	ta.SetValue("Hello\nWorld")
	setBounds(tree, ta, 10, 10, 300, 200)

	// With text
	ta.Draw(buf)

	// Focused with cursor
	tree.SetFocused(ta.ElementID(), true)
	ta.Draw(buf)

	// With selection
	ta.selAnchor = 0
	ta.cursorPos = 5
	ta.Draw(buf)

	// Placeholder
	ta.SetValue("")
	ta.SetPlaceholder("Type here...")
	ta.Draw(buf)
}

func TestTextAreaHitTestWithTextRenderer(t *testing.T) {
	tree := newTestTree()
	cfg := cfgWithTextRenderer()
	ta := NewTextArea(tree, cfg)
	ta.SetValue("Hello\nWorld")
	setBounds(tree, ta, 10, 10, 300, 200)

	// Hit test at various positions
	pos := ta.hitTestChar(10+8+30, 10+8+5) // first line
	if pos < 0 || pos > 5 {t.Errorf("unexpected pos %d", pos)}

	pos = ta.hitTestChar(10+8+30, 10+8+25) // second line
	if pos < 5 {t.Errorf("unexpected pos %d for second line", pos)}
}

func TestRadioClickWithoutGroup(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	r := NewRadio(tree, "Solo", cfg)
	var called bool
	r.OnChange(func(bool) { called = true })

	dispatcher := core.NewDispatcher(tree)
	dispatcher.Dispatch(r.ElementID(), &event.Event{Type: event.MouseClick})
	if !r.IsChecked() {t.Error("should be checked")}
	if !called {t.Error("onChange should be called")}
}

func TestSelectClickToggle(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	opts := []SelectOption{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b"},
	}
	s := NewSelect(tree, opts, cfg)
	setBounds(tree, s, 10, 10, 200, 32)

	dispatcher := core.NewDispatcher(tree)

	// Open
	dispatcher.Dispatch(s.ElementID(), &event.Event{Type: event.MouseClick})
	if !s.IsOpen() {t.Error("should be open")}

	// Close
	dispatcher.Dispatch(s.ElementID(), &event.Event{Type: event.MouseClick})
	if s.IsOpen() {t.Error("should be closed")}

	// Disabled: no toggle
	s.SetDisabled(true)
	dispatcher.Dispatch(s.ElementID(), &event.Event{Type: event.MouseClick})
	if s.IsOpen() {t.Error("should stay closed when disabled")}

	s.Destroy()
}

func TestTextAreaCopyPaste(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	ta := NewTextArea(tree, cfg)
	ta.SetValue("Hello World")
	setBounds(tree, ta, 10, 10, 300, 200)

	// Copy without platform: no panic
	ta.selAnchor = 0
	ta.cursorPos = 5
	ta.copySelection()
	ta.cutSelection()

	// Paste without platform: no panic
	ta.paste()

	// Show context menu without window: no panic
	ta.showContextMenu(0, 0)
}

func TestTextAreaDeleteForwardAtEnd(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	ta := NewTextArea(tree, cfg)
	ta.SetValue("AB")
	ta.cursorPos = 2
	ta.deleteForward() // at end, should be no-op
	if ta.Value() != "AB" {t.Errorf("got '%s'", ta.Value())}
}

func TestTextAreaBackspaceAtStart(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	ta := NewTextArea(tree, cfg)
	ta.SetValue("AB")
	ta.cursorPos = 0
	ta.deleteBack() // at start, should be no-op
	if ta.Value() != "AB" {t.Errorf("got '%s'", ta.Value())}
}

func TestTextAreaSelectedText(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	ta := NewTextArea(tree, cfg)
	ta.SetValue("Hello World")

	// No selection
	if ta.selectedText() != "" {t.Error("expected empty")}

	// With selection
	ta.selAnchor = 0
	ta.cursorPos = 5
	if ta.selectedText() != "Hello" {t.Errorf("got '%s'", ta.selectedText())}

	// Reverse selection
	ta.selAnchor = 5
	ta.cursorPos = 0
	if ta.selectedText() != "Hello" {t.Errorf("got '%s'", ta.selectedText())}
}

func TestTextAreaEmptyLines(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	ta := NewTextArea(tree, cfg)
	ta.SetValue("")
	setBounds(tree, ta, 0, 0, 400, 200)
	vl := ta.visualLines(ta.contentWidth())
	if len(vl) != 1 {t.Errorf("expected 1, got %d", len(vl))}
	if vl[0].text != "" {t.Errorf("expected empty string, got '%s'", vl[0].text)}
}

// ==================== Coverage boost tests ====================

// --- Mock Platform and Window ---

type mockPlatform struct {
	clipboard string
}

func (m *mockPlatform) Init() error                  { return nil }
func (m *mockPlatform) CreateWindow(_ platform.WindowOptions) (platform.Window, error) {
	return nil, nil
}
func (m *mockPlatform) PollEvents() []event.Event    { return nil }
func (m *mockPlatform) ProcessMessages()             {}
func (m *mockPlatform) GetClipboardText() string     { return m.clipboard }
func (m *mockPlatform) SetClipboardText(text string) { m.clipboard = text }
func (m *mockPlatform) GetPrimaryMonitorDPI() float32 { return 1.0 }
func (m *mockPlatform) GetSystemLocale() string       { return "en-US" }
func (m *mockPlatform) Terminate()                    {}

type mockWindow struct {
	imeRect     uimath.Rect
	menuResult  int
}

func (m *mockWindow) Size() (int, int)             { return 800, 600 }
func (m *mockWindow) SetSize(int, int)             {}
func (m *mockWindow) FramebufferSize() (int, int)  { return 800, 600 }
func (m *mockWindow) Position() (int, int)         { return 0, 0 }
func (m *mockWindow) SetPosition(int, int)         {}
func (m *mockWindow) SetTitle(string)              {}
func (m *mockWindow) SetFullscreen(bool)           {}
func (m *mockWindow) IsFullscreen() bool           { return false }
func (m *mockWindow) ShouldClose() bool            { return false }
func (m *mockWindow) SetShouldClose(bool)          {}
func (m *mockWindow) NativeHandle() uintptr        { return 0 }
func (m *mockWindow) DPIScale() float32            { return 1.0 }
func (m *mockWindow) SetVisible(bool)              {}
func (m *mockWindow) ShowDeferred()                {}
func (m *mockWindow) SetMinSize(int, int)          {}
func (m *mockWindow) SetMaxSize(int, int)          {}
func (m *mockWindow) SetCursor(platform.CursorShape) {}
func (m *mockWindow) SetIMEPosition(r uimath.Rect)  { m.imeRect = r }
func (m *mockWindow) ShowContextMenu(_, _ int, _ []platform.ContextMenuItem) int {
	return m.menuResult
}
func (m *mockWindow) ClientToScreen(x, y int) (int, int) { return x, y }
func (m *mockWindow) Destroy()                    {}

func cfgWithPlatform() (*Config, *mockPlatform, *mockWindow) {
	cfg := DefaultConfig()
	p := &mockPlatform{}
	w := &mockWindow{menuResult: -1}
	cfg.Platform = p
	cfg.Window = w
	return cfg, p, w
}

// --- Input: selection + clipboard + context menu ---

func TestInputSelectAll(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.SetValue("hello")

	handlers := tree.Handlers(inp.ElementID(), event.KeyDown)
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyA, Modifiers: event.Modifiers{Ctrl: true}})
	}
	s, e := inp.Selection()
	if s != 0 || e != 5 {
		t.Errorf("expected selection 0-5, got %d-%d", s, e)
	}
}

func TestInputCopyPaste(t *testing.T) {
	tree := newTestTree()
	cfg, plat, _ := cfgWithPlatform()
	inp := NewInput(tree, cfg)
	inp.SetValue("hello world")

	// Select "hello"
	inp.selAnchor = 0
	inp.cursorPos = 5

	// Ctrl+C
	handlers := tree.Handlers(inp.ElementID(), event.KeyDown)
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyC, Modifiers: event.Modifiers{Ctrl: true}})
	}
	if plat.clipboard != "hello" {
		t.Errorf("expected clipboard 'hello', got '%s'", plat.clipboard)
	}

	// Move cursor to end, paste
	inp.selAnchor = -1
	inp.cursorPos = 11
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyV, Modifiers: event.Modifiers{Ctrl: true}})
	}
	if inp.Value() != "hello worldhello" {
		t.Errorf("expected 'hello worldhello', got '%s'", inp.Value())
	}
}

func TestInputCutSelection(t *testing.T) {
	tree := newTestTree()
	cfg, plat, _ := cfgWithPlatform()
	inp := NewInput(tree, cfg)
	inp.SetValue("abcdef")

	inp.selAnchor = 1
	inp.cursorPos = 4

	handlers := tree.Handlers(inp.ElementID(), event.KeyDown)
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyX, Modifiers: event.Modifiers{Ctrl: true}})
	}
	if plat.clipboard != "bcd" {
		t.Errorf("expected 'bcd', got '%s'", plat.clipboard)
	}
	if inp.Value() != "aef" {
		t.Errorf("expected 'aef', got '%s'", inp.Value())
	}
}

func TestInputPasteFiltersNewlines(t *testing.T) {
	tree := newTestTree()
	cfg, plat, _ := cfgWithPlatform()
	inp := NewInput(tree, cfg)
	inp.SetValue("")
	plat.clipboard = "line1\nline2\rline3"

	inp.paste()
	if inp.Value() != "line1line2line3" {
		t.Errorf("expected newlines filtered, got '%s'", inp.Value())
	}
}

func TestInputPasteEmptyClipboard(t *testing.T) {
	tree := newTestTree()
	cfg, plat, _ := cfgWithPlatform()
	inp := NewInput(tree, cfg)
	inp.SetValue("test")
	plat.clipboard = ""
	inp.paste()
	if inp.Value() != "test" {
		t.Error("empty clipboard paste should be no-op")
	}
}

func TestInputContextMenu(t *testing.T) {
	tree := newTestTree()
	cfg, plat, win := cfgWithPlatform()
	inp := NewInput(tree, cfg)
	inp.SetValue("hello world")

	// Test cut via context menu
	inp.selAnchor = 0
	inp.cursorPos = 5
	win.menuResult = 0 // Cut
	inp.showContextMenu(10, 10)
	if plat.clipboard != "hello" {
		t.Errorf("context menu cut: expected 'hello', got '%s'", plat.clipboard)
	}
	if inp.Value() != " world" {
		t.Errorf("expected ' world', got '%s'", inp.Value())
	}

	// Test copy via context menu
	inp.SetValue("abcdef")
	inp.selAnchor = 0
	inp.cursorPos = 3
	win.menuResult = 1 // Copy
	inp.showContextMenu(10, 10)
	if plat.clipboard != "abc" {
		t.Errorf("context menu copy: expected 'abc', got '%s'", plat.clipboard)
	}

	// Test paste via context menu
	plat.clipboard = "XY"
	inp.selAnchor = -1
	inp.cursorPos = 0
	win.menuResult = 2 // Paste
	inp.showContextMenu(10, 10)
	if inp.Value() != "XYabcdef" {
		t.Errorf("expected 'XYabcdef', got '%s'", inp.Value())
	}

	// Test select all via context menu
	win.menuResult = 3 // Select All
	inp.showContextMenu(10, 10)
	s, e := inp.Selection()
	if s != 0 || e != 8 {
		t.Errorf("expected selection 0-8, got %d-%d", s, e)
	}
}

func TestInputContextMenuNoWindow(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.showContextMenu(0, 0) // should not panic
}

func TestInputSelectionHelpers(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.SetValue("abcde")

	// selMin/selMax via Selection()
	inp.selAnchor = 3
	inp.cursorPos = 1
	s, e := inp.Selection()
	if s != 1 || e != 3 {
		t.Errorf("expected 1-3, got %d-%d", s, e)
	}

	// selectedText
	got := inp.selectedText()
	if got != "bc" {
		t.Errorf("expected 'bc', got '%s'", got)
	}

	// deleteSelection
	inp.deleteSelection()
	if inp.Value() != "ade" {
		t.Errorf("expected 'ade', got '%s'", inp.Value())
	}
}

func TestInputShiftArrowSelection(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.SetValue("abcde")
	inp.cursorPos = 2

	handlers := tree.Handlers(inp.ElementID(), event.KeyDown)

	// Shift+Right starts selection
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyArrowRight, Modifiers: event.Modifiers{Shift: true}})
	}
	if inp.selAnchor != 2 || inp.cursorPos != 3 {
		t.Errorf("expected anchor=2, cursor=3, got anchor=%d, cursor=%d", inp.selAnchor, inp.cursorPos)
	}

	// Shift+Left
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyArrowLeft, Modifiers: event.Modifiers{Shift: true}})
	}
	if inp.cursorPos != 2 {
		t.Errorf("expected cursor=2, got %d", inp.cursorPos)
	}

	// ArrowRight with selection collapses to selMax
	inp.selAnchor = 1
	inp.cursorPos = 4
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyArrowRight})
	}
	if inp.cursorPos != 4 || inp.selAnchor != -1 {
		t.Error("right arrow should collapse selection to max")
	}

	// ArrowLeft with selection collapses to selMin
	inp.selAnchor = 1
	inp.cursorPos = 4
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyArrowLeft})
	}
	if inp.cursorPos != 1 || inp.selAnchor != -1 {
		t.Error("left arrow should collapse selection to min")
	}

	// Shift+Home
	inp.cursorPos = 3
	inp.selAnchor = -1
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyHome, Modifiers: event.Modifiers{Shift: true}})
	}
	if inp.cursorPos != 0 || inp.selAnchor != 3 {
		t.Error("Shift+Home should select to start")
	}

	// Shift+End
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyEnd, Modifiers: event.Modifiers{Shift: true}})
	}
	if inp.cursorPos != 5 {
		t.Error("Shift+End should move cursor to end")
	}
}

func TestInputBackspaceWithSelection(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.SetValue("abcde")
	inp.selAnchor = 1
	inp.cursorPos = 3

	handlers := tree.Handlers(inp.ElementID(), event.KeyDown)
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyBackspace})
	}
	if inp.Value() != "ade" {
		t.Errorf("expected 'ade', got '%s'", inp.Value())
	}
}

func TestInputDeleteWithSelection(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.SetValue("abcde")
	inp.selAnchor = 2
	inp.cursorPos = 4

	handlers := tree.Handlers(inp.ElementID(), event.KeyDown)
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyDelete})
	}
	if inp.Value() != "abe" {
		t.Errorf("expected 'abe', got '%s'", inp.Value())
	}
}

func TestInputHitTestChar(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.SetValue("abc")
	setBounds(tree, inp, 10, 0, 200, 32)

	// Hit test at start
	pos := inp.hitTestChar(10) // exactly at bounds.X + padLeft (relX=0)
	if pos != 0 {
		t.Errorf("expected 0, got %d", pos)
	}

	// Hit test far right
	pos = inp.hitTestChar(200)
	if pos != 3 {
		t.Errorf("expected 3, got %d", pos)
	}

	// Hit test negative
	pos = inp.hitTestChar(0)
	if pos != 0 {
		t.Errorf("expected 0, got %d", pos)
	}
}

func TestInputHitTestCharWithTextRenderer(t *testing.T) {
	tree := newTestTree()
	cfg := cfgWithTextRenderer()
	inp := NewInput(tree, cfg)
	inp.SetValue("abc")
	setBounds(tree, inp, 10, 0, 200, 32)

	// Far right should return len
	pos := inp.hitTestChar(200)
	if pos != 3 {
		t.Errorf("expected 3, got %d", pos)
	}

	// At start
	pos = inp.hitTestChar(10)
	if pos != 0 {
		t.Errorf("expected 0, got %d", pos)
	}
}

func TestInputMeasureRunes(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	w := inp.measureRunes([]rune("abc"))
	if w <= 0 {
		t.Error("measure should return positive width")
	}

	// With TextRenderer
	cfg := cfgWithTextRenderer()
	inp2 := NewInput(tree, cfg)
	w2 := inp2.measureRunes([]rune("abc"))
	if w2 <= 0 {
		t.Error("measure with TextRenderer should return positive width")
	}
}

func TestInputCursorXAndTextX(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.SetValue("abc")
	inp.cursorPos = 2

	cx := inp.cursorX()
	if cx <= 0 {
		t.Error("cursorX should be positive")
	}

	tx := inp.textX(1)
	if tx <= 0 {
		t.Error("textX should be positive for pos=1")
	}

	// textX beyond length
	tx2 := inp.textX(100)
	if tx2 <= 0 {
		t.Error("textX should clamp and return positive")
	}
}

func TestInputUpdateIMEPosition(t *testing.T) {
	tree := newTestTree()
	cfg, _, win := cfgWithPlatform()
	inp := NewInput(tree, cfg)
	inp.SetValue("abc")
	inp.cursorPos = 1
	setBounds(tree, inp, 10, 20, 200, 32)

	inp.updateIMEPosition()
	if win.imeRect.Width != 1 {
		t.Errorf("expected IME rect width 1, got %f", win.imeRect.Width)
	}
}

func TestInputUpdateIMEWithTextRenderer(t *testing.T) {
	tree := newTestTree()
	cfg := cfgWithTextRenderer()
	win := &mockWindow{}
	cfg.Window = win
	inp := NewInput(tree, cfg)
	inp.SetValue("abc")
	inp.cursorPos = 2
	setBounds(tree, inp, 10, 20, 200, 32)

	inp.updateIMEPosition()
	if win.imeRect.Width != 1 {
		t.Errorf("expected IME rect width 1, got %f", win.imeRect.Width)
	}
}

func TestInputDrawWithSelection(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.SetValue("hello")
	inp.selAnchor = 1
	inp.cursorPos = 4
	tree.SetFocused(inp.ElementID(), true)
	setBounds(tree, inp, 0, 0, 200, 32)

	buf := render.NewCommandBuffer()
	inp.Draw(buf)
	// rect + selection + text = 3 (maybe +1 for cursor depending on time)
	if buf.Len() < 3 {
		t.Errorf("expected at least 3 commands, got %d", buf.Len())
	}
}

func TestInputDrawWithTextRenderer(t *testing.T) {
	tree := newTestTree()
	cfg := cfgWithTextRenderer()
	inp := NewInput(tree, cfg)
	inp.SetValue("hello")
	setBounds(tree, inp, 0, 0, 200, 32)

	buf := render.NewCommandBuffer()
	inp.Draw(buf)
	// rect + text (via TextRenderer)
	if buf.Len() < 2 {
		t.Errorf("expected at least 2, got %d", buf.Len())
	}
}

func TestInputDrawFocusedWithTextRenderer(t *testing.T) {
	tree := newTestTree()
	cfg := cfgWithTextRenderer()
	inp := NewInput(tree, cfg)
	inp.SetValue("hi")
	inp.selAnchor = 0
	inp.cursorPos = 2
	tree.SetFocused(inp.ElementID(), true)
	setBounds(tree, inp, 0, 0, 200, 32)

	buf := render.NewCommandBuffer()
	inp.Draw(buf)
	if buf.Len() < 2 {
		t.Errorf("expected at least 2, got %d", buf.Len())
	}
}

func TestInputMouseDrag(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.SetValue("abcde")
	setBounds(tree, inp, 0, 0, 200, 32)

	// MouseDown
	downHandlers := tree.Handlers(inp.ElementID(), event.MouseDown)
	for _, h := range downHandlers {
		h(&event.Event{Type: event.MouseDown, Button: event.MouseButtonLeft, X: 10})
	}
	if !inp.dragging {
		t.Error("should be dragging after mouse down")
	}

	// MouseMove
	moveHandlers := tree.Handlers(inp.ElementID(), event.MouseMove)
	for _, h := range moveHandlers {
		h(&event.Event{Type: event.MouseMove, X: 100})
	}

	// MouseUp - should produce a selection
	upHandlers := tree.Handlers(inp.ElementID(), event.MouseUp)
	for _, h := range upHandlers {
		h(&event.Event{Type: event.MouseUp})
	}
	if inp.dragging {
		t.Error("should not be dragging after mouse up")
	}

	// Click after drag with dragSelected=true should not clear selection
	if inp.dragSelected {
		clickHandlers := tree.Handlers(inp.ElementID(), event.MouseClick)
		for _, h := range clickHandlers {
			h(&event.Event{Type: event.MouseClick, X: 50})
		}
		// dragSelected should be cleared
		if inp.dragSelected {
			t.Error("dragSelected should be cleared after click")
		}
	}
}

func TestInputMouseDragNoDragOnDisabled(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.SetValue("abcde")
	inp.SetDisabled(true)
	setBounds(tree, inp, 0, 0, 200, 32)

	downHandlers := tree.Handlers(inp.ElementID(), event.MouseDown)
	for _, h := range downHandlers {
		h(&event.Event{Type: event.MouseDown, Button: event.MouseButtonLeft, X: 10})
	}
	if inp.dragging {
		t.Error("disabled input should not start dragging")
	}
}

func TestInputMouseUpWithoutSelection(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.SetValue("abc")
	setBounds(tree, inp, 0, 0, 200, 32)

	// MouseDown at pos
	downHandlers := tree.Handlers(inp.ElementID(), event.MouseDown)
	for _, h := range downHandlers {
		h(&event.Event{Type: event.MouseDown, Button: event.MouseButtonLeft, X: 10})
	}
	// MouseUp at same pos - no selection
	upHandlers := tree.Handlers(inp.ElementID(), event.MouseUp)
	for _, h := range upHandlers {
		h(&event.Event{Type: event.MouseUp})
	}
	if inp.selAnchor != -1 {
		t.Error("same pos up should clear selection")
	}
}

func TestInputRightClickContextMenu(t *testing.T) {
	tree := newTestTree()
	cfg, _, win := cfgWithPlatform()
	inp := NewInput(tree, cfg)
	inp.SetValue("abc")
	win.menuResult = -1 // dismissed
	setBounds(tree, inp, 0, 0, 200, 32)

	downHandlers := tree.Handlers(inp.ElementID(), event.MouseDown)
	for _, h := range downHandlers {
		h(&event.Event{Type: event.MouseDown, Button: event.MouseButtonRight, GlobalX: 50, GlobalY: 20})
	}
	if !inp.Element().IsFocused() {
		t.Error("right click should focus input")
	}
}

func TestInputShiftClickSelection(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.SetValue("abcde")
	setBounds(tree, inp, 0, 0, 200, 32)

	// First click to set cursor
	clickHandlers := tree.Handlers(inp.ElementID(), event.MouseClick)
	for _, h := range clickHandlers {
		h(&event.Event{Type: event.MouseClick, X: 10})
	}

	// Shift+MouseDown to start shift-selection
	downHandlers := tree.Handlers(inp.ElementID(), event.MouseDown)
	for _, h := range downHandlers {
		h(&event.Event{Type: event.MouseDown, Button: event.MouseButtonLeft, X: 100, Modifiers: event.Modifiers{Shift: true}})
	}
	if inp.selAnchor < 0 {
		t.Error("shift+mousedown should set selection anchor")
	}
}

func TestInputCtrlKeyPressIgnored(t *testing.T) {
	tree := newTestTree()
	inp := NewInput(tree, nil)
	inp.SetValue("")

	handlers := tree.Handlers(inp.ElementID(), event.KeyPress)
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyPress, Char: 'a', Modifiers: event.Modifiers{Ctrl: true}})
	}
	if inp.Value() != "" {
		t.Error("Ctrl+char in KeyPress should be ignored")
	}
}

func TestInputCopyNoSelection(t *testing.T) {
	tree := newTestTree()
	cfg, plat, _ := cfgWithPlatform()
	inp := NewInput(tree, cfg)
	inp.SetValue("abc")
	plat.clipboard = "old"
	inp.selAnchor = -1
	inp.copySelection()
	if plat.clipboard != "old" {
		t.Error("copy with no selection should not change clipboard")
	}
}

// --- TextArea: coverage boost ---

func TestTextAreaDeleteForwardMidText(t *testing.T) {
	tree := newTestTree()
	ta := NewTextArea(tree, nil)
	ta.SetValue("abcde")
	ta.cursorPos = 2
	var changed string
	ta.OnChange(func(v string) { changed = v })

	ta.deleteForward()
	if ta.Value() != "abde" {
		t.Errorf("expected 'abde', got '%s'", ta.Value())
	}
	if changed != "abde" {
		t.Error("onChange should fire")
	}
}

func TestTextAreaPasteWithPlatform(t *testing.T) {
	tree := newTestTree()
	cfg, plat, _ := cfgWithPlatform()
	ta := NewTextArea(tree, cfg)
	ta.SetValue("hello")
	ta.cursorPos = 5
	plat.clipboard = " world"

	ta.paste()
	if ta.Value() != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", ta.Value())
	}
	if ta.cursorPos != 11 {
		t.Errorf("expected cursor at 11, got %d", ta.cursorPos)
	}
}

func TestTextAreaPasteEmptyClipboard(t *testing.T) {
	tree := newTestTree()
	cfg, plat, _ := cfgWithPlatform()
	ta := NewTextArea(tree, cfg)
	ta.SetValue("abc")
	plat.clipboard = ""
	ta.paste()
	if ta.Value() != "abc" {
		t.Error("empty clipboard paste should be no-op")
	}
}

func TestTextAreaPasteWithSelection(t *testing.T) {
	tree := newTestTree()
	cfg, plat, _ := cfgWithPlatform()
	ta := NewTextArea(tree, cfg)
	ta.SetValue("hello world")
	ta.selAnchor = 5
	ta.cursorPos = 11
	plat.clipboard = "!"

	ta.paste()
	if ta.Value() != "hello!" {
		t.Errorf("expected 'hello!', got '%s'", ta.Value())
	}
}

func TestTextAreaCopyPasteCut(t *testing.T) {
	tree := newTestTree()
	cfg, plat, _ := cfgWithPlatform()
	ta := NewTextArea(tree, cfg)
	ta.SetValue("hello world")

	// Copy
	ta.selAnchor = 0
	ta.cursorPos = 5
	ta.copySelection()
	if plat.clipboard != "hello" {
		t.Errorf("expected 'hello', got '%s'", plat.clipboard)
	}

	// Cut
	ta.selAnchor = 0
	ta.cursorPos = 6
	ta.cutSelection()
	if ta.Value() != "world" {
		t.Errorf("expected 'world', got '%s'", ta.Value())
	}
}

func TestTextAreaContextMenu(t *testing.T) {
	tree := newTestTree()
	cfg, plat, win := cfgWithPlatform()
	ta := NewTextArea(tree, cfg)
	ta.SetValue("abcdef")

	// Cut via menu
	ta.selAnchor = 0
	ta.cursorPos = 3
	win.menuResult = 0
	ta.showContextMenu(10, 10)
	if plat.clipboard != "abc" {
		t.Errorf("expected 'abc', got '%s'", plat.clipboard)
	}

	// Copy via menu
	ta.SetValue("xyz123")
	ta.selAnchor = 3
	ta.cursorPos = 6
	win.menuResult = 1
	ta.showContextMenu(10, 10)
	if plat.clipboard != "123" {
		t.Errorf("expected '123', got '%s'", plat.clipboard)
	}

	// Paste via menu
	plat.clipboard = "AB"
	ta.selAnchor = -1
	ta.cursorPos = 0
	win.menuResult = 2
	ta.showContextMenu(10, 10)
	if ta.Value() != "ABxyz123" {
		t.Errorf("expected 'ABxyz123', got '%s'", ta.Value())
	}

	// Select all via menu
	win.menuResult = 3
	ta.showContextMenu(10, 10)
	if ta.selAnchor != 0 || ta.cursorPos != ta.runeLen() {
		t.Error("select all should select everything")
	}
}

func TestTextAreaContextMenuDismissed(t *testing.T) {
	tree := newTestTree()
	cfg, _, win := cfgWithPlatform()
	ta := NewTextArea(tree, cfg)
	ta.SetValue("abc")
	win.menuResult = -1
	ta.showContextMenu(10, 10) // should not panic or change state
	if ta.Value() != "abc" {
		t.Error("dismissed menu should not change value")
	}
}

func TestTextAreaDrawPlaceholder(t *testing.T) {
	tree := newTestTree()
	ta := NewTextArea(tree, nil)
	ta.SetPlaceholder("Enter text...")
	setBounds(tree, ta, 0, 0, 200, 200)

	buf := render.NewCommandBuffer()
	ta.Draw(buf)
	// rect + placeholder text
	if buf.Len() < 2 {
		t.Errorf("expected at least 2, got %d", buf.Len())
	}
}

func TestTextAreaDrawFocusedWithSelection(t *testing.T) {
	tree := newTestTree()
	ta := NewTextArea(tree, nil)
	ta.SetValue("hello\nworld")
	ta.selAnchor = 2
	ta.cursorPos = 8 // spans two lines
	tree.SetFocused(ta.ElementID(), true)
	setBounds(tree, ta, 0, 0, 200, 200)

	buf := render.NewCommandBuffer()
	ta.Draw(buf)
	if buf.Len() < 3 {
		t.Errorf("expected at least 3 commands, got %d", buf.Len())
	}
}

func TestTextAreaDrawDisabled(t *testing.T) {
	tree := newTestTree()
	ta := NewTextArea(tree, nil)
	ta.SetValue("text")
	ta.SetDisabled(true)
	setBounds(tree, ta, 0, 0, 200, 200)

	buf := render.NewCommandBuffer()
	ta.Draw(buf)
	if buf.Len() < 1 {
		t.Error("disabled textarea should still draw")
	}
}

func TestTextAreaLineEnd2(t *testing.T) {
	tree := newTestTree()
	ta := NewTextArea(tree, nil)
	ta.SetValue("abc\nde")

	// lineEnd on first line
	end := ta.lineEnd(1)
	if end != 3 {
		t.Errorf("expected 3, got %d", end)
	}

	// lineEnd on second line
	end = ta.lineEnd(5)
	if end != 6 {
		t.Errorf("expected 6, got %d", end)
	}
}

func TestTextAreaKeyboardNavigation(t *testing.T) {
	tree := newTestTree()
	ta := NewTextArea(tree, nil)
	ta.SetValue("abc\ndef\nghi")
	ta.cursorPos = 5 // 'e' in line 2

	handlers := tree.Handlers(ta.ElementID(), event.KeyDown)

	// Shift+ArrowDown
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyArrowDown, Modifiers: event.Modifiers{Shift: true}})
	}
	if ta.selAnchor != 5 {
		t.Error("Shift+Down should set selection anchor")
	}

	// ArrowUp without shift clears selection
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyArrowUp})
	}
	if ta.selAnchor != -1 {
		t.Error("Up without shift should clear selection")
	}

	// Home
	ta.cursorPos = 5
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyHome})
	}
	if ta.cursorPos != 4 {
		t.Errorf("Home should go to line start, got %d", ta.cursorPos)
	}

	// End
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyEnd})
	}
	if ta.cursorPos != 7 {
		t.Errorf("End should go to line end, got %d", ta.cursorPos)
	}
}

func TestTextAreaIMEDisabledOrEmpty(t *testing.T) {
	tree := newTestTree()
	ta := NewTextArea(tree, nil)

	// Disabled
	ta.SetDisabled(true)
	handlers := tree.Handlers(ta.ElementID(), event.IMECompositionEnd)
	for _, h := range handlers {
		h(&event.Event{Type: event.IMECompositionEnd, Text: "test"})
	}
	if ta.Value() != "" {
		t.Error("disabled should not accept IME")
	}

	// Empty text
	ta.SetDisabled(false)
	for _, h := range handlers {
		h(&event.Event{Type: event.IMECompositionEnd, Text: ""})
	}
	if ta.Value() != "" {
		t.Error("empty IME text should be no-op")
	}
}

func TestTextAreaMouseDrag(t *testing.T) {
	tree := newTestTree()
	ta := NewTextArea(tree, nil)
	ta.SetValue("hello world")
	setBounds(tree, ta, 0, 0, 200, 200)

	// MouseDown
	downHandlers := tree.Handlers(ta.ElementID(), event.MouseDown)
	for _, h := range downHandlers {
		h(&event.Event{Type: event.MouseDown, Button: event.MouseButtonLeft, X: 10, Y: 10})
	}
	if !ta.dragging {
		t.Error("should be dragging")
	}

	// MouseMove
	moveHandlers := tree.Handlers(ta.ElementID(), event.MouseMove)
	for _, h := range moveHandlers {
		h(&event.Event{Type: event.MouseMove, X: 100, Y: 10})
	}

	// MouseUp
	upHandlers := tree.Handlers(ta.ElementID(), event.MouseUp)
	for _, h := range upHandlers {
		h(&event.Event{Type: event.MouseUp})
	}
	if ta.dragging {
		t.Error("should not be dragging after up")
	}
}

func TestTextAreaMouseDownDisabled(t *testing.T) {
	tree := newTestTree()
	ta := NewTextArea(tree, nil)
	ta.SetValue("abc")
	ta.SetDisabled(true)
	setBounds(tree, ta, 0, 0, 200, 200)

	downHandlers := tree.Handlers(ta.ElementID(), event.MouseDown)
	for _, h := range downHandlers {
		h(&event.Event{Type: event.MouseDown, Button: event.MouseButtonLeft, X: 10, Y: 10})
	}
	if ta.dragging {
		t.Error("disabled should not drag")
	}
}

func TestTextAreaRightClickContextMenu(t *testing.T) {
	tree := newTestTree()
	cfg, _, win := cfgWithPlatform()
	ta := NewTextArea(tree, cfg)
	ta.SetValue("abc")
	win.menuResult = -1
	setBounds(tree, ta, 0, 0, 200, 200)

	downHandlers := tree.Handlers(ta.ElementID(), event.MouseDown)
	for _, h := range downHandlers {
		h(&event.Event{Type: event.MouseDown, Button: event.MouseButtonRight, GlobalX: 50, GlobalY: 50})
	}
}

func TestTextAreaShiftClick(t *testing.T) {
	tree := newTestTree()
	ta := NewTextArea(tree, nil)
	ta.SetValue("hello world")
	ta.cursorPos = 3
	ta.selAnchor = 1
	setBounds(tree, ta, 0, 0, 200, 200)

	clickHandlers := tree.Handlers(ta.ElementID(), event.MouseClick)
	for _, h := range clickHandlers {
		h(&event.Event{Type: event.MouseClick, X: 100, Y: 10, Modifiers: event.Modifiers{Shift: true}})
	}
	// selAnchor should remain, cursorPos should change
	if ta.selAnchor != 1 {
		t.Error("shift click should preserve anchor")
	}
}

func TestTextAreaEnter(t *testing.T) {
	tree := newTestTree()
	ta := NewTextArea(tree, nil)
	ta.SetValue("ab")
	ta.cursorPos = 1

	handlers := tree.Handlers(ta.ElementID(), event.KeyDown)
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyEnter})
	}
	if ta.Value() != "a\nb" {
		t.Errorf("expected 'a\\nb', got '%s'", ta.Value())
	}
}

func TestTextAreaCtrlShortcuts(t *testing.T) {
	tree := newTestTree()
	cfg, plat, _ := cfgWithPlatform()
	ta := NewTextArea(tree, cfg)
	ta.SetValue("hello world")

	handlers := tree.Handlers(ta.ElementID(), event.KeyDown)

	// Ctrl+A
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyA, Modifiers: event.Modifiers{Ctrl: true}})
	}
	if ta.selAnchor != 0 || ta.cursorPos != 11 {
		t.Error("Ctrl+A should select all")
	}

	// Ctrl+C
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyC, Modifiers: event.Modifiers{Ctrl: true}})
	}
	if plat.clipboard != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", plat.clipboard)
	}

	// Ctrl+V
	plat.clipboard = "!"
	ta.selAnchor = -1
	ta.cursorPos = 11
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyV, Modifiers: event.Modifiers{Ctrl: true}})
	}
	if ta.Value() != "hello world!" {
		t.Errorf("expected 'hello world!', got '%s'", ta.Value())
	}

	// Ctrl+X
	ta.selAnchor = 0
	ta.cursorPos = 5
	for _, h := range handlers {
		h(&event.Event{Type: event.KeyDown, Key: event.KeyX, Modifiers: event.Modifiers{Ctrl: true}})
	}
	if plat.clipboard != "hello" {
		t.Errorf("expected 'hello', got '%s'", plat.clipboard)
	}
}

// --- Button: coverage boost ---

func TestButtonBgColorSecondary(t *testing.T) {
	tree := newTestTree()
	btn := NewButton(tree, "X", nil)
	btn.SetVariant(ButtonSecondary)

	// Normal
	bg := btn.bgColor()
	if bg != btn.config.BgColor {
		t.Error("secondary normal should use BgColor")
	}

	// Hovered
	tree.SetHovered(btn.ElementID(), true)
	bg = btn.bgColor()
	if bg == btn.config.BgColor {
		t.Error("secondary hovered should differ from normal")
	}

	// Pressed
	btn.pressed = true
	bg = btn.bgColor()
	_ = bg // just ensure no panic
}

func TestButtonBgColorTextLink(t *testing.T) {
	tree := newTestTree()
	for _, v := range []ButtonVariant{ButtonText, ButtonLink} {
		btn := NewButton(tree, "X", nil)
		btn.SetVariant(v)

		bg := btn.bgColor()
		if !bg.IsTransparent() {
			t.Errorf("variant %d normal should be transparent", v)
		}

		tree.SetHovered(btn.ElementID(), true)
		bg = btn.bgColor()
		_ = bg

		btn.pressed = true
		bg = btn.bgColor()
		_ = bg
	}
}

func TestButtonTextColorSecondary(t *testing.T) {
	tree := newTestTree()
	btn := NewButton(tree, "X", nil)
	btn.SetVariant(ButtonSecondary)

	tc := btn.textColor()
	if tc != btn.config.TextColor {
		t.Error("secondary normal text should be TextColor")
	}

	tree.SetHovered(btn.ElementID(), true)
	tc = btn.textColor()
	if tc != btn.config.HoverColor {
		t.Error("secondary hovered text should be HoverColor")
	}

	btn.pressed = true
	tc = btn.textColor()
	if tc != btn.config.ActiveColor {
		t.Error("secondary pressed text should be ActiveColor")
	}
}

func TestButtonTextColorTextLink(t *testing.T) {
	tree := newTestTree()
	for _, v := range []ButtonVariant{ButtonText, ButtonLink} {
		btn := NewButton(tree, "X", nil)
		btn.SetVariant(v)

		tc := btn.textColor()
		if tc != btn.config.PrimaryColor {
			t.Errorf("variant %d normal text should be PrimaryColor", v)
		}

		tree.SetHovered(btn.ElementID(), true)
		tc = btn.textColor()
		if tc != btn.config.HoverColor {
			t.Errorf("variant %d hovered text should be HoverColor", v)
		}

		btn.pressed = true
		tc = btn.textColor()
		if tc != btn.config.ActiveColor {
			t.Errorf("variant %d pressed text should be ActiveColor", v)
		}
	}
}

func TestButtonDrawWithTextRenderer(t *testing.T) {
	tree := newTestTree()
	cfg := cfgWithTextRenderer()
	btn := NewButton(tree, "OK", cfg)
	setBounds(tree, btn, 0, 0, 100, 32)

	buf := render.NewCommandBuffer()
	btn.Draw(buf)
	if buf.Len() < 2 {
		t.Errorf("expected at least 2 commands, got %d", buf.Len())
	}
}

func TestButtonDrawSecondaryHovered(t *testing.T) {
	tree := newTestTree()
	btn := NewButton(tree, "X", nil)
	btn.SetVariant(ButtonSecondary)
	tree.SetHovered(btn.ElementID(), true)
	setBounds(tree, btn, 0, 0, 100, 32)

	buf := render.NewCommandBuffer()
	btn.Draw(buf)
	if buf.Len() < 2 {
		t.Errorf("expected at least 2, got %d", buf.Len())
	}
}

func TestButtonDrawSecondaryPressed(t *testing.T) {
	tree := newTestTree()
	btn := NewButton(tree, "X", nil)
	btn.SetVariant(ButtonSecondary)
	btn.pressed = true
	setBounds(tree, btn, 0, 0, 100, 32)

	buf := render.NewCommandBuffer()
	btn.Draw(buf)
	if buf.Len() < 2 {
		t.Errorf("expected at least 2, got %d", buf.Len())
	}
}

// --- Text: coverage boost ---

func TestTextDrawWithTextRenderer(t *testing.T) {
	tree := newTestTree()
	cfg := cfgWithTextRenderer()
	txt := NewText(tree, "hello", cfg)
	setBounds(tree, txt, 0, 0, 200, 30)

	buf := render.NewCommandBuffer()
	txt.Draw(buf)
	if buf.Len() != 1 {
		t.Errorf("expected 1 command from TextRenderer, got %d", buf.Len())
	}
}

func TestTextDrawLongText(t *testing.T) {
	tree := newTestTree()
	txt := NewText(tree, "this is a very long text that might exceed bounds width easily", nil)
	setBounds(tree, txt, 0, 0, 50, 30)

	buf := render.NewCommandBuffer()
	txt.Draw(buf)
	if buf.Len() != 1 {
		t.Errorf("expected 1, got %d", buf.Len())
	}
}

func TestTextDrawTallText(t *testing.T) {
	tree := newTestTree()
	txt := NewText(tree, "X", nil)
	txt.SetFontSize(100) // very large font
	setBounds(tree, txt, 0, 0, 200, 10) // small height

	buf := render.NewCommandBuffer()
	txt.Draw(buf)
	if buf.Len() != 1 {
		t.Errorf("expected 1, got %d", buf.Len())
	}
}

// --- Tooltip: coverage boost ---

func TestTooltipDrawWithTextRenderer(t *testing.T) {
	tree := newTestTree()
	anchor := tree.CreateElement(core.TypeButton)
	cfg := cfgWithTextRenderer()
	tt := NewTooltip(tree, "Hint", anchor, cfg)
	tt.Show()
	setBounds(tree, tt, 100, 30, 60, 24)

	buf := render.NewCommandBuffer()
	tt.Draw(buf)
	// rect + text via TextRenderer
	if buf.Len() != 2 {
		t.Errorf("expected 2 commands, got %d", buf.Len())
	}
}

func TestTooltipDrawLongText(t *testing.T) {
	tree := newTestTree()
	anchor := tree.CreateElement(core.TypeButton)
	tt := NewTooltip(tree, "This is a very long tooltip text", anchor, nil)
	tt.Show()
	setBounds(tree, tt, 0, 0, 30, 24) // narrow

	buf := render.NewCommandBuffer()
	tt.Draw(buf)
	if buf.Len() != 2 {
		t.Errorf("expected 2, got %d", buf.Len())
	}
}

// --- Dialog: event handler coverage ---

func TestDialogCloseHandler(t *testing.T) {
	tree := newTestTree()
	d := NewDialog(tree, "Test", nil)
	closed := false
	d.OnClose(func() { closed = true })
	d.Open()

	// Trigger click handler
	handlers := tree.Handlers(d.ElementID(), event.MouseClick)
	for _, h := range handlers {
		h(&event.Event{Type: event.MouseClick})
	}
	if !closed {
		t.Error("click should trigger onClose")
	}
}

func TestDialogCloseHandlerNil(t *testing.T) {
	tree := newTestTree()
	d := NewDialog(tree, "Test", nil)
	d.Open()

	// Click without onClose set
	handlers := tree.Handlers(d.ElementID(), event.MouseClick)
	for _, h := range handlers {
		h(&event.Event{Type: event.MouseClick})
	}
	// should not panic
}

func TestDialogDrawWithContent(t *testing.T) {
	tree := newTestTree()
	d := NewDialog(tree, "Title", nil)
	content := &drawCounter{}
	d.SetContent(content)
	d.Open()
	setBounds(tree, d, 0, 0, 800, 600)

	buf := render.NewCommandBuffer()
	d.Draw(buf)
	if content.count != 1 {
		t.Error("content should be drawn")
	}
}

// --- Tabs: event handler coverage ---

func TestTabsClickHandler(t *testing.T) {
	tree := newTestTree()
	items := []TabItem{{Key: "a", Label: "A"}, {Key: "b", Label: "B"}}
	tabs := NewTabs(tree, items, nil)

	handlers := tree.Handlers(tabs.ElementID(), event.MouseClick)
	setBounds(tree, tabs, 0, 0, 400, 300)

	for _, h := range handlers {
		h(&event.Event{Type: event.MouseClick, X: 200, Y: 15})
	}
}

// --- Select: createOptionElements coverage ---

func TestSelectCreateOptionElements2(t *testing.T) {
	tree := newTestTree()
	cfg := cfgWithTextRenderer()
	opts := []SelectOption{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b", Disabled: true},
		{Label: "C", Value: "c"},
	}
	sel := NewSelect(tree, opts, cfg)
	sel.SetValue("a")

	setBounds(tree, sel, 0, 0, 200, 32)
	handlers := tree.Handlers(sel.ElementID(), event.MouseClick)
	for _, h := range handlers {
		h(&event.Event{Type: event.MouseClick})
	}
	if !sel.IsOpen() {
		t.Error("select should be open after click")
	}
}

// --- Bounds helper for Base ---

func TestBaseBoundsWithLayout(t *testing.T) {
	tree := newTestTree()
	base := NewBase(tree, core.TypeDiv, nil)
	setBounds(tree, &testWidget{Base: base}, 10, 20, 100, 50)
	b := base.Bounds()
	if b.X != 10 || b.Y != 20 || b.Width != 100 || b.Height != 50 {
		t.Errorf("bounds mismatch: %+v", b)
	}
}

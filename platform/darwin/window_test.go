//go:build darwin

package darwin

import (
	"testing"

	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
)

// ========== Event Tests ==========

func TestIsKeyPressed(t *testing.T) {
	skipIfNotDarwin(t)

	w := &Window{
		lastModifiers: event.Modifiers{
			Shift: true,
			Ctrl:  false,
			Alt:   true,
			Super: false,
		},
	}

	tests := []struct {
		key  event.Key
		want bool
	}{
		{event.KeyLeftShift, true},
		{event.KeyRightShift, true},
		{event.KeyLeftCtrl, false},
		{event.KeyRightCtrl, false},
		{event.KeyLeftAlt, true},
		{event.KeyRightAlt, true},
		{event.KeyLeftSuper, false},
		{event.KeyRightSuper, false},
		{event.KeyA, false}, // non-modifier keys always return false
	}

	for _, tt := range tests {
		got := w.isKeyPressed(tt.key)
		if got != tt.want {
			t.Errorf("isKeyPressed(%v) = %v, want %v", tt.key, got, tt.want)
		}
	}
}

func TestGetCurrentModifiers(t *testing.T) {
	skipIfNotDarwin(t)

	expected := event.Modifiers{
		Shift: true,
		Ctrl:  true,
		Alt:   false,
		Super: false,
	}

	w := &Window{
		lastModifiers: expected,
	}

	got := w.getCurrentModifiers()
	if got != expected {
		t.Errorf("getCurrentModifiers() = %+v, want %+v", got, expected)
	}
}

// ========== IME Tests ==========

func TestSetIMERect(t *testing.T) {
	skipIfNotDarwin(t)

	w := &Window{}
	
	w.SetIMERect(100, 200, 20)
	
	if w.imeX != 100 {
		t.Errorf("imeX = %d, want 100", w.imeX)
	}
	if w.imeY != 200 {
		t.Errorf("imeY = %d, want 200", w.imeY)
	}
	if w.imeLineH != 20 {
		t.Errorf("imeLineH = %d, want 20", w.imeLineH)
	}
}

func TestProcessIMEEvent(t *testing.T) {
	skipIfNotDarwin(t)

	p := &Platform{}
	w := &Window{p: p}
	
	// Test composition update
	w.processIMEEvent(event.IMECompositionUpdate, "あ", 1)
	
	if len(p.events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(p.events))
	}
	
	evt := p.events[0]
	if evt.Type != event.IMECompositionUpdate {
		t.Errorf("Event type = %v, want IMECompositionUpdate", evt.Type)
	}
	if evt.Text != "あ" {
		t.Errorf("Event text = %q, want \"あ\"", evt.Text)
	}
	if evt.IMECompositionText != "あ" {
		t.Errorf("IMECompositionText = %q, want \"あ\"", evt.IMECompositionText)
	}
	if evt.IMECursorPos != 1 {
		t.Errorf("IMECursorPos = %d, want 1", evt.IMECursorPos)
	}
	
	// Test composition end
	p.events = nil
	w.processIMEEvent(event.IMECompositionEnd, "あい", 0)
	
	if len(p.events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(p.events))
	}
	
	evt = p.events[0]
	if evt.Type != event.IMECompositionEnd {
		t.Errorf("Event type = %v, want IMECompositionEnd", evt.Type)
	}
	if evt.Text != "あい" {
		t.Errorf("Event text = %q, want \"あい\"", evt.Text)
	}
}

func TestHasMarkedTextIME(t *testing.T) {
	skipIfNotDarwin(t)

	w := &Window{}
	
	if w.hasMarkedTextIME() {
		t.Error("New window should not have marked text")
	}
	
	w.hasMarkedText = true
	if !w.hasMarkedTextIME() {
		t.Error("Window should have marked text after setting")
	}
}

func TestSetMarkedText(t *testing.T) {
	skipIfNotDarwin(t)

	w := &Window{}
	
	// Set non-empty marked text
	w.setMarkedText("あい")
	if !w.hasMarkedText {
		t.Error("hasMarkedText should be true after setting non-empty text")
	}
	if w.markedText != "あい" {
		t.Errorf("markedText = %q, want \"あい\"", w.markedText)
	}
	
	// Set empty marked text
	w.setMarkedText("")
	if w.hasMarkedText {
		t.Error("hasMarkedText should be false after setting empty text")
	}
	if w.markedText != "" {
		t.Errorf("markedText = %q, want \"\"", w.markedText)
	}
}

func TestMarkedRange(t *testing.T) {
	skipIfNotDarwin(t)

	w := &Window{}
	
	// No marked text
	r := w.markedRange()
	if r.Location != NSNotFound || r.Length != 0 {
		t.Errorf("markedRange() without text = {%d, %d}, want {NSNotFound, 0}", r.Location, r.Length)
	}
	
	// With marked text
	w.markedText = "あいうえお"
	w.hasMarkedText = true
	r = w.markedRange()
	if r.Location != 0 {
		t.Errorf("markedRange().Location = %d, want 0", r.Location)
	}
	expectedLen := uint64(len("あいうえお"))
	if r.Length != expectedLen {
		t.Errorf("markedRange().Length = %d, want %d", r.Length, expectedLen)
	}
}

func TestSelectedRange(t *testing.T) {
	skipIfNotDarwin(t)

	w := &Window{}
	
	// No marked text
	r := w.selectedRange()
	if r.Location != 0 || r.Length != 0 {
		t.Errorf("selectedRange() without text = {%d, %d}, want {0, 0}", r.Location, r.Length)
	}
	
	// With marked text
	w.markedText = "test"
	w.hasMarkedText = true
	r = w.selectedRange()
	if r.Location != 0 || r.Length != uint64(len("test")) {
		t.Errorf("selectedRange() = {%d, %d}, want {0, %d}", r.Location, r.Length, len("test"))
	}
}

func TestFirstRectForCharacterRange(t *testing.T) {
	skipIfNotDarwin(t)

	w := &Window{
		imeX:     50,
		imeY:     100,
		imeLineH: 20,
		width:    800,
		height:   600,
	}
	
	// Skip actual Cocoa call test - would require real window
	// Just verify struct is initialized correctly
	if w.imeX != 50 {
		t.Errorf("imeX = %d, want 50", w.imeX)
	}
}

func TestInsertText(t *testing.T) {
	skipIfNotDarwin(t)

	p := &Platform{}
	w := &Window{
		p:             p,
		hasMarkedText: true,
		markedText:    "あい",
	}
	
	w.insertText("あいう")
	
	// Should clear marked text
	if w.hasMarkedText {
		t.Error("hasMarkedText should be false after insertText")
	}
	if w.markedText != "" {
		t.Errorf("markedText = %q, want \"\"", w.markedText)
	}
	
	// Should generate IMECompositionEnd event
	if len(p.events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(p.events))
	}
	evt := p.events[0]
	if evt.Type != event.IMECompositionEnd {
		t.Errorf("Event type = %v, want IMECompositionEnd", evt.Type)
	}
	if evt.Text != "あいう" {
		t.Errorf("Event text = %q, want \"あいう\"", evt.Text)
	}
}

func TestSetMarkedTextIME(t *testing.T) {
	skipIfNotDarwin(t)

	p := &Platform{}
	w := &Window{p: p}
	
	selectedRange := NSRange{Location: 2, Length: 0}
	w.setMarkedTextIME("あいう", selectedRange)
	
	if !w.hasMarkedText {
		t.Error("hasMarkedText should be true after setMarkedTextIME")
	}
	if w.markedText != "あいう" {
		t.Errorf("markedText = %q, want \"あいう\"", w.markedText)
	}
	
	// Should generate IMECompositionUpdate event
	if len(p.events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(p.events))
	}
	evt := p.events[0]
	if evt.Type != event.IMECompositionUpdate {
		t.Errorf("Event type = %v, want IMECompositionUpdate", evt.Type)
	}
	if evt.IMECursorPos != 2 {
		t.Errorf("IMECursorPos = %d, want 2", evt.IMECursorPos)
	}
}

func TestUnmarkText(t *testing.T) {
	skipIfNotDarwin(t)

	p := &Platform{}
	w := &Window{
		p:             p,
		hasMarkedText: true,
		markedText:    "test",
	}
	
	w.unmarkText()
	
	if w.hasMarkedText {
		t.Error("hasMarkedText should be false after unmarkText")
	}
	if w.markedText != "" {
		t.Errorf("markedText = %q, want \"\"", w.markedText)
	}
	
	// Should generate IMECompositionEnd event
	if len(p.events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(p.events))
	}
	evt := p.events[0]
	if evt.Type != event.IMECompositionEnd {
		t.Errorf("Event type = %v, want IMECompositionEnd", evt.Type)
	}
}

func TestUnmarkTextNoMarkedText(t *testing.T) {
	skipIfNotDarwin(t)

	p := &Platform{}
	w := &Window{p: p}
	
	w.unmarkText()
	
	// Should not generate event if there was no marked text
	if len(p.events) != 0 {
		t.Errorf("Expected 0 events, got %d", len(p.events))
	}
}

func TestValidAttributesForMarkedText(t *testing.T) {
	skipIfNotDarwin(t)

	w := &Window{}
	
	// Should return empty NSArray
	result := w.validAttributesForMarkedText()
	// Can't easily verify without Cocoa, just check it doesn't panic
	_ = result
}

func TestAttributedSubstringForProposedRange(t *testing.T) {
	skipIfNotDarwin(t)

	w := &Window{}
	
	r := NSRange{Location: 0, Length: 5}
	result := w.attributedSubstringForProposedRange(r)
	
	// Should return nil
	if result != 0 {
		t.Errorf("attributedSubstringForProposedRange() = %d, want 0", result)
	}
}

func TestCharacterIndexForPoint(t *testing.T) {
	skipIfNotDarwin(t)

	w := &Window{}
	
	pt := NSPoint{X: 100, Y: 200}
	result := w.characterIndexForPoint(pt)
	
	// Should return NSNotFound
	if result != NSNotFound {
		t.Errorf("characterIndexForPoint() = %d, want NSNotFound", result)
	}
}

// ========== Window Tests (Logic Only) ==========

func TestWindowSize(t *testing.T) {
	skipIfNotDarwin(t)

	w := &Window{
		width:  1280,
		height: 720,
	}
	
	width, height := w.Size()
	if width != 1280 || height != 720 {
		t.Errorf("Size() = (%d, %d), want (1280, 720)", width, height)
	}
}

func TestWindowFramebufferSize(t *testing.T) {
	skipIfNotDarwin(t)

	w := &Window{
		fbWidth:  2560,
		fbHeight: 1440,
	}
	
	width, height := w.FramebufferSize()
	if width != 2560 || height != 1440 {
		t.Errorf("FramebufferSize() = (%d, %d), want (2560, 1440)", width, height)
	}
}

func TestWindowDPIScale(t *testing.T) {
	skipIfNotDarwin(t)

	w := &Window{
		dpiScale: 2.0,
	}
	
	scale := w.DPIScale()
	if scale != 2.0 {
		t.Errorf("DPIScale() = %f, want 2.0", scale)
	}
}

func TestWindowShouldClose(t *testing.T) {
	skipIfNotDarwin(t)

	w := &Window{}
	
	if w.ShouldClose() {
		t.Error("New window should not be marked for close")
	}
	
	w.shouldClose = true
	if !w.ShouldClose() {
		t.Error("Window should be marked for close after setting")
	}
}

func TestWindowSetShouldClose(t *testing.T) {
	skipIfNotDarwin(t)

	w := &Window{}
	
	w.SetShouldClose(true)
	if !w.shouldClose {
		t.Error("shouldClose should be true after SetShouldClose(true)")
	}
	
	w.SetShouldClose(false)
	if w.shouldClose {
		t.Error("shouldClose should be false after SetShouldClose(false)")
	}
}

func TestWindowIsFullscreen(t *testing.T) {
	skipIfNotDarwin(t)

	w := &Window{}
	
	if w.IsFullscreen() {
		t.Error("New window should not be fullscreen")
	}
	
	w.fullscreen = true
	if !w.IsFullscreen() {
		t.Error("Window should be fullscreen after setting")
	}
}

func TestWindowNativeHandle(t *testing.T) {
	skipIfNotDarwin(t)

	w := &Window{
		nswindow: 0x12345,
	}
	
	handle := w.NativeHandle()
	if handle != 0x12345 {
		t.Errorf("NativeHandle() = 0x%X, want 0x12345", handle)
	}
}

func TestWindowSetIMEPosition(t *testing.T) {
	skipIfNotDarwin(t)

	w := &Window{
		height: 600,
	}
	
	caretRect := uimath.Rect{
		X:      100,
		Y:      200,
		Width:  2,
		Height: 20,
	}
	
	w.SetIMEPosition(caretRect)
	
	if w.imeX != 100 {
		t.Errorf("imeX = %d, want 100", w.imeX)
	}
	// Note: Y coordinate is flipped
	expectedY := float64(w.height) - float64(caretRect.Y) - float64(caretRect.Height)
	if w.imeY != int32(expectedY) {
		t.Errorf("imeY = %d, want %d", w.imeY, int32(expectedY))
	}
	if w.imeLineH != 20 {
		t.Errorf("imeLineH = %d, want 20", w.imeLineH)
	}
}

func TestUpdateContentSize(t *testing.T) {
	skipIfNotDarwin(t)

	w := &Window{
		dpiScale: 2.0,
	}
	
	// Can't fully test without real NSView, but we can test the logic
	// by directly setting width/height
	w.width = 800
	w.height = 600
	
	// Manually call the framebuffer calculation logic
	w.fbWidth = int(float32(w.width) * w.dpiScale)
	w.fbHeight = int(float32(w.height) * w.dpiScale)
	
	if w.fbWidth != 1600 || w.fbHeight != 1200 {
		t.Errorf("Framebuffer size = (%d, %d), want (1600, 1200)", w.fbWidth, w.fbHeight)
	}
}

func TestWindowCursorState(t *testing.T) {
	skipIfNotDarwin(t)

	w := &Window{
		currentCursor: platform.CursorArrow,
		cursorHidden:  false,
	}
	
	if w.currentCursor != platform.CursorArrow {
		t.Errorf("currentCursor = %v, want CursorArrow", w.currentCursor)
	}
	if w.cursorHidden {
		t.Error("cursorHidden should be false")
	}
}

// Benchmark tests
func BenchmarkIsKeyPressed(b *testing.B) {
	w := &Window{
		lastModifiers: event.Modifiers{Shift: true, Ctrl: true},
	}
	
	for i := 0; i < b.N; i++ {
		_ = w.isKeyPressed(event.KeyLeftShift)
	}
}

func BenchmarkSetMarkedText(b *testing.B) {
	w := &Window{}
	
	for i := 0; i < b.N; i++ {
		w.setMarkedText("あいうえお")
		w.setMarkedText("")
	}
}

func BenchmarkMarkedRange(b *testing.B) {
	w := &Window{
		hasMarkedText: true,
		markedText:    "test text",
	}
	
	for i := 0; i < b.N; i++ {
		_ = w.markedRange()
	}
}

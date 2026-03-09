//go:build darwin

package darwin

import (
	"unsafe"

	"github.com/kasuganosora/ui/event"
)

// SetIMERect sets the position and size of the IME candidate window.
func (w *Window) SetIMERect(x, y, lineHeight int32) {
	w.imeX = x
	w.imeY = y
	w.imeLineH = lineHeight
}

// processIMEEvent handles IME-related events from NSTextInputClient.
// This is called from the view's NSTextInputClient delegate methods.
func (w *Window) processIMEEvent(eventType event.Type, text string, cursorPos int) {
	evt := event.Event{
		Type:      eventType,
		Timestamp: w.p.currentTimestamp(),
		Text:      text,
	}

	if eventType == event.IMECompositionUpdate {
		evt.IMECompositionText = text
		evt.IMECursorPos = cursorPos
	}

	w.p.pushEvent(evt)
}

// hasMarkedText returns whether there is currently marked (composition) text.
func (w *Window) hasMarkedTextIME() bool {
	return w.hasMarkedText
}

// setMarkedText sets the current marked (composition) text.
func (w *Window) setMarkedText(text string) {
	w.hasMarkedText = (text != "")
	w.markedText = text
}

// markedRange returns the range of marked text.
// Returns NSNotFound if there is no marked text.
func (w *Window) markedRange() NSRange {
	if !w.hasMarkedText {
		return NSRange{Location: NSNotFound, Length: 0}
	}
	return NSRange{Location: 0, Length: uint64(len(w.markedText))}
}

// selectedRange returns the selected range within the marked text.
// For simplicity, we return the entire marked text range.
func (w *Window) selectedRange() NSRange {
	if !w.hasMarkedText {
		return NSRange{Location: 0, Length: 0}
	}
	return NSRange{Location: 0, Length: uint64(len(w.markedText))}
}

// firstRectForCharacterRange returns the screen rect for the IME candidate window.
// This tells the IME where to position the candidate list.
func (w *Window) firstRectForCharacterRange() NSRect {
	// Convert IME position to screen coordinates
	// The IME position is in window coordinates
	var contentRect NSRect
	msgSendPtr(w.nswindow, objcSelector("contentRectForFrameRect:"), unsafe.Pointer(&contentRect),
		uintptr(msgSend(w.nswindow, objcSelector("frame"))))

	// IME rect in screen coordinates
	screenX := contentRect.Origin.X + float64(w.imeX)
	screenY := contentRect.Origin.Y + float64(w.height-int(w.imeY)-int(w.imeLineH))

	return NSRect{
		Origin: NSPoint{
			X: screenX,
			Y: screenY,
		},
		Size: NSSize{
			Width:  1.0,
			Height: float64(w.imeLineH),
		},
	}
}

// NSTextInputClient methods implementation
// These are called via the delegate mechanism from the NSView

// insertText is called when text is committed (composition finished).
func (w *Window) insertText(text string) {
	w.setMarkedText("")
	w.processIMEEvent(event.IMECompositionEnd, text, 0)
}

// setMarkedTextIME is called when composition text is updated.
func (w *Window) setMarkedTextIME(text string, selectedRange NSRange) {
	w.setMarkedText(text)
	cursorPos := int(selectedRange.Location)
	w.processIMEEvent(event.IMECompositionUpdate, text, cursorPos)
}

// unmarkText is called to remove marked text.
func (w *Window) unmarkText() {
	if w.hasMarkedText {
		w.setMarkedText("")
		w.processIMEEvent(event.IMECompositionEnd, "", 0)
	}
}

// validAttributesForMarkedText returns the valid attributes for marked text.
// We return empty set for simplicity.
func (w *Window) validAttributesForMarkedText() id {
	// Return empty NSArray
	class := objcClass("NSArray")
	return msgSend(id(class), objcSelector("array"))
}

// attributedSubstringForProposedRange returns a substring for a given range.
// This is used for reconversion (not commonly used in Western languages).
func (w *Window) attributedSubstringForProposedRange(range_ NSRange) id {
	// Return nil for simplicity
	return 0
}

// characterIndexForPoint returns the character index for a given point.
// This is used for advanced IME features (clicking on composition text).
func (w *Window) characterIndexForPoint(point NSPoint) uint64 {
	// Return NSNotFound
	return NSNotFound
}

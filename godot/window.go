package godot

import (
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
)

// HeadlessWindow implements platform.Window for headless/embedded operation.
// It reports a fixed viewport size and DPI scale with no OS window.
type HeadlessWindow struct {
	width     int
	height    int
	dpiScale  float32
	closed    bool
}

// NewHeadlessWindow creates a window with the given logical size and DPI scale.
func NewHeadlessWindow(width, height int, dpiScale float32) *HeadlessWindow {
	if dpiScale <= 0 {
		dpiScale = 1.0
	}
	return &HeadlessWindow{
		width:    width,
		height:   height,
		dpiScale: dpiScale,
	}
}

func (w *HeadlessWindow) Size() (int, int) {
	return w.width, w.height
}

func (w *HeadlessWindow) SetSize(width, height int) {
	w.width = width
	w.height = height
}

func (w *HeadlessWindow) FramebufferSize() (int, int) {
	fw := int(float32(w.width) * w.dpiScale)
	fh := int(float32(w.height) * w.dpiScale)
	return fw, fh
}

func (w *HeadlessWindow) Position() (int, int) { return 0, 0 }
func (w *HeadlessWindow) SetPosition(x, y int) {}
func (w *HeadlessWindow) SetTitle(title string) {}
func (w *HeadlessWindow) SetFullscreen(bool)    {}
func (w *HeadlessWindow) IsFullscreen() bool    { return false }

func (w *HeadlessWindow) ShouldClose() bool     { return w.closed }
func (w *HeadlessWindow) SetShouldClose(c bool) { w.closed = c }
func (w *HeadlessWindow) NativeHandle() uintptr { return 0 }
func (w *HeadlessWindow) DPIScale() float32     { return w.dpiScale }
func (w *HeadlessWindow) SetVisible(bool)       {}
func (w *HeadlessWindow) ShowDeferred()         {}
func (w *HeadlessWindow) SetMinSize(w2, h int)  {}
func (w *HeadlessWindow) SetMaxSize(w2, h int)  {}
func (w *HeadlessWindow) SetCursor(platform.CursorShape)        {}
func (w *HeadlessWindow) SetIMEPosition(uimath.Rect)            {}
func (w *HeadlessWindow) ShowContextMenu(int, int, []platform.ContextMenuItem) int {
	return -1
}
func (w *HeadlessWindow) ClientToScreen(x, y int) (int, int) { return x, y }
func (w *HeadlessWindow) IsTransparent() bool                { return false }
func (w *HeadlessWindow) SetTopMost(topmost bool)            {}
func (w *HeadlessWindow) Destroy()                           {}

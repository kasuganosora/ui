package platform

import (
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
)

// Platform is the top-level interface for OS interaction.
// Each OS provides one implementation (Windows, Linux, macOS, etc.).
// This is a DDD anti-corruption layer boundary.
type Platform interface {
	// Init initializes the platform subsystem.
	Init() error

	// CreateWindow creates a new window.
	CreateWindow(opts WindowOptions) (Window, error)

	// PollEvents polls and returns pending OS events.
	PollEvents() []event.Event

	// ProcessMessages pumps the OS message queue to keep the window responsive.
	// Unlike PollEvents, it does not collect or return events.
	// Call this during long-running operations (e.g., font rasterization) to
	// prevent the OS from marking the window as "Not Responding".
	ProcessMessages()

	// Clipboard operations
	GetClipboardText() string
	SetClipboardText(text string)

	// System information
	GetPrimaryMonitorDPI() float32
	GetSystemLocale() string

	// Terminate shuts down the platform subsystem.
	Terminate()
}

// Window represents a native OS window.
type Window interface {
	// Size returns the window size in logical pixels.
	Size() (width, height int)

	// SetSize sets the window size in logical pixels.
	SetSize(width, height int)

	// FramebufferSize returns the framebuffer size in physical pixels.
	FramebufferSize() (width, height int)

	// Position returns the window position on screen.
	Position() (x, y int)

	// SetPosition sets the window position.
	SetPosition(x, y int)

	// SetTitle sets the window title.
	SetTitle(title string)

	// SetFullscreen sets or exits fullscreen mode.
	SetFullscreen(fullscreen bool)

	// IsFullscreen returns whether the window is fullscreen.
	IsFullscreen() bool

	// ShouldClose returns true if the window has been requested to close.
	ShouldClose() bool

	// SetShouldClose sets the close flag.
	SetShouldClose(close bool)

	// NativeHandle returns the platform-specific handle (HWND, X11 Window, etc.).
	NativeHandle() uintptr

	// DPIScale returns the DPI scale factor for this window.
	DPIScale() float32

	// SetVisible shows or hides the window.
	SetVisible(visible bool)

	// ShowDeferred shows the window if its visibility was deferred at creation.
	// Called by the rendering backend after the first frame is presented, so
	// the window appears with content instead of blank. No-op if not deferred.
	ShowDeferred()

	// SetMinSize sets the minimum window size.
	SetMinSize(width, height int)

	// SetMaxSize sets the maximum window size.
	SetMaxSize(width, height int)

	// SetCursor sets the cursor shape.
	SetCursor(cursor CursorShape)

	// SetIMEPosition sets the IME composition and candidate window position.
	// The rect specifies the cursor location: X,Y is the top-left of the text line,
	// Width is 0 (or caret width), Height is the line height.
	SetIMEPosition(caretRect uimath.Rect)

	// ShowContextMenu displays a native context menu at the given client-area
	// coordinates. Returns the 0-based index of the selected item, or -1 if
	// the menu was cancelled/dismissed.
	ShowContextMenu(clientX, clientY int, items []ContextMenuItem) int

	// ClientToScreen converts client-area coordinates to screen coordinates.
	ClientToScreen(x, y int) (screenX, screenY int)

	// IsTransparent returns whether this window has per-pixel alpha enabled.
	IsTransparent() bool

	// SetTopMost sets or clears always-on-top.
	SetTopMost(topmost bool)

	// SetHitTestFunc sets a callback for per-pixel hit testing on transparent windows.
	// The function receives client-area coordinates (logical pixels) and returns true
	// if the window should handle input at that point, or false to pass it through
	// to the window below. When nil (default), all points are handled.
	SetHitTestFunc(fn func(x, y int) bool)

	// Destroy destroys the window.
	Destroy()
}

// WindowOptions specifies options for creating a new window.
type WindowOptions struct {
	Title      string
	Width      int
	Height     int
	MinWidth   int
	MinHeight  int
	MaxWidth   int
	MaxHeight  int
	Resizable   bool
	Visible     bool
	Decorated   bool // Window decorations (title bar, borders)
	Fullscreen  bool
	VSync       bool
	Transparent bool // Per-pixel alpha transparency (shaped window / desktop pet mode)
	TopMost     bool // Always on top of other windows
}

// DefaultWindowOptions returns sensible default window options.
func DefaultWindowOptions() WindowOptions {
	return WindowOptions{
		Title:     "GoUI",
		Width:     1280,
		Height:    720,
		Resizable: true,
		Visible:   true,
		Decorated: true,
		VSync:     true,
	}
}

// ContextMenuItem describes a single context menu item.
type ContextMenuItem struct {
	Label   string
	Enabled bool
}

// TSFTextProvider provides text content and position info for TSF IME.
// Widgets that support IME input (e.g., TextArea, Input) implement this
// interface so the TSF manager can query and modify their text content.
type TSFTextProvider interface {
	// GetText returns the text between byte offsets start and end.
	GetText(start, end int) string

	// GetSelection returns the current selection range (start, end).
	GetSelection() (start, end int)

	// SetSelection sets the current selection range.
	SetSelection(start, end int)

	// InsertText replaces the text between start and end with the given string.
	InsertText(start, end int, text string)

	// GetTextExtent returns the bounding rectangle (in client-area pixels)
	// of the text between start and end.
	GetTextExtent(start, end int) (x, y, w, h int)
}

// CursorShape identifies a standard cursor appearance.
type CursorShape uint8

const (
	CursorArrow CursorShape = iota
	CursorIBeam
	CursorCrosshair
	CursorHand
	CursorHResize
	CursorVResize
	CursorNWSEResize
	CursorNESWResize
	CursorAllResize
	CursorNotAllowed
	CursorWait
	CursorNone
)

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

	// SetMinSize sets the minimum window size.
	SetMinSize(width, height int)

	// SetMaxSize sets the maximum window size.
	SetMaxSize(width, height int)

	// SetCursor sets the cursor shape.
	SetCursor(cursor CursorShape)

	// SetIMEPosition sets the IME candidate window position.
	SetIMEPosition(pos uimath.Vec2)

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
	Resizable  bool
	Visible    bool
	Decorated  bool // Window decorations (title bar, borders)
	Fullscreen bool
	VSync      bool
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

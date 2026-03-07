package event

import uimath "github.com/kasuganosora/ui/math"

// Type identifies the kind of event.
type Type uint16

const (
	// Mouse events
	MouseMove       Type = iota + 1
	MouseDown            // MouseButton in Button field
	MouseUp              // MouseButton in Button field
	MouseClick           // Synthetic click (down + up on same element)
	MouseDoubleClick
	MouseWheel
	MouseEnter
	MouseLeave

	// Keyboard events
	KeyDown
	KeyUp
	KeyPress // Character input (non-IME)

	// IME events
	IMECompositionStart
	IMECompositionUpdate
	IMECompositionEnd
	IMECandidateOpen
	IMECandidateClose

	// Focus events
	FocusIn
	FocusOut
	Blur
	Focus

	// Touch events
	TouchStart
	TouchMove
	TouchEnd
	TouchCancel

	// Drag events
	DragStart
	DragMove
	DragEnd
	DragEnter
	DragLeave
	Drop

	// Gamepad events
	GamepadButtonDown
	GamepadButtonUp
	GamepadAxis

	// Window events
	WindowResize
	WindowClose
	WindowFocus
	WindowBlur
	WindowDPIChange

	// Custom event
	Custom
)

// MouseButton identifies a mouse button.
type MouseButton uint8

const (
	MouseButtonLeft   MouseButton = 0
	MouseButtonRight  MouseButton = 1
	MouseButtonMiddle MouseButton = 2
	MouseButton4      MouseButton = 3
	MouseButton5      MouseButton = 4
)

// Modifiers holds keyboard modifier key state.
type Modifiers struct {
	Ctrl  bool
	Shift bool
	Alt   bool
	Super bool // Windows key / Cmd
}

// Key represents a keyboard key code (virtual key).
type Key uint16

const (
	KeyUnknown Key = iota

	// Letters
	KeyA
	KeyB
	KeyC
	KeyD
	KeyE
	KeyF
	KeyG
	KeyH
	KeyI
	KeyJ
	KeyK
	KeyL
	KeyM
	KeyN
	KeyO
	KeyP
	KeyQ
	KeyR
	KeyS
	KeyT
	KeyU
	KeyV
	KeyW
	KeyX
	KeyY
	KeyZ

	// Numbers
	Key0
	Key1
	Key2
	Key3
	Key4
	Key5
	Key6
	Key7
	Key8
	Key9

	// Function keys
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12

	// Special keys
	KeyEscape
	KeyEnter
	KeyTab
	KeyBackspace
	KeyDelete
	KeyInsert
	KeyHome
	KeyEnd
	KeyPageUp
	KeyPageDown
	KeySpace

	// Arrow keys
	KeyArrowLeft
	KeyArrowRight
	KeyArrowUp
	KeyArrowDown

	// Modifier keys
	KeyLeftCtrl
	KeyRightCtrl
	KeyLeftShift
	KeyRightShift
	KeyLeftAlt
	KeyRightAlt
	KeyLeftSuper
	KeyRightSuper

	// Punctuation
	KeyMinus
	KeyEqual
	KeyLeftBracket
	KeyRightBracket
	KeyBackslash
	KeySemicolon
	KeyApostrophe
	KeyGraveAccent
	KeyComma
	KeyPeriod
	KeySlash

	// Numpad
	KeyNumpad0
	KeyNumpad1
	KeyNumpad2
	KeyNumpad3
	KeyNumpad4
	KeyNumpad5
	KeyNumpad6
	KeyNumpad7
	KeyNumpad8
	KeyNumpad9
	KeyNumpadAdd
	KeyNumpadSubtract
	KeyNumpadMultiply
	KeyNumpadDivide
	KeyNumpadDecimal
	KeyNumpadEnter

	// Misc
	KeyCapsLock
	KeyNumLock
	KeyScrollLock
	KeyPrintScreen
	KeyPause
	KeyMenu
)

// Phase represents the event propagation phase.
type Phase uint8

const (
	PhaseNone    Phase = iota
	PhaseCapture       // Top-down capture phase
	PhaseTarget        // Arrived at target
	PhaseBubble        // Bottom-up bubble phase
)

// Event is the core event structure. Immutable value object.
type Event struct {
	Type      Type
	Phase     Phase
	Timestamp uint64 // Monotonic time in microseconds

	// Mouse/Touch fields
	X, Y       float32     // Position relative to target element
	GlobalX    float32     // Position relative to window
	GlobalY    float32     // Position relative to window
	Button     MouseButton // Which button for mouse events
	WheelDX    float32     // Horizontal scroll
	WheelDY    float32     // Vertical scroll
	ClickCount int         // 1=single, 2=double, 3=triple

	// Keyboard fields
	Key       Key
	Modifiers Modifiers
	Char      rune   // For KeyPress: the character typed
	Text      string // For IME events: composition/commit text

	// IME fields
	IMECompositionText string         // Current composition string
	IMECursorPos       int            // Cursor position within composition
	IMECandidates      []string       // Candidate list
	IMECandidateIndex  int            // Selected candidate index
	IMECaretRect       uimath.Rect    // Caret position for candidate window

	// Touch fields
	TouchID    int
	TouchCount int // Number of active touches
	Touches    []TouchPoint

	// Gamepad fields
	GamepadID     int
	GamepadButton int
	GamepadAxis   int
	GamepadValue  float32

	// Window fields
	WindowWidth  int
	WindowHeight int
	DPIScale     float32

	// Drag fields
	DragData any

	// Custom event fields
	CustomType string
	CustomData any

	// Propagation control (set by handlers)
	stopped          bool
	defaultPrevented bool
}

// TouchPoint represents a single touch point.
type TouchPoint struct {
	ID      int
	X, Y    float32
	GlobalX float32
	GlobalY float32
}

// StopPropagation stops the event from propagating further.
func (e *Event) StopPropagation() {
	e.stopped = true
}

// IsStopped returns whether propagation has been stopped.
func (e *Event) IsStopped() bool {
	return e.stopped
}

// PreventDefault prevents the default action for this event.
func (e *Event) PreventDefault() {
	e.defaultPrevented = true
}

// IsDefaultPrevented returns whether the default action has been prevented.
func (e *Event) IsDefaultPrevented() bool {
	return e.defaultPrevented
}

// HasModifier returns true if any modifier key is held.
func (e *Event) HasModifier() bool {
	return e.Modifiers.Ctrl || e.Modifiers.Shift || e.Modifiers.Alt || e.Modifiers.Super
}

// IsMouse returns true if this is a mouse-related event.
func (t Type) IsMouse() bool {
	return t >= MouseMove && t <= MouseLeave
}

// IsKeyboard returns true if this is a keyboard-related event.
func (t Type) IsKeyboard() bool {
	return t >= KeyDown && t <= KeyPress
}

// IsIME returns true if this is an IME-related event.
func (t Type) IsIME() bool {
	return t >= IMECompositionStart && t <= IMECandidateClose
}

// IsFocus returns true if this is a focus-related event.
func (t Type) IsFocus() bool {
	return t >= FocusIn && t <= Focus
}

// IsTouch returns true if this is a touch-related event.
func (t Type) IsTouch() bool {
	return t >= TouchStart && t <= TouchCancel
}

// IsDrag returns true if this is a drag-related event.
func (t Type) IsDrag() bool {
	return t >= DragStart && t <= Drop
}

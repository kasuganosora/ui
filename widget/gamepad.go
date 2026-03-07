package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// GamepadButton constants for standard gamepad layout.
const (
	GPBtnA     = 0  // A / Cross
	GPBtnB     = 1  // B / Circle
	GPBtnX     = 2  // X / Square
	GPBtnY     = 3  // Y / Triangle
	GPBtnLB    = 4  // Left bumper
	GPBtnRB    = 5  // Right bumper
	GPBtnBack  = 6  // Back / Select
	GPBtnStart = 7  // Start / Options
	GPBtnLS    = 8  // Left stick press
	GPBtnRS    = 9  // Right stick press
	GPBtnUp    = 10 // D-pad up
	GPBtnDown  = 11 // D-pad down
	GPBtnLeft  = 12 // D-pad left
	GPBtnRight = 13 // D-pad right
)

// GamepadAxis constants for standard gamepad layout.
const (
	GPAxisLeftX  = 0
	GPAxisLeftY  = 1
	GPAxisRightX = 2
	GPAxisRightY = 3
	GPAxisLT     = 4 // Left trigger
	GPAxisRT     = 5 // Right trigger
)

// NavDirection represents a navigation direction.
type NavDirection uint8

const (
	NavUp    NavDirection = iota
	NavDown
	NavLeft
	NavRight
)

// Navigable is implemented by widgets that support gamepad/keyboard navigation.
type Navigable interface {
	Widget
	// NavFocusable returns true if this widget can receive navigation focus.
	NavFocusable() bool
}

// GamepadNavigator manages focus navigation between widgets using gamepad input.
type GamepadNavigator struct {
	Base
	widgets     []Navigable
	focusIndex  int
	enabled     bool
	deadzone    float32  // axis deadzone threshold
	repeatDelay float32  // seconds before repeat starts
	repeatRate  float32  // seconds between repeats
	axisAccum   float32  // accumulated axis time for repeat
	lastDir     NavDirection
	moved       bool     // has axis triggered a move this press
	showFocus   bool     // whether to show focus indicator
	onNavigate  func(widget Navigable)
	onActivate  func(widget Navigable)
	onCancel    func()
}

// NewGamepadNavigator creates a gamepad navigation manager.
func NewGamepadNavigator(tree *core.Tree, cfg *Config) *GamepadNavigator {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	gn := &GamepadNavigator{
		Base:        NewBase(tree, core.TypeCustom, cfg),
		enabled:     true,
		deadzone:    0.3,
		repeatDelay: 0.4,
		repeatRate:  0.15,
		focusIndex:  -1,
		showFocus:   true,
	}
	tree.AddHandler(gn.id, event.GamepadButtonDown, gn.onGamepadButton)
	tree.AddHandler(gn.id, event.GamepadAxis, gn.onGamepadAxis)
	// Also support keyboard navigation
	tree.AddHandler(gn.id, event.KeyDown, gn.onKeyDown)
	return gn
}

// SetEnabled enables or disables gamepad navigation.
func (gn *GamepadNavigator) SetEnabled(e bool)           { gn.enabled = e }
func (gn *GamepadNavigator) IsEnabled() bool              { return gn.enabled }
func (gn *GamepadNavigator) SetDeadzone(d float32)        { gn.deadzone = d }
func (gn *GamepadNavigator) SetRepeatDelay(d float32)     { gn.repeatDelay = d }
func (gn *GamepadNavigator) SetRepeatRate(r float32)      { gn.repeatRate = r }
func (gn *GamepadNavigator) ShowFocus() bool               { return gn.showFocus }
func (gn *GamepadNavigator) SetShowFocus(s bool)           { gn.showFocus = s }
func (gn *GamepadNavigator) FocusIndex() int               { return gn.focusIndex }
func (gn *GamepadNavigator) OnNavigate(fn func(Navigable)) { gn.onNavigate = fn }
func (gn *GamepadNavigator) OnActivate(fn func(Navigable)) { gn.onActivate = fn }
func (gn *GamepadNavigator) OnCancel(fn func())            { gn.onCancel = fn }

// AddWidget registers a navigable widget.
func (gn *GamepadNavigator) AddWidget(w Navigable) {
	gn.widgets = append(gn.widgets, w)
	if gn.focusIndex < 0 && w.NavFocusable() {
		gn.focusIndex = len(gn.widgets) - 1
	}
}

// RemoveWidget removes a navigable widget.
func (gn *GamepadNavigator) RemoveWidget(w Navigable) {
	for i, nw := range gn.widgets {
		if nw.ElementID() == w.ElementID() {
			gn.widgets = append(gn.widgets[:i], gn.widgets[i+1:]...)
			if gn.focusIndex >= len(gn.widgets) {
				gn.focusIndex = len(gn.widgets) - 1
			}
			return
		}
	}
}

// ClearWidgets removes all registered widgets.
func (gn *GamepadNavigator) ClearWidgets() {
	gn.widgets = gn.widgets[:0]
	gn.focusIndex = -1
}

// Widgets returns the registered navigable widgets.
func (gn *GamepadNavigator) Widgets() []Navigable { return gn.widgets }

// FocusedWidget returns the currently focused widget, or nil.
func (gn *GamepadNavigator) FocusedWidget() Navigable {
	if gn.focusIndex >= 0 && gn.focusIndex < len(gn.widgets) {
		return gn.widgets[gn.focusIndex]
	}
	return nil
}

// SetFocus sets focus to the widget at the given index.
func (gn *GamepadNavigator) SetFocus(index int) {
	if index >= 0 && index < len(gn.widgets) {
		gn.focusIndex = index
		if gn.onNavigate != nil {
			gn.onNavigate(gn.widgets[index])
		}
	}
}

// Navigate moves focus in the given direction.
func (gn *GamepadNavigator) Navigate(dir NavDirection) {
	if !gn.enabled || len(gn.widgets) == 0 {
		return
	}

	switch dir {
	case NavDown, NavRight:
		gn.navigateNext()
	case NavUp, NavLeft:
		gn.navigatePrev()
	}
}

func (gn *GamepadNavigator) navigateNext() {
	start := gn.focusIndex
	for i := 1; i <= len(gn.widgets); i++ {
		idx := (start + i) % len(gn.widgets)
		if gn.widgets[idx].NavFocusable() {
			gn.focusIndex = idx
			if gn.onNavigate != nil {
				gn.onNavigate(gn.widgets[idx])
			}
			return
		}
	}
}

func (gn *GamepadNavigator) navigatePrev() {
	start := gn.focusIndex
	if start < 0 {
		start = 0
	}
	for i := 1; i <= len(gn.widgets); i++ {
		idx := (start - i + len(gn.widgets)) % len(gn.widgets)
		if gn.widgets[idx].NavFocusable() {
			gn.focusIndex = idx
			if gn.onNavigate != nil {
				gn.onNavigate(gn.widgets[idx])
			}
			return
		}
	}
}

// Activate triggers the activate action on the focused widget.
func (gn *GamepadNavigator) Activate() {
	w := gn.FocusedWidget()
	if w != nil && gn.onActivate != nil {
		gn.onActivate(w)
	}
}

// Cancel triggers the cancel action.
func (gn *GamepadNavigator) Cancel() {
	if gn.onCancel != nil {
		gn.onCancel()
	}
}

// Tick advances repeat timers. Call each frame with delta time.
func (gn *GamepadNavigator) Tick(dt float32) {
	if !gn.enabled || !gn.moved {
		gn.axisAccum = 0
		return
	}
	gn.axisAccum += dt
	if gn.axisAccum >= gn.repeatDelay {
		gn.axisAccum -= gn.repeatRate
		gn.Navigate(gn.lastDir)
	}
}

func (gn *GamepadNavigator) onGamepadButton(e *event.Event) {
	if !gn.enabled {
		return
	}
	switch e.GamepadButton {
	case GPBtnUp:
		gn.Navigate(NavUp)
	case GPBtnDown:
		gn.Navigate(NavDown)
	case GPBtnLeft:
		gn.Navigate(NavLeft)
	case GPBtnRight:
		gn.Navigate(NavRight)
	case GPBtnA:
		gn.Activate()
	case GPBtnB:
		gn.Cancel()
	}
}

func (gn *GamepadNavigator) onGamepadAxis(e *event.Event) {
	if !gn.enabled {
		return
	}
	v := e.GamepadValue
	switch e.GamepadAxis {
	case GPAxisLeftX:
		if v > gn.deadzone {
			if !gn.moved || gn.lastDir != NavRight {
				gn.Navigate(NavRight)
				gn.lastDir = NavRight
				gn.moved = true
				gn.axisAccum = 0
			}
		} else if v < -gn.deadzone {
			if !gn.moved || gn.lastDir != NavLeft {
				gn.Navigate(NavLeft)
				gn.lastDir = NavLeft
				gn.moved = true
				gn.axisAccum = 0
			}
		} else {
			gn.moved = false
		}
	case GPAxisLeftY:
		if v > gn.deadzone {
			if !gn.moved || gn.lastDir != NavDown {
				gn.Navigate(NavDown)
				gn.lastDir = NavDown
				gn.moved = true
				gn.axisAccum = 0
			}
		} else if v < -gn.deadzone {
			if !gn.moved || gn.lastDir != NavUp {
				gn.Navigate(NavUp)
				gn.lastDir = NavUp
				gn.moved = true
				gn.axisAccum = 0
			}
		} else {
			gn.moved = false
		}
	}
}

func (gn *GamepadNavigator) onKeyDown(e *event.Event) {
	if !gn.enabled {
		return
	}
	switch e.Key {
	case event.KeyArrowUp:
		gn.Navigate(NavUp)
	case event.KeyArrowDown:
		gn.Navigate(NavDown)
	case event.KeyArrowLeft:
		gn.Navigate(NavLeft)
	case event.KeyArrowRight:
		gn.Navigate(NavRight)
	case event.KeyEnter, event.KeySpace:
		gn.Activate()
	case event.KeyEscape:
		gn.Cancel()
	}
}

// Draw renders the focus indicator on the currently focused widget.
func (gn *GamepadNavigator) Draw(buf *render.CommandBuffer) {
	if !gn.enabled || !gn.showFocus {
		return
	}
	w := gn.FocusedWidget()
	if w == nil {
		return
	}
	elem := gn.tree.Get(w.ElementID())
	if elem == nil {
		return
	}
	bounds := elem.Layout().Bounds
	if bounds.IsEmpty() {
		return
	}
	cfg := gn.config
	// Draw focus ring
	buf.DrawOverlay(render.RectCmd{
		Bounds:      bounds.Expand(3),
		BorderColor: cfg.FocusBorderColor,
		BorderWidth: 2,
		Corners:     uimath.CornersAll(cfg.BorderRadius + 2),
	}, 90, 0.8)
}

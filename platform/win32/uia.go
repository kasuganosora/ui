//go:build windows

package win32

import (
	"syscall"
)

// UIA constants
const (
	WM_GETOBJECT = 0x003D

	// UI Automation provider control type IDs
	UIA_ButtonControlTypeId   = 50000
	UIA_TextControlTypeId     = 50020
	UIA_EditControlTypeId     = 50004
	UIA_WindowControlTypeId   = 50032
	UIA_CheckBoxControlTypeId = 50002
	UIA_GroupControlTypeId    = 50026

	// Property IDs
	UIA_NamePropertyId             = 30005
	UIA_ControlTypePropertyId      = 30003
	UIA_IsEnabledPropertyId        = 30010
	UIA_HasKeyboardFocusPropertyId = 30008
	UIA_AutomationIdPropertyId     = 30011
)

var (
	uiautomationcore                = syscall.NewLazyDLL("uiautomationcore.dll")
	procUiaReturnRawElementProvider = uiautomationcore.NewProc("UiaReturnRawElementProvider")
	procUiaHostProviderFromHwnd     = uiautomationcore.NewProc("UiaHostProviderFromHwnd")
	procUiaRaiseAutomationEvent     = uiautomationcore.NewProc("UiaRaiseAutomationEvent")
)

// UIAProvider is the root UIA provider for a window.
type UIAProvider struct {
	hwnd    uintptr
	enabled bool
}

// NewUIAProvider creates a UIA provider for the given window handle.
func NewUIAProvider(hwnd uintptr) *UIAProvider {
	return &UIAProvider{
		hwnd:    hwnd,
		enabled: true,
	}
}

// HandleGetObject processes WM_GETOBJECT messages for UIA.
// Returns (result, handled). If handled is false, the message was not for UIA.
func (p *UIAProvider) HandleGetObject(wParam, lParam uintptr) (uintptr, bool) {
	if p == nil || !p.enabled {
		return 0, false
	}

	// UIA sends WM_GETOBJECT with specific lParam values
	// OBJID_CLIENT = -4 (0xFFFFFFFC)
	const OBJID_CLIENT = -4
	if int32(lParam) != int32(OBJID_CLIENT) {
		return 0, false
	}

	// For now, return 0 to let the system handle it with default behavior.
	// A full implementation would return a COM IRawElementProviderSimple here
	// via UiaReturnRawElementProvider.
	return 0, false
}

// RaiseStructureChanged notifies screen readers that the UI tree changed.
func (p *UIAProvider) RaiseStructureChanged() {
	// Stub for future implementation
}

// RaiseFocusChanged notifies screen readers that focus moved.
func (p *UIAProvider) RaiseFocusChanged() {
	// Stub for future implementation
}

// RaisePropertyChanged notifies screen readers of a property change.
func (p *UIAProvider) RaisePropertyChanged(propertyID int32) {
	// Stub for future implementation
}

// SetEnabled enables or disables UIA support.
func (p *UIAProvider) SetEnabled(enabled bool) {
	if p != nil {
		p.enabled = enabled
	}
}

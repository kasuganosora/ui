// Package accessibility provides basic accessibility support for UI widgets.
// It defines roles, properties, and an accessible tree that maps to the widget tree.
package accessibility

// Role identifies the accessibility role of a UI element.
type Role uint8

const (
	RoleNone        Role = iota
	RoleWindow
	RoleButton
	RoleCheckBox
	RoleRadioButton
	RoleTextBox  // single-line text input
	RoleTextArea // multi-line text input
	RoleLabel    // static text
	RoleImage
	RoleLink
	RoleList
	RoleListItem
	RoleMenu
	RoleMenuItem
	RoleMenuBar
	RoleTab
	RoleTabPanel
	RoleDialog
	RoleAlert
	RoleProgressBar
	RoleSlider
	RoleScrollBar
	RoleTree
	RoleTreeItem
	RoleTable
	RoleRow
	RoleCell
	RoleGroup
	RoleToolbar
	RoleSeparator
	RoleComboBox
	RoleSwitch
)

// String returns a human-readable name for the role.
func (r Role) String() string {
	switch r {
	case RoleNone:
		return "none"
	case RoleWindow:
		return "window"
	case RoleButton:
		return "button"
	case RoleCheckBox:
		return "checkbox"
	case RoleRadioButton:
		return "radio"
	case RoleTextBox:
		return "textbox"
	case RoleTextArea:
		return "textarea"
	case RoleLabel:
		return "label"
	case RoleImage:
		return "image"
	case RoleLink:
		return "link"
	case RoleList:
		return "list"
	case RoleListItem:
		return "listitem"
	case RoleMenu:
		return "menu"
	case RoleMenuItem:
		return "menuitem"
	case RoleMenuBar:
		return "menubar"
	case RoleTab:
		return "tab"
	case RoleTabPanel:
		return "tabpanel"
	case RoleDialog:
		return "dialog"
	case RoleAlert:
		return "alert"
	case RoleProgressBar:
		return "progressbar"
	case RoleSlider:
		return "slider"
	case RoleScrollBar:
		return "scrollbar"
	case RoleTree:
		return "tree"
	case RoleTreeItem:
		return "treeitem"
	case RoleTable:
		return "table"
	case RoleRow:
		return "row"
	case RoleCell:
		return "cell"
	case RoleGroup:
		return "group"
	case RoleToolbar:
		return "toolbar"
	case RoleSeparator:
		return "separator"
	case RoleComboBox:
		return "combobox"
	case RoleSwitch:
		return "switch"
	default:
		return "unknown"
	}
}

// State holds boolean accessibility states.
type State uint16

const (
	StateNone      State = 0
	StateFocused   State = 1 << iota
	StateSelected
	StateChecked
	StateDisabled
	StateExpanded
	StateCollapsed
	StateReadOnly
	StateRequired
	StateHovered
	StatePressed
)

// Has returns true if the state flag is set.
func (s State) Has(flag State) bool { return s&flag != 0 }

// Props holds accessibility properties for a single UI element.
type Props struct {
	Role        Role
	Name        string // accessible name (label text, button text, etc.)
	Description string // additional description
	Value       string // current value (for sliders, inputs, etc.)
	State       State

	// Live region support for dynamic content updates
	Live string // "", "polite", or "assertive"

	// Relationships
	LabelledBy  uint64 // ElementID of the labelling element
	DescribedBy uint64 // ElementID of the describing element

	// Range values (for sliders, progress bars)
	ValueMin float64
	ValueMax float64
	ValueNow float64
}

// Accessible is implemented by widgets that expose accessibility information.
type Accessible interface {
	AccessibleProps() Props
}

// DefaultProps returns sensible default Props for common element types.
func DefaultProps(elemType string, text string) Props {
	p := Props{Name: text}
	switch elemType {
	case "button":
		p.Role = RoleButton
	case "input":
		p.Role = RoleTextBox
	case "text", "div":
		p.Role = RoleGroup
	case "image":
		p.Role = RoleImage
	default:
		p.Role = RoleGroup
	}
	return p
}

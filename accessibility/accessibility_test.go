package accessibility

import "testing"

func TestRoleString(t *testing.T) {
	tests := []struct {
		role Role
		want string
	}{
		{RoleNone, "none"},
		{RoleWindow, "window"},
		{RoleButton, "button"},
		{RoleCheckBox, "checkbox"},
		{RoleRadioButton, "radio"},
		{RoleTextBox, "textbox"},
		{RoleTextArea, "textarea"},
		{RoleLabel, "label"},
		{RoleImage, "image"},
		{RoleLink, "link"},
		{RoleList, "list"},
		{RoleListItem, "listitem"},
		{RoleMenu, "menu"},
		{RoleMenuItem, "menuitem"},
		{RoleMenuBar, "menubar"},
		{RoleTab, "tab"},
		{RoleTabPanel, "tabpanel"},
		{RoleDialog, "dialog"},
		{RoleAlert, "alert"},
		{RoleProgressBar, "progressbar"},
		{RoleSlider, "slider"},
		{RoleScrollBar, "scrollbar"},
		{RoleTree, "tree"},
		{RoleTreeItem, "treeitem"},
		{RoleTable, "table"},
		{RoleRow, "row"},
		{RoleCell, "cell"},
		{RoleGroup, "group"},
		{RoleToolbar, "toolbar"},
		{RoleSeparator, "separator"},
		{RoleComboBox, "combobox"},
		{RoleSwitch, "switch"},
		{Role(255), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.role.String(); got != tt.want {
			t.Errorf("Role(%d).String() = %q, want %q", tt.role, got, tt.want)
		}
	}
}

func TestStateHas(t *testing.T) {
	s := StateFocused | StateChecked | StateHovered

	if !s.Has(StateFocused) {
		t.Error("expected StateFocused to be set")
	}
	if !s.Has(StateChecked) {
		t.Error("expected StateChecked to be set")
	}
	if !s.Has(StateHovered) {
		t.Error("expected StateHovered to be set")
	}
	if s.Has(StateDisabled) {
		t.Error("expected StateDisabled to not be set")
	}
	if s.Has(StateSelected) {
		t.Error("expected StateSelected to not be set")
	}
	if s.Has(StatePressed) {
		t.Error("expected StatePressed to not be set")
	}
}

func TestStateNone(t *testing.T) {
	s := StateNone
	if s.Has(StateFocused) {
		t.Error("StateNone should not have any flags set")
	}
	if s.Has(StateDisabled) {
		t.Error("StateNone should not have any flags set")
	}
}

func TestDefaultProps(t *testing.T) {
	tests := []struct {
		elemType string
		text     string
		wantRole Role
		wantName string
	}{
		{"button", "Click me", RoleButton, "Click me"},
		{"input", "Username", RoleTextBox, "Username"},
		{"text", "Hello", RoleGroup, "Hello"},
		{"div", "Container", RoleGroup, "Container"},
		{"image", "Logo", RoleImage, "Logo"},
		{"unknown", "Stuff", RoleGroup, "Stuff"},
		{"", "", RoleGroup, ""},
	}
	for _, tt := range tests {
		p := DefaultProps(tt.elemType, tt.text)
		if p.Role != tt.wantRole {
			t.Errorf("DefaultProps(%q, %q).Role = %v, want %v", tt.elemType, tt.text, p.Role, tt.wantRole)
		}
		if p.Name != tt.wantName {
			t.Errorf("DefaultProps(%q, %q).Name = %q, want %q", tt.elemType, tt.text, p.Name, tt.wantName)
		}
	}
}

func TestPropsZeroValue(t *testing.T) {
	var p Props
	if p.Role != RoleNone {
		t.Errorf("zero Props.Role = %v, want RoleNone", p.Role)
	}
	if p.Name != "" {
		t.Error("zero Props.Name should be empty")
	}
	if p.Description != "" {
		t.Error("zero Props.Description should be empty")
	}
	if p.Value != "" {
		t.Error("zero Props.Value should be empty")
	}
	if p.State != StateNone {
		t.Error("zero Props.State should be StateNone")
	}
	if p.Live != "" {
		t.Error("zero Props.Live should be empty")
	}
	if p.LabelledBy != 0 {
		t.Error("zero Props.LabelledBy should be 0")
	}
	if p.DescribedBy != 0 {
		t.Error("zero Props.DescribedBy should be 0")
	}
	if p.ValueMin != 0 || p.ValueMax != 0 || p.ValueNow != 0 {
		t.Error("zero Props range values should all be 0")
	}
}

package event

import "testing"

func TestStopPropagation(t *testing.T) {
	e := &Event{Type: MouseClick}
	if e.IsStopped() {
		t.Error("should not be stopped initially")
	}
	e.StopPropagation()
	if !e.IsStopped() {
		t.Error("should be stopped after StopPropagation")
	}
}

func TestPreventDefault(t *testing.T) {
	e := &Event{Type: MouseClick}
	if e.IsDefaultPrevented() {
		t.Error("should not be prevented initially")
	}
	e.PreventDefault()
	if !e.IsDefaultPrevented() {
		t.Error("should be prevented after PreventDefault")
	}
}

func TestHasModifier(t *testing.T) {
	e := &Event{}
	if e.HasModifier() {
		t.Error("no modifier should return false")
	}
	e.Modifiers.Ctrl = true
	if !e.HasModifier() {
		t.Error("ctrl modifier should return true")
	}

	e2 := &Event{}
	e2.Modifiers.Shift = true
	if !e2.HasModifier() {
		t.Error("shift modifier should return true")
	}

	e3 := &Event{}
	e3.Modifiers.Alt = true
	if !e3.HasModifier() {
		t.Error("alt modifier should return true")
	}

	e4 := &Event{}
	e4.Modifiers.Super = true
	if !e4.HasModifier() {
		t.Error("super modifier should return true")
	}
}

func TestTypeIsMouse(t *testing.T) {
	mouseTypes := []Type{MouseMove, MouseDown, MouseUp, MouseClick, MouseDoubleClick, MouseWheel, MouseEnter, MouseLeave}
	for _, mt := range mouseTypes {
		if !mt.IsMouse() {
			t.Errorf("type %d should be mouse", mt)
		}
	}
	nonMouse := []Type{KeyDown, KeyUp, FocusIn, TouchStart, DragStart}
	for _, nm := range nonMouse {
		if nm.IsMouse() {
			t.Errorf("type %d should not be mouse", nm)
		}
	}
}

func TestTypeIsKeyboard(t *testing.T) {
	kbTypes := []Type{KeyDown, KeyUp, KeyPress}
	for _, kt := range kbTypes {
		if !kt.IsKeyboard() {
			t.Errorf("type %d should be keyboard", kt)
		}
	}
	if MouseMove.IsKeyboard() {
		t.Error("MouseMove should not be keyboard")
	}
}

func TestTypeIsIME(t *testing.T) {
	imeTypes := []Type{IMECompositionStart, IMECompositionUpdate, IMECompositionEnd, IMECandidateOpen, IMECandidateClose}
	for _, it := range imeTypes {
		if !it.IsIME() {
			t.Errorf("type %d should be IME", it)
		}
	}
	if KeyDown.IsIME() {
		t.Error("KeyDown should not be IME")
	}
}

func TestTypeIsFocus(t *testing.T) {
	focusTypes := []Type{FocusIn, FocusOut, Blur, Focus}
	for _, ft := range focusTypes {
		if !ft.IsFocus() {
			t.Errorf("type %d should be focus", ft)
		}
	}
	if MouseClick.IsFocus() {
		t.Error("MouseClick should not be focus")
	}
}

func TestTypeIsTouch(t *testing.T) {
	touchTypes := []Type{TouchStart, TouchMove, TouchEnd, TouchCancel}
	for _, tt := range touchTypes {
		if !tt.IsTouch() {
			t.Errorf("type %d should be touch", tt)
		}
	}
	if KeyDown.IsTouch() {
		t.Error("KeyDown should not be touch")
	}
}

func TestTypeIsDrag(t *testing.T) {
	dragTypes := []Type{DragStart, DragMove, DragEnd, DragEnter, DragLeave, Drop}
	for _, dt := range dragTypes {
		if !dt.IsDrag() {
			t.Errorf("type %d should be drag", dt)
		}
	}
	if MouseClick.IsDrag() {
		t.Error("MouseClick should not be drag")
	}
}

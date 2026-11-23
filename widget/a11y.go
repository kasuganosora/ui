package widget

import "github.com/kasuganosora/ui/accessibility"

func focusedState(b Base) accessibility.State {
	if elem := b.Element(); elem != nil && elem.IsFocused() {
		return accessibility.StateFocused
	}
	return accessibility.StateNone
}

func (b *Button) AccessibleProps() accessibility.Props {
	p := accessibility.Props{Role: accessibility.RoleButton, Name: b.content, State: focusedState(b.Base)}
	if b.disabled || b.loading {
		p.State |= accessibility.StateDisabled
	}
	if b.pressed {
		p.State |= accessibility.StatePressed
	}
	return p
}

func (c *Checkbox) AccessibleProps() accessibility.Props {
	p := accessibility.Props{Role: accessibility.RoleCheckBox, Name: c.label, State: focusedState(c.Base)}
	if c.checked {
		p.State |= accessibility.StateChecked
	}
	if c.disabled {
		p.State |= accessibility.StateDisabled
	}
	return p
}

func (r *Radio) AccessibleProps() accessibility.Props {
	p := accessibility.Props{Role: accessibility.RoleRadioButton, Name: r.label, State: focusedState(r.Base)}
	if r.checked {
		p.State |= accessibility.StateChecked
	}
	if r.disabled {
		p.State |= accessibility.StateDisabled
	}
	return p
}

func (s *Switch) AccessibleProps() accessibility.Props {
	p := accessibility.Props{Role: accessibility.RoleSwitch, Name: "switch", State: focusedState(s.Base)}
	if s.value {
		p.State |= accessibility.StateChecked
	}
	if s.disabled {
		p.State |= accessibility.StateDisabled
	}
	return p
}

func (i *Input) AccessibleProps() accessibility.Props {
	return accessibility.Props{Role: accessibility.RoleTextBox, Name: i.placeholder, Value: i.value, State: focusedState(i.Base)}
}

func (t *TextArea) AccessibleProps() accessibility.Props {
	return accessibility.Props{Role: accessibility.RoleTextArea, Name: t.placeholder, Value: t.value, State: focusedState(t.Base)}
}

func (s *Slider) AccessibleProps() accessibility.Props {
	return accessibility.Props{
		Role:     accessibility.RoleSlider,
		Name:     "slider",
		ValueNow: float64(s.Value()),
		ValueMin: float64(s.min),
		ValueMax: float64(s.max),
		State:    focusedState(s.Base),
	}
}

func (p *Progress) AccessibleProps() accessibility.Props {
	return accessibility.Props{
		Role:     accessibility.RoleProgressBar,
		Name:     "progress",
		ValueNow: float64(p.percentage),
		ValueMin: 0,
		ValueMax: 100,
	}
}

func (t *Tabs) AccessibleProps() accessibility.Props {
	return accessibility.Props{Role: accessibility.RoleTab, Name: t.value, State: focusedState(t.Base)}
}

func (d *Dialog) AccessibleProps() accessibility.Props {
	st := accessibility.StateNone
	if d.visible {
		st |= accessibility.StateExpanded
	}
	return accessibility.Props{Role: accessibility.RoleDialog, Name: d.header, State: st}
}

func (s *Select) AccessibleProps() accessibility.Props {
	return accessibility.Props{Role: accessibility.RoleComboBox, Name: s.placeholder, Value: s.value, State: focusedState(s.Base)}
}

func (l *Link) AccessibleProps() accessibility.Props {
	p := accessibility.Props{Role: accessibility.RoleLink, Name: l.text, State: focusedState(l.Base)}
	if l.disabled {
		p.State |= accessibility.StateDisabled
	}
	return p
}

func (a *Avatar) AccessibleProps() accessibility.Props {
	name := a.alt
	if name == "" {
		name = a.content
	}
	return accessibility.Props{Role: accessibility.RoleImage, Name: name}
}

func (g *AvatarGroup) AccessibleProps() accessibility.Props {
	return accessibility.Props{Role: accessibility.RoleGroup, Name: "avatars"}
}

func (b *Badge) AccessibleProps() accessibility.Props {
	name := "badge"
	if elem := b.Element(); elem != nil && elem.TextContent() != "" {
		name = elem.TextContent()
	}
	return accessibility.Props{Role: accessibility.RoleLabel, Name: name}
}

func (b *Breadcrumb) AccessibleProps() accessibility.Props {
	return accessibility.Props{Role: accessibility.RoleGroup, Name: "breadcrumb"}
}

func (i *Image) AccessibleProps() accessibility.Props {
	return accessibility.Props{Role: accessibility.RoleImage, Name: "image"}
}

func (s *Skeleton) AccessibleProps() accessibility.Props {
	return accessibility.Props{Role: accessibility.RoleGroup, Name: "skeleton"}
}

func (s *Space) AccessibleProps() accessibility.Props {
	return accessibility.Props{Role: accessibility.RoleGroup, Name: "space"}
}

func (l *Loading) AccessibleProps() accessibility.Props {
	return accessibility.Props{Role: accessibility.RoleProgressBar, Name: "loading"}
}

func (s *Steps) AccessibleProps() accessibility.Props {
	return accessibility.Props{Role: accessibility.RoleGroup, Name: "steps"}
}

func (t *Text) AccessibleProps() accessibility.Props {
	return accessibility.Props{Role: accessibility.RoleLabel, Name: t.text}
}

func (c *Card) AccessibleProps() accessibility.Props {
	return accessibility.Props{Role: accessibility.RoleGroup, Name: c.title}
}

func (i *Icon) AccessibleProps() accessibility.Props {
	return accessibility.Props{Role: accessibility.RoleImage, Name: i.name}
}

func (c *Canvas) AccessibleProps() accessibility.Props {
	return accessibility.Props{Role: accessibility.RoleImage, Name: "canvas"}
}

func (p *Portal) AccessibleProps() accessibility.Props {
	return accessibility.Props{Role: accessibility.RoleGroup, Name: "portal"}
}

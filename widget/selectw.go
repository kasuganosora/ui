package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// SelectOption represents a single option in a Select dropdown.
type SelectOption struct {
	Label    string
	Value    string
	Disabled bool
}

// Select is a dropdown selector widget.
type Select struct {
	Base
	options         []SelectOption
	value           string   // currently selected value
	selectedValues  []string // for multiple mode
	placeholder     string
	disabled        bool
	open            bool
	size            Size
	status          Status
	clearable       bool
	filterable      bool
	multiple        bool
	borderless      bool
	loading         bool
	readonly        bool
	creatable       bool
	max             int // max selectable in multiple mode, 0 = unlimited
	minCollapsedNum int
	showArrow       bool
	empty           string // empty state text
	inputValue      string // search input value
	popupVisible    bool
	label           string

	optionIDs  []core.ElementID
	backdropID core.ElementID

	onChange             func(value string)
	onBlur               func(value string)
	onFocus              func(value string)
	onClear              func()
	onEnter              func(inputValue string)
	onInputChange        func(inputValue string)
	onPopupVisibleChange func(visible bool)
	onRemove             func(value string)
	onSearch             func(filterWords string)
	onCreate             func(value string)
}

// NewSelect creates a select dropdown.
func NewSelect(tree *core.Tree, options []SelectOption, cfg *Config) *Select {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	s := &Select{
		Base:        NewBase(tree, core.TypeCustom, cfg),
		options:     options,
		placeholder: "请选择",
		showArrow:   true,
	}
	s.style.Display = layout.DisplayFlex
	s.style.AlignItems = layout.AlignCenter
	s.style.Height = layout.Px(cfg.SizeHeight(s.size))
	s.style.Padding = layout.EdgeValues{
		Left:  layout.Px(cfg.SpaceSM),
		Right: layout.Px(cfg.SpaceSM),
	}

	// Toggle dropdown on click
	s.tree.AddHandler(s.id, event.MouseClick, func(e *event.Event) {
		if s.disabled || s.readonly {
			return
		}
		s.open = !s.open
		if s.open {
			s.createOptionElements()
		} else {
			s.destroyOptionElements()
		}
		if s.onPopupVisibleChange != nil {
			s.onPopupVisibleChange(s.open)
		}
	})

	return s
}

func (s *Select) Value() string       { return s.value }
func (s *Select) Placeholder() string { return s.placeholder }
func (s *Select) IsDisabled() bool    { return s.disabled }
func (s *Select) IsOpen() bool        { return s.open }
func (s *Select) Options() []SelectOption { return s.options }

func (s *Select) SetValue(v string) {
	s.value = v
	s.open = false
	s.destroyOptionElements()
}

func (s *Select) SetPlaceholder(p string) { s.placeholder = p }

func (s *Select) SetDisabled(d bool) {
	s.disabled = d
	s.tree.SetEnabled(s.id, !d)
}

func (s *Select) SetOptions(opts []SelectOption) {
	s.options = opts
	s.destroyOptionElements()
}

func (s *Select) OnChange(fn func(string)) { s.onChange = fn }

// SetSize sets the component size, affecting height.
func (s *Select) SetSize(sz Size) {
	s.size = sz
	s.style.Height = layout.Px(s.config.SizeHeight(sz))
}

// SetStatus sets the validation status, affecting border color.
func (s *Select) SetStatus(st Status) { s.status = st }

// SetClearable enables/disables the clear button.
func (s *Select) SetClearable(c bool) { s.clearable = c }

// SetFilterable enables/disables search input in the dropdown.
func (s *Select) SetFilterable(f bool) { s.filterable = f }

// SetMultiple enables/disables multiple selection mode.
func (s *Select) SetMultiple(m bool) { s.multiple = m }

// SelectedValues returns the selected values in multiple mode.
func (s *Select) SelectedValues() []string { return s.selectedValues }

// TDesign additional prop getters/setters
func (s *Select) Borderless() bool              { return s.borderless }
func (s *Select) SetBorderless(v bool)          { s.borderless = v }
func (s *Select) Loading() bool                 { return s.loading }
func (s *Select) SetLoading(v bool)             { s.loading = v }
func (s *Select) Readonly() bool                { return s.readonly }
func (s *Select) SetReadonly(v bool)             { s.readonly = v }
func (s *Select) Creatable() bool               { return s.creatable }
func (s *Select) SetCreatable(v bool)           { s.creatable = v }
func (s *Select) Max() int                      { return s.max }
func (s *Select) SetMax(n int)                  { s.max = n }
func (s *Select) MinCollapsedNum() int          { return s.minCollapsedNum }
func (s *Select) SetMinCollapsedNum(n int)      { s.minCollapsedNum = n }
func (s *Select) ShowArrow() bool               { return s.showArrow }
func (s *Select) SetShowArrow(v bool)           { s.showArrow = v }
func (s *Select) Empty() string                 { return s.empty }
func (s *Select) SetEmpty(e string)             { s.empty = e }
func (s *Select) InputValue() string            { return s.inputValue }
func (s *Select) SetInputValue(v string)        { s.inputValue = v }
func (s *Select) PopupVisible() bool            { return s.popupVisible }
func (s *Select) SetPopupVisible(v bool)        { s.popupVisible = v }
func (s *Select) SelectLabel() string           { return s.label }
func (s *Select) SetLabel(l string)             { s.label = l }

// TDesign event setters
func (s *Select) OnBlur(fn func(string))             { s.onBlur = fn }
func (s *Select) OnFocus(fn func(string))            { s.onFocus = fn }
func (s *Select) OnClear(fn func())                  { s.onClear = fn }
func (s *Select) OnEnter(fn func(string))            { s.onEnter = fn }
func (s *Select) OnInputChange(fn func(string))      { s.onInputChange = fn }
func (s *Select) OnPopupVisibleChange(fn func(bool)) { s.onPopupVisibleChange = fn }
func (s *Select) OnRemove(fn func(string))           { s.onRemove = fn }
func (s *Select) OnSearch(fn func(string))           { s.onSearch = fn }
func (s *Select) OnCreate(fn func(string))           { s.onCreate = fn }

func (s *Select) selectedLabel() string {
	for _, opt := range s.options {
		if opt.Value == s.value {
			return opt.Label
		}
	}
	return ""
}

func (s *Select) createOptionElements() {
	s.destroyOptionElements()
	rootID := s.tree.Root()

	// Create a fullscreen backdrop to capture clicks outside the dropdown
	s.backdropID = s.tree.CreateElement(core.TypeCustom)
	s.tree.AppendChild(rootID, s.backdropID)
	s.tree.SetLayout(s.backdropID, core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 1e6, 1e6),
	})
	s.tree.AddHandler(s.backdropID, event.MouseClick, func(e *event.Event) {
		s.open = false
		s.destroyOptionElements()
	})

	for i, opt := range s.options {
		eid := s.tree.CreateElement(core.TypeCustom)
		s.tree.SetProperty(eid, "text", opt.Label)
		s.tree.AppendChild(rootID, eid)
		s.optionIDs = append(s.optionIDs, eid)

		if !opt.Disabled {
			idx := i
			s.tree.AddHandler(eid, event.MouseClick, func(e *event.Event) {
				s.value = s.options[idx].Value
				s.open = false
				s.destroyOptionElements()
				if s.onChange != nil {
					s.onChange(s.value)
				}
			})
		}
	}
}

func (s *Select) destroyOptionElements() {
	if s.backdropID != core.InvalidElementID {
		s.tree.DestroyElement(s.backdropID)
		s.backdropID = core.InvalidElementID
	}
	for _, eid := range s.optionIDs {
		s.tree.DestroyElement(eid)
	}
	s.optionIDs = nil
}

const selectArrowSize = float32(8)

func (s *Select) Draw(buf *render.CommandBuffer) {
	bounds := s.Bounds()
	if bounds.IsEmpty() {
		return
	}

	cfg := s.config
	elem := s.Element()
	hovered := elem != nil && elem.IsHovered()

	// Border — use StatusBorderColor if status is set
	borderClr := cfg.StatusBorderColor(s.status)
	if s.status == StatusDefault {
		if s.open {
			borderClr = cfg.FocusBorderColor
		} else if hovered {
			borderClr = cfg.HoverColor
		}
	}
	if s.disabled {
		borderClr = cfg.DisabledColor
	}

	bgClr := cfg.BgColor
	if s.disabled {
		bgClr = uimath.ColorHex("#f3f3f3")
	}

	buf.DrawRect(render.RectCmd{
		Bounds:      bounds,
		FillColor:   bgClr,
		BorderColor: borderClr,
		BorderWidth: cfg.BorderWidth,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 0, 1)

	fontSize := cfg.SizeFontSize(s.size)

	// Display text
	label := s.selectedLabel()
	if s.multiple && len(s.selectedValues) > 0 {
		// In multiple mode, show count of selected values
		label = ""
		for i, sv := range s.selectedValues {
			for _, opt := range s.options {
				if opt.Value == sv {
					if i > 0 {
						label += ", "
					}
					label += opt.Label
					break
				}
			}
		}
	}
	textColor := cfg.TextColor
	if label == "" {
		label = s.placeholder
		textColor = cfg.DisabledColor
	}
	if s.disabled {
		textColor = cfg.DisabledColor
	}

	padLeft := cfg.SpaceSM
	arrowArea := selectArrowSize + cfg.SpaceSM
	clearArea := float32(0)
	// Clear button area when clearable, has value, and hovered
	showClear := s.clearable && !s.disabled && s.value != "" && hovered
	if showClear {
		clearArea = selectArrowSize + cfg.SpaceXS
	}
	textMaxW := bounds.Width - padLeft - arrowArea - clearArea

	if label != "" && textMaxW > 0 {
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(fontSize)
			tx := bounds.X + padLeft
			ty := bounds.Y + (bounds.Height-lh)/2
			cfg.TextRenderer.DrawText(buf, label, tx, ty, fontSize, textMaxW, textColor, 1)
		} else {
			textW := float32(len(label)) * fontSize * 0.55
			if textW > textMaxW {
				textW = textMaxW
			}
			textH := fontSize * 1.2
			ty := bounds.Y + (bounds.Height-textH)/2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X+padLeft, ty, textW, textH),
				FillColor: textColor,
				Corners:   uimath.CornersAll(2),
			}, 1, 1)
		}
	}

	// Clear button (X) — shown when clearable, has value, and hovered
	if showClear {
		xSize := float32(8)
		xThick := float32(1.5)
		clearCX := bounds.X + bounds.Width - cfg.SpaceSM - selectArrowSize - cfg.SpaceXS - xSize/2
		clearCY := bounds.Y + bounds.Height/2
		// Draw X as two crossed lines using small rects
		for i := 0; i < 5; i++ {
			t := float32(i) / 4.0
			px := clearCX - xSize/2 + xSize*t
			py1 := clearCY - xSize/2 + xSize*t
			py2 := clearCY + xSize/2 - xSize*t
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(px-xThick/2, py1-xThick/2, xThick, xThick),
				FillColor: cfg.TextColor,
			}, 1, 1)
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(px-xThick/2, py2-xThick/2, xThick, xThick),
				FillColor: cfg.TextColor,
			}, 1, 1)
		}
	}

	// Arrow indicator: down-pointing chevron (V shape with small rects)
	arrowColor := cfg.TextColor
	if s.disabled {
		arrowColor = cfg.DisabledColor
	}
	chevW := float32(8)  // total width of chevron
	chevH := float32(4)  // total height of chevron
	lineW := float32(1.5) // line thickness
	cx := bounds.X + bounds.Width - cfg.SpaceSM - chevW/2 // center X
	cy := bounds.Y + (bounds.Height-chevH)/2              // top Y
	steps := 5
	for i := 0; i <= steps; i++ {
		t := float32(i) / float32(steps)
		// Left leg: top-left to bottom-center
		lx := cx - chevW/2*(1-t)
		ly := cy + chevH*t
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(lx-lineW/2, ly-lineW/2, lineW, lineW),
			FillColor: arrowColor,
		}, 1, 1)
		// Right leg: top-right to bottom-center
		rx := cx + chevW/2*(1-t)
		ry := cy + chevH*t
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(rx-lineW/2, ry-lineW/2, lineW, lineW),
			FillColor: arrowColor,
		}, 1, 1)
	}

	// Dropdown panel
	if s.open {
		s.drawDropdown(buf, bounds)
	}

	s.DrawChildren(buf)
}

func (s *Select) drawDropdown(buf *render.CommandBuffer, triggerBounds uimath.Rect) {
	cfg := s.config
	optH := cfg.InputHeight
	dropH := optH * float32(len(s.options))
	if dropH > 200 {
		dropH = 200
	}
	dropY := triggerBounds.Y + triggerBounds.Height + 4
	dropRect := uimath.NewRect(triggerBounds.X, dropY, triggerBounds.Width, dropH)

	// Draw dropdown as overlay — escapes parent clip regions
	// Shadow
	shadowOffset := float32(2)
	shadowBlur := float32(8)
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(dropRect.X+shadowOffset, dropRect.Y+shadowOffset, dropRect.Width+shadowBlur, dropRect.Height+shadowBlur),
		FillColor: uimath.RGBA(0, 0, 0, 0.12),
		Corners:   uimath.CornersAll(cfg.BorderRadius + 2),
	}, 100, 1)

	// Background
	buf.DrawOverlay(render.RectCmd{
		Bounds:      dropRect,
		FillColor:   uimath.ColorWhite,
		BorderColor: cfg.BorderColor,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 101, 1)

	// Options — also update element bounds for hit testing
	y := dropY
	for i, opt := range s.options {
		optRect := uimath.NewRect(triggerBounds.X, y, triggerBounds.Width, optH)

		// Set layout bounds on option element so hit test can find it
		if i < len(s.optionIDs) {
			s.tree.SetLayout(s.optionIDs[i], core.LayoutResult{
				Bounds: optRect,
			})
		}

		// Hover
		if i < len(s.optionIDs) {
			optElem := s.tree.Get(s.optionIDs[i])
			if optElem != nil && optElem.IsHovered() {
				buf.DrawOverlay(render.RectCmd{
					Bounds:    optRect,
					FillColor: uimath.ColorHex("#f3f3f3"),
				}, 102, 1)
			}
		}

		// Selected indicator
		if opt.Value == s.value {
			buf.DrawOverlay(render.RectCmd{
				Bounds:    optRect,
				FillColor: uimath.ColorHex("#f2f3ff"),
			}, 102, 1)
		}

		// Option label
		textColor := cfg.TextColor
		if opt.Disabled {
			textColor = cfg.DisabledColor
		}
		if opt.Value == s.value {
			textColor = cfg.PrimaryColor
		}

		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			tx := optRect.X + cfg.SpaceSM
			ty := optRect.Y + (optH-lh)/2
			maxW := optRect.Width - cfg.SpaceSM*2
			// Draw text into main commands, then move to overlay
			before := buf.Len()
			cfg.TextRenderer.DrawText(buf, opt.Label, tx, ty, cfg.FontSize, maxW, textColor, 1)
			buf.MoveToOverlay(before, 103)
		} else {
			textW := float32(len(opt.Label)) * cfg.FontSize * 0.55
			maxW := optRect.Width - cfg.SpaceSM*2
			if textW > maxW {
				textW = maxW
			}
			textH := cfg.FontSize * 1.2
			tx := optRect.X + cfg.SpaceSM
			ty := optRect.Y + (optH-textH)/2
			buf.DrawOverlay(render.RectCmd{
				Bounds:    uimath.NewRect(tx, ty, textW, textH),
				FillColor: textColor,
				Corners:   uimath.CornersAll(2),
			}, 103, 1)
		}

		y += optH
	}
}

// Destroy cleans up option elements.
func (s *Select) Destroy() {
	s.destroyOptionElements()
	s.Base.Destroy()
}

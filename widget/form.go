package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// FormLayout determines the form label placement.
type FormLayout uint8

const (
	FormLayoutHorizontal FormLayout = iota
	FormLayoutVertical
)

// Form is a container for FormItem widgets.
type formRule struct {
	item *FormItem
	fn   func() string
}

type Form struct {
	Base
	layout     FormLayout
	labelWidth float32
	rules      []formRule
}

// NewForm creates a form container.
func NewForm(tree *core.Tree, cfg *Config) *Form {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	f := &Form{
		Base:       NewBase(tree, core.TypeCustom, cfg),
		labelWidth: 80,
	}
	f.style.Display = layout.DisplayFlex
	f.style.FlexDirection = layout.FlexDirectionColumn
	f.style.Gap = cfg.SpaceMD
	return f
}

func (f *Form) Layout() FormLayout  { return f.layout }
func (f *Form) LabelWidth() float32 { return f.labelWidth }

func (f *Form) SetLayout(l FormLayout)  { f.layout = l }
func (f *Form) SetLabelWidth(w float32) { f.labelWidth = w }

func (f *Form) Draw(buf *render.CommandBuffer) {
	f.DrawChildren(buf)
}

func (f *Form) AddRule(item *FormItem, fn func() string) {
	if item == nil || fn == nil {
		return
	}
	f.rules = append(f.rules, formRule{item: item, fn: fn})
}

func (f *Form) Validate() bool {
	ok := true
	for _, rule := range f.rules {
		msg := rule.fn()
		rule.item.SetError(msg)
		if msg != "" {
			ok = false
		}
	}
	return ok
}

func (f *Form) Reset() {
	seen := map[*FormItem]bool{}
	for _, rule := range f.rules {
		seen[rule.item] = true
		rule.item.SetError("")
		resetControl(rule.item.Control())
	}
	for _, child := range f.Children() {
		item, ok := child.(*FormItem)
		if !ok || seen[item] {
			continue
		}
		item.SetError("")
		resetControl(item.Control())
	}
}

func resetControl(w Widget) {
	switch c := w.(type) {
	case *Input:
		c.SetValue("")
	case *TextArea:
		c.SetValue("")
	case *Checkbox:
		c.SetChecked(false)
	case *Select:
		c.SetValue("")
	case *Slider:
		c.SetValue(c.Min())
	}
}

// FormItem wraps a label and a control widget.
type FormItem struct {
	Base
	label    string
	required bool
	error    string
	control  Widget
}

// NewFormItem creates a form item with a label and control.
func NewFormItem(tree *core.Tree, label string, control Widget, cfg *Config) *FormItem {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	fi := &FormItem{
		Base:    NewBase(tree, core.TypeCustom, cfg),
		label:   label,
		control: control,
	}
	fi.style.Display = layout.DisplayFlex
	fi.style.AlignItems = layout.AlignCenter
	fi.style.Gap = cfg.SpaceSM

	if control != nil {
		fi.AppendChild(control)
	}

	return fi
}

func (fi *FormItem) Label() string    { return fi.label }
func (fi *FormItem) IsRequired() bool { return fi.required }
func (fi *FormItem) Error() string    { return fi.error }
func (fi *FormItem) Control() Widget  { return fi.control }

func (fi *FormItem) SetLabel(label string) { fi.label = label }
func (fi *FormItem) SetRequired(r bool)    { fi.required = r }
func (fi *FormItem) SetError(err string)   { fi.error = err }

func (fi *FormItem) Draw(buf *render.CommandBuffer) {
	bounds := fi.Bounds()
	if bounds.IsEmpty() {
		return
	}

	cfg := fi.config
	labelW := float32(80) // default label width

	// Draw label
	if fi.label != "" {
		labelColor := cfg.TextColor

		// Required asterisk
		labelText := fi.label
		if fi.required {
			labelText = "* " + labelText
		}

		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			ty := bounds.Y + (bounds.Height-lh)/2
			cfg.TextRenderer.DrawText(buf, labelText, bounds.X, ty, cfg.FontSize, labelW, labelColor, 1)
		} else {
			textW := float32(len(labelText)) * cfg.FontSize * 0.55
			if textW > labelW {
				textW = labelW
			}
			textH := cfg.Leading(cfg.FontSize)
			ty := bounds.Y + (bounds.Height-textH)/2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X, ty, textW, textH),
				FillColor: labelColor,
				Corners:   uimath.CornersAll(2),
			}, 0, 1)
		}

		// Required asterisk in red
		if fi.required && cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			ty := bounds.Y + (bounds.Height-lh)/2
			cfg.TextRenderer.DrawText(buf, "*", bounds.X, ty, cfg.FontSize, 10, cfg.ErrorColor, 1)
		}
	}

	// Error message below
	if fi.error != "" {
		errorY := bounds.Y + bounds.Height + 2
		errorColor := cfg.ErrorColor
		if cfg.TextRenderer != nil {
			cfg.TextRenderer.DrawText(buf, fi.error, bounds.X+labelW+cfg.SpaceSM, errorY, cfg.FontSizeSm, bounds.Width-labelW, errorColor, 1)
		} else {
			textW := float32(len(fi.error)) * cfg.FontSizeSm * 0.55
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(bounds.X+labelW+cfg.SpaceSM, errorY, textW, cfg.Leading(cfg.FontSizeSm)),
				FillColor: errorColor,
				Corners:   uimath.CornersAll(2),
			}, 0, 1)
		}
	}

	fi.DrawChildren(buf)
}

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
type Form struct {
	Base
	layout     FormLayout
	labelWidth float32
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

func (f *Form) Layout() FormLayout     { return f.layout }
func (f *Form) LabelWidth() float32    { return f.labelWidth }

func (f *Form) SetLayout(l FormLayout) { f.layout = l }
func (f *Form) SetLabelWidth(w float32) { f.labelWidth = w }

func (f *Form) Draw(buf *render.CommandBuffer) {
	f.DrawChildren(buf)
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

func (fi *FormItem) Label() string   { return fi.label }
func (fi *FormItem) IsRequired() bool { return fi.required }
func (fi *FormItem) Error() string   { return fi.error }
func (fi *FormItem) Control() Widget { return fi.control }

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
			textH := cfg.FontSize * 1.2
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
				Bounds:    uimath.NewRect(bounds.X+labelW+cfg.SpaceSM, errorY, textW, cfg.FontSizeSm*1.2),
				FillColor: errorColor,
				Corners:   uimath.CornersAll(2),
			}, 0, 1)
		}
	}

	fi.DrawChildren(buf)
}

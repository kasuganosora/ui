// Package ui provides the public API for the GoUI library.
//
// The declarative builder API allows constructing widget trees fluently:
//
//	b := ui.Build(tree, cfg)
//	b.Text("Hello")
//	b.Button("Click me").OnClick(func() { ... })
//	root := b.Widget()
package ui

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/widget"
)

// Builder constructs a widget tree declaratively.
type Builder struct {
	tree   *core.Tree
	config *widget.Config
	root   *widget.Div
}

// Build creates a new builder rooted at a Div container.
func Build(tree *core.Tree, cfg *widget.Config) *Builder {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	root := widget.NewDiv(tree, cfg)
	return &Builder{tree: tree, config: cfg, root: root}
}

// Widget returns the root widget.
func (b *Builder) Widget() *widget.Div {
	return b.root
}

// Style applies a style modifier to the root widget.
func (b *Builder) Style(fn func(s *layout.Style)) *Builder {
	s := b.root.Style()
	fn(&s)
	b.root.SetStyle(s)
	return b
}

// BgColor sets the root background color.
func (b *Builder) BgColor(c uimath.Color) *Builder {
	b.root.SetBgColor(c)
	return b
}

// Div creates a div container and builds its children.
func (b *Builder) Div(children func(b *Builder)) *Builder {
	d := widget.NewDiv(b.tree, b.config)
	child := &Builder{tree: b.tree, config: b.config, root: d}
	if children != nil {
		children(child)
	}
	b.root.AppendChild(d)
	return b
}

// Row creates a horizontal flex container with the given gap.
func (b *Builder) Row(gap float32, children func(b *Builder)) *Builder {
	d := widget.NewDiv(b.tree, b.config)
	s := d.Style()
	s.FlexDirection = layout.FlexDirectionRow
	s.Gap = gap
	s.AlignItems = layout.AlignCenter
	d.SetStyle(s)
	child := &Builder{tree: b.tree, config: b.config, root: d}
	if children != nil {
		children(child)
	}
	b.root.AppendChild(d)
	return b
}

// Column creates a vertical flex container with the given gap.
func (b *Builder) Column(gap float32, children func(b *Builder)) *Builder {
	d := widget.NewDiv(b.tree, b.config)
	s := d.Style()
	s.FlexDirection = layout.FlexDirectionColumn
	s.Gap = gap
	d.SetStyle(s)
	child := &Builder{tree: b.tree, config: b.config, root: d}
	if children != nil {
		children(child)
	}
	b.root.AppendChild(d)
	return b
}

// Text adds a text widget.
func (b *Builder) Text(content string) *TextBuilder {
	t := widget.NewText(b.tree, content, b.config)
	b.root.AppendChild(t)
	return &TextBuilder{text: t}
}

// Button adds a button widget.
func (b *Builder) Button(label string) *ButtonBuilder {
	btn := widget.NewButton(b.tree, label, b.config)
	b.root.AppendChild(btn)
	return &ButtonBuilder{btn: btn}
}

// Input adds an input widget.
func (b *Builder) Input() *InputBuilder {
	inp := widget.NewInput(b.tree, b.config)
	b.root.AppendChild(inp)
	return &InputBuilder{inp: inp}
}

// Checkbox adds a checkbox widget.
func (b *Builder) Checkbox(label string) *CheckboxBuilder {
	cb := widget.NewCheckbox(b.tree, label, b.config)
	b.root.AppendChild(cb)
	return &CheckboxBuilder{cb: cb}
}

// Progress adds a progress bar.
func (b *Builder) Progress(percent float32) *Builder {
	p := widget.NewProgress(b.tree, b.config)
	p.SetPercent(percent)
	b.root.AppendChild(p)
	return b
}

// Custom adds an arbitrary widget.
func (b *Builder) Custom(w widget.Widget) *Builder {
	b.root.AppendChild(w)
	return b
}

// TextBuilder provides fluent methods for Text widget.
type TextBuilder struct {
	text *widget.Text
}

func (tb *TextBuilder) Color(c uimath.Color) *TextBuilder {
	tb.text.SetColor(c)
	return tb
}

func (tb *TextBuilder) FontSize(s float32) *TextBuilder {
	tb.text.SetFontSize(s)
	return tb
}

func (tb *TextBuilder) Widget() *widget.Text { return tb.text }

// ButtonBuilder provides fluent methods for Button widget.
type ButtonBuilder struct {
	btn *widget.Button
}

func (bb *ButtonBuilder) Variant(v widget.ButtonVariant) *ButtonBuilder {
	bb.btn.SetVariant(v)
	return bb
}

func (bb *ButtonBuilder) Disabled(d bool) *ButtonBuilder {
	bb.btn.SetDisabled(d)
	return bb
}

func (bb *ButtonBuilder) OnClick(fn func()) *ButtonBuilder {
	bb.btn.OnClick(fn)
	return bb
}

func (bb *ButtonBuilder) Widget() *widget.Button { return bb.btn }

// InputBuilder provides fluent methods for Input widget.
type InputBuilder struct {
	inp *widget.Input
}

func (ib *InputBuilder) Placeholder(p string) *InputBuilder {
	ib.inp.SetPlaceholder(p)
	return ib
}

func (ib *InputBuilder) Value(v string) *InputBuilder {
	ib.inp.SetValue(v)
	return ib
}

func (ib *InputBuilder) Disabled(d bool) *InputBuilder {
	ib.inp.SetDisabled(d)
	return ib
}

func (ib *InputBuilder) OnChange(fn func(string)) *InputBuilder {
	ib.inp.OnChange(fn)
	return ib
}

func (ib *InputBuilder) Widget() *widget.Input { return ib.inp }

// CheckboxBuilder provides fluent methods for Checkbox widget.
type CheckboxBuilder struct {
	cb *widget.Checkbox
}

func (cb *CheckboxBuilder) Checked(v bool) *CheckboxBuilder {
	cb.cb.SetChecked(v)
	return cb
}

func (cb *CheckboxBuilder) OnChange(fn func(bool)) *CheckboxBuilder {
	cb.cb.OnChange(fn)
	return cb
}

func (cb *CheckboxBuilder) Widget() *widget.Checkbox { return cb.cb }

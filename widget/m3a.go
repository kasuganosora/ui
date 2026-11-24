package widget

import (
	"time"
	"unicode/utf8"

	"github.com/kasuganosora/ui/accessibility"
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Meter shows a 0–1 measurement and optional marks. It is not a Progress.
type Meter struct {
	Base
	value float32
	marks []float32
}

func NewMeter(tree *core.Tree, cfg *Config) *Meter {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Meter{Base: NewBase(tree, core.TypeCustom, cfg)}
}

func (m *Meter) Value() float32 { return m.value }
func (m *Meter) Marks() []float32 {
	out := make([]float32, len(m.marks))
	copy(out, m.marks)
	return out
}
func (m *Meter) SetValue(v float32) {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	m.value = v
}
func (m *Meter) SetMarks(marks []float32) {
	m.marks = append(m.marks[:0], marks...)
}
func (m *Meter) AccessibleProps() accessibility.Props {
	return accessibility.Props{Role: accessibility.RoleProgressBar, Name: "meter", ValueMin: 0, ValueMax: 1, ValueNow: float64(m.value)}
}
func (m *Meter) Draw(buf *render.CommandBuffer) {
	b := m.Bounds()
	if b.IsEmpty() {
		return
	}
	buf.DrawRect(render.RectCmd{Bounds: b, FillColor: m.config.BorderColor}, 1, 1)
	fill := b
	fill.Width *= m.value
	buf.DrawRect(render.RectCmd{Bounds: fill, FillColor: m.config.PrimaryColor}, 2, 1)
	for _, mark := range m.marks {
		x := b.X + b.Width*mark
		buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(x, b.Y, 2, b.Height), FillColor: m.config.TextColor}, 3, 1)
	}
}

// HoldToConfirm fires only after the pointer or Space stays down for Duration.
type HoldToConfirm struct {
	Base
	label    string
	duration time.Duration
	held     float32
	holding  bool
	fired    bool
	onFire   func()
}

func NewHoldToConfirm(tree *core.Tree, label string, cfg *Config) *HoldToConfirm {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	h := &HoldToConfirm{Base: NewBase(tree, core.TypeButton, cfg), label: label, duration: 600 * time.Millisecond}
	tree.SetFocusable(h.id, true)
	tree.AddHandler(h.id, event.MouseDown, func(e *event.Event) {
		if h.disabled() {
			return
		}
		h.holding = true
		h.fired = false
		h.held = 0
	})
	tree.AddHandler(h.id, event.MouseUp, func(e *event.Event) {
		if h.held < 1 {
			h.holding = false
			h.held = 0
		}
	})
	tree.AddHandler(h.id, event.KeyDown, func(e *event.Event) {
		if !targetPhase(e) || e.Key != event.KeySpace || h.disabled() {
			return
		}
		h.holding = true
		h.fired = false
		h.held = 0
		e.StopPropagation()
	})
	tree.AddHandler(h.id, event.KeyUp, func(e *event.Event) {
		if e.Key == event.KeySpace && h.held < 1 {
			h.holding = false
			h.held = 0
		}
	})
	return h
}

func (h *HoldToConfirm) disabled() bool {
	elem := h.tree.Get(h.id)
	return elem != nil && !elem.IsEnabled()
}
func (h *HoldToConfirm) SetDuration(d time.Duration) { h.duration = d }
func (h *HoldToConfirm) Progress() float32           { return h.held }
func (h *HoldToConfirm) OnConfirm(fn func())         { h.onFire = fn }
func (h *HoldToConfirm) Tick(dt float32) {
	if !h.holding || h.fired || h.duration <= 0 {
		return
	}
	h.held += dt / float32(h.duration.Seconds())
	if h.held >= 1 {
		h.held = 1
		h.fired = true
		h.holding = false
		if h.onFire != nil {
			h.onFire()
		}
	}
}
func (h *HoldToConfirm) Draw(buf *render.CommandBuffer) {
	b := h.Bounds()
	if b.IsEmpty() {
		return
	}
	buf.DrawRect(render.RectCmd{Bounds: b, FillColor: h.config.BgColor, BorderColor: h.config.BorderColor, BorderWidth: 1}, 1, 1)
	fill := b
	fill.Width *= h.held
	buf.DrawRect(render.RectCmd{Bounds: fill, FillColor: h.config.PrimaryColor}, 2, 1)
}

// ToggleButton stays selected after a click or Space.
type ToggleButton struct {
	Base
	label    string
	selected bool
	onChange func(bool)
}

func NewToggleButton(tree *core.Tree, label string, cfg *Config) *ToggleButton {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	t := &ToggleButton{Base: NewBase(tree, core.TypeButton, cfg), label: label}
	tree.SetFocusable(t.id, true)
	toggle := func() {
		if t.disabled() {
			return
		}
		t.selected = !t.selected
		if t.onChange != nil {
			t.onChange(t.selected)
		}
	}
	tree.AddHandler(t.id, event.MouseClick, func(e *event.Event) {
		if !targetPhase(e) {
			return
		}
		toggle()
	})
	tree.AddHandler(t.id, event.KeyDown, func(e *event.Event) {
		if !targetPhase(e) || e.Key != event.KeySpace {
			return
		}
		toggle()
		e.StopPropagation()
	})
	return t
}

func (t *ToggleButton) disabled() bool {
	elem := t.tree.Get(t.id)
	return elem != nil && !elem.IsEnabled()
}
func (t *ToggleButton) Label() string          { return t.label }
func (t *ToggleButton) Selected() bool         { return t.selected }
func (t *ToggleButton) SetSelected(v bool)     { t.selected = v }
func (t *ToggleButton) OnChange(fn func(bool)) { t.onChange = fn }
func (t *ToggleButton) AccessibleProps() accessibility.Props {
	p := accessibility.Props{Role: accessibility.RoleButton, Name: t.label}
	if t.selected {
		p.ValueNow = 1
	}
	return p
}
func (t *ToggleButton) Draw(buf *render.CommandBuffer) {
	b := t.Bounds()
	if b.IsEmpty() {
		return
	}
	fill := t.config.BgColor
	if t.selected {
		fill = t.config.PrimaryColor
	}
	buf.DrawRect(render.RectCmd{Bounds: b, FillColor: fill, BorderColor: t.config.BorderColor, BorderWidth: 1}, 1, 1)
}

// ImageCompare splits a frame from 0 to 100. Drag and arrows move it by the pointer or by 5.
type ImageCompare struct {
	Base
	pos    float32
	drag   bool
	before *Image
	after  *Image
}

func NewImageCompare(tree *core.Tree, cfg *Config) *ImageCompare {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	c := &ImageCompare{Base: NewBase(tree, core.TypeCustom, cfg), pos: 50}
	tree.SetFocusable(c.id, true)
	move := func(x float32) {
		b := c.Bounds()
		if b.Width <= 0 || c.disabled() {
			return
		}
		p := (x - b.X) / b.Width * 100
		c.SetPosition(p)
	}
	tree.AddHandler(c.id, event.MouseDown, func(e *event.Event) {
		c.drag = true
		move(e.GlobalX)
	})
	tree.AddHandler(c.id, event.MouseMove, func(e *event.Event) {
		if c.drag {
			move(e.GlobalX)
		}
	})
	tree.AddHandler(c.id, event.MouseUp, func(e *event.Event) { c.drag = false })
	tree.AddHandler(c.id, event.KeyDown, func(e *event.Event) {
		if !targetPhase(e) || c.disabled() {
			return
		}
		switch e.Key {
		case event.KeyArrowLeft:
			c.SetPosition(c.pos - 5)
		case event.KeyArrowRight:
			c.SetPosition(c.pos + 5)
		default:
			return
		}
		e.StopPropagation()
	})
	return c
}

func (c *ImageCompare) disabled() bool {
	elem := c.tree.Get(c.id)
	return elem != nil && !elem.IsEnabled()
}
func (c *ImageCompare) SetBefore(img *Image) { c.before = img }
func (c *ImageCompare) SetAfter(img *Image)  { c.after = img }
func (c *ImageCompare) Before() *Image       { return c.before }
func (c *ImageCompare) After() *Image        { return c.after }
func (c *ImageCompare) Position() float32    { return c.pos }
func (c *ImageCompare) SetPosition(p float32) {
	if p < 0 {
		p = 0
	}
	if p > 100 {
		p = 100
	}
	c.pos = p
}
func (c *ImageCompare) Draw(buf *render.CommandBuffer) {
	b := c.Bounds()
	if b.IsEmpty() {
		return
	}
	if c.before != nil || c.after != nil {
		split := b.Width * c.pos / 100
		if c.before != nil && c.before.Texture() != render.InvalidTexture {
			buf.PushClip(uimath.NewRect(b.X, b.Y, split, b.Height))
			buf.DrawImage(render.ImageCmd{Texture: c.before.Texture(), DstRect: b, Tint: uimath.ColorWhite}, 1, 1)
			buf.PopClip()
		}
		if c.after != nil && c.after.Texture() != render.InvalidTexture {
			buf.PushClip(uimath.NewRect(b.X+split, b.Y, b.Width-split, b.Height))
			buf.DrawImage(render.ImageCmd{Texture: c.after.Texture(), DstRect: b, Tint: uimath.ColorWhite}, 1, 1)
			buf.PopClip()
		}
		buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(b.X+split-1, b.Y, 2, b.Height), FillColor: uimath.ColorWhite}, 3, 1)
		return
	}
	buf.DrawRect(render.RectCmd{Bounds: b, FillColor: c.config.BorderColor}, 1, 1)
	left := b
	left.Width = b.Width * c.pos / 100
	buf.DrawRect(render.RectCmd{Bounds: left, FillColor: c.config.PrimaryColor}, 2, 1)
}

// OtpInput is a fixed-length code field.
type OtpInput struct {
	Base
	length   int
	value    string
	onDone   func(string)
	disabled bool
}

func NewOtpInput(tree *core.Tree, length int, cfg *Config) *OtpInput {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	if length < 1 {
		length = 4
	}
	o := &OtpInput{Base: NewBase(tree, core.TypeInput, cfg), length: length}
	tree.SetFocusable(o.id, true)
	tree.AddHandler(o.id, event.KeyPress, func(e *event.Event) {
		if !targetPhase(e) || o.disabled {
			return
		}
		if e.Text != "" {
			o.insert(e.Text)
			e.StopPropagation()
			return
		}
		if e.Char == 0 {
			return
		}
		o.insert(string(e.Char))
		e.StopPropagation()
	})
	tree.AddHandler(o.id, event.KeyDown, func(e *event.Event) {
		if !targetPhase(e) || o.disabled {
			return
		}
		if e.Text != "" && utf8.RuneCountInString(e.Text) > 1 {
			o.insert(e.Text)
			e.StopPropagation()
			return
		}
		if e.Key == event.KeyBackspace {
			if o.value != "" {
				_, size := utf8.DecodeLastRuneInString(o.value)
				o.value = o.value[:len(o.value)-size]
			}
			e.StopPropagation()
		}
	})
	tree.AddHandler(o.id, event.IMECompositionEnd, func(e *event.Event) {
		if o.disabled {
			return
		}
		o.insert(e.Text)
	})
	return o
}

func (o *OtpInput) insert(s string) {
	if utf8.RuneCountInString(o.value) >= o.length {
		return
	}
	for _, r := range s {
		if utf8.RuneCountInString(o.value) >= o.length {
			break
		}
		o.value += string(r)
	}
	if utf8.RuneCountInString(o.value) >= o.length && o.onDone != nil {
		o.onDone(o.value)
	}
}
func (o *OtpInput) Length() int                { return o.length }
func (o *OtpInput) Value() string              { return o.value }
func (o *OtpInput) SetValue(v string)          { o.value = v }
func (o *OtpInput) OnComplete(fn func(string)) { o.onDone = fn }
func (o *OtpInput) SetDisabled(d bool) {
	o.disabled = d
	o.tree.SetEnabled(o.id, !d)
	o.tree.SetFocusable(o.id, !d)
}
func (o *OtpInput) AccessibleProps() accessibility.Props {
	return accessibility.Props{Role: accessibility.RoleTextBox, Name: "otp", ValueNow: float64(utf8.RuneCountInString(o.value))}
}
func (o *OtpInput) Draw(buf *render.CommandBuffer) {
	b := o.Bounds()
	if b.IsEmpty() {
		return
	}
	cell := b.Width / float32(o.length)
	for i := 0; i < o.length; i++ {
		buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(b.X+float32(i)*cell+2, b.Y, cell-4, b.Height), BorderColor: o.config.BorderColor, BorderWidth: 1, FillColor: o.config.BgColor}, 1, 1)
	}
}

// Scrollbar maps a thumb drag to a scroll offset.
type Scrollbar struct {
	Base
	offset   float32
	view     float32
	content  float32
	dragY    float32
	dragOff  float32
	dragging bool
}

func NewScrollbar(tree *core.Tree, cfg *Config) *Scrollbar {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	s := &Scrollbar{Base: NewBase(tree, core.TypeScroll, cfg), view: 100, content: 100}
	tree.SetFocusable(s.id, true)
	tree.AddHandler(s.id, event.MouseDown, func(e *event.Event) {
		if s.disabled() {
			return
		}
		s.dragging = true
		s.dragY = e.GlobalY
		s.dragOff = s.offset
	})
	tree.AddHandler(s.id, event.MouseMove, func(e *event.Event) {
		if !s.dragging {
			return
		}
		b := s.Bounds()
		span := b.Height
		if span < 1 {
			span = 1
		}
		max := s.content - s.view
		if max < 0 {
			max = 0
		}
		s.SetOffset(s.dragOff + (e.GlobalY-s.dragY)/span*max)
	})
	tree.AddHandler(s.id, event.MouseUp, func(e *event.Event) { s.dragging = false })
	return s
}

func (s *Scrollbar) disabled() bool {
	elem := s.tree.Get(s.id)
	return elem != nil && !elem.IsEnabled()
}
func (s *Scrollbar) Offset() float32  { return s.offset }
func (s *Scrollbar) Content() float32 { return s.content }
func (s *Scrollbar) View() float32    { return s.view }
func (s *Scrollbar) SetMetrics(content, view float32) {
	s.content = content
	s.view = view
	s.SetOffset(s.offset)
}
func (s *Scrollbar) SetOffset(v float32) {
	max := s.content - s.view
	if max < 0 {
		max = 0
	}
	if v < 0 {
		v = 0
	}
	if v > max {
		v = max
	}
	s.offset = v
}
func (s *Scrollbar) AccessibleProps() accessibility.Props {
	return accessibility.Props{Role: accessibility.RoleScrollBar, Name: "scrollbar", ValueNow: float64(s.offset), ValueMax: float64(s.content - s.view)}
}
func (s *Scrollbar) Draw(buf *render.CommandBuffer) {
	b := s.Bounds()
	if b.IsEmpty() {
		return
	}
	buf.DrawRect(render.RectCmd{Bounds: b, FillColor: s.config.BorderColor}, 1, 1)
	max := s.content - s.view
	if max < 1 {
		max = 1
	}
	thumbH := b.Height * (s.view / s.content)
	if s.content < 1 {
		thumbH = b.Height
	}
	if thumbH < 12 {
		thumbH = 12
	}
	y := b.Y + (b.Height-thumbH)*(s.offset/max)
	buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(b.X, y, b.Width, thumbH), FillColor: s.config.PrimaryColor}, 2, 1)
}

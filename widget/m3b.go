package widget

import (
	"strings"

	"github.com/kasuganosora/ui/accessibility"
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// ToggleGroup is a single or multiple selection control, separate from RadioGroup.
type ToggleGroup struct {
	Base
	labels   []string
	values   []string
	multi    bool
	index    int
	picked   map[string]bool
	value    string
	onChange func()
}

func NewToggleGroup(tree *core.Tree, cfg *Config) *ToggleGroup {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	g := &ToggleGroup{Base: NewBase(tree, core.TypeCustom, cfg), picked: map[string]bool{}, index: -1}
	tree.SetFocusable(g.id, true)
	tree.AddHandler(g.id, event.KeyDown, func(e *event.Event) {
		if !targetPhase(e) || len(g.values) == 0 || g.disabled() {
			return
		}
		switch e.Key {
		case event.KeyArrowRight, event.KeyArrowDown:
			g.move(1)
		case event.KeyArrowLeft, event.KeyArrowUp:
			g.move(-1)
		case event.KeySpace, event.KeyEnter:
			if g.multi {
				g.toggle(g.index)
			}
		default:
			return
		}
		e.StopPropagation()
	})
	tree.AddHandler(g.id, event.MouseClick, func(e *event.Event) {
		if !targetPhase(e) || g.disabled() {
			return
		}
		b := g.Bounds()
		if len(g.values) == 0 || b.Width <= 0 {
			return
		}
		i := int((e.GlobalX - b.X) / (b.Width / float32(len(g.values))))
		if i < 0 || i >= len(g.values) {
			return
		}
		g.index = i
		if g.multi {
			g.toggle(i)
		} else {
			g.value = g.values[i]
			if g.onChange != nil {
				g.onChange()
			}
		}
	})
	return g
}

func (g *ToggleGroup) disabled() bool {
	elem := g.tree.Get(g.id)
	return elem != nil && !elem.IsEnabled()
}
func (g *ToggleGroup) Add(label, value string) {
	g.labels = append(g.labels, label)
	g.values = append(g.values, value)
	if g.index < 0 {
		g.index = 0
		if !g.multi {
			g.value = value
		}
	}
}
func (g *ToggleGroup) SetMultiple(v bool) { g.multi = v }
func (g *ToggleGroup) Multiple() bool     { return g.multi }
func (g *ToggleGroup) Value() string      { return g.value }
func (g *ToggleGroup) Values() []string {
	var out []string
	for _, v := range g.values {
		if g.picked[v] {
			out = append(out, v)
		}
	}
	return out
}
func (g *ToggleGroup) Index() int         { return g.index }
func (g *ToggleGroup) OnChange(fn func()) { g.onChange = fn }
func (g *ToggleGroup) move(dir int) {
	g.index = (g.index + dir + len(g.values)) % len(g.values)
	if !g.multi {
		g.value = g.values[g.index]
	}
	if g.onChange != nil {
		g.onChange()
	}
}
func (g *ToggleGroup) toggle(i int) {
	if i < 0 || i >= len(g.values) {
		return
	}
	v := g.values[i]
	g.picked[v] = !g.picked[v]
	if g.onChange != nil {
		g.onChange()
	}
}
func (g *ToggleGroup) Draw(buf *render.CommandBuffer) {
	b := g.Bounds()
	if b.IsEmpty() || len(g.values) == 0 {
		return
	}
	w := b.Width / float32(len(g.values))
	for i := range g.values {
		fill := g.config.BgColor
		on := g.value == g.values[i] || g.picked[g.values[i]]
		if on {
			fill = g.config.PrimaryColor
		}
		buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(b.X+float32(i)*w, b.Y, w-2, b.Height), FillColor: fill, BorderColor: g.config.BorderColor, BorderWidth: 1}, 1, 1)
	}
}

// Toolbar keeps one tab stop. Arrows move the item and do not activate it.
type Toolbar struct {
	Base
	labels []string
	clicks []int
	index  int
}

func NewToolbar(tree *core.Tree, cfg *Config) *Toolbar {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	t := &Toolbar{Base: NewBase(tree, core.TypeCustom, cfg)}
	tree.SetFocusable(t.id, true)
	tree.AddHandler(t.id, event.KeyDown, func(e *event.Event) {
		if !targetPhase(e) || len(t.labels) == 0 {
			return
		}
		switch e.Key {
		case event.KeyArrowRight, event.KeyArrowDown:
			t.index = (t.index + 1) % len(t.labels)
		case event.KeyArrowLeft, event.KeyArrowUp:
			t.index = (t.index - 1 + len(t.labels)) % len(t.labels)
		case event.KeyEnter, event.KeySpace:
			t.activate(t.index)
		default:
			return
		}
		e.StopPropagation()
	})
	tree.AddHandler(t.id, event.MouseClick, func(e *event.Event) {
		if !targetPhase(e) || len(t.labels) == 0 {
			return
		}
		b := t.Bounds()
		w := b.Width / float32(len(t.labels))
		if w <= 0 {
			return
		}
		i := int((e.GlobalX - b.X) / w)
		if i >= 0 && i < len(t.labels) {
			t.index = i
			t.activate(i)
		}
	})
	return t
}

func (t *Toolbar) Add(label string) {
	t.labels = append(t.labels, label)
	t.clicks = append(t.clicks, 0)
}
func (t *Toolbar) Index() int { return t.index }
func (t *Toolbar) Clicks(i int) int {
	if i < 0 || i >= len(t.clicks) {
		return 0
	}
	return t.clicks[i]
}
func (t *Toolbar) activate(i int) {
	if i >= 0 && i < len(t.clicks) {
		t.clicks[i]++
	}
}
func (t *Toolbar) AccessibleProps() accessibility.Props {
	return accessibility.Props{Role: accessibility.RoleToolbar, Name: "toolbar"}
}
func (t *Toolbar) Draw(buf *render.CommandBuffer) {
	b := t.Bounds()
	if b.IsEmpty() {
		return
	}
	buf.DrawRect(render.RectCmd{Bounds: b, FillColor: t.config.BgColor}, 1, 1)
}

// HoverCard opens after a hover delay, or immediately when focused.
type HoverCard struct {
	Base
	text       string
	delay      float32
	pending    float32
	hover      bool
	open       bool
	cardID     core.ElementID
	backdropID core.ElementID
}

func NewHoverCard(tree *core.Tree, cfg *Config) *HoverCard {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	h := &HoverCard{Base: NewBase(tree, core.TypeCustom, cfg), delay: 0.15, text: "详情"}
	tree.SetFocusable(h.id, true)
	tree.AddHandler(h.id, event.MouseEnter, func(e *event.Event) {
		h.hover = true
		h.pending = 0
	})
	tree.AddHandler(h.id, event.MouseLeave, func(e *event.Event) {
		h.hover = false
		h.pending = 0
	})
	tree.AddHandler(h.id, event.Focus, func(e *event.Event) { h.SetOpen(true) })
	tree.AddHandler(h.id, event.KeyDown, func(e *event.Event) {
		if !targetPhase(e) || e.Key != event.KeyEscape || !h.open {
			return
		}
		h.SetOpen(false)
		e.StopPropagation()
	})
	return h
}

func (h *HoverCard) SetDelay(seconds float32) { h.delay = seconds }
func (h *HoverCard) SetText(s string)         { h.text = s }
func (h *HoverCard) IsOpen() bool             { return h.open }
func (h *HoverCard) Tick(dt float32) {
	if !h.hover || h.open {
		return
	}
	h.pending += dt
	if h.pending >= h.delay {
		h.SetOpen(true)
	}
}
func (h *HoverCard) SetOpen(v bool) {
	if v == h.open {
		return
	}
	h.open = v
	h.clearPop()
	if !v {
		h.tree.PopTransient(h.id)
		return
	}
	h.tree.PushTransient(h.id)
	h.backdropID = h.tree.CreateElement(core.TypeCustom)
	h.tree.AppendChild(h.id, h.backdropID)
	h.tree.SetOverlay(h.backdropID, true)
	h.tree.SetProperty(h.backdropID, "dismiss", true)
	h.tree.SetLayout(h.backdropID, core.LayoutResult{Bounds: uimath.NewRect(0, 0, 1e6, 1e6)})
	h.tree.AddHandler(h.backdropID, event.MouseClick, func(e *event.Event) {
		if !targetPhase(e) {
			return
		}
		e.StopPropagation()
		h.SetOpen(false)
	})
	h.cardID = h.tree.CreateElement(core.TypeCustom)
	h.tree.AppendChild(h.id, h.cardID)
	h.tree.SetOverlay(h.cardID, true)
	b := h.Bounds()
	h.tree.SetLayout(h.cardID, core.LayoutResult{Bounds: uimath.NewRect(b.X, b.Y+b.Height+4, 160, 64)})
}
func (h *HoverCard) clearPop() {
	if h.cardID != core.InvalidElementID {
		h.tree.DestroyElement(h.cardID)
		h.cardID = core.InvalidElementID
	}
	if h.backdropID != core.InvalidElementID {
		h.tree.DestroyElement(h.backdropID)
		h.backdropID = core.InvalidElementID
	}
}
func (h *HoverCard) Destroy() {
	h.tree.PopTransient(h.id)
	h.clearPop()
	h.open = false
	h.Base.Destroy()
}
func (h *HoverCard) Draw(buf *render.CommandBuffer) {
	defer withLayer(buf, h.tree, h.id)()
	b := h.Bounds()
	if b.IsEmpty() {
		return
	}
	buf.DrawRect(render.RectCmd{Bounds: b, FillColor: h.config.BgColor, BorderColor: h.config.BorderColor, BorderWidth: 1}, 1, 1)
	if h.open {
		card := uimath.NewRect(b.X, b.Y+b.Height+4, 160, 64)
		buf.DrawOverlay(render.RectCmd{Bounds: card, FillColor: uimath.ColorWhite, BorderColor: h.config.BorderColor, BorderWidth: 1}, 20, 1)
	}
}

// Combobox is an editable field plus a suggestion list. It does not replace Select or AutoComplete.
type Combobox struct {
	Base
	options []string
	text    string
	value   string
	open    bool
	index   int
	opts    []core.ElementID
}

func NewCombobox(tree *core.Tree, options []string, cfg *Config) *Combobox {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	c := &Combobox{Base: NewBase(tree, core.TypeInput, cfg), options: append([]string(nil), options...)}
	tree.SetFocusable(c.id, true)
	tree.AddHandler(c.id, event.KeyPress, func(e *event.Event) {
		if !targetPhase(e) || c.disabled() || e.Char == 0 {
			return
		}
		c.text += string(e.Char)
		c.openList()
		e.StopPropagation()
	})
	tree.AddHandler(c.id, event.KeyDown, func(e *event.Event) {
		if !targetPhase(e) || c.disabled() {
			return
		}
		shown := c.shown()
		switch e.Key {
		case event.KeyBackspace:
			if c.text != "" {
				c.text = c.text[:len(c.text)-1]
			}
			c.openList()
		case event.KeyArrowDown:
			if !c.open {
				c.openList()
			} else if len(shown) > 0 {
				c.index = (c.index + 1) % len(shown)
			}
		case event.KeyArrowUp:
			if len(shown) > 0 {
				c.index = (c.index - 1 + len(shown)) % len(shown)
			}
		case event.KeyEnter:
			if c.open && len(shown) > 0 {
				c.choose(shown[c.index])
			}
		case event.KeyEscape:
			if !c.open {
				return
			}
			c.closeList()
		default:
			return
		}
		e.StopPropagation()
	})
	return c
}

func (c *Combobox) disabled() bool {
	elem := c.tree.Get(c.id)
	return elem != nil && !elem.IsEnabled()
}
func (c *Combobox) Text() string  { return c.text }
func (c *Combobox) Value() string { return c.value }
func (c *Combobox) IsOpen() bool  { return c.open }
func (c *Combobox) shown() []string {
	q := strings.ToLower(c.text)
	var out []string
	for _, opt := range c.options {
		if q == "" || strings.Contains(strings.ToLower(opt), q) {
			out = append(out, opt)
		}
	}
	return out
}
func (c *Combobox) openList() {
	c.closeList()
	c.open = true
	b := c.Bounds()
	y := b.Y + b.Height + 2
	for i, opt := range c.shown() {
		id := c.tree.CreateElement(core.TypeCustom)
		c.tree.AppendChild(c.id, id)
		c.tree.SetOverlay(id, true)
		c.tree.SetProperty(id, "text", opt)
		c.tree.SetLayout(id, core.LayoutResult{Bounds: uimath.NewRect(b.X, y+float32(i)*28, b.Width, 28)})
		choice := opt
		c.tree.AddHandler(id, event.MouseClick, func(e *event.Event) {
			if !targetPhase(e) {
				return
			}
			e.StopPropagation()
			c.choose(choice)
		})
		c.opts = append(c.opts, id)
	}
}
func (c *Combobox) closeList() {
	c.open = false
	for _, id := range c.opts {
		c.tree.DestroyElement(id)
	}
	c.opts = nil
}
func (c *Combobox) choose(v string) {
	c.value = v
	c.text = v
	c.closeList()
}
func (c *Combobox) Destroy() {
	c.closeList()
	c.Base.Destroy()
}
func (c *Combobox) AccessibleProps() accessibility.Props {
	return accessibility.Props{Role: accessibility.RoleComboBox, Name: c.text}
}
func (c *Combobox) Draw(buf *render.CommandBuffer) {
	defer withLayer(buf, c.tree, c.id)()
	b := c.Bounds()
	if b.IsEmpty() {
		return
	}
	buf.DrawRect(render.RectCmd{Bounds: b, FillColor: c.config.BgColor, BorderColor: c.config.BorderColor, BorderWidth: 1}, 1, 1)
}

// SuggestList anchors to a caret rectangle and does not own the input value.
type SuggestList struct {
	Base
	items  []string
	index  int
	open   bool
	anchor uimath.Rect
	onPick func(string)
	rows   []core.ElementID
}

func NewSuggestList(tree *core.Tree, cfg *Config) *SuggestList {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	s := &SuggestList{Base: NewBase(tree, core.TypeCustom, cfg)}
	tree.SetFocusable(s.id, true)
	tree.AddHandler(s.id, event.KeyDown, func(e *event.Event) {
		if !targetPhase(e) || !s.open {
			return
		}
		switch e.Key {
		case event.KeyArrowDown:
			if len(s.items) > 0 {
				s.index = (s.index + 1) % len(s.items)
			}
		case event.KeyArrowUp:
			if len(s.items) > 0 {
				s.index = (s.index - 1 + len(s.items)) % len(s.items)
			}
		case event.KeyEnter:
			if len(s.items) > 0 && s.onPick != nil {
				s.onPick(s.items[s.index])
			}
			s.Close()
		case event.KeyEscape:
			s.Close()
		default:
			return
		}
		e.StopPropagation()
	})
	return s
}

func (s *SuggestList) SetItems(items []string) {
	s.items = append(s.items[:0], items...)
	s.index = 0
	if s.open {
		s.place()
	}
}
func (s *SuggestList) SetAnchor(x, y, w, h float32) {
	s.anchor = uimath.NewRect(x, y, w, h)
}
func (s *SuggestList) OnSelect(fn func(string)) { s.onPick = fn }
func (s *SuggestList) Index() int               { return s.index }
func (s *SuggestList) IsOpen() bool             { return s.open }
func (s *SuggestList) Open() {
	s.open = true
	s.place()
}
func (s *SuggestList) Close() {
	s.open = false
	for _, id := range s.rows {
		s.tree.DestroyElement(id)
	}
	s.rows = nil
}
func (s *SuggestList) place() {
	s.Close()
	s.open = true
	for i, item := range s.items {
		id := s.tree.CreateElement(core.TypeCustom)
		s.tree.AppendChild(s.id, id)
		s.tree.SetOverlay(id, true)
		s.tree.SetProperty(id, "text", item)
		s.tree.SetLayout(id, core.LayoutResult{Bounds: uimath.NewRect(s.anchor.X, s.anchor.Y+s.anchor.Height+2+float32(i)*24, s.anchor.Width, 24)})
		choice := item
		s.tree.AddHandler(id, event.MouseClick, func(e *event.Event) {
			if !targetPhase(e) {
				return
			}
			e.StopPropagation()
			if s.onPick != nil {
				s.onPick(choice)
			}
			s.Close()
		})
		s.rows = append(s.rows, id)
	}
}
func (s *SuggestList) Destroy() {
	s.Close()
	s.Base.Destroy()
}
func (s *SuggestList) Draw(buf *render.CommandBuffer) {
	if !s.open {
		return
	}
	defer withLayer(buf, s.tree, s.id)()
	for i := range s.items {
		r := uimath.NewRect(s.anchor.X, s.anchor.Y+s.anchor.Height+2+float32(i)*24, s.anchor.Width, 24)
		buf.DrawOverlay(render.RectCmd{Bounds: r, FillColor: uimath.ColorWhite, BorderColor: s.config.BorderColor, BorderWidth: 1}, 30, 1)
	}
}

// Command filters a command list and runs the highlighted item.
type Command struct {
	Base
	labels []string
	ids    []string
	query  string
	index  int
	onRun  func(string)
	rows   []core.ElementID
}

func NewCommand(tree *core.Tree, cfg *Config) *Command {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	c := &Command{Base: NewBase(tree, core.TypeCustom, cfg)}
	tree.SetFocusable(c.id, true)
	tree.AddHandler(c.id, event.KeyPress, func(e *event.Event) {
		if !targetPhase(e) || e.Char == 0 {
			return
		}
		c.query += string(e.Char)
		c.index = 0
		c.place()
		e.StopPropagation()
	})
	tree.AddHandler(c.id, event.KeyDown, func(e *event.Event) {
		if !targetPhase(e) {
			return
		}
		shown := c.Shown()
		switch e.Key {
		case event.KeyBackspace:
			if c.query != "" {
				c.query = c.query[:len(c.query)-1]
				c.index = 0
				c.place()
			}
		case event.KeyArrowDown:
			if len(shown) > 0 {
				c.index = (c.index + 1) % len(shown)
				c.place()
			}
		case event.KeyArrowUp:
			if len(shown) > 0 {
				c.index = (c.index - 1 + len(shown)) % len(shown)
				c.place()
			}
		case event.KeyEnter:
			if len(shown) > 0 && c.onRun != nil {
				c.onRun(shown[c.index].id)
			}
		case event.KeyEscape:
			return
		default:
			return
		}
		e.StopPropagation()
	})
	return c
}

type commandHit struct{ label, id string }

func (c *Command) Add(label, id string) {
	c.labels = append(c.labels, label)
	c.ids = append(c.ids, id)
	c.place()
}
func (c *Command) Query() string         { return c.query }
func (c *Command) OnRun(fn func(string)) { c.onRun = fn }
func (c *Command) Shown() []commandHit {
	q := strings.ToLower(c.query)
	var out []commandHit
	for i, label := range c.labels {
		if q == "" || strings.Contains(strings.ToLower(label), q) {
			out = append(out, commandHit{label: label, id: c.ids[i]})
		}
	}
	return out
}
func (c *Command) place() {
	for _, id := range c.rows {
		c.tree.DestroyElement(id)
	}
	c.rows = nil
	start, end, items := c.window()
	vis := 0
	for i := start; i < end; i++ {
		id := c.tree.CreateElement(core.TypeCustom)
		c.tree.AppendChild(c.id, id)
		c.tree.SetOverlay(id, true)
		c.tree.SetProperty(id, "text", items[i].label)
		c.tree.SetLayout(id, core.LayoutResult{Bounds: c.rowBounds(vis)})
		runID := items[i].id
		c.tree.AddHandler(id, event.MouseClick, func(e *event.Event) {
			if !targetPhase(e) {
				return
			}
			e.StopPropagation()
			if c.onRun != nil {
				c.onRun(runID)
			}
		})
		c.rows = append(c.rows, id)
		vis++
	}
}

func (c *Command) window() (start, end int, items []commandHit) {
	items = c.Shown()
	const n = 6
	if len(items) == 0 {
		c.index = 0
		return 0, 0, items
	}
	if c.index >= len(items) {
		c.index = len(items) - 1
	}
	if c.index < 0 {
		c.index = 0
	}
	start = 0
	if c.index >= n {
		start = c.index - n + 1
	}
	end = start + n
	if end > len(items) {
		end = len(items)
	}
	return start, end, items
}
func (c *Command) Destroy() {
	for _, id := range c.rows {
		c.tree.DestroyElement(id)
	}
	c.Base.Destroy()
}
func (c *Command) rowBounds(i int) uimath.Rect {
	b := c.Bounds()
	return uimath.NewRect(b.X, b.Y+32+float32(i)*28, b.Width, 28)
}

func (c *Command) Draw(buf *render.CommandBuffer) {
	b := c.Bounds()
	if b.IsEmpty() {
		return
	}
	buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(b.X, b.Y, b.Width, 28), FillColor: c.config.BgColor, BorderColor: c.config.BorderColor, BorderWidth: 1}, 1, 1)
	start, end, _ := c.window()
	vis := 0
	for i := start; i < end; i++ {
		fill := uimath.ColorWhite
		if i == c.index {
			fill = c.config.PrimaryColor
		}
		buf.DrawOverlay(render.RectCmd{Bounds: c.rowBounds(vis), FillColor: fill, BorderColor: c.config.BorderColor, BorderWidth: 1}, 40, 1)
		vis++
	}
}

// CommandDialog shows a Command in a modal layer and restores focus on Escape.
type CommandDialog struct {
	Base
	cmd    *Command
	open   bool
	prior  core.ElementID
	scoped bool
}

func NewCommandDialog(tree *core.Tree, cfg *Config) *CommandDialog {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	d := &CommandDialog{Base: NewBase(tree, core.TypeCustom, cfg)}
	d.cmd = NewCommand(tree, cfg)
	d.AppendChild(d.cmd)
	tree.SetVisible(d.id, false)
	tree.AddHandler(d.id, event.KeyDown, func(e *event.Event) {
		if e.Phase == event.PhaseCapture || e.Key != event.KeyEscape || !d.open {
			return
		}
		d.Close()
		e.StopPropagation()
	})
	return d
}

func (d *CommandDialog) Command() *Command { return d.cmd }
func (d *CommandDialog) IsOpen() bool      { return d.open }
func (d *CommandDialog) Open() {
	if !d.open {
		d.prior = d.tree.Focused()
		d.tree.PushFocusScope(d.id)
		d.scoped = true
	}
	d.open = true
	d.tree.SetVisible(d.id, true)
	d.tree.PushLayer(d.id)
	if elem := d.tree.Get(d.id); elem != nil && elem.ParentID() != core.InvalidElementID {
		d.tree.BringChildToFront(elem.ParentID(), d.id)
	}
	d.tree.SetFocusable(d.cmd.ElementID(), true)
	d.tree.SetFocused(d.cmd.ElementID(), true)
	d.cmd.place()
}
func (d *CommandDialog) Close() {
	if !d.open && !d.scoped {
		return
	}
	inside := d.tree.InScope(d.tree.Focused(), d.id)
	d.open = false
	d.tree.SetVisible(d.id, false)
	d.tree.PopLayer(d.id)
	if d.scoped {
		d.tree.PopFocusScopeOwned(d.id)
		d.scoped = false
	}
	if inside {
		scope := d.tree.FocusScope()
		if d.prior != core.InvalidElementID && d.tree.Get(d.prior) != nil && d.tree.InScope(d.prior, scope) && d.tree.ChainAllows(d.prior) {
			d.tree.SetFocused(d.prior, true)
		} else if scope != core.InvalidElementID {
			d.tree.FocusNext(core.InvalidElementID)
		}
	}
}
func (d *CommandDialog) Destroy() {
	d.Close()
	d.Base.Destroy()
}
func (d *CommandDialog) AccessibleProps() accessibility.Props {
	return accessibility.Props{Role: accessibility.RoleDialog, Name: "command"}
}
func (d *CommandDialog) Draw(buf *render.CommandBuffer) {
	if !d.open {
		return
	}
	defer withLayer(buf, d.tree, d.id)()
	b := d.Bounds()
	if b.IsEmpty() {
		return
	}
	buf.DrawRect(render.RectCmd{Bounds: b, FillColor: uimath.Color{R: 0, G: 0, B: 0, A: 0.45}}, 10, 1)
	d.cmd.Draw(buf)
}

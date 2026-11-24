package widget

import (
	"github.com/kasuganosora/ui/accessibility"
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// NavHeader shows the title of the current navigation page.
type NavHeader struct {
	Base
	title string
}

func NewNavHeader(tree *core.Tree, cfg *Config) *NavHeader {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &NavHeader{Base: NewBase(tree, core.TypeCustom, cfg)}
}
func (h *NavHeader) Title() string     { return h.title }
func (h *NavHeader) SetTitle(s string) { h.title = s }
func (h *NavHeader) Draw(buf *render.CommandBuffer) {
	b := h.Bounds()
	if b.IsEmpty() {
		return
	}
	buf.DrawRect(render.RectCmd{Bounds: b, FillColor: h.config.BgColor}, 1, 1)
}

type navPage struct {
	title string
	body  Widget
}

// Nav pushes and pops pages. Backspace and the back button pop. The header follows the top page.
type Nav struct {
	Base
	header *NavHeader
	stack  []navPage
	backID core.ElementID
}

func NewNav(tree *core.Tree, cfg *Config) *Nav {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	n := &Nav{Base: NewBase(tree, core.TypeCustom, cfg), header: NewNavHeader(tree, cfg)}
	n.AppendChild(n.header)
	n.backID = tree.CreateElement(core.TypeButton)
	tree.AppendChild(n.id, n.backID)
	tree.SetFocusable(n.id, true)
	tree.AddHandler(n.backID, event.MouseClick, func(e *event.Event) {
		if targetPhase(e) {
			n.Pop()
			e.StopPropagation()
		}
	})
	tree.AddHandler(n.id, event.KeyDown, func(e *event.Event) {
		if !targetPhase(e) || e.Key != event.KeyBackspace {
			return
		}
		n.Pop()
		e.StopPropagation()
	})
	return n
}

func (n *Nav) Header() *NavHeader { return n.header }
func (n *Nav) Title() string      { return n.header.Title() }
func (n *Nav) Depth() int         { return len(n.stack) }
func (n *Nav) Push(title string, body Widget) {
	if len(n.stack) > 0 {
		n.tree.SetVisible(n.stack[len(n.stack)-1].body.ElementID(), false)
	}
	n.stack = append(n.stack, navPage{title: title, body: body})
	if body != nil {
		n.AppendChild(body)
		n.tree.SetVisible(body.ElementID(), true)
	}
	n.header.SetTitle(title)
}
func (n *Nav) Pop() {
	if len(n.stack) == 0 {
		return
	}
	top := n.stack[len(n.stack)-1]
	n.stack = n.stack[:len(n.stack)-1]
	if top.body != nil {
		n.tree.SetVisible(top.body.ElementID(), false)
	}
	title := ""
	if len(n.stack) > 0 {
		title = n.stack[len(n.stack)-1].title
		if n.stack[len(n.stack)-1].body != nil {
			n.tree.SetVisible(n.stack[len(n.stack)-1].body.ElementID(), true)
		}
	}
	n.header.SetTitle(title)
}
func (n *Nav) Draw(buf *render.CommandBuffer) {
	b := n.Bounds()
	if b.IsEmpty() {
		return
	}
	buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(b.X, b.Y, b.Width, 32), FillColor: n.config.BgColor}, 1, 1)
	n.tree.SetLayout(n.backID, core.LayoutResult{Bounds: uimath.NewRect(b.X+4, b.Y+4, 24, 24)})
	n.header.Draw(buf)
	if len(n.stack) > 0 && n.stack[len(n.stack)-1].body != nil {
		n.stack[len(n.stack)-1].body.Draw(buf)
	}
}

// FormSection groups form rows. Validation stays on Form.
type FormSection struct {
	Base
	title string
}

func NewFormSection(tree *core.Tree, title string, cfg *Config) *FormSection {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &FormSection{Base: NewBase(tree, core.TypeCustom, cfg), title: title}
}
func (s *FormSection) Title() string { return s.title }
func (s *FormSection) Add(row *FormRow) {
	if row != nil {
		s.AppendChild(row)
	}
}
func (s *FormSection) Draw(buf *render.CommandBuffer) {
	b := s.Bounds()
	if !b.IsEmpty() {
		buf.DrawRect(render.RectCmd{Bounds: b, BorderColor: s.config.BorderColor, BorderWidth: 1}, 1, 1)
	}
	s.DrawChildren(buf)
}

// FormRow lays out fields in one line.
type FormRow struct{ Base }

func NewFormRow(tree *core.Tree, cfg *Config) *FormRow {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &FormRow{Base: NewBase(tree, core.TypeCustom, cfg)}
}
func (r *FormRow) Add(field *FormField) {
	if field != nil {
		r.AppendChild(field)
	}
}
func (r *FormRow) Draw(buf *render.CommandBuffer) { r.DrawChildren(buf) }

// FormField is a label plus a control. It does not validate by itself.
type FormField struct {
	Base
	label   string
	control Widget
}

func NewFormField(tree *core.Tree, label string, cfg *Config) *FormField {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &FormField{Base: NewBase(tree, core.TypeCustom, cfg), label: label}
}
func (f *FormField) Label() string   { return f.label }
func (f *FormField) Control() Widget { return f.control }
func (f *FormField) SetControl(w Widget) {
	f.control = w
	if w != nil {
		f.AppendChild(w)
	}
}
func (f *FormField) Draw(buf *render.CommandBuffer) { f.DrawChildren(buf) }

// PlainShell is a title bar plus a content slot.
type PlainShell struct {
	Base
	title   string
	content Widget
}

func NewPlainShell(tree *core.Tree, cfg *Config) *PlainShell {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &PlainShell{Base: NewBase(tree, core.TypeCustom, cfg)}
}
func (s *PlainShell) SetTitle(t string) { s.title = t }
func (s *PlainShell) Title() string     { return s.title }
func (s *PlainShell) SetContent(w Widget) {
	s.content = w
	if w != nil {
		s.AppendChild(w)
	}
}
func (s *PlainShell) Content() Widget { return s.content }
func (s *PlainShell) Draw(buf *render.CommandBuffer) {
	b := s.Bounds()
	if !b.IsEmpty() {
		buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(b.X, b.Y, b.Width, 36), FillColor: s.config.PrimaryColor}, 1, 1)
	}
	if s.content != nil {
		s.content.Draw(buf)
	}
}

type sideEntry struct {
	label  string
	group  bool
	parent int
	open   bool
}

// Sidebar lists items and collapsible groups.
type Sidebar struct {
	Base
	entries []sideEntry
	index   int
}

func NewSidebar(tree *core.Tree, cfg *Config) *Sidebar {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	s := &Sidebar{Base: NewBase(tree, core.TypeCustom, cfg)}
	tree.SetFocusable(s.id, true)
	tree.AddHandler(s.id, event.KeyDown, func(e *event.Event) {
		if !targetPhase(e) {
			return
		}
		vis := s.visible()
		if len(vis) == 0 {
			return
		}
		pos := 0
		for i, idx := range vis {
			if idx == s.index {
				pos = i
			}
		}
		switch e.Key {
		case event.KeyArrowDown:
			s.index = vis[(pos+1)%len(vis)]
		case event.KeyArrowUp:
			s.index = vis[(pos-1+len(vis))%len(vis)]
		case event.KeyEnter, event.KeySpace:
			if s.entries[s.index].group {
				s.entries[s.index].open = !s.entries[s.index].open
			}
		default:
			return
		}
		e.StopPropagation()
	})
	tree.AddHandler(s.id, event.MouseClick, func(e *event.Event) {
		if !targetPhase(e) {
			return
		}
		vis := s.visible()
		b := s.Bounds()
		i := int((e.GlobalY - b.Y) / 28)
		if i < 0 || i >= len(vis) {
			return
		}
		s.index = vis[i]
		if s.entries[s.index].group {
			s.entries[s.index].open = !s.entries[s.index].open
		}
	})
	return s
}

func (s *Sidebar) AddItem(label string) int {
	s.entries = append(s.entries, sideEntry{label: label, parent: -1})
	return len(s.entries) - 1
}
func (s *Sidebar) AddGroup(label string) int {
	s.entries = append(s.entries, sideEntry{label: label, group: true, parent: -1, open: true})
	return len(s.entries) - 1
}
func (s *Sidebar) AddGroupItem(group int, label string) int {
	s.entries = append(s.entries, sideEntry{label: label, parent: group})
	return len(s.entries) - 1
}
func (s *Sidebar) Expanded(group int) bool {
	if group < 0 || group >= len(s.entries) {
		return false
	}
	return s.entries[group].open
}
func (s *Sidebar) Index() int { return s.index }
func (s *Sidebar) Label() string {
	if s.index < 0 || s.index >= len(s.entries) {
		return ""
	}
	return s.entries[s.index].label
}
func (s *Sidebar) visible() []int {
	var out []int
	for i, e := range s.entries {
		if e.parent >= 0 && (e.parent >= len(s.entries) || !s.entries[e.parent].open) {
			continue
		}
		out = append(out, i)
	}
	return out
}
func (s *Sidebar) Draw(buf *render.CommandBuffer) {
	b := s.Bounds()
	if b.IsEmpty() {
		return
	}
	buf.DrawRect(render.RectCmd{Bounds: b, FillColor: s.config.BgColor}, 1, 1)
	for row, idx := range s.visible() {
		buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(b.X, b.Y+float32(row)*28, b.Width, 26), FillColor: s.config.BorderColor}, 2, 1)
		_ = idx
	}
}

// SidebarShell pairs a sidebar with a content slot.
type SidebarShell struct {
	Base
	side    *Sidebar
	content Widget
}

func NewSidebarShell(tree *core.Tree, cfg *Config) *SidebarShell {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	s := &SidebarShell{Base: NewBase(tree, core.TypeCustom, cfg), side: NewSidebar(tree, cfg)}
	s.AppendChild(s.side)
	return s
}
func (s *SidebarShell) Sidebar() *Sidebar { return s.side }
func (s *SidebarShell) SetContent(w Widget) {
	s.content = w
	if w != nil {
		s.AppendChild(w)
	}
}
func (s *SidebarShell) Content() Widget { return s.content }
func (s *SidebarShell) Draw(buf *render.CommandBuffer) {
	s.side.Draw(buf)
	if s.content != nil {
		s.content.Draw(buf)
	}
}

// TabShell switches which content page is visible.
type TabShell struct {
	Base
	titles []string
	pages  []Widget
	cur    int
}

func NewTabShell(tree *core.Tree, cfg *Config) *TabShell {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	t := &TabShell{Base: NewBase(tree, core.TypeCustom, cfg)}
	tree.SetFocusable(t.id, true)
	tree.AddHandler(t.id, event.MouseClick, func(e *event.Event) {
		if !targetPhase(e) || len(t.titles) == 0 {
			return
		}
		b := t.Bounds()
		w := b.Width / float32(len(t.titles))
		if w <= 0 || e.GlobalY > b.Y+32 {
			return
		}
		i := int((e.GlobalX - b.X) / w)
		if i >= 0 && i < len(t.pages) {
			t.SetCurrent(i)
		}
	})
	tree.AddHandler(t.id, event.KeyDown, func(e *event.Event) {
		if !targetPhase(e) || len(t.pages) == 0 {
			return
		}
		switch e.Key {
		case event.KeyArrowRight:
			t.SetCurrent((t.cur + 1) % len(t.pages))
		case event.KeyArrowLeft:
			t.SetCurrent((t.cur - 1 + len(t.pages)) % len(t.pages))
		default:
			return
		}
		e.StopPropagation()
	})
	return t
}
func (t *TabShell) Add(title string, page Widget) {
	t.titles = append(t.titles, title)
	t.pages = append(t.pages, page)
	if page != nil {
		t.AppendChild(page)
		t.tree.SetVisible(page.ElementID(), len(t.pages) == 1)
	}
}
func (t *TabShell) Current() int { return t.cur }
func (t *TabShell) Title() string {
	if t.cur < 0 || t.cur >= len(t.titles) {
		return ""
	}
	return t.titles[t.cur]
}
func (t *TabShell) SetCurrent(i int) {
	if i < 0 || i >= len(t.pages) {
		return
	}
	for n, page := range t.pages {
		if page != nil {
			t.tree.SetVisible(page.ElementID(), n == i)
		}
	}
	t.cur = i
}
func (t *TabShell) Draw(buf *render.CommandBuffer) {
	b := t.Bounds()
	if !b.IsEmpty() && len(t.titles) > 0 {
		w := b.Width / float32(len(t.titles))
		for i := range t.titles {
			fill := t.config.BgColor
			if i == t.cur {
				fill = t.config.PrimaryColor
			}
			buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(b.X+float32(i)*w, b.Y, w, 32), FillColor: fill}, 1, 1)
		}
	}
	if t.cur >= 0 && t.cur < len(t.pages) && t.pages[t.cur] != nil {
		t.pages[t.cur].Draw(buf)
	}
}

type settingsCat struct {
	name string
	rows []string
}

// Settings shows a category list that opens that category's rows.
type Settings struct {
	Base
	cats []settingsCat
	cur  int
	side *Sidebar
}

func NewSettings(tree *core.Tree, cfg *Config) *Settings {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	s := &Settings{Base: NewBase(tree, core.TypeCustom, cfg), side: NewSidebar(tree, cfg)}
	s.AppendChild(s.side)
	return s
}
func (s *Settings) AddCategory(name string, rows []string) {
	s.cats = append(s.cats, settingsCat{name: name, rows: append([]string(nil), rows...)})
	s.side.AddItem(name)
}
func (s *Settings) Current() string {
	if s.cur < 0 || s.cur >= len(s.cats) {
		return ""
	}
	return s.cats[s.cur].name
}
func (s *Settings) Rows() []string {
	if s.cur < 0 || s.cur >= len(s.cats) {
		return nil
	}
	return append([]string(nil), s.cats[s.cur].rows...)
}
func (s *Settings) OpenCategory(i int) {
	if i < 0 || i >= len(s.cats) {
		return
	}
	s.cur = i
}
func (s *Settings) Draw(buf *render.CommandBuffer) { s.side.Draw(buf) }

// SettingsDialog is a modal around Settings. Escape closes it and restores focus.
type SettingsDialog struct {
	Base
	body   *Settings
	open   bool
	prior  core.ElementID
	scoped bool
}

func NewSettingsDialog(tree *core.Tree, cfg *Config) *SettingsDialog {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	d := &SettingsDialog{Base: NewBase(tree, core.TypeCustom, cfg), body: NewSettings(tree, cfg)}
	d.AppendChild(d.body)
	tree.SetVisible(d.id, false)
	tree.AddHandler(d.id, event.KeyDown, func(e *event.Event) {
		if e.Phase == event.PhaseCapture || e.Key != event.KeyEscape || !d.open {
			return
		}
		d.Close()
		e.StopPropagation()
	})
	tree.AddHandler(d.id, event.MouseClick, func(e *event.Event) {
		if !targetPhase(e) || !d.open {
			return
		}
		b := d.body.side.Bounds()
		if b.Height <= 0 {
			b = d.Bounds()
		}
		i := int((e.GlobalY - b.Y) / 28)
		if i >= 0 && i < len(d.body.cats) {
			d.body.OpenCategory(i)
		}
	})
	return d
}
func (d *SettingsDialog) Settings() *Settings { return d.body }
func (d *SettingsDialog) IsOpen() bool        { return d.open }
func (d *SettingsDialog) Open() {
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
	d.tree.SetFocusable(d.id, true)
	d.tree.SetFocused(d.id, true)
}
func (d *SettingsDialog) Close() {
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
		}
	}
}
func (d *SettingsDialog) AccessibleProps() accessibility.Props {
	return accessibility.Props{Role: accessibility.RoleDialog, Name: "settings"}
}
func (d *SettingsDialog) Draw(buf *render.CommandBuffer) {
	if !d.open {
		return
	}
	defer withLayer(buf, d.tree, d.id)()
	b := d.Bounds()
	if b.IsEmpty() {
		return
	}
	buf.DrawRect(render.RectCmd{Bounds: b, FillColor: uimath.Color{A: 0.45}}, 10, 1)
	d.body.Draw(buf)
}

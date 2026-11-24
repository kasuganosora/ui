package widget

import (
	"strings"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// FileEntry is one name returned by a FileSource. The picker never writes,
// deletes, or opens operating-system dialogs.
type FileEntry struct {
	Name string
	Dir  bool
}

// FileSource lists a directory. Paths use slash separators. The picker does
// not call the operating system itself.
type FileSource interface {
	List(path string) ([]FileEntry, error)
}

// FilePicker browses an injected directory listing. Enter opens a directory
// or reports a file. Backspace goes to the parent. It does not mutate the source.
type FilePicker struct {
	Base
	src    FileSource
	path   string
	rows   []FileEntry
	err    string
	index  int
	scroll float32
	rowH   float32
	pathH  float32
	onOpen func(path string, dir bool)
}

// NewFilePicker creates an empty browser at "/".
func NewFilePicker(tree *core.Tree, cfg *Config) *FilePicker {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	f := &FilePicker{Base: NewBase(tree, core.TypeCustom, cfg), path: "/", rowH: 22, pathH: 24}
	tree.SetFocusable(f.id, true)
	tree.AddHandler(f.id, event.KeyDown, func(e *event.Event) {
		if !targetPhase(e) {
			return
		}
		switch e.Key {
		case event.KeyArrowDown:
			f.move(1)
		case event.KeyArrowUp:
			f.move(-1)
		case event.KeyEnter:
			f.activate()
		case event.KeyBackspace:
			f.Up()
		default:
			return
		}
		e.StopPropagation()
	})
	tree.AddHandler(f.id, event.MouseClick, func(e *event.Event) {
		if !targetPhase(e) {
			return
		}
		if i := f.rowAt(e.GlobalY); i >= 0 {
			f.index = i
			f.reveal()
		}
	})
	return f
}

// SetSource installs the read-only listing callback and reloads the current path.
func (f *FilePicker) SetSource(src FileSource) {
	f.src = src
	_ = f.Open(f.path)
}

// Open lists path. A listing error is kept for Error and drawn in the body.
func (f *FilePicker) Open(path string) error {
	if path == "" {
		path = "/"
	}
	f.path = path
	f.index = 0
	f.scroll = 0
	f.rows = nil
	f.err = ""
	if f.src == nil {
		f.err = "no source"
		return errString(f.err)
	}
	rows, err := f.src.List(path)
	if err != nil {
		f.err = err.Error()
		return err
	}
	f.rows = rows
	return nil
}

func (f *FilePicker) Path() string         { return f.path }
func (f *FilePicker) Entries() []FileEntry { return f.rows }
func (f *FilePicker) Error() string        { return f.err }
func (f *FilePicker) SetScrollY(y float32) {
	if y < 0 {
		y = 0
	}
	f.scroll = y
}
func (f *FilePicker) OnOpen(fn func(path string, dir bool)) { f.onOpen = fn }

// Selected is the current entry name, or "".
func (f *FilePicker) Selected() string {
	if f.index < 0 || f.index >= len(f.rows) {
		return ""
	}
	return f.rows[f.index].Name
}

// Up lists the parent directory. The root stays at "/".
func (f *FilePicker) Up() {
	if f.path == "/" || f.path == "" {
		_ = f.Open("/")
		return
	}
	parent := f.path
	if i := strings.LastIndex(parent, "/"); i <= 0 {
		parent = "/"
	} else {
		parent = parent[:i]
	}
	_ = f.Open(parent)
}

func (f *FilePicker) move(dir int) {
	if len(f.rows) == 0 {
		return
	}
	f.index += dir
	if f.index < 0 {
		f.index = 0
	}
	if f.index >= len(f.rows) {
		f.index = len(f.rows) - 1
	}
	f.reveal()
}

func (f *FilePicker) reveal() {
	b := f.Bounds()
	bodyH := b.Height - f.pathH
	if bodyH < f.rowH {
		bodyH = f.rowH
	}
	top := float32(f.index) * f.rowH
	if top < f.scroll {
		f.scroll = top
	}
	if top+f.rowH > f.scroll+bodyH {
		f.scroll = top + f.rowH - bodyH
	}
}

func (f *FilePicker) activate() {
	if f.index < 0 || f.index >= len(f.rows) {
		return
	}
	ent := f.rows[f.index]
	path := f.join(ent.Name)
	if ent.Dir {
		_ = f.Open(path)
		return
	}
	if f.onOpen != nil {
		f.onOpen(path, false)
	}
}

func (f *FilePicker) join(name string) string {
	if f.path == "/" || f.path == "" {
		return "/" + name
	}
	return f.path + "/" + name
}

func (f *FilePicker) rowAt(y float32) int {
	b := f.Bounds()
	if y < b.Y+f.pathH || y >= b.Y+b.Height {
		return -1
	}
	i := int((y - (b.Y + f.pathH) + f.scroll) / f.rowH)
	if i < 0 || i >= len(f.rows) {
		return -1
	}
	return i
}

func (f *FilePicker) Draw(buf *render.CommandBuffer) {
	b := f.Bounds()
	if b.IsEmpty() {
		return
	}
	cfg := f.config
	buf.PushClip(b)
	pathBand := uimath.NewRect(b.X, b.Y, b.Width, f.pathH)
	bodyH := b.Height - f.pathH
	if bodyH < 0 {
		bodyH = 0
	}
	body := uimath.NewRect(b.X, b.Y+f.pathH, b.Width, bodyH)
	buf.PushClip(pathBand)
	buf.DrawRect(render.RectCmd{Bounds: pathBand, FillColor: cfg.SuccessColor}, 1, 1)
	emitText(cfg, buf, f.path, pathBand.X+8, pathBand.Y+4, cfg.FontSizeSm, pathBand.Width-16, cfg.TextColor, 1, 4)
	buf.PopClip()
	buf.PushClip(body)
	if f.err != "" {
		emitText(cfg, buf, f.err, body.X+8, body.Y+4, cfg.FontSize, body.Width-16, cfg.ErrorColor, 1, 4)
	}
	for i, ent := range f.rows {
		y := body.Y + float32(i)*f.rowH - f.scroll
		if y+f.rowH <= body.Y {
			continue
		}
		if y >= body.Y+body.Height {
			break
		}
		row := uimath.NewRect(body.X, y, body.Width, f.rowH)
		fill := uimath.ColorWhite
		if i == f.index {
			fill = cfg.PrimaryColor
		}
		buf.DrawRect(render.RectCmd{Bounds: row, FillColor: fill}, 1, 1)
		mark := cfg.BorderColor
		if ent.Dir {
			mark = cfg.TextColor
		}
		buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(row.X+6, y+6, 8, 8), FillColor: mark}, 2, 1)
		emitText(cfg, buf, ent.Name, row.X+20, y+2, cfg.FontSizeSm, row.Width-28, cfg.TextColor, 1, 4)
	}
	buf.PopClip()
	buf.PopClip()
}

type errString string

func (e errString) Error() string { return string(e) }

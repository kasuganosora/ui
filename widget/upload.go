package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// UploadFile represents an uploaded file entry.
type UploadFile struct {
	Name     string
	Size     int64
	Status   string // "uploading", "done", "error"
	Progress float32
}

// Upload is a file upload area widget.
type Upload struct {
	Base
	files    []UploadFile
	multiple bool
	accept   string
	drag     bool
	maxCount int
	onUpload func([]UploadFile)
}

func NewUpload(tree *core.Tree, cfg *Config) *Upload {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	u := &Upload{
		Base:     NewBase(tree, core.TypeCustom, cfg),
		multiple: false,
		drag:     true,
	}
	tree.AddHandler(u.id, event.MouseClick, func(e *event.Event) {
		// In a real implementation, this would open a file dialog
	})
	return u
}

func (u *Upload) Files() []UploadFile        { return u.files }
func (u *Upload) SetMultiple(m bool)          { u.multiple = m }
func (u *Upload) SetAccept(a string)          { u.accept = a }
func (u *Upload) SetDrag(d bool)              { u.drag = d }
func (u *Upload) SetMaxCount(m int)           { u.maxCount = m }
func (u *Upload) OnUpload(fn func([]UploadFile)) { u.onUpload = fn }

func (u *Upload) AddFile(f UploadFile) {
	if u.maxCount > 0 && len(u.files) >= u.maxCount {
		return
	}
	u.files = append(u.files, f)
}

func (u *Upload) RemoveFile(index int) {
	if index >= 0 && index < len(u.files) {
		u.files = append(u.files[:index], u.files[index+1:]...)
	}
}

func (u *Upload) ClearFiles() {
	u.files = u.files[:0]
}

func (u *Upload) Draw(buf *render.CommandBuffer) {
	bounds := u.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := u.config

	// Drop zone
	buf.DrawRect(render.RectCmd{
		Bounds:      bounds,
		FillColor:   uimath.RGBA(0, 0, 0, 0.01),
		BorderColor: cfg.BorderColor,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 1, 1)

	// Upload icon placeholder (+ symbol)
	cx := bounds.X + bounds.Width/2
	cy := bounds.Y + bounds.Height/2 - 10
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(cx-12, cy-1, 24, 2),
		FillColor: cfg.DisabledColor,
	}, 2, 1)
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(cx-1, cy-12, 2, 24),
		FillColor: cfg.DisabledColor,
	}, 2, 1)

	// Text
	if cfg.TextRenderer != nil {
		text := "Click or drag to upload"
		lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
		tw := cfg.TextRenderer.MeasureText(text, cfg.FontSizeSm)
		cfg.TextRenderer.DrawText(buf, text, cx-tw/2, cy+18, cfg.FontSizeSm, bounds.Width, cfg.DisabledColor, 1)
		_ = lh
	}
}

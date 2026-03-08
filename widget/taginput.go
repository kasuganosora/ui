package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// TagInput displays a list of tags with an inline input field for adding new tags.
type TagInput struct {
	Base
	tags        []string
	maxTags     int
	tagH        float32
	tagGap      float32
	inputText   string
	placeholder string
	size        Size
	clearable   bool
	onAdd       func(string)
	onRemove    func(string, int)
	onChange    func([]string)
}

func NewTagInput(tree *core.Tree, cfg *Config) *TagInput {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	ti := &TagInput{
		Base:    NewBase(tree, core.TypeCustom, cfg),
		maxTags: 0, // 0 = unlimited
		tagH:    26,
		tagGap:  6,
		size:    SizeMedium,
	}
	// Keyboard handler for Enter (add tag) and Backspace (remove last tag when input empty).
	tree.AddHandler(ti.id, event.KeyDown, func(e *event.Event) {
		switch e.Key {
		case event.KeyEnter:
			if ti.inputText != "" {
				ti.AddTag(ti.inputText)
				ti.inputText = ""
			}
		case event.KeyBackspace:
			if ti.inputText == "" && len(ti.tags) > 0 {
				ti.RemoveTag(len(ti.tags) - 1)
			}
		}
	})
	return ti
}

func (ti *TagInput) Tags() []string                    { return ti.tags }
func (ti *TagInput) InputText() string                 { return ti.inputText }
func (ti *TagInput) Placeholder() string               { return ti.placeholder }
func (ti *TagInput) MaxTags() int                      { return ti.maxTags }
func (ti *TagInput) Clearable() bool                   { return ti.clearable }
func (ti *TagInput) Size() Size                        { return ti.size }
func (ti *TagInput) SetMaxTags(m int)                  { ti.maxTags = m }
func (ti *TagInput) SetInputText(s string)             { ti.inputText = s }
func (ti *TagInput) SetPlaceholder(s string)           { ti.placeholder = s }
func (ti *TagInput) SetSize(s Size)                    { ti.size = s }
func (ti *TagInput) SetClearable(c bool)               { ti.clearable = c }
func (ti *TagInput) OnAdd(fn func(string))             { ti.onAdd = fn }
func (ti *TagInput) OnRemove(fn func(string, int))     { ti.onRemove = fn }
func (ti *TagInput) OnChange(fn func([]string))        { ti.onChange = fn }

func (ti *TagInput) AddTag(tag string) {
	if tag == "" {
		return
	}
	if ti.maxTags > 0 && len(ti.tags) >= ti.maxTags {
		return
	}
	ti.tags = append(ti.tags, tag)
	if ti.onAdd != nil {
		ti.onAdd(tag)
	}
	if ti.onChange != nil {
		ti.onChange(ti.tags)
	}
}

func (ti *TagInput) RemoveTag(index int) {
	if index < 0 || index >= len(ti.tags) {
		return
	}
	tag := ti.tags[index]
	ti.tags = append(ti.tags[:index], ti.tags[index+1:]...)
	if ti.onRemove != nil {
		ti.onRemove(tag, index)
	}
	if ti.onChange != nil {
		ti.onChange(ti.tags)
	}
}

func (ti *TagInput) ClearTags() {
	ti.tags = ti.tags[:0]
	if ti.onChange != nil {
		ti.onChange(ti.tags)
	}
}

// heightForSize returns the container height based on the Size setting.
func (ti *TagInput) heightForSize() float32 {
	switch ti.size {
	case SizeSmall:
		return 28
	case SizeLarge:
		return 40
	default: // SizeMedium
		return 32
	}
}

func (ti *TagInput) Draw(buf *render.CommandBuffer) {
	bounds := ti.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := ti.config

	// Container border
	borderClr := cfg.BorderColor
	if elem := ti.Element(); elem != nil && elem.IsFocused() {
		borderClr = cfg.FocusBorderColor
	}
	buf.DrawRect(render.RectCmd{
		Bounds:      bounds,
		FillColor:   cfg.BgColor,
		BorderColor: borderClr,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 1, 1)

	clearBtnW := float32(0)
	if ti.clearable && len(ti.tags) > 0 {
		clearBtnW = 20
	}
	contentRight := bounds.X + bounds.Width - cfg.SpaceXS - clearBtnW

	// Tags
	x := bounds.X + cfg.SpaceXS
	y := bounds.Y + cfg.SpaceXS
	for _, tag := range ti.tags {
		tw := float32(len(tag)) * cfg.FontSizeSm * 0.6
		tagW := tw + cfg.SpaceSM*2
		if cfg.TextRenderer != nil {
			tw = cfg.TextRenderer.MeasureText(tag, cfg.FontSizeSm)
			tagW = tw + cfg.SpaceSM*2
		}

		// Wrap
		if x+tagW > contentRight {
			x = bounds.X + cfg.SpaceXS
			y += ti.tagH + ti.tagGap
		}

		// Tag background
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x, y, tagW, ti.tagH),
			FillColor: uimath.RGBA(0, 0, 0, 0.04),
			Corners:   uimath.CornersAll(cfg.BorderRadius),
		}, 2, 1)

		// Tag text
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			cfg.TextRenderer.DrawText(buf, tag, x+cfg.SpaceSM, y+(ti.tagH-lh)/2, cfg.FontSizeSm, tagW-cfg.SpaceSM*2, cfg.TextColor, 1)
		}

		x += tagW + ti.tagGap
	}

	// Inline input area after last tag
	inputX := x
	inputY := y
	inputW := contentRight - inputX
	if inputW < 40 {
		// Wrap to next line if not enough space
		inputX = bounds.X + cfg.SpaceXS
		inputY += ti.tagH + ti.tagGap
		inputW = contentRight - inputX
	}

	if ti.inputText != "" {
		// Draw input text
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			cfg.TextRenderer.DrawText(buf, ti.inputText, inputX+cfg.SpaceXS, inputY+(ti.tagH-lh)/2, cfg.FontSizeSm, inputW-cfg.SpaceXS, cfg.TextColor, 1)
		} else {
			itw := float32(len(ti.inputText)) * cfg.FontSizeSm * 0.6
			ith := cfg.FontSizeSm * 1.2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(inputX+cfg.SpaceXS, inputY+(ti.tagH-ith)/2, itw, ith),
				FillColor: cfg.TextColor,
				Corners:   uimath.CornersAll(2),
			}, 2, 1)
		}
	} else if len(ti.tags) == 0 && ti.placeholder != "" {
		// Draw placeholder text
		placeholderColor := uimath.RGBA(0, 0, 0, 0.25)
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			cfg.TextRenderer.DrawText(buf, ti.placeholder, inputX+cfg.SpaceXS, inputY+(ti.tagH-lh)/2, cfg.FontSizeSm, inputW-cfg.SpaceXS, placeholderColor, 1)
		} else {
			pw := float32(len(ti.placeholder)) * cfg.FontSizeSm * 0.6
			ph := cfg.FontSizeSm * 1.2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(inputX+cfg.SpaceXS, inputY+(ti.tagH-ph)/2, pw, ph),
				FillColor: placeholderColor,
				Corners:   uimath.CornersAll(2),
			}, 2, 1)
		}
	}

	// Clear all button (X icon on the right)
	if ti.clearable && len(ti.tags) > 0 {
		clearX := bounds.X + bounds.Width - cfg.SpaceXS - clearBtnW
		clearY := bounds.Y + (bounds.Height-clearBtnW)/2
		// X shape: two crossing lines
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(clearX+4, clearY+clearBtnW/2-0.5, clearBtnW-8, 1),
			FillColor: uimath.RGBA(0, 0, 0, 0.25),
		}, 3, 1)
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(clearX+clearBtnW/2-0.5, clearY+4, 1, clearBtnW-8),
			FillColor: uimath.RGBA(0, 0, 0, 0.25),
		}, 3, 1)
	}
}

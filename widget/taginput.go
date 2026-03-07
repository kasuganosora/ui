package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// TagInput displays a list of tags with add/remove capability.
type TagInput struct {
	Base
	tags     []string
	maxTags  int
	tagH     float32
	tagGap   float32
	onAdd    func(string)
	onRemove func(string, int)
}

func NewTagInput(tree *core.Tree, cfg *Config) *TagInput {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &TagInput{
		Base:   NewBase(tree, core.TypeCustom, cfg),
		maxTags: 0, // 0 = unlimited
		tagH:   26,
		tagGap: 6,
	}
}

func (ti *TagInput) Tags() []string             { return ti.tags }
func (ti *TagInput) SetMaxTags(m int)           { ti.maxTags = m }
func (ti *TagInput) OnAdd(fn func(string))      { ti.onAdd = fn }
func (ti *TagInput) OnRemove(fn func(string, int)) { ti.onRemove = fn }

func (ti *TagInput) AddTag(tag string) {
	if ti.maxTags > 0 && len(ti.tags) >= ti.maxTags {
		return
	}
	ti.tags = append(ti.tags, tag)
	if ti.onAdd != nil {
		ti.onAdd(tag)
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
}

func (ti *TagInput) ClearTags() {
	ti.tags = ti.tags[:0]
}

func (ti *TagInput) Draw(buf *render.CommandBuffer) {
	bounds := ti.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := ti.config

	// Container border
	buf.DrawRect(render.RectCmd{
		Bounds:      bounds,
		FillColor:   cfg.BgColor,
		BorderColor: cfg.BorderColor,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 1, 1)

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
		if x+tagW > bounds.X+bounds.Width-cfg.SpaceXS {
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
}

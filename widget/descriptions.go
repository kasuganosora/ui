package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// DescriptionItem is a label-value pair.
type DescriptionItem struct {
	Label string
	Value string
}

// Descriptions displays a list of label-value pairs.
type Descriptions struct {
	Base
	title   string
	items   []DescriptionItem
	columns int
	bordered bool
	rowH    float32
}

func NewDescriptions(tree *core.Tree, cfg *Config) *Descriptions {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Descriptions{
		Base:     NewBase(tree, core.TypeCustom, cfg),
		columns:  3,
		bordered: true,
		rowH:     36,
	}
}

func (d *Descriptions) Title() string               { return d.title }
func (d *Descriptions) Items() []DescriptionItem     { return d.items }
func (d *Descriptions) SetTitle(t string)            { d.title = t }
func (d *Descriptions) SetColumns(c int)             { d.columns = c }
func (d *Descriptions) SetBordered(b bool)           { d.bordered = b }

func (d *Descriptions) AddItem(item DescriptionItem) {
	d.items = append(d.items, item)
}

func (d *Descriptions) ClearItems() {
	d.items = d.items[:0]
}

func (d *Descriptions) Draw(buf *render.CommandBuffer) {
	bounds := d.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := d.config
	headerH := float32(0)

	// Title
	if d.title != "" && cfg.TextRenderer != nil {
		headerH = 36
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, d.title, bounds.X, bounds.Y+(headerH-lh)/2, cfg.FontSize, bounds.Width, cfg.TextColor, 1)
	}

	if d.columns <= 0 {
		d.columns = 1
	}
	colW := bounds.Width / float32(d.columns)
	labelW := colW * 0.4

	y := bounds.Y + headerH
	for i, item := range d.items {
		col := i % d.columns
		row := i / d.columns
		cx := bounds.X + float32(col)*colW
		cy := y + float32(row)*d.rowH

		if cy+d.rowH > bounds.Y+bounds.Height {
			break
		}

		// Border
		if d.bordered {
			buf.DrawRect(render.RectCmd{
				Bounds:      uimath.NewRect(cx, cy, colW, d.rowH),
				BorderColor: cfg.BorderColor,
				BorderWidth: 1,
			}, 1, 1)
			// Label background
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(cx, cy, labelW, d.rowH),
				FillColor: uimath.RGBA(0, 0, 0, 0.02),
			}, 1, 1)
		}

		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			cfg.TextRenderer.DrawText(buf, item.Label, cx+cfg.SpaceSM, cy+(d.rowH-lh)/2, cfg.FontSizeSm, labelW-cfg.SpaceSM*2, cfg.DisabledColor, 1)
			cfg.TextRenderer.DrawText(buf, item.Value, cx+labelW+cfg.SpaceSM, cy+(d.rowH-lh)/2, cfg.FontSizeSm, colW-labelW-cfg.SpaceSM*2, cfg.TextColor, 1)
		}
	}
}

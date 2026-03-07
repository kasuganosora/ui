package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Empty displays an empty state placeholder with optional description text.
type Empty struct {
	Base
	description string
}

// NewEmpty creates an empty state widget.
func NewEmpty(tree *core.Tree, cfg *Config) *Empty {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	e := &Empty{
		Base:        NewBase(tree, core.TypeCustom, cfg),
		description: "暂无数据",
	}
	e.style.Display = layout.DisplayFlex
	e.style.AlignItems = layout.AlignCenter
	e.style.JustifyContent = layout.JustifyCenter
	e.style.FlexDirection = layout.FlexDirectionColumn
	e.style.Height = layout.Px(120)
	return e
}

func (e *Empty) Description() string          { return e.description }
func (e *Empty) SetDescription(desc string) { e.description = desc }

func (e *Empty) Draw(buf *render.CommandBuffer) {
	bounds := e.Bounds()
	if bounds.IsEmpty() {
		return
	}

	cfg := e.config
	color := cfg.DisabledColor

	// Draw a simple icon placeholder (empty box)
	iconSize := float32(48)
	iconX := bounds.X + (bounds.Width-iconSize)/2
	iconY := bounds.Y + (bounds.Height-iconSize)/2 - cfg.FontSize
	buf.DrawRect(render.RectCmd{
		Bounds:      uimath.NewRect(iconX, iconY, iconSize, iconSize),
		FillColor:   uimath.ColorTransparent,
		BorderColor: color,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 0, 1)

	// Description text
	if e.description != "" {
		textY := iconY + iconSize + cfg.SpaceSM
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			tw := cfg.TextRenderer.MeasureText(e.description, cfg.FontSize)
			tx := bounds.X + (bounds.Width-tw)/2
			cfg.TextRenderer.DrawText(buf, e.description, tx, textY, cfg.FontSize, bounds.Width, color, 1)
			_ = lh
		} else {
			textW := float32(len(e.description)) * cfg.FontSize * 0.55
			textH := cfg.FontSize * 1.2
			tx := bounds.X + (bounds.Width-textW)/2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(tx, textY, textW, textH),
				FillColor: color,
				Corners:   uimath.CornersAll(2),
			}, 1, 1)
		}
	}

	e.DrawChildren(buf)
}

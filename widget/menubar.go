package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// MenuBarItem represents a top-level menu bar item.
type MenuBarItem struct {
	Label    string
	SubItems []MenuBarSubItem
}

// MenuBarSubItem is an item within a dropdown menu.
type MenuBarSubItem struct {
	Label    string
	Shortcut string
	Disabled bool
	Divider  bool
	OnClick  func()
}

// MenuBar is a horizontal application menu bar.
type MenuBar struct {
	Base
	items    []MenuBarItem
	activeIdx int
	open     bool
	height   float32
}

func NewMenuBar(tree *core.Tree, cfg *Config) *MenuBar {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &MenuBar{
		Base:      NewBase(tree, core.TypeCustom, cfg),
		activeIdx: -1,
		height:    32,
	}
}

func (mb *MenuBar) Items() []MenuBarItem   { return mb.items }
func (mb *MenuBar) SetHeight(h float32)    { mb.height = h }
func (mb *MenuBar) IsOpen() bool           { return mb.open }

func (mb *MenuBar) AddItem(item MenuBarItem) {
	mb.items = append(mb.items, item)
}

func (mb *MenuBar) ClearItems() {
	mb.items = mb.items[:0]
	mb.activeIdx = -1
}

func (mb *MenuBar) Draw(buf *render.CommandBuffer) {
	bounds := mb.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := mb.config

	// Bar background
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(bounds.X, bounds.Y, bounds.Width, mb.height),
		FillColor: uimath.RGBA(0, 0, 0, 0.02),
	}, 1, 1)

	// Bottom border
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(bounds.X, bounds.Y+mb.height-1, bounds.Width, 1),
		FillColor: cfg.BorderColor,
	}, 1, 1)

	// Items
	x := bounds.X
	for i, item := range mb.items {
		itemW := float32(len(item.Label))*cfg.FontSize*0.6 + cfg.SpaceMD*2
		if cfg.TextRenderer != nil {
			itemW = cfg.TextRenderer.MeasureText(item.Label, cfg.FontSize) + cfg.SpaceMD*2
		}

		if i == mb.activeIdx && mb.open {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(x, bounds.Y, itemW, mb.height),
				FillColor: uimath.RGBA(0, 0, 0, 0.06),
			}, 2, 1)
		}

		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			cfg.TextRenderer.DrawText(buf, item.Label, x+cfg.SpaceMD, bounds.Y+(mb.height-lh)/2, cfg.FontSize, itemW-cfg.SpaceMD*2, cfg.TextColor, 1)
		}
		x += itemW
	}
}

package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/render"
)

// Portal renders its children at the root overlay level,
// escaping the normal widget tree rendering order.
type Portal struct {
	Base
	content Widget
	visible bool
	zBase   int
}

func NewPortal(tree *core.Tree, cfg *Config) *Portal {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Portal{
		Base:    NewBase(tree, core.TypeCustom, cfg),
		visible: true,
		zBase:   100,
	}
}

func (p *Portal) SetContent(w Widget) { p.content = w }
func (p *Portal) Content() Widget     { return p.content }
func (p *Portal) IsVisible() bool     { return p.visible }
func (p *Portal) SetVisible(v bool)   { p.visible = v }
func (p *Portal) SetZBase(z int)      { p.zBase = z }

func (p *Portal) Draw(buf *render.CommandBuffer) {
	if !p.visible || p.content == nil {
		return
	}
	// Portal renders content directly — the content widget is responsible
	// for using DrawOverlay if it needs overlay-level rendering.
	p.content.Draw(buf)
}

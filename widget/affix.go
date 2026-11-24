package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/render"
)

// Affix pins content to a fixed position when scrolled past a threshold.
type Affix struct {
	Base
	content   Widget
	offsetTop float32
	affixed   bool
	scrollY   float32
	restY     float32
	haveRest  bool
}

func NewAffix(tree *core.Tree, cfg *Config) *Affix {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Affix{
		Base: NewBase(tree, core.TypeCustom, cfg),
	}
}

func (a *Affix) Content() Widget        { return a.content }
func (a *Affix) SetContent(w Widget)    { a.content = w }
func (a *Affix) SetOffsetTop(o float32) { a.offsetTop = o }
func (a *Affix) IsAffixed() bool        { return a.affixed }
func (a *Affix) SetAffixed(v bool)      { a.affixed = v }

func (a *Affix) SetScrollY(y float32) {
	a.scrollY = y
	if a.content == nil {
		a.affixed = y >= a.offsetTop
		return
	}
	elem := a.tree.Get(a.content.ElementID())
	if elem == nil {
		a.affixed = y >= a.offsetTop
		return
	}
	b := elem.Layout().Bounds
	if !a.haveRest {
		a.restY = b.Y
		a.haveRest = true
	}
	a.affixed = y >= a.offsetTop
	if a.affixed {
		b.Y = a.offsetTop
	} else {
		b.Y = a.restY
	}
	a.tree.SetLayout(a.content.ElementID(), core.LayoutResult{Bounds: b})
}

func (a *Affix) Draw(buf *render.CommandBuffer) {
	if a.content == nil {
		return
	}
	a.content.Draw(buf)
}

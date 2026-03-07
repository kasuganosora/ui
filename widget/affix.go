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
}

func NewAffix(tree *core.Tree, cfg *Config) *Affix {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Affix{
		Base: NewBase(tree, core.TypeCustom, cfg),
	}
}

func (a *Affix) Content() Widget          { return a.content }
func (a *Affix) SetContent(w Widget)      { a.content = w }
func (a *Affix) SetOffsetTop(o float32)   { a.offsetTop = o }
func (a *Affix) IsAffixed() bool          { return a.affixed }
func (a *Affix) SetAffixed(v bool)        { a.affixed = v }

func (a *Affix) Draw(buf *render.CommandBuffer) {
	if a.content == nil {
		return
	}
	a.content.Draw(buf)
}

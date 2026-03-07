package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Rate is a star rating input.
type Rate struct {
	Base
	value    int
	count    int
	starSize float32
	gap      float32
	disabled bool
	onChange func(int)
}

func NewRate(tree *core.Tree, cfg *Config) *Rate {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	r := &Rate{
		Base:     NewBase(tree, core.TypeCustom, cfg),
		count:    5,
		starSize: 24,
		gap:      4,
	}
	tree.AddHandler(r.id, event.MouseClick, func(e *event.Event) {
		if r.disabled {
			return
		}
		bounds := r.Bounds()
		relX := e.GlobalX - bounds.X
		idx := int(relX / (r.starSize + r.gap))
		if idx >= 0 && idx < r.count {
			r.value = idx + 1
			if r.onChange != nil {
				r.onChange(r.value)
			}
		}
	})
	return r
}

func (r *Rate) Value() int            { return r.value }
func (r *Rate) SetValue(v int)        { r.value = v }
func (r *Rate) SetCount(c int)        { r.count = c }
func (r *Rate) SetStarSize(s float32) { r.starSize = s }
func (r *Rate) SetDisabled(d bool)    { r.disabled = d }
func (r *Rate) OnChange(fn func(int)) { r.onChange = fn }

func (r *Rate) Draw(buf *render.CommandBuffer) {
	bounds := r.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := r.config
	for i := 0; i < r.count; i++ {
		x := bounds.X + float32(i)*(r.starSize+r.gap)
		y := bounds.Y + (bounds.Height-r.starSize)/2
		color := cfg.BorderColor
		if i < r.value {
			color = uimath.ColorHex("#fadb14")
		}
		// Star as diamond shape (simplified)
		s := r.starSize
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x+s*0.15, y+s*0.15, s*0.7, s*0.7),
			FillColor: color,
			Corners:   uimath.CornersAll(s * 0.15),
		}, 1, 1)
	}
}

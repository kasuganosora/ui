package game

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// BuffType distinguishes buffs from debuffs.
type BuffType uint8

const (
	BuffPositive BuffType = iota
	BuffNegative
)

// Buff represents a single buff/debuff icon.
type Buff struct {
	ID       string
	Icon     render.TextureHandle
	Label    string
	Duration float32 // seconds remaining, 0 = permanent
	Stacks   int
	Type     BuffType
}

// BuffBar displays a row of buff/debuff icons.
type BuffBar struct {
	widget.Base
	buffs    []Buff
	iconSize float32
	gap      float32
	maxIcons int
	filter   BuffType // only draw this type (0=positive default)
	filterOn bool     // if true, only draw buffs matching filter
	onCancel func(id string)
}

func NewBuffBar(tree *core.Tree, cfg *widget.Config) *BuffBar {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	return &BuffBar{
		Base:     widget.NewBase(tree, core.TypeCustom, cfg),
		iconSize: 32,
		gap:      4,
		maxIcons: 16,
	}
}

func (bb *BuffBar) Buffs() []Buff          { return bb.buffs }
func (bb *BuffBar) SetIconSize(s float32)  { bb.iconSize = s }
func (bb *BuffBar) SetGap(g float32)       { bb.gap = g }
func (bb *BuffBar) SetMaxIcons(m int)      { bb.maxIcons = m }
func (bb *BuffBar) OnCancel(fn func(string)) { bb.onCancel = fn }

// SetFilter restricts drawing to only buffs of the given type.
func (bb *BuffBar) SetFilter(t BuffType) {
	bb.filter = t
	bb.filterOn = true
}

func (bb *BuffBar) AddBuff(b Buff) {
	bb.buffs = append(bb.buffs, b)
}

func (bb *BuffBar) RemoveBuff(id string) {
	for i, b := range bb.buffs {
		if b.ID == id {
			bb.buffs = append(bb.buffs[:i], bb.buffs[i+1:]...)
			return
		}
	}
}

func (bb *BuffBar) ClearBuffs() {
	bb.buffs = bb.buffs[:0]
}

// visibleBuffs returns the buffs that should be drawn (respecting filter).
func (bb *BuffBar) visibleBuffs() []Buff {
	if !bb.filterOn {
		return bb.buffs
	}
	var out []Buff
	for _, b := range bb.buffs {
		if b.Type == bb.filter {
			out = append(out, b)
		}
	}
	return out
}

func (bb *BuffBar) Draw(buf *render.CommandBuffer) {
	bounds := bb.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := bb.Config()
	s := bb.iconSize
	visible := bb.visibleBuffs()

	for i, b := range visible {
		if i >= bb.maxIcons {
			break
		}
		x := bounds.X + float32(i)*(s+bb.gap)
		y := bounds.Y

		// Icon background
		borderColor := uimath.ColorHex("#4488ff")
		if b.Type == BuffNegative {
			borderColor = uimath.ColorHex("#ff4444")
		}
		buf.DrawRect(render.RectCmd{
			Bounds:      uimath.NewRect(x, y, s, s),
			FillColor:   uimath.RGBA(0.15, 0.15, 0.15, 0.85),
			BorderColor: borderColor,
			BorderWidth: 1,
			Corners:     uimath.CornersAll(4),
		}, 1, 1)

		// Label text inside icon
		if b.Label != "" && cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			tw := cfg.TextRenderer.MeasureText(b.Label, cfg.FontSizeSm)
			cfg.TextRenderer.DrawText(buf, b.Label, x+(s-tw)/2, y+(s-lh)/2, cfg.FontSizeSm, s, uimath.ColorWhite, 1)
		}

		// Duration overlay (cooldown sweep from bottom)
		if b.Duration > 0 && b.Duration < 30 {
			frac := b.Duration / 30
			coolH := s * (1 - frac)
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(x, y, s, coolH),
				FillColor: uimath.RGBA(0, 0, 0, 0.5),
			}, 2, 1)
		}

		// Stack count
		if b.Stacks > 1 && cfg.TextRenderer != nil {
			txt := itoa(b.Stacks)
			tw := cfg.TextRenderer.MeasureText(txt, cfg.FontSizeSm)
			cfg.TextRenderer.DrawText(buf, txt, x+s-tw-2, y+s-cfg.TextRenderer.LineHeight(cfg.FontSizeSm)-1, cfg.FontSizeSm, s, uimath.ColorWhite, 1)
		}
	}
}

// HandleRightClick checks if (mx,my) hits a buff icon and calls onCancel.
// Returns true if a buff was cancelled.
func (bb *BuffBar) HandleRightClick(mx, my float32) bool {
	if bb.onCancel == nil {
		return false
	}
	bounds := bb.Bounds()
	if bounds.IsEmpty() {
		return false
	}
	s := bb.iconSize
	visible := bb.visibleBuffs()
	for i, b := range visible {
		if i >= bb.maxIcons {
			break
		}
		x := bounds.X + float32(i)*(s+bb.gap)
		y := bounds.Y
		if mx >= x && mx <= x+s && my >= y && my <= y+s {
			bb.onCancel(b.ID)
			return true
		}
	}
	return false
}

package game

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// NameplateType determines the visual style.
type NameplateType uint8

const (
	NameplateFriendly NameplateType = iota
	NameplateHostile
	NameplateNeutral
	NameplatePlayer
)

// Nameplate is a floating name/health indicator above an entity.
type Nameplate struct {
	widget.Base
	name      string
	title     string
	level     int
	hp        float32
	hpMax     float32
	npType    NameplateType
	x, y      float32
	barWidth  float32
	barHeight float32
	visible   bool
}

func NewNameplate(tree *core.Tree, name string, cfg *widget.Config) *Nameplate {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	return &Nameplate{
		Base:      widget.NewBase(tree, core.TypeCustom, cfg),
		name:      name,
		hp:        100,
		hpMax:     100,
		npType:    NameplateFriendly,
		barWidth:  100,
		barHeight: 6,
		visible:   true,
	}
}

func (np *Nameplate) Name() string                { return np.name }
func (np *Nameplate) SetName(n string)            { np.name = n }
func (np *Nameplate) SetTitle(t string)           { np.title = t }
func (np *Nameplate) SetLevel(l int)              { np.level = l }
func (np *Nameplate) SetHP(current, max float32)  { np.hp = current; np.hpMax = max }
func (np *Nameplate) SetType(t NameplateType)     { np.npType = t }
func (np *Nameplate) SetPosition(x, y float32)    { np.x = x; np.y = y }
func (np *Nameplate) SetBarSize(w, h float32)     { np.barWidth = w; np.barHeight = h }
func (np *Nameplate) IsVisible() bool             { return np.visible }
func (np *Nameplate) SetVisible(v bool)           { np.visible = v }

func nameplateColor(t NameplateType) uimath.Color {
	switch t {
	case NameplateHostile:
		return uimath.ColorHex("#ff4444")
	case NameplateNeutral:
		return uimath.ColorHex("#ffdd44")
	case NameplatePlayer:
		return uimath.ColorHex("#44aaff")
	default:
		return uimath.ColorHex("#44ff44")
	}
}

func (np *Nameplate) hpRatio() float32 {
	if np.hpMax <= 0 {
		return 0
	}
	r := np.hp / np.hpMax
	if r < 0 {
		r = 0
	}
	if r > 1 {
		r = 1
	}
	return r
}

func (np *Nameplate) Draw(buf *render.CommandBuffer) {
	if !np.visible {
		return
	}
	cfg := np.Config()
	color := nameplateColor(np.npType)
	cx := np.x

	// Name text
	nameY := np.y
	if cfg.TextRenderer != nil {
		tw := cfg.TextRenderer.MeasureText(np.name, cfg.FontSizeSm)
		lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
		cfg.TextRenderer.DrawText(buf, np.name, cx-tw/2, nameY, cfg.FontSizeSm, tw+4, color, 1)
		nameY += lh + 2
	} else {
		nameY += 16
	}

	// HP bar background
	bx := cx - np.barWidth/2
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(bx, nameY, np.barWidth, np.barHeight),
		FillColor: uimath.RGBA(0, 0, 0, 0.6),
		Corners:   uimath.CornersAll(np.barHeight / 2),
	}, 25, 1)

	// HP bar fill
	ratio := np.hpRatio()
	if ratio > 0 {
		hpColor := uimath.ColorHex("#52c41a")
		if ratio < 0.3 {
			hpColor = uimath.ColorHex("#ff4d4f")
		} else if ratio < 0.6 {
			hpColor = uimath.ColorHex("#faad14")
		}
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(bx, nameY, np.barWidth*ratio, np.barHeight),
			FillColor: hpColor,
			Corners:   uimath.CornersAll(np.barHeight / 2),
		}, 26, 1)
	}
}

package game

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// UnitFrameData holds the data for a unit frame (team member, target, etc.)
type UnitFrameData struct {
	Name     string
	Level    int
	HP       float32
	HPMax    float32
	MP       float32
	MPMax    float32
	Class    string
	Portrait render.TextureHandle
	Dead     bool
}

// TeamFrame displays party/team member unit frames.
type TeamFrame struct {
	widget.Base
	members  []UnitFrameData
	frameW   float32
	frameH   float32
	gap      float32
	maxSlots int
}

func NewTeamFrame(tree *core.Tree, cfg *widget.Config) *TeamFrame {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	return &TeamFrame{
		Base:     widget.NewBase(tree, core.TypeCustom, cfg),
		frameW:   160,
		frameH:   48,
		gap:      4,
		maxSlots: 5,
	}
}

func (tf *TeamFrame) Members() []UnitFrameData  { return tf.members }
func (tf *TeamFrame) SetFrameSize(w, h float32) { tf.frameW = w; tf.frameH = h }
func (tf *TeamFrame) SetGap(g float32)          { tf.gap = g }
func (tf *TeamFrame) SetMaxSlots(m int)         { tf.maxSlots = m }

func (tf *TeamFrame) SetMembers(members []UnitFrameData) {
	tf.members = make([]UnitFrameData, len(members))
	copy(tf.members, members)
}

func (tf *TeamFrame) UpdateMember(index int, data UnitFrameData) {
	if index >= 0 && index < len(tf.members) {
		tf.members[index] = data
	}
}

func (tf *TeamFrame) Draw(buf *render.CommandBuffer) {
	bounds := tf.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := tf.Config()

	for i, m := range tf.members {
		if i >= tf.maxSlots {
			break
		}
		y := bounds.Y + float32(i)*(tf.frameH+tf.gap)
		drawUnitFrame(buf, cfg, bounds.X, y, tf.frameW, tf.frameH, m)
	}
}

// TargetFrame displays the currently selected target.
type TargetFrame struct {
	widget.Base
	target   *UnitFrameData
	width    float32
	height   float32
	visible  bool
}

func NewTargetFrame(tree *core.Tree, cfg *widget.Config) *TargetFrame {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	return &TargetFrame{
		Base:   widget.NewBase(tree, core.TypeCustom, cfg),
		width:  200,
		height: 56,
	}
}

func (tf *TargetFrame) Target() *UnitFrameData    { return tf.target }
func (tf *TargetFrame) IsVisible() bool            { return tf.visible }
func (tf *TargetFrame) SetSize(w, h float32)       { tf.width = w; tf.height = h }

func (tf *TargetFrame) SetTarget(data *UnitFrameData) {
	tf.target = data
	tf.visible = data != nil
}

func (tf *TargetFrame) ClearTarget() {
	tf.target = nil
	tf.visible = false
}

func (tf *TargetFrame) Draw(buf *render.CommandBuffer) {
	if !tf.visible || tf.target == nil {
		return
	}
	bounds := tf.Bounds()
	x, y := bounds.X, bounds.Y
	if bounds.IsEmpty() {
		x, y = 0, 0
	}
	drawUnitFrame(buf, tf.Config(), x, y, tf.width, tf.height, *tf.target)
}

// drawUnitFrame renders a single unit frame (shared by TeamFrame and TargetFrame).
func drawUnitFrame(buf *render.CommandBuffer, cfg *widget.Config, x, y, w, h float32, data UnitFrameData) {
	barH := float32(8)

	// Background
	buf.DrawOverlay(render.RectCmd{
		Bounds:      uimath.NewRect(x, y, w, h),
		FillColor:   uimath.RGBA(0.08, 0.08, 0.12, 0.9),
		BorderColor: uimath.RGBA(0.3, 0.3, 0.4, 0.8),
		BorderWidth: 1,
		Corners:     uimath.CornersAll(4),
	}, 20, 1)

	// Name + level
	if cfg.TextRenderer != nil {
		nameText := data.Name
		if data.Level > 0 {
			nameText = "[" + itoa(data.Level) + "] " + data.Name
		}
		nameColor := uimath.ColorWhite
		if data.Dead {
			nameColor = uimath.RGBA(0.5, 0.5, 0.5, 1)
		}
		cfg.TextRenderer.DrawText(buf, nameText, x+cfg.SpaceXS, y+2, cfg.FontSizeSm, w-cfg.SpaceXS*2, nameColor, 1)
	}

	// HP bar
	hpY := y + h - barH*2 - 4
	drawResourceBar(buf, x+4, hpY, w-8, barH, data.HP, data.HPMax, uimath.ColorHex("#52c41a"))

	// MP bar
	mpY := hpY + barH + 2
	drawResourceBar(buf, x+4, mpY, w-8, barH, data.MP, data.MPMax, uimath.ColorHex("#1890ff"))
}

func drawResourceBar(buf *render.CommandBuffer, x, y, w, h, current, max float32, color uimath.Color) {
	// Background
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(x, y, w, h),
		FillColor: uimath.RGBA(0, 0, 0, 0.5),
		Corners:   uimath.CornersAll(h / 2),
	}, 21, 1)

	// Fill
	if max > 0 && current > 0 {
		ratio := current / max
		if ratio > 1 {
			ratio = 1
		}
		buf.DrawOverlay(render.RectCmd{
			Bounds:    uimath.NewRect(x, y, w*ratio, h),
			FillColor: color,
			Corners:   uimath.CornersAll(h / 2),
		}, 22, 1)
	}
}

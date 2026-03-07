package game

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// DialogueChoice represents a player response option.
type DialogueChoice struct {
	Text    string
	OnClick func()
}

// DialogueBox is an NPC dialogue window with speaker name, text, and choices.
type DialogueBox struct {
	widget.Base
	speaker    string
	text       string
	portrait   render.TextureHandle
	choices    []DialogueChoice
	visible    bool
	width      float32
	height     float32
	onAdvance  func()
}

func NewDialogueBox(tree *core.Tree, cfg *widget.Config) *DialogueBox {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	db := &DialogueBox{
		Base:   widget.NewBase(tree, core.TypeCustom, cfg),
		width:  500,
		height: 160,
	}
	tree.AddHandler(db.ElementID(), event.MouseClick, func(e *event.Event) {
		if len(db.choices) == 0 && db.onAdvance != nil {
			db.onAdvance()
		}
	})
	return db
}

func (db *DialogueBox) Speaker() string              { return db.speaker }
func (db *DialogueBox) Text() string                 { return db.text }
func (db *DialogueBox) Choices() []DialogueChoice    { return db.choices }
func (db *DialogueBox) IsVisible() bool              { return db.visible }
func (db *DialogueBox) SetSpeaker(s string)          { db.speaker = s }
func (db *DialogueBox) SetText(t string)             { db.text = t }
func (db *DialogueBox) SetPortrait(t render.TextureHandle) { db.portrait = t }
func (db *DialogueBox) SetSize(w, h float32)         { db.width = w; db.height = h }
func (db *DialogueBox) OnAdvance(fn func())          { db.onAdvance = fn }

func (db *DialogueBox) SetChoices(choices []DialogueChoice) {
	db.choices = make([]DialogueChoice, len(choices))
	copy(db.choices, choices)
}

func (db *DialogueBox) ClearChoices() {
	db.choices = db.choices[:0]
}

func (db *DialogueBox) Show(speaker, text string) {
	db.speaker = speaker
	db.text = text
	db.visible = true
}

func (db *DialogueBox) Hide() {
	db.visible = false
}

func (db *DialogueBox) Draw(buf *render.CommandBuffer) {
	if !db.visible {
		return
	}
	cfg := db.Config()
	x := float32(0)
	y := float32(0)
	bounds := db.Bounds()
	if !bounds.IsEmpty() {
		x = bounds.X
		y = bounds.Y
	}

	// Background
	buf.DrawOverlay(render.RectCmd{
		Bounds:      uimath.NewRect(x, y, db.width, db.height),
		FillColor:   uimath.RGBA(0.05, 0.05, 0.1, 0.92),
		BorderColor: uimath.RGBA(0.4, 0.4, 0.5, 0.8),
		BorderWidth: 2,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 60, 1)

	textX := x + cfg.SpaceMD
	textY := y + cfg.SpaceSM

	// Speaker name
	if cfg.TextRenderer != nil && db.speaker != "" {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, db.speaker, textX, textY, cfg.FontSize, db.width-cfg.SpaceMD*2, uimath.ColorHex("#ffd700"), 1)
		textY += lh + 4
	}

	// Dialogue text
	if cfg.TextRenderer != nil && db.text != "" {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, db.text, textX, textY, cfg.FontSize, db.width-cfg.SpaceMD*2, uimath.RGBA(0.9, 0.9, 0.9, 1), 1)
		textY += lh + 8
	}

	// Choices
	choiceH := float32(24)
	for i, choice := range db.choices {
		cy := textY + float32(i)*(choiceH+4)
		if cfg.TextRenderer != nil {
			label := itoa(i+1) + ". " + choice.Text
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			cfg.TextRenderer.DrawText(buf, label, textX+cfg.SpaceSM, cy+(choiceH-lh)/2, cfg.FontSizeSm, db.width-cfg.SpaceMD*3, uimath.ColorHex("#88bbff"), 1)
		}
	}

	// "Click to continue" hint if no choices
	if len(db.choices) == 0 && cfg.TextRenderer != nil {
		hint := "▼"
		tw := cfg.TextRenderer.MeasureText(hint, cfg.FontSizeSm)
		cfg.TextRenderer.DrawText(buf, hint, x+db.width-tw-cfg.SpaceMD, y+db.height-cfg.TextRenderer.LineHeight(cfg.FontSizeSm)-cfg.SpaceSM, cfg.FontSizeSm, tw+4, uimath.RGBA(0.6, 0.6, 0.6, 0.8), 1)
	}
}

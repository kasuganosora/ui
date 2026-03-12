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

	// Measure content height for auto-sizing
	contentH := float32(0)
	if cfg.TextRenderer != nil && db.speaker != "" {
		contentH += cfg.TextRenderer.LineHeight(cfg.FontSize) + 4
	}
	if cfg.TextRenderer != nil && db.text != "" {
		contentH += cfg.TextRenderer.LineHeight(cfg.FontSize) + 8
	}
	choiceH := float32(24)
	if len(db.choices) > 0 {
		contentH += float32(len(db.choices))*(choiceH+4) + 4
	}
	contentH += cfg.SpaceSM

	panel := Panel{
		Title:       db.speaker,
		Width:       db.width,
		Height:      db.height,
		TitleH:      30,
		BgColor:     uimath.RGBA(0.05, 0.05, 0.1, 0.92),
		BorderWidth: 2,
	}
	if db.speaker == "" {
		panel.TitleH = -1
	}
	r := panel.Draw(buf, db.Bounds(), cfg, contentH)

	textY := r.ContentY + 4

	// Dialogue text
	if cfg.TextRenderer != nil && db.text != "" {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, db.text, r.ContentX, textY, cfg.FontSize, r.ContentW, uimath.RGBA(0.9, 0.9, 0.9, 1), 1)
		textY += lh + 8
	}

	// Choices
	for i, choice := range db.choices {
		cy := textY + float32(i)*(choiceH+4)
		if cfg.TextRenderer != nil {
			label := itoa(i+1) + ". " + choice.Text
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			cfg.TextRenderer.DrawText(buf, label, r.ContentX+cfg.SpaceSM, cy+(choiceH-lh)/2, cfg.FontSizeSm, r.ContentW-cfg.SpaceSM, uimath.ColorHex("#88bbff"), 1)
		}
	}

	// "Click to continue" hint if no choices
	if len(db.choices) == 0 && cfg.TextRenderer != nil {
		hint := "▼"
		tw := cfg.TextRenderer.MeasureText(hint, cfg.FontSizeSm)
		cfg.TextRenderer.DrawText(buf, hint, r.PanelX+r.PanelW-tw-cfg.SpaceMD, r.PanelY+r.PanelH-cfg.TextRenderer.LineHeight(cfg.FontSizeSm)-cfg.SpaceSM, cfg.FontSizeSm, tw+4, uimath.RGBA(0.6, 0.6, 0.6, 0.8), 1)
	}
}

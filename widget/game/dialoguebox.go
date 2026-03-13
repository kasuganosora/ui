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
	embedded   bool // if true, skip Panel chrome (used inside Window)
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
func (db *DialogueBox) SetEmbedded(v bool)           { db.embedded = v }

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
	bounds := db.Bounds()
	choiceH := float32(24)

	var cx, cy, cw, px, py, pw, ph float32

	if db.embedded {
		// Embedded mode: no Panel chrome, use bounds directly
		pad := float32(8)
		cx = bounds.X + pad
		cy = bounds.Y + pad
		cw = bounds.Width - pad*2
		if cw <= 0 {
			cw = db.width - pad*2
		}
		px, py = bounds.X, bounds.Y
		pw = bounds.Width
		if pw == 0 {
			pw = db.width
		}
		ph = bounds.Height
		if ph == 0 {
			ph = db.height
		}
	} else {
		// Standalone mode: draw Panel chrome
		contentH := float32(0)
		if cfg.TextRenderer != nil && db.speaker != "" {
			contentH += cfg.TextRenderer.LineHeight(cfg.FontSize) + 4
		}
		if cfg.TextRenderer != nil && db.text != "" {
			contentH += cfg.TextRenderer.LineHeight(cfg.FontSize) + 8
		}
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
		r := panel.Draw(buf, bounds, cfg, contentH)
		cx, cy, cw = r.ContentX, r.ContentY, r.ContentW
		px, py, pw, ph = r.PanelX, r.PanelY, r.PanelW, r.PanelH
	}

	textY := cy + 4

	// Speaker name (in embedded mode, drawn as content text)
	if db.embedded && db.speaker != "" && cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, db.speaker, cx, textY, cfg.FontSize, cw, uimath.ColorHex("#ffd700"), 1)
		textY += lh + 4
	}

	// Dialogue text
	if cfg.TextRenderer != nil && db.text != "" {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, db.text, cx, textY, cfg.FontSize, cw, uimath.RGBA(0.9, 0.9, 0.9, 1), 1)
		textY += lh + 8
	}

	// Choices
	for i, choice := range db.choices {
		cy := textY + float32(i)*(choiceH+4)
		if cfg.TextRenderer != nil {
			label := itoa(i+1) + ". " + choice.Text
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			cfg.TextRenderer.DrawText(buf, label, cx+cfg.SpaceSM, cy+(choiceH-lh)/2, cfg.FontSizeSm, cw-cfg.SpaceSM, uimath.ColorHex("#88bbff"), 1)
		}
	}

	// "Click to continue" hint if no choices
	if len(db.choices) == 0 && cfg.TextRenderer != nil {
		hint := "▼"
		tw := cfg.TextRenderer.MeasureText(hint, cfg.FontSizeSm)
		cfg.TextRenderer.DrawText(buf, hint, px+pw-tw-cfg.SpaceMD, py+ph-cfg.TextRenderer.LineHeight(cfg.FontSizeSm)-cfg.SpaceSM, cfg.FontSizeSm, tw+4, uimath.RGBA(0.6, 0.6, 0.6, 0.8), 1)
	}
}

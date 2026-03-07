package game

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// QuestObjective represents a single quest objective.
type QuestObjective struct {
	Text      string
	Current   int
	Required  int
	Completed bool
}

// Quest represents a tracked quest.
type Quest struct {
	Title      string
	Objectives []QuestObjective
	Active     bool
}

// QuestTracker displays a list of active quests and their objectives.
type QuestTracker struct {
	widget.Base
	quests    []Quest
	width     float32
	maxQuests int
}

func NewQuestTracker(tree *core.Tree, cfg *widget.Config) *QuestTracker {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	return &QuestTracker{
		Base:      widget.NewBase(tree, core.TypeCustom, cfg),
		width:     250,
		maxQuests: 5,
	}
}

func (qt *QuestTracker) Quests() []Quest       { return qt.quests }
func (qt *QuestTracker) SetWidth(w float32)    { qt.width = w }
func (qt *QuestTracker) SetMaxQuests(m int)    { qt.maxQuests = m }

func (qt *QuestTracker) AddQuest(q Quest) {
	qt.quests = append(qt.quests, q)
}

func (qt *QuestTracker) RemoveQuest(index int) {
	if index >= 0 && index < len(qt.quests) {
		qt.quests = append(qt.quests[:index], qt.quests[index+1:]...)
	}
}

func (qt *QuestTracker) ClearQuests() {
	qt.quests = qt.quests[:0]
}

func (qt *QuestTracker) Draw(buf *render.CommandBuffer) {
	bounds := qt.Bounds()
	if bounds.IsEmpty() {
		bounds = uimath.NewRect(0, 0, qt.width, 400)
	}
	cfg := qt.Config()
	y := bounds.Y

	for i, quest := range qt.quests {
		if i >= qt.maxQuests {
			break
		}
		// Quest title
		titleColor := uimath.ColorHex("#ffd700")
		if !quest.Active {
			titleColor = uimath.RGBA(0.6, 0.6, 0.6, 1)
		}
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			cfg.TextRenderer.DrawText(buf, quest.Title, bounds.X, y, cfg.FontSize, bounds.Width, titleColor, 1)
			y += lh + 2
		} else {
			y += 20
		}

		// Objectives
		for _, obj := range quest.Objectives {
			objColor := uimath.RGBA(0.8, 0.8, 0.8, 1)
			prefix := "  - "
			if obj.Completed {
				objColor = uimath.RGBA(0.5, 0.5, 0.5, 1)
				prefix = "  ✓ "
			}
			text := prefix + obj.Text
			if obj.Required > 0 {
				text += " (" + itoa(obj.Current) + "/" + itoa(obj.Required) + ")"
			}
			if cfg.TextRenderer != nil {
				lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
				cfg.TextRenderer.DrawText(buf, text, bounds.X, y, cfg.FontSizeSm, bounds.Width, objColor, 1)
				y += lh + 1
			} else {
				y += 16
			}
		}
		y += 6 // gap between quests
	}
}

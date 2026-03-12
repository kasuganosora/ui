package game

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// ScoreEntry represents a single player's score.
type ScoreEntry struct {
	Name   string
	Score  int
	Kills  int
	Deaths int
	Team   int // 0, 1, etc.
}

// Scoreboard displays a list of player scores.
type Scoreboard struct {
	widget.Base
	entries  []ScoreEntry
	title    string
	visible  bool
	width    float32
	rowH     float32
	headerH  float32
}

func NewScoreboard(tree *core.Tree, cfg *widget.Config) *Scoreboard {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	return &Scoreboard{
		Base:    widget.NewBase(tree, core.TypeCustom, cfg),
		title:   "Scoreboard",
		visible: false,
		width:   400,
		rowH:    28,
		headerH: 36,
	}
}

func (sb *Scoreboard) Entries() []ScoreEntry   { return sb.entries }
func (sb *Scoreboard) IsVisible() bool         { return sb.visible }
func (sb *Scoreboard) SetVisible(v bool)       { sb.visible = v }
func (sb *Scoreboard) SetTitle(t string)       { sb.title = t }
func (sb *Scoreboard) SetWidth(w float32)      { sb.width = w }

func (sb *Scoreboard) AddEntry(e ScoreEntry) {
	sb.entries = append(sb.entries, e)
}

func (sb *Scoreboard) ClearEntries() {
	sb.entries = sb.entries[:0]
}

// SortByScore sorts entries by score descending (simple insertion sort).
func (sb *Scoreboard) SortByScore() {
	for i := 1; i < len(sb.entries); i++ {
		key := sb.entries[i]
		j := i - 1
		for j >= 0 && sb.entries[j].Score < key.Score {
			sb.entries[j+1] = sb.entries[j]
			j--
		}
		sb.entries[j+1] = key
	}
}

func (sb *Scoreboard) Draw(buf *render.CommandBuffer) {
	if !sb.visible || len(sb.entries) == 0 {
		return
	}
	cfg := sb.Config()
	colHeaderH := float32(24)
	contentH := colHeaderH + float32(len(sb.entries))*sb.rowH

	panel := Panel{
		Title:   sb.title,
		Width:   sb.width,
		TitleH:  sb.headerH,
		BgColor: uimath.RGBA(0, 0, 0, 0.85),
	}
	r := panel.Draw(buf, sb.Bounds(), cfg, contentH)

	// Column headers
	colW := r.PanelW / 4
	headers := [4]string{"Name", "Score", "K", "D"}
	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
		hy := r.ContentY + (colHeaderH-lh)/2
		for ci, h := range headers {
			cfg.TextRenderer.DrawText(buf, h, r.PanelX+float32(ci)*colW+cfg.SpaceXS, hy, cfg.FontSizeSm, colW, uimath.RGBA(0.6, 0.6, 0.6, 1), 1)
		}
	}

	// Rows
	dataY := r.ContentY + colHeaderH
	for i, e := range sb.entries {
		ry := dataY + float32(i)*sb.rowH
		if i%2 == 1 {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(r.PanelX, ry, r.PanelW, sb.rowH),
				FillColor: uimath.RGBA(1, 1, 1, 0.03),
			}, 3, 1)
		}
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			ty := ry + (sb.rowH-lh)/2
			cfg.TextRenderer.DrawText(buf, e.Name, r.PanelX+cfg.SpaceXS, ty, cfg.FontSizeSm, colW, uimath.ColorWhite, 1)
			cfg.TextRenderer.DrawText(buf, itoa(e.Score), r.PanelX+colW+cfg.SpaceXS, ty, cfg.FontSizeSm, colW, uimath.ColorWhite, 1)
			cfg.TextRenderer.DrawText(buf, itoa(e.Kills), r.PanelX+colW*2+cfg.SpaceXS, ty, cfg.FontSizeSm, colW, uimath.ColorHex("#52c41a"), 1)
			cfg.TextRenderer.DrawText(buf, itoa(e.Deaths), r.PanelX+colW*3+cfg.SpaceXS, ty, cfg.FontSizeSm, colW, uimath.ColorHex("#ff4d4f"), 1)
		}
	}
}

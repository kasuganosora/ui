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
	totalH := sb.headerH + float32(len(sb.entries))*sb.rowH
	x := float32(0)
	y := float32(0)
	bounds := sb.Bounds()
	if !bounds.IsEmpty() {
		x = bounds.X
		y = bounds.Y
	}

	// Background
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(x, y, sb.width, totalH),
		FillColor: uimath.RGBA(0, 0, 0, 0.85),
		Corners:   uimath.CornersAll(cfg.BorderRadius),
	}, 50, 1)

	// Header
	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, sb.title, x+cfg.SpaceSM, y+(sb.headerH-lh)/2, cfg.FontSize, sb.width-cfg.SpaceSM*2, uimath.ColorHex("#ffd700"), 1)
	}

	// Column headers
	colW := sb.width / 4
	headers := [4]string{"Name", "Score", "K", "D"}
	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
		hy := y + sb.headerH - lh - 2
		for ci, h := range headers {
			cfg.TextRenderer.DrawText(buf, h, x+float32(ci)*colW+cfg.SpaceXS, hy, cfg.FontSizeSm, colW, uimath.RGBA(0.6, 0.6, 0.6, 1), 1)
		}
	}

	// Rows
	for i, e := range sb.entries {
		ry := y + sb.headerH + float32(i)*sb.rowH
		// Alternating row
		if i%2 == 1 {
			buf.DrawOverlay(render.RectCmd{
				Bounds:    uimath.NewRect(x, ry, sb.width, sb.rowH),
				FillColor: uimath.RGBA(1, 1, 1, 0.03),
			}, 51, 1)
		}
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			ty := ry + (sb.rowH-lh)/2
			cfg.TextRenderer.DrawText(buf, e.Name, x+cfg.SpaceXS, ty, cfg.FontSizeSm, colW, uimath.ColorWhite, 1)
			cfg.TextRenderer.DrawText(buf, itoa(e.Score), x+colW+cfg.SpaceXS, ty, cfg.FontSizeSm, colW, uimath.ColorWhite, 1)
			cfg.TextRenderer.DrawText(buf, itoa(e.Kills), x+colW*2+cfg.SpaceXS, ty, cfg.FontSizeSm, colW, uimath.ColorHex("#52c41a"), 1)
			cfg.TextRenderer.DrawText(buf, itoa(e.Deaths), x+colW*3+cfg.SpaceXS, ty, cfg.FontSizeSm, colW, uimath.ColorHex("#ff4d4f"), 1)
		}
	}
}

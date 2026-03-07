package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// AutoComplete is an input with suggestion dropdown.
type AutoComplete struct {
	Base
	text        string
	suggestions []string
	filtered    []string
	open        bool
	selected    int
	onSelect    func(string)
	filterFn    func(input string, options []string) []string
}

func NewAutoComplete(tree *core.Tree, cfg *Config) *AutoComplete {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	ac := &AutoComplete{
		Base:     NewBase(tree, core.TypeCustom, cfg),
		selected: -1,
	}
	ac.filterFn = defaultAutoFilter
	return ac
}

func (ac *AutoComplete) Text() string                { return ac.text }
func (ac *AutoComplete) IsOpen() bool                { return ac.open }
func (ac *AutoComplete) SetOpen(o bool)              { ac.open = o }
func (ac *AutoComplete) OnSelect(fn func(string))    { ac.onSelect = fn }
func (ac *AutoComplete) Filtered() []string          { return ac.filtered }

func (ac *AutoComplete) SetSuggestions(s []string) {
	ac.suggestions = make([]string, len(s))
	copy(ac.suggestions, s)
}

func (ac *AutoComplete) SetText(t string) {
	ac.text = t
	ac.filtered = ac.filterFn(t, ac.suggestions)
	ac.open = len(ac.filtered) > 0 && t != ""
	ac.selected = -1
}

func (ac *AutoComplete) SetFilterFn(fn func(string, []string) []string) {
	ac.filterFn = fn
}

func (ac *AutoComplete) SelectItem(index int) {
	if index >= 0 && index < len(ac.filtered) {
		ac.text = ac.filtered[index]
		ac.open = false
		if ac.onSelect != nil {
			ac.onSelect(ac.text)
		}
	}
}

func defaultAutoFilter(input string, options []string) []string {
	if input == "" {
		return nil
	}
	var result []string
	for _, opt := range options {
		if len(opt) >= len(input) && containsPrefix(opt, input) {
			result = append(result, opt)
		}
	}
	return result
}

func containsPrefix(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		a, b := s[i], prefix[i]
		if a >= 'A' && a <= 'Z' {
			a += 32
		}
		if b >= 'A' && b <= 'Z' {
			b += 32
		}
		if a != b {
			return false
		}
	}
	return true
}

func (ac *AutoComplete) Draw(buf *render.CommandBuffer) {
	bounds := ac.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := ac.config

	// Input field
	buf.DrawRect(render.RectCmd{
		Bounds:      bounds,
		FillColor:   cfg.BgColor,
		BorderColor: cfg.BorderColor,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 1, 1)
	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		cfg.TextRenderer.DrawText(buf, ac.text, bounds.X+cfg.SpaceSM, bounds.Y+(bounds.Height-lh)/2, cfg.FontSize, bounds.Width-cfg.SpaceSM*2, cfg.TextColor, 1)
	}

	// Dropdown
	if !ac.open || len(ac.filtered) == 0 {
		return
	}
	itemH := float32(32)
	dropH := float32(len(ac.filtered)) * itemH
	if dropH > 200 {
		dropH = 200
	}
	dx := bounds.X
	dy := bounds.Y + bounds.Height + 4

	buf.DrawOverlay(render.RectCmd{
		Bounds:      uimath.NewRect(dx, dy, bounds.Width, dropH),
		FillColor:   uimath.ColorWhite,
		BorderColor: cfg.BorderColor,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 40, 1)

	for i, item := range ac.filtered {
		iy := dy + float32(i)*itemH
		if iy+itemH > dy+dropH {
			break
		}
		if i == ac.selected {
			buf.DrawOverlay(render.RectCmd{
				Bounds:    uimath.NewRect(dx, iy, bounds.Width, itemH),
				FillColor: uimath.RGBA(0, 0, 0, 0.04),
			}, 41, 1)
		}
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			cfg.TextRenderer.DrawText(buf, item, dx+cfg.SpaceSM, iy+(itemH-lh)/2, cfg.FontSize, bounds.Width-cfg.SpaceSM*2, cfg.TextColor, 1)
		}
	}
}

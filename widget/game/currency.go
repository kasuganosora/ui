package game

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// CurrencyEntry represents a single currency type.
type CurrencyEntry struct {
	Icon   render.TextureHandle
	Symbol string
	Amount int
	Color  uimath.Color
}

// CurrencyDisplay shows one or more currency values.
type CurrencyDisplay struct {
	widget.Base
	currencies []CurrencyEntry
	gap        float32
}

func NewCurrencyDisplay(tree *core.Tree, cfg *widget.Config) *CurrencyDisplay {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	return &CurrencyDisplay{
		Base: widget.NewBase(tree, core.TypeCustom, cfg),
		gap:  12,
	}
}

func (cd *CurrencyDisplay) Currencies() []CurrencyEntry { return cd.currencies }
func (cd *CurrencyDisplay) SetGap(g float32)            { cd.gap = g }

func (cd *CurrencyDisplay) AddCurrency(c CurrencyEntry) {
	cd.currencies = append(cd.currencies, c)
}

func (cd *CurrencyDisplay) SetAmount(index, amount int) {
	if index >= 0 && index < len(cd.currencies) {
		cd.currencies[index].Amount = amount
	}
}

func (cd *CurrencyDisplay) ClearCurrencies() {
	cd.currencies = cd.currencies[:0]
}

func (cd *CurrencyDisplay) Draw(buf *render.CommandBuffer) {
	bounds := cd.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := cd.Config()
	x := bounds.X

	for _, c := range cd.currencies {
		color := c.Color
		if color.A == 0 {
			color = uimath.ColorHex("#ffd700")
		}

		// Symbol
		text := c.Symbol + " " + formatAmount(c.Amount)
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			tw := cfg.TextRenderer.MeasureText(text, cfg.FontSize)
			cfg.TextRenderer.DrawText(buf, text, x, bounds.Y+(bounds.Height-lh)/2, cfg.FontSize, tw+4, color, 1)
			x += tw + cd.gap
		} else {
			x += 80 + cd.gap
		}
	}
}

func formatAmount(n int) string {
	if n < 1000 {
		return itoa(n)
	}
	if n < 1000000 {
		return itoa(n/1000) + "." + itoa((n%1000)/100) + "K"
	}
	return itoa(n/1000000) + "." + itoa((n%1000000)/100000) + "M"
}

package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// TransferItem represents an item that can be transferred.
type TransferItem struct {
	Key   string
	Label string
}

// Transfer is a dual-list transfer widget.
type Transfer struct {
	Base
	source   []TransferItem
	target   []TransferItem
	onChange func(targetKeys []string)
}

func NewTransfer(tree *core.Tree, cfg *Config) *Transfer {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Transfer{
		Base: NewBase(tree, core.TypeCustom, cfg),
	}
}

func (t *Transfer) Source() []TransferItem { return t.source }
func (t *Transfer) Target() []TransferItem { return t.target }
func (t *Transfer) OnChange(fn func([]string)) { t.onChange = fn }

func (t *Transfer) SetSource(items []TransferItem) {
	t.source = make([]TransferItem, len(items))
	copy(t.source, items)
}

func (t *Transfer) MoveToTarget(keys []string) {
	keySet := make(map[string]bool, len(keys))
	for _, k := range keys {
		keySet[k] = true
	}
	var remaining []TransferItem
	for _, item := range t.source {
		if keySet[item.Key] {
			t.target = append(t.target, item)
		} else {
			remaining = append(remaining, item)
		}
	}
	t.source = remaining
	t.fireChange()
}

func (t *Transfer) MoveToSource(keys []string) {
	keySet := make(map[string]bool, len(keys))
	for _, k := range keys {
		keySet[k] = true
	}
	var remaining []TransferItem
	for _, item := range t.target {
		if keySet[item.Key] {
			t.source = append(t.source, item)
		} else {
			remaining = append(remaining, item)
		}
	}
	t.target = remaining
	t.fireChange()
}

func (t *Transfer) fireChange() {
	if t.onChange != nil {
		keys := make([]string, len(t.target))
		for i, item := range t.target {
			keys[i] = item.Key
		}
		t.onChange(keys)
	}
}

func (t *Transfer) Draw(buf *render.CommandBuffer) {
	bounds := t.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := t.config
	halfW := (bounds.Width - 40) / 2 // 40px gap for arrows
	itemH := float32(28)

	// Left panel (source)
	t.drawPanel(buf, bounds.X, bounds.Y, halfW, bounds.Height, t.source, cfg)

	// Arrow area
	arrowX := bounds.X + halfW + 12
	arrowY := bounds.Y + bounds.Height/2
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(arrowX, arrowY-1, 16, 2),
		FillColor: cfg.TextColor,
	}, 2, 0.5)

	// Right panel (target)
	t.drawPanel(buf, bounds.X+halfW+40, bounds.Y, halfW, bounds.Height, t.target, cfg)
	_ = itemH
}

func (t *Transfer) drawPanel(buf *render.CommandBuffer, x, y, w, h float32, items []TransferItem, cfg *Config) {
	// Panel border
	buf.DrawRect(render.RectCmd{
		Bounds:      uimath.NewRect(x, y, w, h),
		FillColor:   cfg.BgColor,
		BorderColor: cfg.BorderColor,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 1, 1)

	// Items
	itemH := float32(28)
	for i, item := range items {
		iy := y + float32(i)*itemH
		if iy+itemH > y+h {
			break
		}
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			cfg.TextRenderer.DrawText(buf, item.Label, x+cfg.SpaceSM, iy+(itemH-lh)/2, cfg.FontSizeSm, w-cfg.SpaceSM*2, cfg.TextColor, 1)
		}
	}
}

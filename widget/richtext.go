package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// RichSpanType identifies the kind of rich text span.
type RichSpanType uint8

const (
	RichSpanText  RichSpanType = iota // Text with styling
	RichSpanImage                      // Inline image
	RichSpanBreak                      // Line break
)

// RichSpan is a single styled fragment within rich text.
type RichSpan struct {
	Type     RichSpanType
	Text     string
	FontSize float32      // 0 = use default
	Color    uimath.Color // zero value = use default
	Bold     bool
	Italic   bool
	// For RichSpanImage
	Texture render.TextureHandle
	ImgW    float32
	ImgH    float32
}

// RichText displays mixed-style text with inline images.
type RichText struct {
	Base
	spans     []RichSpan
	lineSpace float32 // extra vertical spacing between lines
}

// NewRichText creates a rich text widget.
func NewRichText(tree *core.Tree, cfg *Config) *RichText {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &RichText{
		Base:      NewBase(tree, core.TypeCustom, cfg),
		lineSpace: 4,
	}
}

// SetSpans sets the content spans.
func (rt *RichText) SetSpans(spans []RichSpan) {
	rt.spans = make([]RichSpan, len(spans))
	copy(rt.spans, spans)
}

// AddSpan appends a span.
func (rt *RichText) AddSpan(span RichSpan) {
	rt.spans = append(rt.spans, span)
}

// AddText appends a text span with default styling.
func (rt *RichText) AddText(text string) {
	rt.spans = append(rt.spans, RichSpan{Type: RichSpanText, Text: text})
}

// AddStyledText appends a text span with custom color and size.
func (rt *RichText) AddStyledText(text string, color uimath.Color, fontSize float32, bold bool) {
	rt.spans = append(rt.spans, RichSpan{
		Type:     RichSpanText,
		Text:     text,
		Color:    color,
		FontSize: fontSize,
		Bold:     bold,
	})
}

// AddImage appends an inline image span.
func (rt *RichText) AddImage(tex render.TextureHandle, w, h float32) {
	rt.spans = append(rt.spans, RichSpan{
		Type:    RichSpanImage,
		Texture: tex,
		ImgW:    w,
		ImgH:    h,
	})
}

// AddBreak appends a line break.
func (rt *RichText) AddBreak() {
	rt.spans = append(rt.spans, RichSpan{Type: RichSpanBreak})
}

// ClearSpans removes all spans.
func (rt *RichText) ClearSpans() {
	rt.spans = rt.spans[:0]
}

// Spans returns the current spans.
func (rt *RichText) Spans() []RichSpan { return rt.spans }

// SetLineSpacing sets extra vertical space between lines.
func (rt *RichText) SetLineSpacing(s float32) { rt.lineSpace = s }

// Draw renders the rich text content.
func (rt *RichText) Draw(buf *render.CommandBuffer) {
	bounds := rt.Bounds()
	if bounds.IsEmpty() || len(rt.spans) == 0 {
		return
	}
	cfg := rt.config

	x := bounds.X
	y := bounds.Y
	maxW := bounds.Width

	defaultSize := cfg.FontSize
	defaultColor := cfg.TextColor
	lh := defaultSize * cfg.LineHeight

	for _, span := range rt.spans {
		switch span.Type {
		case RichSpanBreak:
			x = bounds.X
			y += lh + rt.lineSpace
			continue

		case RichSpanImage:
			imgW := span.ImgW
			imgH := span.ImgH
			if imgW <= 0 {
				imgW = cfg.IconSize
			}
			if imgH <= 0 {
				imgH = cfg.IconSize
			}
			// Wrap to next line if needed
			if x+imgW > bounds.X+maxW && x > bounds.X {
				x = bounds.X
				y += lh + rt.lineSpace
			}
			if span.Texture != 0 {
				buf.DrawImage(render.ImageCmd{
					Texture: span.Texture,
					DstRect: uimath.NewRect(x, y+(lh-imgH)/2, imgW, imgH),
					Tint:    uimath.ColorWhite,
				}, 0, 1)
			} else {
				// Placeholder rect for missing texture
				buf.DrawRect(render.RectCmd{
					Bounds:      uimath.NewRect(x, y+(lh-imgH)/2, imgW, imgH),
					FillColor:   uimath.RGBA(0.8, 0.8, 0.8, 1),
					BorderColor: cfg.BorderColor,
					BorderWidth: 1,
				}, 0, 1)
			}
			x += imgW + 2
			continue

		case RichSpanText:
			fontSize := span.FontSize
			if fontSize <= 0 {
				fontSize = defaultSize
			}
			color := span.Color
			if color.A <= 0 {
				color = defaultColor
			}
			// Bold: slightly larger as approximation (no real bold font support yet)
			if span.Bold {
				fontSize *= 1.05
			}

			if cfg.TextRenderer != nil {
				tw := cfg.TextRenderer.MeasureText(span.Text, fontSize)
				spanLH := cfg.TextRenderer.LineHeight(fontSize)

				// Simple word wrapping
				if x+tw > bounds.X+maxW && x > bounds.X {
					x = bounds.X
					y += spanLH + rt.lineSpace
				}
				// If still too wide, draw anyway (single long word)
				cfg.TextRenderer.DrawText(buf, span.Text, x, y+(lh-spanLH)/2, fontSize, maxW-(x-bounds.X), color, 1)
				x += tw
			} else {
				// Placeholder: approximate text with a colored rect
				tw := float32(len(span.Text)) * fontSize * 0.5
				if tw > maxW {
					tw = maxW
				}
				if x+tw > bounds.X+maxW && x > bounds.X {
					x = bounds.X
					y += lh + rt.lineSpace
				}
				h := fontSize
				buf.DrawRect(render.RectCmd{
					Bounds:    uimath.NewRect(x, y+(lh-h)/2, tw, h),
					FillColor: color.WithAlpha(0.3),
				}, 0, 1)
				x += tw
			}
		}
	}
}

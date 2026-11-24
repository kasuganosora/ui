package widget

import (
	"strings"
	"unicode/utf8"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// TextBlock is one laid-out piece of a markdown subset: a heading, list item,
// fenced or inline code, a link, or plain text. Bounds are in screen space.
type TextBlock struct {
	Kind   string
	Level  int
	Text   string
	URL    string
	Inline bool
	Bounds uimath.Rect
}

// TextView draws a hand-written markdown subset. Headings, lists, fenced code,
// inline code, and links are laid out from the widget bounds. It does not use
// goldmark, and it does not parse tables or images.
type TextView struct {
	Base
	source string
	onLink func(string)
}

// NewTextView creates an empty document view.
func NewTextView(tree *core.Tree, cfg *Config) *TextView {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	v := &TextView{Base: NewBase(tree, core.TypeCustom, cfg)}
	tree.AddHandler(v.id, event.MouseClick, func(e *event.Event) {
		if e.Phase == event.PhaseCapture || v.onLink == nil {
			return
		}
		if url, ok := v.LinkAt(e.GlobalX, e.GlobalY); ok {
			v.onLink(url)
		}
	})
	return v
}

// SetMarkdown replaces the document. Previous link targets are dropped.
func (v *TextView) SetMarkdown(src string) {
	v.source = strings.ReplaceAll(src, "\r\n", "\n")
}

// OnLink receives a click on a link target.
func (v *TextView) OnLink(fn func(string)) { v.onLink = fn }

// Blocks returns the current layout. Positions follow the widget bounds.
func (v *TextView) Blocks() []TextBlock { return v.layout(v.Bounds()) }

// LinkAt reports the link under a screen point.
func (v *TextView) LinkAt(x, y float32) (string, bool) {
	for _, b := range v.layout(v.Bounds()) {
		if b.Kind == "link" && b.Bounds.Contains(uimath.Vec2{X: x, Y: y}) {
			return b.URL, true
		}
	}
	return "", false
}

func (v *TextView) Draw(buf *render.CommandBuffer) {
	b := v.Bounds()
	if b.IsEmpty() || v.source == "" {
		return
	}
	cfg := v.config
	buf.PushClip(b)
	for _, block := range v.layout(b) {
		switch block.Kind {
		case "code":
			if !block.Inline {
				buf.DrawRect(render.RectCmd{Bounds: block.Bounds, FillColor: cfg.SuccessColor}, 1, 1)
				v.paintLines(buf, block.Text, block.Bounds.X+8, block.Bounds.Y+4, cfg.FontSize, cfg.TextColor)
			} else {
				buf.DrawRect(render.RectCmd{Bounds: block.Bounds, FillColor: cfg.BorderColor}, 1, 1)
				emitText(cfg, buf, block.Text, block.Bounds.X, block.Bounds.Y, cfg.FontSize, block.Bounds.Width, cfg.TextColor, 1, 4)
			}
		case "link":
			emitText(cfg, buf, block.Text, block.Bounds.X, block.Bounds.Y, cfg.FontSize, block.Bounds.Width+1, cfg.PrimaryColor, 1, 4)
		case "heading":
			emitText(cfg, buf, block.Text, block.Bounds.X, block.Bounds.Y, cfg.FontSizeLg, block.Bounds.Width, cfg.TextColor, 1, 4)
		case "list":
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(block.Bounds.X, block.Bounds.Y+6, 6, 6),
				FillColor: cfg.PrimaryColor,
			}, 2, 1)
			emitText(cfg, buf, block.Text, block.Bounds.X+16, block.Bounds.Y, cfg.FontSize, block.Bounds.Width-16, cfg.TextColor, 1, 4)
		default:
			emitText(cfg, buf, block.Text, block.Bounds.X, block.Bounds.Y, cfg.FontSize, block.Bounds.Width+1, cfg.TextColor, 1, 4)
		}
	}
	buf.PopClip()
}

func (v *TextView) paintLines(buf *render.CommandBuffer, text string, x, y, size float32, color uimath.Color) {
	lineH := v.config.FontSize * 1.3
	if lineH < 16 {
		lineH = 16
	}
	for _, line := range strings.Split(text, "\n") {
		emitText(v.config, buf, line, x, y, size, v.Bounds().Width, color, 1, 4)
		y += lineH
	}
}

func (v *TextView) measure(text string, size float32) float32 {
	if size <= 0 {
		size = v.config.FontSize
	}
	if v.config.TextRenderer != nil {
		return v.config.TextRenderer.MeasureText(text, size)
	}
	return float32(utf8.RuneCountInString(text)) * size * 0.55
}

type mdLine struct {
	kind  string
	level int
	text  string
}

func (v *TextView) layout(bounds uimath.Rect) []TextBlock {
	if bounds.IsEmpty() || v.source == "" {
		return nil
	}
	lines := parseMarkdown(v.source)
	lineH := v.config.FontSize * 1.3
	if lineH < 16 {
		lineH = 16
	}
	y := bounds.Y
	var out []TextBlock
	for _, line := range lines {
		switch line.kind {
		case "blank":
			y += lineH * 0.4
		case "heading":
			h := lineH * 1.6
			out = append(out, TextBlock{
				Kind: "heading", Level: line.level, Text: line.text,
				Bounds: uimath.NewRect(bounds.X, y, bounds.Width, h),
			})
			y += h
		case "list":
			out = append(out, TextBlock{
				Kind: "list", Text: line.text,
				Bounds: uimath.NewRect(bounds.X, y, bounds.Width, lineH),
			})
			y += lineH
		case "code":
			n := strings.Count(line.text, "\n") + 1
			if line.text == "" {
				n = 1
			}
			h := float32(n)*lineH + 8
			out = append(out, TextBlock{
				Kind: "code", Text: line.text,
				Bounds: uimath.NewRect(bounds.X, y, bounds.Width, h),
			})
			y += h
		case "paragraph":
			out = append(out, v.placeParagraph(bounds, line.text, y, lineH)...)
			if len(out) > 0 {
				last := out[len(out)-1]
				y = last.Bounds.Y + last.Bounds.Height
			}
		}
	}
	return out
}

func (v *TextView) placeParagraph(bounds uimath.Rect, text string, y, lineH float32) []TextBlock {
	segs := tokenizeInline(text)
	x := bounds.X
	right := bounds.X + bounds.Width
	var out []TextBlock
	advance := func(w float32) {
		if x > bounds.X && x+w > right {
			x = bounds.X
			y += lineH
		}
	}
	for _, seg := range segs {
		if seg.kind == "text" {
			for _, word := range wrapWords(seg.text) {
				w := v.measure(word, v.config.FontSize)
				advance(w)
				out = append(out, TextBlock{
					Kind: "text", Text: word,
					Bounds: uimath.NewRect(x, y, w, lineH),
				})
				x += w
			}
			continue
		}
		w := v.measure(seg.text, v.config.FontSize)
		if w < 8 {
			w = 8
		}
		advance(w)
		out = append(out, TextBlock{
			Kind: seg.kind, Text: seg.text, URL: seg.url, Inline: seg.kind == "code",
			Bounds: uimath.NewRect(x, y, w, lineH),
		})
		x += w
	}
	return out
}

func wrapWords(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	start := 0
	for i, r := range s {
		if r == ' ' {
			if i > start {
				out = append(out, s[start:i])
			}
			out = append(out, " ")
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}

type inlineSeg struct {
	kind string
	text string
	url  string
}

func tokenizeInline(s string) []inlineSeg {
	var out []inlineSeg
	var b strings.Builder
	flush := func() {
		if b.Len() == 0 {
			return
		}
		out = append(out, inlineSeg{kind: "text", text: b.String()})
		b.Reset()
	}
	for i := 0; i < len(s); {
		if s[i] == '`' {
			flush()
			j := strings.IndexByte(s[i+1:], '`')
			if j < 0 {
				b.WriteByte(s[i])
				i++
				continue
			}
			out = append(out, inlineSeg{kind: "code", text: s[i+1 : i+1+j]})
			i += j + 2
			continue
		}
		if s[i] == '[' {
			end := strings.IndexByte(s[i:], ']')
			if end > 1 && i+end+1 < len(s) && s[i+end+1] == '(' {
				close := strings.IndexByte(s[i+end+2:], ')')
				if close >= 0 {
					flush()
					label := s[i+1 : i+end]
					url := s[i+end+2 : i+end+2+close]
					out = append(out, inlineSeg{kind: "link", text: label, url: url})
					i += end + 2 + close + 1
					continue
				}
			}
		}
		b.WriteByte(s[i])
		i++
	}
	flush()
	return out
}

func parseMarkdown(src string) []mdLine {
	raw := strings.Split(src, "\n")
	var out []mdLine
	var para []string
	flushPara := func() {
		if len(para) == 0 {
			return
		}
		out = append(out, mdLine{kind: "paragraph", text: strings.Join(para, " ")})
		para = nil
	}
	for i := 0; i < len(raw); i++ {
		line := raw[i]
		trim := strings.TrimSpace(line)
		if trim == "```" {
			flushPara()
			var body []string
			i++
			for i < len(raw) && strings.TrimSpace(raw[i]) != "```" {
				body = append(body, raw[i])
				i++
			}
			out = append(out, mdLine{kind: "code", text: strings.Join(body, "\n")})
			continue
		}
		if trim == "" {
			flushPara()
			out = append(out, mdLine{kind: "blank"})
			continue
		}
		if n, text, ok := headingLine(trim); ok {
			flushPara()
			out = append(out, mdLine{kind: "heading", level: n, text: text})
			continue
		}
		if strings.HasPrefix(trim, "- ") || strings.HasPrefix(trim, "* ") {
			flushPara()
			out = append(out, mdLine{kind: "list", text: trim[2:]})
			continue
		}
		para = append(para, trim)
	}
	flushPara()
	return out
}

func headingLine(trim string) (int, string, bool) {
	n := 0
	for n < len(trim) && trim[n] == '#' && n < 3 {
		n++
	}
	if n == 0 || n >= len(trim) || trim[n] != ' ' {
		return 0, "", false
	}
	return n, strings.TrimSpace(trim[n+1:]), true
}

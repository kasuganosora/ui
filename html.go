package ui

import (
	"strconv"
	"strings"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/widget"
)

// LoadHTML parses a simple HTML string and builds a widget tree.
// Supported tags: div, span, button, input, p, h1-h6, img, br.
// Inline styles are parsed from the style attribute.
func LoadHTML(tree *core.Tree, cfg *widget.Config, html string) widget.Widget {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	p := &htmlParser{tree: tree, cfg: cfg, src: html, pos: 0}
	return p.parse()
}

type htmlParser struct {
	tree *core.Tree
	cfg  *widget.Config
	src  string
	pos  int
}

func (p *htmlParser) parse() widget.Widget {
	root := widget.NewDiv(p.tree, p.cfg)
	p.parseChildren(root)
	return root
}

func (p *htmlParser) parseChildren(parent *widget.Div) {
	for p.pos < len(p.src) {
		p.skipWhitespace()
		if p.pos >= len(p.src) {
			break
		}
		if p.src[p.pos] == '<' {
			if p.pos+1 < len(p.src) && p.src[p.pos+1] == '/' {
				return // closing tag
			}
			p.parseElement(parent)
		} else {
			// Text content
			text := p.readUntil('<')
			text = strings.TrimSpace(text)
			if text != "" {
				t := widget.NewText(p.tree, text, p.cfg)
				parent.AppendChild(t)
			}
		}
	}
}

func (p *htmlParser) parseElement(parent *widget.Div) {
	p.expect('<')
	tag := p.readTagName()
	attrs := p.readAttributes()
	selfClose := false
	if p.pos < len(p.src) && p.src[p.pos] == '/' {
		selfClose = true
		p.pos++
	}
	p.expect('>')

	switch tag {
	case "br":
		return
	case "img":
		img := widget.NewImage(p.tree, 0, p.cfg)
		_ = attrs["src"] // src stored as property for later resolution
		applyInlineStyle(img, attrs["style"])
		parent.AppendChild(img)
		return
	case "input":
		inp := widget.NewInput(p.tree, p.cfg)
		if ph, ok := attrs["placeholder"]; ok {
			inp.SetPlaceholder(ph)
		}
		if v, ok := attrs["value"]; ok {
			inp.SetValue(v)
		}
		if _, ok := attrs["disabled"]; ok {
			inp.SetDisabled(true)
		}
		parent.AppendChild(inp)
		return
	}

	if selfClose {
		return
	}

	switch tag {
	case "button":
		label := p.readTextContent(tag)
		btn := widget.NewButton(p.tree, label, p.cfg)
		if _, ok := attrs["disabled"]; ok {
			btn.SetDisabled(true)
		}
		parent.AppendChild(btn)

	case "p", "span":
		text := p.readTextContent(tag)
		t := widget.NewText(p.tree, text, p.cfg)
		applyTextStyle(t, attrs["style"])
		parent.AppendChild(t)

	case "h1", "h2", "h3", "h4", "h5", "h6":
		text := p.readTextContent(tag)
		t := widget.NewText(p.tree, text, p.cfg)
		level := int(tag[1] - '0')
		sizes := []float32{32, 28, 24, 20, 16, 14}
		if level >= 1 && level <= 6 {
			t.SetFontSize(sizes[level-1])
		}
		parent.AppendChild(t)

	default: // div and unknown tags
		d := widget.NewDiv(p.tree, p.cfg)
		applyDivStyle(d, attrs["style"])
		if cls, ok := attrs["class"]; ok {
			p.tree.SetClasses(d.ElementID(), strings.Fields(cls))
		}
		p.parseChildren(d)
		parent.AppendChild(d)
		p.skipClosingTag(tag)
	}
}

func (p *htmlParser) readTextContent(tag string) string {
	start := p.pos
	closing := "</" + tag + ">"
	idx := strings.Index(p.src[p.pos:], closing)
	if idx < 0 {
		p.pos = len(p.src)
		return strings.TrimSpace(p.src[start:])
	}
	text := p.src[start : p.pos+idx]
	p.pos += idx + len(closing)
	return strings.TrimSpace(text)
}

func (p *htmlParser) readTagName() string {
	start := p.pos
	for p.pos < len(p.src) && p.src[p.pos] != ' ' && p.src[p.pos] != '>' && p.src[p.pos] != '/' {
		p.pos++
	}
	return strings.ToLower(p.src[start:p.pos])
}

func (p *htmlParser) readAttributes() map[string]string {
	attrs := make(map[string]string)
	for {
		p.skipWhitespace()
		if p.pos >= len(p.src) || p.src[p.pos] == '>' || p.src[p.pos] == '/' {
			break
		}
		key := p.readAttrName()
		if key == "" {
			break
		}
		p.skipWhitespace()
		if p.pos < len(p.src) && p.src[p.pos] == '=' {
			p.pos++
			p.skipWhitespace()
			attrs[key] = p.readAttrValue()
		} else {
			attrs[key] = key // boolean attribute
		}
	}
	return attrs
}

func (p *htmlParser) readAttrName() string {
	start := p.pos
	for p.pos < len(p.src) && p.src[p.pos] != '=' && p.src[p.pos] != ' ' && p.src[p.pos] != '>' && p.src[p.pos] != '/' {
		p.pos++
	}
	return strings.ToLower(p.src[start:p.pos])
}

func (p *htmlParser) readAttrValue() string {
	if p.pos >= len(p.src) {
		return ""
	}
	if p.src[p.pos] == '"' || p.src[p.pos] == '\'' {
		quote := p.src[p.pos]
		p.pos++
		start := p.pos
		for p.pos < len(p.src) && p.src[p.pos] != quote {
			p.pos++
		}
		val := p.src[start:p.pos]
		if p.pos < len(p.src) {
			p.pos++
		}
		return val
	}
	return p.readUntilAny(" >")
}

func (p *htmlParser) readUntil(ch byte) string {
	start := p.pos
	for p.pos < len(p.src) && p.src[p.pos] != ch {
		p.pos++
	}
	return p.src[start:p.pos]
}

func (p *htmlParser) readUntilAny(chars string) string {
	start := p.pos
	for p.pos < len(p.src) && !strings.ContainsRune(chars, rune(p.src[p.pos])) {
		p.pos++
	}
	return p.src[start:p.pos]
}

func (p *htmlParser) expect(ch byte) {
	if p.pos < len(p.src) && p.src[p.pos] == ch {
		p.pos++
	}
}

func (p *htmlParser) skipWhitespace() {
	for p.pos < len(p.src) && (p.src[p.pos] == ' ' || p.src[p.pos] == '\n' || p.src[p.pos] == '\r' || p.src[p.pos] == '\t') {
		p.pos++
	}
}

func (p *htmlParser) skipClosingTag(tag string) {
	closing := "</" + tag + ">"
	if p.pos+len(closing) <= len(p.src) && strings.ToLower(p.src[p.pos:p.pos+len(closing)]) == closing {
		p.pos += len(closing)
	}
}

// --- Inline CSS parsing ---

func parseInlineCSS(style string) map[string]string {
	props := make(map[string]string)
	for _, decl := range strings.Split(style, ";") {
		decl = strings.TrimSpace(decl)
		if decl == "" {
			continue
		}
		parts := strings.SplitN(decl, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			props[key] = val
		}
	}
	return props
}

func applyDivStyle(d *widget.Div, style string) {
	if style == "" {
		return
	}
	props := parseInlineCSS(style)
	for key, val := range props {
		switch key {
		case "background-color", "background":
			d.SetBgColor(parseColor(val))
		case "border-radius":
			if v, err := parsePx(val); err == nil {
				d.SetBorderRadius(v)
			}
		case "border-color":
			d.SetBorderColor(parseColor(val))
		case "border-width":
			if v, err := parsePx(val); err == nil {
				d.SetBorderWidth(v)
			}
		case "display":
			s := d.Style()
			switch val {
			case "flex":
				s.Display = layout.DisplayFlex
			case "grid":
				s.Display = layout.DisplayGrid
			case "none":
				s.Display = layout.DisplayNone
			}
			d.SetStyle(s)
		case "flex-direction":
			s := d.Style()
			switch val {
			case "row":
				s.FlexDirection = layout.FlexDirectionRow
			case "column":
				s.FlexDirection = layout.FlexDirectionColumn
			}
			d.SetStyle(s)
		case "gap":
			s := d.Style()
			if v, err := parsePx(val); err == nil {
				s.Gap = v
			}
			d.SetStyle(s)
		case "width":
			s := d.Style()
			if v, err := parsePx(val); err == nil {
				s.Width = layout.Px(v)
			}
			d.SetStyle(s)
		case "height":
			s := d.Style()
			if v, err := parsePx(val); err == nil {
				s.Height = layout.Px(v)
			}
			d.SetStyle(s)
		case "padding":
			s := d.Style()
			if v, err := parsePx(val); err == nil {
				s.Padding = layout.EdgeValues{
					Top: layout.Px(v), Right: layout.Px(v),
					Bottom: layout.Px(v), Left: layout.Px(v),
				}
			}
			d.SetStyle(s)
		}
	}
}

func applyInlineStyle(w widget.Widget, style string) {
	if style == "" {
		return
	}
	props := parseInlineCSS(style)
	s := w.Style()
	for key, val := range props {
		switch key {
		case "width":
			if v, err := parsePx(val); err == nil {
				s.Width = layout.Px(v)
			}
		case "height":
			if v, err := parsePx(val); err == nil {
				s.Height = layout.Px(v)
			}
		}
	}
	w.SetStyle(s)
}

func applyTextStyle(t *widget.Text, style string) {
	if style == "" {
		return
	}
	props := parseInlineCSS(style)
	for key, val := range props {
		switch key {
		case "color":
			t.SetColor(parseColor(val))
		case "font-size":
			if v, err := parsePx(val); err == nil {
				t.SetFontSize(v)
			}
		}
	}
}

func parseColor(val string) uimath.Color {
	if strings.HasPrefix(val, "#") {
		return uimath.ColorHex(val)
	}
	switch val {
	case "white":
		return uimath.ColorWhite
	case "black":
		return uimath.Color{R: 0, G: 0, B: 0, A: 1}
	case "red":
		return uimath.Color{R: 1, G: 0, B: 0, A: 1}
	case "green":
		return uimath.Color{R: 0, G: 0.5, B: 0, A: 1}
	case "blue":
		return uimath.Color{R: 0, G: 0, B: 1, A: 1}
	case "transparent":
		return uimath.Color{}
	}
	return uimath.Color{}
}

func parsePx(val string) (float32, error) {
	val = strings.TrimSuffix(val, "px")
	val = strings.TrimSpace(val)
	f, err := strconv.ParseFloat(val, 32)
	return float32(f), err
}

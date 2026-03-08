package ui

import (
	"strconv"
	"strings"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/css"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/widget"
)

// Document represents a parsed HTML document with query capabilities.
type Document struct {
	Root    *widget.Div
	ids     map[string]widget.Widget
	classes map[string][]widget.Widget
	tags    map[string][]widget.Widget
}

// QueryByID returns the widget with the given HTML id attribute, or nil.
func (d *Document) QueryByID(id string) widget.Widget {
	return d.ids[id]
}

// QueryByClass returns all widgets with the given CSS class.
func (d *Document) QueryByClass(class string) []widget.Widget {
	return d.classes[class]
}

// QueryByTag returns all widgets with the given HTML tag name.
func (d *Document) QueryByTag(tag string) []widget.Widget {
	return d.tags[tag]
}

// OnClick binds a click handler to a button identified by HTML id.
func (d *Document) OnClick(id string, fn func()) {
	if btn, ok := d.QueryByID(id).(*widget.Button); ok {
		btn.OnClick(fn)
	}
}

// OnChange binds a change handler to an input/textarea/select identified by HTML id.
// The handler receives the new string value.
func (d *Document) OnChange(id string, fn func(string)) {
	w := d.QueryByID(id)
	if w == nil {
		return
	}
	switch v := w.(type) {
	case *widget.Input:
		v.OnChange(fn)
	case *widget.TextArea:
		v.OnChange(fn)
	case *widget.Select:
		v.OnChange(fn)
	}
}

// OnToggle binds a toggle handler to a checkbox/switch identified by HTML id.
// The handler receives the new boolean value.
func (d *Document) OnToggle(id string, fn func(bool)) {
	w := d.QueryByID(id)
	if w == nil {
		return
	}
	switch v := w.(type) {
	case *widget.Checkbox:
		v.OnChange(fn)
	case *widget.Switch:
		v.OnChange(fn)
	}
}

// LoadHTML parses a simple HTML string and builds a widget tree.
// Supported tags: div, span, button, input, p, h1-h6, img, br, a, select, textarea,
// header, footer, aside, main, nav, section, article, space, row, col, layout,
// checkbox, switch, radio, tag, progress, message, empty, loading, tooltip.
// Inline styles are parsed from the style attribute.
// If the HTML contains <style> blocks, CSS rules are extracted and applied.
func LoadHTML(tree *core.Tree, cfg *widget.Config, html string) widget.Widget {
	doc := LoadHTMLDocument(tree, cfg, html, "")
	return doc.Root
}

// LoadHTMLWithCSS parses HTML with an external CSS stylesheet.
// The CSS rules are applied to elements based on selectors.
func LoadHTMLWithCSS(tree *core.Tree, cfg *widget.Config, html, cssText string) widget.Widget {
	doc := LoadHTMLDocument(tree, cfg, html, cssText)
	return doc.Root
}

// LoadHTMLDocument parses HTML with optional CSS and returns a Document for querying.
func LoadHTMLDocument(tree *core.Tree, cfg *widget.Config, html, cssText string) *Document {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	var sheet *css.Stylesheet
	if cssText != "" {
		sheet = css.Parse(cssText)
	}
	p := &htmlParser{
		tree:  tree,
		cfg:   cfg,
		src:   html,
		pos:   0,
		sheet: sheet,
		doc: &Document{
			ids:     make(map[string]widget.Widget),
			classes: make(map[string][]widget.Widget),
			tags:    make(map[string][]widget.Widget),
		},
	}
	p.doc.Root = p.parse()
	return p.doc
}

type htmlParser struct {
	tree  *core.Tree
	cfg   *widget.Config
	src   string
	pos   int
	sheet *css.Stylesheet
	doc   *Document
	// widgetInfo tracks element info for CSS matching after tree construction.
	widgetInfo []widgetStyleInfo
	// radioGroups maps group name to RadioGroup for radio button grouping.
	radioGroups map[string]*widget.RadioGroup
}

// widgetStyleInfo records a widget and its CSS-matching metadata.
type widgetStyleInfo struct {
	widget    widget.Widget
	tag       string
	id        string
	classes   []string
	inlineCSS string
	ancestors []css.ElementInfo
}

func (p *htmlParser) parse() *widget.Div {
	// Extract <style> blocks before parsing
	p.extractStyleBlocks()

	root := widget.NewDiv(p.tree, p.cfg)
	p.parseChildren(root, nil)

	// Apply CSS rules to all collected widgets
	if p.sheet != nil && len(p.sheet.Rules) > 0 {
		p.applyCSS()
	}
	return root
}

// extractStyleBlocks finds and parses all <style>...</style> blocks.
func (p *htmlParser) extractStyleBlocks() {
	src := p.src
	var cssText strings.Builder
	for {
		idx := strings.Index(strings.ToLower(src), "<style")
		if idx < 0 {
			break
		}
		// Find end of opening tag
		closeTag := strings.Index(src[idx:], ">")
		if closeTag < 0 {
			break
		}
		contentStart := idx + closeTag + 1
		endTag := strings.Index(strings.ToLower(src[contentStart:]), "</style>")
		if endTag < 0 {
			break
		}
		cssText.WriteString(src[contentStart : contentStart+endTag])
		cssText.WriteByte('\n')

		// Remove the <style> block from src
		blockEnd := contentStart + endTag + len("</style>")
		src = src[:idx] + src[blockEnd:]
	}
	p.src = src

	if cssText.Len() > 0 {
		parsed := css.Parse(cssText.String())
		if p.sheet == nil {
			p.sheet = parsed
		} else {
			// Merge: append rules and variables
			for k, v := range parsed.Variables {
				p.sheet.Variables[k] = v
			}
			p.sheet.Rules = append(p.sheet.Rules, parsed.Rules...)
		}
	}
}

// containerWidget is any widget that supports AppendChild.
type containerWidget interface {
	widget.Widget
	AppendChild(child widget.Widget)
}

func (p *htmlParser) parseChildren(parent containerWidget, ancestorStack []css.ElementInfo) {
	for p.pos < len(p.src) {
		p.skipWhitespace()
		if p.pos >= len(p.src) {
			break
		}
		if p.src[p.pos] == '<' {
			if p.pos+1 < len(p.src) && p.src[p.pos+1] == '/' {
				return // closing tag
			}
			// Skip comments
			if p.pos+3 < len(p.src) && p.src[p.pos:p.pos+4] == "<!--" {
				p.skipComment()
				continue
			}
			p.parseElement(parent, ancestorStack)
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

func (p *htmlParser) parseElement(parent containerWidget, ancestors []css.ElementInfo) {
	p.expect('<')
	tag := p.readTagName()
	attrs := p.readAttributes()
	selfClose := false
	if p.pos < len(p.src) && p.src[p.pos] == '/' {
		selfClose = true
		p.pos++
	}
	p.expect('>')

	id := attrs["id"]
	classes := strings.Fields(attrs["class"])
	inlineStyle := attrs["style"]

	// Self-closing / void elements
	switch tag {
	case "br":
		return
	case "img":
		img := widget.NewImage(p.tree, 0, p.cfg)
		if src, ok := attrs["src"]; ok {
			img.SetSrc(src)
		}
		applyInlineStyle(img, inlineStyle)
		p.registerWidget(img, tag, id, classes, inlineStyle, ancestors)
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
		p.registerWidget(inp, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(inp)
		return
	}

	if selfClose {
		return
	}

	// Build ancestor info for this element (used by children)
	selfInfo := css.ElementInfo{
		Tag:     tag,
		ID:      id,
		Classes: classes,
	}

	switch tag {
	case "button":
		label := p.readTextContent(tag)
		btn := widget.NewButton(p.tree, label, p.cfg)
		if _, ok := attrs["disabled"]; ok {
			btn.SetDisabled(true)
		}
		if v, ok := attrs["variant"]; ok {
			switch v {
			case "secondary", "outline":
				btn.SetVariant(widget.ButtonOutline)
			case "dashed":
				btn.SetVariant(widget.ButtonDashed)
			case "text":
				btn.SetVariant(widget.ButtonText)
			case "link":
				btn.SetVariant(widget.ButtonLink)
			}
		}
		if v, ok := attrs["theme"]; ok {
			switch v {
			case "primary":
				btn.SetTheme(widget.ThemePrimary)
			case "danger":
				btn.SetTheme(widget.ThemeDanger)
			case "warning":
				btn.SetTheme(widget.ThemeWarning)
			case "success":
				btn.SetTheme(widget.ThemeSuccess)
			}
		}
		if _, ok := attrs["ghost"]; ok {
			btn.SetGhost(true)
		}
		p.registerWidget(btn, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(btn)

	case "p", "span":
		text := p.readTextContent(tag)
		t := widget.NewText(p.tree, text, p.cfg)
		applyTextStyle(t, inlineStyle)
		p.registerWidget(t, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(t)

	case "a":
		text := p.readTextContent(tag)
		href := attrs["href"]
		lnk := widget.NewLink(p.tree, text, href, p.cfg)
		p.registerWidget(lnk, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(lnk)

	case "h1", "h2", "h3", "h4", "h5", "h6":
		text := p.readTextContent(tag)
		t := widget.NewText(p.tree, text, p.cfg)
		level := int(tag[1] - '0')
		sizes := []float32{32, 28, 24, 20, 16, 14}
		if level >= 1 && level <= 6 {
			t.SetFontSize(sizes[level-1])
		}
		p.registerWidget(t, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(t)

	case "select":
		p.readTextContent(tag) // consume content
		sel := widget.NewSelect(p.tree, nil, p.cfg)
		p.registerWidget(sel, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(sel)

	case "textarea":
		text := p.readTextContent(tag)
		ta := widget.NewTextArea(p.tree, p.cfg)
		ta.SetValue(text)
		if ph, ok := attrs["placeholder"]; ok {
			ta.SetPlaceholder(ph)
		}
		if r, ok := attrs["rows"]; ok {
			if n, err := strconv.Atoi(r); err == nil {
				ta.SetRows(n)
			}
		}
		p.registerWidget(ta, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(ta)

	// --- Layout semantic tags ---
	case "layout":
		l := widget.NewLayout(p.tree, p.cfg)
		applyLayoutBgColor(l, inlineStyle)
		p.registerWidget(l, tag, id, classes, inlineStyle, ancestors)
		// Layout uses Div embedding, so we cast to *Div for parseChildren
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(l, childAncestors)
		parent.AppendChild(l)
		p.skipClosingTag(tag)

	case "header":
		h := widget.NewHeader(p.tree, p.cfg)
		if v, ok := attrs["height"]; ok {
			if n, err := parsePx(v); err == nil {
				h.SetHeight(n)
			}
		}
		applyHeaderBgColor(h, inlineStyle)
		p.registerWidget(h, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(h, childAncestors)
		parent.AppendChild(h)
		p.skipClosingTag(tag)

	case "footer":
		f := widget.NewFooter(p.tree, p.cfg)
		applyFooterBgColor(f, inlineStyle)
		p.registerWidget(f, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(f, childAncestors)
		parent.AppendChild(f)
		p.skipClosingTag(tag)

	case "aside":
		a := widget.NewAside(p.tree, p.cfg)
		if v, ok := attrs["width"]; ok {
			if n, err := parsePx(v); err == nil {
				a.SetWidth(n)
			}
		}
		applyAsideBgColor(a, inlineStyle)
		p.registerWidget(a, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(a, childAncestors)
		parent.AppendChild(a)
		p.skipClosingTag(tag)

	case "main":
		c := widget.NewContent(p.tree, p.cfg)
		applyContentBgColor(c, inlineStyle)
		p.registerWidget(c, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(c, childAncestors)
		parent.AppendChild(c)
		p.skipClosingTag(tag)

	case "space":
		s := widget.NewSpace(p.tree, p.cfg)
		if v, ok := attrs["gap"]; ok {
			if n, err := parsePx(v); err == nil {
				s.SetGap(n)
			}
		}
		p.registerWidget(s, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(s, childAncestors)
		parent.AppendChild(s)
		p.skipClosingTag(tag)

	case "row":
		r := widget.NewRow(p.tree, p.cfg)
		if v, ok := attrs["gutter"]; ok {
			if n, err := parsePx(v); err == nil {
				r.SetGutter(n)
			}
		}
		p.registerWidget(r, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(r, childAncestors)
		parent.AppendChild(r)
		p.skipClosingTag(tag)

	case "col":
		span := 1
		if v, ok := attrs["span"]; ok {
			if n, err := strconv.Atoi(v); err == nil {
				span = n
			}
		}
		c := widget.NewCol(p.tree, span, p.cfg)
		p.registerWidget(c, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(c, childAncestors)
		parent.AppendChild(c)
		p.skipClosingTag(tag)

	// --- Custom widget tags ---
	case "checkbox":
		label := p.readTextContent(tag)
		cb := widget.NewCheckbox(p.tree, label, p.cfg)
		if _, ok := attrs["checked"]; ok {
			cb.SetChecked(true)
		}
		if _, ok := attrs["disabled"]; ok {
			cb.SetDisabled(true)
		}
		p.registerWidget(cb, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(cb)

	case "switch":
		p.readTextContent(tag) // consume
		sw := widget.NewSwitch(p.tree, p.cfg)
		if _, ok := attrs["checked"]; ok {
			sw.SetValue(true)
		}
		if _, ok := attrs["disabled"]; ok {
			sw.SetDisabled(true)
		}
		p.registerWidget(sw, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(sw)

	case "radio":
		label := p.readTextContent(tag)
		r := widget.NewRadio(p.tree, label, p.cfg)
		if _, ok := attrs["checked"]; ok {
			r.SetChecked(true)
		}
		// Group radios by group attribute
		group := attrs["group"]
		if group == "" {
			group = "_default"
		}
		if p.radioGroups == nil {
			p.radioGroups = make(map[string]*widget.RadioGroup)
		}
		rg, ok := p.radioGroups[group]
		if !ok {
			rg = widget.NewRadioGroup()
			p.radioGroups[group] = rg
		}
		rg.Add(r)
		p.registerWidget(r, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(r)

	case "tag":
		label := p.readTextContent(tag)
		tg := widget.NewTag(p.tree, label, p.cfg)
		if v, ok := attrs["type"]; ok {
			switch v {
			case "success":
				tg.SetTheme(widget.TagThemeSuccess)
			case "warning":
				tg.SetTheme(widget.TagThemeWarning)
			case "error":
				tg.SetTheme(widget.TagThemeDanger)
			case "processing":
				tg.SetTheme(widget.TagThemePrimary)
			}
		}
		p.registerWidget(tg, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(tg)

	case "progress":
		p.readTextContent(tag) // consume
		pg := widget.NewProgress(p.tree, p.cfg)
		if v, ok := attrs["percent"]; ok {
			if n, err := parsePx(v); err == nil {
				pg.SetPercent(n)
			}
		}
		p.registerWidget(pg, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(pg)

	case "message":
		text := p.readTextContent(tag)
		msg := widget.NewMessage(p.tree, text, p.cfg)
		if v, ok := attrs["type"]; ok {
			switch v {
			case "success":
				msg.SetTheme(widget.MessageThemeSuccess)
			case "warning":
				msg.SetTheme(widget.MessageThemeWarning)
			case "error":
				msg.SetTheme(widget.MessageThemeError)
			}
		}
		p.registerWidget(msg, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(msg)

	case "empty":
		p.readTextContent(tag) // consume
		e := widget.NewEmpty(p.tree, p.cfg)
		p.registerWidget(e, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(e)

	case "loading":
		p.readTextContent(tag) // consume
		l := widget.NewLoading(p.tree, p.cfg)
		if v, ok := attrs["tip"]; ok {
			l.SetText(v)
		}
		p.registerWidget(l, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(l)

	case "tooltip":
		text := p.readTextContent(tag)
		// Tooltip attaches to the previous sibling
		children := parent.Children()
		if len(children) > 0 {
			anchorID := children[len(children)-1].ElementID()
			widget.NewTooltip(p.tree, text, anchorID, p.cfg)
		}

	default: // div, nav, section, article, and unknown tags
		d := widget.NewDiv(p.tree, p.cfg)
		applyDivStyle(d, inlineStyle)
		if len(classes) > 0 {
			p.tree.SetClasses(d.ElementID(), classes)
		}
		if id != "" {
			p.tree.SetProperty(d.ElementID(), "id", id)
		}
		p.registerWidget(d, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(d, childAncestors)
		parent.AppendChild(d)
		p.skipClosingTag(tag)
	}
}

// registerWidget stores widget info for CSS application and indexes it for querying.
func (p *htmlParser) registerWidget(w widget.Widget, tag, id string, classes []string, inlineStyle string, ancestors []css.ElementInfo) {
	// Index for Document queries
	if p.doc != nil {
		if id != "" {
			p.doc.ids[id] = w
		}
		for _, cls := range classes {
			p.doc.classes[cls] = append(p.doc.classes[cls], w)
		}
		p.doc.tags[tag] = append(p.doc.tags[tag], w)
	}

	// Index for CSS matching
	if p.sheet == nil {
		return
	}
	info := widgetStyleInfo{
		widget:    w,
		tag:       tag,
		id:        id,
		classes:   classes,
		inlineCSS: inlineStyle,
	}
	// Copy ancestors slice
	info.ancestors = make([]css.ElementInfo, len(ancestors))
	copy(info.ancestors, ancestors)
	p.widgetInfo = append(p.widgetInfo, info)
}

// applyCSS resolves CSS rules for each recorded widget and applies the computed style.
func (p *htmlParser) applyCSS() {
	for _, info := range p.widgetInfo {
		el := &css.ElementInfo{
			Tag:     info.tag,
			ID:      info.id,
			Classes: info.classes,
		}
		inline := css.ParseInlineDeclarations(info.inlineCSS)
		computed := css.ResolveStyle(p.sheet, el, info.ancestors, inline)

		// Apply layout style
		info.widget.SetStyle(computed.Layout)

		// Apply visual properties to specific widget types
		applyVisualProps(info.widget, &computed)
	}
}

// applyVisualProps applies non-layout CSS properties (color, background, font-size, etc.)
func applyVisualProps(w widget.Widget, cs *css.ComputedStyle) {
	if cs.Color != "" {
		if c, ok := css.ParseColor(cs.Color); ok {
			if t, isText := w.(*widget.Text); isText {
				t.SetColor(c)
			}
		}
	}
	if cs.BackgroundColor != "" {
		if c, ok := css.ParseColor(cs.BackgroundColor); ok {
			switch d := w.(type) {
			case *widget.Div:
				d.SetBgColor(c)
			case *widget.Layout:
				d.SetBgColor(c)
			case *widget.Header:
				d.SetBgColor(c)
			case *widget.Footer:
				d.SetBgColor(c)
			case *widget.Aside:
				d.SetBgColor(c)
			case *widget.Content:
				d.SetBgColor(c)
			}
		}
	}
	if cs.FontSize != "" {
		size := css.ParseFloat(cs.FontSize)
		if size > 0 {
			if t, isText := w.(*widget.Text); isText {
				t.SetFontSize(size)
			}
		}
	}
	if cs.BorderRadius != "" {
		r := css.ParseFloat(cs.BorderRadius)
		if d, isDiv := w.(*widget.Div); isDiv {
			d.SetBorderRadius(r)
		}
	}
	if cs.BorderColor != "" {
		if c, ok := css.ParseColor(cs.BorderColor); ok {
			if d, isDiv := w.(*widget.Div); isDiv {
				d.SetBorderColor(c)
			}
		}
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

func (p *htmlParser) skipComment() {
	p.pos += 4 // skip <!--
	idx := strings.Index(p.src[p.pos:], "-->")
	if idx < 0 {
		p.pos = len(p.src)
	} else {
		p.pos += idx + 3
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

// Background color helpers for layout widgets (parse from inline style).
func applyLayoutBgColor(l *widget.Layout, style string) {
	if style == "" {
		return
	}
	props := parseInlineCSS(style)
	if v, ok := props["background-color"]; ok {
		l.SetBgColor(parseColor(v))
	}
	if v, ok := props["background"]; ok {
		l.SetBgColor(parseColor(v))
	}
}

func applyHeaderBgColor(h *widget.Header, style string) {
	if style == "" {
		return
	}
	props := parseInlineCSS(style)
	if v, ok := props["background-color"]; ok {
		h.SetBgColor(parseColor(v))
	}
	if v, ok := props["background"]; ok {
		h.SetBgColor(parseColor(v))
	}
}

func applyFooterBgColor(f *widget.Footer, style string) {
	if style == "" {
		return
	}
	props := parseInlineCSS(style)
	if v, ok := props["background-color"]; ok {
		f.SetBgColor(parseColor(v))
	}
	if v, ok := props["background"]; ok {
		f.SetBgColor(parseColor(v))
	}
}

func applyAsideBgColor(a *widget.Aside, style string) {
	if style == "" {
		return
	}
	props := parseInlineCSS(style)
	if v, ok := props["background-color"]; ok {
		a.SetBgColor(parseColor(v))
	}
	if v, ok := props["background"]; ok {
		a.SetBgColor(parseColor(v))
	}
}

func applyContentBgColor(c *widget.Content, style string) {
	if style == "" {
		return
	}
	props := parseInlineCSS(style)
	if v, ok := props["background-color"]; ok {
		c.SetBgColor(parseColor(v))
	}
	if v, ok := props["background"]; ok {
		c.SetBgColor(parseColor(v))
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

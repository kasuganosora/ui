package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/css"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/theme"
	"github.com/kasuganosora/ui/widget"
)

// Document represents a parsed HTML document with query capabilities.
type Document struct {
	Root    *widget.Div
	ids     map[string]widget.Widget
	classes map[string][]widget.Widget
	tags    map[string][]widget.Widget

	// cleanups tracks unsubscribe functions for reactive bindings.
	cleanups []func()

	// data holds template data for interpolation.
	data map[string]interface{}

	// bindings tracks data-bound widgets for reactive updates.
	bindings []*binding

	// sheet is the parsed CSS stylesheet (retained for theme variable injection).
	sheet *css.Stylesheet
}

// binding describes a data-bound relationship between a key and a widget.
type binding struct {
	key    string
	widget widget.Widget
	kind   string // "text", "if", "model"
	tmpl   string // original template string for text bindings
	update func(interface{})
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

// On registers an event handler on all elements matching the CSS selector.
// Supported selectors: "#id", ".class", "tag".
func (d *Document) On(selector string, eventType string, handler func(widget.Widget)) {
	widgets := d.QueryAll(selector)
	for _, w := range widgets {
		w := w // capture for closure
		switch eventType {
		case "click":
			if btn, ok := w.(*widget.Button); ok {
				btn.OnClick(func() { handler(w) })
			}
		case "change":
			switch v := w.(type) {
			case *widget.Input:
				v.OnChange(func(string) { handler(w) })
			case *widget.TextArea:
				v.OnChange(func(string) { handler(w) })
			case *widget.Select:
				v.OnChange(func(string) { handler(w) })
			}
		case "toggle":
			switch v := w.(type) {
			case *widget.Checkbox:
				v.OnChange(func(bool) { handler(w) })
			case *widget.Switch:
				v.OnChange(func(bool) { handler(w) })
			}
		}
	}
}

// QueryAll finds widgets matching a simple CSS selector.
// Supported: "#id", ".class", "tag".
func (d *Document) QueryAll(selector string) []widget.Widget {
	selector = strings.TrimSpace(selector)
	if selector == "" {
		return nil
	}
	if strings.HasPrefix(selector, "#") {
		if w := d.ids[selector[1:]]; w != nil {
			return []widget.Widget{w}
		}
		return nil
	}
	if strings.HasPrefix(selector, ".") {
		return d.classes[selector[1:]]
	}
	return d.tags[selector]
}

// Dispose cleans up all event handlers and reactive bindings.
func (d *Document) Dispose() {
	for _, fn := range d.cleanups {
		fn()
	}
	d.cleanups = nil
	d.bindings = nil
	if d.Root != nil {
		d.Root.Destroy()
		d.Root = nil
	}
}

// addCleanup registers a cleanup function.
func (d *Document) addCleanup(fn func()) {
	d.cleanups = append(d.cleanups, fn)
}

// SetData sets template data and triggers re-render of bound elements.
func (d *Document) SetData(key string, value interface{}) {
	if d.data == nil {
		d.data = make(map[string]interface{})
	}
	d.data[key] = value
	for _, b := range d.bindings {
		if b.key == key {
			b.update(value)
		}
	}
}

// GetData retrieves template data.
func (d *Document) GetData(key string) interface{} {
	if d.data == nil {
		return nil
	}
	return d.data[key]
}

// SetTheme injects theme CSS variables into the document's stylesheet.
// This allows CSS rules using var(--ui-*) to pick up the new theme values.
func (d *Document) SetTheme(t *theme.Theme) {
	if d.sheet == nil {
		d.sheet = &css.Stylesheet{Variables: make(map[string]string)}
	}
	if d.sheet.Variables == nil {
		d.sheet.Variables = make(map[string]string)
	}
	for k, v := range t.ToCSSVariables() {
		d.sheet.Variables[k] = v
	}
}

// interpolate replaces {{key}} patterns in a template with data values.
func (d *Document) interpolate(tmpl string) string {
	result := tmpl
	for {
		start := strings.Index(result, "{{")
		if start < 0 {
			break
		}
		end := strings.Index(result[start:], "}}")
		if end < 0 {
			break
		}
		end += start
		key := strings.TrimSpace(result[start+2 : end])
		var replacement string
		if d.data != nil {
			if v, ok := d.data[key]; ok {
				replacement = fmt.Sprint(v)
			}
		}
		result = result[:start] + replacement + result[end+2:]
	}
	return result
}

// isTruthy returns whether a data value is considered truthy.
func isTruthy(v interface{}) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case int:
		return val != 0
	case float64:
		return val != 0
	case string:
		return val != ""
	default:
		return true
	}
}

// LoadHTML parses a simple HTML string and builds a widget tree.
// Supported tags: div, span, button, input, p, h1-h6, img, br, a, select, textarea,
// header, footer, aside, main, nav, section, article, space, row, col, layout,
// checkbox, switch, radio, tag, progress, message, empty, loading, tooltip,
// divider, icon, badge, avatar, alert, statistic, rate, skeleton, watermark,
// slider, inputnumber, card, collapse, tabs, dialog, drawer, panel, splitter,
// form, list, table, menu, portal, breadcrumb, pagination, steps, timeline,
// anchor, backtop, colorpicker, datepicker, daterangepicker, cascader,
// treeselect, transfer, autocomplete, taginput, upload, notification,
// popover, popconfirm, contextmenu, calendar, swiper, subwindow, imageviewer,
// richtext, tree/treew.
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
	p.doc.sheet = p.sheet
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
				if strings.Contains(text, "{{") {
					// Template text — create binding
					rendered := p.doc.interpolate(text)
					t := widget.NewText(p.tree, rendered, p.cfg)
					parent.AppendChild(t)
					p.addTextBinding(t, text)
				} else {
					t := widget.NewText(p.tree, text, p.cfg)
					parent.AppendChild(t)
				}
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
		img := widget.NewImg(p.tree, p.cfg)
		if src, ok := attrs["src"]; ok {
			img.SetSrc(src)
		}
		if alt, ok := attrs["alt"]; ok {
			img.SetAlt(alt)
		}
		if w, ok := attrs["width"]; ok {
			if v, err := strconv.ParseFloat(w, 32); err == nil {
				s := img.Style()
				s.Width = layout.Px(float32(v))
				img.SetStyle(s)
			}
		}
		if h, ok := attrs["height"]; ok {
			if v, err := strconv.ParseFloat(h, 32); err == nil {
				s := img.Style()
				s.Height = layout.Px(float32(v))
				img.SetStyle(s)
			}
		}
		applyInlineStyle(img, inlineStyle)
		p.registerWidgetWithAttrs(img, tag, id, classes, inlineStyle, ancestors, attrs)
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
		p.registerWidgetWithAttrs(inp, tag, id, classes, inlineStyle, ancestors, attrs)
		parent.AppendChild(inp)
		return
	case "icon":
		name := attrs["name"]
		ic := widget.NewIcon(p.tree, name, p.cfg)
		p.registerWidget(ic, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(ic)
		return
	case "colorpicker":
		w := widget.NewColorPicker(p.tree, p.cfg)
		p.registerWidget(w, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(w)
		return
	case "datepicker":
		w := widget.NewDatePicker(p.tree, p.cfg)
		p.registerWidget(w, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(w)
		return
	case "daterangepicker":
		w := widget.NewDateRangePicker(p.tree, p.cfg)
		p.registerWidget(w, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(w)
		return
	case "cascader":
		w := widget.NewCascader(p.tree, p.cfg)
		p.registerWidget(w, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(w)
		return
	case "treeselect":
		w := widget.NewTreeSelect(p.tree, p.cfg)
		p.registerWidget(w, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(w)
		return
	case "transfer":
		w := widget.NewTransfer(p.tree, p.cfg)
		p.registerWidget(w, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(w)
		return
	case "taginput":
		w := widget.NewTagInput(p.tree, p.cfg)
		p.registerWidget(w, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(w)
		return
	case "upload":
		w := widget.NewUpload(p.tree, p.cfg)
		p.registerWidget(w, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(w)
		return
	case "backtop":
		w := widget.NewBackTop(p.tree, p.cfg)
		p.registerWidget(w, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(w)
		return
	case "calendar":
		w := widget.NewCalendar(p.tree, p.cfg)
		p.registerWidget(w, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(w)
		return
	case "imageviewer":
		w := widget.NewImageViewer(p.tree, p.cfg)
		p.registerWidget(w, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(w)
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
		if v, ok := attrs["shape"]; ok {
			switch v {
			case "square":
				btn.SetShape(widget.ShapeSquare)
			case "round":
				btn.SetShape(widget.ShapeRound)
			case "circle":
				btn.SetShape(widget.ShapeCircle)
			}
		}
		p.registerWidgetWithAttrs(btn, tag, id, classes, inlineStyle, ancestors, attrs)
		parent.AppendChild(btn)

	case "p", "span":
		text := p.readTextContent(tag)
		rendered := text
		if strings.Contains(text, "{{") {
			rendered = p.doc.interpolate(text)
		}
		t := widget.NewText(p.tree, rendered, p.cfg)
		applyTextStyle(t, inlineStyle)
		p.registerWidgetWithAttrs(t, tag, id, classes, inlineStyle, ancestors, attrs)
		parent.AppendChild(t)
		if strings.Contains(text, "{{") {
			p.addTextBinding(t, text)
		}

	case "a":
		text := p.readTextContent(tag)
		href := attrs["href"]
		lnk := widget.NewLink(p.tree, text, href, p.cfg)
		p.registerWidgetWithAttrs(lnk, tag, id, classes, inlineStyle, ancestors, attrs)
		parent.AppendChild(lnk)

	case "h1", "h2", "h3", "h4", "h5", "h6":
		text := p.readTextContent(tag)
		rendered := text
		if strings.Contains(text, "{{") {
			rendered = p.doc.interpolate(text)
		}
		t := widget.NewText(p.tree, rendered, p.cfg)
		level := int(tag[1] - '0')
		sizes := []float32{32, 28, 24, 20, 16, 14}
		if level >= 1 && level <= 6 {
			t.SetFontSize(sizes[level-1])
		}
		p.registerWidgetWithAttrs(t, tag, id, classes, inlineStyle, ancestors, attrs)
		parent.AppendChild(t)
		if strings.Contains(text, "{{") {
			p.addTextBinding(t, text)
		}

	case "select":
		p.readTextContent(tag) // consume content
		sel := widget.NewSelect(p.tree, nil, p.cfg)
		p.registerWidgetWithAttrs(sel, tag, id, classes, inlineStyle, ancestors, attrs)
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
		p.registerWidgetWithAttrs(ta, tag, id, classes, inlineStyle, ancestors, attrs)
		parent.AppendChild(ta)

	// --- Layout semantic tags ---
	case "layout":
		l := widget.NewLayout(p.tree, p.cfg)
		applyLayoutBgColor(l, inlineStyle)
		p.registerWidgetWithAttrs(l, tag, id, classes, inlineStyle, ancestors, attrs)
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
		p.registerWidgetWithAttrs(h, tag, id, classes, inlineStyle, ancestors, attrs)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(h, childAncestors)
		parent.AppendChild(h)
		p.skipClosingTag(tag)

	case "footer":
		f := widget.NewFooter(p.tree, p.cfg)
		applyFooterBgColor(f, inlineStyle)
		p.registerWidgetWithAttrs(f, tag, id, classes, inlineStyle, ancestors, attrs)
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
		p.registerWidgetWithAttrs(a, tag, id, classes, inlineStyle, ancestors, attrs)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(a, childAncestors)
		parent.AppendChild(a)
		p.skipClosingTag(tag)

	case "main":
		c := widget.NewContent(p.tree, p.cfg)
		applyContentBgColor(c, inlineStyle)
		p.registerWidgetWithAttrs(c, tag, id, classes, inlineStyle, ancestors, attrs)
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
		p.registerWidgetWithAttrs(s, tag, id, classes, inlineStyle, ancestors, attrs)
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
		p.registerWidgetWithAttrs(r, tag, id, classes, inlineStyle, ancestors, attrs)
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
		p.registerWidgetWithAttrs(c, tag, id, classes, inlineStyle, ancestors, attrs)
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
		p.registerWidgetWithAttrs(cb, tag, id, classes, inlineStyle, ancestors, attrs)
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
		p.registerWidgetWithAttrs(sw, tag, id, classes, inlineStyle, ancestors, attrs)
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
		p.registerWidgetWithAttrs(r, tag, id, classes, inlineStyle, ancestors, attrs)
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
		p.registerWidgetWithAttrs(tg, tag, id, classes, inlineStyle, ancestors, attrs)
		parent.AppendChild(tg)

	case "progress":
		p.readTextContent(tag) // consume
		pg := widget.NewProgress(p.tree, p.cfg)
		if v, ok := attrs["percent"]; ok {
			if n, err := parsePx(v); err == nil {
				pg.SetPercent(n)
			}
		}
		p.registerWidgetWithAttrs(pg, tag, id, classes, inlineStyle, ancestors, attrs)
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
		p.registerWidgetWithAttrs(msg, tag, id, classes, inlineStyle, ancestors, attrs)
		parent.AppendChild(msg)

	case "empty":
		p.readTextContent(tag) // consume
		e := widget.NewEmpty(p.tree, p.cfg)
		p.registerWidgetWithAttrs(e, tag, id, classes, inlineStyle, ancestors, attrs)
		parent.AppendChild(e)

	case "loading":
		p.readTextContent(tag) // consume
		l := widget.NewLoading(p.tree, p.cfg)
		if v, ok := attrs["tip"]; ok {
			l.SetText(v)
		}
		p.registerWidgetWithAttrs(l, tag, id, classes, inlineStyle, ancestors, attrs)
		parent.AppendChild(l)

	case "tooltip":
		text := p.readTextContent(tag)
		// Tooltip attaches to the previous sibling
		children := parent.Children()
		if len(children) > 0 {
			anchorID := children[len(children)-1].ElementID()
			widget.NewTooltip(p.tree, text, anchorID, p.cfg)
		}

	// --- Simple widgets (text content) ---
	case "divider":
		p.readTextContent(tag)
		dv := widget.NewDivider(p.tree, p.cfg)
		if v, ok := attrs["layout"]; ok && v == "vertical" {
			dv.SetLayout(widget.DividerVertical)
		}
		if _, ok := attrs["dashed"]; ok {
			dv.SetDashed(true)
		}
		if v, ok := attrs["content"]; ok {
			dv.SetContent(v)
		}
		p.registerWidget(dv, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(dv)

	case "badge":
		p.readTextContent(tag)
		bg := widget.NewBadge(p.tree, p.cfg)
		if v, ok := attrs["count"]; ok {
			if n, err := strconv.Atoi(v); err == nil {
				bg.SetCount(n)
			}
		}
		if _, ok := attrs["dot"]; ok {
			bg.SetDot(true)
		}
		if v, ok := attrs["max-count"]; ok {
			if n, err := strconv.Atoi(v); err == nil {
				bg.SetMaxCount(n)
			}
		}
		p.registerWidget(bg, tag, id, classes, inlineStyle, ancestors)
		children := parent.Children()
		if len(children) > 0 {
			bg.AppendChild(children[len(children)-1])
		}
		parent.AppendChild(bg)

	case "avatar":
		p.readTextContent(tag)
		av := widget.NewAvatar(p.tree, p.cfg)
		if v, ok := attrs["shape"]; ok && v == "square" {
			av.SetShape(widget.AvatarSquare)
		}
		if v, ok := attrs["size"]; ok {
			switch v {
			case "small":
				av.SetSize(widget.SizeSmall)
			case "large":
				av.SetSize(widget.SizeLarge)
			}
		}
		if v, ok := attrs["content"]; ok {
			av.SetContent(v)
		}
		if v, ok := attrs["image"]; ok {
			av.SetImage(v)
		}
		p.registerWidget(av, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(av)

	case "alert":
		msg := p.readTextContent(tag)
		al := widget.NewAlert(p.tree, msg, p.cfg)
		if v, ok := attrs["theme"]; ok {
			switch v {
			case "success":
				al.SetTheme(widget.AlertThemeSuccess)
			case "warning":
				al.SetTheme(widget.AlertThemeWarning)
			case "error":
				al.SetTheme(widget.AlertThemeError)
			case "info":
				al.SetTheme(widget.AlertThemeInfo)
			}
		}
		if v, ok := attrs["title"]; ok {
			al.SetTitle(v)
		}
		if _, ok := attrs["closable"]; ok {
			al.SetCloseBtn(true)
		}
		p.registerWidget(al, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(al)

	case "statistic":
		p.readTextContent(tag)
		title := attrs["title"]
		value := attrs["value"]
		st := widget.NewStatistic(p.tree, title, value, p.cfg)
		if v, ok := attrs["prefix"]; ok {
			st.SetPrefix(v)
		}
		if v, ok := attrs["suffix"]; ok {
			st.SetSuffix(v)
		}
		if v, ok := attrs["trend"]; ok {
			switch v {
			case "up":
				st.SetTrend(widget.TrendIncrease)
			case "down":
				st.SetTrend(widget.TrendDecrease)
			}
		}
		p.registerWidget(st, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(st)

	case "rate":
		p.readTextContent(tag)
		rt := widget.NewRate(p.tree, p.cfg)
		if v, ok := attrs["value"]; ok {
			if n, err := parsePx(v); err == nil {
				rt.SetValue(n)
			}
		}
		if v, ok := attrs["count"]; ok {
			if n, err := strconv.Atoi(v); err == nil {
				rt.SetCount(n)
			}
		}
		if _, ok := attrs["disabled"]; ok {
			rt.SetDisabled(true)
		}
		p.registerWidget(rt, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(rt)

	case "skeleton":
		p.readTextContent(tag)
		sk := widget.NewSkeleton(p.tree, p.cfg)
		if v, ok := attrs["rows"]; ok {
			if n, err := strconv.Atoi(v); err == nil {
				sk.SetRows(n)
			}
		}
		if _, ok := attrs["avatar"]; ok {
			sk.SetAvatar(true)
		}
		if _, ok := attrs["loading"]; ok {
			sk.SetLoading(true)
		}
		p.registerWidget(sk, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(sk)

	case "watermark":
		p.readTextContent(tag)
		wmText := attrs["text"]
		wm := widget.NewWatermark(p.tree, wmText, p.cfg)
		p.registerWidget(wm, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(wm)

	case "slider":
		p.readTextContent(tag)
		sl := widget.NewSlider(p.tree, p.cfg)
		if v, ok := attrs["min"]; ok {
			if n, err := parsePx(v); err == nil {
				sl.SetMin(n)
			}
		}
		if v, ok := attrs["max"]; ok {
			if n, err := parsePx(v); err == nil {
				sl.SetMax(n)
			}
		}
		if v, ok := attrs["value"]; ok {
			if n, err := parsePx(v); err == nil {
				sl.SetValue(n)
			}
		}
		if v, ok := attrs["step"]; ok {
			if n, err := parsePx(v); err == nil {
				sl.SetStep(n)
			}
		}
		if _, ok := attrs["disabled"]; ok {
			sl.SetDisabled(true)
		}
		p.registerWidget(sl, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(sl)

	case "inputnumber":
		p.readTextContent(tag)
		in := widget.NewInputNumber(p.tree, p.cfg)
		if v, ok := attrs["value"]; ok {
			if n, err := strconv.ParseFloat(v, 64); err == nil {
				in.SetValue(n)
			}
		}
		if v, ok := attrs["min"]; ok {
			if n, err := strconv.ParseFloat(v, 64); err == nil {
				in.SetMin(n)
			}
		}
		if v, ok := attrs["max"]; ok {
			if n, err := strconv.ParseFloat(v, 64); err == nil {
				in.SetMax(n)
			}
		}
		if v, ok := attrs["step"]; ok {
			if n, err := strconv.ParseFloat(v, 64); err == nil {
				in.SetStep(n)
			}
		}
		if _, ok := attrs["disabled"]; ok {
			in.SetDisabled(true)
		}
		if v, ok := attrs["placeholder"]; ok {
			in.SetPlaceholder(v)
		}
		p.registerWidget(in, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(in)

	case "autocomplete":
		p.readTextContent(tag)
		ac := widget.NewAutoComplete(p.tree, p.cfg)
		if v, ok := attrs["placeholder"]; ok {
			ac.SetText(v)
		}
		p.registerWidget(ac, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(ac)

	case "notification":
		p.readTextContent(tag)
		ntTitle := attrs["title"]
		ntContent := attrs["content"]
		nt := widget.NewNotification(p.tree, ntTitle, ntContent, p.cfg)
		p.registerWidget(nt, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(nt)

	case "pagination":
		p.readTextContent(tag)
		pg := widget.NewPagination(p.tree, p.cfg)
		if v, ok := attrs["total"]; ok {
			if n, err := strconv.Atoi(v); err == nil {
				pg.SetTotal(n)
			}
		}
		if v, ok := attrs["page-size"]; ok {
			if n, err := strconv.Atoi(v); err == nil {
				pg.SetPageSize(n)
			}
		}
		if v, ok := attrs["current"]; ok {
			if n, err := strconv.Atoi(v); err == nil {
				pg.SetCurrent(n)
			}
		}
		p.registerWidget(pg, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(pg)

	// --- Container widgets ---
	case "card":
		cd := widget.NewCard(p.tree, p.cfg)
		if v, ok := attrs["title"]; ok {
			cd.SetTitle(v)
		}
		if _, ok := attrs["bordered"]; ok {
			cd.SetBordered(true)
		}
		if _, ok := attrs["shadow"]; ok {
			cd.SetShadow(true)
		}
		p.registerWidget(cd, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(cd, childAncestors)
		parent.AppendChild(cd)
		p.skipClosingTag(tag)

	case "collapse":
		cl := widget.NewCollapse(p.tree, p.cfg)
		if _, ok := attrs["accordion"]; ok {
			cl.SetAccordion(true)
		}
		if _, ok := attrs["expand-mutex"]; ok {
			cl.SetExpandMutex(true)
		}
		if _, ok := attrs["borderless"]; ok {
			cl.SetBorderless(true)
		}
		p.registerWidget(cl, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(cl, childAncestors)
		parent.AppendChild(cl)
		p.skipClosingTag(tag)

	case "tabs":
		tb := widget.NewTabs(p.tree, nil, p.cfg)
		p.registerWidget(tb, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(tb, childAncestors)
		parent.AppendChild(tb)
		p.skipClosingTag(tag)

	case "dialog":
		dlTitle := attrs["title"]
		dl := widget.NewDialog(p.tree, dlTitle, p.cfg)
		if v, ok := attrs["width"]; ok {
			if n, err := parsePx(v); err == nil {
				dl.SetWidth(n)
			}
		}
		if _, ok := attrs["visible"]; ok {
			dl.Open()
		}
		p.registerWidget(dl, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(dl, childAncestors)
		parent.AppendChild(dl)
		p.skipClosingTag(tag)

	case "drawer":
		drTitle := attrs["title"]
		dr := widget.NewDrawer(p.tree, drTitle, p.cfg)
		if v, ok := attrs["placement"]; ok {
			switch v {
			case "left":
				dr.SetPlacement(widget.DrawerLeft)
			case "right":
				dr.SetPlacement(widget.DrawerRight)
			case "top":
				dr.SetPlacement(widget.DrawerTop)
			case "bottom":
				dr.SetPlacement(widget.DrawerBottom)
			}
		}
		if _, ok := attrs["visible"]; ok {
			dr.Open()
		}
		p.registerWidget(dr, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(dr, childAncestors)
		parent.AppendChild(dr)
		p.skipClosingTag(tag)

	case "panel":
		pnTitle := attrs["title"]
		pn := widget.NewPanel(p.tree, pnTitle, p.cfg)
		p.registerWidget(pn, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(pn, childAncestors)
		parent.AppendChild(pn)
		p.skipClosingTag(tag)

	case "splitter":
		sp := widget.NewSplitter(p.tree, p.cfg)
		p.registerWidget(sp, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(sp, childAncestors)
		parent.AppendChild(sp)
		p.skipClosingTag(tag)

	case "form":
		fm := widget.NewForm(p.tree, p.cfg)
		p.registerWidget(fm, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(fm, childAncestors)
		parent.AppendChild(fm)
		p.skipClosingTag(tag)

	case "list":
		ls := widget.NewList(p.tree, p.cfg)
		if _, ok := attrs["selectable"]; ok {
			ls.SetSelectable(true)
			ls.SetBordered(false)
			ls.SetSplit(false)
		}
		if v, ok := attrs["value"]; ok {
			ls.SetSelectedValue(v)
		}
		if v, ok := attrs["item-height"]; ok {
			if h, err := parsePx(v); err == nil && h > 0 {
				ls.SetItemHeight(h)
			}
		}
		// Parse <list-item> children inline
		items := p.parseListItems(tag)
		if len(items) > 0 {
			ls.SetItems(items)
		}
		p.registerWidgetWithAttrs(ls, tag, id, classes, inlineStyle, ancestors, attrs)
		parent.AppendChild(ls)

	case "table":
		tbl := widget.NewTable(p.tree, nil, p.cfg)
		p.registerWidget(tbl, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(tbl, childAncestors)
		parent.AppendChild(tbl)
		p.skipClosingTag(tag)

	case "menu":
		mn := widget.NewMenu(p.tree, p.cfg)
		p.registerWidget(mn, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(mn, childAncestors)
		parent.AppendChild(mn)
		p.skipClosingTag(tag)

	case "portal":
		pt := widget.NewPortal(p.tree, p.cfg)
		p.registerWidget(pt, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(pt, childAncestors)
		parent.AppendChild(pt)
		p.skipClosingTag(tag)

	case "contextmenu":
		cm := widget.NewContextMenu(p.tree, p.cfg)
		p.registerWidget(cm, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(cm, childAncestors)
		parent.AppendChild(cm)
		p.skipClosingTag(tag)

	case "swiper":
		sw := widget.NewSwiper(p.tree, p.cfg)
		p.registerWidget(sw, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(sw, childAncestors)
		parent.AppendChild(sw)
		p.skipClosingTag(tag)

	case "richtext":
		rtw := widget.NewRichText(p.tree, p.cfg)
		p.registerWidget(rtw, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(rtw, childAncestors)
		parent.AppendChild(rtw)
		p.skipClosingTag(tag)

	case "treew", "tree":
		tw := widget.NewTree(p.tree, p.cfg)
		p.registerWidget(tw, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(tw, childAncestors)
		parent.AppendChild(tw)
		p.skipClosingTag(tag)

	// --- Navigation widgets ---
	case "breadcrumb":
		bc := widget.NewBreadcrumb(p.tree, p.cfg)
		p.registerWidget(bc, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(bc, childAncestors)
		parent.AppendChild(bc)
		p.skipClosingTag(tag)

	case "steps":
		stw := widget.NewSteps(p.tree, p.cfg)
		if v, ok := attrs["current"]; ok {
			if n, err := strconv.Atoi(v); err == nil {
				stw.SetCurrent(n)
			}
		}
		p.registerWidget(stw, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(stw, childAncestors)
		parent.AppendChild(stw)
		p.skipClosingTag(tag)

	case "timeline":
		tl := widget.NewTimeline(p.tree, p.cfg)
		p.registerWidget(tl, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(tl, childAncestors)
		parent.AppendChild(tl)
		p.skipClosingTag(tag)

	case "anchor":
		an := widget.NewAnchor(p.tree, p.cfg)
		p.registerWidget(an, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(an, childAncestors)
		parent.AppendChild(an)
		p.skipClosingTag(tag)

	case "subwindow":
		swTitle := attrs["title"]
		sww := widget.NewSubWindow(p.tree, swTitle, p.cfg)
		p.registerWidget(sww, tag, id, classes, inlineStyle, ancestors)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(sww, childAncestors)
		parent.AppendChild(sww)
		p.skipClosingTag(tag)

	// --- Popover-like widgets ---
	case "popover":
		p.readTextContent(tag)
		pv := widget.NewPopover(p.tree, p.cfg)
		if v, ok := attrs["title"]; ok {
			pv.SetTitle(v)
		}
		p.registerWidget(pv, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(pv)

	case "popconfirm":
		p.readTextContent(tag)
		pcTitle := attrs["title"]
		pc := widget.NewPopconfirm(p.tree, pcTitle, p.cfg)
		p.registerWidget(pc, tag, id, classes, inlineStyle, ancestors)
		parent.AppendChild(pc)

	default: // div, nav, section, article, and unknown tags
		d := widget.NewDiv(p.tree, p.cfg)
		applyDivStyle(d, inlineStyle)
		if len(classes) > 0 {
			p.tree.SetClasses(d.ElementID(), classes)
		}
		if id != "" {
			p.tree.SetProperty(d.ElementID(), "id", id)
		}
		p.registerWidgetWithAttrs(d, tag, id, classes, inlineStyle, ancestors, attrs)
		childAncestors := append([]css.ElementInfo{selfInfo}, ancestors...)
		p.parseChildren(d, childAncestors)
		parent.AppendChild(d)
		p.skipClosingTag(tag)
	}
}

// registerWidget stores widget info for CSS application and indexes it for querying.
func (p *htmlParser) registerWidget(w widget.Widget, tag, id string, classes []string, inlineStyle string, ancestors []css.ElementInfo) {
	p.registerWidgetWithAttrs(w, tag, id, classes, inlineStyle, ancestors, nil)
}

// registerWidgetWithAttrs stores widget info and processes data-binding attributes.
func (p *htmlParser) registerWidgetWithAttrs(w widget.Widget, tag, id string, classes []string, inlineStyle string, ancestors []css.ElementInfo, attrs map[string]string) {
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

	// Process data-binding attributes
	if attrs != nil {
		p.applyDataAttributes(w, attrs)
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

// addTextBinding creates a text binding for template interpolation.
func (p *htmlParser) addTextBinding(t *widget.Text, tmpl string) {
	keys := extractTemplateKeys(tmpl)
	for _, key := range keys {
		key := key // capture
		b := &binding{
			key:    key,
			widget: t,
			kind:   "text",
			tmpl:   tmpl,
			update: func(interface{}) {
				t.SetText(p.doc.interpolate(tmpl))
			},
		}
		p.doc.bindings = append(p.doc.bindings, b)
	}
}

// applyDataAttributes handles data-if, data-model, and data-for attributes.
func (p *htmlParser) applyDataAttributes(w widget.Widget, attrs map[string]string) {
	// data-if: show/hide based on truthiness
	if key, ok := attrs["data-if"]; ok {
		if p.doc.data != nil {
			if !isTruthy(p.doc.data[key]) {
				s := w.Style()
				s.Display = layout.DisplayNone
				w.SetStyle(s)
			}
		} else {
			s := w.Style()
			s.Display = layout.DisplayNone
			w.SetStyle(s)
		}
		b := &binding{
			key:    key,
			widget: w,
			kind:   "if",
			update: func(v interface{}) {
				s := w.Style()
				if isTruthy(v) {
					s.Display = layout.DisplayBlock
				} else {
					s.Display = layout.DisplayNone
				}
				w.SetStyle(s)
			},
		}
		p.doc.bindings = append(p.doc.bindings, b)
	}

	// data-model: two-way binding for input widgets
	if key, ok := attrs["data-model"]; ok {
		if p.doc.data != nil {
			if v, exists := p.doc.data[key]; exists {
				setWidgetValue(w, fmt.Sprint(v))
			}
		}
		b := &binding{
			key:    key,
			widget: w,
			kind:   "model",
			update: func(v interface{}) {
				setWidgetValue(w, fmt.Sprint(v))
			},
		}
		p.doc.bindings = append(p.doc.bindings, b)

		// Wire up change handler for two-way binding
		doc := p.doc
		switch inp := w.(type) {
		case *widget.Input:
			inp.OnChange(func(val string) {
				if doc.data == nil {
					doc.data = make(map[string]interface{})
				}
				doc.data[key] = val
			})
		case *widget.TextArea:
			inp.OnChange(func(val string) {
				if doc.data == nil {
					doc.data = make(map[string]interface{})
				}
				doc.data[key] = val
			})
		case *widget.Select:
			inp.OnChange(func(val string) {
				if doc.data == nil {
					doc.data = make(map[string]interface{})
				}
				doc.data[key] = val
			})
		}
	}
}

// setWidgetValue sets the value of an input-like widget.
func setWidgetValue(w widget.Widget, val string) {
	switch v := w.(type) {
	case *widget.Input:
		v.SetValue(val)
	case *widget.TextArea:
		v.SetValue(val)
	case *widget.Text:
		v.SetText(val)
	}
}

// extractTemplateKeys returns all unique keys from {{key}} patterns.
func extractTemplateKeys(tmpl string) []string {
	var keys []string
	seen := make(map[string]bool)
	s := tmpl
	for {
		start := strings.Index(s, "{{")
		if start < 0 {
			break
		}
		end := strings.Index(s[start:], "}}")
		if end < 0 {
			break
		}
		end += start
		key := strings.TrimSpace(s[start+2 : end])
		if key != "" && !seen[key] {
			keys = append(keys, key)
			seen[key] = true
		}
		s = s[end+2:]
	}
	return keys
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

		// Apply layout style.
		// Form controls (input, textarea, button) manage their own height and padding
		// via their constructors/SetRows. Only override those dimensions when CSS
		// explicitly specifies them; otherwise preserve the widget's own values.
		newStyle := computed.Layout
		switch info.widget.(type) {
		case *widget.Input, *widget.TextArea, *widget.Button:
			existing := info.widget.Style()
			if _, hasH := computed.Raw["height"]; !hasH {
				newStyle.Height = existing.Height
			}
			hasPadding := false
			for _, pp := range []string{"padding", "padding-top", "padding-bottom", "padding-left", "padding-right"} {
				if _, ok := computed.Raw[pp]; ok {
					hasPadding = true
					break
				}
			}
			if !hasPadding {
				newStyle.Padding = existing.Padding
			}
		}
		info.widget.SetStyle(newStyle)

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
			case *widget.Card:
				d.SetBgColor(c)
			case *widget.Panel:
				d.SetBgColor(c)
			}
		}
	}
	if cs.BackgroundImage != "" {
		if strings.HasPrefix(cs.BackgroundImage, "linear-gradient") {
			if grad, ok := css.ParseLinearGradient(cs.BackgroundImage); ok {
				if d, isDiv := w.(*widget.Div); isDiv {
					if len(grad.Stops) >= 2 {
						d.SetGradient(grad.Stops[0].Color, grad.Stops[len(grad.Stops)-1].Color, grad.Angle)
					}
				}
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
	if cs.BoxShadow != "" && cs.BoxShadow != "none" {
		if ox, oy, blur, color, ok := css.ParseBoxShadow(cs.BoxShadow); ok {
			if d, isDiv := w.(*widget.Div); isDiv {
				d.SetBoxShadow(ox, oy, blur, color)
			}
		}
	}
	if cs.Opacity != "" {
		op := css.ParseFloat(cs.Opacity)
		if op < 0 {
			op = 0
		}
		if op > 1 {
			op = 1
		}
		// Store clamped opacity in Raw for the rendering layer to read.
		if cs.Raw == nil {
			cs.Raw = make(map[string]string)
		}
		cs.Raw["opacity"] = fmt.Sprintf("%g", op)
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

// parseListItems parses <list-item> children within a <list> element.
// Returns the parsed items and advances past the closing </list> tag.
func (p *htmlParser) parseListItems(parentTag string) []widget.ListItem {
	var items []widget.ListItem
	for p.pos < len(p.src) {
		p.skipWhitespace()
		if p.pos >= len(p.src) {
			break
		}
		// Check for closing </list>
		closing := "</" + parentTag + ">"
		if p.pos+len(closing) <= len(p.src) && strings.ToLower(p.src[p.pos:p.pos+len(closing)]) == closing {
			p.pos += len(closing)
			break
		}
		if p.src[p.pos] != '<' {
			p.readUntil('<')
			continue
		}
		// Skip comments
		if p.pos+3 < len(p.src) && p.src[p.pos:p.pos+4] == "<!--" {
			p.skipComment()
			continue
		}
		// Read tag
		p.pos++ // skip '<'
		tagName := strings.ToLower(p.readTagName())
		attrs := p.readAttributes()
		// Skip self-closing or close '>'
		if p.pos < len(p.src) && p.src[p.pos] == '/' {
			p.pos++
		}
		if p.pos < len(p.src) && p.src[p.pos] == '>' {
			p.pos++
		}

		if tagName == "list-item" || tagName == "item" {
			text := p.readTextContent(tagName)
			item := widget.ListItem{Title: text}
			if v, ok := attrs["value"]; ok {
				item.Value = v
			} else {
				item.Value = text
			}
			if v, ok := attrs["group"]; ok {
				item.Group = v
			}
			items = append(items, item)
		} else {
			// Unknown child tag, skip it
			p.skipClosingTag(tagName)
		}
	}
	return items
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

package widget

import (
	"fmt"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Formula lays out a small expression grammar with the caller's font.
//
//	expr   = term { (+|-) term }
//	term   = factor { (*|/) factor }
//	factor = atom | frac(expr,expr) | sup(expr,expr) | sub(expr,expr) | sqrt(expr) | (expr)
//	atom   = letters and digits
//
// Fractions stack, superscripts rise, subscripts drop, and a radical is a
// rule plus the content glyphs. This is not TeX and does not embed a math font.
type Formula struct {
	Base
	src     string
	root    *fnode
	scrollX float32
}

type fnode struct {
	kind string
	role string
	text string
	x, y float32
	w, h float32
	size float32
	kids []*fnode
}

// NewFormula creates an empty formula view.
func NewFormula(tree *core.Tree, cfg *Config) *Formula {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Formula{Base: NewBase(tree, core.TypeCustom, cfg)}
}

// Expr is the last expression that parsed.
func (f *Formula) Expr() string { return f.src }

// SetExpr parses and lays out s. A parse error leaves the previous layout.
func (f *Formula) SetExpr(s string) error {
	root, err := parseFormula(s)
	if err != nil {
		return err
	}
	f.src = s
	f.root = root
	f.scrollX = 0
	f.layout(root)
	return nil
}

// SetScrollX moves the content horizontally. Vertical position stays put.
func (f *Formula) SetScrollX(x float32) {
	if x < 0 {
		x = 0
	}
	f.scrollX = x
}

// AtomScreen is the first atom with this text, in screen space.
func (f *Formula) AtomScreen(text string) uimath.Rect {
	n := findAtom(f.root, text)
	if n == nil {
		return uimath.Rect{}
	}
	b := f.Bounds()
	return uimath.NewRect(b.X+n.x-f.scrollX, b.Y+n.y, n.w, n.h)
}

// Hit returns the atom under the screen point and its role
// (num, den, base, exp, sub, or atom).
func (f *Formula) Hit(x, y float32) (string, string) {
	b := f.Bounds()
	if f.root == nil || x < b.X || y < b.Y || x >= b.X+b.Width || y >= b.Y+b.Height {
		return "", ""
	}
	lx := x - b.X + f.scrollX
	ly := y - b.Y
	var text, role string
	var walk func(*fnode, string)
	walk = func(n *fnode, inherited string) {
		if n == nil || lx < n.x || ly < n.y || lx >= n.x+n.w || ly >= n.y+n.h {
			return
		}
		r := inherited
		switch n.role {
		case "num", "den", "base", "exp", "sub":
			r = n.role
		}
		if n.kind == "atom" {
			text, role = n.text, r
		}
		for _, k := range n.kids {
			walk(k, r)
		}
	}
	walk(f.root, "atom")
	return text, role
}

// Draw paints atom glyphs and the fraction or radical rules. Nodes outside
// the widget are not emitted. The source string is not drawn.
func (f *Formula) Draw(buf *render.CommandBuffer) {
	b := f.Bounds()
	if b.IsEmpty() || f.root == nil {
		return
	}
	buf.PushClip(b)
	buf.DrawRect(render.RectCmd{Bounds: b, FillColor: f.config.BgColor}, 0, 1)
	f.drawNode(buf, f.root)
	buf.PopClip()
}

func (f *Formula) drawNode(buf *render.CommandBuffer, n *fnode) {
	if n == nil {
		return
	}
	b := f.Bounds()
	sx := b.X + n.x - f.scrollX
	sy := b.Y + n.y
	if sx+n.w <= b.X || sx >= b.X+b.Width || sy+n.h <= b.Y || sy >= b.Y+b.Height {
		return
	}
	cfg := f.config
	switch n.kind {
	case "atom":
		emitText(cfg, buf, n.text, sx, sy, n.size, n.w+1, cfg.TextColor, 1, 4)
		return
	case "frac":
		if len(n.kids) == 2 {
			num := n.kids[0]
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(b.X+n.x-f.scrollX, b.Y+num.y+num.h, n.w, 1),
				FillColor: cfg.TextColor,
			}, 2, 1)
		}
	case "sqrt":
		buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(sx, sy, 1, n.h), FillColor: cfg.TextColor}, 2, 1)
		buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(sx, sy, n.w, 1), FillColor: cfg.TextColor}, 2, 1)
	}
	for _, k := range n.kids {
		f.drawNode(buf, k)
	}
}

func (f *Formula) layout(n *fnode) {
	size := f.config.FontSizeSm
	if size <= 0 {
		size = 12
	}
	f.layoutAt(n, size)
	absPlace(n, 0, 0)
}

func (f *Formula) measure(s string, size float32) float32 {
	if f.config != nil && f.config.TextRenderer != nil {
		if w := f.config.TextRenderer.MeasureText(s, size); w > 0 {
			return w
		}
	}
	return float32(len([]rune(s))) * size * 0.6
}

func (f *Formula) layoutAt(n *fnode, size float32) {
	switch n.kind {
	case "atom":
		n.size = size
		n.w = f.measure(n.text, size)
		if n.w < 1 {
			n.w = size * 0.6
		}
		n.h = size
	case "row":
		var h float32
		for _, k := range n.kids {
			f.layoutAt(k, size)
			if k.h > h {
				h = k.h
			}
		}
		x := float32(0)
		const gap = float32(2)
		for i, k := range n.kids {
			if i > 0 {
				x += gap
			}
			k.x = x
			k.y = (h - k.h) / 2
			x += k.w
		}
		n.w = x
		n.h = h
	case "frac":
		f.layoutAt(n.kids[0], size)
		f.layoutAt(n.kids[1], size)
		const pad = float32(4)
		const bar = float32(4)
		num, den := n.kids[0], n.kids[1]
		n.w = num.w
		if den.w > n.w {
			n.w = den.w
		}
		n.w += pad
		n.h = num.h + bar + den.h
		num.x = (n.w - num.w) / 2
		den.x = (n.w - den.w) / 2
		num.y = 0
		den.y = num.h + bar
	case "sup":
		f.layoutAt(n.kids[0], size)
		expSize := size * 0.7
		f.layoutAt(n.kids[1], expSize)
		base, exp := n.kids[0], n.kids[1]
		raise := exp.h * 0.55
		n.w = base.w + exp.w
		n.h = base.h + raise
		base.x = 0
		base.y = raise
		exp.x = base.w
		exp.y = 0
	case "sub":
		f.layoutAt(n.kids[0], size)
		f.layoutAt(n.kids[1], size*0.7)
		base, sub := n.kids[0], n.kids[1]
		drop := sub.h * 0.45
		n.w = base.w + sub.w
		n.h = base.h + drop
		base.x = 0
		base.y = 0
		sub.x = base.w
		sub.y = base.h - sub.h + drop
	case "sqrt":
		f.layoutAt(n.kids[0], size)
		c := n.kids[0]
		n.w = c.w + 10
		n.h = c.h + 4
		c.x = 8
		c.y = 3
	}
}

func absPlace(n *fnode, ox, oy float32) {
	n.x += ox
	n.y += oy
	for _, k := range n.kids {
		absPlace(k, n.x, n.y)
	}
}

func findAtom(n *fnode, text string) *fnode {
	if n == nil {
		return nil
	}
	if n.kind == "atom" && n.text == text {
		return n
	}
	for _, k := range n.kids {
		if hit := findAtom(k, text); hit != nil {
			return hit
		}
	}
	return nil
}

type fparser struct {
	s string
	i int
}

func parseFormula(s string) (*fnode, error) {
	p := &fparser{s: s}
	n, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if p.peek() != 0 {
		return nil, fmt.Errorf("formula: trailing input")
	}
	if n == nil {
		return nil, fmt.Errorf("formula: empty")
	}
	return n, nil
}

func (p *fparser) skip() {
	for p.i < len(p.s) {
		switch p.s[p.i] {
		case ' ', '\t', '\n', '\r':
			p.i++
		default:
			return
		}
	}
}

func (p *fparser) peek() byte {
	p.skip()
	if p.i >= len(p.s) {
		return 0
	}
	return p.s[p.i]
}

func (p *fparser) eat(c byte) bool {
	if p.peek() != c {
		return false
	}
	p.i++
	return true
}

func (p *fparser) parseExpr() (*fnode, error) {
	n, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	for {
		op := p.peek()
		if op != '+' && op != '-' {
			break
		}
		p.i++
		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		n = joinRow(n, atomNode(string(op), "op"), right)
	}
	return n, nil
}

func (p *fparser) parseTerm() (*fnode, error) {
	n, err := p.parseFactor()
	if err != nil {
		return nil, err
	}
	for {
		op := p.peek()
		if op != '*' && op != '/' {
			break
		}
		p.i++
		right, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		n = joinRow(n, atomNode(string(op), "op"), right)
	}
	return n, nil
}

func (p *fparser) parseFactor() (*fnode, error) {
	if p.eat('(') {
		n, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if !p.eat(')') {
			return nil, fmt.Errorf("formula: missing )")
		}
		return n, nil
	}
	name := p.ident()
	if name == "" {
		return nil, fmt.Errorf("formula: expected atom")
	}
	if !p.eat('(') {
		return atomNode(name, "atom"), nil
	}
	switch name {
	case "frac", "sup", "sub":
		a, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if !p.eat(',') {
			return nil, fmt.Errorf("formula: missing ,")
		}
		b, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if !p.eat(')') {
			return nil, fmt.Errorf("formula: missing )")
		}
		return call2(name, a, b), nil
	case "sqrt":
		a, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if !p.eat(')') {
			return nil, fmt.Errorf("formula: missing )")
		}
		a.role = "rad"
		return &fnode{kind: "sqrt", kids: []*fnode{a}}, nil
	default:
		return nil, fmt.Errorf("formula: unknown call %s", name)
	}
}

func (p *fparser) ident() string {
	p.skip()
	start := p.i
	for p.i < len(p.s) {
		c := p.s[p.i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			p.i++
			continue
		}
		break
	}
	return p.s[start:p.i]
}

func atomNode(text, role string) *fnode {
	return &fnode{kind: "atom", role: role, text: text}
}

func call2(name string, a, b *fnode) *fnode {
	switch name {
	case "frac":
		a.role = "num"
		b.role = "den"
	case "sup":
		a.role = "base"
		b.role = "exp"
	case "sub":
		a.role = "base"
		b.role = "sub"
	}
	return &fnode{kind: name, kids: []*fnode{a, b}}
}

func joinRow(a, op, b *fnode) *fnode {
	if a != nil && a.kind == "row" {
		a.kids = append(a.kids, op, b)
		return a
	}
	return &fnode{kind: "row", kids: []*fnode{a, op, b}}
}

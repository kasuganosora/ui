package core

import (
	"fmt"
	"strconv"

	uimath "github.com/kasuganosora/ui/math"
)

// Shader is a tiny original program. It is not a GPU language and it does
// not compile for a graphics API. Eval runs on the CPU.
//
//	assign = name '=' expr
//	expr   = term { (+|-) term }
//	term   = unary { (*|/) unary }
//	unary  = factor | '-' unary
//	factor = number | name | name.swizzle | name '(' args ')' | '(' expr ')'
//
// uv is a vec2. out must be assigned. vec4, min, max, clamp, and mix are built in.
// Uniforms are floats supplied at Eval.
type Shader struct {
	code []shaderAssign
}

type shaderAssign struct {
	name string
	expr shaderExpr
}

// CompileShader parses src. A failure returns a nil program.
func CompileShader(src string) (*Shader, error) {
	p := &shaderParser{s: src}
	code, err := p.parseProgram()
	if err != nil {
		return nil, err
	}
	return &Shader{code: code}, nil
}

// Eval runs the program. A missing uniform or a missing out returns an error.
func (s *Shader) Eval(uv uimath.Vec2, uniforms map[string]float32) (uimath.Color, error) {
	if s == nil {
		return uimath.Color{}, fmt.Errorf("%w: nil shader", ErrState)
	}
	env := &shaderEnv{uv: uv, uniforms: uniforms, vars: map[string]shaderVal{}}
	for _, a := range s.code {
		v, err := a.expr.eval(env)
		if err != nil {
			return uimath.Color{}, err
		}
		env.vars[a.name] = v
	}
	out, ok := env.vars["out"]
	if !ok {
		return uimath.Color{}, fmt.Errorf("%w: out", ErrState)
	}
	return out.color(), nil
}

type shaderVal struct {
	f   float32
	v   [4]float32
	dim int
}

func (v shaderVal) color() uimath.Color {
	if v.dim == 1 {
		return uimath.Color{R: v.f, G: v.f, B: v.f, A: 1}
	}
	return uimath.Color{R: v.v[0], G: v.v[1], B: v.v[2], A: v.v[3]}
}

func shaderFloat(f float32) shaderVal { return shaderVal{f: f, dim: 1} }

type shaderEnv struct {
	uv       uimath.Vec2
	uniforms map[string]float32
	vars     map[string]shaderVal
}

type shaderExpr interface {
	eval(*shaderEnv) (shaderVal, error)
}

type numExpr struct{ f float32 }
type identExpr struct{ name string }
type swizzleExpr struct{ base, sw string }
type negExpr struct{ inner shaderExpr }
type binExpr struct {
	op   byte
	a, b shaderExpr
}
type callExpr struct {
	name string
	args []shaderExpr
}

func (e numExpr) eval(*shaderEnv) (shaderVal, error) { return shaderFloat(e.f), nil }

func (e identExpr) eval(env *shaderEnv) (shaderVal, error) {
	if e.name == "uv" {
		return shaderVal{v: [4]float32{env.uv.X, env.uv.Y}, dim: 2}, nil
	}
	if v, ok := env.vars[e.name]; ok {
		return v, nil
	}
	if f, ok := env.uniforms[e.name]; ok {
		return shaderFloat(f), nil
	}
	return shaderVal{}, fmt.Errorf("%w: %s", ErrState, e.name)
}

func (e swizzleExpr) eval(env *shaderEnv) (shaderVal, error) {
	base, err := (identExpr{e.base}).eval(env)
	if err != nil {
		return shaderVal{}, err
	}
	idx, ok := swizzleIndex(e.sw)
	if !ok || idx >= base.dim {
		return shaderVal{}, fmt.Errorf("%w: swizzle", ErrState)
	}
	if base.dim == 1 {
		return shaderFloat(base.f), nil
	}
	return shaderFloat(base.v[idx]), nil
}

func swizzleIndex(s string) (int, bool) {
	switch s {
	case "x", "r":
		return 0, true
	case "y", "g":
		return 1, true
	case "z", "b":
		return 2, true
	case "w", "a":
		return 3, true
	default:
		return 0, false
	}
}

func (e negExpr) eval(env *shaderEnv) (shaderVal, error) {
	v, err := e.inner.eval(env)
	if err != nil || v.dim != 1 {
		if err == nil {
			err = fmt.Errorf("%w: negate", ErrState)
		}
		return shaderVal{}, err
	}
	return shaderFloat(-v.f), nil
}

func (e binExpr) eval(env *shaderEnv) (shaderVal, error) {
	a, err := e.a.eval(env)
	if err != nil {
		return shaderVal{}, err
	}
	b, err := e.b.eval(env)
	if err != nil {
		return shaderVal{}, err
	}
	if a.dim != 1 || b.dim != 1 {
		return shaderVal{}, fmt.Errorf("%w: arithmetic", ErrState)
	}
	switch e.op {
	case '+':
		return shaderFloat(a.f + b.f), nil
	case '-':
		return shaderFloat(a.f - b.f), nil
	case '*':
		return shaderFloat(a.f * b.f), nil
	case '/':
		return shaderFloat(a.f / b.f), nil
	default:
		return shaderVal{}, fmt.Errorf("%w: op", ErrState)
	}
}

func (e callExpr) eval(env *shaderEnv) (shaderVal, error) {
	args := make([]shaderVal, len(e.args))
	for i, a := range e.args {
		v, err := a.eval(env)
		if err != nil {
			return shaderVal{}, err
		}
		args[i] = v
	}
	need := func(n int) error {
		if len(args) != n {
			return fmt.Errorf("%w: %s", ErrState, e.name)
		}
		for _, a := range args {
			if a.dim != 1 {
				return fmt.Errorf("%w: %s", ErrState, e.name)
			}
		}
		return nil
	}
	switch e.name {
	case "vec4":
		if err := need(4); err != nil {
			return shaderVal{}, err
		}
		return shaderVal{v: [4]float32{args[0].f, args[1].f, args[2].f, args[3].f}, dim: 4}, nil
	case "min":
		if err := need(2); err != nil {
			return shaderVal{}, err
		}
		if args[0].f < args[1].f {
			return shaderFloat(args[0].f), nil
		}
		return shaderFloat(args[1].f), nil
	case "max":
		if err := need(2); err != nil {
			return shaderVal{}, err
		}
		if args[0].f > args[1].f {
			return shaderFloat(args[0].f), nil
		}
		return shaderFloat(args[1].f), nil
	case "clamp":
		if err := need(3); err != nil {
			return shaderVal{}, err
		}
		x := args[0].f
		if x < args[1].f {
			x = args[1].f
		}
		if x > args[2].f {
			x = args[2].f
		}
		return shaderFloat(x), nil
	case "mix":
		if err := need(3); err != nil {
			return shaderVal{}, err
		}
		return shaderFloat(args[0].f + (args[1].f-args[0].f)*args[2].f), nil
	default:
		return shaderVal{}, fmt.Errorf("%w: %s", ErrState, e.name)
	}
}

type shaderParser struct {
	s string
	i int
}

func (p *shaderParser) parseProgram() ([]shaderAssign, error) {
	var out []shaderAssign
	for {
		p.skip()
		if p.i >= len(p.s) {
			break
		}
		name := p.ident()
		if name == "" || !p.eat('=') {
			return nil, fmt.Errorf("%w: shader syntax", ErrState)
		}
		ex, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		out = append(out, shaderAssign{name: name, expr: ex})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%w: empty shader", ErrState)
	}
	return out, nil
}

func (p *shaderParser) parseExpr() (shaderExpr, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	for {
		switch {
		case p.eat('+'):
			right, err := p.parseTerm()
			if err != nil {
				return nil, err
			}
			left = binExpr{op: '+', a: left, b: right}
		case p.eat('-'):
			right, err := p.parseTerm()
			if err != nil {
				return nil, err
			}
			left = binExpr{op: '-', a: left, b: right}
		default:
			return left, nil
		}
	}
}

func (p *shaderParser) parseTerm() (shaderExpr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for {
		switch {
		case p.eat('*'):
			right, err := p.parseUnary()
			if err != nil {
				return nil, err
			}
			left = binExpr{op: '*', a: left, b: right}
		case p.eat('/'):
			right, err := p.parseUnary()
			if err != nil {
				return nil, err
			}
			left = binExpr{op: '/', a: left, b: right}
		default:
			return left, nil
		}
	}
}

func (p *shaderParser) parseUnary() (shaderExpr, error) {
	if p.eat('-') {
		inner, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return negExpr{inner: inner}, nil
	}
	return p.parseFactor()
}

func (p *shaderParser) parseFactor() (shaderExpr, error) {
	p.skip()
	if p.i >= len(p.s) {
		return nil, fmt.Errorf("%w: shader syntax", ErrState)
	}
	if p.eat('(') {
		ex, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if !p.eat(')') {
			return nil, fmt.Errorf("%w: shader syntax", ErrState)
		}
		return ex, nil
	}
	if f, ok := p.number(); ok {
		return numExpr{f: f}, nil
	}
	name := p.ident()
	if name == "" {
		return nil, fmt.Errorf("%w: shader syntax", ErrState)
	}
	if p.eat('(') {
		var args []shaderExpr
		p.skip()
		if !p.peek(')') {
			for {
				ex, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				args = append(args, ex)
				if !p.eat(',') {
					break
				}
			}
		}
		if !p.eat(')') {
			return nil, fmt.Errorf("%w: shader syntax", ErrState)
		}
		return callExpr{name: name, args: args}, nil
	}
	if p.eat('.') {
		sw := p.ident()
		if sw == "" {
			return nil, fmt.Errorf("%w: swizzle", ErrState)
		}
		return swizzleExpr{base: name, sw: sw}, nil
	}
	return identExpr{name: name}, nil
}

func (p *shaderParser) skip() {
	for p.i < len(p.s) {
		switch p.s[p.i] {
		case ' ', '\t', '\n', '\r':
			p.i++
		default:
			return
		}
	}
}

func (p *shaderParser) peek(c byte) bool {
	p.skip()
	return p.i < len(p.s) && p.s[p.i] == c
}

func (p *shaderParser) eat(c byte) bool {
	if !p.peek(c) {
		return false
	}
	p.i++
	return true
}

func (p *shaderParser) ident() string {
	p.skip()
	if p.i >= len(p.s) {
		return ""
	}
	c := p.s[p.i]
	if c != '_' && (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') {
		return ""
	}
	start := p.i
	p.i++
	for p.i < len(p.s) {
		c = p.s[p.i]
		if c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			p.i++
			continue
		}
		break
	}
	return p.s[start:p.i]
}

func (p *shaderParser) number() (float32, bool) {
	p.skip()
	start := p.i
	saw := false
	for p.i < len(p.s) && p.s[p.i] >= '0' && p.s[p.i] <= '9' {
		p.i++
		saw = true
	}
	if p.i < len(p.s) && p.s[p.i] == '.' {
		p.i++
		for p.i < len(p.s) && p.s[p.i] >= '0' && p.s[p.i] <= '9' {
			p.i++
			saw = true
		}
	}
	if !saw {
		p.i = start
		return 0, false
	}
	f, err := strconv.ParseFloat(p.s[start:p.i], 32)
	if err != nil {
		p.i = start
		return 0, false
	}
	return float32(f), true
}

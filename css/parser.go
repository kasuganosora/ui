package css

import "strings"

// Stylesheet is a parsed CSS stylesheet.
type Stylesheet struct {
	Rules []Rule
	// Variables stores :root CSS custom properties (--name: value).
	Variables map[string]string
}

// Rule is a CSS rule: selectors + declarations.
type Rule struct {
	Selectors    []Selector
	Declarations []Declaration
	// Source order index for cascade tie-breaking.
	Order int
}

// Declaration is a single CSS property: value pair.
type Declaration struct {
	Property  string
	Value     string
	Important bool
}

// Parse parses a CSS stylesheet string into a Stylesheet.
func Parse(src string) *Stylesheet {
	p := &cssParser{
		tokens: newTokenizer(src).tokenize(),
		sheet:  &Stylesheet{Variables: make(map[string]string)},
	}
	p.parse()
	return p.sheet
}

type cssParser struct {
	tokens []Token
	pos    int
	sheet  *Stylesheet
	order  int
}

func (p *cssParser) peek() Token {
	for p.pos < len(p.tokens) && p.tokens[p.pos].Type == TokenWhitespace {
		p.pos++
	}
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos]
}

// peekRaw does not skip whitespace.
func (p *cssParser) peekRaw() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *cssParser) advance() Token {
	tok := p.tokens[p.pos]
	p.pos++
	return tok
}

func (p *cssParser) parse() {
	for p.peek().Type != TokenEOF {
		tok := p.peek()
		if tok.Type == TokenAt {
			p.skipAtRule()
			continue
		}
		p.parseRule()
	}
}

// skipAtRule skips @media, @keyframes, etc. — not supported yet.
func (p *cssParser) skipAtRule() {
	p.advance() // @
	// Skip until { or ;
	for {
		tok := p.peek()
		if tok.Type == TokenEOF {
			return
		}
		if tok.Type == TokenSemicolon {
			p.advance()
			return
		}
		if tok.Type == TokenLBrace {
			p.skipBlock()
			return
		}
		p.advance()
	}
}

func (p *cssParser) skipBlock() {
	p.advance() // {
	depth := 1
	for depth > 0 {
		tok := p.peekRaw()
		if tok.Type == TokenEOF {
			return
		}
		p.advance()
		if tok.Type == TokenLBrace {
			depth++
		} else if tok.Type == TokenRBrace {
			depth--
		}
	}
}

func (p *cssParser) parseRule() {
	// Read selector tokens until {
	selectorStart := p.pos
	for {
		tok := p.peekRaw()
		if tok.Type == TokenEOF || tok.Type == TokenLBrace {
			break
		}
		p.pos++
	}
	selectorText := p.tokensToString(selectorStart, p.pos)

	if p.peek().Type != TokenLBrace {
		return
	}
	p.advance() // {

	declarations := p.parseDeclarations()

	if p.peek().Type == TokenRBrace {
		p.advance()
	}

	selectors := ParseSelectorList(selectorText)
	if len(selectors) == 0 {
		return
	}

	// Check for :root selector → extract variables
	isRoot := false
	for _, sel := range selectors {
		if len(sel.Parts) == 1 && sel.Parts[0].Compound != nil {
			c := sel.Parts[0].Compound
			if len(c.PseudoClass) == 1 && c.PseudoClass[0] == "root" && c.Tag == "" && c.ID == "" && len(c.Classes) == 0 {
				isRoot = true
			}
		}
	}

	if isRoot {
		for _, decl := range declarations {
			if strings.HasPrefix(decl.Property, "--") {
				p.sheet.Variables[decl.Property] = decl.Value
			}
		}
		// Still add as a rule (might have non-variable properties too)
	}

	rule := Rule{
		Selectors:    selectors,
		Declarations: declarations,
		Order:        p.order,
	}
	p.order++
	p.sheet.Rules = append(p.sheet.Rules, rule)
}

func (p *cssParser) parseDeclarations() []Declaration {
	var decls []Declaration
	for {
		tok := p.peek()
		if tok.Type == TokenEOF || tok.Type == TokenRBrace {
			break
		}

		decl, ok := p.parseDeclaration()
		if ok {
			decls = append(decls, decl)
		}
	}
	return decls
}

func (p *cssParser) parseDeclaration() (Declaration, bool) {
	tok := p.peek()
	if tok.Type != TokenIdent && !(tok.Type == TokenDelim && tok.Value == "-") {
		// Try to recover: skip to next ; or }
		p.skipToSemicolonOrBrace()
		return Declaration{}, false
	}

	// Property name (could be custom property like --foo)
	var propParts []string
	for {
		t := p.peekRaw()
		if t.Type == TokenColon || t.Type == TokenEOF || t.Type == TokenRBrace || t.Type == TokenSemicolon {
			break
		}
		if t.Type == TokenWhitespace {
			p.pos++
			continue
		}
		propParts = append(propParts, t.Value)
		p.pos++
	}
	if p.peek().Type != TokenColon {
		p.skipToSemicolonOrBrace()
		return Declaration{}, false
	}
	p.advance() // :

	property := strings.ToLower(strings.Join(propParts, ""))

	// Read value tokens until ; or } or !important
	var valueParts []string
	important := false
	for {
		t := p.peekRaw()
		if t.Type == TokenEOF || t.Type == TokenRBrace {
			break
		}
		if t.Type == TokenSemicolon {
			p.advance()
			break
		}
		p.advance()
		// Function tokens include the name but not the '('
		if t.Type == TokenFunction {
			valueParts = append(valueParts, t.Value+"(")
			continue
		}
		if t.Type == TokenExcl {
			// Check for !important
			next := p.peek()
			if next.Type == TokenIdent && strings.ToLower(next.Value) == "important" {
				important = true
				p.advance()
				// Skip to ; or }
				for {
					tt := p.peekRaw()
					if tt.Type == TokenSemicolon {
						p.advance()
						break
					}
					if tt.Type == TokenEOF || tt.Type == TokenRBrace {
						break
					}
					p.advance()
				}
				break
			}
			valueParts = append(valueParts, "!")
			continue
		}
		valueParts = append(valueParts, t.Value)
	}

	value := strings.TrimSpace(strings.Join(valueParts, ""))
	if property == "" || value == "" {
		return Declaration{}, false
	}

	return Declaration{
		Property:  property,
		Value:     value,
		Important: important,
	}, true
}

func (p *cssParser) skipToSemicolonOrBrace() {
	for {
		tok := p.peekRaw()
		if tok.Type == TokenEOF || tok.Type == TokenRBrace {
			return
		}
		if tok.Type == TokenSemicolon {
			p.advance()
			return
		}
		p.advance()
	}
}

func (p *cssParser) tokensToString(start, end int) string {
	var sb strings.Builder
	for i := start; i < end; i++ {
		sb.WriteString(p.tokens[i].Value)
	}
	return strings.TrimSpace(sb.String())
}

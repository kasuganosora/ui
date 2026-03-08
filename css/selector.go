package css

import "strings"

// Selector represents a CSS selector (e.g. "div.foo > #bar:hover").
// A selector is a list of compound selectors joined by combinators.
type Selector struct {
	Parts []SelectorPart
}

// SelectorPart is either a compound selector or a combinator.
type SelectorPart struct {
	Compound   *CompoundSelector // non-nil for compound selectors
	Combinator Combinator        // non-zero for combinators
}

// Combinator defines the relationship between compound selectors.
type Combinator uint8

const (
	CombinatorNone       Combinator = iota
	CombinatorDescendant            // space (A B)
	CombinatorChild                 // > (A > B)
	CombinatorAdjacent              // + (A + B)
	CombinatorSibling               // ~ (A ~ B)
)

// CompoundSelector is a sequence of simple selectors that must all match.
type CompoundSelector struct {
	Tag          string   // "" means any
	ID           string   // without #
	Classes      []string // without .
	PseudoClass  []string // e.g. "hover", "focus", "first-child"
	Universal    bool     // *
}

// Specificity is (inline, ids, classes, tags) used for cascade ordering.
type Specificity [4]int

// Less returns true if s has lower priority than other.
func (s Specificity) Less(other Specificity) bool {
	for i := 0; i < 4; i++ {
		if s[i] != other[i] {
			return s[i] < other[i]
		}
	}
	return false
}

// SelectorSpecificity computes the specificity of a selector.
func SelectorSpecificity(sel *Selector) Specificity {
	var spec Specificity
	for _, part := range sel.Parts {
		if part.Compound == nil {
			continue
		}
		c := part.Compound
		if c.ID != "" {
			spec[1]++
		}
		spec[2] += len(c.Classes) + len(c.PseudoClass)
		if c.Tag != "" {
			spec[3]++
		}
	}
	return spec
}

// ParseSelector parses a CSS selector string into a Selector.
// Supports: tag, .class, #id, *, combinators (space, >, +, ~), pseudo-classes (:hover etc.).
// Multiple selectors separated by commas are NOT handled here; use ParseSelectorList.
func ParseSelector(s string) Selector {
	p := &selectorParser{tokens: newTokenizer(s).tokenize()}
	return p.parseSelector()
}

// ParseSelectorList parses comma-separated selectors: "a, b, c".
func ParseSelectorList(s string) []Selector {
	p := &selectorParser{tokens: newTokenizer(s).tokenize()}
	return p.parseSelectorList()
}

type selectorParser struct {
	tokens []Token
	pos    int
}

func (p *selectorParser) peek() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *selectorParser) advance() Token {
	tok := p.tokens[p.pos]
	p.pos++
	return tok
}

func (p *selectorParser) parseSelectorList() []Selector {
	var sels []Selector
	sels = append(sels, p.parseSelector())
	for p.peek().Type == TokenComma {
		p.advance() // skip ,
		p.skipWS()
		sels = append(sels, p.parseSelector())
	}
	return sels
}

func (p *selectorParser) parseSelector() Selector {
	p.skipWS()
	sel := Selector{}

	compound := p.parseCompound()
	if compound == nil {
		return sel
	}
	sel.Parts = append(sel.Parts, SelectorPart{Compound: compound})

	for {
		// Check for combinator or end
		tok := p.peek()
		if tok.Type == TokenEOF || tok.Type == TokenComma || tok.Type == TokenLBrace {
			break
		}

		var comb Combinator
		if tok.Type == TokenGT {
			comb = CombinatorChild
			p.advance()
			p.skipWS()
		} else if tok.Type == TokenPlus {
			comb = CombinatorAdjacent
			p.advance()
			p.skipWS()
		} else if tok.Type == TokenTilde {
			comb = CombinatorSibling
			p.advance()
			p.skipWS()
		} else if tok.Type == TokenWhitespace {
			p.skipWS()
			// Check if next is a combinator
			next := p.peek()
			if next.Type == TokenGT {
				comb = CombinatorChild
				p.advance()
				p.skipWS()
			} else if next.Type == TokenPlus {
				comb = CombinatorAdjacent
				p.advance()
				p.skipWS()
			} else if next.Type == TokenTilde {
				comb = CombinatorSibling
				p.advance()
				p.skipWS()
			} else if next.Type == TokenEOF || next.Type == TokenComma || next.Type == TokenLBrace {
				break
			} else {
				comb = CombinatorDescendant
			}
		} else {
			break
		}

		compound = p.parseCompound()
		if compound == nil {
			break
		}
		sel.Parts = append(sel.Parts, SelectorPart{Combinator: comb})
		sel.Parts = append(sel.Parts, SelectorPart{Compound: compound})
	}

	return sel
}

func (p *selectorParser) parseCompound() *CompoundSelector {
	cs := &CompoundSelector{}
	matched := false

	for {
		tok := p.peek()
		switch tok.Type {
		case TokenIdent:
			cs.Tag = strings.ToLower(tok.Value)
			p.advance()
			matched = true
		case TokenStar:
			cs.Universal = true
			p.advance()
			matched = true
		case TokenDot:
			p.advance()
			next := p.peek()
			if next.Type == TokenIdent {
				cs.Classes = append(cs.Classes, next.Value)
				p.advance()
				matched = true
			}
		case TokenHash:
			cs.ID = tok.Value[1:] // remove #
			p.advance()
			matched = true
		case TokenColon:
			p.advance()
			// Double colon (::) pseudo-elements — skip for now
			if p.peek().Type == TokenColon {
				p.advance()
			}
			next := p.peek()
			if next.Type == TokenIdent {
				cs.PseudoClass = append(cs.PseudoClass, next.Value)
				p.advance()
				matched = true
			} else if next.Type == TokenFunction {
				// e.g. :nth-child(2n+1)
				fname := next.Value
				p.advance()
				// skip until matching )
				depth := 1
				for depth > 0 && p.peek().Type != TokenEOF {
					if p.peek().Type == TokenLParen {
						depth++
					} else if p.peek().Type == TokenRParen {
						depth--
					}
					p.advance()
				}
				cs.PseudoClass = append(cs.PseudoClass, fname)
				matched = true
			}
		default:
			if !matched {
				return nil
			}
			return cs
		}
	}
}

func (p *selectorParser) skipWS() {
	for p.peek().Type == TokenWhitespace {
		p.advance()
	}
}

// MatchElement checks if a selector matches an element with the given properties.
type ElementInfo struct {
	Tag      string
	ID       string
	Classes  []string
	Hovered  bool
	Focused  bool
	Active   bool
	Disabled bool
	// For structural pseudo-classes
	ChildIndex int // 0-based index among siblings
	SiblingCount int
}

// MatchCompound checks if a compound selector matches an element.
func MatchCompound(cs *CompoundSelector, el *ElementInfo) bool {
	if cs.Tag != "" && cs.Tag != el.Tag {
		return false
	}
	if cs.ID != "" && cs.ID != el.ID {
		return false
	}
	for _, cls := range cs.Classes {
		if !containsStr(el.Classes, cls) {
			return false
		}
	}
	for _, pseudo := range cs.PseudoClass {
		if !matchPseudo(pseudo, el) {
			return false
		}
	}
	return true
}

func matchPseudo(pseudo string, el *ElementInfo) bool {
	switch pseudo {
	case "hover":
		return el.Hovered
	case "focus":
		return el.Focused
	case "active":
		return el.Active
	case "disabled":
		return el.Disabled
	case "first-child":
		return el.ChildIndex == 0
	case "last-child":
		return el.SiblingCount > 0 && el.ChildIndex == el.SiblingCount-1
	default:
		return false
	}
}

func containsStr(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

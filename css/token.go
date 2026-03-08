// Package css provides CSS parsing, selector matching, and style resolution.
//
// It parses CSS stylesheets into rules, matches selectors against elements,
// computes specificity, and resolves cascaded styles to layout.Style values.
package css

// TokenType identifies the kind of CSS token.
type TokenType uint8

const (
	TokenEOF TokenType = iota
	TokenIdent         // e.g. "color", "div", "flex"
	TokenHash          // e.g. "#id", "#ff0000"
	TokenString        // e.g. "hello", 'world'
	TokenNumber        // e.g. "12", "3.14"
	TokenDimension     // e.g. "12px", "50%", "1.5em"
	TokenColon         // :
	TokenSemicolon     // ;
	TokenComma         // ,
	TokenDot           // .
	TokenLBrace        // {
	TokenRBrace        // }
	TokenLParen        // (
	TokenRParen        // )
	TokenLBracket      // [
	TokenRBracket      // ]
	TokenGT            // >
	TokenPlus          // +
	TokenTilde         // ~
	TokenStar          // *
	TokenSlash         // /
	TokenAt            // @
	TokenExcl          // !
	TokenWhitespace    // spaces/tabs/newlines
	TokenFunction      // e.g. "rgb(", "var("
	TokenDelim         // any other single character
)

// Token is a single CSS token.
type Token struct {
	Type  TokenType
	Value string
	// For TokenNumber/TokenDimension:
	Num  float32
	Unit string // "px", "%", "em", "rem", etc.
}

// tokenizer splits CSS source into tokens.
type tokenizer struct {
	src []byte
	pos int
}

func newTokenizer(src string) *tokenizer {
	return &tokenizer{src: []byte(src)}
}

func (t *tokenizer) peek() byte {
	if t.pos >= len(t.src) {
		return 0
	}
	return t.src[t.pos]
}

func (t *tokenizer) advance() byte {
	ch := t.src[t.pos]
	t.pos++
	return ch
}

func (t *tokenizer) tokenize() []Token {
	var tokens []Token
	for t.pos < len(t.src) {
		tok := t.next()
		if tok.Type == TokenEOF {
			break
		}
		tokens = append(tokens, tok)
	}
	return tokens
}

func (t *tokenizer) next() Token {
	if t.pos >= len(t.src) {
		return Token{Type: TokenEOF}
	}

	ch := t.peek()

	// Whitespace
	if isWhitespace(ch) {
		return t.readWhitespace()
	}

	// Comments
	if ch == '/' && t.pos+1 < len(t.src) && t.src[t.pos+1] == '*' {
		t.skipComment()
		return t.next()
	}

	// String
	if ch == '"' || ch == '\'' {
		return t.readString()
	}

	// Hash
	if ch == '#' {
		t.advance()
		name := t.readName()
		return Token{Type: TokenHash, Value: "#" + name}
	}

	// Number or dimension
	if isDigit(ch) || (ch == '.' && t.pos+1 < len(t.src) && isDigit(t.src[t.pos+1])) {
		return t.readNumeric()
	}
	if ch == '-' && t.pos+1 < len(t.src) && (isDigit(t.src[t.pos+1]) || t.src[t.pos+1] == '.') {
		return t.readNumeric()
	}

	// Ident or function
	if isNameStart(ch) || (ch == '-' && t.pos+1 < len(t.src) && (isNameStart(t.src[t.pos+1]) || t.src[t.pos+1] == '-')) {
		return t.readIdentOrFunction()
	}

	// Single-character tokens
	t.advance()
	switch ch {
	case ':':
		return Token{Type: TokenColon, Value: ":"}
	case ';':
		return Token{Type: TokenSemicolon, Value: ";"}
	case ',':
		return Token{Type: TokenComma, Value: ","}
	case '.':
		return Token{Type: TokenDot, Value: "."}
	case '{':
		return Token{Type: TokenLBrace, Value: "{"}
	case '}':
		return Token{Type: TokenRBrace, Value: "}"}
	case '(':
		return Token{Type: TokenLParen, Value: "("}
	case ')':
		return Token{Type: TokenRParen, Value: ")"}
	case '[':
		return Token{Type: TokenLBracket, Value: "["}
	case ']':
		return Token{Type: TokenRBracket, Value: "]"}
	case '>':
		return Token{Type: TokenGT, Value: ">"}
	case '+':
		return Token{Type: TokenPlus, Value: "+"}
	case '~':
		return Token{Type: TokenTilde, Value: "~"}
	case '*':
		return Token{Type: TokenStar, Value: "*"}
	case '/':
		return Token{Type: TokenSlash, Value: "/"}
	case '@':
		return Token{Type: TokenAt, Value: "@"}
	case '!':
		return Token{Type: TokenExcl, Value: "!"}
	default:
		return Token{Type: TokenDelim, Value: string(ch)}
	}
}

func (t *tokenizer) readWhitespace() Token {
	start := t.pos
	for t.pos < len(t.src) && isWhitespace(t.src[t.pos]) {
		t.pos++
	}
	return Token{Type: TokenWhitespace, Value: string(t.src[start:t.pos])}
}

func (t *tokenizer) skipComment() {
	t.pos += 2 // skip /*
	for t.pos+1 < len(t.src) {
		if t.src[t.pos] == '*' && t.src[t.pos+1] == '/' {
			t.pos += 2
			return
		}
		t.pos++
	}
	t.pos = len(t.src) // unterminated comment
}

func (t *tokenizer) readString() Token {
	quote := t.advance()
	start := t.pos
	for t.pos < len(t.src) && t.src[t.pos] != quote {
		if t.src[t.pos] == '\\' && t.pos+1 < len(t.src) {
			t.pos++ // skip escape
		}
		t.pos++
	}
	val := string(t.src[start:t.pos])
	if t.pos < len(t.src) {
		t.pos++ // skip closing quote
	}
	return Token{Type: TokenString, Value: val}
}

func (t *tokenizer) readNumeric() Token {
	start := t.pos
	if t.src[t.pos] == '-' {
		t.pos++
	}
	for t.pos < len(t.src) && isDigit(t.src[t.pos]) {
		t.pos++
	}
	if t.pos < len(t.src) && t.src[t.pos] == '.' {
		t.pos++
		for t.pos < len(t.src) && isDigit(t.src[t.pos]) {
			t.pos++
		}
	}
	numStr := string(t.src[start:t.pos])
	num := parseFloat32(numStr)

	// Check for unit suffix
	if t.pos < len(t.src) && (isNameStart(t.src[t.pos]) || t.src[t.pos] == '%') {
		unitStart := t.pos
		if t.src[t.pos] == '%' {
			t.pos++
		} else {
			for t.pos < len(t.src) && isNameChar(t.src[t.pos]) {
				t.pos++
			}
		}
		unit := string(t.src[unitStart:t.pos])
		return Token{Type: TokenDimension, Value: numStr + unit, Num: num, Unit: unit}
	}

	return Token{Type: TokenNumber, Value: numStr, Num: num}
}

func (t *tokenizer) readIdentOrFunction() Token {
	name := t.readName()
	if t.pos < len(t.src) && t.src[t.pos] == '(' {
		t.pos++ // consume '('
		return Token{Type: TokenFunction, Value: name}
	}
	return Token{Type: TokenIdent, Value: name}
}

func (t *tokenizer) readName() string {
	start := t.pos
	for t.pos < len(t.src) && isNameChar(t.src[t.pos]) {
		t.pos++
	}
	return string(t.src[start:t.pos])
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == '\f'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isNameStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isNameChar(ch byte) bool {
	return isNameStart(ch) || isDigit(ch) || ch == '-'
}

func parseFloat32(s string) float32 {
	var neg bool
	i := 0
	if i < len(s) && s[i] == '-' {
		neg = true
		i++
	}
	var whole float32
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		whole = whole*10 + float32(s[i]-'0')
		i++
	}
	var frac float32
	if i < len(s) && s[i] == '.' {
		i++
		var div float32 = 10
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			frac += float32(s[i]-'0') / div
			div *= 10
			i++
		}
	}
	v := whole + frac
	if neg {
		v = -v
	}
	return v
}

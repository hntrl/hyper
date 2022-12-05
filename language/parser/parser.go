package parser

import (
	"github.com/hntrl/lang/language/tokens"
)

type BufferItem struct {
	pos tokens.Position
	tok tokens.Token
	lit string
}
type TokenBuffer struct {
	tokens []BufferItem
	n      int
}
type Parser struct {
	lex    *Lexer
	buffer TokenBuffer
}

func NewParser(lex *Lexer) *Parser {
	return &Parser{
		lex: lex,
		buffer: TokenBuffer{
			tokens: []BufferItem{},
			n:      -1,
		},
	}
}

func (p *Parser) Scan() (tokens.Position, tokens.Token, string) {
	p.buffer.n += 1
	if p.buffer.n < len(p.buffer.tokens) {
		token := p.buffer.tokens[p.buffer.n]
		return token.pos, token.tok, token.lit
	}
	pos, tok, lit := p.lex.Lex()
	p.buffer.tokens = append(p.buffer.tokens, BufferItem{pos, tok, lit})
	return pos, tok, lit
}

func (p *Parser) ScanIgnore(tokens ...tokens.Token) (tokens.Position, tokens.Token, string) {
	for {
		ignore := false
		pos, tok, lit := p.Scan()
		for _, ignoredToken := range tokens {
			if tok == ignoredToken {
				ignore = true
				break
			}
		}
		if !ignore {
			return pos, tok, lit
		}
	}
}

func (p *Parser) Unscan() {
	p.buffer.n -= 1
}

func (p *Parser) Rollback(n int) {
	p.buffer.n = n
}

func (p *Parser) Index() int {
	return p.buffer.n
}

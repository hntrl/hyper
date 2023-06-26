package parser

import (
	"github.com/hntrl/hyper/src/hyper/tokens"
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

func (p *Parser) ScanUntil(stopChars ...rune) (tokens.Position, string) {
	pos := p.lex.pos
	val := ""
	if p.buffer.n < len(p.buffer.tokens) {
		panic("cannot ScanUntil when buffer cursor is not current: would be rewriting buffer")
	}
scan:
	for {
		r := p.lex.read()
		if r == '\\' {
			next := p.lex.peek()
			for _, stopChar := range stopChars {
				if next == stopChar {
					p.lex.read()
					val += string(next)
					continue scan
				}
			}
			p.lex.backup()
			val += p.lex.lexEscaped()
		} else {
			for _, stopChar := range stopChars {
				if r == stopChar {
					p.lex.backup()
					p.buffer.n += 1
					p.buffer.tokens = append(p.buffer.tokens, BufferItem{pos, tokens.IDENT, val})
					return pos, val
				}
			}
			val += string(r)
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

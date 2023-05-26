package ast

import (
	"strconv"

	"github.com/hntrl/hyper/internal/parser"
	"github.com/hntrl/hyper/internal/tokens"
)

// Selector :: IDENT
//
//	| IDENT DOT Selector
type Selector struct {
	pos     tokens.Position
	Members []string
}

func (s Selector) Validate() error {
	return nil
}

func (s Selector) Pos() tokens.Position {
	return s.pos
}

func ParseSelector(p *parser.Parser) (*Selector, error) {
	selector := Selector{Members: []string{}}
	for {
		pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		if selector.pos == (tokens.Position{}) {
			selector.pos = pos
		}
		if tok != tokens.IDENT {
			p.Unscan()
			break
		}
		selector.Members = append(selector.Members, lit)
		_, tok, _ = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		if tok != tokens.PERIOD {
			p.Unscan()
			break
		}
	}
	return &selector, nil
}

// Literal :: STRING
//
//	| INT
//	| FLOAT
type Literal struct {
	pos   tokens.Position
	Value interface{}
}

func (l Literal) Validate() error {
	return nil
}

func (l Literal) Pos() tokens.Position {
	return l.pos
}

func ParseLiteral(p *parser.Parser) (*Literal, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	switch tok {
	case tokens.INT:
		val, err := strconv.ParseInt(lit, 10, 64)
		if err != nil {
			return nil, err
		}
		return &Literal{pos: pos, Value: val}, nil
	case tokens.FLOAT:
		val, err := strconv.ParseFloat(lit, 64)
		if err != nil {
			return nil, err
		}
		return &Literal{pos: pos, Value: val}, nil
	case tokens.STRING:
		return &Literal{pos: pos, Value: lit}, nil
	default:
		return nil, ExpectedError(pos, tokens.INT, lit)
	}
}

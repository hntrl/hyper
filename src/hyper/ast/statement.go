package ast

import (
	"fmt"

	"github.com/hntrl/hyper/src/hyper/parser"
	"github.com/hntrl/hyper/src/hyper/tokens"
)

// FieldStatement ::
//
//	| COMMENT? FieldAssignmentExpression
//	| COMMENT? EnumExpression
//	| COMMENT? FieldExpression
type FieldStatement struct {
	pos     tokens.Position
	Init    Node `types:"FieldAssignmentExpression,EnumExpression,FieldExpression"`
	Comment string
}

func (f FieldStatement) Validate() error {
	if as, ok := f.Init.(FieldAssignmentExpression); ok {
		if err := as.Validate(); err != nil {
			return err
		}
	} else if es, ok := f.Init.(EnumExpression); ok {
		if err := es.Validate(); err != nil {
			return err
		}
	} else if ts, ok := f.Init.(FieldExpression); ok {
		if err := ts.Validate(); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("parsing %T not allowed in FieldStatement", f.Init)
	}
	return nil
}

func (f FieldStatement) Pos() tokens.Position {
	return f.pos
}

func ParseFieldStatement(p *parser.Parser) (*FieldStatement, error) {
	field := FieldStatement{}

	for {
		_, tok, lit := p.ScanIgnore(tokens.NEWLINE)
		if tok == tokens.COMMENT {
			_, tok, _ := p.Scan()
			if tok == tokens.NEWLINE || tok == tokens.RCURLY {
				continue
			} else {
				field.Comment = lit
				p.Unscan()
				break
			}
		}
		p.Unscan()
		break
	}

	pos, _, _ := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	startIndex := p.Index() - 1
	field.pos = pos

	_, startingToken, _ := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	p.Rollback(startIndex)
	switch startingToken {
	case tokens.ASSIGN:
		as, err := ParseFieldAssignmentExpression(p)
		if err != nil {
			return nil, err
		}
		field.Init = *as
	case tokens.STRING:
		es, err := ParseEnumExpression(p)
		if err != nil {
			return nil, err
		}
		field.Init = *es
	default:
		ts, err := ParseFieldExpression(p)
		if err != nil {
			return nil, err
		}
		field.Init = *ts
	}
	return &field, nil
}

// FieldAssignmentExpression :: IDENT ASSIGN Expression
type FieldAssignmentExpression struct {
	pos  tokens.Position
	Name string
	Init Expression
}

func (a FieldAssignmentExpression) Validate() error {
	if err := a.Init.Validate(); err != nil {
		return err
	}
	return nil
}

func (a FieldAssignmentExpression) Pos() tokens.Position {
	return a.pos
}

func ParseFieldAssignmentExpression(p *parser.Parser) (*FieldAssignmentExpression, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	as := FieldAssignmentExpression{pos: pos}
	if tok != tokens.IDENT {
		return nil, ExpectedError(pos, tokens.IDENT, lit)
	}
	as.Name = lit

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.ASSIGN {
		return nil, ExpectedError(pos, tokens.ASSIGN, lit)
	}

	expr, err := ParseExpression(p)
	if err != nil {
		return nil, err
	}
	as.Init = *expr

	return &as, nil
}

// EnumExpression :: IDENT STRING
type EnumExpression struct {
	pos  tokens.Position
	Name string
	Init string
}

func (e EnumExpression) Validate() error {
	return nil
}

func (e EnumExpression) Pos() tokens.Position {
	return e.pos
}

func ParseEnumExpression(p *parser.Parser) (*EnumExpression, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	es := EnumExpression{pos: pos}
	if tok != tokens.IDENT {
		return nil, ExpectedError(pos, tokens.IDENT, lit)
	}
	es.Name = lit

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.STRING {
		return nil, ExpectedError(pos, tokens.STRING, lit)
	}
	es.Init = lit

	return &es, nil
}

// FieldExpression :: IDENT TypeExpression
type FieldExpression struct {
	pos  tokens.Position
	Name string
	Init TypeExpression
}

func (t FieldExpression) Validate() error {
	if err := t.Init.Validate(); err != nil {
		return err
	}
	return nil
}

func (t FieldExpression) Pos() tokens.Position {
	return t.pos
}

func ParseFieldExpression(p *parser.Parser) (*FieldExpression, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	ts := FieldExpression{pos: pos}
	if tok != tokens.IDENT {
		return nil, ExpectedError(pos, tokens.IDENT, lit)
	}
	ts.Name = lit

	te, err := ParseTypeExpression(p)
	if err != nil {
		return nil, err
	}
	ts.Init = *te

	return &ts, nil
}

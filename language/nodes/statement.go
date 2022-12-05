package nodes

import (
	"fmt"

	"github.com/hntrl/lang/language/parser"
	"github.com/hntrl/lang/language/tokens"
)

// FieldStatement ::
//
//	| COMMENT? AssignmentStatement
//	| COMMENT? EnumStatement
//	| COMMENT? TypeStatement
type FieldStatement struct {
	pos     tokens.Position
	Init    Node `types:"AssignmentStatement,EnumStatement,TypeStatement"`
	Comment string
}

func (f FieldStatement) Validate() error {
	if as, ok := f.Init.(AssignmentStatement); ok {
		if err := as.Validate(); err != nil {
			return err
		}
	} else if es, ok := f.Init.(EnumStatement); ok {
		if err := es.Validate(); err != nil {
			return err
		}
	} else if ts, ok := f.Init.(TypeStatement); ok {
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
		as, err := ParseAssignmentStatement(p)
		if err != nil {
			return nil, err
		}
		field.Init = *as
	case tokens.STRING:
		es, err := ParseEnumStatement(p)
		if err != nil {
			return nil, err
		}
		field.Init = *es
	default:
		ts, err := ParseTypeStatement(p)
		if err != nil {
			return nil, err
		}
		field.Init = *ts
	}
	return &field, nil
}

// AssignmentStatement :: IDENT ASSIGN Expression
type AssignmentStatement struct {
	pos  tokens.Position
	Name string
	Init Expression
}

func (a AssignmentStatement) Validate() error {
	if err := a.Init.Validate(); err != nil {
		return err
	}
	return nil
}

func (a AssignmentStatement) Pos() tokens.Position {
	return a.pos
}

func ParseAssignmentStatement(p *parser.Parser) (*AssignmentStatement, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	as := AssignmentStatement{pos: pos}
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

// EnumStatement :: IDENT STRING
type EnumStatement struct {
	pos  tokens.Position
	Name string
	Init string
}

func (e EnumStatement) Validate() error {
	return nil
}

func (e EnumStatement) Pos() tokens.Position {
	return e.pos
}

func ParseEnumStatement(p *parser.Parser) (*EnumStatement, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	es := EnumStatement{pos: pos}
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

// TypeStatement :: IDENT TypeExpression
type TypeStatement struct {
	pos  tokens.Position
	Name string
	Init TypeExpression
}

func (t TypeStatement) Validate() error {
	if err := t.Init.Validate(); err != nil {
		return err
	}
	return nil
}

func (t TypeStatement) Pos() tokens.Position {
	return t.pos
}

func ParseTypeStatement(p *parser.Parser) (*TypeStatement, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	ts := TypeStatement{pos: pos}
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

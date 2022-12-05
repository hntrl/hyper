package nodes

import (
	"fmt"

	"github.com/hntrl/lang/language/parser"
	"github.com/hntrl/lang/language/tokens"
)

// ObjectPattern :: LCURLY PropertyList RCURLY
type ObjectPattern struct {
	pos        tokens.Position
	Properties PropertyList
}

func (o ObjectPattern) Validate() error {
	if err := o.Properties.Validate(); err != nil {
		return err
	}
	return nil
}

func (o ObjectPattern) Pos() tokens.Position {
	return o.pos
}

func ParseObjectPattern(p *parser.Parser) (*ObjectPattern, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	op := ObjectPattern{pos: pos}
	if tok != tokens.LCURLY {
		return nil, ExpectedError(pos, tokens.LCURLY, lit)
	}

	pl, err := ParsePropertyList(p)
	if err != nil {
		return nil, err
	}
	op.Properties = *pl

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.RCURLY {
		return nil, ExpectedError(pos, tokens.RCURLY, lit)
	}
	return &op, nil
}

// PropertyList :: (Property | SpreadElement) (COMMA PropertyList)?
type PropertyList []Node

func (p PropertyList) Validate() error {
	for _, prop := range p {
		if spread, ok := prop.(SpreadElement); ok {
			if err := spread.Validate(); err != nil {
				return err
			}
		} else if val, ok := prop.(Property); ok {
			if err := val.Validate(); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("parsing: %T not allowed in PropertyList", prop)
		}
	}
	return nil
}

func (p PropertyList) Pos() tokens.Position {
	return p[0].Pos()
}

func ParsePropertyList(p *parser.Parser) (*PropertyList, error) {
	pl := PropertyList{}
	for {
		pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		p.Unscan()
		if tok == tokens.RCURLY {
			break
		} else if tok == tokens.ELLIPSIS {
			se, err := ParseSpreadElement(p)
			if err != nil {
				return nil, err
			}
			pl = append(pl, *se)
		} else if tok == tokens.IDENT {
			prop, err := ParseProperty(p)
			if err != nil {
				return nil, err
			}
			pl = append(pl, *prop)
		} else {
			return nil, ExpectedError(pos, tokens.IDENT, lit)
		}
		_, tok, _ = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		if tok != tokens.COMMA {
			p.Unscan()
			break
		}
	}
	return &pl, nil
}

// Property :: IDENT COLON Expression
type Property struct {
	pos  tokens.Position
	Key  string
	Init Expression
}

func (p Property) Validate() error {
	if err := p.Init.Validate(); err != nil {
		return err
	}
	return nil
}

func (p Property) Pos() tokens.Position {
	return p.pos
}

func ParseProperty(p *parser.Parser) (*Property, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	prop := Property{pos: pos}
	if tok != tokens.IDENT {
		return nil, ExpectedError(pos, tokens.IDENT, lit)
	}
	prop.Key = lit

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.COLON {
		return nil, ExpectedError(pos, tokens.COLON, lit)
	}

	expr, err := ParseExpression(p)
	if err != nil {
		return nil, err
	}
	prop.Init = *expr

	return &prop, nil
}

// SpreadElement :: ELLIPSIS Expression
type SpreadElement struct {
	pos  tokens.Position
	Init Expression
}

func (s SpreadElement) Validate() error {
	return s.Init.Validate()
}

func (s SpreadElement) Pos() tokens.Position {
	return s.pos
}

func ParseSpreadElement(p *parser.Parser) (*SpreadElement, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	se := SpreadElement{pos: pos}
	if tok != tokens.ELLIPSIS {
		return nil, ExpectedError(pos, tokens.ELLIPSIS, lit)
	}

	expr, err := ParseExpression(p)
	if err != nil {
		return nil, err
	}
	se.Init = *expr

	return &se, nil
}

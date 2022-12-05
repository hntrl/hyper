package nodes

import (
	"fmt"
	"strings"

	"github.com/hntrl/lang/language/parser"
	"github.com/hntrl/lang/language/tokens"
)

// Context :: COMMENT? CONTEXT Selector LCURLY (ContextObject | ContextObjectMethod | ContextMethod)* RCURLY
type Context struct {
	pos     tokens.Position
	Name    string
	Objects []Node `types:"ContextObject,ContextObjectMethod,ContextMethod"`
	Comment string
}

func (c Context) Validate() error {
	for _, obj := range c.Objects {
		if ctxObj, ok := obj.(ContextObject); ok {
			if err := ctxObj.Validate(); err != nil {
				return err
			}
		} else if objMethod, ok := obj.(ContextObjectMethod); ok {
			if err := objMethod.Validate(); err != nil {
				return err
			}
		} else if method, ok := obj.(ContextMethod); ok {
			if err := method.Validate(); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("parsing: %T not allowed in Context", obj)
		}
	}
	return nil
}

func (c Context) Pos() tokens.Position {
	return c.pos
}

func ParseContext(p *parser.Parser) (*Context, error) {
	context := Context{Objects: make([]Node, 0)}

	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE)
	if tok == tokens.COMMENT {
		context.Comment = lit
		pos, tok, lit = p.ScanIgnore(tokens.NEWLINE)
	}
	context.pos = pos
	if tok != tokens.CONTEXT {
		return nil, ExpectedError(pos, tokens.CONTEXT, lit)
	}

	selector, err := ParseSelector(p)
	if err != nil {
		return nil, err
	}
	context.Name = strings.Join(selector.Members, ".")

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE)
	if tok != tokens.LCURLY {
		return nil, ExpectedError(pos, tokens.LCURLY, lit)
	}

	for {
		_, tok, _ = p.ScanIgnore(tokens.NEWLINE)
		if tok == tokens.RCURLY {
			break
		}
		startIndex := p.Index()
		if tok == tokens.COMMENT {
			_, tok, _ = p.Scan()
			if tok != tokens.IDENT && tok != tokens.PRIVATE && tok != tokens.RCURLY {
				continue
			}
		}
		p.Unscan()
		_, tok, _ = p.ScanIgnore(tokens.NEWLINE)
		if tok == tokens.FUNC {
			p.Unscan()
			method, err := ParseContextObjectMethod(p)
			if err != nil {
				return nil, err
			}
			context.Objects = append(context.Objects, *method)
		} else {
			if tok == tokens.PRIVATE {
				p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
			}
			p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
			_, tok, _ = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
			p.Rollback(startIndex - 1)
			if tok == tokens.LPAREN {
				method, err := ParseContextMethod(p)
				if err != nil {
					return nil, err
				}
				context.Objects = append(context.Objects, *method)
			} else {
				obj, err := ParseContextObject(p)
				if err != nil {
					return nil, err
				}
				context.Objects = append(context.Objects, *obj)
			}
		}
	}
	return &context, nil
}

// ContextObject :: COMMENT? PRIVATE? IDENT IDENT (EXTENDS Selector)? LCURLY FieldStatement* RCURLY
type ContextObject struct {
	pos     tokens.Position
	Private bool
	Class   string
	Name    string
	Extends *Selector
	Fields  []FieldStatement
	Comment string
}

func (c ContextObject) Validate() error {
	if c.Extends != nil {
		if err := c.Extends.Validate(); err != nil {
			return err
		}
	}
	for _, field := range c.Fields {
		if err := field.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c ContextObject) Pos() tokens.Position {
	return c.pos
}

func ParseContextObject(p *parser.Parser) (*ContextObject, error) {
	obj := ContextObject{Fields: make([]FieldStatement, 0)}

	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE)
	if tok == tokens.COMMENT {
		obj.Comment = lit
		pos, tok, lit = p.ScanIgnore(tokens.NEWLINE)
	}
	if tok == tokens.PRIVATE {
		obj.Private = true
		pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	}
	if tok != tokens.IDENT {
		return nil, ExpectedError(pos, tokens.IDENT, lit)
	}
	obj.Class = lit
	obj.pos = pos
	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.IDENT {
		return nil, ExpectedError(pos, tokens.IDENT, lit)
	}
	obj.Name = lit

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok == tokens.EXTENDS {
		selector, err := ParseSelector(p)
		if err != nil {
			return nil, err
		}
		obj.Extends = selector
		pos, tok, lit = p.ScanIgnore(tokens.NEWLINE)
	}

	if tok != tokens.LCURLY {
		return nil, ExpectedError(pos, tokens.LCURLY, lit)
	}

	for {
		_, tok, _ = p.ScanIgnore(tokens.NEWLINE)
		if tok == tokens.RCURLY {
			break
		}
		if tok == tokens.COMMENT {
			startIndex := p.Index()
			_, tok, _ = p.Scan()
			if tok == tokens.NEWLINE {
				continue
			} else if tok == tokens.RCURLY {
				break
			} else {
				p.Rollback(startIndex)
			}
		}
		p.Unscan()
		field, err := ParseFieldStatement(p)
		if err != nil {
			return nil, err
		}
		obj.Fields = append(obj.Fields, *field)
	}
	return &obj, nil
}

// ContextObjectMethod :: FUNC LPAREN IDENT RPAREN IDENT FunctionBlock
type ContextObjectMethod struct {
	pos    tokens.Position
	Target string
	Name   string
	Block  FunctionBlock
}

func (c ContextObjectMethod) Validate() error {
	if err := c.Block.Validate(); err != nil {
		return err
	}
	return nil
}

func (c ContextObjectMethod) Pos() tokens.Position {
	return c.pos
}

func ParseContextObjectMethod(p *parser.Parser) (*ContextObjectMethod, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE)
	if tok != tokens.FUNC {
		return nil, ExpectedError(pos, tokens.FUNC, lit)
	}
	method := ContextObjectMethod{pos: pos}

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE)
	if tok != tokens.LPAREN {
		return nil, ExpectedError(pos, tokens.LPAREN, lit)
	}

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE)
	if tok != tokens.IDENT {
		return nil, ExpectedError(pos, tokens.IDENT, lit)
	}
	method.Target = lit

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE)
	if tok != tokens.RPAREN {
		return nil, ExpectedError(pos, tokens.RPAREN, lit)
	}

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE)
	if tok != tokens.IDENT {
		return nil, ExpectedError(pos, tokens.IDENT, lit)
	}
	method.Name = lit

	block, err := ParseFunctionBlock(p)
	if err != nil {
		return nil, err
	}
	method.Block = *block
	return &method, nil
}

// ContextMethod :: COMMENT? PRIVATE? IDENT IDENT FunctionBlock
type ContextMethod struct {
	pos     tokens.Position
	Private bool
	Class   string
	Name    string
	Block   FunctionBlock
	Comment string
}

func (c ContextMethod) Validate() error {
	return c.Block.Validate()
}

func (c ContextMethod) Pos() tokens.Position {
	return c.pos
}

func ParseContextMethod(p *parser.Parser) (*ContextMethod, error) {
	method := ContextMethod{}

	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE)
	if tok == tokens.COMMENT {
		method.Comment = lit
		pos, tok, lit = p.ScanIgnore(tokens.NEWLINE)
	}
	if tok == tokens.PRIVATE {
		method.Private = true
		pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	}
	if tok != tokens.IDENT {
		return nil, ExpectedError(pos, tokens.IDENT, lit)
	}
	method.Class = lit
	method.pos = pos

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.IDENT {
		return nil, ExpectedError(pos, tokens.IDENT, lit)
	}
	method.Name = lit

	block, err := ParseFunctionBlock(p)
	if err != nil {
		return nil, err
	}
	method.Block = *block
	return &method, nil
}

package ast

import (
	"fmt"
	"strings"

	"github.com/hntrl/hyper/src/hyper/parser"
	"github.com/hntrl/hyper/src/hyper/tokens"
)

// Context :: COMMENT? CONTEXT Selector LCURLY (UseStatement | ContextItem)* RCURLY
type Context struct {
	pos     tokens.Position
	Name    string
	Remotes []UseStatement
	Items   []ContextItem
	Comment string
}

func (c Context) Validate() error {
	for _, obj := range c.Remotes {
		if err := obj.Validate(); err != nil {
			return err
		}
	}
	for _, obj := range c.Items {
		if err := obj.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c Context) Pos() tokens.Position {
	return c.pos
}

func ParseContext(p *parser.Parser) (*Context, error) {
	context := Context{Remotes: make([]UseStatement, 0), Items: make([]ContextItem, 0)}

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
		if tok == tokens.USE {
			p.Unscan()
			stmt, err := ParseUseStatement(p)
			if err != nil {
				return nil, err
			}
			context.Remotes = append(context.Remotes, *stmt)
		} else {
			p.Rollback(startIndex - 1)
			item, err := ParseContextItem(p)
			if err != nil {
				return nil, err
			}
			context.Items = append(context.Items, *item)
		}
	}
	return &context, nil
}

// ContextItemSet :: ContextItem*
type ContextItemSet struct {
	Items []ContextItem
}

func (is ContextItemSet) Validate() error {
	for _, obj := range is.Items {
		if err := obj.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (is ContextItemSet) Pos() tokens.Position {
	return is.Items[0].Pos()
}

func ParseContextItemSet(p *parser.Parser) (*ContextItemSet, error) {
	set := ContextItemSet{Items: make([]ContextItem, 0)}
	for {
		_, tok, _ := p.ScanIgnore(tokens.NEWLINE)
		if tok == tokens.EOF {
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
		if tok == tokens.USE {
			return nil, fmt.Errorf("use directive not used in top level context")
		} else {
			p.Rollback(startIndex - 1)
			item, err := ParseContextItem(p)
			if err != nil {
				return nil, err
			}
			set.Items = append(set.Items, *item)
		}
	}
	return &set, nil
}

// ContextItem :: (ContextObject | ContextObjectMethod | ContextMethod | FunctionExpression)
type ContextItem struct {
	Init Node `types:"ContextObject,ContextObjectMethod,ContextMethod,FunctionExpression"`
}

func (i ContextItem) Validate() error {
	if obj, ok := i.Init.(ContextObject); ok {
		if err := obj.Validate(); err != nil {
			return err
		}
	} else if method, ok := i.Init.(ContextObjectMethod); ok {
		if err := method.Validate(); err != nil {
			return err
		}
	} else if method, ok := i.Init.(ContextMethod); ok {
		if err := method.Validate(); err != nil {
			return err
		}
	} else if expr, ok := i.Init.(FunctionExpression); ok {
		if err := expr.Validate(); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("parsing %T not allowed in ContextItem", i.Init)
	}
	return nil
}

func (i ContextItem) Pos() tokens.Position {
	return i.Init.Pos()
}

func ParseContextItem(p *parser.Parser) (*ContextItem, error) {
	startIndex := p.Index()
	_, tok, _ := p.ScanIgnore(tokens.NEWLINE)
	if tok == tokens.FUNC {
		_, tok, _ = p.ScanIgnore(tokens.NEWLINE)
		p.Rollback(startIndex - 1)
		if tok == tokens.LPAREN {
			method, err := ParseContextObjectMethod(p)
			if err != nil {
				return nil, err
			}
			return &ContextItem{*method}, nil
		} else {
			fn, err := ParseFunctionExpression(p)
			if err != nil {
				return nil, err
			}
			return &ContextItem{*fn}, nil
		}
	} else {
		if tok == tokens.COMMENT {
			_, tok, _ = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		}
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
			return &ContextItem{*method}, nil
		} else {
			obj, err := ParseContextObject(p)
			if err != nil {
				return nil, err
			}
			return &ContextItem{*obj}, nil
		}
	}
}

// UseStatement :: USE STRING
type UseStatement struct {
	pos    tokens.Position
	Source string
}

func (u UseStatement) Validate() error {
	return nil
}
func (u UseStatement) Pos() tokens.Position {
	return u.pos
}

func ParseUseStatement(p *parser.Parser) (*UseStatement, error) {
	statement := &UseStatement{}
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.USE {
		return nil, ExpectedError(pos, tokens.USE, lit)
	}
	statement.pos = pos
	_, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.STRING {
		return nil, ExpectedError(pos, tokens.STRING, lit)
	}
	statement.Source = lit
	return statement, nil
}

// ContextObject :: COMMENT? PRIVATE? IDENT IDENT (EXTENDS Selector)? LCURLY FieldStatement* RCURLY
type ContextObject struct {
	pos       tokens.Position
	Private   bool
	Interface string
	Name      string
	Extends   *Selector
	Fields    []FieldStatement
	Comment   string
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
	obj.Interface = lit
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
	pos       tokens.Position
	Private   bool
	Interface string
	Name      string
	Block     FunctionBlock
	Comment   string
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
	method.Interface = lit
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

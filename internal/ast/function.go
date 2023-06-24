package ast

import (
	"fmt"

	"github.com/hntrl/hyper/internal/parser"
	"github.com/hntrl/hyper/internal/tokens"
)

// ArgumentList :: (ArgumentItem | ArgumentObject) (COMMA ArgumentList)?
type ArgumentList struct {
	pos   tokens.Position
	Items []Node `types:"ArgumentItem,ArgumentObject"`
}

func (a ArgumentList) Validate() error {
	for _, arg := range a.Items {
		if item, ok := arg.(ArgumentItem); ok {
			if err := item.Validate(); err != nil {
				return err
			}
		} else if obj, ok := arg.(ArgumentObject); ok {
			if err := obj.Validate(); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("parsing: %T not allowed in ArgumentList", arg)
		}
	}
	return nil
}

func (a ArgumentList) Pos() tokens.Position {
	return a.pos
}

func ParseArgumentList(p *parser.Parser) (*ArgumentList, error) {
	args := ArgumentList{Items: make([]Node, 0)}
	for {
		pos, tok, _ := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		args.pos = pos
		p.Unscan()
		if tok == tokens.LCURLY {
			obj, err := ParseArgumentObject(p)
			if err != nil {
				return nil, err
			}
			args.Items = append(args.Items, *obj)
		} else if tok == tokens.IDENT {
			arg, err := ParseArgumentItem(p)
			if err != nil {
				return nil, err
			}
			args.Items = append(args.Items, *arg)
		} else {
			break
		}
		_, tok, _ = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		if tok != tokens.COMMA {
			p.Unscan()
			break
		}
	}
	return &args, nil
}

// ArgumentItem :: IDENT COLON TypeExpression
type ArgumentItem struct {
	pos  tokens.Position
	Key  string
	Init TypeExpression
}

func (a ArgumentItem) Validate() error {
	if a.Key == "" {
		return fmt.Errorf("parsing: ArgumentItem.Key is empty")
	}
	if err := a.Init.Validate(); err != nil {
		return err
	}
	return nil
}

func (a ArgumentItem) Pos() tokens.Position {
	return a.pos
}

func ParseArgumentItem(p *parser.Parser) (*ArgumentItem, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.IDENT {
		return nil, ExpectedError(pos, tokens.IDENT, lit)
	}
	item := ArgumentItem{pos: pos, Key: lit}

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.COLON {
		return nil, ExpectedError(pos, tokens.COLON, lit)
	}

	expr, err := ParseTypeExpression(p)
	if err != nil {
		return nil, err
	}
	item.Init = *expr
	return &item, nil
}

// ArgumentObject :: LCURLY (ArgumentItem)? (ArgumentItem COMMA)* RCURLY
type ArgumentObject struct {
	pos   tokens.Position
	Items []ArgumentItem
}

func (a ArgumentObject) Validate() error {
	for _, item := range a.Items {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (a ArgumentObject) Pos() tokens.Position {
	return a.pos
}

func ParseArgumentObject(p *parser.Parser) (*ArgumentObject, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.LCURLY {
		return nil, ExpectedError(pos, tokens.LCURLY, lit)
	}
	_, tok, _ = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok == tokens.RCURLY {
		return &ArgumentObject{pos: pos, Items: make([]ArgumentItem, 0)}, nil
	}

	p.Unscan()
	obj := ArgumentObject{pos: pos, Items: make([]ArgumentItem, 0)}
	for {
		_, tok, _ := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		if tok == tokens.RCURLY {
			break
		}
		p.Unscan()
		item, err := ParseArgumentItem(p)
		if err != nil {
			return nil, err
		}
		obj.Items = append(obj.Items, *item)

		_, tok, _ = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		if tok != tokens.COMMA {
			p.Unscan()
			break
		}
	}
	_, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.RCURLY {
		return nil, ExpectedError(pos, tokens.RCURLY, lit)
	}
	return &obj, nil
}

// FunctionParameters :: LPAREN ArgumentList? RPAREN TypeExpression?
type FunctionParameters struct {
	pos        tokens.Position
	Arguments  ArgumentList
	ReturnType *TypeExpression
}

func (p FunctionParameters) Validate() error {
	if err := p.Arguments.Validate(); err != nil {
		return err
	}
	if p.ReturnType != nil {
		if err := p.ReturnType.Validate(); err != nil {
			return err
		}
	}
	return nil
}
func (p FunctionParameters) Pos() tokens.Position {
	return p.pos
}

func ParseFunctionParameters(p *parser.Parser) (*FunctionParameters, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.LPAREN {
		return nil, ExpectedError(pos, tokens.LPAREN, lit)
	}
	args, err := ParseArgumentList(p)
	if err != nil {
		return nil, err
	}
	params := FunctionParameters{pos: pos, Arguments: *args}

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.RPAREN {
		return nil, ExpectedError(pos, tokens.RPAREN, lit)
	}

	_, tok, _ = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	p.Unscan()
	if tok == tokens.IDENT || tok == tokens.LSQUARE {
		ret, err := ParseTypeExpression(p)
		if err != nil {
			return nil, err
		}
		params.ReturnType = ret
	}
	return &params, nil
}

// FunctionBlock :: FunctionParameters LCURLY Block RCURLY
type FunctionBlock struct {
	Parameters FunctionParameters
	Body       Block
}

func (f FunctionBlock) Validate() error {
	if err := f.Parameters.Validate(); err != nil {
		return err
	}
	if err := f.Body.Validate(); err != nil {
		return err
	}
	return nil
}

func (f FunctionBlock) Pos() tokens.Position {
	return f.Parameters.pos
}

func ParseFunctionBlock(p *parser.Parser) (*FunctionBlock, error) {
	params, err := ParseFunctionParameters(p)
	if err != nil {
		return nil, err
	}
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.LCURLY {
		return nil, ExpectedError(pos, tokens.LCURLY, lit)
	}
	block, err := ParseBlock(p)
	if err != nil {
		return nil, err
	}
	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.RCURLY {
		return nil, ExpectedError(pos, tokens.RCURLY, lit)
	}
	return &FunctionBlock{Parameters: *params, Body: *block}, nil
}

// FunctionExpression :: FUNC IDENT FunctionBlock
type FunctionExpression struct {
	pos  tokens.Position
	Name string
	Body FunctionBlock
}

func (f FunctionExpression) Validate() error {
	if err := f.Body.Validate(); err != nil {
		return err
	}
	return nil
}

func (f FunctionExpression) Pos() tokens.Position {
	return f.pos
}

func ParseFunctionExpression(p *parser.Parser) (*FunctionExpression, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.FUNC {
		return nil, ExpectedError(pos, tokens.FUNC, lit)
	}
	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.IDENT {
		return nil, ExpectedError(pos, tokens.IDENT, lit)
	}
	block, err := ParseFunctionBlock(p)
	if err != nil {
		return nil, err
	}
	return &FunctionExpression{pos: pos, Name: lit, Body: *block}, nil
}

// Block :: BlockStatement*
type Block struct {
	pos        tokens.Position
	Statements []BlockStatement
}

func (b Block) Validate() error {
	for _, stmt := range b.Statements {
		if err := stmt.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (b Block) Pos() tokens.Position {
	return b.pos
}

func ParseBlock(p *parser.Parser) (*Block, error) {
	block := Block{Statements: make([]BlockStatement, 0)}
	for {
		pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		block.pos = pos
		switch tok {
		case tokens.RCURLY, tokens.CASE, tokens.DEFAULT:
			p.Unscan()
			return &block, nil
		case tokens.IDENT:
			p.Unscan()
			stmt, err := ParseBlockStatement(p)
			if err != nil {
				return nil, err
			}
			block.Statements = append(block.Statements, *stmt)
		default:
			if tok.IsKeyword() {
				p.Unscan()
				stmt, err := ParseBlockStatement(p)
				if err != nil {
					return nil, err
				}
				block.Statements = append(block.Statements, *stmt)
			} else {
				return nil, ExpectedError(pos, tokens.IDENT, lit)
			}
		}
	}
}

// InlineBlock :: (BlockStatement | LCURLY Block RCURLY)
type InlineBlock struct {
	pos  tokens.Position
	Body Block
}

func (b InlineBlock) Validate() error {
	if err := b.Body.Validate(); err != nil {
		return err
	}
	return nil
}

func (b InlineBlock) Pos() tokens.Position {
	return b.pos
}

func ParseInlineBlock(p *parser.Parser) (*InlineBlock, error) {
	firstPos, tok, _ := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok == tokens.LCURLY {
		block, err := ParseBlock(p)
		if err != nil {
			return nil, err
		}
		pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		if tok != tokens.RCURLY {
			return nil, ExpectedError(pos, tokens.RCURLY, lit)
		}
		return &InlineBlock{pos: firstPos, Body: *block}, nil
	} else {
		p.Unscan()
		stmt, err := ParseBlockStatement(p)
		if err != nil {
			return nil, err
		}
		block := Block{Statements: []BlockStatement{*stmt}}
		return &InlineBlock{pos: firstPos, Body: block}, nil
	}
}

// BlockStatement :: Expression
//
//		| DeclarationStatement
//		| AssignmentStatement
//		| IfStatement
//		| WhileStatement
//		| ForStatement
//		| ContinueStatement
//		| BreakStatement
//		| SwitchBlock
//		| GuardStatement
//		| ReturnStatement
//		| ThrowStatement
//	  | TryStatement
type BlockStatement struct {
	Init Node `types:"Expression,DeclarationStatement,AssignmentStatement,IfStatement,WhileStatement,ForStatement,ContinueStatement,BreakStatement,SwitchBlock,GuardStatement,ReturnStatement,ThrowStatement,TryStatement"`
}

func (b BlockStatement) Validate() error {
	if expr, ok := b.Init.(Expression); ok {
		if err := expr.Validate(); err != nil {
			return err
		}
	} else if decl, ok := b.Init.(DeclarationStatement); ok {
		if err := decl.Validate(); err != nil {
			return err
		}
	} else if assign, ok := b.Init.(AssignmentStatement); ok {
		if err := assign.Validate(); err != nil {
			return err
		}
	} else if ifstmt, ok := b.Init.(IfStatement); ok {
		if err := ifstmt.Validate(); err != nil {
			return err
		}
	} else if whilestmt, ok := b.Init.(WhileStatement); ok {
		if err := whilestmt.Validate(); err != nil {
			return err
		}
	} else if forstmt, ok := b.Init.(ForStatement); ok {
		if err := forstmt.Validate(); err != nil {
			return err
		}
	} else if contstmt, ok := b.Init.(ContinueStatement); ok {
		if err := contstmt.Validate(); err != nil {
			return err
		}
	} else if breakstmt, ok := b.Init.(BreakStatement); ok {
		if err := breakstmt.Validate(); err != nil {
			return err
		}
	} else if switchblock, ok := b.Init.(SwitchBlock); ok {
		if err := switchblock.Validate(); err != nil {
			return err
		}
	} else if guardstmt, ok := b.Init.(GuardStatement); ok {
		if err := guardstmt.Validate(); err != nil {
			return err
		}
	} else if ret, ok := b.Init.(ReturnStatement); ok {
		if err := ret.Validate(); err != nil {
			return err
		}
	} else if throw, ok := b.Init.(ThrowStatement); ok {
		if err := throw.Validate(); err != nil {
			return err
		}
	} else if try, ok := b.Init.(TryStatement); ok {
		if err := try.Validate(); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("parsing: %T not allowed in BlockStatement", b.Init)
	}
	return nil
}

func (b BlockStatement) Pos() tokens.Position {
	return b.Init.Pos()
}

func ParseBlockStatement(p *parser.Parser) (*BlockStatement, error) {
	stmt := BlockStatement{}

	startIndex := p.Index()
	_, tok, _ := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)

	switch tok {
	case tokens.IF:
		p.Unscan()
		ifstmt, err := ParseIfStatement(p)
		if err != nil {
			return nil, err
		}
		stmt.Init = *ifstmt
	case tokens.WHILE:
		p.Unscan()
		whilestmt, err := ParseWhileStatement(p)
		if err != nil {
			return nil, err
		}
		stmt.Init = *whilestmt
	case tokens.FOR:
		p.Unscan()
		forstmt, err := ParseForStatement(p)
		if err != nil {
			return nil, err
		}
		stmt.Init = *forstmt
	case tokens.CONTINUE:
		p.Unscan()
		contstmt, err := ParseContinueStatement(p)
		if err != nil {
			return nil, err
		}
		stmt.Init = *contstmt
	case tokens.BREAK:
		p.Unscan()
		breakstmt, err := ParseBreakStatement(p)
		if err != nil {
			return nil, err
		}
		stmt.Init = *breakstmt
	case tokens.SWITCH:
		p.Unscan()
		switchblock, err := ParseSwitchBlock(p)
		if err != nil {
			return nil, err
		}
		stmt.Init = *switchblock
	case tokens.GUARD:
		p.Unscan()
		guard, err := ParseGuardStatement(p)
		if err != nil {
			return nil, err
		}
		stmt.Init = *guard
	case tokens.RETURN:
		p.Unscan()
		ret, err := ParseReturnStatement(p)
		if err != nil {
			return nil, err
		}
		stmt.Init = *ret
	case tokens.THROW:
		p.Unscan()
		throw, err := ParseThrowStatement(p)
		if err != nil {
			return nil, err
		}
		stmt.Init = *throw
	case tokens.TRY:
		p.Unscan()
		try, err := ParseTryStatement(p)
		if err != nil {
			return nil, err
		}
		stmt.Init = *try
	default:
		_, tok, _ := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		if tok == tokens.COMMA {
			p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
			_, tok, _ = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		}
		p.Rollback(startIndex)
		if tok == tokens.DEFINE {
			decl, err := ParseDeclarationStatement(p)
			if err != nil {
				return nil, err
			}
			stmt.Init = *decl
		} else {
			p.Unscan()
			for {
				_, tok, _ := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
				if tok == tokens.PERIOD || tok == tokens.IDENT {
					continue
				}
				if tok == tokens.LSQUARE {
					// parse index expression and ignore it; still just trying to figure out if this is an assignment
					p.Unscan()
					_, err := ParseIndexExpression(p)
					if err != nil {
						return nil, err
					}
					continue
				}
				p.Rollback(startIndex)
				if tok == tokens.INC || tok == tokens.DEC || tok.IsAssignmentOperator() || tok == tokens.COMMA {
					assign, err := ParseAssignmentStatement(p)
					if err != nil {
						return nil, err
					}
					stmt.Init = *assign
				} else {
					expr, err := ParseExpression(p)
					if err != nil {
						return nil, err
					}
					stmt.Init = *expr
				}
				break
			}
		}
	}
	return &stmt, nil
}

// DeclarationStatement :: IDENT (COMMA IDENT)? DEFINE (Expression | TryStatement)
type DeclarationStatement struct {
	pos             tokens.Position
	Target          string
	SecondaryTarget *string
	Init            Node `types:"Expression,TryStatement"`
}

func (d DeclarationStatement) Validate() error {
	if d.Target == "" {
		return fmt.Errorf("parsing: empty target in DeclarationStatement")
	}
	if expr, ok := d.Init.(Expression); ok {
		if err := expr.Validate(); err != nil {
			return err
		}
	} else if try, ok := d.Init.(TryStatement); ok {
		if err := try.Validate(); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("parsing: %T not allowed in DeclarationStatement", d.Init)
	}
	return nil
}

func (d DeclarationStatement) Pos() tokens.Position {
	return d.pos
}

func ParseDeclarationStatement(p *parser.Parser) (*DeclarationStatement, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.IDENT {
		return nil, ExpectedError(pos, tokens.IDENT, lit)
	}
	stmt := DeclarationStatement{pos: pos, Target: lit}

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok == tokens.COMMA {
		pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		if tok != tokens.IDENT {
			return nil, ExpectedError(pos, tokens.IDENT, lit)
		}
		stmt.SecondaryTarget = &lit
		pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	}
	if tok != tokens.DEFINE {
		return nil, ExpectedError(pos, tokens.DEFINE, lit)
	}

	_, tok, _ = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	p.Unscan()
	if tok == tokens.TRY {
		try, err := ParseTryStatement(p)
		if err != nil {
			return nil, err
		}
		stmt.Init = *try
	} else {
		expr, err := ParseExpression(p)
		if err != nil {
			return nil, err
		}
		stmt.Init = *expr
	}
	return &stmt, nil
}

// AssignmentStatement :: AssignmentTargetExpression token(IsAssignmentOperator) Expression
//
//	| AssignmentTargetExpression (INC | DEC)
type AssignmentStatement struct {
	pos             tokens.Position
	Target          AssignmentTargetExpression
	SecondaryTarget *string
	Operator        tokens.Token
	Init            Node `types:"Expression,TryStatement"`
}

func (a AssignmentStatement) Validate() error {
	if err := a.Target.Validate(); err != nil {
		return err
	}
	if !a.Operator.IsAssignmentOperator() {
		return fmt.Errorf("AssignmentStatement has invalid operator: %s", a.Operator)
	}
	if expr, ok := a.Init.(Expression); ok {
		if err := expr.Validate(); err != nil {
			return err
		}
	} else if try, ok := a.Init.(TryStatement); ok {
		if err := try.Validate(); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("parsing: %T not allowed in AssignmentStatement", a.Init)
	}
	return nil
}

func (a AssignmentStatement) Pos() tokens.Position {
	return a.pos
}

func ParseAssignmentStatement(p *parser.Parser) (*AssignmentStatement, error) {
	target, err := ParseAssignmentTargetExpression(p)
	if err != nil {
		return nil, err
	}
	assign := AssignmentStatement{pos: target.pos, Target: *target}
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok == tokens.COMMA {
		pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		if tok != tokens.IDENT {
			return nil, ExpectedError(pos, tokens.IDENT, lit)
		}
		assign.SecondaryTarget = &lit
		pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	} else {
		if tok == tokens.INC || tok == tokens.DEC {
			if tok == tokens.INC {
				assign.Operator = tokens.ADD
			} else {
				assign.Operator = tokens.SUB
			}
			assign.Init = Expression{pos, Literal{pos, int64(1)}}
			return &assign, nil
		}
	}
	if !tok.IsAssignmentOperator() {
		return nil, ExpectedError(pos, tokens.ILLEGAL, lit)
	}
	assign.Operator = tok

	p.Unscan()
	if tok == tokens.TRY {
		try, err := ParseTryStatement(p)
		if err != nil {
			return nil, err
		}
		assign.Init = *try
	} else {
		expr, err := ParseExpression(p)
		if err != nil {
			return nil, err
		}
		assign.Init = *expr
	}
	return &assign, nil
}

// AssignmentTargetExpression :: IDENT AssignmentTargetExpressionMember*
type AssignmentTargetExpression struct {
	pos     tokens.Position
	Members []AssignmentTargetExpressionMember
}

func (v AssignmentTargetExpression) Validate() error {
	for _, m := range v.Members {
		if err := m.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (v AssignmentTargetExpression) Pos() tokens.Position {
	return v.pos
}

func ParseAssignmentTargetExpression(p *parser.Parser) (*AssignmentTargetExpression, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	ve := AssignmentTargetExpression{pos: pos, Members: []AssignmentTargetExpressionMember{}}

	if tok != tokens.IDENT {
		return nil, ExpectedError(pos, tokens.IDENT, lit)
	}
	ve.Members = append(ve.Members, AssignmentTargetExpressionMember{Init: lit})

	for {
		_, tok, _ := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		p.Unscan()
		if tok != tokens.PERIOD && tok != tokens.LSQUARE && tok != tokens.LPAREN {
			break
		}
		member, err := ParseAssignmentTargetExpressionMember(p)
		if err != nil {
			return nil, err
		}
		ve.Members = append(ve.Members, *member)
	}
	return &ve, nil
}

// AssignmentTargetExpressionMember :: PERIOD IDENT
//
//	| IndexExpression
type AssignmentTargetExpressionMember struct {
	pos  tokens.Position
	Init interface{} `types:"string,IndexExpression"`
}

func (v AssignmentTargetExpressionMember) Validate() error {
	if _, ok := v.Init.(string); ok {
		// do nothing
	} else if index, ok := v.Init.(IndexExpression); ok {
		if err := index.Validate(); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("parsing: %T not allowed in AssignmentTargetExpression", v.Init)
	}
	return nil
}

func (v AssignmentTargetExpressionMember) Pos() tokens.Position {
	return v.pos
}

func ParseAssignmentTargetExpressionMember(p *parser.Parser) (*AssignmentTargetExpressionMember, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	member := AssignmentTargetExpressionMember{pos: pos}

	switch tok {
	case tokens.PERIOD:
		pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		if tok != tokens.IDENT {
			return nil, ExpectedError(pos, tokens.IDENT, lit)
		}
		member.Init = lit
	case tokens.LSQUARE:
		p.Unscan()
		index, err := ParseIndexExpression(p)
		if err != nil {
			return nil, err
		}
		member.Init = *index
	default:
		return nil, ExpectedError(pos, tokens.PERIOD, lit)
	}
	return &member, nil
}

// IfStatement :: IF LPAREN Expression RPAREN InlineBlock (ELSE IfStatement)? (ELSE Block)?
type IfStatement struct {
	pos       tokens.Position
	Condition Expression
	Body      Block
	Alternate Node `types:"IfStatement,Block"`
}

func (i IfStatement) Validate() error {
	if err := i.Condition.Validate(); err != nil {
		return err
	}
	if err := i.Body.Validate(); err != nil {
		return err
	}
	if i.Alternate != nil {
		if alt, ok := i.Alternate.(IfStatement); ok {
			if err := alt.Validate(); err != nil {
				return err
			}
		} else if block, ok := i.Alternate.(Block); ok {
			if err := block.Validate(); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("parsing: %T not allowed in IfStatement", i.Alternate)
		}
	}
	return nil
}

func (i IfStatement) Pos() tokens.Position {
	return i.pos
}

func ParseIfStatement(p *parser.Parser) (*IfStatement, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.IF {
		return nil, ExpectedError(pos, tokens.IF, lit)
	}
	stmt := IfStatement{pos: pos}

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.LPAREN {
		return nil, ExpectedError(pos, tokens.LPAREN, lit)
	}

	expr, err := ParseExpression(p)
	if err != nil {
		return nil, err
	}
	stmt.Condition = *expr

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.RPAREN {
		return nil, ExpectedError(pos, tokens.RPAREN, lit)
	}

	block, err := ParseInlineBlock(p)
	if err != nil {
		return nil, err
	}
	stmt.Body = Block(block.Body)

	_, tok, _ = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok == tokens.ELSE {
		_, tok, _ = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		if tok == tokens.IF {
			p.Unscan()
			ifstmt, err := ParseIfStatement(p)
			if err != nil {
				return nil, err
			}
			stmt.Alternate = *ifstmt
		} else {
			p.Unscan()
			block, err := ParseInlineBlock(p)
			if err != nil {
				return nil, err
			}
			stmt.Alternate = block.Body
		}
	} else {
		p.Unscan()
	}
	return &stmt, nil
}

// WhileStatement :: WHILE LPAREN Expression RPAREN InlineBlock
type WhileStatement struct {
	pos       tokens.Position
	Condition Expression
	Body      Block
}

func (w WhileStatement) Validate() error {
	if err := w.Condition.Validate(); err != nil {
		return err
	}
	if err := w.Body.Validate(); err != nil {
		return err
	}
	return nil
}

func (w WhileStatement) Pos() tokens.Position {
	return w.pos
}

func ParseWhileStatement(p *parser.Parser) (*WhileStatement, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.WHILE {
		return nil, ExpectedError(pos, tokens.WHILE, lit)
	}
	stmt := WhileStatement{pos: pos}

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.LPAREN {
		return nil, ExpectedError(pos, tokens.LPAREN, lit)
	}

	expr, err := ParseExpression(p)
	if err != nil {
		return nil, err
	}
	stmt.Condition = *expr

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.RPAREN {
		return nil, ExpectedError(pos, tokens.RPAREN, lit)
	}

	block, err := ParseInlineBlock(p)
	if err != nil {
		return nil, err
	}
	stmt.Body = block.Body

	return &stmt, nil
}

// ForStatement :: FOR LPAREN (ForCondition | RangeCondition) RPAREN InlineBlock
type ForStatement struct {
	pos       tokens.Position
	Condition Node `types:"ForCondition,RangeCondition"`
	Body      Block
}

func (f ForStatement) Validate() error {
	if forCond, ok := f.Condition.(ForCondition); ok {
		if err := forCond.Validate(); err != nil {
			return err
		}
	} else if rangeCond, ok := f.Condition.(RangeCondition); ok {
		if err := rangeCond.Validate(); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("parsing: %T not allowed in ForStatement", f.Condition)
	}
	if err := f.Body.Validate(); err != nil {
		return err
	}
	return nil
}

func (f ForStatement) Pos() tokens.Position {
	return f.pos
}

func ParseForStatement(p *parser.Parser) (*ForStatement, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.FOR {
		return nil, ExpectedError(pos, tokens.FOR, lit)
	}
	stmt := ForStatement{pos: pos}

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.LPAREN {
		return nil, ExpectedError(pos, tokens.LPAREN, lit)
	}

	startIndex := p.Index()
	for {
		_, tok, _ = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		if tok == tokens.SEMICOLON {
			p.Rollback(startIndex)
			forCond, err := ParseForCondition(p)
			if err != nil {
				return nil, err
			}
			stmt.Condition = *forCond
			break
		}
		if tok == tokens.RPAREN {
			p.Rollback(startIndex)
			rangeCond, err := ParseRangeCondition(p)
			if err != nil {
				return nil, err
			}
			stmt.Condition = *rangeCond
			break
		}
	}

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.RPAREN {
		return nil, ExpectedError(pos, tokens.RPAREN, lit)
	}

	block, err := ParseInlineBlock(p)
	if err != nil {
		return nil, err
	}
	stmt.Body = block.Body

	return &stmt, nil
}

// ForCondition :: (DeclarationStatement | Expression) SEMICOLON Expression (SEMICOLON (Expression | AssignmentStatement))?
type ForCondition struct {
	pos       tokens.Position
	Init      *DeclarationStatement
	Condition Expression
	Update    Node `types:"Expression,AssignmentStatement"`
}

func (f ForCondition) Validate() error {
	if err := f.Init.Validate(); err != nil {
		return err
	}
	if err := f.Condition.Validate(); err != nil {
		return err
	}
	if expr, ok := f.Update.(Expression); ok {
		if err := expr.Validate(); err != nil {
			return err
		}
	} else if assign, ok := f.Update.(AssignmentStatement); ok {
		if err := assign.Validate(); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("parsing: %T not allowed in ForCondition", f.Update)
	}
	return nil
}

func (f ForCondition) Pos() tokens.Position {
	return f.pos
}

func ParseForCondition(p *parser.Parser) (*ForCondition, error) {
	startIndex := p.Index()
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	stmt := ForCondition{pos: pos}

	if tok == tokens.IDENT {
		_, tok, _ = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		p.Rollback(startIndex)
		if tok == tokens.DEFINE {
			decl, err := ParseDeclarationStatement(p)
			if err != nil {
				return nil, err
			}
			stmt.Init = decl
		}
	} else {
		return nil, ExpectedError(pos, tokens.IDENT, lit)
	}

	if stmt.Init != nil {
		pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		if tok != tokens.SEMICOLON {
			return nil, ExpectedError(pos, tokens.SEMICOLON, lit)
		}
	}

	expr, err := ParseExpression(p)
	if err != nil {
		return nil, err
	}
	stmt.Condition = *expr

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.SEMICOLON {
		return nil, ExpectedError(pos, tokens.SEMICOLON, lit)
	}

	startIndex = p.Index()
	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok == tokens.IDENT {
		_, tok, _ = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		p.Rollback(startIndex)
		if tok.IsAssignmentOperator() {
			assign, err := ParseAssignmentStatement(p)
			if err != nil {
				return nil, err
			}
			stmt.Update = *assign
		} else {
			expr, err := ParseExpression(p)
			if err != nil {
				return nil, err
			}
			stmt.Update = *expr
		}
	} else {
		return nil, ExpectedError(pos, tokens.IDENT, lit)
	}
	return &stmt, nil
}

// RangeCondition :: IDENT COMMA IDENT IN Expression
type RangeCondition struct {
	pos    tokens.Position
	Index  string
	Value  string
	Target Expression
}

func (r RangeCondition) Validate() error {
	if err := r.Target.Validate(); err != nil {
		return err
	}
	return nil
}

func (r RangeCondition) Pos() tokens.Position {
	return r.pos
}

func ParseRangeCondition(p *parser.Parser) (*RangeCondition, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	stmt := RangeCondition{pos: pos}
	if tok != tokens.IDENT {
		return nil, ExpectedError(pos, tokens.IDENT, lit)
	}
	stmt.Index = lit

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.COMMA {
		return nil, ExpectedError(pos, tokens.COMMA, lit)
	}

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.IDENT {
		return nil, ExpectedError(pos, tokens.IDENT, lit)
	}
	stmt.Value = lit

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.IN {
		return nil, ExpectedError(pos, tokens.IN, lit)
	}

	expr, err := ParseExpression(p)
	if err != nil {
		return nil, err
	}
	stmt.Target = *expr

	return &stmt, nil
}

// ContinueStatement :: CONTINUE
type ContinueStatement struct {
	pos tokens.Position
}

func (c ContinueStatement) Validate() error {
	return nil
}

func (c ContinueStatement) Pos() tokens.Position {
	return c.pos
}

func ParseContinueStatement(p *parser.Parser) (*ContinueStatement, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.CONTINUE {
		return nil, ExpectedError(pos, tokens.CONTINUE, lit)
	}
	return &ContinueStatement{pos: pos}, nil
}

// BreakStatement :: BREAK
type BreakStatement struct {
	pos tokens.Position
}

func (b BreakStatement) Validate() error {
	return nil
}

func (b BreakStatement) Pos() tokens.Position {
	return b.pos
}

func ParseBreakStatement(p *parser.Parser) (*BreakStatement, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.BREAK {
		return nil, ExpectedError(pos, tokens.BREAK, lit)
	}
	return &BreakStatement{pos: pos}, nil
}

// SwitchBlock :: SWITCH LPAREN Expression RPAREN LCURLY SwitchStatement* RCURLY
type SwitchBlock struct {
	pos        tokens.Position
	Target     Expression
	Statements []SwitchStatement
}

func (s SwitchBlock) Validate() error {
	if err := s.Target.Validate(); err != nil {
		return err
	}
	for _, stmt := range s.Statements {
		if err := stmt.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (s SwitchBlock) Pos() tokens.Position {
	return s.pos
}

func ParseSwitchBlock(p *parser.Parser) (*SwitchBlock, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	stmt := SwitchBlock{pos: pos, Statements: make([]SwitchStatement, 0)}
	if tok != tokens.SWITCH {
		return nil, ExpectedError(pos, tokens.SWITCH, lit)
	}

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.LPAREN {
		return nil, ExpectedError(pos, tokens.LPAREN, lit)
	}

	expr, err := ParseExpression(p)
	if err != nil {
		return nil, err
	}
	stmt.Target = *expr

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.RPAREN {
		return nil, ExpectedError(pos, tokens.RPAREN, lit)
	}

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.LCURLY {
		return nil, ExpectedError(pos, tokens.LCURLY, lit)
	}

	for {
		_, tok, _ = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		if tok == tokens.RCURLY {
			break
		}
		p.Unscan()

		switchStmt, err := ParseSwitchStatement(p)
		if err != nil {
			return nil, err
		}
		stmt.Statements = append(stmt.Statements, switchStmt)
	}

	return &stmt, nil
}

// SwitchStatement :: ((CASE Expression) | DEFAULT) COLON Block
type SwitchStatement struct {
	pos       tokens.Position
	Condition *Expression
	IsDefault bool
	Body      Block
}

func (s SwitchStatement) Validate() error {
	if s.Condition != nil {
		if err := s.Condition.Validate(); err != nil {
			return err
		}
	}
	if !s.IsDefault {
		if err := s.Condition.Validate(); err != nil {
			return err
		}
	}
	if err := s.Body.Validate(); err != nil {
		return err
	}
	return nil
}

func (s SwitchStatement) Pos() tokens.Position {
	return s.pos
}

func ParseSwitchStatement(p *parser.Parser) (SwitchStatement, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	stmt := SwitchStatement{pos: pos}
	if tok == tokens.DEFAULT {
		stmt.IsDefault = true
	} else if tok == tokens.CASE {
		expr, err := ParseExpression(p)
		if err != nil {
			return stmt, err
		}
		stmt.Condition = expr
	} else {
		return stmt, ExpectedError(pos, tokens.CASE, lit)
	}

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.COLON {
		return stmt, ExpectedError(pos, tokens.COLON, lit)
	}

	block, err := ParseBlock(p)
	if err != nil {
		return stmt, err
	}
	stmt.Body = *block

	return stmt, nil
}

// GuardStatement :: GUARD Expression
type GuardStatement struct {
	pos  tokens.Position
	Init Expression
}

func (g GuardStatement) Validate() error {
	return g.Init.Validate()
}

func (g GuardStatement) Pos() tokens.Position {
	return g.pos
}

func ParseGuardStatement(p *parser.Parser) (*GuardStatement, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.GUARD {
		return nil, ExpectedError(pos, tokens.GUARD, lit)
	}
	stmt := GuardStatement{pos: pos}

	expr, err := ParseExpression(p)
	if err != nil {
		return nil, err
	}
	stmt.Init = *expr

	return &stmt, nil
}

// ReturnStatement :: RETURN Expression
type ReturnStatement struct {
	pos  tokens.Position
	Init Expression
}

func (r ReturnStatement) Validate() error {
	return r.Init.Validate()
}

func (r ReturnStatement) Pos() tokens.Position {
	return r.pos
}

func ParseReturnStatement(p *parser.Parser) (*ReturnStatement, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	stmt := ReturnStatement{pos: pos}
	if tok != tokens.RETURN {
		return nil, ExpectedError(pos, tokens.RETURN, lit)
	}

	expr, err := ParseExpression(p)
	if err != nil {
		return nil, err
	}
	stmt.Init = *expr

	return &stmt, nil
}

// ThrowStatement :: THROW Expression
type ThrowStatement struct {
	pos  tokens.Position
	Init Expression
}

func (t ThrowStatement) Validate() error {
	return t.Init.Validate()
}

func (t ThrowStatement) Pos() tokens.Position {
	return t.pos
}

func ParseThrowStatement(p *parser.Parser) (*ThrowStatement, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	stmt := ThrowStatement{pos: pos}
	if tok != tokens.THROW {
		return nil, ExpectedError(pos, tokens.THROW, lit)
	}

	expr, err := ParseExpression(p)
	if err != nil {
		return nil, err
	}
	stmt.Init = *expr

	return &stmt, nil
}

// TryStatement :: TRY Expression
type TryStatement struct {
	pos  tokens.Position
	Init Expression
}

func (t TryStatement) Validate() error {
	return t.Init.Validate()
}

func (t TryStatement) Pos() tokens.Position {
	return t.pos
}

func ParseTryStatement(p *parser.Parser) (*TryStatement, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	stmt := TryStatement{pos: pos}
	if tok != tokens.TRY {
		return nil, ExpectedError(pos, tokens.TRY, lit)
	}
	expr, err := ParseExpression(p)
	if err != nil {
		return nil, err
	}
	stmt.Init = *expr
	return &stmt, nil
}

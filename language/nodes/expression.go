package nodes

import (
	"encoding/json"
	"fmt"

	"github.com/hntrl/lang/language/parser"
	"github.com/hntrl/lang/language/tokens"
)

// TypeExpression :: (LSQUARE RSQUARE)? Selector QUESTION?
//
//	| (LSQUARE RSQUARE)? PARTIAL LT Selector GT QUESTION?
type TypeExpression struct {
	pos        tokens.Position
	IsArray    bool
	IsPartial  bool
	IsOptional bool
	Selector   Selector
}

func (t TypeExpression) Validate() error {
	if err := t.Selector.Validate(); err != nil {
		return err
	}
	return nil
}

func (t TypeExpression) Pos() tokens.Position {
	return t.pos
}

func ParseTypeExpression(p *parser.Parser) (*TypeExpression, error) {
	pos, tok, _ := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	te := TypeExpression{pos: pos}

	if tok == tokens.LSQUARE {
		_, tok, _ = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		if tok != tokens.RSQUARE {
			return nil, ExpectedError(pos, tokens.RSQUARE, "")
		}
		te.IsArray = true
	} else {
		p.Unscan()
	}

	_, tok, _ = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok == tokens.PARTIAL {
		te.IsPartial = true
		_, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		if tok != tokens.LESS {
			return nil, ExpectedError(pos, tokens.LESS, lit)
		}
	} else {
		p.Rollback(p.Index() - 1)
	}

	sel, err := ParseSelector(p)
	if err != nil {
		return nil, err
	}
	te.Selector = *sel

	if te.IsPartial {
		_, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		if tok != tokens.GREATER {
			return nil, ExpectedError(pos, tokens.GREATER, lit)
		}
	}

	_, tok, _ = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok == tokens.QUESTION {
		te.IsOptional = true
	} else {
		p.Unscan()
	}
	return &te, nil
}

// Expression :: Literal
//
//	| ArrayExpression
//	| InstanceExpression
//	| UnaryExpression
//	| BinaryExpression
//	| ObjectPattern
//	| FunctionExpression
//	| ValueExpression
//	| LPAREN Expression RPAREN
type Expression struct {
	pos  tokens.Position
	Init Node `types:"Literal,ArrayExpression,InstanceExpression,UnaryExpression,BinaryExpression,ObjectPattern,FunctionExpression,ValueExpression,Expression"`
}

func (e Expression) Validate() error {
	if lit, ok := e.Init.(Literal); ok {
		if err := lit.Validate(); err != nil {
			return err
		}
	} else if arr, ok := e.Init.(ArrayExpression); ok {
		if err := arr.Validate(); err != nil {
			return err
		}
	} else if inst, ok := e.Init.(InstanceExpression); ok {
		if err := inst.Validate(); err != nil {
			return err
		}
	} else if un, ok := e.Init.(UnaryExpression); ok {
		if err := un.Validate(); err != nil {
			return err
		}
	} else if bin, ok := e.Init.(BinaryExpression); ok {
		if err := bin.Validate(); err != nil {
			return err
		}
	} else if obj, ok := e.Init.(ObjectPattern); ok {
		if err := obj.Validate(); err != nil {
			return err
		}
	} else if fn, ok := e.Init.(FunctionExpression); ok {
		if err := fn.Validate(); err != nil {
			return err
		}
	} else if val, ok := e.Init.(ValueExpression); ok {
		if err := val.Validate(); err != nil {
			return err
		}
	} else if expr, ok := e.Init.(Expression); ok {
		if err := expr.Validate(); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("parsing %T not allowed in Expression", e.Init)
	}
	return nil
}

func (e Expression) Pos() tokens.Position {
	return e.pos
}

func ParseExpression(p *parser.Parser) (*Expression, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	expr := Expression{pos: pos}

	switch tok {
	case tokens.INT, tokens.FLOAT, tokens.STRING:
		p.Unscan()
		literal, err := ParseLiteral(p)
		if err != nil {
			return nil, err
		}
		expr.Init = *literal
	case tokens.LSQUARE:
		p.Unscan()
		arr, err := ParseArrayExpression(p)
		if err != nil {
			return nil, err
		}
		expr.Init = *arr
	case tokens.ADD, tokens.SUB, tokens.NOT:
		p.Unscan()
		un, err := ParseUnaryExpression(p)
		if err != nil {
			return nil, err
		}
		expr.Init = *un
	case tokens.LCURLY:
		p.Unscan()
		obj, err := ParseObjectPattern(p)
		if err != nil {
			return nil, err
		}
		expr.Init = *obj
	case tokens.FUNC:
		p.Unscan()
		fn, err := ParseFunctionExpression(p)
		if err != nil {
			return nil, err
		}
		expr.Init = *fn
	case tokens.IDENT:
		switch lit {
		case "true":
			expr.Init = Literal{Value: true}
		case "false":
			expr.Init = Literal{Value: false}
		case "nil":
			expr.Init = Literal{Value: nil}
		default:
			startingIndex := p.Index()
		checkLoop:
			for {
				_, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
				switch tok {
				case tokens.LCURLY:
					p.Rollback(startingIndex - 1)
					inst, err := ParseInstanceExpression(p)
					if err != nil {
						return nil, err
					}
					expr.Init = *inst
					break checkLoop
				case tokens.PERIOD:
					pos, tok, _ := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
					if tok != tokens.IDENT {
						return nil, ExpectedError(pos, tokens.IDENT, lit)
					}
					continue
				default:
					p.Rollback(startingIndex - 1)
					val, err := ParseValueExpression(p)
					if err != nil {
						return nil, err
					}
					expr.Init = *val
					break checkLoop
				}
			}
		}
	case tokens.LPAREN:
		newExpr, err := ParseExpression(p)
		if err != nil {
			return nil, err
		}
		expr.Init = *newExpr
		pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		if tok != tokens.RPAREN {
			return nil, ExpectedError(pos, tokens.RPAREN, lit)
		}
	default:
		return nil, ExpectedError(pos, tokens.IDENT, lit)
	}

	_, tok, _ = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	p.Unscan()
	if tok.IsOperator() || tok.IsComparableOperator() {
		bin, err := ParseBinaryExpression(p, expr)
		if err != nil {
			return nil, err
		}
		expr.Init = *bin
		expr = orderBinaryExpression(expr)
	}
	return &expr, nil
}

// ArrayExpression :: LSQUARE RSQUARE TypeExpression LCURLY ((Expression COMMA) | Expression)* RCURLY
type ArrayExpression struct {
	pos      tokens.Position
	Init     TypeExpression
	Elements []Expression
}

func (a ArrayExpression) Validate() error {
	if err := a.Init.Validate(); err != nil {
		return err
	}
	for _, e := range a.Elements {
		if err := e.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (a ArrayExpression) Pos() tokens.Position {
	return a.pos
}

func ParseArrayExpression(p *parser.Parser) (*ArrayExpression, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	ae := ArrayExpression{pos: pos, Elements: make([]Expression, 0)}

	if tok != tokens.LSQUARE {
		return nil, ExpectedError(pos, tokens.LSQUARE, lit)
	}
	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.RSQUARE {
		return nil, ExpectedError(pos, tokens.RSQUARE, lit)
	}

	expr, err := ParseTypeExpression(p)
	if err != nil {
		return nil, err
	}
	ae.Init = *expr

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.LCURLY {
		return nil, ExpectedError(pos, tokens.LCURLY, lit)
	}

	for {
		_, tok, _ := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		if tok == tokens.RCURLY {
			break
		}
		p.Unscan()
		expr, err := ParseExpression(p)
		if err != nil {
			return nil, err
		}
		ae.Elements = append(ae.Elements, *expr)
		_, tok, _ = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		if tok != tokens.COMMA {
			p.Unscan()
		}
	}
	return &ae, nil
}

// InstanceExpression :: Selector LCURLY PropertyList RCURLY
type InstanceExpression struct {
	pos        tokens.Position
	Selector   Selector
	Properties PropertyList
}

func (i InstanceExpression) Validate() error {
	if err := i.Selector.Validate(); err != nil {
		return err
	}
	if err := i.Properties.Validate(); err != nil {
		return err
	}
	return nil
}

func (i InstanceExpression) Pos() tokens.Position {
	return i.pos
}

func ParseInstanceExpression(p *parser.Parser) (*InstanceExpression, error) {
	selector, err := ParseSelector(p)
	ie := InstanceExpression{pos: selector.pos}

	if err != nil {
		return nil, err
	}
	ie.Selector = *selector

	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.LCURLY {
		return nil, ExpectedError(pos, tokens.LCURLY, lit)
	}

	properties, err := ParsePropertyList(p)
	if err != nil {
		return nil, err
	}
	ie.Properties = *properties

	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.RCURLY {
		return nil, ExpectedError(pos, tokens.RCURLY, lit)
	}
	return &ie, nil
}

// UnaryExpression :: EXCLAIM Expression
//
//	| PLUS Expression
//	| MINUS Expression
type UnaryExpression struct {
	pos      tokens.Position
	Operator tokens.Token
	Init     Expression
}

func (u UnaryExpression) Validate() error {
	if u.Operator != tokens.ADD && u.Operator != tokens.SUB && u.Operator != tokens.NOT {
		return fmt.Errorf("parsing: invalid unary operator %s", u.Operator)
	}
	if err := u.Init.Validate(); err != nil {
		return err
	}
	return nil
}

func (u UnaryExpression) Pos() tokens.Position {
	return u.pos
}

func ParseUnaryExpression(p *parser.Parser) (*UnaryExpression, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	ue := UnaryExpression{pos: pos}

	switch tok {
	case tokens.ADD, tokens.SUB, tokens.NOT:
		ue.Operator = tok
	default:
		return nil, ExpectedError(pos, tokens.NOT, lit)
	}

	expr, err := ParseExpression(p)
	if err != nil {
		return nil, err
	}
	ue.Init = *expr
	return &ue, nil
}

// BinaryExpression :: Expression token(IsOperator) Expression
type BinaryExpression struct {
	pos      tokens.Position
	Left     Expression
	Operator tokens.Token
	Right    Expression
}

func (b BinaryExpression) Validate() error {
	if !b.Operator.IsOperator() && !b.Operator.IsComparableOperator() {
		return fmt.Errorf("parsing: invalid binary operator %s", b.Operator)
	}
	if err := b.Left.Validate(); err != nil {
		return err
	}
	if err := b.Right.Validate(); err != nil {
		return err
	}
	return nil
}

func (b BinaryExpression) Pos() tokens.Position {
	return b.pos
}

func (b BinaryExpression) MarshalText() ([]byte, error) {
	left, err := json.Marshal(b.Left)
	if err != nil {
		return nil, err
	}
	right, err := json.Marshal(b.Right)
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("BinEx(%s, %s, %s)", string(left), b.Operator.String(), string(right))), nil
}

// order a binary expression by operator precedence
func orderBinaryExpression(expr Expression) Expression {
	if bin, ok := expr.Init.(BinaryExpression); ok {
		if rightBin, ok := bin.Right.Init.(BinaryExpression); ok {
			if rightBin.Operator.Precedence() <= bin.Operator.Precedence() {
				return Expression{
					Init: BinaryExpression{
						pos: bin.pos,
						Left: orderBinaryExpression(Expression{
							pos: rightBin.pos,
							Init: BinaryExpression{
								Left:     bin.Left,
								Operator: bin.Operator,
								Right:    rightBin.Left,
							},
						}),
						Operator: rightBin.Operator,
						Right:    rightBin.Right,
					},
				}
			}
		}
	}
	return expr
}

func ParseBinaryExpression(p *parser.Parser, left Expression) (*BinaryExpression, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	be := BinaryExpression{pos: pos, Left: left}

	if !tok.IsOperator() && !tok.IsComparableOperator() {
		return nil, ExpectedError(pos, tokens.ADD, lit)
	}
	be.Operator = tok

	right, err := ParseExpression(p)
	if err != nil {
		return nil, err
	}
	be.Right = *right
	return &be, nil
}

// ValueExpression :: IDENT ValueExpressionMember*
type ValueExpression struct {
	pos     tokens.Position
	Members []ValueExpressionMember
}

func (v ValueExpression) Validate() error {
	for _, m := range v.Members {
		if err := m.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (v ValueExpression) Pos() tokens.Position {
	return v.pos
}

func ParseValueExpression(p *parser.Parser) (*ValueExpression, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	ve := ValueExpression{pos: pos, Members: []ValueExpressionMember{}}

	if tok != tokens.IDENT {
		return nil, ExpectedError(pos, tokens.IDENT, lit)
	}
	ve.Members = append(ve.Members, ValueExpressionMember{Init: lit})

	for {
		_, tok, _ := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		p.Unscan()
		if tok != tokens.PERIOD && tok != tokens.LSQUARE && tok != tokens.LPAREN {
			break
		}
		member, err := ParseValueExpressionMember(p)
		if err != nil {
			return nil, err
		}
		ve.Members = append(ve.Members, *member)
	}
	return &ve, nil
}

// ValueExpressionMember :: PERIOD IDENT
//
//	| CallExpression
//	| IndexExpression
type ValueExpressionMember struct {
	pos  tokens.Position
	Init interface{} `types:"string,CallExpression,IndexExpression"`
}

func (v ValueExpressionMember) Validate() error {
	if _, ok := v.Init.(string); ok {
		// do nothing
	} else if call, ok := v.Init.(CallExpression); ok {
		if err := call.Validate(); err != nil {
			return err
		}
	} else if index, ok := v.Init.(IndexExpression); ok {
		if err := index.Validate(); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("parsing: %T not allowed in ValueExpression", v.Init)
	}
	return nil
}

func (v ValueExpressionMember) Pos() tokens.Position {
	return v.pos
}

func ParseValueExpressionMember(p *parser.Parser) (*ValueExpressionMember, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	member := ValueExpressionMember{pos: pos}

	switch tok {
	case tokens.PERIOD:
		pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		if tok != tokens.IDENT {
			return nil, ExpectedError(pos, tokens.IDENT, lit)
		}
		member.Init = lit
	case tokens.LPAREN:
		p.Unscan()
		call, err := ParseCallExpression(p)
		if err != nil {
			return nil, err
		}
		member.Init = *call
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

// CallExpression :: LPAREN ((Expression COMMA) | Expression)* RPAREN
type CallExpression struct {
	pos       tokens.Position
	Arguments []Expression
}

func (c CallExpression) Validate() error {
	for _, arg := range c.Arguments {
		if err := arg.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c CallExpression) Pos() tokens.Position {
	return c.pos
}

func ParseCallExpression(p *parser.Parser) (*CallExpression, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	call := CallExpression{pos: pos, Arguments: []Expression{}}

	if tok != tokens.LPAREN {
		return nil, ExpectedError(pos, tokens.LPAREN, lit)
	}

	for {
		_, tok, _ = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		if tok == tokens.RPAREN {
			break
		}
		p.Unscan()

		arg, err := ParseExpression(p)
		if err != nil {
			return nil, err
		}
		call.Arguments = append(call.Arguments, *arg)

		pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		if tok == tokens.RPAREN {
			break
		}
		if tok != tokens.COMMA {
			return nil, ExpectedError(pos, tokens.COMMA, lit)
		}
	}
	return &call, nil
}

// IndexExpression :: LSQUARE Expression? SEMICOLON? Expression? RSQUARE
type IndexExpression struct {
	pos     tokens.Position
	Left    *Expression
	IsRange bool
	Right   *Expression
}

func (i IndexExpression) Validate() error {
	if i.Left != nil {
		if err := i.Left.Validate(); err != nil {
			return err
		}
	}
	if i.Right != nil {
		if err := i.Right.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (i IndexExpression) Pos() tokens.Position {
	return i.pos
}

func ParseIndexExpression(p *parser.Parser) (*IndexExpression, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	index := IndexExpression{pos: pos}

	if tok != tokens.LSQUARE {
		return nil, ExpectedError(pos, tokens.LSQUARE, lit)
	}

	_, tok, _ = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	p.Unscan()
	if tok != tokens.COLON {
		expr, err := ParseExpression(p)
		if err != nil {
			return nil, err
		}
		index.Left = expr
	}

	_, tok, _ = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok == tokens.RSQUARE {
		return &index, nil
	} else if tok == tokens.COLON {
		index.IsRange = true
		_, tok, _ = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		p.Unscan()
		if tok != tokens.RSQUARE {
			expr, err := ParseExpression(p)
			if err != nil {
				return nil, err
			}
			index.Right = expr
			_, tok, _ = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
			if tok != tokens.RSQUARE {
				return nil, ExpectedError(pos, tokens.RSQUARE, lit)
			}
		}
	}
	return &index, nil
}

// AssignmentExpression :: Selector token(IsAssignmentOperator) Expression
//
//	| Selector (INC | DEC)
type AssignmentExpression struct {
	pos      tokens.Position
	Name     Selector
	Operator tokens.Token
	Init     Expression
}

func (a AssignmentExpression) Validate() error {
	if err := a.Name.Validate(); err != nil {
		return err
	}
	if !a.Operator.IsAssignmentOperator() {
		return fmt.Errorf("AssignmentExpression has invalid operator: %s", a.Operator)
	}
	if err := a.Init.Validate(); err != nil {
		return err
	}
	return nil
}

func (a AssignmentExpression) Pos() tokens.Position {
	return a.pos
}

func ParseAssignmentExpression(p *parser.Parser) (*AssignmentExpression, error) {
	name, err := ParseSelector(p)
	if err != nil {
		return nil, err
	}
	assign := AssignmentExpression{pos: name.pos, Name: *name}

	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok == tokens.INC || tok == tokens.DEC {
		if tok == tokens.INC {
			assign.Operator = tokens.ADD
		} else {
			assign.Operator = tokens.SUB
		}
		assign.Init = Expression{pos, Literal{pos, int64(1)}}
	} else if tok.IsAssignmentOperator() {
		assign.Operator = tok
		init, err := ParseExpression(p)
		if err != nil {
			return nil, err
		}
		assign.Init = *init
	} else {
		return nil, ExpectedError(pos, tokens.ILLEGAL, lit)
	}
	return &assign, nil
}

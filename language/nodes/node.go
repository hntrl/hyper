package nodes

import (
	"encoding/gob"
	"fmt"

	"github.com/hntrl/lang/language/tokens"
)

type Node interface {
	Validate() error
	Pos() tokens.Position
}

// type SymbolParser interface {
// 	Scan() (tokens.Position, tokens.Token, string)
// 	ScanIgnore(...tokens.Token) (tokens.Position, tokens.Token, string)
// 	Unscan()
// 	Rollback(int)
// 	Index() int
// }

func ExpectedError(pos tokens.Position, expected tokens.Token, lit string) error {
	return fmt.Errorf("syntax (%s): expected %s but got %s", pos.String(), expected.String(), lit)
}

func init() {
	for _, node := range NodeTypes {
		gob.Register(node)
	}
}

// export of node types for registering gob encoding
var NodeTypes = []Node{
	Manifest{},
	ImportStatement{},
	Selector{},
	Literal{},
	Context{},
	ContextObject{},
	ContextObjectMethod{},
	ContextMethod{},
	FieldStatement{},
	AssignmentStatement{},
	EnumStatement{},
	TypeStatement{},
	// FunctionStatement{},
	TypeExpression{},
	Expression{},
	ArrayExpression{},
	InstanceExpression{},
	UnaryExpression{},
	BinaryExpression{},
	ValueExpression{},
	ValueExpressionMember{},
	CallExpression{},
	IndexExpression{},
	AssignmentExpression{},
	ObjectPattern{},
	PropertyList{},
	Property{},
	SpreadElement{},
	ArgumentList{},
	ArgumentItem{},
	ArgumentObject{},
	FunctionBlock{},
	FunctionExpression{},
	Block{},
	InlineBlock{},
	BlockStatement{},
	DeclarationStatement{},
	IfStatement{},
	WhileStatement{},
	ForStatement{},
	ForCondition{},
	RangeCondition{},
	ContinueStatement{},
	BreakStatement{},
	SwitchBlock{},
	SwitchStatement{},
	GuardStatement{},
	ReturnStatement{},
	ThrowStatement{},
}

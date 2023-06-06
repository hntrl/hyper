package symbols

import (
	"fmt"
	"strings"

	"github.com/hntrl/hyper/internal/ast"
	"github.com/hntrl/hyper/internal/tokens"
)

type InterpreterError struct {
	Node ast.Node
	Msg  string
}

func (e InterpreterError) Error() string {
	if e.Node != nil {
		return fmt.Sprintf("(%s) %s", e.Node.Pos(), e.Msg)
	}
	return e.Msg
}

type InterpreterErrorList []InterpreterError

func (p InterpreterErrorList) Error() string {
	out := make([]string, len(p))
	for i, err := range p {
		out[i] = err.Error()
	}
	return strings.Join(out, "\n")
}

func (p *InterpreterErrorList) Add(err InterpreterError) {
	*p = append(*p, err)
}
func (p *InterpreterErrorList) AddList(list InterpreterErrorList) {
	*p = append(*p, list...)
}

func NodeError(node ast.Node, format string, a ...any) InterpreterError {
	return InterpreterError{node, fmt.Sprintf(format, a...)}
}

func InvalidClassError(node ast.TypeExpression, val ScopeValue) InterpreterError {
	return NodeError(node, "cannot use %T in type expression", val)
}

func UnknownExpressionTypeError(node ast.Node) InterpreterError {
	return NodeError(node, "unknown expression type %T", node)
}

func NotInstanceableError(node ast.InstanceExpression, target string) InterpreterError {
	return NodeError(node, "%s is not instanceable", target)
}

func InvalidNotUnaryOperandError(node ast.UnaryExpression, class Class) InterpreterError {
	return NodeError(node, "cannot apply unary ! to %s", class.Name())
}

func UnknownUnaryOperator(node ast.UnaryExpression, operator tokens.Token) InterpreterError {
	return NodeError(node, "unknown unary opreator %s", operator)
}

func InvalidConstructExpressionError(node ast.CallExpression) InterpreterError {
	return NodeError(node, "invalid class construction")
}

func UncallableError(node ast.CallExpression, val ScopeValue) InterpreterError {
	switch obj := val.(type) {
	case ValueObject:
		return NodeError(node, "%s is not callable", obj.Class().Name())
	default:
		return NodeError(node, "%T is not callable", val)
	}
}

func InvalidValueExpressionError(node ast.ValueExpression) InterpreterError {
	return NodeError(node, "invalid value expression")
}

func NonEnumerableIndexError(node ast.Node, val ScopeValue) InterpreterError {
	switch obj := val.(type) {
	case Class:
		return NodeError(node, "cannot take index of non-enumerable class %s", obj.Name())
	default:
		return NodeError(node, "cannot take index of non-enumerable type %T", obj)
	}
}

func InvalidIndexError(node ast.Node, class Class) InterpreterError {
	return NodeError(node, "index in assignment must be an Integer, got %s", class.Name())
}

func UnknownSelectorError(node ast.Node, selector string) InterpreterError {
	return NodeError(node, "unknown selector %s", selector)
}

func NotIterableError(node ast.Node, val ScopeValue) InterpreterError {
	switch obj := val.(type) {
	case Class:
		return NodeError(node, "%s is not iterable", obj.Name())
	default:
		return NodeError(node, "%T is not iterable", obj)
	}
}

func CannotReassignImmutableValueError(node ast.Node, name string) InterpreterError {
	return NodeError(node, "cannot reassign immutable value %s", name)
}

func CannotRedeclareValueError(node ast.DeclarationStatement) InterpreterError {
	return NodeError(node, "cannot redeclare value %s", node.Name)
}

func InvalidAssignmentIndicesError(node ast.IndexExpression) InterpreterError {
	return NodeError(node, "end index cannot be greater than start index in assignment")
}

func InvalidAssignmentStatementTargetError(node ast.AssignmentStatement) InterpreterError {
	return NodeError(node.Target, "invalid assignment statement target")
}

func CannotAssignNonValueObjectError(node ast.AssignmentStatement) InterpreterError {
	return NodeError(node.Target, "cannot assign non-value object")
}

func BadIfConditionError(node ast.IfStatement) InterpreterError {
	return NodeError(node.Condition, "if condition must be a boolean")
}

func BadIfAlternateError(node ast.Node) InterpreterError {
	return NodeError(node, "invalid if condition alternate")
}

func BadWhileConditionError(node ast.WhileStatement) InterpreterError {
	return NodeError(node.Condition, "while condition must be a boolean")
}

func BadForConditionError(node ast.Node) InterpreterError {
	return NodeError(node, "for condition must be a boolean")
}

func InoperableSwitchTargetError(node ast.SwitchBlock, val ScopeValue) InterpreterError {
	switch obj := val.(type) {
	case ValueObject:
		return NodeError(node, "switch target %s is not operable", obj.Class().Name())
	case Class:
		return NodeError(node, "switch target %s is not operable", obj.Name())
	default:
		return NodeError(node, "switch target %T is not operable", obj)
	}
}

func MultipleSwitchDefaultBlocksError(node ast.SwitchBlock) InterpreterError {
	return NodeError(node, "switch statement can only have one default block")
}

func InvalidThrowValueError(node ast.ThrowStatement) InterpreterError {
	return NodeError(node, "throw statement must be an error")
}

func UnknownBlockStatementError(node ast.Node) InterpreterError {
	return NodeError(node, "unknown block statement type %T", node)
}

func ContinueOutsideLoopError(node ast.ContinueStatement) InterpreterError {
	return NodeError(node, "continue statement outside loop")
}

func BreakOutsideLoopError(node ast.BreakStatement) InterpreterError {
	return NodeError(node, "break statement outside loop")
}

func MismatchedReturnClassError(node ast.ReturnStatement, from, to Class) InterpreterError {
	return NodeError(node, "should return %s, got %s", to.Name(), from.Name())
}

func MissingReturnError(node ast.Node) InterpreterError {
	return NodeError(node, "missing return")
}

func NotSpreadableError(node ast.SpreadElement) InterpreterError {
	return NodeError(node, "cannot spread value without properties")
}

func NoSelectorMemberError(node ast.Selector, chain string, member string) InterpreterError {
	return NodeError(node, "%s has no member %s", chain, member)
}

func UnknownLiteralTypeError(node ast.Literal) error {
	return NodeError(node, "unknown literal type %T", node.Value)
}

func InvalidDestructuredArgumentObjectError() error {
	return fmt.Errorf("destructured argument must be a map value")
}

func MismatchedArgumentLengthError(expected int, got int) error {
	return fmt.Errorf("expected %v arguments, got %v", expected, got)
}

func NoPropertyError(val ScopeValue, key string) error {
	switch obj := val.(type) {
	case ValueObject:
		return fmt.Errorf("%s has no property %s", obj.Class().Name(), key)
	case Class:
		return fmt.Errorf("%s has no property %s", obj.Name(), key)
	default:
		return fmt.Errorf("%T has no property %s", val, key)
	}
}

func CannotAccessPropertyError(val ScopeValue, key string) error {
	switch obj := val.(type) {
	case Class:
		return fmt.Errorf("cannot access property %s on %s", key, obj.Name())
	default:
		return fmt.Errorf("cannot access property %s on %T", key, obj)
	}
}

func CannotSetPropertyError(val ScopeValue, key string) error {
	switch obj := val.(type) {
	case ValueObject:
		return fmt.Errorf("cannot set property %s on %s", key, obj.Class().Name())
	case Class:
		return fmt.Errorf("cannot set property %s on %s", key, obj.Name())
	default:
		return fmt.Errorf("cannot set property %s on %T", key, obj)
	}
}

func CannotSetImmutablePropertyError(key string) error {
	return fmt.Errorf("cannot set immutable property %s", key)
}

func CannotConstructError(from, to Class) error {
	return fmt.Errorf("cannot construct %s from %s", to.Name(), from.Name())
}

func CannotSetNonEnumerableIndexError(class Class) error {
	return fmt.Errorf("cannot set property on non-enumerable class %s", class.Name())
}

func IndexOutOfRangeError() error {
	return fmt.Errorf("index out of range")
}

func StartIndexOutOfRangeError() error {
	return fmt.Errorf("start index out of range")
}

func EndIndexOutOfRangeError() error {
	return fmt.Errorf("end index out of range")
}

func InvalidIndicesError() error {
	return fmt.Errorf("start index cannot be greater than end index")
}

func UndefinedOperatorError(token tokens.Token, target Class, value Class) error {
	return fmt.Errorf("%s operator not defined between %s and %s", token, target.Name(), value.Name())
}

func InvalidOperatorError(operator tokens.Token) error {
	return fmt.Errorf("invalid binary operator %s", operator)
}

func InvalidCompareOperatorError(operator tokens.Token) error {
	return fmt.Errorf("invalid compare operator %s", operator)
}

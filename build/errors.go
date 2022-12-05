package build

import (
	"fmt"
	"strings"

	"github.com/hntrl/lang/language/nodes"
)

type ParserError struct {
	Node nodes.Node
	Msg  string
}

func (e ParserError) Error() string {
	return fmt.Sprintf("(%s) %s", e.Node.Pos(), e.Msg)
}

type ParserErrorList []ParserError

func (p ParserErrorList) Error() string {
	out := make([]string, len(p))
	for i, err := range p {
		out[i] = err.Error()
	}
	return strings.Join(out, "\n")
}

func (p *ParserErrorList) Add(err ParserError) {
	*p = append(*p, err)
}
func (p *ParserErrorList) AddList(list ParserErrorList) {
	*p = append(*p, list...)
}

// Applies a position to an error
func NodeError(node nodes.Node, format string, a ...any) ParserError {
	return ParserError{node, fmt.Sprintf(format, a...)}
}

func UnknownSelector(node nodes.Node, selector string) ParserError {
	return NodeError(node, "unknown selector %s", selector)
}

func AmbiguousObjectError(node nodes.Node, obj Object) ParserError {
	return NodeError(node, "cannot evaluate ambiguous object %T", obj)
}

func NoPropertyError(node nodes.Node, resolveChain string, obj Object, key string) ParserError {
	if valueObj, ok := obj.(ValueObject); ok {
		return NodeError(node, "%s (%s) has no property %s", resolveChain, valueObj.Class().ClassName(), key)
	} else if class, ok := obj.(Class); ok {
		return NodeError(node, "%s (%s) has no property %s", resolveChain, class.ClassName(), key)
	} else {
		return NodeError(node, "%s has no property %s", resolveChain, key)
	}
}

func NotIndexableError(node nodes.Node, resolveChain string, obj Object) ParserError {
	if valueObj, ok := obj.(ValueObject); ok {
		return NodeError(node, "%s (%s) is not indexable", obj, valueObj.Class().ClassName())
	} else {
		return NodeError(node, "%s is not indexable", obj)
	}
}

func NotIterableError(node nodes.Node, obj Object) ParserError {
	if valueObj, ok := obj.(ValueObject); ok {
		return NodeError(node, "%s is not iterable", valueObj.Class().ClassName())
	} else {
		return NodeError(node, "expression is not iterable")
	}
}

func UncallableError(node nodes.Node, resolveChain string, obj Object) ParserError {
	if valueObj, ok := obj.(ValueObject); ok {
		return NodeError(node, "%s (%s) is not callable", resolveChain, valueObj.Class().ClassName())
	} else {
		return NodeError(node, "%s is not callable", resolveChain)
	}
}

func InvalidIndexError(node nodes.Node, obj Object) ParserError {
	return NodeError(node, "got %T for index, expected Integer", obj)
}

func InvalidTypeError(node nodes.Node, obj Object) ParserError {
	return NodeError(node, "cannot use %T for type", obj)
}

func InvalidArgumentLengthError(node nodes.Node, expected, got []Object) ParserError {
	return NodeError(node, "expected %d arguments, got %d", len(expected), len(got))
}

func InvalidReturnTypeError(node nodes.Node, obj Class, class Class) ParserError {
	return NodeError(node, "return type %s does not match expected %s", obj.ClassName(), class.ClassName())
}

func InoperableSwitchTargetError(node nodes.Node, obj Object) ParserError {
	if valueObj, ok := obj.(ValueObject); ok {
		return NodeError(node, "switch target %s is not operable", valueObj.Class().ClassName())
	} else {
		return NodeError(node, "switch target is not operable")
	}
}

func CannotSetPropertyError(key string, obj Object) ParserError {
	if valueObj, ok := obj.(ValueObject); ok {
		return ParserError{nil, fmt.Sprintf("cannot set property %s on %s", key, valueObj.Class().ClassName())}
	} else {
		return ParserError{nil, fmt.Sprintf("cannot set property %s on %T", key, obj)}
	}
}

func UnknownInterfaceError(node nodes.Node, className string) ParserError {
	return NodeError(node, "unknown interface %s", className)
}

func InvalidValueExpressionError(node nodes.Node) ParserError {
	return NodeError(node, "invalid value expression")
}

func CannotConstructError(className string, from string) error {
	return fmt.Errorf("cannot construct %s from %s", className, from)
}

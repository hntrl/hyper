package errors

import (
	"fmt"

	"github.com/hntrl/hyper/internal/ast"
)

type InterpreterError struct {
	Node ast.Node
	Code Code
	Msg  string
}

func (e InterpreterError) Error() string {
	if e.Node != nil {
		return fmt.Sprintf("(%s) %s", e.Node.Pos(), e.Msg)
	}
	return e.Msg
}

func NodeError(node ast.Node, code Code, format string, a ...any) InterpreterError {
	return InterpreterError{
		Node: node,
		Code: code,
		Msg:  fmt.Sprintf(format, a...),
	}
}
func WrappedNodeError(node ast.Node, err error) InterpreterError {
	if ie, ok := err.(InterpreterError); ok {
		return InterpreterError{
			Node: node,
			Code: ie.Code,
			Msg:  ie.Msg,
		}
	} else {
		return InterpreterError{
			Node: node,
			Msg:  err.Error(),
		}
	}
}
func StandardError(code Code, format string, a ...any) InterpreterError {
	return InterpreterError{
		Node: nil,
		Code: code,
		Msg:  fmt.Sprintf(format, a...),
	}
}

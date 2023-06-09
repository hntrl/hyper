package symbols

import (
	"reflect"

	"github.com/hntrl/hyper/internal/ast"
	. "github.com/hntrl/hyper/internal/symbols/errors"
)

type Function struct {
	argumentTypes []Class
	returnType    Class
	handler       functionHandlerFn
}

func (fn Function) Arguments() []Class {
	return fn.argumentTypes
}
func (fn Function) Returns() Class {
	return fn.returnType
}
func (fn Function) Call(args ...ValueObject) (ValueObject, error) {
	return fn.handler(args...)
}

type FunctionOptions struct {
	Arguments []Class
	Returns   Class
	Handler   interface{}
}

func NewFunction(opts FunctionOptions) *Function {
	fn, err := makeFunctionHandlerFn(opts.Arguments, opts.Returns, opts.Handler)
	if err != nil {
		panic(err)
	}
	return &Function{
		argumentTypes: opts.Arguments,
		returnType:    opts.Returns,
		handler:       fn,
	}
}

type functionHandlerFn func(...ValueObject) (ValueObject, error)

func makeFunctionHandlerFn(args []Class, returns Class, callback interface{}) (functionHandlerFn, error) {
	expectedSignature := callbackSignature{
		args:    make([]reflect.Type, len(args)),
		returns: []reflect.Type{},
	}
	for idx := range args {
		expectedSignature.args[idx] = emptyValueObjectType
	}
	if returns != nil {
		expectedSignature.returns = []reflect.Type{emptyValueObjectType, emptyErrorType}
	} else {
		expectedSignature.returns = []reflect.Type{emptyErrorType}
	}
	cb := newCallback(callback)
	if !cb.AcceptsParameters(expectedSignature) {
		return nil, StandardError(ExpectedCallbackSignaure, "expected signature %s, got %s", expectedSignature.String(), cb.Signature.String())
	}
	return func(args ...ValueObject) (ValueObject, error) {
		argValues := make([]reflect.Value, len(args))
		for idx, arg := range args {
			argValues[idx] = reflect.ValueOf(arg)
		}
		returnValues := cb.Call(argValues)
		if value, ok := returnValues[0].Interface().(ValueObject); ok {
			err := returnValues[1].Interface().(error)
			return value, err
		} else {
			err := returnValues[0].Interface().(error)
			return nil, err
		}
	}, nil
}

func (st *SymbolTable) ResolveFunctionBlock(node ast.FunctionBlock) (*Function, error) {
	defTable := st.Clone()
	argumentTypes, returns, err := defTable.EvaluateFunctionParameters(node.Parameters)
	if err != nil {
		return nil, err
	}
	blockDoesReturn, err := defTable.EvaluateBlock(node.Body, returns)
	if err != nil {
		return nil, err
	}
	if !blockDoesReturn {
		return nil, NodeError(node.Body, MissingReturn, "missing return")
	}
	return &Function{
		argumentTypes: argumentTypes,
		returnType:    returns,
		handler: func(args ...ValueObject) (ValueObject, error) {
			scopeTable := st.Clone()
			err = scopeTable.ApplyArgumentList(node.Parameters.Arguments, args)
			if err != nil {
				return nil, err
			}
			obj, err := scopeTable.ResolveBlock(node.Body)
			if err != nil {
				return nil, err
			}
			return obj, nil
		},
	}, nil
}

func (st *SymbolTable) EvaluateFunctionParameters(node ast.FunctionParameters) (args []Class, returns Class, err error) {
	if node.Arguments.Items != nil {
		args, err = st.EvaluateArgumentList(node.Arguments)
		if err != nil {
			return nil, nil, err
		}
	}
	if node.ReturnType != nil {
		returns, err = st.EvaluateTypeExpression(*node.ReturnType)
		if err != nil {
			return nil, nil, err
		}
	}
	return
}

func (st *SymbolTable) ApplyArgumentList(node ast.ArgumentList, args []ValueObject) error {
	for idx, item := range node.Items {
		switch argNode := item.(type) {
		case ast.ArgumentItem:
			st.Local[argNode.Key] = args[idx]
		case ast.ArgumentObject:
			for _, item := range argNode.Items {
				mapValue, ok := args[idx].(*MapValue)
				if !ok {
					return NodeError(argNode, InvalidDestructuredArgument, "destructured argument must be a map value")
				}
				propValue := mapValue.Data[item.Key]
				if propValue == nil {
					return NodeError(argNode, UnknownProperty, "%s has no property %s", mapValue.Class().Name(), item.Key)
				}
				st.Local[item.Key] = propValue
			}
		}
	}
	return nil
}
func (st *SymbolTable) EvaluateArgumentList(node ast.ArgumentList) ([]Class, error) {
	args := make([]Class, len(node.Items))
	for idx, item := range node.Items {
		switch argNode := item.(type) {
		case ast.ArgumentItem:
			class, err := st.EvaluateTypeExpression(argNode.Init)
			if err != nil {
				return nil, err
			}
			st.Local[argNode.Key] = class
			args[idx] = class
		case ast.ArgumentObject:
			mapClass := NewMapClass()
			for _, item := range argNode.Items {
				class, err := st.EvaluateTypeExpression(item.Init)
				if err != nil {
					return nil, err
				}
				mapClass.Properties[item.Key] = class
				st.Local[item.Key] = class
			}
			args[idx] = mapClass
		}
	}
	return args, nil
}

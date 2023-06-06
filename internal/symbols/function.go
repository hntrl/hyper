package symbols

import "github.com/hntrl/hyper/internal/ast"

type functionHandler func(args []ValueObject) (ValueObject, error)

type Function struct {
	argumentTypes []Class
	returnType    Class
	handler       functionHandler
}

func (fn Function) Arguments() []Class {
	return fn.argumentTypes
}
func (fn Function) Returns() Class {
	return fn.returnType
}
func (fn Function) Call(args ...ValueObject) (ValueObject, error) {
	return nil, nil
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
		return nil, MissingReturnError(node.Body)
	}
	return &Function{
		argumentTypes: argumentTypes,
		returnType:    returns,
		handler: func(args []ValueObject) (ValueObject, error) {
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
					return InvalidDestructuredArgumentObjectError()
				}
				propValue := mapValue.Data[item.Key]
				if propValue == nil {
					return NoPropertyError(mapValue, item.Key)
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

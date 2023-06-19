package symbols

import (
	"fmt"

	"github.com/hntrl/hyper/internal/ast"
	. "github.com/hntrl/hyper/internal/symbols/errors"
)

type SymbolTable struct {
	Root      Object
	Immutable map[string]ScopeValue
	Local     map[string]ScopeValue
	LoopState *SymbolTableLoopState
}

type SymbolTableLoopState struct {
	ShouldContinue bool
	ShouldBreak    bool
}

func NewSymbolTable(root Object) *SymbolTable {
	return &SymbolTable{
		Root: root,
		Immutable: map[string]ScopeValue{
			"Bool":   Boolean,
			"String": String,
			"Number": Number,
			"Float":  Float,
			"Int":    Integer,
			"Double": Double,
			"print": NewFunction(FunctionOptions{
				Arguments: []Class{Any},
				Returns:   nil,
				Handler: func(a ValueObject) error {
					fmt.Println(a.Value())
					return nil
				},
			}),
		},
		Local:     make(map[string]ScopeValue),
		LoopState: nil,
	}
}

func (st *SymbolTable) Get(key string) (ScopeValue, error) {
	if obj := st.Immutable[key]; obj != nil {
		return obj, nil
	}
	if obj := st.Local[key]; obj != nil {
		return obj, nil
	}
	return st.Root.Get(key)
}
func (st *SymbolTable) Clone() SymbolTable {
	immutable := make(map[string]ScopeValue)
	for k, v := range st.Immutable {
		immutable[k] = v
	}
	local := make(map[string]ScopeValue)
	for k, v := range st.Local {
		local[k] = v
	}
	return SymbolTable{
		Root:      st.Root,
		Immutable: immutable,
		Local:     local,
		LoopState: st.LoopState,
	}
}
func (st *SymbolTable) StartLoop() SymbolTable {
	newTable := st.Clone()
	newTable.LoopState = &SymbolTableLoopState{
		ShouldContinue: false,
		ShouldBreak:    false,
	}
	return newTable
}

func (st *SymbolTable) ResolveSelector(selector ast.Selector) (ScopeValue, error) {
	current, err := st.Get(selector.Members[0])
	if err != nil {
		return nil, err
	}
	resolveChainString := selector.Members[0]
	if current == nil {
		return nil, NodeError(selector, UnknownSelector, "unknown selector %s", resolveChainString)
	}
	if len(selector.Members) > 1 {
		currentObject, ok := current.(Object)
		if !ok {
			return nil, NodeError(selector, CannotAccessProperty, "cannot access property %s on %T", selector.Members[1], current)
		}
		for idx, member := range selector.Members[1 : len(selector.Members)-1] {
			nextObj, err := currentObject.Get(member)
			if err != nil {
				return nil, err
			}
			if nextObj == nil {
				return nil, NodeError(selector, UnknownProperty, "%s has no member %s", resolveChainString, member)
			}
			currentObject, ok = nextObj.(Object)
			if !ok {
				return nil, NodeError(selector, CannotAccessProperty, "cannot access property %s on %T", selector.Members[idx+1], nextObj)
			}
			resolveChainString += "." + member
		}
		return currentObject.Get(selector.Members[len(selector.Members)-1])
	} else {
		return current, nil
	}
}

func (st *SymbolTable) ResolveLiteral(node ast.Literal) (ValueObject, error) {
	switch lit := node.Value.(type) {
	case string:
		return StringValue(lit), nil
	case int64:
		return IntegerValue(lit), nil
	case float64:
		return FloatValue(lit), nil
	case bool:
		return BooleanValue(lit), nil
	case nil:
		return NilValue{}, nil
	default:
		return nil, NodeError(node, InvalidSyntaxTree, "unknown literal type %T", node.Value)
	}
}

func (st *SymbolTable) ResolvePropertyList(node ast.PropertyList) (*MapValue, error) {
	mapValue := NewMapValue()
	for _, prop := range node {
		switch propNode := prop.(type) {
		case ast.Property:
			propertyValue, err := st.ResolveExpression(propNode.Init)
			if err != nil {
				return nil, err
			}
			mapValue.Set(propNode.Key, propertyValue)
		case ast.SpreadElement:
			propertyValue, err := st.ResolveExpression(propNode.Init)
			if err != nil {
				return nil, err
			}
			properties := propertyValue.Class().Descriptors().Properties
			if properties == nil {
				return nil, NodeError(propNode, InvalidSpreadTarget, "cannot spread value without properties")
			}
			for key, attributes := range properties {
				propertyValue, err := attributes.Getter(propertyValue)
				if err != nil {
					return nil, err
				}
				mapValue.Set(key, propertyValue)
			}
		}
	}
	return mapValue, nil
}
func (st *SymbolTable) EvaluatePropertyList(node ast.PropertyList) (*ExpectedValueObject, error) {
	class := NewMapClass()
	for _, prop := range node {
		switch propNode := prop.(type) {
		case ast.Property:
			property, err := st.EvaluateExpression(propNode.Init)
			if err != nil {
				return nil, err
			}
			class.Properties[propNode.Key] = property.Class
		case ast.SpreadElement:
			property, err := st.EvaluateExpression(propNode.Init)
			if err != nil {
				return nil, err
			}
			properties := property.Class.Descriptors().Properties
			if properties == nil {
				return nil, NodeError(propNode, InvalidSpreadTarget, "cannot spread value without properties")
			}
			for key, attributes := range properties {
				class.Properties[key] = attributes.PropertyClass
			}
		}
	}
	return &ExpectedValueObject{class}, nil
}

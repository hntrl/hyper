package symbols

import (
	"github.com/hntrl/hyper/internal/ast"
)

type SymbolTable struct {
	Immutable map[string]ScopeValue
	Local     map[string]ScopeValue
	LoopState *SymbolTableLoopState
}

type SymbolTableLoopState struct {
	ShouldContinue bool
	ShouldBreak    bool
}

func (st *SymbolTable) Get(key string) (ScopeValue, error) {
	if obj := st.Immutable[key]; obj != nil {
		return obj, nil
	}
	if obj := st.Local[key]; obj != nil {
		return obj, nil
	}
	return nil, nil
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
		return nil, UnknownSelectorError(selector, selector.Members[0])
	}
	for _, member := range selector.Members[1:] {
		currentObject, ok := current.(Object)
		if !ok {
			return nil, CannotAccessPropertyError(current, member)
		}
		nextObj, err := currentObject.Get(member)
		if err != nil {
			return nil, err
		}
		if nextObj == nil {
			return nil, NoSelectorMemberError(selector, resolveChainString, member)
		}
		current = nextObj
		resolveChainString += "." + member
	}
	return current, nil
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
		return nil, UnknownLiteralTypeError(node)
	}
}

func (st *SymbolTable) ResolvePropertyList(node ast.PropertyList) (*MapValue, error) {
	value := NewMapValue()
	for _, prop := range node {
		switch propNode := prop.(type) {
		case ast.Property:
			propertyValue, err := st.ResolveExpression(propNode.Init)
			if err != nil {
				return nil, err
			}
			value.ParentClass.Properties[propNode.Key] = propertyValue.Class()
			value.Data[propNode.Key] = propertyValue
		case ast.SpreadElement:
			propertyValue, err := st.ResolveExpression(propNode.Init)
			if err != nil {
				return nil, err
			}
			properties := propertyValue.Class().Descriptors().Properties
			if properties == nil {
				return nil, NotSpreadableError(propNode)
			}
			for key, attributes := range properties {
				propertyValue, err := attributes.Getter(propertyValue)
				if err != nil {
					return nil, err
				}
				value.ParentClass.Properties[key] = attributes.PropertyClass
				value.Data[key] = propertyValue
			}
		}
	}
	return value, nil
}
func (st *SymbolTable) EvaluatePropertyList(node ast.PropertyList) (*MapClass, error) {
	class := NewMapClass()
	for _, prop := range node {
		switch propNode := prop.(type) {
		case ast.Property:
			propertyClass, err := st.EvaluateExpression(propNode.Init)
			if err != nil {
				return nil, err
			}
			class.Properties[propNode.Key] = propertyClass
		case ast.SpreadElement:
			propertyClass, err := st.EvaluateExpression(propNode.Init)
			if err != nil {
				return nil, err
			}
			properties := propertyClass.Descriptors().Properties
			if properties == nil {
				return nil, NotSpreadableError(propNode)
			}
			for key, attributes := range properties {
				class.Properties[key] = attributes.PropertyClass
			}
		}
	}
	return &class, nil
}

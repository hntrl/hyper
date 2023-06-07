package symbols

import (
	"strings"

	"github.com/hntrl/hyper/internal/ast"
	"github.com/hntrl/hyper/internal/tokens"
)

func (st *SymbolTable) EvaluateTypeExpression(node ast.TypeExpression) (Class, error) {
	target, err := st.ResolveSelector(node.Selector)
	if err != nil {
		return nil, err
	}
	class, ok := target.(Class)
	if !ok {
		return nil, InvalidClassError(node, target)
	}
	// if node.IsPartial {

	// }
	if node.IsArray {
		class = NewArrayClass(class)
	}
	if node.IsOptional {
		class = NewNilableClass(class)
	}
	return class, nil
}

func (st *SymbolTable) ResolveExpression(node ast.Expression) (ValueObject, error) {
	switch node := node.Init.(type) {
	case ast.Literal:
		return st.ResolveLiteral(node)
	case ast.ArrayExpression:
		return st.ResolveArrayExpression(node)
	case ast.InstanceExpression:
		return st.ResolveInstanceExpression(node)
	case ast.UnaryExpression:
		return st.ResolveUnaryExpression(node)
	case ast.BinaryExpression:
		return st.ResolveBinaryExpression(node)
	case ast.ObjectPattern:
		return st.ResolvePropertyList(node.Properties)
	case ast.ValueExpression:
		return st.ResolveValueExpression(node)
	case ast.Expression:
		return st.ResolveExpression(node)
	default:
		return nil, UnknownExpressionTypeError(node)
	}
}
func (st *SymbolTable) EvaluateExpression(node ast.Expression) (Class, error) {
	switch node := node.Init.(type) {
	case ast.Literal:
		lit, err := st.ResolveLiteral(node)
		if err != nil {
			return nil, err
		}
		return lit.Class(), nil
	case ast.ArrayExpression:
		return st.EvaluateArrayExpression(node)
	case ast.InstanceExpression:
		return st.EvaluateInstanceExpression(node)
	case ast.UnaryExpression:
		return st.EvaluateUnaryExpression(node)
	case ast.BinaryExpression:
		return st.EvaluateBinaryExpression(node)
	case ast.ObjectPattern:
		return st.EvaluatePropertyList(node.Properties)
	case ast.ValueExpression:
		return st.EvaluateValueExpression(node)
	case ast.Expression:
		return st.EvaluateExpression(node)
	default:
		return nil, UnknownExpressionTypeError(node)
	}
}

func (st *SymbolTable) ResolveArrayExpression(node ast.ArrayExpression) (ValueObject, error) {
	arrayType, err := st.EvaluateTypeExpression(node.Init)
	if err != nil {
		return nil, err
	}
	array := NewArray(arrayType, len(node.Elements))
	for idx, elementNode := range node.Elements {
		element, err := st.ResolveExpression(elementNode)
		if err != nil {
			return nil, NodeError(elementNode, err.Error())
		}
		array.Items[idx], err = Construct(array.ParentClass.ItemClass, element)
		if err != nil {
			return nil, NodeError(elementNode, err.Error())
		}
	}
	return array, nil
}
func (st *SymbolTable) EvaluateArrayExpression(node ast.ArrayExpression) (Class, error) {
	arrayType, err := st.EvaluateTypeExpression(node.Init)
	if err != nil {
		return nil, err
	}
	arrayClass := NewArrayClass(arrayType)
	for _, elementNode := range node.Elements {
		element, err := st.EvaluateExpression(elementNode)
		if err != nil {
			return nil, NodeError(elementNode, err.Error())
		}
		err = ShouldConstruct(arrayClass.ItemClass, element)
		if err != nil {
			return nil, NodeError(elementNode, err.Error())
		}
	}
	return arrayClass, nil
}

func (st *SymbolTable) ResolveInstanceExpression(node ast.InstanceExpression) (ValueObject, error) {
	target, err := st.ResolveSelector(node.Selector)
	if err != nil {
		return nil, err
	}
	class, ok := target.(Class)
	if !ok {
		return nil, NotInstanceableError(node, strings.Join(node.Selector.Members, "."))
	}
	mapObject, err := st.ResolvePropertyList(node.Properties)
	if err != nil {
		return nil, err
	}
	return Construct(class, mapObject)
}
func (st *SymbolTable) EvaluateInstanceExpression(node ast.InstanceExpression) (Class, error) {
	target, err := st.ResolveSelector(node.Selector)
	if err != nil {
		return nil, err
	}
	class, ok := target.(Class)
	if !ok {
		return nil, NotInstanceableError(node, strings.Join(node.Selector.Members, "."))
	}
	mapObject, err := st.EvaluatePropertyList(node.Properties)
	if err != nil {
		return nil, err
	}
	err = ShouldConstruct(class, mapObject)
	if err != nil {
		return nil, err
	}
	return class, nil
}

func (st *SymbolTable) ResolveUnaryExpression(node ast.UnaryExpression) (ValueObject, error) {
	obj, err := st.ResolveExpression(node.Init)
	if err != nil {
		return nil, err
	}
	switch node.Operator {
	case tokens.ADD:
		val, err := Construct(Number, obj)
		if err != nil {
			return nil, err
		}
		num := float64(val.(NumberValue))
		return Construct(obj.Class(), NumberValue(+num))
	case tokens.SUB:
		val, err := Construct(Number, obj)
		if err != nil {
			return nil, err
		}
		num := float64(val.(NumberValue))
		return Construct(obj.Class(), NumberValue(-num))
	case tokens.NOT:
		boolValue, ok := obj.(BooleanValue)
		if !ok {
			return nil, InvalidNotUnaryOperandError(node, obj.Class())
		}
		return BooleanValue(!boolValue), nil
	default:
		return nil, UnknownUnaryOperator(node, node.Operator)
	}
}
func (st *SymbolTable) EvaluateUnaryExpression(node ast.UnaryExpression) (Class, error) {
	class, err := st.EvaluateExpression(node.Init)
	if err != nil {
		return nil, err
	}
	switch node.Operator {
	case tokens.ADD, tokens.SUB:
		err = ShouldConstruct(Number, class)
		if err != nil {
			return nil, NodeError(node, err.Error())
		}
		return class, nil
	case tokens.NOT:
		err = ShouldConstruct(Boolean, class)
		if err != nil {
			return nil, NodeError(node, err.Error())
		}
		return Boolean, nil
	default:
		return nil, UnknownUnaryOperator(node, node.Operator)
	}
}

func (st *SymbolTable) ResolveBinaryExpression(node ast.BinaryExpression) (ValueObject, error) {
	if !node.Operator.IsOperator() && !node.Operator.IsComparableOperator() {
		return nil, InvalidOperatorError(node.Operator)
	}
	left, err := st.ResolveExpression(node.Left)
	if err != nil {
		return nil, err
	}
	right, err := st.ResolveExpression(node.Right)
	if err != nil {
		return nil, err
	}
	obj, err := Operate(node.Operator, left, right)
	if err != nil {
		return nil, NodeError(node, err.Error())
	}
	return obj, nil
}
func (st *SymbolTable) EvaluateBinaryExpression(node ast.BinaryExpression) (Class, error) {
	if !node.Operator.IsOperator() && !node.Operator.IsComparableOperator() {
		return nil, InvalidOperatorError(node.Operator)
	}
	left, err := st.EvaluateExpression(node.Left)
	if err != nil {
		return nil, err
	}
	right, err := st.EvaluateExpression(node.Right)
	if err != nil {
		return nil, err
	}
	err = ShouldOperate(node.Operator, left, right)
	if err != nil {
		return nil, NodeError(node, err.Error())
	}
	if node.Operator.IsComparableOperator() {
		return Boolean, nil
	}
	return left, nil
}

func resolveValueExpressionMember(st *SymbolTable, value ScopeValue, member string) (ScopeValue, error) {
	switch object := value.(type) {
	case ValueObject:
		descriptors := object.Class().Descriptors()
		if descriptors.Prototype != nil {
			classMethod, ok := descriptors.Prototype[member]
			if ok {
				return classMethod, nil
			}
		}
		if descriptors.Properties != nil {
			property, ok := descriptors.Properties[member]
			if ok {
				return property.Getter(object)
			}
		}
		return nil, NoPropertyError(object, member)
	case Object:
		val, err := object.Get(member)
		if err != nil {
			return nil, err
		}
		if val == nil {
			return nil, NoPropertyError(object, member)
		}
		return val, nil
	default:
		return nil, CannotAccessPropertyError(object, member)
	}
}
func evaluateValueExpressionMember(st *SymbolTable, value ScopeValue, member string) (ScopeValue, error) {
	switch object := value.(type) {
	case ValueObject:
		descriptors := object.Class().Descriptors()
		if descriptors.Prototype != nil {
			classMethod, ok := descriptors.Prototype[member]
			if ok {
				return classMethod, nil
			}
		}
		if descriptors.Properties != nil {
			property, ok := descriptors.Properties[member]
			if ok {
				return property.PropertyClass, nil
			}
		}
		return nil, NoPropertyError(object, member)
	case Object:
		val, err := object.Get(member)
		if err != nil {
			return nil, err
		}
		if val == nil {
			return nil, NoPropertyError(object, member)
		}
		return val, nil
	default:
		return nil, CannotAccessPropertyError(object, member)
	}
}
func resolveValueExpressionCallMember(st *SymbolTable, value ScopeValue, node ast.CallExpression) (ValueObject, error) {
	switch object := value.(type) {
	case Callable:
		callableArgs := object.Arguments()
		if len(node.Arguments) != len(callableArgs) {
			return nil, MismatchedArgumentLengthError(len(callableArgs), len(node.Arguments))
		}
		passedArguments := make([]ValueObject, len(node.Arguments))
		for idx, arg := range node.Arguments {
			argValue, err := st.ResolveExpression(arg)
			if err != nil {
				return nil, err
			}
			passedArguments[idx], err = Construct(callableArgs[idx], argValue)
			if err != nil {
				return nil, err
			}
		}
		return object.Call(passedArguments...)
	case Class:
		if len(node.Arguments) != 1 {
			return nil, InvalidConstructExpressionError(node)
		}
		argValue, err := st.ResolveExpression(node.Arguments[0])
		if err != nil {
			return nil, err
		}
		value, err := Construct(object, argValue)
		if err != nil {
			return nil, NodeError(node, err.Error())
		}
		return value, nil
	default:
		return nil, UncallableError(node, object)
	}
}
func evaluateValueExpressionCallMember(st *SymbolTable, value ScopeValue, node ast.CallExpression) (Class, error) {
	switch object := value.(type) {
	case Callable:
		callableArgs := object.Arguments()
		if len(node.Arguments) != len(callableArgs) {
			return nil, MismatchedArgumentLengthError(len(callableArgs), len(node.Arguments))
		}
		for idx, arg := range node.Arguments {
			argClass, err := st.EvaluateExpression(arg)
			if err != nil {
				return nil, err
			}
			err = ShouldConstruct(callableArgs[idx], argClass)
			if err != nil {
				return nil, err
			}
		}
		return object.Returns(), nil
	case Class:
		if len(node.Arguments) != 1 {
			return nil, InvalidConstructExpressionError(node)
		}
		argClass, err := st.EvaluateExpression(node.Arguments[0])
		if err != nil {
			return nil, err
		}
		err = ShouldConstruct(object, argClass)
		if err != nil {
			return nil, NodeError(node, err.Error())
		}
		return object, nil
	default:
		return nil, UncallableError(node, object)
	}
}

func (st *SymbolTable) ResolveValueExpression(node ast.ValueExpression) (ValueObject, error) {
	firstMember, ok := node.Members[0].Init.(string)
	if !ok {
		return nil, InvalidValueExpressionError(node)
	}
	current, err := st.Get(firstMember)
	if err != nil {
		return nil, err
	}
	for _, memberNode := range node.Members[1:] {
		switch member := memberNode.Init.(type) {
		case string:
			current, err = resolveValueExpressionMember(st, current, member)
			if err != nil {
				return nil, err
			}
		case ast.CallExpression:
			current, err = resolveValueExpressionCallMember(st, current, member)
			if err != nil {
				return nil, NodeError(member, err.Error())
			}
		case ast.IndexExpression:
			currentValueObject, ok := current.(ValueObject)
			if !ok {
				return nil, NonEnumerableIndexError(node, current)
			}
			enumerable := currentValueObject.Class().Descriptors().Enumerable
			if enumerable == nil {
				return nil, NonEnumerableIndexError(node, currentValueObject.Class())
			}
			startIndex, endIndex, err := st.ResolveIndexExpression(member, currentValueObject)
			if endIndex != -1 {
				current, err = enumerable.GetRange(currentValueObject, startIndex, endIndex)
				if err != nil {
					return nil, err
				}
			} else {
				current, err = enumerable.GetIndex(currentValueObject, startIndex)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	valueObj, ok := current.(ValueObject)
	if !ok {
		return nil, InvalidValueExpressionError(node)
	}
	return valueObj, nil
}
func (st *SymbolTable) EvaluateValueExpression(node ast.ValueExpression) (Class, error) {
	firstMember, ok := node.Members[0].Init.(string)
	if !ok {
		return nil, InvalidValueExpressionError(node)
	}
	current, err := st.Get(firstMember)
	if err != nil {
		return nil, err
	}
	for _, memberNode := range node.Members[1:] {
		switch member := memberNode.Init.(type) {
		case string:
			current, err = evaluateValueExpressionMember(st, current, member)
		case ast.CallExpression:
			current, err = evaluateValueExpressionCallMember(st, current, member)
		case ast.IndexExpression:
			currentClass, ok := current.(Class)
			if !ok {
				return nil, NonEnumerableIndexError(node, current)
			}
			enumerable := currentClass.Descriptors().Enumerable
			if enumerable == nil {
				return nil, NonEnumerableIndexError(node, currentClass)
			}
			err = st.EvaluateIndexExpression(member, currentClass)
		}
		if err != nil {
			return nil, NodeError(node, err.Error())
		}
	}
	class, ok := current.(Class)
	if !ok {
		return nil, InvalidValueExpressionError(node)
	}
	return class, nil
}

func (st *SymbolTable) ResolveIndexExpression(node ast.IndexExpression, target ValueObject) (int, int, error) {
	enumerable := target.Class().Descriptors().Enumerable
	if enumerable == nil {
		return -1, -1, NonEnumerableIndexError(node, target.Class())
	}
	startIndex := 0
	if node.Left != nil {
		leftValue, err := st.ResolveExpression(*node.Left)
		if err != nil {
			return -1, -1, err
		}
		leftIntValue, ok := leftValue.(IntegerValue)
		if !ok {
			return -1, -1, InvalidIndexError(node.Left, leftValue.Class())
		}
		startIndex = int(leftIntValue)
	}
	if node.IsRange {
		endIndex, err := enumerable.GetLength(target)
		if err != nil {
			return -1, -1, NodeError(node, err.Error())
		}
		if node.Right != nil {
			rightValue, err := st.ResolveExpression(*node.Right)
			if err != nil {
				return -1, -1, err
			}
			rightIntValue, ok := rightValue.(IntegerValue)
			if !ok {
				return -1, -1, InvalidIndexError(node.Right, rightValue.Class())
			}
			endIndex = int(rightIntValue)
		}
		return startIndex, endIndex, nil
	}
	return startIndex, -1, nil
}
func (st *SymbolTable) EvaluateIndexExpression(node ast.IndexExpression, target Class) error {
	enumerable := target.Descriptors().Enumerable
	if enumerable == nil {
		return NonEnumerableIndexError(node, target)
	}
	if node.Left != nil {
		leftClass, err := st.EvaluateExpression(*node.Left)
		if err != nil {
			return err
		}
		if _, ok := leftClass.(IntegerClass); !ok {
			return InvalidIndexError(node.Left, leftClass)
		}
	}
	if node.Right != nil {
		rightClass, err := st.EvaluateExpression(*node.Right)
		if err != nil {
			return err
		}
		if _, ok := rightClass.(IntegerClass); !ok {
			return InvalidIndexError(node, rightClass)
		}
	}
	return nil
}

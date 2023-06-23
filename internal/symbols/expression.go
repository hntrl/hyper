package symbols

import (
	"strings"

	"github.com/hntrl/hyper/internal/ast"
	. "github.com/hntrl/hyper/internal/symbols/errors"
	"github.com/hntrl/hyper/internal/tokens"
)

func (st *SymbolTable) EvaluateTypeExpression(node ast.TypeExpression) (Class, error) {
	target, err := st.ResolveSelector(node.Selector)
	if err != nil {
		return nil, err
	}
	class, ok := target.(Class)
	if !ok {
		return nil, NodeError(node, InvalidClass, "cannot use %T in type expression", target)
	}
	if node.IsPartial {
		class = NewPartialClass(class)
	}
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
	case ast.TemplateLiteral:
		return st.ResolveTemplateLiteral(node)
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
		return nil, NodeError(node, InvalidSyntaxTree, "unknown expression type %T", node)
	}
}
func (st *SymbolTable) EvaluateExpression(node ast.Expression) (*ExpectedValueObject, error) {
	switch node := node.Init.(type) {
	case ast.Literal:
		lit, err := st.ResolveLiteral(node)
		if err != nil {
			return nil, err
		}
		return &ExpectedValueObject{lit.Class()}, nil
	case ast.TemplateLiteral:
		err := st.EvaluateTemplateLiteral(node)
		if err != nil {
			return nil, err
		}
		return &ExpectedValueObject{String}, nil
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
		return nil, NodeError(node, InvalidSyntaxTree, "unknown expression type %T", node)
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
			return nil, err
		}
		constructedElement, err := Construct(array.parentClass.itemClass, element)
		if err != nil {
			return nil, err
		}
		array.Set(idx, constructedElement)
	}
	return array, nil
}
func (st *SymbolTable) EvaluateArrayExpression(node ast.ArrayExpression) (*ExpectedValueObject, error) {
	arrayType, err := st.EvaluateTypeExpression(node.Init)
	if err != nil {
		return nil, err
	}
	arrayClass := NewArrayClass(arrayType)
	for _, elementNode := range node.Elements {
		element, err := st.EvaluateExpression(elementNode)
		if err != nil {
			return nil, err
		}
		err = ShouldConstruct(arrayClass.itemClass, element.Class)
		if err != nil {
			return nil, WrappedNodeError(elementNode, err)
		}
	}
	return &ExpectedValueObject{arrayClass}, nil
}

func (st *SymbolTable) ResolveInstanceExpression(node ast.InstanceExpression) (ValueObject, error) {
	target, err := st.ResolveSelector(node.Selector)
	if err != nil {
		return nil, err
	}
	class, ok := target.(Class)
	if !ok {
		return nil, NodeError(node, InvalidInstanceableTarget, "%s is not instanceable", strings.Join(node.Selector.Members, "."))
	}
	mapObject, err := st.ResolvePropertyList(node.Properties)
	if err != nil {
		return nil, err
	}
	return Construct(class, mapObject)
}
func (st *SymbolTable) EvaluateInstanceExpression(node ast.InstanceExpression) (*ExpectedValueObject, error) {
	target, err := st.ResolveSelector(node.Selector)
	if err != nil {
		return nil, err
	}
	class, ok := target.(Class)
	if !ok {
		return nil, NodeError(node, InvalidInstanceableTarget, "%s is not instanceable", strings.Join(node.Selector.Members, "."))
	}
	expectedMap, err := st.EvaluatePropertyList(node.Properties)
	if err != nil {
		return nil, err
	}
	err = ShouldConstruct(class, expectedMap.Class)
	if err != nil {
		return nil, WrappedNodeError(node, err)
	}
	return &ExpectedValueObject{class}, nil
}

func (st *SymbolTable) ResolveUnaryExpression(node ast.UnaryExpression) (ValueObject, error) {
	obj, err := st.ResolveExpression(node.Init)
	if err != nil {
		return nil, err
	}
	switch node.Operator {
	case tokens.ADD:
		operandValue, err := Construct(Number, obj)
		if err != nil {
			return nil, WrappedNodeError(node, err)
		}
		num := float64(operandValue.(NumberValue))
		newValue, err := Construct(obj.Class(), NumberValue(+num))
		if err != nil {
			return nil, WrappedNodeError(node, err)
		}
		return newValue, nil
	case tokens.SUB:
		operandValue, err := Construct(Number, obj)
		if err != nil {
			return nil, WrappedNodeError(node, err)
		}
		num := float64(operandValue.(NumberValue))
		newValue, err := Construct(obj.Class(), NumberValue(-num))
		if err != nil {
			return nil, WrappedNodeError(node, err)
		}
		return newValue, nil
	case tokens.NOT:
		boolValue, ok := obj.(BooleanValue)
		if !ok {
			return nil, NodeError(node, InvalidUnaryOperand, "cannot apply unary ! to %s", obj.Class().Descriptors().Name)
		}
		return BooleanValue(!boolValue), nil
	default:
		return nil, NodeError(node, BadUnaryOperator, "unknown unary operator %s", node.Operator)
	}
}
func (st *SymbolTable) EvaluateUnaryExpression(node ast.UnaryExpression) (*ExpectedValueObject, error) {
	expectedValue, err := st.EvaluateExpression(node.Init)
	if err != nil {
		return nil, err
	}
	switch node.Operator {
	case tokens.ADD, tokens.SUB:
		err = ShouldConstruct(Number, expectedValue.Class)
		if err != nil {
			return nil, WrappedNodeError(node, err)
		}
		return &ExpectedValueObject{expectedValue.Class}, nil
	case tokens.NOT:
		err = ShouldConstruct(Boolean, expectedValue.Class)
		if err != nil {
			return nil, WrappedNodeError(node, err)
		}
		return &ExpectedValueObject{Boolean}, nil
	default:
		return nil, NodeError(node, BadUnaryOperator, "unknown unary property %s", node.Operator)
	}
}

func (st *SymbolTable) ResolveBinaryExpression(node ast.BinaryExpression) (ValueObject, error) {
	if !node.Operator.IsOperator() && !node.Operator.IsComparableOperator() {
		return nil, NodeError(node, InvalidOperator, "invalid binary operator %s", node.Operator)
	}
	left, err := st.ResolveExpression(node.Left)
	if err != nil {
		return nil, err
	}
	right, err := st.ResolveExpression(node.Right)
	if err != nil {
		return nil, err
	}
	if node.Operator.IsOperator() {
		obj, err := Operate(node.Operator, left, right)
		if err != nil {
			return nil, WrappedNodeError(node, err)
		}
		return obj, nil
	} else if node.Operator.IsComparableOperator() {
		result, err := Compare(node.Operator, left, right)
		if err != nil {
			return nil, WrappedNodeError(node, err)
		}
		return BooleanValue(result), nil
	}
	return nil, NodeError(node, UndefinedOperator, "%s operator not defined between %s and %s", node.Operator, left.Class().Descriptors().Name, right.Class().Descriptors().Name)
}
func (st *SymbolTable) EvaluateBinaryExpression(node ast.BinaryExpression) (*ExpectedValueObject, error) {
	if !node.Operator.IsOperator() && !node.Operator.IsComparableOperator() {
		return nil, NodeError(node, InvalidOperator, "invalid binary operator %s", node.Operator)
	}
	left, err := st.EvaluateExpression(node.Left)
	if err != nil {
		return nil, err
	}
	right, err := st.EvaluateExpression(node.Right)
	if err != nil {
		return nil, err
	}
	if node.Operator.IsOperator() {
		err = ShouldOperate(node.Operator, left.Class, right.Class)
		if err != nil {
			return nil, WrappedNodeError(node, err)
		}
		return &ExpectedValueObject{left.Class}, nil
	} else if node.Operator.IsComparableOperator() {
		err = ShouldCompare(node.Operator, left.Class, right.Class)
		if err != nil {
			return nil, WrappedNodeError(node, err)
		}
		return &ExpectedValueObject{Boolean}, nil
	}
	return nil, NodeError(node, UndefinedOperator, "%s operator not defined between %s and %s", node.Operator, left.Class.Descriptors().Name, right.Class.Descriptors().Name)
}

func resolveValueExpressionMember(st *SymbolTable, value ScopeValue, member string) (ScopeValue, error) {
	switch object := value.(type) {
	case Class:
		descriptors := object.Descriptors()
		if descriptors.ClassProperties != nil {
			property, ok := descriptors.ClassProperties[member]
			if ok {
				return property, nil
			}
		}
		return nil, StandardError(UnknownProperty, "%s has no property %s", object.Descriptors().Name, member)
	case ValueObject:
		descriptors := object.Class().Descriptors()
		if descriptors.Prototype != nil {
			classMethod, ok := descriptors.Prototype[member]
			if ok {
				return classMethod.CallableForValue(object), nil
			}
		}
		if descriptors.Properties != nil {
			property, ok := descriptors.Properties[member]
			if ok {
				return property.Getter(object)
			}
		}
		return nil, StandardError(UnknownProperty, "%s has no property %s", object.Class().Descriptors().Name, member)
	case Object:
		val, err := object.Get(member)
		if err != nil {
			return nil, err
		}
		if val == nil {
			return nil, StandardError(UnknownProperty, "%T has no property %s", object, member)
		}
		return val, nil
	default:
		return nil, StandardError(CannotAccessProperty, "cannot access property %s on %T", member, object)
	}
}
func evaluateValueExpressionMember(st *SymbolTable, value ScopeValue, member string) (ScopeValue, error) {
	switch object := value.(type) {
	case Class:
		descriptors := object.Descriptors()
		if descriptors.ClassProperties != nil {
			property, ok := descriptors.ClassProperties[member]
			if ok {
				return property, nil
			}
		}
		return nil, StandardError(UnknownProperty, "%s has no property %s", object.Descriptors().Name, member)
	case ValueObject:
		return evaluateValueExpressionMember(st, &ExpectedValueObject{object.Class()}, member)
	case *ExpectedValueObject:
		descriptors := object.Class.Descriptors()
		if descriptors.Prototype != nil {
			classMethod, ok := descriptors.Prototype[member]
			if ok {
				return classMethod, nil
			}
		}
		if descriptors.Properties != nil {
			property, ok := descriptors.Properties[member]
			if ok {
				return &ExpectedValueObject{property.PropertyClass}, nil
			}
		}
		return nil, StandardError(UnknownProperty, "%s has no property %s", object.Class.Descriptors().Name, member)
	case Object:
		val, err := object.Get(member)
		if err != nil {
			return nil, err
		}
		if val == nil {
			return nil, StandardError(UnknownProperty, "%T has no property %s", object, member)
		}
		return val, nil
	default:
		return nil, StandardError(CannotAccessProperty, "cannot access property %s on %T", member, object)
	}
}
func resolveValueExpressionCallMember(st *SymbolTable, value ScopeValue, node ast.CallExpression) (ValueObject, error) {
	switch object := value.(type) {
	case Callable:
		callableArgs := object.Arguments()
		if len(node.Arguments) != len(callableArgs) {
			return nil, StandardError(InvalidArgumentLength, "expected %d arguments, got %d", len(callableArgs), len(node.Arguments))
		}
		passedArguments := make([]ValueObject, len(node.Arguments))
		for idx, arg := range node.Arguments {
			argValue, err := st.ResolveExpression(arg)
			if err != nil {
				return nil, err
			}
			passedArguments[idx], err = Construct(callableArgs[idx], argValue)
			if err != nil {
				return nil, WrappedNodeError(node, err)
			}
		}
		return object.Call(passedArguments...)
	case Class:
		if len(node.Arguments) != 1 {
			return nil, StandardError(InvalidClassConstruction, "invalid class construction")
		}
		argValue, err := st.ResolveExpression(node.Arguments[0])
		if err != nil {
			return nil, err
		}
		value, err := Construct(object, argValue)
		if err != nil {
			return nil, WrappedNodeError(node, err)
		}
		return value, nil
	default:
		return nil, NodeError(node, InvalidCallExpression, "%T is not callable", object)
	}
}
func evaluateValueExpressionCallMember(st *SymbolTable, value ScopeValue, node ast.CallExpression) (*ExpectedValueObject, error) {
	switch object := value.(type) {
	case Callable:
		callableArgs := object.Arguments()
		if len(node.Arguments) != len(callableArgs) {
			return nil, StandardError(InvalidArgumentLength, "expected %d arguments, got %d", len(callableArgs), len(node.Arguments))
		}
		for idx, arg := range node.Arguments {
			expectedArg, err := st.EvaluateExpression(arg)
			if err != nil {
				return nil, err
			}
			err = ShouldConstruct(callableArgs[idx], expectedArg.Class)
			if err != nil {
				return nil, WrappedNodeError(node, err)
			}
		}
		return &ExpectedValueObject{object.Returns()}, nil
	case *ClassMethod:
		methodArgs := object.ArgumentTypes
		for idx, arg := range node.Arguments {
			expectedArg, err := st.EvaluateExpression(arg)
			if err != nil {
				return nil, err
			}
			err = ShouldConstruct(methodArgs[idx], expectedArg.Class)
			if err != nil {
				return nil, WrappedNodeError(node, err)
			}
		}
		return &ExpectedValueObject{object.ReturnType}, nil
	case Class:
		if len(node.Arguments) != 1 {
			return nil, StandardError(InvalidClassConstruction, "invalid class construction")
		}
		expectedArg, err := st.EvaluateExpression(node.Arguments[0])
		if err != nil {
			return nil, err
		}
		err = ShouldConstruct(object, expectedArg.Class)
		if err != nil {
			return nil, WrappedNodeError(node, err)
		}
		return &ExpectedValueObject{object}, nil
	default:
		return nil, NodeError(node, InvalidCallExpression, "%T is not callable", object)
	}
}

func (st *SymbolTable) ResolveValueExpression(node ast.ValueExpression) (ValueObject, error) {
	firstMember, ok := node.Members[0].Init.(string)
	if !ok {
		return nil, NodeError(node, InvalidValueExpression, "invalid value expression")
	}
	current, err := st.Get(firstMember)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, NodeError(node, UnknownSelector, "unknown selector %s", firstMember)
	}
	for _, memberNode := range node.Members[1:] {
		switch member := memberNode.Init.(type) {
		case string:
			current, err = resolveValueExpressionMember(st, current, member)
			if err != nil {
				return nil, WrappedNodeError(node, err)
			}
		case ast.CallExpression:
			current, err = resolveValueExpressionCallMember(st, current, member)
			if err != nil {
				return nil, err
			}
		case ast.IndexExpression:
			currentValueObject, ok := current.(ValueObject)
			if !ok {
				return nil, NodeError(node, InvalidIndexTarget, "cannot take index of non-enumerable %T", current)
			}
			enumerable := currentValueObject.Class().Descriptors().Enumerable
			if enumerable == nil {
				return nil, NodeError(node, InvalidIndexTarget, "cannot take index of non-enumerable class %s", currentValueObject.Class().Descriptors().Name)
			}
			startIndex, endIndex, err := st.ResolveIndexExpression(member, currentValueObject)
			if err != nil {
				return nil, err
			}
			if endIndex != -1 {
				current, err = enumerable.GetRange(currentValueObject, startIndex, endIndex)
				if err != nil {
					return nil, WrappedNodeError(node, err)
				}
			} else {
				current, err = enumerable.GetIndex(currentValueObject, startIndex)
				if err != nil {
					return nil, WrappedNodeError(node, err)
				}
			}
		}
	}
	valueObj, ok := current.(ValueObject)
	if !ok {
		return nil, NodeError(node, InvalidValueExpression, "invalid value expression")
	}
	return valueObj, nil
}
func (st *SymbolTable) EvaluateValueExpression(node ast.ValueExpression) (*ExpectedValueObject, error) {
	firstMember, ok := node.Members[0].Init.(string)
	if !ok {
		return nil, NodeError(node, InvalidValueExpression, "invalid value expression")
	}
	current, err := st.Get(firstMember)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, NodeError(node, UnknownSelector, "unknown selector %s", firstMember)
	}
	for _, memberNode := range node.Members[1:] {
		switch member := memberNode.Init.(type) {
		case string:
			current, err = evaluateValueExpressionMember(st, current, member)
			if err != nil {
				return nil, WrappedNodeError(node, err)
			}
		case ast.CallExpression:
			current, err = evaluateValueExpressionCallMember(st, current, member)
			if err != nil {
				return nil, err
			}
		case ast.IndexExpression:
			currentObject, ok := current.(*ExpectedValueObject)
			if !ok {
				return nil, NodeError(node, InvalidIndexTarget, "cannot take index of non-enumerable %T", current)
			}
			enumerable := currentObject.Class.Descriptors().Enumerable
			if enumerable == nil {
				return nil, NodeError(node, InvalidIndexTarget, "cannot take index of non-enumerable class %s", currentObject.Class.Descriptors().Name)
			}
			err = st.EvaluateIndexExpression(member, currentObject)
			if err != nil {
				return nil, WrappedNodeError(node, err)
			}
			if member.IsRange {
				current = &ExpectedValueObject{currentObject.Class}
			} else {
				if arrayClass, ok := currentObject.Class.(ArrayClass); ok {
					current = &ExpectedValueObject{arrayClass.itemClass}
				} else {
					current = &ExpectedValueObject{currentObject.Class}
				}
			}
		}
	}
	if valueObject, ok := current.(ValueObject); ok {
		current = &ExpectedValueObject{valueObject.Class()}
	}
	expected, ok := current.(*ExpectedValueObject)
	if !ok {
		return nil, NodeError(node, InvalidValueExpression, "invalid value expression")
	}
	return expected, nil
}

func (st *SymbolTable) ResolveIndexExpression(node ast.IndexExpression, target ValueObject) (int, int, error) {
	enumerable := target.Class().Descriptors().Enumerable
	if enumerable == nil {
		return -1, -1, NodeError(node, InvalidIndexTarget, "cannot take index of non-enumerable class %s", target.Class().Descriptors().Name)
	}
	startIndex := 0
	if node.Left != nil {
		leftValue, err := st.ResolveExpression(*node.Left)
		if err != nil {
			return -1, -1, err
		}
		leftIntValue, ok := leftValue.(IntegerValue)
		if !ok {
			return -1, -1, NodeError(node, InvalidIndex, "index must be an Integer, got %s", leftValue.Class().Descriptors().Name)
		}
		startIndex = int(leftIntValue)
	}
	if node.IsRange {
		endIndex, err := enumerable.GetLength(target)
		if err != nil {
			return -1, -1, WrappedNodeError(node, err)
		}
		if node.Right != nil {
			rightValue, err := st.ResolveExpression(*node.Right)
			if err != nil {
				return -1, -1, err
			}
			rightIntValue, ok := rightValue.(IntegerValue)
			if !ok {
				return -1, -1, NodeError(node.Right, InvalidIndex, "index must be an Integer, got %s", rightValue.Class().Descriptors().Name)
			}
			endIndex = int(rightIntValue)
		}
		return startIndex, endIndex, nil
	}
	return startIndex, -1, nil
}
func (st *SymbolTable) EvaluateIndexExpression(node ast.IndexExpression, target *ExpectedValueObject) error {
	enumerable := target.Class.Descriptors().Enumerable
	if enumerable == nil {
		return NodeError(node, InvalidIndexTarget, "cannot take index of non-enumerable class %s", target.Class.Descriptors().Name)
	}
	if node.Left != nil {
		left, err := st.EvaluateExpression(*node.Left)
		if err != nil {
			return err
		}
		if _, ok := left.Class.(IntegerClass); !ok {
			return NodeError(node, InvalidIndex, "index must be an Integer, got %s", left.Class.Descriptors().Name)
		}
	}
	if node.Right != nil {
		right, err := st.EvaluateExpression(*node.Right)
		if err != nil {
			return err
		}
		if _, ok := right.Class.(IntegerClass); !ok {
			return NodeError(node, InvalidIndex, "index must be an Integer, got %s", right.Class.Descriptors().Name)
		}
	}
	return nil
}

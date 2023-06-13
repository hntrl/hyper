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
		return nil, NodeError(node, InvalidSyntaxTree, "unknown expression type %T", node)
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
		array.items[idx], err = Construct(array.parentClass.itemClass, element)
		if err != nil {
			return nil, WrappedNodeError(elementNode, err)
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
			return nil, err
		}
		err = ShouldConstruct(arrayClass.itemClass, element)
		if err != nil {
			return nil, WrappedNodeError(elementNode, err)
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
		return nil, NodeError(node, InvalidInstanceableTarget, "%s is not instanceable", strings.Join(node.Selector.Members, "."))
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
		return nil, NodeError(node, InvalidInstanceableTarget, "%s is not instanceable", strings.Join(node.Selector.Members, "."))
	}
	mapObject, err := st.EvaluatePropertyList(node.Properties)
	if err != nil {
		return nil, err
	}
	err = ShouldConstruct(class, mapObject)
	if err != nil {
		return nil, WrappedNodeError(node, err)
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
			return nil, NodeError(node, InvalidUnaryOperand, "cannot apply unary ! to %s", obj.Class().Name())
		}
		return BooleanValue(!boolValue), nil
	default:
		return nil, NodeError(node, BadUnaryOperator, "unknown unary property %s", node.Operator)
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
			return nil, WrappedNodeError(node, err)
		}
		return class, nil
	case tokens.NOT:
		err = ShouldConstruct(Boolean, class)
		if err != nil {
			return nil, WrappedNodeError(node, err)
		}
		return Boolean, nil
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
	obj, err := Operate(node.Operator, left, right)
	if err != nil {
		return nil, WrappedNodeError(node, err)
	}
	return obj, nil
}
func (st *SymbolTable) EvaluateBinaryExpression(node ast.BinaryExpression) (Class, error) {
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
	err = ShouldOperate(node.Operator, left, right)
	if err != nil {
		return nil, WrappedNodeError(node, err)
	}
	if node.Operator.IsComparableOperator() {
		return Boolean, nil
	}
	return left, nil
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
		return nil, StandardError(UnknownProperty, "%s has no property %s", object.Name(), member)
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
		return nil, StandardError(UnknownProperty, "%s has no property %s", object.Class().Name(), member)
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
		return nil, StandardError(UnknownProperty, "%s has no property %s", object.Name(), member)
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
		return nil, StandardError(UnknownProperty, "%s has no property %s", object.Class().Name(), member)
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
func evaluateValueExpressionCallMember(st *SymbolTable, value ScopeValue, node ast.CallExpression) (Class, error) {
	switch object := value.(type) {
	case Callable:
		callableArgs := object.Arguments()
		if len(node.Arguments) != len(callableArgs) {
			return nil, StandardError(InvalidArgumentLength, "expected %d arguments, got %d", len(callableArgs), len(node.Arguments))
		}
		for idx, arg := range node.Arguments {
			argClass, err := st.EvaluateExpression(arg)
			if err != nil {
				return nil, err
			}
			err = ShouldConstruct(callableArgs[idx], argClass)
			if err != nil {
				return nil, WrappedNodeError(node, err)
			}
		}
		return object.Returns(), nil
	case Class:
		if len(node.Arguments) != 1 {
			return nil, StandardError(InvalidClassConstruction, "invalid class construction")
		}
		argClass, err := st.EvaluateExpression(node.Arguments[0])
		if err != nil {
			return nil, err
		}
		err = ShouldConstruct(object, argClass)
		if err != nil {
			return nil, WrappedNodeError(node, err)
		}
		return object, nil
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
				return nil, WrappedNodeError(node, err)
			}
		case ast.IndexExpression:
			currentValueObject, ok := current.(ValueObject)
			if !ok {
				return nil, NodeError(node, InvalidIndexTarget, "cannot take index of non-enumerable %T", current)
			}
			enumerable := currentValueObject.Class().Descriptors().Enumerable
			if enumerable == nil {
				return nil, NodeError(node, InvalidIndexTarget, "cannot take index of non-enumerable class %s", currentValueObject.Class().Name())
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
func (st *SymbolTable) EvaluateValueExpression(node ast.ValueExpression) (Class, error) {
	firstMember, ok := node.Members[0].Init.(string)
	if !ok {
		return nil, NodeError(node, InvalidValueExpression, "invalid value expression")
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
				return nil, NodeError(node, InvalidIndexTarget, "cannot take index of non-enumerable %T", current)
			}
			enumerable := currentClass.Descriptors().Enumerable
			if enumerable == nil {
				return nil, NodeError(node, InvalidIndexTarget, "cannot take index of non-enumerable class %s", currentClass.Name())
			}
			err = st.EvaluateIndexExpression(member, currentClass)
		}
		if err != nil {
			return nil, WrappedNodeError(node, err)
		}
	}
	class, ok := current.(Class)
	if !ok {
		return nil, NodeError(node, InvalidValueExpression, "invalid value expression")
	}
	return class, nil
}

func (st *SymbolTable) ResolveIndexExpression(node ast.IndexExpression, target ValueObject) (int, int, error) {
	enumerable := target.Class().Descriptors().Enumerable
	if enumerable == nil {
		return -1, -1, NodeError(node, InvalidIndexTarget, "cannot take index of non-enumerable class %s", target.Class().Name())
	}
	startIndex := 0
	if node.Left != nil {
		leftValue, err := st.ResolveExpression(*node.Left)
		if err != nil {
			return -1, -1, err
		}
		leftIntValue, ok := leftValue.(IntegerValue)
		if !ok {
			return -1, -1, NodeError(node, InvalidIndex, "index must be an Integer, got %s", leftValue.Class().Name())
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
				return -1, -1, NodeError(node.Right, InvalidIndex, "index must be an Integer, got %s", rightValue.Class().Name())
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
		return NodeError(node, InvalidIndexTarget, "cannot take index of non-enumerable class %s", target.Name())
	}
	if node.Left != nil {
		leftClass, err := st.EvaluateExpression(*node.Left)
		if err != nil {
			return err
		}
		if _, ok := leftClass.(IntegerClass); !ok {
			return NodeError(node, InvalidIndex, "index must be an Integer, got %s", leftClass.Name())
		}
	}
	if node.Right != nil {
		rightClass, err := st.EvaluateExpression(*node.Right)
		if err != nil {
			return err
		}
		if _, ok := rightClass.(IntegerClass); !ok {
			return NodeError(node, InvalidIndex, "index must be an Integer, got %s", rightClass.Name())
		}
	}
	return nil
}

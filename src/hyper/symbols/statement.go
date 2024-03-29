package symbols

import (
	"github.com/hntrl/hyper/src/hyper/ast"
	. "github.com/hntrl/hyper/src/hyper/symbols/errors"
	"github.com/hntrl/hyper/src/hyper/tokens"
)

func validateSecondaryTargetForTryStatement(st *SymbolTable, secondary *string, init ast.Node) error {
	if secondary != nil {
		if _, ok := init.(ast.TryStatement); !ok {
			return NodeError(init, InvalidSyntaxTree, "secondary target in declaration without try statement")
		}
		if secondary, ok := st.Local[*secondary]; ok {
			switch secondaryValue := secondary.(type) {
			case *ExpectedValueObject:
				if err := ShouldConstruct(NewNilableClass(Error), secondaryValue.Class); err != nil {
					return err
				}
			case ValueObject:
				if err := ShouldConstruct(NewNilableClass(Error), secondaryValue.Class()); err != nil {
					return err
				}
			default:
				return StandardError(InvalidSecondaryTarget, "secondary target must be a value object")
			}
		}
	}
	return nil
}

func (st *SymbolTable) ResolveDeclarationStatement(node ast.DeclarationStatement) error {
	if st.Immutable[node.Target] != nil {
		return NodeError(node, CannotReassignImmutableValue, "cannot reassign immutable value %s", node.Target)
	}
	if st.Local[node.Target] != nil {
		return NodeError(node, CannotRedeclareValue, "cannot redeclare value %s", node.Target)
	}
	err := validateSecondaryTargetForTryStatement(st, node.SecondaryTarget, node.Init)
	if err != nil {
		return WrappedNodeError(node, err)
	}
	switch expr := node.Init.(type) {
	case ast.Expression:
		value, err := st.ResolveExpression(expr)
		if err != nil {
			return err
		}
		st.Local[node.Target] = value
	case ast.TryStatement:
		value, err := st.ResolveExpression(expr.Init)
		if node.SecondaryTarget != nil {
			if err != nil {
				st.Local[*node.SecondaryTarget] = NewNilableValue(Error, ErrorValue{
					Name:    "Error",
					Message: err.Error(),
				})
				return nil
			} else {
				st.Local[*node.SecondaryTarget] = NewNilableValue(Error, nil)
			}
		}
		st.Local[node.Target] = value
	default:
		return NodeError(node, InvalidSyntaxTree, "invalid declaration statement")
	}
	return nil
}
func (st *SymbolTable) EvaluateDeclarationStatement(node ast.DeclarationStatement) error {
	if st.Immutable[node.Target] != nil {
		return NodeError(node, CannotReassignImmutableValue, "cannot reassign immutable value %s", node.Target)
	}
	if st.Local[node.Target] != nil {
		return NodeError(node, CannotRedeclareValue, "cannot redeclare value %s", node.Target)
	}
	err := validateSecondaryTargetForTryStatement(st, node.SecondaryTarget, node.Init)
	if err != nil {
		return WrappedNodeError(node, err)
	}
	switch expr := node.Init.(type) {
	case ast.Expression:
		if node.SecondaryTarget != nil {
			return NodeError(node, InvalidSecondaryTarget, "secondary target in declaration without try statement")
		}
		value, err := st.EvaluateExpression(expr)
		if err != nil {
			return err
		}
		st.Local[node.Target] = value
	case ast.TryStatement:
		value, err := st.EvaluateExpression(expr.Init)
		if err != nil {
			return err
		}
		if node.SecondaryTarget != nil {
			st.Local[*node.SecondaryTarget] = &ExpectedValueObject{NewNilableClass(Error)}
		}
		st.Local[node.Target] = value
	default:
		return NodeError(node, InvalidSyntaxTree, "invalid declaration statement")
	}
	return nil
}

type assignmentStatementValuePredicateResolver func(ValueObject) (ValueObject, error)
type assignmentStatementValuePredicateEvaluator func(Class) error

func resolveAssignmentStatementWithSingleMember(st *SymbolTable, memberNode ast.AssignmentTargetExpressionMember, current ValueObject, operandPredicate assignmentStatementValuePredicateResolver) error {
	descriptors := current.Class().Descriptors()
	switch memberInit := memberNode.Init.(type) {
	case string:
		properties := descriptors.Properties
		if properties == nil {
			return StandardError(CannotSetProperty, "cannot set property %s on %s", memberInit, current.Class().Descriptors().Name)
		}
		property := properties[memberInit]
		if property.Setter == nil {
			return StandardError(CannotSetProperty, "cannot set immutable property %s on %s", memberInit, current.Class().Descriptors().Name)
		}
		currentValue, err := property.Getter(current)
		if err != nil {
			return err
		}
		operand, err := operandPredicate(currentValue)
		if err != nil {
			return err
		}
		return property.Setter(current, operand)
	case ast.IndexExpression:
		enumerable := descriptors.Enumerable
		if enumerable == nil {
			return StandardError(InvalidAssignmentTarget, "cannot set index on non-enumerable class %s", current.Class().Descriptors().Name)
		}
		startIndex, endIndex, err := st.ResolveIndexExpression(memberInit, current)
		if err != nil {
			return err
		}
		if memberInit.IsRange {
			currentValue, err := enumerable.GetRange(current, startIndex, endIndex)
			if err != nil {
				return err
			}
			operand, err := operandPredicate(currentValue)
			if err != nil {
				return err
			}
			return enumerable.SetRange(current, startIndex, endIndex, operand)
		} else {
			currentValue, err := enumerable.GetIndex(current, startIndex)
			if err != nil {
				return err
			}
			operand, err := operandPredicate(currentValue)
			if err != nil {
				return err
			}
			return enumerable.SetIndex(current, startIndex, operand)
		}
	default:
		return nil
	}
}
func evaluateAssignmentStatementWithSingleMember(st *SymbolTable, memberNode ast.AssignmentTargetExpressionMember, current *ExpectedValueObject, operandValidator assignmentStatementValuePredicateEvaluator) error {
	descriptors := current.Class.Descriptors()
	switch memberInit := memberNode.Init.(type) {
	case string:
		properties := descriptors.Properties
		if properties == nil {
			return StandardError(CannotSetProperty, "cannot set property %s on %s", memberInit, current.Class.Descriptors().Name)
		}
		property := properties[memberInit]
		if property.Setter == nil {
			return StandardError(CannotSetProperty, "cannot set immutable property %s on %s", memberInit, current.Class.Descriptors().Name)
		}
		return operandValidator(property.PropertyClass)
	case ast.IndexExpression:
		enumerable := descriptors.Enumerable
		if enumerable == nil {
			return StandardError(InvalidAssignmentTarget, "cannot set index on non-enumerable class %s", current.Class.Descriptors().Name)
		}
		err := st.EvaluateIndexExpression(memberInit, current)
		if err != nil {
			return err
		}
		if memberInit.IsRange {
			return operandValidator(current.Class)
		} else {
			// TODO: this should be fixed: right now the interpreter has no concept of what the class type will be returned when calling GetIndex. that class should be used since its safe to say the return class of GetRange will always equal that of the parent class, but not for GetIndex.
			// use case: String[1] = "a" - constructs by String("a"); []Item[1] = "a" - constructs by Item("a") (not []Item("a"))
			return operandValidator(current.Class)
		}
	default:
		return nil
	}
}

func resolveAssignmentStatementForMembers(st *SymbolTable, members []ast.AssignmentTargetExpressionMember, current ValueObject, operandPredicate assignmentStatementValuePredicateResolver) error {
	if len(members) > 1 {
		descriptors := current.Class().Descriptors()
		switch memberNode := members[0].Init.(type) {
		case string:
			properties := descriptors.Properties
			if properties == nil {
				return StandardError(CannotSetProperty, "cannot set property %s on %s", memberNode, current.Class().Descriptors().Name)
			}
			propertyValue, err := properties[memberNode].Getter(current)
			if err != nil {
				return err
			}
			return resolveAssignmentStatementForMembers(st, members[1:], propertyValue, operandPredicate)
		case ast.IndexExpression:
			enumerable := descriptors.Enumerable
			if enumerable == nil {
				return StandardError(InvalidAssignmentTarget, "cannot set index on non-enumerable class %s", current.Class().Descriptors().Name)
			}
			startIndex, endIndex, err := st.ResolveIndexExpression(memberNode, current)
			if err != nil {
				return err
			}
			if memberNode.IsRange {
				if endIndex > startIndex {
					return StandardError(InvalidAssignmentTarget, "end index cannot be greater than start index in assignment")
				}
				for i := startIndex; i <= endIndex; i++ {
					indexedValue, err := enumerable.GetIndex(current, i)
					if err != nil {
						return err
					}
					err = resolveAssignmentStatementForMembers(st, members[1:], indexedValue, operandPredicate)
					if err != nil {
						return err
					}
				}
				return nil
			} else {
				indexedValue, err := enumerable.GetIndex(current, startIndex)
				if err != nil {
					return err
				}
				return resolveAssignmentStatementForMembers(st, members[1:], indexedValue, operandPredicate)
			}
		default:
			return nil
		}
	} else {
		return resolveAssignmentStatementWithSingleMember(st, members[0], current, operandPredicate)
	}
}
func evaluateAssignmentStatementForMembers(st *SymbolTable, members []ast.AssignmentTargetExpressionMember, current *ExpectedValueObject, operandValidator assignmentStatementValuePredicateEvaluator) error {
	if len(members) > 1 {
		descriptors := current.Class.Descriptors()
		switch memberNode := members[0].Init.(type) {
		case string:
			properties := descriptors.Properties
			if properties == nil {
				return StandardError(CannotSetProperty, "cannot set property %s on %s", memberNode, current.Class.Descriptors().Name)
			}
			propertyValue := &ExpectedValueObject{properties[memberNode].PropertyClass}
			return evaluateAssignmentStatementForMembers(st, members[1:], propertyValue, operandValidator)
		case ast.IndexExpression:
			enumerable := descriptors.Enumerable
			if enumerable == nil {
				return StandardError(InvalidAssignmentTarget, "cannot set index on non-enumerable class %s", current.Class.Descriptors().Name)
			}
			err := st.EvaluateIndexExpression(memberNode, current)
			if err != nil {
				return err
			}
			return evaluateAssignmentStatementForMembers(st, members[1:], current, operandValidator)
		default:
			return nil
		}
	} else {
		return evaluateAssignmentStatementWithSingleMember(st, members[0], current, operandValidator)
	}
}

func (st *SymbolTable) ResolveAssignmentStatement(node ast.AssignmentStatement) error {
	firstMember, ok := node.Target.Members[0].Init.(string)
	if !ok {
		return NodeError(node, InvalidSyntaxTree, "invalid assignment statement target")
	}
	if _, ok := st.Immutable[firstMember]; ok {
		return NodeError(node.Target, CannotReassignImmutableValue, "cannot reassign immutable value %s", firstMember)
	}
	scopeValue := st.Local[firstMember]
	if scopeValue == nil {
		return NodeError(node.Target, UnknownSelector, "unknown selector %s", firstMember)
	}
	currentValue, ok := scopeValue.(ValueObject)
	if !ok {
		return NodeError(node.Target, InvalidAssignmentTarget, "assignment target must be a value object")
	}
	err := validateSecondaryTargetForTryStatement(st, node.SecondaryTarget, node.Init)
	if err != nil {
		return WrappedNodeError(node, err)
	}

	var operand ValueObject
	switch expr := node.Init.(type) {
	case ast.Expression:
		operand, err = st.ResolveExpression(expr)
		if err != nil {
			return err
		}
	case ast.TryStatement:
		operand, err = st.ResolveExpression(expr.Init)
		if node.SecondaryTarget != nil {
			if err != nil {
				st.Local[*node.SecondaryTarget] = NewNilableValue(Error, ErrorValue{
					Name:    "Error",
					Message: err.Error(),
				})
				return nil
			} else {
				st.Local[*node.SecondaryTarget] = NewNilableValue(Error, nil)
			}
		}
	}

	if len(node.Target.Members) > 1 {
		effectOperator := tokens.GetEffectOperator(node.Operator)
		operandPredicate := func(currentValue ValueObject) (ValueObject, error) {
			if node.Operator == tokens.ASSIGN {
				return Construct(currentValue.Class(), operand)
			} else {
				return Operate(effectOperator, currentValue, operand)
			}
		}
		err := resolveAssignmentStatementForMembers(st, node.Target.Members[1:], currentValue, operandPredicate)
		if err != nil {
			return WrappedNodeError(node.Target, err)
		}
	} else {
		constructedValue, err := Construct(currentValue.Class(), operand)
		if err != nil {
			return WrappedNodeError(node.Target, err)
		}
		st.Local[firstMember] = constructedValue
	}
	return nil
}
func (st *SymbolTable) EvaluateAssignmentStatement(node ast.AssignmentStatement) error {
	firstMember, ok := node.Target.Members[0].Init.(string)
	if !ok {
		return NodeError(node, InvalidSyntaxTree, "invalid assignment statement target")
	}
	if _, ok := st.Immutable[firstMember]; ok {
		return NodeError(node.Target, CannotReassignImmutableValue, "cannot reassign immutable value %s", firstMember)
	}
	scopeValue := st.Local[firstMember]
	if scopeValue == nil {
		return NodeError(node.Target, UnknownSelector, "unknown selector %s", firstMember)
	}
	currentValue, ok := scopeValue.(*ExpectedValueObject)
	if !ok {
		return NodeError(node.Target, InvalidAssignmentTarget, "assignment target must be a value object")
	}
	err := validateSecondaryTargetForTryStatement(st, node.SecondaryTarget, node.Init)
	if err != nil {
		return WrappedNodeError(node, err)
	}

	var operand *ExpectedValueObject
	switch expr := node.Init.(type) {
	case ast.Expression:
		operand, err = st.EvaluateExpression(expr)
		if err != nil {
			return err
		}
	case ast.TryStatement:
		operand, err = st.EvaluateExpression(expr.Init)
		if err != nil {
			return err
		}
		if node.SecondaryTarget != nil {
			st.Local[*node.SecondaryTarget] = &ExpectedValueObject{NewNilableClass(Error)}
		}
	}

	if len(node.Target.Members) > 1 {
		effectOperator := tokens.GetEffectOperator(node.Operator)
		operandValidator := func(currentClass Class) error {
			if node.Operator == tokens.ASSIGN {
				return ShouldConstruct(currentClass, operand.Class)
			} else {
				return ShouldOperate(effectOperator, currentClass, operand.Class)
			}
		}
		err := evaluateAssignmentStatementForMembers(st, node.Target.Members[1:], currentValue, operandValidator)
		if err != nil {
			return WrappedNodeError(node.Target, err)
		}
	} else {
		err := ShouldConstruct(currentValue.Class, operand.Class)
		if err != nil {
			return WrappedNodeError(node.Target, err)
		}
	}
	return nil
}

func (st *SymbolTable) ResolveIfStatement(node ast.IfStatement) (ValueObject, error) {
	condition, err := st.ResolveExpression(node.Condition)
	if err != nil {
		return nil, err
	}
	conditionResult, ok := condition.(BooleanValue)
	if !ok {
		return nil, NodeError(node.Condition, InvalidIfCondition, "if condition must be a boolean")
	}
	if conditionResult {
		return st.ResolveBlock(node.Body)
	}
	if node.Alternate != nil {
		switch alt := node.Alternate.(type) {
		case ast.IfStatement:
			return st.ResolveIfStatement(alt)
		case ast.Block:
			return st.ResolveBlock(alt)
		default:
			return nil, NodeError(alt, InvalidSyntaxTree, "inavlid if condition alternate")
		}
	}
	return nil, nil
}
func (st *SymbolTable) EvaluateIfStatement(node ast.IfStatement, shouldReturn Class) (bool, error) {
	condition, err := st.EvaluateExpression(node.Condition)
	if err != nil {
		return false, err
	}
	if _, ok := condition.Class.(BooleanClass); !ok {
		return false, NodeError(node.Condition, InvalidIfCondition, "if condition must be a boolean")
	}
	blockReturns, err := st.EvaluateBlock(node.Body, shouldReturn)
	if err != nil {
		return false, err
	}
	if node.Alternate != nil {
		switch alt := node.Alternate.(type) {
		case ast.IfStatement:
			altStatementReturns, err := st.EvaluateIfStatement(alt, shouldReturn)
			if err != nil {
				return false, err
			}
			return blockReturns && altStatementReturns, nil
		case ast.Block:
			elseBlockReturns, err := st.EvaluateBlock(alt, shouldReturn)
			if err != nil {
				return false, err
			}
			return blockReturns && elseBlockReturns, nil
		default:
			return false, NodeError(alt, InvalidSyntaxTree, "invalid if condition alternate")
		}
	}
	return false, nil
}

func (st *SymbolTable) ResolveWhileStatement(node ast.WhileStatement) (ValueObject, error) {
	scopeTable := st.StartLoop()
loopBlock:
	for {
		condition, err := st.ResolveExpression(node.Condition)
		if err != nil {
			return nil, err
		}
		conditionResult, ok := condition.(BooleanValue)
		if !ok {
			return nil, NodeError(node.Condition, InvalidWhileCondition, "while loop condition must be a boolean")
		}
		if !conditionResult {
			break
		}
		obj, err := st.ResolveBlock(node.Body)
		if err != nil {
			return nil, err
		}
		if obj != nil {
			return obj, nil
		}
		if scopeTable.LoopState.ShouldContinue {
			continue loopBlock
		}
		if scopeTable.LoopState.ShouldBreak {
			break loopBlock
		}
	}
	return nil, nil
}
func (st *SymbolTable) EvaluateWhileStatement(node ast.WhileStatement, shouldReturn Class) (bool, error) {
	condition, err := st.EvaluateExpression(node.Condition)
	if err != nil {
		return false, err
	}
	if _, ok := condition.Class.(BooleanClass); !ok {
		return false, NodeError(node.Condition, InvalidWhileCondition, "while loop condition must be a Boolean")
	}
	scopeTable := st.StartLoop()
	return scopeTable.EvaluateBlock(node.Body, shouldReturn)
}

func resolveForStatementWithForCondition(st *SymbolTable, conditionBlock ast.ForCondition, block ast.Block) (ValueObject, error) {
	scopeTable := st.StartLoop()
loopBlock:
	for {
		if conditionBlock.Init != nil {
			err := scopeTable.ResolveDeclarationStatement(*conditionBlock.Init)
			if err != nil {
				return nil, err
			}
		}
		condition, err := scopeTable.ResolveExpression(conditionBlock.Condition)
		if err != nil {
			return nil, err
		}
		conditionResult, ok := condition.(BooleanValue)
		if !ok {
			return nil, NodeError(conditionBlock.Condition, InvalidForCondition, "for loop condition must be a Boolean")
		}
		if !conditionResult {
			break loopBlock
		}
		obj, err := st.ResolveBlock(block)
		if err != nil {
			return nil, err
		}
		if obj != nil {
			return obj, nil
		}
		if scopeTable.LoopState.ShouldContinue {
			continue loopBlock
		}
		if scopeTable.LoopState.ShouldBreak {
			break loopBlock
		}
		switch updateNode := conditionBlock.Update.(type) {
		case ast.Expression:
			_, err := scopeTable.ResolveExpression(updateNode)
			if err != nil {
				return nil, err
			}
		case ast.AssignmentStatement:
			err := scopeTable.ResolveAssignmentStatement(updateNode)
			if err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}
func evaluateForStatementWithForCondition(st *SymbolTable, conditionBlock ast.ForCondition, block ast.Block, shouldReturn Class) (bool, error) {
	scopeTable := st.StartLoop()
	if conditionBlock.Init != nil {
		err := scopeTable.EvaluateDeclarationStatement(*conditionBlock.Init)
		if err != nil {
			return false, err
		}
	}
	condition, err := scopeTable.EvaluateExpression(conditionBlock.Condition)
	if err != nil {
		return false, err
	}
	if _, ok := condition.Class.(BooleanClass); !ok {
		return false, NodeError(conditionBlock.Condition, InvalidForCondition, "for loop condition must be a Boolean")
	}
	switch updateNode := conditionBlock.Update.(type) {
	case ast.Expression:
		_, err = scopeTable.EvaluateExpression(updateNode)
		if err != nil {
			return false, err
		}
	case ast.AssignmentStatement:
		err = scopeTable.EvaluateAssignmentStatement(updateNode)
		if err != nil {
			return false, err
		}
	}
	return scopeTable.EvaluateBlock(block, shouldReturn)
}

func resolveForStatementWithRangeCondition(st *SymbolTable, conditionBlock ast.RangeCondition, block ast.Block) (ValueObject, error) {
	target, err := st.ResolveExpression(conditionBlock.Target)
	if err != nil {
		return nil, err
	}
	// TODO: use enumerable instead of array value
	arr, ok := target.(*ArrayValue)
	if !ok {
		return nil, NodeError(conditionBlock.Target, CannotEnumerate, "%s is not enumerable", target.Class().Descriptors().Name)
	}
	scopeTable := st.StartLoop()
loopBlock:
	for idx, item := range arr.items {
		scopeTable.Local[conditionBlock.Index] = IntegerValue(idx)
		scopeTable.Local[conditionBlock.Value] = item
		obj, err := st.ResolveBlock(block)
		if err != nil {
			return nil, err
		}
		if obj != nil {
			return obj, nil
		}
		if scopeTable.LoopState.ShouldContinue {
			continue loopBlock
		}
		if scopeTable.LoopState.ShouldBreak {
			break loopBlock
		}
	}
	return nil, nil
}
func evaluateForStatementWithRangeCondition(st *SymbolTable, conditionBlock ast.RangeCondition, block ast.Block, shouldReturn Class) (bool, error) {
	target, err := st.EvaluateExpression(conditionBlock.Target)
	if err != nil {
		return false, err
	}
	arrayClass, ok := target.Class.(ArrayClass)
	if !ok {
		return false, NodeError(conditionBlock.Target, CannotEnumerate, "%s is not enumerable", target.Class.Descriptors().Name)
	}
	scopeTable := st.StartLoop()
	scopeTable.Local[conditionBlock.Index] = &ExpectedValueObject{Integer}
	scopeTable.Local[conditionBlock.Value] = &ExpectedValueObject{arrayClass.itemClass}
	return scopeTable.EvaluateBlock(block, shouldReturn)
}

func (st *SymbolTable) ResolveForStatement(node ast.ForStatement) (ValueObject, error) {
	switch conditionBlock := node.Condition.(type) {
	case ast.ForCondition:
		return resolveForStatementWithForCondition(st, conditionBlock, node.Body)
	case ast.RangeCondition:
		return resolveForStatementWithRangeCondition(st, conditionBlock, node.Body)
	}
	return nil, nil
}
func (st *SymbolTable) EvaluateForStatement(node ast.ForStatement, shouldReturn Class) (bool, error) {
	switch conditionBlock := node.Condition.(type) {
	case ast.ForCondition:
		return evaluateForStatementWithForCondition(st, conditionBlock, node.Body, shouldReturn)
	case ast.RangeCondition:
		return evaluateForStatementWithRangeCondition(st, conditionBlock, node.Body, shouldReturn)
	}
	return false, nil
}

func (st *SymbolTable) ResolveSwitchBlock(node ast.SwitchBlock) (ValueObject, error) {
	target, err := st.ResolveExpression(node.Target)
	if err != nil {
		return nil, err
	}
	comparators := target.Class().Descriptors().Comparators
	// TODO: see if the equals comparator exists
	if comparators == nil {
		return nil, NodeError(node.Target, InvalidSwitchTarget, "switch target %s is not operable", target.Class().Descriptors().Name)
	}
	resolved := false
	for _, caseBlock := range node.Statements {
		if caseBlock.IsDefault {
			continue
		}
		caseCondition, err := st.ResolveExpression(*caseBlock.Condition)
		if err != nil {
			return nil, err
		}
		conditionPassed, err := Compare(tokens.EQUALS, target, caseCondition)
		if err != nil {
			return nil, WrappedNodeError(caseBlock.Condition, err)
		}
		if conditionPassed {
			resolved = true
			returnObject, err := st.ResolveBlock(caseBlock.Body)
			if err != nil {
				return nil, err
			}
			if returnObject != nil {
				return returnObject, nil
			}
		}
	}
	if !resolved {
		for _, caseBlock := range node.Statements {
			if !caseBlock.IsDefault {
				continue
			}
			returnObject, err := st.ResolveBlock(caseBlock.Body)
			if err != nil {
				return nil, err
			}
			if returnObject != nil {
				return returnObject, nil
			}
		}
	}
	return nil, nil
}
func (st *SymbolTable) EvaluateSwitchBlock(node ast.SwitchBlock, shouldReturn Class) (bool, error) {
	target, err := st.EvaluateExpression(node.Target)
	if err != nil {
		return false, err
	}
	comparators := target.Class.Descriptors().Comparators
	if comparators == nil {
		return false, NodeError(node.Target, InvalidSwitchTarget, "switch target %s is not operable", target.Class.Descriptors().Name)
	}
	defaultBlockReturns := false
	hasDefaultBlock := false
	for _, caseBlock := range node.Statements {
		if caseBlock.IsDefault {
			if hasDefaultBlock {
				return false, NodeError(caseBlock, DuplicateDefaultSwitchStatements, "switch statement can only have one default block")
			}
			hasDefaultBlock = true
			defaultBlockReturns, err = st.EvaluateBlock(caseBlock.Body, shouldReturn)
			if err != nil {
				return false, err
			}
		} else {
			caseCondition, err := st.EvaluateExpression(*caseBlock.Condition)
			if err != nil {
				return false, err
			}
			err = ShouldCompare(tokens.EQUALS, target.Class, caseCondition.Class)
			if err != nil {
				return false, WrappedNodeError(caseBlock.Condition, err)
			}
			_, err = st.EvaluateBlock(caseBlock.Body, shouldReturn)
			if err != nil {
				return false, err
			}
		}
	}
	return defaultBlockReturns, nil
}

func (st *SymbolTable) ResolveGuardStatement(node ast.GuardStatement) error {
	// TODO: get guard handler
	_, err := st.ResolveExpression(node.Init)
	if err != nil {
		return err
	}
	return nil
}
func (st *SymbolTable) EvaluateGuardStatement(node ast.GuardStatement) error {
	// TODO: get guard handler
	_, err := st.EvaluateExpression(node.Init)
	if err != nil {
		return err
	}
	return nil
}

func (st *SymbolTable) ResolveBlockStatement(node ast.BlockStatement) (returnValue ValueObject, err error) {
	switch node := node.Init.(type) {
	case ast.Expression:
		_, err = st.ResolveExpression(node)
	case ast.DeclarationStatement:
		err = st.ResolveDeclarationStatement(node)
	case ast.AssignmentStatement:
		err = st.ResolveAssignmentStatement(node)
	case ast.IfStatement:
		returnValue, err = st.ResolveIfStatement(node)
	case ast.WhileStatement:
		returnValue, err = st.ResolveWhileStatement(node)
	case ast.ForStatement:
		returnValue, err = st.ResolveForStatement(node)
	case ast.ContinueStatement:
		st.LoopState.ShouldContinue = true
	case ast.BreakStatement:
		st.LoopState.ShouldBreak = true
	case ast.SwitchBlock:
		returnValue, err = st.ResolveSwitchBlock(node)
	case ast.GuardStatement:
		err = st.ResolveGuardStatement(node)
	case ast.ReturnStatement:
		returnValue, err = st.ResolveExpression(node.Init)
	case ast.ThrowStatement:
		returnValue, err = st.ResolveExpression(node.Init)
		if err != nil {
			return nil, err
		}
		thrownError, ok := returnValue.(ErrorValue)
		if !ok {
			return nil, NodeError(node, InvalidThrowValue, "throw statement must be an Error, got %s", returnValue.Class().Descriptors().Name)
		}
		return nil, thrownError
	case ast.TryStatement:
		st.ResolveExpression(node.Init)
	default:
		return nil, NodeError(node, InvalidSyntaxTree, "unknown block statement type %T", node)
	}
	return returnValue, err
}
func (st *SymbolTable) EvaluateBlockStatement(node ast.BlockStatement, shouldReturn Class) (returns bool, err error) {
	switch node := node.Init.(type) {
	case ast.Expression:
		_, err = st.EvaluateExpression(node)
	case ast.DeclarationStatement:
		err = st.EvaluateDeclarationStatement(node)
	case ast.AssignmentStatement:
		err = st.EvaluateAssignmentStatement(node)
	case ast.IfStatement:
		returns, err = st.EvaluateIfStatement(node, shouldReturn)
	case ast.WhileStatement:
		returns, err = st.EvaluateWhileStatement(node, shouldReturn)
	case ast.ForStatement:
		returns, err = st.EvaluateForStatement(node, shouldReturn)
	case ast.ContinueStatement:
		if st.LoopState == nil {
			return false, NodeError(node, BadLoopControlStatement, "continue statement outside loop")
		}
	case ast.BreakStatement:
		if st.LoopState == nil {
			return false, NodeError(node, BadLoopControlStatement, "break statement outside loop")
		}
	case ast.SwitchBlock:
		returns, err = st.EvaluateSwitchBlock(node, shouldReturn)
	case ast.GuardStatement:
		err = st.EvaluateGuardStatement(node)
	case ast.ReturnStatement:
		returned, err := st.EvaluateExpression(node.Init)
		if err != nil {
			return false, err
		}
		if shouldReturn != nil {
			err := ShouldConstruct(shouldReturn, returned.Class)
			if err != nil {
				if returned.Class == nil {
					return false, NodeError(node, InvalidReturnType, "should return %s, got <nil>", shouldReturn.Descriptors().Name)
				}
				return false, NodeError(node, InvalidReturnType, "should return %s, got %s", shouldReturn.Descriptors().Name, returned.Class.Descriptors().Name)
			}
		}
		return true, nil
	case ast.ThrowStatement:
		returned, err := st.EvaluateExpression(node.Init)
		if err != nil {
			return false, err
		}
		if _, ok := returned.Class.(ErrorClass); !ok {
			return false, NodeError(node, InvalidThrowValue, "throw statement must be an Error, got %s", returned.Class.Descriptors().Name)
		}
		return true, nil
	case ast.TryStatement:
		_, err := st.EvaluateExpression(node.Init)
		if err != nil {
			return false, err
		}
	default:
		return false, NodeError(node, InvalidSyntaxTree, "unknown block statement type %T", node)
	}
	return returns, err
}

func (st *SymbolTable) ResolveBlock(node ast.Block) (ValueObject, error) {
	scopeTable := st.Clone()
	for _, stmt := range node.Statements {
		returnObject, err := scopeTable.ResolveBlockStatement(stmt)
		if err != nil {
			return nil, err
		}
		if returnObject != nil {
			return returnObject, nil
		}
	}
	return nil, nil
}
func (st *SymbolTable) EvaluateBlock(node ast.Block, shouldReturn Class) (returns bool, err error) {
	scopeTable := st.Clone()
	for _, stmt := range node.Statements {
		returns, err = scopeTable.EvaluateBlockStatement(stmt, shouldReturn)
		if err != nil {
			return false, err
		}
	}
	return returns, err
}

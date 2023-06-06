package symbols

import (
	"github.com/hntrl/hyper/internal/ast"
	"github.com/hntrl/hyper/internal/tokens"
)

func (st *SymbolTable) ResolveDeclarationStatement(node ast.DeclarationStatement) error {
	if st.Immutable[node.Name] != nil {
		return CannotReassignImmutableValueError(node, node.Name)
	}
	if st.Local[node.Name] != nil {
		return CannotRedeclareValueError(node)
	}
	obj, err := st.ResolveExpression(node.Init)
	if err != nil {
		return err
	}
	st.Local[node.Name] = obj
	return nil
}
func (st *SymbolTable) EvaluateDeclarationStatement(node ast.DeclarationStatement) error {
	if st.Immutable[node.Name] != nil {
		return CannotReassignImmutableValueError(node, node.Name)
	}
	if st.Local[node.Name] != nil {
		return CannotRedeclareValueError(node)
	}
	class, err := st.EvaluateExpression(node.Init)
	if err != nil {
		return err
	}
	st.Local[node.Name] = class
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
			return CannotSetPropertyError(current, memberInit)
		}
		property := properties[memberInit]
		if property.Setter == nil {
			return CannotSetImmutablePropertyError(memberInit)
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
			return CannotSetNonEnumerableIndexError(current.Class())
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
func evaluateAssignmentStatementWithSingleMember(st *SymbolTable, memberNode ast.AssignmentTargetExpressionMember, current Class, operandValidator assignmentStatementValuePredicateEvaluator) error {
	descriptors := current.Descriptors()
	switch memberInit := memberNode.Init.(type) {
	case string:
		properties := descriptors.Properties
		if properties == nil {
			return CannotSetPropertyError(current, memberInit)
		}
		property := properties[memberInit]
		if property.Setter == nil {
			return CannotSetImmutablePropertyError(memberInit)
		}
		return operandValidator(property.PropertyClass)
	case ast.IndexExpression:
		enumerable := descriptors.Enumerable
		if enumerable == nil {
			return CannotSetNonEnumerableIndexError(current)
		}
		err := st.EvaluateIndexExpression(memberInit, current)
		if err != nil {
			return err
		}
		if memberInit.IsRange {
			return operandValidator(current)
		} else {
			// TODO: this should be fixed: right now the interpreter has no concept of what the class type will be returned when calling GetIndex. that class should be used since its safe to say the return class of GetRange will always equal that of the parent class, but not for GetIndex.
			// use case: String[1] = "a" - constructs by String("a"); []Item[1] = "a" - constructs by Item("a") (not []Item("a"))
			return operandValidator(current)
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
				return CannotSetPropertyError(current.Class(), memberNode)
			}
			propertyValue, err := properties[memberNode].Getter(current)
			if err != nil {
				return err
			}
			return resolveAssignmentStatementForMembers(st, members[1:], propertyValue, operandPredicate)
		case ast.IndexExpression:
			enumerable := descriptors.Enumerable
			if enumerable == nil {
				return CannotSetNonEnumerableIndexError(current.Class())
			}
			startIndex, endIndex, err := st.ResolveIndexExpression(memberNode, current)
			if err != nil {
				return err
			}
			if memberNode.IsRange {
				if endIndex > startIndex {
					return InvalidAssignmentIndicesError(memberNode)
				}
				for i := startIndex; i <= endIndex; i++ {
					indexedValue, err := enumerable.GetIndex(current, i)
					if err != nil {
						return err
					}
					err = resolveAssignmentStatementForMembers(st, members[1:], indexedValue, operandPredicate)
					if err != nil {
						return NodeError(memberNode, "problem when evaluting assignment at index %s: %s", i, err.Error())
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
func evaluateAssignmentStatementForMembers(st *SymbolTable, members []ast.AssignmentTargetExpressionMember, current Class, operandValidator assignmentStatementValuePredicateEvaluator) error {
	if len(members) > 1 {
		descriptors := current.Descriptors()
		switch memberNode := members[0].Init.(type) {
		case string:
			properties := descriptors.Properties
			if properties == nil {
				return CannotSetPropertyError(current, memberNode)
			}
			return evaluateAssignmentStatementForMembers(st, members[1:], properties[memberNode].PropertyClass, operandValidator)
		case ast.IndexExpression:
			enumerable := descriptors.Enumerable
			if enumerable == nil {
				return CannotSetNonEnumerableIndexError(current)
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
		return InvalidAssignmentStatementTargetError(node)
	}
	if _, ok := st.Immutable[firstMember]; ok {
		return CannotReassignImmutableValueError(node, firstMember)
	}
	scopeValue := st.Local[firstMember]
	if scopeValue == nil {
		return UnknownSelectorError(node.Target, firstMember)
	}
	currentValue, ok := scopeValue.(ValueObject)
	if !ok {
		return CannotAssignNonValueObjectError(node)
	}
	operand, err := st.ResolveExpression(node.Init)
	if err != nil {
		return err
	}
	if len(node.Target.Members) > 1 {
		effectOperator := tokens.GetEffectOperator(node.Operator)
		operandPredicate := func(currentValue ValueObject) (ValueObject, error) {
			if effectOperator == tokens.ASSIGN {
				return operand, nil
			} else {
				return Operate(effectOperator, currentValue, operand)
			}
		}
		err := resolveAssignmentStatementForMembers(st, node.Target.Members[1:], currentValue, operandPredicate)
		if err != nil {
			return err
		}
	} else {
		constructedValue, err := Construct(currentValue.Class(), operand)
		if err != nil {
			return err
		}
		st.Local[firstMember] = constructedValue
	}
	return nil
}
func (st *SymbolTable) EvaluateAssignmentStatement(node ast.AssignmentStatement) error {
	firstMember, ok := node.Target.Members[0].Init.(string)
	if !ok {
		return InvalidAssignmentStatementTargetError(node)
	}
	if _, ok := st.Immutable[firstMember]; ok {
		return CannotReassignImmutableValueError(node, firstMember)
	}
	scopeValue := st.Local[firstMember]
	if scopeValue == nil {
		return UnknownSelectorError(node.Target, firstMember)
	}
	currentClass, ok := scopeValue.(Class)
	if !ok {
		return CannotAssignNonValueObjectError(node)
	}
	operand, err := st.EvaluateExpression(node.Init)
	if err != nil {
		return err
	}
	if len(node.Target.Members) > 1 {
		effectOperator := tokens.GetEffectOperator(node.Operator)
		operandValidator := func(currentClass Class) error {
			if effectOperator == tokens.ASSIGN {
				return ShouldConstruct(currentClass, operand)
			} else {
				return ShouldOperate(effectOperator, currentClass, operand)
			}
		}
		err := evaluateAssignmentStatementForMembers(st, node.Target.Members[1:], currentClass, operandValidator)
		if err != nil {
			return err
		}
	} else {
		return ShouldConstruct(currentClass, operand)
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
		return nil, BadIfConditionError(node)
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
			return nil, BadIfAlternateError(node.Alternate)
		}
	}
	return nil, nil
}
func (st *SymbolTable) EvaluateIfStatement(node ast.IfStatement, shouldReturn Class) (bool, error) {
	conditionClass, err := st.EvaluateExpression(node.Condition)
	if err != nil {
		return false, err
	}
	if _, ok := conditionClass.(BooleanClass); !ok {
		return false, BadIfConditionError(node)
	}
	blockReturns, err := st.EvaluateBlock(node.Body, shouldReturn)
	if err != nil {
		return false, err
	}
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
			return nil, BadWhileConditionError(node)
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
	conditionClass, err := st.EvaluateExpression(node.Condition)
	if err != nil {
		return false, err
	}
	if _, ok := conditionClass.(BooleanClass); !ok {
		return false, BadWhileConditionError(node)
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
			return nil, BadForConditionError(conditionBlock.Condition)
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
	conditionClass, err := scopeTable.EvaluateExpression(conditionBlock.Condition)
	if err != nil {
		return false, err
	}
	if _, ok := conditionClass.(BooleanClass); !ok {
		return false, BadForConditionError(conditionBlock.Condition)
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
	arr, ok := target.(*ArrayValue)
	if !ok {
		return nil, NotIterableError(conditionBlock.Target, target)
	}
	scopeTable := st.StartLoop()
loopBlock:
	for idx, item := range arr.Items {
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
	targetClass, err := st.EvaluateExpression(conditionBlock.Target)
	if err != nil {
		return false, err
	}
	arrayClass, ok := targetClass.(ArrayClass)
	if !ok {
		return false, NotIterableError(conditionBlock.Target, targetClass)
	}
	scopeTable := st.StartLoop()
	scopeTable.Local[conditionBlock.Index] = Integer
	scopeTable.Local[conditionBlock.Value] = arrayClass.ItemClass
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
	if comparators == nil {
		return nil, InoperableSwitchTargetError(node, target)
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
			return nil, NodeError(caseBlock, err.Error())
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
	targetClass, err := st.EvaluateExpression(node.Target)
	if err != nil {
		return false, err
	}
	comparators := targetClass.Descriptors().Comparators
	if comparators == nil {
		return false, InoperableSwitchTargetError(node, targetClass)
	}
	defaultBlockReturns := false
	hasDefaultBlock := false
	for _, caseBlock := range node.Statements {
		if caseBlock.IsDefault {
			if hasDefaultBlock {
				return false, MultipleSwitchDefaultBlocksError(node)
			}
			hasDefaultBlock = true
			defaultBlockReturns, err = st.EvaluateBlock(caseBlock.Body, shouldReturn)
			if err != nil {
				return false, err
			}
		} else {
			caseConditionClass, err := st.EvaluateExpression(*caseBlock.Condition)
			if err != nil {
				return false, err
			}
			err = ShouldCompare(tokens.EQUALS, targetClass, caseConditionClass)
			if err != nil {
				return false, err
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

func (st *SymbolTable) ResolveBlockStatement(node ast.BlockStatement) (returnObject ValueObject, err error) {
	switch node := node.Init.(type) {
	case ast.Expression:
		_, err = st.ResolveExpression(node)
	case ast.DeclarationStatement:
		err = st.ResolveDeclarationStatement(node)
	case ast.AssignmentStatement:
		err = st.ResolveAssignmentStatement(node)
	case ast.IfStatement:
		returnObject, err = st.ResolveIfStatement(node)
	case ast.WhileStatement:
		returnObject, err = st.ResolveWhileStatement(node)
	case ast.ForStatement:
		returnObject, err = st.ResolveForStatement(node)
	case ast.ContinueStatement:
		st.LoopState.ShouldContinue = true
	case ast.BreakStatement:
		st.LoopState.ShouldBreak = true
	case ast.SwitchBlock:
		returnObject, err = st.ResolveSwitchBlock(node)
	case ast.GuardStatement:
		err = st.ResolveGuardStatement(node)
	case ast.ReturnStatement:
		returnObject, err = st.ResolveExpression(node.Init)
	case ast.ThrowStatement:
		returnObject, err = st.ResolveExpression(node.Init)
		thrownError, ok := returnObject.(ErrorValue)
		if !ok {
			return nil, InvalidThrowValueError(node)
		}
		return nil, thrownError
	default:
		return nil, UnknownBlockStatementError(node)
	}
	return returnObject, err
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
			return false, ContinueOutsideLoopError(node)
		}
	case ast.BreakStatement:
		if st.LoopState == nil {
			return false, BreakOutsideLoopError(node)
		}
	case ast.SwitchBlock:
		returns, err = st.EvaluateSwitchBlock(node, shouldReturn)
	case ast.GuardStatement:
		err = st.EvaluateGuardStatement(node)
	case ast.ReturnStatement:
		returnedClass, err := st.EvaluateExpression(node.Init)
		if err != nil {
			return false, err
		}
		if !classEquals(returnedClass, shouldReturn) {
			return false, MismatchedReturnClassError(node, returnedClass, shouldReturn)
		}
		return true, nil
	case ast.ThrowStatement:
		returnedClass, err := st.EvaluateExpression(node.Init)
		if err != nil {
			return false, err
		}
		if _, ok := returnedClass.(ErrorClass); !ok {
			return false, InvalidThrowValueError(node)
		}
		return true, nil
	default:
		return false, UnknownBlockStatementError(node)
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

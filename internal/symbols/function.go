package symbols

import (
	"fmt"

	"github.com/hntrl/hyper/internal/ast"
	"github.com/hntrl/hyper/internal/tokens"
)

type fnHandler func(args []ValueObject, proto ValueObject) (ValueObject, error)

// Function represents a subroutine that can be executed in an expression
type Function struct {
	ArgumentTypes []Class
	ReturnType    Class
	Handler       fnHandler
}

func (f Function) Get(key string) (Object, error) {
	return nil, nil
}

func (fn Function) Arguments() []Class {
	return fn.ArgumentTypes
}
func (fn Function) Returns() Class {
	return fn.ReturnType
}
func (fn Function) Call(args []ValueObject, proto ValueObject) (ValueObject, error) {
	if len(args) != len(fn.ArgumentTypes) {
		return nil, fmt.Errorf("expected %d arguments, got %d", len(fn.ArgumentTypes), len(args))
	}
	args, err := ResolveMethodArguments(fn, args)
	if err != nil {
		return nil, err
	}
	return fn.Handler(args, proto)
}

type FunctionOptions struct {
	Arguments []Class
	Returns   Class
	Handler   fnHandler
}

func NewFunction(opts FunctionOptions) Function {
	return Function{
		ArgumentTypes: opts.Arguments,
		ReturnType:    opts.Returns,
		Handler:       opts.Handler,
	}
}

func (st *SymbolTable) ResolveFunctionParameters(node ast.FunctionParameters) ([]Class, Class, error) {
	argumentTypes := make([]Class, 0)
	if node.Arguments.Items != nil {
		args, err := st.ResolveArgumentList(node.Arguments)
		if err != nil {
			return nil, nil, err
		}
		argumentTypes = args
	}
	var returnType Class
	if node.ReturnType != nil {
		returns, err := st.ResolveTypeExpression(*node.ReturnType)
		if err != nil {
			return nil, nil, err
		}
		returnType = returns
	}
	return argumentTypes, returnType, nil
}
func (st *SymbolTable) ResolveFunctionBlock(node ast.FunctionBlock, proto ValueObject) (*Function, error) {
	scopeTable := st.Clone()
	args, returns, err := scopeTable.ResolveFunctionParameters(node.Parameters)
	if err != nil {
		return nil, err
	}
	if proto != nil {
		scopeTable.immutable["self"] = proto
	}
	err = scopeTable.ValidateBlock(node.Body)
	if err != nil {
		return nil, err
	}
	if returns != nil {
		passes, err := scopeTable.ValidateBlockReturns(node.Body, returns)
		if err != nil {
			return nil, err
		}
		if !passes {
			return nil, NodeError(node, "expected return")
		}
	}
	return &Function{
		ArgumentTypes: args,
		ReturnType:    returns,
		Handler: func(args []ValueObject, proto ValueObject) (ValueObject, error) {
			execTable := st.Clone()
			err = execTable.ApplyArgumentList(node.Parameters.Arguments, args)
			if err != nil {
				return nil, err
			}
			if proto != nil {
				execTable.immutable["self"] = proto
			}
			obj, err := execTable.ResolveBlock(node.Body)
			if err != nil {
				return nil, err
			}
			if returns != nil {
				return Construct(returns, obj)
			}
			return nil, nil
		},
	}, nil
}

// ---
// FUNCTION ARGUMENTS
// ---

func (st *SymbolTable) ResolveArgumentList(expr ast.ArgumentList) ([]Class, error) {
	args := make([]Class, len(expr.Items))
	for idx, item := range expr.Items {
		switch arg := item.(type) {
		case ast.ArgumentItem:
			obj, err := st.ResolveTypeExpression(arg.Init)
			if err != nil {
				return nil, err
			}
			args[idx] = obj
			castedObject := Object(obj)
			st.local[arg.Key] = &castedObject
		case ast.ArgumentObject:
			typedObject := Type{Properties: make(map[string]Class)}
			for _, item := range arg.Items {
				obj, err := st.ResolveTypeExpression(item.Init)
				if err != nil {
					return nil, err
				}
				typedObject.Properties[item.Key] = obj
				castedObject := Object(obj)
				st.local[item.Key] = &castedObject
			}
			args[idx] = typedObject
		}
	}
	return args, nil
}

func ResolveMethodArguments(method Method, args []ValueObject) ([]ValueObject, error) {
	var err error
	methodArgs := method.Arguments()
	if len(args) != len(methodArgs) {
		return nil, fmt.Errorf("expected %d arguments, got %d", len(methodArgs), len(args))
	}
	for idx, arg := range args {
		argClass := methodArgs[idx]
		args[idx], err = Construct(argClass, arg)
		if err != nil {
			return nil, err
		}
	}
	return args, nil
}
func ValidateMethodArguments(method Method, args []Class) error {
	methodArgs := method.Arguments()
	if len(args) != len(methodArgs) {
		return fmt.Errorf("expected %d arguments, got %d", len(methodArgs), len(args))
	}
	for idx, arg := range args {
		err := ShouldConstruct(methodArgs[idx], arg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (st *SymbolTable) ApplyArgumentList(expr ast.ArgumentList, args []ValueObject) error {
	for idx, item := range expr.Items {
		switch argNode := item.(type) {
		case ast.ArgumentItem:
			castedObject := Object(args[idx])
			st.local[argNode.Key] = &castedObject
		case ast.ArgumentObject:
			for _, item := range argNode.Items {
				propObject, err := args[idx].Get(item.Key)
				if err != nil {
					return err
				}
				if propObject == nil {
					return NodeError(argNode, "object does not have property %s", item.Key)
				}
				castedObject := Object(propObject)
				st.local[item.Key] = &castedObject
			}
		}
	}
	return nil
}

// ---
// DECLARATION STATEMENTS
// ---

func (st *SymbolTable) ResolveDeclarationStatement(expr ast.DeclarationStatement, shouldEvaluate bool) error {
	if st.immutable[expr.Name] != nil {
		return NodeError(expr, "cannot reassign immutable variable %s", expr.Name)
	}
	if st.local[expr.Name] != nil {
		return NodeError(expr, "cannot redeclare variable %s", expr.Name)
	}
	if shouldEvaluate {
		obj, err := st.ResolveValueObject(expr.Init)
		if err != nil {
			return err
		}
		castedObject := Object(obj)
		st.local[expr.Name] = &castedObject
	} else {
		obj, err := st.ValidateExpression(expr.Init)
		if err != nil {
			return err
		}
		castedObject := Object(obj)
		st.local[expr.Name] = &castedObject
	}
	return nil
}

// ---
// ASSIGNMENT EXPRESSIONS
// ---

func getEffectOperator(token tokens.Token) tokens.Token {
	var effectOperator tokens.Token
	switch token {
	case tokens.ADD_ASSIGN, tokens.INC:
		effectOperator = tokens.ADD
	case tokens.SUB_ASSIGN, tokens.DEC:
		effectOperator = tokens.SUB
	case tokens.MUL_ASSIGN:
		effectOperator = tokens.MUL
	case tokens.PWR_ASSIGN:
		effectOperator = tokens.PWR
	case tokens.QUO_ASSIGN:
		effectOperator = tokens.QUO
	case tokens.REM_ASSIGN:
		effectOperator = tokens.REM
	}
	return effectOperator
}
func (st *SymbolTable) ResolveAssignmentExpression(expr ast.AssignmentExpression) error {
	if st.immutable[expr.Name.Members[0]] != nil {
		return NodeError(expr, "cannot reassign immutable variable %s", expr.Name.Members[0])
	}
	parentObject := st.local[expr.Name.Members[0]]
	if parentObject == nil {
		return UnknownSelector(expr.Name, expr.Name.Members[0])
	}
	originalObject, err := st.ResolveSelector(expr.Name)
	if err != nil {
		return err
	}
	object, ok := originalObject.(ValueObject)
	if !ok {
		return NodeError(expr, "cannot assign to non-value object")
	}
	operand, err := st.ResolveValueObject(expr.Init)
	if err != nil {
		return err
	}
	if expr.Operator == tokens.ASSIGN {
		object, err = Construct(object.Class(), operand)
		if err != nil {
			return err
		}
	} else {
		object, err = Operate(getEffectOperator(expr.Operator), object, operand)
		if err != nil {
			return NodeError(expr, err.Error())
		}
	}

	if len(expr.Name.Members) > 1 {
		var eval func(ValueObject, []string) error
		eval = func(current ValueObject, members []string) error {
			if len(members) == 1 {
				return current.Set(members[0], object)
			}
			prop, err := current.Get(members[0])
			if err != nil {
				return err
			}
			propObj, ok := prop.(ValueObject)
			if !ok {
				return fmt.Errorf("cannot set property on non-value object")
			}
			return eval(propObj, members[1:])
		}
		valueObj, ok := (*parentObject).(ValueObject)
		if !ok {
			return fmt.Errorf("cannot set property on non-value object")
		}
		err = eval(valueObj, expr.Name.Members[1:])
		if err != nil {
			return err
		}
		*st.local[expr.Name.Members[0]] = *parentObject
	} else {
		*st.local[expr.Name.Members[0]] = object
	}
	return nil
}
func (st *SymbolTable) ValidateAssignmentExpression(expr ast.AssignmentExpression) error {
	if st.immutable[expr.Name.Members[0]] != nil {
		return NodeError(expr, "cannot reassign immutable variable %s", expr.Name.Members[0])
	}
	parentObject := st.local[expr.Name.Members[0]]
	if parentObject == nil {
		return UnknownSelector(expr.Name, expr.Name.Members[0])
	}
	var err error
	currentObject := *parentObject
	for _, member := range expr.Name.Members[1:] {
		switch object := currentObject.(type) {
		case ObjectClass:
			if field := object.Fields()[member]; field != nil {
				currentObject = field
			} else {
				staticObject, err := object.Get(member)
				if err != nil {
					return err
				}
				if staticObject != nil {
					currentObject = staticObject
				} else {
					return NodeError(expr, "%s does not have field %s", object.ClassName(), member)
				}
			}
		case ValueObject:
			currentObject, err = object.Get(member)
			if err != nil {
				return err
			}
			if currentObject == nil {
				return NodeError(expr, "%s has no member %s", object.Class().ClassName(), member)
			}
		default:
			currentObject, err = object.Get(member)
			if err != nil {
				return err
			}
			if currentObject == nil {
				return NodeError(expr, "%T has no member %s", object, member)
			}
		}
	}

	if class, ok := currentObject.(Class); ok {
		operand, err := st.ValidateExpression(expr.Init)
		if err != nil {
			return err
		}
		if expr.Operator == tokens.ASSIGN {
			err := ShouldConstruct(operand, class)
			if err != nil {
				return NodeError(expr, err.Error())
			}
			return nil
		} else {
			err = ShouldOperate(getEffectOperator(expr.Operator), class, operand)
			if err != nil {
				return NodeError(expr, err.Error())
			}
		}
	} else {
		return NodeError(expr, "cannot assign to non-value object")
	}
	return nil
}

// ---
// IF STATEMENTS
// ---

func (st *SymbolTable) ResolveIfStatement(expr ast.IfStatement) (ValueObject, error) {
	condition, err := st.ResolveValueObject(expr.Condition)
	if err != nil {
		return nil, err
	}
	conditionResult, ok := condition.(BooleanLiteral)
	if !ok {
		return nil, NodeError(expr.Condition, "if condition must be a boolean")
	}
	var returnObject ValueObject
	if conditionResult {
		returnObject, err = st.ResolveBlock(expr.Body)
	} else {
		switch alt := expr.Alternate.(type) {
		case ast.IfStatement:
			returnObject, err = st.ResolveIfStatement(alt)
		case ast.Block:
			returnObject, err = st.ResolveBlock(alt)
		}
	}
	if err != nil {
		return nil, err
	}
	return returnObject, nil
}
func (st *SymbolTable) ValidateIfStatement(expr ast.IfStatement) error {
	condition, err := st.ValidateExpression(expr.Condition)
	if err != nil {
		return err
	}
	if _, ok := condition.(Boolean); !ok {
		return NodeError(expr.Condition, "if condition must be a boolean")
	}
	err = st.ValidateBlock(expr.Body)
	if err != nil {
		return err
	}
	switch alt := expr.Alternate.(type) {
	case ast.IfStatement:
		err = st.ValidateIfStatement(alt)
	case ast.Block:
		err = st.ValidateBlock(alt)
	}
	return err
}
func (st *SymbolTable) ValidateIfStatementReturns(expr ast.IfStatement, shouldReturn Class) (bool, error) {
	blockPassed, err := st.ValidateBlockReturns(expr.Body, shouldReturn)
	if err != nil {
		return false, err
	}
	if blockPassed {
		switch alt := expr.Alternate.(type) {
		case ast.IfStatement:
			return st.ValidateIfStatementReturns(alt, shouldReturn)
		case ast.Block:
			return st.ValidateBlockReturns(alt, shouldReturn)
		}
	}
	return false, nil
}

// ---
// WHILE STATEMENTS
// ---

func (st *SymbolTable) ResolveWhileStatement(expr ast.WhileStatement) (ValueObject, error) {
loopBlock:
	for {
		condition, err := st.ResolveValueObject(expr.Condition)
		if err != nil {
			return nil, err
		}
		conditionResult, ok := condition.(BooleanLiteral)
		if !ok {
			return nil, NodeError(expr.Condition, "while condition must be a boolean")
		}
		if !conditionResult {
			break
		}
		scopeTable := st.Clone()
		for _, stmt := range expr.Body.Statements {
			switch stmt.Init.(type) {
			case ast.ContinueStatement:
				continue loopBlock
			case ast.BreakStatement:
				break loopBlock
			default:
				returnObject, err := scopeTable.ResolveBlockStatement(stmt)
				if err != nil {
					return nil, err
				}
				if returnObject != nil {
					return returnObject, nil
				}
			}
		}
	}
	return nil, nil
}
func (st *SymbolTable) ValidateWhileStatement(expr ast.WhileStatement) error {
	condition, err := st.ValidateExpression(expr.Condition)
	if err != nil {
		return err
	}
	if _, ok := condition.(Boolean); !ok {
		return NodeError(expr.Condition, "if condition must be a boolean")
	}
	err = st.ValidateLoopBlock(expr.Body)
	if err != nil {
		return err
	}
	return nil
}
func (st *SymbolTable) ValidateWhileStatementReturns(expr ast.WhileStatement, shouldReturn Class) (bool, error) {
	return st.ValidateBlockReturns(expr.Body, shouldReturn)
}

// ---
// FOR STATEMENTS
// ---

func (st *SymbolTable) ResolveForStatement(expr ast.ForStatement) (ValueObject, error) {
	switch conditionBlock := expr.Condition.(type) {
	case ast.ForCondition:
	forLoopBlock:
		for {
			scopeTable := st.Clone()
			if conditionBlock.Init != nil {
				err := scopeTable.ResolveDeclarationStatement(*conditionBlock.Init, true)
				if err != nil {
					return nil, err
				}
			}
			condition, err := scopeTable.ResolveValueObject(conditionBlock.Condition)
			if err != nil {
				return nil, err
			}
			conditionResult, ok := condition.(BooleanLiteral)
			if !ok {
				return nil, NodeError(conditionBlock.Condition, "for condition must be a boolean")
			}
			if !conditionResult {
				break
			}
			for _, stmt := range expr.Body.Statements {
				switch stmt.Init.(type) {
				case ast.ContinueStatement:
					continue forLoopBlock
				case ast.BreakStatement:
					break forLoopBlock
				default:
					returnObject, err := scopeTable.ResolveBlockStatement(stmt)
					if err != nil {
						return nil, err
					}
					if returnObject != nil {
						return returnObject, nil
					}
				}
			}
			switch updateExpr := conditionBlock.Update.(type) {
			case ast.Expression:
				_, err := scopeTable.ResolveValueObject(updateExpr)
				if err != nil {
					return nil, err
				}
			case ast.AssignmentExpression:
				err := scopeTable.ResolveAssignmentExpression(updateExpr)
				if err != nil {
					return nil, err
				}
			}
		}
	case ast.RangeCondition:
		targetObject, err := st.ResolveValueObject(conditionBlock.Target)
		if err != nil {
			return nil, err
		}
		iter, ok := targetObject.(Iterable)
		if !ok {
			return nil, NotIterableError(conditionBlock.Target, targetObject)
		}
	rangeLoopBlock:
		for idx, item := range iter.Items {
			scopeTable := st.Clone()
			castedIndexLiteral := Object(IntegerLiteral(idx))
			scopeTable.local[conditionBlock.Index] = &castedIndexLiteral
			castedItemObject := Object(item)
			scopeTable.local[conditionBlock.Value] = &castedItemObject

			for _, stmt := range expr.Body.Statements {
				switch stmt.Init.(type) {
				case ast.ContinueStatement:
					continue rangeLoopBlock
				case ast.BreakStatement:
					break rangeLoopBlock
				default:
					returnObject, err := scopeTable.ResolveBlockStatement(stmt)
					if err != nil {
						return nil, err
					}
					if returnObject != nil {
						return returnObject, nil
					}
				}
			}
		}
	}
	return nil, nil
}

func (st *SymbolTable) applyForConditionForValidation(expr ast.ForStatement) error {
	switch conditionBlock := expr.Condition.(type) {
	case ast.ForCondition:
		if conditionBlock.Init != nil {
			err := st.ResolveDeclarationStatement(*conditionBlock.Init, false)
			if err != nil {
				return err
			}
		}
		condition, err := st.ValidateExpression(conditionBlock.Condition)
		if err != nil {
			return err
		}
		if _, ok := condition.(Boolean); !ok {
			return NodeError(conditionBlock.Condition, "for statement condition must be a boolean")
		}
		switch updateExpr := conditionBlock.Update.(type) {
		case ast.Expression:
			_, err = st.ValidateExpression(updateExpr)
			if err != nil {
				return err
			}
		case ast.AssignmentExpression:
			err = st.ValidateAssignmentExpression(updateExpr)
			if err != nil {
				return err
			}
		}
	case ast.RangeCondition:
		targetObject, err := st.ValidateExpression(conditionBlock.Target)
		if err != nil {
			return err
		}
		if iter, ok := targetObject.(Iterable); ok {
			castedIndexLiteral := Object(IntegerLiteral(0))
			st.local[conditionBlock.Index] = &castedIndexLiteral
			castedItemObject := Object(iter.ParentType)
			st.local[conditionBlock.Value] = &castedItemObject
		} else {
			return NotIterableError(conditionBlock.Target, targetObject)
		}
	default:
		return NodeError(expr.Condition, "invalid for condition")
	}
	return nil
}
func (st *SymbolTable) ValidateForStatement(expr ast.ForStatement) error {
	scopeTable := st.Clone()
	err := scopeTable.applyForConditionForValidation(expr)
	if err != nil {
		return err
	}
	return scopeTable.ValidateLoopBlock(expr.Body)
}
func (st *SymbolTable) ValidateForStatementReturns(expr ast.ForStatement, shouldReturn Class) (bool, error) {
	scopeTable := st.Clone()
	err := scopeTable.applyForConditionForValidation(expr)
	if err != nil {
		return false, err
	}
	return scopeTable.ValidateBlockReturns(expr.Body, shouldReturn)
}

// ---
// SWITCH STATEMENTS
// ---

func (st *SymbolTable) ResolveSwitchBlock(expr ast.SwitchBlock) (ValueObject, error) {
	target, err := st.ResolveValueObject(expr.Target)
	if err != nil {
		return nil, err
	}

	if _, ok := target.Class().(ComparableClass); !ok {
		return nil, InoperableSwitchTargetError(expr.Target, target)
	}
	resolved := false
	for _, caseBlock := range expr.Statements {
		if !caseBlock.IsDefault {
			caseCondition, err := st.ResolveValueObject(*caseBlock.Condition)
			if err != nil {
				return nil, err
			}
			evaluated, err := Operate(tokens.EQUALS, target, caseCondition)
			if err != nil {
				return nil, NodeError(caseBlock, err.Error())
			}
			if conditionResult, ok := evaluated.(BooleanLiteral); ok && bool(conditionResult) {
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
	}
	if !resolved {
		for _, caseBlock := range expr.Statements {
			if caseBlock.IsDefault {
				returnObject, err := st.ResolveBlock(caseBlock.Body)
				if err != nil {
					return nil, err
				}
				if returnObject != nil {
					return returnObject, nil
				}
			}
		}
	}
	return nil, nil
}
func (st *SymbolTable) ValidateSwitchBlock(expr ast.SwitchBlock) error {
	target, err := st.ValidateExpression(expr.Target)
	if err != nil {
		return err
	}
	comparable, ok := target.(ComparableClass)
	if !ok {
		return InoperableSwitchTargetError(expr.Target, target)
	}
	hasDefaultBlock := false
	for _, caseBlock := range expr.Statements {
		if caseBlock.IsDefault {
			if hasDefaultBlock {
				return NodeError(expr, "switch statement can only have one default block")
			}
			hasDefaultBlock = true
		} else {
			caseCondition, err := st.ValidateExpression(*caseBlock.Condition)
			if err != nil {
				return err
			}
			if fn := comparable.ComparableRules().Get(caseCondition, tokens.EQUALS); fn == nil {
				return NodeError(caseBlock.Condition, "switch statement case condition must be comparable to switch target")
			}
		}
	}
	return nil
}
func (st *SymbolTable) ValidateSwitchBlockReturns(expr ast.SwitchBlock, shouldReturn Class) (bool, error) {
	for _, caseBlock := range expr.Statements {
		if !caseBlock.IsDefault {
			blockPassed, err := st.ValidateBlockReturns(caseBlock.Body, shouldReturn)
			if err != nil {
				return false, err
			}
			if !blockPassed {
				return false, nil
			}
		}
	}
	for _, caseBlock := range expr.Statements {
		if caseBlock.IsDefault {
			blockPassed, err := st.ValidateBlockReturns(caseBlock.Body, shouldReturn)
			if err != nil {
				return false, err
			}
			return blockPassed, nil
		}
	}
	return false, nil
}

// ---
// GUARD STATEMENTS
// ---

func (st *SymbolTable) guardStatementHandler(expr ast.GuardStatement) (ValueObject, *Function, error) {
	if proto := st.immutable["self"]; proto != nil {
		if protoObject, ok := proto.(ValueObject); ok {
			fn, err := protoObject.Get("guard")
			if err != nil {
				return nil, nil, err
			}
			if fn != nil {
				if guardFn, ok := fn.(Function); ok {
					return protoObject, &guardFn, nil
				}
			}
		}
	}
	return nil, nil, NodeError(expr, "function has no guard directive")
}
func (st *SymbolTable) ResolveGuardStatement(expr ast.GuardStatement) error {
	protoObject, guardFn, err := st.guardStatementHandler(expr)
	if err != nil {
		return err
	}
	obj, err := st.ResolveValueObject(expr.Init)
	if err != nil {
		return err
	}
	_, err = guardFn.Call([]ValueObject{obj}, protoObject)
	return err
}
func (st *SymbolTable) ValidateGuardStatement(expr ast.GuardStatement) error {
	_, guardFn, err := st.guardStatementHandler(expr)
	if err != nil {
		return err
	}
	class, err := st.ValidateExpression(expr.Init)
	if err != nil {
		return err
	}
	err = ValidateMethodArguments(guardFn, []Class{class})
	if err != nil {
		return NodeError(expr, err.Error())
	}
	return nil
}

// ---
// BLOCK STATEMENTS
// ---

func (st *SymbolTable) ResolveBlockStatement(expr ast.BlockStatement) (ValueObject, error) {
	var err error
	var returnObject ValueObject
	switch expr := expr.Init.(type) {
	case ast.Expression:
		_, err = st.ResolveExpression(expr)
	case ast.DeclarationStatement:
		err = st.ResolveDeclarationStatement(expr, true)
	case ast.AssignmentExpression:
		err = st.ResolveAssignmentExpression(expr)
	case ast.IfStatement:
		returnObject, err = st.ResolveIfStatement(expr)
	case ast.WhileStatement:
		returnObject, err = st.ResolveWhileStatement(expr)
	case ast.ForStatement:
		returnObject, err = st.ResolveForStatement(expr)
	// case ast.ContinueStatement, ast.BreakStatement:
	//   noop
	case ast.SwitchBlock:
		returnObject, err = st.ResolveSwitchBlock(expr)
	case ast.GuardStatement:
		err = st.ResolveGuardStatement(expr)
	case ast.ReturnStatement:
		returnObject, err = st.ResolveValueObject(expr.Init)
		if err != nil {
			return nil, err
		}
	case ast.ThrowStatement:
		returnObject, err = st.ResolveValueObject(expr.Init)
		if err != nil {
			return nil, err
		}
		if thrownErr, ok := returnObject.(Error); ok {
			return nil, thrownErr
		} else {
			return nil, NodeError(expr, "throw statement must be an error")
		}
	default:
		return nil, NodeError(expr, "unknown block statement type %T", expr)
	}
	return returnObject, err
}
func (st *SymbolTable) ValidateBlockStatement(expr ast.BlockStatement) error {
	var err error
	switch expr := expr.Init.(type) {
	case ast.Expression:
		_, err = st.ValidateExpression(expr)
	case ast.DeclarationStatement:
		err = st.ResolveDeclarationStatement(expr, false)
	case ast.AssignmentExpression:
		err = st.ValidateAssignmentExpression(expr)
	case ast.IfStatement:
		err = st.ValidateIfStatement(expr)
	case ast.WhileStatement:
		err = st.ValidateWhileStatement(expr)
	case ast.ForStatement:
		err = st.ValidateForStatement(expr)
	// case ast.ContinueStatement, ast.BreakStatement:
	//   noop
	case ast.SwitchBlock:
		err = st.ValidateSwitchBlock(expr)
	case ast.GuardStatement:
		err = st.ValidateGuardStatement(expr)
	case ast.ReturnStatement:
		_, err = st.ValidateExpression(expr.Init)
	case ast.ThrowStatement:
		returnObject, err := st.ValidateExpression(expr.Init)
		if err != nil {
			return err
		}
		if _, ok := returnObject.(Error); !ok {
			return NodeError(expr, "throw statement must be an error")
		}
	default:
		return NodeError(expr, "unknown block statement type %T", expr)
	}
	return err
}

// ---
// BLOCKS
// ---

func (st *SymbolTable) ResolveBlock(expr ast.Block) (ValueObject, error) {
	scopeTable := st.Clone()
	for _, stmt := range expr.Statements {
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
func (st *SymbolTable) ValidateBlock(expr ast.Block) error {
	scopeTable := st.Clone()
	for _, stmt := range expr.Statements {
		switch stmt.Init.(type) {
		case ast.ContinueStatement:
			return NodeError(stmt, "continue statement outside loop")
		case ast.BreakStatement:
			return NodeError(stmt, "break statement outside loop")
		default:
			err := scopeTable.ValidateBlockStatement(stmt)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func (st *SymbolTable) ValidateBlockReturns(expr ast.Block, shouldReturn Class) (bool, error) {
	scopeTable := st.Clone()
	for _, stmt := range expr.Statements {
		doesReturn := false
		var err error
		switch expr := stmt.Init.(type) {
		case ast.ReturnStatement:
			returnClass, err := scopeTable.ValidateExpression(expr.Init)
			if err != nil {
				return false, err
			}
			err = ShouldConstruct(shouldReturn, returnClass)
			if err != nil {
				return false, NodeError(expr, err.Error())
			}
			doesReturn = true
		case ast.ThrowStatement:
			returnObject, err := scopeTable.ValidateExpression(expr.Init)
			if err != nil {
				return false, err
			}
			if _, ok := returnObject.(Error); !ok {
				return false, NodeError(expr, "throw type %s is not an error", returnObject.ClassName())
			}
			doesReturn = true
		case ast.IfStatement:
			doesReturn, err = scopeTable.ValidateIfStatementReturns(expr, shouldReturn)
		case ast.WhileStatement:
			doesReturn, err = scopeTable.ValidateWhileStatementReturns(expr, shouldReturn)
		case ast.ForStatement:
			doesReturn, err = scopeTable.ValidateForStatementReturns(expr, shouldReturn)
		case ast.SwitchBlock:
			doesReturn, err = scopeTable.ValidateSwitchBlockReturns(expr, shouldReturn)
		default:
			err := scopeTable.ValidateBlockStatement(stmt)
			if err != nil {
				return false, err
			}
		}
		if err != nil {
			return false, err
		}
		if doesReturn {
			return true, nil
		}
	}
	return false, nil
}

// only difference between ValidateBlock is that it allows for continue/break statements
func (st *SymbolTable) ValidateLoopBlock(expr ast.Block) error {
	scopeTable := st.Clone()
	for _, stmt := range expr.Statements {
		err := scopeTable.ValidateBlockStatement(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

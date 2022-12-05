package build

import (
	"strings"

	"github.com/hntrl/lang/language/nodes"
	"github.com/hntrl/lang/language/tokens"
)

// There's a couple of paradigms you should understand here:
//  ValidateSomeExpression represents if the expression is valid
//  ValidateSomeExpressionReturns is reserved for block statements that can return a value to make sure the entire block returns a value
//	ResolveSomeExpression represents what is yielded by an expression -- what is outputted when you run it?
//		in the case that actually evaluating an expression would mean skipping over some steps of semantic analysis, a validate method is supplemented
//		(this is really only true for function expressions)

type SymbolTable struct {
	immutable map[string]Object
	local     map[string]Object
}

func (st SymbolTable) Joined() map[string]Object {
	joined := make(map[string]Object)
	for k, v := range st.immutable {
		joined[k] = v
	}
	for k, v := range st.local {
		joined[k] = v
	}
	return joined
}

func (st SymbolTable) Clone() SymbolTable {
	immutable := make(map[string]Object)
	for k, v := range st.immutable {
		immutable[k] = v
	}
	local := make(map[string]Object)
	for k, v := range st.local {
		local[k] = v
	}
	return SymbolTable{
		immutable: immutable,
		local:     local,
	}
}

// Returns the Object targeted by a selector
func (st SymbolTable) ResolveSelector(selector nodes.Selector) (Object, error) {
	table := st.Joined()
	current := table[selector.Members[0]]
	resolveChainString := selector.Members[0]
	if current == nil {
		return nil, NodeError(selector, "unknown selector %s", selector.Members[0])
	}
	for _, member := range selector.Members[1:] {
		nextObj := current.Get(member)
		if nextObj == nil {
			return nil, NodeError(selector, "%s has no member %s", resolveChainString, member)
		}
		current = nextObj
		resolveChainString += "." + member
	}
	return current, nil
}

// Gets the ValueObject for a given expression. Returns an error if the expression could not be resolved
// If shouldEvalute is true, function calls will be evaluated and the result returned
func (st SymbolTable) ResolveExpression(expr nodes.Expression) (Object, error) {
	switch expr := expr.Init.(type) {
	case nodes.Literal:
		return st.ResolveLiteral(expr)
	case nodes.ArrayExpression:
		return st.ResolveArrayExpression(expr)
	case nodes.InstanceExpression:
		return st.ResolveInstanceExpression(expr)
	case nodes.UnaryExpression:
		return st.ResolveUnaryExpression(expr)
	case nodes.BinaryExpression:
		return st.ResolveBinaryExpression(expr)
	case nodes.ObjectPattern:
		generic, err := st.ParsePropertyList(expr.Properties)
		if err != nil {
			return nil, err
		}
		return *generic, nil
	case nodes.ValueExpression:
		return st.ResolveValueExpression(expr)
	case nodes.Expression:
		return st.ResolveExpression(expr)
	default:
		return nil, NodeError(expr, "unknown expression type %T", expr)
	}
}

func (st SymbolTable) ValidateExpression(expr nodes.Expression) (Class, error) {
	switch expr := expr.Init.(type) {
	case nodes.Literal:
		lit, err := st.ResolveLiteral(expr)
		if err != nil {
			return nil, err
		}
		return lit.Class(), nil
	case nodes.ArrayExpression:
		return st.ValidateArrayExpression(expr)
	case nodes.InstanceExpression:
		return st.ValidateInstanceExpression(expr)
	case nodes.UnaryExpression:
		return st.ValidateUnaryExpression(expr)
	case nodes.BinaryExpression:
		return st.ValidateBinaryExpression(expr)
	case nodes.ObjectPattern:
		obj, err := st.ValidatePropertyList(expr.Properties)
		if err != nil {
			return nil, err
		}
		return *obj, nil
	case nodes.ValueExpression:
		return st.ValidateValueExpression(expr)
	case nodes.Expression:
		return st.ValidateExpression(expr)
	default:
		return nil, NodeError(expr, "unknown expression type %T", expr)
	}
}

func (st SymbolTable) ResolveValueObject(expr nodes.Expression) (ValueObject, error) {
	obj, err := st.ResolveExpression(expr)
	if err != nil {
		return nil, err
	}
	if valueObj, ok := obj.(ValueObject); ok {
		return valueObj, nil
	}
	return nil, AmbiguousObjectError(expr, obj)
}

func (st SymbolTable) ResolveLiteral(expr nodes.Literal) (ValueObject, error) {
	switch lit := expr.Value.(type) {
	case string:
		return StringLiteral(lit), nil
	case int64:
		return IntegerLiteral(lit), nil
	case float64:
		return FloatLiteral(lit), nil
	case bool:
		return BooleanLiteral(lit), nil
	case nil:
		return NilLiteral{}, nil
	default:
		return nil, NodeError(expr, "unknown literal type %T", lit)
	}
}

// --
// ARRAY EXPRESSIONS
// --

func (st SymbolTable) ResolveArrayExpression(expr nodes.ArrayExpression) (ValueObject, error) {
	parentType, err := st.ResolveTypeExpression(expr.Init)
	if err != nil {
		return nil, err
	}
	iterable := NewIterable(parentType, len(expr.Elements))
	for idx, elementExpr := range expr.Elements {
		element, err := st.ResolveValueObject(elementExpr)
		if err != nil {
			return nil, NodeError(elementExpr, err.Error())
		}
		iterable.Items[idx], err = Construct(iterable.ParentType, element)
		if err != nil {
			return nil, NodeError(elementExpr, err.Error())
		}
	}
	return iterable, nil
}
func (st SymbolTable) ValidateArrayExpression(expr nodes.ArrayExpression) (Class, error) {
	parentType, err := st.ResolveTypeExpression(expr.Init)
	if err != nil {
		return nil, err
	}
	iterable := NewIterable(parentType, 0)
	for _, elementExpr := range expr.Elements {
		element, err := st.ValidateExpression(elementExpr)
		if err != nil {
			return nil, err
		}
		err = ShouldConstruct(iterable.ParentType, element)
		if err != nil {
			return nil, NodeError(elementExpr, err.Error())
		}
	}
	return iterable, nil
}

// --
// INSTANCE EXPRESSIONS
// --

func (st SymbolTable) ResolveInstanceExpression(expr nodes.InstanceExpression) (ValueObject, error) {
	targetType, err := st.ResolveSelector(expr.Selector)
	if err != nil {
		return nil, err
	}
	if class, ok := targetType.(Class); ok {
		generic, err := st.ParsePropertyList(expr.Properties)
		if err != nil {
			return nil, err
		}
		return Construct(class, *generic)
	} else {
		return nil, NodeError(expr, "%s is not instanceable", strings.Join(expr.Selector.Members, ","))
	}
}
func (st SymbolTable) ValidateInstanceExpression(expr nodes.InstanceExpression) (Class, error) {
	targetType, err := st.ResolveSelector(expr.Selector)
	if err != nil {
		return nil, err
	}
	if class, ok := targetType.(Class); ok {
		generic, err := st.ValidatePropertyList(expr.Properties)
		if err != nil {
			return nil, err
		}
		err = ShouldConstruct(class, *generic)
		if err != nil {
			return nil, NodeError(expr, err.Error())
		}
		return class, nil
	} else {
		return nil, NodeError(expr, "%s is not instanceable", strings.Join(expr.Selector.Members, ","))
	}
}

// --
// UNARY EXPRESSIONS
// --

func (st SymbolTable) ResolveUnaryExpression(expr nodes.UnaryExpression) (ValueObject, error) {
	obj, err := st.ResolveValueObject(expr.Init)
	if err != nil {
		return nil, err
	}
	switch expr.Operator {
	case tokens.ADD, tokens.SUB:
		class := obj.Class()
		if numValue, ok := obj.(NumberLiteral); ok {
			if expr.Operator == tokens.ADD {
				return Construct(class, NumberLiteral(+float64(numValue)))
			} else {
				return Construct(class, NumberLiteral(-float64(numValue)))
			}
		} else {
			return nil, NodeError(expr, "cannot apply unary %s to %s", expr.Operator, obj.Class().ClassName())
		}
	case tokens.NOT:
		if boolValue, ok := obj.(BooleanLiteral); ok {
			return BooleanLiteral(!boolValue), nil
		} else {
			return nil, NodeError(expr, "cannot apply unary ! to %s", obj.Class().ClassName())
		}
	default:
		return nil, NodeError(expr, "unknown unary operator %s", expr.Operator)
	}
}
func (st SymbolTable) ValidateUnaryExpression(expr nodes.UnaryExpression) (Class, error) {
	class, err := st.ValidateExpression(expr.Init)
	if err != nil {
		return nil, err
	}
	switch expr.Operator {
	case tokens.ADD, tokens.SUB:
		err = ShouldConstruct(Number{}, class)
		if err != nil {
			return nil, NodeError(expr, err.Error())
		}
		return class, nil
	case tokens.NOT:
		err = ShouldConstruct(Boolean{}, class)
		if err != nil {
			return nil, NodeError(expr, err.Error())
		}
		return Boolean{}, nil
	default:
		return nil, NodeError(expr, "unknown unary operator %s", expr.Operator)
	}
}

// --
// BINARY EXPRESSIONS
// --

func (st SymbolTable) ResolveBinaryExpression(expr nodes.BinaryExpression) (ValueObject, error) {
	if !expr.Operator.IsOperator() && !expr.Operator.IsComparableOperator() {
		return nil, NodeError(expr, "invalid binary operator %s", expr.Operator)
	}
	left, err := st.ResolveValueObject(expr.Left)
	if err != nil {
		return nil, err
	}
	right, err := st.ResolveValueObject(expr.Right)
	if err != nil {
		return nil, err
	}
	obj, err := Operate(expr.Operator, left, right)
	if err != nil {
		return nil, NodeError(expr, err.Error())
	}
	return obj, nil
}
func (st SymbolTable) ValidateBinaryExpression(expr nodes.BinaryExpression) (Class, error) {
	if !expr.Operator.IsOperator() && !expr.Operator.IsComparableOperator() {
		return nil, NodeError(expr, "invalid binary operator %s", expr.Operator)
	}
	left, err := st.ValidateExpression(expr.Left)
	if err != nil {
		return nil, err
	}
	right, err := st.ValidateExpression(expr.Right)
	if err != nil {
		return nil, err
	}
	err = ShouldOperate(expr.Operator, left, right)
	if err != nil {
		return nil, NodeError(expr, err.Error())
	}
	if expr.Operator.IsComparableOperator() {
		return Boolean{}, nil
	}
	return left, nil
}

// --
// VALUE EXPRESSIONS
// --

func (st SymbolTable) valueExpressionPredicate(expr nodes.ValueExpression) (string, Object, error) {
	table := st.Joined()
	resolveChainString := ""
	var current Object

	if firstIdent, ok := expr.Members[0].Init.(string); ok {
		current = table[firstIdent]
		resolveChainString += firstIdent
	} else {
		return resolveChainString, nil, NodeError(expr, "invalid value expression")
	}

	if current == nil {
		return resolveChainString, nil, UnknownSelector(expr, resolveChainString)
	}
	return resolveChainString, current, nil
}
func (st SymbolTable) ResolveValueExpression(expr nodes.ValueExpression) (Object, error) {
	resolveChainString, current, err := st.valueExpressionPredicate(expr)
	if err != nil {
		return nil, err
	}

	for _, memberExpr := range expr.Members[1:] {
		switch expr := memberExpr.Init.(type) {
		case string:
			current = current.Get(expr)
			if current == nil {
				return nil, NoPropertyError(memberExpr, resolveChainString, current, expr)
			}
			resolveChainString += "." + expr
		case nodes.CallExpression:
			if method, ok := current.(Method); ok {
				passedArguments := make([]ValueObject, len(expr.Arguments))
				for idx, argExpr := range expr.Arguments {
					arg, err := st.ResolveValueObject(argExpr)
					if err != nil {
						return nil, err
					}
					passedArguments[idx] = arg
				}
				args, err := ResolveMethodArguments(method, passedArguments)
				if err != nil {
					return nil, NodeError(memberExpr, err.Error())
				}
				current, err = method.Call(args, GenericObject{})
				if err != nil {
					return nil, NodeError(memberExpr, err.Error())
				}
				resolveChainString += "()"
			} else if class, ok := current.(Class); ok {
				if len(expr.Arguments) != 1 {
					return nil, InvalidValueExpressionError(memberExpr)
				}
				arg, err := st.ResolveValueObject(expr.Arguments[0])
				if err != nil {
					return nil, err
				}
				current, err = Construct(class, arg)
				if err != nil {
					return nil, NodeError(memberExpr, err.Error())
				}
				resolveChainString += "()"
			} else {
				return nil, UncallableError(memberExpr, resolveChainString, current)
			}
		case nodes.IndexExpression:
			if valueObj, ok := current.(ValueObject); ok {
				if indexable, ok := valueObj.(Indexable[ValueObject]); ok {
					var left int
					if expr.Left != nil {
						leftExpr, err := st.ResolveValueObject(*expr.Left)
						if err != nil {
							return nil, err
						}
						if leftNum, ok := leftExpr.(IntegerLiteral); ok {
							left = int(leftNum)
						} else {
							return nil, InvalidIndexError(memberExpr, leftExpr)
						}
					}
					if expr.IsRange {
						var right int
						if expr.Right != nil {
							rightExpr, err := st.ResolveValueObject(*expr.Right)
							if err != nil {
								return nil, err
							}
							if rightNum, ok := rightExpr.(IntegerLiteral); ok {
								right = int(rightNum)
							} else {
								return nil, InvalidIndexError(memberExpr, rightExpr)
							}
						}
						current, err = indexable.Range(left, right)
						if err != nil {
							return nil, NodeError(memberExpr, err.Error())
						}
					} else {
						current, err = indexable.GetIndex(left)
						if err != nil {
							return nil, NodeError(memberExpr, err.Error())
						}
					}
					resolveChainString += "[]"
				} else {
					return nil, NotIndexableError(memberExpr, resolveChainString, current)
				}
			} else {
				return nil, AmbiguousObjectError(memberExpr, current)
			}
		default:
			return nil, InvalidValueExpressionError(memberExpr)
		}
	}
	return current, nil
}
func (st SymbolTable) ValidateValueExpression(expr nodes.ValueExpression) (Class, error) {
	resolveChainString, current, err := st.valueExpressionPredicate(expr)
	if err != nil {
		return nil, err
	}

	for _, memberExpr := range expr.Members[1:] {
		switch expr := memberExpr.Init.(type) {
		case string:
			if prop := current.Get(expr); prop != nil {
				current = current.Get(expr)
				if current == nil {
					return nil, NoPropertyError(memberExpr, resolveChainString, current, expr)
				}
			} else if class, ok := current.(ObjectClass); ok {
				if field, ok := class.Fields()[expr]; ok {
					current = field
				} else if obj := class.Get(expr); obj != nil {
					current = obj
				} else {
					return nil, NoPropertyError(memberExpr, resolveChainString, current, expr)
				}
			} else {
				return nil, NoPropertyError(memberExpr, resolveChainString, current, expr)
			}
			resolveChainString += "." + expr
		case nodes.CallExpression:
			if method, ok := current.(Method); ok {
				passedArguments := make([]Class, len(expr.Arguments))
				for idx, argExpr := range expr.Arguments {
					arg, err := st.ValidateExpression(argExpr)
					if err != nil {
						return nil, err
					}
					passedArguments[idx] = arg
				}
				err := ValidateMethodArguments(method, passedArguments)
				if err != nil {
					return nil, NodeError(memberExpr, err.Error())
				}
				current = method.Returns()
				resolveChainString += "()"
			} else if class, ok := current.(Class); ok {
				if len(expr.Arguments) != 1 {
					return nil, InvalidValueExpressionError(memberExpr)
				}
				arg, err := st.ValidateExpression(expr.Arguments[0])
				if err != nil {
					return nil, err
				}
				err = ShouldConstruct(class, arg)
				if err != nil {
					return nil, NodeError(memberExpr, err.Error())
				}
				resolveChainString += "()"
			} else {
				return nil, UncallableError(memberExpr, resolveChainString, current)
			}
		case nodes.IndexExpression:
			if valueObj, ok := current.(ValueObject); ok {
				if _, ok := valueObj.(Indexable[ValueObject]); ok {
					if expr.Left != nil {
						leftExpr, err := st.ValidateExpression(*expr.Left)
						if err != nil {
							return nil, err
						}
						if _, ok := leftExpr.(Integer); !ok {
							return nil, InvalidIndexError(memberExpr, leftExpr)
						}
					}
					if expr.IsRange {
						if expr.Right != nil {
							rightExpr, err := st.ValidateExpression(*expr.Right)
							if err != nil {
								return nil, err
							}
							if _, ok := rightExpr.(Integer); ok {
								return nil, InvalidIndexError(memberExpr, rightExpr)
							}
						}
					}
					resolveChainString += "[]"
				} else {
					return nil, NotIndexableError(memberExpr, resolveChainString, current)
				}
			} else {
				return nil, AmbiguousObjectError(memberExpr, current)
			}
		default:
			return nil, InvalidValueExpressionError(memberExpr)
		}
	}
	if class, ok := current.(Class); ok {
		return class, nil
	} else if valueObj, ok := current.(ValueObject); ok {
		return valueObj.Class(), nil
	} else {
		return nil, nil
	}
}

// --
// PROPERTY LIST (GENERIC OBJECT)
// --

func (st SymbolTable) ParsePropertyList(props nodes.PropertyList) (*GenericObject, error) {
	generic := NewGenericObject()
	for _, prop := range props {
		switch expr := prop.(type) {
		case nodes.Property:
			obj, err := st.ResolveValueObject(expr.Init)
			if err != nil {
				return nil, err
			}
			generic.fields[expr.Key] = obj.Class()
			generic.data[expr.Key] = obj
		case nodes.SpreadElement:
			obj, err := st.ResolveValueObject(expr.Init)
			if err != nil {
				return nil, err
			}
			if objectClass, ok := obj.(ObjectClass); ok {
				for key := range objectClass.Fields() {
					val := obj.Get(key)
					if val == nil {
						generic.data[key] = NilLiteral{}
					} else if innerValueObj, ok := val.(ValueObject); ok {
						generic.fields[key] = innerValueObj.Class()
						generic.data[key] = innerValueObj
					} else {
						return nil, AmbiguousObjectError(expr, obj)
					}
				}
			} else {
				return nil, NodeError(expr, "Cannot spread non-object")
			}
		default:
			return nil, NodeError(prop, "invalid property list")
		}
	}
	return &generic, nil
}
func (st SymbolTable) ValidatePropertyList(props nodes.PropertyList) (*GenericObject, error) {
	generic := NewGenericObject()
	for _, prop := range props {
		switch expr := prop.(type) {
		case nodes.Property:
			obj, err := st.ValidateExpression(expr.Init)
			if err != nil {
				return nil, err
			}
			generic.fields[expr.Key] = obj
		case nodes.SpreadElement:
			obj, err := st.ValidateExpression(expr.Init)
			if err != nil {
				return nil, err
			}
			if objectClass, ok := obj.(ObjectClass); ok {
				for key, value := range objectClass.Fields() {
					generic.fields[key] = value
				}
			} else {
				return nil, NodeError(expr, "Cannot spread non-object")
			}
		default:
			return nil, NodeError(prop, "invalid property list")
		}
	}
	return &generic, nil
}

// --
// TYPE EXPRESSIONS
// --

// Returns the Class assumed by a type expression
func (st SymbolTable) ResolveTypeExpression(expr nodes.TypeExpression) (Class, error) {
	parentType, err := st.ResolveSelector(expr.Selector)
	if err != nil {
		return nil, err
	}
	if class, ok := parentType.(Class); ok {
		if expr.IsArray {
			class = NewIterable(class, 0)
		}
		if expr.IsOptional {
			class = NilableObject{class, nil}
		}
		return class, nil
	} else {
		return nil, InvalidTypeError(expr, parentType)
	}
}

package symbols

import (
	"fmt"
	"strings"

	"github.com/hntrl/hyper/internal/ast"
	"github.com/hntrl/hyper/internal/tokens"
	"github.com/kr/pretty"
)

// There's a couple of paradigms you should understand here:
//  ValidateSomeExpression represents if the expression is valid
//  ValidateSomeExpressionReturns is reserved for block statements that can return a value to make sure the entire block returns a value
//	ResolveSomeExpression represents what is yielded by an expression -- what is outputted when you run it?
//		in the case that actually evaluating an expression would mean skipping over some steps of semantic analysis, a validate method is supplemented
//		(this is really only true for function expressions)

type SymbolTable struct {
	root      Object
	immutable map[string]Object
	local     map[string]*Object
}

func NewSymbolTable(root Object) SymbolTable {
	return SymbolTable{
		root:  root,
		local: make(map[string]*Object),
		immutable: map[string]Object{
			"String": String{},
			"Double": Double{},
			"Float":  Float{},
			"Int":    Integer{},
			"Bool":   Boolean{},
			// "Date":     Date{},
			"DateTime": DateTime{},
			"print": NewFunction(FunctionOptions{
				Arguments: []Class{
					AnyClass{},
				},
				Handler: func(args []ValueObject, proto ValueObject) (ValueObject, error) {
					fmt.Println(args[0].Value())
					return nil, nil
				},
			}),
			"deprecated_debug": NewFunction(FunctionOptions{
				Arguments: []Class{
					AnyClass{},
				},
				Handler: func(args []ValueObject, proto ValueObject) (ValueObject, error) {
					fmt.Println(pretty.Print(args[0]))
					return nil, nil
				},
			}),
			"len": NewFunction(FunctionOptions{
				Arguments: []Class{
					// FIXME: this means arguments skip validation, but it should be
					// Indexable. not a good way to put in arguments but it has to since
					// it's an interface
					Iterable{ParentType: AnyClass{}},
				},
				Returns: Integer{},
				Handler: func(args []ValueObject, proto ValueObject) (ValueObject, error) {
					if indexable, ok := args[0].(Indexable); ok {
						return IntegerLiteral(indexable.Len()), nil
					}
					return nil, fmt.Errorf("cannot get length of %s", args[0].Class().ClassName())
				},
			}),
		},
	}
}

func (st SymbolTable) Get(key string) (Object, error) {
	if obj := st.immutable[key]; obj != nil {
		return obj, nil
	}
	if obj := st.local[key]; obj != nil {
		return *obj, nil
	}
	obj, err := st.root.Get(key)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (st SymbolTable) Clone() SymbolTable {
	immutable := make(map[string]Object)
	for k, v := range st.immutable {
		immutable[k] = v
	}
	local := make(map[string]*Object)
	for k, v := range st.local {
		local[k] = v
	}
	return SymbolTable{
		root:      st.root,
		immutable: immutable,
		local:     local,
	}
}

// Returns the Object targeted by a selector
func (st SymbolTable) ResolveSelector(selector ast.Selector) (Object, error) {
	current, err := st.Get(selector.Members[0])
	if err != nil {
		return nil, err
	}
	resolveChainString := selector.Members[0]
	if current == nil {
		return nil, UnknownSelector(selector, selector.Members[0])
	}
	for _, member := range selector.Members[1:] {
		nextObj, err := current.Get(member)
		if err != nil {
			return nil, err
		}
		if nextObj == nil {
			return nil, NodeError(selector, "%s has no member %s", resolveChainString, member)
		}
		current = nextObj
		resolveChainString += "." + member
	}
	return current, nil
}

// Gets the ValueObject for a given expression. Returns an error if the expression could not be resolved
func (st SymbolTable) ResolveExpression(expr ast.Expression) (Object, error) {
	switch expr := expr.Init.(type) {
	case ast.Literal:
		return st.ResolveLiteral(expr)
	case ast.ArrayExpression:
		return st.ResolveArrayExpression(expr)
	case ast.InstanceExpression:
		return st.ResolveInstanceExpression(expr)
	case ast.UnaryExpression:
		return st.ResolveUnaryExpression(expr)
	case ast.BinaryExpression:
		return st.ResolveBinaryExpression(expr)
	case ast.ObjectPattern:
		mapObject, err := st.ParsePropertyList(expr.Properties)
		if err != nil {
			return nil, err
		}
		return mapObject, nil
	case ast.ValueExpression:
		return st.ResolveValueExpression(expr)
	case ast.Expression:
		return st.ResolveExpression(expr)
	default:
		return nil, NodeError(expr, "unknown expression type %T", expr)
	}
}

func (st SymbolTable) ValidateExpression(expr ast.Expression) (Class, error) {
	switch expr := expr.Init.(type) {
	case ast.Literal:
		lit, err := st.ResolveLiteral(expr)
		if err != nil {
			return nil, err
		}
		return lit.Class(), nil
	case ast.ArrayExpression:
		return st.ValidateArrayExpression(expr)
	case ast.InstanceExpression:
		return st.ValidateInstanceExpression(expr)
	case ast.UnaryExpression:
		return st.ValidateUnaryExpression(expr)
	case ast.BinaryExpression:
		return st.ValidateBinaryExpression(expr)
	case ast.ObjectPattern:
		mapObject, err := st.ValidatePropertyList(expr.Properties)
		if err != nil {
			return nil, err
		}
		return mapObject, nil
	case ast.ValueExpression:
		return st.ValidateValueExpression(expr)
	case ast.Expression:
		return st.ValidateExpression(expr)
	default:
		return nil, NodeError(expr, "unknown expression type %T", expr)
	}
}

func (st SymbolTable) ResolveValueObject(expr ast.Expression) (ValueObject, error) {
	obj, err := st.ResolveExpression(expr)
	if err != nil {
		return nil, err
	}
	if valueObj, ok := obj.(ValueObject); ok {
		return valueObj, nil
	}
	return nil, AmbiguousObjectError(expr, obj)
}

func (st SymbolTable) ResolveLiteral(expr ast.Literal) (ValueObject, error) {
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

// ---
// ARRAY EXPRESSIONS
// ---

func (st SymbolTable) ResolveArrayExpression(expr ast.ArrayExpression) (ValueObject, error) {
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
func (st SymbolTable) ValidateArrayExpression(expr ast.ArrayExpression) (Class, error) {
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

// ---
// INSTANCE EXPRESSIONS
// ---

func (st SymbolTable) ResolveInstanceExpression(expr ast.InstanceExpression) (ValueObject, error) {
	targetType, err := st.ResolveSelector(expr.Selector)
	if err != nil {
		return nil, err
	}
	class, ok := targetType.(Class)
	if !ok {
		return nil, NodeError(expr, "%s is not instanceable", strings.Join(expr.Selector.Members, ","))
	}
	mapObject, err := st.ParsePropertyList(expr.Properties)
	if err != nil {
		return nil, err
	}
	return Construct(class, mapObject)
}
func (st SymbolTable) ValidateInstanceExpression(expr ast.InstanceExpression) (Class, error) {
	targetType, err := st.ResolveSelector(expr.Selector)
	if err != nil {
		return nil, err
	}
	class, ok := targetType.(Class)
	if !ok {
		return nil, NodeError(expr, "%s is not instanceable", strings.Join(expr.Selector.Members, ","))
	}
	mapObject, err := st.ValidatePropertyList(expr.Properties)
	if err != nil {
		return nil, err
	}
	err = ShouldConstruct(class, mapObject)
	if err != nil {
		return nil, NodeError(expr, err.Error())
	}
	return class, nil
}

// ---
// UNARY EXPRESSIONS
// ---

func (st SymbolTable) ResolveUnaryExpression(expr ast.UnaryExpression) (ValueObject, error) {
	obj, err := st.ResolveValueObject(expr.Init)
	if err != nil {
		return nil, err
	}
	switch expr.Operator {
	case tokens.ADD, tokens.SUB:
		val, err := Construct(Number{}, obj)
		if err != nil {
			return nil, err
		}
		if expr.Operator == tokens.ADD {
			return Construct(obj.Class(), NumberLiteral(+float64(val.(NumberLiteral))))
		} else {
			return Construct(obj.Class(), NumberLiteral(-float64(val.(NumberLiteral))))
		}
	case tokens.NOT:
		boolValue, ok := obj.(BooleanLiteral)
		if !ok {
			return nil, NodeError(expr, "cannot apply unary ! to %s", obj.Class().ClassName())
		}
		return BooleanLiteral(!boolValue), nil
	default:
		return nil, NodeError(expr, "unknown unary operator %s", expr.Operator)
	}
}
func (st SymbolTable) ValidateUnaryExpression(expr ast.UnaryExpression) (Class, error) {
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

// ---
// BINARY EXPRESSIONS
// ---

func (st SymbolTable) ResolveBinaryExpression(expr ast.BinaryExpression) (ValueObject, error) {
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
func (st SymbolTable) ValidateBinaryExpression(expr ast.BinaryExpression) (Class, error) {
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

// ---
// VALUE EXPRESSIONS
// ---

func (st SymbolTable) valueExpressionPredicate(expr ast.ValueExpression) (string, Object, error) {
	resolveChainString := ""
	firstIdent, ok := expr.Members[0].Init.(string)
	if !ok {
		return resolveChainString, nil, NodeError(expr, "invalid value expression")
	}

	current, err := st.Get(firstIdent)
	if err != nil {
		return "", nil, err
	}
	resolveChainString += firstIdent
	if current == nil {
		return resolveChainString, nil, UnknownSelector(expr, resolveChainString)
	}
	return resolveChainString, current, nil
}
func (st SymbolTable) ResolveValueExpression(expr ast.ValueExpression) (Object, error) {
	resolveChainString, current, err := st.valueExpressionPredicate(expr)
	if err != nil {
		return nil, err
	}

	for _, memberExpr := range expr.Members[1:] {
		switch expr := memberExpr.Init.(type) {
		case string:
			current, err = current.Get(expr)
			if err != nil {
				return nil, err
			}
			if current == nil {
				return nil, NoPropertyError(memberExpr, resolveChainString, current, expr)
			}
			resolveChainString += "." + expr
		case ast.CallExpression:
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
				current, err = method.Call(args, &MapObject{})
				if err != nil {
					return nil, err
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
		case ast.IndexExpression:
			valueObj, ok := current.(ValueObject)
			if !ok {
				return nil, AmbiguousObjectError(memberExpr, current)
			}
			indexable, ok := valueObj.(Indexable)
			if !ok {
				return nil, NotIndexableError(memberExpr, resolveChainString, current)
			}
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
		default:
			return nil, InvalidValueExpressionError(memberExpr)
		}
	}
	return current, nil
}
func (st SymbolTable) ValidateValueExpression(expr ast.ValueExpression) (Class, error) {
	resolveChainString, current, err := st.valueExpressionPredicate(expr)
	if err != nil {
		return nil, err
	}

	for _, memberExpr := range expr.Members[1:] {
		switch expr := memberExpr.Init.(type) {
		case string:
			prop, err := current.Get(expr)
			if err != nil {
				return nil, err
			}
			if prop != nil {
				current = prop
			} else if class, ok := current.(ObjectClass); ok {
				if field, ok := class.Fields()[expr]; ok {
					current = field
				} else {
					obj, err := class.Get(expr)
					if err != nil {
						return nil, err
					}
					if obj != nil {
						current = obj
					} else {
						return nil, NoPropertyError(memberExpr, resolveChainString, current, expr)
					}
				}
			}
			resolveChainString += "." + expr
		case ast.CallExpression:
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
		case ast.IndexExpression:
			valueObj, ok := current.(ValueObject)
			if !ok {
				return nil, AmbiguousObjectError(memberExpr, current)
			}
			_, ok = valueObj.(Indexable)
			if !ok {
				return nil, NotIndexableError(memberExpr, resolveChainString, valueObj)
			}
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
			} else {
				if iter, ok := valueObj.(Iterable); ok {
					current = iter.ParentType
				}
			}
			resolveChainString += "[]"
		default:
			return nil, InvalidValueExpressionError(memberExpr)
		}
	}
	if class, ok := current.(Class); ok {
		return class, nil
	} else if valueObj, ok := current.(ValueObject); ok {
		return valueObj.Class(), nil
	}
	return nil, nil
}

// ---
// PROPERTY LIST (MAP OBJECT)
// ---

func (st SymbolTable) ParsePropertyList(props ast.PropertyList) (*MapObject, error) {
	mapObject := NewMapObject()
	for _, prop := range props {
		switch expr := prop.(type) {
		case ast.Property:
			obj, err := st.ResolveValueObject(expr.Init)
			if err != nil {
				return nil, err
			}
			mapObject.Properties[expr.Key] = obj.Class()
			mapObject.Data[expr.Key] = obj
		case ast.SpreadElement:
			obj, err := st.ResolveValueObject(expr.Init)
			if err != nil {
				return nil, err
			}
			objectClass, ok := obj.Class().(ObjectClass)
			if !ok {
				return nil, NodeError(expr, "Cannot spread non-object")
			}
			for key := range objectClass.Fields() {
				val, err := obj.Get(key)
				if err != nil {
					return nil, err
				}
				if val == nil {
					mapObject.Data[key] = NilLiteral{}
				} else if innerValueObj, ok := val.(ValueObject); ok {
					mapObject.Properties[key] = innerValueObj.Class()
					mapObject.Data[key] = innerValueObj
				} else {
					return nil, AmbiguousObjectError(expr, obj)
				}
			}
		default:
			return nil, NodeError(prop, "invalid property list")
		}
	}
	return mapObject, nil
}
func (st SymbolTable) ValidatePropertyList(props ast.PropertyList) (*MapObject, error) {
	mapObject := NewMapObject()
	for _, prop := range props {
		switch expr := prop.(type) {
		case ast.Property:
			obj, err := st.ValidateExpression(expr.Init)
			if err != nil {
				return nil, err
			}
			mapObject.Properties[expr.Key] = obj
		case ast.SpreadElement:
			obj, err := st.ValidateExpression(expr.Init)
			if err != nil {
				return nil, err
			}
			objectClass, ok := obj.(ObjectClass)
			if !ok {
				return nil, NodeError(expr, "Cannot spread non-object")
			}
			for key, value := range objectClass.Fields() {
				mapObject.Properties[key] = value
			}
		default:
			return nil, NodeError(prop, "invalid property list")
		}
	}
	return mapObject, nil
}

// ---
// TYPE EXPRESSIONS
// ---

// Returns the Class assumed by a type expression
func (st SymbolTable) ResolveTypeExpression(expr ast.TypeExpression) (Class, error) {
	parentType, err := st.ResolveSelector(expr.Selector)
	if err != nil {
		return nil, err
	}
	class, ok := parentType.(Class)
	if !ok {
		return nil, InvalidTypeError(expr, parentType)
	}
	if expr.IsPartial {
		objectClass, ok := class.(ObjectClass)
		if !ok {
			return nil, fmt.Errorf("cannot take partial of non-object type %s", parentType)
		}
		class = &PartialObject{ClassObject: objectClass}
	}
	if expr.IsArray {
		class = NewIterable(class, 0)
	}
	if expr.IsOptional {
		class = NewOptionalClass(class)
	}
	return class, nil
}

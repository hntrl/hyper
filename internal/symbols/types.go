package symbols

import (
	"reflect"

	"github.com/hntrl/hyper/internal/tokens"
	"github.com/mitchellh/hashstructure"
)

// TODO: make this interface constrinaed to one of `Object` `ValueObject` or `Class`
type ScopeValue interface{}

// @ 1.1.1 `Object` Type

type Object interface {
	Get(string) (ScopeValue, error)
}

// @ 1.1.2 `ValueObject` Type

type ValueObject interface {
	// The serialized version of the value object
	Value() interface{}
	// The class assumed by the value object
	Class() Class
}

var emptyValueObjectType = reflect.TypeOf((*ValueObject)(nil))

// @ 1.1.3 `Callable` Type

type Callable interface {
	Arguments() []Class
	Returns() Class
	Call(...ValueObject) (ValueObject, error)
}

func ConstructCallableArguments(callable Callable, args []ValueObject) ([]ValueObject, error) {
	callableArgs := callable.Arguments()
	constructedArgs := make([]ValueObject, len(callableArgs))
	if len(args) != len(callableArgs) {
		return nil, MismatchedArgumentLengthError(len(callableArgs), len(args))
	}
	for idx, arg := range args {
		argClass := callableArgs[idx]
		argValue, err := Construct(argClass, arg)
		if err != nil {
			return nil, err
		}
		constructedArgs[idx] = argValue
	}
	return constructedArgs, nil
}
func ValidateCallableArguments(callable Callable, args []Class) error {
	callableArgs := callable.Arguments()
	if len(args) != len(callableArgs) {
		return MismatchedArgumentLengthError(len(callableArgs), len(args))
	}
	for idx, arg := range args {
		argClass := callableArgs[idx]
		err := ShouldConstruct(argClass, arg)
		if err != nil {
			return err
		}
	}
	return nil
}

// @ 1.1.4 `Class` Type

type ClassHash uint64

type Class interface {
	Name() string
	Descriptors() *ClassDescriptors
}

type ClassDescriptors struct {
	Constructors ClassConstructorSet
	Operators    ClassOperatorSet
	Comparators  ClassComparatorSet
	Properties   ClassPropertyMap
	Enumerable   *ClassEnumerationRules
	Prototype    ClassPrototypeMap
}

func classHash(class Class) ClassHash {
	hash, err := hashstructure.Hash(class, nil)
	if err != nil {
		panic(err)
	}
	return ClassHash(hash)
}

func classEquals(a, b Class) bool {
	return classHash(a) == classHash(b)
}

type ClassMethod struct {
	Class         Class
	ArgumentTypes []Class
	ReturnType    Class
	handler       classMethodFn
}

type ClassMethodOptions struct {
	Class     Class
	Arguments []Class
	Returns   Class
	Handler   interface{}
}

func NewClassMethod(opts ClassMethodOptions) *ClassMethod {
	fn, err := makeClassMethodFn(opts.Class, opts.Arguments, opts.Returns, opts.Handler)
	if err != nil {
		panic(err)
	}
	return &ClassMethod{
		Class:         opts.Class,
		ArgumentTypes: opts.Arguments,
		ReturnType:    opts.Returns,
		handler:       fn,
	}
}

func (cm ClassMethod) CallableForValue(val ValueObject) Callable {
	return Function{
		argumentTypes: cm.ArgumentTypes,
		returnType:    cm.ReturnType,
		handler: func(args []ValueObject) (ValueObject, error) {
			argsWithValue := append([]ValueObject{val}, args...)
			return cm.handler(argsWithValue...)
		},
	}
}

type classMethodFn func(...ValueObject) (ValueObject, error)

func makeClassMethodFn(class Class, args []Class, returns Class, callback interface{}) (classMethodFn, error) {
	expectedSignature := callbackSignature{
		args:    make([]reflect.Type, len(args)+1),
		returns: []reflect.Type{},
	}
	expectedSignature.args[0] = emptyValueObjectType
	for idx := range args {
		expectedSignature.args[idx+1] = emptyValueObjectType
	}
	if returns != nil {
		expectedSignature.returns = []reflect.Type{emptyValueObjectType, emptyErrorType}
	} else {
		expectedSignature.returns = []reflect.Type{emptyErrorType}
	}
	cb := newCallback(callback)
	if !cb.AcceptsParameters(expectedSignature) {
		return nil, ExpectedCallbackSignatureError(expectedSignature, cb.Signature)
	}
	return func(args ...ValueObject) (ValueObject, error) {
		argValues := make([]reflect.Value, len(args))
		for idx, arg := range args {
			argValues[idx] = reflect.ValueOf(arg)
		}
		returnValues := cb.Call(argValues)
		if value, ok := returnValues[0].Interface().(ValueObject); ok {
			err := returnValues[1].Interface().(error)
			return value, err
		} else {
			err := returnValues[0].Interface().(error)
			return nil, err
		}
	}, nil
}

// @ 1.1.4.1 `Constructors` Class Descriptor

type ClassConstructorSet []*ClassConstructor

func (set ClassConstructorSet) Get(forClass Class) *ClassConstructor {
	hash := classHash(forClass)
	for _, constructor := range set {
		if constructor.forClass == hash {
			return constructor
		}
	}
	return nil
}

type ClassConstructor struct {
	forClass ClassHash
	handler  classConstructorFn
}

func Constructor(forClass Class, callback interface{}) *ClassConstructor {
	fn, err := makeClassConstructorFn(forClass, callback)
	if err != nil {
		panic(err)
	}
	return &ClassConstructor{
		forClass: classHash(forClass),
		handler:  fn,
	}
}

type classConstructorFn func(ValueObject) (ValueObject, error)

func makeClassConstructorFn(forClass Class, callback interface{}) (classConstructorFn, error) {
	expectedSignature := callbackSignature{
		args:    []reflect.Type{emptyValueObjectType},
		returns: []reflect.Type{emptyValueObjectType, emptyErrorType},
	}
	cb := newCallback(callback)
	if !cb.AcceptsParameters(expectedSignature) {
		return nil, ExpectedCallbackSignatureError(expectedSignature, cb.Signature)
	}
	return func(a ValueObject) (ValueObject, error) {
		args := []reflect.Value{reflect.ValueOf(a)}
		returnValues := cb.Call(args)
		value := returnValues[0].Interface().(ValueObject)
		err := returnValues[1].Interface().(error)
		return value, err
	}, nil
}

// Returns no error if a class can be constructed into another class.
func ShouldConstruct(target, value Class) error {
	if classEquals(target, value) {
		return nil
	}
	if targetArrayClass, ok := target.(ArrayClass); ok {
		// If both the target class and the value's class are an array, redo ShouldConstruct with their parent types
		if valueArrayClass, ok := value.(ArrayClass); ok {
			return ShouldConstruct(targetArrayClass.ItemClass, valueArrayClass.ItemClass)
		}
	}
	if valueMapClass, ok := value.(MapClass); ok {
		// If the value class is a Map, it should construct IF:
		//	1. The target class has a `Properties` descriptor
		//  2. The target class has a map constructor
		//  3. The map doesn't have any properties that dont exist on the target class
		//  4. The map isn't missing any properties that are defined on the target class
		//  5. The map's properties are constructable to the target class's fields
		targetDescriptors := target.Descriptors()
		if targetProperties := targetDescriptors.Properties; targetProperties != nil {
			if constructor := targetDescriptors.Constructors.Get(valueMapClass); constructor != nil {
				for key := range valueMapClass.Properties {
					if _, ok := targetProperties[key]; !ok {
						return UnknownPropertyError(key)
					}
				}
				for key, targetProperty := range targetProperties {
					valuePropertyClass := valueMapClass.Properties[key]
					if valuePropertyClass == nil {
						if _, ok := targetProperty.PropertyClass.(*NilableClass); !ok {
							return MissingPropertyError(key)
						}
					} else {
						err := ShouldConstruct(targetProperty.PropertyClass, valuePropertyClass)
						if err != nil {
							return err
						}
					}
				}
				return nil
			}
		}
	}
	if constructor := target.Descriptors().Constructors.Get(value); constructor != nil {
		// If the target class has a constructor defined for the value's class, it should construct
		return nil
	}
	if properties := value.Descriptors().Properties; properties != nil {
		// If the value class has properties, redo ShouldConstruct as if the value was a map
		propertyClassMap := map[string]Class{}
		for key, attributes := range properties {
			propertyClassMap[key] = attributes.PropertyClass
		}
		return ShouldConstruct(target, MapClass{Properties: propertyClassMap})
	}
	return CannotConstructError(target, value)
}
func Construct(target Class, value ValueObject) (ValueObject, error) {
	if classEquals(target, value.Class()) {
		return value, nil
	}
	if targetArrayClass, ok := target.(ArrayClass); ok {
		if valueArray, ok := value.(*ArrayValue); ok {
			itemClass := targetArrayClass.ItemClass
			newValueArray := NewArray(itemClass, len(valueArray.Items))
			for idx, val := range valueArray.Items {
				constructedVal, err := Construct(itemClass, val)
				if err != nil {
					return nil, err
				}
				newValueArray.Items[idx] = constructedVal
			}
			return newValueArray, nil
		}
	}
	if valueMap, ok := value.(*MapValue); ok {
		targetDescriptors := target.Descriptors()
		if targetProperties := targetDescriptors.Properties; targetProperties != nil {
			if constructor := targetDescriptors.Constructors.Get(valueMap.Class()); constructor != nil {
				for key := range valueMap.Data {
					if _, ok := targetProperties[key]; !ok {
						return nil, UnknownPropertyError(key)
					}
				}
				validationErrors := make(map[string]string)
				propertyClasses := make(map[string]Class)
				data := make(map[string]ValueObject)
				for key, targetProperty := range targetDescriptors.Properties {
					propertyClasses[key] = targetProperty.PropertyClass
					value := valueMap.Data[key]
					if value == nil {
						if targetNilable, ok := targetProperty.PropertyClass.(NilableClass); ok {
							data[key] = &NilableObject{ObjectClass: targetNilable, Object: nil}
						} else {
							validationErrors[key] = MissingPropertyError(key).Error()
						}
					} else {
						constructedValue, err := Construct(targetProperty.PropertyClass, value)
						if err != nil {
							return nil, err
						}
						data[key] = constructedValue
					}
				}
				if len(validationErrors) > 0 {
					return nil, ErrorValue{
						Name: "ValidationError",
						Data: validationErrors,
					}
				}
				return constructor.handler(&MapValue{
					ParentClass: MapClass{
						Properties: propertyClasses,
					},
					Data: data,
				})
			}
		}
	}
	if constructor := target.Descriptors().Constructors.Get(value.Class()); constructor != nil {
		return constructor.handler(value)
	}
	if properties := value.Class().Descriptors().Properties; properties != nil {
		propertyClasses := make(map[string]Class)
		data := make(map[string]ValueObject)
		for key, property := range properties {
			propertyClasses[key] = property.PropertyClass
			propertyValue, err := property.Getter(value)
			if err != nil {
				return nil, err
			}
			data[key] = propertyValue
		}
		return Construct(target, &MapValue{
			ParentClass: MapClass{Properties: propertyClasses},
			Data:        data,
		})
	}
	return nil, CannotConstructError(target, value.Class())
}

// @ 1.1.4.2 `Operators` Class Descriptor

type ClassOperatorSet []*ClassOperator

func (set ClassOperatorSet) Get(operandClass Class, token tokens.Token) *ClassOperator {
	hash := classHash(operandClass)
	for _, operator := range set {
		if operator.operandClass == hash && operator.token == token {
			return operator
		}
	}
	return nil
}

type ClassOperator struct {
	operandClass ClassHash
	token        tokens.Token
	handler      classOperatorFn
}

func Operator(operandClass Class, token tokens.Token, callback interface{}) *ClassOperator {
	fn, err := makeClassOperatorFn(operandClass, callback)
	if err != nil {
		panic(err)
	}
	return &ClassOperator{
		operandClass: classHash(operandClass),
		token:        token,
		handler:      fn,
	}
}

type classOperatorFn func(ValueObject, ValueObject) (ValueObject, error)

func makeClassOperatorFn(operandClass Class, callback interface{}) (classOperatorFn, error) {
	expectedSignature := callbackSignature{
		args:    []reflect.Type{emptyValueObjectType, emptyValueObjectType},
		returns: []reflect.Type{emptyValueObjectType, emptyErrorType},
	}
	cb := newCallback(callback)
	if !cb.AcceptsParameters(expectedSignature) {
		return nil, ExpectedCallbackSignatureError(expectedSignature, cb.Signature)
	}
	return func(a, b ValueObject) (ValueObject, error) {
		args := []reflect.Value{reflect.ValueOf(a), reflect.ValueOf(b)}
		returnValues := cb.Call(args)
		value := returnValues[0].Interface().(ValueObject)
		err := returnValues[1].Interface().(error)
		return value, err
	}, nil
}

func ShouldOperate(token tokens.Token, target, value Class) error {
	targetDescriptors := target.Descriptors()
	if !token.IsOperator() {
		return InvalidOperatorError(token)
	}
	if targetDescriptors.Operators != nil {
		if operator := targetDescriptors.Operators.Get(value, token); operator != nil {
			return nil
		}
	}
	return UndefinedOperatorError(token, target, value)
}
func Operate(token tokens.Token, target, value ValueObject) (ValueObject, error) {
	targetDescriptors := target.Class().Descriptors()
	if !token.IsOperator() {
		return nil, InvalidOperatorError(token)
	}
	if targetDescriptors.Operators != nil {
		if operator := targetDescriptors.Operators.Get(value.Class(), token); operator != nil {
			return operator.handler(target, value)
		}
	}
	return nil, UndefinedOperatorError(token, target.Class(), value.Class())
}

// @ 1.1.4.3 `Comparators` Class Descriptor

type ClassComparatorSet []*ClassComparator

func (set ClassComparatorSet) Get(operandClass Class, token tokens.Token) *ClassComparator {
	hash := classHash(operandClass)
	for _, comparator := range set {
		if comparator.operandClass == hash && comparator.token == token {
			return comparator
		}
	}
	return nil
}

type ClassComparator struct {
	operandClass ClassHash
	token        tokens.Token
	handler      classComparatorFn
}

func Comparator(operandClass Class, token tokens.Token, callback interface{}) *ClassComparator {
	hash := classHash(operandClass)
	fn, err := makeClassComparatorFn(operandClass, callback)
	if err != nil {
		panic(err)
	}
	return &ClassComparator{
		operandClass: hash,
		token:        token,
		handler:      fn,
	}
}

type classComparatorFn func(ValueObject, ValueObject) (bool, error)

func makeClassComparatorFn(operandClass Class, callback interface{}) (classComparatorFn, error) {
	expectedSignature := callbackSignature{
		args:    []reflect.Type{emptyValueObjectType, emptyValueObjectType},
		returns: []reflect.Type{emptyBoolType, emptyErrorType},
	}
	cb := newCallback(callback)
	if !cb.AcceptsParameters(expectedSignature) {
		return nil, ExpectedCallbackSignatureError(expectedSignature, cb.Signature)
	}
	return func(a, b ValueObject) (bool, error) {
		args := []reflect.Value{reflect.ValueOf(a), reflect.ValueOf(b)}
		returnValues := cb.Call(args)
		value := returnValues[0].Interface().(bool)
		err := returnValues[1].Interface().(error)
		return value, err
	}, nil
}

func ShouldCompare(token tokens.Token, target, value Class) error {
	targetDescriptors := target.Descriptors()
	if !token.IsComparableOperator() {
		return InvalidCompareOperatorError(token)
	}
	if targetDescriptors.Comparators != nil {
		if comparator := targetDescriptors.Comparators.Get(value, token); comparator != nil {
			return nil
		}
	}
	return UndefinedOperatorError(token, target, value)
}
func Compare(token tokens.Token, target, value ValueObject) (bool, error) {
	targetDescriptors := target.Class().Descriptors()
	if !token.IsComparableOperator() {
		return false, InvalidCompareOperatorError(token)
	}
	if targetDescriptors.Comparators != nil {
		if comparator := targetDescriptors.Comparators.Get(value.Class(), token); comparator != nil {
			return comparator.handler(target, value)
		}
	}
	return false, UndefinedOperatorError(token, target.Class(), value.Class())
}

// @ 1.1.4.4 `Properties` Class Descriptor

type ClassPropertyMap map[string]ClassPropertyAttributes

type ClassPropertyAttributes struct {
	PropertyClass Class
	Getter        ClassGetterMethod
	Setter        ClassSetterMethod
}

type PropertyAttributesOptions struct {
	Class  Class
	Getter interface{}
	Setter interface{}
}

func PropertyAttributes(opts PropertyAttributesOptions) ClassPropertyAttributes {
	attributes := ClassPropertyAttributes{PropertyClass: opts.Class}
	if opts.Getter != nil {
		getterMethod, err := makeClassGetterMethod(opts.Class, opts.Getter)
		if err != nil {
			panic(err)
		}
		attributes.Getter = getterMethod
	}
	if opts.Setter != nil {
		setterMethod, err := makeClassSetterMethod(opts.Class, opts.Setter)
		if err != nil {
			panic(err)
		}
		attributes.Setter = setterMethod
	}
	return attributes
}

type ClassGetterMethod func(ValueObject) (ValueObject, error)
type ClassSetterMethod func(ValueObject, ValueObject) error

func makeClassGetterMethod(propertyClass Class, callback interface{}) (ClassGetterMethod, error) {
	expectedSignature := callbackSignature{
		args:    []reflect.Type{emptyValueObjectType},
		returns: []reflect.Type{emptyValueObjectType, emptyErrorType},
	}
	cb := newCallback(callback)
	if !cb.AcceptsParameters(expectedSignature) {
		return nil, ExpectedCallbackSignatureError(expectedSignature, cb.Signature)
	}
	return func(a ValueObject) (ValueObject, error) {
		args := []reflect.Value{reflect.ValueOf(a)}
		returnValues := cb.Call(args)
		value := returnValues[0].Interface().(ValueObject)
		err := returnValues[1].Interface().(error)
		return value, err
	}, nil
}
func makeClassSetterMethod(propertyClass Class, callback interface{}) (ClassSetterMethod, error) {
	expectedSignature := callbackSignature{
		args:    []reflect.Type{emptyValueObjectType, emptyValueObjectType},
		returns: []reflect.Type{emptyErrorType},
	}
	cb := newCallback(callback)
	if !cb.AcceptsParameters(expectedSignature) {
		return nil, ExpectedCallbackSignatureError(expectedSignature, cb.Signature)
	}
	return func(a, b ValueObject) error {
		args := []reflect.Value{reflect.ValueOf(a), reflect.ValueOf(b)}
		returnValues := cb.Call(args)
		err := returnValues[0].Interface().(error)
		return err
	}, nil
}

// @ 1.1.4.5 `Enumerable` Class Descriptor

type ClassEnumerationRules struct {
	GetLength EnumerableGetLengthMethod
	GetIndex  EnumerableGetIndexMethod
	SetIndex  EnumerableSetIndexMethod
	GetRange  EnumerableGetRangeMethod
	SetRange  EnumerableSetRangeMethod
}

type ClassEnumerationOptions struct {
	GetLength interface{}
	GetIndex  interface{}
	SetIndex  interface{}
	GetRange  interface{}
	SetRange  interface{}
}

func NewClassEnumerationRules(opts ClassEnumerationOptions) *ClassEnumerationRules {
	rules := &ClassEnumerationRules{}
	if opts.GetLength != nil {
		getLengthFn, err := makeEnumerableGetLengthMethod(opts.GetLength)
		if err != nil {
			panic(err)
		}
		rules.GetLength = getLengthFn
	}
	if opts.GetIndex != nil {
		getIndexFn, err := makeEnumerableGetIndexMethod(opts.GetIndex)
		if err != nil {
			panic(err)
		}
		rules.GetIndex = getIndexFn
	}
	if opts.SetIndex != nil {
		setIndexFn, err := makeEnumerableSetIndexMethod(opts.SetIndex)
		if err != nil {
			panic(err)
		}
		rules.SetIndex = setIndexFn
	}
	if opts.GetRange != nil {
		getRangeFn, err := makeEnumerableGetRangeMethod(opts.GetRange)
		if err != nil {
			panic(err)
		}
		rules.GetRange = getRangeFn
	}
	if opts.SetRange != nil {
		setRangeFn, err := makeEnumerableSetRangeMethod(opts.SetRange)
		if err != nil {
			panic(err)
		}
		rules.SetRange = setRangeFn
	}
	return rules
}

type EnumerableGetLengthMethod func(ValueObject) (int, error)
type EnumerableGetIndexMethod func(ValueObject, int) (ValueObject, error)
type EnumerableSetIndexMethod func(ValueObject, int, ValueObject) error
type EnumerableGetRangeMethod func(ValueObject, int, int) (ValueObject, error)
type EnumerableSetRangeMethod func(ValueObject, int, int, ValueObject) error

func makeEnumerableGetLengthMethod(callback interface{}) (EnumerableGetLengthMethod, error) {
	expectedSignature := callbackSignature{
		args:    []reflect.Type{emptyValueObjectType},
		returns: []reflect.Type{emptyIntType, emptyErrorType},
	}
	cb := newCallback(callback)
	if !cb.AcceptsParameters(expectedSignature) {
		return nil, ExpectedCallbackSignatureError(expectedSignature, cb.Signature)
	}
	return func(a ValueObject) (int, error) {
		args := []reflect.Value{reflect.ValueOf(a)}
		returnValues := cb.Call(args)
		value := returnValues[0].Interface().(int)
		err := returnValues[1].Interface().(error)
		return value, err
	}, nil
}
func makeEnumerableGetIndexMethod(callback interface{}) (EnumerableGetIndexMethod, error) {
	expectedSignature := callbackSignature{
		args:    []reflect.Type{emptyValueObjectType, emptyIntType},
		returns: []reflect.Type{emptyValueObjectType, emptyErrorType},
	}
	cb := newCallback(callback)
	if !cb.AcceptsParameters(expectedSignature) {
		return nil, ExpectedCallbackSignatureError(expectedSignature, cb.Signature)
	}
	return func(a ValueObject, b int) (ValueObject, error) {
		args := []reflect.Value{reflect.ValueOf(a), reflect.ValueOf(b)}
		returnValues := cb.Call(args)
		value := returnValues[0].Interface().(ValueObject)
		err := returnValues[1].Interface().(error)
		return value, err
	}, nil
}
func makeEnumerableSetIndexMethod(callback interface{}) (EnumerableSetIndexMethod, error) {
	expectedSignature := callbackSignature{
		args:    []reflect.Type{emptyValueObjectType, emptyIntType, emptyValueObjectType},
		returns: []reflect.Type{emptyErrorType},
	}
	cb := newCallback(callback)
	if !cb.AcceptsParameters(expectedSignature) {
		return nil, ExpectedCallbackSignatureError(expectedSignature, cb.Signature)
	}
	return func(a ValueObject, b int, c ValueObject) error {
		args := []reflect.Value{reflect.ValueOf(a), reflect.ValueOf(b), reflect.ValueOf(c)}
		returnValues := cb.Call(args)
		err := returnValues[0].Interface().(error)
		return err
	}, nil
}
func makeEnumerableGetRangeMethod(callback interface{}) (EnumerableGetRangeMethod, error) {
	expectedSignature := callbackSignature{
		args:    []reflect.Type{emptyValueObjectType, emptyIntType, emptyIntType},
		returns: []reflect.Type{emptyValueObjectType, emptyErrorType},
	}
	cb := newCallback(callback)
	if !cb.AcceptsParameters(expectedSignature) {
		return nil, ExpectedCallbackSignatureError(expectedSignature, cb.Signature)
	}
	return func(a ValueObject, b int, c int) (ValueObject, error) {
		args := []reflect.Value{reflect.ValueOf(a), reflect.ValueOf(b), reflect.ValueOf(c)}
		returnValues := cb.Call(args)
		value := returnValues[0].Interface().(ValueObject)
		err := returnValues[1].Interface().(error)
		return value, err
	}, nil
}
func makeEnumerableSetRangeMethod(callback interface{}) (EnumerableSetRangeMethod, error) {
	expectedSignature := callbackSignature{
		args:    []reflect.Type{emptyValueObjectType, emptyIntType, emptyIntType, emptyValueObjectType},
		returns: []reflect.Type{emptyErrorType},
	}
	cb := newCallback(callback)
	if !cb.AcceptsParameters(expectedSignature) {
		return nil, ExpectedCallbackSignatureError(expectedSignature, cb.Signature)
	}
	return func(a ValueObject, b int, c int, d ValueObject) error {
		args := []reflect.Value{reflect.ValueOf(a), reflect.ValueOf(b), reflect.ValueOf(c), reflect.ValueOf(d)}
		returnValues := cb.Call(args)
		err := returnValues[0].Interface().(error)
		return err
	}, nil
}

// @ 1.1.4.6 `Prototype` Class Descriptor

type ClassPrototypeMap map[string]*ClassMethod

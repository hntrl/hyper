package symbols

import (
	"fmt"
	"math"

	. "github.com/hntrl/hyper/internal/symbols/errors"
	"github.com/hntrl/hyper/internal/tokens"
)

// @ 2.2.1 `nil` Primitive

var (
	Nil            = NilClass{}
	NilDescriptors = &ClassDescriptors{}
)

type NilClass struct{}

func (NilClass) Name() string {
	return "<nil>"
}

func (NilClass) Descriptors() *ClassDescriptors {
	return NilDescriptors
}

type NilValue struct{}

func (NilValue) Value() interface{} {
	return nil
}
func (NilValue) Class() Class {
	return Nil
}

// @ 2.2.2 `Boolean` Primitive

var (
	Boolean            = BooleanClass{}
	BooleanDescriptors = &ClassDescriptors{
		Constructors: ClassConstructorSet{},
		Comparators: ClassComparatorSet{
			Comparator(Boolean, tokens.AND, func(a, b BooleanValue) (bool, error) {
				return bool(a) && bool(b), nil
			}),
			Comparator(Boolean, tokens.OR, func(a, b BooleanValue) (bool, error) {
				return bool(a) || bool(b), nil
			}),
		},
	}
)

type BooleanClass struct{}

func (BooleanClass) Name() string {
	return "Boolean"
}

func (BooleanClass) Descriptors() *ClassDescriptors {
	return BooleanDescriptors
}

type BooleanValue bool

func (v BooleanValue) Value() interface{} {
	return bool(v)
}
func (BooleanValue) Class() Class {
	return Boolean
}

// @ 2.2.3 `String Primitive`

var (
	String            = StringClass{}
	StringDescriptors = &ClassDescriptors{
		Constructors: ClassConstructorSet{
			Constructor(Number, func(val NumberValue) (StringValue, error) {
				return "", nil
			}),
			Constructor(Double, func(val DoubleValue) (StringValue, error) {
				return "", nil
			}),
			Constructor(Float, func(val FloatValue) (StringValue, error) {
				return "", nil
			}),
			Constructor(Integer, func(val IntegerValue) (StringValue, error) {
				return "", nil
			}),
			Constructor(Boolean, func(val BooleanValue) (StringValue, error) {
				return "", nil
			}),
		},
		Operators:   ClassOperatorSet{},
		Comparators: ClassComparatorSet{},
		Prototype:   ClassPrototypeMap{},
	}
)

type StringClass struct{}

func (StringClass) Name() string {
	return "String"
}

func (StringClass) Descriptors() *ClassDescriptors {
	return StringDescriptors
}

type StringValue string

func (v StringValue) Value() interface{} {
	return string(v)
}
func (StringValue) Class() Class {
	return String
}

// @ 2.2.4 Numeric Primitives

var NumericClasses = []Class{Number, Double, Integer, Float}

func numericOperatorPredicate(numberConstructor *ClassConstructor, operandConstructor *ClassConstructor, cb func(float64, float64) float64) classOperatorFn {
	return func(a, b ValueObject) (ValueObject, error) {
		na, err := numberConstructor.handler(a)
		if err != nil {
			return nil, err
		}
		nb, err := numberConstructor.handler(b)
		if err != nil {
			return nil, err
		}
		result := cb(float64(na.(NumberValue)), float64(nb.(NumberValue)))
		return operandConstructor.handler(NumberValue(result))
	}
}

func numericComparatorPredicate(constructor *ClassConstructor, cb func(float64, float64) bool) classComparatorFn {
	return func(a, b ValueObject) (bool, error) {
		na, err := constructor.handler(a)
		if err != nil {
			return false, err
		}
		nb, err := constructor.handler(b)
		if err != nil {
			return false, err
		}
		return cb(float64(na.(NumberValue)), float64(nb.(NumberValue))), nil
	}
}

var NumericOperators = ClassOperatorSet{}
var NumericComparators = ClassComparatorSet{}

func init() {
	for _, operandClass := range NumericClasses {
		numberConstructor := NumberDescriptors.Constructors.Get(operandClass)
		operandConstructor := operandClass.Descriptors().Constructors.Get(Number)
		NumericOperators = append(NumericOperators, ClassOperatorSet{
			Operator(operandClass, tokens.ADD, numericOperatorPredicate(numberConstructor, operandConstructor, func(a, b float64) float64 {
				return a + b
			})),
			Operator(operandClass, tokens.SUB, numericOperatorPredicate(numberConstructor, operandConstructor, func(a, b float64) float64 {
				return a - b
			})),
			Operator(operandClass, tokens.MUL, numericOperatorPredicate(numberConstructor, operandConstructor, func(a, b float64) float64 {
				return a * b
			})),
			Operator(operandClass, tokens.PWR, numericOperatorPredicate(numberConstructor, operandConstructor, func(a, b float64) float64 {
				return math.Pow(a, b)
			})),
			Operator(operandClass, tokens.QUO, numericOperatorPredicate(numberConstructor, operandConstructor, func(a, b float64) float64 {
				return a / b
			})),
			Operator(operandClass, tokens.REM, numericOperatorPredicate(numberConstructor, operandConstructor, func(a, b float64) float64 {
				return math.Mod(a, b)
			})),
		}...)
		NumericComparators = append(NumericComparators, ClassComparatorSet{
			Comparator(operandClass, tokens.EQUALS, numericComparatorPredicate(numberConstructor, func(a, b float64) bool {
				return a == b
			})),
			Comparator(operandClass, tokens.NOT_EQUALS, numericComparatorPredicate(numberConstructor, func(a, b float64) bool {
				return a != b
			})),
			Comparator(operandClass, tokens.LESS, numericComparatorPredicate(numberConstructor, func(a, b float64) bool {
				return a < b
			})),
			Comparator(operandClass, tokens.GREATER, numericComparatorPredicate(numberConstructor, func(a, b float64) bool {
				return a > b
			})),
			Comparator(operandClass, tokens.LESS_EQUAL, numericComparatorPredicate(numberConstructor, func(a, b float64) bool {
				return a <= b
			})),
			Comparator(operandClass, tokens.GREATER_EQUAL, numericComparatorPredicate(numberConstructor, func(a, b float64) bool {
				return a >= b
			})),
		}...)
	}
}

var (
	Number            = NumberClass{}
	NumberDescriptors = &ClassDescriptors{
		Constructors: ClassConstructorSet{
			Constructor(Double, func(val DoubleValue) (NumberValue, error) {
				return NumberValue(val), nil
			}),
			Constructor(Float, func(val FloatValue) (NumberValue, error) {
				return NumberValue(val), nil
			}),
			Constructor(Integer, func(val IntegerValue) (NumberValue, error) {
				// ? is this allowed
				return NumberValue(val), nil
			}),
		},
		Operators:   NumericOperators,
		Comparators: NumericComparators,
	}
)

type NumberClass struct{}

func (NumberClass) Name() string {
	return "Number"
}

func (NumberClass) Descriptors() *ClassDescriptors {
	return NumberDescriptors
}

type NumberValue float64

func (v NumberValue) Value() interface{} {
	return float64(v)
}
func (NumberValue) Class() Class {
	return Number
}

var (
	Double            = DoubleClass{}
	DoubleDescriptors = &ClassDescriptors{
		Constructors: ClassConstructorSet{
			Constructor(Number, func(val NumberValue) (DoubleValue, error) {
				return DoubleValue(val), nil
			}),
			Constructor(Float, func(val FloatValue) (DoubleValue, error) {
				return DoubleValue(val), nil
			}),
			Constructor(Integer, func(val IntegerValue) (DoubleValue, error) {
				return DoubleValue(val), nil
			}),
		},
		Operators:   NumericOperators,
		Comparators: NumericComparators,
	}
)

type DoubleClass struct{}

func (DoubleClass) Name() string {
	return "Double"
}
func (DoubleClass) Descriptors() *ClassDescriptors {
	return DoubleDescriptors
}

type DoubleValue float64

func (v DoubleValue) Value() interface{} {
	return float64(v)
}
func (DoubleValue) Class() Class {
	return Double
}

var (
	Float            = FloatClass{}
	FloatDescriptors = &ClassDescriptors{
		Constructors: ClassConstructorSet{
			Constructor(Number, func(val NumberValue) (FloatValue, error) {
				return FloatValue(val), nil
			}),
			Constructor(Double, func(val DoubleValue) (FloatValue, error) {
				return FloatValue(val), nil
			}),
			Constructor(Integer, func(val IntegerValue) (FloatValue, error) {
				return FloatValue(val), nil
			}),
		},
		Operators:   NumericOperators,
		Comparators: NumericComparators,
	}
)

type FloatClass struct{}

func (FloatClass) Name() string {
	return "Float"
}
func (FloatClass) Descriptors() *ClassDescriptors {
	return FloatDescriptors
}

type FloatValue float64

func (v FloatValue) Value() interface{} {
	return float64(v)
}
func (FloatValue) Class() Class {
	return Float
}

var (
	Integer            = IntegerClass{}
	IntegerDescriptors = &ClassDescriptors{
		Constructors: ClassConstructorSet{
			Constructor(Number, func(val NumberValue) (IntegerValue, error) {
				return IntegerValue(val), nil
			}),
			Constructor(Double, func(val DoubleValue) (IntegerValue, error) {
				return IntegerValue(val), nil
			}),
			Constructor(Float, func(val FloatValue) (IntegerValue, error) {
				return IntegerValue(val), nil
			}),
		},
		Operators:   NumericOperators,
		Comparators: NumericComparators,
	}
)

type IntegerClass struct{}

func (IntegerClass) Name() string {
	return "Integer"
}
func (IntegerClass) Descriptors() *ClassDescriptors {
	return IntegerDescriptors
}

type IntegerValue int64

func (v IntegerValue) Value() interface{} {
	return int64(v)
}
func (IntegerValue) Class() Class {
	return Integer
}

// @ 2.2.5 `Map` Object

var Map = MapClass{}

type MapClass struct {
	Properties map[string]Class `hash:"ignore"`
}

func (MapClass) Name() string {
	return "Map"
}
func (mc MapClass) Descriptors() *ClassDescriptors {
	propertyMap := ClassPropertyMap{}
	for key, val := range mc.Properties {
		propertyMap[key] = PropertyAttributes(PropertyOptions{
			Class: val,
			Getter: func(obj *MapValue) (ValueObject, error) {
				return obj.Get(key), nil
			},
			Setter: func(obj *MapValue, val ValueObject) error {
				obj.Set(key, val)
				return nil
			},
		})
	}
	return &ClassDescriptors{
		Properties: propertyMap,
	}
}

func NewMapClass() MapClass {
	return MapClass{
		Properties: map[string]Class{},
	}
}

type MapValue struct {
	parentClass MapClass
	data        map[string]ValueObject
}

func NewMapValue() *MapValue {
	return &MapValue{
		parentClass: NewMapClass(),
		data:        map[string]ValueObject{},
	}
}

func (mv *MapValue) Class() Class {
	return mv.parentClass
}
func (mv *MapValue) Value() interface{} {
	out := make(map[string]interface{})
	for key, value := range mv.data {
		out[key] = value.Value()
	}
	return out
}

func (mv *MapValue) Get(k string) ValueObject {
	return mv.data[k]
}
func (mv *MapValue) Set(k string, v ValueObject) {
	mv.parentClass.Properties[k] = v.Class()
	mv.data[k] = v
}
func (mv *MapValue) Map() map[string]ValueObject {
	return mv.data
}

// @ 2.2.6 `Nilable` Object

type NilableClass struct {
	parentClass Class
	descriptors *ClassDescriptors `hash:"ignore"`
}

func NewNilableClass(parentClass Class) NilableClass {
	parentDescriptors := parentClass.Descriptors()
	nilableClass := NilableClass{parentClass: parentClass}
	nilableClass.descriptors = &ClassDescriptors{
		Constructors: ClassConstructorSet{
			Constructor(Nil, func() (*NilableValue, error) {
				return &NilableValue{
					nilableClass: nilableClass,
					setValue:     nil,
				}, nil
			}),
		},
		Comparators: ClassComparatorSet{
			Comparator(Nil, tokens.EQUALS, func(a *NilableValue) (bool, error) {
				return a.setValue == nil, nil
			}),
			Comparator(Nil, tokens.NOT_EQUALS, func(a *NilableValue) (bool, error) {
				return a.setValue != nil, nil
			}),
		},
		Properties: parentDescriptors.Properties,
	}
	if parentDescriptors.Constructors != nil {
		for hash, constructor := range parentDescriptors.Constructors {
			nilableClass.descriptors.Constructors[hash] = &ClassConstructor{
				forClass: constructor.forClass,
				handler: func(val ValueObject) (ValueObject, error) {
					constructedVal, err := constructor.handler(val)
					if err != nil {
						return nil, err
					}
					return &NilableValue{nilableClass, constructedVal}, nil
				},
			}
		}
	}
	if parentDescriptors.Operators != nil {
		parentDescriptors.Operators = ClassOperatorSet{}
		for hash, operator := range parentDescriptors.Operators {
			nilableClass.descriptors.Operators[hash] = &ClassOperator{
				operandClass: operator.operandClass,
				token:        operator.token,
				handler: func(a ValueObject, b ValueObject) (ValueObject, error) {
					nilable := a.(*NilableValue)
					if nilable.setValue == nil {
						return nil, StandardError(CannotOperateNilValue, "cannot operate on nil value")
					}
					computedVal, err := operator.handler(nilable.setValue, b)
					if err != nil {
						return nil, err
					}
					return &NilableValue{nilableClass, computedVal}, nil
				},
			}
		}
	}
	if parentDescriptors.Comparators != nil {
		for hash, comparator := range parentDescriptors.Comparators {
			nilableClass.descriptors.Comparators[hash] = &ClassComparator{
				operandClass: comparator.operandClass,
				token:        comparator.token,
				handler: func(a ValueObject, b ValueObject) (bool, error) {
					nilable := a.(*NilableValue)
					if nilable.setValue == nil {
						return false, StandardError(CannotOperateNilValue, "cannot operate on nil value")
					}
					return comparator.handler(nilable.setValue, b)
				},
			}
		}
	}
	if parentDescriptors.Enumerable != nil {
		nilableClass.descriptors.Enumerable = &ClassEnumerationRules{}
		rules := parentDescriptors.Enumerable
		if rules.GetLength != nil {
			nilableClass.descriptors.Enumerable.GetLength = func(a ValueObject) (int, error) {
				nilable := a.(*NilableValue)
				if nilable.setValue == nil {
					return -1, StandardError(CannotEnumerateNilValue, "cannot get length of nil")
				}
				return rules.GetLength(nilable.setValue)
			}
		}
		if rules.GetIndex != nil {
			nilableClass.descriptors.Enumerable.GetIndex = func(a ValueObject, idx int) (ValueObject, error) {
				nilable := a.(*NilableValue)
				if nilable.setValue == nil {
					return nil, StandardError(CannotEnumerateNilValue, "cannot get index of nil")
				}
				return rules.GetIndex(nilable.setValue, idx)
			}
		}
		if rules.SetIndex != nil {
			nilableClass.descriptors.Enumerable.SetIndex = func(a ValueObject, idx int, b ValueObject) error {
				nilable := a.(*NilableValue)
				if nilable.setValue == nil {
					return StandardError(CannotEnumerateNilValue, "cannot set length of nil")
				}
				return rules.SetIndex(a, idx, b)
			}
		}
		if rules.GetRange != nil {
			nilableClass.descriptors.Enumerable.GetRange = func(a ValueObject, start int, end int) (ValueObject, error) {
				nilable := a.(*NilableValue)
				if nilable.setValue == nil {
					return nil, StandardError(CannotEnumerateNilValue, "cannot get range of nil")
				}
				return rules.GetRange(nilable.setValue, start, end)
			}
		}
		if rules.SetRange != nil {
			nilableClass.descriptors.Enumerable.SetRange = func(a ValueObject, start int, end int, b ValueObject) error {
				nilable := a.(*NilableValue)
				if nilable.setValue == nil {
					return StandardError(CannotEnumerateNilValue, "cannot set range of nil")
				}
				return rules.SetRange(nilable.setValue, start, end, b)
			}
		}
	}
	return nilableClass
}

func (nc NilableClass) Name() string {
	return fmt.Sprintf("%s?", nc.parentClass.Name())
}

func (nc NilableClass) Descriptors() *ClassDescriptors {
	return nc.descriptors
}

type NilableValue struct {
	nilableClass NilableClass
	setValue     ValueObject
}

func NewNilableValue(parentClass Class, value ValueObject) *NilableValue {
	return &NilableValue{
		nilableClass: NewNilableClass(parentClass),
		setValue:     value,
	}
}

func (no *NilableValue) Class() Class {
	return no.nilableClass
}
func (no *NilableValue) Value() interface{} {
	if no.setValue == nil {
		return nil
	}
	return no.setValue.Value()
}
func (no *NilableValue) ValueObject() ValueObject {
	return no.setValue
}

// @ 2.2.7 `Array` Object

type ArrayClass struct {
	itemClass   Class
	descriptors *ClassDescriptors `hash:"ignore"`
}

// TODO: "cache" the creation of this array class to reduce reflection calls
func NewArrayClass(itemClass Class) ArrayClass {
	arrayClass := ArrayClass{
		itemClass: itemClass,
	}
	arrayClass.descriptors = &ClassDescriptors{
		Constructors: ClassConstructorSet{
			Constructor(arrayClass, func(arr ArrayValue) (ArrayValue, error) {
				newArray := ArrayValue{parentClass: arr.parentClass, items: make([]ValueObject, len(arr.items))}
				copy(arr.items, newArray.items)
				return newArray, nil
			}),
		},
		Enumerable: NewClassEnumerationRules(ClassEnumerationOptions{
			GetIndex: func(arr *ArrayValue, index int) (ValueObject, error) {
				if index > len(arr.items) || index < 0 {
					return nil, StandardError(IndexOutOfRange, "index out of range")
				}
				return arr.items[index], nil
			},
			SetIndex: func(arr *ArrayValue, index int, item ValueObject) error {
				if index > len(arr.items) || index < 0 {
					return StandardError(IndexOutOfRange, "index out of range")
				}
				arr.items[index] = item
				return nil
			},
			GetRange: func(arr *ArrayValue, start int, end int) (*ArrayValue, error) {
				if start > len(arr.items) || start < 0 {
					return nil, StandardError(IndexOutOfRange, "start index out of range")
				}
				if end > len(arr.items) || end < 0 {
					return nil, StandardError(IndexOutOfRange, "end index out of range")
				}
				shouldReverseOrder := start > end
				if shouldReverseOrder {
					start := int(math.Min(float64(start), float64(end)))
					end := int(math.Max(float64(start), float64(end)))
					s := make([]ValueObject, len(arr.items))
					// thank you s/o :) https://stackoverflow.com/a/19239850
					copy(s, arr.items)
					for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
						s[i], s[j] = s[j], s[i]
					}
					return &ArrayValue{
						parentClass: arr.parentClass,
						items:       s[start:end],
					}, nil
				} else {
					return &ArrayValue{
						parentClass: arr.parentClass,
						items:       arr.items[start:end],
					}, nil
				}
			},
			SetRange: func(arr *ArrayValue, start int, end int, insertArr ArrayValue) error {
				if start > len(arr.items) || start < 0 {
					return StandardError(IndexOutOfRange, "start index out of range")
				}
				if end > len(arr.items) || end < 0 {
					return StandardError(IndexOutOfRange, "end index out of range")
				}
				if start > end {
					return StandardError(InvalidRangeIndices, "start index cannot be greater than end index")
				}
				tailItems := arr.items[end:len(arr.items)]
				arr.items = append(arr.items[0:start], insertArr.items...)
				arr.items = append(arr.items, tailItems...)
				return nil
			},
		}),
		Prototype: ClassPrototypeMap{
			"append": NewClassMethod(ClassMethodOptions{
				Class:     arrayClass,
				Arguments: []Class{arrayClass.itemClass},
				Returns:   nil,
				Handler: func(arr *ArrayValue, item ValueObject) error {
					arr.items = append(arr.items, item)
					return nil
				},
			}),
			"length": NewClassMethod(ClassMethodOptions{
				Class:     arrayClass,
				Arguments: []Class{},
				Returns:   Integer,
				Handler: func(arr *ArrayValue) (IntegerValue, error) {
					return IntegerValue(len(arr.items)), nil
				},
			}),
		},
	}
	return arrayClass
}

func (ac ArrayClass) Name() string {
	return fmt.Sprintf("[]%s", ac.itemClass.Name())
}

func (ac ArrayClass) Descriptors() *ClassDescriptors {
	return ac.descriptors
}

type ArrayValue struct {
	parentClass ArrayClass
	items       []ValueObject `hash:"ignore"`
}

func NewArray(itemClass Class, len int) *ArrayValue {
	arrayClass := NewArrayClass(itemClass)
	return &ArrayValue{
		parentClass: arrayClass,
		items:       make([]ValueObject, len),
	}
}

func (av *ArrayValue) Class() Class {
	return av.parentClass
}

func (av *ArrayValue) Value() interface{} {
	out := make([]interface{}, len(av.items))
	for i, item := range av.items {
		out[i] = item.Value()
	}
	return out
}

// @ 2.2.8 `Error` Object

var (
	Error            = ErrorClass{}
	ErrorDescriptors = &ClassDescriptors{
		Properties: ClassPropertyMap{
			"name": PropertyAttributes(PropertyOptions{
				Class: String,
				Getter: func(a ErrorValue) (StringValue, error) {
					return StringValue(a.Name), nil
				},
			}),
			"message": PropertyAttributes(PropertyOptions{
				Class: String,
				Getter: func(a ErrorValue) (StringValue, error) {
					return StringValue(a.Message), nil
				},
			}),
		},
	}
)

type ErrorClass struct{}

func (ErrorClass) Name() string {
	return "Error"
}
func (ErrorClass) Descriptors() *ClassDescriptors {
	return ErrorDescriptors
}

type ErrorValue struct {
	Name    string
	Message string
	Data    interface{}
}

func (ErrorValue) Class() Class {
	return Error
}
func (ev ErrorValue) Value() interface{} {
	return ev
}
func (ev ErrorValue) Error() string {
	return fmt.Sprintf("%s: %s", ev.Name, ev.Message)
}

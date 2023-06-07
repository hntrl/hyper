package symbols

import (
	"math"
	"time"

	"github.com/hntrl/hyper/internal/tokens"
)

// @ 1.2.1 `nil` Primitive

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

// @ 1.2.2 `Boolean` Primitive

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

// @ 1.2.3 `String Primitive`

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

// @ 1.2.4 Numeric Primitives

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

// @ 1.2.4.1 `Number` Primitive

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

// @ 1.2.4.2 `Double` Primitive

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

// @ 1.2.4.3 `Float` Primitive

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

// @ 1.2.4.4 `Integer` Primitive

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

// @ 1.2.5 `DateTime` Primitive

var (
	DateTime            = DateTimeClass{}
	DateTimeDescriptors = &ClassDescriptors{
		Constructors: ClassConstructorSet{},
		Operators:    ClassOperatorSet{},
		Comparators:  ClassComparatorSet{},
		Prototype: ClassPrototypeMap{
			"format": NewClassMethod(ClassMethodOptions{
				Class:     DateTime,
				Arguments: []Class{String},
				Returns:   String,
				Handler: func(dt *DateTimeValue, fmt StringValue) (StringValue, error) {
					// TODO: write this
					return "", nil
				},
			}),
		},
	}
)

type DateTimeClass struct{}

func (DateTimeClass) Name() string {
	return "DateTime"
}
func (DateTimeClass) Descriptors() *ClassDescriptors {
	return DateTimeDescriptors
}

type DateTimeValue struct {
	t time.Time
}

func (v DateTimeValue) Value() interface{} {
	return map[string]string{"$date": v.t.Format(time.RFC3339)}
}
func (DateTimeValue) Class() Class {
	return DateTime
}

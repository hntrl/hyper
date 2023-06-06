package symbols

import (
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
		Operators:   ClassOperatorSet{},
		Comparators: ClassComparatorSet{},
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
		Operators:   ClassOperatorSet{},
		Comparators: ClassComparatorSet{},
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
		Operators:   ClassOperatorSet{},
		Comparators: ClassComparatorSet{},
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
		Operators:   ClassOperatorSet{},
		Comparators: ClassComparatorSet{},
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

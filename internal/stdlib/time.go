package stdlib

import (
	"time"

	"github.com/hntrl/hyper/internal/symbols"
	"github.com/hntrl/hyper/internal/tokens"
)

const (
	microsecond = DurationValue(1)
	millisecond = DurationValue(microsecond * 1000)
	second      = DurationValue(millisecond * 1000)
	minute      = DurationValue(second * 60)
	hour        = DurationValue(minute * 60)
)

type TimePackage struct{}

func (tp TimePackage) Get(key string) (symbols.ScopeValue, error) {
	objects := map[string]symbols.ScopeValue{
		"DateTime":    DateTime,
		"Duration":    Duration,
		"Microsecond": microsecond,
		"Millisecond": millisecond,
		"Second":      second,
		"Minute":      minute,
		"Hour":        hour,
		"now": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{},
			Returns:   DateTime,
			Handler: func() (DateTimeValue, error) {
				return DateTimeValue{t: time.Now()}, nil
			},
		}),
	}
	return objects[key], nil
}

var (
	DateTime            = DateTimeClass{}
	DateTimeDescriptors = &symbols.ClassDescriptors{
		Name:         "DateTime",
		Constructors: symbols.ClassConstructorSet{},
		Operators:    symbols.ClassOperatorSet{},
		Comparators:  symbols.ClassComparatorSet{},
		Prototype: symbols.ClassPrototypeMap{
			"format": symbols.NewClassMethod(symbols.ClassMethodOptions{
				Class:     DateTime,
				Arguments: []symbols.Class{symbols.String},
				Returns:   symbols.String,
				Handler: func(dt *DateTimeValue, fmt symbols.StringValue) (symbols.StringValue, error) {
					// TODO: write this
					return "", nil
				},
			}),
		},
	}
)

type DateTimeClass struct{}

func (DateTimeClass) Descriptors() *symbols.ClassDescriptors {
	return DateTimeDescriptors
}

type DateTimeValue struct {
	t time.Time
}

func (DateTimeValue) Class() symbols.Class {
	return DateTime
}
func (v DateTimeValue) Value() interface{} {
	return map[string]string{"$date": v.t.Format(time.RFC3339)}
}

var (
	Duration            = DurationClass{}
	DurationDescriptors = &symbols.ClassDescriptors{
		Name: "Duration",
		Constructors: symbols.ClassConstructorSet{
			symbols.Constructor(symbols.Integer, func(a symbols.IntegerValue) (DurationValue, error) {
				return DurationValue(a), nil
			}),
		},
		Operators: symbols.ClassOperatorSet{
			symbols.Operator(Duration, tokens.ADD, func(a, b DurationValue) (DurationValue, error) {
				return DurationValue(a + b), nil
			}),
			symbols.Operator(Duration, tokens.SUB, func(a, b DurationValue) (DurationValue, error) {
				return DurationValue(a - b), nil
			}),
			symbols.Operator(Duration, tokens.MUL, func(a, b DurationValue) (DurationValue, error) {
				return DurationValue(a * b), nil
			}),
			symbols.Operator(symbols.Integer, tokens.ADD, func(a DurationValue, b symbols.IntegerValue) (DurationValue, error) {
				return DurationValue(int64(a) + int64(b)), nil
			}),
			symbols.Operator(symbols.Integer, tokens.SUB, func(a DurationValue, b symbols.IntegerValue) (DurationValue, error) {
				return DurationValue(int64(a) - int64(b)), nil
			}),
		},
		Comparators: symbols.ClassComparatorSet{
			symbols.Comparator(Duration, tokens.EQUALS, func(a, b DurationValue) (bool, error) {
				return a == b, nil
			}),
			symbols.Comparator(Duration, tokens.NOT_EQUALS, func(a, b DurationValue) (bool, error) {
				return a != b, nil
			}),
			symbols.Comparator(Duration, tokens.LESS, func(a, b DurationValue) (bool, error) {
				return a < b, nil
			}),
			symbols.Comparator(Duration, tokens.GREATER, func(a, b DurationValue) (bool, error) {
				return a > b, nil
			}),
			symbols.Comparator(Duration, tokens.LESS_EQUAL, func(a, b DurationValue) (bool, error) {
				return a <= b, nil
			}),
			symbols.Comparator(Duration, tokens.GREATER_EQUAL, func(a, b DurationValue) (bool, error) {
				return a >= b, nil
			}),
		},
		Prototype: symbols.ClassPrototypeMap{
			"toUnits": symbols.NewClassMethod(symbols.ClassMethodOptions{
				Class:     Duration,
				Arguments: []symbols.Class{Duration},
				Returns:   symbols.Float,
				Handler: func(d, unit DurationValue) (symbols.FloatValue, error) {
					return symbols.FloatValue(int64(d) / int64(unit)), nil
				},
			}),
		},
	}
)

type DurationClass struct{}

func (DurationClass) Descriptors() *symbols.ClassDescriptors {
	return DurationDescriptors
}

type DurationValue int64

func (DurationValue) Class() symbols.Class {
	return Duration
}
func (v DurationValue) Value() interface{} {
	return map[string]int64{"$duration": int64(v)}
}

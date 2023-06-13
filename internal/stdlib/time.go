package stdlib

import (
	"time"

	"github.com/hntrl/hyper/internal/symbols"
)

type TimePackage struct{}

func (tp TimePackage) Get(key string) (symbols.ScopeValue, error) {
	classes := map[string]symbols.Class{
		"DateTime": DateTime,
	}
	return classes[key], nil
}

var (
	DateTime            = DateTimeClass{}
	DateTimeDescriptors = &symbols.ClassDescriptors{
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
		ClassProperties: symbols.ClassObjectPropertyMap{
			"now": symbols.NewFunction(symbols.FunctionOptions{
				Arguments: []symbols.Class{},
				Returns:   DateTime,
				Handler: func() (DateTimeValue, error) {
					return DateTimeValue{t: time.Now()}, nil
				},
			}),
		},
	}
)

type DateTimeClass struct{}

func (DateTimeClass) Name() string {
	return "DateTime"
}
func (DateTimeClass) Descriptors() *symbols.ClassDescriptors {
	return DateTimeDescriptors
}

type DateTimeValue struct {
	t time.Time
}

func (v DateTimeValue) Value() interface{} {
	return map[string]string{"$date": v.t.Format(time.RFC3339)}
}
func (DateTimeValue) Class() symbols.Class {
	return DateTime
}

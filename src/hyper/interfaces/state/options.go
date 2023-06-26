package state

import (
	"fmt"

	"github.com/hntrl/hyper/src/hyper/symbols"
)

var (
	QueryOptions            = QueryOptionsClass{}
	QueryOptionsDescriptors = &symbols.ClassDescriptors{
		Name: "QueryOptions",
		Constructors: symbols.ClassConstructorSet{
			symbols.Constructor(symbols.Map, func(val *symbols.MapValue) (*QueryOptionsValue, error) {
				qo := &QueryOptionsValue{
					Skip:  -1,
					Limit: -1,
				}
				if skipValue := val.Get("skip"); skipValue != nil {
					qo.Skip = int64(skipValue.(symbols.IntegerValue))
				}
				if limitValue := val.Get("limit"); limitValue != nil {
					qo.Limit = int64(limitValue.(symbols.IntegerValue))
				}
				return qo, nil
			}),
		},
		Properties: symbols.ClassPropertyMap{
			"skip": symbols.PropertyAttributes(symbols.PropertyOptions{
				Class: symbols.NewNilableClass(symbols.Integer),
				Getter: func(val *QueryOptionsValue) (*symbols.NilableValue, error) {
					if val.Skip == -1 {
						return symbols.NewNilableValue(symbols.Integer, nil), nil
					}
					return symbols.NewNilableValue(symbols.Integer, symbols.IntegerValue(val.Skip)), nil
				},
				Setter: func(val *QueryOptionsValue, newPropertyValue symbols.IntegerValue) error {
					if newPropertyValue <= 0 {
						return fmt.Errorf("skip must be greater than 0")
					}
					val.Skip = int64(newPropertyValue)
					return nil
				},
			}),
			"limit": symbols.PropertyAttributes(symbols.PropertyOptions{
				Class: symbols.NewNilableClass(symbols.Integer),
				Getter: func(val *QueryOptionsValue) (*symbols.NilableValue, error) {
					if val.Limit == -1 {
						return symbols.NewNilableValue(symbols.Integer, nil), nil
					}
					return symbols.NewNilableValue(symbols.Integer, symbols.IntegerValue(val.Limit)), nil
				},
				Setter: func(val *QueryOptionsValue, newPropertyValue symbols.IntegerValue) error {
					if newPropertyValue <= 0 {
						return fmt.Errorf("limit must be greater than 0")
					}
					val.Limit = int64(newPropertyValue)
					return nil
				},
			}),
		},
	}
)

type QueryOptionsClass struct{}

func (qo QueryOptionsClass) Descriptors() *symbols.ClassDescriptors {
	return QueryOptionsDescriptors
}

type QueryOptionsValue struct {
	Skip  int64
	Limit int64
}

func (qo QueryOptionsValue) Class() symbols.Class {
	return QueryOptions
}
func (qo QueryOptionsValue) Value() interface{} {
	return qo
}

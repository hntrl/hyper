package symbols_test

import (
	"github.com/hntrl/hyper/src/hyper/symbols"
)

type GenericClass struct {
	Name              string                         `hash:"ignore"`
	Properties        map[string]symbols.Class       `hash:"ignore"`
	PropertyOverrides symbols.ClassPropertyMap       `hash:"ignore"`
	Prototype         symbols.ClassPrototypeMap      `hash:"ignore"`
	ClassProperties   symbols.ClassObjectPropertyMap `hash:"ignore"`
}

func (tc GenericClass) Descriptors() *symbols.ClassDescriptors {
	propertyMap := make(symbols.ClassPropertyMap)
	for name, class := range tc.Properties {
		propertyMap[name] = symbols.PropertyAttributes(symbols.PropertyOptions{
			Class: class,
			Getter: func(val *GenericValue) (symbols.ValueObject, error) {
				return val.data[name], nil
			},
			Setter: func(val *GenericValue, newPropertyValue symbols.ValueObject) error {
				val.data[name] = newPropertyValue
				return nil
			},
		})
	}
	if tc.PropertyOverrides != nil {
		for name, property := range tc.PropertyOverrides {
			propertyMap[name] = property
		}
	}
	return &symbols.ClassDescriptors{
		Name: tc.Name,
		Constructors: symbols.ClassConstructorSet{
			symbols.Constructor(symbols.Map, func(val *symbols.MapValue) (*GenericValue, error) {
				return &GenericValue{
					class: tc,
					data:  val.Map(),
				}, nil
			}),
		},
		Properties:      propertyMap,
		Prototype:       tc.Prototype,
		ClassProperties: tc.ClassProperties,
	}
}

type GenericValue struct {
	class GenericClass
	data  map[string]symbols.ValueObject
}

func (tv GenericValue) Class() symbols.Class {
	return tv.class
}
func (tv GenericValue) Value() interface{} {
	out := make(map[string]interface{})
	for k, v := range tv.data {
		out[k] = v.Value()
	}
	return out
}

type GenericObject struct {
	handler func(key string) (symbols.ScopeValue, error)
}

func (tc GenericObject) Get(key string) (symbols.ScopeValue, error) {
	return tc.handler(key)
}

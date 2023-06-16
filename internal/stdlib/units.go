package stdlib

import (
	"github.com/hntrl/hyper/internal/symbols"
)

type UnitsPackage struct{}

func (up UnitsPackage) Get(key string) (symbols.ScopeValue, error) {
	classes := map[string]symbols.Class{
		"Dimension": Dimension,
	}
	return classes[key], nil
}

var (
	Dimension            = DimensionClass{}
	DimensionDescriptors = &symbols.ClassDescriptors{
		Name: "Dimension",
		Properties: symbols.ClassPropertyMap{
			"width": symbols.PropertyAttributes(symbols.PropertyOptions{
				Class: symbols.Float,
				Getter: func(val *DimensionValue) (symbols.FloatValue, error) {
					return symbols.FloatValue(val.Width), nil
				},
				Setter: func(val *DimensionValue, newWidth symbols.FloatValue) error {
					val.Width = float64(newWidth)
					return nil
				},
			}),
			"height": symbols.PropertyAttributes(symbols.PropertyOptions{
				Class: symbols.Float,
				Getter: func(val *DimensionValue) (symbols.FloatValue, error) {
					return symbols.FloatValue(val.Height), nil
				},
				Setter: func(val *DimensionValue, newHeiht symbols.FloatValue) error {
					val.Height = float64(newHeiht)
					return nil
				},
			}),
		},
	}
)

type DimensionClass struct{}

func (DimensionClass) Descriptors() *symbols.ClassDescriptors {
	return DimensionDescriptors
}

type DimensionValue struct {
	Width  float64
	Height float64
}

func (*DimensionValue) Class() symbols.Class {
	return Dimension
}
func (dv *DimensionValue) Value() interface{} {
	return map[string]float64{
		"width":  dv.Width,
		"height": dv.Height,
	}
}

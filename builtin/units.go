package builtin

import (
	"github.com/hntrl/lang/symbols"
)

type UnitsPackage struct{}

func (up UnitsPackage) Get(key string) (symbols.Object, error) {
	objects := map[string]symbols.Object{
		"Dimension": Dimension{},
	}
	return objects[key], nil
}

type Dimension struct {
	Width  symbols.FloatLiteral
	Height symbols.FloatLiteral
}

func (dm Dimension) ClassName() string {
	return "Dimension"
}
func (dm Dimension) Fields() map[string]symbols.Class {
	return map[string]symbols.Class{
		"width":  symbols.Float{},
		"height": symbols.Float{},
	}
}
func (dm Dimension) Constructors() symbols.ConstructorMap {
	csMap := symbols.NewConstructorMap()
	csMap.AddGenericConstructor(dm, func(fields map[string]symbols.ValueObject) (symbols.ValueObject, error) {
		return Dimension{
			Width:  fields["width"].(symbols.FloatLiteral),
			Height: fields["height"].(symbols.FloatLiteral),
		}, nil
	})
	return csMap
}
func (dm Dimension) Get(key string) (symbols.Object, error) {
	switch key {
	case "width":
		return dm.Width, nil
	case "height":
		return dm.Height, nil
	}
	return nil, nil
}

func (dm Dimension) Class() symbols.Class {
	return dm
}
func (dm Dimension) Value() interface{} {
	return map[string]float64{
		"width":  float64(dm.Width),
		"height": float64(dm.Height),
	}
}
func (dm Dimension) Set(key string, obj symbols.ValueObject) error {
	return symbols.CannotSetPropertyError(key, dm)
}

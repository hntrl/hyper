package builtin

import (
	"github.com/hntrl/lang/build"
)

type UnitsPackage struct{}

func (up UnitsPackage) Get(key string) build.Object {
	objects := map[string]build.Object{
		"Dimension": Dimension{},
	}
	return objects[key]
}

type Dimension struct {
	Width  build.FloatLiteral
	Height build.FloatLiteral
}

func (dm Dimension) ClassName() string {
	return "Dimension"
}
func (dm Dimension) Fields() map[string]build.Class {
	return map[string]build.Class{
		"width":  build.Float{},
		"height": build.Float{},
	}
}
func (dm Dimension) Constructors() build.ConstructorMap {
	csMap := build.NewConstructorMap()
	csMap.AddGenericConstructor(dm, func(fields map[string]build.ValueObject) (build.ValueObject, error) {
		return Dimension{
			Width:  fields["width"].(build.FloatLiteral),
			Height: fields["height"].(build.FloatLiteral),
		}, nil
	})
	return csMap
}
func (dm Dimension) Get(key string) build.Object {
	switch key {
	case "width":
		return dm.Width
	case "height":
		return dm.Height
	}
	return nil
}

func (dm Dimension) Class() build.Class {
	return dm
}
func (dm Dimension) Value() interface{} {
	return map[string]float64{
		"width":  float64(dm.Width),
		"height": float64(dm.Height),
	}
}
func (dm Dimension) Set(key string, obj build.ValueObject) error {
	return build.CannotSetPropertyError(key, dm)
}

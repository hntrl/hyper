package builtin

import (
	"math"

	"github.com/hntrl/lang/build"
)

type MathPackage struct{}

func (mp MathPackage) Get(key string) build.Object {
	methods := map[string]build.Object{
		// "Abs":  AbsFunction{},
		// "Acos": AcosFunction{},
		// "Asin": AsinFunction{},
		// "Atan": AtanFunction{},
		// "Atan2": Atan2Function{},
		"Ceil": build.NewFunction(build.FunctionOptions{
			Arguments: []build.Class{
				build.Float{},
			},
			Returns: build.Integer{},
			Handler: func(args []build.ValueObject, proto build.ValueObject) (build.ValueObject, error) {
				num := args[0].(build.FloatLiteral)
				return build.IntegerLiteral(math.Ceil(float64(num))), nil
			},
		}),
		// "Cos": CosFunction{},
		"Floor": build.NewFunction(build.FunctionOptions{
			Arguments: []build.Class{
				build.Float{},
			},
			Returns: build.Integer{},
			Handler: func(args []build.ValueObject, proto build.ValueObject) (build.ValueObject, error) {
				num := args[0].(build.FloatLiteral)
				return build.IntegerLiteral(math.Floor(float64(num))), nil
			},
		}),
		"Log": build.NewFunction(build.FunctionOptions{
			Arguments: []build.Class{
				build.Float{},
			},
			Returns: build.Float{},
			Handler: func(args []build.ValueObject, proto build.ValueObject) (build.ValueObject, error) {
				num := args[0].(build.FloatLiteral)
				return build.FloatLiteral(math.Log(float64(num))), nil
			},
		}),
		"Max": build.NewFunction(build.FunctionOptions{
			Arguments: []build.Class{
				build.Float{},
				build.Float{},
			},
			Returns: build.Float{},
			Handler: func(args []build.ValueObject, proto build.ValueObject) (build.ValueObject, error) {
				x := args[0].(build.FloatLiteral)
				y := args[1].(build.FloatLiteral)
				return build.FloatLiteral(math.Max(float64(x), float64(y))), nil
			},
		}),
		"Min": build.NewFunction(build.FunctionOptions{
			Arguments: []build.Class{
				build.Float{},
				build.Float{},
			},
			Returns: build.Float{},
			Handler: func(args []build.ValueObject, proto build.ValueObject) (build.ValueObject, error) {
				x := args[0].(build.FloatLiteral)
				y := args[1].(build.FloatLiteral)
				return build.FloatLiteral(math.Min(float64(x), float64(y))), nil
			},
		}),
		"Round": build.NewFunction(build.FunctionOptions{
			Arguments: []build.Class{
				build.Float{},
			},
			Returns: build.Float{},
			Handler: func(args []build.ValueObject, proto build.ValueObject) (build.ValueObject, error) {
				num := args[0].(build.FloatLiteral)
				return build.IntegerLiteral(math.Round(float64(num))), nil
			},
		}),
		// "Sin": SinFunction{},
		// "Sqrt": SqrtFunction{},
		// "Tan": TanFunction{},
	}
	return methods[key]
}

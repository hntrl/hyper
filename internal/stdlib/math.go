package stdlib

import (
	"math"

	"github.com/hntrl/hyper/internal/symbols"
)

type MathPackage struct{}

func (mp MathPackage) Get(key string) (symbols.Object, error) {
	methods := map[string]symbols.Object{
		// "Abs":  AbsFunction{},
		// "Acos": AcosFunction{},
		// "Asin": AsinFunction{},
		// "Atan": AtanFunction{},
		// "Atan2": Atan2Function{},
		"Ceil": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.Float{},
			},
			Returns: symbols.Integer{},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				num := args[0].(symbols.FloatLiteral)
				return symbols.IntegerLiteral(math.Ceil(float64(num))), nil
			},
		}),
		// "Cos": CosFunction{},
		"Floor": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.Float{},
			},
			Returns: symbols.Integer{},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				num := args[0].(symbols.FloatLiteral)
				return symbols.IntegerLiteral(math.Floor(float64(num))), nil
			},
		}),
		"Log": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.Float{},
			},
			Returns: symbols.Float{},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				num := args[0].(symbols.FloatLiteral)
				return symbols.FloatLiteral(math.Log(float64(num))), nil
			},
		}),
		"Max": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.Float{},
				symbols.Float{},
			},
			Returns: symbols.Float{},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				x := args[0].(symbols.FloatLiteral)
				y := args[1].(symbols.FloatLiteral)
				return symbols.FloatLiteral(math.Max(float64(x), float64(y))), nil
			},
		}),
		"Min": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.Float{},
				symbols.Float{},
			},
			Returns: symbols.Float{},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				x := args[0].(symbols.FloatLiteral)
				y := args[1].(symbols.FloatLiteral)
				return symbols.FloatLiteral(math.Min(float64(x), float64(y))), nil
			},
		}),
		"Round": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.Float{},
			},
			Returns: symbols.Float{},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				num := args[0].(symbols.FloatLiteral)
				return symbols.IntegerLiteral(math.Round(float64(num))), nil
			},
		}),
		// "Sin": SinFunction{},
		// "Sqrt": SqrtFunction{},
		// "Tan": TanFunction{},
	}
	return methods[key], nil
}

package stdlib

import (
	"math"

	sym "github.com/hntrl/hyper/src/hyper/symbols"
)

var mathFunctions = map[string]sym.Callable{
	"Abs": sym.NewFunction(sym.FunctionOptions{
		Arguments: []sym.Class{
			sym.Float,
		},
		Returns: sym.Float,
		Handler: func(x sym.FloatValue) (sym.FloatValue, error) {
			return sym.FloatValue(math.Abs(float64(x))), nil
		},
	}),
	"Acos": sym.NewFunction(sym.FunctionOptions{
		Arguments: []sym.Class{
			sym.Float,
		},
		Returns: sym.Float,
		Handler: func(x sym.FloatValue) (sym.FloatValue, error) {
			return sym.FloatValue(math.Acos(float64(x))), nil
		},
	}),
	"Asin": sym.NewFunction(sym.FunctionOptions{
		Arguments: []sym.Class{
			sym.Float,
		},
		Returns: sym.Float,
		Handler: func(x sym.FloatValue) (sym.FloatValue, error) {
			return sym.FloatValue(math.Asin(float64(x))), nil
		},
	}),
	"Atan": sym.NewFunction(sym.FunctionOptions{
		Arguments: []sym.Class{
			sym.Float,
		},
		Returns: sym.Float,
		Handler: func(x sym.FloatValue) (sym.FloatValue, error) {
			return sym.FloatValue(math.Atan(float64(x))), nil
		},
	}),
	"Atan2": sym.NewFunction(sym.FunctionOptions{
		Arguments: []sym.Class{
			sym.Float,
			sym.Float,
		},
		Returns: sym.Float,
		Handler: func(y sym.FloatValue, x sym.FloatValue) (sym.FloatValue, error) {
			return sym.FloatValue(math.Atan2(float64(y), float64(x))), nil
		},
	}),
	"Ceil": sym.NewFunction(sym.FunctionOptions{
		Arguments: []sym.Class{
			sym.Float,
		},
		Returns: sym.Integer,
		Handler: func(x sym.FloatValue) (sym.IntegerValue, error) {
			return sym.IntegerValue(math.Ceil(float64(x))), nil
		},
	}),
	"Cos": sym.NewFunction(sym.FunctionOptions{
		Arguments: []sym.Class{
			sym.Float,
		},
		Returns: sym.Float,
		Handler: func(x sym.FloatValue) (sym.FloatValue, error) {
			return sym.FloatValue(math.Cos(float64(x))), nil
		},
	}),
	"Floor": sym.NewFunction(sym.FunctionOptions{
		Arguments: []sym.Class{
			sym.Float,
		},
		Returns: sym.Integer,
		Handler: func(x sym.FloatValue) (sym.IntegerValue, error) {
			return sym.IntegerValue(math.Floor(float64(x))), nil
		},
	}),
	"Log": sym.NewFunction(sym.FunctionOptions{
		Arguments: []sym.Class{
			sym.Float,
		},
		Returns: sym.Float,
		Handler: func(x sym.FloatValue) (sym.FloatValue, error) {
			return sym.FloatValue(math.Log(float64(x))), nil
		},
	}),
	"Max": sym.NewFunction(sym.FunctionOptions{
		Arguments: []sym.Class{
			sym.Float,
			sym.Float,
		},
		Returns: sym.Float,
		Handler: func(x, y sym.FloatValue) (sym.FloatValue, error) {
			return sym.FloatValue(math.Max(float64(x), float64(y))), nil
		},
	}),
	"Min": sym.NewFunction(sym.FunctionOptions{
		Arguments: []sym.Class{
			sym.Float,
			sym.Float,
		},
		Returns: sym.Float,
		Handler: func(x, y sym.FloatValue) (sym.FloatValue, error) {
			return sym.FloatValue(math.Min(float64(x), float64(y))), nil
		},
	}),
	"Round": sym.NewFunction(sym.FunctionOptions{
		Arguments: []sym.Class{
			sym.Float,
		},
		Returns: sym.Integer,
		Handler: func(x sym.FloatValue) (sym.IntegerValue, error) {
			return sym.IntegerValue(math.Round(float64(x))), nil
		},
	}),
	"Sin": sym.NewFunction(sym.FunctionOptions{
		Arguments: []sym.Class{
			sym.Float,
		},
		Returns: sym.Integer,
		Handler: func(x sym.FloatValue) (sym.IntegerValue, error) {
			return sym.IntegerValue(math.Sin(float64(x))), nil
		},
	}),
	"Sqrt": sym.NewFunction(sym.FunctionOptions{
		Arguments: []sym.Class{
			sym.Float,
		},
		Returns: sym.Integer,
		Handler: func(x sym.FloatValue) (sym.IntegerValue, error) {
			return sym.IntegerValue(math.Sqrt(float64(x))), nil
		},
	}),
	"Tan": sym.NewFunction(sym.FunctionOptions{
		Arguments: []sym.Class{
			sym.Float,
		},
		Returns: sym.Integer,
		Handler: func(x sym.FloatValue) (sym.IntegerValue, error) {
			return sym.IntegerValue(math.Tan(float64(x))), nil
		},
	}),
}

type MathPackage struct{}

func (mp MathPackage) Get(key string) (sym.ScopeValue, error) {
	return mathFunctions[key], nil
}

package stdlib

import (
	"github.com/hntrl/hyper/internal/symbols"
)

type ErrorsPackage struct{}

func (ep ErrorsPackage) Get(key string) (symbols.Object, error) {
	methods := map[string]symbols.Object{
		"New": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.String{},
				symbols.String{},
			},
			Returns: symbols.Error{},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				return symbols.Error{
					Name:    string(args[0].(symbols.StringLiteral)),
					Message: string(args[1].(symbols.StringLiteral)),
				}, nil
			},
		}),
		"BadRequest": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.String{},
			},
			Returns: symbols.Error{},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				return symbols.Error{
					Name:    "BadRequest",
					Message: string(args[0].(symbols.StringLiteral)),
				}, nil
			},
		}),
		"NotFound": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.String{},
			},
			Returns: symbols.Error{},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				return symbols.Error{
					Name:    "NotFound",
					Message: string(args[0].(symbols.StringLiteral)),
				}, nil
			},
		}),
		"Unauthorized": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.String{},
			},
			Returns: symbols.Error{},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				return symbols.Error{
					Name:    "Unauthorized",
					Message: string(args[0].(symbols.StringLiteral)),
				}, nil
			},
		}),
		"InternalError": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.String{},
			},
			Returns: symbols.Error{},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				return symbols.Error{
					Name:    "InternalError",
					Message: string(args[0].(symbols.StringLiteral)),
				}, nil
			},
		}),
	}
	return methods[key], nil
}

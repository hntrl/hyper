package builtin

import (
	"github.com/hntrl/lang/build"
)

type ErrorsPackage struct{}

func (ep ErrorsPackage) Get(key string) build.Object {
	methods := map[string]build.Object{
		"New": build.NewFunction(build.FunctionOptions{
			Arguments: []build.Class{
				build.String{},
				build.String{},
			},
			Returns: build.Error{},
			Handler: func(args []build.ValueObject, proto build.ValueObject) (build.ValueObject, error) {
				return build.Error{
					Name:    string(args[0].(build.StringLiteral)),
					Message: string(args[1].(build.StringLiteral)),
				}, nil
			},
		}),
		"BadRequest": build.NewFunction(build.FunctionOptions{
			Arguments: []build.Class{
				build.String{},
			},
			Returns: build.Error{},
			Handler: func(args []build.ValueObject, proto build.ValueObject) (build.ValueObject, error) {
				return build.Error{
					Name:    "BadRequest",
					Message: string(args[0].(build.StringLiteral)),
				}, nil
			},
		}),
		"NotFound": build.NewFunction(build.FunctionOptions{
			Arguments: []build.Class{
				build.String{},
			},
			Returns: build.Error{},
			Handler: func(args []build.ValueObject, proto build.ValueObject) (build.ValueObject, error) {
				return build.Error{
					Name:    "NotFound",
					Message: string(args[0].(build.StringLiteral)),
				}, nil
			},
		}),
		"Unauthorized": build.NewFunction(build.FunctionOptions{
			Arguments: []build.Class{
				build.String{},
			},
			Returns: build.Error{},
			Handler: func(args []build.ValueObject, proto build.ValueObject) (build.ValueObject, error) {
				return build.Error{
					Name:    "Unauthorized",
					Message: string(args[0].(build.StringLiteral)),
				}, nil
			},
		}),
		"InternalError": build.NewFunction(build.FunctionOptions{
			Arguments: []build.Class{
				build.String{},
			},
			Returns: build.Error{},
			Handler: func(args []build.ValueObject, proto build.ValueObject) (build.ValueObject, error) {
				return build.Error{
					Name:    "InternalError",
					Message: string(args[0].(build.StringLiteral)),
				}, nil
			},
		}),
	}
	return methods[key]
}

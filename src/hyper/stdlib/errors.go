package stdlib

import (
	sym "github.com/hntrl/hyper/src/hyper/symbols"
)

var errorFunctions = map[string]sym.Callable{
	"New": sym.NewFunction(sym.FunctionOptions{
		Arguments: []sym.Class{
			sym.String,
			sym.String,
		},
		Returns: sym.Error,
		Handler: func(name sym.StringValue, message sym.StringValue) (sym.ErrorValue, error) {
			return sym.ErrorValue{
				Name:    string(name),
				Message: string(message),
			}, nil
		},
	}),
	"BadRequest": sym.NewFunction(sym.FunctionOptions{
		Arguments: []sym.Class{
			sym.String,
		},
		Returns: sym.Error,
		Handler: func(message sym.StringValue) (sym.ErrorValue, error) {
			return sym.ErrorValue{
				Name:    "BadRequest",
				Message: string(message),
			}, nil
		},
	}),
	"NotFound": sym.NewFunction(sym.FunctionOptions{
		Arguments: []sym.Class{
			sym.String,
		},
		Returns: sym.Error,
		Handler: func(message sym.StringValue) (sym.ErrorValue, error) {
			return sym.ErrorValue{
				Name:    "NotFound",
				Message: string(message),
			}, nil
		},
	}),
	"Unauthorized": sym.NewFunction(sym.FunctionOptions{
		Arguments: []sym.Class{
			sym.String,
		},
		Returns: sym.Error,
		Handler: func(message sym.StringValue) (sym.ErrorValue, error) {
			return sym.ErrorValue{
				Name:    "Unauthorized",
				Message: string(message),
			}, nil
		},
	}),
	"InternalError": sym.NewFunction(sym.FunctionOptions{
		Arguments: []sym.Class{
			sym.String,
		},
		Returns: sym.Error,
		Handler: func(message sym.StringValue) (sym.ErrorValue, error) {
			return sym.ErrorValue{
				Name:    "InternalError",
				Message: string(message),
			}, nil
		},
	}),
}

type ErrorsPackage struct{}

func (ep ErrorsPackage) Get(key string) (sym.ScopeValue, error) {
	return errorFunctions[key], nil
}

package stdlib

import "github.com/hntrl/hyper/internal/symbols"

var Packages = map[string]symbols.Object{
	"math":    MathPackage{},
	"errors":  ErrorsPackage{},
	"units":   UnitsPackage{},
	"request": RequestPackage{},
	"mime":    MimeTypesPackage{},
}
